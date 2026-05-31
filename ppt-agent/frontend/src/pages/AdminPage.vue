<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
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

// 详情弹窗
const detailItem = ref<Record<string, unknown> | null>(null);
const detailTitle = ref('');
const deletingId = ref<number | null>(null);

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
  navigator.clipboard.writeText(str).then(() => {
    // 静默成功
  }).catch(() => {
    // fallback
    const ta = document.createElement('textarea');
    ta.value = str;
    document.body.appendChild(ta);
    ta.select();
    document.execCommand('copy');
    document.body.removeChild(ta);
  });
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
  <div class="admin-layout">
    <!-- Header -->
    <header class="admin-header">
      <div class="admin-header-left">
        <h1 class="admin-title">管理后台</h1>
        <span class="admin-subtitle">{{ auth.user?.email }}</span>
      </div>
      <div class="header-actions">
        <button class="btn-ghost" @click="reload" :disabled="loading">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M23 4v6h-6M1 20v-6h6"/><path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"/></svg>
          刷新
        </button>
        <button class="btn-ghost" @click="router.push('/dashboard')">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M19 12H5M12 19l-7-7 7-7"/></svg>
          返回
        </button>
      </div>
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
          <div class="stat-value">{{ stats.user_count }}</div>
          <div class="stat-label">用户总数</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">{{ stats.task_count }}</div>
          <div class="stat-label">任务总数</div>
        </div>
        <div class="stat-card">
          <div class="stat-value stat-running">{{ stats.running_count }}</div>
          <div class="stat-label">运行中</div>
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
        <table class="data-table" v-if="users.length">
          <thead>
            <tr>
              <th>ID</th>
              <th>邮箱</th>
              <th>角色</th>
              <th>注册时间</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="u in users" :key="u.id" @dblclick="openDetail(u, '用户详情')">
              <td class="mono">{{ u.id }}</td>
              <td>
                <div class="cell-with-copy">
                  <span>{{ u.email }}</span>
                  <button class="copy-btn" @click.stop="copyText(u.email)" title="复制">
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>
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
        <table class="data-table" v-if="tasks.length">
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
            <tr v-for="t in tasks" :key="t.id" @dblclick="openDetail(t, '任务详情')">
              <td class="mono">{{ String(t.id).slice(0, 12) }}...</td>
              <td class="mono">{{ t.user_id }}</td>
              <td class="query-cell" :title="String(t.query)">
                <div class="cell-with-copy">
                  <span>{{ truncate(t.query) }}</span>
                  <button class="copy-btn" @click.stop="copyText(t.query)" title="复制">
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>
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
        <table class="data-table" v-if="logAnalyses.length">
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
            <tr v-for="l in logAnalyses" :key="l.id" @dblclick="openDetail(l, '日志分析详情')">
              <td class="mono">{{ l.id }}</td>
              <td class="mono" :title="String(l.task_id)">{{ String(l.task_id).slice(0, 8) }}...</td>
              <td><span :class="['badge', l.trigger_type === 'failed' ? 'badge-failed' : 'badge-idle']">{{ l.trigger_type }}</span></td>
              <td class="text-cell" :title="String(l.root_cause)">
                <div class="cell-with-copy">
                  <span>{{ truncate(l.root_cause, 40) }}</span>
                  <button class="copy-btn" @click.stop="copyText(l.root_cause)" title="复制">
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>
                  </button>
                </div>
              </td>
              <td class="text-cell" :title="String(l.suggestion)">
                <div class="cell-with-copy">
                  <span>{{ truncate(l.suggestion, 40) }}</span>
                  <button class="copy-btn" @click.stop="copyText(l.suggestion)" title="复制">
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>
                  </button>
                </div>
              </td>
              <td class="mono">{{ l.tokens_used }}</td>
              <td>{{ l.model_used || '-' }}</td>
              <td>{{ fmtTime(l.created_at) }}</td>
              <td>
                <button class="delete-btn" @click.stop="deleteLog(l.id)" :disabled="deletingId === l.id" title="删除">
                  <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6"/><path d="M10 11v6M14 11v6"/><path d="M9 6V4a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2"/></svg>
                </button>
              </td>
            </tr>
          </tbody>
        </table>
        <div v-else class="empty-state">暂无日志分析数据</div>
      </div>

      <!-- Style profiles table -->
      <div v-if="activeTab === 'profiles'" class="table-container">
        <table class="data-table" v-if="profiles.length">
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
            <tr v-for="p in profiles" :key="p.user_id" @dblclick="openDetail(p, '风格偏好详情')">
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
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>
                复制全部
              </button>
              <button class="modal-close" @click="closeDetail">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
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
                    <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.admin-layout {
  min-height: 100vh;
  background: var(--bg-muted);
  padding: 0 24px 40px;
}

