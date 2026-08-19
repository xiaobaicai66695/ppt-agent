<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue';
import { CircleCheck, LoaderCircle, MessageSquareText, Send, TriangleAlert } from 'lucide-vue-next';
import { fetchRuntimeEvent } from '../api';
import type { ConversationMessage, LiveActivity, RuntimeEvent } from '../types';
import { renderSafeMarkdown } from '../utils/workbench';
import RuntimeJsonTree from './RuntimeJsonTree.vue';

const props = defineProps<{
  taskId?: string;
  modelValue: string;
  mode: 'create' | 'queue' | 'continue';
  taskTitle?: string;
  messages: ConversationMessage[];
  streamingContent?: string;
  historyLoading?: boolean;
  submitting?: boolean;
  error?: string;
  notice?: string;
  activity?: LiveActivity;
  runtimeEvents?: RuntimeEvent[];
}>();

const emit = defineEmits<{
  'update:modelValue': [value: string];
  submit: [];
}>();

const thread = ref<HTMLElement | null>(null);
const showHistory = ref(props.messages.length > 0);
const traceDetailCache = ref<Record<number, RuntimeEvent>>({});
const traceLoading = ref<Record<number, boolean>>({});
const traceErrors = ref<Record<number, string>>({});

const modeLabel = computed(() => ({
  create: '创建演示',
  queue: '排队反馈',
  continue: '继续修改',
})[props.mode]);
const helper = computed(() => ({
  create: '描述主题、受众、页数和希望强调的内容',
  queue: '反馈会在当前生成结束后自动处理',
  continue: '说明要修改的页面、内容或视觉方向',
})[props.mode]);
const placeholder = computed(() => props.mode === 'create'
  ? '例如：为产品评审制作 10 页演示，面向研发负责人，重点说明架构与风险'
  : '描述希望如何改进这套演示');
const traceEvents = computed(() => (props.runtimeEvents || [])
  .filter(event => {
    const kind = (event.kind || '').toLowerCase();
    return kind.includes('llm') && !kind.includes('start');
  })
  .sort((a, b) => a.id - b.id)
  .slice(-30));

watch(() => [props.messages.length, props.streamingContent], async () => {
  if (props.messages.length > 0 || props.streamingContent) showHistory.value = true;
  await nextTick();
  if (thread.value) thread.value.scrollTop = thread.value.scrollHeight;
});

function handleKeydown(event: KeyboardEvent) {
  if (event.key === 'Enter' && !event.shiftKey) {
    event.preventDefault();
    if (props.modelValue.trim() && !props.submitting) emit('submit');
  }
}

function traceTitle(event: RuntimeEvent): string {
  const kind = (event.kind || '').toLowerCase();
  if (kind.includes('start')) return '模型输入上下文';
  if (kind.includes('error')) return '模型调用错误';
  return '模型显式思考';
}

function traceDetail(event: RuntimeEvent): string {
  if (event.detail) return event.detail;
  const metadata = event.metadata || {};
  if (typeof metadata.reasoning_preview === 'string') return metadata.reasoning_preview;
  if (typeof metadata.output_preview === 'string') return metadata.output_preview;
  return '展开查看模型显式输出与上下文';
}

function compactMetadata(event: RuntimeEvent): Record<string, unknown> {
  const metadata = event.metadata || {};
  const preferred = ['assistant_output', 'output_preview', 'reasoning_preview', 'history', 'assistant_message', 'error'];
  const compact: Record<string, unknown> = {};
  for (const key of preferred) {
    if (metadata[key] !== undefined && metadata[key] !== null && metadata[key] !== '') compact[key] = metadata[key];
  }
  return Object.keys(compact).length ? compact : metadata;
}

function traceEventWithDetails(event: RuntimeEvent): RuntimeEvent {
  return traceDetailCache.value[event.id] ? { ...event, ...traceDetailCache.value[event.id], metadata_loaded: true } : event;
}

function traceEventMetadata(event: RuntimeEvent): Record<string, unknown> {
  return compactMetadata(traceEventWithDetails(event));
}

