<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { Download, ImageOff, LoaderCircle, Maximize2, Presentation } from 'lucide-vue-next'
import AppModal from './AppModal.vue'
import { taskDownloadUrl, taskThumbUrl } from '../api'
import type { TaskInfo } from '../types'

const props = withDefaults(defineProps<{
  task: TaskInfo
  revision?: number
  layout?: 'inline' | 'side-rail'
}>(), { layout: 'inline' })
type PreviewState = 'retrying' | 'failed'

const previewStates = ref<Record<string, PreviewState>>({})
const previewAttempts = ref<Record<string, number>>({})
const previewVersions = ref<Record<string, number>>({})
const retryTimers = new Map<string, number>()
const previewFile = ref('')
const presentationFiles = computed(() => (props.task.files || []).filter(file => file.toLowerCase().endsWith('.pptx')))
const previewTitle = computed(() => previewFile.value ? `${previewFile.value.replace(/\.pptx$/i, '')} · 页面预览` : '页面预览')
const previewState = (file: string) => previewStates.value[file]
const thumbnailURL = (file: string) => `${taskThumbUrl(props.task.id, file)}?preview=${props.revision || 0}-${previewVersions.value[file] || 0}`

function clearRetryTimer(file: string) {
  const timer = retryTimers.get(file)
  if (timer !== undefined) window.clearTimeout(timer)
  retryTimers.delete(file)
}
function resetPreview(file: string) {
  clearRetryTimer(file)
  const { [file]: _state, ...states } = previewStates.value
  const { [file]: _attempt, ...attempts } = previewAttempts.value
  previewStates.value = states
  previewAttempts.value = attempts
  previewVersions.value = { ...previewVersions.value, [file]: (previewVersions.value[file] || 0) + 1 }
}
function markPreviewLoaded(file: string) {
  if (!previewState(file) && !previewAttempts.value[file]) return
  resetPreview(file)
}
function markPreviewError(file: string) {
  const attempts = (previewAttempts.value[file] || 0) + 1
  previewAttempts.value = { ...previewAttempts.value, [file]: attempts }
  if (previewFile.value === file) closePreview()
  if (attempts > 2) {
    previewStates.value = { ...previewStates.value, [file]: 'failed' }
    return
  }
  previewStates.value = { ...previewStates.value, [file]: 'retrying' }
  clearRetryTimer(file)
  retryTimers.set(file, window.setTimeout(() => resetPreview(file), attempts * 2500))
}
function openPreview(file: string) { previewFile.value = file }
function closePreview() { previewFile.value = '' }
function resetAllPreviews() {
  for (const file of retryTimers.keys()) clearRetryTimer(file)
  previewStates.value = {}
  previewAttempts.value = {}
  previewVersions.value = {}
  closePreview()
}
watch(() => [props.task.id, props.revision], resetAllPreviews)
onBeforeUnmount(resetAllPreviews)
</script>

<template>
  <section v-if="presentationFiles.length" class="delivery" :class="{ 'delivery--side-rail': layout === 'side-rail' }" aria-label="已交付演示预览">
    <header class="delivery-head"><div><span>交付预览</span><p>缩略图来自已生成的 PPT 页面</p></div><Presentation :size="19" aria-hidden="true" /></header>
    <div class="thumbnail-list">
      <article v-for="file in presentationFiles" :key="`${file}:${revision || 0}`" class="thumbnail-card">
        <button v-if="!previewState(file)" type="button" class="thumbnail-preview" :aria-label="`放大预览 ${file}`" @click="openPreview(file)">
          <img :src="thumbnailURL(file)" alt="" @load="markPreviewLoaded(file)" @error="markPreviewError(file)">
          <span><Maximize2 :size="13" aria-hidden="true" />放大预览</span>
        </button>
        <div v-else class="thumbnail-unavailable" :class="{ 'thumbnail-unavailable--retrying': previewState(file) === 'retrying' }" role="status" aria-live="polite">
          <LoaderCircle v-if="previewState(file) === 'retrying'" :size="20" aria-hidden="true" />
          <ImageOff v-else :size="20" aria-hidden="true" />
          <span>{{ previewState(file) === 'retrying' ? '正在准备缩略图…' : '缩略图生成失败' }}</span>
          <button v-if="previewState(file) === 'failed'" type="button" class="thumbnail-retry" @click="resetPreview(file)">重试缩略图</button>
        </div>
        <footer><span>{{ file.replace(/\.pptx$/i, '') }}</span><a :href="taskDownloadUrl(task.id, file)" :aria-label="`下载 ${file}`"><Download :size="15" />下载</a></footer>
      </article>
    </div>
    <AppModal :open="Boolean(previewFile)" variant="media" :title="previewTitle" description="缩略图预览；可下载 PPT 获取完整演示文件。" @close="closePreview">
      <figure v-if="previewFile" class="preview-figure">
        <img :src="thumbnailURL(previewFile)" :alt="`${previewFile} 的放大预览`" @load="markPreviewLoaded(previewFile)" @error="markPreviewError(previewFile)">
        <figcaption><span>{{ previewFile.replace(/\.pptx$/i, '') }}</span><a :href="taskDownloadUrl(task.id, previewFile)"><Download :size="15" />下载 PPT</a></figcaption>
      </figure>
    </AppModal>
  </section>
