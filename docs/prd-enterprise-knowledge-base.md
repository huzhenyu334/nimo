# PRD：BitFantasy 企业知识库

> 版本：v2.0 | 日期：2026-02-23 | 作者：Lyra | 状态：待评审
> v2.0 变更：整合竞品优势（Dify/Notion/飞书知识库/Wiki.js），知识写入提升至第一期

## 0. 竞品分析与取长补短

### 0.1 各家核心优势

| 产品 | 核心优势 | 我们要学的 |
|------|---------|-----------|
| **Dify** | ① 多格式自动解析（PDF/Word/Excel/HTML） ② Parent-child chunk模式（子chunk精确匹配→返回父chunk上下文） ③ 三种检索策略（向量/全文/混合+Rerank） ④ 检索测试面板（可视化调参+召回日志） ⑤ Q&A自动生成模式 | chunk分层、检索测试、混合检索 |
| **Notion** | ① Block模型（万物皆Block，灵活组合） ② 双向链接（页面互引+反向引用列表） ③ 数据库视图（同一数据，表格/看板/日历/画廊多视图） ④ 模板系统（一键创建标准化文档） ⑤ 团队Wiki空间（导航树+面包屑） | 文档模板、知识域导航树、关联引用 |
| **飞书知识库** | ① 权限精细到节点（Space→节点→文档三级权限） ② 与飞书生态深度绑定（消息/审批/日历联动） ③ 多人协作+评论 ④ 知识空间分组管理 | 节点级权限、飞书联动、空间分组 |
| **Wiki.js** | ① Git同步（内容即代码，版本管理天然） ② Markdown原生支持 ③ 多种搜索引擎可插拔（Elasticsearch/PostgreSQL/Algolia） ④ 资产管理（图片/附件集中管理） ⑤ 多语言支持 | Git式版本管理、资产管理、搜索引擎可插拔 |

### 0.2 我们的独特优势（别人没有的）

| 优势 | 说明 |
|------|------|
| **Agent原生** | md文件直接进prompt，零转换成本。Notion/飞书需要API→JSON→提取→拼接 |
| **Agent可写** | Agent工作流中自动产生知识。传统知识库靠人手动维护 |
| **跨Agent共享** | 4个Agent共用一个知识库，知识自动同步。Dify的knowledge只服务单个app |
| **轻量** | 纯Go+SQLite+QMD+文件系统，7.5G服务器能跑。Dify要Redis+PostgreSQL+Weaviate |
| **工作流驱动** | 知识产生集成在ACP工作流里，不是独立系统 |

### 0.3 整合策略：取长补短

**从Dify学：**
- ✅ **Section-level chunk**：按md的`##`标题自动切分chunk，检索到section级别而非文件级别
- ✅ **混合检索+Rerank**：QMD已支持BM25+向量+rerank，直接利用
- ✅ **检索测试面板**：在ACP UI中提供检索调试界面，可视化看命中了什么
- ✅ **Q&A模式（简化版）**：Agent写入知识时自动生成3-5个"典型问题"作为检索锚点
- ⏳ 多格式解析：v2再做，先聚焦md

**从Notion学：**
- ✅ **文档模板**：每个知识域预置模板（供应商评估模板、技术方案模板、评审结论模板）
- ✅ **知识域导航树**：树形sidebar浏览知识结构，不只是平铺列表
- ✅ **文档关联**：front-matter中`related`字段，详情页展示关联文档
- ⏳ 双向链接：v2再做
- ❌ Block模型：不需要，md足够

**从飞书知识库学：**
- ✅ **空间分组**：知识域(domain)作为一级分组，类似飞书的Knowledge Space
- ✅ **飞书通知联动**：知识写入审批 → 飞书通知；重要知识更新 → 飞书推送
- ⏳ 节点级权限：v2，当前全Agent可读
- ❌ 多人协作编辑：不需要，Agent和人通过不同入口写入

**从Wiki.js学：**
- ✅ **Git式版本管理**：每次修改自动保存版本，diff对比，回滚
- ✅ **资产管理**：知识文档关联的图片/附件集中存储在assets/目录
- ✅ **搜索引擎可插拔**：QMD作为默认，未来可换Elasticsearch
- ✅ **Markdown原生**：这本来就是我们的路线

## 1. 背景与目标

### 1.1 背景

BitFantasy是一家Agent-first的智能眼镜公司。20个人类员工 + 不断扩展的Agent团队（当前4个，未来更多）。业务扩展主要靠增加Agent而非人力。

**现状问题：**
- Agent workspace相互隔离，知识无法跨Agent共享
- 企业知识散落在飞书文档、Agent各自的bank/目录、人脑中
- 业务系统（PLM/ERP）存结构化数据，但"为什么这样做"的决策上下文没有归宿
- Agent每次session醒来是空白的，没写下来的知识等于不存在

### 1.2 目标

构建ACP内的**企业知识库模块**，作为所有Agent的**共享大脑**：
- Agent能通过tool主动检索企业知识（**knowledge_search**）
- Agent能将工作中产生的认知写入知识库（**knowledge_write**）— **第一期核心功能**
- 人类能通过ACP界面管理和审阅知识
- 知识有版本控制、分域管理、模板规范

### 1.3 非目标

- ❌ 不替代PLM/ERP等业务系统（它们管结构化数据）
- ❌ 不替代飞书文档（人类协作仍在飞书）
- ❌ 不做通用RAG平台（不处理任意格式文件上传，聚焦md）
- ❌ 不做全文搜索引擎（有QMD就够）

