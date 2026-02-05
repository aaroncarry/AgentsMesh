"use client";

import { Button } from "@/components/ui/button";
import { ChannelHeader } from "@/components/mesh/ChannelHeader";
import { MessageList } from "@/components/mesh/MessageList";
import { MessageInput } from "@/components/mesh/MessageInput";
import { Markdown } from "@/components/ui/markdown";
import { ChevronLeft } from "lucide-react";
import type { ChannelInfo, MeshTopology } from "@/stores/mesh";
import type { TransformedMessage } from "./types";

interface ChannelDetailViewProps {
  channelId: number;
  topology: MeshTopology | null;
  currentChannel: {
    name?: string;
    description?: string;
    document?: string;
    pods?: { pod_key: string }[];
  } | null;
  messages: TransformedMessage[];
  messagesLoading: boolean;
  onBack: () => void;
  onSendMessage: (content: string) => Promise<void>;
  onLoadMore: () => void;
  onRefresh: () => void;
  t: (key: string, params?: Record<string, string | number>) => string;
}

/**
 * Channel detail view with messages and input
 */
export function ChannelDetailView({
  channelId,
  topology,
  currentChannel,
  messages,
  messagesLoading,
  onBack,
  onSendMessage,
  onLoadMore,
  onRefresh,
  t,
}: ChannelDetailViewProps) {
  const channelInfo = topology?.channels.find((c: ChannelInfo) => c.id === channelId);
  const podCount = channelInfo?.pod_keys.length || currentChannel?.pods?.length || 0;

  return (
    <div className="flex flex-col h-full">
      {/* Channel Header with back button - softer styling */}
      <div className="flex items-center gap-2 px-3 py-1.5 bg-muted/30">
        <Button
          variant="ghost"
          size="sm"
          className="h-6 w-6 p-0 hover:bg-muted"
          onClick={onBack}
        >
          <ChevronLeft className="w-4 h-4" />
        </Button>
        <div className="flex-1 min-w-0">
          <ChannelHeader
            name={currentChannel?.name || channelInfo?.name || "Channel"}
            description={currentChannel?.description}
            podCount={podCount}
            onClose={onBack}
            onRefresh={onRefresh}
            loading={messagesLoading}
            compact
          />
        </div>
      </div>

      {/* Document section - collapsible if exists */}
      {currentChannel?.document && (
        <div className="px-3 py-2 bg-muted/20">
          <details className="text-xs">
            <summary className="cursor-pointer text-muted-foreground hover:text-foreground flex items-center gap-1">
              <span>{t("ide.bottomPanel.channelDocument")}</span>
            </summary>
            <div className="mt-2 text-muted-foreground">
              <Markdown content={currentChannel.document} compact />
            </div>
          </details>
        </div>
      )}

      {/* Messages */}
      <div className="flex-1 overflow-hidden">
        <MessageList
          messages={messages}
          loading={messagesLoading}
          hasMore={messages.length >= 50 && messages.length % 50 === 0}
          onLoadMore={onLoadMore}
        />
      </div>

      {/* Input - softer top border */}
      <div className="flex-shrink-0 bg-muted/20">
        <MessageInput
          onSend={onSendMessage}
          placeholder={t("ide.bottomPanel.sendMessagePlaceholder")}
        />
      </div>
    </div>
  );
}

export default ChannelDetailView;
