import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import viteCompression from 'vite-plugin-compression'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    viteCompression(),
  ],
  esbuild: {
    legalComments: 'none',
  },
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  build: {
    chunkSizeWarningLimit: 666,
    rollupOptions: {
      input: {
        app: './index.html', // the default entry point
        service: './service-worker.ts',
      },
      output: {
        entryFileNames: assetInfo => {
          switch (assetInfo.name) {
            case 'service':
              return 'service-worker.js'

            default:
              return 'assets/[name]-[hash].js'
          }
        }
      }
    },
    sourcemap: false,
  },
  server: {
    host: '0.0.0.0',
    port: 8080,
    strictPort: true,
  },
})
