"use client";

import { useMemo } from "react";
import { Checkbox } from "@/components/ui/checkbox";
import { StatusIcon, TypeIcon, PriorityIcon, getStatusDisplayInfo, getTypeDisplayInfo, getPriorityDisplayInfo } from "@/components/tickets";
import { ChevronDown, ChevronRight } from "lucide-react";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import type { Ticket, TicketStatus, TicketType, TicketPriority } from "@/stores/ticket";
import { statusOptions, typeOptions, priorityOptions } from "./types";

interface FilterSectionProps {
  title: string;
  expanded: boolean;
  onExpandedChange: (expanded: boolean) => void;
  selectedCount: number;
  showBorder?: boolean;
  children: React.ReactNode;
}

/**
 * Generic collapsible filter section
 */
function FilterSection({
  title,
  expanded,
  onExpandedChange,
  selectedCount,
  showBorder = false,
  children,
}: FilterSectionProps) {
  return (
    <Collapsible open={expanded} onOpenChange={onExpandedChange}>
      <CollapsibleTrigger asChild>
        <div className={`flex items-center justify-between px-3 py-2 cursor-pointer hover:bg-muted/50 ${showBorder ? 'border-t border-border' : ''}`}>
          <span className="text-xs font-medium">{title}</span>
          <div className="flex items-center gap-1">
            {selectedCount > 0 && (
              <span className="text-xs bg-primary/10 text-primary px-1.5 rounded">
                {selectedCount}
              </span>
            )}
            {expanded ? (
              <ChevronDown className="w-3.5 h-3.5 text-muted-foreground" />
            ) : (
              <ChevronRight className="w-3.5 h-3.5 text-muted-foreground" />
            )}
          </div>
        </div>
      </CollapsibleTrigger>
      <CollapsibleContent>
        <div className="px-3 pb-2 space-y-1">
          {children}
        </div>
      </CollapsibleContent>
    </Collapsible>
  );
}

interface TicketFilterSectionProps {
  statusExpanded: boolean;
  typeExpanded: boolean;
  priorityExpanded: boolean;
  onStatusExpandedChange: (expanded: boolean) => void;
  onTypeExpandedChange: (expanded: boolean) => void;
  onPriorityExpandedChange: (expanded: boolean) => void;
  selectedStatuses: TicketStatus[];
  selectedTypes: TicketType[];
  selectedPriorities: TicketPriority[];
  onToggleStatus: (status: TicketStatus) => void;
  onToggleType: (type: TicketType) => void;
  onTogglePriority: (priority: TicketPriority) => void;
  allTickets?: Ticket[];
  t: (key: string) => string;
}

/**
 * TicketFilterSection - Renders all filter collapsibles
 */
export function TicketFilterSection({
  statusExpanded,
  typeExpanded,
  priorityExpanded,
  onStatusExpandedChange,
  onTypeExpandedChange,
  onPriorityExpandedChange,
  selectedStatuses,
  selectedTypes,
  selectedPriorities,
  onToggleStatus,
  onToggleType,
  onTogglePriority,
  allTickets,
  t,
}: TicketFilterSectionProps) {
  const statusCounts = useMemo(() => {
    if (!allTickets) return {};
    const counts: Record<string, number> = {};
    for (const ticket of allTickets) {
      counts[ticket.status] = (counts[ticket.status] || 0) + 1;
    }
    return counts;
  }, [allTickets]);

  const typeCounts = useMemo(() => {
    if (!allTickets) return {};
    const counts: Record<string, number> = {};
    for (const ticket of allTickets) {
      counts[ticket.type] = (counts[ticket.type] || 0) + 1;
    }
    return counts;
  }, [allTickets]);

  const priorityCounts = useMemo(() => {
    if (!allTickets) return {};
    const counts: Record<string, number> = {};
    for (const ticket of allTickets) {
      counts[ticket.priority] = (counts[ticket.priority] || 0) + 1;
    }
    return counts;
  }, [allTickets]);

  return (
    <div className="border-t border-border">
      {/* Status Filter */}
      <FilterSection
        title={t("tickets.filters.status")}
        expanded={statusExpanded}
        onExpandedChange={onStatusExpandedChange}
        selectedCount={selectedStatuses.length}
      >
        {statusOptions.map((status) => {
          const info = getStatusDisplayInfo(status, t);
          const count = statusCounts[status];
          return (
            <label
              key={status}
              className="flex items-center gap-2 text-xs cursor-pointer hover:bg-muted/50 px-1 py-0.5 rounded"
            >
              <Checkbox
                checked={selectedStatuses.includes(status)}
                onCheckedChange={() => onToggleStatus(status)}
                className="h-3.5 w-3.5"
              />
              <StatusIcon status={status} size="xs" />
              <span className="flex-1">{info.label}</span>
              {count !== undefined && (
                <span className="text-muted-foreground/60 font-mono">{count}</span>
              )}
            </label>
          );
        })}
      </FilterSection>

      {/* Type Filter */}
      <FilterSection
        title={t("tickets.filters.type")}
        expanded={typeExpanded}
        onExpandedChange={onTypeExpandedChange}
        selectedCount={selectedTypes.length}
        showBorder
      >
        {typeOptions.map((type) => {
          const info = getTypeDisplayInfo(type, t);
          const count = typeCounts[type];
          return (
            <label
              key={type}
              className="flex items-center gap-2 text-xs cursor-pointer hover:bg-muted/50 px-1 py-0.5 rounded"
            >
              <Checkbox
                checked={selectedTypes.includes(type)}
                onCheckedChange={() => onToggleType(type)}
                className="h-3.5 w-3.5"
              />
              <TypeIcon type={type} size="xs" />
              <span className="flex-1">{info.label}</span>
              {count !== undefined && (
                <span className="text-muted-foreground/60 font-mono">{count}</span>
              )}
            </label>
          );
        })}
      </FilterSection>

      {/* Priority Filter */}
      <FilterSection
        title={t("tickets.filters.priority")}
        expanded={priorityExpanded}
        onExpandedChange={onPriorityExpandedChange}
        selectedCount={selectedPriorities.length}
        showBorder
      >
        {priorityOptions.map((priority) => {
          const info = getPriorityDisplayInfo(priority, t);
          const count = priorityCounts[priority];
          return (
            <label
              key={priority}
              className="flex items-center gap-2 text-xs cursor-pointer hover:bg-muted/50 px-1 py-0.5 rounded"
            >
              <Checkbox
                checked={selectedPriorities.includes(priority)}
                onCheckedChange={() => onTogglePriority(priority)}
                className="h-3.5 w-3.5"
              />
              <PriorityIcon priority={priority} size="xs" />
              <span className="flex-1">{info.label}</span>
              {count !== undefined && (
                <span className="text-muted-foreground/60 font-mono">{count}</span>
              )}
            </label>
          );
        })}
      </FilterSection>
    </div>
  );
}

export default TicketFilterSection;
