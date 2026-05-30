<template>
  <n-spin :show="scrapeLoading" description="Scraping...">
    <div>
      <div class="flex gap-2 items-center mb-2">
        <n-input-group>
          <n-input
            v-model:value="downloadUrl"
            placeholder="Enter download URL"
            :disabled="isBusy"
          />
          <n-button
            tertiary
            type="primary"
            :loading="scrapeLoading"
            :disabled="downloadingSingle || downloadingMultiple"
            @click="fetchScrapeManga"
          >
            GO
          </n-button>
        </n-input-group>
        <n-button
          type="primary"
          secondary
          :disabled="isBusy"
          @click="clearDownloadInput"
        >
          <template #icon>
            <n-icon><CancelRound /></n-icon>
          </template>
        </n-button>
        <n-button
          type="primary"
          secondary
          :loading="downloadingMultiple"
          :disabled="
            checkedRowKeysRef.length === 0 ||
            scrapeLoading ||
            downloadingSingle ||
            downloadingMultiple
          "
          @click="downloadMultiple"
        >
          <template #icon>
            <n-icon><DownloadFilled /></n-icon>
          </template>
        </n-button>
      </div>
      <n-scrollbar style="max-height: calc(100vh - 160px)">
        <div class="bg-dark-400 rounded-md p-2 mb-2">
          <div class="text-sm font-medium">Select site rule:</div>
          <div class="bg-dark-500 p-2 rounded-md mt-1 overflow-auto">
            <div class="min-h-[52.72px] flex items-center">
              <n-radio-group v-model:value="selectedSiteKey" name="radiogroup">
                <n-space>
                  <n-radio
                    v-for="site in listScrapeRule"
                    :key="site.id"
                    :value="site.site_key"
                    :label="site.name"
                  />
                </n-space>
              </n-radio-group>
            </div>
          </div>
        </div>
        <div class="bg-dark-400 rounded-md p-2 my-2 min-h-[100px]">
          <n-h4 align-text>
            <n-text type="primary">
              {{ mangaData?.title }}
            </n-text>
          </n-h4>
          <div v-if="selectedChapters.length > 0">
            Download Chapters : {{ selectedChapters.length }} Chapter<br />
            Selected Chapters : {{ selectedChapters.join(', ') }}
          </div>
        </div>
        <div>
          <n-data-table
            :columns="columns"
            :bordered="false"
            :single-line="false"
            :data="chapterData"
            :row-key="rowKey"
            :size="'small'"
            :pagination="{
              pageSize: 10,
            }"
            striped
            v-model:checked-row-keys="checkedRowKeysRef"
            :row-props="rowProps"
          />
        </div>
      </n-scrollbar>
      <!-- modal progress -->
      <n-modal
        v-model:show="progressModal"
        :mask-closable="false"
        preset="card"
        class="w-[600px]"
        title="Download Progress"
      >
        <div class="w-full text-center mb-2">
          <div class="mb-2" v-if="checkedRowKeysRef.length > 0">
            Manga Chapter
            <n-progress
              type="line"
              :percentage="progress.chapterPercentage"
              indicator-placement="inside"
              processing
              :border-radius="4"
            />
            {{ progress.indexChapter }} / {{ progress.totalChapter }}
          </div>
          <div>
            Chapter Pages
            <n-progress
              type="line"
              status="success"
              :percentage="progress.downloadPercentage"
              indicator-placement="inside"
              processing
              :border-radius="4"
            />
            {{ progress.indexPage }} / {{ progress.totalPages }}
          </div>
        </div>
      </n-modal>
    </div>
  </n-spin>
</template>

<script setup lang="ts">
import { DownloadFilled, CancelRound } from '@vicons/material'
import FlagIndonesia from '@/assets/icon/twemoji--flag-indonesia.svg'
import FlagUK from '@/assets/icon/twemoji--flag-united-kingdom.svg'
import {
  DatabaseService,
  ScraperService,
  DownloadService,
} from '../../bindings/mangav5/services'
import {
  ScrapingRule,
  Manga,
  Chapter,
} from '../../bindings/mangav5/internal/models'
import { DownloadProgress } from '@/type/download'
import { MangaData, ChapterData, ChapterPages } from '@/type/scrape'
import { NButton, NIcon, NTag } from 'naive-ui'
import type { DataTableColumns, DataTableRowKey } from 'naive-ui'
import { Window, Events, Clipboard } from '@wailsio/runtime'
import { watchDebounced, useEventListener } from '@vueuse/core'
import {
  getDownloadDir,
  getDownloadMangaDir,
  safeWindowsDirectoryName,
} from '@/utils/filePathHelper'

