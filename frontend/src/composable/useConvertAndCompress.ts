import { getMangaDirectory } from '@/utils/configHelper'
import {
  DatabaseService,
  FileService,
  ImageService,
} from 'bindings/mangav5/services'
import type { Chapter } from 'bindings/mangav5/internal/models'
import type {
  WebPConvertOptions,
  WebPBatchResult,
} from 'bindings/mangav5/services'

export interface ConvertAndCompressOptions {
  manga_id: number
  manga_title: string
  quality: number
  resize: number
  delete: boolean
  compress: boolean
  status_read_only: boolean
  status_resize: boolean
}

export type ConvertAndCompressStep =
  | 'chapter_start'
  | 'scan_images'
  | 'converting'
  | 'compressing'
  | 'skip_already_compressed'
  | 'skip_read_only'
  | 'skip_no_images'
  | 'convert_skip_already_webp'
  | 'convert_success'
  | 'convert_failed'
  | 'compress_success'
  | 'compress_failed'
  | 'chapter_done'

export interface ConvertAndCompressLogItem {
  chapter_id: number
  chapter_number: number
  step: ConvertAndCompressStep
  message: string
}

export interface ConvertAndCompressResult {
  total: number
  selected: number
  success: number
  failed: number
  skipped: number
  logs: ConvertAndCompressLogItem[]
}

export interface ConvertAndCompressRuntimeState {
  total: number
  selected: number
  success: number
  failed: number
  skipped: number
  processed: number
  current_chapter_id: number | null
  current_chapter_number: number | null
  current_step: ConvertAndCompressStep | null
}

export interface ConvertAndCompressHooks {
  onState?: (state: ConvertAndCompressRuntimeState) => void
  onLog?: (
    log: ConvertAndCompressLogItem,
    state: ConvertAndCompressRuntimeState,
  ) => void
}

