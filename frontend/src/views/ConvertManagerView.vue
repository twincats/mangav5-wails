<template>
  <div>
    <div class="flex gap-2">
      <n-card class="w-xl" size="small" content-style="padding: 0.5rem;">
        <span>Quality : </span>
        <strong>{{ config.quality }}</strong>
        <br />
        <n-slider v-model:value="config.quality" :min="50" :max="90" />
      </n-card>
      <div class="w-full">
        <n-button
          v-if="!config.status_resize"
          @click="config.status_resize = true"
          ghost
          block
          class="h-full"
          >Resize</n-button
        >
        <n-card v-else size="small" content-style="padding: 0.5rem;">
          <span>Resize : </span> <strong>{{ config.resize }}</strong>
          <br />
          <div class="flex gap-2">
            <n-button
              @click="config.status_resize = false"
              ghost
              class="w-12"
              size="small"
              >R</n-button
            >
            <n-slider
              v-model:value="config.resize"
              :min="900"
              :max="1400"
              :step="50"
            />
          </div>
        </n-card>
      </div>

      <n-card class="w-xl" size="small" content-style="padding: 0.5rem;">
        <span>Delete :</span> <br />
        <n-radio-group
          class="w-full flex justify-center"
          v-model:value="config.delete"
          name="radiodelete"
        >
          <n-space>
            <n-radio :value="false" label="No" />
            <n-radio :value="true" label="Yes" />
          </n-space>
        </n-radio-group>
      </n-card>
      <n-card class="w-xl" size="small" content-style="padding: 0.5rem;">
        <span>Compress :</span> <br />
        <n-radio-group
          class="w-full flex justify-center"
          v-model:value="config.compress"
          name="radiocompress"
        >
          <n-space>
            <n-radio :value="false" label="No" />
            <n-radio :value="true" label="Yes" />
          </n-space>
        </n-radio-group>
      </n-card>
      <n-card class="w-xl" size="small" content-style="padding: 0.5rem;">
        <span>Status Read Only :</span> <br />
        <n-radio-group
          class="w-full flex justify-center"
          v-model:value="config.status_read_only"
          name="radiostatus_read_only"
        >
          <n-space>
            <n-radio :value="false" label="No" />
            <n-radio :value="true" label="Yes" />
          </n-space>
        </n-radio-group>
      </n-card>
    </div>
    <div class="w-full my-2">
      <n-input v-model:value="config.search" />
    </div>
    <div>
      <n-scrollbar style="height: calc(100vh - 400px)">
        <n-list bordered hoverable clickable class="manga-list">
          <n-list-item
            v-for="item in mangaListFilter"
            :key="item.id"
            :style="{
              cursor: 'pointer',
              backgroundColor:
                selectedManga?.main_title === item.main_title
                  ? 'rgba(24, 160, 88, 0.18)'
                  : '',
            }"
            @click="selectedManga = item"
          >
            {{ item.main_title }}
          </n-list-item>
        </n-list>
      </n-scrollbar>
    </div>
    <n-modal
      v-model:show="progressUi.visible"
      preset="card"
      class="w-[720px]"
      title="Progress"
    >
      <n-tabs type="line" animated>
        <n-tab-pane name="progress" tab="Progress">
          <div class="progress-pane">
            <!-- <stack-progress
              :total="progress.total"
              :success="progress.success"
              :fail="progress.fail"
            /> -->
            <download-progress
              :total="progress.total"
              :success="progress.success"
              :fail="progress.fail"
            />
          </div>
        </n-tab-pane>
        <n-tab-pane name="logp" tab="Log Progress">
          <div id="logprogress"></div>
        </n-tab-pane>
        <n-tab-pane name="log" tab="Log">
          <div class="log-pane">
            <div class="flex items-center justify-between gap-2 mb-2">
              <div class="text-sm text-gray-400">Latest logs</div>
              <div class="flex items-center gap-2">
                <n-button size="tiny" secondary @click="clearLogs"
                  >Clear</n-button
                >
                <n-button size="tiny" secondary @click="copyLogs"
                  >Copy</n-button
                >
              </div>
            </div>
            <n-log :log="progressLogsText" :rows="12" />
          </div>
        </n-tab-pane>
      </n-tabs>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { reactive, computed, ref, onMounted } from 'vue'
import { DatabaseService } from 'bindings/mangav5/services'
import { MangaBasic } from 'bindings/mangav5/internal/models'
import { useMessage } from 'naive-ui'

const message = useMessage()
const config = reactive({
  quality: 60,
  resize: 1000,
  delete: true,
  compress: true,
  status_read_only: true,
  status_resize: true,
  search: '',
})

const selectedManga = ref<MangaBasic | null>(null)

const progressUi = reactive({
  visible: true,
})

const progress = reactive({
  total: 100,
  success: 51,
  fail: 0,
})

const progressLogs = ref<string[]>([
  'Starting convert...',
  'Scanning chapters...',
  'Processing: Chapter 1',
  'Success: Chapter 1',
  'Processing: Chapter 2',
  'Failed: Chapter 2 (file missing)',
])
const progressLogsText = computed(() => progressLogs.value.join('\n'))
const clearLogs = () => {
  progressLogs.value = []
}
const copyLogs = async () => {
  try {
    await navigator.clipboard.writeText(progressLogsText.value)
    message.success('Log copied')
  } catch (_) {
    message.error('Failed to copy log')
  }
}

const mangaList = ref<MangaBasic[]>([])
const getMangaList = async () => {
  try {
    mangaList.value = await DatabaseService.ListMangaBasic()
  } catch (error) {
    message.error(`Error fetching manga list : ${error}`)
  }
}

const mangaListFilter = computed(() => {
  return mangaList.value.filter(item =>
    item.main_title.toLowerCase().includes(config.search.toLowerCase()),
  )
})

onMounted(() => {
  getMangaList()
})
</script>

<style scoped>
.manga-list :deep(.n-list-item) {
  transition: background-color 120ms ease;
}

.manga-list :deep(.n-list-item:hover) {
  background-color: rgba(24, 160, 88, 0.1);
}

.progress-pane {
  min-height: 320px;
  display: flex;
  align-items: flex-start;
}

.log-pane {
  min-height: 320px;
}
</style>