.admin-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 0 16px;
  border-bottom: 1px solid var(--border);
  margin-bottom: 24px;
}

.admin-header-left {
  display: flex;
  align-items: baseline;
  gap: 12px;
}

.admin-title {
  font-size: 20px;
  font-weight: 600;
  color: var(--text);
}

.admin-subtitle {
  font-size: 13px;
  color: var(--text-muted);
}

.header-actions {
  display: flex;
  gap: 8px;
}

.admin-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 80px;
  color: var(--text-secondary);
}

.spinner {
  width: 24px;
  height: 24px;
  border: 2px solid var(--border);
  border-top-color: var(--accent);
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}

@keyframes spin { to { transform: rotate(360deg); } }

.admin-error {
  padding: 16px;
  background: var(--danger-soft);
  border: 1px solid var(--danger-border);
  border-radius: var(--radius);
  color: var(--danger);
  margin: 20px 0;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}

.stat-card {
  background: var(--bg-base);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  padding: 20px 24px;
  box-shadow: var(--shadow-sm);
}

.stat-value {
  font-size: 32px;
  font-weight: 700;
  color: var(--text);
  line-height: 1;
}

.stat-running { color: var(--info); }

.stat-label {
  font-size: 13px;
  color: var(--text-secondary);
  margin-top: 6px;
}

.tabs {
  display: flex;
  gap: 4px;
  border-bottom: 1px solid var(--border);
  margin-bottom: 20px;
}

.tab {
  padding: 10px 16px;
  font-size: 14px;
  font-weight: 500;
  color: var(--text-secondary);
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  cursor: pointer;
  transition: all var(--transition);
  margin-bottom: -1px;
}

.tab:hover { color: var(--text); }
.tab.active { color: var(--accent); border-bottom-color: var(--accent); }

