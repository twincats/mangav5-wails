<template>
  <div>
    <h2>Chapter Pages {{ route.params.mangaId }}</h2>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()

onMounted(() => {
  console.log('Chapter mounted')
})
</script>
