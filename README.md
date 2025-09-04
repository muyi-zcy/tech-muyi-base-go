## Tech MuYi Base Go

基于 Gin 的 Go 服务基础脚手架，内置配置中心（Viper）、日志（Zap + Lumberjack）、数据库（GORM/MySQL）、Redis、统一返回与中间件等，适合作为新服务的快速起步模板。

### 功能特性
- **配置管理**: 基于 Viper，支持多环境 `dev/local/pre/prod`，支持热更新
- **日志体系**: Zap + 滚动日志（lumberjack），可选输出到控制台，支持 TraceId 透传
- **Web 框架**: Gin，已内置基础中间件与健康检查
- **数据访问**: GORM（MySQL 驱动）、Redis 客户端
- **统一返回**: `myResult` 统一成功/失败响应
- **错误与恢复**: 中间件捕获异常并记录

### 环境要求
- Go 1.22.x
- 可选：MySQL、Redis（如不需要可不启用对应功能）

### 目录结构（节选）
```
app/                     # 各环境配置文件（TOML）
config/                  # 配置装载与默认值
core/                    # 应用初始化、启动器
infrastructure/          # 数据库、Redis 初始化与连接
middleware/              # 日志、异常处理等
model/                   # 数据模型
myContext/, myException/ # 上下文与异常封装
myLogger/                # 日志封装
myObject/                # 常用对象与工具
myRepository/            # 仓储基础封装
myResult/                # 统一返回结构
myUtils/                 # 常用工具
main.go                  # 入口，注册路由与启动
Dockerfile               # 多阶段构建镜像
start.sh                 # 容器/本地启动脚本（Linux/容器环境）
```

### 快速开始（本地）
1) 拉取依赖
```bash
go mod download
```

2) 选择配置（默认 `--env dev`，亦可通过 `--config` 指定文件路径）
```bash
# 使用内置环境：dev/local/pre/prod
go run main.go --env dev

# 或者指定配置文件路径（TOML），优先读取当前目录或 app/ 目录
go run main.go --config ./app/app-dev.conf
```

3) 访问服务
- 根路径：`GET /` 返回欢迎信息
- 健康检查：`GET /api/v1/system/health`
- 当前安全配置：`GET /api/v1/system/config`
- 系统信息：`GET /api/v1/system/info`
- 测试：`GET /api/v1/test/ping`、`POST /api/v1/test/echo`、`GET /api/v1/test/error`

示例（curl）
```bash
curl http://127.0.0.1:8080/
curl http://127.0.0.1:8080/api/v1/system/health
curl http://127.0.0.1:8080/api/v1/test/ping
curl -X POST http://127.0.0.1:8080/api/v1/test/echo -H "Content-Type: application/json" -d '{"hello":"world"}'
```

### 配置说明
配置类型为 TOML，文件名按环境约定为：`app-dev.conf`、`app-local.conf`、`app-pre.conf`、`app-prod.conf`。

程序参数：
- `--env`：dev/local/pre/prod（二选一，无 `--config` 时生效）
- `--config`：自定义配置文件路径（优先级更高）

关键项（`config/config.go` 中的默认值可参考）：
- **server**：`port`、`mode`
- **log**：`level`、`filename`、`maxsize`、`maxage`、`maxbackups`、`compress`、`stdout`、`log_sql`
- **database**：
  - 基础：`driver`、`host`、`port`、`username`、`password`、`database`
  - 连接池：`max_open_conns`、`max_idle_conns`、`conn_max_lifetime`、`conn_max_idle_time`
  - 超时（秒）：`conn_timeout_sec`、`read_timeout_sec`、`write_timeout_sec`
  - 行为：`skip_default_transaction`、`prepare_stmt`、`slow_threshold_ms`
  - DSN：`timezone`（如 Local/Asia/Shanghai）、`extra_params`
- **redis**：
  - 基础：`host`、`port`、`password`、`db`
  - 连接池：`pool_size`、`min_idle_conns`
  - 超时（秒）：`dial_timeout_sec`、`read_timeout_sec`、`write_timeout_sec`、`pool_timeout_sec`、`idle_timeout_sec`
  - 重试：`max_retries`、`min_retry_backoff_ms`、`max_retry_backoff_ms`

当配置文件缺失时，会回落到内置默认值并继续运行。

### 运行与构建
- 直接运行
```bash
go run main.go --env dev
```

- 编译二进制
```bash
go build -o bin/app main.go
./bin/app --env dev
```

### Docker 使用
镜像使用多阶段构建，并在容器内执行 `/home/app/start.sh`（默认启动 `./main`）。Dockerfile 暴露端口 `28080`，请按需映射。

1) 构建镜像
```bash
docker build -t tech-muyi-base-go:latest .
```

2) 运行容器（映射端口与配置）
```bash
# 使用容器内默认配置（若镜像包含 app/*.conf）
docker run -d --name tech-muyi-go -p 28080:28080 tech-muyi-base-go:latest

# 或挂载本地配置文件到容器内
docker run -d --name tech-muyi-go \
  -p 28080:28080 \
  -v $(pwd)/app/app-dev.conf:/home/app/app-dev.conf \
  tech-muyi-base-go:latest
```

3) 访问
```bash
curl http://127.0.0.1:28080/
```

注意：容器内启动脚本为 Linux shell；在 Windows 本地直接双击 `start.sh` 可能不可用，建议用 `go run` 或 `go build` 启动。

### 已内置路由（`main.go`）
- `GET /` 欢迎信息
- `GET /api/v1/system/health` 健康检查
- `GET /api/v1/system/config` 当前安全配置
- `GET /api/v1/system/info` 系统信息
- `GET /api/v1/test/ping` 测试连通
- `POST /api/v1/test/echo` 回显请求体
- `GET /api/v1/test/error` 返回测试错误

### 常见问题
- 端口不一致：本地默认端口由配置 `server.port` 决定；Dockerfile 暴露 `28080`，请对应映射。
- 配置未生效：检查 `--config` 路径或 `--env` 是否正确；程序优先读取当前目录下配置文件，其次 `app/` 目录。
- 日志不输出到控制台：将 `log.stdout` 设为 `true`。

### 许可
本项目示例代码可自由使用与修改，具体以仓库实际 LICENSE 为准。


