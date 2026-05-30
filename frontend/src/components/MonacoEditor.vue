<template>
  <div ref="editorContainer" :style="containerStyle"></div>
</template>

<script setup lang="ts">
import * as monaco from 'monaco-editor'
import 'monaco-editor/esm/vs/language/json/monaco.contribution'
import EditorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker'
import JsonWorker from 'monaco-editor/esm/vs/language/json/json.worker?worker'
import { computed, onMounted, onBeforeUnmount, ref, watch, toRaw } from 'vue'
import { scrapingRuleSnippets } from '../config/monacoSnippets'
import { registerSchema, unregisterSchema } from '../utils/monacoSchemaRegistry'
import { useDebounceFn } from '@vueuse/core'

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
  customValidator?: (value: string) => monaco.editor.IMarkerData[]
}
const props = withDefaults(defineProps<Props>(), {
  language: 'javascript',
  theme: 'vs-dark',
  textAlign: 'left',
  jsonValidate: false,
  formatOnLoad: false,
  readOnly: false,
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'change', value: string): void
  (e: 'validate', isValid: boolean): void
}>()

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
let schemaCompletionDisposable: monaco.IDisposable | null = null

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

    const schemaUri =
      props.jsonSchemaUri ??
      registeredSchemaUri ??
      `inmemory://schema/${uniqueId}.json`

    if (registeredSchemaUri && registeredSchemaUri !== schemaUri) {
      unregisterSchema(registeredSchemaUri)
    }
    registeredSchemaUri = schemaUri

    const fileMatch = (() => {
      if (props.jsonSchemaFileMatch && props.jsonSchemaFileMatch.length > 0) {
        return props.jsonSchemaFileMatch
      }

      const matches = new Set<string>()
      const uri = model?.uri ?? modelUri
      const uriStr = uri?.toString()

      if (uriStr) matches.add(uriStr)
      if (uri?.path) {
        matches.add(uri.path)
        const baseName = uri.path.split('/').pop()
        if (baseName) {
          matches.add(`**/${baseName}`)
          matches.add(baseName)
        }
      }

      return Array.from(matches)
    })()

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

function getSchemaRootProperties(schema: any): string[] {
  const propsObj = schema?.properties
  if (!propsObj || typeof propsObj !== 'object') return []
  return Object.keys(propsObj).filter(
    k => typeof k === 'string' && k.length > 0,
  )
}

function getSchemaEnumForProperty(schema: any, prop: string): string[] {
  const def = schema?.properties?.[prop]
  const list = def?.enum
  if (!Array.isArray(list)) return []
  return list.filter((v: any) => typeof v === 'string')
}

function resolveSchemaNode(root: any, node: any): any {
  let cur = node
  for (let i = 0; i < 10; i++) {
    if (!cur || typeof cur !== 'object') return cur
    const ref = (cur as any).$ref
    if (typeof ref !== 'string' || !ref.startsWith('#/')) return cur
    const parts = ref.slice(2).split('/').filter(Boolean)
    let target: any = root
    for (const p of parts) {
      if (target && typeof target === 'object' && p in target) {
        target = target[p]
      } else {
        target = null
        break
      }
    }
    cur = target
  }
  return cur
}

type JsonCtx = {
  kind: 'object' | 'array'
  keyFromParent?: string
  expectingKey?: boolean
  activeValueKey?: string
}

