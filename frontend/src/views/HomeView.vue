<template>
  <div>
    <div>Home view</div>
    <div>
      <n-space>
        <n-button
          type="primary"
          @click="message.success('primary button clicked')"
          >Notif</n-button
        >
        <n-button type="primary" @click="$router.push('/settings')"
          >Goto Settings</n-button
        >
      </n-space>
    </div>
    <div style="margin-top: 20px">
      <n-space vertical>
        <n-date-picker type="date" />
        <n-color-picker />
      </n-space>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useMessage } from 'naive-ui'
const message = useMessage()
</script>
