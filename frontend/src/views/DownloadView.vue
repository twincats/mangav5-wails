<template>
  <div>
    <div>Settings view</div>
    <n-space>
      <n-button type="primary" @click="scrapeFull">Scrap</n-button>
      <n-button type="info" @click="scrapeJsRender">Scrap JS Render</n-button>
      <n-button type="primary" @click="$router.push('/')">Goto Home</n-button>
      <n-input-group>
        <n-input v-model:value="url" />
        <n-button tertiary type="primary" @click="scrapeRules"> GO </n-button>
      </n-input-group>
    </n-space>

    <n-split
      direction="horizontal"
      class="h-[calc(100vh-190px)] mt-4"
      :max="0.75"
      :min="0.5"
    >
      <template #1>
        <div :style="{ height: '100%' }">
          <MonacoEditor
            v-model="code"
            language="json"
            theme="vs-dark"
            :jsonSchema="SiteRuleSchema"
            :formatOnLoad="true"
          />
        </div>
      </template>
      <template #2>
        <div :style="{ height: '100%' }">
          <MonacoEditor
            v-model="resultJson"
            language="json"
            theme="vs-dark"
            :formatOnLoad="true"
          />
        </div>
      </template>
    </n-split>
  </div>
</template>

<script setup lang="ts">
import MonacoEditor from '@/components/MonacoEditor.vue'
import { BrowserService, ScraperService } from '../../bindings/mangav5/services'
import * as SiteRuleSchemaModule from '@/assets/SiteRuleSchema.json'

const SiteRuleSchema =
  (SiteRuleSchemaModule as any).default || SiteRuleSchemaModule
console.log('Schema loaded:', SiteRuleSchema)

const code = ref('')
const resultJson = ref('')
// Contoh scraping
// async function scrape() {
//   try {
//     const result = await BrowserService.ScrapePage(
//       'https://example.com',
//       'h1', // Optional selector
//     )
//     console.log('Title:', result.title)
//     console.log('Images found:', result.images.length)
//   } catch (err) {
//     console.error(err)
//   }
// }

const dialog = useDialog()
const scrapeJsRender = async () => {
  try {
    const result = await BrowserService.ScrapePage(
      'https://westmanga.me/comic/level-count-stop-kara-hajimaru-kamisama-teki-isekai-life-saikyou-status-ni-tensei-shita-node-suki-ni-ikimasu',
      "div[data-slot='card-title'].break-words", // Selector spesifik untuk judul komik
    )
    dialog.success({
      title: 'Success',
      content: 'Berhasil TITLE = ' + result.content,
    })
    console.log('Result:', result)
    console.log('Title:', result.title)
    console.log('Images found:', result.images.length)
  } catch (err) {
    console.error(err)
  }
}

const scrapeFull = async () => {
  try {
    const result = await BrowserService.ScrapFull(
      'https://westmanga.me/comic/level-count-stop-kara-hajimaru-kamisama-teki-isekai-life-saikyou-status-ni-tensei-shita-node-suki-ni-ikimasu',
    )
    dialog.success({
      title: 'Success',
      content: result,
    })
    console.log('Result:', result)
  } catch (err) {
    console.error(err)
  }
}

const url = ref('')
const scrapeRules = async () => {
  try {
    const rules = JSON.parse(code.value)
    const result = await ScraperService.Scrape(rules, url.value)

    resultJson.value = JSON.stringify(result, null, 2)

    dialog.success({
      title: 'Success',
      content: 'Scraping finished',
    })
    console.log('Result:', result)
  } catch (err) {
    console.error(err)
    dialog.error({
      title: 'Error',
      content: String(err),
    })
  }
}
</script>
