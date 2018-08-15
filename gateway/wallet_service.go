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

package gateway

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Loopring/relay-cluster/accountmanager"
	"github.com/Loopring/relay-cluster/dao"
	"github.com/Loopring/relay-cluster/market"
	"github.com/Loopring/relay-cluster/ordermanager/manager"
	"github.com/Loopring/relay-cluster/ordermanager/viewer"
	txtyp "github.com/Loopring/relay-cluster/txmanager/types"
	txmanager "github.com/Loopring/relay-cluster/txmanager/viewer"
	kafkaUtil "github.com/Loopring/relay-cluster/util"
	"github.com/Loopring/relay-lib/cache"
	"github.com/Loopring/relay-lib/crypto"
	"github.com/Loopring/relay-lib/eth/accessor"
	"github.com/Loopring/relay-lib/eth/gasprice_evaluator"
	"github.com/Loopring/relay-lib/eth/loopringaccessor"
	ethtyp "github.com/Loopring/relay-lib/eth/types"
	"github.com/Loopring/relay-lib/kafka"
	"github.com/Loopring/relay-lib/log"
	"github.com/Loopring/relay-lib/marketcap"
	util "github.com/Loopring/relay-lib/marketutil"
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
	"math"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"time"
)

const DefaultCapCurrency = "CNY"
const PendingTxPreKey = "PENDING_TX_"

const SYS_10001 = "10001"
const P2P_50001 = "50001"
const P2P_50002 = "50002"
const P2P_50003 = "50003"
const P2P_50004 = "50004"
const P2P_50005 = "50005"
const P2P_50006 = "50006"
const P2P_50007 = "50007"
const P2P_50008 = "50008"

const OT_STATUS_INIT = "init"
const OT_STATUS_ACCEPT = "accept"
const OT_STATUS_REJECT = "reject"
const OT_REDIS_PRE_KEY = "otrpk_"
const SL_REDIS_PRE_KEY = "slrpk_"
const TS_REDIS_PRE_KEY = "tsrpk_"

type Portfolio struct {
	Token      string `json:"token"`
	Amount     string `json:"amount"`
	Percentage string `json:"percentage"`
}

type PageResult struct {
	Data      []interface{} `json:"data"`
	PageIndex int           `json:"pageIndex"`
	PageSize  int           `json:"pageSize"`
	Total     int           `json:"total"`
}

type Depth struct {
	DelegateAddress string `json:"delegateAddress"`
	Market          string `json:"market"`
	Depth           AskBid `json:"depth"`
}

type AskBid struct {
	Buy  [][]string `json:"buy"`
	Sell [][]string `json:"sell"`
}

type DepthElement struct {
	Price  string   `json:"price"`
	Size   *big.Rat `json:"size"`
	Amount *big.Rat `json:"amount"`
}

type OrderBook struct {
	DelegateAddress string             `json:"delegateAddress"`
	Market          string             `json:"market"`
	Buy             []OrderBookElement `json:"buy"`
	Sell            []OrderBookElement `json:"sell"`
}

type OrderBookElement struct {
	Price      float64 `json:"price"`
	Size       float64 `json:"size"`
	Amount     float64 `json:"amount"`
	OrderHash  string  `json:"orderHash"`
	LrcFee     float64 `json:"lrcFee"`
	SplitS     float64 `json:"splitS"`
	SplitB     float64 `json:"splitB"`
	ValidUntil int64   `json:"validUntil"`
}

type CommonTokenRequest struct {
	DelegateAddress string `json:"delegateAddress"`
	Owner           string `json:"owner"`
}

type SingleDelegateAddress struct {
	DelegateAddress string `json:"delegateAddress"`
}

type SingleMarket struct {
	Market string `json:"market"`
}

type TrendQuery struct {
	Market   string `json:"market"`
	Interval string `json:"interval"`
}

type SingleOwner struct {
	Owner string `json:"owner"`
}

type SingleToken struct {
	Token string `json:"token"`
}

type LatestOrderQuery struct {
	Owner     string `json:"owner"`
	Market    string `json:"market"`
	OrderType string `json:"orderType"`
}

type OrderTransfer struct {
	Hash      string `json:"hash"`
	Origin    string `json:"origin"`
	Status    string `json:"status"`
	Timestamp int64  `json:"timestamp"`
}

type OrderTransferQuery struct {
	Hash string `json:"hash"`
}

type SimpleKey struct {
	Key string `json:"key"`
}

type TempStore struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type NotifyCirculrBody struct {
	Owner string                 `json:"owner"`
	Body  map[string]interface{} `json:"body"`
}

type TxNotify struct {
	Hash     string `json:"hash"`
	Nonce    string `json:"nonce"`
	From     string `json:"from"`
	To       string `json:"to"`
	Value    string `json:"value"`
	GasPrice string `json:"gasPrice"`
	Gas      string `json:"gas"`
	Input    string `json:"input"`
	R        string `json:"r"`
	S        string `json:"s"`
	V        string `json:"v"`
}

type PriceQuoteQuery struct {
	Currency string `json:"currency"`
}

type CutoffRequest struct {
	Address         string `json:"address"`
	DelegateAddress string `json:"delegateAddress"`
	BlockNumber     string `json:"blockNumber"`
}

type EstimatedAllocatedAllowanceQuery struct {
	DelegateAddress string `json:"delegateAddress"`
	Owner           string `json: "owner"`
	Token           string `json: "token"`
}

type EstimatedAllocatedAllowanceResult struct {
	AllocatedResult map[string]string `json:"allocatedResult"`
	FrozenLrcFee    string            `json: "frozenLrcFee"`
}

type TransactionQuery struct {
	ThxHash   string   `json:"thxHash"`
	Owner     string   `json:"owner"`
	Symbol    string   `json: "symbol"`
	Status    string   `json: "status"`
	TxType    string   `json:"txType"`
	TrxHashes []string `json:"trxHashes"`
	PageIndex int      `json:"pageIndex"`
	PageSize  int      `json:"pageSize"`
}

type OrderQuery struct {
	Status          string   `json:"status"`
	PageIndex       int      `json:"pageIndex"`
	PageSize        int      `json:"pageSize"`
	DelegateAddress string   `json:"delegateAddress"`
	Owner           string   `json:"owner"`
	Market          string   `json:"market"`
	OrderHash       string   `json:"orderHash"`
	OrderHashes     []string `json:"orderHashes"`
	Side            string   `json:"side"`
	OrderType       string   `json:"orderType"`
}

type DepthQuery struct {
	DelegateAddress string `json:"delegateAddress"`
	Market          string `json:"market"`
}

type FillQuery struct {
	DelegateAddress string `json:"delegateAddress"`
	Market          string `json:"market"`
	Owner           string `json:"owner"`
	OrderHash       string `json:"orderHash"`
	RingHash        string `json:"ringHash"`
	PageIndex       int    `json:"pageIndex"`
	PageSize        int    `json:"pageSize"`
	Side            string `json:"side"`
	OrderType       string `json:"orderType"`
}

type RingMinedQuery struct {
	DelegateAddress string `json:"delegateAddress"`
	ProtocolAddress string `json:"protocolAddress"`
	RingIndex       string `json:"ringIndex"`
	PageIndex       int    `json:"pageIndex"`
	PageSize        int    `json:"pageSize"`
}

type RawOrderJsonResult struct {
	Protocol        string `json:"protocol"`        // 智能合约地址
	DelegateAddress string `json:"delegateAddress"` // 智能合约地址
	Owner           string `json:"address"`
	Hash            string `json:"hash"`
	TokenS          string `json:"tokenS"`  // 卖出erc20代币智能合约地址
	TokenB          string `json:"tokenB"`  // 买入erc20代币智能合约地址
	AmountS         string `json:"amountS"` // 卖出erc20代币数量上限
	AmountB         string `json:"amountB"` // 买入erc20代币数量上限
	ValidSince      string `json:"validSince"`
	ValidUntil      string `json:"validUntil"` // 订单过期时间
	//Salt                  string `json:"salt"`
	LrcFee                string `json:"lrcFee"` // 交易总费用,部分成交的费用按该次撮合实际卖出代币额与比例计算
	BuyNoMoreThanAmountB  bool   `json:"buyNoMoreThanAmountB"`
	MarginSplitPercentage string `json:"marginSplitPercentage"` // 不为0时支付给交易所的分润比例，否则视为100%
	V                     string `json:"v"`
	R                     string `json:"r"`
	S                     string `json:"s"`
	WalletAddress         string `json:"walletAddress" gencodec:"required"`
	AuthAddr              string `json:"authAddr" gencodec:"required"`       //
	AuthPrivateKey        string `json:"authPrivateKey" gencodec:"required"` //
	Market                string `json:"market"`
	Side                  string `json:"side"`
	CreateTime            int64  `json:"createTime"`
	OrderType             string `json:"orderType"`
}

