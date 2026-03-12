## 9. Executor能力扩展路线图（v4.7）

> 基于v4.6 YAML语法规范和Meta自发现机制，定义executor从当前状态到完全体的扩展路线。

### 9.1 P0 — Command机制（一个executor多个能力）

#### 核心概念

当前一个executor只有一个功能（如http只能发请求），但真实场景中一个executor应该支持多个command，每个command有独立的input/output schema。

**业界对标：**
| 平台 | 术语 |
|------|------|
| 飞书Anycross | Connector + Action |
| MCP | Server + Tool |
| Zapier | App + Action/Trigger |
| n8n | Node + Operation |

#### YAML语法

```yaml
# 同一个executor，不同command
- id: call_api
  executor: http
  command: request              # ← 指定用哪个能力
  input:
    url: "https://api.example.com/v1/data"
    method: POST
    body: { key: "value" }

- id: wait_hook
  executor: http
  command: webhook              # ← 同一个executor，不同能力
  input:
    path: "/callbacks/order"
    timeout: 1h

# Agent executor 的不同 command
- id: write_prd
  executor: agent
  command: chat                 # 对话模式
  input:
    prompt: "写PRD..."
    model: opus

- id: review_code
  executor: agent
  command: review               # 代码审查模式（只读）
  input:
    repo: "github.com/xxx/yyy"
    pr: 42

# Shell executor 的不同 command
- id: run_cmd
  executor: shell
  command: exec                 # 执行命令
  input:
    command: "make build"

- id: run_script
  executor: shell
  command: script               # 执行脚本文件
  input:
    path: "./deploy.sh"
    args: ["prod"]
```

#### Go接口设计

```go
// CommandDef 定义一个command的元数据和schema
type CommandDef struct {
    Name        string      `json:"name"`        // "request", "webhook"
    Label       string      `json:"label"`       // "HTTP Request", "Webhook Wait"
    Description string      `json:"description"`
    InputType   any         `json:"inputType"`   // Go struct → JSON Schema
    OutputType  any         `json:"outputType"`  // Go struct → JSON Schema
}

// Executor接口升级
type Executor interface {
    Type() string                                    // "http", "shell", "agent"
    Meta() ExecutorMeta                              // 元数据（icon, color等）
    Commands() []CommandDef                          // 列出所有支持的command
    Execute(ctx context.Context, rc *RunContext) error // 执行（rc里有Command字段）
}

// RunContext 新增 Command 字段
type RunContext struct {
    // ... 现有字段
    Command string              // 要执行的command（从YAML的command字段来）
}
```

#### 各executor的command设计

| Executor | Commands | 说明 |
|----------|----------|------|
| **agent** | `chat`(默认), `review`, `structured_output` | 对话/代码审查/结构化输出 |
| **human** | `form`(默认), `approval`, `notification` | 表单提交/审批投票/仅通知 |
| **http** | `request`(默认), `webhook`, `graphql` | HTTP请求/等待回调/GraphQL |
| **shell** | `exec`(默认), `script`, `docker` | 命令/脚本文件/Docker容器 |
| **subprocess** | `run`(默认) | 启动子流程 |
| **calculator** | `eval`(默认) | 计算表达式 |

**默认command规则：** 如果YAML没写command字段，使用executor的第一个command（默认command）。向后兼容。

### 9.2 P0 — 三层Schema实现

#### 核心概念

每个command定义3层Schema：
- **InputType** — 该command接受什么输入（YAML的input字段校验）
- **OutputType** — 该command返回什么标准信封（executor级，固定结构）
- **TraceType** — 该command执行过程中产生什么trace数据（entry types + content结构）

#### Go struct即Schema

```go
// 以HTTP executor的request command为例
type HTTPRequestInput struct {
    Method  string            `json:"method" jsonschema:"enum=GET,POST,PUT,DELETE,PATCH,required"`
    URL     string            `json:"url" jsonschema:"required"`
    Headers map[string]string `json:"headers,omitempty"`
    Body    any               `json:"body,omitempty"`
    Query   map[string]string `json:"query,omitempty"`
    Timeout string            `json:"timeout,omitempty" jsonschema:"default=30s"`
}

type HTTPRequestOutput struct {
    StatusCode int               `json:"status_code"`
    Headers    map[string]string `json:"headers"`
    Duration   string            `json:"duration"`
}

// 用 github.com/invopop/jsonschema 自动生成JSON Schema
// schema := jsonschema.Reflect(&HTTPRequestInput{})
```

#### Publish验证链路

```
YAML发布 → 遍历每个step → 找到executor+command → 取InputType
→ 验证step.input是否符合InputType的JSON Schema
→ 验证{{steps.X.data.Y}}引用：X的OutputType是否有字段Y
→ 通过才允许发布
```

### 9.3 P0 — 两层Output

#### 核心概念

executor输出 = **标准信封**（executor级，固定） + **业务Data**（step级，按data_schema）

```json
// tasks表存的output（标准信封，几KB）
{
    "status": "success",
    "duration": "2.3s",
    "status_code": 200,          // HTTP executor特有
    "token_usage": { ... },      // Agent executor特有
    "data": {                    // 业务数据（按YAML的data_schema）
        "prd_content": "...",
        "features": [...]
    }
}
```

