# Supernove2024

2024年腾讯超新星比赛后台项目

## 题目

游戏消息路由 MiniGameRouter 设计

## 题目调研

SDK必备接口：
1. 服务注册，一个服务启动时，将其注册到RouterServer
2. 服务发现，返回Client请求的服务
3. 健康监测
4. 常规路由匹配规则，
5. 定义路由匹配规则，根据相同前缀进行匹配，并且根据路由匹配规则实现路由查找

性能要求：
1. 支持千万级别的路由表
2. 上万服务实例
3. 分析SDK性能占用
4. 对路由规则进行压力测试

进阶要求：
1. 结合grpc实现
2. SideCar可以参考Istio实现

## 架构设计

### 基本设计
采用yaml作为配置文件，提供基本配置，将整体服务分为3个微型服务器
1. 服务发现svr
2. 服务注册svr
3. 健康检查svr

### redis数据设计
对于一个Service，我们在proto中定义了一个ServiceInfo，在redis中储存他的protobuf形式和Revision

| 资源   | redis类型 | key         | filed                | value   | 数据类型        |
|------|---------|-------------|----------------------|---------|-------------|
| 服务版本 | hash    | Service.服务名 | ServiceRevisionFiled | 版本号     | int64       |
| 服务信息 | hash    | Service.服务名 | ServiceInfoFiled     | 服务和实例数据 | ServiceInfo |

对于健康信息，我们采用另一种redis储存方式

| 资源   | redis类型 | key             | filed         | value           |
|------|---------|-----------------|---------------|-----------------|
| 健康数据 | hash    | Health.服务名.实例ID | TTL           | 超时上报间隔          |
| 健康数据 | hash    | Health.服务名.实例ID | LastHeartBeat | 最后一次上报的时间       |
| 服务集合 | set     | Set.服务名         | 无             | 所有实例的InstanceID |


### 配置设计
配置文件要指明三个微务器集群的地址

### SDK设计
使用grpc作为SDK和服务器通信的协议，因为服务器分为三个微服务
SDK也会分为三个部分，每个部分对应一个独立的微服务

#### SDK启动流程
1. 首先需要一个配置文件初始化SDK
2. SDK具备本地缓存数据的能力
3. SDK

### 服务注册svr设计
服务发现每次增量更新后都需要重新写入redis中

### 服务发现svr设计
采用Redis云服务器作为数据库储存服务数据，
添加本地缓存的数据，主要数据在redis内储存，每过一段时间（可配置）就从redis中同步一次数据

### 健康检查svr设计

TTL，服务自己上报健康数据，健康检查服务器自动更新redis中的健康数据

### 动态路由svr设计
1. 一致性哈希
2. 随机
3. 基于权重
4. 指定目标
5. 键值对前缀匹配

### 路由策略和动态路由


## 项目结构

- miniRouterProto: sdk和svr通信所使用的grpc协议库
- sdk: 客户端所使用的sdk
- register-svr: 服务注册服务器

