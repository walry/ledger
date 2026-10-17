<template>
  <div>
    <PageHeader :title="detail?.Title || '会话详情'" :subtitle="detail?.ConversationID">
      <a-button @click="goBack">返回</a-button>
      <a-button type="primary" :loading="compressing" @click="compress">压缩上下文</a-button>
    </PageHeader>

    <a-spin :spinning="loading">
      <a-descriptions v-if="detail" bordered :column="2">
        <a-descriptions-item label="智能体">{{ detail.AgentID }}</a-descriptions-item>
        <a-descriptions-item label="智能体版本">v{{ detail.AgentVersion }}</a-descriptions-item>
        <a-descriptions-item label="状态"><StatusTag :status="detail.Status" /></a-descriptions-item>
        <a-descriptions-item label="创建时间">{{ formatDateTime(detail.CreatedAt) }}</a-descriptions-item>
        <a-descriptions-item v-if="detail.Summary" label="压缩摘要" :span="2">
          <div class="pre-wrap">{{ detail.Summary }}</div>
        </a-descriptions-item>
      </a-descriptions>

      <a-card style="margin-top: 16px" title="对话记录">
        <a-empty v-if="!detail?.Turns?.length" description="暂无对话" />
        <div v-for="turn in detail?.Turns ?? []" :key="turn.TurnID" class="turn">
          <a-divider orientation="left">
            第 {{ turn.Index + 1 }} 轮 · <StatusTag :status="turn.Status" />
          </a-divider>
          <div v-for="(msg, i) in turn.Messages" :key="i" class="chat-bubble" :class="msg.Role">
            <div class="role muted">
              {{ roleLabel(msg.Role) }}<template v-if="msg.Name"> · {{ msg.Name }}</template>
            </div>
            <div v-if="msg.Content">{{ msg.Content }}</div>
            <div v-for="call in msg.ToolCalls" :key="call.ID" class="muted">
              → 调用 {{ call.Name }}({{ call.Arguments }})
            </div>
          </div>
          <div v-if="turn.Failure" class="muted">失败：{{ turn.Failure }}</div>
        </div>
      </a-card>
    </a-spin>
  </div>
</template>

<script setup lang="ts">
import { message } from 'ant-design-vue'
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { conversationsApi } from '@/api/conversations'
import PageHeader from '@/components/PageHeader.vue'
import StatusTag from '@/components/StatusTag.vue'
import type { ConversationDetail } from '@/types'
import { formatDateTime } from '@/utils/format'

const route = useRoute()
const router = useRouter()
const id = route.params.id as string

const detail = ref<ConversationDetail | null>(null)
const loading = ref(false)
const compressing = ref(false)

async function load() {
  loading.value = true
  try {
    detail.value = await conversationsApi.get(id)
  } finally {
    loading.value = false
  }
}

async function compress() {
  compressing.value = true
  try {
    const res = await conversationsApi.compress(id)
    message.success(res.compressed ? '已压缩上下文' : '无需压缩')
    await load()
  } finally {
    compressing.value = false
  }
}

function roleLabel(role: string): string {
  switch (role) {
    case 'user':
      return '用户'
    case 'assistant':
      return '助手'
    case 'system':
      return '系统'
    case 'tool':
      return '工具'
    default:
      return role
  }
}

function goBack() {
  router.push('/conversations')
}

onMounted(load)
</script>

<style scoped>
.role {
  font-size: 12px;
  margin-bottom: 4px;
}
</style>
