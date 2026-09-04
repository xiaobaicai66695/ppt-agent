const terminalStreamEvents = new Set([
  'complete',
  'continue_complete',
  'conversation_complete',
])

// `answer_end` only terminates the planner's text response. A PPT task keeps
// streaming while its DeckSpec is rendered and delivery files are prepared.
export function isTerminalTaskStreamEvent(type?: string) {
  return Boolean(type && terminalStreamEvents.has(type))
}
