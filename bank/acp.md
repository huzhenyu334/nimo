# ACP (Agent Control Panel)

> 内部AI Agent团队管理平台，泽斌发起，Lyra主导建设

## 2026-03-09 更新

### Debug Console Agent文字修复
- 最终方案B：只用chat event final作为text来源（commit 471ed81，净删93行）
- `forwardChatEventToDebug`：只在state="final"时emit agent_output
- 删掉chatDebugState/chatThrottle/200ms sleep等所有hack
- 教训：泽斌要简单方案，我过度设计了（delta段检测、assistant stream回退等）

### Compaction失忆修复
- 根因：跨session失忆（daily reset凌晨4点UTC创建新session，旧20k不可见）
- 修复：session.reset.mode="never"（永不重置，靠compaction管理）
- AGENTS.md Session Startup只处理/new手动重置的情况

### NocoDB + 软件项目管理体系
- NocoDB部署：43.134.86.237:8090，docker host网络
- 软件项目表(mk4cbfl27dbfhlu)：PLM/ACP/SRM三项目
- 文档注册表(mjap0r8h9nxrb67)：13篇知识库文档已注册
- NocoDB Script Executor：nocodb.js，6命令（list/get/create/update/delete/search）
- 泽斌决策：NocoDB管结构化数据，ACP知识库管非结构化文档

### 流程编写Agent（Bob）
- Bob workspace: /home/claw/.openclaw/workspace-Bob/
- 知识分层：L1(StepDef/控制节点)→L2(常用executor)→L3(executor索引)→L4(按需查schema API)
- 4个新plugin tools: acp_list_executors, acp_get_executor_schema, acp_validate_yaml, acp_get_process_yaml
- TOOLS.md 443行/11.7KB：全部11个executor+5个控制节点完整参考卡
- 动态注入模式：http executor拉executor-metas→变量引用注入agent prompt

### 流程体系建设
- 流程需求收集(82866aeb)：11维度表单→agent生成规格书→NocoDB注册→人工确认
- 软件项目立项(80f6194e) v7：代码扫描用subprocess调子流程
- 代码扫描子流程(9286f334)、文档创建与注册子流程(80d2f70d)
- acp-flow-builder skill：5个reference files（design-guide/executor-reference/stepdef-reference/common-pitfalls/examples）
- 流程设计规范知识库(14b2f6dd)、NocoDB Executor知识库(da19b34a)

### Output数据模型重构
- 统一output列，移除5个冗余DB列（structured_output, outputs, executor_output, step_data, executor_type）
- 每个executor独立output命名空间：human/form→output.form.xxx, agent→output.structured_output.xxx
- 模板引擎：reOutputKey支持多级JSON路径 {{steps.X.output.key.subkey}}
- CC budget移除default:5默认值

### ValidateConfig自动化
- schema-driven required校验：ValidateRequiredFields()从InputSchema自动提取required (commit aed4c18)
- agent/human assignee改为必填plain string（commit 086bf5a）
- 嵌套step递归校验（commit 086bf5a）

### Loop/分支bug修复系列
- BFS cascade-skip：被skip分支的所有downstream也被skip (commit b1186b1)
- loop body中resumeLoopBody在迭代完成后emit step_completed (commit f3b3f98)
- loop completedBodyIdx搜索需遍历branches值 (commit b27f04d，调试中)
- scoped loop/foreach body steps的form fields解析 (commit dfdb3e9)
- buildLoopBodyRC helper统一所有loop body RunContext构建 (commit c03e602)

## 2026-03-08 更新

### CC Executor跑通 + 引擎控制节点修复
- CC Executor端到端通过：10个fix commits（nil pointer、CJS bundle、stderr类型、sessionId等）
- 引擎控制节点6个commits：AdvanceDAG支持控制节点、flatten层级、outputs持久化DB、TriggerStep异步dispatch、CC系统性修复、loop/foreach完成时persistFlowVariables
- ppp流程（if/switch/loop/foreach/三层嵌套if）全部通过
- SQLite DB损坏(.recover修复) + WAL模式

### Agent Executor重构（方案D）
- CC重构：596→125行（commit d83a29d），但过度简化需手动补字段（commit 0b951b2）
- 方案D：callback_url+JWT注入→agent POST结构化数据→lifecycle.end触发步骤完成
- Gateway WebSocket事件流完美覆盖agent全生命周期
- 已有基础设施：gateway/client.go WebSocket+事件解析、OnSessionEnd回调
- Debug事件转发链：WriteEvent→SSE、agent lifecycle/tool_call/tool_result

