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

package order_difficulty_test

import (
	"encoding/json"
	"github.com/Loopring/relay-cluster/gateway/order_difficulty"
	"github.com/Loopring/relay-lib/cache"
	"github.com/Loopring/relay-lib/cache/redis"
	"github.com/Loopring/relay-lib/log"
	"github.com/Loopring/relay-lib/types"
	"github.com/Loopring/relay-lib/zklock"
	"go.uber.org/zap"
	"math/big"
	"testing"
	"time"
)

func init() {
	logConfig := `{
	  "level": "debug",
	  "development": false,
	  "encoding": "json",
	  "outputPaths": ["stdout"],
	  "errorOutputPaths": ["stderr"],
	  "encoderConfig": {
	    "messageKey": "message",
	    "levelKey": "level",
	    "levelEncoder": "lowercase"
	  }
	}`
	rawJSON := []byte(logConfig)

	var (
		cfg zap.Config
		err error
	)
	if err = json.Unmarshal(rawJSON, &cfg); err != nil {
		panic(err)
	}
	log.Initialize(cfg)
	cache.NewCache(redis.RedisOptions{Host: "127.0.0.1", Port: "6379"})
	zklock.Initialize(zklock.ZkLockConfig{ZkServers: "127.0.0.1:2181", ConnectTimeOut: 10000})

}

func TestEvaluator(t *testing.T) {
	config := order_difficulty.OrderDifficultyConfig{}
	evaluator := order_difficulty.NewOrderDifficultyEvaluator(config)
	evaluator.Start()
	time.Sleep(1000000 * time.Second)
}

func TestLinearEvaluator_CalcAndSaveDifficulty(t *testing.T) {
	evaluator := order_difficulty.NewLinearEvaluator(big.NewInt(100), 200)
	cntList := make([]*order_difficulty.OrderDifficulty, 0)
	cntList = append(cntList, order_difficulty.NewOrderDiff(100, types.BigintToHex(big.NewInt(100))))
	cntList = append(cntList, order_difficulty.NewOrderDiff(900, types.BigintToHex(big.NewInt(200))))
	cntList = append(cntList, order_difficulty.NewOrderDiff(900, types.BigintToHex(big.NewInt(200))))
	cntList = append(cntList, order_difficulty.NewOrderDiff(400, types.BigintToHex(big.NewInt(200))))
	//cntList = append(cntList, order_difficulty.NewOrderDiff(700, common.BytesToHash(big.NewInt(200).Bytes())))
	//cntList = append(cntList, order_difficulty.NewOrderDiff(700, common.BytesToHash(big.NewInt(200).Bytes())))
	println(evaluator.CalcAndSaveDifficulty(cntList, 10).String())
}
