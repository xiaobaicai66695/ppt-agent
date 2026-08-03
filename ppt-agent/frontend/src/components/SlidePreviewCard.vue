<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { Check, Download, Eye, FileWarning, LoaderCircle, RefreshCcw } from 'lucide-vue-next';
import type { TaskItem } from '../types';

type ThumbnailStatus = 'pending' | 'ready' | 'error';

const props = defineProps<{
  task: TaskItem;
  taskId: string;
  fileReady: boolean;
  selected?: boolean;
  thumbnailStatus?: ThumbnailStatus;
  thumbnailVersion?: number;
  thumbnailError?: string;
}>();

const emit = defineEmits<{
  toggle: [];
  preview: [task: TaskItem, thumbUrl: string];
}>();

const thumbLoaded = ref(false);
const localThumbError = ref(false);
const retryKey = ref(0);
const forceRetry = ref(false);

const hasThumbError = computed(() => !forceRetry.value && (
  localThumbError.value || props.thumbnailStatus === 'error'
));
const shouldLoadThumbnail = computed(() => props.fileReady && (
  props.thumbnailStatus !== 'pending' || forceRetry.value
));
const shortName = computed(() => {
  const name = props.task.output_file.split(/[/\\]/).pop() || props.task.output_file;
  return name.length > 30 ? `${name.slice(0, 27)}...` : name;
});
const stateLabel = computed(() => {
  if (props.fileReady) return hasThumbError.value ? '预览失败' : thumbLoaded.value ? '可预览' : '文件已就绪';
  if (props.task.status === 'generating') return '正在生成';
  if (props.task.status === 'failed') return '生成失败';
  return '等待生成';
});

watch(() => [props.thumbnailVersion, props.thumbnailStatus], () => {
  if (props.thumbnailStatus === 'ready') {
    thumbLoaded.value = false;
    localThumbError.value = false;
    forceRetry.value = false;
  }
});

function getThumbUrl(): string {
  const name = props.task.output_file.split(/[/\\]/).pop() || props.task.output_file;
  return `/api/tasks/${props.taskId}/thumb/${encodeURIComponent(name)}?v=${props.thumbnailVersion || 0}-${retryKey.value}`;
}

function getDownloadUrl(): string {
  const name = props.task.output_file.split(/[/\\]/).pop() || props.task.output_file;
  return `/api/tasks/${props.taskId}/files/${encodeURIComponent(name)}`;
}

function onImgError() {
  thumbLoaded.value = false;
  localThumbError.value = true;
  forceRetry.value = false;
}

function retryThumb() {
  thumbLoaded.value = false;
  localThumbError.value = false;
  forceRetry.value = true;
  retryKey.value++;
}

function handlePreview() {
  if (props.fileReady && thumbLoaded.value && !hasThumbError.value) {
    emit('preview', props.task, getThumbUrl());
  }
}
</script>

<template>
  <article class="slide-card" :class="{ selected, pending: !fileReady }">
    <div class="slide-media">
      <img
        v-if="shouldLoadThumbnail && !hasThumbError"
        :key="`${thumbnailVersion || 0}-${retryKey}`"
        :src="getThumbUrl()"
        :class="{ loaded: thumbLoaded }"
        :alt="`第 ${task.page_index} 页预览：${task.title}`"
        loading="lazy"
        @load="thumbLoaded = true"
        @error="onImgError"
      />

      <div v-if="!fileReady" class="media-state waiting" role="status">
        <span class="page-ghost">{{ task.page_index }}</span>
        <LoaderCircle v-if="task.status === 'generating'" :size="22" class="spin" />
        <span>{{ stateLabel }}</span>
      </div>

      <div v-else-if="hasThumbError" class="media-state failed">
        <FileWarning :size="24" />
        <span>{{ thumbnailError || '缩略图转换失败' }}</span>
        <button type="button" @click.stop="retryThumb">
          <RefreshCcw :size="14" />
          重试预览
        </button>
      </div>

      <div v-else-if="!thumbLoaded" class="media-state loading" role="status">
        <LoaderCircle :size="22" class="spin" />
        <span>{{ thumbnailStatus === 'pending' ? '正在准备预览' : '正在加载预览' }}</span>
      </div>

      <button
        v-if="thumbLoaded && !hasThumbError"
        class="preview-action"
        type="button"
        :aria-label="`放大预览第 ${task.page_index} 页：${task.title}`"
        @click="handlePreview"
      >
        <Eye :size="19" />
        <span>查看</span>
      </button>

      <button
        v-if="fileReady"
        class="select-action"
        type="button"
        :aria-label="selected ? `取消选择第 ${task.page_index} 页` : `选择第 ${task.page_index} 页`"
        :aria-pressed="selected"
        @click.stop="emit('toggle')"
      >
        <Check v-if="selected" :size="15" :stroke-width="2.5" />
      </button>
    </div>

    <footer class="slide-meta">
      <span class="page-index">{{ String(task.page_index).padStart(2, '0') }}</span>
      <span class="slide-copy">
        <strong :title="task.title">{{ task.title }}</strong>
        <small :class="{ error: hasThumbError || task.status === 'failed' }">{{ stateLabel }} · {{ shortName }}</small>
      </span>
      <a
        v-if="fileReady"
        :href="getDownloadUrl()"
        class="download-action"
        :aria-label="`下载第 ${task.page_index} 页：${task.title}`"
        title="下载 PPTX"
        @click.stop
      >
        <Download :size="18" />
      </a>
    </footer>
  </article>
