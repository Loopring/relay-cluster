# 部署zookeeper集群

zookeeper使用场景
* kafka集群配置中心
* motan-rpc配置中心
* relay-cluster分布式锁
* miner负载均衡

zookeeper需要进行集群部署来保证可用性，建议部署3个以上的奇数节点

## 选择zookeeper部署场景

测试场景下仅需部署单实例伪集群即可，简便快捷

生产场景下以部署3个节点为例，保证可用性

### 生产场景部署

申请3台EC2实例，参考启动aws [EC2实例](new_ec2_cn.md)，并且关联`zookeeper-SecurityGroup`安全组

> 如果还没创建，请参考[aws安全组](security_group_cn.md)关于`zookeeper-SecurityGroup`部分的说明，创建后再关联

每台服务器都需要按以下步骤进行部署

```
sudo apt update

sudo apt -y install openjdk-8-jre-headless

sudo mkdir /opt/loopring/data/zookeeper
sudo chown -R ubuntu:ubuntu /opt/loopring
cd /opt/loopring
wget http://mirrors.ocf.berkeley.edu/apache/zookeeper/zookeeper-3.4.10/zookeeper-3.4.10.tar.gz
tar xzf zookeeper-3.4.10.tar.gz
cd zookeeper-3.4.10/conf
cp zoo_sample.cfg zoo.cfg
```

修改以下配置项，依次填入3台zookeeper节点的内网ip

`vim /opt/loopring/zookeeper-3.4.10/conf/zoo.cfg`

```
dataDir=/opt/loopring/data/zookeeper
server.1=x.x.x.x:2888:3888
server.2=x.x.x.x:2888:3888
server.3=x.x.x.x:2888:3888
```

初始化myid，这里"n"在三台服务器的取值依次为1，2，3与上面zoo.cfg对应，每台服务器仅执行一次自身对应取值的命令

`echo "n" > /opt/loopring/data/zookeeper/myid`

#### 启停

* ##### 启动
```
/opt/loopring/zookeeper-3.4.10/bin/zkServer.sh start /opt/loopring/zookeeper-3.4.10/conf/zoo.cfg
```
###### 确认服务正常启动
```
tail -f zookeeper.out
telnet localhost 2181
```

* ##### 终止
```
/opt/loopring/zookeeper-3.4.10/bin/zkServer.sh start /opt/loopring/zookeeper-3.4.10/conf/zoo.cfg
```

#### 日志
`/opt/loopring/zookeeper-3.4.10/bin/zookeeper.out`

### 测试场景部署

申请1台EC2实例，参考[启动aws EC2实例](new_ec2_cn.md)，并且关联`zookeeper-SecurityGroup`安全组。
> 如果未创建该安全组，请参考[aws安全组](security_group_cn.md)关于`zookeeper-SecurityGroup`安全组的说明，创建后再关联

```
sudo apt update
sudo apt -y install openjdk-8-jre-headless

sudo mkdir -p /opt/loopring/data/zookeeper
sudo mkdir -p /opt/loopring/data/zookeeper2
sudo mkdir -p /opt/loopring/data/zookeeper3
sudo chown -R ubuntu:ubuntu /opt/loopring

cd /opt/loopring
wget http://mirrors.ocf.berkeley.edu/apache/zookeeper/zookeeper-3.4.10/zookeeper-3.4.10.tar.gz
tar xzf zookeeper-3.4.10.tar.gz
cd zookeeper-3.4.10
cp conf/zoo_sample.cfg conf/zoo.cfg
cp conf/zoo_sample.cfg conf/zoo2.cfg
cp conf/zoo_sample.cfg conf/zoo3.cfg
```

修改zoo.cfg

`vim /opt/loopring/zookeeper-3.4.10/conf/zoo.cfg`

```
dataDir=/opt/loopring/data/zookeeper
# 修改为本实例的内网ip
server.1=x.x.x.x:2888:3888
server.2=x.x.x.x:2889:3889
server.3=x.x.x.x:2890:3890
clientPort=2181
```

修改zoo2.cfg

`vim /opt/loopring/zookeeper-3.4.10/conf/zoo2.cfg`

```
dataDir=/opt/loopring/data/zookeeper2
# 修改为本实例的内网ip
server.1=x.x.x.x:2888:3888
server.2=x.x.x.x:2889:3889
server.3=x.x.x.x:2890:3890
clientPort=2182
```

修改zoo3.cfg

`vim /opt/loopring/zookeeper-3.4.10/conf/zoo3.cfg`

```
dataDir=/opt/loopring/data/zookeeper3
# 修改为本实例的内网ip
server.1=x.x.x.x:2888:3888
server.2=x.x.x.x:2889:3889
server.3=x.x.x.x:2890:3890
clientPort=2183
```

初始化myid

```
echo "1" > /opt/loopring/data/zookeeper/myid
echo "2" > /opt/loopring/data/zookeeper2/myid
echo "3" > /opt/loopring/data/zookeeper3/myid
```
#### 启停
* ##### 启动
```
/opt/loopring/zookeeper-3.4.10/bin/zkServer.sh start /opt/loopring/zookeeper-3.4.10/conf/zoo.cfg
/opt/loopring/zookeeper-3.4.10/bin/zkServer.sh start /opt/loopring/zookeeper-3.4.10/conf/zoo2.cfg
/opt/loopring/zookeeper-3.4.10/bin/zkServer.sh start /opt/loopring/zookeeper-3.4.10/conf/zoo3.cfg
```
* ###### 确认服务正常启动
```
tail -f /opt/loopring/zookeeper-3.4.10/zookeeper.out
telnet localhost 2181
```

* ##### 终止
```
/opt/loopring/zookeeper-3.4.10/bin/zkServer.sh stop /opt/loopring/zookeeper-3.4.10/conf/zoo.cfg
/opt/loopring/zookeeper-3.4.10/bin/zkServer.sh stop /opt/loopring/zookeeper-3.4.10/conf/zoo2.cfg
/opt/loopring/zookeeper-3.4.10/bin/zkServer.sh stop /opt/loopring/zookeeper-3.4.10/conf/zoo3.cfg
```
#### 日志

`/opt/loopring/zookeeper-3.4.10/zookeeper.out`
