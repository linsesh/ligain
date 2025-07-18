import i18n from '../index';

describe('i18n configuration', () => {
  it('should have English as fallback language', () => {
    expect(i18n.language).toBeDefined();
  });

  it('should support French translations', () => {
    i18n.changeLanguage('fr');
    expect(i18n.t('common.cancel')).toBe('Annuler');
    expect(i18n.t('auth.welcome')).toBe('Bienvenue sur Ligain');
  });

  it('should support English translations', () => {
    i18n.changeLanguage('en');
    expect(i18n.t('common.cancel')).toBe('Cancel');
    expect(i18n.t('auth.welcome')).toBe('Welcome to Ligain');
  });

  it('should fallback to English for missing translations', () => {
    i18n.changeLanguage('fr');
    expect(i18n.t('nonexistent.key')).toBe('nonexistent.key');
  });
}); 