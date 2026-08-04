import { describe, expect, it } from 'vitest';
import type { ConversationMessage, ConversationSession, TaskItem } from '../types';
import {
  canonicalOutputFile, deriveLiveActivity, mergeConversationMessages, mergeSlideDeliveries,
  nextReplayCursor, recoverConversationMessages, renderSafeMarkdown, summarizeTaskTitle,
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

  it('escapes raw html while rendering markdown blocks', () => {
    const html = renderSafeMarkdown('# 标题\n\n- **重点**\n- `<script>`\n\n<script>alert(1)</script>');
    expect(html).toContain('<h1>标题</h1>');
    expect(html).toContain('<ul>');
    expect(html).toContain('&lt;script&gt;alert(1)&lt;/script&gt;');
    expect(html).not.toContain('<script>');
  });
});