function computeJsonContext(text: string): JsonCtx[] {
  const stack: JsonCtx[] = []

  let inStr = false
  let esc = false
  let currentString = ''
  let lastString: string | null = null

  let expectingColonForKey = false
  let valueKeyForNextContainer: string | null = null
  let inPrimitiveValue = false

  const topObject = () => {
    for (let i = stack.length - 1; i >= 0; i--) {
      if (stack[i].kind === 'object') return stack[i]
    }
    return null
  }

  const setObjectExpectingKey = (v: boolean) => {
    const obj = topObject()
    if (!obj) return
    obj.expectingKey = v
    if (v) obj.activeValueKey = undefined
  }

  for (let i = 0; i < text.length; i++) {
    const ch = text[i]

    if (inStr) {
      if (esc) {
        esc = false
        currentString += ch
        continue
      }
      if (ch === '\\') {
        esc = true
        continue
      }
      if (ch === '"') {
        inStr = false
        lastString = currentString
        currentString = ''
        expectingColonForKey = true
        continue
      }
      currentString += ch
      continue
    }

    if (ch === '"') {
      inStr = true
      esc = false
      currentString = ''
      continue
    }

    if (expectingColonForKey) {
      if (ch === ':') {
        const obj = topObject()
        if (obj && obj.expectingKey && typeof lastString === 'string') {
          obj.activeValueKey = lastString
          obj.expectingKey = false
          valueKeyForNextContainer = lastString
          inPrimitiveValue = false
        }
        expectingColonForKey = false
        lastString = null
        continue
      }
      if (ch.trim() !== '') {
        expectingColonForKey = false
        lastString = null
      }
    }

    if (ch === '{') {
      const key = valueKeyForNextContainer ?? undefined
      valueKeyForNextContainer = null
      inPrimitiveValue = false
      stack.push({ kind: 'object', keyFromParent: key, expectingKey: true })
      continue
    }

    if (ch === '[') {
      const key = valueKeyForNextContainer ?? undefined
      valueKeyForNextContainer = null
      inPrimitiveValue = false
      stack.push({ kind: 'array', keyFromParent: key })
      continue
    }

    if (ch === '}' || ch === ']') {
      if (inPrimitiveValue) {
        inPrimitiveValue = false
        setObjectExpectingKey(true)
      }
      if (stack.length > 0) {
        stack.pop()
      }
      continue
    }

    if (ch === ',') {
      if (inPrimitiveValue) {
        inPrimitiveValue = false
      }
      setObjectExpectingKey(true)
      continue
    }

    if (valueKeyForNextContainer) {
      const ws = ch.trim() === ''
      if (!ws && ch !== '{' && ch !== '[' && ch !== '"') {
        inPrimitiveValue = true
        valueKeyForNextContainer = null
      }
    } else if (!inPrimitiveValue) {
      const obj = topObject()
      if (obj && obj.expectingKey === false) {
        const ws = ch.trim() === ''
        if (!ws && ch !== '{' && ch !== '[' && ch !== '"') {
          inPrimitiveValue = true
        }
      }
    }
  }

  return stack
}

function resolveSchemaForContext(schemaRoot: any, ctx: JsonCtx[]): any {
  let node: any = schemaRoot
  for (const c of ctx) {
    node = resolveSchemaNode(schemaRoot, node)
    if (c.keyFromParent) {
      node = node?.properties?.[c.keyFromParent]
      node = resolveSchemaNode(schemaRoot, node)
    }
    if (c.kind === 'array') {
      node = resolveSchemaNode(schemaRoot, node)
      node = node?.items
      node = resolveSchemaNode(schemaRoot, node)
    }
  }
  return resolveSchemaNode(schemaRoot, node)
}

function schemaProperties(node: any): string[] {
  const propsObj = node?.properties
  if (!propsObj || typeof propsObj !== 'object') return []
  return Object.keys(propsObj).filter(
    k => typeof k === 'string' && k.length > 0,
  )
}

function schemaEnum(node: any): string[] {
  const list = node?.enum
  if (!Array.isArray(list)) return []
  return list.filter((v: any) => typeof v === 'string')
}

function getDepthOutsideStrings(text: string) {
  let depth = 0
  let inStr = false
  let esc = false
  for (let i = 0; i < text.length; i++) {
    const ch = text[i]
    if (inStr) {
      if (esc) {
        esc = false
        continue
      }
      if (ch === '\\') {
        esc = true
        continue
      }
      if (ch === '"') {
        inStr = false
      }
      continue
    }
    if (ch === '"') {
      inStr = true
      continue
    }
    if (ch === '{' || ch === '[') depth++
    else if (ch === '}' || ch === ']') depth = Math.max(0, depth - 1)
  }
  return { depth, inStr }
}

