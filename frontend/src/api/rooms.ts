import { api } from './client'
import type { Room, RoomMember } from '../types'

export const roomApi = {
  list: () => api.get<Room[]>('/rooms'),

  get: (id: string) => api.get<{ room: Room; members: RoomMember[] }>(`/rooms/${id}`),

  create: (name: string, description?: string, worldRules?: string) =>
    api.post<Room>('/rooms', { name, description, worldRules }),

  update: (id: string, data: { name?: string; description?: string; worldRules?: string }) =>
    api.put<Room>(`/rooms/${id}`, data),

  delete: (id: string) => api.del<{ status: string }>(`/rooms/${id}`),

  addMember: (roomId: string, characterId: string, overrides?: string) =>
    api.post<{ status: string }>(`/rooms/${roomId}/members`, { characterId, overrides }),

  removeMember: (roomId: string, characterId: string) =>
    api.del<{ status: string }>(`/rooms/${roomId}/members/${characterId}`),
}
