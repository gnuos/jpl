# JPL 设计决策文档

> 本文档记录项目开发过程中的重要设计决策，供后续开发参考和追溯。

---

## D1. 变量命名规则

**决策日期**：2026-03-21

| 标识符 | 含义 | 示例 |
|--------|------|------|
| `$name` | 全局/局部变量 | `$count = 1;` |
| `name` | 全局/局部变量（省略 $） | `count = 1;` |
| `_private` | 仅作用域内可访问（私有变量） | `_temp = 10;` |
| `_` | 丢弃值赋值占位符 | `_, $b = func();` |
| `$_` | 保留特殊变量（执行结果） | `echo $_;` |

**理由**：支持 `$` 前缀和无前缀两种写法，提高灵活性；`_` 和 `$_` 提供特殊用途。

---

## D2. 字符串连接运算符

**决策日期**：2026-03-21

- 使用 `..` 作为字符串连接运算符（参考 Lua）
- 示例：`"Hello" .. " " .. "World"`

**理由**：与 `+` 算术运算符区分，避免类型歧义。

---

## D3. 函数关键字

**决策日期**：2026-03-21

- `function` 和 `fn` 都可用，`fn` 是简写别名
- 示例：`fn add($a, $b) { return $a + $b; }`

**理由**：提供简洁的函数声明语法，同时保持与传统语法的兼容性。

---

## D4. 常量声明

**决策日期**：2026-03-21

- 支持 `const` 关键字：`const MAX = 100;`
- 支持 `define()` 函数（Phase 4 实现）：`define('MAX', 100);`

**理由**：提供不可变常量支持，防止意外修改。

---

## D5. 保留字面量值名称

**决策日期**：2026-03-21

| 类型 | 支持的写法 |
|------|------------|
| null | `null`、`NULL` |
| true | `true`、`TRUE`、`True` |
| false | `false`、`FALSE`、`False` |

**理由**：支持大小写变体，提高代码兼容性。

---

## D6. 箭头函数

**决策日期**：2026-03-21

- 参考 JavaScript 设计
- 支持表达式体：`$x -> $x * 2`
- 支持块体：`$x -> { return $x * 2; }`

**理由**：提供简洁的匿名函数语法，支持多行代码。

---

## D7. 模式匹配

**决策日期**：2026-03-21

- 参考 Python 3.10+ 设计
- 支持完整模式：字面量、变量绑定、OR 模式、守卫条件、数组/对象解构、通配符
- match 作为表达式或语句（可选模式）

**理由**：提供强大的模式匹配能力，同时保持灵活性。

---

## D8. 元编程

**决策日期**：2026-03-21

- 支持 `eval()` 函数：运行时执行字符串代码
- 支持反射 API：类型检查、属性遍历
- 暂不支持宏系统

**理由**：eval() 和反射满足大部分元编程需求，实现简单；宏系统复杂度高，后续根据需求再考虑。

---

## D9. 动态属性访问

**决策日期**：2026-03-21

- 方括号索引：`$obj[$key]`、`$arr[$index]`
- `$$` 动态变量名：`$$varName`（将字符串解析为变量名）

**示例**：
```jpl
// 方括号索引
$obj = {"name": "test", "value": 42};
$key = "name";
$value = $obj[$key];  // "test"

// $$ 动态变量名
$varName = "x";
$$varName = 42;  // 等价于 $x = 42
echo $x;         // 42
```

**理由**：方括号索引是标准做法；$$ 提供动态变量名能力，参考 PHP。

---

## D10. 对象成员访问

**决策日期**：2026-03-21

- 支持点符号访问：`$obj.field`
- 支持方括号访问：`$obj["field"]`
- 支持链式访问：`$obj.field.subfield`
- 支持混合访问：`$obj.array[0].field`
- JSON 字面量访问需括号包裹：`({"name": "test"}).name`

**示例**：
```jpl
// 点符号访问
$data = {"name": "test", "value": 42};
echo $data.name;   // "test"

// 链式访问
$user = {"profile": {"name": "张三"}};
echo $user.profile.name;  // "张三"

// 混合访问
$data = {"users": [{"name": "张三"}]};
echo $data.users[0].name;  // "张三"

// JSON 字面量访问
echo ({"name": "test"}).name;  // "test"
$result = {"name": "test"}.name;  // 赋值语句中可直接访问
```

**理由**：点符号直观易读；括号包裹解决语法歧义。

---

## D11. 特例函数无括号调用

**决策日期**：2026-03-22

- 6 个 I/O 类内置函数支持语句级省略括号调用（返回值有意义的函数不在此列）
- 多参数使用逗号分隔
- 非特例函数使用空格调用语法时报语法错误
- 嵌套调用或作为值传递时必须使用括号

**特例函数列表**（仅 I/O 类语句函数）：
```
echo, print, println, log, format, assert
```

**示例**：
```php
// 单参数
print "hello"
len arr
typeof x

// 多参数（逗号分隔）
format "%d %s", x, name
echo "value:", count

// 嵌套必须括号
print(len(arr))
assert(len(x) > 0)
```

**实现细节**：
- Parser 维护 `specialFuncs` 表，语句分发时匹配特例函数
- 匹配成功走 `parseSpecialCallStatement`，生成标准 `CallExpr` 节点
- Compiler/VM 无需改动，复用已有函数调用机制
- 非特例函数后跟值类型（字符串/数字/标识符）时报告语法错误

**理由**：减少 I/O 和调试函数的语法噪音，同时限制在已知内置函数范围内，避免通用空格调用带来的歧义和解析复杂度。仅 Parser 层改动，Compiler/VM 零影响。

---

## D12. 内置函数分层策略

**决策日期**：2026-03-22

内置函数分三个梯队实现，优先覆盖核心场景：

**第一梯队（已实现）**：
- I/O：print, println, echo, log, format, assert
- 工具：len
- 函数式编程：map, filter, reject, reduce, find, some, every, sort, contains, unique, partition, flattenDeep, difference, union, zip, unzip
- 类型检查：is_null, is_bool, is_int, is_float, is_string, is_array, is_object, is_func

**理由**：第一梯队覆盖 90% 实际场景；跳过 forEach（与 foreach 语句重复）。按需求追加，避免过度设计。

---

## D13. 特例函数表范围

**决策日期**：2026-03-22

特例函数表仅包含 I/O 类语句函数（print/println/echo/log/format/assert）和声明类函数（define），不包含返回值有意义的函数（len/is_*/typeof）。

**理由**：返回值有意义的函数通常在表达式中使用（`x = len(arr)`），无括号语法会导致表达式位置解析困难。I/O 类函数几乎只作为独立语句使用，无歧义。

---

## D14. 指令集扩展策略

**决策日期**：2026-03-22

### 当前指令集状态

当前实现 38 条指令，覆盖以下功能：
- 加载/存储：NOP, LOAD, LOADK, LOADNULL, LOADBOOL, GETGLOBAL, SETGLOBAL, GETVAR, SETVAR
- 算术运算：ADD, SUB, MUL, DIV, MOD, NEG
- 比较运算：EQ, NEQ, LT, GT, LTE, GTE
- 字符串：CONCAT
- 逻辑运算：AND, OR, NOT
- 数组/对象：NEWARRAY, NEWOBJECT, GETINDEX, SETINDEX, GETMEMBER, SETMEMBER
- 控制流：JMP, JMPIF, JMPIFNOT
- 函数调用：CALL, RETURN
- 闭包：CLOSURE, GETUPVAL, SETUPVAL, CLOSE_UPVALS
- 异常处理：THROW, TRY_BEGIN, TRY_END
- 其他：POP, DUP, TYPEOF

### 扩展策略

**核心原则**：按需扩展，避免过度设计。新指令仅在以下情况添加：
1. 用 Go 函数无法高效实现的功能
2. 性能测试证明存在瓶颈的热点路径
3. 语言特性必须的底层支持

### 延后实现的特性

| 特性 | 延后理由 | 触发条件 |
|------|----------|----------|
| 位运算指令 | Hash/编码用 Go 函数实现，性能更好 | 需要脚本级位运算时 |
| 迭代器指令 | 当前循环实现足够 | 性能测试发现循环瓶颈时 |
| 数学内置指令 | 用 Go 函数实现 | 科学计算场景需求时 |
| 字符串指令 | 用 Go 函数实现 | 字符串操作成为热点时 |
| 调试指令 | 基础 REPL 不需要 | 实现高级调试功能时 |
| 尾调用优化 | 非必需 | 递归深度成为问题时 |

### 任务依赖分析

后续任务与指令特性的关系：

| 任务 | 需要新指令？ | 说明 |
|------|-------------|------|
| Phase 4.3 Hash/编码 | 否 | 用 Go 函数实现 MD5/SHA1/CRC32 |
| Phase 3.3 函数重载 | 否 | 在 VM 层面实现分发 |
| Phase 3.1 GC | 否 | 基于 Go GC 实现引用计数 |
| Phase 5 REPL | 否 | 基础功能不需要新指令 |
| Phase 6 性能测试 | 可能 | 如发现瓶颈再优化 |

### 实现优先级

| 优先级 | 特性 | 触发条件 |
|--------|------|----------|
| P0 | 无 | 当前阶段不需要 |
| P1 | 位运算指令 | 需要脚本实现 Hash 算法时 |
| P1 | 调试指令 | 实现断点调试功能时 |
| P2 | FORPREP/FORLOOP | 循环性能测试不达标时 |
| P2 | INC/DEC | 计数器循环成为热点时 |
| P3 | 其他优化指令 | 特定场景需求时 |

**结论**：当前阶段（Phase 4-6）不需要添加新指令，专注于标准库和工具链完善。指令集扩展作为性能优化手段，在 Phase 6 性能测试后根据实际瓶颈决定。

---

## D15. 函数重载策略

**决策日期**：2026-03-22

### 设计

函数重载仅按参数数量区分，不按参数类型区分。

### 理由

1. JPL 是动态类型语言，参数类型在编译时未知
2. 同参数数量的类型差异由脚本侧 `typeof` 手动分发更灵活
3. 实现简单，覆盖 90% 实际场景

### 实现细节

- `funcMap` 从 `map[string]*CompiledFunction` 改为 `map[string][]*CompiledFunction`
- `findFunction(name, argCount)` 按参数数量精确匹配
- 无精确匹配时容错：返回参数最多的函数，多余参数截断，缺少参数补 null
- `OP_CLOSURE` 改为函数索引查找（避免同名函数歧义）
- 重载仅通过 `CallByName` API 生效，脚本内联调用走变量赋值模式

### 使用示例

```go
// Go 侧调用
vm.CallByName("add", engine.NewInt(10))        // 调用 1 参数版本
vm.CallByName("add", engine.NewInt(10), engine.NewInt(20))  // 调用 2 参数版本
```

```jpl
// 脚本内联调用（注意：变量赋值模式，最后声明的函数覆盖前一个）
fn format() { return "empty"; }
fn format($val) { return "value: " + $val; }
// 此时 format 变量指向 1 参数版本

// 通过 CallByName 调用仍可访问所有重载版本
```

---

## D16. REPL 界面设计

**决策日期**：2026-03-24

### 布局设计

```
┌─────────────────────────┬──────────────┐
│                         │ Variables    │
│   历史记录区域           │ x = 10       │
│   (viewport, 可滚动)    │ y = "hello"  │
│                         │ fn = <func>  │
│  语法高亮 (chroma/v2)   │ ... ↑↓ 滚动  │
│                         ├──────────────┤
├─────────────────────────┤ Keywords     │
│ > 输入框                │ print len    │
│   (textarea, 动态高度)  │ if while for │
│                         │ ... ↑↓ 滚动  │
└─────────────────────────┴──────────────┘
```

### 组件说明