## 2. 知识库放什么

### 2.1 内容定位

> **业务系统记录"是什么"，知识库记录"为什么"和"怎么做"。**

| 属于知识库 | 不属于知识库（属于业务系统） |
|-----------|------------------------|
| 为什么选舜宇做摄像头供应商 | 舜宇的报价单（ERP） |
| EVT评审的核心结论和待办 | BOM物料清单（PLM） |
| 良率问题的根因分析 | 工单详情（CRM） |
| 技术方案选型对比 | 代码仓库（Git） |
| 新员工/新Agent入职指南 | 考勤记录（HR系统） |
| 产品开发流程规范 | 订单数据（ERP） |
| 跨部门协作经验教训 | 审批记录（飞书） |

### 2.2 知识分域（知识空间）

借鉴飞书Knowledge Space概念，按业务域组织：

```
enterprise-knowledge/
├── product/              ← 产品相关
│   ├── specs/            ← 产品规格与技术决策
│   ├── design/           ← 设计规范与指南
│   └── roadmap/          ← 产品路线图与里程碑
├── engineering/          ← 工程研发
│   ├── hardware/         ← 硬件设计经验
│   ├── firmware/         ← 固件开发知识
│   ├── testing/          ← 测试标准与方法
│   └── manufacturing/    ← 制造工艺知识（EVT/DVT/PVT/MP）
├── supply-chain/         ← 供应链
│   ├── suppliers/        ← 供应商评估与档案
│   ├── sourcing/         ← 采购策略与经验
│   └── logistics/        ← 物流与交付
├── quality/              ← 质量管理
│   ├── standards/        ← 质量标准与规范
│   ├── issues/           ← 问题库与根因分析
│   └── reviews/          ← 评审记录与结论
├── ops/                  ← 运营管理
│   ├── process/          ← 流程规范
│   ├── onboarding/       ← 入职指南（人+Agent）
│   └── infra/            ← 基础设施知识
└── market/               ← 市场与客户
    ├── competitors/      ← 竞品分析
    ├── customers/        ← 客户洞察
    └── trends/           ← 行业趋势
```

### 2.3 知识文档格式

统一使用**结构化Markdown + front-matter**：

```markdown
---
title: 摄像头模组供应商选型
domain: supply-chain/sourcing
tags: [摄像头, 供应商, EVT]
related: [quality/issues/camera-yield.md, product/specs/nm2000-camera.md]
template: supplier-evaluation     # 使用的文档模板
created_by: supply-chain-agent
created_at: 2026-03-15
updated_by: lyra
updated_at: 2026-03-20
status: active                    # active | archived | draft
confidential: false
qa_anchors:                       # 自动生成的检索锚点
  - 为什么选舜宇做摄像头供应商？
  - 摄像头模组各供应商报价对比？
  - 舜宇的良率和交期怎么样？
---

# 摄像头模组供应商选型

## 结论
选择舜宇光学作为NM2000摄像头模组供应商。

## 评估对比
| 维度 | 舜宇 | 三星 | 丘钛 |
|------|------|------|------|
| 单价 | ¥45.2 | ¥62.8 | ¥41.5 |
| 良率 | 92% | 98% | 87% |
| 交期 | 8周 | 12周 | 6周 |

## 决策理由
1. 综合成本最优（良率×单价）
2. 交期可接受，三星太慢影响EVT节点
3. 丘钛良率风险太高

## 风险与应对
- 舜宇良率92%偏低 → DVT阶段要求提升至95%
- 备选方案：丘钛作为第二供应商培养
```

### 2.4 文档模板系统（学自Notion）

每个知识域预置文档模板，Agent和人类创建文档时可选用：

| 模板 | 适用场景 | 核心字段 |
|------|---------|---------|
| **supplier-evaluation** | 供应商评估 | 结论、对比表、决策理由、风险 |
| **tech-decision** | 技术方案选型 | 背景、方案对比、选型理由、trade-off |
| **review-summary** | 评审结论 | 评审日期、参与人、结论、待办项 |
| **issue-rca** | 问题根因分析 | 问题描述、影响范围、根因、解决方案、预防措施 |
| **process-guide** | 流程规范 | 适用范围、流程步骤、注意事项、模板 |
| **onboarding** | 入职指南 | 角色、第一天/第一周/第一月任务、常用资源 |
| **competitor-analysis** | 竞品分析 | 基本信息、产品对比、优劣势、启示 |

模板定义存储在 `enterprise-knowledge/_templates/` 目录，每个模板是一个md文件。

## 3. 系统架构

### 3.1 整体架构

```
┌───────────────────────────────────────────────────────────────┐
│                      ACP Web 界面                              │
│  知识空间导航树 / 文档CRUD / 搜索 / 检索测试 / 模板管理        │
└──────────────────────────┬────────────────────────────────────┘
                           │
┌──────────────────────────▼────────────────────────────────────┐
│                    ACP 后端 (Go)                                │
│  知识CRUD / 版本管理(Git式) / 模板 / 搜索API / 审批            │
├───────────────────────────────────────────────────────────────┤
│  存储层                                                         │
│  ┌──────────┐  ┌──────────────┐  ┌──────────────────────────┐ │
│  │ SQLite   │  │ 文件系统     │  │ QMD 搜索索引              │ │
│  │ 元数据   │  │ md文件+资产  │  │ Section-level chunk       │ │
│  │ 版本历史 │  │              │  │ BM25+向量+Rerank          │ │
│  │ 模板定义 │  │              │  │ Q&A锚点索引               │ │
│  └──────────┘  └──────────────┘  └──────────────────────────┘ │
└──────────────────────────┬────────────────────────────────────┘
                           │
           ┌───────────────┼───────────────┐
           │               │               │
      ┌────▼────┐    ┌────▼────┐    ┌────▼────┐
      │  Lyra   │    │  Alice  │    │   PM    │
      │k_search │    │k_search │    │k_search │
      │k_write  │    │k_write  │    │k_write  │
      └─────────┘    └─────────┘    └─────────┘
```

