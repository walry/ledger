import { http } from './client'
import type {
  CreateModelConfigPayload,
  ListModelConfigsResult,
  ModelConfig,
  ModelConfigRef,
  PageQuery,
  UpdateModelConfigPayload,
} from '@/types'

export const modelConfigsApi = {
  list: (params: PageQuery = {}) => http.get<ListModelConfigsResult>('/model-configs', params),
  get: (id: string) => http.get<ModelConfig>(`/model-configs/${id}`),
  create: (payload: CreateModelConfigPayload) =>
    http.post<ModelConfigRef>('/model-configs', payload),
  update: (id: string, payload: UpdateModelConfigPayload) =>
    http.put<ModelConfigRef>(`/model-configs/${id}`, payload),
}
