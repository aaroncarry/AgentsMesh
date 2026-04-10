"use client";

import { useState, useCallback } from "react";
import { Markdown } from "@/components/ui/markdown";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { useTranslations } from "next-intl";
import { MessageActions } from "./MessageActions";
import type { TransformedMessage } from "./types";

interface MessageBubbleProps {
  message: TransformedMessage;
  isFirstInGroup: boolean;
  formatTime: (dateString: string) => string;
  currentUserId?: number;
  onEdit?: (messageId: number, content: string) => Promise<void>;
  onDelete?: (messageId: number) => Promise<void>;
}

export function MessageBubble({
  message, isFirstInGroup, formatTime, currentUserId, onEdit, onDelete,
}: MessageBubbleProps) {
  const t = useTranslations("channels.messages");
  const [editing, setEditing] = useState(false);
  const [editContent, setEditContent] = useState(message.content);

  const isOwnMessage = currentUserId != null && message.user?.id === currentUserId;

  const handleEditSubmit = useCallback(async () => {
    if (!onEdit || editContent.trim() === "" || editContent === message.content) {
      setEditing(false);
      return;
    }
    try { await onEdit(message.id, editContent.trim()); setEditing(false); }
    catch { /* Error handled by store */ }
  }, [onEdit, editContent, message.id, message.content]);

  const handleEditKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.nativeEvent.isComposing) return;
      if (e.key === "Enter" && !e.shiftKey) { e.preventDefault(); handleEditSubmit(); }
      if (e.key === "Escape") { setEditing(false); setEditContent(message.content); }
    },
    [handleEditSubmit, message.content]
  );

  const isCode = message.messageType === "code";
  const isCommand = message.messageType === "command";

  return (
    <div className="group/msg relative">
      <MessageActions
        messageId={message.id}
        content={message.content}
        isOwnMessage={isOwnMessage}
        onEdit={onEdit}
        onDelete={onDelete}
        onStartEdit={() => { setEditing(true); setEditContent(message.content); }}
      />

      <div className="flex items-start gap-3">
        {!isFirstInGroup && (
          <span className="w-8 flex-shrink-0 text-[10px] text-muted-foreground opacity-0 group-hover/msg:opacity-100 transition-opacity pt-1 text-center tabular-nums">
            {formatTime(message.createdAt)}
          </span>
        )}

        <div className="flex-1 min-w-0">
          {editing ? (
            <div className="space-y-2">
              <Textarea
                value={editContent}
                onChange={(e) => setEditContent(e.target.value)}
                onKeyDown={handleEditKeyDown}
                className="min-h-[60px] text-sm"
                autoFocus
              />
              <div className="flex items-center gap-2 text-xs text-muted-foreground">
                <Button size="sm" variant="default" className="h-6 text-xs" onClick={handleEditSubmit}>
                  {t("save")}
                </Button>
                <Button size="sm" variant="ghost" className="h-6 text-xs"
                  onClick={() => { setEditing(false); setEditContent(message.content); }}>
                  {t("cancel")}
                </Button>
                <span>{t("editHint")}</span>
              </div>
            </div>
          ) : isCode ? (
            <pre className="p-3 bg-muted rounded-md text-sm overflow-x-auto"><code>{message.content}</code></pre>
          ) : isCommand ? (
            <div className="p-2 bg-muted rounded-md text-sm font-mono text-green-600 dark:text-green-400">$ {message.content}</div>
          ) : (
            <div>
              <Markdown content={message.content} compact highlightMentions
                className="text-sm [&_p:first-child]:mt-0 [&_p:last-child]:mb-0" />
              {message.editedAt && (
                <span className="text-[10px] text-muted-foreground ml-1">({t("edited")})</span>
              )}
            </div>
          )}

          {message.metadata && Object.keys(message.metadata).length > 0 && (
            <div className="mt-1.5 text-xs text-muted-foreground">
              <details>
                <summary className="cursor-pointer hover:text-foreground">{t("metadata")}</summary>
                <pre className="mt-1 p-2 bg-muted rounded text-xs overflow-x-auto">
                  {JSON.stringify(message.metadata, null, 2)}
                </pre>
              </details>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

export default MessageBubble;
