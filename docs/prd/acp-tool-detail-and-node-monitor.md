# PRD: ACP Tool调用详情展示 + Node监控页面

## 背景

ACP的Agent对话和Step过程都能显示agent的工具调用，但**只显示工具名称，缺少详细的输入参数和输出结果**。OpenClaw的`chat.history(includeTools=true)` API返回了完整的toolCall参数和toolResult结果，但后端转换时丢弃了这些数据。

同时需要新增Node监控页面，从chat history中过滤展示跟特定Node相关的工具调用。

## 目标

1. **Agent对话窗口**：完整展示tool调用的参数和结果（前端组件已有，只需后端填数据）
2. **Step过程窗口**：同样展示tool调用详情
3. **Node监控页面**：聚合展示某个Node上的所有工具调用活动

## 需求1：Agent对话 — Tool详情修复

### 后端修改（agent_handler.go History方法）

**现状问题**：`agent_handler.go` 第120-165行，将gateway.Message转换为前端格式时：
- `toolCall`类型只提取了工具名：`content += "[Tool: " + name + "]\n"`，丢弃了参数
- `toolResult`消息转成了`role: "system"`，没有和对应的toolCall关联

**修改方案**：

1. 遍历messages时，识别assistant消息中的toolCall content blocks，提取：
   - `name`：工具名
   - `id`：toolCall ID
   - `arguments`：完整参数JSON

2. 遍历messages时，识别`role: "toolResult"`消息，通过`toolCallId`和前面的toolCall配对

3. 输出格式匹配前端已有的`ToolCall`接口：
```typescript
interface ToolCall {
  id: string;        // toolCall ID
  name: string;      // 工具名称
  paramsSummary: string;  // 参数摘要（单行，用于折叠标题）
  paramsFull: string;     // 完整参数JSON（展开后显示）
  result?: string;        // toolResult的完整内容
}
```

4. assistant消息的输出格式：
```json
{
  "id": "msg-5",
  "role": "assistant",
  "content": "让我来编译项目...",   // 只保留text内容
  "toolCalls": [
    {
      "id": "toolu_01xxx",
      "name": "exec",
      "paramsSummary": "go build ./cmd/plm/",   // 从参数中提取关键信息
      "paramsFull": "{\"command\": \"go build ./cmd/plm/\", \"timeout\": 30}",
      "result": "编译成功\n..."
    }
  ],
  "timestamp": "1771989313745"
}
```

5. `role: "toolResult"`的消息不再单独输出，而是合并到对应assistant消息的`toolCalls[].result`中

### paramsSummary 生成规则

根据工具名提取最有代表性的参数：
- `exec`/`Bash` → command字段
- `Read`/`Write`/`Edit` → file_path或path字段
- `web_search` → query字段
- `web_fetch` → url字段
- `nodes` → action + node字段
- `message` → action + target字段
- 其他 → 参数JSON的前100个字符

### 前端（已基本就绪）

`AgentDetail.tsx` 第130-165行已有Collapse组件展示toolCalls，只需确认：
- 数据能正确绑定
- `paramsFull`用代码块渲染
- `result`也用代码块渲染，带语法高亮或至少等宽字体
- 长结果截断到2000字符，加"查看更多"按钮

## 需求2：Step过程 — Tool详情修复

### 后端修改（workflow_engine.go GetAgentContext方法）

**现状问题**：`GetAgentContext`方法（第3442行起）将messages转成`AgentContextMessage`时：
- toolCall的参数没有结构化提取
- toolResult只保存了content文本，没有和toolCall配对

**修改方案**：

1. 扩展`AgentContextMessage`结构体，增加toolCall详情字段：
```go
type AgentContextMessage struct {
    Role       string          `json:"role"`
    Content    string          `json:"content"`
    ToolName   string          `json:"tool_name,omitempty"`
    ToolData   json.RawMessage `json:"tool_data,omitempty"`
    ToolCalls  []ToolCallInfo  `json:"tool_calls,omitempty"`  // 新增
    Timestamp  int64           `json:"timestamp"`
    Model      string          `json:"model,omitempty"`
}

type ToolCallInfo struct {
    ID           string `json:"id"`
    Name         string `json:"name"`
    ParamsSummary string `json:"params_summary"`
    ParamsFull   string `json:"params_full"`
    Result       string `json:"result,omitempty"`
}
```

2. 解析逻辑与需求1一致：assistant消息提取toolCall，toolResult按ID配对

### 前端修改（WorkflowRunDetail.tsx ChatBubble组件）

**现状**：ChatBubble只显示tool_name，点击展开显示raw content

**改为**：
- assistant消息下方显示toolCalls列表（类似Agent对话的Collapse样式）
- 每个toolCall显示：🔧 工具名 — 参数摘要
- 展开后显示完整参数 + 结果
- 统一使用和Agent对话一样的Collapse组件样式

## 需求3：Node监控页面

