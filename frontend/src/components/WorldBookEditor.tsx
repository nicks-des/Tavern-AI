import { useState, useEffect } from 'react'
import type { WorldBookEntry } from '../types'
import { worldbookApi } from '../api/worldbook'

export function WorldBookEditor({ characterId }: { characterId: string }) {
  const [entries, setEntries] = useState<WorldBookEntry[]>([])
  const [loading, setLoading] = useState(true)

  const [newKeywords, setNewKeywords] = useState('')
  const [newContent, setNewContent] = useState('')

  useEffect(() => {
    loadEntries()
  }, [characterId])

  const loadEntries = async () => {
    try {
      const list = await worldbookApi.list(characterId)
      setEntries(list)
    } catch {
      setEntries([])
    }
    setLoading(false)
  }

  const handleAdd = async () => {
    if (!newKeywords.trim() || !newContent.trim()) return
    try {
      const entry = await worldbookApi.create(characterId, newKeywords.trim(), newContent.trim())
      setEntries([...entries, entry])
      setNewKeywords('')
      setNewContent('')
    } catch {
      // ignore
    }
  }

  const handleToggle = async (entry: WorldBookEntry) => {
    try {
      await worldbookApi.update(entry.id, { enabled: !entry.enabled })
      setEntries(entries.map((e) =>
        e.id === entry.id ? { ...e, enabled: !e.enabled } : e
      ))
    } catch {
      // ignore
    }
  }

  const handleDelete = async (id: string) => {
    try {
      await worldbookApi.delete(id)
      setEntries(entries.filter((e) => e.id !== id))
    } catch {
      // ignore
    }
  }

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <label className="text-xs font-medium text-gray-400">世界书</label>
        <span className="text-xs text-gray-600">{entries.length} 条设定</span>
      </div>

      <div className="flex gap-2">
        <input
          type="text"
          value={newKeywords}
          onChange={(e) => setNewKeywords(e.target.value)}
          placeholder="关键词, 逗号分隔"
          className="flex-1 bg-tavern-800/60 border border-tavern-600/50 rounded-lg px-3 py-1.5 text-xs text-gray-100 placeholder-gray-500 focus:outline-none focus:border-accent/50"
        />
        <input
          type="text"
          value={newContent}
          onChange={(e) => setNewContent(e.target.value)}
          placeholder="触发内容"
          className="flex-[2] bg-tavern-800/60 border border-tavern-600/50 rounded-lg px-3 py-1.5 text-xs text-gray-100 placeholder-gray-500 focus:outline-none focus:border-accent/50"
        />
        <button
          onClick={handleAdd}
          disabled={!newKeywords.trim() || !newContent.trim()}
          className="px-3 py-1.5 bg-accent/20 hover:bg-accent/30 text-accent-light rounded-lg text-xs disabled:opacity-30 transition-colors"
        >
          + 添加
        </button>
      </div>

      {loading ? (
        <p className="text-xs text-gray-600">加载中...</p>
      ) : entries.length === 0 ? (
        <p className="text-xs text-gray-600">暂无设定，添加关键词来丰富角色的世界观</p>
      ) : (
        <div className="space-y-1.5 max-h-[200px] overflow-y-auto">
          {entries.map((entry) => (
            <div
              key={entry.id}
              className={`flex items-center gap-2 px-3 py-2 rounded-lg border transition-colors ${
                entry.enabled
                  ? 'bg-accent/5 border-accent/20'
                  : 'bg-tavern-800/40 border-tavern-700/50 opacity-50'
              }`}
            >
              <button
                onClick={() => handleToggle(entry)}
                className={`w-6 h-6 rounded flex items-center justify-center text-xs transition-colors ${
                  entry.enabled ? 'bg-accent/30 text-accent-light' : 'bg-tavern-700/50 text-gray-600'
                }`}
              >
                {entry.enabled ? 'ON' : 'OFF'}
              </button>
              <div className="min-w-0 flex-1">
                <span className="text-xs text-purple-400 font-mono">
                  {entry.keywords}
                </span>
                <span className="text-xs text-gray-500 mx-1.5">&rarr;</span>
                <span className="text-xs text-gray-300">{entry.content}</span>
              </div>
              <button
                onClick={() => handleDelete(entry.id)}
                className="text-xs text-gray-600 hover:text-rose-400 transition-colors shrink-0"
              >
                删除
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
