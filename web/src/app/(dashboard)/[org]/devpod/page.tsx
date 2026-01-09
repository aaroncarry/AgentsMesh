"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { sessionApi, runnerApi, agentApi, repositoryApi, RepositoryData } from "@/lib/api/client";

interface Session {
  id: number;
  session_key: string;
  status: string;
  agent_status: string;
  created_at: string;
  runner?: {
    node_id: string;
  };
}

interface Runner {
  id: number;
  node_id: string;
  status: string;
  current_sessions: number;
  max_concurrent_sessions?: number;
}

interface AgentType {
  id: number;
  slug: string;
  name: string;
  description?: string;
}

export default function DevPodPage() {
  const router = useRouter();
  const [sessions, setSessions] = useState<Session[]>([]);
  const [runners, setRunners] = useState<Runner[]>([]);
  const [agentTypes, setAgentTypes] = useState<AgentType[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreateModal, setShowCreateModal] = useState(false);

  const handleOpenSession = (sessionKey: string) => {
    router.push(`devpod/${sessionKey}`);
  };

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      const [sessionsRes, runnersRes, agentsRes] = await Promise.all([
        sessionApi.list(),
        runnerApi.list(),
        agentApi.listTypes(),
      ]);
      setSessions(sessionsRes.sessions || []);
      setRunners(runnersRes.runners || []);
      setAgentTypes(agentsRes.agent_types || []);
    } catch (error) {
      console.error("Failed to load data:", error);
    } finally {
      setLoading(false);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case "running":
        return "bg-green-500";
      case "initializing":
        return "bg-yellow-500";
      case "terminated":
        return "bg-gray-500";
      case "failed":
        return "bg-red-500";
      default:
        return "bg-gray-400";
    }
  };

  const handleTerminate = async (sessionKey: string) => {
    try {
      await sessionApi.terminate(sessionKey);
      loadData();
    } catch (error) {
      console.error("Failed to terminate session:", error);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    );
  }

  return (
    <div className="p-6">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-foreground">DevPod</h1>
          <p className="text-muted-foreground">
            Manage your AI development sessions
          </p>
        </div>
        <Button onClick={() => setShowCreateModal(true)}>
          <svg
            className="w-4 h-4 mr-2"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M12 4v16m8-8H4"
            />
          </svg>
          New Session
        </Button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-4 gap-4 mb-6">
        <StatCard
          title="Active Sessions"
          value={sessions.filter((s) => s.status === "running").length}
          icon="terminal"
        />
        <StatCard
          title="Online Runners"
          value={runners.filter((r) => r.status === "online").length}
          icon="server"
        />
        <StatCard
          title="Total Sessions"
          value={sessions.length}
          icon="history"
        />
        <StatCard
          title="Available Agents"
          value={agentTypes.length}
          icon="bot"
        />
      </div>

      {/* Runners */}
      <div className="mb-6">
        <h2 className="text-lg font-semibold mb-3">Runners</h2>
        <div className="grid grid-cols-3 gap-4">
          {runners.map((runner) => (
            <div
              key={runner.id}
              className="p-4 border border-border rounded-lg bg-card"
            >
              <div className="flex items-center justify-between mb-2">
                <span className="font-medium">{runner.node_id}</span>
                <span
                  className={`px-2 py-1 text-xs rounded-full ${
                    runner.status === "online"
                      ? "bg-green-100 text-green-700"
                      : "bg-gray-100 text-gray-700"
                  }`}
                >
                  {runner.status}
                </span>
              </div>
              <div className="text-sm text-muted-foreground">
                {runner.current_sessions} / {runner.max_concurrent_sessions}{" "}
                sessions
              </div>
            </div>
          ))}
          {runners.length === 0 && (
            <div className="col-span-3 text-center py-8 text-muted-foreground">
              No runners registered. Add a runner to get started.
            </div>
          )}
        </div>
      </div>

      {/* Sessions */}
      <div>
        <h2 className="text-lg font-semibold mb-3">Sessions</h2>
        <div className="border border-border rounded-lg overflow-hidden">
          <table className="w-full">
            <thead className="bg-muted">
              <tr>
                <th className="px-4 py-3 text-left text-sm font-medium">
                  Session
                </th>
                <th className="px-4 py-3 text-left text-sm font-medium">
                  Status
                </th>
                <th className="px-4 py-3 text-left text-sm font-medium">
                  Agent Status
                </th>
                <th className="px-4 py-3 text-left text-sm font-medium">
                  Runner
                </th>
                <th className="px-4 py-3 text-left text-sm font-medium">
                  Created
                </th>
                <th className="px-4 py-3 text-right text-sm font-medium">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {sessions.map((session) => (
                <tr key={session.id} className="hover:bg-muted/50">
                  <td className="px-4 py-3">
                    <code className="text-sm bg-muted px-2 py-1 rounded">
                      {session.session_key.substring(0, 8)}...
                    </code>
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-2">
                      <span
                        className={`w-2 h-2 rounded-full ${getStatusColor(
                          session.status
                        )}`}
                      />
                      {session.status}
                    </div>
                  </td>
                  <td className="px-4 py-3 text-muted-foreground">
                    {session.agent_status}
                  </td>
                  <td className="px-4 py-3 text-muted-foreground">
                    {session.runner?.node_id || "-"}
                  </td>
                  <td className="px-4 py-3 text-muted-foreground">
                    {new Date(session.created_at).toLocaleString()}
                  </td>
                  <td className="px-4 py-3 text-right">
                    {session.status === "running" ? (
                      <>
                        <Button
                          size="sm"
                          variant="outline"
                          className="mr-2"
                          onClick={() => handleOpenSession(session.session_key)}
                        >
                          Connect
                        </Button>
                        <Button
                          size="sm"
                          variant="destructive"
                          onClick={() => handleTerminate(session.session_key)}
                        >
                          Terminate
                        </Button>
                      </>
                    ) : (
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => handleOpenSession(session.session_key)}
                      >
                        View
                      </Button>
                    )}
                  </td>
                </tr>
              ))}
              {sessions.length === 0 && (
                <tr>
                  <td
                    colSpan={6}
                    className="px-4 py-8 text-center text-muted-foreground"
                  >
                    No sessions yet. Create one to get started.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Create Modal (simplified) */}
      {showCreateModal && (
        <CreateSessionModal
          agentTypes={agentTypes}
          runners={runners.filter((r) => r.status === "online")}
          onClose={() => setShowCreateModal(false)}
          onCreated={() => {
            setShowCreateModal(false);
            loadData();
          }}
        />
      )}
    </div>
  );
}

