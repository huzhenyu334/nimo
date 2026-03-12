# MEMORY.md - Lyra 核心记忆

> 上次整理：2026-03-10 | 细节见 bank/acp.md

## 泽斌的铁律

1. **编程全走CC**：claude_agent SDK，修完必须commit
2. **Lyra=调度者**：不亲自写代码，派CC
3. **先对齐再执行**：不理解"做到什么程度"就不开工
4. **知识库铁律**：先搜后写，有则增量更新（append/section_write/patch）
5. **ACP工具=Plugin**：改extensions/acp/index.ts，API不存在才改后端
6. **所有事情走流程**：不做一次性操作，可复用流程优先
7. **Executor选择层级**：原子 > llm > agent > cc（能用简单的绝不用复杂的）
8. **派CC前确认需求范围**：已修好的bug不要继续扩散，不脑补需求
9. **API key换了必须同步所有agent**：把main的auth-profiles.json同步到所有agent（3/10）
10. **Context注入=流程编排**：前置步骤拉数据→变量注入（不是executor职责）
11. **复用→独立子流程**：可复用流程做成独立流程，不内联
12. **部署ACP前查running实例**：有运行中→等完成或确认

## ACP API Key
`ak_d9d31c3f27a92d796765c831f701c8d566994314faec3e1174c4459f0bdb07e1`

## 核心洞察
- **Executor口诀**：CRUD→nocodb/http；文本生成→llm；推理+工具→agent；改代码→cc
- **flow-reference**：GET /api/flow-reference，~880 tokens（比Schema省84%）
- **context决定输出质量80%**，冷启动+精准prompt注入

## 架构决策（最新）
- **LLM Executor已上线**（3/10, 755c4c0）：直接调Anthropic API，2.3秒完成 vs agent的30秒+；从credentials store读anthropic-api-key
- **SQLite→PG迁移完成**（3/10）：ACP用PostgreSQL，30+表
- **Output统一**：human→output.form.xxx，agent→output.structured_output.xxx；assignee=必填plain string
- **知识库审批bypass**（3/10, 6ec0925）：approval状态自动转allow
- **Session永不重置**：session.reset.mode="never"
- **Bob=流程设计专家**：workspace-Bob/，写YAML前必须acp_get_executor_schema确认字段

## PRD标准化（3/10）
双轨输出：JSON元数据（引擎校验+doc_id引用）+ Markdown全文（知识库）；structured_output字段加maxLength，总量<5KB

## 当前状态
1. **ACP引擎** 🎉 重构完成：8770→3945行(-55%)，全流程验证通过（3/10里程碑）
2. **流程体系**：需求收集(82866aeb)、立项(80f6194e)、代码扫描(9286f334)、文档创建(80d2f70d)、cc-dev-task(3fcff09c)
3. **NocoDB**：项目表(mk4cbfl27dbfhlu)+文档注册表(mjap0r8h9nxrb67)+ACP需求管理(m0e7hp5d77pczt9)
4. **LLM Executor**已上线，flow-reference API已上线
5. **Debug Mock已上线**（3/11，969bbd3/46ef434/12a75ea）：output_schema生成mock；agent mock包AgentStandardOutput信封；**待实现**：loop限1轮、断点suffix匹配、断点优先于mock、inspect显示mock_preview

## 重要教训
- **派CC前必须grep验证**：summary的"待做"不可信，功能可能已存在
- **go build正确命令**：`go build -o bin/acp ./cmd/acp/`（go build ./... 不更新指定binary）
- **Bob model需anthropic/前缀**：openclaw.json的model必须="anthropic/claude-opus-4-6"
- **Bob 100秒TTFT正常**：claude-opus-4-6+thinking模式，非故障
- 飞书open_id跨应用不通用（ACP ou_e229cd ≠ OpenClaw ou_5b159fc）
- 泽斌不喜欢§符号，引用章节用中文或数字编号
- **Agent mock必须包AgentStandardOutput信封**，否则structured_output路径找不到
- **Loop body=独立执行路径**（executeBodyStep），hook/mock必须在此单独加