### 3.2 数据模型

```sql
-- 知识空间/域
knowledge_domains:
  id              TEXT PRIMARY KEY
  name            TEXT UNIQUE NOT NULL  -- 'product', 'supply-chain'
  display_name    TEXT NOT NULL         -- '产品', '供应链'
  description     TEXT
  icon            TEXT                  -- 前端图标
  parent_id       TEXT                  -- 支持子域层级
  doc_count       INT DEFAULT 0
  created_at      TIMESTAMP

-- 知识文档
knowledge_documents:
  id              TEXT PRIMARY KEY
  title           TEXT NOT NULL
  domain_id       TEXT NOT NULL         -- → knowledge_domains.id
  path            TEXT UNIQUE NOT NULL  -- 'supply-chain/sourcing/camera-vendor.md'
  content         TEXT NOT NULL         -- 完整md内容（含front-matter）
  summary         TEXT                  -- AI生成的摘要
  tags            TEXT                  -- JSON数组 '["摄像头","供应商"]'
  related_docs    TEXT                  -- JSON数组 关联文档path
  template_id     TEXT                  -- 使用的文档模板
  qa_anchors      TEXT                  -- JSON数组 自动生成的Q&A锚点
  status          TEXT DEFAULT 'active' -- active | archived | draft
  confidential    BOOLEAN DEFAULT 0
  created_by      TEXT NOT NULL         -- agent_id 或 'human'
  updated_by      TEXT
  current_version INT DEFAULT 1
  created_at      TIMESTAMP
  updated_at      TIMESTAMP

-- 知识版本历史（Git式）
knowledge_versions:
  id              TEXT PRIMARY KEY
  document_id     TEXT NOT NULL         -- → knowledge_documents.id
  version         INT NOT NULL
  content         TEXT NOT NULL         -- 该版本完整内容
  change_summary  TEXT                  -- 变更说明
  changed_by      TEXT NOT NULL
  diff_from_prev  TEXT                  -- 与上一版本的diff（增量存储，减少空间）
  created_at      TIMESTAMP

-- 文档模板
knowledge_templates:
  id              TEXT PRIMARY KEY
  name            TEXT UNIQUE NOT NULL  -- 'supplier-evaluation'
  display_name    TEXT NOT NULL         -- '供应商评估'
  domain_id       TEXT                  -- 推荐使用域
  content         TEXT NOT NULL         -- 模板md内容
  description     TEXT
  created_at      TIMESTAMP
  updated_at      TIMESTAMP

-- Agent写入审批队列
knowledge_approvals:
  id              TEXT PRIMARY KEY
  document_id     TEXT                  -- 更新时关联已有文档，新建时为空
  action          TEXT NOT NULL         -- 'create' | 'update'
  title           TEXT NOT NULL
  domain_id       TEXT NOT NULL
  content         TEXT NOT NULL         -- 待审批的内容
  submitted_by    TEXT NOT NULL         -- 提交的agent_id
  status          TEXT DEFAULT 'pending' -- pending | approved | rejected
  review_note     TEXT                  -- 审批意见
  reviewed_by     TEXT                  -- 审批人
  reviewed_at     TIMESTAMP
  created_at      TIMESTAMP

-- 检索日志（用于检索测试面板和优化）
knowledge_search_logs:
  id              TEXT PRIMARY KEY
  query           TEXT NOT NULL
  domain_filter   TEXT                  -- 限定域
  results_count   INT
  top_result_id   TEXT
  top_score       REAL
  agent_id        TEXT                  -- 哪个agent发起的
  created_at      TIMESTAMP
```

### 3.3 存储与索引

**双写机制：** 知识文档同时存在：
- **SQLite**：元数据 + 内容 + 版本历史（用于API和版本管理）
- **文件系统**：`/home/claw/.openclaw/shared-knowledge/` 下的md文件（用于QMD索引）

**Section-level chunk索引（学自Dify）：**
QMD索引时不按整个文件，而是按md的`##`标题切分为section。检索结果返回section级别片段，而非整个文件。

```
文件: supply-chain/sourcing/camera-vendor.md
  → chunk1: "## 结论\n选择舜宇光学..." (section)
  → chunk2: "## 评估对比\n| 维度 | 舜宇 |..." (section)
  → chunk3: "## 决策理由\n1. 综合成本..." (section)
  → chunk4: "## 风险与应对\n..." (section)
```

**Q&A锚点索引（学自Dify Q&A模式）：**
每篇文档的`qa_anchors`字段存储3-5个典型问题。这些问题也被QMD索引，用于question→question匹配（比question→paragraph更精准）。

Agent写入知识时自动生成qa_anchors（调LLM），人工编辑时可手动调整。

### 3.4 资产管理（学自Wiki.js）

知识文档关联的非文本资源（图片、图表、附件）存储在统一的assets目录：

