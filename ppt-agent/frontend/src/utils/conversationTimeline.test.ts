import { describe, expect, it } from 'vitest'
import {
  appendTimelineMessage,
  appendRuntimeExecution,
  appendToolInvocation,
  beginObservablePhase,
  completeObservablePhase,
  finishToolPhase,
  hideCompletedToolTraces,
  prepareToolBoundary,
  resetConversationTimeline,
  resolveToolInvocation,
  toggleToolInvocation,
} from './conversationTimeline'

describe('conversation timeline', () => {
  it('keeps observable phases, their tools, and replies in arrival order', () => {
    const items = resetConversationTimeline([{ role: 'user', content: '找两张图片', timestamp: '2026-09-04T00:00:00Z' }])
    const analysisID = beginObservablePhase(items, 'analysis', '分析请求', '正在判断可用工具')
    appendToolInvocation(items, analysisID, 'search_images', '图片搜索', '正在搜索图片参考')
    resolveToolInvocation(items, analysisID, 'search_images', '图片搜索', '已找到 2 张图片参考')
    beginObservablePhase(items, 'answer', '组织回答', '正在组织回答')
    appendTimelineMessage(items, { role: 'assistant', content: '这是两张候选图片。', timestamp: '2026-09-04T00:00:01Z' })

    expect(items.map(item => item.type === 'message' ? item.message.role : item.label)).toEqual([
      'user', '分析请求', '组织回答', 'assistant',
    ])
    expect(items[1]).toMatchObject({ type: 'phase', state: 'success', tools: [{ label: '图片搜索', state: 'success' }] })
  })

  it('updates only the matching invocation and keeps sibling calls in order', () => {
    const items = resetConversationTimeline([])
    const analysisID = beginObservablePhase(items, 'analysis', '分析请求')
    appendToolInvocation(items, analysisID, 'search', '联网检索', '正在检索')
    appendToolInvocation(items, analysisID, 'search_images', '图片搜索', '正在搜索')
    resolveToolInvocation(items, analysisID, 'search', '联网检索', '已获取 5 条资料')

    expect(items[0]).toMatchObject({ type: 'phase', tools: [
      { label: '联网检索', resultDetail: '已获取 5 条资料', state: 'success' },
      { label: '图片搜索', state: 'running' },
    ] })
  })

  it('resolves repeated tool names by provider call id', () => {
    const items = resetConversationTimeline([])
    const phaseID = beginObservablePhase(items, 'analysis', '分析请求')
    appendToolInvocation(items, phaseID, 'search', '联网检索', '第一次', undefined, 'call-1')
    appendToolInvocation(items, phaseID, 'search', '联网检索', '第二次', undefined, 'call-2')
    resolveToolInvocation(items, phaseID, 'search', '联网检索', '第一次完成', 'success', undefined, 'call-1')

    expect(items[0]).toMatchObject({ type: 'phase', tools: [
      { callID: 'call-1', resultDetail: '第一次完成', state: 'success' },
      { callID: 'call-2', state: 'running' },
    ] })
  })

  it('lets each tool invocation independently expand or collapse', () => {
    const items = resetConversationTimeline([])
    const analysisID = beginObservablePhase(items, 'analysis', '分析请求')
    const firstID = appendToolInvocation(items, analysisID, 'search', '联网检索')!
    const secondID = appendToolInvocation(items, analysisID, 'search_images', '图片搜索')!

    toggleToolInvocation(items, firstID)

    expect(items[0]).toMatchObject({ type: 'phase', tools: [
      { id: firstID, expanded: false },
      { id: secondID, expanded: true },
    ] })
  })

  it('recovers from a result replayed without its earlier call event', () => {
    const items = resetConversationTimeline([])
    const analysisID = beginObservablePhase(items, 'analysis', '分析请求')
    resolveToolInvocation(items, analysisID, 'search', '联网检索', '已获取 5 条资料')
    completeObservablePhase(items, 'analysis')

    expect(items[0]).toMatchObject({ type: 'phase', state: 'success', tools: [{ name: 'search', state: 'success' }] })
  })

  it('removes transient tool phases once the task reaches a terminal state', () => {
    const items = resetConversationTimeline([{ role: 'user', content: '找图', timestamp: '' }])
    const phaseID = beginObservablePhase(items, 'analysis', '分析请求')
    appendToolInvocation(items, phaseID, 'search_images', '图片搜索')
    appendTimelineMessage(items, { role: 'assistant', content: '完成', timestamp: '' })
    hideCompletedToolTraces(items)
    expect(items.map(item => item.type)).toEqual(['message', 'message'])
  })

  it('keeps a tool phase after the preceding assistant segment', () => {
    const items = resetConversationTimeline([{ role: 'user', content: '查资料', timestamp: '' }])
    const phaseID = beginObservablePhase(items, 'analysis', '分析请求')
    appendTimelineMessage(items, { role: 'assistant', content: '我先查一下。', timestamp: '' })
    prepareToolBoundary(items, phaseID)
    appendToolInvocation(items, phaseID, 'search', '联网检索')

    expect(items.map(item => item.type === 'message' ? item.message.role : item.type)).toEqual(['user', 'assistant', 'phase'])
    expect(finishToolPhase(items, phaseID)).toBe(true)
    expect(items[2]).toMatchObject({ type: 'phase', state: 'success' })
    appendTimelineMessage(items, { role: 'assistant', content: '查到了一些资料。', timestamp: '' })
    expect(items.map(item => item.type === 'message' ? item.message.role : item.type)).toEqual(['user', 'assistant', 'phase', 'assistant'])
  })

  it('merges LLM start and end events into one execution row', () => {
    const items = resetConversationTimeline([])
    appendRuntimeExecution(items, 'ChatModel', '', 'running', 'llm_start')
    appendRuntimeExecution(items, 'chat_model', '', 'success', 'llm_end')

    expect(items).toHaveLength(1)
    expect(items[0]).toMatchObject({ type: 'execution', label: 'ChatModel', state: 'success' })
  })
})
