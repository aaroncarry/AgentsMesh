import { describe, it, expect, beforeEach, vi } from 'vitest'
import { useTicketStore } from '../ticket'

// Mock the API client
vi.mock('@/lib/api', () => ({
  ticketApi: {
    list: vi.fn(),
    get: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    delete: vi.fn(),
    updateStatus: vi.fn(),
    getBoard: vi.fn(),
    listLabels: vi.fn(),
    createLabel: vi.fn(),
    deleteLabel: vi.fn(),
  },
}))

import { ticketApi } from '@/lib/api'

import type { TicketStatus, TicketPriority } from '@/lib/api'

const mkTicket = (slug: string, status: TicketStatus, priority: TicketPriority = 'medium') => ({
  id: parseInt(slug.replace(/\D/g, '')) || 1,
  number: 1, slug, title: 'T', status, priority,
  created_at: '', updated_at: '',
})

const mkBoardResponse = () => ({
  board: {
    columns: [
      { status: 'backlog' as const, count: 10, tickets: [mkTicket('B-1', 'backlog'), mkTicket('B-2', 'backlog')] },
      { status: 'todo' as const, count: 0, tickets: [] as ReturnType<typeof mkTicket>[] },
      { status: 'in_progress' as const, count: 3, tickets: [mkTicket('IP-1', 'in_progress')] },
      { status: 'in_review' as const, count: 1, tickets: [mkTicket('IR-1', 'in_review')] },
      { status: 'done' as const, count: 50, tickets: [mkTicket('D-1', 'done')] },
    ],
    priority_counts: { high: 5, medium: 8, low: 2 } as Record<string, number>,
  },
})

beforeEach(() => {
  useTicketStore.setState({
    tickets: [], currentTicket: null, labels: [], filters: {},
    loading: false, error: null, totalCount: 0,
    boardColumns: [], priorityCounts: {}, columnPagination: {}, doneCollapsed: true,
  })
  vi.clearAllMocks()
})

describe('fetchBoard', () => {
  it('populates boardColumns, tickets, pagination, and priorityCounts', async () => {
    vi.mocked(ticketApi.getBoard).mockResolvedValue(mkBoardResponse())

    await useTicketStore.getState().fetchBoard()

    const s = useTicketStore.getState()
    expect(s.boardColumns).toHaveLength(5)
    expect(s.tickets).toHaveLength(5) // 2+0+1+1+1
    expect(s.totalCount).toBe(64) // 10+0+3+1+50
    expect(s.priorityCounts).toEqual({ high: 5, medium: 8, low: 2 })
    // Pagination initialized
    expect(s.columnPagination['backlog']).toEqual({ offset: 2, hasMore: true, loading: false })
    expect(s.columnPagination['todo']).toEqual({ offset: 0, hasMore: false, loading: false })
    expect(s.columnPagination['done']).toEqual({ offset: 1, hasMore: true, loading: false })
  })

  it('passes merged filters to API', async () => {
    vi.mocked(ticketApi.getBoard).mockResolvedValue(mkBoardResponse())
    useTicketStore.setState({ filters: { search: 'bug' } })

    await useTicketStore.getState().fetchBoard({ priority: 'high' })

    expect(ticketApi.getBoard).toHaveBeenCalledWith({ search: 'bug', priority: 'high' })
  })
})

describe('loadMoreColumn', () => {
  it('appends tickets to boardColumns and updates pagination', async () => {
    vi.mocked(ticketApi.getBoard).mockResolvedValue(mkBoardResponse())
    await useTicketStore.getState().fetchBoard()

    const newTickets = [mkTicket('B-3', 'backlog'), mkTicket('B-4', 'backlog')]
    vi.mocked(ticketApi.list).mockResolvedValue({ tickets: newTickets, total: 10 })

    await useTicketStore.getState().loadMoreColumn('backlog')

    const s = useTicketStore.getState()
    const backlog = s.boardColumns.find(c => c.status === 'backlog')!
    expect(backlog.tickets).toHaveLength(4) // 2 initial + 2 loaded
    expect(s.tickets).toHaveLength(7) // 5 initial + 2 loaded
    expect(s.columnPagination['backlog']).toEqual({ offset: 4, hasMore: true, loading: false })
  })

  it('passes current filters to list API', async () => {
    vi.mocked(ticketApi.getBoard).mockResolvedValue(mkBoardResponse())
    useTicketStore.setState({ filters: { priority: 'high' } })
    await useTicketStore.getState().fetchBoard()

    vi.mocked(ticketApi.list).mockResolvedValue({ tickets: [], total: 10 })
    await useTicketStore.getState().loadMoreColumn('backlog')

    expect(ticketApi.list).toHaveBeenCalledWith(
      expect.objectContaining({ priority: 'high', status: 'backlog', offset: 2, limit: 20 })
    )
  })

  it('does not fire when hasMore is false', async () => {
    vi.mocked(ticketApi.getBoard).mockResolvedValue(mkBoardResponse())
    await useTicketStore.getState().fetchBoard()

    await useTicketStore.getState().loadMoreColumn('todo') // count=0, hasMore=false

    expect(ticketApi.list).not.toHaveBeenCalled()
  })

  it('does not double-fire when already loading', async () => {
    vi.mocked(ticketApi.getBoard).mockResolvedValue(mkBoardResponse())
    await useTicketStore.getState().fetchBoard()

    vi.mocked(ticketApi.list).mockImplementation(() => new Promise(() => {})) // never resolves
    useTicketStore.getState().loadMoreColumn('backlog')
    await new Promise(r => setTimeout(r, 10)) // let loading state set

    await useTicketStore.getState().loadMoreColumn('backlog') // should bail

    expect(ticketApi.list).toHaveBeenCalledTimes(1)
  })
})

