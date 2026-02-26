# ACP 统一资源调度平台 PRD

> 版本：v1.0 | 作者：Lyra | 日期：2026-02-25
> 状态：Draft | 审批：待泽斌确认

## 1. 背景与动机

### 1.1 现状

OpenClaw 的架构分三层：

```
Agent（智能层）→ Gateway（管控层）→ Node（执行层）
```

**Gateway 是单机视角**——每个 Gateway 只管自己注册的 Agent、Node、Session。跨 Gateway 的资源不可见、不可达。

ACP 已经实现了**多 Gateway 注册与管理**（`GatewayRegistryService`），可以：
- 注册多个 Gateway（DB 存储 URL + Token）
- 按 gatewayId 路由 RPC 调用到目标 Gateway
- 跨 Gateway 聚合 Agent 列表、Node 列表
- 同步远程 Gateway 的 Agent 到本地 DB

**但 Agent 本身无法跨 Gateway 操作。** Lyra（在 Gateway A 上）想在 Catherine-Build（在 Gateway B 上的 Node）执行命令，目前只能因为 Catherine-Build 恰好也 pair 到了 Gateway A。如果 Node 只连在 Gateway B 上，Lyra 无法触达。

### 1.2 核心问题

| 问题 | 影响 |
|---|---|
| Agent 只能操作本 Gateway 的 Node | 跨机器协作需要 Node 双注册（workaround） |
| Agent 只能与本 Gateway 的 Agent 通信 | 跨 Gateway 的 Agent 协作需要 ACP 流程间接实现 |
| Gateway UI 和 ACP UI 功能重叠 | 用户需要切换两个界面 |
| Gateway 配置需手动编辑 JSON | 运维效率低，易出错 |

### 1.3 目标

**将 ACP 从"管理平台"升级为"统一资源调度平台"**——Agent 通过 ACP 提供的工具，透明访问任意 Gateway 上的任何资源（Node、Agent、Session），无需关心资源的物理位置。

同时，ACP 作为**唯一管理入口**，整合 Gateway UI 的核心功能，用户无需再访问 Gateway Web 界面。

## 2. 架构设计

### 2.1 整体架构

```
┌─────────────────────────────────────────────────────┐
│                    ACP 统一平台                       │
│                                                      │
│  ┌──────────┐  ┌──────────┐  ┌──────────────────┐   │
│  │ 管理 UI  │  │ 流程引擎  │  │ 统一资源调度总线  │   │
│  │(Gateway  │  │(Process/ │  │ (Resource Bus)   │   │
│  │ 整合)    │  │ Task)    │  │                  │   │
│  └────┬─────┘  └────┬─────┘  └────────┬─────────┘   │
│       │              │                 │              │
│  ┌────┴──────────────┴─────────────────┴──────────┐  │
│  │          Gateway Registry Service               │  │
│  │    ┌──────────┐  ┌──────────┐  ┌──────────┐    │  │
│  │    │Gateway A │  │Gateway B │  │Gateway C │    │  │
│  │    │(Lyra)    │  │(Catherine│  │(Future)  │    │  │
│  │    └────┬─────┘  └────┬─────┘  └────┬─────┘   │  │
│  └─────────┼─────────────┼──────────────┼─────────┘  │
└────────────┼─────────────┼──────────────┼────────────┘
             │             │              │
    ┌────────┴───┐  ┌──────┴────┐  ┌──────┴────┐
    │Local-Build │  │Catherine- │  │Future-    │
    │(Node)      │  │Build(Node)│  │Node       │
    └────────────┘  └───────────┘  └───────────┘
```

### 2.2 统一资源调度总线（Resource Bus）

**核心概念：ACP 维护一个全局资源注册表，Agent 通过 ACP 工具访问资源，ACP 自动路由到正确的 Gateway。**

#### 2.2.1 全局资源注册表

```
┌─────────────────────────────────────┐
│         Global Resource Registry     │
├─────────────────────────────────────┤
│ Nodes:                               │
│   Local-Build     → Gateway A        │
│   Catherine-Build → Gateway A (*)    │
│   Future-Node     → Gateway C        │
│                                      │
│ Agents:                              │
│   lyra            → Gateway A        │
│   catherine       → Gateway B        │
│                                      │
│ Sessions:                            │
│   agent:lyra:main → Gateway A        │
│   agent:catherine:main → Gateway B   │
│                                      │
│ (*) 标注：同一Node可能被多个Gateway   │
│     看到，Registry去重并记录primary   │
└─────────────────────────────────────┘
```

