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

package accessor

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Loopring/relay-lib/cache"
	"github.com/Loopring/relay-lib/crypto"
	"github.com/Loopring/relay-lib/eth/abi"
	relayethtyp "github.com/Loopring/relay-lib/eth/types"
	"github.com/Loopring/relay-lib/log"
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	"math/big"
	"time"
)

func (accessor *ethNodeAccessor) RetryCall(routeParam string, retry int, result interface{}, method string, args ...interface{}) error {
	var err error
	for i := 0; i < retry; i++ {
		if _, err = accessor.Call(routeParam, result, method, args...); nil != err {
			continue
		} else {
			return nil
		}
	}
	return err
}

func (accessor *ethNodeAccessor) BatchCall(routeParam string, reqElems []rpc.BatchElem) ([]rpc.BatchElem, error) {
	if _, err := accessor.MutilClient.BatchCall(routeParam, reqElems); err != nil {
		return reqElems, err
	}

	return reqElems, nil
}

func (accessor *ethNodeAccessor) RetryBatchCall(routeParam string, reqElems []rpc.BatchElem, retry int) ([]rpc.BatchElem, error) {
	var err error
	for i := 0; i < retry; i++ {
		if _, err = accessor.BatchCall(routeParam, reqElems); err == nil {
			break
		}
		log.Debugf("accessor,RetryBatchCall %d'st time to get block's transactions", i+1)
	}
	return reqElems, err
}

func (accessor *ethNodeAccessor) BatchTransactions(routeParam string, retry int, reqs []*BatchTransactionReq) error {
	if len(reqs) < 1 || retry < 1 {
		return fmt.Errorf("ethaccessor:batchTransactions retry or reqs invalid")
	}

	reqElems := make([]rpc.BatchElem, len(reqs))
	for idx, req := range reqs {
		reqElems[idx] = rpc.BatchElem{
			Method: "eth_getTransactionByHash",
			Args:   []interface{}{req.TxHash},
			Result: &req.TxContent,
		}
	}

	if _, err := accessor.RetryBatchCall(routeParam, reqElems, retry); err != nil {
		return err
	}

	for idx, v := range reqElems {
		repeatCnt := 0
	mark:
		if repeatCnt > accessor.fetchTxRetryCount {
			err := fmt.Errorf("can't get content of this tx:%s", reqs[idx].TxHash)
			log.Errorf("accessor,BatchTransactions :%s", err.Error())
			return err
		}
		if v.Error == nil && v.Result != nil {
			if tx, ok := v.Result.(*relayethtyp.Transaction); ok && len(tx.Hash) > 0 {
				hash := common.HexToHash(tx.Hash)
				if !types.IsZeroHash(hash) {
					continue
				}
			}
		}

		// 第一次后如果没拿到则等待1s
		if repeatCnt > 0 {
			time.Sleep(1 * time.Second)
		}
		repeatCnt++

		log.Debugf("accessor,BatchTransactions %d's time to get Transaction:%s", repeatCnt, reqs[idx].TxHash)
		_, v.Error = accessor.Call(routeParam, &reqs[idx].TxContent, "eth_getTransactionByHash", reqs[idx].TxHash)
		goto mark
	}

	return nil
}

