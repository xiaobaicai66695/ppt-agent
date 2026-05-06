// ── Task types ──────────────────────────────────────────────────────────────

export interface TaskItem {
  task_id: string;
  page_index: number;
  title: string;
  content_type: string;
  output_file: string;
  status: string;
  qa_report?: string;
  fix_attempts?: number;
}

export interface TaskInfo {
  id: string;
  user_id?: number;
  query: string;
  status: 'running' | 'completed' | 'failed' | 'cancelled';
  work_dir: string;
  created_at: string;
  done_count: number;
  total_count: number;
  duration?: string;
  error?: string;
  files?: string[];
}

export interface SSEEvent {
  type: string;
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
  status?: string;
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
  kind: string;
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
