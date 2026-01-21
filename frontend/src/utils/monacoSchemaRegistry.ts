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
    schemas: Array.from(registeredSchemas.values()),
  })
}
