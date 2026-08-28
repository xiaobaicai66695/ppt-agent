import { describe, expect, it } from 'vitest';
import type { ConversationMessage, ConversationSession, RuntimeEvent, TaskItem } from '../types';
import {
  appendAssistantStreamContent, canonicalOutputFile, compactRuntimeEvents, deriveInlineConversationItems, deriveInlineToolPreviews,
  deriveLiveActivity, deriveObservableSteps, formatToolPreviewFields, mergeConversationMessages,
  mergeRuntimeEvents, mergeRuntimeMeta, mergeSlideDeliveries, nextReplayCursor, recoverConversationMessages, renderSafeMarkdown,
  RUNTIME_EVENT_TAIL_LIMIT,
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

  it('drops assistant runtime fragments already contained in a restored full answer', () => {
    const fullAnswer = [
      '我将为您生成一份12页的PPT，主题为“中小后端项目Git版本控制规范方案”。',
      'PPT规划已完成，12页内容已提交。以下是各页概要：',
      '| 页码 | 标题 |',
      '| --- | --- |',
      '| 12 | 致谢 |',
      'DeckSpec已锁定并发布，后端将自动渲染生成PPT文件。',
    ].join('\n');
    const fragment = 'PPT规划已完成，12页内容已提交。以下是各页概要：\n\n| 页码 | 标题 |\n| --- | --- |\n| 12 | 致谢 |';

    const merged = mergeConversationMessages(
      [{ role: 'assistant', content: fullAnswer, timestamp: 'a' }],
      [{ role: 'assistant', content: fragment, timestamp: 'b' }],
    );

    expect(merged).toHaveLength(1);
    expect(merged[0].content).toBe(fullAnswer);
  });

  it('compacts duplicated assistant prefixes already present in restored messages', () => {
    const prefix = '我将为您规划一份关于“银发友好社区：从适老改造到服务运营”的演示文稿。首先，让我读取组件契约文件以了解可用的组件类型。';
    const suffix = '我将为“银发友好社区：从适老改造到服务运营”设计一份15页的演示文稿。先搜索一些背景资料确保内容准确。';

    const merged = mergeConversationMessages([
      { role: 'user', content: '「银发友好社区：从适老改造到服务运营」', timestamp: 'a' },
      { role: 'assistant', content: prefix, timestamp: 'b' },
      { role: 'assistant', content: `${prefix}\n\n${suffix}`, timestamp: 'c' },
    ], []);

    expect(merged.map(message => message.content)).toEqual([
      '「银发友好社区：从适老改造到服务运营」',
      prefix,
      suffix,
    ]);
  });

  it('keeps one live assistant stream when chunks are cumulative snapshots', () => {
    const first = '我将为您规划一份关于“银发友好社区”的演示文稿。';
    const second = `${first}首先，让我读取组件契约文件。`;

    const streamed = appendAssistantStreamContent(
      appendAssistantStreamContent('', first),
      second,
    );

    expect(streamed).toBe(second);
  });

  it('restores readable English word boundaries in live assistant deltas', () => {
    const streamed = ['I\'ll', 'start', 'by', 'reading', 'the', 'required', 'skill', 'files']
      .reduce((content, chunk) => appendAssistantStreamContent(content, chunk), '');

    expect(streamed).toBe('I\'ll start by reading the required skill files');
  });

  it('restores readable English word boundaries in cumulative assistant snapshots', () => {
    const streamed = appendAssistantStreamContent(
      appendAssistantStreamContent('', 'I\'ll'),
      'I\'llstartbyreading',
    );

    expect(streamed).toBe('I\'ll startbyreading');
  });

  it('keeps only the new assistant suffix when a cumulative output follows tools', () => {
    const prefix = '我将为您创建一个关于微服务项目治理的20页PPT。首先，让我读取组件契约文件以了解可用的组件类型和版式。';
    const suffix = '我已了解组件契约。现在让我规划20页的微服务项目治理PPT。';
    const events: RuntimeEvent[] = [
      {
        id: 1,
        task_id: 'task-1',
        timestamp: '2026-08-05T00:00:01Z',
        elapsed_ms: 1000,
        kind: 'llm_end',
        name: 'planner',
        status: 'ok',
        metadata: { assistant_output: prefix },
      },
      {
        id: 2,
        task_id: 'task-1',
        timestamp: '2026-08-05T00:00:02Z',
        elapsed_ms: 2000,
        kind: 'tool_end',
        name: 'read_file',
        status: 'ok',
        metadata: { file_path: 'component_contracts.json' },
      },
      {
        id: 3,
        task_id: 'task-1',
        timestamp: '2026-08-05T00:00:03Z',
        elapsed_ms: 3000,
        kind: 'llm_end',
        name: 'planner',
        status: 'ok',
        metadata: { assistant_output: `${prefix}\n\n${suffix}` },
      },
    ];

    const messages = runtimeAssistantOutputMessages(events);
    const items = deriveInlineConversationItems(messages, events);

    expect(messages.map(message => message.content)).toEqual([prefix, suffix]);
    expect(items.map(item => item.type)).toEqual(['message', 'tool_group', 'message']);
    expect(items[2].type === 'message' ? items[2].message.content : '').not.toContain(prefix);
  });

  it('removes short earlier turns from a much longer final cumulative snapshot', () => {
    const first = '我将完成这个PPT的规划工作。首先读取必要的参考文件。';
    const second = '好的，我已完整读取所有参考文件。现在我来规划“新时代：AI+赋能”这个主题。';
    const final = [
      '我将完成这个 PPT 的规划工作。 首先读取必要的参考文件。',
      '好的，我已完整读取所有参考文件。现在我来规划“新时代：AI+赋能”这个主题。',
      '现在我已收集了足够的数据和图片素材，开始制定完整的10页PPT规划。',
      '纲要结构设计：',
      ...Array.from({ length: 10 }, (_, index) => `${index + 1}. 第${index + 1}页内容与关键数据说明`),
    ].join('\n\n');
    const events: RuntimeEvent[] = [
      {
        id: 1, task_id: 'task-1', timestamp: '2026-08-25T00:00:01Z', elapsed_ms: 1000,
        kind: 'llm_end', name: 'planner', status: 'ok', metadata: { assistant_output: first },
      },
      {
        id: 2, task_id: 'task-1', timestamp: '2026-08-25T00:00:02Z', elapsed_ms: 2000,
        kind: 'tool_end', name: 'read_file', status: 'ok',
      },
      {
        id: 3, task_id: 'task-1', timestamp: '2026-08-25T00:00:03Z', elapsed_ms: 3000,
        kind: 'llm_end', name: 'planner', status: 'ok', metadata: { assistant_output: `${first}\n\n${second}` },
      },
      {
        id: 4, task_id: 'task-1', timestamp: '2026-08-25T00:00:04Z', elapsed_ms: 4000,
        kind: 'tool_end', name: 'search', status: 'ok',
      },
      {
        id: 5, task_id: 'task-1', timestamp: '2026-08-25T00:00:05Z', elapsed_ms: 5000,
        kind: 'llm_end', name: 'planner', status: 'ok', metadata: { assistant_output: final },
      },
    ];

    const messages = runtimeAssistantOutputMessages(events);

    expect(messages).toHaveLength(3);
    expect(messages[0].content).toBe(first);
    expect(messages[1].content).toBe(second);
    expect(messages[2].content).toContain('现在我已收集了足够的数据和图片素材');
    expect(messages[2].content).not.toContain('我将完成这个 PPT 的规划工作');
    expect(messages.map(message => message.content).join('\n')).toContain('10. 第10页内容与关键数据说明');
  });

  it('keeps the fuller assistant output when runtime events are cumulative', () => {
    const events: RuntimeEvent[] = [
      {
        id: 1,
        task_id: 'task-1',
        timestamp: '2026-08-05T00:00:01Z',
        elapsed_ms: 1000,
        kind: 'llm_end',
        name: 'planner',
        status: 'ok',
        metadata: { assistant_output: 'PPT规划已完成，12页内容已提交。以下是各页概要：' },
      },
      {
        id: 2,
        task_id: 'task-1',
        timestamp: '2026-08-05T00:00:02Z',
        elapsed_ms: 2000,
        kind: 'llm_end',
        name: 'planner',
        status: 'ok',
        metadata: { assistant_output: 'PPT规划已完成，12页内容已提交。以下是各页概要：\n\n| 页码 | 标题 |\n| --- | --- |\n| 12 | 致谢 |' },
      },
    ];

    const messages = runtimeAssistantOutputMessages(events);

    expect(messages).toHaveLength(1);
    expect(messages[0].content).toContain('| 12 | 致谢 |');
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
      runtime_event_id: 2,
    }]);
  });

  it('orders runtime assistant text and tools by event id when timestamps drift', () => {
    const first = '先读取组件契约和页面类型说明，确认当前生成器支持哪些稳定布局。';
    const second = '已读取组件契约，继续搜索资料。';
    const events: RuntimeEvent[] = [
      {
        id: 1,
        task_id: 'task-1',
        timestamp: '2026-08-05T00:00:04Z',
        elapsed_ms: 1000,
        kind: 'llm_end',
        name: 'planner',
        status: 'ok',
        metadata: { assistant_output: first },
      },
      {
        id: 2,
        task_id: 'task-1',
        timestamp: '2026-08-05T00:00:02Z',
        elapsed_ms: 2000,
        kind: 'tool_end',
        name: 'read_file',
        status: 'ok',
        metadata: { file_path: 'component_contracts.json' },
      },
      {
        id: 3,
        task_id: 'task-1',
        timestamp: '2026-08-05T00:00:05Z',
        elapsed_ms: 3000,
        kind: 'llm_end',
        name: 'planner',
        status: 'ok',
        metadata: { assistant_output: `${first}\n\n${second}` },
      },
    ];

    const items = deriveInlineConversationItems(runtimeAssistantOutputMessages(events), events);

    expect(items.map(item => item.type)).toEqual(['message', 'tool_group', 'message']);
    expect(items[0].type === 'message' ? items[0].message.content : '').toBe(first);
    expect(items[2].type === 'message' ? items[2].message.content : '').toBe(second);
  });

  it('uses visible assistant output runtime events to split text around tools', () => {
    const events: RuntimeEvent[] = [
      {
        id: 1,
        task_id: 'task-1',
        timestamp: '2026-08-26T00:00:01Z',
        elapsed_ms: 1000,
        kind: 'assistant_output',
        name: 'visible_answer',
        status: 'ok',
        metadata: { assistant_output: '先读取组件契约。' },
      },
      {
        id: 2,
        task_id: 'task-1',
        timestamp: '2026-08-26T00:00:02Z',
        elapsed_ms: 2000,
        kind: 'tool_end',
        name: 'read_file',
        status: 'ok',
        metadata: { file_path: 'component_contracts.json' },
      },
      {
        id: 3,
        task_id: 'task-1',
        timestamp: '2026-08-26T00:00:03Z',
        elapsed_ms: 3000,
        kind: 'assistant_output',
        name: 'visible_answer',
        status: 'ok',
        metadata: { assistant_output: '已读取完成，继续搜索资料。' },
      },
      {
        id: 4,
        task_id: 'task-1',
        timestamp: '2026-08-26T00:00:04Z',
        elapsed_ms: 4000,
        kind: 'llm_end',
        name: 'planner',
        status: 'ok',
        metadata: { assistant_output: '先读取组件契约。已读取完成，继续搜索资料。' },
      },
    ];

    const messages = runtimeAssistantOutputMessages(events);
    const items = deriveInlineConversationItems(messages, events);

    expect(messages.map(message => message.content)).toEqual(['先读取组件契约。', '已读取完成，继续搜索资料。']);
    expect(items.map(item => item.type)).toEqual(['message', 'tool_group', 'message']);
  });

  it('orders live assistant segments by timeline order before timestamps', () => {
    const messages: ConversationMessage[] = [
      {
        role: 'assistant',
        content: '我先说明下一步。',
        timestamp: '2026-08-26T00:00:05Z',
        timeline_order: 1,
      },
      {
        role: 'assistant',
        content: '工具完成后继续输出。',
        timestamp: '2026-08-26T00:00:01Z',
        timeline_order: 3,
      },
    ];
    const events: RuntimeEvent[] = [{
      id: 2,
      task_id: 'task-1',
      timestamp: '2026-08-26T00:00:02Z',
      elapsed_ms: 2000,
      kind: 'tool_end',
      name: 'read_file',
      status: 'ok',
      metadata: { file_path: 'component_contracts.json' },
    }];

    const items = deriveInlineConversationItems(messages, events);

    expect(items.map(item => item.type)).toEqual(['message', 'tool_group', 'message']);
    expect(items[0].type === 'message' ? items[0].message.content : '').toBe('我先说明下一步。');
    expect(items[2].type === 'message' ? items[2].message.content : '').toBe('工具完成后继续输出。');
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
      runtime_event_id: 3,
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

  it('pairs concurrent tool start/end events by query before falling back to tool name', () => {
    const events: RuntimeEvent[] = [
      {
        id: 1,
        task_id: 'task-1',
        timestamp: '2026-08-26T00:00:01Z',
        elapsed_ms: 1000,
        kind: 'tool_start',
        name: 'search_images',
        status: 'running',
        metadata: { image_query: 'world map globe soft light' },
      },
      {
        id: 2,
        task_id: 'task-1',
        timestamp: '2026-08-26T00:00:02Z',
        elapsed_ms: 2000,
        kind: 'tool_start',
        name: 'search_images',
        status: 'running',
        metadata: { image_query: 'international flags government building' },
      },
      {
        id: 3,
        task_id: 'task-1',
        timestamp: '2026-08-26T00:00:03Z',
        elapsed_ms: 3000,
        kind: 'tool_end',
        name: 'search_images',
        status: 'ok',
        metadata: { image_query: 'international flags government building', image_results: [{ id: 'flags', preview_url: 'https://images.example/flags.jpg' }] },
      },
      {
        id: 4,
        task_id: 'task-1',
        timestamp: '2026-08-26T00:00:04Z',
        elapsed_ms: 4000,
        kind: 'tool_end',
        name: 'search_images',
        status: 'ok',
        metadata: { image_query: 'world map globe soft light', image_results: [{ id: 'map', preview_url: 'https://images.example/map.jpg' }] },
      },
    ];

    const tools = deriveInlineToolPreviews(events);

    expect(tools).toHaveLength(2);
    expect(tools[0].start_event_id).toBe(1);
    expect(tools[0].end_event_id).toBe(4);
    expect(tools[0].image_results[0].id).toBe('map');
    expect(tools[1].start_event_id).toBe(2);
    expect(tools[1].end_event_id).toBe(3);
    expect(tools[1].image_results[0].id).toBe('flags');
  });

  it('deduplicates repeated completed image-search previews with the same query', () => {
    const duplicate = (id: number): RuntimeEvent => ({
      id,
      task_id: 'task-1',
      timestamp: `2026-08-26T00:00:0${id}Z`,
      elapsed_ms: id * 1000,
      kind: 'tool_end',
      name: 'search_images',
      status: 'ok',
      metadata: {
        image_query: 'world map globe soft light wide landscape clean negative space',
        image_results: [{ id: 'map', preview_url: 'https://images.example/map.jpg' }],
      },
    });

    const tools = deriveInlineToolPreviews([duplicate(1), duplicate(2)]);

    expect(tools).toHaveLength(1);
    expect(tools[0].end_event_id).toBe(2);
  });

  it('shows image search purpose in preview detail and args fields', () => {
    const [tool] = deriveInlineToolPreviews([{
      id: 1,
      task_id: 'task-1',
      timestamp: '2026-08-26T00:00:01Z',
      elapsed_ms: 1000,
      kind: 'tool_end',
      name: 'search_images',
      status: 'ok',
      metadata: {
        image_query: 'diplomatic negotiation table',
        search_reason: '用于外交协商页面的图文混排示例',
        image_results: [{ id: 'diplomacy', preview_url: 'https://images.example/diplomacy.jpg' }],
      },
    }]);

    expect(tool.detail).toContain('目的：用于外交协商页面');
    expect(formatToolPreviewFields(tool, 'args')).toContainEqual({ label: '搜索原因', value: '用于外交协商页面的图文混排示例' });
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

  it('keeps user message before assistant reply when timestamps are equal', () => {
    const timestamp = '2026-08-05T00:00:01Z';
    const messages: ConversationMessage[] = [
      { role: 'user', content: '你好', timestamp },
      { role: 'assistant', content: '你好，我在。', timestamp },
    ];

    const items = deriveInlineConversationItems(messages, []);

    expect(items.map(item => item.type === 'message' ? item.message.role : item.type)).toEqual(['user', 'assistant']);
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
        result: '{"provider":"unsplash","total":12,"photos":[{"id":"a","preview_url":"https://images.unsplash.com/a.jpg"},{"id":"b","preview_url":"https://images.unsplash.com/b.jpg"}]}',
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
      { label: '图片数量', value: '2' },
      { label: '已下载', value: '0' },
    ]);
    expect(tool.image_results).toHaveLength(2);
  });

  it('prefers full tool result over truncated preview for readable fields', () => {
    const [tool] = deriveInlineToolPreviews([{
      id: 1,
      task_id: 'task-1',
      timestamp: '2026-08-05T00:00:01Z',
      elapsed_ms: 1000,
      kind: 'tool_end',
      name: 'search',
      status: 'ok',
      metadata: {
        result_preview: '{"results":[{"title":"截断',
        result: '{"results":[{"title":"政策原文","url":"https://example.com/policy","description":"政策明确了试点范围和执行时间。","source":"示例政府网","date":"2026-08-01"}],"content":"这是一段很长的搜索正文摘要，用于说明前端应该解析完整 result 而不是截断 preview。"}',
      },
    }]);

    expect(formatToolPreviewFields(tool, 'result')).toEqual([
      { label: '来源数量', value: '1' },
      { label: '主要来源', value: '政策原文' },
    ]);
    expect(tool.search_results).toEqual([{
      title: '政策原文',
      url: 'https://example.com/policy',
      description: '政策明确了试点范围和执行时间。',
      source: '示例政府网',
      date: '2026-08-01',
    }]);
  });

  it('derives concise live activity without exposing tool arguments', () => {
    expect(deriveLiveActivity({ status: 'running', lastTool: 'search', done: 3, total: 12 })).toEqual({
      label: '正在检索并核实资料', detail: '已完成 3/12 页', state: 'running',
    });
    expect(deriveLiveActivity({ status: 'running', phase: 'compressing_context' }).label).toBe('正在压缩较早对话');
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
	const persisted = [event(1, 'read_file'), event(2, 'patch_tasks_draft')];
	const tail = [event(2, 'patch_tasks_draft', { result: 'updated' }), event(3, 'update_tasks_manifest')];
	const merged = mergeRuntimeEvents(persisted, tail);
	expect(merged.map(item => item.name)).toEqual(['read_file', 'patch_tasks_draft', 'update_tasks_manifest']);
	expect(merged[1].metadata).toEqual({ result: 'updated' });

	const meta = mergeRuntimeMeta({ elapsed_ms: 1, recent_events: persisted }, { elapsed_ms: 3, recent_events: tail });
	expect(meta.elapsed_ms).toBe(3);
	expect(meta.recent_events).toHaveLength(3);
  });

  it('bounds repeated runtime snapshot merges to the live event tail', () => {
    const event = (id: number): RuntimeEvent => ({
      id, task_id: 'task-1', timestamp: `2026-08-05T00:00:${String(id).padStart(2, '0')}Z`, elapsed_ms: id,
      kind: 'assistant_output', name: 'visible_answer', status: 'ok', metadata: { assistant_output: `chunk-${id}` },
    });
    const merged = Array.from({ length: RUNTIME_EVENT_TAIL_LIMIT + 30 }, (_, index) => index + 1)
      .reduce<RuntimeEvent[]>((current, id) => mergeRuntimeEvents(current, [event(id)]), []);

    expect(merged).toHaveLength(RUNTIME_EVENT_TAIL_LIMIT);
    expect(merged[0].id).toBe(31);
    expect(merged[merged.length - 1]?.id).toBe(RUNTIME_EVENT_TAIL_LIMIT + 30);
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

  it('normalizes compact assistant markdown into readable sections', () => {
    const html = renderSafeMarkdown('规划完成说明页数：12页叙事主线：观点。 ##规划完成说明页面结构：1.封面（title_slide）2.目录（agenda）3.章节一（section_divider）');
    expect(html).toContain('<h2>规划完成说明</h2>');
    expect(html).toContain('<p>页面结构：</p>');
    expect(html).toContain('<ol>');
    expect(html).toContain('<li>封面（title_slide）</li>');
    expect(html).toContain('<li>目录（agenda）</li>');
    expect(html).not.toContain('section_div ider');
  });
});
