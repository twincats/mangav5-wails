import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import wails from '@wailsio/runtime/plugins/vite'
import unocss from 'unocss/vite'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { NaiveUiResolver } from 'unplugin-vue-components/resolvers'
import _monacoEditorPlugin from 'vite-plugin-monaco-editor'

const monacoEditorPlugin =
  (_monacoEditorPlugin as any).default || _monacoEditorPlugin

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    wails('./bindings'),
    unocss(),
    // monacoEditorPlugin({
    //   languageWorkers: ['json', 'editorWorkerService'], // Load only JSON and core editor workers
    // }),
    AutoImport({
      imports: [
        'vue',
        'vue-router',
        {
          'naive-ui': [
            'useDialog',
            'useMessage',
            'useNotification',
            'useLoadingBar',
          ],
        },
      ],
    }),
    Components({ resolvers: [NaiveUiResolver()] }),
  ],
  resolve: {
    alias: [
      {
        find: '@',
        replacement: fileURLToPath(new URL('./src', import.meta.url)),
      },
      {
        find: 'bindings',
        replacement: fileURLToPath(new URL('./bindings', import.meta.url)),
      },
    ],
  },
  worker: {
    format: 'es',
  },
  build: {
    chunkSizeWarningLimit: 4000,
    rollupOptions: {
      output: {
        manualChunks: {
          monaco: ['monaco-editor'],
          naive: ['naive-ui'],
        },
      },
    },
  },
})
