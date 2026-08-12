<script setup lang="ts">
import { computed } from 'vue';
import type { RuntimeEvent } from '../types';
import RuntimeJsonTree from './RuntimeJsonTree.vue';

const props = defineProps<{
  event: RuntimeEvent;
  loading?: boolean;
  error?: string;
}>();

type RuntimeRecord = Record<string, unknown>;

const metadata = computed<RuntimeRecord>(() => (
  props.event.metadata && typeof props.event.metadata === 'object'
    ? props.event.metadata as RuntimeRecord
    : {}
));

const eventSummary = computed(() => {
  const { metadata: _metadata, metadata_loaded: _loaded, ...summary } = props.event;
  return summary;
});

const isCompression = computed(() => props.event.kind === 'compression' || props.event.kind === 'planner_context_compressed');
const sourceUrls = computed(() => (
  Array.isArray(metadata.value.source_urls)
    ? metadata.value.source_urls.map(item => String(item).trim()).filter(Boolean).slice(0, 5)
    : []
));
const observationRows = computed(() => {
  const rows: Array<{ label: string; value: string }> = [];
  const fields: Array<[string, string]> = [
    ['search_query', '搜索关键词'],
    ['search_reason', '搜索原因'],
    ['file_path', '读取文件'],
    ['slide_count', '规划页数'],
    ['template', '模板'],
    ['theme', '配色'],
    ['background', '背景'],
    ['task_id', '页面任务'],
    ['content_type', '页面类型'],
    ['output_file', '输出文件'],
  ];
  for (const [key, label] of fields) {
    const value = metadata.value[key];
    if (value !== undefined && value !== null && String(value).trim() !== '') {
      rows.push({ label, value: String(value).trim() });
    }
  }
  return rows;
});

function metadataNumber(key: string): number {
  const value = Number(metadata.value[key] ?? 0);
  return Number.isFinite(value) ? value : 0;
}

const compressionRows = computed(() => [
  {
    label: '消息数',
    before: metadataNumber('before_messages').toLocaleString(),
    after: metadataNumber('after_messages').toLocaleString(),
    delta: `移除 ${metadataNumber('removed_messages').toLocaleString()} 条`,
  },
  {
    label: 'Token',
    before: metadataNumber('before_tokens').toLocaleString(),
    after: metadataNumber('after_tokens').toLocaleString(),
    delta: `节省 ${metadataNumber('saved_tokens').toLocaleString()} · ${String(metadata.value.saved_pct || '0%')}`,
  },
]);

const compressionIntent = computed(() => String(metadata.value.user_intent_summary || '').trim());
const preservedRequirements = computed(() => (
  Array.isArray(metadata.value.preserved_requirements)
    ? metadata.value.preserved_requirements.map(item => String(item).trim()).filter(Boolean)
    : []
));

const historyMessages = computed(() => {
  const value = metadata.value.history;
  return Array.isArray(value) ? value.filter(isRecord) : [];
});

const reasoningPreview = computed(() => String(metadata.value.reasoning_preview || '').trim());
const toolCallDetails = computed(() => {
  const value = metadata.value.tool_call_details;
  return Array.isArray(value) ? value.filter(isRecord) : [];
});
const assistantMessage = computed(() => (
  isRecord(metadata.value.assistant_message) ? metadata.value.assistant_message : null
));

const readableSections = computed(() => {
  const sections: Array<{ key: string; title: string; value: unknown; mode: 'markdown' | 'json' }> = [];
  const candidates: Array<[string, string]> = [
    ['args', '工具参数'],
    ['result', '工具结果'],
    ['output_preview', '模型输出摘要'],
    ['error', '错误'],
  ];
  for (const [key, title] of candidates) {
    if (metadata.value[key] !== undefined && metadata.value[key] !== null && metadata.value[key] !== '') {
      const parsed = parseMaybeJSON(metadata.value[key]);
      sections.push({
        key,
        title,
        value: parsed,
        mode: typeof parsed === 'string' ? 'markdown' : 'json',
      });
    }
  }
  return sections;
});

function isRecord(value: unknown): value is RuntimeRecord {
  return value !== null && typeof value === 'object' && !Array.isArray(value);
}