describe('fetchTickets clears board state', () => {
  it('resets boardColumns, priorityCounts, columnPagination', async () => {
    vi.mocked(ticketApi.getBoard).mockResolvedValue(mkBoardResponse())
    await useTicketStore.getState().fetchBoard()

    expect(useTicketStore.getState().boardColumns).toHaveLength(5)

    vi.mocked(ticketApi.list).mockResolvedValue({ tickets: [mkTicket('L-1', 'todo')], total: 1 })
    await useTicketStore.getState().fetchTickets()

    const s = useTicketStore.getState()
    expect(s.boardColumns).toEqual([])
    expect(s.priorityCounts).toEqual({})
    expect(s.columnPagination).toEqual({})
    expect(s.tickets).toHaveLength(1)
  })
})

// ── Error handling ────────────────────────────────────────────────────

describe('fetchBoard error', () => {
  it('sets error state on API failure', async () => {
    vi.mocked(ticketApi.getBoard).mockRejectedValue(new Error('Network'))

    await useTicketStore.getState().fetchBoard()

    expect(useTicketStore.getState().error).toBe('Network')
    expect(useTicketStore.getState().boardColumns).toEqual([])
  })
})

describe('loadMoreColumn error', () => {
  it('resets loading state and sets error on failure', async () => {
    vi.mocked(ticketApi.getBoard).mockResolvedValue(mkBoardResponse())
    await useTicketStore.getState().fetchBoard()

    vi.mocked(ticketApi.list).mockRejectedValue(new Error('Timeout'))
    await useTicketStore.getState().loadMoreColumn('backlog')

    const s = useTicketStore.getState()
    expect(s.columnPagination['backlog'].loading).toBe(false)
    expect(s.columnPagination['backlog'].hasMore).toBe(true) // unchanged
    expect(s.error).toBe('Timeout')
  })
})

// ── Board-mode CRUD via store ─────────────────────────────────────────

describe('board-mode createTicket', () => {
  it('inserts into boardColumns owner, derives tickets', async () => {
    vi.mocked(ticketApi.getBoard).mockResolvedValue(mkBoardResponse())
    await useTicketStore.getState().fetchBoard()

    const created = mkTicket('NEW-1', 'backlog')
    vi.mocked(ticketApi.create).mockResolvedValue(created)

    await useTicketStore.getState().createTicket({ repositoryId: 1, title: 'New' })

    const s = useTicketStore.getState()
    const backlog = s.boardColumns.find(c => c.status === 'backlog')!
    expect(backlog.tickets).toHaveLength(3) // 2 + 1
    expect(backlog.count).toBe(11)
    expect(s.tickets.find(t => t.slug === 'NEW-1')).toBeTruthy()
    expect(s.totalCount).toBe(65)
  })
})

describe('board-mode updateTicket', () => {
  it('updates ticket in column without status change', async () => {
    vi.mocked(ticketApi.getBoard).mockResolvedValue(mkBoardResponse())
    await useTicketStore.getState().fetchBoard()

    const updated = { ...mkTicket('B-1', 'backlog'), title: 'Renamed' }
    vi.mocked(ticketApi.update).mockResolvedValue(updated)

    await useTicketStore.getState().updateTicket('B-1', { title: 'Renamed' })

    const s = useTicketStore.getState()
    expect(s.tickets.find(t => t.slug === 'B-1')!.title).toBe('Renamed')
    expect(s.boardColumns.find(c => c.status === 'backlog')!.tickets[0].title).toBe('Renamed')
  })

  it('moves ticket across columns on status change', async () => {
    vi.mocked(ticketApi.getBoard).mockResolvedValue(mkBoardResponse())
    await useTicketStore.getState().fetchBoard()

    const updated = { ...mkTicket('B-1', 'done'), title: 'Done now' }
    vi.mocked(ticketApi.update).mockResolvedValue(updated)

    await useTicketStore.getState().updateTicket('B-1', { status: 'done' })

    const s = useTicketStore.getState()
    const backlog = s.boardColumns.find(c => c.status === 'backlog')!
    const done = s.boardColumns.find(c => c.status === 'done')!
    expect(backlog.tickets).toHaveLength(1) // was 2, moved 1
    expect(backlog.count).toBe(9)
    expect(done.tickets).toHaveLength(2) // was 1, added 1
    expect(done.count).toBe(51)
  })
})

