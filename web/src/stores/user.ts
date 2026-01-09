import { create } from "zustand";

export interface UserIdentity {
  id: number;
  provider: "github" | "google" | "gitlab" | "gitee";
  provider_user_id: string;
  provider_username?: string;
  created_at: string;
}

export interface User {
  id: number;
  email: string;
  username: string;
  name?: string;
  avatar_url?: string;
  is_active: boolean;
  last_login_at?: string;
  created_at: string;
  updated_at: string;
}

export interface UserProfile extends User {
  identities: UserIdentity[];
  organizations: Array<{
    id: number;
    name: string;
    slug: string;
    role: string;
  }>;
}

interface UserState {
  profile: UserProfile | null;
  isLoading: boolean;
  error: string | null;

  // Actions
  setProfile: (profile: UserProfile | null) => void;
  updateProfile: (updates: Partial<UserProfile>) => void;
  addIdentity: (identity: UserIdentity) => void;
  removeIdentity: (provider: string) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  reset: () => void;
}

const initialState = {
  profile: null,
  isLoading: false,
  error: null,
};

export const useUserStore = create<UserState>((set) => ({
  ...initialState,

  setProfile: (profile) => set({ profile }),

  updateProfile: (updates) =>
    set((state) => ({
      profile: state.profile ? { ...state.profile, ...updates } : null,
    })),

  addIdentity: (identity) =>
    set((state) => ({
      profile: state.profile
        ? {
            ...state.profile,
            identities: [...state.profile.identities, identity],
          }
        : null,
    })),

  removeIdentity: (provider) =>
    set((state) => ({
      profile: state.profile
        ? {
            ...state.profile,
            identities: state.profile.identities.filter(
              (i) => i.provider !== provider
            ),
          }
        : null,
    })),

  setLoading: (isLoading) => set({ isLoading }),

  setError: (error) => set({ error }),

  reset: () => set(initialState),
}));
