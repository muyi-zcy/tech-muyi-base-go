package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/muyi-zcy/tech-muyi-base-go/myException"
	"github.com/muyi-zcy/tech-muyi-base-go/myLogger"
	"github.com/muyi-zcy/tech-muyi-base-go/myResult"
	"go.uber.org/zap"
)

// NotFoundHandler 处理404路由不存在的情况
func NotFoundHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		myLogger.WarnCtx(c, "路由不存在",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("ip", c.ClientIP()),
		)

		err := myException.NewBizError("platform.route.not_found", nil)
		c.JSON(200, buildErrorResult(c, err))
		c.Abort()
	}
}

// MethodNotAllowedHandler 处理405方法不允许的情况
func MethodNotAllowedHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		myLogger.WarnCtx(c, "方法不允许",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("ip", c.ClientIP()),
		)

		err := myException.NewBizError("platform.method.not_allowed", nil)
		c.JSON(200, buildErrorResult(c, err))
		c.Abort()
	}
}

func ExceptionHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				path := c.Request.URL.Path
				myLogger.ErrorCtx(c, "系统异常",
					zap.Any("error", err),
					zap.String("url", path),
				)

				var result myResult.MyResult
				switch e := err.(type) {
				case error:
					result = buildErrorResult(c, e)
				default:
					result = buildErrorResult(c, myException.NewBizError("platform.internal_error", nil))
				}

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

		c.Next()

		if len(c.Errors) > 0 {
			lastErr := c.Errors.Last().Err
			myLogger.ErrorCtx(c, "HTTP Errors",
				zap.Errors("errors", func() []error {
					errs := make([]error, 0, len(c.Errors))
					for _, e := range c.Errors {
						errs = append(errs, e.Err)
					}
					return errs
				}()),
			)

			c.JSON(200, buildErrorResult(c, lastErr))
			c.Abort()
		}
	}
}
