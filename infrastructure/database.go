package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/muyi-zcy/tech-muyi-base-go/config"
	"github.com/muyi-zcy/tech-muyi-base-go/logger"
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
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			dbConfig.Username, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Database)
	default:
		return fmt.Errorf("不支持的数据库驱动: %s", dbConfig.Driver)
	}

	// 配置GORM日志
	var gormConfig *gorm.Config
	if logConfig.LogSQL {
		gormConfig = &gorm.Config{
			Logger: &gormLoggerImpl{},
		}
	} else {
		gormConfig = &gorm.Config{}
	}

	logger.Info("数据库连接字符串", zap.String("driver", dbConfig.Driver))

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

	sqlDB.SetMaxOpenConns(maxOpenConns)                                    // 最大打开连接数
	sqlDB.SetMaxIdleConns(maxIdleConns)                                    // 最大空闲连接数
	sqlDB.SetConnMaxLifetime(time.Duration(connMaxLifetime) * time.Second) // 连接最大生命周期

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("数据库连接测试失败: %v", err)
	}

	logger.Info("GORM数据库连接初始化成功")
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
	logger.InfoCtx(ctx, fmt.Sprintf(msg, data...))
}

// Warn 记录警告日志
func (l *gormLoggerImpl) Warn(ctx context.Context, msg string, data ...interface{}) {
	logger.WarnCtx(ctx, fmt.Sprintf(msg, data...))
}

// Error 记录错误日志
func (l *gormLoggerImpl) Error(ctx context.Context, msg string, data ...interface{}) {
	logger.ErrorCtx(ctx, fmt.Sprintf(msg, data...))
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

	if err != nil {
		fields = append(fields, zap.Error(err))
		logger.ErrorCtx(ctx, "SQL执行", fields...)
	} else if elapsed > 200*time.Millisecond {
		logger.WarnCtx(ctx, "慢SQL", fields...)
	} else {
		logger.DebugCtx(ctx, "SQL执行", fields...)
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
