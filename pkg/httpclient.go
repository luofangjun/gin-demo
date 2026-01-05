package pkg

import (
	"net/http"
	"time"

	"github.com/imroc/req/v3"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var (
	// tracingEnabled 追踪是否启用（通过 InitHTTPClient 设置）
	tracingEnabled bool
	// httpClient 全局 HTTP 客户端
	httpClient *req.Client
)

// InitHTTPClient 初始化 HTTP 客户端
// 根据追踪开关决定是否启用追踪，优化性能
func InitHTTPClient(enabled bool) {
	tracingEnabled = enabled

	client := req.C().
		SetTimeout(10*time.Second).
		SetCommonHeader("Content-Type", "application/json")

	// 仅在追踪启用时包装 Transport，避免不必要的性能开销
	if enabled {
		// 获取底层 http.Client 并设置带追踪的 Transport
		httpClientInstance := client.GetClient()
		baseTransport := httpClientInstance.Transport
		if baseTransport == nil {
			baseTransport = http.DefaultTransport
		}
		// 包装 Transport 以支持 OpenTelemetry 追踪
		httpClientInstance.Transport = otelhttp.NewTransport(baseTransport)
	}

	httpClient = client
}

// HTTPClient 获取全局带追踪的 HTTP 客户端
// 使用 imroc/req v3 封装，根据配置决定是否集成 OpenTelemetry 追踪
// 所有 HTTP 请求自动追踪（如果启用），无需手动添加追踪代码
func HTTPClient() *req.Client {
	if httpClient == nil {
		// 如果未初始化，使用默认配置（启用追踪）
		InitHTTPClient(true)
	}
	return httpClient
}
