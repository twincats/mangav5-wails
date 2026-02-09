<template>
  <div class="p-4">
    <n-spin :show="!mangaDetail">
      <div v-if="mangaDetail">
        <n-card>
          <n-grid x-gap="24" cols="1 600:6">
            <n-gi span="1">
              <div class="aspect-[2/3] w-full relative">
                <n-image
                  :src="`${ImagePath(mangaDetail.main_title)}/cover`"
                  object-fit="cover"
                  class="rounded-md w-full h-full"
                  :img-props="{
                    style: 'width: 100%; height: 100%; object-fit: cover;',
                  }"
                  fallback-src="/placeholder.png"
                />
              </div>
            </n-gi>
            <n-gi span="5">
              <n-space vertical size="large">
                <div>
                  <n-h1 class="mb-2">{{ mangaDetail.main_title }}</n-h1>
                  <n-space align="center">
                    <n-tag :type="getStatusType(mangaDetail.manga_status)">
                      {{ mangaDetail.manga_status }}
                    </n-tag>
                    <n-text depth="3">{{ mangaDetail.year }}</n-text>
                  </n-space>
                </div>

                <n-descriptions bordered label-placement="left" :column="1">
                  <n-descriptions-item
                    label="Alternative Titles"
                    v-if="mangaDetail.alternative_titles?.length"
                  >
                    <n-space size="small">
                      <n-tag
                        size="small"
                        v-for="alt in mangaDetail.alternative_titles"
                        :key="alt.id"
                        :bordered="false"
                      >
                        {{ alt.alternative_title }}
                      </n-tag>
                    </n-space>
                  </n-descriptions-item>
                  <n-descriptions-item label="Description">
                    <div class="whitespace-pre-wrap">
                      {{ mangaDetail.description }}
                    </div>
                  </n-descriptions-item>
                </n-descriptions>
              </n-space>
            </n-gi>
          </n-grid>
        </n-card>

        <n-divider />

        <n-card title="Chapters" size="small">
          <template #header-extra>
            <n-tag type="info" round>
              Total: {{ mangaDetail.chapters?.length || 0 }}
            </n-tag>
          </template>

          <n-data-table
            :columns="columns"
            :data="mangaDetail.chapters || []"
            :pagination="{ pageSize: 20 }"
            :row-key="row => row.id"
            striped
          />
        </n-card>
      </div>
      <div v-else class="h-[50vh] flex items-center justify-center">
        <n-empty description="Manga not found" />
      </div>
    </n-spin>
  </div>
</template>

<script setup lang="ts">
import { h, onMounted, ref } from 'vue'
import { MangaDetail, Chapter } from 'bindings/mangav5/internal/models'
import { DatabaseService } from 'bindings/mangav5/services'
import { ImagePath } from '@/utils/filePathHelper'
import { NButton, NTag, NSpace, useMessage, NIcon, NTooltip } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { useRouter } from 'vue-router'
import {
  ArchiveFilled,
  FolderFilled,
  CheckCircleFilled,
  ErrorFilled,
  VisibilityFilled,
  VisibilityOffFilled,
  WarningFilled,
} from '@vicons/material'

const message = useMessage()
const router = useRouter()
const { mangaId } = defineProps<{ mangaId: number }>()
const mangaDetail = ref<MangaDetail | null>(null)

const getManga = async () => {
  try {
    mangaDetail.value = await DatabaseService.GetMangaDetail(mangaId)
  } catch (error) {
    message.error(`Error fetching manga : ${error}`)
  }
}

const getStatusType = (status: string) => {
  switch (status?.toLowerCase()) {
    case 'ongoing':
      return 'success'
    case 'completed':
      return 'info'
    case 'hiatus':
      return 'warning'
    default:
      return 'default'
  }
}

const formatDate = (ts: number) => {
  if (!ts) return '-'
  return new Date(ts * 1000).toLocaleDateString()
}

const readChapter = (chapterId: number) => {
  router.push(`/read/${chapterId}`)
}

const renderIcon = (icon: any, color: string, tooltip: string) => {
  return h(
    NTooltip,
    { trigger: 'hover' },
    {
      trigger: () =>
        h(NIcon, { color: color, size: 18 }, { default: () => h(icon) }),
      default: () => tooltip,
    },
  )
}

const columns: DataTableColumns<Chapter> = [
  {
    title: 'Status',
    key: 'info',
    width: 90,
    render(row) {
      // Icon Kompresi
      const compressIcon =
        row.is_compressed === 1 ? ArchiveFilled : FolderFilled
      const compressColor = row.is_compressed === 1 ? '#f59e0b' : '#3b82f6' // Amber for zip, Blue for folder
      const compressText =
        row.is_compressed === 1 ? 'Compressed (CBZ/ZIP)' : 'Folder'

      // Icon Status File
      let statusIcon = CheckCircleFilled
      let statusColor = '#10b981' // Green
      let statusText = 'Valid'

      if (row.status === 'missing') {
        statusIcon = ErrorFilled
        statusColor = '#ef4444' // Red
        statusText = 'File Missing'
      } else if (row.status === 'corrupted') {
        statusIcon = WarningFilled
        statusColor = '#f97316' // Orange
        statusText = 'Corrupted'
      }

      // Icon Read
      const isRead = row.status_read === 1
      const readIcon = isRead ? VisibilityFilled : VisibilityOffFilled
      const readColor = isRead ? '#8b5cf6' : '#9ca3af' // Violet if read, Gray if unread
      const readText = isRead ? 'Read' : 'Unread'

      return h(
        NSpace,
        { size: 'small', align: 'center' },
        {
          default: () => [
            renderIcon(compressIcon, compressColor, compressText),
            renderIcon(statusIcon, statusColor, statusText),
            renderIcon(readIcon, readColor, readText),
          ],
        },
      )
    },
  },
  {
    title: 'Chapter',
    key: 'chapter_number',
    width: 100,
    render(row) {
      return `Ch. ${row.chapter_number}`
    },
    sorter: (row1, row2) => row1.chapter_number - row2.chapter_number,
  },
  {
    title: 'Title',
    key: 'chapter_title',
    render(row) {
      return row.chapter_title || '-'
    },
  },
  {
    title: 'Language',
    key: 'language',
    width: 100,
    render(row) {
      const lang = row.language || 'en'
      return h(
        NTag,
        {
          size: 'small',
          bordered: false,
          // 'error' (red) for ID, 'info' (blue) for EN/others
          type: lang.toLowerCase() === 'id' ? 'error' : 'info',
        },
        { default: () => lang.toUpperCase() },
      )
    },
  },
  {
    title: 'Date',
    key: 'release_time_ts',
    width: 150,
    render(row) {
      return formatDate(row.release_time_ts)
    },
  },
  {
    title: 'Action',
    key: 'actions',
    width: 100,
    render(row) {
      return h(
        NButton,
        {
          size: 'small',
          type: 'primary',
          secondary: true,
          disabled: row.status === 'missing', // Disable if missing
          onClick: () => readChapter(row.id),
        },
        { default: () => 'Read' },
      )
    },
  },
]

onMounted(() => {
  getManga()
})
</script>

<style scoped>
.manga-cover {
  width: 100%;
  border-radius: 8px;
  box-shadow:
    0 4px 6px -1px rgba(0, 0, 0, 0.1),
    0 2px 4px -1px rgba(0, 0, 0, 0.06);
}
</style>
