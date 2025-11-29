package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/muyi-zcy/tech-muyi-base-go/myException"
	"github.com/muyi-zcy/tech-muyi-base-go/myLogger"
	"github.com/muyi-zcy/tech-muyi-base-go/myResult"
	"go.uber.org/zap"
)

func ExceptionHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 获取请求信息
				path := c.Request.URL.Path

				// 记录错误日志（自动获取traceId）
				myLogger.ErrorCtx(c, "系统异常",
					zap.Any("error", err),
					zap.String("url", path),
				)

				// 根据错误类型返回不同的响应
				var result myResult.MyResult
				switch e := err.(type) {
				case *myException.MyException:
					result = myResult.FailWithCode(e.Code, e.Message)
				case *myException.ValidationError:
					result = myResult.BadRequest(e.Message)
				case *myException.NotFoundError:
					result = myResult.NotFound(myException.NOT_FOUND.GetResultMsg())
				default:
					result = myResult.Fail("系统内部错误")
				}

				// 记录错误响应（自动获取traceId）
				myLogger.ErrorCtx(c, "ErrorResponse",
					zap.String("url", path),
					zap.String("code", result.Code),
					zap.String("message", result.Message),
				)

				c.JSON(200, result)
				c.Abort()
				return
			}
		}()

		// 处理请求
		c.Next()

		// 如果有业务错误，记录响应
		if len(c.Errors) > 0 {
			lastErr := c.Errors.Last().Err
			// 记录所有错误
			myLogger.ErrorCtx(c, "HTTP Errors",
				zap.Errors("errors", func() []error {
					errs := make([]error, 0, len(c.Errors))
					for _, e := range c.Errors {
						errs = append(errs, e.Err)
					}
					return errs
				}()),
			)

			// 根据错误类型返回响应
			var result myResult.MyResult
			switch e := lastErr.(type) {
			case *myException.MyException:
				result = myResult.FailWithCode(e.Code, e.Message)
			case *myException.ValidationError:
				result = myResult.BadRequest(e.Message)
			case *myException.NotFoundError:
				result = myResult.NotFound(e.Error())
			default:
				result = myResult.Fail(lastErr.Error())
			}

			c.JSON(200, result)
			c.Abort()
		}
	}
}
