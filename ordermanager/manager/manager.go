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
	omcm "github.com/Loopring/relay-cluster/ordermanager/common"
	notify "github.com/Loopring/relay-cluster/util"
	"github.com/Loopring/relay-lib/eventemitter"
	"github.com/Loopring/relay-lib/log"
	"github.com/Loopring/relay-lib/marketcap"
	util "github.com/Loopring/relay-lib/marketutil"
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type OrderManager interface {
	Start()
	Stop()
}

type OrderManagerImpl struct {
	options                 *omcm.OrderManagerOptions
	rds                     *dao.RdsService
	processor               *ForkProcessor
	cutoffCache             *omcm.CutoffCache
	mc                      marketcap.MarketCapProvider
	newOrderWatcher         *eventemitter.Watcher
	ringMinedWatcher        *eventemitter.Watcher
	fillOrderWatcher        *eventemitter.Watcher
	cancelOrderWatcher      *eventemitter.Watcher
	cutoffOrderWatcher      *eventemitter.Watcher
	cutoffPairWatcher       *eventemitter.Watcher
	forkWatcher             *eventemitter.Watcher
	warningWatcher          *eventemitter.Watcher
	submitRingMethodWatcher *eventemitter.Watcher
}

func NewOrderManager(
	options *omcm.OrderManagerOptions,
	rds *dao.RdsService,
	market marketcap.MarketCapProvider) *OrderManagerImpl {

	om := &OrderManagerImpl{}
	om.options = options
	om.rds = rds
	om.processor = NewForkProcess(om.rds, market)
	om.mc = market
	om.cutoffCache = omcm.NewCutoffCache(options.CutoffCacheCleanTime)

	return om
}

// Start start orderbook as a service
func (om *OrderManagerImpl) Start() {
	om.newOrderWatcher = &eventemitter.Watcher{Concurrent: false, Handle: om.handleGatewayOrder}
	om.ringMinedWatcher = &eventemitter.Watcher{Concurrent: false, Handle: om.handleRingMined}
	om.fillOrderWatcher = &eventemitter.Watcher{Concurrent: false, Handle: om.handleOrderFilled}
	om.cancelOrderWatcher = &eventemitter.Watcher{Concurrent: false, Handle: om.handleOrderCancelled}
	om.cutoffOrderWatcher = &eventemitter.Watcher{Concurrent: false, Handle: om.handleCutoff}
	om.cutoffPairWatcher = &eventemitter.Watcher{Concurrent: false, Handle: om.handleCutoffPair}
	om.forkWatcher = &eventemitter.Watcher{Concurrent: false, Handle: om.handleFork}
	om.warningWatcher = &eventemitter.Watcher{Concurrent: false, Handle: om.handleWarning}
	om.submitRingMethodWatcher = &eventemitter.Watcher{Concurrent: false, Handle: om.handleSubmitRingMethod}

	eventemitter.On(eventemitter.NewOrder, om.newOrderWatcher)
	eventemitter.On(eventemitter.RingMined, om.ringMinedWatcher)
	eventemitter.On(eventemitter.OrderFilled, om.fillOrderWatcher)
	eventemitter.On(eventemitter.CancelOrder, om.cancelOrderWatcher)
	eventemitter.On(eventemitter.CutoffAll, om.cutoffOrderWatcher)
	eventemitter.On(eventemitter.CutoffPair, om.cutoffPairWatcher)
	eventemitter.On(eventemitter.ChainForkDetected, om.forkWatcher)
	eventemitter.On(eventemitter.ExtractorWarning, om.warningWatcher)
	eventemitter.On(eventemitter.Miner_SubmitRing_Method, om.submitRingMethodWatcher)
}

