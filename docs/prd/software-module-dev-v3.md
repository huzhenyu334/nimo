# 软件模块开发流程 v3.0 — 重构 PRD

> 版本：v1.0 | 日期：2026-03-23
> 隶属：ACP 流程体系
> 流程文件：`flows/pm-software-dev.yaml`
> 流程 ID：`30bf4e22-1686-4cfb-9038-742f42a7e83e`
> 触发方式：手动
> 当前版本：v5（published）

---

## 一、现状问题分析

对当前 v5 版本（`软件开发 v2.0`）进行完整审查，发现以下 **15 个问题**，按严重程度分类：

### 1.1 严重问题（影响流程正确性）

| # | 问题 | 位置 | 影响 |
|---|------|------|------|
| **P1** | `input_form` 包含无关字段 `target`/`build_cmd`/`test_cmd` | Step: input_form | 这些字段属于项目架构层面，不应在立项表单中收集。编译/测试命令应来自架构 PRD 文档，由 CC 读取后写入项目 CLAUDE.md |
| **P2** | 缺少 CLAUDE.md 初始化步骤 | 全流程缺失 | CC 每次编码都缺乏项目上下文（技术栈、编译命令、目录结构、代码规范），只靠 prompt 注入不可持久化 |
| **P3** | `confirm_scope` 无拒绝门控 | Step: confirm_scope → create_milestone | 用户选择"取消"后流程仍继续执行 `create_milestone`，没有 if/switch 阻断 |
| **P4** | `breakdown` CC 缺少 `acp-jwt` 凭证 | Step: breakdown | prompt 中引用 `{{credentials.acp-jwt}}` 读取知识库，但 `credentials` 字段只有 `[anthropic-api-key]`，CC 无法访问知识库 API |
| **P5** | 迭代模式下工作目录未解析 | Step: dev_work | `{{steps.setup_workdir.output.stdout || vars.workspace}}` — 迭代模式无 `setup_workdir`，回退到 workspace 根目录而非项目目录 |
| **P6** | 迭代模式无代码同步步骤 | 迭代分支缺失 | 没有 `git pull` / `git fetch` 确保本地代码最新，CC 可能在旧代码上开发 |

### 1.2 中等问题（影响流程健壮性）

| # | 问题 | 位置 | 影响 |
|---|------|------|------|
| **P7** | Sprint 审核不通过是死胡同 | Step: review_gate false | 仅添加 PR 评论，没有重试机制。CC 无法根据反馈修改代码，整个 Sprint 作废 |
| **P8** | `query_existing_epics` 硬编码回退 `'FLOW'` | Step: query_existing_epics | 新建模式下 `project_key` 为空，查询 FLOW 项目的 epic 毫无意义 |
| **P9** | Epic 状态转换背靠背 | Steps: start_epic → close_epic | 在 Phase S8 收尾时才 `start_epic`，接着立刻 `close_epic`。Epic 应在第一个 Sprint 启动时就进入 `in_progress` |
| **P10** | 跨分支变量引用脆弱 | 多处 | `{{steps.create_pm_project.output.key || steps.input_form.output.form.project_key}}` 在 switch 分支外使用，模板解析可能不稳定 |
| **P11** | `sprint_loop` on_failure: skip 太宽松 | Step: sprint_loop | Sprint 失败直接跳过，后续 Sprint 可能依赖前序 Sprint 的成果，连锁失败 |

### 1.3 轻微问题（影响用户体验）

| # | 问题 | 位置 | 影响 |
|---|------|------|------|
| **P12** | `done_notify` 输出原始 JSON | Step: done_notify | `Sprint 数: {{sprints}}` 会 dump 整个 sprints 数组 |
| **P13** | 无部署步骤 | Phase S8 缺失 | 所有 PR 合并后没有构建/部署/验证环节 |
| **P14** | CC dev_work 未使用 `system_prompt_append` | Step: dev_work | CC 支持 `system_prompt_append` 注入项目规范，但未使用，每次 sprint 都在 prompt 里堆砌大量上下文 |
| **P15** | `setup_workdir` 的 `git clone` 缺少认证 | Step: setup_workdir | clone URL 如果是 HTTPS 需要 token，当前用 shell script 直接 clone 无 credential |

---

## 二、目标

重构流程 v3.0，解决以上全部问题。核心改进：

