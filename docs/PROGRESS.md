# JPL 项目进度跟踪

> 本文档用于记录项目开发进度，每次完成阶段任务后更新。暂停后恢复工作时，从此文档了解当前状态和下一步任务。

---

## 更新记录

| 日期 | 阶段 | 任务 | 说明 |
|------|------|------|------|
| 2026-04-03 | 用户体验 | 更好的错误消息 | 运行时错误显示行号 + 源码上下文（箭头标记 + 前后 2 行），Program 存储源代码，RuntimeError 新增 FormatWithContext |
| 2026-04-03 | 性能优化 | 尾调用优化 (TCO) | 自递归尾调用栈帧复用，支持 10000+ 深度递归不溢出；编译器 emit OP_TAIL_CALL，VM 自递归检测 + IP 跳转，opReturn 尾调用返回传播 |
| 2026-04-03 | 语言特性 | static 变量 | 函数级持久化变量，调用间保持值；新增 5 个测试 + 示例文件 static-variables.jpl |
| 2026-04-02 | Bug 修复 | 字符串插值解析修复 | parser isValueToken 添加 STRING_FRAG，修复插值后逗号分隔参数解析失败 |
| 2026-04-02 | 类型系统 | BigInt/BigDecimal 独立类型 | 添加 TypeBigInt 枚举，编译器检查 token 类型，添加 is_bigint/is_bigdecimal 函数 |
| 2026-04-02 | 示例修复 | control-flow.jpl 语法修复 | var 关键字改为 $ 前缀，添加未定义变量初始化 |
| 2026-04-02 | Phase A+B | 包管理器实现 | pkg/pm 包（manifest/git/resolver/cache）、cmd/jpl/pm.go（add/remove/install/list）、扩展 lockfile.go Packages 字段，25 个测试 |
| 2026-04-01 | 规划 | 进入维护模式 | LSP/调试器评估为不值得实现（投入产出比低），包管理器设计文档已完成待需求触发，保留已知 bug 列表 |
| 2026-04-01 | 设计 | 包管理器设计文档 | docs/PACKAGE_MANAGER.md：import 解析方案（扁平安装）、jpl.json 清单格式、CLI 命令设计、依赖解析算法、Phase A/B/C 实现计划 |
| 2026-04-01 | Phase 19.2 | jpl lint 静态分析 | 3 条规则：unused-var(未使用变量)、undefined-var(未定义变量)、dead-code(死代码)，lint 包（~350 行 + 10 测试），CLI 子命令 |
| 2026-04-01 | 优化 | 常量折叠 | 编译期常量表达式求值：整数/浮点算术、字符串拼接、比较、布尔逻辑（含短路）、位运算、一元运算、三元表达式、嵌套折叠，8 组测试用例 |
| 2026-04-01 | Phase 8.2 | print/println 流参数 | print/println 首参数支持流类型，print(STDERR, "msg") 输出到指定流，3 个新测试用例 |
| 2026-04-01 | Phase 11.4 | stub 函数实现 | 实现 fseek(文件指针定位)、ftell(获取位置)、rewind(重置指针)、ftruncate(截断文件)、fgetcsv(CSV 解析)，新增 streamValue.Seek/Truncate 方法，8 个测试用例 |
| 2026-04-01 | Phase 22 | jpl fmt 完成 | Token 层新增 COMMENT/BLOCK_COMMENT，Lexer 保留注释，Parser 跳过注释 token，新建 format 包（870 行引擎 + 46 测试），CLI 子命令支持 --write/--check，对象键排序保证幂等性，行尾注释保持同行 |
| 2026-03-31 | 维护 | 已知测试问题修复 | 修复 5 个预存失败测试：TestArrayNames(数量更新)、TestVMArraySetNegativeIndex(parser LBRACKET 排除)、TestVMTraceCustomHook(while 括号)、TestGetDefinedFunctionsContainsBuiltin(if 括号+foreach+赋值语法)、TestGetDefinedConstantsContainsPI(同上) |
| 2026-03-31 | Phase 18 | 正则字面量实现完成 | 12 个文件修改/新增，20 个测试用例全部通过。`#/pattern/flags#` 语法 + `=~` 运算符 + match/case 正则模式 + `as $var` 捕获绑定 + 编译期错误检测 |
| 2026-03-31 | 设计 | 正则字面量设计完成 | D39 设计决策：`#/pattern/flags#` 语法 + `=~` 运算符 + match/case 正则模式 + `as $var` 捕获绑定，详见 docs/DESIGN.md |
| 2026-03-28 | Phase 9.7 | 边界测试完成 | 12 个边界测试：网络超时、错误处理、并发、大数据、资源清理、Unix Socket，全部通过 |
| 2026-03-28 | Phase 9.6 | 代码示例完成 | 5 个网络编程示例：HTTP 服务器、WebSocket、Redis 客户端、聊天室、文件上传服务器 |
| 2026-03-27 | 规划 | 下一步任务规划 | 在 PROGRESS.md 添加 Phase 9.6-9.7 任务清单（补全代码示例、边界测试）；在 PLAN.md 添加方向2（新功能开发）和方向3（工具链完善）的未来计划；创建 file-io.jpl 示例 |
| 2026-03-27 | Phase 9.5 | 集成测试完成 | 7 个集成测试：Unix Echo、TCP 客户端-服务器、Buffer+网络协议、DNS+Connect、UDP Echo、Pack+Send、完整网络栈，验证 binary/ev/net/dns 模块协同工作 |
| 2026-03-27 | Phase 9.4 | DNS 解析完成 | dns 模块（5 个函数）：dns_resolve（所有 IP）、dns_resolve_one（单个 IP）、dns_resolve_v4/v6（IPv4/IPv6 过滤）、dns_get_records（A/AAAA/CNAME/MX/NS/TXT 记录），Go net 包原生支持 |
| 2026-03-27 | Phase 9.3 | 网络层模块完成 | net 模块（17 个函数）：TCP（listen/connect/accept）、Unix Domain（listen/connect/accept）、UDP（bind/sendto/recvfrom）、通用函数（send/recv/close/getsockname/getpeername/set_nonblock/is_unix），统一 API 设计 |
| 2026-03-27 | Phase 9.1 | 二进制处理完成 | pack/unpack 函数（11 种格式字符）、Buffer 对象（24 个操作函数）、字节序切换、游标控制，支持微秒级定时器精度基础 |
| 2026-03-27 | Phase 9.2 | 事件循环核心完成 | ev 模块（23 个函数）：跨平台 epoll(Linux)/kqueue(BSD) 实现、注册表模式、微秒级定时器、POSIX 信号处理（SIGINT/SIGTERM 等）|
| 2026-03-27 | Phase 8 | 流类型系统完成 | 新增TypeStream类型、streamValue实现、8个IO函数（fopen/fread/fgets/fwrite/fclose/feof/fflush/stream_get_meta_data）、is_stream()类型检查、STDIN/STDOUT/STDERR升级为流类型，33个测试 |
| 2026-03-27 | Phase 11 | 内置函数补全完成 | 新增23个函数：md5_file/sha1_file/file_get_contents/file_put_contents/copy/readfile/pathinfo/disk_free_space/disk_total_space/fileatime/filemtime/filectime/touch/getpid/getuid/getgid/umask/uname/usort，修复usort stub |
| 2026-03-26 | 安全增强 | ARGV/ARGC 魔术常量 | 将 $argv/$argc 变量改为 ARGV/ARGC 运行时魔术常量，编译期禁止赋值，提升安全性 |
| 2026-03-26 | Phase 8 | 流类型系统启动 | 启动Phase 8开发，目标：TypeStream类型 + STDIN/STDOUT/STDERR流资源化 + fread/fwrite/fopen等IO函数 |
| 2026-03-26 | 设计文档 | 网络框架设计 | 新增D24决策文档：IO多路复用 + 回调注册表模式，提供TCP/UDP/事件循环/二进制处理/DNS完整API |
| 2026-03-26 | 设计文档 | 流类型系统设计 | 新增D23决策文档：TypeStream类型、streamValue结构、标准流预定义、IO函数扩展方案 |
| 2026-03-26 | Phase 7.9 | VM/反射函数完成 | 新增10个函数：func_num_args/func_get_arg/func_get_args/function_exists/is_callable/get_defined_functions/get_defined_constants/jpl_version/utf8_encode/utf8_decode |
| 2026-03-26 | Phase 7.7 | URL 处理函数完成 | 新增5个函数：urlencode/urldecode/rawurlencode/rawurldecode/parse_url |
| 2026-03-26 | Phase 7.6 | 文件系统扩展完成 | 新增9个函数：chdir/rename/unlink/realpath/is_readable/is_writable/chmod/scandir/glob，文件I/O共26个函数 |
| 2026-03-26 | Phase 7.6 | cwd() 函数 | 新增获取当前工作目录函数 |
| 2026-03-26 | Phase 7.5 | now() 增强 | 支持格式化参数 now("Y-m-d")，新增 millisecond/timezone/timestamp 字段 |
| 2026-03-26 | Phase 7.8 | STDIO 常量 | 新增 STDIN/STDOUT/STDERR 字符串常量标识 |
| 2026-03-26 | Phase 7.8 | EOL 常量 | 新增平台换行符常量 |
| 2026-03-26 | Phase 7.5 | 日期时间扩展 | 新增7个函数：getdate/gettimeofday/strftime/gmdate/localtime/mktime/gmmktime |
| 2026-03-26 | Phase 7.4 | 数学扩展 | 新增18个三角/双曲/对数函数 |
| 2026-03-26 | Phase 7.3 | 字符串扩展 | 新增23个字符串处理函数 |
| 2026-03-26 | Phase 10.3 | 完整表达式插值 | 支持对象访问、数组索引、算术运算等15种表达式类型 |
| 2026-03-26 | Phase 10.2 | 字符串插值 MVP | 实现 Ruby 风格 `#{$var}` 语法，含转义支持 |
| 2026-03-26 | Phase 10.1 | 多行字符串 | Python 风格三引号语法 `'''` 和 `"""` |
| 2026-03-26 | Phase 7.2 | 数组扩展 | 新增12个数组操作函数 |
| 2026-03-26 | Phase 7.1 | 类型与转换 | 新增13个类型检查和转换函数 |
| 2026-03-26 | Phase 7.8 | 魔术常量 | 实现 __FILE__/__DIR__/__LINE__/__TIME__/__DATE__/__OS__/JPL_VERSION |

