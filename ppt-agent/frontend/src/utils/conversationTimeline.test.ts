import { describe, expect, it } from 'vitest'
import {
  appendTimelineMessage,
  appendToolInvocation,
  beginObservablePhase,
  completeObservablePhase,
  hideCompletedToolTraces,
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
})
