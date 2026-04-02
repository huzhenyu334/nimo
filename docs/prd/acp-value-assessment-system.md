# ACP 统一价值评估与自进化流程体系 PRD

> 版本：v2.0 | 日期：2026-03-13 | 作者：泽斌 & Lyra & Claude
> 基于 v1.0 理论框架，整合现有审计基础设施，面向落地实现

---

## 一、核心目标

构建两个元流程（Meta-Process）：

1. **Auto-Create Process** — Agent 根据业务需求自动创建可执行流程
2. **Auto-Iterate Process** — Agent 根据执行数据自动优化现有流程

用这两个元流程构建全公司所有业务流程，形成**自进化流程系统**。

价值评估体系是 Auto-Iterate 的眼睛——没有评估，Agent 无法判断迭代是否是改进。

### 设计原则

1. **评估信号分层**：执行层（P0）→ 产出层（P1）→ 业务层（P2），每层独立有价值
2. **渐进自动化**：人工审批 → 低风险自动 → 高风险人工，逐步放权
3. **复用现有基础设施**：ACP 已有 RunAudit/StepEvent/ExecutorTrace/AgentPerformance，在此基础上扩展
4. **万物皆变换**：Executor/Step/流程/子流程是同一个抽象，评估框架对所有层次对称

---

## 二、现有基础设施盘点（不重建）

| 已有组件 | 表/实体 | 提供的数据 |
|---------|---------|-----------|
| RunAudit | `run_audits` | 每次运行的综合评分(0-100)、耗时、步骤成功/失败数、Gate通过率 |
| StepEvent | `step_events` | 步骤级事件流（type/level/summary/payload/duration_ms） |
| ExecutorTrace | `executor_trace` | Executor 事件溯源（message/tool_call/vote 等） |
| AgentPerformance | `agent_performance` | Agent 级指标（token/cost/gate_loops/escalation） |
| WorkflowEvent | `workflow_events` | 流程生命周期事件（created/completed/failed） |
| AuditService | `audit_service.go` | 聚合分析（瓶颈检测/Agent排名/退化告警/趋势） |
| Task.PrivateData | `tasks.private_data` | Executor 私有审计元数据 |

**本 PRD 不重建上述组件，只在其基础上新增三个模块：血缘追踪、业务结果、迭代管理。**

---

## 三、Phase 列表与依赖关系

```
Phase 0: 血缘采集（Lineage Capture）
    ↓
Phase 1: 流程健康仪表盘（Process Health Dashboard）
    ↓
Phase 2: Auto-Create 元流程（Agent 自动创建流程）
    ↓
Phase 3: Auto-Iterate 元流程（Agent 自动迭代流程）
    ↓
Phase 4: 业务结果接入与归因（Business Outcome Attribution）
```

| Phase | 名称 | 前置依赖 | 预估工作量 | 核心交付 |
|-------|------|---------|-----------|---------|
| P0 | 血缘采集 | 无 | 3-5 天 | lineage_edges 表 + 引擎自动采集 + 查询 API |
| P1 | 健康仪表盘 | P0（可选） | 5-7 天 | 前端流程健康视图 + 步骤瓶颈分析 |
| P2 | Auto-Create | P1（观测能力） | 7-10 天 | 元流程 YAML + 流程生成 Agent + 审批发布 |
| P3 | Auto-Iterate | P1 + P2 | 10-15 天 | 健康巡检 + 迭代建议 + 安全约束 + 灰度发布 |
| P4 | 业务结果归因 | P0 + P3 | 10-15 天 | 结果节点 + Webhook + 归因算法 + 价值视图 |

---

## 四、Phase 0 — 血缘采集

### 4.1 数据模型

```go
// LineageEdge 记录计算单元之间的数据流动关系
type LineageEdge struct {
    ID          string    `json:"id" gorm:"primaryKey"`

    // 上游
    SourceType  string    `json:"source_type"  gorm:"index"`  // "step" | "resource" | "outcome"
    SourceID    string    `json:"source_id"    gorm:"index"`  // step_id / resource URI / outcome_id
    SourceField string    `json:"source_field"`               // 具体字段路径，如 "output.form.scenario"

    // 下游
    TargetType  string    `json:"target_type"  gorm:"index"`
    TargetID    string    `json:"target_id"    gorm:"index"`
    TargetField string    `json:"target_field"`               // 如 "input.prompt"

    // 上下文
    InstanceID  string    `json:"instance_id"  gorm:"index"`  // 所属流程实例
    ProcessID   string    `json:"process_id"   gorm:"index"`  // 所属流程定义
    EdgeType    string    `json:"edge_type"`                  // "VAR_REF" | "URI_CONSUME" | "SUBPROCESS" | "EXPLICIT"
    InferredBy  string    `json:"inferred_by"`                // "variable_resolver" | "uri_scanner" | "manual"

    CreatedAt   time.Time `json:"created_at"   gorm:"autoCreateTime"`
}
```

**索引策略：**
- `(instance_id, source_id)` — 查询某实例内某步骤的下游
- `(instance_id, target_id)` — 查询某实例内某步骤的上游
- `(process_id, source_id)` — 跨实例查询某流程模板的血缘模式

### 4.2 自动采集：引擎变量解析时记录

**采集点：** `resolveTemplateVars()` 函数（引擎解析 `{{steps.X.output.Y}}` 时）

```go
// 在 resolveTemplateVars 中，每解析一个变量引用：
func (e *ProcessEngine) resolveTemplateVars(text string, instance *ProcessInstance, currentStepID string) string {
    // ... 现有逻辑 ...

    // 新增：每次解析 {{steps.X.output.Y}} 时记录血缘边
    for _, match := range stepRefPattern.FindAllStringSubmatch(text, -1) {
        refStepID := match[1]
        refField := match[2] // "output" / "data" / "structured_output"
        e.recordLineageEdge(LineageEdge{
            SourceType:  "step",
            SourceID:    refStepID,
            SourceField: refField,
            TargetType:  "step",
            TargetID:    currentStepID,
            TargetField: "input",
            InstanceID:  instance.ID,
            ProcessID:   instance.ProcessID,
            EdgeType:    "VAR_REF",
            InferredBy:  "variable_resolver",
        })
    }
}
```