---

## 当前状态

**项目状态**：维护模式

**当前阶段**：功能完整，进入维护

**已完成**：
- 尾调用优化 (TCO) ✅ 完成（自递归栈帧复用，支持 10000+ 深度递归）
- static 变量 ✅ 完成（函数级持久化变量）
- Range 惰性加载优化 ✅ 100% 完成（使用 Go 1.25 iter.Seq2 原生泛型迭代器）
- 正则字面量语法 ✅ 100% 完成（`#/pattern/flags#` + `=~` + match/case 正则模式 + `as $var` 捕获绑定）
- 已知测试问题修复 ✅ 全部通过（5 个预存失败测试已修复）
- `jpl fmt` 代码格式化 ✅ 100% 完成（注释保留、4空格缩进、幂等性）
- Phase 11.4 stub 函数 ✅ 100% 完成（fseek/ftell/rewind/ftruncate/fgetcsv）
- print/println 流参数 ✅ 完成（支持 `print(STDERR, "msg")` 语法）
- 常量折叠优化 ✅ 完成（嵌套表达式编译期求值）
- `jpl lint` 静态分析 ✅ 完成（未使用变量/未定义变量/死代码检测，10 个测试）
- 包管理器设计文档 ✅ 完成（docs/PACKAGE_MANAGER.md）
- 包管理器 Phase A ✅ 完成（基础 add/remove/install）
- 包管理器 Phase B ✅ 完成（传递依赖解析、循环检测、全局缓存、jpl list）
- 包管理器 Phase C ✅ 完成（semver 版本约束 + 集成到 add/install）
- `jpl init` 项目初始化 ✅ 完成（创建 jpl.json、示例文件、jpl_modules/ 目录）
- Resolver 集成 ✅ 完成（`jpl install --resolve` 完整依赖解析）
- `jpl update/outdated` 命令 ✅ 完成（依赖更新和过时检查）
- 字符串插值解析修复 ✅ 完成（插值后逗号分隔参数支持）
- BigInt/BigDecimal 类型系统 ✅ 完成（独立类型枚举、编译器支持、类型检查函数）
- `jpl task` 任务系统 ✅ 完成（支持简单/复杂任务定义、依赖解析、循环检测）
- 并行依赖安装 ✅ 完成（goroutine + 信号量控制并发，`--jobs/-j` 标志）
- 示例项目 ✅ 完成（package-manager、tasks 示例）

**已知问题**：无

**搁置计划**：
- LSP 支持 — 投入产出比低，lint + fmt 已覆盖核心价值
- 调试器 (DAP) — 风险高，--debug + REPL 已提供基本调试能力

**最后更新**：2026-04-03（更好的错误消息 + 尾调用优化 + static 变量完成）

---

## 近期完成（2026-04-03）

### 更好的错误消息 ✅

**实现内容**：
- 运行时错误现在显示行号 + 源码上下文（箭头标记 + 前后 2 行）
- Program 结构体新增 `Source` 和 `SourceLines` 字段存储源代码
- 编译器在编译时存储源代码到 Program
- VM 在错误发生时自动从 `vm.currentLine` 附加行号到 RuntimeError
- RuntimeError 新增 `FormatWithContext()` 方法，生成带上下文的格式化输出
- CLI（run/eval/repl）全部更新为使用新的错误格式

**改动文件**：
- `engine/bytecode.go` — Program 添加 `Source` + `SourceLines` 字段
- `engine/compiler.go` — Compiler 添加 `source` 字段；`CompileStringWithName`/`CompileStringWithGlobals`/`buildProgram` 存储源代码；新增 `CompileWithSource`
- `engine/errors.go` — RuntimeError 新增 `FormatWithContext(sourceLines []string)` 方法
- `engine/vm.go` — 新增 `enrichError()` 辅助方法；`run()` 主循环错误处理附加行号；`OP_CALL`/`OP_TAIL_CALL` 错误返回使用 `enrichError`
- `cmd/jpl/run.go` — 错误显示改用 `FormatWithContext`
- `cmd/jpl/eval.go` — 同上
- `cmd/jpl/repl.go` — 新增 `Program` 字段；错误显示带源码上下文
- `engine/vm_test.go` — 新增 3 个测试：`TestRuntimeErrorSourceContext`、`TestRuntimeErrorSourceContextMultiLine`、`TestRuntimeErrorFormatFallback`

**输出示例**：
```
runtime error at line 3, column 0: something went wrong
   1 | fn greet() {
   2 |     $msg = "hello"
 → 3 |     throw "something went wrong"
   4 | }
   5 | 
```

### 尾调用优化 (TCO) ✅

**实现内容**：
- 自递归尾调用栈帧复用，消除递归调用栈增长
- 编译器检测尾位置调用（`return func(args)`），发出 `OP_TAIL_CALL` 指令
- VM 自递归检测：通过闭包身份（`cl.function == vm.function`）或函数名匹配
- 非自递归尾调用（如 `return $fn($x)`）正常执行并正确返回结果
- `opReturn` 增加尾调用返回传播：当恢复帧的 IP 已在函数末尾时，继续向上传递返回值

**改动文件**：
- `engine/compiler.go` — `compileReturnStmt` 发出 `OP_TAIL_CALL` 替代 `OP_CALL` + `OP_RETURN`；隐式 return 检查跳过 `TAIL_CALL` 后的代码
- `engine/vm.go` — 重写 `opTailCall`：自递归时原地更新参数 + 跳转 IP=0；非自递归时 `tailCallClosure`/`tailCallJPLFunction` 正常调用；`opReturn` 增加尾调用返回传播循环
- `engine/stress_test.go` — 更新 `TestStressStackOverflow` 为非尾递归函数（尾递归不再溢出）
- `engine/scope_test.go` — 新增 4 个深度递归测试：`TestTailCallDeepRecursion`(5000)、`TestTailCallVeryDeepRecursion`(10000)、`TestTailCallFactorialDeep`(20!)、`TestTailCallWithMultipleBranches`(Collatz)

**验证结果**：
```jpl
sum(5000)    → 12502500  ✅ 无栈溢出
counter(10000) → "done"  ✅ 无栈溢出
fact(20, 1)  → 2432902008176640000  ✅
collatz(27)  → 111  ✅
apply($n -> $n * 3, 7) → 21  ✅ 非自递归尾调用正确
```

**设计要点**：
- 仅优化自递归调用（调用自身），互递归和调用其他函数不优化
- 自递归通过闭包身份比较检测，因为 JPL 函数以闭包形式存储
- 编译器不再为 `TAIL_CALL` 后的函数体添加隐式 `LOADNULL` + `RETURN`

### static 变量 ✅

**实现内容**：
- 函数级持久化变量，调用之间保持其值
- 实现方式：静态变量存储在全局变量中，使用 `_static:funcName::varName` 前缀
- 初始值仅在首次调用时设置，后续调用跳过初始化
- 支持无初始值声明（默认为 null）
- 每个函数的静态变量独立命名空间

**改动文件**：
- `engine/compiler.go` — `compileStaticDecl` 编译 static 声明，通过全局变量存储
- `engine/scope_test.go` — 新增 5 个测试：多函数独立命名空间、字符串类型、持久化、初始化唯一性、复杂表达式
- `examples/basic/static-variables.jpl` — 新增示例文件

