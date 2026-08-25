import type {
  ConversationMessage, ConversationSession, LiveActivity, RuntimeEvent, RuntimeMeta, TaskItem, TaskStatus,
} from '../types';

export interface SlideDeliveryEntry {
  key: string;
  task: TaskItem;
  fileReady: boolean;
}

export function canonicalOutputFile(file: string): string {
  return (file || '').split(/[/\\]/).pop()?.trim() || '';
}

export function slideIdentity(item: Pick<TaskItem, 'page_index' | 'task_id' | 'output_file'>): string {
  if (Number.isInteger(item.page_index) && item.page_index > 0) return `page:${item.page_index}`;
  const file = canonicalOutputFile(item.output_file).toLocaleLowerCase();
  if (file) return `file:${file}`;
  return `task:${item.task_id}`;
}

function pageIndexFromFile(file: string, fallback: number): number {
  const match = canonicalOutputFile(file).match(/^(\d+)[_-]/);
  return match ? Number.parseInt(match[1], 10) : fallback;
}

export function mergeSlideDeliveries(taskItems: TaskItem[], files: string[]): SlideDeliveryEntry[] {
  const readyFileMap = new Map<string, string>();
  for (const file of files.map(canonicalOutputFile).filter(Boolean)) {
    readyFileMap.set(file.toLocaleLowerCase(), file);
  }
  const readyFiles = [...readyFileMap.values()];
  const readyFilesByPage = new Map<number, string[]>();
  for (const file of readyFiles) {
    const pageIndex = pageIndexFromFile(file, 0);
    if (pageIndex <= 0) continue;
    readyFilesByPage.set(pageIndex, [...(readyFilesByPage.get(pageIndex) || []), file]);
  }
  const entries = new Map<string, SlideDeliveryEntry>();
  const consumedFiles = new Set<string>();

  for (const item of taskItems) {
    const canonicalFile = canonicalOutputFile(item.output_file);
    const exactFile = readyFileMap.get(canonicalFile.toLocaleLowerCase());
    const pageCandidates = readyFilesByPage.get(item.page_index) || [];
    const actualFile = exactFile || (pageCandidates.length === 1 ? pageCandidates[0] : '');
    const normalized: TaskItem = { ...item, output_file: actualFile || canonicalFile };
    const key = slideIdentity(normalized);
    const fileReady = actualFile !== '';
    if (actualFile) consumedFiles.add(actualFile.toLocaleLowerCase());
    const existing = entries.get(key);
    if (!existing || (!existing.task.output_file && canonicalFile)) {
      entries.set(key, { key, task: normalized, fileReady });
    } else if (fileReady) {
      existing.fileReady = true;
    }
  }

  for (const file of readyFiles) {
    if (consumedFiles.has(file.toLocaleLowerCase())) continue;
    const pageIndex = pageIndexFromFile(file, entries.size + 1);
    const fallback: TaskItem = {
      task_id: `file:${file}`,
      page_index: pageIndex,
      title: canonicalOutputFile(file).replace(/\.pptx$/i, ''),
      content_type: '',
      output_file: canonicalOutputFile(file),
      status: 'done',
    };
    const key = slideIdentity(fallback);
    const existing = entries.get(key);
    if (existing) {
      continue;
    } else {
      entries.set(key, { key, task: fallback, fileReady: true });
    }
  }

  return [...entries.values()].sort((a, b) => {
    const pageDelta = a.task.page_index - b.task.page_index;
    return pageDelta || a.key.localeCompare(b.key);
  });
}

