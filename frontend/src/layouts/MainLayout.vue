<template>
  <a-layout style="min-height: 100vh">
    <a-layout-sider v-model:collapsed="collapsed" collapsible :trigger="null">
      <div class="brand">
        <span v-if="!collapsed">Ledger</span>
        <span v-else>L</span>
      </div>
      <a-menu
        v-model:selectedKeys="selectedKeys"
        theme="dark"
        mode="inline"
        :items="menuItems"
        @click="onMenuClick"
      />
    </a-layout-sider>

    <a-layout>
      <a-layout-header class="header">
        <span class="trigger" @click="collapsed = !collapsed">☰</span>
        <span class="title">{{ currentTitle }}</span>
      </a-layout-header>

      <a-layout-content>
        <div class="page">
          <router-view />
        </div>
      </a-layout-content>

      <a-layout-footer style="text-align: center" class="muted">
        智能财务报告分析 · DDD 智能体应用
      </a-layout-footer>
    </a-layout>
  </a-layout>
</template>

<script setup lang="ts">
import { computed, h, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  ApiOutlined,
  CommentOutlined,
  DashboardOutlined,
  DatabaseOutlined,
  FileTextOutlined,
  MessageOutlined,
  RobotOutlined,
  SearchOutlined,
  ThunderboltOutlined,
  ToolOutlined,
} from '@ant-design/icons-vue'
const route = useRoute()
const router = useRouter()
const collapsed = ref(false)

const menuItems = [
  { key: '/', icon: () => h(DashboardOutlined), label: '概览' },
  { key: '/reports', icon: () => h(FileTextOutlined), label: '报告管理' },
  { key: '/knowledge', icon: () => h(DatabaseOutlined), label: '知识库' },
  { key: '/knowledge/search', icon: () => h(SearchOutlined), label: '知识检索' },
  { key: '/agents', icon: () => h(RobotOutlined), label: '智能体' },
  { key: '/conversations', icon: () => h(CommentOutlined), label: '会话' },
  { key: '/qa', icon: () => h(MessageOutlined), label: '问答助手' },
  { key: '/analysis', icon: () => h(ThunderboltOutlined), label: '分析任务' },
  { key: '/tools', icon: () => h(ToolOutlined), label: '工具' },
  { key: '/model-configs', icon: () => h(ApiOutlined), label: '模型配置' },
]

const selectedKeys = ref<string[]>([route.path])
watch(
  () => route.path,
  (path) => {
    selectedKeys.value = [path]
  },
)

const currentTitle = computed(() => (route.meta.title as string) ?? '')

function onMenuClick({ key }: { key: string }) {
  if (key !== route.path) router.push(key)
}
</script>

<style scoped>
.brand {
  height: 48px;
  margin: 8px 16px;
  color: #fff;
  font-size: 18px;
  font-weight: 600;
  line-height: 48px;
  text-align: center;
}

.header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 0 16px;
  background: #fff;
}

.trigger {
  cursor: pointer;
  font-size: 18px;
}

.title {
  font-size: 16px;
  font-weight: 500;
}
</style>
