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
  Presentation,
  Square,
  X,
} from 'lucide-vue-next';
import type { TaskInfo, TaskItem, SSEEvent, RuntimeMeta, RuntimeEvent } from '../types';
import { STATUS_LABELS } from '../types';
import {
  fetchTasks, createTask, fetchTask, cancelTask, deleteTask,
  isLoggedIn, continueTask, fetchConversation, fetchRuntimeEvent, routeMessage,
} from '../api';
import { authState } from '../stores/auth';
import AppShell from '../components/AppShell.vue';
import Sidebar from '../components/Sidebar.vue';
import ProgressBar from '../components/ProgressBar.vue';
import EventLog from '../components/EventLog.vue';
import SlidePreviewCard from '../components/SlidePreviewCard.vue';
import ConversationComposer from '../components/ConversationComposer.vue';
import RuntimeEventDetail from '../components/RuntimeEventDetail.vue';
import {
  appendAssistantStreamContent, compactRuntimeEvents, deriveLiveActivity, deriveObservableSteps, mergeConversationMessages, mergeRuntimeEvents, mergeRuntimeMeta, mergeSlideDeliveries,
  nextReplayCursor, recoverConversationMessages, runtimeEventDetailLabel, runtimeEventKindLabel,
  runtimeEventNameLabel, runtimeEventStatusLabel, summarizeTaskTitle,
} from '../utils/workbench';

const router = useRouter();
const auth = authState;

// ── State (same as old working App.vue) ─────────────────────────────────
const tasks = ref<TaskInfo[]>([]);
const selectedId = ref<string | null>(null);
const taskItems = ref<TaskItem[]>([]);
const doneCount = ref(0);
const totalCount = ref(0);
const currentPhase = ref('');    // preparing/planning/generating/complete
const phaseDetail = ref('');     // "生成第3/21页" / "读取模板" 等
const logLines = ref<{ ts: number; text: string; kind: import('../types').LogKind }[]>([]);
const finalFiles = ref<string[]>([]);
const finalMessage = ref('');
const duration = ref('');
const activeWorkers = ref(0);
const runtimeMeta = ref<RuntimeMeta | null>(null);
const selectedRuntimeEvent = ref<RuntimeEvent | null>(null);
const selectedRuntimeEventLoading = ref(false);
const selectedRuntimeEventError = ref('');
const thumbnailVersions = ref<Record<string, number>>({});
const thumbnailFailures = ref<Record<string, string>>({});
const cancelling = ref(false);
const creating = ref(false);
const loadError = ref('');
const sidebarOpen = ref(false);
const sseConnectionInterrupted = ref(false);

function closeSidebar() {
  sidebarOpen.value = false;
}