资源注册表**实时构建**（不持久化），每次查询时从各 Gateway RPC 聚合：
- `nodes.status` → 聚合所有 Node
- `agents.list` → 聚合所有 Agent
- `sessions.list` → 聚合所有 Session

可加缓存（TTL 30s）避免频繁 RPC。

#### 2.2.2 路由策略

当 Agent 调用 ACP 工具时：

```
1. Agent 调用 acp_node_run(node="Catherine-Build", command=["go","version"])
2. ACP Resource Bus 查找 "Catherine-Build" 所在的 Gateway
3. 找到 → Gateway B
4. 通过 GatewayRegistryService.GetClient("gateway-b-id") 获取连接
5. 调用 client.NodeInvoke(nodeID, ...) 或 client.Call("nodes.run", ...)
6. 返回结果给 Agent
```

**Agent 无需知道 Node 在哪个 Gateway 上——完全透明。**

### 2.3 ACP 工具集（Plugin Tools）

ACP 以 OpenClaw Plugin 形式注册工具，Agent 的 system prompt 中自动可见。

#### 2.3.1 跨 Gateway Node 操作

| 工具名 | 功能 | 对应 Gateway RPC |
|---|---|---|
| `acp_node_list` | 列出所有 Gateway 的 Node（全局视图） | `nodes.status` × N |
| `acp_node_run` | 在指定 Node 上执行命令 | `nodes.run` → 路由到目标 Gateway |
| `acp_node_invoke` | 在指定 Node 上调用 invoke | `node.invoke` → 路由 |
| `acp_node_describe` | 获取 Node 详细信息 | `node.describe` → 路由 |
| `acp_node_browse` | 在指定 Node 上操作浏览器 | `browser.*` → 路由 |

**工具定义示例：**

```json
{
  "name": "acp_node_run",
  "description": "在任意 Node 上执行命令。自动路由到 Node 所在的 Gateway。",
  "parameters": {
    "node": { "type": "string", "description": "Node 名称或 ID", "required": true },
    "command": { "type": "array", "items": { "type": "string" }, "description": "命令及参数", "required": true },
    "cwd": { "type": "string", "description": "工作目录（可选）" },
    "timeoutMs": { "type": "number", "description": "超时毫秒（默认 30000）" }
  }
}
```

#### 2.3.2 跨 Gateway Agent 操作

| 工具名 | 功能 | 对应 Gateway RPC |
|---|---|---|
| `acp_agent_list` | 列出所有 Gateway 的 Agent | `agents.list` × N |
| `acp_agent_send` | 向任意 Agent 发消息 | `sessions.send` → 路由 |
| `acp_agent_spawn` | 在指定 Gateway 上启动子 Agent | `sessions.spawn` → 路由 |
| `acp_agent_history` | 获取任意 Agent 的对话历史 | `chat.history` → 路由 |
| `acp_agent_status` | 获取任意 Agent 的状态 | `session.status` → 路由 |

#### 2.3.3 跨 Gateway 管理操作

| 工具名 | 功能 | 对应 Gateway RPC |
|---|---|---|
| `acp_gateway_list` | 列出所有 Gateway 及状态 | DB + connectivity check |
| `acp_gateway_config` | 查看/修改 Gateway 配置 | `config.get` / `config.patch` → 路由 |
| `acp_cron_list` | 列出所有 Gateway 的 Cron 任务 | `cron.list` × N |
| `acp_cron_manage` | 创建/更新/删除 Cron 任务 | `cron.*` → 路由 |

### 2.4 ACP 管理 UI 整合

**目标：用户只需访问 ACP 界面，不再需要 Gateway Web UI。**

#### 2.4.1 功能对照表

| Gateway UI 页面 | 功能 | ACP 整合方案 | 优先级 |
|---|---|---|---|
| `/overview` | 总览仪表盘 | ✅ 已有 Dashboard，增强为全局视图 | P0 |
| `/agents` | Agent CRUD | ✅ 已有，补充配置编辑能力 | P0 |
| `/chat` | Agent 对话 | ✅ 已有 | Done |
| `/sessions` | Session 列表 | ⚠️ 整合到 Agent 详情页 | P1 |
| `/nodes` | Node 管理 | ✅ 已有 | Done |
| `/config` | Gateway 配置 | 🆕 新增：可视化配置编辑器 | P0 |
| `/usage` | 用量/费用统计 | 🆕 新增：跨 Gateway 费用汇总 | P0 |
| `/cron` | 定时任务管理 | 🆕 新增：可视化 Cron 管理 | P1 |
| `/skills` | Skill 管理 | 🆕 新增：Skill 安装/更新/市场 | P1 |
| `/channels` | 渠道管理 | 🆕 新增：渠道状态/配置 | P2 |
| `/logs` | 系统日志 | 🆕 新增：日志查看 | P2 |
| `/debug` | 调试工具 | 可选，低优先级 | P3 |

