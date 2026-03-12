#!/bin/bash
# Check if CC mobile optimization task is done
# Called by cron at 5:00 AM CST (21:00 UTC)

SESSION_ID="cas-mmazbz1k-iwyetf"
LOG_FILE="/home/claw/.openclaw/workspace/memory/cc-mobile-check.log"

echo "$(date): Checking CC session $SESSION_ID" >> "$LOG_FILE"

# Check if there are recent commits (last 8 hours)
cd /home/claw/.openclaw/workspace/agent-control-panel
RECENT_COMMITS=$(git log --since="8 hours ago" --oneline | wc -l)
LAST_COMMIT=$(git log --oneline -1)
echo "$(date): Recent commits: $RECENT_COMMITS, Last: $LAST_COMMIT" >> "$LOG_FILE"

# Check if build exists
if [ -f "acp-web/dist/index.html" ]; then
    echo "$(date): Build exists" >> "$LOG_FILE"
else
    echo "$(date): No build found!" >> "$LOG_FILE"
fi

echo "$(date): Check complete" >> "$LOG_FILE"