**使用示例**：
```jpl
fn counter() {
    static $count = 0;
    $count = $count + 1;
    return $count;
}

counter()  // → 1
counter()  // → 2
counter()  // → 3
```

---

## 近期完成（2026-04-02）

### 包管理器 Phase A + B + C ✅

**实现内容**：
- 基于 Git 的包管理器，支持 `jpl init/add/remove/install/list` 命令
- 扁平化安装策略：所有依赖安装到 `jpl_modules/`，传递依赖自动解析
- 三色标记 DFS 算法检测循环依赖
- 全局缓存 `~/.jpl/packages/` 避免重复克隆
- 自动生成/更新 `jpl.json` 清单文件和 `jpl.lock.yaml` 锁文件
- `jpl init` 项目初始化，创建清单文件、示例文件和依赖目录
- 语义化版本约束支持（`^`、`~`、`>=`、`>`、`<`、`<=`、`=`）
- 版本约束集成到 add/install 命令
- Resolver 集成到 install 命令（`--resolve` 模式）

**新增文件**：
- `pkg/pm/manifest.go` — 清单文件读写、源地址解析（~200 行）
- `pkg/pm/init.go` — 项目初始化逻辑（~90 行）
- `pkg/pm/git.go` — git 操作：clone/checkout/获取 commit hash + ListTags/CloneWithConstraint（~290 行）
- `pkg/pm/resolver.go` — DFS 传递依赖解析 + 循环检测（~310 行）
- `pkg/pm/cache.go` — 全局包缓存（~240 行）
- `pkg/pm/semver.go` — 版本约束封装（~180 行）
- `pkg/pm/manifest_test.go` — 清单解析测试（9 个）
- `pkg/pm/resolver_test.go` — 解析算法测试（16 个）
- `pkg/pm/semver_test.go` — 版本约束测试（~50 个）
- `cmd/jpl/pm.go` — init/add/remove/install/list 子命令（~700 行）

**修改文件**：
- `engine/lockfile.go` — 添加 `PkgEntry` 结构体和 `Packages` 字段

**外部依赖**：
- `github.com/Masterminds/semver/v3` v3.4.0 — 语义化版本解析
- `pkg/pm/manifest_test.go` — 清单解析测试（9 个）
- `pkg/pm/resolver_test.go` — 解析算法测试（16 个）
- `pkg/pm/semver_test.go` — 版本约束测试（~50 个）
- `cmd/jpl/pm.go` — init/add/remove/install/list 子命令（~700 行）

**修改文件**：
- `engine/lockfile.go` — 添加 `PkgEntry` 结构体和 `Packages` 字段

**外部依赖**：
- `github.com/Masterminds/semver/v3` v3.4.0 — 语义化版本解析

**CLI 命令**：

| 命令 | 说明 | 示例 |
|------|------|------|
| `jpl init [dir]` | 初始化项目 | `jpl init my-project` |
| `jpl init --name <name>` | 指定项目名称 | `jpl init --name my-app` |
| `jpl init --desc <desc>` | 指定描述 | `jpl init --desc "My app"` |
| `jpl init --no-example` | 不创建示例文件 | `jpl init --no-example` |
| `jpl add <source>` | 添加依赖 | `jpl add https://github.com/user/lib.git` |
| `jpl add <source>@<constraint>` | 添加依赖（版本约束） | `jpl add lib.git@^1.2.3` |
| `jpl add <source> --name <name>` | 自定义导入名 | `jpl add ../my-lib --name utils` |
| `jpl add <source> --version <tag>` | 指定版本 | `jpl add lib.git --version v1.0.0` |
| `jpl remove <name>` | 移除依赖 | `jpl remove utils` |
| `jpl install` | 安装全部依赖 | `jpl install` |
| `jpl install --resolve` | 使用依赖解析器 | `jpl install --resolve` |
| `jpl install --no-cache` | 禁用缓存 | `jpl install --no-cache` |
| `jpl update` | 更新所有依赖 | `jpl update` |
| `jpl update <name>` | 更新指定依赖 | `jpl update utils` |
| `jpl outdated` | 检查过时的依赖 | `jpl outdated` |
| `jpl list` | 列出依赖 | `jpl list` |

**设计要点**：
- 扁平安装：传递依赖也安装到项目根 `jpl_modules/`，通过 `resolvePath()` 向上遍历自动解析
- 循环检测：三色标记 DFS，检测到循环时报错并显示循环路径
- 版本冲突：检测并记录冲突警告
- 全局缓存：`~/.jpl/packages/<owner>/<repo>/<commit>/`，避免重复克隆
- 版本约束：支持 `^`、`~`、`>=`、`>`、`<`、`<=`、`=` 运算符

**测试覆盖**：
- 75+ 个单元测试全部通过（manifest 9 + resolver 16 + semver ~50）
- 全量测试套件通过

**Phase A 限制**（已实现）：
- ✅ 基础 add/remove/install/list
- ✅ Git URL 和本地路径源
- ✅ `@tag` 和 `#branch` 版本指定
- ✅ 自动生成 `jpl.json` 和 `jpl.lock.yaml`

**Phase B 新增**（已实现）：
- ✅ 传递依赖自动解析
- ✅ 循环依赖检测（三色标记 DFS）
- ✅ 全局缓存（`~/.jpl/packages/`）
- ✅ `jpl list` 命令

**Phase C 新增**（已实现）：
- ✅ semver 版本解析（基于 github.com/Masterminds/semver/v3）
- ✅ 版本约束求解（^, ~, >=, >, <, <=, =）
- ✅ 版本冲突检测
- ✅ 版本约束集成到 add/install 命令
- ✅ CloneWithConstraint 根据约束自动选择最佳版本

**Resolver 集成**（已实现）：
- ✅ `jpl install --resolve` 完整依赖解析模式
- ✅ DFS 遍历所有传递依赖
- ✅ 拓扑排序安装顺序
- ✅ 循环依赖检测和报错
- ✅ 缓存检查和写入

---

### Bug 修复与类型系统增强 ✅

**修复内容**：

1. **字符串插值解析修复**
   - 问题：`println "pop: #{$last}", $arr` 解析失败（unexpected token COMMA）
   - 原因：`isValueToken` 函数缺少 `STRING_FRAG` token 类型
   - 修复：`parser/parser.go` 在 `isValueToken` 中添加 `token.STRING_FRAG`

2. **BigInt/BigDecimal 类型系统**
   - 问题：大整数被错误解析为 float，`typeof` 返回 "float"
   - 原因：编译器 `compileNumberLiteral` 忽略 token 类型，只尝试 int64/float64
   - 修复：
     - `engine/value.go` 添加 `TypeBigInt` 枚举
     - `engine/compiler.go` 检查 BIGINT/BIGDECIMAL token，创建对应值类型
     - `buildin/typecheck.go` 添加 `is_bigint()`、`is_bigdecimal()` 函数

3. **示例文件语法修复**
   - `arrays.jpl`：修复插值字符串后的逗号问题（已通过解析器修复解决）
   - `datatypes.jpl`：无需修改（已通过解析器修复解决）
   - `control-flow.jpl`：`var` 关键字改为 `$` 前缀，添加未定义变量初始化

**修改文件**：
- `parser/parser.go` — isValueToken 添加 STRING_FRAG
- `engine/value.go` — 添加 TypeBigInt 枚举，BigIntValue.Type() 返回 TypeBigInt
- `engine/compiler.go` — compileNumberLiteral 检查 token 类型，导入 math/big
- `engine/value_ops.go` — IsNumeric 包含 TypeBigInt
- `buildin/typecheck.go` — 添加 is_bigint/is_bigdecimal，更新 is_numeric/is_scalar
- `examples/basic/control-flow.jpl` — 修复 var 语法

**验证结果**：
```bash
# 字符串插值
println "pop: #{$last}, 数组:", $arr  ✅

# BigInt 类型
$x = 999999999999999999999
typeof($x)       → bigint
is_bigint($x)    → true
is_int($x)       → false
is_numeric($x)   → true

# 示例文件
jpl check examples/basic/*.jpl  ✅
jpl run examples/basic/*.jpl    ✅

go test ./...  ✅ 全部通过
```

---

## 近期完成（2026-03-30）

### Range 惰性加载优化 ✅

**实现内容**：
- 使用 Go 1.25 原生泛型迭代器 `iter.Seq2[int, engine.Value]` 替代自定义迭代器接口
- 创建 `toIter(v engine.Value)` 函数，统一处理 Array 和 Range 的迭代
- 创建 `getLength(v engine.Value)` 函数，高效获取长度（O(1) 复杂度）
- 更新 17 个函数使用 `toIter`：map, filter, reduce, find, some, every, contains, reject, partition, first, last, take, drop, sum, arrayMin, arrayMax, size
- 清理未使用的 `iterFunc`、`iterableWithLen` 和 `getIterable` 类型/函数