### 概述

新增 `/nodes` 页面，展示所有已连接的Node，点击进入Node详情，看到该Node上的实时活动。

### 数据来源

Node监控的数据来源于**所有Agent的chat history**，过滤条件：
- `toolName` 为 `nodes` 或 `exec`（host=node）相关的调用
- toolCall的参数中包含该node的nodeId

### 页面设计

#### Node列表页（/nodes）

- 卡片或表格展示所有Node
- 每个Node显示：名称、平台、版本、在线状态、最近活动时间
- 数据来源：`GET /api/nodes?gatewayId=xxx`（已有接口）

#### Node详情页（/nodes/:nodeId）

**顶部**：Node基本信息
- 名称、平台、版本、IP、连接时间
- Capabilities（browser/system等）
- 在线状态指示灯

**主体**：活动流（Activity Stream）

实时展示该Node上的所有工具调用，格式：

```
[11:05:32] 🤖 Lyra → exec: go build ./cmd/plm/
  ├─ 参数: {"command": "go build ./cmd/plm/", "host": "node", "node": "Catherine-Build"}
  └─ 结果: (exit 0) 编译成功

[11:05:45] 🤖 Lyra → nodes.run: system.run ["npm", "run", "build"]
  ├─ 参数: {"command": ["npm", "run", "build"], "nodeId": "0379eee..."}
  └─ 结果: (exit 0) vite build completed in 3.2s
```

每条记录包含：
- 时间戳
- 执行者Agent
- 工具名 + 参数摘要
- 可展开查看完整参数和结果

**右侧面板（可选）**：Node环境信息
- 已安装工具（来自probe结果）
- 磁盘/内存使用（定期system.run获取）

### 后端新接口

```
GET /api/nodes/:nodeId/activity?gatewayId=xxx&limit=50
```

实现逻辑：
1. 获取该gateway上所有agent列表
2. 遍历每个agent的chat.history(includeTools=true)
3. 过滤出与该nodeId相关的tool调用
4. 按时间排序返回

过滤规则：
- `toolName == "exec"` 且参数中 `host == "node"` 或 `node` 字段匹配
- `toolName == "nodes"` 且参数中 `node` 字段匹配该nodeId
- `toolName == "process"` 关联到node exec session的

返回格式：
```json
{
  "activities": [
    {
      "timestamp": 1771989313745,
      "agentId": "main",
      "agentName": "Lyra",
      "toolName": "exec",
      "paramsSummary": "go build ./cmd/plm/",
      "paramsFull": "{...}",
      "result": "...",
      "exitCode": 0
    }
  ]
}
```

### 前端自动刷新

- Node详情页每3秒轮询一次activity接口
- 新活动出现时自动滚动到底部
- 支持暂停自动刷新（用户向上滚动时暂停）

## 实现优先级

1. **P0 — Agent对话toolCall修复**（后端改agent_handler.go，前端已有组件）
2. **P0 — Step过程toolCall修复**（后端改workflow_engine.go + 前端ChatBubble组件）
3. **P1 — Node监控页面**（新增后端接口 + 前端页面）

## 关键代码位置

| 组件 | 文件 | 行号 |
|------|------|------|
| Agent History后端 | `internal/acp/handler/agent_handler.go` | 102-170 |
| Agent History前端 | `acp-web/src/pages/AgentDetail.tsx` | 130-165 |
| Agent History API类型 | `acp-web/src/api/agents.ts` | 38-55 |
| Step Context后端 | `internal/acp/service/workflow_engine.go` | 3432-3507 |
| Step Context前端 | `acp-web/src/pages/WorkflowRunDetail.tsx` | 1033-1070 |
| Step Context API类型 | `acp-web/src/api/processes.ts` | 146-162 |
| Node列表后端（已有） | `internal/acp/handler/gateway_handler.go` | NodeList |
| Node Service | `internal/acp/service/node_service.go` | 全文件 |
| Gateway Client | `internal/acp/gateway/client.go` | SessionsHistory, NodeList |
| Gateway Types | `internal/acp/gateway/types.go` | Message, NodeInfo |
| 路由注册 | `internal/acp/handler/handler.go` | 路由组 |

## 注意事项

1. **toolCall和toolResult配对**：通过`toolCallId`字段关联。一个assistant消息可能包含多个toolCall。
2. **chat.history原始格式**：assistant消息的content是数组，包含`{type:"text"}`和`{type:"toolCall"}`两种block。toolResult是独立消息，`role:"toolResult"`。
3. **Node activity过滤**：需要扫描toolCall的arguments JSON来判断是否跟某个node相关，注意性能（可以限制history条数）。
4. **结果截断**：tool result可能很长（编译输出等），后端截断到最大8KB，前端截断到2000字符带展开按钮。
5. **前端轮询已有**：AgentDetail已有`setInterval`轮询，确保轮询间隔适中（Agent对话5s，Node监控3s）。
