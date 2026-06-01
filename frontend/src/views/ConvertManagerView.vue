<template>
  <div>
    <n-spin :show="ui.running" description="Processing...">
      <div class="flex gap-2 items-stretch">
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
      <div class="flex flex-col gap-2" style="height: calc(100vh - 260px)">
        <n-scrollbar class="flex-1">
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
        <div class="flex items-center justify-between gap-2">
          <div class="text-xs text-gray-400 truncate">
            Size:
            {{ sizeBeforeText }}
            <span v-if="sizeAfterText !== '-'">→ {{ sizeAfterText }}</span>
            <span v-if="sizeDiffText || sizePercentText">
              ({{ sizeDiffText }}{{ sizeDiffText && sizePercentText ? ', ' : ''
              }}{{ sizePercentText }})
            </span>
          </div>
          <div class="flex gap-2 justify-end">
            <n-button tertiary :disabled="ui.running" @click="resetAll">
              Reset
            </n-button>
            <n-button secondary :disabled="ui.running" @click="openProgress">
              Progress
            </n-button>
            <n-button
              type="primary"
              :disabled="!selectedManga || ui.running"
              :loading="ui.running"
              @click="startConvert"
            >
              Start
            </n-button>
          </div>
        </div>
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
              <div class="w-full">
                <div class="grid grid-cols-4 gap-2 mb-3">
                  <n-card size="small" content-style="padding: 0.5rem;">
                    <div class="text-xs text-gray-400">Total</div>
                    <div class="text-lg font-semibold">
                      {{ progress.total }}
                    </div>
                  </n-card>
                  <n-card size="small" content-style="padding: 0.5rem;">
                    <div class="text-xs text-gray-400">Selected</div>
                    <div class="text-lg font-semibold">
                      {{ progress.selected }}
                    </div>
                  </n-card>
                  <n-card size="small" content-style="padding: 0.5rem;">
                    <div class="text-xs text-gray-400">Success</div>
                    <div class="text-lg font-semibold">
                      {{ progress.success }}
                    </div>
                  </n-card>
                  <n-card size="small" content-style="padding: 0.5rem;">
                    <div class="text-xs text-gray-400">Failed</div>
                    <div class="text-lg font-semibold">{{ progress.fail }}</div>
                  </n-card>
                </div>
                <div class="grid grid-cols-2 gap-2 mb-3">
                  <n-card size="small" content-style="padding: 0.5rem;">
                    <div class="text-xs text-gray-400">Skipped</div>
                    <div class="text-lg font-semibold">
                      {{ progress.skipped }}
                    </div>
                  </n-card>
                  <n-card size="small" content-style="padding: 0.5rem;">
                    <div class="text-xs text-gray-400">Selected Manga</div>
                    <div class="text-sm font-medium truncate">
                      {{ selectedManga?.main_title || '-' }}
                    </div>
                  </n-card>
                </div>
                <div class="mb-2">
                  <div class="text-xs text-gray-400 mb-1">Overall</div>
                  <n-progress
                    type="line"
                    :percentage="overallPercent"
                    indicator-placement="inside"
                    processing
                    :border-radius="4"
                  />
                  <div class="flex justify-between text-xs text-gray-400 mt-1">
                    <div>
                      Processed: {{ progress.processed }} / {{ progress.total }}
                    </div>
                    <div>
                      Current:
                      {{
                        progress.currentChapterNumber
                          ? `Ch ${progress.currentChapterNumber}`
                          : '-'
                      }}
                      {{
                        progress.currentStep ? `(${progress.currentStep})` : ''
                      }}
                    </div>
                  </div>
                </div>
                <div class="grid grid-cols-2 gap-2">
                  <div>
                    <div class="text-xs text-gray-400 mb-1">
                      Success Rate (Selected)
                    </div>
                    <n-progress
                      type="line"
                      status="success"
                      :percentage="successPercent"
                      indicator-placement="inside"
                      :border-radius="4"
                    />
                  </div>
                  <div>
                    <div class="text-xs text-gray-400 mb-1">
                      Fail Rate (Selected)
                    </div>
                    <n-progress
                      type="line"
                      status="error"
                      :percentage="failPercent"
                      indicator-placement="inside"
                      :border-radius="4"
                    />
                  </div>
                </div>
              </div>
            </div>
          </n-tab-pane>
          <n-tab-pane name="logp" tab="Log Progress">
            <div class="log-pane">
              <div id="logprogress" class="h-full"></div>
            </div>
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
    </n-spin>
  </div>
