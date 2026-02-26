# Agent Skills ↔ 企业 Skills 库 同步设计

> 2026-02-23 | 状态：方案讨论

## 1. 核心概念

```
┌─────────────────────────────────────────────────────┐
│                   企业 Skills 库                      │
│  (中央仓库，类似 private npm registry)                │
│                                                       │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐           │
│  │ coding   │  │ pr-review│  │ deploy   │  ...      │
│  │ v3       │  │ v2       │  │ v1       │           │
│  └──────────┘  └──────────┘  └──────────┘           │
└──────────┬──────────────────────┬────────────────────┘
           │                      │
      install v3             install v2
           │                      │
           ▼                      ▼
┌──────────────────┐  ┌──────────────────┐
│   Alice 的 Skills │  │   PM 的 Skills   │
│                    │  │                    │
│  coding/ ★v3      │  │  pr-review/ ★v2   │
│  deploy/ v1       │  │  coding/ v3       │
│  tdd/    (本地)   │  │                    │
│                    │  │                    │
│ ★ = 已修改未发布   │  │                    │
└──────────────────┘  └──────────────────┘
```

**两个空间，三种状态：**

| 状态 | 含义 | 标识 |
|------|------|------|
| **synced** | 本地 = 企业库版本，完全同步 | ✅ |
| **modified** | 本地有修改，尚未发布回企业库 | ⚠️ |
| **local-only** | 仅存在于本地，从未发布过 | 🏠 |

## 2. 数据模型变化

### 2.1 新增：Agent Skill 安装跟踪表

现有 `enterprise_skill_installs` 只记录安装动作，缺少**当前状态**追踪。改为：

```sql
-- 改造现有表，增加状态追踪
ALTER TABLE enterprise_skill_installs ADD COLUMN status TEXT DEFAULT 'synced';
-- status: synced | modified | uninstalled
ALTER TABLE enterprise_skill_installs ADD COLUMN local_hash TEXT;
-- 本地文件内容的hash，用于检测本地修改
ALTER TABLE enterprise_skill_installs ADD COLUMN enterprise_version INT;
-- 安装时的企业库版本号
```

### 2.2 EnterpriseSkill 增加来源追踪

```sql
ALTER TABLE enterprise_skills ADD COLUMN source_agent TEXT;
-- 首次发布的agent
ALTER TABLE enterprise_skills ADD COLUMN install_count INT DEFAULT 0;
-- 安装次数统计
```

## 3. 用户交互流程

### 3.1 Agent 详情页 → Skills Tab（新增）

从 Workspace Tab 中把 skills 单独拎出来，作为独立 tab：

```
┌──────────────────────────────────────────────────────┐
│  Agent: Alice                                         │
│  ┌──────┬──────────┬────────┬─────────┬────────────┐ │
│  │ 概览 │ Workspace│ Skills │ 版本历史 │ 对话记录   │ │
│  └──────┴──────────┴────────┴─────────┴────────────┘ │
│                                                        │
│  Skills 管理                         [从企业库安装...]  │
│  ┌─────────────────────────────────────────────────┐  │
│  │ 🏠 tdd-guide          本地  │ [发布] [删除]      │  │
│  │ ✅ coding       v3 = 企业v3 │ [删除]             │  │
│  │ ⚠️ pr-review    v2 → 本地改 │ [发布] [还原] [删除]│  │
│  │ ✅ deploy       v1 = 企业v1 │ [升级v2可用] [删除] │  │
│  └─────────────────────────────────────────────────┘  │
│                                                        │
│  [从企业库安装...] → 弹窗：                              │
│  ┌──────────────────────────────────────────────┐     │
│  │ 从企业 Skills 库安装               [确认安装]  │     │
│  │                                                │     │
│  │ ☑ coding      v3  "编码规范"      (已安装,最新) │     │
│  │ ☑ pr-review   v2  "PR审查"       (已安装,最新) │     │
│  │ ☐ deploy      v2  "部署流程"      (未安装)     │     │
│  │ ☐ git-flow    v1  "Git工作流"     (未安装)     │     │
│  │                                                │     │
│  │ 已选: 2 个新 skill                             │     │
│  └──────────────────────────────────────────────┘     │
└──────────────────────────────────────────────────────┘
```

