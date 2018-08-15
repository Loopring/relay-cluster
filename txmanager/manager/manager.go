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
	"github.com/Loopring/relay-cluster/accountmanager"
	"github.com/Loopring/relay-cluster/dao"
	"github.com/Loopring/relay-cluster/txmanager/cache"
	txtyp "github.com/Loopring/relay-cluster/txmanager/types"
	notify "github.com/Loopring/relay-cluster/util"
	"github.com/Loopring/relay-lib/eth/contract"
	"github.com/Loopring/relay-lib/eventemitter"
	"github.com/Loopring/relay-lib/log"
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type TransactionManager struct {
	db                              *dao.RdsService
	approveEventWatcher             *eventemitter.Watcher
	orderCancelledEventWatcher      *eventemitter.Watcher
	cutoffAllEventWatcher           *eventemitter.Watcher
	cutoffPairEventWatcher          *eventemitter.Watcher
	wethDepositEventWatcher         *eventemitter.Watcher
	wethWithdrawalEventWatcher      *eventemitter.Watcher
	transferEventWatcher            *eventemitter.Watcher
	ethTransferEventWatcher         *eventemitter.Watcher
	unsupportedContractEventWatcher *eventemitter.Watcher
	orderFilledEventWatcher         *eventemitter.Watcher
	forkDetectedEventWatcher        *eventemitter.Watcher
}

func NewTxManager(db *dao.RdsService) TransactionManager {
	var tm TransactionManager
	tm.db = db

	if cache.Invalid() {
		cache.Initialize(db)
	}

	return tm
}

// Start start orderbook as a service
func (tm *TransactionManager) Start() {
	log.Debugf("transaction manager start...")

	tm.approveEventWatcher = &eventemitter.Watcher{Concurrent: false, Handle: tm.SaveApproveEvent}
	eventemitter.On(eventemitter.Approve, tm.approveEventWatcher)

	tm.orderCancelledEventWatcher = &eventemitter.Watcher{Concurrent: false, Handle: tm.SaveOrderCancelledEvent}
	eventemitter.On(eventemitter.CancelOrder, tm.orderCancelledEventWatcher)

	tm.cutoffAllEventWatcher = &eventemitter.Watcher{Concurrent: false, Handle: tm.SaveCutoffAllEvent}
	eventemitter.On(eventemitter.CutoffAll, tm.cutoffAllEventWatcher)

	tm.cutoffPairEventWatcher = &eventemitter.Watcher{Concurrent: false, Handle: tm.SaveCutoffPairEvent}
	eventemitter.On(eventemitter.CutoffPair, tm.cutoffPairEventWatcher)

	tm.wethDepositEventWatcher = &eventemitter.Watcher{Concurrent: false, Handle: tm.SaveWethDepositEvent}
	eventemitter.On(eventemitter.WethDeposit, tm.wethDepositEventWatcher)

	tm.wethWithdrawalEventWatcher = &eventemitter.Watcher{Concurrent: false, Handle: tm.SaveWethWithdrawalEvent}
	eventemitter.On(eventemitter.WethWithdrawal, tm.wethWithdrawalEventWatcher)

	tm.transferEventWatcher = &eventemitter.Watcher{Concurrent: false, Handle: tm.SaveTransferEvent}
	eventemitter.On(eventemitter.Transfer, tm.transferEventWatcher)

	tm.ethTransferEventWatcher = &eventemitter.Watcher{Concurrent: false, Handle: tm.SaveEthTransferEvent}
	eventemitter.On(eventemitter.EthTransfer, tm.ethTransferEventWatcher)

	tm.unsupportedContractEventWatcher = &eventemitter.Watcher{Concurrent: false, Handle: tm.SaveUnSupportedContractEvent}
	eventemitter.On(eventemitter.UnsupportedContract, tm.unsupportedContractEventWatcher)

	tm.orderFilledEventWatcher = &eventemitter.Watcher{Concurrent: false, Handle: tm.SaveOrderFilledEvent}
	eventemitter.On(eventemitter.OrderFilled, tm.orderFilledEventWatcher)

	tm.forkDetectedEventWatcher = &eventemitter.Watcher{Concurrent: false, Handle: tm.ForkProcess}
	eventemitter.On(eventemitter.ChainForkDetected, tm.forkDetectedEventWatcher)
}

