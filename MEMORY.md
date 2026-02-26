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
13. **ACP工具=OpenClaw Plugin**：通过`extensions/acp/index.ts`注册为Tools，Tools调用ACP后端API。派CC加ACP工具时改的是Plugin文件（index.ts），如果API不存在才需要同时改ACP后端加API（2026-02-26）

## 核心洞察

- **Agent框架无壁垒**，真正壁垒：①企业业务prompt ②给agent用的tool（2026-02-22）
- **ACP = Agent的操作系统**，Gateway是runtime，ACP是用户空间（2026-02-25）
- **不需要MCP**，原生Plugin最可靠，agent天生就会用（2026-02-25）

## 泽斌终极目标

ACP驱动PLM/SRM自主开发到产品化，Agent自我进化。详见知识库。

## 架构决策

- **v2.1 Step数据模型（2026-02-26最终定稿）**：form/output_schema/component三者平级
  - form = 输入表单（fields+on_submit），收集人的输入
  - output_schema = 输出约束+只读渲染，约束agent/引擎校验
  - component = 外部SDK组件（旧名form），props支持`{{steps.xxx.output.field}}`
  - structured_output/output_schema命名不变！不改名
  - on_submit：表单提交调API，成功才完成step，output_mapping提取字段合并到structured_output
- **废弃inline steps嵌套** → 改用`group:"xxx"`标签
- **统一assignee** → assignee_type(agent/user/role)+assignee_id
- **甘特图** → 所有步骤≥1h显示宽度

## 当前优先级

1. **🔴 ACP↔PLM集成** — Phase 1进行中（Plugin+表单嵌入+webhook回调）
2. **PLM产品化** — 代码审计完成，差距分析→路线图
3. **ACP持续演进** — 子流程v4+甘特图已上线，统一资源调度PRD待实现

## 进行中

- **EBOM嵌入模式** — CC已完成前端改造，待build部署+ACP侧集成
- **ACP↔PLM架构** — 纲领v1.1已更新，含任务集成+表单嵌入设计（详见知识库）

## 基础设施

- 知识管理三层：shared-bootstrap(宪法) → 知识库(制度) → AGENTS.md(个性化)
- embedding: 智谱embedding-3, memory-lancedb-pro已部署
- Catherine: 43.156.133.145, GLM5, Node已连接(需检查稳定性)

## 重要教训

- compaction后必须memory_search再回答
- 飞书open_id跨应用不通用（ACP ou_e229cd ≠ OpenClaw ou_5b159fc）
- CC僵尸进程占大量资源注意清理；前端build需2048MB
- SIGUSR1热重载不加载新plugin，需完全重启
