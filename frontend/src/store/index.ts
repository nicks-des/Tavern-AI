import { create } from 'zustand'
import { persist, createJSONStorage } from 'zustand/middleware'
import type { Character, Message, Session, Room } from '../types'
import { characterApi } from '../api/characters'
import { sessionApi } from '../api/sessions'
import { mockCharacters } from '../mock/characters'

interface Store {
  characters: Character[]
  sessions: Session[]
  rooms: Room[]
  activeSessionId: string | null
  activeRoomId: string | null
  isStreaming: boolean
  streamContent: string
  currentView: 'chat' | 'manage'
  backendReady: boolean

  setCurrentView: (v: 'chat' | 'manage') => void
  setActiveSession: (id: string) => void
  setActiveRoom: (id: string | null) => void
  createSession: (characterId: string) => Session
  addMessage: (sessionId: string, msg: Message) => void
  updateLastMessage: (sessionId: string, content: string) => void
  setStreaming: (v: boolean) => void
  setStreamContent: (v: string) => void
  loadCharacters: () => Promise<void>
  addCharacter: (c: Character) => void
  updateCharacter: (id: string, c: Partial<Character>) => void
  deleteCharacter: (id: string) => void
  deleteSession: (id: string) => void
  getGlobalCharacters: () => Character[]
  getRoomCharacters: (roomId: string) => Character[]
  getActiveCharacter: () => Character | undefined
  getActiveSession: () => Session | undefined
}

export const useStore = create<Store>()(
  persist(
    (set, get) => ({
      characters: [],
      sessions: [],
      rooms: [],
      activeSessionId: null,
      activeRoomId: null,
      isStreaming: false,
      streamContent: '',
      currentView: 'chat',
      backendReady: false,

      setCurrentView: (v) => set({ currentView: v }),
      setActiveSession: (id) => set({ activeSessionId: id }),
      setActiveRoom: (id) => set({ activeRoomId: id, activeSessionId: null }),

      createSession: (characterId) => {
        const char = get().characters.find((c) => c.id === characterId)
        const sid = Date.now().toString()
        const session: Session = {
          id: sid,
          characterId,
          title: char ? `${char.name}` : '新对话',
          messages: [],
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
        }
        set((s) => ({
          sessions: [...s.sessions, session],
          activeSessionId: sid,
        }))

        sessionApi.create(sid, characterId, session.title).catch(() => {})
        return session
      },

      addMessage: (sessionId, msg) =>
        set((s) => ({
          sessions: s.sessions.map((ss) =>
            ss.id === sessionId
              ? { ...ss, messages: [...ss.messages, msg], updatedAt: new Date().toISOString() }
              : ss
          ),
        })),

      updateLastMessage: (sessionId, content) =>
        set((s) => ({
          sessions: s.sessions.map((ss) => {
            if (ss.id !== sessionId) return ss
            const msgs = [...ss.messages]
            if (msgs.length > 0) {
              msgs[msgs.length - 1] = { ...msgs[msgs.length - 1], content, isStreaming: false }
            }
            return { ...ss, messages: msgs, updatedAt: new Date().toISOString() }
          }),
        })),

      setStreaming: (v) => set({ isStreaming: v }),
      setStreamContent: (v) => set({ streamContent: v }),

      loadCharacters: async () => {
        try {
          const chars = await characterApi.list()
          if (chars && chars.length > 0) {
            set({ characters: chars, backendReady: true })
            return
          }
        } catch {
          // Backend unavailable, use mock data
        }
        set({ characters: mockCharacters, backendReady: false })
      },

      addCharacter: (c) =>
        set((s) => ({ characters: [...s.characters, c] })),

      updateCharacter: (id, partial) =>
        set((s) => ({
          characters: s.characters.map((c) => (c.id === id ? { ...c, ...partial } : c)),
        })),

      deleteCharacter: (id) =>
        set((s) => ({ characters: s.characters.filter((c) => c.id !== id) })),

      deleteSession: (id) =>
        set((s) => ({
          sessions: s.sessions.filter((ss) => ss.id !== id),
          activeSessionId: s.activeSessionId === id ? null : s.activeSessionId,
        })),

      getActiveCharacter: () => {
        const s = get().activeSessionId
        if (!s) return undefined
        const session = get().sessions.find((ss) => ss.id === s)
        if (!session) return undefined
        return get().characters.find((c) => c.id === session.characterId)
      },

      getGlobalCharacters: () => {
        return get().characters.filter((c) => c.scope === 'global')
      },

      getRoomCharacters: (roomId) => {
        return get().characters.filter((c) => c.scope === 'room' && c.roomId === roomId)
      },

      getActiveSession: () => {
        const s = get().activeSessionId
        return s ? get().sessions.find((ss) => ss.id === s) : undefined
      },
    }),
    {
      name: 'tavern-ai-store',
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({
        characters: state.characters,
        sessions: state.sessions,
        rooms: state.rooms,
      }),
    }
  )
)
