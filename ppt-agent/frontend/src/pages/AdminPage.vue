<script setup lang="ts">
import { computed, ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { ArrowLeft, Copy, RefreshCw, Search, Trash2, X } from 'lucide-vue-next';
import {
  fetchAdminStats,
  fetchAdminUsers,
  fetchAdminTasks,
  fetchAdminLogAnalyses,
  fetchAdminStyleProfiles,
  type AdminStats,
  type AdminUser,
  type AdminTaskRecord,
  type AdminLogAnalysis,
  type AdminStyleProfile,
} from '../api';
import { authState } from '../stores/auth';
import AppShell from '../components/AppShell.vue';

const router = useRouter();
const auth = authState;

const stats = ref<AdminStats>({ user_count: 0, task_count: 0, running_count: 0 });
const users = ref<AdminUser[]>([]);
const tasks = ref<AdminTaskRecord[]>([]);
const logAnalyses = ref<AdminLogAnalysis[]>([]);
const profiles = ref<AdminStyleProfile[]>([]);
const loading = ref(true);
const activeTab = ref<'users' | 'tasks' | 'logs' | 'profiles'>('users');
const error = ref('');
const filterQuery = ref('');

// 详情弹窗
const detailItem = ref<Record<string, unknown> | null>(null);
const detailTitle = ref('');
const deletingId = ref<number | null>(null);

const normalizedFilter = computed(() => filterQuery.value.trim().toLowerCase());
function matchesFilter(...values: unknown[]) {
  if (!normalizedFilter.value) return true;
  return values.some(value => String(value ?? '').toLowerCase().includes(normalizedFilter.value));
}
const filteredUsers = computed(() => users.value.filter(item => matchesFilter(item.id, item.email, item.is_admin ? '管理员' : '用户')));
const filteredTasks = computed(() => tasks.value.filter(item => matchesFilter(item.id, item.user_id, item.user_email, item.query, item.status)));
const filteredLogs = computed(() => logAnalyses.value.filter(item => matchesFilter(item.id, item.task_id, item.trigger_type, item.root_cause, item.suggestion, item.model_used)));
const filteredProfiles = computed(() => profiles.value.filter(item => matchesFilter(item.user_id, item.language_tone, item.preferred_themes, item.preferred_colors, item.content_types)));

onMounted(async () => {
  await reload();
});

async function reload() {
  loading.value = true;
  error.value = '';
  try {
    const [s, u, t, l, p] = await Promise.all([
      fetchAdminStats(),
      fetchAdminUsers(),
      fetchAdminTasks(),
      fetchAdminLogAnalyses(),
      fetchAdminStyleProfiles(),
    ]);
    stats.value = s;
    users.value = Array.isArray(u) ? u : [];
    tasks.value = Array.isArray(t) ? t : [];
    logAnalyses.value = Array.isArray(l) ? l : [];
    profiles.value = Array.isArray(p) ? p : [];
  } catch (e: any) {
    if (e.message?.includes('需要管理员')) {
      router.push('/dashboard');
    } else {
      error.value = e.message || '加载失败';
    }
  } finally {
    loading.value = false;
  }
}

function fmtTime(iso: unknown): string {
  if (!iso) return '-';
  try { return new Date(String(iso)).toLocaleString('zh-CN'); } catch { return '-'; }
}

function statusClass(s: string): string {
  switch (s) {
    case 'running': return 'status-running';
    case 'completed': return 'status-completed';
    case 'failed': return 'status-failed';
    case 'cancelled': return 'status-cancelled';
    default: return '';
  }
}

function safeParseJson(s: unknown): unknown[] {
  if (!s) return [];
  try { const r = JSON.parse(String(s)); return Array.isArray(r) ? r : []; } catch { return []; }
}

function truncate(s: unknown, n = 60): string {
  const str = String(s ?? '');
  return str.length > n ? str.slice(0, n) + '...' : str || '-';
}

// ── 复制 ─────────────────────────────────────────────────────────────────
function copyText(text: unknown) {
  const str = String(text ?? '');
  if (navigator.clipboard?.writeText) {
    navigator.clipboard.writeText(str).catch(() => {
      execCopyFallback(str);
    });
  } else {
    execCopyFallback(str);
  }
}

function execCopyFallback(str: string) {
  const ta = document.createElement('textarea');
  ta.value = str;
  ta.style.position = 'fixed';
  ta.style.opacity = '0';
  document.body.appendChild(ta);
  ta.focus();
  ta.select();
  document.execCommand('copy');
  document.body.removeChild(ta);
}

// ── 详情弹窗 ─────────────────────────────────────────────────────────────
function openDetail(row: Record<string, unknown>, title: string) {
  detailItem.value = row;
  detailTitle.value = title;
}

function closeDetail() {
  detailItem.value = null;
  detailTitle.value = '';
}

function detailFieldLabel(key: string): string {
  const map: Record<string, string> = {
    id: 'ID', task_id: '任务ID', user_id: '用户ID', email: '邮箱',
    is_admin: '管理员', created_at: '创建时间', updated_at: '更新时间',
    query: '查询', status: '状态', done_count: '完成数', total_count: '总数',
    duration: '耗时', error: '错误', prompt_tokens: 'Prompt Token',
    completion_tokens: 'Completion Token', total_tokens: '总Token',
    conversation_content: '对话内容',
    trigger_type: '触发类型', log_snippet: '原始错误日志',
    analysis: '分析详情', root_cause: '根本原因',
    suggestion: '修复建议', tokens_used: 'Token用量',
    model_used: '模型',
    preferred_themes: '偏好主题', preferred_colors: '配色偏好',
    content_patterns: '内容模式', content_types: '内容类型',
    language_tone: '语言风格', typical_page_count: '典型页数',
    special_notes: '特殊备注', task_count: '任务数',
  };
  return map[key] ?? key;
}

function detailFieldValue(key: string, val: unknown): string {
  if (val === null || val === undefined) return '-';
  if (typeof val === 'object') {
    if (Array.isArray(val)) {
      if (val.length === 0) return '-';
      return (val as unknown[]).map(v => String(v)).join(', ');
    }
    try { return JSON.stringify(val, null, 2); } catch { return String(val); }
  }
  if (key.endsWith('_at') || key === 'created_at' || key === 'updated_at') {
    return fmtTime(val);
  }
  return String(val);
}

function isDetailLong(key: string): boolean {
  return ['query', 'analysis', 'log_snippet', 'root_cause', 'suggestion', 'conversation_content',
    'preferred_themes', 'preferred_colors', 'content_patterns', 'content_types',
    'special_notes'].includes(key);
}

function copyAll() {
  if (!detailItem.value) return;
  const lines: string[] = [];
  for (const [key, val] of Object.entries(detailItem.value)) {
    const label = detailFieldLabel(key);
    const value = detailFieldValue(key, val);
    lines.push(`${label}:\n${value}`);
  }
  copyText(lines.join('\n\n' + '─'.repeat(40) + '\n\n'));
}

// ── 日志删除 ─────────────────────────────────────────────────────────────
async function deleteLog(id: number) {
  if (!confirm('确定删除这条日志分析记录？')) return;
  deletingId.value = id;
  try {
    const res = await fetch(`/api/admin/log-analyses/${id}`, {
      method: 'DELETE',
      headers: { 'Authorization': `Bearer ${localStorage.getItem('ppt_agent_token')}` },
    });
    if (res.ok) {
      logAnalyses.value = logAnalyses.value.filter(l => l.id !== id);
      stats.value.task_count = Math.max(0, stats.value.task_count - 1);
    }
  } catch { /* ignore */ }
  finally {
    deletingId.value = null;
  }
}
</script>

<template>
  <AppShell title="管理后台" eyebrow="运营与诊断" content-class="admin-shell-content">
    <template #actions>
      <button class="admin-top-action" type="button" :disabled="loading" @click="reload">
        <RefreshCw :size="17" :class="{ spin: loading }" />
        <span>刷新</span>
      </button>
      <button class="admin-top-action" type="button" @click="router.push('/dashboard')">
        <ArrowLeft :size="17" />
        <span>任务</span>
      </button>
    </template>

    <div class="admin-layout">
      <header class="admin-header">
        <div class="admin-header-left">
          <h2 class="admin-title">系统概览</h2>
          <span class="admin-subtitle">{{ auth.user?.email }}</span>
        </div>
        <label class="admin-search">
          <span class="visually-hidden">筛选当前数据</span>
          <Search :size="16" aria-hidden="true" />
          <input v-model="filterQuery" type="search" placeholder="筛选当前数据" />
        </label>
      </header>

    <!-- Loading / Error -->
    <div v-if="loading" class="admin-loading">
      <div class="spinner"></div>
      <span>加载中...</span>
    </div>
    <div v-else-if="error" class="admin-error">{{ error }}</div>

    <template v-else>
      <!-- Stats cards -->
      <div class="stats-grid">
        <div class="stat-card">
          <div class="stat-label">用户总数</div>
          <div class="stat-value">{{ stats.user_count }}</div>
        </div>
        <div class="stat-card">
          <div class="stat-label">任务总数</div>
          <div class="stat-value">{{ stats.task_count }}</div>
        </div>
        <div class="stat-card">
          <div class="stat-label">运行中</div>
          <div class="stat-value stat-running">{{ stats.running_count }}</div>
        </div>
      </div>

      <!-- Tabs -->
      <div class="tabs">
        <button :class="['tab', { active: activeTab === 'users' }]" @click="activeTab = 'users'">
          用户 ({{ users.length }})
        </button>
        <button :class="['tab', { active: activeTab === 'tasks' }]" @click="activeTab = 'tasks'">
          任务 ({{ tasks.length }})
        </button>
        <button :class="['tab', { active: activeTab === 'logs' }]" @click="activeTab = 'logs'">
          日志分析 ({{ logAnalyses.length }})
        </button>
        <button :class="['tab', { active: activeTab === 'profiles' }]" @click="activeTab = 'profiles'">
          风格偏好 ({{ profiles.length }})
        </button>
      </div>

      <!-- Users table -->
      <div v-if="activeTab === 'users'" class="table-container">
        <table class="data-table" v-if="filteredUsers.length">
          <thead>
            <tr>
              <th>ID</th>
              <th>邮箱</th>
              <th>角色</th>
              <th>注册时间</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="u in filteredUsers" :key="u.id" @dblclick="openDetail(u, '用户详情')">
              <td class="mono">{{ u.id }}</td>
              <td>
                <div class="cell-with-copy">
                  <span>{{ u.email }}</span>
                  <button class="copy-btn" @click.stop="copyText(u.email)" title="复制">
                    <Copy :size="13" />
                  </button>
                </div>
              </td>
              <td><span :class="['badge', u.is_admin ? 'badge-admin' : 'badge-user']">{{ u.is_admin ? '管理员' : '用户' }}</span></td>
              <td>{{ fmtTime(u.created_at) }}</td>
            </tr>
          </tbody>
        </table>
        <div v-else class="empty-state">暂无用户数据</div>
      </div>

      <!-- Tasks table -->
      <div v-if="activeTab === 'tasks'" class="table-container">
        <table class="data-table" v-if="filteredTasks.length">
          <thead>
            <tr>
              <th>ID</th>
              <th>用户</th>
              <th>查询</th>
              <th>状态</th>
              <th>进度</th>
              <th>耗时</th>
              <th>创建时间</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="t in filteredTasks" :key="t.id" @dblclick="openDetail(t, '任务详情')">
              <td class="mono">{{ String(t.id).slice(0, 12) }}...</td>
              <td>
                <div class="admin-user-cell">
                  <strong>{{ t.user_email || `用户 ${t.user_id}` }}</strong>
                  <small class="mono">#{{ t.user_id }}</small>
                </div>
              </td>
              <td class="query-cell" :title="String(t.query)">
                <div class="cell-with-copy">
                  <span>{{ truncate(t.query) }}</span>
                  <button class="copy-btn" @click.stop="copyText(t.query)" title="复制">
                    <Copy :size="13" />
                  </button>
                </div>
              </td>
              <td><span :class="['badge', statusClass(t.status)]">{{ t.status }}</span></td>
              <td>{{ t.done_count }}/{{ t.total_count }}</td>
              <td>{{ t.duration || '-' }}</td>
              <td>{{ fmtTime(t.created_at) }}</td>
            </tr>
          </tbody>
        </table>
        <div v-else class="empty-state">暂无任务数据</div>
      </div>

      <!-- Log analyses table -->
      <div v-if="activeTab === 'logs'" class="table-container">
        <table class="data-table" v-if="filteredLogs.length">
          <thead>
            <tr>
              <th style="width:50px">ID</th>
              <th>任务</th>
              <th>触发</th>
              <th>根本原因</th>
              <th>建议</th>
              <th>Token</th>
              <th>模型</th>
              <th>时间</th>
              <th style="width:60px">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="l in filteredLogs" :key="l.id" @dblclick="openDetail(l, '日志分析详情')">
              <td class="mono">{{ l.id }}</td>
              <td class="mono" :title="String(l.task_id)">{{ String(l.task_id).slice(0, 8) }}...</td>
              <td><span :class="['badge', l.trigger_type === 'failed' ? 'badge-failed' : 'badge-idle']">{{ l.trigger_type }}</span></td>
              <td class="text-cell" :title="String(l.root_cause)">
                <div class="cell-with-copy">
                  <span>{{ truncate(l.root_cause, 40) }}</span>
                  <button class="copy-btn" @click.stop="copyText(l.root_cause)" title="复制">
                    <Copy :size="13" />
                  </button>
                </div>
              </td>
              <td class="text-cell" :title="String(l.suggestion)">
                <div class="cell-with-copy">
                  <span>{{ truncate(l.suggestion, 40) }}</span>
                  <button class="copy-btn" @click.stop="copyText(l.suggestion)" title="复制">
                    <Copy :size="13" />
                  </button>
                </div>
              </td>
              <td class="mono">{{ l.tokens_used }}</td>
              <td>{{ l.model_used || '-' }}</td>
              <td>{{ fmtTime(l.created_at) }}</td>
              <td>
                <button class="delete-btn" @click.stop="deleteLog(l.id)" :disabled="deletingId === l.id" title="删除">
                  <Trash2 :size="15" />
                </button>
              </td>
            </tr>
          </tbody>
        </table>
        <div v-else class="empty-state">暂无日志分析数据</div>
      </div>

      <!-- Style profiles table -->
      <div v-if="activeTab === 'profiles'" class="table-container">
        <table class="data-table" v-if="filteredProfiles.length">
          <thead>
            <tr>
              <th>用户</th>
              <th>任务数</th>
              <th>偏好主题</th>
              <th>配色</th>
              <th>内容类型</th>
              <th>语言风格</th>
              <th>更新</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="p in filteredProfiles" :key="p.user_id" @dblclick="openDetail(p, '风格偏好详情')">
              <td class="mono">{{ p.user_id }}</td>
              <td>{{ p.task_count }}</td>
              <td class="tag-cell">
                <span v-for="t in (safeParseJson(p.preferred_themes) as string[]).slice(0, 3)" :key="t" class="tag">{{ t }}</span>
              </td>
              <td class="tag-cell">
                <span v-for="c in (safeParseJson(p.preferred_colors) as string[]).slice(0, 3)" :key="c" class="tag color-tag" :style="{ background: c + '22', color: c }">{{ c }}</span>
              </td>
              <td class="tag-cell">
                <span v-for="ct in (safeParseJson(p.content_types) as string[]).slice(0, 2)" :key="ct" class="tag">{{ ct }}</span>
              </td>
              <td>{{ p.language_tone || '-' }}</td>
              <td>{{ fmtTime(p.updated_at) }}</td>
            </tr>
          </tbody>
        </table>
        <div v-else class="empty-state">暂无风格偏好数据</div>
      </div>
    </template>

    <!-- Detail Modal -->
    <Teleport to="body">
      <div v-if="detailItem" class="modal-overlay" @click.self="closeDetail">
        <div class="modal">
          <div class="modal-header">
            <h2 class="modal-title">{{ detailTitle }}</h2>
            <div class="modal-header-actions">
              <button class="modal-copy-all" @click="copyAll" title="复制全部">
                <Copy :size="15" />
                复制全部
              </button>
              <button class="modal-close" aria-label="关闭详情" @click="closeDetail">
                <X :size="19" />
              </button>
            </div>
          </div>
          <div class="modal-body">
            <div v-for="[key, val] in Object.entries(detailItem)" :key="key" class="detail-row">
              <div class="detail-label">{{ detailFieldLabel(key) }}</div>
              <div class="detail-value" :class="{ 'detail-long': isDetailLong(key) }">
                <div class="cell-with-copy">
                  <pre class="detail-pre" :class="{ 'detail-pre-wrap': isDetailLong(key) }">{{ detailFieldValue(key, val) }}</pre>
                  <button class="copy-btn" @click="copyText(val)" title="复制">
                    <Copy :size="14" />
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Teleport>
    </div>
  </AppShell>
</template>

<style scoped>
/* Rebuilt operations workspace */
.admin-header { display: flex; align-items: center; justify-content: space-between; gap: 18px; }
.admin-header-left { min-width: 0; display: flex; align-items: baseline; }
.admin-loading { display: flex; align-items: center; justify-content: center; gap: 9px; color: var(--text-muted); }
.spinner { width: 24px; height: 24px; border: 3px solid var(--border-strong); border-top-color: var(--info); border-radius: 50%; animation: spin 0.9s linear infinite; }
.tabs button { border: 0; font: inherit; cursor: pointer; }
.data-table { width: 100%; border-collapse: collapse; table-layout: auto; }
.cell-with-copy { min-width: 0; display: flex; align-items: center; gap: 6px; }.cell-with-copy > span, .cell-with-copy > pre { min-width: 0; flex: 1; }
.mono { font-family: ui-monospace, "Cascadia Code", monospace; font-variant-numeric: tabular-nums; }
.admin-user-cell { min-width: 150px; display: flex; flex-direction: column; gap: 2px; }
.admin-user-cell strong { color: var(--text-secondary); font-size: 11px; font-weight: 650; }
.admin-user-cell small { color: var(--text-muted); font-size: 9px; }
.query-cell { min-width: 220px; }.text-cell { min-width: 180px; max-width: 300px; }
.badge { display: inline-flex; align-items: center; white-space: nowrap; }
.badge-admin { color: var(--action-ink); background: var(--action-soft); }.badge-user, .badge-idle { color: var(--text-secondary); background: var(--surface-pressed); }
.status-running { color: var(--info); background: var(--info-soft); }.status-completed { color: var(--success); background: var(--success-soft); }.status-failed, .badge-failed { color: var(--danger); background: var(--danger-soft); }.status-cancelled { color: var(--warning); background: var(--warning-soft); }
.tag-cell { min-width: 150px; }.tag { margin: 2px 3px 2px 0; padding: 3px 6px; display: inline-flex; align-items: center; color: var(--text-secondary); }
.modal-overlay { position: fixed; inset: 0; z-index: var(--z-modal); display: grid; place-items: center; }
.modal { display: flex; flex-direction: column; overflow: hidden; }
.modal-header { display: flex; align-items: center; justify-content: space-between; gap: 12px; flex: 0 0 auto; border-bottom: 1px solid var(--border); }
.modal-header-actions { display: flex; align-items: center; gap: 6px; }
.modal-body { min-height: 0; flex: 1; overflow: auto; }
.detail-value { min-width: 0; }
.detail-pre-wrap { white-space: pre-wrap; word-break: break-word; }

.admin-top-action {
  min-height: 38px;
  padding: 0 11px;
  display: inline-flex;
  align-items: center;
  gap: 7px;
  border: 1px solid var(--border-strong);
  border-radius: 6px;
  color: var(--text-secondary);
  background: var(--surface);
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
}
.admin-top-action:hover { color: var(--text); background: var(--surface-muted); }
.admin-top-action .spin { animation: spin 0.9s linear infinite; }

.admin-layout { min-height: calc(100dvh - var(--topbar-height)); padding: 20px 24px 40px; background: var(--canvas); }
.admin-header { min-height: 48px; margin: 0 0 16px; padding: 0; border: 0; }
.admin-header-left { gap: 8px; }
.admin-title { margin: 0; color: var(--text); font-size: 17px; font-weight: 730; }
.admin-subtitle { color: var(--text-muted); font-size: 11px; }

.admin-search {
  width: min(320px, 42vw);
  height: 40px;
  padding: 0 10px;
  display: flex;
  align-items: center;
  gap: 8px;
  border: 1px solid var(--border-strong);
  border-radius: 6px;
  color: var(--text-muted);
  background: var(--surface);
}
.admin-search:focus-within { border-color: var(--info); box-shadow: 0 0 0 2px var(--info-soft); }
.admin-search input { min-width: 0; width: 100%; border: 0; outline: 0; color: var(--text); background: transparent; }

.admin-loading,
.admin-error { min-height: 240px; border: 1px solid var(--border); border-radius: 6px; background: var(--surface); }
.admin-error { display: grid; place-items: center; color: var(--danger); }

.stats-grid {
  margin: 0 0 16px;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0;
  overflow: hidden;
  border: 1px solid var(--border);
  border-radius: 6px;
  background: var(--surface);
}
.stat-card { min-width: 0; min-height: 84px; padding: 14px 16px; display: flex; flex-direction: column; justify-content: space-between; align-items: flex-start; border: 0; border-right: 1px solid var(--divider); border-radius: 0; background: var(--surface); box-shadow: none; }
.stat-card:last-child { border-right: 0; }
.stat-value { color: var(--text); font-size: 25px; font-weight: 740; line-height: 1; }
.stat-value.stat-running { color: var(--info); }
.stat-label { color: var(--text-muted); font-size: 10px; font-weight: 700; }

.tabs { min-height: 44px; margin: 0; padding: 4px; display: flex; align-items: center; gap: 3px; overflow-x: auto; border: 1px solid var(--border); border-bottom: 0; border-radius: 6px 6px 0 0; background: var(--surface-muted); }
.tab { min-height: 34px; padding: 0 12px; flex: 0 0 auto; border: 1px solid transparent; border-radius: 5px; color: var(--text-secondary); background: transparent; font-size: 11px; font-weight: 700; }
.tab:hover { color: var(--text); background: var(--surface); }
.tab.active { color: var(--action-ink); border-color: var(--border); background: var(--surface); box-shadow: var(--shadow-xs); }

.table-container { margin: 0; overflow: auto; border: 1px solid var(--border); border-radius: 0 0 6px 6px; background: var(--surface); box-shadow: none; }
.data-table { min-width: 760px; font-size: 11px; }
.data-table th { position: sticky; top: 0; z-index: 1; height: 38px; padding: 0 12px; color: var(--text-muted); background: var(--surface-muted); border-bottom: 1px solid var(--border); font-size: 10px; font-weight: 750; letter-spacing: 0; text-transform: none; }
.data-table td { height: 46px; padding: 7px 12px; border-bottom: 1px solid var(--divider); color: var(--text-secondary); }
.data-table tbody tr { cursor: default; }
.data-table tbody tr:hover { background: #f7f9f9; }
.data-table tbody tr:last-child td { border-bottom: 0; }
.badge { padding: 3px 7px; border-radius: 4px; font-size: 9px; font-weight: 750; }
.tag { border-radius: 4px; background: var(--surface-pressed); font-size: 9px; }
.copy-btn,
.delete-btn { width: 32px; height: 32px; display: inline-grid; place-items: center; border: 0; border-radius: 4px; color: var(--text-muted); background: transparent; }
.copy-btn:hover { color: var(--info); background: var(--info-soft); }
.delete-btn:hover { color: var(--danger); background: var(--danger-soft); }
.empty-state { min-height: 220px; display: grid; place-items: center; color: var(--text-muted); }

.modal-overlay { padding: 20px; background: rgba(15, 17, 18, 0.58); }
.modal { width: min(820px, 100%); max-height: calc(100dvh - 40px); border-radius: 8px; background: var(--surface); box-shadow: var(--shadow-lg); }
.modal-header { min-height: 60px; padding: 10px 10px 10px 18px; }
.modal-title { font-size: 15px; font-weight: 730; }
.modal-copy-all { min-height: 38px; padding: 0 10px; border-radius: 5px; }
.modal-close { width: 42px; height: 42px; border: 0; border-radius: 5px; }
.modal-body { padding: 10px 18px 18px; }
.detail-row { grid-template-columns: 116px minmax(0, 1fr); gap: 12px; padding: 9px 0; }
.detail-label { color: var(--text-muted); font-size: 11px; font-weight: 700; }
.detail-pre { padding: 8px 10px; border-color: var(--divider); border-radius: 5px; color: var(--text-secondary); background: var(--surface-muted); font-size: 11px; }

@keyframes spin { to { transform: rotate(360deg); } }

@media (max-width: 720px) {
  .admin-top-action { width: 40px; min-height: 40px; padding: 0; justify-content: center; }
  .admin-top-action span { display: none; }
  .admin-layout { padding: 14px 12px 28px; }
  .admin-header { height: auto; align-items: stretch; flex-direction: column; gap: 10px; }
  .admin-header-left { width: 100%; justify-content: space-between; }
  .admin-search { width: 100%; }
  .stat-card { min-height: 76px; padding: 12px; }
  .stat-value { font-size: 21px; }
  .tabs { border-radius: 6px 6px 0 0; }
  .modal-overlay { padding: 0; align-items: end; }
  .modal { width: 100%; max-height: 92dvh; border-radius: 8px 8px 0 0; }
  .detail-row { grid-template-columns: 1fr; gap: 4px; }
}

@media (prefers-reduced-motion: reduce) {
  .admin-top-action .spin { animation: none; }
}
</style>
