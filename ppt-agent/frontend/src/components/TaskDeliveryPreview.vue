<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { Download, ImageOff, Maximize2, Presentation } from 'lucide-vue-next'
import AppModal from './AppModal.vue'
import { taskDownloadUrl, taskThumbUrl } from '../api'
import type { TaskInfo } from '../types'

const props = withDefaults(defineProps<{
  task: TaskInfo
  revision?: number
  layout?: 'inline' | 'side-rail'
}>(), { layout: 'inline' })
const brokenPreviews = ref<string[]>([])
const previewFile = ref('')
const presentationFiles = computed(() => (props.task.files || []).filter(file => file.toLowerCase().endsWith('.pptx')))
const previewTitle = computed(() => previewFile.value ? `${previewFile.value.replace(/\.pptx$/i, '')} · 页面预览` : '页面预览')
const isBroken = (file: string) => brokenPreviews.value.includes(file)
function markBroken(file: string) { if (!isBroken(file)) brokenPreviews.value = [...brokenPreviews.value, file] }
function openPreview(file: string) { previewFile.value = file }
function closePreview() { previewFile.value = '' }
watch(() => [props.task.id, props.revision], () => { brokenPreviews.value = []; closePreview() })
</script>

<template>
  <section v-if="presentationFiles.length" class="delivery" :class="{ 'delivery--side-rail': layout === 'side-rail' }" aria-label="已交付演示预览">
    <header class="delivery-head"><div><span>交付预览</span><p>缩略图来自已生成的 PPT 页面</p></div><Presentation :size="19" aria-hidden="true" /></header>
    <div class="thumbnail-list">
      <article v-for="file in presentationFiles" :key="`${file}:${revision || 0}`" class="thumbnail-card">
        <button v-if="!isBroken(file)" type="button" class="thumbnail-preview" :aria-label="`放大预览 ${file}`" @click="openPreview(file)">
          <img :src="taskThumbUrl(task.id, file)" alt="" @error="markBroken(file)">
          <span><Maximize2 :size="13" aria-hidden="true" />放大预览</span>
        </button>
        <div v-else class="thumbnail-unavailable"><ImageOff :size="20" /><span>缩略图暂不可用</span></div>
        <footer><span>{{ file.replace(/\.pptx$/i, '') }}</span><a :href="taskDownloadUrl(task.id, file)" :aria-label="`下载 ${file}`"><Download :size="15" />下载</a></footer>
      </article>
    </div>
    <AppModal :open="Boolean(previewFile)" variant="media" :title="previewTitle" description="缩略图预览；可下载 PPT 获取完整演示文件。" @close="closePreview">
      <figure v-if="previewFile" class="preview-figure">
        <img :src="taskThumbUrl(task.id, previewFile)" :alt="`${previewFile} 的放大预览`" @error="markBroken(previewFile)">
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
.thumbnail-preview{position:relative;display:block;width:100%;padding:0;border:0;color:var(--text-strong);background:var(--surface-base);text-align:left}.thumbnail-preview img,.thumbnail-unavailable{display:block;width:100%;aspect-ratio:16/9;object-fit:cover;background:var(--surface-base)}.thumbnail-preview>span{position:absolute;right:7px;bottom:7px;display:flex;align-items:center;gap:4px;padding:4px 6px;border-radius:4px;color:var(--text-strong);background:rgba(3,16,25,.76);font-size:10px;line-height:1}.thumbnail-preview:hover>span{color:var(--accent-on);background:var(--accent)}.thumbnail-preview:focus-visible{outline:2px solid var(--accent);outline-offset:-3px}.thumbnail-unavailable{display:grid;place-content:center;gap:7px;text-align:center;color:var(--text-subtle);font-size:12px}.thumbnail-unavailable svg{margin:auto}
.thumbnail-card footer{display:flex;align-items:center;justify-content:space-between;gap:8px;padding:8px 9px;color:var(--text-muted);font-size:11px}.thumbnail-card footer>span{overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.thumbnail-card a{display:flex;align-items:center;gap:4px;flex:none;color:var(--accent);font-weight:600}.thumbnail-card a:hover{color:var(--accent-strong);text-decoration:underline}
.preview-figure{display:grid;gap:10px;margin:0}.preview-figure>img{display:block;max-width:100%;max-height:calc(100dvh - 195px);margin:auto;border:1px solid var(--border-subtle);border-radius:6px;background:var(--surface-base);object-fit:contain}.preview-figure figcaption{display:flex;align-items:center;justify-content:space-between;gap:12px;color:var(--text-muted);font-size:12px}.preview-figure figcaption span{overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.preview-figure figcaption a{display:flex;flex:none;align-items:center;gap:5px;color:var(--accent);font-weight:600}.preview-figure figcaption a:hover{text-decoration:underline}
</style>
