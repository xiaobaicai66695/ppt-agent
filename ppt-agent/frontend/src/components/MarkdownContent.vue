<script setup lang="ts">
import { computed, ref } from 'vue'
const props = defineProps<{ content: string }>()
const preview = ref('')
function safeUrl(value: string) { try { const url = new URL(value); return ['http:', 'https:'].includes(url.protocol) ? url.href : '' } catch { return '' } }
function esc(value: string) { return value.replace(/[&<>"']/g, char => ({ '&':'&amp;', '<':'&lt;', '>':'&gt;', '"':'&quot;', "'":'&#039;' }[char] || char)) }
const html = computed(() => {
  let output = esc(props.content || '')
  output = output.replace(/!\[([^\]]*)\]\(([^\s)]+)\)/g, (_, alt, url) => safeUrl(url) ? `<button class="md-image" data-image="${safeUrl(url)}"><img src="${safeUrl(url)}" alt="${esc(alt || '图片参考')}" loading="lazy"><span>${esc(alt || '点击预览图片')}</span></button>` : _)
  output = output.replace(/(?<!!)\[([^\]]+)\]\(([^\s)]+)\)/g, (_, label, url) => safeUrl(url) ? `<a href="${safeUrl(url)}" target="_blank" rel="noreferrer">${esc(label)}</a>` : _)
  output = output.replace(/`([^`]+)`/g, '<code>$1</code>').replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')
  return output.split(/\n{2,}/).map(block => {
    if (/^#{1,3}\s/.test(block)) { const heading = block.match(/^(#+)\s+([\s\S]*)/)!; return `<h${heading[1].length}>${heading[2]}</h${heading[1].length}>` }
    if (/^(?:[-*]\s+)/m.test(block)) return `<ul>${block.split('\n').filter(line => /^[-*]\s+/.test(line)).map(line => `<li>${line.replace(/^[-*]\s+/, '')}</li>`).join('')}</ul>`
    return `<p>${block.replace(/\n/g, '<br>')}</p>`
  }).join('')
})
function onClick(event: MouseEvent) { const button = (event.target as HTMLElement).closest<HTMLElement>('[data-image]'); if (button) preview.value = button.dataset.image || '' }
</script>
<template><div class="markdown" v-html="html" @click="onClick" /><div v-if="preview" class="lightbox" role="dialog" aria-label="图片预览" @click.self="preview = ''"><button @click="preview = ''">关闭预览</button><img :src="preview" alt="图片参考预览" /></div></template>
<style scoped>
.markdown :deep(p) { margin: 0 0 11px; line-height: 1.8; }.markdown :deep(h1),.markdown :deep(h2),.markdown :deep(h3) { margin: 16px 0 8px; line-height: 1.35; }.markdown :deep(a) { color: #0e8e84; text-decoration: underline; text-underline-offset: 3px; }.markdown :deep(code) { padding: 2px 5px; border-radius: 4px; color: #b6faf0; background: #163643; }.markdown :deep(ul) { margin: 0 0 12px; padding-left: 21px; }.markdown :deep(.md-image) { display: block; max-width: min(100%, 520px); margin: 12px 0; padding: 0; overflow: hidden; cursor: zoom-in; text-align: left; color: #55727f; background: #f0f5f3; border: 1px solid #d4e1dd; border-radius: 8px; }.markdown :deep(.md-image img) { display: block; width: 100%; max-height: 330px; object-fit: cover; }.markdown :deep(.md-image span) { display: block; padding: 6px 9px; font-size: 12px; }.lightbox { position: fixed; z-index: 30; inset: 0; display: grid; place-items: center; padding: 25px; background: rgba(0,11,19,.83); }.lightbox img { max-width: min(100%, 1100px); max-height: calc(100vh - 100px); border-radius: 6px; }.lightbox button { position: fixed; top: 18px; right: 20px; color: white; background: transparent; border: 1px solid rgba(255,255,255,.5); border-radius: 5px; padding: 8px 10px; }
</style>
