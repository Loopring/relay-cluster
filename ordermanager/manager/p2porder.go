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
	"encoding/json"
	"github.com/Loopring/relay-lib/cache"
	"github.com/Loopring/relay-lib/eventemitter"
	"github.com/Loopring/relay-lib/log"
	"github.com/Loopring/relay-lib/types"
	"math/big"
	"strconv"
	"strings"
	"time"
)

const p2pRelationPreKey = "P2P_RELATION_"
const splitMark = "_"
const p2pTakerPreKey = "P2P_TAKERS_"

type P2pOrderRelation struct {
	Txhash         string
	Makerorderhash string
	Takerorderhash string
	PendingAmount  string
}

func init() {
	submitRingMethodWatcher := &eventemitter.Watcher{Concurrent: false, Handle: HandleP2PSubmitRing}
	eventemitter.On(eventemitter.Miner_SubmitRing_Method, submitRingMethodWatcher)

	p2pRingMinedWatcher := &eventemitter.Watcher{Concurrent: false, Handle: HandleP2POrderFilled}
	eventemitter.On(eventemitter.OrderFilled, p2pRingMinedWatcher)
}

func SaveP2POrderRelation(takerOwner, taker, makerOwner, maker, txHash, pendingAmount, validUntil string) error {
	log.Info("p2pOrder code line 53")
	takerOwner = strings.ToLower(takerOwner)
	taker = strings.ToLower(taker)
	makerOwner = strings.ToLower(makerOwner)
	maker = strings.ToLower(maker)
	txHash = strings.ToLower(txHash)

	untilTime, _ := strconv.ParseInt(validUntil, 10, 64)
	nowTime := time.Now().Unix()
	takerExpiredTime := untilTime - nowTime
	log.Info("p2pOrder code line 63")
	cache.ZAdd(p2pTakerPreKey+maker, takerExpiredTime, []byte(strconv.FormatInt(nowTime, 10)), []byte(txHash+splitMark+pendingAmount))
	log.Info("p2pOrder code line 65")
	//save txhash,maker,taker relations
	p2pOrderRelationStr, _ := GetP2pOrderRelation(maker, taker, txHash, pendingAmount)
	log.Info("p2pOrder code line 68:" + string(p2pOrderRelationStr))
	cache.Set(p2pRelationPreKey+txHash, p2pOrderRelationStr, takerExpiredTime)
	log.Info("p2pOrder code line 70")
	return nil
}

// status failed/pending
func HandleP2PSubmitRing(input eventemitter.EventData) error {
	//release taker's failed p2porder pengdingAmount
	if evt, ok := input.(*types.SubmitRingMethodEvent); ok && evt != nil && evt.Status == types.TX_STATUS_FAILED {
		txHash := strings.ToLower(evt.TxHash.Hex())
		jsonStr, _ := cache.Get(p2pRelationPreKey + txHash)
		p2pOrderRelation := P2pOrderRelation{}
		if err := json.Unmarshal(jsonStr, &p2pOrderRelation); nil != err {
			log.Errorf("mehtod HandleP2PSubmitRing of p2pOrderRelation syncFromCache err:%s", err.Error())
			return err
		} else {
			maker := p2pOrderRelation.Makerorderhash
			cache.ZRem(p2pTakerPreKey+maker, []byte(txHash+splitMark+p2pOrderRelation.PendingAmount))
		}
	}
	return nil
}

// status success
func HandleP2POrderFilled(input eventemitter.EventData) error {
	//release taker's successed p2porder pengdingAmount
	if evt, ok := input.(*types.OrderFilledEvent); ok && evt != nil && evt.Status == types.TX_STATUS_SUCCESS {
		txHash := strings.ToLower(evt.TxHash.Hex())
		jsonStr, _ := cache.Get(p2pRelationPreKey + txHash)
		p2pOrderRelation := P2pOrderRelation{}
		if err := json.Unmarshal(jsonStr, &p2pOrderRelation); nil != err {
			log.Errorf("metho of HandleP2POrderFilled p2pOrderRelation syncFromCache err:%s", err.Error())
			return err
		} else {
			maker := p2pOrderRelation.Makerorderhash
			cache.ZRem(p2pTakerPreKey+maker, []byte(txHash+splitMark+p2pOrderRelation.PendingAmount))
		}
	}
	return nil
}

func GetP2PPendingAmount(maker string) (pendingAmount *big.Rat, err error) {
	pendingAmount = new(big.Rat)
	maker = strings.ToLower(maker)
	if data, err := cache.ZRange(p2pTakerPreKey+maker, 0, -1, false); nil != err {
		return pendingAmount, err
	} else {
		for _, v := range data {
			pendData, _ := new(big.Int).SetString(strings.Split(string(v), splitMark)[1], 0)
			pendingAmount = pendingAmount.Add(pendingAmount, new(big.Rat).SetInt(pendData))
		}
		return pendingAmount, nil
	}
}

func GetP2pOrderRelation(maker, taker, txHash, pendingAmount string) ([]byte, error) {
	var p2pOrderRelation P2pOrderRelation
	p2pOrderRelation.Makerorderhash = maker
	p2pOrderRelation.Takerorderhash = taker
	p2pOrderRelation.Txhash = txHash
	p2pOrderRelation.PendingAmount = pendingAmount
	return json.Marshal(p2pOrderRelation)
}
