import { http } from './client'
import type {
  AddSkillPayload,
  AgentActionRef,
  AgentDetail,
  AgentVersionChangeResult,
  AgentVersionResult,
  BindToolPayload,
  CreateAgentPayload,
  ListAgentsResult,
  PageQuery,
} from '@/types'

export const agentsApi = {
  list: (params: PageQuery = {}) => http.get<ListAgentsResult>('/agents', params),
  get: (id: string) => http.get<AgentDetail>(`/agents/${id}`),
  create: (payload: CreateAgentPayload) => http.post<AgentVersionResult>('/agents', payload),
  addSkill: (id: string, payload: AddSkillPayload) =>
    http.post<AgentActionRef>(`/agents/${id}/skills`, payload),
  bindTool: (id: string, payload: BindToolPayload) =>
    http.post<AgentActionRef>(`/agents/${id}/tools`, payload),
  publish: (id: string) => http.post<AgentVersionResult>(`/agents/${id}/publish`),
  newVersion: (id: string) => http.post<AgentVersionChangeResult>(`/agents/${id}/versions`),
  rollback: (id: string) => http.post<AgentVersionChangeResult>(`/agents/${id}/rollback`),
}
