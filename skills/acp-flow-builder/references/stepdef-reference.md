# StepDef Complete Field Reference

All fields available on a step definition, extracted from Go source.

## All Fields

| Field | YAML Key | Type | Description |
|-------|----------|------|-------------|
| ID | `id` | string | **Required.** Unique step identifier |
| Name | `name` | string | Display name |
| Description | `description` | string | Step description |
| Executor | `executor` | string | Executor type (agent/human/nocodb/shell/http/knowledge/timer/subprocess/calculator) |
| Command | `command` | string | Executor sub-command (e.g. `form`, `approval`, `search`, `fetch`) |
| Control | `control` | string | Control node type: `if`, `switch`, `loop`, `foreach`, `terminate` |
| DependsOn | `depends_on` | []string | List of step IDs this step depends on |
| Input | `input` | map | Executor-specific input fields |
| Timeout | `timeout` | string | Step timeout (default "30m") |
| Retry | `retry` | object | `{max_attempts: N, delay: "5s"}` |
| OnFailure | `on_failure` | string | `abort` (default) / `skip` / `pause` |
| Gate | `gate` | object | Post-execution gate condition |
| OutputSchema | `output_schema` | object | JSON Schema to validate step output |
| Node | `node` | string | Target node for execution |
| Credentials | `credentials` | []string | Credential slugs to inject |
| Group | `group` | string | Step group for visual grouping |
| PlannedDuration | `planned_duration` | string | Expected duration for Gantt display |
| Due | `due` | string | Deadline |
| Process | `process` | string | Sub-process reference |
| OutputMapping | `output_mapping` | map | Step output → flow variable mapping |

## Control Node Fields

| Field | YAML Key | Used By | Description |
|-------|----------|---------|-------------|
| Condition | `condition` | if, loop | Boolean expression |
| Expression | `expression` | switch | Value expression |
| Branches | `branches` | if, switch | Map of value → target step ID |
| Mode | `mode` | loop | `while` / `do-while` |
| MaxIterations | `max_iterations` | loop | Max loop count |
| Items | `items` | foreach | Array expression |
| As | `as` | foreach | Loop variable name |
| Concurrency | `concurrency` | foreach | Parallel execution count |
| Steps | `steps` | loop, foreach | Nested sub-steps |

## WorkflowDef Top-Level Fields

Only these fields are valid at the top level of the YAML:

```yaml
description: "Flow description"    # optional
steps: []                          # required — list of StepDef
```

**NOT valid at top level**: `input_label`, `input_placeholder`, `assignee`, `assign`, or any other undocumented field.

## Gate Definition

```yaml
gate:
  condition: "{{output.data.score}} > 80"
  on_pass: next_step        # step to run if condition is true
  on_fail: retry_step       # step to run if condition is false
```

## Retry Configuration

```yaml
retry:
  max_attempts: 3
  delay: "10s"              # delay between retries
```

## DependsOn Formats

```yaml
# Simple — wait for step to complete (any status)
depends_on: [step_a, step_b]

# Conditional — only run if step completed successfully
depends_on:
  - step_id: step_a
    status: completed
```
