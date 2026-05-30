<template>
  <div class="doc">
    <component :is="docComponent" v-if="docComponent" />
    <div v-else class="p-4">Dokumen tidak ditemukan.</div>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{ docId?: string | string[] }>()
const modules: Record<string, any> = import.meta.glob('../../docs/*.md', {
  eager: true,
})
const moduleMap: Record<string, any> = {}
Object.entries(modules).forEach(([key, mod]) => {
  const name = key.split('/').pop() || ''
  const base = name.replace(/\.md$/i, '').toLowerCase()
  moduleMap[base] = mod
})

const toBaseName = (name?: string | string[]) => {
  const n = Array.isArray(name) ? name[0] : name
  if (!n || !n.trim()) return 'settings'
  const cleaned = n.trim().replace(/[^a-zA-Z0-9-_\.]/g, '')
  if (!cleaned || cleaned === '.') return 'settings'
  return cleaned.endsWith('.md') ? cleaned.slice(0, -3) : cleaned
}

const docComponent = computed(() => {
  const base = toBaseName(props.docId).toLowerCase()
  const mod = moduleMap[base]
  if (mod && mod.default) return markRaw(mod.default)
  const fallbacks = ['settings', 'setting']
  for (const fb of fallbacks) {
    const fm = moduleMap[fb]
    if (fm && fm.default) return markRaw(fm.default)
  }
  return null
})
</script>

<style scoped>
.doc :deep(:not(pre) > code) {
  font-family:
    'Fira Code', ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas,
    'Liberation Mono', 'Courier New', monospace;
  font-size: 0.95em;
  padding: 0.15em 0.35em;
  border-radius: 4px;
  background: rgba(30, 30, 30, 0.9);
  border: 1px solid rgba(255, 255, 255, 0.12);
  color: #dcdcaa;
  user-select: text;
}

.doc :deep(pre > code) {
  user-select: text;
}
</style>
