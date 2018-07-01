# 部署Aws LoadBalancer

ALB（Application Load Balancer）是aws提供的负载局衡器，ALB具有比较高的可用性。

通过把一组服务器映射到目标组，然后关联ALB，我们可以配置请求url和后端服务的映射关系，从而转发请求到正确的后端接口

## 配置目标组
目标组将一组服务器的某个端口汇总为一个组，该组作为ALB的的请求转发目标

### 创建第一个目标组
【EC2-负载平衡-目标群组】，点击【创建目标组】

* 基本信息

输入【目标组名称】输入`relayClusterGroup`，【协议】为HTTP，端口8083，【目标类型】选择instance，VPC选择默认，

* 运行状况检查

【协议】选择HTTP，路径默认`/`

默认情况，健康检查端口，和目标组的端口一致，如果需要修改健康检查端口，【运行状况检查-高级运行状况检查-端口】，选择【覆盖】输入需新的端口号

点击【创建】

### 其他目标群组
参考上面的relayCluster目标组，创建其他目标组

| 目标组名称         		 | 协议 |   端口 | 运行状况检查端口|
|------------------------|-----|--------|---------|
| relayClusterGroup      | HTTP | 8083  |默认8083    |
| relayClusterWebSockets | HTTP | 8087  |覆盖为8083 |
| ethGroup               | HTTP | 8545  |默认8545 |

### 配置目标

目标组中选择某个目标组，选择下面的【目标】，点击编辑，在弹出的对话框中的下面【实例】部分，选择正确的已经部署了服务的实例，点击【添加到已注册】，选择【保存】

在【目标】tab，能够看到新添加实例的状态，如果目标组没有和ALB关联，则应该显示为【unused】

目标组和实例部署服务对应关系如下


| 目标组名称         		 | 实例部署服务类型 |
|------------------------|---------------|
| relayClusterGroup      | relay-cluster |
| relayClusterWebSockets | relay-cluster |
| ethGroup               | ethnode |

## 配置ALB

### 创建ALB
【EC2-负载平衡-负载均衡器】，点击【创建负载均衡器】，选择【应用程序负载均衡器-创建】

* 步骤 1: 配置负载均衡器
【名称】输入relayCluster，【模式】选择【面向 internet】

【侦听器】，添加HTTP，端口选择默认的80。如果在证书颁发机构申请了证书，请额外添加HTTPS侦听器

【可用区】选择你部署了relay-cluster的所有可用区，至少选择两个区

* 步骤 2: 配置安全设置
如果已经在证书颁发机构申请了https的证书，请在这里配置该证书，安全策略选择`ELBSecurityPolicy-2016-08`

* 步骤 3: 配置安全组
选择【选择一个现有的安全组】，选择`alb-SecurityGroup`安全组。如果未创建该安全组，请参考[aws安全组](security_group_cn.md)关于`alb-SecurityGroup`安全组的说明，创建后再关联

* 步骤 4: 配置路由
【目标组】选择【现有目标组】，这里选择`relayClusterGroup`

* 步骤 5: 注册目标
简单确认

* 步骤 6: 审核
点击【创建】

### 配置路由
选择刚刚创建的【relayCluster】负载均衡器，在【侦听器】tab中点击【查看/编辑规则】

在页面中点击【编辑】和【添加】图标，并且【添加条件】选择【路径为..】，逐个添加以下规则，并【保存】


| 路径         | 转发至 |
|-------------|----------|
| /rpc/*      | relayClusterGroup |
| /rpc/v2/*   | relayClusterGroup |
| /eth        | ethGroup |
| /socket.io/*| relayClusterGroup |

## 确认部署状态
【EC2-负载平衡-目标群组】，依次点击查看前面创建的三个目标组的【目标】tab，确认【状态】列为【healthy】，如果提示【unhealthy】，鼠标移到后面的叹号图标，查看提示原因，然后进行解决

## 连接ALB
ALB配置完成后，就可以通过ALB来请求后端服务。

点击创建的`relayCluster`负载均衡器，在【描述】tab中，找到【DNS 名称】对应的值，该域名加上之前配置的路由url就可以访问relay-cluster和ethnode的相关接口

## 查看ALB访问日志
有时需要通过ALB访问日志分析问题，默认该日志功能是禁用的，需要通过配置来打开日志记录功能。

具体操作可参考[aws doc](https://docs.aws.amazon.com/zh_cn/elasticloadbalancing/latest/application/load-balancer-access-logs.html)

