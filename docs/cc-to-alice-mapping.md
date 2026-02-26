# CC System Prompt → Alice OpenClaw 文件映射分析

> 目标：把 CC 的 system prompt 拆解到 Alice 的 OpenClaw workspace 对应文件中
> 原则：按照 Agent 进化架构的文件升级决策矩阵来分配

## CC Prompt 内容拆解

### 1. 身份定义
**CC 原文：**
> "You are an interactive agent that helps users with software engineering tasks."

**应放位置：** `IDENTITY.md`
**理由：** 身份定义，每次 session 自动注入。Alice 已有 IDENTITY.md，需要强化"资深开发工程师"定位。

---

### 2. 安全边界规则
**CC 原文：**
> "Assist with authorized security testing... Refuse requests for destructive techniques..."
> "You must NEVER generate or guess URLs..."

**应放位置：** `AGENTS.md` → Safety 板块
**理由：** 跨所有任务的硬性规则，属于 AGENTS.md 范畴。OpenClaw 的 AGENTS.md 已有 Safety 板块，补充编码安全相关规则即可。

---

### 3. 编码原则（核心！）
**CC 原文：**
> - "Do not propose changes to code you haven't read"
> - "Do not create files unless absolutely necessary"
> - "Avoid over-engineering"
> - "Don't add features beyond what was asked"
> - "Don't add error handling for scenarios that can't happen"
> - "Don't create helpers for one-time operations"
> - "Be careful not to introduce security vulnerabilities (OWASP top 10)"
> - "Avoid backwards-compatibility hacks"

**应放位置：** `skills/coding/SKILL.md`
**理由：** 这是"某类任务"的指引，不是所有任务都需要。当 Alice 执行开发任务时加载，其他任务（文档、沟通）不需要这些规则。符合 Skill 的定义——按需加载的专业指引。

---

### 4. 谨慎操作规则
**CC 原文：**
> - "Carefully consider the reversibility and blast radius of actions"
> - "Destructive operations: deleting files/branches, dropping database tables..."
> - "Hard-to-reverse operations: force-pushing, git reset --hard..."
> - "Actions visible to others: pushing code, creating PRs..."
> - "Do not use destructive actions as a shortcut"

**应放位置：** `AGENTS.md` → Safety 板块
**理由：** 这些是跨所有任务的通用安全规则。不论执行什么任务，"操作前想清楚后果"都适用。属于 agent 的"操作系统"级规则。

---

### 5. 工具使用偏好
**CC 原文：**
> - "Do NOT use Bash when a dedicated tool is provided"
> - "Use Read instead of cat/head/tail"
> - "Use Edit instead of sed/awk"
> - "Use Glob instead of find/ls"
> - "Use Grep instead of grep/rg"
> - "Break down work with TodoWrite"
> - "Use Task tool for parallelizing independent queries"

**应放位置：** `AGENTS.md` → Tools 使用规范
**理由：** 工具使用偏好是跨所有任务的行为规范。但需要翻译成 OpenClaw 的工具名：
- CC `Read` → OpenClaw `read`（一样）
- CC `Edit` → OpenClaw `edit`（一样）
- CC `Write` → OpenClaw `write`（一样）
- CC `Bash` → OpenClaw `exec`
- CC `Glob`/`Grep` → OpenClaw `exec` (ls/find/grep) 或者内置功能
- CC `Task` → OpenClaw `sessions_spawn`
- CC `TodoWrite` → 无直接对应，可用 memory 文件替代

---

### 6. 输出风格
**CC 原文：**
> - "Only use emojis if the user explicitly requests it"
> - "Responses should be short and concise"
> - "When referencing code include file_path:line_number"
> - "Do not use a colon before tool calls"

**应放位置：** `SOUL.md`
**理由：** 沟通风格和个性定义，属于 SOUL.md 范畴。Alice 可以保留自己的风格，不需要完全照搬 CC 的"冷淡"风格。选择性采纳有用的（如代码引用格式）。

---

### 7. Memory 系统
**CC 原文：**
> - "You have a persistent auto memory directory"
> - "MEMORY.md is always loaded into your system prompt"
> - "Create separate topic files for detailed notes"
> - "Update or remove outdated memories"

**应放位置：** `AGENTS.md` → Memory Management 板块
**理由：** Alice 的 AGENTS.md 已经有记忆管理规则（我们之前写的）。CC 的记忆规则跟我们设计的几乎一模一样！说明我们的设计方向是对的。可以对比补充细节。

---

### 8. 环境信息
**CC 原文：**
> - Working directory, platform, shell, OS version
> - Git status, branch, recent commits

**应放位置：** `TOOLS.md`
**理由：** 环境特定的信息。OpenClaw 已经自动注入 runtime 信息。TOOLS.md 放项目特定的路径、编译命令等。

---

## 映射总结

| CC Prompt 板块 | OpenClaw 文件 | 加载方式 | 优先级 |
|---------------|-------------|---------|-------|
| 身份定义 | `IDENTITY.md` | 自动注入 | ✅ 已有 |
| 安全边界 | `AGENTS.md` Safety | 自动注入 | ⚡ 需补充 |
| **编码原则** | `skills/coding/SKILL.md` | **按需加载** | 🔴 需创建 |
| 谨慎操作 | `AGENTS.md` Safety | 自动注入 | ⚡ 需补充 |
| 工具使用偏好 | `AGENTS.md` Tools | 自动注入 | ⚡ 需补充 |
| 输出风格 | `SOUL.md` | 自动注入 | ✅ 选择性采纳 |
| Memory 系统 | `AGENTS.md` Memory | 自动注入 | ✅ 已有 |
| 环境信息 | `TOOLS.md` + runtime | 自动注入 | ✅ 已有 |

## 关键决策

### 为什么编码原则放 Skill 而不是 AGENTS.md？

根据文件升级决策矩阵：
- **频率**：只在开发任务时需要
- **范围**：只适用于编码场景
- → **特定任务类型 = Skills**

如果放 AGENTS.md，Alice 在做非编码任务（写文档、分析需求）时也会被这些规则影响（如"不要过度工程"可能导致文档不够详细）。Skill 的按需加载正好解决这个问题。

### 哪些内容不需要搬？

1. **CC 特有的工具提示**（如 Glob/Grep 的具体用法）→ OpenClaw 工具不同，不适用
2. **Permission mode 相关**（用户审批工具调用）→ Alice 用 `--dangerously-skip-permissions` 等效模式
3. **CC 帮助信息**（/help、GitHub issues 链接）→ 不适用
4. **Plan mode**（EnterPlanMode/ExitPlanMode）→ 可以用 memory 文件替代
5. **CC 版本/billing header** → 不适用

## 执行计划

1. **创建 `skills/coding/SKILL.md`** — 编码原则核心，从 CC prompt 提炼
2. **更新 Alice `AGENTS.md`** — 补充安全边界 + 工具使用规范 + 谨慎操作规则
3. **可选：更新 Alice `SOUL.md`** — 采纳代码引用格式等有用的风格规则
4. **测试** — 同一任务对比 CC vs Alice+coding skill
