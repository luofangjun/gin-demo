# pprof 性能分析使用指南

> 本文档说明如何使用项目集成的 pprof 性能分析工具。

## 启用条件

pprof 仅在 **debug 模式**下启用，通过 `conf.yaml` 中的 `app.mode` 配置：

```yaml
app:
  mode: debug  # 设置为 debug 时启用 pprof
```

**注意：** 生产环境应设置为 `release` 或 `production`，禁用 pprof 以避免安全风险。

## 访问路径

pprof 路由挂载在 `/debug/pprof/` 路径下：

- **首页**：`http://localhost:8080/debug/pprof/`
- **命令行参数**：`http://localhost:8080/debug/pprof/cmdline`
- **CPU 性能分析**：`http://localhost:8080/debug/pprof/profile`
- **符号表**：`http://localhost:8080/debug/pprof/symbol`
- **执行追踪**：`http://localhost:8080/debug/pprof/trace`
- **内存分配**：`http://localhost:8080/debug/pprof/allocs`
- **阻塞分析**：`http://localhost:8080/debug/pprof/block`
- **Goroutine 分析**：`http://localhost:8080/debug/pprof/goroutine`
- **堆内存分析**：`http://localhost:8080/debug/pprof/heap`
- **互斥锁分析**：`http://localhost:8080/debug/pprof/mutex`
- **线程创建分析**：`http://localhost:8080/debug/pprof/threadcreate`

## 使用方法

### 1. 查看性能分析首页

在浏览器中访问：
```
http://localhost:8080/debug/pprof/
```

会显示所有可用的性能分析端点。

### 2. CPU 性能分析

**通过浏览器：**
```
http://localhost:8080/debug/pprof/profile?seconds=30
```
参数 `seconds` 指定采样时长（秒），默认 30 秒。

**通过命令行：**
```bash
# 下载 CPU 性能数据（30秒采样）
curl.exe http://localhost:8080/debug/pprof/profile?seconds=30 -o cpu.prof

# 使用 go tool pprof 分析
go tool pprof cpu.prof
```

**交互式命令：**
```bash
# 直接进入交互式分析
go tool pprof http://localhost:8080/debug/pprof/profile?seconds=30
```

### 3. 堆内存分析

**通过命令行：**
```bash
# 下载堆内存数据
curl.exe http://localhost:8080/debug/pprof/heap -o heap.prof

# 使用 go tool pprof 分析
go tool pprof heap.prof
```

**交互式命令：**
```bash
# 直接进入交互式分析
go tool pprof http://localhost:8080/debug/pprof/heap
```

### 4. Goroutine 分析

**查看 Goroutine 堆栈：**
```bash
# 下载 Goroutine 数据
curl.exe http://localhost:8080/debug/pprof/goroutine -o goroutine.prof

# 使用 go tool pprof 分析
go tool pprof goroutine.prof
```

**交互式命令：**
```bash
# 直接进入交互式分析
go tool pprof http://localhost:8080/debug/pprof/goroutine
```

### 5. 执行追踪

**下载追踪数据：**
```bash
# 下载执行追踪（5秒）
curl.exe http://localhost:8080/debug/pprof/trace?seconds=5 -o trace.out

# 使用 go tool trace 分析
go tool trace trace.out
```

## 常用 pprof 命令

在 `go tool pprof` 交互式界面中，常用命令：

- `top` - 显示占用资源最多的函数
- `top10` - 显示前 10 个占用资源最多的函数
- `list <函数名>` - 显示指定函数的源代码和性能数据
- `web` - 生成 SVG 可视化图表（需要安装 Graphviz）
- `png` - 生成 PNG 可视化图表
- `svg` - 生成 SVG 可视化图表
- `help` - 显示帮助信息
- `exit` 或 `quit` - 退出

## 性能分析示例

### 示例 1：分析 CPU 性能瓶颈

```bash
# 1. 启动性能采样（30秒）
go tool pprof http://localhost:8080/debug/pprof/profile?seconds=30

# 2. 在交互式界面中查看 top 函数
(pprof) top10

# 3. 查看具体函数的详细信息
(pprof) list GetUserByID

# 4. 生成可视化图表
(pprof) web
```

### 示例 2：分析内存泄漏

```bash
# 1. 获取堆内存快照
go tool pprof http://localhost:8080/debug/pprof/heap

# 2. 查看内存占用最多的函数
(pprof) top10

# 3. 查看内存分配详情
(pprof) list CreateUser

# 4. 生成内存分配图
(pprof) web
```

### 示例 3：分析 Goroutine 泄漏

```bash
# 1. 获取 Goroutine 快照
go tool pprof http://localhost:8080/debug/pprof/goroutine

# 2. 查看 Goroutine 数量
(pprof) top

# 3. 查看 Goroutine 堆栈
(pprof) list main
```

## 安全注意事项

⚠️ **重要：** pprof 端点包含敏感信息，生产环境必须禁用！

**推荐配置：**

```yaml
# 开发环境
app:
  mode: debug  # 启用 pprof

# 生产环境
app:
  mode: release  # 禁用 pprof
```

## 与链路追踪的配合使用

pprof 和链路追踪可以配合使用：

1. **通过链路追踪发现性能问题**：在 Jaeger 中查看慢请求
2. **使用 pprof 深入分析**：针对慢请求进行 CPU/内存分析
3. **优化代码**：根据分析结果优化性能瓶颈

## 参考资源

- [Go pprof 官方文档](https://pkg.go.dev/net/http/pprof)
- [Go 性能优化指南](https://go.dev/doc/diagnostics)
- [pprof 可视化工具](https://github.com/google/pprof)

