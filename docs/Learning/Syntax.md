# 📚 语法教程

本文档展示了JPL语言的语法特点和全部用法。

---

## 基础语法

### 变量与数据类型

声明定义变量不需要特殊的关键字，变量名称同时支持Go语言和PHP语言的风格，下划线开头的变量作为函数或者对象内部的私有变量。

```jpl
// 标识符命名规则（这些都是不同的变量名称）
name = "value"
$name = "value"
_name = "value"

// 定义常量（）
const SOME = "Ok"       // 全大写
const require = "node"  // 全小写
const Result = "Null"   // 首字母大写

// 基本类型
$null_val = null
$bool_val = true
$int_val = 42
$float_val = 3.14
$str_val = "Hello, JPL!"

// 复合类型
$arr = [1, 2, 3, 4, 5]
$obj = {name: "Alice", age: 30, active: true}

// 大数（任意精度）
$big_int = 123456789012345678901234567890
// 精确小数，无浮点误差
$big_decimal = 0.123456789012345678901234567890

// 动态类型
$x = 10
$x = "now I'm a string"
$x = [1, 2, 3]

// Go 风格类型转换（语法级）
$i = int("42")             // → 42
$f = float("3.14")         // → 3.14
$str = string(123)         // → "123"
$flag = bool(1)            // → true

// 嵌套转换
$val = int(float("3.7"))   // → 3

// 在表达式中使用
$sum = int("10") + 5       // → 15
```

### 运算符

基本运算符用法：

```jpl
// 算术（自动支持大数运算）
$a = 10 + 5   // 15
$b = 10 - 5   // 5
$c = 10 * 5   // 50
$d = 10 / 3   // 3.333...
$e = 10 % 3   // 1
$f = -10      // -10

// 比较
$a == $b      // 相等
$a != $b      // 不等
$a < $b       // 小于
$a > $b       // 大于
$a <= $b      // 小于等于
$a >= $b      // 大于等于

// 相似
$a =~ #/\w+/# // 正则匹配

// 逻辑
$a && $b      // 与
$a || $b      // 或
!$a           // 非

// 空值合并
$a ?? $b        // $a 为空时表达式值为 $b
$a ?? $b ?? $c  // 左结合性，支持链式求值

// 三元
$max = $a > $b ? $a : $b

// 管道运算（函数式数据流）
$result = 5 |> double()          // 正向管道：double(5) = 10
$result = 10 |> add(20)          // 正向管道：add(10, 20) = 30
$result = double() <| 7          // 反向管道：double(7) = 14
$result = add(100) <| 50         // 反向管道：add(100, 50) = 150

// 链式管道
$result = 5 |> double() |> double()  // double(double(5)) = 20
$result = double() <| double() <| 3  // double(double(3)) = 12
```

### 间接变量引用

JPL 支持使用反引号 `` ` `` 语法实现间接变量引用，类似 PHP 的可变变量或 Perl 的符号引用。

```jpl
// 基本间接引用
a = "hello"
x = "a"
puts `x      // → "hello"（先求 x 得到 "a"，再查找变量 a）

// 引用带 $ 前缀的变量名
$a = "world"
x = "$a"
puts `x      // → "world"

// 在表达式中使用
a = 10
x = "a"
y = `x + 5   // → 15
puts y

// 链式间接引用
a = "hello"
b = "a"
x = "b"
puts `x      // → "a"（`x → "b" → 查找变量 b → "a"）

// 未定义变量返回 null
x = "nonexistent"
val = `x     // → null
```

**工作原理**:
1. `` `varname`` 先求值 `varname` 变量，得到字符串
2. 再用该字符串作为变量名，在作用域中查找目标变量
3. 查找顺序：局部变量 → 全局变量 → 函数表 → 常量

**与 `eval()` 的区别**:
- `` `x`` 是语法级间接引用，仅按名称查找变量，不执行代码
- `eval()` 是运行时执行任意代码，存在安全风险
- 间接引用更安全、性能更好

### 控制流

```jpl
// if/else
if ($score >= 90) {
    puts "优秀"
} else if ($score >= 60) {
    print "及格"
} else {
    puts "不及格"
}

// while 循环
$i = 0
while ($i < 10) {
    print $i
    $i = $i + 1
}

// for 循环（类似 C）
for ($i = 0; $i < 10; $i += 1) {
    puts $i
}

// 范围语法 for...in
// 半开区间 1...10：不包含 10，值为 1,2,3,4,5,6,7,8,9
foreach ($i in 1...10) {
    puts $i
}

// 闭区间 1..=10：包含 10，值为 1,2,3,4,5,6,7,8,9,10
foreach ($i in 1..=10) {
    puts $i
}

// foreach-in 循环
// （遍历对象键值对）
foreach ($key => $val in $obj) {
    puts $key .. ": " .. $val
}

// （遍历数组元素）
foreach ($item in $arr) {
    puts $item
}

// break 和 continue
foreach ($i in $arr) {
    if ($i == 3) continue
    if ($i == 7) break
    puts $i
}

// exit - 立即终止脚本（无视 try/catch）
if (ARGC < 2) {
    exit 1  // 退出码 1
}

// die - 输出消息并终止（模仿 PHP）
if ($config == null) {
    die "配置文件未找到", 2
}
```

### 模式匹配

JPL 提供类似 Rust 风格的 `match/case` 语法，支持字面量、范围、正则表达式等多种模式。

#### 基本用法

```jpl
// 字面量匹配 — 执行单个分支后自动跳出
match ($status) {
    case 200: puts "OK"
    case 404: puts "Not Found"
    case 500: puts "Server Error"
    case _: puts "Unknown"
}

// match 表达式 — 返回结果
$result = match ($code) {
    case 200: "success"
    case 404: "not found"
    case 500: "error"
    case _: "unknown"
}

// OR 模式 — 多值匹配
$day_type = match ($day) {
    case "Saturday", "Sunday": "Weekend"
    case _: "Weekday"
}
```

#### Guard 条件与变量绑定

