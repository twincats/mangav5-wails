<template>
  <div>
    <div>Home view Berisi list manga untuk di read</div>
    <div>
      <n-image width="200" :src="imageUrl" preview-disabled />
      <n-image width="200" :src="imgcover" preview-disabled />
    </div>
  </div>
</template>
<!-- Ore ga Kokuhaku Sarete Kara, Ojou no Yousu ga Okashii/4/01.jpg -->
<script setup lang="ts">
import { ref } from 'vue'

const imageUrl = ref(
  '/filemanga/Ore ga Kokuhaku Sarete Kara, Ojou no Yousu ga Okashii/4/02.jpg',
)

const imgcover = ref(
  '/filemanga/Ore ga Kokuhaku Sarete Kara, Ojou no Yousu ga Okashii/cover.webp',
)
</script>
