# 内置函数表

本文档列出的是已内置的函数签名。

---

## 已实现的标准库内置函数

| 类别 | 函数 |
|------|------|
| I/O | **print, println, puts, pp**, echo, format, assert |
| 工具 | len |
| 函数式编程 | map, filter, reject, reduce, find, some, every, sort, contains, unique, partition, flattenDeep, difference, union, zip, unzip |
| 类型检查 | is_null, is_bool, is_int, is_float, is_string, is_array, is_object, is_func, **is_real, is_double, is_integer, is_long, is_numeric, is_scalar, is_bigint, is_bigdecimal** |
| 类型转换 | **语法级: int(), float(), string(), bool()** / 函数级: **intval, floatval, strval, boolval, empty** |
| 字符串 | strlen, substr, strpos, str_replace, trim, toUpper, toLower, split, join, startsWith, endsWith, charAt, repeat, reverse, **implode, explode, chop, ltrim, rtrim, strcmp, strcasecmp, strncmp, strncasecmp, stripos, strrpos, strripos, strstr, stristr, strchr, sprintf, printf, vsprintf, vprintf, number_format, ord, chr, nl2br, bin2hex** |
| 数组 | push, pop, shift, unshift, splice, indexOf, lastIndexOf, slice, array_reverse, includes, flat, unique, **count, sizeof, array_key_exists, key_exists, array_merge, array_sum, array_product, array_values, array_diff, array_intersect, in_array, array_copy** |
| 数学 | abs, ceil, floor, round, pow, sqrt, min, max, random, randomInt, parseInt, parseFloat, isNaN, isFinite, **sin, cos, tan, asin, acos, atan, atan2, sinh, cosh, tanh, log, log10, exp, pi, fmod, hypot, deg2rad, rad2deg** |
| **日期时间** | **time, date, now, sleep, microtime, getdate, gettimeofday, strftime, gmdate, localtime, mktime, gmmktime** |
| 动态常量 | define, defined |
| Hash/编码 | md5, sha1, base64_encode, base64_decode, crc32 |
| **URL 处理** | **urlencode, urldecode, rawurlencode, rawurldecode, parse_url** |
| **JSON 处理** | **json_encode, json_decode, json_pretty** |
| **对象解析** | **parse_object** |
| 文件 I/O | read, readLines, write, append, exists, mkdir, mkdirAll, rmdir, listDir, stat, fileSize, isFile, isDir, dirname, basename, extname, joinPath, absPath, relPath, cwd, realpath, chdir, rename, unlink, is_readable, is_writable, chmod, scandir, glob |
| GC | gc, gc_info |
| 调试 | errors, last_error, clear_errors |
| 错误处理 | error |
| 反射 | typeof, varexists, getvar, setvar, listvars, listfns, fn_exists, getfninfo, callfn, **function_exists, is_callable, get_defined_functions, get_defined_constants, func_num_args, func_get_arg, func_get_args** |
| **版本信息** | **jpl_version** |
| **UTF-8** | **utf8_encode, utf8_decode** |
| **进程控制** | **exit, die, exec, system, shell_exec, getenv, setenv, getppid, getlogin, tmpdir, hostname, usleep, putenv, proc_open, proc_close, proc_wait, proc_status, spawn, kill, waitpid, fork, pipe, sigwait** |
| **异步文件IO** | **file_get_async, file_put_async, file_append_async, file_get_bytes, file_put_bytes, file_read_lines, file_read_chunks, file_get_batch, file_put_batch, file_parallel, file_with_lock** |
| **网络编程** | **net_tcp_listen, net_tcp_connect, net_tcp_accept, net_unix_listen, net_unix_connect, net_unix_accept, net_udp_bind, net_udp_sendto, net_udp_recvfrom, net_send, net_recv, net_close, net_getsockname, net_getpeername, net_set_nonblock, net_is_unix, dns_resolve, dns_resolve_one, dns_resolve_v4, dns_resolve_v6, dns_get_records** |
| **事件循环** | **ev_registry_new, ev_loop_new, ev_attach, ev_run, ev_run_once, ev_stop, ev_is_running, ev_timer_now, ev_on_read, ev_on_write, ev_on_accept, ev_on_timer, ev_on_timer_once, ev_on_signal, ev_off, ev_off_read, ev_off_write, ev_off_timer, ev_off_signal, ev_clear, ev_count** |
| **二进制处理** | **pack, unpack, buffer_new, buffer_new_from, buffer_set_endian, buffer_write_int8, buffer_write_int16, buffer_write_int32, buffer_write_uint8, buffer_write_uint16, buffer_write_uint32, buffer_write_float32, buffer_write_float64, buffer_write_string, buffer_write_bytes, buffer_read_int8, buffer_read_int16, buffer_read_int32, buffer_read_uint8, buffer_read_uint16, buffer_read_uint32, buffer_read_float32, buffer_read_float64, buffer_read_string, buffer_read_bytes, buffer_seek, buffer_tell, buffer_length, buffer_reset, buffer_to_bytes, buffer_to_string, is_buffer** |

