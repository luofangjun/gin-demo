package middleware

import (
	"context"
	"log"
	"time"

	"gin-project/config"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracer trace.Tracer
	// noopTracer 无操作追踪器，用于追踪未启用时
	noopTracer = otel.Tracer("noop")
)

// InitTracing 初始化追踪
func InitTracing(cfg *config.Config) {
	// 检查是否启用追踪
	if !cfg.Tracing.Enabled {
		log.Println("追踪功能未启用")
		tracer = noopTracer
		return
	}

	// 设置全局传播器
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	// 获取追踪端点，默认使用本地 Jaeger
	endpoint := cfg.Tracing.Endpoint
	if endpoint == "" {
		endpoint = "localhost:4317"
	}

	// 配置批量导出参数
	batchSize := cfg.Tracing.BatchSize
	if batchSize <= 0 {
		batchSize = 512 // 默认批量大小
	}

	batchTimeout := time.Duration(cfg.Tracing.BatchTimeout) * time.Second
	if batchTimeout <= 0 {
		batchTimeout = 5 * time.Second // 默认批量超时
	}

	// 创建 gRPC 导出器
	exporter, err := otlptracegrpc.New(
		context.Background(),
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(), // 在生产环境中应使用安全连接
	)
	if err != nil {
		log.Printf("创建 OTLP 导出器失败: %v，将使用无操作跟踪器", err)
		// 如果连接失败，使用无操作跟踪器，不影响服务运行
		tp := sdktrace.NewTracerProvider()
		otel.SetTracerProvider(tp)
		tracer = noopTracer
		return
	}

	// 创建资源
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.App.Name),
		),
	)
	if err != nil {
		log.Fatalf("创建资源失败: %v", err)
	}

	// 配置采样率（用于性能优化）
	sampleRate := cfg.Tracing.SampleRate
	if sampleRate <= 0 {
		sampleRate = 1.0 // 默认100%采样（开发环境）
	} else if sampleRate > 1.0 {
		sampleRate = 1.0 // 最大100%
	}

	var sampler sdktrace.Sampler
	if sampleRate >= 1.0 {
		// 100%采样（开发/测试环境）
		sampler = sdktrace.AlwaysSample()
		log.Printf("追踪已启用：100%% 采样率（开发模式）")
	} else {
		// 按比例采样（生产环境推荐）
		sampler = sdktrace.TraceIDRatioBased(sampleRate)
		log.Printf("追踪已启用：%.1f%% 采样率（生产模式）", sampleRate*100)
	}

	// 创建跟踪提供者，配置采样率和批量导出
	// 批量导出配置优化性能：减少网络往返，降低性能开销
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(
			exporter,
			sdktrace.WithMaxExportBatchSize(batchSize),    // 批量大小：每次导出的span数量
			sdktrace.WithBatchTimeout(batchTimeout),        // 批量超时：超过此时间即使未达到批量大小也会导出
			sdktrace.WithExportTimeout(30*time.Second),    // 导出超时：防止导出操作阻塞太久
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler), // 采样率控制，减少性能开销
	)

	// 设置全局跟踪提供者
	otel.SetTracerProvider(tp)

	// 创建全局tracer
	tracer = otel.Tracer(cfg.App.Name)

	// 程序退出时刷新跟踪
	cleanup := func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("关闭跟踪提供者失败: %v", err)
		}
	}

	// 设置清理函数到配置中，以便在main函数中调用
	cfg.Tracing.Cleanup = cleanup
}

// TracingMiddleware 追踪中间件
// 自动为所有 HTTP 请求创建追踪 span，提取和传播 TraceID
func TracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果追踪未启用，直接跳过
		if tracer == noopTracer {
			c.Next()
			return
		}

		// 从请求头中提取追踪上下文（支持 W3C Trace Context 标准）
		ctx := otel.GetTextMapPropagator().Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

		// 开始新的 span（使用路由路径作为操作名）
		ctx, span := tracer.Start(ctx, c.FullPath(),
			trace.WithSpanKind(trace.SpanKindServer),
		)
		defer span.End()

		// 将上下文传递给后续处理（确保 TraceID 能够传播）
		c.Request = c.Request.WithContext(ctx)

		// 记录请求信息
		span.SetAttributes(
			semconv.HTTPMethodKey.String(c.Request.Method),
			semconv.HTTPURLKey.String(c.Request.URL.String()),
			semconv.HTTPRouteKey.String(c.FullPath()),
			semconv.NetPeerIPKey.String(c.ClientIP()),
		)

		// 处理请求
		c.Next()

		// 记录响应信息
		span.SetAttributes(
			semconv.HTTPStatusCodeKey.Int(c.Writer.Status()),
		)
	}
}
