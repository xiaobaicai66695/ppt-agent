export const taskStreamEventNames = [
  'llm_start',
  'llm_delta',
  'llm_end',
  'thought',
  'tool_call',
  'tool_result',
  'final_answer',
  'answer',
  'answer_end',
  'system_step',
  'progress',
  'file_ready',
  'thumbnail_ready',
  'error',
  'complete',
  'continue_complete',
  'continue_queued',
  'conversation_complete',
] as const

const terminalStreamEvents = new Set([
  'complete',
  'continue_complete',
  'conversation_complete',
])

// `answer_end` only terminates the planner's text response. A PPT task keeps
// streaming while its PPTSpec is rendered and delivery files are prepared.
export function isTerminalTaskStreamEvent(type?: string) {
  return Boolean(type && terminalStreamEvents.has(type))
}