---

## 内置函数（类型和值处理）

### 动态常量

```jpl
define("MY_CONST", 42);    // 定义常量（带括号）
define "MY_CONST", 42;     // 定义常量（无括号，特例语法）
defined("MY_CONST")         // 检查常量是否已定义，返回 bool
// 注意：预设常量（PI、E 等）不可被 define 重新定义
```

### Hash/编码

```jpl
md5("hello")              // 返回 MD5 十六进制字符串
sha1("hello")             // 返回 SHA1 十六进制字符串
md5_file("config.json")   // 返回文件 MD5 十六进制字符串
sha1_file("data.txt")     // 返回文件 SHA1 十六进制字符串
base64_encode("hello")    // 返回 Base64 编码字符串
base64_decode("aGVsbG8=") // 返回 Base64 解码字符串
crc32("hello")            // 返回 CRC32 校验和（整数）
```

### 函数式编程

```jpl
map($arr, $fn)            // 映射
filter($arr, $fn)         // 过滤
reduce($arr, $fn, $init)  // 归约
find($arr, $fn)           // 查找第一个匹配元素
some($arr, $fn)           // 是否存在匹配元素
every($arr, $fn)          // 是否所有元素都匹配
sort($arr, $cmp)          // 排序
contains($arr, $val)      // 是否包含值
```

### 字符串函数

```jpl
strlen("hello")               // 字符串长度 → 5
substr("hello", 1, 3)         // 子串 → "ell"
strpos("hello", "ll")         // 查找位置 → 2
str_replace("aab", "a", "x") // 替换 → "xxb"
trim("  hi  ")                // 去除首尾空白 → "hi"
toUpper("hello")              // 转大写 → "HELLO"
toLower("HELLO")              // 转小写 → "hello"
split("a,b,c", ",")           // 分割 → ["a", "b", "c"]
join(["a", "b"], "-")         // 连接 → "a-b"
startsWith("hello", "he")     // 是否以...开头 → true
endsWith("hello", "lo")       // 是否以...结尾 → true
charAt("hello", 0)            // 指定位置字符 → "h"
repeat("ab", 3)               // 重复 → "ababab"
reverse("hello")              // 反转 → "olleh"
```

### 数组函数

```jpl
push([1,2], 3)              // 尾部追加 → [1, 2, 3]
pop([1,2,3])                // 尾部弹出 → 3
shift([1,2,3])              // 头部弹出 → 1
unshift([2,3], 1)           // 头部追加 → [1, 2, 3]
splice([1,2,3,4], 1, 2)     // 删除指定位置元素 → [1, 4]
indexOf([1,2,3], 2)         // 查找索引 → 1
lastIndexOf([1,2,3,2], 2)   // 最后出现索引 → 3
slice([1,2,3,4], 1, 3)      // 切片 → [2, 3]
array_reverse([1,2,3])      // 反转 → [3, 2, 1]
array_fill(0, 5, "x")      // 填充数组 → ["x", "x", "x", "x", "x"]
array_fill_keys(["a", "b"], 1) // 键填充 → {a: 1, b: 1}
array_flip(["a" => 1, "b" => 2]) // 键值交换 → {1: "a", 2: "b"}
includes([1,2,3], 2)        // 是否包含 → true
flat([[1],[2,3]])           // 展平 → [1, 2, 3]
unique([1,2,2,3,3])         // 去重 → [1, 2, 3]

// 排序
sort($arr)                  // 升序排序
sort($arr, fn($a, $b) { return $a > $b; }) // 降序排序
usort($arr, fn($a, $b) { return $a < $b; }) // 自定义排序

// 键检查
array_key_exists("a", {a: 1}) // → true
key_exists(0, [1, 2, 3])      // → true

// 合并
array_merge([1, 2], [3, 4])   // → [1, 2, 3, 4]

// 计算
array_sum([1, 2, 3])          // → 6
array_product([2, 3, 4])      // → 24

// 提取
array_values({a: 1, b: 2})    // → [1, 2]

// 集合
array_diff([1, 2, 3], [2, 3])       // → [1]
array_intersect([1, 2, 3], [2, 3, 4]) // → [2, 3]
in_array(2, [1, 2, 3])              // → true

// 复制
array_copy([1, [2, 3]])       // → 深度复制
```

