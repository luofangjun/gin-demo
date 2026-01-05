package controller

import (
	"fmt"

	"gin-project/logic"
	"gin-project/model"
	"gin-project/service"

	"github.com/gin-gonic/gin"
)

// UserController 用户控制器
type UserController struct {
	BaseController
	serviceFactory *service.Factory // 服务工厂，统一管理服务
}

// NewUserController 创建用户控制器
// 通过依赖注入的方式获取服务工厂，便于测试和扩展
func NewUserController(serviceFactory *service.Factory) *UserController {
	return &UserController{
		serviceFactory: serviceFactory,
	}
}

// GetUserByID 查询用户接口 - 数据查询+缓存接口
func (uc *UserController) GetUserByID(c *gin.Context) {
	var req struct {
		ID uint `json:"id" binding:"required"`
	}

	// 绑定请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		uc.ErrorWithMsg(c, "参数错误: "+err.Error())
		return
	}

	// 调用逻辑层查询用户
	user, err := logic.GetUserByID(c.Request.Context(), req.ID)
	if err != nil {
		uc.ErrorWithMsg(c, "查询用户失败: "+err.Error())
		return
	}

	// ========== 链路追踪示例：GetUserByID -> ServiceC.Calculate/Process ==========
	// 展示如何以零侵入的方式实现链路追踪
	// 关键点：
	// 1. HTTP 请求追踪：由 TracingMiddleware 自动处理
	// 2. MySQL/Redis 追踪：由 otelgorm/redisotel 自动处理
	// 3. 外部 HTTP 调用追踪：由 TracedHTTPClient 自动处理
	// 4. 服务层追踪：由 TraceServiceFunc 装饰器自动处理

	// 从服务工厂获取服务C（所有方法自动追踪）
	serviceC := uc.serviceFactory.GetServiceC()

	// 调用计算接口（自动追踪，HTTP请求也自动追踪）
	calculateResult, err := serviceC.Calculate(c.Request.Context(), 5)
	if err != nil {
		fmt.Printf("调用计算接口失败（已记录到追踪）: %v\n", err)
	} else {
		fmt.Printf("计算接口调用成功: %s\n", calculateResult)
	}

	// 调用处理接口（自动追踪，HTTP请求也自动追踪）
	processResult, err := serviceC.Process(c.Request.Context(), "hello world")
	if err != nil {
		fmt.Printf("调用处理接口失败（已记录到追踪）: %v\n", err)
	} else {
		fmt.Printf("处理接口调用成功: %s\n", processResult)
	}

	// 返回成功响应（包含 trace_id）
	uc.Success(c, user)
}

// CreateUser 创建用户接口 - 数据写入接口
func (uc *UserController) CreateUser(c *gin.Context) {
	var req struct {
		Name   string `json:"name" binding:"required"`
		Email  string `json:"email" binding:"required,email"`
		Age    int    `json:"age"`
		Status int    `json:"status"`
	}

	// 绑定请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		uc.ErrorWithMsg(c, "参数错误: "+err.Error())
		return
	}

	// 创建用户对象
	user := model.User{
		Name:   req.Name,
		Email:  req.Email,
		Age:    req.Age,
		Status: req.Status,
	}

	// 调用逻辑层创建用户（传递 context 用于追踪）
	err := logic.CreateUser(c.Request.Context(), &user)
	if err != nil {
		uc.ErrorWithMsg(c, "创建用户失败: "+err.Error())
		return
	}

	// 返回成功响应
	uc.Success(c, user)
}

// UpdateUser 更新用户接口
func (uc *UserController) UpdateUser(c *gin.Context) {
	var req struct {
		ID     uint   `json:"id" binding:"required"`
		Name   string `json:"name" binding:"required"`
		Email  string `json:"email" binding:"required,email"`
		Age    int    `json:"age"`
		Status int    `json:"status"`
	}

	// 绑定请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		uc.ErrorWithMsg(c, "参数错误: "+err.Error())
		return
	}

	// 创建用户对象用于更新
	user := model.User{
		ID:     req.ID,
		Name:   req.Name,
		Email:  req.Email,
		Age:    req.Age,
		Status: req.Status,
	}

	// 调用逻辑层更新用户（传递 context 用于追踪）
	err := logic.UpdateUser(c.Request.Context(), &user)
	if err != nil {
		uc.ErrorWithMsg(c, "更新用户失败: "+err.Error())
		return
	}

	// 返回成功响应
	uc.Success(c, user)
}
