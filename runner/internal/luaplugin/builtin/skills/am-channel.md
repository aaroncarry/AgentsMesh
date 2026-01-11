---
name: am-channel
description: |
  WHEN to use:
  - Check for new messages before starting new task
  - Task status changes need sync (start, complete, blocked)
  - Need to exchange info with other Agents
  - Need to update Ticket status

  WHEN NOT to use:
  - During independent code writing
  - Standalone tasks without collaboration
user-invocable: false
---

# Channel Communication Protocol

You can communicate asynchronously with other Agents via Channels.

## Available Tools

- `search_channels` - Search Channels
- `get_channel` - Get Channel details
- `create_channel` - Create new Channel
- `get_channel_messages` - Get Channel messages
- `send_channel_message` - Send message to Channel
- `update_ticket` - Update Ticket status

## Check Messages

Before starting task, check for new task assignments or messages:

```
// 1. Find related Channels
search_channels(name="project-x")

// 2. Get latest messages
get_channel_messages(channel_id=123, limit=20)
```

Focus on:
- Task assignment messages
- Issue feedback
- Messages @mentioning you

## Status Sync

### Starting Task

```
// Notify team
send_channel_message(
  channel_id=123,
  content="Starting work on user login feature"
)

// Update Ticket status
update_ticket(ticket_id="AM-456", status="in_progress")
```

### Completing Task

```
// Notify team
send_channel_message(
  channel_id=123,
  content="User login feature completed, ready for integration"
)

// Update Ticket status
update_ticket(ticket_id="AM-456", status="done")
```

### Encountering Blockers

```
send_channel_message(
  channel_id=123,
  content="Issue encountered: missing API docs, need backend team to provide interface definitions"
)
```

State the problem and help needed, then continue with other available work.

## Notes

- Sync status changes promptly, keep team informed
- Use @mention to explicitly notify relevant Pods
- If there are follow-up tasks after completion, use am-delegate to delegate