```jpl
// guard 条件 — 额外的匹配过滤
$grade = match ($score) {
    case $x if $x >= 90: "A"
    case $x if $x >= 80: "B"
    case $x if $x >= 70: "C"
    case $x if $x >= 60: "D"
    case _: "F"
}

// 标识符绑定 — 捕获任意值
match ($value) {
    case $x: puts "Got: " .. $x
}
```

#### 多行 case 体

case 分支支持多行语句，使用缩进或 `{ }` 块：

```jpl
// 缩进多行 — case 后的缩进语句属于该分支
match ($status) {
    case 200:
        puts "请求成功"
        $data = parse_json($body)
        process($data)
    case 404:
        puts "资源未找到"
        log_error("404: " .. $url)
    case 500:
        puts "服务器错误"
        retry($request)
    case _:
        puts "未知状态码"
}

// 块语法 — 使用 { } 包裹多行
match ($event) {
    case "click": {
        $x = get_mouse_x()
        $y = get_mouse_y()
        handle_click($x, $y)
    }
    case "keydown": {
        $key = get_key_code()
        handle_key($key)
    }
}
```

#### 正则模式匹配

使用正则字面量 `#/pattern/flags#` 在 `match/case` 中做模式匹配：

```jpl
// 正则匹配 — 子串匹配语义（需锚定时自行加 ^...$）
fn dispatch($input) {
    $input = trim($input)

    match ($input) {
        case "quit", "exit":       exit(0)
        case "help":               show_help()
        case #/^\d+(\.\d+)?$/:     handle_number($input)
        case #/^hello$/i:          puts "greeting!"
        case _:                    puts "unknown: " .. $input
    }
}

// 捕获组绑定 — 使用 as $var 提取匹配结果
match ($input) {
    // 提取 key=value 对
    case #/^set (\w+)=(.+)$/# as $m: {
        $key = $m[1]
        $val = $m[2]
        puts "set " .. $key .. " = " .. $val
    }

    // 提取日期组件
    case #/^(?P<year>\d{4})-(?P<month>\d{2})-(?P<day>\d{2})$/# as $m: {
        puts "year=" .. $m["year"]
        puts "month=" .. $m["month"]
    }

    case _: puts "no match"
}
```

`as $var` 绑定的变量结构：

| 访问方式 | 说明 |
|---------|------|
| `$m[0]` | 完整匹配 |
| `$m[1]` | 第一个捕获组 |
| `$m["name"]` | 命名捕获组 |

---

## 数据类型详解

### 字符串与字面量

JPL 支持多种字符串格式，满足不同场景的需求。

#### 字符串类型

JPL 支持四种字符串字面量形式：

| 引号类型 | 单行 | 多行 | 插值 | 说明 |
|---------|------|------|------|------|
| `'...'` | ✅ | ❌ | ❌ | 单引号字符串，纯文本 |
| `"..."` | ✅ | ❌ | ✅ | 双引号字符串，支持插值 |
| `'''...'''` | ✅ | ✅ | ❌ | 单引号多行，纯文本 |
| `"""..."""` | ✅ | ✅ | ✅ | 双引号多行，支持插值 |

#### 普通字符串

```jpl
// 单行字符串
$name = "Alice"
$path = 'C:\\Users\\Alice'
```

#### 多行字符串（Python 风格）

使用三引号创建保留换行和缩进的多行文本：

```jpl
// 单引号三引号 - 纯文本（不支持插值）
$json = '''
{
    "name": "JPL",
    "version": "1.0.0",
    "features": ["multiline", "interpolation"]
}
'''

// 双引号三引号 - 支持字符串插值
$title = "JPL Language"
$content = "高性能脚本语言"

$html = """
<!DOCTYPE html>
<html>
<head>
    <title>#{$title}</title>
</head>
<body>
    <h1>#{$title}</h1>
    <p>#{$content}</p>
</body>
</html>
"""
```

#### 字符串插值（Ruby 风格）

采用 **Ruby 风格** 的 `#{}` 语法，支持在双引号字符串中嵌入表达式：

```jpl
// 基本变量插值
$name = "World"
$greeting = "Hello, #{$name}!"

// 对象属性访问
$user = {name: "Alice", age: 30}
$msg = "User: #{$user.name}, Age: #{$user.age}"

// 数组索引访问
$arr = ["apple", "banana"]
$msg = "First: #{$arr[0]}, Last: #{$arr[-1]}"

// 算术表达式（计划中）
// $a = 10
// $b = 20
// $msg = "Sum: #{$a + $b}, Product: #{$a * $b}"

// 三元表达式（计划中）
// $score = 85
// $msg = "Score: #{$score}, Result: #{$score >= 60 ? 'Pass' : 'Fail'}"

// 链式属性访问和嵌套数组
$matrix = [[1, 2, 3], [4, 5, 6]]
$msg = "Matrix[1][2] = #{$matrix[1][2]}"

// 多行插值
$report = """
Name: #{$user.name}
Score: #{$score}
Status: #{$score >= 60 ? 'Pass' : 'Fail'}
"""

// 多插值组合
$x = 1
$y = 2
$z = 3
$msg = "#{$x} + #{$y} = #{$z}"
```

**插值支持的内容**：
- 变量：`#{$name}`
- 对象属性：`#{$user.name}`
- 数组索引：`#{$arr[0]}`, `#{$matrix[1][2]}`
- 算术运算：`#{$a + $b}`, `#{$price * (1 + $tax)}` ✅
- 三元表达式：`#{$score >= 60 ? 'Pass' : 'Fail'}` ✅
- 链式访问：`#{$company.ceo.email}`
- 字符串拼接：`#{$first .. ' ' .. $last}` ✅
- 格式化：`#{$value:.2f}`, `#{$num:05d}`, `#{$name:10s}` ✅

**注意事项**：
- 单引号普通字符串 `'...'` 不支持插值
- 单引号三引号字符串 `'''...'''` 不支持插值
- 只有双引号字符串（`"..."` 和 `"""..."""`）支持插值

