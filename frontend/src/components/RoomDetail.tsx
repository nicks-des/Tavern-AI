import { useState, useEffect, useRef } from 'react'
import { useStore } from '../store'
import { roomApi } from '../api/rooms'
import { sessionApi } from '../api/sessions'
import { CharacterAvatar } from './CharacterAvatar'
import { MessageBubble } from './MessageBubble'
import type { Room, RoomMember, Message } from '../types'

export function RoomDetail() {
  const { activeRoomId, setActiveRoom, characters, addCharacter, loadCharacters } = useStore()
  const [room, setRoom] = useState<Room | null>(null)
  const [members, setMembers] = useState<RoomMember[]>([])
  const [showAddMember, setShowAddMember] = useState(false)
  const [selectedCharId, setSelectedCharId] = useState<string>('')
  const [sessionId, setSessionId] = useState<string>('')
  const [messages, setMessages] = useState<Message[]>([])
  const [input, setInput] = useState('')
  const [streaming, setStreaming] = useState(false)
  const [streamContent, setStreamContent] = useState('')
  const [streamCharName, setStreamCharName] = useState('')
  const [autoRunning, setAutoRunning] = useState(false)
  const [paused, setPaused] = useState(false)
  const [showCreateChar, setShowCreateChar] = useState(false)
  const [newChar, setNewChar] = useState({ name: '', personality: '', goal: '', secret: '' })
  const abortRef = useRef<AbortController | null>(null)
  const autoAbortRef = useRef<AbortController | null>(null)
  const chatRef = useRef<HTMLDivElement>(null)
  const globalChars = characters.filter((c) => c.scope === 'global')

  useEffect(() => {
    if (!activeRoomId) return
    roomApi.get(activeRoomId).then((data) => {
      setRoom(data.room)
      setMembers(data.members)
    }).catch(() => {})

    fetch(`http://localhost:8081/api/rooms/${activeRoomId}/messages`)
      .then(r => r.json())
      .then((msgs: any[]) => {
        if (Array.isArray(msgs)) {
          setMessages(msgs.map((m: any) => ({
            id: m.id,
            role: 'assistant' as const,
            content: `**${m.characterName}**: ${m.content}`,
            timestamp: new Date(m.createdAt).getTime(),
          })))
        }
      })
      .catch(() => {})
  }, [activeRoomId])

  useEffect(() => {
    chatRef.current?.scrollTo(0, chatRef.current.scrollHeight)
  }, [messages, streamContent])

  const handleAddMember = async (characterId: string) => {
    if (!activeRoomId) return
    try {
      await roomApi.addMember(activeRoomId, characterId)
      setShowAddMember(false)
      const data = await roomApi.get(activeRoomId)
      setMembers(data.members)
    } catch {}
  }

  const handleRemoveMember = async (characterId: string) => {
    if (!activeRoomId) return
    await roomApi.removeMember(activeRoomId, characterId)
    setMembers(members.filter((m) => m.characterId !== characterId))
  }

  const handleCreateChar = async () => {
    if (!activeRoomId || !newChar.name.trim()) return
    try {
      const res = await fetch('http://localhost:8081/api/characters', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: newChar.name.trim(),
          description: '',
          personality: newChar.personality.trim(),
          goal: newChar.goal.trim(),
          secret: newChar.secret.trim(),
          scenario: '',
          tags: [],
        }),
      })
      if (!res.ok) throw new Error('Failed')
      const char = await res.json()
      addCharacter(char)
      await roomApi.addMember(activeRoomId, char.id)
      roomApi.get(activeRoomId).then((data) => setMembers(data.members))
      setNewChar({ name: '', personality: '', goal: '', secret: '' })
      setShowCreateChar(false)
      setShowAddMember(false)
    } catch (e) {
      console.error('Create char error:', e)
    }
  }

  const handleStop = () => {
    autoAbortRef.current?.abort()
    setAutoRunning(false)
    setPaused(true)
  }

  const handleAutoRun = async () => {
    if (!activeRoomId || autoRunning) return
    setPaused(false)
    setAutoRunning(true)
    setStreaming(true)
    if (!paused) setMessages([])

    try {
      const controller = new AbortController()
      autoAbortRef.current = controller

      const res = await fetch(`http://localhost:8081/api/rooms/${activeRoomId}/run`, { method: 'POST', signal: controller.signal })
      if (!res.ok) throw new Error(`HTTP ${res.status}`)

      const reader = res.body?.getReader()
      if (!reader) return

      const decoder = new TextDecoder('utf-8')
      let accumulated = ''
      let buffer = ''

      while (true) {
        const { done, value } = await reader.read()
        if (done) break
        buffer += decoder.decode(value, { stream: true })
        const lines = buffer.split('\n')
        buffer = lines.pop() || ''
        for (const line of lines) {
          if (!line.startsWith('data: ')) continue
          const d = line.slice(6).trim()
          if (d === '[DONE]') continue
          if (d.startsWith('[ERROR]')) continue

          if (d.startsWith('[') && d.endsWith(']') && d.length < 30) {
            const charName = d.slice(1, -1)
            const finalContent = accumulated
            if (finalContent) {
              setMessages((prev) => {
                const updated = [...prev]
                const last = updated[updated.length - 1]
                if (last && last.isStreaming) updated[updated.length - 1] = { ...last, content: finalContent, isStreaming: false }
                return updated
              })
            }
            accumulated = ''
            setStreamCharName(charName)
            setStreamContent('')
            setMessages((prev) => [...prev, {
              id: Date.now().toString() + Math.random(),
              role: 'assistant', content: '', timestamp: Date.now(), isStreaming: true,
            }])
            continue
          }

          accumulated += d
          setStreamContent(accumulated)
        }
      }

      if (accumulated) {
        const fc = accumulated
        setMessages((prev) => {
          const updated = [...prev]
          const last = updated[updated.length - 1]
          if (last && last.isStreaming) updated[updated.length - 1] = { ...last, content: fc, isStreaming: false }
          return updated
        })
      }
    } catch (err: unknown) {
      console.error('Auto run error:', err)
    } finally {
      setAutoRunning(false)
      setStreaming(false)
      setStreamContent('')
      setStreamCharName('')
      if (activeRoomId) {
        roomApi.get(activeRoomId).then((data) => setRoom(data.room))
      }
    }
  }

  const handleSend = async () => {
    if (!input.trim() || !selectedCharId || streaming || !activeRoomId) return
    const content = input.trim()
    setInput('')

    let sid = sessionId
    if (!sid) {
      try {
        const sess = await sessionApi.create(`${Date.now()}`, selectedCharId)
        sid = sess.id
        setSessionId(sid)
      } catch {
        sid = Date.now().toString()
        setSessionId(sid)
      }
    }

    const userMsg: Message = {
      id: Date.now().toString(),
      role: 'user',
      content,
      timestamp: Date.now(),
    }
    setMessages((prev) => [...prev, userMsg])

    const selectedChar = characters.find((c) => c.id === selectedCharId)
    const aiMsgId = (Date.now() + 1).toString()
    setMessages((prev) => [...prev, {
      id: aiMsgId,
      role: 'assistant',
      content: '',
      timestamp: Date.now(),
      isStreaming: true,
    }])
    setStreaming(true)
    setStreamContent('')
    setStreamCharName(selectedChar?.name ?? '')
    let accumulated = ''

    try {
      const controller = new AbortController()
      abortRef.current = controller

      const res = await sessionApi.chatStream(sid, content, selectedCharId, activeRoomId)
      if (!res.ok) throw new Error(`HTTP ${res.status}`)

      const reader = res.body?.getReader()
      if (!reader) throw new Error('No stream')

      const decoder = new TextDecoder('utf-8')
      let buffer = ''

      while (true) {
        const { done, value } = await reader.read()
        if (done) break
        buffer += decoder.decode(value, { stream: true })
        const lines = buffer.split('\n')
        buffer = lines.pop() || ''
        for (const line of lines) {
          if (!line.startsWith('data: ')) continue
          const data = line.slice(6).trim()
          if (data === '[DONE]') continue
          if (data.startsWith('[ERROR]')) continue

          // Auto-convo character name tag: data: [Ryu]
          if (data.startsWith('[') && data.endsWith(']') && data.length < 30) {
            const charName = data.slice(1, -1)
            const finalContent = accumulated
            setMessages((prev) => {
              const updated = [...prev]
              const last = updated[updated.length - 1]
              if (last && last.isStreaming) {
                updated[updated.length - 1] = { ...last, content: finalContent, isStreaming: false }
              }
              updated.push({
                id: Date.now().toString() + Math.random(),
                role: 'assistant',
                content: '',
                timestamp: Date.now(),
                isStreaming: true,
              })
              return updated
            })
            accumulated = ''
            setStreamCharName(charName)
            setStreamContent('')
            continue
          }

          accumulated += data
          setStreamContent(accumulated)
        }
      }

      const finalAfterLoop = accumulated
      setMessages((prev) => {
        const updated = [...prev]
        const last = updated[updated.length - 1]
        if (last && last.isStreaming) {
          updated[updated.length - 1] = { ...last, content: finalAfterLoop, isStreaming: false }
        }
        return updated
      })
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Unknown'
      setMessages((prev) => {
        const updated = [...prev]
        const last = updated[updated.length - 1]
        if (last && last.isStreaming) {
          updated[updated.length - 1] = { ...last, content: `[错误] ${msg}`, isStreaming: false }
        }
        return updated
      })
    } finally {
      setStreaming(false)
      const finalContent = accumulated
      if (finalContent) {
        setMessages((prev) => {
          const updated = [...prev]
          const last = updated[updated.length - 1]
          if (last && last.isStreaming) {
            updated[updated.length - 1] = { ...last, content: finalContent, isStreaming: false }
          }
          return updated
        })
      }
      setStreamContent('')
      setStreamCharName('')
      abortRef.current = null
    }
  }

  const selectedChar = characters.find((c) => c.id === selectedCharId)
  const addedCharIds = members.map((m) => m.characterId)
  const availableChars = globalChars.filter((c) => !addedCharIds.includes(c.id))

  return (
    <div className="flex-1 flex flex-col min-h-0">
      <header className="flex items-center gap-3 px-4 py-3 border-b border-tavern-700/50 shrink-0">
        <button onClick={() => setActiveRoom(null)} className="text-gray-400 hover:text-gray-200 text-sm">
          &larr; 返回
        </button>
        <span className="text-sm font-medium text-gray-200">{room?.name ?? ''}</span>
        <button
          onClick={handleAutoRun ? handleStop : handleAutoRun}
          disabled={members.length < 2}
          className={`ml-auto px-3 py-1 rounded-lg text-xs font-medium transition-colors ${autoRunning ? 'bg-amber-500/20 text-amber-400 hover:bg-amber-500/30' : paused ? 'bg-blue-500/20 text-blue-400 hover:bg-blue-500/30' : 'bg-emerald-500/20 text-emerald-400 hover:bg-emerald-500/30'} disabled:opacity-30`}
        >
          {autoRunning ? '暂停' : paused ? '继续运行' : '开始运行'}
        </button>
        <button
          onClick={async () => {
            if (!activeRoomId) return
            await fetch(`http://localhost:8081/api/rooms/${activeRoomId}/messages`, { method: 'DELETE' })
            setMessages([])
            if (activeRoomId) roomApi.get(activeRoomId).then((data) => setRoom(data.room))
          }}
          className="px-2 py-1 rounded-lg text-xs font-medium text-rose-400/60 hover:text-rose-400 hover:bg-rose-500/10 transition-colors"
        >
          重置
        </button>
        <button
          onClick={async () => {
            if (!activeRoomId) return
            const res = await fetch(`http://localhost:8081/api/rooms/${activeRoomId}/export`)
            const blob = await res.blob()
            const url = URL.createObjectURL(blob)
            const a = document.createElement('a')
            a.href = url
            a.download = `${room?.name ?? 'export'}.md`
            a.click()
            URL.revokeObjectURL(url)
          }}
          className="px-2 py-1 rounded-lg text-xs font-medium text-gray-400 hover:text-gray-200 hover:bg-tavern-700/50 transition-colors"
        >
          导出
        </button>
      </header>

      {room?.worldRules && (
        <div className="mx-4 mt-3 bg-emerald-500/5 border border-emerald-500/20 rounded-xl p-3 shrink-0">
          <p className="text-xs font-medium text-emerald-400 mb-1">世界规则</p>
          <p className="text-xs text-gray-300">{room.worldRules}</p>
        </div>
      )}

      {room?.worldState && room.worldState !== '{}' && !room.worldState.startsWith('{"round":0,') && (() => {
        try {
          const state = JSON.parse(room.worldState)
          const rels = state.relationships || {}
          const statuses = state.charStatus || {}
          const memories = state.charMemories || {}
          const events = (state.events || []).slice(-5)
          const time = state.time || '08:00'
          const day = state.day || 1
          const round = state.round || 0

          return (
            <div className="mx-4 mt-2 space-y-2 shrink-0 text-xs">
              {/* D: Clock Bar */}
              <div className="flex items-center gap-2 bg-gradient-to-r from-tavern-900/50 to-transparent rounded-lg px-3 py-1.5">
                <span className="px-2 py-0.5 rounded-full bg-amber-500/15 text-amber-300 font-medium border border-amber-500/20">Day {day}</span>
                <span className="text-gray-300 font-mono">{time}</span>
                <div className="flex-1 h-px bg-gradient-to-r from-transparent to-tavern-700/50 mx-2" />
                <span className="text-gray-600">R{round}</span>
              </div>

              {/* D: Character Status badges */}
              {Object.keys(statuses).length > 0 && (
                <div className="bg-teal-500/5 border border-teal-500/20 rounded-xl p-3">
                  <p className="font-medium text-teal-400 mb-2">角色状态</p>
                  <div className="flex flex-wrap gap-1.5">
                    {Object.entries(statuses).map(([name, status]) => (
                      <span key={name} className={`inline-flex gap-1 px-2.5 py-1 rounded-full border text-[11px] ${
                        String(status) === 'dead' ? 'bg-red-500/10 text-red-400 border-red-500/20' :
                        String(status) === 'dying' ? 'bg-orange-500/10 text-orange-400 border-orange-500/20 animate-pulse' :
                        'bg-teal-500/10 text-teal-300 border-teal-500/20'
                      }`}>
                        <span className="opacity-60">{name}</span>
                        <span className="font-bold">{status}</span>
                      </span>
                    ))}
                  </div>
                </div>
              )}

              {/* D: Relationship Graph */}
              {Object.keys(rels).length > 0 && (
                <div className="bg-purple-500/5 border border-purple-500/20 rounded-xl p-3">
                  <p className="font-medium text-purple-400 mb-2">关系图谱</p>
                  <div className="space-y-1.5">
                    {Object.entries(rels as Record<string, {trust?: number; fear?: number; tag?: string}>).map(([pair, r]) => {
                      const parts = pair.split('::')
                      const t = r.trust ?? 5
                      const f = r.fear ?? 0
                      return (
                        <div key={pair} className="flex items-center gap-1.5">
                          <span className="text-gray-400 w-14 truncate">{parts[0]}</span>
                          <span className="text-gray-600 text-[10px]">→</span>
                          <span className="text-gray-400 w-14 truncate">{parts[1] || ''}</span>
                          <div className="flex-1 h-1.5 bg-tavern-800/50 rounded-full mx-1 overflow-hidden">
                            <div className="flex h-full">
                              <div className="bg-emerald-500/70 rounded-full transition-all" style={{width: `${t*10}%`}} />
                              <div className="bg-orange-500/70 rounded-full transition-all" style={{width: `${f*10}%`}} />
                            </div>
                          </div>
                          <div className="flex gap-1.5 shrink-0">
                            <span className="text-emerald-400 w-8 text-right">❤{t}</span>
                            <span className="text-orange-400 w-8 text-right">⚠{f}</span>
                          </div>
                        </div>
                      )
                    })}
                  </div>
                </div>
              )}

              {/* D: Events Timeline */}
              {events.length > 0 && (
                <div className="bg-amber-500/5 border border-amber-500/20 rounded-xl p-3">
                  <p className="font-medium text-amber-400 mb-2">时间线</p>
                  <div className="space-y-1 pl-3 border-l border-amber-500/10">
                    {events.map((e: string, i: number) => {
                      const isWorld = e.startsWith('[WORLD]')
                      const isLegacy = e.startsWith('[LEGACY]')
                      return (
                        <div key={i} className={`relative pl-3 py-0.5 ${isWorld ? 'text-yellow-300/80' : isLegacy ? 'text-purple-300/80' : 'text-amber-300/70'}`}>
                          <div className={`absolute left-[-5px] top-[6px] w-2 h-2 rounded-full border ${isWorld ? 'bg-yellow-500 border-yellow-400' : isLegacy ? 'bg-purple-500 border-purple-400' : 'bg-amber-500/50 border-amber-400/50'}`} />
                          <span className="leading-relaxed">{e}</span>
                        </div>
                      )
                    })}
                  </div>
                </div>
              )}
            </div>
          )
        } catch { return null }
      })()}

      {/* Character selector */}
      <div className="px-4 py-2 flex items-center gap-3 border-b border-tavern-700/50 shrink-0">
        <select
          value={selectedCharId}
          onChange={(e) => setSelectedCharId(e.target.value)}
          className="bg-tavern-800/60 border border-tavern-600/50 rounded-lg px-3 py-1.5 text-sm text-gray-200 focus:outline-none focus:border-accent/50"
        >
          <option value="">选择发言角色...</option>
          {members.map((m) => {
            const c = characters.find((ch) => ch.id === m.characterId)
            return <option key={m.characterId} value={m.characterId}>{c?.name ?? m.characterId}</option>
          })}
        </select>

        <div className="flex items-center gap-1 ml-auto">
          {members.slice(0, 4).map((m) => {
            const c = characters.find((ch) => ch.id === m.characterId)
            return <CharacterAvatar key={m.characterId} name={c?.name ?? '?'} size="sm" />
          })}
          {members.length > 4 && <span className="text-xs text-gray-500">+{members.length - 4}</span>}
        </div>

        <button onClick={() => setShowAddMember(!showAddMember)} className="text-xs text-accent-light hover:text-accent">
          {showAddMember ? '取消' : '+ 成员'}
        </button>
      </div>

      {showAddMember && (
        <div className="mx-4 mt-2 p-3 bg-tavern-800/40 border border-tavern-700/50 rounded-xl max-h-[280px] overflow-y-auto shrink-0 space-y-2">
          {showCreateChar ? (
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <button onClick={() => setShowCreateChar(false)} className="text-xs text-gray-400 hover:text-gray-200">&larr; 返回</button>
                <span className="text-xs font-medium text-accent-light">快速创建角色</span>
              </div>
              <input
                value={newChar.name}
                onChange={(e) => setNewChar({ ...newChar, name: e.target.value })}
                placeholder="角色名称*"
                className="w-full bg-tavern-900/60 border border-tavern-600/50 rounded-lg px-3 py-1.5 text-sm text-gray-200 placeholder-gray-500 focus:outline-none focus:border-accent/50"
              />
              <input
                value={newChar.personality}
                onChange={(e) => setNewChar({ ...newChar, personality: e.target.value })}
                placeholder="性格特点"
                className="w-full bg-tavern-900/60 border border-tavern-600/50 rounded-lg px-3 py-1.5 text-sm text-gray-200 placeholder-gray-500 focus:outline-none focus:border-accent/50"
              />
              <input
                value={newChar.goal}
                onChange={(e) => setNewChar({ ...newChar, goal: e.target.value })}
                placeholder="目标"
                className="w-full bg-tavern-900/60 border border-tavern-600/50 rounded-lg px-3 py-1.5 text-sm text-gray-200 placeholder-gray-500 focus:outline-none focus:border-accent/50"
              />
              <input
                value={newChar.secret}
                onChange={(e) => setNewChar({ ...newChar, secret: e.target.value })}
                placeholder="秘密"
                className="w-full bg-tavern-900/60 border border-tavern-600/50 rounded-lg px-3 py-1.5 text-sm text-gray-200 placeholder-gray-500 focus:outline-none focus:border-accent/50"
              />
              <button
                onClick={handleCreateChar}
                disabled={!newChar.name.trim()}
                className="w-full py-1.5 rounded-lg text-xs font-medium bg-accent hover:bg-blue-600 text-white transition-colors disabled:opacity-30"
              >
                创建并加入房间
              </button>
            </div>
          ) : (
            <>
              {availableChars.length === 0 ? (
                <p className="text-xs text-gray-600 text-center py-2">所有角色已在房间中</p>
              ) : (
                availableChars.map((c) => (
                  <button key={c.id} onClick={() => handleAddMember(c.id)}
                    className="w-full flex items-center gap-2 p-2 rounded-lg hover:bg-tavern-700/50 text-left">
                    <CharacterAvatar name={c.name} size="sm" />
                    <span className="text-xs text-gray-300">{c.name}</span>
                    <span className="text-xs text-accent ml-auto">添加</span>
                  </button>
                ))
              )}
              <button
                onClick={() => setShowCreateChar(true)}
                className="w-full py-2 rounded-lg text-xs font-medium text-accent-light border border-dashed border-accent/30 hover:border-accent/50 hover:bg-accent/5 transition-colors"
              >
                + 快速创建角色
              </button>
            </>
          )}
        </div>
      )}

      {/* Chat area */}
      <div ref={chatRef} className="flex-1 overflow-y-auto py-3 min-h-0">
        {messages.length === 0 && !streaming && (
          <div className="flex items-center justify-center h-full">
            <p className="text-sm text-gray-500">选择角色后开始对话，或点击「开始运行」</p>
          </div>
        )}
        {messages.map((msg) => {
          const charMatch = msg.content.match(/^\*\*(.+?)\*\*:\s*/)
          const charName = charMatch ? charMatch[1] : ''
          const displayContent = charMatch ? msg.content.slice(charMatch[0].length) : msg.content

          if (msg.role === 'user') {
            return (
              <div key={msg.id} className="flex gap-3 px-4 py-2 justify-end">
                <div className="max-w-[70%] bg-accent text-white rounded-2xl rounded-br-md px-4 py-2.5 text-sm">
                  <p className="whitespace-pre-wrap">{msg.content}</p>
                </div>
              </div>
            )
          }

          if (msg.isStreaming) {
            return null // handled by streaming bubble below
          }

          return (
            <div key={msg.id} className="flex gap-2.5 px-4 py-2 items-start">
              <CharacterAvatar name={charName || 'AI'} size="sm" />
              <div className="max-w-[75%]">
                {charName && <p className="text-xs font-medium text-gray-400 mb-1">{charName}</p>}
                <div className="bg-tavern-800/60 rounded-2xl rounded-bl-md px-4 py-2.5 text-sm text-gray-100">
                  <p className="whitespace-pre-wrap">{displayContent}</p>
                </div>
              </div>
            </div>
          )
        })}
        {streaming && streamContent && (
          <div className="flex gap-2.5 px-4 py-2 items-start">
            {streamCharName && <CharacterAvatar name={streamCharName} size="sm" />}
            <div className="max-w-[75%]">
              {streamCharName && <p className="text-xs font-medium text-accent-light mb-1">{streamCharName}</p>}
              <div className="bg-tavern-800/60 rounded-2xl rounded-bl-md px-4 py-2.5">
                <div className="text-sm text-gray-100 whitespace-pre-wrap">
                  {streamContent}
                  <span className="typing-cursor" />
                </div>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Input */}
      <div className="border-t border-tavern-700/50 p-4 shrink-0">
        <div className="flex items-end gap-3">
          <textarea
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); handleSend() }
            }}
            placeholder={selectedCharId ? '输入消息（其他角色会自动回应）...' : '请先选择发言角色'}
            disabled={!selectedCharId || streaming}
            rows={1}
            className="flex-1 bg-tavern-800/60 border border-tavern-600/50 rounded-xl px-4 py-3 text-sm text-gray-100 placeholder-gray-500 resize-none focus:outline-none focus:border-accent/50 disabled:opacity-50"
          />
          <button
            onClick={handleSend}
            disabled={!input.trim() || !selectedCharId || streaming}
            className="px-4 py-3 bg-accent hover:bg-accent-dark text-white rounded-xl text-sm font-medium disabled:opacity-40 disabled:cursor-not-allowed"
          >
            {streaming ? '...' : '发送'}
          </button>
        </div>
      </div>
    </div>
  )
}
