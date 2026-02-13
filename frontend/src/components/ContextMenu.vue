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
  width: 200,
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
  /* teks utama */
  --app-red: #e53935;
  --app-orange: #ff9800;
  --app-green: #4caf50;

  /* background hover */
  --app-red-hover-bg: rgba(229, 57, 53, 0.18);
  --app-orange-hover-bg: rgba(255, 152, 0, 0.18);
  --app-green-hover-bg: rgba(76, 175, 80, 0.18);
}
.context-menu {
  background-color: var(--app-context);
  position: fixed;
  border-radius: 0.5rem;
  box-shadow:
    0 4px 6px rgba(0, 0, 0, 0.1),
    0 1px 3px rgba(0, 0, 0, 0.08);
  z-index: 50;
  width: max-content;
}

.context-menu ul {
  list-style: none;
  margin: 0;
  padding: 0;
}

.context-menu ul li {
  user-select: none;
  margin: 0.25rem;
  padding: 0.25rem 2rem 0.25rem 0.75rem;
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

.context-menu ul li.disabled {
  color: #999;
  /* cursor: not-allowed; */
}

.context-menu ul li:not(.disabled):hover {
  background-color: var(--app-context-light);
}

.context-menu ul .divider {
  height: 1px;
  background-color: var(--app-context-divider);
}

.context-menu ul li small {
  font-size: 0.75rem;
  min-width: 5rem;
  text-align: right;
}

.context-menu ul li.red {
  color: var(--app-red);
}
.context-menu ul li.red:hover {
  background-color: var(--app-red-hover-bg);
}

.context-menu ul li.orange {
  color: var(--app-orange);
}
.context-menu ul li.orange:hover {
  background-color: var(--app-orange-hover-bg);
}

.context-menu ul li.green {
  color: var(--app-green);
}
.context-menu ul li.green:hover {
  background-color: var(--app-green-hover-bg);
}
</style>
