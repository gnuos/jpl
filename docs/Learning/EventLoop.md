# Event Loop 框架使用指南

JPL Event Loop 框架提供基于 goroutine + context 的高性能事件驱动编程能力，支持定时器、信号处理、网络 IO 等多种事件类型。

## 目录

- [核心概念](#核心概念)
- [快速开始](#快速开始)
- [API 参考](#api-参考)
- [使用示例](#使用示例)
- [最佳实践](#最佳实践)
- [高级主题](#高级主题)

## 核心概念

### 架构设计

JPL Event Loop 采用分层架构：

```
┌─────────────────────────────────────────────────────┐
│  JPL 脚本层                                          │
│  $registry.on("accept", $server, fn($e) {})         │
└─────────────────────────────────────────────────────┘
          │
          ▼
┌─────────────────────────────────────────────────────┐
│  Event Loop (ev.go)                                 │
│  - 管理 goroutine 生命周期 (context)                 │
│  - 定时器                                           │
│  - 信号处理                                         │
└─────────────────────────────────────────────────────┘
          │
          │ 通过 context 控制
          ▼
┌─────────────────────────────────────────────────────┐
│  事件源 (各自实现 goroutine)                         │
│  - net.go: acceptLoop, readLoop                     │
│  - asyncio: file_read_lines, file_read_chunks       │
└─────────────────────────────────────────────────────┘
```

### Event Loop（事件循环）

事件循环管理 context 生命周期、定时器和信号处理。各模块（网络、文件 IO）自己管理事件处理 goroutine。

### Registry（事件注册表）

通用事件注册表，支持任意事件类型：
- 存储事件处理器元数据
- 提供 `on/off/emit` 通用接口
- 各模块封装专用 API（如 `on_accept`, `on_read`）

### 工作流程

1. **创建注册表**：`ev_registry_new()` - 创建事件注册表对象
2. **注册事件**：使用 `$registry.on_xxx()` 方法注册各种事件处理器
3. **创建循环**：`ev_loop_new()` - 创建事件循环对象
4. **附加注册表**：`ev_attach(loop, registry)` - 将注册表附加到循环
5. **运行循环**：`ev_run(loop)` - 启动事件循环（阻塞）
6. **停止循环**：`ev_stop(loop)` - 停止事件循环

## 快速开始

```jpl
// 最简单的定时器示例
$registry = ev_registry_new()

// 每秒打印一次
$count = 0
$registry.on_timer(1000000, fn() {
    $count = $count + 1
    println "Tick #{$count}"
})

// Ctrl+C 退出
$registry.on_signal(2, fn() {
    ev_stop($loop)
})

// 创建并运行事件循环
$loop = ev_loop_new()
ev_attach($loop, $registry)
ev_run($loop)

println "Done!"
```

运行：
```bash
jpl run examples/network/event-loop-tutorial.jpl
```

## API 参考

### 核心函数

#### ev_registry_new()
创建新的事件注册表。

**返回值**：`EvRegistry` - 事件注册表对象

```php
$registry = ev_registry_new()
```

#### ev_loop_new()
创建新的事件循环。

**返回值**：`EvLoop` - 事件循环对象

```php
$loop = ev_loop_new()
```

#### ev_attach(loop, registry)
将注册表附加到事件循环。

**参数**：
- `loop`: EvLoop 对象
- `registry`: EvRegistry 对象

**返回值**：`true`

```php
$loop = ev_loop_new()
$registry = ev_registry_new()
ev_attach($loop, $registry)
```

#### ev_run(loop)
启动事件循环（阻塞运行）。

**参数**：
- `loop`: EvLoop 对象

```php
ev_run($loop)
```

#### ev_run_once(loop)
运行一次事件循环（非阻塞）。

**参数**：
- `loop`: EvLoop 对象

**返回值**：`true` 如果有事件被处理

```php
// 非阻塞模式，适用于需要与主循环配合的场景
while (ev_run_once($loop)) {
    // 处理其他任务
}
```

#### ev_stop(loop)
停止事件循环。

**参数**：
- `loop`: EvLoop 对象

```php
ev_stop($loop)
```

#### ev_is_running(loop)
检查事件循环是否正在运行。

**参数**：
- `loop`: EvLoop 对象

**返回值**：`true` / `false`

```php
if (ev_is_running($loop)) {
    println "Loop is running"
}
```

#### ev_timer_now()
获取当前微秒级时间戳。

**返回值**：`int64` - 微秒级时间戳

```php
$now = ev_timer_now()
println "Current time (us): #{$now}"
```

### 注册表方法

#### $registry.on_timer(interval_us, callback)
注册周期性定时器。

**参数**：
- `interval_us`: 间隔时间（微秒）
- `callback`: 回调函数，无参数

**返回值**：`timer_id`（可用于注销）

```php
// 每 5 秒执行一次
$registry.on_timer(5000000, fn() {
    println "Heartbeat"
})
```

#### $registry.on_timer_once(interval_us, callback)
注册一次性定时器。

**参数**：
- `interval_us`: 延迟时间（微秒）
- `callback`: 回调函数，无参数

```php
// 10 秒后超时
$registry.on_timer_once(10000000, fn() {
    println "Timeout!"
    ev_stop($loop)
})
```

#### $registry.on_signal(sig_num, callback)
注册信号处理器。

**参数**：
- `sig_num`: 信号编号（2=SIGINT, 15=SIGTERM）
- `callback`: 回调函数，无参数

```php
// Ctrl+C 处理
$registry.on_signal(2, fn() {
    println "Shutting down..."
    ev_stop($loop)
})
```

#### $registry.on_read(fd, callback)
注册文件描述符可读事件。

**参数**：
- `fd`: 文件描述符（整数）或流对象
- `callback`: 回调函数，接收 `fd` 作为参数

```php
$registry.on_read($socket_fd, fn($fd) {
    $data = net_recv($fd, 1024)
    println "Received: #{$data}"
})
```

#### $registry.on_write(fd, callback)
注册文件描述符可写事件。

**参数**：
- `fd`: 文件描述符或流对象
- `callback`: 回调函数，接收 `fd` 作为参数

```php
$registry.on_write($socket_fd, fn($fd) {
    net_send($fd, "Hello")
})
```

#### $registry.on_accept(server, callback)
注册服务器 socket 接受连接事件。

**参数**：
- `server`: 服务器 socket 对象（由 `net_tcp_listen` 创建）
- `callback`: 回调函数，接收客户端 socket 作为参数

**回调签名**: `fn($client) { ... }`

```jpl
$server = net_tcp_listen("0.0.0.0", 8080)

$registry.on_accept($server, fn($client) {
    $peer = net_getpeername($client)
    println "New connection from: #{$peer.ip}:#{$peer.port}"
    
    // 为客户端注册读事件
    $registry.on_read($client, fn($socket, $data) {
        println "Received: #{$data}"
        net_send($socket, "Echo: " .. $data)
    })
})
```

#### $registry.on_read(socket, callback)
注册读事件。

**参数**：
- `socket`: socket 对象
- `callback`: 回调函数

**回调签名**: `fn($socket, $data) { ... }`

```jpl
$registry.on_read($client, fn($socket, $data) {
    if (empty($data)) {
        net_close($socket)
        return
    }
    process($data)
})
```

#### $registry.on_write(socket, callback)
注册写事件。

**参数**：
- `socket`: socket 对象
- `callback`: 回调函数

**回调签名**: `fn($socket) { ... }`

#### $registry.off(source)
注销某 source 的所有事件。

**参数**：
- `source`: 事件源（socket、fd 等）

```jpl
$registry.off($client)  // 注销该客户端的所有事件
```

#### $registry.off_timer(timer_id)
注销定时器。

#### $registry.off_signal(sig_num)
注销信号处理。

#### ev_clear($registry)
清空注册表中所有事件。

**参数**：
- `registry`: EvRegistry 对象

```php
ev_clear($registry)
```

#### ev_count($registry)
获取注册表中注册的事件数量。

**参数**：
- `registry`: EvRegistry 对象

**返回值**：`int` - 事件数量

```php
$count = ev_count($registry)
println "Registered events: #{$count}"
```

## 使用示例

### 示例 1：基础定时器

```php
$registry = ev_registry_new()

// 计数器
$counter = 0

// 每 1 秒触发
$registry.on_timer(1000000, fn() {
    $counter = $counter + 1
    println "Counter: #{$counter}"
    
    if ($counter >= 10) {
        println "Stopping..."
        ev_stop($loop)
    }
})

$loop = ev_loop_new()
ev_attach($loop, $registry)
ev_run($loop)
```

### 示例 2：TCP Echo 服务器

```php
// 创建监听 socket
$server = net_tcp_listen("0.0.0.0", 8080)
println "Server on port 8080"

// 创建注册表
$registry = ev_registry_new()

// 处理新连接
$registry.on_accept($server, fn($client) {
    $peer = net_getpeername($client)
    println "Connect from: #{$peer.ip}:#{$peer.port}"
    
    // 处理客户端数据
    $registry.on_read($client, fn($fd) {
        $data = net_recv($fd, 1024)
        
        if (empty($data)) {
            println "Disconnect"
            net_close($fd)
            $registry.off($fd)
            return
        }
        
        // 回显
        net_send($fd, "Echo: " .. $data)
    })
})

// Ctrl+C 退出
$registry.on_signal(2, fn() {
    println "Shutdown..."
    net_close($server)
    ev_stop($loop)
})

// 运行
$loop = ev_loop_new()
ev_attach($loop, $registry)
ev_run($loop)
```

### 示例 3：带超时的网络请求

```php
$registry = ev_registry_new()

// 连接服务器
$client = net_tcp_connect("example.com", 80)

// 发送请求
net_send($client, "GET / HTTP/1.0\r\n\r\n")

$response = ""

// 处理响应
$registry.on_read($client, fn($fd) {
    $data = net_recv($fd, 4096)
    if (empty($data)) {
        // 连接关闭，停止循环
        ev_stop($loop)
        return
    }
    $response = $response .. $data
})

// 5 秒超时
$registry.on_timer_once(5000000, fn() {
    println "Timeout!"
    net_close($client)
    ev_stop($loop)
})

// 错误处理
$registry.on_error($client, fn($fd, $err) {
    println "Error: #{$err}"
    ev_stop($loop)
})

// 运行
$loop = ev_loop_new()
ev_attach($loop, $registry)
ev_run($loop)

println "Response length: " .. strlen($response)
```

### 示例 4：多注册表管理

```php
// 为不同服务创建独立注册表
$http_registry = ev_registry_new()
$ws_registry = ev_registry_new()

// HTTP 服务
$http_server = net_tcp_listen("0.0.0.0", 8080)
$http_registry.on_accept($http_server, fn($client) {
    // HTTP 处理...
})

// WebSocket 服务
$ws_server = net_tcp_listen("0.0.0.0", 8081)
$ws_registry.on_accept($ws_server, fn($client) {
    // WebSocket 处理...
})

// 统一事件循环
$loop = ev_loop_new()
ev_attach($loop, $http_registry)
ev_attach($loop, $ws_registry)

ev_run($loop)
```

## 最佳实践

### 1. 资源清理

始终确保在连接关闭时注销事件：

```php
$registry.on_read($fd, fn($fd) {
    $data = net_recv($fd, 1024)
    
    if (empty($data)) {
        // 连接关闭，必须注销事件！
        net_close($fd)
        $registry.off($fd)  // 重要！
        return
    }
    
    // 处理数据...
})
```

### 2. 错误处理

为所有 socket 注册错误处理器：

```php
$registry.on_error($fd, fn($fd, $err) {
    println "Socket error: #{$err}"
    net_close($fd)
    $registry.off($fd)
})
```

### 3. 信号处理

始终处理 SIGINT 以支持优雅退出：

```php
$registry.on_signal(2, fn() {
    println "\nShutting down gracefully..."
    
    // 清理所有连接
    net_close($server)
    foreach ($clients as $fd) {
        net_close($fd)
    }
    
    ev_stop($loop)
})
```

### 4. 定时器管理

保存定时器 ID 以便后续管理：

```php
// 创建定时器并保存 ID
$timer_id = $registry.on_timer(1000000, fn() {
    println "Periodic task"
})

// 后续可以注销
$registry.off_timer($timer_id)
```

### 5. 性能优化

- **避免在回调中执行耗时操作** - 事件循环是单线程的
- **使用非阻塞 IO** - 配合事件循环使用非阻塞 socket
- **批量处理** - 在可读事件中批量读取数据
- **合理设置缓冲区** - 根据协议选择合适的读取大小

## 高级主题

### 微秒级定时

所有时间参数都是微秒（us）：

| 时间 | 微秒值 |
|------|--------|
| 1 毫秒 | 1000 |
| 100 毫秒 | 100000 |
| 1 秒 | 1000000 |
| 5 秒 | 5000000 |
| 1 分钟 | 60000000 |

### 跨平台支持

Event Loop 自动适配底层操作系统：
- **Linux**: 使用 epoll
- **macOS/BSD**: 使用 kqueue
- **Windows**: 使用 select（降级方案）

### 事件优先级

事件处理优先级（从高到低）：
1. 信号事件
2. 定时器事件
3. 文件描述符事件（读/写/错误）
4. 接受连接事件

### 调试技巧

```php
// 监控事件数量
$last_count = 0
$registry.on_timer(5000000, fn() {
    $count = ev_count($registry)
    if ($count != $last_count) {
        println "Event count changed: #{$last_count} -> #{$count}"
        $last_count = $count
    }
})

// 检查循环状态
if (!ev_is_running($loop)) {
    println "Loop stopped unexpectedly!"
}
```

## 相关文档

- [TCP Echo 服务器示例](../examples/network/tcp-echo-server.jpl)
- [Unix Socket 服务器示例](../examples/network/unix-echo-server.jpl)
- [Event Loop 教程](../examples/network/event-loop-tutorial.jpl)
- [API 参考文档](./API.md)

## 常见问题

**Q: Event Loop 是线程安全的吗？**
A: Event Loop 是单线程的。所有回调都在同一线程中顺序执行，不需要额外同步。

**Q: 可以在回调中修改注册表吗？**
A: 可以，可以在回调中注册或注销事件（包括注销当前事件）。

**Q: 定时器精度如何？**
A: 微秒级，但实际精度取决于操作系统调度。通常可以达到毫秒级精度。

**Q: 如何同时监听多个端口？**
A: 为每个端口创建独立的注册表，或使用同一个注册表注册多个 on_accept 事件。

**Q: 支持多少并发连接？**
A: 取决于系统限制（ulimit -n），理论上可以支持万级连接。
