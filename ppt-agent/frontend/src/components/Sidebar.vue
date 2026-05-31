<script setup lang="ts">
import { ref, computed } from 'vue';
import { useRouter } from 'vue-router';
import type { TaskInfo } from '../types';
import { isLoggedIn, summarizeProfile, updateUserProfile, type PreferenceSummary } from '../api';

const props = defineProps<{
  user: { id: number; email: string; is_admin?: boolean } | null;
  tasks: TaskInfo[];
  selectedId: string | null;
  hasRunningTask: boolean;
  creating: boolean;
  error?: string;
}>();

const emit = defineEmits<{
  logout: [];
  selectTask: [id: string];
  createTask: [query: string];
  deleteTask: [id: string];
  compose: [];
  newSession: [];
}>();

const router = useRouter();
const query = ref('');

const hasActiveTask = computed(() =>
  props.tasks.some(t => t.status === 'running')
);

function fmtTime(iso: string): string {
  if (!iso) return '';
  const d = new Date(iso);
  const today = new Date();
  const isToday = d.toDateString() === today.toDateString();
  if (isToday) {
    return d.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });
  }
  return `${d.getMonth() + 1}/${d.getDate()} ` + d.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });
}

function fmtTokens(n: number): string {
  if (!n || n <= 0) return '';
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M';
  if (n >= 1_000) return (n / 1000).toFixed(1) + 'K';
  return String(n);
}

function handleCreate() {
  const q = query.value.trim();
  if (!q || props.creating || props.hasRunningTask) return;
  if (!isLoggedIn()) {
    router.push('/auth');
    return;
  }
  emit('createTask', q);
  query.value = '';
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault();
    handleCreate();
  }
}

function onLogout() { emit('logout'); router.push('/'); }

const taskCount = computed(() => props.tasks.length);

// ── Preference editor ─────────────────────────────────────────────────
const showPrefs = ref(false);
const prefsLoading = ref(false);
const prefsSummary = ref<PreferenceSummary | null>(null);
const prefsEditing = ref<PreferenceSummary | null>(null);
const prefsSaved = ref(false);

async function openPrefs() {
  showPrefs.value = true;
  prefsLoading.value = true;
  prefsSaved.value = false;
  try {
    const data = await summarizeProfile();
    prefsSummary.value = data.summary;
    prefsEditing.value = JSON.parse(JSON.stringify(data.summary)); // deep clone
  } catch {
    prefsSummary.value = null;
    prefsEditing.value = null;
  } finally {
    prefsLoading.value = false;
  }
}

async function savePrefs() {
  if (!prefsEditing.value) return;
  try {
    await updateUserProfile(prefsEditing.value);
    prefsSaved.value = true;
    setTimeout(() => { prefsSaved.value = false; }, 2000);
  } catch { /* ignore */ }
}

function closePrefs() {
  showPrefs.value = false;
  prefsSummary.value = null;
  prefsEditing.value = null;
}
</script>

