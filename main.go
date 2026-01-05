package main

import (
	"gin-project/config"
	"gin-project/database"
	"gin-project/middleware"
	"gin-project/pkg"
	"gin-project/router"
	"log"
	"os"
)

func main() {
	// 加载配置文件
	config.LoadConfig()

	// 初始化追踪（必须在数据库和HTTP客户端之前）
	middleware.InitTracing(config.Cfg)

	// 初始化 HTTP 客户端（根据追踪开关优化性能）
	pkg.InitHTTPClient(config.Cfg.Tracing.Enabled)

	// 初始化数据库连接（根据追踪开关优化性能）
	database.InitMysql(config.Cfg)
	database.InitRedis(config.Cfg)

	// 创建路由
	r := router.SetupRouter()

	// 获取端口配置，默认8080
	port := config.Cfg.App.Port
	if port == "" {
		port = "8080"
	}

	// 确保在程序退出时关闭追踪提供者
	defer func() {
		if config.Cfg.Tracing.Cleanup != nil {
			config.Cfg.Tracing.Cleanup()
		}
	}()

	// 启动服务器
	log.Printf("服务器启动在端口: %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Printf("服务器启动失败: %v", err)
		os.Exit(1)
	}
}
