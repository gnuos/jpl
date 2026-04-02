# JPL 标准库模块索引

本文档按模块分类列出 JPL 标准库的所有函数。

## 目录

- [I/O 函数](#io-函数)
- [工具函数](#工具函数)
- [数组操作](#数组操作)
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
- [函数式编程](#函数式编程)
- [反射 API](#反射-api)
- [类型检查](#类型检查)
- [类型转换](#类型转换)
- [系统函数](#系统函数)
- [进程控制](#进程控制)
- [调试工具](#调试工具)
- [动态执行](#动态执行)
- [位运算](#位运算)
- [二进制操作](#二进制操作)
- [正则表达式](#正则表达式)
- [压缩解压](#压缩解压)
- [归档操作](#归档操作)
- [URL 处理](#url-处理)
- [DNS 查询](#dns-查询)
- [IP 地址](#ip-地址)
- [事件循环](#事件循环)
- [虚拟机函数](#虚拟机函数)
- [对象解析](#对象解析)

## I/O 函数

文件: `pkg/stdlib/io.go`

### 全局函数
- `print(...args)` - 输出到 stdout（无换行）
- `println(...args)` - 输出到 stdout（带换行）
- `puts(...args)` - 输出到 stdout（不带引号，换行）
- `pp(...args)` - Pretty Print 格式化输出
- `echo(...args)` - 拼接参数为字符串返回
- `format(template, ...args)` - 格式化字符串（%s 占位符）
- `assert(condition, message?)` - 断言检查

### 文件操作
- `fopen(path, mode?)` - 打开文件流
- `fread(stream, length?)` - 从流读取数据
- `fgets(stream)` - 从流读取一行
- `fwrite(stream, data)` - 向流写入数据
- `fclose(stream)` - 关闭流
- `feof(stream)` - 检查流是否结束
- `fflush(stream)` - 刷新流缓冲区
- `stream_get_meta_data(stream)` - 获取流元数据
- `is_readable(path|stream)` - 检查是否可读
- `is_writable(path|stream)` - 检查是否可写

## 工具函数

文件: `pkg/stdlib/util.go`

- `len(value)` - 返回值长度
- `type(value)` - 返回值类型名称

## 数组操作

文件: `pkg/stdlib/array.go`

### 修改函数
- `push(arr, ...items)` - 在末尾添加元素
- `pop(arr)` - 移除并返回末尾元素
- `shift(arr)` - 移除并返回开头元素
- `unshift(arr, ...items)` - 在开头添加元素
- `splice(arr, offset, length?, replacement?)` - 删除或替换元素

### 查询函数
- `indexOf(arr, value)` - 返回元素首次出现的索引
- `lastIndexOf(arr, value)` - 返回元素最后出现的索引
- `includes(arr, value)` - 检查是否包含元素
- `in_array(arr, value)` - includes 的别名

### 属性函数
- `count(arr)` - 返回数组长度
- `sizeof(arr)` - count 的别名
- `array_key_exists(obj, key)` - 检查键是否存在
- `key_exists(obj, key)` - array_key_exists 的别名
- `array_values(obj)` - 返回对象的所有值

### 计算函数
- `array_sum(arr)` - 计算数组元素的和
- `array_product(arr)` - 计算数组元素的乘积
- `array_min(arr)` - 返回最小值
- `array_max(arr)` - 返回最大值

### 操作函数
- `slice(arr, start, end?)` - 返回子数组
- `array_reverse(arr)` - 反转数组
- `flat(arr, depth?)` - 展平嵌套数组
- `unique(arr)` - 移除重复元素
- `array_merge(...arrs)` - 合并多个数组
- `array_diff(...arrs)` - 返回差集
- `array_intersect(...arrs)` - 返回交集
- `array_copy(arr)` - 复制数组

### 其他函数
- `range(start, end?, step?)` - 生成范围数组
- `array_fill(start, count, value)` - 用值填充数组
- `array_fill_keys(keys, value)` - 用值填充键
- `array_flip(arr)` - 交换键和值
- `usort(arr, cmp_fn)` - 自定义排序

### 迭代器函数
- `key(arr)` - 返回当前键
- `current(arr)` - 返回当前值
- `each(arr)` - 返回当前键值对并前进
- `next(arr)` - 前进到下一个元素
- `prev(arr)` - 后退到上一个元素
- `end(arr)` - 移动到末尾元素
- `reset(arr)` - 重置到开头元素
- `extract(arr, prefix?)` - 提取变量到当前作用域
- `array_map(arr, fn)` - 映射数组
- `array_walk(arr, fn)` - 遍历数组并应用函数

## 字符串操作

文件: `pkg/stdlib/string.go`

### 基础函数
- `strlen(str)` - 返回字符串长度（UTF-8）
- `substr(str, start, length?)` - 截取子串
- `strpos(haystack, needle, offset?)` - 查找子串首次出现位置
- `str_replace(str, search, replace)` - 替换子串
- `trim(str, chars?)` - 去除首尾空白
- `ltrim(str, chars?)` - 去除开头空白
- `rtrim(str, chars?)` - 去除末尾空白
- `chop(str)` - rtrim 的别名

### 大小写转换
- `toUpper(str)` - 转换为大写
- `toLower(str)` - 转换为小写

### 分割和连接
- `split(str, delim, limit?)` - 分割字符串
- `explode(delim, str, limit?)` - split 的别名
- `join(arr, delim)` - 连接字符串
- `implode(delim, arr)` - join 的别名

### 检查函数
- `startsWith(str, prefix)` - 检查前缀
- `endsWith(str, suffix)` - 检查后缀

### 其他函数
- `charAt(str, index)` - 获取指定位置字符
- `repeat(str, count)` - 重复字符串
- `reverse(str)` - 反转字符串
- `strrev(str)` - reverse 的别名

### 字符串比较
- `strcmp(str1, str2)` - 比较字符串（区分大小写）
- `strcasecmp(str1, str2)` - 比较字符串（不区分大小写）
- `strncmp(str1, str2, length)` - 比较前 n 个字符（区分大小写）
- `strncasecmp(str1, str2, length)` - 比较前 n 个字符（不区分大小写）

### 查找函数
- `stripos(haystack, needle, offset?)` - 不区分大小写查找
- `strrpos(haystack, needle, offset?)` - 查找最后出现位置
- `strripos(haystack, needle, offset?)` - 不区分大小写查找最后出现位置
- `strstr(haystack, needle)` - 查找子串
- `stristr(haystack, needle)` - 不区分大小写查找子串
- `strchr(haystack, needle)` - strstr 的别名

### 格式化
- `sprintf(format, ...args)` - 格式化字符串（返回）
- `printf(format, ...args)` - 格式化字符串（输出）
- `vsprintf(format, args)` - 格式化字符串（数组参数，返回）
- `vprintf(format, args)` - 格式化字符串（数组参数，输出）
- `number_format(num, decimals?, dec_point?, thousands_sep?)` - 格式化数字

### 其他
- `ord(char)` - ASCII 码
- `chr(code)` - ASCII 字符
- `nl2br(str)` - 换行转 HTML `<br>`
- `bin2hex(str)` - 二进制转十六进制
- `hex2bin(str)` - 十六进制转二进制

### 字符串增强
- `substr_compare(str1, str2, offset, length?, case_insensitive?)` - 比较子串
- `substr_count(haystack, needle, offset?, length?)` - 计算子串出现次数
- `str_repeat(str, count)` - repeat 的别名
- `str_pad(str, length, pad_str?, pad_type?)` - 填充字符串
- `str_split(str, length?)` - 拆分字符串为字符数组
- `htmlspecialchars(str)` - HTML 特殊字符转义
- `htmlspecialchars_decode(str)` - HTML 特殊字符反转义
- `strip_tags(str)` - 移除 HTML/PHP 标签
- `wordwrap(str, width?, break?, cut?)` - 自动换行
- `chunk_split(str, chunklen?, end?)` - 分块字符串

## 数学函数

文件: `pkg/stdlib/math.go`

### 基础运算
- `abs(x)` - 绝对值
- `ceil(x)` - 向上取整
- `floor(x)` - 向下取整
- `round(x)` - 四舍五入
- `pow(x, y)` - 幂运算
- `sqrt(x)` - 平方根
- `min(...values)` - 最小值
- `max(...values)` - 最大值

### 三角函数
- `sin(x)` - 正弦
- `cos(x)` - 余弦
- `tan(x)` - 正切
- `asin(x)` - 反正弦
- `acos(x)` - 反余弦
- `atan(x)` - 反正切
- `atan2(y, x)` - 反正切（两个参数）

### 双曲函数
- `sinh(x)` - 双曲正弦
- `cosh(x)` - 双曲余弦
- `tanh(x)` - 双曲正切

### 对数和指数
- `log(x)` - 自然对数
- `log10(x)` - 常用对数
- `exp(x)` - 指数函数
- `pi()` - 圆周率

### 其他函数
- `fmod(x, y)` - 浮点数取模
- `hypot(x, y)` - 直角三角形斜边长度
- `deg2rad(deg)` - 角度转弧度
- `rad2deg(rad)` - 弧度转角度

### 随机数
- `random()` - 0 到 1 之间的随机浮点数
- `randomInt(min, max)` - 指定范围的随机整数
- `rand_str(length?, chars?)` - 生成随机字符串
- `getrandmax()` - 最大随机数

### 进制转换
- `dechex(num)` - 十进制转十六进制
- `decoct(num)` - 十进制转八进制
- `decbin(num)` - 十进制转二进制
- `hexdec(str)` - 十六进制转十进制
- `bindec(str)` - 二进制转十进制
- `octdec(str)` - 八进制转十进制
- `base_convert(num, frombase, tobase)` - 任意进制转换

### 字符串转换
- `parseInt(str, base?)` - 字符串转整数
- `parseFloat(str)` - 字符串转浮点数
- `isNaN(x)` - 检查是否为 NaN
- `isFinite(x)` - 检查是否为有限数

## 哈希与编码

文件: `pkg/stdlib/hash.go`

### 哈希函数
- `md5(str)` - MD5 哈希
- `sha1(str)` - SHA1 哈希
- `md5_file(path)` - 文件 MD5 哈希
- `sha1_file(path)` - 文件 SHA1 哈希
- `crc32(str)` - CRC32 校验和

### 编码函数
- `base64_encode(str)` - Base64 编码
- `base64_decode(str)` - Base64 解码
- `bin2hex(str)` - 二进制转十六进制
- `hex2bin(str)` - 十六进制转二进制

## 加密模块

文件: `pkg/stdlib/crypto.go`

### 强哈希函数
- `sha256(str)` - SHA256 哈希
- `sha512(str)` - SHA512 哈希
- `sha256_file(path)` - 文件 SHA256 哈希
- `sha512_file(path)` - 文件 SHA512 哈希

### HMAC 函数
- `hmac_sha256(key, msg)` - HMAC-SHA256
- `hmac_sha512(key, msg)` - HMAC-SHA512

### 编码函数
- `hex_encode(str)` - 十六进制编码
- `hex_decode(str)` - 十六进制解码

### 加密解密
- `aes_encrypt(plaintext, key, iv?)` - AES-GCM 加密
- `aes_decrypt(ciphertext, key)` - AES-GCM 解密

### 密码哈希
- `bcrypt_hash(password, cost?)` - bcrypt 密码哈希
- `bcrypt_verify(password, hash)` - bcrypt 密码验证

### Ed25519 签名
- `ed25519_keypair()` - 生成 Ed25519 密钥对
- `ed25519_sign(msg, secret_key)` - Ed25519 签名
- `ed25519_verify(msg, signature, public_key)` - Ed25519 验证

### X25519 密钥交换
- `x25519_keypair()` - 生成 X25519 密钥对
- `x25519_shared_secret(my_secret, their_public)` - X25519 密钥交换

### RSA 加密
- `rsa_keypair(bits)` - 生成 RSA 密钥对
- `rsa_encrypt(msg, public_key)` - RSA 加密
- `rsa_decrypt(ciphertext, secret_key)` - RSA 解密
- `rsa_sign(msg, secret_key)` - RSA 签名
- `rsa_verify(msg, signature, public_key)` - RSA 验证

### ECC
- `ecdh_keypair(curve)` - 生成 ECDH 密钥对
- `ecdh_shared_secret(my_private, their_public)` - ECDH 密钥交换

## 日期时间

文件: `pkg/stdlib/datetime.go`

### 基础函数
- `time()` - 当前时间戳（秒）
- `microtime(get_as_float?)` - 当前微秒时间戳
- `now(format?)` - 当前时间对象
- `date(format, timestamp?)` - 格式化时间
- `sleep(ms)` - 休眠毫秒数

### 高级函数
- `getdate(timestamp?)` - 返回日期信息对象
- `getdateof(timestamp)` - 返回指定日期的信息对象
- `strftime(format, timestamp?)` - strftime 格式化
- `gmdate(format, timestamp?)` - GMT 时间格式化
- `localtime(timestamp?)` - 本地时间对象
- `gettimeofday()` - 时间信息对象（类似 C 的 gettimeofday）

### 时间戳转换
- `mktime(hour, minute, second, month, day, year)` - 生成时间戳
- `gmmktime(hour, minute, second, month, day, year)` - GMT 时间戳

## 文件 I/O

文件: `pkg/stdlib/fileio.go`

### 读写操作
- `read(path)` - 读取文件内容
- `readLines(path)` - 读取文件行数组
- `write(path, content)` - 写入文件
- `append(path, content)` - 追加内容到文件
- `file_get_contents(path, use_include_path?, context?, offset?, maxlen?)` - 读取文件内容（PHP 风格）
- `file_put_contents(path, data, flags?, context?)` - 写入文件内容（PHP 风格）

### 目录操作
- `mkdir(path, mode?, recursive?)` - 创建目录
- `mkdirAll(path)` - 递归创建目录
- `rmdir(path, context?)` - 删除目录
- `listDir(path)` - 列出目录内容
- `scandir(path, sorting_order?, context?)` - 扫描目录
- `glob(pattern, flags?)` - 文件通配

### 文件信息
- `stat(path)` - 文件状态
- `fileSize(path)` - 文件大小
- `isFile(path)` - 检查是否为文件
- `isDir(path)` - 检查是否为目录
- `exists(path)` - 检查文件是否存在
- `pathinfo(path, options?)` - 路径信息

### 路径处理
- `dirname(path)` - 目录部分
- `basename(path, suffix?)` - 文件名部分
- `extname(path)` - 扩展名
- `joinPath(...parts)` - 拼接路径
- `absPath(path)` - 绝对路径
- `relPath(target, base)` - 相对路径
- `realpath(path)` - 规范化绝对路径
- `cwd()` - 当前工作目录

### 文件权限
- `chmod(path, mode)` - 修改文件权限
- `chown(path, user, group?)` - 修改文件所有者
- `is_readable(path)` - 检查是否可读
- `is_writable(path)` - 检查是否可写
- `is_executable(path)` - 检查是否可执行

### 文件时间
- `fileatime(path)` - 最后访问时间
- `filemtime(path)` - 最后修改时间
- `filectime(path)` - 状态改变时间
- `touch(path, time?, atime?)` - 更新时间或创建文件

### 文件操作
- `delete(path)` - 删除文件
- `unlink(path)` - delete 的别名
- `copy(source, dest, context?)` - 复制文件
- `move(source, dest)` - 移动文件
- `rename(oldname, newname, context?)` - move 的别名

### 异步文件 I/O
- `async_read(path, callback)` - 异步读取文件
- `async_write(path, content, callback)` - 异步写入文件

## HTTP 客户端

文件: `pkg/stdlib/http.go`

### 请求函数
- `http_get(url, options?)` - GET 请求
- `http_post(url, options?)` - POST 请求
- `http_put(url, options?)` - PUT 请求
- `http_delete(url, options?)` - DELETE 请求
- `http_request(method, url, options?)` - 通用请求

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

## 网络编程

文件: `pkg/stdlib/net.go`

### TCP 连接
- `net_tcp_connect(host, port)` - 连接 TCP 服务器
- `net_tcp_listen(host, port)` - 监听 TCP 端口
- `net_tcp_accept(server)` - 接受 TCP 连接

### TCP 操作
- `net_send(conn, data)` - 发送数据
- `net_recv(conn, size)` - 接收数据
- `net_close(conn)` - 关闭连接

### Unix Domain Socket
- `net_unix_connect(path)` - 连接 Unix Domain Socket
- `net_unix_listen(path)` - 监听 Unix Domain Socket
- `net_unix_accept(server)` - 接受 Unix Domain Socket 连接

### UDP
- `net_udp_bind(host, port)` - 绑定 UDP 端口
- `net_udp_sendto(socket, data, host, port)` - 发送 UDP 数据
- `net_udp_recvfrom(socket, size)` - 接收 UDP 数据

### 网络事件
- `net_on_read(socket, callback)` - 注册可读事件回调
- `net_on_write(socket, callback)` - 注册可写事件回调
- `net_on_close(socket, callback)` - 注册关闭事件回调

## TLS/SSL

文件: `pkg/stdlib/tls.go`

### TLS 客户端
- `tls_connect(host, port, options?)` - 连接 TLS 服务器
- `tls_connect_client_cert(host, port, cert, key, ca?)` - 使用客户端证书连接

### TLS 服务器
- `tls_listen(port, cert, key, options?)` - 监听 TLS 端口
- `tls_accept(server)` - 接受 TLS 连接

### TLS 操作
- `tls_send(conn, data)` - 发送加密数据
- `tls_recv(conn, size)` - 接收加密数据
- `tls_close(conn)` - 关闭 TLS 连接

### TLS 选项
```javascript
{
    verify: true,              // 验证证书
    serverName: "",            // 服务器名称（SNI）
    minVersion: "",            // 最小 TLS 版本
    maxVersion: "",            // 最大 TLS 版本
    cipherSuites: []           // 密码套件列表
}
```

## JSON 处理

文件: `pkg/stdlib/json.go`

### 序列化和反序列化
- `json_encode(value, pretty?)` - 序列化为 JSON
- `json_decode(json_str)` - 解析 JSON 字符串
- `json_pretty(value)` - 美化 JSON

## 函数式编程

文件: `pkg/stdlib/functional.go`

### 映射和过滤
- `map(arr, fn)` - 对每个元素应用函数
- `filter(arr, fn)` - 过滤数组
- `reduce(arr, fn, initial?)` - 归约数组
- `find(arr, fn)` - 查找第一个满足条件的元素

### 集合操作
- `some(arr, fn)` - 检查是否有元素满足条件
- `every(arr, fn)` - 检查是否所有元素都满足条件
- `contains(arr, value)` - 检查数组是否包含指定值
- `difference(...arrs)` - 返回差集
- `union(...arrs)` - 返回并集
- `intersection(...arrs)` - 返回交集

### 切片操作
- `first(arr)` - 返回第一个元素
- `last(arr)` - 返回最后一个元素
- `take(arr, n)` - 返回前 n 个元素
- `drop(arr, n)` - 跳过前 n 个元素
- `head(arr, n?)` - take 的别名
- `tail(arr, n?)` - drop 的别名

### 其他函数
- `sort(arr, cmp?)` - 对数组排序
- `unique(arr)` - 移除重复元素
- `partition(arr, fn)` - 将数组分为两组
- `zip(...arrs)` - 合并多个数组
- `unzip(arr)` - 拆分数组

## 反射 API

文件: `pkg/stdlib/reflect.go`

### 类型查询
- `typeof(value)` - 返回类型名称

### 变量操作
- `varexists(name)` - 检查变量是否存在
- `getvar(name)` - 获取变量值
- `setvar(name, value)` - 设置变量值
- `listvars(scope?)` - 列出所有变量

### 函数操作
- `listfns()` - 列出所有函数
- `fn_exists(name)` - 检查函数是否存在
- `getfninfo(name)` - 获取函数信息
- `callfn(name, ...args)` - 动态调用函数

## 类型检查

文件: `pkg/stdlib/typecheck.go`

### 基础类型
- `is_null(value)` - 检查是否为 null
- `is_bool(value)` - 检查是否为布尔值
- `is_int(value)` - 检查是否为整数
- `is_float(value)` - 检查是否为浮点数
- `is_string(value)` - 检查是否为字符串
- `is_array(value)` - 检查是否为数组
- `is_object(value)` - 检查是否为对象
- `is_func(value)` - 检查是否为函数

### 扩展类型
- `is_numeric(value)` - 检查是否为数字
- `is_scalar(value)` - 检查是否为标量
- `is_empty(value)` - 检查是否为空
- `is_stream(value)` - 检查是否为流
- `is_regex(value)` - 检查是否为正则表达式
- `is_bigint(value)` - 检查是否为大整数
- `is_bigdecimal(value)` - 检查是否为高精度小数

## 类型转换

文件: `pkg/stdlib/typeconvert.go`

### 转换函数
- `intval(value, base?)` - 转换为整数
- `floatval(value)` - 转换为浮点数
- `strval(value)` - 转换为字符串
- `boolval(value)` - 转换为布尔值
- `doubleval(value)` - floatval 的别名
- `settype(value, type)` - 设置变量类型

## 系统函数

文件: `pkg/stdlib/system.go`

### 磁盘空间
- `disk_free_space(path)` - 可用磁盘空间
- `disk_total_space(path)` - 总磁盘空间

### 进程信息
- `getpid()` - 进程 ID
- `getuid()` - 用户 ID
- `getgid()` - 组 ID
- `getmyuid()` - 当前脚本用户 ID
- `getmygid()` - 当前脚本组 ID

### 系统信息
- `uname()` - 系统信息对象
- `getHostname()` - 主机名
- `php_uname(mode?)` - uname 的别名

### 其他
- `umask(mask?)` - 文件创建掩码
- `getcwd()` - 当前工作目录
- `chdir(dir)` - 改变工作目录

## 进程控制

文件: `pkg/stdlib/process.go`

- `exit(status)` - 退出脚本
- `die(message, status?)` - 退出脚本并输出消息

### 进程扩展

文件: `pkg/stdlib/process_ext.go`

- `exec(command, output?, return_var?)` - 执行命令
- `passthru(command, return_var?)` - 执行命令并输出
- `system(command, return_var?)` - 执行系统命令
- `shell_exec(command)` - 执行 shell 命令并返回输出
- `proc_open(command, descriptors, pipes, cwd?, env?, other_options?)` - 打开进程
- `proc_close(process)` - 关闭进程
- `proc_terminate(process, signal?)` - 终止进程
- `proc_get_status(process)` - 获取进程状态
- `proc_nice(priority)` - 改变进程优先级

## 调试工具

文件: `pkg/stdlib/debug.go`

- `errors()` - 返回所有错误列表
- `last_error()` - 返回最后一个错误
- `clear_errors()` - 清除所有错误

## 动态执行

文件: `pkg/stdlib/eval.go`

- `eval(code)` - 动态执行 JPL 代码

## 位运算

文件: `pkg/stdlib/bitwise.go`

- `bit_and(x, y)` - 按位与
- `bit_or(x, y)` - 按位或
- `bit_xor(x, y)` - 按位异或
- `bit_not(x)` - 按位非
- `bit_left_shift(x, n)` - 左移
- `bit_right_shift(x, n)` - 右移

## 二进制操作

文件: `pkg/stdlib/binary.go`

- `pack(format, ...values)` - 打包二进制数据
- `unpack(format, data)` - 解包二进制数据

## 正则表达式

文件: `pkg/stdlib/re.go`

### 匹配函数
- `re_match(pattern, str, flags?)` - 匹配正则表达式
- `re_match_all(pattern, str, flags?)` - 匹配所有
- `re_replace(pattern, replacement, str, flags?)` - 替换匹配
- `re_split(pattern, str, limit?, flags?)` - 分割字符串

### 正则字面量
- `#/pattern/flags#` - 正则表达式字面量

## 压缩解压

### Gzip

文件: `pkg/stdlib/gzip.go`

- `gzencode(data, level?)` - Gzip 压缩
- `gzdecode(data)` - Gzip 解压
- `gzcompress(data, level?)` - Gzip 压缩（DEFLATE）
- `gzuncompress(data)` - Gzip 解压（DEFLATE）

### Zlib

文件: `pkg/stdlib/zlib.go`

- `zlib_encode(data, level?)` - Zlib 压缩
- `zlib_decode(data)` - Zlib 解压

### Brotli

文件: `pkg/stdlib/brotli.go`

- `brotli_compress(data, quality?)` - Brotli 压缩
- `brotli_decompress(data)` - Brotli 解压

## 归档操作

### ZIP

文件: `pkg/stdlib/zip.go`

- `zip_open(filename)` - 打开 ZIP 归档
- `zip_read(zip)` - 读取 ZIP 条目
- `zip_entry_read(entry, length?)` - 读取条目内容
- `zip_entry_name(entry)` - 获取条目名称
- `zip_entry_filesize(entry)` - 获取条目大小
- `zip_close(zip)` - 关闭 ZIP 归档

### TAR

文件: `pkg/stdlib/tar.go`

- `tar_open(filename)` - 打开 TAR 归档
- `tar_read(tar)` - 读取 TAR 条目
- `tar_entry_read(entry, length?)` - 读取条目内容
- `tar_entry_name(entry)` - 获取条目名称
- `tar_entry_size(entry)` - 获取条目大小
- `tar_close(tar)` - 关闭 TAR 归档

## URL 处理

文件: `pkg/stdlib/url.go`

- `parse_url(url, component?)` - 解析 URL
- `urlencode(str)` - URL 编码
- `urldecode(str)` - URL 解码
- `rawurlencode(str)` - 原始 URL 编码
- `rawurldecode(str)` - 原始 URL 解码
- `http_build_query(data, prefix?, arg_sep?, enc_type?)` - 构建查询字符串

## DNS 查询

文件: `pkg/stdlib/dns.go`

- `dns_lookup(hostname, type?)` - DNS 查询
- `dns_get_record(hostname, type?, authns?, addtl?, raw?)` - 获取 DNS 记录
- `gethostbyname(hostname)` - 获取主机 IP 地址
- `gethostbyaddr(ip)` - 获取主机名

## IP 地址

文件: `pkg/stdlib/ip.go`

- `ip2long(ip)` - IP 地址转整数
- `long2ip(long)` - 整数转 IP 地址
- `inet_pton(address)` - 将可读地址转换为二进制格式
- `inet_ntop(in_addr)` - 将二进制地址转换为可读格式
- `is_ipv4(address)` - 检查是否为 IPv4 地址
- `is_ipv6(address)` - 检查是否为 IPv6 地址

## 事件循环

文件: `pkg/stdlib/ev.go`

- `ev_loop()` - 启动事件循环
- `ev_stop()` - 停止事件循环
- `ev_set_timeout(callback, delay)` - 设置定时器
- `ev_set_interval(callback, interval)` - 设置间隔定时器
- `ev_set_io_event(stream, callback, events)` - 设置 IO 事件
- `ev_remove_timer(timer)` - 移除定时器

## 虚拟机函数

文件: `pkg/stdlib/vmfunc.go`

- `vm_info()` - 虚拟机信息
- `vm_stats()` - 虚拟机统计
- `vm_disassemble()` - 反汇编字节码

## 对象解析

文件: `pkg/stdlib/object_parse.go`

- `object_parse(json_str)` - 安全解析 JSON 字符串为对象
- `object_validate(obj, schema)` - 验证对象结构

## 常量函数

文件: `pkg/stdlib/const.go`

- `define(name, value, case_insensitive?)` - 定义常量
- `defined(name)` - 检查常量是否定义
- `constant(name)` - 获取常量值

## 删除操作

文件: `pkg/stdlib/delete.go`

- `delete(arr, key)` - 删除数组元素
- `unset(...vars)` - 销毁变量

## 错误处理

文件: `pkg/stdlib/error.go`

- `error(message, code?)` - 创建错误对象
- `error_get_last()` - 获取最后一个错误
- `error_reporting(level?)` - 设置或获取错误报告级别

## 完整函数列表

JPL 标准库共包含约 300+ 个内置函数，分布在以下模块：

| 模块 | 函数数量 |
|------|----------|
| I/O | 18 |
| 工具 | 2 |
| 数组 | 45 |
| 字符串 | 70 |
| 数学 | 45 |
| 哈希与编码 | 8 |
| 加密 | 20 |
| 日期时间 | 12 |
| 文件 I/O | 35 |
| HTTP | 6 |
| 网络 | 15 |
| TLS | 10 |
| JSON | 3 |
| 函数式编程 | 20 |
| 反射 | 8 |
| 类型检查 | 15 |
| 类型转换 | 6 |
| 系统 | 8 |
| 进程控制 | 2 |
| 进程扩展 | 10 |
| 调试工具 | 3 |
| 动态执行 | 1 |
| 位运算 | 6 |
| 二进制操作 | 2 |
| 正则表达式 | 5 |
| 压缩解压 | 9 |
| 归档操作 | 12 |
| URL 处理 | 8 |
| DNS 查询 | 5 |
| IP 地址 | 7 |
| 事件循环 | 6 |
| 虚拟机函数 | 3 |
| 对象解析 | 2 |
| 常量函数 | 3 |
| 删除操作 | 2 |
| 错误处理 | 3 |

**总计**: ~400 个函数
