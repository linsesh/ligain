/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./app/**/*.{ts,tsx}", "./src/**/*.{ts,tsx}"],
  presets: [require("nativewind/preset")],
  theme: {
    extend: {
      fontFamily: {
        sans: ['HKGroteskWide-Regular'],
        'hk-light': ['HKGroteskWide-Light'],
        'hk-medium': ['HKGroteskWide-Medium'],
        'hk-semibold': ['HKGroteskWide-SemiBold'],
        'hk-bold': ['HKGroteskWide-Bold'],
        'hk-extrabold': ['HKGroteskWide-ExtraBold'],
        'hk-black': ['HKGroteskWide-Black'],
      },
      colors: {
        background: '#e4e9ef',
        card: '#ffffff',
        foreground: '#1d1d1d',
        'foreground-secondary': '#8e8e93',
        border: '#ccd2d7',
        primary: '#f25702',
        secondary: '#e9a317',
        surface: '#eef1f4',
        link: '#469dff',
        success: '#2e7d32',
        error: '#ff0000',
        warning: '#e65100',
      },
    },
  },
  plugins: [],
};
