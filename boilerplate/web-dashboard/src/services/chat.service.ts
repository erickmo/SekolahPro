import { apiClient } from './api.client'

// ─── Types ──────────────────────────────────────────────────────────────────────

export interface ChatChannelAPI {
  id: string
  nama: string
  kategori: string
  deskripsi: string
  tipe: string
  is_archived: boolean
  unread_count: number
  last_message?: {
    isi: string
    pengirim_nama: string
    dikirim_pada: string
  }
}

export interface ChatMessageAPI {
  id: string
  channel_id: string
  pengirim_id: string
  pengirim_nama: string
  pengirim_inisial: string
  isi: string
  tipe: string
  dikirim_pada: string
  is_current_user: boolean
  parent_id?: string
  attachment_url?: string
  mentions?: string[]
  thread_count?: number
}

export interface ChatMemberAPI {
  channel_id: string
  user_id: string
  nama: string
  inisial: string
  peran: string
  joined_at: string
}

export interface ChatRuleAPI {
  id: string
  channel_id: string
  tipe_aturan: string
  user_id?: string
  role?: string
  scope_company_group_id?: string
  scope_company_id?: string
  scope_branch_id?: string
  scope_warehouse_id?: string
  peran: string
  dibuat_oleh: string
  dibuat_pada: string
}

export interface ChatMentionUserAPI {
  id: string
  nama: string
  inisial: string
}

// ─── Request Types ──────────────────────────────────────────────────────────────

export interface CreateChannelRequest {
  nama: string
  kategori: string
  deskripsi?: string
  member_ids?: string[]
  aturan_awal?: {
    tipe_aturan: string
    user_id?: string
    role?: string
    scope_company_group_id?: string
    scope_company_id?: string
    scope_branch_id?: string
    scope_warehouse_id?: string
    permission: string
  }[]
}

export interface SendMessageRequest {
  isi: string
  parent_id?: string
  attachment_url?: string
  mentions?: string[]
}

export interface AddRuleRequest {
  tipe_aturan: string
  user_id?: string
  role?: string
  scope_company_group_id?: string
  scope_company_id?: string
  scope_branch_id?: string
  scope_warehouse_id?: string
  permission: string
}

// ─── Service ────────────────────────────────────────────────────────────────────

const CHAT_BASE = '/api/v1/chat'