async function handleTraceToggle(domEvent: Event, event: RuntimeEvent) {
  const target = domEvent.currentTarget as HTMLDetailsElement | null;
  if (!target?.open || !props.taskId || event.metadata_loaded || traceDetailCache.value[event.id] || traceLoading.value[event.id]) {
    return;
  }
  traceLoading.value = { ...traceLoading.value, [event.id]: true };
  traceErrors.value = { ...traceErrors.value, [event.id]: '' };
  try {
    const detail = await fetchRuntimeEvent(props.taskId, event.id);
    traceDetailCache.value = { ...traceDetailCache.value, [event.id]: detail };
  } catch (error) {
    traceErrors.value = {
      ...traceErrors.value,
      [event.id]: error instanceof Error ? error.message : '事件详情加载失败',
    };
  } finally {
    traceLoading.value = { ...traceLoading.value, [event.id]: false };
  }
}
</script>

<template>
  <section class="conversation-composer" :class="mode" aria-label="演示生成与 AI 对话">
    <header class="composer-head">
      <span class="composer-identity">
        <MessageSquareText :size="18" />
        <span>
          <strong>{{ modeLabel }}</strong>
          <small>{{ taskTitle || helper }}</small>
        </span>
      </span>
      <button
        v-if="messages.length || streamingContent || historyLoading"
        class="history-toggle"
        type="button"
        :aria-expanded="showHistory"
        @click="showHistory = !showHistory"
      >
        {{ showHistory ? '收起对话' : `查看对话 (${messages.length})` }}
      </button>
    </header>

    <div
      v-if="activity"
      class="activity-line"
      :class="activity.state"
      role="status"
      :aria-live="activity.state === 'error' ? 'assertive' : 'polite'"
    >
      <LoaderCircle v-if="activity.state === 'running'" :size="15" class="spin" />
      <CircleCheck v-else-if="activity.state === 'success'" :size="15" />
      <TriangleAlert v-else-if="activity.state === 'error'" :size="15" />
      <span class="activity-copy">
        <strong>{{ activity.label }}</strong>
        <small v-if="activity.detail">{{ activity.detail }}</small>
      </span>
    </div>

    <div v-show="showHistory" ref="thread" class="conversation-thread" aria-live="polite">
      <div v-if="historyLoading && messages.length === 0" class="history-state">
        <LoaderCircle :size="17" class="spin" />正在恢复对话
      </div>
      <article v-for="(message, index) in messages" :key="`${message.timestamp}-${index}`" class="message" :class="message.role">
        <span class="message-role">{{ message.role === 'user' ? '你' : 'AI' }}</span>
        <div class="markdown-body" v-html="renderSafeMarkdown(message.content)"></div>
      </article>
      <article v-if="streamingContent" class="message assistant streaming">
        <span class="message-role">AI</span>
        <div class="markdown-body" v-html="renderSafeMarkdown(streamingContent)"></div>
      </article>
      <section v-if="traceEvents.length" class="agent-trace" aria-label="AI 思维链">
        <div class="trace-head">
          <MessageSquareText :size="14" />
          <span>AI 思维链</span>
        </div>
        <details
          v-for="event in traceEvents"
          :key="event.id"
          class="trace-event"
          @toggle="handleTraceToggle($event, event)"
        >
          <summary>
            <span>{{ traceTitle(event) }}</span>
            <small>{{ traceDetail(event) }}</small>
          </summary>
          <div v-if="traceLoading[event.id]" class="trace-loading">正在加载完整事件...</div>
          <div v-else-if="traceErrors[event.id]" class="trace-loading error">{{ traceErrors[event.id] }}</div>
          <RuntimeJsonTree v-else label="details" :value="traceEventMetadata(event)" :default-open="true" />
        </details>
      </section>
      <div v-if="historyLoading && messages.length > 0" class="history-refresh">
        <LoaderCircle :size="14" class="spin" />同步历史中
      </div>
    </div>

    <label class="input-label" for="unified-conversation-input">{{ helper }}</label>
    <div class="input-shell">
      <textarea
        id="unified-conversation-input"
        :value="modelValue"
        rows="3"
        :placeholder="placeholder"
        :disabled="submitting"
        @input="emit('update:modelValue', ($event.target as HTMLTextAreaElement).value)"
        @keydown="handleKeydown"
      ></textarea>
      <button
        class="submit-button"
        type="button"
        :disabled="!modelValue.trim() || submitting"
        :aria-label="modeLabel"
        :title="modeLabel"
        @click="emit('submit')"
      >
        <LoaderCircle v-if="submitting" :size="19" class="spin" />
        <Send v-else :size="19" />
      </button>
    </div>
    <p v-if="notice" class="composer-notice">{{ notice }}</p>
    <p v-if="error" class="composer-error" role="alert">{{ error }}</p>
  </section>