### 3.2 操作说明

| 操作 | 触发条件 | 效果 |
|------|---------|------|
| **发布** | 本地有skill（local-only 或 modified） | 读取本地文件 → 创建/更新企业库skill + 新版本 |
| **安装** | 企业库有skill，本地没有 | 企业库文件 → 写入本地 skills/ 目录 |
| **升级** | 企业库有新版本，本地版本较旧 | 企业库最新版 → 覆盖本地文件 |
| **还原** | 本地已修改，想恢复到企业库版本 | 企业库对应版本 → 覆盖本地文件 |
| **删除** | 任何状态 | 删除本地文件，标记安装记录为 uninstalled |
| **Diff** | modified 状态 | 对比本地文件 vs 企业库版本 |

### 3.3 企业 Skills 库页面

```
┌──────────────────────────────────────────────────────┐
│  企业 Skills 库                                       │
│                                                        │
│  ┌─────────────┬───────┬────────┬──────┬────────────┐ │
│  │ 名称         │ 版本  │ 来源   │ 安装 │ 操作        │ │
│  ├─────────────┼───────┼────────┼──────┼────────────┤ │
│  │ coding      │ v3    │ alice  │ 3/4  │ [分发] [详情] │ │
│  │ pr-review   │ v2    │ pm     │ 2/4  │ [分发] [详情] │ │
│  │ deploy      │ v2    │ lyra   │ 1/4  │ [分发] [详情] │ │
│  └─────────────┴───────┴────────┴──────┴────────────┘ │
│                                                        │
│  [分发] = 一键安装到所有agent                           │
└──────────────────────────────────────────────────────┘
```

## 4. API 设计

### 4.1 Agent Skills API（新增）

```
GET    /api/agents/:id/skills
  → 列出agent的所有skill，含同步状态

POST   /api/agents/:id/skills/publish
  → 将本地skill发布到企业库（已有，需增强）
  Body: { skill_name, message }

POST   /api/agents/:id/skills/install
  → 从企业库批量安装skills到本地
  Body: { skills: [{ enterprise_skill_id, version? }, ...] }

POST   /api/agents/:id/skills/:name/restore
  → 还原为企业库版本（丢弃本地修改）

DELETE /api/agents/:id/skills/:name
  → 删除本地skill

GET    /api/agents/:id/skills/:name/diff
  → 对比本地 vs 企业库版本
```

### 4.2 Agent Skills List 返回格式

```json
{
  "skills": [
    {
      "name": "coding",
      "local_files": ["SKILL.md", "references/go-style.md"],
      "sync_status": "synced",       // synced | modified | local-only
      "local_version": null,          // 本地没有版本概念
      "enterprise_skill_id": "xxx",   // 关联的企业skill ID，null=local-only
      "enterprise_version": 3,        // 安装时的企业版本
      "enterprise_latest": 3,         // 企业库最新版本
      "upgradable": false,            // enterprise_latest > enterprise_version
      "description": "..."            // 从SKILL.md提取
    },
    {
      "name": "tdd-guide",
      "local_files": ["SKILL.md"],
      "sync_status": "local-only",
      "enterprise_skill_id": null,
      "enterprise_version": null,
      "enterprise_latest": null,
      "upgradable": false,
      "description": "..."
    }
  ]
}
```

## 5. 同步状态判定逻辑

```
读取本地 skills/ 目录下所有 skill
对每个 skill:
  1. 查 enterprise_skill_installs 是否有安装记录
  2. 如果没有 → local-only
  3. 如果有:
     a. 计算本地文件hash
     b. 对比安装时的hash (local_hash)
     c. hash一致 → synced
     d. hash不一致 → modified
  4. 查企业库最新版本
     如果 latest > installed_version → upgradable=true
```

**Hash计算**：对skill目录下所有文件内容排序拼接后SHA256，简单可靠。

## 6. 发布流程详解

发布分两种场景，前端需要明确区分：

### 6.1 首次发布（local-only → 企业库）

状态：skill从未发布过，企业库无此skill。

