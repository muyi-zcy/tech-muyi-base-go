---
name: tech-muyi-base-go-project-init
description: 新建基于 tech-muyi-base-go 的 Go 项目、初始化目录结构、配置 core.Initialize、Viper 配置与基础设施（DB、Redis）。使用场景：新建服务、从零搭建 Go 后端、初始化项目骨架、配置多环境（dev/local/pre/prod）、接入数据库或 Redis。
---

# tech-muyi-base-go 项目初始化

基于 `github.com/muyi-zcy/tech-muyi-base-go` 创建新项目时，按本 Skill 进行初始化与配置。

## 使用场景

- 用户说「新建一个 Go 服务」「基于 tech-muyi-base-go 搭建项目」
- 用户询问项目目录结构、如何初始化、如何配置多环境
- 用户需要配置 `core.Initialize`、接入 database/redis、`--env`、`app/*.conf`

---

## 初始化流程

### 1. 确认是否使用 MySQL / Redis（初始化时必须执行）

在**开始真正初始化（创建目录、写 go.mod、写配置文件）之前**，先确认本服务是否需要 MySQL 或 Redis：

- **若用户已明确说明**（如「只用 MySQL，不用 Redis」「两个都要」「都不用」），直接记录选择，无需再次询问。
- **若用户未说明**，在初始化步骤中**必须询问一次**，推荐使用 Cursor 的 `AskQuestion` 工具，或以自然语言提问：

> 本服务需要接入哪些基础设施？  
> 请选择：  
> - **仅 MySQL 数据库**  
> - **仅 Redis**  
> - **MySQL + Redis 都需要**  
> - **都不需要，只保留基础 HTTP 能力**

后续写入 `app/app-*.conf`、生成 `go.mod` 时，根据上述选择决定是否生成 `[database]` / `[redis]` 段，以及是否建议引入相关依赖。

### 2. 创建目录结构

```
project-root/
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
└── logs/                   # 加入 .gitignore
```

### 3. 直接初始化基础设施配置

根据步骤 1 的选择，在 `app/app-dev.conf` 中写入对应配置段：

| 选择 | 配置段 | 说明 |
|------|--------|------|
| database | `[database]` | 填写 host、port、username、password、database；host/port 非空时 starter 会自动 `InitDatabase()` |
| redis | `[redis]` | 填写 host、port、password、db；host/port 非空时 starter 会自动 `InitRedis()` |
| 都不需要 | 省略或 host/port 留空 | starter 的 `needDatabase()`/`needRedis()` 为 false，不会初始化 |

**核心规则**：`needDatabase()` 为 `host != "" && port > 0`，`needRedis()` 同理。配置正确则 `core.Initialize()` 会自动完成 DB/Redis 初始化。

### 4. 写入 main.go、go.mod、TOML

按下方模板生成文件，基础设施无需额外代码，starter 会根据配置自动注册。

---

## main.go 模板

```go
package main

import (
	"github.com/muyi-zcy/tech-muyi-base-go/core"
)

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

## 启动流程

1. `core.Initialize()` → `config.Init()` → `starter.Initialize()`
2. `starter.Initialize()`：创建 Gin、初始化日志、注册中间件；根据 `needDatabase()`/`needRedis()` 调用 `InitDatabase()`/`InitRedis()`
3. `starter.Run()`：注册 `/ok`、NoRoute/NoMethod、`Engine.Run(:port)`

---

## 配置

### 命令行参数

| 参数 | 说明 | 示例 |
|-----|------|------|
| `--env` | 环境：dev/local/pre/prod | `go run main.go --env dev` |
| `--config` | 指定配置文件路径 | `go run main.go --config ./app/app-dev.conf` |

### 配置文件查找顺序

1. 当前目录：`./app-dev.conf`
2. `app/app-dev.conf`
3. 均不存在则使用默认配置

### TOML 配置完整模板（初始化时一次写全字段）

初始化时，建议直接写出 **完整的配置字段集合**，而不是只写部分字段；不需要的字段可以先用默认值或留空，后续再按环境调整。

下面是一个基于 `config.Config` 结构和示例 `app-dev.conf` 的**完整模板**（dev 环境）：

```toml
app_name = "NewService"
version = "1.0.0"

[server]
port = 8080
mode = "dev"            # dev / debug / pre / release 等

[log]
level      = "debug"
filename   = "logs/new-service-dev.log"
maxsize    = 100        # 单个日志文件最大大小（MB）
maxage     = 30         # 保留天数
maxbackups = 3          # 最大备份数
compress   = true       # 是否压缩
stdout     = true       # 是否输出到控制台
log_sql    = true       # 是否记录 SQL 日志

[database]
driver                = "mysql"
host                  = "localhost"
port                  = 3306
username              = "root"
password              = "your_password"
database              = "your_db"
max_open_conns        = 25        # 最大打开连接数
max_idle_conns        = 25        # 最大空闲连接数
conn_max_lifetime     = 0         # 连接最大生命周期（秒）
conn_max_idle_time    = 0         # 连接最大空闲时间（秒）
conn_timeout_sec      = 10        # 连接超时（秒）
read_timeout_sec      = 3         # 读超时（秒）
write_timeout_sec     = 3         # 写超时（秒）
skip_default_transaction = true   # 是否跳过默认事务
prepare_stmt          = false     # 是否开启 PrepareStmt
slow_threshold_ms     = 200       # 慢 SQL 阈值（毫秒）
timezone              = "Local"   # 时区，如 Local / Asia/Shanghai
extra_params          = ""        # 追加到 DSN 的额外参数