export const useConvertAndCompress = async (
  options: ConvertAndCompressOptions,
  hooks: ConvertAndCompressHooks = {},
): Promise<ConvertAndCompressResult> => {
  const chapterList = (await DatabaseService.GetChaptersByMangaID(
    options.manga_id,
  )) as unknown as Chapter[]

  const mangaDir = (await getMangaDirectory())?.trim() || ''
  if (!mangaDir) {
    throw new Error('Manga directory is not configured')
  }

  const logs: ConvertAndCompressLogItem[] = []

  const plannedSelected = chapterList.filter(ch => {
    if (isChapterCompressed(ch)) return false
    if (options.status_read_only && !isChapterRead(ch)) return false
    return true
  }).length

  let selected = plannedSelected
  let success = 0
  let failed = 0
  let skipped = 0
  let processed = 0

  const state: ConvertAndCompressRuntimeState = {
    total: chapterList.length,
    selected,
    success,
    failed,
    skipped,
    processed,
    current_chapter_id: null,
    current_chapter_number: null,
    current_step: null,
  }

  const emitState = () => {
    hooks.onState?.({ ...state })
  }

  const pushLog = (item: ConvertAndCompressLogItem) => {
    logs.push(item)
    hooks.onLog?.(item, { ...state })
  }

  const setCurrent = (chapter: Chapter, step: ConvertAndCompressStep) => {
    state.current_chapter_id = Number(chapter.id)
    state.current_chapter_number = Number(chapter.chapter_number)
    state.current_step = step
    emitState()
  }

  const markProcessed = () => {
    processed++
    state.processed = processed
    state.success = success
    state.failed = failed
    state.skipped = skipped
    emitState()
  }

  emitState()

  for (const chapter of chapterList) {
    const chapterNumber = Number(chapter.chapter_number)
    const fullChapterPath = `${mangaDir}/${chapter.path}`

    setCurrent(chapter, 'chapter_start')

    if (isChapterCompressed(chapter)) {
      skipped++
      state.skipped = skipped
      pushLog({
        chapter_id: Number(chapter.id),
        chapter_number: chapterNumber,
        step: 'skip_already_compressed',
        message: 'Skip: already compressed',
      })
      markProcessed()
      continue
    }

    if (options.status_read_only && !isChapterRead(chapter)) {
      skipped++
      state.skipped = skipped
      pushLog({
        chapter_id: Number(chapter.id),
        chapter_number: chapterNumber,
        step: 'skip_read_only',
        message: 'Skip: read-only mode (chapter not read)',
      })
      markProcessed()
      continue
    }

    setCurrent(chapter, 'chapter_start')

    setCurrent(chapter, 'scan_images')
    const images = await FileService.GetImageList(chapter.path)
    if (!Array.isArray(images) || images.length === 0) {
      skipped++
      state.skipped = skipped
      pushLog({
        chapter_id: Number(chapter.id),
        chapter_number: chapterNumber,
        step: 'skip_no_images',
        message: 'Skip: no images found',
      })
      markProcessed()
      continue
    }

    const allImagesAreWebp = images.every(isWebpPath)
    if (allImagesAreWebp) {
      setCurrent(chapter, 'convert_skip_already_webp')
      pushLog({
        chapter_id: Number(chapter.id),
        chapter_number: chapterNumber,
        step: 'convert_skip_already_webp',
        message: 'Convert: skipped (already webp)',
      })
    } else {
      const inputPaths = images.map(name => `${fullChapterPath}/${name}`)
      try {
        setCurrent(chapter, 'converting')
        const results = (await ImageService.ConvertBatchToWebP(
          inputPaths,
          '',
          toWebPConvertOptions(options),
        )) as unknown as WebPBatchResult[]

        const errorCount = results.filter(r => (r as any)?.error).length
        if (errorCount > 0) {
          failed++
          state.failed = failed
          setCurrent(chapter, 'convert_failed')
          pushLog({
            chapter_id: Number(chapter.id),
            chapter_number: chapterNumber,
            step: 'convert_failed',
            message: `Convert: failed (${errorCount}/${results.length})`,
          })
          markProcessed()
          continue
        }

        setCurrent(chapter, 'convert_success')
        pushLog({
          chapter_id: Number(chapter.id),
          chapter_number: chapterNumber,
          step: 'convert_success',
          message: `Convert: success (${results.length})`,
        })
      } catch (error) {
        failed++
        state.failed = failed
        setCurrent(chapter, 'convert_failed')
        pushLog({
          chapter_id: Number(chapter.id),
          chapter_number: chapterNumber,
          step: 'convert_failed',
          message: `Convert: failed (${formatError(error)})`,
        })
        markProcessed()
        continue
      }
    }

    if (options.compress) {
      try {
        setCurrent(chapter, 'compressing')
        await FileService.ConvertToCbz(fullChapterPath)
        chapter.is_compressed = 1
        await DatabaseService.UpdateChapter(chapter as any)
        setCurrent(chapter, 'compress_success')
        pushLog({
          chapter_id: Number(chapter.id),
          chapter_number: chapterNumber,
          step: 'compress_success',
          message: 'Compress: success',
        })
      } catch (error) {
        failed++
        state.failed = failed
        setCurrent(chapter, 'compress_failed')
        pushLog({
          chapter_id: Number(chapter.id),
          chapter_number: chapterNumber,
          step: 'compress_failed',
          message: `Compress: failed (${formatError(error)})`,
        })
        markProcessed()
        continue
      }
    }

    success++
    state.success = success
    setCurrent(chapter, 'chapter_done')
    pushLog({
      chapter_id: Number(chapter.id),
      chapter_number: chapterNumber,
      step: 'chapter_done',
      message: 'Done',
    })
    markProcessed()
  }

  state.current_chapter_id = null
  state.current_chapter_number = null
  state.current_step = null
  emitState()

  return {
    total: chapterList.length,
    selected,
    success,
    failed,
    skipped,
    logs,
  }
}

function isWebpPath(path: string): boolean {
  const cleanPath = path.split('?')[0].split('#')[0]
  return cleanPath.toLowerCase().endsWith('.webp')
}

function isChapterCompressed(chapter: Chapter): boolean {
  return Number(chapter.is_compressed) === 1
}

function isChapterRead(chapter: Chapter): boolean {
  return Number(chapter.status_read) === 1
}

function toWebPConvertOptions(
  options: ConvertAndCompressOptions,
): WebPConvertOptions {
  const resizeWidth =
    options.status_resize && Number(options.resize) > 0
      ? Number(options.resize)
      : 0
  return {
    quality: clamp(Number(options.quality), 1, 100),
    lossless: false,
    effort: 4,
    smartSubsample: true,
    resizeWidth,
    resizeHeight: 0,
    allowUpscale: false,
    deleteSource: Boolean(options.delete),
    concurrency: 0,
    overwrite: false,
    skipExisting: true,
    stopOnError: true,
  }
}

function clamp(n: number, min: number, max: number): number {
  if (!Number.isFinite(n)) return min
  return Math.max(min, Math.min(max, n))
}

function formatError(err: unknown): string {
  if (!err) return 'Unknown error'
  if (typeof err === 'string') return err
  if (err instanceof Error) return err.message || 'Unknown error'
  try {
    return JSON.stringify(err)
  } catch {
    return 'Unknown error'
  }
}