### 数学函数

```jpl
abs(-5)                     // 绝对值 → 5
abs(-3.14)                  // 支持浮点 → 3.14
ceil(3.2)                   // 向上取整 → 4
floor(3.8)                  // 向下取整 → 3
round(3.14159, 2)           // 四舍五入 → 3.14
pow(2, 10)                  // 幂运算 → 1024
sqrt(16)                    // 平方根 → 4.0
min(3, 1, 4)                // 最小值 → 1
max(3, 1, 4)                // 最大值 → 4
random()                    // 随机浮点 [0, 1)
randomInt(1, 100)           // 随机整数 [1, 100]
parseInt("42")              // 字符串转整数 → 42
parseFloat("3.14")          // 字符串转浮点 → 3.14
isNaN(NaN)                  // 检查 NaN → true
isFinite(42)                // 检查有限数 → true

// 三角函数
sin(PI / 6)                 // → 0.5
cos(0)                      // → 1.0
tan(PI / 4)                 // → 1.0
asin(0.5)                   // → π/6
acos(1)                     // → 0
atan(1)                     // → π/4
atan2(1, 1)                 // → π/4

// 双曲函数
sinh(0)                     // → 0
cosh(0)                     // → 1
tanh(0)                     // → 0

// 对数/指数
log(E)                      // → 1
log10(100)                  // → 2
exp(1)                      // → E
pi()                        // → 3.141592653589793

// 其他
fmod(5.5, 2)                // → 1.5
hypot(3, 4)                 // → 5
deg2rad(180)                // → PI
rad2deg(PI)                 // → 180
```

### URL 处理函数

```jpl
// URL 编码/解码
urlencode("hello world")       // → "hello+world"
urldecode("hello+world")       // → "hello world"
rawurlencode("hello world")   // → "hello%20world"
rawurldecode("hello%20world") // → "hello world"

// URL 解析
parse_url("https://user:pass@example.com:8080/path?query=1#section")
// → {scheme: "https", host: "example.com", port: 8080, user: "user", pass: "pass", path: "/path", query: "query=1", fragment: "section"}
```

### JSON 处理函数

```jpl
// 将值序列化为 JSON 字符串
json_encode([1, 2, 3])                           // → "[1,2,3]"
json_encode({"name": "Alice", "age": 25})          // → "{\"name\":\"Alice\",\"age\":25}"

// 美化输出
json_encode({"a": 1, "b": 2}, true)
// → {
//      "a": 1,
//      "b": 2
//    }

// 简写形式
json_pretty({"name": "Alice"})                   // 同上，等效于 json_encode($value, true)

// 解析 JSON 字符串
json_decode("[1, 2, 3]")                          // → [1, 2, 3]
json_decode("{\"name\":\"Alice\"}")               // → {"name": "Alice"}
json_decode("null")                              // → null

// 数字解析智能策略
json_decode("123")                               // → Int: 123
json_decode("9007199254740993")                  // → BigInt (超出 int64 精确范围)
json_decode("99999999999999999999")              // → BigInt (超出 int64 最大值)
json_decode("1e10")                              // → Int: 10000000000
json_decode("1e20")                              // → BigInt: 100000000000000000000
json_decode("1.5e-5")                            // → Float: 0.000015
json_decode("1.5")                               // → Float: 1.5
json_decode("0.1")                               // → Float: 0.1
```

### 类型检查/转换函数

