package core

import (
	"fmt"
	"go.uber.org/zap"
	"tech-muyi-base-go/config"
	"tech-muyi-base-go/infrastructure"
	"tech-muyi-base-go/logger"
)

// App 应用实例
type App struct {
	Name    string
	Version string
	Config  *config.Config
}

// NewApp 创建应用实例
func NewApp(name, version string) *App {
	app := &App{
		Name:    name,
		Version: version,
	}

	// 自动初始化应用
	app.initialize()

	return app
}

// NewAppFromConfig 从配置创建并初始化应用实例
func NewAppFromConfig() *App {
	// 先初始化配置
	if err := config.Init(); err != nil {
		// 如果配置初始化失败，使用默认值
		fmt.Printf("配置初始化失败: %v，使用默认配置\n", err)

		app := &App{
			Name:    config.GetAppName(), // 这会返回默认值
			Version: config.GetVersion(), // 这会返回默认值
		}

		// 自动初始化应用
		app.initialize()

		return app
	}

	// 从配置获取应用名称和版本
	appName := config.GetAppName()
	version := config.GetVersion()

	app := &App{
		Name:    appName,
		Version: version,
		Config:  config.GetConfig(),
	}

	// 自动初始化应用
	app.initialize()

	return app
}

// initialize 初始化应用（内部方法）
func (a *App) initialize() error {
	// 如果Config还没有初始化，则初始化配置
	if a.Config == nil {
		if err := config.Init(); err != nil {
			// 记录错误但不返回，因为这可能在NewApp中调用
			fmt.Printf("初始化配置失败: %v\n", err)
			return err
		}
		a.Config = config.GetConfig()
	}

	// 记录应用启动日志
	logger.Info("应用初始化完成",
		zap.String("appName", a.Name),
		zap.String("appVersion", a.Version))

	return nil
}

// Initialize 初始化应用（为了向后兼容保留此方法）
func (a *App) Initialize() error {
	return a.initialize()
}

// Shutdown 关闭应用，清理资源
func (a *App) Shutdown() error {
	// 同步日志
	logger.Sync()

	// 关闭数据库连接
	if err := infrastructure.CloseDB(); err != nil {
		return fmt.Errorf("关闭数据库连接失败: %v", err)
	}

	// 关闭Redis连接
	if err := infrastructure.CloseRedis(); err != nil {
		return fmt.Errorf("关闭Redis连接失败: %v", err)
	}

	return nil
}

// GetConfig 获取配置
func (a *App) GetConfig() *config.Config {
	return a.Config
}
