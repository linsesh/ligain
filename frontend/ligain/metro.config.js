const { getDefaultConfig } = require('expo/metro-config');

const config = getDefaultConfig(__dirname);

// Configure SVG transformer
config.resolver.assetExts = config.resolver.assetExts.filter(ext => ext !== 'svg');
config.resolver.sourceExts.push('svg');
config.transformer.babelTransformerPath = require.resolve('react-native-svg-transformer');

// Exclude testing dependencies from production builds
config.resolver.blockList = [
  // Block testing libraries from being included in production bundles
  /node_modules\/@testing-library\/.*/,
  /node_modules\/jest\/.*/,
  /node_modules\/@types\/jest\/.*/,
  /node_modules\/ts-jest\/.*/,
  /node_modules\/jest-environment-jsdom\/.*/,
  /node_modules\/react-test-renderer\/.*/,
  // Block test files from being included in production bundles
  /.*\.test\.(js|jsx|ts|tsx)$/,
  /.*\.spec\.(js|jsx|ts|tsx)$/,
  /.*\/__tests__\/.*/,
];

// Ensure test files are not included in production bundles
config.resolver.platforms = ['ios', 'android', 'native', 'web'];

// Additional resolver configuration to prevent test dependencies from being bundled
config.resolver.resolverMainFields = ['react-native', 'browser', 'main'];

module.exports = config;
