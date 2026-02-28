"use client";

import { useTranslations } from "next-intl";
import { Ticket } from "@/stores/ticket";
import { StatusIcon, TypeIcon, getStatusDisplayInfo } from "../TicketIcons";
import { ChevronRight } from "lucide-react";
import { cn } from "@/lib/utils";

interface SubTicketsListProps {
  subTickets: Ticket[];
  onTicketClick: (slug: string) => void;
  compact?: boolean;
  className?: string;
}

export function SubTicketsList({
  subTickets,
  onTicketClick,
  compact = false,
  className,
}: SubTicketsListProps) {
  const t = useTranslations();

  if (subTickets.length === 0) return null;

  if (compact) {
    return (
      <div className={cn("space-y-2", className)}>
        <label className="text-[11px] font-medium text-muted-foreground/70 uppercase tracking-wider">
          {t("tickets.detail.subTickets")} ({subTickets.length})
        </label>
        <div className="space-y-1">
          {subTickets.map((subTicket) => (
            <button
              key={subTicket.id}
              className="w-full px-2.5 py-1.5 flex items-center gap-2 hover:bg-muted/50 rounded-md transition-colors text-left group"
              onClick={() => onTicketClick(subTicket.slug)}
            >
              <StatusIcon status={subTicket.status} size="sm" />
              <span className="font-mono text-xs text-muted-foreground">
                {subTicket.slug}
              </span>
              <span className="flex-1 truncate text-sm">{subTicket.title}</span>
              <ChevronRight className="h-3.5 w-3.5 text-muted-foreground/50 opacity-0 group-hover:opacity-100 transition-opacity" />
            </button>
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className={className}>
      <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider mb-2">
        {t("tickets.detail.subTickets")} ({subTickets.length})
      </p>
      <div className="rounded-lg border border-border divide-y divide-border overflow-hidden">
        {subTickets.map((subTicket) => {
          const subStatusInfo = getStatusDisplayInfo(subTicket.status);
          return (
            <button
              key={subTicket.id}
              type="button"
              className="w-full text-left px-3 py-2.5 hover:bg-muted/40 transition-colors flex items-center gap-2 group"
              onClick={() => onTicketClick(subTicket.slug)}
            >
              <TypeIcon type={subTicket.type} size="sm" />
              <span className="font-mono text-xs text-muted-foreground">
                {subTicket.slug}
              </span>
              <span className="flex-1 truncate text-sm">{subTicket.title}</span>
              <span className={cn(
                "inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] shrink-0",
                subStatusInfo.bgColor,
                subStatusInfo.color
              )}>
                <StatusIcon status={subTicket.status} size="xs" />
                {subStatusInfo.label}
              </span>
              <ChevronRight className="h-3.5 w-3.5 text-muted-foreground/40 opacity-0 group-hover:opacity-100 transition-opacity shrink-0" />
            </button>
          );
        })}
      </div>
    </div>
  );
}

export default SubTicketsList;
