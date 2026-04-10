import type { BoardColumn } from "@/lib/api";
import type { Ticket, TicketStoreDeps } from "./ticket";

// ── Helpers ────────────────────────────────────────────────────────────

/** Derive flat tickets array from boardColumns. */
export function flattenColumns(columns: BoardColumn[]): Ticket[] {
  return columns.flatMap((col) => col.tickets) as Ticket[];
}

// ── Strategy Interface ─────────────────────────────────────────────────

/** Abstraction for ticket collection mutations. Board and list each implement this. */
export interface TicketMutations {
  insert(ticket: Ticket, status: string): void;
  update(slug: string, newTicket: Ticket): void;
  remove(slug: string): void;
  moveStatus(slug: string, movedTicket: Ticket, from: string, to: string): void;
}

// ── List Strategy ──────────────────────────────────────────────────────

export function createListMutations(
  get: TicketStoreDeps["get"], set: TicketStoreDeps["set"],
): TicketMutations {
  return {
    insert(ticket) {
      set((s: { tickets: Ticket[]; totalCount: number }) => ({
        tickets: [ticket, ...s.tickets], totalCount: s.totalCount + 1,
      }));
    },
    update(slug, newTicket) {
      set((s: { tickets: Ticket[] }) => ({
        tickets: s.tickets.map((t) => (t.slug === slug ? newTicket : t)),
      }));
    },
    remove(slug) {
      set((s: { tickets: Ticket[]; totalCount: number }) => ({
        tickets: s.tickets.filter((t) => t.slug !== slug),
        totalCount: s.totalCount - 1,
        currentTicket: get().currentTicket?.slug === slug ? null : get().currentTicket,
      }));
    },
    moveStatus(slug, movedTicket) {
      set((s: { tickets: Ticket[] }) => ({
        tickets: s.tickets.map((t) => (t.slug === slug ? movedTicket : t)),
      }));
    },
  };
}

// ── Board Strategy ─────────────────────────────────────────────────────

export function createBoardMutations(
  get: TicketStoreDeps["get"], set: TicketStoreDeps["set"],
): TicketMutations {
  /** Apply a column-level transform, then derive tickets. */
  const applyColumns = (mapper: (col: BoardColumn) => BoardColumn) => {
    const updated = get().boardColumns.map(mapper);
    set({ boardColumns: updated, tickets: flattenColumns(updated) });
  };

  return {
    insert(ticket, status) {
      applyColumns((col) =>
        col.status === status ? { ...col, count: col.count + 1, tickets: [ticket, ...col.tickets] } : col,
      );
      set({ totalCount: get().totalCount + 1 });
    },

    update(slug, newTicket) {
      const oldTicket = get().tickets.find((t) => t.slug === slug);
      if (oldTicket && oldTicket.status !== newTicket.status) {
        // Status changed via field edit — cross-column move
        this.moveStatus(slug, newTicket, oldTicket.status, newTicket.status);
      } else {
        applyColumns((col) => ({
          ...col, tickets: col.tickets.map((t) => (t.slug === slug ? newTicket : t)),
        }));
      }
    },

    remove(slug) {
      applyColumns((col) => {
        if (!col.tickets.some((t) => t.slug === slug)) return col;
        return { ...col, count: Math.max(0, col.count - 1), tickets: col.tickets.filter((t) => t.slug !== slug) };
      });
      set({
        totalCount: Math.max(0, get().totalCount - 1),
        currentTicket: get().currentTicket?.slug === slug ? null : get().currentTicket,
      });
    },

    moveStatus(slug, movedTicket, from, to) {
      applyColumns((col) => {
        if (col.status === from) {
          return { ...col, count: Math.max(0, col.count - 1), tickets: col.tickets.filter((t) => t.slug !== slug) };
        }
        if (col.status === to) {
          return { ...col, count: col.count + 1, tickets: [movedTicket, ...col.tickets] };
        }
        return col;
      });
    },
  };
}
