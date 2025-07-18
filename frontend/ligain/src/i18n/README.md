# Internationalization (i18n) Setup

This project uses `react-i18next` for internationalization, supporting French and English languages based on the user's device locale.

## How it works

1. **Automatic Language Detection**: The app automatically detects the user's device language using `expo-localization`
2. **French Support**: If the device locale starts with 'fr' (French), the app displays French translations
3. **English Fallback**: For all other locales, the app defaults to English
4. **Fallback System**: If a translation is missing, it falls back to English

## File Structure

```
src/i18n/
├── index.ts              # Main i18n configuration
├── locales/
│   ├── en.json          # English translations
│   └── fr.json          # French translations
├── __tests__/
│   └── i18n.test.ts     # Tests for i18n functionality
└── README.md            # This file
```

## Usage

### In Components

```tsx
import { useTranslation } from '../src/hooks/useTranslation';

export default function MyComponent() {
  const { t, currentLanguage, isFrench, isEnglish } = useTranslation();
  
  return (
    <Text>{t('common.cancel')}</Text>
  );
}
```

### Translation Keys

Translation keys are organized in a hierarchical structure:

- `common.*` - Common UI elements (buttons, labels, etc.)
- `auth.*` - Authentication-related text
- `games.*` - Game-related text
- `profile.*` - Profile screen text
- `about.*` - About screen text
- `rules.*` - Rules screen text
- `navigation.*` - Navigation labels
- `notFound.*` - 404 page text

### Adding New Translations

1. **Add to English file** (`src/i18n/locales/en.json`):
```json
{
  "newSection": {
    "newKey": "English text"
  }
}
```

2. **Add to French file** (`src/i18n/locales/fr.json`):
```json
{
  "newSection": {
    "newKey": "Texte français"
  }
}
```

3. **Use in component**:
```tsx
const { t } = useTranslation();
<Text>{t('newSection.newKey')}</Text>
```

## Testing

Run the i18n tests:
```bash
npm test -- src/i18n/__tests__/i18n.test.ts
```

## Manual Language Switching

For development/testing, you can manually change the language:

```tsx
import { useTranslation } from '../src/hooks/useTranslation';

export default function LanguageSwitcher() {
  const { i18n } = useTranslation();
  
  const switchToFrench = () => i18n.changeLanguage('fr');
  const switchToEnglish = () => i18n.changeLanguage('en');
  
  return (
    <View>
      <Button onPress={switchToFrench} title="Français" />
      <Button onPress={switchToEnglish} title="English" />
    </View>
  );
}
```

## Best Practices

1. **Use descriptive keys**: Instead of `t('cancel')`, use `t('common.cancel')`
2. **Keep translations organized**: Group related translations under meaningful namespaces
3. **Test both languages**: Always verify that both French and English translations work
4. **Use interpolation for dynamic content**: `t('welcome', { name: userName })`
5. **Avoid hardcoded strings**: All user-facing text should use translation keys 