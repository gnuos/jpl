# Changelog

## v0.9.5 (2026-04-04)

### 新增

#### 更好的错误消息
- 运行时错误显示行号 + 源码上下文（箭头标记 + 前后 2 行）
- Program 结构体新增 `Source` 和 `SourceLines` 字段存储源代码
- 编译器在编译时自动存储源代码
- VM 在错误发生时自动附加当前行号到 RuntimeError
- RuntimeError 新增 `FormatWithContext()` 方法
- CLI（run/eval/repl）全部更新为使用新的错误格式
- 新增 3 个测试用例

**输出示例**：
```
runtime error at line 3: something went wrong
   1 | fn greet() {
   2 |     $msg = "hello"
 → 3 |     throw "something went wrong"
   4 | }
   5 | 
```

#### REPL 多行续输
- 括号/引号平衡检测：输入未闭合的 `(`、`{`、`[`、`"`、`'`、`'''`、`"""` 时自动进入多行模式
- 提示符从 `> ` 动态切换为 `... `（go-prompt `WithPrefixCallback`）
- 空行提交多行代码
- 支持转义字符、三引号、注释中的括号忽略
- 新增 7 个测试用例

#### :doc 完整函数签名
- 41 个 stdlib 模块全覆盖，500+ 内置函数签名
- 包含参数名、可选参数、返回值类型和简要描述
- 示例：`map(array_or_range, fn(element) → newValue) → array`
- 新增 4 个测试用例

---

## v0.9.4 (2026-04-03)

### 新增

#### 尾调用优化 (TCO)
- 自递归尾调用栈帧复用，消除递归调用栈增长
- 编译器自动检测尾位置调用（`return func(args)`），发出 `OP_TAIL_CALL` 指令
- VM 通过闭包身份匹配检测自递归，原地更新参数并跳转执行
- 非自递归尾调用（如 `return $fn($x)`）正确执行并返回结果
- 支持 10000+ 深度递归不触发栈溢出
- 新增 4 个深度递归测试用例（5000/10000 层、阶乘、Collatz）

#### static 变量
- 函数级持久化变量，调用之间保持其值
- 语法：`static $var = initialValue;`
- 初始值仅在首次调用时设置
- 支持无初始值声明（默认为 null）
- 每个函数的静态变量独立命名空间
- 新增 5 个测试用例 + 示例文件

### 改进
- `opReturn` 增加尾调用返回传播，正确处理尾调用链
- 编译器隐式 return 检查跳过 `TAIL_CALL` 后的代码
- `TestStressStackOverflow` 更新为非尾递归函数

---

## v0.9.3 (2026-04-02)

### 修复

#### match/case 多行体支持
- 修复 `case` 分支不支持多行语句的问题
- 解析器现在支持 `:` 后的缩进语句块，直到下一个 `case` 或 `}`

#### BigInt/BigDecimal 常量折叠
- 修复 `tryEvalConstant` 未检查 token 类型导致 BigDecimal 被错误解析为 float
- 常量折叠（`tryFoldAdd/Sub/Mul/Div`）增加数值类型检查，避免非数值类型被错误折叠为 0
- 修复 `0.1d + 0.2d == 0.3d` 返回 false 的精度问题

#### include 嵌套 bug
- 修复嵌套 include 时函数索引错乱的问题
- 根因：每个 include 文件独立编译，`globalNames` 映射不一致
- 修复：编译期预编译 include 文件，合并函数定义和全局变量名到父编译器

#### 特殊函数优先级
- 修复 `println (a) * b` 被错误解析为 `(println(a)) * b` 的问题
- 现在正确解析为 `println((a) * b)`
- 同样修复了 `puts`、`pp` 等特殊函数

### 新增

#### 字符串插值格式化
- 支持 `#{$value:.2f}`、`#{$num:05d}` 等格式化语法
- 新增 `FormatExpr` AST 节点、`OP_FORMAT` 字节码指令

#### BigInt/BigDecimal 字面量后缀
- Lexer 支持 `n` 后缀显式声明 BigInt（如 `123n`）
- Lexer 支持 `d` 后缀显式声明 BigDecimal（如 `0.1d`）

### 改进
- `Engine.Compile()` 修复：之前返回空 VM，现在正确调用 `CompileStringWithName`
- `Engine.CompileFile()` 实现：之前是返回 `ErrCompileFailed` 的 stub
- 移除 `pkg/stdlib/fileio.go` 中的 stub 注释

---

## v0.9.0 (2026-04-02)

### 新增

