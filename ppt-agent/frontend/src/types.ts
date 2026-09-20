export type TaskStatus = 'conversation' | 'running' | 'completed' | 'paused_retryable' | 'failed' | 'cancelled'

export interface AuthUser {
  id: number
  email: string
  token?: string
  is_new?: boolean
  is_admin?: boolean
  is_guest?: boolean
}

export interface TaskInfo {
  id: string
  query: string
  status: TaskStatus
  created_at: string
  updated_at?: string
  done_count: number
  total_count: number
  duration?: string
  error?: string
  files?: string[]
  total_tokens?: number
  feedback?: DeliveryFeedback
}

export interface DeliveryFeedback { rating: number; suggestion?: string; updated_at: string }

export interface ConversationMessage {
  role: 'user' | 'assistant'
  content: string
  timestamp: string
}

export interface ConversationSession {
  task_id: string
  latest_event_id?: number
  replay_after_event_id?: number
  conversation_streaming?: boolean
  messages: ConversationMessage[]
  timeline?: ConversationTimelineEvent[]
  status?: TaskStatus
  done_count?: number
  total_count?: number
  files?: string[]
  error?: string
}

export interface RuntimeEvent {
  id: number
  timestamp: string
  kind: string
  phase?: string
  name?: string
  status?: string
  detail?: string
  metadata?: Record<string, unknown>
}

export type TaskStreamEventType =
  | 'llm_start'
  | 'llm_delta'
  | 'llm_end'
  | 'thought'
  | 'tool_call'
  | 'tool_result'
  | 'final_answer'
  | 'answer'
  | 'answer_end'
  | 'system_step'
  | 'progress'
  | 'runtime_event'
  | 'file_ready'
  | 'thumbnail_ready'
  | 'error'
  | 'complete'
  | 'continue_complete'
  | 'continue_queued'
  | 'conversation_complete'

export interface TaskStreamEvent {
  id?: number
  type?: TaskStreamEventType
  segment_id?: string
  segment_boundary?: boolean
  delta?: boolean
  content?: string
  error?: string
  message?: string
  phase?: string
  phase_detail?: string
  tool_call_id?: string
  tool_name?: string
  tool_args?: string
  tool_result?: string
  tool_status?: 'success' | 'error' | 'unverified'
  files?: string[]
  status?: TaskStatus
  runtime_event?: RuntimeEvent
  tool_preview?: {
    images?: Array<{ thumbnail_url?: string; image_url?: string; source_url?: string; alt?: string; attribution?: string }>
  }
}

export type ConversationTimelineEvent =
  | { type: 'message'; message: ConversationMessage }
  | TaskStreamEvent

export interface AtomicLayout {
  name: string
  display_name: string
  type: string
  description: string
  fields: Array<{ name: string; label: string; type: string; required: boolean; placeholder?: string }>
  contract?: { best_for?: string[]; capacity?: Record<string, string | number | boolean> }
}

export interface SlideOutline { title: string; content_type: string; content_plan?: Record<string, unknown> }
export interface TaskOutline { title: string; content_mode?: 'user_outline'; slides: SlideOutline[] }
