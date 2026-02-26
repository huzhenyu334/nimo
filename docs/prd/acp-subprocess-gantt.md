# PRD: ACP 子流程 + 时间管理 + 甘特图

## 1. 背景与目标

### 背景

ACP当前的Process是扁平的Step序列，缺少：
1. **子流程**：无法将复杂流程分层嵌套
2. **时间管理**：Step没有计划工期、截止日期概念
3. **全局视图**：缺乏甘特图等项目管理视角

在PLM场景中，产品开发有清晰的阶段（EVT→DVT→PVT→MP），每个阶段内有多个任务，项目经理需要看到全局甘特图掌控进度。

如果ACP要替代PLM的工作流引擎，必须补齐这些能力。

### 目标

1. 支持**子流程嵌套**：Step可以调用另一个完整的Process
2. 支持**时间属性**：Step可选配置计划工期和截止日期
3. 支持**甘特图视图**：在Instance详情页展示时间线视图
4. 支持**Agent流程 + 企业流程**的分层架构

### 核心设计理念

**Agent流程（Agent Process）**：每个Agent维护自己的工作流程，描述自己如何完成某类工作。
- PM Agent: "需求分析流程"、"PRD撰写流程"
- QA Agent: "代码审查流程"、"测试执行流程"
- DevOps Agent: "构建部署流程"

**企业流程（Enterprise Process）**：把多个Agent流程 + 人类任务编排在一起，形成端到端的业务流程。
- "产品开发流程" = PM需求分析 → 硬件设计（人）→ QA审查 → 打样（人）→ 测试
- "采购流程" = PM询价分析 → 供应商选择（人审批）→ 下单 → 入库（人确认）

**好处**：
- Agent流程可独立迭代，不影响企业流程
- Agent流程可被多个企业流程复用
- 企业流程只关心"调谁做什么"，不关心Agent内部怎么做
- 人类任务和Agent流程平等混排

## 2. YAML 设计

### 2.1 Step 新增属性

```yaml
steps:
  - name: step_name
    # 现有属性保持不变: prompt, gate, depends_on, etc.
    
    # ── 责任人（统一设计，取代旧的 agent 和 type:human）──
    assignee_type: agent          # agent | user | role
    assignee_id: "pm"             # agent ID / 飞书open_id / 角色名
    # 引擎根据 assignee_type 自动决定执行方式：
    #   agent → 调agent执行prompt
    #   user  → 创建人类任务，等待完成
    #   role  → 按角色匹配负责人
    
    # ── 子流程引用 ──
    type: subprocess              # Step类型只管结构：subprocess/expand/condition/approval
    process: "process-name"       # 引用的Process定义名称（或ID）
    input: "{{input}}"            # 传给子流程的输入
    
    # ── 内联子步骤（Step嵌套）──
    steps:                        # 子步骤列表，递归结构
      - name: child_step_1
        assignee_type: agent
        assignee_id: pm
        prompt: "..."
      - name: child_step_2
        assignee_type: user
        assignee_id: "ou_xxx"
        title: "人类任务"
    
    # ── 时间属性（所有Step可选）──
    title: "步骤标题"              # 所有步骤都可以有标题
    duration: "5d"                # 计划工期（支持: 30m, 2h, 1d, 2w）
    due: "+5d"                    # 截止时间
```

### 2.1.0 责任人设计（assignee_type + assignee_id）

**核心理念**：Step类型管结构，责任人管执行。不再用 `type: human` 或 `agent:` 字段区分执行者。

| assignee_type | assignee_id 格式 | 引擎行为 |
|---------------|-----------------|---------|
| `agent` | Agent ID（如 `pm`, `qa`） | 调用Agent执行prompt，等待回复 |
| `user` | 飞书open_id（如 `ou_xxx`） | 创建人类任务，发通知，等待人类完成 |
| `role` | 角色名（如 `reviewer`） | 按角色匹配负责人（未来扩展） |

**向后兼容**：现有的 `agent: pm` 写法继续支持，引擎内部自动转换为 `assignee_type: agent, assignee_id: pm`。

**Step type 简化为纯结构类型**：