```
enterprise-knowledge/
├── _assets/                  ← 集中管理
│   ├── images/               ← 图片
│   └── attachments/          ← 附件（PDF/Excel等原始文件）
├── _templates/               ← 文档模板
└── product/...               ← 知识文档
```

md文档中通过相对路径引用：`![对比图](/_assets/images/camera-comparison.png)`

**v1.0最小实现**：资产管理只做文件上传和引用，不做预览和索引。

## 4. Agent Tool 设计

### 4.1 knowledge_search tool

```json
{
  "name": "knowledge_search",
  "description": "搜索企业知识库。查找公司级知识（产品规格、供应商、流程规范、历史决策等）。返回section级别的精确片段。",
  "parameters": {
    "query": "搜索关键词或问题",
    "domain": "可选，限定搜索域：product/engineering/supply-chain/quality/ops/market",
    "tags": "可选，按标签过滤",
    "limit": "返回结果数量，默认5"
  }
}
```

**返回格式：**
```json
{
  "results": [
    {
      "title": "摄像头模组供应商选型",
      "section": "## 评估对比",
      "domain": "supply-chain/sourcing",
      "path": "supply-chain/sourcing/camera-vendor.md",
      "snippet": "| 维度 | 舜宇 | 三星 | 丘钛 |\n|------|------|------|------|\n| 单价 | ¥45.2 | ¥62.8 | ¥41.5 |...",
      "score": 0.87,
      "updated_at": "2026-03-20",
      "updated_by": "lyra"
    }
  ],
  "total": 1
}
```

### 4.2 knowledge_write tool ⭐ 第一期核心

Agent工作流中产生认知时主动写入：

```json
{
  "name": "knowledge_write",
  "description": "将知识写入企业知识库。当工作中产生有价值的结论、经验、分析时使用。新建文档需审批，更新已有文档直接生效。",
  "parameters": {
    "action": "create | update",
    "document_id": "更新时必填，已有文档ID",
    "title": "新建时必填，知识标题",
    "domain": "新建时必填，知识域路径",
    "content": "Markdown内容（无需front-matter，系统自动生成）",
    "tags": ["标签数组"],
    "template": "可选，使用的文档模板名",
    "change_summary": "变更说明（更新时必填）"
  }
}
```

**写入流程：**

```
Agent调用 knowledge_write
    │
    ├─ action=create（新建）
    │   └→ 进入审批队列 → 人工在ACP审批 → 通过后入库 + 生成qa_anchors + QMD索引
    │
    └─ action=update（更新已有文档）
        └→ 直接生效 → 创建新版本 → 更新qa_anchors → 重建QMD索引
```

**为什么"更新"不需要审批：** 已审批入库的文档代表已被认可的知识。Agent更新它（补充数据、修正信息）是正常迭代，有版本历史可以回滚。如果每次更新都审批，会严重阻塞Agent工作流。

**写入后自动处理：**
1. 调LLM生成`summary`（<100字摘要）
2. 调LLM生成`qa_anchors`（3-5个典型问题）
3. 文件写入`shared-knowledge/`目录
4. 触发QMD重建该文件的section索引

### 4.3 knowledge_list tool

查看知识库目录结构，用于Agent了解有哪些知识可用：

```json
{
  "name": "knowledge_list",
  "description": "浏览企业知识库目录。了解某个域下有哪些知识文档。",
  "parameters": {
    "domain": "可选，限定浏览的域",
    "status": "可选，默认active"
  }
}
```

## 5. ACP界面设计

### 5.0 设计原则

> **这是给人用的界面，不是给工程师看的后台。要精美、要流畅、要让人愿意用。**

**视觉基调：**
- 参考 Notion 的简洁感 + Linear 的精致感
- 大量留白，信息密度适中，不拥挤
- 字体层级清晰：标题/正文/辅助信息有明确区分
- 配色方案：以品牌色为点缀，大面积使用中性色（白/灰），知识域用彩色图标区分
- 卡片式布局为主，圆角、微阴影、hover动效

**交互体验：**
- **即时反馈**：所有操作（保存/搜索/审批）有loading状态和成功提示
- **键盘友好**：`⌘+K` 全局搜索、`⌘+N` 新建文档、`Esc` 关闭弹窗
- **过渡动画**：页面切换平滑过渡，侧边栏折叠有动画，搜索结果渐入
- **Markdown编辑体验**：实时预览（左编辑右预览），或所见即所得模式（推荐用开源编辑器如 Milkdown/ByteMD）
- **空状态设计**：每个页面的空状态都有引导（图标+文案+操作按钮），不只是白屏
- **响应式**：适配1280px~1920px屏幕

**微交互细节：**
- 文档卡片hover时微上浮 + 阴影加深
- 标签可点击筛选，点击后有chip样式反馈
- 版本历史的timeline用时间轴样式，不是干巴巴的表格
- 审批通过/驳回用不同颜色动画确认
- 搜索输入时实时高亮关键词，debounce 300ms
- 树形导航的展开/折叠有smooth动画
- 面包屑导航：知识库 > 供应链 > 采购策略 > 文档名

**组件库：** 复用ACP现有的Ant Design（antd）体系，在此基础上定制主题token：
- 圆角：`borderRadius: 8px`（卡片/按钮），`12px`（大容器）
- 阴影：`box-shadow: 0 1px 3px rgba(0,0,0,0.06), 0 1px 2px rgba(0,0,0,0.04)`
- 间距：遵循8px网格系统
- 动画：`transition: all 0.2s ease`

