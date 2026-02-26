"use client";

import { Ticket, TicketStatus } from "@/stores/ticket";
import { KanbanBoard } from "@/components/tickets";

interface BoardViewLayoutProps {
  tickets: Ticket[];
  onStatusChange: (slug: string, newStatus: TicketStatus) => Promise<void>;
  onTicketClick: (ticket: Ticket) => void;
  onCreatePodRequest?: (ticket: Ticket) => void;
}

/**
 * Board view - full height kanban board.
 * Clicking a ticket navigates directly to the detail page.
 */
export function BoardViewLayout({
  tickets,
  onStatusChange,
  onTicketClick,
  onCreatePodRequest,
}: BoardViewLayoutProps) {
  return (
    <div className="h-full flex flex-col">
      <div className="flex-1 min-h-0 p-4">
        <KanbanBoard
          tickets={tickets}
          onStatusChange={onStatusChange}
          onTicketClick={onTicketClick}
          onCreatePodRequest={onCreatePodRequest}
        />
      </div>
    </div>
  );
}
