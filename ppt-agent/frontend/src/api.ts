import type { AtomicLayout, AuthUser, ConversationSession, RuntimeEvent, TaskInfo, TaskOutline } from './types'

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
export const fetchMe = () => request<AuthUser>('/api/auth/me')
export async function logout() { await request('/api/auth/logout', { method: 'POST' }).catch(() => undefined); clearToken() }
export const fetchTasks = () => request<TaskInfo[]>('/api/tasks')
export const fetchTask = (id: string) => request<TaskInfo>(`/api/tasks/${id}`)
export const deleteTask = (id: string) => request(`/api/tasks/${id}`, { method: 'DELETE' })
export const cancelTask = (id: string) => request<TaskInfo>(`/api/tasks/${id}/cancel`, { method: 'POST' })
export const fetchConversation = (id: string) => request<ConversationSession>(`/api/tasks/${id}/conversation`)
export const fetchRuntimeEvent = (id: string, eventId: number) => request<RuntimeEvent>(`/api/tasks/${id}/runtime-events/${eventId}`)
export const routeMessage = (message: string, selectedTaskId = '', manualMode: 'chat' | 'pptagent' = 'chat', webSearch = false, imageSearch = false) => request<{ task_id: string; intent: string; mode: string; reply?: string; after_event_id?: number }>('/api/messages', { method: 'POST', body: JSON.stringify({ message, selected_task_id: selectedTaskId, manual_mode: manualMode, web_search: webSearch, image_search: imageSearch }) })
export const continueTask = (id: string, message: string) => request<{ task_id: string; after_event_id?: number }>(`/api/tasks/${id}/continue`, { method: 'POST', body: JSON.stringify({ message }) })
export const fetchLayouts = async () => (await request<{ layouts?: AtomicLayout[] }>('/api/templates/layouts')).layouts || []
export const createTaskWithOutline = (query: string, outline: TaskOutline) => request<TaskInfo>('/api/tasks', { method: 'POST', body: JSON.stringify({ query, outline }) })
export const taskDownloadUrl = (id: string, name: string) => `/api/tasks/${id}/files/${encodeURIComponent(name)}`
export const taskThumbUrl = (id: string, name: string) => `/api/tasks/${id}/thumb/${encodeURIComponent(name)}`
export const fetchAdminStats = () => request<{ user_count: number; task_count: number; running_count: number }>('/api/admin/stats')
export const fetchAdminUsers = () => request<{ users?: Array<{ id: number; email: string; is_admin: boolean; created_at: string }> }>('/api/admin/users').then(data => data.users || [])
export const fetchAdminTasks = () => request<{ tasks?: Array<{ id: string; user_email?: string; query: string; status: string; done_count: number; total_count: number; created_at: string }> }>('/api/admin/tasks').then(data => data.tasks || [])
