<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue';
import { useRouter } from 'vue-router';
import {
  Activity,
  Check,
  ChevronDown,
  Clock3,
  Coins,
  Download,
  LayoutPanelTop,
  ListFilter,
  MessageSquareText,
  Presentation,
  Send,
  Square,
  X,
} from 'lucide-vue-next';
import type { TaskInfo, TaskItem, SSEEvent, RuntimeMeta } from '../types';
import { STATUS_LABELS } from '../types';
import { fetchTasks, createTask, fetchTask, cancelTask, deleteTask, isLoggedIn, clearToken, continueTask, fetchConversation } from '../api';
import { authState } from '../stores/auth';
import AppShell from '../components/AppShell.vue';
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
const currentPhase = ref('');    // preparing/planning/generating/qa/fixing/complete
const phaseDetail = ref('');     // "生成第3/21页" / "读取模板" 等
const logLines = ref<{ ts: number; text: string; kind: import('../types').LogKind }[]>([]);
const finalFiles = ref<string[]>([]);
const finalMessage = ref('');
const duration = ref('');
const activeWorkers = ref(0);
const runtimeMeta = ref<RuntimeMeta | null>(null);
const thumbnailVersions = ref<Record<string, number>>({});
const thumbnailFailures = ref<Record<string, string>>({});
const cancelling = ref(false);
const creating = ref(false);
const loadError = ref('');
const sidebarOpen = ref(false);

function closeSidebar() {
  sidebarOpen.value = false;
}

function handleSidebarSelect(id: string) {
  closeSidebar();
  selectTask(id);
}

function handleSidebarCreate(query: string) {
  closeSidebar();
  void handleCreateTask(query);
}

function handleSidebarCompose() {
  closeSidebar();
  router.push('/compose');
}

