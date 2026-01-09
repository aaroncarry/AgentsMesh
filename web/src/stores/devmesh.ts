import { create } from "zustand";
import { devmeshApi } from "@/lib/api/client";

// DevMesh node representing a session in the topology
export interface DevMeshNode {
  session_key: string;
  status: string;
  agent_status: string;
  model?: string;
  ticket_id?: number;
  repository_id?: number;
  created_by_id: number;
  runner_id: number;
  started_at?: string;
  position?: {
    x: number;
    y: number;
  };
}

// DevMesh edge representing a binding between sessions
export interface DevMeshEdge {
  id: number;
  source: string;
  target: string;
  granted_scopes: string[];
  pending_scopes?: string[];
  status: string;
}

// Channel information for DevMesh visualization
export interface ChannelInfo {
  id: number;
  name: string;
  description?: string;
  session_keys: string[];
  message_count: number;
  is_archived: boolean;
}

// Complete topology data
export interface DevMeshTopology {
  nodes: DevMeshNode[];
  edges: DevMeshEdge[];
  channels: ChannelInfo[];
}

// Request to create a session for a ticket
export interface CreateSessionForTicketRequest {
  runner_id: number;
  initial_prompt?: string;
  model?: string;
  permission_mode?: string;
  think_level?: string;
}

interface DevMeshState {
  // State
  topology: DevMeshTopology | null;
  selectedNode: string | null;
  selectedChannel: number | null;
  loading: boolean;
  error: string | null;
  pollInterval: number | null;

  // Actions
  fetchTopology: () => Promise<void>;
  selectNode: (sessionKey: string | null) => void;
  selectChannel: (channelId: number | null) => void;
  startPolling: (interval?: number) => void;
  stopPolling: () => void;
  clearError: () => void;

  // Node helpers
  getNodeByKey: (sessionKey: string) => DevMeshNode | undefined;
  getEdgesForNode: (sessionKey: string) => DevMeshEdge[];
  getChannelsForNode: (sessionKey: string) => ChannelInfo[];
  getActiveNodes: () => DevMeshNode[];
}

export const useDevMeshStore = create<DevMeshState>((set, get) => ({
  topology: null,
  selectedNode: null,
  selectedChannel: null,
  loading: false,
  error: null,
  pollInterval: null,

  fetchTopology: async () => {
    set({ loading: true, error: null });
    try {
      const response = await devmeshApi.getTopology();
      set({ topology: response.topology, loading: false });
    } catch (error: unknown) {
      set({
        error: error instanceof Error ? error.message : "Failed to fetch topology",
        loading: false,
      });
    }
  },

  selectNode: (sessionKey) => {
    set({ selectedNode: sessionKey, selectedChannel: null });
  },

  selectChannel: (channelId) => {
    set({ selectedChannel: channelId, selectedNode: null });
  },

  startPolling: (interval = 5000) => {
    const state = get();

    // Clear existing interval if any
    if (state.pollInterval !== null) {
      clearInterval(state.pollInterval);
    }

    // Start new polling
    const pollId = window.setInterval(() => {
      get().fetchTopology();
    }, interval);

    set({ pollInterval: pollId as unknown as number });

    // Fetch immediately
    get().fetchTopology();
  },

  stopPolling: () => {
    const state = get();
    if (state.pollInterval !== null) {
      clearInterval(state.pollInterval);
      set({ pollInterval: null });
    }
  },

  clearError: () => {
    set({ error: null });
  },

  getNodeByKey: (sessionKey) => {
    const { topology } = get();
    return topology?.nodes.find((n) => n.session_key === sessionKey);
  },

  getEdgesForNode: (sessionKey) => {
    const { topology } = get();
    if (!topology) return [];
    return topology.edges.filter(
      (e) => e.source === sessionKey || e.target === sessionKey
    );
  },

  getChannelsForNode: (sessionKey) => {
    const { topology } = get();
    if (!topology) return [];
    return topology.channels.filter((c) =>
      c.session_keys.includes(sessionKey)
    );
  },

  getActiveNodes: () => {
    const { topology } = get();
    if (!topology) return [];
    return topology.nodes.filter(
      (n) => n.status === "running" || n.status === "initializing"
    );
  },
}));

// Helper function to get session status display info
export const getSessionStatusInfo = (status: string) => {
  const statusMap: Record<
    string,
    { label: string; color: string; bgColor: string }
  > = {
    initializing: {
      label: "Initializing",
      color: "text-blue-600",
      bgColor: "bg-blue-100",
    },
    running: {
      label: "Running",
      color: "text-green-600",
      bgColor: "bg-green-100",
    },
    paused: {
      label: "Paused",
      color: "text-yellow-600",
      bgColor: "bg-yellow-100",
    },
    terminated: {
      label: "Terminated",
      color: "text-gray-600",
      bgColor: "bg-gray-100",
    },
    failed: {
      label: "Failed",
      color: "text-red-600",
      bgColor: "bg-red-100",
    },
  };
  return statusMap[status] || statusMap.terminated;
};

// Helper function to get agent status display info
export const getAgentStatusInfo = (agentStatus: string) => {
  const statusMap: Record<
    string,
    { label: string; color: string; icon: string }
  > = {
    idle: { label: "Idle", color: "text-gray-500", icon: "⏸" },
    thinking: { label: "Thinking", color: "text-blue-500", icon: "🤔" },
    coding: { label: "Coding", color: "text-green-500", icon: "💻" },
    testing: { label: "Testing", color: "text-yellow-500", icon: "🧪" },
    reviewing: { label: "Reviewing", color: "text-purple-500", icon: "📝" },
    waiting: { label: "Waiting", color: "text-orange-500", icon: "⏳" },
    error: { label: "Error", color: "text-red-500", icon: "❌" },
  };
  return statusMap[agentStatus] || { label: agentStatus, color: "text-gray-500", icon: "•" };
};

// Helper function to get binding status display info
export const getBindingStatusInfo = (status: string) => {
  const statusMap: Record<
    string,
    { label: string; color: string }
  > = {
    active: { label: "Active", color: "stroke-green-500" },
    pending: { label: "Pending", color: "stroke-yellow-500" },
    revoked: { label: "Revoked", color: "stroke-red-500" },
    expired: { label: "Expired", color: "stroke-gray-500" },
  };
  return statusMap[status] || statusMap.active;
};
