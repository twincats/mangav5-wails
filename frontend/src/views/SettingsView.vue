<template>
  <div>
    <div>Settings view</div>
    <n-space>
      <n-button type="primary" @click="scrape">Scrap</n-button>
      <n-button type="info" @click="scrapeJsRender">Scrap JS Render</n-button>
      <n-button type="primary" @click="$router.push('/')">Goto Home</n-button>
    </n-space>

    <div class="h-[calc(100vh-170px)] mt-4">
      <MonacoEditor
        v-model="code"
        language="json"
        theme="vs-dark"
        :formatOnLoad="true"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import MonacoEditor from '@/components/MonacoEditor.vue'
import { BrowserService } from '../../bindings/mangav5/services'

const code = ref('')
// Contoh scraping
async function scrape() {
  try {
    const result = await BrowserService.ScrapePage(
      'https://example.com',
      'h1', // Optional selector
    )
    console.log('Title:', result.title)
    console.log('Images found:', result.images.length)
  } catch (err) {
    console.error(err)
  }
}

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
</script>
