import { ApiError } from "./base";

/**
 * Extract the error code from an API error
 */
export function getApiErrorCode(error: unknown): string | undefined {
  if (error instanceof ApiError) {
    return error.code;
  }
  return undefined;
}

/**
 * Extract a server-suggested replacement value from a VALIDATION_FAILED error.
 * Returns undefined when the error has no `suggestion` field.
 *
 * Backend convention: `apierr.RespondWithExtra` may attach
 * `{ field: string, suggestion: string }` next to `error` and `code`.
 */
export function getErrorSuggestion(error: unknown): string | undefined {
  if (!(error instanceof ApiError)) return undefined;
  const data = error.data as { suggestion?: unknown } | null | undefined;
  if (typeof data?.suggestion === "string" && data.suggestion.length > 0) {
    return data.suggestion;
  }
  return undefined;
}

/**
 * Check if an error matches a specific API error code
 */
export function isApiErrorCode(error: unknown, code: string): boolean {
  return getApiErrorCode(error) === code;
}

/**
 * Check if an error has a specific HTTP status
 */
export function isApiStatus(error: unknown, status: number): boolean {
  return error instanceof ApiError && error.status === status;
}

/**
 * Get a localized error message, falling back through: i18n code translation → server message → fallback
 *
 * @param error - The caught error
 * @param t - next-intl translation function (must have access to "apiErrors" namespace in common.json)
 * @param fallback - Fallback message if no translation or server message is available
 */
export function getLocalizedErrorMessage(
  error: unknown,
  t: (key: string, values?: Record<string, string>) => string,
  fallback: string
): string {
  if (error instanceof ApiError) {
    const code = error.code;
    if (code) {
      // Try to get i18n translation for this error code
      // next-intl returns the key path itself when translation is missing
      const key = `apiErrors.${code}`;
      try {
        const translated = t(key);
        // next-intl returns the key itself if not found
        if (translated && translated !== key) {
          return translated;
        }
      } catch {
        // Translation not found, fall through
      }
    }
    if (error.serverMessage) {
      return error.serverMessage;
    }
    // Don't expose internal "API Error: 400 Bad Request" to the user.
    return fallback;
  }
  if (error instanceof Error) {
    return error.message;
  }
  return fallback;
}
