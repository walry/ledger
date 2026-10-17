<template>
  <div>
    <PageHeader title="知识库" subtitle="文档切片、向量化与检索的基础">
      <a-button type="primary" @click="createOpen = true">创建知识库</a-button>
    </PageHeader>

    <a-card>
      <a-table
        :columns="columns"
        :data-source="rows"
        :loading="loading"
        :pagination="pagination"
        row-key="KnowledgeBaseID"
        :row-class-name="rowClass"
        @change="onTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'actions'">
            <a-button size="small" @click="select(record.KnowledgeBaseID)">查看文档</a-button>
          </template>
        </template>
      </a-table>
    </a-card>

    <a-card v-if="detail" style="margin-top: 16px">
      <template #title>
        {{ detail.Name }} 的文档
        <a-tag style="margin-left: 8px">{{ detail.Documents.length }} 篇</a-tag>
      </template>
      <template #extra>
        <a-button type="primary" size="small" @click="openAddDoc">新增文档</a-button>
      </template>
      <a-empty v-if="!detail.Documents.length" description="暂无文档" />
      <a-table
        v-else
        :columns="docColumns"
        :data-source="detail.Documents"
        row-key="DocumentID"
        size="small"
        :pagination="false"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'Status'">
            <StatusTag :status="record.Status" />
          </template>
          <template v-else-if="column.key === 'actions'">
            <a-button size="small" @click="openReindex(record.DocumentID)">重建索引</a-button>
          </template>
        </template>
      </a-table>
    </a-card>

    <a-modal v-model:open="createOpen" title="创建知识库" :confirm-loading="submitting" @ok="createKB">
      <a-form layout="vertical">
        <a-form-item label="名称" required>
          <a-input v-model:value="kbForm.name" />
        </a-form-item>
        <a-form-item label="描述">
          <a-input v-model:value="kbForm.description" />
        </a-form-item>
        <a-form-item label="Embedding 模型">
          <a-input v-model:value="kbForm.embeddingModel" placeholder="留空使用服务端默认" />
        </a-form-item>
        <a-form-item label="向量维度">
          <a-input-number v-model:value="kbForm.dimension" :min="1" style="width: 100%" />
        </a-form-item>
      </a-form>
    </a-modal>

    <a-modal v-model:open="docOpen" title="新增文档" :confirm-loading="submitting" @ok="addDoc">
      <a-form layout="vertical">
        <a-form-item label="标题">
          <a-input v-model:value="docForm.title" />
        </a-form-item>
        <a-form-item label="来源类型">
          <a-select v-model:value="docForm.sourceKind">
            <a-select-option value="file">文件</a-select-option>
            <a-select-option value="url">链接</a-select-option>
            <a-select-option value="report">报告</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="来源引用 ID">
          <a-input v-model:value="docForm.sourceRefId" placeholder="如文件名或 URL" />
        </a-form-item>
        <a-form-item label="正文" required>
          <a-textarea v-model:value="docForm.text" :rows="6" />
        </a-form-item>
      </a-form>
    </a-modal>

    <a-modal v-model:open="reindexOpen" title="重建索引" :confirm-loading="submitting" @ok="reindex">
      <a-form layout="vertical">
        <a-form-item label="新的正文（留空则沿用现有切片）">
          <a-textarea v-model:value="reindexText" :rows="6" />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { message } from 'ant-design-vue'
import { computed, onMounted, reactive, ref } from 'vue'

import { knowledgeApi } from '@/api/knowledge'
import PageHeader from '@/components/PageHeader.vue'
import StatusTag from '@/components/StatusTag.vue'
import type { KnowledgeBaseDetail, KnowledgeBaseSummary } from '@/types'

const columns = [
  { title: '名称', dataIndex: 'Name', key: 'Name' },
  { title: '描述', dataIndex: 'Description', key: 'Description' },
  { title: 'Embedding', dataIndex: 'EmbeddingModel', key: 'EmbeddingModel' },
  { title: '维度', dataIndex: 'Dimension', key: 'Dimension', width: 90 },
  { title: '文档数', dataIndex: 'DocumentCount', key: 'DocumentCount', width: 90 },
  { title: '操作', key: 'actions', width: 120 },
]

const docColumns = [
  { title: '标题', dataIndex: 'Title', key: 'Title' },
  { title: '来源', dataIndex: 'SourceKind', key: 'SourceKind', width: 100 },
  { title: '状态', dataIndex: 'Status', key: 'Status', width: 100 },
  { title: '切片数', dataIndex: 'ChunkCount', key: 'ChunkCount', width: 90 },
  { title: '操作', key: 'actions', width: 120 },
]

const rows = ref<KnowledgeBaseSummary[]>([])
const loading = ref(false)
const total = ref(0)
const query = reactive({ offset: 0, limit: 10 })
const selectedId = ref('')

const pagination = computed(() => ({
  current: Math.floor(query.offset / query.limit) + 1,
  pageSize: query.limit,
  total: total.value,
  showSizeChanger: true,
}))

const detail = ref<KnowledgeBaseDetail | null>(null)

function rowClass(record: KnowledgeBaseSummary) {
  return record.KnowledgeBaseID === selectedId.value ? 'ant-table-row-selected' : ''
}

async function load() {
  loading.value = true
  try {
    const res = await knowledgeApi.list({ offset: query.offset, limit: query.limit })
    rows.value = res.KnowledgeBases ?? []
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

async function select(id: string) {
  selectedId.value = id
  detail.value = await knowledgeApi.get(id)
}

// 创建知识库
const createOpen = ref(false)
const submitting = ref(false)
const kbForm = reactive({ name: '', description: '', embeddingModel: '', dimension: 1536 })

async function createKB() {
  submitting.value = true
  try {
    await knowledgeApi.create({
      name: kbForm.name,
      description: kbForm.description,
      embeddingModel: kbForm.embeddingModel || undefined,
      dimension: kbForm.dimension,
    })
    message.success('已创建知识库')
    createOpen.value = false
    kbForm.name = ''
    kbForm.description = ''
    await load()
  } finally {
    submitting.value = false
  }
}

// 新增文档
const docOpen = ref(false)
const docForm = reactive({ title: '', sourceKind: 'file', sourceRefId: '', text: '' })

function openAddDoc() {
  docOpen.value = true
}

async function addDoc() {
  if (!detail.value) return
  submitting.value = true
  try {
    await knowledgeApi.addDocument(detail.value.KnowledgeBaseID, {
      title: docForm.title,
      sourceKind: docForm.sourceKind,
      sourceRefId: docForm.sourceRefId,
      text: docForm.text,
    })
    message.success('已新增文档并建立索引')
    docOpen.value = false
    docForm.title = ''
    docForm.sourceRefId = ''
    docForm.text = ''
    await select(detail.value.KnowledgeBaseID)
  } finally {
    submitting.value = false
  }
}

// 重建索引
const reindexOpen = ref(false)
const reindexText = ref('')
const reindexDocId = ref('')

function openReindex(docId: string) {
  reindexDocId.value = docId
  reindexText.value = ''
  reindexOpen.value = true
}

async function reindex() {
  submitting.value = true
  try {
    await knowledgeApi.reindex(reindexDocId.value, reindexText.value || undefined)
    message.success('已重建索引')
    reindexOpen.value = false
    if (detail.value) await select(detail.value.KnowledgeBaseID)
  } finally {
    submitting.value = false
  }
}

onMounted(load)
</script>
