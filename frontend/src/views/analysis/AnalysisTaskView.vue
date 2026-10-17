<template>
  <div>
    <PageHeader title="分析任务" subtitle="数字员工：面向报告的异步多步分析">
      <a-button @click="load">刷新</a-button>
      <a-button type="primary" @click="submitOpen = true">提交任务</a-button>
    </PageHeader>

    <a-card>
      <a-table
        :columns="columns"
        :data-source="rows"
        :loading="loading"
        :pagination="pagination"
        row-key="TaskID"
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
            <a-space>
              <a-button size="small" :disabled="record.Status === 'running'" @click="run(record.TaskID)">
                执行
              </a-button>
              <a-button size="small" @click="viewResult(record.TaskID)">结果</a-button>
            </a-space>
          </template>
        </template>
      </a-table>
    </a-card>

    <a-modal v-model:open="submitOpen" title="提交分析任务" :confirm-loading="submitting" @ok="submit">
      <a-form layout="vertical">
        <a-form-item label="标题">
          <a-input v-model:value="form.title" placeholder="可选" />
        </a-form-item>
        <a-form-item label="问题" required>
          <a-textarea v-model:value="form.question" :rows="3" />
        </a-form-item>
        <a-form-item label="智能体" required>
          <a-select
            v-model:value="form.agentId"
            :options="agentOptions"
            :loading="agentStore.loading"
            placeholder="选择智能体"
          />
        </a-form-item>
        <a-form-item label="关联报告">
          <a-select
            v-model:value="form.reportIds"
            mode="multiple"
            :options="reportOptions"
            placeholder="可多选"
          />
        </a-form-item>
      </a-form>
    </a-modal>

    <a-modal v-model:open="resultOpen" title="分析结果" :footer="null" width="720px">
      <a-descriptions v-if="result" bordered :column="1">
        <a-descriptions-item label="标题">{{ result.Title }}</a-descriptions-item>
        <a-descriptions-item label="状态"><StatusTag :status="result.Status" /></a-descriptions-item>
        <a-descriptions-item label="会话">{{ result.ConversationID || '-' }}</a-descriptions-item>
        <a-descriptions-item label="关联报告">
          <a-space v-if="result.ReportIDs?.length" wrap>
            <a-tag v-for="r in result.ReportIDs" :key="r">{{ r }}</a-tag>
          </a-space>
          <span v-else class="muted">-</span>
        </a-descriptions-item>
        <a-descriptions-item v-if="result.Failure" label="失败原因">{{ result.Failure }}</a-descriptions-item>
        <a-descriptions-item label="结果">
          <div class="pre-wrap">{{ result.Result || '-' }}</div>
        </a-descriptions-item>
      </a-descriptions>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { message } from 'ant-design-vue'
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'

import { analysisApi } from '@/api/analysis'
import { reportsApi } from '@/api/reports'
import PageHeader from '@/components/PageHeader.vue'
import StatusTag from '@/components/StatusTag.vue'
import { useAgentsStore } from '@/stores/agents'
import type { AnalysisResult, AnalysisTaskSummary } from '@/types'
import { formatDateTime } from '@/utils/format'

const agentStore = useAgentsStore()

const columns = [
  { title: '标题', dataIndex: 'Title', key: 'Title' },
  { title: '状态', dataIndex: 'Status', key: 'Status', width: 110 },
  { title: '会话', dataIndex: 'ConversationID', key: 'ConversationID' },
  { title: '创建时间', dataIndex: 'CreatedAt', key: 'CreatedAt', width: 180 },
  { title: '操作', key: 'actions', width: 160 },
]

const rows = ref<AnalysisTaskSummary[]>([])
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
const reportOptions = ref<{ label: string; value: string }[]>([])

async function load() {
  loading.value = true
  try {
    const res = await analysisApi.list({ offset: query.offset, limit: query.limit })
    rows.value = res.Tasks ?? []
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

async function run(id: string) {
  message.loading({ content: '执行中…', key: 'run' })
  try {
    await analysisApi.run(id)
    message.success({ content: '执行完成', key: 'run' })
  } finally {
    await load()
  }
}

const submitOpen = ref(false)
const submitting = ref(false)
const form = reactive({
  title: '',
  question: '',
  agentId: undefined as string | undefined,
  reportIds: [] as string[],
})

async function submit() {
  if (!form.agentId || !form.question.trim()) {
    message.warning('请填写问题并选择智能体')
    return
  }
  submitting.value = true
  try {
    await analysisApi.submit({
      title: form.title,
      question: form.question,
      agentId: form.agentId,
      reportIds: form.reportIds,
    })
    message.success('已提交，后台将自动执行')
    submitOpen.value = false
    form.title = ''
    form.question = ''
    form.reportIds = []
    await load()
  } finally {
    submitting.value = false
  }
}

const resultOpen = ref(false)
const result = ref<AnalysisResult | null>(null)
async function viewResult(id: string) {
  result.value = await analysisApi.get(id)
  resultOpen.value = true
}

// 存在排队/执行中的任务时轮询刷新。
let timer: number | undefined
function startPolling() {
  timer = window.setInterval(() => {
    const active = rows.value.some((t) => t.Status === 'queued' || t.Status === 'running')
    if (active && !loading.value) void load()
  }, 4000)
}

onMounted(async () => {
  await Promise.all([load(), agentStore.fetchAll()])
  const reports = await reportsApi.list({ limit: 200 })
  reportOptions.value = (reports.Reports ?? []).map((r) => ({ label: r.Name, value: r.ReportID }))
  startPolling()
})

onBeforeUnmount(() => {
  if (timer) window.clearInterval(timer)
})
</script>
