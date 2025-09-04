package infrastructure

import (
	"context"
	"fmt"
	"github.com/muyi-zcy/tech-muyi-base-go/config"
	"github.com/muyi-zcy/tech-muyi-base-go/myLogger"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"time"
)

var (
	// RedisClient Redis客户端实例
	RedisClient *redis.Client

	// ctx Redis操作上下文
	ctx = context.Background()
)

// InitRedis 初始化Redis连接
func InitRedis() error {
	redisConfig := config.GetRedisConfig()

	// 检查配置是否正确加载
	if redisConfig.Host == "" {
		return fmt.Errorf("Redis配置未正确加载")
	}

	myLogger.Info("Redis配置",
		zap.String("host", redisConfig.Host),
		zap.Int("port", redisConfig.Port),
		zap.String("password", func() string {
			if redisConfig.Password != "" {
				return "***"
			}
			return ""
		}()),
		zap.Int("db", redisConfig.DB),
		zap.Int("poolSize", redisConfig.PoolSize),
		zap.Int("minIdleConns", redisConfig.MinIdleConns))

	RedisClient = redis.NewClient(&redis.Options{
		Addr:            fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port),
		Password:        redisConfig.Password,
		DB:              redisConfig.DB,
		PoolSize:        redisConfig.PoolSize,
		MinIdleConns:    redisConfig.MinIdleConns,
		DialTimeout:     time.Duration(redisConfig.DialTimeoutSec) * time.Second,
		ReadTimeout:     time.Duration(redisConfig.ReadTimeoutSec) * time.Second,
		WriteTimeout:    time.Duration(redisConfig.WriteTimeoutSec) * time.Second,
		PoolTimeout:     time.Duration(redisConfig.PoolTimeoutSec) * time.Second,
		IdleTimeout:     time.Duration(redisConfig.IdleTimeoutSec) * time.Second,
		MaxRetries:      redisConfig.MaxRetries,
		MinRetryBackoff: time.Duration(redisConfig.MinRetryBackoff) * time.Millisecond,
		MaxRetryBackoff: time.Duration(redisConfig.MaxRetryBackoff) * time.Millisecond,
	})

	// 测试连接
	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("Redis连接测试失败: %v", err)
	}

	return nil
}

// GetRedis 获取Redis客户端实例
func GetRedis() *redis.Client {
	return RedisClient
}

// GetContext 获取Redis操作上下文
func GetContext() context.Context {
	return ctx
}

// CloseRedis 关闭Redis连接
func CloseRedis() error {
	if RedisClient != nil {
		return RedisClient.Close()
	}
	return nil
}
