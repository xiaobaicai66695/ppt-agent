import { useState, useEffect, useRef, useCallback } from 'react';

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

interface TaskItem {
  task_id: string;
  page_index: number;
  title: string;
  content_type: string;
  output_file: string;
  status: string;
  qa_report?: string;
  fix_attempts?: number;
}

interface TaskInfo {
  id: string;
  query: string;
  status: 'running' | 'completed' | 'failed';
  work_dir: string;
  created_at: string;
  done_count: number;
  total_count: number;
  duration?: string;
  error?: string;
  files?: string[];
}

interface SSEEvent {
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

// ---------------------------------------------------------------------------
// Status helpers
// ---------------------------------------------------------------------------

const STATUS_LABELS: Record<string, string> = {
  pending: '待生成',
  generating: '生成中',
  done: '已完成',
  qa_done: '已质检',
  fixed: '已修复',
  failed: '失败',
};

const STATUS_COLORS: Record<string, string> = {
  pending: '#94a3b8',
  generating: '#3b82f6',
  done: '#22c55e',
  qa_done: '#14b8a6',
  fixed: '#a855f7',
  failed: '#ef4444',
};

// ---------------------------------------------------------------------------
// App
// ---------------------------------------------------------------------------

export default function App() {
  const [tasks, setTasks] = useState<TaskInfo[]>([]);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [query, setQuery] = useState('');
  const [creating, setCreating] = useState(false);

  // Active task streaming state
  const [taskItems, setTaskItems] = useState<TaskItem[]>([]);
  const [doneCount, setDoneCount] = useState(0);
  const [totalCount, setTotalCount] = useState(0);
  const [logLines, setLogLines] = useState<{ ts: number; text: string; kind: string }[]>([]);
  const [finalFiles, setFinalFiles] = useState<string[]>([]);
  const [finalMessage, setFinalMessage] = useState('');
  const [duration, setDuration] = useState('');
  const logEndRef = useRef<HTMLDivElement>(null);
  const esRef = useRef<EventSource | null>(null);

  // Load task list on mount
  useEffect(() => {
    fetch('/api/tasks')
      .then(r => r.json())
      .then(data => {
        if (Array.isArray(data)) setTasks(data);
      })
      .catch(() => {});
  }, []);

  // Auto-scroll log
  useEffect(() => {
    logEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [logLines]);

  const selectedTask = tasks.find(t => t.id === selectedId);

  const addLog = useCallback((kind: string, text: string) => {
    setLogLines(prev => [...prev.slice(-300), { ts: Date.now(), text, kind }]);
  }, []);

  const handleCreate = async () => {
    const q = query.trim();
    if (!q || creating) return;
    setCreating(true);

    try {
      const res = await fetch('/api/tasks', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ query: q }),
      });
      const info: TaskInfo = await res.json();
      setTasks(prev => [info, ...prev]);
      setSelectedId(info.id);
      setQuery('');
      setTaskItems([]);
      setDoneCount(0);
      setTotalCount(0);
      setLogLines([]);
      setFinalFiles([]);
      setFinalMessage('');
      setDuration('');

      // Connect SSE
      connectSSE(info.id);
    } catch (err) {
      addLog('error', '创建任务失败: ' + (err as Error).message);
    } finally {
      setCreating(false);
    }
  };

  const connectSSE = (taskId: string) => {
    if (esRef.current) esRef.current.close();

    const es = new EventSource(`/api/tasks/${taskId}/stream`);
    esRef.current = es;

    const handleEvent = (e: MessageEvent) => {
      try {
        const evt: SSEEvent = JSON.parse(e.data);
        switch (evt.type) {
          case 'answer':
            addLog('answer', evt.content || '');
            break;
          case 'tool_call':
            addLog('tool', `🔧 ${evt.tool_name}(${evt.tool_args ? evt.tool_args.slice(0, 120) : ''})`);
            break;
          case 'progress':
            if (evt.tasks) setTaskItems(evt.tasks);
            if (evt.done !== undefined) setDoneCount(evt.done);
            if (evt.total !== undefined) setTotalCount(evt.total);
            break;
          case 'error':
            addLog('error', '❌ ' + (evt.error || evt.content || ''));
            break;
          case 'complete':
            setDoneCount(evt.done || 0);
            setTotalCount(evt.total || 0);
            if (evt.files) setFinalFiles(evt.files);
            if (evt.message) setFinalMessage(evt.message);
            if (evt.duration) setDuration(evt.duration);
            if (evt.tasks) setTaskItems(evt.tasks);
            es.close();
            // Refresh task info
            fetch(`/api/tasks/${taskId}`)
              .then(r => r.json())
              .then((info: TaskInfo) => {
                setTasks(prev => prev.map(t => (t.id === taskId ? info : t)));
              })
              .catch(() => {});
            break;
        }
      } catch { /* ignore parse errors */ }
    };

    es.addEventListener('answer', handleEvent);
    es.addEventListener('tool_call', handleEvent);
    es.addEventListener('progress', handleEvent);
    es.addEventListener('error', handleEvent);
    es.addEventListener('complete', handleEvent);

    es.onerror = () => {
      // SSE will auto-reconnect; if the task is done the server closes it.
    };
  };

  const handleSelectTask = (id: string) => {
    setSelectedId(id);
    const task = tasks.find(t => t.id === id);
    if (!task) return;

    // Reset display state
    setTaskItems([]);
    setDoneCount(task.done_count);
    setTotalCount(task.total_count);
    setLogLines([]);
    setFinalFiles(task.files || []);
    setFinalMessage('');
    setDuration(task.duration || '');

    if (task.status === 'running') {
      connectSSE(id);
    } else {
      // Load tasks.json for completed tasks
      fetch(`/api/tasks/${id}`)
        .then(r => r.json())
        .then((info: TaskInfo) => {
          setDoneCount(info.done_count);
          setTotalCount(info.total_count);
          setFinalFiles(info.files || []);
          setDuration(info.duration || '');
        })
        .catch(() => {});
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleCreate();
    }
  };

  const progressPct = totalCount > 0 ? Math.round((doneCount / totalCount) * 100) : 0;

  return (
    <div className="layout">
      {/* Sidebar */}
      <aside className="sidebar">
        <div className="sidebar-header">
          <h1 className="sidebar-logo">📊 PPT Agent</h1>
          <span className="sidebar-sub">AI 幻灯片生成</span>
        </div>

        {/* Create form */}
        <div className="create-form">
          <textarea
            className="create-input"
            placeholder="描述你的 PPT 需求..."
            value={query}
            onChange={e => setQuery(e.target.value)}
            onKeyDown={handleKeyDown}
            rows={2}
          />
          <button className="create-btn" onClick={handleCreate} disabled={creating}>
            {creating ? '创建中...' : '生成 PPT'}
          </button>
        </div>

        {/* Task history */}
        <div className="task-list">
          <h3 className="task-list-title">任务历史</h3>
          {tasks.length === 0 && <p className="empty-hint">暂无任务</p>}
          {tasks.map(t => (
            <div
              key={t.id}
              className={`task-item ${t.id === selectedId ? 'active' : ''}`}
              onClick={() => handleSelectTask(t.id)}
            >
              <div className="task-item-top">
                <span className="task-item-query">{t.query.slice(0, 40)}{t.query.length > 40 ? '...' : ''}</span>
              </div>
              <div className="task-item-meta">
                <span className="task-badge" style={{ background: t.status === 'running' ? '#3b82f6' : t.status === 'completed' ? '#22c55e' : '#ef4444' }}>
                  {t.status === 'running' ? '运行中' : t.status === 'completed' ? '已完成' : '失败'}
                </span>
                <span className="task-item-time">{fmtTime(t.created_at)}</span>
              </div>
              {t.total_count > 0 && (
                <div className="task-item-progress">
                  <div className="mini-bar"><div className="mini-bar-fill" style={{ width: `${Math.round((t.done_count / t.total_count) * 100)}%` }} /></div>
                  <span className="mini-count">{t.done_count}/{t.total_count}</span>
                </div>
              )}
            </div>
          ))}
        </div>
      </aside>

      {/* Main */}
      <main className="main">
        {!selectedTask ? (
          <div className="welcome">
            <div className="welcome-icon">📊</div>
            <h2>PPT Agent</h2>
            <p>在左侧输入需求，开始生成 PPT</p>
            <div className="welcome-hints">
              <div className="hint-card">
                <strong>并行生成</strong>
                <span>多个幻灯片同时生成，大幅提升效率</span>
              </div>
              <div className="hint-card">
                <strong>自动质检</strong>
                <span>生成后自动进行视觉质量检查</span>
              </div>
              <div className="hint-card">
                <strong>实时追踪</strong>
                <span>每页状态实时可见，进度一目了然</span>
              </div>
            </div>
          </div>
        ) : (
          <div className="dashboard">
            {/* Header */}
            <div className="dash-header">
              <div>
                <h2 className="dash-title">{selectedTask.query}</h2>
                <span className="dash-id">ID: {selectedTask.id.slice(0, 8)}...</span>
              </div>
              <div className="dash-stats">
                {selectedTask.status === 'running' && <span className="stat-badge running">● 运行中</span>}
                {selectedTask.status === 'completed' && <span className="stat-badge done">● 已完成</span>}
                {selectedTask.status === 'failed' && <span className="stat-badge failed">● 失败</span>}
                {duration && <span className="stat-duration">⏱ {duration}</span>}
              </div>
            </div>

            {/* Progress */}
            {totalCount > 0 && (
              <div className="progress-section">
                <div className="progress-header">
                  <span>进度</span>
                  <span className="progress-num">{doneCount} / {totalCount} 页</span>
                </div>
                <div className="progress-track">
                  <div className="progress-fill" style={{ width: `${progressPct}%` }} />
                </div>
              </div>
            )}

            {/* Task cards grid */}
            {taskItems.length > 0 && (
              <div className="cards-section">
                <h3 className="section-title">页面状态</h3>
                <div className="cards-grid">
                  {taskItems.map(item => (
                    <div key={item.task_id} className="task-card">
                      <div className="card-status" style={{ background: STATUS_COLORS[item.status] || '#94a3b8' }}>
                        {item.page_index}
                      </div>
                      <div className="card-body">
                        <div className="card-title">{item.title}</div>
                        <div className="card-meta">
                          <span className="card-type">{item.content_type}</span>
                          <span className="card-status-label" style={{ color: STATUS_COLORS[item.status] || '#94a3b8' }}>
                            {STATUS_LABELS[item.status] || item.status}
                          </span>
                        </div>
                        {item.output_file && (
                          <div className="card-file">{item.output_file}</div>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Event log */}
            <div className="log-section">
              <h3 className="section-title">事件日志 {logLines.length > 0 && `(${logLines.length})`}</h3>
              <div className="log-box">
                {logLines.length === 0 && <p className="empty-hint">等待事件...</p>}
                {logLines.map((l, i) => (
                  <div key={i} className={`log-line ${l.kind}`}>
                    <span className="log-ts">{new Date(l.ts).toLocaleTimeString()}</span>
                    <span className="log-text">{l.text}</span>
                  </div>
                ))}
                <div ref={logEndRef} />
              </div>
            </div>

            {/* Final message */}
            {finalMessage && (
              <div className="final-message">
                <h3>生成结果</h3>
                <p>{finalMessage}</p>
              </div>
            )}

            {/* File downloads */}
            {finalFiles.length > 0 && (
              <div className="files-section">
                <h3 className="section-title">生成文件 ({finalFiles.length})</h3>
                <div className="files-grid">
                  {finalFiles.map(f => {
                    const name = f.split(/[/\\]/).pop() || f;
                    return (
                      <a
                        key={f}
                        className="file-card"
                        href={`/api/tasks/${selectedTask.id}/files/${encodeURIComponent(name)}`}
                        download
                      >
                        <span className="file-icon">📄</span>
                        <span className="file-name">{name}</span>
                      </a>
                    );
                  })}
                </div>
              </div>
            )}
          </div>
        )}
      </main>
    </div>
  );
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function fmtTime(iso: string): string {
  if (!iso) return '';
  const d = new Date(iso);
  const pad = (n: number) => String(n).padStart(2, '0');
  return `${pad(d.getHours())}:${pad(d.getMinutes())}`;
}