| type | 含义 |
|------|------|
| _(空/默认)_ | 普通任务步骤（看assignee决定谁做） |
| `subprocess` | 引用另一个独立流程 |
| `expand` | 动态展开（遍历列表） |
| `condition` | 条件分支 |
| `approval` | 审批流程 |

### 2.1.1 三种流程组织方式

| 方式 | YAML | 适用场景 | 例子 |
|------|------|---------|------|
| **内联子步骤** | `steps: [...]` | 只在当前流程用的分组，设计时已确定的层级 | EVT阶段里的具体任务 |
| **子流程引用** | `type: subprocess` + `process: "xxx"` | 可复用的独立流程，被多个流程引用 | "测试套件"被EVT/DVT/PVT复用 |
| **扁平步骤** | 无嵌套 | 简单流程 | 3步走的小任务 |

**内联子步骤 vs 子流程引用的选择标准**：
- 这组步骤会被其他流程复用吗？→ **子流程引用**
- 只是当前流程内部的分组？→ **内联子步骤**
- 内联子步骤内也可以引用子流程（混合使用）

**内联子步骤执行规则**：
- 父Step有`steps`时，引擎按顺序执行子步骤（子步骤之间也支持`depends_on`并行）
- 所有子步骤完成 → 父Step标记completed，output为最后一个子步骤的output
- 任一子步骤失败 → 父Step标记failed（遵循子步骤的on_failure策略）
- 最大嵌套深度：5层（与子流程共享深度计数）
- 子步骤内还可以有子步骤（递归），形成任意深度的WBS树

### 2.2 Duration 格式

**YAML输入**：支持人类友好的时间表达：
- `30m` — 30分钟
- `2h` — 2小时
- `1d` — 1自然日（24h）
- `3d` — 3自然日
- `2w` — 2周（14自然日）

**存储**：引擎解析时立即转为毫秒数值（`duration_ms`），同时保留原始字符串（`duration_raw`）。
- YAML输入 `"5d"` → 存储 `duration_ms: 432000000` + `duration_raw: "5d"`
- 统计/排序/比较全部使用 `duration_ms` 数值字段，避免运行时parse
- 前端展示时从ms反向格式化为人类可读形式

Agent step通常不需要设置duration，系统自动记录实际耗时。

### 2.3 完整示例：产品开发流程（推荐写法）

```yaml
# ═══ 产品开发全流程 ═══
# 展示：内联子步骤 + 子流程引用 + agent/user混排
name: "product-development"
description: "智能眼镜产品开发全流程"

steps:
  - name: requirement
    title: "需求阶段"
    duration: "3d"
    steps:
      - name: market_research
        assignee_type: agent
        assignee_id: pm
        prompt: "调研市场需求和竞品: {{input}}"
      - name: write_prd
        assignee_type: agent
        assignee_id: pm
        prompt: "撰写PRD: {{market_research.output}}"
      - name: prd_review
        assignee_type: user
        assignee_id: "ou_xxx"          # CEO的飞书open_id
        title: "需求评审与批准"
        duration: "1d"

  - name: evt
    title: "EVT阶段"
    duration: "30d"
    steps:
      - name: plan
        assignee_type: agent
        assignee_id: pm
        prompt: "制定EVT阶段计划: {{requirement.output}}"
      - name: schematic
        assignee_type: user
        assignee_id: "ou_hw_001"       # 硬件工程师
        title: "原理图设计"
        duration: "10d"
      - name: schematic_review
        assignee_type: agent
        assignee_id: qa
        prompt: "审查原理图设计文件"
      - name: pcb_layout
        assignee_type: user
        assignee_id: "ou_hw_001"
        title: "PCB Layout"
        duration: "7d"
        depends_on: [schematic_review]
      - name: prototype
        assignee_type: user
        assignee_id: "ou_hw_001"
        title: "打样与组装"
        duration: "10d"
      - name: testing
        type: subprocess               # 引用可复用的测试流程
        process: "full-test-suite"
      - name: evt_report
        assignee_type: agent
        assignee_id: pm
        prompt: "汇总EVT结果: {{testing.output}}"

  - name: dvt
    title: "DVT阶段"
    duration: "45d"
    steps:
      - name: mold_design
        assignee_type: user
        assignee_id: "ou_me_001"       # 结构工程师
        title: "模具设计"
        duration: "15d"
      - name: mold_making
        assignee_type: user
        assignee_id: "ou_supplier_001" # 供应商
        title: "开模制造"
        duration: "25d"
      - name: dvt_test
        type: subprocess
        process: "full-test-suite"
```

