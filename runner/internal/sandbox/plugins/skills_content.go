package plugins

// DelegateSkillContent is the content for am-delegate skill.
// This skill enables agents to delegate tasks to other pods.
const DelegateSkillContent = `---
name: am-delegate
description: |
  WHEN to use:
  - 需要将任务委派给其他环境/仓库的 Agent 执行
  - 任务需要并行处理，需要创建子 Pod
  - 当前任务完成后需要触发下一个 Agent 继续

  WHEN NOT to use:
  - 任务可以在当前环境完成
  - 只是简单的信息查询
user-invocable: false
---

# 任务委派协议

你可以通过 MCP 工具将任务委派给其他 Agent Pod。

## 可用工具

- ` + "`list_available_pods`" + ` - 列出可用的 Pod
- ` + "`create_pod`" + ` - 创建新的 Agent Pod
- ` + "`bind_pod`" + ` - 绑定到目标 Pod（获取终端权限）
- ` + "`send_terminal_text`" + ` - 向 Pod 终端发送文本/任务
- ` + "`send_channel_message`" + ` - 通过 Channel 发送消息

## 委派流程

### 1. 检查现有 Pod

首先检查是否有自己创建的空闲 Pod 可以复用：

` + "```" + `
list_available_pods()
` + "```" + `

如果有空闲 Pod 且适合当前任务，直接复用；否则创建新 Pod。

### 2. 创建新 Pod（如需要）

` + "```" + `
create_pod(
  initial_prompt="你是负责 [具体职责] 的 Agent...",
  ticket_id=123  // 可选，关联到 Ticket
)
` + "```" + `

创建后会自动获得对新 Pod 的绑定权限。

### 3. 分配任务

方式一：通过终端直接发送指令
` + "```" + `
send_terminal_text(
  pod_key="target-pod-key",
  text="请实现用户登录功能，完成后通过 Channel 通知我"
)
` + "```" + `

方式二：通过 Channel 分配（推荐用于异步任务）
` + "```" + `
send_channel_message(
  channel_id=456,
  content="@target-pod 请实现用户登录功能"
)
` + "```" + `

### 4. 建立通信

确保与目标 Pod 有共同的 Channel 用于后续沟通：
- 告知目标 Pod 完成后通过 Channel 汇报
- 指定使用哪个 Channel 进行沟通

## 注意事项

- 优先创建新 Pod，除非自己创建的 Pod 正在空闲
- 委派时说明清楚任务目标和完成标准
- 指定通信方式（Channel）以便接收完成通知
`

// ChannelSkillContent is the content for am-channel skill.
// This skill enables agents to communicate via channels and sync status.
const ChannelSkillContent = `---
name: am-channel
description: |
  WHEN to use:
  - 开始新任务前检查是否有新消息
  - 任务状态发生变化需要同步（开始、完成、阻塞）
  - 需要与其他 Agent 交换信息
  - 需要更新 Ticket 状态

  WHEN NOT to use:
  - 独立的代码编写过程中
  - 不涉及协作的单独任务
user-invocable: false
---

# Channel 通信协议

你可以通过 Channel 与其他 Agent 进行异步通信。

## 可用工具

- ` + "`search_channels`" + ` - 搜索 Channel
- ` + "`get_channel`" + ` - 获取 Channel 详情
- ` + "`create_channel`" + ` - 创建新 Channel
- ` + "`get_channel_messages`" + ` - 获取 Channel 消息
- ` + "`send_channel_message`" + ` - 发送消息到 Channel
- ` + "`update_ticket`" + ` - 更新 Ticket 状态

## 检查消息

开始任务前，检查是否有新的任务分配或消息：

` + "```" + `
// 1. 查找相关 Channel
search_channels(name="project-x")

// 2. 获取最新消息
get_channel_messages(channel_id=123, limit=20)
` + "```" + `

重点关注：
- 任务分配消息
- 问题反馈
- @提及你的消息

## 状态同步

### 开始任务

` + "```" + `
// 通知团队
send_channel_message(
  channel_id=123,
  content="开始处理用户登录功能"
)

// 更新 Ticket 状态
update_ticket(ticket_id="AM-456", status="in_progress")
` + "```" + `

### 完成任务

` + "```" + `
// 通知团队
send_channel_message(
  channel_id=123,
  content="用户登录功能已完成，可以进行联调"
)

// 更新 Ticket 状态
update_ticket(ticket_id="AM-456", status="done")
` + "```" + `

### 遇到阻塞

` + "```" + `
send_channel_message(
  channel_id=123,
  content="遇到问题：缺少 API 文档，需要后端同学提供接口定义"
)
` + "```" + `

说明问题和需要的帮助，然后继续其他可进行的工作。

## 注意事项

- 状态变化时及时同步，保持团队信息透明
- 使用 @mention 明确通知相关 Pod
- 完成后如有后续任务，使用 am-delegate 委派
`
