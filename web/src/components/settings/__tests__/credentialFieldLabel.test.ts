import { describe, it, expect } from "vitest";
import { getCredentialFieldLabel } from "../credentialFieldLabel";

describe("getCredentialFieldLabel", () => {
  it("returns translated label when i18n key exists", () => {
    const t = (key: string) =>
      key === "settings.agentCredentials.fields.ANTHROPIC_API_KEY" ? "API Key" : key;

    expect(getCredentialFieldLabel("ANTHROPIC_API_KEY", t)).toBe("API Key");
  });

  it("falls back to raw ENV name when no translation", () => {
    const t = (key: string) => key; // no translation — returns key as-is

    expect(getCredentialFieldLabel("SOME_UNKNOWN_KEY", t)).toBe("SOME_UNKNOWN_KEY");
  });
});
