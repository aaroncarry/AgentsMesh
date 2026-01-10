"use client";

import { useMemo } from "react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { Badge } from "@/components/ui/badge";
import { TicketStatus, TicketType, TicketPriority } from "@/lib/api/ticket";
import { Search, Filter, X, LayoutList, LayoutGrid } from "lucide-react";

export interface TicketFiltersValue {
  search?: string;
  status?: TicketStatus | null;
  type?: TicketType | null;
  priority?: TicketPriority | null;
  assigneeId?: number | null;
}

interface TicketFiltersProps {
  value: TicketFiltersValue;
  onChange: (filters: TicketFiltersValue) => void;
  viewMode?: "list" | "board";
  onViewModeChange?: (mode: "list" | "board") => void;
  showViewToggle?: boolean;
}

const statusOptions: { value: TicketStatus; label: string }[] = [
  { value: "backlog", label: "Backlog" },
  { value: "todo", label: "To Do" },
  { value: "in_progress", label: "In Progress" },
  { value: "in_review", label: "In Review" },
  { value: "done", label: "Done" },
  { value: "cancelled", label: "Cancelled" },
];

const typeOptions: { value: TicketType; label: string }[] = [
  { value: "task", label: "Task" },
  { value: "bug", label: "Bug" },
  { value: "feature", label: "Feature" },
  { value: "improvement", label: "Improvement" },
  { value: "epic", label: "Epic" },
];

const priorityOptions: { value: TicketPriority; label: string }[] = [
  { value: "urgent", label: "Urgent" },
  { value: "high", label: "High" },
  { value: "medium", label: "Medium" },
  { value: "low", label: "Low" },
  { value: "none", label: "None" },
];

// Helper to get label for a value
const getStatusLabel = (status: string | null | undefined) => {
  if (!status) return null;
  return statusOptions.find((o) => o.value === status)?.label || status;
};

const getTypeLabel = (type: string | null | undefined) => {
  if (!type) return null;
  return typeOptions.find((o) => o.value === type)?.label || type;
};

const getPriorityLabel = (priority: string | null | undefined) => {
  if (!priority) return null;
  return priorityOptions.find((o) => o.value === priority)?.label || priority;
};

export function TicketFilters({
  value,
  onChange,
  viewMode = "list",
  onViewModeChange,
  showViewToggle = true,
}: TicketFiltersProps) {
  const updateFilter = <K extends keyof TicketFiltersValue>(
    key: K,
    newValue: TicketFiltersValue[K]
  ) => {
    onChange({ ...value, [key]: newValue });
  };

  const clearFilters = () => {
    onChange({
      search: "",
      status: null,
      type: null,
      priority: null,
      assigneeId: null,
    });
  };

  // Count active filters (excluding search)
  const activeFilterCount = useMemo(() => {
    let count = 0;
    if (value.status) count++;
    if (value.type) count++;
    if (value.priority) count++;
    if (value.assigneeId) count++;
    return count;
  }, [value]);

  return (
    <div className="flex items-center gap-3 flex-wrap">
      {/* Search Input */}
      <div className="relative flex-1 min-w-[200px] max-w-sm">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
        <Input
          placeholder="Search tickets..."
          value={value.search || ""}
          onChange={(e) => updateFilter("search", e.target.value)}
          className="pl-9"
        />
      </div>

      {/* Desktop Filters */}
      <div className="hidden md:flex items-center gap-2">
        {/* Status Filter */}
        <Select
          value={value.status || ""}
          onValueChange={(val) => updateFilter("status", (val as TicketStatus) || null)}
        >
          <SelectTrigger className="w-[130px]">
            <SelectValue placeholder="Status" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="">All Status</SelectItem>
            {statusOptions.map((opt) => (
              <SelectItem key={opt.value} value={opt.value}>
                {opt.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        {/* Type Filter */}
        <Select
          value={value.type || ""}
          onValueChange={(val) => updateFilter("type", (val as TicketType) || null)}
        >
          <SelectTrigger className="w-[130px]">
            <SelectValue placeholder="Type" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="">All Types</SelectItem>
            {typeOptions.map((opt) => (
              <SelectItem key={opt.value} value={opt.value}>
                {opt.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        {/* Priority Filter */}
        <Select
          value={value.priority || ""}
          onValueChange={(val) => updateFilter("priority", (val as TicketPriority) || null)}
        >
          <SelectTrigger className="w-[130px]">
            <SelectValue placeholder="Priority" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="">All Priority</SelectItem>
            {priorityOptions.map((opt) => (
              <SelectItem key={opt.value} value={opt.value}>
                {opt.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        {/* Clear Filters */}
        {activeFilterCount > 0 && (
          <Button variant="ghost" size="sm" onClick={clearFilters}>
            <X className="h-4 w-4 mr-1" />
            Clear
          </Button>
        )}
      </div>

      {/* Mobile Filter Button */}
      <div className="md:hidden">
        <Popover>
          <PopoverTrigger asChild>
            <Button variant="outline" size="sm" className="relative">
              <Filter className="h-4 w-4" />
              {activeFilterCount > 0 && (
                <Badge
                  variant="destructive"
                  className="absolute -top-2 -right-2 h-5 w-5 p-0 flex items-center justify-center text-xs"
                >
                  {activeFilterCount}
                </Badge>
              )}
            </Button>
          </PopoverTrigger>
          <PopoverContent className="w-72" align="end">
            <div className="space-y-4">
              <div className="space-y-2">
                <label className="text-sm font-medium">Status</label>
                <Select
                  value={value.status || ""}
                  onValueChange={(val) => updateFilter("status", (val as TicketStatus) || null)}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="All Status" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="">All Status</SelectItem>
                    {statusOptions.map((opt) => (
                      <SelectItem key={opt.value} value={opt.value}>
                        {opt.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium">Type</label>
                <Select
                  value={value.type || ""}
                  onValueChange={(val) => updateFilter("type", (val as TicketType) || null)}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="All Types" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="">All Types</SelectItem>
                    {typeOptions.map((opt) => (
                      <SelectItem key={opt.value} value={opt.value}>
                        {opt.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium">Priority</label>
                <Select
                  value={value.priority || ""}
                  onValueChange={(val) => updateFilter("priority", (val as TicketPriority) || null)}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="All Priority" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="">All Priority</SelectItem>
                    {priorityOptions.map((opt) => (
                      <SelectItem key={opt.value} value={opt.value}>
                        {opt.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              {activeFilterCount > 0 && (
                <Button variant="outline" size="sm" className="w-full" onClick={clearFilters}>
                  Clear Filters
                </Button>
              )}
            </div>
          </PopoverContent>
        </Popover>
      </div>

      {/* View Mode Toggle */}
      {showViewToggle && onViewModeChange && (
        <div className="flex border border-border rounded-md overflow-hidden">
          <button
            className={`p-2 ${viewMode === "list" ? "bg-muted" : "hover:bg-muted/50"}`}
            onClick={() => onViewModeChange("list")}
            title="List view"
          >
            <LayoutList className="h-4 w-4" />
          </button>
          <button
            className={`p-2 ${viewMode === "board" ? "bg-muted" : "hover:bg-muted/50"}`}
            onClick={() => onViewModeChange("board")}
            title="Board view"
          >
            <LayoutGrid className="h-4 w-4" />
          </button>
        </div>
      )}
    </div>
  );
}

export default TicketFilters;