function StatCard({
  title,
  value,
  icon,
}: {
  title: string;
  value: number;
  icon: string;
}) {
  return (
    <div className="p-4 border border-border rounded-lg bg-card">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm text-muted-foreground">{title}</p>
          <p className="text-2xl font-bold">{value}</p>
        </div>
        <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center text-primary">
          {icon === "terminal" && (
            <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
          )}
          {icon === "server" && (
            <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />
            </svg>
          )}
          {icon === "history" && (
            <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          )}
          {icon === "bot" && (
            <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
            </svg>
          )}
        </div>
      </div>
    </div>
  );
}

function CreateSessionModal({
  agentTypes,
  runners,
  onClose,
  onCreated,
}: {
  agentTypes: AgentType[];
  runners: Runner[];
  onClose: () => void;
  onCreated: () => void;
}) {
  const [selectedAgent, setSelectedAgent] = useState<number | null>(null);
  const [selectedRunner, setSelectedRunner] = useState<number | null>(null);
  const [prompt, setPrompt] = useState("");
  const [loading, setLoading] = useState(false);

  // Repository and Branch state
  const [repositories, setRepositories] = useState<RepositoryData[]>([]);
  const [selectedRepository, setSelectedRepository] = useState<number | null>(null);
  const [branches, setBranches] = useState<string[]>([]);
  const [selectedBranch, setSelectedBranch] = useState<string>("");
  const [loadingRepos, setLoadingRepos] = useState(true);
  const [loadingBranches, setLoadingBranches] = useState(false);

  // Load repositories on mount
  useEffect(() => {
    const loadRepositories = async () => {
      try {
        const res = await repositoryApi.list();
        setRepositories(res.repositories || []);
      } catch (error) {
        console.error("Failed to load repositories:", error);
      } finally {
        setLoadingRepos(false);
      }
    };
    loadRepositories();
  }, []);

  // Load branches when repository is selected
  useEffect(() => {
    if (!selectedRepository) {
      setBranches([]);
      setSelectedBranch("");
      return;
    }

    const loadBranches = async () => {
      setLoadingBranches(true);
      try {
        const res = await repositoryApi.listBranches(selectedRepository);
        setBranches(res.branches || []);
        // Auto-select default branch
        const repo = repositories.find((r) => r.id === selectedRepository);
        if (repo?.default_branch && res.branches?.includes(repo.default_branch)) {
          setSelectedBranch(repo.default_branch);
        } else if (res.branches?.length > 0) {
          setSelectedBranch(res.branches[0]);
        }
      } catch (error) {
        console.error("Failed to load branches:", error);
        // For SSH providers, branches might not be available
        setBranches([]);
        setSelectedBranch("");
      } finally {
        setLoadingBranches(false);
      }
    };
    loadBranches();
  }, [selectedRepository, repositories]);

  const handleCreate = async () => {
    if (!selectedAgent || !selectedRunner) return;

    setLoading(true);
    try {
      await sessionApi.create({
        agent_type_id: selectedAgent,
        runner_id: selectedRunner,
        repository_id: selectedRepository || undefined,
        branch_name: selectedBranch || undefined,
        initial_prompt: prompt,
      });
      onCreated();
    } catch (error) {
      console.error("Failed to create session:", error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-background border border-border rounded-lg w-full max-w-md p-6 max-h-[90vh] overflow-y-auto">
        <h2 className="text-xl font-semibold mb-4">Create New Session</h2>

        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium mb-2">Agent Type</label>
            <select
              className="w-full px-3 py-2 border border-border rounded-md bg-background"
              value={selectedAgent || ""}
              onChange={(e) => setSelectedAgent(Number(e.target.value))}
            >
              <option value="">Select an agent...</option>
              {agentTypes.map((agent) => (
                <option key={agent.id} value={agent.id}>
                  {agent.name}
                </option>
              ))}
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">Runner</label>
            <select
              className="w-full px-3 py-2 border border-border rounded-md bg-background"
              value={selectedRunner || ""}
              onChange={(e) => setSelectedRunner(Number(e.target.value))}
            >
              <option value="">Select a runner...</option>
              {runners.map((runner) => (
                <option key={runner.id} value={runner.id}>
                  {runner.node_id} ({runner.current_sessions}/
                  {runner.max_concurrent_sessions})
                </option>
              ))}
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">
              Repository (optional)
            </label>
            <select
              className="w-full px-3 py-2 border border-border rounded-md bg-background"
              value={selectedRepository || ""}
              onChange={(e) => setSelectedRepository(e.target.value ? Number(e.target.value) : null)}
              disabled={loadingRepos}
            >
              <option value="">
                {loadingRepos ? "Loading repositories..." : "Select a repository..."}
              </option>
              {repositories.map((repo) => (
                <option key={repo.id} value={repo.id}>
                  {repo.full_path}
                </option>
              ))}
            </select>
            {repositories.length === 0 && !loadingRepos && (
              <p className="text-xs text-muted-foreground mt-1">
                No repositories configured. Add one in Settings → Git Providers.
              </p>
            )}
          </div>

          {selectedRepository && (
            <div>
              <label className="block text-sm font-medium mb-2">Branch</label>
              {loadingBranches ? (
                <div className="w-full px-3 py-2 border border-border rounded-md bg-muted text-muted-foreground">
                  Loading branches...
                </div>
              ) : branches.length > 0 ? (
                <select
                  className="w-full px-3 py-2 border border-border rounded-md bg-background"
                  value={selectedBranch}
                  onChange={(e) => setSelectedBranch(e.target.value)}
                >
                  {branches.map((branch) => (
                    <option key={branch} value={branch}>
                      {branch}
                    </option>
                  ))}
                </select>
              ) : (
                <input
                  type="text"
                  className="w-full px-3 py-2 border border-border rounded-md bg-background"
                  placeholder="Enter branch name (e.g., main)"
                  value={selectedBranch}
                  onChange={(e) => setSelectedBranch(e.target.value)}
                />
              )}
              {branches.length === 0 && !loadingBranches && (
                <p className="text-xs text-muted-foreground mt-1">
                  Branch list unavailable. Enter branch name manually.
                </p>
              )}
            </div>
          )}

          <div>
            <label className="block text-sm font-medium mb-2">
              Initial Prompt (optional)
            </label>
            <textarea
              className="w-full px-3 py-2 border border-border rounded-md bg-background resize-none"
              rows={3}
              placeholder="Enter an initial prompt for the agent..."
              value={prompt}
              onChange={(e) => setPrompt(e.target.value)}
            />
          </div>
        </div>

        <div className="flex justify-end gap-3 mt-6">
          <Button variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button
            onClick={handleCreate}
            disabled={!selectedAgent || !selectedRunner || loading}
          >
            {loading ? "Creating..." : "Create Session"}
          </Button>
        </div>
      </div>
    </div>
  );
}
