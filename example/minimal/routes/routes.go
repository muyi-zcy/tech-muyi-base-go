package routes

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/muyi-zcy/tech-muyi-base-go/config"
	"github.com/muyi-zcy/tech-muyi-base-go/infrastructure"
	"github.com/muyi-zcy/tech-muyi-base-go/myResult"
)

func Register(engine *gin.Engine) {
	engine.GET("/", func(c *gin.Context) {
		myResult.Success(c, gin.H{
			"service": "example-minimal",
			"desc":    "测试 base 包：HTTP、MySQL、Redis、日志、统一返回",
		})
	})

	v1 := engine.Group("/api/v1")
	test := v1.Group("/test")
	{
		test.GET("/ping", ping)
		test.GET("/db", dbCheck)
		test.GET("/redis", redisCheck)
		test.POST("/echo", echo)
	}

	v1.GET("/system/info", systemInfo)
}

func ping(c *gin.Context) {
	myResult.Success(c, gin.H{"pong": true})
}

func dbCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	if err := infrastructure.HealthCheck(ctx); err != nil {
		myResult.Error(c, "数据库连接失败: "+err.Error())
		return
	}

	var version string
	if err := infrastructure.GetDB().WithContext(ctx).Raw("SELECT VERSION()").Scan(&version).Error; err != nil {
		myResult.Error(c, "数据库查询失败: "+err.Error())
		return
	}

	myResult.Success(c, gin.H{"mysqlVersion": version})
}

func redisCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	key := "example:minimal:ping"
	client := infrastructure.GetRedis()
	if err := client.Set(ctx, key, "ok", time.Minute).Err(); err != nil {
		myResult.Error(c, "Redis 写入失败: "+err.Error())
		return
	}

	val, err := client.Get(ctx, key).Result()
	if err != nil {
		myResult.Error(c, "Redis 读取失败: "+err.Error())
		return
	}

	myResult.Success(c, gin.H{"key": key, "value": val})
}

func echo(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		myResult.BadRequestResponse(c, "请求体必须是 JSON")
		return
	}
	myResult.Success(c, body)
}

func systemInfo(c *gin.Context) {
	cfg := config.GetConfig()
	myResult.Success(c, gin.H{
		"appName":  cfg.AppName,
		"version":  cfg.Version,
		"server":   cfg.Server,
		"database": gin.H{"host": cfg.Database.Host, "port": cfg.Database.Port, "database": cfg.Database.Database},
		"redis":    gin.H{"host": cfg.Redis.Host, "port": cfg.Redis.Port},
	})
}
