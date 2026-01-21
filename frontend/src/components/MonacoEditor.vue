<template>
  <div ref="editorContainer" :style="containerStyle"></div>
</template>

<script setup lang="ts">
import * as monaco from 'monaco-editor'
import 'monaco-editor/esm/vs/language/json/monaco.contribution'
import EditorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker'
import JsonWorker from 'monaco-editor/esm/vs/language/json/json.worker?worker'
import {
  computed,
  onMounted,
  onBeforeUnmount,
  ref,
  watch,
  shallowRef,
  toRaw,
} from 'vue'
import { scrapingRuleSnippets } from '../config/monacoSnippets'
import { registerSchema, unregisterSchema } from '../utils/monacoSchemaRegistry'

type JsonSchema = Record<string, unknown>

interface Props {
  modelValue: string
  language?: string
  theme?: 'vs-dark' | 'vs-light'
  textAlign?: 'left' | 'center' | 'right'
  modelUri?: string
  jsonValidate?: boolean
  jsonSchema?: JsonSchema
  jsonSchemaUri?: string
  jsonSchemaFileMatch?: string[]
  formatOnLoad?: boolean
  readOnly?: boolean
}
const props = withDefaults(defineProps<Props>(), {
  language: 'javascript',
  theme: 'vs-dark',
  textAlign: 'left',
  jsonValidate: false,
  formatOnLoad: false,
  readOnly: false,
})

const emit = defineEmits(['update:modelValue', 'change'])

const containerStyle = computed(() => {
  return {
    width: '100%',
    height: '100%',
    '--monaco-text-align': props.textAlign,
  } as unknown as Record<string, string>
})

const editorContainer = ref<HTMLDivElement | null>(null)
// Use shallowRef for non-reactive complex objects if we wanted to expose them,
// but local variables are fine for Monaco instances to avoid Proxy overhead.
let editor: monaco.editor.IStandaloneCodeEditor | null = null
let model: monaco.editor.ITextModel | null = null
let modelUri: monaco.Uri | null = null
let modelChangeDisposable: monaco.IDisposable | null = null
let resizeObserver: ResizeObserver | null = null
let registeredSchemaUri: string | null = null

function createDefaultModelUri(language: string) {
  const id =
    globalThis.crypto?.randomUUID?.() ??
    `${Date.now()}-${Math.random().toString(16).slice(2)}`
  const ext = language === 'json' ? 'json' : 'txt'
  return `inmemory://model/${id}.${ext}`
}

function tryFormatJson(value: string) {
  try {
    return JSON.stringify(JSON.parse(value), null, 2)
  } catch {
    return value
  }
}

function applyJsonSchemaOptions() {
  if (!model || model.getLanguageId() !== 'json') return

  // If we have a schema, update the global registry
  if (props.jsonSchema) {
    const uniqueId =
      globalThis.crypto?.randomUUID?.() ??
      `${Date.now()}-${Math.random().toString(16).slice(2)}`

    // Generate a unique URI if not provided, ensuring uniqueness per component
    const schemaUri =
      props.jsonSchemaUri ?? `inmemory://schema/${uniqueId}.json`
    registeredSchemaUri = schemaUri

    // If modelUri is set, use it for file matching
    // We must ensure exact string match for the worker to pick it up
    // Using model.uri.toString() is the safest way to get the exact URI string known to Monaco
    // Prefer using the actual model's URI if available, otherwise the prop
    const targetUriStr = model ? model.uri.toString() : modelUri?.toString()

    // Robust file matching:
    // 1. Exact match (targetUriStr)
    // 2. Basename match (**/filename.json) to handle path prefix differences
    // We avoid global *.json to ensure isolation between multiple editors
    // const fileName = targetUriStr?.split('/').pop()
    const fileMatch = [targetUriStr].filter(Boolean) as string[]

    // Deep clone the schema to avoid any Proxy/Reactivity issues from Vue
    // and ensure it's a plain JSON object
    const schemaCopy = JSON.parse(JSON.stringify(toRaw(props.jsonSchema)))

    // Explicitly set the $id to match the URI we are registering
    schemaCopy.$id = schemaUri
    schemaCopy.id = schemaUri // Legacy support

    registerSchema(schemaUri, fileMatch, schemaCopy)
  } else if (registeredSchemaUri) {
    // If we previously had a schema but now don't, remove it
    unregisterSchema(registeredSchemaUri)
    registeredSchemaUri = null
  }
}

