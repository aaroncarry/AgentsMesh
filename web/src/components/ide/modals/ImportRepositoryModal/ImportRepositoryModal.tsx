"use client";

import { Button } from "@/components/ui/button";
import { useTranslations } from "@/lib/i18n/client";
import { useImportWizard } from "./useImportWizard";
import { SourceStep, BrowseStep, ManualStep, ConfirmStep } from "./steps";
import type { ImportRepositoryModalProps } from "./types";

/**
 * ImportRepositoryModal - Modal for importing repositories from git providers
 *
 * Refactored with step components following SRP:
 * - SourceStep: Select provider or manual entry
 * - BrowseStep: Browse and search repositories from provider
 * - ManualStep: Enter repository details manually
 * - ConfirmStep: Review and confirm import
 */
export function ImportRepositoryModal({
  open,
  onClose,
  onImported,
  existingRepositories = [],
}: ImportRepositoryModalProps) {
  const t = useTranslations();

  const [state, actions] = useImportWizard({
    open,
    onClose,
    onImported,
    existingRepositories,
    t,
  });

  if (!open) return null;

  const stepProps = { state, actions, existingRepositories, t };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-background rounded-lg shadow-lg w-full max-w-2xl mx-4 max-h-[80vh] flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-border">
          <h2 className="text-lg font-semibold">{t("repositories.modal.title")}</h2>
          <button onClick={onClose} className="text-muted-foreground hover:text-foreground">
            <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-auto p-4">
          {state.error && (
            <div className="mb-4 p-3 bg-destructive/10 text-destructive text-sm rounded-lg">
              {state.error}
            </div>
          )}

          {state.step === "source" && <SourceStep {...stepProps} />}
          {state.step === "browse" && <BrowseStep {...stepProps} />}
          {state.step === "manual" && <ManualStep {...stepProps} />}
          {state.step === "confirm" && <ConfirmStep {...stepProps} />}
        </div>

        {/* Footer */}
        <div className="flex justify-end gap-3 p-4 border-t border-border">
          <Button variant="outline" onClick={onClose}>
            {t("common.cancel")}
          </Button>
          {state.step === "manual" && (
            <Button onClick={actions.handleManualContinue}>
              {t("repositories.modal.continue")}
            </Button>
          )}
          {state.step === "confirm" && (
            <Button onClick={actions.handleImport} disabled={state.importing}>
              {state.importing ? "..." : t("repositories.modal.importRepository")}
            </Button>
          )}
        </div>
      </div>
    </div>
  );
}

export default ImportRepositoryModal;
