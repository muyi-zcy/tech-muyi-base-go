package myLocale

import (
	"github.com/gin-gonic/gin"
	"github.com/muyi-zcy/tech-muyi-base-go/myResult"
)

// RegisterRoutes 注册错误文案与契约的公开 HTTP 接口（挂载在 /api/{appCode} 路由组下）
func RegisterRoutes(apiGroup *gin.RouterGroup) {
	if defaultStore == nil {
		return
	}

	open := apiGroup.Group("/v1/open")
	{
		open.GET("/error-messages", getErrorMessages)
		open.GET("/error-contracts", getErrorContracts)
		open.GET("/locales", getLocales)
	}
}

func getErrorMessages(c *gin.Context) {
	locale := c.Query("locale")
	bundle, err := GetMessages(locale)
	if err != nil {
		myResult.Error(c, err.Error())
		return
	}
	myResult.Success(c, bundle)
}

func getErrorContracts(c *gin.Context) {
	contract, err := GetContract()
	if err != nil {
		myResult.Error(c, err.Error())
		return
	}
	myResult.Success(c, contract)
}

func getLocales(c *gin.Context) {
	myResult.Success(c, gin.H{
		"appCode": AppCode(),
		"locales": SupportedLocales(),
		"default": defaultStore.defaultLocale,
	})
}
