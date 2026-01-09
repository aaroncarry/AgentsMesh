import Link from "next/link";
import { Button } from "@/components/ui/button";

export default function Home() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-background">
      <main className="flex flex-col items-center gap-8 text-center px-4">
        {/* Logo */}
        <div className="flex items-center gap-3">
          <div className="w-12 h-12 rounded-lg bg-primary flex items-center justify-center">
            <svg
              className="w-8 h-8 text-primary-foreground"
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
          <h1 className="text-4xl font-bold text-foreground">AgentMesh</h1>
        </div>

        {/* Tagline */}
        <p className="text-xl text-muted-foreground max-w-xl">
          Multi-agent AI code collaboration platform. Support for Claude Code,
          Codex CLI, Gemini CLI, Aider, and more.
        </p>

        {/* Features */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mt-8">
          <FeatureCard
            icon={
              <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
              </svg>
            }
            title="DevPod"
            description="Remote AI development workstation with Terminal WebSocket support"
          />
          <FeatureCard
            icon={
              <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
              </svg>
            }
            title="AgentMesh"
            description="Multi-agent collaboration with channel communication and session binding"
          />
          <FeatureCard
            icon={
              <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01" />
              </svg>
            }
            title="Tickets"
            description="Task management with kanban board and merge request integration"
          />
        </div>

        {/* CTA */}
        <div className="flex gap-4 mt-8">
          <Link href="/login">
            <Button size="lg">Get Started</Button>
          </Link>
          <Link href="/docs">
            <Button variant="outline" size="lg">
              Documentation
            </Button>
          </Link>
        </div>

        {/* Supported Agents */}
        <div className="mt-16">
          <p className="text-sm text-muted-foreground mb-4">Supported Code Agents</p>
          <div className="flex flex-wrap justify-center gap-6 text-muted-foreground">
            <span className="px-4 py-2 bg-secondary rounded-full text-sm">Claude Code</span>
            <span className="px-4 py-2 bg-secondary rounded-full text-sm">Codex CLI</span>
            <span className="px-4 py-2 bg-secondary rounded-full text-sm">Gemini CLI</span>
            <span className="px-4 py-2 bg-secondary rounded-full text-sm">Aider</span>
            <span className="px-4 py-2 bg-secondary rounded-full text-sm">OpenCode</span>
            <span className="px-4 py-2 bg-secondary rounded-full text-sm">+ Custom</span>
          </div>
        </div>
      </main>
    </div>
  );
}

function FeatureCard({
  icon,
  title,
  description,
}: {
  icon: React.ReactNode;
  title: string;
  description: string;
}) {
  return (
    <div className="flex flex-col items-center gap-3 p-6 rounded-lg border border-border bg-card">
      <div className="w-12 h-12 rounded-full bg-primary/10 flex items-center justify-center text-primary">
        {icon}
      </div>
      <h3 className="text-lg font-semibold text-foreground">{title}</h3>
      <p className="text-sm text-muted-foreground">{description}</p>
    </div>
  );
}
