# Gin项目 - 用户管理API

基于Gin框架、GORM ORM和Redis缓存实现的用户管理API项目，集成了完整的链路追踪功能。

## 项目特性

- ✅ **分层架构**：清晰的 Controller-Logic-Model 分层设计
- ✅ **链路追踪**：集成 OpenTelemetry，支持 Jaeger 追踪
- ✅ **缓存策略**：Redis 缓存优先，提升查询性能
- ✅ **统一响应**：标准化的 API 响应格式，包含 trace_id（用于日志关联和问题排查）
- ✅ **健康检查**：提供健康检查、就绪检查和存活检查接口
- ✅ **中间件支持**：追踪、日志、恢复等中间件
- ✅ **最少侵入**：链路追踪代码侵入性最小

## 项目结构

```
gin-project/
├── pkg/                    # 公共工具包
│   ├── tracing.go          # 链路追踪工具
│   └── utils.go           # 通用工具函数
├── config/                 # 配置管理
├── database/              # 数据库连接
├── model/                  # 数据模型
├── logic/                  # 业务逻辑层
├── controller/             # 控制器层
├── middleware/             # 中间件
├── router/                 # 路由配置
└── main.go                 # 应用入口
```

详细结构说明请参考 [项目结构说明](./docs/project_structure.md)

## 快速开始

### 1. 环境要求

- Go 1.24+
- MySQL 5.7+
- Redis 6.0+
- Jaeger (可选，用于链路追踪)

### 2. 安装依赖

```bash
go mod tidy
```

### 3. 配置

编辑 `conf.yaml` 配置文件：

```yaml
app:
  name: gin-project
  port: 8082
  mode: debug

database:
  mysql:
    host: 127.0.0.1
    port: 3306
    username: root
    password: 123456
    database: gin_project
    charset: utf8mb4
    parseTime: true
    loc: Local
    maxIdleConns: 10
    maxOpenConns: 100

redis:
  addr: 127.0.0.1:6379
  password: ""
  db: 0
  poolSize: 10

tracing:
  enabled: true
  endpoint: localhost:4317
```

### 4. 初始化数据库

执行 `create_tables.sql` 创建数据库表：

```bash
mysql -u root -p gin_project < create_tables.sql
```

### 5. 启动服务

```bash
go run main.go
```

服务将在 `http://localhost:8082` 启动

## API 接口

### 健康检查接口

- `GET /health` - 健康检查
- `GET /readiness` - 就绪检查（检查数据库和Redis连接）
- `GET /liveness` - 存活检查

### 用户管理接口

#### 1. 查询用户

- **接口**: `POST /api/user/query`
- **功能**: 根据ID查询用户，优先从Redis缓存获取
- **请求**:
```json
{
    "id": 1
}
```
- **响应**:
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "id": 1,
        "name": "张三",
        "email": "zhangsan@example.com",
        "age": 25,
        "status": 1
    },
    "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736"
}
```

#### 2. 创建用户

- **接口**: `POST /api/user/create`
- **功能**: 创建新用户
- **请求**:
```json
{
    "name": "张三",
    "email": "zhangsan@example.com",
    "age": 25,
    "status": 1
}
```

#### 3. 更新用户

- **接口**: `PUT /api/user/update`
- **功能**: 更新用户信息
- **请求**:
```json
{
    "id": 1,
    "name": "张三",
    "email": "zhangsan@example.com",
    "age": 26,
    "status": 1
}
```

## 链路追踪

项目集成了完整的链路追踪功能，支持：

- 自动追踪 HTTP 请求
- 追踪服务间调用
- 追踪数据库和缓存操作
- 自动包含 trace_id 到响应中（用于日志关联和问题排查）

详细使用说明请参考 [链路追踪完整指南](./docs/tracing_guide.md)

### 查看追踪数据

1. 启动 Jaeger：
```bash
docker run -d -p 16686:16686 -p 4317:4317 jaegertracing/all-in-one:latest
```

2. 访问 Jaeger UI：http://localhost:16686

3. 使用响应中的 `trace_id` 在 Jaeger UI 中查询完整的调用链路

## 技术栈

- **Web框架**: Gin
- **ORM**: GORM
- **缓存**: Redis
- **链路追踪**: OpenTelemetry + Jaeger
- **配置管理**: YAML

## 项目特点

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

### 2. 缓存策略

- 查询优先从 Redis 缓存获取
- 缓存未命中时查询数据库
- 写入操作后自动清除相关缓存

### 3. 链路追踪

- 路由层使用装饰器自动追踪
- 业务层使用辅助函数记录属性
- 最小侵入代码原则

### 4. 统一响应格式

所有接口返回统一格式，包含：
- `code`: 状态码
- `message`: 消息
- `data`: 数据
- `trace_id`: 追踪ID（用于在追踪系统中查找完整的请求链路，也可用于日志关联和问题排查）

## 开发指南

### 添加新接口

1. 在 `controller/` 中添加控制器方法
2. 在 `logic/` 中实现业务逻辑
3. 在 `router/route.go` 中注册路由

### 使用链路追踪

1. 路由层使用装饰器：
```go
api.POST("/user/query", pkg.TraceWithBusinessAttributes("GetUserByID", handler))
```

2. 业务层传递 context：
```go
result, err := logic.GetUserByID(ctx, id)
```

3. 使用辅助函数记录属性：
```go
pkg.SetSpanAttributes(ctx, attribute.String("key", "value"))
pkg.RecordErrorIfNotNil(ctx, err)
```

详细说明请参考 [链路追踪完整指南](./docs/tracing_guide.md)

## 许可证

MIT
