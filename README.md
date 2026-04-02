# JPL - Jx9 Go Language

JPL（Jx9-like Programming Language）是一个**高性能、嵌入式脚本语言引擎**，专为 Go 应用程序设计。它提供了完整的脚本执行能力，支持函数式编程、闭包、异常处理和模块化编程。

**本项目全程由 MiMo-V2-Pro 模型、MiniMax-M2.5 模型、Kimi-K2.5 模型生成代码与文档，当前文档经过了手动整理**

[项目介绍](./PROJECT.md)

下面的资料是设计过程和来时的路：

- [设计文档](./docs/DESIGN.md)
- [计划文档](./docs/PLAN.md)
- [开发进度](./docs/PROGRESS.md)
- [AI提示词](./docs/jx9_golang_prompts.md)
- [包管理设计](./docs/PACKAGE_MANAGER.md)
- [约束提示词](./CLAUDE.md)

---

## 🌟 核心特性

### 语言特性

- 📦 **丰富的数据类型**：null、bool、int64、float64、string、array、object、bigint、bigdecimal、function、error
- 🔧 **动态类型系统**：自动类型转换和推断，支持 Go 风格类型转换 `int(x)`, `string(x)`
- 🎯 **字节码编译**：AST → 字节码 → 基于寄存器的虚拟机执行
- 🔄 **闭包支持**：完整的词法作用域和 Upvalue 捕获机制
- ⚡ **Lambda 箭头函数**：简洁的函数定义语法 `($x) -> $x * 2`
- 🛡️ **异常处理**：try/catch/throw 语句，支持错误码和条件捕获
- 📚 **模块化编程**：import/include 语句，支持文件和 URL 导入
- 🎨 **函数重载**：按参数数量自动匹配
- 🚀 **特例函数语法**：内置函数支持无括号调用（如 `print hello"`）
- 🛑 **进程控制**：exit/die 函数 + 完整进程 API（exec, spawn, kill, fork, pipe 等）
- 📝 **多行字符串**：Python 风格三引号 `'''` 和 `'''`，保留换行和缩进
- 💎 **字符串插值**：Ruby 风格 `#{}` 语法，支持变量、表达式、对象访问、数组索引
- 🔒 **TLS/SSL**：完整的加密通信能力，支持 HTTPS、自签名证书，双向认证（mTLS）
- 🌐 **HTTP Client**：高级 HTTP 客户端，支持 JSON/Form/认证/超时/重定向
- 🔍 **正则表达式**：字面量语法 `#/pattern/flags#`，`=~` 匹配运算符，match/case 正则模式，`as $var` 捕获绑定
- 🔐 **加密模块**：Hash（SHA-256/512）、HMAC、AES-GCM 加密、Hex/Base64 编码
- 🎪 **@member 闭包成员访问**：类似 Ruby 语法，在闭包内访问对象成员
- 📍 **运行时错误定位**：运行时错误显示源码行号，便于调试
- 🔗 **管道运算符**：`|>` 正向管道和 `<|` 反向管道，支持函数式数据流处理

### 性能优化

- 💾 **小整数缓存**：[-256, 1024] 范围内的整数使用预分配单例
- 📝 **字符串内部化**：短字符串缓存，O(1) 相等性比较
- 🎯 **全局变量索引**：编译期分配数组索引，运行时 O(1) 访问
- ♻️ **垃圾回收**：引用计数 + 循环检测，可选 GC 支持
- ⚙️ **int 快速路径**：算术和比较操作优先使用整数运算

### 开发体验

- 🖥️ **REPL 界面**：交互式命令行，支持自动补全、历史记录、调试指令
- 🔍 **调试工具**：字节码反编译、执行追踪、VM 状态转储
- 🔧 **CLI 工具**：run、check、eval、repl、fmt、lint、init、add、remove、install、list、task 子命令
- 📦 **包管理器**：基于 Git 的依赖管理，支持传递依赖解析、全局缓存和并行安装
- ⚡ **任务系统**：在 jpl.json 中定义任务，支持依赖关系和循环检测
- 📖 **完整文档**：详细的 API 注释和示例代码

---

## 📦 安装

### 作为 CLI 工具使用

```bash
# 直接安装最新版本
go install github.com/gnuos/jpl/cmd/jpl@latest

# 验证安装
jpl version
```

### 作为 Go 库使用

```bash
# 添加到项目依赖
go get github.com/gnuos/jpl
```

---

## 🚀 快速开始

### 1. Hello World（脚本）

创建 `hello.jpl`：

```jpl
$name = "World"
puts "Hello, #{$name}!"
```

运行：