describe('board-mode deleteTicket', () => {
  it('removes from boardColumns owner, derives tickets', async () => {
    vi.mocked(ticketApi.getBoard).mockResolvedValue(mkBoardResponse())
    await useTicketStore.getState().fetchBoard()

    vi.mocked(ticketApi.delete).mockResolvedValue({ message: 'ok' })
    await useTicketStore.getState().deleteTicket('B-1')

    const s = useTicketStore.getState()
    const backlog = s.boardColumns.find(c => c.status === 'backlog')!
    expect(backlog.tickets).toHaveLength(1)
    expect(backlog.count).toBe(9)
    expect(s.tickets.find(t => t.slug === 'B-1')).toBeUndefined()
    expect(s.totalCount).toBe(63)
  })
})

// ── Optimistic status update + rollback ───────────────────────────────

describe('board-mode updateTicketStatus', () => {
  it('optimistically moves ticket between columns', async () => {
    vi.mocked(ticketApi.getBoard).mockResolvedValue(mkBoardResponse())
    await useTicketStore.getState().fetchBoard()
    vi.mocked(ticketApi.updateStatus).mockResolvedValue({ message: 'ok' })

    await useTicketStore.getState().updateTicketStatus('B-1', 'in_progress')

    const s = useTicketStore.getState()
    const backlog = s.boardColumns.find(c => c.status === 'backlog')!
    const inProgress = s.boardColumns.find(c => c.status === 'in_progress')!
    expect(backlog.tickets).toHaveLength(1)
    expect(backlog.count).toBe(9)
    expect(inProgress.tickets).toHaveLength(2)
    expect(inProgress.count).toBe(4)
    expect(s.tickets.find(t => t.slug === 'B-1')!.status).toBe('in_progress')
  })

  it('rolls back on API failure', async () => {
    vi.mocked(ticketApi.getBoard).mockResolvedValue(mkBoardResponse())
    await useTicketStore.getState().fetchBoard()

    const prevBacklogCount = useTicketStore.getState().boardColumns.find(c => c.status === 'backlog')!.count
    const prevTickets = useTicketStore.getState().tickets

    vi.mocked(ticketApi.updateStatus).mockRejectedValue(new Error('Server error'))

    await expect(useTicketStore.getState().updateTicketStatus('B-1', 'done')).rejects.toThrow()

    const s = useTicketStore.getState()
    // Rolled back to original state
    expect(s.boardColumns.find(c => c.status === 'backlog')!.count).toBe(prevBacklogCount)
    expect(s.tickets).toEqual(prevTickets)
    expect(s.error).toBe('Server error')
  })

  it('skips mutation when fromStatus equals target status', async () => {
    vi.mocked(ticketApi.getBoard).mockResolvedValue(mkBoardResponse())
    await useTicketStore.getState().fetchBoard()
    vi.mocked(ticketApi.updateStatus).mockResolvedValue({ message: 'ok' })

    const prevColumns = useTicketStore.getState().boardColumns

    await useTicketStore.getState().updateTicketStatus('B-1', 'backlog') // same status

    // Columns unchanged (no-op mutation)
    expect(useTicketStore.getState().boardColumns).toEqual(prevColumns)
    // API still called (server should decide)
    expect(ticketApi.updateStatus).toHaveBeenCalledWith('B-1', 'backlog')
  })
})

describe('list-mode updateTicketStatus', () => {
  it('updates ticket in flat array when no board data', async () => {
    const ticket = mkTicket('L-1', 'todo')
    useTicketStore.setState({ tickets: [ticket], boardColumns: [] })
    vi.mocked(ticketApi.updateStatus).mockResolvedValue({ message: 'ok' })

    await useTicketStore.getState().updateTicketStatus('L-1', 'done')

    expect(useTicketStore.getState().tickets[0].status).toBe('done')
  })

  it('rolls back flat array on API failure', async () => {
    const ticket = mkTicket('L-1', 'todo')
    useTicketStore.setState({ tickets: [ticket], boardColumns: [] })
    vi.mocked(ticketApi.updateStatus).mockRejectedValue(new Error('fail'))

    await expect(useTicketStore.getState().updateTicketStatus('L-1', 'done')).rejects.toThrow()

    expect(useTicketStore.getState().tickets[0].status).toBe('todo') // rolled back
  })
})
