import * as monaco from 'monaco-editor'

export const validateMangaRule = (
  jsonString: string,
): monaco.editor.IMarkerData[] => {
  const markers: monaco.editor.IMarkerData[] = []
  try {
    const data = JSON.parse(jsonString)
    // Basic structure check handled by schema, we check specific required fields in extract
    if (data.extract && Array.isArray(data.extract)) {
      const requiredFields = ['title', 'cover', 'chapters']
      const missingFields = requiredFields.filter(
        req => !data.extract.some((e: any) => e.name === req),
      )

      if (missingFields.length > 0) {
        markers.push({
          message: `Missing required rules in 'extract': ${missingFields.join(
            ', ',
          )}`,
          severity: 8, // MarkerSeverity.Error
          startLineNumber: 1,
          startColumn: 1,
          endLineNumber: 1,
          endColumn: 1,
        })
      }
    }
  } catch {
    // Syntax errors handled by Monaco JSON worker
  }
  return markers
}

export const validateChapterRule = (
  jsonString: string,
): monaco.editor.IMarkerData[] => {
  const markers: monaco.editor.IMarkerData[] = []
  try {
    const data = JSON.parse(jsonString)
    if (data.extract && Array.isArray(data.extract)) {
      const requiredFields = ['pages']
      const missingFields = requiredFields.filter(
        req => !data.extract.some((e: any) => e.name === req),
      )

      if (missingFields.length > 0) {
        markers.push({
          message: `Missing required rules in 'extract': ${missingFields.join(
            ', ',
          )}`,
          severity: 8, // MarkerSeverity.Error
          startLineNumber: 1,
          startColumn: 1,
          endLineNumber: 1,
          endColumn: 1,
        })
      }
    }
  } catch {
    // Syntax errors handled by Monaco JSON worker
  }
  return markers
}

interface PageData {
  pages: string[]
}

export function isValidUrl(str: string): boolean {
  try {
    new URL(str)
    return true
  } catch {
    return false
  }
}

export function isValidPages(data: unknown): data is PageData {
  if (typeof data !== 'object' || data === null) return false

  const obj = data as Record<string, any>
  if (!('pages' in obj)) return false

  const pages = obj.pages
  if (!Array.isArray(pages)) return false

  return pages.every(
    (url: string) => typeof url === 'string' && isValidUrl(url),
  )
}
