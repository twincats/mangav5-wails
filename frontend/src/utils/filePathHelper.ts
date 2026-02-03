import { getMangaDirectory } from './configHelper'
export function safeWindowsDirectoryName(input: string, maxLen = 120): string {
  const reserved = /^(con|prn|aux|nul|com[1-9]|lpt[1-9])$/i
  const rep = '_' // Gantilah dengan karakter pengganti yang sesuai jika perlu

  let s = (input ?? '').normalize('NFKC')

  // Remove control characters and normalize whitespace
  s = s
    .replace(/[\u0000-\u001F\u007F]/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()

  // Replace Windows illegal characters (\/:*?"<>|)
  s = s.replace(/[\/\\:*?"<>|]/g, rep)

  // Remove leading/trailing spaces and dots
  s = s.replace(/^[.\s]+/, '').replace(/[.\s]+$/, '')

  // Collapse multiple occurrences of the separator
  s = s.replace(/_+/g, rep)

  // Handle empty name and reserved Windows names
  if (!s) s = 'untitled'
  if (reserved.test(s)) s = `${s}${rep}dir`

  // Ensure length limit
  if (s.length > maxLen) s = s.slice(0, maxLen).replace(/[.\s-]+$/, '')

  return s
}

export function ImagePath(title: string): string {
  return `/filemanga/${safeWindowsDirectoryName(title)}`
}

export async function getDownloadDir(
  title: string = 'untitled',
  chapter: string | number = '000',
): Promise<string> {
  const downloadDir = await getMangaDirectory()
  return `${downloadDir}/${safeWindowsDirectoryName(title)}/${chapter}`
}

export async function getDownloadMangaDir(
  title: string = 'untitled',
): Promise<string> {
  const downloadDir = await getMangaDirectory()
  return `${downloadDir}/${safeWindowsDirectoryName(title)}`
}
