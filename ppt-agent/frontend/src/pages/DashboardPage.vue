<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue';
import { useRouter } from 'vue-router';
import type { TaskInfo, TaskItem, SSEEvent } from '../types';
import { STATUS_LABELS } from '../types';
import { fetchTasks, createTask, fetchTask, cancelTask, deleteTask, isLoggedIn } from '../api';
import { authState } from '../stores/auth';
import Sidebar from '../components/Sidebar.vue';
import ProgressBar from '../components/ProgressBar.vue';
import EventLog from '../components/EventLog.vue';
import SlidePreviewCard from '../components/SlidePreviewCard.vue';

const router = useRouter();
const auth = authState;

// ── State (same as old working App.vue) ─────────────────────────────────
const tasks = ref<TaskInfo[]>([]);
const selectedId = ref<string | null>(null);
const taskItems = ref<TaskItem[]>([]);
const doneCount = ref(0);
const totalCount = ref(0);
const logLines = ref<{ ts: number; text: string; kind: import('../types').LogKind }[]>([]);
const finalFiles = ref<string[]>([]);
const finalMessage = ref('');
const duration = ref('');
const activeWorkers = ref(0);
const cancelling = ref(false);
const creating = ref(false);

interface Batch { id: number; taskIds: string[]; ts: number; done: boolean; }
const batches = ref<Batch[]>([]);
let batchIdSeq = 0;
let es: EventSource | null = null;

const selectedTask = computed(() => tasks.value.find(t => t.id === selectedId.value));
const hasRunningTask = computed(() => tasks.value.some(t => t.status === 'running'));

function fmtTokens(n: number): string {
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M';
  if (n >= 1_000) return (n / 1000).toFixed(1) + 'K';
  return String(n);
}

const sampleQueries = [
  '做一个关于新能源汽车的行业分析报告',
  '制作一个产品发布会演示文稿',
  '做一个 AI 大模型技术分享的 PPT',
  '制作一个公司季度总结汇报',
];

const sampleThemes = [
  { name: '科技蓝', emoji: '🔷', color: 'linear-gradient(135deg, #2563eb, #3b82f6)', desc: '企业/技术' },
  { name: '商务灰', emoji: '⬜', color: 'linear-gradient(135deg, #475569, #64748b)', desc: '商务/正式' },
  { name: '活力橙', emoji: '🟠', color: 'linear-gradient(135deg, #ea580c, #f97316)', desc: '创意/营销' },
  { name: '自然绿', emoji: '🟢', color: 'linear-gradient(135deg, #16a34a, #22c55e)', desc: '环保/教育' },
  { name: '优雅紫', emoji: '🟣', color: 'linear-gradient(135deg, #7c3aed, #8b5cf6)', desc: '设计/时尚' },
  { name: '海洋风', emoji: '🌊', color: 'linear-gradient(135deg, #0891b2, #06b6d4)', desc: '医疗/金融' },
];

const orderedSlides = computed(() => {
  const slides = new Map<string, { task: TaskItem; fileReady: boolean }>();
  for (const t of taskItems.value) {
    slides.set(t.task_id, { task: t, fileReady: finalFiles.value.includes(t.output_file) });
  }
  // Fallback: when taskItems is empty (e.g. restored from MySQL after restart),
  // generate entries directly from the file list
  if (slides.size === 0 && finalFiles.value.length > 0) {
    finalFiles.value.forEach((f, i) => {
      const name = f.split(/[/\\]/).pop() || f;
      slides.set(f, {
        task: {
          task_id: f,
          page_index: i + 1,
          title: name,
          content_type: '',
          output_file: f,
          status: 'done',
        },
        fileReady: true,
      });
    });
  }
  return [...slides.values()].sort((a, b) => a.task.page_index - b.task.page_index);
});

function findItem(id: string): TaskItem | undefined {
  return taskItems.value.find(t => t.task_id === id);
}

// ── Multi-select download ──────────────────────────────────────────────
const selectedSlides = ref<Set<string>>(new Set());
const readySlides = computed(() => orderedSlides.value.filter(s => s.fileReady));

