# ACP 流程调试模式 PRD — 像调试程序一样调试工作流

> 版本：v1.0 | 日期：2026-03-03 | 作者：Lyra
> 核心理念：企业工作流 = 程序，设计工作流 = 写代码，那就必须有调试器

## 1. 竞品调研

| 平台 | 调试方式 | 优点 | 缺点 |
|------|---------|------|------|
| **n8n** | 单步执行 + Data Pinning（钉住数据） | 可以冻结中间数据反复测试 | 没有断点，只能手动点 |
| **Dify** | 单节点测试 + 变量检查器 | 可以编辑缓存变量测试不同场景 | 不能从中间继续 |
| **Power Automate** | 运行历史 + Copilot诊断 | AI辅助定位错误 | 没有真正的调试模式 |
| **Temporal** | Event History Replay + Reset | 可以"回放时间" | 面向开发者，无可视化调试 |
| **AWS Step Functions** | 执行可视化 + Express/Standard双模式 | 实时看执行状态 | 没有断点/单步 |
| **飞书Anycross** | 测试运行 + 单节点调试 + Mock数据 | 集成度好 | 功能有限 |
| **IDE（VS Code等）** | 断点 + 单步 + 变量监视 + 调用栈 | 完整调试体验 | 只适用于代码 |

**结论：没有一个工作流平台做到了IDE级别的调试体验。** 我们的机会：把IDE调试器的概念完整映射到工作流引擎。

## 2. 设计理念

既然流程 = 程序，那调试体验就应该跟IDE一样：

| IDE概念 | 工作流映射 | ACP实现 |
|---------|----------|---------|
| **断点 (Breakpoint)** | 在某个step前暂停 | step上标记断点 |
| **单步执行 (Step Over)** | 执行当前step，暂停在下一个 | 逐步推进 |
| **步入 (Step Into)** | 进入子流程内部 | 进入subprocess |
| **继续 (Continue/Resume)** | 运行到下一个断点 | 继续执行 |
| **变量监视 (Watch)** | 查看step的输入/输出 | 数据检查器 |
| **修改变量 (Edit & Continue)** | 修改中间数据后继续 | 编辑step输出后继续 |
| **条件断点** | 满足条件时才暂停 | 表达式断点 |
| **调用栈 (Call Stack)** | 查看执行路径 | 已执行的step链 |
| **热重载 (Hot Reload)** | 修改代码不重启 | 修改step配置不重启实例 |

## 3. 调试模式的启动

### 3.1 两种启动方式

```
# 方式1：从头调试（新实例）
POST /api/instances
{
  "process_id": "xxx",
  "mode": "debug",        // ← 调试模式
  "input": { ... }
}

# 方式2：从失败点调试（已有实例）
POST /api/instances/{id}/debug
{
  "from_step": "failed_step_id"    // 从某个step开始调试
}
```

### 3.2 调试模式 vs 正常模式

| 特性 | 正常模式 | 调试模式 |
|------|---------|---------|
| 执行方式 | 全自动 | 受断点控制 |
| 速度 | 尽快完成 | 可暂停 |
| 数据可见性 | 完成后查看 | 实时查看每步输入/输出 |
| 数据可编辑 | ❌ | ✅ 可修改中间数据 |
| step配置可改 | ❌ | ✅ 可临时修改（不影响定义） |
| 实际副作用 | ✅ 真实执行 | 可选Mock模式 |
| 记入历史 | ✅ | 标记为debug执行 |

## 4. 核心功能

### 4.1 断点 (Breakpoint)

在流程设计界面，点击step左侧可以设置/取消断点（红色圆点，跟IDE一模一样）。

**断点类型：**

| 类型 | 说明 | 图标 |
|------|------|------|
| `before` | 执行前暂停（查看输入） | 🔴 红色实心圆 |
| `after` | 执行后暂停（查看输出） | 🔵 蓝色实心圆 |
| `conditional` | 条件满足时暂停 | 🟡 黄色实心圆 |

**断点存储原则：断点不在YAML里。**

IDE的断点存在 `.vscode/launch.json`，不在源代码文件里。源代码是共享的，断点是个人的。同理：YAML是流程定义（"源代码"），断点是调试状态（"用户配置"），两者分离。

