"use client";

import { useState, useEffect, useCallback } from "react";
import { Button } from "@/components/ui/button";
import { ticketApi, runnerApi } from "@/lib/api/client";
import { getSessionStatusInfo, getAgentStatusInfo } from "@/stores/devmesh";

interface TicketSession {
  session_key: string;
  status: string;
  agent_status: string;
  model?: string;
  started_at?: string;
  runner_id: number;
  created_by_id: number;
}

interface Runner {
  id: number;
  node_id: string;
  status: string;
  current_sessions: number;
  max_concurrent_sessions?: number;
}

interface TicketSessionPanelProps {
  ticketIdentifier: string;
  ticketTitle: string;
  onSessionCreated?: () => void;
}

export default function TicketSessionPanel({
  ticketIdentifier,
  ticketTitle,
  onSessionCreated,
}: TicketSessionPanelProps) {
  const [sessions, setSessions] = useState<TicketSession[]>([]);
  const [runners, setRunners] = useState<Runner[]>([]);
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Create form state
  const [selectedRunner, setSelectedRunner] = useState<number | null>(null);
  const [initialPrompt, setInitialPrompt] = useState("");
  const [model, setModel] = useState("claude-sonnet-4-20250514");
  const [permissionMode, setPermissionMode] = useState("default");

  const fetchSessions = useCallback(async () => {
    try {
      const response = await ticketApi.getSessions(ticketIdentifier);
      setSessions(response.sessions || []);
    } catch (err: any) {
      console.error("Failed to fetch sessions:", err);
    }
  }, [ticketIdentifier]);

  const fetchRunners = useCallback(async () => {
    try {
      const response = await runnerApi.list();
      setRunners(response.runners?.filter((r) => r.status === "online") || []);
    } catch (err: any) {
      console.error("Failed to fetch runners:", err);
    }
  }, []);

  useEffect(() => {
    const loadData = async () => {
      setLoading(true);
      await Promise.all([fetchSessions(), fetchRunners()]);
      setLoading(false);
    };
    loadData();
  }, [fetchSessions, fetchRunners]);

  // Poll for session updates
  useEffect(() => {
    const interval = setInterval(fetchSessions, 5000);
    return () => clearInterval(interval);
  }, [fetchSessions]);

  const handleCreateSession = async () => {
    if (!selectedRunner) {
      setError("Please select a runner");
      return;
    }

    setCreating(true);
    setError(null);

    try {
      await ticketApi.createSession(ticketIdentifier, {
        runner_id: selectedRunner,
        initial_prompt: initialPrompt || `Work on ticket: ${ticketTitle}`,
        model,
        permission_mode: permissionMode,
      });

      // Reset form
      setShowCreateForm(false);
      setSelectedRunner(null);
      setInitialPrompt("");
      setModel("claude-sonnet-4-20250514");
      setPermissionMode("default");

      // Refresh sessions
      await fetchSessions();
      onSessionCreated?.();
    } catch (err: any) {
      setError(err.message || "Failed to create session");
    } finally {
      setCreating(false);
    }
  };

  const activeSessions = sessions.filter(
    (s) => s.status === "running" || s.status === "initializing"
  );
  const inactiveSessions = sessions.filter(
    (s) => s.status !== "running" && s.status !== "initializing"
  );

  if (loading) {
    return (
      <div className="p-4 border border-border rounded-lg">
        <div className="flex items-center justify-center py-8">
          <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-primary"></div>
        </div>
      </div>
    );
  }

  return (
    <div className="border border-border rounded-lg">
      {/* Header */}
      <div className="px-4 py-3 border-b border-border flex items-center justify-between">
        <div className="flex items-center gap-2">
          <svg className="w-5 h-5 text-muted-foreground" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
          </svg>
          <h3 className="font-medium">DevPod Sessions</h3>
          {activeSessions.length > 0 && (
            <span className="px-2 py-0.5 text-xs rounded-full bg-green-100 text-green-700">
              {activeSessions.length} active
            </span>
          )}
        </div>
        <Button
          size="sm"
          onClick={() => setShowCreateForm(!showCreateForm)}
          disabled={runners.length === 0}
        >
          <svg className="w-4 h-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
          </svg>
          New Session
        </Button>
      </div>

      {/* Create Form */}
      {showCreateForm && (
        <div className="p-4 border-b border-border bg-muted/30">
          <h4 className="text-sm font-medium mb-3">Create New Session</h4>
          <div className="space-y-3">
            {/* Runner Selection */}
            <div>
              <label className="block text-xs text-muted-foreground mb-1">Runner</label>
              <select
                className="w-full px-3 py-2 text-sm border border-border rounded-md bg-background"
                value={selectedRunner || ""}
                onChange={(e) => setSelectedRunner(Number(e.target.value) || null)}
              >
                <option value="">Select a runner...</option>
                {runners.map((runner) => (
                  <option key={runner.id} value={runner.id}>
                    {runner.node_id} ({runner.current_sessions}{runner.max_concurrent_sessions ? `/${runner.max_concurrent_sessions}` : ""})
                  </option>
                ))}
              </select>
            </div>

            {/* Model Selection */}
            <div>
              <label className="block text-xs text-muted-foreground mb-1">Model</label>
              <select
                className="w-full px-3 py-2 text-sm border border-border rounded-md bg-background"
                value={model}
                onChange={(e) => setModel(e.target.value)}
              >
                <option value="claude-sonnet-4-20250514">Claude Sonnet 4</option>
                <option value="claude-opus-4-20250514">Claude Opus 4</option>
                <option value="claude-3-5-sonnet-20241022">Claude 3.5 Sonnet</option>
              </select>
            </div>

            {/* Permission Mode */}
            <div>
              <label className="block text-xs text-muted-foreground mb-1">Permission Mode</label>
              <select
                className="w-full px-3 py-2 text-sm border border-border rounded-md bg-background"
                value={permissionMode}
                onChange={(e) => setPermissionMode(e.target.value)}
              >
                <option value="default">Default</option>
                <option value="plan">Plan Mode</option>
                <option value="dangerously-skip-permissions">Auto-approve (Dangerous)</option>
              </select>
            </div>

            {/* Initial Prompt */}
            <div>
              <label className="block text-xs text-muted-foreground mb-1">
                Initial Prompt (optional)
              </label>
              <textarea
                className="w-full px-3 py-2 text-sm border border-border rounded-md bg-background resize-none"
                rows={3}
                placeholder={`Work on ticket: ${ticketTitle}`}
                value={initialPrompt}
                onChange={(e) => setInitialPrompt(e.target.value)}
              />
            </div>

            {/* Error */}
            {error && (
              <div className="text-sm text-destructive">{error}</div>
            )}

            {/* Actions */}
            <div className="flex justify-end gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setShowCreateForm(false)}
              >
                Cancel
              </Button>
              <Button
                size="sm"
                onClick={handleCreateSession}
                disabled={!selectedRunner || creating}
              >
                {creating ? "Creating..." : "Create Session"}
              </Button>
            </div>
          </div>
        </div>
      )}

      {/* Sessions List */}
      <div className="divide-y divide-border">
        {/* Active Sessions */}
        {activeSessions.map((session) => (
          <SessionItem key={session.session_key} session={session} />
        ))}

        {/* Inactive Sessions (collapsed by default if there are active ones) */}
        {inactiveSessions.length > 0 && (
          <details className="group">
            <summary className="px-4 py-2 text-sm text-muted-foreground cursor-pointer hover:bg-muted/50">
              {inactiveSessions.length} previous session{inactiveSessions.length !== 1 ? "s" : ""}
            </summary>
            <div className="divide-y divide-border border-t border-border">
              {inactiveSessions.map((session) => (
                <SessionItem key={session.session_key} session={session} />
              ))}
            </div>
          </details>
        )}

        {/* Empty State */}
        {sessions.length === 0 && (
          <div className="px-4 py-8 text-center text-muted-foreground">
            <svg className="w-10 h-10 mx-auto mb-2 text-muted-foreground/50" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
            <p className="text-sm">No sessions for this ticket yet</p>
            {runners.length === 0 && (
              <p className="text-xs mt-1 text-yellow-600">
                No online runners available
              </p>
            )}
          </div>
        )}
      </div>
    </div>
  );
}

