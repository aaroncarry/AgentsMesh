"use client";

import { cn } from "@/lib/utils";
import { ChannelInfo } from "@/stores/mesh";
import {
  Radio,
  Loader2,
  ChevronDown,
  ChevronRight,
} from "lucide-react";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";

interface MeshChannelsListProps {
  channels: ChannelInfo[];
  loading: boolean;
  expanded: boolean;
  onToggle: (expanded: boolean) => void;
  selectedChannelId: number | null;
  onChannelClick: (channel: ChannelInfo) => void;
  t: (key: string, params?: Record<string, string | number>) => string;
}

/**
 * Collapsible channels list in mesh sidebar
 */
export function MeshChannelsList({
  channels,
  loading,
  expanded,
  onToggle,
  selectedChannelId,
  onChannelClick,
  t,
}: MeshChannelsListProps) {
  return (
    <Collapsible open={expanded} onOpenChange={onToggle}>
      <CollapsibleTrigger asChild>
        <div className="flex items-center justify-between px-3 py-2 border-t border-border cursor-pointer hover:bg-muted/50">
          <div className="flex items-center gap-2">
            <Radio className="w-4 h-4 text-muted-foreground" />
            <span className="text-sm font-medium">{t("ide.sidebar.mesh.channelsSection")}</span>
            <span className="text-xs text-muted-foreground">
              ({channels.length})
            </span>
          </div>
          {expanded ? (
            <ChevronDown className="w-4 h-4 text-muted-foreground" />
          ) : (
            <ChevronRight className="w-4 h-4 text-muted-foreground" />
          )}
        </div>
      </CollapsibleTrigger>
      <CollapsibleContent>
        <div className="max-h-40 overflow-y-auto">
          {loading && channels.length === 0 ? (
            <div className="flex items-center justify-center py-4">
              <Loader2 className="w-4 h-4 animate-spin text-muted-foreground" />
            </div>
          ) : channels.length === 0 ? (
            <div className="px-3 py-4 text-center text-xs text-muted-foreground">
              {t("ide.sidebar.mesh.noChannels")}
            </div>
          ) : (
            <div className="py-1">
              {channels.map((channel) => {
                const isSelected = selectedChannelId === channel.id;
                return (
                  <div
                    key={channel.id}
                    className={cn(
                      "flex items-center gap-2 px-3 py-1.5 cursor-pointer hover:bg-muted/50",
                      isSelected && "bg-muted/30"
                    )}
                    onClick={() => onChannelClick(channel)}
                  >
                    <Radio className="w-3 h-3 text-blue-500 dark:text-blue-400" />
                    <span className="text-sm truncate flex-1">{channel.name}</span>
                    <span className="text-xs text-muted-foreground">
                      {t("ide.sidebar.mesh.podsCount", { count: channel.pod_keys.length })}
                    </span>
                  </div>
                );
              })}
            </div>
          )}
        </div>
      </CollapsibleContent>
    </Collapsible>
  );
}

export default MeshChannelsList;
