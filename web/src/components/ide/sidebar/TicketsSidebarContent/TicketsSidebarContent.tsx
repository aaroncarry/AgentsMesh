"use client";

import React, { useEffect, useState, useCallback } from "react";
import { useRouter } from "next/navigation";
import { cn } from "@/lib/utils";
import { useAuthStore } from "@/stores/auth";
import { useTicketStore } from "@/stores/ticket";
import { TicketCreateDialog } from "@/components/tickets";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Loader2,
  Plus,
  Search,
  LayoutList,
  LayoutGrid,
  RefreshCw,
} from "lucide-react";
import { useTranslations } from "next-intl";
import { useTicketFilters } from "./useTicketFilters";
import { TicketFilterSection } from "./TicketFilterSection";
import { TicketListItem } from "./TicketListItem";
import type { TicketsSidebarContentProps } from "./types";

/**
 * TicketsSidebarContent - Sidebar content for browsing and filtering tickets
 */
export function TicketsSidebarContent({ className }: TicketsSidebarContentProps) {
  const t = useTranslations();
  const router = useRouter();
  const { currentOrg } = useAuthStore();
  const { loading, viewMode, tickets: allTickets, fetchTickets, setViewMode } = useTicketStore();

  // Filter state and actions
  const {
    searchQuery,
    selectedStatuses,
    selectedTypes,
    selectedPriorities,
    filteredTickets,
    setSearchQuery,
    toggleStatus,
    toggleType,
    togglePriority,
    clearAllFilters,
    hasActiveFilters,
  } = useTicketFilters();

  // Local UI state
  const [refreshing, setRefreshing] = useState(false);
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [statusExpanded, setStatusExpanded] = useState(true);
  const [typeExpanded, setTypeExpanded] = useState(false);
  const [priorityExpanded, setPriorityExpanded] = useState(false);
  const [visibleCount, setVisibleCount] = useState(20);

  // Load tickets on mount
  useEffect(() => {
    if (currentOrg) {
      fetchTickets();
    }
  }, [currentOrg, fetchTickets]);

  // Refresh handler
  const handleRefresh = useCallback(async () => {
    setRefreshing(true);
    try {
      await fetchTickets();
    } finally {
      setRefreshing(false);
    }
  }, [fetchTickets]);

  const handleTicketClick = useCallback((slug: string) => {
    router.push(`/${currentOrg?.slug}/tickets/${slug}`);
  }, [router, currentOrg]);

  // Handle ticket created
  const handleTicketCreated = useCallback((ticketId: number, slug: string) => {
    fetchTickets();
    if (currentOrg?.slug) {
      router.push(`/${currentOrg.slug}/tickets/${slug}`);
    }
  }, [fetchTickets, currentOrg, router]);

  return (
    <div className={cn("flex flex-col h-full", className)}>
      {/* Create Ticket Dialog */}
      <TicketCreateDialog
        open={createDialogOpen}
        onOpenChange={setCreateDialogOpen}
        onCreated={handleTicketCreated}
      />

      {/* Search */}
      <div className="px-2 py-2">
        <div className="relative">
          <Search className="absolute left-2 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <Input
            placeholder={t("tickets.searchPlaceholder")}
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-8 h-8 text-sm"
          />
        </div>
      </div>

      {/* Action buttons */}
      <div className="flex items-center gap-1 px-2 pb-2">
        <Button
          size="sm"
          variant="outline"
          className="flex-1 h-8 text-xs"
          onClick={() => setCreateDialogOpen(true)}
        >
          <Plus className="w-3 h-3 mr-1" />
          {t("tickets.newTicket")}
        </Button>
        <Button
          size="sm"
          variant="ghost"
          className="h-8 w-8 p-0"
          onClick={handleRefresh}
          disabled={refreshing}
        >
          <RefreshCw className={cn("w-4 h-4", refreshing && "animate-spin")} />
        </Button>
      </div>

      {/* View Mode Toggle */}
      <div className="flex items-center gap-1 px-2 pb-2">
        <span className="text-xs text-muted-foreground mr-2">{t("tickets.view")}:</span>
        <div className="flex bg-muted rounded-full p-0.5">
          <button
            className={cn(
              "flex items-center gap-1 px-2 py-1 rounded-full text-xs transition-all",
              viewMode === "list"
                ? "bg-background text-foreground shadow-sm font-medium"
                : "text-muted-foreground hover:text-foreground"
            )}
            onClick={() => setViewMode("list")}
          >
            <LayoutList className="h-3 w-3" />
            {viewMode === "list" && <span>{t("tickets.list.ticket") || "List"}</span>}
          </button>
          <button
            className={cn(
              "flex items-center gap-1 px-2 py-1 rounded-full text-xs transition-all",
              viewMode === "board"
                ? "bg-background text-foreground shadow-sm font-medium"
                : "text-muted-foreground hover:text-foreground"
            )}
            onClick={() => setViewMode("board")}
          >
            <LayoutGrid className="h-3 w-3" />
            {viewMode === "board" && <span>{t("tickets.board")}</span>}
          </button>
        </div>
        {hasActiveFilters && (
          <Button
            size="sm"
            variant="ghost"
            className="h-7 text-xs ml-auto"
            onClick={clearAllFilters}
          >
            {t("tickets.clear")}
          </Button>
        )}
      </div>

      {/* Filters */}
      <TicketFilterSection
        statusExpanded={statusExpanded}
        typeExpanded={typeExpanded}
        priorityExpanded={priorityExpanded}
        onStatusExpandedChange={setStatusExpanded}
        onTypeExpandedChange={setTypeExpanded}
        onPriorityExpandedChange={setPriorityExpanded}
        selectedStatuses={selectedStatuses}
        selectedTypes={selectedTypes}
        selectedPriorities={selectedPriorities}
        onToggleStatus={toggleStatus}
        onToggleType={toggleType}
        onTogglePriority={togglePriority}
        allTickets={allTickets}
        t={t}
      />

      {/* Ticket list preview */}
      <div className="flex-1 overflow-y-auto border-t border-border">
        <div className="px-3 py-2 text-xs text-muted-foreground border-b border-border">
          {filteredTickets.length} {t("tickets.ticketCount")}
        </div>

        {loading ? (
          <div className="flex items-center justify-center py-8">
            <Loader2 className="w-5 h-5 animate-spin text-muted-foreground" />
          </div>
        ) : filteredTickets.length === 0 ? (
          <div className="px-3 py-8 text-center">
            <p className="text-sm text-muted-foreground">
              {hasActiveFilters ? t("tickets.emptyState.noMatch") : t("tickets.emptyState.title")}
            </p>
          </div>
        ) : (
          <div className="py-1">
            {filteredTickets.slice(0, visibleCount).map((ticket) => (
              <TicketListItem
                key={ticket.id}
                ticket={ticket}
                onClick={handleTicketClick}
              />
            ))}
            {filteredTickets.length > visibleCount && (
              <div className="px-2 py-2">
                <Button
                  variant="ghost"
                  size="sm"
                  className="w-full h-7 text-xs text-muted-foreground"
                  onClick={() => setVisibleCount((prev) => prev + 20)}
                >
                  {t("tickets.moreTickets", { count: filteredTickets.length - visibleCount })}
                </Button>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}

export default TicketsSidebarContent;
