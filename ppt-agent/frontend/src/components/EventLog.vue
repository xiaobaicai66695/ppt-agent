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

// Watch both length (new lines) and last line text (streaming append)
watch(
  () => props.lines.length + (props.lines[props.lines.length - 1]?.text?.length || 0),
  () => {
    if (autoScroll.value) {
      nextTick(() => {
        if (logBox.value) logBox.value.scrollTop = logBox.value.scrollHeight;
      });
    }
  },
);
</script>

<template>
  <div class="event-log-section">
    <div class="log-header" @click="autoScroll = !autoScroll" :title="autoScroll ? '点击暂停滚动' : '点击恢复自动滚动'">
      <span class="log-title">事件日志</span>
      <span v-if="lines.length" class="log-badge">{{ lines.length }}</span>
      <span class="log-scroll-state">{{ autoScroll ? '(自动滚动)' : '(已暂停)' }}</span>
    </div>
    <div
      ref="logBox"
      class="log-box"
      :style="maxHeight ? { maxHeight } : { flex: 1 }"
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
  display: flex; flex-direction: column; height: 100%;
}
.log-header {
  display: flex; align-items: center; gap: 0.4rem;
  margin-bottom: 0.5rem; font-size: 0.8rem; font-weight: 600;
  user-select: none; cursor: pointer;
}
.log-title { color: var(--text); }
.log-badge {
  font-size: 0.65rem; background: var(--surface-muted);
  padding: 0.1rem 0.4rem; border-radius: 999px;
}
.log-scroll-state {
  font-size: 0.65rem; color: var(--text-muted); font-weight: 400;
  margin-left: auto;
}
.log-box {
  flex: 1; min-height: 0;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 0.7rem 0.9rem;
  overflow-y: auto;
  font-family: 'SF Mono', 'Fira Code', 'Cascadia Code', 'JetBrains Mono', monospace;
  font-size: 0.72rem; line-height: 1.65;
  box-shadow: var(--shadow-sm);
}
.log-box::-webkit-scrollbar { width: 4px; }
.log-box::-webkit-scrollbar-track { background: transparent; }
.log-box::-webkit-scrollbar-thumb { background: var(--text-disabled); border-radius: 2px; }

.log-empty {
  display: flex; align-items: center; gap: 0.5rem;
  color: var(--text-muted); font-style: italic; font-family: inherit;
}
.empty-dot { color: var(--text-disabled); animation: pulse 2s infinite; }

.log-line {
  display: flex; align-items: baseline; gap: 0.5rem;
  padding: 0.1rem 0;
}
.log-line.answer { align-items: flex-start; }
.log-ts {
  color: var(--text-disabled); flex-shrink: 0;
  font-size: 0.62rem; min-width: 4.5em;
}
.log-dot {
  width: 4px; height: 4px; border-radius: 50%;
  background: var(--text-disabled); flex-shrink: 0; margin-top: 0.4em;
}
.log-dot.worker { background: var(--warning); animation: pulse 1s infinite; }
.log-dot.file { background: var(--success); }
.log-dot.tool { background: var(--info); }
.log-dot.error { background: var(--danger); }

.log-text {
  color: var(--text-secondary);
  white-space: pre-wrap;
  word-break: break-word;
}
.log-line.tool .log-text { color: var(--info); }
.log-line.error .log-text { color: var(--danger); }
.log-line.worker .log-text { color: var(--warning); font-weight: 500; }
.log-line.file .log-text { color: var(--success); }
.log-line.divider {
  justify-content: center; padding: 0.3rem 0; font-family: inherit;
}
.log-line.divider .log-text {
  color: var(--text-disabled); font-size: 0.62rem;
}

/* Markdown styles */
.log-text :deep(.md-h1) {
  display: block; font-size: 1.1em; font-weight: 700;
  color: var(--text); margin: 0.4em 0 0.2em;
}
.log-text :deep(.md-h2) {
  display: block; font-size: 1em; font-weight: 700;
  color: var(--text); margin: 0.35em 0 0.15em;
}
.log-text :deep(.md-h3) {
  display: block; font-size: 0.9em; font-weight: 600;
  color: var(--text-secondary); margin: 0.3em 0 0.1em;
}
.log-text :deep(.md-hr) {
  display: block; height: 0;
  border-bottom: 1px dashed var(--border); margin: 0.4em 0;
}
.log-text :deep(.md-code) {
  background: var(--surface-muted); padding: 0.1em 0.3em;
  border-radius: 3px; font-size: 0.9em;
}
.log-text :deep(.md-path) {
  background: var(--action-soft); color: var(--action-ink);
  padding: 0.1em 0.3em; border-radius: 3px;
}
.log-text :deep(strong) { font-weight: 600; color: var(--text); }

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
