<script setup lang="ts">
import { ref, computed } from 'vue';
import { useRouter } from 'vue-router';
import type { TaskInfo } from '../types';
import { isLoggedIn } from '../api';

const props = defineProps<{
  user: { id: number; email: string } | null;
  tasks: TaskInfo[];
  selectedId: string | null;
  hasRunningTask: boolean;
  creating: boolean;
}>();

const emit = defineEmits<{
  logout: [];
  selectTask: [id: string];
  createTask: [query: string];
  deleteTask: [id: string];
}>();

const router = useRouter();
const query = ref('');

function fmtTime(iso: string): string {
  if (!iso) return '';
  return new Date(iso).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });
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
</script>

<template>
  <aside class="sidebar">
    <!-- Animated tech background -->
    <div class="sidebar-bg"></div>

    <div class="sidebar-header" @click="router.push('/')">
      <div class="logo-icon">
        <svg viewBox="0 0 40 40" fill="none">
          <rect x="5" y="7" width="30" height="22" rx="4" fill="rgba(129,140,248,0.12)" stroke="rgba(129,140,248,0.25)" stroke-width="1.5"/>
          <rect x="5" y="7" width="30" height="5" rx="4" fill="rgba(129,140,248,0.25)"/>
          <circle cx="12" cy="15" r="2" fill="rgba(165,180,252,0.5)"/>
          <circle cx="19" cy="15" r="2" fill="rgba(165,180,252,0.4)"/>
          <circle cx="26" cy="15" r="2" fill="rgba(165,180,252,0.3)"/>
          <rect x="10" y="20" width="16" height="1.5" rx="0.75" fill="rgba(165,180,252,0.2)"/>
          <rect x="10" y="23" width="20" height="1.5" rx="0.75" fill="rgba(165,180,252,0.15)"/>
        </svg>
      </div>
      <div>
        <h1 class="sidebar-logo">PPT Agent</h1>
        <span class="sidebar-sub">AI 驱动的幻灯片生成</span>
      </div>
    </div>

    <!-- User bar (always shown) -->
    <div class="user-bar">
      <template v-if="user">
        <span class="user-avatar">{{ user.email[0].toUpperCase() }}</span>
        <span class="user-name">{{ user.email }}</span>
        <button class="logout-btn" @click="onLogout" title="退出登录">退出</button>
      </template>
      <template v-else>
        <span class="user-avatar guest">?</span>
        <span class="user-name dim">未登录</span>
        <button class="logout-btn" @click="router.push('/auth')">登录</button>
      </template>
    </div>

    <!-- Create form (always shown) -->
    <div class="create-form">
      <label class="create-label">新建 PPT 任务</label>
      <textarea
        class="create-input"
        placeholder="描述你的 PPT 需求，例如：做一个关于新能源汽车的行业分析报告..."
        v-model="query"
        @keydown="handleKeydown"
        rows="3"
      ></textarea>
      <button
        class="create-btn"
        :class="{ loading: creating, busy: hasRunningTask }"
        :disabled="creating || hasRunningTask"
        @click="handleCreate"
      >
        <span v-if="hasRunningTask" class="btn-spinner"></span>
        <span>{{ hasRunningTask ? '任务执行中...' : creating ? '创建中...' : '✦ 生成 PPT' }}</span>
      </button>
      <p v-if="hasRunningTask" class="busy-hint">当前有任务正在执行，请等待完成后再创建新任务</p>
    </div>

    <!-- Task history (always shown) -->
    <div class="task-list">
      <h3 class="task-list-title">
        任务历史
        <span v-if="taskCount" class="task-count">{{ taskCount }}</span>
      </h3>
      <p v-if="tasks.length === 0" class="empty-hint">暂无任务，在上方创建第一个 PPT</p>
      <TransitionGroup name="task-list" tag="div">
        <div
          v-for="t in tasks"
          :key="t.id"
          class="task-item"
          :class="{ active: t.id === selectedId }"
          @click="emit('selectTask', t.id)"
        >
          <div class="task-item-top">
            <span class="task-item-query">{{ t.query.length > 38 ? t.query.slice(0, 38) + '...' : t.query }}</span>
            <button v-if="t.status !== 'running'" class="task-delete-btn" title="删除" @click.stop="emit('deleteTask', t.id)">×</button>
          </div>
          <div class="task-item-meta">
            <span class="task-badge" :class="t.status">
              <span class="badge-dot"></span>
              {{ t.status === 'running' ? '运行中' : t.status === 'completed' ? '已完成' : t.status === 'cancelled' ? '已中断' : '失败' }}
            </span>
            <span class="task-item-time">{{ fmtTime(t.created_at) }}</span>
          </div>
          <div v-if="t.total_count > 0" class="task-item-progress">
            <div class="mini-bar"><div class="mini-bar-fill" :class="{ done: t.status === 'completed' }" :style="{ width: Math.round((t.done_count / t.total_count) * 100) + '%' }" /></div>
            <span class="mini-count">{{ t.done_count }}/{{ t.total_count }}</span>
          </div>
        </div>
      </TransitionGroup>
    </div>
  </aside>
