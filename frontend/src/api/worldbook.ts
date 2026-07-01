import { api } from './client'
import type { WorldBookEntry } from '../types'

export const worldbookApi = {
  list: (characterId: string) =>
    api.get<WorldBookEntry[]>(`/characters/${characterId}/worldbook`),

  create: (characterId: string, keywords: string, content: string, enabled = true) =>
    api.post<WorldBookEntry>(`/characters/${characterId}/worldbook`, { keywords, content, enabled: true }),

  update: (id: string, data: { keywords?: string; content?: string; enabled?: boolean }) =>
    api.put<WorldBookEntry>(`/worldbook/${id}`, data),

  delete: (id: string) =>
    api.del<{ status: string }>(`/worldbook/${id}`),
}