func (accessor *ethNodeAccessor) BatchTransactionRecipients(routeParam string, retry int, reqs []*BatchTransactionRecipientReq) error {
	if len(reqs) < 1 || retry < 1 {
		return fmt.Errorf("ethaccessor:batchTransactionRecipients retry or reqs invalid")
	}

	reqElems := make([]rpc.BatchElem, len(reqs))
	for idx, req := range reqs {
		reqElems[idx] = rpc.BatchElem{
			Method: "eth_getTransactionReceipt",
			Args:   []interface{}{req.TxHash},
			Result: &req.TxContent,
		}
	}

	if _, err := accessor.RetryBatchCall(routeParam, reqElems, retry); err != nil {
		return err
	}

	for idx, v := range reqElems {
		repeatCnt := 0
	mark:
		if repeatCnt > accessor.fetchTxRetryCount {
			err := fmt.Errorf("can't get receipt of this tx:%s", reqs[idx].TxHash)
			log.Errorf("accessor,BatchTransactions :%s", err.Error())
			return err
		}
		if v.Error == nil && v.Result != nil && !reqs[idx].TxContent.StatusInvalid() {
			if tx, ok := v.Result.(*relayethtyp.TransactionReceipt); ok && len(tx.TransactionHash) > 0 {
				hash := common.HexToHash(tx.TransactionHash)
				if !types.IsZeroHash(hash) {
					continue
				}
			}
		}

		if repeatCnt > 0 {
			time.Sleep(1 * time.Second)
		}
		repeatCnt++

		log.Debugf("accessor,BatchTransactions %d's time to get TransactionReceipt:%s and statusInvalid:%t", repeatCnt, reqs[idx].TxHash, reqs[idx].TxContent.StatusInvalid())
		_, v.Error = accessor.Call(routeParam, &reqs[idx].TxContent, "eth_getTransactionReceipt", reqs[idx].TxHash)
		goto mark
	}

	return nil
}

func (accessor *ethNodeAccessor) EstimateGas(routeParam string, callData []byte, to common.Address) (gas, gasPrice *big.Int, err error) {
	var gasBig, gasPriceBig types.Big
	//if nil == accessor.gasPriceEvaluator.gasPrice || accessor.gasPriceEvaluator.gasPrice.Cmp(big.NewInt(int64(0))) <= 0 {
	//	if err = accessor.RetryCall(routeParam, 2, &gasPriceBig, "eth_gasPrice"); nil != err {
	//		return
	//	}
	//} else {
	//	gasPriceBig = new(types.Big).SetInt(accessor.gasPriceEvaluator.gasPrice)
	//}

	callArg := &relayethtyp.CallArg{}
	callArg.To = to
	callArg.Data = common.ToHex(callData)
	callArg.GasPrice = gasPriceBig
	log.Debugf("EstimateGas gasPrice:%s", gasPriceBig.BigInt().String())
	if err = accessor.RetryCall(routeParam, 2, &gasBig, "eth_estimateGas", callArg); nil != err {
		return
	}
	log.Debugf("EstimateGas finished")
	gasPrice = gasPriceBig.BigInt()
	gas = gasBig.BigInt()
	return
}

func (accessor *ethNodeAccessor) ContractCallMethod(a *abi.ABI, contractAddress common.Address) func(result interface{}, methodName, blockParameter string, args ...interface{}) error {
	return func(result interface{}, methodName string, blockParameter string, args ...interface{}) error {
		if callData, err := a.Pack(methodName, args...); nil != err {
			return err
		} else {
			arg := &relayethtyp.CallArg{}
			arg.From = contractAddress
			arg.To = contractAddress
			arg.Data = common.ToHex(callData)
			return accessor.RetryCall(blockParameter, 2, result, "eth_call", arg, blockParameter)
		}
	}
}

func (ethAccessor *ethNodeAccessor) SignAndSendTransaction(result interface{}, sender common.Address, tx *ethTypes.Transaction) error {
	var err error
	if tx, err = crypto.SignTx(sender, tx, nil); nil != err {
		return err
	}
	if txData, err := rlp.EncodeToBytes(tx); nil != err {
		return err
	} else {
		log.Debugf("txhash:%s, nonce:%d, value:%s, gas:%d, gasPrice:%s", tx.Hash().Hex(), tx.Nonce(), tx.Value().String(), tx.Gas(), tx.GasPrice().String())
		err = ethAccessor.RetryCall("latest", 2, result, "eth_sendRawTransaction", common.ToHex(txData))
		if err != nil {
			log.Errorf("accessor, Sign and send transaction error:%s", err.Error())
		}
		return err
	}
}

