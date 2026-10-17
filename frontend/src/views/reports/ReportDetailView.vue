<template>
  <div>
    <PageHeader :title="detail?.Name || '报告详情'" :subtitle="detail?.ReportID">
      <a-button @click="goBack">返回</a-button>
      <a-button type="primary" :disabled="!detail" @click="parse">解析</a-button>
    </PageHeader>

    <a-spin :spinning="loading">
      <a-descriptions v-if="detail" bordered :column="2">
        <a-descriptions-item label="状态">
          <StatusTag :status="detail.Status" />
        </a-descriptions-item>
        <a-descriptions-item label="类型">{{ detail.ContentType || '-' }}</a-descriptions-item>
        <a-descriptions-item label="文件名">{{ detail.Filename || '-' }}</a-descriptions-item>
        <a-descriptions-item label="章节数">{{ detail.SectionCount }}</a-descriptions-item>
        <a-descriptions-item label="存储路径" :span="2">
          <span class="muted">{{ detail.StoragePath || '-' }}</span>
        </a-descriptions-item>
        <a-descriptions-item label="失败原因" :span="2">
          {{ detail.FailureReason || '-' }}
        </a-descriptions-item>
      </a-descriptions>

      <a-card title="章节" style="margin-top: 16px">
        <a-table
          :columns="columns"
          :data-source="detail?.Sections ?? []"
          row-key="ID"
          :pagination="{ pageSize: 20 }"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'Content'">
              <div class="pre-wrap">{{ record.Content }}</div>
            </template>
          </template>
        </a-table>
      </a-card>
    </a-spin>
  </div>
</template>

<script setup lang="ts">
import { message } from 'ant-design-vue'
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { reportsApi } from '@/api/reports'
import PageHeader from '@/components/PageHeader.vue'
import StatusTag from '@/components/StatusTag.vue'
import type { ReportDetail } from '@/types'

const route = useRoute()
const router = useRouter()
const id = route.params.id as string

const columns = [
  { title: '序号', dataIndex: 'Order', key: 'Order', width: 80 },
  { title: '类型', dataIndex: 'Type', key: 'Type', width: 120 },
  { title: '标题', dataIndex: 'Title', key: 'Title', width: 240 },
  { title: '内容', dataIndex: 'Content', key: 'Content' },
]

const detail = ref<ReportDetail | null>(null)
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    detail.value = await reportsApi.get(id)
  } finally {
    loading.value = false
  }
}

async function parse() {
  await reportsApi.parse(id)
  message.success('解析完成')
  await load()
}

function goBack() {
  router.push('/reports')
}

onMounted(load)
</script>