</template>

<style scoped>
.delivery{max-width:760px;margin:0 auto 22px;border:1px solid var(--border-subtle);border-radius:8px;overflow:hidden;color:var(--text-strong);background:var(--surface-raised)}
.delivery--side-rail{max-width:none;margin:0}
.delivery-head{display:flex;align-items:flex-start;justify-content:space-between;gap:12px;padding:14px 16px;border-bottom:1px solid var(--border-subtle)}.delivery-head>div>span{display:block;font:700 15px 'Noto Serif SC',serif}.delivery-head p{margin:3px 0 0;color:var(--text-subtle);font-size:11px}.delivery-head>svg{color:var(--info)}
.thumbnail-list{display:grid;grid-template-columns:repeat(auto-fit,minmax(178px,1fr));gap:12px;padding:14px;overflow:auto}.thumbnail-card{min-width:0;overflow:hidden;border:1px solid var(--border-subtle);border-radius:6px;background:var(--surface-accent)}
.delivery--side-rail .thumbnail-list{grid-template-columns:1fr}
.thumbnail-preview{position:relative;display:block;width:100%;padding:0;border:0;color:var(--text-strong);background:var(--surface-base);text-align:left}.thumbnail-preview img,.thumbnail-unavailable{display:block;width:100%;aspect-ratio:16/9;object-fit:cover;background:var(--surface-base)}.thumbnail-preview>span{position:absolute;right:7px;bottom:7px;display:flex;align-items:center;gap:4px;padding:4px 6px;border-radius:4px;color:var(--text-strong);background:rgba(3,16,25,.76);font-size:10px;line-height:1}.thumbnail-preview:hover>span{color:var(--accent-on);background:var(--accent)}.thumbnail-preview:focus-visible,.thumbnail-retry:focus-visible{outline:2px solid var(--accent);outline-offset:-3px}.thumbnail-unavailable{display:grid;place-content:center;gap:7px;text-align:center;color:var(--text-subtle);font-size:12px}.thumbnail-unavailable svg{margin:auto}.thumbnail-unavailable--retrying svg{animation:thumbnail-spin 1s linear infinite}.thumbnail-retry{padding:3px 6px;border:1px solid var(--border-subtle);border-radius:4px;color:var(--accent);background:var(--surface-raised);font-size:11px;font-weight:600;cursor:pointer}.thumbnail-retry:hover{border-color:var(--accent);color:var(--accent-strong)}@keyframes thumbnail-spin{to{transform:rotate(360deg)}}@media (prefers-reduced-motion:reduce){.thumbnail-unavailable--retrying svg{animation:none}}
.thumbnail-card footer{display:flex;align-items:center;justify-content:space-between;gap:8px;padding:8px 9px;color:var(--text-muted);font-size:11px}.thumbnail-card footer>span{overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.thumbnail-card a{display:flex;align-items:center;gap:4px;flex:none;color:var(--accent);font-weight:600}.thumbnail-card a:hover{color:var(--accent-strong);text-decoration:underline}
.preview-figure{display:grid;gap:10px;margin:0}.preview-figure>img{display:block;max-width:100%;max-height:calc(100dvh - 195px);margin:auto;border:1px solid var(--border-subtle);border-radius:6px;background:var(--surface-base);object-fit:contain}.preview-figure figcaption{display:flex;align-items:center;justify-content:space-between;gap:12px;color:var(--text-muted);font-size:12px}.preview-figure figcaption span{overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.preview-figure figcaption a{display:flex;flex:none;align-items:center;gap:5px;color:var(--accent);font-weight:600}.preview-figure figcaption a:hover{text-decoration:underline}
</style>
