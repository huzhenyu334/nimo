# OpenClaw Gateway WebSocket RPC 完整方法参考 — 参数与返回格式

> 从源码 `gateway-cli-DbznSfRg.js` (openclaw@2026.2.14) 逆向分析
> 生成日期: 2026-02-27

---

## 目录

1. [cronHandlers](#1-cronhandlers) — wake, cron.*
2. [sendHandlers](#2-sendhandlers) — send, poll
3. [chatHandlers](#3-chathandlers) — chat.*
4. [sessionsHandlers](#4-sessionshandlers) — sessions.*
5. [agentHandlers](#5-agenthandlers) — agent, agent.*
6. [agentsHandlers](#6-agentshandlers) — agents.*
7. [configHandlers](#7-confighandlers) — config.*
8. [modelsHandlers](#8-modelshandlers) — models.*
9. [nodeHandlers](#9-nodehandlers) — node.*
10. [healthHandlers](#10-healthhandlers) — health, status
11. [channelsHandlers](#11-channelshandlers) — channels.*
12. [connectHandlers](#12-connecthandlers) — connect
13. [logsHandlers](#13-logshandlers) — logs.*
14. [deviceHandlers](#14-devicehandlers) — device.*
15. [execApprovalsHandlers](#15-execapprovalshandlers) — exec.approvals.*
16. [skillsHandlers](#16-skillshandlers) — skills.*
17. [systemHandlers](#17-systemhandlers) — last-heartbeat, set-heartbeats, system-*
18. [talkHandlers](#18-talkhandlers) — talk.*
19. [ttsHandlers](#19-ttshandlers) — tts.*
20. [updateHandlers](#20-updatehandlers) — update.*
21. [usageHandlers](#21-usagehandlers) — usage.*, sessions.usage
22. [voicewakeHandlers](#22-voicewakehandlers) — voicewake.*
23. [webHandlers](#23-webhandlers) — web.login.*
24. [wizardHandlers](#24-wizardhandlers) — wizard.*
25. [browserHandlers](#25-browserhandlers) — browser.*

---

## 通用说明

**请求帧格式** (handleGatewayRequest):
```json
{
  "method": "方法名",
  "params": { ... },
  "id": "request-id"
}
```

**每个handler收到的上下文对象**:
```
{ req, params, client, isWebchatConnect, respond, context }
```

**respond 签名**: `respond(ok: boolean, payload?, error?, meta?)`

**错误格式**:
```json
{
  "code": -32600,
  "message": "错误信息",
  "data": { "details": ... }
}
```

**权限级别** (authorizeGatewayMethod):
- `operator.read` — 读取方法
- `operator.write` — 写入方法
- `operator.admin` — 管理方法 (config.*, wizard.*, update.*, agents.create/update/delete, cron.add/update/remove/run, sessions.patch/reset/delete/compact)
- `operator.pairing` — 配对方法
- `operator.approvals` — 执行审批方法

---

## 1. cronHandlers

### wake
**权限**: `operator.write`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| mode | string | ❌ | 唤醒模式 |
| text | string | ❌ | 唤醒文本 |

**返回：**
```json
// context.cron.wake() 的返回值，具体结构取决于 cron 服务实现
{ ... }
```

---

### cron.list
**权限**: `operator.read`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| includeDisabled | boolean | ❌ | 是否包含已禁用的任务 |

**返回：**
```json
{
  "jobs": [
    {
      "id": "job-uuid",
      "schedule": "*/5 * * * *",
      "message": "...",
      "enabled": true,
      "lastRunAt": 1709000000000,
      "nextRunAt": 1709000300000
      // ... 更多 cron job 字段
    }
  ]
}
```

---

### cron.status
**权限**: `operator.read`

**参数：** 无 (空对象或无 params)

**返回：**
```json
// context.cron.status() 返回值
{ ... }
```

---

### cron.add
**权限**: `operator.admin`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| schedule | string | ✅ | cron 表达式或时间戳 |
| message | string | ✅ | 定时任务要发送的消息 |
| sessionKey | string | ❌ | 目标 session key |
| agentId | string | ❌ | 目标 agent ID |
| enabled | boolean | ❌ | 是否启用 (默认 true) |
| label | string | ❌ | 任务标签 |
| channel | string | ❌ | 投递渠道 |
| to | string | ❌ | 投递目标 |
| deliver | boolean | ❌ | 是否投递到外部渠道 |

> 注: 输入会经过 `normalizeCronJobCreate()` 标准化

**返回：**
```json
{
  "id": "job-uuid",
  "schedule": "...",
  "message": "...",
  "enabled": true
  // ... cron job 完整对象
}
```

---

### cron.update
**权限**: `operator.admin`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id / jobId | string | ✅ | 任务 ID |
| patch | object | ✅ | 要修改的字段 |
| patch.schedule | string | ❌ | 新 cron 表达式 |
| patch.message | string | ❌ | 新消息内容 |
| patch.enabled | boolean | ❌ | 启用/禁用 |
| patch.label | string | ❌ | 新标签 |

> 注: `patch` 会经过 `normalizeCronJobPatch()` 标准化

**返回：**
```json
// context.cron.update() 返回值 — 更新后的 job 对象
{ ... }
```

---

### cron.remove
**权限**: `operator.admin`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id / jobId | string | ✅ | 要删除的任务 ID |

**返回：**
```json
// context.cron.remove() 返回值
{ "ok": true }
```

---

### cron.run
**权限**: `operator.admin`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id / jobId | string | ✅ | 要执行的任务 ID |
| mode | string | ❌ | 执行模式，默认 "force" |

**返回：**
```json
// context.cron.run() 返回值
{ ... }
```

---

### cron.runs
**权限**: `operator.read`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id / jobId | string | ✅ | 任务 ID |
| limit | number | ❌ | 返回的日志条数上限 |

**返回：**
```json
{
  "entries": [
    // cron run log entries
  ]
}
```

---

## 2. sendHandlers

### send
**权限**: `operator.write`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| idempotencyKey | string | ✅ | 幂等键，用于去重 |
| to | string | ✅ | 发送目标（手机号/群ID等，取决于渠道） |
| message | string | ❌ | 消息文本 (与 mediaUrl 至少一个) |
| channel | string | ❌ | 渠道 ID (如 "whatsapp", "telegram")，默认 DEFAULT_CHAT_CHANNEL |
| accountId | string | ❌ | 渠道账号 ID |
| mediaUrl | string | ❌ | 单个媒体 URL |
| mediaUrls | string[] | ❌ | 多个媒体 URL |
| sessionKey | string | ❌ | 关联的 session key（用于镜像到 transcript） |
| gifPlayback | any | ❌ | GIF 播放选项 |

**返回（成功）：**
```json
{
  "runId": "idempotency-key",
  "messageId": "平台消息ID",
  "channel": "whatsapp",
  "chatId": "...",         // 如有
  "channelId": "...",      // 如有
  "toJid": "...",          // 如有 (WhatsApp)
  "conversationId": "..."  // 如有
}
```

---

### poll
**权限**: `operator.write`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| idempotencyKey | string | ✅ | 幂等键 |
| to | string | ✅ | 发送目标 |
| question | string | ✅ | 投票问题 |
| options | string[] | ✅ | 投票选项 |
| channel | string | ❌ | 渠道 (默认 DEFAULT_CHAT_CHANNEL) |
| accountId | string | ❌ | 渠道账号 ID |
| threadId | string | ❌ | 线程 ID |
| maxSelections | number | ❌ | 最大选择数 |
| durationSeconds | number | ❌ | 投票持续秒数 (仅 Telegram) |
| durationHours | number | ❌ | 投票持续小时数 |
| silent | boolean | ❌ | 静默发送 |
| isAnonymous | boolean | ❌ | 匿名投票 (仅 Telegram) |

**返回：**
```json
{
  "runId": "idempotency-key",
  "messageId": "...",
  "channel": "telegram",
  "toJid": "...",
  "channelId": "...",
  "conversationId": "...",
  "pollId": "..."
}
```

---

## 3. chatHandlers

### chat.history
**权限**: `operator.read`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| sessionKey | string | ✅ | Session key |
| limit | number | ❌ | 消息数量限制 (默认 200, 最大 1000) |

**返回：**
```json
{
  "sessionKey": "global",
  "sessionId": "uuid",
  "messages": [
    {
      "role": "user" | "assistant",
      "content": [{ "type": "text", "text": "..." }],
      "timestamp": 1709000000000,
      "stopReason": "stop",
      "usage": {
        "input": 100,
        "output": 200,
        "cacheRead": 0,
        "cacheWrite": 0,
        "totalTokens": 300
      }
    }
  ],
  "thinkingLevel": "low" | "medium" | "high" | "xhigh" | null,
  "verboseLevel": "..." | null
}
```

---

### chat.send
**权限**: `operator.write`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| sessionKey | string | ✅ | Session key |
| message | string | ✅ | 消息文本 (或附件) |
| idempotencyKey | string | ✅ | 幂等键 (用作 clientRunId) |
| thinking | string | ❌ | 思考级别指令 (如 "high") |
| timeoutMs | number | ❌ | 超时毫秒数 |
| attachments | Attachment[] | ❌ | 附件列表 |
| attachments[].type | string | ❌ | 附件类型 |
| attachments[].mimeType | string | ❌ | MIME 类型 |
| attachments[].fileName | string | ❌ | 文件名 |
| attachments[].content | string/Buffer | ✅ | 内容 (base64 或字符串) |

**返回（立即响应 — 异步执行）：**
```json
{
  "runId": "client-run-id",
  "status": "started"
}
```

> **注意**: chat.send 是异步的。发送后立即返回 `started`。实际AI响应通过 WebSocket `chat` 广播事件推送：
>
> 广播事件格式:
> ```json
> {
>   "runId": "...",
>   "sessionKey": "global",
>   "seq": 1,
>   "state": "final" | "error",
>   "message": { /* assistant message object */ },
>   "errorMessage": "..." // 仅在 state=error 时
> }
> ```

**发送 /stop 命令时的返回：**
```json
{
  "ok": true,
  "aborted": true,
  "runIds": ["run-id-1"]
}
```

---

### chat.abort
**权限**: `operator.write`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| sessionKey | string | ✅ | Session key |
| runId | string | ❌ | 指定要中止的 run ID，不填则中止该 session 所有运行 |

**返回：**
```json
{
  "ok": true,
  "aborted": true,
  "runIds": ["run-id-1", "run-id-2"]
}
```

---

### chat.inject
**权限**: `operator.write`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| sessionKey | string | ✅ | Session key |
| message | string | ✅ | 注入的消息内容 |
| label | string | ❌ | 消息标签 (会显示为 `[label]` 前缀) |

**返回：**
```json
{
  "ok": true,
  "messageId": "msg-uuid"
}
```

---

## 4. sessionsHandlers

### sessions.list
**权限**: `operator.read`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| includeGlobal | boolean | ❌ | 包含全局 session |
| includeUnknown | boolean | ❌ | 包含未知/未归类 session |
| agentId | string | ❌ | 按 agent ID 过滤 |
| spawnedBy | string | ❌ | 按 spawnedBy 过滤 |
| search | string | ❌ | 搜索关键词 |
| label | string | ❌ | 按 label 过滤 |
| limit | number | ❌ | 返回数量限制 |

**返回：**
```json
{
  "sessions": [
    {
      "key": "global",
      "sessionId": "uuid",
      "agentId": "default",
      "label": "...",
      "updatedAt": 1709000000000,
      "thinkingLevel": "medium",
      "verboseLevel": "...",
      "modelOverride": "...",
      "providerOverride": "...",
      "spawnedBy": "...",
      "channel": "...",
      "groupId": "...",
      "totalTokens": 5000,
      "inputTokens": 2000,
      "outputTokens": 3000,
      "sendPolicy": "allow" | "deny"
      // ... 更多 session entry 字段
    }
  ]
}
```

---

### sessions.preview
**权限**: `operator.read`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| keys | string[] | ✅ | Session key 列表 (最多 64) |
| limit | number | ❌ | 每个 session 预览条目数 (默认 12) |
| maxChars | number | ❌ | 每条预览的最大字符数 (默认 240) |

**返回：**
```json
{
  "ts": 1709000000000,
  "previews": [
    {
      "key": "global",
      "status": "ok" | "missing" | "empty" | "error",
      "items": [
        // 预览消息摘要
      ]
    }
  ]
}
```

---

### sessions.resolve
**权限**: `operator.read`

**参数：** 提供以下之一 (互斥):
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| key | string | ❌* | Session key |
| sessionId | string | ❌* | Session UUID |
| label | string | ❌* | Session 标签 |
| agentId | string | ❌ | Agent ID 辅助定位 |
| spawnedBy | string | ❌ | SpawnedBy 辅助定位 |
| includeGlobal | boolean | ❌ | 包含全局 session |
| includeUnknown | boolean | ❌ | 包含未知 session |

> *三选一必填

**返回：**
```json
{
  "ok": true,
  "key": "canonical-session-key"
}
```

---

### sessions.patch
**权限**: `operator.admin`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| key | string | ✅ | Session key |
| thinkingLevel | string/null | ❌ | "low"/"medium"/"high"/"xhigh", null 清除 |
| verboseLevel | string/null | ❌ | verbose 级别, null 清除 |
| reasoningLevel | string/null | ❌ | "on"/"off"/"stream", null 清除 |
| responseUsage | string/null | ❌ | "off"/"tokens"/"full", null 清除 |
| elevatedLevel | string/null | ❌ | "on"/"off"/"ask"/"full", null 清除 |
| model | string/null | ❌ | 模型 ID (如 "anthropic/claude-sonnet-4-20250514"), null 恢复默认 |
| label | string/null | ❌ | Session 标签, null 清除 |
| spawnedBy | string/null | ❌ | Parent session key (仅 subagent sessions) |
| sendPolicy | string/null | ❌ | "allow"/"deny", null 清除 |
| groupActivation | string/null | ❌ | "mention"/"always", null 清除 |
| execHost | string/null | ❌ | "sandbox"/"gateway"/"node", null 清除 |
| execSecurity | string/null | ❌ | "deny"/"allowlist"/"full", null 清除 |
| execAsk | string/null | ❌ | "off"/"on-miss"/"always", null 清除 |
| execNode | string/null | ❌ | Node ID, null 清除 |

**返回：**
```json
{
  "ok": true,
  "path": "/path/to/sessions.json",
  "key": "canonical-key",
  "entry": {
    "sessionId": "uuid",
    "updatedAt": 1709000000000,
    "thinkingLevel": "high",
    "modelOverride": "claude-sonnet-4-20250514",
    "providerOverride": "anthropic"
    // ... 完整 session entry
  },
  "resolved": {
    "modelProvider": "anthropic",
    "model": "claude-sonnet-4-20250514"
  }
}
```

---

### sessions.reset
**权限**: `operator.admin`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| key | string | ✅ | Session key |
| reason | string | ❌ | 重置原因 ("new" 或其他) |

**返回：**
```json
{
  "ok": true,
  "key": "canonical-key",
  "entry": {
    "sessionId": "new-uuid",
    "updatedAt": 1709000000000,
    "systemSent": false,
    "abortedLastRun": false
    // ... 新 session entry
  }
}
```

---

### sessions.delete
**权限**: `operator.admin`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| key | string | ✅ | Session key (不能是 main session) |
| deleteTranscript | boolean | ❌ | 是否删除 transcript (默认 true) |

**返回：**
```json
{
  "ok": true,
  "key": "canonical-key",
  "deleted": true,
  "archived": ["path/to/archived-transcript.jsonl"]
}
```

---

### sessions.compact
**权限**: `operator.admin`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| key | string | ✅ | Session key |
| maxLines | number | ❌ | 保留的最大行数 (默认 400) |

**返回：**
```json
{
  "ok": true,
  "key": "canonical-key",
  "compacted": true,
  "archived": "path/to/backup.bak",
  "kept": 400
}
```

---

## 5. agentHandlers

### agent
**权限**: `operator.write`

这是主要的 agent 执行入口。发送消息给 agent 处理。

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| idempotencyKey | string | ✅ | 幂等键/运行 ID |
| message | string | ✅ | 消息内容 |
| agentId | string | ❌ | 目标 agent ID |
| sessionKey | string | ❌ | 目标 session key |
| sessionId | string | ❌ | Session UUID |
| channel | string | ❌ | 消息渠道 |
| replyChannel | string | ❌ | 回复渠道 |
| to / replyTo | string | ❌ | 回复目标 |
| threadId | string | ❌ | 线程 ID |
| accountId / replyAccountId | string | ❌ | 账号 ID |
| deliver | boolean | ❌ | 是否投递到外部渠道 |
| thinking | string | ❌ | 思考级别 |
| timeout | number | ❌ | 超时（秒） |
| label | string | ❌ | 运行标签 |
| lane | string | ❌ | 执行通道 |
| extraSystemPrompt | string | ❌ | 额外系统提示 |
| inputProvenance | object | ❌ | 输入来源信息 |
| groupId | string | ❌ | 群组 ID |
| groupChannel | string | ❌ | 群组渠道 |
| groupSpace | string | ❌ | 群组空间 |
| spawnedBy | string | ❌ | 生成此运行的父 session |
| attachments | Attachment[] | ❌ | 附件列表 |
| attachments[].type | string | ❌ | 附件类型 |
| attachments[].mimeType | string | ❌ | MIME 类型 |
| attachments[].fileName | string | ❌ | 文件名 |
| attachments[].content | string/Buffer | ✅ | 内容 |

**返回（立即 — accepted）：**
```json
{
  "runId": "idempotency-key",
  "status": "accepted",
  "acceptedAt": 1709000000000
}
```

**完成后的第二次 respond（异步）：**
```json
{
  "runId": "...",
  "status": "ok",
  "summary": "completed",
  "result": { /* agent 执行结果 */ }
}
```

**错误时：**
```json
{
  "runId": "...",
  "status": "error",
  "summary": "Error: ..."
}
```

---

### agent.identity.get
**权限**: `operator.read`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| agentId | string | ❌ | Agent ID |
| sessionKey | string | ❌ | Session key (可从中解析 agentId) |

**返回：**
```json
{
  "agentId": "default",
  "name": "Agent Name",
  "emoji": "🤖",
  "avatar": "https://..." | "/path/to/avatar.png",
  "description": "..."
}
```

---

### agent.wait
**权限**: `operator.read`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| runId | string | ✅ | 要等待的运行 ID |
| timeoutMs | number | ❌ | 超时毫秒数 (默认 30000) |

**返回（完成时）：**
```json
{
  "runId": "...",
  "status": "ok" | "error",
  "startedAt": 1709000000000,
  "endedAt": 1709000030000,
  "error": null
}
```

**返回（超时时）：**
```json
{
  "runId": "...",
  "status": "timeout"
}
```

---

## 6. agentsHandlers

### agents.list
**权限**: `operator.read`

**参数：** 无（通过 AJV 验证但无特定字段要求）

**返回：**
```json
{
  "agents": [
    {
      "id": "default",
      "name": "Default Agent",
      "workspace": "/path/to/workspace",
      "model": "...",
      "emoji": "...",
      "avatar": "..."
    }
  ]
}
```

---

### agents.create
**权限**: `operator.admin`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | ✅ | Agent 名称 |
| workspace | string | ❌ | 工作目录路径 |
| emoji | string | ❌ | Emoji 标识 |
| avatar | string | ❌ | 头像 URL 或路径 |

**返回：**
```json
{
  "ok": true,
  "agentId": "normalized-agent-id",
  "name": "原始名称",
  "workspace": "/resolved/path"
}
```

---

### agents.update
**权限**: `operator.admin`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| agentId | string | ✅ | Agent ID |
| name | string | ❌ | 新名称 |
| workspace | string | ❌ | 新工作目录 |
| model | string | ❌ | 新模型 |
| avatar | string | ❌ | 新头像 |

**返回：**
```json
{
  "ok": true,
  "agentId": "agent-id"
}
```

---

### agents.delete
**权限**: `operator.admin`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| agentId | string | ✅ | Agent ID (不能是 "default") |
| deleteFiles | boolean | ❌ | 是否删除关联文件 (默认 true) |

**返回：**
```json
{
  "ok": true,
  "agentId": "agent-id",
  "removedBindings": [...]
}
```

---

### agents.files.list
**权限**: `operator.read`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| agentId | string | ✅ | Agent ID |

**返回：**
```json
{
  "agentId": "default",
  "workspace": "/path/to/workspace",
  "files": [
    {
      "name": "SOUL.md",
      "path": "/full/path/SOUL.md",
      "missing": false,
      "size": 1234,
      "updatedAtMs": 1709000000000
    },
    {
      "name": "TOOLS.md",
      "path": "/full/path/TOOLS.md",
      "missing": true
    }
  ]
}
```

> 检查的文件名: AGENTS.md, SOUL.md, TOOLS.md, IDENTITY.md, USER.md, HEARTBEAT.md, bootstrap.md, MEMORY.md (或 memory.md)

---

### agents.files.get
**权限**: `operator.read`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| agentId | string | ✅ | Agent ID |
| name | string | ✅ | 文件名 (必须是允许的文件名) |

**返回（文件存在）：**
```json
{
  "agentId": "default",
  "workspace": "/path",
  "file": {
    "name": "SOUL.md",
    "path": "/full/path/SOUL.md",
    "missing": false,
    "size": 1234,
    "updatedAtMs": 1709000000000,
    "content": "文件完整内容..."
  }
}
```

**返回（文件不存在）：**
```json
{
  "agentId": "default",
  "workspace": "/path",
  "file": {
    "name": "SOUL.md",
    "path": "/full/path/SOUL.md",
    "missing": true
  }
}
```

---

### agents.files.set
**权限**: `operator.admin`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| agentId | string | ✅ | Agent ID |
| name | string | ✅ | 文件名 (必须是允许的文件名) |
| content | string | ✅ | 文件内容 |

**返回：**
```json
{
  "ok": true,
  "agentId": "default",
  "workspace": "/path",
  "file": {
    "name": "SOUL.md",
    "path": "/full/path/SOUL.md",
    "missing": false,
    "size": 1234,
    "updatedAtMs": 1709000000000,
    "content": "新内容"
  }
}
```

---

## 7. configHandlers

### config.get
**权限**: `operator.read`

**参数：** 无

**返回：**
```json
{
  "valid": true,
  "raw": "{ 原始JSON5内容 }",
  "config": { /* 完整配置对象, 敏感字段已脱敏 */ },
  "hash": "sha256-hash",
  "path": "/path/to/config.json5"
}
```

---

### config.schema
**权限**: `operator.read`

**参数：** 无

**返回：**
```json
{
  "schema": { /* JSON Schema */ },
  "uiHints": { /* UI 提示 */ },
  "fieldHelp": { /* 字段帮助文本 */ },
  "fieldLabels": { /* 字段标签 */ }
}
```

---

### config.set
**权限**: `operator.admin`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| raw | string | ✅ | 完整配置 JSON5 字符串 |
| baseHash | string | ❌ | 基准 hash 用于乐观锁 |

**返回：**
```json
{
  "ok": true,
  "path": "/path/to/config.json5",
  "config": { /* 写入后的配置 (脱敏) */ }
}
```

---

### config.patch
**权限**: `operator.admin`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| raw | string | ✅ | 部分配置 JSON5 字符串 (merge patch) |
| baseHash | string | ❌ | 基准 hash 用于乐观锁 |
| sessionKey | string | ❌ | 关联的 session key |
| note | string | ❌ | 变更说明 |
| restartDelayMs | number | ❌ | 重启延迟毫秒数 |
| deliveryContext | object | ❌ | 投递上下文 |
| threadId | string | ❌ | 线程 ID |

**返回：**
```json
{
  "ok": true,
  "path": "/path/to/config.json5",
  "config": { /* merge 后的配置 (脱敏) */ },
  "restart": {
    "scheduled": true,
    "delayMs": 1000,
    "reason": "config.patch"
  },
  "sentinel": {
    "path": "/path/to/sentinel",
    "payload": { ... }
  }
}
```

---

### config.apply
**权限**: `operator.admin`

**参数：** 同 config.set

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| raw | string | ✅ | 完整配置 JSON5 字符串 |
| baseHash | string | ❌ | 基准 hash |
| sessionKey | string | ❌ | 关联 session key |
| note | string | ❌ | 变更说明 |
| restartDelayMs | number | ❌ | 重启延迟 |

**返回：** 同 config.patch，包含 restart 和 sentinel 信息。

---

## 8. modelsHandlers

### models.list
**权限**: `operator.read`

**参数：** 无

**返回：**
```json
{
  "models": [
    {
      "id": "anthropic/claude-sonnet-4-20250514",
      "provider": "anthropic",
      "model": "claude-sonnet-4-20250514",
      "displayName": "Claude Sonnet 4",
      "contextWindow": 200000,
      "maxOutput": 16384,
      "supportsImages": true,
      "supportsThinking": true,
      "thinkingLevels": ["low", "medium", "high"]
      // ... 更多模型属性
    }
  ]
}
```

---

## 9. nodeHandlers

### node.list
**权限**: `operator.read`

**参数：** 无

**返回：**
```json
{
  "ts": 1709000000000,
  "nodes": [
    {
      "nodeId": "device-uuid",
      "displayName": "My MacBook",
      "platform": "darwin",
      "version": "2026.2.14",
      "coreVersion": "...",
      "uiVersion": "...",
      "deviceFamily": "Mac",
      "modelIdentifier": "MacBookPro18,3",
      "remoteIp": "192.168.1.100",
      "caps": ["browser", "exec"],
      "commands": ["system.run", "browser.proxy"],
      "pathEnv": "/usr/bin:...",
      "permissions": { ... },
      "connectedAtMs": 1709000000000,
      "paired": true,
      "connected": true
    }
  ]
}
```

---

### node.describe
**权限**: `operator.read`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| nodeId | string | ✅ | Node ID |

**返回：**
```json
{
  "ts": 1709000000000,
  "nodeId": "device-uuid",
  "displayName": "My MacBook",
  "platform": "darwin",
  "version": "2026.2.14",
  "coreVersion": "...",
  "uiVersion": "...",
  "deviceFamily": "Mac",
  "modelIdentifier": "MacBookPro18,3",
  "remoteIp": "192.168.1.100",
  "caps": ["browser", "exec"],
  "commands": ["system.run", "browser.proxy"],
  "pathEnv": "/usr/bin:...",
  "permissions": { ... },
  "connectedAtMs": 1709000000000,
  "paired": true,
  "connected": true
}
```

---

### node.invoke
**权限**: `operator.write`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| nodeId | string | ✅ | Node ID |
| command | string | ✅ | 要执行的命令 (不能是 system.execApprovals.*) |
| params | any | ❌ | 命令参数 |
| timeoutMs | number | ❌ | 超时毫秒数 |
| idempotencyKey | string | ❌ | 幂等键 |

**返回：**
```json
{
  "ok": true,
  "nodeId": "device-uuid",
  "command": "system.run",
  "payload": { /* node 返回的结果 */ },
  "payloadJSON": "..." | null
}
```

---

### node.invoke.result
**权限**: Node 端调用

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | string | ✅ | Invoke 请求 ID |
| nodeId | string | ✅ | Node ID |
| ok | boolean | ✅ | 是否成功 |
| payload | any | ❌ | 返回数据 |
| payloadJSON | string | ❌ | JSON 格式的返回数据 |
| error | object | ❌ | 错误对象 |

**返回：**
```json
{ "ok": true }
```

---

### node.event
**权限**: Node 端调用

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| event | string | ✅ | 事件类型 |
| payload | any | ❌ | 事件数据 |
| payloadJSON | string | ❌ | JSON 格式的事件数据 |

**返回：**
```json
{ "ok": true }
```

---

### node.pair.request
**权限**: `operator.pairing`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| nodeId | string | ✅ | Node ID |
| displayName | string | ❌ | 显示名称 |
| platform | string | ❌ | 平台 |
| version | string | ❌ | 版本 |
| coreVersion | string | ❌ | 核心版本 |
| uiVersion | string | ❌ | UI 版本 |
| deviceFamily | string | ❌ | 设备系列 |
| modelIdentifier | string | ❌ | 型号标识 |
| caps | string[] | ❌ | 能力列表 |
| commands | string[] | ❌ | 可用命令列表 |
| remoteIp | string | ❌ | 远程 IP |
| silent | boolean | ❌ | 静默配对 |

**返回：**
```json
{
  "status": "pending",
  "request": { /* 配对请求对象 */ },
  "created": true
}
```

---

### node.pair.list
**权限**: `operator.read`

**参数：** 无

**返回：** listNodePairing() 返回值

---

### node.pair.approve
**权限**: `operator.admin`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| requestId | string | ✅ | 配对请求 ID |

**返回：** 批准后的 node 对象

---

### node.pair.reject
**权限**: `operator.admin`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| requestId | string | ✅ | 配对请求 ID |

**返回：** 拒绝后的结果

---

### node.pair.verify
**权限**: `operator.read`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| nodeId | string | ✅ | Node ID |
| token | string | ✅ | Token |

**返回：** 验证结果

---

### node.rename
**权限**: `operator.admin`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| nodeId | string | ✅ | Node ID |
| displayName | string | ✅ | 新显示名称 |

**返回：**
```json
{
  "nodeId": "...",
  "displayName": "新名称"
}
```

---

## 10. healthHandlers

### health
**权限**: `operator.read`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| probe | boolean | ❌ | 是否执行深度探测 (默认 false, 使用缓存) |

**返回：** 健康快照对象，包含各子系统状态

---

### status
**权限**: `operator.read`

**参数：** 无

**返回：** `getStatusSummary()` 返回值

---

## 11. channelsHandlers

### channels.status
**权限**: `operator.read`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| probe | boolean | ❌ | 是否主动探测各渠道连接状态 |
| timeoutMs | number | ❌ | 探测超时 (默认 10000, 最小 1000) |

**返回：**
```json
{
  "ts": 1709000000000,
  "channelOrder": ["whatsapp", "telegram", "slack", ...],
  "channelLabels": { "whatsapp": "WhatsApp", ... },
  "channelDetailLabels": { ... },
  "channelSystemImages": { ... },
  "channelMeta": { ... },
  "channels": {
    "whatsapp": {
      "configured": true,
      // ... channel summary
    }
  },
  "channelAccounts": {
    "whatsapp": [
      {
        "accountId": "default",
        "configured": true,
        "connected": true,
        "lastInboundAt": 1709000000000,
        "lastOutboundAt": 1709000000000,
        "lastProbeAt": 1709000000000
        // ...
      }
    ]
  },
  "channelDefaultAccountId": {
    "whatsapp": "default"
  }
}
```

---

### channels.logout
**权限**: `operator.admin`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| channel | string | ✅ | 渠道 ID |
| accountId | string | ❌ | 账号 ID |

**返回：**
```json
{
  "channel": "whatsapp",
  "accountId": "default",
  "loggedOut": true,
  "cleared": true
}
```

---

## 12. connectHandlers

### connect
**权限**: N/A (仅作为首个请求有效)

**参数：** 无 (在 connect 阶段之外调用会报错)

**返回：** 错误 — "connect is only valid as the first request"

> 注: `connect` 是 WebSocket 连接的握手方法，在 `handleGatewayRequest` 之前已单独处理。

---

## 13. logsHandlers

### logs.tail
**权限**: `operator.read`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| cursor | number | ❌ | 文件偏移游标 (字节) |
| limit | number | ❌ | 最大返回行数 (默认 500, 最大 5000) |
| maxBytes | number | ❌ | 最大读取字节数 (默认 250000, 最大 1000000) |

**返回：**
```json
{
  "file": "/path/to/openclaw-2026-02-27.log",
  "cursor": 123456,
  "size": 200000,
  "lines": ["log line 1", "log line 2", ...],
  "truncated": false,
  "reset": false
}
```

---

## 14. deviceHandlers

### device.pair.list
**权限**: `operator.read`

**参数：** 无

**返回：** 设备配对列表 (pending + paired)

---

### device.pair.approve / device.pair.reject
**权限**: `operator.admin`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| requestId | string | ✅ | 配对请求 ID |

---

### device.token.rotate / device.token.revoke
**权限**: `operator.admin`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| deviceId | string | ✅ | 设备 ID |
| role | string | ✅ | 角色 |
| scopes | string[] | ❌ | 权限范围 |

---

## 15. execApprovalsHandlers

### exec.approvals.get
**权限**: `operator.read`

**参数：** 无

**返回：**
```json
{
  "path": "/path/to/exec-approvals.json",
  "exists": true,
  "hash": "sha256-hash",
  "file": { /* exec approvals 配置 (脱敏) */ }
}
```

---

### exec.approvals.set
**权限**: `operator.admin`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| file | object | ✅ | exec approvals 配置对象 |
| baseHash | string | ❌ | 基准 hash |

**返回：** 同 exec.approvals.get

---

### exec.approvals.node.get / exec.approvals.node.set
**权限**: `operator.admin`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| nodeId | string | ✅ | Node ID |
| file | object | ❌ | (set 时) exec approvals 配置 |
| baseHash | string | ❌ | (set 时) 基准 hash |

---

## 16. skillsHandlers

### skills.status
**权限**: `operator.read`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| agentId | string | ❌ | Agent ID (默认 default agent) |

**返回：** 工作区技能状态对象

---

### skills.bins
**权限**: `operator.read`

**参数：** 无

**返回：**
```json
{
  "bins": ["git", "python", "node", ...]
}
```

---

### skills.install
**权限**: `operator.admin`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | ✅ | 技能名称 |
| installId | string | ❌ | 安装 ID |
| timeoutMs | number | ❌ | 超时毫秒数 |

**返回：** 安装结果对象

---

### skills.update
**权限**: `operator.admin`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| skillKey | string | ✅ | 技能 key |
| enabled | boolean | ❌ | 是否启用 |
| apiKey | string | ❌ | API key |
| env | object | ❌ | 环境变量 { key: value } |

**返回：**
```json
{
  "ok": true,
  "skillKey": "skill-key",
  "config": { "enabled": true, "apiKey": "***", "env": { ... } }
}
```

---

## 17. systemHandlers

### last-heartbeat
**权限**: `operator.read`

**参数：** 无

**返回：** 最后一次心跳事件对象

---

### set-heartbeats
**权限**: `operator.write`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| enabled | boolean | ✅ | 启用/禁用心跳 |

**返回：**
```json
{
  "ok": true,
  "enabled": true
}
```

---

### system-presence
**权限**: `operator.read`

**参数：** 无

**返回：** 系统在线状态列表

---

### system-event
**权限**: `operator.write`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| text | string | ✅ | 事件文本 |
| deviceId | string | ❌ | 设备 ID |
| instanceId | string | ❌ | 实例 ID |
| host | string | ❌ | 主机名 |
| ip | string | ❌ | IP 地址 |
| mode | string | ❌ | 模式 |
| version | string | ❌ | 版本 |
| platform | string | ❌ | 平台 |
| deviceFamily | string | ❌ | 设备系列 |
| modelIdentifier | string | ❌ | 型号标识 |
| lastInputSeconds | number | ❌ | 最后输入距今秒数 |
| reason | string | ❌ | 事件原因 |
| roles | string[] | ❌ | 角色列表 |
| scopes | string[] | ❌ | 权限范围列表 |
| tags | string[] | ❌ | 标签列表 |

**返回：**
```json
{ "ok": true }
```

---

## 18. talkHandlers

### talk.config
**权限**: `operator.read` (secrets 需要 `operator.talk.secrets` 或 `operator.admin`)

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| includeSecrets | boolean | ❌ | 包含 API key 等敏感信息 |

**返回：**
```json
{
  "config": {
    "talk": {
      "voiceId": "...",
      "voiceAliases": { "alias": "voice-id" },
      "modelId": "...",
      "outputFormat": "mp3",
      "apiKey": "..." // 仅 includeSecrets=true
    },
    "session": { "mainKey": "global" },
    "ui": { "seamColor": "#000" }
  }
}
```

---

### talk.mode
**权限**: `operator.write`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| enabled | boolean | ✅ | 启用/禁用通话模式 |
| phase | string | ❌ | 通话阶段 |

**返回：**
```json
{
  "enabled": true,
  "phase": null,
  "ts": 1709000000000
}
```

---

## 19. ttsHandlers

### tts.status
**权限**: `operator.read`

**参数：** 无

**返回：**
```json
{
  "enabled": true,
  "auto": "...",
  "provider": "openai",
  "fallbackProvider": "edge",
  "fallbackProviders": ["edge"],
  "prefsPath": "/path/to/tts-prefs.json",
  "hasOpenAIKey": true,
  "hasElevenLabsKey": false,
  "edgeEnabled": true
}
```

---

### tts.enable / tts.disable
**权限**: `operator.write`

**参数：** 无

**返回：**
```json
{ "enabled": true }  // 或 false
```

---

### tts.convert
**权限**: `operator.write`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| text | string | ✅ | 要转换的文本 |
| channel | string | ❌ | 渠道 |

**返回：**
```json
{
  "audioPath": "/path/to/audio.mp3",
  "provider": "openai",
  "outputFormat": "mp3",
  "voiceCompatible": true
}
```

---

### tts.setProvider
**权限**: `operator.write`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| provider | string | ✅ | "openai" / "elevenlabs" / "edge" |

**返回：**
```json
{ "provider": "openai" }
```

---

### tts.providers
**权限**: `operator.read`

**参数：** 无

**返回：**
```json
{
  "providers": [
    {
      "id": "openai",
      "name": "OpenAI",
      "configured": true,
      "models": ["tts-1", "tts-1-hd", "gpt-4o-mini-tts"],
      "voices": ["alloy", "echo", "fable", "onyx", "nova", "shimmer", "ash", "ballad", "coral", "sage", "verse"]
    },
    {
      "id": "elevenlabs",
      "name": "ElevenLabs",
      "configured": false,
      "models": ["eleven_multilingual_v2", "eleven_turbo_v2_5", "eleven_monolingual_v1"]
    },
    {
      "id": "edge",
      "name": "Edge TTS",
      "configured": true,
      "models": []
    }
  ],
  "active": "openai"
}
```

---

## 20. updateHandlers

### update.run
**权限**: `operator.admin`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| sessionKey | string | ❌ | 关联 session key |
| note | string | ❌ | 更新说明 |
| restartDelayMs | number | ❌ | 重启延迟毫秒数 |
| timeoutMs | number | ❌ | 更新超时 (最小 1000) |

**返回：**
```json
{
  "ok": true,
  "result": {
    "status": "ok" | "error",
    "mode": "npm" | "git" | "unknown",
    "root": "/path/to/openclaw",
    "before": "2026.2.13",
    "after": "2026.2.14",
    "steps": [
      {
        "name": "npm install",
        "command": "npm install -g openclaw",
        "cwd": "/path",
        "durationMs": 5000,
        "log": {
          "stdoutTail": "...",
          "stderrTail": "...",
          "exitCode": 0
        }
      }
    ],
    "reason": null,
    "durationMs": 10000
  },
  "restart": {
    "scheduled": true,
    "delayMs": 1000,
    "reason": "update.run"
  },
  "sentinel": {
    "path": "/path/to/sentinel",
    "payload": { ... }
  }
}
```

---

## 21. usageHandlers

### usage.status
**权限**: `operator.read`

**参数：** 无

**返回：** `loadProviderUsageSummary()` — 各 provider 的 API 用量摘要

---

### usage.cost
**权限**: `operator.read`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| startDate | string | ❌ | 开始日期 YYYY-MM-DD |
| endDate | string | ❌ | 结束日期 YYYY-MM-DD |
| days | number | ❌ | 天数 (与 startDate/endDate 互斥, 默认约 30 天) |

**返回：** 成本用量摘要对象

---

### sessions.usage
**权限**: `operator.read`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| startDate | string | ❌ | 开始日期 YYYY-MM-DD |
| endDate | string | ❌ | 结束日期 YYYY-MM-DD |
| limit | number | ❌ | 返回条数 (默认 50) |
| includeContextWeight | boolean | ❌ | 包含上下文权重 |
| key | string | ❌ | 指定 session key |

**返回：** 各 session 的用量数据

---

## 22. voicewakeHandlers

### voicewake.get
**权限**: `operator.read`

**参数：** 无

**返回：**
```json
{
  "triggers": ["hey openclaw", "ok claw"]
}
```

---

### voicewake.set
**权限**: `operator.write`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| triggers | string[] | ✅ | 唤醒词列表 |

**返回：**
```json
{
  "triggers": ["hey openclaw", "ok claw"]
}
```

---

## 23. webHandlers

### web.login.start
**权限**: `operator.admin`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| force | boolean | ❌ | 强制重新登录 |
| timeoutMs | number | ❌ | 超时毫秒数 |
| verbose | boolean | ❌ | 详细输出 |
| accountId | string | ❌ | 账号 ID |

**返回：** 登录启动结果 (QR 码等)

---

### web.login.wait
**权限**: `operator.admin`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| timeoutMs | number | ❌ | 等待超时 |
| accountId | string | ❌ | 账号 ID |

**返回：** 登录等待结果

---

## 24. wizardHandlers

### wizard.start
**权限**: `operator.admin`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| mode | string | ✅ | 向导模式 |
| workspace | string | ❌ | 工作目录 |

**返回：** 向导 session 状态和第一个 step

---

### wizard.next / wizard.cancel / wizard.status
**权限**: `operator.admin`

交互式向导的后续步骤。

---

## 25. browserHandlers

### browser.request
**权限**: `operator.write`

**参数：**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| method | string | ✅ | HTTP 方法 ("GET" / "POST" / "DELETE") |
| path | string | ✅ | 请求路径 |
| query | object | ❌ | 查询参数 |
| body | any | ❌ | 请求体 |
| timeoutMs | number | ❌ | 超时毫秒数 |

**返回：** 浏览器控制结果 (取决于具体 path)

---

## 方法总览表

| 方法名 | Handler 组 | 权限 | 说明 |
|--------|-----------|------|------|
| `connect` | connectHandlers | - | WebSocket 握手 |
| `health` | healthHandlers | read | 系统健康状态 |
| `status` | healthHandlers | read | 状态摘要 |
| `wake` | cronHandlers | write | 唤醒 agent |
| `send` | sendHandlers | write | 发送消息到外部渠道 |
| `poll` | sendHandlers | write | 发送投票 |
| `agent` | agentHandlers | write | 执行 agent 任务 |
| `agent.identity.get` | agentHandlers | read | 获取 agent 身份 |
| `agent.wait` | agentHandlers | read | 等待 agent 完成 |
| `agents.list` | agentsHandlers | read | 列出所有 agent |
| `agents.create` | agentsHandlers | admin | 创建 agent |
| `agents.update` | agentsHandlers | admin | 更新 agent |
| `agents.delete` | agentsHandlers | admin | 删除 agent |
| `agents.files.list` | agentsHandlers | read | 列出 agent 文件 |
| `agents.files.get` | agentsHandlers | read | 获取 agent 文件 |
| `agents.files.set` | agentsHandlers | admin | 设置 agent 文件 |
| `chat.history` | chatHandlers | read | 聊天历史 |
| `chat.send` | chatHandlers | write | 发送聊天消息 |
| `chat.abort` | chatHandlers | write | 中止聊天 |
| `chat.inject` | chatHandlers | write | 注入 transcript 消息 |
| `sessions.list` | sessionsHandlers | read | 列出 sessions |
| `sessions.preview` | sessionsHandlers | read | 预览 sessions |
| `sessions.resolve` | sessionsHandlers | read | 解析 session key |
| `sessions.patch` | sessionsHandlers | admin | 修改 session 设置 |
| `sessions.reset` | sessionsHandlers | admin | 重置 session |
| `sessions.delete` | sessionsHandlers | admin | 删除 session |
| `sessions.compact` | sessionsHandlers | admin | 压缩 session transcript |
| `sessions.usage` | usageHandlers | read | Session 用量 |
| `config.get` | configHandlers | read | 获取配置 |
| `config.schema` | configHandlers | read | 获取配置 schema |
| `config.set` | configHandlers | admin | 覆盖配置 |
| `config.patch` | configHandlers | admin | 合并配置 |
| `config.apply` | configHandlers | admin | 应用配置 |
| `models.list` | modelsHandlers | read | 列出可用模型 |
| `node.list` | nodeHandlers | read | 列出节点 |
| `node.describe` | nodeHandlers | read | 节点详情 |
| `node.invoke` | nodeHandlers | write | 调用节点命令 |
| `node.invoke.result` | nodeHandlers | - | 节点返回结果 |
| `node.event` | nodeHandlers | - | 节点事件 |
| `node.pair.request` | nodeHandlers | pairing | 请求节点配对 |
| `node.pair.list` | nodeHandlers | read | 列出节点配对 |
| `node.pair.approve` | nodeHandlers | admin | 批准配对 |
| `node.pair.reject` | nodeHandlers | admin | 拒绝配对 |
| `node.pair.verify` | nodeHandlers | read | 验证节点 token |
| `node.rename` | nodeHandlers | admin | 重命名节点 |
| `cron.list` | cronHandlers | read | 列出定时任务 |
| `cron.status` | cronHandlers | read | 定时系统状态 |
| `cron.add` | cronHandlers | admin | 添加定时任务 |
| `cron.update` | cronHandlers | admin | 更新定时任务 |
| `cron.remove` | cronHandlers | admin | 删除定时任务 |
| `cron.run` | cronHandlers | admin | 手动执行任务 |
| `cron.runs` | cronHandlers | read | 任务运行日志 |
| `channels.status` | channelsHandlers | read | 渠道状态 |
| `channels.logout` | channelsHandlers | admin | 渠道登出 |
| `logs.tail` | logsHandlers | read | 日志尾部 |
| `device.pair.list` | deviceHandlers | read | 设备配对列表 |
| `device.pair.approve` | deviceHandlers | admin | 批准设备配对 |
| `device.pair.reject` | deviceHandlers | admin | 拒绝设备配对 |
| `device.token.rotate` | deviceHandlers | admin | 轮转设备 token |
| `device.token.revoke` | deviceHandlers | admin | 吊销设备 token |
| `exec.approvals.get` | execApprovalsHandlers | read | 获取执行审批 |
| `exec.approvals.set` | execApprovalsHandlers | admin | 设置执行审批 |
| `exec.approvals.node.get` | execApprovalsHandlers | admin | 获取节点审批 |
| `exec.approvals.node.set` | execApprovalsHandlers | admin | 设置节点审批 |
| `skills.status` | skillsHandlers | read | 技能状态 |
| `skills.bins` | skillsHandlers | read | 技能依赖二进制 |
| `skills.install` | skillsHandlers | admin | 安装技能 |
| `skills.update` | skillsHandlers | admin | 更新技能配置 |
| `last-heartbeat` | systemHandlers | read | 最后心跳 |
| `set-heartbeats` | systemHandlers | write | 设置心跳开关 |
| `system-presence` | systemHandlers | read | 系统在线状态 |
| `system-event` | systemHandlers | write | 系统事件 |
| `talk.config` | talkHandlers | read | 通话配置 |
| `talk.mode` | talkHandlers | write | 通话模式 |
| `tts.status` | ttsHandlers | read | TTS 状态 |
| `tts.enable` | ttsHandlers | write | 启用 TTS |
| `tts.disable` | ttsHandlers | write | 禁用 TTS |
| `tts.convert` | ttsHandlers | write | 文本转语音 |
| `tts.setProvider` | ttsHandlers | write | 设置 TTS 提供商 |
| `tts.providers` | ttsHandlers | read | TTS 提供商列表 |
| `update.run` | updateHandlers | admin | 执行更新 |
| `usage.status` | usageHandlers | read | API 用量状态 |
| `usage.cost` | usageHandlers | read | 成本统计 |
| `voicewake.get` | voicewakeHandlers | read | 获取唤醒词 |
| `voicewake.set` | voicewakeHandlers | write | 设置唤醒词 |
| `web.login.start` | webHandlers | admin | 开始 Web 登录 |
| `web.login.wait` | webHandlers | admin | 等待 Web 登录 |
| `wizard.start` | wizardHandlers | admin | 启动向导 |
| `wizard.next` | wizardHandlers | admin | 向导下一步 |
| `wizard.cancel` | wizardHandlers | admin | 取消向导 |
| `wizard.status` | wizardHandlers | admin | 向导状态 |
| `browser.request` | browserHandlers | write | 浏览器控制请求 |

---

> **总计: 80 个 RPC 方法**, 分布在 25 个 handler 组中。
