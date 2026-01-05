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
// ServiceA 服务A结构体，封装多个方法
type ServiceA struct {
    serviceB ServiceBInterface // 通过依赖注入获取其他服务
}

// MethodA 方法A - 纯业务逻辑，无追踪代码
func (s *ServiceA) MethodA(ctx context.Context, userID uint) (string, error) {
    // 业务逻辑
    // 调用服务B的方法A
    result, err := s.serviceB.MethodA(ctx, processedData)
    return result, err
}

// MethodB 方法B - 纯业务逻辑，无追踪代码
func (s *ServiceA) MethodB(ctx context.Context, data string) (string, error) {
    // 业务逻辑
    return result, nil
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
// ✅ 好的实践：通过构造函数注入依赖
func NewServiceA(serviceB ServiceBInterface) *ServiceA {
    return &ServiceA{
        serviceB: serviceB,
    }
}

// ❌ 不好的实践：在方法内部创建依赖
func (s *ServiceA) MethodA(ctx context.Context, userID uint) (string, error) {
    serviceB := NewServiceBWithTrace() // 不推荐：硬编码依赖
    result, err := serviceB.MethodA(ctx, data)
    return result, err
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
// 定义服务B的接口
type ServiceBInterface interface {
    MethodA(ctx context.Context, data string) (string, error)
}

// ServiceA 依赖接口，而不是具体实现
type ServiceA struct {
    serviceB ServiceBInterface // 依赖接口
}

// ServiceB 实现接口
type ServiceB struct{}

func (s *ServiceB) MethodA(ctx context.Context, data string) (string, error) {
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
// ServiceFactory 服务工厂
type ServiceFactory struct {
    serviceB ServiceBInterface
    serviceA *ServiceAWithTrace
}

// NewServiceFactory 创建服务工厂
func NewServiceFactory() *ServiceFactory {
    // 创建服务B（带追踪）
    serviceB := NewServiceBWithTrace()

    // 创建服务A（依赖注入服务B，带追踪）
    serviceA := NewServiceAWithTrace(serviceB)

    return &ServiceFactory{
        serviceB: serviceB,
        serviceA: serviceA,
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
    serviceFactory *ServiceFactory // 通过依赖注入获取服务工厂
}

// NewUserController 创建用户控制器
func NewUserController(serviceFactory *ServiceFactory) *UserController {
    return &UserController{
        serviceFactory: serviceFactory,
    }
}

// 使用服务
func (uc *UserController) GetUserByID(c *gin.Context) {
    // 从服务工厂获取服务
    serviceA := uc.serviceFactory.GetServiceA()
    result, err := serviceA.MethodA(c.Request.Context(), userID)
    // ...
}
```

## 完整示例

### 1. 定义服务接口

```go
// ServiceBInterface 服务B的接口定义
type ServiceBInterface interface {
    MethodA(ctx context.Context, data string) (string, error)
}
```

### 2. 实现服务（纯业务逻辑）

```go
// ServiceB 服务B结构体
type ServiceB struct{}

func NewServiceB() *ServiceB {
    return &ServiceB{}
}

// MethodA 方法A - 纯业务逻辑，无追踪代码
func (s *ServiceB) MethodA(ctx context.Context, data string) (string, error) {
    // 业务逻辑
    return result, nil
}
```

### 3. 创建带追踪的包装器

```go
// ServiceBWithTrace 带追踪的服务B包装器
type ServiceBWithTrace struct {
    *ServiceB
    methodA func(context.Context, string) (string, error)
}

func NewServiceBWithTrace() *ServiceBWithTrace {
    serviceB := NewServiceB()
    return &ServiceBWithTrace{
        ServiceB: serviceB,
        methodA: pkg.TraceServiceFunc("ServiceB.MethodA", serviceB.MethodA,
            func(ctx context.Context, data string) []attribute.KeyValue {
                return []attribute.KeyValue{
                    attribute.String("service.name", "ServiceB"),
                    attribute.String("method", "MethodA"),
                }
            }),
    }
}

func (s *ServiceBWithTrace) MethodA(ctx context.Context, data string) (string, error) {
    return s.methodA(ctx, data)
}
```

### 4. 创建服务A（依赖服务B）

```go
// ServiceA 服务A结构体
type ServiceA struct {
    serviceB ServiceBInterface // 依赖接口
}

func NewServiceA(serviceB ServiceBInterface) *ServiceA {
    return &ServiceA{
        serviceB: serviceB,
    }
}

// MethodA 方法A - 调用服务B的方法A
func (s *ServiceA) MethodA(ctx context.Context, userID uint) (string, error) {
    // 业务逻辑
    result, err := s.serviceB.MethodA(ctx, processedData)
    return result, err
}
```

### 5. 创建服务工厂

```go
// Factory 服务工厂（位于 service/factory.go）
type Factory struct {
    serviceB ServiceBInterface
    serviceA *ServiceAWithTrace
}

func NewFactory() *Factory {
    serviceB := NewServiceBWithTrace()
    serviceA := NewServiceAWithTrace(serviceB)
    return &Factory{
        serviceB: serviceB,
        serviceA: serviceA,
    }
}
```

### 6. 在路由中使用

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

**示例：添加 ServiceC**

```go
// 1. 定义服务C
type ServiceC struct{}

func NewServiceC() *ServiceC {
    return &ServiceC{}
}

func (s *ServiceC) Process(ctx context.Context, data string) (string, error) {
    // 业务逻辑
    return result, nil
}

// 2. 创建带追踪的包装器
type ServiceCWithTrace struct {
    *ServiceC
    process func(context.Context, string) (string, error)
}

func NewServiceCWithTrace() *ServiceCWithTrace {
    serviceC := NewServiceC()
    return &ServiceCWithTrace{
        ServiceC: serviceC,
        process: pkg.TraceServiceFunc("ServiceC.Process", serviceC.Process, nil),
    }
}

// 3. 在服务工厂中注册（位于 service/factory.go）
func NewFactory() *Factory {
    // ... 现有代码 ...
    serviceC := NewServiceCWithTrace()
    return &Factory{
        // ... 现有字段 ...
        serviceC: serviceC,
    }
}

// 4. 在控制器中使用
func (uc *UserController) SomeMethod(c *gin.Context) {
    serviceC := uc.serviceFactory.GetServiceC()
    result, err := serviceC.Process(c.Request.Context(), data)
    // ...
}
```

## 服务使用示例

### 问题场景

当需要调用一个服务的多个方法（如 ServiceA 的 MethodA 和 MethodB）时，如果每个方法都单独写装饰器，会导致代码重复。

### 解决方案：服务结构体 + 包装器模式

**完整示例：**

```go
// 1. 定义纯业务服务（无追踪代码）
type ServiceA struct {
    serviceB ServiceBInterface // 通过依赖注入获取其他服务
}

func NewServiceA(serviceB ServiceBInterface) *ServiceA {
    return &ServiceA{
        serviceB: serviceB,
    }
}

// MethodA 方法A - 纯业务逻辑，无追踪代码
func (s *ServiceA) MethodA(ctx context.Context, userID uint) (string, error) {
    // 模拟业务处理
    time.Sleep(100 * time.Millisecond)
    processedData := fmt.Sprintf("user_%d_data", userID)
    
    // 调用服务B（通过接口调用，自动建立链路追踪关系）
    result, err := s.serviceB.MethodA(ctx, processedData)
    if err != nil {
        return "", fmt.Errorf("服务A方法A调用服务B方法A失败: %v", err)
    }
    
    return result, nil
}

// MethodB 方法B - 纯业务逻辑，无追踪代码
func (s *ServiceA) MethodB(ctx context.Context, data string) (string, error) {
    // 模拟业务处理
    time.Sleep(50 * time.Millisecond)
    return fmt.Sprintf("ServiceA.MethodB processed: %s", data), nil
}

// 2. 创建带追踪的包装器
type ServiceAWithTrace struct {
    *ServiceA
    methodA func(context.Context, uint) (string, error)
    methodB func(context.Context, string) (string, error)
}

func NewServiceAWithTrace(serviceB ServiceBInterface) *ServiceAWithTrace {
    serviceA := NewServiceA(serviceB)
    return &ServiceAWithTrace{
        ServiceA: serviceA,
        // 方法A：带属性设置（记录关键业务参数）
        methodA: pkg.TraceServiceFunc("ServiceA.MethodA", serviceA.MethodA,
            func(ctx context.Context, userID uint) []attribute.KeyValue {
                return []attribute.KeyValue{
                    attribute.String("service.name", "ServiceA"),
                    attribute.String("method", "MethodA"),
                    attribute.Int("user.id", int(userID)),
                }
            }),
        // 方法B：不带属性设置（属性设置是可选的）
        methodB: pkg.TraceServiceFunc("ServiceA.MethodB", serviceA.MethodB, nil),
    }
}

// MethodA 带追踪的方法A
func (s *ServiceAWithTrace) MethodA(ctx context.Context, userID uint) (string, error) {
    return s.methodA(ctx, userID)
}

// MethodB 带追踪的方法B
func (s *ServiceAWithTrace) MethodB(ctx context.Context, data string) (string, error) {
    return s.methodB(ctx, data)
}

// 3. 使用方式
func (uc *UserController) GetUserByID(c *gin.Context) {
    // 从服务工厂获取服务A（所有方法自动追踪）
    serviceA := uc.serviceFactory.GetServiceA()
    
    // 调用方法A（自动追踪）
    resultA, err := serviceA.MethodA(c.Request.Context(), userID)
    
    // 调用方法B（自动追踪）
    resultB, err := serviceA.MethodB(c.Request.Context(), "test-data")
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
methodA := pkg.TraceServiceFunc("ServiceA.MethodA", serviceA.MethodA,
    func(ctx context.Context, userID uint) []attribute.KeyValue {
        return []attribute.KeyValue{
            attribute.Int("user.id", int(userID)),        // 关键业务参数
            attribute.String("service.name", "ServiceA"), // 服务标识
        }
    })

// ❌ 不好的实践：记录过多或不必要的信息
methodA := pkg.TraceServiceFunc("ServiceA.MethodA", serviceA.MethodA,
    func(ctx context.Context, userID uint) []attribute.KeyValue {
        return []attribute.KeyValue{
            attribute.Int("user.id", int(userID)),
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
    resp, err := pkg.HTTPClient.R().
        SetContext(ctx).
        SetBody(reqBody).
        Post(url)
    if err != nil {
        return "", fmt.Errorf("调用计算接口失败: %v", err)
    }
    
    // 检查响应状态码
    if !resp.IsSuccess() {
        return "", fmt.Errorf("计算接口返回错误状态码: %d, 响应: %s", resp.StatusCode, resp.String())
    }
    
    return resp.String(), nil
}
```

**关键点：**
- ✅ 使用 `pkg.HTTPClient`，自动追踪所有 HTTP 请求
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

