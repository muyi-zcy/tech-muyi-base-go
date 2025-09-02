# Tech MuYi Go 基础包

一个基于 Go 语言的企业级 Web 服务基础框架，集成了 Gin、MySQL、Redis、Zap 日志等常用组件，提供开箱即用的基础设施和最佳实践。

## 🚀 特性

- **Web 框架**: 基于 Gin 的高性能 HTTP 框架
- **配置管理**: 使用 Viper 支持多环境配置
- **数据库**: MySQL 数据库集成
- **缓存**: Redis 缓存支持
- **日志系统**: 基于 Zap 的结构化日志
- **健康检查**: 内置健康检查接口
- **统一响应**: 标准化的 API 响应格式
- **异常处理**: 全局异常处理中间件
- **ID 生成**: 分布式 ID 生成器

## 📋 系统要求

- Go 1.22.2 或更高版本
- MySQL 5.7+ 或 8.0+
- Redis 6.0+

## 🛠️ 安装和运行

### 1. 克隆项目

```bash
git clone <repository-url>
cd tech-muyi-base-go
```

### 2. 安装依赖

```bash
go mod tidy
```

### 3. 配置环境

项目支持多环境配置，默认使用 `app-dev.conf`：

```bash
# 开发环境
go run main.go -env=dev

# 本地环境
go run main.go -env=local

# 预发布环境
go run main.go -env=pre

# 生产环境
go run main.go -env=prod

# 指定配置文件
go run main.go -config=./custom.conf
```

### 4. 运行服务

```bash
go run main.go
```

服务将在配置的端口上启动（默认 8080）。

## 📁 项目结构

```
tech-muyi-base-go/
├── app/                    # 配置文件目录
│   ├── app-dev.conf       # 开发环境配置
│   ├── app-local.conf     # 本地环境配置
│   ├── app-pre.conf       # 预发布环境配置
│   └── app-prod.conf      # 生产环境配置
├── config/                 # 配置管理
│   └── config.go          # 配置结构和初始化
├── core/                   # 核心应用逻辑
│   ├── app.go             # 应用实例管理
│   ├── health.go          # 健康检查
│   └── starter.go         # 应用启动器
├── exception/              # 异常处理
│   ├── common_error_code.go # 通用错误码
│   └── exception.go       # 异常处理逻辑
├── id/                     # ID 生成器
│   └── myId.go            # 分布式 ID 生成
├── infrastructure/         # 基础设施
│   ├── database.go        # 数据库连接管理
│   └── redis.go           # Redis 连接管理
├── logger/                 # 日志系统
│   ├── logger.go          # 日志配置
│   └── myLog.go           # 日志工具
├── middleware/             # 中间件
│   ├── exception.go       # 异常处理中间件
│   └── logger.go          # 日志中间件
├── myContext/              # 上下文管理
│   └── context.go         # 自定义上下文
├── myResult/               # 响应结果
│   └── response.go        # 统一响应格式
├── utils/                  # 工具函数
│   ├── stringUtils.go     # 字符串工具
│   └── timeUtils.go       # 时间工具
├── main.go                 # 主程序入口
├── go.mod                  # Go 模块文件
└── README.md               # 项目说明文档
```

## 🔧 配置说明

### 基础配置

配置文件使用 TOML 格式，主要配置项包括：

```toml
# 应用配置
app_name = "tech-muyi-base-go"
version = "1.0.0"

# 服务器配置
[server]
port = 8080
mode = "debug"

# 日志配置
[log]
level = "info"
filename = "logs/app.log"
maxsize = 100
maxage = 30
maxbackups = 10
compress = true
stdout = true

# 数据库配置
[database]
driver = "mysql"
host = "localhost"
port = 3306
username = "root"
password = "password"
database = "test"
max_open_conns = 100
max_idle_conns = 10
conn_max_lifetime = 3600

# Redis配置
[redis]
host = "localhost"
port = 6379
password = ""
db = 0
```

## 📡 API 接口文档

### 基础接口

#### 1. 欢迎页面
- **GET** `/`
- **描述**: 服务欢迎页面
- **响应**: 欢迎信息

#### 2. 健康检查
- **GET** `/api/v1/system/health`
- **描述**: 服务健康状态检查
- **响应**: 健康状态信息

#### 3. 系统信息
- **GET** `/api/v1/system/info`
- **描述**: 获取系统基本信息
- **响应**: 系统版本、框架等信息

#### 4. 配置信息
- **GET** `/api/v1/system/config`
- **描述**: 获取安全配置信息
- **响应**: 应用配置（不包含敏感信息）

### 用户管理接口

#### 1. 获取用户列表
- **GET** `/api/v1/users`
- **描述**: 获取所有用户列表
- **响应**: 用户列表和分页信息

#### 2. 获取单个用户
- **GET** `/api/v1/users/:id`
- **描述**: 根据ID获取用户信息
- **参数**: `id` - 用户ID
- **响应**: 用户详细信息

#### 3. 创建用户
- **POST** `/api/v1/users`
- **描述**: 创建新用户
- **请求体**: 用户信息JSON
- **响应**: 创建成功的用户信息

#### 4. 更新用户
- **PUT** `/api/v1/users/:id`
- **描述**: 更新指定用户信息
- **参数**: `id` - 用户ID
- **请求体**: 更新的用户信息JSON
- **响应**: 更新后的用户信息

#### 5. 删除用户
- **DELETE** `/api/v1/users/:id`
- **描述**: 删除指定用户
- **参数**: `id` - 用户ID
- **响应**: 删除成功信息

### 测试接口

#### 1. Ping 测试
- **GET** `/api/v1/test/ping`
- **描述**: 服务连通性测试
- **响应**: pong 响应

#### 2. Echo 测试
- **POST** `/api/v1/test/echo`
- **描述**: 回显请求数据
- **请求体**: 任意JSON数据
- **响应**: 回显的请求数据

#### 3. 错误测试
- **GET** `/api/v1/test/error`
- **描述**: 测试错误处理
- **响应**: 错误信息

## 📊 响应格式

所有API接口都使用统一的响应格式：

### 成功响应
```json
{
  "code": "200",
  "success": true,
  "message": "success",
  "data": {...},
  "query": null
}
```

### 失败响应
```json
{
  "code": "500",
  "success": false,
  "message": "错误描述",
  "data": null,
  "query": null
}
```

### 分页响应
```json
{
  "code": "200",
  "success": true,
  "message": "success",
  "data": [...],
  "query": {
    "size": 10,
    "current": 1,
    "total": 100
  }
}
```

## 🧪 测试

### 运行测试
```bash
go test ./...
```

### 测试覆盖率
```bash
go test -cover ./...
```

## 📝 日志

项目使用 Zap 日志库，支持结构化日志和多种输出格式：

- 控制台输出
- 文件输出（支持日志轮转）
- 结构化字段
- 不同日志级别

## 🔒 安全特性

- 配置信息脱敏
- 全局异常处理
- 请求参数验证
- 日志安全记录

## 🚀 部署

### Docker 部署

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/app/ ./app/
EXPOSE 8080
CMD ["./main"]
```

### 构建二进制文件

```bash
go build -o tech-muyi-base-go main.go
```

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

本项目采用 MIT 许可证。

## 📞 联系方式

如有问题，请通过以下方式联系：

- 提交 Issue
- 发送邮件
- 项目讨论区

---

**Tech MuYi Go 基础包** - 让 Go 开发更简单、更高效！
