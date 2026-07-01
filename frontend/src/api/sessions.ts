import { api } from './client'
import type { Session, Message } from '../types'

export const sessionApi = {
  list: () => api.get<Session[]>('/sessions'),

  get: (id: string) =>
    api.get<{ session: Session; messages: Message[] }>(`/sessions/${id}`),

  create: (id: string, characterId: string, title?: string) =>
    api.post<Session>('/sessions', { id, characterId, title }),

  delete: (id: string) => api.del<{ status: string }>(`/sessions/${id}`),

  sendMessage: (sessionId: string, role: string, content: string) =>
    api.post<Message>(`/sessions/${sessionId}/messages`, { role, content }),

  listMessages: (sessionId: string) =>
    api.get<Message[]>(`/sessions/${sessionId}/messages`),

  chatStream: (sessionId: string, message: string) =>
    fetch(`/api/sessions/${sessionId}/chat`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ message }),
    }),
}
