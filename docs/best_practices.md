# 服务层最佳实践

> 本文档说明服务层架构设计和链路追踪的最佳实践。关于项目结构，请参考 [项目结构说明](./project_structure.md)；关于完整的追踪指南，请参考 [链路追踪完整指南](./tracing_guide.md)。

## 架构设计原则

### 1. 服务结构体封装

**原则：** 将相关的方法封装到服务结构体中，而不是使用独立的函数。

**优势：**
- ✅ 代码组织更清晰
- ✅ 便于管理服务的多个方法
- ✅ 易于扩展和维护

**示例：**
```go
// ServiceC 服务C结构体，封装多个方法
type ServiceC struct {
    baseURL string // API 基础URL
}

// Calculate 方法 - 纯业务逻辑，无追踪代码
func (s *ServiceC) Calculate(ctx context.Context, number int) (string, error) {
    // 业务逻辑：调用外部 API
    reqBody := map[string]int{"number": number}
    resp, err := pkg.HTTPClient().R().
        SetContext(ctx).
        SetBody(reqBody).
        Post(s.baseURL + "/api/calculate")
    return resp.String(), err
}

// Process 方法 - 纯业务逻辑，无追踪代码
func (s *ServiceC) Process(ctx context.Context, content string) (string, error) {
    // 业务逻辑：调用外部 API
    reqBody := map[string]string{"content": content}
    resp, err := pkg.HTTPClient().R().
        SetContext(ctx).
        SetBody(reqBody).
        Post(s.baseURL + "/api/process")
    return resp.String(), err
}
```

### 2. 依赖注入（Dependency Injection）

**原则：** 通过构造函数注入依赖，而不是在方法内部创建依赖。

**优势：**
- ✅ 便于单元测试（可以注入 mock 对象）
- ✅ 解耦服务之间的依赖
- ✅ 便于扩展和维护

**示例：**
```go
// ✅ 好的实践：通过构造函数注入配置
func NewServiceC(baseURL string) *ServiceC {
    return &ServiceC{
        baseURL: baseURL,
    }
}

// ❌ 不好的实践：在方法内部硬编码配置
func (s *ServiceC) Calculate(ctx context.Context, number int) (string, error) {
    baseURL := "http://localhost:8081" // 不推荐：硬编码配置
    resp, err := pkg.HTTPClient().R().Post(baseURL + "/api/calculate")
    return resp.String(), err
}
```

### 3. 接口解耦

**原则：** 使用接口定义服务之间的依赖关系，而不是直接依赖具体实现。

**优势：**
- ✅ 便于测试（可以注入 mock 实现）
- ✅ 便于替换实现（如替换为不同的服务B实现）
- ✅ 降低耦合度

**示例：**
```go
// 定义服务C的接口
type ServiceCInterface interface {
    Calculate(ctx context.Context, number int) (string, error)
    Process(ctx context.Context, content string) (string, error)
}

// ServiceC 实现接口
type ServiceC struct {
    baseURL string
}

func (s *ServiceC) Calculate(ctx context.Context, number int) (string, error) {
    // 实现
}

func (s *ServiceC) Process(ctx context.Context, content string) (string, error) {
    // 实现
}
```

### 4. 服务工厂模式

**原则：** 使用服务工厂统一管理服务的创建和依赖关系。

**优势：**
- ✅ 统一管理服务的生命周期
- ✅ 便于扩展新服务
- ✅ 便于维护服务之间的依赖关系

**示例：**
```go
// Factory 服务工厂
type Factory struct {
    serviceC *ServiceCWithTrace
}

// NewFactory 创建服务工厂
func NewFactory() *Factory {
    // 创建服务C（带追踪，统一管理配置）
    serviceC := NewServiceCWithTrace("http://localhost:8081")

    return &Factory{
        serviceC: serviceC,
    }
}
```

### 5. 控制器依赖注入

**原则：** 控制器通过依赖注入的方式获取服务，而不是在控制器内部创建服务。

**优势：**
- ✅ 便于测试
- ✅ 便于扩展
- ✅ 代码更清晰

**示例：**
```go
// UserController 用户控制器
type UserController struct {
    BaseController
    serviceFactory *service.Factory // 通过依赖注入获取服务工厂
}

// NewUserController 创建用户控制器
func NewUserController(serviceFactory *service.Factory) *UserController {
    return &UserController{
        serviceFactory: serviceFactory,
    }
}

// 使用服务
func (uc *UserController) GetUserByID(c *gin.Context) {
    // 从服务工厂获取服务
    serviceC := uc.serviceFactory.GetServiceC()
    result, err := serviceC.Calculate(c.Request.Context(), 5)
    // ...
}
```

## 完整示例

