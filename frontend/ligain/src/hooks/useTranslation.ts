import { useTranslation as useI18nTranslation } from 'react-i18next';

export const useTranslation = () => {
  const { t, i18n } = useI18nTranslation();
  
  return {
    t,
    i18n,
    currentLanguage: i18n.language,
    isFrench: i18n.language === 'fr',
    isEnglish: i18n.language === 'en',
  };
}; 