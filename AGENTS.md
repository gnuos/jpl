# AGENTS.md

此文件为 AI 编码助手提供 JPL 项目的工作指南。

## 项目概述

JPL（Jx9-like Programming Language）是一个高性能嵌入式脚本语言引擎，专为 Go 应用程序设计。它提供完整的脚本执行能力，包括编译器、虚拟机、包管理器和 CLI 工具链。

## 常用命令

### 构建
```bash
# 构建二进制文件
make build

# 构建发布版本
make release

# 安装到 GOPATH/bin
make install
```

### 测试
```bash
# 运行所有测试
make test
go test ./...

# 运行测试并显示覆盖率
make test-cover
go test -cover ./...

# 运行基准测试
make bench
go test -bench=. ./...

# 运行单个包的测试
go test ./engine

# 运行单个测试函数
go test -run TestCompiler ./engine

# 运行指定文件的测试
go test ./engine/engine_test.go ./engine/vm_test.go
```

### 代码质量
```bash
# 代码格式化
make fmt
go fmt ./...

# 代码检查
make lint

# 依赖整理
make tidy
go mod tidy
```

### 开发
```bash
# 运行 JPL 脚本
make run FILE=script.jpl
go run ./cmd/jpl run script.jpl

# 检查脚本语法
make check FILE=script.jpl
go run ./cmd/jpl check script.jpl

# 启动 REPL
make repl
go run ./cmd/jpl repl
```

## 代码架构

### 核心组件

