package myAuth

import "github.com/gin-gonic/gin"

// BeforeAuthHook 鉴权前流程 Hook，不写 context。
// skipAuth=true 时跳过 token 校验直接放行。
type BeforeAuthHook func(c *gin.Context) (skipAuth bool, err error)
