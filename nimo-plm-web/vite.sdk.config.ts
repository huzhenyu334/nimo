import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'
import { readFileSync } from 'fs'

// Read version from package.json
const pkg = JSON.parse(readFileSync(path.resolve(__dirname, 'package.json'), 'utf-8'))

/**
 * Vite config for building the PLM Component SDK as a single UMD bundle.
 *
 * Output: dist-sdk/nimo-plm-components.js
 *
 * Key decisions:
 *   - UMD format so it works via <script> tag
 *   - All dependencies (React, ReactDOM, antd, etc.) are INLINED
 *     because the host page is NOT expected to have React
 *   - SDK version is injected via define
 */
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  define: {
    'process.env.NODE_ENV': JSON.stringify('production'),
    // Replace __SDK_VERSION__ in source code
    '__SDK_VERSION__': JSON.stringify(pkg.version || '0.1.0'),
  },
  build: {
    outDir: 'dist-sdk',
    emptyOutDir: true,
    sourcemap: true,
    lib: {
      entry: path.resolve(__dirname, 'src/sdk/index.ts'),
      name: 'nimoComponent',
      fileName: () => 'nimo-plm-components.js',
      formats: ['umd'],
    },
    rollupOptions: {
      // Do NOT externalize anything — bundle everything into the single file
      external: [],
      output: {
        // No code splitting for UMD
        inlineDynamicImports: true,
        // Global variable name
        name: 'nimoComponent',
        // Use named exports to avoid "default" wrapper
        exports: 'named' as const,
        // Ensure CSS is extracted alongside
        assetFileNames: 'nimo-plm-components.[ext]',
      },
    },
    // Increase chunk size warning limit since we're bundling everything
    chunkSizeWarningLimit: 5000,
    // Minify for production
    minify: 'esbuild',
  },
  // Disable CSS code splitting — inline all styles
  css: {
    // Antd uses CSS-in-JS (cssinjs), so most styles are runtime.
    // Any imported .css files will be extracted to nimo-plm-components.css
  },
})
