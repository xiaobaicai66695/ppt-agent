import type { ConversationMessage } from '../types'

export type ExecutionState = 'running' | 'success' | 'error'

export type ToolInvocation = {
  id: string
  name: string
  label: string
  callDetail?: string
  resultDetail?: string
  preview?: ToolPreview
  state: ExecutionState
  expanded: boolean
}

export type ToolPreview = { images?: Array<{ thumbnail_url?: string; image_url?: string; source_url?: string; alt?: string; attribution?: string }> }

export type ConversationTimelineItem =
  | { id: string; type: 'message'; message: ConversationMessage }
  | { id: string; type: 'phase'; phase: string; label: string; detail?: string; state: ExecutionState; tools: ToolInvocation[] }
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

export function beginObservablePhase(items: ConversationTimelineItem[], phase: string, label: string, detail = '') {
  const active = [...items].reverse().find((item): item is Extract<ConversationTimelineItem, { type: 'phase' }> => item.type === 'phase' && item.state === 'running')
  if (active?.phase === phase) {
    active.label = label
    active.detail = detail
    return active.id
  }
  if (active) active.state = 'success'
  const id = nextID('phase')
  items.push({ id, type: 'phase', phase, label, detail, state: 'running', tools: [] })
  return id
}

export function completeObservablePhase(items: ConversationTimelineItem[], phase: string, state: ExecutionState = 'success') {
  const item = [...items].reverse().find((candidate): candidate is Extract<ConversationTimelineItem, { type: 'phase' }> => candidate.type === 'phase' && candidate.phase === phase)
  if (item) item.state = state
}

function findPhase(items: ConversationTimelineItem[], phaseID?: string) {
  if (phaseID) return items.find((item): item is Extract<ConversationTimelineItem, { type: 'phase' }> => item.type === 'phase' && item.id === phaseID)
  return [...items].reverse().find((item): item is Extract<ConversationTimelineItem, { type: 'phase' }> => item.type === 'phase')
}

export function appendToolInvocation(items: ConversationTimelineItem[], phaseID: string | undefined, name: string, label: string, detail = '', preview?: ToolPreview) {
  const phase = findPhase(items, phaseID)
  if (!phase) return undefined
  const tool: ToolInvocation = { id: nextID('tool'), name, label, callDetail: detail, preview, state: 'running', expanded: true }
  phase.tools.push(tool)
  return tool.id
}

export function resolveToolInvocation(items: ConversationTimelineItem[], phaseID: string | undefined, name: string, label: string, detail = '', state: ExecutionState = 'success', preview?: ToolPreview) {
  const phase = findPhase(items, phaseID)
  if (!phase) return undefined
  const tool = [...phase.tools].reverse().find(candidate => candidate.name === name && candidate.state === 'running')
  if (tool) {
    tool.label = label
    tool.resultDetail = detail
    tool.preview = preview || tool.preview
    tool.state = state
    return tool.id
  }
  const id = appendToolInvocation(items, phase.id, name, label, '', preview)
  const recovered = phase.tools.find(candidate => candidate.id === id)
  if (recovered) {
    recovered.resultDetail = detail
    recovered.state = state
  }
  return id
}

// Tool calls are transient observability, not durable conversation history.
// A terminal task keeps user/assistant text but drops every tool-bearing phase.
export function hideCompletedToolTraces(items: ConversationTimelineItem[]) {
  for (let index = items.length - 1; index >= 0; index -= 1) {
    const item = items[index]
    if (item?.type === 'phase' && item.tools.length > 0) items.splice(index, 1)
  }
}

export function toggleToolInvocation(items: ConversationTimelineItem[], toolID: string) {
  for (const item of items) {
    if (item.type !== 'phase') continue
    const tool = item.tools.find(candidate => candidate.id === toolID)
    if (tool) {
      tool.expanded = !tool.expanded
      return tool.expanded
    }
  }
  return undefined
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