1. **精简信息收集** — 表单只收集与项目创建直接相关的信息
2. **新增 CLAUDE.md 初始化步骤** — CC 读取架构 PRD，生成项目级 CLAUDE.md，指导后续所有 Sprint 编码
3. **完善门控逻辑** — confirm_scope 拒绝能终止流程，Sprint 审核不通过能重试
4. **统一变量引用** — 消除跨分支变量引用的脆弱性
5. **补全凭证** — 所有需要 `acp-jwt` 的 CC 步骤都声明凭证
6. **增加部署步骤** — Sprint 全部完成后可选编译验证

### 2.1 设计原则

| 原则 | 说明 |
|------|------|
| **表单最小化** | 只收集必要的项目信息，技术细节由 CC 从架构文档提取 |
| **CLAUDE.md 即上下文** | 项目的编译、测试、目录结构等信息持久化在 CLAUDE.md，而非注入每次 prompt |
| **门控严格** | 每个决策点（scope 确认、Sprint 审核）都有明确的通过/拒绝路径 |
| **迭代模式一等公民** | 迭代模式不应是新建模式的降级版本，需要独立的工作目录检测和代码同步 |
| **失败可恢复** | Sprint 编码失败或审核不通过时，支持重试而非直接跳过 |

---

## 三、流程输入

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `approver` | string | * | 审批人 / 表单接收人 ID |

---

## 四、流程变量

```yaml
variables:
  workspace: "/home/claw/.openclaw/workspace"
  master_prd_doc_id: "208581f9-c273-498b-89e1-f2d19e3e0c25"
  max_issues: 30
  default_base_branch: "main"
  github_org: "nicobao"
  max_review_retries: 2
```

| 变量 | 用途 |
|------|------|
| `workspace` | 项目工作空间根目录 |
| `master_prd_doc_id` | Master PRD 文档 ID（模块注册表） |
| `max_issues` | 单次开发最大 Issue 数 |
| `default_base_branch` | 默认基准分支 |
| `github_org` | GitHub 组织名 |
| `max_review_retries` | Sprint 审核不通过最大重试次数 |

---

## 五、流程设计

### Phase 1: 信息收集

**Step 1.1 — input_form（精简版）**

去掉 `target`/`build_cmd`/`test_cmd`，只保留项目创建相关字段。新增 `architecture_doc_id` 用于指定架构 PRD（含编译/部署等信息）。

```yaml
- id: input_form
  name: 开发信息
  executor: human
  command: form
  input:
    assignee: "{{inputs.approver}}"
    title: 软件开发
    prompt: |
      选择新建项目或迭代已有项目。
      - 新建: 自动创建 GitHub 仓库、PM 项目、工作目录
      - 迭代: 在已有项目上创建新的 Milestone/Sprint/Issue 进行开发
    form:
      fields:
        - key: mode
          label: 模式
          type: select
          required: true
          options:
            - {value: new, label: "新建项目"}
            - {value: iterate, label: "迭代已有项目"}
        - key: prd_doc_id
          label: PRD 文档 ID
          type: text
          required: true
          description: "知识库中需求 PRD 文档 ID"
        - key: architecture_doc_id
          label: 架构 PRD 文档 ID
          type: text
          description: "架构 PRD 文档 ID（含技术栈、编译、部署指令），CC 将从此文档提取信息写入项目 CLAUDE.md"
        - key: project_key
          label: PM 项目 Key
          type: text
          description: "迭代模式必填：已有 PM 项目 Key（如 NIMO）"
        - key: project_name
          label: 项目名称
          type: text
          description: "新建模式必填：项目名称"
        - key: project_description
          label: 项目描述
          type: textarea
          description: "新建模式：项目描述"
        - key: repo_name
          label: 仓库名称
          type: text
          description: "新建模式必填：GitHub 仓库名（如 my-project）"
        - key: repo
          label: GitHub 仓库
          type: text
          description: "迭代模式必填：已有仓库 owner/repo 格式（如 nicobao/nimo-workspace）"
```

**变更说明**：
- 删除 `target`（backend/frontend/fullstack）— 由架构 PRD 决定
- 删除 `build_cmd` — 由 CC 从架构 PRD 提取
- 删除 `test_cmd` — 由 CC 从架构 PRD 提取
- 新增 `architecture_doc_id` — 架构 PRD 文档引用