### 5.1 侧边栏

```
Agent管理
  ├── Agent列表
  ├── 企业Skills库
  └── 📚 企业知识库     ← 新增
       ├── 知识空间      ← 按域浏览
       ├── 搜索          ← 全局搜索+检索测试
       ├── 待审批        ← Agent写入审批队列
       └── 模板管理      ← 文档模板CRUD
```

### 5.2 知识空间页（主页面）— 树形导航

借鉴Notion/飞书的树形导航，左侧域树 + 右侧文档列表。

**布局：** 左侧导航240px宽，可拖拽调整，右侧自适应。

**左侧域树：**
- 每个域有彩色圆点图标（产品=蓝、工程=紫、供应链=绿、质量=橙、运营=灰、市场=红）
- 展开/折叠smooth动画（200ms ease）
- 当前选中域高亮背景
- 域名后显示文档数badge
- 底部固定"待审批"和"模板管理"入口

**右侧文档列表：**
- 顶部面包屑导航：📚 知识库 / 🚛 供应链 / 采购策略
- 文档卡片式列表（非表格），每张卡片包含：标题（加粗）、摘要（灰色小字，1行截断）、标签chips、更新时间+作者头像
- 卡片hover：微上浮2px + 阴影加深 + 右侧出现→箭头
- 右上角"新建文档"主按钮（Primary蓝色）+ "搜索"按钮
- 列表支持排序：最近更新/创建时间/标题
- 空状态：大图标 + "这个空间还没有知识文档" + "新建第一篇"按钮

```
┌──────────────────────────────────────────────────────────────────┐
│  📚 企业知识库                        ⌘K 搜索  [+ 新建文档]      │
├──────────────┬───────────────────────────────────────────────────┤
│              │                                                    │
│  知识空间     │  📚 知识库 / 🚛 供应链 / 采购策略    排序: 最近更新▼│
│              │                                                    │
│  🔵 产品 (12)│  ┌──────────────────────────────────────────────┐ │
│    规格      │  │ 📄 摄像头模组供应商选型                     → │ │
│    设计      │  │ 选择舜宇光学作为NM2000摄像头模组供应商...      │ │
│    路线图    │  │ 🏷摄像头  🏷供应商   lyra · 3月20日           │ │
│  🟣 工程 (8) │  └──────────────────────────────────────────────┘ │
│    硬件      │                                                    │
│    固件      │  ┌──────────────────────────────────────────────┐ │
│    测试      │  │ 📄 显示模组BOM成本分析                      → │ │
│    制造      │  │ NM2000显示模组BOM对比三个方案的综合成本...     │ │
│  🟢 供应链(5)│  │ 🏷显示  🏷BOM       alice · 3月18日          │ │
│  ► 采购策略  │  └──────────────────────────────────────────────┘ │
│    供应商    │                                                    │
│    物流      │  ┌──────────────────────────────────────────────┐ │
│  🟠 质量 (6) │  │ 📄 舜宇质量改进跟踪                        → │ │
│  ⚪ 运营 (4) │  │ 追踪舜宇摄像头模组良率从92%提升到95%的进展.. │ │
│  🔴 市场 (3) │  │ 🏷舜宇  🏷质量      pm · 3月15日            │ │
│              │  └──────────────────────────────────────────────┘ │
│  ──────────  │                                                    │
│  📋 待审批(3)│                                                    │
│  📐 模板管理  │                                                    │
└──────────────┴───────────────────────────────────────────────────┘
```

### 5.3 文档详情页

**布局：** 主体内容居中（max-width 800px），两侧留白。右侧浮动侧栏显示元信息。

**顶部区域：**
- 面包屑导航（可点击返回上级）
- 文档标题（28px加粗）
- 元信息行：版本badge(v3)、作者头像+名字、更新时间、模板badge
- 标签：彩色chip样式，可点击跳转到标签筛选
- 操作按钮组：[编辑]主按钮 + [⋯]更多（归档/导出/删除）

**文档正文：**
- Markdown渲染区域，排版参考 GitHub/Notion 的渲染质感
- 代码块有语法高亮
- 表格有斑马纹 + 表头固定
- 标题层级有左侧竖线装饰（类似 Notion 的heading样式）
- 正文右侧可选显示"文档大纲"（TOC），按##标题自动生成

**底部区域三个折叠Panel（Collapse），默认展开第一个：**

1. **📎 关联文档** — 卡片式横排，每张小卡片显示标题+域+摘要首行，点击跳转
2. **🎯 检索锚点** — 展示Agent会通过什么问题找到这篇文档，可手动编辑
3. **📜 版本历史** — 时间轴（Timeline）样式，左侧时间线圆点，右侧版本信息卡片。每个版本卡片有：版本号、时间、作者头像、变更说明、[查看]和[diff]按钮。最新版本圆点高亮为主色。

