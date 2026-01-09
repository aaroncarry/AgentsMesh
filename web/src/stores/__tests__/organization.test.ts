import { describe, it, expect, beforeEach } from 'vitest'
import { useOrganizationStore, Organization, OrganizationMember } from '../organization'

describe('useOrganizationStore', () => {
  const mockOrg: Organization = {
    id: 1,
    name: 'Test Org',
    slug: 'test-org',
    logo_url: 'https://example.com/logo.png',
    subscription_plan: 'pro',
    subscription_status: 'active',
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  }

  const mockMember: OrganizationMember = {
    id: 1,
    user_id: 1,
    username: 'testuser',
    email: 'test@example.com',
    name: 'Test User',
    avatar_url: 'https://example.com/avatar.png',
    role: 'owner',
    joined_at: '2024-01-01T00:00:00Z',
  }

  beforeEach(() => {
    useOrganizationStore.setState({
      organizations: [],
      currentOrganization: null,
      members: [],
      isLoading: false,
      error: null,
    })
  })

  describe('initial state', () => {
    it('should have empty organizations', () => {
      const state = useOrganizationStore.getState()
      expect(state.organizations).toEqual([])
    })

    it('should have null currentOrganization', () => {
      const state = useOrganizationStore.getState()
      expect(state.currentOrganization).toBeNull()
    })

    it('should have empty members', () => {
      const state = useOrganizationStore.getState()
      expect(state.members).toEqual([])
    })
  })

  describe('setOrganizations', () => {
    it('should set organizations', () => {
      useOrganizationStore.getState().setOrganizations([mockOrg])

      const state = useOrganizationStore.getState()
      expect(state.organizations).toHaveLength(1)
      expect(state.organizations[0]).toEqual(mockOrg)
    })
  })

  describe('setCurrentOrganization', () => {
    it('should set current organization', () => {
      useOrganizationStore.getState().setCurrentOrganization(mockOrg)

      const state = useOrganizationStore.getState()
      expect(state.currentOrganization).toEqual(mockOrg)
    })

    it('should set current organization to null', () => {
      useOrganizationStore.setState({ currentOrganization: mockOrg })
      useOrganizationStore.getState().setCurrentOrganization(null)

      const state = useOrganizationStore.getState()
      expect(state.currentOrganization).toBeNull()
    })
  })

  describe('addOrganization', () => {
    it('should add organization', () => {
      useOrganizationStore.getState().addOrganization(mockOrg)

      const state = useOrganizationStore.getState()
      expect(state.organizations).toHaveLength(1)
    })

    it('should append to existing organizations', () => {
      const org2: Organization = { ...mockOrg, id: 2, slug: 'org-2' }
      useOrganizationStore.setState({ organizations: [mockOrg] })

      useOrganizationStore.getState().addOrganization(org2)

      const state = useOrganizationStore.getState()
      expect(state.organizations).toHaveLength(2)
    })
  })

  describe('updateOrganization', () => {
    it('should update organization in list', () => {
      useOrganizationStore.setState({ organizations: [mockOrg] })

      useOrganizationStore.getState().updateOrganization(1, { name: 'Updated Name' })

      const state = useOrganizationStore.getState()
      expect(state.organizations[0].name).toBe('Updated Name')
    })

    it('should update currentOrganization if same id', () => {
      useOrganizationStore.setState({
        organizations: [mockOrg],
        currentOrganization: mockOrg,
      })

      useOrganizationStore.getState().updateOrganization(1, { name: 'Updated Name' })

      const state = useOrganizationStore.getState()
      expect(state.currentOrganization?.name).toBe('Updated Name')
    })

    it('should not update currentOrganization if different id', () => {
      const org2: Organization = { ...mockOrg, id: 2 }
      useOrganizationStore.setState({
        organizations: [mockOrg, org2],
        currentOrganization: mockOrg,
      })

      useOrganizationStore.getState().updateOrganization(2, { name: 'Updated Name' })

      const state = useOrganizationStore.getState()
      expect(state.currentOrganization?.name).toBe('Test Org')
    })
  })

  describe('removeOrganization', () => {
    it('should remove organization from list', () => {
      useOrganizationStore.setState({ organizations: [mockOrg] })

      useOrganizationStore.getState().removeOrganization(1)

      const state = useOrganizationStore.getState()
      expect(state.organizations).toHaveLength(0)
    })

    it('should clear currentOrganization if removed', () => {
      useOrganizationStore.setState({
        organizations: [mockOrg],
        currentOrganization: mockOrg,
      })

      useOrganizationStore.getState().removeOrganization(1)

      const state = useOrganizationStore.getState()
      expect(state.currentOrganization).toBeNull()
    })

    it('should not clear currentOrganization if different id', () => {
      const org2: Organization = { ...mockOrg, id: 2 }
      useOrganizationStore.setState({
        organizations: [mockOrg, org2],
        currentOrganization: mockOrg,
      })

      useOrganizationStore.getState().removeOrganization(2)

      const state = useOrganizationStore.getState()
      expect(state.currentOrganization).toEqual(mockOrg)
    })
  })

  describe('member management', () => {
    describe('setMembers', () => {
      it('should set members', () => {
        useOrganizationStore.getState().setMembers([mockMember])

        const state = useOrganizationStore.getState()
        expect(state.members).toHaveLength(1)
        expect(state.members[0]).toEqual(mockMember)
      })
    })

    describe('addMember', () => {
      it('should add member', () => {
        useOrganizationStore.getState().addMember(mockMember)

        const state = useOrganizationStore.getState()
        expect(state.members).toHaveLength(1)
      })
    })

    describe('updateMember', () => {
      it('should update member by user_id', () => {
        useOrganizationStore.setState({ members: [mockMember] })

        useOrganizationStore.getState().updateMember(1, { role: 'admin' })

        const state = useOrganizationStore.getState()
        expect(state.members[0].role).toBe('admin')
      })

      it('should not update non-matching member', () => {
        useOrganizationStore.setState({ members: [mockMember] })

        useOrganizationStore.getState().updateMember(999, { role: 'admin' })

        const state = useOrganizationStore.getState()
        expect(state.members[0].role).toBe('owner')
      })
    })

    describe('removeMember', () => {
      it('should remove member by user_id', () => {
        useOrganizationStore.setState({ members: [mockMember] })

        useOrganizationStore.getState().removeMember(1)

        const state = useOrganizationStore.getState()
        expect(state.members).toHaveLength(0)
      })
    })
  })

  describe('setLoading', () => {
    it('should set loading state', () => {
      useOrganizationStore.getState().setLoading(true)

      const state = useOrganizationStore.getState()
      expect(state.isLoading).toBe(true)
    })
  })

  describe('setError', () => {
    it('should set error', () => {
      useOrganizationStore.getState().setError('Test error')

      const state = useOrganizationStore.getState()
      expect(state.error).toBe('Test error')
    })
  })

  describe('reset', () => {
    it('should reset to initial state', () => {
      useOrganizationStore.setState({
        organizations: [mockOrg],
        currentOrganization: mockOrg,
        members: [mockMember],
        isLoading: true,
        error: 'Some error',
      })

      useOrganizationStore.getState().reset()

      const state = useOrganizationStore.getState()
      expect(state.organizations).toEqual([])
      expect(state.currentOrganization).toBeNull()
      expect(state.members).toEqual([])
      expect(state.isLoading).toBe(false)
      expect(state.error).toBeNull()
    })
  })
})
