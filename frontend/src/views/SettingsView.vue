<template>
  <div>
    <div>Settings view</div>
    <n-button type="primary" @click="$router.push('/')">Goto Home</n-button>
    <div class="h-[calc(100vh-170px)] mt-4">
      <MonacoEditor
        v-model="code"
        language="json"
        theme="vs-dark"
        :formatOnLoad="true"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import MonacoEditor from '@/components/MonacoEditor.vue'

const code = ref('')
</script>
