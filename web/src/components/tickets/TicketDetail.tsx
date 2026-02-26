"use client";

import { useEffect, useCallback, useRef, lazy, Suspense } from "react";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { ConfirmDialog, useConfirmDialog } from "@/components/ui/confirm-dialog";
import { useAuthStore } from "@/stores/auth";
import { useTicketStore, TicketStatus } from "@/stores/ticket";
import { StatusIcon, TypeIcon, getStatusDisplayInfo } from "./TicketIcons";
import TicketPodPanel from "./TicketPodPanel";
import { useTicketExtraData } from "./hooks";
import { SubTicketsList, RelationsList, CommitsList, LabelsList, CommentsList } from "./shared";
import { TicketDetailSidebar } from "./TicketDetailSidebar";
import { InlineEditableText } from "./InlineEditableText";

// Lazy load BlockEditor for inline editing
const BlockEditor = lazy(() => import("@/components/ui/block-editor"));

interface TicketDetailProps {
  slug: string;
}

export function TicketDetail({ slug }: TicketDetailProps) {
  const router = useRouter();
  const t = useTranslations();
  const { currentOrg } = useAuthStore();
  const { currentTicket, fetchTicket, updateTicket, updateTicketStatus, deleteTicket, loading, error } = useTicketStore();

  // Confirm dialog for delete
  const { dialogProps, confirm } = useConfirmDialog();

  // Use shared hook for extra data
  const { subTickets, relations, commits, comments, addComment, updateComment, deleteComment } = useTicketExtraData(slug, !!currentTicket);

  // Debounce timer for content auto-save
  const contentSaveTimerRef = useRef<NodeJS.Timeout | null>(null);

  // Cleanup debounce timer on unmount
  useEffect(() => {
    return () => {
      if (contentSaveTimerRef.current) {
        clearTimeout(contentSaveTimerRef.current);
      }
    };
  }, []);

  // Fetch ticket data
  useEffect(() => {
    fetchTicket(slug);
  }, [slug, fetchTicket]);

  // Handle inline title save
  const handleTitleSave = useCallback(async (newTitle: string) => {
    if (!newTitle.trim()) return;
    try {
      await updateTicket(slug, { title: newTitle });
    } catch (err) {
      console.error("Failed to update title:", err);
      throw err;
    }
  }, [slug, updateTicket]);

  // Handle content change with debounced auto-save
  const handleContentChange = useCallback((newContent: string) => {
    if (contentSaveTimerRef.current) {
      clearTimeout(contentSaveTimerRef.current);
    }
    contentSaveTimerRef.current = setTimeout(async () => {
      try {
        await updateTicket(slug, { content: newContent });
      } catch (err) {
        console.error("Failed to update content:", err);
      }
    }, 800);
  }, [slug, updateTicket]);

  // Handle status change
  const handleStatusChange = async (newStatus: TicketStatus) => {
    try {
      await updateTicketStatus(slug, newStatus);
    } catch (err) {
      console.error("Failed to update status:", err);
    }
  };

  // Handle repository change
  const handleRepositoryChange = async (repositoryId: number | null) => {
    try {
      await updateTicket(slug, { repositoryId });
    } catch (err) {
      console.error("Failed to update repository:", err);
    }
  };

  // Handle delete with confirmation
  const handleDelete = useCallback(async () => {
    const confirmed = await confirm({
      title: t("tickets.detail.deleteTicket"),
      description: t("tickets.detail.deleteConfirmation", { slug }),
      variant: "destructive",
      confirmText: t("common.delete"),
      cancelText: t("common.cancel"),
    });
    if (confirmed) {
      try {
        await deleteTicket(slug);
        router.push(`/${currentOrg?.slug}/tickets`);
      } catch (err) {
        console.error("Failed to delete ticket:", err);
      }
    }
  }, [confirm, deleteTicket, slug, router, currentOrg, t]);

  // Handle ticket click for sub-tickets and relations
  const handleTicketClick = (ticketSlug: string) => {
    router.push(`/${currentOrg?.slug}/tickets/${ticketSlug}`);
  };

  if (loading && !currentTicket) {
    return <TicketDetailSkeleton />;
  }

  if (error) {
    return (
      <div className="text-center py-12">
        <div className="text-red-600 dark:text-red-400 mb-4">{error}</div>
        <Button onClick={() => fetchTicket(slug)}>{t("tickets.detail.retry")}</Button>
      </div>
    );
  }

  if (!currentTicket) {
    return (
      <div className="text-center py-12 text-muted-foreground">
        {t("tickets.detail.notFound")}
      </div>
    );
  }

  const statusInfo = getStatusDisplayInfo(currentTicket.status);

  return (
    <div className="flex flex-col lg:flex-row gap-6">
      {/* Main Content */}
      <div className="flex-1 min-w-0">
        {/* Header */}
        <div className="mb-6">
          <div className="flex items-center gap-2 mb-2">
            <TypeIcon type={currentTicket.type} size="md" />
            <span className="text-muted-foreground font-mono text-sm">
              {currentTicket.slug}
            </span>
            <span className={`flex items-center gap-1 px-2 py-0.5 rounded text-xs font-medium ${statusInfo.bgColor} ${statusInfo.color}`}>
              <StatusIcon status={currentTicket.status} size="xs" />
              {statusInfo.label}
            </span>
          </div>

          {/* Inline editable title */}
          <InlineEditableText
            value={currentTicket.title}
            onSave={handleTitleSave}
            placeholder={t("tickets.createDialog.titlePlaceholder")}
            className="text-2xl font-semibold leading-tight"
            inputClassName="text-2xl font-semibold"
          />

          {/* Always-editable content */}
          <div className="mt-4 border border-border rounded-md overflow-hidden bg-card min-h-[120px]">
            <Suspense fallback={<div className="h-[120px] animate-pulse bg-muted" />}>
              <BlockEditor
                key={slug}
                initialContent={currentTicket.content || ""}
                onChange={handleContentChange}
                editable={true}
              />
            </Suspense>
          </div>
        </div>

        {/* Labels (using shared component) */}
        <LabelsList labels={currentTicket.labels || []} />

        {/* Sub-tickets (using shared component) */}
        <SubTicketsList
          subTickets={subTickets}
          onTicketClick={handleTicketClick}
        />

        {/* Relations (using shared component) */}
        <RelationsList
          relations={relations}
          onTicketClick={handleTicketClick}
        />

        {/* Commits (using shared component) */}
        <CommitsList commits={commits} />

        {/* Comments */}
        <CommentsList
          comments={comments}
          onAddComment={addComment}
          onUpdateComment={updateComment}
          onDeleteComment={deleteComment}
        />

        {/* AgentPods */}
        <TicketPodPanel
          ticketSlug={slug}
          ticketTitle={currentTicket.title}
          ticketId={currentTicket.id}
          repositoryId={currentTicket.repository_id}
        />
      </div>

      {/* Sidebar */}
      <TicketDetailSidebar
        ticket={currentTicket}
        onDelete={handleDelete}
        onStatusChange={handleStatusChange}
        onRepositoryChange={handleRepositoryChange}
        t={t}
      />

      {/* Delete Confirmation Dialog */}
      <ConfirmDialog {...dialogProps} />
    </div>
  );
}

function TicketDetailSkeleton() {
  return (
    <div className="animate-pulse" data-testid="ticket-detail-skeleton">
      <div className="flex flex-col lg:flex-row gap-6">
        <div className="flex-1">
          <div className="h-6 bg-muted rounded w-48 mb-4" />
          <div className="h-10 bg-muted rounded w-3/4 mb-4" />
          <div className="h-24 bg-muted rounded mb-6" />
          <div className="h-40 bg-muted rounded" />
        </div>
        <div className="lg:w-80 space-y-6">
          <div className="h-32 bg-muted rounded" />
          <div className="h-24 bg-muted rounded" />
          <div className="h-40 bg-muted rounded" />
        </div>
      </div>
    </div>
  );
}

export default TicketDetail;