const listScrapeRuleDb = ref<ScrapingRule[]>([])
const loadListScrapeRuleDb = async () => {
  listScrapeRuleDb.value = await DatabaseService.ListScrapingRulesBasic()
}
const listScrapeRule = computed(() => {
  return listScrapeRuleDb.value.filter(rule => rule.enabled === 1)
})
loadListScrapeRuleDb()

const message = useMessage()
const selectedSiteKey = ref<string>('')
const downloadUrl = ref<string>('')

const mangaData = ref<MangaData | null>(null)
const chapterData = computed<ChapterData[]>(() => {
  return mangaData.value?.chapters || []
})

// auto Maximize window disabled => broken minimum size
const wasMaximizedBefore = ref(false)
Window.IsMaximised().then((isMax: boolean) => {
  wasMaximizedBefore.value = isMax
  if (!isMax) {
    // Window.Maximise()
  }
})

const progressModal = ref(false)
const progress = reactive({
  downloadPercentage: 0,
  chapterPercentage: 0,
  indexPage: 0,
  totalPages: 0,
  indexChapter: 0,
  totalChapter: 0,
})

const scrapeLoading = ref(false)
const downloadingSingle = ref(false)
const downloadingMultiple = ref(false)
const downloadingChapterId = ref<string | null>(null)
const isBusy = computed(
  () =>
    scrapeLoading.value || downloadingSingle.value || downloadingMultiple.value,
)

const resetPageProgress = () => {
  progress.indexPage = 0
  progress.totalPages = 0
  progress.downloadPercentage = 0
}

const resetChapterProgress = (totalChapter: number) => {
  progress.indexChapter = 0
  progress.totalChapter = totalChapter
  progress.chapterPercentage = 0
}

const formatErrorMessage = (err: unknown): string => {
  if (!err) return 'Unknown error'
  if (typeof err === 'string') return err
  if (err instanceof Error) return err.message || 'Unknown error'
  try {
    return JSON.stringify(err)
  } catch {
    return 'Unknown error'
  }
}

