# Supernove2024

详细介绍可以查看文件 document/超新星期末报告.ppt

**文件结构**:
- test：测试文件
- sdk：SDK代码实现
- grpc-sdk：与grpc相关的代码实现
- pb：proto文件以及生成的go文件
- util：通用文件

**测试：**
在启动测试之前，先在compose中启动监控和etcd集群
```shell
docker compose -f etcd.yaml up -d
docker compose -f compose.yaml up -d
```
如果测试中遇到问题，可以尝试重启etcd集群，很可能是因为etcd集群数据没清空

- test-grpc：TestGrpc是grpc熔断测试，同时也是grpc启动，添加路由，路由分发的测试
- test-instance-pressure：TestManyInstance是服务实例压力测试，同时启动多个服务实例
- test-router-pressure：路由压力测试TestRouterWatch，由于一万个实例占用内存比较大，这里的测试不保存数据到本地，但是watch机制仍然存在，只是watch的结果不进行储存了
- test-sdk-pressure：sdk非联网功的测试，
- test-show-router-change：基本路由功能测试
- test-show-router-change2：金丝雀发布演示
- unit-test：sdk基本功能测试

