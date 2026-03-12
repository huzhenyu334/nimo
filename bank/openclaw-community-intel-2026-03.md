# OpenClaw 社区情报汇总（2026年3月）

> 整理时间：2026-03-09 | 来源：官方文档、官网、Twitter社区

---

## 一、OpenClaw 是什么

OpenClaw 是一个**自托管的 AI Agent 网关**，把 WhatsApp、Telegram、Discord、iMessage、飞书、Slack 等聊天应用连接到 AI 编码代理（Pi）。一个 Gateway 进程搞定所有渠道，运行在你自己的机器上。

**核心卖点：**
- 自托管、数据自控
- 多渠道统一（一个Gateway服务所有聊天平台）
- Agent原生：工具使用、Session管理、记忆、多Agent路由
- 开源MIT协议、社区驱动
- 跨平台：macOS/Linux/Windows(WSL)/Raspberry Pi/Docker/各云平台

---

## 二、核心架构与功能

### 2.1 多Agent路由（Multi-Agent Routing）
- 每个Agent有独立的**workspace**（文件、AGENTS.md、SOUL.md）、**状态目录**、**Session存储**
- 支持一台Gateway同时托管多个Agent
- 路由规则：按peer（DM/群）> guildId > accountId > channel匹配，最具体的优先
- **实战玩法：**
  - 一个WhatsApp号码，按发送者分配不同Agent（DM split）
  - WhatsApp用快速模型日常聊天，Telegram用Opus深度工作
  - 家庭Agent绑定到特定WhatsApp群
  - 每个Discord/Telegram bot对应不同Agent

### 2.2 记忆系统（Memory）
- **纯Markdown文件**作为记忆源：`memory/YYYY-MM-DD.md`（日志）+ `MEMORY.md`（长期）
- 向量记忆搜索：基于嵌入的语义搜索
  - 支持OpenAI/Gemini/Voyage/Mistral/Ollama作为嵌入提供者
  - 可选QMD后端（BM25+向量+重排序）
- **自动记忆保存**：Session接近compaction时自动触发记忆写入
- 关键配置：`agents.defaults.compaction.memoryFlush`

### 2.3 Cron调度系统
- Gateway内置调度器，支持一次性和重复任务
- 两种执行模式：
  - **Main session**：在主session心跳中执行
  - **Isolated**：独立session执行，不污染主会话
- 支持模型和thinking级别覆盖
- 支持结果投递到聊天渠道或Webhook
- 时区感知，支持stagger防止峰值

### 2.4 Heartbeat心跳
- 周期性Agent唤醒（默认30分钟）
- 通过`HEARTBEAT.md`自定义检查清单
- 支持活跃时段限制（`activeHours`）
- 支持轻量上下文模式（`lightContext`）减少token消耗
- 响应协议：无事用`HEARTBEAT_OK`，有事直接输出

### 2.5 Webhook集成
- 暴露HTTP端点接收外部触发
- `/hooks/wake`：唤醒主session
- `/hooks/agent`：触发独立Agent任务
- 支持自定义hook映射、JS/TS转换模块
- Gmail PubSub内置支持

### 2.6 工具系统（Tools）
- 内置工具：exec、process、browser、canvas、nodes、cron、web_search、web_fetch等
- **工具策略**：可全局allow/deny，支持tool profiles（minimal/coding/messaging/full）
- **按Provider限制工具**：不同模型可以有不同工具集
- **工具组简写**：`group:runtime`, `group:fs`, `group:web`, `group:ui`等
- **循环检测**：防止Agent陷入工具调用死循环

---

## 三、Skills生态

### 3.1 Skill系统
- 每个Skill是一个包含`SKILL.md`的目录
- 三级加载：workspace skills > managed skills > bundled skills
- 支持按环境/二进制/配置门控（gating）
- 热重载：文件变更自动刷新
- 多Agent场景下支持per-agent和shared skills

### 3.2 ClawHub（技能市场）
- 地址：https://clawhub.ai
- 公开技能注册表，搜索基于嵌入（向量搜索）
- CLI工具：`npm i -g clawhub`
- 核心操作：
  ```bash
  clawhub search "关键词"     # 搜索
  clawhub install <slug>      # 安装
  clawhub update --all         # 更新所有
  clawhub sync --all           # 扫描并发布
  clawhub publish <path>       # 发布
  ```
- 支持版本控制（semver）、star、评论、举报机制

### 3.3 社区插件
- WeChat插件：`@icesword760/openclaw-wechat`（微信个人号连接）
- 安装方式：`openclaw plugins install <npm-spec>`
- 插件可注册额外工具和CLI命令

