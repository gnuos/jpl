# JPL 项目计划

## 设计决策

设计决策详见独立文档：[docs/DESIGN.md](docs/DESIGN.md)

当前决策编号：D1-D42（2026-04-02）

---

## Phase 1：核心基础（8-10 周）

### 1.1 项目初始化（1 周）

| 任务 | 预估工时 | 说明 |
|------|----------|------|
| Go 模块初始化 | 0.5 天 | `go mod init`，配置 `go.mod` |
| 目录结构搭建 | 0.5 天 | 按规范创建所有包目录 |
| CI/CD 配置 | 2 天 | GitHub Actions，lint、test、build |
| Makefile 编写 | 1 天 | build、test、lint 等常用命令 |
| README 文档 | 1 天 | 项目介绍、快速开始 |

### 1.2 Token 定义（1 周）

| 任务 | 预估工时 | 说明 | 决策 |
|------|----------|------|------|
| Token 类型定义 | 2 天 | 所有关键字、运算符、分隔符 | D1 |
| 关键字映射表 | 1 天 | 字符串 → Token 映射 | D2 D3 D4 D5 |
| 位置信息结构 | 1 天 | 行号、列号、文件名 | |
| 单元测试 | 1 天 | Token 覆盖率测试 | |

### 1.3 词法分析器（2-3 周）

| 任务 | 预估工时 | 说明 | 决策 |
|------|----------|------|------|
| Scanner 核心 | 3 天 | 字符流读取、位置跟踪 | |
| 标识符和关键字 | 2 天 | $前缀、大小写敏感 | D1 |
| 数字字面量 | 2 天 | 整数、浮点、十六进制、BigInt | |
| 字符串字面量 | 2 天 | 单引号、双引号、转义、Unicode | |
| 运算符扫描 | 2 天 | 算术、位运算、比较、逻辑 | D2 |
| 注释处理 | 1 天 | 单行、多行注释 | |
| 错误恢复 | 1 天 | 词法错误收集与报告 | |
| 单元测试 | 2 天 | 各类 Token 扫描测试 | |

### 1.4 Pratt Parser（3-4 周）

| 任务 | 预估工时 | 说明 | 决策 |
|------|----------|------|------|
| AST 节点定义 | 3 天 | 表达式、语句、声明节点 | |
| 表达式解析 | 4 天 | Pratt 算法核心、优先级表 | |
| 语句解析 | 3 天 | if/else, while, for, foreach | |
| 函数声明解析 | 2 天 | 命名函数、匿名函数、lambda | D3 D6 |
| 变量声明解析 | 1 天 | $var = expr | D1 D4 |
| 导入语句解析 | 1 天 | import "file.jpl" | |
| 错误恢复 | 2 天 | panic mode 恢复机制 | |
| 单元测试 | 3 天 | 各类语法解析测试 | D7 D9 D10 |

---

## Phase 2：编译与执行（6-8 周）

### 2.1 Value 类型系统（2 周）

| 任务 | 预估工时 | 说明 |
|------|----------|------|
| Value 接口设计 | 2 天 | 核心接口定义 |
| 基础类型实现 | 3 天 | null, bool, int64, float64, string |
| 复合类型实现 | 3 天 | array, object |
| 类型转换 | 2 天 | 隐式/显式类型转换 |
| 大数支持 | 3 天 | BigInt, BigDecimal |
| 函数值类型 | 1 天 | 函数作为一等公民 |
| 单元测试 | 2 天 | 类型操作测试 |

### 2.2 字节码编译器（2-3 周）

| 任务 | 预估工时 | 说明 |
|------|----------|------|
| 指令集设计 | 3 天 | 操作码定义、操作数格式 |
| 寄存器分配 | 3 天 | 寄存器分配算法 |
| 表达式编译 | 3 天 | AST → 字节码 |
| 语句编译 | 3 天 | 控制流、跳转指令 |
| 函数编译 | 2 天 | 函数体、闭包环境 |
| 常量池管理 | 1 天 | 字面量、字符串池 |
| 反编译器 | 2 天 | 调试用反汇编 |
| 单元测试 | 2 天 | 编译结果验证 |

### 2.3 虚拟机（2-3 周）

| 任务 | 预估工时 | 说明 |
|------|----------|------|
| VM 核心循环 | 3 天 | fetch-decode-execute |
| 算术运算 | 2 天 | 整数、浮点运算 |
| 比较和逻辑 | 1 天 | 布尔运算 |
| 控制流执行 | 2 天 | 跳转、循环 |
| 函数调用 | 3 天 | 调用栈、参数传递 |
| 闭包执行 | 2 夤 | upvalue 捕获与访问 |
| 错误处理 | 2 天 | 运行时错误收集 |
| 单元测试 | 3 天 | 各类指令执行测试 |

---

## Phase 3：高级特性（6-8 周）

### 3.1 垃圾回收（2-3 周）

| 任务 | 预估工时 | 说明 |
|------|----------|------|
| 引用计数实现 | 3 天 | 增减引用、自动释放 |
| 循环引用检测 | 4 天 | 可达性分析 |
| GC 接口集成 | 2 天 | 与 Go GC 协作 |
| 内存泄漏测试 | 2 天 | 循环引用场景测试 |
| 性能优化 | 2 天 | 减少 GC 暂停时间 |

### 3.2 闭包和作用域（1-2 周）

| 任务 | 预估工时 | 说明 |
|------|----------|------|
| 词法作用域 | 2 天 | 作用域链实现 |
| Upvalue 捕获 | 2 天 | 闭包变量捕获 |
| 尾递归优化 | 2 天 | TCO 实现 |
| global 关键字 | 1 天 | 局部作用域声明全局变量 |
| 静态变量 | 1 天 | static 变量支持 |

### 3.3 函数重载（1 周）

| 任务 | 预估工时 | 说明 |
|------|----------|------|
| 参数数量重载 | 2 天 | 基于参数数量分发 ✅ |

> 注：动态类型语言无需参数类型重载，同参数数量场景由脚本侧 typeof 分发

### 3.4 异常处理（1-2 周）

| 任务 | 预估工时 | 说明 |
|------|----------|------|
| try/catch 语法 | 2 天 | 解析支持 |
| 异常抛出机制 | 2 天 | throw 语句 |
| 栈展开 | 3 天 | 异常传播 |
| 错误码机制 | 2 天 | 可选的错误处理方式 |

---

### 3.5 反射基础设施（1-2 周）

> 反射 API 的底层支撑，在 Phase 4 内置函数之前完成

| 任务 | 预估工时 | 优先级 | 说明 |
|------|----------|--------|------|
| ParamNames 字段 | 0.5 天 | P0 | CompiledFunction 增加 ParamNames []string |
| OP_GETVAR 操作码 | 2 天 | P0 | 运行时按字符串名读取变量（局部+全局） |
| OP_SETVAR 操作码 | 2 天 | P0 | 运行时按字符串名写入变量（局部+全局） |
| ListFunctions API | 1 天 | P1 | VM 公开 funcMap 查询，返回函数名列表 |
| GetFunctionInfo API | 1 天 | P1 | 返回函数名、参数名列表、参数数量 |
| CallByName API | 2 天 | P1 | VM 按函数名字符串动态调用，传参并获取返回值 |
| 单元测试 | 2 天 | P1 | 反射基础设施集成测试 |

---

## Phase 4：标准库（8-10 周）

### 4.1 核心内置函数（3-4 周）

| 任务 | 预估工时 | 说明 | 决策 |
|------|----------|------|------|
| 类型检查函数 | 2 天 | is_int, is_string, etc. | ✅ 已完成 |
| 字符串函数 | 4 天 | strlen, substr, strpos, etc. | ✅ 已完成 |
| 数组函数 | 4 天 | push/pop/shift/unshift/splice 等 12 个 | ✅ 已完成 |
| 函数式编程函数 | 3 天 | map, filter, reduce, find, some, every, sort, contains, reject, partition, unique, flattenDeep, difference, union, zip, unzip | ✅ 已完成 |
| 动态常量 | 1 天 | define(name, value), defined(name) | ✅ 已完成 |
| 预设数学常量 | 0.5 天 | INF/NaN/PI/TAU/E/SQRT2/LN2/LN10 | ✅ 已完成 |
| 调试函数 | 0.5 天 | errors/last_error/clear_errors | ✅ 已完成 |
| JSON 函数 | 3 天 | json_encode, json_decode | |
| 数学函数 | 2 天 | abs, sqrt, ceil, floor, etc.（14 个） | ✅ 已完成 |
| 日期时间函数 | 2 天 | time, date, etc. | |
| eval() 函数 | 2 天 | 运行时执行字符串代码 | ✅ 已完成 |
| 反射 API — 变量查询 | 2 天 | typeof, varexists, getvar, setvar, listvars | ✅ 已完成 |
| 反射 API — 函数查询 | 2 天 | listfns, fn_exists, getfninfo, callfn | ✅ 已完成 |

### 4.2 文件和 I/O 函数（2-3 周）

| 任务 | 预估工时 | 说明 | 状态 |
|------|----------|------|------|
| 文件读写 | 3 天 | read/readLines/write/append/exists | ✅ 已完成 |
| 目录操作 | 2 天 | mkdir/mkdirAll/rmdir/listDir | ✅ 已完成 |
| 文件信息 | 1 天 | stat/fileSize/isFile/isDir | ✅ 已完成 |
| 路径处理 | 1 天 | dirname/basename/extname/joinPath | ✅ 已完成 |

### 4.3 Hash/编码函数（1-2 周）

| 任务 | 预估工时 | 说明 |
|------|----------|------|
| MD5/SHA1 | 2 天 | hash 计算 |
| Base64 | 1 天 | 编码/解码 |
| CRC32 | 1 天 | 校验和 |

### 4.4 标准库管理机制（1-2 周） ✅ 完成

| 任务 | 预估工时 | 说明 | 状态 |
|------|----------|------|------|
| 模块注册 | 2 天 | Engine.RegisterModule + Go 模块 + 模块缓存 | ✅ |
| 字节码编译 | 2 天 | OP_IMPORT/OP_INCLUDE + Compiler 编译 | ✅ |
| 外部加载 | 2 天 | FileModuleLoader + 多路径搜索 + URL 缓存 + 锁文件 | ✅ |
| URL 导入 | 2 天 | import from URL + SHA256 校验 + jpl.lock.yaml | ✅ |
| import...as | - | 别名语法支持 | ✅ |
| 标准库模块 | - | math/strings/arrays/io/hash 模块导出 | ✅ |

---

## Phase 5：CLI 工具（4-5 周）

### 5.1 命令行框架 ✅ 完成

| 任务 | 预估工时 | 说明 | 状态 |
|------|----------|------|------|
| Cobra 集成 | 2 天 | 主命令和子命令结构 | ✅ |
| 全局配置 | 1 天 | verbose/debug flag | ✅ |
| 版本信息 | 0.5 天 | --version 输出 | ✅ |
| 帮助文档 | 1.5 天 | 各子命令帮助 | ✅ |

