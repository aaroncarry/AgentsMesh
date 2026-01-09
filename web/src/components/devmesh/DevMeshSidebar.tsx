"use client";

import { useState } from "react";
import { useRouter, useParams } from "next/navigation";
import { Button } from "@/components/ui/button";
import {
  useDevMeshStore,
  getSessionStatusInfo,
  getAgentStatusInfo,
  type DevMeshNode,
  type ChannelInfo,
} from "@/stores/devmesh";
import { sessionApi, channelApi } from "@/lib/api/client";

interface DevMeshSidebarProps {
  onClose: () => void;
}

export default function DevMeshSidebar({ onClose }: DevMeshSidebarProps) {
  const { topology, selectedNode, selectedChannel, getNodeByKey, getEdgesForNode, getChannelsForNode } =
    useDevMeshStore();

  const node = selectedNode ? getNodeByKey(selectedNode) : null;
  const channel = selectedChannel
    ? topology?.channels.find((c) => c.id === selectedChannel)
    : null;

  if (!node && !channel) {
    return null;
  }

  return (
    <div className="w-80 border-l border-border bg-background flex flex-col">
      {/* Header */}
      <div className="p-4 border-b border-border flex items-center justify-between">
        <h3 className="font-semibold">
          {node ? "Session Details" : "Channel Details"}
        </h3>
        <Button variant="ghost" size="sm" onClick={onClose}>
          <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </Button>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto p-4">
        {node && <SessionDetails node={node} />}
        {channel && <ChannelDetails channel={channel} />}
      </div>
    </div>
  );
}