function handleSidebarLogout() {
  closeSidebar();
  auth.logout();
  router.push('/');
}

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
	addLog('error', '对话连接暂时中断，正在自动重连...');
  };

  const handler = (e: MessageEvent) => {
    let evt: SSEEvent;
    try { evt = JSON.parse(e.data); } catch { return; }

	if (evt.type === 'answer') {
	  addLog('answer', evt.content || '');
	} else if (evt.type === 'system_step') {
	  addLog('worker', evt.content || '正在处理任务');
    } else if (evt.type === 'tool_call') {
      addLog('tool', `[${evt.tool_name || 'tool'}]`);
    } else if (evt.type === 'file_ready' && evt.files) {
      for (const f of evt.files) {
        if (!finalFiles.value.includes(f)) {
          finalFiles.value = [...finalFiles.value, f];
          addLog('file', `${f} 已更新`);
        }
      }
    } else if (evt.type === 'progress' && evt.tasks) {
      taskItems.value = evt.tasks;
      if (evt.done !== undefined) doneCount.value = evt.done;
      if (evt.total !== undefined) totalCount.value = evt.total;
	} else if (evt.type === 'runtime_meta' && evt.runtime_meta) {
	  runtimeMeta.value = evt.runtime_meta;
	} else if (evt.type === 'thumbnail_ready' && evt.files) {
	  markThumbnailReady(evt.files);
	} else if (evt.type === 'thumbnail_error' && evt.files) {
	  markThumbnailFailed(evt.files, '缩略图转换失败，可重试或直接下载 PPTX');
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
  chatEs.addEventListener('system_step', handler);
  chatEs.addEventListener('tool_call', handler);
  chatEs.addEventListener('progress', handler);
  chatEs.addEventListener('file_ready', handler);
  chatEs.addEventListener('runtime_meta', handler);
  chatEs.addEventListener('thumbnail_ready', handler);
  chatEs.addEventListener('thumbnail_error', handler);
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

function fmtElapsed(ms?: number): string {
  if (!ms || ms < 0) return '0s';
  const seconds = Math.floor(ms / 1000);
  if (seconds < 60) return `${seconds}s`;
  const minutes = Math.floor(seconds / 60);
  const rest = seconds % 60;
  if (minutes < 60) return `${minutes}m ${rest}s`;
  const hours = Math.floor(minutes / 60);
  return `${hours}h ${minutes % 60}m`;
}

function sumCounts(values?: Record<string, number>): number {
  if (!values) return 0;
  return Object.values(values).reduce((sum, n) => sum + n, 0);
}

const runtimeToolTotal = computed(() => sumCounts(runtimeMeta.value?.tool_calls));
const runtimeErrorTotal = computed(() => sumCounts(runtimeMeta.value?.tool_errors));
const runtimeQATotal = computed(() =>
  (runtimeMeta.value?.qa_high_issues || 0)
  + (runtimeMeta.value?.qa_medium_issues || 0)
  + (runtimeMeta.value?.qa_low_issues || 0)
);
const runtimeWarnings = computed(() => runtimeMeta.value?.budget_warnings || []);
const runtimeTimeline = computed(() => (runtimeMeta.value?.recent_events || []).slice(-24).reverse());

function eventStatusLabel(status?: string): string {
  if (!status) return 'ok';
  return status;
}

const sampleQueries = [
  '做一个关于新能源汽车的行业分析报告',
  '制作一个产品发布会演示文稿',
  '做一个 AI 大模型技术分享的 PPT',
  '制作一个公司季度总结汇报',
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
let sseCompleted = false;
let sseConnectionInterrupted = false;

// 追踪上一条 answer 事件的完整内容，用于计算增量。
let lastAnswerContent = '';

function connectSSE(taskId: string) {
  if (!taskId) return;
  if (sseCompleted) return; // already received complete, don't reconnect
  if (es) es.close();
  lastAnswerContent = '';
  es = new EventSource(`/api/tasks/${taskId}/stream`);
  activeWorkers.value = 0;

  es.onopen = () => {
	if (sseConnectionInterrupted) {
	  addLog('worker', '实时连接已恢复');
	}
	sseConnectionInterrupted = false;
  };

  es.onerror = () => {
	if (sseCompleted || sseConnectionInterrupted) return;
	sseConnectionInterrupted = true;
	addLog('error', '实时连接暂时中断，浏览器正在自动重连；任务状态仍会通过轮询同步');
  };

  const handler = (e: MessageEvent) => {
    let evt: SSEEvent;
    try { evt = JSON.parse(e.data); } catch { return; }

    switch (evt.type) {
	  case 'answer': {
        const newContent = evt.content || '';
        if (newContent) {
          // evt.content 是从本轮开头的完整累加文本，只取增量部分追加。
          const delta = newContent.startsWith(lastAnswerContent)
            ? newContent.slice(lastAnswerContent.length)
            : newContent;
          if (delta) {
            if (logLines.value.length === 0 || logLines.value[logLines.value.length - 1].kind !== 'answer') {
              addLog('divider', '── AI 响应 ──');
            }
            addLog('answer', delta);
          }
          lastAnswerContent = newContent;
        }
		break;
	  }

	  case 'system_step':
		if (evt.content) addLog('worker', evt.content);
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
        if (evt.phase) currentPhase.value = evt.phase;
        if (evt.phase_detail) phaseDetail.value = evt.phase_detail;
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
              }
            }
          }
        }
		if (evt.done !== undefined) doneCount.value = evt.done;
		if (evt.total !== undefined) totalCount.value = evt.total;
		const summary = tasks.value.find(task => task.id === taskId);
		if (summary) {
		  if (evt.done !== undefined) summary.done_count = evt.done;
		  if (evt.total !== undefined) summary.total_count = evt.total;
		}
		break;

	  case 'file_ready':
        if (evt.files) {
          for (const f of evt.files) {
            if (!finalFiles.value.includes(f)) {
              finalFiles.value = [...finalFiles.value, f];
              addLog('file', `${f} 已生成`);
            }
          }
        }
		break;

	  case 'thumbnail_ready':
		if (evt.files) markThumbnailReady(evt.files);
		break;

	  case 'thumbnail_error':
		if (evt.files) markThumbnailFailed(evt.files, '缩略图转换失败，可重试或直接下载 PPTX');
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

      case 'runtime_meta':
        if (evt.runtime_meta) {
          runtimeMeta.value = evt.runtime_meta;
        }
        break;

      case 'error':
        addLog('error', evt.error || evt.content || '');
        break;

      case 'complete':
        sseCompleted = true;
        lastAnswerContent = '';
        doneCount.value = evt.done || 0;
        totalCount.value = evt.total || 0;
        if (evt.files) {
          for (const f of evt.files) {
            if (!finalFiles.value.includes(f)) finalFiles.value.push(f);
          }
        }
        if (evt.message) finalMessage.value = evt.message;
        if (evt.duration) duration.value = evt.duration;
        if (evt.tasks) taskItems.value = evt.tasks;
        if (evt.runtime_meta) runtimeMeta.value = evt.runtime_meta;
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
  es.addEventListener('system_step', handler);
  es.addEventListener('tool_call', handler);
  es.addEventListener('progress', handler);
  es.addEventListener('file_ready', handler);
  es.addEventListener('thumbnail_ready', handler);
  es.addEventListener('thumbnail_error', handler);
  es.addEventListener('token_usage', handler);
  es.addEventListener('runtime_meta', handler);
  es.addEventListener('error', handler);
  es.addEventListener('complete', handler);
}

function disconnectSSE() {
  if (es) { es.close(); es = null; }
  stopPolling();
  lastAnswerContent = '';
  currentPhase.value = '';
  phaseDetail.value = '';
  runtimeMeta.value = null;
  sseConnectionInterrupted = false;
}

function markThumbnailReady(files: string[]) {
  const versions = { ...thumbnailVersions.value };
  const failures = { ...thumbnailFailures.value };
  for (const file of files) {
	const key = thumbnailKey(file);
	versions[key] = (versions[key] || 0) + 1;
	delete failures[key];
  }
  thumbnailVersions.value = versions;
  thumbnailFailures.value = failures;
}

function markThumbnailFailed(files: string[], message: string) {
  const failures = { ...thumbnailFailures.value };
  for (const file of files) failures[thumbnailKey(file)] = message;
  thumbnailFailures.value = failures;
}

function thumbnailKey(file: string) {
  return file.split(/[/\\]/).pop() || file;
}

function thumbnailStatus(file: string): 'pending' | 'ready' | 'error' {
  const key = thumbnailKey(file);
  if (thumbnailFailures.value[key]) return 'error';
  if (thumbnailVersions.value[key]) return 'ready';
  return selectedTask.value?.status === 'running' ? 'pending' : 'ready';
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
	  const taskIndex = tasks.value.findIndex(task => task.id === taskId);
	  if (taskIndex >= 0) tasks.value[taskIndex] = info;
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
		sseCompleted = true;
		if (es) { es.close(); es = null; }
		duration.value = info.duration || duration.value;
		currentPhase.value = info.status === 'completed' ? 'complete' : info.status;
		stopPolling();
	  }
    } catch { /* ignore fetch errors during polling */ }
  }, 3000);
}

function stopPolling() {
  if (pollTimer) { clearInterval(pollTimer); pollTimer = null; }
}

