# JPL 示例代码

本目录包含 JPL 脚本语言的各种使用示例，从基础语法到高级网络编程。

## 🚀 快速开始

### 基础示例

```bash
# 基础语法
jpl run basic/hello.jpl
jpl run basic/datatypes.jpl
jpl run basic/functions.jpl
jpl run basic/arrays.jpl

# 字符串处理
jpl run basic/multiline.jpl
jpl run basic/interpolation.jpl
```

### ⭐⭐ Event Loop 网络编程（Phase 9）

```bash
# 1. 学习 Event Loop（推荐首先运行）
jpl run network/event-loop-tutorial.jpl

# 2. TCP Echo 服务器（终端 1）
jpl run network/tcp-echo-server.jpl

# 3. TCP 客户端（终端 2）
jpl run network/tcp-client.jpl

# 4. Unix Domain Socket
jpl run network/unix-echo-server.jpl

# 5. UDP 通信
jpl run network/udp-client-server.jpl

# 6. DNS 解析
jpl run network/dns-lookup.jpl

# 7. 二进制协议
jpl run network/binary-protocol.jpl
```

## 📚 详细说明

### 1. 基础语法示例

| 文件 | 说明 |
|------|------|
| `hello.jpl` | Hello World + 字符串插值 `#{}` |
| `datatypes.jpl` | 数据类型（null/bool/int/float/string/array/object） |
| `operators.jpl` | 运算符（含 `..` 字符串连接） |
| `control-flow.jpl` | 控制流（if/else/for/while/foreach/exit/die） |
| `functions.jpl` | 函数定义、箭头函数、闭包、递归 |
| `arrays.jpl` | 数组操作、函数式编程 |
| `objects.jpl` | 对象操作、属性访问 |
| `strings.jpl` | 字符串操作（trim/split/substr 等） |
| `file-io.jpl` | 文件读写操作 |

### 2. Phase 10: 语法增强

| 文件 | 说明 |
|------|------|
| `multiline.jpl` | 多行字符串（三引号 `'''...'''` 和 `"""..."""`） |
| `interpolation.jpl` | 字符串插值（Ruby 风格 `#{}`） |
| `json-config.jpl` | JSON 配置解析 |
| `templates.jpl` | 模板字符串应用 |

### 3. Phase 9: 网络编程 ⭐⭐

#### Event Loop 框架

Event Loop 是 JPL 网络编程的核心，基于 IO 多路复用（epoll/kqueue）：

```php
// 基本结构
$registry = ev_registry_new()

// 注册事件
$registry.on_timer(1000000, fn() {
    println "Every second"
})

$registry.on_signal(2, fn() {
    ev_stop($loop)  // Ctrl+C 退出
})

// 创建并运行循环
$loop = ev_loop_new()
ev_attach($loop, $registry)
ev_run($loop)
```

**核心 API**：
- `ev_registry_new()` - 创建事件注册表
- `ev_loop_new()` - 创建事件循环
- `ev_attach(loop, registry)` - 附加注册表
- `ev_run(loop)` / `ev_stop(loop)` - 运行/停止循环
- `$registry.on_timer(us, fn)` - 定时器
- `$registry.on_signal(sig, fn)` - 信号处理
- `$registry.on_read(fd, fn)` - 可读事件
- `$registry.on_write(fd, fn)` - 可写事件
- `$registry.on_accept(server, fn)` - 连接事件

#### 网络编程示例

| 文件 | 说明 | 关键技术 |
|------|------|----------|
| `event-loop-tutorial.jpl` | **Event Loop 完整教程** | 定时器、信号、基础 API |
| `tcp-echo-server.jpl` | TCP Echo 服务器 | accept、read、write、timer、signal |
| `tcp-client.jpl` | TCP 客户端 | connect、send、recv |
| `unix-echo-server.jpl` | Unix Domain Socket | Unix socket、性能优化 |
| `udp-client-server.jpl` | UDP 通信 | bind、sendto、recvfrom |
| `dns-lookup.jpl` | DNS 解析 | dns_resolve、dns_resolve_one |
| `binary-protocol.jpl` | 二进制协议 | pack、unpack、Buffer |

#### 二进制处理

支持 pack/unpack 格式：
- `C` - 1 字节无符号
- `S/s` - 2 字节（大端/小端）
- `N/V` - 4 字节（大端/小端）
- `Q/q` - 8 字节（大端/小端）
- `f/d` - float/double
- `a/Z` - 字符串（空填充/零结尾）

