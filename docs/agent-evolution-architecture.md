# Agent进化架构设计 v1.0

> 作者: 泽斌 & Lyra | 日期: 2026-02-22
> 状态: 设计稿，待分步实现

## 核心理念

Agent的能力由三条线决定：**认知 × 知识 × 工具**。三者通过继承和组合构成完整的Agent能力体系，并通过飞轮效应实现自我进化。

```
Agent = 认知(Prompt) + 知识(Knowledge) + 能力(Tools)
```

## 一、认知体系（我是谁、怎么思考）

通过模板继承链构建，从通用到具体逐层覆盖。

```
BaseAgent 模板
  "你是AI助手，遵守输出规范，JSON必须valid..."
  └── PM角色模板（继承BaseAgent）
  │     "你是产品经理，擅长需求分析和拆解..."
  │     └── PM实例（继承PM角色）
  │           "你叫小明，负责nimo产品线..."
  ├── UX角色模板
  │     └── UX实例
  ├── Dev角色模板
  │     └── Dev实例
  └── COO角色模板
        └── Lyra实例
```

**载体文件：**
- `AGENTS.md` — 行为规则（继承链上各层叠加）
- `IDENTITY.md` — 身份定义（实例层）
- `SOUL.md` — 性格与风格（可继承可覆盖）

**数据模型：**
```
AgentTemplate {
  id            string
  parent_id     string     // 继承链，null表示根模板
  name          string     // "BaseAgent" / "PM" / "UX"
  role_type     string     // 角色类型标识
  base_prompt   text       // 该层级的prompt片段
  files[]       {name, content}  // AGENTS.md, IDENTITY.md等
  version       string
}
```

## 二、知识体系（我知道什么）

按作用域和粒度两个维度组织，运行时按需匹配注入。

### 作用域维度（谁能用）

| 层级 | 说明 | 示例 |
|------|------|------|
| 全局通用 | 所有agent都适用 | "输出JSON必须valid" |
| 角色类型通用 | 同类角色适用（所有PM） | "拆需求要考虑前后端分离" |
| 实例专属 | 只有特定agent适用 | "我负责nimo产品线" |

### 粒度维度（什么时候用）

| 层级 | 说明 | 示例 |
|------|------|------|
| 任何时候 | 不限场景 | "API调用要带token" |
| 流程通用 | 跑特定流程时适用 | "功能优化流程要考虑回归测试" |
| Step通用 | 执行特定类型step时适用 | "需求拆解step上次漏了边界case" |

### 知识继承与匹配

Step执行时，引擎沿两个维度做匹配合并：

```
注入的知识 = 
  全局知识
  + 角色类型知识[当前agent角色]
  + 流程知识[当前workflow]
  + Step知识[当前step]
```

越具体的优先级越高，冲突时子级覆盖父级。

### 数据模型

```
KnowledgeEntry {
  id              string
  scope_type      enum     // "global" | "role_type" | "workflow" | "step"
  scope_id        string   // "" | "pm" | "flow-uuid" | "step-name"
  content         text     // 教训/知识内容
  source_run_id   string   // 从哪次运行产生
  confidence      int      // 被验证次数（重复出现次数）
  created_at      datetime
  updated_at      datetime
}
```

### 知识升级机制（系统驱动，非agent自觉）

```
L1(Step级) → L2(流程/角色级) → L3(全局)

触发条件：
- L1→L2: 同类step在3+次运行中出现相同教训 → 生成升级提案
- L2→L3: 多个角色类型都有相同教训 → 生成升级提案
- 所有升级需要人工审批
```

## 三、工具体系（我能做什么）

不同角色拥有不同的工具集，工具也有继承关系。

```
基础工具集（所有agent）:
  web_search, web_fetch, read, write, exec, memory_search, message...

PM角色工具集（继承基础）:
  + feishu_task, feishu_doc, acp_create_task, acp_list_tasks...

UX角色工具集（继承基础）:
  + browser, screenshot, lighthouse, web-qa-bot...

Dev角色工具集（继承基础）:
  + github, coding-agent, docker, deploy-agent...

COO角色工具集（继承基础）:
  + 全部管理工具, 跨agent调度, 数据分析...

实例专属工具:
  + 特定系统API, 内部工具...
```

**关键原则：所有工具封装为MCP Server，Plugin仅用于消息通道。**

### Plugin vs MCP 的分界线

```
Plugin = 消息通道（Channel）
  OpenClaw的感官系统，管理长连接生命周期
  必须在进程内：Feishu WebSocket、Discord Gateway、Telegram polling
  数量极少，只有通道类

MCP Server = 工具（Tool）
  Agent调用的所有能力
  独立进程，标准协议(stdio/HTTP)，可移植
  好处：Cursor/Claude Desktop等任何MCP client都能复用
  鉴权：本机stdio模式零配置，HTTP模式一次配token
```

