import { useState, useRef } from 'react'
import { useStore } from '../store'
import { CharacterAvatar } from './CharacterAvatar'
import { CharacterForm } from './CharacterForm'
import type { Character } from '../types'

export function CharacterManager({ onClose }: { onClose: () => void }) {
  const { characters, addCharacter, updateCharacter, deleteCharacter, getGlobalCharacters } = useStore()
  const globalChars = getGlobalCharacters()
  const [showForm, setShowForm] = useState(false)
  const [editingChar, setEditingChar] = useState<Character | undefined>()
  const [importMsg, setImportMsg] = useState('')
  const fileRef = useRef<HTMLInputElement>(null)

  const handleCreate = () => {
    setEditingChar(undefined)
    setShowForm(true)
  }

  const handleEdit = (c: Character) => {
    setEditingChar(c)
    setShowForm(true)
  }

  const handleSave = (
    data: Omit<Character, 'id' | 'createdAt'>
  ) => {
    if (editingChar) {
      updateCharacter(editingChar.id, data)
    } else {
      addCharacter({
        ...data,
        id: Date.now().toString(),
        createdAt: new Date().toISOString(),
      })
    }
    setShowForm(false)
    setEditingChar(undefined)
  }

  const handleImport = () => {
    fileRef.current?.click()
  }

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    const reader = new FileReader()
    reader.onload = () => {
      try {
        const data = JSON.parse(reader.result as string)
        if (!data.name || !data.description) {
          setImportMsg('角色卡格式不正确，缺少 name 或 description')
          return
        }
        addCharacter({
          id: Date.now().toString(),
          name: data.name,
          avatar: data.avatar ?? '',
          portrait: data.portrait ?? '',
          catchphrase: data.catchphrase ?? '',
          description: data.description,
          personality: data.personality ?? '',
          scenario: data.scenario ?? '',
          firstMessage: data.first_mes ?? data.firstMessage ?? '',
          exampleDialogue: data.mes_example ?? data.exampleDialogue ?? '',
          tags: data.tags ?? [],
          scope: 'global',
          createdAt: new Date().toISOString(),
        })
        setImportMsg('导入成功!')
        setTimeout(() => setImportMsg(''), 2000)
      } catch {
        setImportMsg('JSON 解析失败，请检查文件格式')
        setTimeout(() => setImportMsg(''), 3000)
      }
    }
    reader.readAsText(file)
    e.target.value = ''
  }

  const handleDelete = (id: string, name: string) => {
    if (confirm(`确定要删除角色 "${name}" 吗？此操作不可撤销。`)) {
      deleteCharacter(id)
    }
  }

  const handleExport = (c: Character) => {
    const blob = new Blob(
      [JSON.stringify({
        name: c.name,
        description: c.description,
        catchphrase: c.catchphrase,
        personality: c.personality,
        scenario: c.scenario,
        first_mes: c.firstMessage,
        mes_example: c.exampleDialogue,
        tags: c.tags,
      }, null, 2)],
      { type: 'application/json' }
    )
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${c.name}.json`
    a.click()
    URL.revokeObjectURL(url)
  }

  return (
    <div className="flex-1 flex flex-col min-h-0">
      <header className="flex items-center gap-3 px-4 py-3 border-b border-tavern-700/50">
        <button
          onClick={onClose}
          className="text-gray-400 hover:text-gray-200 text-sm transition-colors"
        >
          &larr; 返回
        </button>
        <span className="text-sm font-medium text-gray-200">角色管理</span>
        <span className="text-xs text-gray-500 ml-auto">{globalChars.length} 个角色</span>
      </header>

      <div className="flex items-center gap-2 px-4 py-3 border-b border-tavern-700/50">
        <button
          onClick={handleCreate}
          className="px-4 py-1.5 bg-accent hover:bg-accent-dark text-white rounded-lg text-xs font-medium transition-colors"
        >
          + 创建角色
        </button>
        <button
          onClick={handleImport}
          className="px-4 py-1.5 border border-tavern-600/50 hover:border-accent/50 text-gray-400 hover:text-gray-200 rounded-lg text-xs transition-colors"
        >
          导入 JSON
        </button>
        <input ref={fileRef} type="file" accept=".json" onChange={handleFileChange} className="hidden" />
        {importMsg && (
          <span
            className={`text-xs ml-2 ${
              importMsg.includes('成功') ? 'text-emerald-400' : 'text-rose-400'
            }`}
          >
            {importMsg}
          </span>
        )}
      </div>

      <div className="flex-1 overflow-y-auto p-4">
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
          {globalChars.map((c) => (
            <div
              key={c.id}
              className="bg-tavern-800/40 border border-tavern-700/50 rounded-xl p-4 hover:border-accent/30 transition-colors group"
            >
              <div className="flex items-start gap-3 mb-3">
                <CharacterAvatar name={c.name} size="md" />
                <div className="min-w-0 flex-1">
                  <h3 className="text-sm font-medium text-gray-200">{c.name}</h3>
                  {c.catchphrase && (
                    <p className="text-xs text-accent-light italic mt-0.5 truncate">
                      &ldquo;{c.catchphrase}&rdquo;
                    </p>
                  )}
                  <p className="text-xs text-gray-500 mt-1 line-clamp-2">{c.description}</p>
                </div>
              </div>

              {c.tags.length > 0 && (
                <div className="flex flex-wrap gap-1 mb-3">
                  {c.tags.map((tag) => (
                    <span
                      key={tag}
                      className="px-2 py-0.5 bg-accent/10 text-accent-light text-xs rounded-full"
                    >
                      {tag}
                    </span>
                  ))}
                </div>
              )}

              <div className="flex gap-2">
                <button
                  onClick={() => handleEdit(c)}
                  className="px-3 py-1 text-xs text-gray-400 hover:text-gray-200 bg-tavern-700/50 hover:bg-tavern-700 rounded-lg transition-colors"
                >
                  编辑
                </button>
                <button
                  onClick={() => handleExport(c)}
                  className="px-3 py-1 text-xs text-gray-400 hover:text-gray-200 bg-tavern-700/50 hover:bg-tavern-700 rounded-lg transition-colors"
                >
                  导出
                </button>
                <button
                  onClick={() => handleDelete(c.id, c.name)}
                  className="px-3 py-1 text-xs text-rose-400 hover:text-rose-300 bg-rose-500/10 hover:bg-rose-500/20 rounded-lg transition-colors ml-auto"
                >
                  删除
                </button>
              </div>
            </div>
          ))}
        </div>
      </div>

      {showForm && (
        <CharacterForm
          character={editingChar}
          onSave={handleSave}
          onCancel={() => {
            setShowForm(false)
            setEditingChar(undefined)
          }}
        />
      )}
    </div>
  )
}
