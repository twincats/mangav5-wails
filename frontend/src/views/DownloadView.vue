<template>
  <div>
    <div class="flex gap-2 items-center">
      <n-input-group>
        <n-input v-model:value="downloadUrl" placeholder="Enter download URL" />
        <n-button tertiary type="primary" @click="fetchScrapeManga">
          GO
        </n-button>
      </n-input-group>
      <n-button type="primary" secondary @click="progressModal = true">
        <template #icon>
          <n-icon><CancelRound /></n-icon>
        </template>
      </n-button>
      <n-button
        type="primary"
        secondary
        :disabled="checkedRowKeysRef.length === 0"
        @click="downloadMultiple"
      >
        <template #icon>
          <n-icon><DownloadFilled /></n-icon>
        </template>
      </n-button>
    </div>
    <div class="bg-dark-400 rounded-md p-2 my-2">
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
import { Window, Events } from '@wailsio/runtime'
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

const wasMaximizedBefore = ref(false)
Window.IsMaximised().then((isMax: boolean) => {
  wasMaximizedBefore.value = isMax
  if (!isMax) {
    Window.Maximise()
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

const mangaIdDb = ref<number>(0)
const downloadOneChapter = async (chapterId: string) => {
  try {
    // fetch rule data
    const rule = await getScrapeRule(selectedSiteKey.value)
    if (!rule) {
      return
    }
    const chapterRule = JSON.parse(rule.chapter_rule_json)
    // scrape chapter list images
    const chapterImages = (await ScraperService.Scrape(
      chapterRule,
      chapterId,
    )) as unknown as ChapterPages

    const listImages = chapterImages.pages
    const chapterInfo = findChapterByChapterId(chapterId)
    const outputDir = await getDownloadDir(
      mangaData.value?.title || 'untitled',
      chapterInfo?.chapter || '000',
    )
    const chapterPath = `${safeWindowsDirectoryName(mangaData.value?.title || 'untitled')}/${chapterInfo?.chapter || '000'}`

    progressModal.value = true
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

    // save chapter
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
        await DatabaseService.CreateChapter(c)
      } catch (err) {
        message.error('Failed to save chapter ' + chapterInfo?.chapter)
        console.error(err)
      }
    }

    message.success(`Chapter ${chapterInfo?.chapter} downloaded successfully`)
  } catch (error) {
    message.error('Failed to download chapter')
    progressModal.value = false
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
  progressModal.value = true
  progress.indexChapter = 0
  progress.totalChapter = checkedRowKeysRef.value.length
  for (const key of checkedRowKeysRef.value) {
    const chapterId = key as string
    await downloadOneChapter(chapterId)
    progress.indexChapter++
    progress.chapterPercentage = Math.round(
      (progress.indexChapter / progress.totalChapter) * 100,
    )
  }
  setTimeout(() => {
    progressModal.value = false
  }, 2000)
}

const fetchScrapeManga = async () => {
  if (!selectedSiteKey.value) {
    message.error('Please select a site rule')
    return
  }
  if (!downloadUrl.value) {
    message.error('Please enter a download URL')
    return
  }
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
    mangaData.value = result
  } catch (error) {
    message.error('Failed to scrape manga')
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
        if (row.language === 'id') {
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
        return h(
          NButton,
          {
            size: 'small',
            onClick: () => downloadChapter(row),
            disabled: checkedRowKeysRef.value.length > 0,
          },
          { default: () => 'Download' },
        )
      },
    },
  ]
}

const columns = createColumns({
  async downloadChapter(rowData: ChapterData) {
    console.info(`chapter = ${rowData.chapter}`)
    downloadOneChapter(rowData.chapter_id)
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
        const url = new URL(newVal)
        console.info(`url.hostname = ${url.hostname}`)
        const siteKey = listScrapeRuleDb.value.find(rule =>
          JSON.parse(rule.domains_json).includes(url.hostname),
        )?.site_key
        if (siteKey) {
          selectedSiteKey.value = siteKey
        } else {
          message.error('Failed to fetch manga')
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
  const clipboardData = e.clipboardData
  if (clipboardData) {
    const text = clipboardData.getData('text')
    if (text) {
      downloadUrl.value = text
    }
  }
})

const findChapterByChapterId = (chapterId: string) => {
  return chapterData.value.find(chap => chap.chapter_id === chapterId)
}
</script>