**性能考虑：** 血缘边写入使用异步批量（channel + 定期 flush），不阻塞主执行路径。

```go
type LineageCollector struct {
    edges chan LineageEdge
    db    *gorm.DB
}

func NewLineageCollector(db *gorm.DB) *LineageCollector {
    c := &LineageCollector{
        edges: make(chan LineageEdge, 1000),
        db:    db,
    }
    go c.flushLoop() // 每秒或攒满 100 条批量写入
    return c
}
```

### 4.3 子流程血缘

当步骤执行器为 `subprocess` 时，记录父子流程之间的血缘关系：

```go
// subprocess executor 启动子流程时
e.lineageCollector.Record(LineageEdge{
    SourceType: "step",
    SourceID:   parentStepID,
    TargetType: "instance",
    TargetID:   childInstanceID,
    EdgeType:   "SUBPROCESS",
    InferredBy: "engine",
    InstanceID: parentInstanceID,
    ProcessID:  parentProcessID,
})
```

### 4.4 血缘查询 API

```
GET /api/lineage/step/{instance_id}/{step_id}/upstream
    → 返回该步骤所有上游数据来源

GET /api/lineage/step/{instance_id}/{step_id}/downstream
    → 返回该步骤的输出被谁消费

GET /api/lineage/instance/{instance_id}/graph
    → 返回整个实例的血缘 DAG（节点 + 边），用于可视化

GET /api/lineage/process/{process_id}/pattern
    → 聚合该流程所有实例的血缘模式（去重后的典型数据流图）
```

**响应格式（graph 接口）：**

```json
{
  "nodes": [
    { "id": "collect_basics", "type": "step", "name": "填写核心需求", "status": "completed" },
    { "id": "generate_spec", "type": "step", "name": "生成规格书", "status": "completed" }
  ],
  "edges": [
    {
      "source": "collect_basics",
      "target": "generate_spec",
      "source_field": "output.form.scenario",
      "target_field": "input.prompt",
      "edge_type": "VAR_REF"
    }
  ]
}
```

### 4.5 文件变更清单

| 文件 | 变更 |
|------|------|
| `internal/acp/entity/lineage.go` | 新增：LineageEdge 实体定义 |
| `internal/acp/service/lineage_service.go` | 新增：LineageCollector + 查询方法 |
| `internal/acp/service/engine_helpers.go` | 修改：resolveTemplateVars 中添加血缘采集调用 |
| `internal/acp/service/subprocess_executor.go` | 修改：启动子流程时记录 SUBPROCESS 边 |
| `internal/acp/handler/handler.go` | 修改：注册血缘查询路由 |
| `cmd/acp/main.go` | 修改：AutoMigrate 添加 LineageEdge；创建 LineageCollector 并注入 |

---

## 五、Phase 1 — 流程健康仪表盘

### 5.1 后端：健康聚合 API

复用现有 `AuditService`（已有 GetSummary/ComputeBottlenecks/ComputeDegradationAlerts），新增以下接口：

```
GET /api/health/processes
    → 所有流程的健康概览（排序：评分最低的排前面）

GET /api/health/processes/{id}
    → 单个流程的详细健康报告

GET /api/health/processes/{id}/steps
    → 该流程每个步骤的聚合指标（平均耗时/失败率/重试率）

GET /api/health/processes/{id}/trend
    → 该流程的健康趋势（日/周粒度，评分 + 成功率 + 平均耗时）
```

**ProcessHealthSummary 响应结构：**

```go
type ProcessHealthSummary struct {
    ProcessID      string  `json:"process_id"`
    ProcessName    string  `json:"process_name"`
    TotalRuns      int     `json:"total_runs"`       // 总运行次数
    SuccessRate    float64 `json:"success_rate"`      // 成功率 0-1
    AvgScore       float64 `json:"avg_score"`         // 平均审计评分 0-100
    AvgDurationMs  int64   `json:"avg_duration_ms"`   // 平均耗时
    P95DurationMs  int64   `json:"p95_duration_ms"`   // P95 耗时
    TotalCostCents int64   `json:"total_cost_cents"`  // 总成本（分）
    AvgTokenUsage  int64   `json:"avg_token_usage"`   // 平均 token 消耗
    LastRunAt      string  `json:"last_run_at"`       // 最后运行时间
    Trend          string  `json:"trend"`             // "improving" | "stable" | "degrading"
    TopBottleneck  string  `json:"top_bottleneck"`    // 最大瓶颈步骤 ID
}
```

**StepHealthDetail 响应结构：**

```go
type StepHealthDetail struct {
    StepID          string  `json:"step_id"`
    StepName        string  `json:"step_name"`
    StepType        string  `json:"step_type"`         // agent/human/control
    Executor        string  `json:"executor"`
    ExecutionCount  int     `json:"execution_count"`
    FailureRate     float64 `json:"failure_rate"`       // 失败率 0-1
    RetryRate       float64 `json:"retry_rate"`         // 重试率 0-1
    AvgDurationMs   int64   `json:"avg_duration_ms"`
    P95DurationMs   int64   `json:"p95_duration_ms"`
    PctOfTotal      float64 `json:"pct_of_total"`       // 占总耗时百分比
    AvgTokenUsage   int64   `json:"avg_token_usage"`
    AvgCostCents    int64   `json:"avg_cost_cents"`
    CommonErrors    []string `json:"common_errors"`     // 最常见错误（top 3）
    DownstreamCount int     `json:"downstream_count"`   // 有多少步骤依赖此步骤的输出
}
```

### 5.2 前端：健康仪表盘 UI

#### 页面一：全局健康概览（/health）

