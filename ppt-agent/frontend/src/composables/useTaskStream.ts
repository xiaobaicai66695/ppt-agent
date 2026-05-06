import { ref, nextTick, type Ref } from 'vue';
import type { TaskItem, TaskInfo, SSEEvent, Batch, LogLine } from '../types';
import { fetchTask } from '../api';

export function useTaskStream() {
  const taskItems = ref<TaskItem[]>([]);
  const doneCount = ref(0);
  const totalCount = ref(0);
  const logLines = ref<LogLine[]>([]);
  const finalFiles = ref<string[]>([]);
  const finalMessage = ref('');
  const duration = ref('');
  const activeWorkers = ref(0);
  const lastProgressTs = ref(0);
  const batches = ref<Batch[]>([]);
  let batchIdSeq = 0;

  let es: EventSource | null = null;

  function addLog(kind: string, text: string, logBox?: HTMLElement | null) {
    if (kind === 'answer' && logLines.value.length > 0) {
      const last = logLines.value[logLines.value.length - 1];
      if (last.kind === 'answer') {
        last.text += text;
        last.ts = Date.now();
        if (logBox) {
          nextTick(() => {
            logBox.scrollTop = logBox.scrollHeight;
          });
        }
        return;
      }
    }
    logLines.value = [...logLines.value.slice(-500), { ts: Date.now(), text, kind }];
    if (logBox) {
      nextTick(() => {
        logBox.scrollTop = logBox.scrollHeight;
      });
    }
  }

  function parseTaskIds(args: string): string[] {
    try {
      const obj = JSON.parse(args);
      if (Array.isArray(obj.task_ids)) return obj.task_ids.map(String);
      if (typeof obj.task_id === 'string') return obj.task_id.split(',').map((s: string) => s.trim());
      if (typeof obj.task_ids === 'string') return obj.task_ids.split(',').map((s: string) => s.trim());
      return [];
    } catch {
      return [];
    }
  }

  function markGenerating(ids: string[]) {
    for (const item of taskItems.value) {
      if (ids.includes(item.task_id) && item.status === 'pending') {
        item.status = 'generating';
      }
    }
  }

  function connectSSE(taskId: string, logBox?: HTMLElement | null) {
    if (es) es.close();
    es = new EventSource(`/api/tasks/${taskId}/stream`);
    activeWorkers.value = 0;

    const handler = (e: MessageEvent) => {
      let evt: SSEEvent;
      try { evt = JSON.parse(e.data); } catch { return; }

      switch (evt.type) {
        case 'answer':
          if (logLines.value.length > 0 && logLines.value[logLines.value.length - 1].kind !== 'answer') {
            addLog('divider', '── AI 响应 ──', logBox);
          }
          addLog('answer', evt.content || '', logBox);
          break;

        case 'tool_call': {
          const name = evt.tool_name || '';
          const args = evt.tool_args || '';

          if (name === 'task' && args.includes('SlideExecutor')) {
            const ids = parseTaskIds(args);
            markGenerating(ids);
            activeWorkers.value = Math.max(1, ids.length || 1);
            batches.value = [...batches.value, { id: ++batchIdSeq, taskIds: ids, ts: Date.now(), done: false }];
            addLog('worker', `⚡ 派发 ${ids.length} 个子任务并行执行`, logBox);
          } else if (name === 'task') {
            const ids = parseTaskIds(args);
            markGenerating(ids);
            activeWorkers.value = Math.max(activeWorkers.value, ids.length || 1);
            addLog('tool', `▶ ${name} (${args.slice(0, 100)})`, logBox);
          } else {
            addLog('tool', `▶ ${name} (${args.slice(0, 120)})`, logBox);
          }
          lastProgressTs.value = Date.now();
          break;
        }

        case 'progress':
          if (evt.tasks) {
            taskItems.value = evt.tasks;
            for (const batch of batches.value) {
              if (batch.done) continue;
              const allDone = batch.taskIds.every(tid => {
                const item = evt.tasks!.find(t => t.task_id === tid);
                return item && (item.status === 'done' || item.status === 'qa_done' || item.status === 'fixed');
              });
              if (allDone) batch.done = true;
            }
            for (const t of evt.tasks) {
              if ((t.status === 'done' || t.status === 'qa_done' || t.status === 'fixed') && t.output_file) {
                if (!finalFiles.value.includes(t.output_file)) {
                  finalFiles.value = [...finalFiles.value, t.output_file];
                  addLog('file', `📄 ${t.output_file} 已生成 → 可下载`, logBox);
                }
              }
            }
          }
          if (evt.done !== undefined) doneCount.value = evt.done;
          if (evt.total !== undefined) totalCount.value = evt.total;
          lastProgressTs.value = Date.now();
          if (Date.now() - lastProgressTs.value > 8000) {
            activeWorkers.value = 0;
          }
          break;

        case 'file_ready':
          if (evt.files) {
            for (const f of evt.files) {
              if (!finalFiles.value.includes(f)) {
                finalFiles.value = [...finalFiles.value, f];
                addLog('file', `📄 ${f} 已生成 → 可下载`, logBox);
              }
              cacheFileInBrowser(taskId, f);
            }
          }
          break;

        case 'error':
          addLog('error', evt.error || evt.content || '', logBox);
          break;

        case 'complete':
          doneCount.value = evt.done || 0;
          totalCount.value = evt.total || 0;
          if (evt.files) finalFiles.value = evt.files;
          if (evt.message) finalMessage.value = evt.message;
          if (evt.duration) duration.value = evt.duration;
          if (evt.tasks) taskItems.value = evt.tasks;
          activeWorkers.value = 0;
          es!.close();
          break;
      }
    };

    es.addEventListener('answer', handler);
    es.addEventListener('tool_call', handler);
    es.addEventListener('progress', handler);
    es.addEventListener('file_ready', handler);
    es.addEventListener('error', handler);
    es.addEventListener('complete', handler);
  }

  function disconnectSSE() {
    if (es) { es.close(); es = null; }
  }

  async function cacheFileInBrowser(taskId: string, filename: string) {
    try {
      const cache = await caches.open('ppt-agent-files');
      const url = `/api/tasks/${taskId}/files/${encodeURIComponent(filename)}`;
      const cached = await cache.match(url);
      if (cached) return;
      const res = await fetch(url);
      if (res.ok) await cache.put(url, res.clone());
    } catch { /* silent */ }
  }

  async function refreshTaskInfo(taskId: string, taskList: Ref<TaskInfo[]>) {
    try {
      const info = await fetchTask(taskId);
      const arr = taskList.value;
      const idx = arr.findIndex(t => t.id === taskId);
      if (idx >= 0) arr[idx] = info;
    } catch { /* ignore */ }
  }

  function resetState(taskInfo: TaskInfo) {
    taskItems.value = [];
    doneCount.value = taskInfo.done_count;
    totalCount.value = taskInfo.total_count;
    logLines.value = [];
    finalFiles.value = taskInfo.files || [];
    finalMessage.value = '';
    duration.value = taskInfo.duration || '';
    activeWorkers.value = 0;
    batches.value = [];
  }

  function selectTask(taskInfo: TaskInfo, taskList: Ref<TaskInfo[]>, logBox?: HTMLElement | null) {
    resetState(taskInfo);
    if (taskInfo.status === 'running') {
      connectSSE(taskInfo.id, logBox);
    } else {
      refreshTaskInfo(taskInfo.id, taskList);
    }
  }

  return {
    taskItems, doneCount, totalCount, logLines, finalFiles,
    finalMessage, duration, activeWorkers, lastProgressTs, batches,
    connectSSE, disconnectSSE, addLog, resetState, selectTask,
    refreshTaskInfo, cacheFileInBrowser,
  };
}
