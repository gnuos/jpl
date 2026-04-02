# JPL 包管理器设计文档

> 本文档定义 `jpl add/remove/install` 包管理器的完整设计方案，指导分阶段实现。

---

## 1. 背景与目标

### 1.1 问题

JPL 现有 `import` 语句支持：
- 相对路径：`import "./lib/utils.jpl"`
- URL 导入：`import "https://example.com/lib.jpl"`
- 裸名：`import "utils"`（在 `jpl_modules/` 中查找）

但缺乏：
- 项目级依赖声明（哪些包是项目需要的）
- 依赖安装/卸载（从外部源获取包）
- 版本锁定（确保不同机器安装相同的版本）
- 传递依赖管理（A 依赖 B，B 依赖 C）

### 1.2 目标

- `jpl add <source>` — 添加依赖到项目
- `jpl remove <name>` — 移除依赖
- `jpl install` — 安装项目全部依赖
- 基于 git 的包源（无需注册中心）
- 传递依赖自动解析
- 版本锁定（commit hash 级别）

---

## 2. Import 解析问题

### 2.1 现有解析逻辑

`engine/module_loader.go` 中 `resolvePath()` 的解析顺序：

```
import "xxx"
  ├─ xxx 以 http:// 或 https:// 开头 → URL 加载
  ├─ xxx 包含 / 或 \ → 相对路径（相对于当前脚本目录）
  └─ xxx 是裸名 → 搜索路径查找：
       1. 当前脚本目录（baseDir）
       2. 额外搜索路径（AddSearchPath）
       3. 向上遍历找到的 jpl_modules/ 目录
       4. ~/.jpl/modules/
       5. ~/.jpl/cache/（仅 URL 缓存）
```

### 2.2 包管理器引入的问题

**问题 1：包安装在哪里？**

`jpl add https://github.com/user/jpl-utils.git` 应该把代码安装到哪里？

方案：安装到 `jpl_modules/<name>/`，与现有搜索路径兼容。

```
my-project/
├── jpl.json
├── jpl_modules/
│   └── utils/              ← jpl add 安装到这里
│       ├── index.jpl
│       └── lib/
│           └── helper.jpl
└── main.jpl                ← import "utils" 找到 jpl_modules/utils/index.jpl
```

**问题 2：导入名称从何而来？**

用户执行 `jpl add https://github.com/user/jpl-utils.git`，导入时应该用什么名字？

规则（按优先级）：
1. `jpl.json` 清单中声明的键名：`{"dependencies": {"my-utils": "..."}}` → `import "my-utils"`
2. git 仓库名推断：`jpl-utils.git` → `jpl-utils`
3. 用户显式指定：`jpl add <url> --name utils`

**问题 3：传递依赖的 import 如何解析？**

```
my-project 依赖 A，A 依赖 B
```

安装后目录结构：
```
my-project/
├── jpl_modules/
│   ├── A/
│   │   ├── index.jpl        ← import "B" 需要找到 B
│   │   └── jpl_modules/     ← ❌ 不安装在这里（扁平化）
│   └── B/
│       └── index.jpl
```

方案：**扁平安装**。所有依赖（包括传递依赖）都安装到项目的 `jpl_modules/` 根目录。

A 的 `import "B"` 通过项目的 `jpl_modules/B/index.jpl` 解析，因为 `resolvePath()` 会向上遍历找到 `jpl_modules/` 目录。

当 A 的源码在 `my-project/jpl_modules/A/` 中执行时，`resolvePath()` 从 `A/` 向上遍历：
```
my-project/jpl_modules/A/  → 没有 jpl_modules/
my-project/jpl_modules/    → 没有 jpl_modules/
my-project/                → 有 jpl_modules/ ✅
```

所以 `import "B"` 会解析到 `my-project/jpl_modules/B/`。

**问题 4：同名冲突**

A 依赖 `lib@v1.0`，B 依赖 `lib@v2.0`，怎么处理？

策略：扁平安装最新版本，记录冲突警告。

```
jpl_modules/
└── lib/       ← 安装 v2.0（最新）
```

如果 A 无法兼容 v2.0，在 `jpl install` 时输出警告。

### 2.3 Import 解析增强（可选）

如果扁平安装无法满足需求，可增强 `FileModuleLoader` 支持包级作用域：

```
my-project/jpl_modules/A/
├── index.jpl
└── jpl_modules/
    └── lib/          ← A 的私有依赖
```

