import { useCallback, useRef } from 'react'
import { useStore } from '../store'
import type { Message } from '../types'
import { sessionApi } from '../api/sessions'

export function useChat() {
  const store = useStore()
  const abortRef = useRef<AbortController | null>(null)

  const sendMessage = useCallback(
    async (content: string) => {
      const sessionId = store.activeSessionId
      if (!sessionId || !content.trim()) return

      const userMsg: Message = {
        id: Date.now().toString(),
        role: 'user',
        content: content.trim(),
        timestamp: Date.now(),
      }
      store.addMessage(sessionId, userMsg)

      const aiMsgId = (Date.now() + 1).toString()
      const aiMsg: Message = {
        id: aiMsgId,
        role: 'assistant',
        content: '',
        timestamp: Date.now(),
        isStreaming: true,
      }
      store.addMessage(sessionId, aiMsg)
      store.setStreaming(true)

      try {
        const controller = new AbortController()
        abortRef.current = controller

        const res = await sessionApi.chatStream(sessionId, content.trim())

        if (!res.ok) {
          const err = await res.json().catch(() => ({ error: 'Unknown' }))
          throw new Error(err.error || `HTTP ${res.status}`)
        }

        const reader = res.body?.getReader()
        if (!reader) throw new Error('No stream body')

        const decoder = new TextDecoder()
        let accumulated = ''

        while (true) {
          const { done, value } = await reader.read()
          if (done) break

          const text = decoder.decode(value, { stream: true })
          const lines = text.split('\n')

          for (const line of lines) {
            if (!line.startsWith('data: ')) continue
            const data = line.slice(6).trim()
            if (data === '[DONE]') break
            if (data.startsWith('[ERROR]')) continue
            accumulated += data
            store.updateLastMessage(sessionId, accumulated)
          }
        }
      } catch (err: unknown) {
        if (err instanceof Error && err.name === 'AbortError') return
        store.updateLastMessage(sessionId, `[错误] ${err instanceof Error ? err.message : '未知错误'}`)
      } finally {
        store.setStreaming(false)
        abortRef.current = null
      }
    },
    [store]
  )

  const stopStreaming = useCallback(() => {
    abortRef.current?.abort()
    store.setStreaming(false)
  }, [store])

  return { sendMessage, stopStreaming, isStreaming: store.isStreaming }
}
