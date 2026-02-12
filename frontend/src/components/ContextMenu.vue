<template>
  <div
    ref="cmenu"
    class="context-menu"
    v-if="data.show"
    :style="{
      left: data.x + 'px',
      top: data.y - 10 + 'px',
      minWidth: prop.width + 'px',
    }"
  >
    <ul class="">
      <slot :item="extData"></slot>
    </ul>
  </div>
</template>

<script setup lang="ts">
import { useWindowSize, useScroll, onClickOutside } from '@vueuse/core'
// setup props
interface Props {
  width?: number
}

// defineProps with default value need reactivityTransform: true
const prop = withDefaults(defineProps<Props>(), {
  width: 230,
})

// setup emit
// const emit = defineEmits<{
//   (e: 'close'): void
// }>()
const { width: winWidth, height: winHeight } = useWindowSize()
const cmenu = ref<HTMLElement | null>(null)
const extData = ref<any>(null)
const data = reactive({
  show: false,
  x: 0,
  y: 0,
})
const el = document.getElementById('main')
const scroll = useScroll(el)
const close = () => (data.show = false)
onClickOutside(cmenu, () => close())
watch(
  () => scroll.isScrolling.value,
  scrollStat => {
    if (scrollStat) {
      close()
    }
  },
)

const open = (event: MouseEvent, dataExt: any = null) => {
  const limitX = winWidth.value - event.x
  const limitY = winHeight.value - event.y
  data.show = true
  data.x = event.x
  data.y = event.y
  extData.value = dataExt
  if (prop.width > limitX) {
    data.x = winWidth.value - prop.width
  }
  if (cmenu.value) {
    const height = cmenu.value.offsetHeight
    if (height > limitY) {
      data.y = winHeight.value - height
    }
  }
}
export interface ExposedProps {
  open: (event: MouseEvent, dataExt: any) => void // Method to toggle the menu
  close: () => void // Expose the title reactive reference
}
// expose tobe available in $refs instance
defineExpose({
  open,
  close,
})
</script>

<style>
:root {
  --app-context: #292929;
  --app-context-light: #313131;
  --app-context-divider: #535353;
}
.context-menu {
  background-color: var(--app-context);
  position: fixed;
  border-radius: 0.5rem;
  box-shadow:
    0 4px 6px rgba(0, 0, 0, 0.1),
    0 1px 3px rgba(0, 0, 0, 0.08);
  z-index: 50;
}

.context-menu ul {
  list-style: none;
  margin: 0;
  padding: 0;
}

.context-menu ul li {
  user-select: none;
  margin: 0.25rem;
  padding: 0.25rem 0.75rem;
  display: flex;
  border-radius: 0.25rem;
  font-size: 0.875rem;
  line-height: 1.25rem;
  font-family:
    'Segoe UI',
    system-ui,
    -apple-system,
    BlinkMacSystemFont,
    sans-serif;
  color: #fff;
  justify-content: space-between;
}

.context-menu ul li:hover {
  background-color: var(--app-context-light);
}

.context-menu ul .divider {
  height: 1px;
  background-color: var(--app-context-divider);
}

.context-menu ul li small {
  font-size: 0.75rem;
  margin-right: 1.5rem;
}
</style>
