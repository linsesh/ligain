/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./app/**/*.{ts,tsx}", "./src/**/*.{ts,tsx}"],
  presets: [require("nativewind/preset")],
  theme: {
    extend: {
      colors: {
        background: '#e4e9ef',
        card: '#ffffff',
        foreground: '#1d1d1d',
        'foreground-secondary': '#8e8e93',
        border: '#ccd2d7',
        primary: '#469dff',
        secondary: '#e9a317',
        link: '#469dff',
        success: '#2e7d32',
        error: '#ff0000',
        warning: '#e65100',
      },
    },
  },
  plugins: [],
};