function SessionDetails({ node }: { node: DevMeshNode }) {
  const router = useRouter();
  const params = useParams();
  const org = params.org as string;
  const { getEdgesForNode, getChannelsForNode, fetchTopology } = useDevMeshStore();
  const edges = getEdgesForNode(node.session_key);
  const channels = getChannelsForNode(node.session_key);
  const statusInfo = getSessionStatusInfo(node.status);
  const agentInfo = getAgentStatusInfo(node.agent_status);

  const [terminating, setTerminating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleConnectToSession = () => {
    router.push(`/${org}/devpod/${node.session_key}`);
  };

  const handleTerminateSession = async () => {
    if (!confirm("Are you sure you want to terminate this session?")) {
      return;
    }

    setTerminating(true);
    setError(null);
    try {
      await sessionApi.terminate(node.session_key);
      // Refresh topology after termination
      await fetchTopology();
    } catch (err) {
      console.error("Failed to terminate session:", err);
      setError("Failed to terminate session");
    } finally {
      setTerminating(false);
    }
  };

  const isActive = node.status === "running" || node.status === "initializing";

  return (
    <div className="space-y-4">
      {/* Error */}
      {error && (
        <div className="bg-destructive/10 border border-destructive text-destructive px-3 py-2 rounded text-sm">
          {error}
        </div>
      )}

      {/* Session Key */}
      <div>
        <label className="text-xs text-muted-foreground">Session Key</label>
        <code className="block mt-1 text-sm font-mono bg-muted px-2 py-1 rounded break-all">
          {node.session_key}
        </code>
      </div>

      {/* Status */}
      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="text-xs text-muted-foreground">Status</label>
          <div className="mt-1">
            <span className={`px-2 py-1 text-xs rounded-full ${statusInfo.bgColor} ${statusInfo.color}`}>
              {statusInfo.label}
            </span>
          </div>
        </div>
        <div>
          <label className="text-xs text-muted-foreground">Agent</label>
          <div className="mt-1 flex items-center gap-1">
            <span>{agentInfo.icon}</span>
            <span className={`text-sm ${agentInfo.color}`}>{agentInfo.label}</span>
          </div>
        </div>
      </div>

      {/* Model */}
      {node.model && (
        <div>
          <label className="text-xs text-muted-foreground">Model</label>
          <p className="mt-1 text-sm font-medium">{node.model}</p>
        </div>
      )}

      {/* Ticket */}
      {node.ticket_id && (
        <div>
          <label className="text-xs text-muted-foreground">Associated Ticket</label>
          <p className="mt-1">
            <button
              onClick={() => router.push(`/${org}/tickets/${node.ticket_id}`)}
              className="text-sm text-primary hover:underline"
            >
              Ticket #{node.ticket_id}
            </button>
          </p>
        </div>
      )}

      {/* Started At */}
      {node.started_at && (
        <div>
          <label className="text-xs text-muted-foreground">Started</label>
          <p className="mt-1 text-sm">{new Date(node.started_at).toLocaleString()}</p>
        </div>
      )}

      {/* Bindings */}
      <div>
        <label className="text-xs text-muted-foreground mb-2 block">
          Bindings ({edges.length})
        </label>
        {edges.length > 0 ? (
          <div className="space-y-2">
            {edges.map((edge) => (
              <div
                key={edge.id}
                className="p-2 border border-border rounded-md text-xs"
              >
                <div className="flex items-center justify-between mb-1">
                  <span className="font-mono">
                    {edge.source === node.session_key
                      ? `→ ${edge.target.substring(0, 8)}...`
                      : `← ${edge.source.substring(0, 8)}...`}
                  </span>
                  <span
                    className={`px-1.5 py-0.5 rounded ${
                      edge.status === "active"
                        ? "bg-green-100 text-green-700"
                        : "bg-yellow-100 text-yellow-700"
                    }`}
                  >
                    {edge.status}
                  </span>
                </div>
                <div className="text-muted-foreground">
                  Scopes: {edge.granted_scopes.join(", ") || "none"}
                </div>
              </div>
            ))}
          </div>
        ) : (
          <p className="text-sm text-muted-foreground">No bindings</p>
        )}
      </div>

      {/* Channels */}
      <div>
        <label className="text-xs text-muted-foreground mb-2 block">
          Channels ({channels.length})
        </label>
        {channels.length > 0 ? (
          <div className="space-y-2">
            {channels.map((ch) => (
              <button
                key={ch.id}
                onClick={() => router.push(`/${org}/channels/${ch.id}`)}
                className="w-full p-2 border border-border rounded-md flex items-center gap-2 hover:bg-muted transition-colors"
              >
                <span className="text-blue-500">#</span>
                <span className="text-sm">{ch.name}</span>
              </button>
            ))}
          </div>
        ) : (
          <p className="text-sm text-muted-foreground">Not in any channels</p>
        )}
      </div>

      {/* Actions */}
      <div className="pt-4 border-t border-border space-y-2">
        <Button
          className="w-full"
          variant="outline"
          size="sm"
          onClick={handleConnectToSession}
          disabled={!isActive}
        >
          {isActive ? "Connect to Session" : "Session Inactive"}
        </Button>
        <Button
          className="w-full"
          variant="destructive"
          size="sm"
          onClick={handleTerminateSession}
          disabled={terminating || node.status === "terminated"}
        >
          {terminating ? "Terminating..." : "Terminate Session"}
        </Button>
      </div>
    </div>
  );
}

function ChannelDetails({ channel }: { channel: ChannelInfo }) {
  const router = useRouter();
  const params = useParams();
  const org = params.org as string;
  const { topology, fetchTopology, selectChannel } = useDevMeshStore();

  const [archiving, setArchiving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Get session details for sessions in this channel
  const sessionsInChannel = topology?.nodes.filter((n) =>
    channel.session_keys.includes(n.session_key)
  ) || [];

  const handleViewMessages = () => {
    router.push(`/${org}/channels/${channel.id}`);
  };

  const handleArchiveChannel = async () => {
    if (!confirm("Are you sure you want to archive this channel?")) {
      return;
    }

    setArchiving(true);
    setError(null);
    try {
      await channelApi.archive(channel.id);
      // Refresh topology after archiving
      await fetchTopology();
      // Close the sidebar
      selectChannel(null);
    } catch (err) {
      console.error("Failed to archive channel:", err);
      setError("Failed to archive channel");
    } finally {
      setArchiving(false);
    }
  };

  return (
    <div className="space-y-4">
      {/* Error */}
      {error && (
        <div className="bg-destructive/10 border border-destructive text-destructive px-3 py-2 rounded text-sm">
          {error}
        </div>
      )}

      {/* Channel Name */}
      <div>
        <label className="text-xs text-muted-foreground">Channel</label>
        <div className="mt-1 flex items-center gap-2">
          <span className="text-xl text-blue-500">#</span>
          <span className="text-lg font-medium">{channel.name}</span>
          {channel.is_archived && (
            <span className="text-xs bg-muted px-2 py-0.5 rounded">Archived</span>
          )}
        </div>
      </div>

      {/* Description */}
      {channel.description && (
        <div>
          <label className="text-xs text-muted-foreground">Description</label>
          <p className="mt-1 text-sm">{channel.description}</p>
        </div>
      )}

      {/* Stats */}
      <div className="grid grid-cols-2 gap-4">
        <div className="p-3 border border-border rounded-md text-center">
          <p className="text-2xl font-bold">{channel.session_keys.length}</p>
          <p className="text-xs text-muted-foreground">Sessions</p>
        </div>
        <div className="p-3 border border-border rounded-md text-center">
          <p className="text-2xl font-bold">{channel.message_count}</p>
          <p className="text-xs text-muted-foreground">Messages</p>
        </div>
      </div>

      {/* Sessions in Channel */}
      <div>
        <label className="text-xs text-muted-foreground mb-2 block">
          Connected Sessions
        </label>
        {sessionsInChannel.length > 0 ? (
          <div className="space-y-2">
            {sessionsInChannel.map((session) => {
              const statusInfo = getSessionStatusInfo(session.status);
              return (
                <button
                  key={session.session_key}
                  onClick={() => router.push(`/${org}/devpod/${session.session_key}`)}
                  className="w-full p-2 border border-border rounded-md hover:bg-muted transition-colors"
                >
                  <div className="flex items-center justify-between">
                    <code className="text-xs font-mono">
                      {session.session_key.substring(0, 12)}...
                    </code>
                    <span
                      className={`px-1.5 py-0.5 text-xs rounded ${statusInfo.bgColor} ${statusInfo.color}`}
                    >
                      {statusInfo.label}
                    </span>
                  </div>
                </button>
              );
            })}
          </div>
        ) : (
          <p className="text-sm text-muted-foreground">No sessions connected</p>
        )}
      </div>

      {/* Actions */}
      <div className="pt-4 border-t border-border space-y-2">
        <Button
          className="w-full"
          variant="outline"
          size="sm"
          onClick={handleViewMessages}
        >
          View Messages ({channel.message_count})
        </Button>
        {!channel.is_archived && (
          <Button
            className="w-full"
            variant="outline"
            size="sm"
            onClick={handleArchiveChannel}
            disabled={archiving}
          >
            {archiving ? "Archiving..." : "Archive Channel"}
          </Button>
        )}
      </div>
    </div>
  );
}