### 1. 定义服务接口

```go
// ServiceCInterface 服务C的接口定义
type ServiceCInterface interface {
    Calculate(ctx context.Context, number int) (string, error)
    Process(ctx context.Context, content string) (string, error)
}
```

### 2. 实现服务（纯业务逻辑）

```go
// ServiceC 服务C结构体
type ServiceC struct {
    baseURL string // API 基础URL
}

func NewServiceC(baseURL string) *ServiceC {
    return &ServiceC{
        baseURL: baseURL,
    }
}

// Calculate 方法 - 纯业务逻辑，无追踪代码
func (s *ServiceC) Calculate(ctx context.Context, number int) (string, error) {
    reqBody := map[string]int{"number": number}
    url := s.baseURL + "/api/calculate"
    resp, err := pkg.HTTPClient().R().
        SetContext(ctx).
        SetBody(reqBody).
        Post(url)
    if err != nil {
        return "", err
    }
    // 解析响应...
    return result, nil
}

// Process 方法 - 纯业务逻辑，无追踪代码
func (s *ServiceC) Process(ctx context.Context, content string) (string, error) {
    reqBody := map[string]string{"content": content}
    url := s.baseURL + "/api/process"
    resp, err := pkg.HTTPClient().R().
        SetContext(ctx).
        SetBody(reqBody).
        Post(url)
    if err != nil {
        return "", err
    }
    // 解析响应...
    return result, nil
}
```

### 3. 创建带追踪的包装器

```go
// ServiceCWithTrace 带追踪的服务C包装器
type ServiceCWithTrace struct {
    *ServiceC
    calculate func(context.Context, int) (string, error)
    process   func(context.Context, string) (string, error)
}

func NewServiceCWithTrace(baseURL string) *ServiceCWithTrace {
    serviceC := NewServiceC(baseURL)
    return &ServiceCWithTrace{
        ServiceC: serviceC,
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

func (s *ServiceCWithTrace) Calculate(ctx context.Context, number int) (string, error) {
    return s.calculate(ctx, number)
}

func (s *ServiceCWithTrace) Process(ctx context.Context, content string) (string, error) {
    return s.process(ctx, content)
}
```

### 4. 创建服务工厂

```go
// Factory 服务工厂（位于 service/factory.go）
type Factory struct {
    serviceC *ServiceCWithTrace
}

func NewFactory() *Factory {
    serviceC := NewServiceCWithTrace("http://localhost:8081")
    return &Factory{
        serviceC: serviceC,
    }
}

func (f *Factory) GetServiceC() *ServiceCWithTrace {
    return f.serviceC
}
```

### 5. 在路由中使用

```go
import (
    "gin-project/controller"
    "gin-project/service"
)

func SetupRouter() *gin.Engine {
    // 创建服务工厂（位于 service 包）
    serviceFactory := service.NewFactory()
    
    // 创建控制器（依赖注入服务工厂）
    userCtrl := controller.NewUserController(serviceFactory)
    
    // 使用控制器
    api.POST("/user/query", userCtrl.GetUserByID)
}
```

## 扩展新服务

当需要添加新服务时，只需：

1. **定义服务接口**（如果需要）
2. **实现服务**（纯业务逻辑）
3. **创建带追踪的包装器**
4. **在服务工厂中注册**
5. **在控制器中使用**

**参考示例：** 当前项目中的 `ServiceC` 就是完整的实现示例，位于 `service/service_c.go` 和 `service/factory.go`。

## 服务使用示例

### 问题场景

当需要调用一个服务的多个方法（如 ServiceA 的 MethodA 和 MethodB）时，如果每个方法都单独写装饰器，会导致代码重复。

### 解决方案：服务结构体 + 包装器模式

**完整示例：**

参考 `service/service_c.go` 文件，其中包含完整的实现：

