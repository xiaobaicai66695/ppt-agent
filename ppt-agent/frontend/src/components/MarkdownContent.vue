<script setup lang="ts">
import { computed, ref } from 'vue'
import { parseMarkdown } from '../utils/markdown'

const props = withDefaults(defineProps<{ content: string; streaming?: boolean }>(), { streaming: false })
const preview = ref('')
const blocks = computed(() => parseMarkdown(props.content))

function imageLabel(alt: string) { return alt || '点击预览图片' }
function openPreview(src: string) { preview.value = src }
</script>

<template>
  <div class="markdown" :aria-live="streaming ? 'polite' : undefined" :aria-busy="streaming || undefined">
    <template v-for="(block, blockIndex) in blocks" :key="blockIndex">
      <component :is="`h${block.level}`" v-if="block.type === 'heading'" class="markdown-heading">
        <template v-for="(token, tokenIndex) in block.content" :key="tokenIndex">
          <a v-if="token.type === 'link'" :href="token.href" target="_blank" rel="noopener noreferrer">{{ token.label }}</a>
          <code v-else-if="token.type === 'code'">{{ token.value }}</code>
          <strong v-else-if="token.type === 'strong'">{{ token.value }}</strong>
          <br v-else-if="token.type === 'break'">
          <span v-else-if="token.type === 'text'">{{ token.value }}</span>
        </template>
      </component>
      <p v-else-if="block.type === 'paragraph'">
        <template v-for="(token, tokenIndex) in block.content" :key="tokenIndex">
          <a v-if="token.type === 'link'" :href="token.href" target="_blank" rel="noopener noreferrer">{{ token.label }}</a>
          <button v-else-if="token.type === 'image'" class="inline-image-trigger" type="button" @click="openPreview(token.src)">{{ imageLabel(token.alt) }}</button>
          <code v-else-if="token.type === 'code'">{{ token.value }}</code>
          <strong v-else-if="token.type === 'strong'">{{ token.value }}</strong>
          <br v-else-if="token.type === 'break'">
          <span v-else-if="token.type === 'text'">{{ token.value }}</span>
        </template>
      </p>
      <ul v-else-if="block.type === 'list'">
        <li v-for="(item, itemIndex) in block.items" :key="itemIndex">
          <template v-for="(token, tokenIndex) in item" :key="tokenIndex">
            <a v-if="token.type === 'link'" :href="token.href" target="_blank" rel="noopener noreferrer">{{ token.label }}</a>
            <code v-else-if="token.type === 'code'">{{ token.value }}</code>
            <strong v-else-if="token.type === 'strong'">{{ token.value }}</strong>
            <br v-else-if="token.type === 'break'">
            <span v-else-if="token.type === 'text'">{{ token.value }}</span>
          </template>
        </li>
      </ul>
      <figure v-else class="md-image">
        <button type="button" :aria-label="`预览：${imageLabel(block.alt)}`" @click="openPreview(block.src)">
          <img :src="block.src" :alt="block.alt || '图片参考'" loading="lazy">
        </button>
        <figcaption>{{ imageLabel(block.alt) }}</figcaption>
      </figure>
    </template>
    <span v-if="streaming" class="streaming-cursor" aria-label="正在生成" />
  </div>
  <div v-if="preview" class="lightbox" role="dialog" aria-modal="true" aria-label="图片预览" @click.self="preview = ''">
    <button type="button" @click="preview = ''">关闭预览</button>
    <img :src="preview" alt="图片参考预览">
  </div>
</template>

<style scoped>
.markdown :deep(p) { margin: 0 0 11px; line-height: 1.8; overflow-wrap: anywhere; }.markdown-heading { margin: 16px 0 8px; line-height: 1.35; }.markdown a { color: #0e8e84; text-decoration: underline; text-underline-offset: 3px; overflow-wrap: anywhere; }.markdown code { padding: 2px 5px; border-radius: 4px; color: #b6faf0; background: #163643; }.markdown ul { margin: 0 0 12px; padding-left: 21px; }.md-image { display: block; max-width: min(100%, 520px); margin: 12px 0; overflow: hidden; color: #55727f; background: #f0f5f3; border: 1px solid #d4e1dd; border-radius: 8px; }.md-image button { display: block; width: 100%; padding: 0; cursor: zoom-in; border: 0; background: transparent; }.md-image img { display: block; width: 100%; max-height: 330px; object-fit: cover; }.md-image figcaption { padding: 6px 9px; font-size: 12px; }.inline-image-trigger { padding: 0; color: #0e8e84; cursor: zoom-in; text-decoration: underline; text-underline-offset: 3px; background: transparent; border: 0; }.streaming-cursor { display: inline-block; width: 7px; height: 1.1em; margin-left: 2px; vertical-align: -0.18em; background: #0e8e84; animation: blink 1s step-end infinite; }.lightbox { position: fixed; z-index: 30; inset: 0; display: grid; place-items: center; padding: 25px; background: rgba(0,11,19,.83); }.lightbox img { max-width: min(100%, 1100px); max-height: calc(100vh - 100px); border-radius: 6px; }.lightbox button { position: fixed; top: 18px; right: 20px; color: white; background: transparent; border: 1px solid rgba(255,255,255,.5); border-radius: 5px; padding: 8px 10px; }@keyframes blink { 50% { opacity: 0; } }@media (prefers-reduced-motion: reduce) { .streaming-cursor { animation: none; } }
</style>
