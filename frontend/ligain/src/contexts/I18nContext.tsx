import React, { createContext, useContext } from 'react';
import { useTranslation as useI18nTranslation } from 'react-i18next';

interface I18nContextType {
  isReady: boolean;
  currentLanguage: string;
  isFrench: boolean;
  isEnglish: boolean;
}

const I18nContext = createContext<I18nContextType>({
  isReady: true,
  currentLanguage: 'en',
  isFrench: false,
  isEnglish: true,
});

export const useI18n = () => {
  const context = useContext(I18nContext);
  if (!context) {
    throw new Error('useI18n must be used within an I18nProvider');
  }
  return context;
};

export const I18nProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { i18n } = useI18nTranslation();

  const value: I18nContextType = {
    isReady: i18n.isInitialized,
    currentLanguage: i18n.language,
    isFrench: i18n.language === 'fr',
    isEnglish: i18n.language === 'en',
  };

  return (
    <I18nContext.Provider value={value}>
      {children}
    </I18nContext.Provider>
  );
}; 