**转义字符**：
所有字符串类型都支持标准转义序列：
```jpl
\n   // 换行
\t   // 制表符
\r   // 回车
\\   // 反斜杠
\'   // 单引号
\"   // 双引号
\#   // 井号（阻止 #{ 插值识别）
```

**特殊转义 - 阻止插值**:
```jpl
// 使用 \# 阻止 #{ 被识别为插值开始
$msg = "Use \#{$var} syntax"    // → "Use #{$var} syntax"

// 混合使用
$name = "World"
$msg = "Say \#{$name}, Hello #{$name}!"
// → "Say #{$name}, Hello World!"
```

#### 插值格式化

在插值表达式中使用 `:` 后跟格式说明符，可以对值进行格式化输出。格式说明符遵循 Go 的 `fmt.Sprintf` 规范：

```jpl
// 浮点数精度
$pi = 3.14159265
println "Pi: #{$pi:.2f}"       // → Pi: 3.14
println "Pi: #{$pi:.4f}"       // → Pi: 3.1416

// 整数宽度补零
$n = 42
println "Num: #{$n:05d}"       // → Num: 00042

// 字符串宽度对齐
$name = "JPL"
println "Hello #{$name:10s}!"  // → Hello        JPL!

// 带符号
$x = -3.14
println "Val: #{$x:+.2f}"      // → Val: -3.14

// 组合使用
$val = 123.456
println "Result: #{$val:010.2f}"  // → Result: 000123.46
```

常用格式说明符：

| 格式 | 说明 | 示例 |
|------|------|------|
| `.2f` | 浮点数，2 位小数 | `3.14` |
| `05d` | 整数，5 位宽，零填充 | `00042` |
| `10s` | 字符串，10 位宽，右对齐 | `       JPL` |
| `+.2f` | 浮点数，带符号 | `+3.14` |
| `.3e` | 科学计数法，3 位小数 | `3.142e+00` |

#### 字符串连接

使用 `..` 运算符连接字符串：
```jpl
$a = "Hello"
$b = "World"
$c = $a .. " " .. $b    // → "Hello World"
```

### 正则表达式

JPL 提供基于 Go RE2 引擎的正则表达式功能。支持正则字面量语法和函数式 API 两种使用方式。

#### 正则字面量

```jpl
// 字面量语法：#/pattern/flags#
$re = #/\d{3}-\d{4}/#
$re_i = #/hello/i#  // 带 flag

// =~ 匹配运算符
if ($input =~ #/\d+/) {
    puts "contains digits"
}

// 赋值给变量后复用
$phone_re = #/^\d{3}-\d{4}$/
if ($input =~ $phone_re) {
    puts "valid phone"
}
```

#### Flags

| Flag | 含义 | 示例 |
|------|------|------|
| `i` | 忽略大小写 | `#/hello/i#` 匹配 "Hello" |
| `m` | 多行模式（`^` `$` 匹配行首行尾） | `#/^line/m#` |
| `s` | `.` 匹配换行符 | `#/a.b/s#` 匹配 "a\nb" |
| `U` | 非贪婪互换 | `#/<.+>/U#` |

### 函数定义

函数是一等公民。

```jpl
// 传统函数
fn add(a, b) {
    return a + b
}

// 箭头函数（无需 fn 关键字）
$multiply = ($a, $b) -> $a * $b
$square = ($x) -> $x * $x

// 传统函数（使用 fn 关键字）
$double = fn($x) { return $x * 2 }

// 闭包
fn makeCounter() {
    $count = 0
    return () -> {
        $count = $count + 1
        return $count
    }
}

$counter = makeCounter()
puts $counter()  // 1
puts $counter()  // 2
puts $counter()  // 3

// 递归
fn fib(n) {
    return n <= 1 ? n : fib(n - 1) + fib(n - 2)
}

puts fib(10)  // 55

// static 变量 — 函数级持久化，调用间保持值
fn counter() {
    static $count = 0
    $count = $count + 1
    return $count
}

puts counter()  // 1
puts counter()  // 2
puts counter()  // 3

// 尾调用优化 — 自递归尾调用复用栈帧，无深度限制
fn sum($n, $acc) {
    if ($n <= 0) return $acc
    return sum($n - 1, $acc + $n)  // 尾调用，不会栈溢出
}

puts sum(10000, 0)  // 50005000
```

### 数组

```jpl
// 数组字面量
$arr = [1, 2, 3, 4, 5]

// 索引访问
puts $arr[0]    // 1
puts $arr[-1]   // 5（Python 风格负索引）

// 内置方法
push($arr, 6)           // [1, 2, 3, 4, 5, 6]
$last = pop($arr)       // 6, arr = [1, 2, 3, 4, 5]
$first = shift($arr)    // 1, arr = [2, 3, 4, 5]
unshift($arr, 0)        // [0, 2, 3, 4, 5]

// 数组函数式编程
$nums = [1, 2, 3, 4, 5]
$doubles = map($nums, ($x) -> $x * 2)        // [2, 4, 6, 8, 10]
$evens = filter($nums, ($x) -> $x % 2 == 0) // [2, 4]
$sum = reduce($nums, ($acc, $x) -> $acc + $x, 0)  // 15

// 查找
found = find($nums, ($x) -> $x > 3)         // 4
hasEven = some($nums, ($x) -> $x % 2 == 0)  // true
allPositive = every($nums, ($x) -> $x > 0) // true

// 排序和变换
$sorted = sort($nums, ($a, $b) -> $b - $a)  // [5, 4, 3, 2, 1]（降序）
$reversed = array_reverse($nums)        // [5, 4, 3, 2, 1]
$unique = unique([1, 2, 2, 3, 3, 3])    // [1, 2, 3]
```

### 对象