<template>
  <aside class="sidebar">
    <div class="sidebar-header" @click="router.push('/')">
      <div class="logo-icon">
        <svg viewBox="0 0 40 40" fill="none">
          <rect x="5" y="7" width="30" height="22" rx="4" fill="var(--accent-soft)" stroke="var(--accent-border)" stroke-width="1.5"/>
          <rect x="5" y="7" width="30" height="5" rx="4" fill="var(--accent)"/>
          <circle cx="12" cy="15" r="2" fill="var(--accent)"/>
          <circle cx="19" cy="15" r="2" fill="var(--accent)" opacity="0.6"/>
          <circle cx="26" cy="15" r="2" fill="var(--accent)" opacity="0.4"/>
          <rect x="10" y="20" width="16" height="1.5" rx="0.75" fill="var(--accent)" opacity="0.4"/>
          <rect x="10" y="23" width="20" height="1.5" rx="0.75" fill="var(--accent)" opacity="0.25"/>
        </svg>
      </div>
      <div>
        <h1 class="sidebar-logo">PPT Agent</h1>
        <span class="sidebar-sub">AI 驱动的幻灯片生成</span>
      </div>
    </div>

    <!-- User bar -->
    <div class="user-bar" role="region" aria-label="用户信息">
      <template v-if="user">
        <span class="user-avatar" :class="{ admin: props.user?.is_admin }" aria-hidden="true">{{ user.email[0].toUpperCase() }}</span>
        <span class="user-name" :title="user.email">{{ user.email.split('@')[0] }}</span>
        <button class="prefs-btn" @click="openPrefs" title="偏好设置">偏好</button>
        <button class="logout-btn" @click="onLogout" title="退出登录">退出</button>
      </template>
      <template v-else>
        <span class="user-avatar guest" aria-hidden="true">?</span>
        <span class="user-name dim">未登录</span>
        <button class="logout-btn" @click="router.push('/auth')">登录</button>
      </template>
    </div>

    <!-- Admin shortcut -->
    <div class="admin-shortcut">
      <button class="admin-shortcut-btn" @click="router.push('/admin')">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="8" r="4"/><path d="M4 20c0-4 3.6-7 8-7s8 3 8 7"/><path d="M18 12a2 2 0 0 0-2-2h-1"/></svg>
        管理后台
      </button>
    </div>

    <!-- Create form -->
    <div class="create-form">
      <label class="create-label" for="create-input">新建 PPT 任务</label>
      <textarea
        id="create-input"
        class="create-input"
        placeholder="描述你的 PPT 需求，例如：做一个关于新能源汽车行业分析报告..."
        v-model="query"
        @keydown="handleKeydown"
        rows="3"
        :disabled="creating || hasActiveTask"
        aria-label="PPT 需求描述"
      ></textarea>
      <button
        class="create-btn"
        :class="{ loading: creating, busy: hasActiveTask }"
        :disabled="creating || hasActiveTask"
        @click="handleCreate"
        :aria-disabled="creating || hasActiveTask"
      >
        <span v-if="creating" class="btn-spinner" aria-hidden="true"></span>
        <span>{{ creating ? '创建中...' : hasActiveTask ? '任务执行中...' : '✦ 生成 PPT' }}</span>
      </button>
      <p v-if="hasActiveTask && !creating" class="busy-hint" role="status">
        有任务正在执行中，请稍候
      </p>
      <p v-if="error" class="error-hint" role="alert">
        {{ error }}
      </p>
    </div>

    <!-- Compose / New Task Action Row -->
    <div class="action-row">
      <button class="compose-btn" @click="emit('compose')" title="单页编排 / 单张幻灯片设计">
        <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true">
          <rect x="2" y="1" width="12" height="10" rx="2"/>
          <line x1="5" y1="5" x2="11" y2="5"/>
          <line x1="5" y1="8" x2="9" y2="8"/>
        </svg>
        单页编排
      </button>
    </div>

    <!-- Task history -->
    <div class="task-list" role="region" aria-label="任务历史">
      <div class="task-list-header">
        <h3 class="task-list-title">
          任务历史
          <span v-if="taskCount" class="task-count" aria-label="任务数量">{{ taskCount }}</span>
        </h3>
        <button class="new-session-btn" @click="emit('newSession')" title="新建会话（清空当前选中的任务）">
          <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true">
            <line x1="8" y1="3" x2="8" y2="13"/>
            <line x1="3" y1="8" x2="13" y2="8"/>
          </svg>
          新建会话
        </button>
      </div>
      <p v-if="tasks.length === 0" class="empty-hint">暂无任务，在上方输入需求开始</p>
      <TransitionGroup name="task-list" tag="div" role="list">
        <div
          v-for="t in tasks"
          :key="t.id"
          class="task-item"
          :class="{ active: t.id === selectedId, running: t.status === 'running' }"
          @click="emit('selectTask', t.id)"
          role="listitem"
          :aria-current="t.id === selectedId ? 'true' : undefined"
        >
          <div class="task-item-top">
            <span class="task-item-query" :title="t.query || ''">{{ t.query || '' }}</span>
            <button
              v-if="t.status !== 'running'"
              class="task-delete-btn"
              :title="'删除: ' + (t.query || t.id)"
              @click.stop="emit('deleteTask', t.id)"
              aria-label="删除任务"
            >
              <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true">
                <line x1="4" y1="4" x2="12" y2="12"/><line x1="12" y1="4" x2="4" y2="12"/>
              </svg>
            </button>
          </div>
          <div class="task-item-meta">
            <span class="task-badge" :class="t.status" :aria-label="t.status">
              <span class="badge-dot" aria-hidden="true"></span>
              {{ t.status === 'running' ? '运行中' : t.status === 'completed' ? '已完成' : t.status === 'cancelled' ? '已中断' : '失败' }}
            </span>
            <span class="task-item-time" aria-label="创建时间">{{ fmtTime(t.created_at) }}</span>
          </div>
          <div v-if="t.total_count > 0" class="task-item-progress" role="progressbar" :aria-valuenow="t.done_count" :aria-valuemax="t.total_count">
            <div class="mini-bar"><div class="mini-bar-fill" :class="{ done: t.status === 'completed' }" :style="{ width: Math.round((t.done_count / t.total_count) * 100) + '%' }" /></div>
            <span class="mini-count" aria-hidden="true">{{ t.done_count }}/{{ t.total_count }}</span>
          </div>
          <div v-if="(t.total_tokens ?? 0) > 0" class="task-item-tokens" aria-label="消耗 tokens">
            {{ fmtTokens(t.total_tokens ?? 0) }} tokens
          </div>
        </div>
      </TransitionGroup>
    </div>

    <!-- Preference editor modal -->
    <Teleport to="body">
      <div v-if="showPrefs" class="prefs-overlay" @click.self="closePrefs">
        <div class="prefs-modal" role="dialog" aria-label="偏好设置">
          <div class="prefs-header">
            <h3>风格偏好设置</h3>
            <button class="prefs-close" @click="closePrefs" aria-label="关闭">&times;</button>
          </div>

          <div v-if="prefsLoading" class="prefs-loading">正在加载 LLM 偏好分析...</div>

          <template v-else-if="prefsEditing">
            <p class="prefs-hint">以下是根据您的历史任务自动总结的偏好，可手动修改后保存。</p>

            <label class="prefs-field">配色主题 (逗号分隔)
              <input v-model="(prefsEditing.preferred_themes as any)" :placeholder="(prefsSummary?.preferred_themes || []).join(', ') || '如 ocean_soft, sage_calm'" class="prefs-input"/>
            </label>

            <label class="prefs-field">偏好颜色 (逗号分隔)
              <input v-model="(prefsEditing.preferred_colors as any)" :placeholder="(prefsSummary?.preferred_colors || []).join(', ') || '如 蓝色系, 暖色调'" class="prefs-input"/>
            </label>

            <label class="prefs-field">语言风格
              <select v-model="prefsEditing.language_tone" class="prefs-input">
                <option value="">自动</option>
                <option value="formal">正式</option>
                <option value="semi-formal">半正式</option>
                <option value="casual">轻松</option>
              </select>
            </label>

            <label class="prefs-field">典型页数
              <input v-model.number="prefsEditing.typical_page_count" type="number" min="4" max="50" class="prefs-input"/>
            </label>

            <label class="prefs-field">布局偏好 (逗号分隔)
              <input v-model="(prefsEditing.layout_preferences as any)" :placeholder="(prefsSummary?.layout_preferences || []).join(', ') || '如 图表优先, 双栏对比'" class="prefs-input"/>
            </label>

            <label class="prefs-field">内容模式 (逗号分隔)
              <input v-model="(prefsEditing.content_patterns as any)" :placeholder="(prefsSummary?.content_patterns || []).join(', ') || '如 案例驱动, 数据支撑'" class="prefs-input"/>
            </label>

            <div class="prefs-actions">
              <button class="prefs-save-btn" @click="savePrefs" :disabled="prefsSaved">
                {{ prefsSaved ? '已保存' : '保存偏好' }}
              </button>
              <button class="prefs-cancel-btn" @click="closePrefs">关闭</button>
            </div>
          </template>

          <p v-else class="prefs-loading">暂无历史数据，完成几个任务后将自动生成偏好分析。</p>
        </div>
      </div>
    </Teleport>
  </aside>
