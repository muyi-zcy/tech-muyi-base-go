package config

import (
	"flag"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
)

var (
	// 全局配置实例
	GlobalConfig *Config
)

// Config 配置结构体
type Config struct {
	AppName  string         `mapstructure:"app_name"`
	Version  string         `mapstructure:"version"`
	Server   ServerConfig   `mapstructure:"server"`
	Log      LogConfig      `mapstructure:"log"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"maxsize"`
	MaxAge     int    `mapstructure:"maxage"`
	MaxBackups int    `mapstructure:"maxbackups"`
	Compress   bool   `mapstructure:"compress"`
	Stdout     bool   `mapstructure:"stdout"`
	LogSQL     bool   `mapstructure:"log_sql"` // 新增：是否记录SQL日志
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver          string `mapstructure:"driver"`
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	Database        string `mapstructure:"database"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime int    `mapstructure:"conn_max_idle_time"`

	// 连接与IO超时（秒）
	ConnTimeoutSec  int `mapstructure:"conn_timeout_sec"`
	ReadTimeoutSec  int `mapstructure:"read_timeout_sec"`
	WriteTimeoutSec int `mapstructure:"write_timeout_sec"`

	// GORM行为
	SkipDefaultTransaction bool `mapstructure:"skip_default_transaction"`
	PrepareStmt            bool `mapstructure:"prepare_stmt"`
	SlowThresholdMS        int  `mapstructure:"slow_threshold_ms"`

	// 其他DSN参数
	Timezone    string `mapstructure:"timezone"`     // e.g. Local, Asia/Shanghai
	ExtraParams string `mapstructure:"extra_params"` // 追加到 DSN 查询串
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`

	// 连接池与超时
	PoolSize        int `mapstructure:"pool_size"`
	MinIdleConns    int `mapstructure:"min_idle_conns"`
	DialTimeoutSec  int `mapstructure:"dial_timeout_sec"`
	ReadTimeoutSec  int `mapstructure:"read_timeout_sec"`
	WriteTimeoutSec int `mapstructure:"write_timeout_sec"`
	PoolTimeoutSec  int `mapstructure:"pool_timeout_sec"`
	IdleTimeoutSec  int `mapstructure:"idle_timeout_sec"`
	MaxRetries      int `mapstructure:"max_retries"`
	MinRetryBackoff int `mapstructure:"min_retry_backoff_ms"`
	MaxRetryBackoff int `mapstructure:"max_retry_backoff_ms"`
}

// Init 初始化配置
func Init() error {
	// 解析命令行参数
	configEnv := flag.String("env", "dev", "配置环境 (dev|local|pre|prod)")
	configFile := flag.String("config", "", "配置文件路径")
	flag.Parse()

	// 确定配置文件名
	var fileName string
	if *configFile != "" {
		fileName = *configFile
	} else {
		// 根据环境变量确定配置文件名
		env := strings.ToLower(*configEnv)
		switch env {
		case "dev":
			fileName = "app-dev.conf"
		case "local":
			fileName = "app-local.conf"
		case "pre":
			fileName = "app-pre.conf"
		case "prod":
			fileName = "app-prod.conf"
		default:
			fileName = "app-dev.conf"
		}
		fmt.Printf("使用配置文件: %s\n", fileName)
	}

	// 初始化 viper
	viper.SetConfigType("toml") // 配置文件类型为 TOML

	// 检查文件是否存在（优先使用当前目录下的文件）
	if _, err := os.Stat(fileName); err == nil {
		// 文件存在，直接读取
		viper.SetConfigFile(fileName)
	} else if _, err := os.Stat(filepath.Join("app", fileName)); err == nil {
		// 文件在app目录下存在
		viper.SetConfigFile(filepath.Join("app", fileName))
	} else {
		// 都不存在，使用默认配置
		fmt.Printf("配置文件 %s 不存在，使用默认配置\n", fileName)
		setDefaultConfig()
		GlobalConfig = &Config{}
		return viper.Unmarshal(GlobalConfig)
	}

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 监听配置文件变化
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("配置文件已更新:", e.Name)
		// 重新加载配置
		tempConfig := &Config{}
		if err := viper.Unmarshal(tempConfig); err != nil {
			fmt.Printf("重新加载配置失败: %v\n", err)
		} else {
			GlobalConfig = tempConfig
		}
	})

	// 解析配置到结构体
	GlobalConfig = &Config{}
	if err := viper.Unmarshal(GlobalConfig); err != nil {
		return fmt.Errorf("解析配置失败: %v", err)
	}

	fmt.Printf("配置加载成功: %+v\n", GlobalConfig)
	return nil
}

