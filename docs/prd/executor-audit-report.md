# ACP Executor架构 PRD vs 实现审计报告

> 审计日期：2026-03-02（v3） | 审计人：Lyra
> 重要更新：反映v4.3两层Output + v4.4三层Schema（砍Config）+ v4.5 Event Sourcing数据层 + data_schema in input

## 总览

| 指标 | 值 |
|------|-----|
| PRD版本 | v2.0（含v4.1~v4.5补充） |
| 代码最新扫描 | 基于当前main分支 |
| EngineAdapter方法数 | 12个签名方法（精简后） |
| workflow_engine.go行数 | 5545 |
| executor/总行数 | 2400（agent:768 + human:881 + subprocess:290 + executor:352 + deps:109） |
| 注册的Executor类型 | 3个（agent, human, subprocess） |
| StepDef字段 | `Executor` + `Control` 正交字段已实现 |
| v4.3两层Output | ❌ 未实现（代码无标准信封/业务Data分层） |
| v4.4三层Schema | ❌ 未实现（InputType/OutputType/TraceType均无，ConfigType已砍掉） |
| v4.5 Event Sourcing | ❌ 未实现（无executor_trace表、无Storage接口、无流式写入） |

## v4.3~v4.5核心决策摘要

**v4.3 两层Output模型（泽斌 2026-03-02确认）：**

| 层 | 名称 | 定义方 | 结构 | 举例 |
|---|---|---|---|---|
| **标准信封** | Executor Output | OutputType() 反射 | 固定，executor级统一 | agent: token_usage, session_key, model; shell: exit_code, duration |
| **业务Data** | Step Data | YAML output_schema | 可变，每step不同 | 写PRD: prd_content; 代码审查: passed, issues |

同一个AgentExecutor的标准信封结构统一（都有token_usage, session_key），但不同step的data字段由output_schema约束（写PRD输出prd_content，代码审查输出passed/issues）。

**v4.4 三层Schema（泽斌 2026-03-02确认砍掉Config层）：**
- 只有 InputType() / OutputType() / TraceType()，不再有 ConfigType()
- 所有"配置"都是input的一部分（model、timeout等），系统级配置走环境变量
- data_schema放在input里（executor的输入参数），引擎通过InputType()声明知道去input里读

**v4.5 Event Sourcing数据层（泽斌 2026-03-02确认）：**
- Data不是大JSON对象，是有序条目流（`executor_trace`表）
- Storage接口：Append(流式写入) / List(分页读取) / Search(搜索)
- 审计 = 三层Schema的自然副产品，不需要FetchDetail/AuditDetail/Timeline等独立抽象
- 审计核心 = executor完整私有数据（agent=完整session含system prompt）
- 实时审计天然支持（agent还在跑就能看已有对话）

**对代码的影响**：
1. 需要新建 `executor_trace` 表（instance_id, step_id, seq, entry_type, content, summary, created_at）
2. 需要实现 Storage 接口（Append/List/Search）
3. 三个executor都需要在执行过程中调用Storage.Append()写入数据条目
4. tasks表继续存标准信封（OutputType()），executor_trace表存完整私有数据（TraceType()）

---

## 逐章节审计

---

### §1 核心洞察 — 统一计算单元模型

**PRD要求：**
- Executor = f(input) → output，所有能接收输入产出输出的都是执行器
- 消除引擎中按type做switch/case的枚举思维
- 引擎只做调度：`registry.Get(step.Executor) → executor.Dispatch → executor.Wait`

**实现状态：** ⚠️ 部分实现

**差异说明：**
- ✅ Executor接口已定义（`Executor` + `StepExecutor`），Registry已实现，`dispatchStep` 统一路由
- ✅ 三个IO-bound executor（agent/human/subprocess）已抽取为独立struct，通过registry注册
- ⚠️ 引擎中仍有大量按类型分支的逻辑：
  - `workflow_engine.go:2179` — `if step.Executor == "approval"` 特判，路由到 `dispatchStep(ctx, "approval", ...)`
  - `workflow_engine.go:2185` — `if step.Executor == "subprocess"` 特判
  - `workflow_engine.go:2259~2489` — 大量 `if step.AssigneeType() == "human"` 分支处理不同assignee场景（candidates/rules/relation/role/direct）
  - 行2489 — 默认路径 `dispatchStep(ctx, "agent", ...)`
- ⚠️ 旧的 `executeStepAttempt` 函数（定义在第2519行）仍存在于engine中，有3处活跃引用：
  - 第4394行：expand内联执行中for_each子步骤的agent执行
  - 第4481行：retryStep中agent步骤的重试执行
  - 第4580行：expand内联执行中escalation路径的agent重试
  这些路径绕过了executor registry，直接走engine内部的旧逻辑
- ❌ PRD设想的 `executor.Dispatch(ctx, task) → handle` + `executor.Wait(ctx, handle) → result` 模式未实现。实际接口是 `ExecuteStep(ctx, adapter, rc) error`——executor自己管全生命周期

---

### §2 两个正交维度：Executor × Control

**PRD要求：**
- 有IO → Executor（agent, human, subprocess, http, shell）
- 无IO → Control（condition, gate, expand）
- YAML用 `executor:` 和 `control:` 两个互斥字段
- subprocess应归类为Executor（有IO），不是Control

**实现状态：** ✅ 已实现

**差异说明：**
- ✅ `StepDef` 已有 `Executor string` 和 `Control string` 两个正交字段
- ✅ subprocess归类为Executor（`SubprocessExecutor` 实现 `StepExecutor` 接口，290行）
- ✅ condition和expand作为Control节点在引擎内部inline执行（`executeConditionInline`、`executeExpandInline`），不走executor registry
- ✅ 未知control类型有清晰错误处理（`workflow_engine.go:2210`：`"unknown control type: %s"`）
- ⚠️ `gate` 作为Control类型在StepDef的enum里声明了，但实际上gate是step-level attribute（`step.Gate`），不是独立的control节点。PRD §2.3列gate为Control，但代码中gate是agent/human步骤的附属配置。gate检查逻辑在 `executor.CheckGate()` 中实现，由AgentExecutor在step完成后调用

---

### §3 type字段消失

**PRD要求：**
- 不再需要 `type: agent/human/approval`
- 统一用 `executor: openclaw_agent / claude_code / feishu_approval / human / shell / http`
- 六种executor用同一YAML语法

**实现状态：** ⚠️ 部分实现

