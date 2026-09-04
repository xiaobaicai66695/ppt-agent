import type { AtomicLayout, AuthUser, ConversationSession, RuntimeEvent, TaskInfo, TaskOutline } from './types'

export type MessageMode = 'chat' | 'pptagent'

export interface MessageRoute {
  task_id: string
  intent: string
  mode: string
  action?: string
  needs_confirmation?: boolean
  reply?: string
  after_event_id?: number
}

const tokenKey = 'ppt_agent_token'
export const getToken = () => localStorage.getItem(tokenKey)
export function setToken(token: string) {
  localStorage.setItem(tokenKey, token)
  document.cookie = `session_token=${token}; path=/; max-age=604800; SameSite=Lax`
}
export function clearToken() {
  localStorage.removeItem(tokenKey)
  document.cookie = 'session_token=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT'
}
export const isLoggedIn = () => Boolean(getToken())
const headers = () => ({ 'Content-Type': 'application/json', ...(getToken() ? { Authorization: `Bearer ${getToken()}` } : {}) })

async function request<T>(url: string, options: RequestInit = {}): Promise<T> {
  const response = await fetch(url, { ...options, headers: { ...headers(), ...options.headers } })
  if (response.status === 401) {
    clearToken()
    throw new Error('登录已过期，请重新登录')
  }
  if (!response.ok) {
    const body = await response.json().catch(() => ({})) as { error?: string }
    throw new Error(body.error || `请求失败 (${response.status})`)
  }
  return response.json() as Promise<T>
}

export async function loginAsGuest() { const user = await request<AuthUser>('/api/auth/guest', { method: 'POST' }); if (user.token) setToken(user.token); return user }
export async function sendCode(email: string) { await request('/api/auth/send-code', { method: 'POST', body: JSON.stringify({ email }) }) }
export async function login(email: string, secret: string, byPassword = false) {
  const user = await request<AuthUser>('/api/auth/login', { method: 'POST', body: JSON.stringify(byPassword ? { email, password: secret } : { email, code: secret }) })
  if (user.token) setToken(user.token)
  return user
}
export async function register(email: string, code: string, password: string) {
  const user = await request<AuthUser>('/api/auth/register', { method: 'POST', body: JSON.stringify({ email, code, password }) })
  if (user.token) setToken(user.token)
  return user
}
export const fetchMe = () => request<AuthUser>('/api/auth/me')
export async function logout() { await request('/api/auth/logout', { method: 'POST' }).catch(() => undefined); clearToken() }
export const fetchTasks = () => request<TaskInfo[]>('/api/tasks')
export const fetchTask = (id: string) => request<TaskInfo>(`/api/tasks/${id}`)
export const deleteTask = (id: string) => request(`/api/tasks/${id}`, { method: 'DELETE' })
export const cancelTask = (id: string) => request<TaskInfo>(`/api/tasks/${id}/cancel`, { method: 'POST' })
export const startTask = (id: string) => request<TaskInfo>(`/api/tasks/${id}/start`, { method: 'POST' })
export const fetchConversation = (id: string) => request<ConversationSession>(`/api/tasks/${id}/conversation`)
export const fetchRuntimeEvent = (id: string, eventId: number) => request<RuntimeEvent>(`/api/tasks/${id}/runtime-events/${eventId}`)
export const routeMessage = (message: string, selectedTaskId = '', manualMode: MessageMode = 'chat', webSearch = false, imageSearch = false) => request<MessageRoute>('/api/messages', { method: 'POST', body: JSON.stringify({ message, selected_task_id: selectedTaskId, manual_mode: manualMode, web_search: webSearch, image_search: imageSearch }) })
export const continueTask = (id: string, message: string) => request<{ task_id: string; after_event_id?: number }>(`/api/tasks/${id}/continue`, { method: 'POST', body: JSON.stringify({ message }) })
export const fetchLayouts = async () => (await request<{ layouts?: AtomicLayout[] }>('/api/templates/layouts')).layouts || []
export const createTaskWithOutline = (query: string, outline: TaskOutline) => request<TaskInfo>('/api/tasks', { method: 'POST', body: JSON.stringify({ query, outline }) })
export const taskDownloadUrl = (id: string, name: string) => `/api/tasks/${id}/files/${encodeURIComponent(name)}`
export const taskThumbUrl = (id: string, name: string) => `/api/tasks/${id}/thumb/${encodeURIComponent(name)}`
export const saveTaskFeedback = (taskId: string, rating: number, suggestion = '') => request<TaskInfo>(`/api/tasks/${taskId}/feedback`, { method: 'PUT', body: JSON.stringify({ rating, suggestion }) })
export interface UserApiKeyStatus { configured: boolean; provider: string; default_provider?: string; masked_key: string; default_configured: boolean; updated_at?: string }
export const fetchUserApiKeyStatus = () => request<UserApiKeyStatus>('/api/users/me/api-key')
export const updateUserApiKey = (apiKey: string, provider = 'ark') => request<UserApiKeyStatus>('/api/users/me/api-key', { method: 'PUT', body: JSON.stringify({ provider, api_key: apiKey }) })
export const deleteUserApiKey = () => request('/api/users/me/api-key', { method: 'DELETE' })
export interface AdminStats {
  user_count: number; task_count: number; running_count: number
  registered_user_count: number; non_root_registered_user_count: number
  ppt_active_user_count: number; custom_api_key_user_count: number
  ppt_generation_count: number; non_root_ppt_generation_count: number
  feedback_count: number; feedback_suggestion_count: number
}
export interface AdminUser { id: number; email: string; is_admin: boolean; created_at: string; ppt_generation_count: number; custom_api_key_configured: boolean }
export interface AdminTask {
  id: string; user_id: number; user_email?: string; query: string; status: string; done_count: number; total_count: number
  duration?: string; generation_started_at?: string; generation_finished_at?: string; generation_duration_ms: number; fixer_run_count: number
  error?: string; created_at: string; updated_at?: string
}
export interface AdminFeedback { task_id: string; user_id: number; user_email?: string; task_query?: string; rating: number; suggestion?: string; created_at?: string; updated_at?: string }

export const fetchAdminStats = () => request<AdminStats>('/api/admin/stats')
export const fetchAdminUsers = () => request<{ users?: AdminUser[] }>('/api/admin/users').then(data => data.users || [])
export const fetchAdminTasks = () => request<{ tasks?: AdminTask[] }>('/api/admin/tasks').then(data => data.tasks || [])
export const fetchAdminFeedback = () => request<{ feedback?: AdminFeedback[] }>('/api/admin/feedback?limit=100').then(data => data.feedback || [])
