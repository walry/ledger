<template>
  <div>
    <PageHeader title="会话" subtitle="运行态的对话记录">
      <a-button type="primary" @click="openCreate">创建会话</a-button>
    </PageHeader>

    <a-card>
      <a-table
        :columns="columns"
        :data-source="rows"
        :loading="loading"
        :pagination="pagination"
        row-key="ConversationID"
        @change="onTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'Status'">
            <StatusTag :status="record.Status" />
          </template>
          <template v-else-if="column.key === 'CreatedAt'">
            {{ formatDateTime(record.CreatedAt) }}
          </template>
          <template v-else-if="column.key === 'actions'">
            <router-link :to="`/conversations/${record.ConversationID}`">
              <a-button size="small">详情</a-button>
            </router-link>
          </template>
        </template>
      </a-table>
    </a-card>

    <a-modal v-model:open="createOpen" title="创建会话" :confirm-loading="submitting" @ok="submit">
      <a-form layout="vertical">
        <a-form-item label="智能体" required>
          <a-select
            v-model:value="form.agentId"
            :options="agentOptions"
            :loading="agentStore.loading"
            placeholder="选择智能体"
          />
        </a-form-item>
        <a-form-item label="标题">
          <a-input v-model:value="form.title" placeholder="可选" />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { message } from 'ant-design-vue'
import { computed, onMounted, reactive, ref } from 'vue'

import { conversationsApi } from '@/api/conversations'
import PageHeader from '@/components/PageHeader.vue'
import StatusTag from '@/components/StatusTag.vue'
import { useAgentsStore } from '@/stores/agents'
import type { ConversationSummary } from '@/types'
import { formatDateTime } from '@/utils/format'

const agentStore = useAgentsStore()

const columns = [
  { title: '标题', dataIndex: 'Title', key: 'Title' },
  { title: '智能体', dataIndex: 'AgentID', key: 'AgentID' },
  { title: '智能体版本', dataIndex: 'AgentVersion', key: 'AgentVersion', width: 120 },
  { title: '状态', dataIndex: 'Status', key: 'Status', width: 100 },
  { title: '创建时间', dataIndex: 'CreatedAt', key: 'CreatedAt', width: 180 },
  { title: '操作', key: 'actions', width: 100 },
]

const rows = ref<ConversationSummary[]>([])
const loading = ref(false)
const total = ref(0)
const query = reactive({ offset: 0, limit: 10 })

const pagination = computed(() => ({
  current: Math.floor(query.offset / query.limit) + 1,
  pageSize: query.limit,
  total: total.value,
  showSizeChanger: true,
}))

const agentOptions = computed(() =>
  agentStore.items.map((a) => ({ label: `${a.Name}（v${a.Version}）`, value: a.AgentID })),
)

async function load() {
  loading.value = true
  try {
    const res = await conversationsApi.list({ offset: query.offset, limit: query.limit })
    rows.value = res.Conversations ?? []
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

const createOpen = ref(false)
const submitting = ref(false)
const form = reactive({ agentId: undefined as string | undefined, title: '' })

function openCreate() {
  createOpen.value = true
}

async function submit() {
  if (!form.agentId) {
    message.warning('请选择智能体')
    return
  }
  submitting.value = true
  try {
    await conversationsApi.create({ agentId: form.agentId, title: form.title })
    message.success('已创建会话')
    createOpen.value = false
    form.agentId = undefined
    form.title = ''
    await load()
  } finally {
    submitting.value = false
  }
}

onMounted(async () => {
  await Promise.all([load(), agentStore.fetchAll()])
})
</script>