</template>

<style scoped>
/* ═══════════════════════════════════════════════════════════════════ */
/* Sidebar — Modern Light Theme                                      */
/* ═══════════════════════════════════════════════════════════════════ */
.sidebar {
  width: var(--sidebar-w); min-width: var(--sidebar-w);
  background: var(--bg-sidebar);
  border-right: 1px solid var(--border);
  color: var(--text);
  display: flex; flex-direction: column;
  height: 100vh; position: sticky; top: 0;
  overflow-y: auto; overflow-x: hidden;
  background-image: radial-gradient(ellipse at 100% 0%, rgba(99,102,241,0.04) 0%, transparent 60%);
}

/* ── Header ─────────────────────────────────────────────────────── */
.sidebar-header {
  padding: 1.25rem 1.25rem 1rem;
  border-bottom: 1px solid var(--border-light);
  cursor: pointer;
  display: flex; align-items: center; gap: 0.75rem;
}
.logo-icon svg { width: 36px; height: 36px; flex-shrink: 0; }
.sidebar-logo { font-size: 1rem; font-weight: 700; color: var(--text); letter-spacing: -0.01em; }
.sidebar-sub { font-size: 0.68rem; color: var(--text-muted); display: block; margin-top: 1px; }

/* ── User Bar ───────────────────────────────────────────────────── */
.user-bar {
  display: flex; align-items: center; gap: 0.6rem;
  padding: 0.65rem 1.25rem;
  border-bottom: 1px solid var(--border-light);
}
.user-avatar {
  width: 28px; height: 28px; border-radius: var(--radius-full);
  background: var(--accent-soft); color: var(--accent);
  display: flex; align-items: center; justify-content: center;
  font-size: 0.72rem; font-weight: 700; flex-shrink: 0;
}
.user-avatar.guest { background: var(--bg-muted); color: var(--text-muted); }
.user-name { font-size: 0.75rem; color: var(--text-secondary); flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.user-name.dim { color: var(--text-muted); }
.logout-btn {
  background: none; border: 1px solid var(--border);
  color: var(--text-muted); padding: 0.18rem 0.55rem; border-radius: var(--radius-sm);
  cursor: pointer; font-size: 0.65rem; font-weight: 500;
  transition: all var(--transition); font-family: inherit;
}
.logout-btn:hover { background: var(--danger-soft); border-color: var(--danger-border); color: var(--danger); }
.user-avatar.admin { background: var(--accent-soft); color: var(--accent-text); }

.admin-shortcut {
  padding: 0.5rem 1.25rem;
  border-bottom: 1px solid var(--border);
}
.admin-shortcut-btn {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 0.5rem 0.8rem;
  background: none;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  color: var(--text-secondary);
  font-size: 0.78rem;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition);
  font-family: inherit;
}
.admin-shortcut-btn:hover {
  background: var(--accent-soft);
  border-color: var(--accent-border);
  color: var(--accent-text);
}