```
┌─────────────────────────────────────────────────────────────────┐
│  流程健康概览                                    [7天 ▼] [刷新]  │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐          │
│  │ 活跃流程  │ │ 总运行次数 │ │ 平均评分  │ │ 平均成功率 │          │
│  │    12    │ │    47    │ │   78.5   │ │   89%    │          │
│  │  ↑2     │ │  ↑12    │ │  ↑3.2   │ │  ↑5%    │          │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘          │
│                                                                 │
│  流程列表                                                        │
│  ┌─────────────┬───────┬──────┬────────┬───────┬───────┬────┐  │
│  │ 流程名称     │ 运行数 │ 成功率 │ 平均评分 │ 平均耗时 │ 趋势   │ 操作 │  │
│  ├─────────────┼───────┼──────┼────────┼───────┼───────┼────┤  │
│  │ ⚠️ 需求收集  │  15   │ 73%  │  62.3  │ 45min │ ↓ 退化 │ 详情 │  │
│  │ ✅ 代码审查  │  22   │ 95%  │  88.7  │ 12min │ ↑ 改善 │ 详情 │  │
│  │ ➡️ 文档创建  │  10   │ 90%  │  79.1  │ 8min  │ → 稳定 │ 详情 │  │
│  └─────────────┴───────┴──────┴────────┴───────┴───────┴────┘  │
│                                                                 │
│  退化告警 (2)                                                    │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │ ⚠️ "需求收集" 成功率从 92% 降至 73%（近7天 vs 前7天）      │    │
│  │ ⚠️ "需求收集" 平均耗时增加 35%（45min → 61min）            │    │
│  └─────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
```

**组件：**
- 顶部指标卡：Ant Design `Statistic` + 环比变化
- 流程列表：`Table` 组件，按评分升序（最差的排前面）
- 退化告警：`Alert` 组件列表，消费 AuditService.ComputeDegradationAlerts()
- 时间范围选择器：`Select`（7天/30天/90天）

#### 页面二：单流程健康详情（/health/processes/:id）

```
┌─────────────────────────────────────────────────────────────────┐
│  ← 需求收集流程 v6                              [查看YAML] [编辑] │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────────────────────────────┐ ┌──────────────────────┐  │
│  │        评分趋势（30天）            │ │    成功率趋势（30天）   │  │
│  │  100 ┤                           │ │  100% ┤              │  │
│  │   80 ┤  ╭─╮  ╭──╮               │ │   80% ┤ ───╮ ╭──     │  │
│  │   60 ┤──╯ ╰──╯  ╰──╮            │ │   60% ┤    ╰─╯       │  │
│  │   40 ┤              ╰──          │ │   40% ┤              │  │
│  │      └────────────────────       │ │       └──────────     │  │
│  │       3/1  3/8  3/15  3/22       │ │                       │  │
│  └──────────────────────────────────┘ └──────────────────────┘  │
│                                                                 │
│  [步骤分析] [运行历史] [血缘图] [迭代记录]                          │
│                                                                 │
│  步骤分析                                                        │
│  ┌──────────────┬──────┬──────┬───────┬────────┬────────────┐  │
│  │ 步骤          │ 失败率 │ 重试率 │ 平均耗时 │ 占比    │ 常见错误    │  │
│  ├──────────────┼──────┼──────┼───────┼────────┼────────────┤  │
│  │ 🔴 生成规格书 │ 25%  │ 40%  │ 18min │ 40%    │ token超限   │  │
│  │ 🟡 审阅循环   │ 10%  │ 20%  │ 15min │ 33%    │ 审批超时    │  │
│  │ 🟢 需求填写   │  2%  │  0%  │  5min │ 11%    │ —          │  │
│  │ 🟢 文档注册   │  0%  │  0%  │  3sec │  0.1%  │ —          │  │
│  └──────────────┴──────┴──────┴───────┴────────┴────────────┘  │
│                                                                 │
│  瓶颈可视化（步骤耗时占比）                                        │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │ 生成规格书  ████████████████████████████████████░░ 40%   │    │
│  │ 审阅循环   ████████████████████████████░░░░░░░░░░ 33%    │    │
│  │ 需求填写   ██████████░░░░░░░░░░░░░░░░░░░░░░░░░░░ 11%    │    │
│  │ 其他      ████████░░░░░░░░░░░░░░░░░░░░░░░░░░░░░ 16%     │    │
│  └─────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
```

**Tab: 血缘图**

使用现有 React Flow (xyflow) 组件渲染：

```
┌─────────────────────────────────────────────────────────────────┐
│  血缘图                                      [实例选择 ▼] [全图] │
│                                                                 │
│   ┌─────────────┐     output.form     ┌─────────────┐          │
│   │ collect_    │ ──────────────────→ │ generate_   │          │
│   │ basics      │    .scenario        │ spec        │          │
│   │ (human)     │    .business_goal   │ (agent)     │          │
│   └─────────────┘    .participants    └──────┬──────┘          │
│                      .acceptance              │                 │
│                                    output.structured_output     │
│                                       .doc_id │                 │
│                                               ↓                 │
│                                      ┌─────────────┐           │
│                                      │ register_   │           │
│                                      │ doc         │           │
│                                      │ (nocodb)    │           │
│                                      └─────────────┘           │
│                                                                 │
│  点击边查看详情：                                                 │
│  ┌──────────────────────────────────────────────────┐           │
│  │ source: collect_basics.output.form.scenario      │           │
│  │ target: generate_spec.input.prompt               │           │
│  │ type: VAR_REF                                    │           │
│  │ 值: "当客户提交需求时自动创建开发任务..."             │           │
│  └──────────────────────────────────────────────────┘           │
└─────────────────────────────────────────────────────────────────┘
```

### 5.3 文件变更清单