**差异说明：**
- ✅ StepDef中 `type` 字段已消失，换成 `Executor` 和 `Control` 字段
- ✅ YAML用 `executor: openclaw_agent / human / approval / subprocess` + `control: condition / expand`
- ❌ PRD设想的 `executor: claude_code / feishu_approval / feishu_message / http / shell` 均未实现
- 🐛 **Approval路由确认Bug**：engine传 `execType="approval"` 给 `dispatchStep`（第2179行），但registry只注册了三个executor：
  ```go
  // initExecutors() 第224-229行
  e.registry.Register(&executor.SubprocessExecutor{Deps: deps})  // Type()="subprocess"
  e.registry.Register(&executor.AgentExecutor{Deps: deps})       // Type()="agent"
  e.registry.Register(&executor.HumanExecutor{Deps: deps})       // Type()="human"
  ```
  `registry.GetStep("approval")` 返回nil。`dispatchStep` 走到 `se == nil` 分支（第588行），将step标记为failed。
  
  **修复方案**：在 `initExecutors()` 中额外注册 `"approval"` 别名指向HumanExecutor（推荐），或在engine中approval特判后改为 `dispatchStep(ctx, "human", ...)`。

  **注意**：retryStep路径（第4440行）也有同样问题——`dispatchStep(ctx, "approval", ...)`。

---

### §4 业界调研

**PRD要求：** 调研Kestra/Temporal/Windmill/Hatchet等，对比统一度

**实现状态：** ✅ 已实现（PRD文档层面）

**差异说明：** 这是设计文档章节，不涉及代码实现。调研结论已纳入架构决策。ACP对标Kestra的Flowable vs Runnable分类（✅ 已体现为Control vs Executor），但ACP在AI Agent场景的差异化（非确定性执行器统一编排）尚需更多executor类型落地才能体现。

---

### §5 统一执行器协议

**PRD要求（v4.2原始 + v4.3两层Output更新）：**
```go
// v4.4 接口（三层Schema，Config层已砍掉）
type Executor interface {
    Type() string
    InputType()  reflect.Type   // 三层Schema：输入参数（含data_schema）
    OutputType() reflect.Type   // 标准输出信封
    TraceType()   reflect.Type   // 私有持久化数据（无状态executor返回nil）
    Execute(ctx context.Trace, task *Task) (*Result, error)
}

type ControlExecutor interface {
    Type() string
    ExecuteControl(ctx context.Trace, rc *RunTrace, adapter EngineAdapter) error
}
```

v4.3/v4.4新增决策：
- **data_schema放在input里**（不是step顶层字段）——它是executor的输入参数，引擎通过InputType()声明知道去input里读
- `OutputType()` 返回的是 **executor级标准信封**（固定结构）
  - AgentExecutor的OutputType → `AgentOutput{TokenUsage, SessionKey, Model, Duration}`
  - ShellExecutor的OutputType → `ShellOutput{ExitCode, Duration, StdoutBytes}`
  - HTTPExecutor的OutputType → `HTTPOutput{StatusCode, LatencyMs}`
- YAML的 `output_schema` 定义的是 **step级业务Data**（每step不同）
  - 写PRD步骤: `{prd_content: string, sections: []string}`
  - 代码审查步骤: `{passed: bool, issues: []Issue}`

两层共同构成完整输出：`Result = 标准信封 + 业务Data`

**实现状态：** ⚠️ 部分实现（v4.3两层Output完全未实现）

**差异说明：**

**基本接口方面：**
- ✅ 基本接口已实现但大幅简化：
  ```go
  // 实际代码 — executor/executor.go
  type Executor interface {
      Type() string
      ValidateConfig(input map[string]any) error
  }
  type StepExecutor interface {
      Executor
      ExecuteStep(ctx context.Trace, adapter EngineAdapter, rc *RunTrace) error
  }
  ```
- ✅ `Registry` 已实现（`executor.Registry`），支持 `Register`/`Get`/`GetStep`
- ❌ `ControlExecutor` 接口未定义——control节点直接在engine中inline实现

**v4.3两层Output方面：**
- ❌ `OutputType() reflect.Type` 方法不存在——无法声明executor级标准信封Schema
- ❌ 没有AgentOutput/ShellOutput等Go struct定义标准信封
- ❌ `ExecutorResult` 统一数据结构未实现。数据通过 `RunTrace` 传入，结果通过adapter回写DB
- ❌ 当前Task entity的数据模型是扁平的：
  ```go
  // entity/entity.go — Task struct
  Output  string  `gorm:"type:text" json:"output"`                         // 文本输出（混合了标准信封和业务数据）
  Data    string  `gorm:"column:structured_output;type:text" json:"data"`  // JSON（未区分信封和业务Data）
  ```
  没有任何字段区分"executor标准信封"（如token_usage）和"step业务Data"（如prd_content）

**其他缺失：**
- ❌ `InputType()` / `TraceType()` — 其余Schema方法也均未实现（注：ConfigType已在v4.4中砍掉）
- ❌ `Init(config, rt Runtime)` / `Close()` — 生命周期方法未实现
- ❌ `Execute(ctx, task) (*Result, error)` — PRD的简洁签名未实现
- ⚠️ ValidateConfig 在三个executor中都返回nil（空实现），未做实际验证：
  ```go
  // agent.go:16, human.go:30, subprocess.go:31
  func (a *AgentExecutor) ValidateConfig(_ map[string]any) error { return nil }
  ```

**v4.3实现建议：**

实现两层Output需要以下改动：

1. **定义标准信封Go struct**：
   ```go
   type AgentStandardOutput struct {
       TokenUsage  *TokenUsage   `json:"token_usage"`
       SessionKey  string        `json:"session_key"`
       Model       string        `json:"model"`
       Duration    time.Duration `json:"duration"`
   }
   ```

2. **Executor接口添加OutputType()**：
   ```go
   func (a *AgentExecutor) OutputType() reflect.Type {
       return reflect.TypeOf(AgentStandardOutput{})
   }
   ```

3. **Task entity分层存储**（见§7）：
   ```go
   // 标准信封（executor级，固定结构）
   ExecutorOutput string `gorm:"type:text" json:"executor_output"`
   // 业务Data（step级，由output_schema定义）
   StepData       string `gorm:"type:text" json:"step_data"`
   ```

4. **完成回调分层写入**：AgentExecutor在saveStepCompletion时分别写入executor_output（token_usage等）和step_data（output_schema验证后的业务数据）

---

### §6 引擎核心简化

**PRD要求：**
- 引擎核心变成"一个DAG调度器 + executor注册表"
- 伪代码：registry.Get → renderInput → Dispatch → Wait → validateOutput → completeStep
- Control节点用有限枚举switch处理

**实现状态：** ⚠️ 部分实现

**差异说明：**
- ✅ DAG调度器已实现（`execute` 函数：拓扑排序、inDegree触发、并行执行、wg等待）
- ✅ executor registry路由已实现（`dispatchStep` 函数，第567行）
- ✅ condition/expand作为control节点inline处理
- ⚠️ 引擎仍然5545行，距PRD理想的~2500行差距大——原因：
  - `executeStepAttempt` 及其调用链（~200行）是旧agent执行逻辑残留
  - 大量assignee解析逻辑（claim/round_robin/load_balance/role resolution，2259-2489行，~230行）在engine中
  - retryStep（~200行）包含agent和approval两条路径
  - expand内联执行（~300行）仍使用engine内部的 `executeStepAttempt`
