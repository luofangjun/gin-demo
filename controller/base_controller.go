package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
)

// APIResponse 定义统一的API响应格式
type APIResponse struct {
	Code    int         `json:"code"`               // 状态码
	Message string      `json:"message"`            // 消息提示
	Data    interface{} `json:"data,omitempty"`     // 数据字段
	TraceID string      `json:"trace_id,omitempty"` // 追踪ID（链路追踪，用于日志关联和问题排查）
}

// BaseController 基础控制器结构
type BaseController struct{}

// getTraceID 从上下文中提取追踪ID
func (bc *BaseController) getTraceID(c *gin.Context) string {
	ctx := c.Request.Context()

	// 从 OpenTelemetry span 中获取 trace ID
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().TraceID().String()
	}

	return ""
}

// Success 成功响应
func (bc *BaseController) Success(c *gin.Context, data interface{}) {
	traceID := bc.getTraceID(c)
	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "success",
		Data:    data,
		TraceID: traceID,
	})
}

// Error 错误响应
func (bc *BaseController) Error(c *gin.Context, code int, message string) {
	traceID := bc.getTraceID(c)
	c.JSON(http.StatusOK, APIResponse{
		Code:    code,
		Message: message,
		Data:    nil,
		TraceID: traceID,
	})
}

// ErrorWithMsg 错误响应（带自定义消息）
func (bc *BaseController) ErrorWithMsg(c *gin.Context, message string) {
	traceID := bc.getTraceID(c)
	c.JSON(http.StatusOK, APIResponse{
		Code:    400,
		Message: message,
		Data:    nil,
		TraceID: traceID,
	})
}

