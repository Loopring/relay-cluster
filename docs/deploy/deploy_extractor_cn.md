# 部署extractor

## 初始化环境

### 启动EC2实例

启动EC2实例，并在启动实例过程中添加对CodeDeploy的支持，参考[启动aws EC2实例](new_ec2_cn.md)

### 配置安全组
为每个实例配置名为`extractor-SecurityGroup`的安全组
> 如果还没有创建，请参考[配置aws安全组](security_group_cn.md)关于`extractor-SecurityGroup`部分的说明进行配置后再进行关联

### 部署配置文件
目前extractor的基本配置是通过静态配置文件来实现的，所以需要将配置文件在本地配置好并上传所有待部署服务器，不过这个工作只在第一次部署的时候必要，后续都会利用这个静态配置文件启动服务【待优化】

#### 创建配置文件
* extractor.toml

在`Loopring/extractor/config/extractor.toml`的基础上进行如下必要的修改
```
    output_paths = ["/var/log/extractor/zap.log", "stderr"]
    error_output_paths = ["/var/log/extractor/err.log"]
...
[mysql]
    hostname = "xx.xx.xx.xx"
    port = "3306"
    user = "xxx"
    password = "xxx"
...
[redis]
    host = "xx.xx.xx.xx"
    port = "6379"
...
#下面是eth节点的内网ip地址
[accessor]
    raw_urls = ["http://xx.xx.xx.xx:8545", "http://xx.xx.xx.xx:8545"]
[extractor]
    start_block_number = 5863727
    end_block_number = 0
    confirm_block_number = 1
    debug = false
#下面是eth主网合约配置，如果非主网，要联系开源人员获取最新的测试配置
[loopring_protocol]
    implAbi = "[{\"constant\":true,\"inputs\":[],\"name\":\"MARGIN_SPLIT_PERCENTAGE_BASE\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"ringIndex\",\"outputs\":[{\"name\":\"\",\"type\":\"uint64\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"RATE_RATIO_SCALE\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"lrcTokenAddress\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"tokenRegistryAddress\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"delegateAddress\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"orderOwner\",\"type\":\"address\"},{\"name\":\"token1\",\"type\":\"address\"},{\"name\":\"token2\",\"type\":\"address\"}],\"name\":\"getTradingPairCutoffs\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"token1\",\"type\":\"address\"},{\"name\":\"token2\",\"type\":\"address\"},{\"name\":\"cutoff\",\"type\":\"uint256\"}],\"name\":\"cancelAllOrdersByTradingPair\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"addresses\",\"type\":\"address[5]\"},{\"name\":\"orderValues\",\"type\":\"uint256[6]\"},{\"name\":\"buyNoMoreThanAmountB\",\"type\":\"bool\"},{\"name\":\"marginSplitPercentage\",\"type\":\"uint8\"},{\"name\":\"v\",\"type\":\"uint8\"},{\"name\":\"r\",\"type\":\"bytes32\"},{\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"cancelOrder\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"MAX_RING_SIZE\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"cutoff\",\"type\":\"uint256\"}],\"name\":\"cancelAllOrders\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"rateRatioCVSThreshold\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"addressList\",\"type\":\"address[4][]\"},{\"name\":\"uintArgsList\",\"type\":\"uint256[6][]\"},{\"name\":\"uint8ArgsList\",\"type\":\"uint8[1][]\"},{\"name\":\"buyNoMoreThanAmountBList\",\"type\":\"bool[]\"},{\"name\":\"vList\",\"type\":\"uint8[]\"},{\"name\":\"rList\",\"type\":\"bytes32[]\"},{\"name\":\"sList\",\"type\":\"bytes32[]\"},{\"name\":\"feeRecipient\",\"type\":\"address\"},{\"name\":\"feeSelections\",\"type\":\"uint16\"}],\"name\":\"submitRing\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"walletSplitPercentage\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"_ringIndex\",\"type\":\"uint256\"},{\"indexed\":true,\"name\":\"_ringHash\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"_miner\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_feeRecipient\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_orderInfoList\",\"type\":\"bytes32[]\"}],\"name\":\"RingMined\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_orderHash\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"_amountCancelled\",\"type\":\"uint256\"}],\"name\":\"OrderCancelled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_address\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_cutoff\",\"type\":\"uint256\"}],\"name\":\"AllOrdersCancelled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_address\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_token1\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_token2\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_cutoff\",\"type\":\"uint256\"}],\"name\":\"OrdersCancelled\",\"type\":\"event\"}]"
    delegateAbi = "[{\"constant\":true,\"inputs\":[{\"name\":\"owners\",\"type\":\"address[]\"},{\"name\":\"tradingPairs\",\"type\":\"bytes20[]\"},{\"name\":\"validSince\",\"type\":\"uint256[]\"}],\"name\":\"checkCutoffsBatch\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"resume\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"max\",\"type\":\"uint256\"}],\"name\":\"getLatestAuthorizedAddresses\",\"outputs\":[{\"name\":\"addresses\",\"type\":\"address[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"orderHash\",\"type\":\"bytes32\"},{\"name\":\"cancelOrFillAmount\",\"type\":\"uint256\"}],\"name\":\"addCancelledOrFilled\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"cancelled\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"token\",\"type\":\"address\"},{\"name\":\"from\",\"type\":\"address\"},{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"transferToken\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"kill\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"lrcTokenAddress\",\"type\":\"address\"},{\"name\":\"miner\",\"type\":\"address\"},{\"name\":\"feeRecipient\",\"type\":\"address\"},{\"name\":\"walletSplitPercentage\",\"type\":\"uint8\"},{\"name\":\"batch\",\"type\":\"bytes32[]\"}],\"name\":\"batchTransferToken\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"authorizeAddress\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"tokenPair\",\"type\":\"bytes20\"},{\"name\":\"t\",\"type\":\"uint256\"}],\"name\":\"setTradingPairCutoffs\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"claimOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"cancelledOrFilled\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"suspended\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"batch\",\"type\":\"bytes32[]\"}],\"name\":\"batchAddCancelledOrFilled\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"},{\"name\":\"\",\"type\":\"bytes20\"}],\"name\":\"tradingPairCutoffs\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"orderHash\",\"type\":\"bytes32\"},{\"name\":\"cancelAmount\",\"type\":\"uint256\"}],\"name\":\"addCancelled\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"addressInfos\",\"outputs\":[{\"name\":\"previous\",\"type\":\"address\"},{\"name\":\"index\",\"type\":\"uint32\"},{\"name\":\"authorized\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"isAddressAuthorized\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"cutoffs\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"pendingOwner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"suspend\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"deauthorizeAddress\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"t\",\"type\":\"uint256\"}],\"name\":\"setCutoffs\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"addr\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"number\",\"type\":\"uint32\"}],\"name\":\"AddressAuthorized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"addr\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"number\",\"type\":\"uint32\"}],\"name\":\"AddressDeauthorized\",\"type\":\"event\"}]"
    tokenRegistryAbi = "[{\"constant\":false,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"},{\"name\":\"symbol\",\"type\":\"string\"}],\"name\":\"unregisterToken\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"symbol\",\"type\":\"string\"}],\"name\":\"getAddressBySymbol\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"addressList\",\"type\":\"address[]\"}],\"name\":\"areAllTokensRegistered\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"isTokenRegistered\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"start\",\"type\":\"uint256\"},{\"name\":\"count\",\"type\":\"uint256\"}],\"name\":\"getTokens\",\"outputs\":[{\"name\":\"addressList\",\"type\":\"address[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"claimOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"},{\"name\":\"symbol\",\"type\":\"string\"}],\"name\":\"registerToken\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"pendingOwner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"addresses\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"symbol\",\"type\":\"string\"}],\"name\":\"isTokenRegisteredBySymbol\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"addr\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"symbol\",\"type\":\"string\"}],\"name\":\"TokenRegistered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"addr\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"symbol\",\"type\":\"string\"}],\"name\":\"TokenUnregistered\",\"type\":\"event\"}]"
    [loopring_protocol.address]
         "v1.5.1" = "0x781870080c8c24a2fd6882296c49c837b06a65e6"
...
[market]
    token_file = "/opt/loopring/extractor/config/tokens.json"
...
#zk内网ip地址
[zk_lock]
    zk_servers = "xx.xx.xx.xx:2181,xx.xx.xx.xx:2181,xx.xx.xx.xx:2181"
    connect_time_out = 10

#kafka内网ip地址
[kafka]
    brokers = ["xx.xx.xx.xx:9092","xx.xx.xx.xx:9092","xx.xx.xx.xx:9092"]
...
[cloud_watch]
    enabled = false
    region = ""
```

