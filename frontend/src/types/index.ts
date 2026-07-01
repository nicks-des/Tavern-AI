export type CharacterScope = 'global' | 'room'

export interface Character {
  id: string
  name: string
  avatar?: string
  portrait?: string
  catchphrase: string
  description: string
  personality: string
  scenario: string
  firstMessage: string
  exampleDialogue: string
  creator?: string
  tags: string[]
  scope: CharacterScope
  roomId?: string
  createdAt: string
}

export interface RoomCharacterOverride {
  scenario?: string
  firstMessage?: string
}

export interface RoomMember {
  characterId: string
  overrides: RoomCharacterOverride
}

export interface Room {
  id: string
  name: string
  description: string
  worldRules: string
  members: RoomMember[]
  createdAt: string
  updatedAt: string
}

export interface Message {
  id: string
  role: 'user' | 'assistant' | 'system'
  content: string
  timestamp: number
  isStreaming?: boolean
}

export interface Session {
  id: string
  characterId: string
  roomId?: string
  title: string
  messages: Message[]
  createdAt: string
  updatedAt: string
}

export interface ChatState {
  sessions: Session[]
  activeSessionId: string | null
  characters: Character[]
  rooms: Room[]
  isStreaming: boolean
  streamContent: string
}