**engine/** - 核心引擎
- `engine.go`: 引擎入口，管理全局环境、变量和函数注册
- `vm.go`: 寄存器虚拟机，执行字节码
- `compiler.go`: 字节码编译器，将 AST 转换为字节码
- `value.go`: 统一值类型系统，支持所有数据类型
- `bytecode.go`: 字节码指令定义
- `context.go`: 执行上下文
- `module_loader.go`: 模块加载器，支持 import/include 语句
- `lockfile.go`: 包管理器锁文件（jpl.lock.yaml）

**lexer/** - 词法分析器
- `lexer.go`: 将源代码转换为 token 流
- 支持所有 JPL 语法特性（字符串插值、多行字符串、正则表达式等）

**parser/** - 语法分析器
- `parser.go`: Pratt Parser，构建 AST
- `ast.go`: 抽象语法树节点定义

**token/** - Token 定义
- 所有 token 类型和常量定义

**gc/** - 垃圾回收器
- 引用计数 + 循环检测
- 集成到 Value 类型系统

**pkg/stdlib/** - 标准库
- 包含 300 多个内置函数
- 按功能模块组织（array.go, string.go, math.go, crypto.go 等）
- 每个模块都有对应的测试文件

**pkg/pm/** - 包管理器
- `manifest.go`: jpl.json 清单文件管理
- `git.go`: Git 操作（克隆、checkout 等）
- `resolver.go`: 依赖解析（传递依赖、循环检测）
- `cache.go`: 全局包缓存（~/.jpl/packages/）
- 支持版本约束（^, ~, >=, >, <, <=）

**pkg/format/** - 代码格式化器
- 将 JPL 代码格式化为标准格式（4 空格缩进）

**pkg/lint/** - 静态分析器
- 检测未使用变量、未定义变量、死代码

**pkg/task/** - 任务系统
- `task.go`: 任务定义和执行计划
- 支持 jpl.json tasks 字段
- 依赖关系解析和拓扑排序

**cmd/jpl/** - CLI 工具
- `main.go`: 入口
- `root.go`: 根命令
- `run.go`: run 子命令
- `check.go`: check 子命令
- `eval.go`: eval 子命令
- `fmt.go`: fmt 子命令
- `lint.go`: lint 子命令
- `repl.go`: REPL 实现
- `pm.go`: 包管理器命令（add/remove/install/list）
- `task.go`: 任务系统命令

### 执行流程

1. **词法分析**: lexer 将源代码转换为 token 流
2. **语法分析**: parser 使用 Pratt Parser 构建 AST
3. **编译**: compiler 遍历 AST 生成字节码
4. **执行**: vm 执行字节码，管理寄存器和栈

### 关键设计决策

**寄存器虚拟机**
- 基于寄存器而非栈，减少指令数量
- 寄存器分配使用线性扫描算法
- 函数调用时保存/恢复寄存器窗口

**闭包支持**
- 使用 upvalue 捕获外部变量
- 闭包对象包含编译函数和捕获的 upvalue 列表
- 支持完整的词法作用域

**值类型系统**
- 统一的 Value 接口，所有类型都实现此接口
- 支持类型转换和自动类型推断
- 小整数缓存 [-256, 1024] 提升性能

**模块系统**
- import/include 语句加载模块
- 模块缓存机制避免重复加载
- 支持文件系统和 URL 导入

## 重要约定

### 变量命名规则
- `$name`: 全局/局部变量（带 $ 前缀）
- `name`: 全局/局部变量（和 $name 是不同的标识符）
- `_private`: 私有变量（仅当前作用域可访问，不能被 global 声明）
- `_`: 丢弃值占位符
- `$_`: 执行结果保留变量

### 字符串操作
- 使用 `..` 作为字符串连接运算符
- 支持字符串插值 `#{expression}`
- 支持多行字符串 `'''` 和 `'''`

### 函数定义
- `function` 和 `fn` 都可用
- 支持箭头函数 `($x) -> $x * 2`
- 支持函数重载（按参数数量）

### Null 合并运算符
- `??` 运算符：左侧为 null 时返回右侧值
- 支持链式使用：`a ?? b ?? c`
- 优先级高于 `||` 和 `&&`

### 异常处理
- try/catch/throw 语句
- 支持多个 catch 分支
- 支持 when 条件捕获：`catch ($e when $e.code == 404)`
- 支持跨函数异常传播

### 模式匹配
- match/case 语句
- 支持值匹配、范围匹配、正则匹配
- 支持 OR 模式和解构绑定

### 依赖版本管理
- 始终使用最新稳定版本
- Go: 使用 `go get -u package-name`
- 不使用固定版本号，除非用户明确要求

## Go 代码风格

### 导入规范
- 标准库在前，空行分隔，然后是外部包，最后是内部包
- 按字母顺序排列

```go
import (
    "fmt"
    "strings"

    "github.com/gnuos/jpl/parser"
    "github.com/gnuos/jpl/token"
)
```

### 命名规范
- 导出类型/函数：`PascalCase`（如 `CompileFunction`, `Value`）
- 未导出：`camelCase`（如 `compileExpr`, `allocReg`）
- 接口：名词或 `-er` 后缀（如 `Value`, `Iterator`）
- 常量：`UPPER_SNAKE`（如 `OP_ADD`, `OP_JMP`）
- 测试函数：`TestXxx`，优先使用表驱动测试

### 错误处理
- 所有公开 API 必须返回 error 并检查
- 编译错误使用 `CompileError`（包含行号/列号）
- 运行时错误使用 `RuntimeError`
- compiler 中使用 panic + recover 传播错误

### 注释规范
- 使用 `//` 注释（中文注释可接受）
- 使用 `// ============================================================================
// 分区` 分隔主要代码段
- 所有导出类型和函数必须有 godoc 注释

### 线程安全
- Engine 是线程安全的（使用 sync.RWMutex）
- Compiler 不是线程安全的，同一时间只能编译一个程序

## 开发注意事项

1. **线程安全**: Engine 的所有方法都是线程安全的，内部使用 sync.RWMutex
2. **Compiler 不是线程安全的**: 同一时间只能编译一个程序
3. **内存管理**: Value 类型使用引用计数，长时间运行的应用建议启用 GC
4. **性能优化**: 优先使用内置函数（比脚本函数快 10-100 倍）
5. **错误处理**: 所有 API 都返回 error，必须检查错误

## 测试策略

- 每个模块都有对应的 `*_test.go` 文件
- 测试覆盖率通过 `make test-cover` 查看
- 集成测试位于 `integration_test.go` 文件
- 压力测试位于 `stress_test.go` 文件

## 常见修改任务指南

### 添加新操作码
1. `engine/bytecode.go`: 添加操作码常量和指令格式
2. `engine/vm.go`: 添加执行逻辑（switch case）
3. `engine/compiler.go`: 添加编译生成逻辑

### 添加内置函数
1. `pkg/stdlib/`: 在对应模块文件中添加函数
2. 注册到引擎的函数表中

### 添加新语法
1. `token/token.go`: 添加 token 类型
2. `lexer/lexer.go`: 添加词法分析
3. `parser/parser.go`: 添加语法解析
4. `parser/ast.go`: 添加 AST 节点
5. `engine/compiler.go`: 添加编译逻辑

### 修改解析器
- `parser/parser.go` 使用 Pratt 解析，优先级表定义在文件开头

## 文档资源

- [README.md](README.md) - 完整使用文档
- [CLAUDE.md](CLAUDE.md) - AI 协作开发规范
- [PROJECT.md](PROJECT.md) - 项目概述
- [docs/DESIGN.md](docs/DESIGN.md) - 设计决策记录
- [docs/PACKAGE_MANAGER.md](docs/PACKAGE_MANAGER.md) - 包管理器设计
- [examples/](examples/) - 示例代码