```yaml
# ═══ 可复用的Agent流程：完整测试套件 ═══
name: "full-test-suite"
description: "标准化测试流程，被多个阶段复用"
owner: "qa"

steps:
  - name: test_plan
    assignee_type: agent
    assignee_id: qa
    prompt: "制定测试计划: {{input}}"
  - name: auto_test
    assignee_type: agent
    assignee_id: qa
    prompt: "执行自动化测试"
  - name: manual_test
    assignee_type: user
    assignee_id: "ou_test_001"        # 测试工程师
    title: "手动测试项"
    duration: "3d"
  - name: test_report
    assignee_type: agent
    assignee_id: qa
    prompt: "汇总测试结果: {{auto_test.output}} {{manual_test.output}}"
```

## 3. 数据模型

### 3.1 Process 定义扩展

```go
// Process定义新增字段
type ProcessDefinition struct {
    // ...现有字段...
    Owner       string `json:"owner,omitempty"`        // 归属Agent ID（Agent流程）
    Category    string `json:"category,omitempty"`     // "agent" | "enterprise" | ""
    Description string `json:"description,omitempty"`
}
```

### 3.2 Step 定义扩展

```go
type StepDefinition struct {
    // ...现有字段...
    Type         string           `yaml:"type,omitempty"`          // subprocess | expand | condition | approval | ""
    ProcessRef   string           `yaml:"process,omitempty"`      // subprocess引用的Process名称
    Steps        []StepDefinition `yaml:"steps,omitempty"`        // 内联子步骤（递归WBS树）
    AssigneeType string           `yaml:"assignee_type,omitempty"`// agent | user | role
    AssigneeID   string           `yaml:"assignee_id,omitempty"` // Agent ID / 飞书open_id / 角色名
    Title        string           `yaml:"title,omitempty"`        // 步骤标题
    Description  string           `yaml:"description,omitempty"`  // 步骤描述
    Duration     string           `yaml:"duration,omitempty"`     // 计划工期 "5d"
    Due          string           `yaml:"due,omitempty"`          // 截止时间 "+5d" 或 "2026-03-15"
}
```

### 3.3 Instance 扩展

```go
type ProcessInstance struct {
    // ...现有字段...
    ParentInstanceID string     `json:"parent_instance_id,omitempty"`  // 父流程实例ID
    ParentStepName   string     `json:"parent_step_name,omitempty"`    // 父流程中的step名
    PlannedStart     *time.Time `json:"planned_start,omitempty"`       // 计划开始时间
    PlannedEnd       *time.Time `json:"planned_end,omitempty"`         // 计划结束时间
}
```

### 3.4 StepExecution 扩展

```go
type StepExecution struct {
    // ...现有字段...
    Type            string     `json:"type,omitempty"`              // step类型
    AssigneeType    string     `json:"assignee_type,omitempty"`     // agent | user | role
    AssigneeID      string     `json:"assignee_id,omitempty"`       // 责任人ID
    Title           string     `json:"title,omitempty"`             // 步骤标题
    PlannedDuration int64      `json:"planned_duration_ms,omitempty"` // 计划工期(毫秒)
    DueAt           *time.Time `json:"due_at,omitempty"`            // 截止时间
    ChildInstanceID string     `json:"child_instance_id,omitempty"` // 子流程实例ID
}
```

## 4. 引擎行为

### 4.0 内联子步骤执行逻辑