type OrderJsonResult struct {
	RawOrder         RawOrderJsonResult `json:"originalOrder"`
	DealtAmountS     string             `json:"dealtAmountS"`
	DealtAmountB     string             `json:"dealtAmountB"`
	CancelledAmountS string             `json:"cancelledAmountS"`
	CancelledAmountB string             `json:"cancelledAmountB"`
	Status           string             `json:"status"`
}

type PriceQuote struct {
	Currency string       `json:"currency"`
	Tokens   []TokenPrice `json:"tokens"`
}

type TokenPrice struct {
	Token string  `json:"symbol"`
	Price float64 `json:"price"`
}

type RingMinedDetail struct {
	RingInfo RingMinedInfo   `json:"ringInfo"`
	Fills    []dao.FillEvent `json:"fills"`
}

type RingMinedInfo struct {
	ID                 int                 `json:"id"`
	Protocol           string              `json:"protocol"`
	DelegateAddress    string              `json:"delegateAddress"`
	RingIndex          string              `json:"ringIndex"`
	RingHash           string              `json:"ringHash"`
	TxHash             string              `json:"txHash"`
	Miner              string              `json:"miner"`
	FeeRecipient       string              `json:"feeRecipient"`
	IsRinghashReserved bool                `json:"isRinghashReserved"`
	BlockNumber        int64               `json:"blockNumber"`
	TotalLrcFee        string              `json:"totalLrcFee"`
	TotalSplitFee      map[string]*big.Int `json:"totalSplitFee"`
	TradeAmount        int                 `json:"tradeAmount"`
	Time               int64               `json:"timestamp"`
}

type Token struct {
	Token     string `json:"symbol"`
	Balance   string `json:"balance"`
	Allowance string `json:"allowance"`
}

type AccountJson struct {
	DelegateAddress string  `json:"delegateAddress"`
	Address         string  `json:"owner"`
	Tokens          []Token `json:"tokens"`
}

type LatestFill struct {
	CreateTime   int64   `json:"createTime"`
	Price        float64 `json:"price"`
	Amount       float64 `json:"amount"`
	Side         string  `json:"side"`
	RingHash     string  `json:"ringHash"`
	LrcFee       string  `json:"lrcFee"`
	SplitS       string  `json:"splitS"`
	SplitB       string  `json:"splitB"`
	OrderHash    string  `json:"orderHash"`
	PreOrderHash string  `json:"preOrderHash"`
}

type CancelOrderQuery struct {
	Sign       SignInfo `json:"sign"`
	OrderHash  string   `json:"orderHash"`
	CutoffTime int64    `json:"cutoff"`
	TokenS     string   `json:"tokenS"`
	TokenB     string   `json:"tokenB"`
	Type       uint8    `json:"type"`
}

type SignInfo struct {
	Timestamp string `json:"timestamp"`
	V         uint8  `json:"v"'`
	R         string `json:"r"`
	S         string `json:"s"`
	Owner     string `json:"owner"`
}

type Ticket struct {
	Sign   SignInfo           `json:"sign"`
	Ticket dao.TicketReceiver `json:"ticket"`
}

type TicketQuery struct {
	Sign  SignInfo `json:"sign"`
	Owner string   `json:"owner"`
}

type LoginInfo struct {
	Owner string `json:"owner"`
	UUID  string `json:"uuid"`
}

type SignedLoginInfo struct {
	Sign SignInfo `json:"sign"`
	UUID string   `json:"uuid"`
}

type P2PRingRequest struct {
	RawTx string `json:"rawTx"`
	//Taker          *types.OrderJsonRequest `json:"taker"`
	TakerOrderHash string `json:"takerOrderHash"`
	MakerOrderHash string `json:"makerOrderHash"`
}

type AddTokenReq struct {
	Owner                string `json:"owner"`
	TokenContractAddress string `json:"tokenContractAddress"`
	Symbol               string `json:"symbol"`
	Decimals             int64  `json:"decimals"`
}

type WalletServiceImpl struct {
	trendManager    market.TrendManager
	orderViewer     viewer.OrderViewer
	accountManager  accountmanager.AccountManager
	marketCap       marketcap.MarketCapProvider
	tickerCollector market.CollectorImpl
	globalMarket    market.GlobalMarket
	rds             *dao.RdsService
	oldWethAddress  string
}

func NewWalletService(trendManager market.TrendManager, orderViewer viewer.OrderViewer, accountManager accountmanager.AccountManager,
	capProvider marketcap.MarketCapProvider, collector market.CollectorImpl, rds *dao.RdsService, oldWethAddress string, globalMarket market.GlobalMarket) *WalletServiceImpl {
	w := &WalletServiceImpl{}
	w.trendManager = trendManager
	w.orderViewer = orderViewer
	w.accountManager = accountManager
	w.marketCap = capProvider
	w.tickerCollector = collector
	w.rds = rds
	w.oldWethAddress = oldWethAddress
	w.globalMarket = globalMarket
	return w
}
func (w *WalletServiceImpl) TestPing(input int) (resp []byte, err error) {

	var res string
	if input > 0 {
		res = "input is bigger than zero " + time.Now().String()
	} else if input == 0 {
		res = "input is equal zero " + time.Now().String()
	} else if input < 0 {
		res = "input is smaller than zero " + time.Now().String()
	}
	resp = []byte("{'abc' : '" + res + "'}")
	return
}

func (w *WalletServiceImpl) GetPortfolio(query SingleOwner) (res []Portfolio, err error) {
	res = make([]Portfolio, 0)
	if !common.IsHexAddress(query.Owner) {
		return nil, errors.New("owner can't be nil")
	}

	balances, _ := accountmanager.GetBalanceWithSymbolResult(common.HexToAddress(query.Owner))
	if len(balances) == 0 {
		return
	}

	priceQuote, err := w.GetPriceQuote(PriceQuoteQuery{DefaultCapCurrency})
	if err != nil {
		return
	}

	priceQuoteMap := make(map[string]*big.Rat)
	for _, pq := range priceQuote.Tokens {
		priceQuoteMap[pq.Token] = new(big.Rat).SetFloat64(pq.Price)
	}

	totalAsset := big.NewRat(0, 1)
	for k, v := range balances {
		asset := new(big.Rat).Set(priceQuoteMap[k])
		asset = asset.Mul(asset, new(big.Rat).SetFrac(v, big.NewInt(1)))
		totalAsset = totalAsset.Add(totalAsset, asset)
	}

	for k, v := range balances {
		portfolio := Portfolio{Token: k, Amount: v.String()}
		asset := new(big.Rat).Set(priceQuoteMap[k])
		asset = asset.Mul(asset, new(big.Rat).SetFrac(v, big.NewInt(1)))
		totalAssetFloat, _ := totalAsset.Float64()
		var percentage float64
		if totalAssetFloat == 0 {
			percentage = 0
		} else {
			percentage, _ = asset.Quo(asset, totalAsset).Float64()
		}
		portfolio.Percentage = fmt.Sprintf("%.4f%%", 100*percentage)
		res = append(res, portfolio)
	}

	sort.Slice(res, func(i, j int) bool {
		percentStrLeft := strings.Replace(res[i].Percentage, "%", "", 1)
		percentStrRight := strings.Replace(res[j].Percentage, "%", "", 1)
		left, _ := strconv.ParseFloat(percentStrLeft, 64)
		right, _ := strconv.ParseFloat(percentStrRight, 64)
		return left > right
	})

	return
}

func (w *WalletServiceImpl) GetPriceQuote(query PriceQuoteQuery) (result PriceQuote, err error) {

	rst := PriceQuote{query.Currency, make([]TokenPrice, 0)}
	for k, v := range util.AllTokens {
		price, err := w.marketCap.GetMarketCapByCurrency(v.Protocol, query.Currency)
		if err != nil {
			log.Debug(">>>>>>>> get market cap error " + err.Error())
			rst.Tokens = append(rst.Tokens, TokenPrice{k, 0.0})
		} else {
			floatPrice, _ := price.Float64()
			rst.Tokens = append(rst.Tokens, TokenPrice{k, floatPrice})
			if k == "WETH" {
				rst.Tokens = append(rst.Tokens, TokenPrice{"ETH", floatPrice})
			}
		}
	}
	return rst, nil
}

func (w *WalletServiceImpl) GetTickers(mkt SingleMarket) (result map[string]market.Ticker, err error) {
	result = make(map[string]market.Ticker)
	loopringTicker, err := w.trendManager.GetTickerByMarket(mkt.Market)
	if err == nil {
		result["loopr"] = loopringTicker
	} else {
		log.Info("get ticker from loopring error" + err.Error())
		return result, err
	}
	outTickers, err := w.tickerCollector.GetTickers(mkt.Market)
	if err == nil {
		for _, v := range outTickers {
			result[v.Exchange] = v
		}
	} else {
		log.Info("get other exchanges ticker error" + err.Error())
	}
	return result, nil
}

func (w *WalletServiceImpl) UnlockWallet(owner SingleOwner) (result string, err error) {
	if len(owner.Owner) == 0 {
		return "", errors.New("owner can't be null string")
	}

	unlockRst := w.accountManager.UnlockedWallet(owner.Owner)
	if unlockRst != nil {
		return "", unlockRst
	} else {
		return "unlock_notice_success", nil
	}
}

