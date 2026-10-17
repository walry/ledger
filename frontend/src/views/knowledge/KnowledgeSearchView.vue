<template>
  <div>
    <PageHeader title="知识检索" subtitle="语义检索知识库切片" />

    <a-card>
      <a-form layout="vertical" @submit.prevent>
        <a-form-item label="查询">
          <a-textarea
            v-model:value="query.query"
            :rows="2"
            placeholder="输入自然语言问题"
            @keydown="onKeydown"
          />
        </a-form-item>
        <a-row :gutter="16">
          <a-col :md="16">
            <a-form-item label="知识库（不选则检索全部）">
              <a-select
                v-model:value="query.knowledgeBaseIds"
                mode="multiple"
                :options="kbOptions"
                placeholder="可多选"
              />
            </a-form-item>
          </a-col>
          <a-col :md="8">
            <a-form-item label="TopK">
              <a-input-number v-model:value="query.topK" :min="1" :max="50" style="width: 100%" />
            </a-form-item>
          </a-col>
        </a-row>
        <a-button type="primary" :loading="loading" @click="search">检索</a-button>
      </a-form>
    </a-card>

    <a-card v-if="searched" title="检索结果" style="margin-top: 16px">
      <a-empty v-if="!chunks.length" description="未命中任何切片" />
      <a-list v-else :data-source="chunks" item-layout="vertical">
        <template #renderItem="{ item }">
          <a-list-item>
            <a-list-item-meta>
              <template #title>
                <a-space>
                  <a-tag color="blue">score {{ item.Score.toFixed(4) }}</a-tag>
                  <span class="muted">文档 {{ item.DocumentID }}</span>
                </a-space>
              </template>
            </a-list-item-meta>
            <div class="pre-wrap">{{ item.Content }}</div>
          </a-list-item>
        </template>
      </a-list>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'

import { knowledgeApi } from '@/api/knowledge'
import PageHeader from '@/components/PageHeader.vue'
import type { SearchChunk } from '@/types'

const query = reactive({
  query: '',
  knowledgeBaseIds: [] as string[],
  topK: 5,
})

const chunks = ref<SearchChunk[]>([])
const loading = ref(false)
const searched = ref(false)
const kbOptions = ref<{ label: string; value: string }[]>([])

async function search() {
  if (!query.query.trim()) return
  loading.value = true
  try {
    const res = await knowledgeApi.search({
      query: query.query,
      knowledgeBaseIds: query.knowledgeBaseIds,
      topK: query.topK,
    })
    chunks.value = res.Chunks ?? []
    searched.value = true
  } finally {
    loading.value = false
  }
}

function onKeydown(e: KeyboardEvent) {
  if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
    e.preventDefault()
    void search()
  }
}

onMounted(async () => {
  const res = await knowledgeApi.list({ limit: 200 })
  kbOptions.value = (res.KnowledgeBases ?? []).map((k) => ({ label: k.Name, value: k.KnowledgeBaseID }))
})
</script>
