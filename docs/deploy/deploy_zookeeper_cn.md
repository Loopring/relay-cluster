# 部署zookeeper集群

zookeeper使用场景
* kafka集群配置中心
* motan-rpc配置中心
* relay-cluster分布式锁
* miner负载均衡

zookeeper需要进行集群部署来保证可用性，建议部署3个以上的奇数节点。

## 部署
3个节点为例

### 申请EC2实例并关联安全组
申请3台EC2服务器，参考[EC2实例](new_ec2_cn.md)


关联`zookeeper-SecurityGroup`安全组。
> 如果未创建该安全组，请参考[aws安全组](security_group_cn.md)关于`zookeeper-SecurityGroup`安全组的说明，创建后再关联

### 部署
使用三台服务器的内网ip地址设置`zoo1~zoo3`三个host，便于后面的配置

`sudo vim /etc/hosts`

设置为
```
x.x.x.x zoo1
x.x.x.x zoo2
x.x.x.x zoo3
```

* 初始化zookeeper环境

```
#如果没有部署jre，需要先部署

sudo apt update
sudo apt -y install openjdk-8-jre-headless

sudo mkdir /opt/loopring
sudo chown -R ubuntu:ubuntu /opt/loopring

cd /opt/loopring
wget http://mirrors.ocf.berkeley.edu/apache/zookeeper/zookeeper-3.4.10/zookeeper-3.4.10.tar.gz
tar xzf zookeeper-3.4.10.tar.gz
cd zookeeper-3.4.10/conf
cp zoo_sample.cfg zoo.cfg
sudo mkdir -p /opt/loopring/data/zookeeper
```

* 修改并添加以下配置项，分别填入zookeeper服务器的内网ip

`vim /opt/loopring/zookeeper-3.4.10/conf/zoo.cfg`

```
dataDir=/opt/loopring/data/zookeeper
server.1=zoo1:2888:3888
server.2=zoo2:2888:3888
server.3=zoo3:2888:3888
```

初始化myid，这里"n"在三台服务器的取值依次为1，2，3，和上面zoo.conf一致，每台服务器仅执行自己对应取值的那条命令

```
echo "n" > /opt/loopring/data/zookeeper/myid
```

## 启停

### 启动
```
cd /opt/loopring/zookeeper-3.4.10/bin/
./zkServer.sh start
```
确认服务正常启动
```
tail -f zookeeper.out
telnet localhost 2181
```

### 终止
```
cd /opt/loopring/zookeeper-3.4.10/bin/
./zkServer.sh stop
```

## 日志
`/opt/loopring/zookeeper-3.4.10/bin/zookeeper.out`
