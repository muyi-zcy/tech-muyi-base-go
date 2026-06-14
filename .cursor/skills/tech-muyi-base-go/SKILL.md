---
name: tech-muyi-base-go
description: tech-muyi-base-go Go 微服务脚手架总览与任务路由。基于 Gin + Viper + GORM + Redis + myResult/myException/myContext，支持可插拔 Nacos 与 gRPC。使用场景：用户询问 tech-muyi-base-go 用法、架构、如何选型、该用哪个子 Skill、从零搭建服务前的整体了解。
---

# tech-muyi-base-go 总览

完整文档见 [docs/USAGE.md](../../../docs/USAGE.md)。

## 何时使用本 Skill

用户提到 `tech-muyi-base-go`、MuYi Go 脚手架、或不确定该用哪个子 Skill 时，先读本 Skill 再路由到对应子 Skill。

## 核心能力速览

| 模块 | 包 | 一句话 |
|------|-----|--------|
| 启动 | `core` | `Initialize()` + `Run()`，自动 DB/Redis/Nacos/RPC |
| 配置 | `config` | TOML，`--env dev/local/pre/prod` 或 `--config` |
| 返回 | `myResult` | HTTP 200 + body success/code/message |
| 异常 | `myException` | BizError / MyException / ValidationError |
| 错误文案 | `myLocale` | contracts/locales + open API + 平台默认文案 |
| 上下文 | `myContext` | traceId、ssoId HTTP/gRPC 透传 |
| 实体 | `model.BaseDO` | 公共字段 + DateTime + GORM Hooks |
| 仓储 | `myRepository.BaseRepository` | CRUD、软删除、分页 |
| RPC | `infrastructure/rpc` | gRPC Server/Client，nacos/static |
| Nacos | `infrastructure/nacos` | 注册发现，失败降级 noop |

## 最小 main.go

```go
package main

import "github.com/muyi-zcy/tech-muyi-base-go/core"

func main() {
    starter, err := core.Initialize()
    if err != nil {
        panic(err)
    }
    if err := starter.Run(); err != nil {
        panic(err)
    }
}
```

## 任务路由 — 选择子 Skill

| 用户意图 | 使用 Skill |
|----------|------------|
| 新建 Go 服务、目录结构、go.mod、app/*.conf | `tech-muyi-base-go-project-init` |
| 根据表生成 Model/Repository/Service/Controller | `tech-muyi-base-go-db-scaffold` |
| 写 HTTP 接口、统一返回、分页、traceId | `tech-muyi-base-go-api` |
| 业务异常、错误码、校验失败 | `tech-muyi-base-go-exception` |
| Nacos 注册、gRPC 服务/客户端、proto | `tech-muyi-base-go-rpc` |

## 基础设施启用条件

- **MySQL**：`database.host != "" && database.port > 0`
- **Redis**：`redis.host != "" && redis.port > 0`
- **Nacos**：`plugins.nacos.enabled = true`
- **gRPC**：`plugins.rpc.enabled = true`

host 留空则 starter 跳过对应组件，无需额外代码。

## 标准分层

```
Controller → Service → Repository → GORM
     ↑           ↑
 myResult    myException（边界映射）
 myContext   c.Request.Context() 全链路传递
```

## 注册路由与 gRPC（Run 之前）

```go
starter, _ := core.Initialize()

// HTTP
routes.Register(starter.GetEngine())

// gRPC（plugins.rpc.enabled=true）
starter.RegisterGrpcServices(func(s *grpc.Server) {
    pb.RegisterXxxServer(s, &XxxServer{})
})

starter.Run()
```

## 示例项目

`example/` 下有三个独立示例，详见 [example/README.md](../../../example/README.md)：

- `minimal/` — HTTP + MySQL + Redis
- `producer/` — gRPC + static 直连
- `consumer/` — Nacos 发现调用 producer

自动化测试：`./example/scripts/test-all.sh`

## 依赖引入

```bash
go get github.com/muyi-zcy/tech-muyi-base-go@latest && go mod tidy
```

本地开发 replace：

```go
replace github.com/muyi-zcy/tech-muyi-base-go => ../tech-muyi-base-go
```
