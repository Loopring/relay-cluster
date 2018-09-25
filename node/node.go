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

package node

import (
	"sync"

	"fmt"
	"github.com/Loopring/relay-cluster/accountmanager"
	"github.com/Loopring/relay-cluster/dao"
	"github.com/Loopring/relay-cluster/gateway"
	"github.com/Loopring/relay-cluster/gateway/order_difficulty"
	"github.com/Loopring/relay-cluster/market"
	ordermanager "github.com/Loopring/relay-cluster/ordermanager/manager"
	orderviewer "github.com/Loopring/relay-cluster/ordermanager/viewer"
	ringtrackerviewer "github.com/Loopring/relay-cluster/ringtrackermanager/viewer"
	txmanager "github.com/Loopring/relay-cluster/txmanager/manager"
	txviewer "github.com/Loopring/relay-cluster/txmanager/viewer"
	"github.com/Loopring/relay-cluster/usermanager"
	socketioutil "github.com/Loopring/relay-cluster/util"
	"github.com/Loopring/relay-lib/cache"
	"github.com/Loopring/relay-lib/cloudwatch"
	"github.com/Loopring/relay-lib/crypto"
	"github.com/Loopring/relay-lib/eth/accessor"
	"github.com/Loopring/relay-lib/eth/loopringaccessor"
	"github.com/Loopring/relay-lib/extractor"
	"github.com/Loopring/relay-lib/kafka"
	"github.com/Loopring/relay-lib/log"
	"github.com/Loopring/relay-lib/marketcap"
	util "github.com/Loopring/relay-lib/marketutil"
	"github.com/Loopring/relay-lib/sns"
	"github.com/Loopring/relay-lib/zklock"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"go.uber.org/zap"
)

type Node struct {
	globalConfig *GlobalConfig
	rdsService   *dao.RdsService
	//ipfsSubService    gateway.IPFSSubService
	orderManager             ordermanager.OrderManager
	orderViewer              orderviewer.OrderViewer
	userManager              usermanager.UserManager
	marketCapProvider        marketcap.MarketCapProvider
	accountManager           accountmanager.AccountManager
	trendManager             market.TrendManager
	tickerCollector          market.CollectorImpl
	tickerManager            market.GetTickerImpl
	globalMarket             market.GlobalMarket
	jsonRpcService           gateway.JsonrpcServiceImpl
	websocketService         gateway.WebsocketServiceImpl
	socketIOService          gateway.SocketIOServiceImpl
	walletService            gateway.WalletServiceImpl
	txManager                txmanager.TransactionManager
	motanService             *gateway.MotanService
	orderDifficultyEvaluator *order_difficulty.OrderDifficultyEvaluator
	wg                       *sync.WaitGroup
	logger                   *zap.Logger

	ringTrackerViewer  ringtrackerviewer.RingTrackerViewer
	ringTrackerService gateway.RingTrackerServiceImpl
}

func NewNode(logger *zap.Logger, globalConfig *GlobalConfig) *Node {
	n := &Node{}
	n.logger = logger
	n.globalConfig = globalConfig
	n.wg = new(sync.WaitGroup)
	// register
	n.registerZklock()
	n.registerSocketIOProducer()
	n.registerSnsNotifier()

	n.registerMysql()
	n.registerCache()

	n.registerMarketUtil()
	n.registerMarketCap()
	n.registerAccessor()
	n.registerUserManager()

	n.registerOrderManager()
	n.registerOrderViewer()

	n.registerAccountManager()
	n.registerGateway()
	n.registerCrypto(nil)

	n.registerTransactionManager()
	n.registerTransactionViewer()

	n.registerTrendManager()
	n.registerTickerCollector()
	n.registerTickerManager()
	n.registerGlobalMarket()
	n.registerWalletService()
	n.registerJsonRpcService()
	n.registerWebsocketService()
	n.registerSocketIOService()

	n.registerExtractor()
	n.registerCloudWatch()

	n.registerRingTrackerViewer()
	n.registerRingTrackerService()
	n.registerOrderDifficultyEvaluator()
	return n
}

func (n *Node) Start() {
	n.orderManager.Start()
	n.marketCapProvider.Start()
	n.accountManager.Start()
	n.txManager.Start()
	//gateway.NewJsonrpcService("8080").Start()
	fmt.Println("step in relay node start")
	n.tickerCollector.Start()
	n.tickerManager.Start()
	n.globalMarket.Start()
	go n.jsonRpcService.Start()
	//n.websocketService.Start()
	go n.socketIOService.Start()
	gateway.StartMotanService(n.globalConfig.MotanServer, n.accountManager, n.orderViewer)
	n.orderDifficultyEvaluator.Start()

	n.wg.Add(1)
}

func (n *Node) Wait() {
	n.wg.Wait()
}

// todo release zklock and kafka producers and consumers
func (n *Node) Stop() {
	n.orderManager.Stop()
	n.txManager.Stop()
	n.wg.Done()
}

func (n *Node) registerCrypto(ks *keystore.KeyStore) {
	c := crypto.NewKSCrypto(true, ks)
	crypto.Initialize(c)
}