/* ── Action Row (compose shortcut) ────────────────────────────────── */
.action-row {
  padding: 0.65rem 1.25rem;
  border-bottom: 1px solid var(--border);
}
.compose-btn {
  width: 100%; padding: 0.55rem 0.8rem;
  border: 1.5px dashed var(--border);
  border-radius: var(--radius);
  background: transparent;
  color: var(--text-muted);
  font-size: 0.75rem; font-weight: 600;
  cursor: pointer;
  display: flex; align-items: center; justify-content: center;
  gap: 0.4rem;
  transition: all var(--transition);
  font-family: inherit;
}
.compose-btn svg { width: 14px; height: 14px; flex-shrink: 0; }
.compose-btn:hover {
  border-color: var(--accent-border);
  background: var(--accent-soft);
  color: var(--accent-text);
}
.compose-btn:active { transform: scale(0.98); }

/* ── Create Form ────────────────────────────────────────────────── */
.create-form {
  padding: 1rem 1.25rem;
  border-bottom: 1px solid var(--border);
  position: relative;
}
/* Decorative top gradient bar on create section */
.create-form::before {
  content: '';
  position: absolute; top: 0; left: 0; right: 0; height: 2px;
  background: linear-gradient(90deg, var(--accent), #8b5cf6);
  border-radius: 0;
}
.create-label {
  font-size: 0.68rem; font-weight: 600; text-transform: uppercase; letter-spacing: 0.05em;
  color: var(--text-muted); margin-bottom: 0.5rem; display: block;
}
.create-input {
  width: 100%; padding: 0.65rem 0.8rem;
  border: 1.5px solid var(--border);
  border-radius: var(--radius);
  background: var(--bg-base);
  color: var(--text); font-size: 0.8rem; resize: none; outline: none;
  font-family: inherit; line-height: 1.5;
  transition: border-color var(--transition), box-shadow var(--transition);
  box-shadow: var(--shadow-xs);
}
.create-input::placeholder { color: var(--text-disabled); }
.create-input:focus {
  border-color: var(--accent);
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
}
.create-btn {
  width: 100%; margin-top: 0.5rem; padding: 0.6rem;
  border: none; border-radius: var(--radius);
  background: var(--accent);
  color: #fff; font-size: 0.82rem; font-weight: 600;
  cursor: pointer; display: flex; align-items: center; justify-content: center;
  gap: 0.4rem;
  transition: transform var(--transition-md), box-shadow var(--transition-md), background var(--transition);
  box-shadow: 0 2px 8px rgba(99, 102, 241, 0.25);
}
.create-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 16px rgba(99, 102, 241, 0.35);
  background: var(--accent-hover);
}
.create-btn:active { transform: translateY(0); }
.create-btn:disabled { opacity: 0.55; cursor: not-allowed; transform: none; box-shadow: none; }
.create-btn.loading { pointer-events: none; }
.create-btn.busy { background: var(--warning); box-shadow: 0 2px 8px rgba(245, 158, 11, 0.25); }
.busy-hint { font-size: 0.63rem; color: var(--warning); margin-top: 0.35rem; text-align: center; }
.error-hint { font-size: 0.63rem; color: var(--error); margin-top: 0.35rem; text-align: center; padding: 0.25rem 0.5rem; background: color-mix(in srgb, var(--error) 10%, transparent); border-radius: 4px; }
.btn-spinner {
  width: 14px; height: 14px;
  border: 2px solid rgba(255,255,255,0.3);
  border-top-color: #fff; border-radius: 50%;
  animation: spin 0.7s linear infinite;
  flex-shrink: 0;
}