const mangaIdDb = ref<number>(0)
const downloadOneChapter = async (chapterId: string): Promise<boolean> => {
  const chapterInfo = findChapterByChapterId(chapterId)
  try {
    // fetch rule data
    const rule = await getScrapeRule(selectedSiteKey.value)
    if (!rule) {
      return false
    }
    const chapterRule = JSON.parse(rule.chapter_rule_json)
    // scrape chapter list images
    const chapterImages = (await ScraperService.Scrape(
      chapterRule,
      chapterId,
    )) as unknown as ChapterPages

    const listImages = chapterImages.pages
    const outputDir = await getDownloadDir(
      mangaData.value?.title || 'untitled',
      chapterInfo?.chapter || '000',
    )
    const chapterPath = `${safeWindowsDirectoryName(mangaData.value?.title || 'untitled')}/${chapterInfo?.chapter || '000'}`

    // download chapter images
    await DownloadService.DownloadImages(listImages, outputDir, null)
    let StatusDownloadDover: number | boolean = 0
    // save manga only if manga_title is not saved
    let isNewManga = false
    if (mangaData.value) {
      const m = new Manga()
      m.main_title = mangaData.value.title
      // SaveManga returns [id, isNew]
      const result = await DatabaseService.SaveManga(m)
      if (Array.isArray(result)) {
        mangaIdDb.value = result[0]
        isNewManga = result[1]
      } else {
        // Fallback incase of single return
        mangaIdDb.value = result as unknown as number
        // If single return, we assume we don't know if it's new, so maybe default to false or check logic
        // But with our backend change, it should be an array.
      }
    }

    // Download cover only if it's a NEW manga
    if (mangaData.value && isNewManga) {
      try {
        const mangaDir = await getDownloadMangaDir(mangaData.value.title)
        await DownloadService.DownloadImage(
          mangaData.value.cover,
          mangaDir,
          'cover',
          null,
        )
      } catch (err) {
        console.error('Failed to download cover:', err)
        // Non-blocking error for cover
      }
    }

    // save or update chapter (unique per manga_id + chapter_number)
    if (mangaIdDb.value !== 0 && chapterInfo) {
      const c = new Chapter()
      c.manga_id = mangaIdDb.value
      c.chapter_number = Number(chapterInfo.chapter)
      c.chapter_title = chapterInfo.chapter_title || ''
      c.volume = Number(chapterInfo.chapter_volume) || 0
      c.language = chapterInfo.language || 'en'
      c.translator_group = chapterInfo.group_name || 'unknown'
      c.path = chapterPath
      c.release_time_raw = chapterInfo.time || new Date().toISOString()
      c.is_compressed = 0

      try {
        const chapters = await DatabaseService.GetChaptersByMangaID(
          mangaIdDb.value,
        )
        const existing = chapters.find(
          ch => Number(ch.chapter_number) === c.chapter_number,
        )
        if (existing) {
          c.id = existing.id
          await DatabaseService.UpdateChapter(c)
        } else {
          await DatabaseService.CreateChapter(c)
        }
      } catch (err) {
        message.error('Failed to save chapter ' + chapterInfo?.chapter)
        console.error(err)
      }
    }

    message.success(`Chapter ${chapterInfo?.chapter} downloaded successfully`)
    return true
  } catch (error) {
    const chapLabel = chapterInfo?.chapter ? ` ${chapterInfo.chapter}` : ''
    message.error(
      `Failed to download chapter${chapLabel}: ${formatErrorMessage(error)}`,
    )
    console.error(error)
    return false
  }
}
// download progress event
Events.On('downloadProgress', event => {
  const data = event.data as DownloadProgress
  progress.indexPage = data.index
  progress.totalPages = data.total
  if (data.total > 0) {
    progress.downloadPercentage = Math.round((data.index / data.total) * 100)
  } else {
    progress.downloadPercentage = 0
  }
})

// download multiple chapters
const downloadMultiple = async () => {
  if (checkedRowKeysRef.value.length === 0) {
    message.error('Please select at least one chapter')
    return
  }
  if (downloadingSingle.value || downloadingMultiple.value) return

  downloadingMultiple.value = true
  progressModal.value = true
  resetPageProgress()
  resetChapterProgress(checkedRowKeysRef.value.length)

  const failedChapters: string[] = []
  try {
    for (const key of checkedRowKeysRef.value) {
      const chapterId = key as string
      resetPageProgress()
      const ok = await downloadOneChapter(chapterId)
      if (!ok) failedChapters.push(chapterId)

      progress.indexChapter++
      progress.chapterPercentage = Math.round(
        (progress.indexChapter / progress.totalChapter) * 100,
      )
    }
  } finally {
    downloadingMultiple.value = false
    setTimeout(() => {
      if (!downloadingSingle.value) progressModal.value = false
    }, 800)
  }

  if (failedChapters.length > 0) {
    const failedNumbers = failedChapters
      .map(id => findChapterByChapterId(id)?.chapter ?? id)
      .join(', ')
    message.error(`Some chapters failed: ${failedNumbers}`)
  } else {
    message.success('All chapters downloaded successfully')
  }
}

