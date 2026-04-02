# JPL - Jx9-like Programming Language

> 高性能嵌入式脚本语言引擎，专为 Go 应用程序设计

## 一句话介绍

JPL 是一门轻量级动态脚本语言，提供完整的包管理器、任务系统和 CLI 工具链，可嵌入 Go 程序或独立使用。

## 核心特性

| 分类 | 特性 |
|------|------|
| **语言** | 动态类型、闭包、Lambda、异常处理、match/case 模式匹配 |
| **数据** | null/bool/int/float/string/array/object/BigInt/BigDecimal |
| **语法** | 字符串插值 `#{}`、多行字符串 `'''`、管道运算符 `\|>`、正则 `#/pat/#` |
| **网络** | TCP/UDP/Unix Socket、HTTP 客户端、事件循环、DNS |
| **IO** | 文件读写、异步 IO、二进制处理（pack/unpack/Buffer） |
| **安全** | TLS/mTLS、AES-GCM、SHA-256/512、HMAC |
| **工具** | REPL、代码格式化、静态分析、包管理器、任务系统 |

## 快速体验

```bash
# 安装
go install github.com/gnuos/jpl/cmd/jpl@latest

# Hello World
echo '$name = "JPL"; println "Hello, #{$name}!"' > hello.jpl
jpl run hello.jpl

# 交互式 REPL
jpl repl

# 初始化项目
jpl init my-project && cd my-project

# 运行任务
jpl task start
```

## 项目结构

```
jpl/
├── engine/       # VM、编译器、值类型系统
├── lexer/        # 词法分析器
├── parser/       # Pratt Parser + AST
├── token/        # Token 定义
├── gc/           # 垃圾回收（引用计数 + 循环检测）
├── pkg/pm/       # 包管理器
├── pkg/task/     # 任务系统
├── pkg/lint/     # 静态分析
├── pkg/format/   # 代码格式化
├── pkg/stdlib/   # 标准库（~260 个内置函数）
├── cmd/jpl/      # CLI 入口
├── examples/     # 示例代码
└── docs/         # 设计文档
```

## CLI 命令

| 命令 | 说明 |
|------|------|
| `jpl run <file>` | 执行脚本 |
| `jpl repl` | 交互式解释器 |
| `jpl fmt <file>` | 格式化代码 |
| `jpl lint <file>` | 静态分析 |
| `jpl init` | 初始化项目 |
| `jpl add <source>` | 添加依赖 |
| `jpl install` | 安装依赖（支持并行） |
| `jpl task <name>` | 运行任务 |
| `jpl version` | 版本信息 |

## 作为 Go 库使用

```go
eng := jpl.NewEngine()
defer eng.Close()

eng.Set("name", jpl.NewString("World"))
vm, _ := eng.Compile(`println "Hello, #{$name}!"`)
vm.Execute()
```

## 文档

- [README.md](README.md) — 完整使用文档
- [docs/DESIGN.md](docs/DESIGN.md) — 设计决策记录
- [docs/PACKAGE_MANAGER.md](docs/PACKAGE_MANAGER.md) — 包管理器设计
- [examples/](examples/) — 示例代码

## 状态

**版本**：0.9.0
**状态**：维护模式（核心功能完整）
