<template>
  <div
    class="reader-layout"
    ref="readerLayoutRef"
    @contextmenu.prevent="handleLayoutContextMenu"
  >
    <div class="reader-header">
      <span>Chapter {{ chapter?.chapter_number }}</span>
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
    <div class="reader-scroll-area" ref="scrollAreaRef">
      <div
        v-for="(row, rowIndex) in displayRows"
        :key="rowIndex"
        class="reader-row"
        :class="{
          'is-double': readingMode === 'double-page',
          'is-rtl': direction === 'rtl',
        }"
        :style="readerRowStyle"
        :ref="el => (rowRefs[rowIndex] = el as HTMLElement)"
        :data-indexes="row.map(r => r.index).join(',')"
      >
        <div
          v-for="(img, imgIndex) in row"
          :key="imgIndex"
          class="image-wrapper"
          :class="{ 'double-page-item': readingMode === 'double-page' }"
          @contextmenu.prevent.stop="handleContextMenu($event, img)"
        >
          <n-image
            :src="`${ImagePath(chapter?.path + '/' + img.fileName)}`"
            preview-disabled
            lazy
            object-fit="contain"
            class="reader-image"
            :img-props="{
              decoding: 'async',
              fetchpriority: priorityIndexes.has(img.index) ? 'high' : 'auto',
            }"
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
      <div ref="endSentinelRef" style="height: 1px; width: 100%"></div>
    </div>
    <teleport v-if="teleportReady" :to="teleportTarget">
      <context-menu ref="refMenu">
        <template #default="{ item }">
          <li @click="toggleFullscreen">
            {{ isFullscreen ? 'Exit Fullscreen' : 'Fullscreen' }}
          </li>
          <li @click="toggleReadingMode">
            {{
              readingMode === 'double-page' ? 'Long Strip Mode' : '2 Pages Mode'
            }}
          </li>
          <li
            :class="{ disabled: readingMode === 'long-strip' }"
            @click="toggleDirection"
          >
            {{ direction === 'rtl' ? 'LTR' : 'RTL' }}
          </li>
          <li
            :class="{ disabled: readingMode === 'double-page' }"
            @click="toggleFullWidth"
          >
            {{
              readingMode === 'double-page'
                ? 'Full Width'
                : fullWidth
                  ? 'Normal Width'
                  : 'Full Width'
            }}
          </li>
          <div class="divider"></div>
          <li :class="{ disabled: !hasPrev }" @click="navigateChapter('prev')">
            Previous Chapter
          </li>
          <li @click="goToHome">Home</li>
          <li :class="{ disabled: !hasNext }" @click="navigateChapter('next')">
            Next Chapter
          </li>
          <template v-if="item">
            <div class="divider"></div>
            <li @click="copyImage(item.src)">Copy Image</li>
            <li @click="copyLink(item.src)">Image Link</li>
            <div class="divider"></div>
            <li class="red" @click="deleteMenu(item.fileName)">
              Delete {{ item.fileName }}
            </li>
          </template>
        </template>
      </context-menu>
    </teleport>
  </div>
</template>

<script setup lang="ts">
import { DatabaseService, FileService } from '../../bindings/mangav5/services'
import { Chapter, MangaDetail } from '../../bindings/mangav5/internal/models'
import { ImagePath } from '@/utils/filePathHelper'
import { Window } from '@wailsio/runtime'
import { onBeforeRouteLeave } from 'vue-router'
import { UseContextMenu } from '@/utils/contextMenuHelper'

const message = useMessage()
const dialog = useDialog()
const route = useRoute()
const router = useRouter()
const { mangaId, chapterId } = defineProps<{
  mangaId: number
  chapterId: number
}>()
const imageList = ref<string[]>([])
const chapter = ref<Chapter | null>(null)
const mangaDetail = ref<MangaDetail | null>(null)
const currentChapterIndex = ref<number>(0)
const { refMenu, openContextMenu, closeContextMenu } = UseContextMenu()

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
const readingMode = ref<'long-strip' | 'double-page'>('long-strip')
const direction = ref<'rtl' | 'ltr'>('rtl')

// Image Dimensions for Layout
const imageDimensions = reactive<
  Record<string, { width: number; height: number }>
>({})

