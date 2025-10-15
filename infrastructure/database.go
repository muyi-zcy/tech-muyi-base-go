package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/muyi-zcy/tech-muyi-base-go/config"
	"github.com/muyi-zcy/tech-muyi-base-go/myLogger"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// DB GORM数据库连接实例
var DB *gorm.DB

// sqlDB 原根数据库连接实例（用于连接池管理）
var sqlDB *sql.DB

// InitDatabase 初始化GORM数据库连接
func InitDatabase() error {
	dbConfig := config.GetDatabaseConfig()
	logConfig := config.GetLogConfig()

	// 检查配置是否正确加载
	if dbConfig.Host == "" {
		return fmt.Errorf("数据库配置未正确加载")
	}

	// 构建数据库连接字符串
	var dsn string
	switch dbConfig.Driver {
	case "mysql":
		// 基础 DSN
		base := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
			dbConfig.Username, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Database)
		// 查询参数
		tz := dbConfig.Timezone
		if tz == "" {
			tz = "Local"
		}
		params := fmt.Sprintf("charset=utf8mb4&parseTime=True&loc=%s&timeout=%ds&readTimeout=%ds&writeTimeout=%ds",
			tz, dbConfig.ConnTimeoutSec, dbConfig.ReadTimeoutSec, dbConfig.WriteTimeoutSec)
		if dbConfig.ExtraParams != "" {
			params = params + "&" + dbConfig.ExtraParams
		}
		dsn = fmt.Sprintf("%s?%s", base, params)
	default:
		return fmt.Errorf("不支持的数据库驱动: %s", dbConfig.Driver)
	}

	// 配置GORM日志
	var gormConfig *gorm.Config
	if logConfig.LogSQL {
		gormConfig = &gorm.Config{
			Logger:                 &gormLoggerImpl{},
			PrepareStmt:            dbConfig.PrepareStmt,
			SkipDefaultTransaction: dbConfig.SkipDefaultTransaction,
		}
	} else {
		gormConfig = &gorm.Config{
			PrepareStmt:            dbConfig.PrepareStmt,
			SkipDefaultTransaction: dbConfig.SkipDefaultTransaction,
		}
	}

	myLogger.Info("数据库连接字符串", zap.String("driver", dbConfig.Driver))

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		return fmt.Errorf("打开数据库连接失败: %v", err)
	}

	// 获取通用数据库对象 sql.DB 以配置连接池
	sqlDB, err = DB.DB()
	if err != nil {
		return fmt.Errorf("获取数据库对象失败: %v", err)
	}

	// 设置连接池参数 - 从配置中读取
	maxOpenConns := dbConfig.MaxOpenConns
	if maxOpenConns <= 0 {
		maxOpenConns = 25 // 默认值
	}

	maxIdleConns := dbConfig.MaxIdleConns
	if maxIdleConns <= 0 {
		maxIdleConns = 25 // 默认值
	}

	connMaxLifetime := dbConfig.ConnMaxLifetime
	connMaxIdleTime := dbConfig.ConnMaxIdleTime

	sqlDB.SetMaxOpenConns(maxOpenConns)                                    // 最大打开连接数
	sqlDB.SetMaxIdleConns(maxIdleConns)                                    // 最大空闲连接数
	sqlDB.SetConnMaxLifetime(time.Duration(connMaxLifetime) * time.Second) // 连接最大生命周期
	if connMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(time.Duration(connMaxIdleTime) * time.Second)
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("数据库连接测试失败: %v", err)
	}

	// 注册BaseDO的GORM Hooks
	if err := RegisterBaseDOHooks(DB); err != nil {
		return fmt.Errorf("注册GORM Hooks失败: %v", err)
	}

	myLogger.Info("GORM数据库连接初始化成功")
	return nil
}

// gormLoggerImpl GORM日志记录器实现
type gormLoggerImpl struct {
	gormLogger.Interface
}

// LogMode 设置日志模式
func (l *gormLoggerImpl) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	return gormLogger.Default.LogMode(level)
}

// Info 记录信息日志
func (l *gormLoggerImpl) Info(ctx context.Context, msg string, data ...interface{}) {
	myLogger.InfoCtx(ctx, fmt.Sprintf(msg, data...))
}

// Warn 记录警告日志
func (l *gormLoggerImpl) Warn(ctx context.Context, msg string, data ...interface{}) {
	myLogger.WarnCtx(ctx, fmt.Sprintf(msg, data...))
}

// Error 记录错误日志
func (l *gormLoggerImpl) Error(ctx context.Context, msg string, data ...interface{}) {
	myLogger.ErrorCtx(ctx, fmt.Sprintf(msg, data...))
}

// Trace 记录SQL执行轨迹
func (l *gormLoggerImpl) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	// 使用我们的日志系统记录SQL执行信息
	sql, rows := fc()
	elapsed := time.Since(begin)

	fields := []zap.Field{
		zap.String("sql", sql),
		zap.Int64("rows", rows),
		zap.Duration("elapsed", elapsed),
	}

	// 慢 SQL 阈值
	slow := time.Duration(config.GetDatabaseConfig().SlowThresholdMS) * time.Millisecond
	if slow <= 0 {
		slow = 200 * time.Millisecond
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
		myLogger.ErrorCtx(ctx, "SQL执行", fields...)
	} else if elapsed > slow {
		myLogger.WarnCtx(ctx, "慢SQL", fields...)
	} else {
		myLogger.DebugCtx(ctx, "SQL执行", fields...)
	}
}

// GetDB 获取GORM数据库连接实例
func GetDB() *gorm.DB {
	return DB
}

// CloseDB 关闭数据库连接
func CloseDB() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
