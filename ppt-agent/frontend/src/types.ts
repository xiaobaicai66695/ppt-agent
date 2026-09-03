export type TaskStatus = 'conversation' | 'running' | 'completed' | 'failed' | 'cancelled'

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
  full_answer?: string
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
