package myLogger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"path/filepath"
	"time"
)

var (
	log *zap.Logger
)

// GlobalConfig 全局日志配置
type GlobalConfig struct {
	EnableTraceId bool   // 是否启用traceId
	TraceIdKey    string // traceId字段名
}

// TimeFormat 时间格式枚举
type TimeFormat string

const (
	TimeFormatISO8601    TimeFormat = "iso8601"    // ISO8601格式: 2006-01-02T15:04:05.000Z07:00
	TimeFormatUnix       TimeFormat = "unix"       // Unix时间戳: 1640995200
	TimeFormatUnixMillis TimeFormat = "unixMillis" // Unix毫秒时间戳: 1640995200000
	TimeFormatRFC3339    TimeFormat = "rfc3339"    // RFC3339格式: 2006-01-02T15:04:05Z07:00
	TimeFormatCustom     TimeFormat = "custom"     // 自定义格式: 2006-01-02 15:04:05
)

// customTimeEncoder 自定义时间编码器，支持多种时间格式
func customTimeEncoder(timeFormat TimeFormat) func(time.Time, zapcore.PrimitiveArrayEncoder) {
	return func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		switch timeFormat {
		case TimeFormatUnix:
			// Unix时间戳（秒）
			enc.AppendInt64(t.Unix())
		case TimeFormatUnixMillis:
			// Unix毫秒时间戳
			enc.AppendInt64(t.UnixMilli())
		case TimeFormatRFC3339:
			// RFC3339格式
			enc.AppendString(t.Format(time.RFC3339))
		case TimeFormatCustom:
			// 自定义格式
			enc.AppendString(t.Format("2006-01-02 15:04:05"))
		case TimeFormatISO8601:
		default:
			// 默认使用Unix毫秒时间戳
			enc.AppendInt64(t.UnixMilli())
		}
	}
}

// Init 初始化日志（使用默认配置）
func Init() error {
	// 获取环境变量
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}

	// 根据环境变量设置日志级别
	var level zapcore.Level
	switch env {
	case "prod", "production":
		level = zapcore.InfoLevel
	default:
		level = zapcore.DebugLevel
	}

	// 创建日志配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "myLogger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     customTimeEncoder(TimeFormatUnixMillis), // 默认使用Unix毫秒时间戳
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 创建encoder
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	// 创建writeSyncer
	writeSyncer := zapcore.AddSync(os.Stdout)

	// 创建core
	core := zapcore.NewCore(encoder, writeSyncer, level)

	// 创建logger，跳过1层调用栈以便正确显示caller
	log = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.Development())

	return nil
}

// InitWithConfig 根据配置初始化日志
func InitWithConfig(logConfig LogConfig) error {
	// 确保日志目录存在
	logDir := filepath.Dir(logConfig.Filename)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	// 解析日志级别
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(logConfig.Level)); err != nil {
		level = zapcore.InfoLevel
	}

	// 创建日志轮转配置
	lumberJackLogger := &lumberjack.Logger{
		Filename:   logConfig.Filename,
		MaxSize:    logConfig.MaxSize,
		MaxAge:     logConfig.MaxAge,
		MaxBackups: logConfig.MaxBackups,
		LocalTime:  true,
		Compress:   logConfig.Compress,
	}

	// 确定时间格式，如果没有配置则使用Unix毫秒时间戳
	timeFormat := TimeFormat(logConfig.TimeFormat)
	if timeFormat == "" {
		timeFormat = TimeFormatUnixMillis // 默认使用Unix毫秒时间戳
	}

	// 创建encoder配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "myLogger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     customTimeEncoder(timeFormat), // 使用配置的时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 创建encoder
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	// 创建writeSyncer列表
	var writeSyncers []zapcore.WriteSyncer

	// 如果启用stdout，则添加控制台输出
	if logConfig.Stdout {
		writeSyncers = append(writeSyncers, zapcore.AddSync(os.Stdout))
	}

	// 添加文件输出
	writeSyncers = append(writeSyncers, zapcore.AddSync(lumberJackLogger))

	// 创建core
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(writeSyncers...), level),
	)

	// 创建logger，跳过1层调用栈以便正确显示caller
	log = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.Development())

	return nil
}

// LogConfig 日志配置结构体（保持向后兼容）
// 注意：这个结构体需要与config包中的LogConfig结构体保持一致
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"maxsize"`
	MaxAge     int    `mapstructure:"maxage"`
	MaxBackups int    `mapstructure:"maxbackups"`
	Compress   bool   `mapstructure:"compress"`
	Stdout     bool   `mapstructure:"stdout"`
	LogSQL     bool   `mapstructure:"log_sql"`     // 新增：是否记录SQL日志
	TimeFormat string `mapstructure:"time_format"` // 可选：时间格式配置
}

// Sync 同步日志
func Sync() {
	if log != nil {
		log.Sync()
	}
}
