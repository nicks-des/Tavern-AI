import ReactMarkdown from 'react-markdown'
import type { Message } from '../types'
import { CharacterAvatar } from './CharacterAvatar'
import { useStore } from '../store'

export function MessageBubble({ message }: { message: Message }) {
  const character = useStore((s) => s.getActiveCharacter())
  const isUser = message.role === 'user'

  return (
    <div className={`message-enter flex gap-3 px-4 py-3 ${isUser ? 'flex-row-reverse' : ''}`}>
      {!isUser && <CharacterAvatar name={character?.name ?? 'AI'} size="sm" />}

      <div
        className={`max-w-[75%] rounded-2xl px-4 py-3 text-sm leading-relaxed ${
          isUser
            ? 'bg-accent text-white rounded-br-md'
            : 'bg-tavern-800/60 text-gray-100 rounded-bl-md'
        }`}
      >
        {message.isStreaming ? (
          <span className="whitespace-pre-wrap">
            <ReactMarkdown className="prose-sm prose-invert">
              {message.content || ''}
            </ReactMarkdown>
            <span className="typing-cursor" />
          </span>
        ) : (
          <div className="whitespace-pre-wrap">
            <ReactMarkdown className="prose-sm prose-invert">
              {message.content}
            </ReactMarkdown>
          </div>
        )}
      </div>
    </div>
  )
}