```jpl
// 类型检查
is_null(null)               // → true
is_bool(true)               // → true
is_int(42)                  // → true
is_float(3.14)              // → true
is_string("hi")             // → true
is_array([1,2])             // → true
is_object({a:1})            // → true
is_func(fn() {})            // → true
is_real(3.14)               // → true (is_float 别名)
is_integer(42)              // → true (is_int 别名)
is_numeric("42")            // → true
is_scalar(42)               // → true
is_scalar([1, 2])           // → false

// 类型转换（函数式）
intval("42")                // → 42
intval("ff", 16)            // → 255 (十六进制)
floatval("3.14")            // → 3.14
strval(42)                  // → "42"
boolval(1)                  // → true
empty("")                   // → true
empty(0)                    // → true
empty(null)                 // → true
empty("hello")              // → false

// 类型转换（Go风格语法级）
$num = int("42")            // → 42
$float = float("3.14")      // → 3.14
$str = string(123)          // → "123"
$flag = bool(1)             // → true

// 在表达式中使用
$sum = int("10") + 5        // → 15
$val = int(float("3.7"))    // → 3 (嵌套转换)
```

### 对象解析（安全）

```jpl
// parse_object() - 安全解析对象字面量字符串
// 只解析字面量，拒绝函数调用和表达式，避免代码注入

$obj = parse_object("{a: 1, b: 2}")                    // {a: 1, b: 2}
$obj = parse_object("{name: 'John', age: 30}")        // {name: "John", age: 30}
$obj = parse_object("{items: [1, 2, 3]}")            // 支持嵌套数组
$obj = parse_object("{user: {name: 'A'}}")            // 支持嵌套对象

// 这些会被拒绝（安全特性）
parse_object("{x: delete_all_files()}")   // 错误！拒绝函数调用
parse_object("{y: $x}")                    // 错误！拒绝变量
parse_object("{z: 1 + 2}")                 // 错误！拒绝表达式
```

---

## 内置函数（高级编程）

### 标准 I/O 函数

```jpl
print("hello", 42)          // 输出（不换行）
println("hello", 42)        // 输出并换行
echo("a", "b")              // 返回拼接字符串 "a b"
log("debug info")           // 输出日志（带前缀）
format("Hi %s", "JPL")      // 格式化字符串 → "Hi JPL"
assert(true)                // 断言，失败抛异常
assert(1 == 1, "msg")       // 带消息断言

// print - 输出（不换行）
print "Hello"                    // 输出到 stdout
print(STDERR, "error!")          // 输出到指定流

// println - 输出（换行）
println "World"                  // 输出到 stdout
println(STDERR, "fatal error")   // 输出到指定流

// print/println 多参数（空格分隔）
print("count:", 42)              // → count: 42
println(STDERR, "code:", 500)    // → code: 500（到 stderr）

// puts - 输出到 stdout（简单输出）
puts "Simple output"

// pp - 美观打印（格式化输出，适合调试）
pp {a: 1, b: 2}           // 带缩进和格式化的对象
pp [1, 2, [3, 4]]         // 带缩进的数组

// echo - 输出字符串或数组
$s = echo "hello"
$a = echo [1, 2, 3]
```

### 调试函数

```jpl
errors()                    // 返回所有错误消息数组
last_error()                // 返回最后一条错误消息，无错误返回 null
clear_errors()              // 清空错误日志
```

### VM/反射函数

```jpl
// 函数参数获取（只能在用户定义函数内调用）
fn test(a, b, c) {
    func_num_args()         // → 3
    func_get_arg(0)         // → a 的值
    func_get_arg(1)         // → b 的值
    func_get_args()         // → [a, b, c]
    func_get_args()[0]      // → a 的值
}

// 函数存在检查
function_exists("print")    // → true
function_exists("nonexist") // → false
is_callable("print")        // → true
is_callable(42)             // → false

// 获取定义列表
get_defined_functions()     // → 所有函数名数组
get_defined_constants()     // → 所有常量名数组

// 版本信息
jpl_version()               // → "1.0.0"

// UTF-8 编解码
utf8_encode("Hello")        // → "48656c6c6f" (十六进制)
utf8_decode("48656c6c6f")   // → "Hello"
utf8_encode("中文")          // → "e4b8ade69687"
utf8_decode("e4b8ade69687") // → "中文"
```

示例：
```jpl
"hello" + 5;                // 类型不匹配，静默记录
println(last_error());      // → runtime error: cannot add string and int
println(len(errors()));     // → 1
clear_errors();             // 清空
println(len(errors()));     // → 0
```

### 错误处理函数

```jpl
error(message)                      // 创建错误对象，默认 code=0, type="Error"
error(message, code)                // 创建错误对象，指定错误码
error(message, code, type)          // 创建错误对象，指定错误码和类型
```

错误对象字段：
- `$e.message` — 错误消息（字符串）
- `$e.code` — 错误码（整数）
- `$e.type` — 错误类型（字符串）