| 数据 | 类比 | 存储位置 | 生命周期 | 共享范围 |
|------|------|---------|---------|---------|
| 流程YAML | 源代码 `.go` | 数据库 process_definitions | 永久 | 全员 |
| 断点 | `.vscode/launch.json` | 前端localStorage + 调试会话 | 跟用户走 | 仅个人 |
| Data Pinning | 测试fixture | 调试配置（debug_profiles） | 可选持久化 | 可共享 |
| Mock数据 | 测试用例 | 调试配置（debug_profiles） | 可选持久化 | 可共享 |

**调试配置（Debug Profiles）：** Pinning和Mock数据可以持久化为"调试配置"，独立于YAML存储，团队可共享常用测试场景：

```json
// debug_profiles — 独立存储，不在YAML里
{
  "process_id": "xxx",
  "profiles": {
    "happy_path": {
      "name": "高分通过路径",
      "input": { "pr_number": 42 },
      "mock_data": { "analyze": { "data": { "score": 90 } } },
      "pinned_steps": ["analyze"]
    },
    "low_score_path": {
      "name": "低分走审批路径",
      "input": { "pr_number": 42 },
      "mock_data": { "analyze": { "data": { "score": 30 } } }
    },
    "error_case": {
      "name": "API异常场景",
      "input": { "pr_number": 999 },
      "mock_data": { "http_call": { "status": "failed", "error": "timeout" } }
    }
  }
}
```

启动调试时可以选择Profile，像IDE选择调试配置一样：

```
🐛 ▼ [happy_path ▾]  ▶ Start Debugging
```

### 4.2 执行控制

调试模式下，引擎在断点处暂停，等待用户指令：

```
POST /api/instances/{id}/debug/action
{
  "action": "step_over"    // 执行当前step，暂停在下一个
}
```

| 动作 | 快捷键 | 说明 |
|------|--------|------|
| **Continue** (F5) | `continue` | 运行到下一个断点 |
| **Step Over** (F10) | `step_over` | 执行当前step，暂停在下一个 |
| **Step Into** (F11) | `step_into` | 如果是subprocess，进入子流程内部 |
| **Step Out** (Shift+F11) | `step_out` | 从子流程跳出，回到父流程 |
| **Run to Here** | `run_to: step_id` | 运行到指定step后暂停 |
| **Restart** | `restart` | 从头重新运行 |
| **Stop** | `stop` | 终止调试 |
| **Skip Step** | `skip` | 跳过当前step（标记skipped） |

### 4.3 数据检查器 (Data Inspector)

暂停时，可以查看当前step的完整上下文：

```
GET /api/instances/{id}/debug/inspect?step=implement

{
  "step_id": "implement",
  "status": "paused_before",     // 暂停在执行前

  "input": {                      // 渲染后的输入（变量已替换）
    "prompt": "实现用户登录API",
    "workdir": "/home/claw/project"
  },

  "input_raw": {                  // 原始输入（含变量模板）
    "prompt": "{{analyze.data.requirement}}",
    "workdir": "/home/claw/project"
  },

  "resolved_variables": {         // 变量解析详情
    "analyze.data.requirement": {
      "value": "实现用户登录API",
      "source_step": "analyze",
      "source_field": "data.requirement"
    }
  },

  "upstream_outputs": {           // 所有上游step的输出
    "analyze": {
      "status": "completed",
      "data": { "requirement": "实现用户登录API", "score": 85 },
      "output": "分析完成..."
    }
  },

  "execution_path": [             // 调用栈（已执行的路径）
    { "step": "analyze", "status": "completed", "duration_ms": 15000 },
    { "step": "quality_check", "status": "completed", "duration_ms": 5 },
    { "step": "implement", "status": "paused_before" }  // ← 当前位置
  ]
}
```

### 4.4 Edit & Continue（修改数据后继续）

**杀手级功能。** 暂停时可以修改step的输入或上游step的输出，然后继续执行——不需要重新跑整个流程。

