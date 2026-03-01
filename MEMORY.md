# MEMORY.md - Lyra 核心记忆

> 上次整理：2026-03-01 | 原则：只保留影响当前行为的信息，细节在bank/

## 泽斌的铁律

1. **编程全走CC**：必须通过sessions_spawn同步调用，禁止async+轮询（浪费context加速compaction）
2. **Lyra=调度者**：开发走Lyra→PM→CC，不亲自写代码
3. **CC调用前后必须写MEMORY.md**：调用前写"最近CC任务"（任务+状态running），完成后更新结果——保证compaction后永远记得最后CC干了什么
4. **先对齐再执行**：不理解"做到什么程度"就不开工
5. **先想再做**：上下文缺失→memory_recall→知识库→问泽斌
6. **知识库铁律**：先搜后写，有则增量更新（append/section_write/patch）
7. **ACP工具=Plugin**：改extensions/acp/index.ts，API不存在才改后端
8. **DB是可选扩展**：ACP核心不能依赖DB
9. **OpenClaw操作**：配置修改/重启先问泽斌
10. **部署第三方**：先读README，优先环境变量不改源码
11. **一次性给最优方案**：先调研行业做法再设计，不要先给次方案再改（2026-03-01）

## 核心洞察

- **Agent框架无壁垒**，真正壁垒：①企业业务prompt ②给agent用的tool
- **ACP = Agent的操作系统**，Gateway是runtime，ACP是用户空间
- **不需要MCP**，原生Plugin最可靠，agent天生就会用
- **企业即求解器，一切皆计算**：流程=分治算法，知识=函数不是信息（详见知识库067dbaea）
- **评估核心指标=人的修改率**：最小可行评估=token数+耗时+成功率+修改率
- **context决定输出质量80%**，模板只占20%且价值递减
- **历史消息是噪音不是记忆**：每个task冷启动+精准prompt注入（2026-03-01）

## 架构决策

- **Task统一模型**：Step=YAML模板，Task=运行时单元（Camunda风格）
- **7状态生命周期**：waiting→pending→running→completed/failed/cancelled/skipped（2026-02-28）
- **动态指派方案C**：统一assignee结构体（type=agent|human + 寻址方式id/role/candidates/rules/relation），废弃旧assignee_type+assignee_id（PRD知识库00e7a2b9）
- **三层分离**：ACP Tools(agent) / Backend API(数据层) / Web UI(人)
- **Push+Pull双模式**：assignee=Push直接分配，candidates=Pull竞争claim
- **Event Log单表设计**：一切皆事件，13种事件类型，workflow_events表（2026-03-01）
- **每个task用sessions_spawn**：不用agent session中间层（2026-03-01）
- **Task表仅加session_id**：审计数据全从event log查，trace内容不存ACP（2026-03-01）
- **变量引用验证**：publish时BFS校验模板变量引用在上游可达集内
- **知识库=纯存储层**，审批走ACP流程

## 当前状态

1. **ACP引擎** — 7状态+统一Task已部署，Event Log CC实现中，待做：任务调度PRD（候选池/级联通知/SLA）
2. **ACP前端** — 卡片redesign进行中（Vercel风格+Pipeline可视化），审计bug修复27项已部署
3. **Agent进化体系** — 知识库eda4fb31 v2（搁置中）

## 重要教训

- compaction后必须memory_search再回答
- **派CC前必须grep验证功能是否已存在**——summary的"待做"不可信（2026-03-01 event log事件）
- 飞书open_id跨应用不通用（ACP ou_e229cd ≠ OpenClaw ou_5b159fc）
- SIGUSR1热重载不加载新plugin，需完全重启
- CC commit经常漏文件（git add不完整），需检查
- CronWake无法定向agent，用SessionsSend传具体内容
- complete_step后必须主动调acp_list_tasks，不能只看返回值
- 引擎wake：去掉agent-busy检查，始终wake（commit 030b7fe）

## 最近CC任务
- **任务**: ACP审批节点重构——激活审批引擎，支持Agent+Human混合审批
- **PRD**: 知识库 f28a5a61（审批节点重构PRD）
- **时间**: 2026-03-01 18:25 启动
- **状态**: 🔄 running
- **上一个CC**: 结构化日志 ✅ 完成（238处log全部迁移slog）