```bash
jpl run hello.jpl
# 输出: Hello, World!
```

### 2. 交互式 REPL

```bash
$ jpl repl
JPL REPL - 输入 :help 查看指令，Ctrl+D 退出
> $x = 10
> y = 20
> $x + y
30
> fn greet(name) { return "Hello, " .. name }
> greet("JPL")
"Hello, JPL"
> exit
```

### 3. 初始化项目

```bash
# 创建新项目
jpl init my-project
cd my-project

# 查看项目结构
ls -la
# jpl.json    - 项目清单
# main.jpl    - 示例脚本
# jpl_modules/ - 依赖目录

# 运行示例
jpl run main.jpl
# 输出: Hello from my-project!
```

生成的 `jpl.json`：

```json
{
    "name": "my-project",
    "version": "0.1.0",
    "dependencies": {}
}
```

### 4. 嵌入 Go 程序

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/gnuos/jpl"
)

func main() {
    // 创建引擎实例
    engine := jpl.NewEngine()
    defer engine.Close()
    
    // 编译并执行脚本
    vm, err := engine.Compile(`
        $items = [10, 20, 30, 40, 50]
        $sum = reduce($items, ($acc, $x) -> $acc + $x, 0)
        return $sum
    `)
    if err != nil {
        log.Fatal(err)
    }
    defer vm.Close()
    
    // 执行
    if err := vm.Execute(); err != nil {
        log.Fatal(err)
    }
    
    // 获取结果
    result := vm.GetResult()
    fmt.Printf("Sum: %d\n", result.Int()) // 输出: Sum: 150
}
```

---

## 📖 文档手册

- [API文档](./docs/Learning/API.md)
- [语法教程](./docs/Learning/Syntax.md)
- [内置函数](./docs/Learning/Builtin.md)
- [标准库签名](./docs/Learning/Stdlib.md)
- [标准库全集](./docs/Stdlib/README.md)
- [事件循环框架](./docs/Learning/EventLoop.md)

---

## 🔧 CLI 工具

### 子命令概览

```bash
jpl <command> [flags] [args]

Commands:
  run      运行 JPL 脚本文件
  check    检查语法（不执行）
  eval     执行代码片段
  repl     启动交互式解释器
  fmt      格式化代码
  lint     静态分析
  init     初始化项目
  add      添加依赖
  remove   移除依赖
  install  安装全部依赖（支持并行克隆）
  update   更新依赖
  outdated 检查过时的依赖
  list     列出依赖
  task     运行项目任务
  version  显示版本信息

Flags:
  -v, --verbose  详细输出
  -d, --debug    调试模式
  -h, --help     帮助信息
```

### run - 运行脚本

```bash
# 运行脚本文件
jpl run script.jpl

# 传递参数
jpl run script.jpl arg1 arg2 arg3
# 脚本内通过 ARGV[1], ARGV[2], ARGV[3] 访问

# 启用调试输出
jpl run --debug script.jpl
```

### check - 语法检查

```bash
# 检查单个文件
jpl check script.jpl

# 检查多个文件
jpl check file1.jpl file2.jpl file3.jpl

# 详细输出
jpl check --verbose script.jpl
```

### eval - 执行代码片段

```bash
# 简单表达式
jpl eval 'print 2 + 2'

# 多行代码（使用分号）
jpl eval '$x = 10; $y = 20; print $x + $y'

# 复杂代码（使用引号）
jpl eval 'for ($i = 0; $i < 5; $i = $i + 1) { print $i }'
```

### repl - 交互式解释器

```bash
# 启动 REPL
jpl repl

# REPL 调试指令
#   :help         - 显示帮助
#   :debug on/off - 切换调试模式（显示执行步骤）
#   :globals      - 显示全局变量
#   :locals       - 显示局部变量
#   :vars         - 显示所有变量
#   :funcs        - 显示所有内置函数
#   :consts       - 显示预设常量
#   :doc <name>   - 查看函数签名
#   :quit         - 退出 REPL

# REPL 快捷键
#   Tab         - 自动补全
#   Up/Down     - 历史导航
#   Ctrl+C      - 中断执行
#   Ctrl+D      - 退出
```

### fmt - 代码格式化

```bash
# 输出格式化后的代码到 stdout
jpl fmt script.jpl

# 原地格式化文件
jpl fmt --write script.jpl

# 检查文件是否已格式化（退出码 0=已格式化 1=需格式化）
jpl fmt --check script.jpl

# 批量格式化
jpl fmt --write *.jpl
```

格式化规则：
- 4 空格缩进
- 保留注释（行尾注释保持同行）
- 对象字面量键按字母排序
- 幂等性保证（重复 fmt 输出一致）

### lint - 静态分析

```bash
# 检查单个文件
jpl lint script.jpl

