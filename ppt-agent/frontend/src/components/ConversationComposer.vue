<script setup lang="ts">
import { computed, nextTick, onUnmounted, ref, watch } from 'vue';
import { CircleCheck, LoaderCircle, MessageSquareText, Send, TriangleAlert } from 'lucide-vue-next';
import type { ConversationMessage, LiveActivity, RuntimeEvent } from '../types';
import {
  deriveInlineConversationItems, formatToolPreviewFields, mergeConversationMessages, renderSafeMarkdown,
  runtimeAssistantOutputMessages,
} from '../utils/workbench';

const props = defineProps<{
  taskId?: string;
  modelValue: string;
  mode: 'create' | 'queue' | 'continue';
  taskTitle?: string;
  messages: ConversationMessage[];
  streamingMessages?: ConversationMessage[];
  streamingContent?: string;
  streamingTimestamp?: string;
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
  'load-tool-detail': [eventId: number];
}>();

const thread = ref<HTMLElement | null>(null);
const showHistory = ref(props.messages.length > 0);
const autoFollowThread = ref(true);
const requestedToolEvents = ref(new Set<number>());

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

const hasLiveAssistantStream = computed(() => Boolean(
  props.streamingContent?.trim() || props.streamingMessages?.some(message => message.content?.trim()),
));
const displayMessages = computed(() => mergeConversationMessages(
  // While a turn is in flight, the runtime snapshot contains the same chunks
  // as streamingContent. Reconstructing both produces duplicated prose and
  // makes every snapshot re-render the full transcript.
  hasLiveAssistantStream.value ? [] : runtimeAssistantOutputMessages(props.runtimeEvents || []),
  props.messages,
));
const streamingAlreadyShown = computed(() => {
  const content = props.streamingContent?.trim();
  return Boolean(content && displayMessages.value.some(message => message.role === 'assistant' && message.content.trim() === content));
});
const visibleMessages = computed(() => {
  const streamingMessages = (props.streamingMessages || []).filter(message => message.content?.trim());
  const messages = mergeConversationMessages(displayMessages.value, streamingMessages);
  const content = props.streamingContent?.trim();
  if (content && streamingMessages.length === 0 && !streamingAlreadyShown.value) {
    return mergeConversationMessages(messages, [{
      role: 'assistant',
      content,
      timestamp: props.streamingTimestamp || new Date(0).toISOString(),
    }]);
  }
  return messages;
});
const conversationItems = computed(() => deriveInlineConversationItems(
  visibleMessages.value,
  props.runtimeEvents || [],
));
const visibleItemCount = computed(() => conversationItems.value.length);

let scrollFrame: number | null = null;

function scheduleThreadScroll() {
  if (conversationItems.value.length > 0 || props.streamingContent) showHistory.value = true;
  if (scrollFrame !== null) return;
  scrollFrame = window.requestAnimationFrame(async () => {
    scrollFrame = null;
    const shouldScroll = autoFollowThread.value || !thread.value;
    const previousScrollTop = thread.value?.scrollTop ?? 0;
    await nextTick();
    if (!thread.value) return;
    if (shouldScroll) {
      thread.value.scrollTop = thread.value.scrollHeight;
    } else {
      thread.value.scrollTop = previousScrollTop;
    }
  });
}

watch(() => [conversationItems.value.length, props.streamingContent], scheduleThreadScroll);

onUnmounted(() => {
  if (scrollFrame !== null) window.cancelAnimationFrame(scrollFrame);
});

function isNearThreadBottom(el: HTMLElement): boolean {
  return el.scrollHeight - el.scrollTop - el.clientHeight < 56;
}

function handleThreadScroll() {
  if (!thread.value) return;
  autoFollowThread.value = isNearThreadBottom(thread.value);
}

function handleKeydown(event: KeyboardEvent) {
  if (event.key === 'Enter' && !event.shiftKey) {
    event.preventDefault();
    if (props.modelValue.trim() && !props.submitting) emit('submit');
  }
}

function toolStatusLabel(status: string): string {
  if (status === 'running') return '运行中';
  if (status === 'error' || status === 'failed') return '失败';
  return '已完成';
}

