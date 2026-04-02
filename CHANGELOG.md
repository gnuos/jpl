# Changelog

## v0.9.0 (2026-04-02)

> 项目进入维护模式，核心功能完整。

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
