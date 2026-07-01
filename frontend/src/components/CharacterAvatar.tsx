export function CharacterAvatar({
  name,
  avatar,
  size = 'md',
}: {
  name: string
  avatar?: string
  size?: 'sm' | 'md' | 'lg'
}) {
  const sizeMap = { sm: 'w-8 h-8 text-xs', md: 'w-10 h-10 text-sm', lg: 'w-16 h-16 text-lg' }

  if (avatar) {
    return (
      <img
        src={avatar}
        alt={name}
        className={`${sizeMap[size]} rounded-full object-cover ring-2 ring-accent/30`}
      />
    )
  }

  const initials = name.slice(0, 2).toUpperCase()
  const colors = [
    'bg-purple-500', 'bg-blue-500', 'bg-emerald-500',
    'bg-amber-500', 'bg-rose-500', 'bg-cyan-500',
  ]
  const colorIdx = name.charCodeAt(0) % colors.length

  return (
    <div
      className={`${sizeMap[size]} rounded-full ${colors[colorIdx]} flex items-center justify-center font-bold text-white ring-2 ring-white/10`}
    >
      {initials}
    </div>
  )
}
