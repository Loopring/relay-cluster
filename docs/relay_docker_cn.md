# Loopring Relay Docker部署文档

为方便合作伙伴部署分布式中继，我们将提供以下几个镜像(除relay-cluster外，其余镜像正在制作中)：

* relay-cluster镜像
* miner镜像
* extractor镜像
* zookeeper&kafka
* mysql & redis 请自行下载官方镜像

各个微服务和中间件的配置文件，都需要独立于镜像配置，官方会及时更新配置文件说明，增加自解释，这部分工作会和docker化一起完成，然后提供给合作伙伴使用，尽请期待！

目前官方提供relay-cluster镜像，最新版本：v1.5.0

## 部署relay-cluster
* 获取docker镜像
```bash
docker pull loopring/relay-cluster
```
* 创建log&config目录
```bash
mkdir your_log_path your_config_path
```
* 配置relay.toml文件，[参考](https://loopring.github.io/relay-cluster/deploy/deploy_relay_cluster_cn.html#%E5%88%9B%E5%BB%BA%E9%85%8D%E7%BD%AE%E6%96%87%E4%BB%B6)
* telnet测试mysql,redis,zk,kafka,ethereum,motan rpc相关端口能否连接

* 运行
运行时需要挂载logs&config目录,并指定config文件
```bash
docker run --name relay -idt -v your_log_path:/opt/loopring/relay/logs -v your_config_path:/opt/loopring/relay/config loopring/relay-cluster:latest --config=/opt/loopring/relay/config/relay.toml /bin/bash
```

## 历史版本

| 版本号         | 描述         |
|--------------|------------|
| v1.5.0| release初始版本|

