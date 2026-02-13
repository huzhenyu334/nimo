# Hook Setup Guide

## Overview

Claude Code hooks trigger shell scripts at specific lifecycle events. We use `Stop` and `SessionEnd` hooks to notify OpenClaw when CC finishes a task.

## Configuration

### ~/.claude/settings.json

Add hooks configuration:

```json
{
  "hooks": {
    "Stop": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "bash ~/.claude/hooks/notify-openclaw.sh"
          }
        ]
      }
    ],
    "SessionEnd": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "bash ~/.claude/hooks/notify-openclaw.sh"
          }
        ]
      }
    ]
  }
}
```

### Hook Script (~/.claude/hooks/notify-openclaw.sh)

The hook script receives JSON on stdin with:
- `stop_reason`: why CC stopped (e.g., "end_turn", "max_tokens")
- `session_id`: CC session identifier
- `transcript_path`: path to JSONL transcript file

The script:
1. **Deduplicates** — 30-second lock prevents Stop+SessionEnd from double-firing
2. **Extracts from transcript** — task, final response, tool calls, errors, test results
3. **Generates report** — saved to `.claude-code-reports/YYYYMMDD-HHMMSS-sessionid.md`
4. **Notifies OpenClaw** — via `openclaw agent` (wakes main session)
5. **Notifies user** — via `openclaw message send` (Feishu DM)

### Transcript JSONL Structure

Each line is a JSON object:
- `type: "user"` — user messages, also contains `tool_result` entries
- `type: "assistant"` — CC responses, contains `tool_use` entries
- `type: "queue-operation"` — internal operations

Content format varies:
- String (simple `-p` mode)
- Array of `{type: "text", text: "..."}` or `{type: "tool_use", name: "...", input: {...}}`

### Test Result Extraction

The hook greps tool results for:
- Go tests: lines matching `ok`, `FAIL`, `PASS`
- Playwright: lines matching `passed`, `failed`, `skipped`

## Troubleshooting

- Debug log: `/tmp/claude-hook-debug.log`
- Lock file: `/tmp/.claude-hook-lock` (delete if stuck)
- Test manually: `echo '{"stop_reason":"test","session_id":"test123"}' | bash ~/.claude/hooks/notify-openclaw.sh`