</template>

<style scoped>
.slide-card {
  min-width: 0;
  overflow: hidden;
  border: 1px solid var(--border);
  border-radius: 6px;
  background: var(--surface);
  box-shadow: var(--shadow-xs);
  transition: border-color var(--motion-fast), box-shadow var(--motion-fast), transform var(--motion-fast);
}

.slide-card:hover { border-color: var(--border-strong); box-shadow: var(--shadow-sm); transform: translateY(-1px); }
.slide-card.selected { border-color: var(--action-ink); box-shadow: 0 0 0 2px var(--action-soft); }
.slide-card.pending { border-style: dashed; box-shadow: none; }

.slide-media {
  position: relative;
  aspect-ratio: 16 / 9;
  overflow: hidden;
  background: #e9eded;
  border-bottom: 1px solid var(--divider);
}

.slide-media img { width: 100%; height: 100%; object-fit: cover; opacity: 0; transition: opacity var(--motion-medium); }
.slide-media img.loaded { opacity: 1; }

.media-state {
  position: absolute;
  inset: 0;
  display: grid;
  place-content: center;
  justify-items: center;
  gap: 8px;
  padding: 20px;
  color: var(--text-muted);
  background: var(--surface-muted);
  font-size: 12px;
  text-align: center;
}

.media-state.failed { color: var(--danger); background: var(--danger-soft); }
.media-state button {
  min-height: 36px;
  padding: 0 10px;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  border: 1px solid #e5b4b4;
  border-radius: 5px;
  color: var(--danger);
  background: var(--surface);
  font-weight: 650;
  cursor: pointer;
}

.page-ghost {
  width: 36px;
  height: 36px;
  display: grid;
  place-items: center;
  border: 1px solid var(--border-strong);
  border-radius: 5px;
  color: var(--text-secondary);
  background: var(--surface);
  font-weight: 750;
}

.spin { animation: spin 0.9s linear infinite; }

.preview-action {
  position: absolute;
  inset: 0;
  width: 100%;
  border: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 7px;
  color: #ffffff;
  background: rgba(15, 20, 21, 0.58);
  font-weight: 700;
  opacity: 0;
  cursor: zoom-in;
  transition: opacity var(--motion-fast);
}

.slide-media:hover .preview-action,
.preview-action:focus-visible { opacity: 1; }

.select-action {
  position: absolute;
  top: 8px;
  left: 8px;
  z-index: 2;
  width: 32px;
  height: 32px;
  display: grid;
  place-items: center;
  border: 1px solid rgba(15, 20, 21, 0.26);
  border-radius: 5px;
  color: #ffffff;
  background: rgba(255, 255, 255, 0.9);
  cursor: pointer;
}

.slide-card.selected .select-action { border-color: var(--action-ink); background: var(--action-ink); }

.slide-meta {
  min-height: 58px;
  padding: 9px 10px;
  display: grid;
  grid-template-columns: 28px minmax(0, 1fr) 36px;
  align-items: center;
  gap: 8px;
}

.page-index { color: var(--text-muted); font-size: 11px; font-variant-numeric: tabular-nums; }
.slide-copy { min-width: 0; display: flex; flex-direction: column; }
.slide-copy strong { overflow: hidden; color: var(--text); font-size: 12px; text-overflow: ellipsis; white-space: nowrap; }
.slide-copy small { overflow: hidden; margin-top: 3px; color: var(--text-muted); font-size: 10px; text-overflow: ellipsis; white-space: nowrap; }
.slide-copy small.error { color: var(--danger); }

.download-action {
  width: 36px;
  height: 36px;
  display: grid;
  place-items: center;
  border-radius: 5px;
  color: var(--text-secondary);
  text-decoration: none;
}
.download-action:hover { color: var(--action-ink); background: var(--action-soft); }

@keyframes spin { to { transform: rotate(360deg); } }

@media (hover: none) {
  .preview-action { inset: auto 8px 8px auto; width: 72px; height: 38px; border-radius: 5px; opacity: 1; cursor: pointer; }
}

@media (prefers-reduced-motion: reduce) {
  .spin { animation: none; }
}
</style>
