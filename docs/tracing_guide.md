# 链路追踪完整指南

## 目录

1. [概述](#概述)
2. [快速开始](#快速开始)
3. [零代码入侵实现](#零代码入侵实现)
4. [服务层追踪](#服务层追踪)
5. [使用示例](#使用示例)
6. [最佳实践](#最佳实践)
7. [Jaeger 配置](#jaeger-配置)

---

## 概述

本项目已按照最佳实践实现了**零代码入侵**的链路追踪，使用 OpenTelemetry 官方/社区标准插件，只需在初始化时配置一次，即可自动追踪所有跨网络操作。

### 核心优势

- ✅ **零代码入侵**：业务代码完全不需要追踪相关代码
- ✅ **自动追踪**：所有跨网络操作自动追踪
- ✅ **统一管理**：追踪配置集中在初始化代码中
- ✅ **易于维护**：使用官方/社区标准插件，稳定可靠

---

## 快速开始

### 1. 配置追踪

在 `conf.yaml` 中启用追踪：

```yaml
tracing:
  enabled: true
  endpoint: "localhost:4317"  # Jaeger OTLP gRPC 端点
```

### 2. 启动 Jaeger

```bash
docker run -d -p 16686:16686 -p 4317:4317 jaegertracing/all-in-one:latest
```

### 3. 启动应用

启动应用后，所有 HTTP 请求、数据库操作、Redis 操作都会自动追踪。

### 4. 查看追踪数据

1. 访问 Jaeger UI: `http://localhost:16686`
2. 选择服务名称（在 `conf.yaml` 中配置的 `app.name`）
3. 点击 "Find Traces" 查看追踪数据
4. 或使用 API 响应中的 `trace_id` 直接搜索

---

## 零代码入侵实现

### 1. HTTP 请求追踪 ✅

**使用插件：** `go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp`

**实现方式：**
- 在 `pkg/httpclient.go` 中创建全局 `HTTPClient`（使用 `imroc/req` v3 封装）
- 使用 `otelhttp.NewTransport` 包装 HTTP Transport
- 自动注入 TraceID 到请求头，自动追踪请求耗时

**代码位置：** `pkg/httpclient.go`

```go
// HTTPClient 全局带追踪的 HTTP 客户端
// 使用 imroc/req v3 封装，集成 OpenTelemetry 追踪，零代码入侵
var HTTPClient = func() *req.Client {
    client := req.C().
        SetTimeout(10*time.Second).
        SetCommonHeader("Content-Type", "application/json")
    
    // 获取底层 http.Client 并设置带追踪的 Transport
    httpClient := client.GetClient()
    baseTransport := httpClient.Transport
    if baseTransport == nil {
        baseTransport = http.DefaultTransport
    }
    // 包装 Transport 以支持 OpenTelemetry 追踪
    httpClient.Transport = otelhttp.NewTransport(baseTransport)
    
    return client
}()
```

**业务代码中使用（零代码入侵）：**

```go
// service/service_c.go
func (s *ServiceC) Calculate(ctx context.Context, number int) (string, error) {
    // 使用带追踪的 HTTP 客户端，自动注入 TraceID 到请求头
    url := s.baseURL + "/api/calculate"
    resp, err := pkg.HTTPClient().R().
        SetContext(ctx).
        SetBody(reqBody).
        Post(url)
    // 自动追踪，无需任何手动代码
    return result, nil
}
```

**优势：**
- ✅ 业务代码无需修改，只需使用 `pkg.HTTPClient`
- ✅ 自动注入 TraceID 到请求头，实现跨服务追踪
- ✅ 自动记录请求耗时、状态码等信息

---

### 2. MySQL 数据库追踪 ✅

**使用插件：** `github.com/uptrace/opentelemetry-go-extra/otelgorm`

**实现方式：**
- 在 `database/mysql.go` 初始化时注册 `otelgorm` 插件
- 所有 GORM 操作自动追踪，无需修改业务代码

**代码位置：** `database/mysql.go`

```go
// 连接到目标数据库
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{...})

// 【最佳实践】使用 otelgorm 插件，自动追踪所有数据库操作（零代码入侵）
if err := db.Use(otelgorm.NewPlugin()); err != nil {
    panic("failed to register otelgorm plugin: " + err.Error())
}
```

**业务代码中使用（零代码入侵）：**

```go
// logic/user_query.go
func GetUserByID(ctx context.Context, id uint) (*model.User, error) {
    // 使用带追踪的数据库客户端，自动追踪
    err := database.DB.WithContext(ctx).First(user, id).Error
    // 自动追踪，自动记录 SQL、执行时间等信息
    return user, nil
}
```

**自动追踪的信息：**
- `db.system`: "mysql"
- `db.operation`: "first", "find", "create", "update", "updates", "delete"
- `db.sql.table`: 表名
- `db.result.rows_affected`: 受影响的行数
- `db.result.not_found`: 是否未找到记录（First 操作）

**优势：**
- ✅ 只需在初始化时配置一次，所有操作自动追踪
- ✅ 自动记录 SQL 语句、执行时间、表名等信息
- ✅ 业务代码完全不需要修改

---

### 3. Redis 缓存追踪 ✅

**使用插件：** `github.com/redis/go-redis/extra/redisotel/v9`

**实现方式：**
- 在 `database/redis.go` 初始化时调用 `redisotel.InstrumentTracing`
- 所有 Redis 操作自动追踪，无需修改业务代码

**代码位置：** `database/redis.go`

```go
RedisClient = redis.NewClient(&redis.Options{...})

// 【最佳实践】使用 redisotel 自动追踪所有 Redis 操作（零代码入侵）
if err := redisotel.InstrumentTracing(RedisClient); err != nil {
    panic("failed to instrument redis tracing: " + err.Error())
}
```

**业务代码中使用（零代码入侵）：**

```go
// logic/user_query.go
func GetUserByID(ctx context.Context, id uint) (*model.User, error) {
    // 从Redis获取数据（使用带追踪的客户端，自动追踪）
    jsonData, err := database.RedisClient.Get(ctx, cacheKey).Result()
    // 自动追踪，自动记录命令、键名、执行时间等信息
    return user, nil
}
```

**自动追踪的信息：**
- `db.system`: "redis"
- `db.operation`: "get", "set", "del", "exists", "ping"
- `db.redis.key`: Redis 键名
- `db.redis.expiration_seconds`: 过期时间（Set 操作）

**优势：**
- ✅ 只需在初始化时配置一次，所有操作自动追踪
- ✅ 自动记录 Redis 命令、键名、执行时间等信息
- ✅ 业务代码完全不需要修改

---

### 4. HTTP 中间件追踪 ✅

**实现方式：**
- 在 `middleware/tracing.go` 中实现 `TracingMiddleware`
- 自动为所有 HTTP 请求创建追踪 span

**代码位置：** `middleware/tracing.go`

```go
// TracingMiddleware 追踪中间件
// 自动为所有 HTTP 请求创建追踪 span，提取和传播 TraceID
func TracingMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 从请求头中提取追踪上下文
        ctx := otel.GetTextMapPropagator().Extract(c.Request.Context(), 
            propagation.HeaderCarrier(c.Request.Header))
        
        // 开始新的 span
        ctx, span := tracer.Start(ctx, c.FullPath(),
            trace.WithSpanKind(trace.SpanKindServer),
        )
        defer span.End()
        
        // 将上下文传递给后续处理
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
```

**使用方式：**

```go
// router/route.go
func SetupRouter() *gin.Engine {
    r := gin.New()
    
    // 添加全局中间件（注意顺序很重要）
    r.Use(middleware.RecoveryMiddleware()) // 恢复中间件（最先添加）
    r.Use(middleware.LoggerMiddleware())   // 日志中间件
    r.Use(middleware.TracingMiddleware())  // 追踪中间件（在日志之后）
    
    // 路由配置...
    return r
}
```

**优势：**
- ✅ 自动为所有 HTTP 请求创建追踪
- ✅ 自动提取和传播 TraceID
- ✅ 自动记录请求和响应信息

---

## 服务层追踪

### 服务函数装饰器

对于服务层的业务逻辑，我们提供了 `TraceServiceFunc` 装饰器，实现**零侵入**的追踪。

**代码位置：** `pkg/tracing.go`

```go
// TraceServiceFunc 服务函数追踪装饰器
// 用于追踪服务层的业务逻辑函数，自动创建 span、处理错误，零代码入侵
func TraceServiceFunc[T any, R any](
    operationName string,
    fn func(context.Context, T) (R, error),
    attrFunc func(context.Context, T) []attribute.KeyValue,
) func(context.Context, T) (R, error)
```

**使用示例：**

```go
// 方式1：不带属性（最简单）
process := pkg.TraceServiceFunc("ServiceC.Process", serviceC.Process, nil)

// 方式2：带属性（可选，用于重要参数）
calculate := pkg.TraceServiceFunc("ServiceC.Calculate", serviceC.Calculate,
    func(ctx context.Context, number int) []attribute.KeyValue {
        return []attribute.KeyValue{
            attribute.String("service.name", "ServiceC"),
            attribute.String("method", "Calculate"),
            attribute.Int("input.number", number),
        }
    })
```

**装饰器优势：**
- ✅ **零侵入**：业务函数不包含任何追踪代码
- ✅ **自动处理**：装饰器自动创建 span、处理错误、设置状态
- ✅ **类型安全**：使用 Go 泛型，编译时类型检查
- ✅ **灵活配置**：支持自定义属性设置（可选）

---

## 使用示例

### 完整追踪链路示例

在 `GetUserByID` 接口中实现了完整的链路追踪示例：

```
GetUserByID (HTTP请求) - TracingMiddleware 自动追踪
  ├── GetUserByID (Logic层) - 业务逻辑
  │     ├── Redis GET user:1 - redisotel 自动追踪
  │     ├── MySQL SELECT * FROM users - otelgorm 自动追踪
  │     └── Redis SET user:1 - redisotel 自动追踪
  ├── ServiceC.Calculate - TraceServiceFunc 装饰器追踪
  │     └── HTTP POST /api/calculate - pkg.HTTPClient 自动追踪
  └── ServiceC.Process - TraceServiceFunc 装饰器追踪
        └── HTTP POST /api/process - pkg.HTTPClient 自动追踪
```

### 代码实现

**Controller 层：**

```go
func (uc *UserController) GetUserByID(c *gin.Context) {
    // HTTP 追踪：由 TracingMiddleware 自动处理
    
    // 调用逻辑层查询用户
    user, err := logic.GetUserByID(c.Request.Context(), req.ID)
    // MySQL/Redis 追踪：由 otelgorm/redisotel 自动处理
    
    // 从服务工厂获取服务C（所有方法自动追踪）
    serviceC := uc.serviceFactory.GetServiceC()
    
    // 调用计算接口（自动追踪，HTTP请求也自动追踪）
    calculateResult, err := serviceC.Calculate(c.Request.Context(), 5)
    // HTTP 追踪：由 pkg.HTTPClient 自动处理
    // 业务逻辑追踪：由 TraceServiceFunc 装饰器处理
    
    // 返回成功响应（包含 trace_id）
    uc.Success(c, user)
}
```

**Logic 层：**

```go
func GetUserByID(ctx context.Context, id uint) (*model.User, error) {
    // 从Redis获取数据（自动追踪）
    jsonData, err := database.RedisClient.Get(ctx, cacheKey).Result()
    
    // 从数据库查询（自动追踪）
    err = database.DB.WithContext(ctx).First(user, id).Error
    
    // 将查询结果存入缓存（异步执行，自动追踪）
    go func() {
        database.RedisClient.Set(ctx, cacheKey, string(jsonData), ttl)
    }()
    
    return user, nil
}
```

**Service 层：**

```go
// ServiceC.Calculate - 纯业务逻辑，无追踪代码
func (s *ServiceC) Calculate(ctx context.Context, number int) (string, error) {
    reqBody := map[string]int{"number": number}
    
    // 使用带追踪的 HTTP 客户端，自动注入 TraceID 到请求头
    url := s.baseURL + "/api/calculate"
    resp, err := pkg.HTTPClient().R().
        SetContext(ctx).
        SetBody(reqBody).
        Post(url)
    // 自动追踪，无需任何手动代码
    
    return resp.String(), nil
}
```

---

## 最佳实践

### ✅ 推荐做法

1. **始终传递 context**
   ```go
   // ✅ 正确：传递 context
   err := database.DB.WithContext(ctx).First(user, id).Error
   err := database.RedisClient.Get(ctx, key).Result()
   resp, err := pkg.HTTPClient().R().SetContext(ctx).Get(url)
   ```

2. **使用标准客户端**
   - HTTP: 使用 `pkg.HTTPClient`
   - MySQL: 使用 `database.DB.WithContext(ctx)`
   - Redis: 使用 `database.RedisClient`

3. **无需手动创建 Span**
   - 所有追踪都是自动的
   - 业务代码完全不需要追踪相关代码

### ❌ 不推荐做法

1. **不传递 context**
   ```go
   // ❌ 错误：不传递 context，无法建立追踪链路
   err := database.DB.First(user, id).Error
   ```

2. **手动创建 Span**
   ```go
   // ❌ 错误：手动创建 span，增加代码复杂度
   ctx, span := tracer.Start(ctx, "operation")
   defer span.End()
   ```

3. **使用未追踪的客户端**
   ```go
   // ❌ 错误：使用标准库 http.Client，无法追踪
   resp, err := http.Get(url)
   ```

### 配置位置总结

所有追踪配置都在初始化时完成：

- **HTTP 中间件**: `middleware/tracing.go` - `TracingMiddleware()`
- **HTTP 客户端**: `pkg/httpclient.go` - `HTTPClient` 全局变量
- **MySQL**: `database/mysql.go` - `InitMysql()` 中注册 `otelgorm` 插件
- **Redis**: `database/redis.go` - `InitRedis()` 中调用 `redisotel.InstrumentTracing`

### 扩展新功能

当需要添加新的 HTTP 调用、数据库操作或 Redis 操作时：

1. **HTTP 调用**：直接使用 `pkg.HTTPClient`，自动追踪
2. **数据库操作**：直接使用 `database.DB.WithContext(ctx)`，自动追踪
3. **Redis 操作**：直接使用 `database.RedisClient`，自动追踪

**无需任何额外配置，完全零代码入侵！**

---

## Jaeger 配置

### 启动 Jaeger

使用 Docker 启动 Jaeger：

```bash
docker run -d \
  --name jaeger \
  -p 16686:16686 \
  -p 4317:4317 \
  jaegertracing/all-in-one:latest
```

### 配置应用

在 `conf.yaml` 中配置：

```yaml
app:
  name: "gin-project"  # 服务名称，会在 Jaeger 中显示

tracing:
  enabled: true
  endpoint: "localhost:4317"  # Jaeger OTLP gRPC 端点
```

### 查看追踪数据

1. 访问 Jaeger UI: `http://localhost:16686`
2. 选择服务名称（`gin-project`）
3. 点击 "Find Traces" 查看追踪数据
4. 或使用 API 响应中的 `trace_id` 直接搜索

### 追踪链路示例

在 Jaeger UI 中，你会看到如下完整的追踪链路：

```
[Span 1] HTTP GET /api/user/query (TracingMiddleware 自动创建)
  ├── [Span 2] Redis GET user:1 (redisotel 自动追踪)
  ├── [Span 3] MySQL SELECT * FROM users (otelgorm 自动追踪)
  ├── [Span 4] ServiceC.Calculate (TraceServiceFunc 装饰器)
  │      └── [Span 5] HTTP POST /api/calculate (pkg.HTTPClient 自动追踪)
  └── [Span 6] ServiceC.Process (TraceServiceFunc 装饰器)
         └── [Span 7] HTTP POST /api/process (pkg.HTTPClient 自动追踪)
```

所有追踪都是自动的，无需手动创建 span！

---

## 参考资料

- [OpenTelemetry 官方文档](https://opentelemetry.io/docs/)
- [otelhttp 文档](https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp)
- [otelgorm 文档](https://github.com/uptrace/opentelemetry-go-extra/tree/main/otelgorm)
- [redisotel 文档](https://github.com/redis/go-redis/tree/master/extra/redisotel)
- [imroc/req v3 文档](https://github.com/imroc/req)