```jpl
// 对象字面量
$person = {
    name: "Alice",
    age: 30,
    hobbies: ["reading", "coding"]
}

// 属性访问
puts $person.name      // "Alice"
puts $person["age"]    // 30

// 动态键
$key = "name"
puts $person[$key]     // "Alice"

// 遍历
foreach ($k => $v in $person) {
    puts $k .. " = " .. $v
}

// 合并
$defaults = {theme: "dark", lang: "en"}
$settings = {lang: "zh"}
$merged = $defaults + $settings  // {theme: "dark", lang: "zh"}

// @member 闭包成员访问
// 在对象字面量的闭包方法内，使用 @ 访问当前对象的成员
// @ 绑定到最近一层的对象（静态作用域）

$counter = {
    count: 0,
    increment: () -> {
        @count = @count + 1
        return @count
    },
    get: () -> {
        return @count
    }
}

println $counter.get()       // → 0
println $counter.increment() // → 1
println $counter.increment() // → 2

$person = {
    name: "Alice",
    age: 30,
    greet: () -> {
        return "Hello, I'm " .. @name .. ", age " .. @age
    },
    birthday: () -> {
        @age = @age + 1
        return "Now I'm " .. @age
    }
}

println $person.greet()     // → Hello, I'm Alice, age 30
println $person.birthday()  // → Now I'm 31
```

#### 对象解析（安全）

```jpl
// parse_object() 安全解析对象字面量字符串
// 只解析字面量，拒绝函数调用和表达式

$config = parse_object("{host: 'localhost', port: 3306}")
puts $config.host         // → "localhost"

// 支持嵌套
$data = parse_object("{user: {name: 'John', age: 30}}")
puts $data.user.name      // → "John"

// 对比：eval() 危险 vs parse_object() 安全
// $obj = eval("{x: delete_all_files()}")        // 危险！会执行代码
// $obj = parse_object("{x: delete_all_files()}") // 安全！拒绝函数调用
```

### 预设常量

| 常量 | 值 | 说明 |
|------|------|------|
| `INF` | +∞ | 正无穷 |
| `NaN` | NaN | 非数字 |
| `PI` | 3.141592653589793 | 圆周率 |
| `TAU` | 6.283185307179586 | 2π |
| `E` | 2.718281828459045 | 自然常数 |
| `SQRT2` | 1.4142135623730951 | √2 |
| `LN2` | 0.6931471805599453 | ln(2) |
| `LN10` | 2.302585092994046 | ln(10) |
| `EOL` | `\n` 或 `\r\n` | 平台换行符（Unix: `\n`, Windows: `\r\n`）|
| `STDIN` | 流资源 | 标准输入流（可读） |
| `STDOUT` | 流资源 | 标准输出流（可写） |
| `STDERR` | 流资源 | 标准错误流（可写） |

```jpl
PI * 2                      // → 6.283185...
TAU                         // → 6.283185...
SQRT2 * SQRT2               // → 2.0
EOL                         // → "\n" (Linux/macOS) 或 "\r\n" (Windows)

// 使用 EOL 写入多行文件
$content = "Line 1" + EOL + "Line 2" + EOL
write("output.txt", $content)

// 标准流 IO 操作
$line = fgets(STDIN)         // 从标准输入读取一行
fwrite(STDOUT, "hello")      // 输出到标准输出
fwrite(STDERR, "error")      // 输出到标准错误

// print/println 默认输出到 STDOUT
print "message"
// print(STDERR, "error")    // 显式指定流参数
```

---

## 高级语法

### 异常处理

#### try/catch/throw 语法

```jpl
// 基本用法
try {
    throw "something went wrong";
} catch ($e) {
    println($e);            // → something went wrong
}

// 抛出错误对象
try {
    throw error("not found", 404, "HttpError");
} catch ($e) {
    println($e.message);    // → not found
    println($e.code);       // → 404
    println($e.type);       // → HttpError
}
```

#### 错误消息与源码上下文

运行时错误会自动显示行号和源码上下文，便于快速定位问题：

```
runtime error at line 3: something went wrong
   1 | fn greet() {
   2 |     $msg = "hello"
 → 3 |     throw "something went wrong"
   4 | }
   5 | 
```

箭头 `→` 标记了出错代码行，并显示前后各 2 行上下文。

#### 条件捕获 (when)

```jpl
// 只捕获特定错误码
try {
    throw error("db error", 1, "DBError");
} catch ($e when $e.code == 404) {
    // 处理 404
} catch ($e when $e.type == "DBError") {
    // 处理数据库错误
} catch ($e) {
    // 处理其他错误
}

// 条件不匹配时自动 re-throw 到外层
try {
    try {
        throw error("inner", 404);
    } catch ($e when $e.code == 500) {
        // 不匹配，re-throw
    }
} catch ($e) {
    println($e.code);       // → 404
}
```

### 管道运算符（函数式数据流）

JPL 支持管道运算符，使函数式编程风格的数据流处理更加清晰。

#### 正向管道 `|>`（左结合）

左侧值作为函数的**首个参数**：

```jpl
fn double(x) { return x * 2 }
fn add(a, b) { return a + b }

// 基本用法
5 |> double()           // = double(5) = 10
10 |> add(20)           // = add(10, 20) = 30

// 链式调用
5 |> double() |> double()  // = double(double(5)) = 20
```

#### 反向管道 `<|`（右结合）

右侧值作为函数的**末尾参数**：

```jpl
double() <| 7           // = double(7) = 14
add(100) <| 50          // = add(100, 50) = 150

// 链式调用
double() <| double() <| 3  // = double(double(3)) = 12
```

#### 实际应用：数据处理管道

```jpl
// 复杂数据处理管道
$data = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]

$result = $data
  |> filter(fn($x) { return $x > 5 })
  |> map(fn($x) { return $x * 10 })
  |> reduce(0, fn($acc, $x) { return $acc + $x })
// = 60 + 70 + 80 + 90 + 100 = 400

// 字符串处理
$text = "hello world"
$processed = $text
  |> split(" ")
  |> map(fn($word) { return upper($word) })
  |> join("-")
// = "HELLO-WORLD"
```

#### 注意事项

- 管道到函数引用（不带括号）返回函数本身，不调用：`5 |> double` 返回函数引用
- 管道到函数调用（带括号）才会执行：`5 |> double()` 调用并返回 10
- 正向管道左结合：`a |> f |> g` = `(a |> f) |> g` = `g(f(a))`
- 反向管道右结合：`f <| g <| a` = `f <| (g <| a)` = `f(g(a))`

