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

package market

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"github.com/Loopring/relay-lib/log"
	util "github.com/Loopring/relay-lib/marketutil"
	"github.com/Loopring/relay-lib/sns"
	"github.com/Loopring/relay-lib/zklock"
	"github.com/pkg/errors"
	"github.com/robfig/cron"
	"io/ioutil"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

var globalMarket *GlobalMarket

type GlobalMarket struct {
	config MyTokenConfig
	cron   *cron.Cron
}

type GlobalTrend struct {
}

type MyTokenConfig struct {
	AppId     string
	AppSecret string
	baseUrl   string
}

type GlobalTrendReq struct {
	TrendAnchor string `json:"trend_anchor"`
	Symbol      string `json:"symbol"`
}

type GlobalTickerReq struct {
	symbol string `json:"symbol"`
}

func NewGlobalMarket(config MyTokenConfig) GlobalMarket {
	return GlobalMarket{config: config}
}

func getResp(url string, result interface{}) error {

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func() {
		if nil != resp && nil != resp.Body {
			resp.Body.Close()
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if nil != err {
		return err
	} else {
		return json.Unmarshal(body, &result)
	}
}

func (g *GlobalMarket) GetGlobalTrend(token string) (trend GlobalTrend, err error) {

	url := g.config.baseUrl + "symbol/trend?"
	request := GlobalTrendReq{TrendAnchor: "USDT", Symbol: token}
	urlParam, err := g.Sign(request)
	if err != nil {
		return trend, err
	}

	var resp GlobalTrend
	getResp(url+urlParam, resp)

	return GlobalTrend{}, nil
}

func (g *GlobalMarket) GetGlobalTicker(token string) (tickers []Ticker, err error) {
	return nil, nil
}

func (g *GlobalMarket) Sign(param interface{}) (urlParam string, err error) {
	jsonStr, err := json.Marshal(param)
	if err != nil {
		return urlParam, err
	}
	var jsonMap map[string]interface{}
	err = json.Unmarshal(jsonStr, &jsonMap)
	if err != nil {
		return urlParam, err
	}

	jsonMap["app_id"] = g.config.AppId
	jsonMap["timestamp"] = time.Now().Unix()

	var keys []string
	for k := range jsonMap {
		keys = append(keys, k)
	}

	signatureList := make([]string, 0)
	sort.Strings(keys)
	for _, k := range keys {
		v := jsonMap[k]
		if reflect.TypeOf(v).Kind() == reflect.String {
			signatureList = append(signatureList, k+"="+v.(string))
		} else if reflect.TypeOf(v).Kind() == reflect.Int64 {
			signatureList = append(signatureList, k+"="+strconv.FormatInt(v.(int64), 10))
		} else {
			return urlParam, errors.New("unsupported data type " + reflect.TypeOf(v).String())
		}
	}

	waitToSign := strings.Join(signatureList, "&") + "&app_secret=" + g.config.AppSecret
	sign := computeHmac256(waitToSign, g.config.AppSecret)

	return waitToSign + "&sign=" + strings.ToUpper(sign), nil
}

func (g *GlobalMarket) Start() {
	go func() {
		if zklock.TryLock(tickerCollectorCronJobZkLock) == nil {
			updateBinanceCache()
			updateOkexCache()
			updateHuobiCache()
			g.cron.AddFunc("@every 20s", updateBinanceCache)
			log.Info("start collect cron jobs......... ")
			g.cron.Start()
		} else {
			err := sns.PublishSns(tryLockFailedMsg, tryLockFailedMsg)
			if err != nil {
				log.Error(err.Error())
			}
		}
	}()
}

func syncGlobalTrend() {
	if globalMarket == nil {
		return
	}
	for k := range util.AllTokens {
		globalMarket.syncGlobalTrend(k)
	}
}

func (g *GlobalMarket) syncGlobalTrend(token string) error {

	return nil
}

func (g *GlobalMarket) syncGlobalTicker(token string) error {
	return nil
}

func (g *GlobalMarket) syncPairOnMarket(market string) error {
	return nil
}

func computeHmac256(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
