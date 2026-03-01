---
name: tech-muyi-base-go-api
description: Controller 层规范、myResult 统一返回、myContext 获取 traceId/ssoId、接口响应格式。使用场景：编写 HTTP 接口、统一成功/失败返回、分页查询返回、需要 traceId/ssoId 的请求链路。
---

# tech-muyi-base-go API 层规范

基于 `github.com/muyi-zcy/tech-muyi-base-go` 的 myResult、myContext 实现 Controller 层与统一返回。

## 使用场景

- 用户编写或规范 Controller、接口返回格式
- 用户需要 traceId、ssoId 透传或从 context 获取
- 用户需要分页查询、成功/失败统一响应

## MyResult 统一返回结构

```go
type MyResult struct {
	Code    string      `json:"code"`
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Query   *MyQuery    `json:"query"`  // 分页时使用
}
```

## 成功返回

### 静态方法（推荐）

```go
import "github.com/muyi-zcy/tech-muyi-base-go/myResult"

// 基础成功
myResult.Success(c, data)

// 自定义 message
myResult.SuccessWithMessage(c, "操作成功", data)

// 带分页信息
query := &myResult.MyQuery{Size: 20, Current: 1, Total: 100}
myResult.SuccessWithQuery(c, list, query)
```

### 构建 MyResult 后 JSON 返回

```go
result := myResult.Ok(data)
result := myResult.OkWithMessage("成功", data)
result := myResult.OkWithQuery(data, query)
myResult.JSON(c, result)
```

## 失败返回

### 通过 c.Error(err) 由中间件处理

```go
c.Error(myException.NewExceptionFromError(myException.NOT_FOUND))
return
```

中间件会识别 `MyException`、`ValidationError`、`NotFoundError` 并返回对应 JSON。

### 手动返回 JSON

```go
myResult.Error(c, "系统错误")
myResult.ErrorWithCode(c, "10001", "参数错误")
myResult.ErrorWithError(c, err)

myResult.BadRequestResponse(c, "参数无效")
myResult.UnauthorizedResponse(c, "未登录")
myResult.NotFoundResponse(c, "资源不存在")
```

## MyQuery 分页

```go
type MyQuery struct {
	Size    int   `json:"size"`
	Current int   `json:"current"`
	Total   int64 `json:"total"`
}

// 方法
query.GetSize()    // 默认 20，最大 2000
query.GetCurrent() // 默认 1
query.GetOffset()  // (Current-1)*Size
query.SetTotal(n)  // 设置总数
```

## myContext：traceId / ssoId

### 从 Gin 上下文获取

```go
import "github.com/muyi-zcy/tech-muyi-base-go/myContext"

traceId := myContext.GetTraceIdFromGinCtx(c)
ssoId := myContext.GetSsoIdFromGinCtx(c)
```

### 从标准 context 获取（Service/Repository 层）

```go
traceId, err := myContext.GetTraceId(ctx)
ssoId, err := myContext.GetSsoId(ctx)
```

### 透传来源

- Header：`x-trace-id`、`x-sso-id`
- Cookie：`x-trace-id`、`x-sso-id`
- 若无则自动生成 UUID（traceId）或 `-1`（ssoId）

`ContextMiddleware` 会注入到 `c.Request.Context()`，供全链路使用。

## Controller 示例（路径使用小驼峰，统一通过 myResult 处理）

```go
package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/muyi-zcy/tech-muyi-base-go/myResult"
	"github.com/muyi-zcy/tech-muyi-base-go/myException"
)

type UserController struct{}

// 路由定义示例：路径段使用小驼峰（lowerCamelCase）
// router.GET("/api/v1/user/getUserInfo", userController.GetUserInfo)
// router.GET("/api/v1/user/listUsers", userController.ListUsers)

func (u *UserController) GetUserInfo(c *gin.Context) {
	id := c.Param("id")
	user, err := userService.GetById(c.Request.Context(), id)
	if err != nil {
		// 异常统一交给 myResult 处理
		myResult.ErrorWithError(c, err)
		return
	}
	// 正常返回也统一使用 myResult
	myResult.Success(c, user)
}

func (u *UserController) ListUsers(c *gin.Context) {
	query := &myResult.MyQuery{}
	if err := c.ShouldBindQuery(query); err != nil {
		myResult.ErrorWithError(c, myException.NewValidationError("query", "分页参数错误"))
		return
	}
	list, total, err := userService.GetPage(c.Request.Context(), query)
	if err != nil {
		myResult.ErrorWithError(c, err)
		return
	}
	query.SetTotal(total)
	myResult.SuccessWithQuery(c, list, query)
}
```

## 响应约定

- HTTP 状态码：接口统一返回 200，错误信息在 body 的 `code`、`message`、`success` 中体现
- 成功：`success: true`，`code: "200"`
- 失败：`success: false`，`code` 为错误码，`message` 为错误描述
