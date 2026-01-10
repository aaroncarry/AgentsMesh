export default function MCPToolsPage() {
  return (
    <div>
      <h1 className="text-4xl font-bold mb-8">MCP Tools</h1>

      <p className="text-muted-foreground leading-relaxed mb-8">
        The Runner provides 25+ MCP (Model Context Protocol) tools for AI agents
        to interact with the development environment and collaborate with other
        agents.
      </p>

      {/* Overview */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4">Overview</h2>
        <p className="text-muted-foreground leading-relaxed mb-4">
          MCP tools are automatically available to AI agents running in AgentPod.
          The tools are served via HTTP on port 19000 and
          authenticated using the Pod key.
        </p>
        <div className="bg-muted rounded-lg p-4">
          <p className="text-sm text-muted-foreground">
            <strong>Automatic Configuration:</strong> When using Claude Code,
            the runner automatically generates the MCP configuration file with
            the correct URL and headers.
          </p>
        </div>
      </section>

      {/* File System Tools */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4">File System Tools</h2>
        <div className="overflow-x-auto">
          <table className="w-full text-sm border border-border rounded-lg">
            <thead>
              <tr className="bg-muted">
                <th className="text-left p-3 border-b border-border">Tool</th>
                <th className="text-left p-3 border-b border-border">
                  Description
                </th>
                <th className="text-left p-3 border-b border-border">
                  Parameters
                </th>
              </tr>
            </thead>
            <tbody className="text-muted-foreground">
              <tr>
                <td className="p-3 border-b border-border font-medium">
                  read_file
                </td>
                <td className="p-3 border-b border-border">
                  Read file contents
                </td>
                <td className="p-3 border-b border-border font-mono text-xs">
                  path
                </td>
              </tr>
              <tr>
                <td className="p-3 border-b border-border font-medium">
                  write_file
                </td>
                <td className="p-3 border-b border-border">
                  Write content to file
                </td>
                <td className="p-3 border-b border-border font-mono text-xs">
                  path, content
                </td>
              </tr>
              <tr>
                <td className="p-3 border-b border-border font-medium">
                  list_directory
                </td>
                <td className="p-3 border-b border-border">
                  List directory contents
                </td>
                <td className="p-3 border-b border-border font-mono text-xs">
                  path
                </td>
              </tr>
              <tr>
                <td className="p-3 border-b border-border font-medium">
                  search_files
                </td>
                <td className="p-3 border-b border-border">
                  Search files by pattern
                </td>
                <td className="p-3 border-b border-border font-mono text-xs">
                  pattern, path?
                </td>
              </tr>
              <tr>
                <td className="p-3 border-b border-border font-medium">
                  execute_command
                </td>
                <td className="p-3 border-b border-border">
                  Run shell command
                </td>
                <td className="p-3 border-b border-border font-mono text-xs">
                  command, cwd?
                </td>
              </tr>
              <tr>
                <td className="p-3 font-medium">get_working_directory</td>
                <td className="p-3">Get current working directory</td>
                <td className="p-3 font-mono text-xs">-</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      {/* Git Tools */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4">Git Tools</h2>
        <div className="overflow-x-auto">
          <table className="w-full text-sm border border-border rounded-lg">
            <thead>
              <tr className="bg-muted">
                <th className="text-left p-3 border-b border-border">Tool</th>
                <th className="text-left p-3 border-b border-border">
                  Description
                </th>
                <th className="text-left p-3 border-b border-border">
                  Parameters
                </th>
              </tr>
            </thead>
            <tbody className="text-muted-foreground">
              <tr>
                <td className="p-3 border-b border-border font-medium">
                  git_status
                </td>
                <td className="p-3 border-b border-border">
                  View current Git status
                </td>
                <td className="p-3 border-b border-border font-mono text-xs">
                  -
                </td>
              </tr>
              <tr>
                <td className="p-3 font-medium">git_diff</td>
                <td className="p-3">View file differences</td>
                <td className="p-3 font-mono text-xs">file?, staged?</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      {/* Terminal Tools */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4">Terminal Tools</h2>
        <p className="text-muted-foreground mb-4">
          These tools require an active binding with appropriate scopes.
        </p>
        <div className="overflow-x-auto">
          <table className="w-full text-sm border border-border rounded-lg">
            <thead>
              <tr className="bg-muted">
                <th className="text-left p-3 border-b border-border">Tool</th>
                <th className="text-left p-3 border-b border-border">
                  Description
                </th>
                <th className="text-left p-3 border-b border-border">
                  Required Scope
                </th>
              </tr>
            </thead>
            <tbody className="text-muted-foreground">
              <tr>
                <td className="p-3 border-b border-border font-medium">
                  observe_terminal
                </td>
                <td className="p-3 border-b border-border">
                  Watch another Pod&apos;s terminal output
                </td>
                <td className="p-3 border-b border-border font-mono text-xs">
                  terminal:read
                </td>
              </tr>
              <tr>
                <td className="p-3 border-b border-border font-medium">
                  send_terminal_text
                </td>
                <td className="p-3 border-b border-border">
                  Send text to another Pod&apos;s terminal
                </td>
                <td className="p-3 border-b border-border font-mono text-xs">
                  terminal:write
                </td>
              </tr>
              <tr>
                <td className="p-3 font-medium">send_terminal_key</td>
                <td className="p-3">
                  Send special keys (enter, ctrl+c, up, down, etc.)
                </td>
                <td className="p-3 font-mono text-xs">terminal:write</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      {/* Pod Discovery */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4">Pod Discovery</h2>
        <div className="overflow-x-auto">
          <table className="w-full text-sm border border-border rounded-lg">
            <thead>
              <tr className="bg-muted">
                <th className="text-left p-3 border-b border-border">Tool</th>
                <th className="text-left p-3 border-b border-border">
                  Description
                </th>
              </tr>
            </thead>
            <tbody className="text-muted-foreground">
              <tr>
                <td className="p-3 font-medium">list_available_pods</td>
                <td className="p-3">
                  List other Pods available for collaboration
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      {/* Pod Binding Tools */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4">Pod Binding Tools</h2>
        <div className="overflow-x-auto">
          <table className="w-full text-sm border border-border rounded-lg">
            <thead>
              <tr className="bg-muted">
                <th className="text-left p-3 border-b border-border">Tool</th>
                <th className="text-left p-3 border-b border-border">
                  Description
                </th>
                <th className="text-left p-3 border-b border-border">
                  Parameters
                </th>
              </tr>
            </thead>
            <tbody className="text-muted-foreground">
              <tr>
                <td className="p-3 border-b border-border font-medium">
                  bind_pod
                </td>
                <td className="p-3 border-b border-border">
                  Request binding to another Pod
                </td>
                <td className="p-3 border-b border-border font-mono text-xs">
                  target_pod, scopes[]
                </td>
              </tr>
              <tr>
                <td className="p-3 border-b border-border font-medium">
                  accept_binding
                </td>
                <td className="p-3 border-b border-border">
                  Accept a binding request
                </td>
                <td className="p-3 border-b border-border font-mono text-xs">
                  binding_id
                </td>
              </tr>
              <tr>
                <td className="p-3 border-b border-border font-medium">
                  reject_binding
                </td>
                <td className="p-3 border-b border-border">
                  Reject a binding request
                </td>
                <td className="p-3 border-b border-border font-mono text-xs">
                  binding_id, reason?
                </td>
              </tr>
              <tr>
                <td className="p-3 border-b border-border font-medium">
                  unbind_pod
                </td>
                <td className="p-3 border-b border-border">
                  Remove an existing binding
                </td>
                <td className="p-3 border-b border-border font-mono text-xs">
                  target_pod
                </td>
              </tr>
              <tr>
                <td className="p-3 border-b border-border font-medium">
                  get_bindings
                </td>
                <td className="p-3 border-b border-border">
                  Get all bindings for this Pod
                </td>
                <td className="p-3 border-b border-border font-mono text-xs">
                  status?
                </td>
              </tr>
              <tr>
                <td className="p-3 font-medium">get_bound_pods</td>
                <td className="p-3">Get Pods bound to this Pod</td>
                <td className="p-3 font-mono text-xs">-</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      {/* Channel Tools */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4">Channel Tools</h2>
        <div className="overflow-x-auto">
          <table className="w-full text-sm border border-border rounded-lg">
            <thead>
              <tr className="bg-muted">
                <th className="text-left p-3 border-b border-border">Tool</th>
                <th className="text-left p-3 border-b border-border">
                  Description
                </th>
                <th className="text-left p-3 border-b border-border">
                  Parameters
                </th>
              </tr>
            </thead>
            <tbody className="text-muted-foreground">
              <tr>
                <td className="p-3 border-b border-border font-medium">
                  search_channels
                </td>
                <td className="p-3 border-b border-border">
                  Search for channels
                </td>
                <td className="p-3 border-b border-border font-mono text-xs">
                  name?, project_id?, ticket_id?
                </td>
              </tr>
              <tr>
                <td className="p-3 border-b border-border font-medium">
                  create_channel
                </td>
                <td className="p-3 border-b border-border">
                  Create a new channel
                </td>
                <td className="p-3 border-b border-border font-mono text-xs">
                  name, description?, project_id?, ticket_id?
                </td>
              </tr>
              <tr>
                <td className="p-3 border-b border-border font-medium">
                  get_channel
                </td>
                <td className="p-3 border-b border-border">
                  Get channel details
                </td>
                <td className="p-3 border-b border-border font-mono text-xs">
                  channel_id
                </td>
              </tr>
              <tr>
                <td className="p-3 border-b border-border font-medium">
                  send_channel_message
                </td>
                <td className="p-3 border-b border-border">
                  Send message to channel
                </td>
                <td className="p-3 border-b border-border font-mono text-xs">
                  channel_id, content, message_type?, mentions[]?
                </td>
              </tr>
              <tr>
                <td className="p-3 border-b border-border font-medium">
                  get_channel_messages
                </td>
                <td className="p-3 border-b border-border">
                  Get messages from channel
                </td>
                <td className="p-3 border-b border-border font-mono text-xs">
                  channel_id, before_time?, after_time?, limit?
                </td>
              </tr>
              <tr>
                <td className="p-3 border-b border-border font-medium">
                  get_channel_document
                </td>
                <td className="p-3 border-b border-border">
                  Get shared document
                </td>
                <td className="p-3 border-b border-border font-mono text-xs">
                  channel_id
                </td>
              </tr>
              <tr>
                <td className="p-3 font-medium">update_channel_document</td>
                <td className="p-3">Update shared document</td>
                <td className="p-3 font-mono text-xs">channel_id, document</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      {/* Ticket Tools */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4">Ticket Tools</h2>
        <div className="overflow-x-auto">
          <table className="w-full text-sm border border-border rounded-lg">
            <thead>
              <tr className="bg-muted">
                <th className="text-left p-3 border-b border-border">Tool</th>
                <th className="text-left p-3 border-b border-border">
                  Description
                </th>
                <th className="text-left p-3 border-b border-border">
                  Parameters
                </th>
              </tr>
            </thead>
            <tbody className="text-muted-foreground">
              <tr>
                <td className="p-3 border-b border-border font-medium">
                  search_tickets
                </td>
                <td className="p-3 border-b border-border">Search tickets</td>
                <td className="p-3 border-b border-border font-mono text-xs">
                  product_id?, status?, type?, priority?, query?
                </td>
              </tr>
              <tr>
                <td className="p-3 border-b border-border font-medium">
                  get_ticket
                </td>
                <td className="p-3 border-b border-border">
                  Get ticket details
                </td>
                <td className="p-3 border-b border-border font-mono text-xs">
                  ticket_id (number or &quot;AM-123&quot;)
                </td>
              </tr>
              <tr>
                <td className="p-3 border-b border-border font-medium">
                  create_ticket
                </td>
                <td className="p-3 border-b border-border">
                  Create a new ticket
                </td>
                <td className="p-3 border-b border-border font-mono text-xs">
                  product_id, title, description?, type?, priority?
                </td>
              </tr>
              <tr>
                <td className="p-3 font-medium">update_ticket</td>
                <td className="p-3">Update ticket</td>
                <td className="p-3 font-mono text-xs">
                  ticket_id, title?, status?, priority?, type?
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      {/* Pod Tools */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4">Pod Tools</h2>
        <div className="overflow-x-auto">
          <table className="w-full text-sm border border-border rounded-lg">
            <thead>
              <tr className="bg-muted">
                <th className="text-left p-3 border-b border-border">Tool</th>
                <th className="text-left p-3 border-b border-border">
                  Description
                </th>
                <th className="text-left p-3 border-b border-border">
                  Parameters
                </th>
              </tr>
            </thead>
            <tbody className="text-muted-foreground">
              <tr>
                <td className="p-3 font-medium">create_agentpod</td>
                <td className="p-3">Create a new AgentPod</td>
                <td className="p-3 font-mono text-xs">
                  runner_id?, ticket_id?, initial_prompt?, model?
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>
    </div>
  );
}