function parseMaybeJSON(value: unknown): unknown {
  if (typeof value !== 'string') return value;
  const text = value.trim();
  if (!text || !/^[\[{]/.test(text)) return normalizeNewlines(value);
  try {
    return JSON.parse(text);
  } catch {
    return normalizeNewlines(value);
  }
}

function normalizeNewlines(value: string): string {
  return value.replace(/\r\n/g, '\n').replace(/\r/g, '\n');
}

function displayRole(message: RuntimeRecord): string {
  return String(message.role || 'message');
}

function messageTitle(message: RuntimeRecord, index: number): string {
  const role = displayRole(message);
  const toolName = typeof message.tool_name === 'string' ? ` · ${message.tool_name}` : '';
  return `#${index + 1} ${role}${toolName}`;
}

function messageContent(message: RuntimeRecord): string {
  if (typeof message.content_preview === 'string') return normalizeNewlines(message.content_preview);
  return typeof message.content === 'string' ? normalizeNewlines(message.content) : '';
}

function messageExtra(message: RuntimeRecord): RuntimeRecord {
  const { content: _content, content_preview: _contentPreview, role: _role, ...rest } = message;
  return rest;
}

function toolCallTitle(call: RuntimeRecord, index: number): string {
  return `#${index + 1} ${String(call.name || 'tool')}`;
}

function toolCallArguments(call: RuntimeRecord): string {
  return normalizeNewlines(String(call.arguments_preview || '').trim());
}

function hasExtra(message: RuntimeRecord): boolean {
  return Object.keys(messageExtra(message)).length > 0;
}

function renderMarkdown(value: unknown): string {
  const text = normalizeNewlines(String(value ?? ''));
  const chunks = text.split(/```/);
  return chunks.map((chunk, index) => (
    index % 2 === 1 ? `<pre><code>${escapeHtml(chunk)}</code></pre>` : renderMarkdownLines(chunk)
  )).join('');
}

function renderMarkdownLines(text: string): string {
  const lines = text.split('\n');
  const html: string[] = [];
  let listType: 'ul' | 'ol' | '' = '';
  const closeList = () => {
    if (listType) {
      html.push(`</${listType}>`);
      listType = '';
    }
  };

  for (const line of lines) {
    const trimmed = line.trim();
    if (!trimmed) {
      closeList();
      html.push('<br>');
      continue;
    }
    const heading = /^(#{1,4})\s+(.+)$/.exec(trimmed);
    if (heading) {
      closeList();
      const level = Math.min(heading[1].length + 2, 6);
      html.push(`<h${level}>${inlineMarkdown(heading[2])}</h${level}>`);
      continue;
    }
    const bullet = /^[-*]\s+(.+)$/.exec(trimmed);
    if (bullet) {
      if (listType !== 'ul') {
        closeList();
        listType = 'ul';
        html.push('<ul>');
      }
      html.push(`<li>${inlineMarkdown(bullet[1])}</li>`);
      continue;
    }
    const numbered = /^\d+\.\s+(.+)$/.exec(trimmed);
    if (numbered) {
      if (listType !== 'ol') {
        closeList();
        listType = 'ol';
        html.push('<ol>');
      }
      html.push(`<li>${inlineMarkdown(numbered[1])}</li>`);
      continue;
    }
    closeList();
    if (trimmed.startsWith('>')) {
      html.push(`<blockquote>${inlineMarkdown(trimmed.replace(/^>\s?/, ''))}</blockquote>`);
    } else {
      html.push(`<p>${inlineMarkdown(line)}</p>`);
    }
  }
  closeList();
  return html.join('');
}

function inlineMarkdown(text: string): string {
  return escapeHtml(text)
    .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
    .replace(/`([^`]+?)`/g, '<code>$1</code>');
}

function escapeHtml(value: string): string {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;');
}
</script>

<template>
  <div class="runtime-event-detail">
    <div v-if="loading" class="runtime-detail-state">正在加载完整事件...</div>
    <div v-else-if="error" class="runtime-detail-state error">{{ error }}</div>

    <section v-if="observationRows.length || sourceUrls.length" class="runtime-detail-section observation-section">
      <h4>观察摘要</h4>
      <div v-if="observationRows.length" class="observation-grid">
        <div v-for="row in observationRows" :key="row.label" class="observation-row">
          <span>{{ row.label }}</span>
          <strong>{{ row.value }}</strong>
        </div>
      </div>
      <div v-if="sourceUrls.length" class="source-list">
        <span>来源 URL</span>
        <a
          v-for="url in sourceUrls"
          :key="url"
          :href="url"
          target="_blank"
          rel="noreferrer"
        >{{ url }}</a>
      </div>
    </section>

    <section v-if="isCompression" class="runtime-detail-section compression-section">
      <h4>压缩前后</h4>
      <div class="compression-diff">
        <div class="compression-head"><span>指标</span><span>压缩前</span><span>压缩后</span><span>变化</span></div>
        <div v-for="row in compressionRows" :key="row.label" class="compression-row">
          <strong>{{ row.label }}</strong><span>{{ row.before }}</span><span>{{ row.after }}</span><small>{{ row.delta }}</small>
        </div>
      </div>
      <div v-if="compressionIntent" class="intent-anchor">
        <strong>用户目标锚点</strong>
        <p>{{ compressionIntent }}</p>
      </div>
      <div v-if="preservedRequirements.length" class="preserved-requirements">
        <strong>保留要求</strong>
        <ul><li v-for="requirement in preservedRequirements" :key="requirement">{{ requirement }}</li></ul>
      </div>
    </section>

    <section class="runtime-detail-section">
      <h4>事件字段</h4>
      <RuntimeJsonTree label="event" :value="eventSummary" :default-open="true" />
    </section>

    <section v-if="reasoningPreview || toolCallDetails.length || assistantMessage" class="runtime-detail-section model-observation-section">
      <h4>模型思考与调用</h4>
      <div v-if="reasoningPreview" class="model-reasoning">
        <strong>显式思考摘要</strong>
        <p>{{ reasoningPreview }}</p>
      </div>
      <details v-if="assistantMessage" class="history-message" open>
        <summary>
          <span class="history-role">assistant output</span>
          <span v-if="messageContent(assistantMessage)" class="history-preview">{{ messageContent(assistantMessage).slice(0, 90) }}</span>
        </summary>
        <div v-if="messageContent(assistantMessage)" class="markdown-body" v-html="renderMarkdown(messageContent(assistantMessage))"></div>
      </details>
      <details
        v-for="(call, index) in toolCallDetails"
        :key="index"
        class="history-message"
        :open="index < 2"
      >
        <summary>
          <span class="history-role">{{ toolCallTitle(call, index) }}</span>
          <span v-if="toolCallArguments(call)" class="history-preview">{{ toolCallArguments(call).slice(0, 90) }}</span>
        </summary>
        <div v-if="toolCallArguments(call)" class="markdown-body" v-html="renderMarkdown(toolCallArguments(call))"></div>
        <RuntimeJsonTree label="工具调用元数据" :value="call" />
      </details>
    </section>

    <section v-if="historyMessages.length" class="runtime-detail-section">
      <h4>模型上下文</h4>
      <details
        v-for="(message, index) in historyMessages"
        :key="index"
        class="history-message"
        :open="index < 2"
      >
        <summary>
          <span class="history-role">{{ messageTitle(message, index) }}</span>
          <span v-if="messageContent(message)" class="history-preview">{{ messageContent(message).slice(0, 90) }}</span>
        </summary>
        <div v-if="messageContent(message)" class="markdown-body" v-html="renderMarkdown(messageContent(message))"></div>
        <RuntimeJsonTree
          v-if="hasExtra(message)"
          label="消息元数据"
          :value="messageExtra(message)"
        />
      </details>
    </section>

    <section
      v-for="section in readableSections"
      :key="section.key"
      class="runtime-detail-section"
    >
      <h4>{{ section.title }}</h4>
      <div
        v-if="section.mode === 'markdown'"
        class="markdown-body"
        v-html="renderMarkdown(section.value)"
      ></div>
      <RuntimeJsonTree
        v-else
        :label="section.key"
        :value="section.value"
        :default-open="true"
      />
    </section>

    <section class="runtime-detail-section metadata-section">
      <h4>原始元数据</h4>
      <RuntimeJsonTree label="metadata" :value="metadata" />
    </section>
  </div>
</template>

<style scoped>
.runtime-event-detail {
  border-top: 1px solid var(--divider);
  background: var(--surface);
}
.runtime-detail-state {
  padding: 8px 10px;
  color: var(--text-muted);
  background: var(--surface-muted);
  font-size: 10px;
}
.runtime-detail-state.error {
  color: var(--danger);
  background: var(--danger-soft);
}
.runtime-detail-section {
  border-top: 1px solid var(--divider);
}
.runtime-detail-section:first-child {
  border-top: 0;
}
.runtime-detail-section h4 {
  margin: 0;
  padding: 9px 10px;
  color: var(--text-secondary);
  background: var(--surface-muted);
  font-size: 10px;
  font-weight: 800;
}
.history-message {
  border-top: 1px solid var(--divider);
}
.history-message summary {
  min-height: 34px;
  padding: 7px 10px;
  display: flex;
  align-items: center;
  gap: 10px;
  color: var(--text-secondary);
  cursor: pointer;
  font-size: 10px;
}
.history-role {
  flex: 0 0 auto;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  color: var(--info);
}
.history-preview {
  min-width: 0;
  overflow: hidden;
  color: var(--text-muted);
  text-overflow: ellipsis;
  white-space: nowrap;
}
.markdown-body {
  max-height: 520px;
  padding: 10px;
  overflow: auto;
  color: var(--text-secondary);
  font-size: 11px;
  line-height: 1.65;
  overflow-wrap: anywhere;
}
.markdown-body :deep(p) {
  margin: 0 0 7px;
  white-space: pre-wrap;
}
.markdown-body :deep(h3),
.markdown-body :deep(h4),
.markdown-body :deep(h5),
.markdown-body :deep(h6) {
  margin: 10px 0 6px;
  color: var(--text);
  font-size: 12px;
}
.markdown-body :deep(ul),
.markdown-body :deep(ol) {
  margin: 6px 0 8px 18px;
  padding: 0;
}
.markdown-body :deep(li) {
  margin: 3px 0;
}
.markdown-body :deep(blockquote) {
  margin: 8px 0;
  padding: 6px 9px;
  border-left: 3px solid var(--info);
  color: var(--text-secondary);
  background: var(--surface-muted);
}
.markdown-body :deep(code) {
  padding: 1px 4px;
  border-radius: 3px;
  color: var(--text);
  background: var(--surface-muted);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
}
.markdown-body :deep(pre) {
  margin: 8px 0;
  padding: 9px;
  overflow: auto;
  border: 1px solid var(--divider);
  border-radius: 4px;
  background: var(--surface-muted);
  white-space: pre-wrap;
}
.metadata-section {
  border-top: 2px solid var(--divider);
}
.model-reasoning {
  padding: 10px;
  border-top: 1px solid var(--divider);
  color: var(--text-secondary);
  font-size: 11px;
}
.model-reasoning strong {
  color: var(--text);
  font-size: 10px;
}
.model-reasoning p {
  margin: 6px 0 0;
  line-height: 1.6;
  white-space: pre-wrap;
}
.observation-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  border-top: 1px solid var(--divider);
}
.observation-row {
  min-width: 0;
  padding: 8px 10px;
  display: grid;
  gap: 4px;
  border-right: 1px solid var(--divider);
  border-bottom: 1px solid var(--divider);
}
.observation-row span,
.source-list > span {
  color: var(--text-muted);
  font-size: 9px;
  font-weight: 800;
}
.observation-row strong {
  min-width: 0;
  overflow: hidden;
  color: var(--text-secondary);
  font-size: 11px;
  font-weight: 650;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.source-list {
  padding: 9px 10px;
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  border-top: 1px solid var(--divider);
}
.source-list a {
  max-width: 100%;
  padding: 3px 6px;
  overflow: hidden;
  border: 1px solid var(--divider);
  border-radius: 3px;
  color: var(--info);
  background: var(--info-soft);
  font-size: 10px;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.compression-diff { display: grid; }
.compression-head,
.compression-row {
  min-height: 34px;
  padding: 0 10px;
  display: grid;
  grid-template-columns: 80px minmax(64px, 1fr) minmax(64px, 1fr) minmax(110px, 1.4fr);
  align-items: center;
  gap: 8px;
  border-top: 1px solid var(--divider);
  font-size: 10px;
}
.compression-head { color: var(--text-muted); background: var(--surface-muted); font-weight: 700; }
.compression-row strong { color: var(--text); }
.compression-row span { color: var(--text-secondary); font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; }
.compression-row small { color: var(--success); }
.intent-anchor,
.preserved-requirements { padding: 10px; border-top: 1px solid var(--divider); color: var(--text-secondary); font-size: 11px; }
.intent-anchor strong,
.preserved-requirements strong { color: var(--text); font-size: 10px; }
.intent-anchor p { margin: 6px 0 0; line-height: 1.6; white-space: pre-wrap; }
.preserved-requirements ul { margin: 6px 0 0 16px; padding: 0; }
.preserved-requirements li { margin: 4px 0; line-height: 1.5; }
@media (max-width: 640px) {
  .compression-head,
  .compression-row { grid-template-columns: 64px 1fr 1fr; }
  .compression-head span:last-child,
  .compression-row small { grid-column: 2 / -1; }
}
</style>
