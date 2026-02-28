"use client";

import { useEffect, useCallback, useRef, lazy, Suspense } from "react";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ConfirmDialog, useConfirmDialog } from "@/components/ui/confirm-dialog";
import { useAuthStore } from "@/stores/auth";
import { useTicketStore, TicketStatus, TicketPriority } from "@/stores/ticket";
import { StatusIcon, TypeIcon, getStatusDisplayInfo } from "./TicketIcons";
import { useTicketExtraData } from "./hooks";
import { SubTicketsList, RelationsList, CommitsList, LabelsList, CommentsList } from "./shared";
import { TicketDetailSidebar } from "./TicketDetailSidebar";
import { InlineEditableText } from "./InlineEditableText";
import { MessageSquare, GitBranch, FileText } from "lucide-react";

const BlockEditor = lazy(() => import("@/components/ui/block-editor"));

interface TicketDetailProps {
  slug: string;
}

export function TicketDetail({ slug }: TicketDetailProps) {
  const router = useRouter();
  const t = useTranslations();
  const { currentOrg } = useAuthStore();
  const { currentTicket, fetchTicket, updateTicket, updateTicketStatus, deleteTicket, loading, error } = useTicketStore();

  const { dialogProps, confirm } = useConfirmDialog();
  const { subTickets, relations, commits, comments, addComment, updateComment, deleteComment } = useTicketExtraData(slug, !!currentTicket);

  const contentSaveTimerRef = useRef<NodeJS.Timeout | null>(null);

  useEffect(() => {
    return () => {
      if (contentSaveTimerRef.current) {
        clearTimeout(contentSaveTimerRef.current);
      }
    };
  }, []);

  useEffect(() => {
    fetchTicket(slug);
  }, [slug, fetchTicket]);

  const handleTitleSave = useCallback(async (newTitle: string) => {
    if (!newTitle.trim()) return;
    try {
      await updateTicket(slug, { title: newTitle });
    } catch (err) {
      console.error("Failed to update title:", err);
      throw err;
    }
  }, [slug, updateTicket]);

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

  const handleStatusChange = async (newStatus: TicketStatus) => {
    try {
      await updateTicketStatus(slug, newStatus);
    } catch (err) {
      console.error("Failed to update status:", err);
    }
  };

  const handlePriorityChange = async (newPriority: TicketPriority) => {
    try {
      await updateTicket(slug, { priority: newPriority });
    } catch (err) {
      console.error("Failed to update priority:", err);
    }
  };

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

  const handleTicketClick = (ticketSlug: string) => {
    router.push(`/${currentOrg?.slug}/tickets/${ticketSlug}`);
  };

  if (loading && !currentTicket) {
    return <TicketDetailSkeleton />;
  }

  if (error) {
    return (
      <div className="text-center py-16">
        <div className="text-destructive mb-4 text-sm">{error}</div>
        <Button variant="outline" size="sm" onClick={() => fetchTicket(slug)}>
          {t("tickets.detail.retry")}
        </Button>
      </div>
    );
  }

  if (!currentTicket) {
    return (
      <div className="text-center py-16 text-muted-foreground text-sm">
        {t("tickets.detail.notFound")}
      </div>
    );
  }

  const statusInfo = getStatusDisplayInfo(currentTicket.status, t);
  const linkedCount = subTickets.length + relations.length + commits.length;

  return (
    <div className="flex flex-col lg:flex-row gap-8">
      {/* Main Content */}
      <div className="flex-1 min-w-0">
        {/* Header */}
        <div className="mb-6">
          {/* Meta row: type icon + slug + status badge */}
          <div className="flex items-center gap-2 mb-3">
            <TypeIcon type={currentTicket.type} size="md" />
            <span className="text-muted-foreground font-mono text-sm">
              {currentTicket.slug}
            </span>
            <span className="mx-1 text-border">·</span>
            <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium ${statusInfo.bgColor} ${statusInfo.color}`}>
              <StatusIcon status={currentTicket.status} size="xs" />
              {statusInfo.label}
            </span>
          </div>

          {/* Title */}
          <InlineEditableText
            value={currentTicket.title}
            onSave={handleTitleSave}
            placeholder={t("tickets.createDialog.titlePlaceholder")}
            className="text-xl sm:text-2xl font-semibold leading-snug"
            inputClassName="text-xl sm:text-2xl font-semibold"
          />

          {/* Labels (inline with header) */}
          {currentTicket.labels && currentTicket.labels.length > 0 && (
            <div className="mt-3">
              <LabelsList labels={currentTicket.labels} compact />
            </div>
          )}
        </div>

        {/* Tabs */}
        <Tabs defaultValue="content" className="mt-2">
          <TabsList className="w-full justify-start border-b border-border rounded-none bg-transparent p-0 h-auto gap-0">
            <TabsTrigger
              value="content"
              className="rounded-none border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:bg-transparent data-[state=active]:shadow-none px-4 py-2 text-sm font-medium text-muted-foreground data-[state=active]:text-foreground"
            >
              <FileText className="w-3.5 h-3.5 mr-1.5" />
              {t("tickets.detail.content") || "Content"}
            </TabsTrigger>
            <TabsTrigger
              value="activity"
              className="rounded-none border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:bg-transparent data-[state=active]:shadow-none px-4 py-2 text-sm font-medium text-muted-foreground data-[state=active]:text-foreground"
            >
              <MessageSquare className="w-3.5 h-3.5 mr-1.5" />
              {t("tickets.detail.comments")}
              {comments.length > 0 && (
                <span className="ml-1.5 text-[10px] bg-muted px-1.5 py-0.5 rounded-full text-muted-foreground tabular-nums">
                  {comments.length}
                </span>
              )}
            </TabsTrigger>
            <TabsTrigger
              value="linked"
              className="rounded-none border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:bg-transparent data-[state=active]:shadow-none px-4 py-2 text-sm font-medium text-muted-foreground data-[state=active]:text-foreground"
            >
              <GitBranch className="w-3.5 h-3.5 mr-1.5" />
              {t("tickets.detail.linked") || "Linked"}
              {linkedCount > 0 && (
                <span className="ml-1.5 text-[10px] bg-muted px-1.5 py-0.5 rounded-full text-muted-foreground tabular-nums">
                  {linkedCount}
                </span>
              )}
            </TabsTrigger>
          </TabsList>

          {/* Content tab */}
          <TabsContent value="content" className="mt-4 focus-visible:outline-none focus-visible:ring-0">
            <div className="rounded-lg border border-border overflow-hidden bg-card min-h-[150px] max-h-[60vh] overflow-y-auto">
              <Suspense fallback={<div className="h-[150px] animate-pulse bg-muted/50" />}>
                <BlockEditor
                  key={slug}
                  initialContent={currentTicket.content || ""}
                  onChange={handleContentChange}
                  editable={true}
                />
              </Suspense>
            </div>
          </TabsContent>

          {/* Activity tab */}
          <TabsContent value="activity" className="mt-4 focus-visible:outline-none focus-visible:ring-0">
            <CommentsList
              comments={comments}
              onAddComment={addComment}
              onUpdateComment={updateComment}
              onDeleteComment={deleteComment}
            />
          </TabsContent>

          {/* Linked tab */}
          <TabsContent value="linked" className="mt-4 focus-visible:outline-none focus-visible:ring-0">
            <div className="space-y-1">
              <SubTicketsList
                subTickets={subTickets}
                onTicketClick={handleTicketClick}
              />
              <RelationsList
                relations={relations}
                onTicketClick={handleTicketClick}
              />
              <CommitsList commits={commits} />
              {linkedCount === 0 && (
                <div className="flex flex-col items-center justify-center py-12 text-muted-foreground">
                  <GitBranch className="w-8 h-8 mb-3 text-muted-foreground/30" />
                  <p className="text-sm">{t("tickets.detail.noLinkedItems") || "No linked items yet."}</p>
                </div>
              )}
            </div>
          </TabsContent>
        </Tabs>

      </div>

      {/* Sidebar */}
      <TicketDetailSidebar
        ticket={currentTicket}
        onDelete={handleDelete}
        onStatusChange={handleStatusChange}
        onPriorityChange={handlePriorityChange}
        ticketSlug={slug}
        t={t}
      />

      <ConfirmDialog {...dialogProps} />
    </div>
  );
}

function TicketDetailSkeleton() {
  return (
    <div className="animate-pulse" data-testid="ticket-detail-skeleton">
      <div className="flex flex-col lg:flex-row gap-8">
        <div className="flex-1">
          <div className="flex items-center gap-2 mb-3">
            <div className="h-5 w-5 bg-muted rounded" />
            <div className="h-4 w-24 bg-muted rounded" />
            <div className="h-5 w-20 bg-muted rounded-full" />
          </div>
          <div className="h-8 bg-muted rounded w-3/4 mb-6" />
          <div className="h-10 bg-muted rounded w-full mb-4" />
          <div className="h-48 bg-muted rounded" />
        </div>
        <div className="lg:w-72 space-y-3">
          <div className="h-16 bg-muted rounded-lg" />
          <div className="h-40 bg-muted rounded-lg" />
          <div className="h-20 bg-muted rounded-lg" />
        </div>
      </div>
    </div>
  );
}

export default TicketDetail;