function toggleSelect(taskId: string) {
  const next = new Set(selectedSlides.value);
  if (next.has(taskId)) next.delete(taskId);
  else next.add(taskId);
  selectedSlides.value = next;
}

function selectAll() {
  selectedSlides.value = new Set(readySlides.value.map(s => s.task.task_id));
}

function deselectAll() {
  selectedSlides.value = new Set();
}

function downloadSelected() {
  const ids = selectedSlides.value;
  if (ids.size === 0) return;
  const slides = readySlides.value.filter(s => ids.has(s.task.task_id));
  // Download each file with a small stagger
  slides.forEach((s, i) => {
    setTimeout(() => {
      const name = s.task.output_file.split(/[/\\]/).pop() || s.task.output_file;
      const a = document.createElement('a');
      a.href = `/api/tasks/${selectedTask.value!.id}/files/${encodeURIComponent(name)}`;
      a.download = name;
      a.style.display = 'none';
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
    }, i * 300);
  });
}

// ── Log ────────────────────────────────────────────────────────────────
function addLog(kind: import('../types').LogKind, text: string) {
  if (kind === 'answer' && logLines.value.length > 0) {
    const last = logLines.value[logLines.value.length - 1];
    if (last.kind === 'answer') {
      last.text += text;
      last.ts = Date.now();
      return;
    }
  }
  logLines.value = [...logLines.value.slice(-500), { ts: Date.now(), text, kind }];
}

// ── SSE ────────────────────────────────────────────────────────────────
function connectSSE(taskId: string) {
  if (!taskId) return;
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
        if (name === 'task' && args.includes('SlideExecutor')) {
          const ids = parseTaskIds(args);
          markGenerating(ids);
          activeWorkers.value = Math.max(1, ids.length || 1);
          batches.value = [...batches.value, { id: ++batchIdSeq, taskIds: ids, ts: Date.now(), done: false }];
          addLog('worker', `派发 ${ids.length} 个子任务并行执行`);
        } else if (name === 'task') {
          const ids = parseTaskIds(args);
          markGenerating(ids);
          activeWorkers.value = Math.max(activeWorkers.value, ids.length || 1);
          addLog('tool', `> ${name} (${args.slice(0, 100)})`);
        } else {
          addLog('tool', `> ${name} (${args.slice(0, 120)})`);
        }
        break;
      }

      case 'progress':
        if (evt.tasks) {
          taskItems.value = evt.tasks;
          for (const batch of batches.value) {
            if (batch.done) continue;
            batch.done = batch.taskIds.every(tid => {
              const item = evt.tasks!.find(t => t.task_id === tid);
              return item && (item.status === 'done' || item.status === 'qa_done' || item.status === 'fixed');
            });
          }
          for (const t of evt.tasks) {
            if ((t.status === 'done' || t.status === 'qa_done' || t.status === 'fixed') && t.output_file) {
              if (!finalFiles.value.includes(t.output_file)) {
                finalFiles.value = [...finalFiles.value, t.output_file];
                addLog('file', `${t.output_file} 已生成`);
                cachePPT(taskId, t.output_file);
              }
            }
          }
        }
        if (evt.done !== undefined) doneCount.value = evt.done;
        if (evt.total !== undefined) totalCount.value = evt.total;
        break;

      case 'file_ready':
        if (evt.files) {
          for (const f of evt.files) {
            if (!finalFiles.value.includes(f)) {
              finalFiles.value = [...finalFiles.value, f];
              addLog('file', `${f} 已生成`);
              cachePPT(taskId, f);
            }
          }
        }
        break;

      case 'error':
        addLog('error', evt.error || evt.content || '');
        break;

      case 'complete':
        doneCount.value = evt.done || 0;
        totalCount.value = evt.total || 0;
        if (evt.files) {
          finalFiles.value = evt.files;
          for (const f of evt.files) cachePPT(taskId, f);
        }
        if (evt.message) finalMessage.value = evt.message;
        if (evt.duration) duration.value = evt.duration;
        if (evt.tasks) taskItems.value = evt.tasks;
        if (evt.total_tokens) {
          const t = tasks.value.find(x => x.id === taskId);
          if (t) {
            t.prompt_tokens = evt.prompt_tokens || 0;
            t.completion_tokens = evt.completion_tokens || 0;
            t.total_tokens = evt.total_tokens || 0;
          }
        }
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