Buffer 对象支持：
- `buffer_new([endian])` - 创建缓冲区（可选字节序）
- `buffer_new_from(bytes, [endian])` - 从字节数组/字符串创建
- `buffer_write_int8/16/32` - 写入有符号整数
- `buffer_write_uint8/16/32` - 写入无符号整数
- `buffer_read_int8/16/32` - 读取有符号整数
- `buffer_read_uint8/16/32` - 读取无符号整数
- `buffer_seek/tell/length` - 位置操作
- `buffer_to_bytes/to_string` - 转换
- `is_buffer(value)` - 类型检查

### 4. 错误处理

| 文件 | 说明 |
|------|------|
| `exceptions.jpl` | try/catch、throw、error_last |

### 5. 模块系统

| 文件 | 说明 |
|------|------|
| `math.jpl` | 数学函数库 |
| `strings.jpl` | 字符串工具库 |

## 🎯 学习路径建议

### 新手入门

```bash
# 第 1 步：基础语法
jpl run basic/hello.jpl
jpl run basic/datatypes.jpl
jpl run basic/operators.jpl

# 第 2 步：控制流和函数
jpl run basic/control-flow.jpl
jpl run basic/functions.jpl

# 第 3 步：数组和对象
jpl run basic/arrays.jpl
jpl run basic/objects.jpl

# 第 4 步：字符串处理
jpl run basic/strings.jpl
jpl run basic/multiline.jpl
jpl run basic/interpolation.jpl
```

### 进阶网络编程

```bash
# 第 1 步：学习 Event Loop 基础
jpl run network/event-loop-tutorial.jpl

# 第 2 步：理解二进制协议
jpl run network/binary-protocol.jpl

# 第 3 步：运行 TCP 服务器和客户端
# 终端 1:
jpl run network/tcp-echo-server.jpl

# 终端 2:
jpl run network/tcp-client.jpl

# 第 4 步：探索 UDP 和 DNS
jpl run network/udp-client-server.jpl
jpl run network/dns-lookup.jpl
```

## 📖 相关文档

- [Event Loop 详细指南](../docs/EVENT_LOOP_GUIDE.md) - 完整的 Event Loop API 文档
- [API 参考文档](../API.md) - 所有内置函数列表
- [设计文档](../docs/DESIGN.md) - D24 网络框架设计决策
- [项目计划](../PLAN.md) - Phase 9 网络编程开发计划

## 💡 实用技巧

### 交互式探索

```bash
# 启动 REPL
jpl repl

# 在 REPL 中测试网络 API
> $ips = dns_resolve("localhost")
> println $ips

# 测试 Event Loop
> $r = ev_registry_new()
> $r.on_timer(1000000, fn() { println "tick" })
> $l = ev_loop_new()
> ev_attach($l, $r)
> ev_run($l)
```

### 调试脚本

```bash
# 使用 debug 模式运行
jpl run --debug network/tcp-echo-server.jpl

# 检查语法
jpl check network/tcp-echo-server.jpl
```

## 🔗 外部资源

