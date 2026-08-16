export default {
  plugins: {
    '@tailwindcss/postcss': {},
    autoprefixer: {
      // Ensure proper vendor prefix ordering
      cascade: true,
      grid: 'autoplace',
    },
  },
};
