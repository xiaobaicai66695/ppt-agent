<script setup lang="ts">
import { computed, onUnmounted, ref, watch } from 'vue';
import { Check, Clock3, LoaderCircle } from 'lucide-vue-next';
import type { TaskItem } from '../types';

const props = defineProps<{
  doneCount: number;
  totalCount: number;
  taskItems: TaskItem[];
  isRunning: boolean;
  phase?: string;
  phaseDetail?: string;
  taskId?: string;
  createdAt?: string;
}>();

const completionTimestamps = ref<number[]>([]);
const taskStartTime = ref(Date.now());
const lastDoneCount = ref(0);
const now = ref(Date.now());
let clockTimer: ReturnType<typeof setInterval> | null = null;

function resetTracking() {
  const created = props.createdAt ? Date.parse(props.createdAt) : NaN;
  taskStartTime.value = Number.isFinite(created) ? created : Date.now();
  now.value = Date.now();
  lastDoneCount.value = props.doneCount;
  completionTimestamps.value = [];
}

watch(() => props.taskId, resetTracking, { immediate: true });
watch(() => props.isRunning, (running) => {
  if (clockTimer) clearInterval(clockTimer);
  clockTimer = null;
  now.value = Date.now();
  if (running) clockTimer = setInterval(() => { now.value = Date.now(); }, 1000);
}, { immediate: true });
watch(() => props.doneCount, (newDone) => {
  if (newDone <= lastDoneCount.value) return;
  completionTimestamps.value.push(Date.now());
  if (completionTimestamps.value.length > 20) completionTimestamps.value.shift();
  lastDoneCount.value = newDone;
});
onUnmounted(() => { if (clockTimer) clearInterval(clockTimer); });

const progressPct = computed(() => props.totalCount > 0
  ? Math.min(100, Math.round((props.doneCount / props.totalCount) * 100))
  : 0);
const elapsedMs = computed(() => Math.max(0, now.value - taskStartTime.value));
const remainingCount = computed(() => Math.max(0, props.totalCount - props.doneCount));
const failedCount = computed(() => props.taskItems.filter(item => item.status === 'failed').length);

const etaMs = computed(() => {
  const recent = completionTimestamps.value.slice(-10);
  if (remainingCount.value <= 0 || recent.length < 2) return 0;
  const duration = recent[recent.length - 1] - recent[0];
  if (duration <= 0) return 0;
  const estimate = remainingCount.value * duration / (recent.length - 1);
  return estimate >= 5000 && estimate <= 7_200_000 ? Math.round(estimate) : 0;
});

function formatDuration(ms: number): string {
  const seconds = Math.max(0, Math.floor(ms / 1000));
  if (seconds < 60) return `${seconds}s`;
  const minutes = Math.floor(seconds / 60);
  const rest = seconds % 60;
  if (minutes < 60) return `${minutes}m${rest ? ` ${rest}s` : ''}`;
  return `${Math.floor(minutes / 60)}h ${minutes % 60}m`;
}

const etaLabel = computed(() => etaMs.value ? `预计还需 ${formatDuration(etaMs.value)}` : '正在估算剩余时间');
const phaseLabel = computed(() => {
  if (!props.isRunning && props.totalCount > 0 && props.doneCount >= props.totalCount) return '全部完成';
  const labels: Record<string, string> = {
    preparing: '准备资源',
    planning: '规划内容',
    generating: '生成页面',
    qa: '整理输出',
    fixing: '完善页面',
    complete: '全部完成',
    completed: '全部完成',
    failed: '生成失败',
    cancelled: '已中断',
  };
  return props.phaseDetail || labels[props.phase || ''] || (props.isRunning ? '正在启动' : '等待开始');
});
const steps = computed(() => [
  { key: 'planning', label: '规划', done: props.totalCount > 0 || props.doneCount > 0 },
  { key: 'generating', label: '生成', done: props.totalCount > 0 && props.doneCount >= props.totalCount },
  { key: 'complete', label: '交付', done: !props.isRunning && props.totalCount > 0 && props.doneCount >= props.totalCount },
]);
</script>

