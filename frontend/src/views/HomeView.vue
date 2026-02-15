<template>
  <div>
    <div>
      <home-search
        class="mb-1"
        v-model:search="search"
        v-model:dateModel="dateModel"
        v-if="breakpoints.greaterOrEqual('2xl').value"
      />
      <div class="grid grid-cols-6 xl:grid-cols-10 gap-2">
        <div
          v-for="(m, index) in mangaView"
          class="relative group select-none rounded-1 transition-all duration-300"
          :class="{ 'today-highlight': isToday(m.download_time) }"
          @click="clickManga(m.manga_id)"
          @contextmenu.prevent="openContextMenu"
        >
          <n-image
            class="rounded-1 block"
            width="100%"
            :key="index"
            :src="`${ImagePath(m.main_title)}/cover`"
            object-fit="cover"
            preview-disabled
            :img-props="{
              style: 'width: 100%; height: 250px; display: block;',
            }"
          />
          <div
            class="absolute top-0 right-0 text-center text-white rounded-bl-1 rounded-tr-1 bg-black bg-opacity-50 px-1"
          >
            {{ formatDate(m.download_time) }}
          </div>
          <div
            class="bg-gradient w-full absolute bottom-0 rounded-b-1 transition-all duration-300 ease-in-out group-hover:h-[40%] h-[4rem]"
          >
            <div class="text-center h-[2.5rem] line-clamp-2 p-2">
              {{ m.main_title }}
            </div>
            <div
              @click.stop="clickChapter(m.manga_id, m.chapter_id)"
              class="absolute bottom-0 left-0 right-0 bg-white/50 text-center rounded-b-1 text-black transition-all duration-300 ease-in-out group-hover:opacity-100 opacity-0"
            >
              <strong>Chapter {{ m.chapter_number }}</strong>
            </div>
          </div>
        </div>
      </div>
      <div
        v-if="totalPages > 1"
        class="mt-5 w-full absolute bottom-0 mb-5 flex justify-center"
      >
        <n-pagination
          v-model:page="currentPage"
          :page-size="pageSize"
          :item-count="totalItems"
        />
      </div>
    </div>
    <teleport to="#main">
      <context-menu ref="refMenu">
        <li class="disabled">Add Alternative</li>
        <li @click="search = ''">Convert Chapter Webp</li>
        <li>Compress Manga Chapter</li>
        <div class="divider"></div>
        <li class="red">Delete Manga</li>
      </context-menu>
    </teleport>
  </div>
</template>
<!-- Ore ga Kokuhaku Sarete Kara, Ojou no Yousu ga Okashii/4/01.jpg -->
<script setup lang="ts">
import { DatabaseService } from '../../bindings/mangav5/services'
import { LatestManga } from '../../bindings/mangav5/internal/models'
import { ImagePath } from '@/utils/filePathHelper'
import { UseContextMenu } from '@/utils/contextMenuHelper'
import { breakpointsTailwind, useBreakpoints } from '@vueuse/core'

const message = useMessage()
const router = useRouter()
const { refMenu, openContextMenu } = UseContextMenu()
const currentPage = ref(1)
const pageSize = ref(12)
const breakpoints = useBreakpoints(breakpointsTailwind)

