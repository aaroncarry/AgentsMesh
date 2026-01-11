import { useState, useEffect } from "react";
import {
  runnerApi,
  agentApi,
  repositoryApi,
  RunnerData,
  AgentTypeData,
  RepositoryData,
} from "@/lib/api/client";

export interface PodCreationData {
  runners: RunnerData[];
  agentTypes: AgentTypeData[];
  repositories: RepositoryData[];
  loading: boolean;
  error: string | null;
}

/**
 * Hook to load data required for pod creation (runners, agents, repositories)
 * Only loads when enabled is true (e.g., when modal is open)
 */
export function usePodCreationData(enabled: boolean): PodCreationData {
  const [runners, setRunners] = useState<RunnerData[]>([]);
  const [agentTypes, setAgentTypes] = useState<AgentTypeData[]>([]);
  const [repositories, setRepositories] = useState<RepositoryData[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!enabled) return;

    let cancelled = false;

    const loadData = async () => {
      setLoading(true);
      setError(null);
      try {
        const [runnersRes, agentsRes, reposRes] = await Promise.allSettled([
          runnerApi.list(),
          agentApi.listTypes(),
          repositoryApi.list(),
        ]);

        if (cancelled) return;

        if (runnersRes.status === "fulfilled") {
          // Only online runners
          setRunners((runnersRes.value.runners || []).filter(r => r.status === "online"));
        }
        if (agentsRes.status === "fulfilled") {
          setAgentTypes(agentsRes.value.agent_types || []);
        }
        if (reposRes.status === "fulfilled") {
          setRepositories(reposRes.value.repositories || []);
        }
      } catch (err) {
        if (cancelled) return;
        const message = err instanceof Error ? err.message : "Failed to load data";
        setError(message);
        console.error("Failed to load pod creation data:", err);
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    };

    loadData();

    return () => {
      cancelled = true;
    };
  }, [enabled]);

  return { runners, agentTypes, repositories, loading, error };
}
