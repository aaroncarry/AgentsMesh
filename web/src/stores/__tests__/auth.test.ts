import { describe, it, expect, beforeEach } from 'vitest'
import { useAuthStore } from '../auth'

describe('useAuthStore', () => {
  beforeEach(() => {
    // Reset store state before each test
    useAuthStore.setState({
      token: null,
      user: null,
      currentOrg: null,
      organizations: [],
    })
  })

  describe('initial state', () => {
    it('should have null token', () => {
      const state = useAuthStore.getState()
      expect(state.token).toBeNull()
    })

    it('should have null user', () => {
      const state = useAuthStore.getState()
      expect(state.user).toBeNull()
    })

    it('should have null currentOrg', () => {
      const state = useAuthStore.getState()
      expect(state.currentOrg).toBeNull()
    })

    it('should have empty organizations', () => {
      const state = useAuthStore.getState()
      expect(state.organizations).toEqual([])
    })
  })

  describe('setAuth', () => {
    it('should set token and user', () => {
      const token = 'test-token'
      const user = {
        id: 1,
        email: 'test@example.com',
        username: 'testuser',
        name: 'Test User',
      }

      useAuthStore.getState().setAuth(token, user)

      const state = useAuthStore.getState()
      expect(state.token).toBe(token)
      expect(state.user).toEqual(user)
    })

    it('should handle user without optional fields', () => {
      const token = 'test-token'
      const user = {
        id: 1,
        email: 'test@example.com',
        username: 'testuser',
      }

      useAuthStore.getState().setAuth(token, user)

      const state = useAuthStore.getState()
      expect(state.user).toEqual(user)
      expect(state.user?.name).toBeUndefined()
      expect(state.user?.avatar_url).toBeUndefined()
    })
  })

  describe('setOrganizations', () => {
    it('should set organizations', () => {
      const orgs = [
        { id: 1, name: 'Org 1', slug: 'org-1', role: 'owner' },
        { id: 2, name: 'Org 2', slug: 'org-2', role: 'member' },
      ]

      useAuthStore.getState().setOrganizations(orgs)

      const state = useAuthStore.getState()
      expect(state.organizations).toEqual(orgs)
    })

    it('should auto-select first org if none selected', () => {
      const orgs = [
        { id: 1, name: 'Org 1', slug: 'org-1', role: 'owner' },
        { id: 2, name: 'Org 2', slug: 'org-2', role: 'member' },
      ]

      useAuthStore.getState().setOrganizations(orgs)

      const state = useAuthStore.getState()
      expect(state.currentOrg).toEqual(orgs[0])
    })

    it('should not change currentOrg if already selected', () => {
      const existingOrg = { id: 3, name: 'Existing', slug: 'existing', role: 'admin' }
      useAuthStore.setState({ currentOrg: existingOrg })

      const orgs = [
        { id: 1, name: 'Org 1', slug: 'org-1', role: 'owner' },
        { id: 2, name: 'Org 2', slug: 'org-2', role: 'member' },
      ]

      useAuthStore.getState().setOrganizations(orgs)

      const state = useAuthStore.getState()
      expect(state.currentOrg).toEqual(existingOrg)
    })

    it('should handle empty organizations array', () => {
      useAuthStore.getState().setOrganizations([])

      const state = useAuthStore.getState()
      expect(state.organizations).toEqual([])
      expect(state.currentOrg).toBeNull()
    })
  })

  describe('setCurrentOrg', () => {
    it('should set current organization', () => {
      const org = { id: 1, name: 'Test Org', slug: 'test-org', role: 'owner' }

      useAuthStore.getState().setCurrentOrg(org)

      const state = useAuthStore.getState()
      expect(state.currentOrg).toEqual(org)
    })

    it('should handle org with logo_url', () => {
      const org = {
        id: 1,
        name: 'Test Org',
        slug: 'test-org',
        role: 'owner',
        logo_url: 'https://example.com/logo.png',
      }

      useAuthStore.getState().setCurrentOrg(org)

      const state = useAuthStore.getState()
      expect(state.currentOrg?.logo_url).toBe('https://example.com/logo.png')
    })
  })

  describe('logout', () => {
    it('should clear all auth state', () => {
      // Setup authenticated state
      useAuthStore.setState({
        token: 'test-token',
        user: { id: 1, email: 'test@example.com', username: 'testuser' },
        currentOrg: { id: 1, name: 'Org', slug: 'org', role: 'owner' },
        organizations: [{ id: 1, name: 'Org', slug: 'org', role: 'owner' }],
      })

      useAuthStore.getState().logout()

      const state = useAuthStore.getState()
      expect(state.token).toBeNull()
      expect(state.user).toBeNull()
      expect(state.currentOrg).toBeNull()
      expect(state.organizations).toEqual([])
    })
  })

  describe('isAuthenticated', () => {
    it('should return false when no token', () => {
      const result = useAuthStore.getState().isAuthenticated()
      expect(result).toBe(false)
    })

    it('should return true when token exists', () => {
      useAuthStore.setState({ token: 'test-token' })

      const result = useAuthStore.getState().isAuthenticated()
      expect(result).toBe(true)
    })

    it('should return false after logout', () => {
      useAuthStore.setState({ token: 'test-token' })
      useAuthStore.getState().logout()

      const result = useAuthStore.getState().isAuthenticated()
      expect(result).toBe(false)
    })
  })
})
