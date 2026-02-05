"use client";

import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import type { CredentialDialogProps, CredentialFormData } from "./types";

/**
 * CredentialDialog - Dialog for adding or editing credential profiles
 *
 * Displays a form with name, description, base URL, and API key fields.
 * Handles both create and edit modes based on editingProfile prop.
 */
export function CredentialDialog({
  open,
  onOpenChange,
  editingProfile,
  onSubmit,
  t,
}: CredentialDialogProps) {
  const [formName, setFormName] = useState("");
  const [formDescription, setFormDescription] = useState("");
  const [formBaseUrl, setFormBaseUrl] = useState("");
  const [formApiKey, setFormApiKey] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Reset form when dialog opens/closes or editing profile changes
  useEffect(() => {
    if (open) {
      if (editingProfile) {
        setFormName(editingProfile.name);
        setFormDescription(editingProfile.description || "");
        setFormBaseUrl("");
        setFormApiKey("");
      } else {
        setFormName("");
        setFormDescription("");
        setFormBaseUrl("");
        setFormApiKey("");
      }
      setError(null);
    }
  }, [open, editingProfile]);

  const handleSubmit = async () => {
    if (!formName.trim()) return;

    try {
      setSubmitting(true);
      setError(null);

      const formData: CredentialFormData = {
        name: formName,
        description: formDescription,
        baseUrl: formBaseUrl,
        apiKey: formApiKey,
      };

      await onSubmit(formData, editingProfile);
      onOpenChange(false);
    } catch (err) {
      console.error("Failed to save profile:", err);
      setError(t("settings.agentCredentials.failedToSave"));
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>
            {editingProfile
              ? t("settings.agentCredentials.editProfile")
              : t("settings.agentCredentials.addProfile")}
          </DialogTitle>
          <DialogDescription>
            {t("settings.agentCredentials.customProfileDescription")}
          </DialogDescription>
        </DialogHeader>

        <div className="grid gap-4 py-4">
          {error && (
            <div className="text-sm text-destructive">{error}</div>
          )}

          <div className="grid gap-2">
            <Label htmlFor="name">{t("settings.agentCredentials.name")}</Label>
            <Input
              id="name"
              value={formName}
              onChange={(e) => setFormName(e.target.value)}
              placeholder={t("settings.agentCredentials.namePlaceholder")}
            />
          </div>

          <div className="grid gap-2">
            <Label htmlFor="description">{t("settings.agentCredentials.descriptionLabel")}</Label>
            <Textarea
              id="description"
              value={formDescription}
              onChange={(e) => setFormDescription(e.target.value)}
              placeholder={t("settings.agentCredentials.descriptionPlaceholder")}
              rows={2}
            />
          </div>

          <div className="grid gap-2">
            <Label htmlFor="base_url">
              {t("settings.agentCredentials.baseUrl")}
              <span className="text-xs text-muted-foreground ml-1">
                ({t("common.optional")})
              </span>
            </Label>
            <Input
              id="base_url"
              value={formBaseUrl}
              onChange={(e) => setFormBaseUrl(e.target.value)}
              placeholder="https://api.anthropic.com"
            />
          </div>

          <div className="grid gap-2">
            <Label htmlFor="api_key">{t("settings.agentCredentials.apiKey")}</Label>
            <Input
              id="api_key"
              type="password"
              value={formApiKey}
              onChange={(e) => setFormApiKey(e.target.value)}
              placeholder={editingProfile ? t("settings.agentCredentials.apiKeyPlaceholder") : "sk-..."}
            />
            {editingProfile && (
              <p className="text-xs text-muted-foreground">
                {t("settings.agentCredentials.apiKeyEditHint")}
              </p>
            )}
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t("common.cancel")}
          </Button>
          <Button onClick={handleSubmit} disabled={submitting || !formName.trim()}>
            {submitting
              ? t("common.saving")
              : editingProfile
              ? t("common.save")
              : t("common.create")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

export default CredentialDialog;
