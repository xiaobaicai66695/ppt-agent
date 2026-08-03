import { describe, expect, it } from 'vitest';
import type { ConversationSession, TaskItem } from '../types';
import { canonicalOutputFile, mergeSlideDeliveries, recoverConversationMessages, renderSafeMarkdown, summarizeTaskTitle } from './workbench';

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

  it('escapes raw html while rendering markdown blocks', () => {
    const html = renderSafeMarkdown('# 标题\n\n- **重点**\n- `<script>`\n\n<script>alert(1)</script>');
    expect(html).toContain('<h1>标题</h1>');
    expect(html).toContain('<ul>');
    expect(html).toContain('&lt;script&gt;alert(1)&lt;/script&gt;');
    expect(html).not.toContain('<script>');
  });
});