增强 `resolvePath()`：先查找 `<当前模块目录>/jpl_modules/`，再查找项目根的 `jpl_modules/`。

但 Phase A 不实现此功能，扁平安装足以覆盖大部分场景。

---

## 3. 清单文件 `jpl.json`

### 3.1 格式

```json
{
    "name": "my-project",
    "version": "0.1.0",
    "description": "My JPL project",
    "dependencies": {
        "utils": "https://github.com/user/jpl-utils.git",
        "http-client": "https://github.com/user/jpl-http.git@v1.2.0"
    }
}
```

### 3.2 字段说明

| 字段 | 必填 | 说明 |
|------|------|------|
| `name` | 是 | 项目名称（标识符） |
| `version` | 否 | 项目版本（语义化版本） |
| `description` | 否 | 项目描述 |
| `dependencies` | 否 | 依赖映射 `{导入名: 源地址}` |

### 3.3 依赖源地址格式

| 格式 | 示例 | 说明 |
|------|------|------|
| git URL | `https://github.com/user/lib.git` | 最新 commit |
| git URL + tag | `https://github.com/user/lib.git@v1.0.0` | 指定 tag |
| git URL + branch | `https://github.com/user/lib.git#main` | 指定分支 |
| 本地路径 | `../shared-lib` | 本地相对路径（开发用） |

### 3.4 设计决策

**为什么用 JSON 而非 YAML？**
- JPL 内置 `json_encode`/`json_decode`，工具链可读写
- JSON 更简单，无缩进陷阱
- 与 npm/cargo 等主流包管理器一致

**为什么依赖用 map 而非 array？**
- 键名即导入名，无需额外推断
- 声明式，一目了然

---

## 4. 锁文件 `jpl.lock.yaml`

### 4.1 现有锁文件

`engine/lockfile.go` 已实现锁文件基础设施：
- `LockFile` 结构体（version, generated, remote map）
- `LockEntry`（hash, downloaded, size）
- SHA256 校验
- YAML 序列化

### 4.2 扩展格式

```yaml
version: 1
generated: "2026-04-01T12:00:00Z"
packages:
  utils:
    source: "https://github.com/user/jpl-utils.git"
    resolved: "https://github.com/user/jpl-utils.git"
    version: "v1.2.3"
    commit: "abc123def456"
    hash: "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
    dependencies: []
  http-client:
    source: "https://github.com/user/jpl-http.git@v1.2.0"
    resolved: "https://github.com/user/jpl-http.git"
    version: "v1.2.0"
    commit: "789abc123def"
    hash: "sha256:..."
    dependencies:
      - utils
```

### 4.3 锁文件与现有 LockFile 的关系

现有 `LockFile` 只锁 URL 源（用于 URL import）。包管理器的锁文件锁定 git 依赖。

方案：复用 YAML 格式，但扩展结构。现有 URL 锁定逻辑保持不变（存于 `remote` 字段）。

```yaml
version: 1
generated: "2026-04-01T12:00:00Z"
# 现有：URL import 缓存
remote:
  https://example.com/lib.jpl:
    hash: "sha256:..."
# 新增：包管理器依赖
packages:
  utils:
    source: "..."
    commit: "..."
    hash: "sha256:..."
```

---

## 5. CLI 命令设计

### 5.1 `jpl add`

```
jpl add <source>
jpl add <source> --name <import-name>
jpl add <source> --version <tag>
```

行为：
1. 解析源地址（git URL + 可选 @tag）
2. 克隆仓库到临时目录
3. 如果有 `@tag`，checkout 到该 tag
4. 读取仓库根目录的 `jpl.json`（如果存在）获取元信息
5. 确定导入名（--name > jpl.json.name > 仓库名推断）
6. 安装到 `jpl_modules/<name>/`
7. 递归安装其 `jpl.json` 中的依赖（传递依赖）
8. 更新项目 `jpl.json`（添加依赖条目）
9. 更新 `jpl.lock.yaml`（锁定 commit hash）

示例：
```bash
$ jpl add https://github.com/user/jpl-utils.git
Added dependency: utils (https://github.com/user/jpl-utils.git @ abc123d)
Installed to: jpl_modules/utils/

$ jpl add https://github.com/user/jpl-http.git@v1.2.0 --name http
Added dependency: http (https://github.com/user/jpl-http.git @ v1.2.0)
Installed to: jpl_modules/http/
```

### 5.2 `jpl remove`

```
jpl remove <name>
```

