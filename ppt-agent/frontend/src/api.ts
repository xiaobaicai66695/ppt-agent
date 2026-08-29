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

export async function createTask(query: string): Promise<TaskInfo> {
  const res = await checkResponse(await apiFetch('/api/tasks', {
    method: 'POST',
    headers: authHeaders(),
	body: JSON.stringify({ query }),
  }));
  return res.json();
}

// Promote an existing workbench conversation into a PPT generation task while
// retaining its task ID and persisted conversation history.
export async function startTask(id: string): Promise<TaskInfo> {
  const res = await checkResponse(await apiFetch(`/api/tasks/${id}/start`, {
    method: 'POST',
    headers: authHeaders(),
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
  content_plan?: ContentPlan;
}

export interface PlanComponent {
  id?: string;
  type: string;
  title?: string;
  text?: string;
  body?: string;
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
}

export interface TaskOutline {
  title: string;
  content_mode?: 'user_outline';
  slides: SlideOutline[];
}

export interface AtomicLayout {
  name: string;
  display_name: string;
  type: string;
  description: string;
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

export async function fetchLayouts(): Promise<AtomicLayout[]> {
	const res = await checkResponse(await apiFetch('/api/templates/layouts'));
  const data = await res.json();
  return data.layouts || [];
}

export async function createTaskWithOutline(query: string, outline: TaskOutline): Promise<TaskInfo> {
  const res = await checkResponse(await apiFetch('/api/tasks', {
    method: 'POST',
    headers: authHeaders(),
    body: JSON.stringify({ query, outline }),
  }));
  return res.json();
}

// ── Unified message / intent API ─────────────────────────────────────────────

export type MessageIntent = 'chat' | 'create' | 'plan' | 'fix';
export type AgentMode = 'chat' | 'pptagent';
export type MessageAction = 'reply' | 'prepare_create' | 'save_plan' | 'update_task' | 'ask_clarification';

export interface MessageRouteResult {
  intent: MessageIntent;
  mode: AgentMode;
  confidence: number;
  needs_confirmation: boolean;
  normalized_request: string;
  task_id: string;
  draft_id?: string;
  missing_fields: string[];
  action: MessageAction;
  reason?: string;
  reply?: string;
  task_candidates?: TaskCandidate[];
}

export interface TaskCandidate {
  id: string;
  title: string;
  status: string;
  created_at: string;
}

export interface PlanDraft {
  id: string;
  user_id: number;
  conversation_id: string;
  source_message_id: string;
  query: string;
  normalized_request: string;
  draft_content: string;
  status: string;
  created_at: string;
  updated_at: string;
}

export async function routeMessage(message: string, selectedTaskId = '', manualMode: AgentMode = 'chat'): Promise<MessageRouteResult> {
	const res = await checkResponse(await apiFetch('/api/messages', {
		method: 'POST',
		headers: authHeaders(),
		body: JSON.stringify({ message, selected_task_id: selectedTaskId, manual_mode: manualMode }),
	}));
  return res.json();
}

export async function fetchPlanDrafts(): Promise<PlanDraft[]> {
	const res = await checkResponse(await apiFetch('/api/plan-drafts', { headers: authHeaders() }));
  const data = await res.json();
  return data.drafts || [];
}

export async function fetchPlanDraft(id: string): Promise<PlanDraft> {
	const res = await checkResponse(await apiFetch(`/api/plan-drafts/${encodeURIComponent(id)}`, { headers: authHeaders() }));
  return res.json();
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

// ── Account API key API ───────────────────────────────────────────────────────

export interface UserApiKeyStatus {
  configured: boolean;
  provider: string;
  default_provider?: string;
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

export async function deleteAdminLogAnalysis(id: number): Promise<void> {
  const res = await apiFetch(`/api/admin/log-analyses/${id}`, {
    method: 'DELETE',
    headers: authHeaders(),
  });
  if (!res.ok) throw new Error('删除失败');
}
