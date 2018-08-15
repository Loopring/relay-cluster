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

package viewer

import (
	"github.com/Loopring/relay-cluster/dao"
	"github.com/Loopring/relay-cluster/ordermanager/cache"
	. "github.com/Loopring/relay-cluster/ordermanager/common"
	"github.com/Loopring/relay-lib/log"
	"github.com/Loopring/relay-lib/marketcap"
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"strconv"
	"strings"
	cache2 "github.com/Loopring/relay-cluster/txmanager/cache"
)

type OrderViewer interface {
	GetOrderBook(protocol, tokenS, tokenB common.Address, length int) ([]types.OrderState, error)
	GetOrders(query map[string]interface{}, statusList []types.OrderStatus, pageIndex, pageSize int) (dao.PageResult, error)
	GetLatestOrders(query map[string]interface{}, length int) ([]types.OrderState, error)
	GetOrderByHash(hash common.Hash) (*types.OrderState, error)
	GetOrdersByHashes(hash []common.Hash) ([]types.OrderState, error)
	FillsPageQuery(query map[string]interface{}, pageIndex, pageSize int) (dao.PageResult, error)
	GetLatestFills(query map[string]interface{}, limit int) ([]dao.FillEvent, error)
	FindFillsByRingHash(ringHash common.Hash) (result []dao.FillEvent, err error)
	RingMinedPageQuery(query map[string]interface{}, pageIndex, pageSize int) (dao.PageResult, error)
	IsOrderCutoff(protocol, owner, token1, token2 common.Address, validsince *big.Int) bool
	GetFrozenAmount(owner common.Address, token common.Address, statusSet []types.OrderStatus, delegateAddress common.Address) (*big.Int, error)
	GetFrozenLRCFee(owner common.Address, statusSet []types.OrderStatus) (*big.Int, error)
}

type OrderViewerImpl struct {
	mc          marketcap.MarketCapProvider
	rds         *dao.RdsService
	cutoffCache *CutoffCache
}

func NewOrderViewer(options *OrderManagerOptions,
	rds *dao.RdsService,
	market marketcap.MarketCapProvider) *OrderViewerImpl {

	var viewer OrderViewerImpl
	viewer.mc = market
	viewer.rds = rds
	viewer.cutoffCache = NewCutoffCache(options.CutoffCacheCleanTime)

	if cache.Invalid() {
		cache.Initialize(viewer.rds)
	}

	return &viewer
}

func (om *OrderViewerImpl) GetOrderBook(protocol, tokenS, tokenB common.Address, length int) ([]types.OrderState, error) {
	var list []types.OrderState
	models, err := om.rds.GetOrderBook(protocol, tokenS, tokenB, length)
	if err != nil {
		return list, err
	}

	for _, v := range models {
		var state types.OrderState
		if err := v.ConvertUp(&state); err != nil {
			continue
		}
		list = append(list, state)
	}

	return list, nil
}

func (om *OrderViewerImpl) GetOrders(query map[string]interface{}, statusList []types.OrderStatus, pageIndex, pageSize int) (dao.PageResult, error) {
	var (
		pageRes dao.PageResult
	)
	sL := make([]int, 0)
	for _, s := range statusList {
		sL = append(sL, int(s))
	}

	var tmp dao.PageResult
	key := orderToKey(query, statusList, pageIndex, pageSize)
	if err, get := cache.GetCacheOrders(key, &tmp); err != nil && !get {
		tmp, err = om.rds.OrderPageQuery(query, sL, pageIndex, pageSize)
		if err != nil {
			return pageRes, err
		}
		cache.SaveCacheOrders(key, &tmp, 24*3600*7)
	}

	pageRes.PageIndex = tmp.PageIndex
	pageRes.PageSize = tmp.PageSize
	pageRes.Total = tmp.Total

	for _, v := range tmp.Data {
		var state types.OrderState
		model := v.(dao.Order)
		if err := model.ConvertUp(&state); err != nil {
			log.Debug("convertUp error occurs " + err.Error())
			continue
		}
		pageRes.Data = append(pageRes.Data, state)
	}

	return pageRes, nil
}

func (om *OrderViewerImpl) GetLatestOrders(query map[string]interface{}, length int) (rst []types.OrderState, err error) {
	tmp, err := om.rds.GetLatestOrders(query, length)
	if err != nil {
		return nil, err
	}

	rst = make([]types.OrderState, 0)

	for _, v := range tmp {
		var state types.OrderState
		if err := v.ConvertUp(&state); err != nil {
			log.Debug("convertUp error occurs " + err.Error())
			continue
		}
		rst = append(rst, state)
	}
	return rst, nil
}

