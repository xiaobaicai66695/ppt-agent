<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue';
import { useRouter } from 'vue-router';
import type { TaskInfo, TaskItem, SSEEvent } from '../types';
import { STATUS_LABELS } from '../types';
import { fetchTasks, createTask, fetchTask, cancelTask, deleteTask, isLoggedIn, clearToken, continueTask, fetchConversation } from '../api';
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
const loadError = ref('');

// ── Chat bar state ─────────────────────────────────────────────────────────────
const showChatInput = ref(false);
const chatInput = ref('');
const chatLoading = ref(false);
const chatHistory = ref<import('../types').ConversationMessage[]>([]);
let chatEs: EventSource | null = null;

function openChatBar() {
  showChatInput.value = true;
  chatHistory.value = [];
  if (selectedId.value) {
    fetchConversation(selectedId.value)
      .then(s => { chatHistory.value = s.messages || []; })
      .catch(() => { /* ignore */ });
  }
}

function closeChatBar() {
  showChatInput.value = false;
  if (chatEs) { chatEs.close(); chatEs = null; }
}

async function submitChat() {
  if (!selectedId.value || !chatInput.value.trim() || chatLoading.value) return;
  chatLoading.value = true;
  const msg = chatInput.value.trim();
  chatInput.value = '';

  chatHistory.value = [...chatHistory.value, {
    role: 'user' as const, content: msg, timestamp: new Date().toISOString(),
  }];

  try {
    const res = await continueTask(selectedId.value, msg);
    if (!res.ok) {
      const err = await res.json();
      throw new Error(err.error || '发送失败');
    }
    const data = await res.json();
    if (data.status === 'queued') {
      // 任务运行中，消息已排队
      chatHistory.value = [...chatHistory.value, {
        role: 'assistant' as const,
        content: data.message || '您的反馈已排队，将在当前任务完成后自动处理',
        timestamp: new Date().toISOString(),
      }];
      chatLoading.value = false;
      return;
    }
    connectChatSSE(selectedId.value);
  } catch (e: any) {
    chatHistory.value = [...chatHistory.value, {
      role: 'assistant' as const, content: '发送失败: ' + e.message, timestamp: new Date().toISOString(),
    }];
    chatLoading.value = false;
  }
}

function connectChatSSE(taskId: string) {
  if (chatEs) chatEs.close();
  chatEs = new EventSource(`/api/tasks/${taskId}/stream`);

  chatEs.onerror = () => {
    chatEs?.close();
    chatEs = null;
    chatLoading.value = false;
  };

  const handler = (e: MessageEvent) => {
    let evt: SSEEvent;
    try { evt = JSON.parse(e.data); } catch { return; }

    if (evt.type === 'answer') {
      addLog('answer', evt.content || '');
    } else if (evt.type === 'tool_call') {
      addLog('tool', `[${evt.tool_name || 'tool'}]`);
    } else if (evt.type === 'file_ready' && evt.files) {
      for (const f of evt.files) {
        if (!finalFiles.value.includes(f)) {
          finalFiles.value = [...finalFiles.value, f];
          addLog('file', `${f} 已更新`);
          cachePPT(taskId, f);
        }
      }
    } else if (evt.type === 'progress' && evt.tasks) {
      taskItems.value = evt.tasks;
      if (evt.done !== undefined) doneCount.value = evt.done;
      if (evt.total !== undefined) totalCount.value = evt.total;
    } else if (evt.type === 'error') {
      addLog('error', evt.error || '');
    } else if (evt.type === 'continue_queued') {
      // 任务完成，已自动触发排队消息的处理
      chatHistory.value = [...chatHistory.value, {
        role: 'assistant' as const,
        content: '检测到您之前提交的反馈，正在自动处理...',
        timestamp: new Date().toISOString(),
      }];
    } else if (evt.type === 'continue_complete' || evt.type === 'complete') {
      stopPolling();
      // Extract assistant reply from log
      const lastAnswer = logLines.value.filter(l => l.kind === 'answer').pop();
      chatHistory.value = [...chatHistory.value, {
        role: 'assistant' as const,
        content: lastAnswer?.text || '处理完成',
        timestamp: new Date().toISOString(),
      }];
      chatLoading.value = false;
      chatEs?.close();
      chatEs = null;
      refreshTask(taskId);
    }
  };

  chatEs.addEventListener('answer', handler);
  chatEs.addEventListener('tool_call', handler);
  chatEs.addEventListener('progress', handler);
  chatEs.addEventListener('file_ready', handler);
  chatEs.addEventListener('error', handler);
  chatEs.addEventListener('continue_complete', handler);
  chatEs.addEventListener('complete', handler);
}

// ── Online Preview Modal ────────────────────────────────────────────────
const previewTask = ref<TaskItem | null>(null);
const previewThumbUrl = ref('');
const previewVisible = ref(false);

function openPreview(task: TaskItem, thumbUrl: string) {
  previewTask.value = task;
  previewThumbUrl.value = thumbUrl;
  previewVisible.value = true;
}

function closePreview() {
  previewVisible.value = false;
  previewTask.value = null;
}

