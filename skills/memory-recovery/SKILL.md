---
name: memory-recovery
description: Recover conversation context after compaction or cold start. Use when you wake up mid-conversation with missing context, receive a message referencing "this problem" or prior work you don't remember, see exec/session completion notifications without context, or need to figure out what was being discussed before compaction happened.
---

# Memory Recovery

After compaction, you lose conversation history but the user expects continuity. Follow this recovery chain in order — each step builds on the last. Stop as soon as you have enough context to answer.

## Recovery Chain (ordered by signal strength)

### 1. Daily Logs (fastest, richest)

```bash
# Read today's and yesterday's logs
cat memory/$(date +%Y-%m-%d).md
cat memory/$(date -d yesterday +%Y-%m-%d).md
```

Daily logs are your #1 source — they contain timestamped work entries with commit hashes, decisions, and problem descriptions. Written by you before each compaction flush.

### 2. MEMORY.md

Already injected in context. Scan for "当前状态", "最近CC任务", "待办" sections. But beware: MEMORY.md is lossy — compaction summaries sometimes mark completed work as "pending".

### 3. Trigger Clues

The inbound message often contains clues:
- **Exec completion notification** (e.g. `good-sea, code 0`) → grep session files for that ID
- **"这个问题"/"this issue"** → user is continuing a prior thread
- **Message timestamp gap** → estimate when compaction happened

```bash
# Find which session used a specific exec name
grep -r "good-sea" /home/claw/.openclaw/agents/main/sessions/*.jsonl | head -5
```

### 4. Compacted Session Tail

The old session file still exists. Read its last 20-30 messages to find what was being discussed right before compaction:

```bash
# Find recent large session files (compacted sessions are big)
ls -lS /home/claw/.openclaw/agents/main/sessions/*.jsonl | head -10

# Extract last messages' text content
tail -30 <session_file> | grep -o '"text":"[^"]*"' | tail -20
```

This is the **key technique for finding the exact last topic**. The compacted session retains all messages — you just need to read the tail.

### 5. Git Log

Recent commits reveal what was being worked on:

```bash
cd /path/to/project && git log --oneline -20
```

Commits with fix/feat prefixes tell you the current focus area.

### 6. memory_recall + acp_knowledge_search

Search long-term memory and knowledge base for relevant context:

```
memory_recall(query="具体关键词", limit=10)
acp_knowledge_search(query="具体关键词")
```

Use specific nouns (project names, function names, error messages) not vague terms.

### 7. Active Sessions & Processes

```
sessions_list(activeMinutes=60, messageLimit=5)
process list
```

Check if related sessions or background tasks are still running.

## Decision Tree

```
User message references prior context?
├─ Has exec completion notification?
│  └─ grep session files for exec ID → find source session → read tail
├─ References "this problem" / continuing discussion?
│  └─ Read daily log → read compacted session tail → find last topic
├─ Asks about work status?
│  └─ Read MEMORY.md → daily log → git log
└─ No clear reference?
   └─ Ask user what they're referring to
```

## Anti-Patterns

- **Don't guess** — if recovery chain doesn't surface the context, ask the user
- **Don't trust MEMORY.md blindly** — cross-check with daily logs and git
- **Don't use vague search terms** — "bug analysis" finds nothing; "debug console truncation" finds everything
- **Don't skip the session tail** — it's the most direct evidence of what was last discussed
