import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { PodDetail } from '../PodDetail'
import { podApi } from '@/lib/api/client'

// Mock next/navigation
const mockRouterBack = vi.fn()
const mockRouterPush = vi.fn()
vi.mock('next/navigation', () => ({
  useRouter: () => ({
    back: mockRouterBack,
    push: mockRouterPush,
  }),
}))

// Mock pod API
vi.mock('@/lib/api/client', () => ({
  podApi: {
    get: vi.fn(),
    terminate: vi.fn(),
    sendPrompt: vi.fn(),
  },
}))

// Mock Terminal component
vi.mock('../Terminal', () => ({
  Terminal: ({ podKey, className }: { podKey: string; className?: string }) => (
    <div data-testid="terminal" data-pod-key={podKey} className={className}>
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

describe('PodDetail Component', () => {
  const mockPod = {
    id: 1,
    pod_key: 'test-pod-key-123',
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
    ;(podApi.get as ReturnType<typeof vi.fn>).mockResolvedValue({ pod: mockPod })
  })

  describe('rendering', () => {
    it('should render pod key', async () => {
      render(<PodDetail podKey="test-pod-key-123" />)
      await waitFor(() => {
        expect(screen.getByText('test-pod-key-123')).toBeInTheDocument()
      })
    })

    it('should render status badge', async () => {
      render(<PodDetail podKey="test-pod-key-123" />)
      await waitFor(() => {
        expect(screen.getByText('Running')).toBeInTheDocument()
      })
    })

    it('should render agent type', async () => {
      render(<PodDetail podKey="test-pod-key-123" />)
      await waitFor(() => {
        expect(screen.getByText('Claude Code')).toBeInTheDocument()
      })
    })

    it('should render runner info', async () => {
      render(<PodDetail podKey="test-pod-key-123" />)
      await waitFor(() => {
        expect(screen.getByText('runner-001')).toBeInTheDocument()
      })
    })

    it('should render terminal', async () => {
      render(<PodDetail podKey="test-pod-key-123" />)
      await waitFor(() => {
        expect(screen.getByTestId('terminal')).toBeInTheDocument()
      })
    })

    it('should call podApi.get on mount', async () => {
      render(<PodDetail podKey="test-pod-key-123" />)
      await waitFor(() => {
        expect(podApi.get).toHaveBeenCalledWith('test-pod-key-123')
      })
    })
  })

  describe('loading state', () => {
    it('should render skeleton when loading', () => {
      ;(podApi.get as ReturnType<typeof vi.fn>).mockImplementation(
        () => new Promise(() => {}) // Never resolves
      )
      render(<PodDetail podKey="test-pod-key-123" />)
      expect(screen.getByTestId('pod-detail-skeleton')).toBeInTheDocument()
    })
  })

  describe('error state', () => {
    it('should render error message', async () => {
      ;(podApi.get as ReturnType<typeof vi.fn>).mockRejectedValue(new Error('Failed'))
      render(<PodDetail podKey="test-pod-key-123" />)
      await waitFor(() => {
        expect(screen.getByText('Failed to load pod')).toBeInTheDocument()
      })
    })

    it('should render retry button on error', async () => {
      ;(podApi.get as ReturnType<typeof vi.fn>).mockRejectedValue(new Error('Failed'))
      render(<PodDetail podKey="test-pod-key-123" />)
      await waitFor(() => {
        expect(screen.getByText('Retry')).toBeInTheDocument()
      })
    })
  })

  describe('not found state', () => {
    it('should render not found message when pod is null', async () => {
      ;(podApi.get as ReturnType<typeof vi.fn>).mockResolvedValue({ pod: null })
      render(<PodDetail podKey="test-pod-key-123" />)
      await waitFor(() => {
        expect(screen.getByText('Pod not found')).toBeInTheDocument()
      })
    })
  })

  describe('repository info', () => {
    it('should display repository path', async () => {
      render(<PodDetail podKey="test-pod-key-123" />)
      await waitFor(() => {
        expect(screen.getByText('org/my-project')).toBeInTheDocument()
      })
    })

    it('should display branch name', async () => {
      render(<PodDetail podKey="test-pod-key-123" />)
      await waitFor(() => {
        expect(screen.getByText('feature/test')).toBeInTheDocument()
      })
    })
  })

  describe('linked ticket', () => {
    it('should display ticket identifier', async () => {
      render(<PodDetail podKey="test-pod-key-123" />)
      await waitFor(() => {
        expect(screen.getByText('PROJ-42')).toBeInTheDocument()
      })
    })

    it('should display ticket title', async () => {
      render(<PodDetail podKey="test-pod-key-123" />)
      await waitFor(() => {
        expect(screen.getByText('Implement new feature')).toBeInTheDocument()
      })
    })

    it('should navigate to ticket on click', async () => {
      render(<PodDetail podKey="test-pod-key-123" />)
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
      render(<PodDetail podKey="test-pod-key-123" />)
      await waitFor(() => {
        expect(screen.getByText('Hello, help me with coding')).toBeInTheDocument()
      })
    })
  })

  describe('created by', () => {
    it('should display creator name', async () => {
      render(<PodDetail podKey="test-pod-key-123" />)
      await waitFor(() => {
        expect(screen.getByText('John Doe')).toBeInTheDocument()
      })
    })

    it('should display creator username', async () => {
      render(<PodDetail podKey="test-pod-key-123" />)
      await waitFor(() => {
        expect(screen.getByText('@john')).toBeInTheDocument()
      })
    })
  })

  describe('navigation', () => {
    it('should navigate back when back button is clicked', async () => {
      render(<PodDetail podKey="test-pod-key-123" />)
      await waitFor(() => {
        const backButton = screen.getByText('Back')
        fireEvent.click(backButton)
        expect(mockRouterBack).toHaveBeenCalled()
      })
    })
  })

  describe('terminate action', () => {
    it('should show terminate button for running pod', async () => {
      render(<PodDetail podKey="test-pod-key-123" />)
      await waitFor(() => {
        expect(screen.getByText('Terminate')).toBeInTheDocument()
      })
    })

    it('should call terminate API when terminate is clicked', async () => {
      ;(podApi.terminate as ReturnType<typeof vi.fn>).mockResolvedValue({})
      render(<PodDetail podKey="test-pod-key-123" />)
      await waitFor(() => {
        const terminateButton = screen.getByText('Terminate')
        fireEvent.click(terminateButton)
      })
      await waitFor(() => {
        expect(podApi.terminate).toHaveBeenCalledWith('test-pod-key-123')
      })
    })

    it('should not show terminate button for terminated pod', async () => {
      ;(podApi.get as ReturnType<typeof vi.fn>).mockResolvedValue({
        pod: { ...mockPod, status: 'terminated' }
      })
      render(<PodDetail podKey="test-pod-key-123" />)
      await waitFor(() => {
        expect(screen.queryByText('Terminate')).not.toBeInTheDocument()
      })
    })
  })

  // Note: send prompt tests removed as the prompt input feature was removed from PodDetail component

  describe('status variations', () => {
    it('should show initializing status', async () => {
      ;(podApi.get as ReturnType<typeof vi.fn>).mockResolvedValue({
        pod: { ...mockPod, status: 'initializing' }
      })
      render(<PodDetail podKey="test-pod-key-123" />)
      await waitFor(() => {
        expect(screen.getByText('Initializing')).toBeInTheDocument()
      })
    })

    it('should show terminated status', async () => {
      ;(podApi.get as ReturnType<typeof vi.fn>).mockResolvedValue({
        pod: { ...mockPod, status: 'terminated' }
      })
      render(<PodDetail podKey="test-pod-key-123" />)
      await waitFor(() => {
        expect(screen.getByText('Terminated')).toBeInTheDocument()
      })
    })

    it('should show failed status', async () => {
      ;(podApi.get as ReturnType<typeof vi.fn>).mockResolvedValue({
        pod: { ...mockPod, status: 'failed' }
      })
      render(<PodDetail podKey="test-pod-key-123" />)
      await waitFor(() => {
        expect(screen.getByText('Failed')).toBeInTheDocument()
      })
    })
  })
})