function registerJsonSnippets() {
  const globalAny = globalThis as any
  if (globalAny.__scrapingRuleSnippetsRegistered) return

  monaco.languages.registerCompletionItemProvider('json', {
    provideCompletionItems(model, position) {
      const word = model.getWordUntilPosition(position)
      const range = new monaco.Range(
        position.lineNumber,
        word.startColumn,
        position.lineNumber,
        word.endColumn,
      )

      const suggestions = scrapingRuleSnippets.map(snippet => ({
        ...snippet,
        range,
      })) as monaco.languages.CompletionItem[]

      return { suggestions }
    },
  })
  globalAny.__scrapingRuleSnippetsRegistered = true
}

function ensureMonacoWorkers() {
  const globalAny = globalThis as any
  if (globalAny.MonacoEnvironment?.getWorker) return

  globalAny.MonacoEnvironment = {
    getWorker(_moduleId: unknown, label: string) {
      if (label === 'json') return new JsonWorker()
      return new EditorWorker()
    },
  }
}

onMounted(() => {
  ensureMonacoWorkers()
  const language = props.language ?? 'javascript'
  const uriString = props.modelUri ?? createDefaultModelUri(language)
  modelUri = monaco.Uri.parse(uriString)

  let initialValue = props.modelValue
  if (language === 'json' && props.formatOnLoad) {
    initialValue = tryFormatJson(initialValue)
  }

  // Check if model already exists to prevent "Model already exists" error
  model = monaco.editor.getModel(modelUri)

  if (!model) {
    model = monaco.editor.createModel(initialValue, language, modelUri)
  } else {
    // If model exists, just update value if needed
    // Be careful not to overwrite user's work if they just navigated away and back?
    // Current logic assumes prop is source of truth.
    if (model.getValue() !== initialValue) {
      model.setValue(initialValue)
    }
    monaco.editor.setModelLanguage(model, language)
  }

  editor = monaco.editor.create(editorContainer.value!, {
    model,
    theme: props.theme,
    automaticLayout: false, // Disabled for performance, using ResizeObserver instead
    minimap: { enabled: false },
    readOnly: props.readOnly,
    scrollBeyondLastLine: false,
    fixedOverflowWidgets: true,
  })

  // Resize Observer Implementation
  resizeObserver = new ResizeObserver(() => {
    editor?.layout()
  })
  if (editorContainer.value) {
    resizeObserver.observe(editorContainer.value)
  }

  applyJsonSchemaOptions()

  if (language === 'json') {
    registerJsonSnippets()
  }

  modelChangeDisposable = editor.onDidChangeModelContent(() => {
    const value = editor!.getValue()
    emit('update:modelValue', value)
    emit('change', value)
  })
})

watch(
  () => props.modelValue,
  newVal => {
    if (editor && model && newVal !== model.getValue()) {
      let nextValue = newVal
      if (props.language === 'json' && props.formatOnLoad) {
        nextValue = tryFormatJson(newVal)
      }
      // Use executeEdits to preserve undo stack if desired, but setValue is standard for full replacement
      model.setValue(nextValue)
    }
  },
)

watch(
  () => props.theme,
  newTheme => {
    monaco.editor.setTheme(newTheme)
  },
)

watch(
  () => props.language,
  newLanguage => {
    if (!model || !newLanguage) return
    monaco.editor.setModelLanguage(model, newLanguage)
    applyJsonSchemaOptions()
  },
)

watch(
  () => props.readOnly,
  newReadOnly => {
    editor?.updateOptions({ readOnly: newReadOnly })
  },
)

watch(
  () => [
    props.jsonValidate,
    props.jsonSchema,
    props.jsonSchemaUri,
    props.jsonSchemaFileMatch,
  ],
  () => applyJsonSchemaOptions(),
  { deep: true },
)

onBeforeUnmount(() => {
  resizeObserver?.disconnect()
  modelChangeDisposable?.dispose()
  editor?.dispose()

  if (registeredSchemaUri) {
    unregisterSchema(registeredSchemaUri)
  }

  // Only dispose model if we created it via default URI (temp model)
  // If user provided a specific URI, they might want to persist it.
  // For safety in this specific implementation:
  // If props.modelUri was NOT provided, we definitely created a temp model -> dispose it.
  if (!props.modelUri && model) {
    model.dispose()
  }
})
</script>

<style>
.monaco-editor,
.monaco-editor .view-lines,
.monaco-editor .margin {
  text-align: var(--monaco-text-align, left);
}
</style>
