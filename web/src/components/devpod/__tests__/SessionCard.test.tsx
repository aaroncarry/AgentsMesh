import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { SessionCard } from '../SessionCard'

// Mock next/link
vi.mock('next/link', () => ({
  default: ({ children, href }: { children: React.ReactNode; href: string }) => (
    <a href={href}>{children}</a>
  ),
}))

describe('SessionCard Component', () => {
  const baseSession = {
    id: 1,
    sessionKey: 'abc12345-1234-5678-9abc-def012345678',
    status: 'running' as const,
    agentStatus: 'coding',
    createdAt: '2024-01-01T00:00:00Z',
  }

  describe('rendering', () => {
    it('should render session key (truncated)', () => {
      render(<SessionCard session={baseSession} />)
      expect(screen.getByText('abc12345')).toBeInTheDocument()
    })

    it('should render status badge', () => {
      render(<SessionCard session={baseSession} />)
      expect(screen.getByText('Running')).toBeInTheDocument()
    })
  })

  describe('status display', () => {
    it('should display initializing status', () => {
      render(<SessionCard session={{ ...baseSession, status: 'initializing' }} />)
      expect(screen.getByText('Initializing')).toBeInTheDocument()
    })

    it('should display running status', () => {
      render(<SessionCard session={{ ...baseSession, status: 'running' }} />)
      expect(screen.getByText('Running')).toBeInTheDocument()
    })

    it('should display paused status', () => {
      render(<SessionCard session={{ ...baseSession, status: 'paused' }} />)
      expect(screen.getByText('Paused')).toBeInTheDocument()
    })

    it('should display terminated status', () => {
      render(<SessionCard session={{ ...baseSession, status: 'terminated' }} />)
      expect(screen.getByText('Terminated')).toBeInTheDocument()
    })

    it('should display failed status', () => {
      render(<SessionCard session={{ ...baseSession, status: 'failed' }} />)
      expect(screen.getByText('Failed')).toBeInTheDocument()
    })
  })

  describe('agent type', () => {
    it('should display agent type when provided', () => {
      const sessionWithAgent = {
        ...baseSession,
        agentType: { id: 1, name: 'Claude Code', slug: 'claude-code' },
      }
      render(<SessionCard session={sessionWithAgent} />)
      expect(screen.getByText('Claude Code')).toBeInTheDocument()
    })

    it('should not display agent type when not provided', () => {
      render(<SessionCard session={baseSession} />)
      expect(screen.queryByText('Claude Code')).not.toBeInTheDocument()
    })
  })

  describe('repository display', () => {
    it('should display repository path when provided', () => {
      const sessionWithRepo = {
        ...baseSession,
        repository: { id: 1, name: 'my-repo', fullPath: 'org/my-repo' },
      }
      render(<SessionCard session={sessionWithRepo} />)
      expect(screen.getByText('org/my-repo')).toBeInTheDocument()
    })

    it('should display branch name when provided', () => {
      const sessionWithBranch = {
        ...baseSession,
        repository: { id: 1, name: 'my-repo', fullPath: 'org/my-repo' },
        branchName: 'feature/new-feature',
      }
      render(<SessionCard session={sessionWithBranch} />)
      expect(screen.getByText('feature/new-feature')).toBeInTheDocument()
    })
  })

  describe('ticket display', () => {
    it('should display ticket link when provided', () => {
      const sessionWithTicket = {
        ...baseSession,
        ticket: { id: 1, identifier: 'PROJ-42', title: 'Fix bug' },
      }
      render(<SessionCard session={sessionWithTicket} />)
      expect(screen.getByText('PROJ-42: Fix bug')).toBeInTheDocument()
      expect(screen.getByRole('link')).toHaveAttribute('href', '/tickets/PROJ-42')
    })
  })

  describe('initial prompt', () => {
    it('should display initial prompt when provided', () => {
      const sessionWithPrompt = {
        ...baseSession,
        initialPrompt: 'Implement the user authentication feature',
      }
      render(<SessionCard session={sessionWithPrompt} />)
      expect(screen.getByText('Implement the user authentication feature')).toBeInTheDocument()
    })
  })

  describe('runner display', () => {
    it('should display runner node ID when provided', () => {
      const sessionWithRunner = {
        ...baseSession,
        runner: { id: 1, nodeId: 'runner-001', status: 'online' },
      }
      render(<SessionCard session={sessionWithRunner} />)
      expect(screen.getByText('runner-001')).toBeInTheDocument()
    })

    it('should display Unknown when runner not provided', () => {
      render(<SessionCard session={baseSession} />)
      expect(screen.getByText('Unknown')).toBeInTheDocument()
    })
  })

  describe('agent status', () => {
    it('should display agent status when provided', () => {
      render(<SessionCard session={baseSession} />)
      expect(screen.getByText('coding')).toBeInTheDocument()
    })

    it('should not display agent status when unknown', () => {
      const sessionWithUnknown = {
        ...baseSession,
        agentStatus: 'unknown',
      }
      render(<SessionCard session={sessionWithUnknown} />)
      // Should not show the "Agent: unknown" line
      expect(screen.queryByText('unknown')).not.toBeInTheDocument()
    })
  })

  describe('actions for active sessions', () => {
    it('should show Open Terminal button for running session', () => {
      render(<SessionCard session={baseSession} />)
      expect(screen.getByText('Open Terminal')).toBeInTheDocument()
    })

    it('should show Terminate button for running session', () => {
      render(<SessionCard session={baseSession} />)
      expect(screen.getByText('Terminate')).toBeInTheDocument()
    })

    it('should show Open Terminal button for initializing session', () => {
      render(<SessionCard session={{ ...baseSession, status: 'initializing' }} />)
      expect(screen.getByText('Open Terminal')).toBeInTheDocument()
    })

    it('should call onOpen when Open Terminal clicked', () => {
      const handleOpen = vi.fn()
      render(<SessionCard session={baseSession} onOpen={handleOpen} />)
      fireEvent.click(screen.getByText('Open Terminal'))
      expect(handleOpen).toHaveBeenCalledWith(baseSession.sessionKey)
    })

    it('should call onTerminate when Terminate clicked', async () => {
      const handleTerminate = vi.fn().mockResolvedValue(undefined)
      render(<SessionCard session={baseSession} onTerminate={handleTerminate} />)
      fireEvent.click(screen.getByText('Terminate'))
      await waitFor(() => {
        expect(handleTerminate).toHaveBeenCalledWith(baseSession.sessionKey)
      })
    })

    it('should show loading state during termination', async () => {
      const handleTerminate = vi.fn().mockImplementation(() => new Promise(resolve => setTimeout(resolve, 100)))
      render(<SessionCard session={baseSession} onTerminate={handleTerminate} />)
      fireEvent.click(screen.getByText('Terminate'))
      expect(screen.getByText('...')).toBeInTheDocument()
    })

    it('should disable terminate button during termination', async () => {
      const handleTerminate = vi.fn().mockImplementation(() => new Promise(resolve => setTimeout(resolve, 100)))
      render(<SessionCard session={baseSession} onTerminate={handleTerminate} />)
      fireEvent.click(screen.getByText('Terminate'))
      expect(screen.getByText('...')).toBeDisabled()
    })
  })

  describe('actions for inactive sessions', () => {
    it('should show View Logs button for terminated session', () => {
      render(<SessionCard session={{ ...baseSession, status: 'terminated' }} />)
      expect(screen.getByText('View Logs')).toBeInTheDocument()
    })

    it('should not show Terminate button for terminated session', () => {
      render(<SessionCard session={{ ...baseSession, status: 'terminated' }} />)
      expect(screen.queryByText('Terminate')).not.toBeInTheDocument()
    })

    it('should show View Logs button for failed session', () => {
      render(<SessionCard session={{ ...baseSession, status: 'failed' }} />)
      expect(screen.getByText('View Logs')).toBeInTheDocument()
    })

    it('should call onOpen when View Logs clicked', () => {
      const handleOpen = vi.fn()
      render(
        <SessionCard
          session={{ ...baseSession, status: 'terminated' }}
          onOpen={handleOpen}
        />
      )
      fireEvent.click(screen.getByText('View Logs'))
      expect(handleOpen).toHaveBeenCalledWith(baseSession.sessionKey)
    })
  })

  describe('time display', () => {
    it('should format duration for active sessions', () => {
      const recentSession = {
        ...baseSession,
        startedAt: new Date(Date.now() - 30000).toISOString(), // 30 seconds ago
      }
      render(<SessionCard session={recentSession} />)
      // Should show duration in seconds
      expect(screen.getByText(/\d+s/)).toBeInTheDocument()
    })

    it('should display started time for inactive sessions', () => {
      const inactiveSession = {
        ...baseSession,
        status: 'terminated' as const,
        startedAt: '2024-01-01T10:00:00Z',
      }
      render(<SessionCard session={inactiveSession} />)
      // Should show formatted date/time
    })

    it('should show dash when startedAt not provided', () => {
      const sessionWithoutStart = {
        ...baseSession,
        status: 'terminated' as const,
      }
      render(<SessionCard session={sessionWithoutStart} />)
      expect(screen.getByText('—')).toBeInTheDocument()
    })
  })

  describe('edge cases', () => {
    it('should handle unknown status gracefully', () => {
      const sessionWithUnknownStatus = {
        ...baseSession,
        status: 'unknown' as any,
      }
      render(<SessionCard session={sessionWithUnknownStatus} />)
      // Should fall back to terminated styling and show View Logs
      expect(screen.getByText('View Logs')).toBeInTheDocument()
    })

    it('should not crash when onTerminate is not provided', () => {
      render(<SessionCard session={baseSession} />)
      fireEvent.click(screen.getByText('Terminate'))
      // Should not throw
    })

    it('should not crash when onOpen is not provided', () => {
      render(<SessionCard session={baseSession} />)
      fireEvent.click(screen.getByText('Open Terminal'))
      // Should not throw
    })
  })
})