Plugin唯一的"优势"是免鉴权（进程内天然信任），但这个成本可忽略。
MCP的可移植性和标准化收益远大于这点便利。

**工具开发策略：新工具一律MCP Server，不做Plugin。**

已有的MCP Server：`cmd/mcp-plm`、`cmd/mcp-erp`
待迁移：ACP当前是Plugin，未来应迁移为MCP Server

**载体：** `TOOLS.md`（配置说明） + MCP Server定义 + OpenClaw config中的skill

**当前限制：** OpenClaw暂不支持per-agent tool限制，所有agent共享工具集。未来需要支持：
```yaml
# 理想状态
agents:
  list:
    - id: pm
      tools:
        allow: [feishu_task, feishu_doc, acp_*]
        deny: [github, exec]
    - id: ux
      tools:
        allow: [browser, screenshot]
```

## 四、三大飞轮

### 飞轮1: 任务执行闭环（已跑通）
```
接任务 → 执行 → 交付 → 积累教训
                         ↓ 发现能力缺口
```

### 飞轮2: 知识进化闭环（建设中）
```
教训收集 → 浓缩 → 注入prompt → 验证效果
                                ↓ 发现工具缺口
```

### 飞轮3: 工具进化闭环（未来方向）
```
识别需求 → 设计工具 → 开发 → 测试 → 注册上线
   ↑                                    │
   └──── 能力更强，能接更复杂的任务 ←────┘
```

三个飞轮嵌套驱动，形成**自我进化的复利效应**：
- 教训不过期，工具不消失
- 每转一圈，下一圈更快
- CEO只需定方向和做关键决策，执行层自动越跑越好

## 五、Prompt组装流程

Agent执行某个Step时，引擎的prompt组装逻辑：

```
1. 模板继承链合并 → System Prompt
   BaseAgent.prompt + RoleType.prompt + Instance.prompt

2. 知识库匹配注入 → Context
   Global知识 + RoleType知识 + Workflow知识 + Step知识

3. 工具集确定 → Available Tools
   BaseTools + RoleTools + InstanceTools

4. 上游数据注入 → User Context
   前置step的structured_output + 模板变量渲染

最终 = System Prompt + Context知识 + Tools定义 + 任务指令
```

## 六、文件升级决策规则

Agent进化的本质是精准修改正确的文件。核心判断依据：**频率 × 范围 × 时效性**。

### 决策矩阵

