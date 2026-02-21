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

**关键原则：所有工具最终封装为OpenClaw plugin/skill。**

裸exec调脚本只是临时方案。Plugin化的好处：
- 可版本管理
- 可分配给特定agent
- 可组合编排
- 可审计调用记录

**载体：** `TOOLS.md`（配置说明） + OpenClaw config中的plugin/skill定义

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

## 六、实现路线图

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

*本文档记录BitFantasy Agent进化架构的完整设计，作为ACP长期发展的北极星方向。*