#### Range + 管道 + 函数式

范围表达式可以直接与管道和函数式方法结合，支持惰性加载（使用 Go 1.25 iter.Seq2）：

```jpl
// Range + 箭头函数
$result = 1...5 |> map(($x) -> $x * 2)      // [2, 4, 6, 8]
$result = 1...10 |> filter(($x) -> $x % 2 == 0)  // [2, 4, 6, 8]
$result = 1...5 |> reduce(($acc, $x) -> $acc + $x, 0)  // 15

// Range + Lambda 函数
$result = map(1...5, fn($x) { return $x * 2; })  // [2, 4, 6, 8]

// 闭区间范围
$result = 1..=5 |> map(($x) -> $x * 2)  // [2, 4, 6, 8, 10]

// 更多函数式操作（全部支持 Range）
$result = 1...10 |> find(($x) -> $x > 5)           // 6
$result = 1...5 |> some(($x) -> $x == 3)            // true
$result = 1...5 |> every(($x) -> $x > 0)            // true
$result = 1...5 |> contains(3)                     // true
$result = 1...5 |> first()                          // 1
$result = 1...5 |> last()                           // 4
$result = 1...5 |> take(2)                          // [1, 2]
$result = 1...5 |> drop(2)                          // [3, 4]
$result = 1...5 |> sum()                            // 10
$result = 1...5 |> arrayMin()                       // 1
$result = 1...5 |> arrayMax()                       // 4
$result = 1...5 |> size()                           // 5

// 链式操作
$result = 1...10 |> filter(($x) -> $x % 2 == 0) |> map(($x) -> $x * 2)  // [4, 8, 12, 16, 20]

// 组合多个 Range
$result = union(1...3, 5...7)  // [1, 2, 3, 5, 6, 7]
$result = difference(1...5, 3...7)  // [1, 2]
$result = zip(1...3, 4...6)  // [[1, 4], [2, 5], [3, 6]]
```

### 模块化编程

### import 语句

```jpl
// 导入本地模块
import "./utils.jpl"
import "./math" as math  // 别名
from "./strings" import split, join, trim

// 导入标准库
import "math"                           // 加载模块，创建 math 命名空间

// 使用导入的模块
$result = math.pow(2, 10)
$nums = arrays.range(1, 100)
println "计算完成"

import "https://example.com/lib.jpl";   // URL 导入（带磁盘缓存 + 锁文件）
import "https://cdn.example.com/v2/lib.jpl" as lib;  // URL + 别名

from "math" import sqrt, abs;           // 选择性导入到全局
sqrt(16)                                // → 4.0

import "utils" as u;                    // 别名导入
u.helper()
```

### include 语句

```jpl
include "utils.jpl";                    // 执行文件，定义注入当前作用域（每次执行）
include_once "config.jpl";              // 只执行一次（缓存）
```

### 模块搜索路径

| 优先级 | 路径 | 说明 |
|--------|------|------|
| 1 | 脚本同目录 | 相对路径优先 |
| 2 | `jpl_modules/` | 项目根目录下 |
| 3 | `~/.jpl/modules/` | 用户级模块 |
| 4 | URL | 仅 import 支持 |

### 标准库模块

内置函数同时注册为全局函数和模块导出：

| 模块 | 全局用法 | 模块用法 |
|------|----------|----------|
| math | `sqrt(16)` | `import "math"; math.sqrt(16)` |
| strings | `strlen("hi")` | `import "strings"; strings.strlen("hi")` |
| arrays | `push(arr, 1)` | `import "arrays"; arrays.push(arr, 1)` |
| io | `print("hi")` | `import "io"; io.print("hi")` |
| hash | `md5("text")` | `import "hash"; hash.md5("text")` |

---

## 标准库用法

这里选择常用的一些场景列出标准库内置函数的使用示例。

### 正则处理

```jpl
import "re"

// 检查是否匹配
if (re_match("\\d+", "abc123")) {
    println("Contains numbers")
}

// 查找邮箱
$email = re_search("[\\w.-]+@[\\w.-]+\\.\\w+", "Contact: john@example.com")
// → "john@example.com"

// 查找所有数字
$numbers = re_findall("\\d+", "Room 101, Floor 5")
// → ["101", "5"]

// 替换
$text = re_sub("\\d+", "[NUM]", "Room 101, Floor 5")
// → "Room [NUM], Floor [NUM]"

// 分割
$parts = re_split("\\s*,\\s*", "apple, banana ,orange")
// → ["apple", "banana", "orange"]

// 命名捕获组
$groups = re_groups("(?P<year>\\d{4})-(?P<month>\\d{2})", "2024-03")
// → {year: "2024", month: "03", 0: "2024-03", ...}
```

### 二进制处理

```jpl
// 创建 Buffer（默认大端/网络字节序）
$buf = buffer_new()

// 或者从现有字节数据创建
$data = [0x48, 0x65, 0x6C, 0x6C, 0x6F]  // "Hello"
$buf = buffer_new_from($data, "big")

// 写入协议头（支持有符号/无符号整数）
buffer_write_uint32($buf, 0x48454C50)  // Magic: "HELP"
buffer_write_int16($buf, -1)           // 有符号整数
buffer_write_uint16($buf, 1)            // Version
buffer_write_uint16($buf, 0x0001)        // Flags
buffer_write_string($buf, "payload")

// 打包二进制数据
$header = pack("NNS", $len, $type, $crc)   // 格式: N=4字节大端, S=2字节大端
$data = unpack("N", $header)              // 解包

// 网络发送
net_send($fd, buffer_to_string($buf))
```

**pack/unpack 格式字符**:
- `C` - 1 字节无符号
- `S/s` - 2 字节（大端/小端）
- `N/V` - 4 字节（大端/小端）
- `Q/q` - 8 字节（大端/小端）
- `f/d` - float/double
- `a/Z` - 字符串（空填充/零结尾）


### 进程控制