function disconnectSSE() {
  if (es) { es.close(); es = null; }
}

// Cache PPT files and thumbnails via Cache API (service worker storage)
async function cachePPT(taskId: string, filename: string) {
  try {
    const cache = await caches.open('ppt-agent-v1');
    const name = filename.split(/[/\\]/).pop() || filename;
    // Cache the file
    const fileUrl = `/api/tasks/${taskId}/files/${encodeURIComponent(name)}`;
    if (!await cache.match(fileUrl)) {
      const res = await fetch(fileUrl);
      if (res.ok) cache.put(fileUrl, res.clone());
    }
    // Cache the thumbnail
    const thumbUrl = `/api/tasks/${taskId}/thumb/${encodeURIComponent(name)}`;
    if (!await cache.match(thumbUrl)) {
      const res = await fetch(thumbUrl);
      if (res.ok) cache.put(thumbUrl, res.clone());
    }
  } catch { /* non-critical */ }
}

// ── Helpers ────────────────────────────────────────────────────────────
function parseTaskIds(args: string): string[] {
  try {
    const obj = JSON.parse(args);
    if (Array.isArray(obj.task_ids)) return obj.task_ids.map(String);
    if (typeof obj.task_id === 'string') return obj.task_id.split(',').map((s: string) => s.trim());
    if (typeof obj.task_ids === 'string') return obj.task_ids.split(',').map((s: string) => s.trim());
    return [];
  } catch { return []; }
}

function markGenerating(ids: string[]) {
  for (const item of taskItems.value) {
    if (ids.includes(item.task_id) && item.status === 'pending') {
      item.status = 'generating';
    }
  }
}

async function refreshTask(id: string) {
  try {
    const info = await fetchTask(id);
    const idx = tasks.value.findIndex(t => t.id === id);
    if (idx >= 0) tasks.value[idx] = info;
  } catch { /* ignore */ }
}

// ── Task state cache (preserve state when switching between tasks) ─────
interface TaskCache {
  taskItems: TaskItem[];
  doneCount: number;
  totalCount: number;
  logLines: { ts: number; text: string; kind: import('../types').LogKind }[];
  finalFiles: string[];
  finalMessage: string;
  duration: string;
  activeWorkers: number;
  batches: Batch[];
}
const taskCache = new Map<string, TaskCache>();

function saveCache(id: string) {
  taskCache.set(id, {
    taskItems: [...taskItems.value],
    doneCount: doneCount.value,
    totalCount: totalCount.value,
    logLines: [...logLines.value],
    finalFiles: [...finalFiles.value],
    finalMessage: finalMessage.value,
    duration: duration.value,
    activeWorkers: activeWorkers.value,
    batches: [...batches.value],
  });
}

function restoreCache(id: string): boolean {
  const c = taskCache.get(id);
  if (!c) return false;
  taskItems.value = c.taskItems;
  doneCount.value = c.doneCount;
  totalCount.value = c.totalCount;
  logLines.value = c.logLines;
  finalFiles.value = c.finalFiles;
  finalMessage.value = c.finalMessage;
  duration.value = c.duration;
  activeWorkers.value = c.activeWorkers;
  batches.value = c.batches;
  return true;
}

