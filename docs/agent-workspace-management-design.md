# Agent Workspace 管理 & 企业 Skills 库 — 设计方案

> 作者：Lyra | 日期：2026-02-23 | 状态：待泽斌确认

## 一、核心理念

**Agent = 它的文件。** 每个 agent 的所有 `.md` 文件 + skills 就是它的完整"基因"。版本管理 = 对这些文件做 git snapshot。

两个模块：
1. **Agent Workspace 版本管理** — 每个 agent 独立管理自己的所有文件
2. **企业 Skills 库** — 公司级的 skill 仓库，agent 可以发布/安装 skill

## 二、Agent Workspace 版本管理

### 2.1 管理什么

每个 agent workspace 的所有可版本化文件：

| 分类 | 文件 | 说明 |
|------|------|------|
| **身份** | IDENTITY.md, SOUL.md | 谁，什么性格 |
| **规则** | AGENTS.md, USER.md, TOOLS.md | 行为准则、用户信息、工具配置 |
| **记忆** | MEMORY.md, bank/*.md | 核心记忆、知识库 |
| **日志** | memory/*.md | 日常记录（可选是否纳入版本） |
| **心跳** | HEARTBEAT.md | 定时任务配置 |
| **技能** | skills/*/SKILL.md + references/* | agent 本地安装的 skills |

### 2.2 页面设计

**Agent 详情页 → Workspace 标签页**，一个界面搞定：

```
┌─────────────────────────────────────────────────┐
│  Agent: Lyra (main)                    [提交版本] │
├────────────┬────────────────────────────────────┤
│ 📁 文件树    │  编辑区 (Monaco Editor)              │
│            │                                     │
│ ▼ 身份      │  # SOUL.md                          │
│   IDENTITY │  _You're not a chatbot..._          │
│   SOUL     │                                     │
│ ▼ 规则      │                                     │
│   AGENTS   │                                     │
│   USER     │                                     │
│   TOOLS    │                                     │
│ ▼ 记忆      │                                     │
│   MEMORY   │                                     │
│   bank/    │                                     │
│ ▼ 技能      │                                     │
│   skills/  │                                     │
│ HEARTBEAT  │                                     │
│            │  [保存] [还原]                        │
├────────────┴────────────────────────────────────┤
│ 版本历史                                          │
│ v3  2026-02-23 14:30  "添加PLM审批流程skill"       │
│ v2  2026-02-20 10:00  "更新AGENTS.md编程规则"      │
│ v1  2026-02-15 09:00  "初始版本"                   │
│                                    [对比] [回滚]   │
└─────────────────────────────────────────────────┘
```

### 2.3 版本提交机制

- **提交** = 对 workspace 所有文件做一次 snapshot（存数据库，JSON blob）
- **回滚** = 把某个版本的文件覆写回 workspace
- **对比** = 两个版本间的 diff 展示
- 每次提交需要填写版本说明（类似 commit message）
- 日志文件（memory/YYYY-MM-DD.md）默认不纳入版本快照（可配置）

### 2.4 后端设计

```
// 新表
agent_workspace_versions
├── id (uuid)
├── agent_id (关联 agents 表)
├── version (自增整数)
├── message (版本说明)
├── snapshot (JSONB: { "files": { "SOUL.md": "content...", "skills/coding/SKILL.md": "..." } })
├── file_count (文件数)
├── total_size (总字节数)
├── created_at
└── created_by (谁提交的：human / agent自身)
```

### 2.5 API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/agents/:id/workspace | 获取 workspace 文件树 + 内容 |
| PUT | /api/agents/:id/workspace/files | 保存单个文件 |
| POST | /api/agents/:id/workspace/commit | 提交新版本（snapshot 所有文件） |
| GET | /api/agents/:id/workspace/versions | 版本列表 |
| GET | /api/agents/:id/workspace/versions/:ver | 某版本详情 |
| POST | /api/agents/:id/workspace/versions/:ver/rollback | 回滚到某版本 |
| GET | /api/agents/:id/workspace/diff?from=v1&to=v2 | 版本对比 |

---

## 三、企业 Skills 库

### 3.1 概念

类似 ClawHub，但是企业私有的：
- Agent 在工作中进化出有用的 skill → 可以**发布**到企业库
- 其他 agent 可以从企业库**安装** skill
- 人工也可以直接在库里创建/编辑 skill

### 3.2 Skill 结构（遵循 OpenClaw 标准）

```
skill-name/
├── SKILL.md          # 核心：描述 + 指令（必须）
└── references/       # 可选：参考文件
    ├── example.md
    └── config.json
```

### 3.3 页面设计

**企业 Skills 库页面：**

```
┌─────────────────────────────────────────────────┐
│  🏢 企业 Skills 库                    [创建 Skill] │
├─────────────────────────────────────────────────┤
│ 🔍 搜索                                          │
├─────────────────────────────────────────────────┤
│                                                   │
│  ┌──────────────┐  ┌──────────────┐              │
│  │ 📦 coding     │  │ 📦 pr-reviewer │              │
│  │ v2 · Lyra发布 │  │ v1 · PM发布   │              │
│  │ 已安装: 3/4   │  │ 已安装: 1/4   │              │
│  │ [详情] [分发]  │  │ [详情] [分发]  │              │
│  └──────────────┘  └──────────────┘              │
│                                                   │
│  ┌──────────────┐  ┌──────────────┐              │
│  │ 📦 deploy     │  │ 📦 git-flow   │              │
│  │ v1 · Alice发布│  │ v3 · 手动创建  │              │
│  │ 已安装: 2/4   │  │ 已安装: 4/4   │              │
│  └──────────────┘  └──────────────┘              │
│                                                   │
└─────────────────────────────────────────────────┘
```

**Skill 详情页：**

```
┌─────────────────────────────────────────────────┐
│  📦 coding (v2)                 来源: Lyra 发布   │
│  "编码规范和最佳实践"                               │
├─────────────────────────────────────────────────┤
│  SKILL.md 编辑区 (Monaco)                         │
│  ┌─────────────────────────────────────────┐    │
│  │ ---                                      │    │
│  │ name: coding                             │    │
│  │ description: 编码规范...                  │    │
│  │ ---                                      │    │
│  │ # Coding Skill                           │    │
│  │ ...                                      │    │
│  └─────────────────────────────────────────┘    │
│                                                   │
│  📎 References: example.md, config.json           │
│                                                   │
│  安装状态:                                         │
│  ✅ Lyra  ✅ PM  ❌ UX  ✅ Alice                  │
│                                                   │
│  [保存] [发布新版本] [分发到所有Agent] [删除]        │
├─────────────────────────────────────────────────┤
│  版本历史                                         │
│  v2  2026-02-23  "优化编码规则，增加Go最佳实践"     │
│  v1  2026-02-20  "初始版本"                        │
└─────────────────────────────────────────────────┘
```

### 3.4 工作流

```
Agent工作中进化出skill
        │
        ▼
  [发布到企业库]  ──→  企业Skills库(数据库)
                            │
                     管理员/Agent浏览
                            │
                    [安装到指定Agent]
                            │
                     写入agent workspace
                     skills/skill-name/SKILL.md
```

### 3.5 后端设计

```
// 新表
enterprise_skills
├── id (uuid)
├── name (唯一，skill目录名)
├── description
├── current_version (整数)
├── created_by (agent_id 或 "human")
├── created_at
└── updated_at

enterprise_skill_versions
├── id (uuid)
├── skill_id (FK)
├── version (整数)
├── message (版本说明)
├── files (JSONB: { "SKILL.md": "content...", "references/example.md": "..." })
├── created_by
└── created_at

enterprise_skill_installs
├── id (uuid)
├── skill_id (FK)
├── agent_id (FK)
├── installed_version (整数)
└── installed_at
```

### 3.6 API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/skills | 企业 skill 列表 |
| POST | /api/skills | 创建新 skill |
| GET | /api/skills/:id | skill 详情 |
| PUT | /api/skills/:id | 更新 skill 内容 |
| DELETE | /api/skills/:id | 删除 skill |
| POST | /api/skills/:id/publish | 发布新版本 |
| GET | /api/skills/:id/versions | 版本列表 |
| POST | /api/skills/:id/install | 安装到指定 agent（写入 workspace） |
| POST | /api/skills/:id/distribute | 分发到所有 agent |
| POST | /api/agents/:id/skills/publish | Agent 从自己 workspace 发布 skill 到企业库 |

---

## 四、要删除的

1. **角色模板模块**（agent-templates CRUD + 版本管理）— 整个删掉
   - 前端：AgentTemplates 相关页面/路由/API
   - 后端：template_handler.go, template_service.go, template 相关表
   
**保留：**
- ✅ 共享模板页面（BootstrapTemplate.tsx + bootstrap_handler.go）— 管理共享 AGENTS.md
- ✅ bootstrap-extra-files hook — 分发共享规则到所有 agent

---

## 五、菜单结构调整

```
Agent管理
├── Agent列表（已有）
├── Agent详情 → 新增 Workspace 标签页（文件管理 + 版本历史）
├── 共享模板（已有，保留）
└── 企业Skills库（新页面）
```

---

## 六、不做什么（保持简单）

- ❌ 不做 skill 自动发现/推荐 — 人工或 agent 主动发布
- ❌ 不做 skill 依赖管理 — 每个 skill 独立
- ❌ 不做跨企业 skill 市场 — 只管企业内部
- ❌ 不做自动版本提交 — 必须人工或 agent 主动触发
- ❌ 不做文件锁/并发编辑 — 单人操作场景
- ❌ memory/日志文件默认不纳入版本 — 太大且变化频繁