[redis]
host                = "localhost"
port                = 6379
password            = ""          # 无密码可留空
db                  = 0
pool_size           = 10          # 连接池大小
min_idle_conns      = 2           # 最小空闲连接数
dial_timeout_sec    = 5           # 建立连接超时（秒）
read_timeout_sec    = 3           # 读超时（秒）
write_timeout_sec   = 3           # 写超时（秒）
pool_timeout_sec    = 4           # 从连接池获取连接超时（秒）
idle_timeout_sec    = 300         # 空闲连接超时（秒）
max_retries         = 2           # 最大重试次数
min_retry_backoff_ms = 8          # 最小重试退避（毫秒）
max_retry_backoff_ms = 512        # 最大重试退避（毫秒）
```

#### 必要但未给出的参数 — 默认使用本地配置

- **当用户选择「要用 MySQL / Redis」但没有给出具体连接信息时**：
  - 对于 **MySQL**：使用上面模板中的本地默认值（`host=localhost, port=3306, username=root, password=localMysqlPasswd 或占位, database=<项目名>`）。
  - 对于 **Redis**：使用本地默认值（`host=localhost, port=6379, db=0, password 为空或占位`）。
- **当用户明确表示「暂时不用数据库 / Redis」时**：
  - 仍可保留完整 `[database]` / `[redis]` 段，但将 `host` 留空或注释掉；starter 会因为 `host 为空` 而跳过对应基础设施的初始化。

这样在「必要内容不确定」的情况下，**初始化阶段也不会卡住**，可以先用本地默认参数跑通服务，后续再按环境（dev/local/pre/prod）分别调整实际连接信息。

> 若初始化时选择「暂时不用数据库 / Redis」，**仍建议写出上述完整段落**，但可以将对应 `host` 留空或改成占位值，starter 会根据 `host 非空且 port > 0` 决定是否初始化对应基础设施。

---

## 基础设施详情

### 数据库初始化

- `infrastructure.InitDatabase()` 由 starter 按配置自动调用
- 配置来自 `config.GetDatabaseConfig()`
- 支持 MySQL、连接池、BaseDO Hooks、可选 SQL 日志

### Redis 初始化

- `infrastructure.InitRedis()` 由 starter 按配置自动调用
- 配置来自 `config.GetRedisConfig()`

### 中间件（已内置）

1. **ContextMiddleware**：traceId、ssoId
2. **ExceptionHandler**：panic 与 `c.Errors` 统一处理
3. **Logger**：请求/响应日志

### BaseRepository、GORM Hooks

- `myRepository.NewBaseRepository()` 使用 `infrastructure.GetDB()`
- BaseDO 的 Id/Creator/GmtCreate 等由 GORM Hooks 自动填充
- 查询默认 `row_status=0`，`DeleteById` 为软删除

---

## 自定义 starter 扩展

```go
starter, _ := core.Initialize()
engine := starter.GetEngine()
v1 := engine.Group("/api/v1")
v1.GET("/hello", helloHandler)
starter.Run()
```

## go.mod 依赖

> **推荐做法**：不要手写具体版本号，直接在项目根目录执行  
> `go get -u github.com/muyi-zcy/tech-muyi-base-go@latest && go mod tidy`  
> 由 Go 工具自动把「当前最新版本」写入 `go.mod`。

### 最小可用模板

```go
module your_module_name

go 1.22

require (
	// 由 `go get ...@latest` 自动写入最新版本
	github.com/muyi-zcy/tech-muyi-base-go vX.X.X
)
```

> **注意**：创建文件后，**立刻执行一次** `go mod tidy` 或 `go mod download`，一次性拉取 `tech-muyi-base-go` 及其全部依赖，避免运行时才报缺依赖错误。

### 推荐完整依赖（与 tech-muyi-base-go 保持一致）

如果你的业务代码会直接使用 Gin、GORM、Redis、日志等组件，可以在 `go.mod` 中显式加入与 `tech-muyi-base-go/go.mod` 相同的依赖版本，例如：

```go
require (
	// Web & 配置
	github.com/gin-gonic/gin v1.10.1
	github.com/spf13/viper v1.18.2

	// 数据库 & ORM
	gorm.io/gorm v1.30.2
	gorm.io/driver/mysql v1.6.0

	// Redis
	github.com/go-redis/redis/v8 v8.11.5

	// 日志
	go.uber.org/zap v1.27.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1

	// 工具
	github.com/google/uuid v1.6.0
	github.com/pkg/errors v0.9.1

	// 基础脚手架
	// 建议通过 `go get github.com/muyi-zcy/tech-muyi-base-go@latest` 让 Go 自动填入最新版本
	github.com/muyi-zcy/tech-muyi-base-go vX.X.X
)
```

如需本地开发并使用当前仓库代码，可在 `go.mod` 末尾增加本地 replace（按实际路径调整）：

```go
replace github.com/muyi-zcy/tech-muyi-base-go => ../tech-muyi-base-go
```