// Restore log lines and files from conversation_content (cold start).
// 优先使用 full_answer（完整累积的 LLM 输出），回退到 conversation_content markdown 解析。
function restoreFromConversation(sess: import('../types').ConversationSession) {
  // 优先使用完整累积的 LLM 回答
  if (sess.full_answer) {
    logLines.value = [{ ts: Date.now(), text: sess.full_answer, kind: 'answer' }];
    return;
  }

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
  currentPhase: string;
  phaseDetail: string;
  logLines: { ts: number; text: string; kind: import('../types').LogKind }[];
  finalFiles: string[];
  finalMessage: string;
  duration: string;
  activeWorkers: number;
  batches: Batch[];
  runtimeMeta: RuntimeMeta | null;
}
const taskCache = new Map<string, TaskCache>();

function saveCache(id: string) {
  taskCache.set(id, {
    taskItems: [...taskItems.value],
    doneCount: doneCount.value,
    totalCount: totalCount.value,
    currentPhase: currentPhase.value,
    phaseDetail: phaseDetail.value,
    logLines: [...logLines.value],
    finalFiles: [...finalFiles.value],
    finalMessage: finalMessage.value,
    duration: duration.value,
    activeWorkers: activeWorkers.value,
    batches: [...batches.value],
    runtimeMeta: runtimeMeta.value ? { ...runtimeMeta.value } : null,
  });
}