function registerSchemaCompletionFallback() {
  schemaCompletionDisposable?.dispose()
  schemaCompletionDisposable = null

  if (!model || model.getLanguageId() !== 'json') return
  if (!props.jsonSchema) return

  const schemaRoot = JSON.parse(JSON.stringify(toRaw(props.jsonSchema)))
  if (schemaProperties(schemaRoot).length === 0) return

  const targetUri = model.uri.toString()

  schemaCompletionDisposable = monaco.languages.registerCompletionItemProvider(
    'json',
    {
      triggerCharacters: ['"', ':'],
      provideCompletionItems(m, position, context) {
        if (m.uri.toString() !== targetUri) return { suggestions: [] }

        const line = m.getLineContent(position.lineNumber)
        const linePrefix = line.slice(0, Math.max(0, position.column - 1))

        const textBefore = m.getValueInRange({
          startLineNumber: 1,
          startColumn: 1,
          endLineNumber: position.lineNumber,
          endColumn: position.column,
        })

        const ctxStack = computeJsonContext(textBefore)
        const { depth } = getDepthOutsideStrings(textBefore)
        const currentSchemaNode = resolveSchemaForContext(schemaRoot, ctxStack)

        const word = m.getWordUntilPosition(position)
        const range = new monaco.Range(
          position.lineNumber,
          word.startColumn,
          position.lineNumber,
          word.endColumn,
        )

        const suggestions: monaco.languages.CompletionItem[] = []
        const keysHere = schemaProperties(currentSchemaNode)

        const trimmed = textBefore.replace(/\s+$/g, '')
        const lastNonSpace =
          trimmed.length > 0 ? trimmed[trimmed.length - 1] : ''

        const isLikelyPropertyName =
          /(^\s*\"[^\"]*$)/.test(linePrefix) ||
          /([,{]\s*\"[^\"]*$)/.test(linePrefix)
        const isAtPropertyInsertionPoint =
          (lastNonSpace === '{' || lastNonSpace === ',') &&
          !/\"[^\"]*$/.test(linePrefix)

        const activeValueKey = (() => {
          for (let i = ctxStack.length - 1; i >= 0; i--) {
            const c = ctxStack[i]
            if (c.kind === 'object' && c.activeValueKey) return c.activeValueKey
          }
          return null
        })()

        const isInvoke =
          context?.triggerKind === monaco.languages.CompletionTriggerKind.Invoke

        if (
          keysHere.length > 0 &&
          (isLikelyPropertyName || (isInvoke && isAtPropertyInsertionPoint))
        ) {
          let idx = 0
          for (const k of keysHere) {
            const insertText =
              isInvoke && isAtPropertyInsertionPoint ? `"${k}": $0` : k
            suggestions.push({
              label: k,
              kind: monaco.languages.CompletionItemKind.Property,
              insertText,
              insertTextRules:
                isInvoke && isAtPropertyInsertionPoint
                  ? monaco.languages.CompletionItemInsertTextRule
                      .InsertAsSnippet
                  : undefined,
              range,
              sortText: `0_${idx.toString().padStart(3, '0')}_${k}`,
            })
            idx++
          }
          return { suggestions }
        }

        if (activeValueKey && currentSchemaNode?.properties?.[activeValueKey]) {
          const valueSchema = resolveSchemaNode(
            schemaRoot,
            currentSchemaNode.properties[activeValueKey],
          )
          const enumVals = schemaEnum(valueSchema)
          const hasOpeningQuote = new RegExp(
            `"${activeValueKey}"\\s*:\\s*\\"[^\\"]*$`,
          ).test(textBefore)

          if (enumVals.length > 0) {
            let idx = 0
            for (const v of enumVals) {
              suggestions.push({
                label: v,
                kind: monaco.languages.CompletionItemKind.Value,
                insertText: hasOpeningQuote ? v : `"${v}"`,
                range,
                sortText: `0_${idx.toString().padStart(3, '0')}_${v}`,
              })
              idx++
            }
            return { suggestions }
          }

          const t = valueSchema?.type
          if (t === 'object') {
            suggestions.push({
              label: '{}',
              kind: monaco.languages.CompletionItemKind.Snippet,
              insertText: `{ $0 }`,
              insertTextRules:
                monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
              range,
              sortText: '0_000_{}',
            })
            return { suggestions }
          }
          if (t === 'array') {
            suggestions.push({
              label: '[]',
              kind: monaco.languages.CompletionItemKind.Snippet,
              insertText: `[ $0 ]`,
              insertTextRules:
                monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
              range,
              sortText: '0_000_[]',
            })
            return { suggestions }
          }
        }

        return { suggestions }
      },
    },
  )
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

      // Check context: are we inside a JSON object/array?
      // Heuristic: scan backwards for unclosed { or [
      const textBefore = model.getValueInRange({
        startLineNumber: 1,
        startColumn: 1,
        endLineNumber: position.lineNumber,
        endColumn: position.column,
      })

      const { depth } = getDepthOutsideStrings(textBefore)
      const isInside = depth > 0

      const suggestions = scrapingRuleSnippets
        .filter(snippet => {
          if (snippet.label.startsWith('field-')) {
            return isInside && depth >= 2
          }
          return true
        })
        .map(snippet => ({
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

  applyJsonSchemaOptions()

  editor = monaco.editor.create(editorContainer.value!, {
    model,
    theme: props.theme,
    automaticLayout: false, // Disabled for performance, using ResizeObserver instead
    minimap: { enabled: false },
    readOnly: props.readOnly,
    scrollBeyondLastLine: false,
    fixedOverflowWidgets: true,
    suggestOnTriggerCharacters: true,
    quickSuggestions: {
      other: true,
      comments: false,
      strings: true,
    },
    acceptSuggestionOnEnter: 'smart',
    tabCompletion: 'off',
  })

  // Resize Observer Implementation
  resizeObserver = new ResizeObserver(() => {
    editor?.layout()
  })
  if (editorContainer.value) {
    resizeObserver.observe(editorContainer.value)
  }

  registerSchemaCompletionFallback()

  if (language === 'json') {
    registerJsonSnippets()
  }

  // Common validation logic
  // Debounce validation to avoid excessive parsing on large files
  const performValidation = useDebounceFn(() => {
    if (!model) return

    // 1. Check for empty/whitespace content
    const value = editor?.getValue() || ''
    if (!value.trim()) {
      emit('validate', false)
      return
    }

    // 2. Immediate syntax check for JSON to prevent race condition
    if (props.language === 'json') {
      try {
        JSON.parse(value)
      } catch (e) {
        emit('validate', false)
        return
      }
    }

    // 2.1 Run custom validator if provided
    if (props.customValidator) {
      const customMarkers = props.customValidator(value)
      monaco.editor.setModelMarkers(model, 'custom-validator', customMarkers)
    }

    // 3. Check for Monaco markers (schema errors or other language errors)
    const markers = monaco.editor.getModelMarkers({ resource: model.uri })
    // Treat both Errors and Warnings as validation failures
    // Missing required fields often appear as Warnings in Monaco JSON
    const hasErrors = markers.some(
      marker =>
        marker.severity === monaco.MarkerSeverity.Error ||
        marker.severity === monaco.MarkerSeverity.Warning,
    )

    emit('validate', !hasErrors)
  }, 300) // 300ms debounce delay

  modelChangeDisposable = editor.onDidChangeModelContent(() => {
    const value = editor!.getValue()
    emit('update:modelValue', value)
    emit('change', value)
    performValidation() // Validate immediately on content change (catches empty string)
  })

  // Listen for marker changes (validation errors)
  monaco.editor.onDidChangeMarkers(() => {
    performValidation() // Re-validate when markers update
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
  () => {
    applyJsonSchemaOptions()
    registerSchemaCompletionFallback()
  },
  { deep: true },
)

onBeforeUnmount(() => {
  resizeObserver?.disconnect()
  modelChangeDisposable?.dispose()
  schemaCompletionDisposable?.dispose()
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
