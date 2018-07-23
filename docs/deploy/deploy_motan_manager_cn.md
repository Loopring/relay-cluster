# 部署motan-manager

motan-manager是weibo motan-rpc的开源组件的一部分，可以用来查看注册到motan-rpc所在zookeeper的rpc服务，并可以进行简单的管理操作

## 申请EC2实例并关联安全组
申请1台EC2服务器，参考[EC2实例](new_ec2_cn.md)

关联`motanManger-SecurityGroup`安全组
> 如果未创建该安全组，请参考[aws安全组](security_group_cn.md)关于`motanManger-SecurityGroup`安全组的说明，创建后再关联

## 部署
```
#部署mysql并记下用户名口令
sudo apt update
sudo apt install mysql-server -y
sudo apt install maven -y
sudo apt install openjdk-8-jdk-headless -y
sudo mkdir -p /opt/loopring/
sudo chown -R ubuntu:ubuntu /opt/loopring/
cd /opt/loopring/
sudo git clone https://github.com/weibocom/motan.git
cd motan
mvn install -DskipTests
cd motan-manager
```

修改初始化sql文件，在“DROP TABLE...”这句前面插入以下命令来创建motan_manager db

`vim /opt/loopring/motan/motan-manager/src/main/resources/motan-manager.sql`

```
create database motan_manager;
use motan_manager;
```

修改配置文件

`vim /opt/loopring/motan/motan-manager/src/main/resources/application.properties`

```
jdbc_url=jdbc:mysql://127.0.0.1:3306/motan-manager?useUnicode=true&characterEncoding=UTF-8
#设置正确的数据库用户和密码
jdbc_username=root
jdbc_password=xxx
#配置motan-rpc对应的zookeeper地址
registry.url=xx.xx.xx.xx:2181
```

初始化motan_manager db

`mysql --host=localhost --port=3306 --user=root -p < src/main/resources/motan-manager.sql`

打jar包

`mvn package`

## 启停

### 启动
`nohup java -jar target/motan-manager.jar &`

### 终止
`pkill -f "motan-manager"`

## 日志
`/opt/loopring/motan/motan-manager/nohup.out`

## 访问
浏览器访问 `http://外网ip:8080`
