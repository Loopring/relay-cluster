/*

  Copyright 2017 Loopring Project Ltd (Loopring Foundation).

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.

*/

package cache

import (
	"github.com/Loopring/relay-lib/cache"
	"github.com/ethereum/go-ethereum/common"
	"strings"
	"github.com/Loopring/relay-lib/log"
	"encoding/json"
	"github.com/Loopring/relay-cluster/dao"
)

// 注意:这里我们不对cache设置过期时间,如果设定过期时间,会导致event通知到ordermanager后,与订单无关的用户查询mysql,消耗太大

const (
	UserPendingOrderKeyPrefix = "om_pending_ordertx_"
	// just for getting list
	OrderSearchPreKey = "ORDER_"
)

func SetPendingOrder(owner common.Address, orderhash common.Hash) error {
	key := getKey(owner)
	member := getMember(orderhash)
	cache.SAdd(key, 0, member)
	return nil
}

func DelPendingOrder(owner common.Address, orderhash common.Hash) error {
	key := getKey(owner)
	member := getMember(orderhash)
	cache.SRem(key, member)
	return nil
}

func ExistPendingOrder(owner common.Address, orderhash common.Hash) bool {
	key := getKey(owner)
	member := getMember(orderhash)
	ok, _ := cache.SIsMember(key, member)
	return ok
}

func GetPendingOrders(owner common.Address) []common.Hash {
	var list []common.Hash
	key := getKey(owner)

	if ok, _ := cache.Exists(key); !ok {
		return list
	}

	bslist, _ := cache.SMembers(key)
	for _, bs := range bslist {
		orderhash := setMember(bs)
		list = append(list, orderhash)
	}

	return list
}

func getKey(owner common.Address) string {
	return UserPendingOrderKeyPrefix + owner.Hex()
}

func getMember(orderhash common.Hash) []byte {
	return []byte(orderhash.Hex())
}

func setMember(bs []byte) common.Hash {
	return common.HexToHash(string(bs))
}

func GetCacheOrders(key string, res *dao.PageResult) (err error, get bool) {
	if ordersByte, err := cache.Get(OrderSearchPreKey + key); err != nil {
		return err, false
	} else if len(ordersByte) > 0 {
		data := make([]interface{}, 0)
		orders := make([]dao.Order, 0)
		json.Unmarshal(ordersByte, &res)
		if r, _ := json.Marshal(res.Data); r != nil {
			json.Unmarshal(r, &orders)
		}
		for _, v := range orders {
			data = append(data, v)
		}
		res.Data = data
		log.Debugf("[GetOrders Cache] from cache key: %s", OrderSearchPreKey+key)
		return err, true
	}
	return nil, false
}

func SaveCacheOrders(key string, res *dao.PageResult, ttl int64) {
	value, _ := json.Marshal(res)
	log.Debugf("[GetOrders Cache] save cache key: %s", OrderSearchPreKey+key)
	cache.Set(OrderSearchPreKey+key, value, ttl)
}

func DelOrderCacheByOwner(owners []string) {
	keyStrs := make([]string, 0)
	if owners != nil && len(owners) > 0 {
		for _, owner := range owners {
			keys, _ := cache.Keys(strings.ToUpper(OrderSearchPreKey + "OWNER:" + owner + "*"))
			for _, key := range keys {
				log.Debugf("[clear cache] clear key: %s", string(key))
				keyStrs = append(keyStrs, string(key))
			}
		}
	}
	cache.Dels(keyStrs)
}
