import type { ConversationMessage, ConversationSession, TaskItem } from '../types';

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
  const readyFiles = new Set(files.map(canonicalOutputFile).filter(Boolean).map(file => file.toLocaleLowerCase()));
  const entries = new Map<string, SlideDeliveryEntry>();

  for (const item of taskItems) {
    const canonicalFile = canonicalOutputFile(item.output_file);
    const normalized: TaskItem = { ...item, output_file: canonicalFile };
    const key = slideIdentity(normalized);
    const fileReady = canonicalFile !== '' && readyFiles.has(canonicalFile.toLocaleLowerCase());
    const existing = entries.get(key);
    if (!existing || (!existing.task.output_file && canonicalFile)) {
      entries.set(key, { key, task: normalized, fileReady });
    } else if (fileReady) {
      existing.fileReady = true;
    }
  }

  for (const file of readyFiles) {
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
      existing.fileReady = true;
      if (!existing.task.output_file) existing.task.output_file = fallback.output_file;
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
