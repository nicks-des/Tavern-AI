import { useEffect, useRef } from 'react'
import { useStore } from '../store'
import { MessageBubble } from './MessageBubble'
import { ChatInput } from './ChatInput'

export function ChatArea() {
  const activeSession = useStore((s) => s.getActiveSession())
  const character = useStore((s) => s.getActiveCharacter())
  const containerRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (containerRef.current) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight
    }
  }, [activeSession?.messages])

  return (
    <div className="flex-1 flex flex-col min-h-0">
      <header className="flex items-center gap-3 px-4 py-3 border-b border-tavern-700/50">
        <div className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
        <span className="text-sm font-medium text-gray-200">
          {character?.name ?? 'AI'}
        </span>
        <span className="text-xs text-gray-500 ml-auto">
          {activeSession?.messages.length ?? 0} 条消息
        </span>
      </header>

      <div ref={containerRef} className="flex-1 overflow-y-auto py-2">
        {activeSession?.messages.length === 0 && (
          <div className="flex items-center justify-center h-full">
            <div className="text-center text-gray-500">
              <p className="text-lg mb-2">{character?.firstMessage ?? '开始一段对话吧'}</p>
              <p className="text-sm">发送第一条消息开始对话</p>
            </div>
          </div>
        )}
        {activeSession?.messages.map((msg) => (
          <MessageBubble key={msg.id} message={msg} />
        ))}
      </div>

      <ChatInput />
    </div>
  )
}