func (tm *TransactionManager) Stop() {
	eventemitter.Un(eventemitter.Approve, tm.approveEventWatcher)
	eventemitter.Un(eventemitter.CancelOrder, tm.orderCancelledEventWatcher)
	eventemitter.Un(eventemitter.CutoffAll, tm.cutoffAllEventWatcher)
	eventemitter.Un(eventemitter.CutoffPair, tm.cutoffPairEventWatcher)
	eventemitter.Un(eventemitter.WethDeposit, tm.wethDepositEventWatcher)
	eventemitter.Un(eventemitter.WethWithdrawal, tm.wethWithdrawalEventWatcher)
	eventemitter.Un(eventemitter.Transfer, tm.transferEventWatcher)
	eventemitter.Un(eventemitter.EthTransfer, tm.ethTransferEventWatcher)
	eventemitter.Un(eventemitter.UnsupportedContract, tm.unsupportedContractEventWatcher)
	eventemitter.Un(eventemitter.OrderFilled, tm.orderFilledEventWatcher)
	eventemitter.Un(eventemitter.ChainForkDetected, tm.forkDetectedEventWatcher)
}

// todo: check and test
func (tm *TransactionManager) ForkProcess(input eventemitter.EventData) error {
	log.Debugf("txmanager,processing chain fork......")

	tm.Stop()
	forkEvent := input.(*types.ForkedEvent)
	from := forkEvent.ForkBlock.Int64()
	to := forkEvent.DetectedBlock.Int64()
	if err := tm.db.RollBackTxEntity(from, to); err != nil {
		log.Debugf("txmanager,process fork error:%s", err.Error())
	}
	if owners, err := tm.db.RollBackTxView(from, to); err != nil {
		log.Debugf("txmanager,process fork error:%s", err.Error())
	} else if owners != nil && len(owners) > 0 {
		//cache.DelTxViewCacheByOwners(owners)
	}
	if err := cache.RollbackEntityCache(from, to); err != nil {
		log.Debugf("txmanager,process cache rollback error:%s", err.Error())
	}
	tm.Start()

	return nil
}

func (tm *TransactionManager) SaveApproveEvent(input eventemitter.EventData) error {
	event := input.(*types.ApprovalEvent)

	var (
		entity txtyp.TransactionEntity
		list   []txtyp.TransactionView
	)

	if err := entity.FromApproveEvent(event); err != nil {
		return err
	}

	view, err := txtyp.ApproveView(event)
	if err != nil {
		return err
	}
	list = append(list, view)

	return tm.saveTransaction(&entity, list)
}

func (tm *TransactionManager) SaveOrderCancelledEvent(input eventemitter.EventData) error {
	event := input.(*types.OrderCancelledEvent)

	var (
		entity txtyp.TransactionEntity
		list   []txtyp.TransactionView
	)

	if err := entity.FromCancelEvent(event); err != nil {
		return err
	}

	view, err := txtyp.CancelView(event)
	if err != nil {
		return err
	}

	list = append(list, view)
	return tm.saveTransaction(&entity, list)
}

func (tm *TransactionManager) SaveCutoffAllEvent(input eventemitter.EventData) error {
	event := input.(*types.CutoffEvent)

	var (
		entity txtyp.TransactionEntity
		list   []txtyp.TransactionView
	)

	if err := entity.FromCutoffEvent(event); err != nil {
		return err
	}
	view, err := txtyp.CutoffView(event)
	if err != nil {
		return err
	}
	list = append(list, view)

	return tm.saveTransaction(&entity, list)
}

func (tm *TransactionManager) SaveCutoffPairEvent(input eventemitter.EventData) error {
	event := input.(*types.CutoffPairEvent)

	var (
		entity txtyp.TransactionEntity
		list   []txtyp.TransactionView
	)

	if err := entity.FromCutoffPairEvent(event); err != nil {
		return err
	}
	view, err := txtyp.CutoffPairView(event)
	if err != nil {
		return err
	}
	list = append(list, view)

	return tm.saveTransaction(&entity, list)
}

func (tm *TransactionManager) SaveWethDepositEvent(input eventemitter.EventData) error {
	event := input.(*types.WethDepositEvent)

	var entity txtyp.TransactionEntity
	if err := entity.FromWethDepositEvent(event); err != nil {
		return err
	}

	list, err := txtyp.WethDepositView(event)
	if err != nil {
		return err
	}

	return tm.saveTransaction(&entity, list)
}

func (tm *TransactionManager) SaveWethWithdrawalEvent(input eventemitter.EventData) error {
	event := input.(*types.WethWithdrawalEvent)

	var entity txtyp.TransactionEntity
	if err := entity.FromWethWithdrawalEvent(event); err != nil {
		return err
	}
	list, err := txtyp.WethWithdrawalView(event)
	if err != nil {
		return err
	}

	return tm.saveTransaction(&entity, list)
}

