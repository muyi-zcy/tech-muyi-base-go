# tech-muyi-base-go

一个基于 Go 语言的企业级 Web 服务基础框架，集成了 Gin、GORM、MySQL、Redis、Zap 日志等常用组件，提供开箱即用的基础设施和最佳实践。

## 🚀 特性

- **Web 框架**: 基于 Gin 的高性能 HTTP 框架
- **ORM框架**: 集成 GORM，提供便捷的数据库操作
- **配置管理**: 使用 Viper 支持多环境配置
- **数据库**: MySQL 数据库集成
- **缓存**: Redis 缓存支持
- **日志系统**: 基于 Zap 的结构化日志，自动携带 traceId
- **健康检查**: 内置健康检查接口
- **统一响应**: 标准化的 API 响应格式
- **异常处理**: 全局异常处理中间件，支持自定义业务异常注册
- **ID 生成**: 分布式 ID 生成器
- **上下文管理**: 自动管理请求上下文，支持 traceId 和 ssoId

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
├── controller/             # 控制器层
│   └── user.go            # 用户控制器示例
├── core/                   # 核心应用逻辑
│   ├── app.go             # 应用实例管理
│   ├── health.go          # 健康检查
│   └── starter.go         # 应用启动器
├── dao/                    # 数据访问层
│   └── user.go            # 用户数据访问对象
├── exception/              # 异常处理
│   ├── common_error_code.go # 通用错误码
│   └── exception.go       # 异常处理逻辑
├── id/                     # ID 生成器
│   └── myId.go            # 分布式 ID 生成
├── infrastructure/         # 基础设施
│   ├── database.go        # 数据库连接管理（已集成GORM）
│   └── redis.go           # Redis 连接管理
├── logger/                 # 日志系统
│   ├── logger.go          # 日志配置
│   └── myLog.go           # 日志工具
├── middleware/             # 中间件
│   ├── exception.go       # 异常处理中间件
│   └── logger.go          # 日志中间件
├── model/                  # 数据模型
│   └── user.go            # 用户模型示例
├── myContext/              # 上下文管理
│   └── context.go         # 自定义上下文
├── myResult/               # 响应结果
│   └── response.go        # 统一响应格式
├── service/                # 服务层
│   └── user.go            # 用户服务示例
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

### 用户管理接口（GORM示例）

#### 1. 获取用户列表
- **GET** `/api/v1/users`
- **描述**: 获取所有用户列表（使用GORM查询）
- **响应**: 用户列表

#### 2. 获取单个用户
- **GET** `/api/v1/users/:id`
- **描述**: 根据ID获取用户信息（使用GORM查询）
- **参数**: `id` - 用户ID
- **响应**: 用户详细信息

#### 3. 创建用户
- **POST** `/api/v1/users`
- **描述**: 创建新用户（使用GORM插入）
- **请求体**: 用户信息JSON
- **响应**: 创建成功的用户信息

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
- 自动携带 traceId

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

## 🎯 核心功能详解

### 1. 上下文管理

项目提供了强大的上下文管理功能，自动处理请求中的 traceId 和 ssoId：

```go
// 自动从请求头或cookie中获取traceId
// 如果都获取不到，则生成新的traceId
// traceId在请求头或cookie中的key为：X-Trace-ID
// ssoId在请求头或cookie中的key为：X-SSO-ID

// 在业务代码中获取traceId
traceId := myContext.GetTraceId(ctx)

// 在业务代码中获取ssoId
ssoId := myContext.GetSsoId(ctx)
```

### 2. 日志系统

日志系统自动携带 traceId，便于问题追踪：

```go
// 业务日志记录时不需要手动处理traceId
// 系统会自动从上下文中提取并携带
logger.InfoCtx(ctx, "用户登录", zap.String("username", username))

// 也可以使用常规日志记录方法
logger.Info("系统启动完成")
```

### 3. 异常处理增强功能

#### 3.1 自定义业务异常注册

支持动态注册新的业务异常码，满足不同业务场景的需求：

