import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import * as Localization from 'expo-localization';

// Import translation files
import en from './locales/en.json';
import fr from './locales/fr.json';

const resources = {
  en: {
    translation: en,
  },
  fr: {
    translation: fr,
  },
};

// Get device locale
const getDeviceLocale = () => {
  try {
    const locale = Localization.getLocales()[0]?.languageCode || 'en';
    console.log('🌍 Device locale detected:', locale);
    if (locale.startsWith('fr')) {
      console.log('🇫🇷 Using French translations');
      return 'fr';
    }
    console.log('🇺🇸 Using English translations');
    return 'en';
  } catch (error) {
    console.log('🌍 Error detecting locale, defaulting to English:', error);
    return 'en';
  }
};

// Initialize i18n immediately and synchronously
i18n
  .use(initReactI18next)
  .init({
    resources,
    lng: getDeviceLocale(),
    fallbackLng: 'en',
    interpolation: {
      escapeValue: false,
    },
    compatibilityJSON: 'v4',
    debug: __DEV__,
    initImmediate: false, // Force synchronous initialization
  });

console.log('🌍 i18n initialized with language:', i18n.language);
console.log('🌍 i18n isInitialized:', i18n.isInitialized);

// Utility function to get the current locale for date/time formatting
export const getCurrentLocale = (): string => {
  return i18n.language === 'fr' ? 'fr-FR' : 'en-US';
};

export default i18n; 