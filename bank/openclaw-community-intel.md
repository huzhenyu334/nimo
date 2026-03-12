# OpenClaw 社区情报（2026-03-09 整理）

> 基于官方文档 docs.openclaw.ai、Showcase、ClawHub 等渠道整理

---

## 一、OpenClaw 是什么

OpenClaw 是一个**自托管 AI Agent 平台**，核心理念是"AI 助手住在你的机器上"。它通过一个 Gateway 进程统一管理：
- **多渠道接入**：WhatsApp、Telegram、Discord、Slack、Signal、iMessage、Mattermost、飞书等
- **多 Agent 路由**：每个 Agent 有独立 workspace/session/persona
- **工具执行**：文件读写、shell 命令、浏览器控制、媒体处理
- **移动节点**：iOS/Android 设备配对，支持摄像头、屏幕录制、位置、语音等

官网: https://docs.openclaw.ai | Discord: https://discord.gg/clawd | Twitter: @openclaw

---

## 二、核心架构与概念

### 2.1 Agent 工作区（Workspace）
每个 Agent 拥有独立工作区（默认 `~/.openclaw/workspace`），包含：
- **AGENTS.md** — 行为规则
- **SOUL.md** — 人格/风格定义
- **USER.md** — 用户信息
- **IDENTITY.md** — Agent 身份
- **TOOLS.md** — 工具配置笔记
- **HEARTBEAT.md** — 心跳任务清单
- **memory/** — 每日日志（memory/YYYY-MM-DD.md）
- **skills/** — 技能包

### 2.2 记忆系统（Memory）
- **纯 Markdown 文件**作为记忆源：`MEMORY.md`（长期记忆）+ `memory/*.md`（日志）
- **向量搜索**：自动对记忆文件建索引，支持语义查询
- 嵌入提供者：OpenAI / Gemini / Voyage / Mistral / Ollama / 本地 node-llama-cpp
- **QMD 后端（实验性）**：本地搜索 sidecar，BM25 + 向量 + 重排序
  - `memory.backend = "qmd"` 启用
  - 完全本地运行，无需外部 API
  - 支持会话 JSONL 索引
- **自动 Memory Flush**：Compaction 前自动触发一轮 agentic turn，提醒 Agent 持久化记忆
- **额外索引路径**：`memorySearch.extraPaths` 可索引工作区外的 Markdown 文件

### 2.3 Session 管理
- 直接对话 → 合并到 `main` session
- 群组聊天 → 独立 session
- **Compaction**：上下文过大时自动压缩，支持配置阈值
- **Session Pruning**：`contextPruning.mode: "cache-ttl"` 按缓存 TTL 修剪旧工具结果

### 2.4 Prompt Caching（省钱利器）
- `cacheRetention`: `none` | `short` | `long`
- 支持 Anthropic 直连、Bedrock、OpenRouter
- **心跳保温**：设置 heartbeat 间隔小于 cache TTL，保持缓存活跃
- **诊断工具**：`diagnostics.cacheTrace` 输出 JSONL 追踪缓存命中

---

## 三、最新玩法与技巧

### 3.1 多 Agent 路由（Multi-Agent）
一个 Gateway 运行多个 Agent，每个有独立人格、工作区、认证：

```bash
openclaw agents add coding    # 创建 coding agent
openclaw agents add social    # 创建 social agent
```

**常见模式：**
- **按渠道分流**：WhatsApp → 快速 Sonnet agent，Telegram → 深度 Opus agent
- **按联系人分流**：同一 WhatsApp 号码，不同联系人路由到不同 Agent
- **按 Discord 机器人分流**：每个 bot 对应一个 Agent
- **家庭 Agent**：绑定到特定 WhatsApp 群组

**路由优先级**：peer > parentPeer > guildId+roles > guildId > teamId > accountId > channel > fallback

### 3.2 Sub-Agents（子 Agent 并行）
- 后台启动子 Agent 执行任务，不阻塞主对话
- 完成后自动 announce 结果到主 session
- **嵌套子 Agent**（Orchestrator 模式）：`maxSpawnDepth: 2`
  - Main → Orchestrator Sub-Agent → Worker Sub-Sub-Agents
- **Thread 绑定**：Discord 支持线程持久化绑定
- 每个子 Agent 可配置不同模型（省钱 trick：子 Agent 用便宜模型）
- 自动归档：完成后 `archiveAfterMinutes`（默认 60 分钟）

```bash
/subagents spawn coding "重构认证模块" --model anthropic/claude-sonnet-4-5
```

### 3.3 Skills 技能系统
技能是 Agent 的能力扩展包，一个文件夹 + `SKILL.md`：

**三层加载**：
1. 工作区 skills/（最高优先）
2. ~/.openclaw/skills/（共享）
3. 内置 skills

**ClawHub 技能市场**（clawhub.ai）：
```bash
clawhub search "calendar"     # 搜索技能
clawhub install my-skill      # 安装技能
clawhub update --all           # 更新所有
clawhub sync --all             # 同步发布
```

**技能门控**：
- `requires.bins` — 依赖二进制
- `requires.env` — 依赖环境变量
- `requires.config` — 依赖配置项
- `os` — 限制操作系统

**跨节点技能**：Linux Gateway + macOS 节点，macOS-only 技能可通过节点远程执行。

### 3.4 浏览器控制
Agent 可控制隔离的浏览器实例：

**两种模式：**
- `openclaw` profile — 托管隔离浏览器（推荐）
- `chrome` profile — 通过扩展中继控制用户浏览器

**远程浏览器：**
- 支持 Browserless（HTTPS CDP）
- 支持 Browserbase（WebSocket CDP，自带验证码解决）
- 支持任意远程 CDP 端点

**沙箱浏览器**：Docker 中运行，noVNC 远程观察

### 3.5 Cron 定时任务
内置调度器，持久化存储：

```bash
# 每天早上 7 点总结
openclaw cron add --name "Morning brief" --cron "0 7 * * *" \
  --tz "Asia/Shanghai" --session isolated \
  --message "总结昨晚的更新" --announce --channel telegram
```

**两种执行模式**：
- **Main session**：注入系统事件到心跳
- **Isolated**：独立 agent turn，完成后 announce 结果

**支持模型覆盖**：Cron 任务可指定不同模型和 thinking 级别
**Webhook 投递**：结果可 POST 到外部 URL
**轻量上下文**：`lightContext: true` 跳过工作区文件注入

### 3.6 Hooks 事件钩子
事件驱动的自动化脚本：

**内置 Hooks**：
- `session-memory` — `/new` 时保存会话上下文
- `bootstrap-extra-files` — 注入额外启动文件
- `command-logger` — 命令审计日志
- `boot-md` — Gateway 启动时执行 BOOT.md

```bash
openclaw hooks enable session-memory
openclaw hooks list
```

**Hook Packs**：通过 npm 安装打包的 hook 集合

### 3.7 沙箱化执行（Sandboxing）
Docker 容器隔离工具执行：

**模式**：
- `off` — 无沙箱
- `non-main` — 仅非主 session 沙箱化
- `all` — 所有 session

**范围**：
- `session` — 每个 session 一个容器
- `agent` — 每个 agent 一个容器
- `shared` — 共享容器

**工作区访问**：`none`（隔离）/ `ro`（只读）/ `rw`（读写）
**自定义挂载**：`docker.binds` 挂载额外目录

---

## 四、社区 Showcase 精选

| 项目 | 作者 | 标签 | 说明 |
|------|------|------|------|
| **PR Review → Telegram** | @bangnokia | review, github | OpenCode 提 PR → OpenClaw 审代码 → Telegram 回复修改建议 |
| **Wine Cellar Skill** | @prades_maxime | skills, local | 让 Agent 创建本地酒窖管理技能，从 CSV 导入 962 瓶酒 |
| **Tesco 自动购物** | @marchattonhere | automation, browser | 周餐计划 → 自动下单 → 预约配送，纯浏览器控制无 API |
| **SNAG 截图转 Markdown** | @am-will | devtools | 热键截屏 → Gemini Vision → 剪贴板 Markdown |
| **Agents UI** | @kitze | ui, skills | 桌面应用管理多个 AI Agent 的 skills/commands |
| **TTS 语音笔记** | Community | voice, telegram | papla.media TTS → Telegram 语音消息 |
| **CodexMonitor** | @odrobnik | devtools | Homebrew CLI 监控 Codex sessions |
| **Bambu 3D 打印控制** | @tobiasbischoff | hardware | 通过 Agent 控制 BambuLab 3D 打印机 |

---

## 五、实用配置模板

### 5.1 省钱配置（Cache + 便宜子 Agent）
```json5
{
  agents: {
    defaults: {
      model: { primary: "anthropic/claude-opus-4-6" },
      models: {
        "anthropic/claude-opus-4-6": {
          params: { cacheRetention: "long" }
        }
      },
      heartbeat: { every: "55m" },  // 保持缓存活跃
      subagents: {
        model: "anthropic/claude-sonnet-4-5",  // 子 Agent 用便宜模型
        runTimeoutSeconds: 900,
      },
      contextPruning: { mode: "cache-ttl", ttl: "1h" },
    },
  },
}
```

### 5.2 多人共用 Gateway
```json5
{
  agents: {
    list: [
      { id: "alex", workspace: "~/.openclaw/workspace-alex" },
      { id: "mia", workspace: "~/.openclaw/workspace-mia" },
    ],
  },
  bindings: [
    { agentId: "alex", match: { channel: "whatsapp", peer: { kind: "direct", id: "+86138..." } } },
    { agentId: "mia", match: { channel: "whatsapp", peer: { kind: "direct", id: "+86139..." } } },
  ],
}
```

### 5.3 远程浏览器（Browserbase）
```json5
{
  browser: {
    enabled: true,
    defaultProfile: "browserbase",
    profiles: {
      browserbase: {
        cdpUrl: "wss://connect.browserbase.com?apiKey=YOUR_KEY",
      },
    },
  },
}
```

---

## 六、关键技巧总结

1. **记忆防丢失**：重要决策立即写 MEMORY.md，不要等 session 结束
2. **Compaction 前自动 flush**：配置 `memoryFlush.enabled: true`
3. **向量搜索**：确保配好嵌入 provider，否则记忆搜索退化为关键词
4. **QMD 本地搜索**：不想用远程 API？试试 `memory.backend = "qmd"`，完全本地
5. **子 Agent 省钱**：主 Agent 用 Opus，子 Agent 用 Sonnet
6. **Prompt Cache 保温**：心跳间隔 < cache TTL（如 55min heartbeat + long cache）
7. **Cron + 隔离 session**：定时任务用 isolated 模式，避免污染主对话历史
8. **技能门控**：用 `requires` 声明依赖，避免技能在不兼容环境加载
9. **沙箱安全**：生产环境建议 `sandbox.mode: "non-main"` + `workspaceAccess: "ro"`
10. **ClawHub 生态**：先 `clawhub search` 看看社区有没有现成技能，别重复造轮子

---

## 七、资源链接

- 官方文档: https://docs.openclaw.ai
- 文档完整索引: https://docs.openclaw.ai/llms.txt
- ClawHub 技能市场: https://clawhub.ai
- Discord 社区: https://discord.gg/clawd
- Twitter: https://x.com/openclaw
- Showcase: https://docs.openclaw.ai/start/showcase
- GitHub 讨论: https://github.com/nicobailon/openclaw
- NPM 包: https://www.npmjs.com/package/openclaw

---

*整理时间: 2026-03-09 | 整理者: Lyra*