```
┌──────────────────────────────────────────────────────────────────┐
│  📚 知识库 / 🚛 供应链 / 采购策略                                  │
│                                                                    │
│  摄像头模组供应商选型                                               │
│  ┌v3┐  lyra · 3月20日更新  📐 supplier-evaluation                 │
│  🏷摄像头  🏷供应商  🏷EVT                  [✏️ 编辑]  [⋯]        │
│                                                                    │
│  ┌─── 文档内容 ─────────────────────────────── 大纲 ▶ ──────┐   │
│  │                                                             │   │
│  │  ▎ 结论                                                    │   │
│  │  选择舜宇光学作为NM2000摄像头模组供应商。                     │   │
│  │                                                             │   │
│  │  ▎ 评估对比                                                │   │
│  │  ┌──────┬──────┬──────┬──────┐                            │   │
│  │  │ 维度 │ 舜宇 │ 三星 │ 丘钛 │  ← 斑马纹表格             │   │
│  │  ├──────┼──────┼──────┼──────┤                            │   │
│  │  │ 单价 │¥45.2 │¥62.8 │¥41.5 │                            │   │
│  │  └──────┴──────┴──────┴──────┘                            │   │
│  │  ...                                                        │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                    │
│  ▼ 📎 关联文档 (2)                                                │
│  ┌────────────────────┐  ┌────────────────────┐                   │
│  │ 📄 摄像头良率问题   │  │ 📄 NM2000摄像头规格│                   │
│  │ quality/issues     │  │ product/specs      │                   │
│  └────────────────────┘  └────────────────────┘                   │
│                                                                    │
│  ▼ 🎯 检索锚点 (3)                                    [编辑]     │
│  • 为什么选舜宇做摄像头供应商？                                     │
│  • 摄像头模组各供应商报价对比？                                     │
│  • 舜宇的良率和交期怎么样？                                         │
│                                                                    │
│  ▼ 📜 版本历史                                                    │
│     ●─── v3  3月20日  lyra   "更新舜宇最新报价"  [diff]           │
│     │                                                              │
│     ○─── v2  3月16日  alice  "补充丘钛评估数据"  [diff]           │
│     │                                                              │
│     ○─── v1  3月15日  pm     "首次创建"                           │
└──────────────────────────────────────────────────────────────────┘
```

**编辑模式：**
- 点击"编辑"进入编辑态，推荐使用 **ByteMD**（字节跳动开源Markdown编辑器，与飞书同源）或 **Milkdown**
- 编辑器特性：工具栏（加粗/标题/列表/表格/链接）、实时预览、图片拖拽上传、快捷键
- 保存时弹出小弹窗填写"变更说明"（必填），自动创建新版本
- 自动保存草稿（localStorage），防丢失

### 5.4 全局搜索 + 检索测试面板（学自Dify）

**搜索入口：** `⌘+K` 快捷键唤出全局搜索弹窗（类似 Spotlight/Raycast），覆盖全屏半透明遮罩。

**搜索弹窗：**
- 顶部大搜索框，自动focus，输入即搜索（debounce 300ms）
- 搜索框下方筛选pills：域过滤 + 标签过滤（可多选）
- 结果列表：每条结果是一张mini卡片

**结果卡片设计：**
- 左侧域彩色圆点 + 文档标题（加粗）+ section名（灰色）
- 中间：匹配片段，关键词高亮（黄色背景）
- 右侧：相关度分数（进度条样式，不是数字）
- 底部小字：命中方式tag（"Q&A锚点" / "内容匹配" / "标签匹配"）
- 点击结果直接跳转文档详情页，定位到匹配section

**检索测试面板（可折叠，默认收起）：**
面向技术人员调试检索效果，展开后显示BM25/向量/Rerank各阶段得分。

```
┌──────────────────────────────────────────────────────────────────┐
│                                                                    │
│  🔍  摄像头良率问题                                        ✕     │
│  ─────────────────────────────────────────────────────────────   │
│  筛选: [全部域 ▼]  [全部标签 ▼]              3条结果 · 45ms      │
│                                                                    │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │ 🟢 摄像头模组供应商选型  › §风险与应对          ████████░ │  │
│  │ "...舜宇良率92%偏低 → DVT阶段要求提升至95%..."             │  │
│  │ 🎯Q&A锚点  📝内容匹配                                     │  │
│  └────────────────────────────────────────────────────────────┘  │
│                                                                    │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │ 🟠 EVT-1评审结论  › §待办项                    ███████░░ │  │
│  │ "...摄像头模组良率不达标，需二次打样..."                    │  │
│  │ 📝内容匹配                                                 │  │
│  └────────────────────────────────────────────────────────────┘  │
│                                                                    │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │ 🟠 NM2000硬件问题追踪  › §摄像头               █████░░░░ │  │
│  │ "...良率改进方案已提交舜宇..."                              │  │
│  │ 🏷标签匹配  📝内容匹配                                     │  │
│  └────────────────────────────────────────────────────────────┘  │
│                                                                    │
│  ▶ 检索调试详情                                                   │
│                                                                    │
└──────────────────────────────────────────────────────────────────┘
```

### 5.5 审批队列页

**设计：** 类似邮件收件箱的左右分栏布局。左侧待审批列表，右侧内容预览+审批操作。

**左侧列表：**
- Tab切换：待审批 / 已通过 / 已驳回
- 每条显示：类型icon(🆕新建)、标题、提交Agent头像+名字、域badge、相对时间
- 未读状态：左侧有蓝色竖条标记

**右侧详情：**
- 顶部：标题 + 提交者 + 域 + 模板信息
- 中间：完整Markdown渲染预览（与文档详情页同等质量）
- 底部固定栏：审批意见输入框 + 驳回(红outline按钮) + 通过(绿实心按钮)
- 通过时有confetti微动画🎉，驳回时按钮shake反馈

