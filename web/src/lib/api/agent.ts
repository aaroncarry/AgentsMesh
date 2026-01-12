import { request, orgPath } from "./base";

// Agent type interface
export interface AgentTypeData {
  id: number;
  slug: string;
  name: string;
  description?: string;
  launch_command?: string;
  is_builtin: boolean;
  is_active: boolean;
}

// Organization agent default config interface
export interface OrganizationAgentConfigData {
  id: number;
  organization_id: number;
  agent_type_id: number;
  agent_type_name?: string;
  agent_type_slug?: string;
  config_values: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

// Agents API
export const agentApi = {
  listTypes: async () => {
    const response = await request<{
      builtin_types: AgentTypeData[];
      custom_types: AgentTypeData[];
    }>(orgPath("/agents/types"));
    // Combine builtin and custom types for frontend compatibility
    return {
      agent_types: [...(response.builtin_types || []), ...(response.custom_types || [])],
    };
  },

  getConfig: () =>
    request<{ config: unknown }>(orgPath("/agents/config")),

  updateConfig: (data: unknown) =>
    request<{ message: string }>(orgPath("/agents/config"), {
      method: "PUT",
      body: data,
    }),

  listCredentials: () =>
    request<{ credentials: unknown[] }>(orgPath("/agents/credentials")),

  updateCredentials: (agentType: string, credentials: Record<string, string>) =>
    request<{ message: string }>(`${orgPath("/agents/credentials")}/${agentType}`, {
      method: "PUT",
      body: { credentials },
    }),

  // Organization default config API
  listDefaultConfigs: () =>
    request<{ configs: OrganizationAgentConfigData[] }>(orgPath("/agents/default-configs")),

  getDefaultConfig: (agentTypeId: number) =>
    request<{ config: OrganizationAgentConfigData }>(`${orgPath("/agents")}/${agentTypeId}/default-config`),

  setDefaultConfig: (agentTypeId: number, configValues: Record<string, unknown>) =>
    request<{ config: OrganizationAgentConfigData }>(`${orgPath("/agents")}/${agentTypeId}/default-config`, {
      method: "PUT",
      body: { config_values: configValues },
    }),

  deleteDefaultConfig: (agentTypeId: number) =>
    request<{ message: string }>(`${orgPath("/agents")}/${agentTypeId}/default-config`, {
      method: "DELETE",
    }),
};