/* ── Task List ─────────────────────────────────────────────────── */
.task-list { flex: 1; overflow-y: auto; padding: 0.6rem 0.75rem; }
.task-list-header {
  display: flex; align-items: center; justify-content: space-between;
  margin-bottom: 0.5rem; padding: 0 0.5rem;
}
.task-list-title {
  font-size: 0.68rem; font-weight: 600;
  text-transform: uppercase; letter-spacing: 0.05em;
  color: var(--text-muted);
  display: flex; align-items: center; gap: 0.4rem;
}
.task-count {
  background: var(--accent-soft); color: var(--accent);
  font-size: 0.6rem; padding: 0.1rem 0.4rem; border-radius: var(--radius-full);
  font-weight: 600;
}
.new-session-btn {
  display: flex; align-items: center; gap: 0.25rem;
  padding: 0.2rem 0.5rem; border-radius: var(--radius-sm);
  border: 1px solid var(--accent-border);
  background: var(--accent-soft); color: var(--accent-text);
  font-size: 0.65rem; font-weight: 600;
  cursor: pointer; transition: all var(--transition);
  font-family: inherit;
}
.new-session-btn svg { width: 11px; height: 11px; }
.new-session-btn:hover { background: var(--accent); color: #fff; border-color: var(--accent); }

.task-item {
  padding: 0.6rem 0.7rem; border-radius: var(--radius);
  cursor: pointer; margin-bottom: 0.2rem;
  transition: background var(--transition), border-color var(--transition);
  border: 1px solid transparent;
}
.task-item:hover {
  background: var(--accent-soft);
  border-color: var(--accent-border);
}
.task-item.active {
  background: var(--accent-soft);
  border-color: var(--accent-border);
}
.task-item.running {
  border-left: 3px solid var(--accent);
  padding-left: calc(0.7rem - 2px);
}

.task-item-top { display: flex; justify-content: space-between; align-items: flex-start; gap: 0.3rem; }
.task-delete-btn {
  background: none; border: none; color: var(--text-disabled);
  cursor: pointer; padding: 0 0.15rem; line-height: 1;
  flex-shrink: 0; opacity: 0.4;
  transition: opacity var(--transition), color var(--transition);
  display: flex; align-items: center; justify-content: center;
}
.task-delete-btn svg { width: 14px; height: 14px; }
.task-item:hover .task-delete-btn, .task-item.active .task-delete-btn { opacity: 1; }
.task-delete-btn:hover { color: var(--danger); }
.task-item-query {
  font-size: 0.78rem; color: var(--text-secondary); line-height: 1.35;
  overflow: hidden; display: -webkit-box;
  -webkit-line-clamp: 2; -webkit-box-orient: vertical;
  flex: 1; min-width: 0;
}
.task-item-meta { display: flex; align-items: center; gap: 0.5rem; }
.task-badge {
  font-size: 0.6rem; font-weight: 600;
  padding: 0.12rem 0.4rem; border-radius: var(--radius-full);
  display: flex; align-items: center; gap: 0.2rem;
}
.task-badge.running { background: var(--info-soft); color: var(--info); }
.task-badge.completed { background: var(--success-soft); color: var(--success); }
.task-badge.cancelled { background: var(--warning-soft); color: var(--warning); }
.task-badge.failed { background: var(--danger-soft); color: var(--danger); }
.badge-dot { width: 5px; height: 5px; border-radius: 50%; background: currentColor; }
.task-badge.running .badge-dot { animation: pulse 1.5s infinite; }
.task-item-time { font-size: 0.6rem; color: var(--text-muted); }
.task-item-progress { display: flex; align-items: center; gap: 0.4rem; margin-top: 0.3rem; }
.mini-bar { flex: 1; height: 3px; background: var(--border); border-radius: 2px; overflow: hidden; }
.mini-bar-fill { height: 100%; background: var(--accent); border-radius: 2px; transition: width 0.6s; }
.mini-bar-fill.done { background: var(--success); }
.mini-count { font-size: 0.6rem; color: var(--text-muted); min-width: 2.5em; text-align: right; }
.task-item-tokens { font-size: 0.6rem; color: var(--accent); margin-top: 0.2rem; opacity: 0.7; }
.empty-hint { font-size: 0.75rem; color: var(--text-muted); padding: 0.5rem 0.25rem; line-height: 1.5; }

.task-list-enter-active { transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1); }
.task-list-leave-active { transition: all 0.2s ease-in; }
.task-list-enter-from { opacity: 0; transform: translateX(-16px); }
.task-list-leave-to { opacity: 0; transform: translateX(-16px); }

