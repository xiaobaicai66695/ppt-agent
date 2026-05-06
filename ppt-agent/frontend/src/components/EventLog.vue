<script setup lang="ts">
import { ref, watch, nextTick } from 'vue';
import type { LogLine } from '../types';
import { renderMd } from '../composables/useMarkdown';

const props = defineProps<{
  lines: LogLine[];
  maxHeight?: string;
}>();

const logBox = ref<HTMLElement | null>(null);
const autoScroll = ref(true);

function onScroll() {
  const el = logBox.value;
  if (!el) return;
  autoScroll.value = el.scrollHeight - el.scrollTop - el.clientHeight < 30;
}

watch(() => props.lines.length, () => {
  if (autoScroll.value) {
    nextTick(() => {
      if (logBox.value) logBox.value.scrollTop = logBox.value.scrollHeight;
    });
  }
});
</script>

<template>
  <div class="event-log-section">
    <div class="log-header" @click="autoScroll = !autoScroll" style="cursor:pointer">
      <span class="log-title">事件日志</span>
      <span v-if="lines.length" class="log-badge">{{ lines.length }}</span>
      <span class="log-scroll-state">{{ autoScroll ? '(自动滚动)' : '(已暂停)' }}</span>
    </div>
    <div
      ref="logBox"
      class="log-box"
      :style="{ maxHeight: maxHeight || '400px' }"
      @scroll="onScroll"
    >
      <div v-if="lines.length === 0" class="log-empty">
        <span class="empty-dot">●</span>
        <span>等待事件...</span>
      </div>
      <TransitionGroup v-else name="log" tag="div">
        <div v-for="(l, i) in lines" :key="i" class="log-line" :class="l.kind">
          <template v-if="l.kind === 'divider'">
            <span class="log-text">{{ l.text }}</span>
          </template>
          <template v-else>
            <span class="log-ts">{{ new Date(l.ts).toLocaleTimeString('zh-CN') }}</span>
            <span class="log-dot" :class="l.kind"></span>
            <span class="log-text" v-if="l.kind === 'answer'" v-html="renderMd(l.text)"></span>
            <span class="log-text" v-else>{{ l.text }}</span>
          </template>
        </div>
      </TransitionGroup>
    </div>
  </div>
</template>

<style scoped>
.event-log-section {
  margin-bottom: 1rem;
}
.log-header {
  display: flex; align-items: center; gap: 0.4rem;
  margin-bottom: 0.5rem; font-size: 0.8rem; font-weight: 600;
  user-select: none;
}
.log-title { color: var(--c-text); }
.log-badge {
  font-size: 0.65rem; background: var(--c-bg);
  padding: 0.1rem 0.4rem; border-radius: 999px;
}
.log-scroll-state {
  font-size: 0.65rem; color: var(--c-text-muted); font-weight: 400;
  margin-left: auto;
}
.log-box {
  background: var(--c-surface);
  border: 1px solid var(--c-border);
  border-radius: var(--radius);
  padding: 0.7rem 0.9rem;
  overflow-y: auto;
  font-family: 'SF Mono', 'Fira Code', 'Cascadia Code', 'JetBrains Mono', monospace;
  font-size: 0.72rem; line-height: 1.65;
  box-shadow: var(--shadow-sm);
}
.log-box::-webkit-scrollbar { width: 4px; }
.log-box::-webkit-scrollbar-track { background: transparent; }
.log-box::-webkit-scrollbar-thumb { background: #d1d5db; border-radius: 2px; }

.log-empty {
  display: flex; align-items: center; gap: 0.5rem;
  color: var(--c-text-muted); font-style: italic; font-family: inherit;
}
.empty-dot { color: #d1d5db; animation: pulse 2s infinite; }

.log-line {
  display: flex; align-items: baseline; gap: 0.5rem;
  padding: 0.1rem 0;
}
.log-line.answer { align-items: flex-start; }
.log-ts {
  color: #cbd5e1; flex-shrink: 0;
  font-size: 0.62rem; min-width: 4.5em;
}
.log-dot {
  width: 4px; height: 4px; border-radius: 50%;
  background: #d1d5db; flex-shrink: 0; margin-top: 0.4em;
}
.log-dot.worker { background: #f59e0b; animation: pulse 1s infinite; }
.log-dot.file { background: #10b981; }
.log-dot.tool { background: #a78bfa; }
.log-dot.error { background: #fca5a5; }

.log-text {
  color: var(--c-text-2);
  white-space: pre-wrap;
  word-break: break-word;
}
.log-line.tool .log-text { color: #6d28d9; }
.log-line.error .log-text { color: #dc2626; }
.log-line.worker .log-text { color: #b45309; font-weight: 500; }
.log-line.file .log-text { color: #059669; }
.log-line.divider {
  justify-content: center; padding: 0.3rem 0; font-family: inherit;
}
.log-line.divider .log-text {
  color: #cbd5e1; font-size: 0.62rem;
}

/* Markdown styles */
.log-text :deep(.md-h1) {
  display: block; font-size: 1.1em; font-weight: 700;
  color: var(--c-text); margin: 0.4em 0 0.2em;
}
.log-text :deep(.md-h2) {
  display: block; font-size: 1em; font-weight: 700;
  color: var(--c-text); margin: 0.35em 0 0.15em;
}
.log-text :deep(.md-h3) {
  display: block; font-size: 0.9em; font-weight: 600;
  color: var(--c-text-2); margin: 0.3em 0 0.1em;
}
.log-text :deep(.md-hr) {
  display: block; height: 0;
  border-bottom: 1px dashed #e2e8f0; margin: 0.4em 0;
}
.log-text :deep(.md-code) {
  background: #f1f5f9; padding: 0.1em 0.3em;
  border-radius: 3px; font-size: 0.9em;
}
.log-text :deep(.md-path) {
  background: var(--c-primary-light); color: var(--c-primary);
  padding: 0.1em 0.3em; border-radius: 3px;
}
.log-text :deep(strong) { font-weight: 600; color: var(--c-text); }

/* Transitions */
.log-enter-active { transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1); }
.log-leave-active { transition: all 0.15s ease-in; }
.log-enter-from { opacity: 0; transform: translateX(-8px); }
.log-leave-to { opacity: 0; }

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}
</style>