```
┌──────────────────────────────────────────────────────────────────┐
│  📋 知识审批                                                       │
│  [待审批(3)]  [已通过]  [已驳回]                                   │
├────────────────────┬─────────────────────────────────────────────┤
│                    │                                               │
│ 🔵 铰链供应商初步.. │  📄 铰链供应商初步评估                        │
│    alice · 10分钟前 │  提交者: alice  域: 🟢供应链/sourcing          │
│                    │  模板: supplier-evaluation                    │
│ ░ DVT-1光学测试..  │                                               │
│    pm · 1小时前     │  ┌── 内容预览 ───────────────────────────┐  │
│                    │  │                                         │  │
│ ░ 竞品XReal Air.. │  │  ▎ 结论                                │  │
│    lyra · 2小时前   │  │  初步筛选出3家铰链供应商进入详细        │  │
│                    │  │  评估阶段：xxx、yyy、zzz               │  │
│                    │  │                                         │  │
│                    │  │  ▎ 评估维度                            │  │
│                    │  │  价格、交期、产能、配合度               │  │
│                    │  │  ...                                    │  │
│                    │  └─────────────────────────────────────────┘  │
│                    │                                               │
│                    │  审批意见 (可选):                              │
│                    │  [补充一下产能数据再入库_______________]       │
│                    │                                               │
│                    │              [❌ 驳回]    [✅ 通过并入库]     │
└────────────────────┴─────────────────────────────────────────────┘
```

### 5.6 模板管理页

**设计：** 卡片网格布局（3列），每张模板卡片是一个精致的预览卡。

**模板卡片：**
- 顶部：模板图标（大emoji）+ 模板名称
- 中间：描述文字（2行截断）+ 推荐域badge
- 底部：使用次数 + 最近更新时间
- hover：边框变主色 + 出现"预览"和"编辑"按钮

**新建模板：** 第一张卡片是虚线边框的"+ 新建模板"占位卡，点击弹出Markdown编辑器。

```
┌──────────────────────────────────────────────────────────────────┐
│  📐 文档模板                                                       │
│  Agent和人类创建知识文档时可选用模板，确保格式规范                   │
│                                                                    │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐           │
│  │  ╭─ ─ ─ ─╮  │  │  📊          │  │  ⚙️          │           │
│  │  │ + 新建 │  │  │  供应商评估   │  │  技术方案选型 │           │
│  │  │  模板  │  │  │  评估和选择.. │  │  对比方案优.. │           │
│  │  ╰─ ─ ─ ─╯  │  │  🟢供应链     │  │  🟣工程       │           │
│  │              │  │  12次 · 3月18 │  │  8次 · 3月15  │           │
│  └──────────────┘  └──────────────┘  └──────────────┘           │
│                                                                    │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐           │
│  │  📝          │  │  🔍          │  │  📋          │           │
│  │  评审结论     │  │  问题根因分析 │  │  流程规范     │           │
│  │  记录评审会.. │  │  分析问题根.. │  │  制定标准化.. │           │
│  │  🟠质量       │  │  🟠质量       │  │  ⚪运营       │           │
│  │  15次 · 3月20│  │  6次 · 3月12  │  │  4次 · 2月28  │           │
│  └──────────────┘  └──────────────┘  └──────────────┘           │
└──────────────────────────────────────────────────────────────────┘
```

### 5.7 新建文档页

**设计：** 两步流程 — ① 选模板 → ② 编辑内容。

**步骤一：选择模板（弹窗）**
- 模板卡片网格，点击选择
- 也可选择"空白文档"跳过模板
- 选择后自动填充模板骨架到编辑器

**步骤二：编辑器页面**
- 顶部：域选择下拉 + 标签输入（tag input组件，回车添加）
- 中间：Markdown编辑器（ByteMD/Milkdown），预填模板内容，标题自动focus
- 底部固定栏：[取消] + [保存草稿] + [发布]
- 编辑过程中自动保存草稿到localStorage

### 5.8 统计概览（知识库首页顶部）

在知识空间页顶部展示4个统计卡片：

```
┌────────────┐  ┌────────────┐  ┌────────────┐  ┌────────────┐
│  📄 42     │  │  🏷 15     │  │  🔍 128    │  │  ✍️ 6      │
│  知识文档   │  │  标签总数   │  │  本月检索   │  │  本月新增   │
│  +3 本周   │  │            │  │  ↑23%      │  │  待审批: 3  │
└────────────┘  └────────────┘  └────────────┘  └────────────┘
```

## 6. API设计

### 6.1 知识文档 CRUD

```
GET    /api/knowledge                       → 列表（domain/tags/status筛选+分页）
POST   /api/knowledge                       → 创建文档（人工创建，直接入库）
GET    /api/knowledge/:id                   → 文档详情（含关联文档+qa_anchors）
PUT    /api/knowledge/:id                   → 更新文档（自动创建新版本+重建索引）
DELETE /api/knowledge/:id                   → 归档文档（不物理删除）
GET    /api/knowledge/:id/versions          → 版本历史列表
GET    /api/knowledge/:id/versions/:v       → 查看历史版本
POST   /api/knowledge/:id/versions/:v/rollback → 回滚到指定版本
GET    /api/knowledge/:id/diff/:v1/:v2      → 两个版本的diff
```

### 6.2 搜索

```
GET    /api/knowledge/search?q=xxx&domain=xxx&tags=xxx&limit=5
  → QMD section-level搜索 + Q&A锚点匹配
  → 返回: results + debug info (scores, match types)
```

### 6.3 知识域管理

```
GET    /api/knowledge/domains               → 树形域列表（含文档计数）
POST   /api/knowledge/domains               → 创建域
PUT    /api/knowledge/domains/:id           → 更新域
DELETE /api/knowledge/domains/:id           → 删除域（需为空）
```