func (w *WalletServiceImpl) NotifyTransactionSubmitted(txNotify TxNotify) (result string, err error) {

	log.Info("input transaciton found > >>>>>>>>" + txNotify.Hash)

	if len(txNotify.Hash) == 0 {
		return "", errors.New("raw tx can't be null string")
	}
	if !common.IsHexAddress(txNotify.From) || !common.IsHexAddress(txNotify.To) {
		return "", errors.New("from or to address is illegal")
	}

	nonce := types.HexToBigint(txNotify.Nonce)

	err = txmanager.ValidateNonce(txNotify.From, nonce)
	if err != nil {
		log.Infof("nonce invalid in tx %s, %s", txNotify.Hash, err.Error())
		return "", err
	}

	tx := &ethtyp.Transaction{}
	tx.Hash = txNotify.Hash

	if len(txNotify.Input) > 2 && !strings.HasPrefix(txNotify.Input, "0x") && !strings.HasPrefix(txNotify.Input, "0X") {
		tx.Input = "0x" + txNotify.Input
	} else {
		tx.Input = txNotify.Input
	}

	tx.From = txNotify.From
	tx.To = txNotify.To
	tx.Gas = *types.NewBigPtr(types.HexToBigint(txNotify.Gas))
	tx.GasPrice = *types.NewBigPtr(types.HexToBigint(txNotify.GasPrice))
	tx.Nonce = *types.NewBigPtr(types.HexToBigint(txNotify.Nonce))
	tx.Value = *types.NewBigPtr(types.HexToBigint(txNotify.Value))
	if len(txNotify.V) > 0 {
		tx.V = txNotify.V
	}
	if len(txNotify.R) > 0 {
		tx.R = txNotify.R
	}
	if len(txNotify.S) > 0 {
		tx.S = txNotify.S
	}
	tx.BlockNumber = *types.NewBigWithInt(0)
	tx.BlockHash = ""
	tx.TransactionIndex = *types.NewBigWithInt(0)

	log.Debug("emit Pending tx >>>>>>>>>>>>>>>> " + tx.Hash)
	kafkaUtil.ProducerNormalMessage(kafka.Kafka_Topic_Extractor_PendingTransaction, tx)
	txByte, err := json.Marshal(txNotify)
	if err == nil {
		err = cache.Set(PendingTxPreKey+strings.ToUpper(txNotify.Hash), txByte, 3600*24*7)
		if err != nil {
			return "", err
		}
	}
	log.Info("emit transaction info " + tx.Hash)
	return tx.Hash, nil
}

func (w *WalletServiceImpl) GetOldVersionWethBalance(owner SingleOwner) (res string, err error) {
	b, err := loopringaccessor.Erc20Balance(common.HexToAddress(w.oldWethAddress), common.HexToAddress(owner.Owner), "latest")
	if err != nil {
		return
	} else {
		return types.BigintToHex(b), nil
	}
}

func (w *WalletServiceImpl) SubmitOrder(order *types.OrderJsonRequest) (res string, err error) {

	if order.OrderType != types.ORDER_TYPE_MARKET && order.OrderType != types.ORDER_TYPE_P2P {
		order.OrderType = types.ORDER_TYPE_MARKET
	}

	return HandleInputOrder(types.ToOrder(order))
}

func (w *WalletServiceImpl) GetOrders(query *OrderQuery) (res PageResult, err error) {
	orderQuery, statusList, pi, ps := convertFromQuery(query)
	src, err := w.orderViewer.GetOrders(orderQuery, statusList, pi, ps)
	if err != nil {
		log.Info("query order error : " + err.Error())
	}

	rst := PageResult{Total: src.Total, PageIndex: src.PageIndex, PageSize: src.PageSize, Data: make([]interface{}, 0)}

	for _, d := range src.Data {
		o := d.(types.OrderState)
		rst.Data = append(rst.Data, orderStateToJson(o))
	}
	return rst, err
}

// 查询p2p订单, 订单类型固定, market不限
//func (w *WalletServiceImpl) GetP2pOrders(query *OrderQuery) (res PageResult, err error) {
//	orderQuery, statusList, pi, ps := convertFromQuery(query)
//	orderQuery["order_type"] = types.ORDER_TYPE_P2P
//	if _, ok := orderQuery["market"]; ok {
//		delete(orderQuery, "market")
//	}
//
//	src, err := w.orderViewer.GetOrders(orderQuery, statusList, pi, ps)
//	if err != nil {
//		log.Info("query order error : " + err.Error())
//	}
//
//	rst := PageResult{Total: src.Total, PageIndex: src.PageIndex, PageSize: src.PageSize, Data: make([]interface{}, 0)}
//
//	for _, d := range src.Data {
//		o := d.(types.OrderState)
//		rst.Data = append(rst.Data, orderStateToJson(o))
//	}
//	return rst, err
//}

func (w *WalletServiceImpl) GetOrderByHash(query OrderQuery) (order OrderJsonResult, err error) {
	if len(query.OrderHash) == 0 {
		return order, errors.New("order hash can't be null")
	} else {
		state, err := w.orderViewer.GetOrderByHash(common.HexToHash(query.OrderHash))
		if err != nil {
			return order, err
		} else {
			return orderStateToJson(*state), err
		}
	}
}

func (w *WalletServiceImpl) GetOrdersByHashes(query OrderQuery) (order []OrderJsonResult, err error) {
	if query.OrderHashes == nil || len(query.OrderHashes) == 0 {
		return order, errors.New("param orderHashes can't be empty")
	}
	if len(query.OrderHashes) > 50 {
		return order, errors.New("param orderHashes's length can't be over 50")
	} else {
		rst := make([]OrderJsonResult, 0)
		orderHashHex := make([]common.Hash, len(query.OrderHashes))
		for _, oh := range query.OrderHashes {
			orderHashHex = append(orderHashHex, common.HexToHash(oh))
		}
		orderList, err := w.orderViewer.GetOrdersByHashes(orderHashHex)
		if err != nil {
			return order, err
		} else {
			for _, order := range orderList {
				rst = append(rst, orderStateToJson(order))
			}
			return rst, err
		}
	}
}

func (w *WalletServiceImpl) SubmitRingForP2P(p2pRing P2PRingRequest) (res string, err error) {

	log.Info("p2pOrder code line 705:" + p2pRing.RawTx + "," + p2pRing.TakerOrderHash + "," + p2pRing.MakerOrderHash)
	maker, err := w.orderViewer.GetOrderByHash(common.HexToHash(p2pRing.MakerOrderHash))
	if err != nil {
		return res, errors.New(P2P_50001)
	}

	taker, err := w.orderViewer.GetOrderByHash(common.HexToHash(p2pRing.TakerOrderHash))
	if err != nil {
		return res, errors.New(P2P_50008)
	}

	if taker.RawOrder.OrderType != types.ORDER_TYPE_P2P || maker.RawOrder.OrderType != types.ORDER_TYPE_P2P {
		//return res, errors.New("only p2p order can be submitted")
		return res, errors.New(P2P_50002)
	}

	if !maker.IsEffective() {
		//return res, errors.New("maker order has been finished, can't be match ring again")
		return res, errors.New(P2P_50003)
	}

	if taker.RawOrder.Owner.Hex() == maker.RawOrder.Owner.Hex() {
		//return res, errors.New("taker and maker's address can't be same")
		return res, errors.New(P2P_50005)
	}

	if manager.IsDustyOrder(maker) {
		//return res, errors.New("It's dusty order")
		return res, errors.New(P2P_50004)
	}
	log.Info("p2pOrder code line 735")
	remainedAmountS, _ := maker.RemainedAmount()
	log.Info("p2pOrder code line 737:" + remainedAmountS.String())
	if pendingAmountB, err := manager.GetP2PPendingAmount(maker.RawOrder.Hash.Hex()); nil != err {
		log.Info("p2pOrder code line 739:" + err.Error())
		return res, err
	} else {
		log.Info("p2pOrder code line 742:" + pendingAmountB.String())
		if pendingAmountB.Cmp(remainedAmountS) >= 0 {
			//return res, errors.New("maker's remainedAmount is not enough")
			log.Info("p2pOrder code line 745")
			return res, errors.New(P2P_50004)
		}
		log.Info("p2pOrder code line 747")
	}

	var txHashRst string
	err = accessor.SendRawTransaction(&txHashRst, p2pRing.RawTx)
	if err != nil {
		return res, err
	}
	log.Info("p2pOrder code line 756")
	err = manager.SaveP2POrderRelation(taker.RawOrder.Owner.Hex(), taker.RawOrder.Hash.Hex(), maker.RawOrder.Owner.Hex(), maker.RawOrder.Hash.Hex(), txHashRst, taker.RawOrder.AmountB.String(), maker.RawOrder.ValidUntil.String())
	if err != nil {
		return res, errors.New(SYS_10001)
	}
	log.Info("p2pOrder code line 761")
	return txHashRst, nil
}

