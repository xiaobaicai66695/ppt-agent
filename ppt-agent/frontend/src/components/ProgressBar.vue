<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import type { TaskItem } from '../types';

const props = defineProps<{
  doneCount: number;
  totalCount: number;
  taskItems: TaskItem[];
  isRunning: boolean;
  phase?: string;
  phaseDetail?: string;
}>();

// ── ETA Estimation ─────────────────────────────────────────────────────
// Track timestamps of each done slide for rate estimation
const completionTimestamps = ref<number[]>([]);
const taskStartTime = ref<number>(Date.now());
const lastDoneCount = ref(0);

// Reset when total count changes (new task)
watch(() => props.totalCount, (newTotal) => {
  if (newTotal > 0) {
    taskStartTime.value = Date.now();
    lastDoneCount.value = 0;
    completionTimestamps.value = [];
  }
});

// Track when doneCount increases
watch(() => props.doneCount, (newDone) => {
  if (newDone > lastDoneCount.value) {
    const now = Date.now();
    completionTimestamps.value.push(now);
    // Keep only last 20 timestamps for rolling average
    if (completionTimestamps.value.length > 20) {
      completionTimestamps.value.shift();
    }
    lastDoneCount.value = newDone;
  }
});

const elapsedMs = computed(() => {
  if (!props.isRunning && props.doneCount === 0) return 0;
  return Date.now() - taskStartTime.value;
});

const etaMs = computed(() => {
  const pending = props.totalCount - props.doneCount;
  if (pending <= 0 || props.totalCount <= 0) return 0;

  const timestamps = completionTimestamps.value;
  if (timestamps.length < 2) return 0;

  // Use the last N intervals for rate calculation
  const windowSize = Math.min(timestamps.length, 10);
  const recent = timestamps.slice(-windowSize);
  const firstTs = recent[0];
  const lastTs = recent[recent.length - 1];
  const duration = lastTs - firstTs;

  if (duration <= 0 || recent.length < 2) return 0;

  const slidesPerMs = (recent.length - 1) / duration;
  const remainingMs = pending / slidesPerMs;

  // Cap ETA at reasonable bounds (max 2 hours, min 5 seconds)
  if (remainingMs > 2 * 60 * 60 * 1000) return 0;
  if (remainingMs < 5000) return 0;

  return Math.round(remainingMs);
});

const etaLabel = computed(() => {
  const ms = etaMs.value;
  if (ms === 0) return '';
  const secs = Math.ceil(ms / 1000);
  if (secs < 60) return `约 ${secs}s`;
  const mins = Math.floor(secs / 60);
  const remSecs = secs % 60;
  if (mins < 60) {
    return remSecs > 0 ? `约 ${mins}m ${remSecs}s` : `约 ${mins}m`;
  }
  const hours = Math.floor(mins / 60);
  const remMins = mins % 60;
  return `约 ${hours}h ${remMins}m`;
});

const elapsedLabel = computed(() => {
  const ms = elapsedMs.value;
  if (ms === 0) return '';
  const secs = Math.floor(ms / 1000);
  if (secs < 60) return `${secs}s`;
  const mins = Math.floor(secs / 60);
  const remSecs = secs % 60;
  if (mins < 60) {
    return remSecs > 0 ? `${mins}m ${remSecs}s` : `${mins}m`;
  }
  const hours = Math.floor(mins / 60);
  const remMins = mins % 60;
  return `${hours}h ${remMins}m`;
});

// ── Progress Stats ────────────────────────────────────────────────────
const progressPct = computed(() =>
  props.totalCount > 0 ? Math.round((props.doneCount / props.totalCount) * 100) : 0
);

const doneItems = computed(() =>
  props.taskItems.filter(t => t.status === 'done' || t.status === 'qa_done' || t.status === 'fixed')
);
const pendingItems = computed(() =>
  props.taskItems.filter(t => t.status === 'pending' || t.status === 'generating')
);
const generatingItems = computed(() =>
  props.taskItems.filter(t => t.status === 'generating')
);
const failedItems = computed(() =>
  props.taskItems.filter(t => t.status === 'failed')
);

// Completion rate: slides per minute
const slidesPerMinute = computed(() => {
  const timestamps = completionTimestamps.value;
  if (timestamps.length < 2) return null;
  const durationMs = timestamps[timestamps.length - 1] - timestamps[0];
  if (durationMs <= 0) return null;
  const rate = ((timestamps.length - 1) / durationMs) * 60 * 1000;
  return Math.round(rate * 10) / 10;
});

const phaseLabel = (phase: string) => {
  const labels: Record<string, string> = {
    preparing: '读取模板',
    planning: '创建任务',
    generating: '生成幻灯片',
    qa: '质量审查',
    fixing: '优化修复',
    complete: '完成',
  };
  return labels[phase] || phase;
};
</script>

