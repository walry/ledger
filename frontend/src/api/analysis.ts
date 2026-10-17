import { http } from './client'
import type {
  AnalysisResult,
  ListTasksResult,
  PageQuery,
  RunTaskResult,
  SubmitTaskPayload,
  SubmitTaskResult,
} from '@/types'

export const analysisApi = {
  list: (params: PageQuery = {}) => http.get<ListTasksResult>('/analysis/tasks', params),
  submit: (payload: SubmitTaskPayload) => http.post<SubmitTaskResult>('/analysis/tasks', payload),
  run: (id: string) => http.post<RunTaskResult>(`/analysis/tasks/${id}/run`),
  get: (id: string) => http.get<AnalysisResult>(`/analysis/tasks/${id}`),
}