function restoreCache(id: string): boolean {
  const c = taskCache.get(id);
  if (!c) return false;
  taskItems.value = c.taskItems;
  doneCount.value = c.doneCount;
  totalCount.value = c.totalCount;
  currentPhase.value = c.currentPhase || '';
  phaseDetail.value = c.phaseDetail || '';
  logLines.value = c.logLines;
  finalFiles.value = c.finalFiles;
  finalMessage.value = c.finalMessage;
  duration.value = c.duration;
  activeWorkers.value = c.activeWorkers;
  batches.value = c.batches;
  runtimeMeta.value = c.runtimeMeta;
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
	  sseCompleted = false;
	  connectSSE(id);
	  startPolling(id);
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
  runtimeMeta.value = null;
  selectedSlides.value = new Set();
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
  <AppShell
    :title="selectedTask?.query || '任务工作台'"
    eyebrow="生成与交付"
    content-class="dashboard-shell-content"
  >
    <template #actions>
      <span v-if="selectedTask" class="top-task-state" :class="selectedTask.status">
        <i aria-hidden="true"></i>
        {{ selectedTask.status === 'running' ? '运行中' : selectedTask.status === 'completed' ? '已完成' : selectedTask.status === 'cancelled' ? '已中断' : '失败' }}
      </span>
      <button
        class="topbar-tool task-list-trigger"
        type="button"
        :aria-expanded="sidebarOpen"
        aria-label="打开任务列表"
        title="任务列表"
        @click="sidebarOpen = true"
      >
        <ListFilter :size="18" />
        <span>任务</span>
      </button>
      <button
        v-if="selectedTask"
        class="topbar-tool"
        type="button"
        title="继续修改"
        @click="openChatBar"
      >
        <MessageSquareText :size="18" />
        <span>继续修改</span>
      </button>
      <button
        v-if="selectedTask?.status === 'running'"
        class="topbar-tool danger"
        type="button"
        :disabled="cancelling"
        @click="handleCancel"
      >
        <Square :size="16" />
        <span>{{ cancelling ? '中断中' : '中断' }}</span>
      </button>
    </template>

    <div class="layout">
    <button
      v-if="sidebarOpen"
      class="sidebar-scrim"
      type="button"
      aria-label="关闭任务导航"
      @click="closeSidebar"
    ></button>
    <div class="sidebar-shell" :class="{ open: sidebarOpen }">
      <Sidebar
        :user="auth.user"
        :tasks="tasks"
        :selected-id="selectedId"
        :has-running-task="hasRunningTask"
        :creating="creating"
        :error="loadError"
        @logout="handleSidebarLogout"
        @select-task="handleSidebarSelect"
        @create-task="handleSidebarCreate"
        @delete-task="handleDeleteTask"
        @compose="handleSidebarCompose"
        @new-session="handleNewSession(); closeSidebar()"
      />
      <button class="sidebar-close-btn" type="button" aria-label="关闭任务导航" @click="closeSidebar">
        <X :size="20" />
      </button>
    </div>

    <main class="main" id="main-content">
      <!-- Welcome -->
      <div v-if="!selectedTask" class="welcome">
        <div class="welcome-hero">
          <div class="welcome-icon">
            <Presentation :size="30" :stroke-width="1.7" />
          </div>
          <span class="welcome-kicker">生成与交付</span>
          <h2>选择一个任务，或开始新的演示</h2>
          <p>页面生成后会在这里逐张出现，可随时预览、选择和下载。</p>
        </div>

        <!-- Quick start examples -->
        <div class="welcome-examples">
          <h3>常用起点</h3>
          <div class="examples-grid">
            <button v-for="ex in sampleQueries" :key="ex" class="example-chip" @click="handleCreateTask(ex)">{{ ex }}</button>
          </div>
        </div>

        <!-- Template Compose -->
        <div class="compose-entry">
          <span>
            <h3>需要先确定结构？</h3>
            <p>进入编排工作区，选择模板并调整每一页的大纲。</p>
          </span>
          <button class="compose-btn" @click="router.push({ name: 'compose' })">
            <LayoutPanelTop :size="17" />
            打开编排
          </button>
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
                <Clock3 :size="14" />
                {{ duration }}
              </span>
              <span v-if="(selectedTask?.total_tokens ?? 0) > 0" class="dash-tokens">
                <Coins :size="14" />
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
              <MessageSquareText :size="16" />
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
		  :phase="currentPhase"
		  :phase-detail="phaseDetail"
		  :task-id="selectedTask?.id"
		  :created-at="selectedTask?.created_at"
		/>

        <details v-if="runtimeMeta" class="dev-status-panel">
          <summary class="dev-status-summary">
            <span class="dev-status-title">运行诊断</span>
            <span class="dev-status-phase">{{ runtimeMeta.phase || currentPhase || 'preparing' }}</span>
            <span v-if="runtimeWarnings.length || runtimeMeta.last_error" class="diagnostic-alert">
              {{ runtimeWarnings.length + (runtimeMeta.last_error ? 1 : 0) }} 条警告
            </span>
            <ChevronDown class="diagnostic-chevron" :size="16" aria-hidden="true" />
          </summary>
          <div class="dev-status-body">
            <div class="dev-status-grid">
            <div class="dev-stat">
              <span class="dev-stat-label">运行</span>
              <strong>{{ fmtElapsed(runtimeMeta.elapsed_ms) }}</strong>
              <small>{{ runtimeMeta.phase_detail || phaseDetail || '等待阶段更新' }}</small>
            </div>
            <div class="dev-stat">
              <span class="dev-stat-label">工具</span>
              <strong>{{ runtimeToolTotal }}</strong>
              <small>错误 {{ runtimeErrorTotal }} · 同参 {{ runtimeMeta.same_tool_args_streak || 0 }}</small>
            </div>
            <div class="dev-stat">
              <span class="dev-stat-label">Token</span>
              <strong>{{ fmtTokens(runtimeMeta.total_tokens || selectedTask?.total_tokens || 0) }}</strong>
              <small>{{ fmtTokens(runtimeMeta.prompt_tokens || 0) }}p + {{ fmtTokens(runtimeMeta.completion_tokens || 0) }}c</small>
            </div>
            <div class="dev-stat">
              <span class="dev-stat-label">页面</span>
              <strong>{{ runtimeMeta.done_slides || doneCount }} / {{ runtimeMeta.total_slides || totalCount }}</strong>
              <small>缺失文件 {{ runtimeMeta.missing_files || 0 }}</small>
            </div>
            <div v-if="runtimeQATotal > 0" class="dev-stat">
              <span class="dev-stat-label">问题记录</span>
              <strong>{{ runtimeQATotal }}</strong>
              <small>H {{ runtimeMeta.qa_high_issues || 0 }} · M {{ runtimeMeta.qa_medium_issues || 0 }} · L {{ runtimeMeta.qa_low_issues || 0 }}</small>
            </div>
            <div class="dev-stat">
              <span class="dev-stat-label">压缩</span>
              <strong>{{ runtimeMeta.compression_saved_pct || '0%' }}</strong>
              <small>{{ fmtTokens(runtimeMeta.compression_before_tokens || 0) }} → {{ fmtTokens(runtimeMeta.compression_after_tokens || 0) }}</small>
            </div>
            </div>
            <div v-if="runtimeWarnings.length > 0 || (runtimeMeta.last_error || '')" class="dev-status-warnings">
            <span v-for="w in runtimeWarnings" :key="w" class="dev-warning">{{ w }}</span>
            <span v-if="runtimeMeta.last_error" class="dev-warning danger">{{ runtimeMeta.last_error }}</span>
          </div>
            <div v-if="runtimeTimeline.length > 0" class="runtime-timeline">
            <div class="timeline-head">
              <span>Timeline</span>
              <small>{{ runtimeTimeline.length }} recent events</small>
            </div>
            <div class="timeline-list">
              <div
                v-for="evt in runtimeTimeline"
                :key="evt.id"
                class="timeline-row"
                :class="evt.status || 'ok'"
              >
                <span class="timeline-time">{{ fmtElapsed(evt.elapsed_ms) }}</span>
                <span class="timeline-kind">{{ evt.kind }}</span>
                <span class="timeline-main">
                  <strong>{{ evt.name || evt.phase || 'task' }}</strong>
                  <small>{{ evt.phase || 'phase' }} · {{ eventStatusLabel(evt.status) }}</small>
                </span>
                <span v-if="evt.detail" class="timeline-detail">{{ evt.detail }}</span>
              </div>
            </div>
            </div>
          </div>
        </details>

        <!-- Left-Right Split -->
        <div class="split-layout">
          <details class="split-left activity-panel">
            <summary>
              <span><Activity :size="16" />任务活动</span>
              <small>{{ logLines.length }} 条</small>
              <ChevronDown :size="16" class="activity-chevron" />
            </summary>
            <EventLog :lines="logLines" max-height="320px" />
          </details>
          <div class="split-right">
            <!-- Batch panels -->
            <TransitionGroup name="batch" tag="div" class="batch-list">
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
                    <Download :size="15" />
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
                  :thumbnail-status="thumbnailStatus(s.task.output_file)"
                  :thumbnail-version="thumbnailVersions[thumbnailKey(s.task.output_file)] || 0"
                  :thumbnail-error="thumbnailFailures[thumbnailKey(s.task.output_file)]"
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
              <Check :size="14" :stroke-width="2.5" />
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
          <div class="preview-modal" role="dialog" aria-modal="true" aria-label="幻灯片预览">
            <div class="preview-modal-header">
              <h3>{{ previewTask?.title || '幻灯片预览' }}</h3>
              <div class="preview-modal-actions">
                <button class="modal-action-btn" title="下载 PPTX" @click="downloadFromPreview">
                  <Download :size="18" />
                  下载 PPTX
                </button>
                <button class="modal-close-btn" aria-label="关闭预览" @click="closePreview">
                  <X :size="20" />
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
      <button v-if="!showChatInput" class="chat-bar-collapsed" type="button" @click="openChatBar">
        <MessageSquareText :size="17" />
        <span>继续修改这套演示</span>
      </button>
      <div v-else class="chat-bar-expanded">
        <div class="chat-bar-header">
          <span class="chat-bar-title">AI 对话</span>
          <button class="chat-bar-close" @click="closeChatBar" aria-label="关闭">
            <X :size="18" />
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
            <Send :size="17" />
          </button>
        </div>
      </div>
    </div>
    </div>
  </AppShell>
</template>

<style scoped>
/* Rebuilt delivery workspace */
.dash-header { display: flex; justify-content: space-between; }
.dash-header-left { min-width: 0; }
.dash-meta { display: flex; align-items: center; flex-wrap: wrap; }
.dash-duration,
.dash-tokens { font-size: 11px; }

.slides-toolbar { display: flex; justify-content: space-between; gap: 12px; }
.toolbar-actions { display: flex; align-items: center; }
.tool-btn { border: 1px solid var(--border-strong); font: inherit; font-size: 11px; font-weight: 700; cursor: pointer; }

.batch-header { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
.batch-label { display: inline-flex; align-items: center; gap: 7px; color: var(--text-secondary); font-size: 11px; font-weight: 700; }
.batch-spinner { width: 13px; height: 13px; border: 2px solid var(--border-strong); border-top-color: var(--info); border-radius: 50%; animation: spin 0.9s linear infinite; }
.batch-meta { color: var(--text-muted); font-size: 10px; }.batch-meta.done { color: var(--success); }.batch-meta.running { color: var(--info); }
.batch-cards { display: flex; flex-wrap: wrap; }
.batch-card { min-width: min(200px, 100%); padding: 8px 10px; display: flex; align-items: center; gap: 9px; border: 1px solid var(--divider); }
.bcard-idx { width: 24px; height: 24px; display: grid; place-items: center; flex: 0 0 auto; border-radius: 4px; color: var(--text-muted); background: var(--surface-muted); font-size: 10px; font-weight: 750; }
.bcard-body { min-width: 0; display: flex; flex-direction: column; }
.bcard-title { overflow: hidden; color: var(--text-secondary); font-size: 11px; text-overflow: ellipsis; white-space: nowrap; }
.bcard-status { margin-top: 2px; color: var(--text-muted); font-size: 9px; }.bcard-status.generating { color: var(--info); }.bcard-status.done, .bcard-status.fixed, .bcard-status.qa_done { color: var(--success); }.bcard-status.failed { color: var(--danger); }

.dev-status-body { display: grid; gap: 10px; }
.dev-status-grid { display: grid; }
.dev-stat { display: flex; flex-direction: column; }
.dev-stat-label { color: var(--text-muted); font-size: 9px; font-weight: 750; }
.dev-stat strong { margin-top: 5px; color: var(--text); font-size: 15px; }.dev-stat small { margin-top: 3px; color: var(--text-muted); font-size: 9px; }
.dev-status-warnings { display: flex; flex-wrap: wrap; gap: 6px; }
.dev-warning { padding: 5px 7px; border-radius: 4px; color: var(--warning); background: var(--warning-soft); font-size: 10px; }.dev-warning.danger { color: var(--danger); background: var(--danger-soft); }
.runtime-timeline { overflow: hidden; }
.timeline-head { min-height: 38px; padding: 0 10px; display: flex; align-items: center; justify-content: space-between; border-bottom: 1px solid var(--divider); color: var(--text-secondary); font-size: 10px; font-weight: 700; }
.timeline-head small { color: var(--text-muted); font-size: 9px; font-weight: 500; }
.timeline-list { max-height: 280px; overflow: auto; }
.timeline-row { min-height: 42px; padding: 7px 10px; display: grid; grid-template-columns: 54px 90px minmax(130px, 1fr) minmax(160px, 2fr); align-items: center; gap: 8px; border-bottom: 1px solid var(--divider); font-size: 9px; }
.timeline-row:last-child { border-bottom: 0; }
.timeline-row.error, .timeline-row.failed, .timeline-row.cancelled { background: var(--danger-soft); }.timeline-row.warning { background: var(--warning-soft); }
.timeline-time, .timeline-kind { color: var(--text-muted); font-family: ui-monospace, monospace; }
.timeline-main { min-width: 0; display: flex; flex-direction: column; }.timeline-main strong { overflow: hidden; color: var(--text-secondary); text-overflow: ellipsis; white-space: nowrap; }.timeline-main small { color: var(--text-muted); }
.timeline-detail { overflow: hidden; color: var(--text-muted); text-overflow: ellipsis; white-space: nowrap; }

.final-message { padding: 14px 16px; }
.final-header { display: flex; align-items: center; gap: 8px; }.final-header h3 { margin: 0; color: var(--success); font-size: 13px; }
.final-icon { width: 24px; height: 24px; display: grid; place-items: center; border-radius: 50%; color: var(--success); background: #fff; }
.final-message p { margin: 7px 0 0 32px; color: var(--text-secondary); font-size: 11px; }

.preview-modal-overlay { position: fixed; inset: 0; z-index: var(--z-modal); padding: 18px; display: grid; place-items: center; }
.preview-modal { max-height: calc(100dvh - 36px); overflow: hidden; display: flex; flex-direction: column; }
.preview-modal-header { display: flex; align-items: center; justify-content: space-between; gap: 12px; }.preview-modal-header h3 { min-width: 0; margin: 0; overflow: hidden; font-size: 13px; text-overflow: ellipsis; white-space: nowrap; }
.preview-modal-actions { display: flex; align-items: center; gap: 6px; }
.modal-action-btn { padding: 0 11px; display: inline-flex; align-items: center; gap: 7px; border: 1px solid var(--border-strong); color: var(--text-secondary); background: var(--surface); font-weight: 700; cursor: pointer; }
.modal-close-btn { width: 40px; display: grid; place-items: center; border: 0; color: var(--text-secondary); background: transparent; cursor: pointer; }
.preview-modal-body { min-height: 0; flex: 1; display: grid; place-items: center; overflow: auto; }
.preview-modal-img { width: 100%; height: auto; }
.preview-modal-loading { min-height: 320px; display: grid; place-items: center; }.preview-spinner { width: 28px; height: 28px; border: 3px solid var(--border-strong); border-top-color: var(--info); border-radius: 50%; animation: spin 0.9s linear infinite; }

.chat-bar-collapsed { width: 100%; border: 0; display: flex; align-items: center; gap: 8px; font: inherit; font-size: 12px; font-weight: 700; cursor: pointer; }
.chat-bar-expanded { display: flex; flex-direction: column; }
.chat-bar-header { display: flex; align-items: center; justify-content: space-between; }
.chat-bar-title { color: var(--text); font-size: 12px; font-weight: 730; }
.chat-bar-close { border: 0; border-radius: 5px; color: var(--text-secondary); background: transparent; cursor: pointer; }
.chat-bar-close:hover { color: var(--danger); background: var(--danger-soft); }
.chat-messages { display: flex; flex-direction: column; gap: 8px; overflow: auto; }
.chat-empty { padding: 22px 10px; color: var(--text-muted); font-size: 11px; text-align: center; }
.chat-msg { display: flex; }.chat-msg.user { justify-content: flex-end; }.chat-msg.assistant { justify-content: flex-start; }
.chat-msg-bubble { max-width: 84%; padding: 8px 10px; border-radius: 6px; color: var(--text); background: var(--surface); border: 1px solid var(--border); font-size: 11px; white-space: pre-wrap; word-break: break-word; }
.chat-msg.user .chat-msg-bubble { color: #fff; border-color: var(--action-ink); background: var(--action-ink); }
.typing { display: inline-flex; align-items: center; gap: 4px; }.typing-dot { width: 5px; height: 5px; border-radius: 50%; background: var(--text-muted); animation: typing 1.1s ease-in-out infinite; }.typing-dot:nth-child(2) { animation-delay: 0.12s; }.typing-dot:nth-child(3) { animation-delay: 0.24s; }
.chat-input-row { display: flex; align-items: center; gap: 7px; }
.chat-input { min-width: 0; flex: 1; padding: 0 10px; border: 1px solid var(--border-strong); color: var(--text); }
.chat-input:focus { border-color: var(--info); outline: 2px solid var(--info-soft); }
.chat-send-btn { display: grid; place-items: center; flex: 0 0 auto; border: 1px solid var(--action-ink); color: #fff; background: var(--action-ink); cursor: pointer; }

.pulse-dot { width: 7px; height: 7px; display: inline-block; border-radius: 50%; background: currentColor; animation: pulse 1.4s ease-in-out infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
@keyframes pulse { 50% { opacity: 0.35; } }
@keyframes typing { 50% { transform: translateY(-3px); opacity: 0.45; } }

.topbar-tool {
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
.topbar-tool:hover { color: var(--text); background: var(--surface-muted); }
.topbar-tool.danger { color: var(--danger); border-color: #e7b7b7; background: var(--danger-soft); }
.task-list-trigger { display: none; }
.top-task-state { display: inline-flex; align-items: center; gap: 6px; color: var(--text-secondary); font-size: 11px; font-weight: 700; }
.top-task-state i { width: 7px; height: 7px; border-radius: 50%; background: var(--text-disabled); }
.top-task-state.running { color: var(--info); }.top-task-state.running i { background: var(--info); animation: pulse 1.4s ease-in-out infinite; }
.top-task-state.completed { color: var(--success); }.top-task-state.completed i { background: var(--success); }
.top-task-state.failed { color: var(--danger); }.top-task-state.failed i { background: var(--danger); }

.layout {
  width: 100%;
  height: calc(100dvh - var(--topbar-height));
  min-height: 0;
  display: grid;
  grid-template-columns: 272px minmax(0, 1fr);
  overflow: hidden;
  background: var(--canvas);
}

.sidebar-shell {
  position: static;
  min-width: 0;
  width: 272px;
  height: 100%;
  overflow: hidden;
}
.sidebar-shell :deep(.task-sidebar) { height: 100%; }
.sidebar-scrim,
.sidebar-close-btn,
.mobile-nav-btn { display: none; }

.main {
  min-width: 0;
  width: 100%;
  height: 100%;
  padding: 22px 24px 96px;
  display: flex;
  flex-direction: column;
  overflow: auto;
  background: var(--canvas);
}
.main > .dash-header { order: 1; }
.main > :deep(.progress-panel) { order: 2; }
.main > .split-layout { order: 3; }
.main > .final-message { order: 4; }
.main > .dev-status-panel { order: 5; }
.main > .welcome { order: 1; }

.dash-header {
  margin: 0 0 14px;
  padding: 0;
  align-items: flex-end;
  border: 0;
}
.dash-title { max-width: 900px; color: var(--text); font-size: 18px; font-weight: 730; line-height: 1.35; letter-spacing: 0; white-space: normal; }
.dash-meta { margin-top: 7px; gap: 12px; }
.dash-id { padding: 3px 6px; border: 1px solid var(--border); border-radius: 4px; color: var(--text-muted); background: var(--surface-muted); font-family: ui-monospace, monospace; }
.dash-duration, .dash-tokens { display: inline-flex; align-items: center; gap: 5px; color: var(--text-muted); }
.dash-duration svg, .dash-tokens svg { width: auto; height: auto; }
.dash-status { display: none; }

.main > :deep(.progress-panel) { margin-bottom: 18px; }

.split-layout {
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: 18px;
  overflow: visible;
}
.split-right { min-width: 0; padding: 0; display: flex; flex-direction: column; gap: 16px; overflow: visible; }
.slides-section { order: 1; margin: 0; }
.batch-list { order: 2; display: grid; gap: 8px; }
.split-left { order: 3; }

.slides-toolbar { min-height: 44px; margin: 0 0 10px; padding: 0; align-items: center; border: 0; }
.section-title { color: var(--text); font-size: 13px; font-weight: 730; }
.file-count { margin-left: 6px; color: var(--text-muted); font-size: 10px; font-weight: 600; }
.toolbar-actions { gap: 5px; }
.tool-btn { min-height: 36px; padding: 0 10px; display: inline-flex; align-items: center; justify-content: center; gap: 6px; border-radius: 5px; color: var(--text-secondary); background: var(--surface); }
.tool-btn.primary { color: #fff; border-color: var(--action-ink); background: var(--action-ink); }
.slides-list { display: grid; grid-template-columns: repeat(auto-fill, minmax(230px, 1fr)); gap: 12px; }

.batch-panel { margin: 0; border: 1px solid var(--border); border-radius: 6px; background: var(--surface-muted); box-shadow: none; }
.batch-panel.done { opacity: 0.72; }
.batch-header { padding: 9px 11px; }
.batch-cards { padding: 0 10px 10px; gap: 6px; }
.batch-card { border-radius: 5px; background: var(--surface); }

.activity-panel,
.dev-status-panel {
  margin: 0;
  overflow: hidden;
  border: 1px solid var(--border);
  border-radius: 6px;
  background: var(--surface);
  box-shadow: none;
}
.activity-panel > summary,
.dev-status-summary {
  min-height: 46px;
  padding: 0 14px;
  display: flex;
  align-items: center;
  gap: 10px;
  color: var(--text-secondary);
  background: var(--surface);
  cursor: pointer;
  list-style: none;
}
.activity-panel > summary::-webkit-details-marker,
.dev-status-summary::-webkit-details-marker { display: none; }
.activity-panel > summary > span { display: inline-flex; align-items: center; gap: 7px; color: var(--text); font-size: 12px; font-weight: 700; }
.activity-panel > summary small { margin-left: auto; color: var(--text-muted); font-size: 10px; }
.activity-chevron,
.diagnostic-chevron { flex: 0 0 auto; transition: transform var(--motion-fast); }
.activity-panel[open] .activity-chevron,
.dev-status-panel[open] .diagnostic-chevron { transform: rotate(180deg); }
.activity-panel :deep(.event-log-section) { height: auto; padding: 0 14px 14px; }
.activity-panel :deep(.log-header) { display: none; }
.activity-panel :deep(.log-box) { border: 0; border-radius: 5px; background: var(--surface-muted); box-shadow: none; }

.dev-status-panel { margin-top: 18px; }
.dev-status-title { font-size: 12px; }
.dev-status-phase { padding: 2px 6px; border-radius: 4px; color: var(--info); background: var(--info-soft); font-size: 10px; }
.diagnostic-alert { margin-left: auto; color: var(--warning); background: var(--warning-soft); }
.dev-status-body { padding: 0 14px 14px; }
.dev-status-grid { grid-template-columns: repeat(auto-fit, minmax(140px, 1fr)); gap: 8px; }
.dev-stat { min-height: 76px; padding: 10px; border: 1px solid var(--divider); border-radius: 5px; background: var(--surface-muted); }
.runtime-timeline { border: 1px solid var(--divider); border-radius: 5px; }

.final-message { margin: 18px 0 0; border: 1px solid #b9dfcf; border-radius: 6px; background: var(--success-soft); box-shadow: none; }

.welcome {
  min-height: 100%;
  width: min(820px, 100%);
  margin: 0 auto;
  padding: 7vh 0 40px;
  display: flex;
  flex-direction: column;
  justify-content: flex-start;
  align-items: stretch;
  text-align: left;
}
.welcome-hero { display: grid; grid-template-columns: 46px minmax(0, 1fr); column-gap: 14px; align-items: center; }
.welcome-icon { grid-row: 1 / 4; width: 46px; height: 46px; display: grid; place-items: center; border-radius: 7px; color: var(--action-ink); background: var(--action-soft); }
.welcome-icon svg { width: 28px; height: 28px; }
.welcome-kicker { color: var(--action-ink); font-size: 10px; font-weight: 800; }
.welcome h2 { margin: 1px 0 0; color: var(--text); font-size: 22px; letter-spacing: 0; }
.welcome p { margin: 4px 0 0; color: var(--text-muted); font-size: 13px; }
.welcome-examples { width: 100%; margin-top: 34px; }
.welcome-examples h3 { margin-bottom: 9px; color: var(--text-secondary); font-size: 11px; }
.examples-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 8px; }
.example-chip { min-height: 48px; padding: 0 13px; border: 1px solid var(--border); border-radius: 6px; color: var(--text-secondary); background: var(--surface); text-align: left; box-shadow: none; cursor: pointer; }
.example-chip:hover { color: var(--text); border-color: var(--border-strong); background: var(--surface-muted); transform: none; }
.compose-entry { width: 100%; margin-top: 28px; padding: 16px 0; display: flex; align-items: center; justify-content: space-between; gap: 18px; border-top: 1px solid var(--border); border-radius: 0; background: transparent; box-shadow: none; }
.compose-entry h3 { margin: 0; color: var(--text); font-size: 13px; }
.compose-entry p { margin-top: 3px; font-size: 11px; }
.compose-entry .compose-btn { min-height: 40px; padding: 0 12px; display: inline-flex; align-items: center; gap: 7px; flex: 0 0 auto; border: 1px solid var(--border-strong); border-radius: 5px; color: var(--text); background: var(--surface); cursor: pointer; }

.chat-bar {
  position: fixed;
  z-index: calc(var(--z-nav) - 1);
  right: 20px;
  bottom: 18px;
  left: auto;
  width: min(420px, calc(100vw - var(--rail-width) - 40px));
  overflow: hidden;
  border: 1px solid var(--border-strong);
  border-radius: 7px;
  background: var(--surface);
  box-shadow: var(--shadow-md);
}
.chat-bar-collapsed { min-height: 46px; padding: 0 14px; color: var(--text); background: var(--surface); }
.chat-bar-collapsed:hover { background: var(--surface-muted); }
.chat-bar-expanded { max-height: min(560px, 72dvh); }
.chat-bar-header { min-height: 46px; padding: 0 8px 0 14px; border-bottom: 1px solid var(--border); }
.chat-bar-close { width: 38px; height: 38px; display: grid; place-items: center; }
.chat-messages { min-height: 150px; padding: 12px; background: var(--surface-muted); }
.chat-input-row { padding: 10px; border-top: 1px solid var(--border); }
.chat-input { min-height: 42px; border-radius: 5px; background: var(--surface); }
.chat-send-btn { width: 42px; height: 42px; border-radius: 5px; }

.preview-modal-overlay { background: rgba(15, 17, 18, 0.62); }
.preview-modal { width: min(1100px, 96vw); border-radius: 8px; background: var(--surface); box-shadow: var(--shadow-lg); }
.preview-modal-header { min-height: 58px; padding: 8px 10px 8px 16px; border-bottom: 1px solid var(--border); }
.preview-modal-body { padding: 14px; background: #dde2e3; }
.preview-modal-img { max-height: calc(100dvh - 110px); object-fit: contain; }
.modal-action-btn, .modal-close-btn { min-height: 40px; border-radius: 5px; }

@media (max-width: 1180px) {
  .layout { display: block; height: calc(100dvh - var(--topbar-height)); }
  .task-list-trigger { display: inline-flex; }
  .sidebar-shell {
    position: fixed;
    inset: 0 auto 0 0;
    z-index: calc(var(--z-modal) + 1);
    width: min(88vw, 320px);
    height: 100dvh;
    transform: translateX(-102%);
    transition: transform var(--motion-medium);
    box-shadow: var(--shadow-lg);
  }
  .sidebar-shell.open { transform: translateX(0); }
  .sidebar-scrim { position: fixed; inset: 0; z-index: var(--z-modal); display: block; border: 0; background: rgba(15, 17, 18, 0.55); }
  .sidebar-close-btn { position: fixed; top: 6px; left: min(calc(88vw - 50px), 270px); z-index: calc(var(--z-modal) + 2); width: 42px; height: 42px; display: grid; place-items: center; border: 0; border-radius: 5px; color: var(--text-secondary); background: var(--surface); cursor: pointer; }
  .main { height: 100%; padding: 20px 22px 92px; }
  .chat-bar { width: min(420px, calc(100vw - 40px)); }
}

@media (max-width: 720px) {
  .top-task-state { display: none; }
  .topbar-tool { width: 40px; min-height: 40px; padding: 0; justify-content: center; }
  .topbar-tool span { display: none; }
  .layout { display: flex; flex-direction: column; }
  .main { height: auto; min-height: 0; flex: 1; padding: 16px 14px; }
  .dash-header { margin-bottom: 12px; }
  .dash-title { font-size: 16px; }
  .token-detail { display: none; }
  .slides-list { grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 9px; }
  .slides-toolbar { align-items: flex-start; }
  .toolbar-actions { width: auto; display: flex; flex-wrap: wrap; }
  .tool-btn.primary { grid-column: auto; }
  .examples-grid { grid-template-columns: 1fr; }
  .chat-bar { position: relative; right: auto; bottom: auto; left: auto; width: auto; margin: 0 8px 8px; flex: 0 0 auto; }
  .preview-modal-overlay { padding: 0; }
  .preview-modal { width: 100vw; height: 100dvh; border-radius: 0; }
  .preview-modal-body { height: calc(100dvh - 58px); }
  .preview-modal-img { width: 100%; height: 100%; max-height: none; }
}

@media (max-width: 520px) {
  .slides-list { grid-template-columns: minmax(0, 1fr); }
  .slides-toolbar { gap: 8px; }
  .toolbar-actions { width: 100%; display: grid; grid-template-columns: 1fr 1fr; }
  .tool-btn.primary { grid-column: 1 / -1; }
  .welcome { padding-top: 5vh; }
  .welcome-hero { grid-template-columns: 40px minmax(0, 1fr); }
  .welcome-icon { width: 40px; height: 40px; }
  .welcome h2 { font-size: 18px; }
  .compose-entry { align-items: flex-start; flex-direction: column; }
}
</style>