func (accessor *ethNodeAccessor) ContractSendTransactionByData(routeParam string, sender common.Address, to common.Address, gas, gasPrice, value *big.Int, callData []byte, needPreExe bool) (string, *ethTypes.Transaction, error) {
	if nil == gasPrice || gasPrice.Cmp(big.NewInt(0)) <= 0 {
		return "", nil, errors.New("gasPrice must be setted.")
	}
	if nil == gas || gas.Cmp(big.NewInt(0)) <= 0 {
		return "", nil, errors.New("gas must be setted.")
	}
	var txHash string
	if needPreExe {
		if estimagetGas, _, err := EstimateGas(callData, to, "latest"); nil != err {
			return txHash, nil, err
		} else {
			gas = estimagetGas
		}
	}
	nonce := accessor.addressCurrentNonce(sender)
	log.Infof("nonce:%s, gas:%s", nonce.String(), gas.String())
	if value == nil {
		value = big.NewInt(0)
	}
	//todo:modify it
	//if gas.Cmp(big.NewInt(int64(350000)))  {
	gas.SetString("500000", 0)
	//}
	transaction := ethTypes.NewTransaction(nonce.Uint64(),
		common.HexToAddress(to.Hex()),
		value,
		gas.Uint64(),
		gasPrice,
		callData)
	if err := accessor.SignAndSendTransaction(&txHash, sender, transaction); nil != err {
		//if err.Error() == "nonce too low" {
		accessor.resetAddressNonce(sender)
		nonce = accessor.addressCurrentNonce(sender)
		transaction = ethTypes.NewTransaction(nonce.Uint64(),
			common.HexToAddress(to.Hex()),
			value,
			gas.Uint64(),
			gasPrice,
			callData)
		if err := accessor.SignAndSendTransaction(&txHash, sender, transaction); nil != err {
			log.Errorf("send raw transaction err:%s, manual check it please.", err.Error())
			return "", nil, err
		}
		//} else {
		//
		//}
	}
	accessor.addressNextNonce(sender)
	return txHash, transaction, nil
}

//gas, gasPrice can be set to nil
func (accessor *ethNodeAccessor) ContractSendTransactionMethod(routeParam string, a *abi.ABI, contractAddress common.Address) func(sender common.Address, methodName string, gas, gasPrice, value *big.Int, args ...interface{}) (string, *ethTypes.Transaction, error) {
	return func(sender common.Address, methodName string, gas, gasPrice, value *big.Int, args ...interface{}) (string, *ethTypes.Transaction, error) {
		if callData, err := a.Pack(methodName, args...); nil != err {
			return "", nil, err
		} else {
			if nil == gas || nil == gasPrice {
				if gas, gasPrice, err = accessor.EstimateGas(routeParam, callData, contractAddress); nil != err {
					return "", nil, err
				}
			}
			gas.Add(gas, big.NewInt(int64(1000)))
			log.Infof("sender:%s, %s", sender.Hex(), gasPrice.String())
			return accessor.ContractSendTransactionByData(routeParam, sender, contractAddress, gas, gasPrice, value, callData, false)
		}
	}
}

func (iterator *BlockIterator) Next() (interface{}, error) {
	if nil != iterator.endNumber && iterator.endNumber.Cmp(big.NewInt(0)) > 0 && iterator.endNumber.Cmp(iterator.currentNumber) < 0 {
		return nil, errors.New("finished")
	}

	var blockNumber types.Big
	if err := iterator.ethClient.RetryCall("latest", 2, &blockNumber, "eth_blockNumber"); nil != err {
		log.Errorf("err:%s", err.Error())
		return nil, err
	} else {
		confirmNumber := iterator.currentNumber.Uint64() + iterator.confirms
		if blockNumber.Uint64() < confirmNumber {
		hasNext:
			for {
				select {
				// todo(fk):modify this duration
				case <-time.After(time.Duration(5 * time.Second)):
					if err1 := iterator.ethClient.RetryCall("latest", 2, &blockNumber, "eth_blockNumber"); nil == err1 && blockNumber.Uint64() >= confirmNumber {
						break hasNext
					}
				}
			}
		}
	}

	block, err := iterator.ethClient.GetFullBlock(iterator.currentNumber, iterator.withTxData)
	if nil == err {
		iterator.currentNumber.Add(iterator.currentNumber, big.NewInt(1))
	}
	return block, err
}

