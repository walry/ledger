import { http } from './client'
import type {
  AddDocumentPayload,
  CreateKnowledgeBasePayload,
  CreateKnowledgeBaseResult,
  DocumentIndexResult,
  KnowledgeBaseDetail,
  ListKnowledgeBasesResult,
  PageQuery,
  SearchPayload,
  SearchResult,
} from '@/types'

export const knowledgeApi = {
  list: (params: PageQuery = {}) => http.get<ListKnowledgeBasesResult>('/knowledge-bases', params),
  get: (id: string) => http.get<KnowledgeBaseDetail>(`/knowledge-bases/${id}`),
  create: (payload: CreateKnowledgeBasePayload) =>
    http.post<CreateKnowledgeBaseResult>('/knowledge-bases', payload),
  addDocument: (kbID: string, payload: AddDocumentPayload) =>
    http.post<DocumentIndexResult>(`/knowledge-bases/${kbID}/documents`, payload),
  reindex: (documentID: string, text?: string) =>
    http.post<DocumentIndexResult>(`/documents/${documentID}/reindex`, { text }),
  search: (payload: SearchPayload) => http.post<SearchResult>('/knowledge/search', payload),
}
