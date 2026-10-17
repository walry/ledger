import { defineStore } from 'pinia'
import { ref } from 'vue'

import { agentsApi } from '@/api/agents'
import type { AgentSummary } from '@/types'

/** 缓存智能体列表，供各页面下拉选择复用。 */
export const useAgentsStore = defineStore('agents', () => {
  const items = ref<AgentSummary[]>([])
  const loading = ref(false)
  const loaded = ref(false)

  async function fetchAll(force = false): Promise<AgentSummary[]> {
    if (loaded.value && !force) return items.value
    loading.value = true
    try {
      const res = await agentsApi.list({ offset: 0, limit: 200 })
      items.value = res.Agents ?? []
      loaded.value = true
    } finally {
      loading.value = false
    }
    return items.value
  }

  return { items, loading, loaded, fetchAll }
})
