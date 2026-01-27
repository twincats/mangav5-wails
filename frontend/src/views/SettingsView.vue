<template>
  <div>
    <div class="mx-auto">
      <div class="flex gap-2">
        <n-input-group>
          <n-input
            v-model:value="urlRule.url_manga_rule"
            placeholder="Manga Rule URL"
          />
          <n-button
            tertiary
            type="primary"
            @click="
              clickScrapeTest(
                urlRule.url_manga_rule,
                scrapingRuleInput.manga_rule_json,
              )
            "
          >
            GO
          </n-button>
        </n-input-group>
        <n-input-group>
          <n-input
            v-model:value="urlRule.url_chapter_rule"
            placeholder="Chapter Rule URL"
          />
          <n-button
            tertiary
            type="primary"
            @click="
              clickScrapeTest(
                urlRule.url_chapter_rule,
                scrapingRuleInput.chapter_rule_json,
              )
            "
          >
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
        <n-button
          type="primary"
          @click="saveScrapingRules"
          :disabled="!readyToSave"
        >
          Save Rules
        </n-button>
        <n-button type="primary" @click="openLoadModal"> Load Rules </n-button>
      </div>
    </div>
    <!-- second row -->
    <div class="grid grid-cols-10 mt-2">
      <div class="col-span-9 flex gap-2">
        <n-input
          type="text"
          placeholder="site_key"
          v-model:value="scrapingRuleInput.site_key"
        />
        <n-input
          type="text"
          placeholder="name"
          v-model:value="scrapingRuleInput.name"
        />
        <n-input
          type="text"
          placeholder="domains"
          v-model:value="scrapingRuleInput.domains_json"
        />
        <div class="flex gap-2 items-center">
          <n-switch
            v-model:value="scrapingRuleInput.enabled"
            :checked-value="1"
            :unchecked-value="0"
          >
            <template #checked> enabled </template>
            <template #unchecked> disabled </template>
          </n-switch>
          <div class="flex gap-2">
            <n-button type="primary" text :disabled="!statusJson.manga_rule">
              <template #icon>
                <n-icon>
                  <CheckCircleFilled v-if="statusJson.manga_rule" />
                  <CancelRound v-else />
                </n-icon>
              </template>
              Manga Rule
            </n-button>
            <n-button type="primary" text :disabled="!statusJson.chapter_rule">
              <template #icon>
                <n-icon>
                  <CheckCircleFilled v-if="statusJson.chapter_rule" />
                  <CancelRound v-else />
                </n-icon>
              </template>
              Chapter Rule
            </n-button>
          </div>
        </div>
      </div>
      <div class="flex justify-end gap-2">
        <n-button tertiary type="primary">
          <n-icon>
            <PlaylistAddFilled />
          </n-icon>
        </n-button>
        <n-button secondary type="primary" @click="clearInput">
          <n-icon>
            <ClearFilled />
          </n-icon>
        </n-button>
      </div>
    </div>
    <!-- third row -->
    <!-- editor row -->
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
                v-model="scrapingRuleInput.manga_rule_json"
                language="json"
                theme="vs-dark"
                :jsonSchema="MangaRuleSchema"
                :formatOnLoad="true"
                :customValidator="validateMangaRule"
                @validate="statusJson.manga_rule = $event"
              />
            </div>
          </n-tab-pane>
          <n-tab-pane name="editor2" tab="Chapter Rule">
            <div class="h-full">
              <MonacoEditor
                v-model="scrapingRuleInput.chapter_rule_json"
                language="json"
                theme="vs-dark"
                :jsonSchema="ChapterRuleSchema"
                :formatOnLoad="true"
                :customValidator="validateChapterRule"
                @validate="statusJson.chapter_rule = $event"
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
    <!-- modal for loading scraping rules -->
    <n-modal
      v-model:show="loadRuleModalVisible"
      preset="dialog"
      title="Load Scraping Rules"
      size="medium"
      transform-origin="center"
      :auto-focus="false"
      :icon="() => h(NIcon, null, { default: () => h(AssignmentOutlined) })"
    >
      <div>
        <n-scrollbar class="pr-4 max-h-300px">
          <n-list hoverable clickable>
            <n-list-item
              v-for="rule in listScrapeRuleDb"
              :key="rule.id"
              @click="loadRuleToInput(rule.site_key)"
            >
              {{ rule.site_key }} -
              {{ JSON.parse(rule.domains_json).join(', ') }}
            </n-list-item>
          </n-list>
        </n-scrollbar>
      </div>
      <template #footer> Footer </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import {
  PlaylistAddFilled,
  ClearFilled,
  CheckCircleFilled,
  CancelRound,
  AssignmentOutlined,
} from '@vicons/material'
import MangaRuleSchema from '@/assets/MangaRuleSchema.json'
import ChapterRuleSchema from '@/assets/ChapterRuleSchema.json'
import {
  DownloadService,
  ScraperService,
} from '../../bindings/mangav5/services'
import { Events } from '@wailsio/runtime'
import { watchDebounced } from '@vueuse/core'
import { DatabaseService } from '../../bindings/mangav5/services'
import { ScrapingRule } from '../../bindings/mangav5/internal/models'
import {
  validateMangaRule,
  validateChapterRule,
  isValidPages,
} from '../utils/validationHelpers'
import { NIcon } from 'naive-ui'
import { h } from 'vue'