**引擎职责分离：**
- 引擎验证标准信封（OutputType()定义的字段）
- executor验证业务data（data_schema定义的字段）
- 跨步骤引用：`{{steps.X.duration}}` = 信封字段，`{{steps.X.data.prd_content}}` = 业务字段

### 9.4 P1 — Trace数据层

#### executor_trace表

```sql
CREATE TABLE executor_trace (
    instance_id TEXT NOT NULL,
    step_id     TEXT NOT NULL,
    seq         INTEGER NOT NULL,      -- 自增序号
    entry_type  TEXT NOT NULL,          -- "system_prompt", "message", "tool_call", "request", "response"
    content     TEXT NOT NULL,          -- 完整JSON（每条entry是独立完整的事件）
    summary     TEXT DEFAULT '',        -- 搜索用摘要
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (instance_id, step_id, seq)
);
CREATE INDEX idx_trace_search ON executor_trace(instance_id, step_id, summary);
```

#### Storage接口

```go
type Storage interface {
    Append(ctx context.Context, instanceID, stepID, entryType string, content any, summary string) error
    List(ctx context.Context, instanceID, stepID string, opts ListOpts) ([]TraceEntry, error)
    Search(ctx context.Context, instanceID, stepID, keyword string) ([]TraceEntry, error)
}

type ListOpts struct {
    Offset    int
    Limit     int
    EntryType string   // 按类型过滤
    Since     int      // 从某个seq开始（实时轮询用）
}
```

#### 各executor的trace entry types

| Executor | Entry Types |
|----------|------------|
| **agent** | `system_prompt`, `message`, `tool_call`, `token_snapshot` |
| **human** | `notification`, `form_submit`, `vote`, `reminder` |
| **http** | `request`, `response`, `retry` |
| **shell** | `start`, `stdout_chunk`, `stderr_chunk`, `exit` |
| **subprocess** | `start`, `step_completed`, `step_failed`, `finish` |
| **calculator** | `eval` |

#### API

```
GET /api/instances/:id/steps/:step_id/trace?offset=0&limit=50           # 分页
GET /api/instances/:id/steps/:step_id/trace?entry_type=tool_call        # 按类型
GET /api/instances/:id/steps/:step_id/trace?since=42                    # 实时轮询
GET /api/instances/:id/steps/:step_id/trace/search?q=error              # 搜索
```

### 9.5 P1 — 凭证注入

#### YAML语法

```yaml
- id: call_github
  executor: http
  command: request
  input:
    url: "https://api.github.com/repos/xxx"
    headers:
      Authorization: "Bearer {{credentials.github_token}}"
```

#### 实现方案

```go
type CredentialStore interface {
    Get(ctx context.Context, name string) (string, error)
    List(ctx context.Context) ([]CredentialMeta, error)
}
```

凭证存在ACP的credentials表（已有），executor通过Runtime SDK访问。YAML变量替换时，`{{credentials.xxx}}` 从CredentialStore获取。

### 9.6 P1 — Publish验证

#### 验证规则

1. **executor/control互斥** — 一个step不能同时有executor和control
2. **command存在性** — executor+command组合必须在registry中存在
3. **input字段验证** — input字段符合该command的InputType JSON Schema
4. **跨步骤类型检查** — `{{steps.X.data.Y}}` 中，X的output schema必须包含字段Y
5. **依赖图无环** — depends_on不能形成循环
6. **变量引用可达** — 引用的步骤必须在依赖链上游

### 9.7 P2 — Flow Schema API + Monaco自动校验

#### API

```
GET /api/flow/schema → 返回完整的JSON Schema
```

后端自动合并：
- 流程通用字段（name, description, input, steps）
- 所有executor的InputType（按command分组）
- 所有control的字段定义

#### 前端

```typescript
// Monaco编辑器加载Flow Schema
import { configureMonacoYaml } from 'monaco-yaml';
const schema = await fetch('/api/flow/schema').then(r => r.json());
configureMonacoYaml(monaco, { schemas: [{ uri: 'flow', fileMatch: ['*'], schema }] });
```

用户在YAML编辑器里写流程时，自动补全executor名称、command名称、input字段，实时校验。

### 9.8 实施顺序

| 阶段 | 内容 | 优先级 | 预估工作量 |
|------|------|--------|-----------|
| **Phase A** | Command机制 — Executor接口+RunContext+YAML解析+默认command兼容 | P0 | 1天 |
| **Phase B** | 三层Schema — InputType/OutputType/TraceType Go struct + jsonschema反射 | P0 | 1天 |
| **Phase C** | 两层Output — 标准信封+业务Data分离，引擎验证逻辑 | P0 | 0.5天 |
| **Phase D** | Trace数据层 — executor_trace表+Storage接口+API | P1 | 1天 |
| **Phase E** | 凭证注入 — CredentialStore+YAML变量替换 | P1 | 0.5天 |
| **Phase F** | Publish验证 — Schema校验+跨步骤类型检查 | P1 | 1天 |
| **Phase G** | Flow Schema API + Monaco自动校验 | P2 | 1天 |

**总计约6天。Phase A→B→C是基础，必须按顺序；D/E/F可以并行；G最后。**

