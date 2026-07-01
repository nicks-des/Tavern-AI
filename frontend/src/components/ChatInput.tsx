import { useState, useRef, useEffect, type KeyboardEvent } from 'react'
import { useChat } from '../hooks/useChat'

export function ChatInput() {
  const [input, setInput] = useState('')
  const textareaRef = useRef<HTMLTextAreaElement>(null)
  const { sendMessage, stopStreaming, isStreaming } = useChat()

  const handleSend = () => {
    if (!input.trim() || isStreaming) return
    sendMessage(input)
    setInput('')
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto'
    }
  }

  const handleKeyDown = (e: KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  useEffect(() => {
    textareaRef.current?.focus()
  }, [])

  const handleInput = () => {
    const el = textareaRef.current
    if (!el) return
    el.style.height = 'auto'
    el.style.height = Math.min(el.scrollHeight, 160) + 'px'
  }

  return (
    <div className="border-t border-tavern-700/50 p-4">
      <div className="flex items-end gap-3 max-w-3xl mx-auto">
        <textarea
          ref={textareaRef}
          value={input}
          onChange={(e) => {
            setInput(e.target.value)
            handleInput()
          }}
          onKeyDown={handleKeyDown}
          placeholder={isStreaming ? 'AI 正在回复...' : '输入消息... (Enter 发送, Shift+Enter 换行)'}
          disabled={isStreaming}
          rows={1}
          className="flex-1 bg-tavern-800/60 border border-tavern-600/50 rounded-xl px-4 py-3 text-sm text-gray-100 placeholder-gray-500 resize-none focus:outline-none focus:border-accent/50 focus:ring-1 focus:ring-accent/30 disabled:opacity-50 transition-colors"
        />

        {isStreaming ? (
          <button
            onClick={stopStreaming}
            className="px-4 py-3 bg-rose-500/20 hover:bg-rose-500/30 text-rose-400 rounded-xl text-sm font-medium transition-colors"
          >
            停止
          </button>
        ) : (
          <button
            onClick={handleSend}
            disabled={!input.trim()}
            className="px-4 py-3 bg-accent hover:bg-accent-dark text-white rounded-xl text-sm font-medium disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
          >
            发送
          </button>
        )}
      </div>
    </div>
  )
}
