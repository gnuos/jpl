# JPL 标准库文档

JPL 标准库提供了丰富的内置函数，涵盖字符串操作、数学计算、数组处理、文件 I/O、网络通信、加密解密等功能。

## 目录

- [概述](#概述)
- [基础类型操作](#基础类型操作)
- [数组操作](#数组操作)
- [函数式编程](#函数式编程)
- [字符串操作](#字符串操作)
- [数学函数](#数学函数)
- [哈希与编码](#哈希与编码)
- [加密模块](#加密模块)
- [日期时间](#日期时间)
- [文件 I/O](#文件-io)
- [HTTP 客户端](#http-客户端)
- [网络编程](#网络编程)
- [TLS/SSL](#tlsssl)
- [JSON 处理](#json-处理)
- [反射 API](#反射-api)
- [系统函数](#系统函数)
- [进程控制](#进程控制)
- [调试工具](#调试工具)
- [类型检查](#类型检查)
- [类型转换](#类型转换)
- [动态执行](#动态执行)

## 概述

JPL 标准库包含约 520+ 个内置函数，分布在 35+ 个模块文件中，总计约 38,000+ 行代码。

### 模块列表

| 模块 | 文件 | 功能描述 |
|------|------|----------|
| I/O | io.go | 输入输出函数 |
| 工具 | util.go | 通用工具函数 |
| 数组 | array.go | 数组操作 |
| 字符串 | string.go | 字符串操作 |
| 数学 | math.go | 数学函数 |
| 哈希 | hash.go | 哈希和编码 |
| 加密 | crypto.go | 加密解密 |
| 日期时间 | datetime.go | 日期时间处理 |
| 文件 I/O | fileio.go | 文件系统操作 |
| HTTP | http.go | HTTP 客户端 |
| 网络 | net.go | TCP/UDP/Unix Socket |
| TLS | tls.go | TLS/SSL 加密通信 |
| JSON | json.go | JSON 序列化/反序列化 |
| 函数式 | functional.go | 函数式编程 |
| 反射 | reflect.go | 反射 API |
| 类型检查 | typecheck.go | 类型检查函数 |
| 类型转换 | typeconvert.go | 类型转换函数 |
| 求值 | eval.go | 动态执行代码 |
| 调试 | debug.go | 调试工具 |
| 错误 | error.go | 错误处理 |
| 进程 | process.go | 进程控制 |
| 系统 | system.go | 系统信息 |
| 位运算 | bitwise.go | 位操作 |
| 二进制 | binary.go | 二进制数据操作 |
| 正则 | re.go | 正则表达式 |
| 进程扩展 | process_ext.go | 进程扩展函数 |
| 压缩 | gzip.go, zlib.go, brotli.go | 压缩/解压缩 |
| 归档 | zip.go, tar.go | ZIP/TAR 归档 |
| URL | url.go | URL 处理 |
| DNS | dns.go | DNS 查询 |
| IP | ip.go | IP 地址处理 |
| 事件循环 | ev.go | 事件循环 |
| VM 函数 | vmfunc.go | 虚拟机函数 |
| 删除 | delete.go | 删除操作 |
| 常量 | const.go | 常量函数 |
| 对象解析 | object_parse.go | 安全的对象解析 |

## 基础类型操作

### I/O 函数

#### print
输出到标准输出（stdout），不换行。

```javascript
print "Hello"                    // 输出: Hello
print "The answer is" 42         // 输出: The answer is 42
print [1, 2, 3]                  // 输出: [1, 2, 3]
```

#### println
输出到标准输出（stdout），末尾自动添加换行符。

```javascript
println "Hello"                    // 输出: Hello\n
println "Line 1"
println "Line 2"                    // 每行单独输出
```

#### puts
输出到标准输出（stdout），所有值都不带引号，末尾自动换行。

```javascript
puts "Hello"                    // 输出: Hello
puts [1, 2, 3]                  // 输出: [1, 2, 3]（不带引号）
```

#### pp
Pretty Print 格式化输出对象和数组。

```javascript
pp {name: "Alice", items: [1, 2, 3]}
// 输出:
// {
//   name: "Alice",
//   items: [1, 2, 3]
// }
```

#### echo
将参数拼接为字符串返回，不输出。

```javascript
$msg = echo "Hello" "World"     // $msg = "Hello World"
$str = echo 10 20 30            // $str = "10 20 30"
```

#### format
格式化字符串，使用 %s 作为占位符。

```javascript
$str = format("Hello, %s!", "World")           // "Hello, World!"
$msg = format("Name: %s, Age: %s", "Alice", 30) // "Name: Alice, Age: 30"
```

#### assert
断言检查，条件为 false 时抛出运行时错误。

```javascript
assert($x > 0)                      // 如果 $x <= 0，抛出 "assertion failed"
assert($x > 0, "x must be positive") // 自定义错误消息
```

#### len
返回值的长度。

```javascript
len("hello")          // 5
len([1, 2, 3])        // 3
len({a: 1, b: 2})     // 2
```

#### type
返回值的类型名称。

```javascript
type(42)              // "int"
type("hello")         // "string"
type([1, 2, 3])       // "array"
type({a: 1})          // "object"
```

## 数组操作

### 修改函数

#### push
在数组末尾添加一个或多个元素。

```javascript
$arr = [1, 2, 3]
push($arr, 4, 5)       // $arr = [1, 2, 3, 4, 5]
```

#### pop
移除并返回数组末尾的元素。

```javascript
$arr = [1, 2, 3]
$last = pop($arr)      // $last = 3, $arr = [1, 2]
```

#### shift
移除并返回数组开头的元素。

```javascript
$arr = [1, 2, 3]
$first = shift($arr)   // $first = 1, $arr = [2, 3]
```

#### unshift
在数组开头添加一个或多个元素。

```javascript
$arr = [2, 3, 4]
unshift($arr, 0, 1)    // $arr = [0, 1, 2, 3, 4]
```

#### splice
删除或替换数组中的元素。

```javascript
$arr = [1, 2, 3, 4, 5]
splice($arr, 1, 2)              // 删除 2 个元素，从索引 1 开始，$arr = [1, 4, 5]
splice($arr, 1, 0, [2, 3])      // 插入元素，$arr = [1, 2, 3, 4, 5]
```

### 查询函数

#### indexOf
返回元素在数组中首次出现的索引。

```javascript
indexOf([1, 2, 3, 2], 2)         // 1
indexOf([1, 2, 3], 4)            // -1
```

#### lastIndexOf
返回元素在数组中最后出现的索引。

```javascript
lastIndexOf([1, 2, 3, 2], 2)     // 3
```

#### includes
检查数组是否包含指定元素。

```javascript
includes([1, 2, 3], 2)           // true
includes([1, 2, 3], 4)           // false
```

#### in_array
includes 的别名。

```javascript
in_array([1, 2, 3], 2)           // true
```

### 属性函数

#### count
返回数组长度。

```javascript
count([1, 2, 3])                 // 3
```

#### sizeof
count 的别名。

#### array_key_exists
检查数组或对象是否存在指定的键。

```javascript
array_key_exists({a: 1, b: 2}, "a")  // true
array_key_exists({a: 1, b: 2}, "c")  // false
```

#### key_exists
array_key_exists 的别名。

#### array_values
返回对象的所有值组成的数组。

```javascript
array_values({a: 1, b: 2})      // [1, 2]
```

### 计算函数

#### array_sum
计算数组元素的和。

```javascript
array_sum([1, 2, 3, 4])         // 10
```

#### array_product
计算数组元素的乘积。

```javascript
array_product([1, 2, 3, 4])     // 24
```

#### array_min
返回数组中的最小值。

```javascript
array_min([3, 1, 4, 1, 5])      // 1
```

#### array_max
返回数组中的最大值。

```javascript
array_max([3, 1, 4, 1, 5])      // 5
```

### 操作函数

#### slice
返回数组的子数组。

```javascript
slice([1, 2, 3, 4, 5], 1, 3)   // [2, 3]
slice([1, 2, 3, 4, 5], 2)       // [3, 4, 5]
```

#### array_reverse
反转数组。

```javascript
array_reverse([1, 2, 3])         // [3, 2, 1]
```

#### flat
展平嵌套数组。

```javascript
flat([1, [2, [3, 4]], 5])       // [1, 2, [3, 4], 5]
flat([1, [2, [3, 4]], 5], 2)     // [1, 2, 3, 4, 5]
```

#### unique
移除数组中的重复元素。

```javascript
unique([1, 2, 2, 3, 3, 3])      // [1, 2, 3]
```

#### array_merge
合并多个数组。

```javascript
array_merge([1, 2], [3, 4])      // [1, 2, 3, 4]
```

#### array_diff
返回数组的差集。

```javascript
array_diff([1, 2, 3], [2, 3, 4]) // [1]
```

#### array_intersect
返回数组的交集。

```javascript
array_intersect([1, 2, 3], [2, 3, 4]) // [2, 3]
```

#### array_copy
复制数组。

```javascript
$arr = [1, 2, 3]
$copy = array_copy($arr)          // [1, 2, 3]
```

### 其他函数

#### range
生成包含指定范围元素的数组。

```javascript
range(1, 5)                      // [1, 2, 3, 4, 5]
range(0, 10, 2)                  // [0, 2, 4, 6, 8]
```

#### array_fill
用值填充数组。

```javascript
array_fill(0, 5, 10)             // [10, 10, 10, 10, 10]
```

#### array_flip
交换数组的键和值。

```javascript
array_flip({a: 1, b: 2})         // {1: "a", 2: "b"}
```

#### usort
使用自定义比较函数对数组排序。

```javascript
$arr = [3, 1, 4, 1, 5]
usort($arr, ($a, $b) -> $a - $b)  // [1, 1, 3, 4, 5]
```

#### array_column
从多维数组/对象数组中提取单列。

```javascript
$users = [
    {name: "Alice", age: 30},
    {name: "Bob", age: 25}
]
array_column($users, "name")       // ["Alice", "Bob"]
```

#### array_chunk
将数组分割为多个指定大小的块。

```javascript
array_chunk([1, 2, 3, 4, 5], 2)   // [[1, 2], [3, 4], [5]]
array_chunk([1, 2, 3, 4, 5], 3)   // [[1, 2, 3], [4, 5]]
```

## 函数式编程

### 映射和过滤

#### map
对数组的每个元素应用函数，返回新数组。

```javascript
map([1, 2, 3], ($x) -> $x * 2)    // [2, 4, 6]
map(["a", "b", "c"], ($x) -> toUpper($x))  // ["A", "B", "C"]
```

#### filter
过滤数组，返回满足条件的元素。

```javascript
filter([1, 2, 3, 4, 5], ($x) -> $x % 2 == 0)  // [2, 4]
```

#### reduce
归约数组为单个值。

```javascript
reduce([1, 2, 3, 4], ($acc, $x) -> $acc + $x, 0)  // 10
```

#### find
查找第一个满足条件的元素。

```javascript
find([1, 2, 3, 4, 5], ($x) -> $x > 3)  // 4
```

### 集合操作

#### some
检查是否有元素满足条件。

```javascript
some([1, 2, 3, 4, 5], ($x) -> $x > 3)  // true
```

#### every
检查是否所有元素都满足条件。

```javascript
every([2, 4, 6, 8], ($x) -> $x % 2 == 0)  // true
```

#### contains
检查数组是否包含指定值。

```javascript
contains([1, 2, 3], 2)  // true
```

#### difference
返回数组的差集。

```javascript
difference([1, 2, 3], [2, 3, 4])  // [1]
```

#### union
返回数组的并集。

```javascript
union([1, 2], [3, 4])  // [1, 2, 3, 4]
```

### 切片操作

#### first
返回数组的第一个元素。

```javascript
first([1, 2, 3])  // 1
```

#### last
返回数组的最后一个元素。

```javascript
last([1, 2, 3])  // 3
```

#### take
返回数组的前 n 个元素。

```javascript
take([1, 2, 3, 4, 5], 3)  // [1, 2, 3]
```

#### drop
跳过数组的前 n 个元素。

```javascript
drop([1, 2, 3, 4, 5], 2)  // [3, 4, 5]
```

### 其他函数

#### sort
对数组排序。

```javascript
sort([3, 1, 4, 1, 5])  // [1, 1, 3, 4, 5]
```

#### unique
移除数组中的重复元素。

```javascript
unique([1, 2, 2, 3, 3, 3])  // [1, 2, 3]
```

#### partition
将数组分为两组。

```javascript
partition([1, 2, 3, 4, 5], ($x) -> $x % 2 == 0)  // [[2, 4], [1, 3, 5]]
```

#### zip
合并多个数组。

```javascript
zip([1, 2, 3], ["a", "b", "c"])  // [[1, "a"], [2, "b"], [3, "c"]]
```

#### unzip
拆分数组。

```javascript
unzip([[1, "a"], [2, "b"], [3, "c"]])  // [[1, 2, 3], ["a", "b", "c"]]
```

## 字符串操作

### 基础函数

#### strlen
返回字符串长度（UTF-8 字符数）。

```javascript
strlen("hello")           // 5
strlen("你好")            // 2
```

#### substr
截取子串。

```javascript
substr("hello", 1, 3)     // "ell"
substr("hello", 1)        // "ello"
```

#### strpos
查找子串首次出现的位置。

```javascript
strpos("hello world", "world")  // 6
strpos("hello", "x")            // -1
```

#### str_replace
替换子串。

```javascript
str_replace("hello world", "world", "JPL")  // "hello JPL"
```

#### trim
去除首尾空白。

```javascript
trim("  hello  ")         // "hello"
```

#### ltrim
去除开头空白。

```javascript
ltrim("  hello  ")        // "hello  "
```

#### rtrim
去除末尾空白。

```javascript
rtrim("  hello  ")        // "  hello"
```

#### chop
rtrim 的别名。

### 大小写转换

#### toUpper
转换为大写。

```javascript
toUpper("hello")          // "HELLO"
```

#### toLower
转换为小写。

```javascript
toLower("HELLO")          // "hello"
```

### 分割和连接

#### split
分割字符串。

```javascript
split("a,b,c", ",")       // ["a", "b", "c"]
```

#### explode
split 的别名。

#### join
连接字符串。

```javascript
join(["a", "b", "c"], ",")  // "a,b,c"
```

#### implode
join 的别名。

### 检查函数

#### startsWith
检查字符串是否以指定前缀开头。

```javascript
startsWith("hello", "he")  // true
```

#### endsWith
检查字符串是否以指定后缀结尾。

```javascript
endsWith("hello", "lo")    // true
```

### 其他函数

#### charAt
获取指定位置的字符。

```javascript
charAt("hello", 1)         // "e"
```

#### repeat
重复字符串。

```javascript
repeat("ab", 3)            // "ababab"
```

#### reverse
反转字符串。

```javascript
reverse("hello")           // "olleh"
```

#### strrev
reverse 的别名。

#### strcmp
比较字符串（区分大小写）。

```javascript
strcmp("a", "b")           // -1
strcmp("a", "a")           // 0
strcmp("b", "a")           // 1
```

#### strcasecmp
比较字符串（不区分大小写）。

```javascript
strcasecmp("A", "a")       // 0
```

#### ord
获取字符的 ASCII 码。

```javascript
ord("A")                   // 65
```

#### chr
将 ASCII 码转换为字符。

```javascript
chr(65)                    // "A"
```

#### nl2br
将换行符转换为 HTML `<br>`。

```javascript
nl2br("hello\nworld")     // "hello<br>world"
```

#### str_pad
填充字符串。

```javascript
str_pad("5", 3, "0", "STR_PAD_LEFT")  // "005"
str_pad("5", 3, " ", "STR_PAD_RIGHT")  // "5  "
```

#### str_getcsv
解析 CSV 字符串为数组。

```javascript
str_getcsv("Alice,30,Beijing")         // ["Alice", "30", "Beijing"]
str_getcsv('"Smith, John",25,NY')     // ["Smith, John", "25", "NY"]
```

#### levenshtein
计算两个字符串的编辑距离（Levenshtein 距离）。

```javascript
levenshtein("kitten", "sitting")       // 3
levenshtein("hello", "hallo")          // 1
```

#### similar_text
计算两个字符串的相似字符数，可选返回相似度百分比。

```javascript
similar_text("hello", "hallo")         // 4（相似字符数）
similar_text("hello", "world")         // 1
```

#### strtok
按分隔符逐个返回字符串片段（状态保持）。

```javascript
strtok("one,two;three", ",;")          // "one"
strtok()                               // "two"
strtok()                               // "three"
strtok()                               // null（结束）
```

#### parse_str
解析 URL 查询字符串为对象。

```javascript
$result = parse_str("name=Alice&age=30")
$result.name                           // "Alice"
$result.age                            // "30"
```

#### http_build_query
将对象或数组构建为 URL 查询字符串。

```javascript
http_build_query({name: "Alice", age: "30"})  // "name=Alice&age=30"
http_build_query(["a", "b", "c"])             // "0=a&1=b&2=c"
```

## 数学函数

### 基础运算

#### abs
绝对值。

```javascript
abs(-5)                    // 5
```

#### ceil
向上取整。

```javascript
ceil(3.14)                 // 4
```

#### floor
向下取整。

```javascript
floor(3.14)                // 3
```

#### round
四舍五入。

```javascript
round(3.14)                // 3
round(3.5)                 // 4
```

#### pow
幂运算。

```javascript
pow(2, 3)                  // 8
```

#### sqrt
平方根。

```javascript
sqrt(16)                   // 4
```

#### min
返回最小值。

```javascript
min(1, 5, 3)               // 1
```

#### max
返回最大值。

```javascript
max(1, 5, 3)               // 5
```

### 三角函数

#### sin
正弦函数。

```javascript
sin(0)                     // 0
```

#### cos
余弦函数。

```javascript
cos(0)                     // 1
```

#### tan
正切函数。

```javascript
tan(0)                     // 0
```

#### asin
反正弦函数。

```javascript
asin(0)                    // 0
```

#### acos
反余弦函数。

```javascript
acos(1)                    // 0
```

#### atan
反正切函数。

```javascript
atan(0)                    // 0
```

#### atan2
反正切函数（两个参数）。

```javascript
atan2(0, 1)                // 0
```

### 双曲函数

#### sinh
双曲正弦。

```javascript
sinh(0)                    // 0
```

#### cosh
双曲余弦。

```javascript
cosh(0)                    // 1
```

#### tanh
双曲正切。

```javascript
tanh(0)                    // 0
```

### 对数和指数

#### log
自然对数。

```javascript
log(2.718)                 // ≈ 1
```

#### log10
常用对数（以 10 为底）。

```javascript
log10(100)                 // 2
```

#### exp
指数函数。

```javascript
exp(1)                     // ≈ 2.718
```

#### pi
圆周率。

```javascript
pi()                       // ≈ 3.14159
```

### 其他函数

#### fmod
取模（浮点数）。

```javascript
fmod(5.5, 2)               // 1.5
```

#### hypot
直角三角形斜边长度。

```javascript
hypot(3, 4)                // 5
```

#### deg2rad
角度转弧度。

```javascript
deg2rad(180)               // ≈ 3.14159
```

#### rad2deg
弧度转角度。

```javascript
rad2deg(pi())              // 180
```

### 随机数

#### random
生成 0 到 1 之间的随机浮点数。

```javascript
random()                   // 0.123456789
```

#### randomInt
生成指定范围的随机整数。

```javascript
randomInt(1, 10)           // 5
```

### 进制转换

#### dechex
十进制转十六进制。

```javascript
dechex(255)                // "ff"
```

#### decoct
十进制转八进制。

```javascript
decoct(8)                  // "10"
```

#### decbin
十进制转二进制。

```javascript
decbin(10)                 // "1010"
```

#### hexdec
十六进制转十进制。

```javascript
hexdec("ff")               // 255
```

#### bindec
二进制转十进制。

```javascript
bindec("1010")             // 10
```

#### octdec
八进制转十进制。

```javascript
octdec("10")               // 8
```

#### base_convert
任意进制转换。

```javascript
base_convert("ff", 16, 2)  // "11111111"
```

### 字符串转换

#### parseInt
字符串转整数。

```javascript
parseInt("42")             // 42
parseInt("42", 16)         // 66
```

#### parseFloat
字符串转浮点数。

```javascript
parseFloat("3.14")         // 3.14
```

#### isNaN
检查是否为 NaN。

```javascript
isNaN(0/0)                 // true
```

#### isFinite
检查是否为有限数。

```javascript
isFinite(42)               // true
isFinite(1/0)              // false
```

## 哈希与编码

### 哈希函数

#### md5
计算 MD5 哈希值。

```javascript
md5("hello")               // "5d41402abc4b2a76b9719d911017c592"
```

#### sha1
计算 SHA-1 哈希值。

```javascript
sha1("hello")              // "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d"
```

#### md5_file
计算文件的 MD5 哈希值。

```javascript
md5_file("test.txt")       // 文件的 MD5 哈希值
```

#### sha1_file
计算文件的 SHA-1 哈希值。

```javascript
sha1_file("test.txt")      // 文件的 SHA-1 哈希值
```

#### crc32
计算 CRC32 校验和。

```javascript
crc32("hello")             // 907060870
```

### 编码函数

#### base64_encode
Base64 编码。

```javascript
base64_encode("hello")     // "aGVsbG8="
```

#### base64_decode
Base64 解码。

```javascript
base64_decode("aGVsbG8=")  // "hello"
```

#### bin2hex
二进制转十六进制。

```javascript
bin2hex("hello")           // "68656c6c6f"
```

#### hex2bin
十六进制转二进制。

```javascript
hex2bin("68656c6c6f")      // "hello"
```

## 加密模块

### 强哈希函数

#### sha256
计算 SHA-256 哈希值。

```javascript
sha256("hello")            // "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
```

#### sha512
计算 SHA-512 哈希值。

```javascript
sha512("hello")            // "9b71d224bd62f3785d96d46ad3ea3d73319bfbc2890caadae2dff72519673ca72323c3d99ba5c11d7c7acc6e14b8c5da0c4663475c2e5c3adef46f73bcdec043"
```

### HMAC 函数

#### hmac_sha256
计算 HMAC-SHA256。

```javascript
hmac_sha256("secret", "message")  // HMAC-SHA256 值
```

#### hmac_sha512
计算 HMAC-SHA512。

```javascript
hmac_sha512("secret", "message")  // HMAC-SHA512 值
```

### 编码函数

#### hex_encode
十六进制编码。

```javascript
hex_encode("hello")         // "68656c6c6f"
```

#### hex_decode
十六进制解码。

```javascript
hex_decode("68656c6c6f")   // "hello"
```

### 加密解密

#### aes_encrypt
AES-GCM 加密。

```javascript
aes_encrypt("plaintext", "key")  // 加密后的密文
```

#### aes_decrypt
AES-GCM 解密。

```javascript
aes_decrypt("ciphertext", "key")  // 解密后的明文
```

### 密码哈希

#### bcrypt_hash
bcrypt 密码哈希。

```javascript
bcrypt_hash("password", 10)  // bcrypt 哈希值（cost=10）
```

#### bcrypt_verify
bcrypt 密码验证。

```javascript
bcrypt_verify("password", $hash)  // true 或 false
```

### 现代加密

#### ed25519_keypair
生成 Ed25519 密钥对。

```javascript
$keys = ed25519_keypair()
// $keys = {public: "...", secret: "..."}
```

#### ed25519_sign
Ed25519 签名。

```javascript
$signature = ed25519_sign("message", $secret_key)
```

#### ed25519_verify
Ed25519 验证。

```javascript
$valid = ed25519_verify("message", $signature, $public_key)  // true 或 false
```

#### x25519_keypair
生成 X25519 密钥对。

```javascript
$keys = x25519_keypair()
// $keys = {public: "...", secret: "..."}
```

#### x25519_shared_secret
X25519 密钥交换。

```javascript
$shared = x25519_shared_secret($my_secret, $their_public)
```

### RSA 加密

#### rsa_keypair
生成 RSA 密钥对。

```javascript
$keys = rsa_keypair(2048)
// $keys = {public: "...", secret: "..."}
```

#### rsa_encrypt
RSA 加密。

```javascript
$ciphertext = rsa_encrypt("message", $public_key)
```

#### rsa_decrypt
RSA 解密。

```javascript
$plaintext = rsa_decrypt($ciphertext, $secret_key)
```

#### rsa_sign
RSA 签名。

```javascript
$signature = rsa_sign("message", $secret_key)
```

#### rsa_verify
RSA 验证。

```javascript
$valid = rsa_verify("message", $signature, $public_key)  // true 或 false
```

## 日期时间

### 基础函数

#### time
当前时间戳（秒）。

```javascript
time()                     // 当前 Unix 时间戳
```

#### microtime
当前微秒时间戳。

```javascript
microtime()                // 微秒时间戳
```

#### now
当前时间对象。

```javascript
$now = now()               // 当前时间对象
$now = now("Y-m-d H:i:s")  // 格式化的当前时间
```

#### date
格式化时间。

```javascript
date("Y-m-d H:i:s")       // "2026-04-02 14:30:00"
date("Y-m-d", time())      // "2026-04-02"
```

#### sleep
休眠指定毫秒数。

```javascript
sleep(1000)                // 休眠 1 秒
```

### 高级函数

#### getdate
返回日期信息对象。

```javascript
$info = getdate()
// $info = {
//   seconds: 0,
//   minutes: 30,
//   hours: 14,
//   mday: 2,
//   wday: 4,
//   mon: 4,
//   year: 2026,
//   yday: 92,
//   weekday: "Thursday",
//   month: "April",
//   0: 1722628200
// }
```

#### getdateof
返回指定日期的信息对象。

```javascript
$info = getdateof(time() - 86400)  // 昨天的日期信息
```

#### strftime
strftime 格式化。

```javascript
strftime("%Y-%m-%d", time())  // "2026-04-02"
```

#### gmdate
GMT 时间格式化。

```javascript
gmdate("Y-m-d H:i:s")     // GMT 时间
```

#### localtime
本地时间。

```javascript
$time = localtime()        // 本地时间对象
```

### 时间戳转换

#### mktime
生成时间戳。

```javascript
mktime(14, 30, 0, 4, 2, 2026)  // 指定日期的时间戳
```

#### gmmktime
GMT 时间戳。

```javascript
gmmktime(14, 30, 0, 4, 2, 2026)  // GMT 时间戳
```

## 文件 I/O

### 读写操作

#### read
读取文件内容。

```javascript
$content = read("test.txt")
```

#### readLines
读取文件行数组。

```javascript
$lines = readLines("test.txt")
```

#### write
写入文件。

```javascript
write("test.txt", "Hello, World!")
```

#### append
追加内容到文件。

```javascript
append("test.txt", "\nNew line")
```

#### file_get_contents
读取文件内容（PHP 风格）。

```javascript
$content = file_get_contents("test.txt")
```

#### file_put_contents
写入文件内容（PHP 风格）。

```javascript
file_put_contents("test.txt", "Hello, World!")
```

### 目录操作

#### mkdir
创建目录。

```javascript
mkdir("newdir")
```

#### mkdirAll
递归创建目录。

```javascript
mkdirAll("path/to/newdir")
```

#### rmdir
删除目录。

```javascript
rmdir("olddir")
```

#### listDir
列出目录内容。

```javascript
$files = listDir("/path/to/dir")
```

#### scandir
扫描目录。

```javascript
$files = scandir("/path/to/dir")
```

#### glob
文件通配。

```javascript
$files = glob("*.jpl")
```

### 文件信息

#### stat
文件状态。

```javascript
$info = stat("test.txt")
// $info = {
//   size: 1024,
//   mode: 33188,
//   modTime: 1722628200,
//   isDir: false,
//   isRegular: true
// }
```

#### fileSize
文件大小。

```javascript
fileSize("test.txt")       // 1024
```

#### isFile
检查是否为文件。

```javascript
isFile("test.txt")         // true
```

#### isDir
检查是否为目录。

```javascript
isDir("/path/to/dir")      // true
```

#### exists
检查文件是否存在。

```javascript
exists("test.txt")         // true
```

#### pathinfo
路径信息。

```javascript
$info = pathinfo("/path/to/test.txt")
// $info = {
//   dirname: "/path/to",
//   basename: "test.txt",
//   extension: "txt",
//   filename: "test"
// }
```

### 路径处理

#### dirname
目录部分。

```javascript
dirname("/path/to/test.txt")  // "/path/to"
```

#### basename
文件名部分。

```javascript
basename("/path/to/test.txt") // "test.txt"
```

#### extname
扩展名。

```javascript
extname("test.txt")           // ".txt"
```

#### joinPath
拼接路径。

```javascript
joinPath("/path", "to", "file.txt")  // "/path/to/file.txt"
```

#### absPath
绝对路径。

```javascript
absPath("relative/path")     // "/absolute/path"
```

#### relPath
相对路径。

```javascript
relPath("/a/b/c", "/a/d")    // "../b/c"
```

#### cwd
当前工作目录。

```javascript
$cwd = cwd()                 // "/current/working/directory"
```

### 文件权限

#### chmod
修改文件权限。

```javascript
chmod("test.txt", 0755)
```

#### is_readable
检查是否可读。

```javascript
is_readable("test.txt")      // true
```

#### is_writable
检查是否可写。

```javascript
is_writable("test.txt")     // true
```

### 文件时间

#### fileatime
最后访问时间。

```javascript
fileatime("test.txt")       // 访问时间戳
```

#### filemtime
最后修改时间。

```javascript
filemtime("test.txt")       // 修改时间戳
```

#### filectime
状态改变时间。

```javascript
filectime("test.txt")       // 状态改变时间戳
```

#### touch
更新时间或创建文件。

```javascript
touch("test.txt")           // 更新访问和修改时间
touch("newfile.txt")        // 创建新文件
```

### 文件操作

#### delete
删除文件。

```javascript
delete("test.txt")
```

#### unlink
delete 的别名。

#### copy
复制文件。

```javascript
copy("src.txt", "dst.txt")
```

#### move
移动文件。

```javascript
move("old.txt", "new.txt")
```

#### rename
move 的别名。

## HTTP 客户端

### 请求函数

#### http_get
GET 请求。

```javascript
$response = http_get("https://api.example.com/data")
```

#### http_post
POST 请求。

```javascript
$response = http_post("https://api.example.com/data", {
    json: {name: "Alice", age: 30}
})
```

#### http_put
PUT 请求。

```javascript
$response = http_put("https://api.example.com/data/1", {
    json: {name: "Bob"}
})
```

#### http_delete
DELETE 请求。

```javascript
$response = http_delete("https://api.example.com/data/1")
```

#### http_request
通用请求。

```javascript
$response = http_request("GET", "https://api.example.com/data", {
    headers: {Authorization: "Bearer token"},
    timeout: 30,
    follow_redirects: true,
    verify_ssl: true
})
```

### 选项说明

```javascript
{
    headers: {},              // 请求头
    timeout: 30,              // 超时时间（秒）
    follow_redirects: true,   // 跟随重定向
    max_redirects: 10,        // 最大重定向次数
    verify_ssl: true,         // 验证 SSL 证书
    proxy: "",                // 代理地址
    body: "",                 // 请求体
    json: {},                 // JSON 数据
    form: {},                 // 表单数据
    auth: {                   // 基本认证
        username: "",
        password: ""
    }
}
```

### 响应格式

```javascript
{
    status: 200,              // HTTP 状态码
    headers: {},              // 响应头
    body: "",                 // 响应体
    time: 0.5,                // 请求耗时（秒）
    size: 1024                // 响应大小（字节）
}
```

## 网络编程

### TCP 连接

#### net_tcp_connect
连接 TCP 服务器。

```javascript
$conn = net_tcp_connect("127.0.0.1", 8080)
```

#### net_tcp_listen
监听 TCP 端口。

```javascript
$server = net_tcp_listen("0.0.0.0", 8080)
```

#### net_tcp_accept
接受 TCP 连接。

```javascript
$client = net_tcp_accept($server)
```

### TCP 操作

#### net_send
发送数据。

```javascript
net_send($conn, "Hello, Server!")
```

#### net_recv
接收数据。

```javascript
$data = net_recv($conn, 1024)
```

#### net_close
关闭连接。

```javascript
net_close($conn)
```

### Unix Domain Socket

#### net_unix_connect
连接 Unix Domain Socket。

```javascript
$conn = net_unix_connect("/tmp/server.sock")
```

#### net_unix_listen
监听 Unix Domain Socket。

```javascript
$server = net_unix_listen("/tmp/server.sock")
```

#### net_unix_accept
接受 Unix Domain Socket 连接。

```javascript
$client = net_unix_accept($server)
```

### UDP

#### net_udp_bind
绑定 UDP 端口。

```javascript
$socket = net_udp_bind("0.0.0.0", 8080)
```

#### net_udp_sendto
发送 UDP 数据。

```javascript
net_udp_sendto($socket, "Hello", "127.0.0.1", 9999)
```

#### net_udp_recvfrom
接收 UDP 数据。

```javascript
[$data, $ip, $port] = net_udp_recvfrom($socket, 1024)
```

### 网络事件

#### net_on_read
注册可读事件回调。

```javascript
net_on_read($socket, ($sock) -> {
    $data = net_recv($sock, 1024)
    println "Received: " .. $data
})
```

#### net_on_write
注册可写事件回调。

```javascript
net_on_write($socket, ($sock) -> {
    net_send($sock, "Hello")
})
```

## TLS/SSL

### TLS 客户端

#### tls_connect
连接 TLS 服务器。

```javascript
$conn = tls_connect("example.com", 443, {
    verify: true,
    serverName: "example.com"
})
```

### TLS 服务器

#### tls_listen
监听 TLS 端口。

```javascript
$server = tls_listen(8443, "/path/to/cert.pem", "/path/to/key.pem")
```

#### tls_accept
接受 TLS 连接。

```javascript
$conn = tls_accept($server)
```

### TLS 操作

#### tls_send
发送加密数据。

```javascript
tls_send($conn, "Encrypted data")
```

#### tls_recv
接收加密数据。

```javascript
$data = tls_recv($conn, 1024)
```

#### tls_close
关闭 TLS 连接。

```javascript
tls_close($conn)
```

### mTLS（双向认证）

#### tls_connect_client_cert
使用客户端证书连接。

```javascript
$conn = tls_connect_client_cert(
    "example.com", 443,
    "/path/to/client.crt",
    "/path/to/client.key",
    "/path/to/ca.crt"
)
```

## JSON 处理

### 序列化和反序列化

#### json_encode
序列化为 JSON。

```javascript
json_encode([1, 2, 3])              // "[1,2,3]"
json_encode({name: "Alice"})        // "{\"name\":\"Alice\"}"
json_encode({name: "Alice"}, true)  // 美化输出
```

#### json_decode
解析 JSON 字符串。

```javascript
json_decode("[1, 2, 3]")             // [1, 2, 3]
json_decode("{\"name\":\"Alice\"}") // {name: "Alice"}
```

#### json_pretty
美化 JSON。

```javascript
json_pretty({name: "Alice", age: 30})
// {
//   "name": "Alice",
//   "age": 30
// }
```

### 智能数字解析

- 纯整数 → Int（范围内）或 BigInt
- 科学计数法 → Int、BigInt 或 Float
- 普通小数 → Float（精确）或 BigDecimal（不精确）

```javascript
json_decode("123")                   // Int(123)
json_decode("1e20")                  // BigInt(100000000000000000000)
json_decode("3.14")                 // Float(3.14)
```

## 反射 API

### 类型查询

#### typeof
返回类型名称。

```javascript
typeof(42)              // "int"
typeof("hello")         // "string"
typeof([1, 2, 3])       // "array"
```

### 变量操作

#### varexists
检查变量是否存在。

```javascript
$x = 42
varexists("$x")         // true
varexists("$y")         // false
```

#### getvar
获取变量值。

```javascript
$x = 42
getvar("$x")            // 42
```

#### setvar
设置变量值。

```javascript
setvar("$x", 100)
```

#### listvars
列出所有变量。

```javascript
$vars = listvars()
// ["$x", "$y", "$z", ...]
```

### 函数操作

#### listfns
列出所有函数。

```javascript
$fns = listfns()
// ["print", "println", "map", "filter", ...]
```

#### fn_exists
检查函数是否存在。

```javascript
fn_exists("print")      // true
fn_exists("unknown")    // false
```

#### getfninfo
获取函数信息。

```javascript
$info = getfninfo("map")
// {
//   name: "map",
//   paramCount: 2,
//   paramNames: ["arr", "fn"]
// }
```

#### callfn
动态调用函数。

```javascript
$result = callfn("print", "Hello")  // 输出 "Hello"
```

## 系统函数

### 磁盘空间

#### disk_free_space
可用磁盘空间。

```javascript
$free = disk_free_space("/")  // 可用空间（字节）
```

#### disk_total_space
总磁盘空间。

```javascript
$total = disk_total_space("/")  // 总空间（字节）
```

### 进程信息

#### getpid
进程 ID。

```javascript
$pid = getpid()           // 当前进程 ID
```

#### getuid
用户 ID。

```javascript
$uid = getuid()           // 当前用户 ID
```

#### getgid
组 ID。

```javascript
$gid = getgid()           // 当前组 ID
```

### 系统信息

#### uname
系统信息。

```javascript
$info = uname()
// {
//   sysname: "Linux",
//   nodename: "hostname",
//   release: "5.15.0",
//   version: "...",
//   machine: "x86_64"
// }
```

#### getHostname
主机名。

```javascript
$hostname = getHostname()  // 当前主机名
```

#### umask
文件创建掩码。

```javascript
$mask = umask()           // 当前 umask
umask(0755)               // 设置 umask
```

## 进程控制

#### exit
退出脚本。

```javascript
exit(0)                   // 正常退出
exit(1)                   // 错误退出
```

#### die
退出脚本并输出消息。

```javascript
die("Error occurred")     // 输出消息后退出
```

## 调试工具

#### errors
返回所有错误列表。

```javascript
$errs = errors()
// [
//   {message: "...", file: "...", line: 10},
//   ...
// ]
```

#### last_error
返回最后一个错误。

```javascript
$err = last_error()
// {message: "...", file: "...", line: 10}
```

#### clear_errors
清除所有错误。

```javascript
clear_errors()
```

## 类型检查

### 基础类型

#### is_null
检查是否为 null。

```javascript
is_null(null)            // true
```

#### is_bool
检查是否为布尔值。

```javascript
is_bool(true)            // true
```

#### is_int
检查是否为整数。

```javascript
is_int(42)               // true
```

#### is_float
检查是否为浮点数。

```javascript
is_float(3.14)            // true
```

#### is_string
检查是否为字符串。

```javascript
is_string("hello")       // true
```

#### is_array
检查是否为数组。

```javascript
is_array([1, 2, 3])      // true
```

#### is_object
检查是否为对象。

```javascript
is_object({a: 1})        // true
```

#### is_func
检查是否为函数。

```javascript
is_func(($x) -> $x * 2)   // true
```

### 扩展类型

#### is_numeric
检查是否为数字。

```javascript
is_numeric(42)           // true
is_numeric("42")         // true
```

#### is_scalar
检查是否为标量。

```javascript
is_scalar(42)            // true
is_scalar(null)          // false
```

#### is_empty
检查是否为空。

```javascript
is_empty([])             // true
is_empty({})             // true
is_empty("")             // true
is_empty(null)           // true
```

#### is_stream
检查是否为流。

```javascript
$stream = fopen("test.txt", "r")
is_stream($stream)        // true
```

#### is_regex
检查是否为正则表达式。

```javascript
$re = #/test/#
is_regex($re)             // true
```

#### is_bigint
检查是否为大整数。

```javascript
is_bigint(BigInt("100000000000000000000"))  // true
```

#### is_bigdecimal
检查是否为高精度小数。

```javascript
is_bigdecimal(BigDecimal("3.141592653589793"))  // true
```

## 类型转换

### 转换函数

#### intval
转换为整数。

```javascript
intval("42")              // 42
intval("42", 16)          // 66（十六进制）
intval(3.14)              // 3
intval(true)              // 1
```

#### floatval
转换为浮点数。

```javascript
floatval("3.14")          // 3.14
floatval(42)              // 42.0
```

#### strval
转换为字符串。

```javascript
strval(42)                // "42"
strval(3.14)              // "3.14"
```

#### boolval
转换为布尔值。

```javascript
boolval(1)                // true
boolval(0)                // false
boolval("")               // false
boolval("0")              // false
boolval("hello")          // true
```

## 动态执行

#### eval
动态执行 JPL 代码。

```javascript
$result = eval("1 + 2 * 3")     // 7
$code = 'return $x + $y'
$result = eval($code)            // $x + $y
```

**安全警告**: eval() 可能导致代码注入漏洞，建议使用 JSON/YAML 配置替代。

---

## 总结

JPL 标准库提供了约 300+ 个内置函数，涵盖：

- **基础类型操作**: I/O、工具、类型检查、类型转换
- **数组操作**: 40+ 个函数，支持增删改查、排序、聚合
- **函数式编程**: map、filter、reduce 等 20+ 个函数
- **字符串操作**: 70+ 个函数，支持 UTF-8
- **数学函数**: 40+ 个函数，包括三角函数、对数、进制转换
- **哈希与编码**: MD5、SHA1、Base64、CRC32
- **加密模块**: SHA-256/512、HMAC、AES-GCM、bcrypt、Ed25519、RSA
- **日期时间**: 12 个函数，支持 PHP 风格格式化
- **文件 I/O**: 35+ 个函数，支持路径处理、权限管理
- **HTTP 客户端**: 6 个函数，支持 JSON/Form/认证/重定向
- **网络编程**: 20+ 个函数，支持 TCP/UDP/Unix Socket
- **TLS/SSL**: 完整的加密通信支持
- **JSON 处理**: 智能数字解析，支持 BigInt/BigDecimal
- **反射 API**: 运行时类型检查和元编程
- **系统函数**: 磁盘空间、进程信息、系统信息
- **进程控制**: exit、die
- **调试工具**: 错误查询和清除

标准库为 JPL 语言提供了强大的后盾支持，使其能够胜任从简单脚本到复杂后端服务的各种开发任务。
