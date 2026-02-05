"use client";

import React from "react";
import { cn } from "@/lib/utils";
import { getPodDisplayName } from "@/lib/pod-utils";
import { MeshNode, ChannelInfo, getPodStatusInfo } from "@/stores/mesh";
import { Button } from "@/components/ui/button";
import { Terminal } from "lucide-react";

interface MeshSelectedDetailsProps {
  selectedNode: MeshNode | null;
  selectedChannel: ChannelInfo | null;
  nodeChannels: ChannelInfo[];
  onOpenTerminal: (podKey: string, e: React.MouseEvent) => void;
  t: (key: string, params?: Record<string, string | number>) => string;
}

/**
 * Details panel for selected node or channel in mesh sidebar
 */
export function MeshSelectedDetails({
  selectedNode,
  selectedChannel,
  nodeChannels,
  onOpenTerminal,
  t,
}: MeshSelectedDetailsProps) {
  if (!selectedNode && !selectedChannel) {
    return null;
  }

  return (
    <div className="border-t border-border p-3 space-y-2">
      <div className="text-xs font-medium text-muted-foreground">
        {selectedNode ? t("ide.sidebar.mesh.selectedPod") : t("ide.sidebar.mesh.selectedChannel")}
      </div>

      {selectedNode && (
        <div className="space-y-2">
          <div className="text-sm font-medium truncate">{getPodDisplayName(selectedNode)}</div>
          <div className="flex items-center gap-2 text-xs">
            <span className={cn("px-1.5 py-0.5 rounded", getPodStatusInfo(selectedNode.status).bgColor, getPodStatusInfo(selectedNode.status).color)}>
              {getPodStatusInfo(selectedNode.status).label}
            </span>
            {selectedNode.model && (
              <span className="text-muted-foreground">{selectedNode.model}</span>
            )}
          </div>
          {nodeChannels.length > 0 && (
            <div className="text-xs text-muted-foreground">
              {t("ide.sidebar.mesh.channelsLabel")}: {nodeChannels.map(c => c.name).join(", ")}
            </div>
          )}
          <div className="flex gap-2">
            <Button
              size="sm"
              variant="outline"
              className="h-7 text-xs flex-1"
              onClick={(e) => onOpenTerminal(selectedNode.pod_key, e)}
            >
              <Terminal className="w-3 h-3 mr-1" />
              {t("ide.sidebar.mesh.terminal")}
            </Button>
          </div>
        </div>
      )}

      {selectedChannel && (
        <div className="space-y-2">
          <div className="text-sm font-medium">{selectedChannel.name}</div>
          <div className="text-xs text-muted-foreground">
            {t("ide.sidebar.mesh.connectedPods", { count: selectedChannel.pod_keys.length })}
          </div>
          <div className="text-xs text-muted-foreground break-all">
            {t("ide.sidebar.mesh.podsLabel")}: {selectedChannel.pod_keys.map(k => k.substring(0, 8)).join(", ")}
          </div>
        </div>
      )}
    </div>
  );
}

export default MeshSelectedDetails;