const resultJson = ref('')
const dialog = useDialog()
const message = useMessage()
const activeTab = ref('editor1')
const statusJson = reactive({
  manga_rule: false,
  chapter_rule: false,
})
const urlRule = reactive({
  url_manga_rule: '',
  url_chapter_rule: '',
})
const loadRuleModalVisible = ref(false)
const listScrapeRuleDb = ref<ScrapingRule[]>([])
const openLoadModal = async () => {
  if (listScrapeRuleDb.value.length > 0) {
    loadRuleModalVisible.value = true
    return
  }
  listScrapeRuleDb.value = await DatabaseService.ListScrapingRulesBasic()
  loadRuleModalVisible.value = true
}

const loadRuleToInput = async (site_key: string) => {
  try {
    const rule = await DatabaseService.GetScrapingRule(site_key)
    if (!rule) {
      message.error('Rule not found')
      return
    }
    Object.assign(scrapingRuleInput, rule)
    loadRuleModalVisible.value = false

    // Manually trigger validation since editors might be inactive
    if (rule.manga_rule_json) {
      statusJson.manga_rule =
        validateMangaRule(rule.manga_rule_json).length === 0
    }
    if (rule.chapter_rule_json) {
      statusJson.chapter_rule =
        validateChapterRule(rule.chapter_rule_json).length === 0
    }
  } catch (error) {
    message.error(`${error}`)
  }
}

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

// scrape test
const clickScrapeTest = async (url: string, json_rule: string) => {
  if (!json_rule) {
    console.log('JSON Rule is empty')
    return
  }
  const rules = JSON.parse(json_rule)
  try {
    const res = await ScraperService.Scrape(rules, url)
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

const scrapingRuleDefault = {
  site_key: '',
  name: '',
  domains_json: '[]',
  enabled: 1,
  chapter_rule_json: '',
  manga_rule_json: '',
}
const scrapingRuleInput = reactive({ ...scrapingRuleDefault })

const clearInput = () => {
  // Reset fields to default
  Object.assign(scrapingRuleInput, scrapingRuleDefault)
  // Remove ID if it exists (from loaded rule)
  if ('id' in scrapingRuleInput) {
    delete (scrapingRuleInput as any).id
  }
  // Reset ID for validation status as well if needed
  statusJson.manga_rule = false
  statusJson.chapter_rule = false
}
/* ====== SAVE RULES ====== */
const saveScrapingRules = async () => {
  try {
    const scrapingRuleRaw = toRaw(scrapingRuleInput)
    const ruleData = new ScrapingRule(scrapingRuleRaw)
    await DatabaseService.SaveScrapingRule(ruleData)
    dialog.success({
      title: 'Success',
      content: 'Scraping Rule saved successfully',
      positiveText: 'OK',
      onPositiveClick: () => {
        clearInput()
      },
    })
  } catch (error) {
    console.log(error)
    message.error(`${error}`)
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
// removed manual validation watcher since we use @validate event now

watchDebounced(
  () => scrapingRuleInput.manga_rule_json,
  m => {
    if (!m) return
    try {
      const mm = JSON.parse(m)
      const obj = mm as Record<string, any>
      if ('domains' in obj && Array.isArray(obj.domains)) {
        scrapingRuleInput.domains_json = JSON.stringify(obj.domains)
      }
      if ('site' in obj) {
        scrapingRuleInput.name = obj.site
        scrapingRuleInput.site_key = obj.site.replace(/\s+/g, '').toLowerCase()
      }
    } catch (error) {
      console.log(error)
    }
  },
  { debounce: 500, maxWait: 1000 },
)

/* ====== COMPUTED VALUES ====== */
const readyToSave = computed(() => {
  return (
    scrapingRuleInput.site_key &&
    scrapingRuleInput.name &&
    scrapingRuleInput.domains_json &&
    scrapingRuleInput.chapter_rule_json &&
    scrapingRuleInput.manga_rule_json &&
    statusJson.manga_rule &&
    statusJson.chapter_rule
  )
})

/* ====== HELPER FUNCTIONS ====== */
</script>