```
POST /api/instances/{id}/debug/edit
{
  "step_id": "implement",
  "edit_type": "input",           // 修改当前step的输入
  "data": {
    "prompt": "实现用户登录API，使用JWT认证，增加rate limiting"   // 改了prompt
  }
}

// 或者修改上游输出
POST /api/instances/{id}/debug/edit
{
  "step_id": "analyze",
  "edit_type": "output",          // 修改上游step的输出
  "data": {
    "score": 50                    // 改成低分，看看走不走另一条分支
  }
}
```

**使用场景：**
- Agent输出不理想 → 手动修正输出 → 继续后续步骤
- 想测试不同分支 → 修改gate条件的上游数据 → 看走哪条路
- 外部API返回异常 → Mock一个正常结果 → 继续调试后面的逻辑

### 4.5 数据钉住 (Data Pinning)

借鉴n8n的概念：把某个step的输出"钉住"，后续重新运行时跳过执行，直接使用钉住的数据。

```
POST /api/instances/{id}/debug/pin
{
  "step_id": "analyze",
  "pinned_output": {                // 钉住的输出
    "data": { "score": 85, "requirement": "..." },
    "output": "分析结果..."
  }
}
```

**钉住的step在调试时：**
- 显示📌图标
- 不真正执行executor
- 直接返回钉住的数据
- 节省时间和API费用（特别是Agent/CC步骤）

**典型流程：**
1. 第一次完整调试，Agent步骤执行了5分钟花了$2
2. 钉住Agent步骤的输出
3. 后续调试只跑Agent后面的步骤，秒级完成，零费用

### 4.6 Mock模式

整个调试可以选择Mock模式——所有executor不真正执行，使用Mock数据。适合测试流程逻辑（control节点走向是否正确），不关心具体执行结果时使用。

```
POST /api/instances
{
  "process_id": "xxx",
  "mode": "debug",
  "mock": true,              // Mock模式
  "mock_data": {             // 预定义Mock输出
    "analyze": { "data": { "score": 85 } },
    "implement": { "data": { "result": "mock implementation" } }
  }
}
```

**Mock模式下：**
- Executor不执行
- 如果有预定义mock_data → 使用预定义数据
- 如果没有 → 暂停等用户输入，或使用executor的default output schema生成空数据
- Control节点正常评估（测试分支逻辑）
- 零费用、秒级完成

## 5. UI设计

### 5.1 设计视图 vs 调试视图

调试不能用设计视图（DAG编辑器）。设计视图关注**结构**（谁连谁），调试关注**执行**（走到哪了、数据是什么）。就像IDE里编辑代码用**编辑器**，调试代码用**调试器**——是两个不同的视图。

顶部Tab切换：

```
[📐 设计] [🐛 调试] [📊 甘特图] [📋 历史]
```

| Tab | 用途 | 布局 |
|-----|------|------|
| 📐 设计 | DAG编辑器，拖拽连线 | 自由画布 |
| 🐛 调试 | 调试执行，断点+数据检查 | 三面板+控制台 |
| 📊 甘特图 | 时间维度查看执行 | 甘特图（见甘特图PRD） |
| 📋 历史 | 执行记录列表 | 表格 |

### 5.2 调试视图：三面板布局

```
┌─────────────────────────────────────────────────────────────────────────┐
│ 🐛 调试模式  [happy_path ▾]  ▶F5  ⏭F10  ⏬F11  ⏫⇧F11  ⏹Stop  🔄    │
│ ⏸ 暂停在 [implement] (断点)  │  已执行: 3/8步  │  耗时: 2m15s  │ $1.2  │
├───────────────┬─────────────────────────────┬───────────────────────────┤
│               │                             │                           │
│  执行序列      │        执行时间线             │     数据检查器             │
│  (Call Stack)  │       (Timeline)            │    (Inspector)            │
│               │                             │                           │
│ ✅ analyze    │ ████████ ✅ 12s              │ ▼ Input                  │
│   12s  $0.5   │                             │   prompt: "实现..." ✏️    │
│               │                             │   model: opus             │
│ ✅ design     │ ████████████████ ✅ 18s      │   workdir: /home/...     │
│   18s  $0.3   │                             │                           │
│               │                             │ ▼ Variables               │
│ ✅ if(>80)    │          ◇→true             │   analyze.data.score: 85  │
│   <1ms        │                             │   analyze.data.req: "..." │
│               │                             │                           │
│ ⏸ implement  │           ▓▓▓▓▓▓ ⏸          │ ▼ Upstream Outputs        │
│  🔴← 当前     │                ↑ HERE       │   ▶ analyze  ✅ {..}     │
│               │                             │   ▶ design   ✅ {..}     │
│ ⬜ review     │                □□□□□ ~4h    │                           │
│  🔴           │                             │ ▼ Output                  │
│ ⬜ deploy     │                     □□ ~5m  │   (尚未执行)               │
│               │                             │                           │
├───────────────┴─────────────────────────────┴───────────────────────────┤
│ 📋 Console                                                              │
│ 10:00:00  analyze started                                               │
│ 10:00:12  analyze completed (score=85, cost=$0.5)                       │
│ 10:00:12  design started                                                │
│ 10:00:30  design completed                                              │
│ 10:00:30  if(score>=80) evaluated: 85>=80 → true → implement            │
│ 10:00:30  implement → PAUSED at breakpoint                              │
│ _                                                                       │
└─────────────────────────────────────────────────────────────────────────┘
```

