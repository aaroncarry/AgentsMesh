import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { CredentialFormFields } from "../CredentialFormFields";
import type { CredentialField } from "@/lib/api";

const mockT = (key: string) => {
  const translations: Record<string, string> = {
    "common.optional": "Optional",
    "settings.agentCredentials.secretPlaceholder": "Leave empty to keep existing",
    "settings.agentCredentials.secretEditHint": "Leave empty to keep the existing value",
    "settings.agentCredentials.fields.ANTHROPIC_API_KEY": "API Key",
    "settings.agentCredentials.fields.ANTHROPIC_BASE_URL": "Base URL",
  };
  return translations[key] || key;
};

const claudeFields: CredentialField[] = [
  { name: "ANTHROPIC_API_KEY", type: "secret", optional: true },
  { name: "ANTHROPIC_AUTH_TOKEN", type: "secret", optional: true },
  { name: "ANTHROPIC_BASE_URL", type: "text", optional: true },
];

describe("CredentialFormFields", () => {
  it("renders fields dynamically from credentialFields", () => {
    render(
      <CredentialFormFields
        credentialFields={claudeFields}
        fieldValues={{}}
        onFieldChange={vi.fn()}
        isEditing={false}
        t={mockT}
      />
    );

    // i18n translated labels
    expect(screen.getByLabelText(/API Key/)).toBeInTheDocument();
    expect(screen.getByLabelText(/Base URL/)).toBeInTheDocument();
    // Fallback to ENV name when no translation
    expect(screen.getByLabelText(/ANTHROPIC_AUTH_TOKEN/)).toBeInTheDocument();
  });

  it("renders secret fields as password inputs", () => {
    render(
      <CredentialFormFields
        credentialFields={claudeFields}
        fieldValues={{}}
        onFieldChange={vi.fn()}
        isEditing={false}
        t={mockT}
      />
    );

    const apiKeyInput = screen.getByLabelText(/API Key/);
    expect(apiKeyInput).toHaveAttribute("type", "password");

    const baseUrlInput = screen.getByLabelText(/Base URL/);
    expect(baseUrlInput).toHaveAttribute("type", "text");
  });

  it("shows optional badge for optional fields", () => {
    render(
      <CredentialFormFields
        credentialFields={claudeFields}
        fieldValues={{}}
        onFieldChange={vi.fn()}
        isEditing={false}
        t={mockT}
      />
    );

    const optionalBadges = screen.getAllByText("(Optional)");
    expect(optionalBadges).toHaveLength(3);
  });

  it("shows edit hints for secret fields in editing mode", () => {
    render(
      <CredentialFormFields
        credentialFields={claudeFields}
        fieldValues={{}}
        onFieldChange={vi.fn()}
        isEditing={true}
        t={mockT}
      />
    );

    // Secret fields should have edit hints
    const hints = screen.getAllByText("Leave empty to keep the existing value");
    expect(hints).toHaveLength(2); // API_KEY + AUTH_TOKEN, not BASE_URL (text type)
  });

  it("calls onFieldChange with ENV name as key", () => {
    const onFieldChange = vi.fn();
    render(
      <CredentialFormFields
        credentialFields={claudeFields}
        fieldValues={{}}
        onFieldChange={onFieldChange}
        isEditing={false}
        t={mockT}
      />
    );

    const apiKeyInput = screen.getByLabelText(/API Key/);
    fireEvent.change(apiKeyInput, { target: { value: "sk-test-123" } });

    expect(onFieldChange).toHaveBeenCalledWith("ANTHROPIC_API_KEY", "sk-test-123");
  });

  it("renders nothing for empty credentialFields", () => {
    const { container } = render(
      <CredentialFormFields
        credentialFields={[]}
        fieldValues={{}}
        onFieldChange={vi.fn()}
        isEditing={false}
        t={mockT}
      />
    );

    expect(container.innerHTML).toBe("");
  });

  it("populates field values for text fields in editing mode", () => {
    render(
      <CredentialFormFields
        credentialFields={claudeFields}
        fieldValues={{ ANTHROPIC_BASE_URL: "https://api.example.com" }}
        onFieldChange={vi.fn()}
        isEditing={true}
        t={mockT}
      />
    );

    const baseUrlInput = screen.getByLabelText(/Base URL/);
    expect(baseUrlInput).toHaveValue("https://api.example.com");
  });
});