</template>

<style scoped>
.sidebar {
  width: var(--sidebar-w); min-width: var(--sidebar-w);
  background: linear-gradient(180deg, #0f0d2e 0%, #151345 40%, #0d1137 100%);
  color: #e2e8f0;
  display: flex; flex-direction: column;
  height: 100vh; position: sticky; top: 0;
  overflow-y: auto; overflow-x: hidden;
  position: relative;
}
.sidebar-bg {
  position: absolute; inset: 0; pointer-events: none; overflow: hidden;
  background:
    radial-gradient(ellipse at 50% 20%, rgba(99,102,241,0.08) 0%, transparent 60%),
    radial-gradient(ellipse at 80% 70%, rgba(139,92,246,0.06) 0%, transparent 50%);
}
.sidebar-bg::before {
  content: ''; position: absolute; inset: 0;
  background-image:
    linear-gradient(rgba(99,102,241,0.02) 1px, transparent 1px),
    linear-gradient(90deg, rgba(99,102,241,0.02) 1px, transparent 1px);
  background-size: 40px 40px;
}

.sidebar-header {
  padding: 1.25rem; display: flex; align-items: center; gap: 0.75rem;
  border-bottom: 1px solid rgba(99,102,241,0.1);
  cursor: pointer; position: relative;
}
.logo-icon svg { width: 38px; height: 38px; }
.sidebar-logo { font-size: 1.1rem; font-weight: 700; color: #e0e7ff; letter-spacing: -0.01em; }
.sidebar-sub { font-size: 0.68rem; color: #6366f1; margin-top: 1px; display: block; opacity: 0.7; }

.user-bar {
  display: flex; align-items: center; gap: 0.55rem;
  padding: 0.6rem 1.25rem;
  border-bottom: 1px solid rgba(99,102,241,0.08);
  position: relative;
}
.user-avatar {
  width: 28px; height: 28px; border-radius: 50%;
  background: rgba(99,102,241,0.2); color: #a5b4fc;
  display: flex; align-items: center; justify-content: center;
  font-size: 0.7rem; font-weight: 700; flex-shrink: 0;
}
.user-avatar.guest { background: rgba(148,163,184,0.1); color: #64748b; }
.user-name { font-size: 0.75rem; color: #cbd5e1; flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.user-name.dim { color: #64748b; }
.logout-btn {
  background: rgba(99,102,241,0.08); border: 1px solid rgba(99,102,241,0.15);
  color: #818cf8; padding: 0.2rem 0.55rem; border-radius: 5px;
  cursor: pointer; font-size: 0.65rem; font-weight: 500;
  transition: all var(--transition); font-family: inherit;
}
.logout-btn:hover { background: rgba(239,68,68,0.1); border-color: rgba(239,68,68,0.3); color: #f87171; }

.create-form {
  padding: 1.25rem;
  border-bottom: 1px solid rgba(99,102,241,0.08);
  position: relative;
}
.create-label {
  font-size: 0.7rem; font-weight: 600;
  text-transform: uppercase; letter-spacing: 0.06em;
  color: #818cf8; margin-bottom: 0.5rem; display: block;
  opacity: 0.8;
}
.create-input {
  width: 100%; padding: 0.7rem 0.8rem;
  border: 1px solid rgba(99,102,241,0.12);
  border-radius: var(--radius); background: rgba(99,102,241,0.04);
  color: #e2e8f0; font-size: 0.8rem; resize: none; outline: none;
  font-family: inherit; line-height: 1.5;
  transition: border-color var(--transition), box-shadow var(--transition);
}
.create-input::placeholder { color: #475569; }
.create-input:focus {
  border-color: rgba(129,140,248,0.4);
  box-shadow: 0 0 0 3px rgba(99,102,241,0.08);
}
.create-btn {
  width: 100%; margin-top: 0.625rem; padding: 0.6rem;
  border: none; border-radius: var(--radius);
  background: linear-gradient(135deg, #4f46e5, #7c3aed);
  color: #fff; font-size: 0.8rem; font-weight: 600;
  cursor: pointer; display: flex; align-items: center; justify-content: center;
  gap: 0.4rem;
  transition: transform var(--transition), box-shadow var(--transition);
}
.create-btn:hover { transform: translateY(-1px); box-shadow: 0 4px 16px rgba(99,102,241,0.3); }
.create-btn:active { transform: translateY(0); }
.create-btn:disabled { opacity: 0.5; cursor: not-allowed; transform: none; }
.create-btn.loading { pointer-events: none; }
.create-btn.busy { background: linear-gradient(135deg, #d97706, #b45309); }
.busy-hint { font-size: 0.63rem; color: #f59e0b; margin-top: 0.4rem; text-align: center; }
.btn-spinner {
  width: 15px; height: 15px;
  border: 2px solid rgba(255,255,255,0.3);
  border-top-color: #fff; border-radius: 50%;
  animation: spin 0.7s linear infinite;
}

.task-list { flex: 1; overflow-y: auto; padding: 0.75rem 1rem; position: relative; }
.task-list-title {
  font-size: 0.68rem; font-weight: 600;
  text-transform: uppercase; letter-spacing: 0.06em;
  color: #818cf8; margin-bottom: 0.6rem; padding: 0 0.25rem;
  display: flex; align-items: center; justify-content: space-between;
  opacity: 0.8;
}
.task-count {
  background: rgba(99,102,241,0.12); color: #818cf8;
  font-size: 0.6rem; padding: 0.1rem 0.4rem; border-radius: 999px;
}

.task-item {
  padding: 0.65rem 0.7rem; border-radius: var(--radius-sm);
  cursor: pointer; margin-bottom: 0.2rem;
  transition: background var(--transition), border-color var(--transition);
  border-left: 2px solid transparent;
}
.task-item:hover { background: rgba(99,102,241,0.06); }
.task-item.active {
  background: rgba(99,102,241,0.1);
  border-left-color: #818cf8;
}
.task-item-top { margin-bottom: 0.25rem; display: flex; justify-content: space-between; align-items: flex-start; gap: 0.3rem; }
.task-delete-btn {
  background: none; border: none; color: #4b5563; font-size: 0.85rem;
  cursor: pointer; padding: 0 0.1rem; line-height: 1;
  flex-shrink: 0; opacity: 0;
  transition: opacity var(--transition), color var(--transition);
}
.task-item:hover .task-delete-btn, .task-item.active .task-delete-btn { opacity: 1; }
.task-delete-btn:hover { color: #ef4444; }
.task-item-query { font-size: 0.78rem; color: #cbd5e1; line-height: 1.35; word-break: break-word; }
.task-item-meta { display: flex; align-items: center; gap: 0.5rem; }
.task-badge {
  font-size: 0.6rem; font-weight: 600;
  padding: 0.12rem 0.4rem; border-radius: 999px;
  display: flex; align-items: center; gap: 0.2rem;
}
.task-badge.running { background: rgba(59,130,246,0.15); color: #60a5fa; }
.task-badge.completed { background: rgba(16,185,129,0.12); color: #34d399; }
.task-badge.cancelled { background: rgba(251,191,36,0.12); color: #fbbf24; }
.task-badge.failed { background: rgba(239,68,68,0.12); color: #f87171; }
.badge-dot { width: 5px; height: 5px; border-radius: 50%; background: currentColor; }
.task-badge.running .badge-dot { animation: pulse 1.5s infinite; }
.task-item-time { font-size: 0.6rem; color: #475569; }
.task-item-progress { display: flex; align-items: center; gap: 0.4rem; margin-top: 0.35rem; }
.mini-bar { flex: 1; height: 3px; background: rgba(99,102,241,0.08); border-radius: 2px; overflow: hidden; }
.mini-bar-fill { height: 100%; background: #6366f1; border-radius: 2px; transition: width 0.6s; }
.mini-bar-fill.done { background: #10b981; }
.mini-count { font-size: 0.6rem; color: #64748b; min-width: 2.5em; text-align: right; }
.empty-hint { font-size: 0.75rem; color: #475569; padding: 0.5rem 0.25rem; line-height: 1.5; }

.task-list-enter-active { transition: all 0.35s cubic-bezier(0.4, 0, 0.2, 1); }
.task-list-leave-active { transition: all 0.2s ease-in; }
.task-list-enter-from { opacity: 0; transform: translateX(-20px); }
.task-list-leave-to { opacity: 0; transform: translateX(-20px); }

@keyframes spin { to { transform: rotate(360deg); } }
@keyframes pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.5; } }
</style>
