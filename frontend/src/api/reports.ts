import { http } from './client'
import type { ListReportsResult, PageQuery, ParseReportResult, ReportDetail, UploadReportResult } from '@/types'

export const reportsApi = {
  list: (params: PageQuery & { status?: string } = {}) =>
    http.get<ListReportsResult>('/reports', params),
  get: (id: string) => http.get<ReportDetail>(`/reports/${id}`),
  upload: (file: File, name?: string) => {
    const form = new FormData()
    form.append('file', file)
    if (name) form.append('name', name)
    return http.postForm<UploadReportResult>('/reports', form)
  },
  parse: (id: string) => http.post<ParseReportResult>(`/reports/${id}/parse`),
}
