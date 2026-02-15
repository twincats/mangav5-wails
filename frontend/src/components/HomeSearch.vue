<template>
  <div
    class="bg-dark-600 rounded-1 p-2 text-white w-screen-lg text-center mx-auto mb-2 grid grid-rows-2 gap-2"
  >
    <div>
      <n-input-group>
        <n-input
          placeholder="Enter Search Query"
          :value="localSearch"
          @input="onInput"
          @blur="commitFromEvent"
          @keyup.enter="commitFromEvent"
        />
        <n-button tertiary type="primary">
          <template #icon><ClearOutlined /></template>
        </n-button>
        <n-button tertiary type="primary">
          <template #icon><SearchOutlined /></template>
        </n-button>
      </n-input-group>
    </div>
    <div>
      <n-radio-group
        v-model:value="dateModel"
        name="radiobuttongroup1"
        size="small"
      >
        <n-radio-button
          v-for="(label, i) in labelDate"
          :key="i"
          :value="i"
          :label="label"
        />
      </n-radio-group>
    </div>
  </div>
</template>

<script setup lang="ts">
import { SearchOutlined, ClearOutlined } from '@vicons/material'
interface ListDate {
  [index: number]: string
}

const search = defineModel<string>('search', { required: true, default: '' })
const dateModel = defineModel<number>('dateModel', {
  required: true,
  default: 0,
})
const props = withDefaults(defineProps<{ totalDate?: number }>(), {
  totalDate: 12,
})

const localSearch = ref(search.value)
const onInput = (v: string) => {
  localSearch.value = v
}
const commitFromEvent = () => {
  search.value = localSearch.value
}
watch(search, v => {
  localSearch.value = v
})

const labelDate = computed(() => {
  const data: ListDate = ['All', 'Today', 'Yesterday']
  const dateoptions = {
    month: 'short',
    day: 'numeric',
  } as Intl.DateTimeFormatOptions
  for (let i = 0; i < props.totalDate; i++) {
    const index = i + 3
    const d = new Date()
    d.setDate(d.getDate() - index + 1)
    data[index] = d.toLocaleString('id-ID', dateoptions)
  }

  return data
})
</script>