行为：
1. 检查 `jpl.json` 中是否存在该依赖
2. 删除 `jpl_modules/<name>/` 目录
3. 从 `jpl.json` 中移除条目
4. 从 `jpl.lock.yaml` 中移除条目
5. 检查是否有其他依赖依赖于被移除的包（警告）

示例：
```bash
$ jpl remove utils
Removed dependency: utils
Deleted: jpl_modules/utils/
```

### 5.3 `jpl install`

```
jpl install
```

行为：
1. 读取 `jpl.json` 的 `dependencies`
2. 读取 `jpl.lock.yaml` 的 `packages`（如果存在）
3. 对每个依赖：
   a. 如果 lock 文件有条目，按 lock 文件的 commit 安装
   b. 如果 lock 文件无条目，安装最新版本
4. 递归处理传递依赖
5. 更新 `jpl.lock.yaml`
6. 检测并报告版本冲突

示例：
```bash
$ jpl install
Installing dependencies...
  utils (https://github.com/user/jpl-utils.git @ abc123d) - OK
  http (https://github.com/user/jpl-http.git @ 789abc1) - OK
  (transitive) net (https://github.com/user/jpl-net.git @ def456a) - OK
3 packages installed.
```

### 5.4 `jpl list`

```
jpl list
```

行为：列出项目已安装的依赖。

```bash
$ jpl list
my-project@0.1.0
├── utils@1.2.3 (https://github.com/user/jpl-utils.git)
├── http@1.2.0 (https://github.com/user/jpl-http.git)
│   └── net@0.5.0 (https://github.com/user/jpl-net.git) [transitive]
└── (3 packages)
```

---

## 6. 依赖解析算法

### 6.1 扁平安装策略

```
输入：项目的 jpl.json
输出：所有依赖（含传递依赖）的安装列表

1. 读取项目的 jpl.json dependencies
2. 初始化 resolved = {}（已解析的包映射）
3. 初始化 queue = dependencies 列表
4. 循环 queue 不为空：
   a. 取出一个 (name, source)
   b. 如果 name 已在 resolved 中：
      - 检查版本是否冲突
      - 如果冲突，记录警告，使用已安装版本
      - 跳过
   c. 克隆/更新源码到临时位置
   d. 读取该包的 jpl.json（如果存在）
   e. 将其 dependencies 加入 queue
   f. resolved[name] = {source, commit, hash}
5. 返回 resolved
```

### 6.2 循环依赖检测

```
A 依赖 B，B 依赖 A → 循环依赖

检测方法：DFS + 三色标记
- 白色（未访问）
- 灰色（正在访问的祖先节点）
- 黑色（已完成访问）

如果在 DFS 中遇到灰色节点 → 检测到循环依赖
```

### 6.3 版本冲突处理

Phase A（无版本约束）：
- 所有包使用最新版本
- 如果传递依赖之间版本不同，输出警告
- 使用最新版本安装

Phase B（有版本约束）：
- 解析 semver 约束
- 如果无法满足所有约束 → 报错
- 如果可以满足 → 安装满足的版本

---

## 7. 目录结构

### 7.1 项目目录

```
my-project/
├── jpl.json                ← 项目清单
├── jpl.lock.yaml           ← 锁文件（自动生成）
├── jpl_modules/            ← 依赖安装目录
│   ├── utils/
│   │   ├── jpl.json        ← 包自身的清单
│   │   ├── index.jpl       ← 入口文件
│   │   └── lib/
│   │       └── helper.jpl
│   └── http/
│       ├── jpl.json
│       └── index.jpl
├── src/
│   └── main.jpl            ← 用户代码
└── ...
```

### 7.2 全局缓存

```
~/.jpl/
├── cache/                  ← URL import 缓存（现有）
│   └── <sha256>.jpl
├── packages/               ← 包管理器全局缓存
│   └── <owner>/
│       └── <repo>/
│           └── <commit>/
│               └── <files>
└── repl_history            ← REPL 历史（现有）
```

全局缓存避免重复克隆同一仓库。`jpl add` 时：
1. 检查 `~/.jpl/packages/` 是否有该 commit
2. 有 → 直接复制到 `jpl_modules/`
3. 无 → 克隆仓库，存入缓存，再复制

---

## 8. 实现阶段

### Phase A：最小可用版本（1 周）

