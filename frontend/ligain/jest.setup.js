require('@testing-library/jest-dom');

// Mock expo-constants
jest.mock('expo-constants', () => ({
  expoConfig: {
    extra: {
      apiBaseUrl: 'https://test-api.example.com',
      apiKey: 'test-api-key',
      environment: 'test'
    }
  }
}));

// Mock AsyncStorage
jest.mock('@react-native-async-storage/async-storage', () => ({
  getItem: jest.fn(),
  setItem: jest.fn(),
  multiRemove: jest.fn(),
}));

// Mock expo-localization
jest.mock('expo-localization', () => ({
  locale: 'en',
  locales: ['en', 'fr'],
  isRTL: false,
  getLocales: () => [
    {
      languageCode: 'en',
      countryCode: 'US',
      languageTag: 'en-US',
      decimalSeparator: '.',
      groupingSeparator: ',',
    },
  ],
  getCalendars: () => [
    {
      id: 'gregorian',
      calendar: 'gregorian',
      locale: 'en-US',
    },
  ],
  getTimeZone: () => 'America/New_York',
  is24Hour: () => false,
  getCurrencyCode: () => 'USD',
  getDecimalSeparator: () => '.',
  getGroupingSeparator: () => ',',
}));

// Mock react-native
jest.mock('react-native', () => {
  const RN = jest.requireActual('react-native');
  return {
    ...RN,
    Alert: {
      alert: jest.fn(),
    },
    Clipboard: {
      setString: jest.fn(),
    },
    Keyboard: {
      dismiss: jest.fn(),
    },
  };
});

// Mock @expo/vector-icons
jest.mock('@expo/vector-icons', () => ({
  Ionicons: 'Ionicons',
  AntDesign: 'AntDesign',
  MaterialIcons: 'MaterialIcons',
  FontAwesome: 'FontAwesome',
}));

// Mock console methods to reduce noise in tests
global.console = {
  ...console,
  log: jest.fn(),
  debug: jest.fn(),
  info: jest.fn(),
  warn: jest.fn(),
  error: jest.fn(),
};

// Mock React Native globals
global.__DEV__ = false; 