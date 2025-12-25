import vue from '@vitejs/plugin-vue'
import AutoImport from 'unplugin-auto-import/vite'
import { NaiveUiResolver } from 'unplugin-vue-components/resolvers'
import Components from 'unplugin-vue-components/vite'
import { defineConfig } from 'vite'
import { fileURLToPath } from 'url'
import { dirname, resolve } from 'path'
import pkg from './package.json'
import vitePluginBundleObfuscator from 'vite-plugin-bundle-obfuscator'

const __filename = fileURLToPath(import.meta.url)
const __dirname = dirname(__filename)

export default defineConfig({
  define: {
    __APP_VERSION__: JSON.stringify(pkg.version)
  },
  resolve: {
    alias: {
      '@': resolve(__dirname, './src')
    }
  },
  base: './',
  build: {
    outDir: 'dist',
    minify: 'esbuild',
    chunkSizeWarningLimit: 800,
    rollupOptions: {
      output: {
        manualChunks: id => {
          if (id.includes('node_modules')) {
            if (id.includes('naive-ui')) return 'naive-ui'
            if (id.includes('vue')) return 'vue-vendor'
            if (id.includes('pinia')) return 'pinia-vendor'
            return 'vendor'
          }
        },
        chunkFileNames: 'assets/js/[name]-[hash].js',
        entryFileNames: 'assets/js/[name]-[hash].js',
        assetFileNames: () => {
          return 'assets/[ext]/[name]-[hash].[ext]'
        }
      }
    },
    terserOptions: {
      compress: {
        drop_console: false,
        drop_debugger: true
      }
    }
  },
  preview: {
    allowedHosts: true
  },
  plugins: [
    vue(),
    AutoImport({
      imports: [
        'vue',
        {
          'naive-ui': ['useDialog', 'useMessage', 'useNotification', 'useLoadingBar']
        }
      ]
    }),
    Components({
      resolvers: [NaiveUiResolver()]
    }),
    vitePluginBundleObfuscator({
      log: false,
      enable: true,
      options: {
        log: false,
        compact: true,
        stringArray: true,
        renameGlobals: false,
        selfDefending: false,
        debugProtection: false,
        rotateStringArray: true,
        deadCodeInjection: false,
        stringArrayEncoding: ['none'],
        disableConsoleOutput: true,
        stringArrayThreshold: 0.75,
        controlFlowFlattening: false,
        unicodeEscapeSequence: true,
        identifierNamesGenerator: 'hexadecimal'
      },
      // excludes: ['router.js'],
      autoExcludeNodeModules: true
    })
  ],
  worker: {
    format: 'es'
  },
  server: {
    port: parseInt(process.env.VITE_PORT) || 2025,
    proxy: {
      '/api': {
        target: 'http://localhost:3000',
        changeOrigin: true,
        secure: false
      }
    }
  }
})
