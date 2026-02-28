"use client";

import { useState, useCallback } from "react";
import { useTranslations } from "next-intl";
import { MessageSquare, Reply, Pencil, Trash2 } from "lucide-react";
import { TicketComment } from "@/lib/api";
import { useAuthStore } from "@/stores/auth";
import { CommentInput } from "./CommentInput";

interface CommentsListProps {
  comments: TicketComment[];
  onAddComment: (
    content: string,
    parentId?: number,
    mentions?: Array<{ user_id: number; username: string }>
  ) => Promise<void>;
  onUpdateComment: (
    commentId: number,
    content: string,
    mentions?: Array<{ user_id: number; username: string }>
  ) => Promise<void>;
  onDeleteComment: (commentId: number) => Promise<void>;
  className?: string;
}

export function CommentsList({
  comments,
  onAddComment,
  onUpdateComment,
  onDeleteComment,
  className,
}: CommentsListProps) {
  const t = useTranslations();
  const { user } = useAuthStore();
  const [replyTo, setReplyTo] = useState<{
    id: number;
    username: string;
  } | null>(null);
  const [editingId, setEditingId] = useState<number | null>(null);

  const handleAddComment = useCallback(
    async (
      content: string,
      mentions: Array<{ user_id: number; username: string }>
    ) => {
      await onAddComment(content, replyTo?.id, mentions);
      setReplyTo(null);
    },
    [onAddComment, replyTo]
  );

  const handleSaveEdit = useCallback(
    async (
      content: string,
      mentions: Array<{ user_id: number; username: string }>
    ) => {
      if (editingId === null) return;
      await onUpdateComment(editingId, content, mentions);
      setEditingId(null);
    },
    [editingId, onUpdateComment]
  );

  const handleDelete = async (commentId: number) => {
    if (!window.confirm(t("tickets.detail.deleteCommentConfirm"))) return;
    await onDeleteComment(commentId);
  };

  const formatRelativeDate = (dateString: string) => {
    const date = new Date(dateString);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMin = Math.floor(diffMs / 60000);
    const diffHr = Math.floor(diffMin / 60);
    const diffDay = Math.floor(diffHr / 24);

    if (diffDay > 7) {
      return date.toLocaleDateString(undefined, {
        month: "short",
        day: "numeric",
        ...(date.getFullYear() !== now.getFullYear() ? { year: "numeric" } : {}),
      });
    }
    if (diffDay > 0) return `${diffDay}d ago`;
    if (diffHr > 0) return `${diffHr}h ago`;
    if (diffMin > 0) return `${diffMin}m ago`;
    return "just now";
  };

  const renderContent = (content: string) => {
    const parts = content.split(/(@\w+)/g);
    return parts.map((part, i) => {
      if (part.startsWith("@")) {
        return (
          <span
            key={i}
            className="text-primary font-medium bg-primary/10 rounded px-0.5"
          >
            {part}
          </span>
        );
      }
      return <span key={i}>{part}</span>;
    });
  };

  const renderComment = (comment: TicketComment, isReply = false) => {
    const isAuthor = user?.id === comment.user_id;
    const isEdited =
      comment.updated_at !== comment.created_at &&
      new Date(comment.updated_at).getTime() -
        new Date(comment.created_at).getTime() >
        1000;

    return (
      <div
        key={comment.id}
        className={`group ${isReply ? "ml-8 border-l-2 border-border pl-4" : ""}`}
      >
        <div className="flex items-start gap-3 py-3">
          {comment.user?.avatar_url ? (
            // eslint-disable-next-line @next/next/no-img-element
            <img
              src={comment.user.avatar_url}
              alt=""
              className="w-7 h-7 rounded-full shrink-0 mt-0.5"
            />
          ) : (
            <div className="w-7 h-7 rounded-full bg-primary/10 flex items-center justify-center text-xs font-medium text-primary shrink-0 mt-0.5">
              {(comment.user?.username || "?")[0].toUpperCase()}
            </div>
          )}

          <div className="flex-1 min-w-0">
            <div className="flex items-baseline gap-2 mb-0.5">
              <span className="text-sm font-medium">
                {comment.user?.name || comment.user?.username || "Unknown"}
              </span>
              <span className="text-xs text-muted-foreground" title={new Date(comment.created_at).toLocaleString()}>
                {formatRelativeDate(comment.created_at)}
              </span>
              {isEdited && (
                <span className="text-[10px] text-muted-foreground/60 italic">
                  ({t("tickets.detail.edited")})
                </span>
              )}
            </div>

            {editingId === comment.id ? (
              <CommentInput
                initialContent={comment.content}
                onSubmit={handleSaveEdit}
                onCancel={() => setEditingId(null)}
              />
            ) : (
              <div className="text-sm text-foreground/90 whitespace-pre-wrap leading-relaxed">
                {renderContent(comment.content)}
              </div>
            )}

            {editingId !== comment.id && (
              <div className="flex items-center gap-0.5 mt-1 opacity-0 group-hover:opacity-100 transition-opacity">
                {!isReply && (
                  <button
                    type="button"
                    onClick={() =>
                      setReplyTo({
                        id: comment.id,
                        username: comment.user?.username || "unknown",
                      })
                    }
                    className="flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground transition-colors px-1.5 py-0.5 rounded hover:bg-muted/50"
                  >
                    <Reply className="w-3 h-3" />
                    {t("tickets.detail.reply")}
                  </button>
                )}
                {isAuthor && (
                  <>
                    <button
                      type="button"
                      onClick={() => setEditingId(comment.id)}
                      className="flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground transition-colors px-1.5 py-0.5 rounded hover:bg-muted/50"
                    >
                      <Pencil className="w-3 h-3" />
                      {t("tickets.detail.edit")}
                    </button>
                    <button
                      type="button"
                      onClick={() => handleDelete(comment.id)}
                      className="flex items-center gap-1 text-xs text-destructive/70 hover:text-destructive transition-colors px-1.5 py-0.5 rounded hover:bg-destructive/5"
                    >
                      <Trash2 className="w-3 h-3" />
                    </button>
                  </>
                )}
              </div>
            )}
          </div>
        </div>

        {comment.replies?.map((reply) => renderComment(reply, true))}
      </div>
    );
  };

  return (
    <div className={className}>
      {/* Comment list */}
      {comments.length > 0 ? (
        <div className="divide-y divide-border mb-4">
          {comments.map((comment) => renderComment(comment))}
        </div>
      ) : (
        <div className="flex flex-col items-center justify-center py-10 mb-4 text-muted-foreground">
          <MessageSquare className="w-8 h-8 mb-2 text-muted-foreground/25" />
          <p className="text-sm">{t("tickets.detail.noComments")}</p>
        </div>
      )}

      {/* Comment input */}
      <CommentInput
        onSubmit={handleAddComment}
        replyTo={replyTo || undefined}
        onCancelReply={() => setReplyTo(null)}
      />
    </div>
  );
}

export default CommentsList;
