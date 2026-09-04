<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { Download, ImageOff, Presentation } from 'lucide-vue-next'
import { taskDownloadUrl, taskThumbUrl } from '../api'
import type { TaskInfo } from '../types'

const props = defineProps<{ task: TaskInfo; revision?: number }>()
const brokenPreviews = ref<string[]>([])
const presentationFiles = computed(() => (props.task.files || []).filter(file => file.toLowerCase().endsWith('.pptx')))
const isBroken = (file: string) => brokenPreviews.value.includes(file)
function markBroken(file: string) { if (!isBroken(file)) brokenPreviews.value = [...brokenPreviews.value, file] }
watch(() => [props.task.id, props.revision], () => { brokenPreviews.value = [] })
</script>

<template>
  <section v-if="presentationFiles.length" class="delivery" aria-label="已交付演示预览">
    <header class="delivery-head"><div><span>交付预览</span><p>缩略图来自已生成的 PPT 页面</p></div><Presentation :size="19" aria-hidden="true" /></header>
    <div class="thumbnail-list">
      <article v-for="file in presentationFiles" :key="`${file}:${revision || 0}`" class="thumbnail-card">
        <img v-if="!isBroken(file)" :src="taskThumbUrl(task.id, file)" :alt="`${file} 的缩略图`" @error="markBroken(file)">
        <div v-else class="thumbnail-unavailable"><ImageOff :size="20" /><span>缩略图暂不可用</span></div>
        <footer><span>{{ file.replace(/\.pptx$/i, '') }}</span><a :href="taskDownloadUrl(task.id, file)" :aria-label="`下载 ${file}`"><Download :size="15" />下载</a></footer>
      </article>
    </div>
  </section>
</template>

<style scoped>
.delivery{max-width:760px;margin:0 auto 22px;border:1px solid var(--border-subtle);border-radius:8px;overflow:hidden;color:var(--text-strong);background:var(--surface-raised)}
.delivery-head{display:flex;align-items:flex-start;justify-content:space-between;gap:12px;padding:14px 16px;border-bottom:1px solid var(--border-subtle)}.delivery-head>div>span{display:block;font:700 15px 'Noto Serif SC',serif}.delivery-head p{margin:3px 0 0;color:var(--text-subtle);font-size:11px}.delivery-head>svg{color:var(--info)}
.thumbnail-list{display:grid;grid-template-columns:repeat(auto-fit,minmax(178px,1fr));gap:12px;padding:14px;overflow:auto}.thumbnail-card{min-width:0;overflow:hidden;border:1px solid var(--border-subtle);border-radius:6px;background:var(--surface-accent)}
.thumbnail-card img,.thumbnail-unavailable{display:block;width:100%;aspect-ratio:16/9;object-fit:cover;background:var(--surface-base)}.thumbnail-unavailable{display:grid;place-content:center;gap:7px;text-align:center;color:var(--text-subtle);font-size:12px}.thumbnail-unavailable svg{margin:auto}
.thumbnail-card footer{display:flex;align-items:center;justify-content:space-between;gap:8px;padding:8px 9px;color:var(--text-muted);font-size:11px}.thumbnail-card footer>span{overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.thumbnail-card a{display:flex;align-items:center;gap:4px;flex:none;color:var(--accent);font-weight:600}.thumbnail-card a:hover{color:var(--accent-strong);text-decoration:underline}
</style>
