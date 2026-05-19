<template>
  <div style="width: 700px; margin: 0 auto">
    <!-- Global stacked progress bar -->
    <h3>Overall Chapter Status</h3>
    <div
      style="
        position: relative;
        height: 30px;
        background: #e5e5e5;
        border-radius: 6px;
        overflow: hidden;
      "
    >
      <div
        :style="{
          width: successPercent + '%',
          backgroundColor: '#4caf50',
          height: '100%',
          position: 'absolute',
          left: '0',
          top: '0',
        }"
        :title="`Success: ${chapterCounts.success} chapters`"
      ></div>
      <div
        :style="{
          width: partialPercent + '%',
          backgroundColor: '#ffeb3b',
          height: '100%',
          position: 'absolute',
          left: successPercent + '%',
          top: '0',
        }"
        :title="`Partial: ${chapterCounts.partial} chapters`"
      ></div>
      <div
        :style="{
          width: failedPercent + '%',
          backgroundColor: '#f44336',
          height: '100%',
          position: 'absolute',
          left: successPercent + partialPercent + '%',
          top: '0',
        }"
        :title="`Failed: ${chapterCounts.failed} chapters`"
      ></div>
      <div
        :style="{
          width: pendingPercent + '%',
          backgroundColor: '#e0e0e0',
          height: '100%',
          position: 'absolute',
          left: successPercent + partialPercent + failedPercent + '%',
          top: '0',
        }"
        :title="`Pending: ${chapterCounts.pending} chapters`"
      ></div>
    </div>
    <p style="margin-top: 4px">
      Total Chapters: {{ chapters.length }} | Success:
      {{ chapterCounts.success }} | Partial: {{ chapterCounts.partial }} |
      Failed: {{ chapterCounts.failed }} | Pending: {{ chapterCounts.pending }}
    </p>

    <hr style="margin: 20px 0" />

    <!-- Current chapter progress bar -->
    <h3>Current Chapter: {{ currentChapter.name }}</h3>
    <n-progress
      :percentage="currentChapterPercent"
      show-label
      :stroke-color="getChapterColor"
    />
    <p>
      {{ currentChapter.success }}/{{ currentChapter.total }} pages downloaded
      (Failed: {{ currentChapter.failed }})
    </p>

    <!-- Expandable chapter log -->
    <h4>Chapter Logs</h4>
    <div v-for="chap in chapters" :key="chap.name" style="margin-bottom: 8px">
      <n-collapse :default-expanded="false">
        <n-collapse-item :title="`${chap.name} - ${chap.statusText}`">
          <ul style="list-style: none; padding-left: 0">
            <li v-for="(status, index) in chap.pages" :key="index">
              Page {{ index + 1 }}:
              <span :style="{ color: statusColor(status) }">{{ status }}</span>
            </li>
          </ul>
        </n-collapse-item>
      </n-collapse>
    </div>
  </div>
</template>

<script setup>
import { reactive, computed, onMounted } from 'vue'
import { NProgress, NCollapse, NCollapseItem } from 'naive-ui'

// Simulasi 10 chapter, 20 halaman per chapter
const chapters = reactive(
  Array.from({ length: 10 }, (_, i) => ({
    name: `Chapter ${i + 1}`,
    total: 20,
    success: 0,
    failed: 0,
    pages: Array(20).fill('Pending'),
  })),
)

// Update status per chapter
const updateChapterStatus = () => {
  chapters.forEach(chap => {
    chap.remaining = chap.total - chap.success - chap.failed
    if (chap.success === chap.total && chap.total > 0)
      chap.statusText = '✅ Success'
    else if (chap.success > 0) chap.statusText = '⚠️ Partial'
    else if (chap.failed === 0) chap.statusText = '⏳ Pending'
    else chap.statusText = '❌ Failed'
  })
}

// Status counts
const chapterCounts = computed(() => {
  let success = 0,
    partial = 0,
    failed = 0,
    pending = 0
  chapters.forEach(chap => {
    if (chap.success === chap.total && chap.total > 0) success++
    else if (chap.success > 0) partial++
    else if (chap.failed > 0) failed++
    else pending++
  })
  return { success, partial, failed, pending }
})

const totalChapters = computed(() => chapters.length)
const successPercent = computed(
  () => (chapterCounts.value.success / totalChapters.value) * 100,
)
const partialPercent = computed(
  () => (chapterCounts.value.partial / totalChapters.value) * 100,
)
const failedPercent = computed(
  () => (chapterCounts.value.failed / totalChapters.value) * 100,
)
const pendingPercent = computed(
  () => (chapterCounts.value.pending / totalChapters.value) * 100,
)

// Current chapter
const currentChapterIndex = reactive({ index: 0 })
const currentChapter = computed(() => chapters[currentChapterIndex.index])
const currentChapterPercent = computed(
  () => (currentChapter.value.success / currentChapter.value.total) * 100,
)
const getChapterColor = computed(() => {
  if (currentChapter.value.success === currentChapter.value.total)
    return '#4caf50'
  if (currentChapter.value.success > 0) return '#ffeb3b'
  return '#2196f3'
})

// Color mapping for pages
const statusColor = status => {
  if (status === 'Success') return '#4caf50'
  if (status === 'Failed') return '#f44336'
  return '#999'
}

// Live download simulation
onMounted(() => {
  const interval = setInterval(() => {
    const chap = currentChapter.value
    if (!chap) return

    if (chap.success + chap.failed < chap.total) {
      const idx = chap.success + chap.failed
      const r = Math.random()
      if (r < 0.9) {
        chap.pages[idx] = 'Success'
        chap.success++
      } else {
        chap.pages[idx] = 'Failed'
        chap.failed++
      }
    } else {
      if (currentChapterIndex.index < chapters.length - 1)
        currentChapterIndex.index++
      else clearInterval(interval)
    }
    updateChapterStatus()
  }, 200)
})
</script>
