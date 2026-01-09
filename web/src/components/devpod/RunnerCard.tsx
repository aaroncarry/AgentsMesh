"use client";

import { Button } from "@/components/ui/button";

interface Runner {
  id: number;
  nodeId: string;
  description?: string;
  status: "online" | "offline" | "maintenance";
  lastHeartbeat?: string;
  currentSessions: number;
  maxConcurrentSessions: number;
  runnerVersion?: string;
  hostInfo?: {
    os?: string;
    arch?: string;
    memory?: number;
    cpuCores?: number;
    hostname?: string;
  };
  createdAt: string;
}

interface RunnerCardProps {
  runner: Runner;
  onDelete?: (id: number) => void;
  onCreateSession?: (runnerId: number) => void;
}

const statusConfig: Record<string, { color: string; bg: string; label: string }> = {
  online: { color: "text-green-600", bg: "bg-green-500", label: "Online" },
  offline: { color: "text-gray-500", bg: "bg-gray-400", label: "Offline" },
  maintenance: { color: "text-yellow-600", bg: "bg-yellow-500", label: "Maintenance" },
};

export function RunnerCard({ runner, onDelete, onCreateSession }: RunnerCardProps) {
  const statusStyle = statusConfig[runner.status] || statusConfig.offline;
  const canCreateSession =
    runner.status === "online" &&
    runner.currentSessions < runner.maxConcurrentSessions;

  const formatMemory = (bytes?: number) => {
    if (!bytes) return "—";
    const gb = bytes / (1024 * 1024 * 1024);
    return `${gb.toFixed(1)} GB`;
  };

  const formatLastSeen = (dateString?: string) => {
    if (!dateString) return "Never";
    const date = new Date(dateString);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffSec = Math.floor(diffMs / 1000);

    if (diffSec < 60) return "Just now";
    if (diffSec < 3600) return `${Math.floor(diffSec / 60)}m ago`;
    if (diffSec < 86400) return `${Math.floor(diffSec / 3600)}h ago`;
    return date.toLocaleDateString();
  };

  return (
    <div className="border rounded-lg p-4 bg-card">
      {/* Header */}
      <div className="flex items-start justify-between mb-4">
        <div>
          <div className="flex items-center gap-2">
            <h3 className="font-medium">{runner.nodeId}</h3>
            <span
              className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium ${statusStyle.color}`}
            >
              <span className={`w-1.5 h-1.5 rounded-full ${statusStyle.bg}`} />
              {statusStyle.label}
            </span>
          </div>
          {runner.description && (
            <p className="text-sm text-muted-foreground mt-1">{runner.description}</p>
          )}
        </div>
        {runner.runnerVersion && (
          <span className="text-xs text-muted-foreground bg-muted px-2 py-1 rounded">
            v{runner.runnerVersion}
          </span>
        )}
      </div>

      {/* Session Capacity */}
      <div className="mb-4">
        <div className="flex justify-between text-sm mb-1">
          <span className="text-muted-foreground">Sessions</span>
          <span>
            {runner.currentSessions} / {runner.maxConcurrentSessions}
          </span>
        </div>
        <div className="w-full h-2 bg-muted rounded-full overflow-hidden">
          <div
            className={`h-full rounded-full transition-all ${
              runner.currentSessions >= runner.maxConcurrentSessions
                ? "bg-red-500"
                : runner.currentSessions >= runner.maxConcurrentSessions * 0.8
                ? "bg-yellow-500"
                : "bg-green-500"
            }`}
            style={{
              width: `${(runner.currentSessions / runner.maxConcurrentSessions) * 100}%`,
            }}
          />
        </div>
      </div>

      {/* Host Info */}
      {runner.hostInfo && (
        <div className="grid grid-cols-2 gap-2 mb-4 text-sm">
          {runner.hostInfo.hostname && (
            <div>
              <span className="text-muted-foreground">Hostname: </span>
              <span>{runner.hostInfo.hostname}</span>
            </div>
          )}
          {runner.hostInfo.os && (
            <div>
              <span className="text-muted-foreground">OS: </span>
              <span>
                {runner.hostInfo.os}
                {runner.hostInfo.arch && ` (${runner.hostInfo.arch})`}
              </span>
            </div>
          )}
          {runner.hostInfo.cpuCores && (
            <div>
              <span className="text-muted-foreground">CPU: </span>
              <span>{runner.hostInfo.cpuCores} cores</span>
            </div>
          )}
          {runner.hostInfo.memory && (
            <div>
              <span className="text-muted-foreground">Memory: </span>
              <span>{formatMemory(runner.hostInfo.memory)}</span>
            </div>
          )}
        </div>
      )}

      {/* Last Seen */}
      <div className="text-xs text-muted-foreground mb-4">
        Last heartbeat: {formatLastSeen(runner.lastHeartbeat)}
      </div>

      {/* Actions */}
      <div className="flex gap-2">
        {canCreateSession && (
          <Button
            size="sm"
            variant="default"
            className="flex-1"
            onClick={() => onCreateSession?.(runner.id)}
          >
            New Session
          </Button>
        )}
        {!canCreateSession && runner.status === "online" && (
          <Button size="sm" variant="outline" className="flex-1" disabled>
            At Capacity
          </Button>
        )}
        {runner.status !== "online" && (
          <Button size="sm" variant="outline" className="flex-1" disabled>
            Offline
          </Button>
        )}
        <Button
          size="sm"
          variant="ghost"
          onClick={() => onDelete?.(runner.id)}
        >
          <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
          </svg>
        </Button>
      </div>
    </div>
  );
}

export default RunnerCard;
