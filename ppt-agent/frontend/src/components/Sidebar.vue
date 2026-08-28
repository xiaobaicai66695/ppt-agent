<script setup lang="ts">
import { computed, ref } from 'vue';
import { Clock3, History, Plus, Trash2 } from 'lucide-vue-next';
import type { TaskInfo } from '../types';
import { summarizeTaskTitle } from '../utils/workbench';

const props = defineProps<{
  user: { id: number; email: string; is_admin?: boolean } | null;
  tasks: TaskInfo[];
  selectedId: string | null;
}>();

const emit = defineEmits<{
  logout: [];
  selectTask: [id: string];
  deleteTask: [id: string];
  compose: [];
  newSession: [];
}>();

const taskCount = computed(() => props.tasks.length);

function fmtTime(iso: string): string {
  if (!iso) return '';
  const date = new Date(iso);
  return date.toDateString() === new Date().toDateString()
    ? date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
    : `${date.getMonth() + 1}/${date.getDate()}`;
}

function fmtTokens(value: number): string {
  if (!value || value <= 0) return '';
  if (value >= 1_000_000) return `${(value / 1_000_000).toFixed(1)}M`;
  if (value >= 1_000) return `${(value / 1000).toFixed(1)}K`;
  return String(value);
}

function statusLabel(status: string): string {
  return ({ running: '运行中', completed: '已完成', cancelled: '已中断', failed: '失败' } as Record<string, string>)[status] || status;
}

</script>

<template>
  <aside class="task-sidebar" aria-label="任务工作区">
    <header class="task-sidebar-head">
      <span class="head-copy">
        <History :size="17" />
        <strong>任务</strong>
        <small>{{ taskCount }}</small>
      </span>
      <span class="head-actions">
        <button type="button" title="新建会话" aria-label="新建会话" @click="emit('newSession')">
          <Plus :size="18" />
        </button>
      </span>
    </header>

    <div class="task-list" role="list">
      <div v-if="tasks.length === 0" class="task-empty">
        <History :size="22" />
        <strong>还没有生成记录</strong>
        <span>使用工作台底部输入框创建演示</span>
      </div>

      <article
        v-for="task in tasks"
        :key="task.id"
        class="task-item"
        :class="{ active: task.id === selectedId }"
        role="listitem"
      >
        <button
          class="task-select"
          type="button"
          :aria-current="task.id === selectedId ? 'true' : undefined"
          @click="emit('selectTask', task.id)"
        >
          <strong :title="task.query || ''">{{ summarizeTaskTitle(task.query) }}</strong>
          <span class="task-meta">
            <span class="task-state" :class="task.status">
              <i aria-hidden="true"></i>{{ statusLabel(task.status) }}
            </span>
            <span><Clock3 :size="12" />{{ fmtTime(task.created_at) }}</span>
          </span>
          <span v-if="task.total_count > 0" class="mini-progress">
            <i :style="{ width: `${Math.round((task.done_count / task.total_count) * 100)}%` }"></i>
          </span>
          <small v-if="(task.total_tokens || 0) > 0">{{ fmtTokens(task.total_tokens || 0) }} tokens</small>
        </button>
        <button
          v-if="task.status !== 'running'"
          class="task-delete"
          type="button"
          :aria-label="`删除任务：${task.query || task.id}`"
          title="删除任务"
          @click="emit('deleteTask', task.id)"
        >
          <Trash2 :size="15" />
        </button>
      </article>
    </div>

  </aside>
</template>

<style scoped>
.task-sidebar { height: 100%; min-height: 0; display: flex; flex-direction: column; background: var(--surface-muted); border-right: 1px solid var(--border); }
.task-sidebar-head { min-height: 52px; padding: 0 12px 0 16px; display: flex; align-items: center; justify-content: space-between; border-bottom: 1px solid var(--border); }
.head-copy { display: flex; align-items: center; gap: 8px; color: var(--text-secondary); }
.head-copy strong { color: var(--text); font-size: 13px; }
.head-copy small { min-width: 20px; height: 20px; padding: 0 5px; display: grid; place-items: center; border-radius: 10px; color: var(--text-muted); background: var(--surface-pressed); font-size: 10px; }
.head-actions { display: flex; gap: 2px; }
.head-actions button { width: 40px; height: 40px; display: grid; place-items: center; border: 0; border-radius: 5px; color: var(--text-secondary); background: transparent; cursor: pointer; }
.head-actions button:hover { color: var(--text); background: var(--surface-hover); }

.task-list { min-height: 0; padding: 8px; display: grid; align-content: start; gap: 4px; overflow-y: auto; }
.task-empty { min-height: 150px; padding: 22px; display: grid; place-content: center; justify-items: center; gap: 5px; color: var(--text-muted); text-align: center; }
.task-empty strong { color: var(--text-secondary); font-size: 12px; }
.task-empty span { font-size: 11px; }
.task-item { position: relative; min-width: 0; border: 1px solid transparent; border-radius: 6px; }
.task-item:hover { background: var(--surface); }
.task-item.active { border-color: var(--border-strong); background: var(--surface); box-shadow: inset 3px 0 0 var(--action-ink); }
.task-select { width: 100%; min-height: 84px; padding: 10px 34px 10px 11px; display: flex; flex-direction: column; align-items: stretch; border: 0; color: inherit; background: transparent; text-align: left; cursor: pointer; }
.task-select strong { overflow: hidden; color: var(--text); font-size: 12px; font-weight: 650; line-height: 1.45; text-overflow: ellipsis; white-space: nowrap; }
.task-meta { margin-top: 7px; display: flex; align-items: center; justify-content: space-between; color: var(--text-muted); font-size: 10px; }
.task-meta > span { display: inline-flex; align-items: center; gap: 4px; }
.task-state i { width: 6px; height: 6px; border-radius: 50%; background: var(--text-disabled); }
.task-state.running { color: var(--info); }.task-state.running i { background: var(--info); animation: pulse 1.4s ease-in-out infinite; }
.task-state.completed { color: var(--success); }.task-state.completed i { background: var(--success); }
.task-state.failed { color: var(--danger); }.task-state.failed i { background: var(--danger); }
.mini-progress { height: 3px; margin-top: 9px; overflow: hidden; border-radius: 2px; background: var(--surface-pressed); }
.mini-progress i { display: block; height: 100%; background: var(--action-ink); }
.task-select > small { margin-top: 5px; color: var(--text-muted); font-size: 9px; }
.task-delete { position: absolute; top: 6px; right: 4px; width: 32px; height: 32px; display: grid; place-items: center; border: 0; border-radius: 4px; color: var(--text-muted); background: transparent; opacity: 0; cursor: pointer; }
.task-item:hover .task-delete, .task-delete:focus-visible { opacity: 1; }
.task-delete:hover { color: var(--danger); background: var(--danger-soft); }

@keyframes pulse { 50% { opacity: 0.4; } }
</style>