export function summarizeTaskTitle(query: string, limit = 42): string {
  const lines = (query || '').replace(/\r/g, '').split('\n').map(line => line.trim()).filter(Boolean);
  const heading = lines.find(line => /^#{1,6}\s+/.test(line));
  let candidate = heading || lines[0] || '未命名任务';
  candidate = candidate
    .replace(/^#{1,6}\s+/, '')
    .replace(/^[-*>\s]+/, '')
    .replace(/[`*_~]/g, '')
    .replace(/\s+/g, ' ')
    .trim();
  const sentence = candidate.split(/(?<=[。！？!?])\s*/)[0] || candidate;
  const chars = [...sentence];
  return chars.length > limit ? `${chars.slice(0, limit).join('')}...` : sentence;
}

function legacyConversationMessages(content: string, timestamp: string): ConversationMessage[] {
  const marker = /^\*\*(用户|助手)\*\*[:：]?\s*(.*)$/;
  const messages: ConversationMessage[] = [];
  let active: ConversationMessage | null = null;

  for (const rawLine of content.replace(/\r/g, '').split('\n')) {
    const match = rawLine.trim().match(marker);
    if (match) {
      if (active?.content.trim()) messages.push({ ...active, content: active.content.trim() });
      active = {
        role: match[1] === '用户' ? 'user' : 'assistant',
        content: match[2] || '',
        timestamp,
      };
    } else if (active) {
      active.content += `${active.content ? '\n' : ''}${rawLine}`;
    }
  }
  if (active?.content.trim()) messages.push({ ...active, content: active.content.trim() });
  return messages;
}

export function recoverConversationMessages(session: ConversationSession): ConversationMessage[] {
  const structured = (session.messages || []).filter(message => message.content?.trim());
  if (structured.length > 0) return structured.map(message => ({ ...message }));
  const timestamp = session.created_at || new Date(0).toISOString();
  if (session.full_answer?.trim()) {
    return [{ role: 'assistant', content: session.full_answer.trim(), timestamp }];
  }
  if (session.conversation_content?.trim()) {
    const parsed = legacyConversationMessages(session.conversation_content, timestamp);
    return parsed.length > 0
      ? parsed
      : [{ role: 'assistant', content: session.conversation_content.trim(), timestamp }];
  }
  return [];
}

export function mergeConversationMessages(
  current: ConversationMessage[], incoming: ConversationMessage[],
): ConversationMessage[] {
  const merged = current.map(message => ({ ...message }));
  for (const message of incoming) {
    const content = message.content?.trim();
    if (!content) continue;
    mergeConversationMessage(merged, { ...message, content });
  }
  return merged;
}

export function runtimeAssistantOutputMessages(events: RuntimeEvent[]): ConversationMessage[] {
  const messages: ConversationMessage[] = [];
  const seen = new Set<string>();
  let toolEventSinceAssistant = false;
  for (const event of [...events].sort((a, b) => {
    if (a.id > 0 && b.id > 0 && a.id !== b.id) return a.id - b.id;
    return Date.parse(a.timestamp || '') - Date.parse(b.timestamp || '');
  })) {
    if (isToolLikeRuntimeEvent(event)) {
      toolEventSinceAssistant = true;
      continue;
    }
    const kind = (event.kind || '').toLowerCase();
    if (!kind.includes('llm') || kind.includes('start')) continue;
    const output = event.metadata?.assistant_output;
    if (typeof output !== 'string') continue;
    const content = output.trim();
    if (!content || seen.has(content)) continue;
    seen.add(content);
    mergeConversationMessage(messages, {
      role: 'assistant',
      content,
      timestamp: event.timestamp || new Date(0).toISOString(),
    }, { splitCumulativePrefix: toolEventSinceAssistant });
    toolEventSinceAssistant = false;
  }
  return messages;
}

function mergeConversationMessage(
  messages: ConversationMessage[],
  incoming: ConversationMessage,
  options: { splitCumulativePrefix?: boolean } = { splitCumulativePrefix: true },
) {
  let content = incoming.content.trim();
  if (!content) return;
  let incomingNormalized = normalizeConversationContent(content);
  for (let index = 0; index < messages.length; index += 1) {
    const existing = messages[index];
    if (existing.role !== incoming.role) continue;
    const existingNormalized = normalizeConversationContent(existing.content);
    const relation = duplicateContentRelation(existingNormalized, incomingNormalized);
    if (relation === 'same' || relation === 'existing_contains_incoming') return;
    if (relation === 'incoming_contains_existing') {
      const suffix = options.splitCumulativePrefix ? cumulativeConversationSuffix(existing.content, content) : null;
      if (suffix !== null) {
        content = suffix;
        if (!content) return;
        incomingNormalized = normalizeConversationContent(content);
        continue;
      }
      messages[index] = { ...incoming, content };
      return;
    }
  }
  messages.push({ ...incoming, content });
}

function normalizeConversationContent(content: string): string {
  return content
    .replace(/\r/g, '')
    .replace(/\s+/g, '')
    .toLowerCase();
}

function duplicateContentRelation(
  existing: string,
  incoming: string,
): 'none' | 'same' | 'existing_contains_incoming' | 'incoming_contains_existing' {
  if (!existing || !incoming) return 'none';
  if (existing === incoming) return 'same';
  const minLength = Math.min(existing.length, incoming.length);
  const maxLength = Math.max(existing.length, incoming.length);
  if (minLength < 20 || minLength / maxLength < 0.18) return 'none';
  if (existing.includes(incoming)) return 'existing_contains_incoming';
  if (incoming.includes(existing)) return 'incoming_contains_existing';
  return 'none';
}

function cumulativeConversationSuffix(existing: string, incoming: string): string | null {
  const existingTrimmed = existing.trim();
  const incomingTrimmed = incoming.trim();
  if (!existingTrimmed || !incomingTrimmed) return null;
  if (!incomingTrimmed.startsWith(existingTrimmed)) return null;
  return incomingTrimmed.slice(existingTrimmed.length).replace(/^\s+/, '').trim();
}

export function nextReplayCursor(cachedEventID = 0, sessionBoundary = 0): number {
  return Math.max(0, cachedEventID, sessionBoundary);
}

export interface LiveActivityInput {
  status?: TaskStatus;
  phase?: string;
  phaseDetail?: string;
  lastTool?: string;
  connectionInterrupted?: boolean;
  done?: number;
  total?: number;
  error?: string;
}

export interface ObservableRuntimeStep {
  id: number;
  label: string;
  detail: string;
  status: string;
  elapsed_ms: number;
  category: string;
  urls: string[];
}

export interface InlineToolImagePreview {
  id: string;
  preview_url?: string;
  image_url?: string;
  source_url?: string;
  photographer?: string;
  attribution?: string;
  local_path?: string;
  description?: string;
  download_error?: string;
}

export interface InlineToolSearchResult {
  title: string;
  url: string;
  description?: string;
  source?: string;
  date?: string;
}

export interface InlineToolPreview {
  key: string;
  name: string;
  label: string;
  detail: string;
  status: string;
  timestamp: string;
  elapsed_ms: number;
  start_event_id?: number;
  end_event_id?: number;
  args_preview?: string;
  result_preview?: string;
  source_urls: string[];
  search_results: InlineToolSearchResult[];
  image_results: InlineToolImagePreview[];
  metadata_loaded?: boolean;
}

export interface InlineToolGroupPreview {
  key: string;
  timestamp: string;
  status: string;
  label: string;
  detail: string;
  tools: InlineToolPreview[];
}

export interface ToolPreviewField {
  label: string;
  value: string;
}

export type InlineConversationItem =
  | { type: 'message'; key: string; timestamp: string; message: ConversationMessage }
  | { type: 'tool_group'; key: string; timestamp: string; group: InlineToolGroupPreview };

export function deriveInlineConversationItems(
  messages: ConversationMessage[] = [],
  events: RuntimeEvent[] = [],
): InlineConversationItem[] {
  const items: InlineConversationItem[] = [];
  messages.forEach((message, index) => {
    items.push({
      type: 'message',
      key: `message:${message.timestamp || index}:${message.role}:${index}`,
      timestamp: message.timestamp || new Date(0).toISOString(),
      message,
    });
  });
  const toolItems = deriveInlineToolPreviews(events).map(tool => ({
    type: 'tool' as const,
    key: tool.key,
    timestamp: tool.timestamp,
    tool,
  }));
  toolItems.forEach(tool => {
    items.push({
      type: 'tool_group',
      key: `pending:${tool.key}`,
      timestamp: tool.timestamp,
      group: {
        key: `pending:${tool.key}`,
        timestamp: tool.timestamp,
        status: tool.tool.status,
        label: '工具调用',
        detail: tool.tool.label,
        tools: [tool.tool],
      },
    });
  });
  const sorted = items.sort((a, b) => {
    const timeDelta = Date.parse(a.timestamp || '') - Date.parse(b.timestamp || '');
    if (timeDelta !== 0) return timeDelta;
    if (a.type !== b.type) return a.type === 'tool_group' ? -1 : 1;
    return a.key.localeCompare(b.key);
  });
  return groupAdjacentInlineTools(sorted);
}

export function deriveInlineToolPreviews(events: RuntimeEvent[] = []): InlineToolPreview[] {
  const sorted = [...events].sort((a, b) => {
    if (a.id > 0 && b.id > 0 && a.id !== b.id) return a.id - b.id;
    return Date.parse(a.timestamp || '') - Date.parse(b.timestamp || '');
  });
  const previews: InlineToolPreview[] = [];
  const running = new Map<string, InlineToolPreview[]>();

  for (const event of sorted) {
    if (!isToolLikeRuntimeEvent(event)) continue;
    const kind = (event.kind || '').toLowerCase();
    const name = event.name || event.phase || 'tool';
    if (kind.endsWith('_start')) {
      const preview = toolPreviewFromEvent(event, 'running');
      previews.push(preview);
      running.set(name, [...(running.get(name) || []), preview]);
      continue;
    }
    if (kind.endsWith('_end') || kind.endsWith('_error')) {
      const queue = running.get(name) || [];
      const existing = queue.shift();
      if (queue.length) running.set(name, queue);
      else running.delete(name);
      const status = kind.endsWith('_error') || event.status === 'error' || event.status === 'failed' ? 'error' : 'ok';
      if (existing) {
        mergeToolPreview(existing, event, status);
      } else {
        previews.push(toolPreviewFromEvent(event, status));
      }
      continue;
    }
    previews.push(toolPreviewFromEvent(event, event.status || 'ok'));
  }
  return previews;
}

function isToolLikeRuntimeEvent(event: RuntimeEvent): boolean {
  const kind = (event.kind || '').toLowerCase();
  return kind.startsWith('tool_') || kind.startsWith('slide_render_');
}

function toolPreviewFromEvent(event: RuntimeEvent, status: string): InlineToolPreview {
  const metadata = event.metadata || {};
  const resultPayload = metadataResultRecord(metadata);
  const imageResults = metadataImageResults(metadata.image_results).length
    ? metadataImageResults(metadata.image_results)
    : metadataImageResults(resultPayload?.photos);
  const searchResults = metadataSearchResults(metadata.search_results).length
    ? metadataSearchResults(metadata.search_results)
    : metadataSearchResults(resultPayload?.results);
  return {
    key: `tool:${event.id || runtimeEventFingerprint(event)}`,
    name: event.name || event.phase || 'tool',
    label: runtimeEventNameLabel(event),
    detail: runtimeEventDetailLabel(event),
    status,
    timestamp: event.timestamp || new Date(0).toISOString(),
    elapsed_ms: event.elapsed_ms || 0,
    start_event_id: event.kind?.toLowerCase().endsWith('_start') ? event.id : undefined,
    end_event_id: event.kind?.toLowerCase().endsWith('_start') ? undefined : event.id,
    args_preview: firstMetadataString(metadata, 'args', 'arguments', 'args_preview', 'arguments_preview')
      || synthesizeToolArgsPreview(event),
    result_preview: preferredToolResultPreview(event, imageResults, searchResults, resultPayload),
    source_urls: metadataArray(metadata.source_urls),
    search_results: searchResults,
    image_results: imageResults,
    metadata_loaded: event.metadata_loaded,
  };
}

function mergeToolPreview(preview: InlineToolPreview, event: RuntimeEvent, status: string) {
  const metadata = event.metadata || {};
  preview.status = status;
  preview.end_event_id = event.id;
  preview.elapsed_ms = event.elapsed_ms || preview.elapsed_ms;
  preview.detail = runtimeEventDetailLabel(event);
  const resultPayload = metadataResultRecord(metadata);
  const images = metadataImageResults(metadata.image_results).length
    ? metadataImageResults(metadata.image_results)
    : metadataImageResults(resultPayload?.photos);
  const searchResults = metadataSearchResults(metadata.search_results).length
    ? metadataSearchResults(metadata.search_results)
    : metadataSearchResults(resultPayload?.results);
  preview.result_preview = preferredToolResultPreview(event, images, searchResults, resultPayload)
    || preview.result_preview;
  preview.source_urls = metadataArray(metadata.source_urls) || preview.source_urls;
  if (images.length > 0) preview.image_results = images;
  if (searchResults.length > 0) preview.search_results = searchResults;
  preview.metadata_loaded = event.metadata_loaded ?? preview.metadata_loaded;
}

function preferredToolResultPreview(
  event: RuntimeEvent,
  images: InlineToolImagePreview[],
  searchResults: InlineToolSearchResult[],
  resultPayload?: Record<string, unknown>,
): string {
  const name = event.name || '';
  const synthesized = synthesizeToolResultPreview(event, images, searchResults, resultPayload);
  if ((name === 'search' || name === 'search_images') && synthesized) return synthesized;
  return firstMetadataString(event.metadata || {}, 'result', 'result_preview', 'output_preview', 'error') || synthesized;
}

function groupAdjacentInlineTools(items: InlineConversationItem[]): InlineConversationItem[] {
  const grouped: InlineConversationItem[] = [];
  let pending: InlineToolPreview[] = [];
  let pendingTimestamp = '';

  const flush = () => {
    if (pending.length === 0) return;
    const group = inlineToolGroupFromTools(pending, pendingTimestamp);
    grouped.push({
      type: 'tool_group',
      key: group.key,
      timestamp: group.timestamp,
      group,
    });
    pending = [];
    pendingTimestamp = '';
  };

  for (const item of items) {
    if (item.type === 'tool_group') {
      if (!pendingTimestamp) pendingTimestamp = item.timestamp;
      pending.push(...item.group.tools);
      continue;
    }
    flush();
    grouped.push(item);
  }
  flush();
  return grouped;
}

function inlineToolGroupFromTools(tools: InlineToolPreview[], timestamp: string): InlineToolGroupPreview {
  const status = tools.some(tool => tool.status === 'running')
    ? 'running'
    : tools.some(tool => tool.status === 'error' || tool.status === 'failed')
      ? 'error'
      : 'ok';
  const names = [...new Set(tools.map(tool => tool.label).filter(Boolean))].slice(0, 4);
  return {
    key: `tool-group:${tools.map(tool => tool.key).join(':')}`,
    timestamp: timestamp || tools[0]?.timestamp || new Date(0).toISOString(),
    status,
    label: status === 'running' ? '正在调用工具' : '本轮工具调用',
    detail: `${tools.length} 次调用${names.length ? `：${names.join('、')}` : ''}`,
    tools,
  };
}

function synthesizeToolArgsPreview(event: RuntimeEvent): string {
  const metadata = event.metadata || {};
  const name = event.name || '';
  const data: Record<string, unknown> = {};
  if (name === 'search') {
    data.query = metadata.search_query;
  } else if (name === 'search_images') {
    data.query = metadata.image_query;
    data.asset_purpose = metadata.asset_purpose;
    data.asset_subject = metadata.asset_subject;
    data.composition = metadata.composition;
  } else if (name === 'read_file') {
    data.file = metadata.file_path;
  } else if (name === 'update_tasks_manifest') {
    data.mode = metadata.mode;
    data.slides = metadata.slide_count;
    data.template = metadata.template;
    data.theme = metadata.theme;
  }
  return compactJSONPreview(data);
}

function synthesizeToolResultPreview(
  event: RuntimeEvent,
  images: InlineToolImagePreview[],
  searchResults: InlineToolSearchResult[],
  resultPayload?: Record<string, unknown>,
): string {
  const metadata = event.metadata || {};
  const name = event.name || '';
  const data: Record<string, unknown> = {};
  if (event.status === 'error' || event.status === 'failed') {
    data.error = event.detail || metadata.error;
  }
  if (name === 'search') {
    const urls = metadataArray(metadata.source_urls);
    const sourceCount = Number(metadata.source_count || 0) || searchResults.length || urls.length;
    data.query = metadata.search_query;
    if (sourceCount > 0) data.sources = sourceCount;
    data.top_sources = searchResults.map(item => item.title).filter(Boolean).slice(0, 3);
    if (sourceCount === 0) data.note = metadata.error || '本次检索没有匹配来源，可调整关键词后重试';
  } else if (name === 'search_images') {
    data.query = metadata.image_query || metadata.asset_query;
    data.provider = metadata.provider || resultPayload?.provider;
    data.total = metadata.total || resultPayload?.total;
    if (images.length > 0) data.images = images.length;
    data.downloaded = images.filter(image => image.local_path).length;
    data.attribution = images.map(image => image.attribution || image.photographer).filter(Boolean).slice(0, 3);
    if (images.length === 0) data.note = metadata.error || '本次检索没有匹配图片，可调整视觉主体后重试';
  } else if (name === 'read_file') {
    data.status = runtimeEventStatusLabel(event.status);
    data.file = metadata.file_path;
  } else if (name === 'update_tasks_manifest') {
    data.status = runtimeEventStatusLabel(event.status);
    data.slides = metadata.slide_count;
    data.target = metadata.target;
  }
  return compactJSONPreview(data);
}

function compactJSONPreview(data: Record<string, unknown>): string {
  const entries = Object.entries(data).filter(([, value]) => {
    if (value === undefined || value === null || value === '') return false;
    return !(Array.isArray(value) && value.length === 0);
  });
  if (entries.length === 0) return '';
  return JSON.stringify(Object.fromEntries(entries));
}

export function formatToolPreviewFields(tool: InlineToolPreview, part: 'args' | 'result'): ToolPreviewField[] {
  const raw = part === 'args' ? tool.args_preview : tool.result_preview;
  const parsed = parsePreviewPayload(raw || '');
  const fallbackLabel = part === 'args' ? '输入' : '结果';
  if (parsed === undefined) return [];
  if (typeof parsed === 'string') return [{ label: fallbackLabel, value: parsed }];
  if (Array.isArray(parsed)) return [{ label: fallbackLabel, value: summarizePreviewValue(parsed) }];
  return previewFieldsFromObject(parsed as Record<string, unknown>);
}

function parsePreviewPayload(raw: string): unknown {
  const value = raw.trim();
  if (!value) return undefined;
  if ((value.startsWith('{') && value.endsWith('}')) || (value.startsWith('[') && value.endsWith(']'))) {
    try {
      return JSON.parse(value);
    } catch {
      return value;
    }
  }
  return value;
}

function previewFieldsFromObject(record: Record<string, unknown>): ToolPreviewField[] {
  return Object.entries(record)
    .filter(([, value]) => value !== undefined && value !== null && value !== '')
    .slice(0, 10)
    .map(([key, value]) => ({
      label: previewFieldLabel(key),
      value: summarizePreviewValue(value),
    }))
    .filter(field => field.value);
}

function summarizePreviewValue(value: unknown): string {
  if (Array.isArray(value)) {
    if (value.length === 0) return '';
    const scalarItems = value.filter(item => ['string', 'number', 'boolean'].includes(typeof item)).slice(0, 3);
    if (scalarItems.length > 0) {
      const suffix = value.length > scalarItems.length ? ` 等 ${value.length} 项` : '';
      return `${scalarItems.map(item => String(item)).join('、')}${suffix}`;
    }
    return `${value.length} 项`;
  }
  if (value !== null && typeof value === 'object') {
    const keys = Object.keys(value as Record<string, unknown>);
    return keys.length > 0 ? `${keys.length} 个字段：${keys.slice(0, 4).map(previewFieldLabel).join('、')}` : '';
  }
  if (typeof value === 'boolean') return value ? '是' : '否';
  return truncatePreviewText(String(value).trim(), 260);
}

function previewFieldLabel(key: string): string {
  const labels: Record<string, string> = {
    query: '检索词',
    search_query: '检索词',
    image_query: '图片检索词',
    asset_purpose: '图片用途',
    asset_subject: '视觉主体',
    composition: '构图要求',
    file: '文件',
    file_path: '文件',
    mode: '写入模式',
    slides: '页数',
    slide_count: '页数',
    template: '模板',
    theme: '配色',
    sources: '来源数量',
    source_count: '来源数量',
    top_sources: '主要来源',
    results: '搜索结果',
    content: '结果摘要',
    images: '图片数量',
    downloaded: '已下载',
    download: '下载图片',
    attribution: '署名',
    status: '状态',
    target: '目标',
    error: '错误',
    photos: '图片结果',
    total: '总数',
    total_pages: '页数',
    provider: '服务',
    asset_query: '图片检索词',
    note: '说明',
  };
  return labels[key] || key.replace(/_/g, ' ');
}

function truncatePreviewText(value: string, limit: number): string {
  const normalized = value.replace(/\s+/g, ' ').trim();
  if (normalized.length <= limit) return normalized;
  return `${normalized.slice(0, Math.max(1, limit - 1)).trim()}...`;
}

function firstMetadataString(metadata: Record<string, unknown>, ...keys: string[]): string {
  for (const key of keys) {
    const value = metadata[key];
    if (value === undefined || value === null) continue;
    const text = typeof value === 'string' ? value.trim() : JSON.stringify(value);
    if (text) return text;
  }
  return '';
}

function metadataImageResults(value: unknown): InlineToolImagePreview[] {
  if (!Array.isArray(value)) return [];
  return value
    .filter(item => item !== null && typeof item === 'object' && !Array.isArray(item))
    .slice(0, 6)
    .map((item, index) => {
      const record = item as Record<string, unknown>;
      const text = (key: string) => String(record[key] || '').trim();
      return {
        id: text('id') || String(index + 1),
        preview_url: text('preview_url'),
        image_url: text('image_url'),
        source_url: text('source_url'),
        photographer: text('photographer'),
        attribution: text('attribution'),
        local_path: text('local_path'),
        description: text('description'),
        download_error: text('download_error'),
      };
    })
    .filter(item => item.preview_url || item.image_url || item.source_url || item.local_path);
}

function metadataSearchResults(value: unknown): InlineToolSearchResult[] {
  if (!Array.isArray(value)) return [];
  return value
    .filter(item => item !== null && typeof item === 'object' && !Array.isArray(item))
    .slice(0, 5)
    .map(item => {
      const record = item as Record<string, unknown>;
      const text = (key: string) => String(record[key] || '').trim();
      return {
        title: text('title') || text('url') || '未命名来源',
        url: text('url'),
        description: text('description'),
        source: text('source'),
        date: text('date'),
      };
    })
    .filter(item => item.title || item.url || item.description);
}

function metadataResultRecord(metadata: Record<string, unknown>): Record<string, unknown> | undefined {
  const raw = metadata.result ?? metadata.result_preview;
  if (raw !== null && typeof raw === 'object' && !Array.isArray(raw)) {
    return raw as Record<string, unknown>;
  }
  if (typeof raw !== 'string' || !raw.trim()) return undefined;
  const parsed = parsePreviewPayload(raw);
  return parsed !== null && typeof parsed === 'object' && !Array.isArray(parsed)
    ? parsed as Record<string, unknown>
    : undefined;
}

export function deriveLiveActivity(input: LiveActivityInput): LiveActivity {
  if (input.connectionInterrupted) {
    return { label: '正在恢复实时连接', detail: '任务仍在服务器继续执行', state: 'running' };
  }
  if (input.status === 'failed') {
    return { label: '生成遇到错误', detail: input.error || '请查看执行轨迹', state: 'error' };
  }
  if (input.status === 'cancelled') {
    return { label: '任务已中断', state: 'idle' };
  }
  if (input.status === 'completed' || input.phase === 'complete') {
    return { label: '演示生成完成', detail: input.total ? `已完成 ${input.done || input.total}/${input.total} 页` : undefined, state: 'success' };
  }

  const toolLabels: Record<string, string> = {
    search: '正在检索并核实资料',
    search_images: '正在搜索并下载图片',
    read_file: '正在读取模板与设计规范',
    update_tasks_manifest: '正在整理页面内容',
    task: '正在并行生成幻灯片',
    python3: '正在渲染幻灯片',
    bash: '正在执行生成工具',
    batch_convert: '正在整理演示文件',
  };
  const phaseLabels: Record<string, string> = {
    preparing: '正在准备任务',
    planning: '正在规划演示内容',
    generating: '正在生成幻灯片',
    qa: '正在检查演示质量',
    fixing: '正在修正页面',
  };
  const label = input.phaseDetail
    || (input.lastTool ? toolLabels[input.lastTool] : '')
    || (input.phase ? phaseLabels[input.phase] : '')
    || '正在思考并推进任务';
  const detail = input.total && input.total > 0 ? `已完成 ${input.done || 0}/${input.total} 页` : '任务正在持续运行';
  return { label, detail, state: input.status === 'running' ? 'running' : 'idle' };
}

function runtimeEventFingerprint(event: RuntimeEvent): string {
  return [event.timestamp, event.kind, event.name, event.status, event.phase, event.detail].join('\u0000');
}

function runtimeEventIdentity(event: RuntimeEvent): string {
  if (event.id > 0) return `${event.task_id || ''}\u0000${event.id}`;
  return runtimeEventFingerprint(event);
}

export function mergeRuntimeEvents(current: RuntimeEvent[] = [], incoming: RuntimeEvent[] = []): RuntimeEvent[] {
  const merged = new Map<string, RuntimeEvent>();
  for (const event of [...current, ...incoming]) {
    const key = runtimeEventIdentity(event);
    const previous = merged.get(key);
    if (!previous) {
      merged.set(key, { ...event });
      continue;
    }
    merged.set(key, {
      ...previous,
      ...event,
      metadata: event.metadata ?? previous.metadata,
      metadata_loaded: event.metadata_loaded ?? previous.metadata_loaded,
    });
  }
  return [...merged.values()].sort((a, b) => {
    if (a.id > 0 && b.id > 0 && a.id !== b.id) return a.id - b.id;
    return Date.parse(a.timestamp || '') - Date.parse(b.timestamp || '');
  });
}

export function mergeRuntimeMeta(current: RuntimeMeta | null | undefined, incoming: RuntimeMeta): RuntimeMeta {
  return {
    ...(current || {}),
    ...incoming,
    recent_events: mergeRuntimeEvents(current?.recent_events || [], incoming.recent_events || []),
  } as RuntimeMeta;
}

export function compactRuntimeEvents(events: RuntimeEvent[]): RuntimeEvent[] {
  const compacted: RuntimeEvent[] = [];
  for (const event of events) {
    const previous = compacted[compacted.length - 1];
    if (
      (event.kind === 'manifest_validated' || event.kind === 'deck_spec_validated')
      && event.kind === previous?.kind
      && runtimeEventFingerprint(event) === runtimeEventFingerprint(previous)
    ) {
      continue;
    }
    compacted.push(event);
  }
  return compacted;
}

export function runtimeEventKindLabel(event: RuntimeEvent): string {
  const labels: Record<string, string> = {
    manifest_validated: '交付进度核对',
    deck_spec_validated: 'DeckSpec 校验',
    deck_spec_frozen: '页面计划冻结',
    deck_spec_alignment: '计划对齐检查',
    intent_classified: '意图分类',
    phase_changed: '阶段切换',
    slide_progress: '页面生成进度',
    delivery_progress: '交付进度',
    delivery_file_created: '文件已生成',
    slide_render_start: '开始渲染页面',
    slide_render_end: '页面渲染完成',
    slide_render_error: '页面渲染失败',
    task_terminal: '任务结束',
    file_created: '文件已生成',
    compression: '上下文压缩',
    planner_context_compressed: '上下文压缩',
    llm_start: '模型调用开始',
    llm_end: '模型调用完成',
    llm_error: '模型调用失败',
    tool_start: '工具调用开始',
    tool_end: '工具调用完成',
    tool_error: '工具调用失败',
  };
  return labels[event.kind] || event.kind;
}

export function runtimeEventNameLabel(event: RuntimeEvent): string {
  if ((event.kind === 'manifest_validated' || event.kind === 'deck_spec_validated') && event.name === 'tasks.json') return 'PPT 页清单';
  if (event.kind === 'compression' || event.kind === 'planner_context_compressed') return '对话上下文';
  if (event.name === 'search') return '资料搜索';
  if (event.name === 'search_images') return '图片搜索';
  if (event.name === 'read_file') return '读取文件';
  if (event.name === 'update_tasks_manifest') return '写入 DeckSpec';
  if (event.name === 'generate_slide') return '页面渲染器';
  return event.name || event.phase || '任务';
}

export function runtimeEventStatusLabel(status?: string): string {
  const labels: Record<string, string> = {
    ok: '正常',
    running: '进行中',
    warning: '需注意',
    error: '错误',
    failed: '失败',
    cancelled: '已取消',
  };
  return labels[(status || 'ok').toLowerCase()] || status || '正常';
}

export function runtimeEventDetailLabel(event: RuntimeEvent): string {
  if (event.detail) return event.detail;
  if (event.kind === 'manifest_validated' || event.kind === 'deck_spec_validated') return 'PPT 页清单状态已核对';
  if (event.kind === 'deck_spec_frozen') {
    const count = Number((event.metadata || {}).slides_count || (event.metadata || {}).slide_count || 0);
    return count > 0 ? `已冻结 ${count} 页计划` : '页面计划已冻结';
  }
  if (event.kind === 'compression' || event.kind === 'planner_context_compressed') {
    const metadata = event.metadata || {};
    const before = Number(metadata.before_tokens || 0);
    const after = Number(metadata.after_tokens || 0);
    const saved = String(metadata.saved_pct || '');
    if (before > 0 || after > 0) return `Token ${before.toLocaleString()} → ${after.toLocaleString()}${saved ? `，节省 ${saved}` : ''}`;
    return '用户要求已锚定，早期轨迹已压缩';
  }
  if (event.name === 'search') {
    const metadata = event.metadata || {};
    const query = String(metadata.search_query || '').trim();
    const urls = metadataArray(metadata.source_urls);
    if (query && event.kind === 'tool_start') return `搜索关键词：${query}`;
    if (query && urls.length > 0) return `搜索完成：${query}，${urls.length} 个来源`;
    if (query) return `搜索关键词：${query}`;
  }
  if (event.name === 'search_images') {
    const metadata = event.metadata || {};
    const query = String(metadata.image_query || '').trim();
    const images = Array.isArray(metadata.image_results) ? metadata.image_results.length : 0;
    if (query && event.kind === 'tool_start') return `图片关键词：${query}`;
    if (query && images > 0) return `图片搜索完成：${query}，${images} 张候选`;
    if (query) return `图片关键词：${query}`;
  }
  if (event.name === 'update_tasks_manifest') {
    const count = Number((event.metadata || {}).slide_count || 0);
    return count > 0 ? `正在整理 ${count} 页 DeckSpec` : '正在整理 DeckSpec';
  }
  if (event.name === 'read_file') {
    const path = String((event.metadata || {}).file_path || '').split(/[/\\]/).pop();
    return path ? `读取 ${path}` : '读取项目文件';
  }
  return '可展开查看详情';
}

export function deriveObservableSteps(events: RuntimeEvent[], limit = 6): ObservableRuntimeStep[] {
  return events
    .filter(isObservableEvent)
    .slice(0, limit)
    .map(event => {
      const metadata = event.metadata || {};
      return {
        id: event.id,
        label: observableEventLabel(event),
        detail: runtimeEventDetailLabel(event),
        status: event.status || 'ok',
        elapsed_ms: event.elapsed_ms,
        category: observableEventCategory(event),
        urls: metadataArray(metadata.source_urls),
      };
    });
}

function isObservableEvent(event: RuntimeEvent): boolean {
  const kind = (event.kind || '').toLowerCase();
  const name = (event.name || '').toLowerCase();
  return kind.includes('intent')
    || kind.includes('phase')
    || kind.includes('llm')
    || kind.includes('tool')
    || kind.includes('compress')
    || kind.includes('deck_spec')
    || kind.includes('delivery')
    || kind.includes('slide_render')
    || kind.includes('terminal')
    || name === 'search'
    || name === 'search_images'
    || name === 'update_tasks_manifest'
    || name === 'read_file'
    || name === 'generate_slide';
}

function observableEventLabel(event: RuntimeEvent): string {
  const kind = (event.kind || '').toLowerCase();
  const name = (event.name || '').toLowerCase();
  const metadata = event.metadata || {};
  if (kind === 'intent_classified') return '已完成意图分类';
  if (kind === 'llm_start') return 'Planner 正在规划';
  if (kind === 'llm_end') return 'Planner 输出完成';
  if (kind === 'llm_error') return 'Planner 调用失败';
  if (kind === 'compression' || kind === 'planner_context_compressed') return '正在压缩上下文';
  if (name === 'search') {
    const query = String(metadata.search_query || '').trim();
    if (kind === 'tool_start') return query ? `正在搜索：${query}` : '正在搜索资料';
    return query ? `搜索完成：${query}` : '搜索完成';
  }
  if (name === 'search_images') {
    const query = String(metadata.image_query || '').trim();
    if (kind === 'tool_start') return query ? `正在搜图：${query}` : '正在搜索图片';
    return query ? `图片已返回：${query}` : '图片搜索完成';
  }
  if (name === 'update_tasks_manifest') return kind === 'tool_end' ? 'DeckSpec 已写入' : '正在写入 DeckSpec';
  if (name === 'read_file') return '正在读取规范';
  if (name === 'generate_slide' || kind.startsWith('slide_render')) return runtimeEventKindLabel(event);
  if (kind === 'deck_spec_frozen') return '页面计划已冻结';
  if (kind === 'deck_spec_validated') return '正在核对交付进度';
  if (kind === 'delivery_file_created') return '检测到新文件';
  if (kind === 'task_terminal') return '任务进入终态';
  return runtimeEventKindLabel(event);
}

function observableEventCategory(event: RuntimeEvent): string {
  const kind = (event.kind || '').toLowerCase();
  const name = (event.name || '').toLowerCase();
  if (kind.includes('error') || event.status === 'error' || event.status === 'failed') return 'error';
  if (name === 'search' || name === 'search_images') return 'search';
  if (kind.includes('llm')) return 'planner';
  if (kind.includes('compress')) return 'delivery';
  if (kind.includes('slide_render') || name === 'generate_slide') return 'render';
  if (kind.includes('delivery') || kind.includes('deck_spec') || name === 'update_tasks_manifest') return 'delivery';
  return 'step';
}

function metadataArray(value: unknown): string[] {
  if (!Array.isArray(value)) return [];
  return value.map(item => String(item).trim()).filter(Boolean).slice(0, 5);
}

function escapeHtml(value: string): string {
  return value.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}

function renderInline(value: string): string {
  const code: string[] = [];
  let safe = escapeHtml(value).replace(/`([^`]+)`/g, (_match, body: string) => {
    const index = code.push(`<code>${body}</code>`) - 1;
    return `@@CODE_${index}@@`;
  });
  safe = safe
    .replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')
    .replace(/__([^_]+)__/g, '<strong>$1</strong>')
    .replace(/\*([^*]+)\*/g, '<em>$1</em>');
  return safe.replace(/@@CODE_(\d+)@@/g, (_match, index: string) => code[Number(index)] || '');
}

function isTableDivider(line: string): boolean {
  const cells = line.trim().replace(/^\||\|$/g, '').split('|');
  return cells.length > 0 && cells.every(cell => /^\s*:?-{3,}:?\s*$/.test(cell));
}

function tableCells(line: string): string[] {
  return line.trim().replace(/^\||\|$/g, '').split('|').map(cell => cell.trim());
}

export function renderSafeMarkdown(markdown: string): string {
  const lines = (markdown || '').replace(/\r/g, '').split('\n');
  const output: string[] = [];
  let index = 0;

  while (index < lines.length) {
    const line = lines[index];
    if (!line.trim()) { index += 1; continue; }

    if (/^```/.test(line.trim())) {
      const language = line.trim().slice(3).replace(/[^\w-]/g, '');
      const code: string[] = [];
      index += 1;
      while (index < lines.length && !/^```/.test(lines[index].trim())) code.push(lines[index++]);
      index += index < lines.length ? 1 : 0;
      output.push(`<pre><code${language ? ` class="language-${language}"` : ''}>${escapeHtml(code.join('\n'))}</code></pre>`);
      continue;
    }

    const heading = line.match(/^(#{1,6})\s+(.+)$/);
    if (heading) {
      const level = heading[1].length;
      output.push(`<h${level}>${renderInline(heading[2])}</h${level}>`);
      index += 1;
      continue;
    }

    if (index + 1 < lines.length && line.includes('|') && isTableDivider(lines[index + 1])) {
      const headers = tableCells(line);
      index += 2;
      const rows: string[][] = [];
      while (index < lines.length && lines[index].includes('|') && lines[index].trim()) rows.push(tableCells(lines[index++]));
      output.push(`<div class="md-table-wrap"><table><thead><tr>${headers.map(cell => `<th>${renderInline(cell)}</th>`).join('')}</tr></thead><tbody>${rows.map(row => `<tr>${headers.map((_, cellIndex) => `<td>${renderInline(row[cellIndex] || '')}</td>`).join('')}</tr>`).join('')}</tbody></table></div>`);
      continue;
    }

    const unordered = line.match(/^\s*[-+*]\s+(.+)$/);
    const ordered = line.match(/^\s*\d+[.)]\s+(.+)$/);
    if (unordered || ordered) {
      const tag = unordered ? 'ul' : 'ol';
      const items: string[] = [];
      while (index < lines.length) {
        const match = unordered
          ? lines[index].match(/^\s*[-+*]\s+(.+)$/)
          : lines[index].match(/^\s*\d+[.)]\s+(.+)$/);
        if (!match) break;
        items.push(`<li>${renderInline(match[1])}</li>`);
        index += 1;
      }
      output.push(`<${tag}>${items.join('')}</${tag}>`);
      continue;
    }

    if (/^>\s?/.test(line)) {
      const quoted: string[] = [];
      while (index < lines.length && /^>\s?/.test(lines[index])) quoted.push(lines[index++].replace(/^>\s?/, ''));
      output.push(`<blockquote>${quoted.map(renderInline).join('<br>')}</blockquote>`);
      continue;
    }

    if (/^\s*---+\s*$/.test(line)) {
      output.push('<hr>');
      index += 1;
      continue;
    }

    const paragraph = [line.trim()];
    index += 1;
    while (index < lines.length && lines[index].trim()
      && !/^(#{1,6})\s+/.test(lines[index])
      && !/^```/.test(lines[index].trim())
      && !/^\s*[-+*]\s+/.test(lines[index])
      && !/^\s*\d+[.)]\s+/.test(lines[index])
      && !/^>\s?/.test(lines[index])) {
      paragraph.push(lines[index++].trim());
    }
    output.push(`<p>${paragraph.map(renderInline).join('<br>')}</p>`);
  }

  return output.join('');
}
