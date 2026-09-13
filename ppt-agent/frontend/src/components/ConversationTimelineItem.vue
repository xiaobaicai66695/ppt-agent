<script setup lang="ts">
import { computed } from 'vue'
import { BrainCircuit, CheckCircle2, ChevronDown, ChevronRight, CircleAlert, CircleDashed, LoaderCircle } from 'lucide-vue-next'
import MarkdownContent from './MarkdownContent.vue'
import type { ConversationTimelineItem, ExecutionState } from '../utils/conversationTimeline'

const props = defineProps<{ item: ConversationTimelineItem }>()
const emit = defineEmits<{ toggle: [itemID: string] }>()

const detailID = computed(() => `timeline-detail-${props.item.id}`)

function stateLabel(state: ExecutionState) {
  return ({ running: '执行中', success: '已完成', error: '失败' } as Record<ExecutionState, string>)[state]
}

function formatPayload(value?: string) {
  const source = value?.trim()
  if (!source) return ''
  try {
    return JSON.stringify(JSON.parse(source), null, 2)
  } catch {
    return source
  }
}

function toggle() {
  emit('toggle', props.item.id)
}
</script>

<template>
  <article v-if="item.type === 'message'" class="timeline-item message" :class="item.message.role" role="listitem">
    <div class="message-label">{{ item.message.role === 'user' ? '你' : 'PPTform' }}</div>
    <MarkdownContent :content="item.message.content" />
  </article>

  <article v-else-if="item.type === 'thought'" class="timeline-item trace-card thought-card" :class="item.state" role="listitem">
    <button type="button" class="trace-toggle" :aria-expanded="item.expanded" :aria-controls="detailID" @click="toggle">
      <span class="trace-icon thought-icon"><BrainCircuit :size="16" aria-hidden="true" /></span>
      <span class="trace-copy"><small>思考摘要</small><b>{{ item.phase || '正在推进下一步' }}</b></span>
      <span class="trace-status">{{ stateLabel(item.state) }}</span>
      <ChevronDown v-if="item.expanded" :size="16" aria-hidden="true" />
      <ChevronRight v-else :size="16" aria-hidden="true" />
    </button>
    <div v-if="item.expanded" :id="detailID" class="trace-detail thought-detail" :aria-busy="item.streaming || undefined">
      <p>{{ item.content }}</p>
      <span v-if="item.streaming" class="streaming-mark">正在更新</span>
    </div>
  </article>

  <article v-else-if="item.type === 'tool_call'" class="timeline-item trace-card tool-pair-card" :class="item.state" role="listitem">
    <button type="button" class="trace-toggle" :aria-expanded="item.expanded" :aria-controls="detailID" @click="toggle">
      <span class="trace-icon" :class="item.state"><LoaderCircle v-if="item.state === 'running'" :size="16" aria-hidden="true" /><CheckCircle2 v-else-if="item.state === 'success'" :size="16" aria-hidden="true" /><CircleAlert v-else :size="16" aria-hidden="true" /></span>
      <span class="trace-copy"><small>工具调用</small><b>{{ item.label }}</b></span>
      <span class="trace-status">{{ stateLabel(item.state) }}</span>
      <ChevronDown v-if="item.expanded" :size="16" aria-hidden="true" />
      <ChevronRight v-else :size="16" aria-hidden="true" />
    </button>
    <div v-if="item.expanded" :id="detailID" class="trace-detail tool-detail" :aria-busy="item.state === 'running' || undefined">
      <section class="tool-call-panel">
        <p v-if="item.detail" class="detail-copy"><span>调用说明</span>{{ item.detail }}</p>
        <div v-if="item.args" class="payload-block">
          <span>参数</span>
          <pre>{{ formatPayload(item.args) }}</pre>
        </div>
        <p v-if="!item.detail && !item.args && item.state === 'running'" class="detail-copy"><span>当前状态</span>正在等待工具返回结果</p>
      </section>
      <section class="tool-result-panel">
        <div class="tool-result-head">
          <span>返回内容</span>
          <small>{{ item.state === 'running' ? '等待中' : stateLabel(item.state) }}</small>
        </div>
        <pre class="tool-result-pre">{{ formatPayload(item.result || (item.state === 'running' ? '工具结果尚未返回' : '工具调用未返回结果')) }}</pre>
      </section>
      <div v-if="item.preview?.images?.length" class="tool-image-preview">
        <a v-for="(image, index) in item.preview.images" :key="image.image_url || image.thumbnail_url || index" :href="image.source_url || image.image_url || image.thumbnail_url" target="_blank" rel="noopener noreferrer">
          <img :src="image.thumbnail_url || image.image_url" :alt="image.alt || '图片工具结果预览'" width="136" height="88">
          <small v-if="image.attribution">{{ image.attribution }}</small>
        </a>
      </div>
    </div>
  </article>

  <article v-else-if="item.type === 'tool_batch'" class="timeline-item trace-card tool-batch-card" :class="item.state" role="listitem">
    <button type="button" class="trace-toggle" :aria-expanded="item.expanded" :aria-controls="detailID" @click="toggle">
      <span class="trace-icon" :class="item.state"><LoaderCircle v-if="item.state === 'running'" :size="16" aria-hidden="true" /><CheckCircle2 v-else-if="item.state === 'success'" :size="16" aria-hidden="true" /><CircleAlert v-else :size="16" aria-hidden="true" /></span>
      <span class="trace-copy"><small>工具调用</small><b>{{ item.tools.length }} 项工具调用</b></span>
      <span class="trace-status">{{ stateLabel(item.state) }}</span>
      <ChevronDown v-if="item.expanded" :size="16" aria-hidden="true" />
      <ChevronRight v-else :size="16" aria-hidden="true" />
    </button>
    <div v-if="item.expanded" :id="detailID" class="trace-detail tool-batch-list" :aria-busy="item.state === 'running' || undefined">
      <section v-for="tool in item.tools" :key="tool.id" class="tool-batch-entry">
        <div class="tool-result-head"><span>{{ tool.label }}</span><small>{{ stateLabel(tool.state) }}</small></div>
        <section class="tool-call-panel">
          <p v-if="tool.detail" class="detail-copy"><span>调用说明</span>{{ tool.detail }}</p>
          <div v-if="tool.args" class="payload-block"><span>参数</span><pre>{{ formatPayload(tool.args) }}</pre></div>
        </section>
        <section class="tool-result-panel">
          <div class="tool-result-head"><span>返回内容</span><small>{{ tool.state === 'running' ? '等待中' : stateLabel(tool.state) }}</small></div>
          <pre class="tool-result-pre">{{ formatPayload(tool.result || (tool.state === 'running' ? '工具结果尚未返回' : '工具调用未返回结果')) }}</pre>
        </section>
        <div v-if="tool.preview?.images?.length" class="tool-image-preview">
          <a v-for="(image, index) in tool.preview.images" :key="image.image_url || image.thumbnail_url || index" :href="image.source_url || image.image_url || image.thumbnail_url" target="_blank" rel="noopener noreferrer">
            <img :src="image.thumbnail_url || image.image_url" :alt="image.alt || '图片工具结果预览'" width="136" height="88">
            <small v-if="image.attribution">{{ image.attribution }}</small>
          </a>
        </div>
      </section>
    </div>
  </article>

  <article v-else-if="item.type === 'final_answer'" class="timeline-item message assistant final-answer" role="listitem">
    <div class="message-label">最终回答</div>
    <MarkdownContent :content="item.content" :streaming="item.streaming" :live="false" />
  </article>

  <article v-else-if="item.type === 'error'" class="timeline-item error-card" role="listitem">
    <div class="error-content" role="alert">
      <CircleAlert :size="16" aria-hidden="true" />
      <div><small>需要处理</small><p>{{ item.content }}</p></div>
    </div>
  </article>

  <article v-else class="timeline-item execution-event" :class="item.state" role="listitem">
    <span class="trace-icon" :class="item.state"><LoaderCircle v-if="item.state === 'running'" :size="14" aria-hidden="true" /><CheckCircle2 v-else-if="item.state === 'success'" :size="14" aria-hidden="true" /><CircleDashed v-else :size="14" aria-hidden="true" /></span>
    <b>{{ item.label }}</b>
    <em v-if="item.detail">{{ item.detail }}</em>
  </article>