JPL 提供完整的系统进程管理 API，支持命令执行、环境变量、进程控制等功能。

```jpl
// 执行命令
$output = exec("ls -la")              // 返回输出字符串
$code = system("ping -c 1 8.8.8.8")   // 返回退出码
$output = shell_exec("whoami")         // 返回完整输出

// 环境变量
$home = getenv("HOME")                 // 获取环境变量
setenv("APP_ENV", "production")        // 设置环境变量
putenv("DEBUG=true")                   // KEY=VALUE 格式

// 进程信息
$pid = getpid()                        // 当前进程 ID
$ppid = getppid()                      // 父进程 ID
$user = getlogin()                     // 登录用户名
$host = hostname()                     // 主机名
$tmp = tmpdir()                        // 临时目录

// 进程管理
$proc = spawn("sleep", ["10"])         // 创建子进程（不等待）
kill($proc.pid, 9)                     // 发送 SIGKILL
$code = waitpid($proc)                 // 等待进程结束

// 进程管道
$proc = proc_open("sort", {stdout: "pipe"})
$code = proc_wait($proc)

// 多进程并发
$procs = []
for ($i = 0; $i < 3; $i = $i + 1) {
    push($procs, spawn("echo 'Task #{$i}'"))
}
for ($proc of $procs) {
    waitpid($proc)
}
```

**模块导入**: `import "process"`

| 函数 | 说明 |
|------|------|
| `exec($cmd)` | 执行命令，返回输出字符串 |
| `system($cmd)` | 执行命令，返回退出码 |
| `shell_exec($cmd)` | 执行 shell 命令 |
| `getenv($name)` | 获取环境变量 |
| `setenv($name, $val)` | 设置环境变量 |
| `spawn($cmd)` | 创建子进程 |
| `kill($pid, $sig)` | 发送信号 |
| `waitpid($proc)` | 等待子进程 |
| `fork()` | 创建子进程（Unix） |
| `pipe()` | 创建管道对 |

### 网络编程

JPL 提供完整的网络编程能力，支持 TCP、Unix Domain Socket 和 UDP。

#### TCP 服务器/客户端

```jpl
// TCP Echo 服务器
$server = net_tcp_listen("0.0.0.0", 8080)
$registry = ev_registry_new()

$registry.on_accept($server, fn($client) {
    $registry.on_read($client, fn($fd) {
        $data = net_recv($fd, 1024)
        if (empty($data)) {
            net_close($fd)
            return
        }
        net_send($fd, "Echo: " .. $data)
    })
})

$loop = ev_loop_new()
ev_attach($loop, $registry)
ev_run($loop)
```

#### Unix Domain Socket

```jpl
// 服务器
$server = net_unix_listen("/tmp/myapp.sock")
$client = net_unix_accept($server)
$data = net_recv($client, 1024)
net_send($client, "Response: " .. $data)
net_close($client)
net_close($server)

// 客户端
$conn = net_unix_connect("/tmp/myapp.sock")
net_send($conn, "Hello")
$response = net_recv($conn, 1024)
net_close($conn)
```

#### UDP

```jpl
// UDP 服务器
$udp = net_udp_bind("0.0.0.0", 5353)
[$data, $from, $port] = net_udp_recvfrom($udp, 512)
net_udp_sendto($udp, "Response", $from, $port)
net_close($udp)

// UDP 客户端
$client = net_udp_bind("0.0.0.0", 0)
net_udp_sendto($client, "Query", "1.1.1.1", 53)
[$response] = net_udp_recvfrom($client, 512)
```

#### DNS 解析

```jpl
// 解析域名
$ips = dns_resolve("example.com")           // ["93.184.216.34", "2606:2800:220:1:248:1893:25c8:1946"]
$ip = dns_resolve_one("example.com")        // 单个 IP
$ipv4s = dns_resolve_v4("example.com")      // 仅 IPv4
$ipv6s = dns_resolve_v6("example.com")      // 仅 IPv6

// 获取详细记录
$records = dns_get_records("example.com")
for ($rec of $records) {
    print $rec.type .. ": " .. $rec.ip
}
```

#### TLS/SSL 加密连接

JPL 提供完整的 TLS/SSL 加密通信能力，支持标准 HTTPS、自签名证书、双向认证（mTLS）。

```jpl
import "tls"

// 标准 HTTPS 连接
$conn = tls_connect("api.example.com", 443)
tls_send($conn, "GET / HTTP/1.1\r\nHost: api.example.com\r\n\r\n")
$response = tls_recv($conn, 4096)
tls_close($conn)

// 自签名证书（开发/测试环境）
$conn = tls_connect("internal.dev", 8443, {verify: false})

// 使用自定义 CA
$conn = tls_connect("internal.company.com", 443, {
    ca_file: "/etc/ssl/company-ca.crt"
})

// 双向认证 (mTLS)
$conn = tls_connect("secure.example.com", 443, {
    cert_file: "/path/to/client.crt",
    key_file: "/path/to/client.key",
    ca_file: "/path/to/ca.crt"
})

// 生成自签名证书
$paths = tls_gen_cert({
    bits: 4096,
    days: 365,
    common_name: "My Server"
})
println("Cert: " + $paths.cert_path)
println("Key: " + $paths.key_path)

// TLS 服务端
$server = tls_listen(8443, "/path/to/server.crt", "/path/to/server.key")
$client = tls_accept($server)
$data = tls_recv($client, 1024)
tls_send($client, "Response: " .. $data)
tls_close($client)
tls_close($server)
```

#### HTTP Client

JPL 提供高级 HTTP 客户端，支持 HTTPS、JSON、Form、认证、超时、重定向等。

