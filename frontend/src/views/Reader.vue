<template>
  <div class="reader-layout">
    <div class="reader-header">
      <span>Reader {{ route.params.chapterId }}</span>
      <n-divider vertical />
      <n-radio-group v-model:value="readingMode" size="small">
        <n-radio-button value="long-strip">Long Strip</n-radio-button>
        <n-radio-button value="double-page">Double Page</n-radio-button>
      </n-radio-group>
      <n-divider vertical v-if="readingMode === 'double-page'" />
      <n-radio-group
        v-model:value="direction"
        size="small"
        v-if="readingMode === 'double-page'"
      >
        <n-radio-button value="rtl">RTL</n-radio-button>
        <n-radio-button value="ltr">LTR</n-radio-button>
      </n-radio-group>
    </div>

    <!-- Unified Scroll Area -->
    <div class="reader-scroll-area">
      <div
        v-for="(row, rowIndex) in displayRows"
        :key="rowIndex"
        class="reader-row"
        :class="{
          'is-double': readingMode === 'double-page',
          'is-rtl': direction === 'rtl',
        }"
      >
        <div
          v-for="(img, imgIndex) in row"
          :key="imgIndex"
          class="image-wrapper"
          :class="{ 'double-page-item': readingMode === 'double-page' }"
          :style="{
            '--custom-contextmenu': 'read-menu',
          }"
        >
          <n-image
            :src="`${ImagePath(chapter?.path + '/' + img.fileName)}`"
            preview-disabled
            lazy
            object-fit="contain"
            class="reader-image"
          >
            <template #placeholder>
              <div
                class="flex items-center justify-center h-[50vh] w-full bg-gray-800/30 rounded"
              >
                <n-spin size="large" />
              </div>
            </template>
          </n-image>
          <div class="image-counter">
            <!-- Calculate actual index based on row -->
            {{ img.index + 1 }} /
            {{ imageList.length }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { DatabaseService, FileService } from '../../bindings/mangav5/services'
import { Chapter } from '../../bindings/mangav5/internal/models'
import { ImagePath } from '@/utils/filePathHelper'
import { Window } from '@wailsio/runtime'
import { onBeforeRouteLeave } from 'vue-router'

const message = useMessage()
const route = useRoute()
const { chapterId } = defineProps<{ chapterId: number }>()
const imageList = ref<string[]>([])
const chapter = ref<Chapter | null>(null)

// Window State Management
const wasMaximizedBefore = ref(false)
Window.IsMaximised().then((isMax: boolean) => {
  wasMaximizedBefore.value = isMax
  if (!isMax) {
    Window.Maximise()
  }
})

onBeforeRouteLeave((_to, _from, next) => {
  if (!wasMaximizedBefore.value) {
    Window.Restore()
  }
  next()
})

// Reading State
const readingMode = ref<'long-strip' | 'double-page'>('double-page')
const direction = ref<'rtl' | 'ltr'>('rtl')

// Image Dimensions for Layout
const imageDimensions = reactive<
  Record<string, { width: number; height: number }>
>({})

const preloadImages = () => {
  if (!chapter.value) return
  const basePath = chapter.value.path
  imageList.value.forEach(img => {
    // Only load if not already known
    if (imageDimensions[img]) return

    const src = ImagePath(basePath + '/' + img)
    const image = new Image()
    image.onload = () => {
      imageDimensions[img] = {
        width: image.naturalWidth,
        height: image.naturalHeight,
      }
    }
    image.src = src
  })
}

const getChapterImageList = async (chapter_id: number) => {
  try {
    chapter.value = await DatabaseService.GetChapter(chapter_id)
    if (chapter.value) {
      try {
        imageList.value = await FileService.GetImageList(chapter.value.path)
        preloadImages()
      } catch (error) {
        message.error(`Error fetching chapter image list : ${error}`)
      }
    }
  } catch (error) {
    message.error(`Error fetching chapter image list : ${error}`)
  }
}

interface ImageItem {
  fileName: string
  index: number
}

const displayRows = computed(() => {
  if (readingMode.value === 'long-strip') {
    // 1 image per row
    return imageList.value.map((img, index) => [{ fileName: img, index }])
  } else {
    // 2 images per row, but handle wide images
    const rows: ImageItem[][] = []
    let i = 0
    while (i < imageList.value.length) {
      const img = imageList.value[i]
      const dim = imageDimensions[img]
      const isWide = dim ? dim.width > dim.height : false // Default to portrait if not loaded

      if (isWide) {
        rows.push([{ fileName: img, index: i }])
        i++
      } else {
        // Current is portrait
        if (i + 1 < imageList.value.length) {
          const nextImg = imageList.value[i + 1]
          const nextDim = imageDimensions[nextImg]
          const nextIsWide = nextDim ? nextDim.width > nextDim.height : false

          if (!nextIsWide) {
            // Both portrait -> Pair
            rows.push([
              { fileName: img, index: i },
              { fileName: nextImg, index: i + 1 },
            ])
            i += 2
          } else {
            // Next is wide -> Current alone
            rows.push([{ fileName: img, index: i }])
            i++
          }
        } else {
          // Last one alone
          rows.push([{ fileName: img, index: i }])
          i++
        }
      }
    }
    return rows
  }
})

// calculateIndex is no longer needed but kept empty to avoid breaking template if referenced (though removed above)
const calculateIndex = (rowIndex: number, imgIndex: number) => 0

onMounted(() => {
  getChapterImageList(chapterId)
})
</script>

<style scoped>
.reader-layout {
  display: flex;
  flex-direction: column;
  height: 100%;
  width: 100%;
  background-color: #121212;
  overflow: hidden;
}

.reader-header {
  padding: 10px 20px;
  background-color: #1e1e1e;
  color: #e0e0e0;
  text-align: center;
  font-weight: bold;
  border-bottom: 1px solid #333;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
}

.reader-scroll-area {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 0;
  gap: 4px; /* Small gap between rows */
}

/* Row Styles */
.reader-row {
  display: flex;
  width: 100%;
  max-width: 1000px; /* Base width for long strip */
  justify-content: center;
  /* Remove gaps */
  line-height: 0;
  font-size: 0;
}

.reader-row.is-double {
  max-width: 100%; /* Allow full width for double page */
  flex-direction: row; /* Default LTR */
  gap: 4px; /* Small gap between pages */
  align-items: flex-start;
}

.reader-row.is-double.is-rtl {
  flex-direction: row-reverse;
}

/* Image Wrapper */
.image-wrapper {
  position: relative;
  width: 100%; /* Default for long-strip */
  display: flex;
  justify-content: center;
}

.image-wrapper.double-page-item {
  flex: 1;
  width: 0; /* Ensure flex items split space evenly */
  min-width: 0;
}

/* Image */
.reader-image {
  width: 100%;
  height: auto;
  display: block;
}

.image-wrapper.double-page-item :deep(img) {
  width: 100%;
  height: auto;
  display: block;
}

.image-wrapper :deep(img) {
  display: block;
  width: 100%;
  height: auto;
  border-radius: 4px;
}

.image-counter {
  position: absolute;
  bottom: 10px;
  right: 10px;
  background: rgba(0, 0, 0, 0.6);
  color: #fff;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  line-height: normal;
  pointer-events: none;
  opacity: 0;
  transition: opacity 0.2s ease-in-out;
}

.image-wrapper:hover .image-counter {
  opacity: 1;
}

/* Scrollbar */
.reader-scroll-area::-webkit-scrollbar {
  width: 8px;
}
.reader-scroll-area::-webkit-scrollbar-track {
  background: #1e1e1e;
}
.reader-scroll-area::-webkit-scrollbar-thumb {
  background: #444;
  border-radius: 4px;
}
.reader-scroll-area::-webkit-scrollbar-thumb:hover {
  background: #555;
}
</style>