---

### Phase 2: 模式路由 + 项目初始化

**Step 2.0 — 加载模块注册表（并行前置）**

```yaml
- id: read_master_registry
  name: 读取模块注册表
  executor: knowledge
  command: fetchsection
  depends_on: [input_form]
  input:
    doc_id: "{{vars.master_prd_doc_id}}"
    heading: "八、模块注册表"
```

**Step 2.1 — mode_route（新建 vs 迭代分支）**

两个分支的末尾都生成统一的工作目录和项目信息，供后续步骤引用。

```yaml
- id: mode_route
  control: switch
  depends_on: [input_form]
  expression: "{{steps.input_form.output.form.mode}}"
  branches:
    new:
      # 2.1a — 创建 GitHub 仓库
      - id: create_repo
        name: 创建仓库
        executor: github
        command: create_repo
        credentials: [github-token]
        input:
          name: "{{steps.input_form.output.form.repo_name}}"
          org: "{{vars.github_org}}"
          description: "{{steps.input_form.output.form.project_description}}"
          private: true
          gitignore: Go
          credential: "{{credentials.github-token}}"

      # 2.1b — 创建 PM 项目
      - id: create_pm_project
        name: 创建项目
        executor: pm
        command: create_project
        depends_on: [create_repo]
        input:
          name: "{{steps.input_form.output.form.project_name}}"
          key: "{{steps.input_form.output.form.repo_name}}"
          description: "{{steps.input_form.output.form.project_description}}"

      # 2.1c — 关联 PRD 文档
      - id: link_prd
        name: 关联 PRD
        executor: pm
        command: link_project_doc
        depends_on: [create_pm_project]
        input:
          project_key: "{{steps.create_pm_project.output.key}}"
          doc_id: "{{steps.input_form.output.form.prd_doc_id}}"

      # 2.1d — 关联 GitHub Repo
      - id: link_repo
        name: 关联仓库
        executor: pm
        command: add_project_repo
        depends_on: [create_pm_project]
        credentials: [github-token]
        input:
          project_key: "{{steps.create_pm_project.output.key}}"
          repo: "{{steps.create_repo.output.full_name}}"
          is_default: true
          credential: "{{credentials.github-token}}"

      # 2.1e — 克隆仓库到工作目录
      - id: setup_workdir
        name: 初始化工作目录
        executor: shell
        command: script
        depends_on: [create_repo]
        credentials: [github-token]
        input:
          script: |
            WORKDIR="{{vars.workspace}}/{{steps.input_form.output.form.repo_name}}"
            mkdir -p "$WORKDIR"
            cd "$WORKDIR"
            git clone "https://{{credentials.github-token}}@github.com/{{vars.github_org}}/{{steps.input_form.output.form.repo_name}}.git" . 2>/dev/null || echo "already cloned"
            echo "$WORKDIR"

    iterate:
      # 2.2a — 获取项目信息
      - id: get_project_info
        name: 项目信息
        executor: pm
        command: get_project_stats
        input:
          project_key: "{{steps.input_form.output.form.project_key}}"

      # 2.2b — 获取关联仓库
      - id: list_repos
        name: 项目仓库
        executor: pm
        command: list_project_repos
        input:
          project_key: "{{steps.input_form.output.form.project_key}}"

      # 2.2c — 确保工作目录存在并拉取最新代码
      - id: sync_workdir
        name: 同步工作目录
        executor: shell
        command: script
        depends_on: [list_repos]
        credentials: [github-token]
        input:
          script: |
            REPO="{{steps.input_form.output.form.repo}}"
            REPO_NAME=$(echo "$REPO" | cut -d'/' -f2)
            WORKDIR="{{vars.workspace}}/$REPO_NAME"
            if [ -d "$WORKDIR/.git" ]; then
              cd "$WORKDIR"
              git fetch origin
              git pull origin {{vars.default_base_branch}} --ff-only 2>/dev/null || true
            else
              mkdir -p "$WORKDIR"
              cd "$WORKDIR"
              git clone "https://{{credentials.github-token}}@github.com/$REPO.git" .
            fi
            echo "$WORKDIR"
```

**关键改进**：
- 迭代模式新增 `sync_workdir`：clone 或 pull 确保代码最新
- clone 命令注入 github-token 解决 HTTPS 认证问题
- 两个分支都生成 `stdout` 输出工作目录路径

