"use client";

import { useState, useEffect, useMemo, useCallback } from "react";
import { useTicketStore, type TicketStatus, type TicketType, type TicketPriority, type Ticket } from "@/stores/ticket";
import type { TicketFilterState, TicketFilterActions } from "./types";

/**
 * Custom hook for managing ticket filter state and actions
 */
export function useTicketFilters(): TicketFilterState & TicketFilterActions & {
  filteredTickets: Ticket[];
} {
  const { tickets, filters, setFilters } = useTicketStore();

  // Local state
  const [searchQuery, setSearchQuery] = useState(filters.search || "");
  const [selectedStatuses, setSelectedStatuses] = useState<TicketStatus[]>([]);
  const [selectedTypes, setSelectedTypes] = useState<TicketType[]>([]);
  const [selectedPriorities, setSelectedPriorities] = useState<TicketPriority[]>([]);

  // Debounce search
  useEffect(() => {
    const timer = setTimeout(() => {
      setFilters({ ...filters, search: searchQuery || undefined });
    }, 300);
    return () => clearTimeout(timer);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [searchQuery]);

  // Filter tickets locally
  const filteredTickets = useMemo(() => {
    return tickets.filter((ticket) => {
      // Search filter
      if (searchQuery) {
        const query = searchQuery.toLowerCase();
        const matchesTitle = ticket.title.toLowerCase().includes(query);
        const matchesId = ticket.identifier.toLowerCase().includes(query);
        if (!matchesTitle && !matchesId) return false;
      }

      // Status filter
      if (selectedStatuses.length > 0 && !selectedStatuses.includes(ticket.status)) {
        return false;
      }

      // Type filter
      if (selectedTypes.length > 0 && !selectedTypes.includes(ticket.type)) {
        return false;
      }

      // Priority filter
      if (selectedPriorities.length > 0 && !selectedPriorities.includes(ticket.priority)) {
        return false;
      }

      return true;
    });
  }, [tickets, searchQuery, selectedStatuses, selectedTypes, selectedPriorities]);

  // Toggle functions
  const toggleStatus = useCallback((status: TicketStatus) => {
    setSelectedStatuses(prev =>
      prev.includes(status)
        ? prev.filter(s => s !== status)
        : [...prev, status]
    );
  }, []);

  const toggleType = useCallback((type: TicketType) => {
    setSelectedTypes(prev =>
      prev.includes(type)
        ? prev.filter(t => t !== type)
        : [...prev, type]
    );
  }, []);

  const togglePriority = useCallback((priority: TicketPriority) => {
    setSelectedPriorities(prev =>
      prev.includes(priority)
        ? prev.filter(p => p !== priority)
        : [...prev, priority]
    );
  }, []);

  const clearAllFilters = useCallback(() => {
    setSearchQuery("");
    setSelectedStatuses([]);
    setSelectedTypes([]);
    setSelectedPriorities([]);
    setFilters({});
  }, [setFilters]);

  const hasActiveFilters = searchQuery.length > 0 ||
    selectedStatuses.length > 0 ||
    selectedTypes.length > 0 ||
    selectedPriorities.length > 0;

  return {
    // State
    searchQuery,
    selectedStatuses,
    selectedTypes,
    selectedPriorities,
    filteredTickets,

    // Actions
    setSearchQuery,
    toggleStatus,
    toggleType,
    togglePriority,
    clearAllFilters,
    hasActiveFilters,
  };
}
