package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/muyi-zcy/tech-muyi-base-go/myContext"
	"github.com/muyi-zcy/tech-muyi-base-go/myException"
	"github.com/muyi-zcy/tech-muyi-base-go/myLocale"
	"github.com/muyi-zcy/tech-muyi-base-go/myResult"
)

func buildErrorResult(c *gin.Context, err error) myResult.MyResult {
	locale := myContext.GetLocaleFromGinCtx(c)
	code := myException.GetErrorCode(err)
	args := myException.GetErrorArgs(err)

	switch e := err.(type) {
	case *myException.BizError:
		message := resolveMessage(code, locale, args)
		return myResult.FailWithCode(code, message)
	case *myException.MyException:
		message := e.Message
		if message == "" {
			message = resolveMessage(code, locale, args)
		}
		return myResult.FailWithCode(code, message)
	case *myException.ValidationError:
		if args == nil {
			args = map[string]string{"field": e.Field, "message": e.Message}
		}
		message := resolveMessage(code, locale, args)
		if message == code && e.Message != "" {
			message = e.Message
		}
		return myResult.FailWithCode(code, message)
	case *myException.NotFoundError:
		message := resolveMessage(code, locale, args)
		if message == code {
			message = e.Error()
		}
		return myResult.FailWithCode(code, message)
	default:
		message := resolveMessage("platform.internal_error", locale, nil)
		if !myLocale.Initialized() {
			return myResult.Fail(err.Error())
		}
		return myResult.FailWithCode("platform.internal_error", message)
	}
}

func resolveMessage(code, locale string, args map[string]string) string {
	if myLocale.Initialized() {
		message := myLocale.Resolve(code, locale, args)
		if message != "" && message != code {
			return message
		}
	}
	if args != nil {
		if msg, ok := args["message"]; ok && msg != "" {
			return msg
		}
	}
	return code
}