// ── Actions ────────────────────────────────────────────────────────────
function selectTask(id: string) {
  if (!id) return;
  if (selectedId.value && selectedId.value !== id) {
    saveCache(selectedId.value);
    disconnectSSE();
  }
  selectedId.value = id;
  const t = tasks.value.find(x => x.id === id);
  if (!t) return;

  // Restore cached state if switching back to a previously viewed task
  if (restoreCache(id)) {
    // Check if cached data is incomplete (replay was interrupted, or new data arrived)
    const cachedDone = doneCount.value;
    const cachedFiles = finalFiles.value.length;
    const needReplay = t.status !== 'failed' && (
      t.status === 'running' || // running: always reconnect for live updates
      t.done_count > cachedDone || // more done than cached → replay was incomplete
      (t.files?.length || 0) > cachedFiles // more files than cached
    );
    if (needReplay) {
      // Clear volatile state before replay to avoid duplicates
      logLines.value = [];
      taskItems.value = [];
      batches.value = [];
      connectSSE(id);
    }
    return;
  }

  // Fresh task: connect SSE to replay all stored events (works for running/completed/cancelled)
  taskItems.value = [];
  doneCount.value = t.done_count;
  totalCount.value = t.total_count;
  logLines.value = [];
  finalFiles.value = t.files || [];
  finalMessage.value = '';
  duration.value = t.duration || '';
  activeWorkers.value = 0;
  batches.value = [];
  selectedSlides.value = new Set();
  if (t.status !== 'failed') {
    connectSSE(id);
  }
}

async function handleCreateTask(query: string) {
  if (creating.value) return;
  creating.value = true;
  try {
    const info = await createTask(query);
    tasks.value = [info, ...tasks.value];
    selectTask(info.id);
  } catch (e: any) {
    console.error('创建任务失败:', e);
  } finally {
    creating.value = false;
  }
}

async function handleCancel() {
  if (!selectedId.value || cancelling.value) return;
  cancelling.value = true;
  try {
    const info = await cancelTask(selectedId.value);
    const idx = tasks.value.findIndex(t => t.id === info.id);
    if (idx >= 0) tasks.value[idx] = info;
    disconnectSSE();
  } catch (e) {
    console.error('取消失败:', e);
  } finally {
    cancelling.value = false;
  }
}

async function handleDeleteTask(id: string) {
  try {
    await deleteTask(id);
    tasks.value = tasks.value.filter(t => t.id !== id);
    taskCache.delete(id);
    if (selectedId.value === id) {
      selectedId.value = null;
      disconnectSSE();
    }
  } catch (e) {
    console.error('删除失败:', e);
  }
}

// ── Lifecycle ──────────────────────────────────────────────────────────
async function loadTasks() {
  try { tasks.value = await fetchTasks(); } catch { /* noop */ }
}

onMounted(async () => {
  if (!isLoggedIn()) { router.push('/auth'); return; }
  await auth.init();
  if (!auth.loggedIn) { router.push('/auth'); return; }
  await loadTasks();
});

onUnmounted(() => { disconnectSSE(); });
</script>

