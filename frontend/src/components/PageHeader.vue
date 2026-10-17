<template>
  <div class="page-toolbar">
    <div>
      <h2 style="margin: 0">{{ title }}</h2>
      <div v-if="subtitle" class="muted">{{ subtitle }}</div>
    </div>
    <div class="page-actions">
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{ title: string; subtitle?: string }>()
</script>

<style scoped>
.page-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}
</style>
