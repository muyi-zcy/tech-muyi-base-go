package core

import (
	"github.com/muyi-zcy/tech-muyi-base-go/config"
	"github.com/muyi-zcy/tech-muyi-base-go/infrastructure"
	"github.com/muyi-zcy/tech-muyi-base-go/infrastructure/nacos"
	"github.com/muyi-zcy/tech-muyi-base-go/infrastructure/rpc"
	"github.com/muyi-zcy/tech-muyi-base-go/middleware"
	"github.com/muyi-zcy/tech-muyi-base-go/myContext"
	"github.com/muyi-zcy/tech-muyi-base-go/myLogger"
	"github.com/pkg/errors"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Starter 应用启动器
type Starter struct {
	App    *App
	Engine *gin.Engine
}

// NewStarter 创建应用启动器
func NewStarter(app *App) *Starter {
	return &Starter{
		App: app,
	}
}

// NewStarterFromConfig 从配置创建应用启动器
func NewStarterFromConfig() *Starter {
	// 从配置创建并初始化应用实例
	app := NewAppFromConfig()

	// 创建启动器
	return &Starter{
		App: app,
	}
}

// NewStarterFromConfigAndInitialize 从配置创建并初始化应用启动器
func Initialize() (*Starter, error) {
	// 从配置创建启动器
	starter := NewStarterFromConfig()

	// 初始化启动器
	if err := starter.Initialize(); err != nil {
		return nil, errors.Wrap(err, "初始化启动器失败")
	}

	return starter, nil
}

// Initialize 初始化启动器
func (s *Starter) Initialize() error {
	// 创建Gin引擎
	s.Engine = gin.New()

	// 初始化日志系统
	if err := s.InitializeLogger(); err != nil {
		return errors.Wrap(err, "初始化日志系统失败")
	}

	// 注册默认中间件
	s.RegisterDefaultMiddlewares()

	// 根据配置自动注册基础设施
	if err := s.RegisterInfrastructure(); err != nil {
		return errors.Wrap(err, "注册基础设施失败")
	}

	return nil
}

// InitializeLogger 初始化日志系统
func (s *Starter) InitializeLogger() error {
	// 检查是否有配置
	if s.App.Config == nil {
		myLogger.Info("未找到配置，使用默认日志配置")
		return myLogger.Init()
	}

	// 根据配置初始化日志
	logConfig := myLogger.LogConfig{
		Level:      s.App.Config.Log.Level,
		Filename:   s.App.Config.Log.Filename,
		MaxSize:    s.App.Config.Log.MaxSize,
		MaxAge:     s.App.Config.Log.MaxAge,
		MaxBackups: s.App.Config.Log.MaxBackups,
		Compress:   s.App.Config.Log.Compress,
		Stdout:     s.App.Config.Log.Stdout,
	}

	if err := myLogger.InitWithConfig(logConfig); err != nil {
		myLogger.Warn("使用配置初始化日志失败，使用默认初始化", zap.Error(err))
		return myLogger.Init()
	}

	myLogger.Info("日志系统初始化成功")
	return nil
}

// RegisterDefaultMiddlewares 注册默认中间件
func (s *Starter) RegisterDefaultMiddlewares() {
	// 上下文管理中间件（必须在最前面，为其他中间件提供traceId）
	s.Engine.Use(myContext.ContextMiddleware())

	defaultLocale := "zh-CN"
	if s.App.Config != nil && s.App.Config.Locale.DefaultLocale != "" {
		defaultLocale = s.App.Config.Locale.DefaultLocale
	}
	s.Engine.Use(myContext.LocaleMiddleware(defaultLocale))

	// 异常处理中间件
	s.Engine.Use(middleware.ExceptionHandler())

	//Gin恢复中间件
	//s.Engine.Use(gin.Recovery())

	// 日志中间件（必须在上下文中间件之后）
	s.Engine.Use(middleware.Logger())
}

// RegisterInfrastructure 根据配置自动注册基础设施
func (s *Starter) RegisterInfrastructure() error {
	// 检查是否有配置
	if s.App.Config == nil {
		myLogger.Info("未找到配置，跳过基础设施注册")
		return nil
	}

	// 检查是否需要注册数据库
	if s.needDatabase() {
		myLogger.Info("初始化数据库连接")
		if err := infrastructure.InitDatabase(); err != nil {
			myLogger.Error("数据库连接初始化失败", zap.Error(err))
			return errors.Wrap(err, "数据库连接初始化失败")
		}
		myLogger.Info("数据库连接初始化成功")
	}

	// 检查是否需要注册Redis
	if s.needRedis() {
		myLogger.Info("初始化Redis连接")
		if err := infrastructure.InitRedis(); err != nil {
			myLogger.Error("Redis连接初始化失败", zap.Error(err))
			return errors.Wrap(err, "Redis连接初始化失败")
		}
		myLogger.Info("Redis连接初始化成功")
	}

	// 可插拔：Nacos + RPC
	if err := s.registerPlugins(); err != nil {
		myLogger.Warn("插件初始化部分失败，服务继续启动", zap.Error(err))
	}

	return nil
}

func (s *Starter) registerPlugins() error {
	if s.App.Config == nil {
		return nil
	}
	if _, err := nacos.InitWithError(s.App.Config); err != nil {
		myLogger.Warn("Nacos 插件初始化失败，已降级为 noop", zap.Error(err))
	}
	rpc.Init(s.App.Config, nacos.GetRegistry())
	return nil
}

// needDatabase 检查是否需要数据库
func (s *Starter) needDatabase() bool {
	// 检查数据库配置是否存在且有效
	if s.App.Config == nil {
		return false
	}
	dbConfig := config.GetDatabaseConfig()
	return dbConfig.Host != "" && dbConfig.Port > 0
}

// needRedis 检查是否需要Redis
func (s *Starter) needRedis() bool {
	// 检查Redis配置是否存在且有效
	if s.App.Config == nil {
		return false
	}
	redisConfig := config.GetRedisConfig()
	return redisConfig.Host != "" && redisConfig.Port > 0
}

// GetEngine 获取Gin引擎
func (s *Starter) GetEngine() *gin.Engine {
	return s.Engine
}

// GetAPIGroup 获取带 /api/{appCode} 前缀的路由组，业务路由应注册在此组下
func (s *Starter) GetAPIGroup() *gin.RouterGroup {
	return s.Engine.Group(config.APIPrefix())
}
