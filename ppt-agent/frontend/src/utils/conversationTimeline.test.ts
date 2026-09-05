import { describe, expect, it } from 'vitest'
import {
  appendFinalAnswer,
  appendThought,
  appendToolCall,
  appendToolResult,
  finishStreamingEntries,
  finishTimelineEntries,
  resetConversationTimeline,
  toggleTimelineItem,
} from './conversationTimeline'

describe('conversation timeline', () => {
  it('keeps thoughts, paired tool cards, and final answers in arrival order', () => {
    const items = resetConversationTimeline([{ role: 'user', content: '找两张图片', timestamp: '2026-09-04T00:00:00Z' }])

    appendThought(items, '先判断需要哪些资料。', { eventID: 1, segmentID: 'thought-1', phase: 'analysis', delta: true })
    appendToolCall(items, { eventID: 2, callID: 'call-1', name: 'search_images', label: '图片搜索', args: '{"query":"PPT"}' })
    appendToolResult(items, { eventID: 3, callID: 'call-1', name: 'search_images', label: '图片搜索', result: '已找到 2 张图片参考' })
    appendThought(items, '根据候选图片组织回答。', { eventID: 4, segmentID: 'thought-2', phase: 'answer', delta: true })
    appendFinalAnswer(items, '这是两张候选图片。', { eventID: 5, segmentID: 'answer-1', delta: true })

    expect(items.map(item => item.type)).toEqual(['message', 'thought', 'tool_call', 'thought', 'final_answer'])
    expect(items[1]).toMatchObject({ type: 'thought', state: 'success', content: '先判断需要哪些资料。' })
    expect(items[2]).toMatchObject({ type: 'tool_call', callID: 'call-1', state: 'success', result: '已找到 2 张图片参考' })
  })

  it('updates one streaming segment and deduplicates replayed event IDs', () => {
    const items = resetConversationTimeline([])

    appendThought(items, '正在', { eventID: 11, segmentID: 'thought-1', delta: true })
    appendThought(items, '分析', { eventID: 12, segmentID: 'thought-1', delta: true })
    appendThought(items, '正在', { eventID: 11, segmentID: 'thought-1', delta: true })
    appendFinalAnswer(items, '答案', { eventID: 13, segmentID: 'answer-1', delta: true })
    appendFinalAnswer(items, '如下。', { eventID: 14, segmentID: 'answer-1', delta: true })
    appendFinalAnswer(items, '答案', { eventID: 13, segmentID: 'answer-1', delta: true })

    expect(items).toHaveLength(2)
    expect(items[0]).toMatchObject({ type: 'thought', content: '正在分析', eventIDs: [11, 12] })
    expect(items[1]).toMatchObject({ type: 'final_answer', content: '答案如下。', eventIDs: [13, 14] })
  })

  it('keeps a completed call in place when a later observation arrives', () => {
    const items = resetConversationTimeline([])
    appendToolCall(items, { eventID: 21, callID: 'call-1', name: 'search', label: '联网检索' })
    appendThought(items, '工具仍在返回资料。', { eventID: 22, segmentID: 'thought-2', delta: true })
    appendToolResult(items, { eventID: 23, callID: 'call-1', name: 'search', label: '联网检索', result: '已获取 5 条资料' })

    expect(items.map(item => item.type)).toEqual(['tool_call', 'thought'])
    expect(items[0]).toMatchObject({ type: 'tool_call', callID: 'call-1', state: 'success' })
    expect(items[0]).toMatchObject({ result: '已获取 5 条资料' })
  })

  it('recovers an observation replayed without its earlier call frame', () => {
    const items = resetConversationTimeline([])
    appendToolResult(items, { eventID: 31, callID: 'call-1', name: 'search', label: '联网检索', result: '已获取 5 条资料' })

    expect(items).toEqual([expect.objectContaining({ type: 'tool_call', callID: 'call-1', state: 'success', result: '已获取 5 条资料' })])
  })

  it('keeps long tool results intact inside the paired call card', () => {
    const items = resetConversationTimeline([])
    const thoughtID = appendThought(items, '安全的过程摘要', { eventID: 41, segmentID: 'thought-1' })!
    const resultID = appendToolResult(items, { eventID: 42, callID: 'call-1', name: 'search', label: '联网检索', result: '资料'.repeat(140) })!

    expect(items[0]).toMatchObject({ id: thoughtID, expanded: true })
    expect(items[1]).toMatchObject({ id: resultID, type: 'tool_call', expanded: true, result: '资料'.repeat(140) })
    expect(toggleTimelineItem(items, thoughtID)).toBe(false)
    expect(toggleTimelineItem(items, resultID)).toBe(false)
    expect(items[0]).toMatchObject({ expanded: false })
    expect(items[1]).toMatchObject({ expanded: false })
  })

  it('keeps trace cards after text streaming ends', () => {
    const items = resetConversationTimeline([])
    appendThought(items, '先检索资料。', { eventID: 51, segmentID: 'thought-1', delta: true })
    appendToolCall(items, { eventID: 52, callID: 'call-1', name: 'search', label: '联网检索' })
    appendToolResult(items, { eventID: 53, callID: 'call-1', name: 'search', label: '联网检索', result: '已获取 5 条资料' })
    appendFinalAnswer(items, '根据资料，答案如下。', { eventID: 54, segmentID: 'answer-1', delta: true })

    finishStreamingEntries(items)

    expect(items.map(item => item.type)).toEqual(['thought', 'tool_call', 'final_answer'])
    expect(items[2]).toMatchObject({ type: 'final_answer', streaming: false })
  })

  it('keeps a tool call collapsible between llm_end and the next llm_start', () => {
    const items = resetConversationTimeline([])

    appendThought(items, '先确定需要检索。', { eventID: 51, segmentID: 'thought-1', delta: true })
    finishTimelineEntries(items, { includeTools: false })
    appendToolCall(items, { eventID: 52, name: 'search', label: '联网检索', args: '{"query":"同义词"}' })
    finishTimelineEntries(items, { includeTools: false })
    appendToolResult(items, { eventID: 53, name: 'search', label: '联网检索', args: '{"query":"同义词"}', result: '已获取同义词资料' })
    appendThought(items, '根据检索结果继续回答。', { eventID: 54, segmentID: 'thought-2', delta: true })

    expect(items.map(item => item.type)).toEqual(['thought', 'tool_call', 'thought'])
    expect(items[1]).toMatchObject({ type: 'tool_call', state: 'success', expanded: true, result: '已获取同义词资料' })
    expect(toggleTimelineItem(items, items[1].id)).toBe(false)
    expect(items[1]).toMatchObject({ expanded: false })
  })

  it('repairs a late tool result after an older boundary marked the call unfinished', () => {
    const items = resetConversationTimeline([])

    appendToolCall(items, { eventID: 61, name: 'search', label: '联网检索' })
    finishStreamingEntries(items)
    appendToolResult(items, { eventID: 62, name: 'search', label: '联网检索', result: '迟到的真实结果' })

    expect(items).toHaveLength(1)
    expect(items[0]).toMatchObject({ type: 'tool_call', state: 'success', result: '迟到的真实结果' })
  })

  it('starts a fresh thought card after a finished segment receives the same segment id again', () => {
    const items = resetConversationTimeline([])
    appendThought(items, '先检索资料。', { eventID: 51, segmentID: 'thought-1', delta: true })
    finishStreamingEntries(items)
    appendThought(items, '继续整理。', { eventID: 52, segmentID: 'thought-1', delta: true })

    expect(items).toHaveLength(2)
    expect(items[0]).toMatchObject({ type: 'thought', content: '先检索资料。', streaming: false })
    expect(items[1]).toMatchObject({ type: 'thought', content: '继续整理。', streaming: true })
  })

  it('starts a fresh final-answer card after a finished segment receives the same segment id again', () => {
    const items = resetConversationTimeline([])
    appendFinalAnswer(items, '第一段答案。', { eventID: 61, segmentID: 'answer-1', delta: true })
    finishStreamingEntries(items)
    appendFinalAnswer(items, '第二段答案。', { eventID: 62, segmentID: 'answer-1', delta: true })

    expect(items).toHaveLength(2)
    expect(items[0]).toMatchObject({ type: 'final_answer', content: '第一段答案。', streaming: false })
    expect(items[1]).toMatchObject({ type: 'final_answer', content: '第二段答案。', streaming: true })
  })

  it('keeps legacy answer chunks in a single final-answer segment', () => {
    const items = resetConversationTimeline([])
    appendFinalAnswer(items, '旧协议', { eventID: 61, segmentID: 'legacy-final-answer', delta: true })
    appendFinalAnswer(items, '仍可显示。', { eventID: 62, segmentID: 'legacy-final-answer', delta: true })

    expect(items).toEqual([expect.objectContaining({ type: 'final_answer', content: '旧协议仍可显示。' })])
  })

  it('deduplicates repeated tool results by call id', () => {
    const items = resetConversationTimeline([])
    appendToolCall(items, { eventID: 71, callID: 'call-1', name: 'search', label: '联网检索' })
    const first = appendToolResult(items, { eventID: 72, callID: 'call-1', name: 'search', label: '联网检索', result: '首个结果' })
    const second = appendToolResult(items, { eventID: 73, callID: 'call-1', name: 'search', label: '联网检索', result: '重复结果' })

    expect(first).toBeDefined()
    expect(second).toBe(first)
    expect(items).toHaveLength(1)
    expect(items[0]).toMatchObject({ type: 'tool_call', callID: 'call-1', result: '首个结果', eventIDs: [71, 72, 73] })
  })
})
