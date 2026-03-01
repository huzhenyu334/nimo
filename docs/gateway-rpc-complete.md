# OpenClaw Gateway RPC Complete Reference

> Auto-generated from source analysis of `openclaw@2026.2.14` gateway compiled JS.

---

## Table of Contents

- [Transport Protocol](#transport-protocol)
- [Authentication & Scopes](#authentication--scopes)
- [RPC Methods by Namespace](#rpc-methods-by-namespace)
  - [Sessions](#sessions)
  - [Chat](#chat)
  - [Agent](#agent)
  - [Agents (CRUD)](#agents-crud)
  - [Agents Files](#agents-files)
  - [Skills](#skills)
  - [Models](#models)
  - [Config](#config)
  - [Cron](#cron)
  - [Nodes](#nodes)
  - [Devices](#devices)
  - [Exec Approvals](#exec-approvals)
  - [Channels](#channels)
  - [Talk / TTS](#talk--tts)
  - [Logs](#logs)
  - [Usage](#usage)
  - [Wizard](#wizard)
  - [Update](#update)
  - [Web Login](#web-login)
  - [System](#system)
  - [Send / Poll](#send--poll)
  - [Browser](#browser)
- [WebSocket Event Stream](#websocket-event-stream)
  - [Envelope Format](#envelope-format)
  - [Stream: lifecycle](#stream-lifecycle)
  - [Stream: tool](#stream-tool)
  - [Stream: assistant](#stream-assistant)
  - [Stream: compaction](#stream-compaction)
  - [Stream: thinking](#stream-thinking)
  - [Chat Events](#chat-events)
  - [Broadcast Events](#broadcast-events)
  - [Subagent Lifecycle](#subagent-lifecycle)
  - [Plugin Hook Events](#plugin-hook-events)

---

## Transport Protocol

### Frame Types

All communication uses JSON frames over WebSocket:

**Request Frame (client -> server):**
```json
{
  "type": "req",
  "id": "<unique-request-id>",
  "method": "<method-name>",
  "params": { ... }
}
```

**Response Frame (server -> client):**
```json
{
  "type": "res",
  "id": "<matching-request-id>",
  "ok": true,
  "payload": { ... },
  "error": null
}
```

**Error Response:**
```json
{
  "type": "res",
  "id": "<matching-request-id>",
  "ok": false,
  "payload": null,
  "error": {
    "code": "<error-code>",
    "message": "<human-readable>",
    "details": { ... },
    "retryable": false,
    "retryAfterMs": 0
  }
}
```

**Event Frame (server -> client, push):**
```json
{
  "type": "event",
  "event": "<event-name>",
  "payload": { ... },
  "seq": 0,
  "stateVersion": { "presence": 0, "health": 0 }
}
```

### Connection Handshake (`connect`)

The first frame from the client **must** be a `connect` request. Sending `connect` after the handshake returns an error.

**Params:**
```json
{
  "minProtocol": 1,
  "maxProtocol": 1,
  "client": {
    "id": "webchat-ui | openclaw-control-ui | webchat | cli | gateway-client | openclaw-macos | openclaw-ios | openclaw-android | node-host | test | fingerprint | openclaw-probe",
    "displayName": "string (optional)",
    "version": "string",
    "platform": "string",
    "deviceFamily": "string (optional)",
    "modelIdentifier": "string (optional)",
    "mode": "webchat | cli | ui | backend | node | probe | test",
    "instanceId": "string (optional)"
  },
  "caps": ["string"],
  "commands": ["string"],
  "permissions": { "<key>": true },
  "pathEnv": "string (optional)",
  "role": "operator | node",
  "scopes": ["operator.admin", "operator.read", "operator.write", "operator.approvals", "operator.pairing"],
  "device": {
    "id": "string",
    "publicKey": "string",
    "signature": "string",
    "signedAt": 0,
    "nonce": "string (optional)"
  },
  "auth": {
    "token": "string (optional)",
    "password": "string (optional)"
  },
  "locale": "string (optional)",
  "userAgent": "string (optional)"
}
```

**hello-ok Response:**
```json
{
  "type": "hello-ok",
  "protocol": 1,
  "server": {
    "version": "string",
    "commit": "string (optional)",
    "host": "string (optional)",
    "connId": "string"
  },
  "features": {
    "methods": ["sessions.list", "chat.send", "..."],
    "events": ["agent-event", "chat", "..."]
  },
  "snapshot": {
    "presence": [{ "host": "...", "ip": "...", "version": "...", "ts": 0, "..." }],
    "health": {},
    "stateVersion": { "presence": 0, "health": 0 },
    "uptimeMs": 0,
    "configPath": "string",
    "stateDir": "string",
    "sessionDefaults": {
      "defaultAgentId": "string",
      "mainKey": "string",
      "mainSessionKey": "string",
      "scope": "per-sender | global"
    },
    "authMode": "none | token | password | trusted-proxy"
  },
  "canvasHostUrl": "string (optional)",
  "auth": {
    "deviceToken": "string",
    "role": "string",
    "scopes": ["string"],
    "issuedAtMs": 0
  },
  "policy": {
    "maxPayload": 1048576,
    "maxBufferedBytes": 4194304,
    "tickIntervalMs": 30000
  }
}
```

---

## Authentication & Scopes

### Scope Constants

| Scope | Value |
|---|---|
| Admin | `operator.admin` |
| Read | `operator.read` |
| Write | `operator.write` |
| Approvals | `operator.approvals` |
| Pairing | `operator.pairing` |

### Authorization Rules

1. **Admin** scope is superuser: grants access to all methods.
2. **Node** role can only call: `node.invoke.result`, `node.event`, `skills.bins`.
3. **Approval methods** require `operator.approvals` scope.
4. **Pairing methods** require `operator.pairing` scope.
5. **Read methods** require `operator.read` OR `operator.write` scope.
6. **Write methods** require `operator.write` scope.
7. Everything else requires `operator.admin`.

### Methods by Required Scope

<details>
<summary><strong>operator.read</strong> (read OR write)</summary>

`health`, `logs.tail`, `channels.status`, `status`, `usage.status`, `usage.cost`,
`tts.status`, `tts.providers`, `models.list`, `agents.list`, `agent.identity.get`,
`skills.status`, `voicewake.get`, `sessions.list`, `sessions.preview`,
`cron.list`, `cron.status`, `cron.runs`, `system-presence`, `last-heartbeat`,
`node.list`, `node.describe`, `chat.history`, `config.get`, `talk.config`

</details>

<details>
<summary><strong>operator.write</strong></summary>

`send`, `agent`, `agent.wait`, `wake`, `talk.mode`, `tts.enable`, `tts.disable`,
`tts.convert`, `tts.setProvider`, `voicewake.set`, `node.invoke`, `chat.send`,
`chat.abort`, `browser.request`

</details>

<details>
<summary><strong>operator.approvals</strong></summary>

`exec.approval.request`, `exec.approval.waitDecision`, `exec.approval.resolve`

</details>

<details>
<summary><strong>operator.pairing</strong></summary>

`node.pair.request`, `node.pair.list`, `node.pair.approve`, `node.pair.reject`,
`node.pair.verify`, `device.pair.list`, `device.pair.approve`, `device.pair.reject`,
`device.token.rotate`, `device.token.revoke`, `node.rename`

</details>

<details>
<summary><strong>operator.admin</strong></summary>

`config.set`, `config.apply`, `config.patch`, `config.schema`, `wizard.*`,
`update.run`, `channels.logout`, `agents.create`, `agents.update`, `agents.delete`,
`agents.files.*`, `skills.install`, `skills.update`, `cron.add`, `cron.update`,
`cron.remove`, `cron.run`, `sessions.patch`, `sessions.reset`, `sessions.delete`,
`sessions.compact`, `sessions.resolve`, `sessions.usage`, `sessions.usage.timeseries`,
`sessions.usage.logs`, `chat.inject`, `poll`, `exec.approvals.*`, `set-heartbeats`,
`system-event`

</details>

<details>
<summary><strong>node role</strong></summary>

`node.invoke.result`, `node.event`, `skills.bins`

</details>

---

## RPC Methods by Namespace

---

### Sessions

#### `sessions.list`

> **Scope:** read

**Params:**
```json
{
  "limit": "integer >= 1 (optional)",
  "activeMinutes": "integer >= 1 (optional)",
  "includeGlobal": "boolean (optional)",
  "includeUnknown": "boolean (optional)",
  "includeDerivedTitles": "boolean (optional)",
  "includeLastMessage": "boolean (optional)",
  "label": "string 1-64 chars (optional)",
  "spawnedBy": "string (optional)",
  "agentId": "string (optional)",
  "search": "string (optional)"
}
```

**Response:**
```json
{
  "sessions": [
    {
      "key": "string",
      "sessionId": "string",
      "agentId": "string",
      "label": "string | null",
      "updatedAt": 0,
      "channel": "string",
      "chatType": "string",
      "model": "string",
      "thinkingLevel": "string",
      "totalTokens": 0,
      "inputTokens": 0,
      "outputTokens": 0,
      "origin": {},
      "lastMessage": "string (if includeLastMessage)"
    }
  ]
}
```

---

#### `sessions.preview`

> **Scope:** read

**Params:**
```json
{
  "keys": ["string (minItems: 1, max 64)"],
  "limit": "integer >= 1 (optional, default 12)",
  "maxChars": "integer >= 20 (optional, default 240)"
}
```

**Response:**
```json
{
  "ts": 0,
  "previews": [
    {
      "key": "string",
      "status": "ok | empty | missing | error",
      "items": ["<truncated message previews>"]
    }
  ]
}
```

---

#### `sessions.resolve`

> **Scope:** admin | **Note:** Not advertised in BASE_METHODS

**Params:**
```json
{
  "key": "string (optional)",
  "sessionId": "string (optional)",
  "label": "string 1-64 chars (optional)",
  "agentId": "string (optional)",
  "spawnedBy": "string (optional)",
  "includeGlobal": "boolean (optional)",
  "includeUnknown": "boolean (optional)"
}
```

Provide exactly one of `key`, `sessionId`, or `label`.

**Response:**
```json
{
  "ok": true,
  "key": "string (resolved canonical session key)"
}
```

---

#### `sessions.patch`

> **Scope:** admin

**Params:**
```json
{
  "key": "string (required)",
  "label": "string 1-64 | null (optional)",
  "thinkingLevel": "string | null (optional, e.g. 'off', 'low', 'medium', 'high', 'xhigh')",
  "verboseLevel": "string | null (optional)",
  "reasoningLevel": "string | null (optional, e.g. 'on', 'off', 'stream')",
  "responseUsage": "'off' | 'tokens' | 'full' | 'on' | null (optional)",
  "elevatedLevel": "string | null (optional, e.g. 'on', 'off', 'ask', 'full')",
  "execHost": "string | null (optional, e.g. 'sandbox', 'gateway', 'node')",
  "execSecurity": "string | null (optional, e.g. 'deny', 'allowlist', 'full')",
  "execAsk": "string | null (optional, e.g. 'off', 'on-miss', 'always')",
  "execNode": "string | null (optional)",
  "model": "string | null (optional, 'provider/model')",
  "spawnedBy": "string | null (optional)",
  "sendPolicy": "'allow' | 'deny' | null (optional)",
  "groupActivation": "'mention' | 'always' | null (optional)"
}
```

**Response:**
```json
{
  "ok": true,
  "path": "string (store path)",
  "key": "string",
  "entry": { "...full session store entry..." },
  "resolved": {
    "modelProvider": "string",
    "model": "string"
  }
}
```

---

#### `sessions.reset`

> **Scope:** admin

**Params:**
```json
{
  "key": "string (required)",
  "reason": "'new' | 'reset' (optional)"
}
```

**Response:**
```json
{
  "ok": true,
  "key": "string",
  "entry": {
    "sessionId": "string (new UUID)",
    "updatedAt": 0,
    "systemSent": false,
    "abortedLastRun": false,
    "thinkingLevel": "string (preserved)",
    "model": "string (preserved)",
    "inputTokens": 0,
    "outputTokens": 0,
    "totalTokens": 0,
    "totalTokensFresh": true
  }
}
```

**Notes:** Archives old transcripts. Aborts running agent. Stops subagents. Clears session queues.

---

#### `sessions.delete`

> **Scope:** admin

**Params:**
```json
{
  "key": "string (required)",
  "deleteTranscript": "boolean (optional, default true)"
}
```

**Response:**
```json
{
  "ok": true,
  "key": "string",
  "deleted": true,
  "archived": ["string (archived file paths)"]
}
```

**Notes:** Cannot delete the main session. Aborts running agent, stops subagents.

---

#### `sessions.compact`

> **Scope:** admin

**Params:**
```json
{
  "key": "string (required)",
  "maxLines": "integer >= 1 (optional, default 400)"
}
```

**Response (compacted):**
```json
{
  "ok": true,
  "key": "string",
  "compacted": true,
  "archived": "string (backup file path)",
  "kept": 400
}
```

**Response (not compacted):**
```json
{
  "ok": true,
  "key": "string",
  "compacted": false,
  "reason": "no sessionId | no transcript",
  "kept": 123
}
```

---

#### `sessions.usage`

> **Scope:** admin

**Params:**
```json
{
  "key": "string (optional, for single-session)",
  "startDate": "string 'YYYY-MM-DD' (optional)",
  "endDate": "string 'YYYY-MM-DD' (optional)",
  "limit": "integer >= 1 (optional, default 50)",
  "includeContextWeight": "boolean (optional)"
}
```

**Response:**
```json
{
  "updatedAt": 0,
  "startDate": "YYYY-MM-DD",
  "endDate": "YYYY-MM-DD",
  "sessions": [
    {
      "key": "string",
      "label": "string",
      "sessionId": "string",
      "updatedAt": 0,
      "agentId": "string",
      "channel": "string",
      "model": "string",
      "usage": {
        "input": 0, "output": 0, "cacheRead": 0, "cacheWrite": 0,
        "totalTokens": 0, "totalCost": 0.0,
        "inputCost": 0.0, "outputCost": 0.0,
        "cacheReadCost": 0.0, "cacheWriteCost": 0.0,
        "missingCostEntries": 0
      }
    }
  ],
  "totals": {
    "input": 0, "output": 0, "cacheRead": 0, "cacheWrite": 0,
    "totalTokens": 0, "totalCost": 0.0,
    "inputCost": 0.0, "outputCost": 0.0
  },
  "aggregates": {
    "messages": { "total": 0, "user": 0, "assistant": 0, "toolCalls": 0, "toolResults": 0, "errors": 0 },
    "tools": { "totalCalls": 0, "uniqueTools": 0, "tools": [{ "name": "...", "count": 0 }] },
    "byModel": [{ "provider": "...", "model": "...", "count": 0, "totals": {} }],
    "byProvider": [{ "provider": "...", "count": 0, "totals": {} }],
    "byAgent": [{ "agentId": "...", "totals": {} }],
    "byChannel": [{ "channel": "...", "totals": {} }],
    "latency": { "count": 0, "avgMs": 0, "minMs": 0, "maxMs": 0, "p95Ms": 0 },
    "daily": [{ "date": "YYYY-MM-DD", "tokens": 0, "cost": 0.0, "messages": 0, "toolCalls": 0, "errors": 0 }]
  }
}
```

---

### Chat

#### `chat.send`

> **Scope:** write

**Params:**
```json
{
  "sessionKey": "string (required)",
  "message": "string (required)",
  "thinking": "string (optional)",
  "deliver": "boolean (optional)",
  "attachments": ["unknown (optional)"],
  "timeoutMs": "integer >= 0 (optional)",
  "idempotencyKey": "string (required)"
}
```

**Immediate Response:**
```json
{
  "runId": "string (same as idempotencyKey)",
  "status": "started"
}
```

**Notes:** Two-phase delivery. The initial response returns `"started"`. Completion is delivered asynchronously via broadcast `chat` events (see [Chat Events](#chat-events)). Idempotent: if the same `idempotencyKey` is sent again, returns `"in_flight"` or the cached result. If the message is `"/stop"`, aborts the session instead.

---

#### `chat.history`

> **Scope:** read

**Params:**
```json
{
  "sessionKey": "string (required)",
  "limit": "integer 1-1000 (optional, default 200)"
}
```

**Response:**
```json
{
  "sessionKey": "string",
  "sessionId": "string | undefined",
  "messages": ["<transcript messages>"],
  "thinkingLevel": "string | undefined",
  "verboseLevel": "string | undefined"
}
```

---

#### `chat.abort`

> **Scope:** write

**Params:**
```json
{
  "sessionKey": "string (required)",
  "runId": "string (optional)"
}
```

**Response:**
```json
{
  "ok": true,
  "aborted": true,
  "runIds": ["string (aborted run IDs)"]
}
```

---

#### `chat.inject`

> **Scope:** admin | **Note:** Not in BASE_METHODS

**Params:**
```json
{
  "sessionKey": "string (required)",
  "message": "string (required)",
  "label": "string max 100 chars (optional)"
}
```

**Response:**
```json
{
  "ok": true,
  "messageId": "string"
}
```

**Notes:** Injects an assistant-role message into the transcript. Broadcasts a `chat` event with `state: "final"`.

---

### Agent

#### `agent`

> **Scope:** write

**Params:**
```json
{
  "message": "string (required)",
  "agentId": "string (optional)",
  "to": "string (optional)",
  "replyTo": "string (optional)",
  "sessionId": "string (optional)",
  "sessionKey": "string (optional)",
  "thinking": "string (optional)",
  "deliver": "boolean (optional)",
  "attachments": ["unknown (optional)"],
  "channel": "string (optional)",
  "replyChannel": "string (optional)",
  "accountId": "string (optional)",
  "replyAccountId": "string (optional)",
  "threadId": "string (optional)",
  "groupId": "string (optional)",
  "groupChannel": "string (optional)",
  "groupSpace": "string (optional)",
  "timeout": "integer >= 0 (optional)",
  "lane": "string (optional, 'main' | 'cron' | 'subagent' | 'nested')",
  "extraSystemPrompt": "string (optional)",
  "inputProvenance": {
    "kind": "'external_user' | 'inter_session' | 'internal_system'",
    "sourceSessionKey": "string (optional)",
    "sourceChannel": "string (optional)",
    "sourceTool": "string (optional)"
  },
  "idempotencyKey": "string (required)",
  "label": "string 1-64 chars (optional)",
  "spawnedBy": "string (optional)"
}
```

**Immediate Response:**
```json
{
  "runId": "string",
  "status": "accepted",
  "acceptedAt": 0
}
```

**Async Completion (second response):**
```json
{
  "runId": "string",
  "status": "ok",
  "summary": "completed",
  "result": { "...agent command return..." }
}
```

**Async Error (second response):**
```json
{
  "runId": "string",
  "status": "error",
  "summary": "error message"
}
```

**Notes:** Two-phase. First respond returns immediately. Second fires on completion/failure.

---

#### `agent.identity.get`

> **Scope:** read

**Params:**
```json
{
  "agentId": "string (optional)",
  "sessionKey": "string (optional)"
}
```

**Response:**
```json
{
  "agentId": "string",
  "name": "string (optional)",
  "avatar": "string (optional, URL or path)",
  "emoji": "string (optional)"
}
```

---

#### `agent.wait`

> **Scope:** write

**Params:**
```json
{
  "runId": "string (required)",
  "timeoutMs": "integer >= 0 (optional, default 30000)"
}
```

**Response (completed):**
```json
{
  "runId": "string",
  "status": "ok | error",
  "startedAt": 0,
  "endedAt": 0,
  "error": "string (if error)"
}
```

**Response (timeout):**
```json
{
  "runId": "string",
  "status": "timeout"
}
```

---

### Agents (CRUD)

#### `agents.list`

> **Scope:** read

**Params:** `{}` (none)

**Response:**
```json
{
  "defaultId": "string",
  "mainKey": "string",
  "scope": "per-sender | global",
  "agents": [
    {
      "id": "string",
      "name": "string (optional)",
      "identity": {
        "name": "string (optional)",
        "theme": "string (optional)",
        "emoji": "string (optional)",
        "avatar": "string (optional)",
        "avatarUrl": "string (optional)"
      }
    }
  ]
}
```

---

#### `agents.create`

> **Scope:** admin

**Params:**
```json
{
  "name": "string (required)",
  "workspace": "string (required)",
  "emoji": "string (optional)",
  "avatar": "string (optional)"
}
```

**Response:**
```json
{
  "ok": true,
  "agentId": "string (normalized)",
  "name": "string",
  "workspace": "string (resolved path)"
}
```

---

#### `agents.update`

> **Scope:** admin

**Params:**
```json
{
  "agentId": "string (required)",
  "name": "string (optional)",
  "workspace": "string (optional)",
  "model": "string (optional)",
  "avatar": "string (optional)"
}
```

**Response:**
```json
{
  "ok": true,
  "agentId": "string"
}
```

---

#### `agents.delete`

> **Scope:** admin

**Params:**
```json
{
  "agentId": "string (required)",
  "deleteFiles": "boolean (optional)"
}
```

**Response:**
```json
{
  "ok": true,
  "agentId": "string",
  "removedBindings": 0
}
```

**Notes:** Cannot delete the default agent. Files are moved to trash by default.

---

### Agents Files

#### `agents.files.list`

> **Scope:** admin

**Params:**
```json
{
  "agentId": "string (required)"
}
```

**Response:**
```json
{
  "agentId": "string",
  "workspace": "string",
  "files": [
    {
      "name": "string (e.g. AGENTS.md, SOUL.md, TOOLS.md, ...)",
      "path": "string (full path)",
      "missing": false,
      "size": 1234,
      "updatedAtMs": 0
    }
  ]
}
```

**Notes:** Allowed files: `AGENTS.md`, `SOUL.md`, `TOOLS.md`, `IDENTITY.md`, `USER.md`, `HEARTBEAT.md`, `BOOTSTRAP.md`, `MEMORY.md`, `MEMORY_ALT.md`.

---

#### `agents.files.get`

> **Scope:** admin

**Params:**
```json
{
  "agentId": "string (required)",
  "name": "string (required)"
}
```

**Response:**
```json
{
  "agentId": "string",
  "workspace": "string",
  "file": {
    "name": "string",
    "path": "string",
    "missing": false,
    "size": 1234,
    "updatedAtMs": 0,
    "content": "string (full file contents)"
  }
}
```

---

#### `agents.files.set`

> **Scope:** admin

**Params:**
```json
{
  "agentId": "string (required)",
  "name": "string (required)",
  "content": "string (required)"
}
```

**Response:**
```json
{
  "ok": true,
  "agentId": "string",
  "workspace": "string",
  "file": {
    "name": "string",
    "path": "string",
    "missing": false,
    "size": 0,
    "updatedAtMs": 0,
    "content": "string"
  }
}
```

---

### Skills

#### `skills.status`

> **Scope:** read

**Params:**
```json
{
  "agentId": "string (optional, defaults to default agent)"
}
```

**Response:** Workspace skill status object (shape varies by installed skills).

---

#### `skills.bins`

> **Scope:** node role only

**Params:** `{}` (none)

**Response:**
```json
{
  "bins": ["string (sorted unique binary names)"]
}
```

---

#### `skills.install`

> **Scope:** admin

**Params:**
```json
{
  "name": "string (required)",
  "installId": "string (required)",
  "timeoutMs": "integer >= 1000 (optional)"
}
```

**Response:** Install result object (shape varies).

---

#### `skills.update`

> **Scope:** admin

**Params:**
```json
{
  "skillKey": "string (required)",
  "enabled": "boolean (optional)",
  "apiKey": "string (optional)",
  "env": { "<key>": "string (optional)" }
}
```

**Response:**
```json
{
  "ok": true,
  "skillKey": "string",
  "config": { "enabled": true, "apiKey": "...", "env": {} }
}
```

---

### Models

#### `models.list`

> **Scope:** read

**Params:** `{}` (none)

**Response:**
```json
{
  "models": [
    {
      "id": "string",
      "name": "string",
      "provider": "string",
      "contextWindow": 128000,
      "reasoning": true
    }
  ]
}
```

---

### Config

#### `config.get`

> **Scope:** read

**Params:** `{}` (none)

**Response:** Full config snapshot (sensitive values redacted). Includes `config`, `valid`, `raw`, `hash`, `path`.

---

#### `config.set`

> **Scope:** admin

**Params:**
```json
{
  "raw": "string (required, full YAML/JSON config)",
  "baseHash": "string (optional, for optimistic concurrency)"
}
```

**Response:**
```json
{
  "ok": true,
  "path": "string (config file path)",
  "config": { "...redacted validated config..." }
}
```

---

#### `config.apply`

> **Scope:** admin

**Params:**
```json
{
  "raw": "string (required, full YAML/JSON config)",
  "baseHash": "string (optional)",
  "sessionKey": "string (optional)",
  "note": "string (optional)",
  "restartDelayMs": "integer >= 0 (optional)"
}
```

**Response:**
```json
{
  "ok": true,
  "path": "string",
  "config": {},
  "restart": { "...restart schedule..." },
  "sentinel": { "path": "string", "payload": {} }
}
```

**Notes:** Same as `config.set` but also schedules a SIGUSR1 restart.

---

#### `config.patch`

> **Scope:** admin

**Params:**
```json
{
  "raw": "string (required, merge-patch YAML/JSON)",
  "baseHash": "string (optional)",
  "sessionKey": "string (optional)",
  "note": "string (optional)",
  "restartDelayMs": "integer >= 0 (optional)"
}
```

**Response:**
```json
{
  "ok": true,
  "path": "string",
  "config": {},
  "restart": {},
  "sentinel": { "path": "string", "payload": {} }
}
```

---

#### `config.schema`

> **Scope:** admin

**Params:** `{}` (none)

**Response:**
```json
{
  "schema": { "...JSON schema..." },
  "uiHints": { "<path>": { "label": "...", "help": "...", "group": "...", "sensitive": true } },
  "version": "string",
  "generatedAt": "string"
}
```

---

### Cron

#### `cron.list`

> **Scope:** read

**Params:**
```json
{
  "includeDisabled": "boolean (optional)"
}
```

**Response:**
```json
{
  "jobs": [
    {
      "id": "string",
      "agentId": "string",
      "name": "string",
      "description": "string",
      "enabled": true,
      "deleteAfterRun": false,
      "createdAtMs": 0,
      "updatedAtMs": 0,
      "schedule": { "kind": "cron", "expr": "*/5 * * * *", "tz": "UTC" },
      "sessionTarget": "main | isolated",
      "wakeMode": "next-heartbeat | now",
      "payload": { "kind": "agentTurn", "message": "..." },
      "delivery": { "mode": "none | announce", "channel": "last" },
      "state": {
        "nextRunAtMs": 0,
        "lastRunAtMs": 0,
        "lastStatus": "ok | error | skipped",
        "consecutiveErrors": 0
      }
    }
  ]
}
```

---

#### `cron.status`

> **Scope:** read

**Params:** `{}` (none)

**Response:** Cron system status object.

---

#### `cron.add`

> **Scope:** admin

**Params:**
```json
{
  "name": "string (required)",
  "agentId": "string | null (optional)",
  "description": "string (optional)",
  "enabled": "boolean (optional)",
  "deleteAfterRun": "boolean (optional)",
  "schedule": {
    "kind": "'at' | 'every' | 'cron'",
    "at": "string (for kind=at, ISO timestamp)",
    "everyMs": "integer (for kind=every)",
    "expr": "string (for kind=cron, cron expression)",
    "tz": "string (optional timezone)"
  },
  "sessionTarget": "'main' | 'isolated' (required)",
  "wakeMode": "'next-heartbeat' | 'now' (required)",
  "payload": {
    "kind": "'systemEvent' | 'agentTurn'",
    "message": "string",
    "model": "string (optional)",
    "thinking": "string (optional)",
    "timeoutSeconds": "integer (optional)",
    "deliver": "boolean (optional)"
  },
  "delivery": {
    "mode": "'none' | 'announce'",
    "channel": "string (optional)",
    "to": "string (optional)"
  }
}
```

**Response:** The created cron job object.

---

#### `cron.update`

> **Scope:** admin

**Params:**
```json
{
  "id": "string (required, or use jobId)",
  "jobId": "string (alternative to id)",
  "patch": { "...partial CronJob fields..." }
}
```

**Response:** The updated cron job object.

---

#### `cron.remove`

> **Scope:** admin

**Params:**
```json
{
  "id": "string (required, or use jobId)",
  "jobId": "string (alternative to id)"
}
```

**Response:** The removed cron job.

---

#### `cron.run`

> **Scope:** admin

**Params:**
```json
{
  "id": "string (required, or use jobId)",
  "jobId": "string (alternative to id)",
  "mode": "'due' | 'force' (optional, default 'force')"
}
```

**Response:** Run result object.

---

#### `cron.runs`

> **Scope:** read

**Params:**
```json
{
  "id": "string (required, or use jobId)",
  "jobId": "string (alternative to id)",
  "limit": "integer 1-5000 (optional)"
}
```

**Response:**
```json
{
  "entries": [
    {
      "ts": 0,
      "jobId": "string",
      "action": "finished",
      "status": "ok | error | skipped",
      "error": "string",
      "summary": "string",
      "sessionId": "string",
      "sessionKey": "string",
      "runAtMs": 0,
      "durationMs": 0,
      "nextRunAtMs": 0
    }
  ]
}
```

---

#### `wake`

> **Scope:** write

**Params:**
```json
{
  "mode": "'now' | 'next-heartbeat' (required)",
  "text": "string (required)"
}
```

**Response:** Wake result from cron engine.

---

### Nodes

#### `node.pair.request`

> **Scope:** pairing

**Params:**
```json
{
  "nodeId": "string (required)",
  "displayName": "string (optional)",
  "platform": "string (optional)",
  "version": "string (optional)",
  "coreVersion": "string (optional)",
  "uiVersion": "string (optional)",
  "deviceFamily": "string (optional)",
  "modelIdentifier": "string (optional)",
  "caps": ["string (optional)"],
  "commands": ["string (optional)"],
  "remoteIp": "string (optional)",
  "silent": "boolean (optional)"
}
```

**Response:**
```json
{
  "status": "pending",
  "request": { "requestId": "...", "nodeId": "...", "..." },
  "created": true
}
```

---

#### `node.pair.list`

> **Scope:** pairing

**Params:** `{}` (none)

**Response:**
```json
{
  "pending": ["<pairing requests>"],
  "paired": ["<paired node entries>"]
}
```

---

#### `node.pair.approve`

> **Scope:** pairing

**Params:**
```json
{
  "requestId": "string (required)"
}
```

**Response:**
```json
{
  "requestId": "string",
  "node": { "...approved node object..." }
}
```

---

#### `node.pair.reject`

> **Scope:** pairing

**Params:**
```json
{
  "requestId": "string (required)"
}
```

**Response:**
```json
{
  "requestId": "string",
  "nodeId": "string"
}
```

---

#### `node.pair.verify`

> **Scope:** pairing

**Params:**
```json
{
  "nodeId": "string (required)",
  "token": "string (required)"
}
```

**Response:** `{ "ok": true, ... }`

---

#### `node.rename`

> **Scope:** pairing

**Params:**
```json
{
  "nodeId": "string (required)",
  "displayName": "string (required)"
}
```

**Response:**
```json
{
  "nodeId": "string",
  "displayName": "string"
}
```

---

#### `node.list`

> **Scope:** read

**Params:** `{}` (none)

**Response:**
```json
{
  "ts": 0,
  "nodes": [
    {
      "nodeId": "string",
      "displayName": "string",
      "platform": "string",
      "version": "string",
      "coreVersion": "string",
      "uiVersion": "string",
      "deviceFamily": "string",
      "modelIdentifier": "string",
      "remoteIp": "string",
      "caps": ["string"],
      "commands": ["string"],
      "pathEnv": "string",
      "permissions": {},
      "connectedAtMs": 0,
      "paired": true,
      "connected": true
    }
  ]
}
```

---

#### `node.describe`

> **Scope:** read

**Params:**
```json
{
  "nodeId": "string (required)"
}
```

**Response:** Same shape as a single node entry (see `node.list`), plus `ts`.

---

#### `node.invoke`

> **Scope:** write

**Params:**
```json
{
  "nodeId": "string (required)",
  "command": "string (required)",
  "params": "unknown (optional)",
  "timeoutMs": "integer >= 0 (optional)",
  "idempotencyKey": "string (required)"
}
```

**Response:**
```json
{
  "ok": true,
  "nodeId": "string",
  "command": "string",
  "payload": { "...parsed result..." },
  "payloadJSON": "string (raw JSON)"
}
```

**Notes:** Blocked commands: `system.execApprovals.get`, `system.execApprovals.set` (use `exec.approvals.node.*`). Sanitizes `system.run` params.

---

#### `node.invoke.result`

> **Scope:** node role only

**Params:**
```json
{
  "id": "string (required)",
  "nodeId": "string (required)",
  "ok": "boolean (required)",
  "payload": "unknown (optional)",
  "payloadJSON": "string (optional)",
  "error": { "code": "string", "message": "string" }
}
```

**Response:**
```json
{
  "ok": true,
  "ignored": "boolean (true if late/unmatched)"
}
```

---

#### `node.event`

> **Scope:** node role only

**Params:**
```json
{
  "event": "string (required)",
  "payload": "unknown (optional)",
  "payloadJSON": "string (optional)"
}
```

**Response:** `{ "ok": true }`

---

### Devices

#### `device.pair.list`

> **Scope:** pairing

**Params:** `{}` (none)

**Response:**
```json
{
  "pending": [
    {
      "requestId": "string",
      "deviceId": "string",
      "publicKey": "string",
      "displayName": "string",
      "platform": "string",
      "clientId": "string",
      "clientMode": "string",
      "role": "string",
      "roles": ["string"],
      "scopes": ["string"],
      "remoteIp": "string",
      "silent": false,
      "isRepair": false,
      "ts": 0
    }
  ],
  "paired": [
    {
      "...device info...",
      "tokens": { "role": "...", "scopes": [], "createdAtMs": 0, "rotatedAtMs": 0, "revokedAtMs": 0, "lastUsedAtMs": 0 }
    }
  ]
}
```

---

#### `device.pair.approve`

> **Scope:** pairing

**Params:**
```json
{
  "requestId": "string (required)"
}
```

**Response:**
```json
{
  "requestId": "string",
  "device": { "...redacted device..." }
}
```

---

#### `device.pair.reject`

> **Scope:** pairing

**Params:**
```json
{
  "requestId": "string (required)"
}
```

**Response:**
```json
{
  "requestId": "string",
  "deviceId": "string"
}
```

---

#### `device.token.rotate`

> **Scope:** pairing

**Params:**
```json
{
  "deviceId": "string (required)",
  "role": "string (required)",
  "scopes": ["string (optional)"]
}
```

**Response:**
```json
{
  "deviceId": "string",
  "role": "string",
  "token": "string (the new token, NOT redacted)",
  "scopes": ["string"],
  "rotatedAtMs": 0
}
```

---

#### `device.token.revoke`

> **Scope:** pairing

**Params:**
```json
{
  "deviceId": "string (required)",
  "role": "string (required)"
}
```

**Response:**
```json
{
  "deviceId": "string",
  "role": "string",
  "revokedAtMs": 0
}
```

---

### Exec Approvals

#### `exec.approvals.get`

> **Scope:** admin

**Params:** `{}` (none)

**Response:**
```json
{
  "path": "string (exec approvals file path)",
  "exists": true,
  "hash": "string (content hash for concurrency)",
  "file": {
    "version": 1,
    "socket": { "path": "string" },
    "defaults": {
      "security": "string",
      "ask": "string",
      "askFallback": "string",
      "autoAllowSkills": true
    },
    "agents": {
      "<agentId>": {
        "security": "string",
        "ask": "string",
        "allowlist": [
          { "id": "string", "pattern": "string", "lastUsedAt": 0, "lastUsedCommand": "string" }
        ]
      }
    }
  }
}
```

**Notes:** `socket.token` is redacted (omitted) in the response.

---

#### `exec.approvals.set`

> **Scope:** admin

**Params:**
```json
{
  "file": { "...ExecApprovalsFile..." },
  "baseHash": "string (optional, for optimistic concurrency)"
}
```

**Response:** Same shape as `exec.approvals.get` (with updated content).

---

#### `exec.approvals.node.get`

> **Scope:** admin

**Params:**
```json
{
  "nodeId": "string (required)"
}
```

**Response:** Forwarded from node's `system.execApprovals.get`.

---

#### `exec.approvals.node.set`

> **Scope:** admin

**Params:**
```json
{
  "nodeId": "string (required)",
  "file": { "...ExecApprovalsFile..." },
  "baseHash": "string (optional)"
}
```

**Response:** Forwarded from node's `system.execApprovals.set`.

---

#### `exec.approval.request`

> **Scope:** approvals

**Params:**
```json
{
  "id": "string (optional)",
  "command": "string (required)",
  "cwd": "string | null (optional)",
  "host": "string | null (optional)",
  "security": "string | null (optional)",
  "ask": "string | null (optional)",
  "agentId": "string | null (optional)",
  "resolvedPath": "string | null (optional)",
  "sessionKey": "string | null (optional)",
  "timeoutMs": "integer >= 1 (optional)",
  "twoPhase": "boolean (optional)"
}
```

**Response (single phase, default):**
```json
{
  "id": "string",
  "decision": "allow-once | allow-always | deny",
  "createdAtMs": 0,
  "expiresAtMs": 0
}
```

**Response (twoPhase=true, immediate):**
```json
{
  "status": "accepted",
  "id": "string",
  "createdAtMs": 0,
  "expiresAtMs": 0
}
```

Then async second response with `decision` field.

**Notes:** Broadcasts `exec.approval.requested` event.

---

#### `exec.approval.resolve`

> **Scope:** approvals

**Params:**
```json
{
  "id": "string (required)",
  "decision": "string (required: 'allow-once' | 'allow-always' | 'deny')"
}
```

**Response:** `{ "ok": true }`

---

#### `exec.approval.waitDecision`

> **Scope:** approvals

Long-polls until the approval resolves or expires.

**Response:**
```json
{
  "id": "string",
  "decision": "string",
  "createdAtMs": 0,
  "expiresAtMs": 0
}
```

---

### Channels

#### `channels.status`

> **Scope:** read

**Params:**
```json
{
  "probe": "boolean (optional, actively probe channels)",
  "timeoutMs": "integer >= 0 (optional, default 10000)"
}
```

**Response:**
```json
{
  "ts": 0,
  "channelOrder": ["telegram", "discord", "whatsapp", "..."],
  "channelLabels": { "telegram": "Telegram", "..." },
  "channelDetailLabels": { "telegram": "Telegram Bot", "..." },
  "channelSystemImages": {},
  "channelMeta": [{ "id": "...", "label": "...", "detailLabel": "...", "systemImage": "..." }],
  "channels": { "telegram": { "...summary..." } },
  "channelAccounts": {
    "telegram": [
      {
        "accountId": "string",
        "name": "string",
        "enabled": true,
        "configured": true,
        "linked": true,
        "running": true,
        "connected": true,
        "reconnectAttempts": 0,
        "lastConnectedAt": 0,
        "lastError": "string",
        "lastInboundAt": 0,
        "lastOutboundAt": 0,
        "mode": "string",
        "dmPolicy": "string",
        "allowFrom": ["string"],
        "probe": {}
      }
    ]
  },
  "channelDefaultAccountId": { "telegram": "default" }
}
```

---

#### `channels.logout`

> **Scope:** admin

**Params:**
```json
{
  "channel": "string (required)",
  "accountId": "string (optional)"
}
```

**Response:**
```json
{
  "channel": "string",
  "accountId": "string",
  "cleared": true,
  "...plugin logout result..."
}
```

---

### Talk / TTS

#### `talk.config`

> **Scope:** read

**Params:**
```json
{
  "includeSecrets": "boolean (optional)"
}
```

**Response:**
```json
{
  "config": {
    "talk": {
      "voiceId": "string",
      "voiceAliases": { "alias": "voiceId" },
      "modelId": "string",
      "outputFormat": "string",
      "apiKey": "string (only if includeSecrets + admin scope)",
      "interruptOnSpeech": true
    },
    "session": { "mainKey": "string" },
    "ui": { "seamColor": "string" }
  }
}
```

---

#### `talk.mode`

> **Scope:** write

**Params:**
```json
{
  "enabled": "boolean (required)",
  "phase": "string (optional)"
}
```

**Response:**
```json
{
  "enabled": true,
  "phase": "string | null",
  "ts": 0
}
```

---

#### `tts.status`

> **Scope:** read

**Params:** (none)

**Response:**
```json
{
  "enabled": true,
  "auto": "...",
  "provider": "openai",
  "fallbackProvider": "edge",
  "fallbackProviders": ["edge"],
  "prefsPath": "string",
  "hasOpenAIKey": true,
  "hasElevenLabsKey": false,
  "edgeEnabled": true
}
```

---

#### `tts.providers`

> **Scope:** read

**Params:** (none)

**Response:**
```json
{
  "providers": [
    { "id": "openai", "name": "OpenAI", "configured": true, "models": ["tts-1", "tts-1-hd"], "voices": ["alloy", "echo", "fable", "onyx", "nova", "shimmer"] },
    { "id": "elevenlabs", "name": "ElevenLabs", "configured": false, "models": ["eleven_multilingual_v2", "eleven_turbo_v2_5", "eleven_monolingual_v1"] },
    { "id": "edge", "name": "Edge TTS", "configured": true, "models": [] }
  ],
  "active": "openai"
}
```

---

#### `tts.enable`

> **Scope:** write

**Response:** `{ "enabled": true }`

---

#### `tts.disable`

> **Scope:** write

**Response:** `{ "enabled": false }`

---

#### `tts.convert`

> **Scope:** write

**Params:** `{ "text": "string", "channel": "string (optional)" }`

**Response:**
```json
{
  "audioPath": "string",
  "provider": "string",
  "outputFormat": "string",
  "voiceCompatible": true
}
```

---

#### `tts.setProvider`

> **Scope:** write

**Params:** `{ "provider": "string" }`

**Response:** `{ "provider": "openai" }`

---

### Logs

#### `logs.tail`

> **Scope:** read

**Params:**
```json
{
  "cursor": "integer >= 0 (optional, byte offset for polling)",
  "limit": "integer 1-5000 (optional, default 500)",
  "maxBytes": "integer 1-1000000 (optional, default 250000)"
}
```

**Response:**
```json
{
  "file": "string (log file path)",
  "cursor": 12345,
  "size": 67890,
  "lines": ["log line 1", "log line 2"],
  "truncated": false,
  "reset": false
}
```

**Notes:** Supports cursor-based incremental polling. Auto-resolves rolling log files (`openclaw-YYYY-MM-DD.log`).

---

### Usage

#### `usage.status`

> **Scope:** read

**Params:** (none)

**Response:** Provider usage summary object.

---

#### `usage.cost`

> **Scope:** read

**Params:** `{ "startDate": "string", "endDate": "string", "days": 30 }`

**Response:**
```json
{
  "updatedAt": 0,
  "days": 30,
  "daily": [
    { "date": "YYYY-MM-DD", "input": 0, "output": 0, "totalTokens": 0, "totalCost": 0.0 }
  ],
  "totals": { "input": 0, "output": 0, "totalTokens": 0, "totalCost": 0.0 }
}
```

---

#### `sessions.usage.timeseries`

> **Scope:** admin | **Note:** Not in BASE_METHODS

**Params:** `{ "key": "string", "maxPoints": 200 }`

**Response:**
```json
{
  "sessionId": "string",
  "points": [
    { "timestamp": 0, "input": 0, "output": 0, "totalTokens": 0, "cost": 0.0, "cumulativeTokens": 0, "cumulativeCost": 0.0 }
  ]
}
```

---

#### `sessions.usage.logs`

> **Scope:** admin | **Note:** Not in BASE_METHODS

**Params:** `{ "key": "string", "limit": 200 }`

**Response:**
```json
{
  "logs": [
    { "timestamp": 0, "role": "user | assistant | tool | toolResult", "content": "string (max 2000 chars)", "tokens": 0, "cost": 0.0 }
  ]
}
```

---

### Wizard

#### `wizard.start`

> **Scope:** admin

**Params:**
```json
{
  "mode": "'local' | 'remote' (optional)",
  "workspace": "string (optional)"
}
```

**Response:**
```json
{
  "sessionId": "string (UUID)",
  "done": false,
  "step": {
    "id": "string",
    "type": "note | select | multiselect | text | confirm | progress | action",
    "title": "string",
    "message": "string",
    "options": [{ "value": "...", "label": "...", "hint": "..." }],
    "initialValue": "...",
    "placeholder": "string",
    "sensitive": false,
    "executor": "client | gateway"
  },
  "status": "running",
  "error": null
}
```

---

#### `wizard.next`

> **Scope:** admin

**Params:**
```json
{
  "sessionId": "string (required)",
  "answer": {
    "stepId": "string (required)",
    "value": "unknown (optional)"
  }
}
```

**Response:**
```json
{
  "done": false,
  "step": { "...next step..." },
  "status": "running",
  "error": null
}
```

---

#### `wizard.cancel`

> **Scope:** admin

**Params:**
```json
{
  "sessionId": "string (required)"
}
```

**Response:**
```json
{
  "status": "cancelled",
  "error": null
}
```

---

#### `wizard.status`

> **Scope:** admin

**Params:**
```json
{
  "sessionId": "string (required)"
}
```

**Response:**
```json
{
  "status": "running | done | cancelled | error",
  "error": "string (optional)"
}
```

---

### Update

#### `update.run`

> **Scope:** admin

**Params:**
```json
{
  "sessionKey": "string (optional)",
  "note": "string (optional)",
  "restartDelayMs": "integer >= 0 (optional)",
  "timeoutMs": "integer >= 1 (optional)"
}
```

**Response:**
```json
{
  "ok": true,
  "result": {
    "status": "ok | error | skipped",
    "mode": "string",
    "root": "string (package root)",
    "before": "string (version before)",
    "after": "string (version after)",
    "steps": [
      {
        "name": "string",
        "command": "string",
        "cwd": "string",
        "durationMs": 0,
        "log": { "stdoutTail": "string", "stderrTail": "string", "exitCode": 0 }
      }
    ],
    "reason": "string",
    "durationMs": 0
  },
  "restart": {},
  "sentinel": { "path": "string", "payload": {} }
}
```

---

### Web Login

#### `web.login.start`

> **Scope:** admin (via `config.*` prefix match)

**Params:**
```json
{
  "force": "boolean (optional)",
  "timeoutMs": "integer >= 0 (optional)",
  "verbose": "boolean (optional)",
  "accountId": "string (optional)"
}
```

**Response:** Channel plugin-specific (e.g., QR code data for WhatsApp).

---

#### `web.login.wait`

> **Scope:** admin

**Params:**
```json
{
  "timeoutMs": "integer >= 0 (optional)",
  "accountId": "string (optional)"
}
```

**Response:** `{ "connected": true, "..." }`

---

### System

#### `health`

> **Scope:** read

**Response:** Full health snapshot object.

---

#### `status`

> **Scope:** read

**Response:** Status summary object.

---

#### `system-presence`

> **Scope:** read

**Response:** Current system presence list.

---

#### `last-heartbeat`

> **Scope:** read

**Response:** Last heartbeat event object.

---

#### `set-heartbeats`

> **Scope:** admin

**Params:** `{ "enabled": true }`

**Response:** `{ "ok": true, "enabled": true }`

---

#### `system-event`

> **Scope:** admin

**Params:** System event with presence metadata (deviceId, host, ip, version, platform, etc.)

**Response:** `{ "ok": true }`

---

#### `voicewake.get`

> **Scope:** read

**Response:** `{ "triggers": ["hey openclaw", "..."] }`

---

#### `voicewake.set`

> **Scope:** write

**Params:** `{ "triggers": ["string"] }`

**Response:** `{ "triggers": ["string (normalized)"] }`

---

### Send / Poll

#### `send`

> **Scope:** write

**Params:**
```json
{
  "to": "string (required)",
  "message": "string (optional)",
  "mediaUrl": "string (optional)",
  "mediaUrls": ["string (optional)"],
  "gifPlayback": "boolean (optional)",
  "channel": "string (optional)",
  "accountId": "string (optional)",
  "sessionKey": "string (optional)",
  "idempotencyKey": "string (required)"
}
```

**Response:**
```json
{
  "runId": "string",
  "messageId": "string",
  "channel": "string",
  "chatId": "string (optional)",
  "channelId": "string (optional)",
  "toJid": "string (optional)",
  "conversationId": "string (optional)"
}
```

---

#### `poll`

> **Scope:** admin | **Note:** Not in BASE_METHODS

**Params:**
```json
{
  "to": "string (required)",
  "question": "string (required)",
  "options": ["string (2-12 items)"],
  "maxSelections": "integer 1-12 (optional)",
  "durationSeconds": "integer 1-604800 (optional)",
  "durationHours": "integer >= 1 (optional)",
  "silent": "boolean (optional)",
  "isAnonymous": "boolean (optional)",
  "threadId": "string (optional)",
  "channel": "string (optional)",
  "accountId": "string (optional)",
  "idempotencyKey": "string (required)"
}
```

**Response:**
```json
{
  "runId": "string",
  "messageId": "string",
  "channel": "string",
  "pollId": "string (optional)"
}
```

---

### Browser

#### `browser.request`

> **Scope:** write

**Params:** Browser control request (varies by action).

**Response:** Browser control result body (shape varies).

---

## WebSocket Event Stream

### Envelope Format

Every agent event is wrapped in this envelope:

```json
{
  "runId": "string",
  "seq": 0,
  "stream": "lifecycle | tool | assistant | compaction | thinking",
  "ts": 0,
  "data": { "...stream-specific..." },
  "sessionKey": "string (optional)"
}
```

Events are delivered via the `agent-event` WebSocket event type:
```json
{
  "type": "event",
  "event": "agent-event",
  "payload": { "...AgentEventPayload..." }
}
```

---

### Stream: `lifecycle`

Agent run start/end/error.

#### Phase: `start`
```json
{
  "stream": "lifecycle",
  "data": {
    "phase": "start",
    "startedAt": 1709312400000
  }
}
```

#### Phase: `end`
```json
{
  "stream": "lifecycle",
  "data": {
    "phase": "end",
    "endedAt": 1709312500000,
    "startedAt": 1709312400000
  }
}
```

#### Phase: `error`
```json
{
  "stream": "lifecycle",
  "data": {
    "phase": "error",
    "error": "LLM request failed.",
    "endedAt": 1709312500000,
    "startedAt": 1709312400000
  }
}
```

**Transitions:** `start` -> `end` (success) or `start` -> `error` (failure)

---

### Stream: `tool`

Tool execution lifecycle.

#### Phase: `start`
```json
{
  "stream": "tool",
  "data": {
    "phase": "start",
    "name": "exec",
    "toolCallId": "toolu_abc123",
    "args": { "command": "ls -la" }
  }
}
```

#### Phase: `update`
```json
{
  "stream": "tool",
  "data": {
    "phase": "update",
    "name": "exec",
    "toolCallId": "toolu_abc123",
    "partialResult": "partial output..."
  }
}
```

#### Phase: `result`
```json
{
  "stream": "tool",
  "data": {
    "phase": "result",
    "name": "exec",
    "toolCallId": "toolu_abc123",
    "meta": "ls -la",
    "isError": false,
    "result": { "stdout": "...", "exitCode": 0 }
  }
}
```

**Transitions:** `start` -> `update`* -> `result`

**Notes:**
- `args` is only included in the full event, NOT in the callback version.
- `result` is sanitized: image `data` fields are stripped, text truncated.
- `meta` is inferred from args (e.g., file path for `read`, command for `exec`).

---

### Stream: `assistant`

Streaming text from the LLM.

```json
{
  "stream": "assistant",
  "data": {
    "text": "Here is the full text so far...",
    "delta": "so far...",
    "mediaUrls": ["https://example.com/image.png"]
  }
}
```

**Notes:**
- `text` accumulates (full cleaned text so far).
- `delta` is the incremental portion since the last event.
- `mediaUrls` are extracted from reply directives (`[[media:...]]`).
- Thinking/final tags are stripped from text.

---

### Stream: `compaction`

Auto-compaction of context window.

#### Phase: `start`
```json
{
  "stream": "compaction",
  "data": { "phase": "start" }
}
```

#### Phase: `end`
```json
{
  "stream": "compaction",
  "data": {
    "phase": "end",
    "willRetry": false
  }
}
```

**Notes:** If `willRetry` is true, another compaction cycle will follow.

---

### Stream: `thinking`

Extended thinking / reasoning output (when `reasoningMode === "stream"`).

```json
{
  "stream": "thinking",
  "data": {
    "text": "Full reasoning text so far...",
    "delta": "so far..."
  }
}
```

---

### Chat Events

Broadcast events for `chat.send` completion. Delivered as:

```json
{
  "type": "event",
  "event": "chat",
  "payload": {
    "runId": "string",
    "sessionKey": "string",
    "seq": 0,
    "state": "delta | final | aborted | error",
    "message": { "...message content..." },
    "errorMessage": "string (if error)",
    "usage": { "input": 0, "output": 0, "total": 0 },
    "stopReason": "string"
  }
}
```

**States:**
- `delta` - Streaming partial updates during agent execution
- `final` - Agent completed successfully
- `aborted` - Agent was aborted by user
- `error` - Agent encountered an error

---

### Broadcast Events

Other broadcast events the server may push:

| Event | Payload | When |
|---|---|---|
| `agent-event` | `AgentEventPayload` (see above) | During agent runs |
| `chat` | `ChatEventSchema` | Chat completion updates |
| `tick` | `{ "ts": 0 }` | Periodic keepalive |
| `shutdown` | `{ "reason": "string", "restartExpectedMs": 0 }` | Server shutting down |
| `presence` | `PresenceEntry` | Presence changes |
| `health` | Health snapshot | Health state changes |
| `node.pair.requested` | Pairing request | New node pairing request |
| `node.pair.resolved` | `{ "requestId", "nodeId", "decision", "ts" }` | Node pairing resolved |
| `device.pair.requested` | Device pairing request | New device pairing request |
| `device.pair.resolved` | `{ "requestId", "deviceId", "decision", "ts" }` | Device pairing resolved |
| `exec.approval.requested` | `{ "id", "request", "createdAtMs", "expiresAtMs" }` | New exec approval |
| `exec.approval.resolved` | `{ "id", "decision", "resolvedBy", "ts" }` | Exec approval resolved |
| `talk.mode` | `{ "enabled", "phase", "ts" }` | Talk mode toggled |
| `node.invoke.request` | `{ "id", "nodeId", "command", "paramsJSON", "timeoutMs" }` | Node invoke forwarded to node |

---

### Subagent Lifecycle

Subagents are spawned via the `sessions_spawn` tool during agent execution.

**Session key format:** `agent:<targetAgentId>:subagent:<UUID>`

**Subagent Run Record:**
```json
{
  "runId": "string",
  "childSessionKey": "string",
  "requesterSessionKey": "string",
  "requesterDisplayKey": "string",
  "task": "string",
  "cleanup": "delete | keep",
  "label": "string",
  "createdAt": 0,
  "startedAt": 0,
  "endedAt": 0,
  "outcome": {
    "status": "ok | error | timeout | unknown",
    "error": "string (if error)"
  }
}
```

**Lifecycle flow:**
1. `lifecycle.start` event emitted with subagent's `runId` and `sessionKey`
2. `tool.*` and `assistant` events stream during execution
3. `lifecycle.end` or `lifecycle.error` emitted on completion
4. Announce flow: reads final reply, builds stats, sends summary back to requester session
5. Cleanup: session is either deleted or archived with TTL

**Command lanes:**
```
main       - Primary user conversation
cron       - Cron-triggered runs
subagent   - Spawned subagent runs
nested     - Nested agent calls
```

---

### Plugin Hook Events

Lifecycle hooks available to plugins (not WebSocket events, but gateway-internal):

| Hook | When | Payload |
|---|---|---|
| `message_received` | Inbound message arrives | Message metadata |
| `before_agent_start` | Before LLM call | Session info |
| `agent_end` | After LLM call | Result info |
| `before_tool_call` | Before tool execution | `{ toolName, params }` |
| `after_tool_call` | After tool execution | `{ toolName, params, result, error?, durationMs }` |
| `before_compaction` | Before auto-compaction | `{ messageCount }` |
| `after_compaction` | After auto-compaction | `{ messageCount, compactedCount }` |
| `tool_result_persist` | Transform tool result for storage | `{ message, meta }` |
| `before_reset` | Before session reset | `{ messages }` |
| `session_start` | Session initialization | Session info |
| `session_end` | Session teardown | Session info |

---

## Complete Event Taxonomy

```
AgentEventPayload
  |
  |-- stream: "lifecycle"
  |     |-- data.phase: "start"    { startedAt }
  |     |-- data.phase: "end"      { endedAt, startedAt? }
  |     |-- data.phase: "error"    { error, endedAt, startedAt? }
  |
  |-- stream: "tool"
  |     |-- data.phase: "start"    { name, toolCallId, args }
  |     |-- data.phase: "update"   { name, toolCallId, partialResult }
  |     |-- data.phase: "result"   { name, toolCallId, meta?, isError, result }
  |
  |-- stream: "assistant"
  |     |-- (no phase)             { text, delta, mediaUrls? }
  |
  |-- stream: "compaction"
  |     |-- data.phase: "start"    {}
  |     |-- data.phase: "end"      { willRetry }
  |
  |-- stream: "thinking"
        |-- (no phase)             { text, delta }
```

All events share the envelope: `{ runId, seq, stream, ts, data, sessionKey? }`

---

## Appendix: Advertised Methods

Methods returned in `hello-ok.features.methods` (BASE_METHODS):

```
health, logs.tail, channels.status, channels.logout, status, usage.status,
usage.cost, tts.status, tts.providers, tts.enable, tts.disable, tts.convert,
tts.setProvider, config.get, config.set, config.apply, config.patch, config.schema,
exec.approvals.get, exec.approvals.set, exec.approvals.node.get,
exec.approvals.node.set, exec.approval.request, exec.approval.waitDecision,
exec.approval.resolve, wizard.start, wizard.next, wizard.cancel, wizard.status,
talk.config, talk.mode, models.list, agents.list, agents.create, agents.update,
agents.delete, agents.files.list, agents.files.get, agents.files.set, skills.status,
skills.bins, skills.install, skills.update, update.run, voicewake.get, voicewake.set,
sessions.list, sessions.preview, sessions.patch, sessions.reset, sessions.delete,
sessions.compact, last-heartbeat, set-heartbeats, wake, node.pair.request,
node.pair.list, node.pair.approve, node.pair.reject, node.pair.verify,
device.pair.list, device.pair.approve, device.pair.reject, device.token.rotate,
device.token.revoke, node.rename, node.list, node.describe, node.invoke,
node.invoke.result, node.event, cron.list, cron.status, cron.add, cron.update,
cron.remove, cron.run, cron.runs, system-presence, system-event, send, agent,
agent.identity.get, agent.wait, browser.request, chat.history, chat.abort, chat.send
```

**Unadvertised but functional methods:**
```
poll, chat.inject, sessions.resolve, sessions.usage, sessions.usage.timeseries,
sessions.usage.logs, connect (always errors post-handshake)
```