# 检查多个文件
jpl lint *.jpl
```

检测规则：
- `unused-var` — 声明但未使用的变量（warning）
- `undefined-var` — 使用未声明的变量（error）
- `dead-code` — return/break/continue 后的不可达代码（warning）

### 包管理器

JPL 提供基于 Git 的包管理器，支持项目初始化、依赖安装、卸载和传递依赖解析。

#### 初始化项目

```bash
# 在当前目录初始化
jpl init

# 创建新目录并初始化
jpl init my-project

# 指定项目名称和描述
jpl init --name my-app --desc "My awesome app"

# 不创建示例文件
jpl init --no-example
```

生成的项目结构：

```
my-project/
├── jpl.json        # 项目清单
├── main.jpl        # 示例脚本
└── jpl_modules/    # 依赖目录（空）
```

生成的 `jpl.json` 示例：

```json
{
    "name": "my-project",
    "version": "0.1.0",
    "description": "",
    "dependencies": {}
}
```

#### 添加依赖

```bash
# 从 Git 仓库添加依赖
jpl add https://github.com/user/jpl-utils.git

# 指定版本（tag）
jpl add https://github.com/user/jpl-utils.git@v1.0.0

# 指定分支
jpl add https://github.com/user/jpl-utils.git#main

# 从本地路径添加
jpl add ../my-lib

# 自定义导入名称
jpl add https://github.com/user/jpl-utils.git --name utils

# 移除依赖
jpl remove utils

# 安装全部依赖（根据 jpl.json）
jpl install

# 列出已安装的依赖
jpl list
```

#### 更新依赖

```bash
# 更新所有依赖
jpl update

# 更新指定依赖
jpl update utils

# 检查过时的依赖
jpl outdated
```

#### 版本约束

支持语义化版本约束，使用 `@` 后缀指定：

```bash
# 兼容版本（>=1.2.3, <2.0.0）
jpl add https://github.com/user/lib.git@^1.2.3

# 补丁版本（>=1.2.3, <1.3.0）
jpl add https://github.com/user/lib.git@~1.2.3

# 大于等于
jpl add https://github.com/user/lib.git@">=1.0.0"

# 精确版本
jpl add https://github.com/user/lib.git@1.2.3
```

约束语法：

| 语法 | 含义 | 示例匹配 |
|------|------|----------|
| `^1.2.3` | 兼容版本 | `1.2.3` ~ `1.x.x` |
| `~1.2.3` | 补丁版本 | `1.2.3` ~ `1.2.x` |
| `>=1.2.3` | 大于等于 | `1.2.3`, `1.3.0`, `2.0.0` |
| `>1.2.3` | 大于 | `1.2.4`, `1.3.0`, `2.0.0` |
| `<1.2.3` | 小于 | `1.2.2`, `1.1.0`, `0.9.0` |
| `<=1.2.3` | 小于等于 | `1.2.3`, `1.2.2`, `1.1.0` |
| `1.2.3` | 精确匹配 | `1.2.3` |
| `*` | 任意版本 | 所有版本 |

#### 项目结构

添加依赖后，项目目录结构如下：

```
my-project/
├── jpl.json              # 项目清单（自动生成）
├── jpl.lock.yaml         # 锁文件（自动生成）
├── jpl_modules/          # 依赖安装目录
│   ├── utils/
│   │   ├── index.jpl     # 包入口文件
│   │   └── lib/
│   └── http/
│       └── index.jpl
└── src/
    └── main.jpl          # 你的代码
```

#### 清单文件 (jpl.json)

```json
{
    "name": "my-project",
    "version": "0.1.0",
    "dependencies": {
        "utils": "https://github.com/user/jpl-utils.git",
        "http": "https://github.com/user/jpl-http.git@v1.2.0"
    }
}
```

#### 使用依赖

安装后直接使用裸名称导入：

```jpl
// 导入已安装的包
import "utils"
import "http"