---

## 四、支持的渠道（超全）

WhatsApp | Telegram | Discord | Slack | iMessage | Signal | 飞书(Feishu) | Google Chat | Microsoft Teams | Mattermost | IRC | LINE | Matrix | Nextcloud Talk | Nostr | Twitch | Zalo | Synology Chat | Tlon | BlueBubbles | WeChat(插件)

---

## 五、支持的模型Provider

Anthropic | OpenAI | Amazon Bedrock | GitHub Copilot | Mistral | Ollama(本地) | OpenRouter | Qwen | Moonshot | MiniMax | NVIDIA | vLLM | Venice AI | Vercel AI Gateway | Cloudflare AI Gateway | Hugging Face | GLM | Qianfan | Z.AI | 小米MiMo | Claude Max API Proxy | LiteLLM | Together | Deepgram(语音)

---

## 六、社区实战玩法精选（Twitter收集）

### 🔥 高频用法
1. **从手机控制电脑开发** — Telegram/WhatsApp发消息触发Claude Code，边遛狗边写代码
2. **自建AI助手生态** — 给agent起名字、设个性、分角色（工作/生活/编码）
3. **智能家居控制** — 连接Philips Hue灯、空气净化器等IoT设备
4. **Email/日历/航班** — 自动处理邮箱、管理日历、航班值机
5. **自我进化** — Agent自己发现需要API key → 自动打开浏览器 → 配置OAuth
6. **个性化冥想** — 生成定制冥想文本 + TTS + 环境音
7. **多实例克隆** — 一个Agent搞定配置后自我克隆到多台机器

### 💡 高级技巧
1. **Skill自创建** — 跟Agent说需要什么能力，它自己创建skill并开始使用
2. **YouTube → Skills** — 喂YouTube视频让Agent提炼为可复用的workflow skill
3. **Sentry → Auto-fix** — webhook接入Sentry，自动捕获错误并开PR修复
4. **Claude Max代理** — 用Claude Max订阅作为API endpoint节省成本
5. **GitHub Copilot代理** — 通过代理把Copilot订阅作为模型endpoint
6. **跨渠道信息关联** — Agent自动关联不同聊天渠道的信息形成洞察
7. **保险理赔** — Agent"意外"帮用户跟保险公司打架（还赢了）

### 🏗️ 部署实践
1. **Raspberry Pi** — 低成本7x24运行，配合Cloudflare
2. **Mac Mini** — 桌面级AI助手，有键盘鼠标浏览器完整控制
3. **Docker** — 容器化部署到各种云
4. **多Agent协作** — 不同Agent负责不同领域，共享一台Gateway

---

## 七、关键社区语录

> "Using OpenClaw for a week and it genuinely feels like early AGI." — @tobi_bsf

> "It's running my company." — @therno

> "The fact that it's hackable (and more importantly, self-hackable) and hostable on-prem will make sure tech like this DOMINATES conventional SaaS." — @rovensky

> "After years of AI hype, I thought nothing could faze me. Then I installed OpenClaw." — @lycfyi

> "At this point I don't even know what to call OpenClaw. It is something new." — @davemorin

> "My OpenClaw realised it needed an API key… it opened my browser… opened the Google Cloud Console… Configured oauth and provisioned a new token" — @Infoxicador

---

## 八、对BitFantasy的参考价值

### 可以借鉴的
1. **飞书渠道已支持** — 我们已在用，可以深化集成
2. **Multi-Agent架构** — ACP的多Agent流程可以参考OpenClaw的路由和隔离模式
3. **Skill生态** — ClawHub的skill发布/安装模式可以给ACP的模板系统做参考
4. **Webhook集成** — PLM/ERP的事件可以通过webhook触发Agent任务
5. **记忆系统** — QMD后端的BM25+向量+重排序方案值得关注
6. **Cron调度** — 可参考其isolated job模式优化我们的定时任务

### 值得关注的趋势
1. **Agent自进化** — 让Agent自己创建/修改skill是社区最热的方向
2. **IoT集成** — 从软件走向硬件控制（我们做智能眼镜更有优势）
3. **跨平台统一** — 一个Agent入口覆盖所有沟通渠道
4. **开源社区驱动** — MIT协议+ClawHub生态快速聚拢开发者

---

## 九、相关链接

- 官网：https://openclaw.ai
- 文档：https://docs.openclaw.ai
- Skills市场：https://clawhub.ai
- 社区：Discord（官方社区）
- 安装：`npm install -g openclaw@latest`
- CLI文档索引：https://docs.openclaw.ai/llms.txt