```
引擎执行到有 steps 子步骤的 Step:
  1. 父Step状态 → "running"
  2. 按顺序执行子步骤（子步骤之间支持 depends_on 并行）
  3. 子步骤可以是任何类型：normal/human/subprocess/expand/approval/condition
  4. 子步骤内还可以有子步骤（递归），与子流程共享5层深度限制
  5. 所有子步骤完成：
     - 最后一个子步骤的output → 父Step output
     - 父Step状态 → "completed"
  6. 任一子步骤失败：
     - 遵循子步骤自身的 on_failure 策略
     - 如果子步骤 on_failure=abort → 父Step状态 → "failed"
  7. 变量引用：
     - 子步骤之间可以用 {{sibling_step.output}} 互相引用
     - 子步骤可以用 {{parent_step.output}} 引用同流程内其他顶层步骤
     - 外部步骤引用父Step时，得到的是父Step的聚合output
```

### 4.1 Subprocess Step 执行逻辑

```
父流程执行到 subprocess step:
  1. 查找引用的Process定义（by name or ID）
  2. 创建子Instance（设置parent_instance_id, parent_step_name）
  3. 父step状态 → "waiting"
  4. 子Instance开始独立执行（复用现有引擎）
  5. 子Instance所有step完成：
     - 子Instance output → 父step output
     - 父step状态 → "completed"
     - 触发父流程继续执行下一步
  6. 子Instance失败：
     - 父step状态 → "failed"
     - 父流程暂停，支持重试（重试 = 重新创建子Instance）
```

### 4.2 基于 assignee_type 的执行分发

```
引擎执行普通Step时，根据 assignee_type 分发：

【assignee_type = agent】（默认，兼容旧的 agent: xxx 写法）
  1. 向Agent发送prompt
  2. 等待Agent回复（轮询）
  3. 收到回复 → Step "completed"

【assignee_type = user】
  1. 创建"人类任务"记录（复用现有Task系统）
  2. 通知负责人（飞书推送，按assignee_id查找用户）
  3. Step状态 → "waiting_human"
  4. 人类通过ACP前端或飞书完成任务
  5. Step状态 → "completed"
  6. 超期 → 催办通知 + 前端标红

【assignee_type = role】（未来扩展）
  1. 按角色匹配可用负责人
  2. 根据匹配到的是agent还是user，走对应逻辑

【向后兼容】
  - 旧写法 `agent: pm` → 引擎内部转为 assignee_type=agent, assignee_id=pm
  - 无assignee_type且无agent字段 → 报错（必须指定执行者）
```

### 4.3 时间计算

Instance启动时自动计算时间线：

```go
func calculateTimeline(instance *Instance, steps []StepDefinition) {
    cursor := instance.StartedAt
    for _, step := range steps {
        step.PlannedStart = cursor
        if step.Duration != "" {
            d := parseDuration(step.Duration) // "5d" → 5 * 8h
            step.PlannedEnd = cursor.Add(d)
            cursor = step.PlannedEnd
        } else if step is agent type {
            // Agent step默认预估5分钟
            step.PlannedEnd = cursor.Add(5 * time.Minute)
            cursor = step.PlannedEnd
        }
    }
    instance.PlannedEnd = cursor
}
```

注意：并行step（depends_on）的时间计算需要用关键路径算法。

## 5. API 设计

### 5.1 新增/扩展接口

```
# 子流程相关
GET  /api/instances/:id/children          # 获取子流程实例列表
GET  /api/instances/:id/parent            # 获取父流程实例
GET  /api/instances/:id/timeline          # 获取甘特图数据（含子流程展开）

# 人类任务相关
GET  /api/my-tasks?assignee=xxx           # 获取我的待办任务
POST /api/instances/:id/steps/:name/complete-human  # 人类完成任务

# 流程分类
GET  /api/processes?category=agent        # 获取Agent流程
GET  /api/processes?category=enterprise   # 获取企业流程
GET  /api/processes?owner=pm              # 获取PM的Agent流程
```

### 5.2 甘特图数据接口

```
GET /api/instances/:id/timeline
```

