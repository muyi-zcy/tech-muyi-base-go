# tech-muyi-base-go 使用说明

本文档是 `github.com/muyi-zcy/tech-muyi-base-go` 的完整使用指南，涵盖从零搭建服务、配置、分层开发、RPC/Nacos 插件、示例联调与部署。

---

## 目录

1. [概述](#1-概述)
2. [环境要求](#2-环境要求)
3. [快速开始](#3-快速开始)
4. [项目结构](#4-项目结构)
5. [配置系统](#5-配置系统)
6. [启动与生命周期](#6-启动与生命周期)
7. [中间件](#7-中间件)
8. [日志系统](#8-日志系统)
9. [数据库与 GORM](#9-数据库与-gorm)
10. [Model 与 BaseDO](#10-model-与-basedo)
11. [Repository 层](#11-repository-层)
12. [Service 层](#12-service-层)
13. [Controller 与 API 规范](#13-controller-与-api-规范)
14. [异常处理](#14-异常处理)
15. [上下文透传 traceId / ssoId](#15-上下文透传-traceid--ssoid)
16. [Nacos 服务注册与发现](#16-nacos-服务注册与发现)
17. [gRPC 插件](#17-grpc-插件)
18. [示例服务联调](#18-示例服务联调)
19. [Docker 部署](#19-docker-部署)
20. [常见问题](#20-常见问题)
21. [最佳实践清单](#21-最佳实践清单)

---

## 1. 概述

`tech-muyi-base-go` 是基于 **Gin** 的 Go 微服务基础脚手架，提供：

| 能力 | 包/模块 | 说明 |
|------|---------|------|
| 应用启动 | `core` | `Initialize()` + `Run()`，自动初始化配置、日志、中间件、基础设施 |
| 配置管理 | `config` | Viper + TOML，支持 dev/local/pre/prod 多环境、热更新 |
| 日志 | `myLogger` | Zap + Lumberjack 滚动日志，支持 traceId 透传 |
| 数据库 | `infrastructure` | GORM + MySQL，连接池、慢 SQL、BaseDO Hooks |
| 缓存 | `infrastructure` | Redis 客户端（go-redis v8） |
| 统一返回 | `myResult` | HTTP 200 + body 内 success/code/message |
| 异常体系 | `myException` | 业务异常、校验异常、404 等 |
| 上下文 | `myContext` | traceId、ssoId、token 在 HTTP/gRPC 间透传 |
| 数据访问 | `myRepository` | BaseRepository CRUD + 软删除 + 分页 |
| 实体基类 | `model` | BaseDO + DateTime 自定义类型 |
| Nacos（可插拔） | `infrastructure/nacos` | 服务注册/发现，失败降级 noop |
| gRPC（可插拔） | `infrastructure/rpc` | Server/Client/Resolver/拦截器 |

**设计原则：**

- 最小启动只需 `core.Initialize()` + `starter.Run()`，DB/Redis/Nacos/RPC 按配置自动启用或跳过。
- 插件失败不阻断 HTTP 启动（Nacos 连接失败 → noop；RPC 关闭 → 仅 HTTP）。
- 业务代码按 **Model → Repository → Service → Controller** 分层组织。

---

## 2. 环境要求

| 依赖 | 版本 | 是否必须 |
|------|------|----------|
| Go | 1.22.x | 是 |
| MySQL | 5.7+ / 8.x | 可选（配置 host+port 时启用） |
| Redis | 6.x+ | 可选（配置 host+port 时启用） |
| Nacos | 2.x | 可选（`plugins.nacos.enabled=true` 时） |

---

## 3. 快速开始

### 3.1 作为依赖引入新项目

```bash
mkdir my-service && cd my-service
go mod init github.com/your-org/my-service
go get github.com/muyi-zcy/tech-muyi-base-go@latest
go mod tidy
```

**main.go（最小可用）：**

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

**目录：**

```
my-service/
├── main.go
├── go.mod
├── app/
│   ├── app-dev.conf
│   ├── app-local.conf
│   ├── app-pre.conf
│   └── app-prod.conf
├── controller/
├── service/
├── repository/
├── model/
└── logs/          # 加入 .gitignore
```

### 3.2 本地运行脚手架本身

```bash
cd tech-muyi-base-go
go mod download
go run main.go --env dev
# 或指定配置文件
go run main.go --config ./app/app-dev.conf
```

### 3.3 验证

```bash
curl http://127.0.0.1:8080/ok
curl http://127.0.0.1:8080/api/v1/system/health
curl http://127.0.0.1:8080/api/v1/test/ping
```

### 3.4 本地开发 replace

在业务项目 `go.mod` 末尾：

```go
replace github.com/muyi-zcy/tech-muyi-base-go => ../tech-muyi-base-go
```

---

## 4. 项目结构

```
tech-muyi-base-go/
├── app/                          # 各环境 TOML 配置
├── config/                       # Config 结构体、Viper 初始化、plugins 配置
├── core/                         # Starter、App、健康检查、优雅退出、RPC 入口
├── infrastructure/               # DB、Redis、GORM Hooks、Nacos、RPC
│   ├── nacos/
│   └── rpc/
│       ├── interceptor/          # Recovery、Context、Logging、ErrorMapping
│       └── resolver/             # static / nacos 解析器
├── middleware/                   # 日志、异常、404/405
├── model/                        # BaseDO、DateTime
├── myContext/                    # HTTP + gRPC 上下文
├── myException/                  # 异常与错误码
├── myId/                         # 分布式 ID 生成（BaseDO Hook 使用）
├── myLogger/                     # Zap 封装
├── myRepository/                 # BaseRepository
├── myResult/                     # 统一返回
├── myUtils/                      # 工具函数
├── example/                      # minimal / producer / consumer 示例
│   ├── minimal/                  # HTTP + MySQL + Redis
│   ├── producer/                 # gRPC 生产者 + static 直连
│   ├── consumer/                 # Nacos 发现调用 producer
│   ├── proto/                    # echo.proto
│   └── scripts/                  # 自动化测试脚本
├── docs/                         # 本文档
├── main.go                       # 脚手架入口
├── Dockerfile
└── start.sh
```

---

## 5. 配置系统

### 5.1 命令行参数

| 参数 | 说明 | 示例 |
|------|------|------|
| `--env` | 环境名，映射到配置文件 | `go run main.go --env dev` |
| `--config` | 直接指定 TOML 路径（优先级高于 `--env`） | `go run main.go --config ./app/app-dev.conf` |

**环境 → 文件名映射：**

| `--env` | 配置文件 |
|---------|----------|
| dev | app-dev.conf |
| local | app-local.conf |
| pre | app-pre.conf |
| prod | app-prod.conf |

### 5.2 配置文件查找顺序

1. 当前工作目录：`./app-dev.conf`
2. `app/app-dev.conf`
3. 均不存在 → 使用 `config.setDefaultConfig()` 内置默认值

### 5.3 热更新

Viper 监听配置文件变更（`fsnotify`），变更后自动重新 `Unmarshal` 到 `GlobalConfig`。注意：已建立的 DB/Redis 连接不会自动重建，仅内存中的配置对象更新。

### 5.4 完整配置模板（dev）

```toml
app_name = "MyService"
version = "1.0.0"

[server]
port = 8080
mode = "dev"            # dev / debug / release 等，影响 Gin 模式

[log]
level      = "debug"    # debug / info / warn / error
filename   = "logs/my-service-dev.log"
maxsize    = 100        # 单文件最大 MB
maxage     = 30         # 保留天数
maxbackups = 3
compress   = true
stdout     = true       # 是否同时输出到控制台
log_sql    = false      # 是否打印 GORM SQL

[database]
driver                = "mysql"
host                  = "localhost"     # 留空则跳过 DB 初始化
port                  = 3306
username              = "root"
password              = "your_password"
database              = "your_db"
max_open_conns        = 25
max_idle_conns        = 25
conn_max_lifetime     = 0
conn_max_idle_time    = 0
conn_timeout_sec      = 10
read_timeout_sec      = 3
write_timeout_sec     = 3
skip_default_transaction = true
prepare_stmt          = false
slow_threshold_ms     = 200
timezone              = "Local"
extra_params          = ""

[redis]
host                = "localhost"       # 留空则跳过 Redis 初始化
port                = 6379
password            = ""
db                  = 0
pool_size           = 10
min_idle_conns      = 2
dial_timeout_sec    = 5
read_timeout_sec    = 3
write_timeout_sec   = 3
pool_timeout_sec    = 4
idle_timeout_sec    = 300
max_retries         = 2
min_retry_backoff_ms = 8
max_retry_backoff_ms = 512

# ---- 可插拔插件（默认关闭）----

[plugins.nacos]
enabled = false
serverAddr = "127.0.0.1:8848"
namespace = "dev"
group = "XI_PLATFORM"
serviceName = "my.service"
configEnabled = false
configDataId = ""
username = ""
password = ""

[plugins.rpc]
enabled = false
protocol = "grpc"
registry = "nacos"          # nacos | static

[plugins.rpc.server]
port = 9080
maxRecvMsgSize = 4194304
enableReflection = false    # dev 可开 true 便于 grpcurl

[plugins.rpc.client]
defaultTimeoutMs = 3000
maxRetry = 0

[plugins.rpc.client.services]
# 配置键 -> Nacos 服务名
xi.app = "xi.app"

[plugins.rpc.static]
# static 模式下的地址映射（配置键 -> host:port）
xi.app = "127.0.0.1:9080"
```

### 5.5 基础设施启用条件

| 组件 | 启用条件 | 代码位置 |
|------|----------|----------|
| MySQL | `database.host != "" && database.port > 0` | `core/starter.go needDatabase()` |
| Redis | `redis.host != "" && redis.port > 0` | `core/starter.go needRedis()` |
| Nacos | `plugins.nacos.enabled = true` | `infrastructure/nacos` |
| gRPC | `plugins.rpc.enabled = true` | `infrastructure/rpc` |

---

## 6. 启动与生命周期

### 6.1 启动流程

```
main()
  └─ core.Initialize()
       ├─ config.Init()           # 解析 --env / --config，加载 TOML
       ├─ NewAppFromConfig()      # 创建 App{Name, Version, Config}
       └─ starter.Initialize()
            ├─ gin.New()
            ├─ InitializeLogger()
            ├─ RegisterDefaultMiddlewares()
            └─ RegisterInfrastructure()
                 ├─ InitDatabase()   (条件)
                 ├─ InitRedis()      (条件)
                 └─ registerPlugins() → Nacos + RPC Init

starter.Run() / RunWithOptions()
  ├─ RegisterHealthCheckRoutes → GET /ok
  ├─ NoRoute / NoMethod 处理器
  ├─ RPC Listen + Start + Nacos Register (条件)
  ├─ HTTP ListenAndServe (goroutine)
  └─ 等待 SIGINT/SIGTERM → 优雅关闭
       ├─ Nacos Deregister
       ├─ RPC GracefulStop + Client Close
       ├─ HTTP Shutdown (15s timeout)
       └─ App.Shutdown() → CloseDB + CloseRedis + Logger Sync
```

### 6.2 注册业务路由

在 `Run()` 之前获取 Engine 并注册：

```go
starter, _ := core.Initialize()

engine := starter.GetEngine()
v1 := engine.Group("/api/v1")
v1.GET("/user/getUserInfo", userController.GetUserInfo)

starter.Run()
```

### 6.3 注册 gRPC 服务

```go
starter, _ := core.Initialize()

starter.RegisterGrpcServices(func(s *grpc.Server) {
    pb.RegisterYourServiceServer(s, &YourServerImpl{})
})

starter.Run()
// 或一步完成：
starter.RunWithGrpc(func(s *grpc.Server) {
    pb.RegisterYourServiceServer(s, &YourServerImpl{})
})
```

### 6.4 获取 RPC 客户端

```go
conn, err := starter.GetRpcManager().Client().GetConn("xi.app")
client := pb.NewCatalogServiceClient(conn)
```

`serviceKey` 对应 `[plugins.rpc.client.services]` 或 `[plugins.rpc.static]` 中的配置键。

---

## 7. 中间件

启动器默认注册（顺序不可颠倒）：

| 顺序 | 中间件 | 包 | 作用 |
|------|--------|-----|------|
| 1 | ContextMiddleware | myContext | 注入 traceId、ssoId 到 Gin + Request.Context |
| 2 | ExceptionHandler | middleware | panic 捕获、c.Errors 统一 JSON 返回 |
| 3 | Logger | middleware | 请求/响应日志（含 traceId） |

**404/405：** 在 `Run()` 时注册 `NoRoute(NotFoundHandler)`、`NoMethod(MethodNotAllowedHandler)`，HTTP 状态码仍为 200，body 中 code 为 404/405。

---

## 8. 日志系统

### 8.1 配置项

见 `[log]` 段：`level`、`filename`、`maxsize`、`maxage`、`maxbackups`、`compress`、`stdout`。

### 8.2 常用 API

```go
import "github.com/muyi-zcy/tech-muyi-base-go/myLogger"

myLogger.Info("消息", zap.String("key", "value"))
myLogger.Error("错误", zap.Error(err))

// 从 Gin Context 自动取 traceId
myLogger.InfoCtx(c, "请求处理", zap.String("path", c.Request.URL.Path))

// 从标准 context 取 traceId（Service/Repository 层）
myLogger.InfoCtx(ctx, "业务日志", zap.Int64("id", id))
```

### 8.3 SQL 日志

`log.log_sql = true` 时，GORM 使用自定义 logger 输出 SQL；慢 SQL 阈值由 `database.slow_threshold_ms` 控制。

---

## 9. 数据库与 GORM

### 9.1 初始化

`infrastructure.InitDatabase()` 由 starter 自动调用，成功后可通过 `infrastructure.GetDB()` 获取 `*gorm.DB`。

### 9.2 DSN 构建

MySQL DSN 格式：

```
{username}:{password}@tcp({host}:{port})/{database}?charset=utf8mb4&parseTime=True&loc={timezone}&timeout={conn_timeout_sec}s&readTimeout={read_timeout_sec}s&writeTimeout={write_timeout_sec}s&{extra_params}
```

### 9.3 BaseDO Hooks

`infrastructure.RegisterBaseDOHooks(db)` 在 InitDatabase 时自动注册，行为：

**Create 时自动填充（若字段为空）：**

| 字段 | 值来源 |
|------|--------|
| Id | `myId.NextId()` 雪花 ID |
| Creator / Operator | context 中 ssoId，失败则 `"system"` |
| GmtCreate / GmtModified | `model.Now()`（精确到秒） |
| RowVersion | 0 |
| RowStatus | 0 |

**Update 时自动填充：**

| 字段 | 行为 |
|------|------|
| Operator | 从 ssoId 填充 |
| GmtModified | 当前时间 |
| RowVersion | 自增（乐观锁） |

**Map 更新注意：** 使用 `Updates(map[string]interface{}{...})` 时，必须显式传入 `row_version`，否则 Hook 报错以保证乐观锁生效。

---

## 10. Model 与 BaseDO

### 10.1 BaseDO 字段

```go
type BaseDO struct {
    Id          int64    `gorm:"column:id;primaryKey" json:"id,string"`
    RowVersion  int64    `gorm:"column:row_version" json:"rowVersion"`
    Creator     *string  `gorm:"column:creator" json:"creator"`
    GmtCreate   DateTime `gorm:"column:gmt_create" json:"gmtCreate"`
    Operator    *string  `gorm:"column:operator" json:"operator"`
    GmtModified DateTime `gorm:"column:gmt_modified" json:"gmtModified"`
    ExtAtt      *string  `gorm:"column:ext_att" json:"extAtt"`
    RowStatus   int      `gorm:"column:row_status" json:"rowStatus"`
    TenantID    *string  `gorm:"column:tenant_id" json:"tenantId"`
}
```

### 10.2 建表规范

业务表应包含 BaseDO 全部列。示例：

```sql
CREATE TABLE `user` (
  `id`           BIGINT       NOT NULL COMMENT '主键',
  `row_version`  BIGINT       NOT NULL DEFAULT 0 COMMENT '乐观锁',
  `creator`      VARCHAR(64)  DEFAULT NULL COMMENT '创建人',
  `gmt_create`   DATETIME     NOT NULL COMMENT '创建时间',
  `operator`     VARCHAR(64)  DEFAULT NULL COMMENT '更新人',
  `gmt_modified` DATETIME     NOT NULL COMMENT '修改时间',
  `ext_att`      TEXT         DEFAULT NULL COMMENT '扩展属性',
  `row_status`   TINYINT      NOT NULL DEFAULT 0 COMMENT '0正常 1删除',
  `tenant_id`    VARCHAR(64)  DEFAULT NULL COMMENT '租户',
  `username`     VARCHAR(64)  NOT NULL COMMENT '用户名',
  PRIMARY KEY (`id`)
) COMMENT='用户表';
```

### 10.3 DateTime 类型

时间/日期字段统一使用 `model.DateTime`（非 `time.Time`），支持 JSON `"2006-01-02 15:04:05"` 格式及 GORM Scan/Value。

### 10.4 字段类型映射

| MySQL 类型 | Go 类型 |
|------------|---------|
| BIGINT | int64 |
| INT | int |
| TINYINT(1) 布尔 | bool 或 *bool |
| VARCHAR/TEXT | string |
| DATETIME/TIMESTAMP/DATE | model.DateTime |
| DECIMAL | float64（或自定义 decimal） |
| 可空字段 | 指针类型 *string、*int64、*model.DateTime |

### 10.5 Model 示例

```go
package model

type UserDO struct {
    BaseDO
    Username string `gorm:"column:username" json:"username"`
    Email    string `gorm:"column:email" json:"email"`
}

func (UserDO) TableName() string {
    return "user"
}
```

---

## 11. Repository 层

### 11.1 BaseRepository 接口

```go
type BaseRepository interface {
    GetDB() (*gorm.DB, error)
    Insert(ctx, entity) error
    Update(ctx, entity, id) error
    DeleteById(ctx, entity, id) error      // 软删除 row_status=1
    GetById(ctx, entity, id) error
    GetAll(ctx, entity, sortFields...) error
    GetByCondition(ctx, entity, conditions, sortFields...) error
    GetPageByCondition(ctx, entity, conditions, query, sortFields...) error
    CountByCondition(ctx, entity, conditions) (int64, error)
}
```

### 11.2 默认行为

- 所有查询自动附加 `row_status = 0`
- `DeleteById` 软删除，设置 `row_status=1`、`operator`、`gmt_modified`
- 分页使用 `myResult.MyQuery`（默认 size=20，最大 2000）

### 11.3 实现模板

```go
type UserRepository struct {
    myRepository.BaseRepository
}

func NewUserRepository() *UserRepository {
    return &UserRepository{BaseRepository: myRepository.NewBaseRepository()}
}
```

### 11.4 简单 vs 复杂查询

| 场景 | 推荐方式 |
|------|----------|
| 单表等值条件 | `GetByCondition` / `GetPageByCondition` |
| 多表 JOIN、OR、子查询 | `GetDB()` + GORM 链式调用 |

```go
func (r *UserRepository) FindByRole(ctx context.Context, role string) ([]*model.UserDO, error) {
    db, err := r.GetDB()
    if err != nil {
        return nil, err
    }
    var list []*model.UserDO
    err = db.WithContext(ctx).
        Where("row_status = 0 AND role = ?", role).
        Order("gmt_create DESC").
        Find(&list).Error
    return list, err
}
```

---

## 12. Service 层

无基类，按业务组织，调用 Repository，内层错误用 `github.com/pkg/errors` Wrap 保留堆栈：

```go
func (s *UserService) GetById(ctx context.Context, id int64) (*model.UserDO, error) {
    user := &model.UserDO{}
    if err := s.repo.GetById(ctx, user, id); err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, myException.NewExceptionFromError(myException.NOT_FOUND)
        }
        return nil, perrors.Wrap(err, "查询用户失败")
    }
    return user, nil
}
```

**原则：** Repository 不直接构造 `MyException`；在 Service/Controller 边界映射业务错误码。

---

## 13. Controller 与 API 规范

### 13.1 统一返回结构

```json
{
  "code": "200",
  "success": true,
  "message": "success",
  "data": { ... },
  "query": null
}
```

分页时 `query` 含 `size`、`current`、`total`。

### 13.2 HTTP 状态码约定

**所有接口 HTTP 状态码统一 200**，成功/失败由 body 的 `success`、`code`、`message` 表达。

### 13.3 路由命名

路径段使用 **小驼峰（lowerCamelCase）**：

```
GET  /api/v1/user/getUserInfo
GET  /api/v1/user/listUsers
POST /api/v1/user/createUser
```

### 13.4 Controller 模板

```go
func (u *UserController) GetUserInfo(c *gin.Context) {
    id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
    user, err := u.svc.GetById(c.Request.Context(), id)
    if err != nil {
        myResult.ErrorWithError(c, err)
        return
    }
    myResult.Success(c, user)
}

func (u *UserController) ListUsers(c *gin.Context) {
    query := &myResult.MyQuery{}
    if err := c.ShouldBindQuery(query); err != nil {
        myResult.ErrorWithError(c, myException.NewValidationError("query", "分页参数错误"))
        return
    }
    list, total, err := u.svc.GetPage(c.Request.Context(), query)
    if err != nil {
        myResult.ErrorWithError(c, err)
        return
    }
    query.SetTotal(total)
    myResult.SuccessWithQuery(c, list, query)
}
```

### 13.5 myResult API 速查

| 方法 | 用途 |
|------|------|
| `Success(c, data)` | 成功返回 |
| `SuccessWithMessage(c, msg, data)` | 成功 + 自定义 message |
| `SuccessWithQuery(c, data, query)` | 分页成功 |
| `ErrorWithError(c, err)` | 错误（交给中间件处理） |
| `Error(c, msg)` | 直接返回错误 JSON |
| `BadRequestResponse(c, msg)` | 400 业务码 |

---

## 14. 异常处理

### 14.1 异常类型

| 类型 | 创建 | 中间件映射 |
|------|------|------------|
| MyException | `NewException(code, msg)` / `NewExceptionFromError(enum)` | FailWithCode |
| ValidationError | `NewValidationError(field, msg)` | BadRequest |
| NotFoundError | `NewNotFoundError(resource, id)` | NotFound |

### 14.2 常用错误码

| 枚举 | Code | 说明 |
|------|------|------|
| OK | 200 | 成功 |
| BAD_REQUEST | 400 | 请求错误 |
| UNAUTHORIZED | 401 | 未授权 |
| NOT_FOUND | 404 | 资源不存在 |
| INTERNAL_SERVER_ERROR | 500 | 服务器错误 |
| INVALID_PARAM | 10000 | 参数无效 |
| DB_EXCEPTION | 10001 | 数据库异常 |

完整枚举见 `myException/common_error_code.go`。

### 14.3 注册自定义错误码

```go
const MY_ERROR myException.CommonErrorCodeEnum = 100

func init() {
    myException.RegisterErrorCode(MY_ERROR, "20001", "自定义业务错误")
}
```

### 14.4 错误处理分层

```
Repository  → pkg/errors.Wrap 保留堆栈
Service     → 映射 gorm.ErrRecordNotFound → NOT_FOUND；其他 Wrap
Controller  → myResult.ErrorWithError(c, err)
```

---

## 15. 上下文透传 traceId / ssoId

### 15.1 HTTP

| 来源 | Header / Cookie |
|------|-----------------|
| traceId | x-trace-id |
| ssoId | x-sso-id |

无则自动生成：traceId=UUID，ssoId=`-1`。

### 15.2 获取方式

```go
// Controller
traceId := myContext.GetTraceIdFromGinCtx(c)
ssoId := myContext.GetSsoIdFromGinCtx(c)

// Service / Repository（使用 c.Request.Context()）
traceId, _ := myContext.GetTraceId(ctx)
ssoId, _ := myContext.GetSsoId(ctx)

// 不报错版本
traceId := myContext.TryGetTraceId(ctx)
```

### 15.3 gRPC

客户端拦截器注入 metadata（x-trace-id、x-sso-id、x-source-service）；服务端 `ContextExtract` 拦截器还原到 context。

---

## 16. Nacos 服务注册与发现

### 16.1 配置

```toml
[plugins.nacos]
enabled = true
serverAddr = "127.0.0.1:8848"
namespace = "dev"
group = "XI_PLATFORM"
serviceName = "my.service"
```

### 16.2 注册时机

`starter.Run()` 时，若 RPC 已启用且 Nacos enabled，自动将 gRPC 端口注册到 Nacos。

### 16.3 降级

- `enabled=false` → noop，跳过注册
- 连接失败 → Warn 日志，HTTP 正常启动
- 退出时自动 Deregister

---

## 17. gRPC 插件

### 17.1 模式对比

| registry | 说明 | 适用场景 |
|----------|------|----------|
| nacos | 通过 Nacos 服务发现 | 生产、多实例 |
| static | 直连 `[plugins.rpc.static]` 配置的地址 | 本地联调、降级 |

### 17.2 完整配置示例（Nacos 模式）

```toml
[plugins.nacos]
enabled = true
serverAddr = "192.168.0.181:8848"
namespace = "dev"
group = "XI_PLATFORM"
serviceName = "example.consumer"

[plugins.rpc]
enabled = true
registry = "nacos"

[plugins.rpc.server]
port = 9082
enableReflection = true

[plugins.rpc.client]
defaultTimeoutMs = 3000

[plugins.rpc.client.services]
example_producer = "example.producer"
```

### 17.3 定义 proto 与生成代码

```protobuf
// example/proto/echo.proto
syntax = "proto3";
package example;
option go_package = "github.com/muyi-zcy/tech-muyi-base-go/example/proto/echo";

service EchoService {
  rpc Ping(PingRequest) returns (PingReply);
  rpc Echo(EchoRequest) returns (EchoReply);
}
```

```bash
cd example/proto
protoc --go_out=. --go-grpc_out=. echo.proto
```

### 17.4 实现 gRPC Server

```go
type EchoServer struct {
    pb.UnimplementedEchoServiceServer
}

func (s *EchoServer) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingReply, error) {
    return &pb.PingReply{Message: "pong"}, nil
}

func RegisterEchoService(s *grpc.Server) {
    pb.RegisterEchoServiceServer(s, &EchoServer{})
}
```

### 17.5 调用远程 gRPC

```go
conn, err := starter.GetRpcManager().Client().GetConn("example_producer")
client := pb.NewEchoServiceClient(conn)
ctx, cancel := context.WithTimeout(ctx, starter.GetRpcManager().Client().DefaultTimeout())
defer cancel()
resp, err := client.Echo(ctx, &pb.EchoRequest{Message: "hello"})
```

### 17.6 内置拦截器链

**Server：** Recovery → ContextExtract → Logging → ErrorMapping

**Client：** ContextInject → ClientLogging → ClientErrorDecode

### 17.7 grpcurl 调试

开启 `enableReflection = true` 后：

```bash
grpcurl -plaintext 127.0.0.1:9081 list
grpcurl -plaintext 127.0.0.1:9081 example.EchoService/Ping
grpcurl -plaintext -d '{"message":"hello"}' 127.0.0.1:9081 example.EchoService/Echo
```

---

## 18. 示例服务联调

详见 [example/README.md](../example/README.md)。

| 示例 | HTTP | gRPC | 说明 |
|------|------|------|------|
| minimal | 8080 | - | HTTP + MySQL + Redis 基础能力 |
| producer | 8081 | 9081 | gRPC 服务 + static 直连 consumer |
| consumer | 8082 | 9082 | Nacos 发现调用 producer |

**联调顺序：**

```bash
# 终端 1
cd example/producer && go run main.go --config ./app/app-dev.conf

# 终端 2
cd example/consumer && go run main.go --config ./app/app-dev.conf

# 验证
curl "http://127.0.0.1:8082/api/v1/call/producer?message=hello"
curl http://127.0.0.1:8081/api/v1/loop/status
curl http://127.0.0.1:8082/api/v1/loop/status
```

**自动化测试：**

```bash
./example/scripts/test-all.sh
```

---

## 19. Docker 部署

```bash
docker build -t tech-muyi-base-go:latest .

docker run -d --name tech-muyi-go \
  -p 28080:28080 \
  -v $(pwd)/app/app-prod.conf:/home/app/app-prod.conf \
  tech-muyi-base-go:latest
```

注意：Dockerfile 暴露端口 **28080**，与本地默认 8080 不同，映射时需对应。

容器内执行 `/home/app/start.sh` 启动 `./main`。

---

## 20. 常见问题

### Q: 数据库/Redis 未初始化？

检查 `[database]` / `[redis]` 的 `host` 和 `port` 是否非空且 port > 0。

### Q: 配置未生效？

确认 `--config` 路径或 `--env` 是否正确；程序优先读当前目录，其次 `app/` 目录。

### Q: 日志不输出到控制台？

设置 `log.stdout = true`。

### Q: Nacos 注册失败但服务仍启动？

设计如此，Nacos 失败降级 noop，HTTP 不受影响。检查 serverAddr、namespace、网络。

### Q: gRPC 调用 static 模式报「未配置地址」？

在 `[plugins.rpc.static]` 中添加 serviceKey 对应的 `host:port`。

### Q: 软删除后查不到数据？

BaseRepository 默认过滤 `row_status=0`，符合预期。管理后台需查已删除数据时用 `GetDB()` 自定义查询。

### Q: Update 时乐观锁不生效？

避免仅用 map 更新且不传 `row_version`；推荐传入完整 struct 或显式指定 row_version。

### Q: Id 在 JSON 中精度丢失？

BaseDO 的 Id 使用 `json:"id,string"`，前端以字符串接收。

---

## 21. 最佳实践清单

- [ ] 新建服务先确认是否需要 MySQL / Redis / Nacos / RPC
- [ ] `app/app-*.conf` 四个环境都维护，敏感信息 prod 用环境变量或密钥管理
- [ ] 业务表包含完整 BaseDO 字段
- [ ] Model 嵌入 BaseDO，时间字段用 `model.DateTime`
- [ ] Repository 组合 BaseRepository，复杂查询用 GetDB()
- [ ] Service 内层 Wrap error，边界映射 MyException
- [ ] Controller 统一 myResult 返回，路径小驼峰
- [ ] 全链路传递 `c.Request.Context()`，不丢 traceId
- [ ] RPC 生产用 nacos，本地联调可用 static
- [ ] dev 环境开 `enableReflection` 便于 grpcurl 调试
- [ ] logs/ 加入 .gitignore
- [ ] 依赖通过 `go get ...@latest && go mod tidy` 管理版本

---

## 相关 Cursor Skills

在 Cursor 中使用以下 Skills 可加速开发：

| Skill | 用途 |
|-------|------|
| `tech-muyi-base-go` | 总览与任务路由 |
| `tech-muyi-base-go-project-init` | 新建项目、目录、配置 |
| `tech-muyi-base-go-db-scaffold` | Model/Repository/Service/Controller 脚手架 |
| `tech-muyi-base-go-api` | Controller 规范、myResult、myContext |
| `tech-muyi-base-go-exception` | 异常与错误码 |
| `tech-muyi-base-go-rpc` | Nacos + gRPC 接入 |
