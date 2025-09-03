package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/muyi-zcy/tech-muyi-base-go/config"
	"github.com/muyi-zcy/tech-muyi-base-go/logger"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

// DB 数据库连接实例
var DB *sql.DB

// InitDatabase 初始化数据库连接
func InitDatabase() error {
	dbConfig := config.GetDatabaseConfig()

	// 检查配置是否正确加载
	if dbConfig.Host == "" {
		return fmt.Errorf("数据库配置未正确加载")
	}

	// 构建数据库连接字符串
	var dataSourceName string
	switch dbConfig.Driver {
	case "mysql":
		dataSourceName = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
			dbConfig.Username, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Database)
	default:
		return fmt.Errorf("不支持的数据库驱动: %s", dbConfig.Driver)
	}

	logger.Info("数据库连接字符串", zap.String("dataSourceName", dataSourceName))
	var err error
	DB, err = sql.Open(dbConfig.Driver, dataSourceName)
	if err != nil {
		return fmt.Errorf("打开数据库连接失败: %v", err)
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

	DB.SetMaxOpenConns(maxOpenConns)                                    // 最大打开连接数
	DB.SetMaxIdleConns(maxIdleConns)                                    // 最大空闲连接数
	DB.SetConnMaxLifetime(time.Duration(connMaxLifetime) * time.Second) // 连接最大生命周期

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = DB.PingContext(ctx); err != nil {
		return fmt.Errorf("数据库连接测试失败: %v", err)
	}

	return nil
}

// GetDB 获取数据库连接实例
func GetDB() *sql.DB {
	return DB
}

// CloseDB 关闭数据库连接
func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
