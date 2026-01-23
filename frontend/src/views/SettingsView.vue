<template>
  <div>
    <div class="mx-auto">
      <div class="flex gap-2">
        <n-input-group>
          <n-input v-model:value="url" />
          <n-button tertiary type="primary" @click="clickScrapeTest">
            GO
          </n-button>
        </n-input-group>
        <n-button tertiary type="primary" @click="clickDownloadTest">
          Download
        </n-button>
        <n-button type="primary"> Save Rules </n-button>
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
    <div>
      <n-split
        direction="horizontal"
        class="h-[calc(100vh-210px)] mt-4"
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

const code = ref('')
const resultJson = ref('')
const dialog = useDialog()

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
  if (!code.value) {
    console.log('code.value is empty')
    return
  }
  const rules = JSON.parse(code.value)
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
</script>
