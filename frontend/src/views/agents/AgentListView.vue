<template>
  <div>
    <PageHeader title="智能体" subtitle="配置态的 Agent（提示词 / 技能 / 工具 / 模型 / 知识库）">
      <a-button type="primary" @click="openCreate">创建智能体</a-button>
    </PageHeader>

    <a-card>
      <a-table
        :columns="columns"
        :data-source="rows"
        :loading="loading"
        :pagination="pagination"
        row-key="AgentID"
        @change="onTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'Status'">
            <StatusTag :status="record.Status" />
          </template>
          <template v-else-if="column.key === 'actions'">
            <router-link :to="`/agents/${record.AgentID}`">
              <a-button size="small">详情</a-button>
            </router-link>
          </template>
        </template>
      </a-table>
    </a-card>

    <a-modal
      v-model:open="createOpen"
      title="创建智能体（草稿）"
      width="640px"
      :confirm-loading="submitting"
      @ok="submit"
    >
      <a-form layout="vertical">
        <a-form-item label="名称" required>
          <a-input v-model:value="form.name" />
        </a-form-item>
        <a-form-item label="描述">
          <a-input v-model:value="form.description" />
        </a-form-item>
        <a-form-item label="系统提示词" required>
          <a-textarea v-model:value="form.systemPrompt" :rows="4" />
        </a-form-item>
        <a-form-item label="模型配置" required>
          <a-select
            v-model:value="form.modelConfigId"
            placeholder="选择模型配置"
            :options="modelOptions"
            :loading="modelStore.loading"
          />
        </a-form-item>
        <a-form-item label="关联知识库">
          <a-select
            v-model:value="form.knowledgeBaseIds"
            mode="multiple"
            placeholder="可多选"
            :options="kbOptions"
          />
        </a-form-item>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="最大步数">
              <a-input-number v-model:value="form.maxSteps" :min="1" :max="20" style="width: 100%" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="检索 TopK">
              <a-input-number v-model:value="form.retrievalTopK" :min="1" :max="20" style="width: 100%" />
            </a-form-item>
          </a-col>
        </a-row>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { message } from 'ant-design-vue'
import { computed, onMounted, reactive, ref } from 'vue'

import { agentsApi } from '@/api/agents'
import { knowledgeApi } from '@/api/knowledge'
import PageHeader from '@/components/PageHeader.vue'
import StatusTag from '@/components/StatusTag.vue'
import { useModelConfigsStore } from '@/stores/modelConfigs'
import type { AgentSummary } from '@/types'

const modelStore = useModelConfigsStore()

const columns = [
  { title: '名称', dataIndex: 'Name', key: 'Name' },
  { title: '状态', dataIndex: 'Status', key: 'Status', width: 120 },
  { title: '版本', dataIndex: 'Version', key: 'Version', width: 80 },
  { title: '模型', dataIndex: 'Model', key: 'Model' },
  { title: '操作', key: 'actions', width: 100 },
]

const rows = ref<AgentSummary[]>([])
const loading = ref(false)
const total = ref(0)
const query = reactive({ offset: 0, limit: 10 })

const pagination = computed(() => ({
  current: Math.floor(query.offset / query.limit) + 1,
  pageSize: query.limit,
  total: total.value,
  showSizeChanger: true,
}))

async function load() {
  loading.value = true
  try {
    const res = await agentsApi.list({ offset: query.offset, limit: query.limit })
    rows.value = res.Agents ?? []
    total.value = res.Total ?? 0
  } finally {
    loading.value = false
  }
}

function onTableChange(p: { current?: number; pageSize?: number }) {
  const pageSize = p.pageSize ?? query.limit
  query.limit = pageSize
  query.offset = ((p.current ?? 1) - 1) * pageSize
  void load()
}

// 下拉选项
const modelOptions = computed(() =>
  modelStore.items.map((m) => ({ label: `${m.Name}（${m.Model}）`, value: m.ConfigID })),
)
const kbOptions = ref<{ label: string; value: string }[]>([])

async function loadOptions() {
  await modelStore.fetchAll()
  const res = await knowledgeApi.list({ limit: 200 })
  kbOptions.value = (res.KnowledgeBases ?? []).map((k) => ({ label: k.Name, value: k.KnowledgeBaseID }))
}

// 创建
const createOpen = ref(false)
const submitting = ref(false)
const form = reactive({
  name: '',
  description: '',
  systemPrompt: '',
  modelConfigId: undefined as string | undefined,
  knowledgeBaseIds: [] as string[],
  maxSteps: 6,
  retrievalTopK: 5,
})

function openCreate() {
  createOpen.value = true
}

async function submit() {
  if (!form.modelConfigId) {
    message.warning('请选择模型配置')
    return
  }
  submitting.value = true
  try {
    await agentsApi.create({
      name: form.name,
      description: form.description,
      systemPrompt: form.systemPrompt,
      modelConfigId: form.modelConfigId,
      knowledgeBaseIds: form.knowledgeBaseIds,
      maxSteps: form.maxSteps,
      retrievalTopK: form.retrievalTopK,
    })
    message.success('已创建草稿')
    createOpen.value = false
    form.name = ''
    form.description = ''
    form.systemPrompt = ''
    form.modelConfigId = undefined
    form.knowledgeBaseIds = []
    await load()
  } finally {
    submitting.value = false
  }
}

onMounted(async () => {
  await Promise.all([load(), loadOptions()])
})
</script>