---

### Phase 3: 初始化 CLAUDE.md（NEW）

**Step 3.1 — resolve_workdir（统一工作目录变量）**

用 shell 统一两个分支的工作目录输出，避免后续步骤跨分支引用。

```yaml
- id: resolve_workdir
  name: 解析工作目录
  executor: shell
  command: script
  depends_on: [mode_route]
  input:
    script: |
      # 新建模式输出在 setup_workdir，迭代模式在 sync_workdir
      if [ -n "{{steps.setup_workdir.output.stdout}}" ]; then
        echo "{{steps.setup_workdir.output.stdout}}"
      elif [ -n "{{steps.sync_workdir.output.stdout}}" ]; then
        echo "{{steps.sync_workdir.output.stdout}}"
      else
        REPO_NAME="{{steps.input_form.output.form.repo_name || steps.input_form.output.form.repo}}"
        REPO_NAME=$(echo "$REPO_NAME" | sed 's|.*/||')
        echo "{{vars.workspace}}/$REPO_NAME"
      fi
```

**Step 3.2 — resolve_project_key（统一项目 Key）**

```yaml
- id: resolve_project_key
  name: 解析项目 Key
  executor: shell
  command: script
  depends_on: [mode_route]
  input:
    script: |
      if [ -n "{{steps.create_pm_project.output.key}}" ]; then
        echo "{{steps.create_pm_project.output.key}}"
      else
        echo "{{steps.input_form.output.form.project_key}}"
      fi
```

**Step 3.3 — resolve_repo（统一仓库名）**

```yaml
- id: resolve_repo
  name: 解析仓库名
  executor: shell
  command: script
  depends_on: [mode_route]
  input:
    script: |
      if [ -n "{{steps.create_repo.output.full_name}}" ]; then
        echo "{{steps.create_repo.output.full_name}}"
      else
        echo "{{steps.input_form.output.form.repo}}"
      fi
```

**Step 3.4 — init_claude_md（CC 初始化 CLAUDE.md）**

这是本次重构的**核心新增步骤**。CC 读取架构 PRD（如果有），分析项目代码结构，生成项目级 CLAUDE.md。

```yaml
- id: init_claude_md
  name: 初始化 CLAUDE.md
  executor: cc
  command: run
  depends_on: [resolve_workdir, resolve_project_key, resolve_repo]
  timeout: 15m
  credentials: [anthropic-api-key, acp-jwt]
  input:
    model: claude-sonnet-4-6
    workdir: "{{steps.resolve_workdir.output.stdout}}"
    prompt: |
      你的任务是为当前项目生成或更新 CLAUDE.md 文件，让后续 CC 编码步骤有完整的项目上下文。

      ## 架构 PRD（如有）
      架构 PRD 文档 ID: {{steps.input_form.output.form.architecture_doc_id}}
      如果 ID 非空，通过以下命令读取：
      curl -s http://localhost:3001/api/knowledge/{{steps.input_form.output.form.architecture_doc_id}} \
        -H "Authorization: Bearer {{credentials.acp-jwt}}" | jq -r '.document.content'

      请从架构 PRD 中提取以下信息（如果文档存在）：
      - 技术栈（语言、框架、数据库等）
      - 编译命令
      - 测试命令
      - 部署步骤
      - 代码规范 / lint 配置
      - 目录结构约定

      ## 需求 PRD（项目用途参考）
      需求 PRD 文档 ID: {{steps.input_form.output.form.prd_doc_id}}
      读取摘要：
      curl -s "http://localhost:3001/api/knowledge/{{steps.input_form.output.form.prd_doc_id}}/sections" \
        -H "Authorization: Bearer {{credentials.acp-jwt}}"

      ## 当前目录结构
      请先运行 `ls -la` 和 `find . -maxdepth 3 -type f | head -50` 了解项目状态。
      如果有现成的 CLAUDE.md、README.md、Makefile、go.mod、package.json 等，先读取理解。

      ## CLAUDE.md 应包含
      1. **项目概述** — 一句话说明项目是什么
      2. **技术栈** — 语言、框架、主要依赖
      3. **目录结构** — 关键目录和文件说明
      4. **编译命令** — 从架构 PRD 或项目推断
      5. **测试命令** — 从架构 PRD 或项目推断
      6. **部署步骤**（如有） — 从架构 PRD 获取
      7. **开发规范** — 代码风格、命名约定
      8. **常用命令速查**

      ## 规则
      - 如果 CLAUDE.md 已存在，在现有内容基础上补充缺失部分，不要丢失已有信息
      - 如果是全新项目（空仓库），基于架构 PRD 创建骨架 CLAUDE.md
      - 如果没有架构 PRD 且项目为空，创建最小 CLAUDE.md 模板
      - 写入完成后 commit: git add CLAUDE.md && git commit -m "chore: init/update CLAUDE.md"
      - 不要 push（后续 Sprint 步骤会 push）
```

