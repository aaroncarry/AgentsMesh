import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { SessionDetail } from '../SessionDetail'
import { sessionApi } from '@/lib/api/client'

// Mock next/navigation
const mockRouterBack = vi.fn()
const mockRouterPush = vi.fn()
vi.mock('next/navigation', () => ({
  useRouter: () => ({
    back: mockRouterBack,
    push: mockRouterPush,
  }),
}))

// Mock session API
vi.mock('@/lib/api/client', () => ({
  sessionApi: {
    get: vi.fn(),
    terminate: vi.fn(),
    sendPrompt: vi.fn(),
  },
}))

// Mock Terminal component
vi.mock('../Terminal', () => ({
  Terminal: ({ sessionKey, className }: { sessionKey: string; className?: string }) => (
    <div data-testid="terminal" data-session-key={sessionKey} className={className}>
      Mock Terminal
    </div>
  ),
  useTerminal: () => ({
    terminalRef: { current: null },
    connect: vi.fn(),
    disconnect: vi.fn(),
    sendInput: vi.fn(),
    sendResize: vi.fn(),
  }),
}))

describe('SessionDetail Component', () => {
  const mockSession = {
    id: 1,
    session_key: 'test-session-key-123',
    status: 'running' as const,
    agent_status: 'idle',
    initial_prompt: 'Hello, help me with coding',
    branch_name: 'feature/test',
    worktree_path: '/tmp/worktree',
    started_at: '2024-01-15T10:00:00Z',
    finished_at: undefined,
    last_activity: '2024-01-15T11:00:00Z',
    created_at: '2024-01-15T09:00:00Z',
    runner: {
      id: 1,
      node_id: 'runner-001',
      status: 'online',
    },
    agent_type: {
      id: 1,
      name: 'Claude Code',
      slug: 'claude-code',
    },
    repository: {
      id: 1,
      name: 'my-project',
      full_path: 'org/my-project',
    },
    ticket: {
      id: 1,
      identifier: 'PROJ-42',
      title: 'Implement new feature',
    },
    created_by: {
      id: 1,
      username: 'john',
      name: 'John Doe',
    },
  }

  beforeEach(() => {
    vi.clearAllMocks()
    ;(sessionApi.get as ReturnType<typeof vi.fn>).mockResolvedValue({ session: mockSession })
  })

  describe('rendering', () => {
    it('should render session key', async () => {
      render(<SessionDetail sessionKey="test-session-key-123" />)
      await waitFor(() => {
        expect(screen.getByText('test-session-key-123')).toBeInTheDocument()
      })
    })

    it('should render status badge', async () => {
      render(<SessionDetail sessionKey="test-session-key-123" />)
      await waitFor(() => {
        expect(screen.getByText('Running')).toBeInTheDocument()
      })
    })

    it('should render agent type', async () => {
      render(<SessionDetail sessionKey="test-session-key-123" />)
      await waitFor(() => {
        expect(screen.getByText('Claude Code')).toBeInTheDocument()
      })
    })

    it('should render runner info', async () => {
      render(<SessionDetail sessionKey="test-session-key-123" />)
      await waitFor(() => {
        expect(screen.getByText('runner-001')).toBeInTheDocument()
      })
    })

    it('should render terminal', async () => {
      render(<SessionDetail sessionKey="test-session-key-123" />)
      await waitFor(() => {
        expect(screen.getByTestId('terminal')).toBeInTheDocument()
      })
    })

    it('should call sessionApi.get on mount', async () => {
      render(<SessionDetail sessionKey="test-session-key-123" />)
      await waitFor(() => {
        expect(sessionApi.get).toHaveBeenCalledWith('test-session-key-123')
      })
    })
  })

  describe('loading state', () => {
    it('should render skeleton when loading', () => {
      ;(sessionApi.get as ReturnType<typeof vi.fn>).mockImplementation(
        () => new Promise(() => {}) // Never resolves
      )
      render(<SessionDetail sessionKey="test-session-key-123" />)
      expect(screen.getByTestId('session-detail-skeleton')).toBeInTheDocument()
    })
  })

  describe('error state', () => {
    it('should render error message', async () => {
      ;(sessionApi.get as ReturnType<typeof vi.fn>).mockRejectedValue(new Error('Failed'))
      render(<SessionDetail sessionKey="test-session-key-123" />)
      await waitFor(() => {
        expect(screen.getByText('Failed to load session')).toBeInTheDocument()
      })
    })

    it('should render retry button on error', async () => {
      ;(sessionApi.get as ReturnType<typeof vi.fn>).mockRejectedValue(new Error('Failed'))
      render(<SessionDetail sessionKey="test-session-key-123" />)
      await waitFor(() => {
        expect(screen.getByText('Retry')).toBeInTheDocument()
      })
    })
  })

  describe('not found state', () => {
    it('should render not found message when session is null', async () => {
      ;(sessionApi.get as ReturnType<typeof vi.fn>).mockResolvedValue({ session: null })
      render(<SessionDetail sessionKey="test-session-key-123" />)
      await waitFor(() => {
        expect(screen.getByText('Session not found')).toBeInTheDocument()
      })
    })
  })

  describe('repository info', () => {
    it('should display repository path', async () => {
      render(<SessionDetail sessionKey="test-session-key-123" />)
      await waitFor(() => {
        expect(screen.getByText('org/my-project')).toBeInTheDocument()
      })
    })

    it('should display branch name', async () => {
      render(<SessionDetail sessionKey="test-session-key-123" />)
      await waitFor(() => {
        expect(screen.getByText('feature/test')).toBeInTheDocument()
      })
    })
  })

  describe('linked ticket', () => {
    it('should display ticket identifier', async () => {
      render(<SessionDetail sessionKey="test-session-key-123" />)
      await waitFor(() => {
        expect(screen.getByText('PROJ-42')).toBeInTheDocument()
      })
    })

    it('should display ticket title', async () => {
      render(<SessionDetail sessionKey="test-session-key-123" />)
      await waitFor(() => {
        expect(screen.getByText('Implement new feature')).toBeInTheDocument()
      })
    })

    it('should navigate to ticket on click', async () => {
      render(<SessionDetail sessionKey="test-session-key-123" />)
      await waitFor(() => {
        const ticketCard = screen.getByText('Implement new feature').closest('div[class*="cursor-pointer"]')
        expect(ticketCard).toBeInTheDocument()
        fireEvent.click(ticketCard!)
        expect(mockRouterPush).toHaveBeenCalledWith('/tickets/PROJ-42')
      })
    })
  })

  describe('initial prompt', () => {
    it('should display initial prompt', async () => {
      render(<SessionDetail sessionKey="test-session-key-123" />)
      await waitFor(() => {
        expect(screen.getByText('Hello, help me with coding')).toBeInTheDocument()
      })
    })
  })

  describe('created by', () => {
    it('should display creator name', async () => {
      render(<SessionDetail sessionKey="test-session-key-123" />)
      await waitFor(() => {
        expect(screen.getByText('John Doe')).toBeInTheDocument()
      })
    })

    it('should display creator username', async () => {
      render(<SessionDetail sessionKey="test-session-key-123" />)
      await waitFor(() => {
        expect(screen.getByText('@john')).toBeInTheDocument()
      })
    })
  })

  describe('navigation', () => {
    it('should navigate back when back button is clicked', async () => {
      render(<SessionDetail sessionKey="test-session-key-123" />)
      await waitFor(() => {
        const backButton = screen.getByText('Back')
        fireEvent.click(backButton)
        expect(mockRouterBack).toHaveBeenCalled()
      })
    })
  })

  describe('terminate action', () => {
    it('should show terminate button for running session', async () => {
      render(<SessionDetail sessionKey="test-session-key-123" />)
      await waitFor(() => {
        expect(screen.getByText('Terminate')).toBeInTheDocument()
      })
    })

    it('should call terminate API when terminate is clicked', async () => {
      ;(sessionApi.terminate as ReturnType<typeof vi.fn>).mockResolvedValue({})
      render(<SessionDetail sessionKey="test-session-key-123" />)
      await waitFor(() => {
        const terminateButton = screen.getByText('Terminate')
        fireEvent.click(terminateButton)
      })
      await waitFor(() => {
        expect(sessionApi.terminate).toHaveBeenCalledWith('test-session-key-123')
      })
    })

    it('should not show terminate button for terminated session', async () => {
      ;(sessionApi.get as ReturnType<typeof vi.fn>).mockResolvedValue({
        session: { ...mockSession, status: 'terminated' }
      })
      render(<SessionDetail sessionKey="test-session-key-123" />)
      await waitFor(() => {
        expect(screen.queryByText('Terminate')).not.toBeInTheDocument()
      })
    })
  })

  describe('send prompt', () => {
    it('should show prompt input for running session', async () => {
      render(<SessionDetail sessionKey="test-session-key-123" />)
      await waitFor(() => {
        expect(screen.getByPlaceholderText('Send a prompt to the AI agent...')).toBeInTheDocument()
      })
    })

    it('should call sendPrompt API when prompt is sent', async () => {
      ;(sessionApi.sendPrompt as ReturnType<typeof vi.fn>).mockResolvedValue({})
      render(<SessionDetail sessionKey="test-session-key-123" />)

      await waitFor(() => {
        const input = screen.getByPlaceholderText('Send a prompt to the AI agent...')
        fireEvent.change(input, { target: { value: 'Help me fix this bug' } })
      })

      const sendButton = screen.getByText('Send')
      fireEvent.click(sendButton)

      await waitFor(() => {
        expect(sessionApi.sendPrompt).toHaveBeenCalledWith('test-session-key-123', 'Help me fix this bug')
      })
    })

    it('should clear input after sending prompt', async () => {
      ;(sessionApi.sendPrompt as ReturnType<typeof vi.fn>).mockResolvedValue({})
      render(<SessionDetail sessionKey="test-session-key-123" />)

      await waitFor(() => {
        const input = screen.getByPlaceholderText('Send a prompt to the AI agent...')
        fireEvent.change(input, { target: { value: 'Help me' } })
      })

      const sendButton = screen.getByText('Send')
      fireEvent.click(sendButton)

      await waitFor(() => {
        const input = screen.getByPlaceholderText('Send a prompt to the AI agent...')
        expect(input).toHaveValue('')
      })
    })

    it('should not show prompt input for terminated session', async () => {
      ;(sessionApi.get as ReturnType<typeof vi.fn>).mockResolvedValue({
        session: { ...mockSession, status: 'terminated' }
      })
      render(<SessionDetail sessionKey="test-session-key-123" />)
      await waitFor(() => {
        expect(screen.queryByPlaceholderText('Send a prompt to the AI agent...')).not.toBeInTheDocument()
      })
    })
  })

  describe('status variations', () => {
    it('should show initializing status', async () => {
      ;(sessionApi.get as ReturnType<typeof vi.fn>).mockResolvedValue({
        session: { ...mockSession, status: 'initializing' }
      })
      render(<SessionDetail sessionKey="test-session-key-123" />)
      await waitFor(() => {
        expect(screen.getByText('Initializing')).toBeInTheDocument()
      })
    })

    it('should show terminated status', async () => {
      ;(sessionApi.get as ReturnType<typeof vi.fn>).mockResolvedValue({
        session: { ...mockSession, status: 'terminated' }
      })
      render(<SessionDetail sessionKey="test-session-key-123" />)
      await waitFor(() => {
        expect(screen.getByText('Terminated')).toBeInTheDocument()
      })
    })

    it('should show failed status', async () => {
      ;(sessionApi.get as ReturnType<typeof vi.fn>).mockResolvedValue({
        session: { ...mockSession, status: 'failed' }
      })
      render(<SessionDetail sessionKey="test-session-key-123" />)
      await waitFor(() => {
        expect(screen.getByText('Failed')).toBeInTheDocument()
      })
    })
  })
})
