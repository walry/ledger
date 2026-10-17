import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

import MainLayout from '@/layouts/MainLayout.vue'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    component: MainLayout,
    children: [
      {
        path: '',
        name: 'dashboard',
        component: () => import('@/views/DashboardView.vue'),
        meta: { title: '概览' },
      },
      {
        path: 'reports',
        name: 'reports',
        component: () => import('@/views/reports/ReportListView.vue'),
        meta: { title: '报告管理' },
      },
      {
        path: 'reports/:id',
        name: 'report-detail',
        component: () => import('@/views/reports/ReportDetailView.vue'),
        meta: { title: '报告详情' },
      },
      {
        path: 'knowledge',
        name: 'knowledge',
        component: () => import('@/views/knowledge/KnowledgeBaseView.vue'),
        meta: { title: '知识库' },
      },
      {
        path: 'knowledge/search',
        name: 'knowledge-search',
        component: () => import('@/views/knowledge/KnowledgeSearchView.vue'),
        meta: { title: '知识检索' },
      },
      {
        path: 'agents',
        name: 'agents',
        component: () => import('@/views/agents/AgentListView.vue'),
        meta: { title: '智能体' },
      },
      {
        path: 'agents/:id',
        name: 'agent-detail',
        component: () => import('@/views/agents/AgentDetailView.vue'),
        meta: { title: '智能体详情' },
      },
      {
        path: 'conversations',
        name: 'conversations',
        component: () => import('@/views/conversations/ConversationListView.vue'),
        meta: { title: '会话' },
      },
      {
        path: 'conversations/:id',
        name: 'conversation-detail',
        component: () => import('@/views/conversations/ConversationDetailView.vue'),
        meta: { title: '会话详情' },
      },
      {
        path: 'qa',
        name: 'qa',
        component: () => import('@/views/qa/QAChatView.vue'),
        meta: { title: '问答助手' },
      },
      {
        path: 'analysis',
        name: 'analysis',
        component: () => import('@/views/analysis/AnalysisTaskView.vue'),
        meta: { title: '分析任务' },
      },
      {
        path: 'tools',
        name: 'tools',
        component: () => import('@/views/tools/ToolListView.vue'),
        meta: { title: '工具' },
      },
      {
        path: 'model-configs',
        name: 'model-configs',
        component: () => import('@/views/modelconfigs/ModelConfigListView.vue'),
        meta: { title: '模型配置' },
      },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

export default router
