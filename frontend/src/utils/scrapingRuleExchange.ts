export interface ScrapingRuleExchangeFile {
  format_version: 1
  kind: 'scraping_rule'
  site_key: string
  name: string
  enabled: boolean
  domains: string[]
  manga_rule: Record<string, any>
  chapter_rule: Record<string, any>
}

export interface ScrapingRuleEditorState {
  site_key: string
  name: string
  domains_json: string
  enabled: number
  manga_rule_json: string
  chapter_rule_json: string
}

const EXCHANGE_KIND = 'scraping_rule'
const EXCHANGE_VERSION = 1 as const

function parseJSON(value: string, label: string): any {
  try {
    return JSON.parse(value)
  } catch {
    throw new Error(`${label} bukan JSON yang valid`)
  }
}

function ensureObject(value: unknown, label: string): Record<string, any> {
  if (!value || typeof value !== 'object' || Array.isArray(value)) {
    throw new Error(`${label} harus berupa object JSON`)
  }
  return value as Record<string, any>
}

function ensureNonEmptyString(value: unknown, label: string): string {
  if (typeof value !== 'string' || !value.trim()) {
    throw new Error(`${label} harus berupa string yang tidak kosong`)
  }
  return value.trim()
}

function ensureDomains(value: unknown): string[] {
  if (!Array.isArray(value) || value.length === 0) {
    throw new Error('domains harus berupa array string dan minimal berisi 1 domain')
  }

  const domains = value.map(domain => {
    if (typeof domain !== 'string' || !domain.trim()) {
      throw new Error('domains harus berupa array string dan minimal berisi 1 domain')
    }
    return domain.trim()
  })

  return Array.from(new Set(domains))
}

function normalizeMangaRuleMetadata(
  mangaRule: Record<string, any>,
  name: string,
  domains: string[],
) {
  return {
    ...mangaRule,
    site: name,
    domains,
  }
}

export function buildScrapingRuleExchange(
  rule: ScrapingRuleEditorState,
): ScrapingRuleExchangeFile {
  const siteKey = ensureNonEmptyString(rule.site_key, 'site_key')
  const name = ensureNonEmptyString(rule.name, 'name')
  const domains = ensureDomains(parseJSON(rule.domains_json, 'domains_json'))
  const mangaRule = ensureObject(
    parseJSON(rule.manga_rule_json, 'manga_rule_json'),
    'manga_rule',
  )
  const chapterRule = ensureObject(
    parseJSON(rule.chapter_rule_json, 'chapter_rule_json'),
    'chapter_rule',
  )

  return {
    format_version: EXCHANGE_VERSION,
    kind: EXCHANGE_KIND,
    site_key: siteKey,
    name,
    enabled: rule.enabled === 1,
    domains,
    manga_rule: normalizeMangaRuleMetadata(mangaRule, name, domains),
    chapter_rule: chapterRule,
  }
}

export function stringifyScrapingRuleExchange(
  rule: ScrapingRuleEditorState,
): string {
  return JSON.stringify(buildScrapingRuleExchange(rule), null, 2)
}

export function parseScrapingRuleExchange(
  content: string,
): ScrapingRuleEditorState {
  const parsed = ensureObject(parseJSON(content, 'File import'), 'File import')

  const formatVersion = parsed.format_version
  if (formatVersion !== EXCHANGE_VERSION) {
    throw new Error(
      `format_version ${String(formatVersion)} belum didukung oleh aplikasi ini`,
    )
  }

  if (parsed.kind !== EXCHANGE_KIND) {
    throw new Error('kind file import tidak dikenali')
  }

  const siteKey = ensureNonEmptyString(parsed.site_key, 'site_key')
  const name = ensureNonEmptyString(parsed.name, 'name')
  const domains = ensureDomains(parsed.domains)
  const mangaRule = normalizeMangaRuleMetadata(
    ensureObject(parsed.manga_rule, 'manga_rule'),
    name,
    domains,
  )
  const chapterRule = ensureObject(parsed.chapter_rule, 'chapter_rule')

  return {
    site_key: siteKey,
    name,
    domains_json: JSON.stringify(domains),
    enabled: parsed.enabled ? 1 : 0,
    manga_rule_json: JSON.stringify(mangaRule, null, 2),
    chapter_rule_json: JSON.stringify(chapterRule, null, 2),
  }
}
