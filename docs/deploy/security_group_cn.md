# 配置安全组

> 因为默认安全组会拒绝除了22端口之外的流量进入，为了能够正常对外提供服务，我们需要配置安全组

aws安全组是通过设置准入规则来拦截非法流量访问Ec2服务器，以此提高服务器的网络安全性，详情可以参考[aws安全组](https://docs.aws.amazon.com/zh_cn/AWSEC2/latest/UserGuide/using-network-security.html)

## 配置入口
【EC2-网络与安全-安全组】是编辑和查看安全组的入口

## 默认安全组
如果未自建安全组，请选择新建`launch-wizard-1`的默认安全组，该安全组开通ssh端口22，允许通过ssh方式访问

## 自定义安全组
当实例部署的服务需要提供额外的端口供外界访问，或者需要访问其他实例的特定端口时候，就需要新建自定义安全组

### 只需要访问外部实例
如果只需要访问其他实例的端口，本身不需要提供端口供外界访问，只需要新建服务组，不需要额外配置规则

### 允许外部访问本实例
设置允许的入站规则，通常我们部署的服务都是通过TCP协议来对外提供服务的，所以下面只介绍入站TCP的设置

*  允许特定ip

只允许特定的ip来访问，通常只用在临时的特例场景

通常我们可以会允许一个特定的ipv4地址来访问，做法是在来源中配置 `ip+/32` ，也就是类似 `203.0.113.1/32`

对于服务之间的访问不要使用该规则，因为一旦服务做了扩容或者迁移，还需要手动进行修改

*  允许一组实例

这是常用的做法，允许绑定了特定安全组A的一组实例访问当前编辑安全组B会绑定的一组实例。

> 比如已经有名称为 ethnode-SecurityGroup 的安全组绑定了部署eth节点的一组服务器，这时候我们需要允许relay-cluster能够访问这组eth节点的8545端口。我们会新建relayCluster-SecurityGroup安全组，并得到组ID，假设为sg-123456。然后编辑ethnode-SecurityGroup，添加新规则允许8545端口的来源为sg-123456。这样我们把relayCluster-SecurityGroup关联到relay-cluster部署的节点，eth的8545端口就对relay-cluster开放了。后续eth或者relay-cluster扩容，该规则都会自动生效

> 设置安全组为来源的场景，需要在来源字段输入【组ID】，而非组名

* 允许所有流量

对于需要向公网提供入口的类似http服务，需要允许公网的所有ip都能访问
这时候只需要配置两个规则

|类型         | 协议 | 端口范围| 来源     |
|------------|-----|--------|---------|
|自定义 TCP 规则| TCP | 80     | 0.0.0.0/0|
|自定义 TCP 规则| TCP | 80     | ::/0    |

## 修改实例关联安全组
【EC2-实例-实例-操作-联网-更改安全组】

## 安全组叠加
对于单个服务器可能部署多个服务的情况，建议针对每个服务设置安全组，然后编辑该服务器关联所需要的多个安全组，这样可以能提高安全组配置的灵活性，也便于命名。

## relay-cluster涉及全部安全组配置
安全组独立于EC2实例，可以预先创建，建立他们之间的关联关系，待服务部署时只需要关联就可以
* alb-securityGroup

|类型         | 协议 | 端口范围| 来源     |
|------------|-----|--------|---------|
| HTTP | TCP | 80  | 0.0.0.0/0,::/0|
| 自定义 TCP 规则 | TCP | 443  | 0.0.0.0/0,::/0|

* miner-securityGroup

|类型         | 协议 | 端口范围| 来源     |
|------------|-----|--------|---------|
| SSH | TCP | 22    | 	0.0.0.0/0    |

* relayCluster-securityGroup

|类型         | 协议 | 端口范围| 来源     |
|------------|-----|--------|---------|
| SSH | TCP | 22    | 	0.0.0.0/0    |
| 自定义 TCP 规则 | TCP | 8087    |alb-securityGroup|
| 自定义 TCP 规则 | TCP | 8083    |alb-securityGroup|
| 自定义 TCP 规则 | TCP | 8100    |miner-SecurityGroup|

* extractor-securityGroup

|类型         | 协议 | 端口范围| 来源     |
|------------|-----|--------|---------|
| SSH | TCP | 22    | 	0.0.0.0/0    |

* mysql-securityGroup

|类型         | 协议 | 端口范围| 来源     |
|------------|-----|--------|---------|
| SSH | TCP | 22    | 	0.0.0.0/0    |
| 自定义 TCP 规则 | TCP | 3306    |relayCluster-securityGroup|
| 自定义 TCP 规则 | TCP | 3306    |miner-SecurityGroup|
| 自定义 TCP 规则 | TCP | 3306    |extractor-SecurityGroup|

* redis-securityGroup

|类型         | 协议 | 端口范围| 来源     |
|------------|-----|--------|---------|
| SSH | TCP | 22    | 	0.0.0.0/0    |
| 自定义 TCP 规则 | TCP | 6379    |relayCluster-securityGroup|
| 自定义 TCP 规则 | TCP | 6379    |miner-SecurityGroup|
| 自定义 TCP 规则 | TCP | 6379    |extractor-SecurityGroup|

* ethnode-securityGroup

|类型         | 协议 | 端口范围| 来源     |
|------------|-----|--------|---------|
| SSH | TCP | 22    | 	0.0.0.0/0    |
| 自定义 TCP 规则 | TCP | 8545    |alb-SecurityGroup|
| 自定义 TCP 规则 | TCP | 8545    |relayCluster-securityGroup|
| 自定义 TCP 规则 | TCP | 8545    |miner-SecurityGroup|
| 自定义 TCP 规则 | TCP | 8545    |extractor-SecurityGroup|

* kafkaManager-securityGroup

|类型         | 协议 | 端口范围| 来源     |
|------------|-----|--------|---------|
| SSH | TCP | 22    | 	0.0.0.0/0    |
| 自定义 TCP 规则 | TCP | 9000    |yourSpecialIp|

* kafka-securityGroup

|类型         | 协议 | 端口范围| 来源     |
|------------|-----|--------|---------|
| SSH | TCP | 22    | 	0.0.0.0/0    |
| 自定义 TCP 规则 | TCP | 9092    |relayCluster-securityGroup|
| 自定义 TCP 规则 | TCP | 9092    |miner-SecurityGroup|
| 自定义 TCP 规则 | TCP | 9092    |extractor-SecurityGroup|
| 自定义 TCP 规则 | TCP | 9092    |kafka-SecurityGroup|

* motanManager-securityGroup

|类型         | 协议 | 端口范围| 来源     |
|------------|-----|--------|---------|
| SSH | TCP | 22    | 	0.0.0.0/0    |
| 自定义 TCP 规则 | TCP | 8080    |yourSpecialIp|

* zookeeperBrowser-securityGroup

|类型         | 协议 | 端口范围| 来源     |
|------------|-----|--------|---------|
| SSH | TCP | 22      | 	0.0.0.0/0    |
| 自定义 TCP 规则 | TCP | 3000    |yourSpecialIp|

* zookeeper-securityGroup

|类型         | 协议 | 端口范围| 来源     |
|------------|-----|--------|---------|
| SSH | TCP | 22    | 	0.0.0.0/0    |
| 自定义 TCP 规则 | TCP | 2181    |relayCluster-securityGroup|
| 自定义 TCP 规则 | TCP | 2181    |miner-SecurityGroup|
| 自定义 TCP 规则 | TCP | 2181    |extractor-SecurityGroup|
| 自定义 TCP 规则 | TCP | 2181    |kafka-SecurityGroup|
| 自定义 TCP 规则 | TCP | 2181    |motanManager-SecurityGroup|
| 自定义 TCP 规则 | TCP | 2181    |zookeeperBrowser-SecurityGroup|
| 自定义 TCP 规则 | TCP | 2181    |kafkaManager-SecurityGroup|
| 自定义 TCP 规则 | TCP | 2888    |zookeeper-SecurityGroup|
| 自定义 TCP 规则 | TCP | 3888    |zookeeper-SecurityGroup|