## 3. 实现方案

### 3.1 Phase 1：资源调度总线核心（2-3 天）

**目标：Agent 可以通过 ACP 工具跨 Gateway 操作 Node。**

#### 3.1.1 后端：ResourceBusService

新增 `internal/acp/service/resource_bus_service.go`：

```go
type ResourceBusService struct {
    GatewayRegistry *GatewayRegistryService
    nodeCache       sync.Map // nodeName -> nodeLocation
    cacheTTL        time.Duration
}

type NodeLocation struct {
    GatewayID string
    NodeID    string
    CachedAt  time.Time
}

// ResolveNode 查找 Node 所在的 Gateway
func (s *ResourceBusService) ResolveNode(nameOrID string) (*NodeLocation, error)

// ExecOnNode 在任意 Node 上执行命令（自动路由）
func (s *ResourceBusService) ExecOnNode(node string, command []string, cwd string, timeoutMs int) (*ExecResult, error)

// InvokeOnNode 在任意 Node 上调用 invoke
func (s *ResourceBusService) InvokeOnNode(node string, command string, params map[string]interface{}, timeoutMs int) (*NodeInvokeResult, error)

// ListAllNodes 聚合所有 Gateway 的 Node
func (s *ResourceBusService) ListAllNodes() ([]GlobalNodeInfo, error)
```

核心逻辑：
1. `ResolveNode`：遍历所有 Gateway 调 `nodes.status`，找到匹配的 Node，缓存结果
2. `ExecOnNode`：先 Resolve，再路由到目标 Gateway 调 `nodes.run`
3. `ListAllNodes`：聚合所有 Gateway 的 Node 列表，附加 `gatewayId` 和 `gatewayName` 字段

#### 3.1.2 后端：API Endpoints

```
GET    /api/resource-bus/nodes              # 全局 Node 列表
POST   /api/resource-bus/nodes/:name/exec   # 跨 Gateway 执行命令
POST   /api/resource-bus/nodes/:name/invoke # 跨 Gateway invoke
GET    /api/resource-bus/nodes/:name/describe # 跨 Gateway describe
```

#### 3.1.3 Tool 注册：原生 Plugin 机制

**决策：使用 OpenClaw 原生 Plugin 机制，不引入 MCP。**

理由：
- ACP 现有的 `acp_*` 工具（`acp_create_task`、`acp_list_tasks` 等）已经通过此机制工作，**已验证**
- Agent 天然会使用 Plugin 工具，工具定义直接注入 system prompt
- 链路最短：Agent → Plugin HTTP call → ACP API → Gateway RPC
- 无额外协议层，故障面最小

新增的 Resource Bus 工具与现有 ACP 工具注册方式完全一致，ACP 启动时向 Gateway 注册：

```go
// ACP Plugin 注册的工具列表（追加到现有工具之后）
tools := []PluginTool{
    // 跨 Gateway Node 操作
    {Name: "acp_node_run", Description: "在任意Node上执行命令，自动路由到Node所在的Gateway", ...},
    {Name: "acp_node_list", Description: "列出所有Gateway的Node（全局视图）", ...},
    {Name: "acp_node_invoke", Description: "在指定Node上调用invoke命令", ...},
    {Name: "acp_node_describe", Description: "获取Node详细信息", ...},
    // 跨 Gateway Agent 操作
    {Name: "acp_agent_send", Description: "向任意Agent发消息（跨Gateway）", ...},
    {Name: "acp_agent_spawn", Description: "在指定Gateway上启动子Agent", ...},
    // 管理操作
    {Name: "acp_gateway_config", Description: "查看/修改Gateway配置", ...},
    {Name: "acp_cron_manage", Description: "管理Cron定时任务", ...},
}
```

#### 3.1.4 前端：增强 Node 页面

Node 列表页增加 `Gateway` 列，显示 Node 所属的 Gateway：

