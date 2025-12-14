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
  // Ensure test dependencies are only used in test environment
  testPathIgnorePatterns: [
    '/node_modules/',
    '/android/',
    '/ios/',
  ],
  testMatch: [
    '**/src/**/__tests__/**/*.test.(ts|tsx)',
    '**/src/**/*.test.(ts|tsx)',
    '**/hooks/**/*.test.(ts|tsx)'
  ],
  moduleNameMapper: {
    // SVG mock must come before other asset mappings
    '\\.svg$': '<rootDir>/src/__mocks__/svgMock.js',
    '^@/(.*)$': '<rootDir>/src/$1',
    '^@assets/(.*)$': '<rootDir>/assets/$1',
    '^react-native$': 'react-native-web',
    '^expo-localization$': '<rootDir>/src/__mocks__/expo-localization.js',
  },
  setupFilesAfterEnv: ['<rootDir>/jest.setup.js'],
  testEnvironmentOptions: {
    customExportConditions: [''],
  },
}; 