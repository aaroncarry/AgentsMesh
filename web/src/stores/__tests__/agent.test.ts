import { describe, it, expect, beforeEach } from 'vitest'
import {
  useAgentStore,
  AgentType,
  CustomAgentType,
  OrganizationAgent,
  UserAgentCredentials,
} from '../agent'

describe('useAgentStore', () => {
  const mockAgentType: AgentType = {
    id: 1,
    slug: 'claude-code',
    name: 'Claude Code',
    description: 'AI coding assistant',
    launch_command: 'claude',
    default_args: '--model opus',
    credential_schema: [
      { name: 'api_key', type: 'secret', env_var: 'ANTHROPIC_API_KEY', required: true },
    ],
    is_builtin: true,
    is_active: true,
  }

  const mockCustomAgentType: CustomAgentType = {
    id: 2,
    slug: 'custom-agent',
    name: 'Custom Agent',
    launch_command: './custom-agent',
    organization_id: 1,
    credential_schema: [],
    is_active: true,
  }

  const mockOrgAgent: OrganizationAgent = {
    id: 1,
    organization_id: 1,
    agent_type_id: 1,
    agent_type: mockAgentType,
    is_enabled: true,
    is_default: true,
    has_credentials: true,
  }

  const mockUserCredential: UserAgentCredentials = {
    agent_type_id: 1,
    agent_slug: 'claude-code',
    has_credentials: true,
  }

  beforeEach(() => {
    useAgentStore.setState({
      builtinAgentTypes: [],
      customAgentTypes: [],
      organizationAgents: [],
      userCredentials: [],
      isLoading: false,
      error: null,
    })
  })

  describe('initial state', () => {
    it('should have empty builtinAgentTypes', () => {
      const state = useAgentStore.getState()
      expect(state.builtinAgentTypes).toEqual([])
    })

    it('should have empty customAgentTypes', () => {
      const state = useAgentStore.getState()
      expect(state.customAgentTypes).toEqual([])
    })

    it('should have empty organizationAgents', () => {
      const state = useAgentStore.getState()
      expect(state.organizationAgents).toEqual([])
    })

    it('should have empty userCredentials', () => {
      const state = useAgentStore.getState()
      expect(state.userCredentials).toEqual([])
    })

    it('should not be loading', () => {
      const state = useAgentStore.getState()
      expect(state.isLoading).toBe(false)
    })

    it('should have no error', () => {
      const state = useAgentStore.getState()
      expect(state.error).toBeNull()
    })
  })

  describe('setBuiltinAgentTypes', () => {
    it('should set builtin agent types', () => {
      useAgentStore.getState().setBuiltinAgentTypes([mockAgentType])

      const state = useAgentStore.getState()
      expect(state.builtinAgentTypes).toHaveLength(1)
      expect(state.builtinAgentTypes[0]).toEqual(mockAgentType)
    })
  })

  describe('setCustomAgentTypes', () => {
    it('should set custom agent types', () => {
      useAgentStore.getState().setCustomAgentTypes([mockCustomAgentType])

      const state = useAgentStore.getState()
      expect(state.customAgentTypes).toHaveLength(1)
      expect(state.customAgentTypes[0]).toEqual(mockCustomAgentType)
    })
  })

  describe('addCustomAgentType', () => {
    it('should add custom agent type', () => {
      useAgentStore.getState().addCustomAgentType(mockCustomAgentType)

      const state = useAgentStore.getState()
      expect(state.customAgentTypes).toHaveLength(1)
    })

    it('should append to existing custom agent types', () => {
      useAgentStore.setState({ customAgentTypes: [mockCustomAgentType] })
      const newAgent: CustomAgentType = { ...mockCustomAgentType, id: 3, slug: 'another-agent' }

      useAgentStore.getState().addCustomAgentType(newAgent)

      const state = useAgentStore.getState()
      expect(state.customAgentTypes).toHaveLength(2)
    })
  })

  describe('updateCustomAgentType', () => {
    it('should update custom agent type', () => {
      useAgentStore.setState({ customAgentTypes: [mockCustomAgentType] })

      useAgentStore.getState().updateCustomAgentType(2, { name: 'Updated Name' })

      const state = useAgentStore.getState()
      expect(state.customAgentTypes[0].name).toBe('Updated Name')
    })

    it('should not update non-matching agent type', () => {
      useAgentStore.setState({ customAgentTypes: [mockCustomAgentType] })

      useAgentStore.getState().updateCustomAgentType(999, { name: 'Updated' })

      const state = useAgentStore.getState()
      expect(state.customAgentTypes[0].name).toBe('Custom Agent')
    })
  })

  describe('removeCustomAgentType', () => {
    it('should remove custom agent type', () => {
      useAgentStore.setState({ customAgentTypes: [mockCustomAgentType] })

      useAgentStore.getState().removeCustomAgentType(2)

      const state = useAgentStore.getState()
      expect(state.customAgentTypes).toHaveLength(0)
    })

    it('should only remove matching agent type', () => {
      const agent2: CustomAgentType = { ...mockCustomAgentType, id: 3 }
      useAgentStore.setState({ customAgentTypes: [mockCustomAgentType, agent2] })

      useAgentStore.getState().removeCustomAgentType(2)

      const state = useAgentStore.getState()
      expect(state.customAgentTypes).toHaveLength(1)
      expect(state.customAgentTypes[0].id).toBe(3)
    })
  })

  describe('setOrganizationAgents', () => {
    it('should set organization agents', () => {
      useAgentStore.getState().setOrganizationAgents([mockOrgAgent])

      const state = useAgentStore.getState()
      expect(state.organizationAgents).toHaveLength(1)
      expect(state.organizationAgents[0]).toEqual(mockOrgAgent)
    })
  })

  describe('enableAgent', () => {
    it('should enable existing agent', () => {
      const disabledAgent = { ...mockOrgAgent, is_enabled: false, is_default: false }
      useAgentStore.setState({ organizationAgents: [disabledAgent] })

      useAgentStore.getState().enableAgent(1, false)

      const state = useAgentStore.getState()
      expect(state.organizationAgents[0].is_enabled).toBe(true)
    })

    it('should set agent as default', () => {
      const agent = { ...mockOrgAgent, is_enabled: false, is_default: false }
      useAgentStore.setState({ organizationAgents: [agent] })

      useAgentStore.getState().enableAgent(1, true)

      const state = useAgentStore.getState()
      expect(state.organizationAgents[0].is_default).toBe(true)
    })

    it('should clear is_default from other agents when setting new default', () => {
      const agent1 = { ...mockOrgAgent, agent_type_id: 1, is_default: true }
      const agent2 = { ...mockOrgAgent, id: 2, agent_type_id: 2, is_default: false }
      useAgentStore.setState({ organizationAgents: [agent1, agent2] })

      useAgentStore.getState().enableAgent(2, true)

      const state = useAgentStore.getState()
      expect(state.organizationAgents[0].is_default).toBe(false)
      expect(state.organizationAgents[1].is_default).toBe(true)
    })

    it('should do nothing for non-existent agent', () => {
      useAgentStore.setState({ organizationAgents: [mockOrgAgent] })

      useAgentStore.getState().enableAgent(999, true)

      const state = useAgentStore.getState()
      expect(state.organizationAgents).toHaveLength(1)
    })
  })

  describe('disableAgent', () => {
    it('should disable agent', () => {
      useAgentStore.setState({ organizationAgents: [mockOrgAgent] })

      useAgentStore.getState().disableAgent(1)

      const state = useAgentStore.getState()
      expect(state.organizationAgents[0].is_enabled).toBe(false)
    })

    it('should not affect other agents', () => {
      const agent2: OrganizationAgent = { ...mockOrgAgent, id: 2, agent_type_id: 2 }
      useAgentStore.setState({ organizationAgents: [mockOrgAgent, agent2] })

      useAgentStore.getState().disableAgent(1)

      const state = useAgentStore.getState()
      expect(state.organizationAgents[1].is_enabled).toBe(true)
    })
  })

  describe('setUserCredentials', () => {
    it('should set user credentials', () => {
      useAgentStore.getState().setUserCredentials([mockUserCredential])

      const state = useAgentStore.getState()
      expect(state.userCredentials).toHaveLength(1)
      expect(state.userCredentials[0]).toEqual(mockUserCredential)
    })
  })

  describe('updateUserCredential', () => {
    it('should update user credential', () => {
      useAgentStore.setState({ userCredentials: [mockUserCredential] })

      useAgentStore.getState().updateUserCredential(1, false)

      const state = useAgentStore.getState()
      expect(state.userCredentials[0].has_credentials).toBe(false)
    })

    it('should not update non-matching credential', () => {
      useAgentStore.setState({ userCredentials: [mockUserCredential] })

      useAgentStore.getState().updateUserCredential(999, false)

      const state = useAgentStore.getState()
      expect(state.userCredentials[0].has_credentials).toBe(true)
    })
  })

  describe('setLoading', () => {
    it('should set loading state', () => {
      useAgentStore.getState().setLoading(true)

      const state = useAgentStore.getState()
      expect(state.isLoading).toBe(true)
    })
  })

  describe('setError', () => {
    it('should set error', () => {
      useAgentStore.getState().setError('Test error')

      const state = useAgentStore.getState()
      expect(state.error).toBe('Test error')
    })

    it('should clear error', () => {
      useAgentStore.setState({ error: 'Previous error' })
      useAgentStore.getState().setError(null)

      const state = useAgentStore.getState()
      expect(state.error).toBeNull()
    })
  })

  describe('reset', () => {
    it('should reset to initial state', () => {
      useAgentStore.setState({
        builtinAgentTypes: [mockAgentType],
        customAgentTypes: [mockCustomAgentType],
        organizationAgents: [mockOrgAgent],
        userCredentials: [mockUserCredential],
        isLoading: true,
        error: 'Some error',
      })

      useAgentStore.getState().reset()

      const state = useAgentStore.getState()
      expect(state.builtinAgentTypes).toEqual([])
      expect(state.customAgentTypes).toEqual([])
      expect(state.organizationAgents).toEqual([])
      expect(state.userCredentials).toEqual([])
      expect(state.isLoading).toBe(false)
      expect(state.error).toBeNull()
    })
  })
})
