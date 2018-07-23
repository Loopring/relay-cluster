# 启动aws EC2实例

## 启动新实例

登录 aws控制台 [https://console.aws.amazon.com/console/home] ，选择【实例-实例-启动实例】

在【步骤 1: 选择一个 Amazon 系统映像(AMI)】页面，选择ubuntu实例类型，因为后续所有部署都是基于ubuntu实例进行操作的。建议使用通用实例`Ubuntu Server 16.04 LTS (HVM), SSD Volume Type`，然后点击【选择】

在【步骤 2: 选择一个实例类型】页面选择合适的类型，如果做实验，可以选择免费的实例。

在【步骤 3: 配置实例详细信息】页面，【网络】选择默认VPC即可。如果有多个实例，建议分批创建，然后每批选择不同的【子网】，可以避免单aws机房故障导致服务不可用。对于IAM角色，可以不设置，后续根据需要进行配置。其他选择默认即可
> 如果需要支持通过CodeDeploy在该实例部署服务，这里需要额外配置IAM角色和初始化脚本，参考[支持codedeploy](#支持codedeploy)

在【步骤 4: 添加存储】页面，磁盘大小建议选择30G（若部署ethnode，须选择大于300G的实例），【卷类型】为【默认通用SSD】

在【步骤 5: 添加标签】，【添加标签】-【密钥：Name】-【值：自定义组件标记，如：extrator、relay_cluster或miner等】，该选项也可后续手动配置

在【步骤 6: 配置安全组】，选择【选择一个现有的安全组】，选择您将要部署的组件对应的安全组，然后点击【审核和启动】

在【步骤 7: 核查实例启动】，确认相关配置正确，点击【启动】，在启动弹出框中，会提示选择登录私钥，如果没有则选择创建新的，否则选择现有的即可。如果创建秘钥对，下载新秘钥然后进行保存，后续所有ssh操作都依赖这个绑定的xx.pem私钥文件，点击【启动实例】

点击【查看实例】进入EC2实例列表页面，能看到实例的启动状态

## 登录实例
修改xx.pem 的权限 `chmod 400 xx.pem`

对于【实例-实例-实例状态】为running的实例，执行`ssh -i xx.pem ubuntu@ipv4_add` 来登录实例，这里ipv4_add是该实例的【IPv4 公有 IP】

## 支持CodeDeploy

> 角色`CodeDeployEc2InstanceProfile`的创建，请参考[接入CodeDeploy](codedeploy_cn.md)

IAM角色选择`CodeDeployEc2InstanceProfile`，高级详细信息选择“以文本形式”，然后输入以下文本
```
#!/bin/bash
apt-get -y update
apt-get -y install ruby
apt-get -y install wget
cd /home/ubuntu
wget https://aws-codedeploy-ap-northeast-1.s3.amazonaws.com/latest/install
chmod +x ./install
./install auto
service codedeploy-agent stop
service codedeploy-agent start
```

## 部署aws sdk鉴权文件
通过aws sdk可以实现对aws相关服务的接入，目前用到的两个服务是cloudwatch和SNS(Simple Notification Service)两个功能。

> aws sdk会用到鉴权文件，如果打开上面两个服务的开关，需要在实例上部署该鉴权文件。如果不需要以上aws服务，在配置中将开关关闭并跳过下面配置即可。

### 创建鉴权信息
参考[aws doc](https://docs.aws.amazon.com/zh_cn/cli/latest/userguide/cli-chap-getting-started.html)

打开IAM控制台，点击【添加用户】

步骤1，输入用户名`sdkUser`，【访问类型】选择【编程访问】

步骤2，选择【直接附加现有策略】

	如果开通cloudwatch 服务，请添加`CloudWatchAgentServerPolicy`, `AmazonAPIGatewayPushToCloudWatchLogs`, `CloudWatchActionsEC2Access`三个策略

	如果开通SNS服务，请添加`AmazonSNSFullAccess`

步骤3，点击【创建用户】

步骤4，记录页面中的【访问密钥 ID】和【私有访问密钥】，在后面的部署文件会用到

### 部署鉴权文件

文件部署路径 `~/.aws/credentials`

将前面创建的鉴权信息输入该配置文件中
```
[default]
aws_access_key_id = xxxx
aws_secret_access_key = xxxx
```