### 5.3 四个区域详解

#### 工具栏（顶部）

```
┌───────────────────────────────────────────────────────────────────┐
│ 🐛  [happy_path ▾]  ▶F5  ⏭F10  ⏬F11  ⏫⇧F11  ⏹  🔄           │
│ ⏸ implement (断点)  │  3/8步  │  2m15s  │  $1.2  │  Mock: OFF  │
└───────────────────────────────────────────────────────────────────┘
```

| 元素 | 说明 |
|------|------|
| Profile选择器 | 下拉选择调试配置（happy_path / error_case / ...） |
| 执行控制按钮 | Continue/StepOver/StepInto/StepOut/Stop/Restart |
| 当前状态 | 暂停在哪个step、原因 |
| 统计信息 | 已执行步骤数、总耗时、总费用 |
| Mock开关 | 全局Mock模式开关 |

#### 执行序列（左侧）— 调用栈

按执行顺序纵向排列所有步骤，显示状态和耗时：

| 图标 | 状态 | 说明 |
|------|------|------|
| ✅ | completed | 已完成，显示耗时和费用 |
| ⏸ + 🔴 | paused + breakpoint | 断点暂停，当前位置 |
| ⏸ | paused (step mode) | 单步暂停 |
| 🔄 | running | 执行中（动画） |
| ⬜ | waiting | 未到达 |
| ⏭ | skipped | 被跳过 |
| ❌ | failed | 失败 |
| 📌 | pinned | 使用钉住数据（不真正执行） |
| ◇ | control node | if/switch/loop评估结果 |

点击任意步骤 → 右侧数据检查器切换到该步骤。

#### 执行时间线（中间）— 甘特图增强

横向甘特图，展示时间维度：

- 已完成的步骤：绿色实心条（长度 = 实际耗时）
- 运行中/暂停中：蓝色动画条 + "HERE"标记
- 未执行的步骤：灰色虚线条（长度 = 预估耗时）
- Control节点：菱形标记 + 评估结果（`→true`）
- 并行步骤分行显示
- 当前位置有明显的竖线光标（类似播放器的播放头）

**时间线的独特价值：** 让用户直观看到"流程走了多远、还剩多少、卡在哪里"。这是序列视图和DAG图都做不到的。

#### 数据检查器（右侧）— 变量面板

显示选中步骤的完整数据上下文：

| 区块 | 内容 | 交互 |
|------|------|------|
| **Input** | 渲染后的输入参数 | ✏️ 可编辑（Edit & Continue） |
| **Input (Raw)** | 原始模板（含`{{}}`） | 只读，展开查看 |
| **Variables** | 引用的变量值及来源 | 点击跳转到来源步骤 |
| **Upstream Outputs** | 所有上游步骤的输出 | 可展开查看完整JSON |
| **Output** | 当前步骤的输出（已完成时） | ✏️ 可编辑 + 📌可钉住 |
| **Error** | 失败信息（失败时） | 展示错误详情 |

#### 控制台（底部）— 事件流

实时滚动的事件日志，类似IDE的Debug Console：