func (w *WalletServiceImpl) GetLatestOrders(query LatestOrderQuery) (res []OrderJsonResult, err error) {
	orderQuery, _, _, _ := convertFromQuery(&OrderQuery{Owner: query.Owner, Market: query.Market, OrderType: query.OrderType})
	queryRst, err := w.orderViewer.GetLatestOrders(orderQuery, 40)
	if err != nil {
		return res, err
	}

	res = make([]OrderJsonResult, 0)
	for _, d := range queryRst {
		res = append(res, orderStateToJson(d))
	}
	return res, err
}

func (w *WalletServiceImpl) GetLatestMarketOrders(query LatestOrderQuery) (res []OrderJsonResult, err error) {
	query.OrderType = types.ORDER_TYPE_MARKET
	return w.GetLatestOrders(query)
}

func (w *WalletServiceImpl) GetLatestP2POrders(query LatestOrderQuery) (res []OrderJsonResult, err error) {
	query.OrderType = types.ORDER_TYPE_P2P
	return w.GetLatestOrders(query)
}

func (w *WalletServiceImpl) GetDepth(query DepthQuery) (res Depth, err error) {

	defaultDepthLength := 100
	asks, bids, err := w.getInnerOrderBook(query, defaultDepthLength)
	if err != nil {
		return
	}

	mkt := strings.ToUpper(query.Market)
	delegateAddress := query.DelegateAddress
	a, b := util.UnWrap(mkt)

	empty := make([][]string, 0)

	for i := range empty {
		empty[i] = make([]string, 0)
	}
	askBid := AskBid{Buy: empty, Sell: empty}
	depth := Depth{DelegateAddress: delegateAddress, Market: mkt, Depth: askBid}
	depth.Depth.Sell = w.calculateDepth(asks, defaultDepthLength, true, util.AllTokens[a].Decimals, util.AllTokens[b].Decimals)
	depth.Depth.Buy = w.calculateDepth(bids, defaultDepthLength, false, util.AllTokens[b].Decimals, util.AllTokens[a].Decimals)
	return depth, err
}

func (w *WalletServiceImpl) GetUnmergedOrderBook(query DepthQuery) (res OrderBook, err error) {

	defaultDepthLength := 40
	asks, bids, err := w.getInnerOrderBook(query, defaultDepthLength)
	if err != nil {
		return
	}

	mkt := strings.ToUpper(query.Market)
	delegateAddress := query.DelegateAddress
	a, b := util.UnWrap(mkt)
	orderBook := OrderBook{DelegateAddress: delegateAddress, Market: mkt, Buy: make([]OrderBookElement, 0), Sell: make([]OrderBookElement, 0)}
	orderBook.Sell, _ = w.generateOrderBook(asks, true, util.AllTokens[a].Decimals, util.AllTokens[b].Decimals, defaultDepthLength)
	orderBook.Buy, _ = w.generateOrderBook(bids, false, util.AllTokens[b].Decimals, util.AllTokens[a].Decimals, defaultDepthLength)
	return orderBook, err
}

func (w *WalletServiceImpl) getInnerOrderBook(query DepthQuery, defaultDepthLength int) (asks, bids []types.OrderState, err error) {

	mkt := strings.ToUpper(query.Market)
	delegateAddress := query.DelegateAddress

	if mkt == "" || !common.IsHexAddress(delegateAddress) {
		err = errors.New("market and correct contract address must be applied")
		return
	}

	a, b := util.UnWrap(mkt)

	_, err = util.WrapMarket(a, b)
	if err != nil {
		err = errors.New("unsupported market type")
		return
	}

	empty := make([][]string, 0)

	for i := range empty {
		empty[i] = make([]string, 0)
	}

	//(TODO) 考虑到需要聚合的情况，所以每次取2倍的数据，先聚合完了再cut, 不是完美方案，后续再优化
	asks, askErr := w.orderViewer.GetOrderBook(
		common.HexToAddress(delegateAddress),
		util.AllTokens[a].Protocol,
		util.AllTokens[b].Protocol, defaultDepthLength)

	if askErr != nil {
		err = errors.New("get ask order error , please refresh again")
		return
	}

	bids, bidErr := w.orderViewer.GetOrderBook(
		common.HexToAddress(delegateAddress),
		util.AllTokens[b].Protocol,
		util.AllTokens[a].Protocol, defaultDepthLength)

	if bidErr != nil {
		err = errors.New("get bid order error , please refresh again")
		return
	}

	return asks, bids, err
}

func (w *WalletServiceImpl) GetFills(query FillQuery) (dao.PageResult, error) {
	res, err := w.orderViewer.FillsPageQuery(fillQueryToMap(query))

	if err != nil {
		return dao.PageResult{}, nil
	}

	result := dao.PageResult{PageIndex: res.PageIndex, PageSize: res.PageSize, Total: res.Total, Data: make([]interface{}, 0)}

	for _, f := range res.Data {
		fill := f.(dao.FillEvent)
		//if util.IsBuy(fill.TokenB) {
		//	fill.Side = "buy"
		//} else {
		//	fill.Side = "sell"
		//}
		fill.TokenS = util.AddressToAlias(fill.TokenS)
		fill.TokenB = util.AddressToAlias(fill.TokenB)

		result.Data = append(result.Data, fill)
	}
	return result, nil
}

func (w *WalletServiceImpl) GetLatestFills(query FillQuery) ([]LatestFill, error) {

	rst := make([]LatestFill, 0)
	fillQuery, _, _ := fillQueryToMap(query)
	res, err := w.orderViewer.GetLatestFills(fillQuery, 40)

	if err != nil {
		return rst, err
	}

	for _, f := range res {
		lf, err := toLatestFill(f)
		if err == nil && lf.Price > 0 && lf.Amount > 0 {
			rst = append(rst, lf)
		}
	}
	return rst, nil
}

func (w *WalletServiceImpl) GetTicker() (res []market.Ticker, err error) {
	return w.trendManager.GetTicker()
}

func (w *WalletServiceImpl) GetTrend(query TrendQuery) (res []market.Trend, err error) {
	res, err = w.trendManager.GetTrends(query.Market, query.Interval)
	sort.Slice(res, func(i, j int) bool {
		return res[i].Start > res[j].Start
	})
	return
}

func (w *WalletServiceImpl) GetRingMined(query RingMinedQuery) (res dao.PageResult, err error) {
	return w.orderViewer.RingMinedPageQuery(ringMinedQueryToMap(query))
}

func (w *WalletServiceImpl) GetRingMinedDetail(query RingMinedQuery) (res RingMinedDetail, err error) {

	if query.RingIndex == "" {
		return res, errors.New("ringIndex must be supplied")
	}

	if query.DelegateAddress == "" {
		return res, errors.New("delegate address must be supplied")
	}

	rings, err := w.orderViewer.RingMinedPageQuery(ringMinedQueryToMap(query))

	// todo:如果ringhash重复暂时先取第一条
	if err != nil || rings.Total > 1 {
		log.Errorf("query ring error, %s, %d", err.Error(), rings.Total)
		return res, errors.New("query ring error occurs")
	}

	if rings.Total == 0 {
		return res, errors.New("no ring found by hash")
	}

	ring := rings.Data[0].(dao.RingMinedEvent)
	fills, err := w.orderViewer.FindFillsByRingHash(common.HexToHash(ring.RingHash))
	if err != nil {
		return res, err
	}
	return fillDetail(ring, fills)
}

func (w *WalletServiceImpl) GetBalance(balanceQuery CommonTokenRequest) (res AccountJson, err error) {
	if !common.IsHexAddress(balanceQuery.Owner) {
		return res, errors.New("owner can't be null")
	}
	if !common.IsHexAddress(balanceQuery.DelegateAddress) {
		return res, errors.New("delegate must be address")
	}
	owner := common.HexToAddress(balanceQuery.Owner)
	balances, _ := accountmanager.GetBalanceWithSymbolResult(owner)
	allowances, _ := accountmanager.GetAllowanceWithSymbolResult(owner, common.HexToAddress(balanceQuery.DelegateAddress))

	res = AccountJson{}
	res.DelegateAddress = balanceQuery.DelegateAddress
	res.Address = balanceQuery.Owner
	res.Tokens = []Token{}
	for symbol, balance := range balances {
		token := Token{}
		token.Token = symbol

		if allowance, exists := allowances[symbol]; exists {
			token.Allowance = allowance.String()
		} else {
			token.Allowance = "0"
		}
		token.Balance = balance.String()
		res.Tokens = append(res.Tokens, token)
	}

	return
}

func (w *WalletServiceImpl) GetCutoff(query CutoffRequest) (result int64, err error) {
	var cutoff *big.Int
	err = loopringaccessor.GetCutoff(&cutoff, common.HexToAddress(query.DelegateAddress), common.HexToAddress(query.Address), query.BlockNumber)
	if err != nil {
		return 0, err
	}
	return cutoff.Int64(), nil
}