**为什么这个步骤重要**：
- CC 的 `system_prompt_append` 会自动加载工作目录下的 CLAUDE.md
- 一次初始化，后续所有 Sprint 的 CC 都能读到项目规范
- 编译/测试命令从架构 PRD 提取，不需要人在表单里手动填写
- 迭代模式下也能更新 CLAUDE.md（比如项目新增了前端模块）

---

### Phase 4: AI 拆解 PRD

**Step 4.1 — query_existing_epics（修复查询范围）**

```yaml
- id: query_existing_epics
  name: 查询已有 Epic
  executor: pm
  command: query_issues
  depends_on: [resolve_project_key]
  on_failure: skip
  input:
    project_key: "{{steps.resolve_project_key.output.stdout}}"
    type: epic
    limit: 100
```

**修复**：用 `resolve_project_key` 统一输出替代硬编码 `'FLOW'` 回退。

**Step 4.2 — breakdown（修复凭证）**

```yaml
- id: breakdown
  name: AI 拆解 PRD
  executor: cc
  command: structured_output
  depends_on: [read_master_registry, query_existing_epics, init_claude_md]
  timeout: 30m
  credentials: [anthropic-api-key, acp-jwt]
  input:
    model: claude-opus-4-6
    workdir: "{{steps.resolve_workdir.output.stdout}}"
    schema: |
      { ... }  # 同现有 schema，略
    prompt: |
      ... # 同现有 prompt，但凭证引用已正确
```

**修复**：`credentials` 添加 `acp-jwt`；`depends_on` 包含 `init_claude_md`（确保 CLAUDE.md 就绪）。

---

### Phase 5: 确认 Scope + 门控

**Step 5.1 — confirm_scope（同现有）**

```yaml
- id: confirm_scope
  name: 确认 Scope
  executor: human
  command: form
  depends_on: [breakdown]
  input:
    assignee: "{{inputs.approver}}"
    title: "确认开发 Scope"
    prompt: |
      ## 摘要
      {{steps.breakdown.output.structured_output.summary}}

      ## Milestone / Epic / Sprint 计划
      {{steps.breakdown.output.structured_output.milestone}}
      {{steps.breakdown.output.structured_output.epic}}
      {{steps.breakdown.output.structured_output.sprints}}

      请确认是否按此方案执行。
    form:
      fields:
        - key: decision
          label: 决定
          type: select
          required: true
          options:
            - {value: approve, label: 确认执行}
            - {value: reject, label: 取消}
```

**Step 5.2 — scope_gate（NEW — 门控拒绝）**

```yaml
- id: scope_gate
  control: if
  depends_on: [confirm_scope]
  condition: "{{steps.confirm_scope.output.form.decision}} == approve"
  branches:
    "true":
      - id: scope_approved
        name: Scope 已确认
        executor: shell
        command: exec
        input:
          command: "echo approved"
    "false":
      - id: scope_rejected_notify
        name: 流程取消通知
        executor: human
        command: notification
        input:
          assignee: "{{inputs.approver}}"
          title: "软件开发已取消"
          prompt: "Scope 确认被拒绝，流程终止。"
```

**关键**：后续 Phase 6/7/8 全部 `depends_on: [scope_approved]`，false 分支只发通知后流程自然结束。

---

### Phase 6: 创建 PM 结构

**Step 6.1 — create_milestone**

```yaml
- id: create_milestone
  name: 创建 Milestone
  executor: pm
  command: create_milestone
  depends_on: [scope_approved]
  input:
    project_key: "{{steps.resolve_project_key.output.stdout}}"
    title: "{{steps.breakdown.output.structured_output.milestone.title}}"
    description: "{{steps.breakdown.output.structured_output.milestone.description}}"
    due_date: "{{steps.breakdown.output.structured_output.milestone.due_date}}"
```

