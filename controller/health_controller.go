package controller

import (
	"gin-project/database"

	"github.com/gin-gonic/gin"
)

// HealthController 健康检查控制器
type HealthController struct {
	BaseController
}

// Health 健康检查接口
func (hc *HealthController) Health(c *gin.Context) {
	hc.Success(c, gin.H{
		"status": "ok",
		"service": "gin-project",
	})
}

// Readiness 就绪检查接口
func (hc *HealthController) Readiness(c *gin.Context) {
	// 检查数据库连接
	if database.DB == nil {
		hc.Error(c, 503, "数据库未初始化")
		return
	}

	// 检查 Redis 连接
	if database.RedisClient == nil {
		hc.Error(c, 503, "Redis未初始化")
		return
	}

	// 测试数据库连接
	sqlDB, err := database.DB.DB()
	if err != nil {
		hc.Error(c, 503, "数据库连接失败: "+err.Error())
		return
	}

	if err := sqlDB.Ping(); err != nil {
		hc.Error(c, 503, "数据库连接失败: "+err.Error())
		return
	}

	// 测试 Redis 连接
	ctx := c.Request.Context()
	if err := database.RedisClient.Ping(ctx).Err(); err != nil {
		hc.Error(c, 503, "Redis连接失败: "+err.Error())
		return
	}

	hc.Success(c, gin.H{
		"status": "ready",
		"database": "ok",
		"redis": "ok",
	})
}

// Liveness 存活检查接口
func (hc *HealthController) Liveness(c *gin.Context) {
	hc.Success(c, gin.H{
		"status": "alive",
	})
}