export const chatService = {
  // ─── Channels ───────────────────────────────────────────────────────────────

  listChannels: async (): Promise<ChatChannelAPI[]> => {
    const response = await apiClient.get<{ data: ChatChannelAPI[] }>(`${CHAT_BASE}/channels`)
    return response.data ?? []
  },

  createChannel: async (data: CreateChannelRequest): Promise<string> => {
    const response = await apiClient.post<{ data: { channel_id: string } }>(
      `${CHAT_BASE}/channels`,
      data,
    )
    return response.data.channel_id
  },

  // ─── Messages ───────────────────────────────────────────────────────────────

  listMessages: async (
    channelId: string,
    params?: { limit?: number; offset?: number },
  ): Promise<{ items: ChatMessageAPI[]; total: number }> => {
    const query = new URLSearchParams()
    if (params?.limit) query.set('limit', String(params.limit))
    if (params?.offset) query.set('offset', String(params.offset))
    const qs = query.toString()
    const response = await apiClient.get<{
      data: ChatMessageAPI[]
      total: number
    }>(`${CHAT_BASE}/channels/${channelId}/messages${qs ? `?${qs}` : ''}`)
    return { items: response.data ?? [], total: response.total ?? 0 }
  },

  sendMessage: async (
    channelId: string,
    data: SendMessageRequest,
  ): Promise<{ pesan_id: string; dikirim_pada: string }> => {
    const response = await apiClient.post<{
      data: { pesan_id: string; dikirim_pada: string }
    }>(`${CHAT_BASE}/channels/${channelId}/messages`, data)
    return response.data
  },

  markRead: async (channelId: string): Promise<void> => {
    await apiClient.post(`${CHAT_BASE}/channels/${channelId}/read`, {})
  },

  // ─── Thread ─────────────────────────────────────────────────────────────────

  listThread: async (
    messageId: string,
    params?: { limit?: number; offset?: number },
  ): Promise<{ items: ChatMessageAPI[]; total: number }> => {
    const query = new URLSearchParams()
    if (params?.limit) query.set('limit', String(params.limit))
    if (params?.offset) query.set('offset', String(params.offset))
    const qs = query.toString()
    const response = await apiClient.get<{
      data: ChatMessageAPI[]
      total: number
    }>(`${CHAT_BASE}/channels/*/messages/${messageId}/thread${qs ? `?${qs}` : ''}`)
    return { items: response.data ?? [], total: response.total ?? 0 }
  },

  // ─── Members ────────────────────────────────────────────────────────────────

  listMembers: async (
    channelId: string,
    params?: { limit?: number; offset?: number },
  ): Promise<{ items: ChatMemberAPI[]; total: number }> => {
    const query = new URLSearchParams()
    if (params?.limit) query.set('limit', String(params.limit))
    if (params?.offset) query.set('offset', String(params.offset))
    const qs = query.toString()
    const response = await apiClient.get<{
      data: ChatMemberAPI[]
      total: number
    }>(`${CHAT_BASE}/channels/${channelId}/members${qs ? `?${qs}` : ''}`)
    return { items: response.data ?? [], total: response.total ?? 0 }
  },

  removeMember: async (channelId: string, userId: string): Promise<void> => {
    await apiClient.delete(`${CHAT_BASE}/channels/${channelId}/members/${userId}`)
  },

  updateMemberPermission: async (
    channelId: string,
    userId: string,
    permission: string,
  ): Promise<void> => {
    await apiClient.put(
      `${CHAT_BASE}/channels/${channelId}/members/${userId}/permission`,
      { permission },
    )
  },

  syncMembers: async (channelId: string): Promise<number> => {
    const response = await apiClient.post<{
      data: { jumlah_disinkronkan: number }
    }>(`${CHAT_BASE}/channels/${channelId}/members/sync`, {})
    return response.data.jumlah_disinkronkan
  },

  // ─── Access Rules ───────────────────────────────────────────────────────────

  listRules: async (channelId: string): Promise<ChatRuleAPI[]> => {
    const response = await apiClient.get<{ data: ChatRuleAPI[] }>(
      `${CHAT_BASE}/channels/${channelId}/rules`,
    )
    return response.data ?? []
  },

  addRule: async (channelId: string, data: AddRuleRequest): Promise<string> => {
    const response = await apiClient.post<{ data: { aturan_id: string } }>(
      `${CHAT_BASE}/channels/${channelId}/rules`,
      data,
    )
    return response.data.aturan_id
  },

  updateRule: async (
    channelId: string,
    ruleId: string,
    permission: string,
  ): Promise<void> => {
    await apiClient.put(
      `${CHAT_BASE}/channels/${channelId}/rules/${ruleId}`,
      { permission },
    )
  },

  deleteRule: async (channelId: string, ruleId: string): Promise<void> => {
    await apiClient.delete(`${CHAT_BASE}/channels/${channelId}/rules/${ruleId}`)
  },

  // ─── Search ─────────────────────────────────────────────────────────────────

  searchUsers: async (q: string): Promise<ChatMentionUserAPI[]> => {
    const response = await apiClient.get<{ data: ChatMentionUserAPI[] }>(
      `${CHAT_BASE}/users/search?q=${encodeURIComponent(q)}`,
    )
    return response.data ?? []
  },

  // ─── Upload ─────────────────────────────────────────────────────────────────

  uploadImage: async (file: File): Promise<string> => {
    const formData = new FormData()
    formData.append('file', file)
    const authState = (await import('@/stores/auth.store')).useAuthStore.getState()
    const headers: Record<string, string> = {}
    if (authState.token) headers['Authorization'] = `Bearer ${authState.token}`
    if (authState.selectedGroup) headers['X-Company-Group-ID'] = authState.selectedGroup.id
    if (authState.selectedCompany) headers['X-Company-ID'] = authState.selectedCompany.id

    const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080'
    const res = await fetch(`${BASE_URL}${CHAT_BASE}/upload`, {
      method: 'POST',
      headers,
      body: formData,
    })
    if (!res.ok) throw new Error('Upload gagal')
    const json = await res.json()
    return json.data.url as string
  },
}