| 区域 | 组件 | 说明 |
|------|------|------|
| 左上：历史记录 | `viewport` | 显示已执行的输入语句和执行结果，支持垂直滚动，语法高亮 |
| 左下：输入框 | `textarea` | 多行编辑，初始 1 行，Shift+Enter 换行动态增长，最大 40% 终端高度，超出后内部滚动，Enter 提交执行 |
| 右上：变量面板 | `viewport` | 显示用户声明的变量和常量及其值，执行后刷新，溢出时可滚动 |
| 右下：关键字面板 | `viewport` | 显示语言核心关键字和内置函数名（静态列表），溢出时可滚动 |

### 交互行为

| 操作 | 行为 |
|------|------|
| Enter | 提交当前输入执行 |
| Shift+Enter | 在输入框内换行 |
| Ctrl+C | 清空当前输入 |
| Ctrl+D / exit 命令 | 退出 REPL |
| 上下箭头 | 浏览历史输入 |
| Page Up/Down | 滚动历史记录区域 |
| Tab | 自动补全（关键字、变量名、函数名） |

### 语法高亮

- 使用 `alecthomas/chroma/v2` 进行语法高亮
- 历史记录中的代码和当前输入均实时高亮
- 需要为 JPL 语言实现自定义 Lexer（或复用类 C/JS 语法的 Lexer）

### 右侧面板数据源

- **变量面板**：`VM.GetGlobals()` API 获取全局变量快照，每次语句执行后刷新
- **关键字面板**：`token.Keywords()` + `buildin.FunctionNames()` 返回的静态列表

### 技术栈

- `charm.land/bubbletea/v2` — TUI 框架
- `charm.land/bubbles/v2` — textarea + viewport 组件
- `charm.land/lipgloss/v2` — 样式渲染（布局分栏）
- `alecthomas/chroma/v2` — 语法高亮

### 参考

