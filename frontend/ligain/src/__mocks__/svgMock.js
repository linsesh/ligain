// Mock for SVG files in Jest tests
const React = require('react');

const SvgMock = React.forwardRef((props, ref) => {
  return React.createElement('svg', { ...props, ref });
});

SvgMock.displayName = 'SvgMock';

module.exports = SvgMock;
