<template>
  <div>
    <div>Home view Berisi list manga untuk di read</div>
    <div>
      <!-- <n-image width="200" :src="imageUrl" preview-disabled />
      <n-image width="200" :src="imgcover" preview-disabled /> -->
      <div class="grid grid-cols-5 gap-2">
        <n-card v-for="(m, index) in mangaList" :key="index">
          <template #header>
            <n-ellipsis>{{ m.main_title }}</n-ellipsis>
          </template>
          <template #cover>
            <n-image
              width="200"
              :src="`${ImagePath(m.main_title)}/cover`"
              preview-disabled
            />
          </template>
          <n-ellipsis style="max-width: 100px"
            >Chapter {{ m.chapter_number }}</n-ellipsis
          >
        </n-card>
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
// const imageUrl = ref(
//   '/filemanga/Ore ga Kokuhaku Sarete Kara, Ojou no Yousu ga Okashii/4/02.jpg',
// )

// const imgcover = ref(
//   '/filemanga/Ore ga Kokuhaku Sarete Kara, Ojou no Yousu ga Okashii/cover.webp',
// )
const mangaList = ref<LatestManga[]>([])
const fetchMangaList = async () => {
  try {
    mangaList.value = await DatabaseService.GetLatestManga()
  } catch (error) {
    message.error(`Error fetching manga list : ${error}`)
  }
}
fetchMangaList()
</script>