```jpl
import "http"

// 简单 GET
$resp = http_get("https://api.example.com/users")
$users = $resp.json()

// POST JSON
$resp = http_post("https://api.example.com/users", {
    json: {name: "Alice", email: "alice@example.com"}
})
println("Created: " + $resp.json().id)

// POST Form 数据
$resp = http_post("https://api.example.com/login", {
    form: {username: "user", password: "pass"}
})

// 带认证和自定义头
$resp = http_request("GET", "https://api.example.com/private", {
    headers: {
        "Authorization": "Bearer " + $token,
        "X-API-Key": "secret-key"
    },
    timeout: 30
})

// 检查响应
if ($resp.status == 200) {
    println("Success: " + $resp.body)
} else {
    println("Error " .. $resp.status .. ": " .. $resp.status_text)
}

// 下载文件
$resp = http_get("https://example.com/file.zip")
file_put_contents("download.zip", $resp.body)
println("Downloaded " .. $resp.content_length .. " bytes in " .. $resp.time .. " seconds")
```

### 加密模块 (crypto)

JPL 提供 Hash、HMAC、AES 加密和编码功能。

```jpl
import "crypto"

// SHA-256 哈希
$hash = crypto.sha256("Hello World")
// → "a591a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e"

// HMAC-SHA256
$signature = crypto.hmac_sha256("secret_key", "message")

// Hex 编解码
$hex = crypto.hex_encode("Hello")      // → "48656c6c6f"
$text = crypto.hex_decode("48656c6c6f") // → "Hello"

// AES-256-GCM 加密（自动处理 IV）
$key = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
$encrypted = crypto.aes_encrypt("Secret data", $key)
$decrypted = crypto.aes_decrypt($encrypted, $key)
// → "Secret data"
```

### 事件循环（Event Loop）

JPL 提供基于 goroutine + context 的高性能事件循环框架，支持网络 IO、定时器、信号处理。

**架构设计**:
- **Event Loop**: 管理 context 生命周期、定时器、信号
- **Registry**: 通用事件注册表，支持任意事件类型
- **各模块**: 自己管理事件处理 goroutine（net、fileio 等）

**基本用法**:

```jpl
// 1. 创建注册表
$registry = ev_registry_new()

// 2. 注册事件处理器
$registry.on_timer(1000000, fn() {
    puts "Tick"  // 每 1 秒执行
})

$registry.on_signal(2, fn() {
    puts "Ctrl+C pressed"
    ev_stop($loop)
})

// 3. 创建并运行循环
$loop = ev_loop_new()
ev_attach($loop, $registry)
ev_run($loop)
```

**网络应用示例**（TCP Echo 服务器）:

```jpl
$server = net_tcp_listen("0.0.0.0", 8080)
$registry = ev_registry_new()

// 处理新连接（回调签名：fn($client)）
$registry.on_accept($server, fn($client) {
    $peer = net_getpeername($client)
    println "Connect from: #{$peer.ip}:#{$peer.port}"
    
    // 处理客户端数据（回调签名：fn($socket, $data)）
    $registry.on_read($client, fn($socket, $data) {
        if (empty($data)) {
            net_close($socket)
            return
        }
        net_send($socket, "Echo: " .. $data)
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

**Event Loop API**:

| 函数 | 说明 | 示例 |
|------|------|------|
| `ev_registry_new()` | 创建事件注册表 | `$r = ev_registry_new()` |
| `ev_loop_new()` | 创建事件循环 | `$l = ev_loop_new()` |
| `ev_attach(loop, reg)` | 附加注册表 | `ev_attach($l, $r)` |
| `ev_run(loop)` | 运行循环（阻塞） | `ev_run($l)` |
| `ev_stop(loop)` | 停止循环 | `ev_stop($l)` |
| `$r.on_timer(us, fn)` | 周期性定时器 | `$r.on_timer(1000000, fn() {...})` |
| `$r.on_timer_once(us, fn)` | 一次性定时器 | `$r.on_timer_once(3000000, fn() {...})` |
| `$r.on_signal(sig, fn)` | 信号处理 | `$r.on_signal(2, fn() {...})` |
| `$r.on_accept(srv, fn)` | 接受连接 | `$r.on_accept($srv, fn($client) {...})` |
| `$r.on_read(sock, fn)` | 可读事件 | `$r.on_read($sock, fn($s, $data) {...})` |
| `$r.on_write(sock, fn)` | 可写事件 | `$r.on_write($sock, fn($s) {...})` |
| `$r.off(source)` | 注销事件 | `$r.off($socket)` |
| `ev_timer_now()` | 微秒级时间 | `$now = ev_timer_now()` |

**模块导入方式**:
```jpl
import "ev"
$loop = ev.loop_new()
ev.run($loop)
```

**完整文档**: [Event Loop 指南](docs/EVENT_LOOP_GUIDE.md)
**示例代码**: [network/tcp-echo-server.jpl](examples/network/tcp-echo-server.jpl)

### 异步文件 IO

JPL 提供 Python asyncio 风格的异步文件操作 API，避免阻塞事件循环。

**基本用法**:

```jpl
// 异步读取文件
file_get_async("data.txt", fn($data) {
    println "Got: #{$data}"
})

// 异步写入文件
file_put_async("output.txt", "Hello!", fn($success) {
    println "Write: #{$success}"
})

// 逐行读取大文件
file_read_lines("big.txt", fn($line) {
    process($line)
}, fn() {
    println "Done"
})