返回：
```json
{
  "instance_id": "xxx",
  "name": "product-development",
  "planned_start": "2026-02-25T00:00:00+08:00",
  "planned_end": "2026-05-15T00:00:00+08:00",
  "items": [
    {
      "step_name": "requirement",
      "type": "subprocess",
      "title": "需求分析",
      "planned_start": "2026-02-25",
      "planned_end": "2026-02-27",
      "actual_start": "2026-02-25T10:00:00+08:00",
      "actual_end": "2026-02-25T10:02:00+08:00",
      "status": "completed",
      "progress": 100,
      "children": [
        {
          "step_name": "market_research",
          "type": "agent",
          "title": "市场调研",
          "agent": "pm",
          "actual_start": "...",
          "actual_end": "...",
          "status": "completed",
          "duration_ms": 45000
        },
        ...
      ]
    },
    {
      "step_name": "evt",
      "type": "subprocess",
      "title": "EVT阶段",
      "planned_start": "2026-02-28",
      "planned_end": "2026-03-30",
      "actual_start": "2026-02-28T09:00:00+08:00",
      "actual_end": null,
      "status": "running",
      "progress": 35,
      "overdue": false,
      "children": [...]
    },
    {
      "step_name": "schematic",
      "type": "human",
      "title": "原理图设计",
      "assignee": "张工",
      "planned_start": "2026-03-01",
      "planned_end": "2026-03-11",
      "due_at": "2026-03-11T18:00:00+08:00",
      "status": "running",
      "overdue": true,           // 超期标记
      "overdue_days": 2
    }
  ]
}
```

## 6. 前端设计

### 6.1 甘特图视图

在Instance详情页新增"时间线"Tab，使用甘特图展示：

- **横轴**：时间（自动适配粒度：小时/天/周/月）
- **纵轴**：Steps（支持折叠子流程）
- **色条**：
  - 🟢 绿色：已完成
  - 🔵 蓝色：进行中
  - ⚪ 灰色：未开始（planned）
  - 🔴 红色：超期
- **条形最小宽度规则**：
  - 所有步骤统一使用条形显示（包括Agent步骤）
  - **实际耗时低于1小时的步骤，按1小时宽度显示**（`displayDuration = max(actualDuration, 1h)`）
  - 这解决了Agent步骤（秒级）和人类步骤（天级）在同一甘特图上的尺度差异问题
  - Hover时显示真实耗时（如"32秒"），条形宽度仅用于视觉可读性
- **交互**：
  - 点击色条 → 查看Step详情
  - 折叠/展开子流程
  - Hover显示真实时间信息（实际耗时、开始/结束时间）

推荐方案：使用 `frappe-gantt` 或手绘SVG（更灵活），避免重量级库。

### 6.2 流程分类视图

流程列表页增加分类Tab：
- **全部** | **企业流程** | **Agent流程**
- Agent流程显示归属Agent标签
- 企业流程显示包含的子流程引用

### 6.3 人类任务看板

新增"我的任务"页面（或Dashboard卡片）：
- 待办任务列表（按截止日期排序）
- 超期任务高亮
- 快捷完成入口

## 7. 实现路线

### Phase 1: 基础子流程 + 内联子步骤（MVP）
- Step `type: subprocess` 支持（引用独立流程）
- Step `steps` 内联子步骤支持（递归WBS树）
- 子Instance创建和完成回调
- 父子Instance关联查询
- 引擎递归执行内联子步骤
- 前端：子流程/子步骤折叠展开查看

### Phase 2: 时间管理
- Step `duration`/`due` 属性解析
- 时间线自动计算
- 超期检测和通知
- 甘特图视图（基础版）

### Phase 3: 人类任务
- Step `type: human` 支持
- 人类任务通知（飞书集成）
- 人类任务完成接口
- "我的任务"看板

### Phase 4: Agent流程生态
- 流程 `owner`/`category` 分类
- Agent自主创建和管理自己的流程
- 流程市场/复用机制
- 企业流程可视化编排（拖拽式？）

## 8. 注意事项

1. **向后兼容**：现有的Process定义无需修改即可继续运行。所有新字段都是可选的。
2. **无限嵌套防护**：子流程不允许调用自身或形成循环引用。引擎启动子流程前检查。最大嵌套深度建议限制为5层。
3. **DB可选原则**：子流程的父子关系可以用Instance字段存储（parent_instance_id），不需要新建关联表。时间属性直接存在Step/Instance记录中。
4. **Agent流程归属**：`owner`字段标记归属Agent，但执行时仍然通过正常的Agent调度机制。Agent可以通过ACP知识库维护和更新自己的流程定义。
