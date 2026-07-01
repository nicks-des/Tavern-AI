import { useState } from 'react'
import { useStore } from '../store'
import { CharacterAvatar } from './CharacterAvatar'
import { CharacterForm } from './CharacterForm'
import type { Character } from '../types'

type SidebarTab = 'chars' | 'chats'

export function Sidebar() {
  const store = useStore()
  const { characters, sessions, activeSessionId, setActiveSession, createSession, setCurrentView, addCharacter, getGlobalCharacters } = store
  const [tab, setTab] = useState<SidebarTab>('chars')
  const [showCreate, setShowCreate] = useState(false)
  const globalChars = getGlobalCharacters()

  const handleCreateChar = (data: Omit<Character, 'id' | 'createdAt'>) => {
    addCharacter({
      ...data,
      id: Date.now().toString(),
      scope: 'global',
      createdAt: new Date().toISOString(),
    })
    setShowCreate(false)
  }

  return (
    <aside className="w-72 bg-tavern-900 border-r border-tavern-700/50 flex flex-col shrink-0">
      <div className="p-4 border-b border-tavern-700/50">
        <h1 className="text-base font-semibold text-gray-100 flex items-center gap-2">
          <span className="text-lg">&#x1F3FA;</span> Tavern AI
        </h1>
      </div>

      <div className="flex border-b border-tavern-700/50">
        <button
          onClick={() => setTab('chars')}
          className={`flex-1 py-2.5 text-xs font-medium transition-colors ${
            tab === 'chars' ? 'text-accent border-b-2 border-accent' : 'text-gray-500 hover:text-gray-300'
          }`}
        >
          角色
        </button>
        <button
          onClick={() => setTab('chats')}
          className={`flex-1 py-2.5 text-xs font-medium transition-colors ${
            tab === 'chats' ? 'text-accent border-b-2 border-accent' : 'text-gray-500 hover:text-gray-300'
          }`}
        >
          对话
        </button>
      </div>

      <div className="flex-1 overflow-y-auto">
        {tab === 'chars' ? (
          <div className="flex flex-col h-full">
            <div className="p-2 flex-1">
              {globalChars.map((char) => (
                <button
                  key={char.id}
                  onClick={() => createSession(char.id)}
                  className="w-full flex items-center gap-3 p-3 rounded-lg hover:bg-tavern-800/60 transition-colors text-left group"
                >
                  <CharacterAvatar name={char.name} size="sm" />
                  <div className="min-w-0 flex-1">
                    <p className="text-sm font-medium text-gray-200 truncate">{char.name}</p>
                    <p className="text-xs text-gray-500 truncate">{char.description}</p>
                  </div>
                </button>
              ))}
            </div>
            <div className="p-2 border-t border-tavern-700/50 space-y-1">
              <button
                onClick={() => setShowCreate(true)}
                className="w-full py-2.5 bg-accent/20 hover:bg-accent/30 text-accent-light rounded-lg text-xs font-medium transition-colors"
              >
                + 创建角色
              </button>
              <button
                onClick={() => setCurrentView('manage')}
                className="w-full py-2 border border-tavern-600/50 hover:border-accent/50 text-gray-500 hover:text-gray-300 rounded-lg text-xs transition-colors"
              >
                管理角色
              </button>
            </div>
          </div>
        ) : (
          <div className="p-2">
            {sessions.length === 0 && (
              <p className="text-xs text-gray-500 text-center py-8">暂无对话，选择一个角色开始</p>
            )}
            {sessions.map((s) => {
              const char = characters.find((c) => c.id === s.characterId)
              const isActive = s.id === activeSessionId
              return (
                <button
                  key={s.id}
                  onClick={() => setActiveSession(s.id)}
                  className={`w-full flex items-center gap-3 p-3 rounded-lg transition-colors text-left ${
                    isActive ? 'bg-accent/20 border border-accent/30' : 'hover:bg-tavern-800/60'
                  }`}
                >
                  <CharacterAvatar name={char?.name ?? '?'} size="sm" />
                  <div className="min-w-0 flex-1">
                    <p className="text-sm font-medium text-gray-200 truncate">{s.title}</p>
                    <p className="text-xs text-gray-500">{s.messages.length} 条消息</p>
                  </div>
                </button>
              )
            })}
          </div>
        )}
      </div>

      <div className="p-3 border-t border-tavern-700/50">
        <p className="text-xs text-gray-600 text-center">MVP v0.1</p>
      </div>

      {showCreate && (
        <CharacterForm onSave={handleCreateChar} onCancel={() => setShowCreate(false)} />
      )}
    </aside>
  )
}
