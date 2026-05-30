import * as monaco from 'monaco-editor'

// Global state shared across all instances
const registeredSchemas = new Map<string, any>()

export function registerSchema(uri: string, fileMatch: string[], schema: any) {
  registeredSchemas.set(uri, {
    uri,
    fileMatch,
    schema,
  })
  updateJsonDiagnostics()
}

export function unregisterSchema(uri: string) {
  if (registeredSchemas.has(uri)) {
    registeredSchemas.delete(uri)
    updateJsonDiagnostics()
  }
}

function updateJsonDiagnostics() {
  const jsonDefaults = (monaco.languages as any).json?.jsonDefaults
  if (!jsonDefaults) return

  jsonDefaults.setDiagnosticsOptions({
    validate: true,
    enableSchemaRequest: false,
    allowComments: true,
    trailingCommas: 'ignore',
    schemas: Array.from(registeredSchemas.values()),
  })

  if (typeof jsonDefaults.setModeConfiguration === 'function') {
    jsonDefaults.setModeConfiguration({
      completionItems: true,
      diagnostics: true,
      documentFormattingEdits: true,
      documentRangeFormattingEdits: true,
      documentSymbols: true,
      colors: true,
      foldingRanges: true,
      hover: true,
      tokens: true,
    })
  }

  if (typeof jsonDefaults.setEagerModelSync === 'function') {
    jsonDefaults.setEagerModelSync(true)
  }
}
