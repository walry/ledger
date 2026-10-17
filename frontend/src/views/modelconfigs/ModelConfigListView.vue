<template>
  <div>
    <PageHeader title="模型配置" subtitle="大模型连接与调用参数">
      <a-button type="primary" @click="openCreate">新建配置</a-button>
    </PageHeader>

    <a-card>
      <a-table
        :columns="columns"
        :data-source="rows"
        :loading="loading"
        :pagination="pagination"
        row-key="ConfigID"
        @change="onTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'Enabled'">
            <a-tag :color="record.Enabled ? 'green' : 'default'">
              {{ record.Enabled ? '启用' : '停用' }}
            </a-tag>
          </template>
          <template v-else-if="column.key === 'actions'">
            <a-button size="small" @click="openEdit(record)">编辑</a-button>
          </template>
        </template>
      </a-table>
    </a-card>

    <a-modal
      v-model:open="modalOpen"
      :title="editing ? '编辑模型配置' : '新建模型配置'"
      :confirm-loading="submitting"
      @ok="submit"
    >
      <a-form layout="vertical">
        <template v-if="!editing">
          <a-form-item label="名称" required>
            <a-input v-model:value="form.name" />
          </a-form-item>
          <a-form-item label="供应商" required>
            <a-select v-model:value="form.provider">
              <a-select-option value="openai">OpenAI</a-select-option>
              <a-select-option value="anthropic">Anthropic</a-select-option>
              <a-select-option value="openai_compatible">OpenAI 兼容</a-select-option>
            </a-select>
          </a-form-item>
        </template>
        <a-form-item label="Base URL">
          <a-input v-model:value="form.baseUrl" placeholder="https://api.openai.com/v1" />
        </a-form-item>
        <a-form-item :label="editing ? 'API Key（留空表示不修改）' : 'API Key'">
          <a-input-password v-model:value="form.apiKey" />
        </a-form-item>
        <a-form-item label="模型名" required>
          <a-input v-model:value="form.model" placeholder="gpt-4o-mini" />
        </a-form-item>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="Temperature">
              <a-input-number v-model:value="form.temperature" :min="0" :max="2" :step="0.1" style="width: 100%" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="Top P">
              <a-input-number v-model:value="form.topP" :min="0" :max="1" :step="0.05" style="width: 100%" />
            </a-form-item>
          </a-col>
        </a-row>
        <a-form-item label="Max Tokens">
          <a-input-number v-model:value="form.maxTokens" :min="1" style="width: 100%" />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { message } from 'ant-design-vue'
import { computed, onMounted, reactive, ref } from 'vue'

import { modelConfigsApi } from '@/api/modelConfigs'
import PageHeader from '@/components/PageHeader.vue'
import type { ModelConfig, ModelParameters } from '@/types'

const columns = [
  { title: '名称', dataIndex: 'Name', key: 'Name' },
  { title: '供应商', dataIndex: 'Provider', key: 'Provider', width: 140 },
  { title: '模型', dataIndex: 'Model', key: 'Model' },
  { title: 'Base URL', dataIndex: 'BaseURL', key: 'BaseURL' },
  { title: '状态', key: 'Enabled', width: 90 },
  { title: '操作', key: 'actions', width: 100 },
]

const rows = ref<ModelConfig[]>([])
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
    const res = await modelConfigsApi.list({ offset: query.offset, limit: query.limit })
    rows.value = res.Configs ?? []
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

const modalOpen = ref(false)
const submitting = ref(false)
const editing = ref(false)
const editingId = ref('')
const form = reactive({
  name: '',
  provider: 'openai',
  baseUrl: '',
  apiKey: '',
  model: '',
  temperature: 0.3,
  topP: 1,
  maxTokens: 2048,
})

function resetForm() {
  form.name = ''
  form.provider = 'openai'
  form.baseUrl = ''
  form.apiKey = ''
  form.model = ''
  form.temperature = 0.3
  form.topP = 1
  form.maxTokens = 2048
}

function openCreate() {
  editing.value = false
  editingId.value = ''
  resetForm()
  modalOpen.value = true
}

function openEdit(record: ModelConfig) {
  editing.value = true
  editingId.value = record.ConfigID
  form.name = record.Name
  form.provider = record.Provider
  form.baseUrl = record.BaseURL
  form.apiKey = ''
  form.model = record.Model
  form.temperature = record.Parameters?.Temperature ?? 0.3
  form.topP = record.Parameters?.TopP ?? 1
  form.maxTokens = record.Parameters?.MaxTokens ?? 2048
  modalOpen.value = true
}

function buildParameters(): ModelParameters {
  return {
    Temperature: form.temperature,
    TopP: form.topP,
    MaxTokens: form.maxTokens,
    PresencePenalty: 0,
    FrequencyPenalty: 0,
    Stop: [],
  }
}

async function submit() {
  submitting.value = true
  try {
    if (editing.value) {
      await modelConfigsApi.update(editingId.value, {
        baseUrl: form.baseUrl,
        apiKey: form.apiKey || undefined,
        model: form.model,
        parameters: buildParameters(),
      })
      message.success('已更新')
    } else {
      await modelConfigsApi.create({
        name: form.name,
        provider: form.provider,
        baseUrl: form.baseUrl,
        apiKey: form.apiKey,
        model: form.model,
        parameters: buildParameters(),
      })
      message.success('已创建')
    }
    modalOpen.value = false
    await load()
  } finally {
    submitting.value = false
  }
}

onMounted(load)
</script>
