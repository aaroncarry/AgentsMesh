import { cookies, headers } from "next/headers";
import { Locale, defaultLocale, isValidLocale, LOCALE_COOKIE, getLocaleFromHeaders } from "./config";

// Type for nested translation object
type TranslationValue = string | { [key: string]: TranslationValue };
type Translations = { [key: string]: TranslationValue };

// Get locale from request
export async function getLocale(): Promise<Locale> {
  // Try to get from cookie first
  const cookieStore = await cookies();
  const localeCookie = cookieStore.get(LOCALE_COOKIE);
  if (localeCookie && isValidLocale(localeCookie.value)) {
    return localeCookie.value;
  }

  // Fall back to Accept-Language header
  const headersList = await headers();
  const acceptLanguage = headersList.get("accept-language");
  return getLocaleFromHeaders(acceptLanguage);
}

// Load translations for a locale
export async function getTranslations(locale?: Locale): Promise<Translations> {
  const effectiveLocale = locale ?? (await getLocale());

  try {
    const translations = await import(`@/messages/${effectiveLocale}.json`);
    return translations.default || translations;
  } catch {
    console.error(`Failed to load translations for ${effectiveLocale}`);
    // Fall back to default locale
    if (effectiveLocale !== defaultLocale) {
      try {
        const fallback = await import(`@/messages/${defaultLocale}.json`);
        return fallback.default || fallback;
      } catch {
        return {};
      }
    }
    return {};
  }
}

// Get nested value from object using dot notation
function getNestedValue(obj: Translations, path: string): string | undefined {
  const keys = path.split(".");
  let current: TranslationValue | undefined = obj;

  for (const key of keys) {
    if (current && typeof current === "object" && key in current) {
      current = current[key];
    } else {
      return undefined;
    }
  }

  return typeof current === "string" ? current : undefined;
}

// Replace parameters in translation string
function interpolate(
  str: string,
  params?: Record<string, string | number>
): string {
  if (!params) return str;

  return str.replace(/\{(\w+)\}/g, (_, key) => {
    return params[key]?.toString() ?? `{${key}}`;
  });
}

// Create a translation function for server components
export async function createTranslator(locale?: Locale) {
  const translations = await getTranslations(locale);

  return function t(key: string, params?: Record<string, string | number>): string {
    const value = getNestedValue(translations, key);
    if (value === undefined) {
      console.warn(`Translation missing for key: ${key}`);
      return key;
    }
    return interpolate(value, params);
  };
}

// Helper to get a specific translation
export async function t(
  key: string,
  params?: Record<string, string | number>,
  locale?: Locale
): Promise<string> {
  const translator = await createTranslator(locale);
  return translator(key, params);
}