func (n *Node) registerMysql() {
	n.rdsService = dao.NewDb(&n.globalConfig.Mysql)
}

func (n *Node) registerCache() {
	cache.NewCache(n.globalConfig.Redis)
}

func (n *Node) registerAccessor() {
	err := accessor.Initialize(n.globalConfig.Accessor)
	err = loopringaccessor.Initialize(n.globalConfig.LoopringProtocol)
	if nil != err {
		log.Fatalf("err:%s", err.Error())
	}
}

//func (n *Node) registerIPFSSubService() {
//	n.ipfsSubService = gateway.NewIPFSSubService(&n.globalConfig.Ipfs)
//}

func (n *Node) registerOrderManager() {
	n.orderManager = ordermanager.NewOrderManager(&n.globalConfig.OrderManager, n.rdsService, n.marketCapProvider, n.globalConfig.Kafka.Brokers)
}

func (n *Node) registerOrderViewer() {
	n.orderViewer = orderviewer.NewOrderViewer(&n.globalConfig.OrderManager, n.rdsService, n.marketCapProvider)
}

func (n *Node) registerTrendManager() {
	n.trendManager = market.NewTrendManager(n.rdsService, n.orderViewer)
}

func (n *Node) registerAccountManager() {
	n.accountManager = accountmanager.Initialize(&n.globalConfig.AccountManager, n.globalConfig.Kafka.Brokers)
}

func (n *Node) registerTransactionManager() {
	n.txManager = txmanager.NewTxManager(n.rdsService)
}

func (n *Node) registerTransactionViewer() {
	txviewer.NewTxView(n.rdsService)
}

func (n *Node) registerTickerCollector() {
	n.tickerCollector = *market.NewCollector()
}

func (n *Node) registerTickerManager() {
	n.tickerManager = *market.NewTickManager(n.trendManager, n.marketCapProvider)
}

func (n *Node) registerGlobalMarket() {
	n.globalMarket = market.NewGlobalMarket(n.globalConfig.MyToken)
}

func (n *Node) registerWalletService() {
	n.walletService = *gateway.NewWalletService(n.trendManager, n.orderViewer,
		n.accountManager, n.marketCapProvider, n.tickerCollector, n.tickerManager, n.rdsService, n.globalConfig.Market.OldVersionWethAddress, n.globalMarket)
}

func (n *Node) registerJsonRpcService() {
	n.jsonRpcService = *gateway.NewJsonrpcService(n.globalConfig.Jsonrpc.Port, &n.walletService, &n.ringTrackerService)
}

func (n *Node) registerWebsocketService() {
	n.websocketService = *gateway.NewWebsocketService(n.globalConfig.Websocket.Port, n.trendManager, n.accountManager, n.marketCapProvider)
}

func (n *Node) registerSocketIOService() {
	n.socketIOService = *gateway.NewSocketIOService(n.globalConfig.Websocket.Port, n.walletService, n.globalConfig.Kafka.Brokers)
}

func (n *Node) registerGateway() {
	gateway.Initialize(&n.globalConfig.GatewayFilters, &n.globalConfig.Gateway, n.orderViewer, n.marketCapProvider, n.accountManager)
}

func (n *Node) registerUserManager() {
	n.userManager = usermanager.NewUserManager(&n.globalConfig.UserManager, n.rdsService)
}

func (n *Node) registerMarketUtil() {
	util.Initialize(&n.globalConfig.Market)
}

func (n *Node) registerMarketCap() {
	n.marketCapProvider = marketcap.NewMarketCapProvider(&n.globalConfig.MarketCap)
}

func (n *Node) registerZklock() {
	if _, err := zklock.Initialize(n.globalConfig.ZkLock); err != nil {
		log.Fatalf("node start, register zklock error:%s", err.Error())
	}
}

func (n *Node) registerSocketIOProducer() {
	socketioutil.Initialize(n.globalConfig.Kafka.Brokers)
}

func (n *Node) registerSnsNotifier() {
	sns.Initialize(n.globalConfig.Sns)
}

func (n *Node) registerExtractor() {
	if err := extractor.Initialize(n.globalConfig.Kafka, kafka.Kafka_Group_RelayCluster_EventOnChain); err != nil {
		log.Fatalf("node start, register extractor error:%s", err.Error())
	}
}

func (n *Node) registerCloudWatch() {
	cloudwatch.Initialize(n.globalConfig.CloudWatch)
}

func (n *Node) registerOrderDifficultyEvaluator() {
	n.orderDifficultyEvaluator = order_difficulty.NewOrderDifficultyEvaluator(n.globalConfig.OrderDifficulty)
}

func (n *Node) registerRingTrackerViewer() {
	n.ringTrackerViewer = ringtrackerviewer.NewRingTrackerViewer(n.rdsService, n.marketCapProvider)
}

func (n *Node) registerRingTrackerService() {
	n.ringTrackerService = *gateway.NewRingTrackerService(n.ringTrackerViewer, n.globalMarket)
}
