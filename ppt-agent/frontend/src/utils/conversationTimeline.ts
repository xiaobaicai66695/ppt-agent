import type { ConversationMessage } from '../types'

export type ExecutionState = 'running' | 'success' | 'error'
export type TimelineEventID = string | number

export type ToolPreview = {
  images?: Array<{
    thumbnail_url?: string
    image_url?: string
    source_url?: string
    alt?: string
    attribution?: string
  }>
}

type TimelineBase = { id: string; eventID?: TimelineEventID; eventIDs?: TimelineEventID[] }

export type ConversationTimelineItem =
  | (TimelineBase & { type: 'message'; message: ConversationMessage })
  | (TimelineBase & { type: 'thought'; segmentID: string; content: string; phase?: string; state: ExecutionState; expanded: boolean; streaming: boolean })
  | (TimelineBase & { type: 'tool_call'; callID: string; name: string; label: string; args?: string; detail?: string; result?: string; state: ExecutionState; expanded: boolean; preview?: ToolPreview })
  | (TimelineBase & { type: 'final_answer'; segmentID: string; content: string; streaming: boolean })
  | (TimelineBase & { type: 'execution'; label: string; detail?: string; state: ExecutionState })
  | (TimelineBase & { type: 'error'; content: string })

let ordinal = 0

function nextID(prefix: string) {
  ordinal += 1
  return `${prefix}-${ordinal}`
}

function eventItemID(type: string, eventID?: TimelineEventID) {
  return eventID === undefined || eventID === '' ? nextID(type) : `event-${eventID}-${type}`
}

function eventAlreadyRendered(items: ConversationTimelineItem[], type: ConversationTimelineItem['type'], eventID?: TimelineEventID) {
  if (eventID === undefined || eventID === '') return false
  return items.some(item => item.type === type && (String(item.eventID) === String(eventID) || item.eventIDs?.some(id => String(id) === String(eventID))))
}

function rememberEventID(item: TimelineBase, eventID?: TimelineEventID) {
  if (eventID === undefined || eventID === '') return
  item.eventID = eventID
  if (!item.eventIDs?.some(id => String(id) === String(eventID))) {
    item.eventIDs = [...(item.eventIDs || []), eventID]
  }
}

function appendDelta(current: string, incoming: string) {
  if (!incoming) return current
  if (!current) return incoming
  if (incoming === current) return current
  if (incoming.startsWith(current)) return incoming
  if (current.endsWith(incoming)) return current
  return current + incoming
}

function isPlaceholderToolResult(result?: string) {
  const normalized = result?.trim()
  return !normalized || normalized === '工具调用已完成' || normalized === '工具调用失败' || normalized === '工具调用未返回结果' || normalized === '工具调用未完成' || normalized === '工具结果尚未返回'
}

function findToolCallForResult(
  items: ConversationTimelineItem[],
  payload: { callID?: string; name?: string; args?: string },
) {
  const callID = payload.callID?.trim()
  const name = payload.name?.trim()
  const args = payload.args?.trim()
  if (callID) {
    const byID = [...items]
      .reverse()
      .find((item): item is Extract<ConversationTimelineItem, { type: 'tool_call' }> => item.type === 'tool_call' && item.callID === callID)
    if (byID) return byID
  }
  const candidates = [...items]
    .reverse()
    .filter((item): item is Extract<ConversationTimelineItem, { type: 'tool_call' }> => item.type === 'tool_call' && (!name || item.name === name))
  const unresolved = candidates.find(item => item.state === 'running' || isPlaceholderToolResult(item.result))
  if (unresolved) {
    if (callID) unresolved.callID = callID
    return unresolved
  }
  if (args) {
    const sameArgs = candidates.find(item => item.args?.trim() === args)
    if (sameArgs && callID) sameArgs.callID = callID
    if (sameArgs) return sameArgs
  }
  return undefined
}

export function resetConversationTimeline(messages: ConversationMessage[]): ConversationTimelineItem[] {
  ordinal = 0
  return messages.map(message => ({ id: nextID('message'), type: 'message' as const, message }))
}

export function appendTimelineMessage(items: ConversationTimelineItem[], message: ConversationMessage) {
  items.push({ id: nextID('message'), type: 'message', message })
}

export function appendThought(
  items: ConversationTimelineItem[],
  content: string,
  options: { eventID?: TimelineEventID; segmentID?: string; phase?: string; delta?: boolean } = {},
) {
  if (!content) return undefined
  if (eventAlreadyRendered(items, 'thought', options.eventID)) return undefined
  const segmentID = options.segmentID || `thought-${options.eventID || nextID('segment')}`
  const current = [...items]
    .reverse()
    .find((item): item is Extract<ConversationTimelineItem, { type: 'thought' }> => item.type === 'thought' && item.segmentID === segmentID && item.streaming)
  if (current) {
    current.content = options.delta === false ? content : appendDelta(current.content, content)
    current.streaming = options.delta !== false
    rememberEventID(current, options.eventID)
    return current.id
  }
  const id = eventItemID('thought', options.eventID)
  items.push({ id, eventID: options.eventID, eventIDs: options.eventID === undefined ? undefined : [options.eventID], type: 'thought', segmentID, content, phase: options.phase, state: 'running', expanded: true, streaming: options.delta === true })
  return id
}

