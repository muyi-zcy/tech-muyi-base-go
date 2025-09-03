package main

import (
	"github.com/gin-gonic/gin"
	"github.com/muyi-zcy/tech-muyi-base-go/config"
	"github.com/muyi-zcy/tech-muyi-base-go/core"
	"github.com/muyi-zcy/tech-muyi-base-go/myResult"
	"time"
)

func main() {

	// 从配置创建并初始化启动器（自动创建并初始化应用实例）
	starter, err := core.Initialize()
	if err != nil {
		panic(err)
	}

	// 获取Gin引擎
	engine := starter.GetEngine()

	// 注册业务路由
	registerRoutes(engine)

	// 启动应用（自动注册日志中间件和健康检查）
	if err := starter.Run(); err != nil {
		panic(err)
	}
}

// registerRoutes 注册路由
func registerRoutes(engine *gin.Engine) {
	// 基础路由
	engine.GET("/", func(c *gin.Context) {
		time.Sleep(1)
		myResult.Success(c, "欢迎使用Tech MuYi Go基础包新服务")
	})

	// API版本分组
	api := engine.Group("/api/v1")
	{
		// 系统相关接口
		system := api.Group("/system")
		{
			system.GET("/health", healthCheck)
			system.GET("/config", getConfig)
			system.GET("/info", getSystemInfo)
		}

		// 测试接口
		test := api.Group("/test")
		{
			test.GET("/ping", ping)
			test.POST("/echo", echo)
			test.GET("/error", testError)
		}
	}
}

// 系统相关接口
func healthCheck(c *gin.Context) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": "2024-01-01T00:00:00Z",
		"version":   "1.0.0",
		"uptime":    "1h 30m",
	}

	myResult.Success(c, health)
}

func getConfig(c *gin.Context) {
	cfg := config.GetConfig()
	if cfg == nil {
		myResult.Fail("配置未初始化")
		return
	}

	// 只返回安全的配置信息
	safeConfig := map[string]interface{}{
		"app_name": cfg.AppName,
		"version":  cfg.Version,
		"server": map[string]interface{}{
			"port": cfg.Server.Port,
			"mode": cfg.Server.Mode,
		},
		"log": map[string]interface{}{
			"level":    cfg.Log.Level,
			"filename": cfg.Log.Filename,
		},
	}

	myResult.Success(c, safeConfig)
}

func getSystemInfo(c *gin.Context) {
	info := map[string]interface{}{
		"app_name":   "Tech MuYi Go Base",
		"version":    "1.0.0",
		"go_version": "1.22.2",
		"framework":  "Gin",
		"database":   "MySQL",
		"cache":      "Redis",
		"logger":     "Zap",
		"config":     "Viper",
	}

	myResult.Success(c, info)
}

// 测试接口
func ping(c *gin.Context) {
	myResult.Success(c, map[string]string{"message": "pong", "timestamp": "2024-01-01T00:00:00Z"})
}

func echo(c *gin.Context) {
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		myResult.Fail("请求参数错误: " + err.Error())
		return
	}

	// 回显请求数据
	myResult.Success(c, data)
}

func testError(c *gin.Context) {
	// 测试错误处理
	myResult.Fail("这是一个测试错误")
}
