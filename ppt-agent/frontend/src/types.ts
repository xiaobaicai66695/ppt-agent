// ── Status types ────────────────────────────────────────────────────────────

export type TaskItemStatus = 'pending' | 'generating' | 'done' | 'qa_done' | 'fixed' | 'failed';
export type TaskStatus = 'running' | 'completed' | 'failed' | 'cancelled';
export type SSEEventType =
  | 'answer'
  | 'answer_end'
  | 'system_step'
  | 'system_step_end'
  | 'tool_call'
  | 'progress'
  | 'file_ready'
  | 'thumbnail_ready'
  | 'thumbnail_error'
  | 'token_usage'
  | 'runtime_event'
  | 'error'
  | 'complete'
  | 'continue_complete'
  | 'continue_queued';
export type LogKind = 'answer' | 'tool' | 'worker' | 'file' | 'error' | 'divider';

// ── Session types ────────────────────────────────────────────────────────────

export interface ConversationMessage {
  role: 'user' | 'assistant';
  content: string;
  timestamp: string;
  runtime_event_id?: number;
  timeline_order?: number;
}

export interface ConversationSession {
  task_id: string;
  latest_event_id?: number;
  replay_after_event_id?: number;
  messages: ConversationMessage[];
  // 冷启动时从 task_records 重建的完整拼接内容。
  conversation_content?: string;
  // 完整累积的 LLM 回答（任务结束时一次性写入 DB），优先用于冷加载恢复。
  full_answer?: string;
  status?: TaskStatus;
  done_count?: number;
  total_count?: number;
  files?: string[];
  duration?: string;
  prompt_tokens?: number;
  completion_tokens?: number;
  total_tokens?: number;
  runtime_meta?: RuntimeMeta;
  created_at: string;
  updated_at: string;
}

export interface LiveActivity {
  label: string;
  detail?: string;
  state: 'idle' | 'running' | 'success' | 'error';
}

// ── Task types ──────────────────────────────────────────────────────────────

export interface TaskItem {
  task_id: string;
  page_index: number;
  title: string;
  content_type: string;
  output_file: string;
  status: TaskItemStatus;
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

export interface RuntimeBudgets {
  same_tool_args_warn?: number;
  max_tool_calls_per_tool?: number;
  max_total_tool_calls?: number;
  token_warn?: number;
  phase_duration_warn_sec?: number;
}

export interface TaskInputAnchor {
  summary?: string;
  original_length?: number;
  template?: string;
  theme?: string;
  recommendation?: string;
}

export interface PlanSlide {
  page_index?: number;
  task_id?: string;
  title?: string;
  content_type?: string;
  output_file?: string;
  status?: string;
}

export interface AlignmentWarning {
  code: string;
  step: string;
  severity: string;
  message: string;
  page_index?: number;
  expected?: string;
  observed?: string;
}

export interface RuntimeEvent {
  id: number;
  task_id?: string;
  timestamp: string;
  elapsed_ms: number;
  kind: string;
  phase?: string;
  name?: string;
  status?: string;
  detail?: string;
  metadata?: Record<string, unknown>;
  metadata_loaded?: boolean;
}

export interface RuntimeMeta {
  task_id?: string;
  work_dir?: string;
  elapsed_ms: number;
  phase?: string;
  phase_detail?: string;
  last_error?: string;
  last_tool?: string;
  tool_calls?: Record<string, number>;
  tool_errors?: Record<string, number>;
  same_tool_args_streak?: number;
  prompt_tokens?: number;
  completion_tokens?: number;
  total_tokens?: number;
  compression_before_tokens?: number;
  compression_after_tokens?: number;
  compression_saved_pct?: string;
  compression_before_messages?: number;
  compression_after_messages?: number;
  compression_removed_messages?: number;
  compression_saved_tokens?: number;
  budgets?: RuntimeBudgets;
  budget_warnings?: string[];
  done_slides?: number;
  total_slides?: number;
  missing_files?: number;
  qa_high_issues?: number;
  qa_medium_issues?: number;
  qa_low_issues?: number;
  task_input?: TaskInputAnchor;
  plan_slides?: PlanSlide[];
  current_slide?: PlanSlide;
  alignment_status?: string;
  alignment_warnings?: AlignmentWarning[];
  event_counts?: Record<string, number>;
  recent_events?: RuntimeEvent[];
}

export interface SSEEvent {
  id?: number;
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
  phase?: string;       // 当前阶段: preparing/planning/generating/complete
  phase_detail?: string; // 阶段详情
  runtime_event?: RuntimeEvent;
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
