package router

import (
	"gin-project/controller"
	"gin-project/middleware"
	"gin-project/service"

	"github.com/gin-gonic/gin"
)

// SetupRouter 配置路由信息
func SetupRouter() *gin.Engine {
	// 使用 gin.New() 而不是 gin.Default()，因为我们需要自定义中间件
	r := gin.New()

	// 添加全局中间件（注意顺序很重要）
	r.Use(middleware.RecoveryMiddleware()) // 恢复中间件（最先添加，确保能捕获所有 panic）
	r.Use(middleware.LoggerMiddleware())   // 日志中间件
	r.Use(middleware.TracingMiddleware())  // 追踪中间件（在日志之后，确保日志能记录追踪信息）

	// 健康检查路由（不需要追踪）
	healthCtrl := &controller.HealthController{}
	r.GET("/health", healthCtrl.Health)
	r.GET("/readiness", healthCtrl.Readiness)
	r.GET("/liveness", healthCtrl.Liveness)

	// 创建控制器实例
	// 创建服务工厂（统一管理所有服务）
	serviceFactory := service.NewFactory()

	// 创建用户控制器（依赖注入服务工厂）
	userCtrl := controller.NewUserController(serviceFactory)

	// API 路由组
	api := r.Group("/api")
	{
		// 用户相关接口
		// 注意：HTTP 请求追踪已由 TracingMiddleware 自动处理，无需装饰器
		users := api.Group("/user")
		{
			users.POST("/query", userCtrl.GetUserByID)
			users.POST("/create", userCtrl.CreateUser)
			users.PUT("/update", userCtrl.UpdateUser)
		}
	}

	return r
}
