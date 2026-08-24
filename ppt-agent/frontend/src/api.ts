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
  is_admin?: boolean;
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
	const res = await checkResponse(await apiFetch('/api/tasks', { headers: authHeaders() }));
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

export interface TemplateSelection {
  mode: 'recommended' | 'preset';
  template?: string;
}

export interface TemplateStrategy {
  mode: string;
  template: string;
  theme: string;
  use_visual_assets: boolean;
  visual_hint?: string;
  reason: string;
  page_count?: number;
}

export interface TemplateCandidate {
  name: string;
  display_name: string;
  description: string;
  category: string;
  thumbnail: string;
  slide_count: number;
  tags: string[];
  reason: string;
}

export interface TemplateRecommendation {
  strategy: TemplateStrategy;
  primary_template: TemplateCandidate;
  ranked_templates: TemplateCandidate[];
  theme?: ThemeInfo;
  visual_policy: string;
  component_focus: string[];
  risks?: string[];
}

export async function createTask(query: string, templateSelection?: TemplateSelection): Promise<TaskInfo> {
  const res = await checkResponse(await apiFetch('/api/tasks', {
    method: 'POST',
    headers: authHeaders(),
    body: JSON.stringify({
      query,
      ...(templateSelection ? { template_selection: templateSelection } : {}),
    }),
  }));
  return res.json();
}

export async function fetchTask(id: string): Promise<TaskInfo> {
  const res = await checkResponse(await apiFetch(`/api/tasks/${id}`, { headers: authHeaders() }));
  return res.json();
}

export async function cancelTask(id: string): Promise<TaskInfo> {
	const res = await checkResponse(await apiFetch(`/api/tasks/${id}/cancel`, {
		method: 'POST',
		headers: authHeaders(),
	}));
	return res.json();
}

export async function deleteTask(id: string): Promise<void> {
	const res = await checkResponse(await apiFetch(`/api/tasks/${id}`, {
		method: 'DELETE',
		headers: authHeaders(),
	}));
	await res.json().catch(() => undefined);
}

// ── Template API ──────────────────────────────────────────────────────────────

export interface SlideOutline {
  title: string;
  content_type: string;
  description: string;
  content_plan?: ContentPlan;
}

export interface PlanComponent {
  id?: string;
  type: string;
  title?: string;
  text?: string;
  body?: string;
  description?: string;
  items?: string[];
  local_path?: string;
  asset_query?: string;
  asset_purpose?: string;
  [key: string]: unknown;
}

export interface VisualIntent {
  asset_query?: string;
  asset_purpose?: string;
  local_path?: string;
  [key: string]: unknown;
}

export interface ContentPlan {
  summary?: string;
  slide_intent?: string;
  components?: PlanComponent[];
  visual_intent?: VisualIntent;
  capacity_hint?: Record<string, unknown>;
}

export interface TaskOutline {
  template: string;
  theme: string;
  title: string;
  content_mode?: 'template_scaffold' | 'user_outline' | 'recommended_style';
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
    placeholder?: string;
    options?: Array<string | { value: string; label: string }>;
  }[];
  contract?: {
    capacity?: Record<string, string | number | boolean>;
    required_fields?: string[];
    best_for?: string[];
    avoid_for?: string[];
    overflow_strategy?: string;
    background_policy?: string;
    visual_primitives?: string[];
  };
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
	const res = await checkResponse(await apiFetch('/api/templates'));
  const data = await res.json();
  return data.presets || [];
}

export async function recommendTemplate(query: string): Promise<TemplateRecommendation> {
	const res = await checkResponse(await apiFetch('/api/templates/recommend', {
		method: 'POST',
		headers: authHeaders(),
		body: JSON.stringify({ query }),
	}));
  return res.json();
}

export async function fetchPreset(name: string): Promise<PresetTemplate | null> {
  const res = await apiFetch(`/api/templates/${name}`);
  if (!res.ok) return null;
  return res.json();
}

export async function fetchLayouts(): Promise<AtomicLayout[]> {
	const res = await checkResponse(await apiFetch('/api/templates/layouts'));
  const data = await res.json();
  return data.layouts || [];
}

