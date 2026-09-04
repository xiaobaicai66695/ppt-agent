import { describe, expect, it } from 'vitest'
import { isTerminalTaskStreamEvent } from './taskStream'

describe('task stream lifecycle', () => {
  it('keeps the stream open after the planner response ends', () => {
    expect(isTerminalTaskStreamEvent('answer_end')).toBe(false)
  })

  it.each(['complete', 'continue_complete', 'conversation_complete'])('closes the stream only for %s', event => {
    expect(isTerminalTaskStreamEvent(event)).toBe(true)
  })
})