```
┌────────────────────────────────────────────────────┐
│ Nodes                                    [Refresh] │
├──────────────┬──────────┬────────┬────────┬────────┤
│ Name         │ Gateway  │ Status │ OS     │ Caps   │
├──────────────┼──────────┼────────┼────────┼────────┤
│ Local-Build  │ Lyra     │ 🟢    │ Linux  │ sys,br │
│ Catherine-   │ Lyra     │ 🟢    │ Linux  │ sys,br │
│   Build      │          │        │        │        │
│ Remote-Node  │ Catherine│ 🟢    │ Linux  │ sys    │
└──────────────┴──────────┴────────┴────────┴────────┘
```

### 3.2 Phase 2：跨 Gateway Agent 操作（2 天）

**目标：ACP 工具支持跨 Gateway 的 Agent 通信。**

#### 3.2.1 后端扩展 ResourceBusService

```go
// ResolveAgent 查找 Agent 所在的 Gateway
func (s *ResourceBusService) ResolveAgent(agentID string) (*AgentLocation, error)

// SendToAgent 向任意 Agent 发消息
func (s *ResourceBusService) SendToAgent(agentID string, message string) (string, error)

// SpawnOnGateway 在指定 Gateway 上 spawn 子 Agent
func (s *ResourceBusService) SpawnOnGateway(gatewayID string, task string, opts SpawnOpts) (string, error)
```

#### 3.2.2 API Endpoints

```
GET    /api/resource-bus/agents              # 全局 Agent 列表
POST   /api/resource-bus/agents/:id/send     # 跨 Gateway 发消息
POST   /api/resource-bus/agents/:id/spawn    # 跨 Gateway spawn
GET    /api/resource-bus/agents/:id/history   # 跨 Gateway 历史
GET    /api/resource-bus/agents/:id/status    # 跨 Gateway 状态
```

### 3.3 Phase 3：Gateway 管理整合（3 天）

**目标：在 ACP 中管理 Gateway 配置、Cron、Usage。**

#### 3.3.1 Gateway 配置管理

```
GET    /api/gateways/:id/config       # 获取配置（调 config.get RPC）
PATCH  /api/gateways/:id/config       # 修改配置（调 config.patch RPC）
GET    /api/gateways/:id/config/schema # 获取配置 schema
```

前端：可视化配置编辑器
- 基于 JSON Schema 自动生成表单
- 分组显示：Agent 配置、Channel 配置、Tool 配置、安全配置
- 修改预览 + 确认 → 调 `config.patch` → 自动重启 Gateway

#### 3.3.2 Cron 管理

```
GET    /api/gateways/:id/cron         # 列出 Cron 任务
POST   /api/gateways/:id/cron         # 创建 Cron 任务
PUT    /api/gateways/:id/cron/:jobId  # 更新
DELETE /api/gateways/:id/cron/:jobId  # 删除
POST   /api/gateways/:id/cron/:jobId/run # 手动触发
```

前端：Cron 管理页
- 列表：任务名、Cron 表达式、上次运行、下次运行、状态
- 创建/编辑：表单 + Cron 表达式可视化（human-readable 转换）
- 手动触发按钮

#### 3.3.3 Usage 统计

```
GET    /api/usage                     # 跨 Gateway 汇总
GET    /api/gateways/:id/usage        # 单 Gateway 用量
```

前端：Usage 页
- Token 用量趋势图（按天/周/月）
- 按 Agent 分组的费用排行
- 按 Model 分组的费用排行
- 支持跨 Gateway 汇总

### 3.4 Phase 4：Skill 市场与 Channel 管理（2 天）

#### 3.4.1 Skill 管理

ACP 已有 `SkillService`，扩展：
- Skill 列表（按 Gateway 分组）
- Skill 安装/更新/卸载
- ClawHub 浏览与搜索（调 clawhub.com API）

#### 3.4.2 Channel 管理

- 渠道状态一览（飞书/TG/Discord 等连接状态）
- 基本配置查看

## 4. 技术实现细节

### 4.1 现有代码基础

ACP 已有的多 Gateway 基础设施：