- ⚠️ `executeStepAttempt`（定义在第2519行）是遗留的agent执行逻辑，3处活跃引用：
  - 第4394行（expand子步骤执行）
  - 第4481行（retryStep agent执行）
  - 第4580行（expand escalation执行）
  这些执行路径绕过executor registry，直接在engine内部处理prompt渲染、agent spawn、callback等待

---

### §7 执行器数据存储（v4.5 — Event Sourcing模型）

**PRD要求（v4.5重写）：**
- **审计 = 三层Schema的自然副产品**，不是独立功能（InputType+OutputType+TraceType=完整审计）
- **Data是有序条目流**（Event Sourcing），不是大JSON对象
- 统一的 `executor_trace` 表，所有executor共用
- Storage接口：`Append()`流式写入 / `List()`分页读取 / `Search()`搜索
- executor运行时流式append，不需要完成时dump → 实时审计天然支持
- 审计核心 = executor完整私有数据（agent=完整session含system prompt，不是摘要）
- 前端StepTraceViewer根据executor type选渲染组件（SessionViewer/FormHistoryViewer/RequestViewer等）
- 两层数据分离：tasks表存标准信封（几KB），executor_trace表存完整数据（几KB~几MB）

**实现状态：** ❌ 未实现

**差异说明：**
- ❌ `executor_trace` 表不存在（需新建）
- ❌ `Storage` 接口（Append/List/Search）不存在
- ❌ 无流式写入机制，executor不在运行过程中产生TraceEntry
- ❌ 审计API不存在（无分页/搜索/实时增量拉取）
- ❌ 前端无StepTraceViewer组件
- 当前数据全部存在Task表的扁平字段中：
  ```go
  Output       string  // 文本输出
  Data         string  // column:structured_output — JSON结构化输出
  Outputs      string  // JSON — [output:key=value] 解析的变量map
  SessionID    string  // session标识
  AgentTrace string  // agent上下文（用于rejection feedback）
  ```
- `entity/executor.go` 只有 `ExecutorConfig`（并发配置，4个字段），不是executor数据存储：
  ```go
  type ExecutorConfig struct {
      ID            string    // 主键
      Name          string    // e.g. "main", "pm", "zebin"
      Type          string    // agent | human
      MaxConcurrent int       // 默认3
      QueueStrategy string    // fifo | priority
  }
  ```

**v4.3两层Output对存储模型的影响：**

当前的`structured_output`字段混合存储了所有输出数据，无法区分哪些是executor标准信封、哪些是step业务Data。实现两层Output需要：

| 字段 | 内容 | 来源 | 举例 |
|------|------|------|------|
| `executor_output` (新) | 标准信封JSON | executor完成时自动填充 | `{"token_usage":{"input":1200,"output":3400},"session_key":"agent:main:sub:xxx","model":"claude-opus-4-6"}` |
| `step_data` (新) | 业务数据JSON | output_schema验证后的data | `{"prd_content":"...","sections":["需求","设计"]}` |
| `output` (已有) | 文本输出 | executor文本回复 | 保持不变 |

下游步骤通过 `{{steps.X.data.Y}}` 引用的是 `step_data`，而审计/监控看的是 `executor_output`。

---

### §8 YAML Schema验证

**PRD要求：**
- 两层验证：引擎验证公共字段 + Executor验证专属input
- Executor通过 `ValidateConfig(input)` 自带验证逻辑

**实现状态：** ⚠️ 部分实现

**差异说明：**
- ✅ 接口已定义：`Executor.ValidateConfig(input map[string]any) error`
- ✅ 引擎公共字段验证已有（`ValidateYAML` 函数存在于 `workflow_service.go`）
- ❌ 三个executor的 `ValidateConfig` 都是空实现（`return nil`），没有做任何实际验证：
  - `AgentExecutor.ValidateConfig` — 不验证prompt/assignee是否存在
  - `HumanExecutor.ValidateConfig` — 不验证form定义/approval配置是否完整
  - `SubprocessExecutor.ValidateConfig` — 不验证process引用是否存在
- ❌ publish时executor-specific验证无效（因为都返回nil）

---

### §9 通知架构

**PRD要求：**
- 步骤级通知：生命周期钩子（`on_complete`/`on_fail` YAML配置）
- 全局通知规则：事件驱动，订阅event log

**实现状态：** ⚠️ 部分实现

**差异说明：**
- ✅ HumanExecutor内有飞书通知（`sendHumanNotification` 方法，使用 `FeishuNotifier` 接口）
- ✅ Approval投票通知（`sendVoteNotification` 同时支持飞书人类和agent spawn）
- ✅ Workflow完成通知（`e.Notifier.NotifyWorkflowComplete`）
- ✅ TriggerEmitter已实现，AgentExecutor在 `emitStepCompletedEvent` 中发出 `step.completed` 事件
- ❌ `on_complete` / `on_fail` 钩子未在StepDef中定义
- ❌ 全局事件驱动通知规则未实现（event log存在但无消费者订阅通知）

---

### §10 行业标准对齐

**PRD要求：** 对齐Agent Protocol / Responses API / A2A Protocol / NIST标准

**实现状态：** ❌ 未实现

**差异说明：** 长期方向性目标，当前无代码对接任何标准协议。

---

### §11 实施路径

**PRD要求：**
- P0：接口抽象（不改行为）—— 定义Executor interface，包装现有逻辑
- P1：YAML格式迁移 —— `type:` → `executor:` / `control:`
- P2：新增执行器 —— Shell/HTTP/ClaudeCode/FeishuApproval
- P3：生态扩展 —— 插件热加载、Agent Protocol对接

**实现状态：** ⚠️ P0~P1部分实现

**差异说明：**
- ✅ P0 大部分完成：Executor/StepExecutor接口已定义，三个executor已抽取为独立struct
- ✅ P1 已完成：StepDef用 `Executor` + `Control` 字段替代旧 `type` 字段
- ⚠️ P0 残留：
  - `executeStepAttempt` 仍在engine中（3处活跃引用），expand和retryStep走旧路径
  - Approval路由bug（registry中无"approval"别名）
- ❌ P2 未开始：Shell/HTTP/ClaudeCode/FeishuApproval executor均未实现
- ❌ P3 未开始
- ❌ v4.3两层Output需要在P0.5（接口增强）或P2（新executor实现时）落地

---

### §12 战略意义

**PRD要求：** 产品定位——ACP成为通用编排平台

**实现状态：** ✅ 已实现（设计层面）

**差异说明：** 这是愿景章节。代码架构已朝这个方向迈出第一步（executor抽象 + registry），但距"通用编排平台"还有较大差距（缺shell/http/cc executor，缺两层Output，缺Runtime SDK）。

---

### §13 实现反思与补充

**PRD要求（多个子节）：**

**§13.1 控制节点需要独立接口：** 两个接口——Executor + ControlExecutor

- **实现状态：** ⚠️ 部分实现
- ✅ Executor(StepExecutor)已有独立接口
- ❌ ControlExecutor接口未定义——condition/expand直接在engine中作为私有方法实现（`executeConditionInline`、`executeExpandInline`）