function downloadFromPreview() {
  if (!previewTask.value || !selectedId.value) return;
  const name = previewTask.value.output_file.split(/[/\\]/).pop() || previewTask.value.output_file;
  const a = document.createElement('a');
  a.href = `/api/tasks/${selectedId.value}/files/${encodeURIComponent(name)}`;
  a.download = name;
  a.style.display = 'none';
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
}

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
  // 1. Add all entries from taskItems (progress events)
  for (const t of taskItems.value) {
    slides.set(t.output_file, { task: t, fileReady: finalFiles.value.includes(t.output_file) });
  }
  // 2. Add files that arrived via file_ready but don't yet have a taskItem entry
  // This ensures slides appear immediately when the .pptx file is detected on disk,
  // without waiting for the master agent to update tasks.json
  for (const f of finalFiles.value) {
    if (!taskItems.value.some(t => t.output_file === f)) {
      const name = f.split(/[/\\]/).pop() || f;
      const idxMatch = name.match(/^(\d+)_/);
      const pageIdx = idxMatch ? parseInt(idxMatch[1]) : slides.size + 1;
      slides.set(f, {
        task: {
          task_id: f,
          page_index: pageIdx,
          title: name,
          content_type: '',
          output_file: f,
          status: 'done',
        },
        fileReady: true,
      });
    }
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
  if (slides.length === 0) return;

  // Download all selected files synchronously within the same user gesture.
  // setTimeout breaks the gesture chain and triggers browser popup blocking.
  for (const s of slides) {
    const name = s.task.output_file.split(/[/\\]/).pop() || s.task.output_file;
    const a = document.createElement('a');
    a.href = `/api/tasks/${selectedTask.value!.id}/files/${encodeURIComponent(name)}`;
    a.download = name;
    a.style.display = 'none';
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
  }
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
let sseRetryCount = 0;
let sseCompleted = false;
const SSE_MAX_RETRIES = 10;

function connectSSE(taskId: string) {
  if (!taskId) return;
  if (sseCompleted) return; // already received complete, don't reconnect
  if (es) es.close();
  es = new EventSource(`/api/tasks/${taskId}/stream`);
  activeWorkers.value = 0;

  es.onerror = () => {
    es!.close();
    es = null;

    // Already received complete event — don't retry
    if (sseCompleted) return;

    const task = tasks.value.find(t => t.id === taskId);

    // Running task: keep retrying indefinitely with backoff (master agent may be idle/thinking)
    if (task?.status === 'running') {
      sseRetryCount++;
      const delay = Math.min(2000 * Math.pow(1.5, sseRetryCount - 1), 15000);
      addLog('error', `连接中断，${(delay / 1000).toFixed(0)}s 后自动重连 (第 ${sseRetryCount} 次)...`);
      setTimeout(() => connectSSE(taskId), delay);
      return;
    }

    // Completed/cancelled task: limited retries
    if (task && sseRetryCount < 3) {
      sseRetryCount++;
      setTimeout(() => connectSSE(taskId), 2000);
      return;
    }

    // Only force logout if we can confirm a 401 via auth check (skip in dev mode)
    if (!import.meta.env.DEV && !isLoggedIn()) {
      clearToken();
      window.location.href = '/auth';
    } else {
      addLog('error', 'SSE 连接失败，请刷新页面重试');
    }
  };

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

      case 'token_usage':
        if (evt.total_tokens) {
          const t = tasks.value.find(x => x.id === taskId);
          if (t) {
            t.prompt_tokens = evt.prompt_tokens || 0;
            t.completion_tokens = evt.completion_tokens || 0;
            t.total_tokens = evt.total_tokens || 0;
          }
        }
        break;

      case 'error':
        addLog('error', evt.error || evt.content || '');
        break;

      case 'complete':
        sseCompleted = true;
        doneCount.value = evt.done || 0;
        totalCount.value = evt.total || 0;
        if (evt.files) {
          for (const f of evt.files) {
            if (!finalFiles.value.includes(f)) finalFiles.value.push(f);
            cachePPT(taskId, f);
          }
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
  es.addEventListener('token_usage', handler);
  es.addEventListener('error', handler);
  es.addEventListener('complete', handler);
}

function disconnectSSE() {
  if (es) { es.close(); es = null; }
  stopPolling();
}

// ── Polling fallback ────────────────────────────────────────────────────
// Polls task state every 3s to fill gaps when SSE events are lost.
let pollTimer: ReturnType<typeof setInterval> | null = null;

function startPolling(taskId: string) {
  stopPolling();
  pollTimer = setInterval(async () => {
    try {
      const info = await fetchTask(taskId);
      if (!info) return;
      // Update file list from API (catches files missed by SSE file_ready)
      if (info.files?.length) {
        for (const f of info.files) {
          if (!finalFiles.value.includes(f)) {
            finalFiles.value = [...finalFiles.value, f];
          }
        }
      }
      doneCount.value = info.done_count;
      totalCount.value = info.total_count;
      // Stop polling when task finishes
      if (info.status !== 'running') {
        stopPolling();
      }
    } catch { /* ignore fetch errors during polling */ }
  }, 3000);
}

function stopPolling() {
  if (pollTimer) { clearInterval(pollTimer); pollTimer = null; }
}

// Restore log lines and files from conversation_content (cold start).
function restoreFromConversation(sess: import('../types').ConversationSession) {
  if (!sess.conversation_content) return;

  const lines: { ts: number; text: string; kind: import('../types').LogKind }[] = [];
  const linesArr = sess.conversation_content.split('\n');

  for (const raw of linesArr) {
    const line = raw.trim();
    if (!line) continue;
    if (line.startsWith('**助手**')) {
      lines.push({ ts: Date.now(), text: line.replace('**助手**:', '').trim(), kind: 'answer' });
    } else if (line.startsWith('**用户**')) {
      lines.push({ ts: Date.now(), text: line.replace('**用户**:', '').trim(), kind: 'answer' });
    } else if (line.startsWith('**错误**')) {
      lines.push({ ts: Date.now(), text: line.replace('**错误**:', '').trim(), kind: 'error' });
    } else if (line.startsWith('**完成摘要**')) {
      lines.push({ ts: Date.now(), text: line.replace('**完成摘要**:', '').trim(), kind: 'answer' });
    } else if (line.startsWith('**生成文件**')) {
      lines.push({ ts: Date.now(), text: line, kind: 'file' });
    } else if (line.startsWith('##') || line.startsWith('|')) {
      // 任务信息 / 幻灯片概览表格：不加入日志流，但保留可读性
    }
  }

  if (lines.length > 0) {
    logLines.value = lines;
  }
}

// Cache PPT files in background for completed tasks (no SSE connection).
async function preloadFiles(taskId: string, files?: string[]) {
  if (!files?.length) return;
  for (const f of files) {
    await cachePPT(taskId, f);
  }
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
      sseRetryCount = 0;
      sseCompleted = false;
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
  sseRetryCount = 0;
  sseCompleted = false;

  if (t.status !== 'failed') {
    if (t.status === 'running') {
      // 正在运行：通过 SSE 实时接收事件 + polling 兜底
      connectSSE(id);
      startPolling(id);
    } else {
      // 已完成 / 已取消：后端任务已不在内存，SSE 无事件可重放。
      // 改用 /conversation 接口从数据库恢复日志区和文件。
      fetchConversation(id).then(sess => {
        if (sess.conversation_content) {
          restoreFromConversation(sess);
          if (sess.duration) duration.value = sess.duration;
        }
        if (sess.files?.length) {
          finalFiles.value = sess.files;
        }
        if (sess.done_count !== undefined) doneCount.value = sess.done_count;
        if (sess.total_count !== undefined) totalCount.value = sess.total_count;
        // 预缓存所有 PPT 文件（供后续预览 / 下载）。
        preloadFiles(id, sess.files);
      }).catch(() => { /* ignore */ });
    }
  }
}

async function handleCreateTask(query: string) {
  if (creating.value) return;
  creating.value = true;
  loadError.value = '';
  try {
    const info = await createTask(query);
    tasks.value = [info, ...tasks.value];
    selectTask(info.id);
  } catch (e: any) {
    loadError.value = e?.message || '创建任务失败，请重试';
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
  } catch (e: any) {
    loadError.value = e?.message || '取消任务失败';
  } finally {
    cancelling.value = false;
  }
}

function handleNewSession() {
  // Clear current selection to return to welcome view
  if (selectedId.value) {
    saveCache(selectedId.value);
    disconnectSSE();
    if (chatEs) { chatEs.close(); chatEs = null; }
    selectedId.value = null;
  }
  // Scroll main to top
  document.querySelector('.main')?.scrollTo({ top: 0 });
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
  try {
    tasks.value = await fetchTasks();
    loadError.value = '';
  } catch (e: any) {
    loadError.value = e?.message || '加载任务列表失败，请刷新重试';
  }
}

onMounted(async () => {
  if (!import.meta.env.DEV) {
    if (!isLoggedIn()) { router.push('/auth'); return; }
    await auth.init();
    if (!auth.loggedIn) { router.push('/auth'); return; }
  }
  await loadTasks();
  // Auto-select task from query param (e.g. navigated from ComposePage)
  const selectId = router.currentRoute.value.query?.select as string;
  if (selectId && tasks.value.some(t => t.id === selectId)) {
    nextTick(() => selectTask(selectId));
    router.replace({ query: {} }); // clean URL so reload doesn't re-select
  }
});

onUnmounted(() => { disconnectSSE(); stopPolling(); if (chatEs) { chatEs.close(); chatEs = null; } });
</script>

<template>
  <div class="layout">
    <Sidebar
      :user="auth.user"
      :tasks="tasks"
      :selected-id="selectedId"
      :has-running-task="hasRunningTask"
      :creating="creating"
      :error="loadError"
      @logout="auth.logout(); router.push('/')"
      @select-task="selectTask"
      @create-task="handleCreateTask"
      @delete-task="handleDeleteTask"
      @compose="router.push('/compose')"
      @new-session="handleNewSession"
    />

    <main class="main" id="main-content">
      <!-- Welcome -->
      <div v-if="!selectedTask" class="welcome">
        <div class="welcome-hero">
          <div class="welcome-icon">
            <svg viewBox="0 0 120 120" fill="none">
              <rect x="12" y="18" width="96" height="72" rx="8" fill="var(--accent-soft)" stroke="var(--accent-border)" stroke-width="2"/>
              <rect x="12" y="18" width="96" height="16" rx="8" fill="var(--accent)"/>
              <circle cx="28" cy="26" r="3" fill="#fff" opacity="0.9"/>
              <circle cx="38" cy="26" r="3" fill="#fff" opacity="0.7"/>
              <circle cx="48" cy="26" r="3" fill="#fff" opacity="0.5"/>
              <rect x="22" y="44" width="40" height="5" rx="2.5" fill="var(--accent-border)"/>
              <rect x="22" y="54" width="56" height="5" rx="2.5" fill="var(--border)"/>
              <rect x="22" y="64" width="48" height="5" rx="2.5" fill="var(--border)"/>
              <rect x="22" y="74" width="32" height="5" rx="2.5" fill="var(--border)"/>
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

        <!-- Template Compose -->
        <div class="compose-entry">
          <h3>高级编排 — 自由组合模板</h3>
          <p>从预设模板或原子布局开始，自由拖拽组合幻灯片，精细控制每页内容</p>
          <button class="compose-btn" @click="router.push({ name: 'compose' })">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="3" y="3" width="18" height="18" rx="2"/><path d="M3 9h18M9 21V9"/>
            </svg>
            打开模板编排
          </button>
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
              class="chat-trigger-btn"
              @click="openChatBar"
              title="与 AI 聊天改进 PPT"
            >
              <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true">
                <path d="M14 10c0 .5-.3 1-.7 1.2L9.5 13 6 11.5V10c0-.5.3-1 .7-1.2l3.8-1.8L14 10z"/>
                <path d="M2 10c0 4.4 3.6 8 8 8s8-3.6 8-8-3.6-8-8-8-8 3.6-8 8z"/>
              </svg>
              AI 对话
            </button>
            <button
              v-if="selectedTask?.status === 'running'"
              class="cancel-btn"
              :disabled="cancelling"
              @click="handleCancel"
            >{{ cancelling ? '取消中...' : '中断' }}</button>
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
            <EventLog :lines="logLines" />
          </div>
          <div class="split-right">
            <!-- Batch panels -->
            <TransitionGroup name="batch" tag="div">
              <div v-for="batch in batches" :key="batch.id" class="batch-panel" :class="{ done: batch.done }">
                <div class="batch-header">
                  <span class="batch-label">
                    <span v-if="!batch.done" class="batch-spinner" aria-hidden="true"></span>
                    <span>批次 #{{ batch.id }}</span>
                  </span>
                  <span class="batch-meta" :class="batch.done ? 'done' : 'running'">{{ batch.done ? '已完成' : '执行中' }} · {{ batch.taskIds.length }} 页</span>
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
                  @preview="openPreview"
                />
              </TransitionGroup>
            </div>
          </div>
        </div>

        <!-- Final message -->
        <div v-if="finalMessage && selectedTask?.status !== 'running'" class="final-message">
          <div class="final-header">
            <span class="final-icon" aria-hidden="true">
              <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="2.5">
                <path d="M2 7l3.5 3.5L12 4"/>
              </svg>
            </span>
            <h3>任务完成</h3>
          </div>
          <p>{{ finalMessage }}</p>
        </div>
      </template>
    </main>

    <!-- Online PPT Preview Modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="previewVisible" class="preview-modal-overlay" @click.self="closePreview">
          <div class="preview-modal">
            <div class="preview-modal-header">
              <h3>{{ previewTask?.title || '幻灯片预览' }}</h3>
              <div class="preview-modal-actions">
                <button class="modal-action-btn" title="下载 PPTX" @click="downloadFromPreview">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/>
                    <polyline points="7 10 12 15 17 10"/>
                    <line x1="12" y1="15" x2="12" y2="3"/>
                  </svg>
                  下载 PPTX
                </button>
                <button class="modal-close-btn" @click="closePreview">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
                  </svg>
                </button>
              </div>
            </div>
            <div class="preview-modal-body">
              <img
                v-if="previewThumbUrl"
                :src="previewThumbUrl"
                :alt="previewTask?.title"
                class="preview-modal-img"
              />
              <div v-else class="preview-modal-loading">
                <span class="preview-spinner"></span>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- Persistent Chat Bar (bottom of main) -->
    <div v-if="selectedTask" class="chat-bar" :class="{ expanded: showChatInput }">
      <div v-if="!showChatInput" class="chat-bar-collapsed" @click="openChatBar">
        <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true">
          <path d="M14 10c0 .5-.3 1-.7 1.2L9.5 13 6 11.5V10c0-.5.3-1 .7-1.2l3.8-1.8L14 10z"/>
          <path d="M2 10c0 4.4 3.6 8 8 8s8-3.6 8-8-3.6-8-8-8-8 3.6-8 8z"/>
        </svg>
        <span>与 AI 对话，改进 PPT 内容</span>
      </div>
      <div v-else class="chat-bar-expanded">
        <div class="chat-bar-header">
          <span class="chat-bar-title">AI 对话</span>
          <button class="chat-bar-close" @click="closeChatBar" aria-label="关闭">
            <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5">
              <line x1="4" y1="4" x2="12" y2="12"/><line x1="12" y1="4" x2="4" y2="12"/>
            </svg>
          </button>
        </div>
        <div class="chat-messages" ref="chatMessagesEl">
          <div v-if="chatHistory.length === 0 && !chatLoading" class="chat-empty">
            对这个 PPT 有什么想改进的？直接告诉我吧
          </div>
          <div
            v-for="(msg, idx) in chatHistory"
            :key="idx"
            :class="['chat-msg', msg.role]"
          >
            <div class="chat-msg-bubble">{{ msg.content }}</div>
          </div>
          <div v-if="chatLoading" class="chat-msg assistant">
            <div class="chat-msg-bubble typing">
              <span class="typing-dot"></span><span class="typing-dot"></span><span class="typing-dot"></span>
            </div>
          </div>
        </div>
        <div class="chat-input-row">
          <input
            v-model="chatInput"
            class="chat-input"
            placeholder="描述想改进的内容，按 Enter 发送..."
            :disabled="chatLoading"
            @keydown.enter.prevent="submitChat"
            aria-label="聊天输入"
          />
          <button
            class="chat-send-btn"
            :disabled="!chatInput.trim() || chatLoading"
            @click="submitChat"
            aria-label="发送"
          >
            <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="2" y1="8" x2="14" y2="8"/><polyline points="10,4 14,8 10,12"/>
            </svg>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.layout { display: flex; height: 100vh; overflow: hidden; }
.main {
  flex: 1;
  background: var(--bg-muted);
  padding: 1.5rem 2rem;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}

/* Welcome */
.welcome {
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  min-height: 60vh; text-align: center;
  max-width: 680px;
  margin: 0 auto;
  width: 100%;
}
.welcome-hero { animation: fadeInUp 0.6s cubic-bezier(0.4, 0, 0.2, 1); }
.welcome-icon svg { width: 100px; height: 100px; margin: 0 auto 1.25rem; }
.welcome h2 {
  font-size: 1.5rem; font-weight: 700; margin-bottom: 0.5rem;
  background: linear-gradient(135deg, var(--text) 0%, var(--text-secondary) 100%);
  -webkit-background-clip: text; -webkit-text-fill-color: transparent;
}
.welcome p { color: var(--text-muted); font-size: 0.9rem; }

/* Welcome examples */
.welcome-examples { max-width: 650px; margin: 0 auto 2rem; text-align: center; }
.welcome-examples h3 { font-size: 0.85rem; font-weight: 600; color: var(--text-secondary); margin-bottom: 0.75rem; }
.examples-grid { display: flex; flex-wrap: wrap; gap: 0.5rem; justify-content: center; }
.example-chip {
  padding: 0.45rem 0.9rem; border-radius: var(--radius-full);
  border: 1px solid var(--border); background: var(--bg-base);
  color: var(--text-secondary); font-size: 0.78rem; cursor: pointer;
  transition: all var(--transition); font-family: inherit;
}
.example-chip:hover {
  border-color: var(--accent); color: var(--accent);
  background: var(--accent-soft); transform: translateY(-1px);
}

/* Compose entry */
.compose-entry { max-width: 650px; margin: 0 auto 2rem; text-align: center; }
.compose-entry h3 { font-size: 0.85rem; font-weight: 600; color: var(--text-secondary); margin-bottom: 0.5rem; }
.compose-entry p { font-size: 0.78rem; color: var(--text-muted); margin-bottom: 0.75rem; }
.compose-btn {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 8px 20px; border-radius: var(--radius);
  background: var(--accent); color: white; border: none;
  font-size: 0.85rem; font-weight: 500;
  cursor: pointer; transition: all var(--transition-md);
  font-family: inherit;
  box-shadow: 0 2px 8px rgba(99,102,241,0.25);
}
.compose-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 6px 20px rgba(99,102,241,0.35);
  background: var(--accent-hover);
}

/* Welcome themes */
.welcome-themes { max-width: 600px; margin: 0 auto; text-align: center; }
.welcome-themes h3 { font-size: 0.85rem; font-weight: 600; color: var(--text-secondary); margin-bottom: 0.75rem; }
.themes-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 0.6rem; }
.theme-card {
  background: var(--bg-base); border: 1px solid var(--border);
  border-radius: var(--radius); padding: 0.75rem 0.5rem;
  text-align: center; transition: transform var(--transition), box-shadow var(--transition);
  cursor: default;
}
.theme-card:hover { transform: translateY(-2px); box-shadow: var(--shadow); }
.theme-swatch {
  width: 40px; height: 40px; border-radius: var(--radius);
  margin: 0 auto 0.4rem; display: flex; align-items: center; justify-content: center;
  font-size: 1.1rem;
}
.theme-name { display: block; font-size: 0.75rem; font-weight: 600; color: var(--text); margin-bottom: 0.1rem; }
.theme-desc { font-size: 0.65rem; color: var(--text-muted); }

