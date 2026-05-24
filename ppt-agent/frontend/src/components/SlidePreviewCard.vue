<script setup lang="ts">
import { ref, computed } from 'vue';
import type { TaskItem } from '../types';

const props = defineProps<{
  task: TaskItem;
  taskId: string;
  fileReady: boolean;
  selected?: boolean;
}>();

const emit = defineEmits<{
  toggle: [];
  preview: [task: TaskItem, thumbUrl: string];
}>();

const thumbLoaded = ref(false);
const thumbError = ref(false);
const retryKey = ref(0);

function getThumbUrl(): string {
  const name = props.task.output_file.split(/[/\\]/).pop() || props.task.output_file;
  return `/api/tasks/${props.taskId}/thumb/${encodeURIComponent(name)}?t=${retryKey.value}`;
}

function getDownloadUrl(): string {
  const name = props.task.output_file.split(/[/\\]/).pop() || props.task.output_file;
  return `/api/tasks/${props.taskId}/files/${encodeURIComponent(name)}`;
}

function onImgError() {
  thumbLoaded.value = false;
  if (retryKey.value < 3) {
    setTimeout(() => { retryKey.value++; }, 1500);
  } else {
    thumbError.value = true;
  }
}

function handlePreview() {
  if (props.fileReady && !thumbError.value) {
    emit('preview', props.task, getThumbUrl());
  }
}

const shortName = computed(() => {
  const name = props.task.output_file.split(/[/\\]/).pop() || props.task.output_file;
  return name.length > 28 ? name.slice(0, 25) + '...' : name;
});
</script>

<template>
  <div class="slide-preview" :class="{ placeholder: !fileReady, selected }">
    <template v-if="fileReady">
      <div class="select-overlay" @click.stop="emit('toggle')">
        <span class="check-mark" :class="{ on: selected }">
          <svg v-if="selected" viewBox="0 0 16 16" fill="currentColor"><path d="M6 10.8L3.2 8l-.9.9L6 12.6l8-8-.9-.9z"/></svg>
        </span>
      </div>
      <div class="preview-inner">
        <div class="preview-thumb" @click.stop="handlePreview">
          <img
            v-if="!thumbError"
            :key="retryKey"
            :src="getThumbUrl()"
            :class="{ loaded: thumbLoaded }"
            @load="thumbLoaded = true"
            @error="onImgError"
            loading="lazy"
            alt=""
          />
          <span v-if="!thumbLoaded && !thumbError" class="thumb-spinner"></span>
          <span v-if="thumbError" class="thumb-fallback">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/>
              <polyline points="14 2 14 8 20 8"/>
              <line x1="12" y1="18" x2="12" y2="12"/>
              <polyline points="9 15 12 18 15 15"/>
            </svg>
            <span class="thumb-note">生成完成</span>
          </span>
          <!-- Preview button overlay -->
          <button v-if="!thumbError" class="preview-btn" title="在线预览" @click.stop="handlePreview">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/>
              <circle cx="12" cy="12" r="3"/>
            </svg>
          </button>
        </div>
        <div class="preview-info">
          <span class="preview-idx">#{{ task.page_index }}</span>
          <span class="preview-title">{{ task.title }}</span>
          <span class="preview-name">{{ shortName }}</span>
          <a
            :href="getDownloadUrl()"
            class="download-btn"
            title="下载"
            @click.stop
          >
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/>
              <polyline points="7 10 12 15 17 10"/>
              <line x1="12" y1="15" x2="12" y2="3"/>
            </svg>
          </a>
        </div>
      </div>
    </template>
    <template v-else>
      <div class="placeholder-body">
        <div class="ph-idx">{{ task.page_index }}</div>
        <div class="ph-info">
          <div class="ph-title">{{ task.title }}</div>
          <div class="ph-status" :class="task.status">
            <span class="ph-dot"></span>
            {{ task.status === 'generating' ? '生成中...' : task.status === 'pending' ? '等待中' : task.status }}
          </div>
          <!-- Show QA report details when the slide has issues -->
          <div v-if="task.qa_report" class="qa-report-summary" :class="{ failed: task.status === 'failed' }">
            <span class="qa-badge">QA 报告</span>
            <span class="qa-preview">{{ task.qa_report.slice(0, 150) }}{{ task.qa_report.length > 150 ? '...' : '' }}</span>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.slide-preview {
  background: var(--bg-base);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  overflow: hidden;
  transition: transform var(--transition), box-shadow var(--transition), border-color var(--transition);
  box-shadow: var(--shadow-sm);
  position: relative;
}
.slide-preview:not(.placeholder):hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
  border-color: var(--accent);
}
.slide-preview.placeholder {
  border-style: dashed; background: var(--bg-muted); opacity: 0.85;
}
.slide-preview.selected {
  border-color: var(--accent);
  box-shadow: 0 0 0 2px rgba(59,130,246,0.25);
}