**§13.2 RunTrace打包传递：**

- **实现状态：** ✅ 已实现
- ✅ `executor.RunTrace` 已定义（352行的executor.go中），包含所有DAG调度状态
- ✅ `buildRunTrace` 函数在engine中将多个参数打包为RunTrace
- ✅ RunTrace包含丰富的步骤配置：Conditions、Approval、Gate、Component、FormJSON等

**§13.3 错误恢复与幂等性：** Handle可序列化、Resume方法、幂等key

- **实现状态：** ⚠️ 部分实现
- ✅ 引擎有recovery mode（`isRecoveryMode(ctx)`）：
  - AgentExecutor：`spawnAgent` 在recovery模式跳过重复spawn（`agent.go:182`）
  - HumanExecutor：`sendHumanNotification` 在recovery模式跳过飞书通知
  - HumanExecutor：`createOrReuseVoteTask` 支持crash recovery（检查已有vote task，`human.go:270`）
  - SubprocessExecutor：无显式recovery处理（依赖子流程自身状态）
- ❌ `Resume(ctx, handle)` 方法未实现
- ❌ 没有通用的Handle序列化/持久化机制
- ❌ HTTP executor幂等key机制不存在（因为HTTP executor本身不存在）

**§13.4 Executor生命周期管理：** Init/Close方法

- **实现状态：** ❌ 未实现
- ❌ Executor接口中无 `Init()` / `Close()` 方法
- 当前executor通过 `ExecutorDeps` 注入依赖，在 `initExecutors()` 中一次性构建：
  ```go
  // workflow_engine.go:224-229
  deps := e.executorDeps()
  e.registry = executor.NewRegistry()
  e.registry.Register(&executor.SubprocessExecutor{Deps: deps})
  e.registry.Register(&executor.AgentExecutor{Deps: deps})
  e.registry.Register(&executor.HumanExecutor{Deps: deps})
  ```

**§13.5 Human与Approval合并：** 方案A——合并为单一HumanExecutor

- **实现状态：** ✅ 已实现
- ✅ HumanExecutor内部按 `rc.Approval != nil` 分支处理form和approval两种模式（`human.go:288`）
- ✅ 共享逻辑（任务创建、通知、回调等待）复用
- ✅ Approval子流程完整实现：resolveApprovers → createOrReuseVoteTask → sendVoteNotification → WaitForApprovalVotes → handle result（approved/rejected/timeout）

**§13.6 Executor级别可观测性：** Result增加Metrics字段

- **实现状态：** ⚠️ 部分实现
- ✅ `StepPerformanceMetrics` 已定义（executor.go中）包含：
  - GateLoops, GatePassed, ApprovalResult, Escalated
  - ExecutionTimeMs, Attempt, FinalStatus
  - IssueCategory, IssueDetail
- ✅ AgentExecutor在完成时调用 `recordPerformance` 记录指标（agent.go多处）
- ❌ 没有统一的 `Result.Metrics` 字段——metrics直接通过 `Deps.Performance.RecordPerformance` 写入，不经过统一接口
- ❌ v4.3两层Output中的标准信封（token_usage, duration等）目前不作为结构化metrics返回，而是散落在各处

**§13.7 优雅降级策略：** 未注册executor类型的处理

- **实现状态：** ✅ 已实现
- ✅ `dispatchStep` 中 `registry.GetStep(execType)` 返回nil时，step标记failed + 清晰错误消息（第588行）：
  ```go
  logEngine.Error("no step executor registered", "type", execType, "step_id", stepID)
  errMsg := fmt.Sprintf("no executor registered for type: %s", execType)
  ```
- ✅ 未知control类型同样有错误处理（`workflow_engine.go:2210`）
- ❌ 未来扩展方向（fallback到HTTPExecutor）未实现

---

### §14 确定性光谱：ACP的哲学基础

**PRD要求：** 确定性executor（http/shell）vs 非确定性executor（agent/human）在同一DAG共存

**实现状态：** ⚠️ 部分实现

**差异说明：**
- ✅ 架构支持在同一流程中混合不同executor类型
- ❌ 确定性executor（http/shell）不存在，无法在实际流程中体现"确定性光谱"
- ✅ output_schema作为约束手段已实现（AgentExecutor.validateOutputSchema，`agent.go:209`）
- ⚠️ v4.3两层Output与确定性光谱的关系：确定性executor（shell/http）的标准信封是完全确定的（exit_code, status_code），而非确定性executor（agent）的标准信封也是确定的（token_usage是精确数字），不确定性只存在于业务Trace层。这个分层有助于在混合流程中统一处理确定性和非确定性输出

---

### §15 企业即程序：ACP的产品叙事

**PRD要求：** 产品叙事——企业=程序，员工=函数，管理=编程

**实现状态：** ✅ 已实现（设计层面）

**差异说明：** 概念章节，不涉及代码。核心架构已体现这一理念。

---

### §16 Executor Runtime：平台能力SDK

**PRD要求：**
```go
type Runtime interface {
    Storage() Storage
    Logger() Logger
    Credentials() map[string]string
    Notify(target, message string) error
    Config() map[string]any
}
```
- Executor不主动获取资源，一切通过声明+注入
- Runtime vs EngineAdapter 分工明确
- 双层Schema模型：Input Schema（参数验证，含data_schema）+ Trace Schema（存储声明）

**实现状态：** ❌ 未实现

**差异说明：**
- ❌ `Runtime` 接口未定义
- ❌ `Storage` 接口未定义（v4.5明确定义为Append/List/Search三个方法，基于executor_trace表）
- ❌ 声明式凭证注入未实现
- 当前executor通过 `ExecutorDeps` 直接持有依赖——这是PRD Phase 1的过渡方案（§17.3确认），Phase 2才引入Runtime/Storage
- ✅ 依赖接口化已部分实现。`deps.go`（109行）定义了8个依赖接口：
  ```go
  type GatewayClient interface { AgentSpawnTask(...); GetAgentState(...) }
  type FeishuNotifier interface { IsConfigured(); BuildHumanTaskCard(...); SendInteractiveCard(...) }
  type CredentialProvider interface { BuildCredentialTrace(slugs []string) string }
  type TriggerEmitter interface { Emit(event TriggerEvent) }
  type LessonRecorder interface { RecordLesson(...); FetchLessons(...); FetchStepLessons(...) }
  type PerformanceRecorder interface { RecordPerformance(...) }
  type AgentTraceCapture interface { CaptureAgentTrace(...) (string, error) }
  type OutputValidator interface { ValidateOutputSchema(...) string }
  type RoleResolver interface { GetRoleMembers(...) ([]RoleMember, error) }
  type SubprocessStart interface { StartSubprocessRun(...) (string, error) }
  ```