| 任务 | 文件 | 说明 |
|------|------|------|
| `jpl.json` 读写 | `pkg/pm/manifest.go` | 清单文件解析和生成 |
| `jpl add` | `cmd/jpl/add.go` | git 克隆 + 安装到 jpl_modules/ |
| `jpl remove` | `cmd/jpl/remove.go` | 删除目录 + 清理清单 |
| `jpl install` | `cmd/jpl/install.go` | 读取清单，安装全部依赖 |
| git 操作 | `pkg/pm/git.go` | clone, checkout, 获取 commit hash |
| 测试 | `pkg/pm/manifest_test.go` | 清单解析测试 |

Phase A 限制：
- 无版本约束（总是最新）
- 无传递依赖
- 无全局缓存
- 无锁文件更新

### Phase B：依赖解析（1 周）

| 任务 | 文件 | 说明 |
|------|------|------|
| 传递依赖 | `pkg/pm/resolver.go` | DFS 解析传递依赖 |
| 循环检测 | `pkg/pm/resolver.go` | 三色标记检测 |
| 锁文件 | 复用 `engine/lockfile.go` | 更新 commit hash |
| `jpl list` | `cmd/jpl/list.go` | 列出依赖树 |
| 全局缓存 | `pkg/pm/cache.go` | ~/.jpl/packages/ 缓存 |
| 测试 | `pkg/pm/resolver_test.go` | 解析算法测试 |

### Phase C：版本约束（按需）

| 任务 | 文件 | 说明 |
|------|------|------|
| semver 解析 | `pkg/pm/semver.go` | 语义化版本 |
| 约束求解 | `pkg/pm/resolver.go` | ^, >=, ~ 约束 |
| 版本冲突 | `pkg/pm/resolver.go` | 冲突检测和报错 |

---

## 9. 边界情况

| 场景 | 处理方式 |
|------|----------|
| 无 `jpl.json` 的项目 | `jpl add` 自动创建 |
| `jpl_modules/` 不存在 | 自动创建 |
| git clone 失败 | 报错，不修改清单 |
| 依赖的 `jpl.json` 不存在 | 使用仓库名作为包名，无传递依赖 |
| 同名不同源的依赖 | 报错：name 冲突 |
| 循环依赖 | 报错并显示循环路径 |
| 网络离线 | 如果有缓存，使用缓存；否则报错 |
| 删除被其他包依赖的包 | 输出警告，但仍删除 |
| `jpl_modules/` 手动删除 | `jpl install` 可重新安装 |

---

## 10. 实现约束

### 10.1 必须遵守的规则

- **3-File Rule**：单次任务修改不超过 3 个文件
- **Describe Before Code**：每阶段开始前描述方案
- **Test-First**：先写测试再实现
- **渐进式**：Phase A 完成后验证，再进入 Phase B

### 10.2 代码规范

- 新包放在 `pkg/pm/` 目录
- CLI 子命令放在 `cmd/jpl/` 目录
- 复用现有 `engine/lockfile.go` 的锁文件基础设施
- 复用现有 `FileModuleLoader` 的搜索路径逻辑（不修改）

### 10.3 验证方法

每个阶段完成后：
1. `go test ./pkg/pm/...` — 单元测试
2. `go test ./cmd/jpl/...` — CLI 集成测试
3. 手动测试完整流程：
   ```bash
   mkdir test-project && cd test-project
   jpl add https://github.com/user/some-jpl-lib.git
   jpl list
   jpl remove some-jpl-lib
   jpl install
   ```

---

## 11. 与现有系统的兼容性

### 11.1 不需要修改的部分

| 组件 | 原因 |
|------|------|
| `engine/module_loader.go` | `jpl_modules/` 已在搜索路径中 |
| `parser/parser.go` | `import` 语句解析不变 |
| `engine/vm.go` | `opImport` 逻辑不变 |
| `buildin/` | 内置函数不受影响 |

### 11.2 可能需要修改的部分

| 组件 | 修改原因 |
|------|----------|
| `engine/lockfile.go` | 扩展锁文件格式（packages 字段） |
| `cmd/jpl/root.go` | 注册新子命令 |

### 11.3 Import 语句无需修改

核心结论：**`import` 语句的查找逻辑不需要任何修改**。

原因：
1. 包安装到 `jpl_modules/<name>/`
2. `FileModuleLoader.searchForModule()` 已搜索 `jpl_modules/`
3. 裸名 `import "utils"` 自动解析到 `jpl_modules/utils/index.jpl`
4. 传递依赖通过扁平安装到同一 `jpl_modules/` 解决

包管理器只负责**安装和管理** `jpl_modules/` 目录中的文件，不修改 import 解析逻辑。