</template>

<script setup lang="ts">
import { reactive, computed, ref, onMounted, watch } from 'vue'
import { DatabaseService, FileService } from 'bindings/mangav5/services'
import { MangaBasic } from 'bindings/mangav5/internal/models'
import { useDialog, useMessage } from 'naive-ui'
import { useConvertAndCompress } from '@/composable/useConvertAndCompress'
import { safeWindowsDirectoryName } from '@/utils/filePathHelper'

const message = useMessage()
const dialog = useDialog()
const defaultConfig = {
  quality: 60,
  resize: 1000,
  delete: true,
  compress: true,
  status_read_only: false,
  status_resize: true,
  search: '',
}
const config = reactive({
  ...defaultConfig,
})

const selectedManga = ref<MangaBasic | null>(null)

const progressUi = reactive({
  visible: false,
})

const progress = reactive({
  total: 0,
  selected: 0,
  success: 0,
  fail: 0,
  skipped: 0,
  processed: 0,
  currentChapterNumber: null as number | null,
  currentStep: '' as string,
})

const ui = reactive({
  running: false,
})

const progressLogs = ref<string[]>([])
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

const overallPercent = computed(() => {
  if (progress.total <= 0) return 0
  return Math.min(100, Math.round((progress.processed / progress.total) * 100))
})

const successPercent = computed(() => {
  if (progress.selected <= 0) return 0
  return Math.min(100, Math.round((progress.success / progress.selected) * 100))
})

const failPercent = computed(() => {
  if (progress.selected <= 0) return 0
  return Math.min(100, Math.round((progress.fail / progress.selected) * 100))
})

const resetProgress = () => {
  progress.total = 0
  progress.selected = 0
  progress.success = 0
  progress.fail = 0
  progress.skipped = 0
  progress.processed = 0
  progress.currentChapterNumber = null
  progress.currentStep = ''
  clearLogs()
}

const resetAll = () => {
  resetProgress()
  config.quality = defaultConfig.quality
  config.resize = defaultConfig.resize
  config.delete = defaultConfig.delete
  config.compress = defaultConfig.compress
  config.status_read_only = defaultConfig.status_read_only
  config.status_resize = defaultConfig.status_resize
  config.search = defaultConfig.search
}

const openProgress = () => {
  progressUi.visible = true
}

const sizeUi = reactive({
  beforeBytes: null as number | null,
  afterBytes: null as number | null,
  beforeLoading: false,
  afterLoading: false,
})

const selectedMangaDir = computed(() => {
  const t = selectedManga.value?.main_title?.trim() || ''
  if (!t) return ''
  return safeWindowsDirectoryName(t)
})

const formatBytes = (bytes: number) => {
  if (!Number.isFinite(bytes)) return '-'
  if (bytes === 0) return '0 B'
  const k = 1024
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.min(
    units.length - 1,
    Math.floor(Math.log(bytes) / Math.log(k)),
  )
  const v = bytes / Math.pow(k, i)
  const decimals = i === 0 ? 0 : v >= 100 ? 0 : v >= 10 ? 1 : 2
  return `${v.toFixed(decimals)} ${units[i]}`
}

const sizeBeforeText = computed(() => {
  if (sizeUi.beforeLoading && sizeUi.beforeBytes == null) return '...'
  if (sizeUi.beforeBytes == null) return '-'
  return formatBytes(sizeUi.beforeBytes)
})

const sizeAfterText = computed(() => {
  if (sizeUi.afterLoading && sizeUi.afterBytes == null) return '...'
  if (sizeUi.afterBytes == null) return '-'
  return formatBytes(sizeUi.afterBytes)
})