function handleSidebarSelect(id: string) {
  closeSidebar();
  selectTask(id);
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

// ── Unified conversation composer ─────────────────────────────────────────────
const composerInput = ref('');
const composerLoading = ref(false);
const composerError = ref('');
const composerNotice = ref('');
const conversationLoading = ref(false);
const conversationMessages = ref<import('../types').ConversationMessage[]>([]);
const streamingAssistant = ref('');
const streamingAssistantStartedAt = ref('');
const streamingAssistantSegments = ref<import('../types').ConversationMessage[]>([]);
const continuationQueued = ref(false);
const manualAgentMode = ref<'chat' | 'pptagent'>('chat');

const composerMode = computed<'create' | 'queue' | 'continue'>(() => {
  if (!selectedTask.value) return 'create';
  return selectedTask.value.status === 'running' ? 'queue' : 'continue';
});
async function loadConversation(taskId: string, replace = false) {
  conversationLoading.value = true;
  try {
    const session = await fetchConversation(taskId);
    if (selectedId.value !== taskId) return;
    const recovered = recoverConversationMessages(session);
    conversationMessages.value = replace
      ? mergeConversationMessages([], recovered)
      : mergeConversationMessages(conversationMessages.value, recovered);
    if (logLines.value.length === 0) restoreFromConversation(session);
    if (session.files?.length) finalFiles.value = session.files;
    if (session.duration) duration.value = session.duration;
    if (session.done_count !== undefined) doneCount.value = session.done_count;
    if (session.total_count !== undefined) totalCount.value = session.total_count;
    if (session.runtime_meta) {
	  const merged = mergeRuntimeMeta(runtimeMeta.value, session.runtime_meta);
      runtimeMeta.value = merged;
      if (selectedRuntimeEvent.value && !merged.recent_events?.some(evt => evt.id === selectedRuntimeEvent.value?.id)) {
        selectedRuntimeEvent.value = null;
      }
    }
    return session;
  } catch {
    if (selectedId.value === taskId) composerError.value = '对话历史暂时无法恢复，任务产物仍可正常查看';
  } finally {
    if (selectedId.value === taskId) conversationLoading.value = false;
  }
}

function appendAssistantDelta(delta: string, timelineOrder = 0) {
  if (!delta) return;
  if (!streamingAssistant.value && !streamingAssistantStartedAt.value) {
    streamingAssistantStartedAt.value = new Date().toISOString();
  }
  const previous = streamingAssistant.value;
  const next = appendAssistantStreamContent(previous, delta);
  const visibleDelta = visibleAssistantDelta(previous, next, delta);
  streamingAssistant.value = next;
  appendStreamingAssistantSegment(visibleDelta, timelineOrder);
}

function finalizeAssistantTurn() {
  const content = streamingAssistant.value.trim();
  const segments = streamingAssistantSegments.value.filter(message => message.content.trim());
  if (segments.length > 0) {
    conversationMessages.value = mergeConversationMessages(conversationMessages.value, segments);
  } else if (content) {
    conversationMessages.value = mergeConversationMessages(conversationMessages.value, [{
      role: 'assistant', content, timestamp: streamingAssistantStartedAt.value || new Date().toISOString(),
    }]);
  }
  streamingAssistant.value = '';
  streamingAssistantStartedAt.value = '';
  streamingAssistantSegments.value = [];
  composerLoading.value = false;
}

function visibleAssistantDelta(previous: string, next: string, rawDelta: string): string {
  if (!next) return '';
  if (!previous) return next;
  if (next === previous) return '';
  if (next.startsWith(previous)) return next.slice(previous.length);
  return rawDelta;
}

function appendStreamingAssistantSegment(content: string, timelineOrder: number) {
  if (!content.trim()) return;
  const now = new Date().toISOString();
  const segments = [...streamingAssistantSegments.value];
  const last = segments[segments.length - 1];
  if (last && !hasRuntimeBoundarySince(last.timestamp, now)) {
    segments[segments.length - 1] = {
      ...last,
      content: appendAssistantStreamContent(last.content, content),
      timestamp: now,
      timeline_order: timelineOrder || last.timeline_order,
    };
  } else {
    segments.push({
      role: 'assistant',
      content,
      timestamp: now,
      timeline_order: timelineOrder || undefined,
    });
  }
  streamingAssistantSegments.value = segments;
}

function hasRuntimeBoundarySince(start: string, end: string): boolean {
  const from = Date.parse(start || '');
  const to = Date.parse(end || '');
  if (Number.isNaN(from) || Number.isNaN(to)) return false;
  return runtimeTimelineAll.value.some(event => {
    const timestamp = Date.parse(event.timestamp || '');
    if (Number.isNaN(timestamp) || timestamp <= from || timestamp > to) return false;
    const kind = (event.kind || '').toLowerCase();
    return kind.startsWith('tool_') || kind.startsWith('slide_render_') || kind === 'task_terminal';
  });
}

async function submitComposer() {
  const message = composerInput.value.trim();
  if (!message || composerLoading.value) return;
  composerError.value = '';
  composerNotice.value = '';

  if (!selectedTask.value) {
    composerLoading.value = true;
    try {
      composerInput.value = '';
      const routed = await routeMessage(message);
      if (routed.intent === 'chat' || routed.action === 'reply') {
        conversationMessages.value = mergeConversationMessages(conversationMessages.value, [
          { role: 'user', content: message, timestamp: new Date().toISOString() },
          { role: 'assistant', content: routed.reply || '这是普通对话，不会创建 PPT 任务。', timestamp: new Date().toISOString() },
        ]);
        return;
      }
      if (routed.intent === 'plan' || routed.action === 'save_plan') {
        conversationMessages.value = mergeConversationMessages(conversationMessages.value, [
          { role: 'user', content: message, timestamp: new Date().toISOString() },
          { role: 'assistant', content: withDraftNote(routed.reply || '已进入 PPT Agent 规划状态，不会生成 PPT 文件。', routed.draft_id), timestamp: new Date().toISOString() },
        ]);
        composerNotice.value = routed.draft_id ? `已保存规划草稿 ${routed.draft_id}，未创建任务。` : '已识别为规划请求，未创建任务。可打开高级编排继续完善 DeckSpec。';
        return;
      }
      if (routed.intent === 'fix') {
        composerError.value = withTaskCandidates(routed.reply || '这是修复请求，请先选择要修改的任务。', routed.task_candidates || []);
        return;
      }
      if (routed.needs_confirmation || routed.action === 'ask_clarification') {
        conversationMessages.value = mergeConversationMessages(conversationMessages.value, [
          { role: 'user', content: message, timestamp: new Date().toISOString() },
          { role: 'assistant', content: routed.reply || '已识别为 PPT 意图，但还需要补充信息。', timestamp: new Date().toISOString() },
        ]);
        return;
      }
      const info = await createTask(routed.normalized_request || message);
      tasks.value = [info, ...tasks.value.filter(task => task.id !== info.id)];
      selectTask(info.id);
    } catch (error) {
      composerError.value = error instanceof Error ? error.message : '创建任务失败，请重试';
    } finally {
      composerLoading.value = false;
    }
    return;
  }

  const taskId = selectedTask.value.id;
  composerLoading.value = true;
  const now = new Date().toISOString();
  conversationMessages.value = [...conversationMessages.value, {
    role: 'user', content: message, timestamp: new Date().toISOString(),
  }];
  composerInput.value = '';
  streamingAssistant.value = '';
  streamingAssistantStartedAt.value = '';
  streamingAssistantSegments.value = [];
  try {
    const routed = await routeMessage(message, taskId, manualAgentMode.value);
    manualAgentMode.value = routed.mode;
    if (routed.intent === 'chat' || routed.action === 'reply') {
      conversationMessages.value = mergeConversationMessages(conversationMessages.value, [{
        role: 'assistant', content: routed.reply || '这是普通对话，不会进入修复流程。', timestamp: now,
      }]);
      return;
    }
    if (routed.intent === 'plan' || routed.action === 'save_plan') {
      conversationMessages.value = mergeConversationMessages(conversationMessages.value, [{
        role: 'assistant', content: withDraftNote(routed.reply || '已进入 PPT Agent 规划状态，不会生成 PPT 文件。', routed.draft_id), timestamp: now,
      }]);
      composerNotice.value = routed.draft_id ? `已保存规划草稿 ${routed.draft_id}，未修改当前任务。` : '已识别为规划请求，未修改当前任务。';
      return;
    }
    if (routed.intent === 'create') {
      conversationMessages.value = mergeConversationMessages(conversationMessages.value, [{
        role: 'assistant', content: '这像是新建 PPT 请求。为避免覆盖当前任务，请回到新建入口创建新的演示。', timestamp: now,
      }]);
      return;
    }
    if (routed.needs_confirmation || routed.action === 'ask_clarification') {
      conversationMessages.value = mergeConversationMessages(conversationMessages.value, [{
        role: 'assistant', content: routed.reply || '请说明要修复的任务、页码或具体问题。', timestamp: now,
      }]);
      return;
    }
    const accepted = await continueTask(taskId, message);
    composerNotice.value = accepted.message;
    continuationQueued.value = accepted.status === 'queued';
    sseCompleted = false;
    connectSSE(taskId, accepted.after_event_id || 0);
    startPolling(taskId);
    if (accepted.status === 'queued') composerLoading.value = false;
  } catch (error) {
    composerError.value = error instanceof Error ? error.message : '发送失败，请重试';
    composerLoading.value = false;
  }
}

function withDraftNote(content: string, draftId?: string): string {
  if (!draftId) return content;
  return `${content}\n\n草稿已保存：${draftId}`;
}

function withTaskCandidates(content: string, candidates: import('../api').TaskCandidate[]): string {
  if (!candidates.length) return content;
  return `${content}\n\n最近可选任务：\n${candidates.map(item => `- ${item.title || item.id} (${item.id})`).join('\n')}`;
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
type RuntimeCategory = 'all' | 'llm' | 'tool' | 'compression' | 'phase' | 'delivery' | 'error' | 'other';

const selectedTask = computed(() => tasks.value.find(t => t.id === selectedId.value));
const selectedTaskTitle = computed(() => summarizeTaskTitle(selectedTask.value?.query || '任务工作台'));

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
const selectedRuntimeCategory = ref<RuntimeCategory>('all');
const runtimeTimelineAll = computed(() => compactRuntimeEvents(
  (runtimeMeta.value?.recent_events || []).slice().reverse(),
));
const observableSteps = computed(() => deriveObservableSteps(runtimeTimelineAll.value, 6));
const runtimeTimeline = computed(() => (
  selectedRuntimeCategory.value === 'all'
    ? runtimeTimelineAll.value
    : runtimeTimelineAll.value.filter(evt => runtimeEventCategory(evt) === selectedRuntimeCategory.value)
));
const runtimeCategoryOptions = computed(() => {
  const counts: Record<RuntimeCategory, number> = {
    all: runtimeTimelineAll.value.length,
    llm: 0,
    tool: 0,
    compression: 0,
    phase: 0,
    delivery: 0,
    error: 0,
    other: 0,
  };
  for (const evt of runtimeTimelineAll.value) counts[runtimeEventCategory(evt)] += 1;
  return ([
    ['all', '全部'],
    ['llm', '模型'],
    ['tool', '工具'],
    ['compression', '压缩'],
    ['phase', '阶段'],
    ['delivery', '交付'],
    ['error', '错误'],
    ['other', '其他'],
  ] as Array<[RuntimeCategory, string]>).map(([key, label]) => ({ key, label, count: counts[key] }));
});
const liveActivity = computed(() => deriveLiveActivity({
  status: selectedTask.value?.status,
  phase: currentPhase.value,
  phaseDetail: phaseDetail.value,
  lastTool: runtimeMeta.value?.last_tool,
  connectionInterrupted: sseConnectionInterrupted.value,
  done: doneCount.value,
  total: totalCount.value,
  error: selectedTask.value?.error,
}));

function runtimeEventCategory(evt: RuntimeEvent): RuntimeCategory {
  const kind = (evt.kind || '').toLowerCase();
  const name = (evt.name || '').toLowerCase();
  const status = (evt.status || '').toLowerCase();
  if (status === 'error' || status === 'failed' || kind.includes('error')) return 'error';
  if (kind.startsWith('llm') || kind === 'model_request' || name.includes('chatmodel') || name.includes('chat_model')) return 'llm';
  if (kind.startsWith('tool') || name === 'toolnode') return 'tool';
  if (kind === 'compression' || kind === 'planner_context_compressed') return 'compression';
  if (kind.includes('manifest') || kind.includes('file') || kind.includes('terminal') || kind.includes('delivery') || kind.includes('plan') || kind.includes('intent')) return 'delivery';
  if (kind.includes('phase') || kind.includes('progress')) return 'phase';
  return 'other';
}

function runtimeEventCategoryLabel(evt: RuntimeEvent): string {
  const labels: Record<RuntimeCategory, string> = {
    all: '全部',
    llm: '模型',
    tool: '工具',
    compression: '压缩',
    phase: '阶段',
    delivery: '交付',
    error: '错误',
    other: '其他',
  };
  return labels[runtimeEventCategory(evt)];
}

function setRuntimeCategory(category: RuntimeCategory) {
  selectedRuntimeCategory.value = category;
  if (
    selectedRuntimeEvent.value
    && category !== 'all'
    && runtimeEventCategory(selectedRuntimeEvent.value) !== category
  ) {
    selectedRuntimeEvent.value = null;
    selectedRuntimeEventError.value = '';
    selectedRuntimeEventLoading.value = false;
  }
}

async function selectRuntimeEvent(evt: RuntimeEvent) {
  if (selectedRuntimeEvent.value?.id === evt.id) {
    selectedRuntimeEvent.value = null;
    selectedRuntimeEventError.value = '';
    return;
  }
  selectedRuntimeEvent.value = evt;
  selectedRuntimeEventError.value = '';
  if (!selectedId.value || evt.metadata_loaded) return;
  const taskId = selectedId.value;
  const eventId = evt.id;
  selectedRuntimeEventLoading.value = true;
  try {
    const detail = await fetchRuntimeEvent(taskId, eventId);
    if (selectedRuntimeEvent.value?.id === eventId && selectedId.value === taskId) {
      selectedRuntimeEvent.value = { ...evt, ...detail, metadata_loaded: true };
    }
  } catch (error) {
    if (selectedRuntimeEvent.value?.id === eventId && selectedId.value === taskId) {
      selectedRuntimeEventError.value = error instanceof Error ? error.message : '事件详情加载失败';
    }
  } finally {
    if (selectedRuntimeEvent.value?.id === eventId && selectedId.value === taskId) {
      selectedRuntimeEventLoading.value = false;
    }
  }
}

function selectRuntimeEventById(id: number) {
  const event = runtimeTimelineAll.value.find(evt => evt.id === id);
  if (event) selectRuntimeEvent(event);
}

async function loadInlineRuntimeEvent(eventId: number) {
  const taskId = selectedId.value;
  if (!taskId || eventId <= 0) return;
  const existing = runtimeMeta.value?.recent_events?.find(event => event.id === eventId);
  if (existing?.metadata_loaded) return;
  try {
    const detail = await fetchRuntimeEvent(taskId, eventId);
    if (selectedId.value !== taskId) return;
    const current = runtimeMeta.value || { elapsed_ms: 0 };
    runtimeMeta.value = {
      ...current,
      recent_events: mergeRuntimeEvents(current.recent_events || [], [{ ...detail, metadata_loaded: true }]),
    };
  } catch (error) {
    if (selectedId.value === taskId) {
      composerError.value = error instanceof Error ? `工具结果加载失败：${error.message}` : '工具结果加载失败';
    }
  }
}

const sampleQueries = [
  '做一个关于新能源汽车的行业分析报告',
  '制作一个产品发布会演示文稿',
  '做一个 AI 大模型技术分享的 PPT',
  '制作一个公司季度总结汇报',
];

const orderedSlides = computed(() => mergeSlideDeliveries(taskItems.value, finalFiles.value));

function findItem(id: string): TaskItem | undefined {
  return taskItems.value.find(t => t.task_id === id);
}

// ── Multi-select download ──────────────────────────────────────────────
const selectedSlides = ref<Set<string>>(new Set());
const readySlides = computed(() => orderedSlides.value.filter(s => s.fileReady));

function toggleSelect(slideKey: string) {
  const next = new Set(selectedSlides.value);
  if (next.has(slideKey)) next.delete(slideKey);
  else next.add(slideKey);
  selectedSlides.value = next;
}

function selectAll() {
  selectedSlides.value = new Set(readySlides.value.map(s => s.key));
}

function deselectAll() {
  selectedSlides.value = new Set();
}

function downloadSelected() {
  const ids = selectedSlides.value;
  if (ids.size === 0) return;
  const slides = readySlides.value.filter(s => ids.has(s.key));
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

let lastSeenEventID = 0;

function connectSSE(taskId: string, afterEventID = 0) {
  if (!taskId) return;
  if (sseCompleted) return; // already received complete, don't reconnect
  if (es) es.close();
  lastSeenEventID = afterEventID;
  const streamURL = afterEventID > 0
    ? `/api/tasks/${taskId}/stream?after_id=${encodeURIComponent(afterEventID)}`
    : `/api/tasks/${taskId}/stream`;
  es = new EventSource(streamURL);
  activeWorkers.value = 0;

  es.onopen = () => {
	if (sseConnectionInterrupted.value) {
	  addLog('worker', '实时连接已恢复');
	}
	sseConnectionInterrupted.value = false;
  };

  es.onerror = () => {
	if (sseCompleted || sseConnectionInterrupted.value) return;
	sseConnectionInterrupted.value = true;
	addLog('error', '实时连接暂时中断，浏览器正在自动重连；任务状态仍会通过轮询同步');
  };

  const handler = (e: MessageEvent) => {
    let evt: SSEEvent;
    try { evt = JSON.parse(e.data); } catch { return; }
    const eventID = e.lastEventId ? Number.parseInt(e.lastEventId, 10) : (evt.id || 0);
    if (eventID > 0 && eventID <= lastSeenEventID) return;
    if (eventID > 0) lastSeenEventID = eventID;

    switch (evt.type) {
      case 'answer': {
        const chunk = evt.content || '';
        if (chunk) {
          appendAssistantDelta(chunk);
        }
		break;
	  }

      case 'answer_end':
        finalizeAssistantTurn();
        break;

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
          runtimeMeta.value = mergeRuntimeMeta(runtimeMeta.value, evt.runtime_meta);
        }
        break;

      case 'error':
        addLog('error', evt.error || evt.content || '');
        break;

      case 'continue_queued':
        composerNotice.value = '已进入反馈处理阶段';
        break;

      case 'continue_complete':
        finalizeAssistantTurn();
        continuationQueued.value = false;
        sseCompleted = true;
        activeWorkers.value = 0;
        es?.close();
        es = null;
        stopPolling();
        void refreshTask(taskId);
        window.setTimeout(() => void loadConversation(taskId, true), 200);
        break;

      case 'complete':
        sseCompleted = !continuationQueued.value;
        finalizeAssistantTurn();
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
        if (evt.runtime_meta) runtimeMeta.value = mergeRuntimeMeta(runtimeMeta.value, evt.runtime_meta);
        if (evt.total_tokens) {
          const t = tasks.value.find(x => x.id === taskId);
          if (t) {
            t.prompt_tokens = evt.prompt_tokens || 0;
            t.completion_tokens = evt.completion_tokens || 0;
            t.total_tokens = evt.total_tokens || 0;
          }
        }
        activeWorkers.value = 0;
        es?.close();
        es = null;
        void refreshTask(taskId);
        if (continuationQueued.value) {
          window.setTimeout(() => {
            if (selectedId.value !== taskId || !continuationQueued.value) return;
            sseCompleted = false;
            connectSSE(taskId, lastSeenEventID);
            startPolling(taskId);
          }, 300);
        } else {
          stopPolling();
          window.setTimeout(() => void loadConversation(taskId, true), 200);
        }
        break;
    }
  };

  es.addEventListener('answer', handler);
  es.addEventListener('answer_end', handler);
  es.addEventListener('system_step', handler);
  es.addEventListener('tool_call', handler);
  es.addEventListener('progress', handler);
  es.addEventListener('file_ready', handler);
  es.addEventListener('thumbnail_ready', handler);
  es.addEventListener('thumbnail_error', handler);
  es.addEventListener('token_usage', handler);
  es.addEventListener('runtime_meta', handler);
  es.addEventListener('error', handler);
  es.addEventListener('continue_queued', handler);
  es.addEventListener('continue_complete', handler);
  es.addEventListener('complete', handler);
}

function disconnectSSE() {
  if (es) { es.close(); es = null; }
  stopPolling();
  streamingAssistant.value = '';
  streamingAssistantStartedAt.value = '';
  streamingAssistantSegments.value = [];
  composerLoading.value = false;
  continuationQueued.value = false;
  currentPhase.value = '';
  phaseDetail.value = '';
  runtimeMeta.value = null;
  sseConnectionInterrupted.value = false;
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

// Restore operational log lines from conversation_content on cold start.
// Assistant/user prose belongs exclusively to ConversationComposer.
function restoreFromConversation(sess: import('../types').ConversationSession) {
  if (!sess.conversation_content) return;

  const lines: { ts: number; text: string; kind: import('../types').LogKind }[] = [];
  const linesArr = sess.conversation_content.split('\n');

  for (const raw of linesArr) {
    const line = raw.trim();
    if (!line) continue;
    if (line.startsWith('**错误**')) {
      lines.push({ ts: Date.now(), text: line.replace('**错误**:', '').trim(), kind: 'error' });
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
  conversationMessages: import('../types').ConversationMessage[];
  streamingAssistant: string;
  streamingAssistantStartedAt: string;
  streamingAssistantSegments: import('../types').ConversationMessage[];
  lastSeenEventID: number;
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
    conversationMessages: mergeConversationMessages([], conversationMessages.value),
    streamingAssistant: streamingAssistant.value,
    streamingAssistantStartedAt: streamingAssistantStartedAt.value,
    streamingAssistantSegments: mergeConversationMessages([], streamingAssistantSegments.value),
    lastSeenEventID,
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
  conversationMessages.value = mergeConversationMessages([], c.conversationMessages);
  streamingAssistant.value = c.streamingAssistant;
  streamingAssistantStartedAt.value = c.streamingAssistantStartedAt || '';
  streamingAssistantSegments.value = mergeConversationMessages([], c.streamingAssistantSegments || []);
  lastSeenEventID = c.lastSeenEventID;
  return true;
}

// ── Actions ────────────────────────────────────────────────────────────
let selectionEpoch = 0;

async function selectTask(id: string) {
  if (!id) return;
  if (selectedId.value === id) return;
  const epoch = ++selectionEpoch;
  if (selectedId.value && selectedId.value !== id) {
    saveCache(selectedId.value);
    disconnectSSE();
  }
  selectedId.value = id;
  conversationMessages.value = [];
  conversationLoading.value = true;
  composerError.value = '';
  composerNotice.value = '';
  const t = tasks.value.find(x => x.id === id);
  if (!t) return;

  // Restore cached state if switching back to a previously viewed task
  if (restoreCache(id)) {
    conversationLoading.value = false;
    if (t.status === 'running') {
	  sseCompleted = false;
	  connectSSE(id, lastSeenEventID);
	  startPolling(id);
    } else {
      await loadConversation(id, true);
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
	selectedRuntimeEvent.value = null;
	selectedRuntimeEventLoading.value = false;
	selectedRuntimeEventError.value = '';
	selectedSlides.value = new Set();
	sseCompleted = false;
	lastSeenEventID = 0;
  streamingAssistant.value = '';
  streamingAssistantStartedAt.value = '';
  streamingAssistantSegments.value = [];

  if (t.status === 'running') {
    // Restore finalized turns first, then replay only the unfinished turn.
    const session = await loadConversation(id, true);
    if (epoch !== selectionEpoch || selectedId.value !== id) return;
    lastSeenEventID = nextReplayCursor(0, session?.replay_after_event_id || 0);
    connectSSE(id, lastSeenEventID);
    startPolling(id);
  } else {
    await loadConversation(id, true);
  }
}

async function handleCreateTask(query: string) {
  if (creating.value) return;
  creating.value = true;
  loadError.value = '';
  try {
    const routed = await routeMessage(query);
    if (routed.intent !== 'create' || routed.needs_confirmation || routed.action === 'ask_clarification') {
      loadError.value = routed.reply || '当前输入还不能直接创建 PPT。';
      return;
    }
    const info = await createTask(routed.normalized_request || query);
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
    selectedId.value = null;
  }
  conversationMessages.value = [];
  composerError.value = '';
  composerNotice.value = '';
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

onUnmounted(() => { disconnectSSE(); stopPolling(); });
</script>

<template>
  <AppShell
    :title="selectedTaskTitle"
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
        @logout="handleSidebarLogout"
        @select-task="handleSidebarSelect"
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
            <p>进入编排工作区，自行调整每一页的布局和内容约束。</p>
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
            <h2 class="dash-title">{{ selectedTaskTitle }}</h2>
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
            <details v-if="selectedTask?.query" class="task-query-details">
              <summary>查看原始需求</summary>
              <pre>{{ selectedTask.query }}</pre>
            </details>
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

        <section v-if="observableSteps.length > 0" class="execution-watch" aria-label="执行观察">
          <div class="watch-head">
            <span>
              <Activity :size="15" />
              执行观察
            </span>
            <small>最近 {{ observableSteps.length }} 步 · 可展开诊断查看完整事件</small>
          </div>
          <div class="watch-list">
            <article
              v-for="step in observableSteps"
              :key="step.id"
              role="button"
              tabindex="0"
              class="watch-step"
              :class="[step.status, step.category, { selected: selectedRuntimeEvent?.id === step.id }]"
              @click="selectRuntimeEventById(step.id)"
              @keydown.enter="selectRuntimeEventById(step.id)"
              @keydown.space.prevent="selectRuntimeEventById(step.id)"
            >
              <span class="watch-time">{{ fmtElapsed(step.elapsed_ms) }}</span>
              <span class="watch-dot" aria-hidden="true"></span>
              <span class="watch-copy">
                <strong>{{ step.label }}</strong>
                <small>{{ step.detail }}</small>
                <span v-if="step.urls.length" class="watch-sources">
                  <a
                    v-for="url in step.urls.slice(0, 3)"
                    :key="url"
                    :href="url"
                    target="_blank"
                    rel="noreferrer"
                    @click.stop
                  >{{ url }}</a>
                </span>
              </span>
            </article>
          </div>
          <div v-if="selectedRuntimeEvent" class="watch-detail-panel">
            <div class="timeline-detail-head">
              <strong>{{ runtimeEventKindLabel(selectedRuntimeEvent) }} · {{ runtimeEventNameLabel(selectedRuntimeEvent) }}</strong>
              <button type="button" class="timeline-close" @click="selectedRuntimeEvent = null">
                <X :size="14" />
              </button>
            </div>
            <RuntimeEventDetail
              :event="selectedRuntimeEvent"
              :loading="selectedRuntimeEventLoading"
              :error="selectedRuntimeEventError"
            />
          </div>
        </section>

        <details v-if="runtimeMeta" class="dev-status-panel">
          <summary class="dev-status-summary">
            <span class="dev-status-title">运行诊断</span>
            <span class="dev-status-phase">{{ runtimeMeta.phase || currentPhase || 'preparing' }}</span>
            <span v-if="runtimeWarnings.length || runtimeMeta.alignment_warnings?.length || runtimeMeta.last_error" class="diagnostic-alert">
              {{ runtimeWarnings.length + (runtimeMeta.alignment_warnings?.length || 0) + (runtimeMeta.last_error ? 1 : 0) }} 条警告
            </span>
            <ChevronDown class="diagnostic-chevron" :size="16" aria-hidden="true" />
          </summary>
          <div class="dev-status-body">
            <div class="dev-status-grid">
            <div class="dev-stat intent-stat">
              <span class="dev-stat-label">用户意图</span>
              <strong>{{ runtimeMeta.task_input?.summary || selectedTaskTitle }}</strong>
              <small>{{ totalCount || '?' }} 页 · {{ runtimeMeta.task_input?.template || '动态规划' }}</small>
            </div>
            <div class="dev-stat">
              <span class="dev-stat-label">契约对齐</span>
              <strong :class="['alignment-value', runtimeMeta.alignment_status || 'pending']">{{ runtimeMeta.alignment_status === 'aligned' ? '正常' : runtimeMeta.alignment_status === 'warning' ? '发现偏离' : '等待规划' }}</strong>
              <small>冻结计划 {{ runtimeMeta.plan_slides?.length || 0 }} 页</small>
            </div>
            <div class="dev-stat">
              <span class="dev-stat-label">当前执行</span>
              <strong>{{ runtimeMeta.current_slide?.page_index ? `第 ${runtimeMeta.current_slide.page_index} 页` : runtimeMeta.phase || '准备中' }}</strong>
              <small>{{ runtimeMeta.current_slide?.title || runtimeMeta.phase_detail || '等待执行点更新' }}</small>
            </div>
            <div class="dev-stat">
              <span class="dev-stat-label">运行</span>
              <strong>{{ fmtElapsed(runtimeMeta.elapsed_ms) }}</strong>
              <small>{{ runtimeMeta.phase_detail || phaseDetail || '等待阶段更新' }}</small>
            </div>
            <div v-if="runtimeMeta.plan_slides?.length" class="plan-strip">
              <span
                v-for="slide in runtimeMeta.plan_slides"
                :key="slide.task_id || slide.page_index"
                :class="{ current: runtimeMeta.current_slide?.page_index === slide.page_index }"
                :title="`${slide.title || '未命名'} · ${slide.content_type || 'unknown'}`"
              >{{ slide.page_index }}</span>
            </div>
            <div v-if="runtimeMeta.alignment_warnings?.length" class="alignment-warnings">
              <article v-for="warning in runtimeMeta.alignment_warnings" :key="`${warning.code}-${warning.page_index || 0}`" :class="warning.severity">
                <strong>{{ warning.step }}{{ warning.page_index ? ` · 第 ${warning.page_index} 页` : '' }}</strong>
                <span>{{ warning.message }}</span>
                <small v-if="warning.expected || warning.observed">期望：{{ warning.expected || '-' }} · 实际：{{ warning.observed || '-' }}</small>
              </article>
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
              <small>{{ runtimeTimeline.length }} / {{ runtimeTimelineAll.length }} events</small>
            </div>
            <div class="timeline-filters">
              <button
                v-for="category in runtimeCategoryOptions"
                :key="category.key"
                type="button"
                class="timeline-filter"
                :class="{ active: selectedRuntimeCategory === category.key }"
                :disabled="category.count === 0"
                @click="setRuntimeCategory(category.key)"
              >
                <span>{{ category.label }}</span>
                <strong>{{ category.count }}</strong>
              </button>
            </div>
            <div class="timeline-list">
              <button
                v-for="evt in runtimeTimeline"
                :key="evt.id"
                type="button"
                class="timeline-row"
                :class="[evt.status || 'ok', { selected: selectedRuntimeEvent?.id === evt.id }]"
                @click="selectRuntimeEvent(evt)"
              >
                <span class="timeline-time">{{ fmtElapsed(evt.elapsed_ms) }}</span>
                <span class="timeline-category" :class="runtimeEventCategory(evt)">{{ runtimeEventCategoryLabel(evt) }}</span>
                <span class="timeline-kind">{{ runtimeEventKindLabel(evt) }}</span>
                <span class="timeline-main">
                  <strong>{{ runtimeEventNameLabel(evt) }}</strong>
                  <small>{{ evt.phase || '任务' }} · {{ runtimeEventStatusLabel(evt.status) }}</small>
                </span>
                <span class="timeline-detail" :class="{ muted: !evt.detail }">{{ runtimeEventDetailLabel(evt) }}</span>
              </button>
            </div>
            <div v-if="selectedRuntimeEvent" class="timeline-detail-panel">
              <div class="timeline-detail-head">
                <strong>{{ runtimeEventKindLabel(selectedRuntimeEvent) }} · {{ runtimeEventNameLabel(selectedRuntimeEvent) }}</strong>
                <button type="button" class="timeline-close" @click="selectedRuntimeEvent = null">
                  <X :size="14" />
                </button>
              </div>
              <RuntimeEventDetail
                :event="selectedRuntimeEvent"
                :loading="selectedRuntimeEventLoading"
                :error="selectedRuntimeEventError"
              />
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
                  <span class="file-count">{{ readySlides.length }} / {{ orderedSlides.length }}</span>
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
                  :key="s.key"
                  :task="s.task"
                  :task-id="selectedTask?.id"
                  :file-ready="s.fileReady"
                  :selected="selectedSlides.has(s.key)"
                  :thumbnail-status="thumbnailStatus(s.task.output_file)"
                  :thumbnail-version="thumbnailVersions[thumbnailKey(s.task.output_file)] || 0"
                  :thumbnail-error="thumbnailFailures[thumbnailKey(s.task.output_file)]"
                  @toggle="toggleSelect(s.key)"
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

      <ConversationComposer
          v-model="composerInput"
          class="workspace-composer"
          :task-id="selectedTask ? selectedTask.id : undefined"
          :mode="composerMode"
          :task-title="selectedTask ? selectedTaskTitle : undefined"
        :messages="conversationMessages"
        :streaming-messages="streamingAssistantSegments"
        :streaming-content="streamingAssistant"
        :streaming-timestamp="streamingAssistantStartedAt"
        :history-loading="conversationLoading"
        :submitting="composerLoading || creating"
        :error="composerError || (!selectedTask ? loadError : '')"
        :notice="composerNotice"
        :activity="selectedTask ? liveActivity : undefined"
        :runtime-events="runtimeTimelineAll"
        @load-tool-detail="loadInlineRuntimeEvent"
        @submit="submitComposer"
      />
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
.timeline-filters { padding: 7px 8px; display: flex; gap: 6px; overflow-x: auto; border-bottom: 1px solid var(--divider); background: var(--surface); }
.timeline-filter { min-width: 82px; height: 28px; padding: 0 8px; display: inline-flex; align-items: center; justify-content: space-between; gap: 8px; border: 1px solid var(--divider); border-radius: 4px; color: var(--text-muted); background: var(--surface-muted); font-size: 9px; font-weight: 750; cursor: pointer; }
.timeline-filter strong { color: var(--text-secondary); font-size: 10px; }
.timeline-filter:hover:not(:disabled), .timeline-filter.active { border-color: var(--info); color: var(--text); background: var(--info-soft); }
.timeline-filter:disabled { opacity: 0.45; cursor: default; }
.timeline-list { max-height: 320px; overflow: auto; }
.timeline-row { width: 100%; min-height: 42px; padding: 7px 10px; display: grid; grid-template-columns: 54px 76px 90px minmax(130px, 1fr) minmax(160px, 2fr); align-items: center; gap: 8px; border: 0; border-bottom: 1px solid var(--divider); color: inherit; background: transparent; font-size: 9px; text-align: left; cursor: pointer; }
.timeline-row:last-child { border-bottom: 0; }
.timeline-row.error, .timeline-row.failed, .timeline-row.cancelled { background: var(--danger-soft); }.timeline-row.warning { background: var(--warning-soft); }
.timeline-row:hover, .timeline-row.selected { background: var(--surface-muted); }
.timeline-row.error:hover, .timeline-row.error.selected, .timeline-row.failed:hover, .timeline-row.failed.selected, .timeline-row.cancelled:hover, .timeline-row.cancelled.selected { background: var(--danger-soft); }
.timeline-row.warning:hover, .timeline-row.warning.selected { background: var(--warning-soft); }
.timeline-time, .timeline-kind { color: var(--text-muted); font-family: ui-monospace, monospace; }
.timeline-category { min-width: 0; padding: 3px 5px; border-radius: 3px; color: var(--text-muted); background: var(--surface-muted); font-family: ui-monospace, monospace; font-size: 8px; font-weight: 800; text-align: center; }
.timeline-category.llm { color: var(--info); background: var(--info-soft); }
.timeline-category.tool { color: var(--success); background: var(--success-soft); }
.timeline-category.phase { color: var(--warning); background: var(--warning-soft); }
.timeline-category.delivery { color: var(--action-ink); background: var(--action-soft); }
.timeline-category.compression { color: var(--info); background: var(--info-soft); box-shadow: inset 0 0 0 1px var(--divider); }
.timeline-category.error { color: var(--danger); background: var(--danger-soft); }
.timeline-main { min-width: 0; display: flex; flex-direction: column; }.timeline-main strong { overflow: hidden; color: var(--text-secondary); text-overflow: ellipsis; white-space: nowrap; }.timeline-main small { color: var(--text-muted); }
.timeline-detail { overflow: hidden; color: var(--text-muted); text-overflow: ellipsis; white-space: nowrap; }.timeline-detail.muted { opacity: 0.65; }
.timeline-detail-panel { border-top: 1px solid var(--divider); background: var(--surface-muted); }
.timeline-detail-head { min-height: 36px; padding: 0 10px; display: flex; align-items: center; gap: 8px; color: var(--text-secondary); font-size: 10px; }
.timeline-detail-head strong { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.timeline-close { margin-left: auto; width: 28px; height: 28px; display: grid; place-items: center; border: 0; border-radius: 4px; color: var(--text-muted); background: transparent; cursor: pointer; }
.timeline-close:hover { color: var(--text); background: var(--surface); }
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

.pulse-dot { width: 7px; height: 7px; display: inline-block; border-radius: 50%; background: currentColor; animation: pulse 1.4s ease-in-out infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
@keyframes pulse { 50% { opacity: 0.35; } }

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
  padding: 22px 24px 28px;
  display: flex;
  flex-direction: column;
  overflow: auto;
  background: var(--canvas);
}
.main > .dash-header { order: 1; }
.main > :deep(.progress-panel) { order: 2; }
.main > .execution-watch { order: 3; }
.main > .split-layout { order: 4; }
.main > .final-message { order: 5; }
.main > .dev-status-panel { order: 6; }
.main > .workspace-composer { order: 7; }
.main > .welcome { order: 1; }
.main > .dash-header,
.main > :deep(.progress-panel),
.main > .execution-watch,
.main > .split-layout,
.main > .final-message,
.main > .dev-status-panel,
.main > .workspace-composer,
.main > .welcome { flex-shrink: 0; }

.dash-header {
  margin: 0 0 14px;
  padding: 0;
  align-items: flex-end;
  border: 0;
}
.dash-title { max-width: 900px; color: var(--text); font-size: 18px; font-weight: 730; line-height: 1.35; letter-spacing: 0; white-space: normal; }
.task-query-details { margin-top: 8px; width: min(900px, 100%); }
.task-query-details summary { width: max-content; min-height: 28px; display: flex; align-items: center; color: var(--info); font-size: 10px; cursor: pointer; }
.task-query-details pre { max-height: 260px; margin: 6px 0 0; padding: 10px 12px; overflow: auto; border: 1px solid var(--border); border-radius: 5px; color: var(--text-secondary); background: var(--surface); font: 11px/1.65 ui-monospace, monospace; white-space: pre-wrap; overflow-wrap: anywhere; }
.dash-meta { margin-top: 7px; gap: 12px; }
.dash-id { padding: 3px 6px; border: 1px solid var(--border); border-radius: 4px; color: var(--text-muted); background: var(--surface-muted); font-family: ui-monospace, monospace; }
.dash-duration, .dash-tokens { display: inline-flex; align-items: center; gap: 5px; color: var(--text-muted); }
.dash-duration svg, .dash-tokens svg { width: auto; height: auto; }
.dash-status { display: none; }

.main > :deep(.progress-panel) { margin-bottom: 18px; }

.execution-watch {
  margin: 0 0 18px;
  overflow: hidden;
  border: 1px solid var(--border);
  border-radius: 6px;
  background: var(--surface);
}
.watch-head {
  min-height: 42px;
  padding: 0 12px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  border-bottom: 1px solid var(--divider);
}
.watch-head span {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  color: var(--text);
  font-size: 12px;
  font-weight: 750;
}
.watch-head small {
  color: var(--text-muted);
  font-size: 10px;
}
.watch-list {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
}
.watch-step {
  min-width: 0;
  min-height: 72px;
  padding: 10px 12px;
  display: grid;
  grid-template-columns: 46px 12px minmax(0, 1fr);
  align-items: start;
  gap: 8px;
  border-right: 1px solid var(--divider);
  border-bottom: 1px solid var(--divider);
  background: var(--surface);
  cursor: pointer;
}
.watch-step:hover,
.watch-step.selected { background: var(--surface-muted); }
.watch-step:focus-visible {
  outline: 2px solid var(--info);
  outline-offset: -2px;
}
.watch-step.error,
.watch-step.failed { background: var(--danger-soft); }
.watch-step.warning { background: var(--warning-soft); }
.watch-time {
  color: var(--text-muted);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 10px;
}
.watch-dot {
  width: 8px;
  height: 8px;
  margin-top: 3px;
  border-radius: 50%;
  background: var(--text-muted);
}
.watch-step.planner .watch-dot,
.watch-step.search .watch-dot { background: var(--info); }
.watch-step.render .watch-dot { background: var(--success); }
.watch-step.delivery .watch-dot { background: var(--action-ink); }
.watch-step.error .watch-dot,
.watch-step.failed .watch-dot { background: var(--danger); }
.watch-copy {
  min-width: 0;
  display: grid;
  gap: 4px;
}
.watch-copy strong {
  min-width: 0;
  overflow: hidden;
  color: var(--text-secondary);
  font-size: 11px;
  line-height: 1.35;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.watch-copy small {
  display: -webkit-box;
  overflow: hidden;
  color: var(--text-muted);
  font-size: 10px;
  line-height: 1.45;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}
.watch-sources {
  min-width: 0;
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}
.watch-sources a {
  max-width: 180px;
  padding: 2px 5px;
  overflow: hidden;
  border: 1px solid var(--divider);
  border-radius: 3px;
  color: var(--info);
  background: var(--info-soft);
  font-size: 9px;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.watch-detail-panel {
  border-top: 1px solid var(--divider);
  background: var(--surface);
}
.watch-detail-panel :deep(.runtime-event-detail) {
  border-top: 0;
}

.split-layout {
  min-height: 0;
  flex: 0 0 auto;
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
.intent-stat { grid-column: span 2; }
.intent-stat strong { font-size: 12px; line-height: 1.5; }
.alignment-value.aligned { color: var(--success); }.alignment-value.warning { color: var(--warning); }.alignment-value.pending { color: var(--text-muted); }
.plan-strip { grid-column: 1 / -1; min-height: 42px; padding: 7px; display: flex; flex-wrap: wrap; align-items: center; gap: 5px; border: 1px solid var(--divider); border-radius: 5px; background: var(--surface-muted); }
.plan-strip span { width: 28px; height: 28px; display: grid; place-items: center; border: 1px solid var(--border); border-radius: 4px; color: var(--text-muted); background: var(--surface); font-size: 10px; font-weight: 750; }
.plan-strip span.current { border-color: var(--info); color: #fff; background: var(--info); }
.alignment-warnings { grid-column: 1 / -1; display: grid; gap: 6px; }
.alignment-warnings article { padding: 8px 10px; display: grid; grid-template-columns: 120px minmax(160px, 1fr) minmax(180px, 1.4fr); gap: 8px; border-left: 3px solid var(--warning); color: var(--text-secondary); background: var(--warning-soft); font-size: 10px; }
.alignment-warnings article.error { border-left-color: var(--danger); background: var(--danger-soft); }
.alignment-warnings strong { color: var(--text); }.alignment-warnings small { color: var(--text-muted); overflow-wrap: anywhere; }
.runtime-timeline { border: 1px solid var(--divider); border-radius: 5px; }

.final-message { margin: 18px 0 0; border: 1px solid #b9dfcf; border-radius: 6px; background: var(--success-soft); box-shadow: none; }

.welcome {
  min-height: auto;
  width: min(820px, 100%);
  margin: 0 auto;
  padding: 7vh 0 40px;
  display: flex;
  flex-direction: column;
  justify-content: flex-start;
  align-items: stretch;
  text-align: left;
}
.workspace-composer { width: min(920px, 100%); margin: 22px auto 0; flex: 0 0 auto; }
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
  .main { height: 100%; padding: 20px 22px 26px; }
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
  .watch-head { align-items: flex-start; flex-direction: column; padding: 10px 12px; }
  .watch-list { grid-template-columns: 1fr; }
  .watch-step { grid-template-columns: 42px 12px minmax(0, 1fr); border-right: 0; }
  .slides-list { grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 9px; }
  .slides-toolbar { align-items: flex-start; }
  .toolbar-actions { width: auto; display: flex; flex-wrap: wrap; }
  .tool-btn.primary { grid-column: auto; }
  .examples-grid { grid-template-columns: 1fr; }
  .intent-stat { grid-column: 1 / -1; }
  .alignment-warnings article { grid-template-columns: 1fr; }
  .workspace-composer { margin-top: 16px; }
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