func (om *OrderManagerImpl) Stop() {
	eventemitter.Un(eventemitter.NewOrder, om.newOrderWatcher)
	eventemitter.Un(eventemitter.RingMined, om.ringMinedWatcher)
	eventemitter.Un(eventemitter.OrderFilled, om.fillOrderWatcher)
	eventemitter.Un(eventemitter.CancelOrder, om.cancelOrderWatcher)
	eventemitter.Un(eventemitter.CutoffAll, om.cutoffOrderWatcher)
	eventemitter.Un(eventemitter.ChainForkDetected, om.forkWatcher)
	eventemitter.Un(eventemitter.ExtractorWarning, om.warningWatcher)
	eventemitter.Un(eventemitter.Miner_SubmitRing_Method, om.submitRingMethodWatcher)
}

func (om *OrderManagerImpl) handleFork(input eventemitter.EventData) error {
	log.Debugf("order manager processing chain fork......")

	om.Stop()
	if err := om.processor.Fork(input.(*types.ForkedEvent)); err != nil {
		log.Fatalf("order manager,handle fork error:%s", err.Error())
	}
	om.Start()

	return nil
}

func (om *OrderManagerImpl) handleWarning(input eventemitter.EventData) error {
	log.Debugf("order manager processing extractor warning")
	om.Stop()
	return nil
}

func (om *OrderManagerImpl) handleSubmitRingMethod(input eventemitter.EventData) error {
	event := input.(*types.SubmitRingMethodEvent)

	if event.Status != types.TX_STATUS_FAILED && event.Status != types.TX_STATUS_PENDING {
		return nil
	}

	if err := om.saveSubmitRingEvent(event); err != nil {
		log.Errorf(err.Error())
	}

	return nil
}

// 所有来自gateway的订单都是新订单
func (om *OrderManagerImpl) handleGatewayOrder(input eventemitter.EventData) error {
	state := input.(*types.OrderState)

	model, err := NewOrderEntity(state, om.mc, nil)
	if err != nil {
		log.Errorf("order manager,handle gateway order:%s error", state.RawOrder.Hash.Hex())
		return err
	}

	if err = om.rds.Add(model); err != nil {
		return err
	}

	log.Debugf("order manager,handle gateway order,order.hash:%s amountS:%s", state.RawOrder.Hash.Hex(), state.RawOrder.AmountS.String())

	notify.NotifyOrderUpdate(state)
	return nil
}

func (om *OrderManagerImpl) handleRingMined(input eventemitter.EventData) error {
	event := input.(*types.RingMinedEvent)

	if event.Status != types.TX_STATUS_SUCCESS {
		return nil
	}

	if err := om.saveRingMinedEvent(event); err != nil {
		log.Errorf(err.Error())
	}

	return nil
}

