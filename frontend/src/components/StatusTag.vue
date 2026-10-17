<template>
  <a-tag :color="color">{{ label }}</a-tag>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ status: string }>()

const COLORS: Record<string, string> = {
  // agent
  draft: 'default',
  published: 'green',
  archived: 'default',
  // report
  uploaded: 'blue',
  parsing: 'orange',
  parsed: 'green',
  // conversation / turn
  active: 'green',
  running: 'orange',
  completed: 'green',
  // analysis
  queued: 'default',
  succeeded: 'green',
  // document
  pending: 'default',
  indexed: 'green',
  // tool source
  builtin: 'blue',
  http: 'purple',
  mcp: 'gold',
  // common
  failed: 'red',
}

const color = computed(() => COLORS[props.status] ?? 'default')
const label = computed(() => props.status || '-')
</script>