<template>
  <div class="layout">
    <Sidebar
      :user="auth.user"
      :tasks="tasks"
      :selected-id="selectedId"
      :has-running-task="hasRunningTask"
      :creating="creating"
      @logout="auth.logout(); router.push('/')"
      @select-task="selectTask"
      @create-task="handleCreateTask"
      @delete-task="handleDeleteTask"
    />

    <main class="main">
      <!-- Welcome -->
      <div v-if="!selectedTask" class="welcome">
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
            </svg>
          </div>
          <h2>在左侧输入需求，开始生成 PPT</h2>
          <p>AI 将自动规划大纲、生成幻灯片、视觉质检、迭代修复</p>
        </div>

        <!-- Quick start examples -->
        <div class="welcome-examples">
          <h3>快速开始 — 点击示例即可填入</h3>
          <div class="examples-grid">
            <button v-for="ex in sampleQueries" :key="ex" class="example-chip" @click="handleCreateTask(ex)">{{ ex }}</button>
          </div>
        </div>

        <!-- Theme showcase -->
        <div class="welcome-themes">
          <h3>内置主题风格</h3>
          <div class="themes-grid">
            <div v-for="t in sampleThemes" :key="t.name" class="theme-card">
              <div class="theme-swatch" :style="{ background: t.color }">{{ t.emoji }}</div>
              <span class="theme-name">{{ t.name }}</span>
              <span class="theme-desc">{{ t.desc }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Dashboard -->
      <template v-else>
        <!-- Header -->
        <div class="dash-header">
          <div class="dash-header-left">
            <h2 class="dash-title">{{ selectedTask?.query }}</h2>
            <div class="dash-meta">
              <span class="dash-id">{{ selectedTask?.id?.slice(0, 8) }}</span>
              <span v-if="duration" class="dash-duration">
                <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="8" cy="8" r="7"/><path d="M8 4v4l3 2"/></svg>
                {{ duration }}
              </span>
              <span v-if="(selectedTask?.total_tokens ?? 0) > 0" class="dash-tokens">
                <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5"><rect x="2" y="3" width="12" height="10" rx="1"/><line x1="5" y1="7" x2="5" y2="10"/><line x1="8" y1="5" x2="8" y2="10"/><line x1="11" y1="6" x2="11" y2="10"/></svg>
                {{ fmtTokens(selectedTask?.total_tokens ?? 0) }} tokens
                <span class="token-detail">({{ fmtTokens(selectedTask?.prompt_tokens ?? 0) }}p + {{ fmtTokens(selectedTask?.completion_tokens ?? 0) }}c)</span>
              </span>
            </div>
          </div>
          <div class="dash-status">
            <span v-if="activeWorkers > 0 && selectedTask?.status === 'running'" class="stat-badge workers">
              <span class="pulse-dot"></span>{{ activeWorkers }} 并行执行中
            </span>
            <span v-if="selectedTask?.status === 'running'" class="stat-badge running">
              <span class="pulse-dot"></span>运行中
            </span>
            <span v-if="selectedTask?.status === 'completed'" class="stat-badge done"> 已完成</span>
            <span v-if="selectedTask?.status === 'cancelled'" class="stat-badge cancelled"> 已中断</span>
            <span v-if="selectedTask?.status === 'failed'" class="stat-badge failed"> 失败</span>
            <button
              v-if="selectedTask?.status === 'running'"
              class="cancel-btn"
              :disabled="cancelling"
              @click="handleCancel"
            >{{ cancelling ? '取消中...' : ' 中断' }}</button>
          </div>
        </div>

        <!-- Progress -->
        <ProgressBar
          :done-count="doneCount"
          :total-count="totalCount"
          :task-items="taskItems"
          :is-running="selectedTask?.status === 'running'"
        />

        <!-- Left-Right Split -->
        <div class="split-layout">
          <div class="split-left">
            <EventLog :lines="logLines" max-height="calc(100vh - 300px)" />
          </div>
          <div class="split-right">
            <!-- Batch panels -->
            <TransitionGroup name="batch" tag="div">
              <div v-for="batch in batches" :key="batch.id" class="batch-panel" :class="{ done: batch.done }">
                <div class="batch-header">
                  <span class="batch-label">
                    <span v-if="!batch.done" class="batch-spinner"></span>
                    {{ batch.done ? '' : '' }} 批次 #{{ batch.id }}
                  </span>
                  <span class="batch-meta">{{ batch.done ? '已完成' : '执行中' }}  {{ batch.taskIds.length }} 页</span>
                </div>
                <div class="batch-cards">
                  <div v-for="tid in batch.taskIds" :key="tid" class="batch-card" :class="(findItem(tid) || {}).status || 'pending'">
                    <div class="bcard-idx">{{ (findItem(tid) || {}).page_index || '?' }}</div>
                    <div class="bcard-body">
                      <div class="bcard-title">{{ (findItem(tid) || {}).title || tid }}</div>
                      <div class="bcard-status" :class="(findItem(tid) || {}).status || 'pending'">
                        {{ STATUS_LABELS[(findItem(tid) || {}).status || 'pending'] }}
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </TransitionGroup>

            <!-- Slides grid -->
            <div v-if="orderedSlides.length > 0" class="slides-section">
              <div class="slides-toolbar">
                <h3 class="section-title">
                  生成文件
                  <span class="file-count">{{ finalFiles.length }} / {{ orderedSlides.length }}</span>
                </h3>
                <div class="toolbar-actions">
                  <button class="tool-btn" @click="selectAll">全选</button>
                  <button class="tool-btn" @click="deselectAll">取消</button>
                  <button
                    class="tool-btn primary"
                    :disabled="selectedSlides.size === 0"
                    @click="downloadSelected"
                  >
                    下载选中 ({{ selectedSlides.size }})
                  </button>
                </div>
              </div>
              <TransitionGroup name="slide" tag="div" class="slides-list">
                <SlidePreviewCard
                  v-for="s in orderedSlides"
                  :key="s.task.task_id"
                  :task="s.task"
                  :task-id="selectedTask?.id"
                  :file-ready="s.fileReady"
                  :selected="selectedSlides.has(s.task.task_id)"
                  @toggle="toggleSelect(s.task.task_id)"
                />
              </TransitionGroup>
            </div>
          </div>
        </div>

        <!-- Final message -->
        <div v-if="finalMessage && selectedTask?.status !== 'running'" class="final-message">
          <div class="final-header">
            <span class="final-icon"></span>
            <h3>任务完成</h3>
          </div>
          <p>{{ finalMessage }}</p>
        </div>
      </template>
    </main>
  </div>
</template>

<style scoped>
.layout { display: flex; min-height: 100vh; }
.main { flex: 1; padding: 1.5rem 2rem; overflow-y: auto; max-height: 100vh; }

/* Welcome */
.welcome { display: flex; flex-direction: column; align-items: center; justify-content: center; min-height: 60vh; text-align: center; }
.welcome-hero { animation: fadeInUp 0.6s cubic-bezier(0.4, 0, 0.2, 1); }
.welcome-icon svg { width: 100px; height: 100px; margin: 0 auto 1.25rem; }
.welcome h2 { font-size: 1.5rem; font-weight: 700; margin-bottom: 0.5rem; background: linear-gradient(135deg, #1e293b, #475569); -webkit-background-clip: text; -webkit-text-fill-color: transparent; }
.welcome p { color: var(--c-text-muted); font-size: 0.9rem; }

/* Welcome examples */
.welcome-examples { max-width: 650px; margin: 0 auto 2rem; text-align: center; }
.welcome-examples h3 { font-size: 0.85rem; font-weight: 600; color: var(--c-text-2); margin-bottom: 0.75rem; }
.examples-grid { display: flex; flex-wrap: wrap; gap: 0.5rem; justify-content: center; }
.example-chip {
  padding: 0.45rem 0.9rem; border-radius: 999px;
  border: 1px solid var(--c-border); background: var(--c-surface);
  color: var(--c-text-2); font-size: 0.78rem; cursor: pointer;
  transition: all var(--transition); font-family: inherit;
}
.example-chip:hover {
  border-color: var(--c-primary); color: var(--c-primary);
  background: var(--c-primary-light); transform: translateY(-1px);
}

/* Welcome themes */
.welcome-themes { max-width: 600px; margin: 0 auto; text-align: center; }
.welcome-themes h3 { font-size: 0.85rem; font-weight: 600; color: var(--c-text-2); margin-bottom: 0.75rem; }
.themes-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 0.6rem; }
.theme-card {
  background: var(--c-surface); border: 1px solid var(--c-border);
  border-radius: var(--radius); padding: 0.75rem 0.5rem;
  text-align: center; transition: transform var(--transition), box-shadow var(--transition);
  cursor: default;
}
.theme-card:hover { transform: translateY(-2px); box-shadow: var(--shadow-md); }
.theme-swatch {
  width: 40px; height: 40px; border-radius: 10px;
  margin: 0 auto 0.4rem; display: flex; align-items: center; justify-content: center;
  font-size: 1.1rem;
}
.theme-name { display: block; font-size: 0.75rem; font-weight: 600; color: var(--c-text); margin-bottom: 0.1rem; }
.theme-desc { font-size: 0.65rem; color: var(--c-text-muted); }

/* Header */
.dash-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 1rem; animation: fadeInUp 0.4s both; }
.dash-header-left { flex: 1; min-width: 0; }
.dash-title { font-size: 1.1rem; font-weight: 650; line-height: 1.4; }
.dash-meta { display: flex; align-items: center; gap: 1rem; margin-top: 0.3rem; }
.dash-id { font-size: 0.68rem; color: var(--c-text-muted); font-family: 'SF Mono', monospace; background: var(--c-border-light); padding: 0.12rem 0.4rem; border-radius: 4px; }
.dash-duration { font-size: 0.72rem; color: var(--c-text-muted); display: flex; align-items: center; gap: 0.2rem; }
.dash-duration svg { width: 12px; height: 12px; }
.dash-tokens { font-size: 0.72rem; color: var(--c-text-muted); display: flex; align-items: center; gap: 0.2rem; }
.dash-tokens svg { width: 13px; height: 13px; }
.token-detail { font-size: 0.64rem; color: var(--c-text-muted); opacity: 0.7; }
.dash-status { display: flex; gap: 0.5rem; flex-shrink: 0; align-items: center; }
.stat-badge { font-size: 0.72rem; font-weight: 600; padding: 0.3rem 0.7rem; border-radius: 999px; display: flex; align-items: center; gap: 0.3rem; }
.stat-badge.running { background: var(--c-primary-light); color: #1d4ed8; }
.stat-badge.workers { background: var(--c-warning-light); color: #b45309; }
.stat-badge.done { background: var(--c-success-light); color: #059669; }
.stat-badge.cancelled { background: #fef3c7; color: #92400e; }
.stat-badge.failed { background: var(--c-danger-light); color: #b91c1c; }
.pulse-dot { width: 7px; height: 7px; border-radius: 50%; background: currentColor; animation: pulse 1.5s infinite; }
.cancel-btn { font-size: 0.72rem; font-weight: 600; padding: 0.3rem 0.7rem; border-radius: 999px; border: 1px solid #fca5a5; background: var(--c-danger-light); color: #dc2626; cursor: pointer; display: flex; align-items: center; gap: 0.2rem; transition: all var(--transition); font-family: inherit; }
.cancel-btn:hover { background: #fecaca; border-color: #f87171; }
.cancel-btn:disabled { opacity: 0.5; cursor: not-allowed; }

/* Split */
.split-layout { display: grid; grid-template-columns: 33% 67%; gap: 1rem; margin-bottom: 1rem; }
.split-left { min-width: 0; }
.split-right { min-width: 0; }

/* Section title */
.section-title { font-size: 0.82rem; font-weight: 600; margin-bottom: 0.6rem; display: flex; align-items: center; gap: 0.5rem; }
.file-count { font-size: 0.66rem; background: var(--c-bg); padding: 0.15rem 0.45rem; border-radius: 999px; font-weight: 500; color: var(--c-text-muted); }

/* Slides grid */
.slides-section { margin-bottom: 1rem; }
.slides-toolbar { display: flex; align-items: center; justify-content: space-between; margin-bottom: 0.6rem; flex-wrap: wrap; gap: 0.5rem; }
.toolbar-actions { display: flex; gap: 0.4rem; }
.tool-btn {
  font-size: 0.7rem; font-weight: 500;
  padding: 0.3rem 0.65rem; border-radius: 6px;
  border: 1px solid var(--c-border);
  background: var(--c-surface); color: var(--c-text-2);
  cursor: pointer; transition: all var(--transition);
  font-family: inherit;
}
.tool-btn:hover { border-color: #cbd5e1; background: var(--c-bg); }
.tool-btn.primary {
  background: var(--c-primary); color: #fff; border-color: var(--c-primary);
}
.tool-btn.primary:hover { background: #2563eb; }
.tool-btn.primary:disabled { opacity: 0.5; cursor: not-allowed; }
.slides-list { display: flex; flex-direction: column; gap: 0.6rem; }

/* Batch panels */
.batch-panel { background: var(--c-surface); border: 1px solid var(--c-border); border-radius: var(--radius); padding: 0.75rem 0.9rem; margin-bottom: 0.65rem; box-shadow: var(--shadow-sm); transition: border-color var(--transition); }
.batch-panel:not(.done) { border-color: #fcd34d; box-shadow: 0 0 0 1px rgba(251,191,36,0.15); }
.batch-panel.done { border-color: #a7f3d0; }
.batch-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 0.5rem; font-size: 0.78rem; font-weight: 600; }
.batch-label { display: flex; align-items: center; gap: 0.3rem; }
.batch-spinner { width: 11px; height: 11px; border: 2px solid #fcd34d; border-top-color: transparent; border-radius: 50%; animation: spin 0.8s linear infinite; }
.batch-meta { font-size: 0.65rem; color: var(--c-text-muted); font-weight: 400; }
.batch-cards { display: flex; gap: 0.4rem; flex-wrap: wrap; }
.batch-card { display: flex; align-items: center; background: var(--c-bg); border: 1px solid var(--c-border); border-radius: var(--radius-sm); padding: 0.35rem 0.5rem; min-width: 140px; transition: all var(--transition); }
.batch-card.generating { border-color: #f59e0b; background: #fffbeb; }
.batch-card.done, .batch-card.qa_done, .batch-card.fixed { border-color: #10b981; background: #ecfdf5; }
.batch-card.failed { border-color: #ef4444; background: #fef2f2; }
.bcard-idx { width: 20px; height: 20px; border-radius: 50%; background: #e2e8f0; color: #64748b; font-size: 0.6rem; font-weight: 700; display: flex; align-items: center; justify-content: center; margin-right: 0.35rem; flex-shrink: 0; }
.batch-card.generating .bcard-idx { background: #f59e0b; color: #fff; }
.batch-card.done .bcard-idx, .batch-card.qa_done .bcard-idx, .batch-card.fixed .bcard-idx { background: #10b981; color: #fff; }
.bcard-body { flex: 1; min-width: 0; }
.bcard-title { font-size: 0.68rem; font-weight: 500; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.bcard-status { font-size: 0.58rem; font-weight: 600; margin-top: 0.05rem; color: var(--c-text-muted); }
.bcard-status.generating { color: #d97706; }
.bcard-status.done, .bcard-status.qa_done, .bcard-status.fixed { color: #059669; }

/* Final message */
.final-message { background: linear-gradient(135deg, #ecfdf5, #f0fdf4); border: 1px solid #a7f3d0; border-radius: var(--radius); padding: 1.1rem; margin-bottom: 1rem; animation: fadeInUp 0.4s both; }
.final-header { display: flex; align-items: center; gap: 0.5rem; margin-bottom: 0.4rem; }
.final-icon { width: 22px; height: 22px; border-radius: 50%; background: #10b981; color: #fff; display: flex; align-items: center; justify-content: center; font-size: 0.65rem; font-weight: 700; }
.final-header h3 { font-size: 0.85rem; }
.final-message p { font-size: 0.78rem; color: var(--c-text-2); white-space: pre-wrap; max-height: 120px; overflow-y: auto; }

/* Transitions */
.slide-enter-active { transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1); }
.slide-leave-active { transition: all 0.2s ease-in; }
.slide-enter-from { opacity: 0; transform: scale(0.9); }
.slide-leave-to { opacity: 0; }
.batch-enter-active { transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1); }
.batch-leave-active { transition: all 0.2s ease-in; }
.batch-enter-from { opacity: 0; transform: translateY(-8px); }
.batch-leave-to { opacity: 0; }

@keyframes fadeInUp { from { opacity: 0; transform: translateY(12px); } to { opacity: 1; transform: translateY(0); } }
@keyframes pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.5; } }
@keyframes spin { to { transform: rotate(360deg); } }
</style>
