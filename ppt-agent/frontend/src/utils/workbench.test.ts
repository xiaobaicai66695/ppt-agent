import { describe, expect, it } from 'vitest';
import type { ConversationMessage, ConversationSession, RuntimeEvent, TaskItem } from '../types';
import {
  canonicalOutputFile, compactRuntimeEvents, deriveInlineConversationItems, deriveInlineToolPreviews,
  deriveLiveActivity, deriveObservableSteps, formatToolPreviewFields, mergeConversationMessages,
  mergeRuntimeEvents, mergeRuntimeMeta, mergeSlideDeliveries, nextReplayCursor, recoverConversationMessages, renderSafeMarkdown,
  runtimeAssistantOutputMessages, runtimeEventDetailLabel, runtimeEventKindLabel, runtimeEventNameLabel, runtimeEventStatusLabel,
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
      role: 'assistant', content: '# 完成\n\n- 第一页\n- 第二页', timestamp: session.created_at,
    }]);
  });

  it('does not pin recovered full answers after newer runtime tools', () => {
    const session: ConversationSession = {
      task_id: 't1',
      messages: [],
      full_answer: '固定的历史总结',
      created_at: '2026-01-01T00:00:00Z',
      updated_at: '2026-01-01T00:10:00Z',
    };
    const events: RuntimeEvent[] = [{
      id: 1,
      task_id: 't1',
      timestamp: '2026-01-01T00:05:00Z',
      elapsed_ms: 1000,
      kind: 'tool_end',
      name: 'search_images',
      status: 'ok',
      metadata: { image_query: 'conference hall wide landscape' },
    }];

    const items = deriveInlineConversationItems(recoverConversationMessages(session), events);

    expect(items.map(item => item.type)).toEqual(['message', 'tool_group']);
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

  it('extracts assistant_output from model events as markdown assistant messages', () => {
    const events: RuntimeEvent[] = [
      {
        id: 2,
        task_id: 'task-1',
        timestamp: '2026-08-05T00:00:02Z',
        elapsed_ms: 2000,
        kind: 'llm_end',
        name: 'planner',
        status: 'ok',
        metadata: { assistant_output: '# 规划\n\n- 正在拆分章节' },
      },
      {
        id: 1,
        task_id: 'task-1',
        timestamp: '2026-08-05T00:00:01Z',
        elapsed_ms: 1000,
        kind: 'llm_start',
        name: 'planner',
        status: 'running',
        metadata: { assistant_output: '不应展示模型输入' },
      },
    ];

    expect(runtimeAssistantOutputMessages(events)).toEqual([{
      role: 'assistant',
      content: '# 规划\n\n- 正在拆分章节',
      timestamp: '2026-08-05T00:00:02Z',
    }]);
  });

  it('does not expose tool output previews as assistant chat content', () => {
    const events: RuntimeEvent[] = [
      {
        id: 1,
        task_id: 'task-1',
        timestamp: '2026-08-05T00:00:01Z',
        elapsed_ms: 1000,
        kind: 'tool_end',
        name: 'update_tasks_manifest',
        status: 'ok',
        metadata: { assistant_output: '工具事件不应进入 AI 正文', output_preview: '{"tasks":[]}' },
      },
      {
        id: 2,
        task_id: 'task-1',
        timestamp: '2026-08-05T00:00:02Z',
        elapsed_ms: 2000,
        kind: 'llm_end',
        name: 'planner',
        status: 'ok',
        metadata: { output_preview: '摘要不应兜底展示' },
      },
      {
        id: 3,
        task_id: 'task-1',
        timestamp: '2026-08-05T00:00:03Z',
        elapsed_ms: 3000,
        kind: 'llm_end',
        name: 'planner',
        status: 'ok',
        metadata: { assistant_output: '## 可见规划\n\n继续执行。' },
      },
      {
        id: 4,
        task_id: 'task-1',
        timestamp: '2026-08-05T00:00:04Z',
        elapsed_ms: 4000,
        kind: 'llm_end',
        name: 'planner',
        status: 'ok',
        metadata: { assistant_output: '## 可见规划\n\n继续执行。' },
      },
    ];

    expect(runtimeAssistantOutputMessages(events)).toEqual([{
      role: 'assistant',
      content: '## 可见规划\n\n继续执行。',
      timestamp: '2026-08-05T00:00:03Z',
    }]);
  });

  it('derives inline tool previews with paired start/end and image results', () => {
    const events: RuntimeEvent[] = [
      {
        id: 1,
        task_id: 'task-1',
        timestamp: '2026-08-05T00:00:01Z',
        elapsed_ms: 1000,
        kind: 'tool_start',
        name: 'search_images',
        status: 'running',
        metadata: { image_query: 'aerial city skyline', args_preview: '{"query":"aerial city skyline"}' },
      },
      {
        id: 2,
        task_id: 'task-1',
        timestamp: '2026-08-05T00:00:02Z',
        elapsed_ms: 2000,
        kind: 'tool_end',
        name: 'search_images',
        status: 'ok',
        metadata: {
          image_query: 'aerial city skyline',
          result_preview: '{"photos":[...]}',
          source_urls: ['https://unsplash.com/photos/abc'],
          image_results: [{
            id: 'abc',
            preview_url: 'https://images.unsplash.com/small.jpg',
            source_url: 'https://unsplash.com/photos/abc',
            local_path: 'assets/images/abc.jpg',
            attribution: 'Photo by Demo on Unsplash',
          }],
        },
      },
    ];

    const tools = deriveInlineToolPreviews(events);

    expect(tools).toHaveLength(1);
    expect(tools[0].status).toBe('ok');
    expect(tools[0].start_event_id).toBe(1);
    expect(tools[0].end_event_id).toBe(2);
    expect(tools[0].args_preview).toContain('aerial city skyline');
    expect(tools[0].source_urls).toEqual(['https://unsplash.com/photos/abc']);
    expect(tools[0].image_results[0].local_path).toBe('assets/images/abc.jpg');
  });

  it('interleaves conversation messages and tool previews by timestamp', () => {
    const messages: ConversationMessage[] = [
      { role: 'user', content: '做大兴安岭 PPT', timestamp: '2026-08-05T00:00:01Z' },
      { role: 'assistant', content: '我先检索事实。', timestamp: '2026-08-05T00:00:03Z' },
    ];
    const events: RuntimeEvent[] = [{
      id: 2,
      task_id: 'task-1',
      timestamp: '2026-08-05T00:00:02Z',
      elapsed_ms: 2000,
      kind: 'tool_end',
      name: 'search',
      status: 'ok',
      metadata: { search_query: '大兴安岭 文化', source_urls: ['https://example.com'] },
    }];

    const items = deriveInlineConversationItems(messages, events);

    expect(items.map(item => item.type)).toEqual(['message', 'tool_group', 'message']);
    expect(items[1].type === 'tool_group' ? items[1].group.tools[0].name : '').toBe('search');
  });

  it('groups adjacent tool previews into one visible tool round', () => {
    const events: RuntimeEvent[] = [
      {
        id: 1,
        task_id: 'task-1',
        timestamp: '2026-08-05T00:00:01Z',
        elapsed_ms: 1000,
        kind: 'tool_end',
        name: 'read_file',
        status: 'ok',
        metadata: { file_path: '/tmp/component_contracts.json' },
      },
      {
        id: 2,
        task_id: 'task-1',
        timestamp: '2026-08-05T00:00:02Z',
        elapsed_ms: 2000,
        kind: 'tool_end',
        name: 'search',
        status: 'ok',
        metadata: { search_query: '2026 两会', source_urls: ['https://example.com'] },
      },
    ];
    const items = deriveInlineConversationItems([], events);

    expect(items).toHaveLength(1);
    expect(items[0].type).toBe('tool_group');
    expect(items[0].type === 'tool_group' ? items[0].group.label : '').toBe('本轮工具调用');
    expect(items[0].type === 'tool_group' ? items[0].group.tools.map(tool => tool.name) : []).toEqual(['read_file', 'search']);
  });

  it('formats tool preview JSON into readable fields', () => {
    const [tool] = deriveInlineToolPreviews([{
      id: 1,
      task_id: 'task-1',
      timestamp: '2026-08-05T00:00:01Z',
      elapsed_ms: 1000,
      kind: 'tool_end',
      name: 'search_images',
      status: 'ok',
      metadata: {
        args_preview: '{"query":"aerial city skyline","asset_purpose":"background","download":true}',
        result_preview: '{"provider":"unsplash","total":12,"photos":[{"id":"a"},{"id":"b"}]}',
      },
    }]);

    expect(formatToolPreviewFields(tool, 'args')).toEqual([
      { label: '检索词', value: 'aerial city skyline' },
      { label: '图片用途', value: 'background' },
      { label: '下载图片', value: '是' },
    ]);
    expect(formatToolPreviewFields(tool, 'result')).toEqual([
      { label: '服务', value: 'unsplash' },
      { label: '总数', value: '12' },
      { label: '图片结果', value: '2 项' },
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
	  kind: 'planner_context_compressed', name: 'context_compressor', status: 'ok',
	  metadata: { before_tokens: 42000, after_tokens: 11000, saved_pct: '73.8%' },
	};
	expect(runtimeEventKindLabel(event)).toBe('上下文压缩');
	expect(runtimeEventNameLabel(event)).toBe('对话上下文');
	expect(runtimeEventDetailLabel(event)).toContain('42,000 → 11,000');
  });

  it('derives observable steps with search query and source urls', () => {
    const steps = deriveObservableSteps([{
      id: 12,
      timestamp: '2026-08-05T00:00:12Z',
      elapsed_ms: 12000,
      kind: 'tool_end',
      name: 'search',
      status: 'ok',
      metadata: {
        search_query: '延安 红色旅游 数据',
        source_urls: ['https://www.yanan.gov.cn/a', 'https://example.com/b'],
      },
    }]);

    expect(steps).toHaveLength(1);
    expect(steps[0].label).toBe('搜索完成：延安 红色旅游 数据');
    expect(steps[0].detail).toContain('2 个来源');
    expect(steps[0].urls).toEqual(['https://www.yanan.gov.cn/a', 'https://example.com/b']);
  });

  it('escapes raw html while rendering markdown blocks', () => {
    const html = renderSafeMarkdown('# 标题\n\n- **重点**\n- `<script>`\n\n<script>alert(1)</script>');
    expect(html).toContain('<h1>标题</h1>');
    expect(html).toContain('<ul>');
    expect(html).toContain('&lt;script&gt;alert(1)&lt;/script&gt;');
    expect(html).not.toContain('<script>');
  });
});