```
10:00:00  [START] 流程启动，input: { pr_number: 42 }
10:00:00  [STEP] analyze started (executor: agent, command: chat)
10:00:12  [STEP] analyze completed (12s, $0.5)
10:00:12  [CTRL] if(score>=80): 85>=80 → true → implement
10:00:12  [BREAK] implement → paused at breakpoint (before)
10:00:30  [USER] Edit & Continue: implement.input.prompt modified
10:00:30  [USER] Continue (F5)
10:00:30  [STEP] implement started (executor: claudecode, command: run)
```

**可过滤：** 只看 `[STEP]` / `[CTRL]` / `[BREAK]` / `[USER]` / `[ERROR]`

### 5.4 节点状态图标（统一）

| 状态 | 执行序列 | 时间线 | 设计视图 |
|------|---------|--------|---------|
| completed | ✅ 绿色 | 绿色条 | 绿色边框 |
| running | 🔄 蓝色动画 | 蓝色动画条 | 蓝色脉冲 |
| paused | ⏸ 蓝色 | 蓝色条+光标 | 蓝色脉冲+⏸ |
| breakpoint | 🔴 红点 | 红色竖线标记 | 🔴 左侧红点 |
| failed | ❌ 红色 | 红色条 | 红色边框 |
| waiting | ⬜ 灰色 | 灰色虚线条 | 灰色 |
| skipped | ⏭ 灰色 | 无 | 灰色虚线 |
| pinned | 📌 | 📌标记 | 📌 |
| edited | ✏️ 橙色 | 橙色标记 | ✏️ |

### 5.5 当前步骤高亮（Current Step Highlight）

IDE调试时，当前执行行会整行高亮（黄色背景 + 箭头 ▶），这是调试器最核心的视觉线索。工作流调试必须有同等强度的"当前位置"指示：

**三面板中的高亮规则：**

| 区域 | 当前步骤的高亮方式 | 非当前步骤 |
|------|------------------|----------|
| **执行序列（左）** | 整行高亮背景（蓝色/黄色）+ ▶ 箭头指示器 + 加粗文字 | 普通背景 |
| **时间线（中）** | 条形脉冲动画 + 竖线播放头光标 + 发光效果(glow) | 静态条形 |
| **数据检查器（右）** | 自动切换到当前步骤的数据 | 手动点击切换 |
| **控制台（底）** | 最新事件自动滚动 + 当前步骤事件高亮 | 普通日志样式 |

**视觉示意：**

```
执行序列：
  ✅ analyze      12s
  ✅ design       18s
  ✅ if(>80)      <1ms
▶ ⏸ implement    ← 当前   ← 整行蓝色高亮背景 + ▶箭头
  ⬜ review
  ⬜ deploy
```

**高亮行为跟随调试动作：**

| 动作 | 高亮变化 |
|------|--------|
| Step Over (F10) | 高亮从当前step跳到下一个step（带平滑过渡动画） |
| Continue (F5) | 高亮快速扫过中间步骤，停在下一个断点 |
| Step Into (F11) | 高亮进入子流程的第一个step |
| 点击其他步骤 | 数据检查器切换，但高亮不移动（高亮 ≠ 选中，就像IDE里光标位置 ≠ 当前执行行） |

**高亮 vs 选中的区别：**

- **高亮（Highlight）** = 当前执行位置，▶ 箭头，蓝色/黄色背景。只有调试引擎能移动它。
- **选中（Selected）** = 用户点击查看的步骤，虚线边框。用户随意点击，数据面板跟随。
- 两者可以不同：用户可以点击已完成的步骤查看数据，同时高亮仍在当前暂停位置。

### 5.1 调试工具栏

流程设计器顶部，调试模式下显示工具栏：

```
┌──────────────────────────────────────────────────────────────────────┐
│ 🐛 调试模式  │ ▶ Continue(F5) │ ⏭ Step Over(F10) │ ⏬ Step Into(F11) │
│              │ ⏫ Step Out    │ ⏹ Stop           │ 🔄 Restart        │
│              │                                                        │
│  状态: ⏸ 暂停在 [implement] (断点)  │  已执行: 3/8步  │  耗时: 2m 15s  │
└──────────────────────────────────────────────────────────────────────┘
```

### 5.2 流程图上的调试状态