func (om *OrderManagerImpl) handleOrderFilled(input eventemitter.EventData) error {
	event := input.(*types.OrderFilledEvent)

	if event.Status != types.TX_STATUS_SUCCESS {
		return nil
	}

	// find fill event
	if _, err := om.rds.FindFillEvent(event.TxHash.Hex(), event.FillIndex.Int64()); err == nil {
		log.Debugf("order manager,handle order filled event tx:%s fillIndex:%d fill already exist", event.TxHash.String(), event.FillIndex)
		return nil
	}

	// get rds.Order and types.OrderState
	state := &types.OrderState{UpdatedBlock: event.BlockNumber}
	model, err := om.rds.GetOrderByHash(event.OrderHash)
	if err != nil {
		return err
	}
	if err := model.ConvertUp(state); err != nil {
		return err
	}

	newFillModel := &dao.FillEvent{}
	newFillModel.ConvertDown(event)
	newFillModel.Fork = false
	newFillModel.OrderType = state.RawOrder.OrderType
	newFillModel.Side = util.GetSide(util.AddressToAlias(event.TokenS.Hex()), util.AddressToAlias(event.TokenB.Hex()))
	newFillModel.Market, _ = util.WrapMarketByAddress(event.TokenB.Hex(), event.TokenS.Hex())

	if err := om.rds.Add(newFillModel); err != nil {
		log.Debugf("order manager,handle order filled event tx:%s fillIndex:%s orderhash:%s error:insert failed",
			event.TxHash.Hex(), event.FillIndex.String(), event.OrderHash.Hex())
		return err
	}

	// judge order status
	if state.Status == types.ORDER_CUTOFF || state.Status == types.ORDER_FINISHED || state.Status == types.ORDER_UNKNOWN {
		log.Debugf("order manager,handle order filled event tx:%s fillIndex:%s orderhash:%s status:%d invalid",
			event.TxHash.Hex(), event.FillIndex.String(), event.OrderHash.Hex(), state.Status)
		return nil
	}

	// calculate dealt amount
	state.UpdatedBlock = event.BlockNumber
	state.DealtAmountS = new(big.Int).Add(state.DealtAmountS, event.AmountS)
	state.DealtAmountB = new(big.Int).Add(state.DealtAmountB, event.AmountB)
	state.SplitAmountS = new(big.Int).Add(state.SplitAmountS, event.SplitS)
	state.SplitAmountB = new(big.Int).Add(state.SplitAmountB, event.SplitB)

	// update order status
	SettleOrderStatus(state, om.mc, false)

	// update rds.Order
	if err := model.ConvertDown(state); err != nil {
		log.Errorf(err.Error())
		return err
	}
	if err := om.rds.UpdateOrderWhileFill(state.RawOrder.Hash, state.Status, state.DealtAmountS, state.DealtAmountB, state.SplitAmountS, state.SplitAmountB, state.UpdatedBlock); err != nil {
		return err
	}

	log.Debugf("order manager,handle order filled event tx:%s, fillIndex:%s, orderhash:%s, dealAmountS:%s, dealtAmountB:%s",
		event.TxHash.Hex(), event.FillIndex.String(), state.RawOrder.Hash.Hex(), state.DealtAmountS.String(), state.DealtAmountB.String())

	notify.NotifyOrderFilled(newFillModel)
	return nil
}

func (om *OrderManagerImpl) handleOrderCancelled(input eventemitter.EventData) error {
	event := input.(*types.OrderCancelledEvent)

	var (
		state       = &types.OrderState{}
		orderentity = &dao.Order{}
		err         error
	)

	if err = om.saveCancelEvent(event); err != nil {
		log.Errorf(err.Error())
		return nil
	}

	// get rds.Order and types.OrderState
	orderentity, err = om.rds.GetOrderByHash(event.OrderHash)
	if err != nil {
		return err
	}
	if err := orderentity.ConvertUp(state); err != nil {
		return err
	}

	// calculate remainAmount and cancelled amount should be saved whether order is finished or not
	if state.RawOrder.BuyNoMoreThanAmountB {
		state.CancelledAmountB = new(big.Int).Add(state.CancelledAmountB, event.AmountCancelled)
		log.Debugf("order manager,handle order cancelled event tx:%s, order:%s cancelled amountb:%s", event.TxHash.Hex(), state.RawOrder.Hash.Hex(), state.CancelledAmountB.String())
	} else {
		state.CancelledAmountS = new(big.Int).Add(state.CancelledAmountS, event.AmountCancelled)
		log.Debugf("order manager,handle order cancelled event tx:%s, order:%s cancelled amounts:%s", event.TxHash.Hex(), state.RawOrder.Hash.Hex(), state.CancelledAmountS.String())
	}

	// update order status
	SettleOrderStatus(state, om.mc, true)
	state.UpdatedBlock = event.BlockNumber

	// update rds.Order
	if err = orderentity.ConvertDown(state); err != nil {
		return err
	}
	if err = om.rds.UpdateOrderWhileCancel(state.RawOrder.Hash, state.Status, state.CancelledAmountS, state.CancelledAmountB, state.UpdatedBlock); err != nil {
		return err
	}

	notify.NotifyOrderUpdate(state)

	return nil
}