**技术细节**：
- `iter.Seq2` 是 Go 1.25 引入的原生泛型迭代器，支持 `for i, val := range toIter(value)` 语法
- Range 迭代采用惰性生成，而非预先展开为数组
- 保持与 Array 完全一致的函数式 API

---

## 近期完成（2026-03-28）

### 编码功能：Email编码和HTML实体编码 ✅
- 实现 quoted_printable_encode 和 quoted_printable_decode 函数（邮件传输编码）
- 实现 htmlentities、html_entity_decode 和 get_html_translation_table 函数（HTML实体编码/解码）
- 所有函数已添加到 buildin/string.go 并注册到 strings 模块
- 需要后续添加单元测试

### 进程 API ✅

**实现内容**：
- 参考 PHP/Python/Node.js，实现 21 个进程管理函数
- 分 P0-P3 四个优先级实现，跳过低必要性函数
- 支持命令执行、环境变量、进程管理、信号处理

**改动文件**：
- `buildin/process_ext.go` — 新建，21 个进程函数（~1100 行）
- `buildin/builtin.go` — 注册 ProcessExt 模块

**新增函数列表**：

| 类别 | 函数 | 说明 |
|------|------|------|
| 命令执行 | exec, system, shell_exec | 执行系统命令 |
| 环境变量 | getenv, setenv, putenv | 管理环境变量 |
| 进程信息 | getpid, getppid, getlogin, hostname, tmpdir | 获取进程信息 |
| 进程管道 | proc_open, proc_close, proc_wait, proc_status | 管理进程管道 |
| 进程控制 | spawn, kill, waitpid, fork | 创建和控制子进程 |
| 其他 | usleep, pipe, sigwait | 延迟、管道、信号等待 |

**使用示例**：
```jpl
// 执行命令
$output = exec("ls -la")
$code = system("ping -c 1 8.8.8.8")

// 环境变量
setenv("APP_ENV", "production")
$env = getenv("HOME")

// 进程管理
$proc = spawn("sleep", ["5"])
kill($proc.pid, 9)  // SIGKILL
$code = waitpid($proc)

// 多进程并发
$procs = []
for ($i = 0; $i < 3; $i = $i + 1) {
    push($procs, spawn("echo 'Task #{$i}'"))
}
for ($proc of $procs) {
    waitpid($proc)
}

// 模块导入
import "process"
$host = process.hostname()
```

**测试覆盖**：手动测试验证

### 事件循环架构重构 ✅

**实现内容**：
- 从自定义 epoll/kqueue 重构为 goroutine + context 模式
- 设计通用事件注册表接口
- 网络事件注册与事件循环解耦

**改动文件**：
- 删除: `engine/evpoll_linux.go`, `engine/evpoll_bsd.go`, `engine/evpoll_stub.go`
- 重写: `buildin/ev.go` — Event Loop 核心
- 重写: `buildin/net.go` — 网络事件注册

**架构变更**：
- 删除 ~500 行平台相关代码
- 每个事件处理器独立的 goroutine
- 使用 context.Context 控制生命周期

### 异步文件 IO（asyncio）✅

**实现内容**：
- 参考 Python asyncio，实现异步文件 IO API
- 支持一次性读写、流式读取、批量操作
- 实现文件访问冲突检测

**改动文件**：
- `buildin/fileio_async.go` — 新建，11 个异步 IO 函数

**新增函数**：
- file_get_async, file_put_async, file_append_async
- file_get_bytes, file_put_bytes
- file_read_lines, file_read_chunks
- file_get_batch, file_put_batch, file_parallel
- file_with_lock

**模块导入**：`import "asyncio"`

---

## 近期完成（2026-03-27）

### 流类型系统 ✅

**实现内容**：
- 新增 TypeStream 值类型，STDIN/STDOUT/STDERR 升级为流资源
- 实现 streamValue 结构体，组合 io.Reader/io.Writer/io.Closer
- 实现 8 个核心 IO 函数：fopen, fread, fgets, fwrite, fclose, feof, fflush, stream_get_meta_data
- 实现 is_stream() 类型检查函数
- 新增 5 个 stub 函数（需 Phase 8 完善）：fseek, ftell, rewind, ftruncate, fgetcsv

**改动文件**：
- `engine/stream.go` — 新建，streamValue 实现（292 行）
- `engine/stream_test.go` — 新建，33 个测试
- `engine/value.go` — 新增 TypeStream 枚举
- `buildin/const.go` — STDIN/STDOUT/STDERR 改为流类型
- `buildin/io.go` — 新增 8 个 IO 函数
- `buildin/typecheck.go` — 新增 is_stream() 函数

**新增函数列表**：

| 类别 | 函数 | 说明 |
|------|------|------|
| 流操作 | fopen(path, mode) | 打开文件流（r/w/rw） |
| 流操作 | fread(stream, length) | 读取指定字节数 |
| 流操作 | fgets(stream) | 读取一行 |
| 流操作 | fwrite(stream, data) | 写入数据 |
| 流操作 | fclose(stream) | 关闭流 |
| 流操作 | feof(stream) | 检查是否到达末尾 |
| 流操作 | fflush(stream) | 刷新缓冲区 |
| 流操作 | stream_get_meta_data(stream) | 获取流元数据 |
| 类型检查 | is_stream(value) | 检查是否为流类型 |

**测试覆盖**：33 个单元测试

**使用示例**：
```jpl
// 打开文件流
$f = fopen("data.txt", "r")
$content = fread($f, 1024)
fclose($f)

// 使用标准流
fwrite(STDOUT, "Hello, World!\n")
$line = fgets(STDIN)

// 流元数据
$info = stream_get_meta_data($f)
println($info["mode"])  // → "r"
```

### 内置函数补全 ✅

**实现内容**：
- 根据 Jx9 对比分析，补充缺失的常用内置函数
- 新增 23 个函数，覆盖文件 IO、Hash、系统信息、数组排序等领域
- 修复 usort stub 为完整实现

**改动文件**：
- `buildin/hash.go` — 新增 md5_file, sha1_file
- `buildin/fileio.go` — 新增 9 个函数（file_get_contents, file_put_contents, copy, readfile, pathinfo 等）
- `buildin/system.go` — 新建，11 个系统函数
- `buildin/array.go` — 修复 usort + 注册
- `buildin/builtin.go` — 注册 System 模块

**新增函数列表**：

| 类别 | 函数 | 说明 |
|------|------|------|
| Hash | md5_file, sha1_file | 计算文件 MD5/SHA1 |
| 文件 IO | file_get_contents, file_put_contents, copy, readfile | 文件读写复制 |
| 文件 IO | pathinfo | 返回路径信息对象 |
| 系统 | disk_free_space, disk_total_space | 磁盘空间查询 |
| 系统 | fileatime, filemtime, filectime | 文件时间戳 |
| 系统 | touch, umask | 文件操作 |
| 系统 | getpid, getuid, getgid | 进程/用户信息 |
| 系统 | uname | 系统信息对象 |
| 数组 | usort | 自定义排序（修复 stub） |

**测试覆盖**：77 个新测试用例

---

## 近期完成（2026-03-26）

### JSON 解码智能数字解析 ✅

**改动内容**：
- 实现 `json_decode()` 智能数字解析，支持 BigInt 和 BigDecimal
- **精度保持**：使用 `json.Decoder.UseNumber()` 获取原始数字字符串，避免 float64 精度丢失
- **内存优化**：根据数值自动选择最优类型（Int < Float < BigInt/BigDecimal）
- **科学计数法支持**：`1e10` → Int, `1e20` → BigInt, `1.5e-5` → Float

**实现细节**：
- `parseJSONNumber()`: 判断数字类型（整数/小数/科学计数法），路由到对应解析器
- `parseScientificNotation()`: 使用 `big.Rat.SetString()` 精确解析科学计数法
- 通过 `big.Rat.IsInt()` 判断结果是否为整数
- Float64 精度检查：格式化后重新解析回 Rat 比较，确保精度无损

**改动文件**：
- `buildin/json.go` — 重写 `builtinJSONDecode()` 使用 Decoder.UseNumber()，新增 `parseJSONNumber()`, `parseScientificNotation()` 等函数，利用 big.Rat 精确解析
- `buildin/json_test.go` — 新增 `TestJSONNumberParsing` 测试用例，覆盖整数、小数、科学计数法、负数、大数等场景

**类型选择策略**：
| 输入格式 | int64 范围内 | 超出 int64 | 小数部分 |
|---------|------------|-----------|---------|
| 纯整数 | Int | BigInt | N/A |
| 科学计数法 | Int (如 1e10) | BigInt (如 1e20) | Float (如 1.5e-5) |
| 普通小数 | N/A | N/A | Float（精确时）或 BigDecimal |

