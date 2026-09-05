import { describe, expect, it } from 'vitest'
import { appendDeliveryDirectives, shouldStartPPTGeneration } from './messageRouting'

describe('message routing', () => {
  it('starts PPT generation when intent recognition returns create in chat mode', () => {
    expect(shouldStartPPTGeneration({ intent: 'create', action: 'prepare_create' })).toBe(true)
  })

  it('appends every selected delivery directive to the routed message', () => {
    expect(appendDeliveryDirectives('介绍人工智能趋势', 'pptagent', true, true)).toBe('介绍人工智能趋势\n\n（生成PPT） （网络搜索） （图片搜索）')
  })

  it('streams an ordinary reply for non-create chat intent', () => {
    expect(shouldStartPPTGeneration({ intent: 'chat' })).toBe(false)
  })

  it('does not start from an uncertain creation recommendation', () => {
    expect(shouldStartPPTGeneration({ intent: 'create', action: 'ask_clarification', needs_confirmation: true })).toBe(false)
  })
})
