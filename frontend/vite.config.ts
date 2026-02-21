import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import wails from '@wailsio/runtime/plugins/vite'
import unocss from 'unocss/vite'
import Markdown from 'unplugin-vue-markdown/vite'
import hljs from 'highlight.js'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { NaiveUiResolver } from 'unplugin-vue-components/resolvers'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    vue({ include: [/\.vue$/, /\.md$/] }),
    Markdown({
      markdownItSetup(md) {
        md.options.highlight = (str, lang) => {
          if (lang && hljs.getLanguage(lang)) {
            try {
              return hljs.highlight(str, {
                language: lang,
                ignoreIllegals: true,
              }).value
            } catch {}
          }
          try {
            return hljs.highlightAuto(str).value
          } catch {
            return md.utils.escapeHtml(str)
          }
        }
        const fence = md.renderer.rules.fence
        md.renderer.rules.fence = (tokens, idx, options, env, self) => {
          const token = tokens[idx]
          token.attrJoin('class', 'hljs')
          return (fence || self.renderToken).call(
            self,
            tokens,
            idx,
            options,
            env,
            self,
          )
        }
      },
    }),
    wails('./bindings'),
    unocss(),
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
      {
        find: /^monaco-editor$/,
        replacement: 'monaco-editor/esm/vs/editor/editor.api',
      },
    ],
  },
  optimizeDeps: {
    exclude: ['monaco-editor'],
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
