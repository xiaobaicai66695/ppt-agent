import type { ConversationMessage } from '../types'

export type ExecutionState = 'running' | 'success' | 'error'

export type ConversationTimelineItem =
  | { id: string; type: 'message'; message: ConversationMessage }
  | { id: string; type: 'execution'; label: string; detail?: string; state: ExecutionState }

let ordinal = 0

function nextID(prefix: string) {
  ordinal += 1
  return `${prefix}-${ordinal}`
}

export function resetConversationTimeline(messages: ConversationMessage[]): ConversationTimelineItem[] {
  ordinal = 0
  return messages.map(message => ({ id: nextID('message'), type: 'message' as const, message }))
}

export function appendTimelineMessage(items: ConversationTimelineItem[], message: ConversationMessage) {
  items.push({ id: nextID('message'), type: 'message', message })
}

export function appendExecutionStep(items: ConversationTimelineItem[], label: string, detail = '', state: ExecutionState = 'running') {
  const id = nextID('execution')
  items.push({ id, type: 'execution', label, detail, state })
  return id
}

export function updateExecutionStep(items: ConversationTimelineItem[], id: string, label: string, detail = '', state: ExecutionState = 'running') {
  const step = items.find((item): item is Extract<ConversationTimelineItem, { type: 'execution' }> => item.type === 'execution' && item.id === id)
  if (step) {
    step.label = label
    step.detail = detail
    step.state = state
    return id
  }
  return appendExecutionStep(items, label, detail, state)
}
