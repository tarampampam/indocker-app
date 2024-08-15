/// <reference types="vite/client" />

import { defineConfig, type PluginOption } from 'vite'
import react from '@vitejs/plugin-react'
import { resolve, join } from 'path'
import { buildSync } from 'esbuild'

const rootDir = resolve(__dirname)
const [distDir, srcDir] = [join(rootDir, 'dist'), join(rootDir, 'src')]

const buildServiceWorkerPlugin: PluginOption = {
  name: 'service-worker',
  apply: 'build',
  enforce: 'post',
  transformIndexHtml() {
    buildSync({
      minify: true,
      bundle: true,
      loader: { '.svg': 'text', '.png': 'text' },
      entryPoints: [join(srcDir, 'service-worker.ts')],
      outfile: join(distDir, 'service-worker.js'),
    })
  },
}

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react(), buildServiceWorkerPlugin],
  resolve: {
    alias: {
      '~': srcDir,
    },
  },
  define: {
    __LATEST_RELEASE_LINK__: JSON.stringify('https://github.com/tarampampam/indocker-app/releases/latest'),
  },
  build: {
    emptyOutDir: true,
    rollupOptions: {
      input: {
        app: join(rootDir, 'index.html'), // the default entry point
      },
      output: {
        entryFileNames: 'js/[name]-[hash].js',
        chunkFileNames: 'js/[name]-[hash].js',
        assetFileNames: 'assets/[name]-[hash].[ext]',
      },
    },
    sourcemap: false,
  },
  esbuild: {
    legalComments: 'none',
  },
})
