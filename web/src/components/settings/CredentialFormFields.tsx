"use client";

import { useCallback } from "react";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import type { CredentialField } from "@/lib/api";
import { getCredentialFieldLabel } from "./credentialFieldLabel";

interface CredentialFormFieldsProps {
  credentialFields: CredentialField[];
  fieldValues: Record<string, string>;
  onFieldChange: (fieldName: string, value: string) => void;
  isEditing: boolean;
  t: (key: string) => string;
}

/**
 * CredentialFormFields - Dynamic form fields for credential profiles.
 *
 * Renders inputs based on AgentFile ENV SECRET/TEXT declarations.
 * Secret fields render as password inputs and are never echoed on edit.
 * Text fields render as plain inputs and are echoed on edit.
 */
export function CredentialFormFields({
  credentialFields,
  fieldValues,
  onFieldChange,
  isEditing,
  t,
}: CredentialFormFieldsProps) {
  const handleChange = useCallback(
    (fieldName: string, value: string) => onFieldChange(fieldName, value),
    [onFieldChange]
  );

  if (credentialFields.length === 0) return null;

  return (
    <>
      {credentialFields.map((field) => (
        <div key={field.name} className="grid gap-2">
          <Label htmlFor={`cred-${field.name}`}>
            {getCredentialFieldLabel(field.name, t)}
            {field.optional && (
              <span className="text-xs text-muted-foreground ml-1">
                ({t("common.optional")})
              </span>
            )}
          </Label>
          <Input
            id={`cred-${field.name}`}
            type={field.type === "secret" ? "password" : "text"}
            value={fieldValues[field.name] || ""}
            onChange={(e) => handleChange(field.name, e.target.value)}
            placeholder={
              isEditing && field.type === "secret"
                ? t("settings.agentCredentials.secretPlaceholder")
                : ""
            }
          />
          {isEditing && field.type === "secret" && (
            <p className="text-xs text-muted-foreground">
              {t("settings.agentCredentials.secretEditHint")}
            </p>
          )}
        </div>
      ))}
    </>
  );
}

export default CredentialFormFields;