- ⚠️ `ExecutorDeps` 直接持有 `*gorm.DB`（deps.go最后）——这是Phase 2需要替换的核心依赖：
  ```go
  type ExecutorDeps struct {
      DB              *gorm.DB          // ← Phase 2需要替换为Storage
      Gateway         GatewayClient
      Feishu          FeishuNotifier
      Credentials     CredentialProvider
      Triggers        TriggerEmitter
      Lessons         LessonRecorder
      Performance     PerformanceRecorder
      AgentTrace    AgentTraceCapture
      OutputValidate  OutputValidator
      Roles           RoleResolver
      SubprocessStart SubprocessStart
  }
  ```

---

### §17 Executor全自治重构方案

**PRD要求：**
- executor持有DB/Gateway，直接操作DB（Phase 1过渡）
- EngineAdapter从63方法精简到~10方法
- AgentExecutor吸收~20个方法，HumanExecutor吸收~15个，SubprocessExecutor吸收~5个
- deps.go定义依赖接口，解决循环依赖

**实现状态：** ✅ 已实现（Phase 1目标基本达成）

**差异说明：**

**AgentExecutor（768行）吸收的方法：**
| 方法 | 行号 | 说明 |
|------|------|------|
| `sessionKeyFor` | 30 | agent session key生成 |
| `updateAttemptCount` | 34 | 更新重试次数 |
| `readStepStartTime` | 40 | 读取执行开始时间 |
| `updateStepDispatched` | 48 | 记录dispatch元数据 |
| `updateStepOutputsJSON` | 63 | 更新outputs JSON |
| `readTaskData` | 69 | 读取structured_output |
| `recordLesson` | 79 | 记录经验教训 |
| `recordPerformance` | 86 | 记录性能指标 |
| `emitStepCompletedEvent` | 92 | 发出step.completed事件 |
| `readRejectionFeedback` | 115 | 读取approval拒绝反馈 |
| `renderAndSaveComponent` | 123 | 渲染前端组件 |
| `spawnAgent` | 146 | Gateway agent spawn |
| `completeStepImmediate` | 197 | wait_reply=false时立即完成 |
| `saveStepCompletion` | 205 | 保存完成结果 |
| `validateOutputSchema` | 227 | output schema验证 |
| `fetchLessonsPrompt` | 234 | 获取lessons注入prompt |
| `buildCredentialTrace` | 250 | 构建凭证上下文 |
| `createStepTask` | 257 | 创建/更新step task |
| `executeAgentAttempt` | 307 | 单次执行尝试 |
| `ExecuteStep` | 383 | 完整执行入口（gate循环+重试+escalation） |

**HumanExecutor（881行）吸收的方法：**
| 方法 | 说明 |
|------|------|
| `prepareHumanStep` | 渲染form/component模板 |
| `updateTaskForHumanStart` | 设置pending状态和started_at |
| `sendHumanNotification` | 发送飞书通知 |
| `updateTaskCompletedWithData` | 更新完成状态 |
| `findMainTaskID` | 查找主task ID |
| `markApprovalRunning` | 标记approval运行中 |
| `resolveApprovers` | 解析审批人（feishu_user/agent/role） |
| `createOrReuseVoteTask` | 创建投票子任务（含crash recovery） |
| `sendVoteNotification` | 发送投票通知（飞书/agent spawn） |
| `collectVoteComments` | 收集投票评论 |
| `completeApprovalStep` | 标记审批完成 |
| `cancelPendingVotes` | 取消pending投票 |
| `prepareApprovalRestart` | 准备rejection后重启 |
| `executeApproval` | 完整审批子流程 |
| `handleApprovalApproved/Rejected/Timeout` | 审批结果处理 |

**SubprocessExecutor（290行）吸收的方法：**
| 方法 | 说明 |
|------|------|
| `markTaskRunning` | 更新running状态 |
| `lookupProcessByNameOrID` | 按名称/ID查找流程 |
| `checkSubprocessDepth` | 检查嵌套深度（上限5） |
| `pollSubprocessRun` | 轮询子流程状态 |
| `completeSubprocessTask` | 复制子流程输出到父步骤 |
| `failStep` | 标记失败helper |

**EngineAdapter当前12个方法：**
```go
type EngineAdapter interface {
    // 状态持久化（2个）
    UpdateStepStatus(runID, stepID, status, output string, errMsg *string)
    WriteStepLog(runID, stepID, level, msg string, data map[string]any)
    // 模板渲染（1个）
    RenderPrompt(tmpl string, inputs map[string]string, stepOutputs interface{}) string
    // Step Output管理（1个）
    SetStepOutput(stepOutputs interface{}, stepID, output string, outputs map[string]string, data string)
    // Agent步骤（6个 — 仍有业务逻辑）
    CheckAndWaitCapacity(ctx context.Trace) error
    EscalateToAgent(ctx context.Trace, gate *GateConfig, stepID string, history []string) (string, error)
    RegisterCallback(runID, stepID string) chan CallbackResult
    UnregisterCallback(runID, stepID string)
    WaitForCallback(ctx context.Trace, ch chan CallbackResult, ...) (CallbackResult, error)
    ResetTaskForGateRetry(runID, stepID, taskID string)
    // Human步骤（2个 — 仍有业务逻辑）
    WaitForHumanCallback(ctx context.Trace, runID, stepID string) (...)
    WaitForApprovalVotes(ctx context.Trace, runID, stepID string, ...) (...)
}
```

**残留问题：**
- ⚠️ EngineAdapter仍有12个方法，其中 `CheckAndWaitCapacity`、`EscalateToAgent`、`WaitForHumanCallback`、`WaitForApprovalVotes`、`ResetTaskForGateRetry` 是业务逻辑方法。PRD §17.4目标是精简到~7个纯引擎方法
- ⚠️ `executeStepAttempt` 仍存在于engine中（3处活跃引用），expand和retryStep路径未完全迁移到AgentExecutor
- ✅ executor持有 `*gorm.DB` 直接操作——符合Phase 1预期，Phase 2需替换为Storage接口

---

### §18 Step三层Schema模型与自动校验（v4.4 砍掉Config层）

**PRD要求：**
- InputType() / OutputType() / TraceType() — 三个方法返回reflect.Type（ConfigType已砍掉，配置参数统一放Input里）
- Go struct定义Schema → 反射自动生成JSON Schema
- 统一Flow Schema API（`GET /api/flow/schema`）
- Monaco编辑器自动校验
- Publish时跨步骤类型安全验证

**v4.3/v4.4对Schema模型的影响：**

v4.3的两层Output决策 + v4.4砍掉Config层，Schema模型从四层简化为三层。**OutputType()定义的是executor级标准信封，而input.data_schema定义的是step级业务Data**。这意味着：

1. **OutputType()** 返回的Go struct是executor固有的（如`AgentOutput{TokenUsage, SessionKey, Model}`），所有使用该executor的step都有这个标准信封
2. **output_schema**（YAML配置）是per-step的，约束的是业务Data字段
3. 下游步骤通过 `{{steps.X.data.Y}}` 引用的是业务Data（output_schema定义的字段），不是标准信封
4. 标准信封用于审计、监控、计费——不参与流程数据流
5. Publish时的跨步骤类型安全验证只验证 output_schema 定义的字段，不验证标准信封

