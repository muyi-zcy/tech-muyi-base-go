# 平台鉴权体系（myAuth + myContext）

本文档描述 `tech-muyi-base-go` 中 **HTTP/gRPC 鉴权** 与 **请求上下文传播** 的架构、配置与使用方式。

相关包：

| 包 | 职责 |
|----|------|
| `myAuth` | Token 签发/校验、Session、中间件、权限钩子、SessionEnricher |
| `myContext` | traceId / token / ssoId 传播、GORM 审计字段、RPC metadata |

---

## 目录

1. [架构概览](#1-架构概览)
2. [安全模型](#2-安全模型)
3. [请求生命周期](#3-请求生命周期)
4. [配置](#4-配置)
5. [Token Provider](#5-token-provider)
6. [HTTP 鉴权](#6-http-鉴权)
7. [gRPC 鉴权](#7-grpc-鉴权)
8. [Session 与 Extras](#8-session-与-extras)
9. [SessionEnricher](#9-sessionenricher)
10. [权限校验](#10-权限校验)
11. [上下文 API 速查](#11-上下文-api-速查)
12. [业务接入指南](#12-业务接入指南)
13. [错误码](#13-错误码)
14. [最佳实践](#14-最佳实践)
15. [常见问题](#15-常见问题)

---

## 1. 架构概览

```
┌─────────────────────────────────────────────────────────────────┐
│ Layer 0 · myContext HTTP Ingress                                │
│   traceId（必生成） + token（可选预置，来自 x-token）            │
│   不读取 x-sso-id，不写入默认 ssoId                              │
└────────────────────────────┬────────────────────────────────────┘
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│ Layer 1 · myAuth                                                │
│   ExtractToken → TokenProvider.Parse → SessionEnricher          │
│   bindAuthContext → 写入 Session + ssoId + token                │
└────────────────────────────┬────────────────────────────────────┘
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│ Layer 2 · 业务 / 基础设施                                       │
│   身份/权限：myAuth.GetSession                                   │
│   审计/GORM：myContext.ResolveActor                              │
│   RPC 出站：myContext metadata 透传                              │
└─────────────────────────────────────────────────────────────────┘
```

**职责边界：**

- **标量上下文**（traceId、ssoId、token）→ 仅 `myContext` 读写
- **会话对象**（Session 及 extras）→ 仅 `myAuth` 读写
- **ssoId 权威来源** → Token 校验成功后，由 `Session.UserID` 派生

---

## 2. 安全模型

### 2.1 核心原则

| 原则 | 说明 |
|------|------|
| **只信任 Token** | 客户端/网关可传 `x-token`；服务端校验后才认定用户身份 |
| **不信任裸 x-sso-id** | HTTP Ingress 与 gRPC `WithMetadata` **均不读取** metadata/header 中的 `x-sso-id` |
| **ssoId 两态** | 有值 = 已鉴权用户 ID 字符串；无值 = 匿名（审计 fallback 为 `system`） |
| **废除 `-1` 伪用户** | 未鉴权请求不再写入虚假 ssoId |

### 2.2 传输头约定

| Header / Cookie | 方向 | 含义 |
|-----------------|------|------|
| `x-trace-id` | 双向 | 链路追踪 ID |
| `x-token` | 请求 → 服务 | 登录凭证（Header 或 Cookie） |
| `x-sso-id` | 响应 / RPC 出站 | **仅**在服务端鉴权后随 metadata 带出，不可作为入站身份依据 |

Token 提取顺序：`Header x-token` → `Cookie x-token`；支持 `Bearer <token>` 前缀（`NormalizeToken`）。

---

## 3. 请求生命周期

### 3.1 HTTP 中间件顺序

`core.Initialize()` 默认注册：

```
1. myContext.ContextMiddleware()   // traceId + 可选 token 预置
2. myContext.LocaleMiddleware()
3. middleware.ExceptionHandler()
4. middleware.Logger()
```

业务路由组再挂载：

```
myAuth.Required(...)   // 或 Optional
myAuth.Permission(...) // 可选，须在 Required 之后
```

### 3.2 鉴权中间件流程

```
白名单命中？ ──是──► 放行（无 Session）
     │
     否
     ▼
BeforeAuthHook？ ──skip──► 放行
     │
     否
     ▼
提取 token
     │
     ├─ 无 token + Required  ──► token_missing
     ├─ 无 token + Optional ──► 放行
     ▼
TokenProvider.Parse
     │
     ├─ 失败/无效 + Required  ──► token_invalid
     ▼
SessionEnricher（可选）
     ▼
bindAuthContext → Session + ssoId + token 写入上下文
     ▼
c.Next()
```

### 3.3 gRPC 拦截器链

默认 RPC Server 链（`infrastructure/rpc`）：

```
Recovery → ContextExtract → [RegisterGRPCAuth 等] → Logging → ErrorMapping
```

- `ContextExtract`：从 metadata 恢复 traceId、token（**不含** ssoId）
- `GRPCAuthInterceptor`：有 token 时校验并写入 Session/ssoId

---

## 4. 配置

配置段 `[auth]`，通过 `myAuth.MustInitFromViper()` 或 `myAuth.Init(cfg)` 加载。

### 4.1 完整示例（TOML）

```toml
[auth]
provider = "encrypted_jwt"          # 或 redis_opaque
tokenExpireSeconds = 28800          # 默认 8 小时
whiteList = [
  "/api/user/v1/auth/login",
  "/api/user/v1/open/**",
]

[auth.jwt]
key = "0123456789abcdef0123456789abcdef"   # 32 字节或 base64
issuer = "my-xi"
audience = ""                               # 可选

[auth.session]
keyPrefix = "auth:session:"                 # redis_opaque 会话前缀
revokePrefix = "auth:revoked:"              # 吊销 JTI 前缀（JWT 也使用）
```

### 4.2 配置项说明

| 字段 | 类型 | 默认 | 说明 |
|------|------|------|------|
| `provider` | string | `encrypted_jwt` | Token 实现，见 [Token Provider](#5-token-provider) |
| `tokenExpireSeconds` | int | `28800` | Token/Session TTL（秒） |
| `whiteList` | []string | `[]` | 全局白名单路径，支持 `*`、`**` |
| `auth.jwt.key` | string | — | JWE 加密密钥（32 字节） |
| `auth.jwt.issuer` | string | `my-xi` | JWT iss |
| `auth.jwt.audience` | string | — | JWT aud（可选） |
| `auth.session.keyPrefix` | string | `auth:session:` | Redis  opaque token 存储前缀 |
| `auth.session.revokePrefix` | string | `auth:revoked:` | 吊销列表前缀 |

### 4.3 初始化时机

```go
func main() {
    starter, _ := core.Initialize()

    myAuth.MustInitFromViper()   // 必须在路由 / gRPC 注册前
    auth.Setup()                 // 注册 SessionEnricher（业务包）

    routes.Register(starter.GetAPIGroup())

    myAuth.RegisterGRPCAuth(     // 可选：gRPC 需要 Session 时
        myAuth.WithGRPCEnrichers("user.profile"),
    )

    starter.Run()
}
```

---

## 5. Token Provider

### 5.1 内置 Provider

| 名称 | 常量 | 说明 |
|------|------|------|
| 加密 JWT | `encrypted_jwt` | JWE（A256GCM）+ 可选 Redis 吊销 JTI |
| Redis Opaque | `redis_opaque` | 随机 token 作 key，Session JSON 存 Redis |

### 5.2 encrypted_jwt

- 签发：`Issue(ctx, claims)` → JWE 字符串
- 解析：解密 → 校验 exp/iss → 检查 Redis 吊销
- 吊销：`Revoke` 将 JTI 写入 Redis（TTL = 剩余有效期）

Claims 标准字段：

```go
type Claims struct {
    UserID      int64
    Username    string
    DisplayName string
    ExpireAt    time.Time
    JTI         string
}
```

### 5.3 redis_opaque

- 签发：生成 UUID token → Redis `keyPrefix + token`
- 解析：Redis GET → 反序列化 Session
- 吊销：DELETE key + 写入 revoke 前缀

### 5.4 自定义 Provider

```go
type TokenProvider interface {
    Name() string
    Issue(ctx context.Context, claims *Claims) (token string, err error)
    Parse(ctx context.Context, token string) (*Claims, error)
    Revoke(ctx context.Context, claims *Claims) error
}

// 注册（Init 前或 provider 名称与配置一致时）
myAuth.RegisterTokenProvider(myProvider)
```

---

## 6. HTTP 鉴权

### 6.1 中间件 API

```go
// 必须登录
secured.Use(myAuth.Required(opts...))

// 有 token 则加载 Session，无 token 放行
optional.Use(myAuth.Optional(opts...))

// 权限（须在 Required 之后）
userGroup.Use(myAuth.Permission("user:manage:edit", checkFn, "user.auth.permission_denied"))
```

> `Required`、白名单、`Optional` 及「不挂中间件」的差异与选型，见 [6.5 节](#65-required白名单与-optional-的区别)。

### 6.2 中间件 Option

| Option | 说明 |
|--------|------|
| `WithProvider(name)` | 指定 TokenProvider 名称 |
| `WithProviderSelector(fn)` | 按请求动态选择 Provider |
| `WithWhiteList(paths...)` | 路由组级白名单（叠加全局） |
| `WithTokenMissingCode(code)` | 自定义 token 缺失错误码 |
| `WithTokenInvalidCode(code)` | 自定义 token 无效错误码 |
| `WithEnrichers(names...)` | 鉴权后执行的 SessionEnricher |
| `WithBeforeAuth(hook)` | 鉴权前 Hook（可 skip） |

### 6.3 路由示例（my-xi-user）

```go
authOpts := []myAuth.Option{
    myAuth.WithEnrichers("user.profile"),
    myAuth.WithTokenMissingCode("user.auth.token_missing"),
    myAuth.WithTokenInvalidCode("user.auth.token_invalid"),
}

v1 := apiGroup.Group("/v1")

// 公开
v1.POST("/auth/login", authCtrl.Login)

// 受保护
secured := v1.Group("")
secured.Use(myAuth.Required(authOpts...))
{
    secured.POST("/auth/logout", authCtrl.Logout)
    secured.GET("/auth/getCurrentUser", authCtrl.GetCurrentUser)

    userGroup := secured.Group("/user")
    userGroup.Use(myAuth.Permission("user:manage:edit", auth.CheckPerm, "user.auth.permission_denied"))
    {
        userGroup.POST("/save", userCtrl.Save)
    }
}
```

### 6.4 Optional 示例（my-xi-file）

公开读接口有 token 则识别用户，无 token 也可访问：

```go
optional := v1.Group("")
optional.Use(myAuth.Optional(authOpts...))
optional.GET("/getAccessUrl", fileCtrl.GetAccessUrl)
```

### 6.5 Required、白名单与 Optional 的区别

`myAuth.Required` 和**白名单**不是二选一，而是**同一条鉴权链路里的两种机制**：前者决定「要不要验 token」，后者决定「在已挂鉴权的路由组里，哪些路径可以跳过验 token」。

#### 6.5.1 `Required` 是什么

`Required` 是挂在路由组上的 **Gin 中间件**。请求进入该组后，中间件按以下顺序处理：

1. 路径是否命中白名单？→ 是则直接放行（见下文）
2. 提取并校验 `x-token`
3. 无 token / token 无效 → 返回错误（Required 模式）
4. 校验成功 → SessionEnricher → `bindAuthContext` → 放行

| 情况 | Required 行为 |
|------|---------------|
| 无 token | 返回 `token_missing` |
| token 无效 | 返回 `token_invalid` |
| token 有效 | 写入 Session / ssoId / token 到上下文 |

只有路由**挂上了** `Required`（或 `Optional`），才会进入上述流程。

#### 6.5.2 白名单是什么

白名单是 **Required / Optional 内部的例外规则**。路径命中后**直接 `c.Next()`**：

- 不检查 token
- 不加载 Session
- 不执行 Enricher
- 不报错

白名单有两个来源（**满足任一即跳过**）：

| 来源 | 配置方式 | 作用域 |
|------|----------|--------|
| **全局白名单** | 配置文件 `[auth].whiteList` | 所有挂了 `Required` / `Optional` 的路由组 |
| **路由组白名单** | `myAuth.WithWhiteList("/path/**")` | 仅该中间件实例 |

路径匹配规则（`NormalizePath` 后）：

- 精确匹配：`/api/user/v1/auth/login`
- 单段通配：`/api/user/v1/open/*`
- 多级通配：`/api/user/v1/open/**`

#### 6.5.3 三者对比

| | `Required` | 白名单 | `Optional` |
|---|-----------|--------|------------|
| **本质** | 鉴权中间件（强制登录） | 鉴权中间件内的**跳过规则** | 鉴权中间件（宽松模式） |
| **关系** | 父级约束 | Required/Optional 的子级例外 | Required 的替代模式 |
| **无 token** | 报错 | 直接放行 | 放行 |
| **有 token** | 校验并写 Session | **仍不校验**（整段跳过） | 校验，成功则写 Session |
| **无效 token** | 报错 | 放行（不报错） | 放行（不报错） |
| **Session** | 校验成功后写入 | 不写入 | 有有效 token 时写入 |

> **注意**：白名单与 `Optional` 不同。白名单是「完全跳过鉴权逻辑」；`Optional` 是「没 token 可以访问，有 token 则尽量识别用户」。

#### 6.5.4 与「不挂 Required」的区别

实现「无需登录」还有第三种方式：**路由根本不挂鉴权中间件**。

`my-xi-user` 同时使用了两种方式：

```go
// 方式 A：不挂 Required —— login、open 等公开接口
v1.POST("/auth/login", authCtrl.Login)

openGroup := v1.Group("/open")
{
    openGroup.POST("/checkPermission", openCtrl.CheckPermission)
}

// 方式 B：挂 Required —— 业务接口必须登录
secured := v1.Group("")
secured.Use(myAuth.Required(authOpts...))
```

| 方式 | 是否进入鉴权中间件 | 典型场景 |
|------|-------------------|----------|
| **不挂 Required** | 否 | 登录、注册、open API（**推荐**，结构清晰） |
| **Required + 白名单** | 是，但命中路径跳过 | 大路由组统一挂 Required，少数路径例外 |
| **Optional** | 是，无 token 也放行 | 公开读接口，有 token 时识别用户（如 file 的 getAccessUrl） |

效果上「不挂 Required」与「白名单跳过」对客户端都是无需 token，但：

- **不挂 Required**：路由层就不存在鉴权逻辑，语义最明确
- **白名单**：路由仍在 `secured` 组内，靠路径匹配绕过，适合全局统一配置或同组内个别例外

#### 6.5.5 选型建议

```
登录 / 注册 / open API     → 单独路由组，不挂 Required（推荐）
大部分业务 API             → secured.Use(myAuth.Required(...))
同一 secured 组内个别公开  → WithWhiteList 或配置 auth.whiteList
有 token 更好、无 token 也行 → myAuth.Optional
```

### 6.6 登录 / 登出

**签发 Token（Service 层）：**

```go
sess, token, err := myAuth.Manager().Create(ctx, &myAuth.SessionInput{
    UserID:      user.Id,
    Username:    user.Username,
    DisplayName: user.DisplayName,
    TTL:         ttl, // 可选，覆盖配置
})
// 将 token 返回给客户端（LoginResp.Token）
```

**登出：**

```go
_ = myAuth.Manager().Destroy(ctx, myAuth.Token(c))
```

**客户端携带 Token：**

```
Header: x-token: <token>
或 Cookie: x-token=<token>
```

---

## 7. gRPC 鉴权

### 7.1 注册

```go
myAuth.MustInitFromViper()
auth.Setup()

myAuth.RegisterGRPCAuth(
    myAuth.WithGRPCEnrichers("user.profile"),
    // myAuth.WithGRPCProvider("encrypted_jwt"), // 可选，默认用配置 provider
)
```

须在 `starter.Run()` **之前**调用。Server 在 `Start()` 时构建，可在此前注册拦截器。

### 7.2 拦截器行为

| 条件 | 行为 |
|------|------|
| metadata 无 token | 直接 `handler(ctx, req)`，无 Session |
| token 无效 | 同上（不阻断，由业务决定是否 Require） |
| token 有效 | 执行 Enricher → 写入 Session/ssoId/token |

### 7.3 Handler 读取 Session

```go
func (s *registerServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterReply, error) {
    sess, ok := myAuth.SessionFromContext(ctx)
    if ok {
        _ = sess.UserID
    }
    // ...
}
```

### 7.4 客户端透传

出站拦截器（`interceptor.ContextInject`）自动携带：

- `x-trace-id`
- `x-token`（若 context 中有）
- `x-sso-id`（**仅**鉴权后 context 中已有 ssoId 时）

---

## 8. Session 与 Extras

### 8.1 Session 结构

```go
type Session struct {
    UserID      int64
    Username    string
    DisplayName string
    ExpireAt    time.Time
    Token       string
    JTI         string
    // extras 通过 SetExtra / Extra 访问
}
```

平台 Session **只存身份字段**；部门、权限等业务数据放入 `extras`。

### 8.2 读取 API

```go
// HTTP Handler
sess, ok := myAuth.GetSession(c)
sess := myAuth.MustSession(c) // 不存在 panic

// Service / gRPC（使用 c.Request.Context() 或 ctx）
sess, ok := myAuth.SessionFromContext(ctx)

// 泛型 extras
deptID, ok := myAuth.SessionExtra[int64](c, auth.ExtraDeptID)
permSet, ok := myAuth.SessionExtra[*auth.PermSet](c, auth.ExtraPermSet)
```

### 8.3 写入 extras

仅在 **SessionEnricher** 或 **登录流程** 中写入：

```go
sess.SetExtra("permSet", permSet)
```

不要在 Controller 中随意 `SetExtra`（Session 已绑定上下文后 extras 不应再变）。

---

## 9. SessionEnricher

鉴权成功后、写入上下文之前，按名称链式执行。

### 9.1 注册

```go
// my-xi-user/auth/setup.go
func Setup() {
    myAuth.RegisterSessionEnricher("user.profile", enrichUserProfile)
}

func enrichUserProfile(ctx context.Context, sess *myAuth.Session) error {
    profile, err := LoadUserProfile(ctx, sess.UserID)
    if err != nil {
        return err
    }
    sess.SetExtra(ExtraDeptID, profile.DeptID)
    sess.SetExtra(ExtraPermSet, permSet)
    return nil
}
```

### 9.2 使用

HTTP：

```go
myAuth.Required(myAuth.WithEnrichers("user.profile"))
```

gRPC：

```go
myAuth.RegisterGRPCAuth(myAuth.WithGRPCEnrichers("user.profile"))
```

未注册的 enricher 名称会导致鉴权失败（返回 error）。

---

## 10. 权限校验

`Permission` 中间件不内置 RBAC 模型，通过回调注入：

```go
type PermissionCheck func(sess *myAuth.Session, code string) bool

// my-xi-user/auth/accessors.go
func CheckPerm(sess *myAuth.Session, code string) bool {
    ps := PermSetFromSess(sess)
    return ps != nil && ps.Allow(code)
}
```

权限数据通常由 SessionEnricher 注入 `PermSet` extra。

---

## 11. 上下文 API 速查

### 11.1 读取约定

| 场景 | API | 未鉴权行为 |
|------|-----|------------|
| 用户身份 / 权限 | `myAuth.GetSession` / `SessionFromContext` | 无 Session |
| 必须登录的操作 | `myAuth.Required` + `myContext.RequireSsoId` | 返回 `platform.unauthorized` |
| GORM Creator/Operator | `myContext.ResolveActor` | 返回 `"system"` |
| 日志字段 | `myContext.TryGetSsoId` | 空则不打 ssoId 字段 |
| RPC 出站 | `TryGetTraceId` / `TryGetToken` / `TryGetSsoId` | 空则不带 |

### 11.2 myContext 主要 API

```go
// traceId
myContext.TryGetTraceId(ctx)
myContext.GetTraceId(ctx)          // 缺失 → error

// ssoId
myContext.TryGetSsoId(ctx)         // 可能为空
myContext.RequireSsoId(ctx)        // 缺失 → platform.unauthorized
myContext.GetSsoId(ctx)            // RequireSsoId 别名
myContext.ResolveActor(ctx)        // 缺失 → "system"

// token
myContext.TryGetToken(ctx)
myContext.WithToken(ctx, token)

// 绑定（框架内部，业务一般不直接调用）
myContext.BindScalars(c, myContext.ScalarBinding{SsoId: "...", Token: "..."})
```

### 11.3 myAuth Token 辅助

```go
myAuth.ExtractToken(c)   // 从 Header/Cookie 提取
myAuth.Token(c)          // 优先 context，否则 ExtractToken
myAuth.NormalizeToken(raw)
```

---

## 12. 业务接入指南

### 12.1 新 HTTP 微服务（最小步骤）

```go
// main.go
starter, _ := core.Initialize()
myAuth.MustInitFromViper()
routes.Register(starter.GetAPIGroup())
starter.Run()
```

```go
// routes.go
secured := apiGroup.Group("/v1")
secured.Use(myAuth.Required(
    myAuth.WithTokenMissingCode("myservice.auth.token_missing"),
    myAuth.WithTokenInvalidCode("myservice.auth.token_invalid"),
))
secured.GET("/resource/get", ctrl.Get)
```

```go
// controller.go
func (ctrl *Controller) Get(c *gin.Context) {
    sess, ok := myAuth.GetSession(c)
    if !ok { /* Required 已保证，通常不会走到 */ }
    data, err := ctrl.svc.Get(c.Request.Context(), sess.UserID)
    // ...
}
```

### 12.2 带 Enricher 的服务（参考 my-xi-user）

1. `auth/setup.go` → `RegisterSessionEnricher`
2. `main.go` → `auth.Setup()` 在 `MustInitFromViper` 之后
3. 路由 → `WithEnrichers("your.enricher")`
4. 权限 → `Permission(code, CheckPerm, deniedCode)`

### 12.3 仅依赖用户 ID 的服务（参考 my-xi-desktop）

```go
// main.go
myAuth.MustInitFromViper()

// routes.go
secured.Use(myAuth.Required(
    myAuth.WithTokenMissingCode("desktop.auth.token_missing"),
    myAuth.WithTokenInvalidCode("desktop.auth.token_invalid"),
))

// service.go
sess, ok := myAuth.SessionFromContext(ctx)
userId := sess.UserID
```

无需自建 Redis Session 读取；与用户服务共用 `[auth]` 配置与 Token。

### 12.4 gRPC 服务

```go
myAuth.MustInitFromViper()
auth.Setup()
myAuth.RegisterGRPCAuth(myAuth.WithGRPCEnrichers("user.profile"))

starter.RegisterGrpcServices(func(s *grpc.Server) {
    pb.RegisterYourServiceServer(s, &yourServer{})
})
starter.Run()
```

### 12.5 自定义 TokenProvider 服务

1. 实现 `TokenProvider`
2. `myAuth.RegisterTokenProvider(p)`（`Init` 前）
3. 配置 `auth.provider = "<Name()>"`

---

## 13. 错误码

### 13.1 平台默认

| 常量 | 错误码 | 场景 |
|------|--------|------|
| `CodeTokenMissing` | `platform.auth.token_missing` | Required 且无 token |
| `CodeTokenInvalid` | `platform.auth.token_invalid` | token 解析/校验失败 |
| `CodePermissionDenied` | `platform.auth.permission_denied` | Permission 不通过 |

### 13.2 业务覆盖示例

| 服务 | token_missing | token_invalid |
|------|---------------|---------------|
| user | `user.auth.token_missing` | `user.auth.token_invalid` |
| file | `file.auth.token_missing` | `file.auth.token_invalid` |
| desktop | `desktop.auth.token_missing` | `desktop.auth.token_invalid` |

通过 `WithTokenMissingCode` / `WithTokenInvalidCode` 覆盖。

---

## 14. 最佳实践

1. **中间件顺序**：`ContextMiddleware` 必须在 `myAuth` 之前（`core` 默认已满足）。
2. **Service 层传 `ctx`**：Controller 使用 `c.Request.Context()`，保证 traceId/token/ssoId/Session 向下传递。
3. **不要读 Header ssoId**：统一 `GetSession` 或 `RequireSsoId` / `ResolveActor`。
4. **公开 API 选型**：登录/open 等优先**不挂 Required**；同组内少数例外用白名单；需要「有 token 识别用户、无 token 也可访问」用 `Optional`（见 [6.5 节](#65-required白名单与-optional-的区别)）。
5. **权限放 extras**：Session 保持平台字段稳定，业务扩展走 Enricher。
6. **登出调 Destroy**：确保 JTI/Redis 会话被吊销。
7. **gRPC 需要 Session 时注册 `RegisterGRPCAuth`**：否则仅有 token 字符串，无 Session 对象。
8. **多服务共享 auth 配置**：同一 `auth.jwt.key` 或 Redis 前缀，Token 可跨服务校验。

---

## 15. 常见问题

### Q1：未登录时 GORM 的 Creator/Operator 是什么？

`myContext.ResolveActor(ctx)` 返回 `"system"`。不再使用 `"-1"`。

### Q2：Optional 路由无 token 时 ssoId 是什么？

空。不会写入虚假 ID。

### Q3：网关透传 x-sso-id 能否识别用户？

**不能。** 必须传 `x-token`，由服务端校验。ssoId 仅作为鉴权后的出站/日志字段。

### Q4：HTTP 鉴权后 RPC 调用会带身份吗？

会。`ContextInject` 携带 context 中的 `x-token`；若已鉴权还会带 `x-sso-id`。对端需 `RegisterGRPCAuth` 才能还原 Session。

### Q5：`Required` 和白名单有什么区别？

二者不是二选一。**`Required` 是鉴权中间件；白名单是该中间件内的跳过规则。**

- 挂了 `Required` 且无 token → 报错
- 挂了 `Required` 且路径命中白名单 → 直接放行，不验 token、无 Session
- 没挂 `Required` → 根本不会进鉴权逻辑（与白名单无关，推荐用于 login/open）

详见 [6.5 Required、白名单与 Optional 的区别](#65-required白名单与-optional-的区别)。

### Q6：JWT key 格式？

32 字节原始字符串，或 Base64 编码的 32 字节。见 `JWTConfig.decodeKey()`。

### Q7：RegisterGRPCAuth 调用时机？

`myAuth.MustInitFromViper()` 之后、`starter.Run()` 之前。

---

## 附录：包文件索引

```
myAuth/
  config.go       配置与 Init
  manager.go      SessionManager（Create/Load/Destroy）
  middleware.go   Required/Optional/Permission
  context.go      Session 存取、bindAuthContext
  token.go        ExtractToken/Token
  session.go      Session 模型
  claims.go       Claims
  provider.go     TokenProvider 接口
  provider_impl.go encrypted_jwt / redis_opaque
  store_redis.go  Redis 存储与吊销
  enricher.go     SessionEnricher
  hooks.go        BeforeAuthHook
  whitelist.go    白名单匹配
  grpc.go         GRPCAuthInterceptor / RegisterGRPCAuth

myContext/
  keys.go         typed key + header 常量
  bind.go         BindScalars / BindTrace
  trace.go        traceId
  actor.go        ssoId / ResolveActor
  token.go        token
  context.go      HTTPIngressMiddleware
  rpc.go          gRPC metadata
  locale.go       locale（独立）
```