export async function fetchThemes(): Promise<ThemeInfo[]> {
	const res = await checkResponse(await apiFetch('/api/themes'));
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
	const res = await checkResponse(await apiFetch('/api/ai/expand', {
		method: 'POST',
		headers: authHeaders(),
		body: JSON.stringify({ title, content_type: contentType, description, theme }),
	}));
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

export interface ContinueTaskResult {
  status: 'accepted' | 'queued';
  message: string;
  task_id: string;
  after_event_id?: number;
}

export async function continueTask(taskId: string, message: string): Promise<ContinueTaskResult> {
	const res = await checkResponse(await apiFetch(`/api/tasks/${taskId}/continue`, {
		method: 'POST',
		headers: authHeaders(),
		body: JSON.stringify({ message }),
	}));
	return res.json();
}

export async function fetchConversation(taskId: string): Promise<import('./types').ConversationSession> {
	const res = await checkResponse(await apiFetch(`/api/tasks/${taskId}/conversation`, { headers: authHeaders() }));
  return res.json();
}

export async function fetchRuntimeEvent(taskId: string, eventId: number): Promise<import('./types').RuntimeEvent> {
	const res = await checkResponse(await apiFetch(`/api/tasks/${taskId}/runtime-events/${eventId}`, { headers: authHeaders() }));
  return res.json();
}

// ── User profile API ──────────────────────────────────────────────────────────────

export async function fetchUserProfile(): Promise<import('./types').UserStyleProfile> {
	const res = await checkResponse(await apiFetch('/api/users/me/profile', { headers: authHeaders() }));
  const data = await res.json();
  return data.profile;
}

export async function updateUserProfile(profile: Partial<import('./types').UserStyleProfile>): Promise<void> {
	await checkResponse(await apiFetch('/api/users/me/profile', {
		method: 'PUT',
		headers: authHeaders(),
		body: JSON.stringify(profile),
	}));
}

export async function resetUserProfile(): Promise<void> {
	await checkResponse(await apiFetch('/api/users/me/profile/reset', {
		method: 'POST',
		headers: authHeaders(),
	}));
}

export interface PreferenceSummary {
  preferred_themes: string[];
  preferred_colors: string[];
  content_patterns: string[];
  layout_preferences: string[];
  language_tone: string;
  typical_page_count: number;
  special_notes: string[];
  user_facts?: import('./types').UserFacts;
}

export async function summarizeProfile(): Promise<{ summary: PreferenceSummary; task_count: number; updated_at: string }> {
	const res = await checkResponse(await apiFetch('/api/users/me/profile/summarize', {
		method: 'POST',
		headers: authHeaders(),
	}));
  const data = await res.json();
  return data;
}

// ── Account API key API ───────────────────────────────────────────────────────

export interface UserApiKeyStatus {
  configured: boolean;
  provider: string;
  masked_key: string;
  default_configured: boolean;
  updated_at?: string;
}

export async function fetchUserApiKeyStatus(): Promise<UserApiKeyStatus> {
	const res = await checkResponse(await apiFetch('/api/users/me/api-key', { headers: authHeaders() }));
  return res.json();
}

export async function updateUserApiKey(apiKey: string, provider = 'ark'): Promise<UserApiKeyStatus> {
	const res = await checkResponse(await apiFetch('/api/users/me/api-key', {
		method: 'PUT',
		headers: authHeaders(),
		body: JSON.stringify({ provider, api_key: apiKey }),
	}));
  return res.json();
}

export async function deleteUserApiKey(): Promise<void> {
	await checkResponse(await apiFetch('/api/users/me/api-key', {
		method: 'DELETE',
		headers: authHeaders(),
	}));
}

// ── Recommendations API ──────────────────────────────────────────────────────────────

export interface Recommendation {
  template: string;
  theme: string;
  page_count: number;
  animation: string;
  tips: string[];
}

export async function fetchRecommendations(domain?: string): Promise<Recommendation | null> {
	const url = domain ? `/api/recommendations?domain=${encodeURIComponent(domain)}` : '/api/recommendations';
	const res = await checkResponse(await apiFetch(url, { headers: authHeaders() }));
  const data = await res.json();
  return data.recommendation;
}

// ── Admin API ──────────────────────────────────────────────────────────────────

export interface AdminStats {
  user_count: number;
  task_count: number;
  running_count: number;
}

export interface AdminUser {
  id: number;
  email: string;
  is_admin: boolean;
  created_at: string;
}

export interface AdminTaskRecord {
  id: string;
  user_id: number;
  user_email?: string;
  query: string;
  status: string;
  done_count: number;
  total_count: number;
  duration: string;
  error: string;
  created_at: string;
  updated_at: string;
}

export interface AdminLogAnalysis {
  id: number;
  task_id: string;
  trigger_type: string;
  log_snippet: string;
  analysis: string;
  root_cause: string;
  suggestion: string;
  tokens_used: number;
  model_used: string;
  created_at: string;
}

export interface AdminStyleProfile {
  user_id: number;
  preferred_themes: string;
  preferred_colors: string;
  content_patterns: string;
  language_tone: string;
  typical_page_count: number;
  content_types: string;
  special_notes: string;
  task_count: number;
  updated_at: string;
}

export async function fetchAdminStats(): Promise<AdminStats> {
  const res = await checkResponse(await apiFetch('/api/admin/stats', { headers: authHeaders() }));
  return res.json();
}

export async function fetchAdminUsers(): Promise<AdminUser[]> {
  const res = await checkResponse(await apiFetch('/api/admin/users', { headers: authHeaders() }));
  const data = await res.json();
  return data.users || [];
}

export async function fetchAdminTasks(): Promise<AdminTaskRecord[]> {
  const res = await checkResponse(await apiFetch('/api/admin/tasks', { headers: authHeaders() }));
  const data = await res.json();
  return data.tasks || [];
}

export async function fetchAdminLogAnalyses(): Promise<AdminLogAnalysis[]> {
  const res = await checkResponse(await apiFetch('/api/admin/log-analyses', { headers: authHeaders() }));
  const data = await res.json();
  return data.analyses || [];
}

export async function fetchAdminStyleProfiles(): Promise<AdminStyleProfile[]> {
  const res = await checkResponse(await apiFetch('/api/admin/style-profiles', { headers: authHeaders() }));
  const data = await res.json();
  return data.profiles || [];
}

export async function deleteAdminLogAnalysis(id: number): Promise<void> {
  const res = await apiFetch(`/api/admin/log-analyses/${id}`, {
    method: 'DELETE',
    headers: authHeaders(),
  });
  if (!res.ok) throw new Error('删除失败');
}
