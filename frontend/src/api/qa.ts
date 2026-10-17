import { http } from './client'
import type { AskPayload, AskResult } from '@/types'

export const qaApi = {
  ask: (payload: AskPayload) => http.post<AskResult>('/qa/ask', payload),
}

export { askStream } from './sse'
export type { StreamController, StreamHandlers } from './sse'
