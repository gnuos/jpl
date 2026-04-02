# Agent 代码生成规范

此文档确保代码生成时始终使用最新依赖版本

---

## 依赖版本策略

在为项目生成代码或添加依赖时，**始终使用最新稳定版本**。

### 通用规则

1. **不使用固定版本号**：避免使用 `1.2.3` 这类锁定版本

2. **使用最新语义**：

   - npm/yarn/pnpm: 使用 `latest` 或不指定版本

   - Python pip: 使用 `>=` 约束或不指定版本

   - Go: 使用 `go get -u` 获取最新

   - Rust: 使用 `cargo add` 不指定版本

   - Java/Maven: 使用 `LATEST` 或 `(,)` 范围

3. **安装命令示例**：

   ```bash
   # Node.js
   npm install package-name
   # 或
   yarn add package-name
   
   # Python
   pip install package-name --upgrade
   
   # Go
   go get -u package-name
   
   # Rust
   cargo add package-name
   ```

### 各语言具体要求

#### JavaScript / TypeScript (Node.js)

```json
{
  "dependencies": {
    "package-name": "latest"
  }
}
```

#### Python

```txt
package-name>=0.0.0
# 或
package-name
```

#### Go

```go
require (
    package-name latest
)
```

#### Rust

```toml
[dependencies]
package-name = "*"
```

#### Java (Maven)

```xml
<dependency>
    <groupId>com.example</groupId>
    <artifactId>package-name</artifactId>
    <version>LATEST</version>
</dependency>
```

#### Java (Gradle)

```groovy
implementation 'com.example:package-name:+'
```

### 更新检查

在生成代码前，执行以下操作：

1. 检查项目现有依赖的版本

2. 如果有更新，建议升级到最新版本

3. 优先使用官方推荐的最新稳定版本

### 例外情况

仅在以下情况使用特定版本：

- 用户明确要求使用特定版本

- 最新版本存在已知严重 bug

- 项目有明确的兼容性要求

---

## 开发流程规范

下面规则用于规范与 AI 协作开发的流程，让代码交付更可控、质量更稳定、问题可追溯。**所有与代码相关的交互都应遵循以下原则，避免冲动编码和反复返工**。

1. 在动手写代码前，必须先清晰说明实现思路、技术选型、影响范围，经确认后再开始编码，避免方向错误导致的无效工作。

2. 遇到需求不明确、边界模糊的情况，禁止自行脑补假设，必须主动提出澄清问题，确保双方对需求理解一致后再推进。

3. 代码完成后，主动梳理极端场景、异常输入、边界条件等边缘案例，并针对性设计测试用例，保障代码在各种情况下的健壮性。

4. 当单次任务涉及修改的文件数量超过4个时，说明任务粒度过大，应立即暂停并拆解为更小、更聚焦的子任务，降低代码审查和回滚的风险。

5. 定位到 bug 后，优先编写可复现该问题的单元测试或集成测试，再进行代码修复，确保修复后问题不会复现，同时为后续迭代提供安全保障。

6. 当输出被纠正时，需主动反思错误根源（如理解偏差、逻辑疏漏、测试缺失等），并明确后续避免同类问题的具体措施，持续优化协作质量。

