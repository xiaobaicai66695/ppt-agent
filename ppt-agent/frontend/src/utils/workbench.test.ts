import { describe, expect, it } from 'vitest';
import type { ConversationMessage, ConversationSession, RuntimeEvent, TaskItem } from '../types';
import {
  canonicalOutputFile, compactRuntimeEvents, deriveLiveActivity, mergeConversationMessages,
  mergeRuntimeEvents, mergeRuntimeMeta, mergeSlideDeliveries, nextReplayCursor, recoverConversationMessages, renderSafeMarkdown,
  runtimeEventDetailLabel, runtimeEventKindLabel, runtimeEventNameLabel, runtimeEventStatusLabel,
  summarizeTaskTitle,
} from './workbench';

describe('workbench utilities', () => {
  it('deduplicates relative and absolute slide paths by page identity', () => {
    const items: TaskItem[] = [{
      task_id: 'slide-1', page_index: 1, title: '封面', content_type: 'title_slide',
      output_file: '1_cover.pptx', status: 'done',
    }];
    const result = mergeSlideDeliveries(items, ['/srv/output/1_cover.pptx', 'C:\\task\\1_cover.pptx']);
    expect(result).toHaveLength(1);
    expect(result[0].fileReady).toBe(true);
    expect(result[0].task.output_file).toBe('1_cover.pptx');
    expect(canonicalOutputFile('/srv/output/1_cover.pptx')).toBe('1_cover.pptx');
  });

  it('uses the unique ready filename when a legacy artifact drifted from the manifest', () => {
    const items: TaskItem[] = [{
      task_id: 'slide-6', page_index: 6, title: '2025年经济社会发展成就', content_type: 'content_slide',
      output_file: '6_2025年经济社会发展成就.pptx', status: 'done',
    }];
    const result = mergeSlideDeliveries(items, ['/srv/output/6_2025 年经济社会发展成就.pptx']);
    expect(result).toHaveLength(1);
    expect(result[0].fileReady).toBe(true);
    expect(result[0].task.output_file).toBe('6_2025 年经济社会发展成就.pptx');
  });

  it('does not guess among multiple ready filenames for the same page', () => {
    const items: TaskItem[] = [{
      task_id: 'slide-9', page_index: 9, title: '2026年主要预期目标', content_type: 'content_slide',
      output_file: '9_2026年主要预期目标.pptx', status: 'done',
    }];
    const result = mergeSlideDeliveries(items, ['9_候选一.pptx', '9_候选二.pptx']);
    expect(result).toHaveLength(1);
    expect(result[0].fileReady).toBe(false);
    expect(result[0].task.output_file).toBe('9_2026年主要预期目标.pptx');
  });

  it('creates a compact deterministic task title', () => {
    const query = '# PPT Agent 项目汇报完整文字模板（共12页，技术实习汇报风格）\n## 基础全局设定';
    expect(summarizeTaskTitle(query, 16)).toBe('PPT Agent 项目汇报完整...');
  });

  it('recovers one markdown-preserving legacy assistant turn', () => {
    const session: ConversationSession = {
      task_id: 't1', messages: [], full_answer: '# 完成\n\n- 第一页\n- 第二页',
      created_at: '2026-01-01T00:00:00Z', updated_at: '2026-01-01T00:00:01Z',
    };
    expect(recoverConversationMessages(session)).toEqual([{
      role: 'assistant', content: '# 完成\n\n- 第一页\n- 第二页', timestamp: session.updated_at,
    }]);
  });

  it('resumes from the newest task-specific event boundary', () => {
    expect(nextReplayCursor(18, 12)).toBe(18);
    expect(nextReplayCursor(0, 12)).toBe(12);
  });

  it('merges restored conversation without duplicating an assistant turn', () => {
    const existing: ConversationMessage[] = [{ role: 'assistant', content: '已完成规划', timestamp: 'a' }];
    const incoming: ConversationMessage[] = [
      { role: 'assistant', content: '已完成规划', timestamp: 'b' },
      { role: 'user', content: '修改第二页', timestamp: 'c' },
    ];
    expect(mergeConversationMessages(existing, incoming)).toEqual([
      existing[0], incoming[1],
    ]);
  });

  it('derives concise live activity without exposing tool arguments', () => {
    expect(deriveLiveActivity({ status: 'running', lastTool: 'search', done: 3, total: 12 })).toEqual({
      label: '正在检索并核实资料', detail: '已完成 3/12 页', state: 'running',
    });
    expect(deriveLiveActivity({ status: 'running', connectionInterrupted: true }).label).toBe('正在恢复实时连接');
  });

  it('collapses repeated legacy manifest polling events but keeps progress changes', () => {
    const event = (id: number, detail = ''): RuntimeEvent => ({
      id, timestamp: '2026-08-05T00:00:00Z', elapsed_ms: id * 1000,
      kind: 'manifest_validated', name: 'tasks.json', phase: 'preparing', status: 'warning', detail,
    });
    const result = compactRuntimeEvents([
      event(4, '已完成 1/2 页'), event(3), event(2), event(1),
    ]);
    expect(result.map(item => item.id)).toEqual([4, 3]);
  });

  it('uses readable labels for manifest delivery events', () => {
    const event: RuntimeEvent = {
      id: 1, timestamp: '2026-08-05T00:00:00Z', elapsed_ms: 1000,
      kind: 'manifest_validated', name: 'tasks.json', status: 'running',
    };
    expect(runtimeEventKindLabel(event)).toBe('交付进度核对');
    expect(runtimeEventNameLabel(event)).toBe('PPT 页清单');
    expect(runtimeEventStatusLabel(event.status)).toBe('进行中');
    expect(runtimeEventDetailLabel(event)).toBe('PPT 页清单状态已核对');
  });

  it('merges persisted tool history with the bounded SSE tail by event id', () => {
	const event = (id: number, name: string, metadata?: Record<string, unknown>): RuntimeEvent => ({
	  id, task_id: 'task-1', timestamp: `2026-08-05T00:00:0${id}Z`, elapsed_ms: id * 1000,
	  kind: 'tool_end', name, status: 'ok', metadata,
	});
	const persisted = [event(1, 'read_file'), event(2, 'edit_file')];
	const tail = [event(2, 'edit_file', { result: 'updated' }), event(3, 'python3')];
	const merged = mergeRuntimeEvents(persisted, tail);
	expect(merged.map(item => item.name)).toEqual(['read_file', 'edit_file', 'python3']);
	expect(merged[1].metadata).toEqual({ result: 'updated' });

	const meta = mergeRuntimeMeta({ elapsed_ms: 1, recent_events: persisted }, { elapsed_ms: 3, recent_events: tail });
	expect(meta.elapsed_ms).toBe(3);
	expect(meta.recent_events).toHaveLength(3);
  });

  it('labels compression as an observable before-after event', () => {
	const event: RuntimeEvent = {
	  id: 9, timestamp: '2026-08-05T00:00:09Z', elapsed_ms: 9000,
	  kind: 'compression', name: 'context_compressor', status: 'ok',
	  metadata: { before_tokens: 42000, after_tokens: 11000, saved_pct: '73.8%' },
	};
	expect(runtimeEventKindLabel(event)).toBe('上下文压缩');
	expect(runtimeEventNameLabel(event)).toBe('对话上下文');
	expect(runtimeEventDetailLabel(event)).toContain('42,000 → 11,000');
  });

  it('escapes raw html while rendering markdown blocks', () => {
    const html = renderSafeMarkdown('# 标题\n\n- **重点**\n- `<script>`\n\n<script>alert(1)</script>');
    expect(html).toContain('<h1>标题</h1>');
    expect(html).toContain('<ul>');
    expect(html).toContain('&lt;script&gt;alert(1)&lt;/script&gt;');
    expect(html).not.toContain('<script>');
  });
});
