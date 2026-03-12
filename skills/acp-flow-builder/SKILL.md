---
name: acp-flow-builder
description: Build, validate, and publish ACP workflow YAML definitions. Use when creating new ACP processes/flows, debugging YAML validation errors, choosing the right executor for a step, or designing multi-step workflows with branching, human approval, NocoDB operations, or agent tasks. Triggers on "创建流程", "build flow", "workflow YAML", "ACP process", "流程定义", "publish process".
---

# ACP Flow Builder

Build ACP workflow YAML that publishes on the first try.

## Workflow

1. **Clarify requirements** — Understand what the flow does, what inputs it takes, what outputs it produces
2. **Choose executors** — Pick the right executor for each step (see `references/executor-reference.md`)
3. **Design step graph** — Map dependencies with `depends_on`, add control nodes (if/switch/loop) as needed
4. **Write YAML** — Follow StepDef schema exactly (see `references/stepdef-reference.md`)
5. **Validate mentally** — Check against common pitfalls (see `references/common-pitfalls.md`)
6. **Create + Publish** — Use `acp_create_process` then `acp_publish_process`
7. **If publish fails** — Read error message, fix, `acp_update_process`, re-publish

## Quick Reference: StepDef Top-Level Fields

```yaml
- id: step_id              # required, unique
  name: "Display Name"     # optional
  executor: agent           # executor type (agent/human/nocodb/shell/http/knowledge/timer/subprocess/calculator)
  command: chat             # executor sub-command (optional, defaults to first command)
  control: switch           # control node type (if/switch/loop/foreach/terminate) — mutually exclusive with executor
  input: {}                 # executor-specific input fields
  depends_on: [prev_step]   # dependency list
  timeout: "30m"            # step timeout
  on_failure: abort         # abort/skip/pause
  gate: {}                  # post-execution gate
  output_schema: {}         # validate step output
```

**Key rule**: `executor` and `control` are mutually exclusive. A step is either an executor step OR a control node.

## Control Nodes

```yaml
# IF node
- id: check
  control: if
  condition: "{{steps.prev.output.data.approved}} == true"
  branches:
    "true": next_step
    "false": other_step
  depends_on: [prev]

# SWITCH node
- id: route
  control: switch
  expression: "{{steps.prev.output.data.type}}"
  branches:
    "A": step_a
    "B": step_b
    default: step_default
  depends_on: [prev]
```

## Variable References

- `{{input}}` — Flow input string
- `{{steps.STEP_ID.output.text}}` — Step's text output
- `{{steps.STEP_ID.output.data.FIELD}}` — Step's structured data field
- `{{steps.STEP_ID.output.data.records[0].Id}}` — Array access

## Essential Rules

1. **human executor always needs `title` in input** — It's required for all 3 commands (form/approval/notification)
2. **human executor `assignee` goes in `input`**, not as a top-level StepDef field
3. **Control nodes use `control:` field** — Not `executor: switch` or `executor: gate`
4. **`branches` not `cases`** — Switch/if use `branches` map
5. **Script executors (nocodb, cc, etc.)** — Put `command` in `input`, not as top-level `command`... UNLESS using built-in executors where `command` is top-level
6. **No `input_label`/`input_placeholder`** — These are NOT WorkflowDef fields

## References

| File | When to read |
|------|-------------|
| `references/design-guide.md` | **First** — 设计方法论、六步法、命名规范、复审清单 |
| `references/executor-reference.md` | 写YAML时 — 所有executor的input/output schema |
| `references/stepdef-reference.md` | 写YAML时 — StepDef全字段、控制节点语法 |
| `references/common-pitfalls.md` | 发布前 — 常见错误和pre-publish checklist |
| `references/examples.md` | 需要参考时 — 3个验证过的完整流程 |