- [Go net 包文档](https://pkg.go.dev/net) - 底层网络实现
- [epoll 介绍](https://man7.org/linux/man-pages/man7/epoll.7.html) - Linux IO 多路复用
- [kqueue 介绍](https://man.freebsd.org/kqueue/) - BSD/macOS IO 多路复用

```

## 📚 示例说明

### 基础示例 (basic/)

适合 JPL 初学者，展示语言的核心特性。

| 文件 | 内容 |
|------|------|
| `hello.jpl` | Hello World + 字符串插值 `#{}` |
| `datatypes.jpl` | 所有数据类型（含 BigInt/BigDecimal） |
| `operators.jpl` | 运算符（`..` 字符串连接、`+` 等） |
| `control-flow.jpl` | if/else、循环、break/continue/exit/die |
| `functions.jpl` | 传统函数、箭头函数 `->`、闭包、递归 |
| `arrays.jpl` | 数组操作、函数式方法（map/filter/reduce） |
| `objects.jpl` | 对象创建、属性访问、遍历 |
| `strings.jpl` | 字符串操作、多行字符串、插值 |
| `file-io.jpl` | 文件读写、流操作、目录管理、JSON处理 |
| `type_cast.jpl` | ⭐ Go 风格类型转换 `int(x)`, `float(x)`, `string(x)`, `bool(x)` |
| `parse_object.jpl` | ⭐ 安全对象解析，对比 `eval()` 的安全性 |

### I/O 函数示例 (io/) ⭐ 新增

展示新的输出函数特性。

| 文件 | 内容 | 特性 |
|------|------|------|
| `puts_and_pp.jpl` | puts 和 pp 输出函数 | puts（纯文本输出）、pp（Pretty Print 格式化） |

### 网络编程示例 (network/) ⭐ 新增！

展示 Phase 9 网络编程能力。

| 文件 | 内容 | 特性 |
|------|------|------|
| `tcp-echo-server.jpl` | TCP Echo 服务器 | 事件循环、非阻塞 IO、信号处理 |
| `tcp-client.jpl` | TCP 客户端 | DNS 解析、连接、收发 |
| `unix-echo-server.jpl` | Unix Domain Socket | 本地 IPC、socket 文件管理 |
| `udp-client-server.jpl` | UDP 通信 | 无连接协议、sendto/recvfrom |
| `dns-lookup.jpl` | DNS 解析 | resolve/resolve_v4/resolve_v6/get_records |
| `binary-protocol.jpl` | 二进制协议 | pack/unpack、Buffer、字节序 |

**网络编程特色功能**：
- **事件循环**：`ev_loop_new`、`ev_registry_new`、`ev_run`
- **异步 IO**：`on_read`、`on_write`、`on_accept`
- **定时器**：微秒级精度 `on_timer`、`on_timer_once`
- **信号处理**：POSIX 信号 `on_signal`
- **DNS 解析**：A/AAAA/CNAME/MX/NS/TXT 记录
- **二进制处理**：pack/unpack、Buffer 对象、字节序控制

### Phase 10 语法增强 (phase10/)

展示字符串增强特性。

- **多行字符串**：Python 风格三引号 `'''` 和 `"""`
- **字符串插值**：Ruby 风格 `#{}`，支持表达式、对象访问、数组索引
- **JSON 支持**：多行字符串解析 JSON 配置

### 其他示例

- **functional/**: 高阶函数、闭包、柯里化
- **modules/**: 可复用模块创建
- **error-handling/**: try/catch/throw 异常处理
- **integration/**: Go 程序嵌入 JPL
- **package-manager/**: 包管理器使用示例（jpl.json、依赖管理）
- **tasks/**: 任务系统示例（定义任务、依赖关系）

## 🎯 学习路径建议

### 初学者

1. **Hello World** (`basic/hello.jpl`)
2. **数据类型** (`basic/datatypes.jpl`)
3. **运算符** (`basic/operators.jpl`)
4. **控制流** (`basic/control-flow.jpl`)
5. **函数** (`basic/functions.jpl`)
6. **类型转换** (`basic/type_cast.jpl`) ⭐ 新增

### 进阶学习

7. **数组** (`basic/arrays.jpl`)
8. **对象** (`basic/objects.jpl`)
9. **安全解析** (`basic/parse_object.jpl`) ⭐ 新增
10. **文件 I/O** (`basic/file-io.jpl`)
11. **I/O 函数** (`io/puts_and_pp.jpl`) ⭐ 新增
12. **字符串增强** (`phase10/`)
13. **错误处理** (`error-handling/exceptions.jpl`)

### 网络编程（新！）

11. **DNS 解析** (`network/dns-lookup.jpl`)
12. **TCP 客户端** (`network/tcp-client.jpl`)
13. **TCP 服务器** (`network/tcp-echo-server.jpl`)
14. **UDP 通信** (`network/udp-client-server.jpl`)
15. **二进制协议** (`network/binary-protocol.jpl`)

### 高级特性

16. **函数式编程** (`functional/functional.jpl`)
17. **模块开发** (`modules/`)
18. **Go 集成** (`integration/`)

## 💡 实用技巧

### 调试示例

```bash
# 使用 --debug 查看字节码
jpl run --debug basic/functions.jpl

# 使用 --verbose 查看执行过程
jpl run --verbose network/tcp-echo-server.jpl
```

### 测试网络示例

```bash
# 启动服务器（后台）
jpl run network/tcp-echo-server.jpl &

# 运行客户端
jpl run network/tcp-client.jpl

# 停止服务器
kill %1
```

### 性能测试

```bash
# DNS 解析性能
time jpl run network/dns-lookup.jpl

# 带性能分析的运行
jpl run --profile network/binary-protocol.jpl
```

## 📝 注意事项

1. **网络示例需要权限**：某些端口（如 80）需要管理员权限
2. **Unix Socket 清理**：示例会自动清理旧的 socket 文件
3. **信号处理**：按 Ctrl+C 可以优雅地关闭服务器示例
4. **DNS 缓存**：本地 DNS 可能有缓存，测试结果可能不同

## 🔗 相关文档

- [README.md](../README.md) - 项目介绍
- [API.md](../API.md) - API 参考
- [DESIGN.md](../docs/DESIGN.md) - 设计决策
- [CHANGELOG.md](../CHANGELOG.md) - 更新日志