**Step 6.2 — create_epic**

```yaml
- id: create_epic
  name: 创建 Epic
  executor: pm
  command: create_issue
  depends_on: [create_milestone]
  input:
    project_key: "{{steps.resolve_project_key.output.stdout}}"
    title: "{{steps.breakdown.output.structured_output.epic.title}}"
    type: epic
    priority: "{{steps.breakdown.output.structured_output.epic.priority || 'P1'}}"
    description: "{{steps.breakdown.output.structured_output.epic.description}}"
    milestone_id: "{{steps.create_milestone.output.milestone_id}}"
```

**Step 6.3 — link_epic_prd / link_epic_process**（同现有，略）

**Step 6.4 — start_epic（移到此处，不再在收尾阶段）**

```yaml
- id: start_epic
  name: 开始 Epic
  executor: pm
  command: transition_issue
  depends_on: [create_epic]
  input:
    issue_id: "{{steps.create_epic.output.issue_id}}"
    status: in_progress
```

**修复**：Epic 在创建后立即进入 `in_progress`，而非在所有 Sprint 完成后。

---

### Phase 7: Sprint 迭代开发

**Step 7.0 — sprint_loop**

```yaml
- id: sprint_loop
  control: foreach
  depends_on: [start_epic]
  items: "{{steps.breakdown.output.structured_output.sprints}}"
  as: sprint
  concurrency: 1
  on_failure: pause
  steps:
    # 7.1-7.5 同现有（create_sprint, create_sprint_issues, add_to_sprint, start_sprint, create_branch）
    # 但所有 project_key 引用改为 {{steps.resolve_project_key.output.stdout}}
    # 所有 repo 引用改为 {{steps.resolve_repo.output.stdout}}
    ...
```

**修复**：`on_failure: pause`（而非 skip），失败时暂停让人工介入。

**Step 7.6 — dev_work（精简 prompt，依赖 CLAUDE.md）**

```yaml
- id: dev_work
  name: AI 编码
  executor: cc
  command: run
  depends_on: [create_branch]
  timeout: 60m
  credentials: [anthropic-api-key, acp-jwt]
  input:
    model: claude-opus-4-6
    workdir: "{{steps.resolve_workdir.output.stdout}}"
    prompt: |
      你正在执行 Sprint "{{sprint.name}}" 的开发任务。

      ## Sprint 目标
      {{sprint.goal}}

      ## 待实现的 Issue 列表
      {{sprint.issues}}

      ## PRD 文档（按需读取）
      PRD 文档 ID: {{steps.input_form.output.form.prd_doc_id}}
      - 章节目录: curl -s http://localhost:3001/api/knowledge/{{steps.input_form.output.form.prd_doc_id}}/sections -H "Authorization: Bearer {{credentials.acp-jwt}}"
      - 读取章节: curl -s "http://localhost:3001/api/knowledge/{{steps.input_form.output.form.prd_doc_id}}/section?heading=章节名" -H "Authorization: Bearer {{credentials.acp-jwt}}"

      ## Git 操作
      git fetch origin
      git checkout {{sprint.branch_name}} || git checkout -b {{sprint.branch_name}} origin/{{vars.default_base_branch}}

      ## 编译和测试命令
      参照项目根目录的 CLAUDE.md（已自动加载）。

      ## 任务要求
      1. 切换到分支 {{sprint.branch_name}}
      2. 按 Issue 优先级顺序逐个实现
      3. 每完成一组改动就 commit（commit message 引用 Issue 标题）
      4. 编译必须通过
      5. 测试必须通过
      6. push 到远程: git push origin {{sprint.branch_name}}
      7. 完成后输出每个 Issue 的实现摘要
```

**关键改进**：
- 删除了 `target`/`build_cmd`/`test_cmd` 注入 — 这些信息在 CLAUDE.md 中
- CC 的 `system_prompt_append` 会自动加载 workdir 下的 CLAUDE.md
- prompt 更精简，聚焦于 Sprint 任务而非重复技术细节

**Step 7.7-7.8 — create_pr, review**（同现有，略）

**Step 7.9 — review_loop（NEW — 替代 review_gate，支持重试）**

当审核不通过时，CC 根据反馈修改代码，重新提交 PR，再次审核。

