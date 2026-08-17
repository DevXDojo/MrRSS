import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'
import { pathToFileURL } from 'url'

const configDir = import.meta.dirname

// Mock plugin for static assets
function mockAssets() {
  return {
    name: 'vite-mock-assets',
    enforce: 'pre',
    resolveId(id) {
      // Redirect public asset imports to proper file paths
      if (id.startsWith('/assets/') || id.startsWith('/plugin_icons/')) {
        const resolvedPath = resolve(configDir, 'public', id.substring(1))
        return resolvedPath
      }
      return null
    },
    load(id) {
      // Mock SVG and image imports
      if (/\.(svg|png|jpg|jpeg|gif|webp|ico|woff|woff2|ttf|eot)$/.test(id)) {
        // Return the file path as a URL string for src attributes
        const fileUrl = pathToFileURL(id).href
        return {
          code: `export default ${JSON.stringify(fileUrl)}`,
          map: null
        }
      }
      return null
    },
  }
}

export default defineConfig({
  plugins: [vue(), mockAssets()],
  resolve: {
    alias: {
      '@': resolve(configDir, './src'),
      '@wailsio/runtime': resolve(configDir, './src/test/mocks/wails.ts')
    }
  },
  publicDir: 'public',
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/test/setup.ts'],
    server: {
      fs: {
        // Allow access to the public directory in tests
        allow: [resolve(configDir, '.'), resolve(configDir, 'public')]
      }
    },
    deps: {
      inline: ['@wailsio/runtime']
    }
  }
})