**使用示例**：
```php
json_decode("123")                    // Int: 123
json_decode("99999999999999999999")   // BigInt (超出 int64)
json_decode("1e10")                   // Int: 10000000000
json_decode("1e20")                   // BigInt: 100000000000000000000
json_decode("1.5")                    // Float: 1.5
```

### 运行时魔术常量：ARGV/ARGC ✅

**改动内容**：
- 将 `$argv` 和 `$argc` 变量改为 `ARGV` 和 `ARGC` 运行时魔术常量
- **安全性提升**：从可变变量改为编译期禁止赋值的只读常量
- **实现方式**：新增 `OP_GETARGV` 和 `OP_GETARGC` 字节码指令，运行时从 VM 获取参数

**改动文件**：
- `engine/bytecode.go` — 新增 OP_GETARGV、OP_GETARGC 操作码
- `engine/compiler.go` — compileIdentifier() 识别 ARGV/ARGC，compileAssign() 禁止赋值
- `engine/vm.go` — 新增 args 字段、SetArgs() 方法、opGetArgv()、opGetArgc() 指令实现
- `cmd/jpl/run.go` — 移除 $argv/$argc 变量设置，改用 vm.SetArgs()
- `README.md` — 更新示例代码（$argc → ARGC, $argv → ARGV）
- `API.md` — 新增"运行时魔术常量"章节，说明 ARGV/ARGC 使用方法和安全特性

**安全特性**：
- 编译期禁止对 ARGV/ARGC 赋值，尝试修改会直接报错
- 运行时从 VM 内部获取，不可被脚本篡改
- 符合常量命名惯例（Ruby 风格）

---

## Phase 7：内置函数完善（已完成 - 100% 完成）

| 子阶段 | 函数/常量数 | 状态 | 完成率 |
|--------|-------------|------|--------|
| 7.1 类型与转换 | 13 个 | ✅ 完成 | 100% |
| 7.2 数组扩展 | 12 个 | ✅ 完成 | 100% |
| 7.3 字符串扩展 | 23 个 | ✅ 完成 | 100% |
| 7.4 数学扩展 | 32 个 | ✅ 完成 | 100% |
| 7.5 日期时间 | 12 个 | ✅ 完成 | 100% |
| 7.6 文件系统 | 26 个 | ✅ 完成 | 100% |
| 7.7 URI/URL | 5 个 | ✅ 完成 | 100% |
| 7.8 常量 | 11 个 | ✅ 完成 | 100% |
| 7.9 VM/反射 | 12 个 | ✅ 完成 | 100% |

**Phase 7 总计**：107项（68函数 + 11常量 + 28其他）

---

### Phase 7.1 类型与转换 ✅

**已实现函数（13个）**：
- [x] `is_real` / `is_double` — is_float 别名
- [x] `is_integer` / `is_long` — is_int 别名  
- [x] `is_numeric` — 检查是否为数字
- [x] `is_scalar` — 检查是否为标量
- [x] `intval` — 转换为整数（支持多进制）
- [x] `floatval` — 转换为浮点数
- [x] `strval` — 转换为字符串
- [x] `boolval` — 转换为布尔值
- [x] `empty` — 检查值是否为空

**文件变更**：
- `buildin/typecheck.go` — 6个新类型检查函数
- `buildin/typeconvert.go` — 4个类型转换函数
- `buildin/typecheck_test.go` — 9个新测试
- `buildin/typeconvert_test.go` — 5个测试套件
- `buildin/builtin.go` — 注册类型转换模块

---

### Phase 7.2 数组扩展 ✅

**已实现函数（12个）**：
- [x] `count` / `sizeof` — 数组长度统计
- [x] `array_key_exists` / `key_exists` — 键存在检查（支持负数索引）
- [x] `array_merge` — 合并多个数组（支持标量元素）
- [x] `array_sum` — 数组求和
- [x] `array_product` — 数组乘积
- [x] `array_values` — 获取所有值
- [x] `array_diff` — 数组差集
- [x] `array_intersect` — 数组交集
- [x] `in_array` — 值存在检查（includes 别名）
- [x] `array_copy` — 深度复制

**文件变更**：
- `buildin/array.go` — 12个新函数
- `buildin/array_phase7_test.go` — 11个测试套件

---

### Phase 7.3 字符串扩展 ✅

**已实现函数（23个）**：
- [x] `implode` / `explode` / `chop` — 别名（join/split/rtrim）
- [x] `ltrim` / `rtrim` — 左右修剪（支持自定义字符集）
- [x] `strcmp`, `strcasecmp` — 字符串比较
- [x] `strncmp`, `strncasecmp` — 前N字符比较
- [x] `stripos`, `strrpos`, `strripos` — 查找位置
- [x] `strstr`, `stristr`, `strchr` — 子串查找
- [x] `sprintf`, `printf`, `vsprintf`, `vprintf` — 格式化输出
- [x] `ord`, `chr` — ASCII 转换
- [x] `nl2br` — 换行转 `<br>`
- [x] `bin2hex` — 二进制转十六进制

**文件变更**：
- `buildin/string.go` — 扩展23个函数
- `buildin/string_phase7_test.go` — 13个测试套件

---

### Phase 7.4 数学扩展 ✅

**已实现函数（18个新增）**：
- [x] `sin`, `cos`, `tan` — 三角函数
- [x] `asin`, `acos`, `atan`, `atan2` — 反三角函数
- [x] `sinh`, `cosh`, `tanh` — 双曲函数
- [x] `log`, `log10`, `exp`, `pi` — 对数/指数
- [x] `fmod` — 浮点取模
- [x] `hypot` — 斜边计算
- [x] `deg2rad`, `rad2deg` — 角度/弧度转换

**数学函数总计**：32个（原有14个 + 新增18个）

**文件变更**：
- `buildin/math.go` — 扩展18个函数
- `buildin/math_test.go` — 更新函数数量期望

---

### Phase 7.5 日期时间扩展 ✅

**已实现函数（7个新增）**：
- [x] `getdate` — 返回日期数组
- [x] `gettimeofday` — 返回时间信息
- [x] `strftime` — 格式化本地时间
- [x] `gmdate` — 格式化 GMT 时间
- [x] `localtime` — 返回本地时间信息
- [x] `mktime` — 生成本地时间戳
- [x] `gmmktime` — 生成 GMT 时间戳

**增强函数**：
- [x] `now()` — 增强版，支持 `now("Y-m-d")` 格式化调用，新增字段：`millisecond`, `timezone`, `timestamp`

**日期时间函数总计**：12个

**文件变更**：
- `buildin/datetime.go` — 7个新函数 + now() 增强
- `buildin/datetime_phase7_test.go` — 测试文件

---

### Phase 7.6 文件系统扩展 ✅

**已实现函数（9个新增）**：
- [x] `chdir` — 改变当前工作目录
- [x] `rename` — 重命名/移动文件
- [x] `unlink` — 删除文件
- [x] `realpath` — 规范化绝对路径
- [x] `is_readable` — 检查可读性
- [x] `is_writable` — 检查可写性
- [x] `chmod` — 修改文件权限
- [x] `scandir` — 扫描目录内容
- [x] `glob` — 模式匹配查找文件

**新增函数**：
- [x] `cwd()` — 获取当前工作目录（独立实现）

**文件 I/O 函数总计**：26个（原有17个 + 新增9个）

**文件变更**：
- `buildin/fileio.go` — 9个新函数 + cwd() 函数
- `buildin/fileio_phase7_test.go` — 测试文件

---

### Phase 7.7 URI/URL 处理 ✅

**已实现函数（5个）**：
- [x] `urlencode` / `urldecode` — 标准 URL 编码（空格→+）
- [x] `rawurlencode` / `rawurldecode` — 原始 URL 编码（空格→%20）
- [x] `parse_url` — 解析 URL 返回组成部分数组

**设计说明**：
- UU 编码（convert_uuencode/decode）因使用场景少未实现
- `parse_url` 暂不支持第二个参数（指定返回部分），可通过数组访问

**文件变更**：
- `buildin/url.go` — 5个 URL 处理函数
- `buildin/url_test.go` — 测试文件

---

### Phase 7.8 魔术常量 ✅

**已实现常量（8个）**：
- [x] `__FILE__` — 当前源文件名
- [x] `__DIR__` — 当前源文件目录
- [x] `__LINE__` — 当前代码行号
- [x] `__TIME__` — 编译时间（HH:MM:SS）
- [x] `__DATE__` — 编译日期（Mon Jan 2 2006）
- [x] `__OS__` — 操作系统名称
- [x] `JPL_VERSION` — JPL 版本号
- [x] `EOL` — 平台换行符（\n 或 \r\n）

**平台常量（3个）**：
- [x] `STDIN` — 标准输入流标识符
- [x] `STDOUT` — 标准输出流标识符
- [x] `STDERR` — 标准错误流标识符

