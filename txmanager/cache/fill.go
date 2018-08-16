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
	"github.com/Loopring/relay-cluster/dao"
	"encoding/json"
)

// 该缓存模块用于处理fill数据
// 从extractor过来的数据经过排序后,fill先到，然后是transfer
// 如果是普通用户,我们只需要处理fill,如果是miner我们需要只需要处理transfer

const (
	FillOwnerPrefix  = "txm_fill_owner_"
	FillOwnerTtl     = 864000  // todo 临时数据,只存储10分钟,系统性宕机后无法重启后丢失?
	FillSearchPreKey = "FILL_" // list cache
)

func SetFillOwnerCache(txhash common.Hash, owner common.Address) error {
	key := generateFillOwnerKey(txhash)
	field := []byte(owner.Hex())
	return cache.SAdd(key, FillOwnerTtl, field)
}

func ExistFillOwnerCache(txhash common.Hash, owner common.Address) (bool, error) {
	key := generateFillOwnerKey(txhash)
	field := []byte(owner.Hex())
	return cache.SIsMember(key, field)
}

func generateFillOwnerKey(txhash common.Hash) string {
	return FillOwnerPrefix + txhash.Hex()
}

func GetCacheFills(key string, res *dao.PageResult) (err error, get bool) {
	if fillsByte, err := cache.Get(FillSearchPreKey + key); err != nil {
		return err, false
	} else if len(fillsByte) > 0 {
		data := make([]interface{}, 0)
		fills := make([]dao.FillEvent, 0)
		json.Unmarshal(fillsByte, &res)
		if r, _ := json.Marshal(res.Data); r != nil {
			json.Unmarshal(r, &fills)
		}
		for _, v := range fills {
			data = append(data, v)
		}
		res.Data = data
		log.Debugf("[GetFills Cache] from cache key: %s", FillSearchPreKey+key)
		return err, true
	}
	return nil, false
}

func SaveCacheFills(key string, res *dao.PageResult, ttl int64) {
	value, _ := json.Marshal(res)
	log.Debugf("[GetFills Cache] save cache key: %s", FillSearchPreKey+key)
	cache.Set(FillSearchPreKey+key, value, ttl)
}

func DelFillCacheByOwner(owners []string) {
	keyStrs := make([]string, 0)
	if owners != nil && len(owners) > 0 {
		for _, owner := range owners {
			keys, _ := cache.Keys(strings.ToUpper(FillSearchPreKey + "OWNER:" + owner + "*"))
			for _, key := range keys {
				log.Debugf("[clear cache] clear key: %s", string(key))
				keyStrs = append(keyStrs, string(key))
			}
		}
	}
	cache.Dels(keyStrs)
}
