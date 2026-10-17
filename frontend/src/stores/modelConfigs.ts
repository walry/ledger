import { defineStore } from 'pinia'
import { ref } from 'vue'

import { modelConfigsApi } from '@/api/modelConfigs'
import type { ModelConfig } from '@/types'

/** 缓存模型配置列表，供智能体创建等场景下拉选择复用。 */
export const useModelConfigsStore = defineStore('modelConfigs', () => {
  const items = ref<ModelConfig[]>([])
  const loading = ref(false)
  const loaded = ref(false)

  async function fetchAll(force = false): Promise<ModelConfig[]> {
    if (loaded.value && !force) return items.value
    loading.value = true
    try {
      const res = await modelConfigsApi.list({ offset: 0, limit: 200 })
      items.value = res.Configs ?? []
      loaded.value = true
    } finally {
      loading.value = false
    }
    return items.value
  }

  return { items, loading, loaded, fetchAll }
})
