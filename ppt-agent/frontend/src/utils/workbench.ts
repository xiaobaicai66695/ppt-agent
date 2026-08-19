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
  const timestamp = session.updated_at || session.created_at || new Date(0).toISOString();
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
  const seen = new Set(merged.map(message => `${message.role}\u0000${message.content.trim()}`));
  for (const message of incoming) {
    const content = message.content?.trim();
    if (!content) continue;
    const key = `${message.role}\u0000${content}`;
    if (seen.has(key)) continue;
    seen.add(key);
    merged.push({ ...message, content });
  }
  return merged;
}

export function runtimeAssistantOutputMessages(events: RuntimeEvent[]): ConversationMessage[] {
  const messages: ConversationMessage[] = [];
  const seen = new Set<string>();
  for (const event of [...events].sort((a, b) => {
    if (a.id > 0 && b.id > 0 && a.id !== b.id) return a.id - b.id;
    return Date.parse(a.timestamp || '') - Date.parse(b.timestamp || '');
  })) {
    const kind = (event.kind || '').toLowerCase();
    if (!kind.includes('llm') || kind.includes('start')) continue;
    const output = event.metadata?.assistant_output;
    if (typeof output !== 'string') continue;
    const content = output.trim();
    if (!content || seen.has(content)) continue;
    seen.add(content);
    messages.push({
      role: 'assistant',
      content,
      timestamp: event.timestamp || new Date(0).toISOString(),
    });
  }
  return messages;
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
  if (name === 'search') return 'search';
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
