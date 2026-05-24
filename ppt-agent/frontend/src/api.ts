import type { TaskInfo } from './types';

// ── Token management ──────────────────────────────────────────────────────

const TOKEN_KEY = 'ppt_agent_token';

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string) {
  localStorage.setItem(TOKEN_KEY, token);
  // Also set cookie for EventSource (can't send custom headers)
  document.cookie = `session_token=${token}; path=/; max-age=604800; SameSite=Lax`;
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

// Global fetch wrapper: handle 401 without disrupting active workflows
async function apiFetch(url: string, options?: RequestInit): Promise<Response> {
  const res = await fetch(url, options);
  if (res.status === 401) {
    clearToken();
    // Only auto-redirect if user is on a page that requires auth (not already on /auth)
    if (!window.location.pathname.startsWith('/auth')) {
      window.location.href = '/auth';
    }
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

// ── Template API ──────────────────────────────────────────────────────────────

export interface SlideOutline {
  title: string;
  content_type: string;
  description: string;
  content_plan?: ContentPlan;
}

export interface ContentElement {
  type: string;
  items?: string[];
  text?: string;
  title?: string;
  description?: string;
  layout_hint?: string;
}

export interface ContentPlan {
  summary?: string;
  elements?: ContentElement[];
}

export interface TaskOutline {
  template: string;
  theme: string;
  title: string;
  slides: SlideOutline[];
}

export interface PresetTemplate {
  name: string;
  display_name: string;
  type: string;
  description: string;
  category: string;
  default_palette: string;
  tags: string[];
  thumbnail: string;
  slide_count: number;
  default_slides: {
    title: string;
    content_type: string;
    description: string;
  }[];
}

export interface AtomicLayout {
  name: string;
  display_name: string;
  type: string;
  description: string;
  allowed_palettes: string[];
  fields: {
    name: string;
    label: string;
    type: string;
    required: boolean;
  }[];
}

export interface ThemeInfo {
  name: string;
  display_name: string;
  primary: string;
  secondary: string;
  accent: string;
  background: string;
  tags: string[];
}

export async function fetchPresets(): Promise<PresetTemplate[]> {
  const res = await apiFetch('/api/templates');
  const data = await res.json();
  return data.presets || [];
}

export async function fetchPreset(name: string): Promise<PresetTemplate | null> {
  const res = await apiFetch(`/api/templates/${name}`);
  if (!res.ok) return null;
  return res.json();
}

export async function fetchLayouts(): Promise<AtomicLayout[]> {
  const res = await apiFetch('/api/templates/layouts');
  const data = await res.json();
  return data.layouts || [];
}

export async function fetchThemes(): Promise<ThemeInfo[]> {
  const res = await apiFetch('/api/themes');
  const data = await res.json();
  return data.themes || [];
}

export async function createTaskWithOutline(query: string, outline: TaskOutline): Promise<TaskInfo> {
  const res = await checkResponse(await apiFetch('/api/tasks', {
    method: 'POST',
    headers: authHeaders(),
    body: JSON.stringify({ query, outline }),
  }));
  return res.json();
}

export async function expandWithAI(title: string, contentType: string, description: string, theme: string): Promise<string> {
  const res = await apiFetch('/api/ai/expand', {
    method: 'POST',
    headers: authHeaders(),
    body: JSON.stringify({ title, content_type: contentType, description, theme }),
  });
  const data = await res.json();
  return data.description || '';
}

export async function generateOutlineWithAI(query: string, outline: TaskOutline): Promise<SlideOutline[]> {
  try {
    const res = await apiFetch('/api/ai/generate-outline', {
      method: 'POST',
      headers: authHeaders(),
      body: JSON.stringify({ query, outline }),
    });
    if (!res.ok) {
      const text = await res.text();
      let errMsg = '生成大纲失败';
      try { const j = JSON.parse(text); errMsg = j.error || errMsg; } catch { if (text) errMsg = text; }
      throw new Error(errMsg);
    }
    const text = await res.text();
    if (!text.trim()) throw new Error('服务器返回为空，请重试');
    const body = JSON.parse(text);
    if (!body || !Array.isArray(body.slides)) throw new Error('服务器返回格式错误: ' + text.slice(0, 100));
    return body.slides;
  } catch (e) {
    if (e instanceof SyntaxError) {
      throw new Error('服务器返回格式错误，请重试');
    }
    throw e;
  }
}

// ── Continue / Session API ──────────────────────────────────────────────────────

export async function continueTask(taskId: string, message: string): Promise<Response> {
  return apiFetch(`/api/tasks/${taskId}/continue`, {
    method: 'POST',
    headers: authHeaders(),
    body: JSON.stringify({ message }),
  });
}

export async function fetchConversation(taskId: string): Promise<import('./types').ConversationSession> {
  const res = await apiFetch(`/api/tasks/${taskId}/conversation`, { headers: authHeaders() });
  return res.json();
}

// ── User profile API ──────────────────────────────────────────────────────────────

export async function fetchUserProfile(): Promise<import('./types').UserStyleProfile> {
  const res = await apiFetch('/api/users/me/profile', { headers: authHeaders() });
  const data = await res.json();
  return data.profile;
}

export async function updateUserProfile(profile: Partial<import('./types').UserStyleProfile>): Promise<void> {
  await apiFetch('/api/users/me/profile', {
    method: 'PUT',
    headers: authHeaders(),
    body: JSON.stringify(profile),
  });
}

export async function resetUserProfile(): Promise<void> {
  await apiFetch('/api/users/me/profile/reset', {
    method: 'POST',
    headers: authHeaders(),
  });
}