func (tm *TransactionManager) SaveTransferEvent(input eventemitter.EventData) error {
	event := input.(*types.TransferEvent)

	var entity txtyp.TransactionEntity
	if err := entity.FromTransferEvent(event); err != nil {
		return err
	}
	list, err := txtyp.TransferView(event)
	if err != nil {
		return err
	}

	// view过滤fill owner的转账,entity仍然存储transfer
	if contract.TxIsSubmitRing(event.Identify) {
		var filterList []txtyp.TransactionView
		for _, v := range list {
			if ok, _ := cache.ExistFillOwnerCache(v.TxHash, v.Owner); !ok {
				filterList = append(filterList, v)
			}
		}
		list = filterList
	}

	return tm.saveTransaction(&entity, list)
}

// 普通的transaction
// 当value大于0时认为是eth转账
// 当value等于0时认为是调用系统不支持的合约,默认使用fromTransferEvent/send type为unsupported_contract
func (tm *TransactionManager) SaveEthTransferEvent(input eventemitter.EventData) error {
	event := input.(*types.EthTransferEvent)

	var entity txtyp.TransactionEntity
	if err := entity.FromEthTransferEvent(event); err != nil {
		return err
	}
	list, err := txtyp.EthTransferView(event)
	if err != nil {
		return err
	}
	return tm.saveTransaction(&entity, list)
}

func (tm *TransactionManager) SaveUnSupportedContractEvent(input eventemitter.EventData) error {
	event := input.(*types.UnsupportedContractEvent)

	var entity txtyp.TransactionEntity
	if err := entity.FromUnsupportedContractEvent(event); err != nil {
		return err
	}
	list, err := txtyp.UnsupportedContractView(event)
	if err != nil {
		return err
	}
	return tm.saveTransaction(&entity, list)
}

func (tm *TransactionManager) SaveOrderFilledEvent(input eventemitter.EventData) error {
	event := input.(*types.OrderFilledEvent)

	// 存储fill关联的用户及tx
	cache.SetFillOwnerCache(event.TxHash, event.Owner)

	// 一个ringmined可以生成多个fill,他们的tx&logIndex都相等,这里将其放大存储到entity及view
	event.TxLogIndex = event.TxLogIndex*10 + event.FillIndex.Int64()

	var entity txtyp.TransactionEntity
	if err := entity.FromOrderFilledEvent(event); err != nil {
		return err
	}
	list, err := txtyp.OrderFilledView(event)
	if err != nil {
		return err
	}

	return tm.saveTransaction(&entity, list)
}

func (tm *TransactionManager) saveTransaction(tx *txtyp.TransactionEntity, list []txtyp.TransactionView) error {
	if tx.Status == types.TX_STATUS_PENDING {
		return tm.savePendingTx(tx, list)
	}
	return tm.saveMinedTx(tx, list)
}

func (tm *TransactionManager) savePendingTx(tx *txtyp.TransactionEntity, list []txtyp.TransactionView) error {
	// get users unlocked map
	ump := tm.getUnlockedMap(list)

	// save entity if either owner unlocked
	if !ump.invalidEntity() {
		return nil
	}

	// find pending tx entity, return if tx already exist
	if _, err := tm.db.FindPendingTxEntity(tx.Hash.Hex()); err == nil {
		log.Debugf("transaction manager,tx pending entity:%s already exist", tx.Hash.Hex())
		return nil
	}

	// save entity
	if err := tm.addEntity(tx); err != nil {
		log.Errorf("transaction manager,add tx pending entity:%s error:%s", tx.Hash.Hex(), err.Error())
		return err
	}

	// save views
	for _, view := range list {
		if !ump.invalidView(view.Owner) {
			continue
		}
		if err := tm.addView(&view); err != nil {
			log.Errorf("transaction manager,add tx pending view:%s owner:%s error:%s", tx.Hash.Hex(), err.Error())
		}
	}

	return nil
}