<template>
  <section v-if="isRunning || totalCount > 0" class="progress-panel" aria-live="polite">
    <header class="progress-head">
      <div class="progress-heading">
        <span class="phase-icon" :class="{ running: isRunning }" aria-hidden="true">
          <LoaderCircle v-if="isRunning" :size="18" />
          <Check v-else :size="18" />
        </span>
        <span>
          <strong>{{ phaseLabel }}</strong>
          <small v-if="isRunning">生成完成的页面会立即出现在下方</small>
          <small v-else>{{ failedCount ? `${failedCount} 页需要关注` : '页面已经可以预览和下载' }}</small>
        </span>
      </div>
      <div class="progress-count">
        <strong>{{ progressPct }}%</strong>
        <span>{{ doneCount }} / {{ totalCount || '—' }} 页</span>
      </div>
    </header>

    <div
      class="progress-track"
      role="progressbar"
      :aria-label="totalCount > 0 ? 'PPT 生成进度' : '正在规划 PPT'"
      :aria-valuenow="totalCount > 0 ? progressPct : undefined"
      aria-valuemin="0"
      aria-valuemax="100"
    >
      <span
        class="progress-fill"
        :class="{ indeterminate: isRunning && totalCount === 0 }"
        :style="totalCount > 0 ? { width: `${progressPct}%` } : undefined"
      />
    </div>

    <footer class="progress-foot">
      <ol class="progress-steps" aria-label="生成阶段">
        <li v-for="step in steps" :key="step.key" :class="{ done: step.done, active: phase === step.key || (step.key === 'generating' && isRunning && totalCount > 0) }">
          <span aria-hidden="true"><Check v-if="step.done" :size="11" /></span>
          {{ step.label }}
        </li>
      </ol>
      <div v-if="isRunning" class="time-copy">
        <Clock3 :size="15" />
        <span>已运行 {{ formatDuration(elapsedMs) }}</span>
        <span class="divider" aria-hidden="true"></span>
        <span>{{ etaLabel }}</span>
      </div>
    </footer>
  </section>
</template>

<style scoped>
.progress-panel {
  padding: 18px 20px;
  border: 1px solid var(--border);
  border-radius: 6px;
  background: var(--surface);
  box-shadow: var(--shadow-xs);
}

.progress-head { display: flex; align-items: center; justify-content: space-between; gap: 20px; }
.progress-heading { min-width: 0; display: flex; align-items: center; gap: 11px; }
.progress-heading > span:last-child { min-width: 0; display: flex; flex-direction: column; }
.progress-heading strong { overflow: hidden; color: var(--text); font-size: 14px; text-overflow: ellipsis; white-space: nowrap; }
.progress-heading small { margin-top: 2px; color: var(--text-muted); font-size: 11px; }

.phase-icon {
  width: 36px;
  height: 36px;
  display: grid;
  place-items: center;
  flex: 0 0 auto;
  border-radius: 6px;
  color: var(--success);
  background: var(--success-soft);
}
.phase-icon.running { color: var(--info); background: var(--info-soft); }
.phase-icon.running svg { animation: spin 1s linear infinite; }

.progress-count { flex: 0 0 auto; display: flex; align-items: baseline; gap: 8px; font-variant-numeric: tabular-nums; }
.progress-count strong { color: var(--text); font-size: 20px; line-height: 1; }
.progress-count span { color: var(--text-muted); font-size: 11px; }

.progress-track { height: 6px; margin: 17px 0 14px; overflow: hidden; border-radius: 3px; background: var(--surface-pressed); }
.progress-fill { display: block; height: 100%; border-radius: inherit; background: var(--action-ink); transition: width var(--motion-medium); }
.progress-fill.indeterminate { width: 34%; animation: indeterminate 1.35s ease-in-out infinite; }

.progress-foot { display: flex; align-items: center; justify-content: space-between; gap: 16px; }
.progress-steps { margin: 0; padding: 0; display: flex; align-items: center; gap: 18px; list-style: none; }
.progress-steps li { display: flex; align-items: center; gap: 6px; color: var(--text-muted); font-size: 11px; }
.progress-steps li > span { width: 16px; height: 16px; display: grid; place-items: center; border: 1px solid var(--border-strong); border-radius: 50%; }
.progress-steps li.done { color: var(--success); }
.progress-steps li.done > span { border-color: var(--success); background: var(--success-soft); }
.progress-steps li.active { color: var(--info); font-weight: 700; }

.time-copy { display: flex; align-items: center; justify-content: flex-end; gap: 7px; color: var(--text-muted); font-size: 11px; }
.divider { width: 1px; height: 12px; margin: 0 2px; background: var(--border); }

@keyframes spin { to { transform: rotate(360deg); } }
@keyframes indeterminate { 0% { transform: translateX(-115%); } 100% { transform: translateX(310%); } }

@media (max-width: 720px) {
  .progress-panel { padding: 15px; }
  .progress-heading small { white-space: normal; }
  .progress-count { flex-direction: column; align-items: flex-end; gap: 2px; }
  .progress-foot { align-items: flex-start; flex-direction: column; }
  .time-copy { width: 100%; justify-content: flex-start; flex-wrap: wrap; }
}

@media (max-width: 420px) {
  .progress-steps { width: 100%; justify-content: space-between; gap: 8px; }
  .progress-heading small { display: none; }
}

@media (prefers-reduced-motion: reduce) {
  .phase-icon.running svg,
  .progress-fill.indeterminate { animation: none; }
}
</style>
