import { useState, useCallback, useMemo, useEffect } from "react";
import { podApi, PodData, AgentTypeData, RepositoryData } from "@/lib/api/client";

/**
 * Validation errors for the form
 */
export interface FormValidationErrors {
  agent?: string;
  runner?: string;
  repository?: string;
  branch?: string;
  prompt?: string;
}

export interface CreatePodFormState {
  // Selection state
  selectedAgent: number | null;
  selectedRunner: number | null;
  selectedRepository: number | null;
  selectedBranch: string;
  prompt: string;

  // Actions
  setSelectedAgent: (id: number | null) => void;
  setSelectedRunner: (id: number | null) => void;
  setSelectedRepository: (id: number | null) => void;
  setSelectedBranch: (branch: string) => void;
  setPrompt: (prompt: string) => void;

  // Computed
  selectedAgentSlug: string;

  // Form state
  loading: boolean;
  error: string | null;
  validationErrors: FormValidationErrors;
  isValid: boolean;

  // Actions
  reset: () => void;
  validate: () => boolean;
  submit: (pluginConfig: Record<string, unknown>) => Promise<PodData | null>;
}

/**
 * Hook to manage Create Pod form state and submission
 */
export function useCreatePodForm(
  agentTypes: AgentTypeData[],
  repositories: RepositoryData[],
  onSuccess?: (pod: PodData) => void
): CreatePodFormState {
  const [selectedAgent, setSelectedAgent] = useState<number | null>(null);
  const [selectedRunner, setSelectedRunner] = useState<number | null>(null);
  const [selectedRepository, setSelectedRepository] = useState<number | null>(null);
  const [selectedBranch, setSelectedBranch] = useState<string>("");
  const [prompt, setPrompt] = useState<string>("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [validationErrors, setValidationErrors] = useState<FormValidationErrors>({});

  // Compute agent slug from selected agent
  const selectedAgentSlug = useMemo(() => {
    if (!selectedAgent) return "";
    const agent = agentTypes.find((a) => a.id === selectedAgent);
    return agent?.slug || "";
  }, [selectedAgent, agentTypes]);

  // Compute form validity
  const isValid = useMemo(() => {
    return selectedAgent !== null && selectedRunner !== null;
  }, [selectedAgent, selectedRunner]);

  // Auto-select default branch when repository is selected
  useEffect(() => {
    if (!selectedRepository) {
      setSelectedBranch("");
      return;
    }

    const repo = repositories.find((r) => r.id === selectedRepository);
    if (repo?.default_branch) {
      setSelectedBranch(repo.default_branch);
    }
  }, [selectedRepository, repositories]);

  // Clear validation error when field changes
  useEffect(() => {
    if (selectedAgent && validationErrors.agent) {
      setValidationErrors((prev) => ({ ...prev, agent: undefined }));
    }
  }, [selectedAgent, validationErrors.agent]);

  useEffect(() => {
    if (selectedRunner && validationErrors.runner) {
      setValidationErrors((prev) => ({ ...prev, runner: undefined }));
    }
  }, [selectedRunner, validationErrors.runner]);

  // Validate form
  const validate = useCallback((): boolean => {
    const errors: FormValidationErrors = {};

    if (!selectedAgent) {
      errors.agent = "Please select an agent type";
    }

    if (!selectedRunner) {
      errors.runner = "Please select a runner";
    }

    // Branch validation: if repository is selected but branch is empty, warn
    if (selectedRepository && !selectedBranch.trim()) {
      errors.branch = "Branch name is recommended when using a repository";
    }

    // Validate branch name format (optional, only if provided)
    if (selectedBranch.trim()) {
      const branchRegex = /^[a-zA-Z0-9._/-]+$/;
      if (!branchRegex.test(selectedBranch)) {
        errors.branch = "Branch name contains invalid characters";
      }
    }

    setValidationErrors(errors);
    return Object.keys(errors).filter(k => errors[k as keyof FormValidationErrors]).length === 0;
  }, [selectedAgent, selectedRunner, selectedRepository, selectedBranch]);

  // Reset form
  const reset = useCallback(() => {
    setSelectedAgent(null);
    setSelectedRunner(null);
    setSelectedRepository(null);
    setSelectedBranch("");
    setPrompt("");
    setError(null);
    setValidationErrors({});
  }, []);

  // Submit form
  const submit = useCallback(
    async (pluginConfig: Record<string, unknown>): Promise<PodData | null> => {
      // Validate before submission
      if (!validate()) {
        return null;
      }

      if (!selectedAgent || !selectedRunner) {
        setError("Please select an agent and runner");
        return null;
      }

      setLoading(true);
      setError(null);

      try {
        // Build plugin config for API
        // Keep the full key format (plugin.field) to avoid name collisions
        // between plugins that have fields with the same name
        const config: Record<string, unknown> = {
          agent_type: selectedAgentSlug,
          // Spread plugin config directly - keys are already namespaced as "plugin.field"
          ...pluginConfig,
        };

        const response = await podApi.create({
          agent_type_id: selectedAgent,
          runner_id: selectedRunner,
          repository_id: selectedRepository || undefined,
          branch_name: selectedBranch || undefined,
          initial_prompt: prompt,
          plugin_config: config,
        });

        if (response.pod) {
          onSuccess?.(response.pod);
          return response.pod;
        }
        return null;
      } catch (err) {
        const message = err instanceof Error ? err.message : "Failed to create pod";
        setError(message);
        console.error("Failed to create pod:", err);
        return null;
      } finally {
        setLoading(false);
      }
    },
    [selectedAgent, selectedRunner, selectedRepository, selectedBranch, prompt, selectedAgentSlug, onSuccess, validate]
  );

  return {
    selectedAgent,
    selectedRunner,
    selectedRepository,
    selectedBranch,
    prompt,
    setSelectedAgent,
    setSelectedRunner,
    setSelectedRepository,
    setSelectedBranch,
    setPrompt,
    selectedAgentSlug,
    loading,
    error,
    validationErrors,
    isValid,
    reset,
    validate,
    submit,
  };
}
