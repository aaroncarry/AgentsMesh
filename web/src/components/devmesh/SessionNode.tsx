"use client";

import { memo } from "react";
import { Handle, Position } from "@xyflow/react";
import { getSessionStatusInfo, getAgentStatusInfo, type DevMeshNode } from "@/stores/devmesh";

interface SessionNodeProps {
  data: {
    node: DevMeshNode;
    isSelected?: boolean;
  };
}

function SessionNode({ data }: SessionNodeProps) {
  const { node, isSelected } = data;
  const statusInfo = getSessionStatusInfo(node.status);
  const agentInfo = getAgentStatusInfo(node.agent_status);

  return (
    <div
      className={`px-4 py-3 rounded-lg border-2 bg-background shadow-md min-w-[180px] transition-all ${
        isSelected
          ? "border-primary ring-2 ring-primary/20"
          : "border-border hover:border-primary/50"
      }`}
    >
      {/* Handles for edges */}
      <Handle
        type="target"
        position={Position.Left}
        className="w-3 h-3 !bg-primary"
      />
      <Handle
        type="source"
        position={Position.Right}
        className="w-3 h-3 !bg-primary"
      />

      {/* Session Header */}
      <div className="flex items-center justify-between mb-2">
        <code className="text-xs font-mono text-muted-foreground">
          {node.session_key.substring(0, 8)}...
        </code>
        <span
          className={`px-2 py-0.5 text-xs rounded-full ${statusInfo.bgColor} ${statusInfo.color}`}
        >
          {statusInfo.label}
        </span>
      </div>

      {/* Agent Status */}
      <div className="flex items-center gap-2 mb-2">
        <span className="text-lg">{agentInfo.icon}</span>
        <span className={`text-sm font-medium ${agentInfo.color}`}>
          {agentInfo.label}
        </span>
      </div>

      {/* Model */}
      {node.model && (
        <div className="text-xs text-muted-foreground mb-1">
          Model: <span className="font-medium">{node.model}</span>
        </div>
      )}

      {/* Ticket ID if exists */}
      {node.ticket_id && (
        <div className="text-xs text-muted-foreground">
          Ticket: <span className="font-medium text-primary">#{node.ticket_id}</span>
        </div>
      )}

      {/* Started At */}
      {node.started_at && (
        <div className="text-xs text-muted-foreground mt-1">
          {new Date(node.started_at).toLocaleTimeString()}
        </div>
      )}
    </div>
  );
}

export default memo(SessionNode);