### 5.2 run 子命令 ✅ 完成

| 任务 | 预估工时 | 说明 | 状态 |
|------|----------|------|------|
| 脚本执行 | 2 天 | 编译并执行脚本文件 | ✅ |
| 参数传递 | 1 天 | $argv/$argc 参数传递，__FILE__/__DIR__ 常量 | ✅ |
| 错误输出 | 1 天 | 友好的错误信息 | ✅ |
| 退出码 | 1 天 | 正确的退出码处理 | ✅ |

### 5.4 check 子命令 ✅ 完成

| 任务 | 预估工时 | 说明 | 状态 |
|------|----------|------|------|
| 语法检查 | 1 天 | 仅编译不执行，支持多文件 | ✅ |
| 错误报告 | 1 天 | 行号、列号、错误信息 | ✅ |

### 5.5 eval 子命令 ✅ 完成

| 任务 | 预估工时 | 说明 | 状态 |
|------|----------|------|------|
| 代码执行 | 1 天 | 从命令行直接执行代码，如 `jpl eval 'print "hello"'` | ✅ |
| 表达式求值 | 0.5 天 | 支持表达式结果输出 | ✅ |

### 5.3 REPL 界面（1 周）✅ 已重写

> **2026-03-26 更新**：已从 Bubble Tea TUI 简化为 go-prompt 实现，详见 [D17. REPL 简化重写](docs/DESIGN.md#d17-repl-简化重写)
>
> 旧版设计详见 [D16. REPL 界面设计](docs/DESIGN.md#d16-repl-界面设计)

| 任务 | 预估工时 | 说明 | 状态 |
|------|----------|------|------|
| go-prompt 集成 | 0.5 天 | 基础框架搭建、命令循环 | ✅ |
| 自动补全 | 0.5 天 | Tab 触发、关键字/函数/变量补全 | ✅ |
| 历史导航 | 0.5 天 | 上下箭头浏览、Ctrl+R 搜索 | ✅ |
| 调试指令 | 0.5 天 | `:debug/:globals/:funcs/:help` 等指令 | ✅ |
| 历史持久化 | 0.5 天 | ~/.jpl/repl_history 保存/加载 | ✅ |
| 执行超时 | 0.5 天 | 10秒超时保护 + Ctrl+C 中断 | ✅ |
| 集成测试 | 0.5 天 | 子进程方式测试 | ✅ |

---

## Phase 6：测试与优化（4-5 周）

### 6.1 兼容性测试（已取消）

> **状态**：❌ 已取消
> 
> **原因**：JPL 语法已与原版 Jx9 产生差异，不再追求完全兼容。
> 主要差异包括：变量命名规则、函数语法、内置函数等。
> 转而专注于 JPL 自身的功能和稳定性测试。

~~| 任务 | 预估工时 | 说明 |
|------|----------|------|
| 测试用例移植 | 3 天 | 原版 Jx9 测试套件 |
| 自动化运行 | 2 天 | 测试脚本和 CI 集成 |
| 差异分析 | 2 天 | 兼容性问题记录 |~~

### 6.2 性能测试（1 周）

| 任务 | 预估工时 | 说明 |
|------|----------|------|
| 基准测试 | 2 夤 | 与 C 版本性能对比 |
| 压力测试 | 2 天 | 递归、循环、内存 |
| 并发测试 | 2 天 | 100 goroutine 测试 |

### 6.3 REPL 测试（1 周）

| 任务 | 预估工时 | 说明 |
|------|----------|------|
| Expect 脚本 | 3 天 | 自动化 REPL 测试 |
| 交互场景 | 2 天 | 各种输入场景覆盖 |

### 6.4 文档与示例（1 周）

#### 目标
完善项目文档，编写使用示例，为正式发布做准备。

| 任务 | 预估工时 | 说明 |
|------|----------|------|
| API 文档 | 2 天 | GoDoc 注释完善 |
| 使用指南 | 2 天 | README 和 wiki |
| 示例代码 | 2 天 | example_test.go |

#### 任务清单

| 序号 | 任务 | 优先级 | 工时 | 说明 | 状态 |
|------|------|--------|------|------|------|
| 1 | API GoDoc 完善 | P0 | 2 天 | engine/ 包所有导出函数/类型添加详细注释（1000+ 行） | ✅ 完成 |
| 2 | README.md 重构 | P0 | 1 天 | 完整重写，包含项目介绍、安装、快速开始、语言教程、CLI、API | ✅ 完成 |
| 3 | 使用指南编写 | P0 | 2 天 | 通过 examples/ 目录提供 14 个示例文件代替 GUIDE.md | ✅ 完成 |
| 4 | 示例代码编写 | P1 | 2 天 | 创建 examples/ 目录，14 个示例覆盖所有语言特性 | ✅ 完成 |

#### 已知限制

| 问题 | 说明 | 影响 |
|------|------|------|
| 嵌套 include | 模块间的全局变量索引映射在嵌套 include 时可能不一致 | TestNestedInclude 失败，不影响核心功能 |

### 6.5 性能优化

> 基于 Phase 6.2 性能测试结果，对比 V8 (Ignition) 和 CPython 3.12，识别瓶颈并分级优化。

#### 性能评估基准

**测试环境**: AMD Ryzen 9 9955HX, Go 1.22, Linux x86_64

| 操作 | JPL (优化前) | V8 Ignition | CPython 3.12 |
|------|-------------|-------------|--------------|
| 空循环迭代 | 89 ns | ~2 ns | ~50 ns |
| 整数加法 | 141 ns | ~3 ns | ~70 ns |
| 函数调用 | 227 ns | ~10 ns | ~100 ns |
| 递归 fib(20) | 4.2 ms | ~0.1 ms | ~1.5 ms |

#### 三大核心瓶颈

**1. 完全装箱的值表示**：每个值都是 `Value` 接口（16 字节 fat pointer），整数 `NewInt()` 分配 `*intValue{}` 结构体（24 字节）。1M 次循环 = 3M 次堆分配。V8 用 SMI 直接编码，CPython 有小整数缓存。

**2. 无特化指令**：所有算术走统一接口派发 `Value.Add()` → `switch b.Type()` → `NewInt()`。V8 有 SMI 快速路径，CPython 3.11 有自适应特化字节码。

**3. 函数调用开销过高**：每次调用 3 次堆分配（参数切片、寄存器窗口、null 填充）+ O(n) 寄存器初始化。V8 寄存器复用，CPython 有 frame freelist。

#### 优化分级

| 优先级 | 改进项 | 预期收益 | 实现难度 | 状态 |
|--------|--------|----------|----------|------|
| **P0** | Null/Bool/Int 小值池化 | 消除 40%+ 堆分配 | 低（1 天） | ✅ 完成 |
| **P0** | 算术 int-int 快速路径 | 算术 11-18% 提升 | 低（1 天） | ✅ 完成 |
| **P1** | 字符串内部化 | 字符串比较 O(n)→O(1) | 低（1-2 天） | ✅ 完成 |
| **P1** | 全局变量缓存 | 全局访问 3-5x | 低（1 天） | ✅ 完成 |
| ~~P1~~ | ~~寄存器窗口池化~~ | ~~函数调用 2-3x~~ | ~~中（2 天）~~ | ❌ 跳过（与 upvalue 机制冲突） |
| **P1** | 常量折叠 | 编译质量提升 | 低（1 天） | ✅ 完成 |
| ~~P2~~ | ~~NaN-boxing 值表示~~ | ~~全面 5-10x~~ | ~~高（1-2 周）~~ | ❌ **已取消**（风险过高） |
| ~~P2~~ | ~~特化字节码~~ | ~~算术+比较 3-5x~~ | ~~高（1 周）~~ | ❌ **已取消**（收益重叠） |
| ~~P3~~ | ~~Inline Cache~~ | ~~属性访问 5-10x~~ | ~~高（1-2 周）~~ | ❌ **已取消**（架构限制） |
| ~~P3~~ | ~~JIT 编译~~ | ~~全面 10-100x~~ | ~~极高（数月）~~ | ❌ **已取消**（非优先级） |

> **说明**：经过评估，P2/P3 级别的优化实现难度高、风险大，与当前架构存在冲突，暂不实施。当前性能已满足大部分场景需求。

#### P0 优化结果

**P0-1 小值池化**（`engine/value.go`）：
- 全局单例：`nullSingleton`、`boolTrueValue`、`boolFalseValue`
- 小整数缓存：`[-256, 1024]` 范围预分配 1281 个 `intValue`
- `NewNull()`/`NewBool()`/`NewInt()` 在范围内返回缓存，零分配

**P0-2 int-int 快速路径**（`engine/vm.go`）：
- `opAdd/opSub/opMul/opDiv/opMod` + `opEq/opNeq/opLt/opGt/opLte/opGte/opNeg`
- 使用 concrete type assertion `*intValue` 跳过接口派发
- `opDiv` 保留 float 除法语义，`opMod` 保留零返回 NaN 语义

**优化前后对比**：

| 测试 | 优化前 | 优化后 | 提升 |
|------|--------|--------|------|
| ArithmeticAdd | 141 ns | 122 ns | -13% |
| ConditionalBranching | 199 ns | 163 ns | -18% |
| FunctionCall | 227 ns | 203 ns | -11% |
| RecursionFib(20) | 4.2 ms | 3.7 ms | -12% |
| 1M 循环 | 154ms (6.5M/s) | 123ms (8.2M/s) | +27% |
| fib(30) | 564ms | 480ms | -15% |

#### P1 优化结果

**P1-1 字符串内部化**（`engine/value.go`）：
- `stringInternPool`（`sync.Map`）线程安全缓存池
- 仅缓存长度 ≤64 字节的短字符串，避免内存爆炸
- 重复字符串（如变量名、键名）返回同一 `*stringValue` 实例
- 字符串比较从内容比较 O(n) 变为指针比较 O(1)

**P1-2 全局变量缓存**（`engine/vm.go` + `engine/compiler.go`）：
- `globals` 从 `map[string]Value` 改为 `[]Value` slice
- 编译期分配整数索引：`Compiler.allocateGlobalIndex()`
- `OP_GETGLOBAL`/`OP_SETGLOBAL` Bx 字段改为直接索引而非常量池索引
- 子编译器通过指针共享 `globalNames` 确保一致性
- 全局变量访问从字符串 hash 查找变为数组索引 O(1)

---

## 指令集扩展计划（按需）

> 基于 D14 设计决策，指令集扩展作为性能优化手段，在 Phase 6 性能测试后根据实际瓶颈决定。

### 当前指令集

**已实现 38 条指令**，覆盖：加载/存储(9)、算术(6)、比较(6)、字符串(1)、逻辑(3)、数组/对象(6)、控制流(3)、函数调用(2)、闭包(4)、异常(3)、其他(3)

### 待扩展指令（按优先级）

| 优先级 | 指令类别 | 数量 | 触发条件 | 预估工时 |
|--------|----------|------|----------|----------|
| P1 | 位运算指令 | 6 | 需要脚本实现 Hash 算法 | 1 天 |
| P1 | 调试指令 | 3 | 实现断点调试功能 | 2 天 |
| P2 | 循环优化指令 | 4 | 循环性能测试不达标 | 1 天 |
| P2 | 快速路径指令 | 4 | 计数器循环成为热点 | 1 天 |
| P3 | 数学内置指令 | 5 | 科学计算场景需求 | 1 天 |
| P3 | 字符串指令 | 3 | 字符串操作成为热点 | 1 天 |

### 位运算指令详情

```
OP_BITAND   // R[A] = R[B] & R[C]
OP_BITOR    // R[A] = R[B] | R[C]
OP_BITXOR   // R[A] = R[B] ^ R[C]
OP_BITNOT   // R[A] = ~R[B]
OP_SHL      // R[A] = R[B] << R[C]
OP_SHR      // R[A] = R[B] >> R[C]
```

### 调试指令详情

```
OP_LINE     // 设置当前源码行号（调试用）
OP_BREAK    // 断点支持
OP_STEP     // 单步执行
```

### 循环优化指令详情

```
OP_FORPREP  // 初始化循环：R[A]=起始值, R[A+1]=结束值, R[A+2]=步长
OP_FORLOOP  // 循环迭代：PC += sBx 直到 R[A] >= R[A+1]
OP_INC      // R[A]++
OP_DEC      // R[A]--
```

---

## Phase 7：内置函数完善（4-6 周）

> 基于 Jx9 内置函数对比分析，补充缺失的核心功能函数和常量。

### 7.1 核心类型与转换（第 1 周）

| 任务 | 预估工时 | 说明 | 依赖 |
|------|----------|------|------|
| 类型别名 | 0.5 天 | is_real/is_double/is_integer/is_long 作为别名 | 无 |
| 类型转换函数 | 2 天 | intval(), floatval(), strval(), boolval() | 无 |
| empty() 函数 | 1 天 | 检查值是否为空（null/空串/空数组/0/false） | 无 |
| is_numeric() | 0.5 天 | 检查是否为数字（整数或浮点数） | 无 |
| is_scalar() | 0.5 天 | 检查是否为标量类型 | 无 |
| 单元测试 | 1 天 | 每个函数 3-5 个测试用例 | 上述完成 |

### 7.2 数组操作扩展（第 1 周）

| 任务 | 预估工时 | 说明 | 依赖 |
|------|----------|------|------|
| count/sizeof | 0.5 天 | 数组长度别名 | 无 |
| array_key_exists/key_exists | 1 天 | 检查键是否存在 | 无 |
| array_merge | 1 天 | 合并多个数组 | 无 |
| array_sum/array_product | 1 天 | 计算数组和/乘积 | 无 |
| in_array | 0.5 天 | 检查值是否在数组中 | 无 |
| array_values | 0.5 天 | 获取所有值 | 无 |
| array_diff/array_intersect | 1.5 天 | 数组差集/交集 | 无 |
| 单元测试 | 1 天 | 每个函数 3-5 个测试用例 | 上述完成 |

### 7.3 字符串处理扩展（第 2 周）

| 任务 | 预估工时 | 说明 | 依赖 |
|------|----------|------|------|
| 别名函数 | 0.5 天 | implode(explode别名), chop(rtrim别名) | 无 |
| trim 系列 | 1 天 | rtrim(), ltrim(), 增强 trim() 支持字符集 | 无 |
| 字符串比较 | 1.5 天 | strcmp, strcasecmp, strncmp, strncasecmp | 无 |
| 查找函数 | 2 天 | stripos, strrpos, strripos, strstr, stristr, strchr | 无 |
| 大小写 | 0.5 天 | mb_strtolower, mb_strtoupper 别名 | 无 |
| 格式化 | 1.5 天 | sprintf, printf, vsprintf, vprintf | 无 |
| 其他基础 | 2 天 | ord, chr, bin2hex, str_repeat, nl2br | 无 |
| 单元测试 | 1.5 天 | 每个函数 3-5 个测试用例 | 上述完成 |

### 7.4 数学与三角函数（第 1 周）

| 任务 | 预估工时 | 说明 | 依赖 |
|------|----------|------|------|
| 三角函数 | 1.5 天 | sin, cos, tan, asin, acos, atan, atan2 | 无 |
| 双曲函数 | 1 天 | sinh, cosh, tanh | 无 |
| 对数/指数 | 1 天 | log, log10, exp, pi() | 无 |
| 进制转换 | 1.5 天 | dechex, decoct, decbin, hexdec, bindec, octdec, base_convert | 无 |
| 随机数增强 | 1 天 | rand, rand_str, getrandmax | 无 |
| 其他 | 1 天 | fmod, hypot, round 增强 | 无 |
| 单元测试 | 1 天 | 每个函数 3-5 个测试用例 | 上述完成 |

### 7.5 日期时间扩展（第 1 周）

| 任务 | 预估工时 | 说明 | 依赖 |
|------|----------|------|------|
| 获取函数 | 1 天 | getdate, gettimeofday, localtime | 无 |
| 格式化 | 1.5 天 | strftime, gmdate, idate | 无 |
| 时间戳 | 1.5 天 | mktime, gmmktime | 无 |
| 单元测试 | 1 天 | 每个函数 3-5 个测试用例 | 上述完成 |

### 7.6 文件系统扩展（第 1 周）

| 任务 | 预估工时 | 说明 | 依赖 |
|------|----------|------|------|
| 目录操作 | 1 天 | chdir, getcwd, rename | 无 |
| 文件删除 | 0.5 天 | unlink, delete | 无 |
| 权限 | 1 天 | chmod, chown, chgrp | 无 |
| 属性检查 | 1.5 天 | is_link, is_readable, is_writable, is_executable, filetype | 无 |
| 路径 | 1 天 | realpath, pathinfo | 无 |
| 目录扫描 | 1.5 天 | scandir, glob, fnmatch/strglob | 无 |
| 单元测试 | 1.5 天 | 每个函数 3-5 个测试用例 | 上述完成 |

### 7.7 URI/URL 处理（第 1 周，可选）

| 任务 | 预估工时 | 说明 | 依赖 |
|------|----------|------|------|
| URL 编码 | 1 天 | urlencode, urldecode, rawurlencode, rawurldecode | 无 |
| URL 解析 | 1.5 天 | parse_url | 无 |
| UU 编码 | 1 天 | convert_uuencode, convert_uudecode | 无 |
| 单元测试 | 1 天 | 每个函数 3-5 个测试用例 | 上述完成 |

### 7.8 常量与魔术常量（第 1 周）

| 任务 | 预估工时 | 说明 | 依赖 |
|------|----------|------|------|
| 平台常量 | 0.5 天 | JPL_VERSION, JPL_OS, __OS__, PATH_SEPARATOR, DIRECTORY_SEPARATOR | 无 |
| 整数常量 | 0.5 天 | JPL_INT_MAX, MAXINT, JPL_INT_SIZE | 无 |
| 魔术常量 | 2 天 | __FILE__, __DIR__, __LINE__, __TIME__, __DATE__ | 编译器支持 |
| 数组常量 | 0.5 天 | COUNT_NORMAL, COUNT_RECURSIVE | 无 |
| 排序常量 | 0.5 天 | SORT_ASC, SORT_DESC, SORT_REGULAR, SORT_NUMERIC, SORT_STRING | 无 |
| 标准流 | 0.5 天 | STDIN, STDOUT, STDERR | 无 |
| 单元测试 | 1 天 | 常量值验证 | 上述完成 |

### 7.9 VM 与反射函数（第 1 周，可选）

| 任务 | 预估工时 | 说明 | 依赖 |
|------|----------|------|------|
| 函数参数 | 1.5 天 | func_num_args, func_get_arg, func_get_args | 无 |
| 函数检查 | 1 天 | function_exists, is_callable, get_defined_functions | 无 |
| 常量检查 | 0.5 天 | get_defined_constants | 无 |
| 引擎信息 | 1 天 | jpl_version, jpl_credits, jpl_info, jpl_copyright | 无 |
| 编码 | 1 天 | utf8_encode, utf8_decode | 无 |
| 单元测试 | 1 天 | 每个函数 3-5 个测试用例 | 上述完成 |

### Phase 7 工时汇总

| 子阶段 | 预估工时 | 实际工时 | 优先级 | 状态 | 说明 |
|--------|----------|----------|--------|------|------|
| 7.1 类型与转换 | 5.5 天 | ~4 小时 | 🔴 高 | ✅ 完成 | 13个函数：类型别名、转换函数、empty等 |
| 7.2 数组扩展 | 6 天 | ~4 小时 | 🔴 高 | ✅ 完成 | 12个函数：count/merge/sum/diff/intersect等 |
| 7.3 字符串扩展 | 10.5 天 | ~6 小时 | 🔴 高 | ✅ 完成 | 23个函数：trim系列/比较/查找/格式化/sprintf等 |
| 7.4 数学扩展 | 7 天 | ~3 小时 | 🟡 中 | ✅ 完成 | 18个新函数：三角/双曲/对数（总计32个） |
| 7.5 日期时间 | 5 天 | ~3 小时 | 🟡 中 | ✅ 完成 | 7个新函数：getdate/strftime/mktime等 + now()增强 |
| 7.6 文件系统 | 7 天 | ~3 小时 | 🟡 中 | ✅ 完成 | 9个新函数：chdir/rename/chmod/scandir/glob等 + cwd() |
| 7.7 URI/URL | 4.5 天 | ~2 小时 | 🟢 低 | ✅ 完成 | 5个函数：urlencode/urldecode/rawurlencode/rawurldecode/parse_url |
| 7.8 常量 | 5 天 | ~4 小时 | 🔴 高 | ✅ 完成 | 8魔术常量 + EOL + STDIN/STDOUT/STDERR |
| 7.9 VM/反射 | 5 天 | ~2 小时 | 🟢 低 | ✅ 完成 | 10个函数：func_num_args/function_exists/jpl_version/utf8_encode等 |
| **总计** | **~46 天** | **~31 小时** | | **100% 完成** | **已交付 117 项** |

---

## Phase 8：流 IO 系统

> 引入流资源类型，实现真正的 IO 操作能力，为 pipe 和 socket 标准库奠定基础。

### 8.1 流类型基础 ✅ 完成

| 任务 | 预估工时 | 说明 | 状态 |
|------|----------|------|------|
| TypeStream 枚举 | 0.5 天 | ValueType 新增 TypeStream = 10 | ✅ |
| streamValue 结构 | 1 天 | 实现 Value 接口，组合 io.Reader/Writer/Closer | ✅ |
| 流构造函数 | 1 天 | NewStream, NewFileStream, NewStdinStream 等 | ✅ |
| 预定义标准流 | 0.5 天 | STDIN/STDOUT/STDERR 改为流类型注册 | ✅ |
| 单元测试 | 1 天 | 流创建、读写、关闭、类型检查 | ✅ |

### 8.2 IO 函数扩展 ✅ 完成

| 任务 | 预估工时 | 说明 | 状态 |
|------|----------|------|------|
| fopen(path, mode) | 1 天 | 打开文件流，支持 r/w/rw 模式 | ✅ |
| fread(stream, length) | 0.5 天 | 读取指定字节数 | ✅ |
| fgets(stream) | 0.5 天 | 读取一行 | ✅ |
| fwrite(stream, data) | 0.5 天 | 写入数据 | ✅ |
| fclose(stream) | 0.5 天 | 关闭流 | ✅ |
| feof(stream) | 0.5 天 | 是否到达末尾 | ✅ |
| fflush(stream) | 0.5 天 | 刷新缓冲区 | ✅ |
| print/println 扩展 | 1 天 | 支持可选流参数：print(STDERR, "msg") | ✅ 已实现 |
| is_stream() 类型检查 | 0.5 天 | 检查值是否为流类型 | ✅ |
| 单元测试 | 1 天 | 每个函数 3-5 个测试用例 | ✅ |

### 8.3 流高级功能 ✅ 完成

| 任务 | 预估工时 | 说明 | 状态 |
|------|----------|------|------|
| 流模式枚举 | 0.5 天 | StreamRead/StreamWrite/StreamReadWrite | ✅ |
| 流状态检查 | 0.5 天 | is_readable/is_writable 增强 | ✅ |
| 流信息函数 | 1 天 | stream_get_meta_data 等 | ✅ |
| 临时流 | 1 天 | NewBufferStream（内存流） | ✅ |
| 单元测试 | 1 天 | 高级功能测试 | ✅ |

### Phase 8 实际工时

| 子阶段 | 预估工时 | 实际工时 | 状态 | 说明 |
|--------|----------|----------|------|------|
| 8.1 流类型基础 | 4 天 | ~2 小时 | ✅ 完成 | TypeStream + streamValue + 标准流 |
| 8.2 IO 函数扩展 | 5.5 天 | ~2 小时 | ✅ 完成 | fopen/fread/fwrite/fclose 等 8 个函数 |
| 8.3 流高级功能 | 4 天 | ~1 小时 | ✅ 完成 | 流模式、状态检查、内存流 |
| **总计** | **~13.5 天** | **~5 小时** | **✅ 100%** | |

---

## Phase 9：网络/Socket 标准库（3-4 周）

> 基于 IO 多路复用 + 回调注册表模式，提供完整的网络编程能力。
> 设计文档: [D24. 网络框架设计](docs/DESIGN.md#d24-网络框架设计)

### 设计要点

- **并发模型**：IO 多路复用（非协程），单线程处理万级连接
- **API 风格**：回调注册表模式，分离事件定义和循环执行
- **二进制处理**：pack/unpack 函数 + Buffer 对象

### 9.1 二进制处理（第 1 周）✅ 完成

| 任务 | 预估工时 | 说明 | 依赖 | 状态 |
|------|----------|------|------|------|
| pack 函数 | 1 天 | 格式化打包，支持 C/S/s/N/V/Q/q/f/d/a/Z/x | 无 | ✅ |
| unpack 函数 | 1 天 | 格式化解包 | pack | ✅ |
| bufferValue 类型 | 1.5 天 | 新增 TypeBuffer，可增长字节缓冲区 | 无 | ✅ |
| buffer 读写函数 | 1 天 | write_int8/16/32, read_int8/16/32 等 | bufferValue | ✅ |
| buffer 转换函数 | 0.5 天 | to_bytes, to_string, length, reset | bufferValue | ✅ |
| 单元测试 | 1 天 | pack/unpack 各格式、buffer 读写 | 上述完成 | ✅ |

### 9.2 事件循环核心（第 1-2 周）✅ 完成

| 任务 | 预估工时 | 说明 | 依赖 | 状态 |
|------|----------|------|------|------|
| evloopValue 类型 | 1 天 | 新增 TypeEvloop，封装 epoll/kqueue | 无 | ✅ |
| registryValue 类型 | 1 天 | 新增 TypeRegistry，事件注册表 | 无 | ✅ |
| registry 事件注册 | 1 天 | on_read/on_write/on_accept/on_error/on_timer | registryValue | ✅ |
| registry 事件注销 | 0.5 天 | off/off_read/off_write | registryValue | ✅ |
| evloop 运行控制 | 1 天 | attach/run/run_once/stop | evloopValue | ✅ |
| Go netpoll 集成 | 2 天 | 封装 Go runtime 网络轮询 | evloopValue | ✅ |
| 单元测试 | 1.5 天 | 注册/注销/运行/停止 | 上述完成 | ✅ |

### 9.3 TCP 网络层（第 2 周）✅ 完成

| 任务 | 预估工时 | 说明 | 依赖 | 状态 |
|------|----------|------|------|------|
| tcp_listen | 1 天 | 创建监听 socket | 9.2 | ✅ |
| tcp_connect | 0.5 天 | 连接远程主机 | 9.2 | ✅ |
| tcp_accept | 0.5 天 | 接受连接（非阻塞） | 9.2 | ✅ |
| tcp_send/recv | 1 天 | 发送/接收数据 | 9.2 | ✅ |
| tcp_close | 0.5 天 | 关闭连接 | 9.2 | ✅ |
| tcp 信息函数 | 0.5 天 | peername/sockname/set_nonblock | 9.2 | ✅ |
| 单元测试 | 1 天 | echo 服务器测试 | 上述完成 | ✅ |

### 9.4 UDP 网络层（第 2 周）✅ 完成

| 任务 | 预估工时 | 说明 | 依赖 | 状态 |
|------|----------|------|------|------|
| udp_bind | 0.5 天 | 创建 UDP socket | 9.2 | ✅ |
| udp_sendto/recvfrom | 1 天 | 发送/接收 UDP 数据 | 9.2 | ✅ |
| 单元测试 | 0.5 天 | UDP echo 测试 | 上述完成 | ✅ |

### 9.5 DNS 解析（第 1 周）✅ 完成

| 任务 | 预估工时 | 说明 | 依赖 | 状态 |
|------|----------|------|------|------|
| dns_resolve | 0.5 天 | DNS 解析所有 IP | 无 | ✅ |
| dns_resolve_one | 0.5 天 | 解析单个 IP | 无 | ✅ |
| 单元测试 | 0.5 天 | 域名解析测试 | 上述完成 | ✅ |

### 9.6 Unix Domain Socket（扩展）✅ 完成

| 任务 | 预估工时 | 说明 | 依赖 | 状态 |
|------|----------|------|------|------|
| unix_listen | 0.5 天 | 创建 Unix Domain 监听 socket | 9.2 | ✅ |
| unix_connect | 0.5 天 | 连接 Unix Domain socket | 9.2 | ✅ |
| unix_accept | 0.5 天 | 接受 Unix Domain 连接 | 9.2 | ✅ |
| 单元测试 | 0.5 天 | Unix socket 测试 | 上述完成 | ✅ |

### Phase 9 工时汇总

| 子阶段 | 预估工时 | 实际工时 | 优先级 | 状态 |
|--------|----------|----------|--------|------|
| 9.1 二进制处理 | 6 天 | ~8 小时 | 🔴 高 | ✅ 完成 |
| 9.2 事件循环 | 8 天 | ~10 小时 | 🔴 高 | ✅ 完成 |
| 9.3 TCP 网络 | 5 天 | ~6 小时 | 🔴 高 | ✅ 完成 |
| 9.4 UDP 网络 | 2 天 | ~2 小时 | 🟡 中 | ✅ 完成 |
| 9.5 DNS 解析 | 1.5 天 | ~1 小时 | 🟡 中 | ✅ 完成 |
| 9.6 代码示例 | 5 天 | ~2 小时 | 🟡 中 | ✅ 完成 |
| 9.7 边界测试 | 5 天 | ~0.5 小时 | 🟡 中 | ✅ 完成 |
| **总计** | **~32 天** | **~27.5 小时** | | **100%** |

### 功能特性

**二进制处理**：
- `pack(format, ...values)` - 支持 C/S/s/N/V/Q/q/f/d/a/Z/x 格式
- `unpack(format, bytes)` - 解包二进制数据
- `buffer_new()` - 创建二进制缓冲区
- `buffer_write_uint8/16/32`, `buffer_read_uint8/16/32` 等 - 缓冲区读写
- `buffer_to_bytes/string`, `buffer_seek/tell/length/reset` - 缓冲区操作

**事件循环（IO 多路复用）**：
- `ev_loop_new()` - 创建事件循环（封装 epoll/kqueue）
- `ev_registry_new()` - 创建事件注册表
- `ev_attach(loop, registry)` - 附加注册表到循环
- `ev_run(loop)`, `ev_run_once(loop)`, `ev_stop(loop)` - 运行控制
- `ev_on_read/registry/fd, fn)`, `ev_on_write(...)`, `ev_on_accept(...)` - 事件注册
- `ev_on_timer(registry, interval, fn)`, `ev_on_timer_once(...)` - 定时器
- `ev_on_signal(registry, sig, fn)` - 信号处理
- `ev_off_xxx()` - 事件注销
- `ev_timer_now()` - 微秒级时间

**TCP/UDP/Unix Socket**：
- `net_tcp_listen(host, port)`, `net_tcp_connect(host, port)`, `net_tcp_accept(server)`
- `net_udp_bind(host, port)`, `net_udp_sendto(fd, data, host, port)`, `net_udp_recvfrom(fd, len)`
- `net_unix_listen(path)`, `net_unix_connect(path)`, `net_unix_accept(server)`
- `net_send(fd, data)`, `net_recv(fd, len)`, `net_close(fd)`
- `net_getsockname(fd)`, `net_getpeername(fd)`, `net_set_nonblock(fd)`

**DNS 解析**：
- `dns_resolve(domain)` - 解析所有 IP
- `dns_resolve_one(domain)` - 解析单个 IP

### 测试覆盖

- 9 个事件循环测试（注册表、定时器、信号）✅
- TCP/UDP/Unix Socket 测试 ✅
- DNS 解析测试 ✅
- Buffer + Network 集成测试 ✅
- 完整网络栈集成测试 ✅

### 9.6 代码示例（1 周）

> 完善示例代码库，覆盖所有主要功能模块，展示 JPL 网络编程能力。

| 批次 | 示例 | 说明 | 状态 |
|------|------|------|------|
| 9.6.1 | HTTP 服务器 | 基于 TCP 实现完整 HTTP 服务器 | ✅ 已完成 |
| 9.6.2 | WebSocket 协议 | 基于 TCP 实现 WebSocket 握手 | ✅ 已完成 |
| 9.6.3 | Redis 客户端 | RESP 协议实现 | ✅ 已完成 |
| 9.6.4 | 聊天室 | 多客户端广播（Unix Socket）| ✅ 已完成 |
| 9.6.5 | 文件上传服务器 | 展示二进制协议处理 | ✅ 已完成 |

### 9.7 边界测试补充（1 周）

> 补充网络/IO/事件循环的边界测试，提升代码健壮性。

| 测试类型 | 待测场景 | 优先级 | 状态 |
|----------|----------|--------|------|
| 网络超时 | 连接超时、读取超时、写入超时 | ⭐⭐⭐ | ✅ 完成 |
| 错误处理 | 网络中断、对端关闭、半开连接 | ⭐⭐⭐ | ✅ 完成 |
| 重连机制 | 自动重连、指数退避 | ⭐⭐ | ⏭️ 跳过（应用层实现） |
| 资源清理 | FD 泄漏、内存泄漏测试 | ⭐⭐ | ✅ 完成 |
| 并发测试 | 多连接并发、竞态条件 | ⭐ | ✅ 完成 |
| 大数据量 | 大文件传输、长连接稳定性 | ⭐ | ✅ 完成 |

> **注意**：性能测试暂不进行，待后续有明确需求时再做。

---

## Phase 10：语法增强 - 多行字符串与插值（1.5-2 周）

> 实现多行字符串和字符串插值语法，提升代码可读性和模板能力。
> 设计文档：[D25. 多行字符串语法](docs/DESIGN.md#d25-多行字符串语法)、[D26. 字符串插值语法](docs/DESIGN.md#d26-字符串插值语法)

### 10.1 多行字符串基础（3-4 天）✅ 完成

| 任务 | 预估工时 | 说明 | 决策 | 状态 |
|------|----------|------|------|------|
| Token 定义 | 0.5 天 | 新增 TRIPLE_SINGLE/TRIPLE_DOUBLE token | D25 | ✅ |
| Lexer 三引号扫描 | 1.5 天 | 识别 `"""` 和 `'''`，支持多行换行 | D25 | ✅ |
| Parser 多行字符串 | 1 天 | 解析多行字符串字面量 | D25 | ✅ |
| Compiler 支持 | 0.5 天 | 复用现有字符串编译逻辑 | | ✅ |
| 单元测试 | 0.5 天 | 覆盖纯文本多行字符串 | | ✅ |

### 10.2 字符串插值 - MVP 阶段（3-4 天）✅ 完成

| 任务 | 预估工时 | 说明 | 决策 | 状态 |
|------|----------|------|------|------|
| Lexer 插值扫描 | 1.5 天 | 双引号字符串识别 `#{$var}`，返回 token 序列 | D26 | ✅ |
| Parser 插值解析 | 1 天 | 构建 ConcatExpr 链 | D26 | ✅ |
| 编译器支持 | 0.5 天 | 复用现有 compileConcatExpr | | ✅ |
| 单元测试 | 0.5 天 | 简单变量插值测试 | | ✅ |

### 10.3 字符串插值 - 完整表达式（2-3 天）✅ 完成

| 任务 | 预估工时 | 说明 | 决策 | 状态 |
|------|----------|------|------|------|
| Lexer 表达式扫描 | 1 天 | 支持 `#{$obj.prop}`、`#{$arr[0]}` 等 | D26 | ✅ |
| Parser 表达式解析 | 1 天 | 递归解析复杂表达式 | D26 | ✅ |
| 单元测试 | 0.5 天 | 复杂表达式插值测试 | | ✅ |

### Phase 10 工时汇总

| 子阶段 | 预估工时 | 实际工时 | 优先级 | 状态 |
|--------|----------|----------|--------|------|
| 10.1 多行字符串 | 3-4 天 | ~2 小时 | 🟡 中 | ✅ 完成 |
| 10.2 插值 MVP | 3-4 天 | ~2 小时 | 🟡 中 | ✅ 完成 |
| 10.3 完整插值 | 2-3 天 | ~1 小时 | 🟢 低 | ✅ 完成 |
| **总计** | **8-11 天** | **~5 小时** | | **100%** |

### 功能特性

**多行字符串（D25）**：
- `'''...'''` 单引号三引号：纯文本，无插值
- `"""..."""` 双引号三引号：支持 `#{}` 插值
- 支持跨行、转义字符

**字符串插值（D26）**：
- 基本变量：`"Hello #{$name}!"`
- 对象属性：`"User: #{$user.name}"`
- 数组索引：`"First: #{$arr[0]}"`
- 算术运算：`"Sum: #{$a + $b}"`
- 三元表达式：`"Result: #{$score >= 60 ? 'Pass' : 'Fail'}"`
- 函数调用：`"Name: #{getName()}"`
- 字符串拼接：`"Full: #{$first .. ' ' .. $last}"`
- 转义机制：`\#{}` 输出字面量 `#{}`

---

## Phase 11：内置函数补全计划（参考 Jx9）

### 背景

Jx9 内置函数共 303 个，JPL 目前实现约 173 个。Phase 11 补充了 23 个常用函数。

### 11.1 第一批：字符串增强（高优先级）✅ 完成

| 函数 | 说明 | 文件 | 状态 |
|------|------|------|------|
| `substr_compare` | 比较两个字符串 | string.go | ✅ |
| `substr_count` | 统计子串出现次数 | string.go | ✅ |
| `str_repeat` | 重复字符串 | string.go | ✅ |
| `str_pad` | 字符串填充 | string.go | ✅ |
| `str_split` | 分割字符串为数组 | string.go | ✅ |
| `strrev` | 反转字符串 | string.go | ✅ |
| `htmlspecialchars` | HTML 实体转义 | string.go | ✅ |
| `htmlspecialchars_decode` | HTML 实体反转义 | string.go | ✅ |
| `strip_tags` | 移除 HTML 标签 | string.go | ✅ |
| `wordwrap` | 字符串换行 | string.go | ✅ |
| `strtolower` | 转为小写（toLower 别名） | string.go | ✅ |
| `strtoupper` | 转为大写（toUpper 别名） | string.go | ✅ |
| `chunk_split` | 分块插入字符 | string.go | ✅ |
| `md5` | MD5 哈希 | hash.go | ✅ |
| `sha1` | SHA1 哈希 | hash.go | ✅ |

### 11.2 第二批：数组排序/遍历（高优先级）✅ 完成

| 函数 | 说明 | 文件 | 状态 |
|------|------|------|------|
| `sort` | 升序排序（原数组） | functional.go | ✅ |
| `rsort` | 降序排序（原数组） | functional.go | ✅ |
| `usort` | 自定义排序 | array.go | ✅ 修复 stub |
| `key` | 获取当前键 | array.go | ✅ |
| `current` | 获取当前值 | array.go | ✅ |
| `each` | 获取当前键值对 | array.go | ✅ |
| `next` | 移动到下一个 | array.go | ✅ |
| `prev` | 移动到上一个 | array.go | ✅ |
| `end` | 移动到最后 | array.go | ✅ |
| `reset` | 重置指针 | array.go | ✅ |
| `extract` | 导入变量 | array.go | ✅ |
| `array_map` | 数组映射 | array.go | ✅ |
| `array_walk` | 数组遍历 | array.go | ✅ |

### 11.3 第三批：数学增强（中优先级）✅ 完成

| 函数 | 说明 | 文件 | 状态 |
|------|------|------|------|
| `rand_str` | 随机字符串 | math.go | ✅ |
| `getrandmax` | 随机数最大值 | math.go | ✅ |
| `round` | 四舍五入（需完善） | math.go | ✅ |
| `dechex` | 十进制转十六进制 | math.go | ✅ |
| `decoct` | 十进制转八进制 | math.go | ✅ |
| `decbin` | 十进制转二进制 | math.go | ✅ |
| `hexdec` | 十六进制转十进制 | math.go | ✅ |
| `bindec` | 二进制转十进制 | math.go | ✅ |
| `octdec` | 八进制转十进制 | math.go | ✅ |
| `base_convert` | 进制转换 | math.go | ✅ |

### 11.4 第四批：文件 IO 增强（中优先级）✅ 完成

| 函数 | 说明 | 文件 | 状态 |
|------|------|------|------|
| `fseek` | 文件指针移动 | fileio.go | ✅ 已实现 |
| `ftell` | 获取文件指针位置 | fileio.go | ✅ 已实现 |
| `rewind` | 重置文件指针 | fileio.go | ✅ 已实现 |
| `ftruncate` | 截断文件 | fileio.go | ✅ 已实现 |
| `fgets` | 已实现（Phase 8.2） | — | — |
| `fgetcsv` | 读取 CSV 行 | fileio.go | ✅ 已实现 |
| `file_get_contents` | 读取文件内容 | fileio.go | ✅ |
| `file_put_contents` | 写入文件内容 | fileio.go | ✅ |
| `copy` | 复制文件 | fileio.go | ✅ |
| `readfile` | 读取并输出文件 | fileio.go | ✅ |
| `pathinfo` | 路径信息 | fileio.go | ✅ |

### 11.5 第五批：Hash 增强（中优先级）✅ 完成

| 函数 | 说明 | 文件 | 状态 |
|------|------|------|------|
| `md5_file` | 文件 MD5 | hash.go | ✅ |
| `sha1_file` | 文件 SHA1 | hash.go | ✅ |
| `crc32` | CRC32 校验 | hash.go | ✅ |

### 11.6 第六批：系统/其他（低优先级）✅ 完成

| 函数 | 说明 | 文件 | 状态 |
|------|------|------|------|
| `disk_free_space` | 磁盘可用空间 | system.go | ✅ |
| `disk_total_space` | 磁盘总空间 | system.go | ✅ |
| `fileatime` | 文件访问时间 | system.go | ✅ |
| `filemtime` | 文件修改时间 | system.go | ✅ |
| `filectime` | 文件创建时间 | system.go | ✅ |
| `touch` | 修改文件时间 | system.go | ✅ |
| `umask` | 设置文件掩码 | system.go | ✅ |
| `getpid` | 获取进程 ID | system.go | ✅ |
| `getuid` | 获取用户 ID | system.go | ✅ |
| `getgid` | 获取组 ID | system.go | ✅ |
| `uname` | 系统信息 | system.go | ✅ |
| `dirname` | 目录名 | fileio.go | ✅ |
| `basename` | 文件名 | fileio.go | ✅ |
| `pathinfo` | 路径信息 | fileio.go | ✅ |

### 实际工时

| 批次 | 预估工时 | 实际工时 | 函数数 | 状态 |
|------|----------|----------|--------|------|
| 11.1 字符串增强 | 3 天 | 已完成 | ~15 | ✅ |
| 11.2 数组排序/遍历 | 3 天 | 已完成 | ~13 | ✅ |
| 11.3 数学增强 | 2 天 | 已完成 | ~10 | ✅ |
| 11.4 文件 IO 增强 | 2 天 | ~1 小时 | 9 | ✅ |
| 11.5 Hash 增强 | 1 天 | ~0.5 小时 | 2 | ✅ |
| 11.6 系统/其他 | 2 天 | ~1 小时 | 11 | ✅ |
| **总计** | **~13 天** | **~0.5 天** | **~65** | **✅ 完成** |

---

## Phase 12：进程 API

> 参考 PHP/Python/Node.js 的进程管理 API，为 JPL 提供完整的系统进程操作能力。
> 设计文档：[D32. 进程 API 设计](docs/DESIGN.md#d32-进程-api-设计)

### 12.1 P0 - 核心功能 ✅ 完成

| 函数 | 说明 | 状态 |
|------|------|------|
| `exec($cmd)` | 执行命令，返回输出字符串 | ✅ |
| `system($cmd)` | 执行命令，返回退出码 | ✅ |
| `shell_exec($cmd)` | 执行命令并返回完整输出 | ✅ |
| `getenv($name)` | 获取环境变量 | ✅ |
| `setenv($name, $val)` | 设置环境变量 | ✅ |
| `getppid()` | 获取父进程 ID | ✅ |
| `tmpdir()` | 获取系统临时目录 | ✅ |
| `hostname()` | 获取主机名 | ✅ |

### 12.2 P1 - 常用功能 ✅ 完成

| 函数 | 说明 | 状态 |
|------|------|------|
| `proc_open($cmd, $opts)` | 执行命令，返回进程管道对象 | ✅ |
| `proc_close($proc)` | 关闭进程管道 | ✅ |
| `proc_wait($proc)` | 等待进程结束，返回退出码 | ✅ |
| `proc_status($proc)` | 获取进程状态 | ✅ |
| `getlogin()` | 获取当前登录用户名 | ✅ |
| `usleep($us)` | 暂停执行（微秒） | ✅ |
| `putenv($expr)` | 设置环境变量（"KEY=VALUE"格式） | ✅ |

### 12.3 P2 - 进阶功能 ✅ 完成

| 函数 | 说明 | 状态 |
|------|------|------|
| `spawn($cmd, $args)` | 创建子进程（不等待） | ✅ |
| `kill($pid, $signal)` | 向进程发送信号 | ✅ |
| `waitpid($proc)` | 等待指定子进程 | ✅ |
| `fork()` | 创建子进程（Unix） | ✅ |
| `pipe()` | 创建管道对 | ✅ |

### 12.4 P3 - 高级功能 ✅ 完成

| 函数 | 说明 | 状态 |
|------|------|------|
| `sigwait($sigs)` | 阻塞等待信号 | ✅ |

### P3 跳过的函数

| 函数 | 原因 |
|------|------|
| `signal()` | 已有 `$registry.on_signal()` 替代 |
| `execv()` | 风险高，调用后进程被替换 |
| `daemon()` | 实现复杂，必要性低 |
| `nice()` | 必要性低 |

### Phase 12 工时汇总

| 子阶段 | 预估工时 | 实际工时 | 函数数 | 状态 |
|--------|----------|----------|--------|------|
| P0 核心功能 | 1-2 天 | ~1 小时 | 8 | ✅ 完成 |
| P1 常用功能 | 1-2 天 | ~1 小时 | 7 | ✅ 完成 |
| P2 进阶功能 | 2-3 天 | ~1 小时 | 5 | ✅ 完成 |
| P3 高级功能 | 按需 | ~0.5 小时 | 1 | ✅ 完成 |
| **总计** | **4-7 天** | **~3.5 小时** | **21** | **✅ 完成** |

---

## Phase 13：TLS/SSL 模块（1-2 周）

> 实现 TLS/SSL 加密通信模块，为 HTTPS 提供基础支持。
> 设计文档：[D34. TLS/SSL 模块设计](docs/DESIGN.md#d34-tlsssl-模块设计)

### 13.1 TLS 连接管理（第 1 周）

| 任务 | 预估工时 | 说明 | 状态 |
|------|----------|------|------|
| `tls_connect()` | 0.5 天 | 建立 TLS 客户端连接 | ✅ 完成 |
| `tls_listen()` | 0.5 天 | 创建 TLS 服务端监听 | ✅ 完成 |
| `tls_accept()` | 0.5 天 | 接受 TLS 连接 | ✅ 完成 |
| `tls_close()` | 0.5 天 | 关闭 TLS 连接 | ✅ 完成 |
| 证书验证 | 0.5 天 | CA 证书链验证 | ✅ 完成 |
| 单元测试 | 0.5 天 | TLS 连接测试 | ✅ 完成 |

### 13.2 TLS 数据传输（第 1 周）

| 任务 | 预估工时 | 说明 | 状态 |
|------|----------|------|------|
| `tls_send()` | 0.5 天 | 发送加密数据 | ✅ 完成 |
| `tls_recv()` | 0.5 天 | 接收解密数据 | ✅ 完成 |
| TLS 信息函数 | 0.5 天 | 获取加密套件、版本、证书信息 | ✅ 完成 |
| mTLS 支持 | 0.5 天 | 客户端证书双向认证 | ✅ 完成 |
| 单元测试 | 0.5 天 | 数据传输测试 | ✅ 完成 |

### Phase 13 工时汇总

| 子阶段 | 预估工时 | 实际工时 | 函数数 | 状态 |
|--------|----------|----------|--------|------|
| 13.1 连接管理 | 2.5 天 | ~0.75 小时 | 4 | ✅ 完成 |
| 13.2 数据传输 | 2 天 | ~0.75 小时 | 7 | ✅ 完成 |
| **总计** | **~4.5 天** | **~1.5 小时** | **11** | **✅ 完成** |

### 新增函数列表

| 类别 | 函数 | 说明 |
|------|------|------|
| 连接 | tls_connect | 建立 TLS 连接 |
| 连接 | tls_listen | TLS 服务端监听 |
| 连接 | tls_accept | 接受 TLS 连接 |
| 连接 | tls_close | 关闭连接 |
| 传输 | tls_send | 发送加密数据 |
| 传输 | tls_recv | 接收解密数据 |
| 信息 | tls_get_cipher | 获取加密套件 |
| 信息 | tls_get_version | 获取 TLS 版本 |
| 信息 | tls_get_cert_info | 获取证书信息 |
| 高级 | tls_set_cert | 设置客户端证书 (mTLS) |

---

## Phase 14：HTTP Client 模块（1-2 周）

> 实现高级 HTTP 客户端功能，支持 HTTPS、JSON、Form、认证等。
> 设计文档：[D35. HTTP Client 模块设计](docs/DESIGN.md#d35-http-client-模块设计)

### 14.1 基础 HTTP 请求（第 1 周）

| 任务 | 预估工时 | 说明 | 状态 |
|------|----------|------|------|
| `http_get()` | 0.5 天 | GET 请求 | ✅ 完成 |
| `http_post()` | 0.5 天 | POST 请求 | ✅ 完成 |
| `http_request()` | 0.5 天 | 通用 HTTP 请求 | ✅ 完成 |
| HTTP 响应对象 | 0.5 天 | 统一响应结构 | ✅ 完成 |
| 单元测试 | 0.5 天 | 基础请求测试 | ✅ 完成 |

### 14.2 高级 HTTP 功能（第 1-2 周）

| 任务 | 预估工时 | 说明 | 状态 |
|------|----------|------|------|
| `http_put/delete/head/patch()` | 0.5 天 | 其他 HTTP 方法 | ✅ 完成 |
| Options 参数解析 | 0.5 天 | headers, timeout, auth 等 | ✅ 完成 |
| JSON/Form 自动处理 | 0.5 天 | 自动序列化/反序列化 | ✅ 完成 |
| HTTPS 自动支持 | 0.5 天 | URL 检测自动使用 TLS | ✅ 完成 |
| 重定向处理 | 0.5 天 | follow_redirects, max_redirects | ✅ 完成 |
| 代理支持 | 0.5 天 | HTTP/HTTPS 代理 | ✅ 完成 |
| 单元测试 | 0.5 天 | 高级功能测试 | ✅ 完成 |

### Phase 14 工时汇总

| 子阶段 | 预估工时 | 实际工时 | 函数数 | 状态 |
|--------|----------|----------|--------|------|
| 14.1 基础请求 | 2.5 天 | ~2 小时 | 7 | ✅ 完成 |
| 14.2 高级功能 | 3 天 | ~1 小时 | 1 | ✅ 完成 |
| **总计** | **~5.5 天** | **~3 小时** | **8** | **✅ 完成** |

### 新增函数列表

| 类别 | 函数 | 说明 |
|------|------|------|
| 简单请求 | http_get | GET 请求 |
| 简单请求 | http_post | POST 请求 |
| 简单请求 | http_put | PUT 请求 |
| 简单请求 | http_delete | DELETE 请求 |
| 简单请求 | http_head | HEAD 请求 |
| 简单请求 | http_patch | PATCH 请求 |
| 通用请求 | http_request | 通用 HTTP 请求 |

---

## Phase 15：正则与加密模块

> 实现正则表达式、加密哈希、编码解码功能。

### 15.1 正则表达式（第 1 周）✅ 完成

| 函数 | 说明 | 状态 |
|------|------|------|
| `re_match($pattern, $string)` | 正则匹配 | ✅ |
| `re_search($pattern, $string)` | 查找第一个 | ✅ |
| `re_findall($pattern, $string)` | 全局匹配 | ✅ |
| `re_sub($pattern, $replace, $string)` | 正则替换 | ✅ |
| `re_split($pattern, $string)` | 正则分割 | ✅ |
| `re_groups($pattern, $string)` | 捕获组 | ✅ |

### 15.2 加密模块（第 1 周）✅ 完成

| 函数 | 说明 | 状态 |
|------|------|------|
| `sha256($data)` | SHA256 哈希 | ✅ |
| `sha512($data)` | SHA512 哈希 | ✅ |
| `hmac_sha256($key, $data)` | HMAC-SHA256 | ✅ |
| `hmac_sha512($key, $data)` | HMAC-SHA512 | ✅ |
| `aes_encrypt($data, $key)` | AES-256-GCM 加密 | ✅ |
| `aes_decrypt($data, $key)` | AES-256-GCM 解密 | ✅ |
| `bcrypt_hash($password)` | bcrypt 哈希 | ✅ |
| `bcrypt_verify($password, $hash)` | bcrypt 验证 | ✅ |
| `bcrypt_cost($hash)` | 获取 cost 值 | ✅ |
| `ed25519_generate_key()` | Ed25519 密钥对 | ✅ |
| `ed25519_sign($msg, $privkey)` | Ed25519 签名 | ✅ |
| `ed25519_verify($msg, $sig, $pubkey)` | Ed25519 验签 | ✅ |
| `ed25519_public_key($privkey)` | Ed25519 公钥 | ✅ |
| `x25519_generate_key()` | X25519 密钥对 | ✅ |
| `x25519_shared_secret($privkey, $pubkey)` | 共享密钥 | ✅ |
| `x25519_public_key($privkey)` | X25519 公钥 | ✅ |
| `rsa_generate_key($bits)` | RSA 密钥对 | ✅ |
| `rsa_encrypt($msg, $pubkey)` | RSA 加密 | ✅ |
| `rsa_decrypt($cipher, $privkey)` | RSA 解密 | ✅ |
| `rsa_sign($msg, $privkey)` | RSA 签名 | ✅ |
| `rsa_verify($msg, $sig, $pubkey)` | RSA 验签 | ✅ |
| `rsa_public_key($privkey)` | RSA 公钥 | ✅ |

### 15.3 编码增强（第 1 周）✅ 完成

| 函数 | 说明 | 状态 |
|------|------|------|
| `bin2hex($str)` | 二进制转十六进制 | ✅ |
| `hex2bin($str)` | 十六进制转二进制 | ✅ |
| `ord($char)` | 字符转 ASCII 码 | ✅ |
| `chr($code)` | ASCII 码转字符 | ✅ |

### 15.4 压缩模块（第 2 周）✅ 完成

> 使用 Go 标准库实现：`compress/gzip`、`compress/zlib`、`github.com/andybalholm/brotli`、`archive/zip`、`archive/tar`

#### 15.4.1 gzip 压缩

| 函数 | 说明 | 状态 |
|------|------|------|
| `gzencode($data)` | gzip 压缩 | ✅ |
| `gzdecode($data)` | gzip 解压 | ✅ |
| `gzfile($filename)` | 读取 gzip 文件 | ✅ |
| `writegzfile($filename, $data)` | 写入 gzip 文件 | ✅ |
| `gzopen($filename, $mode)` | 打开 gzip 文件 | ✅ |
| `gzread($gz, $length)` | 读取 gzip 数据 | ✅ |
| `gzwrite($gz, $data)` | 写入 gzip 数据 | ✅ |
| `gzclose($gz)` | 关闭 gzip 文件 | ✅ |
| `gzgets($gz)` | 读取一行 | ✅ |
| `gzeof($gz)` | 检查是否到达末尾 | ✅ |

#### 15.4.2 zlib 压缩

| 函数 | 说明 | 状态 |
|------|------|------|
| `zlib_encode($data)` | zlib 压缩 | ✅ |
| `zlib_decode($data)` | zlib 解压 | ✅ |
| `deflate($data)` | deflate 压缩 | ✅ |
| `inflate($data)` | deflate 解压 | ✅ |

#### 15.4.3 brotli 压缩

| 函数 | 说明 | 状态 |
|------|------|------|
| `brotli_encode($data)` | brotli 压缩 | ✅ |
| `brotli_decode($data)` | brotli 解压 | ✅ |
| `brotli_compress_file($src, $dest)` | 压缩文件 | ✅ |
| `brotli_decompress_file($src, $dest)` | 解压文件 | ✅ |
| `brotli_open($filename, $mode)` | 打开 brotli 文件 | ✅ |
| `brotli_read($handle, $length)` | 读取数据 | ✅ |
| `brotli_write($handle, $data)` | 写入数据 | ✅ |
| `brotli_close($handle)` | 关闭文件 | ✅ |

#### 15.4.4 ZIP 归档

| 函数 | 说明 | 状态 |
|------|------|------|
| `zip_open($filename)` | 打开 zip 文件 | ✅ |
| `zip_read($zip)` | 读取 zip 条目 | ✅ |
| `zip_entry_name($entry)` | 获取 zip 条目名 | ✅ |
| `zip_entry_filesize($entry)` | 获取压缩前大小 | ✅ |
| `zip_entry_compressedsize($entry)` | 获取压缩后大小 | ✅ |
| `zip_entry_read($entry, $len)` | 读取 zip 条目内容 | ✅ |
| `zip_entry_close($entry)` | 关闭 zip 条目 | ✅ |
| `zip_close($zip)` | 关闭 zip 文件 | ✅ |
| `zip_create($filename, $entries)` | 创建 zip 文件 | ✅ |

#### 15.4.5 TAR 归档

| 函数 | 说明 | 状态 |
|------|------|------|
| `tar_open($filename)` | 打开 tar 文件 | ✅ |
| `tar_read($tar)` | 读取 tar 条目 | ✅ |
| `tar_entry_name($entry)` | 获取 tar 条目名 | ✅ |
| `tar_entry_size($entry)` | 获取条目大小 | ✅ |
| `tar_entry_isdir($entry)` | 检查是否目录 | ✅ |
| `tar_entry_read($entry, $len)` | 读取 tar 条目内容 | ✅ |
| `tar_entry_close($entry)` | 关闭 tar 条目 | ✅ |
| `tar_close($tar)` | 关闭 tar 文件 | ✅ |
| `tar_create($filename, $entries)` | 创建 tar 文件 | ✅ |

### Phase 15 工时汇总

| 子阶段 | 预估工时 | 实际工时 | 函数数 | 状态 |
|--------|----------|----------|--------|------|
| 15.1 正则表达式 | 2 天 | ~2 小时 | 8 | ✅ 完成 |
| 15.2 加密模块 | 3 天 | ~2 小时 | 22 | ✅ 完成 |
| 15.3 编码增强 | 1 天 | ~0.5 小时 | 4 | ✅ 完成 |
| 15.4 压缩模块 | 3 天 | ~4 小时 | 41 | ✅ 完成 |
| **总计** | **~9 天** | **~7.5 小时** | **59** | **100%** |

---

## Phase 16：管道运算符（1-2 天）

> 实现管道运算符，支持函数式编程风格的数据流处理。
> 设计文档：[D41. 管道运算符设计](docs/DESIGN.md#d41-管道运算符设计)

### 16.1 正向管道 `|>` ✅ 完成

| 任务 | 预估工时 | 说明 | 状态 |
|------|----------|------|------|
| Token 定义 | 0.5 天 | 新增 PIPE_FWD token | ✅ 完成 |
| Lexer 识别 | 0.5 天 | 识别 `\|>` 两字符组合 | ✅ 完成 |
| AST 节点 | 0.5 天 | 新增 PipeExpr 节点 | ✅ 完成 |
| Parser 解析 | 0.5 天 | 左结合优先级解析 | ✅ 完成 |
| Compiler 编译 | 0.5 天 | 编译为函数调用 | ✅ 完成 |

### 16.2 反向管道 `<|` ✅ 完成

| 任务 | 预估工时 | 说明 | 状态 |
|------|----------|------|------|
| Token 定义 | 0.5 天 | 新增 PIPE_BWD token | ✅ 完成 |
| Lexer 识别 | 0.5 天 | 识别 `<\|` 两字符组合 | ✅ 完成 |
| Parser 解析 | 0.5 天 | 右结合优先级解析 | ✅ 完成 |
| Compiler 编译 | 0.5 天 | 编译为函数调用（末尾参数） | ✅ 完成 |

### Phase 16 工时汇总

| 子阶段 | 预估工时 | 实际工时 | 状态 |
|--------|----------|----------|------|
| 16.1 正向管道 | 3 天 | ~0.5 小时 | ✅ 完成 |
| 16.2 反向管道 | 2 天 | ~0.5 小时 | ✅ 完成 |
| **总计** | **~5 天** | **~1 小时** | **✅ 完成** |

### 功能特性

**正向管道 `|>`**（左结合）：
- `a |> f(b,c)` = `f(a, b, c)` — 左侧值作为首个参数
- `a |> f` — 返回函数引用（不调用）
- `a |> f()` — 调用函数，a 作为首个参数

**反向管道 `<|`**（右结合）：
- `f(b,c) <| a` = `f(b, c, a)` — 右侧值作为末尾参数
- `f <| a` — 调用函数，a 作为参数

**链式调用**：
- `a |> f |> g` = `g(f(a))` — 正向链式
- `f <| g <| a` = `f(g(a))` — 反向链式

---

## Phase 17：match/case 语法

> 修复 match/case 语法 bug 并完善实现。
> 设计文档：[D42. match/case 语法设计](docs/DESIGN.md#d42-matchcase-语法设计)

### 17.1 Bug 修复

| 任务 | 说明 | 状态 |
|------|------|------|
| 编译器缺失 case | compileStmt() 缺少 MatchStmt 处理，语句被静默跳过 | ✅ 完成 |
| Guard 解析修复 | 解析 guard 条件时使用错误的 token 类型 | ✅ 完成 |

### 17.2 功能实现

| 任务 | 说明 | 状态 |
|------|------|------|
| 字面量匹配 | 支持数字、字符串、布尔值匹配 | ✅ 完成 |
| 标识符绑定 | 匹配时绑定到变量 | ✅ 完成 |
| OR 模式 | 使用 `||` 连接多个模式 | ✅ 完成 |
| Guard 条件 | `if` 关键字添加额外匹配条件 | ✅ 完成 |
| Match 表达式 | 支持返回值 | ✅ 完成 |

### 17.3 范围语法设计

| 语法 | 说明 | 状态 |
|------|------|------|
| `...` | 半开区间 [start, end) | ✅ 完成 |
| `..=` | 闭区间 [start, end] | ✅ 完成 |

### 语法示例

```jpl
// match 语句
match ($status) {
    case 200: puts "OK"
    case 404: puts "Not Found"
    case 500: puts "Server Error"
    case _: puts "Unknown"
}

// match 表达式
$result = match ($code) {
    case 200: "success"
    case 404: "not found"
    case _: "unknown"
}

// 带 guard 条件
match ($score) {
    case $x if $x >= 90: "A"
    case $x if $x >= 80: "B"
    case $x if $x >= 60: "C"
    case _: "F"
}

// OR 模式
match ($day) {
    case "Saturday", "Sunday": "Weekend"
    case _: "Weekday"
}
```

## Phase 18：正则字面量语法 ✅ 完成

> 为 JPL 添加正则字面量语法 `#/pattern/flags#` 和 `=~` 匹配运算符，支持 match/case 正则模式。
> 设计文档：[D39. 正则字面量语法设计](docs/DESIGN.md#d39-正则字面量语法设计)

### 18.1 Token + Value 类型定义（实际 ~1h）

| 任务 | 说明 | 文件 | 状态 |
|------|------|------|------|
| REGEX token | 新增 REGEX token 类型 | `token/token.go` | ✅ 完成 |
| MATCH_EQ token | 新增 `=~` 运算符 token | `token/token.go` | ✅ 完成 |
| regexValue 类型 | 新增值类型，包装 Go regexp.Regexp | `engine/value.go` | ✅ 完成 |

### 18.2 Lexer 正则字面量扫描（实际 ~2h）

| 任务 | 说明 | 文件 | 状态 |
|------|------|------|------|
| `#/` 扫描 | scanRegexLiteral 函数，处理转义和 flags | `lexer/lexer.go` | ✅ 完成 |
| `=~` 识别 | scanOperator 中识别两字符组合 | `lexer/lexer.go` | ✅ 完成 |
| 编译期验证 | 空模式、无效正则、未知 flag 报错 | `lexer/lexer.go` | ✅ 完成 |

### 18.3 AST + Parser（实际 ~2h）

| 任务 | 说明 | 文件 | 状态 |
|------|------|------|------|
| RegexLiteral 节点 | 正则字面量表达式 AST | `parser/ast.go` | ✅ 完成 |
| RegexPattern 节点 | match/case 正则模式 AST | `parser/ast.go` | ✅ 完成 |
| parseRegexLiteral | 注册 prefix 解析函数 | `parser/parser.go` | ✅ 完成 |
| =~ infix | 注册 infix 解析函数 | `parser/parser.go` | ✅ 完成 |
| parsePattern 扩展 | 支持 REGEX token + as 绑定 | `parser/parser.go` | ✅ 完成 |

### 18.4 Compiler + VM 正则执行（实际 ~2h）

| 任务 | 说明 | 文件 | 状态 |
|------|------|------|------|
| =~ 编译 | 编译为 OP_REGEX_MATCH | `engine/compiler.go` | ✅ 完成 |
| regex case 编译 | RegexPattern → 字节码生成 | `engine/compiler.go` | ✅ 完成 |
| 捕获组绑定 | as $var → re_groups_raw 提取 | `engine/compiler.go` | ✅ 完成 |
| VM 执行 | opRegexMatch + OP_REGEX_MATCH | `engine/vm.go` | ✅ 完成 |

### 18.5 测试 + 边界修复（实际 ~2h）

| 任务 | 说明 | 状态 |
|------|------|------|
| =~ 运算符测试 | 6 个基本匹配场景 | ✅ 完成 |
| 编译期错误测试 | 空模式、缺少 #、无效正则 | ✅ 完成 |
| Flags 测试 | i/m/s/im 组合 | ✅ 完成 |
| match/case 正则测试 | 4 个分支匹配场景 | ✅ 完成 |
| 捕获组绑定测试 | 3 个 as $m 绑定场景 | ✅ 完成 |
| 混合测试 | 字面量 + 正则混合 | ✅ 完成 |

### Phase 18 工时汇总

| 子阶段 | 预估工时 | 实际工时 | 状态 |
|--------|----------|----------|------|
| 18.1 Token + Value | 3h | ~1h | ✅ 完成 |
| 18.2 Lexer | 4h | ~2h | ✅ 完成 |
| 18.3 AST + Parser | 4h | ~2h | ✅ 完成 |
| 18.4 Compiler + VM | 6h | ~2h | ✅ 完成 |
| 18.5 测试 | 4h | ~2h | ✅ 完成 |
| **总计** | **~21h (2-3天)** | **~9h** | **✅ 完成** |

### 语法速览

```jpl
// 字面量
$re = #/^\d{3}-\d{4}/#

// =~ 运算符
if ($input =~ #/\d+/#) { puts "has digits" }

// match/case 正则
match ($input) {
    case #/^quit$/:       exit(0)
    case #/^set (\w+)=(.+)$/# as $m:
        $config[$m[1]] = $m[2]
    case #/^hello$/i#:    puts "greeting"
    case _:               puts "unknown"
}
```

---

## Phase 19：工具链完善

> 完善 JPL 开发工具链，提升开发体验。

### 19.1 代码格式化 `jpl fmt`（预估 ~3 天）⭐⭐⭐

| 任务 | 说明 | 状态 |
|------|------|------|
| 缩进规则 | 统一 4 空格缩进 | 🔲 待开始 |
| 大括号风格 | K&R 风格（左大括号不换行） | 🔲 待开始 |
| 运算符空格 | 统一运算符两侧空格 | 🔲 待开始 |
| 字符串引号 | 统一使用双引号 | 🔲 待开始 |
| 行宽限制 | 默认 120 字符 | 🔲 待开始 |
| CLI 集成 | `jpl fmt [file]` 命令 | 🔲 待开始 |

### 19.2 静态分析 `jpl lint`（预估 ~3 天）⭐⭐ ✅ 完成

| 任务 | 说明 | 状态 |
|------|------|------|
| 未使用变量 | 检测声明但未使用的变量 | ✅ |
| 未定义变量 | 使用未声明的变量 | ✅ |
| 死代码检测 | 不可达代码 | ✅ |
| CLI 集成 | `jpl lint [file]` 命令 | ✅ |

### 19.3 LSP 支持（预估 ~1 周）⭐⭐

| 任务 | 说明 | 状态 |
|------|------|------|
| LSP Server | Language Server Protocol 实现 | 🔲 待开始 |
| 语法高亮 | TextMate 语法定义 | 🔲 待开始 |
| 自动补全 | 函数名、变量名补全 | 🔲 待开始 |
| 跳转定义 | 函数/变量定义跳转 | 🔲 待开始 |
| VS Code 插件 | VS Code 扩展 | 🔲 待开始 |

### 19.4 调试器 `jpl debug`（预估 ~1 周）⭐

| 任务 | 说明 | 状态 |
|------|------|------|
| DAP 协议 | Debug Adapter Protocol 实现 | 🔲 待开始 |
| 断点 | 源码行断点 | 🔲 待开始 |
| 单步执行 | step in/out/over | 🔲 待开始 |
| 变量查看 | 运行时变量值 | 🔲 待开始 |

### 19.5 包管理器 `jpl add/remove`（预估 ~2 周）⭐ ✅ 完成

> 设计文档：[docs/PACKAGE_MANAGER.md](docs/PACKAGE_MANAGER.md)

| 任务 | 说明 | 状态 |
|------|------|------|
| `jpl.json` 清单 | 项目依赖声明 | ✅ 完成 |
| `jpl add` | git 克隆 + 安装到 jpl_modules/ | ✅ 完成 |
| `jpl remove` | 移除依赖 + 清理清单 | ✅ 完成 |
| `jpl install` | 安装全部依赖（并行克隆） | ✅ 完成 |
| 传递依赖 | DFS 解析 + 循环检测 | ✅ 完成 |
| 锁文件 | 复用现有 LockFile，锁定 commit hash | ✅ 完成 |
| `jpl list` | 列出依赖树 | ✅ 完成 |
| 全局缓存 | ~/.jpl/packages/ 缓存 git 仓库 | ✅ 完成 |
| 版本约束 | semver 语义化版本（^, ~, >= 等） | ✅ 完成 |
| `jpl update/outdated` | 依赖更新和过时检查 | ✅ 完成 |
| `jpl init` | 项目初始化 | ✅ 完成 |
| Resolver 集成 | `--resolve` 完整依赖解析 | ✅ 完成 |
| 并行安装 | goroutine 并行克隆 + `--jobs/-j` 标志 | ✅ 完成 |

### 19.6 任务系统 `jpl task`（新增）⭐ ✅ 完成

| 任务 | 说明 | 状态 |
|------|------|------|
| `tasks` 字段 | jpl.json 中定义任务 | ✅ 完成 |
| 简单格式 | `"name": "command"` | ✅ 完成 |
| 复杂格式 | `"name": {"cmd": "...", "deps": [...]}` | ✅ 完成 |
| 依赖解析 | 拓扑排序 + 循环检测 + 去重 | ✅ 完成 |
| CLI 命令 | `jpl task <name>`, `--list`, `--dry-run` | ✅ 完成 |

### 19.7 文档生成 `jpl doc`（暂不实现）⭐

| 任务 | 说明 | 状态 |
|------|------|------|
| 注释解析 | 解析函数/模块文档注释 | 🔲 搁置 |
| HTML 输出 | 生成 HTML 文档站点 | 🔲 搁置 |
| CLI 集成 | `jpl doc [file]` 命令 | 🔲 搁置 |

### 暂不计划

| 功能 | 原因 |
|------|------|
| **jpl doc** | 核心功能已完整，文档生成投入产出比低 |
| **LSP 支持** | lint + fmt 已覆盖核心开发体验 |
| **调试器 (DAP)** | --debug + REPL 已提供基本调试能力 |
| **性能测试/优化** | 用户暂不需求，后续根据实际场景再做 |
| **协程/async-await** | 当前事件循环模型已足够 |

---

## Phase 20：间接变量引用（反引号语法）

> 使用反引号 `` ` `` 实现运行时间接变量引用，类似宏展开语义但延迟到运行时求值。
> 参考语言：PHP 可变变量 `$$x`、Perl 符号引用 `${$x}`、Bash 间接展开 `${!x}`

### 20.1 语法设计

```jpl
a = "hello"
$x = "world"

// 基本间接引用
x = "a"
puts `x      // → "hello"（读取 x 的值 "a"，再查找变量 a）

// 引用带 $ 前缀的变量名
x = "$x"
puts `x      // → "world"（读取 x 的值 "$x"，再查找变量 $x）

// 链式间接引用
b = "a"
x = "b"
puts `x      // → "hello"（`x → "b" → `b → "a" → "hello"）
```

### 20.2 实现方案

| 任务 | 说明 | 文件 | 状态 |
|------|------|------|------|
| BACKTICK token | 新增 `` ` `` token 类型 | `token/token.go` | 🔲 待实现 |
| Lexer 识别 | 识别反引号字符 | `lexer/lexer.go` | 🔲 待实现 |
| IndirectRef 节点 | AST 间接引用节点 | `parser/ast.go` | 🔲 待实现 |
| 前缀解析 | Pratt parser 解析 `` `identifier`` | `parser/parser.go` | 🔲 待实现 |
| OP_GET_INDIRECT | 新增间接查找操作码 | `engine/bytecode.go` | 🔲 待实现 |
| 编译逻辑 | 编译间接引用为字节码 | `engine/compiler.go` | 🔲 待实现 |
| VM 执行 | 运行时按名称查找变量 | `engine/vm.go` | 🔲 待实现 |
| 单元测试 | 覆盖基本/链式/边界场景 | `engine/vm_test.go` | 🔲 待实现 |

### 20.3 设计要点

- **运行时求值**：非编译期宏展开，支持动态变量名
- **作用域查找**：按局部→全局顺序查找
- **错误处理**：变量不存在时返回 null 或运行时错误
- **性能考量**：需字符串查找变量名，比直接引用稍慢

### 20.4 参考语言对比

| 语言 | 语法 | 变量名含 `$` | 求值时机 |
|------|------|-------------|----------|
| PHP | `$$x` | ❌ `$` 是 sigil | 运行时 |
| Perl | `${$x}` | ❌ 同上 | 运行时 |
| Bash | `${!x}` | ❌ `$` 是前缀 | 运行时 |
| **JPL** | `` `x`` | ✅ | 运行时 |

---

*项目已进入维护模式，核心功能完整。*
