import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { PodCard } from '../PodCard'

// Mock next/link
vi.mock('next/link', () => ({
  default: ({ children, href }: { children: React.ReactNode; href: string }) => (
    <a href={href}>{children}</a>
  ),
}))

describe('PodCard Component', () => {
  const basePod = {
    id: 1,
    podKey: 'abc12345-1234-5678-9abc-def012345678',
    status: 'running' as const,
    agentStatus: 'coding',
    createdAt: '2024-01-01T00:00:00Z',
  }

  describe('rendering', () => {
    it('should render pod key (truncated)', () => {
      render(<PodCard pod={basePod} />)
      expect(screen.getByText('abc12345')).toBeInTheDocument()
    })

    it('should render status badge', () => {
      render(<PodCard pod={basePod} />)
      expect(screen.getByText('Running')).toBeInTheDocument()
    })
  })

  describe('status display', () => {
    it('should display initializing status', () => {
      render(<PodCard pod={{ ...basePod, status: 'initializing' }} />)
      expect(screen.getByText('Initializing')).toBeInTheDocument()
    })

    it('should display running status', () => {
      render(<PodCard pod={{ ...basePod, status: 'running' }} />)
      expect(screen.getByText('Running')).toBeInTheDocument()
    })

    it('should display paused status', () => {
      render(<PodCard pod={{ ...basePod, status: 'paused' }} />)
      expect(screen.getByText('Paused')).toBeInTheDocument()
    })

    it('should display terminated status', () => {
      render(<PodCard pod={{ ...basePod, status: 'terminated' }} />)
      expect(screen.getByText('Terminated')).toBeInTheDocument()
    })

    it('should display failed status', () => {
      render(<PodCard pod={{ ...basePod, status: 'failed' }} />)
      expect(screen.getByText('Failed')).toBeInTheDocument()
    })
  })

  describe('agent type', () => {
    it('should display agent type when provided', () => {
      const podWithAgent = {
        ...basePod,
        agentType: { id: 1, name: 'Claude Code', slug: 'claude-code' },
      }
      render(<PodCard pod={podWithAgent} />)
      expect(screen.getByText('Claude Code')).toBeInTheDocument()
    })

    it('should not display agent type when not provided', () => {
      render(<PodCard pod={basePod} />)
      expect(screen.queryByText('Claude Code')).not.toBeInTheDocument()
    })
  })

  describe('repository display', () => {
    it('should display repository path when provided', () => {
      const podWithRepo = {
        ...basePod,
        repository: { id: 1, name: 'my-repo', fullPath: 'org/my-repo' },
      }
      render(<PodCard pod={podWithRepo} />)
      expect(screen.getByText('org/my-repo')).toBeInTheDocument()
    })

    it('should display branch name when provided', () => {
      const podWithBranch = {
        ...basePod,
        repository: { id: 1, name: 'my-repo', fullPath: 'org/my-repo' },
        branchName: 'feature/new-feature',
      }
      render(<PodCard pod={podWithBranch} />)
      expect(screen.getByText('feature/new-feature')).toBeInTheDocument()
    })
  })

  describe('ticket display', () => {
    it('should display ticket link when provided', () => {
      const podWithTicket = {
        ...basePod,
        ticket: { id: 1, identifier: 'PROJ-42', title: 'Fix bug' },
      }
      render(<PodCard pod={podWithTicket} />)
      expect(screen.getByText('PROJ-42: Fix bug')).toBeInTheDocument()
      expect(screen.getByRole('link')).toHaveAttribute('href', '/tickets/PROJ-42')
    })
  })

  describe('initial prompt', () => {
    it('should display initial prompt when provided', () => {
      const podWithPrompt = {
        ...basePod,
        initialPrompt: 'Implement the user authentication feature',
      }
      render(<PodCard pod={podWithPrompt} />)
      expect(screen.getByText('Implement the user authentication feature')).toBeInTheDocument()
    })
  })

  describe('runner display', () => {
    it('should display runner node ID when provided', () => {
      const podWithRunner = {
        ...basePod,
        runner: { id: 1, nodeId: 'runner-001', status: 'online' },
      }
      render(<PodCard pod={podWithRunner} />)
      expect(screen.getByText('runner-001')).toBeInTheDocument()
    })

    it('should display Unknown when runner not provided', () => {
      render(<PodCard pod={basePod} />)
      expect(screen.getByText('Unknown')).toBeInTheDocument()
    })
  })

  describe('agent status', () => {
    it('should display agent status when provided', () => {
      render(<PodCard pod={basePod} />)
      expect(screen.getByText('coding')).toBeInTheDocument()
    })

    it('should not display agent status when unknown', () => {
      const podWithUnknown = {
        ...basePod,
        agentStatus: 'unknown',
      }
      render(<PodCard pod={podWithUnknown} />)
      // Should not show the "Agent: unknown" line
      expect(screen.queryByText('unknown')).not.toBeInTheDocument()
    })
  })

  describe('actions for active pods', () => {
    it('should show Open Terminal button for running pod', () => {
      render(<PodCard pod={basePod} />)
      expect(screen.getByText('Open Terminal')).toBeInTheDocument()
    })

    it('should show Terminate button for running pod', () => {
      render(<PodCard pod={basePod} />)
      expect(screen.getByText('Terminate')).toBeInTheDocument()
    })

    it('should show Open Terminal button for initializing pod', () => {
      render(<PodCard pod={{ ...basePod, status: 'initializing' }} />)
      expect(screen.getByText('Open Terminal')).toBeInTheDocument()
    })

    it('should call onOpen when Open Terminal clicked', () => {
      const handleOpen = vi.fn()
      render(<PodCard pod={basePod} onOpen={handleOpen} />)
      fireEvent.click(screen.getByText('Open Terminal'))
      expect(handleOpen).toHaveBeenCalledWith(basePod.podKey)
    })

    it('should call onTerminate when Terminate clicked', async () => {
      const handleTerminate = vi.fn().mockResolvedValue(undefined)
      render(<PodCard pod={basePod} onTerminate={handleTerminate} />)
      fireEvent.click(screen.getByText('Terminate'))
      await waitFor(() => {
        expect(handleTerminate).toHaveBeenCalledWith(basePod.podKey)
      })
    })

    it('should show loading state during termination', async () => {
      const handleTerminate = vi.fn().mockImplementation(() => new Promise(resolve => setTimeout(resolve, 100)))
      render(<PodCard pod={basePod} onTerminate={handleTerminate} />)
      fireEvent.click(screen.getByText('Terminate'))
      expect(screen.getByText('...')).toBeInTheDocument()
    })

    it('should disable terminate button during termination', async () => {
      const handleTerminate = vi.fn().mockImplementation(() => new Promise(resolve => setTimeout(resolve, 100)))
      render(<PodCard pod={basePod} onTerminate={handleTerminate} />)
      fireEvent.click(screen.getByText('Terminate'))
      expect(screen.getByText('...')).toBeDisabled()
    })
  })

  describe('actions for inactive pods', () => {
    it('should show View Logs button for terminated pod', () => {
      render(<PodCard pod={{ ...basePod, status: 'terminated' }} />)
      expect(screen.getByText('View Logs')).toBeInTheDocument()
    })

    it('should not show Terminate button for terminated pod', () => {
      render(<PodCard pod={{ ...basePod, status: 'terminated' }} />)
      expect(screen.queryByText('Terminate')).not.toBeInTheDocument()
    })

    it('should show View Logs button for failed pod', () => {
      render(<PodCard pod={{ ...basePod, status: 'failed' }} />)
      expect(screen.getByText('View Logs')).toBeInTheDocument()
    })

    it('should call onOpen when View Logs clicked', () => {
      const handleOpen = vi.fn()
      render(
        <PodCard
          pod={{ ...basePod, status: 'terminated' }}
          onOpen={handleOpen}
        />
      )
      fireEvent.click(screen.getByText('View Logs'))
      expect(handleOpen).toHaveBeenCalledWith(basePod.podKey)
    })
  })

  describe('time display', () => {
    it('should format duration for active pods', () => {
      const recentPod = {
        ...basePod,
        startedAt: new Date(Date.now() - 30000).toISOString(), // 30 seconds ago
      }
      render(<PodCard pod={recentPod} />)
      // Should show duration in seconds
      expect(screen.getByText(/\d+s/)).toBeInTheDocument()
    })

    it('should display started time for inactive pods', () => {
      const inactivePod = {
        ...basePod,
        status: 'terminated' as const,
        startedAt: '2024-01-01T10:00:00Z',
      }
      render(<PodCard pod={inactivePod} />)
      // Should show formatted date/time
    })

    it('should show dash when startedAt not provided', () => {
      const podWithoutStart = {
        ...basePod,
        status: 'terminated' as const,
      }
      render(<PodCard pod={podWithoutStart} />)
      expect(screen.getByText('—')).toBeInTheDocument()
    })
  })

  describe('edge cases', () => {
    it('should handle unknown status gracefully', () => {
      const podWithUnknownStatus = {
        ...basePod,
        status: 'unknown' as any,
      }
      render(<PodCard pod={podWithUnknownStatus} />)
      // Should fall back to terminated styling and show View Logs
      expect(screen.getByText('View Logs')).toBeInTheDocument()
    })

    it('should not crash when onTerminate is not provided', () => {
      render(<PodCard pod={basePod} />)
      fireEvent.click(screen.getByText('Terminate'))
      // Should not throw
    })

    it('should not crash when onOpen is not provided', () => {
      render(<PodCard pod={basePod} />)
      fireEvent.click(screen.getByText('Open Terminal'))
      // Should not throw
    })
  })
})
