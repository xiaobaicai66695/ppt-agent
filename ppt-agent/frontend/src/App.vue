<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue';
import {
  type TaskInfo,
  type TaskItem,
  type SSEEvent,
  STATUS_LABELS,
  STATUS_COLORS,
  fetchTasks,
  createTask,
  fetchTask,
} from './api';

// ---------------------------------------------------------------------------
// State
// ---------------------------------------------------------------------------

const tasks = ref<TaskInfo[]>([]);
const selectedId = ref<string | null>(null);
const query = ref('');
const creating = ref(false);

const taskItems = ref<TaskItem[]>([]);
const doneCount = ref(0);
const totalCount = ref(0);
const logLines = ref<{ ts: number; text: string; kind: string; html?: boolean }[]>([]);
const finalFiles = ref<string[]>([]);
const finalMessage = ref('');
const duration = ref('');
const logAutoScroll = ref(true);

// 当前正在并行执行的子任务数
const activeWorkers = ref(0);
const lastProgressTs = ref(0);

// 并行批次追踪：每个 SlideExecutor 派发创建一个批次面板
interface Batch {
  id: number;
  taskIds: string[];
  ts: number;
  done: boolean;
}
const batches = ref<Batch[]>([]);
let batchIdSeq = 0;

// 缩略图加载状态
const thumbLoaded = ref<Record<string, boolean>>({});
const thumbError = ref<Record<string, boolean>>({});

let es: EventSource | null = null;
const logBox = ref<HTMLElement | null>(null);

const selectedTask = computed(() => tasks.value.find(t => t.id === selectedId.value));
const progressPct = computed(() =>
  totalCount.value > 0 ? Math.round((doneCount.value / totalCount.value) * 100) : 0
);
const doneItems = computed(() => taskItems.value.filter(t => t.status === 'done' || t.status === 'qa_done' || t.status === 'fixed'));
const pendingItems = computed(() => taskItems.value.filter(t => t.status === 'pending' || t.status === 'generating'));
const failedItems = computed(() => taskItems.value.filter(t => t.status === 'failed'));
const generatingItems = computed(() => taskItems.value.filter(t => t.status === 'generating'));

// ---------------------------------------------------------------------------
// Markdown → HTML (lightweight)
// ---------------------------------------------------------------------------