<template>
  <div v-if="totalCount > 0" class="progress-section">
    <div class="progress-header">
      <div class="progress-title-row">
        <span>生成进度</span>
        <span v-if="phase && phase !== 'complete'" class="phase-badge" :class="`phase-${phase}`">
          {{ phaseDetail || phaseLabel(phase) }}
        </span>
      </div>
      <div class="progress-meta">
        <span class="progress-num">
          <strong>{{ doneCount }}</strong> / {{ totalCount }} 页
          <span class="progress-pct">({{ progressPct }}%)</span>
        </span>
        <span v-if="elapsedMs > 0 && isRunning" class="elapsed-time">
          <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5">
            <circle cx="7" cy="7" r="6"/><path d="M7 3.5v3.5l2.5 1.5"/>
          </svg>
          {{ elapsedLabel }}
        </span>
        <span v-if="etaLabel" class="eta-time">
          <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5">
            <circle cx="7" cy="7" r="6"/><path d="M7 3.5v3.5l2.5 1.5"/>
          </svg>
          {{ etaLabel }}
        </span>
        <span v-if="slidesPerMinute !== null && isRunning" class="rate-badge">
          {{ slidesPerMinute }} 页/分钟
        </span>
      </div>
    </div>
    <div class="progress-track">
      <div
        class="progress-fill"
        :class="{ shimmer: isRunning }"
        :style="{ width: progressPct + '%' }"
      />
    </div>
    <div class="progress-legend">
      <span class="legend-item">
        <span class="legend-dot done"></span>
        已完成 {{ doneItems.length }}
      </span>
      <span class="legend-item">
        <span class="legend-dot pending"></span>
        处理中 {{ pendingItems.length }}
      </span>
      <span v-if="generatingItems.length" class="legend-item active">
        <span class="legend-dot generating"></span>
        并行生成 {{ generatingItems.length }}
      </span>
      <span v-if="failedItems.length" class="legend-item">
        <span class="legend-dot failed"></span>
        失败 {{ failedItems.length }}
      </span>
    </div>
  </div>
</template>

<style scoped>
.progress-section {
  background: var(--bg-base);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 0.85rem 1.1rem;
  margin-bottom: 1rem;
  box-shadow: var(--shadow-sm);
}
.progress-header {
  display: flex; justify-content: space-between;
  margin-bottom: 0.5rem; font-size: 0.8rem;
  align-items: flex-start;
  gap: 0.5rem;
  flex-wrap: wrap;
}
.progress-title-row {
  display: flex; align-items: center; gap: 0.5rem;
}
.phase-badge {
  font-size: 0.65rem;
  padding: 0.15rem 0.5rem;
  border-radius: 999px;
  font-weight: 600;
  letter-spacing: 0.02em;
  animation: fadeIn 0.3s ease;
}
.phase-preparing { background: #fef3c7; color: #92400e; }
.phase-planning  { background: #dbeafe; color: #1e40af; }
.phase-generating { background: #ede9fe; color: #5b21b6; }
.phase-qa        { background: #d1fae5; color: #065f46; }
.phase-fixing    { background: #fee2e2; color: #991b1b; }
.phase-complete  { background: #d1fae5; color: #065f46; }

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(-2px); }
  to   { opacity: 1; transform: translateY(0); }
}
.progress-meta {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
}
.progress-num { color: var(--text-muted); font-size: 0.78rem; }
.progress-num strong { color: var(--accent); font-size: 1rem; }
.progress-pct { font-size: 0.72rem; color: var(--text-muted); }

.elapsed-time, .eta-time {
  font-size: 0.7rem;
  color: var(--text-muted);
  display: flex;
  align-items: center;
  gap: 0.2rem;
}
.eta-time { color: #7c3aed; font-weight: 500; }
.elapsed-time svg, .eta-time svg { width: 12px; height: 12px; }

.rate-badge {
  font-size: 0.65rem;
  color: var(--text-muted);
  background: var(--bg-muted);
  border: 1px solid var(--border);
  border-radius: 999px;
  padding: 0.1rem 0.4rem;
}

.progress-track {
  height: 8px; background: var(--border-light);
  border-radius: 4px; overflow: hidden;
}
.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, #3b82f6, #818cf8, #6366f1);
  background-size: 200% 100%;
  border-radius: 4px;
  transition: width 0.6s cubic-bezier(0.4, 0, 0.2, 1);
}
.progress-fill.shimmer {
  animation: shimmer 2s linear infinite;
}
.progress-legend {
  display: flex; gap: 1rem; margin-top: 0.5rem; flex-wrap: wrap;
}
.legend-item {
  font-size: 0.68rem; color: var(--text-muted);
  display: flex; align-items: center; gap: 0.25rem;
}
.legend-item.active { color: #b45309; font-weight: 500; }
.legend-dot { width: 7px; height: 7px; border-radius: 50%; }
.legend-dot.done { background: var(--success); }
.legend-dot.pending { background: var(--accent); animation: pulse 1.5s infinite; }
.legend-dot.generating { background: var(--warning); animation: pulse 0.8s infinite; }
.legend-dot.failed { background: var(--danger); }

@keyframes shimmer {
  0% { background-position: -200% 0; }
  100% { background-position: 200% 0; }
}
@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}
</style>
