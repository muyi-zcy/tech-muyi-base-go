---
name: tech-muyi-base-go-exception
description: tech-muyi-base-go myException 错误处理：BizError、MyException、ValidationError、NotFoundError、ExceptionHandler 中间件配合、myLocale 国际化、Repository/Service 层 pkg/errors Wrap。使用场景：业务异常、错误码定义、参数校验失败、404、错误分层映射。
---

# tech-muyi-base-go 异常处理

完整文档：[docs/USAGE.md §14](../../../docs/USAGE.md#14-异常处理)。

基于 `github.com/muyi-zcy/tech-muyi-base-go/myException` 进行统一错误处理。

## 使用场景

- 用户需要抛出业务异常、校验异常、404 等
- 用户询问错误码如何定义、如何配置国际化文案
- 用户需要将 error 转为统一返回结构

## 异常类型

| 类型 | 用途 | 创建方式 |
|------|------|----------|
| `BizError` | 业务异常（推荐） | `NewBizError(code, args)`，message 由中间件按 locale 解析 |
| `MyException` | 兼容旧写法 | `NewException(code, msg)` |
| `ValidationError` | 参数校验失败 | `NewValidationError(field, message)` |
| `NotFoundError` | 资源不存在 | `NewNotFoundError(resource, id)` |

## 核心 API

### 创建异常

```go
import "github.com/muyi-zcy/tech-muyi-base-go/myException"

// 推荐：只抛 code + args
err := myException.NewBizError("user.username.exists", map[string]string{
    "username": "admin",
})

// 参数校验
err := myException.NewBizError("user.validation.required", map[string]string{
    "field": "username",
})

// 资源不存在
err := myException.NewBizError("user.user.not_found", nil)

// 校验异常（映射为 platform.validation.required）
err := myException.NewValidationError("username", "用户名不能为空")

// 资源不存在（映射为 platform.resource.not_found）
err := myException.NewNotFoundError("User", 123)
```

### 从 error 提取 code / message / args

```go
code := myException.GetErrorCode(err)
msg := myException.GetErrorMessage(err)
args := myException.GetErrorArgs(err)
```

---

## 错误码格式

格式为 `{appCode}.{module}.{semantic}`，第一段必须是 appCode：

```
user.username.exists
platform.validation.required
platform.route.not_found
```

平台公共错误码定义在 `myLocale/platform/contracts/errors.yaml`，业务错误码在各服务 `contracts/errors.yaml` 中定义。

---

## 与中间件配合

`middleware.ExceptionHandler` 会处理：

- `*BizError`：按 locale 解析 message，返回 `FailWithCode(code, message)`
- `*MyException`：返回 `FailWithCode(e.Code, e.Message)`
- `*ValidationError`：映射为 `platform.validation.required`
- `*NotFoundError`：映射为 `platform.resource.not_found`
- 其他 error：返回 `platform.internal_error`

Controller 中通过 `c.Error(err)` 或 `panic(err)` 传递异常，由中间件统一返回 JSON。推荐在 Controller 层统一使用 `myResult.ErrorWithError(c, err)`。

---

## 内层（Repository / Service）错误处理：统一使用 errors 包保留堆栈

为了在日志中看到完整调用链和堆栈信息，**建议在 Repository / Service 等内层统一使用 `github.com/pkg/errors` 对错误进行 Wrap**，在边界层（Controller 或 API Facade）再转换为 `BizError`：

```go
import (
	perrors "github.com/pkg/errors"
)

func (r *UserRepository) FindByID(ctx context.Context, id int64) (*model.UserDO, error) {
	user := &model.UserDO{}
	if err := r.GetById(ctx, user, id); err != nil {
		return nil, perrors.Wrap(err, "FindByID 查询 user 失败")
	}
	return user, nil
}

func (s *UserService) GetUser(ctx context.Context, id int64) (*model.UserDO, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, myException.NewBizError("user.user.not_found", nil)
		}
		return nil, perrors.Wrap(err, "GetUser 业务处理失败")
	}
	return user, nil
}
```

在最外层（Controller / Handler）将 error 交给中间件：

```go
func (u *UserController) Get(c *gin.Context) {
	user, err := u.svc.GetUser(c.Request.Context(), id)
	if err != nil {
		myResult.ErrorWithError(c, err)
		return
	}
	myResult.Success(c, user)
}
```

**要点**：

- **内层统一用 `pkg/errors` 包装底层 error（`Wrap` / `Wrapf` / `WithStack`）保留堆栈；**
- **不在 Repository 层直接构造 `BizError`，而是在 Service 层将业务错误映射为错误码；**
- 日志封装（`myLogger`）中如需打印堆栈，可在格式化时使用 `%+v`。
