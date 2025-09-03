package main

import (
	"github.com/gin-gonic/gin"
	"github.com/muyi-zcy/tech-muyi-base-go/config"
	"github.com/muyi-zcy/tech-muyi-base-go/core"
	"github.com/muyi-zcy/tech-muyi-base-go/logger"
	"github.com/muyi-zcy/tech-muyi-base-go/myResult"
	"go.uber.org/zap"
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
		// 用户相关接口
		users := api.Group("/users")
		{
			users.GET("", getUsers)
			users.GET("/:id", getUserByID)
			users.POST("", createUser)
			users.PUT("/:id", updateUser)
			users.DELETE("/:id", deleteUser)
		}

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

// 用户相关接口
func getUsers(c *gin.Context) {
	// 模拟用户数据
	users := []map[string]interface{}{
		{"id": 1, "name": "张三", "email": "zhangsan@example.com", "age": 25},
		{"id": 2, "name": "李四", "email": "lisi@example.com", "age": 30},
		{"id": 3, "name": "王五", "email": "wangwu@example.com", "age": 28},
	}

	query := &myResult.MyQuery{
		Size:    10,
		Current: 1,
		Total:   3,
	}

	logger.InfoCtx(c, "获取用户列表", zap.Any("query", query))

	myResult.SuccessWithQuery(c, users, query)
}

func getUserByID(c *gin.Context) {
	id := c.Param("id")

	// 模拟根据ID查询用户
	user := map[string]interface{}{
		"id":    id,
		"name":  "用户" + id,
		"email": "user" + id + "@example.com",
		"age":   25,
	}

	myResult.Success(c, user)
}

func createUser(c *gin.Context) {
	var user map[string]interface{}
	if err := c.ShouldBindJSON(&user); err != nil {
		myResult.Fail("请求参数错误: " + err.Error())
		return
	}

	// 模拟创建用户
	user["id"] = 999
	user["created_at"] = "2024-01-01T00:00:00Z"

	logger.Info("创建用户", zap.Any("user", user))
	myResult.Success(c, user)
}

func updateUser(c *gin.Context) {
	id := c.Param("id")
	var user map[string]interface{}
	if err := c.ShouldBindJSON(&user); err != nil {
		myResult.Fail("请求参数错误: " + err.Error())
		return
	}

	// 模拟更新用户
	user["id"] = id
	user["updated_at"] = "2024-01-01T00:00:00Z"

	logger.Info("更新用户", zap.String("id", id), zap.Any("user", user))
	myResult.Success(c, user)
}

func deleteUser(c *gin.Context) {
	id := c.Param("id")

	// 模拟删除用户
	logger.Info("删除用户", zap.String("id", id))
	myResult.Success(c, map[string]string{"message": "用户删除成功", "id": id})
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
