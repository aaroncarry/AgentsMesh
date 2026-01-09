import { describe, it, expect, beforeEach } from 'vitest'
import { useUserStore, UserProfile, UserIdentity } from '../user'

describe('useUserStore', () => {
  const mockProfile: UserProfile = {
    id: 1,
    email: 'test@example.com',
    username: 'testuser',
    name: 'Test User',
    avatar_url: 'https://example.com/avatar.png',
    is_active: true,
    last_login_at: '2024-01-01T00:00:00Z',
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
    identities: [],
    organizations: [],
  }

  const mockIdentity: UserIdentity = {
    id: 1,
    provider: 'github',
    provider_user_id: '12345',
    provider_username: 'testuser',
    created_at: '2024-01-01T00:00:00Z',
  }

  beforeEach(() => {
    useUserStore.setState({
      profile: null,
      isLoading: false,
      error: null,
    })
  })

  describe('initial state', () => {
    it('should have null profile', () => {
      const state = useUserStore.getState()
      expect(state.profile).toBeNull()
    })

    it('should have isLoading false', () => {
      const state = useUserStore.getState()
      expect(state.isLoading).toBe(false)
    })

    it('should have null error', () => {
      const state = useUserStore.getState()
      expect(state.error).toBeNull()
    })
  })

  describe('setProfile', () => {
    it('should set profile', () => {
      useUserStore.getState().setProfile(mockProfile)

      const state = useUserStore.getState()
      expect(state.profile).toEqual(mockProfile)
    })

    it('should set profile to null', () => {
      useUserStore.setState({ profile: mockProfile })
      useUserStore.getState().setProfile(null)

      const state = useUserStore.getState()
      expect(state.profile).toBeNull()
    })
  })

  describe('updateProfile', () => {
    it('should update profile fields', () => {
      useUserStore.setState({ profile: mockProfile })

      useUserStore.getState().updateProfile({ name: 'Updated Name' })

      const state = useUserStore.getState()
      expect(state.profile?.name).toBe('Updated Name')
      expect(state.profile?.email).toBe(mockProfile.email)
    })

    it('should update multiple fields', () => {
      useUserStore.setState({ profile: mockProfile })

      useUserStore.getState().updateProfile({
        name: 'New Name',
        avatar_url: 'https://new-avatar.com/img.png',
      })

      const state = useUserStore.getState()
      expect(state.profile?.name).toBe('New Name')
      expect(state.profile?.avatar_url).toBe('https://new-avatar.com/img.png')
    })

    it('should do nothing if profile is null', () => {
      useUserStore.getState().updateProfile({ name: 'Updated Name' })

      const state = useUserStore.getState()
      expect(state.profile).toBeNull()
    })
  })

  describe('addIdentity', () => {
    it('should add identity to profile', () => {
      useUserStore.setState({
        profile: { ...mockProfile, identities: [] },
      })

      useUserStore.getState().addIdentity(mockIdentity)

      const state = useUserStore.getState()
      expect(state.profile?.identities).toHaveLength(1)
      expect(state.profile?.identities[0]).toEqual(mockIdentity)
    })

    it('should append to existing identities', () => {
      const existingIdentity: UserIdentity = {
        id: 2,
        provider: 'google',
        provider_user_id: '67890',
        created_at: '2024-01-01T00:00:00Z',
      }
      useUserStore.setState({
        profile: { ...mockProfile, identities: [existingIdentity] },
      })

      useUserStore.getState().addIdentity(mockIdentity)

      const state = useUserStore.getState()
      expect(state.profile?.identities).toHaveLength(2)
    })

    it('should do nothing if profile is null', () => {
      useUserStore.getState().addIdentity(mockIdentity)

      const state = useUserStore.getState()
      expect(state.profile).toBeNull()
    })
  })

  describe('removeIdentity', () => {
    it('should remove identity by provider', () => {
      useUserStore.setState({
        profile: { ...mockProfile, identities: [mockIdentity] },
      })

      useUserStore.getState().removeIdentity('github')

      const state = useUserStore.getState()
      expect(state.profile?.identities).toHaveLength(0)
    })

    it('should only remove matching provider', () => {
      const googleIdentity: UserIdentity = {
        id: 2,
        provider: 'google',
        provider_user_id: '67890',
        created_at: '2024-01-01T00:00:00Z',
      }
      useUserStore.setState({
        profile: { ...mockProfile, identities: [mockIdentity, googleIdentity] },
      })

      useUserStore.getState().removeIdentity('github')

      const state = useUserStore.getState()
      expect(state.profile?.identities).toHaveLength(1)
      expect(state.profile?.identities[0].provider).toBe('google')
    })

    it('should do nothing if profile is null', () => {
      useUserStore.getState().removeIdentity('github')

      const state = useUserStore.getState()
      expect(state.profile).toBeNull()
    })
  })

  describe('setLoading', () => {
    it('should set loading to true', () => {
      useUserStore.getState().setLoading(true)

      const state = useUserStore.getState()
      expect(state.isLoading).toBe(true)
    })

    it('should set loading to false', () => {
      useUserStore.setState({ isLoading: true })
      useUserStore.getState().setLoading(false)

      const state = useUserStore.getState()
      expect(state.isLoading).toBe(false)
    })
  })

  describe('setError', () => {
    it('should set error message', () => {
      useUserStore.getState().setError('Test error')

      const state = useUserStore.getState()
      expect(state.error).toBe('Test error')
    })

    it('should clear error', () => {
      useUserStore.setState({ error: 'Previous error' })
      useUserStore.getState().setError(null)

      const state = useUserStore.getState()
      expect(state.error).toBeNull()
    })
  })

  describe('reset', () => {
    it('should reset to initial state', () => {
      useUserStore.setState({
        profile: mockProfile,
        isLoading: true,
        error: 'Some error',
      })

      useUserStore.getState().reset()

      const state = useUserStore.getState()
      expect(state.profile).toBeNull()
      expect(state.isLoading).toBe(false)
      expect(state.error).toBeNull()
    })
  })
})