| 文件 | 变更 |
|------|------|
| `internal/acp/service/health_service.go` | 新增：ProcessHealthSummary/StepHealthDetail 聚合查询 |
| `internal/acp/handler/handler.go` | 修改：注册 /api/health/* 路由 |
| `acp-web/src/pages/ProcessHealth.tsx` | 新增：全局健康概览页 |
| `acp-web/src/pages/ProcessHealthDetail.tsx` | 新增：单流程健康详情页 |
| `acp-web/src/components/LineageGraph.tsx` | 新增：血缘图组件（基于 React Flow） |
| `acp-web/src/api/health.ts` | 新增：健康 API 客户端 |

---

## 六、Phase 2 — Auto-Create 元流程

### 6.1 流程设计

Auto-Create 本身就是一个 ACP 流程，YAML 定义如下：

```yaml
description: 自动创建业务流程
input_schema:
  fields:
    - key: process_name
      label: 流程名称
      type: text
      required: true
    - key: requirements
      label: 业务需求描述
      type: textarea
      required: true
    - key: complexity
      label: 预估复杂度
      type: select
      options: [simple, medium, complex]

steps:
  # 第一步：Agent 分析需求，生成流程规格书
  - id: analyze_requirements
    name: 分析需求并生成规格书
    executor: agent
    command: chat
    input:
      prompt: |
        你是 ACP 流程架构 Agent。根据以下业务需求设计一个完整的流程规格书。

        ## 需求
        流程名称：{{input.process_name}}
        需求描述：{{input.requirements}}
        复杂度：{{input.complexity}}

        ## 要求
        1. 分析需求，拆解为具体步骤
        2. 确定每个步骤的执行者类型（agent/human/系统集成）
        3. 确定步骤间的依赖关系和数据流
        4. 识别需要人工审批的关键决策点
        5. 设计错误处理和回退策略

        ## 输出格式（structured_output）
        - spec_summary: 规格书摘要（200字内）
        - step_count: 建议步骤数
        - human_steps: 需要人工介入的步骤数
        - risk_factors: 主要风险点列表
      structured_output_schema:
        type: object
        properties:
          spec_summary: { type: string }
          step_count: { type: number }
          human_steps: { type: number }
          risk_factors: { type: string }
        required: [spec_summary, step_count]

  # 第二步：Agent 根据规格书生成可执行 YAML
  - id: generate_yaml
    name: 生成流程 YAML
    executor: agent
    command: chat
    input:
      prompt: |
        根据以下规格书，生成一个完整的 ACP 流程 YAML 定义。

        ## 规格书
        {{steps.analyze_requirements.output.response}}

        ## ACP YAML 规范
        - 每个步骤必须有 id, name, executor, command, input
        - 步骤间依赖用 depends_on 声明
        - 数据传递用 {{steps.xxx.output.yyy}} 模板语法
        - 支持的 executor: agent, human, nocodb, echo, subprocess
        - 支持的控制节点: if, switch, loop, foreach
        - human 步骤需要 form 定义
        - agent 步骤需要 prompt

        ## 要求
        1. YAML 必须语法正确，可以直接发布执行
        2. 每个 agent 步骤的 prompt 要详细、具体
        3. 关键决策点添加 human 审批步骤
        4. 用 structured_output_schema 约束 agent 输出格式

        请直接输出完整 YAML，不要包含 ```yaml 标记。
      structured_output_schema:
        type: object
        properties:
          yaml_content: { type: string, description: "完整的流程 YAML" }
          validation_notes: { type: string, description: "需要人工确认的设计决策" }
        required: [yaml_content]
    depends_on:
      - analyze_requirements

  # 第三步：自动校验 YAML（系统步骤）
  - id: validate_yaml
    name: 校验 YAML
    executor: acp
    command: validate_yaml
    input:
      yaml_content: "{{steps.generate_yaml.output.structured_output.yaml_content}}"
    depends_on:
      - generate_yaml

  # 第四步：人工审阅规格书和 YAML
  - id: human_review
    name: 人工审阅
    executor: human
    command: form
    input:
      title: "审阅自动生成的流程：{{input.process_name}}"
      assignee: ou_e229cd56698a8e15e629af2447a8e0ed
      form:
        fields:
          - key: spec_review
            label: 规格书概要
            type: textarea
            value: "{{steps.analyze_requirements.output.structured_output.spec_summary}}"
          - key: validation_result
            label: YAML 校验结果
            type: textarea
            value: "{{steps.validate_yaml.output.result}}"
          - key: approved
            label: 审批结果
            type: select
            required: true
            options:
              - { value: "approved", label: 通过 }
              - { value: "revision", label: 需要修改 }
              - { value: "rejected", label: 驳回 }
          - key: revision_notes
            label: 修改意见
            type: textarea
    depends_on:
      - validate_yaml

  # 第五步：根据审批结果分支
  - id: review_gate
    name: 审批分支
    control: if
    condition: '{{steps.human_review.output.form.approved}} == "approved"'
    depends_on:
      - human_review
    branches:
      'true':
        # 通过：自动创建流程
        - id: create_process
          name: 创建并发布流程
          executor: acp
          command: create_process
          input:
            name: "{{input.process_name}}"
            yaml_content: "{{steps.generate_yaml.output.structured_output.yaml_content}}"
            auto_publish: true
      'false':
        # 不通过：记录原因
        - id: log_rejection
          name: 记录结果
          executor: echo
          command: echo
          input:
            message: "流程创建未通过审批。原因：{{steps.human_review.output.form.revision_notes}}"
```

### 6.2 新增 Executor：acp（系统内部操作）

```go
// ACP Executor — 执行 ACP 平台内部操作
// 命令：
//   validate_yaml  — 校验流程 YAML（调用 ValidateYAML + ValidateFlow）
//   create_process — 创建流程定义（可选自动发布）
//   get_health     — 获取流程健康数据（供 auto-iterate 使用）
//   get_lineage    — 获取血缘图（供分析用）
type ACPExecutor struct {
    processSvc *ProcessService
    healthSvc  *HealthService
    lineageSvc *LineageService
}
```

### 6.3 文件变更清单

| 文件 | 变更 |
|------|------|
| `internal/acp/service/executor/acp_executor.go` | 新增：ACP 内部操作 executor |
| `internal/acp/service/executor/registry.go` | 修改：注册 acp executor |
| `cmd/acp/main.go` | 修改：创建并注入 ACPExecutor |
| DB seed / 管理界面 | 导入 auto-create 流程 YAML |

---

## 七、Phase 3 — Auto-Iterate 元流程

### 7.1 架构概览

```
┌──────────────┐    健康数据     ┌──────────────┐    YAML diff    ┌──────────────┐
│ Health       │ ──────────→ │ Iterate      │ ────────────→ │ Safety       │
│ Scanner      │             │ Agent        │               │ Validator    │
│ (定时触发)    │             │ (分析+建议)   │               │ (变更分类)    │
└──────────────┘             └──────────────┘               └──────┬───────┘
                                                                   │
                                                    ┌──────────────┼──────────────┐
                                                    ↓              ↓              ↓
                                              ┌──────────┐  ┌──────────┐  ┌──────────┐
                                              │ 自动应用  │  │ 人工审批  │  │ 拒绝     │
                                              │ (低风险)  │  │ (高风险)  │  │ (禁止)   │
                                              └────┬─────┘  └────┬─────┘  └──────────┘
                                                   ↓              ↓
                                              ┌──────────────────────┐
                                              │ Version Manager      │
                                              │ (创建新版本+记录diff) │
                                              └──────────┬───────────┘
                                                         ↓
                                              ┌──────────────────────┐
                                              │ Rollback Monitor     │
                                              │ (自动检测退化→回滚)   │
                                              └──────────────────────┘
```

### 7.2 安全约束体系（核心设计）

```go
// ChangeClassification 对每次迭代变更进行分类
type ChangeClassification struct {
    Type     string // "prompt_tune" | "param_adjust" | "step_reorder" | "step_add" | "step_remove" | "dep_change"
    Risk     string // "low" | "medium" | "high" | "forbidden"
    Approval string // "auto" | "human" | "blocked"
}

// 分类规则
var changeRiskRules = map[string]string{
    "prompt_tune":    "low",      // 修改 agent prompt 文本 → 自动应用
    "param_adjust":   "low",      // 修改超时/重试参数 → 自动应用
    "output_schema":  "medium",   // 修改 structured_output_schema → 人工审批
    "step_reorder":   "medium",   // 调整步骤执行顺序 → 人工审批
    "step_add":       "high",     // 新增步骤 → 人工审批
    "step_remove":    "forbidden",// 删除步骤 → 禁止自动（必须人工操作）
    "dep_change":     "high",     // 修改依赖关系 → 人工审批
    "control_change": "forbidden",// 修改控制流结构 → 禁止自动
    "executor_change":"forbidden",// 更换 executor → 禁止自动
}
```

### 7.3 迭代数据模型

```go
// ProcessIteration 记录每次流程迭代
type ProcessIteration struct {
    ID             string    `json:"id" gorm:"primaryKey"`
    ProcessID      string    `json:"process_id" gorm:"index"`
    VersionFrom    int       `json:"version_from"`
    VersionTo      int       `json:"version_to"`

    // 变更内容
    ChangeType     string    `json:"change_type"`     // 变更分类
    ChangeSummary  string    `json:"change_summary"`  // 人类可读的变更描述
    YAMLDiff       string    `json:"yaml_diff"`       // unified diff 格式
    RiskLevel      string    `json:"risk_level"`      // low/medium/high

    // 审批
    ProposedBy     string    `json:"proposed_by"`     // "auto_iterate_agent" | human user
    ProposedAt     time.Time `json:"proposed_at"`
    ApprovedBy     string    `json:"approved_by"`     // 审批人（auto/human ID）
    ApprovedAt     *time.Time `json:"approved_at"`

    // 状态
    Status         string    `json:"status"`          // proposed → approved → applied → monitoring → confirmed | rolled_back

    // 效果评估
    MetricsBefore  string    `json:"metrics_before"`  // JSON: {success_rate, avg_duration, avg_score}
    MetricsAfter   string    `json:"metrics_after"`   // JSON: 同上，应用后采集
    EvalWindowDays int       `json:"eval_window_days"`// 评估窗口（默认7天）
    EvalResult     string    `json:"eval_result"`     // "improved" | "neutral" | "degraded" | "pending"

    CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
}
```

### 7.4 Auto-Iterate 流程 YAML

```yaml
description: 自动迭代优化流程（每周运行一次）
steps:
  # 第一步：扫描所有流程的健康状况
  - id: scan_health
    name: 扫描流程健康
    executor: acp
    command: get_health
    input:
      time_range_days: 7
      min_runs: 3              # 至少跑过3次才有统计意义
      sort_by: score_asc       # 评分最低的排前面

  # 第二步：Agent 分析健康数据，识别优化机会
  - id: analyze_opportunities
    name: 识别优化机会
    executor: agent
    command: chat
    input:
      prompt: |
        你是 ACP 流程优化 Agent。分析以下流程健康数据，识别可以优化的地方。

        ## 健康数据
        {{steps.scan_health.output.health_data}}

        ## 分析要求
        1. 找出失败率最高的步骤，分析失败原因
        2. 找出耗时占比最大的瓶颈步骤
        3. 找出重试率异常的步骤
        4. 检查是否有步骤的输出从未被下游使用（可能冗余）

        ## 输出要求
        对每个发现的优化机会：
        - target_process_id: 目标流程 ID
        - target_step_id: 目标步骤 ID
        - issue_type: 问题类型（high_failure_rate / bottleneck / high_retry / unused_output）
        - severity: 严重程度（critical / warning / info）
        - suggested_action: 建议的优化动作
        - expected_improvement: 预期改善效果

        只输出有明确数据支撑的优化建议，不要猜测。
      structured_output_schema:
        type: object
        properties:
          opportunities:
            type: array
            items:
              type: object
              properties:
                target_process_id: { type: string }
                target_step_id: { type: string }
                issue_type: { type: string }
                severity: { type: string }
                suggested_action: { type: string }
                expected_improvement: { type: string }
          summary: { type: string }
        required: [opportunities, summary]
    depends_on:
      - scan_health

  # 第三步：对每个优化机会生成具体的 YAML 变更
  - id: generate_patches
    name: 生成优化补丁
    executor: agent
    command: chat
    input:
      prompt: |
        根据以下优化机会列表，为每个机会生成具体的 YAML 修改。

        ## 优化机会
        {{steps.analyze_opportunities.output.structured_output.opportunities}}

        ## 要求
        1. 读取目标流程的当前 YAML（通过 acp_get_process 工具）
        2. 生成 unified diff 格式的修改补丁
        3. 对每个补丁标注变更类型：
           - prompt_tune: 仅修改 prompt 文本
           - param_adjust: 修改超时/重试等参数
           - step_add: 新增步骤
           - step_remove: 删除步骤（谨慎！）
           - dep_change: 修改依赖关系
        4. 保守原则：优先选择低风险的 prompt_tune 和 param_adjust

        ## 安全约束
        - 禁止删除 human/approval 类型的步骤
        - 禁止修改 control 节点的结构
        - 禁止更换步骤的 executor
      structured_output_schema:
        type: object
        properties:
          patches:
            type: array
            items:
              type: object
              properties:
                process_id: { type: string }
                step_id: { type: string }
                change_type: { type: string }
                diff: { type: string }
                explanation: { type: string }
        required: [patches]
    depends_on:
      - analyze_opportunities

  # 第四步：安全分类 + 自动/人工分流
  - id: classify_and_route
    name: 安全分类与审批路由
    executor: acp
    command: classify_iterations
    input:
      patches: "{{steps.generate_patches.output.structured_output.patches}}"
    depends_on:
      - generate_patches

  # 第五步：人工审批高风险变更
  - id: human_approval
    name: 审批高风险变更
    executor: human
    command: form
    input:
      title: "流程优化审批（{{steps.classify_and_route.output.high_risk_count}} 项需审批）"
      assignee: ou_e229cd56698a8e15e629af2447a8e0ed
      form:
        fields:
          - key: patches_review
            label: 待审批变更
            type: textarea
            value: "{{steps.classify_and_route.output.review_summary}}"
          - key: approved_ids
            label: 批准的变更ID（逗号分隔，留空=全部拒绝）
            type: text
    depends_on:
      - classify_and_route

  # 第六步：应用已批准的变更
  - id: apply_iterations
    name: 应用迭代变更
    executor: acp
    command: apply_iterations
    input:
      auto_approved: "{{steps.classify_and_route.output.auto_approved}}"
      human_approved: "{{steps.human_approval.output.form.approved_ids}}"
    depends_on:
      - human_approval
```

### 7.5 回滚监控

应用迭代后，系统自动进入监控期（默认 7 天）：

```go
// RollbackMonitor 在迭代应用后自动监控效果
func (s *IterationService) MonitorIteration(iterationID string) {
    iteration := s.GetIteration(iterationID)
    // 等待评估窗口
    after := time.After(time.Duration(iteration.EvalWindowDays) * 24 * time.Hour)

    // 定期采样
    ticker := time.NewTicker(24 * time.Hour)
    for {
        select {
        case <-ticker.C:
            currentMetrics := s.healthSvc.GetProcessHealth(iteration.ProcessID, 1) // 最近1天
            beforeMetrics := parseMetrics(iteration.MetricsBefore)

            // 如果关键指标显著退化（成功率下降>15% 或评分下降>20分），立即回滚
            if currentMetrics.SuccessRate < beforeMetrics.SuccessRate-0.15 ||
                currentMetrics.AvgScore < beforeMetrics.AvgScore-20 {
                s.RollbackIteration(iterationID, "auto_degradation_detected")
                return
            }

        case <-after:
            // 评估窗口结束，对比 before/after 指标
            afterMetrics := s.healthSvc.GetProcessHealth(iteration.ProcessID, iteration.EvalWindowDays)
            s.FinalizeIteration(iterationID, afterMetrics)
            return
        }
    }
}
```

### 7.6 前端：迭代管理 UI

#### 迭代提案审批页（/iterations）

```
┌─────────────────────────────────────────────────────────────────┐
│  流程迭代管理                                    [仅待审批] [全部] │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  待审批 (2)                                                      │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │ 📝 需求收集 v6 → v7                         [审批] [拒绝] │    │
│  │ 变更：prompt_tune (低风险)                                │    │
│  │ 目标：生成规格书步骤 prompt 优化                            │    │
│  │ 原因：该步骤失败率 25%，主要因 token 超限                    │    │
│  │ ┌─ YAML Diff ──────────────────────────────────────┐    │    │
│  │ │ - prompt: |                                      │    │    │
│  │ │ -   你是BitFantasy的流程架构Agent...（2048字）      │    │    │
│  │ │ + prompt: |                                      │    │    │
│  │ │ +   你是流程架构Agent。简明扼要地...（1200字）        │    │    │
│  │ │ +   ## 约束                                      │    │    │
│  │ │ +   - 总输出不超过2000字                           │    │    │
│  │ └──────────────────────────────────────────────────┘    │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                 │
│  监控中 (1)                                                      │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │ ⏳ 代码审查 v3 → v4               剩余 5天     [回滚]    │    │
│  │ 变更：param_adjust (自动应用)                              │    │
│  │ 指标对比：                                                │    │
│  │   成功率: 88% → 92% (+4%) ✅                              │    │
│  │   平均耗时: 15min → 13min (-13%) ✅                       │    │
│  │   评分: 75.2 → 79.8 (+4.6) ✅                             │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                 │
│  历史 (15)                                                      │
│  ┌──────────┬────────┬──────────┬─────────┬────────┬──────┐    │
│  │ 流程      │ 版本    │ 变更类型  │ 风险     │ 结果   │ 效果  │    │
│  ├──────────┼────────┼──────────┼─────────┼────────┼──────┤    │
│  │ 代码审查  │ v2→v3  │ prompt   │ low     │ 已确认 │ +12% │    │
│  │ 需求收集  │ v5→v6  │ param    │ low     │ 已回滚 │ -8%  │    │
│  └──────────┴────────┴──────────┴─────────┴────────┴──────┘    │
└─────────────────────────────────────────────────────────────────┘
```

### 7.7 文件变更清单

| 文件 | 变更 |
|------|------|
| `internal/acp/entity/iteration.go` | 新增：ProcessIteration 实体 |
| `internal/acp/service/iteration_service.go` | 新增：迭代管理（提案/审批/应用/回滚/监控） |
| `internal/acp/service/executor/acp_executor.go` | 修改：新增 classify_iterations / apply_iterations 命令 |
| `internal/acp/handler/handler.go` | 修改：注册 /api/iterations/* 路由 |
| `acp-web/src/pages/Iterations.tsx` | 新增：迭代管理页 |
| `acp-web/src/components/YAMLDiff.tsx` | 新增：YAML Diff 渲染组件 |
| DB seed | 导入 auto-iterate 流程 YAML |

---

## 八、Phase 4 — 业务结果接入与归因

### 8.1 数据模型

```go
// BusinessOutcome 业务结果节点
type BusinessOutcome struct {
    ID          string    `json:"id" gorm:"primaryKey"`
    Name        string    `json:"name"`
    Type        string    `json:"type"`        // REVENUE | EFFICIENCY | QUALITY | GROWTH
    ValueCents  int64     `json:"value_cents"`  // 货币化价值（分）
    OccurredAt  time.Time `json:"occurred_at"`
    TriggeredBy string    `json:"triggered_by"` // "webhook" | "process" | "manual"
    SourceURI   string    `json:"source_uri"`   // 关联资源 URI（可选）
    InstanceID  string    `json:"instance_id"`  // 直接关联的流程实例（可选）
    Metadata    string    `json:"metadata"`     // JSON
    CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// OutcomeAttribution 归因记录
type OutcomeAttribution struct {
    ID                string  `json:"id" gorm:"primaryKey"`
    OutcomeID         string  `json:"outcome_id" gorm:"index"`
    UnitType          string  `json:"unit_type"`          // "step" | "instance" | "process"
    UnitID            string  `json:"unit_id"`
    UnitName          string  `json:"unit_name"`
    AttributionWeight float64 `json:"attribution_weight"` // 0.0 - 1.0
    AttributionMethod string  `json:"attribution_method"` // "direct" | "lineage_1hop" | "lineage_2hop" | "manual"
    PathDescription   string  `json:"path_description"`   // 人类可读的归因路径描述
    CreatedAt         time.Time `json:"created_at" gorm:"autoCreateTime"`
}
```

### 8.2 归因算法 v1（保守版：直接血缘 + 距离衰减）

不做 PageRank，不做 Shapley 值。第一版用最简单的逻辑：

```go
func (s *AttributionService) ComputeAttribution(outcomeID string) []OutcomeAttribution {
    outcome := s.GetOutcome(outcomeID)
    var attributions []OutcomeAttribution

    // 策略一：如果 outcome 直接关联了流程实例
    if outcome.InstanceID != "" {
        instance := s.GetInstance(outcome.InstanceID)
        steps := s.GetInstanceSteps(instance.ID)

        // 简单归因：按步骤耗时占比分配权重
        totalDuration := int64(0)
        for _, step := range steps {
            if step.Status == "completed" {
                totalDuration += step.DurationMs
            }
        }

        for _, step := range steps {
            if step.Status == "completed" && totalDuration > 0 {
                weight := float64(step.DurationMs) / float64(totalDuration)
                attributions = append(attributions, OutcomeAttribution{
                    OutcomeID:         outcomeID,
                    UnitType:          "step",
                    UnitID:            step.StepID,
                    UnitName:          step.Title,
                    AttributionWeight: weight,
                    AttributionMethod: "direct",
                    PathDescription:   fmt.Sprintf("直接执行步骤，耗时占比 %.0f%%", weight*100),
                })
            }
        }
    }

    // 策略二：通过血缘追溯间接关联的实例
    if outcome.SourceURI != "" {
        // 查找消费了 outcome.SourceURI 资源的步骤
        edges := s.lineageSvc.GetUpstream(outcome.SourceURI)
        for i, edge := range edges {
            hop := i + 1
            decay := 1.0 / float64(hop+1) // 距离衰减
            attributions = append(attributions, OutcomeAttribution{
                OutcomeID:         outcomeID,
                UnitType:          edge.SourceType,
                UnitID:            edge.SourceID,
                AttributionWeight: decay,
                AttributionMethod: fmt.Sprintf("lineage_%dhop", hop),
                PathDescription:   fmt.Sprintf("通过血缘图追溯，%d 跳距离", hop),
            })
        }
    }

    // 归一化权重
    normalize(attributions)
    return attributions
}
```

### 8.3 API 设计

```
# 业务结果
POST /api/outcomes                           # 注册业务结果（Webhook）
GET  /api/outcomes                           # 列表
GET  /api/outcomes/{id}                      # 详情
GET  /api/outcomes/{id}/attribution          # 归因分析

# 价值查询
GET  /api/value/process/{id}                 # 某流程的累计价值
GET  /api/value/step/{process_id}/{step_id}  # 某步骤的累计价值
GET  /api/value/ranking                      # 全局价值排名
```

### 8.4 前端：价值视图

#### 业务结果与归因（/outcomes）

```
┌─────────────────────────────────────────────────────────────────┐
│  业务结果                           [+ 手动添加] [Webhook 配置]   │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐                       │
│  │ 本月结果  │ │ 总价值    │ │ 关联流程  │                       │
│  │    8     │ │ ¥325,000 │ │   5/12   │                       │
│  └──────────┘ └──────────┘ └──────────┘                       │
│                                                                 │
│  结果列表                                                       │
│  ┌──────────────────┬───────┬──────────┬─────────┬────────┐    │
│  │ 名称              │ 类型   │ 价值      │ 时间     │ 来源   │    │
│  ├──────────────────┼───────┼──────────┼─────────┼────────┤    │
│  │ 客户A签约         │ 营收   │ ¥150,000 │ 03-10   │ Webhook│    │
│  │ v2.0功能上线      │ 质量   │ ¥80,000  │ 03-08   │ 流程末 │    │
│  │ 运维效率提升30%   │ 效率   │ ¥50,000  │ 03-05   │ 手动   │    │
│  └──────────────────┴───────┴──────────┴─────────┴────────┘    │
│                                                                 │
│  点击"客户A签约"查看归因：                                        │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │  价值瀑布图：客户A签约 ¥150,000                           │    │
│  │                                                         │    │
│  │  销售跟进流程                                             │    │
│  │  ├── 需求对接 (agent)     ¥45,000  ████████████████      │    │
│  │  ├── 方案定制 (agent)     ¥37,500  ████████████          │    │
│  │  ├── 报价审批 (human)     ¥30,000  ██████████            │    │
│  │  ├── 合同生成 (agent)     ¥22,500  ████████              │    │
│  │  └── 客户通知 (echo)      ¥15,000  █████                 │    │
│  └─────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
```

#### 流程价值排名（/value/ranking）

```
┌─────────────────────────────────────────────────────────────────┐
│  价值排名                                          [30天 ▼]     │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────────────┬──────────┬───────┬──────────┬──────────┐ │
│  │ 流程              │ 累计价值  │ 运行数 │ 效率比    │ 趋势     │ │
│  ├──────────────────┼──────────┼───────┼──────────┼──────────┤ │
│  │ 1. 销售跟进       │ ¥180,000 │  12   │ 15.0     │ ↑        │ │
│  │ 2. 功能开发       │ ¥95,000  │   8   │ 8.2      │ →        │ │
│  │ 3. 客户成功       │ ¥50,000  │  15   │ 5.0      │ ↑        │ │
│  │ ...              │          │       │          │          │ │
│  │ 10. 内部周报      │ ¥0       │  22   │ 0.0      │ →        │ │
│  │ 11. 测试数据清理  │ ¥0       │   5   │ 0.0      │ →        │ │
│  └──────────────────┴──────────┴───────┴──────────┴──────────┘ │
│                                                                 │
│  效率比 = 累计价值 / 累计成本                                     │
│  ¥0 表示该流程尚未关联到任何业务结果，不代表无价值                   │
└─────────────────────────────────────────────────────────────────┘
```

### 8.5 文件变更清单

| 文件 | 变更 |
|------|------|
| `internal/acp/entity/outcome.go` | 新增：BusinessOutcome + OutcomeAttribution 实体 |
| `internal/acp/service/outcome_service.go` | 新增：结果管理 + Webhook 接入 |
| `internal/acp/service/attribution_service.go` | 新增：归因算法 v1 |
| `internal/acp/handler/handler.go` | 修改：注册 /api/outcomes/* 和 /api/value/* 路由 |
| `acp-web/src/pages/Outcomes.tsx` | 新增：业务结果管理页 |
| `acp-web/src/pages/ValueRanking.tsx` | 新增：价值排名页 |
| `acp-web/src/components/ValueWaterfall.tsx` | 新增：价值瀑布图组件 |

---

## 九、前端路由总览

```
/health                         → 全局流程健康概览      (P1)
/health/processes/:id           → 单流程健康详情         (P1)
/health/processes/:id/lineage   → 血缘图               (P0+P1)
/iterations                     → 迭代管理（提案/审批）   (P3)
/iterations/:id                 → 单次迭代详情           (P3)
/outcomes                       → 业务结果管理           (P4)
/outcomes/:id                   → 结果详情+归因          (P4)
/value/ranking                  → 价值排名               (P4)
```

**导航集成：** 在现有 ACP 左侧导航栏新增分组：

```
📊 评估与优化
  ├── 流程健康        (P1)
  ├── 迭代管理        (P3)
  ├── 业务结果        (P4)
  └── 价值排名        (P4)
```

---

## 十、技术风险与缓解

| 风险 | 概率 | 影响 | 缓解措施 |
|------|------|------|---------|
| Agent 生成的 YAML 语法错误 | 高 | 低 | ValidateYAML + ValidateFlow 双重校验，已有 |
| Agent 迭代导致流程退化 | 中 | 高 | 安全分类 + 人工审批 + 自动回滚监控 |
| 血缘采集影响引擎性能 | 低 | 中 | 异步批量写入，不阻塞主路径 |
| 归因算法结果不准确 | 高 | 低 | v1 用最简单逻辑，标注"近似值"，后续迭代 |
| 评估窗口内样本不足 | 高 | 中 | 设最小运行次数阈值（默认 3 次） |

---

## 十一、开放问题（留给后续版本）

1. **评估窗口自适应**：高频流程（日跑 10 次）用 3 天窗口，低频流程（周跑 1 次）用 30 天窗口
2. **多版本灰度**：同一流程的新旧版本按比例分流（需要引擎层支持）
3. **跨流程归因**：流程 A 的输出被流程 B 消费，归因如何跨流程传播
4. **LLM Judge**：对无结构化产出（纯文本）用另一个 Agent 评估质量
5. **元流程的自我迭代**：Auto-Iterate 流程本身也可以被 Auto-Iterate 优化（递归自举）

---

*本文档基于 v1.0 理论框架 + 现有 ACP 审计基础设施 + 三轮讨论结论，面向可落地实现。*
