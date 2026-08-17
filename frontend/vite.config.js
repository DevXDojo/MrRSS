import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

const configDir = import.meta.dirname

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': resolve(configDir, './src')
    }
  },
  test: {
    globals: true,
    environment: 'jsdom'
  }
})