func (om *OrderViewerImpl) GetOrderByHash(hash common.Hash) (orderState *types.OrderState, err error) {
	var result types.OrderState
	order, err := om.rds.GetOrderByHash(hash)
	if err != nil {
		return nil, err
	}

	if err := order.ConvertUp(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (om *OrderViewerImpl) GetOrdersByHashes(orders []common.Hash) (orderState []types.OrderState, err error) {
	var result types.OrderState
	orderList, err := om.rds.GetOrdersByHashes(orders)
	if err != nil {
		return nil, err
	}
	rst := make([]types.OrderState, 0)
	for _, order := range orderList {
		if err := order.ConvertUp(&result); err != nil {
			return nil, err
		}
		rst = append(rst, result)
	}

	return rst, nil
}

func (om *OrderViewerImpl) FillsPageQuery(query map[string]interface{}, pageIndex, pageSize int) (result dao.PageResult, err error) {
	key := orderToKey(query, nil, pageIndex, pageSize)
	if err, get := cache2.GetCacheFills(key, &result); err != nil || !get {
		result, err = om.rds.FillsPageQuery(query, pageIndex, pageSize)
		cache2.SaveCacheFills(key, &result, 24*3600*7)
	}
	return
}

func (om *OrderViewerImpl) GetLatestFills(query map[string]interface{}, limit int) (result []dao.FillEvent, err error) {
	return om.rds.GetLatestFills(query, limit)
}

func (om *OrderViewerImpl) FindFillsByRingHash(ringHash common.Hash) (result []dao.FillEvent, err error) {
	return om.rds.FindFillsByRingHash(ringHash)
}

func (om *OrderViewerImpl) RingMinedPageQuery(query map[string]interface{}, pageIndex, pageSize int) (result dao.PageResult, err error) {
	return om.rds.RingMinedPageQuery(query, pageIndex, pageSize)
}

func (om *OrderViewerImpl) IsOrderCutoff(protocol, owner, token1, token2 common.Address, validsince *big.Int) bool {
	return om.cutoffCache.IsOrderCutoff(protocol, owner, token1, token2, validsince)
}

func (om *OrderViewerImpl) GetFrozenAmount(owner common.Address, token common.Address, statusSet []types.OrderStatus, delegateAddress common.Address) (*big.Int, error) {
	orderList, err := om.rds.GetFrozenAmount(owner, token, statusSet, delegateAddress)
	if err != nil {
		return nil, err
	}

	totalAmount := big.NewInt(0)

	if len(orderList) == 0 {
		return totalAmount, nil
	}

	for _, v := range orderList {
		var state types.OrderState
		if err := v.ConvertUp(&state); err != nil {
			continue
		}
		rs, _ := state.RemainedAmount()
		totalAmount.Add(totalAmount, rs.Num())
	}

	return totalAmount, nil
}

func (om *OrderViewerImpl) GetFrozenLRCFee(owner common.Address, statusSet []types.OrderStatus) (*big.Int, error) {
	orderList, err := om.rds.GetFrozenLrcFee(owner, statusSet)
	if err != nil {
		return nil, err
	}

	totalAmount := big.NewInt(0)

	if len(orderList) == 0 {
		return totalAmount, nil
	}

	for _, v := range orderList {
		lrcFee, _ := new(big.Int).SetString(v.LrcFee, 0)
		totalAmount.Add(totalAmount, lrcFee)
	}

	return totalAmount, nil
}

func orderToKey(query map[string]interface{}, statusList []types.OrderStatus, pageIndex, pageSize int) (res string) {
	keys := make([]string, 0)
	mapKeys := []string{"owner", "delegate_address", "market", "side", "order_hash", "order_type", "ring_hash"}
	for _, mk := range mapKeys {
		if key, ok := query[mk]; ok {
			keys = append(append(keys, mk), key.(string))
		}
	}
	if statusList != nil && len(statusList) > 0 {
		statusStr := ""
		for _, s := range statusList {
			statusStr += strconv.Itoa(int(s))
		}
		keys = append(append(keys, "status"), statusStr)
	}
	if pageIndex <= 0 {
		pageIndex = 1
	}

	if pageSize <= 0 {
		pageSize = 20
	}
	keys = append(append(keys, "index"), strconv.Itoa(pageIndex))
	keys = append(append(keys, "size"), strconv.Itoa(pageSize))
	return strings.ToUpper(strings.Join(keys, ":"))
}