// 所有cutoff event都应该存起来,但不是所有event都会影响订单
func (om *OrderManagerImpl) handleCutoff(input eventemitter.EventData) error {
	event := input.(*types.CutoffEvent)

	// save event
	if err := om.saveCutoffEvent(event); err != nil {
		log.Errorf(err.Error())
		return nil
	}

	if event.Status != types.TX_STATUS_SUCCESS {
		return nil
	}

	// update cache
	if lastCutoff := om.cutoffCache.GetCutoff(event.Protocol, event.Owner); event.Cutoff.Cmp(lastCutoff) < 0 {
		log.Debugf("order manager,handle cutoff event tx:%s, protocol:%s - owner:%s lastCutofftime:%s > currentCutoffTime:%s", event.TxHash.Hex(), event.Protocol.Hex(), event.Owner.Hex(), lastCutoff.String(), event.Cutoff.String())
		return nil
	}
	om.cutoffCache.UpdateCutoff(event.Protocol, event.Owner, event.Cutoff)

	// update order status
	om.rds.SetCutOffOrders(event.OrderHashList, event.BlockNumber)

	notify.NotifyCutoff(event)

	return nil
}

func (om *OrderManagerImpl) handleCutoffPair(input eventemitter.EventData) error {
	event := input.(*types.CutoffPairEvent)

	if err := om.saveCutoffPairEvent(event); err != nil {
		log.Errorf(err.Error())
		return nil
	}

	if event.Status != types.TX_STATUS_SUCCESS {
		return nil
	}

	// save cutoff cache
	if last := om.cutoffCache.GetCutoffPair(event.Protocol, event.Owner, event.Token1, event.Token2); event.Cutoff.Cmp(last) < 0 {
		log.Debugf("order manager,handle cutoffPair event tx:%s, protocol:%s - owner:%s lastCutoffPairtime:%s > currentCutoffPairTime:%s", event.TxHash.Hex(), event.Protocol.Hex(), event.Owner.Hex(), last.String(), event.Cutoff.String())
		return nil
	}
	om.cutoffCache.UpdateCutoffPair(event.Protocol, event.Owner, event.Token1, event.Token2, event.Cutoff)

	// cutoff orders
	om.rds.SetCutOffOrders(event.OrderHashList, event.BlockNumber)

	notify.NotifyCutoffPair(event)

	return nil
}

func (om *OrderManagerImpl) saveSubmitRingEvent(event *types.SubmitRingMethodEvent) error {
	var (
		model = &dao.RingMinedEvent{}
		err   error
	)

	if model, err = om.rds.FindRingMined(event.TxHash.Hex()); err != nil {
		log.Debugf("order manager,handle submitRing method,tx:%s status:%s inserted", event.TxHash.Hex(), types.StatusStr(event.Status))
		model.FromSubmitRingMethod(event)
		return om.rds.Add(model)
	}

	if model.Status != uint8(event.Status) {
		log.Debugf("order manager,handle submitRing method,tx:%s status:%s updated", event.TxHash.Hex(), types.StatusStr(event.Status))
		model.FromSubmitRingMethod(event)
		return om.rds.Save(model)
	}

	return fmt.Errorf("order manager,handle submitRing method,tx %s already exist", event.TxHash.Hex())
}

func (om *OrderManagerImpl) saveRingMinedEvent(event *types.RingMinedEvent) error {
	var (
		model = &dao.RingMinedEvent{}
		err   error
	)

	if model, err = om.rds.FindRingMined(event.TxHash.Hex()); err != nil {
		log.Debugf("order manager,handle ringmined event,tx:%s, ringhash:%s inserted", event.TxHash.Hex(), event.Ringhash.Hex())
		model.ConvertDown(event)
		return om.rds.Add(model)
	}

	if model.Status != uint8(event.Status) {
		log.Debugf("order manager,handle ringmined event,tx:%s, ringhash:%s updated", event.TxHash.Hex(), event.Ringhash.Hex())
		model.ConvertDown(event)
		return om.rds.Save(model)
	}

	return fmt.Errorf("order manager,handle ringmined event,tx:%s ringhash:%s already exist", event.TxHash.Hex(), event.Ringhash.Hex())
}