</template>

<style scoped>
.conversation-composer {
  width: 100%;
  border: 1px solid var(--border-strong);
  border-radius: 8px;
  background: var(--surface);
  box-shadow: var(--shadow-sm);
}
.composer-head { min-height: 52px; padding: 9px 12px; display: flex; align-items: center; justify-content: space-between; gap: 12px; border-bottom: 1px solid var(--divider); }
.composer-identity { min-width: 0; display: flex; align-items: center; gap: 9px; color: var(--action-ink); }
.composer-identity > span { min-width: 0; display: flex; flex-direction: column; }
.composer-identity strong { color: var(--text); font-size: 13px; }
.composer-identity small { margin-top: 2px; overflow: hidden; color: var(--text-muted); font-size: 11px; text-overflow: ellipsis; white-space: nowrap; }
.history-toggle { min-height: 34px; padding: 0 9px; flex: 0 0 auto; border: 0; border-radius: 5px; color: var(--text-secondary); background: transparent; font-size: 11px; cursor: pointer; }
.history-toggle:hover { background: var(--surface-muted); }
.activity-line { min-height: 38px; padding: 7px 12px; display: flex; align-items: center; gap: 8px; border-bottom: 1px solid var(--divider); color: var(--text-secondary); background: var(--surface-muted); }
.activity-line.running { color: var(--info); }.activity-line.success { color: var(--success); }.activity-line.error { color: var(--danger); }
.activity-copy { min-width: 0; display: flex; align-items: baseline; gap: 8px; }
.activity-copy strong { color: currentColor; font-size: 11px; }.activity-copy small { overflow: hidden; color: var(--text-muted); font-size: 10px; text-overflow: ellipsis; white-space: nowrap; }

