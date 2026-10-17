<template>
  <div>
    <PageHeader :title="detail?.Name || '智能体详情'" :subtitle="detail?.AgentID">
      <a-button @click="goBack">返回</a-button>
      <a-button @click="newVersion">派生新版本</a-button>
      <a-button @click="rollback">回滚</a-button>
      <a-button type="primary" :disabled="detail?.Status !== 'draft'" @click="publish">发布</a-button>
    </PageHeader>

    <a-spin :spinning="loading">
      <a-descriptions v-if="detail" bordered :column="2">
        <a-descriptions-item label="状态"><StatusTag :status="detail.Status" /></a-descriptions-item>
        <a-descriptions-item label="版本">
          v{{ detail.Version }}（父版本 {{ detail.ParentVersion || '-' }}）
        </a-descriptions-item>
        <a-descriptions-item label="模型">{{ detail.Model }}（{{ detail.Provider }}）</a-descriptions-item>
        <a-descriptions-item label="模型配置">{{ detail.ModelConfigID }}</a-descriptions-item>
        <a-descriptions-item label="最大步数">{{ detail.MaxSteps }}</a-descriptions-item>
        <a-descriptions-item label="检索 TopK">{{ detail.RetrievalTopK }}</a-descriptions-item>
        <a-descriptions-item label="关联知识库" :span="2">
          <a-space v-if="detail.KnowledgeBaseIDs?.length" wrap>
            <a-tag v-for="kb in detail.KnowledgeBaseIDs" :key="kb">{{ kb }}</a-tag>
          </a-space>
          <span v-else class="muted">-</span>
        </a-descriptions-item>
        <a-descriptions-item label="系统提示词" :span="2">
          <div class="pre-wrap">{{ detail.SystemPrompt }}</div>
        </a-descriptions-item>
      </a-descriptions>

      <a-row :gutter="16" style="margin-top: 16px">
        <a-col :md="12">
          <a-card title="技能绑定">
            <template #extra>
              <a-button size="small" @click="skillOpen = true">绑定技能</a-button>
            </template>
            <a-table
              :columns="skillColumns"
              :data-source="detail?.Skills ?? []"
              row-key="SkillID"
              size="small"
              :pagination="false"
            />
          </a-card>
        </a-col>
        <a-col :md="12">
          <a-card title="工具绑定">
            <template #extra>
              <a-button size="small" @click="openBindTool">绑定工具</a-button>
            </template>
            <a-table
              :columns="toolColumns"
              :data-source="detail?.Tools ?? []"
              row-key="ToolID"
              size="small"
              :pagination="false"
            />
          </a-card>
        </a-col>
      </a-row>
    </a-spin>

    <a-modal v-model:open="skillOpen" title="绑定技能" :confirm-loading="submitting" @ok="submitSkill">
      <a-form layout="vertical">
        <a-form-item label="技能 ID" required>
          <a-input v-model:value="skillForm.skillId" />
        </a-form-item>
        <a-form-item label="版本">
          <a-input-number v-model:value="skillForm.version" :min="1" style="width: 100%" />
        </a-form-item>
      </a-form>
    </a-modal>

    <a-modal v-model:open="toolOpen" title="绑定工具" :confirm-loading="submitting" @ok="submitTool">
      <a-form layout="vertical">
        <a-form-item label="工具" required>
          <a-select v-model:value="toolForm.toolId" :options="toolOptions" placeholder="选择工具" />
        </a-form-item>
        <a-form-item label="版本">
          <a-input-number v-model:value="toolForm.version" :min="1" style="width: 100%" />
        </a-form-item>
        <a-form-item label="配置（JSON，可选）">
          <a-textarea v-model:value="toolForm.config" :rows="4" placeholder='{"key":"value"}' />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { message } from 'ant-design-vue'
import { onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { agentsApi } from '@/api/agents'
import { toolsApi } from '@/api/tools'
import PageHeader from '@/components/PageHeader.vue'
import StatusTag from '@/components/StatusTag.vue'
import type { AgentDetail } from '@/types'

const route = useRoute()
const router = useRouter()
const id = route.params.id as string

const skillColumns = [
  { title: '技能 ID', dataIndex: 'SkillID', key: 'SkillID' },
  { title: '版本', dataIndex: 'Version', key: 'Version', width: 80 },
  { title: '启用', dataIndex: 'Enabled', key: 'Enabled', width: 80 },
]
const toolColumns = [
  { title: '工具 ID', dataIndex: 'ToolID', key: 'ToolID' },
  { title: '版本', dataIndex: 'Version', key: 'Version', width: 80 },
  { title: '启用', dataIndex: 'Enabled', key: 'Enabled', width: 80 },
]

const detail = ref<AgentDetail | null>(null)
const loading = ref(false)
const submitting = ref(false)

async function load() {
  loading.value = true
  try {
    detail.value = await agentsApi.get(id)
  } finally {
    loading.value = false
  }
}

async function publish() {
  await agentsApi.publish(id)
  message.success('已发布')
  await load()
}

async function newVersion() {
  const res = await agentsApi.newVersion(id)
  message.success(`已派生 v${res.version}`)
  router.push(`/agents/${res.agentId}`)
}

async function rollback() {
  const res = await agentsApi.rollback(id)
  message.success(`已回滚为 v${res.version}（新草稿）`)
  router.push(`/agents/${res.agentId}`)
}

// 技能
const skillOpen = ref(false)
const skillForm = reactive({ skillId: '', version: 1 })

async function submitSkill() {
  submitting.value = true
  try {
    await agentsApi.addSkill(id, { skillId: skillForm.skillId, version: skillForm.version })
    message.success('已绑定技能')
    skillOpen.value = false
    skillForm.skillId = ''
    await load()
  } finally {
    submitting.value = false
  }
}

// 工具
const toolOpen = ref(false)
const toolForm = reactive({ toolId: undefined as string | undefined, version: 1, config: '' })
const toolOptions = ref<{ label: string; value: string }[]>([])

async function openBindTool() {
  if (toolOptions.value.length === 0) {
    const res = await toolsApi.list({ limit: 200 })
    toolOptions.value = (res.Tools ?? []).map((t) => ({ label: t.Name, value: t.ToolID }))
  }
  toolOpen.value = true
}

async function submitTool() {
  if (!toolForm.toolId) {
    message.warning('请选择工具')
    return
  }
  let config: Record<string, string> | undefined
  if (toolForm.config.trim()) {
    try {
      config = JSON.parse(toolForm.config) as Record<string, string>
    } catch {
      message.error('配置不是合法 JSON')
      return
    }
  }
  submitting.value = true
  try {
    await agentsApi.bindTool(id, {
      toolId: toolForm.toolId,
      version: toolForm.version,
      config,
    })
    message.success('已绑定工具')
    toolOpen.value = false
    toolForm.toolId = undefined
    toolForm.config = ''
    await load()
  } finally {
    submitting.value = false
  }
}

function goBack() {
  router.push('/agents')
}

onMounted(load)
</script>