func (om *OrderManagerImpl) saveFillEvent(event *types.OrderFilledEvent) error {
	return nil
}

func (om *OrderManagerImpl) saveCancelEvent(event *types.OrderCancelledEvent) error {
	var (
		model dao.CancelEvent
		err   error
	)

	// save cancel event
	if model, err = om.rds.GetCancelEvent(event.TxHash); err != nil {
		log.Debugf("order manager,handle order cancelled event tx:%s, orderhash:%s inserted", event.TxHash.Hex(), event.OrderHash.Hex())
		model.ConvertDown(event)
		model.Fork = false
		return om.rds.Add(model)
	}

	if model.Status != uint8(event.Status) {
		log.Debugf("order manager,handle order cancelled event tx:%s, orderhash:%s updated", event.TxHash.Hex(), event.OrderHash.Hex())
		model.ConvertDown(event)
		return om.rds.Save(model)
	}

	return fmt.Errorf("order manager,handle order cancelled event tx:%s, orderhash:%s already exist", event.TxHash.Hex(), event.OrderHash.Hex())
}

func (om *OrderManagerImpl) saveCutoffEvent(event *types.CutoffEvent) error {
	var (
		model         dao.CutOffEvent
		orderhashlist []common.Hash
		err           error
	)

	if model, err = om.rds.GetCutoffEvent(event.TxHash); err == nil && model.Status == uint8(event.Status) {
		return fmt.Errorf("order manager,handle order cutoff event tx:%s already exist", event.TxHash.Hex())
	}

	if err == nil && model.Status != uint8(event.Status) {
		log.Debugf("order manager,handle cutoff event tx:%s, owner:%s, cutoffTimestamp:%s updated", event.TxHash.Hex(), event.Owner.Hex(), event.Cutoff.String())

		model.ConvertDown(event)
		return om.rds.Save(&model)
	}

	log.Debugf("order manager,handle cutoff event tx:%s, owner:%s, cutoffTimestamp:%s inserted", event.TxHash.Hex(), event.Owner.Hex(), event.Cutoff.String())

	orders, _ := om.rds.GetCutoffOrders(event.Owner, event.Cutoff)
	for _, v := range orders {
		orderhashlist = append(orderhashlist, common.HexToHash(v.OrderHash))
		event.OrderHashList = orderhashlist
	}

	model.ConvertDown(event)
	model.Fork = false
	return om.rds.Add(&model)
}

func (om *OrderManagerImpl) saveCutoffPairEvent(event *types.CutoffPairEvent) error {
	var (
		model         dao.CutOffPairEvent
		orderHashList []common.Hash
		err           error
	)
	if model, err = om.rds.GetCutoffPairEvent(event.TxHash); err == nil && model.Status == uint8(event.Status) {
		return fmt.Errorf("order manager,handle order cutoffPair event tx:%s already exist", event.TxHash.Hex())
	}

	if err == nil && model.Status != uint8(event.Status) {
		log.Debugf("order manager,handle cutoffPair event tx:%s, owner:%s, token1:%s, token2:%s, cutoffTimestamp:%s updated", event.TxHash.Hex(), event.Owner.Hex(), event.Token1.Hex(), event.Token2.Hex(), event.Cutoff.String())
		model.ConvertDown(event)
		return om.rds.Save(&model)
	}

	orders, _ := om.rds.GetCutoffPairOrders(event.Owner, event.Token1, event.Token2, event.Cutoff)
	for _, v := range orders {
		orderHashList = append(orderHashList, common.HexToHash(v.OrderHash))
		event.OrderHashList = orderHashList
	}
	log.Debugf("order manager,handle cutoffPair event tx:%s, owner:%s, token1:%s, token2:%s, cutoffTimestamp:%s", event.TxHash.Hex(), event.Owner.Hex(), event.Token1.Hex(), event.Token2.Hex(), event.Cutoff.String())

	model.ConvertDown(event)
	model.Fork = false
	return om.rds.Add(&model)
}