### cc-dev-task流程
- 两层设计：cc-dev-task(模块级) + project-dev(项目级)
- cc-dev-task v9：task_input(form)→cc_code(CC)→review_code(form+switch三选)→push/rollback/skip→result
- YAML踩坑：assignee在input内、human需title、foreach用items/as、subprocess的process在顶层
- 首次实战：用cc-dev-task驱动CC开发agent executor

### Recovery控制节点bug修复
- 根因：resumeRun把所有步骤reset为pending，CAS guard跳过非root pending步骤
- 修复：有依赖步骤reset为waiting，root保持pending，human/agent保持pending不清output
- 部署时重启ACP破坏了泽斌正在操作的流程实例（严重教训→写入AGENTS.md）

### 服务器基础设施
- 腾讯云S5.MEDIUM8（2核8G），新加坡ap-singapore-3
- OpenClaw高负载时占满CPU（100%+571MB），导致ACP变慢
- nginx已配443反代到3001（自签名证书），考虑Cloudflare Tunnel
- API Key修复流程：POST /api/auth/api-keys → ACP自动推到gateway → restart gateway

### 前端修复
- 导航栏超出→Nodes/System Logs移到Settings子菜单
- 甘特图拓扑排序（Kahn's算法）+ loop内嵌套if/switch树构建
- 知识库TOC：左侧sticky、slug队列预计算、scroll事件联动
- 全局滚动条暗色风格（6px宽、半透明白色）
- on_submit移除会导致前端command消失（schema $defs结构变化影响parse）→留到重构时改

## 2026-03-04 更新

### Debug模式架构完善
- **DAG并行断点**：PausedSteps map[string]string支持多步同时暂停，sendAction恢复后自动切下一个
- **断点暂停时间排除**：重置started_at为恢复时间+清零StepPausedDuration，匹配VS Code行为
- **SSE replay机制**：eventHistory缓冲+Subscribe()时replay历史事件，解决执行速度>SSE连接速度
- **前端状态：SSE是唯一状态源**，action handler不应强制setState覆盖SSE推送

### Loop/Foreach引擎重构
- **Loop body同步重构**：从异步goroutine+轮询→同步ExecuteStep()直接调用（与foreach统一）
- Scoped task ID（`loop_double-3-loop_worker`）替代复用+归档模式
- Output映射：scoped output拷回原始body step ID
- 嵌套控制流：loop/foreach body支持if/switch/loop/foreach（commit d936f91）

### 流程级变量+Output Mapping
- WorkflowDef.Variables初始值，StepDef.OutputMapping映射规则
- RunContext.Vars全局变量，RenderPrompt支持`{{vars.xxx}}`
- 泽斌三层递进：变量→Output Mapping→脚本Executor

### 可视化改进
- if/switch分支容器：Post-dominator算法检测分支归属，ConditionGroupNode组件
- 甘特图树形重写：TreeNode递归结构，真实wall-clock时间轴，服务端时间戳
- Loop/Foreach展示统一：iteration→body steps嵌套，迭代>10条截断(前5+后5)

### 审计系统（统一Event模型）
- **step_events表**统一log+审计，summary/payload分离，parent_id嵌套
- Executor API：`executor.Emit(ctx, type, summary, payload)` / `EmitChild(ctx, parentID, ...)`
- Phase 1（commit e757b7e）：entity+API+引擎自动采集+前端Events section
- Phase 2（commit 360c600）：控制流事件(condition_eval/branch_selected/iteration/variable_changed/failure_handled)
- 审计看板三层：流程图热力图→Step下钻→性能火焰图

### 性能优化
- **SQLite WAL模式**：30-799ms→4-8ms（100x提升），synchronous=NORMAL
- **单写入队列**：消除async goroutine写锁竞争（dbWriteQueue chan 1024）
- **Loop同步调用**：927ms→364ms
- Inspect API缓存+去重，breakpoints useMemo稳定引用

### 技术要点
- RenderPrompt是简单字符串替换，不支持管道语法`| default:`
- PC端DebugStepGrid和移动端DebugTimeline是完全独立的组件
- foreach body需在task创建前完成input渲染（否则存原始模板）
- Loop scoped body events：每次迭代为body step发射scoped step_started/step_completed

## 2026-03-03 更新

### Control Flow审计+重构
- 审计报告：`docs/audit/control-flow-audit.md`
- 现状：if/foreach有部分实现，switch/loop/when完全缺失
- GAP：9个新字段中6个不存在，DependsOn需从`[]string`→`[]Dependency`
- 风险：Gate深嵌AgentExecutor(~300行)、RunContext"上帝对象"(35+字段)、DAG推进逻辑重复5+处
- 重构Phase 1-7，约10工作日，CC session `63052e78` 执行中

