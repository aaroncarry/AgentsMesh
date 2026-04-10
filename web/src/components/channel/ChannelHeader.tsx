"use client";

import { Button } from "@/components/ui/button";
import { X, Radio, RefreshCw, LogIn, Lock } from "lucide-react";
import { cn } from "@/lib/utils";
import { ChannelPodManager } from "./ChannelPodManager";
import { ChannelMemberManager } from "./ChannelMemberManager";
import { useChannelStore } from "@/stores/channel";
import { useTranslations } from "next-intl";
import { useState } from "react";

interface ChannelHeaderProps {
  name: string;
  description?: string;
  podCount: number;
  channelId: number;
  visibility?: "public" | "private";
  isMember?: boolean;
  memberCount?: number;
  onClose?: () => void;
  onRefresh?: () => void;
  loading?: boolean;
  compact?: boolean;
  onPodsChanged?: () => void;
}

export function ChannelHeader({
  name,
  description,
  podCount,
  channelId,
  visibility = "public",
  isMember = true,
  memberCount = 0,
  onClose,
  onRefresh,
  loading,
  compact = false,
  onPodsChanged,
}: ChannelHeaderProps) {
  const t = useTranslations();
  const joinUserChannel = useChannelStore((s) => s.joinUserChannel);
  const [joining, setJoining] = useState(false);

  const handleJoin = async () => {
    setJoining(true);
    try {
      await joinUserChannel(channelId);
    } finally {
      setJoining(false);
    }
  };

  const isPrivate = visibility === "private";

  if (compact) {
    return (
      <div className="flex items-center justify-between flex-1 min-w-0">
        <div className="flex items-center gap-2 min-w-0">
          {isPrivate ? (
            <Lock className="w-3.5 h-3.5 text-amber-500 flex-shrink-0" />
          ) : (
            <Radio className="w-3.5 h-3.5 text-blue-500 dark:text-blue-400 flex-shrink-0" />
          )}
          <span className="font-medium text-xs truncate">#{name}</span>
          {isMember && (
            <ChannelPodManager
              channelId={channelId}
              podCount={podCount}
              compact
              onPodsChanged={onPodsChanged}
            />
          )}
        </div>
        <div className="flex items-center gap-1 flex-shrink-0">
          {onRefresh && (
            <Button variant="ghost" size="sm" className="h-6 w-6 p-0" onClick={onRefresh} disabled={loading}>
              <RefreshCw className={cn("w-3.5 h-3.5", loading && "animate-spin")} />
            </Button>
          )}
        </div>
      </div>
    );
  }

  return (
    <div className="flex-shrink-0 border-b border-border">
      <div className="flex items-center justify-between px-4 py-3">
        <div className="flex items-center gap-3 min-w-0">
          <div className={cn(
            "w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0",
            isPrivate ? "bg-amber-500/10" : "bg-blue-500/10"
          )}>
            {isPrivate ? (
              <Lock className="w-4 h-4 text-amber-500" />
            ) : (
              <Radio className="w-4 h-4 text-blue-500 dark:text-blue-400" />
            )}
          </div>
          <div className="min-w-0">
            <h3 className="font-semibold text-sm truncate">#{name}</h3>
            {description && (
              <p className="text-xs text-muted-foreground truncate">{description}</p>
            )}
          </div>
        </div>

        <div className="flex items-center gap-2 flex-shrink-0">
          {isMember && (
            <ChannelMemberManager
              channelId={channelId}
              memberCount={memberCount}
            />
          )}

          {!isMember && !isPrivate && (
            <Button size="sm" onClick={handleJoin} disabled={joining}>
              <LogIn className="w-3.5 h-3.5 mr-1.5" />
              {t("channels.actions.join")}
            </Button>
          )}

          <ChannelPodManager
            channelId={channelId}
            podCount={podCount}
            onPodsChanged={onPodsChanged}
          />

          {onRefresh && (
            <Button variant="ghost" size="icon" className="h-8 w-8" onClick={onRefresh} disabled={loading}>
              <RefreshCw className={cn("w-4 h-4", loading && "animate-spin")} />
            </Button>
          )}

          {onClose && (
            <Button variant="ghost" size="icon" className="h-8 w-8" onClick={onClose}>
              <X className="w-4 h-4" />
            </Button>
          )}
        </div>
      </div>
    </div>
  );
}

export default ChannelHeader;