.preview-inner { display: flex; flex-direction: column; width: 100%; }

.select-overlay {
  position: absolute; top: 0.4rem; right: 0.4rem; z-index: 3; cursor: pointer;
}
.check-mark {
  width: 22px; height: 22px; border-radius: 4px;
  border: 2px solid #cbd5e1; background: rgba(255,255,255,0.9);
  display: flex; align-items: center; justify-content: center;
  transition: all var(--transition);
}
.check-mark svg { width: 14px; height: 14px; color: #fff; }
.check-mark.on { border-color: var(--accent); background: var(--accent); }

.preview-thumb {
  width: 100%; aspect-ratio: 16 / 9;
  background: var(--bg-muted);
  display: flex; align-items: center; justify-content: center;
  overflow: hidden; position: relative;
  cursor: zoom-in;
}
.preview-thumb img {
  width: 100%; height: 100%; object-fit: cover;
  opacity: 0; transition: opacity 0.4s;
}
.preview-thumb img.loaded { opacity: 1; }

.thumb-spinner {
  position: absolute; width: 28px; height: 28px;
  border: 2px solid var(--border); border-top-color: var(--accent);
  border-radius: 50%; animation: spin 0.7s linear infinite;
}

.thumb-fallback {
  display: flex; flex-direction: column; align-items: center; gap: 0.4rem;
  color: var(--text-muted);
}
.thumb-fallback svg { width: 40px; height: 40px; }
.thumb-note { font-size: 0.7rem; }

.preview-btn {
  position: absolute; inset: 0;
  display: flex; align-items: center; justify-content: center;
  background: rgba(0,0,0,0.4); opacity: 0; transition: opacity 0.2s;
  border: none; cursor: zoom-in;
}
.preview-btn svg { width: 36px; height: 36px; color: #fff; }
.preview-thumb:hover .preview-btn { opacity: 1; }

.preview-info {
  display: flex; align-items: center; gap: 0.5rem; padding: 0.65rem 0.75rem;
}
.preview-idx {
  font-size: 0.72rem; font-weight: 700;
  color: var(--accent); background: var(--accent-light);
  padding: 0.15rem 0.5rem; border-radius: 4px; flex-shrink: 0;
}
.preview-title {
  font-size: 0.85rem; font-weight: 600;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
  line-height: 1.3; flex: 1;
}
.preview-name {
  font-size: 0.62rem; color: var(--text-muted);
  font-family: 'SF Mono', monospace;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
  max-width: 30%; flex-shrink: 0; margin-left: auto;
}

.download-btn {
  display: flex; align-items: center; justify-content: center;
  width: 24px; height: 24px; border-radius: 4px;
  color: var(--text-muted); text-decoration: none;
  transition: color var(--transition), background var(--transition);
  flex-shrink: 0;
}
.download-btn svg { width: 14px; height: 14px; }
.download-btn:hover { color: var(--accent); background: var(--accent-light); }

.placeholder-body {
  display: flex; align-items: center; gap: 0.75rem; padding: 1rem 0.75rem; min-height: 72px;
}
.ph-idx {
  width: 36px; height: 36px; border-radius: 50%;
  background: var(--border-light); color: var(--text-muted);
  display: flex; align-items: center; justify-content: center;
  font-size: 0.85rem; font-weight: 700; flex-shrink: 0;
}
.ph-info { flex: 1; min-width: 0; }
.ph-title { font-size: 0.82rem; font-weight: 500; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; margin-bottom: 0.2rem; }
.ph-status { font-size: 0.65rem; font-weight: 600; display: flex; align-items: center; gap: 0.25rem; color: var(--text-muted); }
.ph-dot { width: 5px; height: 5px; border-radius: 50%; background: currentColor; }
.ph-status.generating { color: #d97706; }
.ph-status.generating .ph-dot { animation: pulse 0.8s infinite; }
.ph-status.failed { color: var(--error); }
.ph-status.failed .ph-dot { background: var(--error); }
.qa-report-summary {
  margin-top: 0.25rem;
  padding: 0.2rem 0.35rem;
  background: color-mix(in srgb, var(--error) 8%, transparent);
  border-radius: 3px;
  font-size: 0.6rem;
  color: var(--text-muted);
  line-height: 1.3;
  display: flex;
  flex-direction: column;
  gap: 0.1rem;
}
.qa-badge {
  font-size: 0.55rem;
  font-weight: 600;
  color: var(--error);
  text-transform: uppercase;
  letter-spacing: 0.03em;
}
.qa-preview {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--text-muted);
}

@keyframes spin { to { transform: rotate(360deg); } }
@keyframes pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.5; } }
</style>
