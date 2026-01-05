package pkg

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// Tracer 服务层追踪器
// 用于服务层业务逻辑的追踪，HTTP 请求追踪由 middleware.TracingMiddleware() 处理
var Tracer = otel.Tracer("gin-service-tracer")

// 注意：HTTP 请求追踪已由 middleware.TracingMiddleware() 自动处理
// 本文件仅提供服务层业务逻辑追踪装饰器

// TraceServiceFunc 服务函数追踪装饰器
// 用于追踪服务层的业务逻辑函数，自动创建 span、处理错误，零代码入侵
//
// 参数说明：
//   - operationName: 操作名称，用于标识该操作
//   - fn: 要追踪的业务函数
//   - attrFunc: 可选的属性设置函数，传入 nil 表示不设置属性
//
// 使用示例:
//
//	// 不带属性
//	tracedFunc := TraceServiceFunc("ServiceA.MethodA", serviceA.MethodA, nil)
//	result, err := tracedFunc(ctx, userID)
//
//	// 带属性
//	tracedFunc := TraceServiceFunc("ServiceA.MethodA", serviceA.MethodA,
//		func(ctx context.Context, userID uint) []attribute.KeyValue {
//			return []attribute.KeyValue{
//				attribute.Int("user.id", int(userID)),
//				attribute.String("service.name", "ServiceA"),
//			}
//		})
func TraceServiceFunc[T any, R any](
	operationName string,
	fn func(context.Context, T) (R, error),
	attrFunc func(context.Context, T) []attribute.KeyValue,
) func(context.Context, T) (R, error) {
	return func(ctx context.Context, arg T) (R, error) {
		ctx, span := Tracer.Start(ctx, operationName)
		defer span.End()

		// 设置属性（可选）
		if attrFunc != nil {
			if attrs := attrFunc(ctx, arg); len(attrs) > 0 {
				span.SetAttributes(attrs...)
			}
		}

		// 执行业务函数
		result, err := fn(ctx, arg)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		return result, err
	}
}
