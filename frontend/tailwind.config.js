/** @type {import('tailwindcss').Config} */
export default {
  darkMode: ['class', '.dark-mode'],
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  theme: {
    extend: {
      fontFamily: {
        sans: ['Inter', 'Noto Sans CJK SC', 'system-ui', 'sans-serif'],
      },
      fontSize: {
        xs: [
          'calc(0.75rem * var(--ui-font-scale, 1))',
          { lineHeight: 'calc(1rem * var(--ui-font-scale, 1))' },
        ],
        sm: [
          'calc(0.875rem * var(--ui-font-scale, 1))',
          { lineHeight: 'calc(1.25rem * var(--ui-font-scale, 1))' },
        ],
        base: [
          'calc(1rem * var(--ui-font-scale, 1))',
          { lineHeight: 'calc(1.5rem * var(--ui-font-scale, 1))' },
        ],
        lg: [
          'calc(1.125rem * var(--ui-font-scale, 1))',
          { lineHeight: 'calc(1.75rem * var(--ui-font-scale, 1))' },
        ],
        xl: [
          'calc(1.25rem * var(--ui-font-scale, 1))',
          { lineHeight: 'calc(1.75rem * var(--ui-font-scale, 1))' },
        ],
        '2xl': [
          'calc(1.5rem * var(--ui-font-scale, 1))',
          { lineHeight: 'calc(2rem * var(--ui-font-scale, 1))' },
        ],
      },
      colors: {
        bg: {
          primary: 'var(--bg-primary)',
          secondary: 'var(--bg-secondary)',
          tertiary: 'var(--bg-tertiary)',
        },
        text: {
          primary: 'var(--text-primary)',
          secondary: 'var(--text-secondary)',
        },
        accent: {
          DEFAULT: 'var(--accent-color)',
          hover: 'var(--accent-hover)',
        },
        border: 'var(--border-color)',
      },
    },
  },
  plugins: [],
};
