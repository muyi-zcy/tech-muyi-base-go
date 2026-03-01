---
name: tech-muyi-base-go-exception
description: 使用 tech-muyi-base-go 的 myException 进行错误处理，包括 MyException、错误码枚举、NewException、NewExceptionFromError、RegisterErrorCode、ValidationError、NotFoundError。使用场景：业务异常抛出、错误码定义、校验失败、资源不存在等错误处理。
---

# tech-muyi-base-go 异常处理

基于 `github.com/muyi-zcy/tech-muyi-base-go/myException` 进行统一错误处理。

## 使用场景

- 用户需要抛出业务异常、校验异常、404 等
- 用户询问错误码如何定义、如何注册自定义错误码
- 用户需要将 error 转为统一返回结构

## 异常类型

| 类型 | 用途 | 创建方式 |
|------|------|----------|
| `MyException` | 通用业务异常 | `NewException(code, msg)` / `NewExceptionFromError(enum)` |
| `ValidationError` | 参数校验失败 | `NewValidationError(field, message)` |
| `NotFoundError` | 资源不存在 | `NewNotFoundError(resource, id)` |

## 核心 API

### 创建异常

```go
import "github.com/muyi-zcy/tech-muyi-base-go/myException"

// 自定义 code + message
err := myException.NewException("10001", "用户不存在")

// 从错误码枚举创建
err := myException.NewExceptionFromError(myException.NOT_FOUND)
err := myException.NewExceptionFromError(myException.INVALID_PARAM)

// 校验异常（会被中间件映射为 400）
err := myException.NewValidationError("username", "用户名不能为空")

// 资源不存在（会被中间件映射为 404）
err := myException.NewNotFoundError("User", 123)
```

### 从 error 提取 code / message

```go
code := myException.GetErrorCode(err)
msg := myException.GetErrorMessage(err)
```

### 错误码枚举转 MyException

```go
err := myException.CommonErrorCodeEnum.INVALID_PARAM.ToBusinessError()
```

### 按 code / HTTP 状态创建

```go
err := myException.GetErrorCodeByCode("10001", "自定义错误信息")
err := myException.GetErrorCodeByHTTPStatus(403, "无权限")
```

---

## 错误码枚举 (CommonErrorCodeEnum)

常用枚举（来自 common_error_code.go）：

| 枚举 | Code | 说明 |
|------|------|------|
| OK | 200 | 成功 |
| BAD_REQUEST | 400 | 请求错误 |
| UNAUTHORIZED | 401 | 未授权 |
| FORBIDDEN | 403 | 禁止访问 |
| NOT_FOUND | 404 | 资源不存在 |
| METHOD_NOT_ALLOWED | 405 | 方法不允许 |
| INTERNAL_SERVER_ERROR | 500 | 服务器内部错误 |
| INVALID_PARAM | 10000 | 参数无效 |
| DB_EXCEPTION | 10001 | 数据库异常 |
| NULL_POINTER | 10002 | 空指针 |
| QUERY_PARAM_ERROR | 10006 | 查询参数不符合要求 |

完整枚举见 `myException/common_error_code.go`（1xx–5xx HTTP 风格，6xx 未知，10000+ 系统级）。

---

## 注册自定义错误码

当内置枚举不满足时，使用 `RegisterErrorCode` 注册：

```go
// 定义新枚举值（需在 const 块中追加）
const (
	// ...
	MY_CUSTOM_ERROR CommonErrorCodeEnum = 100
)

func init() {
	myException.RegisterErrorCode(MY_CUSTOM_ERROR, "20001", "业务自定义错误描述")
}

// 使用
err := myException.NewExceptionFromError(myException.MY_CUSTOM_ERROR)
```

注意：枚举值需在 `common_error_code.go` 的 const 中定义，或在业务包中扩展后调用 `RegisterErrorCode`。

---

## 与中间件配合

`middleware.ExceptionHandler` 会处理：

- `*MyException`：返回 `FailWithCode(e.Code, e.Message)`
- `*ValidationError`：返回 `BadRequest(e.Message)`
- `*NotFoundError`：返回 `NotFound`
- 其他 error：返回 `Fail(err.Error())`

Controller 中通过 `c.Error(err)` 或 `panic(err)` 传递异常，由中间件统一返回 JSON。推荐在 Controller 层统一使用 `myResult.ErrorWithError(c, err)` 来交给 MyResult + 中间件处理错误响应。

---

## 内层（Repository / Service）错误处理：统一使用 errors 包保留堆栈

为了在日志中看到完整调用链和堆栈信息，**建议在 Repository / Service 等内层统一使用 `github.com/pkg/errors` 对错误进行 Wrap**，在边界层（Controller 或 API Facade）再转换为 `MyException`：

```go
import (
	perrors "github.com/pkg/errors"
)

func (r *UserRepository) FindByID(ctx context.Context, id int64) (*model.UserDO, error) {
	user := &model.UserDO{}
	if err := r.GetById(ctx, user, id); err != nil {
		// 包装底层错误，附加语义 + 堆栈
		return nil, perrors.Wrap(err, "FindByID 查询 user 失败")
	}
	return user, nil
}

func (s *UserService) GetUser(ctx context.Context, id int64) (*model.UserDO, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		// 继续向上 Wrap，链路保留
		return nil, perrors.Wrap(err, "GetUser 业务处理失败")
	}
	return user, nil
}
```

在最外层（Controller / Handler）再将最终的 error 转成业务异常或直接交给中间件：

```go
func (u *UserController) Get(c *gin.Context) {
	// ...
	user, err := u.svc.GetUser(c.Request.Context(), id)
	if err != nil {
		// 如果是业务异常，可直接 panic / c.Error(MyException)
		// 否则可以统一转成 DB_EXCEPTION 等业务异常
		be := myException.DB_EXCEPTION.ToBusinessError()
		// 日志里可使用 %+v 打印完整堆栈（在自定义日志封装中实现）
		c.Error(be)
		return
	}
	myResult.Success(c, user)
}
```

**要点**：

- **内层统一用 `pkg/errors` 包装底层 error（`Wrap` / `Wrapf` / `WithStack`）保留堆栈；**
- **不在 Repository 层直接构造 `MyException`，而是在上层 Service/Controller 将“技术错误”映射为业务错误码；**
- 日志封装（`myLogger`）中如需打印堆栈，可在格式化时使用 `%+v` 或针对 `interface{ StackTrace() }` 做专门处理。
