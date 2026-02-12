import type { ExposedProps } from '@/components/ContextMenu.vue'
// context menu instance
export const UseContextMenu = () => {
  const refMenu = ref<ExposedProps | null>(null)
  const openContextMenu = (ev: MouseEvent, data: any = null) => {
    ev.preventDefault()
    refMenu.value?.open(ev, data)
  }
  const closeContextMenu = () => {
    refMenu.value?.close()
  }
  return {
    refMenu,
    openContextMenu,
    closeContextMenu,
  }
}