| 特征 | 写入位置 | 注入方式 | 示例 |
|------|---------|---------|------|
| 所有任务都要遵守的铁律 | AGENTS.md | 每次加载 | "调API前验证token" |
| 当前状态/优先级/进行中的事 | MEMORY.md | 每次加载 | "PLM v2.1正在开发" |
| 某领域的详细知识 | bank/*.md | 检索命中 | "PLM API文档详情" |
| 做某类任务的方法论和步骤 | skills/SKILL.md | 匹配时加载 | "如何做PR Review" |
| 工具配置和路径 | TOOLS.md | 每次加载 | "PLM服务地址:8080" |

### 升级路径（教训从产生到沉淀）

```
Step运行产生教训（L1）
  ↓ 出现1次
Step级知识库（ACP存储，step执行时注入）

  ↓ 同类step出现3+次
领域知识 → bank/*.md 或 flow级知识库

  ↓ 跨多个流程反复出现
行为规则 → AGENTS.md

  ↓ 涉及具体操作步骤
方法论 → skills/SKILL.md
```

### 写入约束

- **AGENTS.md**: <8KB，只放规则不放细节，用"详见bank/xxx.md"引用
- **MEMORY.md**: <3KB，只放影响当前行为的信息
- **bank/*.md**: 每个<3KB，按主题拆分
- **skills/SKILL.md**: 无硬限，但保持聚焦单一能力
- **TOOLS.md**: <8KB，纯配置信息

## 七、实现路线图

### Phase 1: 知识注入基础（当前）
- [x] Step级教训收集（acp_complete_step的lessons字段）
- [ ] Step级教训自动注入（渲染prompt时匹配历史教训）
- [ ] 知识库数据模型（scope_type + scope_id）
- [ ] ACP Web知识库管理界面

### Phase 2: 继承与升级
- [ ] 模板继承链（parent_id）
- [ ] 知识升级提案（L1→L2→L3自动检测）
- [ ] 人工审批升级
- [ ] 知识冲突解决（子级覆盖父级）

### Phase 3: 工具体系
- [ ] Per-agent tool限制（需OpenClaw支持或ACP层面实现）
- [ ] 工具使用审计
- [ ] 工具缺口识别（从obstacles字段分析）
- [ ] 工具开发工作流模板

### Phase 4: 自我进化闭环
- [ ] 自动识别需要造的工具
- [ ] 工具开发走ACP工作流
- [ ] 工具自动注册和分配
- [ ] 进化效果度量和报告

---

## 五、记忆巩固机制（Sleep Consolidation）

> 2026-02-23 新增。基于人类神经科学的记忆巩固原理设计。

### 问题

Agent 的 daily log（memory/YYYY-MM-DD.md）相当于人类的短期记忆（海马体），MEMORY.md 相当于长期记忆（大脑皮层）。但目前缺少从短期到长期的**自动巩固机制**——全靠 agent 自觉判断什么该记，什么不该记。

这就像一个**从不睡觉的人**——信息不断输入，但没有整理、筛选、巩固的过程。

### 人类记忆的编码标准（神经科学）

人类大脑在睡眠时通过海马体回放，按以下因素决定什么该转入长期记忆：

| 编码因素 | 神经机制 | Agent 对应 |
|---------|---------|-----------|
| **情绪显著性** | 杏仁核标记 → 优先巩固 | 老板的指令/批评 |
| **目标相关性** | 前额叶评估 → 跟当前目标有关的优先 | 影响当前项目/任务的决策 |
| **重复/频率** | 突触强化（Hebb定律）| 反复出现的模式或问题 |
| **新奇/反直觉** | 多巴胺释放 → 增强编码 | 跟预期不符的发现 |
| **关联密度** | 与已有网络连接越多越牢 | 跟多个项目/决策关联的信息 |

核心原则：**记住 what 和 why，忘掉 how**（具体操作步骤可重新推导）。

### Agent 记忆层次对应

| 人类记忆类型 | Agent 载体 | 加载方式 | 类比 |
|------------|-----------|---------|------|
| 工作记忆（7±2项） | Session 上下文 | 自动（当前对话） | 正在想的事 |
| 情景记忆（具体事件） | `memory/YYYY-MM-DD.md` | memory_search | "那天发生了什么" |
| 语义记忆（知识事实） | `bank/*.md` | memory_search / read | "X 是什么" |
| 程序性记忆（怎么做） | `AGENTS.md` + `skills/` | 自动注入 / 按需加载 | "遇到 X 应该 Y" |
| 长期核心记忆 | `MEMORY.md` | 自动注入（每次） | "我是谁、在做什么" |

### 自动巩固协议（Consolidation Protocol）

每日运行一次（相当于"睡眠巩固"），对前一天的 daily log 进行评分和巩固：

**评分维度（每条信息 0-10 分）：**

1. **Authority（权威性）: 0-3 分**
   - 泽斌的直接指令/规则 → 3分
   - 泽斌的偏好/倾向 → 2分
   - 自己的发现/总结 → 1分
   - 过程性描述 → 0分

2. **Impact（影响性）: 0-3 分**
   - 改变当前工作方式/流程 → 3分
   - 影响某个项目的决策 → 2分
   - 可能未来有用 → 1分
   - 一次性事件 → 0分

3. **Recurrence（再现性）: 0-2 分**
   - 必定会被再次提到 → 2分
   - 可能被提到 → 1分
   - 大概率不会再提 → 0分

4. **Novelty（新奇性）: 0-2 分**
   - 完全颠覆之前认知 → 2分
   - 有一定新意 → 1分
   - 常规/预期内 → 0分

**阈值：总分 ≥ 5 → 必须进 MEMORY.md；3-4 → 进 bank/；≤ 2 → 仅留 daily log**

### 实现方案

通过 cron job 驱动，每日自动执行：

```
Schedule: 每天 UTC 18:00（北京凌晨 2:00，泽斌通常已休息）
Agent: isolated session
Task:
  1. 读取昨天的 memory/YYYY-MM-DD.md
  2. 提取所有独立信息条目
  3. 对每条按 4 个维度评分（Authority + Impact + Recurrence + Novelty）
  4. 总分 ≥ 5 的条目 → 写入/更新 MEMORY.md
  5. 总分 3-4 的条目 → 写入对应 bank/*.md
  6. 检查 MEMORY.md 大小，超 3KB 则精简低分条目
  7. 输出巩固报告
```

### 与人类记忆的关键差异

1. **人类遗忘是渐进的**（艾宾浩斯曲线），agent 的 daily log 不会自动衰减
2. **人类有情绪标记**，agent 需要用"Authority"维度模拟
3. **人类睡眠巩固是无意识的**，agent 需要显式 cron 触发
4. **人类可以通过"回想"强化记忆**，agent 的 memory_search 部分等效

### Phase 规划

- [x] Phase 0: MEMORY.md 准入标准（三问测试）— 已部署到 4 个 agent
- [ ] Phase 1: 每日自动巩固 cron — **当前优先**
- [ ] Phase 2: 多维评分系统（替代简单的三问测试）
- [ ] Phase 3: 跨 agent 记忆共享（shared/ 目录 + ACP lessons 联动）
- [ ] Phase 4: 记忆衰减机制（长期未被 search 命中的信息自动降级）

---

*本文档记录BitFantasy Agent进化架构的完整设计，作为ACP长期发展的北极星方向。*