function renderMd(text: string): string {
  let html = text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;');
  // headings
  html = html.replace(/^### (.+)$/gm, '<span class="md-h3">$1</span>');
  html = html.replace(/^## (.+)$/gm, '<span class="md-h2">$1</span>');
  html = html.replace(/^# (.+)$/gm, '<span class="md-h1">$1</span>');
  // bold
  html = html.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>');
  // inline code
  html = html.replace(/`([^`]+)`/g, '<code class="md-code">$1</code>');
  // horizontal rules
  html = html.replace(/^---$/gm, '<span class="md-hr"></span>');
  // file paths (heuristically)
  html = html.replace(/((?:\/[\w.-]+)+\.\w+)/g, '<code class="md-path">$1</code>');
  return html;
}

// ---------------------------------------------------------------------------
// Parse task IDs from tool call arguments
// ---------------------------------------------------------------------------

function parseTaskIds(args: string): string[] {
  try {
    const obj = JSON.parse(args);
    // Try "task_id" or "task_ids" field (comma-separated string or array)
    if (Array.isArray(obj.task_ids)) return obj.task_ids.map(String);
    if (typeof obj.task_id === 'string') return obj.task_id.split(',').map((s: string) => s.trim());
    if (typeof obj.task_ids === 'string') return obj.task_ids.split(',').map((s: string) => s.trim());
    return [];
  } catch {
    return [];
  }
}

// 将 task card 标记为 generating
function markGenerating(ids: string[]) {
  for (const item of taskItems.value) {
    if (ids.includes(item.task_id) && item.status === 'pending') {
      item.status = 'generating';
    }
  }
}

// ---------------------------------------------------------------------------
// Log
// ---------------------------------------------------------------------------

function addLog(kind: string, text: string) {
  if (kind === 'answer' && logLines.value.length > 0) {
    const last = logLines.value[logLines.value.length - 1];
    if (last.kind === 'answer') {
      last.text += text;
      last.ts = Date.now();
      if (logAutoScroll.value) {
        nextTick(() => {
          if (logBox.value) logBox.value.scrollTop = logBox.value.scrollHeight;
        });
      }
      return;
    }
  }
  logLines.value = [...logLines.value.slice(-500), { ts: Date.now(), text, kind }];
  if (logAutoScroll.value) {
    nextTick(() => {
      if (logBox.value) logBox.value.scrollTop = logBox.value.scrollHeight;
    });
  }
}

// ---------------------------------------------------------------------------
// SSE
// ---------------------------------------------------------------------------

function connectSSE(taskId: string) {
  if (es) es.close();
  es = new EventSource(`/api/tasks/${taskId}/stream`);
  activeWorkers.value = 0;

  const handler = (e: MessageEvent) => {
    let evt: SSEEvent;
    try { evt = JSON.parse(e.data); } catch { return; }

    switch (evt.type) {
      case 'answer':
        if (logLines.value.length > 0 && logLines.value[logLines.value.length - 1].kind !== 'answer') {
          addLog('divider', '── AI 响应 ──');
        }
        addLog('answer', evt.content || '');
        break;

      case 'tool_call': {
        const name = evt.tool_name || '';
        const args = evt.tool_args || '';

        // SlideExecutor dispatch → 标记对应卡片为 generating，创建批次面板
        if (name === 'task' && args.includes('SlideExecutor')) {
          const ids = parseTaskIds(args);
          markGenerating(ids);
          activeWorkers.value = Math.max(1, ids.length || 1);
          batches.value = [...batches.value, { id: ++batchIdSeq, taskIds: ids, ts: Date.now(), done: false }];
          addLog('worker', `⚡ 派发 ${ids.length} 个子任务并行执行`);
        } else if (name === 'task') {
          const ids = parseTaskIds(args);
          markGenerating(ids);
          activeWorkers.value = Math.max(activeWorkers.value, ids.length || 1);
          addLog('tool', `▶ ${name} (${args.slice(0, 100)})`);
        } else {
          addLog('tool', `▶ ${name} (${args.slice(0, 120)})`);
        }
        lastProgressTs.value = Date.now();
        break;
      }

      case 'progress':
        if (evt.tasks) {
          taskItems.value = evt.tasks;
          // 更新批次状态
          for (const batch of batches.value) {
            if (batch.done) continue;
            const allDone = batch.taskIds.every(tid => {
              const item = evt.tasks!.find(t => t.task_id === tid);
              return item && (item.status === 'done' || item.status === 'qa_done' || item.status === 'fixed');
            });
            if (allDone) batch.done = true;
          }
          // 双保险：从 progress 事件中提取已完成文件的 output_file
          for (const t of evt.tasks) {
            if ((t.status === 'done' || t.status === 'qa_done' || t.status === 'fixed') && t.output_file) {
              if (!finalFiles.value.includes(t.output_file)) {
                finalFiles.value = [...finalFiles.value, t.output_file];
                addLog('file', `📄 ${t.output_file} 已生成 → 可下载`);
              }
            }
          }
        }
        if (evt.done !== undefined) doneCount.value = evt.done;
        if (evt.total !== undefined) totalCount.value = evt.total;
        lastProgressTs.value = Date.now();
        if (Date.now() - lastProgressTs.value > 8000) {
          activeWorkers.value = 0;
        }
        break;

      case 'file_ready':
        if (evt.files) {
          for (const f of evt.files) {
            if (!finalFiles.value.includes(f)) {
              finalFiles.value = [...finalFiles.value, f];
              addLog('file', `📄 ${f} 已生成 → 可下载`);
            }
            cacheFileInBrowser(selectedTask.value!.id, f);
          }
        }
        break;

      case 'error':
        addLog('error', evt.error || evt.content || '');
        break;

      case 'complete':
        doneCount.value = evt.done || 0;
        totalCount.value = evt.total || 0;
        if (evt.files) finalFiles.value = evt.files;
        if (evt.message) finalMessage.value = evt.message;
        if (evt.duration) duration.value = evt.duration;
        if (evt.tasks) taskItems.value = evt.tasks;
        activeWorkers.value = 0;
        es!.close();
        refreshTask(taskId);
        break;
    }
  };

  es.addEventListener('answer', handler);
  es.addEventListener('tool_call', handler);
  es.addEventListener('progress', handler);
  es.addEventListener('file_ready', handler);
  es.addEventListener('error', handler);
  es.addEventListener('complete', handler);
}

// ---------------------------------------------------------------------------
// Task actions
// ---------------------------------------------------------------------------

async function refreshTask(id: string) {
  try {
    const info = await fetchTask(id);
    const idx = tasks.value.findIndex(t => t.id === id);
    if (idx >= 0) tasks.value[idx] = info;
  } catch { /* ignore */ }
}

async function handleCreate() {
  const q = query.value.trim();
  if (!q || creating.value) return;
  creating.value = true;

  try {
    const info = await createTask(q);
    tasks.value = [info, ...tasks.value];
    selectTask(info.id);
    query.value = '';
  } catch (err) {
    addLog('error', '创建任务失败: ' + (err as Error).message);
  } finally {
    creating.value = false;
  }
}

function selectTask(id: string) {
  selectedId.value = id;
  const t = tasks.value.find(x => x.id === id);
  if (!t) return;

  taskItems.value = [];
  doneCount.value = t.done_count;
  totalCount.value = t.total_count;
  logLines.value = [];
  finalFiles.value = t.files || [];
  finalMessage.value = '';
  duration.value = t.duration || '';
  activeWorkers.value = 0;
  batches.value = [];
  thumbLoaded.value = {};
  thumbError.value = {};

  if (t.status === 'running') {
    connectSSE(id);
  } else {
    refreshTask(id);
  }
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault();
    handleCreate();
  }
}

function onLogScroll() {
  const el = logBox.value;
  if (!el) return;
  logAutoScroll.value = el.scrollHeight - el.scrollTop - el.clientHeight < 30;
}

function fmtTime(iso: string): string {
  if (!iso) return '';
  const d = new Date(iso);
  return d.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });
}

async function cacheFileInBrowser(taskId: string, filename: string) {
  try {
    const cache = await caches.open('ppt-agent-files');
    const url = `/api/tasks/${taskId}/files/${encodeURIComponent(filename)}`;
    const cached = await cache.match(url);
    if (cached) return;
    // 后台预缓存，不阻塞
    const res = await fetch(url);
    if (res.ok) await cache.put(url, res.clone());
  } catch { /* 静默失败，缓存不是必需的 */ }
}

function getThumbUrl(taskId: string, filename: string): string {
  const name = filename.split(/[/\\]/).pop() || filename;
  return `/api/tasks/${taskId}/thumb/${encodeURIComponent(name)}`;
}

function onThumbLoaded(filename: string) {
  thumbLoaded.value = { ...thumbLoaded.value, [filename]: true };
}
function onThumbError(filename: string) {
  thumbError.value = { ...thumbError.value, [filename]: true };
}

onMounted(async () => {
  try { tasks.value = await fetchTasks(); } catch { /* noop */ }
});

onUnmounted(() => {
  if (es) es.close();
});
</script>

<template>
  <div class="layout">
    <!-- ================================================================== -->
    <!-- SIDEBAR                                                             -->
    <!-- ================================================================== -->
    <aside class="sidebar">
      <div class="sidebar-header">
        <div class="logo-icon">
          <svg viewBox="0 0 40 40" fill="none">
            <rect x="4" y="6" width="32" height="24" rx="3" fill="rgba(255,255,255,0.15)" />
            <rect x="4" y="6" width="32" height="5" rx="3" fill="rgba(255,255,255,0.3)" />
            <rect x="10" y="14" width="14" height="2" rx="1" fill="rgba(255,255,255,0.2)" />
            <rect x="10" y="18" width="20" height="2" rx="1" fill="rgba(255,255,255,0.15)" />
            <rect x="10" y="22" width="16" height="2" rx="1" fill="rgba(255,255,255,0.12)" />
          </svg>
        </div>
        <div>
          <h1 class="sidebar-logo">PPT Agent</h1>
          <span class="sidebar-sub">AI 驱动的幻灯片生成</span>
        </div>
      </div>

      <div class="create-form">
        <label class="create-label">新建任务</label>
        <div class="create-input-wrap">
          <textarea
            class="create-input"
            placeholder="描述你的 PPT 需求，例如：做一个关于新能源汽车的行业分析报告..."
            v-model="query"
            @keydown="handleKeydown"
            rows="3"
          ></textarea>
        </div>
        <button
          class="create-btn"
          :class="{ loading: creating }"
          :disabled="creating"
          @click="handleCreate"
        >
          <span v-if="creating" class="btn-spinner"></span>
          <span>{{ creating ? '创建中...' : '✦ 生成 PPT' }}</span>
        </button>
      </div>

      <div class="task-list">
        <h3 class="task-list-title">
          任务历史
          <span v-if="tasks.length" class="task-count">{{ tasks.length }}</span>
        </h3>
        <p v-if="tasks.length === 0" class="empty-hint">暂无任务</p>
        <TransitionGroup name="task-list" tag="div">
          <div
            v-for="t in tasks"
            :key="t.id"
            class="task-item"
            :class="{ active: t.id === selectedId }"
            @click="selectTask(t.id)"
          >
            <div class="task-item-top">
              <span class="task-item-query">{{ t.query.length > 38 ? t.query.slice(0, 38) + '…' : t.query }}</span>
            </div>
            <div class="task-item-meta">
              <span class="task-badge" :class="t.status">
                <span class="badge-dot"></span>
                {{ t.status === 'running' ? '运行中' : t.status === 'completed' ? '已完成' : '失败' }}
              </span>
              <span class="task-item-time">{{ fmtTime(t.created_at) }}</span>
            </div>
            <div v-if="t.total_count > 0" class="task-item-progress">
              <div class="mini-bar">
                <div
                  class="mini-bar-fill"
                  :class="{ done: t.status === 'completed' }"
                  :style="{ width: Math.round((t.done_count / t.total_count) * 100) + '%' }"
                />
              </div>
              <span class="mini-count">{{ t.done_count }}/{{ t.total_count }}</span>
            </div>
          </div>
        </TransitionGroup>
      </div>
    </aside>

    <!-- ================================================================== -->
    <!-- MAIN                                                                -->
    <!-- ================================================================== -->
    <main class="main">
      <Transition name="view" mode="out-in">
        <!-- Welcome -->
        <div v-if="!selectedTask" key="welcome" class="welcome">
          <div class="welcome-hero">
            <div class="welcome-icon">
              <svg viewBox="0 0 120 120" fill="none">
                <rect x="12" y="18" width="96" height="72" rx="8" fill="#e0e7ff" stroke="#c7d2fe" stroke-width="2"/>
                <rect x="12" y="18" width="96" height="16" rx="8" fill="#a5b4fc"/>
                <circle cx="28" cy="26" r="3" fill="#6366f1"/>
                <circle cx="38" cy="26" r="3" fill="#818cf8"/>
                <circle cx="48" cy="26" r="3" fill="#a5b4fc"/>
                <rect x="22" y="44" width="40" height="5" rx="2.5" fill="#c7d2fe"/>
                <rect x="22" y="54" width="56" height="5" rx="2.5" fill="#e0e7ff"/>
                <rect x="22" y="64" width="48" height="5" rx="2.5" fill="#e0e7ff"/>
                <rect x="22" y="74" width="32" height="5" rx="2.5" fill="#e0e7ff"/>
                <path d="M90 30l2 5 5 1-5 1-2 5-2-5-5-1 5-1 2-5z" fill="#6366f1"/>
                <path d="M100 72l1 3 3 1-3 1-1 3-1-3-3-1 3-1 1-3z" fill="#a5b4fc"/>
              </svg>
            </div>
            <h2>开始创建演示文稿</h2>
            <p>在左侧描述你的需求，AI 将自动规划、生成、质检每一页幻灯片</p>
          </div>
          <div class="welcome-hints">
            <div class="hint-card" style="animation-delay: 0.1s">
              <div class="hint-icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"/></svg></div>
              <strong>并行生成</strong>
              <span>多幻灯片并行处理，效率提升 3-5 倍</span>
            </div>
            <div class="hint-card" style="animation-delay: 0.2s">
              <div class="hint-icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/></svg></div>
              <strong>自动质检</strong>
              <span>生成后自动进行视觉质量审查</span>
            </div>
            <div class="hint-card" style="animation-delay: 0.3s">
              <div class="hint-icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2"/><line x1="7" y1="8" x2="17" y2="8"/><line x1="7" y1="12" x2="17" y2="12"/><line x1="7" y1="16" x2="12" y2="16"/></svg></div>
              <strong>实时追踪</strong>
              <span>每页状态实时可见，进度一目了然</span>
            </div>
          </div>
        </div>

        <!-- Dashboard -->
        <div v-else key="dashboard" class="dashboard">
          <!-- Header -->
          <div class="dash-header">
            <div class="dash-header-left">
              <h2 class="dash-title">{{ selectedTask!.query }}</h2>
              <div class="dash-meta-row">
                <span class="dash-id">{{ selectedTask!.id.slice(0, 8) }}</span>
                <span v-if="duration" class="dash-duration">
                  <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="8" cy="8" r="7"/><path d="M8 4v4l3 2"/></svg>
                  {{ duration }}
                </span>
              </div>
            </div>
            <div class="dash-stats">
              <span v-if="activeWorkers > 0 && selectedTask!.status === 'running'" class="stat-badge workers">
                <span class="pulse-dot"></span>{{ activeWorkers }} 并行执行中
              </span>
              <span v-if="selectedTask!.status === 'running'" class="stat-badge running">
                <span class="pulse-dot"></span>运行中
              </span>
              <span v-if="selectedTask!.status === 'completed'" class="stat-badge done">✓ 已完成</span>
              <span v-if="selectedTask!.status === 'failed'" class="stat-badge failed">✗ 失败</span>
            </div>
          </div>

          <!-- Progress -->
          <Transition name="slide-down">
            <div v-if="totalCount > 0" class="progress-section">
              <div class="progress-header">
                <span>生成进度</span>
                <span class="progress-num">
                  <strong>{{ doneCount }}</strong> / {{ totalCount }} 页
                  <span class="progress-pct">({{ progressPct }}%)</span>
                </span>
              </div>
              <div class="progress-track">
                <div class="progress-fill" :class="{ shimmer: selectedTask!.status === 'running' }" :style="{ width: progressPct + '%' }" />
              </div>
              <div class="progress-legend">
                <span class="legend-item"><span class="legend-dot done"></span>已完成 {{ doneItems.length }}</span>
                <span class="legend-item"><span class="legend-dot pending"></span>处理中 {{ pendingItems.length }}</span>
                <span v-if="generatingItems.length" class="legend-item active-workers">
                  <span class="legend-dot generating"></span>并行生成 {{ generatingItems.length }}
                </span>
                <span v-if="failedItems.length" class="legend-item"><span class="legend-dot failed"></span>失败 {{ failedItems.length }}</span>
              </div>
            </div>
          </Transition>

          <!-- Parallel Batch Panels -->
          <TransitionGroup name="slide-down" tag="div">
            <div v-for="batch in batches" :key="batch.id" class="batch-panel" :class="{ done: batch.done }">
              <div class="batch-header">
                <span class="batch-label">
                  <span class="batch-spinner" v-if="!batch.done"></span>
                  {{ batch.done ? '✓' : '⚡' }} 批次 #{{ batch.id }}
                </span>
                <span class="batch-meta">
                  {{ batch.done ? '已完成' : '执行中' }}
                  · {{ batch.taskIds.length }} 页
                </span>
              </div>
              <div class="batch-cards">
                <div
                  v-for="tid in batch.taskIds"
                  :key="tid"
                  class="batch-card"
                  :class="(taskItems.find(t => t.task_id === tid) || {}).status || 'pending'"
                >
                  <div class="bcard-idx">{{ (taskItems.find(t => t.task_id === tid) || {}).page_index || '?' }}</div>
                  <div class="bcard-body">
                    <div class="bcard-title">{{ (taskItems.find(t => t.task_id === tid) || {}).title || 'task_id=' + tid }}</div>
                    <div class="bcard-status" :class="(taskItems.find(t => t.task_id === tid) || {}).status || 'pending'">
                      {{ STATUS_LABELS[(taskItems.find(t => t.task_id === tid) || {}).status || 'pending'] }}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </TransitionGroup>

          <!-- Task Cards -->
          <Transition name="slide-down">
            <div v-if="taskItems.length > 0" class="cards-section">
              <h3 class="section-title">页面详情</h3>
              <TransitionGroup name="card" tag="div" class="cards-grid">
                <div
                  v-for="(item, idx) in taskItems"
                  :key="item.task_id"
                  class="task-card"
                  :class="{ generating: item.status === 'generating' }"
                  :style="{ animationDelay: Math.min(idx * 0.02, 0.6) + 's' }"
                >
                  <div class="card-status" :class="item.status" :style="(item.status === 'done' || item.status === 'qa_done' || item.status === 'fixed' || item.status === 'failed' || item.status === 'generating') ? { background: STATUS_COLORS[item.status] } : {}">
                    <Transition name="pop" mode="out-in">
                      <span v-if="item.status === 'done' || item.status === 'qa_done' || item.status === 'fixed'" key="check">✓</span>
                      <span v-else-if="item.status === 'failed'" key="fail">✗</span>
                      <span v-else-if="item.status === 'generating'" key="spin" class="card-spinner"></span>
                      <span v-else key="num">{{ item.page_index }}</span>
                    </Transition>
                  </div>
                  <div class="card-body">
                    <div class="card-title">{{ item.title }}</div>
                    <div class="card-meta">
                      <span class="card-type">{{ item.content_type }}</span>
                      <span class="card-status-label" :class="item.status">
                        <span class="status-dot-sm"></span>
                        {{ STATUS_LABELS[item.status] || item.status }}
                      </span>
                    </div>
                    <div v-if="item.output_file" class="card-file">{{ item.output_file }}</div>
                  </div>
                </div>
              </TransitionGroup>
            </div>
          </Transition>

          <!-- Event Log -->
          <div class="log-section">
            <h3 class="section-title" @click="logAutoScroll = !logAutoScroll" style="cursor:pointer">
              事件日志
              <span v-if="logLines.length" class="log-count">{{ logLines.length }}</span>
              <span class="log-scroll-hint">{{ logAutoScroll ? '(自动滚动)' : '(已暂停)' }}</span>
            </h3>
            <div ref="logBox" class="log-box" @scroll="onLogScroll">
              <div v-if="logLines.length === 0" class="log-empty">
                <span class="log-empty-icon">●</span>
                <span>等待事件...</span>
              </div>
              <TransitionGroup v-else name="log" tag="div">
                <div v-for="(l, i) in logLines" :key="i" class="log-line" :class="l.kind">
                  <template v-if="l.kind === 'divider'">
                    <span class="log-text">{{ l.text }}</span>
                  </template>
                  <template v-else-if="l.kind === 'worker'">
                    <span class="log-ts">{{ new Date(l.ts).toLocaleTimeString('zh-CN') }}</span>
                    <span class="log-dot worker"></span>
                    <span class="log-text">{{ l.text }}</span>
                  </template>
                  <template v-else>
                    <span class="log-ts">{{ new Date(l.ts).toLocaleTimeString('zh-CN') }}</span>
                    <span class="log-dot"></span>
                    <span class="log-text" v-if="l.kind === 'answer'" v-html="renderMd(l.text)"></span>
                    <span class="log-text" v-else>{{ l.text }}</span>
                  </template>
                </div>
              </TransitionGroup>
            </div>
          </div>

          <!-- Final Message -->
          <Transition name="fade-up">
            <div v-if="finalMessage && selectedTask!.status !== 'running'" class="final-message">
              <div class="final-header">
                <span class="final-icon">✓</span>
                <h3>任务完成</h3>
              </div>
              <p>{{ finalMessage }}</p>
            </div>
          </Transition>

          <!-- File Downloads -->
          <Transition name="fade-up">
            <div v-if="finalFiles.length > 0" class="files-section">
              <h3 class="section-title">生成文件 ({{ finalFiles.length }})</h3>
              <TransitionGroup name="file" tag="div" class="files-grid">
                <a
                  v-for="f in finalFiles"
                  :key="f"
                  class="file-card"
                  :href="`/api/tasks/${selectedTask!.id}/files/${encodeURIComponent(f.split(/[/\\]/).pop() || '')}`"
                  download
                >
                  <div class="file-thumb">
                    <img
                      v-if="!thumbError[f]"
                      :src="getThumbUrl(selectedTask!.id, f)"
                      :class="{ loaded: thumbLoaded[f] }"
                      @load="onThumbLoaded(f)"
                      @error="onThumbError(f)"
                      loading="lazy"
                      alt=""
                    />
                    <span v-if="!thumbLoaded[f] && !thumbError[f]" class="thumb-spinner"></span>
                    <span v-if="thumbError[f]" class="file-icon">
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="12" y1="18" x2="12" y2="12"/><polyline points="9 15 12 18 15 15"/></svg>
                    </span>
                  </div>
                  <span class="file-name">{{ f.split(/[/\\]/).pop() }}</span>
                  <span class="file-arrow">↓</span>
                </a>
              </TransitionGroup>
            </div>
          </Transition>
        </div>
      </Transition>
    </main>
  </div>
</template>

<style>
/* ==================================================================== */
/* Reset & Variables                                                    */
/* ==================================================================== */
*, *::before, *::after { margin: 0; padding: 0; box-sizing: border-box; }

:root {
  --sidebar-w: 340px;
  --c-bg: #f8fafc;
  --c-surface: #ffffff;
  --c-sidebar: #0b1121;
  --c-sidebar-hover: rgba(255,255,255,0.06);
  --c-sidebar-active: rgba(59,130,246,0.15);
  --c-primary: #3b82f6;
  --c-primary-light: #eff6ff;
  --c-success: #10b981;
  --c-success-light: #ecfdf5;
  --c-danger: #ef4444;
  --c-danger-light: #fef2f2;
  --c-warning: #f59e0b;
  --c-warning-light: #fffbeb;
  --c-text: #0f172a;
  --c-text-2: #334155;
  --c-text-muted: #94a3b8;
  --c-border: #e2e8f0;
  --c-border-light: #f1f5f9;
  --radius-sm: 6px;
  --radius: 10px;
  --radius-lg: 14px;
  --shadow-sm: 0 1px 2px rgba(0,0,0,0.04);
  --shadow: 0 1px 3px rgba(0,0,0,0.06), 0 1px 2px rgba(0,0,0,0.04);
  --shadow-md: 0 4px 12px rgba(0,0,0,0.06), 0 2px 4px rgba(0,0,0,0.04);
  --shadow-lg: 0 10px 30px rgba(0,0,0,0.08), 0 4px 10px rgba(0,0,0,0.04);
  --transition: 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  --transition-slow: 0.4s cubic-bezier(0.4, 0, 0.2, 1);
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', sans-serif;
  background: var(--c-bg);
  color: var(--c-text);
  line-height: 1.5;
  -webkit-font-smoothing: antialiased;
}

/* ==================================================================== */
/* Keyframes                                                            */
/* ==================================================================== */
@keyframes fadeInUp {
  from { opacity: 0; transform: translateY(16px); }
  to   { opacity: 1; transform: translateY(0); }
}
@keyframes fadeIn {
  from { opacity: 0; }
  to   { opacity: 1; }
}
@keyframes slideDown {
  from { opacity: 0; transform: translateY(-8px); max-height: 0; }
  to   { opacity: 1; transform: translateY(0); max-height: 600px; }
}
@keyframes shimmer {
  0%   { background-position: -200% 0; }
  100% { background-position: 200% 0; }
}
@keyframes pulse {
  0%, 100% { opacity: 1; }
  50%      { opacity: 0.5; }
}
@keyframes pulseScale {
  0%, 100% { transform: scale(1); }
  50%      { transform: scale(1.08); }
}
@keyframes spin {
  to { transform: rotate(360deg); }
}
@keyframes popIn {
  0%   { opacity: 0; transform: scale(0.7); }
  70%  { transform: scale(1.08); }
  100% { opacity: 1; transform: scale(1); }
}
@keyframes borderPulse {
  0%, 100% { border-color: var(--c-primary); box-shadow: 0 0 0 2px rgba(59,130,246,0.2); }
  50%      { border-color: #818cf8; box-shadow: 0 0 0 4px rgba(99,102,241,0.15); }
}

/* ==================================================================== */
/* Layout                                                               */
/* ==================================================================== */
.layout { display: flex; min-height: 100vh; }

.sidebar {
  width: var(--sidebar-w); min-width: var(--sidebar-w);
  background: var(--c-sidebar); color: #e2e8f0;
  display: flex; flex-direction: column;
  height: 100vh; position: sticky; top: 0;
  overflow-y: auto; overflow-x: hidden;
}

.sidebar-header {
  padding: 1.25rem 1.25rem;
  display: flex; align-items: center; gap: 0.75rem;
  border-bottom: 1px solid rgba(255,255,255,0.06);
}
.logo-icon svg { width: 40px; height: 40px; }
.sidebar-logo { font-size: 1.15rem; font-weight: 700; color: #f1f5f9; letter-spacing: -0.01em; }
.sidebar-sub { font-size: 0.7rem; color: #64748b; margin-top: 1px; display: block; }

/* Create Form */
.create-form { padding: 1.25rem; border-bottom: 1px solid rgba(255,255,255,0.06); }
.create-label { font-size: 0.7rem; font-weight: 600; text-transform: uppercase; letter-spacing: 0.06em; color: #64748b; margin-bottom: 0.5rem; display: block; }
.create-input-wrap { position: relative; }
.create-input {
  width: 100%; padding: 0.75rem;
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: var(--radius); background: rgba(255,255,255,0.04);
  color: #e2e8f0; font-size: 0.8125rem; resize: none; outline: none;
  font-family: inherit; line-height: 1.5;
  transition: border-color var(--transition), box-shadow var(--transition);
}
.create-input::placeholder { color: #475569; }
.create-input:focus {
  border-color: rgba(59,130,246,0.5);
  box-shadow: 0 0 0 3px rgba(59,130,246,0.1);
}
.create-btn {
  width: 100%; margin-top: 0.625rem; padding: 0.65rem;
  border: none; border-radius: var(--radius);
  background: linear-gradient(135deg, #3b82f6, #6366f1);
  color: #fff; font-size: 0.8125rem; font-weight: 600;
  cursor: pointer; display: flex; align-items: center; justify-content: center;
  gap: 0.4rem; position: relative; overflow: hidden;
  transition: transform var(--transition), box-shadow var(--transition);
}
.create-btn::after {
  content: ''; position: absolute; inset: 0;
  background: linear-gradient(135deg, transparent 40%, rgba(255,255,255,0.15));
  transition: opacity var(--transition);
}
.create-btn:hover { transform: translateY(-1px); box-shadow: 0 4px 12px rgba(59,130,246,0.35); }
.create-btn:active { transform: translateY(0); }
.create-btn:disabled { opacity: 0.6; cursor: not-allowed; transform: none; }
.create-btn.loading { pointer-events: none; }
.btn-spinner {
  width: 16px; height: 16px; border: 2px solid rgba(255,255,255,0.3);
  border-top-color: #fff; border-radius: 50%; animation: spin 0.7s linear infinite;
}

/* Task List */
.task-list { flex: 1; overflow-y: auto; padding: 0.75rem 1rem; }
.task-list-title { font-size: 0.7rem; font-weight: 600; text-transform: uppercase; letter-spacing: 0.06em; color: #64748b; margin-bottom: 0.6rem; padding: 0 0.25rem; display: flex; align-items: center; justify-content: space-between; }
.task-count { background: rgba(255,255,255,0.08); font-size: 0.6rem; padding: 0.1rem 0.4rem; border-radius: 999px; }
.task-item {
  padding: 0.7rem 0.75rem; border-radius: var(--radius-sm);
  cursor: pointer; margin-bottom: 0.25rem;
  transition: background var(--transition), transform var(--transition);
}
.task-item:hover { background: var(--c-sidebar-hover); }
.task-item.active { background: var(--c-sidebar-active); border-left: 2px solid var(--c-primary); padding-left: calc(0.75rem - 2px); }
.task-item-top { margin-bottom: 0.3rem; }
.task-item-query { font-size: 0.8rem; color: #cbd5e1; line-height: 1.35; }
.task-item-meta { display: flex; align-items: center; gap: 0.5rem; }
.task-badge {
  font-size: 0.6rem; font-weight: 600; padding: 0.15rem 0.45rem;
  border-radius: 999px; display: flex; align-items: center; gap: 0.25rem;
}
.task-badge.running { background: rgba(59,130,246,0.15); color: #60a5fa; }
.task-badge.completed { background: rgba(16,185,129,0.15); color: #34d399; }
.task-badge.failed { background: rgba(239,68,68,0.15); color: #f87171; }
.badge-dot { width: 5px; height: 5px; border-radius: 50%; background: currentColor; }
.task-badge.running .badge-dot { animation: pulse 1.5s infinite; }
.task-item-time { font-size: 0.6rem; color: #475569; }
.task-item-progress { display: flex; align-items: center; gap: 0.4rem; margin-top: 0.4rem; }
.mini-bar { flex: 1; height: 3px; background: rgba(255,255,255,0.08); border-radius: 2px; overflow: hidden; }
.mini-bar-fill { height: 100%; background: #3b82f6; border-radius: 2px; transition: width 0.6s cubic-bezier(0.4, 0, 0.2, 1); }
.mini-bar-fill.done { background: #10b981; }
.mini-count { font-size: 0.6rem; color: #64748b; min-width: 2.5em; text-align: right; }

/* Task list transition */
.task-list-enter-active { transition: all 0.35s cubic-bezier(0.4, 0, 0.2, 1); }
.task-list-leave-active { transition: all 0.2s ease-in; }
.task-list-enter-from { opacity: 0; transform: translateX(-20px); }
.task-list-leave-to { opacity: 0; transform: translateX(-20px); }

/* ==================================================================== */
/* Main                                                                 */
/* ==================================================================== */
.main { flex: 1; padding: 2rem 2.5rem; overflow-y: auto; max-height: 100vh; }

/* View transition */
.view-enter-active { transition: all 0.35s cubic-bezier(0.4, 0, 0.2, 1); }
.view-leave-active { transition: all 0.2s ease-in; }
.view-enter-from { opacity: 0; transform: translateY(12px); }
.view-leave-to { opacity: 0; transform: translateY(-8px); }

/* ==================================================================== */
/* Welcome                                                              */
/* ==================================================================== */
.welcome {
  display: flex; flex-direction: column; align-items: center;
  justify-content: center; min-height: 70vh; text-align: center;
}
.welcome-hero { margin-bottom: 2.5rem; animation: fadeInUp 0.6s cubic-bezier(0.4, 0, 0.2, 1); }
.welcome-icon svg { width: 120px; height: 120px; margin: 0 auto 1.5rem; }
.welcome h2 {
  font-size: 1.75rem; font-weight: 700; margin-bottom: 0.5rem;
  background: linear-gradient(135deg, #1e293b, #475569);
  -webkit-background-clip: text; -webkit-text-fill-color: transparent;
}
.welcome > p { color: var(--c-text-muted); font-size: 0.95rem; }

.welcome-hints { display: grid; grid-template-columns: repeat(3, 1fr); gap: 1rem; max-width: 680px; width: 100%; }
.hint-card {
  background: var(--c-surface); border: 1px solid var(--c-border);
  border-radius: var(--radius-lg); padding: 1.5rem 1rem; text-align: center;
  transition: transform var(--transition), box-shadow var(--transition);
  animation: fadeInUp 0.5s cubic-bezier(0.4, 0, 0.2, 1) both;
}
.hint-card:hover { transform: translateY(-4px); box-shadow: var(--shadow-lg); }
.hint-icon {
  width: 40px; height: 40px; margin: 0 auto 0.75rem;
  background: var(--c-primary-light); color: var(--c-primary);
  border-radius: 10px; display: flex; align-items: center; justify-content: center;
}
.hint-icon svg { width: 20px; height: 20px; }
.hint-card strong { display: block; margin-bottom: 0.3rem; font-size: 0.9rem; color: var(--c-text); }
.hint-card span { font-size: 0.8rem; color: var(--c-text-muted); line-height: 1.4; }

/* ==================================================================== */
/* Dashboard                                                            */
/* ==================================================================== */
.dashboard { max-width: 1100px; }

.dash-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 1.5rem; animation: fadeInUp 0.4s both; }
.dash-header-left { flex: 1; min-width: 0; }
.dash-title { font-size: 1.2rem; font-weight: 650; line-height: 1.4; }
.dash-meta-row { display: flex; align-items: center; gap: 1rem; margin-top: 0.35rem; }
.dash-id { font-size: 0.7rem; color: var(--c-text-muted); font-family: 'SF Mono', monospace; background: var(--c-border-light); padding: 0.15rem 0.4rem; border-radius: 4px; }
.dash-duration { font-size: 0.75rem; color: var(--c-text-muted); display: flex; align-items: center; gap: 0.25rem; }
.dash-duration svg { width: 13px; height: 13px; }
.dash-stats { display: flex; align-items: center; gap: 0.75rem; flex-shrink: 0; flex-wrap: wrap; }
.stat-badge { font-size: 0.78rem; font-weight: 600; padding: 0.35rem 0.8rem; border-radius: 999px; display: flex; align-items: center; gap: 0.35rem; }
.stat-badge.running { background: var(--c-primary-light); color: #1d4ed8; }
.stat-badge.workers { background: var(--c-warning-light); color: #b45309; animation: borderPulse 2s infinite; }
.stat-badge.done { background: var(--c-success-light); color: #059669; }
.stat-badge.failed { background: var(--c-danger-light); color: #b91c1c; }
.pulse-dot { width: 7px; height: 7px; border-radius: 50%; background: currentColor; animation: pulse 1.5s infinite; }

/* Progress */
.progress-section {
  background: var(--c-surface); border: 1px solid var(--c-border);
  border-radius: var(--radius-lg); padding: 1rem 1.25rem; margin-bottom: 1.25rem;
  box-shadow: var(--shadow-sm);
}
.progress-header { display: flex; justify-content: space-between; margin-bottom: 0.6rem; font-size: 0.85rem; }
.progress-num { color: var(--c-text-muted); }
.progress-num strong { color: var(--c-primary); font-size: 1rem; }
.progress-pct { font-size: 0.8rem; color: var(--c-text-muted); }
.progress-track {
  height: 10px; background: var(--c-border-light);
  border-radius: 5px; overflow: hidden; position: relative;
}
.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, #3b82f6, #818cf8, #6366f1);
  background-size: 200% 100%;
  border-radius: 5px;
  transition: width 0.6s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
}
.progress-fill.shimmer { animation: shimmer 2s linear infinite; }
.progress-fill::after {
  content: ''; position: absolute; right: 0; top: 50%; transform: translateY(-50%);
  width: 20px; height: 20px; border-radius: 50%;
  background: #fff; box-shadow: 0 0 8px rgba(99,102,241,0.4);
}
.progress-legend { display: flex; gap: 1.25rem; margin-top: 0.5rem; flex-wrap: wrap; }
.legend-item { font-size: 0.7rem; color: var(--c-text-muted); display: flex; align-items: center; gap: 0.3rem; }
.legend-dot { width: 7px; height: 7px; border-radius: 50%; }
.legend-dot.done { background: #10b981; }
.legend-dot.pending { background: #3b82f6; animation: pulse 1.5s infinite; }
.legend-dot.generating { background: #f59e0b; animation: pulse 0.8s infinite; }
.legend-dot.failed { background: #ef4444; }
.active-workers { color: #b45309; font-weight: 500; }

/* Cards */
.cards-section { margin-bottom: 1.25rem; }
.section-title { font-size: 0.85rem; font-weight: 600; margin-bottom: 0.7rem; display: flex; align-items: center; gap: 0.4rem; }
.log-count { background: var(--c-bg); font-size: 0.65rem; padding: 0.1rem 0.4rem; border-radius: 999px; font-weight: 500; }
.log-scroll-hint { font-size: 0.65rem; color: var(--c-text-muted); font-weight: 400; }
.cards-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(270px, 1fr)); gap: 0.65rem; }

.task-card {
  background: var(--c-surface); border: 1px solid var(--c-border);
  border-radius: var(--radius); display: flex; overflow: hidden;
  transition: transform var(--transition), box-shadow var(--transition), border-color var(--transition);
  box-shadow: var(--shadow-sm);
}
.task-card:hover {
  transform: translateY(-2px); box-shadow: var(--shadow-md);
  border-color: #cbd5e1;
}
.task-card.generating {
  border-color: #f59e0b;
  box-shadow: 0 0 0 1px rgba(245,158,11,0.2);
  animation: borderPulse 2s infinite;
}
.card-status {
  width: 44px; min-width: 44px; display: flex; align-items: center;
  justify-content: center; font-size: 0.85rem; font-weight: 700; color: #fff;
  background: #94a3b8; transition: background var(--transition);
}
.card-status.generating { background: #f59e0b; }
.card-status.done, .card-status.qa_done { background: #10b981; }
.card-status.fixed { background: #8b5cf6; }
.card-status.failed { background: #ef4444; }
.card-status.pending { background: #94a3b8; }
.card-spinner {
  width: 16px; height: 16px; border: 2px solid rgba(255,255,255,0.4);
  border-top-color: #fff; border-radius: 50%; animation: spin 0.8s linear infinite;
}
.card-body { padding: 0.7rem 0.8rem; flex: 1; min-width: 0; }
.card-title { font-size: 0.82rem; font-weight: 600; line-height: 1.35; margin-bottom: 0.35rem; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.card-meta { display: flex; justify-content: space-between; align-items: center; gap: 0.5rem; }
.card-type { font-size: 0.65rem; color: var(--c-text-muted); background: var(--c-bg); padding: 0.15rem 0.4rem; border-radius: 4px; }
.card-status-label { font-size: 0.65rem; font-weight: 600; display: flex; align-items: center; gap: 0.2rem; }
.status-dot-sm { width: 5px; height: 5px; border-radius: 50%; background: currentColor; }
.card-status-label.generating { color: #d97706; }
.card-status-label.generating .status-dot-sm { animation: pulse 0.8s infinite; }
.card-status-label.done, .card-status-label.qa_done { color: #10b981; }
.card-status-label.fixed { color: #8b5cf6; }
.card-status-label.failed { color: #ef4444; }
.card-status-label.pending { color: #94a3b8; }
.card-file { font-size: 0.6rem; color: var(--c-text-muted); margin-top: 0.3rem; font-family: 'SF Mono', monospace; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

/* Card transitions */
.card-enter-active { transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1); }
.card-leave-active { transition: all 0.25s ease-in; }
.card-enter-from { opacity: 0; transform: translateY(12px) scale(0.96); }
.card-leave-to { opacity: 0; transform: scale(0.95); }

/* Pop */
.pop-enter-active { transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1); }
.pop-leave-active { transition: all 0.15s ease-in; }
.pop-enter-from { opacity: 0; transform: scale(0.5); }
.pop-leave-to { opacity: 0; transform: scale(0.5); }

/* Fade-up / Slide-down */
.fade-up-enter-active { transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1); }
.fade-up-leave-active { transition: all 0.2s ease-in; }
.fade-up-enter-from { opacity: 0; transform: translateY(10px); }
.fade-up-leave-to { opacity: 0; }
.slide-down-enter-active { transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1); }
.slide-down-leave-active { transition: all 0.2s ease-in; }
.slide-down-enter-from { opacity: 0; transform: translateY(-8px); }
.slide-down-leave-to { opacity: 0; }

/* ==================================================================== */
/* Event Log                                                            */
/* ==================================================================== */
.log-section { margin-bottom: 1.25rem; }
.log-box {
  background: var(--c-surface); border: 1px solid var(--c-border);
  border-radius: var(--radius); padding: 0.65rem 0.9rem;
  max-height: 320px; overflow-y: auto;
  font-family: 'SF Mono', 'Fira Code', 'Cascadia Code', 'JetBrains Mono', monospace;
  font-size: 0.75rem; line-height: 1.65;
  box-shadow: var(--shadow-sm);
}
.log-box::-webkit-scrollbar { width: 4px; }
.log-box::-webkit-scrollbar-track { background: transparent; }
.log-box::-webkit-scrollbar-thumb { background: #d1d5db; border-radius: 2px; }
.log-empty { display: flex; align-items: center; gap: 0.5rem; color: var(--c-text-muted); font-style: italic; }
.log-empty-icon { color: #d1d5db; animation: pulse 2s infinite; }
.log-line { display: flex; align-items: baseline; gap: 0.5rem; padding: 0.12rem 0; }
.log-line.answer { align-items: flex-start; }
.log-ts { color: #cbd5e1; flex-shrink: 0; font-size: 0.65rem; min-width: 4.5em; }
.log-dot { width: 4px; height: 4px; border-radius: 50%; background: #d1d5db; flex-shrink: 0; margin-top: 0.4em; }
.log-dot.worker { background: #f59e0b; animation: pulse 1s infinite; }
.log-text { color: var(--c-text-2); }
.log-line.tool .log-text { color: #6d28d9; }
.log-line.tool .log-dot { background: #a78bfa; }
.log-line.error .log-text { color: #dc2626; }
.log-line.error .log-dot { background: #fca5a5; }
.log-line.answer .log-text {
  color: var(--c-text);
  white-space: pre-wrap;
  word-break: break-word;
  line-height: 1.7;
}
.log-line.worker .log-text { color: #b45309; font-weight: 500; }
.log-line.file .log-dot { background: #10b981; }
.log-line.file .log-text { color: #059669; }
.log-line.divider { justify-content: center; padding: 0.3rem 0; }
.log-line.divider .log-text { color: #cbd5e1; font-size: 0.65rem; font-family: inherit; }

/* Markdown in log */
.log-text :deep(.md-h1) { display: block; font-size: 1.1em; font-weight: 700; color: var(--c-text); margin: 0.4em 0 0.2em; }
.log-text :deep(.md-h2) { display: block; font-size: 1em; font-weight: 700; color: var(--c-text); margin: 0.35em 0 0.15em; }
.log-text :deep(.md-h3) { display: block; font-size: 0.9em; font-weight: 600; color: var(--c-text-2); margin: 0.3em 0 0.1em; }
.log-text :deep(.md-hr) { display: block; height: 0; border-bottom: 1px dashed #e2e8f0; margin: 0.4em 0; }
.log-text :deep(.md-code) { background: #f1f5f9; padding: 0.1em 0.3em; border-radius: 3px; font-size: 0.9em; }
.log-text :deep(.md-path) { background: #eff6ff; color: #3b82f6; padding: 0.1em 0.3em; border-radius: 3px; }
.log-text :deep(strong) { font-weight: 600; color: var(--c-text); }

/* Log transitions */
.log-enter-active { transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1); }
.log-leave-active { transition: all 0.15s ease-in; }
.log-enter-from { opacity: 0; transform: translateX(-8px); }
.log-leave-to { opacity: 0; }

/* ==================================================================== */
/* Final Message                                                        */
/* ==================================================================== */
.final-message {
  background: linear-gradient(135deg, #ecfdf5, #f0fdf4);
  border: 1px solid #a7f3d0; border-radius: var(--radius-lg);
  padding: 1.25rem; margin-bottom: 1.25rem;
}
.final-header { display: flex; align-items: center; gap: 0.5rem; margin-bottom: 0.5rem; }
.final-icon {
  width: 24px; height: 24px; border-radius: 50%; background: #10b981;
  color: #fff; display: flex; align-items: center; justify-content: center;
  font-size: 0.7rem; font-weight: 700;
}
.final-header h3 { font-size: 0.9rem; }
.final-message p { font-size: 0.82rem; color: var(--c-text-2); white-space: pre-wrap; max-height: 160px; overflow-y: auto; }

/* ==================================================================== */
/* Batch Panels                                                         */
/* ==================================================================== */
.batch-panel {
  background: var(--c-surface); border: 1px solid var(--c-border);
  border-radius: var(--radius-lg); padding: 0.85rem 1rem; margin-bottom: 0.75rem;
  box-shadow: var(--shadow-sm); transition: border-color var(--transition);
}
.batch-panel:not(.done) { border-color: #fcd34d; box-shadow: 0 0 0 1px rgba(251,191,36,0.15); }
.batch-panel.done { border-color: #a7f3d0; }
.batch-header {
  display: flex; justify-content: space-between; align-items: center;
  margin-bottom: 0.6rem; font-size: 0.8rem; font-weight: 600;
}
.batch-label { display: flex; align-items: center; gap: 0.35rem; }
.batch-spinner {
  width: 12px; height: 12px; border: 2px solid #fcd34d;
  border-top-color: transparent; border-radius: 50%; animation: spin 0.8s linear infinite;
}
.batch-meta { font-size: 0.7rem; color: var(--c-text-muted); font-weight: 400; }
.batch-cards { display: flex; gap: 0.5rem; flex-wrap: wrap; }
.batch-card {
  display: flex; align-items: center; background: var(--c-bg);
  border: 1px solid var(--c-border); border-radius: var(--radius-sm);
  padding: 0.4rem 0.6rem; min-width: 150px; transition: all var(--transition);
}
.batch-card.generating { border-color: #f59e0b; background: #fffbeb; }
.batch-card.done, .batch-card.qa_done, .batch-card.fixed { border-color: #10b981; background: #ecfdf5; }
.batch-card.failed { border-color: #ef4444; background: #fef2f2; }
.bcard-idx {
  width: 22px; height: 22px; border-radius: 50%;
  background: #e2e8f0; color: #64748b; font-size: 0.65rem; font-weight: 700;
  display: flex; align-items: center; justify-content: center; margin-right: 0.4rem;
  flex-shrink: 0;
}
.batch-card.generating .bcard-idx { background: #f59e0b; color: #fff; }
.batch-card.done .bcard-idx, .batch-card.qa_done .bcard-idx, .batch-card.fixed .bcard-idx { background: #10b981; color: #fff; }
.bcard-body { flex: 1; min-width: 0; }
.bcard-title { font-size: 0.7rem; font-weight: 500; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.bcard-status {
  font-size: 0.6rem; font-weight: 600; margin-top: 0.1rem;
  color: var(--c-text-muted);
}
.bcard-status.generating { color: #d97706; }
.bcard-status.done, .bcard-status.qa_done, .bcard-status.fixed { color: #059669; }

/* ==================================================================== */
/* Files                                                                */
/* ==================================================================== */
.files-section { margin-bottom: 1.25rem; }
.files-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(240px, 1fr)); gap: 0.65rem; }
.file-card {
  display: flex; flex-direction: column;
  background: var(--c-surface); border: 1px solid var(--c-border);
  border-radius: var(--radius); text-decoration: none; color: var(--c-text);
  transition: transform var(--transition), box-shadow var(--transition), border-color var(--transition);
  box-shadow: var(--shadow-sm); overflow: hidden;
}
.file-card:hover {
  transform: translateY(-2px); box-shadow: var(--shadow-md);
  border-color: var(--c-primary);
}
.file-thumb {
  width: 100%; height: 120px; background: var(--c-bg);
  display: flex; align-items: center; justify-content: center;
  overflow: hidden; position: relative;
}
.file-thumb img {
  width: 100%; height: 100%; object-fit: cover;
  opacity: 0; transition: opacity 0.3s;
}
.file-thumb img.loaded { opacity: 1; }
.thumb-spinner {
  position: absolute; width: 20px; height: 20px;
  border: 2px solid #e2e8f0; border-top-color: var(--c-primary);
  border-radius: 50%; animation: spin 0.7s linear infinite;
}
.file-icon svg { width: 28px; height: 28px; color: #94a3b8; }
.file-name {
  font-size: 0.72rem; font-family: 'SF Mono', monospace;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
  padding: 0.5rem 0.6rem 0; flex: 1;
}
.file-arrow {
  font-size: 0.65rem; color: var(--c-text-muted); padding: 0.15rem 0.6rem 0.5rem;
  transition: transform var(--transition); align-self: flex-start;
}
.file-card:hover .file-arrow { transform: translateY(2px); color: var(--c-primary); }

.file-enter-active { transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1); }
.file-leave-active { transition: all 0.2s ease-in; }
.file-enter-from { opacity: 0; transform: scale(0.9); }
.file-leave-to { opacity: 0; }

/* ==================================================================== */
/* Shared                                                               */
/* ==================================================================== */
.empty-hint { font-size: 0.78rem; color: #475569; font-style: italic; padding: 0.5rem 0.25rem; }

@media (max-width: 900px) {
  .layout { flex-direction: column; }
  .sidebar { width: 100%; min-width: 0; height: auto; position: static; max-height: 50vh; }
  .welcome-hints { grid-template-columns: 1fr; }
  .cards-grid { grid-template-columns: 1fr; }
  .main { padding: 1rem; }
}
</style>