```yaml
- id: review_loop
  control: loop
  depends_on: [create_pr]
  condition: "{{steps.review_loop.review_result.output.form.result}} != pass"
  mode: do-while
  max_iterations: "{{vars.max_review_retries}}"
  steps:
    # 人工审核
    - id: review_result
      name: 审核
      executor: human
      command: form
      input:
        assignee: "{{inputs.approver}}"
        title: "审核: {{sprint.name}} (第{{loop.iteration}}轮)"
        prompt: |
          ## Sprint: {{sprint.name}}
          ## PR: {{steps.create_pr.output.url}}
          ## 开发结果
          {{steps.dev_work.output.summary}}
          {{steps.dev_work.output.diff_stat}}
        form:
          fields:
            - key: result
              label: 审核结果
              type: select
              required: true
              options:
                - {value: pass, label: 通过}
                - {value: fail, label: 不通过}
            - key: comment
              label: 修改意见
              type: textarea

    # 审核不通过 → CC 修改
    - id: fix_gate
      control: if
      depends_on: [review_result]
      condition: "{{steps.review_result.output.form.result}} == fail"
      branches:
        "true":
          - id: fix_work
            name: 修改代码
            executor: cc
            command: run
            timeout: 30m
            credentials: [anthropic-api-key, acp-jwt]
            input:
              model: claude-opus-4-6
              workdir: "{{steps.resolve_workdir.output.stdout}}"
              prompt: |
                PR 审核未通过，请根据以下反馈修改代码：

                ## 审核意见
                {{steps.review_result.output.form.comment}}

                ## 当前分支
                git checkout {{sprint.branch_name}}

                ## 原始 Issue 列表
                {{sprint.issues}}

                修改后 commit 并 push。
        "false":
          - id: review_passed
            name: 审核通过
            executor: shell
            command: exec
            input:
              command: "echo passed"
```

**Step 7.10 — merge_pr（审核通过后合并）**

```yaml
- id: merge_pr
  name: 合并 PR
  executor: github
  command: merge_pr
  depends_on: [review_loop]
  credentials: [github-token]
  input:
    repo: "{{steps.resolve_repo.output.stdout}}"
    pr_number: "{{steps.create_pr.output.pr_number}}"
    merge_method: squash
    delete_branch: true
    credential: "{{credentials.github-token}}"
```

**Step 7.11 — close_sprint_issues + complete_sprint**（同现有，略）

---

### Phase 8: 收尾

**Step 8.1 — close_epic（不再需要 start_epic，已在 Phase 6 完成）**

```yaml
- id: close_epic
  name: 关闭 Epic
  executor: pm
  command: transition_issue
  depends_on: [sprint_loop]
  input:
    issue_id: "{{steps.create_epic.output.issue_id}}"
    status: done
    comment: "所有 Sprint 完成"
```

**Step 8.2 — close_milestone**

```yaml
- id: close_milestone
  name: 关闭 Milestone
  executor: pm
  command: close_milestone
  depends_on: [close_epic]
  input:
    milestone_id: "{{steps.create_milestone.output.milestone_id}}"
```

**Step 8.3 — build_verify（NEW — 可选编译验证）**

```yaml
- id: build_verify
  name: 编译验证
  executor: cc
  command: run
  depends_on: [close_milestone]
  timeout: 15m
  credentials: [anthropic-api-key]
  on_failure: skip
  input:
    model: claude-sonnet-4-6
    workdir: "{{steps.resolve_workdir.output.stdout}}"
    prompt: |
      所有 Sprint 已合并到 {{vars.default_base_branch}}。
      请 checkout {{vars.default_base_branch}}，拉取最新代码，按 CLAUDE.md 中的编译命令做一次完整编译验证。
      如果失败，输出详细错误信息。
      不要修改代码，只验证。
```

**Step 8.4 — done_notify（修复输出格式）**

```yaml
- id: done_notify
  name: 完成通知
  executor: human
  command: notification
  depends_on: [build_verify]
  input:
    assignee: "{{inputs.approver}}"
    title: "软件开发完成"
    prompt: |
      ## 完成
      项目: {{steps.breakdown.output.structured_output.epic.title}}
      Milestone: {{steps.breakdown.output.structured_output.milestone.title}}
      全部 Sprint 已完成，PR 已合并，Milestone 已关闭。

      ## 编译验证
      {{steps.build_verify.output.summary}}
```

