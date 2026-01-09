"use client";

import { useEffect, useState, useCallback } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Terminal, useTerminal } from "./Terminal";
import { sessionApi, SessionData } from "@/lib/api/client";

interface SessionDetailProps {
  sessionKey: string;
}

const statusColors: Record<string, { bg: string; text: string; dot: string }> = {
  initializing: { bg: "bg-yellow-50", text: "text-yellow-700", dot: "bg-yellow-500" },
  running: { bg: "bg-green-50", text: "text-green-700", dot: "bg-green-500" },
  paused: { bg: "bg-blue-50", text: "text-blue-700", dot: "bg-blue-500" },
  terminated: { bg: "bg-gray-50", text: "text-gray-700", dot: "bg-gray-400" },
  failed: { bg: "bg-red-50", text: "text-red-700", dot: "bg-red-500" },
};

const statusLabels: Record<string, string> = {
  initializing: "Initializing",
  running: "Running",
  paused: "Paused",
  terminated: "Terminated",
  failed: "Failed",
};

export function SessionDetail({ sessionKey }: SessionDetailProps) {
  const router = useRouter();
  const [session, setSession] = useState<SessionData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [isTerminating, setIsTerminating] = useState(false);

  const { terminalRef, connect, disconnect, sendInput, sendResize } = useTerminal(sessionKey);

  // Fetch session data
  const fetchSession = useCallback(async () => {
    try {
      setLoading(true);
      const response = await sessionApi.get(sessionKey);
      setSession(response.session);
      setError(null);
    } catch (err) {
      console.error("Failed to fetch session:", err);
      setError("Failed to load session");
    } finally {
      setLoading(false);
    }
  }, [sessionKey]);

  useEffect(() => {
    fetchSession();
  }, [fetchSession]);

  // Auto-connect terminal when session is running
  useEffect(() => {
    if (session?.status === "running" && !isConnected) {
      handleConnect();
    }
  }, [session?.status]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (isConnected) {
        disconnect();
      }
    };
  }, [isConnected, disconnect]);

  const handleConnect = useCallback(() => {
    connect();
    setIsConnected(true);
  }, [connect]);

  const handleDisconnect = useCallback(() => {
    disconnect();
    setIsConnected(false);
  }, [disconnect]);

  const handleTerminate = async () => {
    if (!session) return;

    setIsTerminating(true);
    try {
      await sessionApi.terminate(sessionKey);
      handleDisconnect();
      await fetchSession();
    } catch (err) {
      console.error("Failed to terminate session:", err);
    } finally {
      setIsTerminating(false);
    }
  };

  const handleTerminalData = useCallback((data: string) => {
    sendInput(data);
  }, [sendInput]);

  const handleTerminalResize = useCallback((rows: number, cols: number) => {
    sendResize(rows, cols);
  }, [sendResize]);

  const formatTime = (dateString?: string) => {
    if (!dateString) return "—";
    return new Date(dateString).toLocaleString();
  };

  const formatDuration = (startedAt?: string) => {
    if (!startedAt) return "—";
    const start = new Date(startedAt);
    const now = new Date();
    const diff = Math.floor((now.getTime() - start.getTime()) / 1000);

    if (diff < 60) return `${diff}s`;
    if (diff < 3600) return `${Math.floor(diff / 60)}m ${diff % 60}s`;
    return `${Math.floor(diff / 3600)}h ${Math.floor((diff % 3600) / 60)}m`;
  };

  if (loading) {
    return <SessionDetailSkeleton />;
  }

  if (error) {
    return (
      <div className="text-center py-12">
        <div className="text-red-600 mb-4">{error}</div>
        <Button onClick={fetchSession}>Retry</Button>
      </div>
    );
  }

  if (!session) {
    return (
      <div className="text-center py-12 text-muted-foreground">
        Session not found
      </div>
    );
  }

  const statusStyle = statusColors[session.status] || statusColors.terminated;
  const isActive = session.status === "running" || session.status === "initializing";

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="flex-shrink-0 border-b border-border p-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <Button variant="ghost" size="sm" onClick={() => router.back()}>
              <svg className="w-4 h-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 19l-7-7m0 0l7-7m-7 7h18" />
              </svg>
              Back
            </Button>
            <div className="h-6 w-px bg-border" />
            <div className="flex items-center gap-3">
              <span
                className={`inline-flex items-center gap-1.5 px-2 py-1 rounded-full text-xs font-medium ${statusStyle.bg} ${statusStyle.text}`}
              >
                <span className={`w-1.5 h-1.5 rounded-full ${statusStyle.dot}`} />
                {statusLabels[session.status]}
              </span>
              <code className="text-sm font-mono bg-muted px-2 py-1 rounded">
                {sessionKey}
              </code>
            </div>
          </div>
          <div className="flex items-center gap-2">
            {isActive && !isConnected && (
              <Button size="sm" onClick={handleConnect}>
                <svg className="w-4 h-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
                </svg>
                Connect
              </Button>
            )}
            {isConnected && (
              <Button size="sm" variant="outline" onClick={handleDisconnect}>
                Disconnect
              </Button>
            )}
            {isActive && (
              <Button
                size="sm"
                variant="destructive"
                onClick={handleTerminate}
                disabled={isTerminating}
              >
                {isTerminating ? "Terminating..." : "Terminate"}
              </Button>
            )}
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex flex-1 min-h-0">
        {/* Terminal Area */}
        <div className="flex-1 flex flex-col min-w-0">
          {/* Terminal */}
          <div className="flex-1 min-h-0 p-4">
            <div className="w-full h-full rounded-lg overflow-hidden border border-border">
              <Terminal
                ref={terminalRef}
                sessionKey={sessionKey}
                onData={handleTerminalData}
                onResize={handleTerminalResize}
                className="h-full"
              />
            </div>
          </div>

        </div>

        {/* Sidebar */}
        <div className="w-80 flex-shrink-0 border-l border-border p-4 overflow-y-auto">
          {/* Session Info */}
          <div className="space-y-4">
            <div className="border border-border rounded-lg p-4">
              <h3 className="font-medium mb-3">Session Info</h3>
              <dl className="space-y-2 text-sm">
                {session.agent_type && (
                  <div className="flex justify-between">
                    <dt className="text-muted-foreground">Agent</dt>
                    <dd className="font-medium">{session.agent_type.name}</dd>
                  </div>
                )}
                <div className="flex justify-between">
                  <dt className="text-muted-foreground">Agent Status</dt>
                  <dd className="font-medium">{session.agent_status || "—"}</dd>
                </div>
                {session.runner && (
                  <div className="flex justify-between">
                    <dt className="text-muted-foreground">Runner</dt>
                    <dd className="font-mono text-xs">{session.runner.node_id}</dd>
                  </div>
                )}
                {isActive && (
                  <div className="flex justify-between">
                    <dt className="text-muted-foreground">Duration</dt>
                    <dd className="text-green-600 font-medium">
                      {formatDuration(session.started_at)}
                    </dd>
                  </div>
                )}
              </dl>
            </div>

            {/* Repository */}
            {session.repository && (
              <div className="border border-border rounded-lg p-4">
                <h3 className="font-medium mb-3">Repository</h3>
                <dl className="space-y-2 text-sm">
                  <div>
                    <dt className="text-muted-foreground text-xs">Path</dt>
                    <dd className="font-medium truncate" title={session.repository.full_path}>
                      {session.repository.full_path}
                    </dd>
                  </div>
                  {session.branch_name && (
                    <div>
                      <dt className="text-muted-foreground text-xs">Branch</dt>
                      <dd className="flex items-center gap-1">
                        <svg className="w-3 h-3 text-muted-foreground" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" />
                        </svg>
                        <span className="font-mono text-xs">{session.branch_name}</span>
                      </dd>
                    </div>
                  )}
                </dl>
              </div>
            )}

            {/* Ticket */}
            {session.ticket && (
              <div className="border border-border rounded-lg p-4">
                <h3 className="font-medium mb-3">Linked Ticket</h3>
                <div
                  className="p-3 bg-muted/50 rounded cursor-pointer hover:bg-muted transition-colors"
                  onClick={() => router.push(`/tickets/${session.ticket!.identifier}`)}
                >
                  <div className="text-xs text-primary font-mono mb-1">
                    {session.ticket.identifier}
                  </div>
                  <div className="text-sm truncate">{session.ticket.title}</div>
                </div>
              </div>
            )}

            {/* Initial Prompt */}
            {session.initial_prompt && (
              <div className="border border-border rounded-lg p-4">
                <h3 className="font-medium mb-3">Initial Prompt</h3>
                <p className="text-sm text-muted-foreground whitespace-pre-wrap">
                  {session.initial_prompt}
                </p>
              </div>
            )}

            {/* Timestamps */}
            <div className="border border-border rounded-lg p-4">
              <h3 className="font-medium mb-3">Timestamps</h3>
              <dl className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <dt className="text-muted-foreground">Created</dt>
                  <dd>{formatTime(session.created_at)}</dd>
                </div>
                {session.started_at && (
                  <div className="flex justify-between">
                    <dt className="text-muted-foreground">Started</dt>
                    <dd>{formatTime(session.started_at)}</dd>
                  </div>
                )}
                {session.finished_at && (
                  <div className="flex justify-between">
                    <dt className="text-muted-foreground">Finished</dt>
                    <dd>{formatTime(session.finished_at)}</dd>
                  </div>
                )}
                {session.last_activity && (
                  <div className="flex justify-between">
                    <dt className="text-muted-foreground">Last Activity</dt>
                    <dd>{formatTime(session.last_activity)}</dd>
                  </div>
                )}
              </dl>
            </div>

            {/* Created By */}
            {session.created_by && (
              <div className="border border-border rounded-lg p-4">
                <h3 className="font-medium mb-3">Created By</h3>
                <div className="flex items-center gap-2">
                  <div className="w-8 h-8 rounded-full bg-muted flex items-center justify-center text-sm font-medium">
                    {(session.created_by.name || session.created_by.username)[0].toUpperCase()}
                  </div>
                  <div>
                    <div className="text-sm font-medium">
                      {session.created_by.name || session.created_by.username}
                    </div>
                    <div className="text-xs text-muted-foreground">
                      @{session.created_by.username}
                    </div>
                  </div>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

function SessionDetailSkeleton() {
  return (
    <div className="animate-pulse" data-testid="session-detail-skeleton">
      <div className="border-b border-border p-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <div className="h-8 w-16 bg-muted rounded" />
            <div className="h-6 w-24 bg-muted rounded-full" />
            <div className="h-6 w-48 bg-muted rounded" />
          </div>
          <div className="flex gap-2">
            <div className="h-8 w-24 bg-muted rounded" />
            <div className="h-8 w-24 bg-muted rounded" />
          </div>
        </div>
      </div>
      <div className="flex">
        <div className="flex-1 p-4">
          <div className="h-96 bg-muted rounded-lg" />
        </div>
        <div className="w-80 border-l border-border p-4 space-y-4">
          <div className="h-40 bg-muted rounded-lg" />
          <div className="h-32 bg-muted rounded-lg" />
          <div className="h-24 bg-muted rounded-lg" />
        </div>
      </div>
    </div>
  );
}

export default SessionDetail;
