<template>
  <div>
    <PageHeader title="工具" subtitle="工具定义与启用状态">
      <a-button type="primary" @click="openRegister">注册工具</a-button>
    </PageHeader>

    <a-card>
      <a-table
        :columns="columns"
        :data-source="rows"
        :loading="loading"
        :pagination="pagination"
        row-key="ToolID"
        @change="onTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'SourceKind'">
            <StatusTag :status="record.SourceKind" />
          </template>
          <template v-else-if="column.key === 'Enabled'">
            <a-switch
              :checked="record.Enabled"
              @change="(checked: boolean) => toggle(record, checked)"
            />
          </template>
          <template v-else-if="column.key === 'actions'">
            <a-button size="small" @click="showDetail(record.ToolID)">详情</a-button>
          </template>
        </template>
      </a-table>
    </a-card>

    <a-modal v-model:open="registerOpen" title="注册工具" :confirm-loading="submitting" @ok="submit">
      <a-form layout="vertical">
        <a-form-item label="名称" required>
          <a-input v-model:value="form.name" placeholder="唯一名称" />
        </a-form-item>
        <a-form-item label="描述">
          <a-input v-model:value="form.description" />
        </a-form-item>
        <a-form-item label="处理器名称" required>
          <a-input v-model:value="form.handlerName" placeholder="如 calculator / http_request" />
        </a-form-item>
        <a-form-item label="来源类型">
          <a-select v-model:value="form.sourceKind">
            <a-select-option value="builtin">内置</a-select-option>
            <a-select-option value="http">HTTP</a-select-option>
            <a-select-option value="mcp">MCP</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item v-if="form.sourceKind !== 'builtin'" label="来源地址">
          <a-input v-model:value="form.sourceUri" />
        </a-form-item>
        <a-form-item label="超时（秒）">
          <a-input-number v-model:value="form.timeoutSeconds" :min="1" style="width: 100%" />
        </a-form-item>
        <a-form-item label="参数 Schema（JSON）" required>
          <a-textarea v-model:value="form.schema" :rows="6" />
        </a-form-item>
      </a-form>
    </a-modal>

    <a-modal v-model:open="detailOpen" title="工具详情" :footer="null" width="640px">
      <a-descriptions v-if="detail" bordered :column="1">
        <a-descriptions-item label="名称">{{ detail.Name }}</a-descriptions-item>
        <a-descriptions-item label="描述">{{ detail.Description || '-' }}</a-descriptions-item>
        <a-descriptions-item label="处理器">{{ detail.Handler }}</a-descriptions-item>
        <a-descriptions-item label="来源">{{ detail.SourceKind }} {{ detail.SourceURI }}</a-descriptions-item>
        <a-descriptions-item label="版本">{{ detail.Version }}</a-descriptions-item>
        <a-descriptions-item label="超时">{{ detail.TimeoutSeconds }} 秒</a-descriptions-item>
        <a-descriptions-item label="Schema">
          <pre class="pre-wrap">{{ JSON.stringify(detail.Schema, null, 2) }}</pre>
        </a-descriptions-item>
      </a-descriptions>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { message } from 'ant-design-vue'
import { computed, onMounted, reactive, ref } from 'vue'

import { toolsApi } from '@/api/tools'
import PageHeader from '@/components/PageHeader.vue'
import StatusTag from '@/components/StatusTag.vue'
import type { JSONSchema, ToolDetail, ToolSummary } from '@/types'

const columns = [
  { title: '名称', dataIndex: 'Name', key: 'Name' },
  { title: '描述', dataIndex: 'Description', key: 'Description' },
  { title: '来源', dataIndex: 'SourceKind', key: 'SourceKind', width: 120 },
  { title: '版本', dataIndex: 'Version', key: 'Version', width: 80 },
  { title: '启用', dataIndex: 'Enabled', key: 'Enabled', width: 90 },
  { title: '操作', key: 'actions', width: 100 },
]

const rows = ref<ToolSummary[]>([])
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
    const res = await toolsApi.list({ offset: query.offset, limit: query.limit })
    rows.value = res.Tools ?? []
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

async function toggle(record: ToolSummary, checked: boolean) {
  await toolsApi.enable(record.ToolID, checked)
  record.Enabled = checked
  message.success(checked ? '已启用' : '已停用')
}

const registerOpen = ref(false)
const submitting = ref(false)
const form = reactive({
  name: '',
  description: '',
  handlerName: '',
  sourceKind: 'builtin',
  sourceUri: '',
  timeoutSeconds: 30,
  schema: '{"type":"object","properties":{},"required":[]}',
})

function openRegister() {
  registerOpen.value = true
}

async function submit() {
  let schema: JSONSchema
  try {
    schema = JSON.parse(form.schema) as JSONSchema
  } catch {
    message.error('Schema 不是合法 JSON')
    return
  }
  submitting.value = true
  try {
    await toolsApi.register({
      name: form.name,
      description: form.description,
      handlerName: form.handlerName,
      sourceKind: form.sourceKind,
      sourceUri: form.sourceUri,
      timeoutSeconds: form.timeoutSeconds,
      schema,
    })
    message.success('注册成功')
    registerOpen.value = false
    await load()
  } finally {
    submitting.value = false
  }
}

const detailOpen = ref(false)
const detail = ref<ToolDetail | null>(null)
async function showDetail(id: string) {
  detail.value = await toolsApi.get(id)
  detailOpen.value = true
}

onMounted(load)
</script>
