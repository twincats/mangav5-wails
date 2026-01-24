<template>
  <div>
    <div class="mx-auto">
      <div class="flex gap-2">
        <n-input-group>
          <n-input v-model:value="url" placeholder="Manga Rule URL" />
          <n-button tertiary type="primary" @click="clickScrapeTest">
            GO
          </n-button>
        </n-input-group>
        <n-input-group>
          <n-input v-model:value="url" placeholder="Chapter Rule URL" />
          <n-button tertiary type="primary" @click="clickScrapeTest">
            GO
          </n-button>
        </n-input-group>
        <n-button
          tertiary
          type="primary"
          @click="clickDownloadTest"
          :disabled="!statusDownload"
        >
          Download
        </n-button>
        <n-button type="primary" @click="saveScrapingRules">
          Save Rules
        </n-button>
        <n-button type="primary"> Load Rules </n-button>
      </div>
    </div>
    <!-- second row -->
    <div class="grid grid-cols-10 mt-2">
      <div class="col-span-9 grid grid-cols-6 gap-2">
        <n-input type="text" placeholder="site_key" />
        <n-input type="text" placeholder="name" />
        <n-input type="text" placeholder="domains" />
        <div class="col-span-3 gap-2 grid grid-cols-4 items-center">
          <n-switch>
            <template #checked> enabled </template>
            <template #unchecked> disabled </template>
          </n-switch>
          <n-checkbox> manga_rule </n-checkbox>
          <n-checkbox> chapter_rule </n-checkbox>
          <n-switch>
            <template #checked> chapter_rule </template>
            <template #unchecked> manga_rule </template>
          </n-switch>
        </div>
      </div>
      <div class="flex justify-end gap-2">
        <n-button tertiary type="primary">
          <n-icon>
            <PlaylistAddFilled />
          </n-icon>
        </n-button>
        <n-button secondary type="primary">
          <n-icon>
            <PostAddFilled />
          </n-icon>
        </n-button>
      </div>
    </div>
    <!-- editor row -->
    {{ activeTab }}
    <div class="grid grid-cols-2 gap-2">
      <div>
        <n-tabs
          type="line"
          paneClass="h-[calc(100vh-280px)]"
          v-model:value="activeTab"
        >
          <n-tab-pane name="editor1" tab="Manga Rule">
            <div class="h-full">
              <MonacoEditor
                v-model="codeMangaRule"
                language="json"
                theme="vs-dark"
                :jsonSchema="SiteRuleSchema"
                :formatOnLoad="true"
              />
            </div>
          </n-tab-pane>
          <n-tab-pane name="editor2" tab="Chapter Rule">
            <div class="h-full">
              <MonacoEditor
                v-model="codeChapterRule"
                language="json"
                theme="vs-dark"
                :jsonSchema="SiteRuleSchema"
                :formatOnLoad="true"
              />
            </div>
          </n-tab-pane>
        </n-tabs>
      </div>
      <div>
        <n-tabs type="line" paneClass="h-[calc(100vh-280px)]">
          <n-tab-pane name="editor3" tab="Scrape Result">
            <div class="h-full">
              <MonacoEditor
                v-model="resultJson"
                language="json"
                theme="vs-dark"
                :formatOnLoad="true"
              />
            </div>
          </n-tab-pane>
        </n-tabs>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { PlaylistAddFilled, PostAddFilled } from '@vicons/material'
import SiteRuleSchema from '@/assets/SiteRuleSchema.json'
import {
  DownloadService,
  ScraperService,
} from '../../bindings/mangav5/services'
import { Events } from '@wailsio/runtime'
import { watchDebounced } from '@vueuse/core'
import { DatabaseService } from '../../bindings/mangav5/services'
import { ScrapingRule } from '../../bindings/mangav5/internal/models'

const codeMangaRule = ref('')
const codeChapterRule = ref('')
const resultJson = ref('')
const dialog = useDialog()

const activeTab = ref('editor1')

const clickDownloadTest = async () => {
  console.log('clickDownloadTest')
  if (!resultJson.value) {
    console.log('resultJson.value is empty')
    return
  }
  const urlImages: string[] = JSON.parse(resultJson.value).pages
  try {
    const res = await DownloadService.DownloadImages(
      urlImages,
      'D:/Tutorial/mangago/ikimen',
      null,
    )
    console.log(res)
  } catch (error) {
    console.log(error)
  }
}
Events.On('downloadProgress', data => {
  console.log('downloadProgress', data)
})

const url = ref('')
// scrape test
const clickScrapeTest = async () => {
  console.log('clickScrapeTest')
  if (!codeMangaRule.value) {
    console.log('code.value is empty')
    return
  }
  const rules = JSON.parse(codeMangaRule.value)
  try {
    const res = await ScraperService.Scrape(rules, url.value)
    resultJson.value = JSON.stringify(res, null, 2)
    console.log(res)
  } catch (error) {
    console.log(error)
    dialog.error({
      title: 'Error',
      content: `${error}`,
    })
  }
}

const scrapingRuleInput = reactive({
  site_key: '',
  name: '',
  domains: '',
  enabled: true,
})
/* ====== SAVE RULES ====== */
const saveScrapingRules = async () => {
  try {
    const ruleData = new ScrapingRule()
    ruleData.site_key = scrapingRuleInput.site_key
    ruleData.name = scrapingRuleInput.name
    ruleData.domains_json = JSON.stringify(scrapingRuleInput.domains)
    ruleData.enabled = scrapingRuleInput.enabled ? 1 : 0
    ruleData.chapter_rule_json =
      typeof codeChapterRule.value === 'string'
        ? codeChapterRule.value
        : JSON.stringify(codeChapterRule.value)
    ruleData.manga_rule_json =
      typeof codeMangaRule.value === 'string'
        ? codeMangaRule.value
        : JSON.stringify(codeMangaRule.value)
    await DatabaseService.SaveScrapingRule(ruleData)
  } catch (error) {
    console.log(error)
  }
}

const statusDownload = ref(false)
/* ====== WATCHERR ====== */
// watch resultJson and update statusDownload
watchDebounced(
  resultJson,
  v => {
    // exit if empty
    if (!v.trim()) {
      statusDownload.value = false
      return
    }
    // parse JSON
    try {
      const p = JSON.parse(v)
      if (isValidPages(p)) {
        statusDownload.value = true
      } else {
        statusDownload.value = false
      }
    } catch (error) {
      statusDownload.value = false
    }
  },
  { debounce: 500, maxWait: 1000 },
)

// watch codeMangaRule
watchDebounced(codeMangaRule, v => {
  // exit if empty
  if (!v.trim()) {
    return
  }
})

/* ====== HELPER FUNCTIONS ====== */
interface PageData {
  pages: string[]
}

function isValidUrl(str: string): boolean {
  try {
    new URL(str)
    return true
  } catch {
    return false
  }
}

function isValidPages(data: unknown): data is PageData {
  if (typeof data !== 'object' || data === null) return false

  const obj = data as Record<string, any>
  if (!('pages' in obj)) return false

  const pages = obj.pages
  if (!Array.isArray(pages)) return false

  return pages.every(
    (url: string) => typeof url === 'string' && isValidUrl(url),
  )
}
</script>