// 批量读取
file_get_batch(["a.txt", "b.txt", "c.txt"], fn($results) {
    for ($r of $results) { println $r }
})
```

**异步文件 IO API**:

| 函数 | 说明 | 示例 |
|------|------|------|
| `file_get_async(path, cb)` | 异步读取文本 | `file_get_async("f.txt", fn($d) {...})` |
| `file_put_async(path, data, cb)` | 异步写入文本 | `file_put_async("f.txt", "data", fn() {...})` |
| `file_append_async(path, data, cb)` | 异步追加 | `file_append_async("f.txt", "data", fn() {...})` |
| `file_get_bytes(path, cb)` | 异步读取二进制 | `file_get_bytes("img.png", fn($buf) {...})` |
| `file_put_bytes(path, buf, cb)` | 异步写入二进制 | `file_put_bytes("out.bin", $buf, fn() {...})` |
| `file_read_lines(path, onLine, onDone)` | 逐行读取 | `file_read_lines("f.txt", fn($l) {...}, fn() {...})` |
| `file_read_chunks(path, size, onChunk, onDone)` | 分块读取 | `file_read_chunks("f.txt", 4096, fn($c) {...}, fn() {...})` |
| `file_get_batch(paths, cb)` | 批量读取 | `file_get_batch(["a.txt", "b.txt"], fn($r) {...})` |
| `file_put_batch(items, cb)` | 批量写入 | `file_put_batch([{path:"a",data:"b"}], fn() {...})` |
| `file_parallel(ops, cb)` | 并行操作 | `file_parallel([{op:"read",path:"a"}], fn($r) {...})` |
| `file_with_lock(path, cb)` | 文件锁 | `file_with_lock("f.txt", fn($lock) {...})` |

**模块导入方式**（参考 Python asyncio）:
```jpl
import "asyncio"
asyncio.file_get("data.txt", fn($data) { ... })
asyncio.read_lines("big.txt", fn($line) { ... }, fn() { ... })
```

---

## 附录A：IEEE 754 除零行为

除法和取模运算遵循 IEEE 754 标准：

| 表达式 | 结果 |
|--------|------|
| `5 / 0` | `INF` |
| `-5 / 0` | `-INF` |
| `0 / 0` | `NaN` |
| `5 % 0` | `NaN` |

```jpl
x = 1 / 0;          // → INF
is_float(x)         // → true
x = 0 / 0;          // → NaN
isNaN(x)            // → true
```

---

## 附录B：魔术常量

JPL 支持 PHP 风格的魔术常量，它们在编译时被替换为对应的值。

### 编译时魔术常量

| 常量 | 类型 | 说明 | 示例值 |
|------|------|------|--------|
| `__FILE__` | string | 当前编译的源文件名 | `"/home/user/script.jpl"` |
| `__DIR__` | string | 当前源文件所在目录 | `"/home/user"` |
| `__LINE__` | int | 当前代码行号 | `42` |
| `__TIME__` | string | 编译时间 | `"15:04:05"` |
| `__DATE__` | string | 编译日期 | `"Jan 2 2006"` |
| `__OS__` | string | 操作系统名称 | `"linux"` / `"darwin"` / `"windows"` |
| `JPL_VERSION` | string | JPL 版本号 | `"1.0.0"` |

### 运行时魔术常量

| 常量 | 类型 | 说明 | 示例值 |
|------|------|------|--------|
| `ARGV` | array | 命令行参数数组（包含脚本名） | `["script.jpl", "arg1", "arg2"]` |
| `ARGC` | int | 命令行参数数量（包含脚本名） | `3` |

**说明**：
- `ARGV[0]` 始终是脚本文件名
- `ARGV[1]` 开始是传递给脚本的实际参数
- `ARGC` 是 `ARGV` 数组的长度（包含脚本名）

### 平台预设常量

| 常量 | 类型 | 说明 | 值 |
|------|------|------|-----|
| `EOL` | string | 平台换行符 | `\n` (Unix) / `\r\n` (Windows) |

### 使用示例

```jpl
// 获取文件位置信息
print "Running: " + __FILE__
print "In directory: " + __DIR__
print "At line: " + __LINE__

// 跨平台文件写入
$data = "Line 1" + EOL + "Line 2" + EOL
write("output.txt", $data)

// 版本检查
if (JPL_VERSION != "1.0.0") {
    print "Warning: Different JPL version"
}

// 命令行参数处理
if (ARGC < 2) {
    print "Usage: " + ARGV[0] + " <input_file>"
    exit 1
}

$inputFile = ARGV[1]
print "Processing: " + $inputFile
```

### 注意事项

1. **编译时魔术常量**（`__FILE__`, `__DIR__`, `__LINE__` 等）在编译时确定
2. **运行时魔术常量**（`ARGV`, `ARGC`）在执行时从命令行获取
3. `ARGV` 和 `ARGC` 是**只读常量**，任何修改尝试都会导致错误
4. `__LINE__` 指的是编译时所在源代码的行号
5. `__TIME__` 和 `__DATE__` 使用编译时刻的时间
6. `EOL` 根据编译时的目标平台确定（不是运行时）

### 平台常量

| 常量 | 类型 | 值 | 说明 |
|------|------|-----|------|
| `EOL` | string | `"\n"` / `"\r\n"` | 平台换行符（Unix: `\n`, Windows: `\r\n`）|
| `STDIN` | stream | 流资源 | 标准输入流（可读） |
| `STDOUT` | stream | 流资源 | 标准输出流（可写） |
| `STDERR` | stream | 流资源 | 标准错误流（可写） |

### 标准 IO 流常量设计说明

**当前实现**：
- STDIN/STDOUT/STDERR 作为**流资源类型**实现
- 类型为 `TypeStream`，包含底层 io.Reader/io.Writer
- 在 `buildin/const.go` 中通过 `RegisterPresetConstants()` 注册
- 实现位置：`engine/stream.go` - 流值类型和构造函数

**设计目的**：
1. **IO 操作**：支持真正的读写操作（fread/fwrite/fgets 等）
2. **统一接口**：与其他流（文件流、管道流、socket）共享同一接口
3. **扩展性**：为 pipe 和 socket 标准库预留接口
4. **兼容性**：与 PHP 流资源设计理念保持一致

**使用示例**：

```jpl
// 从标准输入读取
$line = fgets(STDIN)           // 读取一行
$data = fread(STDIN, 1024)     // 读取 1024 字节

// 输出到标准流
fwrite(STDOUT, "hello\n")      // 输出到标准输出
fwrite(STDERR, "error\n")      // 输出到标准错误

// print/println 默认输出到 STDOUT
print "message"
println(STDERR, "error")       // 显式指定流参数

// 文件流操作
$f = fopen("data.txt", "r")   // 打开文件
$content = fread($f, 100)      // 读取内容
fclose($f)                     // 关闭文件
```

**技术说明**：
- 流类型在运行时创建，包含实际的 IO 句柄
- 支持 is_stream() 类型检查
- 流关闭后操作会返回错误
