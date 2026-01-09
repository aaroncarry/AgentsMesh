import { create } from "zustand";

export interface CredentialField {
  name: string;
  type: "secret" | "text";
  env_var: string;
  required: boolean;
}

export interface AgentType {
  id: number;
  slug: string;
  name: string;
  description?: string;
  launch_command: string;
  default_args?: string;
  credential_schema: CredentialField[];
  is_builtin: boolean;
  is_active: boolean;
}

export interface CustomAgentType extends Omit<AgentType, "is_builtin"> {
  organization_id: number;
  status_detection?: Record<string, unknown>;
}

export interface OrganizationAgent {
  id: number;
  organization_id: number;
  agent_type_id: number;
  agent_type: AgentType;
  is_enabled: boolean;
  is_default: boolean;
  has_credentials: boolean;
}

export interface UserAgentCredentials {
  agent_type_id: number;
  agent_slug: string;
  has_credentials: boolean;
}

interface AgentState {
  builtinAgentTypes: AgentType[];
  customAgentTypes: CustomAgentType[];
  organizationAgents: OrganizationAgent[];
  userCredentials: UserAgentCredentials[];
  isLoading: boolean;
  error: string | null;

  // Actions
  setBuiltinAgentTypes: (types: AgentType[]) => void;
  setCustomAgentTypes: (types: CustomAgentType[]) => void;
  addCustomAgentType: (type: CustomAgentType) => void;
  updateCustomAgentType: (id: number, updates: Partial<CustomAgentType>) => void;
  removeCustomAgentType: (id: number) => void;
  setOrganizationAgents: (agents: OrganizationAgent[]) => void;
  enableAgent: (agentTypeId: number, isDefault: boolean) => void;
  disableAgent: (agentTypeId: number) => void;
  setUserCredentials: (credentials: UserAgentCredentials[]) => void;
  updateUserCredential: (agentTypeId: number, hasCredentials: boolean) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  reset: () => void;
}

const initialState = {
  builtinAgentTypes: [],
  customAgentTypes: [],
  organizationAgents: [],
  userCredentials: [],
  isLoading: false,
  error: null,
};

export const useAgentStore = create<AgentState>((set) => ({
  ...initialState,

  setBuiltinAgentTypes: (builtinAgentTypes) => set({ builtinAgentTypes }),

  setCustomAgentTypes: (customAgentTypes) => set({ customAgentTypes }),

  addCustomAgentType: (type) =>
    set((state) => ({
      customAgentTypes: [...state.customAgentTypes, type],
    })),

  updateCustomAgentType: (id, updates) =>
    set((state) => ({
      customAgentTypes: state.customAgentTypes.map((t) =>
        t.id === id ? { ...t, ...updates } : t
      ),
    })),

  removeCustomAgentType: (id) =>
    set((state) => ({
      customAgentTypes: state.customAgentTypes.filter((t) => t.id !== id),
    })),

  setOrganizationAgents: (organizationAgents) => set({ organizationAgents }),

  enableAgent: (agentTypeId, isDefault) =>
    set((state) => {
      const existing = state.organizationAgents.find(
        (a) => a.agent_type_id === agentTypeId
      );
      if (existing) {
        return {
          organizationAgents: state.organizationAgents.map((a) =>
            a.agent_type_id === agentTypeId
              ? { ...a, is_enabled: true, is_default: isDefault }
              : isDefault
              ? { ...a, is_default: false }
              : a
          ),
        };
      }
      return state;
    }),

  disableAgent: (agentTypeId) =>
    set((state) => ({
      organizationAgents: state.organizationAgents.map((a) =>
        a.agent_type_id === agentTypeId ? { ...a, is_enabled: false } : a
      ),
    })),

  setUserCredentials: (userCredentials) => set({ userCredentials }),

  updateUserCredential: (agentTypeId, hasCredentials) =>
    set((state) => ({
      userCredentials: state.userCredentials.map((c) =>
        c.agent_type_id === agentTypeId
          ? { ...c, has_credentials: hasCredentials }
          : c
      ),
    })),

  setLoading: (isLoading) => set({ isLoading }),

  setError: (error) => set({ error }),

  reset: () => set(initialState),
}));
