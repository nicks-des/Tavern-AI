import { api } from './client'
import type { Character } from '../types'

export interface CreateCharacterInput {
  name: string
  avatar?: string
  portrait?: string
  catchphrase?: string
  description: string
  personality?: string
  scenario?: string
  firstMessage?: string
  exampleDialogue?: string
  tags?: string[]
}

export const characterApi = {
  list: () => api.get<Character[]>('/characters'),

  get: (id: string) => api.get<Character>(`/characters/${id}`),

  create: (data: CreateCharacterInput) =>
    api.post<Character>('/characters', {
      ...data,
      tags: data.tags ?? [],
      scope: 'global',
    }),

  update: (id: string, data: Partial<CreateCharacterInput>) =>
    api.put<Character>(`/characters/${id}`, data),

  delete: (id: string) => api.del<{ status: string }>(`/characters/${id}`),
}
