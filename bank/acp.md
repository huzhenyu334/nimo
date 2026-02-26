# ACP (Agent Control Panel)

> 内部AI Agent团队管理平台，泽斌发起，Lyra主导建设

## 基本信息
- 技术栈：Go + React + SQLite，零依赖单二进制
- 端口：3001 | 地址：http://43.134.86.237:3001
- 密码：openclaw2026
- 代码目录：/home/claw/.openclaw/workspace/agent-control-panel/
- systemd管理：`systemctl --user restart acp`
- 飞书App ID：cli_a9122d58b5f8dcca

## 定位
- **ACP是给Lyra用的**，泽斌只看前端页面
- 两套界面：泽斌=前端Web，Lyra=MCP接口（plugin tools）
- 目标：7×24持续工作，系统驱动agent团队不间断
- 核心理念："平台只是平台，核心是流程"（泽斌语）

## 引擎能力
- 结构化输入输出 + schema校验（422打回）
- Gate条件循环（所有step通用）
- Expand子任务集（PM输出JSON数组→引擎展开→逐个分配）
- Approval审批节点（飞书card + on_reject策略）
- Condition条件分支
- Escalate升级 + Blocked状态
- Lessons自动采集（4字段必填）
- Credential vault（AES-GCM加密）
- 崩溃恢复（已修4个bug）

## 工作流模板
- 新模块开发：6f0b1040（PRD→拆任务→expand开发→交付报告）
- 功能优化/修复：b0995080（发现→检查→审批→开发→验收→报告）v11

## Plugin Tools
14个原生agent tools，通过OpenClaw plugins.load.paths加载

## 性能优化（2026-02-21）
- Run详情API裁剪：1.3MB→66KB（agent_context+大output剥离）
- Runs列表N+1查询消除
- Gzip预压缩+Code splitting：首次加载3MB→449KB
- systemd user service管理（`systemctl --user restart acp`）

## 工作流v11首次完整运行（2026-02-21）
- 45min47s，20步全completed，0 failure，3 agent参与
- 修复10个PLM问题，3个commit
- dev_cycle占85%时间，Task1耗时异常(19min vs 7min)

## Lessons自动采集（2026-02-22）
- commit 5d6e75c: 4字段必填(context_sufficient/context_gap/obstacles/reusable_approach)
- 浓缩机制(condensed_lessons)待实现

## 版本历史
- v2.2: 审批系统+飞书OAuth
- v2.3: 审批升级+条件分支
- v2.4: 工作流停止+Agent进化模块
- v2.5: Flow审计系统+分析Dashboard

## 关键技术决策
- API响应：不返回agent_context，truncate output>2000字符
- 前端部署：`rm -rf web && cp -r acp-web/dist web`，生成.gz文件
- 模板变量限制：Max 32000/64000/100000 chars
- 泽斌ACP-app open_id：ou_e229cd56698a8e15e629af2447a8e0ed（和Lyra app不同！）
- Plugin vs MCP：新工具默认MCP，除非需要进程内能力
- 文件升级决策矩阵(commit 915b5ed)：frequency × scope → 目标文件
- 升级路径：Step lesson → bank/(3+次) → AGENTS.md(跨流程重复) → skills/(程序化)

## 知识体系设计参考（2026-02-22调研）
- Letta MemFS：system/永驻context + git版本 + sleep反思 + defrag
- Anthropic Context Engineering：context rot + 最小有效context + progressive disclosure
- OpenClaw hybrid search已启用：vectorWeight=0.7, textWeight=0.3

## 2026-02-24 更新

### Agent CRUD
- 通过config.patch实现（OpenClaw无独立agent CRUD RPC）
- Gateway统一将agent ID转小写（Bob → agent:bob:main）
- 创建agent需同步更新tools.agentToAgent.allow

### 远程能力
- /bash斜杠命令可远程执行shell（需commands.bash+tools.elevated）
- 远程skill安装通过agent.send + /bash heredoc实现
- 远程memory日志通过ExecBash读取
- agents.files.* RPC仅支持bootstrap白名单文件

### Node管理（Phase 1 PRD已写）
- Node列表/详情 + 环境探测 + 快速命令 + 工作流node字段
- PRD: agent-control-panel/docs/prd/acp-node-management-v1.md

### 移动端适配
- useIsMobile hook + mobile.css + 汉堡菜单，已部署

## 2026-02-25 更新

### 架构决策
- **废弃inline steps嵌套** → 改用`group:"xxx"`标签（纯前端分组，引擎不处理）
- **统一assignee** → assignee_type(agent/user/role)+assignee_id，废弃旧agent字段
- **Duration** → YAML友好格式(5d)，存储转ms，统一自然日
- **ACP=Agent OS** → Gateway是runtime，ACP是用户空间，支持多Gateway联邦管理

### 新功能上线
- Node监控页：从agent history过滤node相关tool调用（严格匹配nodes/exec(host=node)/process）
- Agent对话+Step过程的tool详情展示（修复后端丢弃toolCall参数和toolResult）
- 甘特图ClickUp风格SVG重写（双层刻度+Today标记+时间尺度切换月/周/天/时）
- 子流程+内联子步骤+统一assignee（后改为group标签）
- 时间管理duration/due/timeline API
- 人类任务飞书通知（飞书卡片）
- Role CRUD + 实例角色映射
- 前端skeleton loading（6个页面）+ gzip预压缩（3MB→860KB）+ 路由预取
- Schema API: GET /api/processes/schema → JSON Schema 2020-12（invopop/jsonschema从struct生成）
- Plugin工具扩展到32个（新增acp_process_schema + acp_knowledge_*）

### 统一资源调度PRD（v1.1）
- Resource Bus架构 + 全局资源注册表 + 路由策略
- 原生Plugin Tool集（acp_node_run/list/invoke, acp_agent_send/spawn等）
- ACL权限模型 + 审计日志
- Gateway UI整合计划（Config/Usage/Cron/Skills/Channels）
- 4 Phase路线图
- PRD: docs/prd/acp-unified-resource-bus.md

### 技术要点
- SIGUSR1热重载不加载新plugin，需完全重启服务器
- 前端build需2048MB（需先kill CC僵尸进程释放内存）
- 甘特图最小bar宽度：天级以上min 1h，hover显示真实耗时