- [charmbracelet/crush](https://github.com/charmbracelet/crush) — 底部输入框 + 可滚动历史区域布局

---

## D17. REPL 简化重写

**决策日期**：2026-03-26

### 背景
原 REPL 使用 Bubble Tea TUI 框架实现，虽然功能丰富（四象限布局、语法高亮、变量面板等），但存在以下问题：
- 代码量过大（1085 行），维护困难
- 依赖过多（6 个 charm 相关库）
- 架构复杂（Model-View-Update 模式）

### 新方案
使用 `github.com/elk-language/go-prompt` 库重写 REPL，实现简化的命令行交互：

| 特性 | 旧方案 | 新方案 |
|------|--------|--------|
| 框架 | Bubble Tea TUI | go-prompt |
| 代码量 | 1085 行 | 150 行 |
| 依赖数 | 6 个 charm 库 | 1 个 go-prompt |
| 布局 | 四象限 TUI | 简单命令行 |
| 语法高亮 | chroma/v2 | 暂不支持（可扩展） |
| 变量面板 | 实时显示 | `:globals`/`:vars` 指令 |
| 调试模式 | F9/F10 快捷键 | `:debug on/off` 指令 |

### 调试指令设计

所有调试功能通过 `:` 前缀指令实现：

| 指令 | 功能 |
|------|------|
| `:debug on/off` | 切换调试模式（打印执行步骤） |
| `:globals` | 显示全局变量（紧凑格式） |
| `:locals` | 显示局部变量 |
| `:vars` | 显示所有变量 |
| `:funcs` | 显示所有内置函数 |
| `:consts` | 显示预设常量 |
| `:doc <name>` | 查看函数签名 |
| `:help` | 显示帮助 |
| `:quit` | 退出 REPL |

### 调试输出示例

```bash
> :debug on
调试模式已开启

> for ($i = 0; $i < 3; $i++) { }
[EXEC] for 初始化: $i = 0
[EXEC] for 条件: 0 < 3 = true → 进入循环体
[EXEC] for 迭代: $i = 1
[EXEC] for 条件: 1 < 3 = true → 进入循环体
[EXEC] for 迭代: $i = 2
[EXEC] for 条件: 2 < 3 = true → 进入循环体
[EXEC] for 迭代: $i = 3
[EXEC] for 条件: 3 < 3 = false → 退出循环
```

### 技术实现

1. **移除依赖**：
   - `charm.land/bubbles/v2`
   - `charm.land/bubbletea/v2`
   - `charm.land/lipgloss/v2`
   - `github.com/alecthomas/chroma/v2`

2. **新增依赖**：
   - `github.com/elk-language/go-prompt`

3. **VM 扩展**：
   - 添加 `debugMode` 字段
   - 添加 `SetDebugMode/GetDebugMode` 方法
   - 在执行循环和条件分支处打印调试信息

### 决策理由

1. **简化维护**：代码量减少 86%，大幅降低维护成本
2. **减少依赖**：从 6 个库减少到 1 个库
3. **保持功能**：核心交互功能保留（补全、历史、调试）
4. **更可靠**：go-prompt 经过充分测试，稳定性高
5. **开发效率**：重写仅需 3 小时，远低于预期

### 备份与回滚

原实现已备份为：
- `cmd/jpl/repl.go.bak`
- `cmd/jpl/repl_test.go.bak`

如需回滚，可恢复备份文件并重新添加 charm 依赖。

---

## D18. 流类型系统设计

**决策日期**：2026-03-26

### 背景

当前 STDIN/STDOUT/STDERR 为纯字符串常量（`"stdin"`/`"stdout"`/`"stderr"`），仅具标识作用，无法进行真正的 IO 操作。为支持完整的流 IO 能力（包括未来的 pipe 和 socket），需引入流资源类型。

### 设计目标

1. 支持真正的流读写操作（fread/fwrite/fgets 等）
2. 预定义标准流常量（STDIN/STDOUT/STDERR）为可操作的流资源
3. 为未来 pipe 和 socket 标准库预留扩展接口
4. 保持与 PHP/Perl 设计理念的一致性

### 核心决策

#### 1. 新增值类型 `TypeStream`

```go
// ValueType 枚举扩展
const (
    // ... 现有类型
    TypeStream ValueType = 10  // 流资源类型
)
```

#### 2. 流对象结构

```go
// streamValue 实现 Value 接口
type streamValue struct {
    mode     StreamMode     // 读/写/读写
    reader   io.Reader      // 读取器（可为nil）
    writer   io.Writer      // 写入器（可为nil）
    closer   io.Closer      // 关闭器（可为nil）
    path     string         // 源标识（如 "stdin", "file.txt"）
    closed   bool           // 是否已关闭
}

// StreamMode 流模式
type StreamMode int
const (
    StreamRead      StreamMode = iota  // 只读
    StreamWrite                        // 只写
    StreamReadWrite                    // 读写
)
```

#### 3. 预定义标准流

```go
// engine/stream.go
func NewStdinStream() Value   // 包装 os.Stdin
func NewStdoutStream() Value  // 包装 os.Stdout
func NewStderrStream() Value  // 包装 os.Stderr

// buildin/const.go 注册修改
e.RegisterConst("STDIN", engine.NewStdinStream())
e.RegisterConst("STDOUT", engine.NewStdoutStream())
e.RegisterConst("STDERR", engine.NewStderrStream())
```

#### 4. IO 函数扩展

| 函数 | 说明 | 示例 |
|------|------|------|
| `fopen(path, mode)` | 打开文件流 | `$f = fopen("test.txt", "r")` |
| `fread(stream, length)` | 读取指定字节数 | `$data = fread(STDIN, 1024)` |
| `fgets(stream)` | 读取一行 | `$line = fgets(STDIN)` |
| `fwrite(stream, data)` | 写入数据 | `fwrite(STDOUT, "hello")` |
| `fclose(stream)` | 关闭流 | `fclose($f)` |
| `feof(stream)` | 是否到达末尾 | `feof($f)` |
| `fflush(stream)` | 刷新缓冲区 | `fflush(STDOUT)` |

#### 5. print/println 函数扩展

```jpl
// 现有用法保持兼容
print "hello"                    // 输出到 STDOUT（默认）
println "world"                  // 输出到 STDOUT（默认）

// 新增：显式指定流
print(STDERR, "error message")   // 输出到标准错误
println(STDOUT, "info")          // 显式输出到标准输出
```

### 与 Perl/PHP 设计对比

| 特性 | Perl | PHP | JPL |
|------|------|-----|-----|
| 类型 | 文件句柄（glob） | 流资源（resource） | TypeStream |
| 预定义 | 全局句柄 STDIN/STDOUT/STDERR | 常量资源 | 常量流值 |
| 读取 | `<FH>` / `readline` | `fgets` / `fread` | `fgets` / `fread` |
| 写入 | `print FH "text"` | `fwrite(STDOUT, "text")` | `fwrite(STDOUT, "text")` |
| 关闭 | `close(FH)` | `fclose($fh)` | `fclose($f)` |
| 重定向 | `select` / `open` | `fopen` | `fopen` + 赋值 |
| 管道 | `open(FH, "\| cmd")` | `popen()` | 未来：`pipe()` |

### 未来扩展接口

此设计为以下功能预留接口：

```go
// Pipe 支持
func NewPipe() (read, write Value, error)

// Socket 支持
func NewSocketStream(conn net.Conn) Value

// Buffer 支持（内存流）
func NewBufferStream(buf *bytes.Buffer) Value

// Tee 支持（输出复制）
func NewTeeStream(writers ...io.Writer) Value
```

### 实现步骤

| 步骤 | 文件 | 任务 |
|------|------|------|
| 1 | `engine/value.go` | 新增 `TypeStream` 枚举值 |
| 2 | `engine/stream.go`（新文件） | `streamValue` 结构体和 Value 接口实现 |
| 3 | `engine/stream.go` | 流构造函数（NewStdinStream 等） |
| 4 | `buildin/const.go` | 修改 STDIN/STDOUT/STDERR 注册为流类型 |
| 5 | `buildin/io.go` | 新增 fread/fwrite/fgets/fclose/feof/fflush 函数 |
| 6 | `buildin/io.go` | 修改 print/println 支持可选流参数 |
| 7 | `engine/value.go` | ValueType.String() 添加 TypeStream 分支 |

### 决策理由

1. **TypeStream 独立类型**：与 TypeArray/TypeObject 平级，语义清晰，便于类型检查（`is_stream($f)`）
2. **组合 io.Reader/io.Writer/io.Closer**：复用 Go 标准库，支持任意流源
3. **预定义常量改为流值**：破坏性变更（原字符串→流），但这是正确设计的必要代价
4. **PHP 风格函数命名**：fread/fwrite/fgets 与 PHP 一致，降低学习成本
5. **流模式枚举**：明确读写权限，为权限检查和错误提示提供基础

---

## D19. 网络框架设计

**决策日期**：2026-03-26
**实现日期**：2026-03-27
**状态**：✅ 已完成（Phase 9）

### 背景

用户期望在 JPL 中实现自定义 socket 网络框架（如 HTTP 服务器、MQTT broker 等）。当前标准库仅有文件 IO 和基础函数，缺少网络通信、并发处理、二进制数据处理等能力。

### 设计目标

1. 提供完整的 TCP/UDP/Unix Domain Socket 网络编程能力
2. 支持高性能并发连接处理（IO 多路复用）
3. 提供二进制协议解析能力（pack/unpack + Buffer）
4. 为未来扩展（WebSocket、TLS）预留接口

### 实现概要

**代码规模**：约 3541 行 Go 代码
- `buildin/binary.go` (1033 行)：pack/unpack + Buffer 类型
- `buildin/ev.go` (589 行)：Event Loop 核心（EvLoopValue + EvRegistryValue）
- `buildin/ev_registry.go` (595 行)：事件注册表方法
- `buildin/net.go` (968 行)：TCP/UDP/Unix Socket + NetSocketValue
- `buildin/dns.go`：DNS 解析

**函数总数**：约 60 个新函数
- 二进制处理：pack, unpack + 23 个 buffer_xxx 函数
- 事件循环：ev_loop_new, ev_registry_new, ev_attach, ev_run, ev_run_once, ev_stop, ev_is_running, ev_timer_now
- 事件注册：ev_on_read/write/accept/timer/timer_once/signal, ev_off/off_read/off_write/off_timer/off_signal, ev_clear, ev_count
- 网络：net_tcp_listen/connect/accept, net_unix_listen/connect/accept, net_udp_bind/sendto/recvfrom, net_send/recv/close, net_getsockname/getpeername/set_nonblock/is_unix
- DNS：dns_resolve, dns_resolve_one

**测试覆盖**：
- 9 个事件循环测试（ev_test.go）
- 网络集成测试（net_test.go）
- 完整网络栈集成测试（integration_test.go）

### 并发模型选择

**决策**：采用 IO 多路复用 + 回调注册表模式

| 方案 | 优点 | 缺点 | 选择 |
|------|------|------|------|
| Go 协程模型 | 编程简单 | 协程泄漏风险、内存开销大 | ❌ |
| IO 多路复用 + 回调 | 资源可控、无泄漏、高性能 | 回调嵌套 | ✅ |
| async/await 生成器 | 用户友好 | 实现复杂、需事件循环 | ❌ |

**理由**：
- 单线程处理万级连接，内存可控
- 无协程泄漏风险
- 生命周期明确（手动注册/注销）
- 与底层 OS 能力（epoll/kqueue）直接对接

### 核心设计：注册表模式

```jpl
// 注册表模式：分离"事件定义"和"循环执行"
$loop = evloop_new()                    // 创建事件循环
$handlers = registry_new()              // 创建事件注册表

// 注册事件处理器
$handlers.on_read($fd, fn($fd) { ... })
$handlers.on_write($fd, fn($fd) { ... })
$handlers.on_accept($fd, fn($client) { ... })
$handlers.on_error($fd, fn($fd, $err) { ... })
$handlers.on_timer($ms, fn() { ... })

// 附加并运行
evloop_attach($loop, $handlers)
evloop_run($loop)
```

**注册表模式优势**：
1. 集中管理所有事件处理器
2. 处理器可复用、可批量操作
3. 代码结构清晰，避免回调地狱
4. 便于调试和追踪

### API 设计

#### 1. 网络层 API

| 函数 | 说明 | 示例 |
|------|------|------|
| `tcp_listen(host, port)` | 创建监听 socket | `$server = tcp_listen("0.0.0.0", 8080)` |
| `tcp_connect(host, port)` | 连接远程主机 | `$client = tcp_connect("example.com", 80)` |
| `tcp_accept($fd)` | 接受连接 | `$conn = tcp_accept($server)` |
| `tcp_send($fd, $data)` | 发送数据 | `tcp_send($conn, "hello")` |
| `tcp_recv($fd, $len)` | 接收数据 | `$data = tcp_recv($conn, 4096)` |
| `tcp_close($fd)` | 关闭连接 | `tcp_close($conn)` |
| `tcp_peername($fd)` | 获取对端地址 | `$addr = tcp_peername($conn)` |
| `tcp_sockname($fd)` | 获取本地地址 | `$addr = tcp_sockname($conn)` |
| `tcp_set_nonblock($fd)` | 设置非阻塞 | `tcp_set_nonblock($fd)` |
| `udp_bind(host, port)` | 创建 UDP socket | `$sock = udp_bind("0.0.0.0", 5353)` |
| `udp_sendto($fd, $data, $host, $port)` | 发送 UDP 数据 | `udp_sendto($sock, $data, "1.1.1.1", 53)` |
| `udp_recvfrom($fd, $len)` | 接收 UDP 数据 | `[$data, $addr, $port] = udp_recvfrom($sock, 512)` |

#### 2. 事件循环层 API

| 函数 | 说明 | 示例 |
|------|------|------|
| `evloop_new()` | 创建事件循环 | `$loop = evloop_new()` |
| `registry_new()` | 创建事件注册表 | `$h = registry_new()` |
| `$h.on_read($fd, $cb)` | 注册可读事件 | `$h.on_read($fd, fn($fd) {...})` |
| `$h.on_write($fd, $cb)` | 注册可写事件 | `$h.on_write($fd, fn($fd) {...})` |
| `$h.on_accept($fd, $cb)` | 注册连接事件 | `$h.on_accept($server, fn($client) {...})` |
| `$h.on_error($fd, $cb)` | 注册错误事件 | `$h.on_error($fd, fn($fd, $err) {...})` |
| `$h.on_timer($ms, $cb)` | 注册定时器 | `$h.on_timer(1000, fn() {...})` |
| `$h.off($fd)` | 注销 fd 所有事件 | `$h.off($client)` |
| `$h.off_read($fd)` | 注销可读事件 | `$h.off_read($fd)` |
| `$h.count()` | 获取注册的事件数 | `$n = $h.count()` |
| `evloop_attach($loop, $h)` | 附加注册表到循环 | `evloop_attach($loop, $h)` |
| `evloop_run($loop)` | 运行事件循环 | `evloop_run($loop)` |
| `evloop_run_once($loop)` | 运行一次（非阻塞） | `evloop_run_once($loop)` |
| `evloop_stop($loop)` | 停止循环 | `evloop_stop($loop)` |

#### 3. 二进制处理层 API

**pack/unpack 函数**：

| 函数 | 说明 | 示例 |
|------|------|------|
| `pack($fmt, $val)` | 打包二进制数据 | `$bin = pack("N", 1234)` |
| `unpack($fmt, $bin)` | 解包二进制数据 | `$val = unpack("N", $bin)` |

格式字符：
| 格式 | 含义 | 大小 |
|------|------|------|
| `C` | unsigned char | 1 字节 |
| `S` | unsigned short (大端) | 2 字节 |
| `s` | unsigned short (小端) | 2 字节 |
| `N` | unsigned long (大端) | 4 字节 |
| `V` | unsigned long (小端) | 4 字节 |
| `Q` | unsigned long long (大端) | 8 字节 |
| `q` | unsigned long long (小端) | 8 字节 |
| `f` | float | 4 字节 |
| `d` | double | 8 字节 |
| `a` | 字符串（空填充） | N 字节 |
| `Z` | 字符串（零结尾） | N+1 字节 |
| `x` | 空字节 | 1 字节 |

**Buffer 对象**（29 个函数）：

| 函数 | 说明 | 示例 |
|------|------|------|
| **创建与配置** | | |
| `buffer_new([endian])` | 创建空 buffer | `$buf = buffer_new()` |
| `buffer_new_from($bytes, [endian])` | 从字节数组或字符串创建 | `$buf = buffer_new_from([0x48, 0x65, 0x6C, 0x6C, 0x6F])` |
| `buffer_set_endian($buf, $endian)` | 设置字节序（big/little） | `buffer_set_endian($buf, "little")` |
| **有符号整数写入** | 按当前字节序写入 | |
| `buffer_write_int8($buf, $v)` | 写入 8 位有符号整数（-128~127） | `buffer_write_int8($buf, -50)` |
| `buffer_write_int16($buf, $v)` | 写入 16 位有符号整数（-32768~32767） | `buffer_write_int16($buf, -1000)` |
| `buffer_write_int32($buf, $v)` | 写入 32 位有符号整数 | `buffer_write_int32($buf, -50000)` |
| **无符号整数写入** | 按当前字节序写入 | |
| `buffer_write_uint8($buf, $v)` | 写入 8 位无符号整数（0~255） | `buffer_write_uint8($buf, 255)` |
| `buffer_write_uint16($buf, $v)` | 写入 16 位无符号整数 | `buffer_write_uint16($buf, 50000)` |
| `buffer_write_uint32($buf, $v)` | 写入 32 位无符号整数 | `buffer_write_uint32($buf, 100000)` |
| **浮点数写入** | 固定大端序 | |
| `buffer_write_float32($buf, $v)` | 写入 32 位浮点数 | `buffer_write_float32($buf, 3.14)` |
| `buffer_write_float64($buf, $v)` | 写入 64 位浮点数 | `buffer_write_float64($buf, 2.718281828)` |
| **其他写入** | | |
| `buffer_write_string($buf, $s)` | 写入字符串（无编码转换） | `buffer_write_string($buf, "hello")` |
| `buffer_write_bytes($buf, $b)` | 写入字节数组 | `buffer_write_bytes($buf, [0x01, 0x02, 0x03])` |
| **有符号整数读取** | 按当前字节序读取 | |
| `buffer_read_int8($buf)` | 读取 8 位有符号整数 | `$v = buffer_read_int8($buf)` |
| `buffer_read_int16($buf)` | 读取 16 位有符号整数 | `$v = buffer_read_int16($buf)` |
| `buffer_read_int32($buf)` | 读取 32 位有符号整数 | `$v = buffer_read_int32($buf)` |
| **无符号整数读取** | 按当前字节序读取 | |
| `buffer_read_uint8($buf)` | 读取 8 位无符号整数 | `$v = buffer_read_uint8($buf)` |
| `buffer_read_uint16($buf)` | 读取 16 位无符号整数 | `$v = buffer_read_uint16($buf)` |
| `buffer_read_uint32($buf)` | 读取 32 位无符号整数 | `$v = buffer_read_uint32($buf)` |
| **浮点数读取** | 固定大端序 | |
| `buffer_read_float32($buf)` | 读取 32 位浮点数 | `$v = buffer_read_float32($buf)` |
| `buffer_read_float64($buf)` | 读取 64 位浮点数 | `$v = buffer_read_float64($buf)` |
| **其他读取** | | |
| `buffer_read_string($buf, $len)` | 读取指定长度字符串 | `$s = buffer_read_string($buf, 10)` |
| `buffer_read_bytes($buf, $len)` | 读取指定长度字节数组 | `$b = buffer_read_bytes($buf, 5)` |
| **游标操作** | | |
| `buffer_seek($buf, $offset, $whence)` | 设置读取位置 | `buffer_seek($buf, 0, 0)` // SEEK_SET |
| `buffer_tell($buf)` | 获取当前读取位置 | `$pos = buffer_tell($buf)` |
| **信息查询** | | |
| `buffer_length($buf)` | 获取缓冲区字节长度 | `$len = buffer_length($buf)` |
| **转换与重置** | | |
| `buffer_to_bytes($buf)` | 转为字节数组 | `$bytes = buffer_to_bytes($buf)` |
| `buffer_to_string($buf)` | 转为字符串 | `$str = buffer_to_string($buf)` |
| `buffer_reset($buf)` | 清空缓冲区，读取位置归零 | `buffer_reset($buf)` |
| **类型检查** | | |
| `is_buffer($value)` | 检查是否为 Buffer 类型 | `is_buffer($buf)` → true |

#### 4. DNS 层 API

| 函数 | 说明 | 示例 |
|------|------|------|
| `dns_resolve($host)` | DNS 解析（返回所有 IP） | `$ips = dns_resolve("example.com")` |
| `dns_resolve_one($host)` | 解析单个 IP | `$ip = dns_resolve_one("example.com")` |

### 完整示例：Echo 服务器

```jpl
// Echo 服务器 - 回显客户端发送的内容
$server = tcp_listen("0.0.0.0", 8080)
println("Listening on :8080")

$loop = evloop_new()
$handlers = registry_new()

// 处理新连接
$handlers.on_accept($server, fn($client) {
    println("New connection: " .. tcp_peername($client))
    
    // 为新连接注册读事件
    $handlers.on_read($client, fn($fd) {
        $data = tcp_recv($fd, 4096)
        
        // 连接关闭
        if (empty($data)) {
            println("Connection closed")
            $handlers.off($fd)
            tcp_close($fd)
            return
        }
        
        // 回显数据
        tcp_send($fd, "Echo: " .. $data)
    })
    
    // 错误处理
    $handlers.on_error($client, fn($fd, $err) {
        println("Error: " .. $err)
        $handlers.off($fd)
        tcp_close($fd)
    })
})

evloop_attach($loop, $handlers)
evloop_run($loop)
```

### 完整示例：HTTP 服务器

```jpl
// 简单的 HTTP 服务器
$server = tcp_listen("0.0.0.0", 8080)
$loop = evloop_new()
$handlers = registry_new()

fn http_response($status, $body) {
    return "HTTP/1.1 " .. $status .. "\r\n" ..
           "Content-Length: " .. strlen($body) .. "\r\n" ..
           "Connection: close\r\n" ..
           "\r\n" ..
           $body
}

$handlers.on_accept($server, fn($client) {
    $handlers.on_read($client, fn($fd) {
        $data = tcp_recv($fd, 8192)
        
        if (empty($data)) {
            $handlers.off($fd)
            tcp_close($fd)
            return
        }
        
        // 解析请求行
        $lines = split($data, "\r\n")
        $parts = split($lines[0], " ")
        $method = $parts[0]
        $path = $parts[1]
        
        // 路由
        $body = ""
        if ($path == "/") {
            $body = "<h1>Hello JPL!</h1>"
        } else if ($path == "/time") {
            $body = "<p>" .. date("Y-m-d H:i:s") .. "</p>"
        } else {
            $body = "<h1>404 Not Found</h1>"
        }
        
        tcp_send($fd, http_response("200 OK", $body))
        $handlers.off($fd)
        tcp_close($fd)
    })
})

evloop_attach($loop, $handlers)
println("HTTP Server on http://localhost:8080")
evloop_run($loop)
```

### 实现依赖

| 组件 | Go 实现 | 说明 |
|------|---------|------|
| IO 多路复用 | `runtime/epoll` (Linux) / `runtime/kqueue` (BSD) | 或使用 Go netpoll |
| TCP | `net.TCPConn` | Go 标准库 |
| UDP | `net.UDPConn` | Go 标准库 |
| DNS | `net.LookupHost` | Go 标准库 |
| 二进制 | `encoding/binary` | Go 标准库 |

### 决策理由

1. **IO 多路复用优先于协程**：避免协程泄漏、资源可控、高性能
2. **注册表模式**：事件集中管理、代码清晰、便于调试
3. **回调风格**：实现简单、性能好、与底层 epoll/kqueue 模型一致
4. **pack/unpack + Buffer**：覆盖简单和复杂两种场景
5. **Go 标准库实现**：复用成熟代码，无需从头实现网络栈

---

## D20. 多行字符串语法 ✅ 已实现

**决策日期**：2026-03-26
**实现日期**：2026-03-26
**状态**：✅ 已完成

### 语法设计

采用 **Python 风格的三引号** 语法：

```php
// 单引号三引号：纯文本多行字符串（不插值）
$json = '''
{
    "name": "JPL",
    "version": "1.0"
}
'''

// 双引号三引号：支持插值的多行字符串
$template = """
Dear #{$name},

Welcome to #{$service}!

Best regards,
#{$sender}
"""
```

### 设计原则

| 引号类型 | 单行 | 多行 | 插值 | 说明 |
|---------|------|------|------|------|
| `'...'` | ✅ | ❌ | ❌ | 单引号字符串，无转义无插值 |
| `"..."` | ✅ | ❌ | ✅ | 双引号字符串，支持插值（D26） |
| `'''...'''` | ✅ | ✅ | ❌ | 单引号多行，纯文本 |
| `"""..."""` | ✅ | ✅ | ✅ | 双引号多行，支持插值 |

### 决策理由

1. **与现有语法一致**：复用单/双引号的语义（单=纯文本，双=动态）
2. **Python 用户熟悉**：三引号是多行字符串的业界标准
3. **示例兼容**：hello.jpl 示例已使用双引号，自然扩展到三引号
4. **实现简单**：Lexer 只需识别 `"""` 和 `'''` 作为字符串开始/结束标记

### 实现细节

**新增 Token 类型**：
- `TRIPLE_SINGLE`：单引号三引号开始/结束
- `TRIPLE_DOUBLE`：双引号三引号开始/结束

**Lexer 实现**：
- `scanTripleString()`：扫描三引号字符串，支持转义和插值检测
- `scanTripleStringContinue()`：插值结束后继续扫描
- 支持转义字符：`\n`, `\t`, `\r`, `\\`, `\'`, `\"`

**测试覆盖**：7 个测试用例，包括空字符串、JSON、转义等场景

---

## D21. 字符串插值语法

**决策日期**：2026-03-26
**实现日期**：2026-03-26
**状态**：✅ 已完成（MVP + 完整表达式）

### 语法设计

采用 **Ruby 风格的 `#{}`** 语法：

```php
// 基本变量插值
$name = "World"
$greeting = "Hello #{$name}!"

// 表达式插值（第二阶段实现）
$sum = "Result: #{$a + $b}"

// 多行插值
$email = """
Dear #{$name},
Your score is #{$score} / 100.
"""
```

### 分阶段实现

**阶段 1（MVP）**：简单变量插值 ✅ 完成
- 支持：`#{$var}`
- 实现：`parseInterpolatedString()` 构建 ConcatExpr 链

**阶段 2**：完整表达式插值 ✅ 完成
- 支持任意表达式：`#{$obj.name}`、`#{$arr[0]}`、`#{$x + $y * 2}`、`#{$score >= 60 ? 'Pass' : 'Fail'}`、`#{getName()}`
- 实现：`parseExpression(LOWEST)` 解析任意表达式

**阶段 3（可选）**：格式化语法
- Python f-string 风格的格式化：`#{$value:.2f}`
- 状态：未实现（暂缓）

### 决策理由

1. **Ruby/PHP 用户熟悉**：目标用户群体习惯 `#{}` 语法
2. **与示例一致**：hello.jpl 示例已展示 `{...}` 风格
3. **避免语法冲突**：
   - 不采用 `${}`（与 JavaScript/TypeScript 变量声明冲突）
   - 不采用 `{{}}`（与模板引擎语法冲突）
   - `#{}` 在 PHP 中是注释开始，但在字符串内无歧义
4. **分阶段实施**：降低复杂度，快速交付核心价值

### 实现细节

**新增 Token 类型**：
- `INTERP_START`：`#{` 插值开始标记
- `INTERP_END`：`}` 插值结束标记
- `STRING_FRAG`：插值之间的字符串片段

**Lexer 状态管理**：
- `inString`：当前是否在字符串内
- `isTripleQuote`：是否三引号字符串
- `inInterp`：是否在插值表达式内
- `interpDepth`：嵌套插值深度（支持嵌套）

**Parser 实现**：
- `parseInterpolatedString()`：构建 ConcatExpr 链
- `parseInterpStartError()` / `parseInterpEndError()`：处理字符串外的插值标记
- 字符串插值转换为 `"Hello " .. $name .. "!"` 形式

**支持的表达式类型**：
- 变量：`#{$var}`
- 对象属性：`#{$user.name}`、`#{$user.profile.name}`
- 数组索引：`#{$arr[0]}`、`#{$matrix[0][1]}`
- 算术运算：`#{$a + $b}`、`#{$x * $y + 2}`、`#{$price * (1 + $tax)}`
- 逻辑运算：`#{$x > 0}`、`#{$score >= 60}`
- 三元表达式：`#{$score >= 60 ? 'Pass' : 'Fail'}`
- 函数调用：`#{getName()}`、`#{strtoupper($name)}`
- 字符串拼接：`#{$first .. ' ' .. $last}`
- 负号表达式：`#{-$temp}`

**转义机制**：
- `\#{}` 输出字面量 `#{}`
- 用于在双引号字符串中显示原始 `#{}` 语法

**测试覆盖**：
- `interpolation_test.go`：14 个 MVP 测试（基本变量、转义、多行）
- `interp_expression_test.go`：15 个完整表达式测试（对象、数组、算术、三元、函数等）
- 总计 36 个测试用例，全部通过

### 与其他方案对比

| 方案 | 示例 | 优点 | 缺点 | 决策 |
|------|------|------|------|------|
| **Python f-string** | `f"Hello {name}"` | 显式标记 | 需 `f` 前缀，改动大 | ❌ |
| **Ruby #{}** | `"Hello #{name}"` | 双引号自动插值 | `#` 需转义 | ✅ |
| **JS `${}`** | `` `Hello ${name}` `` | 现代语法 | 反引号输入不便 | ❌ |
| **PHP {$}** | `"Hello {$name}"` | 最简洁 | 与数组访问冲突 | ❌ |

### 实现要点

**Lexer 策略**：
```go
// 双引号字符串扫描时，遇到 {# 进入插值模式
// 返回 token 序列：
// STRING_FRAG("Hello ") + INTERP_START + IDENT($name) + INTERP_END + STRING_FRAG("!")
```

**Parser 策略**：
```go
// 将插值字符串展开为 ConcatExpr 链
// "Hello #{$name}!" → ConcatExpr("Hello ", $name, "!")
```

**Compiler 策略**：
```go
// 复用现有的 compileConcatExpr，无需新增指令
```

---

### 联合使用示例

```php
// 多行 + 插值（推荐用于模板）
$report = """
================================
Execution Report
================================

Script: #{$scriptName}
Status: #{$status}
Duration: #{$duration} seconds

Details:
#{$details}

Generated at: #{date("Y-m-d H:i:s")}
================================
"""
```

## D22. 网络模块 API 分层设计

**决策日期**：2026-03-27

### 背景

Phase 9 实现网络编程能力，需要设计 TCP、UDP、Unix Domain Socket 的 API。讨论两种方案：
- 方案A：底层 socket API + 高级 net API 分层
- 方案B：统一的 net API 直接暴露给用户

### 决策

采用 **方案B：统一 net API，不暴露底层 socket**

### 架构设计

```
┌─────────────────────────────────────────┐
│  net 模块（统一 API 层）                   │
│  ├── TCP: net_tcp_listen/connect/accept    │
│  ├── Unix: net_unix_listen/connect/accept  │
│  └── UDP: net_udp_bind/sendto/recvfrom     │
├─────────────────────────────────────────┤
│  Go net 包（底层实现）                     │
└─────────────────────────────────────────┘
```

### API 统一性

| 协议类型 | 创建 | 连接/接受 | 数据传输 | 关闭 |
|---------|------|----------|----------|------|
| **TCP** | `net_tcp_listen(host, port)` | `net_tcp_accept(server)` | `net_send/recv(fd)` | `net_close(fd)` |
| **Unix** | `net_unix_listen(path)` | `net_unix_accept(server)` | `net_send/recv(fd)` | `net_close(fd)` |
| **UDP** | `net_udp_bind(host, port)` | N/A | `net_udp_sendto/recvfrom(fd, ...)` | `net_close(fd)` |

### Unix Domain Socket 处理

Unix Domain Socket 与 TCP 使用 **完全相同的 API**，通过地址类型区分：
- TCP: `net_tcp_listen("0.0.0.0", 8080)`
- Unix: `net_unix_listen("/tmp/server.sock")`
- 通用函数: `net_send`, `net_recv`, `net_close`

不单独创建 `unix` 库，原因：
1. **API 一致性**：Unix socket 本质上是本地 IPC，行为与 TCP 几乎相同
2. **学习成本**：用户只需记忆 `net_` 前缀的函数
3. **代码复用**：错误处理、地址管理逻辑可复用

### 与事件循环集成

```jpl
// 所有 socket 类型都可以注册到事件循环
$registry = ev_registry_new()
$registry.on_read($fd, fn($fd) { ... })
$registry.on_accept($server, fn($client) { ... })
```

### 决策理由

1. **简洁优先**：71 个新函数已经足够多（binary 26 + ev 23 + net 17 + dns 5），避免 API 爆炸
2. **Go 标准库简化**：Go 的 `net` 包已经将底层 socket 细节封装得很好
3. **99% 场景覆盖**：应用开发者不需要精细控制 socket 选项
4. **未来可扩展**：如确实需要底层控制，可后续添加 `socket_xxx` 系列函数

---

## D23. 事件循环注册表模式

**决策日期**：2026-03-27

### 背景

设计事件循环 API 时，讨论两种回调风格：
- 方案A：回调函数风格 `ev.io(fd, fn() { ... })`
- 方案B：注册表风格 `$registry.on_read(fd, fn() { ... })`

### 决策

采用 **方案B：纯注册表模式**

### 设计

```jpl
// 创建注册表（集中管理）
$registry = ev_registry_new()

// 注册事件处理器（声明式）
$registry.on_read($fd, fn($fd) { ... })
$registry.on_write($fd, fn($fd) { ... })
$registry.on_accept($server, fn($client) { ... })
$registry.on_timer(1000000, fn() { ... })  // 微秒级
$registry.on_signal(2, fn() { ... })       // SIGINT

// 批量管理
$registry.off($fd)          // 注销所有
$registry.off_read($fd)     // 只注销读
$registry.clear()           // 清空所有

// 附加到循环并运行
$loop = ev_loop_new()
ev_attach($loop, $registry)
ev_run($loop)
```

### 优势

1. **集中管理**：所有处理器在一个对象中，便于调试和监控
2. **批量操作**：可批量暂停/恢复/清理
3. **生命周期清晰**：显式注册/注销，无泄漏风险
4. **模块化**：不同模块可拥有自己的注册表

### 跨平台实现

```
engine/
├── evpoll_linux.go    // epoll (Linux)
├── evpoll_bsd.go      // kqueue (BSD/Darwin)
└── evpoll_stub.go     // 其他平台 stub
```

**决策理由**：
- 注册表模式比回调风格更易读
- 避免回调地狱
- 与底层 epoll/kqueue 模型一致（注册/注销事件）

---

## D24. 管道运算符决策

**决策日期**：2026-03-27

### 背景

讨论是否为 JPL 语言添加管道运算符（如 OCaml 的 `|>`），用于支持更流畅的函数式编程风格。

### 讨论要点

**现状分析**：
- JPL 已实现丰富的函数式编程支持（map, filter, reduce, find 等 16+ 个函数）
- 管道运算符会新增语法复杂度（token、优先级、结合性）

**已实现的函数式函数**：
- map, filter, reject, reduce, find, some, every
- sort, contains, unique, partition
- flattenDeep, difference, union, zip, unzip

**管道运算符形式**：
```php
// OCaml/Elixir 风格
$data |> filter($x => $x > 0) |> map($x => $x * 2) |> sum()
```

**决策**：暂不添加管道运算符

- 现有函数式函数已覆盖大部分需求
- 管道语法糖可作为未来扩展方向

---

## D25. 事件循环架构重构：goroutine + context

**决策日期**：2026-03-28

### 背景

v0.9.0 使用自定义 epoll/kqueue 实现事件循环，存在以下问题：
1. **fd 生命周期问题**：需要从 `net.Conn` 获取底层 fd，`file.Close()` 导致 fd 无效
2. **平台相关代码**：需要维护 Linux epoll 和 BSD kqueue 两套实现
3. **与 Go runtime 不集成**：无法利用 Go scheduler 的并发能力

### 决策

**重构为 goroutine + context 模式**

### 设计

```
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
│  - fileio.go: readLoop (未来)                        │
└─────────────────────────────────────────────────────┘
```

**核心组件**：
- `EvLoopValue`: 管理 context、定时器、信号
- `EvRegistryValue`: 通用事件注册表，存储 handler 元数据
- 各模块（net.go）：自己启动 goroutine，使用 registry 的 context

**优势**：
1. 消除 fd 生命周期问题
2. 跨平台，删除 ~500 行平台相关代码
3. 更好的并发性能
4. 代码更清晰，职责分离

---

## D26. 通用事件注册表接口

**决策日期**：2026-03-28

### 背景

v0.9.0 的事件注册表 API 直接暴露网络概念（on_accept, on_read），无法复用于文件 IO 等其他事件源。

### 决策

**设计通用抽象接口，各模块封装专用 API**

### 设计

**通用接口**:
```jpl
// 注册事件
$registry.on("accept", $server, fn($source, $data) { ... })
$registry.on("read", $socket, fn($source, $data) { ... })

// 注销事件
$registry.off($source)  // 注销某 source 的所有事件
$registry.off($source, "read")  // 只注销特定类型

// 触发事件（内部使用）
$registry.emit($source, "data", $payload)
```

**各模块封装**:
```jpl
// net.go 提供语法糖
$registry.on_accept($server, fn($client) { ... })
$registry.on_read($socket, fn($socket, $data) { ... })

// fileio.go 可复用（未来）
$registry.on("data", $file, fn($file, $data) { ... })
$registry.on("change", $path, fn($path, $event) { ... })
```

**Go 实现**:
```go
// EvRegistryValue 提供通用接口
RegisterHandler(eventType, source, callback, ctx, cancel) → handlerID
UnregisterHandler(id)
UnregisterBySource(source)
UnregisterBySourceAndType(source, eventType)
Emit(source, eventType, data...)
```

---

## D27. 异步文件 IO 设计（asyncio）

**决策日期**：2026-03-28

### 背景

JPL 需要异步文件 IO 能力，避免大文件操作阻塞事件循环。

### 决策

**采用 Python asyncio 风格的回调 API**

### 设计

**一次性读写**:
```jpl
file_get_async($path, fn($data) { ... })
file_put_async($path, $data, fn() { ... })
file_get_bytes($path, fn($buffer) { ... })
```

**流式读写**:
```jpl
file_read_lines($path, fn($line) { ... }, fn() { ... })
file_read_chunks($path, 4096, fn($chunk) { ... }, fn() { ... })
```

**批量操作**:
```jpl
file_get_batch(["a.txt", "b.txt"], fn($results) { ... })
file_parallel([{op: "read", path: "a.txt"}], fn($results) { ... })
```

**冲突检测**:
- `fileAccessTracker` 全局追踪器
- 批量操作自动检测同文件冲突
- `file_with_lock` 提供手动锁控制

**模块导入**: `import "asyncio"`

**文件**: `buildin/fileio_async.go`

---

## D28. 进程 API 设计

**决策日期**：2026-03-28

### 背景

JPL 缺少系统进程相关的 API，无法执行外部命令、管理环境变量、获取进程信息等。

### 决策

**分优先级实现进程 API，参考 PHP/Python/Node.js 设计**

### 优先级划分

#### P0 - 核心功能（最常用）

| 函数 | 说明 | 参考 |
|------|------|------|
| `exec($cmd)` | 执行命令，返回输出字符串 | PHP exec() |
| `system($cmd)` | 执行命令，返回退出码 | Python os.system() |
| `shell_exec($cmd)` | 执行命令并返回完整输出 | PHP shell_exec() |
| `getenv($name)` | 获取环境变量 | PHP/Python getenv() |
| `setenv($name, $val)` | 设置环境变量 | Python os.putenv() |
| `getppid()` | 获取父进程 ID | Python os.getppid() |
| `tmpdir()` | 获取系统临时目录 | Node.js os.tmpdir() |
| `hostname()` | 获取主机名 | Node.js os.hostname() |

#### P1 - 常用功能

| 函数 | 说明 | 参考 |
|------|------|------|
| `proc_open($cmd, $opts)` | 执行命令，返回进程管道对象 | PHP proc_open() |
| `proc_close($proc)` | 关闭进程管道 | PHP proc_close() |
| `proc_wait($proc)` | 等待进程结束，返回退出码 | PHP proc_wait() |
| `proc_status($proc)` | 获取进程状态 | PHP proc_get_status() |
| `getlogin()` | 获取当前登录用户名 | PHP getlogin() |
| `usleep($us)` | 暂停执行（微秒） | PHP usleep() |
| `putenv($expr)` | 设置环境变量（"KEY=VALUE"格式） | PHP putenv() |

#### P2 - 进阶功能

| 函数 | 说明 | 参考 |
|------|------|------|
| `spawn($cmd, $args)` | 创建子进程（不等待） | Python subprocess.Popen() |
| `kill($pid, $signal)` | 向进程发送信号 | Python os.kill() |
| `waitpid($proc)` | 等待指定子进程 | Python os.waitpid() |
| `fork()` | 创建子进程（Unix） | Unix fork() |
| `pipe()` | 创建管道对 | Python os.pipe() |

#### P3 - 高级功能（精选）

| 函数 | 说明 | 参考 |
|------|------|------|
| `sigwait($sigs)` | 阻塞等待信号 | Python signal.sigwait() |

### P3 跳过的函数

| 函数 | 原因 |
|------|------|
| `signal()` | 已有 `$registry.on_signal()` 替代 |
| `execv()` | 风险高，调用后进程被替换 |
| `daemon()` | 实现复杂，必要性低 |
| `nice()` | 必要性低 |

### 实现

- **文件**: `buildin/process_ext.go`
- **模块**: `import "process"`

---

## D29. pipe/flow 命名决策

**决策日期**：2026-03-28

### 背景

系统管道函数 `pipe()` 与未来函数式编程中的管道组合函数会产生命名冲突。

### 决策

**保留 `pipe()` 给系统管道，函数式管道使用 `flow()` 函数名**

### 命名规范

| 类别 | 函数名 | 说明 |
|------|--------|------|
| 系统 | `pipe()` | OS 管道，进程间通信，返回 PipeValue |
| 函数式 | `flow()` | 数据流处理管道，返回处理结果 |
| 函数式 | `compose()` | 函数组合（右到左） |
| 函数式 | `pipe_fn()` | 函数组合（左到右） |

### 使用示例

```jpl
// 系统管道 - pipe()
$p = pipe()
$p.read   // 读端 fd
$p.write  // 写端 fd

// 函数式管道 - flow()（未来实现）
$result = flow($data, [
    fn($x) { return filter($x, fn($_) { return $_ > 2 }) },
    fn($x) { return map($x, fn($_) { return $_ * 2 }) },
    fn($x) { return reduce($x, 0, fn($acc, $_) { return $acc + $_ }) }
])

// 函数组合 - compose()（未来实现）
$transform = compose($f3, $f2, $f1)  // $f1 -> $f2 -> $f3
$result = $transform($data)
```

### 决策理由

1. **语义清晰**：`pipe()` 保持 Unix/系统编程的语义
2. **函数式命名**：`flow()` 体现数据流动的含义
3. **避免冲突**：两种用途使用不同名称
4. **未来扩展**：可添加 `|>` 运算符作为语法糖

---

## D30. TLS/SSL 模块设计

**决策日期**：2026-03-28

### 决策

**创建独立 `tls` 模块，与 `net` 模块对称设计，不依赖 `crypto` 模块**

### 模块命名

| 方案 | 结果 | 理由 |
|------|------|------|
| `tls` | ✅ 采用 | 简洁专业，现代术语（TLS 取代 SSL） |
| `ssl` | ❌ 拒绝 | 过时术语 |
| `crypto/tls` | ❌ 拒绝 | 嵌套过于复杂 |
| `secure` | ❌ 拒绝 | 抽象不够具体 |

### 模块结构

```
文件名:      buildin/tls.go
模块名:      tls
函数前缀:    tls_
导入方式:    import "tls"
```

### 函数设计

#### 连接管理

| 函数 | 说明 | 对应 net 函数 |
|------|------|---------------|
| `tls_connect(host, port, options?)` | 建立 TLS 连接 | `net_tcp_connect` |
| `tls_listen(port, cert, key, options?)` | 创建 TLS 监听 | `net_tcp_listen` |
| `tls_accept(server)` | 接受 TLS 连接 | `net_tcp_accept` |
| `tls_close(conn)` | 关闭连接 | `net_close` |

#### 数据传输

| 函数 | 说明 |
|------|------|
| `tls_send(conn, data)` | 发送加密数据 |
| `tls_recv(conn, len)` | 接收解密数据 |

#### 信息获取

| 函数 | 说明 |
|------|------|
| `tls_get_cipher(conn)` | 获取协商的加密套件 |
| `tls_get_version(conn)` | 获取 TLS 版本 |
| `tls_get_cert_info(conn)` | 获取对端证书信息 |
| `tls_set_cert(conn, cert, key)` | 设置客户端证书（mTLS） |

### 使用示例

```jpl
// 简单 HTTPS 请求
$conn = tls_connect("api.example.com", 443)
tls_send($conn, "GET / HTTP/1.1\r\nHost: api.example.com\r\n\r\n")
$response = tls_recv($conn, 4096)
tls_close($conn)

// 模块导入方式
import "tls"
$conn = tls.connect("example.com", 443, {
    verify: true,
    ca_file: "/path/to/ca.crt"
})

// mTLS（双向认证）
$conn = tls_connect("secure.example.com", 443, {
    cert_file: "/path/to/client.crt",
    key_file: "/path/to/client.key"
})
```

### 技术实现

- **底层依赖**: Go `crypto/tls` 标准库
- **证书验证**: 默认启用，可配置 CA 证书
- **协议版本**: TLS 1.2+（TLS 1.0/1.1 已弃用）
- **独立实现**: 不依赖独立的 `crypto` 模块，内部使用 Go 加密功能

#### 证书生成

| 函数 | 说明 |
|------|------|
| `tls_gen_cert(options?)` | 生成自签名证书对 |

### 证书生成选项

```jpl
{
    bits: 2048,           // RSA 密钥位数（1024/2048/4096）
    days: 365,            // 有效期天数
    common_name: "CN",    // 证书主题名称
    out_dir: "/tmp",      // 输出目录
    out_prefix: "jpl_tls" // 文件名前缀
}
```

### 点对点证书认证使用示例

```jpl
import "tls"

// 1. 生成自签名证书（用于测试或内部服务）
$paths = tls_gen_cert({
    bits: 4096,
    days: 730,
    common_name: "My Server",
    out_dir: "/opt/certs"
})
println("Cert: " + $paths.cert_path)  # → /opt/certs/jpl_tls_xxx.crt
println("Key: " + $paths.key_path)    # → /opt/certs/jpl_tls_xxx.key

// 2. 服务端使用自签名证书
$server = tls_listen(8443, $paths.cert_path, $paths.key_path)
$client = tls_accept($server)

// 3. 客户端连接自签名服务端（提供 CA 证书）
$conn = tls_connect("server.example.com", 8443, {
    ca_file: $paths.cert_path,  // 使用同一证书作为 CA 验证
    verify: true
})

// 4. 双向认证（mTLS）场景
// 生成 CA 证书
$ca = tls_gen_cert({common_name: "My CA", days: 3650})

// 生成服务端证书
$server_cert = tls_gen_cert({common_name: "Server"})

// 生成客户端证书
$client_cert = tls_gen_cert({common_name: "Client"})

// 客户端连接（提供客户端证书）
$conn = tls_connect("server.example.com", 443, {
    cert_file: $client_cert.cert_path,  // 客户端证书
    key_file: $client_cert.key_path,    // 客户端私钥
    ca_file: $ca.cert_path,              // CA 验证服务端
    verify: true
})
```

### 理由

1. **与 net 对称**: API 设计保持一致，降低学习成本
2. **现代术语**: 使用 TLS 而非 SSL，符合行业标准
3. **独立模块**: TLS 加密是传输层功能，与数据加密（crypto）分离
4. **Go 标准库**: 直接使用 `crypto/tls`，无需重复实现加密算法
5. **证书生成**: 内置 `tls_gen_cert()` 方便测试和内部服务部署

---

## D31. HTTP Client 模块设计

**决策日期**：2026-03-28

### 决策

**创建 `http` 模块，提供高级 HTTP 客户端功能，内部使用 `net` + `tls` 模块**

### 模块结构

```
文件名:      buildin/http.go
模块名:      http
导入方式:    import "http"
```

### 函数设计

#### 简单请求

| 函数 | 说明 |
|------|------|
| `http_get(url, options?)` | GET 请求，返回响应体 |
| `http_post(url, data, options?)` | POST 请求 |
| `http_put(url, data, options?)` | PUT 请求 |
| `http_delete(url, options?)` | DELETE 请求 |
| `http_head(url, options?)` | HEAD 请求 |
| `http_patch(url, data, options?)` | PATCH 请求 |

#### 通用请求

| 函数 | 说明 |
|------|------|
| `http_request(method, url, options?)` | 通用 HTTP 请求 |

### Options 参数结构

```jpl
{
    headers: {          // 自定义请求头
        "Authorization": "Bearer token",
        "Content-Type": "application/json"
    },
    timeout: 30,        // 超时时间（秒）
    follow_redirects: true,  // 是否跟随重定向
    max_redirects: 10,    // 最大重定向次数
    verify_ssl: true,     // 是否验证 SSL 证书
    proxy: "http://proxy:8080",  // 代理设置
    body: "raw body",     // 原始请求体
    json: {key: "value"}, // JSON 请求体（自动设置 Content-Type）
    form: {key: "value"}, // Form 请求体（自动设置 Content-Type）
    auth: {              // 基本认证
        username: "user",
        password: "pass"
    }
}
```

### 响应结构

```jpl
{
    status: 200,              // HTTP 状态码
    status_text: "OK",        // 状态文本
    headers: {...},           // 响应头对象
    body: "response body",     // 响应体字符串
    json(): {...},            // 将 body 解析为 JSON
    text(): "...",            // 响应体文本（别名）
    content_length: 1234,     // 内容长度
    time: 0.5                 // 请求耗时（秒）
}
```

### 使用示例

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

// 带认证和超时
$resp = http_request("GET", "https://api.example.com/private", {
    headers: {"Authorization": "Bearer " + $token},
    timeout: 10
})

// 检查状态码
if ($resp.status >= 200 && $resp.status < 300) {
    println("Success: " + $resp.body)
} else {
    println("Error: " + $resp.status_text)
}

// 下载文件
$resp = http_get("https://example.com/file.zip")
file_put_contents("file.zip", $resp.body)
```

### 技术实现

- **依赖模块**: `net` (TCP) + `tls` (HTTPS)
- **HTTP 版本**: HTTP/1.1（HTTP/2 未来扩展）
- **自动 HTTPS**: URL 以 https:// 开头时自动使用 TLS
- **Cookie 支持**: options 中可配置 `cookies: true`

### 理由

1. **简化使用**: 封装底层网络细节，一行代码完成 HTTP 请求
2. **现代设计**: 参考 Python requests、Node.js fetch 等流行库
3. **灵活配置**: 丰富的 options 参数支持各种场景
4. **响应对象**: 统一响应结构，方便链式处理
---

## D32. 正则表达式模块设计

**决策日期**：2026-03-28

### 决策

**创建 `preg` 模块，使用 Go RE2 语法，提供简化的 7 个核心函数**

### 模块命名

| 方案 | 结果 | 理由 |
|------|------|------|
| `re` | ✅ 采用 | 参考 Python re 模块，语义清晰 |
| `preg` | ❌ 拒绝 | 与 PHP 命名相同但语法不同，易产生混淆 |
| `regex` | ❌ 拒绝 | 过于通用 |
| `regexp` | ❌ 拒绝 | 与 Go 包名重复 |

### 设计原则

1. **无需预编译**：直接使用 pattern 字符串，简化 API
2. **函数命名参考 Python**：re_match, re_sub 等，降低学习成本
3. **返回简单类型**：bool, string, array，不返回复杂对象
4. **支持命名捕获**：使用 Go RE2 语法 `(?P<name>...)`
5. **RE2 语法**：与 Perl/Python/JavaScript 兼容，安全性高（无回溯炸弹）

### 核心函数

| 函数 | 说明 | 参考 |
|------|------|------|
| `re_match(pattern, string)` | 检查是否匹配 | Python re.match() |
| `re_search(pattern, string)` | 查找第一个匹配 | Python re.search() |
| `re_findall(pattern, string)` | 查找所有匹配 | Python re.findall() |
| `re_sub(pattern, replacement, string)` | 替换所有匹配 | Python re.sub() |
| `re_split(pattern, string)` | 按正则分割 | Python re.split() |
| `re_groups(pattern, string)` | 返回捕获组对象 | 简化设计 |

### 使用示例

```jpl
import "re"

// 检查是否匹配
if (re_match("\\d+", "abc123")) {
    println("Contains numbers")
}

// 查找第一个匹配
$email = re_search("[\\w.-]+@[\\w.-]+\\.\\w+", "Contact: john@example.com")
// → "john@example.com"

// 查找所有数字
$numbers = re_findall("\\d+", "Room 101, Floor 5, Building 3")
// → ["101", "5", "3"]

// 替换所有
$text = re_sub("\\d+", "[NUMBER]", "Room 101, Floor 5")
// → "Room [NUMBER], Floor [NUMBER]"

// 分割
$parts = re_split("\\s*,\\s*", "apple, banana ,orange")
// → ["apple", "banana", "orange"]

// 命名捕获组
$groups = re_groups("(?P<year>\\d{4})-(?P<month>\\d{2})", "2024-03")
// → {year: "2024", month: "03", 0: "2024-03", 1: "2024", 2: "03"}
```

### 与 Go regexp 的关系

- 底层使用 Go `regexp` 包（RE2 引擎）
- 语法完全兼容 RE2: https://github.com/google/re2/wiki/Syntax
- 不支持反向引用（与 RE2 一致）
- 线性时间复杂度，无回溯炸弹风险

---

## D33. 加密模块设计

**决策日期**：2026-03-28

### 决策

**创建 `crypto` 模块，提供 Hash、HMAC、对称加密功能，不依赖 TLS 模块**

### 模块命名

| 方案 | 结果 | 理由 |
|------|------|------|
| `crypto` | ✅ 采用 | 标准命名，与 Go crypto 包一致 |
| `cipher` | ❌ 拒绝 | 过于狭窄，仅指加密 |
| `hash` | ❌ 拒绝 | 与现有 hash 模块冲突 |

### 功能范围

| 类别 | 函数 | 说明 |
|------|------|------|
| **Hash** | `sha256(data)` | SHA-256 哈希 |
| | `sha512(data)` | SHA-512 哈希 |
| | `sha1(data)` | SHA-1 哈希（兼容旧系统） |
| **HMAC** | `hmac_sha256(key, data)` | HMAC-SHA256 |
| | `hmac_sha512(key, data)` | HMAC-SHA512 |
| **编码** | `base64_encode(data)` | Base64 编码 |
| | `base64_decode(data)` | Base64 解码 |
| | `hex_encode(data)` | Hex 编码 |
| | `hex_decode(data)` | Hex 解码 |
| **AES** | `aes_encrypt(data, key)` | AES-GCM 加密 |
| | `aes_decrypt(data, key)` | AES-GCM 解密 |

### 设计原则

1. **独立模块**：与 `tls` 模块完全独立，可单独使用
2. **字符串接口**：输入输出均为字符串，自动处理编码
3. **安全默认**：AES 使用 GCM 模式，自动处理 IV/Nounce
4. **密钥处理**：支持 hex/base64 编码的密钥字符串

### 使用示例

```jpl
import "crypto"

// Hash
$hash = crypto.sha256("Hello World")
// → "a591a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e"

// HMAC
$signature = crypto.hmac_sha256("secret_key", "message")
// → 64 字符 hex 字符串

// Base64
$encoded = crypto.base64_encode("Hello")
// → "SGVsbG8="
$decoded = crypto.base64_decode("SGVsbG8=")
// → "Hello"

// AES 加密（GCM 模式，自动处理 IV）
$key = "0123456789abcdef0123456789abcdef"  // 32 字节 hex
$encrypted = crypto.aes_encrypt("Secret data", $key)
// → base64 编码的加密数据（包含 IV）

$decrypted = crypto.aes_decrypt($encrypted, $key)
// → "Secret data"
```

### 实现细节

- **Hash**: Go `crypto/sha256`, `crypto/sha512`
- **HMAC**: Go `crypto/hmac` + `crypto/sha256`
- **Base64**: Go `encoding/base64`
- **AES**: Go `crypto/aes` with GCM mode (`crypto/cipher`)
- **Key 长度**: AES-256 要求 32 字节密钥

---

## D34. 闭包内成员访问语法（@member）

**决策日期**：2026-03-29

- 使用 `@member` 语法在闭包内访问对象成员
- `@member` 引用闭包定义所在的对象字面量的成员
- 嵌套对象时，`@` 绑定到最近一层的对象（静态作用域）
- 对象字面量外的闭包使用 `@member` 会报编译错误

**示例**：
```jpl
// 基本用法
$obj = {value: 10, get: fn() { return @value }}
$obj.get()  // 返回 10

// 嵌套对象
$obj = {
    outer: {
        value: 5,
        get: fn() { return @value }  // @value 绑定到 outer 对象
    }
}
$obj.outer.get()  // 返回 5

// 编译错误：@ 在对象外使用
$fn = fn() { return @value }  // 编译错误
```

**语法设计**：
- `@member` 是特殊语法，不是变量名
- `@$value` 严格解析为成员名 `$value`（不会回退到外层变量）
- 编译时检测闭包是否在对象字面量内，自动捕获 self

**实现机制**：
1. 编译器追踪对象字面量上下文（objectDepth）
2. 对象字面量内创建隐式 `__self__` 变量
3. `@member` 编译为 `self.member` 访问
4. 运行时通过 upvalue 机制访问 self

**理由**：类似 Ruby 的 `@var` 实例变量语法，简洁直观；静态作用域绑定避免闭包传递后的行为不确定性。

---

## D34. 运行时错误定位

**决策日期**：2026-03-29

- 字节码记录每条指令对应的源码行号
- VM 执行时追踪当前源码行号
- 运行时错误包含源码位置信息
- 成员访问保持灵活（不存在返回 null，不报错）

**示例**：
```jpl
// 运行时错误带行号
$ a = 1
$ b = $undefined_var  // 运行时错误: :2:0: undefined variable: $undefined_var
```

**实现机制**：
1. `CompiledFunction.SourceLines` 记录指令→行号映射
2. VM 执行循环中更新 `vm.currentLine`
3. 错误创建时使用 `NewRuntimeErrorWithLocation`

**例外**：
- 成员访问 `$obj.notexist` 返回 null，不报错（保持灵活性）
- 变量未定义返回 null（保持原有行为）

**理由**：运行时错误定位对调试至关重要；成员访问保持灵活以支持动态类型特性。

---

## D35. delete/unset 函数

**决策日期**：2026-03-29

- `delete($obj, "key")` - 删除对象成员或数组元素
- `unset($var)` - 将变量设为 null

**示例**：
```jpl
// 删除对象成员
$obj = {a: 1, b: 2, c: 3}
delete($obj, "b")
println $obj  // {"a": 1, "c": 3}

// 删除数组元素（返回新数组）
$arr = [1, 2, 3, 4, 5]
$newArr = delete($arr, 2)
println $newArr  // [1, 2, 4, 5]

// unset（效果等同于赋值 null）
$x = 10
$x = null  // 推荐写法
```

**实现机制**：
- `delete` 对对象是原地修改，对数组返回新数组
- `unset` 由于语言设计限制，实际效果等同于 `$var = null`

**理由**：提供删除能力，增强语言实用性；保持与 PHP 的兼容性。

---

## D36. Language Server Protocol (LSP) 实现

**决策日期**：2026-03-29

### 背景

JPL 语言已接近完善状态，需要 LSP 支持以提升开发体验。LSP 是现代语言编辑器的标准协议，支持语法检查、代码补全、跳转定义、悬停提示等功能。

### 现状分析

JPL 编译器已具备 LSP 所需的基础设施：

| 组件 | 状态 | 说明 |
|------|------|------|
| **Lexer** | ✅ 完整 | `lexer/` 词法分析 |
| **Parser** | ✅ 完整 | `parser/` 解析 + AST |
| **AST 位置信息** | ✅ 完整 | 所有节点实现 `Pos() token.Position` |
| **符号表** | ✅ 完整 | `Symbol`, `scopes` 作用域管理 |
| **编译错误** | ✅ 包含位置 | `CompileError` 含 Line/Column/File |
| **检查命令** | ✅ 已有 | `jpl check` 命令 |

### 决策

**实现 LSP 支持**，分阶段进行：

### 第一阶段（1-2 天）

| LSP 功能 | 说明 |
|----------|------|
| **诊断 (diagnostics)** | 调用 parser/compiler，获取错误转 LSP 诊断格式 |
| **格式化 (formatting)** | 基于 AST 重新生成代码 |
| **符号搜索 (document symbols)** | 遍历 AST 提取函数/变量声明 |

### 第二阶段（2-3 天）

| LSP 功能 | 说明 |
|----------|------|
| **补全 (completion)** | 分析 AST 上下文，提取可用符号 |
| **定义跳转 (goto definition)** | 利用现有符号表解析标识符 |
| **悬停 (hover)** | 提取符号类型信息 |

### 第三阶段（2-3 天）

| LSP 功能 | 说明 |
|----------|------|
| **引用 (find references)** | 基于符号表查找变量引用 |
| **重构工具** | 符号重命名等 |

### 技术方案

推荐使用 Go 的 LSP 框架：

1. **Bingo** (https://github.com/saiblo/bingo) - 成熟的 Go LSP 骨架
2. 或从头实现，使用 `golang.org/x/tools/lsp` 基础库

### 预估工作量

| 阶段 | 功能 | 预估工时 |
|------|------|----------|
| 第一阶段 | 诊断 + 格式化 + 符号搜索 | 1-2 天 |
| 第二阶段 | 补全 + 跳转定义 + 悬停 | 2-3 天 |
| 第三阶段 | 引用查找 + 重构工具 | 2-3 天 |
| **总计** | | **5-8 天** |

### 结论

**难度：低至中等**

JPL 的编译器架构设计良好，已具备 LSP 所需的核心基础设施。主要工作量在于：
1. 将现有错误信息转换为 LSP 诊断格式
2. 补全/跳转功能需要基于符号表构建索引

建议从第一阶段开始，实现成本低且收益高（语法检查、格式化、Outline 视图）。

---

## D37. 管道运算符设计

**决策日期**：2026-03-29

### 背景

函数式编程中，管道运算符是一种方便的数据流处理语法，可以将多个函数调用串联起来，使代码更清晰易读。JPL 需要支持管道运算符以提升函数式编程体验。

### 设计方案

#### 语法

- **正向管道**：`|>`（两个字符：竖线 + 大于号）
- **反向管道**：`<|`（两个字符：小于号 + 竖线）

#### 语义

**正向管道 `|>`**（左结合）：
- `a |> f(b, c)` = `f(a, b, c)` — 左侧值作为函数的**首个参数**
- `a |> f` — 返回函数引用（不调用）
- `a |> f()` — 调用函数，a 作为首个参数

**反向管道 `<|`**（右结合）：
- `f(b, c) <| a` = `f(b, c, a)` — 右侧值作为函数的**末尾参数**
- `f <| a` — 调用函数，a 作为参数

#### 结合性

- `|>` 左结合：`a |> f |> g` = `(a |> f) |> g` = `g(f(a))`
- `<|` 右结合：`f <| g <| a` = `f <| (g <| a)` = `f(g(a))`

#### 优先级

管道运算符优先级位于箭头函数 `->` 和逻辑运算符 `||` 之间：
```
TernARY (?:)   ← 最低
Assign (=)
Arrow (->)
PIPE_FWD (|>)  ← 新增
PIPE_BWD (<|)  ← 新增
OR (||)
AND (&&)
...
```

### 技术实现

#### 修改的组件

| 组件 | 修改内容 |
|------|----------|
| `token/token.go` | 新增 `PIPE_FWD`、`PIPE_BWD` token |
| `lexer/lexer.go` | 识别 `\|>` 和 `<\|` 两字符组合 |
| `parser/ast.go` | 新增 `PipeExpr` AST 节点 |
| `parser/parser.go` | 新增优先级和解析函数 |
| `engine/compiler.go` | 编译为函数调用 |

#### AST 节点

```go
// PipeExpr 管道表达式
type PipeExpr struct {
    Token   token.Token // |> 或 <| Token
    Left    Expression  // |> 左侧: 值; <| 左侧: 函数
    Right   Expression  // |> 右侧: 函数; <| 右侧: 值
    Forward bool        // true = |>, false = <|
}
```

#### 编译策略

管道表达式编译为普通函数调用：
- `a |> f(b, c)` → 编译为 `f(a, b, c)` 调用
- `f(b, c) <| a` → 编译为 `f(b, c, a)` 调用

### 使用示例

```jpl
fn double(x) { return x * 2 }
fn add(a, b) { return a + b }

// 正向管道
5 |> double()           // = 10
10 |> add(20)           // = 30
5 |> double() |> double()  // = 20

// 反向管道
double() <| 7           // = 14
add(100) <| 50          // = 150
double() <| double() <| 3  // = 12

// 数据处理管道
[1,2,3,4,5]
  |> filter(fn($x) { return $x > 2 })
  |> map(fn($x) { return $x * 10 })
  |> reduce(0, fn($a, $b) { return $a + $b })
// = 90 (30 + 40 + 50)
```

### 与其他语言对比

| 语言 | 正向管道 | 反向管道 | 说明 |
|------|----------|----------|------|
| Elixir | `\|>` | 无 | 左侧值作为首个参数 |
| F# | `\|>` | `<\|` | 与 JPL 相同 |
| OCaml | `\|>` | `<\|` | 与 JPL 相同 |
| JavaScript (TC39) | `\|>` | 无 | 提案阶段 |

### 设计决策

1. **选择 `|>` 而非 `|`**：避免与按位或运算符 `|` 和 match/case 模式冲突
2. **首个参数插入**：`a |> f(b, c)` = `f(a, b, c)`，左侧值作为首个参数
3. **无占位符语法**：保持简单，管道值始终作为首个/末尾参数
4. **双运算符设计**：同时支持正向和反向管道，提供更灵活的代码组织方式

---

## D38. match/case 语法设计

**决策日期**：2026-03-29
**状态**：✅ 已完成（Phase 16）

### 背景

JPL 语言设计参考 Python 3.10+ 的 match/case 语法，但在实现过程中发现以下问题：
1. **编译器缺失 case**：`compileStmt()` 缺少对 `*parser.MatchStmt` 的处理，导致 match 语句被静默跳过
2. **Guard 解析错误**：解析 guard 条件时使用了错误的 token 类型（`token.IDENTIFIER` 而非 `token.IF`）

### 语法设计

参考 **Rust 风格**的 match 语法，执行单个分支后自动跳出：

```jpl
// match 语句（无返回值）
match ($status) {
    case 200: puts "OK"
    case 404: puts "Not Found"
    case 500: puts "Server Error"
    case _: puts "Unknown"  // 默认分支
}

// match 表达式（返回值）
$result = match ($code) {
    case 200: "success"
    case 404: "not found"
    case 500: "error"
    case _: "unknown"
}

// 带 Guard 条件
match ($score) {
    case $x if $x >= 90: "A"
    case $x if $x >= 80: "B"
    case $x if $x >= 70: "C"
    case $x if $x >= 60: "D"
    case _: "F"
}

// OR 模式（多值匹配）
match ($day) {
    case "Saturday", "Sunday" => "Weekend"
    case _: "Weekday"
}

// 标识符绑定
match ($value) {
    case $x: print "Got: " .. $x
}
```

### 支持的模式类型

| 模式 | 示例 | 说明 |
|------|------|------|
| 字面量 | `200`, `"ok"`, `true` | 精确匹配 |
| 标识符绑定 | `$x` | 绑定任意值到变量 |
| 通配符 | `_` | 匹配任意值，不绑定 |
| OR 模式 | `A \| B \| C` | 多选一匹配 |
| Guard 条件 | `if $x > 10` | 额外的布尔条件 |

### 实现细节

#### 编译器实现（engine/compiler.go）

```go
// compileMatchStmt - 编译 match 语句
func (c *Compiler) compileMatchStmt(node *parser.MatchStmt) error

// compileMatchExpr - 编译 match 表达式
func (c *Compiler) compileMatchExpr(node *parser.MatchExpr) error

// compileMatchCase - 编译单个 case 分支
func (c *Compiler) compileMatchCase(node *parser.MatchCase, endLabel string) error
```

#### 编译策略

1. 为每个 case 生成条件跳转指令
2. 使用标签（Label）实现分支结束后的跳出
3. Guard 条件编译为额外的布尔判断
4. OR 模式编译为多个条件分支共享同一处理代码

#### Guard 解析修复

```go
// 错误：使用 IDENTIFIER
case.Condition = c.parseExpression(token.IDENTIFIER)  // ❌

// 正确：使用 IF
case.Condition = c.parseExpression(token.IF)  // ✅
```

### 范围语法设计 ✅ 已完成

由于 `..` 已被定义为字符串连接运算符，范围语法采用不同形式：

| 语法 | 含义 | 示例 |
|------|------|------|
| `...` | 半开区间 [start, end) | `1...10` = 1,2,3,4,5,6,7,8,9 |
| `..=` | 闭区间 [start, end] | `1..=10` = 1,2,3,4,5,6,7,8,9,10 |

**状态**：✅ 已完成（Phase 18）

**实现细节**：
- 词法分析器识别 `...`（ELLIPSIS）和 `..=`（DOT_DOT_EQUAL）token
- 解析器生成 RangeExpr AST 节点，包含 Start、End、Inclusive 字段
- 编译器生成范围迭代代码，支持负数范围
- VM 通过内置 `range` 函数实现范围迭代

### 测试覆盖

- 字面量匹配（数字、字符串、布尔）
- 标识符绑定
- OR 模式
- Guard 条件
- Match 表达式（返回值）
- Match 语句（无返回值）

### 决策理由

1. **Rust 风格跳出**：单个分支执行后自动跳出，避免多个分支意外执行
2. **表达式支持**：match 可作为表达式返回值，增强语言表达能力
3. **Guard 条件**：提供额外的条件过滤能力
4. **OR 模式**：简化多值匹配代码
5. **区分范围语法**：使用 `...` 和 `..=` 避免与字符串连接冲突

---

## D39. 正则字面量语法设计

**决策日期**：2026-03-31
**状态**：设计完成，待实现

### 背景

JPL 现有正则功能通过 `re_match()`、`re_search()` 等函数调用实现，模式以字符串形式传入，写法冗长且不直观。需要引入正则字面量语法，作为一等公民支持条件判断和 match/case 模式匹配。

### 方案选型

考虑了多种语法方案后，综合 JPL 现有语法特征做出选择：

| 方案 | 歧义风险 | 结论 |
|------|---------|------|
| `/pattern/` (JS/Ruby 风格) | 与除法 `/` 和注释 `//` 严重冲突 | ❌ 放弃 |
| `~/pattern/` | `~` 已用作位运算 NOT (TILDE token) | ❌ 放弃 |
| `~r/pattern/` (Elixir 风格) | `~` 已占用，需 lookahead | ⚠️ 复杂度高 |
| `re"pattern"` | 与函数调用 `re()` 视觉混淆 | ⚠️ 可行但不直观 |
| **`#/pattern/flags#`** | `#` 在字符串外未使用，无歧义 | ✅ 采用 |

**唯一已知先例**：Guile Scheme 使用 `#/pattern/` 作为正则字面量。

### 语法规范

```
#/pattern/flags#
```

| 组成 | 说明 |
|------|------|
| `#/` | 正则开始定界符 |
| `pattern` | RE2 模式内容 |
| `/` | 模式结束，flags 起始（无 flags 时紧跟结尾 `#`） |
| `flags` | 可选，`i` `m` `s` `U` 的任意组合 |
| `#` | 正则结束定界符 |

#### Flags（与 Go RE2 引擎对齐）

| Flag | 含义 | RE2 对应 |
|------|------|---------|
| `i` | 忽略大小写 | `(?i)` |
| `m` | 多行模式（`^` `$` 匹配行首行尾） | `(?m)` |
| `s` | `.` 匹配换行符 | `(?s)` |
| `U` | 非贪婪互换 | `(?U)` |

注：Go RE2 不支持 `g`（全局），全局匹配由 API 层面处理（如 `re_findall`）。

#### 转义规则

| 字符 | 写法 | 说明 |
|------|------|------|
| `/` | `\/` | 模式内斜杠 |
| `#` | `\#` | 模式内井号 |
| `\` | `\\` | 模式内反斜杠 |

#### 合法写法

```jpl
#/hello/#              // 无 flags
#/hello/i#             // 忽略大小写
#/hello/ims#           // 组合 flags
#/<a\/b>/#             // 转义内部的 /
#/^(\w+)=(.+)$/ims#    // 捕获组 + flags
```

### `=~` 匹配运算符

新增 `MATCH_EQ` token，语义：`左值(字符串) =~ 右值(正则)`，返回 bool。

```jpl
if ($input =~ #/\d+/) {
    puts "contains digits"
}
```

### match/case 正则模式

#### 基本匹配

子串匹配语义（非精确匹配），用户自行 `^...$` 锚定。第一条匹配的分支执行后跳出，不 fallthrough。

```jpl
match ($input) {
    case #/^quit$/:      exit(0)
    case #/^help$/:      show_help()
    case #/^\d+$/:       handle_number($input)
    case _:              puts "unknown"
}
```

#### 捕获组绑定：`as $var`

使用 `as $var` 显式绑定捕获结果，`$m` 的结构与 `re_groups()` 返回值一致。

```jpl
match ($input) {
    case #/^set (\w+)=(.+)$/ as $m:
        $config[$m[1]] = $m[2]

    case #/^(?P<area>\d{3})-(?P<num>\d{4})$/ as $m:
        puts "area=#{$m['area']}, num=#{$m['num']}"

    case #/^(\d{4})-(\d{2})-(\d{2})$/ as $m if int($m[1]) > 2000:
        puts "year=#{$m[1]}"

    case _: puts "no match"
}
```

`$m` 绑定变量结构：

```jpl
$m[0]           // 完整匹配 "123-4567"
$m[1]           // 第一个捕获组 "123"
$m[2]           // 第二个捕获组 "4567"
$m["area"]      // 命名捕获组 "123"
$m["num"]       // 命名捕获组 "4567"
```

#### 混合使用

```jpl
fn dispatch($input) {
    $input = trim($input)

    match ($input) {
        // 字面量精确匹配
        case "quit", "exit": exit(0)
        case "help":         show_help()

        // 正则子串匹配
        case #/^-?\d+(\.\d+)?$/ as $m:
            $val = float($m[0])
            puts "number: #{$val}"

        // 正则 + guard
        case #/^(\w+)\s*=\s*(.+)$/ as $m:
            puts "set #{$m[1]} = #{$m[2]}"

        // 大小写不敏感
        case #/^hello|hi$/i:
            puts "greeting!"

        case _: puts "unknown: #{$input}"
    }
}
```

### 编译期错误处理

| 错误类型 | 阶段 | 示例 |
|---------|------|------|
| 无效正则语法 | 编译期 | `#/[/invalid/#` → `line 5: invalid regex: [...]` |
| 空模式 | 编译期 | `#//#` → `line 3: empty regex pattern` |
| 未知 flag | 编译期 | `#/pat/x#` → `line 2: unknown regex flag 'x'` |

正则在编译期验证，报错信息包含源码行号和列号。不支持跨行正则字面量。

### 实现涉及文件

| 层级 | 文件 | 改动 |
|------|------|------|
| Token | `token/token.go` | 新增 `REGEX`、`MATCH_EQ` |
| Value | `engine/value.go` | 新增 `regexValue` 类型 |
| Lexer | `lexer/lexer.go` | 新增 `#/` 扫描 + `=~` 识别 |
| AST | `parser/ast.go` | 新增 `RegexLiteral`、`RegexPattern` 节点 |
| Parser | `parser/parser.go` | 注册 prefix/infix，扩展 `parsePattern` |
| Compiler | `engine/compiler.go` | 编译 `=~`、regex case 分支 |
| VM | `engine/vm.go` | 执行 regex 匹配逻辑 |

### 决策理由

1. **`#/.../#` 语法无歧义**：`#` 在字符串外未被 JPL 使用，不会与任何现有语法冲突
2. **Guile Scheme 先例**：虽然认知度低，但有成熟语言的实践验证
3. **`as $var` 显式绑定**：比隐式 `$1` `$2` 更清晰，避免魔法变量，利于调试和静态分析
4. **子串匹配语义**：比精确匹配更实用，用户需要精确匹配时可自行锚定
5. **编译期验证**：正则错误提前暴露，避免运行时失败
6. **与 `re` 模块互补**：字面量用于简单场景，`re_groups()` 等函数用于复杂场景

---

## D40. 包管理器设计

**决策日期**：2026-04-02

### 关键决策

| 决策 | 选择 | 理由 |
|------|------|------|
| 依赖源 | Git URL（无中心化 registry） | 降低基础设施依赖，利用现有 Git 生态 |
| 安装策略 | 扁平化（`jpl_modules/` 根目录） | 类似 npm v3+，避免深层嵌套 |
| 锁文件 | YAML 格式（`jpl.lock.yaml`） | 比 JSON 更易读，复用已有 YAML 依赖 |
| 版本控制 | semver 约束（^, ~, >= 等） | 业界标准，使用 Masterminds/semver 库 |
| 缓存 | `~/.jpl/packages/` 全局缓存 | 避免重复克隆，按 owner/repo/commit 组织 |
| 循环检测 | DFS + 三色标记 | O(V+E) 复杂度，与 resolver 共用算法 |

### 源地址格式

```
https://github.com/user/repo.git           # 最新
https://github.com/user/repo.git@v1.0.0    # 精确 tag
https://github.com/user/repo.git@^1.2.0    # semver 约束
https://github.com/user/repo.git#main      # 分支
../local-path                               # 本地路径
```

### 实现文件

| 文件 | 说明 |
|------|------|
| `pkg/pm/manifest.go` | jpl.json 清单读写 |
| `pkg/pm/git.go` | Git 操作（clone/checkout/tags） |
| `pkg/pm/resolver.go` | DFS 依赖解析 + 循环检测 |
| `pkg/pm/cache.go` | 全局包缓存 |
| `pkg/pm/semver.go` | 版本约束封装 |
| `cmd/jpl/pm.go` | CLI 子命令（init/add/remove/install/list/update/outdated） |

---

## D41. 任务系统设计

**决策日期**：2026-04-02

### 关键决策

| 决策 | 选择 | 理由 |
|------|------|------|
| 定义位置 | `jpl.json` 的 `tasks` 字段 | 项目自包含，无需额外配置文件 |
| 格式 | 字符串或 `{cmd, deps}` 对象 | 简单场景用字符串，复杂场景用对象 |
| 命令类型 | JPL 脚本 + Shell 命令 | 自动判断：`jpl run` 或 `.jpl` → JPL，其他 → shell |
| 依赖执行 | 拓扑排序 + 循环检测 | 与包管理器 resolver 共用算法 |
| 去重 | 依赖只执行一次 | 类似 Make/Rake，避免重复执行 |

### 格式示例

```json
{
    "tasks": {
        "clean": "rm -rf build",
        "build": {
            "cmd": "jpl run scripts/build.jpl",
            "deps": ["clean", "lint"]
        }
    }
}
```

### 实现文件

| 文件 | 说明 |
|------|------|
| `pkg/task/task.go` | TaskDef 类型、ResolveTaskOrder、TaskRunner |
| `cmd/jpl/task.go` | CLI 子命令（run/--list/--dry-run） |

---

## D42. 并行依赖安装

**决策日期**：2026-04-02

### 关键决策

| 决策 | 选择 | 理由 |
|------|------|------|
| 并行粒度 | 克隆阶段并行，安装阶段串行 | 克隆是网络 IO 瓶颈，安装涉及文件写入需串行 |
| 并发控制 | channel 信号量 + `--jobs/-j` 标志 | 避免过多 git 进程，默认 4 并发 |
| 缓存集成 | 先检查缓存，未命中再并行克隆 | 减少不必要的网络请求 |

### 两阶段流程

```
Phase 1: 收集任务 → 检查缓存 → 并行克隆（信号量控制）
Phase 2: 串行安装 → 更新锁文件
```

### 实现文件

| 文件 | 说明 |
|------|------|
| `pkg/pm/git.go` | ParallelClone、CloneJob、CloneJobResult |
| `cmd/jpl/pm.go` | runInstallStandard/runInstallWithResolver 改为两阶段 |