### 流程调试模式
- 修复：循环依赖检测bug（拓扑排序downstream图缺branch边）
- 修复：Input类型不匹配（map[string]any→string需json.Marshal）
- 调试器bug 11个→CC派发（引擎6个 session `bc09e3e2` + 前端5个 session `e4624e7d`）
- 竞品调研结论：ACP是行业首个IDE级工作流调试器

## 2026-02-28 更新

- **DSTE实例ed5d9903**：pending任务prompt含未解析`{{steps.xxx.output}}`模板变量 — 引擎创建任务时未resolve，是ACP引擎bug
- **cancelled/skipped颜色**：#faad14 (orange/yellow)，泽斌明确要求
- **代码审计结果**：6P0(concurrent map panic/defer close closed channel等)/8P1/13P2，详见AUDIT_REPORT.md，修复延后
- **Pipeline可视化**：≤20步用blocks，>20步用聚合进度条（hybrid方案），hover tooltip被否决
- **Race condition防护**：用defensive DB writes（`AND status != 'cancelled'`），不kill goroutine
- **YAML schema**：yaml.Unmarshal默认忽略未知字段，需注意字段名对齐
- **Event Log PRD**：已提交知识库审批，CC实现中

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

## 2026-02-27/28 重大变更

### Task统一模型重构
- StepExecution废弃→统一Task表（Camunda风格），延迟创建，3状态pending→running→completed
- 前端路由：/my-tasks/:taskId

### Agent实时状态
- diagnostics.enabled → lifecycle事件 → 4档状态(idle/processing/responding/offline)

### Bug修复
- wake RPC方法名是`wake`不是`cron.wake`；CronWake无法定向agent→用SessionsSend
- complete_step加前序依赖校验；取消加/stop强制中断

### DSTE实例
- `3ca28ee5`，已跑gap/industry/SWOT/plan，depends_on丢失需修复重跑

### Trigger系统
- P0已实现部署，API路由待注册到router

## 2026-03-01 更新

### Subagent Session改造（4 commits）
- Gateway `agent` RPC：有sessionKey→创建/路由session；只有agentId→路由到主session
- 正确格式：`agent:{id}:subagent:{uuid}`
- 修复：usage获取提前到task查询之前、stepRun.SessionID优先、human step prompt渲染

### Event Log已部署运行
- CC实现 commit `9e21c8d`：WorkflowEvent实体+emitEvent+8个事件插入点+API
- 验证OK：启动流程产生5条event（instance.created→task全生命周期→instance.completed）

### 审批节点重构
- commit `d30c786`（813行）：executeApprovalBranch + evaluateApprovalStrategy
- Task表新增parent_task_id + vote字段
- 策略引擎：any/all，agent自动spawn投票，on_reject/on_timeout策略

### Flow Editor
- P0精确渲染 commit `20e61bf`（1263行）— React Flow + dagre + 4种自定义节点
- P1可视化编辑 commit `b6f1d05` — 拖拽+属性面板+undo/redo
- P2 UX Polish进行中

### API全面鉴权改造
- commit `a89fd0e`：mine=true身份匹配、StepCallback身份校验、knowledge/workspace context取值、/me端点

### PLM剥离
- 移除PLM的workflow/task/approval模块（~25文件），PLM=纯领域数据服务
- 企业架构文档 知识库 `d25a1510`：三层权限设计

### 命名决策
- 泽斌：对外统一叫**ACP Tools**（不叫Plugin），避免与OpenClaw Plugin机制概念撞车

### 详情页设计方向
- Session Context是排查最有价值的信息（泽斌指出）
- 分层展示：事件Tab+对话Tab，tool call参数默认折叠
- Human step用不同模板：表单结果+上游内容+事件时间线

## 2026-03-06 更新

### YAML格式经验（容易踩坑）
- `input_fields` 不是合法 WorkflowDef 字段（不存在）
- shell exec 字段：`command` + `workdir`（不是 `cmd`）
- loop body步骤必须在顶层steps里声明，且需手动声明depends_on上游步骤
- **gate已废弃**：control=loop + condition + max_rounds + body=[step_ids]

### 重要bug根因记录（3/6）
- **控制节点拖入无容器**：getMetaByType缓存未就绪时返回undefined导致isControl=false，修法：用静态Set判断控制类型
- **属性面板硬编码**：控制类型字段应从meta.InputFields动态生成（不能写死JSX）
- **FlattenNestedYAML空数组崩溃**：空分支数组`[]`被continue跳过，保留[]interface{}，unmarshal到map[string]string崩溃。修法：空数组赋空字符串""

