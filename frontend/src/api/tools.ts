import { http } from './client'
import type {
  EnableToolResult,
  ListToolsResult,
  PageQuery,
  RegisterToolPayload,
  RegisterToolResult,
  ToolDetail,
} from '@/types'

export const toolsApi = {
  list: (params: PageQuery = {}) => http.get<ListToolsResult>('/tools', params),
  get: (id: string) => http.get<ToolDetail>(`/tools/${id}`),
  register: (payload: RegisterToolPayload) => http.post<RegisterToolResult>('/tools', payload),
  enable: (id: string, enabled: boolean) =>
    http.post<EnableToolResult>(`/tools/${id}/enable`, { enabled }),
}