```
┌──────────┐     ┌──────────┐     ◇──────◇     🔴┌──────────┐
│  分析需求  │ ──→ │  技术设计  │ ──→ ╱ 分数>80 ╲ ──→ │   开发    │
│ ✅ 12min  │     │ ✅ 8min   │     ◇──────◇     │ ⏸ 暂停中  │
└──────────┘     └──────────┘         │ No        └──────────┘
   已完成            已完成            ↓              ← 当前位置
                                ┌──────────┐      断点(红点)
                                │  返工     │
                                │ ⬜ 未执行  │
                                └──────────┘
```

**节点状态指示：**
| 状态 | 视觉 | 说明 |
|------|------|------|
| 已完成 | ✅ 绿色边框 + 耗时 | 执行成功 |
| 暂停中 | ⏸ 蓝色脉冲边框 | 当前位置 |
| 有断点 | 🔴 左侧红点 | 将在此暂停 |
| 已钉住 | 📌 右上角图标 | 使用钉住数据 |
| 未执行 | ⬜ 灰色 | 尚未到达 |
| 已跳过 | ⏭ 灰色虚线 | 被Skip |
| 数据已修改 | ✏️ 橙色标记 | Edit & Continue |

### 5.3 数据面板

暂停时，右侧显示数据检查面板（类比IDE的Variables面板）：

```
┌─────────────────────────────────┐
│ 📊 数据检查器 — implement       │
├─────────────────────────────────┤
│ ▼ 输入 (Input)                  │
│   prompt: "实现用户登录API"  ✏️  │
│   workdir: "/home/claw/..."     │
│   model: "opus"                 │
│                                 │
│ ▼ 变量来源 (Variables)          │
│   analyze.data.score: 85        │
│   analyze.data.requirement:     │
│     "实现用户登录API"            │
│                                 │
│ ▼ 上游输出 (Upstream)           │
│   ▶ analyze (✅ completed)       │
│   ▶ quality_check (✅ completed) │
│                                 │
│ ▼ 执行路径 (Call Stack)         │
│   → analyze          12s        │
│   → quality_check    5ms        │
│   → implement        ⏸ 当前     │
└─────────────────────────────────┘
```

点击 ✏️ 可以编辑数据值（Edit & Continue）。

## 6. API设计

### 6.1 调试会话管理

```
# 创建调试实例
POST /api/instances
  body: { process_id, mode: "debug", input?, mock?, mock_data?, breakpoints? }
  → { instance_id, debug_session_id }

# 从失败实例进入调试
POST /api/instances/{id}/debug
  body: { from_step? }
  → { debug_session_id }
```

### 6.2 断点管理

```
# 设置断点
PUT /api/debug/{session_id}/breakpoints
  body: { breakpoints: [{ step_id, type, condition? }] }

# 切换单个断点
POST /api/debug/{session_id}/breakpoints/toggle
  body: { step_id }
```

### 6.3 执行控制

```
# 执行动作
POST /api/debug/{session_id}/action
  body: { action: "continue" | "step_over" | "step_into" | "step_out" | "run_to" | "restart" | "stop" | "skip", target_step? }
```

### 6.4 数据操作

```
# 检查数据
GET /api/debug/{session_id}/inspect?step={step_id}

# 修改数据
POST /api/debug/{session_id}/edit
  body: { step_id, edit_type: "input" | "output", data }

# 钉住数据
POST /api/debug/{session_id}/pin
  body: { step_id, pinned_output }

# 取消钉住
DELETE /api/debug/{session_id}/pin/{step_id}
```

### 6.5 实时推送

调试状态变化通过WebSocket/SSE推送给前端：

```
Event: step_started      { step_id, input }
Event: step_completed    { step_id, output, duration_ms }
Event: step_paused       { step_id, reason: "breakpoint" | "step_over" | "user_skip" }
Event: control_evaluated { step_id, condition, result, activated_step }
Event: debug_ended       { reason: "completed" | "stopped" | "error" }
```

## 7. 引擎实现

### 7.1 调试状态机

```
正常模式的step生命周期：
  waiting → pending → running → completed/failed

调试模式新增 paused 状态：
  waiting → pending → [paused_before] → running → [paused_after] → completed/failed
                          ↑                              ↑
                     断点/单步暂停                   after断点暂停
```

