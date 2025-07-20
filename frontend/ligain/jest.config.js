module.exports = {
  preset: 'ts-jest',
  testEnvironment: 'jsdom',
  moduleFileExtensions: ['ts', 'tsx', 'js', 'jsx', 'json', 'node'],
  transform: {
    '^.+\\.(ts|tsx)$': ['ts-jest', {
      tsconfig: {
        jsx: 'react-jsx',
      },
    }],
  },
  transformIgnorePatterns: [
    'node_modules/(?!(expo-localization|@expo|expo|react-native|@react-native|react-i18next|i18next|@testing-library|@expo/vector-icons)/)',
  ],
  testMatch: [
    '**/src/**/__tests__/**/*.test.(ts|tsx)',
    '**/src/**/*.test.(ts|tsx)',
    '**/hooks/**/*.test.(ts|tsx)'
  ],
  moduleNameMapper: {
    '^@/(.*)$': '<rootDir>/src/$1',
    '^react-native$': 'react-native-web',
    '^expo-localization$': '<rootDir>/src/__mocks__/expo-localization.js',
  },
  setupFilesAfterEnv: ['<rootDir>/jest.setup.js'],
  testEnvironmentOptions: {
    customExportConditions: [''],
  },
}; 