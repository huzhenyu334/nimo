# PLM 项目甘特图 × ACP 流程集成

> 版本: v1.0 | 日期: 2026-04-01 | 编制: Claude Code

---

## 一、背景与问题

### 1.1 现状

PLM 项目管理模块已有甘特图功能，任务数据存储在 PLM 的 `tasks` 表中，支持阶段分组、父子任务、依赖关系、里程碑标记。

ACP（Agent Control Panel）是流程执行引擎，通过 YAML 定义流程，支持 if/switch/loop/foreach 等控制流。ACP 也有自己的甘特图组件 `FlowGantt`，支持嵌套树形渲染、实时执行状态、断点调试。

### 1.2 核心矛盾

| 维度 | PLM 甘特图 | ACP 流程 |
|------|-----------|---------|
| 数据性质 | 计划视图（固定任务列表+时间线） | 执行图（含运行时逻辑分支） |
| 节点类型 | 任务/里程碑 | 执行步骤 + 控制节点（if/switch/loop/foreach） |
| 确定性 | 项目立项时可排定 | 运行时才知道走哪条分支 |
| 受众 | 项目经理 | 流程设计者/执行引擎 |

**问题**：PLM 项目的任务实际由 ACP 流程的 step 定义和驱动，但 ACP 流程中包含大量控制流节点，不能直接作为甘特图的任务项展示给项目经理。

### 1.3 设计原则

1. **ACP 零改动** — 不在 ACP 的 StepDef 中添加 PLM 相关字段，不修改 ACP 引擎逻辑
2. **组件解耦复用** — FlowGantt 提取为独立 UI 包，PLM 和 ACP 均通过包引用使用
3. **过滤逻辑归属 PLM** — 哪些 step 可见、哪些是里程碑，完全由 PLM 侧决定
4. **HTTP 接口通信** — PLM 通过 ACP REST API 获取数据，不 import ACP Go 包

---

## 二、方案总览

### 2.1 架构

```
┌─────────────────────────────────────────────────────┐
│                    @nimo/gantt                       │
│              (独立 UI 组件包，纯渲染)                  │
│                                                     │
│  Props: nodes: GanttNode[], mode, onNodeClick       │
└──────────┬──────────────────────────┬───────────────┘
           │                          │
    ┌──────┴──────┐           ┌──────┴──────┐
    │  ACP 前端    │           │  PLM 前端    │
    │  传完整节点   │           │  传过滤后节点  │
    └──────┬──────┘           └──────┬──────┘
           │                          │
           │                   ┌──────┴──────┐
           │                   │  PLM 后端    │
           │                   │ gantt_filter │
           │                   └──────┬──────┘
           │                          │ HTTP API
           │                   ┌──────┴──────┐
           └───────────────────│  ACP 后端    │
                               │  (零改动)    │
                               └─────────────┘
```

### 2.2 数据流

```
1. PLM 项目关联 ACP process_id
2. PLM 后端调用 ACP API：
   - GET /api/processes/:id          → 获取流程定义（steps 树）
   - GET /api/instances/:id/tasks    → 获取执行状态（如有运行中实例）
3. PLM 后端执行过滤：
   - 递归遍历所有 steps（含 subprocess 展开）
   - 应用过滤规则，去除控制节点和不可见 executor
   - 修复被过滤节点导致的依赖链断裂
   - 输出 GanttNode[] 树
4. PLM 前端接收 GanttNode[]，传入 @nimo/gantt 组件渲染
```

---

## 三、FlowGantt 组件提取

### 3.1 包结构

```
packages/
  gantt/
    ├── package.json          # @nimo/gantt
    ├── tsconfig.json
    └── src/
        ├── index.ts          # 导出入口
        ├── FlowGantt.tsx     # 甘特图主组件（从 acp-web 迁移）
        ├── types.ts          # 纯数据类型定义
        └── constants.ts      # 颜色、tick 配置等
```

### 3.2 组件接口

```typescript
// packages/gantt/src/types.ts

export interface GanttNode {
  id: string                    // step ID
  label: string                 // 显示名称
  status: string                // pending | running | completed | failed | skipped
  depth: number                 // 树深度（用于缩进）
  children: GanttNode[]         // 子节点

  // 时间
  started_at?: number           // 执行开始时间戳 (ms)
  completed_at?: number         // 执行结束时间戳 (ms)
  duration_ms?: number          // 实际耗时
  planned_start?: number        // 计划开始时间戳 (ms)
  planned_duration_ms?: number  // 计划工期

  // 标记
  isMilestone?: boolean         // 里程碑
  isConditional?: boolean       // 来自条件分支（可能被跳过）
  isRepeating?: boolean         // 来自循环体（可能执行多次）

  // 元数据（可选，用于 tooltip）
  executor?: string             // 步骤执行器类型
  assignee?: string             // 负责人
}

export interface FlowGanttProps {
  nodes: GanttNode[]                      // 树形数据
  mode: 'plan' | 'execution'              // 计划模式 or 执行模式
  onNodeClick?: (nodeId: string) => void  // 点击回调
  className?: string
}
```