func (accessor *ethNodeAccessor) getFullBlockFromCacheByHash(hash string) (*relayethtyp.BlockWithTxAndReceipt, error) {
	blockWithTxAndReceipt := &relayethtyp.BlockWithTxAndReceipt{}

	if blockData, err := cache.Get(hash); nil == err {
		if err = json.Unmarshal(blockData, blockWithTxAndReceipt); nil == err {
			return blockWithTxAndReceipt, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (accessor *ethNodeAccessor) GetFullBlock(blockNumber *big.Int, withTxObject bool) (interface{}, error) {
	blockWithTxHash := &relayethtyp.BlockWithTxHash{}

	if err := accessor.RetryCall(blockNumber.String(), 2, &blockWithTxHash, "eth_getBlockByNumber", fmt.Sprintf("%#x", blockNumber), false); nil != err || blockWithTxHash == nil {
		blockNumberStr := "0"
		if nil != blockNumber {
			blockNumberStr = blockNumber.String()
		}
		if blockWithTxHash == nil {
			return nil, fmt.Errorf("err:%s, blockNumber:%s", "can't get blockWithTxHash by ", blockNumberStr)
		}
		log.Errorf("err:%s, blockNumber:%s", err.Error(), blockNumberStr)
		return nil, err
	} else {
		if !withTxObject {
			return blockWithTxHash, nil
		} else {
			if blockWithTxAndReceipt, err := accessor.getFullBlockFromCacheByHash(blockWithTxHash.Hash.Hex()); nil == err && nil != blockWithTxAndReceipt {
				return blockWithTxAndReceipt, nil
			} else {
				blockWithTxAndReceipt := &relayethtyp.BlockWithTxAndReceipt{}
				blockWithTxAndReceipt.Block = blockWithTxHash.Block
				blockWithTxAndReceipt.Transactions = []relayethtyp.Transaction{}
				blockWithTxAndReceipt.Receipts = []relayethtyp.TransactionReceipt{}

				txno := len(blockWithTxHash.Transactions)
				if txno == 0 {
					return blockWithTxAndReceipt, nil
				}

				var (
					txReqs = make([]*BatchTransactionReq, txno)
					rcReqs = make([]*BatchTransactionRecipientReq, txno)
				)
				for idx, txstr := range blockWithTxHash.Transactions {
					var (
						txreq        BatchTransactionReq
						rcreq        BatchTransactionRecipientReq
						tx           relayethtyp.Transaction
						rc           relayethtyp.TransactionReceipt
						txerr, rcerr error
					)
					txreq.TxHash = txstr
					txreq.TxContent = tx
					txreq.Err = txerr

					rcreq.TxHash = txstr
					rcreq.TxContent = rc
					rcreq.Err = rcerr

					txReqs[idx] = &txreq
					rcReqs[idx] = &rcreq
				}

				if err := BatchTransactions(txReqs, blockWithTxAndReceipt.Number.BigInt().String()); err != nil {
					log.Errorf("err:%s, blockNumber:%s", err.Error(), blockWithTxAndReceipt.Number.BigInt().String())
					return nil, err
				}
				if err := BatchTransactionRecipients(rcReqs, blockWithTxAndReceipt.Number.BigInt().String()); err != nil {
					log.Errorf("err:%s, blockNumber:%s", err.Error(), blockWithTxAndReceipt.Number.BigInt().String())
					return nil, err
				}

				for idx, _ := range txReqs {
					blockWithTxAndReceipt.Transactions = append(blockWithTxAndReceipt.Transactions, txReqs[idx].TxContent)
					blockWithTxAndReceipt.Receipts = append(blockWithTxAndReceipt.Receipts, rcReqs[idx].TxContent)
				}

				var txcnt types.Big
				if err := accessor.RetryCall("latest", 2, &txcnt, "eth_getBlockTransactionCountByHash", blockWithTxAndReceipt.Hash.Hex()); err != nil {
					return blockWithTxAndReceipt, err
				}
				txcntinblock := len(blockWithTxAndReceipt.Transactions)
				if txcntinblock != txcnt.Int() || txcntinblock != len(blockWithTxAndReceipt.Receipts) {
					err := fmt.Errorf("tx count isn't equal,txcount in chain:%d, txcount in block:%d, receipt count:%d", txcnt.Int(), txcntinblock, len(blockWithTxAndReceipt.Receipts))
					log.Errorf("err:%s", err.Error())
					return blockWithTxAndReceipt, err
				}

				if blockData, err := json.Marshal(blockWithTxAndReceipt); nil == err {
					cache.Set(blockWithTxHash.Hash.Hex(), blockData, int64(36000))
				}

				return blockWithTxAndReceipt, nil
			}

		}
	}
}

func (iterator *BlockIterator) Prev() (interface{}, error) {
	var block interface{}
	if iterator.withTxData {
		block = &relayethtyp.BlockWithTxObject{}
	} else {
		block = &relayethtyp.BlockWithTxHash{}
	}
	if nil != iterator.startNumber && iterator.startNumber.Cmp(big.NewInt(0)) > 0 && iterator.startNumber.Cmp(iterator.currentNumber) > 0 {
		return nil, errors.New("finished")
	}
	prevNumber := new(big.Int).Sub(iterator.currentNumber, big.NewInt(1))
	if err := iterator.ethClient.RetryCall(prevNumber.String(), 2, &block, "eth_getBlockByNumber", fmt.Sprintf("%#x", prevNumber), iterator.withTxData); nil != err {
		return nil, err
	} else {
		if nil == block {
			return nil, errors.New("there isn't a block with number:" + prevNumber.String())
		}
		iterator.currentNumber.Sub(iterator.currentNumber, big.NewInt(1))
		return block, nil
	}
}

func (ethAccessor *ethNodeAccessor) BlockIterator(startNumber, endNumber *big.Int, withTxData bool, confirms uint64) *BlockIterator {
	iterator := &BlockIterator{
		startNumber:   new(big.Int).Set(startNumber),
		endNumber:     endNumber,
		currentNumber: new(big.Int).Set(startNumber),
		ethClient:     ethAccessor,
		withTxData:    withTxData,
		confirms:      confirms,
	}
	return iterator
}

func (accessor *ethNodeAccessor) addressCurrentNonce(address common.Address) *big.Int {
	if _, exists := accessor.AddressNonce[address]; !exists {
		var nonce types.Big
		if err := accessor.RetryCall("pending", 2, &nonce, "eth_getTransactionCount", address.Hex(), "pending"); nil != err {
			nonce = *(types.NewBigWithInt(0))
		}
		accessor.AddressNonce[address] = nonce.BigInt()
	}
	nonce := new(big.Int)
	nonce.Set(accessor.AddressNonce[address])
	return nonce
}

func (accessor *ethNodeAccessor) resetAddressNonce(address common.Address) {
	var nonce types.Big
	if err := accessor.RetryCall("pending", 2, &nonce, "eth_getTransactionCount", address.Hex(), "pending"); nil != err {
		nonce = *(types.NewBigWithInt(0))
	}
	accessor.AddressNonce[address] = nonce.BigInt()
}

func (accessor *ethNodeAccessor) addressNextNonce(address common.Address) *big.Int {
	accessor.mtx.Lock()
	defer accessor.mtx.Unlock()

	nonce := accessor.addressCurrentNonce(address)
	nonce.Add(nonce, big.NewInt(int64(1)))
	accessor.AddressNonce[address].Set(nonce)
	return nonce
}
