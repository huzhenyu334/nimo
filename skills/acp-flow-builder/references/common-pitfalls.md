# Common Pitfalls & Fixes

Real errors encountered during flow development. Check these BEFORE publishing.

## ❌ Field Not Found Errors

### `field input_label not found in type service.WorkflowDef`

**Cause**: Using non-existent top-level fields like `input_label`, `input_placeholder`.
**Fix**: Remove them. WorkflowDef only has `description` and `steps`.

```yaml
# ❌ Wrong
input_label: "项目名称"
input_placeholder: "例如：PLM"

# ✅ Right — just omit these fields
description: "项目立项流程"
steps: [...]
```

### `field assignee not found in type service.StepDef`

**Cause**: Putting `assignee` or `assign` as a top-level step field.
**Fix**: Move `assignee` inside `input`.

```yaml
# ❌ Wrong
- id: review
  executor: human
  assignee:
    type: human
    id: "ou_xxx"
  input:
    title: "Review"

# ✅ Right
- id: review
  executor: human
  command: form
  input:
    title: "Review"
    assignee: "ou_xxx"
```

### `field assign not found in type service.StepDef`

Same as above — `assign` doesn't exist either. Use `input.assignee`.

## ❌ Required Field Errors

### `required input field "title" is missing for executor "human"`

**Cause**: human executor requires `title` in input for ALL commands.
**Fix**: Add `title` to input.

```yaml
# ❌ Wrong
- id: review
  executor: human
  input:
    prompt: "Please review"
    assignee: "ou_xxx"

# ✅ Right
- id: review
  executor: human
  command: form
  input:
    title: "Review Request"       # ← REQUIRED
    prompt: "Please review"
    assignee: "ou_xxx"
```

## ❌ Executor vs Control Confusion

### Using `executor: switch` or `executor: gate`

**Cause**: switch and gate are NOT executors — they're control node types.
**Fix**: Use `control: switch` or `control: if`.

```yaml
# ❌ Wrong
- id: route
  executor: switch
  input:
    expression: "..."
    cases:
      "0": step_a

# ✅ Right
- id: route
  control: switch
  expression: "{{steps.prev.output.data.total}}"
  branches:
    "0": step_a
    default: step_b
  depends_on: [prev]
```

### Using `executor: gate`

```yaml
# ❌ Wrong
- id: check
  executor: gate
  input:
    condition: "..."
    pass_to: yes
    fail_to: no

# ✅ Right
- id: check
  control: if
  condition: "{{steps.prev.output.data.approved}} == true"
  branches:
    "true": yes
    "false": no
  depends_on: [prev]
```

## ❌ Branch Key Naming

### Using `cases` instead of `branches`

```yaml
# ❌ Wrong
  cases:
    "0": new_project

# ✅ Right
  branches:
    "0": new_project
    default: existing_project
```

### Using `pass_to`/`fail_to` instead of `branches`

```yaml
# ❌ Wrong (gate-style)
  pass_to: complete
  fail_to: rejected

# ✅ Right (if-style)
  branches:
    "true": complete
    "false": rejected
```

## ❌ Script Executor Command Placement

For script executors (nocodb, cc, echo), `command` goes in `input`, not as top-level:

```yaml
# Both work for built-in executors:
- id: step1
  executor: human
  command: form          # ← top-level OK for built-in executors
  input: { title: "..." }

# For script executors, command goes in input:
- id: step2
  executor: nocodb
  input:
    command: search      # ← inside input for script executors
    table: 软件项目
```

## ❌ Variable Reference Errors

### Referencing non-existent output fields

Agent steps output `[output:key=value]` tags → accessible as `output.data.key`.
If the agent doesn't output the tag, the variable will be empty.

**Best practice**: Document expected `[output:...]` tags clearly in agent prompts.

### Array index on potentially empty results

```yaml
# ⚠️ Risky — crashes if search returns 0 records
record_id: "{{steps.search.output.data.records[0].Id}}"

# Better — add a control node to check results first
- id: check_found
  control: if
  condition: "{{steps.search.output.data.total}} > 0"
  branches:
    "true": process_result
    "false": handle_not_found
```

## ❌ Subprocess Process Field

subprocess的`process`字段需要**同时放在顶层和input里**（两个验证逻辑不一致）：

```yaml
# ❌ Wrong — 只放顶层
- id: run_scan
  executor: subprocess
  process: "代码扫描"
  input:
    sub_input: "{{input}}"

# ❌ Wrong — 只放input
- id: run_scan
  executor: subprocess
  input:
    process: "代码扫描"
    sub_input: "{{input}}"

# ✅ Right — 两边都放
- id: run_scan
  executor: subprocess
  process: "代码扫描"
  input:
    process: "代码扫描"
    sub_input: "{{input}}"
```

## ❌ Human Form 用 form.fields，不用 output_schema

human executor 的表单定义用 `form.fields`（FieldDef数组），**不是** `form.output_schema`（JSON Schema）：

```yaml
# ❌ Wrong — output_schema 是给 agent 用的
form:
  output_schema:
    type: object
    properties:
      name: { type: string, title: "姓名" }

# ✅ Right — human 用 form.fields
form:
  fields:
    - key: name
      label: "姓名"
      type: text
      required: true
      placeholder: "请输入姓名"
```

FieldDef支持的type：`text | textarea | number | select | multi_select | date | switch`

select/multi_select 需要 options：`options: [{ value: "x", label: "X" }]`

## Pre-Publish Checklist

Before calling `acp_publish_process`, verify:

- [ ] No top-level fields except `description` and `steps`
- [ ] Every human step has `title` in input
- [ ] `assignee` is inside `input`, not a top-level step field
- [ ] Control nodes use `control:` not `executor:`
- [ ] Branches use `branches:` not `cases:` or `pass_to:`/`fail_to:`
- [ ] All `depends_on` reference existing step IDs
- [ ] Variable references use correct path (steps.X.output.data.Y)
- [ ] Each step has either `executor` or `control`, never both
