<script setup lang="ts">
import { computed } from 'vue';
import type { TaskItem } from '../types';

const props = defineProps<{
  doneCount: number;
  totalCount: number;
  taskItems: TaskItem[];
  isRunning: boolean;
}>();

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
</script>

<template>
  <div v-if="totalCount > 0" class="progress-section">
    <div class="progress-header">
      <span>生成进度</span>
      <span class="progress-num">
        <strong>{{ doneCount }}</strong> / {{ totalCount }} 页
        <span class="progress-pct">({{ progressPct }}%)</span>
      </span>
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
  background: var(--c-surface);
  border: 1px solid var(--c-border);
  border-radius: var(--radius);
  padding: 0.85rem 1.1rem;
  margin-bottom: 1rem;
  box-shadow: var(--shadow-sm);
}
.progress-header {
  display: flex; justify-content: space-between;
  margin-bottom: 0.5rem; font-size: 0.8rem;
}
.progress-num { color: var(--c-text-muted); font-size: 0.78rem; }
.progress-num strong { color: var(--c-primary); font-size: 1rem; }
.progress-pct { font-size: 0.72rem; color: var(--c-text-muted); }
.progress-track {
  height: 8px; background: var(--c-border-light);
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
  font-size: 0.68rem; color: var(--c-text-muted);
  display: flex; align-items: center; gap: 0.25rem;
}
.legend-item.active { color: #b45309; font-weight: 500; }
.legend-dot { width: 7px; height: 7px; border-radius: 50%; }
.legend-dot.done { background: #10b981; }
.legend-dot.pending { background: #3b82f6; animation: pulse 1.5s infinite; }
.legend-dot.generating { background: #f59e0b; animation: pulse 0.8s infinite; }
.legend-dot.failed { background: #ef4444; }

@keyframes shimmer {
  0% { background-position: -200% 0; }
  100% { background-position: 200% 0; }
}
@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}
</style>
