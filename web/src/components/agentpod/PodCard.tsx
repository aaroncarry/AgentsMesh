"use client";

import { useState } from "react";
import Link from "next/link";
import { Button } from "@/components/ui/button";

interface Pod {
  id: number;
  podKey: string;
  status: "initializing" | "running" | "paused" | "terminated" | "failed";
  agentStatus: string;
  initialPrompt?: string;
  branchName?: string;
  startedAt?: string;
  lastActivity?: string;
  createdAt: string;
  runner?: {
    id: number;
    nodeId: string;
    status: string;
  };
  agentType?: {
    id: number;
    name: string;
    slug: string;
  };
  repository?: {
    id: number;
    name: string;
    fullPath: string;
  };
  ticket?: {
    id: number;
    identifier: string;
    title: string;
  };
  createdBy?: {
    id: number;
    username: string;
    name?: string;
  };
}

interface PodCardProps {
  pod: Pod;
  onTerminate?: (podKey: string) => void;
  onOpen?: (podKey: string) => void;
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

export function PodCard({ pod, onTerminate, onOpen }: PodCardProps) {
  const [isTerminating, setIsTerminating] = useState(false);

  const handleTerminate = async () => {
    if (!onTerminate) return;
    setIsTerminating(true);
    try {
      await onTerminate(pod.podKey);
    } finally {
      setIsTerminating(false);
    }
  };

  const statusStyle = statusColors[pod.status] || statusColors.terminated;
  const isActive = pod.status === "running" || pod.status === "initializing";

  const formatTime = (dateString?: string) => {
    if (!dateString) return "—";
    const date = new Date(dateString);
    return date.toLocaleString();
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

  return (
    <div className="border rounded-lg p-4 bg-card hover:shadow-md transition-shadow">
      {/* Header */}
      <div className="flex items-start justify-between mb-3">
        <div className="flex items-center gap-2">
          <span
            className={`inline-flex items-center gap-1.5 px-2 py-1 rounded-full text-xs font-medium ${statusStyle.bg} ${statusStyle.text}`}
          >
            <span className={`w-1.5 h-1.5 rounded-full ${statusStyle.dot}`} />
            {statusLabels[pod.status]}
          </span>
          {pod.agentType && (
            <span className="text-xs text-muted-foreground bg-muted px-2 py-1 rounded">
              {pod.agentType.name}
            </span>
          )}
        </div>
        <span className="text-xs text-muted-foreground font-mono">
          {pod.podKey.slice(0, 8)}
        </span>
      </div>

      {/* Repository & Branch */}
      {pod.repository && (
        <div className="mb-2">
          <div className="text-sm font-medium truncate">{pod.repository.fullPath}</div>
          {pod.branchName && (
            <div className="text-xs text-muted-foreground flex items-center gap-1 mt-0.5">
              <svg className="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" />
              </svg>
              {pod.branchName}
            </div>
          )}
        </div>
      )}

      {/* Ticket */}
      {pod.ticket && (
        <div className="mb-2">
          <Link
            href={`/tickets/${pod.ticket.identifier}`}
            className="text-xs text-primary hover:underline"
          >
            {pod.ticket.identifier}: {pod.ticket.title}
          </Link>
        </div>
      )}

      {/* Initial Prompt */}
      {pod.initialPrompt && (
        <div className="mb-3">
          <p className="text-xs text-muted-foreground line-clamp-2">
            {pod.initialPrompt}
          </p>
        </div>
      )}

      {/* Runner & Time Info */}
      <div className="flex items-center justify-between text-xs text-muted-foreground mb-3">
        <div className="flex items-center gap-1">
          <svg className="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2" />
          </svg>
          {pod.runner?.nodeId || "Unknown"}
        </div>
        <div>
          {isActive ? (
            <span className="text-green-600">{formatDuration(pod.startedAt)}</span>
          ) : (
            formatTime(pod.startedAt)
          )}
        </div>
      </div>

      {/* Agent Status */}
      {pod.agentStatus && pod.agentStatus !== "unknown" && (
        <div className="mb-3 text-xs">
          <span className="text-muted-foreground">Agent: </span>
          <span className="font-medium">{pod.agentStatus}</span>
        </div>
      )}

      {/* Actions */}
      <div className="flex gap-2">
        {isActive && (
          <>
            <Button
              size="sm"
              variant="default"
              className="flex-1"
              onClick={() => onOpen?.(pod.podKey)}
            >
              Open Terminal
            </Button>
            <Button
              size="sm"
              variant="destructive"
              onClick={handleTerminate}
              disabled={isTerminating}
            >
              {isTerminating ? "..." : "Terminate"}
            </Button>
          </>
        )}
        {!isActive && (
          <Button
            size="sm"
            variant="outline"
            className="flex-1"
            onClick={() => onOpen?.(pod.podKey)}
          >
            View Logs
          </Button>
        )}
      </div>
    </div>
  );
}

export default PodCard;