const search = ref('')
const dateModel = ref(0)
const mangaList = ref<LatestManga[]>([])
const mangaView = computed<LatestManga[]>(() => {
  let mangaView = mangaList.value
  // filtered mangaList
  if (search.value) {
    // filter by search
    mangaView = mangaList.value.filter(m =>
      m.main_title.toLowerCase().includes(search.value.toLowerCase()),
    )
  } else if (dateModel.value > 0) {
    // filter by date
    const filterDate = getRelativeDate(dateModel.value)
    mangaView = mangaList.value.filter(m =>
      isSameDate(new Date(m.download_time), filterDate),
    )
  } else {
    mangaView = mangaList.value
  }
  // pagination
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return mangaView.slice(start, end)
})
const totalItems = computed(() => {
  if (search.value) {
    return mangaList.value.filter(m =>
      m.main_title.toLowerCase().includes(search.value.toLowerCase()),
    ).length
  } else if (dateModel.value > 0) {
    const filterDate = getRelativeDate(dateModel.value)
    return mangaList.value.filter(m =>
      isSameDate(new Date(m.download_time), filterDate),
    ).length
  } else {
    return mangaList.value.length
  }
})
const totalPages = computed(() => Math.ceil(totalItems.value / pageSize.value))
// const mangaView = computed<LatestManga[]>(() => {
//   const r = repeatArray(mangaList.value, 5)
watch(totalPages, tp => {
  const maxPage = tp > 0 ? tp : 1
  if (currentPage.value > maxPage) {
    currentPage.value = maxPage
  }
})
// const mangaView = computed<LatestManga[]>(() => {
//   const r = repeatArray(mangaList.value, 5)
//   const start = (currentPage.value - 1) * pageSize.value
//   const end = start + pageSize.value
//   return r.slice(start, end)
// })
watch(search, v => {
  if (v) {
    dateModel.value = 0
  }
})
watch(dateModel, v => {
  if (v > 0) {
    search.value = ''
  }
})
const is2xlUp = breakpoints.greaterOrEqual('2xl')
watch(
  is2xlUp,
  b => {
    pageSize.value = b ? 30 : 12
  },
  { immediate: true },
)

const fetchMangaList = async () => {
  try {
    mangaList.value = await DatabaseService.GetLatestManga()
  } catch (error) {
    message.error(`Error fetching manga list : ${error}`)
  }
}
fetchMangaList()

const clickManga = (manga_id: number) => {
  router.push(`/chapters/${manga_id}`)
}

const clickChapter = (manga_id: number, chapter_id: number) => {
  router.push(`/read/${manga_id}/${chapter_id}`)
}

onMounted(() => {
  refMenu.value
})
/* ========= HELPER FUNCTION ============== */
const isToday = (date_string: string) => {
  const date = new Date(date_string)
  const today = new Date()
  return (
    date.getDate() === today.getDate() &&
    date.getMonth() === today.getMonth() &&
    date.getFullYear() === today.getFullYear()
  )
}

function getRelativeDate(n: number) {
  const date = new Date() // sekarang
  date.setDate(date.getDate() - (n - 1))
  return date
}

function isSameDate(d1: Date, d2: Date) {
  return (
    d1.getFullYear() === d2.getFullYear() &&
    d1.getMonth() === d2.getMonth() &&
    d1.getDate() === d2.getDate()
  )
}

const formatter = new Intl.DateTimeFormat('en-GB', {
  day: '2-digit',
  month: '2-digit',
})
const formatDate = (date_string: string) => {
  const date = new Date(date_string)
  return formatter.format(date)
}
function repeatArray<T>(arr: T[], times: number): T[] {
  const result: T[] = []
  for (let i = 0; i < times; i++) {
    result.push(...arr)
  }
  return result
}
</script>

<style>
.bg-gradient {
  padding-top: 3rem;
  color: white;
  background: linear-gradient(
    to top,
    rgba(0, 0, 0, 0.9) 0%,
    rgba(0, 0, 0, 0.6) 60%,
    transparent 100%
  );
}

.today-highlight {
  box-shadow: 0 0 10px 2px rgba(255, 165, 0, 0.6); /* Orange glow */
  border: 2px solid rgba(255, 165, 0, 0.8); /* Solid orange border */
  transform: scale(1.02); /* Sedikit membesar */
  z-index: 10;
  transition: all 0.3s ease-in-out;
}

/* Animasi opsional: pulse halus */
@keyframes pulse-orange {
  0% {
    box-shadow: 0 0 0 0 rgba(255, 165, 0, 0.4);
  }
  70% {
    box-shadow: 0 0 10px 4px rgba(255, 165, 0, 0);
  }
  100% {
    box-shadow: 0 0 0 0 rgba(255, 165, 0, 0);
  }
}
</style>