### 7.2 引擎调度逻辑（调试增强）

```go
func (e *Engine) scheduleStep(ctx context.Context, step *Step) {
    // 正常模式：直接执行
    if !e.isDebugMode() {
        e.executeStep(ctx, step)
        return
    }

    // 调试模式：检查断点
    if e.hasBreakpoint(step.ID, "before") || e.isStepMode() {
        e.pauseAt(step.ID, "paused_before")
        e.waitForDebugAction(ctx)   // 阻塞等待用户指令
    }

    // 检查是否钉住
    if pinned := e.getPinnedData(step.ID); pinned != nil {
        e.completeWithData(step.ID, pinned)
        return
    }

    // 检查数据是否被编辑
    if edited := e.getEditedInput(step.ID); edited != nil {
        step.Input = edited
    }

    // 执行
    e.executeStep(ctx, step)

    // 执行后检查断点
    if e.hasBreakpoint(step.ID, "after") {
        e.pauseAt(step.ID, "paused_after")
        e.waitForDebugAction(ctx)
    }
}
```

### 7.3 调试会话存储

调试会话是临时的，存内存即可（进程重启后丢失没关系）：

```go
type DebugSession struct {
    ID           string
    InstanceID   string
    Breakpoints  map[string]Breakpoint    // step_id → breakpoint
    PinnedData   map[string]interface{}   // step_id → pinned output
    EditedInput  map[string]interface{}   // step_id → edited input
    StepMode     bool                     // 当前是否单步模式
    CurrentStep  string                   // 当前暂停在哪个step
    ActionChan   chan DebugAction          // 用户指令通道
}
```

## 8. 实施计划

| 阶段 | 内容 | 工作量 | 价值 |
|------|------|--------|------|
| **Phase 1** | 断点 + Continue/StepOver + 基础暂停 | 后端2天 | 核心调试能力 |
| **Phase 2** | 数据检查器（inspect API） | 后端1天 | 看到中间数据 |
| **Phase 3** | 前端调试工具栏 + 流程图状态 | 前端3天 | 可视化调试 |
| **Phase 4** | Edit & Continue | 后端1天 + 前端1天 | 杀手级功能 |
| **Phase 5** | Data Pinning | 后端1天 + 前端1天 | 节省调试时间 |
| **Phase 6** | Mock模式 | 后端1天 | 逻辑测试 |
| **Phase 7** | 条件断点 | 后端半天 | 高级调试 |
| **Phase 8** | WebSocket实时推送 | 前后端各1天 | 流畅体验 |
| **Phase 9** | Step Into/Out（subprocess） | 后端1天 | 子流程调试 |

**总计约 13-15天。Phase 1-3 约6天出第一个可用的调试器。**

## 9. 完整调试示例

用户设计了一个"代码审查与部署"流程，想调试分支逻辑：

### Step 1：设置断点
在 `quality_check`（control: if）和 `deploy` 上设置断点。

### Step 2：启动调试
点击 🐛 调试按钮，输入测试数据 `{ "pr_number": 42 }`。

### Step 3：analyze执行
引擎执行 analyze 步骤（AI分析代码），完成后显示输出：
```json
{ "score": 75, "max_severity": "warning", "issues": [...] }
```

### Step 4：暂停在quality_check
断点命中，用户看到：
- **输入**：condition = `{{analyze.data.score}} >= 80`
- **变量**：`analyze.data.score` = 75
- **评估结果**：`75 >= 80` = false → 将走 `severity_route`

### Step 5：Edit & Continue
用户想测试"如果分数够高会怎样"，修改 `analyze.data.score` 为 90。
点击 Continue。

### Step 6：观察新路径
condition重新评估：`90 >= 80` = true → 走 `auto_merge` 分支。
用户验证了高分路径的逻辑正确。

### Step 7：暂停在deploy（第二次调试）
用户 Restart，这次不改数据。分数75走了 `severity_route` → `human_review`。
人工审批步骤暂停（Mock模式下直接返回approved=true）。
到 `deploy` 断点暂停，用户检查部署命令是否正确。

### Step 8：Continue到结束
确认没问题，点Continue运行到结束。调试完成。

**整个过程就像在VS Code里按F5/F10调试代码一样自然。**
