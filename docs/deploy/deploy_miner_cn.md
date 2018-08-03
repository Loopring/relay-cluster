# 部署miner

## 初始化环境

申请EC2实例，配置时添加对CodeDeploy的支持，参考启动aws [EC2实例](new_ec2_cn.md)，并且关联`miner-SecurityGroup`安全组

> 如果还没创建，请参考[aws安全组](security_group_cn.md)关于`miner-SecurityGroup`部分的说明，创建后再关联

### 部署配置文件
目前miner是通过静态文件来实现基本配置的，所以需要先在本地修改好配置文件，再上传到待部署的miner服务器

> 该操作仅首次部署有必要，后续会利用该静态配置文件直接启动服务【待优化】

#### 创建配置文件
* miner.toml

下载`https://github.com/Loopring/miner/blob/master/config/miner.toml`到本地，并在此基础上进行如下修改
```
output_paths = ["/var/log/miner/zap.log"]
error_output_paths = ["/var/log/miner/err.log"]
...
[mysql]
    hostname = "x.x.x.x"
    port = "3306"
    user = "xxx"
    password = "xxx"
...
[redis]
    host = "x.x.x.x"
    port = "6379"
#ethnode实例内网ip，多个实例之间用逗号进行分割
[accessor]
    raw_urls = ["http://x.x.x.x:8545", "http://x.x.x.x:8545"]
#eth主网合约配置，如果非主网，要联系开源人员获取最新的测试配置
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

#zookeeper内网ip
[zk_lock]
    zk_servers = "x.x.x.x:2181,x.x.x.x:2181,x.x.x.x:2181"
    #测试场景下修改为：zk_servers = "x.x.x.x:2181"
...
#kafka内网ip
[kafka]
    brokers = ["x.x.x.x:9092","x.x.x.x:9092","x.x.x.x:9092"]
    #测试场景下修改为：brokers = ["x.x.x.x:9092"]
[cloud_watch]
    enabled = false
    region = ""
```

> cloudwatch如果设置`enabled`为true，请参考[ec2](new_ec2_cn.md)部署鉴权文件，region取值请参考[aws doc](https://docs.aws.amazon.com/zh_cn/AWSEC2/latest/UserGuide/using-regions-availability-zones.html)

* motan_client.yaml

下载`https://github.com/Loopring/miner/blob/master/config/motan_client.yaml`到本地，并在此基础上进行如下修改
```
log_dir: "/var/log/miner"
...
#设置zookeeper内网ip
  zk-registry:
    protocol: zookeeper
    host: x.x.x.x
    port: 2181
```
* tokens.json

在[tokens.json](tokens_main.md)的基础上根据实际需要进行必要的裁剪

#### 配置EC2实例
* 部署配置文件

在EC2实例执行脚本
```
sudo mkdir -p /opt/loopring/miner/bin
sudo mkdir -p /opt/loopring/miner/config
sudo mkdir -p /opt/loopring/miner/src
sudo mkdir -p /opt/loopring/miner/config/keystore
sudo chown -R ubuntu:ubuntu /opt/loopring
```
上传本地配置文件
```
scp -i xx.pem miner.toml ubuntu@x.x.x.x:/opt/loopring/miner/config
scp -i xx.pem motan_client.yaml ubuntu@x.x.x.x:/opt/loopring/miner/config
scp -i xx.pem tokens.json ubuntu@x.x.x.x:/opt/loopring/miner/config
```
* 部署keystore

通过私钥生成keystore文件并自动导入到miner中，其中--private-key填私钥，--passphrase填该keystore的密码
```
cd /opt/loopring/miner
bin/miner account import --datadir config/keystore --private-key xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx --passphrase xxxxxx
```

#### 部署deamontools配置

和其他两个服务不同，因为miner启动脚本包含本地参数，因此不能放在自动启动脚本中进行每次覆盖部署，在第一次部署前需要手动配置启动脚本【待优化】

在EC2实例执行下面命令创建临时目录

`mkdir -p /tmp/svc/log`

在`https://github.com/Loopring/miner/blob/master/bin/svc/run`的基础上修改`run`文件

```
chmod +x run
sudo vim run

#修改unlocks为矿工费用接受地址，password为该地址对应口令，这里的地址应该和上面配置的keystore地址一致
exec setuidgid ubuntu $WORK_DIR/bin/miner --unlocks ‘0x1111111111111111111111111111’ --passwords ‘xxxxxxxx’ --config $WORK_DIR/config/miner.toml 2>&1
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
## 部署miner
通过CodeDeploy进行配置，详细步骤参考[接入CodeDeloy](codedeploy_cn.md)

## 启停
通过CodeDeploy的方式部署会为服务添加daemontools支持，也就是服务如果意外终止，会自动启动，所以不能通过kill的方式手动停止

* ### 启动
`sudo svc -u /etc/service/miner`

* ### 终止
`sudo svc -d /etc/service/miner`

## 日志

* ## miner业务日志
`/var/log/miner/zap.log`

* ## motan-rpc日志
`/var/log/miner/miner.INFO`

* ## stdout
`/var/log/svc/miner/current`
