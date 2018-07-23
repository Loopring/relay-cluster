# 部署kafka

> kafka依赖zookeeper，所以需要先部署zookeeper

kafka是extractor和relay-cluster之间的消息通信服务

## 部署
建议部署3个以上的节点来保证可用性，下面以3个节点的kafka集群为例

### 申请EC2实例并关联安全组
申请3台EC2服务器，参考[EC2实例](new_ec2_cn.md)

关联`kafka-SecurityGroup`安全组。
> 如果未创建该安全组，请参考[aws安全组](security_group_cn.md)关于`kafka-SecurityGroup`安全组的说明，创建后再关联

### 部署kafka
```
#如果没有部署jre，需要执行下面两步操作
sudo apt update
sudo apt -y install openjdk-9-jre-headless

sudo mkdir /opt/loopring
sudo chown -R ubuntu:ubuntu /opt/loopring

cd /opt/loopring
wget http://apache.mirrors.lucidnetworks.net/kafka/0.11.0.2/kafka_2.12-0.11.0.2.tgz
tar xzf kafka_2.12-0.11.0.2.tgz
cd kafka_2.12-0.11.0.2/
sudo mkdir -p /opt/loopring/data/kafka-logs
sudo chown -R ubuntu:ubuntu /opt/loopring/data/kafka-logs
```
修改config/server.properties配置项

`vim config/server.properties`
```
#不同节点设置不同的id
broker.id=0
# 修改ip地址为当前节点的内网ip地址
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
#修改为zookeeper节点的内网ip和端口，多个节点间使用逗号分隔
zookeeper.connect=xx.xx.xx.xx:2181,xx.xx.xx.xx:2181,xx.xx.xx.xx:2181
default.replication.factor=3
```

## 启停

### 启动
`nohup bin/kafka-server-start.sh config/server.properties &`

### 终止
`bin/kafka-server-stop.sh`

## 日志
`/opt/loopring/kafka_2.12-0.11.0.2/logs`
