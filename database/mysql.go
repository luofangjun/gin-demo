package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"gin-project/config"

	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitMysql 初始化MySQL数据库连接
func InitMysql(cfg *config.Config) {
	// 先连接到系统数据库
	sysDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=%s&parseTime=%t&loc=%s",
		cfg.Database.Mysql.Username,
		cfg.Database.Mysql.Password,
		cfg.Database.Mysql.Host,
		cfg.Database.Mysql.Port,
		cfg.Database.Mysql.Charset,
		cfg.Database.Mysql.ParseTime,
		cfg.Database.Mysql.Loc,
	)

	// 连接系统数据库
	sysDB, err := gorm.Open(mysql.Open(sysDSN), &gorm.Config{
		Logger: logger.Default,
	})
	if err != nil {
		panic("failed to connect to system database: " + err.Error())
	}

	// 创建数据库（如果不存在）
	dbName := cfg.Database.Mysql.Database
	err = sysDB.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", dbName)).Error
	if err != nil {
		panic("failed to create database: " + err.Error())
	}

	// 关闭系统数据库连接
	sqlDB, _ := sysDB.DB()
	sqlDB.Close()

	// 构建目标数据库的DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
		cfg.Database.Mysql.Username,
		cfg.Database.Mysql.Password,
		cfg.Database.Mysql.Host,
		cfg.Database.Mysql.Port,
		cfg.Database.Mysql.Database,
		cfg.Database.Mysql.Charset,
		cfg.Database.Mysql.ParseTime,
		cfg.Database.Mysql.Loc,
	)

	// 配置GORM日志级别
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // 慢SQL阈值
			LogLevel:                  logger.Info, // 日志级别
			IgnoreRecordNotFoundError: false,       // 忽略ErrRecordNotFound错误
			Colorful:                  true,        // 彩色打印
		},
	)

	// 连接到目标数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})

	if err != nil {
		panic("failed to connect database: " + err.Error())
	}

	// 【最佳实践】使用 otelgorm 插件，自动追踪所有数据库操作（零代码入侵）
	// 仅在追踪启用时注册插件，避免不必要的性能开销
	if cfg.Tracing.Enabled {
		if err := db.Use(otelgorm.NewPlugin()); err != nil {
			panic("failed to register otelgorm plugin: " + err.Error())
		}
		log.Println("MySQL 追踪已启用")
	} else {
		log.Println("MySQL 追踪未启用（性能优化模式）")
	}

	// 设置连接池
	sqlDB, err = db.DB()
	if err != nil {
		panic("failed to get database instance: " + err.Error())
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(cfg.Database.Mysql.MaxIdleConns) // 设置最大空闲连接数
	sqlDB.SetMaxOpenConns(cfg.Database.Mysql.MaxOpenConns) // 设置最大打开连接数
	sqlDB.SetConnMaxLifetime(time.Hour)                    // 设置连接最大生存时间

	DB = db
}