func (w *WalletServiceImpl) GetEstimatedAllocatedAllowance(query EstimatedAllocatedAllowanceQuery) (frozenAmount string, err error) {
	statusSet := make([]types.OrderStatus, 0)
	statusSet = append(statusSet, types.ORDER_NEW)
	statusSet = append(statusSet, types.ORDER_PARTIAL)

	token := query.Token
	owner := query.Owner

	tokenAddress := util.AliasToAddress(token)
	if tokenAddress.Hex() == "" {
		return "", errors.New("unsupported token alias " + token)
	}
	amount, err := w.orderViewer.GetFrozenAmount(common.HexToAddress(owner), tokenAddress, statusSet, common.HexToAddress(query.DelegateAddress))
	if err != nil {
		return "", err
	}

	return types.BigintToHex(amount), err
}

func (w *WalletServiceImpl) GetFrozenLRCFee(query SingleOwner) (frozenAmount string, err error) {
	statusSet := make([]types.OrderStatus, 0)
	statusSet = append(statusSet, types.ORDER_NEW)
	statusSet = append(statusSet, types.ORDER_PARTIAL)

	owner := query.Owner

	allLrcFee, err := w.orderViewer.GetFrozenLRCFee(common.HexToAddress(owner), statusSet)
	if err != nil {
		return "", err
	}

	return types.BigintToHex(allLrcFee), err
}

func (w *WalletServiceImpl) GetAllEstimatedAllocatedAmount(query EstimatedAllocatedAllowanceQuery) (result EstimatedAllocatedAllowanceResult, err error) {

	if len(query.Owner) == 0 || len(query.DelegateAddress) == 0 {
		return result, errors.New("owner and delegateAddress must be applied")
	}

	allOrders, err := w.getAllOrdersByOwner(query.Owner, query.DelegateAddress)
	if err != nil {
		return result, err
	}

	if len(allOrders) == 0 {
		return result, nil
	}

	tmpResult := make(map[string]*big.Int)

	for _, v := range allOrders {
		token := util.AddressToAlias(v.RawOrder.TokenS.Hex())
		if len(token) == 0 {
			continue
		}

		amountS, _ := v.RemainedAmount()
		amount, ok := tmpResult[token]
		if ok {
			amount = amount.Add(amount, amountS.Num())
		} else {
			tmpResult[token] = amountS.Num()
		}
	}

	resultMap := make(map[string]string)

	for k, v := range tmpResult {
		resultMap[k] = types.BigintToHex(v)
	}

	lrcFee, err := w.GetFrozenLRCFee(SingleOwner{query.Owner})
	if err != nil {
		return result, err
	}

	return EstimatedAllocatedAllowanceResult{resultMap, lrcFee}, err
}

func (w *WalletServiceImpl) getAllOrdersByOwner(owner, delegateAddress string) (orders []types.OrderState, err error) {

	allOrders := make([]interface{}, 0)

	orderQuery := OrderQuery{Owner: owner, DelegateAddress: delegateAddress, PageIndex: 1, PageSize: 200, Status: "ORDER_OPENED"}
	pageRst, err := w.orderViewer.GetOrders(convertFromQuery(&orderQuery))
	if err != nil {
		return orders, err
	}

	allOrders = append(allOrders, pageRst.Data[:]...)

	if pageRst.Total > 200 {
		for i := 2; i < (pageRst.Total/200)+1; i++ {
			orderQuery = OrderQuery{Owner: owner, DelegateAddress: delegateAddress, PageIndex: i, PageSize: 200, Status: "ORDER_OPENED"}
			pageRst, err = w.orderViewer.GetOrders(convertFromQuery(&orderQuery))
			if err != nil {
				return orders, err
			}
			allOrders = append(allOrders, pageRst.Data[:]...)
		}
	}

	for _, v := range allOrders {
		orders = append(orders, v.(types.OrderState))
	}

	return orders, nil
}

func (w *WalletServiceImpl) GetLooprSupportedMarket() (markets []string, err error) {
	return w.GetSupportedMarket()
}

func (w *WalletServiceImpl) GetLooprSupportedTokens() (markets []types.Token, err error) {
	return w.GetSupportedTokens()
}

func (w *WalletServiceImpl) GetContracts() (contracts map[string][]string, err error) {
	rst := make(map[string][]string)
	for k, protocol := range loopringaccessor.ProtocolAddresses() {
		lprP := k.Hex()
		lprDP := protocol.DelegateAddress.Hex()

		v, ok := rst[lprDP]
		if ok {
			v = append(v, lprP)
			rst[lprDP] = v
		} else {
			lprPS := make([]string, 0)
			lprPS = append(lprPS, lprP)
			rst[lprDP] = lprPS
		}
	}
	return rst, nil
}

func (w *WalletServiceImpl) GetSupportedMarket() (markets []string, err error) {
	return util.AllMarkets, err
}

func (w *WalletServiceImpl) GetSupportedTokens() (markets []types.Token, err error) {
	markets = make([]types.Token, 0)
	for _, v := range util.AllTokens {
		markets = append(markets, v)
	}
	return markets, err
}

func (w *WalletServiceImpl) GetTransactions(query TransactionQuery) (PageResult, error) {
	var (
		rst PageResult
		// should be make
		txs           = make([]txtyp.TransactionJsonResult, 0)
		limit, offset int
		err           error
	)

	rst.Data = make([]interface{}, 0)
	rst.PageIndex, rst.PageSize, limit, offset = pagination(query.PageIndex, query.PageSize)
	rst.Total, err = txmanager.GetAllTransactionCount(query.Owner, query.Symbol, query.Status, query.TxType)
	if err != nil {
		return rst, err
	}
	txs, err = txmanager.GetAllTransactions(query.Owner, query.Symbol, query.Status, query.TxType, limit, offset)
	for _, v := range txs {
		rst.Data = append(rst.Data, v)
	}

	if err != nil {
		return rst, err
	}

	return rst, nil
}

func (w *WalletServiceImpl) GetLatestTransactions(query TransactionQuery) ([]txtyp.TransactionJsonResult, error) {
	return txmanager.GetAllTransactions(query.Owner, query.Symbol, query.Status, query.TxType, 40, 0)
}

