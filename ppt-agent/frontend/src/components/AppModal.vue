<script setup lang="ts">
import { nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { X } from 'lucide-vue-next'

const props = defineProps<{ open: boolean; title: string; description?: string }>()
const emit = defineEmits<{ close: [] }>()

const dialog = ref<HTMLElement>()
let previousFocus: HTMLElement | null = null

function focusableElements() {
  return Array.from(dialog.value?.querySelectorAll<HTMLElement>(
    'button:not([disabled]), [href], textarea:not([disabled]), input:not([disabled]), select:not([disabled]), [tabindex]:not([tabindex="-1"])',
  ) || [])
}

function close() { emit('close') }

function handleKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') { event.preventDefault(); close(); return }
  if (event.key !== 'Tab') return
  const elements = focusableElements()
  if (!elements.length) { event.preventDefault(); return }
  const first = elements[0]
  const last = elements[elements.length - 1]
  if (event.shiftKey && document.activeElement === first) { event.preventDefault(); last.focus() }
  else if (!event.shiftKey && document.activeElement === last) { event.preventDefault(); first.focus() }
}

watch(() => props.open, async open => {
  document.removeEventListener('keydown', handleKeydown)
  if (!open) { previousFocus?.focus(); previousFocus = null; return }
  previousFocus = document.activeElement instanceof HTMLElement ? document.activeElement : null
  document.addEventListener('keydown', handleKeydown)
  await nextTick()
  focusableElements()[0]?.focus()
}, { immediate: true })

onBeforeUnmount(() => document.removeEventListener('keydown', handleKeydown))
</script>

<template>
  <Teleport to="body">
    <div v-if="open" class="modal-backdrop" @mousedown.self="close">
      <section ref="dialog" class="modal" role="dialog" aria-modal="true" :aria-label="title">
        <header class="modal-head">
          <div><h2>{{ title }}</h2><p v-if="description">{{ description }}</p></div>
          <button class="modal-close" type="button" aria-label="关闭弹窗" @click="close"><X :size="18" /></button>
        </header>
        <div class="modal-content"><slot /></div>
      </section>
    </div>
  </Teleport>
</template>

<style scoped>
.modal-backdrop{position:fixed;z-index:50;inset:0;display:grid;place-items:center;padding:20px;background:rgba(3,16,25,.62)}
.modal{width:min(100%,460px);max-height:min(680px,calc(100vh - 40px));overflow:auto;border:1px solid var(--border-strong);border-radius:10px;color:var(--text-strong);background:var(--surface-raised);box-shadow:0 22px 70px rgba(0,0,0,.38)}
.modal-head{display:flex;align-items:flex-start;justify-content:space-between;gap:16px;padding:19px 20px 15px;border-bottom:1px solid var(--border-subtle)}
.modal-head h2{margin:0;color:var(--text-strong);font:700 19px 'Noto Serif SC',serif}.modal-head p{margin:5px 0 0;color:var(--text-subtle);font-size:12px;line-height:1.55}
.modal-close{display:grid;place-items:center;flex:none;width:31px;height:31px;border:1px solid var(--border-strong);border-radius:5px;color:var(--text-muted);background:transparent}.modal-close:hover{color:var(--text-strong);background:var(--surface-hover)}
.modal-content{padding:18px 20px 20px}
@media(max-width:520px){.modal-backdrop{padding:12px}.modal-head{padding:16px 16px 13px}.modal-content{padding:15px 16px 16px}}
</style>
