"use client";

import { useState, useEffect, useCallback } from "react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { repositoryApi, gitProviderApi, RepositoryData, GitProviderData } from "@/lib/api";

export default function RepositoriesPage() {
  const [repositories, setRepositories] = useState<RepositoryData[]>([]);
  const [gitProviders, setGitProviders] = useState<GitProviderData[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState("");
  const [providerFilter, setProviderFilter] = useState<number | "">("");
  const [showCreateModal, setShowCreateModal] = useState(false);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      const [reposRes, providersRes] = await Promise.all([
        repositoryApi.list(),
        gitProviderApi.list(),
      ]);
      setRepositories(reposRes.repositories || []);
      setGitProviders(providersRes.git_providers || []);
    } catch (error) {
      console.error("Failed to load data:", error);
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = useCallback(async (id: number, name: string) => {
    if (!confirm(`Are you sure you want to delete repository "${name}"?`)) {
      return;
    }
    try {
      await repositoryApi.delete(id);
      setRepositories((prev) => prev.filter((r) => r.id !== id));
    } catch (error) {
      console.error("Failed to delete repository:", error);
    }
  }, []);

  const filteredRepositories = repositories.filter((repo) => {
    const matchesSearch =
      repo.name.toLowerCase().includes(filter.toLowerCase()) ||
      repo.full_path.toLowerCase().includes(filter.toLowerCase());
    const matchesProvider = !providerFilter || repo.git_provider_id === providerFilter;
    return matchesSearch && matchesProvider;
  });

  const getProviderIcon = (providerType?: string) => {
    switch (providerType) {
      case "github":
        return (
          <svg className="w-5 h-5" viewBox="0 0 24 24" fill="currentColor">
            <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z" />
          </svg>
        );
      case "gitlab":
        return (
          <svg className="w-5 h-5" viewBox="0 0 24 24" fill="currentColor">
            <path d="M22.65 14.39L12 22.13 1.35 14.39a.84.84 0 01-.3-.94l1.22-3.78 2.44-7.51A.42.42 0 014.82 2a.43.43 0 01.58 0 .42.42 0 01.11.18l2.44 7.49h8.1l2.44-7.51A.42.42 0 0118.6 2a.43.43 0 01.58 0 .42.42 0 01.11.18l2.44 7.51L23 13.45a.84.84 0 01-.35.94z" />
          </svg>
        );
      case "gitee":
        return (
          <svg className="w-5 h-5" viewBox="0 0 24 24" fill="currentColor">
            <path d="M11.984 0A12 12 0 000 12a12 12 0 0012 12 12 12 0 0012-12A12 12 0 0012 0a12 12 0 00-.016 0zm6.09 5.333c.328 0 .593.266.592.593v1.482a.594.594 0 01-.593.592H9.777c-.982 0-1.778.796-1.778 1.778v5.63c0 .327.266.592.593.592h5.63c.982 0 1.778-.796 1.778-1.778v-.296a.593.593 0 00-.592-.593h-4.15a.592.592 0 01-.592-.592v-1.482a.593.593 0 01.593-.592h6.815c.327 0 .593.265.593.592v3.408a4 4 0 01-4 4H5.926a.593.593 0 01-.593-.593V9.778a4.444 4.444 0 014.445-4.444h8.296z" />
          </svg>
        );
      case "ssh":
        return (
          <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
          </svg>
        );
      default:
        return (
          <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
          </svg>
        );
    }
  };

  const getProviderName = (providerId: number) => {
    const provider = gitProviders.find((p) => p.id === providerId);
    return provider?.name || provider?.provider_type || "Unknown";
  };

  const getProviderType = (providerId: number) => {
    const provider = gitProviders.find((p) => p.id === providerId);
    return provider?.provider_type;
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
          <h1 className="text-2xl font-bold text-foreground">Repositories</h1>
          <p className="text-muted-foreground">
            Manage your Git repositories for DevPod sessions
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
          Add Repository
        </Button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-3 gap-4 mb-6">
        <div className="p-4 border border-border rounded-lg bg-card">
          <p className="text-sm text-muted-foreground">Total Repositories</p>
          <p className="text-2xl font-bold">{repositories.length}</p>
        </div>
        <div className="p-4 border border-border rounded-lg bg-card">
          <p className="text-sm text-muted-foreground">Active</p>
          <p className="text-2xl font-bold">
            {repositories.filter((r) => r.is_active).length}
          </p>
        </div>
        <div className="p-4 border border-border rounded-lg bg-card">
          <p className="text-sm text-muted-foreground">Git Providers</p>
          <p className="text-2xl font-bold">{gitProviders.length}</p>
        </div>
      </div>

      {/* Filters */}
      <div className="flex items-center gap-4 mb-6">
        <div className="flex-1 max-w-sm">
          <Input
            placeholder="Search repositories..."
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
            className="w-full"
          />
        </div>
        <select
          className="px-3 py-2 border border-border rounded-md bg-background text-sm"
          value={providerFilter}
          onChange={(e) =>
            setProviderFilter(e.target.value ? Number(e.target.value) : "")
          }
        >
          <option value="">All Providers</option>
          {gitProviders.map((provider) => (
            <option key={provider.id} value={provider.id}>
              {provider.name || provider.provider_type}
            </option>
          ))}
        </select>
      </div>

      {/* Repository List */}
      <div className="grid gap-4">
        {filteredRepositories.map((repo) => (
          <div
            key={repo.id}
            className="p-4 border border-border rounded-lg bg-card hover:border-primary/50 transition-colors"
          >
            <div className="flex items-start justify-between">
              <div className="flex items-start gap-3">
                <div className="mt-1 text-muted-foreground">
                  {getProviderIcon(getProviderType(repo.git_provider_id))}
                </div>
                <div>
                  <div className="flex items-center gap-2">
                    <Link
                      href={`repositories/${repo.id}`}
                      className="font-medium text-foreground hover:text-primary"
                    >
                      {repo.name}
                    </Link>
                    {!repo.is_active && (
                      <span className="px-2 py-0.5 text-xs bg-gray-100 text-gray-600 rounded">
                        Inactive
                      </span>
                    )}
                  </div>
                  <p className="text-sm text-muted-foreground">{repo.full_path}</p>
                  <div className="flex items-center gap-3 mt-2">
                    <span className="inline-flex items-center gap-1 text-xs text-muted-foreground">
                      <svg
                        className="w-3 h-3"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"
                        />
                      </svg>
                      {repo.default_branch}
                    </span>
                    <span className="text-xs text-muted-foreground">
                      {getProviderName(repo.git_provider_id)}
                    </span>
                    {repo.ticket_prefix && (
                      <span className="px-2 py-0.5 text-xs bg-primary/10 text-primary rounded">
                        {repo.ticket_prefix}
                      </span>
                    )}
                  </div>
                </div>
              </div>
              <div className="flex items-center gap-2">
                <Link href={`repositories/${repo.id}`}>
                  <Button variant="outline" size="sm">
                    View
                  </Button>
                </Link>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => handleDelete(repo.id, repo.name)}
                  className="text-destructive hover:text-destructive"
                >
                  <svg
                    className="w-4 h-4"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                    />
                  </svg>
                </Button>
              </div>
            </div>
          </div>
        ))}
        {filteredRepositories.length === 0 && (
          <div className="text-center py-12 text-muted-foreground">
            {repositories.length === 0 ? (
              <>
                <svg
                  className="w-12 h-12 mx-auto mb-4 text-muted-foreground/50"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"
                  />
                </svg>
                <p className="mb-2">No repositories yet</p>
                <p className="text-sm">
                  Add a repository to use Git-based workflows in DevPod
                </p>
              </>
            ) : (
              <p>No repositories match your search</p>
            )}
          </div>
        )}
      </div>

      {/* Create Modal */}
      {showCreateModal && (
        <CreateRepositoryModal
          gitProviders={gitProviders}
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

function CreateRepositoryModal({
  gitProviders,
  onClose,
  onCreated,
}: {
  gitProviders: GitProviderData[];
  onClose: () => void;
  onCreated: () => void;
}) {
  const [selectedProvider, setSelectedProvider] = useState<number | null>(null);
  const [name, setName] = useState("");
  const [fullPath, setFullPath] = useState("");
  const [externalId, setExternalId] = useState("");
  const [defaultBranch, setDefaultBranch] = useState("main");
  const [ticketPrefix, setTicketPrefix] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const selectedProviderData = gitProviders.find((p) => p.id === selectedProvider);
  const isSSHProvider = selectedProviderData?.provider_type === "ssh";

  const handleCreate = async () => {
    if (!selectedProvider || !name || !fullPath) {
      setError("Please fill in all required fields");
      return;
    }

    setLoading(true);
    setError("");

    try {
      await repositoryApi.create({
        git_provider_id: selectedProvider,
        external_id: externalId || fullPath.replace(/[^a-zA-Z0-9]/g, "-"),
        name,
        full_path: fullPath,
        default_branch: defaultBranch || "main",
        ticket_prefix: ticketPrefix || undefined,
      });
      onCreated();
    } catch (err) {
      console.error("Failed to create repository:", err);
      setError("Failed to create repository. Please check your inputs.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-background border border-border rounded-lg w-full max-w-md p-6">
        <h2 className="text-xl font-semibold mb-4">Add Repository</h2>

        {error && (
          <div className="mb-4 p-3 bg-destructive/10 text-destructive text-sm rounded-md">
            {error}
          </div>
        )}

        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium mb-2">
              Git Provider <span className="text-destructive">*</span>
            </label>
            <select
              className="w-full px-3 py-2 border border-border rounded-md bg-background"
              value={selectedProvider || ""}
              onChange={(e) => setSelectedProvider(Number(e.target.value))}
            >
              <option value="">Select a provider...</option>
              {gitProviders.map((provider) => (
                <option key={provider.id} value={provider.id}>
                  {provider.name || provider.provider_type}
                </option>
              ))}
            </select>
            {gitProviders.length === 0 && (
              <p className="text-xs text-muted-foreground mt-1">
                No Git providers configured. Add one in Settings.
              </p>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">
              Repository Name <span className="text-destructive">*</span>
            </label>
            <Input
              placeholder="my-project"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">
              Full Path <span className="text-destructive">*</span>
            </label>
            <Input
              placeholder={isSSHProvider ? "git@github.com:org/repo.git" : "org/my-project"}
              value={fullPath}
              onChange={(e) => setFullPath(e.target.value)}
            />
            <p className="text-xs text-muted-foreground mt-1">
              {isSSHProvider
                ? "SSH clone URL (e.g., git@github.com:org/repo.git)"
                : "Repository path (e.g., org/repo-name)"}
            </p>
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">Default Branch</label>
            <Input
              placeholder="main"
              value={defaultBranch}
              onChange={(e) => setDefaultBranch(e.target.value)}
            />
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">
              Ticket Prefix (optional)
            </label>
            <Input
              placeholder="PROJ"
              value={ticketPrefix}
              onChange={(e) => setTicketPrefix(e.target.value.toUpperCase())}
            />
            <p className="text-xs text-muted-foreground mt-1">
              Used for ticket identifiers (e.g., PROJ-123)
            </p>
          </div>
        </div>

        <div className="flex justify-end gap-3 mt-6">
          <Button variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button
            onClick={handleCreate}
            disabled={!selectedProvider || !name || !fullPath || loading}
          >
            {loading ? "Creating..." : "Add Repository"}
          </Button>
        </div>
      </div>
    </div>
  );
}