**修复**：不再 dump `sprints` 数组原始 JSON，改为人类可读摘要。

---

## 六、错误处理与边界

| 场景 | 处理方式 |
|------|---------|
| GitHub 仓库创建失败 | on_failure: abort（默认），流程终止 |
| PM 项目创建失败 | on_failure: abort，流程终止 |
| AI 拆解超时 | timeout: 30m，超时后 abort |
| Scope 确认拒绝 | scope_gate false 分支通知后自然结束 |
| Sprint CC 编码失败 | sprint_loop on_failure: pause，人工介入 |
| Sprint 审核不通过 | review_loop 最多重试 max_review_retries 次 |
| 编译验证失败 | on_failure: skip，不阻断完成通知 |
| 迭代模式 project_key 无效 | get_project_info 返回错误，abort |
| 架构 PRD 不存在 | init_claude_md 中 CC 会 gracefully 处理空 ID |

---

## 七、变更汇总

### 7.1 删除的字段

| 字段 | 原位置 | 原因 |
|------|--------|------|
| `target` | input_form | 由架构 PRD 决定技术栈，非用户手动选择 |
| `build_cmd` | input_form | 由 CC 从架构 PRD 提取写入 CLAUDE.md |
| `test_cmd` | input_form | 同上 |

### 7.2 新增的步骤

| Step ID | 类型 | 用途 |
|---------|------|------|
| `sync_workdir` | shell | 迭代模式下 clone 或 pull 代码 |
| `resolve_workdir` | shell | 统一两分支的工作目录输出 |
| `resolve_project_key` | shell | 统一两分支的 project_key |
| `resolve_repo` | shell | 统一两分支的 repo 名 |
| `init_claude_md` | cc | 读取架构 PRD，生成项目 CLAUDE.md |
| `scope_gate` | if | confirm_scope 拒绝门控 |
| `review_loop` | loop | Sprint 审核 + 重试机制 |
| `fix_work` | cc | 审核不通过时 CC 修改代码 |
| `build_verify` | cc | 最终编译验证 |

### 7.3 修复的步骤

| Step ID | 修复内容 |
|---------|---------|
| `input_form` | 删除 target/build_cmd/test_cmd，新增 architecture_doc_id |
| `breakdown` | credentials 添加 acp-jwt |
| `query_existing_epics` | project_key 改用 resolve_project_key |
| `start_epic` | 从 Phase S8 移到 Phase 6（Epic 创建后立即开始） |
| `dev_work` | prompt 精简，依赖 CLAUDE.md；删除 target/build_cmd/test_cmd 引用 |
| `sprint_loop` | on_failure 从 skip 改为 pause |
| `done_notify` | 输出人类可读摘要 |
| `setup_workdir` | git clone 加入 github-token 认证 |
| 所有步骤 | project_key/repo/workdir 统一用 resolve_* 步骤输出 |

### 7.4 删除的步骤

| Step ID | 原因 |
|---------|------|
| `start_epic`（Phase S8 版本） | 已移到 Phase 6，不需要背靠背 start→close |
| `review_gate` | 被 `review_loop` 替代 |

---

## 八、预期产物

流程完成后产生：
1. **PM 项目** — 含 Milestone、Epic、Sprint、Issue 完整结构
2. **GitHub 仓库** — 含多个已合并的 Sprint PR
3. **CLAUDE.md** — 项目级配置文件，指导 CC 编码
4. **代码** — 通过编译和测试的实现代码

---

## 九、实现优先级

| 优先级 | 变更 | 理由 |
|--------|------|------|
| **P0** | 精简 input_form + 新增 init_claude_md | 用户明确要求，解决 CC 上下文问题 |
| **P0** | 修复 breakdown 凭证 | CC 无法读取知识库是阻断性 bug |
| **P0** | 新增 scope_gate | 拒绝后仍执行是严重 bug |
| **P0** | 统一变量引用（resolve_*） | 消除跨分支引用的脆弱性 |
| **P1** | 迭代模式 sync_workdir | 确保代码最新 |
| **P1** | review_loop 重试机制 | 提升审核不通过时的可恢复性 |
| **P2** | build_verify | 可选的最终验证 |
| **P2** | done_notify 格式优化 | 用户体验改善 |