示例：
```jpl
// 创建并抛出错误
throw error("not found", 404, "HttpError");

// 条件捕获
try {
    throw error("db error", 1, "DBError");
} catch ($e when $e.code == 404) {
    println("not found: " .. $e.message);
} catch ($e) {
    println($e.type .. ": " .. $e.message);
}

// 使用 typeof 检查
$e = error("test");
println(typeof($e));        // → error
println($e.message);        // → test
println($e.code);           // → 0
println($e.type);           // → Error
```

### 进程控制函数

```jpl
exit()                      // 终止脚本，退出码 0
exit(1)                     // 终止脚本，指定退出码（0-255）
die("Fatal error")          // 输出消息并终止，退出码 0
die("Fatal error", 2)       // 输出消息并终止，指定退出码
```

**特性说明**：
- exit/die 会**立即终止脚本执行**，无视 try/catch 块
- 支持**特例函数语法**（无括号调用）：`exit 5`、`die "message"`
- 多参数调用需使用逗号分隔：`die "msg", 1`
- CLI 会将退出码传递给操作系统

**使用示例**：
```jpl
// 验证输入后提前退出
if (ARGC < 2) {
    die "Usage: script.jpl <input_file>", 1
}

// 条件退出
if ($config == null) {
    exit 2
}

// 无视 try/catch 的终止
try {
    die "Critical error", 99
} catch ($e) {
    // 不会执行到这里
}
```

### GC 函数

```jpl
gc()                        // 手动触发垃圾回收
gc_info()                   // 返回 GC 统计信息
```

---

## 内置函数（标准库简写）

### 文件和 I/O

```jpl
// 文件读写
write("test.txt", "hello")  // 写入文件
append("test.txt", " world")// 追加文件
read("test.txt")            // 读取文件内容
readLines("test.txt")       // 按行读取 → 数组
exists("test.txt")          // 文件是否存在 → bool

// PHP 风格文件操作（Phase 11）
file_get_contents("input.txt")   // 读取文件内容（read 别名）
file_put_contents("out.txt", $data) // 写入文件内容，返回字节数
copy("source.txt", "backup.txt") // 复制文件
readfile("config.txt")          // 读取并返回文件内容

// 路径信息
pathinfo("/path/to/file.txt")   // → {dirname: "/path/to", basename: "file.txt", extension: "txt", filename: "file"}

// 目录操作
mkdir("mydir")              // 创建目录
mkdirAll("a/b/c")           // 递归创建目录
rmdir("mydir")              // 删除目录
listDir(".")                // 列出目录内容 → 数组

// 文件信息
stat("test.txt")            // 文件信息对象
fileSize("test.txt")        // 文件大小（字节）
isFile("test.txt")          // 是否为文件 → bool
isDir("mydir")              // 是否为目录 → bool

// 路径处理
dirname("/a/b/c.txt")       // → "/a/b"
basename("/a/b/c.txt")      // → "c.txt"
extname("/a/b/c.txt")       // → ".txt"
joinPath("a", "b", "c")     // → "a/b/c"

// 当前工作目录
cwd()                       // → 返回当前工作目录路径
realpath(".")              // → 返回规范化的绝对路径

// 文件系统操作
chdir("/tmp")              // 改变当前工作目录
rename("old.txt", "new.txt") // 重命名文件
unlink("temp.txt")          // 删除文件
chmod("script.sh", 0755)    // 修改文件权限

// 目录扫描
scandir(".")               // 扫描目录内容 → 数组
glob("*.go")                // 查找匹配文件 → 数组

// 权限检查
is_readable("file.txt")     // 检查是否可读 → bool
is_writable("file.txt")     // 检查是否可写 → bool

// 文件时间
fileatime("file.txt")       // 最后访问时间（Unix 时间戳）
filemtime("file.txt")       // 最后修改时间（Unix 时间戳）
filectime("file.txt")       // 状态改变时间（Unix 时间戳）
touch("newfile.txt")        // 创建空文件或更新时间
```

### 流 IO 函数

