import { http } from './client'
import type {
  CompressContextResult,
  ConversationDetail,
  CreateConversationPayload,
  CreateConversationResult,
  ListConversationsResult,
  PageQuery,
  StartTurnResult,
} from '@/types'

export const conversationsApi = {
  list: (params: PageQuery & { agentId?: string } = {}) =>
    http.get<ListConversationsResult>('/conversations', params),
  get: (id: string) => http.get<ConversationDetail>(`/conversations/${id}`),
  create: (payload: CreateConversationPayload) =>
    http.post<CreateConversationResult>('/conversations', payload),
  startTurn: (id: string, message: string) =>
    http.post<StartTurnResult>(`/conversations/${id}/turns`, { message }),
  compress: (id: string) => http.post<CompressContextResult>(`/conversations/${id}/compress`),
}