const preloadImages = () => {
  if (!chapter.value) return
  const basePath = chapter.value.path
  const limit = 8
  imageList.value.slice(0, limit).forEach(img => {
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

const ensureDimensionsForIndexes = (indexes: number[]) => {
  if (!chapter.value) return
  const basePath = chapter.value.path
  indexes.forEach(i => {
    const fname = imageList.value[i]
    if (!fname || imageDimensions[fname]) return
    const src = ImagePath(basePath + '/' + fname)
    const image = new Image()
    image.onload = () => {
      imageDimensions[fname] = {
        width: image.naturalWidth,
        height: image.naturalHeight,
      }
    }
    image.src = src
  })
}

const getChapterImageList = async (chapter_path: string) => {
  try {
    imageList.value = await FileService.GetImageList(chapter_path)
    preloadImages()
  } catch (error) {
    message.error(`Error fetching chapter image list : ${error}`)
  }
}

const getCurrentChapter = (): Chapter | null => {
  if (!mangaDetail.value) {
    getMangaDetail()
  }
  if (mangaDetail.value && mangaDetail.value.chapters) {
    const chapters = mangaDetail.value?.chapters ?? []
    const currentChapter = chapters.find(chap => chap.id === chapterId) ?? null
    chapter.value = currentChapter
    return currentChapter
  }
  return null
}

const getMangaDetail = async () => {
  try {
    // get mangaDetail
    mangaDetail.value = await DatabaseService.GetMangaDetail(mangaId)
    if (mangaDetail.value?.chapters) {
      mangaDetail.value.chapters.sort(
        (a, b) =>
          Number(a.chapter_number ?? a.id ?? 0) -
          Number(b.chapter_number ?? b.id ?? 0),
      )
    }
  } catch (error) {
    message.error(`Error fetching mangaDetail : ${error}`)
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

const handleContextMenu = (ev: MouseEvent, img: ImageItem) => {
  const src = ImagePath((chapter.value?.path || '') + '/' + img.fileName)
  openContextMenu(ev, {
    src,
    fileName: img.fileName,
    index: img.index,
  })
}

const handleLayoutContextMenu = (ev: MouseEvent) => {
  openContextMenu(ev)
}
const copyLink = async (src?: string) => {
  if (!src) return
  try {
    await navigator.clipboard.writeText(src)
    message.success('Image link copied')
  } catch (_) {
    message.error('Failed to copy link')
  } finally {
    closeContextMenu()
  }
}

const copyImage = async (src?: string) => {
  if (!src) return
  try {
    const res = await fetch(src)
    const blob = await res.blob()
    const type = blob.type || 'image/png'
    const ClipboardItemAny: any = (window as any).ClipboardItem
    if (ClipboardItemAny) {
      await (navigator as any).clipboard.write([
        new ClipboardItemAny({ [type]: blob }),
      ])
      message.success('Image copied to clipboard')
    } else {
      await navigator.clipboard.writeText(src)
      message.warning('Clipboard image not supported; link copied instead')
    }
  } catch (_) {
    try {
      await navigator.clipboard.writeText(src)
      message.warning('Failed copying image; link copied instead')
    } catch {
      message.error('Failed to copy image')
    }
  } finally {
    closeContextMenu()
  }
}

const toggleReadingMode = () => {
  readingMode.value =
    readingMode.value === 'double-page' ? 'long-strip' : 'double-page'
  closeContextMenu()
}

const toggleDirection = () => {
  if (readingMode.value !== 'double-page') {
    return
  }
  direction.value = direction.value === 'rtl' ? 'ltr' : 'rtl'
  closeContextMenu()
}

const isFullscreen = ref(!!document.fullscreenElement)
const toggleFullscreen = async () => {
  try {
    if (!document.fullscreenElement) {
      const el = readerLayoutRef.value || document.documentElement
      await el.requestFullscreen()
    } else {
      await document.exitFullscreen()
    }
  } catch (_) {
  } finally {
    isFullscreen.value = document.fullscreenElement === readerLayoutRef.value
    closeContextMenu()
  }
}

const fullWidth = ref(false)
const readerRowStyle = computed(() => {
  if (readingMode.value === 'long-strip') {
    return { maxWidth: fullWidth.value ? '100%' : '1000px' }
  }
  return {}
})
const toggleFullWidth = () => {
  if (readingMode.value !== 'long-strip') {
    return
  }
  fullWidth.value = !fullWidth.value
  closeContextMenu()
}

const goToHome = () => {
  router.push({ name: 'home' })
}

const deleteMenu = (filename: string) => {
  dialog.error({
    title: 'Confirm Delete File',
    content: `Are you sure to delete ${filename}?`,
    positiveText: 'Sure',
    negativeText: 'Not Sure',
    maskClosable: false,
    negativeButtonProps: {
      color: 'grey',
    },
    onPositiveClick: () => {
      message.success('Sure')
    },
    onNegativeClick: () => {
      message.error('Not Sure')
    },
  })
  closeContextMenu()
}

const readerLayoutRef = ref<HTMLElement | null>(null)
const scrollAreaRef = ref<HTMLElement | null>(null)
const rowRefs: HTMLElement[] = []
const priorityIndexes = new Set<number>()
let rowObserver: IntersectionObserver | null = null
let bottomObserver: IntersectionObserver | null = null
const endSentinelRef = ref<HTMLElement | null>(null)
const hasMarkedRead = ref(false)
const isAtBottom = () => {
  const el = scrollAreaRef.value
  if (!el) return false
  return el.scrollTop + el.clientHeight >= el.scrollHeight - 2
}
const markChapterAsRead = async () => {
  if (hasMarkedRead.value) return
  try {
    await DatabaseService.MarkChapterAsRead(chapterId)
    hasMarkedRead.value = true
    if (chapter.value) {
      chapter.value.status_read = 1
    }
  } catch (_) {}
}
const onScrollCheck = () => {
  if (!hasMarkedRead.value && isAtBottom()) {
    markChapterAsRead()
  }
}
const teleportTarget = computed(() =>
  isFullscreen.value ? readerLayoutRef.value || '#main' : '#main',
)
const teleportReady = computed(() =>
  isFullscreen.value ? !!readerLayoutRef.value : true,
)

const hasPrev = computed(() => {
  return currentChapterIndex.value > 0
})
const hasNext = computed(() => {
  const len = mangaDetail.value?.chapters?.length ?? 0
  return currentChapterIndex.value < len - 1
})

const navigateChapter = (direction: 'prev' | 'next') => {
  let targetIndex: number
  if (direction === 'prev') {
    targetIndex = currentChapterIndex.value - 1
  } else {
    targetIndex = currentChapterIndex.value + 1
  }
  if (mangaDetail.value) {
    if (targetIndex >= 0 && targetIndex < mangaDetail.value?.chapters.length) {
      const targetChapter = mangaDetail.value?.chapters[targetIndex]
      router.push(`/read/${mangaId}/${targetChapter.id}`)
    }
  }
}
// Update chapter when route prop changes
watch(
  () => chapterId,
  newId => {
    if (!mangaDetail.value) return
    hasMarkedRead.value = false
    const idx = mangaDetail.value.chapters.findIndex(chap => chap.id === newId)
    if (idx >= 0) {
      currentChapterIndex.value = idx
      const ch = mangaDetail.value.chapters[idx]
      chapter.value = ch
      getChapterImageList(ch.path)
    }
  },
)
watch(
  () => displayRows.value,
  () => {
    setTimeout(() => {
      if (!rowObserver) return
      rowObserver.disconnect()
      rowRefs.forEach(el => el && rowObserver?.observe(el))
      if (bottomObserver && endSentinelRef.value) {
        bottomObserver.disconnect()
        bottomObserver.observe(endSentinelRef.value)
      }
    }, 0)
  },
)
onMounted(async () => {
  // load first time reader
  await getMangaDetail()
  const currentChapter = getCurrentChapter()
  if (currentChapter) {
    getChapterImageList(currentChapter.path)
    currentChapterIndex.value =
      mangaDetail.value?.chapters.findIndex(
        chapter => chapter.id === currentChapter.id,
      ) || 0
    for (let i = 0; i < Math.min(6, imageList.value.length); i++) {
      priorityIndexes.add(i)
    }
  }
  refMenu.value
  document.addEventListener('fullscreenchange', () => {
    isFullscreen.value = document.fullscreenElement === readerLayoutRef.value
  })
  if (rowObserver) {
    rowObserver.disconnect()
  }
  rowObserver = new IntersectionObserver(
    entries => {
      entries.forEach(entry => {
        const el = entry.target as HTMLElement
        const data = el.getAttribute('data-indexes') || ''
        const indexes = data ? data.split(',').map(s => parseInt(s)) : []
        if (entry.isIntersecting) {
          indexes.forEach(i => priorityIndexes.add(i))
          ensureDimensionsForIndexes(indexes)
        } else {
          indexes.forEach(i => priorityIndexes.delete(i))
        }
      })
    },
    {
      root: scrollAreaRef.value || null,
      rootMargin: '2000px 0px',
      threshold: 0,
    },
  )
  rowRefs.forEach(el => el && rowObserver?.observe(el))
  scrollAreaRef.value?.addEventListener('scroll', onScrollCheck, {
    passive: true,
  })
  if (bottomObserver) {
    bottomObserver.disconnect()
  }
  bottomObserver = new IntersectionObserver(
    entries => {
      entries.forEach(entry => {
        if (entry.isIntersecting) {
          markChapterAsRead()
        }
      })
    },
    {
      root: scrollAreaRef.value || null,
      rootMargin: '0px',
      threshold: 0,
    },
  )
  if (endSentinelRef.value) {
    bottomObserver.observe(endSentinelRef.value)
  }
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
  content-visibility: auto;
  contain-intrinsic-size: 800px;
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