```jpl
// 打开文件流
$f = fopen("data.txt", "r")     // 只读模式
$f = fopen("output.txt", "w")   // 写入模式
$f = fopen("log.txt", "rw")     // 读写模式

// 读取操作
$data = fread($f, 1024)         // 读取指定字节数
$line = fgets($f)               // 读取一行
$eof = feof($f)                 // 检查是否到达末尾

// 写入操作
$n = fwrite($f, "Hello\n")      // 写入数据，返回写入字节数
fflush($f)                      // 刷新缓冲区

// 关闭流
fclose($f)                      // 关闭流

// 流元数据
$info = stream_get_meta_data($f)
$info["mode"]                   // → "r"
$info["uri"]                    // → "data.txt"
$info["seekable"]               // → true
$info["timed_out"]              // → false

// 类型检查
is_stream($f)                   // → true
is_stream("not a stream")       // → false

// 使用标准流
fwrite(STDOUT, "Hello\n")       // 输出到标准输出
fwrite(STDERR, "Error\n")       // 输出到标准错误
$line = fgets(STDIN)            // 从标准输入读取
```

### 流定位与操作函数

```jpl
// 流操作过程
$f = fopen("data.txt", "r")
fseek($f, 10, 0)            // 移动到位置 10（0=SEEK_SET）
$pos = ftell($f)            // 获取当前位置
rewind($f)                  // 回到开头
ftruncate($f, 100)          // 截断到 100 字节
fclose($f)

// fseek — 移动文件指针
// 参数: fseek(stream, offset, whence)
// whence: 0=SEEK_SET(开头), 1=SEEK_CUR(当前), 2=SEEK_END(末尾)
// 返回: 新位置（字节偏移）
$f = fopen("data.txt", "r")
$newPos = fseek($f, 10, 0)      // 从开头偏移 10 字节
$newPos = fseek($f, 5, 1)       // 从当前位置前进 5 字节
$newPos = fseek($f, -10, 2)     // 从末尾回退 10 字节
fclose($f)

// ftell — 获取文件指针当前位置
$pos = ftell($f)                // 返回当前字节偏移

// rewind — 重置文件指针到开头
rewind($f)                      // 等价于 fseek($f, 0, 0)

// ftruncate — 截断文件到指定大小
$f = fopen("data.txt", "rw")
ftruncate($f, 100)              // 截断到 100 字节
fclose($f)

// fgetcsv — 从流中读取一行 CSV 数据
// 参数: fgetcsv(stream, [delimiter])
// 返回: 字符串数组，到达末尾返回 null
$f = fopen("data.csv", "r")
$header = fgetcsv($f)           // 读取表头 ["name", "age", "city"]
$row = fgetcsv($f)              // 读取数据行 ["Alice", "30", "Beijing"]
fclose($f)

// 自定义分隔符（如 TSV）
$f = fopen("data.tsv", "r")
$row = fgetcsv($f, "\t")        // Tab 分隔
fclose($f)
```

**函数签名**：

| 函数 | 参数 | 返回值 | 说明 |
|------|------|--------|------|
| `fseek(stream, offset, whence)` | 流、偏移量、起始位置 | 新位置 (int) | 移动文件指针 |
| `ftell(stream)` | 流 | 当前位置 (int) | 获取文件指针位置 |
| `rewind(stream)` | 流 | null | 重置到开头 |
| `ftruncate(stream, size)` | 流、新大小 | true | 截断文件 |
| `fgetcsv(stream, [delimiter])` | 流、可选分隔符 | 数组或 null | 读取 CSV 行 |

**注意**：
- `fseek`/`ftell`/`rewind`/`ftruncate` 仅支持文件流（`fopen` 创建的流），不支持标准流
- `fgetcsv` 支持 CSV 引号转义（`""` 表示字面量 `"`）
- 默认分隔符为逗号 `,`，可通过第二个参数自定义

**流模式说明**：

| 模式 | 说明 | 文件不存在时 |
|------|------|-------------|
| `"r"` | 只读 | 返回错误 |
| `"w"` | 写入（覆盖） | 创建文件 |
| `"rw"` | 读写 | 创建文件 |

### 系统函数

```jpl
// 磁盘空间
disk_free_space("/")        // 可用空间（字节）
disk_total_space("/")       // 总空间（字节）

// 进程信息
getpid()                    // 当前进程 ID
getuid()                    // 当前用户 ID（Unix）
getgid()                    // 当前组 ID（Unix）

// 文件掩码
umask()                     // 获取当前 umask
umask(0022)                 // 设置 umask，返回旧值

// 系统信息
$info = uname()             // 返回系统信息对象
$info["sysname"]            // → "linux"
$info["machine"]            // → "x86_64"
$info["version"]            // → "go1.21"
```
