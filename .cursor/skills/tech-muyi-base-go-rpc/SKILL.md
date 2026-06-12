---
name: tech-muyi-base-go-rpc
description: 在 tech-muyi-base-go 中接入可插拔 Nacos 服务注册发现与 gRPC RPC。涵盖 plugins.nacos/plugins.rpc 配置、RegisterGrpcServices、GetRpcManager 客户端调用、static/nacos 两种 registry、proto 生成与拦截器透传。使用场景：接入 gRPC、Nacos 注册发现、服务间 RPC 调用、grpcurl 调试、producer/consumer 联调。
---

# tech-muyi-base-go Nacos + gRPC

完整文档见 [docs/USAGE.md §16-17](../../../docs/USAGE.md#16-nacos-服务注册与发现)。

## 使用场景

- 开启/配置 Nacos 注册与 gRPC 端口
- 定义 proto、实现 Server、注册到 starter
- 通过 Nacos 或 static 调用其他服务
- 参考 example/producer、example/consumer 联调

## 配置模板

### Nacos + gRPC（consumer 模式）

```toml
[plugins.nacos]
enabled = true
serverAddr = "127.0.0.1:8848"
namespace = "dev"
group = "XI_PLATFORM"
serviceName = "my.service"

[plugins.rpc]
enabled = true
registry = "nacos"

[plugins.rpc.server]
port = 9080
enableReflection = true

[plugins.rpc.client]
defaultTimeoutMs = 3000

[plugins.rpc.client.services]
peer_service = "peer.service.name"
```

### Static 直连（本地联调 / 降级）

```toml
[plugins.nacos]
enabled = false

[plugins.rpc]
enabled = true
registry = "static"

[plugins.rpc.server]
port = 9081
enableReflection = true

[plugins.rpc.client.services]
peer_service = "peer.service.name"

[plugins.rpc.static]
peer_service = "127.0.0.1:9082"
```

## 降级行为

| 条件 | 行为 |
|------|------|
| `plugins.nacos.enabled=false` | 跳过注册，RPC 可用 static |
| `plugins.rpc.enabled=false` | 仅 HTTP，不监听 gRPC |
| Nacos 连接失败 | Warn + noop，**不阻断** HTTP |

## main.go 模板

```go
package main

import (
    "github.com/muyi-zcy/tech-muyi-base-go/core"
    "your-module/routes"
    "your-module/server"
    "google.golang.org/grpc"
)

func main() {
    starter, err := core.Initialize()
    if err != nil {
        panic(err)
    }

    routes.Register(starter.GetEngine(), starter)

    starter.RegisterGrpcServices(func(s *grpc.Server) {
        server.RegisterYourService(s)
    })

    if err := starter.Run(); err != nil {
        panic(err)
    }
}
```

等价写法：`starter.RunWithGrpc(func(s *grpc.Server) { ... })`

## 实现 gRPC Server

```go
type YourServer struct {
    pb.UnimplementedYourServiceServer
}

func (s *YourServer) YourMethod(ctx context.Context, req *pb.YourRequest) (*pb.YourReply, error) {
    // ctx 已含 traceId/ssoId（ContextExtract 拦截器注入）
    traceId := myContext.TryGetTraceId(ctx)
    return &pb.YourReply{}, nil
}

func RegisterYourService(s *grpc.Server) {
    pb.RegisterYourServiceServer(s, &YourServer{})
}
```

## 调用远程 gRPC

```go
conn, err := starter.GetRpcManager().Client().GetConn("peer_service")
if err != nil {
    return err
}
client := pb.NewPeerServiceClient(conn)

timeout := starter.GetRpcManager().Client().DefaultTimeout()
ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
defer cancel()

resp, err := client.SomeMethod(ctx, &pb.SomeRequest{})
```

- `peer_service` 对应 `[plugins.rpc.client.services]` 的配置键
- static 模式下地址来自 `[plugins.rpc.static]`

## proto 工作流

```bash
# 1. 编写 proto（go_package 指向业务 module）
# 2. 生成
protoc --go_out=. --go-grpc_out=. your.proto
# 3. 实现 Server + RegisterGrpcServices
```

参考：`example/proto/echo.proto`

## 拦截器链（内置，无需业务配置）

**Server：** Recovery → ContextExtract → Logging → ErrorMapping

**Client：** ContextInject(sourceService) → ClientLogging → ClientErrorDecode

HTTP 请求中的 x-trace-id / x-sso-id 会通过 Client 拦截器注入 gRPC metadata。

## grpcurl 调试

`enableReflection = true` 时：

```bash
grpcurl -plaintext 127.0.0.1:9081 list
grpcurl -plaintext 127.0.0.1:9081 package.Service/Method
grpcurl -plaintext -d '{"field":"value"}' 127.0.0.1:9081 package.Service/Method
```

## 联调参考

| 示例 | 端口 | registry | 说明 |
|------|------|----------|------|
| example/producer | HTTP 8081, gRPC 9081 | static | 注册 Nacos 供 consumer 发现 |
| example/consumer | HTTP 8082, gRPC 9082 | nacos | 通过 Nacos 调 producer |

```bash
cd example/producer && go run main.go --config ./app/app-dev.conf
cd example/consumer && go run main.go --config ./app/app-dev.conf
curl "http://127.0.0.1:8082/api/v1/call/producer?message=hello"
```

## 检查清单

- [ ] `plugins.rpc.enabled = true`
- [ ] server.port 与 static 映射地址一致
- [ ] Nacos 模式：`serviceName` 与 client.services 映射的服务名匹配
- [ ] static 模式：`[plugins.rpc.static]` 含所有 peer 的 host:port
- [ ] Run 之前调用 `RegisterGrpcServices`
- [ ] 远程调用使用 `c.Request.Context()` 保留 traceId