```go
// 定义自定义错误码枚举
const (
    // 用户相关错误 20000-29999
    USER_NOT_FOUND CommonErrorCodeEnum = iota + 20000
    USER_ALREADY_EXISTS
    USER_PASSWORD_INVALID
    
    // 订单相关错误 30000-39999
    ORDER_NOT_FOUND
    ORDER_ALREADY_PAID
)

// 在init函数中注册自定义错误码
func init() {
    RegisterErrorCode(USER_NOT_FOUND, "20000", "用户不存在")
    RegisterErrorCode(USER_ALREADY_EXISTS, "20001", "用户已存在")
    RegisterErrorCode(USER_PASSWORD_INVALID, "20002", "密码错误")
    
    RegisterErrorCode(ORDER_NOT_FOUND, "30000", "订单不存在")
    RegisterErrorCode(ORDER_ALREADY_PAID, "30001", "订单已支付")
}

// 在业务代码中使用
if userNotFound {
    // 方式1：使用注册的错误码创建异常
    return nil, USER_NOT_FOUND.ToBusinessError()
}
```

#### 3.2 直接抛状态码异常

支持直接通过自定义状态码或HTTP状态码创建异常：

```go
// 方式1：通过自定义错误码创建异常
err := GetErrorCodeByCode("40001", "业务参数错误")
return nil, err

// 方式2：通过HTTP状态码创建异常
err := GetErrorCodeByHTTPStatus(404, "资源不存在")
panic(err) // 由异常处理中间件捕获
```

#### 3.3 返回异常状态码

提供多种方式返回异常响应：

```go
// 在控制器中使用
func GetUser(ctx *gin.Context) {
    user, err := userService.GetUser(id)
    if err != nil {
        // 方式1：使用myResult.ErrorWithError返回错误
        myResult.ErrorWithError(ctx, err)
        return
    }
    
    myResult.Success(ctx, user)
}

// 方式2：直接返回指定状态码的错误
func HandleNotFoundError(ctx *gin.Context) {
    myResult.NotFoundResponse(ctx, "用户不存在")
}

// 方式3：返回自定义错误码
func HandleCustomError(ctx *gin.Context) {
    myResult.ErrorWithCode(ctx, "20000", "用户不存在")
}
```

#### 3.4 异常处理中间件

全局异常处理中间件会自动捕获并处理各种类型的异常：

```go
// 异常处理中间件会自动处理以下类型的异常：
// 1. BusinessError - 业务异常
// 2. ValidationError - 验证异常
// 3. NotFoundError - 资源不存在异常
// 4. 其他标准error类型

// 中间件会根据异常类型返回对应的HTTP状态码和错误信息
```

## 📦 核心组件使用说明

### 上下文管理组件

上下文管理组件位于 `myContext` 包中，提供了完整的上下文管理机制：

1. **ContextMiddleware**: 上下文管理中间件，自动处理请求中的 traceId 和 ssoId
2. **GetTraceId**: 从上下文中获取 traceId
3. **GetSsoId**: 从上下文中获取 ssoId
4. **SetTraceId**: 设置上下文中的 traceId
5. **SetSsoId**: 设置上下文中的 ssoId

### 日志组件

日志组件位于 `logger` 包中，提供了强大的日志记录功能：

1. **InfoCtx/ErrorCtx/DebugCtx**: 带上下文的日志记录方法，自动携带 traceId
2. **Info/Error/Debug**: 常规日志记录方法
3. **InitWithConfig**: 根据配置初始化日志系统
4. **Logger**: 获取底层 zap.Logger 实例

### 异常处理组件

异常处理组件位于 `exception` 包中，提供了完整的异常处理机制：

1. **CommonErrorCodeEnum**: 预定义的HTTP状态码和系统级错误码
2. **BusinessError**: 业务异常类型
3. **ValidationError**: 验证异常类型
4. **NotFoundError**: 资源不存在异常类型
5. **RegisterErrorCode**: 注册自定义业务异常
6. **GetErrorCodeByCode**: 通过自定义错误码创建异常
7. **GetErrorCodeByHTTPStatus**: 通过HTTP状态码创建异常

### 响应结果组件

响应结果组件位于 `myResult` 包中，提供了统一的API响应格式：

1. **MyResult**: 统一响应结构体
2. **Success**: 成功响应
3. **ErrorWithError**: 错误响应
4. **ErrorWithCode**: 自定义错误码响应
5. **BadRequestResponse**: 400错误响应
6. **UnauthorizedResponse**: 401错误响应
7. **NotFoundResponse**: 404错误响应

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

本项目采用 MIT 许可证。

## 📞 联系方式

如有问题，请通过以下方式联系：

- 提交 Issue
- 发送邮件
- 项目讨论区