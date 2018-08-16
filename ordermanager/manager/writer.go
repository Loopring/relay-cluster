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

package manager

import (
	"fmt"
	"github.com/Loopring/relay-cluster/dao"
	cm "github.com/Loopring/relay-cluster/ordermanager/common"
	"github.com/Loopring/relay-lib/log"
	util "github.com/Loopring/relay-lib/marketutil"
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/Loopring/relay-cluster/ordermanager/cache"
)

func MinerOrders(delegate, tokenS, tokenB common.Address, length int, reservedTime, startBlockNumber, endBlockNumber int64, filterOrderHashLists ...*types.OrderDelayList) []*types.OrderState {
	var list []*types.OrderState

	var (
		modelList []*dao.Order
		err       error
	)

	for _, orderDelay := range filterOrderHashLists {
		orderHashes := []string{}
		for _, hash := range orderDelay.OrderHash {
			orderHashes = append(orderHashes, hash.Hex())
		}
		if len(orderHashes) > 0 && orderDelay.DelayedCount != 0 {
			if err = rds.MarkMinerOrders(orderHashes, orderDelay.DelayedCount); err != nil {
				log.Debugf("order manager,provide orders for miner error:%s", err.Error())
			}
		}
	}

	// 从数据库获取订单
	if modelList, err = rds.GetOrdersForMiner(delegate.Hex(), tokenS.Hex(), tokenB.Hex(), length, cm.ValidMinerStatus, reservedTime, startBlockNumber, endBlockNumber); err != nil {
		log.Errorf("err:%s", err.Error())
		return list
	}

	for _, v := range modelList {
		state := &types.OrderState{}
		v.ConvertUp(state)
		list = append(list, state)

		//if um.InWhiteList(state.RawOrder.Owner) {
		//	list = append(list, state)
		//} else {
		//	log.Debugf("order manager,owner:%s not in white list", state.RawOrder.Owner.Hex())
		//}
	}

	return list
}

func UpdateBroadcastTimeByHash(hash common.Hash, bt int) error {
	return rds.UpdateBroadcastTimeByHash(hash.Hex(), bt)
}

func FlexCancelOrder(event *types.FlexCancelOrderEvent) error {
	if types.IsZeroAddress(event.Owner) {
		return fmt.Errorf("params owner invalid")
	}

	validStatus := cm.ValidFlexCancelStatus
	status := types.ORDER_FLEX_CANCEL

	var nums int64 = 0
	switch event.Type {
	case types.FLEX_CANCEL_BY_HASH:
		if types.IsZeroHash(event.OrderHash) {
			return fmt.Errorf("params orderhash invalid")
		}
		nums = rds.FlexCancelOrderByHash(event.Owner, event.OrderHash, validStatus, status)

	case types.FLEX_CANCEL_BY_OWNER:
		nums = rds.FlexCancelOrderByOwner(event.Owner, validStatus, status)

	case types.FLEX_CANCEL_BY_TIME:
		if event.CutoffTime <= 0 {
			return fmt.Errorf("params cutoffTimeStamp invalid")
		}
		nums = rds.FlexCancelOrderByTime(event.Owner, event.CutoffTime, validStatus, status)

	case types.FLEX_CANCEL_BY_MARKET:
		market, err := util.WrapMarketByAddress(event.TokenS.Hex(), event.TokenB.Hex())
		if err != nil {
			return fmt.Errorf("params market invalid")
		}
		nums = rds.FlexCancelOrderByMarket(event.Owner, event.CutoffTime, market, validStatus, status)

	default:
		return fmt.Errorf("event type invalid")
	}

	if nums == 0 {
		return fmt.Errorf("no valid order exist")
	}

	cache.DelOrderCacheByOwner(event.Owner)
	return nil
}
