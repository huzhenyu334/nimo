# ACP Executor Reference

Complete input/output schema for every executor.

## agent — AI Agent Execution

**Commands**: `chat` (default), `review`, `structured_output`

```yaml
- id: analyze
  executor: agent
  input:
    prompt: "Analyze this: {{input}}"           # required — prompt text
    assignee:                                    # optional — specific agent
      type: agent
      id: "main"
    model: "claude-opus-4-6"                       # optional — LLM override
    max_tokens: 4096                             # optional
    tools: ["read", "exec", "web_search"]        # optional — tool whitelist
    credentials: ["github-token"]                # optional — credential slugs
    data_schema:                                 # optional — structured output constraint
      type: object
      properties:
        score: { type: number }
```

**Output**: `output.text` (full text), `output.data.*` (parsed from `[output:key=value]` tags)

---

## human — Human Tasks

### form command (default)

```yaml
- id: review
  executor: human
  command: form
  input:
    title: "Review Request"               # REQUIRED
    prompt: "Please review..."            # optional — description
    assignee: "ou_xxxxx"                  # optional — Feishu open_id
    due: "3d"                             # optional — deadline
    form:                                 # optional — form fields
      output_schema:
        type: object
        properties:
          approved: { type: boolean, title: "Approve?" }
          notes: { type: string, title: "Notes" }
    data_schema: {}                       # optional — output validation
```

### approval command

```yaml
- id: approve
  executor: human
  command: approval
  input:
    title: "Approve deployment?"          # REQUIRED
    assignee: "ou_xxxxx"                  # optional
    form:                                 # optional — additional fields
      output_schema:
        type: object
        properties:
          reason: { type: string }
```

**Output**: `output.data.approved` (boolean), `output.data.reason`, plus any form fields

### notification command

```yaml
- id: notify
  executor: human
  command: notification
  input:
    title: "Build Complete"               # REQUIRED
    prompt: "Version 2.1 deployed"        # optional
    assignee: "ou_xxxxx"                  # optional
```

**Output**: notification is fire-and-forget, no meaningful output

---

## nocodb — NocoDB Structured Data (Script Executor)

**Commands**: `search`, `list_records`, `get_record`, `create_record`, `update_record`, `delete_record`

All commands take `command` inside `input`:

```yaml
- id: find
  executor: nocodb
  input:
    command: search
    table: 软件项目                      # table alias or ID
    where: "(项目名称,eq,{{input}})"     # NocoDB where syntax

- id: create
  executor: nocodb
  input:
    command: create_record
    table: 文档注册表
    fields: '{"文档标题":"test","类型":"PRD","状态":"草稿"}'

- id: update
  executor: nocodb
  input:
    command: update_record
    table: 软件项目
    record_id: "{{steps.find.output.data.records[0].Id}}"
    fields: '{"状态":"进行中"}'

- id: list
  executor: nocodb
  input:
    command: list_records
    table: 软件项目
    limit: "50"
    sort: "-CreatedAt"
```

**Table aliases**: `软件项目`/`projects` → `mk4cbfl27dbfhlu`, `文档注册表`/`documents` → `mjap0r8h9nxrb67`

**Output**: `output.data.records` (array), `output.data.total` (count)

---

## shell — Shell Commands

**Commands**: `exec` (default), `script`

```yaml
- id: build
  executor: shell
  input:
    command: "go build -o bin/plm ./cmd/plm/"   # required
    workdir: "/home/claw/.openclaw/workspace"   # optional
    timeout: "5m"                                # optional (default 5m)
```

**Output**: `output.data.stdout`, `output.data.stderr`, `output.data.exit_code`

---

## http — HTTP Requests

**Commands**: `request` (default)

```yaml
- id: call_api
  executor: http
  input:
    url: "https://api.example.com/data"         # required
    method: "POST"                               # GET/POST/PUT/DELETE/PATCH (default GET)
    headers:
      Authorization: "Bearer {{token}}"
      Content-Type: "application/json"
    body: '{"key": "value"}'                     # optional
    timeout: "30s"                               # optional
    retries: 2                                   # optional
```

**Output**: `output.data.status_code`, `output.data.body`, `output.data.headers`

---

## knowledge — Knowledge Base

**Commands**: `fetch`, `search`, `fetch_section`

```yaml
- id: load_doc
  executor: knowledge
  command: fetch
  input:
    doc_ids: ["doc-uuid-1", "doc-uuid-2"]       # required
    max_length: 50000                            # optional
    format: "full"                               # full/summary

- id: search_kb
  executor: knowledge
  command: search
  input:
    query: "PLM architecture"                    # required
    domain: "engineering"                        # required
    limit: 3                                     # optional
    format: "summary"                            # full/summary
```

**Output**: `output.data.contents` (merged markdown), `output.data.documents` (metadata array)

---

## timer — Delays and Scheduling

**Commands**: `delay`, `until`

```yaml
- id: wait
  executor: timer
  command: delay
  input:
    duration: "1h30m"                            # Go duration format

- id: schedule
  executor: timer
  command: until
  input:
    time: "2026-03-10T09:00:00+08:00"           # RFC3339
```

---

## subprocess — Child Process

**Commands**: `run` (default)

```yaml
- id: run_child
  executor: subprocess
  input:
    process: "child-process-name"                # required — process name or ID
    sub_input: "{{input}}"                       # optional — input to child
```

**Output**: `output.data.child_instance_id`

---

## calculator — Math Expressions

**Commands**: `eval` (default), `convert`, `compare`

```yaml
- id: calc
  executor: calculator
  input:
    expression: "(2 + 3) * 4"                    # required
```
