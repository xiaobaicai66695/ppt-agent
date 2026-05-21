<script setup lang="ts">
import { STATUS_LABELS } from '../types';

defineProps<{
  pageIndex: number;
  title: string;
  status: string;
}>();
</script>

<template>
  <div class="slide-placeholder" :class="status">
    <div class="ph-idx">{{ pageIndex }}</div>
    <div class="ph-body">
      <div class="ph-title">{{ title }}</div>
      <div class="ph-status-label">
        <span class="ph-dot"></span>
        {{ STATUS_LABELS[status] || status }}
      </div>
    </div>
  </div>
</template>

<style scoped>
.slide-placeholder {
  display: flex; align-items: center; gap: 0.75rem;
  padding: 1rem;
  background: var(--bg-base);
  border: 1px dashed var(--border);
  border-radius: var(--radius);
  transition: all var(--transition);
}
.slide-placeholder.generating {
  border-color: #f59e0b;
  border-style: solid;
  background: #fffbeb;
  animation: borderGlow 2s infinite;
}
.ph-idx {
  width: 36px; height: 36px; border-radius: 50%;
  background: var(--border-light);
  color: var(--text-muted);
  display: flex; align-items: center; justify-content: center;
  font-size: 0.85rem; font-weight: 700; flex-shrink: 0;
}
.slide-placeholder.generating .ph-idx {
  background: #f59e0b;
  color: #fff;
}
.ph-body { flex: 1; min-width: 0; }
.ph-title {
  font-size: 0.85rem; font-weight: 500;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
  margin-bottom: 0.25rem;
}
.ph-status-label {
  font-size: 0.7rem; font-weight: 600;
  display: flex; align-items: center; gap: 0.3rem;
  color: var(--text-muted);
}
.ph-dot { width: 5px; height: 5px; border-radius: 50%; background: currentColor; }
.ph-status-label .ph-dot { animation: pulse 0.8s infinite; }

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}
@keyframes borderGlow {
  0%, 100% { border-color: #f59e0b; box-shadow: 0 0 0 1px rgba(245,158,11,0.1); }
  50% { border-color: #fbbf24; box-shadow: 0 0 0 3px rgba(245,158,11,0.2); }
}
</style>