```
internal/acp/
├── gateway/
│   ├── client.go          # WebSocket RPC 客户端
│   │   ├── NodeList()      # nodes.status RPC
│   │   ├── NodeInvoke()    # node.invoke RPC
│   │   ├── NodeDescribe()  # node.describe RPC
│   │   ├── SessionsList()  # sessions.list RPC
│   │   ├── SessionsSend()  # sessions.send RPC
│   │   ├── ConfigGet()     # config.get RPC
│   │   ├── ConfigPatch()   # config.patch RPC
│   │   ├── CronList()      # cron.list RPC
│   │   ├── AgentsList()    # agents.list RPC
│   │   ├── SkillsStatus()  # skills.status RPC
│   │   └── ...
│   └── types.go           # RPC 类型定义
├── entity/
│   └── entity.go          # Gateway 实体（ID/Name/URL/Token）
├── service/
│   ├── gateway_registry_service.go  # 多 Gateway 连接管理
│   │   ├── GetClient(gatewayID)     # 获取/创建连接
│   │   ├── List()                   # 列出所有 Gateway
│   │   ├── SyncAgents()             # 同步远程 Agent
│   │   └── TestConnection()         # 测试连接
│   ├── agent_service.go             # Agent 管理（已支持 gatewayId 路由）
│   ├── node_service.go              # Node 管理（已支持 gatewayId 路由）
│   └── ...
```

**关键：大部分 RPC 方法和路由逻辑已存在**，Resource Bus 主要是：
1. 增加「按资源名自动 Resolve Gateway」的逻辑
2. 包装为 Plugin Tool
3. 增加缓存层

### 4.2 Gateway RPC 方法补充

当前 `client.go` 缺少的 RPC 方法，需新增：

```go
// nodes.run — 在 Node 上执行命令（目前只有 NodeInvoke）
func (c *Client) NodeRun(nodeID string, command []string, cwd string, timeoutMs int) (*NodeRunResult, error)

// sessions.spawn — 在 Gateway 上 spawn 子 Agent
func (c *Client) SessionsSpawn(task string, opts SpawnOpts) (*SpawnResult, error)

// cron.add / cron.update / cron.remove — Cron 管理
func (c *Client) CronAdd(job CronJobDef) (*CronJob, error)
func (c *Client) CronUpdate(jobID string, patch map[string]interface{}) error
func (c *Client) CronRemove(jobID string) error
func (c *Client) CronRun(jobID string) error

// usage.get — 用量统计
func (c *Client) UsageGet(opts UsageOpts) (*UsageResult, error)

// channels.list — 渠道列表
func (c *Client) ChannelsList() ([]ChannelInfo, error)
```

### 4.3 ACP 权限模型（Resource Bus ACL）

**跨 Gateway 操作的核心安全问题：** ACP 持有所有 Gateway 的 operator token，Agent 通过 ACP 工具调用时，实际以 ACP 的 operator 权限执行，可能绕过 Gateway 原本对 Agent 的权限约束。

**解决方案：ACP 层自建资源访问控制（ACL）。**

#### 4.3.1 权限配置

```go
// internal/acp/service/acl.go
type ResourceACL struct {
    // AgentID → 允许访问的资源
    Rules map[string]*AgentACLRule
    // 默认规则（未配置的 Agent）
    DefaultRule *AgentACLRule
}

type AgentACLRule struct {
    AllowedNodes    []string  // Node 名称/ID 白名单，"*" 表示全部
    AllowedGateways []string  // Gateway ID 白名单，"*" 表示全部
    AllowedActions  []string  // 操作白名单：exec, invoke, browse, send, spawn
    DenyNodes       []string  // Node 黑名单（优先于白名单）
    MaxConcurrent   int       // 最大并发操作数
}
```

ACP 配置文件中定义：
```yaml
# acp-config.yaml
resource_acl:
  default:
    allowed_nodes: ["*"]      # 默认允许所有 Node
    allowed_gateways: ["*"]   # 默认允许所有 Gateway
    allowed_actions: ["exec", "invoke", "describe"]
    max_concurrent: 5
  
  agents:
    lyra:
      allowed_nodes: ["*"]
      allowed_gateways: ["*"]
      allowed_actions: ["*"]   # COO 级别，全部权限
      max_concurrent: 10
    
    catherine:
      allowed_nodes: ["Catherine-Build"]
      allowed_gateways: ["catherine"]
      allowed_actions: ["exec", "invoke", "browse"]
      max_concurrent: 5
    
    intern-agent:
      allowed_nodes: ["Sandbox-Node"]
      allowed_gateways: ["default"]
      allowed_actions: ["exec"]
      deny_nodes: ["Production-*"]  # 禁止访问生产 Node
      max_concurrent: 2
```

#### 4.3.2 权限检查流程