func pagination(pageIndex, pageSize int) (int, int, int, int) {
	if pageIndex <= 0 {
		pageIndex = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	limit := pageSize
	offset := (pageIndex - 1) * pageSize

	return pageIndex, pageSize, limit, offset
}

func (w *WalletServiceImpl) GetTransactionsByHash(query TransactionQuery) (result []txtyp.TransactionJsonResult, err error) {
	return txmanager.GetTransactionsByHash(query.Owner, query.TrxHashes)
}

func (w *WalletServiceImpl) GetPendingTransactions(query SingleOwner) (result []txtyp.TransactionJsonResult, err error) {
	return txmanager.GetPendingTransactions(query.Owner)
}

func (w *WalletServiceImpl) GetPendingRawTxByHash(query TransactionQuery) (result TxNotify, err error) {
	if len(query.ThxHash) == 0 {
		return result, errors.New("tx hash can't be nil")
	}

	txBytes, err := cache.Get(PendingTxPreKey + strings.ToUpper(query.ThxHash))
	if err != nil {
		return result, err
	}

	var tx TxNotify

	err = json.Unmarshal(txBytes, &tx)
	if err != nil {
		return result, err
	}

	return tx, nil

}

func (w *WalletServiceImpl) GetEstimateGasPrice() (result string, err error) {
	return types.BigintToHex(gasprice_evaluator.EstimateGasPrice(nil, nil)), nil
}

func (w *WalletServiceImpl) ApplyTicket(ticket Ticket) (result string, err error) {

	ticket.Ticket.Address = ticket.Sign.Owner
	isSignCorrect, err := verifySign(ticket.Sign)
	if isSignCorrect {
		exist, err := w.rds.QueryTicketByAddress(ticket.Ticket.Address)
		if err == nil && exist.ID > 0 {
			log.Debugf("update ticket id %d", exist.ID)
			ticket.Ticket.ID = exist.ID
		}
		return "", w.rds.Save(&ticket.Ticket)
	} else {
		return "", err
	}
}

func (w *WalletServiceImpl) QueryTicket(query TicketQuery) (ticket dao.TicketReceiver, err error) {

	isSignCorrect, err := verifySign(query.Sign)
	if isSignCorrect {
		return w.rds.QueryTicketByAddress(query.Sign.Owner)
	} else {
		return ticket, err
	}
}

func (w *WalletServiceImpl) TicketCount() (int, error) {
	return w.rds.TicketCount()
}

func (w *WalletServiceImpl) AddCustomToken(req AddTokenReq) (result string, err error) {
	if !util.IsAddress(req.Owner) || !util.IsAddress(req.TokenContractAddress) {
		return "", errors.New("illegal address format in request")
	}

	decimals := new(big.Int)
	decimals.SetInt64(req.Decimals)
	return req.TokenContractAddress, util.AddToken(
		common.HexToAddress(req.Owner),
		util.CustomToken{Address: common.HexToAddress(req.TokenContractAddress), Symbol: req.Symbol, Decimals: decimals})
}

func convertFromQuery(orderQuery *OrderQuery) (query map[string]interface{}, statusList []types.OrderStatus, pageIndex int, pageSize int) {

	query = make(map[string]interface{})
	statusList = convertStatus(orderQuery.Status)
	if orderQuery.Owner != "" {
		query["owner"] = orderQuery.Owner
	}
	if common.IsHexAddress(orderQuery.DelegateAddress) {
		query["delegate_address"] = orderQuery.DelegateAddress
	}

	if orderQuery.Market != "" {
		query["market"] = orderQuery.Market
	}

	if orderQuery.Side != "" {
		query["side"] = orderQuery.Side
	}

	if orderQuery.OrderHash != "" {
		query["order_hash"] = orderQuery.OrderHash
	}

	if orderQuery.OrderType == types.ORDER_TYPE_MARKET || orderQuery.OrderType == types.ORDER_TYPE_P2P {
		query["order_type"] = orderQuery.OrderType
	} else {
		query["order_type"] = types.ORDER_TYPE_MARKET
	}

	pageIndex = orderQuery.PageIndex
	pageSize = orderQuery.PageSize
	return

}

func convertStatus(s string) []types.OrderStatus {
	switch s {
	case "ORDER_OPENED":
		return []types.OrderStatus{types.ORDER_NEW, types.ORDER_PARTIAL}
	case "ORDER_NEW":
		return []types.OrderStatus{types.ORDER_NEW}
	case "ORDER_PARTIAL":
		return []types.OrderStatus{types.ORDER_PARTIAL}
	case "ORDER_FINISHED":
		return []types.OrderStatus{types.ORDER_FINISHED}
	case "ORDER_CANCELLED":
		return []types.OrderStatus{types.ORDER_CANCEL, types.ORDER_FLEX_CANCEL, types.ORDER_CUTOFF}
	case "ORDER_CUTOFF":
		return []types.OrderStatus{types.ORDER_CUTOFF}
	case "ORDER_EXPIRE":
		return []types.OrderStatus{types.ORDER_EXPIRE}
	case "ORDER_PENDING":
		return []types.OrderStatus{types.ORDER_PENDING}
	case "ORDER_CANCELLING":
		return []types.OrderStatus{types.ORDER_CANCELLING, types.ORDER_CUTOFFING}
	}
	return []types.OrderStatus{}
}

func getStringStatus(order types.OrderState) string {
	s := order.Status

	if order.IsExpired() {
		return "ORDER_EXPIRE"
	}

	switch s {
	case types.ORDER_NEW:
		return "ORDER_OPENED"
	case types.ORDER_PARTIAL:
		return "ORDER_OPENED"
	case types.ORDER_FINISHED:
		return "ORDER_FINISHED"
	case types.ORDER_CANCEL:
		return "ORDER_CANCELLED"
	case types.ORDER_FLEX_CANCEL:
		return "ORDER_CANCELLED"
	case types.ORDER_CUTOFF:
		return "ORDER_CUTOFF"
	case types.ORDER_PENDING:
		return "ORDER_PENDING"
	case types.ORDER_CANCELLING, types.ORDER_CUTOFFING:
		return "ORDER_CANCELLING"
	case types.ORDER_EXPIRE:
		return "ORDER_EXPIRE"
	}
	return "ORDER_UNKNOWN"
}

func (w *WalletServiceImpl) calculateDepth(states []types.OrderState, length int, isAsk bool, tokenSDecimal, tokenBDecimal *big.Int) [][]string {

	if len(states) == 0 {
		return [][]string{}
	}

	depth := make([][]string, 0)
	for i := range depth {
		depth[i] = make([]string, 0)
	}

	depthMap := make(map[string]DepthElement)

	for _, s := range states {

		price := *s.RawOrder.Price
		minAmountS, minAmountB, err := w.calculateOrderBookAmount(s, isAsk, tokenSDecimal, tokenBDecimal)

		if err != nil {
			//log.Errorf("calculate min amount error " + err.Error())
			continue
		}

		if isAsk {
			price = *price.Inv(&price)
			priceFloatStr := price.FloatString(10)
			if v, ok := depthMap[priceFloatStr]; ok {
				amount := v.Amount
				size := v.Size
				amount = amount.Add(amount, minAmountS)
				size = size.Add(size, minAmountB)
				depthMap[priceFloatStr] = DepthElement{Price: v.Price, Amount: amount, Size: size}
			} else {
				depthMap[priceFloatStr] = DepthElement{Price: priceFloatStr, Amount: minAmountS, Size: minAmountB}
			}
		} else {
			priceFloatStr := price.FloatString(10)
			if v, ok := depthMap[priceFloatStr]; ok {
				amount := v.Amount
				size := v.Size
				amount = amount.Add(amount, minAmountB)
				size = size.Add(size, minAmountS)
				depthMap[priceFloatStr] = DepthElement{Price: v.Price, Amount: amount, Size: size}
			} else {
				depthMap[priceFloatStr] = DepthElement{Price: priceFloatStr, Amount: minAmountB, Size: minAmountS}
			}
		}
	}

	for k, v := range depthMap {
		amount, _ := v.Amount.Float64()
		size, _ := v.Size.Float64()
		depth = append(depth, []string{k, strconv.FormatFloat(amount, 'f', 10, 64), strconv.FormatFloat(size, 'f', 10, 64)})
	}

	sort.Slice(depth, func(i, j int) bool {
		cmpA, _ := strconv.ParseFloat(depth[i][0], 64)
		cmpB, _ := strconv.ParseFloat(depth[j][0], 64)
		//if isAsk {
		//	return cmpA < cmpB
		//} else {
		//	return cmpA > cmpB
		//}
		return cmpA > cmpB

	})

	if len(depth) > length {
		if isAsk {
			return depth[len(depth)-length-1:]
		} else {
			return depth[:length]
		}
	} else {
		return depth
	}
}

func (w *WalletServiceImpl) generateOrderBook(states []types.OrderState, isAsk bool, tokenSDecimal, tokenBDecimal *big.Int, length int) (elements []OrderBookElement, err error) {
	if len(states) == 0 {
		return nil, errors.New("orders can't be nil")
	}
	elements = make([]OrderBookElement, 0)

	for _, s := range states {
		o := OrderBookElement{}
		o.OrderHash = s.RawOrder.Hash.Hex()
		o.SplitS = fmtFloat(new(big.Rat).SetFrac(s.SplitAmountS, tokenSDecimal))
		o.SplitB = fmtFloat(new(big.Rat).SetFrac(s.SplitAmountB, tokenBDecimal))
		lrcToken := util.AllTokens["LRC"]
		o.LrcFee = fmtFloat(new(big.Rat).SetFrac(s.RawOrder.LrcFee, lrcToken.Decimals))
		o.ValidUntil = s.RawOrder.ValidUntil.Int64()

		price := *s.RawOrder.Price
		amountS, amountB, err := w.calculateOrderBookAmount(s, isAsk, tokenSDecimal, tokenBDecimal)
		if err != nil {
			continue
		}

		if isAsk {
			price = *price.Inv(&price)
			o.Price = fmtFloat(&price)
			o.Amount = fmtFloat(amountS)
			o.Size = fmtFloat(amountB)
		} else {
			o.Price = fmtFloat(&price)
			o.Amount = fmtFloat(amountB)
			o.Size = fmtFloat(amountS)
		}

		elements = append(elements, o)
	}

	if len(elements) > length {
		if isAsk {
			return elements[len(elements)-length-1:], nil
		} else {
			return elements[:length], nil
		}
	}

	return elements, nil
}

func (w *WalletServiceImpl) calculateOrderBookAmount(state types.OrderState, isAsk bool, tokenSDecimal, tokenBDecimal *big.Int) (amountS, amountB *big.Rat, err error) {

	amountS, amountB = state.RemainedAmount()
	amountS = amountS.Quo(amountS, new(big.Rat).SetFrac(tokenSDecimal, big.NewInt(1)))
	amountB = amountB.Quo(amountB, new(big.Rat).SetFrac(tokenBDecimal, big.NewInt(1)))

	if amountS.Cmp(new(big.Rat).SetFloat64(0)) == 0 {
		log.Debug("amount s is zero, skipped")
		return nil, nil, errors.New("amount s is zero")
	}

	if amountB.Cmp(new(big.Rat).SetFloat64(0)) == 0 {
		log.Debug("amount b is zero, skipped")
		return nil, nil, errors.New("amount b is zero")
	}

	minAmountB := amountB
	minAmountS := amountS

	minAmountS, err = w.getAvailableMinAmount(amountS, state.RawOrder.Owner, state.RawOrder.TokenS, state.RawOrder.DelegateAddress, tokenSDecimal)
	if err != nil {
		return nil, nil, err
	}

	sellPrice := new(big.Rat).SetFrac(state.RawOrder.AmountS, state.RawOrder.AmountB)
	buyPrice := new(big.Rat).SetFrac(state.RawOrder.AmountB, state.RawOrder.AmountS)
	if state.RawOrder.BuyNoMoreThanAmountB {
		limitedAmountS := new(big.Rat).Mul(minAmountB, sellPrice)
		if limitedAmountS.Cmp(minAmountS) < 0 {
			minAmountS = limitedAmountS
		}

		minAmountB = minAmountB.Mul(minAmountS, buyPrice)
	} else {
		limitedAmountB := new(big.Rat).Mul(minAmountS, buyPrice)
		if limitedAmountB.Cmp(minAmountB) < 0 {
			minAmountB = limitedAmountB
		}
		minAmountS = minAmountS.Mul(minAmountB, sellPrice)
	}

	return minAmountS, minAmountB, nil
}

func (w *WalletServiceImpl) getAvailableMinAmount(depthAmount *big.Rat, owner, token, spender common.Address, decimal *big.Int) (amount *big.Rat, err error) {

	amount = depthAmount

	balance, allowance, err := accountmanager.GetBalanceAndAllowance(owner, token, spender)
	if err != nil {
		return
	}

	balanceRat := new(big.Rat).SetFrac(balance, decimal)
	allowanceRat := new(big.Rat).SetFrac(allowance, decimal)

	//log.Info(amount.String())
	//log.Info(balanceRat.String())
	//log.Info(allowanceRat.String())

	if amount.Cmp(balanceRat) > 0 {
		amount = balanceRat
	}

	if amount.Cmp(allowanceRat) > 0 {
		amount = allowanceRat
	}

	if amount.Cmp(new(big.Rat).SetFloat64(1e-8)) < 0 {
		return nil, errors.New("amount is zero, skipped")
	}

	//log.Infof("get reuslt amount is  %s", amount)

	return
}

func (w *WalletServiceImpl) GetGlobalTrend(req SingleToken) (trend []market.GlobalTrend, err error) {
	if len(req.Token) == 0 {
		return nil, errors.New("token required")
	}

	tokenMap, err := w.globalMarket.GetGlobalTrendCache(req.Token)
	if err != nil {
		return nil, err
	}
	return tokenMap[req.Token], err
}

func (w *WalletServiceImpl) GetAllGlobalTrend(req SingleToken) (trend map[string][]market.GlobalTrend, err error) {
	return w.globalMarket.GetGlobalTrendCache("")
}

func (w *WalletServiceImpl) GetGlobalTicker(req SingleToken) (ticker map[string]market.GlobalTicker, err error) {
	return w.globalMarket.GetGlobalTickerCache(req.Token)
}

func (w *WalletServiceImpl) GetGlobalMarketTicker(req SingleToken) (tickers map[string][]market.GlobalMarketTicker, err error) {
	return w.globalMarket.GetGlobalMarketTickerCache(req.Token)
}

func (w *WalletServiceImpl) GetOrderTransfer(req OrderTransferQuery) (ot OrderTransfer, err error) {
	otByte, err := cache.Get(OT_REDIS_PRE_KEY + strings.ToLower(req.Hash))
	if err != nil {
		return ot, err
	} else {

		var orderTransfer OrderTransfer
		err = json.Unmarshal(otByte, &orderTransfer)
		if err != nil {
			return ot, err
		}
		return orderTransfer, err
	}
}

func (w *WalletServiceImpl) GetTempStore(req SimpleKey) (ts string, err error) {
	otByte, err := cache.Get(TS_REDIS_PRE_KEY + strings.ToLower(req.Key))
	if err != nil {
		return ts, err
	} else {
		return string(otByte[:]), err
	}
}

func (w *WalletServiceImpl) SetTempStore(req TempStore) (hash string, err error) {
	if len(req.Key) == 0 {
		return hash, errors.New("key can't be nil")
	}

	err = cache.Set(TS_REDIS_PRE_KEY+strings.ToLower(req.Key), []byte(req.Value), 3600*24)
	return req.Key, err
}

func (w *WalletServiceImpl) NotifyCirculr(req NotifyCirculrBody) (owner string, err error) {
	if len(req.Owner) == 0 {
		return owner, errors.New("owner can't be nil")
	}
	kafkaUtil.ProducerSocketIOMessage(Kafka_Topic_SocketIO_Notify_Circulr, &req)
	return req.Owner, err
}

func (w *WalletServiceImpl) FlexCancelOrder(req CancelOrderQuery) (rst string, err error) {

	isCorrect, err := verifySign(req.Sign)
	if !isCorrect {
		return rst, err
	}

	cancelOrderEvent := types.FlexCancelOrderEvent{}
	cancelOrderEvent.OrderHash = common.HexToHash(req.OrderHash)
	cancelOrderEvent.Owner = common.HexToAddress(req.Sign.Owner)
	cancelOrderEvent.TokenS = common.HexToAddress(req.TokenS)
	cancelOrderEvent.TokenB = common.HexToAddress(req.TokenB)
	cancelOrderEvent.CutoffTime = req.CutoffTime
	cancelOrderEvent.Type = types.FlexCancelType(req.Type)
	err = manager.FlexCancelOrder(&cancelOrderEvent)
	if err == nil {
		go func() {

			orderQuery := OrderQuery{Owner: req.Sign.Owner, OrderType: types.ORDER_TYPE_MARKET}
			if cancelOrderEvent.Type == types.FLEX_CANCEL_BY_HASH {
				orderQuery.OrderHash = cancelOrderEvent.OrderHash.Hex()
			} else if cancelOrderEvent.Type == types.FLEX_CANCEL_BY_MARKET {
				orderQuery.Market, _ = util.WrapMarketByAddress(req.TokenS, req.TokenB)
			} else {
				// other conditions will notify all market, not implement now
			}
			orderQueryMap, _, _, _ := convertFromQuery(&orderQuery)
			ot, err := w.orderViewer.GetLatestOrders(orderQueryMap, 1)
			if err == nil && len(ot) > 0 {
				kafkaUtil.ProducerSocketIOMessage(kafka.Kafka_Topic_SocketIO_Order_Updated, ot[0])
			}
		}()
	}
	return rst, err
}

func (w *WalletServiceImpl) SetOrderTransfer(req OrderTransfer) (hash string, err error) {
	if len(req.Hash) == 0 {
		return hash, errors.New("hash can't be nil")
	}
	req.Status = OT_STATUS_INIT
	req.Timestamp = time.Now().Unix()
	otByte, err := json.Marshal(req)
	if err != nil {
		return hash, err
	}
	err = cache.Set(OT_REDIS_PRE_KEY+strings.ToLower(req.Hash), otByte, 3600*24)
	return req.Hash, err
}

func (w *WalletServiceImpl) UpdateOrderTransfer(req OrderTransfer) (hash string, err error) {
	if len(req.Hash) == 0 {
		return hash, errors.New("hash can't be nil")
	}

	ot, err := w.GetOrderTransfer(OrderTransferQuery{Hash: req.Hash})
	if err != nil {
		return hash, err
	}

	ot.Status = req.Status

	otByte, err := json.Marshal(ot)
	if err != nil {
		return hash, err
	}
	err = cache.Set(OT_REDIS_PRE_KEY+strings.ToLower(req.Hash), otByte, 3600*24)
	if err != nil {
		return hash, err
	} else {
		kafkaUtil.ProducerSocketIOMessage(Kafka_Topic_SocketIO_Order_Transfer, &OrderTransfer{Hash: ot.Hash, Status: ot.Status})
		return req.Hash, err
	}
}

func (w *WalletServiceImpl) NotifyScanLogin(req SignedLoginInfo) (rst string, err error) {

	isCorrect, err := verifySign(req.Sign)
	if !isCorrect {
		return req.UUID, err
	}

	kafkaUtil.ProducerSocketIOMessage(Kafka_Topic_SocketIO_Scan_Login, &LoginInfo{UUID: req.UUID, Owner: req.Sign.Owner})
	return req.UUID, err
}

func (w *WalletServiceImpl) GetNonce(owner SingleOwner) (n int64, err error) {
	nonce, err := txmanager.GetNonce(owner.Owner)
	if err != nil {
		return n, err
	}
	return nonce.Int64(), err
}

func fillQueryToMap(q FillQuery) (map[string]interface{}, int, int) {
	rst := make(map[string]interface{})
	var pi, ps int
	if q.Market != "" {
		rst["market"] = q.Market
	}
	if q.PageIndex <= 0 {
		pi = 1
	} else {
		pi = q.PageIndex
	}
	if q.PageSize <= 0 || q.PageSize > 20 {
		ps = 20
	} else {
		ps = q.PageSize
	}
	if common.IsHexAddress(q.DelegateAddress) {
		rst["delegate_address"] = q.DelegateAddress
	}
	if q.Owner != "" {
		rst["owner"] = q.Owner
	}
	if q.OrderHash != "" {
		rst["order_hash"] = q.OrderHash
	}
	if q.RingHash != "" {
		rst["ring_hash"] = q.RingHash
	}

	if q.Side != "" {
		rst["side"] = q.Side
	}

	if q.OrderType == types.ORDER_TYPE_MARKET || q.OrderType == types.ORDER_TYPE_P2P {
		rst["order_type"] = q.OrderType
	} else {
		rst["order_type"] = types.ORDER_TYPE_MARKET
	}

	return rst, pi, ps
}

func ringMinedQueryToMap(q RingMinedQuery) (map[string]interface{}, int, int) {
	rst := make(map[string]interface{})
	var pi, ps int
	if q.PageIndex <= 0 {
		pi = 1
	} else {
		pi = q.PageIndex
	}
	if q.PageSize <= 0 || q.PageSize > 20 {
		ps = 20
	} else {
		ps = q.PageSize
	}
	if common.IsHexAddress(q.DelegateAddress) {
		rst["delegate_address"] = q.DelegateAddress
	}
	if common.IsHexAddress(q.ProtocolAddress) {
		rst["contract_address"] = q.ProtocolAddress
	}

	if q.RingIndex != "" {
		rst["ring_index"] = types.HexToBigint(q.RingIndex).String()
	}

	return rst, pi, ps
}

func orderStateToJson(src types.OrderState) OrderJsonResult {

	rst := OrderJsonResult{}
	rst.DealtAmountB = types.BigintToHex(src.DealtAmountB)
	rst.DealtAmountS = types.BigintToHex(src.DealtAmountS)
	rst.CancelledAmountB = types.BigintToHex(src.CancelledAmountB)
	rst.CancelledAmountS = types.BigintToHex(src.CancelledAmountS)
	rst.Status = getStringStatus(src)
	rawOrder := RawOrderJsonResult{}
	rawOrder.Protocol = src.RawOrder.Protocol.Hex()
	rawOrder.DelegateAddress = src.RawOrder.DelegateAddress.Hex()
	rawOrder.Owner = src.RawOrder.Owner.Hex()
	rawOrder.Hash = src.RawOrder.Hash.Hex()
	rawOrder.TokenS = util.AddressToAlias(src.RawOrder.TokenS.String())
	rawOrder.TokenB = util.AddressToAlias(src.RawOrder.TokenB.String())
	rawOrder.AmountS = types.BigintToHex(src.RawOrder.AmountS)
	rawOrder.AmountB = types.BigintToHex(src.RawOrder.AmountB)
	rawOrder.ValidSince = types.BigintToHex(src.RawOrder.ValidSince)
	rawOrder.ValidUntil = types.BigintToHex(src.RawOrder.ValidUntil)
	rawOrder.LrcFee = types.BigintToHex(src.RawOrder.LrcFee)
	rawOrder.BuyNoMoreThanAmountB = src.RawOrder.BuyNoMoreThanAmountB
	rawOrder.MarginSplitPercentage = types.BigintToHex(big.NewInt(int64(src.RawOrder.MarginSplitPercentage)))
	rawOrder.V = types.BigintToHex(big.NewInt(int64(src.RawOrder.V)))
	rawOrder.R = src.RawOrder.R.Hex()
	rawOrder.S = src.RawOrder.S.Hex()
	rawOrder.WalletAddress = src.RawOrder.WalletAddress.Hex()
	rawOrder.AuthAddr = src.RawOrder.AuthAddr.Hex()
	rawOrder.Market = src.RawOrder.Market
	auth, _ := src.RawOrder.AuthPrivateKey.MarshalText()
	rawOrder.AuthPrivateKey = string(auth)
	rawOrder.CreateTime = src.RawOrder.CreateTime
	rawOrder.Side = src.RawOrder.Side
	rawOrder.OrderType = src.RawOrder.OrderType
	rst.RawOrder = rawOrder
	return rst
}

func fillDetail(ring dao.RingMinedEvent, fills []dao.FillEvent) (rst RingMinedDetail, err error) {
	rst = RingMinedDetail{Fills: fills}
	ringInfo := RingMinedInfo{}
	ringInfo.ID = ring.ID
	ringInfo.RingHash = ring.RingHash
	ringInfo.BlockNumber = ring.BlockNumber
	ringInfo.Protocol = ring.Protocol
	ringInfo.DelegateAddress = ring.DelegateAddress
	ringInfo.TxHash = ring.TxHash
	ringInfo.Time = ring.Time
	ringInfo.RingIndex = ring.RingIndex
	ringInfo.Miner = ring.Miner
	ringInfo.FeeRecipient = ring.FeeRecipient
	ringInfo.IsRinghashReserved = ring.IsRinghashReserved
	ringInfo.TradeAmount = ring.TradeAmount
	ringInfo.TotalLrcFee = ring.TotalLrcFee
	ringInfo.TotalSplitFee = make(map[string]*big.Int)

	for _, f := range fills {
		if len(f.SplitS) > 0 && f.SplitS != "0" {
			symbol := util.AddressToAlias(f.TokenS)
			if len(symbol) > 0 {
				splitS, _ := new(big.Int).SetString(f.SplitS, 0)
				totalSplitS, ok := ringInfo.TotalSplitFee[symbol]
				if ok {
					ringInfo.TotalSplitFee[symbol] = totalSplitS.Add(splitS, totalSplitS)
				} else {
					ringInfo.TotalSplitFee[symbol] = splitS
				}
			}
		}
		if len(f.SplitB) > 0 && f.SplitB != "0" {
			symbol := util.AddressToAlias(f.TokenB)
			if len(symbol) > 0 {
				splitB, _ := new(big.Int).SetString(f.SplitB, 0)
				totalSplitB, ok := ringInfo.TotalSplitFee[symbol]
				if ok {
					ringInfo.TotalSplitFee[symbol] = totalSplitB.Add(splitB, totalSplitB)
				} else {
					ringInfo.TotalSplitFee[symbol] = splitB
				}
			}
		}
	}

	rst.RingInfo = ringInfo
	return rst, nil
}

func toLatestFill(f dao.FillEvent) (latestFill LatestFill, err error) {
	rst := LatestFill{CreateTime: f.CreateTime}
	price := util.CalculatePrice(f.AmountS, f.AmountB, f.TokenS, f.TokenB)
	rst.Price, _ = strconv.ParseFloat(fmt.Sprintf("%0.8f", price), 64)
	rst.Side = f.Side
	rst.RingHash = f.RingHash
	rst.LrcFee = f.LrcFee
	rst.SplitS = f.SplitS
	rst.SplitB = f.SplitB
	rst.OrderHash = f.OrderHash
	rst.PreOrderHash = f.PreOrderHash
	var amount float64
	if util.GetSide(f.TokenS, f.TokenB) == util.SideBuy {
		amountB, _ := new(big.Int).SetString(f.AmountB, 0)
		tokenB, ok := util.AllTokens[util.AddressToAlias(f.TokenB)]
		if !ok {
			return latestFill, err
		}
		ratAmount := new(big.Rat).SetFrac(amountB, tokenB.Decimals)
		amount, _ = ratAmount.Float64()
		rst.Amount, _ = strconv.ParseFloat(fmt.Sprintf("%0.8f", amount), 64)
	} else {
		amountS, _ := new(big.Int).SetString(f.AmountS, 0)
		tokenS, ok := util.AllTokens[util.AddressToAlias(f.TokenS)]
		if !ok {
			return latestFill, err
		}
		ratAmount := new(big.Rat).SetFrac(amountS, tokenS.Decimals)
		amount, _ = ratAmount.Float64()
		rst.Amount, _ = strconv.ParseFloat(fmt.Sprintf("%0.8f", amount), 64)
	}
	return rst, nil
}

func fmtFloat(src *big.Rat) float64 {
	f, _ := src.Float64()
	rst, _ := strconv.ParseFloat(fmt.Sprintf("%0.8f", f), 64)
	return rst
}

func verifySign(sign SignInfo) (bool, error) {

	now := time.Now().Unix()
	ts, err := strconv.ParseInt(sign.Timestamp, 10, 64)
	if err != nil {
		return false, err
	}

	if math.Abs(float64(now-ts)) > 60*10 {
		return false, errors.New("timestamp had expired")
	}

	h := &common.Hash{}
	address := &common.Address{}
	hashBytes := crypto.GenerateHash([]byte(sign.Timestamp))
	h.SetBytes(hashBytes)
	sig, _ := crypto.VRSToSig(sign.V, types.HexToBytes32(sign.R).Bytes(), types.HexToBytes32(sign.S).Bytes())
	if addressBytes, err := crypto.SigToAddress(h.Bytes(), sig); nil != err {
		log.Errorf("signer address error:%s", err.Error())
		return false, errors.New("sign is incorrect")
	} else {
		address.SetBytes(addressBytes)
		if strings.ToLower(address.Hex()) == strings.ToLower(sign.Owner) {
			return true, nil
		} else {
			return false, errors.New("sign address not matched")
		}
	}
}