export function appendToolCall(
  items: ConversationTimelineItem[],
  payload: { eventID?: TimelineEventID; callID?: string; name: string; label: string; args?: string; detail?: string },
) {
  if (eventAlreadyRendered(items, 'tool_call', payload.eventID)) return undefined
  const callID = payload.callID || `call-${payload.eventID || nextID('call')}`
  const existing = items.find((item): item is Extract<ConversationTimelineItem, { type: 'tool_call' }> => item.type === 'tool_call' && item.callID === callID)
  if (existing) {
    existing.name = payload.name || existing.name
    existing.label = payload.label || existing.label
    if (payload.args && !existing.args) existing.args = payload.args
    if (payload.detail && !existing.detail) existing.detail = payload.detail
    rememberEventID(existing, payload.eventID)
    return existing.id
  }
  const activeThought = [...items].reverse().find((item): item is Extract<ConversationTimelineItem, { type: 'thought' }> => item.type === 'thought' && item.state === 'running')
  if (activeThought) {
    activeThought.state = 'success'
    activeThought.streaming = false
  }
  const id = eventItemID('tool-call', payload.eventID)
  items.push({ id, eventID: payload.eventID, eventIDs: payload.eventID === undefined ? undefined : [payload.eventID], type: 'tool_call', callID, name: payload.name, label: payload.label, args: payload.args, detail: payload.detail, state: 'running', expanded: true })
  return id
}

export function appendToolResult(
  items: ConversationTimelineItem[],
  payload: { eventID?: TimelineEventID; callID?: string; name: string; label: string; args?: string; result?: string; state?: ExecutionState; preview?: ToolPreview },
) {
  if (eventAlreadyRendered(items, 'tool_call', payload.eventID)) return undefined
  const callID = payload.callID || `call-${payload.eventID || nextID('call')}`
  const call = findToolCallForResult(items, { callID: payload.callID, name: payload.name, args: payload.args })
  const result = payload.result || (payload.state === 'error' ? '工具调用失败' : '工具调用已完成')
  if (call) {
    if (payload.state) call.state = payload.state
    else if (result) call.state = 'success'
    if (result && (isPlaceholderToolResult(call.result) || (call.result?.length ?? 0) < result.length)) {
      call.result = result
    }
    if (payload.preview && !call.preview) call.preview = payload.preview
    if (payload.eventID !== undefined) rememberEventID(call, payload.eventID)
    return call.id
  }
  const id = eventItemID('tool-call', payload.eventID)
  items.push({
    id,
    eventID: payload.eventID,
    eventIDs: payload.eventID === undefined ? undefined : [payload.eventID],
    type: 'tool_call',
    callID,
    name: payload.name || 'unknown',
    label: payload.label || '工具调用',
    result,
    state: payload.state || 'success',
    preview: payload.preview,
    expanded: true,
  })
  return id
}

export function appendFinalAnswer(
  items: ConversationTimelineItem[],
  content: string,
  options: { eventID?: TimelineEventID; segmentID?: string; delta?: boolean } = {},
) {
  if (!content) return undefined
  if (eventAlreadyRendered(items, 'final_answer', options.eventID)) return undefined
  const activeThought = [...items].reverse().find((item): item is Extract<ConversationTimelineItem, { type: 'thought' }> => item.type === 'thought' && item.state === 'running')
  if (activeThought) {
    activeThought.state = 'success'
    activeThought.streaming = false
  }
  const segmentID = options.segmentID || 'legacy-final-answer'
  const current = [...items]
    .reverse()
    .find((item): item is Extract<ConversationTimelineItem, { type: 'final_answer' }> => item.type === 'final_answer' && item.segmentID === segmentID && item.streaming)
  if (current) {
    current.content = options.delta === false ? content : appendDelta(current.content, content)
    current.streaming = options.delta !== false
    rememberEventID(current, options.eventID)
    return current.id
  }
  const id = eventItemID('final-answer', options.eventID)
  items.push({ id, eventID: options.eventID, eventIDs: options.eventID === undefined ? undefined : [options.eventID], type: 'final_answer', segmentID, content, streaming: options.delta !== false })
  return id
}

export function finishStreamingEntries(items: ConversationTimelineItem[]) {
  finishTimelineEntries(items)
}

export function finishTimelineEntries(items: ConversationTimelineItem[], options: { includeTools?: boolean } = {}) {
  const includeTools = options.includeTools ?? true
  for (const item of items) {
    if (item.type === 'thought') {
      item.streaming = false
      if (item.state === 'running') item.state = 'success'
    }
    if (includeTools && item.type === 'tool_call' && item.state === 'running') {
      item.state = item.result ? 'success' : 'error'
      if (!item.result) item.result = '工具调用未返回结果'
    }
    if (item.type === 'final_answer') item.streaming = false
  }
}

export function appendTimelineError(items: ConversationTimelineItem[], content: string, eventID?: TimelineEventID) {
  if (!content || eventAlreadyRendered(items, 'error', eventID)) return undefined
  const id = eventItemID('error', eventID)
  items.push({ id, eventID, eventIDs: eventID === undefined ? undefined : [eventID], type: 'error', content })
  return id
}

export function toggleTimelineItem(items: ConversationTimelineItem[], itemID: string) {
  const item = items.find(candidate => candidate.id === itemID)
  if (item?.type === 'thought' || item?.type === 'tool_call') {
    item.expanded = !item.expanded
    return item.expanded
  }
  return undefined
}

export function appendExecutionStep(items: ConversationTimelineItem[], label: string, detail = '', state: ExecutionState = 'running', eventID?: TimelineEventID) {
  if (eventAlreadyRendered(items, 'execution', eventID)) return undefined
  const id = eventItemID('execution', eventID)
  items.push({ id, eventID, eventIDs: eventID === undefined ? undefined : [eventID], type: 'execution', label, detail, state })
  return id
}
