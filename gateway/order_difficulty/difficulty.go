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

package order_difficulty

import (
	"encoding/json"
	"github.com/Loopring/relay-lib/cache"
	"github.com/Loopring/relay-lib/eventemitter"
	"github.com/Loopring/relay-lib/log"
	"github.com/Loopring/relay-lib/types"
	"github.com/Loopring/relay-lib/zklock"
	"github.com/ethereum/go-ethereum/common"
	"gonum.org/v1/gonum/stat"
	"math/big"
	"strconv"
	"time"
)

const (
	OrderCountPerSecond = "o_cnt_per_s_"
	OrderDifficultyKey  = "order_diff"
	ZklockDifficulty    = "zklock_diff"
)

type OrderDifficultyEvaluator struct {
	//currentDifficult *OrderDifficulty
	//parentDifficult  *OrderDifficulty
	evaluator Evaluator
	stopFuns  []func()
	calCount  int64 //must be odd
	duration  int64
}

type OrderDifficulty struct {
	Difficulty string
	OrdersNum  int64
	TimeStamp  int64
}

func (evaluator *OrderDifficultyEvaluator) getOrderCacheKey(createTime int64) (key string, expireAt int64) {
	mod := createTime % evaluator.calCount
	orderSection := createTime - mod
	orderSectionStr := strconv.FormatInt(orderSection, 10)
	return OrderCountPerSecond + orderSectionStr, createTime + evaluator.calCount*evaluator.duration
}

func (evaluator *OrderDifficultyEvaluator) getDiffCacheKey(createTime int64) (key string, expireAt int64) {
	mod := createTime % evaluator.calCount
	orderSection := createTime - mod
	orderSectionStr := strconv.FormatInt(orderSection, 10)
	return OrderDifficultyKey + orderSectionStr, createTime + evaluator.calCount*evaluator.duration
}

func NewOrderDifficultyEvaluator(config OrderDifficultyConfig) *OrderDifficultyEvaluator {
	if "" == config.BaseDifficulty {
		config.BaseDifficulty = "0x10"
	}
	if 0 == config.CalCount {
		config.CalCount = 1000
	}
	if 0 == config.Duration {
		config.Duration = 2
	}
	if 0 == config.Threshold {
		config.Threshold = 10
	}
	evaluator := &OrderDifficultyEvaluator{calCount: config.CalCount, duration: config.Duration}
	baseDifficulty := types.HexToBigint(config.BaseDifficulty)
	evaluator.evaluator = &LinearEvaluator{baseDifficulty: baseDifficulty, threshold: config.Threshold}
	return evaluator
}

type OrderDifficultyConfig struct {
	BaseDifficulty string
	Threshold      int64
	CalCount       int64
	Duration       int64
}

func (evaluator *OrderDifficultyEvaluator) Start() {
	evaluator.HandleNewOrder()
	go func() {
		if err := zklock.TryLock(ZklockDifficulty); nil != err {
			log.Errorf("erro:%s", err.Error())
		} else {
			if _, err1 := GetDifficulty(); nil != err1 {
				diffHex := types.BigintToHex(evaluator.evaluator.getBaseDifficulty())
				cache.Set(OrderDifficultyKey, []byte(diffHex), int64(10000))
			}
			//now := time.Now().Unix()
			//orderCntList := []*OrderDifficulty{}
			//for i := evaluator.calCount; i > 0; i-- {
			//	t := now - i*evaluator.duration
			//	cacheKey, _ := evaluator.getDiffCacheKey(t)
			//	orderDifficulty := &OrderDifficulty{}
			//	if data, err := cache.Get(cacheKey); nil == err {
			//		json.Unmarshal(data, orderDifficulty)
			//	}
			//	orderCntList = append(orderCntList, orderDifficulty)
			//}
			for {
				select {
				case <-time.After(time.Duration(evaluator.duration) * time.Second):
					now := time.Now().Unix()
					if diffHex, err := GetDifficulty(); nil == err {
						currentDiff := diffHex
						orderDiff := &OrderDifficulty{}
						cacheKey, _ := evaluator.getOrderCacheKey(now)
						orderDiff.OrdersNum = 0
						if data, err1 := cache.Get(cacheKey); nil == err1 {
							if orderNum, err3 := strconv.Atoi(string(data)); nil == err3 {
								orderDiff.OrdersNum = int64(orderNum)
							}
						}
						diffCacheKey, _ := evaluator.getDiffCacheKey(now)
						orderDiff.Difficulty = currentDiff
						if data, err4 := json.Marshal(orderDiff); nil == err4 {
							cache.Set(diffCacheKey, data, (evaluator.calCount+2)*evaluator.duration)
						}
					}
					orderCntList := []*OrderDifficulty{}
					for i := evaluator.calCount; i > 0; i-- {
						t := now - i*evaluator.duration
						diffCacheKey, _ := evaluator.getDiffCacheKey(t)
						orderDifficulty := &OrderDifficulty{}
						if data, err := cache.Get(diffCacheKey); nil == err {
							json.Unmarshal(data, orderDifficulty)
						}
						orderCntList = append(orderCntList, orderDifficulty)
					}
					diff := evaluator.evaluator.CalcAndSaveDifficulty(orderCntList, evaluator.duration)
					diffHash := common.BytesToHash(diff.Bytes())
					log.Infof("current order difficulty:%s", diff.String())
					cache.Set(OrderDifficultyKey, []byte(diffHash.Hex()), int64(0))
				}
			}
		}
	}()
}

