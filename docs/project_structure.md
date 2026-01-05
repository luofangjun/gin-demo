# 项目结构说明

> 本文档详细说明项目的目录结构。关于链路追踪，请参考 [链路追踪完整指南](./tracing_guide.md)。

## 目录结构

```
gin-project/
├── service/               # 服务层（外部服务调用）
│   ├── service_a.go       # 服务A实现
│   ├── service_b.go       # 服务B实现
│   ├── service_c.go       # 服务C实现
│   └── factory.go         # 服务工厂
├── pkg/                    # 公共工具包（可被外部引用）
│   ├── tracing.go          # 链路追踪装饰器
│   └── httpclient.go       # HTTP 客户端（带追踪）
├── config/                 # 配置管理
│   └── config.go          # 配置结构定义和加载
├── database/              # 数据库相关
│   ├── mysql.go           # MySQL 连接和初始化（集成 otelgorm 插件）
│   └── redis.go           # Redis 连接和初始化（集成 redisotel）
├── model/                  # 数据模型
│   └── user.go            # 用户模型
├── logic/                  # 业务逻辑层
│   ├── constants.go       # 常量定义
│   ├── user_query.go      # 用户查询逻辑
│   └── user_write.go      # 用户写入逻辑
├── controller/             # 控制器层（处理 HTTP 请求）
│   ├── base_controller.go # 基础控制器（统一响应格式）
│   ├── user_controller.go # 用户控制器
│   └── health_controller.go # 健康检查控制器
├── middleware/             # 中间件
│   ├── tracing.go         # 追踪中间件
│   ├── logger.go          # 日志中间件
│   └── recovery.go        # 恢复中间件
├── router/                 # 路由配置
│   └── route.go           # 路由定义
├── docs/                   # 文档目录
│   ├── README.md          # 文档索引
│   ├── project_structure.md  # 项目结构说明
│   ├── tracing_guide.md   # 链路追踪完整指南
│   └── best_practices.md  # 服务层最佳实践
├── main.go                 # 应用入口
├── conf.yaml              # 配置文件
├── go.mod                  # Go 模块定义
├── go.sum                  # Go 模块校验和
└── README.md              # 项目说明
```

## 目录说明

### `/pkg` - 公共工具包
可被外部项目引用的公共工具包：
- `tracing.go`: 链路追踪装饰器（`TraceServiceFunc`）
- `httpclient.go`: HTTP 客户端（使用 `imroc/req` v3，集成 OpenTelemetry 追踪）

### `/internal` - 内部包
不对外暴露的内部包，用于存放项目内部使用的代码。

### `/config` - 配置管理
- 配置结构定义
- 配置文件加载
- 配置验证

### `/database` - 数据库
- MySQL 连接管理（集成 `otelgorm` 插件，自动追踪）
- Redis 连接管理（集成 `redisotel`，自动追踪）
- 连接池配置

### `/model` - 数据模型
- GORM 模型定义
- 数据表结构映射

### `/logic` - 业务逻辑层
- 业务逻辑实现
- 数据访问抽象
- 缓存逻辑

### `/controller` - 控制器层
- HTTP 请求处理
- 参数验证
- 响应格式化
- 健康检查接口

### `/middleware` - 中间件
- 追踪中间件：自动追踪 HTTP 请求
- 日志中间件：请求日志记录
- 恢复中间件：panic 恢复和错误处理

### `/router` - 路由配置
- 路由定义
- 路由分组
- 中间件注册

## 设计原则

### 1. 分层架构
```
Controller (控制器层)
    ↓
Logic (业务逻辑层)
    ↓
Model (数据模型层)
    ↓
Database (数据访问层)
```

### 2. 依赖方向
- Controller 依赖 Logic
- Logic 依赖 Model 和 Database
- 避免循环依赖

### 3. 命名规范
- 包名使用小写字母
- 文件名使用下划线分隔（snake_case）
- 结构体和方法使用驼峰命名（PascalCase）

### 4. 错误处理
- 统一使用 BaseController 返回错误
- 错误信息包含 trace_id（用于日志关联和问题排查）
- 使用辅助函数简化错误记录

### 5. 链路追踪
- HTTP 请求：中间件自动追踪
- HTTP 客户端：使用 `pkg.HTTPClient`，自动追踪
- MySQL 操作：`otelgorm` 插件自动追踪
- Redis 操作：`redisotel` 自动追踪
- 服务层：使用 `TraceServiceFunc` 装饰器（可选）
- 零代码入侵原则

## 最佳实践

1. **新增功能时**：
   - 在对应的层级添加代码
   - 遵循分层架构原则
   - 使用统一的错误处理

2. **添加新接口**：
   - 在 controller 中添加处理方法
   - 在 logic 中实现业务逻辑
   - 在 router 中注册路由

3. **使用链路追踪**：
   - HTTP 请求：中间件自动追踪
   - HTTP 调用：使用 `pkg.HTTPClient`
   - 数据库操作：传递 `context`，自动追踪
   - 服务层：使用 `TraceServiceFunc` 装饰器（可选）

4. **错误处理**：
   - 使用 BaseController 的方法返回错误
   - 错误信息要清晰明确
   - 记录到追踪系统