</template>

<style scoped>
.timeline-item{max-width:760px;margin:0 auto 16px}.message{padding:18px 20px;border-radius:3px 13px 13px 13px}.message.user{max-width:650px;margin-right:0;color:var(--text-on-accent);background:var(--message-user)}.message.assistant{color:var(--text-strong);background:var(--message-assistant);border:1px solid var(--border-subtle)}.message-label{margin-bottom:10px;font:500 10px 'DM Mono',monospace;letter-spacing:.08em;color:var(--accent-strong)}.message.user .message-label{color:var(--message-user-label)}.final-answer{border-left:3px solid var(--accent)}
.trace-card{position:relative;overflow:hidden;border:1px solid var(--border-subtle);border-left-width:3px;border-radius:8px;background:var(--surface-accent)}.trace-card::before{position:absolute;top:-17px;bottom:100%;left:15px;width:1px;background:var(--border-strong);content:''}.thought-card{border-left-color:var(--text-subtle)}.tool-pair-card.running,.tool-batch-card.running{border-left-color:var(--info)}.tool-pair-card.success,.tool-batch-card.success{border-left-color:var(--accent)}.tool-pair-card.error,.tool-batch-card.error{border-left-color:var(--danger)}
.trace-toggle{display:grid;grid-template-columns:auto minmax(0,1fr) auto auto;align-items:center;width:100%;gap:9px;padding:10px 11px;border:0;color:var(--text-muted);background:transparent;text-align:left}.trace-toggle:hover{background:var(--surface-hover)}.trace-toggle:active{background:var(--surface-raised)}.trace-copy{display:grid;gap:2px;min-width:0}.trace-copy small,.payload-block>span,.detail-copy span,.streaming-mark,.error-card small{color:var(--text-subtle);font:500 10px 'DM Mono',monospace;letter-spacing:.06em}.trace-copy b{overflow:hidden;color:var(--text-strong);font-size:12px;text-overflow:ellipsis;white-space:nowrap}.trace-status{color:var(--text-subtle);font-size:11px;white-space:nowrap}.trace-icon{display:grid;flex:none;place-items:center;color:var(--accent)}.thought-icon{color:var(--text-muted)}.trace-icon.running svg{animation:spin 1s linear infinite}.trace-icon.error{color:var(--danger)}.trace-icon.success{color:var(--accent)}
.trace-detail{display:grid;gap:10px;padding:10px 12px;border-top:1px solid var(--border-subtle);color:var(--text-muted);background:color-mix(in srgb,var(--surface-raised) 88%,transparent);font-size:12px;line-height:1.65}.thought-detail p,.detail-copy{margin:0;white-space:pre-wrap;overflow-wrap:anywhere}.detail-copy span{display:block;margin-bottom:2px}.tool-call-panel,.tool-result-panel{display:grid;gap:7px}.tool-batch-list{gap:8px}.tool-batch-entry{display:grid;gap:8px;padding:10px;border:1px solid var(--border-subtle);border-radius:7px;background:var(--surface-base)}.payload-block{min-width:0}.payload-block>span{display:block;margin-bottom:4px}.payload-block pre,.tool-result-pre{margin:0;padding:9px 10px;border:1px solid var(--border-subtle);border-radius:6px;color:var(--text-strong);background:var(--surface-raised);font:400 11px/1.55 'DM Mono',monospace;white-space:pre-wrap;overflow-wrap:anywhere;max-height:196px;overflow:auto;overscroll-behavior:contain;scrollbar-gutter:stable}.tool-result-panel{padding:10px;border:1px solid var(--border-subtle);border-radius:7px;background:var(--surface-base)}.tool-result-head{display:flex;align-items:center;justify-content:space-between;gap:10px;color:var(--text-subtle);font:500 10px 'DM Mono',monospace;letter-spacing:.06em}.streaming-mark{display:inline-flex;align-items:center;width:max-content;color:var(--accent)}.streaming-mark::before{width:5px;height:5px;margin-right:5px;border-radius:50%;background:currentColor;content:''}.tool-image-preview{display:flex;gap:8px;overflow:auto;padding-bottom:2px}.tool-image-preview a{display:grid;gap:3px;min-width:136px;color:var(--text-muted);font-size:10px;line-height:1.35}.tool-image-preview img{display:block;width:136px;height:88px;border:1px solid var(--border-subtle);border-radius:4px;object-fit:cover;background:var(--surface-hover)}.tool-image-preview a:focus-visible{outline:2px solid var(--accent);outline-offset:2px}
.execution-event{display:flex;align-items:center;gap:7px;padding:3px;color:var(--text-muted);font-size:12px}.execution-event b{font-weight:600}.execution-event em{overflow:hidden;color:var(--text-subtle);font-style:normal;text-overflow:ellipsis;white-space:nowrap}.error-card{padding:11px 12px;border:1px solid color-mix(in srgb,var(--danger) 65%,var(--border-subtle));border-radius:8px;color:var(--danger);background:color-mix(in srgb,var(--danger) 9%,var(--surface-raised))}.error-content{display:flex;align-items:flex-start;gap:9px}.error-card p{margin:2px 0 0;color:var(--text-strong);font-size:12px;line-height:1.6;overflow-wrap:anywhere}
@keyframes spin{to{transform:rotate(360deg)}}@media (prefers-reduced-motion: reduce){.trace-icon.running svg{animation:none}}@media(max-width:850px){.message{padding:15px}.trace-toggle{grid-template-columns:auto minmax(0,1fr) auto}.trace-status{grid-column:2;grid-row:2}.trace-toggle>svg{grid-column:3;grid-row:1 / span 2}.payload-block pre{max-height:210px}}
</style>
