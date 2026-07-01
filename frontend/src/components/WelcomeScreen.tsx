import { useStore } from '../store'

export function WelcomeScreen() {
  const characters = useStore((s) => s.characters)

  return (
    <div className="flex-1 flex items-center justify-center">
      <div className="text-center max-w-md px-6">
        <div className="text-5xl mb-4">&#x1F3FA;</div>
        <h1 className="text-xl font-semibold text-gray-100 mb-3">欢迎来到 Tavern AI</h1>
        <p className="text-sm text-gray-400 mb-8 leading-relaxed">
          选择一个角色开始对话。每个角色都有独特的个性和故事背景，带给你沉浸式的角色扮演体验。
        </p>
        <div className="space-y-2">
          <p className="text-xs text-gray-600 mb-3">
            {characters.length} 个角色已就绪，点击左侧角色即可开始
          </p>
          <div className="flex justify-center gap-2 flex-wrap">
            {characters.map((c) => (
              <span
                key={c.id}
                className="text-xs px-3 py-1.5 rounded-full bg-tavern-800/60 text-gray-400 border border-tavern-700/50"
              >
                {c.name}
              </span>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}