```go
// 1. 定义纯业务服务（无追踪代码）
type ServiceC struct {
    baseURL string
}

func NewServiceC(baseURL string) *ServiceC {
    return &ServiceC{baseURL: baseURL}
}

// Calculate 方法 - 纯业务逻辑，无追踪代码
func (s *ServiceC) Calculate(ctx context.Context, number int) (string, error) {
    reqBody := map[string]int{"number": number}
    url := s.baseURL + "/api/calculate"
    resp, err := pkg.HTTPClient().R().
        SetContext(ctx).
        SetBody(reqBody).
        Post(url)
    // 解析响应...
    return result, nil
}

// Process 方法 - 纯业务逻辑，无追踪代码
func (s *ServiceC) Process(ctx context.Context, content string) (string, error) {
    reqBody := map[string]string{"content": content}
    url := s.baseURL + "/api/process"
    resp, err := pkg.HTTPClient().R().
        SetContext(ctx).
        SetBody(reqBody).
        Post(url)
    // 解析响应...
    return result, nil
}

// 2. 创建带追踪的包装器
type ServiceCWithTrace struct {
    *ServiceC
    calculate func(context.Context, int) (string, error)
    process   func(context.Context, string) (string, error)
}

func NewServiceCWithTrace(baseURL string) *ServiceCWithTrace {
    serviceC := NewServiceC(baseURL)
    return &ServiceCWithTrace{
        ServiceC: serviceC,
        calculate: pkg.TraceServiceFunc("ServiceC.Calculate", serviceC.Calculate, ...),
        process: pkg.TraceServiceFunc("ServiceC.Process", serviceC.Process, ...),
    }
}

// 3. 使用方式
func (uc *UserController) GetUserByID(c *gin.Context) {
    // 从服务工厂获取服务C（所有方法自动追踪）
    serviceC := uc.serviceFactory.GetServiceC()
    
    // 调用方法（自动追踪）
    result, err := serviceC.Calculate(c.Request.Context(), 5)
    result, err := serviceC.Process(c.Request.Context(), "hello world")
}
```

### 关于属性设置的最佳实践

**属性设置是可选的**

`TraceServiceFunc` 的第三个参数 `attrFunc` 是可选的，可以传入 `nil`：

```go
// 方式1：不带属性（最简单）
methodB := pkg.TraceServiceFunc("ServiceA.MethodB", serviceA.MethodB, nil)

// 方式2：带属性（用于重要参数或调试）
methodA := pkg.TraceServiceFunc("ServiceA.MethodA", serviceA.MethodA,
    func(ctx context.Context, userID uint) []attribute.KeyValue {
        return []attribute.KeyValue{
            attribute.Int("user.id", int(userID)), // 只记录重要参数
        }
    })
```

**何时设置属性？**

**推荐设置属性的情况：**
- ✅ 关键业务参数（如用户ID、订单ID）
- ✅ 用于问题排查的重要信息
- ✅ 需要统计分析的维度数据

**不推荐设置属性的情况：**
- ❌ 敏感信息（密码、token）
- ❌ 大量数据（整个对象、长字符串）
- ❌ 频繁变化的参数（时间戳、随机数）

**属性设置示例：**

```go
// ✅ 好的实践：只记录关键信息
calculate := pkg.TraceServiceFunc("ServiceC.Calculate", serviceC.Calculate,
    func(ctx context.Context, number int) []attribute.KeyValue {
        return []attribute.KeyValue{
            attribute.Int("input.number", number),        // 关键业务参数
            attribute.String("service.name", "ServiceC"), // 服务标识
        }
    })

// ❌ 不好的实践：记录过多或不必要的信息
calculate := pkg.TraceServiceFunc("ServiceC.Calculate", serviceC.Calculate,
    func(ctx context.Context, number int) []attribute.KeyValue {
        return []attribute.KeyValue{
            attribute.Int("input.number", number),
            attribute.String("timestamp", time.Now().String()), // 不必要
            attribute.String("random", uuid.New().String()),    // 不必要
        }
    })
```

### HTTP 调用示例

**使用 `pkg.HTTPClient` 进行外部 HTTP 调用：**

```go
// ServiceC.Calculate - 纯业务逻辑，无追踪代码
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
    
    // 解析响应（标准 API 响应格式：{"code":0, "message":"xxx", "data":{...}}）
    var apiResp APIResponse
    if err := resp.UnmarshalJson(&apiResp); err != nil {
        return "", fmt.Errorf("解析响应失败: %v", err)
    }
    
    // 检查业务状态码（code==0 表示成功）
    if apiResp.Code != 0 {
        return "", fmt.Errorf("计算接口返回错误: %s", apiResp.Message)
    }
    
    // 返回 data 字段
    dataBytes, _ := json.Marshal(apiResp.Data)
    return string(dataBytes), nil
}
```

**关键点：**
- ✅ 使用 `pkg.HTTPClient()`（函数调用），自动追踪所有 HTTP 请求
- ✅ 传递 `ctx`，自动建立追踪链路
- ✅ 业务代码完全不需要追踪相关代码

## 优势总结

1. **零代码侵入**：业务方法不包含任何追踪代码
2. **避免重复**：一个服务的所有方法统一管理，无需为每个方法重复写装饰器
3. **灵活配置**：每个方法可以独立配置是否设置属性
4. **依赖注入**：便于测试和扩展
5. **接口解耦**：降低服务之间的耦合度
6. **统一管理**：服务工厂统一管理所有服务
7. **易于扩展**：添加新服务只需几个步骤
8. **类型安全**：使用 Go 接口和泛型，编译时检查

