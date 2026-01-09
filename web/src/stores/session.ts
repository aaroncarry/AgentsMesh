import { create } from "zustand";
import { sessionApi, SessionData } from "@/lib/api/client";

// Re-export SessionData as Session for backward compatibility
export type Session = SessionData;

interface SessionState {
  // State
  sessions: Session[];
  currentSession: Session | null;
  loading: boolean;
  error: string | null;

  // Actions
  fetchSessions: (filters?: {
    status?: string;
    runnerId?: number;
  }) => Promise<void>;
  fetchSession: (sessionKey: string) => Promise<void>;
  createSession: (data: {
    runnerId: number;
    agentTypeId?: number;
    repositoryId?: number;
    ticketId?: number;
    initialPrompt?: string;
    branchName?: string;
  }) => Promise<Session>;
  terminateSession: (sessionKey: string) => Promise<void>;
  setCurrentSession: (session: Session | null) => void;
  updateSessionStatus: (sessionKey: string, status: Session["status"]) => void;
  updateAgentStatus: (sessionKey: string, agentStatus: string) => void;
  clearError: () => void;
}

export const useSessionStore = create<SessionState>((set, get) => ({
  sessions: [],
  currentSession: null,
  loading: false,
  error: null,

  fetchSessions: async (filters) => {
    set({ loading: true, error: null });
    try {
      const response = await sessionApi.list(filters);
      set({ sessions: response.sessions || [], loading: false });
    } catch (error: any) {
      set({
        error: error.message || "Failed to fetch sessions",
        loading: false,
      });
    }
  },

  fetchSession: async (sessionKey) => {
    set({ loading: true, error: null });
    try {
      const response = await sessionApi.get(sessionKey);
      set({ currentSession: response.session, loading: false });
    } catch (error: any) {
      set({
        error: error.message || "Failed to fetch session",
        loading: false,
      });
    }
  },

  createSession: async (data) => {
    set({ loading: true, error: null });
    try {
      // Convert camelCase to snake_case for API
      const apiData = {
        agent_type_id: data.agentTypeId ?? 0,
        runner_id: data.runnerId,
        repository_id: data.repositoryId,
        ticket_id: data.ticketId,
        initial_prompt: data.initialPrompt,
        branch_name: data.branchName,
      };
      const response = await sessionApi.create(apiData);
      set((state) => ({
        sessions: [response.session, ...state.sessions],
        currentSession: response.session,
        loading: false,
      }));
      return response.session;
    } catch (error: any) {
      set({
        error: error.message || "Failed to create session",
        loading: false,
      });
      throw error;
    }
  },

  terminateSession: async (sessionKey) => {
    try {
      await sessionApi.terminate(sessionKey);
      set((state) => ({
        sessions: state.sessions.map((s) =>
          s.session_key === sessionKey ? { ...s, status: "terminated" as const } : s
        ),
        currentSession:
          state.currentSession?.session_key === sessionKey
            ? { ...state.currentSession, status: "terminated" as const }
            : state.currentSession,
      }));
    } catch (error: any) {
      set({ error: error.message || "Failed to terminate session" });
      throw error;
    }
  },

  setCurrentSession: (session) => {
    set({ currentSession: session });
  },

  updateSessionStatus: (sessionKey, status) => {
    set((state) => ({
      sessions: state.sessions.map((s) =>
        s.session_key === sessionKey ? { ...s, status } : s
      ),
      currentSession:
        state.currentSession?.session_key === sessionKey
          ? { ...state.currentSession, status }
          : state.currentSession,
    }));
  },

  updateAgentStatus: (sessionKey, agentStatus) => {
    set((state) => ({
      sessions: state.sessions.map((s) =>
        s.session_key === sessionKey ? { ...s, agent_status: agentStatus } : s
      ),
      currentSession:
        state.currentSession?.session_key === sessionKey
          ? { ...state.currentSession, agent_status: agentStatus }
          : state.currentSession,
    }));
  },

  clearError: () => {
    set({ error: null });
  },
}));