### 3.3 workspace 配置

根目录 `package.json`:

```json
{
  "private": true,
  "workspaces": [
    "packages/*",
    "nimo-plm-web",
    "agent-control-panel/acp-web"
  ]
}
```

ACP 和 PLM 前端均通过 `"@nimo/gantt": "workspace:*"` 引用。

### 3.4 迁移策略

1. 从 `acp-web/src/components/flow/FlowGantt/` 复制到 `packages/gantt/src/`
2. 剥离 ACP 特有逻辑（debug API 调用、breakpoint 操作），仅保留纯渲染
3. ACP 前端改为 `import { FlowGantt } from '@nimo/gantt'`，debug/breakpoint 逻辑留在 ACP 侧通过 `onNodeClick` 等回调处理
4. PLM 前端引入同一个包

---

## 四、Step 过滤规则

### 4.1 规则架构

过滤逻辑完全在 PLM 后端实现，分两层：

```
Layer 1: 节点类型过滤（硬规则，不可配置）
  └── control 字段有值 → 节点本身 SKIP，递归处理其 branches/body steps

Layer 2: executor 类型过滤（软规则，可调整默认值）
  └── 按 executor 名称判断是否可见
```

### 4.2 Layer 1 — 控制节点过滤

控制节点（`control` 字段非空）**永远不作为任务项显示**，但其内部的子步骤需要递归处理：

| control 类型 | 处理方式 |
|-------------|---------|
| `if` | 节点本身跳过；所有分支（true/false）内的步骤递归处理，标记 `isConditional: true` |
| `switch` | 节点本身跳过；所有 case 分支内的步骤递归处理，标记 `isConditional: true` |
| `loop` | 节点本身跳过；body 内的步骤递归处理，标记 `isRepeating: true`。**计划模式**展示一次（模板）；**执行模式**按实际轮次展开（见 4.6） |
| `foreach` | 节点本身跳过；body 内的步骤递归处理，标记 `isRepeating: true`。**计划模式**展示一次；**执行模式**按实际迭代展开 |
| `terminate` | 跳过，不产生任何可见节点 |

### 4.6 Loop 迭代的甘特图渲染（EVT/DVT 多轮场景）

硬件开发的阶段迭代（EVT1→EVT2、DVT1→DVT2→DVT3）是典型的 loop + do-while 模式。甘特图需要区分两种模式：

**计划模式**（流程未启动）：
- Loop body 步骤展示一次，标记 `isRepeating: true`
- 计划工期按单轮估算
- 甘特图 tooltip 提示"循环步骤，实际轮次取决于评审结果"

**执行模式**（流程实例运行中）：
- 从 ACP tasks 表读取实际产生的 task 记录
- ACP 的 scoped step ID 格式：`{loop_step}-{round}-{body_step}`（如 `evt_phase-2-evt_build`）
- PLM 按轮次（round）分组，渲染为可折叠的子树：

```
├── EVT 阶段
│   ├── 第1轮
│   │   ├── EVT 样品制作     ✅ completed
│   │   ├── EVT 测试验证     ✅ completed
│   │   └── ◆ EVT 评审      ❌ fail（需重做）
│   └── 第2轮
│       ├── EVT 样品制作     🔵 running
│       ├── EVT 测试验证     ⏳ pending
│       └── ◆ EVT 评审      ⏳ pending
```

- 每轮的实际日期和状态从 ACP task 记录获取
- 未通过的评审轮次标记为 `fail`，帮助 PM 追溯迭代历史

### 4.3 Layer 2 — executor 类型过滤

| executor | 默认可见 | 理由 |
|----------|---------|------|
| `human` | **true** | 人工任务，项目管理核心关注点 |
| `agent` | **true** | AI 执行任务，有实际产出物 |
| `subprocess` | **递归展开** | 读取子流程定义，对子流程的 steps 应用同样的过滤规则 |
| `shell` | false | 自动化脚本，基础设施 |
| `http` | false | API 调用，基础设施 |
| `timer` | false | 延时/定时器 |
| `calculator` | false | 计算节点 |
| `llm` | false | LLM 调用 |
| `knowledge` | false | 知识库操作 |
| `github` | false | GitHub 操作 |
| 未知 executor | false | 安全默认 |

默认规则以 Go map 形式硬编码在 PLM 代码中。后续如需调整，改代码即可，不需要 UI。