/* ═══════════════════════════════════════════════════════════════════ */
/* Keyframes                                                        */
/* ═══════════════════════════════════════════════════════════════════ */
@keyframes spin { to { transform: rotate(360deg); } }
@keyframes pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.5; } }

/* ── Preference Editor ────────────────────────────────────────── */
.prefs-btn {
  background: transparent; border: 1px solid var(--accent-border);
  color: var(--accent-text); padding: 0.15rem 0.5rem; border-radius: var(--radius-sm);
  font-size: 0.65rem; cursor: pointer; margin-right: 0.25rem;
}
.prefs-btn:hover { background: var(--accent-soft); }

.prefs-overlay {
  position: fixed; inset: 0; background: rgba(0,0,0,0.35);
  display: flex; align-items: center; justify-content: center;
  z-index: 9999;
}
.prefs-modal {
  background: var(--bg-primary); border-radius: 12px;
  padding: 1.5rem; width: min(480px, 90vw); max-height: 80vh;
  overflow-y: auto; box-shadow: 0 8px 32px rgba(0,0,0,0.18);
}
.prefs-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 0.75rem; }
.prefs-header h3 { font-size: 1rem; font-weight: 600; }
.prefs-close { background: none; border: none; font-size: 1.25rem; cursor: pointer; color: var(--text-muted); }
.prefs-loading { text-align: center; padding: 2rem; color: var(--text-muted); font-size: 0.85rem; }
.prefs-hint { font-size: 0.72rem; color: var(--text-muted); margin-bottom: 0.75rem; }
.prefs-field { display: block; margin-bottom: 0.6rem; font-size: 0.75rem; color: var(--text-secondary); }
.prefs-input {
  display: block; width: 100%; margin-top: 0.15rem;
  padding: 0.4rem 0.5rem; border: 1px solid var(--accent-border);
  border-radius: 6px; font-size: 0.82rem; background: var(--bg-primary);
}
.prefs-input:focus { outline: none; border-color: var(--accent); box-shadow: 0 0 0 2px var(--accent-soft); }
.prefs-actions { display: flex; gap: 0.5rem; margin-top: 1rem; }
.prefs-save-btn {
  padding: 0.4rem 1.25rem; border-radius: 6px; border: none;
  background: var(--accent); color: #fff; font-size: 0.82rem; font-weight: 600; cursor: pointer;
}
.prefs-save-btn:disabled { opacity: 0.5; cursor: default; }
.prefs-save-btn:hover:not(:disabled) { background: var(--accent-hover); }
.prefs-cancel-btn {
  padding: 0.4rem 1rem; border-radius: 6px; border: 1px solid var(--accent-border);
  background: transparent; color: var(--text-secondary); font-size: 0.82rem; cursor: pointer;
}
</style>
