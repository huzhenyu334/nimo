#!/bin/bash
# Claude Code Stop Hook — notify OpenClaw agent via Feishu
LOCK_FILE="/tmp/.claude-hook-lock"
LOCK_TIMEOUT=30
LOG="/tmp/claude-hook-debug.log"
WORKDIR="/home/claw/.openclaw/workspace"

echo "$(date): Hook triggered" >> "$LOG"

# Dedup
if [ -f "$LOCK_FILE" ]; then
    last=$(cat "$LOCK_FILE" 2>/dev/null)
    now=$(date +%s)
    if [ -n "$last" ] && [ $((now - last)) -lt $LOCK_TIMEOUT ]; then
        echo "$(date): Skipped (dedup)" >> "$LOG"
        exit 0
    fi
fi
date +%s > "$LOCK_FILE"

# Read stdin
INPUT=$(cat)
STOP_REASON=$(echo "$INPUT" | jq -r '.stop_reason // "unknown"' 2>/dev/null)

# Extract summary
SUMMARY=$(echo "$INPUT" | jq -r '
    .transcript // [] |
    [.[] | select(.role == "assistant")] | last // empty |
    if .content | type == "array" then
        [.content[] | select(.type == "text") | .text] | join(" ")
    elif .content | type == "string" then
        .content
    else
        empty
    end
' 2>/dev/null | head -c 800)

[ -z "$SUMMARY" ] && SUMMARY="(无摘要)"

# Git diff stats
GIT_STATS=$(cd "$WORKDIR" && git diff --stat 2>/dev/null | tail -1)
GIT_FILES=$(cd "$WORKDIR" && git diff --name-only 2>/dev/null | head -20)
[ -z "$GIT_STATS" ] && GIT_STATS="无文件改动"

MSG="收到Claude Code HOOK的回复：

📋 摘要：${SUMMARY}
📁 统计：${GIT_STATS}
📄 文件：${GIT_FILES}"

echo "$(date): stop_reason=$STOP_REASON" >> "$LOG"

# Send via openclaw agent
echo "$(date): Sending agent message..." >> "$LOG"
openclaw agent --agent main --message "$MSG" --deliver --reply-channel feishu >/dev/null 2>&1
echo "$(date): agent exit=$?" >> "$LOG"

# Backup: direct Feishu DM (always works)
echo "$(date): Sending Feishu DM backup..." >> "$LOG"
openclaw message send --channel feishu --target "user:ou_5b159fc157d4042f1e8088b1ffebb2da" --message "$MSG" >/dev/null 2>&1
echo "$(date): message send exit=$?" >> "$LOG"

echo "$(date): Hook done" >> "$LOG"
exit 0