三层Schema模型（v4.4）：

| 层 | 方法 | 说明 | v4.3备注 |
|---|---|---|---|
| Input Schema | `InputType()` | executor接受什么输入 | 不受v4.3影响 |
| Output Schema | `OutputType()` | executor标准信封（固定结构） | **v4.3核心**：定义token_usage等executor级输出 |
| ~~Config Schema~~ | ~~`ConfigType()`~~ | ~~executor配置参数~~ | v4.4砍掉，配置统一放Input |
| Trace Schema | `TraceType()` | executor私有数据（v4.5: 有序条目流，存executor_trace表） | **v4.5核心**：定义entry types + content结构，审计=读取TraceType()数据 |
| **(新增) data_schema** | **YAML input字段** | **step业务Data（每step不同）** | **v4.3核心**：放在input里，约束prd_content等业务字段 |

**实现状态：** ❌ 未实现

**差异说明：**
- ❌ Executor接口中无 `InputType()` / `OutputType()` / `TraceType()` 方法（ConfigType已在v4.4中砍掉）
- ❌ 没有Go struct定义的Schema（如 `AgentInput`、`AgentOutput`、`HumanInput`、`HumanConfig`）
- ❌ 无JSON Schema自动生成（未引入 `github.com/invopop/jsonschema`）
- ❌ 无Flow Schema API
- ❌ 无跨步骤类型安全验证
- ❌ 无标准信封struct定义（v4.3要求的AgentStandardOutput等）
- 当前YAML验证是手动检查字段存在性，不基于executor声明的Schema
- 当前output_schema验证只在runtime做（AgentExecutor.validateOutputSchema），不在publish时做

---

## 代码超前于PRD的部分

以下功能在代码中已实现，但PRD未提及或未充分讨论：

1. **动态Assignee解析（AssigneeResolver）**
   - 支持 `candidates`/`rules`/`relation`/`role` 等多种分配策略
   - claim（认领）、round_robin（轮询）、load_balance（负载均衡）策略
   - 代码中大量逻辑处理动态分配（engine第2259-2489行，~230行），PRD完全未涉及
   - 这些逻辑在engine中而非executor中——因为是"在确定用哪个executor前的前置路由"

2. **Executor并发控制（ExecutorConfig）**
   - `entity/executor.go` 定义了 `ExecutorConfig`（per-executor并发上限、队列策略），4个字段
   - `checkExecutorCapacity` / `waitForExecutorSlot` 在engine中实现
   - 通过EngineAdapter的 `CheckAndWaitCapacity` 暴露给executor
   - PRD只在§16提到capacity概念，未详细设计

3. **On-Submit Server-Side Hook**
   - `executeOnSubmitServerSide` 支持human步骤提交后执行服务端逻辑
   - PRD §9只提到通知钩子，未提到on_submit

4. **Component系统（前端组件渲染）**
   - `ComponentConfig` struct在executor包中定义（System/Component/Props三个字段）
   - agent和human executor都支持渲染component props模板
   - AgentExecutor的 `renderAndSaveComponent` 方法处理模板变量渲染
   - PRD未提及component概念

5. **Trigger系统（事件驱动触发）**
   - `TriggerEmitter` 接口 + `TriggerEvent` 类型在deps.go中定义
   - AgentExecutor在 `emitStepCompletedEvent` 中发出 `step.completed` 事件
   - ProcessInstance.TriggerChain用于防止循环触发
   - PRD §9提到事件驱动但未详细设计

6. **Step Lessons（步骤级经验沉淀）**
   - `LessonRecorder` 接口支持 `FetchStepLessons(processID, stepID)` 按process+step维度积累经验
   - AgentExecutor在 `fetchLessonsPrompt` 中自动将domain lessons和step lessons注入prompt
   - Gate解决后通过 `recordLesson` 记录经验
   - PRD未提及

7. **Agent Trace Capture**
   - `AgentTraceCapture` 接口在deps.go中定义
   - AgentExecutor的 `saveStepCompletion` 在步骤完成时捕获agent上下文
   - 用于审批拒绝后重新执行时注入 rejection feedback（`readRejectionFeedback`）
   - PRD未提及

8. **Round Variables（循环变量注入）**
   - `injectRoundVars` 在condition loop_to场景注入 `{{round}}` 变量
   - ProcessInstance.CurrentRound / MaxRounds 字段支持全局循环计数
   - PRD未提及

9. **Instance Role Assignment**
   - `RoleResolver` 接口在deps.go中定义
   - 支持per-instance的角色绑定（entity.ProcessInstance有相关字段）
   - HumanExecutor的 `resolveApprovers` 使用RoleResolver解析角色到具体人/agent
   - PRD未提及

---

## 实现程度统计

| 状态 | 章节数 | 占比 |
|------|--------|------|
| ✅ 已实现 | 5/18 | 28% |
| ⚠️ 部分实现 | 8/18 | 44% |
| ❌ 未实现 | 5/18 | 28% |

**已实现（5）：** §2（正交维度）、§4（调研）、§12/§15（愿景/叙事）、§17（全自治重构Phase 1）
**部分实现（8）：** §1（统一模型）、§3（type消失）、§5（统一协议）、§6（引擎简化）、§8（Schema验证）、§9（通知）、§11（实施路径）、§13/§14（实现反思+确定性光谱）
**未实现（5）：** §7（数据存储）、§10（行业标准）、§16（Runtime SDK）、§18（三层Schema）、以及§5中的两层Output

---

## PRD与代码不一致的关键差异清单

1. **Executor接口签名差异**：PRD要求 `Execute(ctx, task) (*Result, error)` + 三层Schema方法（InputType/OutputType/TraceType） + OutputType()定义标准信封，实际是 `ExecuteStep(ctx, adapter, rc) error` + 仅ValidateConfig（空实现）
2. **🐛 Approval路由Bug**：engine传 `execType="approval"` 给 `dispatchStep`（第2179、4440行），但registry只注册了 `Type()="human/agent/subprocess"` 的executor。`GetStep("approval")` 返回nil → step标记failed
3. **两层Output完全缺失（v4.3）**：没有executor标准信封struct、没有OutputType()方法、Task entity无分层存储字段（executor_output/step_data），当前structured_output混合存储一切
4. **EngineAdapter仍过重**：12个方法中至少5个（CheckAndWaitCapacity, EscalateToAgent, WaitForHumanCallback, WaitForApprovalVotes, ResetTaskForGateRetry）是业务逻辑方法
5. **旧代码残留**：`executeStepAttempt`（engine第2519行）仍有3处活跃引用（expand子步骤、retryStep、expand escalation），这些路径绕过executor registry
6. **ValidateConfig空实现**：三个executor的 `ValidateConfig` 都返回nil，publish时无法校验executor-specific配置
7. **Control节点无独立接口**：PRD§13.1要求 `ControlExecutor` 接口，实际condition/expand直接是engine私有方法
8. **没有Dispatch/Wait分离**：PRD§5核心设计（Dispatch→Handle→Wait模式）未实现，executor内部自己管全生命周期
9. **executor持有裸DB**：Phase 1过渡方案，需要Phase 2引入Runtime/Storage替换
10. **缺少确定性executor**：http/shell/claude_code executor未实现，"确定性光谱"无法在实际流程中体现
11. **Event Sourcing数据层完全缺失**：`executor_trace`表不存在，Storage接口（Append/List/Search）不存在，executor无流式写入机制，审计API不存在
12. **三层Schema + JSON Schema自动生成完全缺失**：无Go struct定义（InputType/OutputType/TraceType）、无jsonschema反射、无Flow Schema API