```
按钮文案: [发布到企业库]
确认弹窗: "将 tdd-guide 作为新 Skill 发布到企业库？"
         输入: Skill描述（可选）、发布说明

流程:
1. 读取 skills/tdd-guide/ 下所有文件
2. 企业库创建新 EnterpriseSkill + v1
3. 创建 enterprise_skill_installs 记录（status=synced）
4. 本地状态从 🏠 local-only → ✅ synced
```

### 6.2 更新发布（modified → 企业库新版本）

状态：skill已在企业库，本地做了修改。

```
按钮文案: [发布更新]
确认弹窗: "将本地修改发布为 coding v4？（当前企业版本: v3）"
         显示: 变更diff预览
         输入: 更新说明（必填）

流程:
1. 读取本地文件，对比企业库当前版本
2. 展示diff让用户确认
3. 企业库创建新版本 (v3 → v4)
4. 更新 enterprise_skill_installs（version=4, status=synced）
5. 本地状态从 ⚠️ modified → ✅ synced
```

### 6.3 前端按钮逻辑

| 状态 | 按钮 | 样式 |
|------|------|------|
| local-only | [发布到企业库] | primary蓝色 |
| modified | [发布更新 v3→v4] | warning橙色 |
| synced | 无发布按钮 | — |

## 7. 批量安装流程详解

```
从企业库批量安装 skills 到 Alice

请求: POST /api/agents/alice/skills/install
Body: { skills: [
  { enterprise_skill_id: "xxx", version: 2 },
  { enterprise_skill_id: "yyy" }           // 不指定version=最新
]}

对每个skill:
  1. 读取 enterprise_skill_versions 中对应版本的文件内容
  2. 写入 Alice 的 workspace/skills/{name}/ 目录
  3. 创建/更新 enterprise_skill_installs 记录:
     - agent_id = alice
     - installed_version = 对应版本
     - local_hash = 写入后的hash
     - status = synced

返回: { installed: ["pr-review@v2", "git-flow@v1"], failed: [] }
```

前端交互：点"从企业库安装" → 弹窗展示企业库所有skill（checkbox多选）→ 已安装的默认勾选且disabled → 勾选要装的 → 确认 → 批量安装。

## 8. 前端变化

### 8.1 AgentDetail.tsx 新增 Skills Tab

独立于 Workspace Tab，专门管理skills：
- 卡片列表展示每个skill
- 状态徽标（synced/modified/local-only）
- 操作按钮（发布/安装/还原/删除/diff）
- "从企业库安装"弹窗

### 8.2 Workspace Tab 变化

- skills 文件仍然在 workspace 文件树中**可见**（只读参考）
- 但编辑/删除操作引导到 Skills Tab

### 8.3 企业 Skills 库页面（重新设计）

#### 列表页

简洁表格，一眼看清每个skill的状态：

```
┌─────────────────────────────────────────────────────────────────────┐
│  企业 Skills 库                                                      │
│                                                                       │
│  名称          描述              版本   来源    安装情况    操作       │
│  ─────────────────────────────────────────────────────────────────── │
│  coding        编码规范与最佳实践  v3    alice   3/4 ██▓░   [详情]    │
│  pr-review     PR审查清单         v2    pm      2/4 ██░░   [详情]    │
│  deploy        部署流程自动化      v2    lyra    4/4 ████   [详情]    │
│  git-flow      Git工作流规范      v1    alice   1/4 █░░░   [详情]    │
│                                                                       │
│  安装情况: 已安装agent数 / 总agent数  (进度条直观展示覆盖率)           │
└─────────────────────────────────────────────────────────────────────┘
```

**列表字段设计：**

| 字段 | 来源 | 说明 |
|------|------|------|
| 名称 | enterprise_skills.name | skill唯一标识 |
| 描述 | enterprise_skills.description | 一句话说明用途，从SKILL.md的description提取 |
| 版本 | enterprise_skills.current_version | 最新版本号 |
| 来源 | enterprise_skills.source_agent | 首次发布的agent |
| 安装情况 | COUNT(installs) / total_agents | 覆盖率，直观看出哪些skill还没全员安装 |
| 最近更新 | enterprise_skill_versions.created_at | 最新版本的发布时间 |

#### 详情页

点击skill进入详情，三个区域：