function toolStatusIcon(status: string): 'running' | 'error' | 'ok' {
  if (status === 'running') return 'running';
  if (status === 'error' || status === 'failed') return 'error';
  return 'ok';
}

function toolGroupStatusLabel(status: string): string {
  if (status === 'running') return '持续运行';
  if (status === 'error' || status === 'failed') return '需要处理';
  return '已完成';
}

function previewImageUrl(item: { preview_url?: string; image_url?: string }): string {
  return String(item.preview_url || item.image_url || '').trim();
}

function previewImageLabel(item: { attribution?: string; photographer?: string; description?: string }): string {
  return String(item.attribution || item.photographer || item.description || '图片预览').trim();
}

function sourceHost(url: string): string {
  try {
    return new URL(url).hostname.replace(/^www\./, '');
  } catch {
    return '查看来源';
  }
}

function handleToolToggle(
  event: Event,
  tool: { end_event_id?: number; start_event_id?: number; metadata_loaded?: boolean },
) {
  const details = event.currentTarget as HTMLDetailsElement | null;
  if (!details?.open || tool.metadata_loaded) return;
  const eventId = tool.end_event_id || tool.start_event_id;
  if (!eventId || requestedToolEvents.value.has(eventId)) return;
  requestedToolEvents.value = new Set([...requestedToolEvents.value, eventId]);
  emit('load-tool-detail', eventId);
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
        v-if="conversationItems.length || streamingContent || historyLoading"
        class="history-toggle"
        type="button"
        :aria-expanded="showHistory"
        @click="showHistory = !showHistory"
      >
        {{ showHistory ? '收起对话' : `查看对话 (${visibleItemCount})` }}
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

    <div v-show="showHistory" ref="thread" class="conversation-thread" aria-live="polite" @scroll="handleThreadScroll">
      <div v-if="historyLoading && conversationItems.length === 0" class="history-state">
        <LoaderCircle :size="17" class="spin" />正在恢复对话
      </div>
      <template v-for="item in conversationItems" :key="item.key">
        <article v-if="item.type === 'message'" class="message" :class="item.message.role">
          <span class="message-role">{{ item.message.role === 'user' ? '你' : 'AI' }}</span>
          <div class="markdown-body" v-html="renderSafeMarkdown(item.message.content)"></div>
        </article>
        <article v-else class="message assistant tool-message" :class="toolStatusIcon(item.group.status)">
          <span class="message-role">AI</span>
          <details class="tool-group-card" :open="item.group.status === 'running'">
            <summary class="tool-group-summary">
              <span class="tool-state group" :class="toolStatusIcon(item.group.status)">
                <LoaderCircle v-if="toolStatusIcon(item.group.status) === 'running'" :size="14" class="spin" />
                <TriangleAlert v-else-if="toolStatusIcon(item.group.status) === 'error'" :size="14" />
                <CircleCheck v-else :size="14" />
              </span>
              <span class="tool-copy">
                <strong>{{ item.group.label }}</strong>
                <small>{{ toolGroupStatusLabel(item.group.status) }} · {{ item.group.detail }}</small>
              </span>
            </summary>
            <div class="tool-group-body">
              <details
                v-for="tool in item.group.tools"
                :key="tool.key"
                class="inline-tool-card"
                :open="tool.status === 'running' || tool.image_results.length > 0"
                @toggle="handleToolToggle($event, tool)"
              >
                <summary>
                  <span class="tool-state" :class="toolStatusIcon(tool.status)">
                    <LoaderCircle v-if="toolStatusIcon(tool.status) === 'running'" :size="14" class="spin" />
                    <TriangleAlert v-else-if="toolStatusIcon(tool.status) === 'error'" :size="14" />
                    <CircleCheck v-else :size="14" />
                  </span>
                  <span class="tool-copy">
                    <strong>{{ tool.label }}</strong>
                    <small>{{ toolStatusLabel(tool.status) }} · {{ tool.detail }}</small>
                  </span>
                </summary>
                <div class="tool-card-body">
                  <div v-if="tool.search_results.length" class="tool-search-results">
                    <a
                      v-for="result in tool.search_results"
                      :key="`${result.url}:${result.title}`"
                      class="tool-search-result"
                      :href="result.url || undefined"
                      :target="result.url ? '_blank' : undefined"
                      rel="noreferrer"
                    >
                      <span class="tool-result-head">
                        <strong>{{ result.title }}</strong>
                        <small>{{ result.source || sourceHost(result.url) }}<template v-if="result.date"> · {{ result.date }}</template></small>
                      </span>
                      <p v-if="result.description">{{ result.description }}</p>
                    </a>
                  </div>
                  <div v-else-if="tool.source_urls.length" class="tool-source-list">
                    <a v-for="url in tool.source_urls" :key="url" :href="url" target="_blank" rel="noreferrer">
                      <strong>{{ sourceHost(url) }}</strong><span>{{ url }}</span>
                    </a>
                  </div>
                  <div v-if="tool.image_results.length" class="tool-image-grid">
                    <a
                      v-for="image in tool.image_results"
                      :key="image.id"
                      class="tool-image-preview"
                      :href="image.source_url || image.image_url || previewImageUrl(image)"
                      target="_blank"
                      rel="noreferrer"
                    >
                      <img v-if="previewImageUrl(image)" :src="previewImageUrl(image)" :alt="previewImageLabel(image)" loading="lazy" />
                      <span v-else>无缩略图</span>
                      <small>{{ previewImageLabel(image) }}</small>
                    </a>
                  </div>
                  <div class="tool-preview-columns">
                    <div v-if="formatToolPreviewFields(tool, 'args').length" class="tool-preview-block">
                      <span>调用参数</span>
                      <dl>
                        <template v-for="field in formatToolPreviewFields(tool, 'args')" :key="`${field.label}:${field.value}`">
                          <dt>{{ field.label }}</dt>
                          <dd>{{ field.value }}</dd>
                        </template>
                      </dl>
                    </div>
                    <div v-if="formatToolPreviewFields(tool, 'result').length" class="tool-preview-block">
                      <span>返回结果</span>
                      <dl>
                        <template v-for="field in formatToolPreviewFields(tool, 'result')" :key="`${field.label}:${field.value}`">
                          <dt>{{ field.label }}</dt>
                          <dd>{{ field.value }}</dd>
                        </template>
                      </dl>
                    </div>
                  </div>
                  <p
                    v-if="requestedToolEvents.has(tool.end_event_id || tool.start_event_id || 0) && !tool.metadata_loaded && !tool.search_results.length && !tool.image_results.length"
                    class="tool-detail-loading"
                  >
                    <LoaderCircle :size="13" class="spin" />正在读取完整调用结果
                  </p>
                </div>
              </details>
            </div>
          </details>
        </article>
      </template>
      <div v-if="historyLoading && conversationItems.length > 0" class="history-refresh">
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
.markdown-body { min-width: 0; padding: 6px 9px; border: 1px solid var(--border); border-radius: 6px; color: var(--text); background: var(--surface); font-size: 13px; line-height: 1.65; overflow-wrap: break-word; word-break: normal; }
.message.user .markdown-body { border-color: #bdd7d3; background: var(--action-soft); }
.message.streaming .markdown-body { border-left: 3px solid var(--info); }
.tool-message .message-role { color: var(--info); background: var(--info-soft); }
.tool-message.error .message-role { color: var(--danger); background: var(--danger-soft); }
.tool-group-card,
.inline-tool-card {
  min-width: 0;
  overflow: hidden;
  border: 1px solid var(--border);
  border-radius: 6px;
  background: var(--surface);
}
.tool-group-card {
  border-color: #cdd8dc;
  background: #fbfcfc;
}
.tool-group-summary,
.inline-tool-card summary {
  min-height: 38px;
  padding: 7px 9px;
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  color: var(--text-secondary);
}
.tool-group-summary {
  min-height: 42px;
  background: #f8fafb;
}
.tool-group-summary:hover,
.inline-tool-card summary:hover { background: var(--surface-muted); }
.tool-group-body {
  padding: 9px;
  display: grid;
  gap: 8px;
  border-top: 1px solid var(--divider);
  background: #fff;
}
.tool-state {
  width: 22px;
  height: 22px;
  flex: 0 0 auto;
  display: grid;
  place-items: center;
  border-radius: 999px;
  color: var(--success);
  background: var(--success-soft);
}
.tool-state.running { color: var(--info); background: var(--info-soft); }
.tool-state.error { color: var(--danger); background: var(--danger-soft); }
.tool-state.group {
  width: 24px;
  height: 24px;
}
.tool-copy {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.tool-copy strong {
  color: var(--text);
  font-size: 12px;
}
.tool-copy small {
  min-width: 0;
  overflow: hidden;
  color: var(--text-muted);
  font-size: 10px;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.tool-card-body {
  padding: 9px;
  display: grid;
  gap: 9px;
  border-top: 1px solid var(--divider);
  background: var(--surface);
}
.tool-search-results {
  display: grid;
  gap: 7px;
}
.tool-search-result {
  min-width: 0;
  padding: 9px 10px;
  display: grid;
  gap: 5px;
  border: 1px solid var(--divider);
  border-radius: 5px;
  color: var(--text);
  background: #fbfcfd;
  text-decoration: none;
}
.tool-search-result:hover { border-color: #aac9c5; background: var(--action-soft); }
.tool-result-head { min-width: 0; display: flex; align-items: baseline; justify-content: space-between; gap: 10px; }
.tool-result-head strong { min-width: 0; color: var(--action-ink); font-size: 12px; line-height: 1.45; }
.tool-result-head small { flex: 0 0 auto; color: var(--text-muted); font-size: 10px; }
.tool-search-result p { margin: 0; color: var(--text-secondary); font-size: 11px; line-height: 1.6; }
.tool-source-list {
  display: grid;
  gap: 6px;
}
.tool-source-list a {
  max-width: 100%;
  padding: 6px 8px;
  display: flex;
  align-items: baseline;
  gap: 8px;
  overflow: hidden;
  border: 1px solid var(--divider);
  border-radius: 4px;
  color: var(--info);
  background: var(--info-soft);
  font-size: 10px;
  text-decoration: none;
}
.tool-source-list a strong { flex: 0 0 auto; color: var(--action-ink); }
.tool-source-list a span { min-width: 0; overflow: hidden; color: var(--text-muted); text-overflow: ellipsis; white-space: nowrap; }
.tool-image-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(118px, 1fr));
  gap: 8px;
}
.tool-image-preview {
  min-width: 0;
  overflow: hidden;
  border: 1px solid var(--divider);
  border-radius: 5px;
  color: var(--text-secondary);
  background: var(--surface-muted);
  text-decoration: none;
}
.tool-image-preview img,
.tool-image-preview > span {
  width: 100%;
  height: 74px;
  display: grid;
  place-items: center;
  object-fit: cover;
  color: var(--text-muted);
  font-size: 10px;
}
.tool-image-preview small {
  min-width: 0;
  padding: 5px 6px;
  display: block;
  overflow: hidden;
  color: var(--text-muted);
  font-size: 9px;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.tool-preview-columns {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 8px;
}
.tool-detail-loading { margin: 0; display: flex; align-items: center; gap: 6px; color: var(--text-muted); font-size: 10px; }
.tool-preview-block {
  min-width: 0;
  display: grid;
  gap: 6px;
}
.tool-preview-block span {
  color: var(--text-muted);
  font-size: 9px;
  font-weight: 800;
}
.tool-preview-block dl {
  margin: 0;
  display: grid;
  grid-template-columns: minmax(64px, max-content) minmax(0, 1fr);
  overflow: hidden;
  border: 1px solid var(--divider);
  border-radius: 5px;
  color: var(--text-secondary);
  background: var(--surface-muted);
  font-size: 11px;
}
.tool-preview-block dt,
.tool-preview-block dd {
  min-width: 0;
  margin: 0;
  padding: 6px 8px;
  border-bottom: 1px solid var(--divider);
}
.tool-preview-block dt {
  color: var(--text-muted);
  background: #fff;
  font-weight: 800;
  white-space: nowrap;
}
.tool-preview-block dd {
  color: var(--text);
  overflow-wrap: anywhere;
}
.tool-preview-block dt:last-of-type,
.tool-preview-block dd:last-of-type {
  border-bottom: 0;
}
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
  .tool-result-head { align-items: flex-start; flex-direction: column; gap: 2px; }
  .input-shell textarea { font-size: 16px; }
}

@media (prefers-reduced-motion: reduce) { .spin { animation: none; } }
</style>
