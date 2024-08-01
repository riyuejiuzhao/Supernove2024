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

1. 对于一个Service，我们在proto中定义了一个ServiceInfo，在redis中储存他的protobuf形式和Revision

| 资源   | redis类型 | key              | filed                | value   | 数据类型        |
|------|---------|------------------|----------------------|---------|-------------|
| 服务版本 | hash    | Hash.Service.服务名 | ServiceRevisionFiled | 版本号     | int64       |
| 服务信息 | hash    | Hash.Service.服务名 | ServiceInfoFiled     | 服务和实例数据 | ServiceInfo |

2. 健康信息的redis储存方式如下

| 资源     | redis类型 | key                  | filed         | value           |
|--------|---------|----------------------|---------------|-----------------|
| 健康数据   | hash    | Hash.Health.服务名.实例ID | TTL           | 超时上报间隔          |
| 健康数据   | hash    | Hash.Health.服务名.实例ID | LastHeartBeat | 最后一次上报的时间       |

3. 路由信息的redis储存方式

| 资源   | redis类型 | key             | filed               | value | 数据类型              |
|------|---------|-----------------|---------------------|-------|-------------------|
| 路由版本 | hash    | Hash.Router.服务名 | RouterRevisionFiled | 版本号   | int64             |
| 路由信息 | hash    | Hash.Router.服务名 | RouterInfoFiled     | 路由数据  | ServiceRouterInfo |


### 配置设计
配置文件要指明三个微务器集群的地址和其他配置

### SDK设计
使用grpc作为SDK和服务器通信的协议，因为服务器分为三个微服务
SDK也会分为三个部分，每个部分对应一个独立的微服务

#### 数据同步设计

第一次启动SDK的时候会自动拉去一次数据，之后定时触发，
根据配置中注册的本地数据更新任务，每过一段时间就自动拉取一次，

> 一开始设计的，配置文件中没有设置需要拉去的服务名，然后当用户GetInstances的时候，
> SDK检查本地是否有数据，如果没有就拉取对应的数据，
> 但是这个有一个问题，每次用户Get都有可能触发一次拉取，
> 这导致Get不能利用读写锁。
> 
> 我们设想一下这样一个场景，用户多线程调用GetInstance(ServiceA)，
> 查询时如果使用读锁，那么第一个请求和第二个请求同时查询，同时发现ServiceA不存在，
> 然后两个都请求写锁进行写入，一个先一个后会写入两次，
> 想要避免这个状况就必须在查询的时候使用写锁，这样只有一个会进行查询，但是这样做很明显不太符合
> SDK查多写少的情况，所以提前注册好需要拉取的服务

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