```
┌─────────────────────────────────────────────────────────────────┐
│  ← 返回列表                                                      │
│                                                                   │
│  coding — 编码规范与最佳实践                           [全员分发]  │
│  来源: alice | 当前版本: v3 | 创建: 2026-02-20                    │
│                                                                   │
│  ┌─── 安装分布 ───────────────────────────────────────────────┐  │
│  │  alice  ✅ v3 (synced)     最新                             │  │
│  │  pm     ✅ v3 (synced)     最新                             │  │
│  │  ux     ⚠️ v2 (outdated)   可升级v3                         │  │
│  │  main   ➖ 未安装           [安装]                           │  │
│  └─────────────────────────────────────────────────────────────┘  │
│                                                                   │
│  ┌─── 版本历史 ───────────────────────────────────────────────┐  │
│  │  v3  2026-02-22  alice  "增加错误处理规范"      [查看文件]  │  │
│  │  v2  2026-02-18  alice  "补充Go测试规范"        [查看文件]  │  │
│  │  v1  2026-02-15  alice  "首次发布"              [查看文件]  │  │
│  └─────────────────────────────────────────────────────────────┘  │
│                                                                   │
│  ┌─── 文件内容 (v3) ─────────────────────────────────────────┐  │
│  │  📄 SKILL.md                                               │  │
│  │  📁 references/                                            │  │
│  │     📄 go-style.md                                         │  │
│  │     📄 error-handling.md                                   │  │
│  │                                                             │  │
│  │  (点击文件名展开内容预览)                                    │  │
│  └─────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

**详情页三大区块：**

| 区块 | 回答的问题 | 核心数据 |
|------|-----------|---------|
| **安装分布** | 谁装了？谁没装？谁落后了？ | 每个agent的安装版本 + 状态(synced/outdated/未安装) |
| **版本历史** | 这个skill怎么演进的？ | 版本号 + 时间 + 发布者 + 更新说明 |
| **文件内容** | 这个skill具体是什么？ | 文件树 + 内容预览（当前版本） |

#### 数据模型完善

```sql
-- EnterpriseSkill 增加字段
enterprise_skills:
  id              TEXT PRIMARY KEY
  name            TEXT UNIQUE NOT NULL    -- skill名称（目录名）
  description     TEXT                     -- 一句话描述
  current_version INT DEFAULT 1            -- 最新版本号
  source_agent    TEXT                     -- 首次发布的agent
  created_by      TEXT                     -- 发布人（agent名）
  created_at      TIMESTAMP
  updated_at      TIMESTAMP

-- EnterpriseSkillVersion 不变
enterprise_skill_versions:
  id              TEXT PRIMARY KEY
  skill_id        TEXT NOT NULL            -- → enterprise_skills.id
  version         INT NOT NULL
  message         TEXT                     -- 发布说明（必填）
  files           TEXT NOT NULL            -- JSON: {"SKILL.md":"...","references/x.md":"..."}
  created_by      TEXT                     -- 哪个agent发布的
  created_at      TIMESTAMP

-- EnterpriseSkillInstall 增强
enterprise_skill_installs:
  id                TEXT PRIMARY KEY
  skill_id          TEXT NOT NULL          -- → enterprise_skills.id
  agent_id          TEXT NOT NULL
  installed_version INT                    -- 安装的企业库版本
  local_hash        TEXT                   -- 安装时的文件hash
  status            TEXT DEFAULT 'synced'  -- synced | modified | uninstalled
  installed_at      TIMESTAMP
  UNIQUE(skill_id, agent_id)              -- 每个agent每个skill只有一条记录
```

## 9. 不做的事

- ❌ 不做自动同步（所有操作都是显式的，人工触发）
- ❌ 不做冲突合并（还原=覆盖，发布=新版本，不merge）
- ❌ 不做skill依赖管理（太复杂，当前不需要）
- ❌ 不改企业skill的直接创建功能（保留从UI直接创建的入口）

## 10. 实现优先级

1. **P0**: Agent Skills Tab（列表+状态） + 发布 + 删除
2. **P1**: 从企业库安装 + 还原
3. **P2**: Diff对比 + 升级提示 + 分发统计
