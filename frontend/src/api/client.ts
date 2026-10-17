import axios, { AxiosError, type AxiosRequestConfig, type InternalAxiosRequestConfig } from 'axios'
import { message } from 'ant-design-vue'

const baseURL = import.meta.env.VITE_API_BASE || '/api'

const client = axios.create({ baseURL, timeout: 120000 })

client.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const isForm = config.data instanceof FormData
  if (config.data != null && !isForm && !config.headers['Content-Type']) {
    config.headers['Content-Type'] = 'application/json'
  }
  return config
})

// 成功时拆掉 {"data": ...} 外层封装；失败时把 {"error": ...} 转成异常并提示。
client.interceptors.response.use(
  (response) => {
    const body = response.data
    if (body && typeof body === 'object' && 'data' in body) {
      return body.data
    }
    return body
  },
  (error: AxiosError<{ error?: string }>) => {
    const text = error.response?.data?.error || error.message || '请求失败'
    message.error(text)
    return Promise.reject(new Error(text))
  },
)

/** 发起请求：拦截器已拆包，返回值即业务数据。 */
export async function request<T>(config: AxiosRequestConfig): Promise<T> {
  const res = await client.request<T>(config)
  return res as unknown as T
}

/** 简化的 HTTP 方法封装，泛型即业务数据类型。 */
export const http = {
  get: <T>(url: string, params?: object) => request<T>({ url, method: 'GET', params }),
  post: <T>(url: string, data?: unknown) => request<T>({ url, method: 'POST', data }),
  put: <T>(url: string, data?: unknown) => request<T>({ url, method: 'PUT', data }),
  postForm: <T>(url: string, form: FormData) => request<T>({ url, method: 'POST', data: form }),
}

export { baseURL }