const fetchScrapeManga = async () => {
  if (scrapeLoading.value) return
  if (!downloadUrl.value) {
    const clip = await Clipboard.Text()
    if (clip && isValidUrl(clip)) {
      downloadUrl.value = clip.trim()
    } else {
      downloadUrl.value = ''
      message.error('Please enter a download URL')
      return
    }
  }
  downloadUrl.value = downloadUrl.value.trim()
  if (!isValidUrl(downloadUrl.value)) {
    downloadUrl.value = ''
    message.error('Please enter a download URL')
    return
  }
  if (!selectedSiteKey.value) {
    if (listScrapeRuleDb.value.length === 0) {
      await loadListScrapeRuleDb()
    }
    const siteKey = inferSiteKeyFromUrl(downloadUrl.value)
    if (siteKey) {
      selectedSiteKey.value = siteKey
    } else {
      message.error('Please select a site rule')
      return
    }
  }
  scrapeLoading.value = true
  try {
    const rule = await getScrapeRule(selectedSiteKey.value)
    if (!rule) {
      return
    }
    const mangaRule = JSON.parse(rule.manga_rule_json)
    const result = (await ScraperService.Scrape(
      mangaRule,
      downloadUrl.value,
    )) as unknown as MangaData
    // set chapter status
    // check if manga title is in db or not
    const [isExist, mangaId] = await DatabaseService.CheckStatusMangaByTitle(
      result.title,
    )
    const existingChapters = new Set<number>()
    if (isExist) {
      // get data chapters by mangaID
      const chap = await DatabaseService.GetChaptersByMangaID(mangaId)
      chap.forEach(c => existingChapters.add(c.chapter_number))
    }

    if (result.chapters) {
      result.chapters.forEach(c => {
        const cNum =
          typeof c.chapter === 'string' ? parseFloat(c.chapter) : c.chapter
        c.status = existingChapters.has(cNum)
      })
    }
    // const fe
    mangaData.value = result
  } catch (error) {
    message.error(`Failed to scrape manga: ${formatErrorMessage(error)}`)
    console.error(error)
  } finally {
    scrapeLoading.value = false
  }
}

const getScrapeRule = async (siteKey: string) => {
  try {
    return await DatabaseService.GetScrapingRule(siteKey)
  } catch (error) {
    message.error('Failed to get scrape rule')
    return null
  }
}

const clearDownloadInput = () => {
  downloadUrl.value = ''
  mangaData.value = null
  selectedSiteKey.value = ''
}

/* ======== TABLE FUNCTION ========== */
const rowKey = (record: ChapterData) => record.chapter_id
const checkedRowKeysRef = ref<DataTableRowKey[]>([])
const rowProps = (row: ChapterData) => {
  return {
    style: 'cursor: pointer;',
    onClick: (e: MouseEvent) => {
      if (
        (e.target as HTMLElement).closest('.n-checkbox') ||
        (e.target as HTMLElement).closest('.n-button')
      ) {
        return
      }
      const key = rowKey(row)
      const index = checkedRowKeysRef.value.indexOf(key)
      if (index > -1) {
        checkedRowKeysRef.value.splice(index, 1)
      } else {
        checkedRowKeysRef.value.push(key)
      }
    },
  }
}
function createColumns({
  downloadChapter,
}: {
  downloadChapter: (rowData: ChapterData) => Promise<void> | void
}): DataTableColumns<ChapterData> {
  return [
    {
      type: 'selection',
    },
    {
      title: 'Chapter ID',
      key: 'chapter_id',
      width: 120,
      ellipsis: true,
    },
    {
      title: 'Chapter',
      key: 'chapter',
      align: 'center',
      width: 80,
    },
    {
      title: 'Chapter Title',
      key: 'chapter_title',
      ellipsis: true,
    },
    {
      title: 'Chapter Volume',
      key: 'chapter_volume',
      align: 'center',
      width: 80,
    },
    {
      title: 'Group Name',
      key: 'group_name',
      align: 'center',
      ellipsis: true,
    },
    {
      title: 'Language',
      key: 'language',
      align: 'center',
      width: 80,
      render(row) {
        const iconStyle = { width: '1em', height: '1em', display: 'block' }
        if (row.language === 'id') {
          return h(
            NIcon,
            { size: '1.2em' },
            {
              default: () => h('img', { src: FlagIndonesia, style: iconStyle }),
            },
          )
        }
        if (row.language === 'en') {
          return h(
            NIcon,
            { size: '1.2em' },
            {
              default: () => h('img', { src: FlagUK, style: iconStyle }),
            },
          )
        }
        return row.language
      },
    },
    {
      title: 'Release Time',
      key: 'time',
      align: 'center',
      width: 120,
      ellipsis: true,
    },
    {
      title: 'Status',
      key: 'status',
      align: 'center',
      width: 80,
      ellipsis: true,
      render(row) {
        if (row.status === true) {
          return h(
            NTag,
            { type: 'success', size: 'small' },
            { default: () => 'YES' },
          )
        }
        return h(
          NTag,
          { type: 'error', size: 'small' },
          { default: () => 'No' },
        )
      },
    },
    {
      title: 'Action',
      key: 'actions',
      align: 'center',
      width: 100,
      render(row) {
        const rowLoading =
          downloadingSingle.value &&
          downloadingChapterId.value === row.chapter_id
        const disabled =
          checkedRowKeysRef.value.length > 0 ||
          scrapeLoading.value ||
          downloadingMultiple.value ||
          (downloadingSingle.value && !rowLoading)
        return h(
          NButton,
          {
            size: 'small',
            onClick: () => downloadChapter(row),
            disabled,
            loading: rowLoading,
          },
          { default: () => 'Download' },
        )
      },
    },
  ]
}

