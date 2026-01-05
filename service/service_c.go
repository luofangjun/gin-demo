package service

import (
	"context"
	"encoding/json"
	"fmt"

	"gin-project/pkg"

	"go.opentelemetry.io/otel/attribute"
)

// APIResponse 标准 API 响应格式
type APIResponse struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

// ServiceCInterface 服务C的接口定义
type ServiceCInterface interface {
	Calculate(ctx context.Context, number int) (string, error)
	Process(ctx context.Context, content string) (string, error)
}

// ServiceC 服务C结构体
type ServiceC struct {
	baseURL string // API 基础URL
}

// NewServiceC 创建服务C实例
func NewServiceC(baseURL string) *ServiceC {
	return &ServiceC{
		baseURL: baseURL,
	}
}

// Calculate 计算接口 - 纯业务逻辑，无追踪代码
// HTTP 请求追踪：由 pkg.HTTPClient 自动处理（零代码入侵）
func (s *ServiceC) Calculate(ctx context.Context, number int) (string, error) {
	// 构建请求体
	reqBody := map[string]int{"number": number}

	// 使用带追踪的 HTTP 客户端，自动注入 TraceID 到请求头
	url := s.baseURL + "/api/calculate"
	resp, err := pkg.HTTPClient().R().
		SetContext(ctx).
		SetBody(reqBody).
		Post(url)
	if err != nil {
		return "", fmt.Errorf("调用计算接口失败: %v", err)
	}

	// 解析响应
	var apiResp APIResponse
	if err := resp.UnmarshalJson(&apiResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	// 检查业务状态码（code==0 表示成功）
	if apiResp.Code != 0 {
		return "", fmt.Errorf("计算接口返回错误: %s", apiResp.Message)
	}

	// 将 data 转换为 JSON 字符串返回
	dataBytes, err := json.Marshal(apiResp.Data)
	if err != nil {
		return "", fmt.Errorf("序列化响应数据失败: %v", err)
	}

	return string(dataBytes), nil
}

// Process 处理接口 - 纯业务逻辑，无追踪代码
// HTTP 请求追踪：由 pkg.HTTPClient 自动处理（零代码入侵）
func (s *ServiceC) Process(ctx context.Context, content string) (string, error) {
	// 构建请求体
	reqBody := map[string]string{"content": content}

	// 使用带追踪的 HTTP 客户端，自动注入 TraceID 到请求头
	url := s.baseURL + "/api/process"
	resp, err := pkg.HTTPClient().R().
		SetContext(ctx).
		SetBody(reqBody).
		Post(url)
	if err != nil {
		return "", fmt.Errorf("调用处理接口失败: %v", err)
	}

	// 解析响应
	var apiResp APIResponse
	if err := resp.UnmarshalJson(&apiResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	// 检查业务状态码（code==0 表示成功）
	if apiResp.Code != 0 {
		return "", fmt.Errorf("处理接口返回错误: %s", apiResp.Message)
	}

	// 将 data 转换为 JSON 字符串返回
	dataBytes, err := json.Marshal(apiResp.Data)
	if err != nil {
		return "", fmt.Errorf("序列化响应数据失败: %v", err)
	}

	return string(dataBytes), nil
}

// ServiceCWithTrace 带追踪的服务C包装器
// HTTP 请求追踪：由 pkg.HTTPClient 自动处理
// 业务逻辑追踪：由 TraceServiceFunc 装饰器处理
type ServiceCWithTrace struct {
	*ServiceC
	calculate func(context.Context, int) (string, error)
	process   func(context.Context, string) (string, error)
}

// NewServiceCWithTrace 创建带追踪的服务C实例
func NewServiceCWithTrace(baseURL string) *ServiceCWithTrace {
	serviceC := NewServiceC(baseURL)
	return &ServiceCWithTrace{
		ServiceC: serviceC,
		// 追踪业务逻辑层（HTTP 请求已由 pkg.HTTPClient 自动追踪）
		calculate: pkg.TraceServiceFunc("ServiceC.Calculate", serviceC.Calculate,
			func(ctx context.Context, number int) []attribute.KeyValue {
				return []attribute.KeyValue{
					attribute.String("service.name", "ServiceC"),
					attribute.String("method", "Calculate"),
					attribute.Int("input.number", number),
				}
			}),
		process: pkg.TraceServiceFunc("ServiceC.Process", serviceC.Process,
			func(ctx context.Context, content string) []attribute.KeyValue {
				return []attribute.KeyValue{
					attribute.String("service.name", "ServiceC"),
					attribute.String("method", "Process"),
					attribute.String("input.content", content),
				}
			}),
	}
}

// Calculate 带追踪的计算方法
func (s *ServiceCWithTrace) Calculate(ctx context.Context, number int) (string, error) {
	return s.calculate(ctx, number)
}

// Process 带追踪的处理方法
func (s *ServiceCWithTrace) Process(ctx context.Context, content string) (string, error) {
	return s.process(ctx, content)
}