/* ── Dashboard working area ───────────────────────────────────── */
.dash-header {
  display: flex; justify-content: space-between; align-items: flex-start;
  margin-bottom: 1rem; animation: fadeInUp 0.4s both;
  background: var(--bg-base);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 1rem 1.25rem;
  box-shadow: var(--shadow-xs);
}
.dash-header-left { flex: 1; min-width: 0; max-width: calc(100% - 200px); }
.dash-title {
  font-size: 1rem; font-weight: 700; line-height: 1.4; color: var(--text);
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.dash-meta { display: flex; align-items: center; gap: 0.75rem; margin-top: 0.3rem; flex-wrap: wrap; }
.dash-id { font-size: 0.68rem; color: var(--text-muted); font-family: 'JetBrains Mono', monospace; background: var(--bg-muted); padding: 0.12rem 0.4rem; border-radius: 4px; }
.dash-duration { font-size: 0.72rem; color: var(--text-muted); display: flex; align-items: center; gap: 0.2rem; }
.dash-duration svg { width: 12px; height: 12px; }
.dash-tokens { font-size: 0.72rem; color: var(--text-muted); display: flex; align-items: center; gap: 0.2rem; }
.dash-tokens svg { width: 13px; height: 13px; }
.token-detail { font-size: 0.64rem; opacity: 0.7; }
.dash-status { display: flex; gap: 0.4rem; flex-shrink: 0; align-items: center; flex-wrap: wrap; }
.stat-badge { font-size: 0.72rem; font-weight: 600; padding: 0.3rem 0.7rem; border-radius: var(--radius-full); display: flex; align-items: center; gap: 0.3rem; }
.stat-badge.running { background: var(--info-soft); color: var(--info); }
.stat-badge.workers { background: var(--warning-soft); color: var(--warning); }
.stat-badge.done { background: var(--success-soft); color: var(--success); }
.stat-badge.cancelled { background: var(--warning-soft); color: var(--warning); }
.stat-badge.failed { background: var(--danger-soft); color: var(--danger); }
.pulse-dot { width: 7px; height: 7px; border-radius: 50%; background: currentColor; animation: pulse 1.5s infinite; }
.cancel-btn {
  font-size: 0.72rem; font-weight: 600; padding: 0.3rem 0.7rem;
  border-radius: var(--radius-full); border: 1px solid var(--danger-border);
  background: var(--danger-soft); color: var(--danger);
  cursor: pointer; display: flex; align-items: center; gap: 0.2rem;
  transition: all var(--transition); font-family: inherit;
}
.cancel-btn:hover { background: var(--danger); color: #fff; border-color: var(--danger); }
.cancel-btn:disabled { opacity: 0.5; cursor: not-allowed; }
/* ── Split layout with distinct section backgrounds ─────────── */
.split-layout {
  display: grid; grid-template-columns: 33% 67%; gap: 1rem;
  flex: 1; min-height: 0;
}
.split-left {
  background: var(--bg-base);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 1rem;
  box-shadow: var(--shadow-xs);
  min-width: 0; overflow-y: auto;
}
.split-right {
  min-width: 0; overflow-y: auto; padding-right: 0.3rem;
  display: flex; flex-direction: column; gap: 0.75rem;
}

/* Section title */
.section-title {
  font-size: 0.8rem; font-weight: 700; color: var(--text);
  margin-bottom: 0.6rem; display: flex; align-items: center; gap: 0.5rem;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid var(--border);
}
.file-count { font-size: 0.66rem; background: var(--bg-muted); padding: 0.15rem 0.45rem; border-radius: var(--radius-full); font-weight: 500; color: var(--text-muted); }

/* Slides grid */
.slides-section { margin-bottom: 1rem; }
.slides-toolbar { display: flex; align-items: center; justify-content: space-between; margin-bottom: 0.6rem; flex-wrap: wrap; gap: 0.5rem; }
.toolbar-actions { display: flex; gap: 0.4rem; }
.tool-btn {
  font-size: 0.7rem; font-weight: 500;
  padding: 0.3rem 0.65rem; border-radius: var(--radius-sm);
  border: 1px solid var(--border);
  background: var(--bg-base); color: var(--text-secondary);
  cursor: pointer; transition: all var(--transition); font-family: inherit;
}
.tool-btn:hover { border-color: var(--text-disabled); background: var(--bg-muted); }
.tool-btn.primary { background: var(--accent); color: #fff; border-color: var(--accent); }
.tool-btn.primary:hover { background: var(--accent-hover); border-color: var(--accent-hover); }
.tool-btn.primary:disabled { opacity: 0.5; cursor: not-allowed; }
.slides-list { display: flex; flex-direction: column; gap: 0.6rem; }

/* Batch panels */
.batch-panel {
  background: var(--bg-base); border: 1px solid var(--border);
  border-radius: var(--radius); padding: 0.75rem 0.9rem;
  box-shadow: var(--shadow-xs); transition: border-color var(--transition);
}
.batch-panel:not(.done) { border-color: var(--warning-border); box-shadow: var(--shadow-xs); }
.batch-panel.done { border-color: var(--success-border); }
.batch-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 0.5rem; font-size: 0.78rem; font-weight: 600; color: var(--text); }
.batch-label { display: flex; align-items: center; gap: 0.3rem; }
.batch-spinner { width: 11px; height: 11px; border: 2px solid var(--warning); border-top-color: transparent; border-radius: 50%; animation: spin 0.8s linear infinite; }
.batch-meta { font-size: 0.65rem; font-weight: 500; }
.batch-meta.running { color: var(--warning); }
.batch-meta.done { color: var(--success); }
.batch-cards { display: flex; gap: 0.4rem; flex-wrap: wrap; }
.batch-card {
  display: flex; align-items: center;
  background: var(--bg-muted); border: 1px solid var(--border);
  border-radius: var(--radius-sm); padding: 0.35rem 0.5rem; min-width: 140px;
  transition: all var(--transition);
}
.batch-card.generating { border-color: var(--warning); background: var(--warning-soft); }
.batch-card.done, .batch-card.qa_done, .batch-card.fixed { border-color: var(--success); background: var(--success-soft); }
.batch-card.failed { border-color: var(--danger); background: var(--danger-soft); }
.bcard-idx {
  width: 20px; height: 20px; border-radius: 50%;
  background: var(--border-light); color: var(--text-muted);
  font-size: 0.6rem; font-weight: 700; display: flex; align-items: center; justify-content: center;
  margin-right: 0.35rem; flex-shrink: 0;
}
.batch-card.generating .bcard-idx { background: var(--warning); color: #fff; }
.batch-card.done .bcard-idx, .batch-card.qa_done .bcard-idx, .batch-card.fixed .bcard-idx { background: var(--success); color: #fff; }
.bcard-body { flex: 1; min-width: 0; }
.bcard-title { font-size: 0.68rem; font-weight: 500; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.bcard-status { font-size: 0.58rem; font-weight: 600; margin-top: 0.05rem; color: var(--text-muted); }
.bcard-status.generating { color: var(--warning); }
.bcard-status.done, .bcard-status.qa_done, .bcard-status.fixed { color: var(--success); }

/* Final message */
.final-message {
  background: var(--success-soft); border: 1px solid var(--success-border);
  border-radius: var(--radius); padding: 1rem 1.1rem; margin-bottom: 0.75rem;
  animation: fadeInUp 0.4s both;
}
.final-header { display: flex; align-items: center; gap: 0.5rem; margin-bottom: 0.4rem; }
.final-icon {
  width: 22px; height: 22px; border-radius: 50%;
  background: var(--success); color: #fff;
  display: flex; align-items: center; justify-content: center;
  font-size: 0.65rem; font-weight: 700;
  flex-shrink: 0;
}
.final-icon svg { width: 10px; height: 10px; }
.final-header h3 { font-size: 0.85rem; font-weight: 600; color: var(--text); }
.final-message p { font-size: 0.78rem; color: var(--text-secondary); white-space: pre-wrap; max-height: 120px; overflow-y: auto; }

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

/* Online Preview Modal */
.preview-modal-overlay {
  position: fixed; inset: 0; z-index: 1000;
  background: rgba(15,23,42,0.5); backdrop-filter: blur(6px);
  display: flex; align-items: center; justify-content: center; padding: 2rem;
}
.preview-modal {
  background: var(--bg-base); border-radius: var(--radius-lg);
  max-width: 900px; width: 100%; max-height: 90vh;
  display: flex; flex-direction: column;
  box-shadow: 0 25px 50px rgba(15,23,42,0.2);
  overflow: hidden;
}
.preview-modal-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 1rem 1.25rem; border-bottom: 1px solid var(--border);
  gap: 1rem;
}
.preview-modal-header h3 { font-size: 0.95rem; font-weight: 600; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.preview-modal-actions { display: flex; align-items: center; gap: 0.5rem; flex-shrink: 0; }
.modal-action-btn {
  display: flex; align-items: center; gap: 0.35rem;
  padding: 0.4rem 0.8rem; border-radius: var(--radius-sm);
  border: 1px solid var(--border); background: var(--bg-base);
  color: var(--text-secondary); font-size: 0.75rem; font-weight: 500;
  cursor: pointer; transition: all var(--transition); font-family: inherit;
}
.modal-action-btn:hover { border-color: var(--accent); color: var(--accent); background: var(--accent-soft); }
.modal-action-btn svg { width: 14px; height: 14px; }
.modal-close-btn {
  width: 32px; height: 32px; border-radius: var(--radius-sm);
  border: 1px solid var(--border); background: var(--bg-base);
  color: var(--text-muted); cursor: pointer; display: flex;
  align-items: center; justify-content: center;
  transition: all var(--transition); font-family: inherit;
}
.modal-close-btn:hover { background: var(--danger-soft); border-color: var(--danger); color: var(--danger); }
.modal-close-btn svg { width: 16px; height: 16px; }
.preview-modal-body {
  flex: 1; overflow: auto; padding: 1.5rem;
  display: flex; align-items: center; justify-content: center;
  background: var(--bg-muted);
}
.preview-modal-img {
  max-width: 100%; max-height: calc(90vh - 120px);
  object-fit: contain; border-radius: var(--radius);
  box-shadow: var(--shadow-lg);
}
.preview-modal-loading { display: flex; align-items: center; justify-content: center; min-height: 200px; }
.preview-spinner {
  width: 36px; height: 36px;
  border: 3px solid var(--border); border-top-color: var(--accent);
  border-radius: 50%; animation: spin 0.7s linear infinite;
}

.modal-enter-active { transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1); }
.modal-leave-active { transition: all 0.2s ease-in; }
.modal-enter-from, .modal-leave-to { opacity: 0; }
.modal-enter-from .preview-modal, .modal-leave-to .preview-modal { transform: scale(0.95) translateY(10px); }

/* Continue Panel */
.continue-overlay {
  position: fixed; inset: 0; z-index: 1100;
  background: rgba(15,23,42,0.5); backdrop-filter: blur(6px);
  display: flex; align-items: center; justify-content: center; padding: 2rem;
}
.continue-panel {
  background: var(--bg-base); border-radius: var(--radius-lg);
  max-width: 600px; width: 100%; max-height: 80vh;
  display: flex; flex-direction: column;
  box-shadow: 0 25px 50px rgba(15,23,42,0.2); overflow: hidden;
}
.continue-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 1rem 1.25rem; border-bottom: 1px solid var(--border);
  gap: 1rem; flex-shrink: 0;
}
.continue-header h3 { font-size: 0.95rem; font-weight: 600; color: var(--text); }
.continue-close-btn {
  width: 32px; height: 32px; border-radius: var(--radius-sm);
  border: 1px solid var(--border); background: var(--bg-base);
  color: var(--text-muted); cursor: pointer; display: flex;
  align-items: center; justify-content: center;
  transition: all var(--transition); font-family: inherit;
}
.continue-close-btn:hover { background: var(--danger-soft); border-color: var(--danger); color: var(--danger); }
.continue-close-btn svg { width: 16px; height: 16px; }
.continue-history {
  flex: 1; overflow-y: auto; padding: 1rem;
  display: flex; flex-direction: column; gap: 0.75rem; max-height: 400px;
}
.continue-msg { display: flex; flex-direction: column; gap: 0.25rem; }
.continue-msg.user { align-items: flex-end; }
.continue-msg.assistant { align-items: flex-start; }
.continue-msg-role { font-size: 0.65rem; font-weight: 600; color: var(--text-muted); padding: 0 0.25rem; }
.continue-msg-content {
  font-size: 0.8rem; padding: 0.6rem 0.75rem; border-radius: var(--radius);
  line-height: 1.5; white-space: pre-wrap; word-break: break-word; max-width: 85%;
}
.continue-msg.user .continue-msg-content { background: var(--accent); color: white; }
.continue-msg.assistant .continue-msg-content { background: var(--bg-muted); color: var(--text); border: 1px solid var(--border); }
.continue-input-area { padding: 0.75rem 1rem; border-top: 1px solid var(--border); flex-shrink: 0; }
.continue-textarea {
  width: 100%; padding: 0.6rem 0.75rem; border: 1px solid var(--border);
  border-radius: var(--radius); background: var(--bg-muted); color: var(--text);
  font-size: 0.82rem; line-height: 1.5; resize: none;
  font-family: inherit; transition: border-color var(--transition);
}
.continue-textarea:focus { outline: none; border-color: var(--accent); }
.continue-textarea:disabled { opacity: 0.6; cursor: not-allowed; }
.continue-input-footer { display: flex; align-items: center; justify-content: space-between; margin-top: 0.5rem; }
.continue-hint { font-size: 0.65rem; color: var(--text-muted); }
.continue-submit-btn {
  padding: 0.4rem 1rem; border-radius: var(--radius);
  border: none; background: var(--accent); color: white;
  font-size: 0.8rem; font-weight: 600; cursor: pointer;
  transition: all var(--transition-md); font-family: inherit;
  box-shadow: 0 2px 8px rgba(99,102,241,0.25);
}
.continue-submit-btn:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 14px rgba(99,102,241,0.35);
  background: var(--accent-hover);
}
.continue-submit-btn:disabled { opacity: 0.5; cursor: not-allowed; }

/* Chat Bar */
.chat-bar {
  position: fixed; bottom: 0; right: 0;
  width: calc(100% - var(--sidebar-w));
  background: var(--bg-base);
  border-top: 1px solid var(--border);
  box-shadow: 0 -4px 20px rgba(15,23,42,0.08);
  z-index: 100;
  transition: height 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}
.chat-bar-collapsed {
  display: flex; align-items: center; gap: 0.5rem;
  padding: 0.6rem 1.25rem;
  cursor: pointer;
  color: var(--accent); font-size: 0.8rem; font-weight: 500;
  transition: background var(--transition);
}
.chat-bar-collapsed:hover { background: var(--accent-soft); }
.chat-bar-collapsed svg { width: 16px; height: 16px; flex-shrink: 0; }

.chat-bar-expanded { display: flex; flex-direction: column; max-height: 360px; }
.chat-bar-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 0.6rem 1rem; border-bottom: 1px solid var(--border);
  flex-shrink: 0;
}
.chat-bar-title { font-size: 0.82rem; font-weight: 600; color: var(--text); }
.chat-bar-close {
  width: 24px; height: 24px; border-radius: var(--radius-sm);
  border: 1px solid var(--border); background: transparent;
  color: var(--text-muted); cursor: pointer; display: flex;
  align-items: center; justify-content: center;
  transition: all var(--transition); font-family: inherit;
}
.chat-bar-close:hover { background: var(--danger-soft); color: var(--danger); border-color: var(--danger); }
.chat-bar-close svg { width: 12px; height: 12px; }

.chat-messages {
  flex: 1; overflow-y: auto; padding: 0.75rem 1rem;
  display: flex; flex-direction: column; gap: 0.5rem;
  max-height: 220px;
}
.chat-empty { font-size: 0.78rem; color: var(--text-muted); text-align: center; padding: 0.5rem; }
.chat-msg { display: flex; }
.chat-msg.user { justify-content: flex-end; }
.chat-msg.assistant { justify-content: flex-start; }
.chat-msg-bubble {
  max-width: 70%; padding: 0.45rem 0.7rem;
  border-radius: var(--radius); font-size: 0.78rem;
  line-height: 1.5; word-break: break-word;
}
.chat-msg.user .chat-msg-bubble { background: var(--accent); color: #fff; }
.chat-msg.assistant .chat-msg-bubble { background: var(--bg-muted); color: var(--text); border: 1px solid var(--border); }
.typing { display: flex; gap: 3px; align-items: center; padding: 0.5rem 0.7rem; }
.typing-dot {
  width: 6px; height: 6px; border-radius: 50%;
  background: var(--text-muted); animation: typingBounce 1.2s infinite;
}
.typing-dot:nth-child(2) { animation-delay: 0.2s; }
.typing-dot:nth-child(3) { animation-delay: 0.4s; }

.chat-input-row {
  display: flex; align-items: center; gap: 0.5rem;
  padding: 0.6rem 0.75rem; border-top: 1px solid var(--border);
  flex-shrink: 0;
}
.chat-input {
  flex: 1; padding: 0.5rem 0.75rem;
  border: 1px solid var(--border); border-radius: var(--radius);
  background: var(--bg-muted); color: var(--text); font-size: 0.8rem;
  outline: none; font-family: inherit; transition: border-color var(--transition);
}
.chat-input:focus { border-color: var(--accent); }
.chat-input:disabled { opacity: 0.6; }
.chat-send-btn {
  width: 32px; height: 32px; border-radius: var(--radius);
  border: none; background: var(--accent); color: #fff;
  cursor: pointer; display: flex; align-items: center; justify-content: center;
  transition: all var(--transition); flex-shrink: 0;
}
.chat-send-btn:hover:not(:disabled) { background: var(--accent-hover); transform: translateY(-1px); }
.chat-send-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.chat-send-btn svg { width: 14px; height: 14px; }

/* Adjust main content to not be hidden behind fixed chat bar */
.main { padding-bottom: 52px; }

/* Chat trigger button */
.chat-trigger-btn {
  display: flex; align-items: center; gap: 0.3rem;
  padding: 0.3rem 0.7rem; border-radius: var(--radius-full);
  border: none; background: var(--accent); color: #fff;
  font-size: 0.72rem; font-weight: 600; cursor: pointer;
  transition: all var(--transition-md); font-family: inherit;
  box-shadow: 0 2px 8px rgba(99,102,241,0.25);
}
.chat-trigger-btn svg { width: 14px; height: 14px; }
.chat-trigger-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 14px rgba(99,102,241,0.35);
  background: var(--accent-hover);
}

@keyframes typingBounce { 0%, 80%, 100% { transform: translateY(0); } 40% { transform: translateY(-4px); } }

.continue-panel-enter-active { transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1); }
.continue-panel-leave-active { transition: all 0.2s ease-in; }
.continue-panel-enter-from, .continue-panel-leave-to { opacity: 0; }
.continue-panel-enter-from .continue-panel, .continue-panel-leave-to .continue-panel { transform: scale(0.95) translateY(10px); }
</style>
