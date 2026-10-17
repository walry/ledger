<template>
  <div>
    <PageHeader title="问答助手" subtitle="基于 RAG + ReAct 的流式问答">
      <a-button @click="createOpen = true">新建会话</a-button>
      <a-button @click="loadConversations">刷新会话</a-button>
    </PageHeader>

    <a-card>
      <a-space style="margin-bottom: 12px" wrap>
        <span class="muted">会话：</span>
        <a-select
          v-model:value="conversationId"
          style="min-width: 320px"
          placeholder="选择会话"
          :options="conversationOptions"
        />
        <a-tag v-if="conversationId" color="blue">{{ conversationId }}</a-tag>
      </a-space>

      <div ref="chatBox" class="chat-list">
        <a-empty v-if="!messages.length" description="输入问题开始对话" />
        <div v-for="(msg, i) in messages" :key="i" class="chat-bubble" :class="msg.role">
          <div class="role muted">{{ roleLabel(msg.role) }}</div>
          <div v-if="msg.content">{{ msg.content }}</div>
        </div>
      </div>

      <a-space direction="vertical" style="width: 100%; margin-top: 12px">
        <a-textarea
          v-model:value="question"
          :rows="3"
          placeholder="请输入问题（Ctrl/⌘ + Enter 发送）"
          @keydown="onKeydown"
        />
        <a-space>
          <a-button type="primary" :loading="streaming" @click="ask">发送（流式）</a-button>
          <a-button :disabled="!streaming" @click="stop">停止</a-button>
        </a-space>
      </a-space>
    </a-card>

    <a-modal v-model:open="createOpen" title="新建会话" :confirm-loading="creating" @ok="createConversation">
      <a-form layout="vertical">
        <a-form-item label="智能体" required>
          <a-select
            v-model:value="newConversation.agentId"
            :options="agentOptions"
            :loading="agentStore.loading"
            placeholder="选择智能体"
          />
        </a-form-item>
        <a-form-item label="标题">
          <a-input v-model:value="newConversation.title" placeholder="可选" />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { message } from 'ant-design-vue'
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref } from 'vue'

import { conversationsApi } from '@/api/conversations'
import { askStream, type StreamController } from '@/api/qa'
import PageHeader from '@/components/PageHeader.vue'
import { useAgentsStore } from '@/stores/agents'
import type { ConversationSummary } from '@/types'
import { truncate } from '@/utils/format'

interface ChatMsg {
  role: 'user' | 'assistant' | 'system'
  content: string
}

const agentStore = useAgentsStore()

const messages = ref<ChatMsg[]>([])
const question = ref('')
const streaming = ref(false)
const chatBox = ref<HTMLElement | null>(null)
let controller: StreamController | null = null

const conversations = ref<ConversationSummary[]>([])
const conversationId = ref<string | undefined>(undefined)

const conversationOptions = computed(() =>
  conversations.value.map((c) => ({ label: `${c.Title}（${c.ConversationID.slice(0, 8)}）`, value: c.ConversationID })),
)

const agentOptions = computed(() =>
  agentStore.items.map((a) => ({ label: `${a.Name}（v${a.Version}）`, value: a.AgentID })),
)

async function loadConversations() {
  const res = await conversationsApi.list({ limit: 100 })
  conversations.value = res.Conversations ?? []
  if (!conversationId.value && conversations.value.length) {
    conversationId.value = conversations.value[0].ConversationID
  }
}

function roleLabel(role: string): string {
  if (role === 'user') return '用户'
  if (role === 'assistant') return '助手'
  return '系统'
}

async function scrollToBottom() {
  await nextTick()
  if (chatBox.value) chatBox.value.scrollTop = chatBox.value.scrollHeight
}

function ask() {
  if (!conversationId.value) {
    message.warning('请先选择或新建会话')
    return
  }
  const q = question.value.trim()
  if (!q || streaming.value) return

  messages.value.push({ role: 'user', content: q })
  const index = messages.value.push({ role: 'assistant', content: '' }) - 1
  question.value = ''
  streaming.value = true
  void scrollToBottom()

  const patch = (delta: string) => {
    if (delta) messages.value[index].content += delta
    void scrollToBottom()
  }

  controller = askStream(
    { conversationId: conversationId.value, question: q },
    {
      onDelta: patch,
      onToolCall: (e) => {
        messages.value.push({ role: 'system', content: `调用工具 ${e.name}…` })
        void scrollToBottom()
      },
      onToolResult: (e) => {
        messages.value.push({ role: 'system', content: `工具 ${e.name} 返回：${truncate(e.result, 200)}` })
        void scrollToBottom()
      },
      onError: (msg) => {
        message.error(msg)
        streaming.value = false
      },
      onDone: async (e) => {
        streaming.value = false
        if (!messages.value[index].content && e.answer) {
          messages.value[index].content = e.answer
        }
        // 兜底：流式未拿到答案时，从会话记录里取最后一条助手消息。
        if (!messages.value[index].content) {
          await recoverAnswer(index)
        }
        void scrollToBottom()
      },
    },
  )
}

async function recoverAnswer(index: number) {
  if (!conversationId.value) return
  try {
    const detail = await conversationsApi.get(conversationId.value)
    const turns = detail.Turns ?? []
    for (let t = turns.length - 1; t >= 0; t--) {
      const msgs = turns[t].Messages ?? []
      for (let m = msgs.length - 1; m >= 0; m--) {
        if (msgs[m].Role === 'assistant' && msgs[m].Content) {
          messages.value[index].content = msgs[m].Content
          return
        }
      }
    }
  } catch {
    // 忽略：答案缺失时保持为空即可
  }
}

function stop() {
  controller?.abort()
  controller = null
  streaming.value = false
  message.info('已停止')
}

function onKeydown(e: KeyboardEvent) {
  if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
    e.preventDefault()
    ask()
  }
}

// 新建会话
const createOpen = ref(false)
const creating = ref(false)
const newConversation = reactive({ agentId: undefined as string | undefined, title: '' })

async function createConversation() {
  if (!newConversation.agentId) {
    message.warning('请选择智能体')
    return
  }
  creating.value = true
  try {
    const res = await conversationsApi.create({
      agentId: newConversation.agentId,
      title: newConversation.title,
    })
    message.success('已创建会话')
    createOpen.value = false
    newConversation.agentId = undefined
    newConversation.title = ''
    await loadConversations()
    conversationId.value = res.conversationId
  } finally {
    creating.value = false
  }
}

onMounted(async () => {
  await Promise.all([loadConversations(), agentStore.fetchAll()])
})

onBeforeUnmount(() => {
  controller?.abort()
})
</script>

<style scoped>
.role {
  font-size: 12px;
  margin-bottom: 4px;
}
</style>