**设计说明**：
- STDIN/STDOUT/STDERR 目前作为字符串常量（"stdin"/"stdout"/"stderr"）
- 用于标识标准 IO 流，为未来 IO 重定向功能预留
- 在编译器层通过 `getMagicConstant()` 方法在编译时替换为常量值

**文件变更**：
- `engine/compiler.go` — 编译器扩展支持魔术常量
- `buildin/const.go` — EOL 和 STDIO 常量注册
- `engine/magic_const_test.go` — 测试文件

### Phase 7.9 VM/反射函数 ✅

**已实现函数（10个）**：
- [x] `func_num_args` — 返回当前函数参数数量
- [x] `func_get_arg` — 获取指定索引的参数
- [x] `func_get_args` — 获取所有参数数组
- [x] `function_exists` — 检查函数是否存在
- [x] `is_callable` — 检查是否可调用
- [x] `get_defined_functions` — 获取所有已定义函数名
- [x] `get_defined_constants` — 获取所有已定义常量名
- [x] `jpl_version` — 返回 JPL 版本号
- [x] `utf8_encode` — UTF-8 编码（转十六进制）
- [x] `utf8_decode` — UTF-8 解码（从十六进制）

**设计说明**：
- func_num_args/get_arg/get_args 只能在用户定义函数内调用
- function_exists 同时检查编译函数和引擎注册函数
- is_callable 支持函数名字符串和函数值
- utf8_encode/decode 使用十六进制表示 UTF-8 字节序列

**文件变更**：
- `buildin/vmfunc.go` — 10个 VM/反射函数
- `buildin/vmfunc_test.go` — 测试文件
- `engine/vm.go` — 新增 CurrentFunction() 和 CurrentRegisters() 方法
- `engine/engine.go` — 新增 GetConstantNames() 方法
- `buildin/builtin.go` — 注册 VMFunc 模块

---

## Phase 8：流类型系统 ✅ 完成

### 实现内容

**核心组件**：
- `TypeStream` 值类型，集成到 Value 类型系统
- `streamValue` 结构体，组合 io.Reader/io.Writer/io.Closer
- `StreamMode` 枚举：StreamRead, StreamWrite, StreamReadWrite
- 标准流：STDIN/STDOUT/STDERR 升级为流类型

**新增函数（8 个）**：
- fopen(path, mode) — 打开文件流（r/w/rw）
- fread(stream, length) — 读取指定字节数
- fgets(stream) — 读取一行
- fwrite(stream, data) — 写入数据
- fclose(stream) — 关闭流
- feof(stream) — 检查是否到达末尾
- fflush(stream) — 刷新缓冲区
- stream_get_meta_data(stream) — 获取流元数据

**类型检查**：
- is_stream(value) — 检查是否为流类型

**已完善（Phase 11.4 + Phase 8.2）**：
- ✅ fseek/ftell/rewind — 文件指针定位操作
- ✅ ftruncate — 文件截断
- ✅ fgetcsv — CSV 行读取
- ✅ print/println 流参数 — 支持 `print(STDERR, "msg")` 语法

**阶段目标**：引入 TypeStream 值类型，将 STDIN/STDOUT/STDERR 从字符串常量升级为可操作的流资源。