---

## 建议的Phase清单（按优先级排序）

### Phase 1: 修复Approval路由Bug + 清理旧代码残留 ⚡
- **前置依赖**: 无
- **风险**: 低（修bug + 删dead code，不改行为）
- **价值/成本**: 高/低 — 消除技术债 + 修复生产bug
- **受v4.3影响**: 否
- **具体任务**:
  1. **修复Approval路由**：在 `initExecutors()` 中额外注册 `"approval"` 别名指向HumanExecutor：
     ```go
     humanExec := &executor.HumanExecutor{Deps: deps}
     e.registry.Register(humanExec)                          // Type()="human"
     e.registry.RegisterAlias("approval", humanExec)         // 别名
     ```
     或在Registry中添加 `RegisterAlias` 方法
  2. **排查 `executeStepAttempt` 的3处活跃引用**：
     - 第4394行（expand子步骤）→ 需要用AgentExecutor替代或保留（expand的子步骤执行路径较特殊）
     - 第4481行（retryStep agent执行）→ 应迁移到通过executor registry路由
     - 第4580行（expand escalation）→ 同上
  3. 如果expand路径可以改为通过executor registry路由，删除 `executeStepAttempt` 函数
  4. 确认所有主要执行路径（包括retryStep）都通过executor registry路由

### Phase 2: EngineAdapter进一步精简
- **前置依赖**: Phase 1
- **风险**: 中（涉及方法迁移，需要仔细测试回调逻辑）
- **价值/成本**: 高/中 — 让EngineAdapter真正只做引擎内存状态操作
- **受v4.3影响**: 否
- **具体任务**:
  1. 将 `CheckAndWaitCapacity` 移入executor内部（从adapter.CheckAndWaitCapacity改为executor自己持有concurrency config通过Deps）
  2. 将 `EscalateToAgent` 移入AgentExecutor（gate escalation是agent-only逻辑）
  3. 将 `WaitForHumanCallback` 移入HumanExecutor内部（需要传入pendingCallbacks map或用channel替代）
  4. 将 `WaitForApprovalVotes` 移入HumanExecutor内部
  5. 将 `ResetTaskForGateRetry` 移入AgentExecutor
  6. 目标：EngineAdapter只保留 `UpdateStepStatus`、`WriteStepLog`、`RenderPrompt`、`SetStepOutput`、`RegisterCallback`、`UnregisterCallback`、`WaitForCallback`（~7个方法）

### Phase 3: ValidateConfig实现 + YAML Schema增强
- **前置依赖**: 无（可与Phase 1/2并行）
- **风险**: 低（只在publish时增加校验，不影响运行时）
- **价值/成本**: 中/低 — 提前发现YAML配置错误
- **受v4.3影响**: 否（v4.3的output_schema验证在Phase 9）
- **具体任务**:
  1. AgentExecutor.ValidateConfig：验证prompt或assignee存在、credentials引用合法、output_schema格式正确
  2. HumanExecutor.ValidateConfig：验证form定义合法、approval配置完整（approvers/strategy必填）
  3. SubprocessExecutor.ValidateConfig：验证process引用存在且已published
  4. 在publish流程中确认调用executor.ValidateConfig并返回错误

### Phase 4: ControlExecutor接口定义
- **前置依赖**: Phase 1
- **风险**: 低（纯重构，行为不变）
- **价值/成本**: 中/中 — 架构一致性
- **受v4.3影响**: 否
- **具体任务**:
  1. 在 `executor/executor.go` 中定义 `ControlExecutor` 接口：
     ```go
     type ControlExecutor interface {
         Type() string
         ExecuteControl(ctx context.Trace, rc *RunTrace, adapter EngineAdapter) error
     }
     ```
  2. 将 `executeConditionInline` 抽取为 `ConditionController.ExecuteControl`
  3. 将 `executeExpandInline` 抽取为 `ExpandController.ExecuteControl`
  4. 在engine中通过控制节点registry路由（或保持inline但统一接口签名）

### Phase 5: 两层Output基础设施 — Task entity分层存储 🆕 v4.3
- **前置依赖**: Phase 1
- **风险**: 中（涉及DB schema变更）
- **价值/成本**: 高/中 — v4.3核心决策落地的第一步
- **受v4.3影响**: ✅ 这就是v4.3的核心实现
- **具体任务**:
  1. **Task entity增加分层字段**：
     ```go
     // 标准信封 — executor级固定输出
     ExecutorOutput string `gorm:"column:executor_output;type:text;default:''" json:"executor_output,omitempty"`
     // Step业务Data — YAML output_schema定义的业务数据
     // 注：原structured_output字段保留作为向后兼容，新代码写StepData
     StepData       string `gorm:"column:step_data;type:text;default:''" json:"step_data,omitempty"`
     ```
  2. **Task entity增加executor_type字段**：
     ```go
     ExecutorType string `gorm:"column:executor_type;default:''" json:"executor_type,omitempty"`
     ```
  3. **DB migration**：添加executor_output、step_data、executor_type三个字段
  4. **AgentExecutor.saveStepCompletion改造**：分层写入
     - executor_output ← `{"token_usage":..., "session_key":..., "model":..., "duration":...}`
     - step_data ← output_schema验证后的业务数据
     - output ← 文本输出（保持不变）
  5. **HumanExecutor.updateTaskCompletedWithData改造**：
     - executor_output ← `{"response_time":..., "channel":...}`
     - step_data ← 表单提交数据
  6. **SubprocessExecutor.completeSubprocessTask改造**：
     - executor_output ← `{"child_instance_id":..., "step_count":..., "duration":...}`
     - step_data ← 子流程最终输出
  7. **下游引用路径**：确认 `{{steps.X.data.Y}}` 引用的是step_data而非executor_output
  8. **向后兼容**：structured_output字段保留，新代码写step_data + executor_output，读取时先查step_data后fallback structured_output

### Phase 6: ShellExecutor实现
- **前置依赖**: Phase 1, Phase 5（用两层Output写executor_output.exit_code）
- **风险**: 中（新功能，需要安全考虑）
- **价值/成本**: 高/中 — 开启"确定性光谱"
- **受v4.3影响**: ✅ ShellExecutor的标准信封：`{exit_code, duration, stdout_bytes}`
- **具体任务**:
  1. 定义 `ShellExecutor` struct，实现 StepExecutor 接口
  2. 定义标准信封 `ShellStandardOutput{ExitCode int, Duration time.Duration, StdoutBytes int64}`
  3. 支持 YAML: `executor: shell`, `input: { command: "make deploy", workdir: "/path" }`
  4. 进程管理：超时kill、stdout/stderr capture、exit code处理
  5. 安全：工作目录限制、命令白名单（可选）
  6. 运行时通过Storage.Append()写入TraceEntry（start/stdout_chunk/stderr_chunk/exit）
  7. 完成时返回标准信封ShellStandardOutput
  8. 注册到registry

