# 部署kafka-manager
kafka-manager是yahoo开源的kafka管理工具，可以用来查看集群内的节点和topic状态，以及一些性能指标

## 申请EC2实例并关联安全组
申请1台EC2服务器，参考[EC2实例](new_ec2_cn.md)

关联`kafkaManger-SecurityGroup`安全组。
> 如果未创建该安全组，请参考[aws安全组](security_group_cn.md)关于`kafkaManger-SecurityGroup`安全组的说明，创建后再关联

## 部署
```
sudo mkdir /opt/loopring
sudo chown -R ubuntu:ubuntu /opt/loopring
cd /opt/loopring
sudo apt update
sudo apt install openjdk-8-jdk-headless -y
wget https://downloads.lightbend.com/scala/2.12.6/scala-2.12.6.deb
sudo dpkg -i scala-2.12.6.deb 
wget https://dl.bintray.com/sbt/debian/sbt-1.1.5.deb
sudo dpkg -i sbt-1.1.5.deb
git clone https://github.com/yahoo/kafka-manager.git
cd kafka-manager
sbt clean dist
mv ./target/universal/kafka-manager-1.3.3.17.zip ./
unzip kafka-manager-1.3.3.17.zip
cd kafka-manager-1.3.3.17
```
修改conf/application.conf

`vim /opt/loopring/kafka-manager/kafka-manager-1.3.3.17/conf/application.conf`
```
#修改为zookeeper节点的内网ip和端口，多个节点间使用逗号分隔

kafka-manager.zkhosts="xx.xx.xx.xx:2181,xx.xx.xx.xx:2181,xx.xx.xx.xx:2181"

#配置认证选项

basicAuthentication.enabled=true
basicAuthentication.username="admin"
basicAuthentication.password="admin"
```
## 启停

### 启动
`nohup ./bin/kafka-manager &`

### 终止
`pkill -f "play.core.server.ProdServerStart"`

## 日志
`/opt/loopring/kafka-manager/kafka-manager-1.3.3.17/nohup.out`

## 访问
浏览器访问 `http://外网ip:9000`
