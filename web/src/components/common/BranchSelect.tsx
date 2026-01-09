"use client";

import { useState, useEffect, useCallback } from "react";
import { repositoryApi, RepositoryData } from "@/lib/api";

export interface BranchSelectProps {
  repositoryId: number | null;
  repository?: RepositoryData;
  value: string;
  onChange: (value: string) => void;
  disabled?: boolean;
  placeholder?: string;
  className?: string;
  /** Allow manual input when branches are unavailable */
  allowManualInput?: boolean;
  /** Git access token for fetching branches (optional, will use manual mode if not provided) */
  accessToken?: string;
}

export function BranchSelect({
  repositoryId,
  repository,
  value,
  onChange,
  disabled = false,
  placeholder = "Select a branch...",
  className = "",
  allowManualInput = true,
  accessToken,
}: BranchSelectProps) {
  const [branches, setBranches] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [manualMode, setManualMode] = useState(false);

  const loadBranches = useCallback(async () => {
    if (!repositoryId) {
      setBranches([]);
      setManualMode(false);
      return;
    }

    // If no access token provided, use manual mode by default
    if (!accessToken) {
      setBranches([]);
      if (allowManualInput) {
        setManualMode(true);
      }
      return;
    }

    setLoading(true);
    setError(null);
    setManualMode(false);

    try {
      const res = await repositoryApi.listBranches(repositoryId, accessToken);
      const branchList = res.branches || [];
      setBranches(branchList);

      // If no branches available, switch to manual mode
      if (branchList.length === 0 && allowManualInput) {
        setManualMode(true);
      }

      // Auto-select default branch if available
      if (branchList.length > 0 && !value) {
        const defaultBranch = repository?.default_branch;
        if (defaultBranch && branchList.includes(defaultBranch)) {
          onChange(defaultBranch);
        } else {
          onChange(branchList[0]);
        }
      }
    } catch (err) {
      console.error("Failed to load branches:", err);
      setError("Failed to load branches");
      // Switch to manual mode on error if allowed
      if (allowManualInput) {
        setManualMode(true);
      }
    } finally {
      setLoading(false);
    }
  }, [repositoryId, repository, value, onChange, allowManualInput, accessToken]);

  useEffect(() => {
    loadBranches();
  }, [loadBranches]);

  const handleSelectChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const selectedValue = e.target.value;
    if (selectedValue === "__manual__") {
      setManualMode(true);
      onChange("");
    } else {
      onChange(selectedValue);
    }
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    onChange(e.target.value);
  };

  const handleBackToSelect = () => {
    setManualMode(false);
    if (branches.length > 0) {
      const defaultBranch = repository?.default_branch;
      if (defaultBranch && branches.includes(defaultBranch)) {
        onChange(defaultBranch);
      } else {
        onChange(branches[0]);
      }
    }
  };

  // No repository selected
  if (!repositoryId) {
    return (
      <select
        className={`w-full px-3 py-2 border border-border rounded-md bg-background ${className}`}
        disabled
      >
        <option value="">Select a repository first</option>
      </select>
    );
  }

  // Loading state
  if (loading) {
    return (
      <div className={`w-full px-3 py-2 border border-border rounded-md bg-muted text-muted-foreground ${className}`}>
        Loading branches...
      </div>
    );
  }

  // Manual input mode
  if (manualMode) {
    return (
      <div className={`flex gap-2 ${className}`}>
        <input
          type="text"
          className="flex-1 px-3 py-2 border border-border rounded-md bg-background"
          placeholder="Enter branch name (e.g., main)"
          value={value}
          onChange={handleInputChange}
          disabled={disabled}
        />
        {branches.length > 0 && (
          <button
            type="button"
            onClick={handleBackToSelect}
            className="px-3 py-2 text-sm text-muted-foreground hover:text-foreground border border-border rounded-md"
            disabled={disabled}
          >
            List
          </button>
        )}
      </div>
    );
  }

  // Error state with retry
  if (error && !manualMode) {
    return (
      <div className={`text-sm text-destructive ${className}`}>
        {error}
        <button
          type="button"
          onClick={loadBranches}
          className="ml-2 underline hover:no-underline"
        >
          Retry
        </button>
        {allowManualInput && (
          <button
            type="button"
            onClick={() => setManualMode(true)}
            className="ml-2 underline hover:no-underline"
          >
            Enter manually
          </button>
        )}
      </div>
    );
  }

  // Dropdown mode
  return (
    <select
      className={`w-full px-3 py-2 border border-border rounded-md bg-background ${className}`}
      value={value}
      onChange={handleSelectChange}
      disabled={disabled}
    >
      <option value="">{placeholder}</option>
      {branches.map((branch) => (
        <option key={branch} value={branch}>
          {branch}
          {branch === repository?.default_branch ? " (default)" : ""}
        </option>
      ))}
      {allowManualInput && (
        <option value="__manual__">Enter manually...</option>
      )}
    </select>
  );
}

export default BranchSelect;
