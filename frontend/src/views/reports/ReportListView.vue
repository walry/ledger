<template>
  <div>
    <PageHeader title="报告管理" subtitle="上传、解析与查看财务报告">
      <a-button type="primary" @click="openUpload">
        <template #icon><UploadOutlined /></template>
        上传报告
      </a-button>
    </PageHeader>

    <a-card>
      <a-table
        :columns="columns"
        :data-source="rows"
        :loading="loading"
        :pagination="pagination"
        row-key="ReportID"
        @change="onTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'Status'">
            <StatusTag :status="record.Status" />
          </template>
          <template v-else-if="column.key === 'actions'">
            <a-space>
              <router-link :to="`/reports/${record.ReportID}`">
                <a-button size="small">详情</a-button>
              </router-link>
              <a-button size="small" type="primary" @click="parse(record.ReportID)">解析</a-button>
            </a-space>
          </template>
        </template>
      </a-table>
    </a-card>

    <a-modal v-model:open="uploadOpen" title="上传报告" :confirm-loading="uploading" @ok="submitUpload">
      <a-form layout="vertical">
        <a-form-item label="文件">
          <a-upload :before-upload="beforeUpload" :max-count="1">
            <a-button>
              <template #icon><UploadOutlined /></template>
              选择文件
            </a-button>
          </a-upload>
          <div v-if="uploadFile" class="muted" style="margin-top: 8px">{{ uploadFile.name }}</div>
        </a-form-item>
        <a-form-item label="报告名称（可选）">
          <a-input v-model:value="uploadName" placeholder="默认使用文件名" />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { message } from 'ant-design-vue'
import { UploadOutlined } from '@ant-design/icons-vue'
import { computed, onMounted, reactive, ref } from 'vue'

import { reportsApi } from '@/api/reports'
import PageHeader from '@/components/PageHeader.vue'
import StatusTag from '@/components/StatusTag.vue'
import type { ReportSummary } from '@/types'

const columns = [
  { title: '名称', dataIndex: 'Name', key: 'Name' },
  { title: '状态', dataIndex: 'Status', key: 'Status', width: 120 },
  { title: '章节数', dataIndex: 'SectionCount', key: 'SectionCount', width: 100 },
  { title: '操作', key: 'actions', width: 180 },
]

const rows = ref<ReportSummary[]>([])
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
    const res = await reportsApi.list({ offset: query.offset, limit: query.limit })
    rows.value = res.Reports ?? []
    total.value = res.Total ?? 0
  } finally {
    loading.value = false
  }
}

function onTableChange(pagination: { current?: number; pageSize?: number }) {
  const pageSize = pagination.pageSize ?? query.limit
  const current = pagination.current ?? 1
  query.limit = pageSize
  query.offset = (current - 1) * pageSize
  void load()
}

async function parse(id: string) {
  await reportsApi.parse(id)
  message.success('解析完成')
  await load()
}

// 上传相关
const uploadOpen = ref(false)
const uploadName = ref('')
const uploadFile = ref<File | null>(null)
const uploading = ref(false)

function openUpload() {
  uploadOpen.value = true
}

function beforeUpload(file: unknown) {
  uploadFile.value = file as File
  return false
}

async function submitUpload() {
  if (!uploadFile.value) {
    message.warning('请先选择文件')
    return
  }
  uploading.value = true
  try {
    await reportsApi.upload(uploadFile.value, uploadName.value || undefined)
    message.success('上传成功')
    uploadOpen.value = false
    uploadFile.value = null
    uploadName.value = ''
    await load()
  } finally {
    uploading.value = false
  }
}

onMounted(load)
</script>
