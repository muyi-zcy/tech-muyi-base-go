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
)

// DB GORM数据库连接实例
var DB *gorm.DB

// sqlDB 原生数据库连接实例（用于连接池管理）
var sqlDB *sql.DB

// InitDatabase 初始化GORM数据库连接
func InitDatabase() error {
	dbConfig := config.GetDatabaseConfig()

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

	logger.Info("数据库连接字符串", zap.String("driver", dbConfig.Driver))

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
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

// GetDB 获取GORM数据库连接实例
func GetDB() *gorm.DB {
	return DB
}

// GetSqlDB 获取原生数据库连接实例
func GetSqlDB() *sql.DB {
	return sqlDB
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
