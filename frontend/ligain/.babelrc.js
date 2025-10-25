module.exports = function (api) {
  api.cache(true);
  return {
    presets: ['babel-preset-expo'],
    plugins: [
      // Ensure test dependencies are not included in production builds
      process.env.NODE_ENV === 'production' && [
        'babel-plugin-transform-remove-imports',
        {
          test: /@testing-library|jest|react-test-renderer/,
        },
      ],
    ].filter(Boolean),
  };
};
