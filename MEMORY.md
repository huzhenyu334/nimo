# MEMORY.md - Lyra 核心记忆

> 上次整理：2026-02-25 (sleep consolidation) | 原则：只保留影响当前行为的信息，细节在bank/

## 泽斌的铁律

1. **工作流纪律**：开发任务走Lyra调度→PM拆任务→CC开发（2026-02-20）
2. **OpenClaw操作**：任何配置修改/重启必须先问泽斌（2026-02-17）
3. **部署第三方**：先读README再动手，优先环境变量不改源码（2026-02-19）
4. **编程铁律**：所有代码通过Claude Code，无例外（2026-02-11）
5. **流程纪律**：Lyra是调度者不是执行者（2026-02-20）
6. **PM第一守则**：先对齐再执行——不理解"做到什么程度"就不开工（2026-02-24）
7. **DB是可选扩展**：ACP核心功能不能依赖DB（2026-02-24）
8. **CC调用不阻塞**：用sessions_spawn调CC，不用同步claude_agent（2026-02-25）
9. **先想再做**：上下文缺失时 memory_recall → 知识库 → 主动询问（2026-02-25）
10. **subagent超时要长**：runTimeoutSeconds: 1800（2026-02-25）
11. **CC修完必须commit**（2026-02-25）
12. **subagent必须同步调CC**：task里写明"用claude_agent同步调用"（2026-02-25）

## 核心洞察

- **Agent框架无壁垒**，真正壁垒：①企业业务prompt ②给agent用的tool（2026-02-22）
- **ACP = Agent的操作系统**，Gateway是runtime，ACP是用户空间（2026-02-25）
- **不需要MCP**，原生Plugin最可靠，agent天生就会用（2026-02-25）

## 泽斌终极目标

ACP驱动PLM/SRM自主开发到产品化，Agent自我进化。详见知识库。

## 架构决策（2026-02-25新增）

- **废弃inline steps嵌套** → 改用`group:"xxx"`标签，纯前端分组，引擎不处理
- **统一assignee** → assignee_type(agent/user/role)+assignee_id，Step type只管结构
- **Duration** → YAML用友好格式(5d)，存储转ms(duration_ms+duration_raw)，统一自然日
- **甘特图** → 所有步骤≥1h显示宽度，时间尺度切换(月/周/天/时)

## 当前优先级

1. **🔴 PLM产品化** — 代码审计(cas-mlzgm31w) → 差距分析 → 路线图
2. **ACP持续演进** — 子流程v4+甘特图已上线，统一资源调度PRD待实现(4Phase)
3. **知识库体系** — 8篇文档pending审批

## 进行中

- **Skills RPC改造** — CC任务cas-mm0d2ga4
- **gatewayId前端改造** — CC任务cas-mm0dy170
- **ACP Node功能规划** — Phase 1待启动

## 基础设施

- 知识管理三层：shared-bootstrap(宪法) → 知识库(制度) → AGENTS.md(个性化)
- embedding: 智谱embedding-3, memory-lancedb-pro已部署
- Catherine: 43.156.133.145, GLM5, Node已连接(需检查稳定性)

## 重要教训

- compaction后必须memory_search再回答
- 飞书open_id跨应用不通用（ACP ou_e229cd ≠ OpenClaw ou_5b159fc）
- CC僵尸进程占大量资源注意清理；前端build需2048MB
- SIGUSR1热重载不加载新plugin，需完全重启