### 4.4 里程碑识别规则

唯一的自动规则：**ACP step 的 `type` == `approval` → 里程碑**。审批节点是结构性信息，天然代表阶段性签字确认，不依赖命名猜测。

```go
func isMilestone(step ACPStepDef) bool {
    return step.Type == "approval"
}
```

不做关键词匹配、不做位置推断。简单、可靠、零误判。

**后续扩展（按需）**：如果 PM 反馈有非 approval 步骤也需要标记为里程碑，在 PLM 数据库增加映射表：

```sql
-- PLM 数据库，非 ACP 数据库
CREATE TABLE plm_milestone_config (
    process_id  VARCHAR(64),   -- ACP 流程 ID
    step_id     VARCHAR(100),  -- ACP step ID
    PRIMARY KEY (process_id, step_id)
);
```

这是流程模板级配置，配一次适用于所有使用该流程的项目。v1 不实现，留作扩展点。

### 4.5 依赖链穿透

控制节点被过滤后，其下游步骤的 `depends_on` 指向了一个不存在的节点。需要依赖穿透修复：

```
原始 DAG:
  A → if_node → B
        ├── branch_step_1
        └── branch_step_2

过滤后:
  A → branch_step_1 (conditional)
  A → branch_step_2 (conditional)
  branch_step_1 → B (or branch_step_2 → B)
```

穿透算法：

```
对每个可见步骤 S:
  对 S.depends_on 中的每个依赖 D:
    if D 是可见步骤:
      保留依赖
    else if D 是控制节点:
      替换为 D.depends_on（递归，直到找到可见步骤）
    else if D 被 executor 规则过滤:
      替换为 D.depends_on（递归）
```

---

## 五、PLM 后端实现

### 5.1 ACP HTTP Client

```go
// internal/plm/service/acp_client.go

type ACPClient struct {
    BaseURL    string // http://localhost:3001
    Token      string // JWT token
    HTTPClient *http.Client
}

// 获取流程定义（含完整 steps 树）
func (c *ACPClient) GetProcessDef(processID string) (*ACPProcessDef, error)

// 获取流程实例下所有 task 的执行状态
func (c *ACPClient) GetInstanceTasks(instanceID string) ([]ACPTaskStatus, error)

// 获取项目关联的最新流程实例
func (c *ACPClient) GetLatestInstance(processID string) (*ACPInstance, error)
```

数据结构仅定义 PLM 需要的字段，不复制 ACP 的完整 entity：

```go
type ACPProcessDef struct {
    ID    string        `json:"id"`
    Name  string        `json:"name"`
    Steps []ACPStepDef  `json:"steps"`
}

type ACPStepDef struct {
    ID              string                 `json:"id"`
    Name            string                 `json:"name"`
    Executor        string                 `json:"executor"`
    Control         string                 `json:"control"`
    DependsOn       []string               `json:"depends_on"`
    PlannedDuration string                 `json:"planned_duration"`
    Branches        map[string][]ACPStepDef `json:"branches,omitempty"`
    Steps           []ACPStepDef           `json:"steps,omitempty"` // loop/foreach body
    Items           string                 `json:"items,omitempty"`
    ProcessID       string                 `json:"process_id,omitempty"` // subprocess 引用
}

type ACPTaskStatus struct {
    StepID      string  `json:"step_id"`
    Status      string  `json:"status"`
    StartedAt   *int64  `json:"started_at"`
    CompletedAt *int64  `json:"completed_at"`
}
```

### 5.2 过滤服务

```go
// internal/plm/service/gantt_filter.go

type GanttFilterService struct {
    acpClient *ACPClient
}

// 核心方法：读取 ACP 流程 → 过滤 → 输出甘特图节点
func (s *GanttFilterService) BuildProjectGantt(processID string, instanceID string) ([]GanttNode, error)

// 递归过滤 steps
func (s *GanttFilterService) filterSteps(steps []ACPStepDef, depth int) []GanttNode

// 依赖链穿透修复
func (s *GanttFilterService) fixDependencies(nodes []GanttNode, originalSteps []ACPStepDef)

// 里程碑识别
func (s *GanttFilterService) detectMilestones(nodes []GanttNode)

// 合并执行状态
func (s *GanttFilterService) mergeTaskStatus(nodes []GanttNode, tasks []ACPTaskStatus)
```

### 5.3 PLM API 端点

```
GET /api/v1/projects/:id/gantt
```

Response:

