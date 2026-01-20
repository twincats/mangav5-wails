<template>
  <div>
    <div>Settings view</div>
    <n-space>
      <n-button type="primary" @click="scrape">Scrap</n-button>
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
</script>
