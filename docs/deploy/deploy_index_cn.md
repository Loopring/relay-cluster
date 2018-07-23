# Loopring部署文档

本文档是所有组件部署说明的索引，并依次按照组件的依赖顺序进行了详细的介绍，使用者也应该按照此顺序去阅读和部署。

## 依赖说明

> relay-cluster部分功能强依赖aws提供的相关服务，若采用其他云服务提供商，可能会造成部分功能不可用，或出现一些非预期的异常。

> relay-cluster及其依赖的extractor服务，都需要通过集群的方式进行部署来避免单点故障，虽然可以只部署单个节点，但是单节点的正常启动依然需要依赖于组成集群的基础组件，如zookeeper，kafka。

## 注册AWS

官方网站：https://aws.amazon.com/

## 通用

* [aws安全组](security_group_cn.md)

限制非预期的端口访问，提高服务的网络安全，本文档集中定义了所有组件的安全组配置，按照文档所列顺序依次配置即可。

* [EC2实例](new_ec2_cn.md)

启动合适性能的EC2实例，并可根据部署组件的需要，在启动向导中进行必要的配置，或者启动后进行配置。

## 存储及通信
* [ethnode](deploy_geth_cn.md)

relay-cluster通过和ethereum(geth)节点的交互来实现以太坊网络的接入。

* [mysql](deploy_mysql_cn.md)

是后端服务所操作数据的主要存储方式。

* [redis](deploy_redis_cn.md)

用来提高系统的存取速度，或者存放非关键的数据。

* [zookeeper](deploy_zookeeper_cn.md)

用来做系统的配置管理以及kafka的元数据存储。

* [kafka](deploy_kafka_cn.md)

实现服务间的异步通信，方便系统解耦和扩展。

## 服务
* [接入CodeDeploy](codedeploy_cn.md)

目前后端服务是通过aws CodeDeploy进行部署的，在实际开始部署前需要进行相关的配置。

* [extractor](deploy_extractor_cn.md)

解析eth网络交易，通过kafka将解析结果同步到relay-cluster和miner。

* [relay-cluster](deploy_relay_cluster_cn.md)

是后台服务的核心组件，对外提供jsonRpc接口，支持钱包和交易功能的接入，同时通过motan-rpc暴露接口给miner。

* [miner](deploy_miner_cn.md)

撮合服务，接收订单，发现环路，提交环路到以太坊网络。

## web接入

* [ALB](deploy_alb_cn.md)

目前通过aws ALB作为relay-cluster和ethnode的统一请求代理入口。

## 辅助管理系统【可选】
* [kafka-manager](deploy_kafka_manager_cn.md)

是一个开源的kafka集群浏览和管理系统。

* [node-zookeeper-browser](deploy_zk_browser_cn.md)

提供了一个web界面用来辅助查看和编辑kafka，方便对系统进行配制管理。

* [motan-manager](deploy_motan_manager_cn.md)

提供了web界面方便查看motan-rpc的启动状态，并能执行简单的配置操作。

## 监控和告警【可选】
* [cloudwatch](cloudwatch_cn.md)

cloudwatch可以实现指标的上报、查看和基于规则的报警，用来辅助发现和解决问题。

* [SNS](sns_cn.md)

SNS是aws的通知服务，可以通过接入SNS的API来进行直接的系统通知，通知方式包含短信，邮件。可以在系统的关键业务逻辑中插入该通知功能，方便运维人员及时发现故障点。
