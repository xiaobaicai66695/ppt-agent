import { describe, expect, it } from 'vitest'
import { shouldStartPPTGeneration } from './messageRouting'

describe('message routing', () => {
  it('starts PPT generation when intent recognition returns create in chat mode', () => {
    expect(shouldStartPPTGeneration({ intent: 'create', action: 'prepare_create' }, 'chat')).toBe(true)
  })

  it('keeps the manual PPT choice as a create override', () => {
    expect(shouldStartPPTGeneration({ intent: 'chat' }, 'pptagent')).toBe(true)
  })

  it('streams an ordinary reply for non-create chat intent', () => {
    expect(shouldStartPPTGeneration({ intent: 'chat' }, 'chat')).toBe(false)
  })

  it('does not start from an uncertain creation recommendation', () => {
    expect(shouldStartPPTGeneration({ intent: 'create', action: 'ask_clarification', needs_confirmation: true }, 'chat')).toBe(false)
  })
})
