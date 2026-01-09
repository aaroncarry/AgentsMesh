"use client";

import { useState, useEffect, useCallback } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { KanbanBoard } from "@/components/tickets";
import { ticketApi } from "@/lib/api/client";

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
  const [filter, setFilter] = useState("");
  const [statusFilter, setStatusFilter] = useState("");
  const [viewMode, setViewMode] = useState<"list" | "board">("list");

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

  const filteredTickets = tickets.filter((ticket) => {
    const matchesSearch =
      ticket.title.toLowerCase().includes(filter.toLowerCase()) ||
      ticket.identifier.toLowerCase().includes(filter.toLowerCase());
    const matchesStatus = !statusFilter || ticket.status === statusFilter;
    return matchesSearch && matchesStatus;
  });

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
        <Button>
          <svg
            className="w-4 h-4 mr-2"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M12 4v16m8-8H4"
            />
          </svg>
          New Ticket
        </Button>
      </div>

      {/* Filters */}
      <div className="flex items-center gap-4 mb-6">
        <div className="flex-1 max-w-sm">
          <Input
            placeholder="Search tickets..."
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
            className="w-full"
          />
        </div>
        <select
          className="px-3 py-2 border border-border rounded-md bg-background text-sm"
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
        >
          <option value="">All Status</option>
          <option value="backlog">Backlog</option>
          <option value="todo">To Do</option>
          <option value="in_progress">In Progress</option>
          <option value="in_review">In Review</option>
          <option value="done">Done</option>
        </select>
        <div className="flex border border-border rounded-md overflow-hidden">
          <button
            className={`px-3 py-2 text-sm ${
              viewMode === "list" ? "bg-muted" : ""
            }`}
            onClick={() => setViewMode("list")}
          >
            <svg
              className="w-4 h-4"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M4 6h16M4 10h16M4 14h16M4 18h16"
              />
            </svg>
          </button>
          <button
            className={`px-3 py-2 text-sm ${
              viewMode === "board" ? "bg-muted" : ""
            }`}
            onClick={() => setViewMode("board")}
          >
            <svg
              className="w-4 h-4"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M9 17V7m0 10a2 2 0 01-2 2H5a2 2 0 01-2-2V7a2 2 0 012-2h2a2 2 0 012 2m0 10a2 2 0 002 2h2a2 2 0 002-2M9 7a2 2 0 012-2h2a2 2 0 012 2m0 10V7m0 10a2 2 0 002 2h2a2 2 0 002-2V7a2 2 0 00-2-2h-2a2 2 0 00-2 2"
              />
            </svg>
          </button>
        </div>
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

