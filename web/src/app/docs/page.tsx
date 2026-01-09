import Link from "next/link";
import { Button } from "@/components/ui/button";

export default function DocsPage() {
  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="border-b border-border sticky top-0 bg-background z-10">
        <div className="container mx-auto px-4 py-4 flex items-center justify-between">
          <Link href="/" className="flex items-center gap-2">
            <div className="w-8 h-8 rounded-lg bg-primary flex items-center justify-center">
              <svg
                className="w-5 h-5 text-primary-foreground"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z"
                />
              </svg>
            </div>
            <span className="text-xl font-bold">AgentMesh</span>
          </Link>
          <Link href="/login">
            <Button variant="outline">Sign In</Button>
          </Link>
        </div>
      </header>

      <div className="flex">
        {/* Sidebar */}
        <aside className="w-64 border-r border-border min-h-[calc(100vh-65px)] p-4 hidden md:block sticky top-[65px] h-[calc(100vh-65px)] overflow-y-auto">
          <nav className="space-y-6">
            <div>
              <h3 className="font-semibold text-sm mb-2">Getting Started</h3>
              <ul className="space-y-1">
                <li>
                  <a href="#introduction" className="text-sm text-muted-foreground hover:text-foreground block py-1">
                    Introduction
                  </a>
                </li>
                <li>
                  <a href="#quick-start" className="text-sm text-muted-foreground hover:text-foreground block py-1">
                    Quick Start
                  </a>
                </li>
              </ul>
            </div>
            <div>
              <h3 className="font-semibold text-sm mb-2">Features</h3>
              <ul className="space-y-1">
                <li>
                  <a href="#devpod" className="text-sm text-muted-foreground hover:text-foreground block py-1">
                    DevPod
                  </a>
                </li>
                <li>
                  <a href="#devmesh" className="text-sm text-muted-foreground hover:text-foreground block py-1">
                    DevMesh
                  </a>
                </li>
                <li>
                  <a href="#tickets" className="text-sm text-muted-foreground hover:text-foreground block py-1">
                    Tickets
                  </a>
                </li>
                <li>
                  <a href="#channels" className="text-sm text-muted-foreground hover:text-foreground block py-1">
                    Channels
                  </a>
                </li>
              </ul>
            </div>
            <div>
              <h3 className="font-semibold text-sm mb-2">Runners</h3>
              <ul className="space-y-1">
                <li>
                  <a href="#runner-setup" className="text-sm text-muted-foreground hover:text-foreground block py-1">
                    Runner Setup
                  </a>
                </li>
                <li>
                  <a href="#mcp-tools" className="text-sm text-muted-foreground hover:text-foreground block py-1">
                    MCP Tools
                  </a>
                </li>
              </ul>
            </div>
            <div>
              <h3 className="font-semibold text-sm mb-2">API Reference</h3>
              <ul className="space-y-1">
                <li>
                  <a href="#rest-api" className="text-sm text-muted-foreground hover:text-foreground block py-1">
                    REST API
                  </a>
                </li>
                <li>
                  <a href="#webhooks" className="text-sm text-muted-foreground hover:text-foreground block py-1">
                    Webhooks
                  </a>
                </li>
              </ul>
            </div>
          </nav>
        </aside>

        {/* Content */}
        <main className="flex-1 p-8 max-w-4xl">
          <h1 className="text-4xl font-bold mb-8">Documentation</h1>

          {/* Introduction */}
          <section id="introduction" className="mb-12">
            <h2 className="text-2xl font-semibold mb-4">Introduction</h2>
            <p className="text-muted-foreground leading-relaxed mb-4">
              AgentMesh is a multi-agent AI code collaboration platform that enables teams to
              leverage AI coding assistants at scale. The platform provides remote development
              workstations (DevPod), multi-agent coordination (DevMesh), and integrated task
              management.
            </p>
            <div className="bg-muted rounded-lg p-4 mt-4">
              <p className="font-medium mb-2">Supported AI Agents:</p>
              <ul className="list-disc list-inside text-muted-foreground space-y-1">
                <li>Claude Code (Anthropic)</li>
                <li>Codex CLI (OpenAI)</li>
                <li>Gemini CLI (Google)</li>
                <li>Aider</li>
                <li>OpenCode</li>
                <li>Custom agents via MCP protocol</li>
              </ul>
            </div>
          </section>

          {/* Quick Start */}
          <section id="quick-start" className="mb-12">
            <h2 className="text-2xl font-semibold mb-4">Quick Start</h2>
            <div className="space-y-4">
              <div className="border border-border rounded-lg p-4">
                <h3 className="font-medium mb-2">1. Create an Account</h3>
                <p className="text-muted-foreground text-sm">
                  Sign up at{" "}
                  <Link href="/register" className="text-primary hover:underline">
                    /register
                  </Link>{" "}
                  to create your account and organization.
                </p>
              </div>
              <div className="border border-border rounded-lg p-4">
                <h3 className="font-medium mb-2">2. Add API Keys</h3>
                <p className="text-muted-foreground text-sm">
                  Go to Settings → Agents to configure your AI provider API keys (Anthropic,
                  OpenAI, Google).
                </p>
              </div>
              <div className="border border-border rounded-lg p-4">
                <h3 className="font-medium mb-2">3. Setup a Runner</h3>
                <p className="text-muted-foreground text-sm">
                  Download and configure the AgentMesh Runner on your development machine or
                  server.
                </p>
              </div>
              <div className="border border-border rounded-lg p-4">
                <h3 className="font-medium mb-2">4. Start a Session</h3>
                <p className="text-muted-foreground text-sm">
                  Create a new DevPod session and start coding with AI assistance!
                </p>
              </div>
            </div>
          </section>

          {/* DevPod */}
          <section id="devpod" className="mb-12">
            <h2 className="text-2xl font-semibold mb-4">DevPod</h2>
            <p className="text-muted-foreground leading-relaxed mb-4">
              DevPod provides remote AI development workstations with integrated terminal access.
              Each session runs on a Runner and provides a full development environment with AI
              assistance.
            </p>
            <h3 className="text-lg font-medium mt-6 mb-3">Features</h3>
            <ul className="list-disc list-inside text-muted-foreground space-y-2">
              <li>Web-based terminal with WebSocket support</li>
              <li>Real-time AI agent status monitoring</li>
              <li>Git worktree isolation for each session</li>
              <li>Automatic session cleanup</li>
              <li>Multiple concurrent sessions per user</li>
            </ul>
          </section>

          {/* DevMesh */}
          <section id="devmesh" className="mb-12">
            <h2 className="text-2xl font-semibold mb-4">DevMesh</h2>
            <p className="text-muted-foreground leading-relaxed mb-4">
              DevMesh visualizes and coordinates multiple AI agents working together. See the
              topology of active sessions, their bindings, and communication channels.
            </p>
            <h3 className="text-lg font-medium mt-6 mb-3">Session Bindings</h3>
            <p className="text-muted-foreground leading-relaxed">
              Sessions can bind to each other to share capabilities. Bindings support scoped
              permissions including terminal observation, file access, and command execution.
            </p>
            <h3 className="text-lg font-medium mt-6 mb-3">Binding Scopes</h3>
            <ul className="list-disc list-inside text-muted-foreground space-y-1">
              <li><code className="bg-muted px-1 rounded">observe_terminal</code> - View terminal output</li>
              <li><code className="bg-muted px-1 rounded">send_text</code> - Send text to terminal</li>
              <li><code className="bg-muted px-1 rounded">send_control</code> - Send control keys</li>
              <li><code className="bg-muted px-1 rounded">read_files</code> - Read workspace files</li>
              <li><code className="bg-muted px-1 rounded">write_files</code> - Write workspace files</li>
            </ul>
          </section>

          {/* Tickets */}
          <section id="tickets" className="mb-12">
            <h2 className="text-2xl font-semibold mb-4">Tickets</h2>
            <p className="text-muted-foreground leading-relaxed mb-4">
              Integrated task management with Kanban board view. Create tickets, assign them to
              AI sessions, and track progress through your workflow.
            </p>
            <h3 className="text-lg font-medium mt-6 mb-3">Ticket Types</h3>
            <ul className="list-disc list-inside text-muted-foreground space-y-1">
              <li>Task</li>
              <li>Bug</li>
              <li>Feature</li>
              <li>Epic</li>
              <li>Subtask</li>
              <li>Story</li>
            </ul>
          </section>

          {/* Channels */}
          <section id="channels" className="mb-12">
            <h2 className="text-2xl font-semibold mb-4">Channels</h2>
            <p className="text-muted-foreground leading-relaxed mb-4">
              Channels provide communication hubs for AI agents. Multiple sessions can join a
              channel to collaborate on tasks, share information, and coordinate work.
            </p>
            <h3 className="text-lg font-medium mt-6 mb-3">Message Types</h3>
            <ul className="list-disc list-inside text-muted-foreground space-y-1">
              <li><code className="bg-muted px-1 rounded">text</code> - Plain text messages</li>
              <li><code className="bg-muted px-1 rounded">code</code> - Code snippets</li>
              <li><code className="bg-muted px-1 rounded">system</code> - System notifications</li>
              <li><code className="bg-muted px-1 rounded">command</code> - Command execution results</li>
            </ul>
          </section>

          {/* Runner Setup */}
          <section id="runner-setup" className="mb-12">
            <h2 className="text-2xl font-semibold mb-4">Runner Setup</h2>
            <p className="text-muted-foreground leading-relaxed mb-4">
              Runners are the execution environments for AI agent sessions. Set up a runner on
              any machine with Git and your preferred development tools installed.
            </p>
            <div className="bg-muted rounded-lg p-4 font-mono text-sm overflow-x-auto">
              <pre>{`# Download the runner binary
curl -LO https://github.com/agentmesh/runner/releases/latest/download/runner

# Make executable
chmod +x runner

# Configure with your registration token
./runner configure --token <YOUR_TOKEN>

# Start the runner
./runner start`}</pre>
            </div>
          </section>

          {/* MCP Tools */}
          <section id="mcp-tools" className="mb-12">
            <h2 className="text-2xl font-semibold mb-4">MCP Tools</h2>
            <p className="text-muted-foreground leading-relaxed mb-4">
              The Runner provides 30+ MCP (Model Context Protocol) tools for AI agents to
              interact with the development environment.
            </p>
            <h3 className="text-lg font-medium mt-6 mb-3">Tool Categories</h3>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mt-4">
              <div className="border border-border rounded-lg p-4">
                <h4 className="font-medium mb-2">File System</h4>
                <ul className="text-sm text-muted-foreground space-y-1">
                  <li>read_file, write_file</li>
                  <li>list_directory</li>
                  <li>search_files</li>
                </ul>
              </div>
              <div className="border border-border rounded-lg p-4">
                <h4 className="font-medium mb-2">Git Operations</h4>
                <ul className="text-sm text-muted-foreground space-y-1">
                  <li>git_status, git_diff</li>
                  <li>git_commit, git_push</li>
                  <li>git_branch</li>
                </ul>
              </div>
              <div className="border border-border rounded-lg p-4">
                <h4 className="font-medium mb-2">Collaboration</h4>
                <ul className="text-sm text-muted-foreground space-y-1">
                  <li>send_message, get_messages</li>
                  <li>bind_session, unbind_session</li>
                  <li>observe_terminal</li>
                </ul>
              </div>
              <div className="border border-border rounded-lg p-4">
                <h4 className="font-medium mb-2">Tickets</h4>
                <ul className="text-sm text-muted-foreground space-y-1">
                  <li>get_ticket, create_ticket</li>
                  <li>update_ticket</li>
                  <li>search_tickets</li>
                </ul>
              </div>
            </div>
          </section>

          {/* REST API */}
          <section id="rest-api" className="mb-12">
            <h2 className="text-2xl font-semibold mb-4">REST API</h2>
            <p className="text-muted-foreground leading-relaxed mb-4">
              AgentMesh provides a comprehensive REST API for integration with your tools and
              workflows.
            </p>
            <div className="bg-muted rounded-lg p-4 font-mono text-sm overflow-x-auto">
              <pre>{`# Base URL
https://api.agentmesh.dev/api/v1

# Authentication
Authorization: Bearer <your-token>
X-Organization-Slug: <org-slug>

# Example: List sessions
GET /api/v1/org/sessions

# Example: Create session
POST /api/v1/org/sessions
{
  "agent_type_id": 1,
  "runner_id": 1,
  "initial_prompt": "Help me refactor this code"
}`}</pre>
            </div>
          </section>

          {/* Webhooks */}
          <section id="webhooks" className="mb-12">
            <h2 className="text-2xl font-semibold mb-4">Webhooks</h2>
            <p className="text-muted-foreground leading-relaxed mb-4">
              AgentMesh can receive webhooks from Git providers to trigger automated workflows.
            </p>
            <h3 className="text-lg font-medium mt-6 mb-3">Supported Providers</h3>
            <ul className="list-disc list-inside text-muted-foreground space-y-1">
              <li>GitHub (HMAC-SHA256 signature verification)</li>
              <li>GitLab (Token-based verification)</li>
              <li>Gitee (Token and HMAC verification)</li>
            </ul>
            <h3 className="text-lg font-medium mt-6 mb-3">Webhook Events</h3>
            <ul className="list-disc list-inside text-muted-foreground space-y-1">
              <li>Push events</li>
              <li>Merge/Pull request events</li>
              <li>Pipeline/CI events</li>
              <li>Issue events</li>
              <li>Comment events</li>
            </ul>
          </section>
        </main>
      </div>

      {/* Footer */}
      <footer className="border-t border-border mt-16">
        <div className="container mx-auto px-4 py-8">
          <div className="flex flex-col md:flex-row justify-between items-center gap-4">
            <p className="text-sm text-muted-foreground">
              &copy; 2025 AgentMesh. All rights reserved.
            </p>
            <div className="flex gap-6">
              <Link href="/privacy" className="text-sm text-muted-foreground hover:text-foreground">
                Privacy Policy
              </Link>
              <Link href="/terms" className="text-sm text-muted-foreground hover:text-foreground">
                Terms of Service
              </Link>
              <Link href="/docs" className="text-sm text-muted-foreground hover:text-foreground">
                Documentation
              </Link>
            </div>
          </div>
        </div>
      </footer>
    </div>
  );
}
