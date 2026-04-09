import { describe, it, expect, beforeEach } from 'vitest'
import {
  createListMutations, createBoardMutations, flattenColumns, type TicketMutations,
} from '../ticket-mutations'
import type { Ticket } from '../ticket'
import type { TicketStoreDeps } from '../ticket'

// ── Test helpers ────────────────────────────────────────────────────────

const mkTicket = (slug: string, status: string, title = 'T'): Ticket => ({
  id: parseInt(slug.replace(/\D/g, '')) || 1,
  number: 1, slug, title, status, priority: 'medium',
  created_at: '', updated_at: '',
} as Ticket)

const mkColumn = (status: string, tickets: Ticket[], count?: number) => ({
  status, tickets, count: count ?? tickets.length,
})

function createMockDeps(initial: Partial<ReturnType<TicketStoreDeps["get"]>> = {}) {
  let state: ReturnType<TicketStoreDeps["get"]> = {
    tickets: [], currentTicket: null, filters: {},
    totalCount: 0, boardColumns: [], priorityCounts: {},
    columnPagination: {},
    ...initial,
  }
  const get = () => state
  const set = (updater: object | ((s: typeof state) => object)) => {
    const patch = typeof updater === 'function' ? updater(state) : updater
    state = { ...state, ...patch }
  }
  return { get, set, state: () => state }
}

// ── flattenColumns ──────────────────────────────────────────────────────

describe('flattenColumns', () => {
  it('flattens columns into a single array', () => {
    const cols = [
      mkColumn('backlog', [mkTicket('A-1', 'backlog')]),
      mkColumn('done', [mkTicket('A-2', 'done'), mkTicket('A-3', 'done')]),
    ]
    expect(flattenColumns(cols)).toHaveLength(3)
  })

  it('returns empty for empty columns', () => {
    expect(flattenColumns([])).toEqual([])
  })
})

// ── List Strategy ───────────────────────────────────────────────────────

describe('createListMutations', () => {
  let deps: ReturnType<typeof createMockDeps>
  let mut: TicketMutations

  beforeEach(() => {
    deps = createMockDeps({
      tickets: [mkTicket('T-1', 'backlog'), mkTicket('T-2', 'todo')],
      totalCount: 2,
    })
    mut = createListMutations(deps.get, deps.set)
  })

  it('insert prepends ticket and increments totalCount', () => {
    mut.insert(mkTicket('T-3', 'backlog'), 'backlog')
    expect(deps.state().tickets).toHaveLength(3)
    expect(deps.state().tickets[0].slug).toBe('T-3')
    expect(deps.state().totalCount).toBe(3)
  })

  it('update replaces ticket by slug', () => {
    mut.update('T-1', mkTicket('T-1', 'backlog', 'Updated'))
    expect(deps.state().tickets.find(t => t.slug === 'T-1')!.title).toBe('Updated')
  })

  it('remove filters ticket and decrements totalCount', () => {
    mut.remove('T-1')
    expect(deps.state().tickets).toHaveLength(1)
    expect(deps.state().totalCount).toBe(1)
  })

  it('remove clears currentTicket if matching', () => {
    deps = createMockDeps({
      tickets: [mkTicket('T-1', 'backlog')],
      totalCount: 1,
      currentTicket: mkTicket('T-1', 'backlog'),
    })
    mut = createListMutations(deps.get, deps.set)
    mut.remove('T-1')
    expect(deps.state().currentTicket).toBeNull()
  })

  it('moveStatus replaces ticket in place', () => {
    mut.moveStatus('T-1', mkTicket('T-1', 'done'), 'backlog', 'done')
    expect(deps.state().tickets.find(t => t.slug === 'T-1')!.status).toBe('done')
  })
})

// ── Board Strategy ──────────────────────────────────────────────────────

describe('createBoardMutations', () => {
  let deps: ReturnType<typeof createMockDeps>
  let mut: TicketMutations

  beforeEach(() => {
    const t1 = mkTicket('T-1', 'backlog')
    const t2 = mkTicket('T-2', 'in_progress')
    deps = createMockDeps({
      boardColumns: [
        mkColumn('backlog', [t1], 10),
        mkColumn('todo', [], 0),
        mkColumn('in_progress', [t2], 5),
        mkColumn('in_review', [], 0),
        mkColumn('done', [], 20),
      ],
      tickets: [t1, t2],
      totalCount: 35,
    })
    mut = createBoardMutations(deps.get, deps.set)
  })

  it('insert adds to correct column, increments count, derives tickets', () => {
    mut.insert(mkTicket('T-3', 'backlog'), 'backlog')
    const col = deps.state().boardColumns.find(c => c.status === 'backlog')!
    expect(col.tickets).toHaveLength(2)
    expect(col.count).toBe(11)
    expect(deps.state().tickets).toHaveLength(3)
    expect(deps.state().totalCount).toBe(36)
  })

  it('update replaces ticket in column and derives tickets', () => {
    mut.update('T-1', mkTicket('T-1', 'backlog', 'New Title'))
    const col = deps.state().boardColumns.find(c => c.status === 'backlog')!
    expect((col.tickets[0] as Ticket).title).toBe('New Title')
    expect(deps.state().tickets.find(t => t.slug === 'T-1')!.title).toBe('New Title')
  })

  it('update with status change moves ticket across columns', () => {
    // T-1 is in backlog, update changes status to in_progress
    mut.update('T-1', mkTicket('T-1', 'in_progress', 'Moved'))
    const backlog = deps.state().boardColumns.find(c => c.status === 'backlog')!
    const inProgress = deps.state().boardColumns.find(c => c.status === 'in_progress')!

    expect(backlog.tickets).toHaveLength(0)
    expect(backlog.count).toBe(9)
    expect(inProgress.tickets).toHaveLength(2)
    expect(inProgress.count).toBe(6)
  })

  it('remove removes from correct column, decrements count', () => {
    mut.remove('T-1')
    const col = deps.state().boardColumns.find(c => c.status === 'backlog')!
    expect(col.tickets).toHaveLength(0)
    expect(col.count).toBe(9)
    expect(deps.state().totalCount).toBe(34)
    expect(deps.state().tickets).toHaveLength(1)
  })

  it('moveStatus moves ticket between columns with correct counts', () => {
    mut.moveStatus('T-1', mkTicket('T-1', 'done'), 'backlog', 'done')
    const backlog = deps.state().boardColumns.find(c => c.status === 'backlog')!
    const done = deps.state().boardColumns.find(c => c.status === 'done')!

    expect(backlog.tickets).toHaveLength(0)
    expect(backlog.count).toBe(9)
    expect(done.tickets).toHaveLength(1)
    expect(done.count).toBe(21)
    // tickets array re-derived
    expect(deps.state().tickets.find(t => t.slug === 'T-1')!.status).toBe('done')
  })
})
