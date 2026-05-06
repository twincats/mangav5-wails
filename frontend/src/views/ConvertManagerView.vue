<template>
  <div>
    <div class="flex gap-2">
      <n-card class="w-xl" size="small" content-style="padding: 0.5rem;">
        <span>Quality :</span>
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
          <span>Resize :</span> <br />
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
    </div>
    <div class="w-full my-2">
      <n-input v-model:value="config.search" />
    </div>
    <div>
      <n-scrollbar style="height: 400px">
        {{ mangaList }}
      </n-scrollbar>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { DatabaseService } from 'bindings/mangav5/services'
import { MangaBasic } from 'bindings/mangav5/internal/models'
import { useMessage } from 'naive-ui'

const message = useMessage()
const config = ref({
  quality: 60,
  resize: 1000,
  delete: true,
  compress: true,
  status_resize: true,
  search: '',
})

const mangaList = ref<MangaBasic[]>([])
const getMangaList = async () => {
  try {
    mangaList.value = await DatabaseService.ListMangaBasic()
  } catch (error) {
    message.error(`Error fetching manga list : ${error}`)
  }
}

onMounted(() => {
  getMangaList()
})
</script>
