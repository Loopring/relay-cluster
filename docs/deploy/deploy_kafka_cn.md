# 部署kafka集群

> kafka依赖zookeeper，所以应先部署zookeeper

kafka是extractor和relay-cluster之间的消息通信服务

## 选择kafka部署场景

测试场景下仅需部署单实例伪集群即可，简便快捷

生产场景下以部署3个节点为例，保证可用性

### 生产环境部署

申请3台EC2实例，参考启动aws [EC2实例](new_ec2_cn.md)，并且关联`kafka-SecurityGroup`安全组

> 如果还没创建，请参考[aws安全组](security_group_cn.md)关于`kafka-SecurityGroup`部分的说明，创建后再关联

每台服务器都需要按以下步骤进行部署

```
sudo apt update
sudo apt -y install openjdk-8-jre-headless

sudo mkdir /opt/loopring/data/kafka-logs
sudo chown -R ubuntu:ubuntu /opt/loopring

cd /opt/loopring
wget http://apache.mirrors.lucidnetworks.net/kafka/0.11.0.2/kafka_2.12-0.11.0.2.tgz
tar xzf kafka_2.12-0.11.0.2.tgz
cd kafka_2.12-0.11.0.2
```
修改config/server.properties配置项

`vim /opt/loopring/kafka_2.12-0.11.0.2/config/server.properties`
```
#不同节点设置不同的id
broker.id=0
# 修改为当前节点的内网ip
listeners=PLAINTEXT://x.x.x.x:9092

log.dirs=/opt/loopring/data/kafka-logs
offsets.topic.replication.factor=3
min.insync.replicas=1
transaction.state.log.replication.factor=2
log.flush.interval.messages=300
log.flush.interval.ms=300
log.flush.scheduler.interval.ms=300
log.flush.start.offset.checkpoint.interval.ms=2000
log.retention.hours=168
#修改为3台zookeeper节点的内网ip，多个节点间使用逗号分隔
zookeeper.connect=x.x.x.x:2181,x.x.x.x:2181,x.x.x.x:2181
default.replication.factor=3
```
#### 启停

* ##### 启动
```
nohup /opt/loopring/kafka_2.12-0.11.0.2/bin/kafka-server-start.sh /opt/loopring/kafka_2.12-0.11.0.2/config/server.properties &
```
###### 确认服务正常启动
```
tail -f /opt/loopring/kafka_2.12-0.11.0.2/nohup.out
telnet 本实列内网ip 9092
```

* ##### 终止
`/opt/loopring/kafka_2.12-0.11.0.2/bin/kafka-server-stop.sh`

#### 日志
`/opt/loopring/kafka_2.12-0.11.0.2/logs`

### 测试环境部署

申请1台EC2实例，参考[启动aws EC2实例](new_ec2_cn.md)，并且关联`kafka-SecurityGroup`安全组

> 如果未创建该安全组，请参考[aws安全组](security_group_cn.md)关于`kafka-SecurityGroup`安全组的说明，创建后再关联

```
sudo apt update
sudo apt -y install openjdk-8-jre-headless

sudo mkdir -p /opt/loopring/data/kafka-logs
sudo mkdir -p /opt/loopring/data/kafka-logs2
sudo mkdir -p /opt/loopring/data/kafka-logs3
sudo chown -R ubuntu:ubuntu /opt/loopring

cd /opt/loopring
wget http://apache.mirrors.lucidnetworks.net/kafka/0.11.0.2/kafka_2.12-0.11.0.2.tgz
tar xzf kafka_2.12-0.11.0.2.tgz
cd kafka_2.12-0.11.0.2
cp config/server.properties config/server.properties
cp config/server.properties config/server.properties2
cp config/server.properties config/server.properties3
```
修改config/server.properties

`vim /opt/loopring/kafka_2.12-0.11.0.2/config/server.properties`
```
broker.id=0
# 修改为本实列的内网ip
listeners=PLAINTEXT://x.x.x.x:9092

log.dirs=/opt/loopring/data/kafka-logs
offsets.topic.replication.factor=3
min.insync.replicas=1
transaction.state.log.replication.factor=2
log.flush.interval.messages=300
log.flush.interval.ms=300
log.flush.scheduler.interval.ms=300
log.flush.start.offset.checkpoint.interval.ms=2000
log.retention.hours=168
#修改为zookeeper实列的内网ip
zookeeper.connect=x.x.x.x:2181
default.replication.factor=3
```

修改config/server.properties2

`vim /opt/loopring/kafka_2.12-0.11.0.2/config/server.properties2`
```
broker.id=1
# 修改为本实列的内网ip
listeners=PLAINTEXT://x.x.x.x:9093

log.dirs=/opt/loopring/data/kafka-logs2
offsets.topic.replication.factor=3
min.insync.replicas=1
transaction.state.log.replication.factor=2
log.flush.interval.messages=300
log.flush.interval.ms=300
log.flush.scheduler.interval.ms=300
log.flush.start.offset.checkpoint.interval.ms=2000
log.retention.hours=168
#修改为zookeeper实列的内网ip
zookeeper.connect=x.x.x.x:2181
default.replication.factor=3
```

修改config/server.properties3

`vim /opt/loopring/kafka_2.12-0.11.0.2/config/server.properties3`
```
broker.id=2
# 修改为本实列的内网ip
listeners=PLAINTEXT://x.x.x.x:9094

log.dirs=/opt/loopring/data/kafka-logs3
offsets.topic.replication.factor=3
min.insync.replicas=1
transaction.state.log.replication.factor=2
log.flush.interval.messages=300
log.flush.interval.ms=300
log.flush.scheduler.interval.ms=300
log.flush.start.offset.checkpoint.interval.ms=2000
log.retention.hours=168
#修改为zookeeper实列的内网ip
zookeeper.connect=x.x.x.x:2181
default.replication.factor=3
```

若采用免费aws实例做测试，由于内存不足，启动后会立即exit，并在日志报错“Cannot allocate memory”，进行如下修改即可

`sudo vim /opt/loopring/kafka_2.12-0.11.0.2/bin/kafka-server-start.sh`

`export KAFKA_HEAP_OPTS="-Xmx256M -Xms256M"`

#### 启停

* ##### 启动
```
nohup /opt/loopring/kafka_2.12-0.11.0.2/bin/kafka-server-start.sh /opt/loopring/kafka_2.12-0.11.0.2/config/server.properties &
nohup /opt/loopring/kafka_2.12-0.11.0.2/bin/kafka-server-start.sh /opt/loopring/kafka_2.12-0.11.0.2/config/server.properties2 &
nohup /opt/loopring/kafka_2.12-0.11.0.2/bin/kafka-server-start.sh /opt/loopring/kafka_2.12-0.11.0.2/config/server.properties3 &
```

* ###### 确认服务正常启动
```
tail -f /opt/loopring/kafka_2.12-0.11.0.2/nohup.out
telnet 本实列内网ip 9092
```

* ##### 终止
```
/opt/loopring/kafka_2.12-0.11.0.2/bin/kafka-server-stop.sh /opt/loopring/kafka_2.12-0.11.0.2/config/server.properties
/opt/loopring/kafka_2.12-0.11.0.2/bin/kafka-server-stop.sh /opt/loopring/kafka_2.12-0.11.0.2/config/server.properties2
/opt/loopring/kafka_2.12-0.11.0.2/bin/kafka-server-stop.sh /opt/loopring/kafka_2.12-0.11.0.2/config/server.properties3
```
#### 日志
`/opt/loopring/kafka_2.12-0.11.0.2/logs`