func (tm *TransactionManager) saveMinedTx(tx *txtyp.TransactionEntity, list []txtyp.TransactionView) error {
	// get users unlocked map
	ump := tm.getUnlockedMap(list)
	if !ump.invalidEntity() {
		return nil
	}

	// process pending txs
	tm.processPendingTxWhileMined(tx)

	// save entity
	if model, err := tm.db.FindTxEntity(tx.Hash.Hex(), tx.LogIndex); err == nil {
		// 判断是否为同一个txhash
		if model.Nonce == tx.Nonce.Int64() {
			log.Debugf("transaction manager,tx mined entity:%s logIndex:%d already exist", tx.Hash.Hex(), tx.LogIndex)
			return nil
		} else {
			tm.db.DelDuplicateTxEntity(model.TxHash, model.LogIndex, model.Nonce)
		}
	}
	if err := tm.addEntity(tx); err != nil {
		log.Errorf("transaction manager,tx mined entity:%s error:%s", tx.Hash.Hex(), err.Error())
		return err
	}

	for _, view := range list {
		if !ump.invalidView(view.Owner) {
			continue
		}
		if err := tm.addView(&view); err != nil {
			log.Errorf("transaction manager,add tx mined view:%s error:%s", tx.Hash.Hex(), err.Error())
		}
		log.Debugf("transaction manager,tx mined view:%s type:%s owner:%s logIndex:%d status:%s", view.TxHash.Hex(), txtyp.TypeStr(view.Type), view.Owner.Hex(), view.LogIndex, types.StatusStr(view.Status))
	}

	return nil
}

// todo(fuk): redo it as cron task
func (tm *TransactionManager) processPendingTxWhileMined(tx *txtyp.TransactionEntity) {
	// find the same nonce pending txs and delete
	txs, _ := tm.db.GetPendingTxEntity(tx.From.Hex(), tx.Nonce.Int64())
	if len(txs) == 0 {
		return
	}

	var (
		preHashList          []string
		currentHashIsPending bool = false
	)

	for _, v := range txs {
		if common.HexToHash(v.TxHash) != tx.Hash {
			preHashList = append(preHashList, v.TxHash)
		} else {
			currentHashIsPending = true
		}
	}

	// 将相同nonce的其他hash更新为failed
	if len(preHashList) > 0 {
		if err := tm.db.SetPendingTxEntityFailed(preHashList); err != nil {
			log.Errorf("transaction manager,set pending tx entities:%s err:", err.Error())
		}
		if owners, err := tm.db.SetPendingTxViewFailed(preHashList); err != nil {
			log.Errorf("transaction manager,set pending tx view:%s err:", err.Error())
		} else if owners != nil && len(owners) > 0 {
			cache.DelTxViewCacheByOwners(owners)
		}
	}

	// 删除当前pending tx
	if currentHashIsPending {
		if err := tm.db.DelPendingTxEntity(tx.Hash.Hex()); err != nil {
			log.Errorf("transaction manager,delete pending tx entity:%s err:", tx.Hash.Hex(), err.Error())
		}
		if owners, err := tm.db.DelPendingTxView(tx.Hash.Hex()); err != nil {
			log.Errorf("transaction manager,delete pending tx view:%s err:", tx.Hash.Hex(), err.Error())
		} else if owners != nil && len(owners) > 0 {
			cache.DelTxViewCacheByOwners(owners)
		}
	}
}

func (tm *TransactionManager) addEntity(tx *txtyp.TransactionEntity) error {
	var item dao.TransactionEntity
	item.ConvertDown(tx)
	setNonce(tx.Status, tx.From, tx.Nonce)
	return tm.db.Add(&item)
}

func (tm *TransactionManager) addView(tx *txtyp.TransactionView) error {
	var item dao.TransactionView

	item.ConvertDown(tx)
	if err := tm.db.Add(&item); err != nil {
		return err
	}

	notify.NotifyTransactionView(tx)

	cache.DelTxViewCacheByOwners([]string{item.Owner})

	return nil
}

type unlockedMap map[common.Address]bool

func (tm *TransactionManager) getUnlockedMap(list []txtyp.TransactionView) unlockedMap {
	ret := make(map[common.Address]bool)

	for _, v := range list {
		if ok, _ := accountmanager.HasUnlocked(v.Owner.Hex()); ok {
			ret[v.Owner] = true
		} else {
			ret[v.Owner] = false
		}
	}

	return ret
}

func (m unlockedMap) invalidEntity() bool {
	ret := false
	for _, unlocked := range m {
		if unlocked {
			ret = true
			break
		}
	}
	return ret
}

func (m unlockedMap) invalidView(owner common.Address) bool {
	if unlocked, ok := m[owner]; ok && unlocked {
		return true
	}
	return false
}

func setNonce(status types.TxStatus, owner common.Address, currentNonce *big.Int) error {
	// mined
	if status != types.TX_STATUS_PENDING {
		if preNonce, err := cache.GetTxMinedMaxNonceValue(owner); err != nil {
			return err
		} else {
			cache.SetTxMinedMaxNonceValue(owner, preNonce, currentNonce)
		}
	}

	// pending/success/failed
	if preNonce, err := cache.GetMaxNonceValue(owner); err != nil {
		return err
	} else {
		return cache.SetMaxNonceValue(owner, preNonce, currentNonce)
	}
}
