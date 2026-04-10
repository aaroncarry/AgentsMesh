import { ticketApi } from "@/lib/api";
import { getErrorMessage } from "@/lib/utils";
import type { Ticket, ColumnPagination, TicketFilters, TicketStoreDeps } from "./ticket";
import { flattenColumns } from "./ticket-mutations";

/** Build initial ColumnPagination from board API response. */
function initPagination(columns: { tickets: unknown[]; count: number; status: string }[]) {
  const pag: Record<string, ColumnPagination> = {};
  for (const col of columns) {
    pag[col.status] = { offset: col.tickets.length, hasMore: col.tickets.length < col.count, loading: false };
  }
  return pag;
}

export function createBoardActions(set: TicketStoreDeps["set"], get: TicketStoreDeps["get"]) {
  return {
    /** Fetch board data. boardColumns becomes the ticket owner; tickets is derived. */
    fetchBoard: async (filters?: TicketFilters) => {
      const mergedFilters = { ...get().filters, ...filters };
      set({ error: null, filters: mergedFilters });
      try {
        const response = await ticketApi.getBoard(mergedFilters);
        const board = response.board;
        set({
          boardColumns: board.columns,
          tickets: flattenColumns(board.columns),
          totalCount: board.columns.reduce((sum, col) => sum + col.count, 0),
          priorityCounts: board.priority_counts || {},
          columnPagination: initPagination(board.columns),
        });
      } catch (error: unknown) {
        set({ error: getErrorMessage(error, "Failed to fetch board") });
      }
    },

    /** Load next page of tickets for a single column. Appends to boardColumns. */
    loadMoreColumn: async (status: string) => {
      const { columnPagination, boardColumns, filters } = get();
      const pag = columnPagination[status];
      if (!pag || !pag.hasMore || pag.loading) return;

      set({ columnPagination: { ...columnPagination, [status]: { ...pag, loading: true } } });
      try {
        const response = await ticketApi.list({ ...filters, status, offset: pag.offset, limit: 20 });
        const newTickets = (response.tickets || []) as Ticket[];
        const newOffset = pag.offset + newTickets.length;

        const updatedColumns = boardColumns.map((col) =>
          col.status === status ? { ...col, tickets: [...col.tickets, ...newTickets] } : col,
        );
        set({
          boardColumns: updatedColumns,
          tickets: flattenColumns(updatedColumns),
          columnPagination: {
            ...get().columnPagination,
            [status]: { offset: newOffset, hasMore: newOffset < (response.total || 0), loading: false },
          },
        });
      } catch (error: unknown) {
        set({
          columnPagination: { ...get().columnPagination, [status]: { ...pag, loading: false } },
          error: getErrorMessage(error, "Failed to load more tickets"),
        });
      }
    },
  };
}