const sizeDiffText = computed(() => {
  if (sizeUi.beforeBytes == null || sizeUi.afterBytes == null) return ''
  const diff = sizeUi.afterBytes - sizeUi.beforeBytes
  const abs = Math.abs(diff)
  const sign = diff > 0 ? '+' : diff < 0 ? '-' : ''
  return `${sign}${formatBytes(abs)}`
})

const sizePercentText = computed(() => {
  if (sizeUi.beforeBytes == null || sizeUi.afterBytes == null) return ''
  const before = sizeUi.beforeBytes
  if (before <= 0) return ''
  const diff = sizeUi.afterBytes - before
  if (diff === 0) return '0%'
  const pct = (Math.abs(diff) / before) * 100
  const pctText =
    pct >= 100
      ? Math.round(pct).toString()
      : pct >= 10
        ? pct.toFixed(1)
        : pct.toFixed(2)
  return diff < 0 ? `decrease ${pctText}%` : `increase ${pctText}%`
})

const refreshSizeBefore = async (relDir: string) => {
  if (!relDir) {
    sizeUi.beforeBytes = null
    return
  }
  sizeUi.beforeLoading = true
  try {
    sizeUi.beforeBytes = await FileService.GetDirectorySize(relDir)
  } catch (_) {
    sizeUi.beforeBytes = null
  } finally {
    sizeUi.beforeLoading = false
  }
}

const refreshSizeAfter = async (relDir: string) => {
  if (!relDir) {
    sizeUi.afterBytes = null
    return
  }
  sizeUi.afterLoading = true
  try {
    sizeUi.afterBytes = await FileService.GetDirectorySize(relDir)
  } catch (_) {
    sizeUi.afterBytes = null
  } finally {
    sizeUi.afterLoading = false
  }
}

const startConvert = async () => {
  if (!selectedManga.value) {
    message.error('Pilih manga dulu')
    return
  }
  if (ui.running) return

  const ok = await new Promise<boolean>(resolve => {
    dialog.warning({
      title: 'Konfirmasi',
      content: `Mulai convert/compress untuk "${selectedManga.value?.main_title}"?`,
      positiveText: 'Start',
      negativeText: 'Cancel',
      onPositiveClick: () => resolve(true),
      onNegativeClick: () => resolve(false),
    })
  })
  if (!ok) return

  const relDir = selectedMangaDir.value
  const selectedId = selectedManga.value.id
  sizeUi.afterBytes = null
  await refreshSizeBefore(relDir)

  ui.running = true
  resetProgress()
  progressUi.visible = true
  progressLogs.value.push('Starting...')
  try {
    const result = await useConvertAndCompress(
      {
        manga_id: selectedManga.value.id,
        manga_title: selectedManga.value.main_title,
        quality: config.quality,
        resize: config.resize,
        delete: config.delete,
        compress: config.compress,
        status_read_only: config.status_read_only,
        status_resize: config.status_resize,
      },
      {
        onState: s => {
          progress.total = s.total
          progress.selected = s.selected
          progress.success = s.success
          progress.fail = s.failed
          progress.skipped = s.skipped
          progress.processed = s.processed
          progress.currentChapterNumber = s.current_chapter_number
          progress.currentStep = s.current_step ?? ''
        },
        onLog: log => {
          progressLogs.value.push(
            `Ch ${log.chapter_number}: ${log.step} - ${log.message}`,
          )
        },
      },
    )

    progress.total = result.total
    progress.selected = result.selected
    progress.success = result.success
    progress.fail = result.failed
    progress.skipped = result.skipped
    progress.processed = result.total

    if (result.failed > 0) {
      message.error(`Selesai dengan error: ${result.failed} failed`)
    } else {
      message.success('Selesai')
    }
  } catch (error) {
    message.error(`Gagal: ${error}`)
  } finally {
    ui.running = false
    if (selectedManga.value?.id === selectedId) {
      await refreshSizeAfter(relDir)
    }
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

watch(
  () => selectedManga.value?.id,
  () => {
    sizeUi.beforeBytes = null
    sizeUi.afterBytes = null
    const relDir = selectedMangaDir.value
    if (!relDir) return
    refreshSizeBefore(relDir)
  },
)

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
