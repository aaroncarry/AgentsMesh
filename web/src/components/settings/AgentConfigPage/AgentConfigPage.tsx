"use client";

import { useState, useCallback } from "react";
import { CenteredSpinner } from "@/components/ui/spinner";
import { AlertMessage } from "@/components/ui/alert-message";
import { useTranslations } from "@/lib/i18n/client";
import { Bot, AlertCircle } from "lucide-react";
import type { CredentialProfileData } from "@/lib/api";
import { useAgentConfig } from "./useAgentConfig";
import { CredentialsSection } from "./CredentialsSection";
import { RuntimeConfigSection } from "./RuntimeConfigSection";
import { CredentialDialog } from "./CredentialDialog";
import type { AgentConfigPageProps, CredentialFormData } from "./types";

/**
 * AgentConfigPage - Unified configuration page for a single agent type
 *
 * Combines credentials management and runtime configuration in one place.
 * Acts as the coordinator for the extracted sub-components.
 */
export function AgentConfigPage({ agentSlug }: AgentConfigPageProps) {
  const t = useTranslations();

  // Dialog state
  const [showCredentialDialog, setShowCredentialDialog] = useState(false);
  const [editingProfile, setEditingProfile] = useState<CredentialProfileData | null>(null);

  // Use the custom hook for data and actions
  const {
    loading,
    savingConfig,
    agentType,
    configFields,
    configValues,
    credentialProfiles,
    isRunnerHostDefault,
    error,
    success,
    handleConfigChange,
    handleSaveConfig,
    handleSetRunnerHostDefault,
    handleSetDefault,
    handleDeleteProfile,
    handleSaveProfile,
    setError,
    setSuccess,
  } = useAgentConfig(agentSlug, t);

  // Open credential add dialog
  const handleOpenAddDialog = useCallback(() => {
    setEditingProfile(null);
    setShowCredentialDialog(true);
  }, []);

  // Open credential edit dialog
  const handleOpenEditDialog = useCallback((profile: CredentialProfileData) => {
    setEditingProfile(profile);
    setShowCredentialDialog(true);
  }, []);

  // Handle credential form submission
  const handleCredentialSubmit = useCallback(async (
    data: CredentialFormData,
    profile: CredentialProfileData | null
  ) => {
    await handleSaveProfile(data, profile);
    setShowCredentialDialog(false);
  }, [handleSaveProfile]);

  if (loading) {
    return <CenteredSpinner className="py-12" />;
  }

  if (!agentType) {
    return (
      <div className="flex flex-col items-center justify-center py-12">
        <AlertCircle className="w-12 h-12 text-muted-foreground mb-4" />
        <p className="text-muted-foreground">{error || t("settings.agentConfig.agentNotFound")}</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-3">
        <Bot className="w-8 h-8 text-primary" />
        <div>
          <h2 className="text-xl font-semibold">{agentType.name}</h2>
          {agentType.description && (
            <p className="text-sm text-muted-foreground">{agentType.description}</p>
          )}
        </div>
      </div>

      {/* Error/Success messages */}
      {error && <AlertMessage type="error" message={error} onDismiss={() => setError(null)} />}
      {success && <AlertMessage type="success" message={success} onDismiss={() => setSuccess(null)} />}

      {/* Credentials Section */}
      <CredentialsSection
        isRunnerHostDefault={isRunnerHostDefault}
        credentialProfiles={credentialProfiles}
        onSetRunnerHostDefault={handleSetRunnerHostDefault}
        onSetDefault={handleSetDefault}
        onEdit={handleOpenEditDialog}
        onDelete={handleDeleteProfile}
        onAdd={handleOpenAddDialog}
        t={t}
      />

      {/* Runtime Config Section */}
      <RuntimeConfigSection
        configFields={configFields}
        configValues={configValues}
        agentSlug={agentSlug}
        saving={savingConfig}
        onChange={handleConfigChange}
        onSave={handleSaveConfig}
        t={t}
      />

      {/* Add/Edit Credential Dialog */}
      <CredentialDialog
        open={showCredentialDialog}
        onOpenChange={setShowCredentialDialog}
        editingProfile={editingProfile}
        onSubmit={handleCredentialSubmit}
        t={t}
      />
    </div>
  );
}

export default AgentConfigPage;
