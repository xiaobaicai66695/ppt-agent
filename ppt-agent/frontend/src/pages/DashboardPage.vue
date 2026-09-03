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
  ListFilter,
  Presentation,
  Square,
  X,
} from 'lucide-vue-next';
import type { TaskInfo, TaskItem, SSEEvent, RuntimeMeta, RuntimeEvent } from '../types';
import { STATUS_LABELS } from '../types';
import {
  fetchTasks, fetchTask, startTask, cancelTask, deleteTask,
  isLoggedIn, continueTask, fetchConversation, fetchRuntimeEvent, routeMessage,
} from '../api';
import { authState } from '../stores/auth';
import AppShell from '../components/AppShell.vue';
import Sidebar from '../components/Sidebar.vue';
import ProgressBar from '../components/ProgressBar.vue';
import EventLog from '../components/EventLog.vue';
import SlidePreviewCard from '../components/SlidePreviewCard.vue';
import ConversationComposer from '../components/ConversationComposer.vue';
import RuntimeTracePanel from '../components/RuntimeTracePanel.vue';
import {
  appendAssistantStreamContent, compactRuntimeEvents, deriveLiveActivity, mergeConversationMessages, mergeRuntimeEvents, mergeRuntimeMeta, mergeSlideDeliveries,
  nextReplayCursor, recoverConversationMessages, summarizeTaskTitle,
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
const conversationStreaming = ref(false);
const manualAgentMode = ref<'chat' | 'pptagent'>('chat');
const chatWebSearch = ref(false);
const chatImageSearch = ref(false);
const STREAMING_RENDER_INTERVAL_MS = 100;
let pendingAssistantDeltas: { content: string; timelineOrder: number }[] = [];
let streamingRenderTimer: ReturnType<typeof setTimeout> | null = null;

const composerMode = computed<'create' | 'queue' | 'continue'>(() => {
  if (!selectedTask.value) return 'create';
  if (selectedTask.value.status === 'conversation') return 'create';
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

function flushAssistantDeltas() {
  if (streamingRenderTimer) {
    clearTimeout(streamingRenderTimer);
    streamingRenderTimer = null;
  }
  const deltas = pendingAssistantDeltas;
  pendingAssistantDeltas = [];
  for (const delta of deltas) appendAssistantDelta(delta.content, delta.timelineOrder);
}

function queueAssistantDelta(content: string, timelineOrder = 0) {
  if (!content) return;
  pendingAssistantDeltas.push({ content, timelineOrder });
  if (streamingRenderTimer) return;
  streamingRenderTimer = setTimeout(flushAssistantDeltas, STREAMING_RENDER_INTERVAL_MS);
}

function clearPendingAssistantDeltas() {
  if (streamingRenderTimer) clearTimeout(streamingRenderTimer);
  streamingRenderTimer = null;
  pendingAssistantDeltas = [];
}

function finalizeAssistantTurn() {
  flushAssistantDeltas();
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
      const routed = await routeMessage(message, '', manualAgentMode.value, chatWebSearch.value, chatImageSearch.value);
      manualAgentMode.value = routed.mode;
      chatWebSearch.value = false;
      chatImageSearch.value = false;
      await adoptConversationTask(routed.task_id);
      if (routed.streaming) {
        conversationStreaming.value = true;
        sseCompleted = false;
        connectSSE(routed.task_id, routed.after_event_id || 0);
      }
      if (routed.intent === 'create' && !routed.needs_confirmation && routed.action !== 'ask_clarification') {
        await promoteConversationTask(routed.task_id);
      } else if (routed.intent === 'plan' || routed.action === 'save_plan') {
        composerNotice.value = routed.draft_id ? `已保存规划草稿 ${routed.draft_id}，任务会话会继续保留。` : '已进入规划对话，可继续补充 DeckSpec。';
      } else if (routed.intent === 'fix') {
        composerError.value = withTaskCandidates(routed.reply || '请先说明要修改的已有演示。', routed.task_candidates || []);
      }
    } catch (error) {
      composerError.value = error instanceof Error ? error.message : '创建任务失败，请重试';
    } finally {
      composerLoading.value = false;
    }
    return;
  }

  const taskId = selectedTask.value.id;
  composerLoading.value = true;
  composerInput.value = '';
  clearPendingAssistantDeltas();
  streamingAssistant.value = '';
  streamingAssistantStartedAt.value = '';
  streamingAssistantSegments.value = [];
  try {
    const routed = await routeMessage(message, taskId, manualAgentMode.value, chatWebSearch.value, chatImageSearch.value);
    manualAgentMode.value = routed.mode;
    chatWebSearch.value = false;
    chatImageSearch.value = false;
    await loadConversation(taskId, true);
    if (routed.intent === 'create' && !routed.needs_confirmation && routed.action !== 'ask_clarification') {
      await promoteConversationTask(taskId);
      return;
    }
    if (routed.intent === 'chat' || routed.action === 'reply') {
      if (routed.streaming) {
        conversationStreaming.value = true;
        sseCompleted = false;
        connectSSE(taskId, routed.after_event_id || 0);
        return;
      }
      if (routed.reply) {
        conversationMessages.value = mergeConversationMessages(conversationMessages.value, [{
          role: 'assistant', content: routed.reply, timestamp: new Date().toISOString(),
        }]);
      }
      composerLoading.value = false;
      return;
    }
    if (routed.intent === 'plan' || routed.action === 'save_plan') {
      composerNotice.value = routed.draft_id ? `已保存规划草稿 ${routed.draft_id}，任务会话会继续保留。` : '已识别为规划请求，可继续补充当前任务。';
      return;
    }
    if (routed.needs_confirmation || routed.action === 'ask_clarification') {
      return;
    }
    if (selectedTask.value?.status === 'conversation') return;
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

async function adoptConversationTask(taskId: string) {
  if (!taskId) throw new Error('服务未返回任务 ID');
  const info = await fetchTask(taskId);
  tasks.value = [info, ...tasks.value.filter(task => task.id !== info.id)];
  await selectTask(info.id);
  return info;
}

async function promoteConversationTask(taskId: string) {
  const info = await startTask(taskId);
  const index = tasks.value.findIndex(task => task.id === info.id);
  if (index >= 0) tasks.value[index] = info;
  else tasks.value = [info, ...tasks.value];
  sseCompleted = false;
  connectSSE(info.id, 0);
  startPolling(info.id);
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

const runtimeTimelineAll = computed(() => compactRuntimeEvents(
  (runtimeMeta.value?.recent_events || []).slice().reverse(),
));
const liveActivity = computed(() => {
  if (conversationStreaming.value) {
    return { label: '正在生成回复', detail: '实时输出中', state: 'running' as const };
  }
  // A durable conversation remains in the `conversation` status between
  // turns. It is idle, not a task that is still running.
  if (selectedTask.value?.status === 'conversation') return undefined;
  return deriveLiveActivity({
    status: selectedTask.value?.status,
    phase: currentPhase.value,
    phaseDetail: phaseDetail.value,
    lastTool: runtimeMeta.value?.last_tool,
    connectionInterrupted: sseConnectionInterrupted.value,
    done: doneCount.value,
    total: totalCount.value,
    error: selectedTask.value?.error,
  });
});

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

const chatSuggestions = [
  '帮我梳理一个产品需求的关键风险',
  '解释一下 RAG 和 Agent 的区别',
  '把这段想法整理成可执行的待办清单',
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
let sseConnectionGeneration = 0;
let terminalSSEFallbackTimer: ReturnType<typeof setTimeout> | null = null;
const TERMINAL_SSE_GRACE_MS = 3000;

function clearTerminalSSEFallback() {
  if (terminalSSEFallbackTimer) clearTimeout(terminalSSEFallbackTimer);
  terminalSSEFallbackTimer = null;
}

function appendRuntimeEvent(event: RuntimeEvent) {
  const current = runtimeMeta.value;
  const kind = (event.kind || '').toLowerCase();
  const isToolEvent = kind.startsWith('tool_') || kind.startsWith('slide_render_');
  runtimeMeta.value = {
    ...(current || { elapsed_ms: 0 }),
    task_id: event.task_id || current?.task_id,
    elapsed_ms: event.elapsed_ms || current?.elapsed_ms || 0,
    phase: event.phase || current?.phase,
    phase_detail: kind === 'phase_changed' ? event.detail || current?.phase_detail : current?.phase_detail,
    last_tool: isToolEvent ? event.name || current?.last_tool : current?.last_tool,
    last_error: event.status === 'error' || event.status === 'failed'
      ? event.detail || current?.last_error
      : current?.last_error,
    recent_events: mergeRuntimeEvents(current?.recent_events || [], [event]),
  };
}

function scheduleTerminalSSEFallback(taskId: string) {
  if (terminalSSEFallbackTimer || sseCompleted || selectedId.value !== taskId) return;
  terminalSSEFallbackTimer = setTimeout(async () => {
    terminalSSEFallbackTimer = null;
    if (sseCompleted || selectedId.value !== taskId) return;
    const latest = await fetchTask(taskId).catch(() => null);
    if (latest?.status === 'running') {
      sseCompleted = false;
      connectSSE(taskId, lastSeenEventID);
      startPolling(taskId);
      return;
    }
    sseCompleted = true;
    finalizeAssistantTurn();
    if (es) { es.close(); es = null; }
    void refreshTask(taskId);
    void loadConversation(taskId, true);
  }, TERMINAL_SSE_GRACE_MS);
}

function connectSSE(taskId: string, afterEventID = 0) {
  if (!taskId) return;
  if (sseCompleted) return; // already received complete, don't reconnect
  clearTerminalSSEFallback();
  const connectionGeneration = ++sseConnectionGeneration;
  if (es) es.close();
  lastSeenEventID = afterEventID;
  const streamURL = afterEventID > 0
    ? `/api/tasks/${taskId}/stream?after_id=${encodeURIComponent(afterEventID)}`
    : `/api/tasks/${taskId}/stream`;
  const source = new EventSource(streamURL);
  es = source;
  activeWorkers.value = 0;


  source.onopen = () => {
	if (connectionGeneration !== sseConnectionGeneration || es !== source || selectedId.value !== taskId) return;
	if (sseConnectionInterrupted.value) {
	  addLog('worker', '实时连接已恢复');
	}
	sseConnectionInterrupted.value = false;
  };


  source.onerror = async (event) => {
	// Server-side `event: error` carries MessageEvent.data and is handled below.
	// Only transport failures should put the connection into reconnecting state.
	if ('data' in event) return;
	if (connectionGeneration !== sseConnectionGeneration || es !== source || selectedId.value !== taskId) return;
	if (sseCompleted || sseConnectionInterrupted.value) return;
	if (conversationStreaming.value) {
	  // A short chat reply can finish before EventSource dispatches the final
	  // named event. Its assistant message is already durable at answer_end, so
	  // reconcile from the conversation snapshot instead of looping forever on
	  // a closed stream.
	  source.close();
	  if (es === source) es = null;
	  const session = await loadConversation(taskId, true);
	  if (connectionGeneration !== sseConnectionGeneration || selectedId.value !== taskId) return;
	  if (!session?.conversation_streaming) {
		conversationStreaming.value = false;
		sseCompleted = true;
		finalizeAssistantTurn();
		return;
	  }
	  lastSeenEventID = nextReplayCursor(lastSeenEventID, session.replay_after_event_id || 0);
	  connectSSE(taskId, lastSeenEventID);
	  return;
	}
	sseConnectionInterrupted.value = true;
	addLog('error', '实时连接暂时中断，浏览器正在自动重连；任务状态仍会通过轮询同步');
  };

  const handler = (e: MessageEvent) => {
	if (connectionGeneration !== sseConnectionGeneration || es !== source || selectedId.value !== taskId) return;
    let evt: SSEEvent;
    try { evt = JSON.parse(e.data); } catch { return; }
    const eventID = e.lastEventId ? Number.parseInt(e.lastEventId, 10) : (evt.id || 0);
    if (eventID > 0 && eventID <= lastSeenEventID) return;
    if (eventID > 0 && lastSeenEventID > 0 && eventID > lastSeenEventID + 1) {
      sseConnectionInterrupted.value = true;
      addLog('error', '实时事件出现缺口，正在从最后确认的位置恢复');
      source.close();
      if (es === source) es = null;
      window.setTimeout(() => {
        if (selectedId.value === taskId && !sseCompleted) connectSSE(taskId, lastSeenEventID);
      }, 0);
      return;
    }
    if (eventID > 0) lastSeenEventID = eventID;

    switch (evt.type) {
      case 'answer': {
        const chunk = evt.content || '';
        if (chunk) {
          queueAssistantDelta(chunk);
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

      case 'runtime_event':
        if (evt.runtime_event) appendRuntimeEvent(evt.runtime_event);
        break;

      case 'error':
        addLog('error', evt.error || evt.content || '');
        break;

      case 'continue_queued':
        composerNotice.value = '已进入反馈处理阶段';
        break;

      case 'continue_complete':
        clearTerminalSSEFallback();
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

      case 'conversation_complete':
        clearTerminalSSEFallback();
        conversationStreaming.value = false;
        sseCompleted = true;
        finalizeAssistantTurn();
        if (es === source) { source.close(); es = null; }
        window.setTimeout(() => void loadConversation(taskId, true), 200);
        break;

      case 'complete':
        clearTerminalSSEFallback();
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

  source.addEventListener('answer', handler);
  source.addEventListener('answer_end', handler);
  source.addEventListener('system_step', handler);
  source.addEventListener('tool_call', handler);
  source.addEventListener('progress', handler);
  source.addEventListener('file_ready', handler);
  source.addEventListener('thumbnail_ready', handler);
  source.addEventListener('thumbnail_error', handler);
  source.addEventListener('token_usage', handler);
  source.addEventListener('runtime_event', handler);
  source.addEventListener('error', handler);
  source.addEventListener('continue_queued', handler);
  source.addEventListener('continue_complete', handler);
  source.addEventListener('conversation_complete', handler);
  source.addEventListener('complete', handler);
}

function disconnectSSE() {
	++sseConnectionGeneration;
  clearTerminalSSEFallback();
  if (es) { es.close(); es = null; }
  stopPolling();
  clearPendingAssistantDeltas();
  streamingAssistant.value = '';
  streamingAssistantStartedAt.value = '';
  streamingAssistantSegments.value = [];
  composerLoading.value = false;
  continuationQueued.value = false;
  conversationStreaming.value = false;
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
		duration.value = info.duration || duration.value;
		currentPhase.value = info.status === 'completed' ? 'complete' : info.status;
		stopPolling();
		scheduleTerminalSSEFallback(taskId);
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
      const session = await loadConversation(id, true);
      if (session?.conversation_streaming) {
        conversationStreaming.value = true;
        sseCompleted = false;
        lastSeenEventID = nextReplayCursor(lastSeenEventID, session.replay_after_event_id || 0);
        connectSSE(id, lastSeenEventID);
      }
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
    const session = await loadConversation(id, true);
    if (session?.conversation_streaming) {
      conversationStreaming.value = true;
      sseCompleted = false;
      lastSeenEventID = nextReplayCursor(0, session.replay_after_event_id || 0);
      connectSSE(id, lastSeenEventID);
    }
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
    :title="selectedTask ? selectedTaskTitle : '工作助手'"
    :eyebrow="selectedTask ? '任务工作区' : '多意图 Agent'"
    content-class="dashboard-shell-content"
  >
    <template #actions>
      <span v-if="selectedTask && selectedTask.status !== 'conversation'" class="top-task-state" :class="selectedTask.status">
        <i aria-hidden="true"></i>
        {{ selectedTask.status === 'running' ? '正在处理' : selectedTask.status === 'completed' ? '演示已交付' : selectedTask.status === 'cancelled' ? '已中断' : '失败' }}
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
        <span>会话</span>
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
          <span class="welcome-kicker">多意图 Agent</span>
          <h2>有什么我可以帮你的？</h2>
          <p>直接告诉我你的需求；需要演示文稿时，系统会自动识别并开始创建。</p>
        </div>

        <div class="welcome-examples">
          <h3>可以这样开始</h3>
          <div class="examples-grid chat-examples">
            <button v-for="suggestion in chatSuggestions" :key="suggestion" class="example-chip" @click="composerInput = suggestion">{{ suggestion }}</button>
          </div>
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
              <span class="pulse-dot"></span>正在处理
            </span>
            <span v-if="selectedTask?.status === 'completed'" class="stat-badge done"> 演示已交付</span>
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

        <RuntimeTracePanel
          :events="runtimeTimelineAll"
          :runtime-meta="runtimeMeta"
          :selected-event="selectedRuntimeEvent"
          :loading="selectedRuntimeEventLoading"
          :error="selectedRuntimeEventError"
          :done="doneCount"
          :total="totalCount"
          @select="selectRuntimeEvent"
          @close="selectedRuntimeEvent = null"
        />

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
          :agent-mode="manualAgentMode"
          :web-search="chatWebSearch"
          :image-search="chatImageSearch"
          :task-title="selectedTask ? selectedTaskTitle : undefined"
        :messages="conversationMessages"
        :streaming-messages="streamingAssistantSegments"
        :streaming-content="streamingAssistant"
        :streaming-timestamp="streamingAssistantStartedAt"
        :history-loading="conversationLoading"
        :submitting="composerLoading"
        :error="composerError || (!selectedTask ? loadError : '')"
        :notice="composerNotice"
        :activity="selectedTask ? liveActivity : undefined"
        :runtime-events="runtimeTimelineAll"
        @load-tool-detail="loadInlineRuntimeEvent"
        @update:agent-mode="manualAgentMode = $event"
        @update:web-search="chatWebSearch = $event"
        @update:image-search="chatImageSearch = $event"
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
.chat-examples { grid-template-columns: 1fr; }
.example-chip { min-height: 48px; padding: 0 13px; border: 1px solid var(--border); border-radius: 6px; color: var(--text-secondary); background: var(--surface); text-align: left; box-shadow: none; cursor: pointer; }
.example-chip:hover { color: var(--text); border-color: var(--border-strong); background: var(--surface-muted); transform: none; }

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
}
</style>
