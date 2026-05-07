// ── Token management ──────────────────────────────────────────────────────

const TOKEN_KEY = 'ppt_agent_token';

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string) {
  localStorage.setItem(TOKEN_KEY, token);
  // Also set cookie for EventSource (can't send custom headers)
  document.cookie = `session_token=${token}; path=/; SameSite=Lax`;
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY);
  document.cookie = 'session_token=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT';
}

export function isLoggedIn(): boolean {
  return !!getToken();
}

function authHeaders(): Record<string, string> {
  const h: Record<string, string> = { 'Content-Type': 'application/json' };
  const t = getToken();
  if (t) h['Authorization'] = `Bearer ${t}`;
  return h;
}

// Global fetch wrapper: auto-redirect to /auth on 401
async function apiFetch(url: string, options?: RequestInit): Promise<Response> {
  const res = await fetch(url, options);
  if (res.status === 401) {
    clearToken();
    window.location.href = '/auth';
    throw new Error('登录已过期，请重新登录');
  }
  return res;
}

// ── Auth API ──────────────────────────────────────────────────────────────

export interface AuthUser {
  id: number;
  email: string;
  token?: string;
  is_new?: boolean;
}

export async function sendCode(email: string): Promise<void> {
  const res = await apiFetch('/api/auth/send-code', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email }),
  });
  if (!res.ok) {
    const e = await res.json();
    throw new Error(e.error || '发送失败');
  }
}

export async function loginWithCode(email: string, code: string): Promise<AuthUser> {
  const res = await apiFetch('/api/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, code }),
  });
  if (!res.ok) {
    const e = await res.json();
    throw new Error(e.error || '登录失败');
  }
  const data = await res.json();
  if (data.token) setToken(data.token);
  return data;
}

export async function loginWithPassword(email: string, password: string): Promise<AuthUser> {
  const res = await apiFetch('/api/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  });
  if (!res.ok) {
    const e = await res.json();
    throw new Error(e.error || '登录失败');
  }
  const data = await res.json();
  if (data.token) setToken(data.token);
  return data;
}

export async function setPassword(password: string): Promise<void> {
  const res = await apiFetch('/api/auth/set-password', {
    method: 'POST',
    headers: authHeaders(),
    body: JSON.stringify({ password }),
  });
  if (!res.ok) {
    const e = await res.json();
    throw new Error(e.error || '设置失败');
  }
}

export async function logout() {
  try {
    await apiFetch('/api/auth/logout', { method: 'POST', headers: authHeaders() });
  } catch { /* ignore */ }
  clearToken();
}

export async function fetchMe(): Promise<AuthUser> {
  const res = await apiFetch('/api/auth/me', { headers: authHeaders() });
  if (!res.ok) throw new Error('未登录');
  return res.json();
}

// ── Types ─────────────────────────────────────────────────────────────────

export interface TaskItem {
  task_id: string;
  page_index: number;
  title: string;
  content_type: string;
  output_file: string;
  status: string;
  qa_report?: string;
  fix_attempts?: number;
}

export interface TaskInfo {
  id: string;
  user_id?: number;
  query: string;
  status: 'running' | 'completed' | 'failed';
  work_dir: string;
  created_at: string;
  done_count: number;
  total_count: number;
  duration?: string;
  error?: string;
  files?: string[];
}

export interface SSEEvent {
  type: string;
  content?: string;
  tool_name?: string;
  tool_args?: string;
  error?: string;
  tasks?: TaskItem[];
  done?: number;
  total?: number;
  files?: string[];
  message?: string;
  duration?: string;
  status?: string;
}

export const STATUS_LABELS: Record<string, string> = {
  pending: '待生成',
  generating: '生成中',
  done: '已完成',
  qa_done: '已质检',
  fixed: '已修复',
  failed: '失败',
};

export const STATUS_COLORS: Record<string, string> = {
  pending: '#94a3b8',
  generating: '#3b82f6',
  done: '#22c55e',
  qa_done: '#14b8a6',
  fixed: '#a855f7',
  failed: '#ef4444',
};

// ── Task API ──────────────────────────────────────────────────────────────

export async function fetchTasks(): Promise<TaskInfo[]> {
  const res = await apiFetch('/api/tasks', { headers: authHeaders() });
  return res.json();
}

async function checkResponse(res: Response): Promise<Response> {
  if (!res.ok) {
    let msg = `请求失败 (${res.status})`;
    try { const e = await res.json(); if (e.error) msg = e.error; } catch {}
    throw new Error(msg);
  }
  return res;
}

export async function createTask(query: string): Promise<TaskInfo> {
  const res = await checkResponse(await apiFetch('/api/tasks', {
    method: 'POST',
    headers: authHeaders(),
    body: JSON.stringify({ query }),
  }));
  return res.json();
}

export async function fetchTask(id: string): Promise<TaskInfo> {
  const res = await checkResponse(await apiFetch(`/api/tasks/${id}`, { headers: authHeaders() }));
  return res.json();
}

export async function cancelTask(id: string): Promise<TaskInfo> {
  const res = await apiFetch(`/api/tasks/${id}/cancel`, {
    method: 'POST',
    headers: authHeaders(),
  });
  return res.json();
}

export async function deleteTask(id: string): Promise<void> {
  const res = await apiFetch(`/api/tasks/${id}`, {
    method: 'DELETE',
    headers: authHeaders(),
  });
  if (!res.ok) throw new Error('删除失败');
}