func (evaluator *OrderDifficultyEvaluator) Stop() {
	for _, f := range evaluator.stopFuns {
		f()
	}
}

//add OrdersNum
func (evaluator *OrderDifficultyEvaluator) HandleNewOrder() {
	watcher := &eventemitter.Watcher{
		Concurrent: false, Handle: func(input eventemitter.EventData) error {
			state := input.(*types.OrderState)
			cacheKey, expireAt := evaluator.getOrderCacheKey(state.RawOrder.CreateTime)
			_, err := cache.Incr(cacheKey)
			if nil == err {
				err = cache.ExpireAt(cacheKey, expireAt)
			}
			return err
		},
	}

	evaluator.stopFuns = append(evaluator.stopFuns, func() {
		eventemitter.Un(eventemitter.NewOrder, watcher)
	})
	eventemitter.On(eventemitter.NewOrder, watcher)
}

type Evaluator interface {
	CalcAndSaveDifficulty(orderCntList []*OrderDifficulty, duration int64) *big.Int
	getBaseDifficulty() *big.Int
}

type LinearEvaluator struct {
	baseDifficulty *big.Int
	threshold      int64
}

func (evaluator *LinearEvaluator) getBaseDifficulty() *big.Int {
	return new(big.Int).Set(evaluator.baseDifficulty)
}

func (evaluator *LinearEvaluator) CalNextOrderCnt(orderCntList []*OrderDifficulty, duration int64) int64 {
	xes := []float64{}
	yes := []float64{}
	now := time.Now().Unix()
	for idx, cnt := range orderCntList {
		//println("sum",cnt.OrdersNum)
		xes = append(xes, float64(now-duration*int64(len(orderCntList)-idx)))
		yes = append(yes, float64(cnt.OrdersNum))
	}
	alpha, beta := stat.LinearRegression(xes, yes, nil, false)
	return int64(beta*float64(now) + alpha)
}

//控制订单的提交速度，随着订单的流量增大而增大
func (evaluator *LinearEvaluator) CalcAndSaveDifficulty(orderCntList []*OrderDifficulty, duration int64) *big.Int {
	nextCnt := evaluator.CalNextOrderCnt(orderCntList, duration)
	log.Infof("next order count:%d", nextCnt)
	orderList := append(orderCntList, &OrderDifficulty{OrdersNum: nextCnt})
	return evaluator.nextDifficulty(orderList)
}

func NewOrderDiff(ordersNum int64, diff string) *OrderDifficulty {
	return &OrderDifficulty{OrdersNum: ordersNum, Difficulty: diff}
}

func NewLinearEvaluator(baseDifficulty *big.Int, threshold int64) *LinearEvaluator {
	return &LinearEvaluator{baseDifficulty: baseDifficulty, threshold: threshold}
}

func (evaluator *LinearEvaluator) nextDifficulty(orderCntList []*OrderDifficulty) *big.Int {
	nextOrderDiff := orderCntList[len(orderCntList)-1]
	if nextOrderDiff.OrdersNum < evaluator.threshold {
		return evaluator.baseDifficulty
	} else {
		currentDiff := new(big.Int).Set(evaluator.baseDifficulty)
		if currentDiffHex, err := GetDifficulty(); nil == err {
			currentDiff = types.HexToBigint(currentDiffHex)
		}
		addRatio := big.NewInt(3)
		if nextOrderDiff.OrdersNum > int64(float64(evaluator.threshold)*1.5) {
			addRatio = big.NewInt(1)
		} else if nextOrderDiff.OrdersNum < int64(float64(evaluator.threshold)*1.2) {
			addRatio = big.NewInt(5)
		}
		return new(big.Int).Add(currentDiff, new(big.Int).Quo(currentDiff, addRatio))
	}
}

func GetDifficulty() (string, error) {
	if data, err := cache.Get(OrderDifficultyKey); nil == err {
		return string(data), nil
	} else {
		return "0x0", err
	}
}