### 6.4 Agent Tool接口

```
POST   /api/knowledge/tool/search           → knowledge_search tool
POST   /api/knowledge/tool/write            → knowledge_write tool（新建进审批，更新直接生效）
POST   /api/knowledge/tool/list             → knowledge_list tool
```

### 6.5 审批

```
GET    /api/knowledge/approvals             → 审批队列（pending/approved/rejected）
POST   /api/knowledge/approvals/:id/approve → 通过审批
POST   /api/knowledge/approvals/:id/reject  → 驳回审批
```

### 6.6 模板

```
GET    /api/knowledge/templates             → 模板列表
POST   /api/knowledge/templates             → 创建模板
GET    /api/knowledge/templates/:id         → 模板详情
PUT    /api/knowledge/templates/:id         → 更新模板
DELETE /api/knowledge/templates/:id         → 删除模板
```

## 7. OpenClaw集成

### 7.1 注册knowledge tools

ACP插件初始化时向OpenClaw gateway注册3个tools：
- `knowledge_search` — 所有Agent可用
- `knowledge_write` — 所有Agent可用
- `knowledge_list` — 所有Agent可用

### 7.2 QMD索引配置

知识库文件目录 `/home/claw/.openclaw/shared-knowledge/` 作为QMD独立collection：
- 与memory_search的索引分离（不混合个人记忆和企业知识）
- Section-level chunking（按`##`标题切分）
- Q&A锚点作为额外索引条目

### 7.3 工作流集成

ACP工作流中可配置knowledge_write步骤，让Agent在工作流执行过程中自动将结论写入知识库：

```yaml
steps:
  - name: research
    agent: alice
    prompt: "调研NM2000显示方案..."
  - name: write_knowledge
    agent: alice
    prompt: "将调研结论写入企业知识库，域为engineering/hardware，使用tech-decision模板"
    tools: [knowledge_write]
```

### 7.4 AGENTS.md使用指引

在shared-bootstrap的AGENTS.md中增加：
```markdown
## 企业知识库
- 需要公司级知识（产品规格、供应商、流程等）→ knowledge_search
- 需要个人记忆（之前做了什么、老板说了什么）→ memory_search
- 工作中产生有价值的结论 → knowledge_write 写入企业知识库
- 写入前确认：这个知识对其他Agent/人类也有用吗？如果只对自己有用，写bank/就好
```

## 8. 权限与安全

### 8.1 v1.0 简单模型

| 操作 | 权限 |
|------|------|
| 搜索/读取 | 所有Agent + 人工（通过ACP界面） |
| 新建文档（Agent） | 需人工审批 |
| 新建文档（人工） | ACP界面直接创建，无需审批 |
| 更新文档 | Agent和人工均可直接更新，有版本历史可回滚 |
| 归档/删除 | 仅人工通过ACP界面 |
| 模板管理 | 仅人工通过ACP界面 |

### 8.2 未来扩展

- 按域设置写入权限（供应链Agent只能写supply-chain/）
- confidential标记限制访问范围
- Agent写入信任等级（高信任Agent自动入库）
- 操作审计日志

## 9. 实现优先级

### 🔴 P0 — 第一版（核心读写）

**后端：**
- [ ] 知识文档CRUD + 版本管理（Git式：每次保存自动版本，diff，rollback）
- [ ] 知识域树形管理
- [ ] 文档模板CRUD
- [ ] **knowledge_write tool + 审批流程** ⭐
- [ ] knowledge_search API（调QMD，section-level）
- [ ] knowledge_list API
- [ ] 审批队列CRUD
- [ ] 写入后自动生成summary + qa_anchors（调LLM）
- [ ] QMD索引集成（shared-knowledge/目录，section chunking）

**前端：**
- [ ] 知识空间树形导航页
- [ ] 文档列表页（按域浏览）
- [ ] 文档详情页（内容渲染+关联文档+版本历史+Q&A锚点）
- [ ] 文档编辑页（Markdown编辑器）
- [ ] 搜索页 + 检索测试面板
- [ ] 审批队列页
- [ ] 模板管理页
- [ ] 侧边栏菜单整合

**Agent集成：**
- [ ] 注册knowledge_search/write/list到OpenClaw gateway
- [ ] 更新shared-bootstrap AGENTS.md使用指引

### 🟡 P1 — 第二版（增强）

- [ ] 文档关联双向链接（A引用B时，B自动显示被A引用）
- [ ] 标签管理与标签云
- [ ] 知识使用统计（哪些被搜索、被引用最多）
- [ ] 按域权限控制
- [ ] 资产管理（图片/附件上传和引用）
- [ ] 飞书通知集成（审批提醒、重要知识更新推送）

### 🟢 P2 — 未来

- [ ] 多格式解析入库（PDF/PPT → 自动md摘要）
- [ ] 知识生命周期管理（过期提醒、健康度评分）
- [ ] 知识图谱可视化（文档关联关系图）
- [ ] 跨知识库联邦搜索
- [ ] Agent写入自动入库（基于信任等级）

## 10. 技术约束

- ACP后端Go + SQLite，不引入新依赖
- 搜索走QMD（已部署），不额外部署向量数据库
- 知识文件存本地文件系统，不依赖外部存储
- 服务器资源有限（7.5G RAM），方案必须轻量
- LLM调用（summary/qa_anchors生成）走现有gateway API
