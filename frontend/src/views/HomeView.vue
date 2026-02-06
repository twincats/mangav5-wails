<template>
  <div>
    <div>
      <!-- <n-image width="200" :src="imageUrl" preview-disabled />
      <n-image width="200" :src="imgcover" preview-disabled /> -->
      <div class="grid grid-cols-6 xl:grid-cols-10 gap-2">
        <div
          v-for="(m, index) in mangaList"
          class="relative group select-none rounded-1 transition-all duration-300"
          :class="{ 'today-highlight': isToday(m.download_time) }"
          @click="clickManga(m.manga_id)"
          :style="{
            '--custom-contextmenu': 'home-menu',
            '--custom-contextmenu-data': m.manga_id,
          }"
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
              @click.stop="clickChapter(m.chapter_id)"
              class="absolute bottom-0 left-0 right-0 bg-white/50 text-center rounded-b-1 text-black transition-all duration-300 ease-in-out group-hover:opacity-100 opacity-0"
            >
              <strong>Chapter {{ m.chapter_number }}</strong>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
<!-- Ore ga Kokuhaku Sarete Kara, Ojou no Yousu ga Okashii/4/01.jpg -->
<script setup lang="ts">
import { DatabaseService } from '../../bindings/mangav5/services'
import { LatestManga } from '../../bindings/mangav5/internal/models'
import { ImagePath } from '@/utils/filePathHelper'

const message = useMessage()
const router = useRouter()

const mangaList = ref<LatestManga[]>([])
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

const clickChapter = (chapter_id: number) => {
  router.push(`/read/${chapter_id}`)
}

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

const formatter = new Intl.DateTimeFormat('en-GB', {
  day: '2-digit',
  month: '2-digit',
})
const formatDate = (date_string: string) => {
  const date = new Date(date_string)
  return formatter.format(date)
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