function SessionItem({ session }: { session: TicketSession }) {
  const statusInfo = getSessionStatusInfo(session.status);
  const agentInfo = getAgentStatusInfo(session.agent_status);
  const isActive = session.status === "running" || session.status === "initializing";

  return (
    <div className={`px-4 py-3 ${isActive ? "bg-green-50/50" : ""}`}>
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          {/* Status Indicator */}
          <div
            className={`w-2 h-2 rounded-full ${
              session.status === "running"
                ? "bg-green-500 animate-pulse"
                : session.status === "initializing"
                ? "bg-yellow-500 animate-pulse"
                : session.status === "failed"
                ? "bg-red-500"
                : "bg-gray-400"
            }`}
          />

          {/* Session Info */}
          <div>
            <code className="text-xs font-mono text-muted-foreground">
              {session.session_key.substring(0, 12)}...
            </code>
            <div className="flex items-center gap-2 mt-0.5">
              <span
                className={`px-1.5 py-0.5 text-xs rounded ${statusInfo.bgColor} ${statusInfo.color}`}
              >
                {statusInfo.label}
              </span>
              {isActive && (
                <span className={`text-xs flex items-center gap-1 ${agentInfo.color}`}>
                  <span>{agentInfo.icon}</span>
                  {agentInfo.label}
                </span>
              )}
            </div>
          </div>
        </div>

        {/* Actions */}
        <div className="flex items-center gap-2">
          {session.model && (
            <span className="text-xs text-muted-foreground">{session.model}</span>
          )}
          {isActive && (
            <Button size="sm" variant="outline">
              Connect
            </Button>
          )}
        </div>
      </div>

      {/* Started At */}
      {session.started_at && (
        <div className="mt-1 text-xs text-muted-foreground ml-5">
          Started: {new Date(session.started_at).toLocaleString()}
        </div>
      )}
    </div>
  );
}