// GetConfigByType 根据配置类型获取配置并填充到指定结构体
// configType: 配置类型，如 "log", "server" 等
func GetConfigByType(configType string, target interface{}) error {
	// 检查配置是否已初始化
	if GlobalConfig == nil {
		return fmt.Errorf("配置未初始化")
	}

	// 根据配置类型获取子配置并反序列化到目标结构体
	switch strings.ToLower(configType) {
	case "log":
		return viper.UnmarshalKey("log", target)
	case "server":
		return viper.UnmarshalKey("server", target)
	case "database", "db":
		return viper.UnmarshalKey("database", target)
	case "redis":
		return viper.UnmarshalKey("redis", target)
	default:
		return viper.UnmarshalKey(configType, target)
	}
}

// setDefaultConfig 设置默认配置
func setDefaultConfig() {
	viper.SetDefault("app_name", "TechMuYiApp")
	viper.SetDefault("version", "1.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "dev")
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.filename", "app.log")
	viper.SetDefault("log.maxsize", 100)
	viper.SetDefault("log.maxage", 30)
	viper.SetDefault("log.maxbackups", 3)
	viper.SetDefault("log.compress", true)
	viper.SetDefault("log.stdout", false)  // 默认不启用控制台输出
	viper.SetDefault("log.log_sql", false) // 默认不记录SQL日志
	viper.SetDefault("log.enable_trace_id", false)
	viper.SetDefault("log.trace_id_key", "X-Trace-Id")
	viper.SetDefault("database.driver", "mysql")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 3306)
	viper.SetDefault("database.username", "root")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.database", "test")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 25)
	viper.SetDefault("database.conn_max_lifetime", 0)
	viper.SetDefault("database.conn_max_idle_time", 0)
	viper.SetDefault("database.conn_timeout_sec", 10)
	viper.SetDefault("database.read_timeout_sec", 3)
	viper.SetDefault("database.write_timeout_sec", 3)
	viper.SetDefault("database.skip_default_transaction", true)
	viper.SetDefault("database.prepare_stmt", false)
	viper.SetDefault("database.slow_threshold_ms", 200)
	viper.SetDefault("database.timezone", "Local")
	viper.SetDefault("database.extra_params", "")
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.pool_size", 10)
	viper.SetDefault("redis.min_idle_conns", 2)
	viper.SetDefault("redis.dial_timeout_sec", 5)
	viper.SetDefault("redis.read_timeout_sec", 3)
	viper.SetDefault("redis.write_timeout_sec", 3)
	viper.SetDefault("redis.pool_timeout_sec", 4)
	viper.SetDefault("redis.idle_timeout_sec", 300)
	viper.SetDefault("redis.max_retries", 2)
	viper.SetDefault("redis.min_retry_backoff_ms", 8)
	viper.SetDefault("redis.max_retry_backoff_ms", 512)
}

// GetConfig 获取全局配置
func GetConfig() *Config {
	return GlobalConfig
}

// GetAppName 获取应用名称
func GetAppName() string {
	if GlobalConfig != nil && GlobalConfig.AppName != "" {
		return GlobalConfig.AppName
	}
	return "TechMuYiApp" // 默认应用名称
}

// GetVersion 获取应用版本
func GetVersion() string {
	if GlobalConfig != nil && GlobalConfig.Version != "" {
		return GlobalConfig.Version
	}
	return "1.0.0" // 默认版本
}

// GetServerConfig 获取服务器配置
func GetServerConfig() ServerConfig {
	if GlobalConfig != nil {
		return GlobalConfig.Server
	}
	return ServerConfig{}
}

// GetLogConfig 获取日志配置
func GetLogConfig() LogConfig {
	if GlobalConfig != nil {
		return GlobalConfig.Log
	}
	return LogConfig{}
}

// GetDatabaseConfig 获取数据库配置
func GetDatabaseConfig() DatabaseConfig {
	if GlobalConfig != nil {
		return GlobalConfig.Database
	}
	return DatabaseConfig{}
}

// GetRedisConfig 获取Redis配置
func GetRedisConfig() RedisConfig {
	if GlobalConfig != nil {
		return GlobalConfig.Redis
	}
	return RedisConfig{}
}