> cloud_watch如果设置`enabled`为true，请参考[ec2](new_ec2_cn.md)部署鉴权文件，region取值请参考[aws doc](https://docs.aws.amazon.com/zh_cn/AWSEC2/latest/UserGuide/using-regions-availability-zones.html)

* tokens.json

在[tokens.json](tokens_main.md)的基础上根据实际需要进行必要的裁剪

#### 配置EC2实例
在EC2实例执行脚本
```
sudo mkdir -p /opt/loopring/extractor
sudo chown -R ubuntu:ubuntu /opt/loopring
cd /opt/loopring/extractor 
sudo mkdir bin/ config/ src/
```
上传本地配置文件
```
scp -i xx.pem extractor.toml ubuntu@x.x.x.x:/opt/loopring/extractor/config
scp -i xx.pem tokens.json ubuntu@x.x.x.x:/opt/loopring/extractor/config
```

## 部署
通过CodeDeploy进行配置，详细步骤参考[接入CodeDeloy](codedeploy_cn.md)

## 服务日志

### extractor业务日志
`/var/log/extractor/zap.log`

### stdout
`/var/log/svc/extractor/current`

## 启停
通过CodeDeploy的方式部署会为服务添加daemontools支持，也就是服务如果意外终止，会自动启动，所以不能通过kill的方式手动停止

### 启动
`sudo svc -u /etc/service/extractor`

### 停止
`sudo svc -d /etc/service/extractor`
