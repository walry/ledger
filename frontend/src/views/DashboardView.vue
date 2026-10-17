<template>
  <div>
    <PageHeader title="概览" subtitle="智能财务报告分析应用" />

    <a-spin :spinning="loading">
      <a-row :gutter="[16, 16]">
        <a-col v-for="card in cards" :key="card.key" :xs="12" :md="8" :lg="6">
          <router-link :to="card.to" class="stat-link">
            <a-card hoverable>
              <a-statistic :title="card.title" :value="card.value" />
            </a-card>
          </router-link>
        </a-col>
      </a-row>
    </a-spin>

    <a-card title="快速入口" style="margin-top: 16px">
      <a-space wrap>
        <router-link to="/reports"><a-button>上传报告</a-button></router-link>
        <router-link to="/agents"><a-button>创建智能体</a-button></router-link>
        <router-link to="/qa"><a-button type="primary">开始问答</a-button></router-link>
        <router-link to="/analysis"><a-button>提交分析任务</a-button></router-link>
      </a-space>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'

import { agentsApi } from '@/api/agents'
import { analysisApi } from '@/api/analysis'
import { conversationsApi } from '@/api/conversations'
import { knowledgeApi } from '@/api/knowledge'
import { modelConfigsApi } from '@/api/modelConfigs'
import { reportsApi } from '@/api/reports'
import { toolsApi } from '@/api/tools'
import PageHeader from '@/components/PageHeader.vue'

const loading = ref(false)

const cards = reactive([
  { key: 'reports', title: '报告', value: 0, to: '/reports' },
  { key: 'knowledge', title: '知识库', value: 0, to: '/knowledge' },
  { key: 'agents', title: '智能体', value: 0, to: '/agents' },
  { key: 'conversations', title: '会话', value: 0, to: '/conversations' },
  { key: 'analysis', title: '分析任务', value: 0, to: '/analysis' },
  { key: 'tools', title: '工具', value: 0, to: '/tools' },
  { key: 'modelConfigs', title: '模型配置', value: 0, to: '/model-configs' },
])

function setValue(key: string, value: number) {
  const card = cards.find((c) => c.key === key)
  if (card) card.value = value
}

onMounted(async () => {
  loading.value = true
  const tasks: Array<Promise<void>> = [
    reportsApi.list({ limit: 1 }).then((r) => setValue('reports', r.Total)),
    knowledgeApi.list({ limit: 1 }).then((r) => setValue('knowledge', r.Total)),
    agentsApi.list({ limit: 1 }).then((r) => setValue('agents', r.Total)),
    conversationsApi.list({ limit: 1 }).then((r) => setValue('conversations', r.Total)),
    analysisApi.list({ limit: 1 }).then((r) => setValue('analysis', r.Total)),
    toolsApi.list({ limit: 1 }).then((r) => setValue('tools', r.Total)),
    modelConfigsApi.list({ limit: 1 }).then((r) => setValue('modelConfigs', r.Total)),
  ]
  // 单个统计失败不应阻塞其它统计。
  await Promise.allSettled(tasks)
  loading.value = false
})
</script>

<style scoped>
.stat-link {
  display: block;
  color: inherit;
  text-decoration: none;
}
</style>
