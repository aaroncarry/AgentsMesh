"use client";

import { useState, useCallback, useEffect } from "react";
import {
  repositoryApi,
  userRepositoryProviderApi,
  RepositoryProviderData,
  UserRemoteRepositoryData,
  RepositoryData,
} from "@/lib/api";
import type { ImportWizardState, ImportWizardActions, ImportWizardStep } from "./types";

const initialState: ImportWizardState = {
  step: "source",
  providers: [],
  selectedProvider: null,
  repositories: [],
  selectedRepo: null,
  search: "",
  page: 1,
  loadingProviders: true,
  loadingRepos: false,
  importing: false,
  error: null,
  manualProviderType: "github",
  manualBaseURL: "https://github.com",
  manualCloneURL: "",
  manualName: "",
  manualFullPath: "",
  manualDefaultBranch: "main",
  ticketPrefix: "",
  visibility: "organization",
};

interface UseImportWizardOptions {
  open: boolean;
  onClose: () => void;
  onImported?: () => void;
  existingRepositories?: RepositoryData[];
  t: (key: string) => string;
}

/**
 * Hook for managing import repository wizard state and actions
 */
export function useImportWizard({
  open,
  onClose,
  onImported,
  existingRepositories = [],
  t,
}: UseImportWizardOptions): [ImportWizardState, ImportWizardActions] {
  const [state, setState] = useState<ImportWizardState>(initialState);

  // Load providers
  const loadProviders = useCallback(async () => {
    try {
      setState(s => ({ ...s, loadingProviders: true }));
      const response = await userRepositoryProviderApi.list();
      const activeProviders = (response.providers || []).filter(
        (p) => p.is_active && (p.has_identity || p.has_bot_token)
      );
      setState(s => ({ ...s, providers: activeProviders, loadingProviders: false }));
    } catch (err) {
      console.error("Failed to load providers:", err);
      setState(s => ({
        ...s,
        error: t("repositories.modal.failedToLoadConnections"),
        loadingProviders: false,
      }));
    }
  }, [t]);

  // Load repositories for selected provider
  const loadRepositories = useCallback(async () => {
    if (!state.selectedProvider) return;
    try {
      setState(s => ({ ...s, loadingRepos: true, error: null }));
      const response = await userRepositoryProviderApi.listRepositories(state.selectedProvider.id, {
        page: state.page,
        perPage: 20,
        search: state.search || undefined,
      });
      setState(s => ({
        ...s,
        repositories: response.repositories || [],
        loadingRepos: false,
      }));
    } catch (err) {
      console.error("Failed to load repositories:", err);
      setState(s => ({
        ...s,
        error: t("repositories.modal.failedToLoadRepos"),
        loadingRepos: false,
      }));
    }
  }, [state.selectedProvider, state.page, state.search, t]);

  // Load providers when modal opens
  useEffect(() => {
    if (open) {
      loadProviders();
    }
  }, [open, loadProviders]);

  // Load repositories when step changes to browse
  useEffect(() => {
    if (state.step === "browse" && state.selectedProvider) {
      loadRepositories();
    }
  }, [state.step, state.selectedProvider, loadRepositories]);

  // Reset state when modal closes
  useEffect(() => {
    if (!open) {
      setState(initialState);
    }
  }, [open]);

  // Actions
  const actions: ImportWizardActions = {
    setStep: (step: ImportWizardStep) => setState(s => ({ ...s, step })),
    setSearch: (search: string) => setState(s => ({ ...s, search })),
    setPage: (page) => setState(s => ({
      ...s,
      page: typeof page === "function" ? page(s.page) : page,
    })),
    setError: (error) => setState(s => ({ ...s, error })),

    selectProvider: (provider: RepositoryProviderData) => {
      setState(s => ({ ...s, selectedProvider: provider, step: "browse" }));
    },

    clearProvider: () => {
      setState(s => ({
        ...s,
        selectedProvider: null,
        repositories: [],
        step: "source",
      }));
    },

    selectRepo: (repo: UserRemoteRepositoryData, existingRepos: RepositoryData[]) => {
      const existingRepo = existingRepos.find(
        (r) => r.clone_url === repo.clone_url || r.full_path === repo.full_path
      );
      setState(s => ({
        ...s,
        selectedRepo: repo,
        manualName: repo.name,
        manualFullPath: repo.full_path,
        manualDefaultBranch: repo.default_branch || "main",
        manualCloneURL: repo.clone_url,
        manualProviderType: s.selectedProvider?.provider_type || "github",
        manualBaseURL: s.selectedProvider?.base_url || "https://github.com",
        ticketPrefix: existingRepo?.ticket_prefix || "",
        step: "confirm",
      }));
    },

    setManualProviderType: (type: string) => {
      let baseURL = "";
      switch (type) {
        case "github":
          baseURL = "https://github.com";
          break;
        case "gitlab":
          baseURL = "https://gitlab.com";
          break;
        case "gitee":
          baseURL = "https://gitee.com";
          break;
        default:
          baseURL = "";
      }
      setState(s => ({ ...s, manualProviderType: type, manualBaseURL: baseURL }));
    },
    setManualBaseURL: (url: string) => setState(s => ({ ...s, manualBaseURL: url })),
    setManualCloneURL: (url: string) => setState(s => ({ ...s, manualCloneURL: url })),
    setManualName: (name: string) => setState(s => ({ ...s, manualName: name })),
    setManualFullPath: (path: string) => setState(s => ({ ...s, manualFullPath: path })),
    setManualDefaultBranch: (branch: string) => setState(s => ({ ...s, manualDefaultBranch: branch })),

    setTicketPrefix: (prefix: string) => setState(s => ({ ...s, ticketPrefix: prefix.toUpperCase() })),
    setVisibility: (visibility: string) => setState(s => ({ ...s, visibility })),

    loadProviders,
    loadRepositories,

    handleManualContinue: () => {
      if (!state.manualCloneURL || !state.manualName || !state.manualFullPath) {
        setState(s => ({ ...s, error: t("repositories.modal.fillRequiredFields") }));
        return false;
      }
      setState(s => ({ ...s, step: "confirm" }));
      return true;
    },

    handleImport: async () => {
      setState(s => ({ ...s, importing: true, error: null }));
      try {
        await repositoryApi.create({
          provider_type: state.manualProviderType,
          provider_base_url: state.manualBaseURL,
          clone_url: state.manualCloneURL,
          external_id: state.selectedRepo?.id || state.manualFullPath.replace(/[^a-zA-Z0-9]/g, "-"),
          name: state.manualName,
          full_path: state.manualFullPath,
          default_branch: state.manualDefaultBranch || "main",
          ticket_prefix: state.ticketPrefix || undefined,
          visibility: state.visibility,
        });
        onImported?.();
        onClose();
      } catch (err) {
        console.error("Failed to import repository:", err);
        setState(s => ({
          ...s,
          error: t("repositories.modal.failedToImport"),
          importing: false,
        }));
      }
    },

    goBack: () => {
      setState(s => {
        if (s.step === "browse") {
          return { ...s, step: "source", selectedProvider: null, repositories: [] };
        }
        if (s.step === "manual") {
          return { ...s, step: "source" };
        }
        if (s.step === "confirm") {
          return { ...s, step: s.selectedRepo ? "browse" : "manual" };
        }
        return s;
      });
    },

    reset: () => setState(initialState),
  };

  return [state, actions];
}

export default useImportWizard;
