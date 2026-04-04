
# 标准库函数签名

本文档列出了已实现的标准库的函数签名和功能简介。

---

## 正则表达式

> 模块导入: `import "preg"`
> 语法: Go RE2 (https://github.com/google/re2/wiki/Syntax)
> 模块导入: `import "re"`

### 匹配与查找

```go
// re_match 检查正则是否匹配
// re_match(pattern, string) → bool
// 示例: re_match(`\d+`, "abc123") → true
func builtinReMatch(ctx *Context, args []Value) (Value, error)

// re_search 查找第一个匹配
// re_search(pattern, string) → string | ""
// 示例: re_search(`\d+`, "Room 101") → "101"
func builtinReSearch(ctx *Context, args []Value) (Value, error)

// re_findall 查找所有匹配
// re_findall(pattern, string) → [match1, match2, ...]
// 示例: re_findall(`\d+`, "Room 101, Floor 5") → ["101", "5"]
func builtinReFindall(ctx *Context, args []Value) (Value, error)

// re_groups 返回捕获组（支持命名捕获）
// re_groups(pattern, string) → {0: full, 1: group1, name: named_group} | null
// 示例: re_groups(`(?P<year>\d{4})-(?P<month>\d{2})`, "2024-03")
//       → {0: "2024-03", 1: "2024", 2: "03", year: "2024", month: "03"}
func builtinReGroups(ctx *Context, args []Value) (Value, error)
```

### 替换与分割

```go
// re_sub 替换所有匹配
// re_sub(pattern, replacement, string) → string
// 示例: re_sub(`\d+`, "[NUM]", "Room 101") → "Room [NUM]"
func builtinReSub(ctx *Context, args []Value) (Value, error)

// re_split 按正则分割字符串
// re_split(pattern, string) → [part1, part2, ...]
// 示例: re_split(`\s*,\s*`, "a, b ,c") → ["a", "b", "c"]
func builtinReSplit(ctx *Context, args []Value) (Value, error)
```

---

## 进程 API

> 模块导入: `import "process"`

### 命令执行

```go
// exec 执行系统命令，返回输出字符串
// exec($cmd) → string | null
// exec($cmd, $args) → string | null
func builtinExec(ctx *Context, args []Value) (Value, error)

// system 执行系统命令，返回退出码
// system($cmd) → int
// 返回 0 表示成功
func builtinSystem(ctx *Context, args []Value) (Value, error)

// shell_exec 执行 shell 命令，返回完整输出
// shell_exec($cmd) → string | null
func builtinShellExec(ctx *Context, args []Value) (Value, error)
```

### 环境变量

```go
// getenv 获取环境变量
// getenv($name) → string | null
// getenv($name, $default) → string
func builtinGetenv(ctx *Context, args []Value) (Value, error)

// setenv 设置环境变量
// setenv($name, $value) → bool
func builtinSetenv(ctx *Context, args []Value) (Value, error)

// putenv 设置环境变量（KEY=VALUE 格式）
// putenv($expr) → bool
func builtinPutenv(ctx *Context, args []Value) (Value, error)
```

### 进程信息

```go
// getpid 获取当前进程 ID
// getpid() → int

// getppid 获取父进程 ID
// getppid() → int
func builtinGetppid(ctx *Context, args []Value) (Value, error)

// getlogin 获取登录用户名
// getlogin() → string | null
func builtinGetlogin(ctx *Context, args []Value) (Value, error)

// hostname 获取主机名
// hostname() → string
func builtinHostname(ctx *Context, args []Value) (Value, error)

// tmpdir 获取系统临时目录
// tmpdir() → string
func builtinTmpdir(ctx *Context, args []Value) (Value, error)
```

### 进程管道

```go
// proc_open 创建进程管道
// proc_open($cmd, $opts) → Process | null
// opts: {stdin: "pipe"|"null", stdout: "pipe"|"null", stderr: "pipe"|"null"}
func builtinProcOpen(ctx *Context, args []Value) (Value, error)

// proc_close 关闭进程管道
// proc_close($proc) → int
// 返回退出码
func builtinProcClose(ctx *Context, args []Value) (Value, error)

// proc_wait 等待进程结束
// proc_wait($proc) → int
// 返回退出码
func builtinProcWait(ctx *Context, args []Value) (Value, error)

// proc_status 获取进程状态
// proc_status($proc) → {pid, running, exited, exit_code}
func builtinProcStatus(ctx *Context, args []Value) (Value, error)
```

### 进程控制

```go
// spawn 创建子进程（不等待完成）
// spawn($cmd) → Process | null
// spawn($cmd, $args) → Process | null
func builtinSpawn(ctx *Context, args []Value) (Value, error)

// kill 向进程发送信号
// kill($pid) → bool
// kill($pid, $signal) → bool
// 常用信号: 1=SIGHUP, 2=SIGINT, 9=SIGKILL, 15=SIGTERM
func builtinKill(ctx *Context, args []Value) (Value, error)

// waitpid 等待指定子进程
// waitpid($proc) → int
// waitpid($pid) → int
func builtinWaitpid(ctx *Context, args []Value) (Value, error)

// fork 创建子进程（Unix）
// fork() → int
// 返回 0 表示子进程，正数表示父进程（子进程PID）
func builtinFork(ctx *Context, args []Value) (Value, error)

// pipe 创建管道对
// pipe() → {read: fd, write: fd}
func builtinPipe(ctx *Context, args []Value) (Value, error)

// sigwait 阻塞等待信号
// sigwait($signal) → int
// sigwait([$signals]) → int
func builtinSigwait(ctx *Context, args []Value) (Value, error)

// usleep 微秒级暂停
// usleep($microseconds) → null
func builtinUsleep(ctx *Context, args []Value) (Value, error)
```

### 脚本示例

```jpl
import "process"

// 执行命令
$output = process.exec("ls -la")
$code = process.system("ping -c 1 8.8.8.8")

// 环境变量
process.setenv("APP_ENV", "production")
$home = process.getenv("HOME")

// 进程管理
$proc = process.spawn("sleep", ["5"])
process.kill($proc.pid, 9)  // SIGKILL
$code = process.waitpid($proc)

// 进程信息
$pid = process.getpid()
$ppid = process.getppid()
$host = process.hostname()
```

---

## 二进制处理 API

### pack/unpack 函数

```go
// pack 打包二进制数据
// pack(format, val1, val2, ...) → bytes_array
// 格式字符: C(1B), S/s(2B), N/V(4B), Q/q(8B), f(4B), d(8B), a/Z(字符串), x(填充)
// 大端: S/N/Q, 小端: s/V/q
func builtinPack(ctx *Context, args []Value) (Value, error)

// unpack 解包二进制数据
// unpack(format, bytes_array) → value or [val1, val2, ...]
func builtinUnpack(ctx *Context, args []Value) (Value, error)
```

### Buffer 对象

```go
// BufferValue 二进制缓冲区对象
type BufferValue struct {
    data  *bytes.Buffer
    order binary.ByteOrder  // 默认大端
}

// NewBuffer 创建 Buffer（默认大端）
func NewBuffer(endian string) *BufferValue

// buffer_new([endian]) → Buffer
// endian: "big" 或 "little"
func builtinBufferNew(ctx *Context, args []Value) (Value, error)

// buffer_new_from(bytes, [endian]) → Buffer
// bytes: 字节数组或字符串，endian: 可选字节序
func builtinBufferNewFrom(ctx *Context, args []Value) (Value, error)

// buffer_set_endian(buf, endian) → bool
func builtinBufferSetEndian(ctx *Context, args []Value) (Value, error)

// 写入函数 - 有符号整数（使用当前字节序）
func builtinBufferWriteInt8(ctx *Context, args []Value) (Value, error)
func builtinBufferWriteInt16(ctx *Context, args []Value) (Value, error)
func builtinBufferWriteInt32(ctx *Context, args []Value) (Value, error)

// 写入函数 - 无符号整数（使用当前字节序）
func builtinBufferWriteUint8(ctx *Context, args []Value) (Value, error)
func builtinBufferWriteUint16(ctx *Context, args []Value) (Value, error)
func builtinBufferWriteUint32(ctx *Context, args []Value) (Value, error)
func builtinBufferWriteFloat32(ctx *Context, args []Value) (Value, error)
func builtinBufferWriteFloat64(ctx *Context, args []Value) (Value, error)
func builtinBufferWriteString(ctx *Context, args []Value) (Value, error)
func builtinBufferWriteBytes(ctx *Context, args []Value) (Value, error)

// 读取函数 - 有符号整数（使用当前字节序）
func builtinBufferReadInt8(ctx *Context, args []Value) (Value, error)
func builtinBufferReadInt16(ctx *Context, args []Value) (Value, error)
func builtinBufferReadInt32(ctx *Context, args []Value) (Value, error)

// 读取函数 - 无符号整数（使用当前字节序）
func builtinBufferReadUint8(ctx *Context, args []Value) (Value, error)
func builtinBufferReadUint16(ctx *Context, args []Value) (Value, error)
func builtinBufferReadUint32(ctx *Context, args []Value) (Value, error)
func builtinBufferReadFloat32(ctx *Context, args []Value) (Value, error)
func builtinBufferReadFloat64(ctx *Context, args []Value) (Value, error)
func builtinBufferReadString(ctx *Context, args []Value) (Value, error)
func builtinBufferReadBytes(ctx *Context, args []Value) (Value, error)

// 游标操作
func builtinBufferSeek(ctx *Context, args []Value) (Value, error)
func builtinBufferTell(ctx *Context, args []Value) (Value, error)
func builtinBufferLength(ctx *Context, args []Value) (Value, error)
func builtinBufferReset(ctx *Context, args []Value) (Value, error)

// 转换函数
func builtinBufferToBytes(ctx *Context, args []Value) (Value, error)
func builtinBufferToString(ctx *Context, args []Value) (Value, error)

// 类型检查
func builtinIsBuffer(ctx *Context, args []Value) (Value, error)
```

---

## 加密模块 API

> 模块导入: `import "crypto"`

### Hash 函数

```go
// sha256 计算 SHA-256 哈希
// sha256(data) → hex_string (64 chars)
// 示例: sha256("Hello") → "185f8db32271fe25f561a6fc938b2e264306ec304eda518007d1764826381969"
func builtinSHA256(ctx *Context, args []Value) (Value, error)

// sha512 计算 SHA-512 哈希
// sha512(data) → hex_string (128 chars)
func builtinSHA512(ctx *Context, args []Value) (Value, error)
```

### HMAC 函数

```go
// hmac_sha256 计算 HMAC-SHA256
// hmac_sha256(key, data) → hex_string (64 chars)
func builtinHMACSHA256(ctx *Context, args []Value) (Value, error)

// hmac_sha512 计算 HMAC-SHA512
// hmac_sha512(key, data) → hex_string (128 chars)
func builtinHMACSHA512(ctx *Context, args []Value) (Value, error)
```

### 编码函数

```go
// hex_encode Hex 编码
// hex_encode(data) → hex_string
// 示例: hex_encode("Hello") → "48656c6c6f"
func builtinHexEncode(ctx *Context, args []Value) (Value, error)

// hex_decode Hex 解码
// hex_decode(data) → string | null
// 示例: hex_decode("48656c6c6f") → "Hello"
func builtinHexDecode(ctx *Context, args []Value) (Value, error)

// 注意: base64_encode/decode 在 hash 模块中定义，crypto 模块重新导出
```

### AES 加密

```go
// aes_encrypt AES-256-GCM 加密
// aes_encrypt(data, key) → base64_string
// key: 32 字节 hex 字符串 (64 hex chars)
// 返回值包含 nonce(12) + ciphertext + tag(16)
// 示例:
//   $key = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
//   $encrypted = aes_encrypt("Secret", $key)
func builtinAESEncrypt(ctx *Context, args []Value) (Value, error)

// aes_decrypt AES-256-GCM 解密
// aes_decrypt(data, key) → string | null
// data: base64 编码的加密数据
// 解密失败返回 null
// 示例:
//   $decrypted = aes_decrypt($encrypted, $key)
func builtinAESDecrypt(ctx *Context, args []Value) (Value, error)
```

### bcrypt 密码哈希

```go
// bcrypt_hash 生成 bcrypt 哈希
// bcrypt_hash(password) → hash_string
// 示例:
//   $hash = bcrypt_hash("mypassword")
func builtinBcryptHash(ctx *Context, args []Value) (Value, error)

// bcrypt_verify 验证密码
// bcrypt_verify(password, hash) → bool
// 示例:
//   $valid = bcrypt_verify("mypassword", $hash)
func builtinBcryptVerify(ctx *Context, args []Value) (Value, error)

// bcrypt_cost 获取哈希 cost 值
// bcrypt_cost(hash) → int
func builtinBcryptCost(ctx *Context, args []Value) (Value, error)
```

### ECC 椭圆曲线（Ed25519 + X25519）

```go
// ed25519_generate_key 生成 Ed25519 密钥对
// ed25519_generate_key() → {public_key, private_key}
func builtinEd25519GenerateKey(ctx *Context, args []Value) (Value, error)

// ed25519_sign Ed25519 签名
// ed25519_sign(message, private_key) → signature
func builtinEd25519Sign(ctx *Context, args []Value) (Value, error)

// ed25519_verify Ed25519 验签
// ed25519_verify(message, signature, public_key) → bool
func builtinEd25519Verify(ctx *Context, args []Value) (Value, error)

// ed25519_public_key 获取 Ed25519 公钥
// ed25519_public_key(private_key) → public_key
func builtinEd25519PublicKey(ctx *Context, args []Value) (Value, error)

// x25519_generate_key 生成 X25519 密钥对
// x25519_generate_key() → {public_key, private_key}
func builtinX25519GenerateKey(ctx *Context, args []Value) (Value, error)

// x25519_shared_secret 计算共享密钥
// x25519_shared_secret(private_key, public_key) → shared_secret
func builtinX25519SharedSecret(ctx *Context, args []Value) (Value, error)

// x25519_public_key 获取 X25519 公钥
// x25519_public_key(private_key) → public_key
func builtinX25519PublicKey(ctx *Context, args []Value) (Value, error)
```

### RSA 加密/签名

```go
// rsa_generate_key 生成 RSA 密钥对
// rsa_generate_key(bits?) → {public_key, private_key, n, e}
// bits: 密钥位数，默认 2048
func builtinRSAGenerateKey(ctx *Context, args []Value) (Value, error)

// rsa_encrypt RSA 加密
// rsa_encrypt(message, public_key) → ciphertext
func builtinRSAEncrypt(ctx *Context, args []Value) (Value, error)

// rsa_decrypt RSA 解密
// rsa_decrypt(ciphertext, private_key) → message
func builtinRSADecrypt(ctx *Context, args []Value) (Value, error)

// rsa_sign RSA 签名
// rsa_sign(message, private_key) → signature
func builtinRSASign(ctx *Context, args []Value) (Value, error)

// rsa_verify RSA 验签
// rsa_verify(message, signature, public_key) → bool
func builtinRSAVerify(ctx *Context, args []Value) (Value, error)

// rsa_public_key 获取 RSA 公钥
// rsa_public_key(private_key) → public_key
func builtinRSAPublicKey(ctx *Context, args []Value) (Value, error)
```

---

## 异步文件 IO API (asyncio)

> 参考 Python asyncio 设计，提供非阻塞的文件操作。
> 模块导入: `import "asyncio"`

### 一次性读写

```go
// file_get_async 异步读取整个文件（文本模式）
// file_get_async($path, fn($data) { ... }) → null
// 回调接收文件内容字符串，失败时为 null
func builtinFileGetAsync(ctx *Context, args []Value) (Value, error)

// file_put_async 异步写入整个文件（文本模式）
// file_put_async($path, $data, fn($success) { ... }) → null
// 回调接收布尔值表示成功/失败
func builtinFilePutAsync(ctx *Context, args []Value) (Value, error)

// file_append_async 异步追加到文件
// file_append_async($path, $data, fn($success) { ... }) → null
func builtinFileAppendAsync(ctx *Context, args []Value) (Value, error)

// file_get_bytes 异步读取二进制文件
// file_get_bytes($path, fn($buffer) { ... }) → null
// 回调接收 Buffer 对象，失败时为 null
func builtinFileGetBytes(ctx *Context, args []Value) (Value, error)

// file_put_bytes 异步写入二进制文件
// file_put_bytes($path, $buffer, fn($success) { ... }) → null
func builtinFilePutBytes(ctx *Context, args []Value) (Value, error)
```

### 流式读取

```go
// file_read_lines 逐行读取文件
// file_read_lines($path, fn($line) { ... }, fn($success) { ... }) → null
// onLine 回调每行调用一次，onDone 完成时调用
func builtinFileReadLines(ctx *Context, args []Value) (Value, error)

// file_read_chunks 分块读取文件
// file_read_chunks($path, $chunkSize, fn($chunk) { ... }, fn($success) { ... }) → null
// chunkSize 默认 4096 字节
func builtinFileReadChunks(ctx *Context, args []Value) (Value, error)
```

### 批量操作

```go
// file_get_batch 批量读取文件
// file_get_batch($paths, fn($results) { ... }) → null
// paths: 文件路径数组，results: 对应的结果数组
func builtinFileGetBatch(ctx *Context, args []Value) (Value, error)

// file_put_batch 批量写入文件
// file_put_batch($items, fn($results) { ... }) → null
// items: [{path: "...", data: "..."}, ...]
func builtinFilePutBatch(ctx *Context, args []Value) (Value, error)

// file_parallel 并行执行多个文件操作（带冲突检测）
// file_parallel($ops, fn($results) { ... }) → null
// ops: [{op: "read", path: "..."}, {op: "write", path: "...", data: "..."}]
// results: [{success: true, data: "..."}, {success: false, error: "..."}, ...]
func builtinFileParallel(ctx *Context, args []Value) (Value, error)
```

### 文件锁

```go
// file_with_lock 获取文件锁并执行操作
// file_with_lock($path, fn($lock) { ... }) → null
// lock.release() 释放锁
func builtinFileWithLock(ctx *Context, args []Value) (Value, error)
```

### 脚本示例

```jpl
import "asyncio"

// 异步读取
asyncio.file_get("data.txt", fn($data) {
    println "Got: #{$data}"
})

// 逐行处理大文件
asyncio.file_read_lines("big.txt", fn($line) {
    process($line)
}, fn() {
    println "Done"
})

// 批量读取
asyncio.file_get_batch(["a.txt", "b.txt"], fn($results) {
    foreach ($r in $results) {
        println $r
    }
})

// 文件锁保护
asyncio.file_with_lock("shared.txt", fn($lock) {
    asyncio.file_get("shared.txt", fn($data) {
        // 安全读写...
        $lock.release()
    })
})
```

---

## 事件循环 API

### 事件循环结构

```go
// EvRegistryValue 事件注册表对象
type EvRegistryValue struct {
    onRead      map[int]func(int)
    onWrite     map[int]func(int)
    onAccept    map[int]func(int)
    timers      map[int]*evTimer
    signals     map[int]func()
}

// EvLoopValue 事件循环对象
type EvLoopValue struct {
    poll       *EvPoll
    registries []*EvRegistryValue
    running    bool
    stopChan   chan bool
}

// 创建事件循环
// ev_loop_new() → loop
func builtinEvLoopNew(ctx *Context, args []Value) (Value, error)

// 创建注册表
// ev_registry_new() → registry
func builtinEvRegistryNew(ctx *Context, args []Value) (Value, error)

// 附加注册表到循环
// ev_attach(loop, registry) → bool
func builtinEvAttach(ctx *Context, args []Value) (Value, error)

// 运行事件循环（阻塞）
// ev_run(loop) → null
func builtinEvRun(ctx *Context, args []Value) (Value, error)

// 运行一次（非阻塞）
// ev_run_once(loop) → bool
func builtinEvRunOnce(ctx *Context, args []Value) (Value, error)

// 停止事件循环
// ev_stop(loop) → bool
func builtinEvStop(ctx *Context, args []Value) (Value, error)

// 检查是否运行中
// ev_is_running(loop) → bool
func builtinEvIsRunning(ctx *Context, args []Value) (Value, error)

// 获取当前时间（微秒级时间戳）
// ev_timer_now() → int
func builtinEvTimerNow(ctx *Context, args []Value) (Value, error)
```

### 注册表事件注册

```go
// ev_on_read 注册读事件处理器
// registry.on_read(fd, fn(fd) { ... }) → bool
func builtinRegistryOnRead(ctx *Context, args []Value) (Value, error)

// ev_on_write 注册写事件处理器
// registry.on_write(fd, fn(fd) { ... }) → bool
func builtinRegistryOnWrite(ctx *Context, args []Value) (Value, error)

// ev_on_accept 注册连接事件处理器
// registry.on_accept(server_fd, fn(client_fd) { ... }) → bool
func builtinRegistryOnAccept(ctx *Context, args []Value) (Value, error)

// ev_on_timer 注册周期性定时器（微秒）
// registry.on_timer(microseconds, fn() { ... }) → timer_id
func builtinRegistryOnTimer(ctx *Context, args []Value) (Value, error)

// ev_on_timer_once 注册一次性定时器（微秒）
// registry.on_timer_once(microseconds, fn() { ... }) → timer_id
func builtinRegistryOnTimerOnce(ctx *Context, args []Value) (Value, error)

// ev_on_signal 注册 POSIX 信号处理器
// registry.on_signal(signal_num, fn() { ... }) → bool
// 常用信号: 2=SIGINT(Ctrl+C), 15=SIGTERM
func builtinRegistryOnSignal(ctx *Context, args []Value) (Value, error)

// ev_off 注销 fd 所有事件
// registry.off(fd) → bool
func builtinRegistryOff(ctx *Context, args []Value) (Value, error)

// ev_off_read 只注销读事件
// registry.off_read(fd) → bool
func builtinRegistryOffRead(ctx *Context, args []Value) (Value, error)

// ev_off_write 只注销写事件
// registry.off_write(fd) → bool
func builtinRegistryOffWrite(ctx *Context, args []Value) (Value, error)

// ev_off_timer 注销定时器
// registry.off_timer(timer_id) → bool
func builtinRegistryOffTimer(ctx *Context, args []Value) (Value, error)

// ev_off_signal 注销信号处理器
// registry.off_signal(signal_num) → bool
func builtinRegistryOffSignal(ctx *Context, args []Value) (Value, error)

// ev_clear 清空所有注册
// registry.clear() → bool
func builtinRegistryClear(ctx *Context, args []Value) (Value, error)

// ev_count 返回注册的事件总数
// registry.count() → int
func builtinRegistryCount(ctx *Context, args []Value) (Value, error)
```

---

## 网络编程 API

### TCP/Unix/UDP Socket

```go
// NetSocketValue 表示网络 socket 对象
type NetSocketValue struct {
    fd       int
    isUnix   bool
    isUDP    bool
    conn     net.Conn
    listener net.Listener
    udpConn  *net.UDPConn
}

// Type 返回类型标识（TypeStream）
func (s *NetSocketValue) Type() ValueType

// String 返回字符串表示
func (s *NetSocketValue) String() string
```

### TCP 函数

```go
// net_tcp_listen 创建 TCP 监听 socket
// net_tcp_listen(host, port) → socket
func builtinNetTcpListen(ctx *Context, args []Value) (Value, error)

// net_tcp_connect 建立 TCP 连接
// net_tcp_connect(host, port) → socket
func builtinNetTcpConnect(ctx *Context, args []Value) (Value, error)

// net_tcp_accept 接受 TCP 连接
// net_tcp_accept(server_socket) → client_socket
func builtinNetTcpAccept(ctx *Context, args []Value) (Value, error)
```

### Unix Domain Socket 函数

```go
// net_unix_listen 创建 Unix Domain 监听 socket
// net_unix_listen(path) → socket
func builtinNetUnixListen(ctx *Context, args []Value) (Value, error)

// net_unix_connect 连接 Unix Domain Socket
// net_unix_connect(path) → socket
func builtinNetUnixConnect(ctx *Context, args []Value) (Value, error)

// net_unix_accept 接受 Unix Domain 连接
// net_unix_accept(server_socket) → client_socket
func builtinNetUnixAccept(ctx *Context, args []Value) (Value, error)

// net_is_unix 检查是否为 Unix Domain Socket
// net_is_unix(socket) → bool
func builtinNetIsUnix(ctx *Context, args []Value) (Value, error)
```

### UDP 函数

```go
// net_udp_bind 创建 UDP socket 并绑定地址
// net_udp_bind(host, port) → socket
func builtinNetUdpBind(ctx *Context, args []Value) (Value, error)

// net_udp_sendto 发送 UDP 数据到指定地址
// net_udp_sendto(socket, data, host, port) → bytes_sent
func builtinNetUdpSendto(ctx *Context, args []Value) (Value, error)

// net_udp_recvfrom 接收 UDP 数据
// net_udp_recvfrom(socket, len) → [data, from_ip, from_port]
func builtinNetUdpRecvfrom(ctx *Context, args []Value) (Value, error)
```

### 通用网络函数

```go
// net_send 发送数据（TCP/Unix）
// net_send(socket, data) → bytes_sent
func builtinNetSend(ctx *Context, args []Value) (Value, error)

// net_recv 接收数据（TCP/Unix）
// net_recv(socket, len) → data
func builtinNetRecv(ctx *Context, args []Value) (Value, error)

// net_close 关闭 socket
// net_close(socket) → bool
func builtinNetClose(ctx *Context, args []Value) (Value, error)

// net_getsockname 获取本地地址
// net_getsockname(socket) → {ip, port} or {path}
func builtinNetGetsockname(ctx *Context, args []Value) (Value, error)

// net_getpeername 获取对端地址
// net_getpeername(socket) → {ip, port} or {path}
func builtinNetGetpeername(ctx *Context, args []Value) (Value, error)

// net_set_nonblock 设置非阻塞模式
// net_set_nonblock(socket) → bool
func builtinNetSetNonblock(ctx *Context, args []Value) (Value, error)
```

### DNS 解析函数

```go
// dns_resolve 解析域名返回所有 IP
// dns_resolve(host) → [ip1, ip2, ...]
func builtinDNSResolve(ctx *Context, args []Value) (Value, error)

// dns_resolve_one 解析域名返回单个 IP
// dns_resolve_one(host) → ip
func builtinDNSResolveOne(ctx *Context, args []Value) (Value, error)

// dns_resolve_v4 解析域名返回所有 IPv4
// dns_resolve_v4(host) → [ip1, ...]
func builtinDNSResolveV4(ctx *Context, args []Value) (Value, error)

// dns_resolve_v6 解析域名返回所有 IPv6
// dns_resolve_v6(host) → [ip1, ...]
func builtinDNSResolveV6(ctx *Context, args []Value) (Value, error)

// dns_get_records 获取 DNS 记录详情
// dns_get_records(host [, type]) → [{type, ...}, ...]
// 支持类型: A, AAAA, CNAME, MX, NS, TXT
func builtinDNSGetRecords(ctx *Context, args []Value) (Value, error)
```

### TLS/SSL 加密连接

> 模块导入: `import "tls"`

#### 连接管理

```go
// tls_connect 建立 TLS 客户端连接
// tls_connect(host, port, options?) → TLSSocketValue
// 选项: verify, ca_file, cert_file, key_file, server_name
func builtinTLSConnect(ctx *Context, args []Value) (Value, error)

// tls_listen 创建 TLS 服务端监听
// tls_listen(port, cert_file, key_file, options?) → TLSListenerValue
// 选项: host
func builtinTLSListen(ctx *Context, args []Value) (Value, error)

// tls_accept 接受 TLS 连接
// tls_accept(server_socket) → client_socket
func builtinTLSAccept(ctx *Context, args []Value) (Value, error)

// tls_close 关闭 TLS 连接或监听
// tls_close(conn) → bool
func builtinTLSClose(ctx *Context, args []Value) (Value, error)
```

#### 数据传输

```go
// tls_send 发送加密数据
// tls_send(socket, data) → bytes_sent
func builtinTLSSend(ctx *Context, args []Value) (Value, error)

// tls_recv 接收解密数据
// tls_recv(socket, length) → data
func builtinTLSRecv(ctx *Context, args []Value) (Value, error)
```

#### 信息获取

```go
// tls_get_cipher 获取协商的加密套件
// tls_get_cipher(socket) → cipher_suite_name
func builtinTLSGetCipher(ctx *Context, args []Value) (Value, error)

// tls_get_version 获取 TLS 版本
// tls_get_version(socket) → version_string
// 返回值: "TLS 1.0", "TLS 1.1", "TLS 1.2", "TLS 1.3"
func builtinTLSGetVersion(ctx *Context, args []Value) (Value, error)

// tls_get_cert_info 获取证书信息
// tls_get_cert_info(socket) → {subject, issuer, not_before, not_after, serial_number, dns_names}
func builtinTLSGetCertInfo(ctx *Context, args []Value) (Value, error)

// tls_set_cert 设置客户端证书（提示使用 options）
// tls_set_cert(socket, cert_file, key_file) → bool
// 注意: 实际应通过 tls_connect 的 options 设置
func builtinTLSSetCert(ctx *Context, args []Value) (Value, error)
```

#### 证书生成

```go
// tls_gen_cert 生成自签名证书
// tls_gen_cert(options?) → {cert_path, key_path}
// 选项: bits, days, common_name, out_dir, out_prefix
// 默认: bits=2048, days=365, common_name="JPL Generated"
func builtinTLSGenCert(ctx *Context, args []Value) (Value, error)
```

#### 类型定义

```go
// TLSSocketValue 表示 TLS 连接对象
type TLSSocketValue struct {
    conn     net.Conn    // TLS 连接
    config   *tls.Config // TLS 配置
    isServer bool        // 是否为服务端
}

// TLSListenerValue 表示 TLS 监听对象
type TLSListenerValue struct {
    listener net.Listener
    config   *tls.Config
}
```

### HTTP Client

> 模块导入: `import "http"`
>
> **自动解压缩**：http_get 等函数会自动发送 `Accept-Encoding: gzip, deflate, br` 头，并自动解压缩响应体（支持 gzip、deflate、brotli）。

#### 简单请求

```go
// http_get 执行 GET 请求
// http_get(url, options?) → HTTPResponseValue
// 自动检测 HTTPS 并使用 TLS
func builtinHTTPGet(ctx *Context, args []Value) (Value, error)

// http_post 执行 POST 请求
// http_post(url, options?) → HTTPResponseValue
// 选项支持: body, json, form, headers, timeout, auth, etc.
func builtinHTTPPost(ctx *Context, args []Value) (Value, error)

// http_put 执行 PUT 请求
// http_put(url, options?) → HTTPResponseValue
func builtinHTTPPut(ctx *Context, args []Value) (Value, error)

// http_delete 执行 DELETE 请求
// http_delete(url, options?) → HTTPResponseValue
func builtinHTTPDelete(ctx *Context, args []Value) (Value, error)

// http_head 执行 HEAD 请求
// http_head(url, options?) → HTTPResponseValue
func builtinHTTPHead(ctx *Context, args []Value) (Value, error)

// http_patch 执行 PATCH 请求
// http_patch(url, options?) → HTTPResponseValue
func builtinHTTPPatch(ctx *Context, args []Value) (Value, error)
```

#### 通用请求

```go
// http_request 执行通用 HTTP 请求
// http_request(method, url, options?) → HTTPResponseValue
// method: "GET", "POST", "PUT", "DELETE", "HEAD", "PATCH", etc.
func builtinHTTPRequest(ctx *Context, args []Value) (Value, error)
```

#### 请求选项

```go
// HTTPOptions 结构定义
type httpOptions struct {
    Headers         map[string]string  // 自定义请求头
    Timeout         int                // 超时时间（秒）
    FollowRedirects bool               // 是否跟随重定向
    MaxRedirects    int                // 最大重定向次数
    VerifySSL       bool               // 是否验证 SSL 证书
    Proxy           string             // 代理地址
    Body            []byte             // 原始请求体
    ContentType     string             // 内容类型
}

// 支持的 options 字段:
// headers: map[string]string    - 自定义请求头
// timeout: int                   - 超时秒数（默认 30）
// follow_redirects: bool         - 跟随重定向（默认 true）
// max_redirects: int              - 最大重定向次数（默认 10）
// verify_ssl: bool                - 验证 SSL 证书（默认 true）
// proxy: string                  - HTTP/HTTPS 代理地址
// body: string                    - 原始请求体
// json: object                    - JSON 请求体（自动设置 Content-Type: application/json）
// form: map[string]string         - Form 请求体（自动设置 Content-Type: application/x-www-form-urlencoded）
// auth: {username, password}      - 基本认证（自动设置 Authorization 头）
```

#### 响应对象

```go
// HTTPResponseValue 表示 HTTP 响应
type HTTPResponseValue struct {
    Status        int               // HTTP 状态码
    StatusText    string            // 状态文本
    Headers       map[string]string // 响应头
    Body          []byte            // 响应体
    ContentLength int64             // 内容长度
    Time          float64           // 请求耗时（秒）
}

// 响应对象方法:
// Type() → TypeObject
// Bool() → status >= 200 && status < 300
// Int() → status code
// Object() → {status, status_text, headers, body, content_length, time}
// String() → "HTTPResponse(status status_text)"
```

---

## 压缩/归档模块

> 模块导入: `import "gzip"`, `import "zlib"`, `import "brotli"`, `import "zip"`, `import "tar"`

### gzip 压缩

```go
// gzencode 使用 gzip 压缩数据
// gzencode(data) → string
func builtinGzencode(ctx *Context, args []Value) (Value, error)

// gzdecode 解压 gzip 数据
// gzdecode(data) → string | null
func builtinGzdecode(ctx *Context, args []Value) (Value, error)

// gzfile 读取 gzip 文件并返回行数组
// gzfile(filename) → array | null
func builtinGzfile(ctx *Context, args []Value) (Value, error)

// writegzfile 写入 gzip 文件
// writegzfile(filename, data) → int | null
func builtinWriteGzfile(ctx *Context, args []Value) (Value, error)

// gzopen 打开 gzip 文件
// gzopen(filename, mode) → GzipValue | null
// mode: "r", "w", "a"
func builtinGzopen(ctx *Context, args []Value) (Value, error)

// gzread 读取 gzip 数据
// gzread(gz, length?) → string | null
func builtinGzread(ctx *Context, args []Value) (Value, error)

// gzwrite 写入 gzip 数据
// gzwrite(gz, data) → int
func builtinGzwrite(ctx *Context, args []Value) (Value, error)

// gzclose 关闭 gzip 文件
// gzclose(gz) → int
func builtinGzclose(ctx *Context, args []Value) (Value, error)

// gzgets 读取一行
// gzgets(gz, length?) → string | null
func builtinGzgets(ctx *Context, args []Value) (Value, error)

// gzeof 检查是否到达文件末尾
// gzeof(gz) → bool
func builtinGzeof(ctx *Context, args []Value) (Value, error)

// GzipValue gzip 文件对象
type GzipValue struct {
	file   *os.File
	reader *gzip.Reader
	writer *gzip.Writer
	mode   string
	path   string
	eof    bool
}
```

### brotli 压缩

```go
// brotli_encode 使用 brotli 压缩数据
// brotli_encode(data) → string
func builtinBrotliEncode(ctx *Context, args []Value) (Value, error)

// brotli_decode 解压 brotli 数据
// brotli_decode(data) → string | null
func builtinBrotliDecode(ctx *Context, args []Value) (Value, error)

// brotli_compress_file 压缩文件为 brotli 格式
// brotli_compress_file(source, dest) → int | null
func builtinBrotliCompressFile(ctx *Context, args []Value) (Value, error)

// brotli_decompress_file 解压 brotli 文件
// brotli_decompress_file(source, dest) → int | null
func builtinBrotliDecompressFile(ctx *Context, args []Value) (Value, error)

// brotli_open 打开 brotli 文件
// brotli_open(filename, mode) → BrotliValue | null
// mode: "r", "w", "a"
func builtinBrotliOpen(ctx *Context, args []Value) (Value, error)

// brotli_read 读取 brotli 数据
// brotli_read(handle, length?) → string
func builtinBrotliRead(ctx *Context, args []Value) (Value, error)

// brotli_write 写入 brotli 数据
// brotli_write(handle, data) → int
func builtinBrotliWrite(ctx *Context, args []Value) (Value, error)

// brotli_close 关闭 brotli 文件
// brotli_close(handle) → null
func builtinBrotliClose(ctx *Context, args []Value) (Value, error)

// BrotliValue brotli 文件句柄
type BrotliValue struct {
	path      string
	file      *os.File
	reader    *brotli.Reader
	writer    *brotli.Writer
	writerBuf *bytes.Buffer
	isClosed  bool
}
```

### zlib 压缩

```go
// zlib_encode 使用 zlib 压缩数据
// zlib_encode(data) → string
func builtinZlibEncode(ctx *Context, args []Value) (Value, error)

// zlib_decode 解压 zlib 数据
// zlib_decode(data) → string | null
func builtinZlibDecode(ctx *Context, args []Value) (Value, error)

// deflate 使用 deflate 压缩
// deflate(data) → string
func builtinDeflate(ctx *Context, args []Value) (Value, error)

// inflate 解压 deflate 数据
// inflate(data) → string | null
func builtinInflate(ctx *Context, args []Value) (Value, error)
```

### zip 归档

```go
// zip_open 打开 zip 文件
// zip_open(filename) → ZipHandle | null
func builtinZipOpen(ctx *Context, args []Value) (Value, error)

// zip_read 读取下一个条目
// zip_read(zip) → ZipEntry | bool
func builtinZipRead(ctx *Context, args []Value) (Value, error)

// zip_entry_name 获取条目名称
// zip_entry_name(entry) → string | null
func builtinZipEntryName(ctx *Context, args []Value) (Value, error)

// zip_entry_filesize 获取原始文件大小
// zip_entry_filesize(entry) → int | null
func builtinZipEntryFilesize(ctx *Context, args []Value) (Value, error)

// zip_entry_compressedsize 获取压缩后大小
// zip_entry_compressedsize(entry) → int | null
func builtinZipEntryCompressedSize(ctx *Context, args []Value) (Value, error)

// zip_entry_read 读取条目内容
// zip_entry_read(entry, length?) → string | null
func builtinZipEntryRead(ctx *Context, args []Value) (Value, error)

// zip_entry_close 关闭条目
// zip_entry_close(entry) → int
func builtinZipEntryClose(ctx *Context, args []Value) (Value, error)

// zip_close 关闭 zip 文件
// zip_close(zip) → int
func builtinZipClose(ctx *Context, args []Value) (Value, error)

// zip_create 创建 zip 文件
// zip_create(filename, entries) → int | null
// entries: [{"name": "file.txt", "content": "..."}]
func builtinZipCreate(ctx *Context, args []Value) (Value, error)

// ZipHandle zip 文件句柄
type ZipHandle struct {
	filename string
	reader   *zip.ReadCloser
	entries  []*zip.File
	current  int
}

// ZipEntry zip 文件条目
type ZipEntry struct {
	file *zip.File
	name string
}
```

### tar 归档

```go
// tar_open 打开 tar 文件
// tar_open(filename) → TarHandle | null
func builtinTarOpen(ctx *Context, args []Value) (Value, error)

// tar_read 读取下一个条目
// tar_read(tar) → TarEntry | bool
func builtinTarRead(ctx *Context, args []Value) (Value, error)

// tar_entry_name 获取条目名称
// tar_entry_name(entry) → string | null
func builtinTarEntryName(ctx *Context, args []Value) (Value, error)

// tar_entry_size 获取文件大小
// tar_entry_size(entry) → int | null
func builtinTarEntrySize(ctx *Context, args []Value) (Value, error)

// tar_entry_isdir 检查是否为目录
// tar_entry_isdir(entry) → bool
func builtinTarEntryIsdir(ctx *Context, args []Value) (Value, error)

// tar_entry_read 读取条目内容
// tar_entry_read(entry, length?) → string | null
func builtinTarEntryRead(ctx *Context, args []Value) (Value, error)

// tar_entry_close 关闭条目
// tar_entry_close(entry) → int
func builtinTarEntryClose(ctx *Context, args []Value) (Value, error)

// tar_close 关闭 tar 文件
// tar_close(tar) → int
func builtinTarClose(ctx *Context, args []Value) (Value, error)

// tar_create 创建 tar 文件
// tar_create(filename, entries) → int | null
// entries: [{"name": "file.txt", "content": "..."}]
func builtinTarCreate(ctx *Context, args []Value) (Value, error)

// TarHandle tar 文件句柄
type TarHandle struct {
	filename string
	reader   *tar.Reader
	file     *os.File
}

// TarEntry tar 文件条目
type TarEntry struct {
	header *tar.Header
	name   string
	size   int64
	isdir  bool
}
```

---

## 日期时间模块 API

> 模块导入: `import "datetime"`

### 时间获取

```go
// time 返回当前 Unix 时间戳（秒级 float64）
// time() → float
// 示例: time() → 1711209600.123456
func builtinTime(ctx *Context, args []Value) (Value, error)

// now 返回当前时间对象或格式化字符串
// now([format]) → object/string
// 示例: now() → {year: 2026, month: 3, ...}
//       now("Y-m-d") → "2026-03-26"
func builtinNow(ctx *Context, args []Value) (Value, error)

// date 格式化时间戳
// date(format, [timestamp]) → string
// 示例: date("Y-m-d") → "2026-03-26"
func builtinDate(ctx *Context, args []Value) (Value, error)

// gmdate 格式化 GMT/UTC 时间
// gmdate(format, [timestamp]) → string
func builtinGmdate(ctx *Context, args []Value) (Value, error)

// microtime 微秒级时间戳
// microtime() → float
func builtinMicrotime(ctx *Context, args []Value) (Value, error)
```

### 日期解析与验证

```go
// strtotime 解析日期时间字符串为 Unix 时间戳
// strtotime(datetime_string, [base_timestamp]) → float | null
// 支持格式:
//   - "2006-01-02 15:04:05"
//   - "2006-01-02"
//   - "01/02/2006"
//   - "Jan 2, 2006"
//   - ISO 8601: "2006-01-02T15:04:05Z"
//   - 相对时间: "+1 day", "-2 weeks", "+3 months", "+1 year"
// 示例:
//   strtotime("2026-03-26") → 1742947200
//   strtotime("+1 day") → 明天此时的时间戳
//   strtotime("-2 weeks") → 两周前的时间戳
func builtinStrtotime(ctx *Context, args []Value) (Value, error)

// checkdate 验证公历日期是否有效
// checkdate(month, day, year) → bool
// 示例:
//   checkdate(2, 29, 2024) → true  (2024 是闰年)
//   checkdate(2, 29, 2023) → false (2023 不是闰年)
//   checkdate(13, 1, 2026) → false (月份无效)
func builtinCheckdate(ctx *Context, args []Value) (Value, error)
```

### 睡眠

```go
// sleep 暂停执行指定毫秒数
// sleep(ms) → null
// 示例: sleep(1000)  // 暂停 1 秒
func builtinSleep(ctx *Context, args []Value) (Value, error)

// usleep 微秒级暂停（在 process 模块中定义）
// usleep(microseconds) → null
// 示例: usleep(500000)  // 暂停 0.5 秒
func builtinUsleep(ctx *Context, args []Value) (Value, error)
```

### 日期信息

```go
// getdate 返回日期信息对象
// getdate([timestamp]) → object
func builtinGetdate(ctx *Context, args []Value) (Value, error)

// gettimeofday 返回时间信息对象
// gettimeofday() → object
func builtinGettimeofday(ctx *Context, args []Value) (Value, error)

// strftime 按 strftime 格式化
// strftime(format, [timestamp]) → string
func builtinStrftime(ctx *Context, args []Value) (Value, error)

// localtime 返回本地时间信息
// localtime([timestamp], [is_assoc]) → array/object
func builtinLocaltime(ctx *Context, args []Value) (Value, error)

// mktime 生成本地时间戳
// mktime(hour, minute, second, month, day, year) → float
func builtinMktime(ctx *Context, args []Value) (Value, error)

// gmmktime 生成 GMT 时间戳
// gmmktime(hour, minute, second, month, day, year) → float
func builtinGmmktime(ctx *Context, args []Value) (Value, error)
```

---

## 字符串模块 API

> 模块导入: `import "strings"`

### 基础函数

```go
// strlen 字符串长度（字节数）
// strlen(str) → int
func builtinStrlen(ctx *Context, args []Value) (Value, error)

// substr 截取子串
// substr(str, start, [length]) → string
func builtinSubstr(ctx *Context, args []Value) (Value, error)

// strpos 查找子串位置
// strpos(haystack, needle, [offset]) → int | -1
func builtinStrpos(ctx *Context, args []Value) (Value, error)

// str_replace 替换子串
// str_replace(search, replace, subject) → string
func builtinStrReplace(ctx *Context, args []Value) (Value, error)
```

### 大小写转换

```go
// toUpper 转大写
// toUpper(str) → string
func builtinToUpper(ctx *Context, args []Value) (Value, error)

// toLower 转小写
// toLower(str) → string
func builtinToLower(ctx *Context, args []Value) (Value, error)

// ucfirst 首字母大写
// ucfirst(str) → string
// 示例: ucfirst("hello world") → "Hello world"
func builtinUcfirst(ctx *Context, args []Value) (Value, error)

// ucwords 每个单词首字母大写
// ucwords(str) → string
// 示例: ucwords("hello world") → "Hello World"
func builtinUcwords(ctx *Context, args []Value) (Value, error)
```

### 子串替换

```go
// substr_replace 替换指定位置的子串
// substr_replace(str, replacement, start, [length]) → string
// 示例:
//   substr_replace("Hello World", "JPL", 6) → "Hello JPL"
//   substr_replace("Hello World", "JPL", 6, 5) → "Hello JPL"
//   substr_replace("Hello World", "JPL", -5) → "Hello JPL"
func builtinSubstrReplace(ctx *Context, args []Value) (Value, error)
```

---

## I/O 模块 API

> 模块导入: `import "io"`

### 输出函数

```go
// print 输出到 stdout（无换行）
// print(args...) → null
func builtinPrint(ctx *Context, args []Value) (Value, error)

// println 输出到 stdout（带换行）
// println(args...) → null
func builtinPrintln(ctx *Context, args []Value) (Value, error)

// puts 输出到 stdout（不带引号，带换行）
// puts(args...) → null
func builtinPuts(ctx *Context, args []Value) (Value, error)

// pp Pretty Print 格式化输出
// pp(args...) → null
func builtinPP(ctx *Context, args []Value) (Value, error)
```

### 输入函数

```go
// input 从标准输入读取一行（带可选提示）
// input([prompt]) → string | null
// 示例:
//   $name = input("Enter your name: ")
//   $line = input()  // 无提示，直接读取
func builtinInput(ctx *Context, args []Value) (Value, error)

// readline input 的别名
// readline([prompt]) → string | null
func builtinReadline(ctx *Context, args []Value) (Value, error)
```

### 字符串工具

```go
// echo 拼接参数为字符串（不输出）
// echo(args...) → string
func builtinEcho(ctx *Context, args []Value) (Value, error)

// format 格式化字符串
// format(template, args...) → string
func builtinFormat(ctx *Context, args []Value) (Value, error)

// assert 断言检查
// assert(condition, [message]) → null
func builtinAssert(ctx *Context, args []Value) (Value, error)
```

---

## 工具模块 API

> 模块导入: `import "util"`（仅全局注册，无模块）

```go
// len 返回字符串、数组或对象的长度
// len(value) → int
// 示例: len([1, 2, 3]) → 3
//       len("hello") → 5
//       len({a: 1, b: 2}) → 2
func builtinLen(ctx *Context, args []Value) (Value, error)

// typeof 返回值的类型名称
// typeof(value) → string
// 返回值: "null", "bool", "int", "float", "string", "array", "object", "func", "bigint", "bigdecimal", "stream", "error", "range", "regex"
// 示例:
//   typeof(42) → "int"
//   typeof("hello") → "string"
//   typeof([1, 2, 3]) → "array"
//   typeof({a: 1}) → "object"
func builtinTypeof(ctx *Context, args []Value) (Value, error)

// dump 返回值的详细调试信息
// dump(value) → string
// 示例:
//   dump([1, {name: "Alice"}])
//   // 输出:
//   // [
//   //   1,
//   //   {
//   //     name: "Alice"
//   //   }
//   // ]
func builtinDump(ctx *Context, args []Value) (Value, error)
```

---

## IP 地址模块 API

> 模块导入: `import "ip"`

### 转换函数

```go
// ip2long IPv4 转整数
// ip2long(ip) → int
func builtinIP2Long(ctx *Context, args []Value) (Value, error)

// long2ip 整数转 IPv4
// long2ip(long) → string
func builtinLong2IP(ctx *Context, args []Value) (Value, error)
```

### 验证与范围

```go
// ip_valid 验证 IP 地址
// ip_valid(ip) → bool
func builtinIPValid(ctx *Context, args []Value) (Value, error)

// ip_version 检测 IP 版本
// ip_version(ip) → int (4 or 6)
func builtinIPVersion(ctx *Context, args []Value) (Value, error)

// ip_in_range 检查 IP 是否在 CIDR 范围内
// ip_in_range(ip, range) → bool
// 示例:
//   ip_in_range("192.168.1.100", "192.168.1.0/24") → true
//   ip_in_range("10.0.0.1", "192.168.1.0/24") → false
//   ip_in_range("127.0.0.1", "127.0.0.1") → true  (纯 IP 匹配)
func builtinIPInRange(ctx *Context, args []Value) (Value, error)
```

---

## JSON 模块 API

> 模块导入: `import "json"`

```go
// json_encode 序列化为 JSON
// json_encode(value, [pretty]) → string
func builtinJSONEncode(ctx *Context, args []Value) (Value, error)

// json_decode 解析 JSON 字符串
// json_decode(str) → value
func builtinJSONDecode(ctx *Context, args []Value) (Value, error)

// json_pretty 美化输出 JSON
// json_pretty(value) → string
func builtinJSONPretty(ctx *Context, args []Value) (Value, error)

// json_validate 验证 JSON 是否合法（不返回解析结果）
// json_validate(str) → bool
// 示例:
//   json_validate('{"name": "Alice"}') → true
//   json_validate('{invalid}') → false
func builtinJSONValidate(ctx *Context, args []Value) (Value, error)
```

---

## 数学模块增强 API

> 模块导入: `import "math"`

### 新增函数

```go
// cbrt 计算立方根
// cbrt(n) → float
// 示例: cbrt(27) → 3.0
//       cbrt(-8) → -2.0
func builtinCbrt(ctx *Context, args []Value) (Value, error)

// log2 计算以 2 为底的对数
// log2(n) → float
// 示例: log2(1024) → 10.0
func builtinLog2(ctx *Context, args []Value) (Value, error)

// clamp 将值限制在 [min, max] 范围内
// clamp(value, min, max) → number
// 示例:
//   clamp(5, 0, 10) → 5
//   clamp(-5, 0, 10) → 0
//   clamp(15, 0, 10) → 10
func builtinClamp(ctx *Context, args []Value) (Value, error)

// sign 返回数字的符号
// sign(n) → int (-1, 0, 1)
// 示例:
//   sign(-42) → -1
//   sign(0) → 0
//   sign(42) → 1
func builtinSign(ctx *Context, args []Value) (Value, error)

// intdiv 整数除法
// intdiv(a, b) → int
// 示例:
//   intdiv(7, 2) → 3
//   intdiv(-7, 2) → -3
func builtinIntDiv(ctx *Context, args []Value) (Value, error)
```

---

## 加密模块增强 API

> 模块导入: `import "crypto"`

### 新增函数

```go
// random_bytes 生成密码学安全的随机字节
// random_bytes(length) → string
// 示例:
//   $bytes = random_bytes(32)  // 32 字节随机数据
func builtinRandomBytes(ctx *Context, args []Value) (Value, error)

// hash 通用 hash 函数
// hash(algo, data, [raw_output]) → string
// 支持的算法: md5, sha1, sha256, sha512
// 示例:
//   hash("sha256", "Hello") → "185f8db32271fe25f561a6fc938b2e264306ec304eda518007d1764826381969"
//   hash("md5", "Hello") → "8b1a9953c4611296a827abf8c47804d7"
//   hash("sha256", "Hello", true) → 原始二进制字符串
func builtinHash(ctx *Context, args []Value) (Value, error)
```

---

## 文件 I/O 模块别名

> 模块导入: `import "file"`

### 新增别名

```go
// file_exists exists 的别名（PHP 风格）
// file_exists(path) → bool
// 等价于 exists(path)

// is_file isFile 的别名（snake_case）
// is_file(path) → bool
// 等价于 isFile(path)

// is_dir isDir 的别名（snake_case）
// is_dir(path) → bool
// 等价于 isDir(path)

// file_size fileSize 的别名（snake_case）
// file_size(path) → int
// 等价于 fileSize(path)
```