```json
{
  "nodes": [
    {
      "id": "requirements_review",
      "label": "需求评审",
      "status": "completed",
      "depth": 0,
      "isMilestone": true,
      "started_at": 1711929600000,
      "completed_at": 1711944000000,
      "duration_ms": 14400000,
      "executor": "human",
      "children": []
    },
    {
      "id": "structural_design",
      "label": "结构设计",
      "status": "running",
      "depth": 0,
      "isMilestone": false,
      "isConditional": false,
      "started_at": 1711958400000,
      "planned_duration_ms": 604800000,
      "executor": "human",
      "children": []
    }
  ],
  "mode": "execution",
  "process_name": "nimo-v2-开发流程",
  "instance_status": "running"
}
```

---

## 六、PLM 前端集成

### 6.1 项目详情页甘特图 Tab

现有 `ProjectDetail.tsx` 的甘特图 Tab 改为从新 API 获取数据，使用 `@nimo/gantt` 组件渲染：

```typescript
// nimo-plm-web/src/pages/ProjectDetail.tsx (甘特图 Tab)

import { FlowGantt } from '@nimo/gantt'

function GanttTab({ projectId }: { projectId: string }) {
  const { data } = useQuery(['project-gantt', projectId], () =>
    api.get(`/projects/${projectId}/gantt`)
  )

  return (
    <FlowGantt
      nodes={data.nodes}
      mode={data.mode}
      onNodeClick={(id) => openTaskDetail(id)}
    />
  )
}
```

### 6.2 视觉标记

| 标记 | 甘特图表现 |
|------|-----------|
| `isMilestone` | 菱形标记 ◆，不显示时间条 |
| `isConditional` | 虚线边框，tooltip 提示"条件分支，可能跳过" |
| `isRepeating` | 循环图标 🔄，tooltip 提示"循环步骤，可能执行多次" |
| `status: skipped` | 灰色，删除线 |

---

## 七、PLM 项目与 ACP 流程的关联

### 7.1 关联方式

`projects` 表新增字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| `acp_process_id` | VARCHAR(64) | 关联的 ACP 流程定义 ID |
| `acp_instance_id` | VARCHAR(64) | 关联的 ACP 流程实例 ID（项目启动后填入） |

### 7.2 生命周期

```
1. 创建项目时，选择关联的 ACP 流程（可选）
2. 项目启动时，调 ACP API 启动流程实例，记录 instance_id
3. 甘特图渲染时，读取流程定义 + 实例状态
4. 流程实例完成时，项目自动标记为 completed
```

---

## 八、Subprocess 递归展开

当过滤遇到 `executor: subprocess` 的步骤时：

1. 读取该步骤的 `process_id`（子流程引用）
2. 调 ACP API 获取子流程定义
3. 对子流程的 steps 递归应用同样的过滤规则
4. 将过滤后的子步骤作为该 subprocess 步骤的 `children`
5. 甘特图中以可折叠的子树形式展示

```
甘特图:
├── 需求评审                    ← 顶层流程的 human step
├── ▶ 硬件开发子流程              ← subprocess，可折叠
│   ├── PCB 设计                ← 子流程的 human step
│   ├── 结构设计                ← 子流程的 human step
│   └── ◆ 硬件评审              ← 子流程的 milestone
├── 软件开发                    ← 顶层流程的 human step
└── ◆ EVT 评审                 ← 顶层流程的 milestone
```

递归深度限制：最大 5 层，防止循环引用。

---

## 九、实施计划

### Phase 1 — 组件提取 + 基础集成

1. 创建 `packages/gantt/` 包，从 ACP 迁移 FlowGantt 组件
2. ACP 前端切换到 `@nimo/gantt` 引用，确认功能不退化
3. PLM 后端实现 `ACPClient` + `GanttFilterService`
4. PLM 后端实现 `GET /api/v1/projects/:id/gantt` 端点
5. PLM 前端甘特图 Tab 对接新组件

### Phase 2 — Subprocess 递归 + 状态同步

1. 实现 subprocess 递归展开
2. 实现执行状态合并（从 ACP instance tasks 读取）
3. 实现依赖链穿透修复

### Phase 3 — 里程碑 + 视觉增强

1. 实现里程碑自动识别
2. 条件分支/循环体的视觉标记
3. 计划模式 vs 执行模式切换

---

## 十、风险与边界

| 风险 | 应对 |
|------|------|
| ACP API 不可用 | PLM 甘特图显示"流程数据加载失败"，不影响项目其他功能 |
| Subprocess 循环引用 | 递归深度限制 5 层 + 已访问 processID 集合去重 |
| 大型流程步骤过多 | 过滤后通常只保留 human/agent 步骤，数量可控 |
| ACP 流程定义变更 | PLM 每次渲染时实时读取 ACP，不缓存流程定义 |
| 条件分支导致甘特图任务不确定 | `isConditional` 标记 + 运行时自动标记 `skipped` |