const columns = createColumns({
  async downloadChapter(rowData: ChapterData) {
    if (downloadingSingle.value || downloadingMultiple.value) return
    downloadingSingle.value = true
    downloadingChapterId.value = rowData.chapter_id
    progressModal.value = true
    resetPageProgress()
    resetChapterProgress(0)
    try {
      await downloadOneChapter(rowData.chapter_id)
    } finally {
      downloadingSingle.value = false
      downloadingChapterId.value = null
      setTimeout(() => {
        if (!downloadingMultiple.value) progressModal.value = false
      }, 800)
    }
  },
})

const selectedChapters = computed(() => {
  return checkedRowKeysRef.value
    .map(key => findChapterByChapterId(key.toString())!.chapter)
    .sort((a, b) => Number(a) - Number(b))
})
/* ======== WATCHER FUNCTION ========== */
watchDebounced(
  downloadUrl,
  newVal => {
    if (newVal) {
      try {
        if (!isValidUrl(newVal)) {
          return
        }
        const siteKey = inferSiteKeyFromUrl(newVal)
        if (siteKey && siteKey !== selectedSiteKey.value) {
          selectedSiteKey.value = siteKey
        }
      } catch (error) {}
    }
  },
  { debounce: 100 },
)

/* ======== ROUTER FUNCTION ========== */
onBeforeRouteLeave((_to, _from, next) => {
  if (!wasMaximizedBefore.value) {
    Window.Restore()
  }
  next()
})

/* ======== HELPER FUNCTION ========== */
useEventListener(document, 'paste', e => {
  const target = e.target as HTMLElement | null
  if (
    target &&
    (target.tagName === 'INPUT' ||
      target.tagName === 'TEXTAREA' ||
      target.isContentEditable)
  ) {
    return
  }

  const activeElement = document.activeElement as HTMLElement | null
  if (
    activeElement &&
    (activeElement.tagName === 'INPUT' ||
      activeElement.tagName === 'TEXTAREA' ||
      activeElement.isContentEditable)
  ) {
    return
  }

  const clipboardData = e.clipboardData
  if (clipboardData) {
    const text = clipboardData.getData('text')
    if (text && isValidUrl(text)) {
      downloadUrl.value = text.trim()
    }
  }
})

function isValidUrl(text: string): boolean {
  try {
    const trimmed = text.trim()
    if (!/^https?:\/\//i.test(trimmed)) {
      return false
    }
    new URL(trimmed)
    return true
  } catch {
    return false
  }
}

function inferSiteKeyFromUrl(text: string): string | null {
  const trimmed = text.trim()
  if (!isValidUrl(trimmed)) return null

  let hostname = ''
  try {
    hostname = new URL(trimmed).hostname
  } catch {
    return null
  }

  const matchedRule = listScrapeRuleDb.value.find(rule => {
    try {
      return JSON.parse(rule.domains_json).includes(hostname)
    } catch {
      return false
    }
  })

  return matchedRule?.site_key ?? null
}

const findChapterByChapterId = (chapterId: string) => {
  return chapterData.value.find(chap => chap.chapter_id === chapterId)
}
</script>