// 使用包中的函数
$result = utils.format("Hello")
$response = http.get("https://api.example.com")
```

#### 传递依赖

如果包 A 依赖包 B，安装 A 时会自动安装 B：

```bash
# 假设 jpl-utils 依赖 jpl-core
$ jpl add https://github.com/user/jpl-utils.git
添加依赖: utils (https://github.com/user/jpl-utils.git @ abc123d)
安装到: jpl_modules/utils/
  (传递依赖) core (https://github.com/user/jpl-core.git @ def456a)
```

#### 全局缓存

已安装的包会缓存到 `~/.jpl/packages/`，避免重复克隆：

```bash
# 禁用缓存（强制重新克隆）
jpl add https://github.com/user/lib.git --no-cache
jpl install --no-cache
```

#### 锁文件 (jpl.lock.yaml)

锁文件记录每个依赖的确切 commit hash，确保不同机器安装相同版本：

```yaml
version: 1
generated: "2026-04-02T12:00:00Z"
packages:
  utils:
    source: "https://github.com/user/jpl-utils.git"
    resolved: "https://github.com/user/jpl-utils.git"
    version: "v1.0.0"
    commit: "abc123def456"
    hash: "sha256:e3b0c44298fc..."
    dependencies:
      - core
  core:
    source: "https://github.com/user/jpl-core.git"
    commit: "def456abc789"
    ...
```

#### 并行安装

依赖安装时自动并行克隆，提升安装速度：

```bash
# 默认并发数 4
jpl install

# 指定并发数 8
jpl install -j 8

# 结合 resolve 模式
jpl install --resolve -j 4
```

### 任务系统

JPL 支持在 `jpl.json` 中定义项目任务，类似 npm scripts 和 deno task。

#### 定义任务

在 `jpl.json` 中添加 `tasks` 字段：

```json
{
    "name": "my-project",
    "version": "0.1.0",
    "dependencies": {},
    "tasks": {
        "clean": "rm -rf build",
        "lint": "jpl run scripts/lint.jpl",
        "build": {
            "cmd": "jpl run scripts/build.jpl",
            "deps": ["clean", "lint"]
        },
        "test": {
            "cmd": "jpl run tests/main.jpl",
            "deps": ["build"]
        },
        "dev": "jpl run scripts/dev.jpl --watch"
    }
}
```

#### 任务格式

支持两种格式：

1. **简单格式**：`"name": "command"` — 字符串形式
2. **复杂格式**：`"name": {"cmd": "...", "deps": [...]}` — 带依赖的对象形式

命令类型：
- **JPL 脚本**：`jpl run script.jpl` 或 `script.jpl` — 使用 jpl 执行
- **Shell 命令**：`rm -rf build`、`echo hello` 等 — 使用 sh 执行

#### 运行任务

```bash
# 运行单个任务（自动执行其依赖）
jpl task test

# 列出所有可用任务
jpl task --list

# 显示执行计划但不执行
jpl task test --dry-run
```

#### 执行顺序

任务按依赖关系拓扑排序执行，自动处理：
- **循环依赖检测**：检测并报告循环依赖
- **依赖去重**：菱形依赖场景下每个任务只执行一次

示例执行顺序：

```bash
$ jpl task test
Running "test" (with 2 dependencies):
  → clean (rm -rf build)
  → lint  (jpl run scripts/lint.jpl)
  → build (jpl run scripts/build.jpl)
  → test  (jpl run tests/main.jpl)
```

---

## 🔌 API 使用示范

### 引擎 API

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/gnuos/jpl"
)

func main() {
    // 创建引擎
    eng := jpl.NewEngine()
    defer eng.Close()
    
    // 注册变量
    eng.Set("version", jpl.NewString("1.0.0"))
    eng.Set("debug", jpl.NewBool(true))
    
    // 注册函数
    eng.RegisterFunc("add", func(ctx *jpl.Context, args []jpl.Value) (jpl.Value, error) {
        if len(args) != 2 {
            return nil, fmt.Errorf("add() requires 2 arguments")
        }
        a := args[0].Int()
        b := args[1].Int()
        return jpl.NewInt(a + b), nil
    })
    
    // 编译脚本
    vm, err := eng.Compile(`
        $result = add(10, 20)
        if ($debug) {
            print "Debug: result = " + $result
        }
        return $result
    `)
    if err != nil {
        log.Fatal(err)
    }
    defer vm.Close()
    
    // 执行
    if err := vm.Execute(); err != nil {
        log.Fatal(err)
    }
    
    // 获取结果
    fmt.Printf("Result: %d\n", vm.GetResult().Int())
}
```

### VM 反射 API

```go
// 列出所有函数
funcs := vm.ListFunctions()
fmt.Printf("定义了 %d 个函数\n", len(funcs))

// 获取函数信息
if infos, ok := vm.GetFunctionInfo("add"); ok {
    for _, info := range infos {
        fmt.Printf("函数 %s 有 %d 个参数: %v\n", 
            info.Name, info.ParamCount, info.ParamNames)
    }
}

// 动态调用函数
result, err := vm.CallByName("add", jpl.NewInt(5), jpl.NewInt(10))
if err != nil {
    log.Fatal(err)
}
fmt.Printf("add(5, 10) = %d\n", result.Int())

// 获取全局变量
if value, ok := vm.GetGlobal("result"); ok {
    fmt.Printf("result = %v\n", value.Int())
}
```

### 调试与追踪

```go
// 查看字节码
asm := vm.Disassemble()
fmt.Println("=== Bytecode ===")
fmt.Println(asm)

// 设置调试追踪（可选）
vm.SetTraceConfig(&engine.TraceConfig{
    Enabled:     true,
    ShowState:   true,
    ShowGlobals: true,
})

// 执行并查看追踪输出
vm.Execute()
```

### 错误处理

```go
// 执行并捕获错误
if err := vm.Execute(); err != nil {
    if err == engine.ErrVMClosed {
        log.Println("VM 已关闭")
    } else if engine.IsRuntimeError(vm.GetResult()) {
        log.Printf("运行时错误: %v\n", vm.GetResult().Stringify())
    } else {
        log.Printf("执行错误: %v\n", err)
    }
}
```

## 📊 性能

### 优化建议

1. **使用小整数**：[-256, 1024] 范围内的整数无内存分配
2. **缓存短字符串**：重复字符串会自动共享
3. **避免全局变量修改**：使用局部变量提高性能
4. **使用内置函数**：内置函数比脚本函数快 10-100 倍
5. **启用 GC**：对于长时间运行的应用，启用 GC 防止内存泄漏

## 🏗️ 项目结构

```
jpl/
├── cmd/jpl/          # CLI 工具
│   ├── main.go       # 入口
│   ├── root.go       # 根命令
│   ├── run.go        # run 子命令
│   ├── check.go      # check 子命令
│   ├── eval.go       # eval 子命令
│   ├── fmt.go        # fmt 子命令
│   ├── repl.go       # REPL 实现
│   ├── repl_test.go  # REPL 测试
│   └── pm.go         # 包管理器命令（add/remove/install/list）
│
├── engine/           # 核心引擎
│   ├── engine.go     # 引擎 API
│   ├── vm.go         # 虚拟机实现
│   ├── compiler.go   # 字节码编译器
│   ├── value.go      # 值类型系统
│   ├── bytecode.go   # 字节码定义
│   ├── errors.go     # 错误类型
│   ├── lockfile.go   # 锁文件（含包管理器扩展）
│   ├── module_loader.go  # 模块加载
│   └── ...           # 测试文件
│
├── token/            # 词法单元
├── lexer/            # 词法分析器
├── parser/           # 语法分析器（Pratt Parser）
├── format/           # 代码格式化器
├── lint/             # 静态分析器
├── pkg/stdlib/       # 标准库内置函数
│   ├── builtin.go    # 核心函数
│   ├── string.go     # 字符串函数
│   ├── array.go      # 数组函数
│   ├── math.go       # 数学函数
│   ├── fileio.go     # 文件 I/O
│   ├── hash.go       # Hash 函数
│   ├── functional.go # 函数式编程
│   ├── gc.go         # GC 函数
│   └── ...
│
├── pkg/              # 扩展包
│   └── pm/           # 包管理器
│       ├── manifest.go       # 清单文件读写
│       ├── git.go            # Git 操作
│       ├── resolver.go       # 依赖解析
│       ├── cache.go          # 全局缓存
│       ├── manifest_test.go  # 清单测试
│       └── resolver_test.go  # 解析测试
│
├── gc/               # 垃圾回收器
├── api.go            # 顶层 API 导出
├── version.go        # 版本信息
└── README.md         # 本文件
```

## 📝 设计决策

### 为什么选择字节码虚拟机？

- **性能**：比 AST 解释器快 5-10 倍
- **可移植性**：字节码与平台无关
- **可调试性**：易于反编译和追踪
- **扩展性**：便于添加新指令

### 为什么基于寄存器而非栈？

- **性能**：减少指令数量，更高效
- **简洁**：生成的字节码更紧凑
- **调试**：寄存器状态比栈更直观

### 为什么小整数缓存？

- **内存**：避免频繁分配小对象
- **缓存友好性**：常用值在 CPU 缓存中
- **实践中常见**：大多数脚本整数在此范围内

## 🤝 贡献

欢迎贡献！请阅读 [CONTRIBUTING.md](CONTRIBUTING.md) 了解贡献指南。

### 提交 Issue

- 使用清晰的标题描述问题
- 提供最小可复现示例
- 说明期望行为和实际行为
- 注明操作系统和 Go 版本

### 提交 PR

1. Fork 仓库
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

## 📄 许可证

JPL 使用 [MIT 许可证](LICENSE)。

---

<p align="center">
  用 ❤️ 和 Go 编写
</p>