.conversation-thread { max-height: 340px; padding: 12px; display: flex; flex-direction: column; gap: 10px; overflow: auto; border-bottom: 1px solid var(--divider); background: var(--surface-muted); }
.message { display: grid; grid-template-columns: 28px minmax(0, 1fr); align-items: start; gap: 8px; }
.message-role { width: 28px; height: 28px; display: grid; place-items: center; border-radius: 5px; color: var(--text-secondary); background: var(--surface-pressed); font-size: 10px; font-weight: 800; }
.message.user .message-role { color: #fff; background: var(--action-ink); }
.markdown-body { min-width: 0; padding: 6px 9px; border: 1px solid var(--border); border-radius: 6px; color: var(--text); background: var(--surface); font-size: 13px; line-height: 1.65; overflow-wrap: anywhere; }
.message.user .markdown-body { border-color: #bdd7d3; background: var(--action-soft); }
.message.streaming .markdown-body { border-left: 3px solid var(--info); }
.agent-trace { margin-left: 36px; border: 1px solid var(--border); border-radius: 6px; overflow: hidden; background: var(--surface); }
.trace-head { min-height: 34px; padding: 0 10px; display: flex; align-items: center; gap: 7px; color: var(--text-secondary); border-bottom: 1px solid var(--divider); background: var(--surface-muted); font-size: 11px; font-weight: 800; }
.trace-event { border-top: 1px solid var(--divider); }
.trace-event:first-of-type { border-top: 0; }
.trace-event summary { min-height: 34px; padding: 7px 10px; display: grid; grid-template-columns: 112px minmax(0, 1fr); align-items: center; gap: 8px; color: var(--text-secondary); cursor: pointer; list-style: none; font-size: 11px; }
.trace-event summary::-webkit-details-marker { display: none; }
.trace-event summary span { color: var(--text); font-weight: 750; }
.trace-event summary small { overflow: hidden; color: var(--text-muted); text-overflow: ellipsis; white-space: nowrap; }
.trace-event :deep(.json-tree) { padding: 8px 10px 10px; border-top: 1px solid var(--divider); background: var(--surface); }
.trace-loading { padding: 8px 10px; border-top: 1px solid var(--divider); color: var(--text-muted); background: var(--surface); font-size: 11px; }
.trace-loading.error { color: var(--danger); }
.markdown-body :deep(h1), .markdown-body :deep(h2), .markdown-body :deep(h3), .markdown-body :deep(h4) { margin: 0 0 7px; font-size: 14px; line-height: 1.4; }
.markdown-body :deep(p) { margin: 0 0 7px; }.markdown-body :deep(p:last-child) { margin-bottom: 0; }
.markdown-body :deep(ul), .markdown-body :deep(ol) { margin: 4px 0 8px; padding-left: 22px; }
.markdown-body :deep(li + li) { margin-top: 3px; }
.markdown-body :deep(pre) { margin: 7px 0; padding: 10px; overflow: auto; border-radius: 5px; color: #eef2f2; background: #202527; }
.markdown-body :deep(code) { padding: 1px 4px; border-radius: 3px; background: var(--surface-pressed); font-family: ui-monospace, monospace; font-size: .9em; }
.markdown-body :deep(pre code) { padding: 0; background: transparent; }
.markdown-body :deep(blockquote) { margin: 7px 0; padding-left: 10px; border-left: 3px solid var(--border-strong); color: var(--text-secondary); }
.markdown-body :deep(.md-table-wrap) { max-width: 100%; overflow: auto; }.markdown-body :deep(table) { width: 100%; border-collapse: collapse; font-size: 11px; }.markdown-body :deep(th), .markdown-body :deep(td) { padding: 6px 7px; border: 1px solid var(--border); text-align: left; }.markdown-body :deep(th) { background: var(--surface-muted); }
.history-state, .history-refresh { min-height: 48px; display: flex; align-items: center; justify-content: center; gap: 7px; color: var(--text-muted); font-size: 11px; }
.history-refresh { min-height: 28px; }

.input-label { display: block; padding: 10px 12px 0; color: var(--text-muted); font-size: 10px; font-weight: 700; }
.input-shell { padding: 7px 9px 10px 12px; display: grid; grid-template-columns: minmax(0, 1fr) 44px; align-items: end; gap: 8px; }
.input-shell textarea { width: 100%; min-height: 66px; max-height: 180px; padding: 7px 0; resize: vertical; border: 0; outline: 0; color: var(--text); background: transparent; font: inherit; font-size: 14px; line-height: 1.55; }
.submit-button { width: 44px; height: 44px; display: grid; place-items: center; border: 1px solid var(--action-ink); border-radius: 6px; color: #fff; background: var(--action-ink); cursor: pointer; }
.submit-button:disabled { border-color: var(--border-strong); color: var(--text-disabled); background: var(--surface-pressed); cursor: not-allowed; }
.composer-notice, .composer-error { margin: -3px 12px 10px; font-size: 11px; }.composer-notice { color: var(--info); }.composer-error { color: var(--danger); }
.spin { animation: spin .9s linear infinite; } @keyframes spin { to { transform: rotate(360deg); } }

@media (max-width: 620px) {
  .composer-head { align-items: flex-start; }
  .composer-identity small { white-space: normal; }
  .conversation-thread { max-height: 46dvh; }
  .markdown-body { font-size: 12px; }
  .agent-trace { margin-left: 0; }
  .trace-event summary { grid-template-columns: 1fr; }
  .input-shell textarea { font-size: 16px; }
}

@media (prefers-reduced-motion: reduce) { .spin { animation: none; } }
</style>