#### 任务系统 `jpl task`
- 在 `jpl.json` 中定义项目任务
- 支持简单格式 (`"name": "cmd"`) 和复杂格式 (`"name": {"cmd": "...", "deps": [...]}`)
- 自动拓扑排序执行依赖任务
- 循环依赖检测和依赖去重
- CLI：`jpl task <name>`, `--list`, `--dry-run`

#### 并行依赖安装
- `jpl install` 自动并行克隆依赖
- `--jobs/-j` 标志控制并发数（默认 4）
- 两阶段流程：并行克隆 → 串行安装

#### 示例项目
- `examples/package-manager/` — 包管理器使用示例
- `examples/tasks/` — 任务系统示例

### 包管理器完善

#### 版本约束
- 支持 semver 语义化版本：`^1.2.3`, `~1.2.3`, `>=1.0.0` 等
- `jpl add <url>@^1.2.0` 语法
- 自动选择满足约束的最佳版本

#### 新增命令
- `jpl init` — 项目初始化（创建 jpl.json、示例文件）
- `jpl update` — 更新依赖到最新版本
- `jpl outdated` — 检查过时的依赖
- `jpl install --resolve` — 使用完整依赖解析器

#### Resolver 集成
- DFS 传递依赖解析
- 版本冲突检测和警告
- 按拓扑顺序安装

### 文档更新
- PROJECT.md 精简为项目简介
- docs/DESIGN.md 新增 D40-D42 决策记录
- README.md 新增任务系统和并行安装文档
- examples/README.md 新增示例链接

---

## v0.8.0 (2026-03-31)

### 新增

#### 正则表达式
- 字面量语法：`#/pattern/flags#`
- 匹配运算符：`=~`
- match/case 正则模式
- `as $var` 捕获组绑定
- 编译期正则验证

#### 包管理器 Phase A/B/C
- 基于 Git 的依赖管理
- `jpl add/remove/install/list` 命令
- 传递依赖解析 + 循环检测
- 全局缓存 `~/.jpl/packages/`
- 锁文件 `jpl.lock.yaml`

### 改进
- BigInt/BigDecimal 独立类型枚举
- 字符串插值解析修复

---

## v0.7.0 (2026-03-27)

### 新增

#### 代码格式化 `jpl fmt`
- 4 空格缩进
- 注释保留
- 对象键排序
- `--write` / `--check` 模式

#### 静态分析 `jpl lint`
- 未使用变量检测
- 未定义变量检测
- 死代码检测

#### 常量折叠优化
- 编译期嵌套表达式求值

---

## v0.6.0 (2026-03-25)

### 新增

#### 网络编程
- TCP/UDP/Unix Socket 支持
- 事件循环（epoll/kqueue）
- DNS 解析（A/AAAA/CNAME/MX/NS/TXT）
- HTTP 客户端
- TLS/mTLS 支持

#### 进程管理
- exec/spawn/kill/fork/pipe
- 21 个进程相关函数

#### 异步文件 IO
- asyncio 模块
- 流式读取、批量操作

---

## v0.5.0 (2026-03-24)

### 新增

#### 语法增强
- 多行字符串（三引号 `'''` 和 `"""`）
- 字符串插值（Ruby 风格 `#{}`）
- 管道运算符（`|>` 正向、`<|` 反向）
- 范围运算符（`...`）
- match/case 模式匹配

#### 类型系统
- BigInt 原生支持
- BigDecimal 原生支持
- Go 风格类型转换（`int(x)`, `string(x)`）

---

## v0.4.0 (2026-03-23)

### 新增

#### 模块系统
- import/include 语句
- URL 导入支持
- 模块路径解析

#### 加密模块
- Hash（SHA-256/512, MD5）
- HMAC
- AES-GCM 加密
- Hex/Base64 编码

---

## v0.3.0 (2026-03-22)

### 新增

#### 闭包和 Lambda
- 完整词法作用域
- Upvalue 捕获
- Lambda/箭头函数语法

#### 异常处理
- try/catch/throw
- 错误码和条件捕获

#### 垃圾回收
- 引用计数
- 循环引用检测

---

## v0.2.0 (2026-03-21)

### 新增

#### 字节码编译器
- Pratt Parser
- 42 条操作码
- 寄存器分配

#### 虚拟机
- fetch-decode-execute 循环
- ABC/ABx/AsBx 指令格式

#### REPL
- 交互式解释器
- 自动补全
- 历史记录

---

## v0.1.0 (2026-03-20)

### 新增

#### 核心基础
- Token 定义
- 词法分析器
- 基本数据类型（null/bool/int/float/string/array/object）
- CLI 工具（run/check/eval）
