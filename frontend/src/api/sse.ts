import { baseURL } from './client'
import type { AskPayload, TokenUsage } from '@/types'

export interface StreamToolCall {
  id: string
  name: string
}

export interface StreamToolResult extends StreamToolCall {
  result: string
}

export interface StreamDone {
  answer: string
  usage?: TokenUsage
}

export interface StreamHandlers {
  onDelta?: (delta: string) => void
  onToolCall?: (e: StreamToolCall) => void
  onToolResult?: (e: StreamToolResult) => void
  onError?: (msg: string) => void
  onDone?: (e: StreamDone) => void
}

export interface StreamController {
  abort: () => void
}

/**
 * 订阅问答流式接口。
 *
 * 后端 SSE 仅支持 GET（EventSource 无法自定义中止/自定义事件，故用 fetch 手工解析）。
 */
export function askStream(payload: AskPayload, handlers: StreamHandlers): StreamController {
  const controller = new AbortController()
  const params = new URLSearchParams({
    conversationId: payload.conversationId,
    question: payload.question,
  })
  const url = `${baseURL}/qa/ask/stream?${params.toString()}`

  void run(url, controller.signal, handlers)
  return { abort: () => controller.abort() }
}

async function run(url: string, signal: AbortSignal, handlers: StreamHandlers): Promise<void> {
  let response: Response
  try {
    response = await fetch(url, {
      method: 'GET',
      headers: { Accept: 'text/event-stream' },
      signal,
    })
  } catch (e) {
    if (!isAbort(e)) handlers.onError?.(errorText(e))
    return
  }

  if (!response.ok || !response.body) {
    handlers.onError?.(`流式请求失败：HTTP ${response.status}`)
    return
  }

  const reader = response.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''

  try {
    for (;;) {
      const { done, value } = await reader.read()
      if (done) break
      buffer += decoder.decode(value, { stream: true })

      let sep = buffer.indexOf('\n\n')
      while (sep >= 0) {
        const frame = buffer.slice(0, sep)
        buffer = buffer.slice(sep + 2)
        if (dispatch(frame, handlers)) {
          await reader.cancel()
          return
        }
        sep = buffer.indexOf('\n\n')
      }
    }
  } catch (e) {
    if (!isAbort(e)) handlers.onError?.(errorText(e))
  }
}

/** 解析一帧 SSE，返回 true 表示流应结束。 */
function dispatch(frame: string, handlers: StreamHandlers): boolean {
  let event = 'message'
  const dataLines: string[] = []
  for (const line of frame.split('\n')) {
    if (line.startsWith(':')) continue // 心跳注释
    if (line.startsWith('event:')) {
      event = line.slice(6).trim()
    } else if (line.startsWith('data:')) {
      dataLines.push(line.slice(5).replace(/^ /, ''))
    }
  }
  const data = dataLines.join('\n')
  if (data === '') return false

  switch (event) {
    case 'message': {
      const payload = safeParse<{ delta?: string }>(data)
      if (payload) handlers.onDelta?.(payload.delta ?? '')
      return false
    }
    case 'tool_call': {
      const payload = safeParse<StreamToolCall>(data)
      if (payload) handlers.onToolCall?.({ id: payload.id, name: payload.name })
      return false
    }
    case 'tool_result': {
      const payload = safeParse<StreamToolResult>(data)
      if (payload) handlers.onToolResult?.({ id: payload.id, name: payload.name, result: payload.result })
      return false
    }
    case 'error': {
      const payload = safeParse<{ error?: string }>(data)
      handlers.onError?.(payload?.error ?? '未知错误')
      return false
    }
    case 'done': {
      if (data === '[DONE]') {
        handlers.onDone?.({ answer: '' })
        return true
      }
      const payload = safeParse<StreamDone>(data)
      handlers.onDone?.({ answer: payload?.answer ?? '', usage: payload?.usage })
      return true
    }
    default:
      return false
  }
}

function safeParse<T>(raw: string): T | null {
  try {
    return JSON.parse(raw) as T
  } catch {
    return null
  }
}

function isAbort(e: unknown): boolean {
  return e instanceof DOMException && e.name === 'AbortError'
}

function errorText(e: unknown): string {
  return e instanceof Error ? e.message : String(e)
}
