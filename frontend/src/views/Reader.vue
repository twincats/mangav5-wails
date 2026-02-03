<template>
  <div>
    <h2>Reader {{ route.params.chapterId }}</h2>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()

onMounted(() => {
  console.log('Reader mounted')
})
</script>
