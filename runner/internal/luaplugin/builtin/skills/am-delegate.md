---
name: am-delegate
description: |
  WHEN to use:
  - Need to delegate tasks to agents in other environments/repositories
  - Tasks need parallel processing, need to create sub-pods
  - After current task completion, need to trigger next agent

  WHEN NOT to use:
  - Task can be completed in current environment
  - Simple information queries
user-invocable: false
---

# Task Delegation Protocol

You can delegate tasks to other Agent Pods via MCP tools.

## Available Tools

- `list_available_pods` - List available Pods
- `create_pod` - Create new Agent Pod
- `bind_pod` - Bind to target Pod (get terminal access)
- `send_terminal_text` - Send text/task to Pod terminal
- `send_channel_message` - Send message via Channel

## Delegation Flow

### 1. Check Existing Pods

First check if there are idle Pods you created that can be reused:

```
list_available_pods()
```

If there's an idle Pod suitable for current task, reuse it; otherwise create new Pod.

### 2. Create New Pod (if needed)

```
create_pod(
  initial_prompt="You are an Agent responsible for [specific duty]...",
  ticket_id=123  // Optional, link to Ticket
)
```

After creation, you automatically get binding permissions to the new Pod.

### 3. Assign Task

Method 1: Send command directly via terminal
```
send_terminal_text(
  pod_key="target-pod-key",
  text="Please implement user login feature, notify me via Channel when done"
)
```

Method 2: Assign via Channel (recommended for async tasks)
```
send_channel_message(
  channel_id=456,
  content="@target-pod Please implement user login feature"
)
```

### 4. Establish Communication

Ensure you have a common Channel with target Pod for follow-up communication:
- Tell target Pod to report via Channel when done
- Specify which Channel to use for communication

## Notes

- Prefer creating new Pods unless your created Pod is idle
- Clearly state task goals and completion criteria when delegating
- Specify communication method (Channel) to receive completion notification
