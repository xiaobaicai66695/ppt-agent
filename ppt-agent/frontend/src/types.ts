// ── Status types ────────────────────────────────────────────────────────────

export type TaskItemStatus = 'pending' | 'generating' | 'done' | 'qa_done' | 'fixed' | 'failed';
export type TaskStatus = 'running' | 'completed' | 'failed' | 'cancelled';
export type SSEEventType = 'answer' | 'tool_call' | 'progress' | 'file_ready' | 'token_usage' | 'error' | 'complete' | 'continue_complete' | 'continue_queued';
export type LogKind = 'answer' | 'tool' | 'worker' | 'file' | 'error' | 'divider';

// ── Session types ────────────────────────────────────────────────────────────

export interface ConversationMessage {
  role: 'user' | 'assistant';
  content: string;
  timestamp: string;
}

export interface ConversationSession {
  task_id: string;
  messages: ConversationMessage[];
  // 冷启动时从 task_records 重建的完整拼接内容。
  conversation_content?: string;
  status?: TaskStatus;
  done_count?: number;
  total_count?: number;
  files?: string[];
  duration?: string;
  prompt_tokens?: number;
  completion_tokens?: number;
  total_tokens?: number;
  created_at: string;
  updated_at: string;
}

// ── User profile types ──────────────────────────────────────────────────────

export interface ContentTypeCount {
  [key: string]: number;
}

export interface UserStyleProfile {
  user_id: number;
  preferred_themes: string[];
  preferred_colors: string[];
  content_patterns: string[];
  layout_preferences: string[];
  language_tone: string;
  typical_page_count: number;
  content_types: ContentTypeCount;
  special_notes: string[];
  task_count: number;
  updated_at: string;
}

// ── Task types ──────────────────────────────────────────────────────────────

export interface TaskItem {
  task_id: string;
  page_index: number;
  title: string;
  content_type: string;
  output_file: string;
  status: TaskItemStatus;
  qa_report?: string;
  fix_attempts?: number;
}

export interface TaskInfo {
  id: string;
  user_id?: number;
  query: string;
  status: TaskStatus;
  work_dir: string;
  created_at: string;
  done_count: number;
  total_count: number;
  duration?: string;
  error?: string;
  files?: string[];
  prompt_tokens?: number;
  completion_tokens?: number;
  total_tokens?: number;
}

export interface SSEEvent {
  type: SSEEventType;
  content?: string;
  tool_name?: string;
  tool_args?: string;
  error?: string;
  tasks?: TaskItem[];
  done?: number;
  total?: number;
  files?: string[];
  message?: string;
  duration?: string;
  status?: TaskStatus;
  prompt_tokens?: number;
  completion_tokens?: number;
  total_tokens?: number;
}

// ── Batch tracking ──────────────────────────────────────────────────────────

export interface Batch {
  id: number;
  taskIds: string[];
  ts: number;
  done: boolean;
}

// ── Log ─────────────────────────────────────────────────────────────────────

export interface LogLine {
  ts: number;
  text: string;
  kind: LogKind;
}

// ── Status helpers ──────────────────────────────────────────────────────────

export const STATUS_LABELS: Record<string, string> = {
  pending: '待生成',
  generating: '生成中',
  done: '已完成',
  qa_done: '已质检',
  fixed: '已修复',
  failed: '失败',
};

export const STATUS_COLORS: Record<string, string> = {
  pending: '#94a3b8',
  generating: '#3b82f6',
  done: '#22c55e',
  qa_done: '#14b8a6',
  fixed: '#a855f7',
  failed: '#ef4444',
};
