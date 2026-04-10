/**
 * Translates a credential field ENV name (e.g. "ANTHROPIC_API_KEY") to a display label.
 * Uses i18n key `settings.agentCredentials.fields.{name}`, falls back to raw ENV name.
 */
export function getCredentialFieldLabel(
  fieldName: string,
  t: (key: string) => string
): string {
  const i18nKey = `settings.agentCredentials.fields.${fieldName}`;
  const translated = t(i18nKey);
  return translated !== i18nKey ? translated : fieldName;
}
