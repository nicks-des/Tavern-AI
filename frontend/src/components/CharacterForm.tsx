import { useState, useRef } from 'react'
import type { Character } from '../types'

interface Props {
  character?: Character
  onSave: (data: Omit<Character, 'id' | 'createdAt'>) => void
  onCancel: () => void
}

export function CharacterForm({ character, onSave, onCancel }: Props) {
  const [avatar, setAvatar] = useState(character?.avatar ?? '')
  const [portrait, setPortrait] = useState(character?.portrait ?? '')
  const [name, setName] = useState(character?.name ?? '')
  const [catchphrase, setCatchphrase] = useState(character?.catchphrase ?? '')
  const [personality, setPersonality] = useState(character?.personality ?? '')
  const [description, setDescription] = useState(character?.description ?? '')
  const [scenario, setScenario] = useState(character?.scenario ?? '')
  const [firstMessage, setFirstMessage] = useState(character?.firstMessage ?? '')
  const [exampleDialogue, setExampleDialogue] = useState(character?.exampleDialogue ?? '')
  const [goal, setGoal] = useState(character?.goal ?? '')
  const [secret, setSecret] = useState(character?.secret ?? '')
  const [tags, setTags] = useState(character?.tags.join(', ') ?? '')
  const [showAdvanced, setShowAdvanced] = useState(false)
  const avatarInputRef = useRef<HTMLInputElement>(null)
  const portraitInputRef = useRef<HTMLInputElement>(null)

  const handleFileToDataUrl = (file: File, cb: (url: string) => void) => {
    const reader = new FileReader()
    reader.onload = () => cb(reader.result as string)
    reader.readAsDataURL(file)
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim() || !description.trim()) return
    onSave({
      name: name.trim(),
      avatar,
      portrait,
      catchphrase: catchphrase.trim(),
      description: description.trim(),
      personality: personality.trim(),
      scenario: scenario.trim(),
      firstMessage: firstMessage.trim(),
      exampleDialogue: exampleDialogue.trim(),
      goal: goal.trim(),
      secret: secret.trim(),
      tags: tags.split(',').map((t) => t.trim()).filter(Boolean),
    })
  }

  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4">
      <div className="bg-tavern-900 border border-tavern-700/50 rounded-2xl w-full max-w-3xl max-h-[90vh] overflow-hidden shadow-2xl flex flex-col">
        <div className="px-6 py-4 border-b border-tavern-700/50 shrink-0">
          <h2 className="text-base font-semibold text-gray-100">
            {character ? '编辑角色' : '创建角色'}
          </h2>
        </div>

        <form onSubmit={handleSubmit} className="flex-1 overflow-y-auto">
          <div className="flex flex-row min-h-0">
            {/* ===== 左侧：立绘区 ===== */}
            <div className="w-[42%] p-6 flex flex-col items-center border-r border-tavern-700/50">
              <label className="block text-xs font-medium text-gray-400 mb-3 self-start">
                角色立绘
              </label>

              <div className="relative w-full aspect-[3/4] max-w-[240px] mb-4">
                {portrait ? (
                  <div className="w-full h-full rounded-xl overflow-hidden border border-tavern-600/50 group relative">
                    <img
                      src={portrait}
                      alt="角色立绘"
                      className="w-full h-full object-cover"
                    />
                    <div className="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center gap-2">
                      <button
                        type="button"
                        onClick={() => portraitInputRef.current?.click()}
                        className="px-3 py-1.5 bg-tavern-800/80 text-gray-200 rounded-lg text-xs hover:bg-tavern-700/80 transition-colors"
                      >
                        更换
                      </button>
                      <button
                        type="button"
                        onClick={() => setPortrait('')}
                        className="px-3 py-1.5 bg-rose-500/40 text-rose-300 rounded-lg text-xs hover:bg-rose-500/60 transition-colors"
                      >
                        移除
                      </button>
                    </div>
                  </div>
                ) : (
                  <button
                    type="button"
                    onClick={() => portraitInputRef.current?.click()}
                    className="w-full h-full border-2 border-dashed border-tavern-600/50 rounded-xl flex flex-col items-center justify-center gap-3 hover:border-accent/50 hover:bg-accent/5 transition-colors group"
                  >
                    <div className="text-3xl text-gray-600 group-hover:text-accent-light transition-colors">
                      &#x1F5BC;
                    </div>
                    <span className="text-xs text-gray-500 group-hover:text-gray-400 transition-colors">
                      点击上传立绘
                    </span>
                    <span className="text-xs text-gray-600">
                      建议比例 3:4
                    </span>
                  </button>
                )}
              </div>

              <input
                ref={portraitInputRef}
                type="file"
                accept="image/*"
                onChange={(e) => {
                  const f = e.target.files?.[0]
                  if (f) handleFileToDataUrl(f, setPortrait)
                  e.target.value = ''
                }}
                className="hidden"
              />

              <div className="flex gap-2 w-full max-w-[240px]">
                <button
                  type="button"
                  onClick={() => portraitInputRef.current?.click()}
                  className="flex-1 py-2 border border-tavern-600/50 hover:border-accent/50 text-gray-400 hover:text-gray-200 rounded-lg text-xs transition-colors"
                >
                  本地上传
                </button>
                <button
                  type="button"
                  className="flex-1 py-2 bg-purple-500/20 hover:bg-purple-500/30 text-purple-400 rounded-lg text-xs transition-colors flex items-center justify-center gap-1"
                  title="AI 生成功能即将推出"
                >
                  <span className="text-sm">&#x2728;</span> AI 生成
                </button>
              </div>
            </div>

            {/* ===== 右侧：信息表单 ===== */}
            <div className="w-[58%] p-6 space-y-4">
              {/* 头像 + 名称 */}
              <div className="flex items-end gap-4">
                <div className="relative shrink-0">
                  <button
                    type="button"
                    onClick={() => avatarInputRef.current?.click()}
                    className="w-16 h-16 rounded-full border-2 border-dashed border-tavern-600/50 flex items-center justify-center overflow-hidden hover:border-accent/50 transition-colors group"
                  >
                    {avatar ? (
                      <img src={avatar} alt="" className="w-full h-full object-cover" />
                    ) : (
                      <span className="text-xs text-gray-500 group-hover:text-gray-400">头像</span>
                    )}
                  </button>
                  <input
                    ref={avatarInputRef}
                    type="file"
                    accept="image/*"
                    onChange={(e) => {
                      const f = e.target.files?.[0]
                      if (f) handleFileToDataUrl(f, setAvatar)
                      e.target.value = ''
                    }}
                    className="hidden"
                  />
                </div>
                <div className="flex-1">
                  <label className="block text-xs font-medium text-gray-400 mb-1.5">角色名称 *</label>
                  <input
                    type="text"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    placeholder="给角色起个名字"
                    className="w-full bg-tavern-800/60 border border-tavern-600/50 rounded-lg px-3 py-2 text-sm text-gray-100 placeholder-gray-500 focus:outline-none focus:border-accent/50"
                  />
                </div>
              </div>

              {/* 口头禅 */}
              <div>
                <label className="block text-xs font-medium text-gray-400 mb-1.5">
                  口头禅 <span className="text-gray-600">- 角色最常说的一句话</span>
                </label>
                <input
                  type="text"
                  value={catchphrase}
                  onChange={(e) => setCatchphrase(e.target.value)}
                  placeholder="例如：真相只有一个！"
                  className="w-full bg-tavern-800/60 border border-tavern-600/50 rounded-lg px-3 py-2 text-sm text-gray-100 placeholder-gray-500 focus:outline-none focus:border-accent/50"
                />
              </div>

              {/* 性格特征 */}
              <div>
                <label className="block text-xs font-medium text-gray-400 mb-1.5">角色性格</label>
                <textarea
                  value={personality}
                  onChange={(e) => setPersonality(e.target.value)}
                  placeholder="描述角色的性格、说话风格、行为习惯..."
                  rows={3}
                  className="w-full bg-tavern-800/60 border border-tavern-600/50 rounded-lg px-3 py-2 text-sm text-gray-100 placeholder-gray-500 resize-none focus:outline-none focus:border-accent/50"
                />
              </div>

              {/* 简介 */}
              <div>
                <label className="block text-xs font-medium text-gray-400 mb-1.5">角色简介 *</label>
                <textarea
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  placeholder="一句话概括这个角色的背景和特点"
                  rows={2}
                  className="w-full bg-tavern-800/60 border border-tavern-600/50 rounded-lg px-3 py-2 text-sm text-gray-100 placeholder-gray-500 resize-none focus:outline-none focus:border-accent/50"
                />
              </div>

              {/* 标签 */}
              <div>
                <label className="block text-xs font-medium text-gray-400 mb-1.5">标签</label>
                <input
                  type="text"
                  value={tags}
                  onChange={(e) => setTags(e.target.value)}
                  placeholder="友好, 智慧, 冒险 (逗号分隔)"
                  className="w-full bg-tavern-800/60 border border-tavern-600/50 rounded-lg px-3 py-2 text-sm text-gray-100 placeholder-gray-500 focus:outline-none focus:border-accent/50"
                />
              </div>

              {/* 高级设置折叠 */}
              <div>
                <button
                  type="button"
                  onClick={() => setShowAdvanced(!showAdvanced)}
                  className="flex items-center gap-1.5 text-xs text-gray-500 hover:text-gray-300 transition-colors"
                >
                  <span className={`transition-transform ${showAdvanced ? 'rotate-90' : ''}`}>&#9654;</span>
                  高级设置
                </button>

                {showAdvanced && (
                  <div className="mt-3 space-y-3">
                    <div>
                      <label className="block text-xs font-medium text-gray-400 mb-1.5">场景设定</label>
                      <textarea
                        value={scenario}
                        onChange={(e) => setScenario(e.target.value)}
                        placeholder="对话发生的场景背景"
                        rows={2}
                        className="w-full bg-tavern-800/60 border border-tavern-600/50 rounded-lg px-3 py-2 text-sm text-gray-100 placeholder-gray-500 resize-none focus:outline-none focus:border-accent/50"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-medium text-gray-400 mb-1.5">开场白</label>
                      <textarea
                        value={firstMessage}
                        onChange={(e) => setFirstMessage(e.target.value)}
                        placeholder="角色的第一句话"
                        rows={2}
                        className="w-full bg-tavern-800/60 border border-tavern-600/50 rounded-lg px-3 py-2 text-sm text-gray-100 placeholder-gray-500 resize-none focus:outline-none focus:border-accent/50"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-medium text-gray-400 mb-1.5">示例对话</label>
                      <textarea
                        value={exampleDialogue}
                        onChange={(e) => setExampleDialogue(e.target.value)}
                        placeholder="User: ...&#10;Character: ..."
                        rows={3}
                        className="w-full bg-tavern-800/60 border border-tavern-600/50 rounded-lg px-3 py-2 text-sm text-gray-100 placeholder-gray-500 resize-none focus:outline-none focus:border-accent/50 font-mono text-xs"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-medium text-amber-400 mb-1.5">角色目标</label>
                      <input
                        type="text"
                        value={goal}
                        onChange={(e) => setGoal(e.target.value)}
                        placeholder="这个角色想要达成什么？如：找到国王的下落"
                        className="w-full bg-tavern-800/60 border border-amber-500/30 rounded-lg px-3 py-2 text-sm text-gray-100 placeholder-gray-500 focus:outline-none focus:border-amber-500/50"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-medium text-rose-400 mb-1.5">角色秘密</label>
                      <input
                        type="text"
                        value={secret}
                        onChange={(e) => setSecret(e.target.value)}
                        placeholder="这个角色隐藏了什么？如：其实我知道国王在哪"
                        className="w-full bg-tavern-800/60 border border-rose-500/30 rounded-lg px-3 py-2 text-sm text-gray-100 placeholder-gray-500 focus:outline-none focus:border-rose-500/50"
                      />
                    </div>
                  </div>
                )}
              </div>
            </div>
          </div>

          {/* 底部按钮 */}
          <div className="flex gap-3 px-6 py-4 border-t border-tavern-700/50 shrink-0">
            <button
              type="button"
              onClick={onCancel}
              className="flex-1 py-2.5 border border-tavern-600/50 rounded-lg text-sm text-gray-400 hover:text-gray-200 hover:border-tavern-500/50 transition-colors"
            >
              取消
            </button>
            <button
              type="submit"
              disabled={!name.trim() || !description.trim()}
              className="flex-1 py-2.5 bg-accent hover:bg-accent-dark text-white rounded-lg text-sm font-medium disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
            >
              {character ? '保存修改' : '创建角色'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