### Phase 7: HTTPExecutor实现
- **前置依赖**: Phase 1, Phase 5
- **风险**: 中（外部依赖，需要重试/超时策略）
- **价值/成本**: 高/中 — 连接第三方API/Agent
- **受v4.3影响**: ✅ HTTPExecutor的标准信封：`{status_code, latency_ms, retries}`
- **具体任务**:
  1. 定义 `HTTPExecutor` struct
  2. 定义标准信封 `HTTPStandardOutput{StatusCode int, LatencyMs int64, Retries int}`
  3. 支持 YAML: `executor: http`, `input: { url, method, headers, body, timeout }`
  4. 响应处理：状态码判断、body解析为step_data
  5. 重试策略：可配置重试次数、backoff
  6. 凭证注入：从credentials配置注入Authorization headers

### Phase 8: Event Sourcing数据层 — executor_trace表 + Storage接口 🆕 v4.5
- **前置依赖**: Phase 5（两层Output需要先就位）
- **风险**: 中（新表+新接口，需要迁移现有executor的数据写入方式）
- **价值/成本**: 高/中 — 审计可观测性+实时查看的基础设施
- **受v4.5影响**: ✅ 核心，v4.5的实现主体
- **具体任务**:
  1. 新建 `executor_trace` 表（instance_id, step_id, seq, entry_type, content, summary, created_at）
  2. 实现 Storage 接口（Append/List/Search）
  3. AgentExecutor改造：执行完成时调Gateway拿完整session，逐条Append（system_prompt/message/tool_call/token_snapshot）
  4. HumanExecutor改造：通知发送时Append(notification)，表单提交时Append(form_submit)，投票时Append(vote)
  5. SubprocessExecutor改造：子流程事件Append(start/step_completed/finish)
  6. 审计API：`GET /api/instances/{id}/steps/{stepID}/data` 支持分页(?offset&limit)、搜索(?q)、类型过滤(?type)、增量拉取(?after_seq)
  7. 前端StepTraceViewer组件（根据executor type选渲染：SessionViewer/FormHistoryViewer/GenericEntryList）

### Phase 9: 三层Schema + JSON Schema自动生成 🆕 v4.4
- **前置依赖**: Phase 5, Phase 8
- **风险**: 中（需要引入reflect依赖 + jsonschema库）
- **价值/成本**: 高/高 — 实现publish时类型安全验证 + 前端自动补全
- **受v4.3影响**: ✅ OutputType()返回标准信封Schema，output_schema验证业务Data
- **具体任务**:
  1. Executor接口增加 `InputType()` / `OutputType()` / `TraceType()` 返回 `reflect.Type`（无ConfigType）
  2. 为每个executor定义Go struct：
     - AgentExecutor: `AgentInput{Prompt, Assignee}`, `AgentStandardOutput{TokenUsage, SessionKey, Model}`, `AgentConfig{Model, MaxTokens, Timeout, Tools}`, `AgentData{SessionID, TokenUsage, AgentTrace}`
     - HumanExecutor: `HumanInput{Title, Assignee, Due}`, `HumanStandardOutput{ResponseTime, Channel}`, `HumanConfig{NotifyChannel, ReminderHours}`, `HumanData{NotifiedAt, ReminderSent}`
     - ShellExecutor: `ShellInput{Command, WorkDir}`, `ShellStandardOutput{ExitCode, Duration}`, `ShellConfig{Timeout, Sandbox}`, `ShellData{LogPath}`
     - HTTPExecutor: `HTTPInput{URL, Method, Headers, Body}`, `HTTPStandardOutput{StatusCode, LatencyMs}`, `HTTPConfig{Retries, Timeout}`, `HTTPData{}`
  3. 引入 `github.com/invopop/jsonschema` 反射生成JSON Schema
  4. 实现 `/api/flow/schema` API — 合并引擎通用schema + 所有executor的schema
  5. Publish时跨步骤验证：
     - 检查 `{{steps.X.data.Y}}` 引用的Y字段是否存在于步骤X的output_schema中（业务Trace层）
     - 标准信封字段（token_usage等）不参与跨步骤引用验证
  6. Monaco编辑器集成：加载Flow Schema → 自动补全、字段校验、类型检查

### Phase 10: Runtime/Storage接口（长期目标）
- **前置依赖**: Phase 8, Phase 2
- **风险**: 高（大规模接口变更，所有executor需要重写依赖注入方式）
- **价值/成本**: 高/高 — 标准化第三方executor开发体验
- **受v4.3影响**: ✅ Storage需要区分标准信封和业务Data的存储
- **具体任务**:
  1. 定义 `Runtime` 接口（Storage/Logger/Credentials/Notify/Config）
  2. 定义 `Storage` 接口（Set/Get/SetForStep/GetForStep/StoreBlob/GetBlob）
  3. 实现基于SQLite/PostgreSQL的Storage backend
  4. 从executor中移除 `*gorm.DB` 直接依赖
  5. executor通过 `Init(config, rt)` 接收Runtime
  6. 声明式Schema：executor定义 `DataSchema()` → engine自动建表

---

## Phase总结 — v4.3影响标注

| Phase | 优先级 | 受v4.3影响 | 核心交付 |
|-------|--------|-----------|---------|
| **Phase 1** | P0 ⚡ | ❌ | 修复Approval bug + 清理executeStepAttempt残留 |
| **Phase 2** | P0 | ❌ | EngineAdapter精简到~7方法 |
| **Phase 3** | P1 | ❌ | ValidateConfig实际实现 |
| **Phase 4** | P1 | ❌ | ControlExecutor接口 |
| **Phase 5** | P0 🆕 | ✅ **核心** | Task entity两层Output字段 + 分层写入 |
| **Phase 6** | P1 | ✅ | ShellExecutor + 标准信封 |
| **Phase 7** | P1 | ✅ | HTTPExecutor + 标准信封 |
| **Phase 8** | P2 🆕 | ✅ **核心** | Event Sourcing数据层 + executor_trace表 + Storage接口 |
| **Phase 9** | P2 🆕 | ✅ **核心** | 三层Schema + JSON Schema自动生成 |
| **Phase 10** | P3 | ✅ | Runtime/Storage SDK |

**v4.3决策直接影响的Phase：5、6、7、8、9、10（6/10个Phase）**

Phase 5是v4.3落地的最小可行改动（DB字段 + 分层写入），建议与Phase 1同优先级推进。Phase 9是v4.3的完整形态（Schema自动生成 + 跨步骤验证），属于中期目标。
