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
	"fmt"
	"github.com/Loopring/relay-lib/eth/abi"
	libEthTypes "github.com/Loopring/relay-lib/eth/types"
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"math/big"
	"sync"
)

var accessor *ethNodeAccessor

func BlockNumber(result interface{}) error {
	return accessor.RetryCall("latest", 5, result, "eth_blockNumber")
}

func GetBalance(result interface{}, address common.Address, blockNumber string) error {
	return accessor.RetryCall(blockNumber, 2, result, "eth_getBalance", address, blockNumber)
}

func SendRawTransaction(result interface{}, tx string) error {
	return accessor.RetryCall("latest", 2, result, "eth_sendRawTransaction", tx)
}

func GetTransactionCount(result interface{}, address common.Address, blockNumber string) error {
	return accessor.RetryCall(blockNumber, 2, result, "eth_getTransactionCount", address, blockNumber)
}

func Call(result interface{}, ethCall *libEthTypes.CallArg, blockNumber string) error {
	return accessor.RetryCall(blockNumber, 2, result, "eth_call", ethCall, blockNumber)
}

func GetBlockByNumber(result interface{}, blockNumber *big.Int, withObject bool) error {
	return accessor.RetryCall(blockNumber.String(), 2, result, "eth_getBlockByNumber", fmt.Sprintf("%#x", blockNumber), withObject)
}

func GetBlockByHash(result types.CheckNull, blockHash string, withObject bool) error {
	for _, c := range accessor.clients {
		//todo:is it need retrycall
		if err := c.client.Call(result, "eth_getBlockByHash", blockHash, withObject); nil == err {
			if !result.IsNull() {
				return nil
			}
		}
	}
	return fmt.Errorf("no block with blockhash:%s", blockHash)

	//return accessor.RetryCall("latest", 2, result, "eth_getBlockByHash", blockHash, withObject)
}

func GetTransactionReceipt(result interface{}, txHash string, blockParameter string) error {
	return accessor.RetryCall(blockParameter, 2, result, "eth_getTransactionReceipt", txHash)
}

func GetTransactionByHash(result types.CheckNull, txHash string, blockParameter string) error {
	for _, c := range accessor.clients {
		if err := c.client.Call(result, "eth_getTransactionByHash", txHash); nil == err {
			if !result.IsNull() {
				return nil
			}
		}
	}
	return fmt.Errorf("no transaction with hash:%s", txHash)
}

func GetBlockTransactionCountByHash(result interface{}, blockHash string, blockParameter string) error {
	return accessor.RetryCall("latest", 5, result, "eth_getBlockTransactionCountByHash", blockHash)

}

func GetBlockTransactionCountByNumber(result interface{}, blockNumber string) error {
	return accessor.RetryCall(blockNumber, 2, result, "eth_getBlockTransactionCountByNumber", blockNumber)
}

func EstimateGas(callData []byte, to common.Address, blockNumber string) (gas, gasPrice *big.Int, err error) {
	return accessor.EstimateGas(blockNumber, callData, to)
}

func SignAndSendTransaction(sender common.Address, to common.Address, gas, gasPrice, value *big.Int, callData []byte, needPreExe bool) (string, *ethTypes.Transaction, error) {
	return accessor.ContractSendTransactionByData("latest", sender, to, gas, gasPrice, value, callData, needPreExe)
}

func ContractSendTransactionMethod(routeParam string, a *abi.ABI, contractAddress common.Address) func(sender common.Address, methodName string, gas, gasPrice, value *big.Int, args ...interface{}) (string, *ethTypes.Transaction, error) {
	return accessor.ContractSendTransactionMethod(routeParam, a, contractAddress)
}

func ContractCallMethod(a *abi.ABI, contractAddress common.Address) func(result interface{}, methodName, blockParameter string, args ...interface{}) error {
	return accessor.ContractCallMethod(a, contractAddress)
}

func BatchCall(routeParam string, reqs []BatchReq) error {
	var err error
	elems := []rpc.BatchElem{}
	elemsLength := []int{}
	for _, req := range reqs {
		elems1 := req.ToBatchElem()
		elemsLength = append(elemsLength, len(elems1))
		elems = append(elems, elems1...)
	}
	if elems, err = accessor.BatchCall(routeParam, elems); nil != err {
		return err
	} else {
		startId := 0
		for idx, req := range reqs {
			endId := startId + elemsLength[idx]
			req.FromBatchElem(elems[startId:endId])
			startId = endId
		}
		return nil
	}
}

func BatchTransactions(reqs []*BatchTransactionReq, blockNumber string) error {
	return accessor.BatchTransactions(blockNumber, 5, reqs)
}

func BatchTransactionRecipients(reqs []*BatchTransactionRecipientReq, blockNumber string) error {
	return accessor.BatchTransactionRecipients(blockNumber, 5, reqs)
}

func NewBlockIterator(startNumber, endNumber *big.Int, withTxData bool, confirms uint64) *BlockIterator {
	return accessor.BlockIterator(startNumber, endNumber, withTxData, confirms)
}

func GetFullBlock(blockNumber *big.Int, withObject bool) (interface{}, error) {
	return accessor.GetFullBlock(blockNumber, withObject)
}

func IsInit() bool {
	return nil != accessor
}

func Initialize(accessorOptions AccessorOptions) error {
	var err error
	accessor = &ethNodeAccessor{}
	accessor.mtx = sync.RWMutex{}
	if accessorOptions.FetchTxRetryCount > 0 {
		accessor.fetchTxRetryCount = accessorOptions.FetchTxRetryCount
	} else {
		accessor.fetchTxRetryCount = 60
	}
	accessor.AddressNonce = make(map[common.Address]*big.Int)
	accessor.MutilClient = NewMutilClient(accessorOptions.RawUrls)
	if nil != err {
		return err
	}
	accessor.MutilClient.startSyncBlockNumber()
	return nil
}