.table-container {
  background: var(--bg-base);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  overflow: auto;
  box-shadow: var(--shadow-sm);
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.data-table th {
  text-align: left;
  padding: 10px 16px;
  background: var(--bg-muted);
  color: var(--text-secondary);
  font-weight: 500;
  border-bottom: 1px solid var(--border);
  white-space: nowrap;
}

.data-table td {
  padding: 10px 16px;
  border-bottom: 1px solid var(--border-light);
  color: var(--text);
  max-width: 300px;
}

.data-table tr:last-child td { border-bottom: none; }
.data-table tr:hover td { background: var(--bg-muted); cursor: pointer; }

.mono {
  font-family: ui-monospace, monospace;
  font-size: 12px;
  color: var(--text-secondary);
}

.query-cell, .text-cell {
  max-width: 260px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tag-cell {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}

.cell-with-copy {
  display: flex;
  align-items: center;
  gap: 4px;
  min-width: 0;
}

.cell-with-copy > span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
}

.copy-btn {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  border: none;
  background: none;
  color: var(--text-muted);
  cursor: pointer;
  border-radius: var(--radius-xs);
  opacity: 0;
  transition: opacity var(--transition), color var(--transition);
  padding: 0;
}

tr:hover .copy-btn,
.detail-value:hover .copy-btn {
  opacity: 1;
}

.copy-btn:hover { color: var(--accent); }

.badge {
  display: inline-flex;
  align-items: center;
  padding: 2px 8px;
  border-radius: var(--radius-full);
  font-size: 12px;
  font-weight: 500;
}

.badge-user { background: var(--bg-muted); color: var(--text-secondary); }
.badge-admin { background: var(--accent-soft); color: var(--accent-text); }
.badge-failed { background: var(--danger-soft); color: var(--danger); }
.badge-idle { background: var(--info-soft); color: var(--info); }

.status-running { background: var(--info-soft); color: var(--info); }
.status-completed { background: var(--success-soft); color: var(--success); }
.status-failed { background: var(--danger-soft); color: var(--danger); }
.status-cancelled { background: var(--bg-muted); color: var(--text-muted); }

.tag {
  display: inline-block;
  padding: 2px 7px;
  border-radius: var(--radius-full);
  font-size: 11px;
  background: var(--accent-soft);
  color: var(--accent-text);
}

.color-tag { background: transparent; border: 1px solid currentColor; }

.delete-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: 1px solid var(--border);
  background: none;
  color: var(--text-muted);
  cursor: pointer;
  border-radius: var(--radius-sm);
  transition: all var(--transition);
  opacity: 0;
  padding: 0;
}

tr:hover .delete-btn { opacity: 1; }
.delete-btn:hover { background: var(--danger-soft); border-color: var(--danger-border); color: var(--danger); }
.delete-btn:disabled { opacity: 0.4; cursor: not-allowed; }

.empty-state {
  padding: 60px;
  text-align: center;
  color: var(--text-muted);
  font-size: 14px;
}

.btn-ghost {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  background: var(--bg-base);
  color: var(--text-secondary);
  font-size: 13px;
  cursor: pointer;
  transition: all var(--transition);
}

.btn-ghost:hover { background: var(--bg-muted); color: var(--text); }
.btn-ghost:disabled { opacity: 0.5; cursor: not-allowed; }

/* ── Modal ─────────────────────────────────────────────────────────────── */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 20px;
  backdrop-filter: blur(2px);
}

.modal {
  background: var(--bg-base);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  width: 100%;
  max-width: 640px;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
  box-shadow: var(--shadow-lg);
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border);
  flex-shrink: 0;
}

.modal-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text);
}

.modal-copy-all {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 6px 12px;
  border: 1px solid var(--border);
  background: var(--bg-base);
  color: var(--text-secondary);
  font-size: 13px;
  cursor: pointer;
  border-radius: var(--radius);
  transition: all var(--transition);
}

.modal-copy-all:hover { background: var(--accent-soft); border-color: var(--accent-border); color: var(--accent-text); }

.modal-header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.modal-close {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: 1px solid var(--border);
  background: none;
  color: var(--text-muted);
  cursor: pointer;
  border-radius: var(--radius);
  transition: all var(--transition);
  padding: 0;
}

.modal-close:hover { background: var(--bg-muted); color: var(--text); }

.modal-body {
  overflow-y: auto;
  padding: 12px 20px 20px;
  flex: 1;
}

.detail-row {
  display: grid;
  grid-template-columns: 120px 1fr;
  gap: 12px;
  padding: 8px 0;
  border-bottom: 1px solid var(--border-light);
  align-items: start;
}

.detail-row:last-child { border-bottom: none; }

.detail-label {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
  padding-top: 2px;
  flex-shrink: 0;
}

.detail-value {
  min-width: 0;
}

.detail-pre {
  margin: 0;
  font-size: 13px;
  color: var(--text);
  white-space: pre;
  overflow-x: auto;
  max-height: 300px;
  overflow-y: auto;
  font-family: ui-monospace, monospace;
  background: var(--bg-muted);
  padding: 6px 10px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border);
}

.detail-pre-wrap { white-space: pre-wrap; word-break: break-all; }
</style>
