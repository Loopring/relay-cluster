# 部署miner

## 初始化环境

### 启动EC2实例
启动EC2实例，并在启动实例过程中添加对CodeDeploy的支持，参考[启动aws EC2实例](new_ec2_cn.md)

### 配置安全组
关联`miner-SecurityGroup`安全组。如果未创建该安全组，请参考[aws安全组](security_group_cn.md)关于`miner-SecurityGroup`安全组的说明，创建后再关联

## 部署配置文件

目前miner的基本配置是通过静态配置文件来实现的，所以需要将配置文件在本地配置好并上传所有待部署服务器，不过这个工作只在第一次部署的时候必要，后续都会利用这个静态配置文件启动服务【待优化】

### 创建配置文件
* miner.toml

在`Loopring/miner/config/miner.toml`的基础上进行如下必要的修改
```
    output_paths = ["/var/log/miner/zap.log"]
    error_output_paths = ["/var/log/miner/err.log"]
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
#下面是eth节点的内网ip地址
[accessor]
    raw_urls = ["http://xx.xx.xx.xx:8545", "http://xx.xx.xx.xx:8545"]
#下面是eth主网合约配置，如果非主网，要联系开源人员获取最新的测试配置
[loopring_accessor.address]
    "v1.5" = "0x8d8812b72d1e4ffCeC158D25f56748b7d67c1e78"
...
[miner]
    ....
    feeReceipt = "0x111111111111111111111111111111"
    [[miner.normal_miners]]
        address = "0x111111111111111111111111111111"
...
[keystore]
    keydir = "/opt/loopring/miner/config/keystore"
...
[market_cap]
        base_url = "https://api.coinmarketcap.com/v2/ticker/?convert=%s&start=%d&limit=%d"
        currency = "CNY"
...
[market_util]
    token_file = "/opt/loopring/miner/config/tokens.json"
    old_version_weth_address = "0x88699e7fee2da0462981a08a15a3b940304cc516"
...
[data_source]
    type = "motan"
    [data_source.motan_client]
        client_id="miner-client"
        conf_file="/opt/loopring/miner/config/motan_client.yaml"

#zk内网ip地址
[zk_lock]
    zk_servers = "xx.xx.xx.xx:2181,xx.xx.xx.xx:2181,xx.xx.xx.xx:2181"
...
#kafka内网ip地址
[kafka]
    brokers = ["xx.xx.xx.xx:9092","xx.xx.xx.xx:9092","xx.xx.xx.xx:9092"]

[cloudwatch]
    enabled = false
    region = ""
```

> cloudwatch如果设置`enabled`为true，请参考[ec2](new_ec2_cn.md)部署鉴权文件，region取值请参考[aws doc](https://docs.aws.amazon.com/zh_cn/AWSEC2/latest/UserGuide/using-regions-availability-zones.html)

* motan_client.yaml

在`Loopring/miner/config/motan_client.yaml`的基础上进行如下必要的修改
```
log_dir: "/var/log/miner"
...
#设置zookeeper内网ip地址
  zk-registry:
    protocol: zookeeper
    host: xx.xx.xx.xx
    port: 2181
```
* tokens.json

在[tokens.json](tokens_main.md)的基础上根据实际需要进行必要的裁剪

### 配置EC2实例
* 部署配置文件

在EC2实例执行脚本
```
sudo mkdir -p /opt/loopring/miner
sudo chown -R ubuntu:ubuntu /opt/loopring
cd /opt/loopring/miner 
mkdir bin/ config/ src/
```
上传本地配置文件
```
scp -i xx.pem miner.toml ubuntu@x.x.x.x:/opt/loopring/miner/config
scp -i xx.pem motan_client.yaml ubuntu@x.x.x.x:/opt/loopring/miner/config
scp -i xx.pem tokens.json ubuntu@x.x.x.x:/opt/loopring/miner/config
```
* 部署keystore

将接受矿工费用的eth地址对应keystore文件复制到目录 `/opt/loopring/miner/config/keystore`

### 部署deamontools配置

和其他两个服务不同，因为miner启动脚本包含本地参数，因此不能放在自动启动脚本中进行每次覆盖部署，在第一次部署前需要手动配置启动脚本【待优化】

在EC2实例执行下面脚本创建临时目录
```
mkdir -p /tmp/svc/log
```
在`Loopring/miner/bin/svc/run`的基础上修改svc/run
```
#修改unlocks为矿工费用接受地址，password为该地址对应口令，这里的地址应该和上面配置的keystore地址一致
exec setuidgid ubuntu $WORK_DIR/bin/miner --unlocks=0x1111111111111111111111111111 --passwords xxxx --config $WORK_DIR/config/miner.toml 2>&1
```
上传配置脚本
```
scp -i xx.pem svc/run ubuntu@x.x.x.x:/tmp/svc
scp -i xx.pem svc/log/run ubuntu@x.x.x.x:/tmp/svc/log
```
部署配置文件
```
sudo mkdir -p /etc/service/miner
sudo cp -rf /tmp/svc/* /etc/service/miner
sudo chmod -R 755 /etc/service/miner
rm -rf /tmp/svc
```
## 部署
通过CodeDeploy进行配置，详细步骤参考[接入CodeDeloy](codedeploy_cn.md)

## 服务日志

## miner业务日志
`/var/log/miner/zap.log`

## motan-rpc日志
`/var/log/miner/miner.INFO`

## stdout
`/var/log/svc/miner/current`

## 启停
通过CodeDeploy的方式部署会为服务添加daemontools支持，也就是服务如果意外终止，会自动启动，所以不能通过kill的方式手动停止

### 启动
`sudo svc -u /etc/service/miner`

### 停止
`sudo svc -d /etc/service/miner`
