<template>
  <div :style="{ width: normalizedWidth }">
    <div
      v-if="showHeader"
      :style="{ marginBottom: '4px', color: themeVars.textColor2 }"
    >
      {{ headerLabel }}: {{ processed }}/{{ safeTotal }} ({{
        processedPercent.toFixed(1)
      }}%)
    </div>

    <div
      :style="{
        position: 'relative',
        height: normalizedHeight,
        background: themeVars.dividerColor,
        borderRadius: '4px',
        overflow: 'hidden',
      }"
    >
      <div
        :style="{
          width: successPercent + '%',
          backgroundColor: themeVars.successColor,
          height: '100%',
          position: 'absolute',
          left: '0',
          top: '0',
        }"
      />
      <div
        :style="{
          width: failPercent + '%',
          backgroundColor: themeVars.errorColor,
          height: '100%',
          position: 'absolute',
          left: successPercent + '%',
          top: '0',
        }"
      />
    </div>

    <div
      v-if="showFooter"
      :style="{
        marginTop: '4px',
        fontSize: '14px',
        display: 'flex',
        justifyContent: 'space-between',
        color: themeVars.textColor3,
      }"
    >
      <span>Success: {{ safeSuccess }}</span>
      <span>Failed: {{ safeFail }}</span>
      <span>Remaining: {{ remaining }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useThemeVars } from 'naive-ui'

const props = withDefaults(
  defineProps<{
    total: number
    success?: number
    fail?: number
    width?: number | string
    height?: number | string
    showHeader?: boolean
    showFooter?: boolean
    headerLabel?: string
  }>(),
  {
    success: 0,
    fail: 0,
    width: '100%',
    height: 24,
    showHeader: true,
    showFooter: true,
    headerLabel: 'Processed',
  },
)

const themeVars = useThemeVars()

const safeTotal = computed(() => Math.max(0, Number(props.total) || 0))
const safeSuccess = computed(() =>
  Math.min(Math.max(0, Number(props.success) || 0), safeTotal.value),
)
const safeFail = computed(() =>
  Math.min(
    Math.max(0, Number(props.fail) || 0),
    Math.max(0, safeTotal.value - safeSuccess.value),
  ),
)
const processed = computed(() => safeSuccess.value + safeFail.value)
const remaining = computed(() => Math.max(0, safeTotal.value - processed.value))

const toPercent = (value: number) =>
  safeTotal.value > 0 ? (value / safeTotal.value) * 100 : 0
const processedPercent = computed(() => toPercent(processed.value))
const successPercent = computed(() => toPercent(safeSuccess.value))
const failPercent = computed(() => toPercent(safeFail.value))

const normalizedWidth = computed(() => {
  return typeof props.width === 'number' ? `${props.width}px` : props.width
})
const normalizedHeight = computed(() => {
  return typeof props.height === 'number' ? `${props.height}px` : props.height
})
</script>
