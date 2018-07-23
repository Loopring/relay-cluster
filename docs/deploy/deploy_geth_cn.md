# 部署ethnode
我们选择部署官方的go-ethereum(geth)版本的eth服务，作为接入eth网络的入口

## 启动新实例
建议部署两台以上eth节点，避免单台故障，或者由于节点时延或者运维原因导致服务不可用

由于需要部署geth全节点，相对会比较耗费资源，需要选择8core 32g及以上配置的实例，可以参考[EC2实例](new_ec2_cn.md)

启动实例后需要关联`ethnode-SecurityGroup`安全组。如果未创建该安全组，请参考[aws安全组](security_group_cn.md)关于`ethnode-SecurityGroup`安全组的说明，创建后再关联

## 部署
具体部署操作，参考eth官方文档[go-ethereum for ubuntu](https://github.com/ethereum/go-ethereum/wiki/Installation-Instructions-for-Ubuntu)

## 启动
```
sudo mkdir -p /data/ethereum
sudo chown -R ubuntu:ubuntu /data/ethereum
#启动脚本请参考如下进行配置，其中ip地址为本机内网地址
geth --datadir /data/ethereum --fast --cache=1024 --rpc --rpcaddr x.x.x.x --rpcport 8545 --rpccorsdomain *
```
