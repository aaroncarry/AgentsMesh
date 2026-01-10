"use client";

import { useState, useEffect, useCallback, useMemo } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { KanbanBoard, TicketFilters, TicketFiltersValue, TicketCreateDialog } from "@/components/tickets";
import { ticketApi } from "@/lib/api/client";
import { Plus } from "lucide-react";

interface Ticket {
  id: number;
  number: number;
  identifier: string;
  title: string;
  description?: string;
  status: "backlog" | "todo" | "in_progress" | "in_review" | "done" | "cancelled";
  priority: "none" | "low" | "medium" | "high" | "urgent";
  type: "task" | "bug" | "feature" | "epic";
  created_at: string;
  assignees?: Array<{ id: number; username: string; name?: string; avatarUrl?: string }>;
  labels?: Array<{ id: number; name: string; color: string }>;
  repository?: { id: number; name: string };
}

const statusColors: Record<string, string> = {
  backlog: "bg-gray-100 text-gray-700",
  todo: "bg-blue-100 text-blue-700",
  in_progress: "bg-yellow-100 text-yellow-700",
  in_review: "bg-purple-100 text-purple-700",
  done: "bg-green-100 text-green-700",
  cancelled: "bg-red-100 text-red-700",
};

const priorityColors: Record<string, string> = {
  none: "text-gray-400",
  low: "text-blue-500",
  medium: "text-yellow-500",
  high: "text-orange-500",
  urgent: "text-red-500",
};

export default function TicketsPage() {
  const router = useRouter();
  const [tickets, setTickets] = useState<Ticket[]>([]);
  const [loading, setLoading] = useState(true);
  const [filters, setFilters] = useState<TicketFiltersValue>({});
  const [viewMode, setViewMode] = useState<"list" | "board">("list");
  const [createDialogOpen, setCreateDialogOpen] = useState(false);

  useEffect(() => {
    loadTickets();
  }, []);

  const loadTickets = async () => {
    try {
      const response = await ticketApi.list();
      setTickets((response.tickets || []) as Ticket[]);
    } catch (error) {
      console.error("Failed to load tickets:", error);
    } finally {
      setLoading(false);
    }
  };

  const handleStatusChange = useCallback(async (identifier: string, newStatus: Ticket["status"]) => {
    try {
      await ticketApi.updateStatus(identifier, newStatus);
      setTickets(prev =>
        prev.map(t => (t.identifier === identifier ? { ...t, status: newStatus } : t))
      );
    } catch (error) {
      console.error("Failed to update ticket status:", error);
    }
  }, []);

  const handleTicketClick = useCallback((ticket: Ticket) => {
    router.push(`tickets/${ticket.identifier}`);
  }, [router]);

  const handleTicketCreated = useCallback((ticketId: number, identifier: string) => {
    // Reload tickets to show the new one
    loadTickets();
    // Optionally navigate to the new ticket
    // router.push(`tickets/${identifier}`);
  }, []);

  const filteredTickets = useMemo(() => {
    return tickets.filter((ticket) => {
      // Search filter
      const searchTerm = filters.search?.toLowerCase() || "";
      const matchesSearch =
        !searchTerm ||
        ticket.title.toLowerCase().includes(searchTerm) ||
        ticket.identifier.toLowerCase().includes(searchTerm);

      // Status filter
      const matchesStatus = !filters.status || ticket.status === filters.status;

      // Type filter
      const matchesType = !filters.type || ticket.type === filters.type;

      // Priority filter
      const matchesPriority = !filters.priority || ticket.priority === filters.priority;

      return matchesSearch && matchesStatus && matchesType && matchesPriority;
    });
  }, [tickets, filters]);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    );
  }

  return (
    <div className="p-6">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-foreground">Tickets</h1>
          <p className="text-muted-foreground">
            Track and manage your tasks and issues
          </p>
        </div>
        <Button onClick={() => setCreateDialogOpen(true)}>
          <Plus className="w-4 h-4 mr-2" />
          New Ticket
        </Button>
      </div>

      {/* Create Ticket Dialog */}
      <TicketCreateDialog
        open={createDialogOpen}
        onOpenChange={setCreateDialogOpen}
        onCreated={handleTicketCreated}
      />

      {/* Filters */}
      <div className="mb-6">
        <TicketFilters
          value={filters}
          onChange={setFilters}
          viewMode={viewMode}
          onViewModeChange={setViewMode}
          showViewToggle
        />
      </div>

      {/* Content */}
      {viewMode === "list" ? (
        <ListView tickets={filteredTickets} />
      ) : (
        <div className="flex-1 min-h-0">
          <KanbanBoard
            tickets={filteredTickets}
            onStatusChange={handleStatusChange}
            onTicketClick={handleTicketClick}
          />
        </div>
      )}
    </div>
  );
}

function ListView({ tickets }: { tickets: Ticket[] }) {
  return (
    <div className="border border-border rounded-lg overflow-hidden">
      <table className="w-full">
        <thead className="bg-muted">
          <tr>
            <th className="px-4 py-3 text-left text-sm font-medium">ID</th>
            <th className="px-4 py-3 text-left text-sm font-medium">Title</th>
            <th className="px-4 py-3 text-left text-sm font-medium">Status</th>
            <th className="px-4 py-3 text-left text-sm font-medium">Priority</th>
            <th className="px-4 py-3 text-left text-sm font-medium">Type</th>
            <th className="px-4 py-3 text-left text-sm font-medium">Created</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-border">
          {tickets.map((ticket) => (
            <tr key={ticket.id} className="hover:bg-muted/50">
              <td className="px-4 py-3">
                <code className="text-sm text-primary">{ticket.identifier}</code>
              </td>
              <td className="px-4 py-3">
                <Link
                  href={`tickets/${ticket.identifier}`}
                  className="text-foreground hover:text-primary"
                >
                  {ticket.title}
                </Link>
              </td>
              <td className="px-4 py-3">
                <span
                  className={`px-2 py-1 text-xs rounded-full ${
                    statusColors[ticket.status] || "bg-gray-100"
                  }`}
                >
                  {ticket.status.replace("_", " ")}
                </span>
              </td>
              <td className="px-4 py-3">
                <span className={priorityColors[ticket.priority] || ""}>
                  {ticket.priority}
                </span>
              </td>
              <td className="px-4 py-3 text-muted-foreground">
                {ticket.type}
              </td>
              <td className="px-4 py-3 text-muted-foreground">
                {new Date(ticket.created_at).toLocaleDateString()}
              </td>
            </tr>
          ))}
          {tickets.length === 0 && (
            <tr>
              <td
                colSpan={6}
                className="px-4 py-8 text-center text-muted-foreground"
              >
                No tickets found.
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  );
}