**参考设计**：[D23. 流类型系统设计](docs/DESIGN.md#d23-流类型系统设计)

#### Phase 8.1 流类型基础 ✅ 完成

| 文件 | 任务 | 优先级 | 状态 |
|------|------|--------|------|
| `engine/value.go` | 新增 TypeStream 枚举值 | 🔴 高 | ✅ |
| `engine/value.go` | ValueType.String() 添加 TypeStream 分支 | 🔴 高 | ✅ |
| `engine/stream.go` | 新建文件，streamValue 结构体 | 🔴 高 | ✅ |
| `engine/stream.go` | 实现 Value 接口（Type/Bool/Int/Stringify等） | 🔴 高 | ✅ |
| `engine/stream.go` | 流构造函数（NewStream/NewFileStream/NewStdinStream等） | 🔴 高 | ✅ |
| `engine/stream.go` | StreamMode 枚举和流操作方法 | 🔴 高 | ✅ |
| 单元测试 | streamValue 基本操作测试 | 🟡 中 | ✅ |

#### Phase 8.2 IO 函数扩展 ✅ 完成

| 文件 | 任务 | 优先级 | 状态 |
|------|------|--------|------|
| `buildin/const.go` | 修改 STDIN/STDOUT/STDERR 注册为流类型 | 🔴 高 | ✅ |
| `buildin/io.go` | fopen(path, mode) 函数 | 🔴 高 | ✅ |
| `buildin/io.go` | fread(stream, length) 函数 | 🔴 高 | ✅ |
| `buildin/io.go` | fgets(stream) 函数 | 🔴 高 | ✅ |
| `buildin/io.go` | fwrite(stream, data) 函数 | 🔴 高 | ✅ |
| `buildin/io.go` | fclose(stream) 函数 | 🔴 高 | ✅ |
| `buildin/io.go` | feof(stream) 函数 | 🔴 高 | ✅ |
| `buildin/io.go` | fflush(stream) 函数 | 🔴 高 | ✅ |
| `buildin/io.go` | print/println 扩展流参数支持 | 🔴 高 | ✅ 已实现 |
| `buildin/typecheck.go` | is_stream() 类型检查函数 | 🟡 中 | ✅ |
| 单元测试 | IO函数测试套件 | 🟡 中 | ✅ |

#### Phase 8.3 流高级功能 ✅ 完成

| 文件 | 任务 | 优先级 | 状态 |
|------|------|--------|------|
| `engine/stream.go` | StreamRead/StreamWrite/StreamReadWrite 模式 | 🟡 中 | ✅ |
| `buildin/io.go` | is_readable/is_writable 增强 | 🟡 中 | ✅ |
| `buildin/io.go` | stream_get_meta_data 函数 | 🟡 中 | ✅ |
| `engine/stream.go` | NewBufferStream（内存流） | 🟡 中 | ✅ |
| 单元测试 | 高级功能测试 | 🟡 中 | ✅ |

**设计要点**：
- streamValue 组合 io.Reader/io.Writer/io.Closer
- 预定义标准流包装 os.Stdin/os.Stdout/os.Stderr
- pack/unpack 格式支持 PHP 风格
- 注册表模式管理事件处理器，避免协程泄漏

### Phase 8 实际工时

| 子阶段 | 预估工时 | 实际工时 | 状态 |
|--------|----------|----------|------|
| 8.1 流类型基础 | 4 天 | ~2 小时 | ✅ 完成 |
| 8.2 IO 函数扩展 | 5.5 天 | ~2 小时 | ✅ 完成 |
| 8.3 流高级功能 | 4 天 | ~1 小时 | ✅ 完成 |
| **总计** | **~13.5 天** | **~5 小时** | **✅ 100% 完成** |

---

## Phase 9 补全示例和测试

### Phase 9.1 - 补全代码示例 ✅ 完成

**目标**：完善示例代码库，覆盖所有主要功能模块

**任务清单**（遵循 3-File Rule，分批完成）：

| 批次 | 示例类型 | 状态 | 说明 |
|------|----------|------|------|
| 9.6.1 | HTTP 服务器示例 | ✅ 已完成 | 基于 TCP 实现完整 HTTP 服务器，支持路由、JSON API、HTML 响应 |
| 9.6.2 | WebSocket 协议示例 | ✅ 已完成 | 基于 TCP 实现 WebSocket 握手 |
| 9.6.3 | Redis 客户端示例 | ✅ 已完成 | RESP 协议实现 |
| 9.6.4 | 聊天室示例 | ✅ 已完成 | 多客户端广播（Unix Socket）|
| 9.6.5 | 文件上传服务器 | ✅ 已完成 | 展示二进制协议处理 |

### Phase 9.1 - 边界测试补充 ✅ 完成

**目标**：补充网络/IO/事件循环的边界测试

**新增文件**：`buildin/net_boundary_test.go`（12 个测试，~350 行）

| 测试类型 | 测试用例 | 优先级 | 状态 |
|----------|----------|--------|------|
| 网络超时 | TestConnectInvalidPort - 连接无效端口 | ⭐⭐⭐ | ✅ |
| 网络超时 | TestConnectInvalidHost - 连接无效主机 | ⭐⭐⭐ | ✅ |
| 错误处理 | TestListenInvalidAddress - 绑定无效地址 | ⭐⭐⭐ | ✅ |
| 错误处理 | TestSendOnClosedSocket - 向关闭 socket 发送 | ⭐⭐⭐ | ✅ |
| 并发测试 | TestMultipleConnections - 10 并发连接 | ⭐ | ✅ |
| 大数据量 | TestLargeDataTransfer - 256KB 数据传输 | ⭐ | ✅ |
| 快速断开 | TestRapidConnectDisconnect - 50 次快速连接断开 | ⭐⭐ | ✅ |
| UDP | TestUDPBasicSendRecv - UDP 基本收发 | ⭐ | ✅ |
| UDP | TestUDPLargePacket - UDP 大数据包 (32KB) | ⭐ | ✅ |
| 资源清理 | TestFDCleanup - FD/内存泄漏检测 | ⭐⭐ | ✅ |
| 并发测试 | TestConcurrentEcho - 并发回显 | ⭐ | ✅ |
| Unix Socket | TestUnixSocketConnection - Unix Socket 连接 | ⭐ | ✅ |

---

## Phase 10：语法增强 - 完成总结 ✅

### 10.1 多行字符串基础 ✅

**已实现功能**：
- [x] 单引号多行字符串 `'''...'''`（纯文本，不插值）
- [x] 双引号多行字符串 `"""..."""`（支持插值）
- [x] 保留换行符和缩进
- [x] 支持转义字符（\\n、\\t 等）

**测试覆盖**：7 个单元测试
- `TestTripleSingleQuoteString` - 单引号多行
- `TestTripleDoubleQuoteString` - 双引号多行
- `TestTripleQuoteJSON` - JSON 格式多行
- `TestTripleQuoteEmpty` - 空多行字符串
- `TestTripleQuoteWithEscapes` - 转义字符
- `TestTripleQuoteVsSingleQuote` - 与普通字符串对比
- `TestTripleQuoteAssignment` - 赋值和使用

### 10.2 字符串插值 MVP ✅

**已实现功能**：
- [x] 基本变量插值 `#{$var}`
- [x] 特殊变量插值 `#{$_}`
- [x] 多变量插值 `#{$a} #{$b}`
- [x] 转义支持 `\#{}`（字面量）
- [x] 多行字符串插值

**测试覆盖**：14 个单元测试（含 4 个转义测试）
- `TestStringInterpolationBasic` - 基本插值
- `TestStringInterpolationMultipleVars` - 多变量
- `TestStringInterpolationEscaped` - 转义 `#{}`
- `TestStringInterpolationEscapedWithVar` - 混合转义和插值

### 10.3 完整表达式插值 ✅

**已支持表达式类型**：
- [x] 对象属性访问 `#{$obj.prop}`
- [x] 数组索引访问 `#{$arr[0]}`
- [x] 链式对象访问 `#{$user.profile.name}`
- [x] 嵌套数组索引 `#{$matrix[0][1]}`
- [x] 算术运算 `#{$a + $b}`、`#{$x * $y + 2}`
- [x] 逻辑运算 `#{$x > 0}`
- [x] 三元运算符 `#{$cond ? 'a' : 'b'}`
- [x] 函数调用 `#{getName()}`
- [x] 字符串连接 `#{$a .. $b}`
- [x] 负数 `#{$temp}`（$temp = -10）

**测试覆盖**：15 个单元测试
- `TestInterpObjectAccess` - 对象访问
- `TestInterpArrayIndex` - 数组索引
- `TestInterpArithmetic` - 算术运算
- `TestInterpChainedProperty` - 链式属性
- `TestInterpMultipleExpressions` - 多表达式
- `TestInterpMultilineWithExpression` - 多行插值
---

## Phase 11：内置函数补全

### 实现内容

**新增函数（23 个）**：
- Hash：md5_file, sha1_file
- 文件 IO：file_get_contents, file_put_contents, copy, readfile, pathinfo
- 系统：disk_free_space, disk_total_space, fileatime, filemtime, filectime, touch, getpid, getuid, getgid, umask, uname
- 数组：usort（修复 stub）

**Stub 函数（5 个，需要 Phase 8 流类型支持）**：
- fseek, ftell, rewind, ftruncate, fgetcsv

| 批次 | 函数 | 文件 |
|------|------|------|
| 11.1 字符串 | substr_compare, substr_count, str_repeat, str_pad, str_split, strrev, htmlspecialchars, htmlspecialchars_decode, strip_tags, wordwrap, strtolower, strtoupper, chunk_split | string.go |
| 11.2 数组 | sort, rsort, usort, key, current, each, next, prev, end, reset, extract, array_map, array_walk | array.go, functional.go |
| 11.2 数组（函数式） | map, filter, reduce, find, some, every, sort, contains, reject, partition, unique, flattenDeep, difference, union, zip, unzip | functional.go |
| 11.3 数学 | rand_str, getrandmax, round, dechex, decoct, decbin, hexdec, bindec, octdec, base_convert | math.go |
| 11.4 文件IO | file_get_contents, file_put_contents, copy, readfile, pathinfo, fseek*, ftell*, rewind*, ftruncate*, fgetcsv* | fileio.go |
| 11.5 Hash | md5_file, sha1_file, crc32 | hash.go |
| 11.6 系统 | disk_free_space, disk_total_space, fileatime, filemtime, filectime, touch, umask, getpid, getuid, getgid, uname, dirname, basename, pathinfo | system.go, fileio.go |

### Phase 12：@member 闭包成员访问语法 ✅

**实现内容**：
- 添加 `@member` 语法，在闭包内访问对象成员
- 嵌套对象时 `@` 绑定到最近一层的对象（静态作用域）
- 对象字面量外使用 `@member` 报编译错误

**改动文件**：
- `token/token.go` — 添加 INSTANCE_VAR token 类型
- `lexer/lexer.go` — 添加 `@member` 词法识别
- `parser/parser.go` — 解析 INSTANCE_VAR 表达式
- `engine/compiler.go` — 添加对象上下文跟踪和编译逻辑

**设计决策**：D38

### Phase 13：运行时错误定位 ✅

**实现内容**：
- 字节码记录每条指令对应的源码行号
- VM 执行时追踪当前源码行号
- 运行时错误包含源码位置信息

**改动文件**：
- `engine/bytecode.go` — CompiledFunction 添加 SourceLines 字段
- `engine/compiler.go` — 编译时记录源码行号
- `engine/vm.go` — 执行时更新 currentLine

**设计决策**：D39


### Phase 14：delete/unset 函数 ✅

**实现内容**：
- `delete($obj, "key")` — 删除对象成员或数组元素
- `unset($var)` — 将变量设为 null

**改动文件**：
- `buildin/delete.go` — 新建，实现 delete 和 unset 函数

**设计决策**：D40

---

## Phase 15：TLS/SSL 模块

### 当前状态

**阶段**：Phase 15 - 100% 完成  
**进度**：100%（所有功能已实现并通过测试）  
**实际工时**：~1.5 小时（预估 4.5 天）

### 新增函数（11 个）

| 函数 | 状态 | 说明 |
|------|------|------|
| `tls_connect()` | ✅ | 建立 TLS 客户端连接 |
| `tls_listen()` | ✅ | 创建 TLS 服务端监听 |
| `tls_accept()` | ✅ | 接受 TLS 连接 |
| `tls_close()` | ✅ | 关闭连接 |
| `tls_send()` | ✅ | 发送加密数据 |
| `tls_recv()` | ✅ | 接收解密数据 |
| `tls_get_cipher()` | ✅ | 获取加密套件 |
| `tls_get_version()` | ✅ | 获取 TLS 版本 |
| `tls_get_cert_info()` | ✅ | 获取证书信息 |
| `tls_set_cert()` | ✅ | 设置客户端证书（提示使用 options） |
| `tls_gen_cert()` | ✅ | 生成自签名证书 |

### 支持场景

1. **HTTPS 连接** ✅ - 连接标准 HTTPS 服务
2. **自签名证书** ✅ - 连接使用自签名证书的服务
3. **双向认证（mTLS）** ✅ - 客户端和服务端相互认证
4. **证书生成** ✅ - 内置生成自签名证书对

### 测试覆盖

- ✅ TestTLSGenCert - 证书生成
- ✅ TestMTLS - 双向认证场景
- ✅ TestTLSConnectWithOptions - 连接选项
- ✅ TestTLSConnectInvalidHost - 错误处理
- ✅ TestTLSConnectInvalidPort - 端口验证
- ✅ TestTLSListenInvalidCert - 证书验证

### 设计决策

- **模块名**: `tls`（现代术语，取代 SSL）
- **与 net 对称**: API 设计保持一致
- **Go 标准库**: 使用 `crypto/tls`
- **内置证书生成**: `tls_gen_cert()` 使用 `crypto/x509` 和 `crypto/rsa`
- **证书选项**: 支持 `ca_file`, `cert_file`, `key_file`, `verify`, `server_name`

---

## Phase 16：HTTP Client 模块

### 已实现函数

| 函数 | 状态 | 说明 |
|------|------|------|
| `http_get()` | ✅ | GET 请求 |
| `http_post()` | ✅ | POST 请求 |
| `http_put()` | ✅ | PUT 请求 |
| `http_delete()` | ✅ | DELETE 请求 |
| `http_head()` | ✅ | HEAD 请求 |
| `http_patch()` | ✅ | PATCH 请求 |
| `http_request()` | ✅ | 通用 HTTP 请求 |

### 设计决策

- **模块名**: `http`
- **依赖**: `net` + `tls` 模块
- **响应对象**: 统一结构包含 status, headers, body, json() 方法
- **自动 HTTPS**: URL 以 https:// 开头时自动使用 TLS

---

### Phase 17 - 管道运算符 ✅

**实现内容**：
- 添加管道运算符 `|>`（正向管道）和 `<|`（反向管道）
- 正向管道：左结合，`a |> f(b,c)` = `f(a, b, c)`，左侧值作为首个参数
- 反向管道：右结合，`f(b,c) <| a` = `f(b, c, a)`，右侧值作为末尾参数
- 支持链式调用：`a |> f |> g` = `g(f(a))`

---

### Phase 18 - match/case 语法 ✅

**实现内容**：
- 添加 match/case 语法，类似 Rust 风格
- 支持字面量匹配（数字、字符串、布尔）
- 支持标识符绑定（捕获任意值）
- 支持通配符 `_`（匹配任意值，不绑定）
- 支持 OR 模式（`|` 连接多值）
- 支持 Guard 条件（`if` 关键字）
- 支持 match 表达式（返回值）

---

### Phase 19 - 范围语法 ✅

**实现内容**：
- 添加范围语法 `...`（半开区间）和 `..=`（闭区间）
- `1...10` = 1,2,3,4,5,6,7,8,9（不含 10）
- `1..=10` = 1,2,3,4,5,6,7,8,9,10（含 10）
- 支持负数范围

**改动文件**：
- `token/token.go` — 新增 ELLIPSIS、DOUBLE_DOT_EQUAL token
- `lexer/lexer.go` — 识别 `...` 和 `..=`
- `parser/ast.go` — 新增 RangeExpr 节点
- `parser/parser.go` — 解析范围表达式
- `engine/compiler.go` — 编译范围迭代

**设计决策**：D38

**使用示例**：
```jpl
// 半开区间 [1, 10)
for ($i = 1; $i < 10; $i++) {
    print($i)  // 1,2,3,4,5,6,7,8,9
}

// 闭区间 [1, 10]
for ($i = 1; $i <= 10; $i++) {
    print($i)  // 1,2,3,4,5,6,7,8,9,10
}

// 负数范围
for ($i = -3; $i <= -1; $i++) {
    print($i)  // -3,-2,-1
}
```

### Phase 20 - Range + 函数式编程 + 管道 ✅

**实现内容**：
- 范围表达式与管道运算符结合使用
- 内置函数 `map`、`filter`、`reduce` 支持 Range 类型
- 添加 `getIterable()` 辅助函数统一处理数组和 Range 迭代
- Range 支持负数范围

**改动文件**：
- `engine/value.go` — 为 rangeValue 添加 Start()、End()、IsInclusive() 方法
- `buildin/functional.go` — 添加 getIterable()，修改 builtinMap/Filter/Reduce 支持 Range

**使用示例**：
```jpl
// Range + 管道 + 箭头函数
$result = 1...5 |> map(($x) -> $x * 2)      // [2, 4, 6, 8]
$result = 1...10 |> filter(($x) -> $x % 2 == 0)  // [2, 4, 6, 8]
$result = 1...5 |> reduce(($acc, $x) -> $acc + $x, 0)  // 15

// Range + 管道 + Lambda 函数
$result = map(1...5, fn($x) { return $x * 2; })  // [2, 4, 6, 8]
```

### Phase 21 - 正则字面量语法 ✅ 完成

**设计文档**：[D39. 正则字面量语法设计](docs/DESIGN.md#d39-正则字面量语法设计)

**目标**：为 JPL 添加正则字面量语法 `#/pattern/flags#` 和 `=~` 匹配运算符

**实现计划**（按 3-File Rule 拆分）：

| 步骤 | 内容 | 文件数 | 预计 | 实际 | 状态 |
|------|------|--------|------|------|------|
| Step 1 | Token + Value 类型定义 | 2 | 3h | ~1h | ✅ |
| Step 2 | Lexer 正则字面量扫描 + `=~` | 1 | 4h | ~2h | ✅ |
| Step 3 | AST 节点 + Parser 解析 | 2 | 4h | ~2h | ✅ |
| Step 4 | Compiler + VM 正则执行 | 2 | 6h | ~2h | ✅ |
| Step 5 | match/case 正则分支 + `as` 绑定 | 2-3 | 6h | ~2h | ✅ |
| Step 6 | 测试 + 边界修复 | - | 4h | ~2h | ✅ |
| **总计** | | **~7 文件** | **~21h** | **~11h** | **✅ 完成** |

**核心语法**：
- 字面量：`#/pattern/flags#`
- Flags：`i`(忽略大小写) `m`(多行) `s`(dot匹配换行) `U`(非贪婪)
- 匹配运算符：`=~`
- match/case：`case #/pat/#:` 匹配，`case #/pat/# as $m:` 绑定捕获组
- 转义：`\/` `\#` `\\`
- 编译期错误检测：空模式、缺少结尾 `#`、无效正则语法

**修改文件清单**：
- `token/token.go` — 新增 `REGEX`, `MATCH_EQ` token
- `engine/value.go` — 新增 `TypeRegex`, `regexValue`, `NewRegex()`, `IsRegex()`
- `engine/value_ops.go` — `IsTruthy` 增加 regex/range 分支
- `lexer/lexer.go` — 新增 `scanRegexLiteral()`, `=~` 识别
- `parser/ast.go` — 新增 `RegexLiteral`, `RegexPattern` AST 节点
- `parser/parser.go` — 注册 prefix/infix, 扩展 `parsePattern`
- `engine/bytecode.go` — 新增 `OP_REGEX_MATCH` opcode
- `engine/compiler.go` — 新增 `compileRegexLiteral()`, `compileRegexPattern()`, parser 错误检查
- `engine/vm.go` — 新增 `opRegexMatch()`
- `buildin/typecheck.go` — 新增 `is_regex()`
- `buildin/re.go` — 新增 `re_groups_raw()` 内部函数
- `engine/regex_test.go` — 新建，20 个测试用例

---

### Phase 22 - 代码格式化 `jpl fmt` ✅ 完成

**实现内容**：
- 新增 `jpl fmt` 子命令，格式化 JPL 脚本文件
- Lexer 改造：注释不再丢弃，返回 COMMENT/BLOCK_COMMENT token
- Parser 适配：跳过注释 token，不影响现有编译逻辑
- Formatter 引擎：基于 AST 递归格式化，4 空格缩进
- 注释保留：行尾注释保持同行，leading 注释保留位置
- 对象字面量键按字母排序，保证不同运行输出一致
- 支持 `--write` 原地写入和 `--check` 检查模式

**改动文件**：
- `token/token.go` — 新增 COMMENT、BLOCK_COMMENT token 类型
- `lexer/lexer.go` — scanLineComment/scanBlockComment 返回注释 token
- `parser/parser.go` — nextToken 跳过注释 + skipNewlines/peekIsNewline 辅助方法
- `format/formatter.go` — **新建**，格式化引擎（~870 行）
- `format/formatter_test.go` — **新建**，46 个测试用例
- `cmd/jpl/fmt.go` — **新建**，CLI 子命令
- `lexer/lexer_test.go` — 更新注释测试期望值

**功能特性**：

| 特性 | 说明 |
|------|------|
| 解析验证 | 先编译检查，通过后再格式化 |
| 注释保留 | 单行 `//`、多行 `/* */` 均保留 |
| 行尾注释 | `$x = 1 // comment` 保持在同一行 |
| 4 空格缩进 | 所有层级统一 4 空格 |
| 对象键排序 | 按字母排序，消除 Go map 随机性 |
| 幂等性 | 重复 fmt 输出一致 |
| --write | 原地格式化文件 |
| --check | 检查模式，退出码表示状态 |

**支持的语法元素**：
- 语句：变量/常量/函数声明、if/else、while、for、foreach、try/catch、match/case、import/include、global/static、return/break/continue/throw
- 表达式：二元/一元/三元、函数调用、索引/成员访问、lambda/箭头函数、数组/对象字面量、管道运算符、字符串拼接、范围表达式、正则字面量、类型转换

**测试覆盖**：46 个测试用例（语句 18 + 表达式 8 + 注释 6 + 边界 14）

**CLI 用法**：
```bash
jpl fmt script.jpl            # 输出到 stdout
jpl fmt --write script.jpl    # 原地格式化
jpl fmt --check script.jpl    # 检查格式（退出码 0=已格式化 1=需格式化）
jpl fmt *.jpl                  # 批量输出
jpl fmt --write *.jpl          # 批量原地格式化
```

**已知限制**：
- 部分高级语法（如某些数组操作语法）可能解析失败
- 空行不保留（格式化后统一为单换行分隔）
