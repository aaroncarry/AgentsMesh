// Configuration
export {
  locales,
  defaultLocale,
  localeNames,
  LOCALE_COOKIE,
  isValidLocale,
  type Locale,
} from "./config";

// Client-side hooks and provider
export {
  I18nProvider,
  useI18n,
  useTranslations,
  useLocale,
} from "./client";

// Server-side utilities
export {
  getLocale,
  getTranslations,
  createTranslator,
  t,
} from "./server";
