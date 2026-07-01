import { useState, useEffect } from 'react'
import { useStore } from '../store'
import { CharacterAvatar } from './CharacterAvatar'
import { CharacterForm } from './CharacterForm'
import { roomApi } from '../api/rooms'
import type { Character, Room } from '../types'

type SidebarTab = 'chars' | 'chats' | 'rooms'

export function Sidebar() {
  const store = useStore()
  const { characters, sessions, activeSessionId, setActiveSession, createSession, setCurrentView, addCharacter, getGlobalCharacters, setActiveRoom } = store
  const [tab, setTab] = useState<SidebarTab>('chars')
  const [showCreate, setShowCreate] = useState(false)
  const [showCreateRoom, setShowCreateRoom] = useState(false)
  const [rooms, setRooms] = useState<Room[]>([])
  const globalChars = getGlobalCharacters()

  useEffect(() => {
    if (tab === 'rooms') {
      roomApi.list().then(setRooms).catch(() => {})
    }
  }, [tab])

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
        {(['chars', 'rooms', 'chats'] as const).map((t) => (
          <button
            key={t}
            onClick={() => setTab(t)}
            className={`flex-1 py-2.5 text-xs font-medium transition-colors ${
              tab === t ? 'text-accent border-b-2 border-accent' : 'text-gray-500 hover:text-gray-300'
            }`}
          >
            {{ chars: '角色', rooms: '房间', chats: '对话' }[t]}
          </button>
        ))}
      </div>

      <div className="flex-1 overflow-y-auto">
        {tab === 'chars' && (
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
              <button onClick={() => setShowCreate(true)} className="w-full py-2.5 bg-accent/20 hover:bg-accent/30 text-accent-light rounded-lg text-xs font-medium transition-colors">
                + 创建角色
              </button>
              <button onClick={() => setCurrentView('manage')} className="w-full py-2 border border-tavern-600/50 hover:border-accent/50 text-gray-500 hover:text-gray-300 rounded-lg text-xs transition-colors">
                管理角色
              </button>
            </div>
          </div>
        )}

        {tab === 'rooms' && (
          <div className="flex flex-col h-full">
            <div className="p-2 flex-1">
              {rooms.length === 0 && (
                <p className="text-xs text-gray-500 text-center py-8">暂无房间，创建一个开始冒险</p>
              )}
              {rooms.map((room) => (
                <button
                  key={room.id}
                  onClick={() => setActiveRoom(room.id)}
                  className="w-full flex items-center gap-3 p-3 rounded-lg hover:bg-tavern-800/60 transition-colors text-left"
                >
                  <div className="w-8 h-8 rounded-lg bg-emerald-500/20 flex items-center justify-center text-sm">
                    &#x1F3E0;
                  </div>
                  <div className="min-w-0 flex-1">
                    <p className="text-sm font-medium text-gray-200 truncate">{room.name}</p>
                    <p className="text-xs text-gray-500 truncate">{room.description || '无简介'}</p>
                  </div>
                </button>
              ))}
            </div>
            <div className="p-2 border-t border-tavern-700/50">
              <button onClick={() => setShowCreateRoom(true)} className="w-full py-2.5 bg-emerald-500/20 hover:bg-emerald-500/30 text-emerald-400 rounded-lg text-xs font-medium transition-colors">
                + 创建房间
              </button>
            </div>
          </div>
        )}

        {tab === 'chats' && (
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
        <p className="text-xs text-gray-600 text-center">MVP v0.2</p>
      </div>

      {showCreate && (
        <CharacterForm onSave={handleCreateChar} onCancel={() => setShowCreate(false)} />
      )}

      {showCreateRoom && (
        <CreateRoomModal
          onClose={() => setShowCreateRoom(false)}
          onCreated={(room) => {
            setRooms([room, ...rooms])
            setShowCreateRoom(false)
          }}
        />
      )}
    </aside>
  )
}

function CreateRoomModal({ onClose, onCreated }: { onClose: () => void; onCreated: (r: Room) => void }) {
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [worldRules, setWorldRules] = useState('')

  const handleCreate = async () => {
    if (!name.trim()) return
    try {
      const room = await roomApi.create(name.trim(), description.trim(), worldRules.trim())
      onCreated(room)
    } catch {
      // ignore
    }
  }

  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4">
      <div className="bg-tavern-900 border border-tavern-700/50 rounded-2xl w-full max-w-md shadow-2xl">
        <div className="p-6 border-b border-tavern-700/50">
          <h2 className="text-base font-semibold text-gray-100">创建房间</h2>
        </div>

        <div className="p-6 space-y-4">
          <div>
            <label className="block text-xs font-medium text-gray-400 mb-1.5">房间名称 *</label>
            <input type="text" value={name} onChange={(e) => setName(e.target.value)} placeholder="如：冒险者酒馆" className="w-full bg-tavern-800/60 border border-tavern-600/50 rounded-lg px-3 py-2 text-sm text-gray-100 placeholder-gray-500 focus:outline-none focus:border-accent/50" />
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-400 mb-1.5">房间简介</label>
            <textarea value={description} onChange={(e) => setDescription(e.target.value)} placeholder="一句话描述这个房间" rows={2} className="w-full bg-tavern-800/60 border border-tavern-600/50 rounded-lg px-3 py-2 text-sm text-gray-100 placeholder-gray-500 resize-none focus:outline-none focus:border-accent/50" />
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-400 mb-1.5">世界规则</label>
            <textarea value={worldRules} onChange={(e) => setWorldRules(e.target.value)} placeholder="描述这个房间的世界规则和场景设定" rows={3} className="w-full bg-tavern-800/60 border border-tavern-600/50 rounded-lg px-3 py-2 text-sm text-gray-100 placeholder-gray-500 resize-none focus:outline-none focus:border-accent/50" />
          </div>
        </div>

        <div className="flex gap-3 px-6 py-4 border-t border-tavern-700/50">
          <button onClick={onClose} className="flex-1 py-2.5 border border-tavern-600/50 rounded-lg text-sm text-gray-400 hover:text-gray-200 transition-colors">取消</button>
          <button onClick={handleCreate} disabled={!name.trim()} className="flex-1 py-2.5 bg-accent hover:bg-accent-dark text-white rounded-lg text-sm font-medium disabled:opacity-40 transition-colors">创建</button>
        </div>
      </div>
    </div>
  )
}