```
Agent 调用 acp_node_run(node="Production-DB", command=["rm", "-rf", "/"])
  → ACP 收到请求
  → 识别调用者 Agent ID
  → 查 ACL：该 Agent 是否允许操作 "Production-DB"？
  → 查 ACL：该 Agent 是否允许 "exec" 操作？
  → 不允许 → 返回 403 "Permission denied: agent 'intern-agent' cannot access node 'Production-DB'"
  → 允许 → 继续路由到目标 Gateway
```

#### 4.3.3 审计日志

所有跨 Gateway 操作记录审计日志：

```go
type AuditEntry struct {
    Timestamp   time.Time
    AgentID     string
    Action      string    // exec, invoke, send, spawn, ...
    TargetType  string    // node, agent, gateway
    TargetID    string    // node name, agent id, gateway id
    GatewayID   string    // 实际路由到的 Gateway
    Allowed     bool      // ACL 判定结果
    Result      string    // success, denied, error
    Details     string    // 命令内容、错误信息等
}
```

审计日志存 DB（可选），ACP 管理 UI 可查看。

### 4.4 安全考虑

| 风险 | 缓解方案 |
|---|---|
| Agent 通过 ACP 绕过 Gateway 权限 | ACP 检查源 Agent 的权限级别，不允许越权 |
| 跨 Gateway Token 泄露 | Token 只存 DB，API 永不返回，日志脱敏 |
| Node 命令注入 | ACP 继承目标 Gateway 的安全策略（deny/allowlist/full） |
| DDoS：Agent 频繁查全局资源 | 缓存 + 速率限制 |

### 4.5 性能考虑

| 场景 | 策略 |
|---|---|
| 全局 Node 列表 | 缓存 30s，惰性刷新 |
| Agent Resolve | 缓存 + LRU，Agent 不会频繁换 Gateway |
| 跨 Gateway RPC | 已有长连接（WebSocket），延迟 < 50ms |
| 多 Gateway 并发查询 | `sync.WaitGroup` 并发调用，不串行 |

## 5. 实施路线图

```
Phase 1（Week 1）：资源调度总线核心
  ├── ResourceBusService + Node Resolve/Exec
  ├── Resource Bus API endpoints
  ├── 原生 Plugin Tool 注册（acp_node_run 等）
  ├── Resource Bus ACL 权限模型
  └── 前端 Node 列表增加 Gateway 列

Phase 2（Week 2）：跨 Gateway Agent 操作
  ├── Agent Resolve/Send/Spawn
  ├── Plugin Tool 扩展（acp_agent_send 等）
  └── 前端 Agent 列表增加 Gateway 标识

Phase 3（Week 2-3）：Gateway 管理整合
  ├── Config 可视化编辑器
  ├── Usage 统计页
  ├── Cron 可视化管理
  └── 全局 Dashboard 增强

Phase 4（Week 3-4）：Skill 市场 + Channel 管理
  ├── Skill 安装/更新 UI
  ├── Channel 状态/配置
  └── 日志查看器
```

## 6. 成功标准

1. **Agent 透明跨 Gateway**：Lyra 调 `acp_node_run(node="Remote-Node", ...)` 能自动路由到正确 Gateway 并返回结果，Agent 不感知 Gateway 存在
2. **统一入口**：用户日常管理工作 100% 在 ACP 界面完成，不需要打开 Gateway Web UI
3. **零侵入**：不修改 OpenClaw Gateway 源码，仅通过 RPC + 原生 Plugin 机制集成
4. **性能无损**：跨 Gateway 操作延迟 < 200ms（同机房），Agent 执行效率不受影响

## 7. ACP 平台定位总结

```
┌─────────────────────────────────────────────┐
│              ACP 统一平台                     │
│                                              │
│  ✅ 业务编排（流程/任务/知识库/经验库）        │
│  ✅ 多 Gateway 统一管理                       │
│  ✅ 跨 Gateway 资源调度（Resource Bus）       │
│  ✅ Gateway 功能整合（Config/Cron/Usage/...） │
│  ✅ 唯一管理 UI 入口                          │
│                                              │
│  类比：                                       │
│  • Gateway = 单个 K8s 集群                    │
│  • ACP = Rancher（多集群管理 + 应用编排）     │
│  • Node = Worker Node                         │
│  • Agent = Pod/Workload                       │
│  • Process = Workflow/Pipeline                │
│  • Resource Bus = 联邦 API Gateway            │
└─────────────────────────────────────────────┘
```

**ACP = Agent 的操作系统。Gateway 是底层 runtime，ACP 是用户空间。**
