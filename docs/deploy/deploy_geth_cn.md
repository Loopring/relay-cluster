# 部署ethnode
部署官方的PPA(geth)版本，作为接入eth网络的入口

## 启动新实例
建议部署两台以上eth节点，避免因单台故障、节点延时或运维不当等等造成的服务不可用

由于需部署geth全节点，会相当的耗费资源，所以应选择8核_32g内存_300GB SSD及以上配置的实例，可参考[EC2实例](new_ec2_cn.md)


实例需关联`ethnode-SecurityGroup`安全组
> 若未创建该安全组，请务必参考[aws安全组](security_group_cn.md)关于`ethnode-SecurityGroup`安全组的说明，创建后再关联

## 部署
```
sudo apt-get -y install software-properties-common
sudo add-apt-repository -y ppa:ethereum/ethereum
sudo apt-get update
sudo apt-get -y install ethereum
```
## 启动
```
sudo mkdir -p /data/ethereum
sudo chown -R ubuntu:ubuntu /data/ethereum

#启动脚本请参考如下配置，x.x.x.x为本机内网ip地址
sudo geth --datadir /data/ethereum --fast --cache=1024 --rpc --rpcaddr x.x.x.x --rpcport 8545 --rpccorsdomain * &
```