### Script Executor架构（已实装 e6254af）
- executor-scripts/ 目录，fsnotify热加载
- 协议：`node xxx.js --describe` 输出ScriptMeta；stdin JSON→stdout NDJSON事件流
- cc.ts已加载：`loaded script executor: cc` ✅（commit 457f21b）

### CC Executor关键环境变量
- `CLAUDE_CODE_MAX_OUTPUT_TOKENS=128000`（32000太小导致CC截断失败，是3/5三次CC失败根因）

### 企业开发工作流（prd-to-production 6bb955d2）
- fetch_prd → loop(implement→review, max_rounds=3) → build_test → deploy
- TS作为流程定义DSL：泽斌认可，待实现（内部仍是WorkflowDef IR）

### Bob工作流设计Agent
- workspace: /home/claw/.openclaw/workspace-Bob/
- 风格：引导式（先问清楚需求再生成YAML），直接用ACP Tools创建发布
- SCHEMA.md包含所有executor类型+命令+字段+控制节点+典型模板

## 2026-03-07 更新

### ACP Schema系统升级
- `/api/processes/schema` 现返回 workflow_schema + executors（10个executor完整schema）
- 删除21条 legacy 路由（/workflows/* 8条、/workflow-runs/* 11条）
- v1/v2 executor统一通过 Commands() + InputSchema() 生成schema（不再依赖SchemaProvider接口）
- Human executor：3个per-command struct（form/approval/notification），assignee=string（飞书open_id）
- 验证系统schema-driven重构：删140+行hardcoded验证（commit 1066936）
- omitempty标注规则：有omitempty→optional，jsonschema:"required"→required
- **ACP API key (main)**：`ak_80cc1217c8f65413a69f2a6fa6328dbf580333c95768e58930e52306455dcdca`
- 知识库「ACP Executor完整Schema参考」ID：2d99a1f3（注：文档ID可能已失效，以schema_full.json为准）
- schema_full.json 保存路径：agent-control-panel/schema_full.json（66KB）

### 核心PRD文档索引
| PRD | 知识库ID | 状态 |
|-----|----------|------|
| Executor函数调用模型 v2.0 | dd745fc6 | 完成 |
| Control Flow v2.0 | 7aa5fcba | Phase 1-7已实现 |
| 流程调试模式 | 4d229435 | 完成 |
| 审计数据架构 | 8992dfa0 | Phase 1-2已实现 |
| Script Executor PRD | c2eadd84 | 已实现 |
| CC Executor PRD v3.1 | 31f7128b | cc.ts已部署 |
| Knowledge Executor PRD | da9ac180 | 已实现 |

### React Flow 流程图编辑器技术要点（血的教训）
- **两套FlowGraph组件**：flow-editor/（废弃）vs flow/FlowGraph/（实际使用），WorkflowDetail用的是新版
- **PropertyPanel.tsx已删除**，统一使用 StepDetail/StepDetailEdit.tsx
- **`<ReactFlow fitView>` prop**：声明即启用，节点每次变化都fit，是viewport乱跳根本原因
- **Handle偏移根因**：ExecutorMeta异步加载→节点高度变化→React Flow未重测Handle位置
  - 修复：PlaceholderNode（和DynamicNode同尺寸）在metasReady=false时注册，保持Handle位置一致
  - 绝不在DynamicNode.tsx加新import（Vite TDZ循环依赖）
  - 绝不用 setNodes(prev=>[...prev])（触发无限循环+黑屏）
  - 绝不条件渲染FlowGraph（会导致ACP全局崩溃）
- **Handle overflow**：不改容器根div的overflow为visible；用CSS::before伪元素扩大感应区（不碰布局）
- **layoutContainerOnly**：onDrop只重排容器内部（不全局layout），保持其他节点位置不变
- **CC性能优化陷阱**：删掉useState改ref时，需验证所有依赖该state的UI效果（3/7高亮丢失教训）
- **Monaco Editor lazy mount**：tab切换才渲染，onMount异步，不能在tab切换useEffect里直接读editorRef
- **并行节点YAML顺序**：yamlIndex字段 + 同y-rank按yamlIndex分配x坐标，实现deterministic布局

### ACP执行模型（当前架构）
- goroutine per step 模式（不是全局异步）
- human：RegisterCallback → channel阻塞等 → POST callback → goroutine解除阻塞
- CC/Script：exec.Cmd + goroutine阻塞读stdout，等进程退出
- 全局异步（Temporal模式）改动大，需完整PRD，泽斌未决定是否重构

### 属性面板通用化（SchemaFieldEditor.tsx d70ef2a）
- 根据executor input schema自动递归渲染UI（不再为每个executor定制）
- string+enum→Select / string→Input|TextArea / number→InputNumber / boolean→Switch
- object(with properties)→展开子字段 / array of primitive→tags / array of object→可增删卡片
- 泽斌原则：不该改后端兼容错误的写法，让schema正确声明格式，用户按格式填

## 2026-03-10 更新

### 引擎重构完成（8770→3945行，-55%）
- 引擎重构全流程验证通过（PPP流程+流程需求收集双双ok），里程碑commit: 812c5ce
- 审计发现8个bug全修复（C1 foreach并发/C2 terminate loop/W1 resumeRun嵌套/W3 dagMutex泄漏等）
- loop do-while修复：resumeLoopBody缺__loop stepOutput导致{{loop.iteration}}不渲染 (a89f99a)
- findBodyStepIndex两趟扫描，direct match优先（29c9dc5）

### SQLite→PG迁移
- ACP完全迁移到PostgreSQL，30+表AutoMigrate创建
- credentials表迁移：plm bearer + anthropic-api-key两条记录

### LLM Executor
- 文件：executor/llm.go，反射引擎模式，Generate命令
- Input: prompt/system/model/max_tokens/temperature/provider
- Output: text/model/input_tokens/output_tokens/duration_ms/stop_reason
- 默认model: claude-sonnet-4-6，从credentials store读anthropic-api-key (slug: anthropic-api-key)
- 对比：2.3秒完成 vs agent的30秒+，commit: 755c4c0

### flow-reference API
- GET /api/flow-reference（无需认证），~900 tokens，比schema省84%
- 动态从executor registry生成；schema类型*jsonschema.Schema需json.Marshal/Unmarshal转换
- v2加静态YAML示例；字段去重（agent/cc三command input相同→合并一行），commit: 31c10c4

### 架构决策（context注入）
- Context注入=流程编排职责：前置步骤(http/knowledge)拉数据→变量注入prompt
- 不要把context拉取逻辑塞进executor（业务耦合）
- wake-on-remaining移除：不再发"还有N条任务"消息，纯靠agent executor prompt自驱（5dcee71）

### 各种bug修复
- acp_list_processes性能优化：List()加.Omit("yaml_content")，145KB→8KB (a7d6514)
- Human executor commands为空：registry.Types()改为用e.Type()返回canonical名 (34f0022)
- 验证器upstream依赖：flattenNestedSteps给分支入口步骤加addImplicitDep(parentID) (75ed2b2)
- array input模板渲染：新增renderInputRecursive递归渲染 string/slice/map (engine_template.go)
- 知识库审批bypass：CheckPermission里approval状态→allow (6ec0925)
- system prompt最小化：只保留"实例ID+步骤ID+调用acp_complete_step"，移除HTTP callback残留 (fcd7a0c)

### Bob配置修复
- openclaw.json中Bob/cloine的model需anthropic/前缀："anthropic/claude-opus-4-6"
- API key换了→必须同步所有agent auth-profiles.json（用python3批量同步main→其他agent）
- TOOLS.md重写14KB→2.7KB：移除静态executor字段，改为动态查acp_get_executor_schema
- AGENTS.md更新：25个有效domain列表（禁止software/acp），executor output路径对照表

### Output路径对照表（引用规范）
- human executor: `output.form.fieldName`
- agent executor: `output.structured_output.fieldName`
- knowledge executor: `output.contents`, `output.total_chars`（不是output.data.xxx！）
- nocodb executor: 参见executor schema
- llm executor: `output.text`

## 2026-03-11 更新

### Debug Mock功能
- getMockDataForStep：基于step的output_schema自动生成mock数据（boolean→true, number→1）
- human步骤mock: `{"form": {"field_key": "mock_value"}}`
- agent步骤mock: 必须包AgentStandardOutput信封 `{"executor":"agent","structured_output":{...}}`（裸schema数据不行，模板引用找不到路径）
- loop body mock不生效根因：executeBodyStep没有调用debugBeforeHook → 修复加入 (46ef434)
- Commits: 969bbd3（基础）/ 46ef434（loop body）/ 12a75ea（agent信封）

### Mock待实现（PRD 4d229435 v6）
- loop mock模式限1轮（evalLoopCondition加MockMode检查）
- 断点静态ID suffix匹配（shouldPauseBefore/After加strings.HasSuffix）
- 断点优先于mock（debugBeforeHook里断点检查移到mock检查之前）
- inspect面板显示mock_preview字段

### Loop Body独立执行路径
- engine_control_body.go的executeBodyStep是loop body步骤专用路径
- 任何引擎新功能（hook/mock/debug）都要在此路径单独添加，不能只改主路径
