package middleware

import (
	"gin-project/controller"

	"github.com/gin-gonic/gin"
)

// RecoveryMiddleware 恢复中间件
// 捕获 panic 并返回统一错误响应，确保服务不会因为 panic 而崩溃
func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		// 使用 BaseController 返回统一错误格式（包含 trace_id）
		baseCtrl := &controller.BaseController{}
		baseCtrl.Error(c, 500, "服务器内部错误")
		c.Abort()
	})
}
