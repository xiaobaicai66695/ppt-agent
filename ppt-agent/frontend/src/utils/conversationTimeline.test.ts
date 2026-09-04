import { describe, expect, it } from 'vitest'
import { appendExecutionStep, appendTimelineMessage, resetConversationTimeline, updateExecutionStep } from './conversationTimeline'

describe('conversation timeline', () => {
  it('keeps messages and observable execution steps in arrival order', () => {
    const items = resetConversationTimeline([{ role: 'user', content: '找两张图片', timestamp: '2026-09-04T00:00:00Z' }])
    appendExecutionStep(items, '分析请求', '正在判断可用工具')
    appendExecutionStep(items, '图片搜索', '已找到 2 张图片参考', 'success')
    appendTimelineMessage(items, { role: 'assistant', content: '这是两张候选图片。', timestamp: '2026-09-04T00:00:01Z' })

    expect(items.map(item => item.type === 'message' ? item.message.role : item.label)).toEqual([
      'user', '分析请求', '图片搜索', 'assistant',
    ])
  })

  it('updates one tool invocation without moving it in the conversation', () => {
    const items = resetConversationTimeline([])
    const id = appendExecutionStep(items, '联网检索', '正在调用')
    updateExecutionStep(items, id, '联网检索', '已获得 5 条资料', 'success')

    expect(items).toHaveLength(1)
    expect(items[0]).toMatchObject({ type: 'execution', label: '联网检索', detail: '已获得 5 条资料', state: 'success' })
  })
})
