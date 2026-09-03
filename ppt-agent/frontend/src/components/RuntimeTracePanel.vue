<script setup lang="ts">
import { computed, ref } from 'vue';
import { Activity, Bot, ChevronDown, CircleAlert, CircleCheck, LoaderCircle, Wrench } from 'lucide-vue-next';
import type { RuntimeEvent, RuntimeMeta } from '../types';
import {
  runtimeEventDetailLabel,
  runtimeEventKindLabel,
  runtimeEventNameLabel,
  runtimeEventStatusLabel,
} from '../utils/workbench';
import RuntimeEventDetail from './RuntimeEventDetail.vue';

const props = defineProps<{
  events: RuntimeEvent[];
  runtimeMeta?: RuntimeMeta | null;
  selectedEvent?: RuntimeEvent | null;
  loading?: boolean;
  error?: string;
  done?: number;
  total?: number;
}>();

const emit = defineEmits<{
  select: [event: RuntimeEvent];
  close: [];
}>();

type TraceFilter = 'all' | 'model' | 'tool' | 'delivery' | 'issue';
const filter = ref<TraceFilter>('all');

function category(event: RuntimeEvent): TraceFilter {
  const kind = (event.kind || '').toLowerCase();
  const name = (event.name || '').toLowerCase();
  const status = (event.status || '').toLowerCase();
  if (status === 'error' || status === 'failed' || kind.includes('error')) return 'issue';
  if (kind.startsWith('llm') || kind === 'model_request' || kind === 'assistant_output' || name.includes('chatmodel')) return 'model';
  if (kind.startsWith('tool') || kind.startsWith('slide_render') || name === 'toolnode') return 'tool';
  return 'delivery';
}

function categoryLabel(value: TraceFilter): string {
  return ({ all: '全部', model: '模型', tool: '工具', delivery: '交付', issue: '异常' })[value];
}

const counts = computed(() => {
  const result: Record<TraceFilter, number> = { all: props.events.length, model: 0, tool: 0, delivery: 0, issue: 0 };
  props.events.forEach(event => { result[category(event)] += 1; });
  return result;
});
const visibleEvents = computed(() => props.events.filter(event => filter.value === 'all' || category(event) === filter.value));
const currentPhase = computed(() => props.runtimeMeta?.phase_detail || props.runtimeMeta?.phase || '等待任务启动');
const issueCount = computed(() => counts.value.issue + (props.runtimeMeta?.budget_warnings?.length || 0));

function iconFor(event: RuntimeEvent) {
  const kind = category(event);
  if (kind === 'issue') return CircleAlert;
  if (kind === 'model') return Bot;
  if (kind === 'tool') return Wrench;
  return CircleCheck;
}

function eventState(event: RuntimeEvent): string {
  const status = (event.status || '').toLowerCase();
  if (status === 'running') return 'running';
  if (status === 'error' || status === 'failed') return 'issue';
  return category(event);
}

function elapsed(ms?: number): string {
  const value = Number(ms || 0);
  if (!value || value < 0) return '刚刚';
  const seconds = Math.floor(value / 1000);
  if (seconds < 60) return `${seconds}s`;
  return `${Math.floor(seconds / 60)}m ${seconds % 60}s`;
}
</script>

<template>
  <section v-if="events.length || runtimeMeta" class="runtime-trace" aria-label="执行轨迹">
    <header class="trace-head">
      <div class="trace-title">
        <span class="trace-mark"><Activity :size="17" /></span>
        <span><strong>执行轨迹</strong><small>{{ currentPhase }}</small></span>
      </div>
      <span v-if="issueCount" class="trace-alert"><CircleAlert :size="14" />{{ issueCount }} 项需留意</span>
      <span v-else class="trace-state"><LoaderCircle v-if="runtimeMeta?.phase" :size="14" class="spin" />过程正常</span>
    </header>

    <div class="trace-overview">
      <span><strong>{{ done || runtimeMeta?.done_slides || 0 }}</strong>/<span>{{ total || runtimeMeta?.total_slides || '—' }}</span> 页</span>
      <span><strong>{{ counts.tool }}</strong> 次工具调用</span>
      <span><strong>{{ counts.model }}</strong> 个模型节点</span>
    </div>

    <div class="trace-filters" role="tablist" aria-label="筛选执行事件">
      <button
        v-for="option in (['all', 'model', 'tool', 'delivery', 'issue'] as TraceFilter[])"
        :key="option"
        type="button"
        :class="{ active: filter === option }"
        :aria-selected="filter === option"
        :disabled="counts[option] === 0"
        @click="filter = option"
      >{{ categoryLabel(option) }} <span>{{ counts[option] }}</span></button>
    </div>

    <ol v-if="visibleEvents.length" class="trace-list">
      <li v-for="event in visibleEvents" :key="event.id || `${event.timestamp}-${event.kind}`">
        <button
          type="button"
          class="trace-event"
          :class="[eventState(event), { selected: selectedEvent?.id === event.id }]"
          @click="emit('select', event)"
        >
          <span class="trace-icon"><component :is="iconFor(event)" :size="15" /></span>
          <span class="trace-copy">
            <strong>{{ runtimeEventNameLabel(event) }}</strong>
            <small>{{ runtimeEventDetailLabel(event) }}</small>
          </span>
          <span class="trace-meta">
            <small>{{ elapsed(event.elapsed_ms) }}</small>
            <em>{{ runtimeEventStatusLabel(event.status) }}</em>
          </span>
        </button>
      </li>
    </ol>
    <p v-else class="trace-empty">此分类暂时没有事件。</p>

    <div v-if="selectedEvent" class="trace-detail">
      <div class="trace-detail-head">
        <span><strong>{{ runtimeEventKindLabel(selectedEvent) }}</strong><small>{{ runtimeEventNameLabel(selectedEvent) }}</small></span>
        <button type="button" aria-label="关闭事件详情" @click="emit('close')"><ChevronDown :size="16" /></button>
      </div>
      <RuntimeEventDetail :event="selectedEvent" :loading="loading" :error="error" />
    </div>
  </section>
</template>

<style scoped>
.runtime-trace { overflow: hidden; border: 1px solid var(--border); border-radius: var(--radius-lg); background: var(--surface); box-shadow: var(--shadow-xs); }
.trace-head { min-height: 62px; padding: 11px 15px; display: flex; align-items: center; justify-content: space-between; gap: 12px; border-bottom: 1px solid var(--divider); }
.trace-title { min-width: 0; display: flex; align-items: center; gap: 10px; }.trace-title > span:last-child { min-width: 0; display: grid; gap: 2px; }.trace-title strong { color: var(--text); font-size: 14px; }.trace-title small { overflow: hidden; color: var(--text-muted); font-size: 11px; text-overflow: ellipsis; white-space: nowrap; }
.trace-mark { width: 32px; height: 32px; display: grid; place-items: center; border-radius: 9px; color: var(--action-ink); background: var(--action-soft); }.trace-state,.trace-alert { flex: 0 0 auto; display: inline-flex; align-items: center; gap: 5px; font-size: 11px; }.trace-state { color: var(--success); }.trace-alert { color: var(--danger); }
.trace-overview { padding: 10px 15px; display: flex; flex-wrap: wrap; gap: 7px 18px; color: var(--text-muted); border-bottom: 1px solid var(--divider); font-size: 11px; }.trace-overview strong { color: var(--text); font-size: 13px; }
.trace-filters { padding: 9px 12px; display: flex; gap: 5px; overflow-x: auto; border-bottom: 1px solid var(--divider); background: var(--surface-muted); }.trace-filters button { min-height: 29px; padding: 0 8px; border: 1px solid transparent; border-radius: var(--radius-sm); color: var(--text-secondary); background: transparent; font-size: 11px; cursor: pointer; white-space: nowrap; }.trace-filters button span { margin-left: 2px; color: var(--text-muted); }.trace-filters button.active { border-color: #a8cbc8; color: var(--action-ink); background: var(--surface); }.trace-filters button:disabled { opacity: .42; cursor: default; }
.trace-list { max-height: 366px; margin: 0; padding: 0; overflow: auto; list-style: none; }.trace-event { width: 100%; min-height: 58px; padding: 9px 14px; display: grid; grid-template-columns: 28px minmax(0, 1fr) max-content; align-items: center; gap: 10px; border: 0; border-bottom: 1px solid var(--divider); color: inherit; background: var(--surface); text-align: left; cursor: pointer; }.trace-event:hover,.trace-event.selected { background: #f3f8f7; }.trace-event.selected { box-shadow: inset 3px 0 0 var(--action-ink); }.trace-icon { width: 27px; height: 27px; display: grid; place-items: center; border-radius: 50%; color: var(--success); background: var(--success-soft); }.trace-event.model .trace-icon { color: var(--info); background: var(--info-soft); }.trace-event.tool .trace-icon { color: var(--action-ink); background: var(--action-soft); }.trace-event.issue .trace-icon { color: var(--danger); background: var(--danger-soft); }.trace-event.running .trace-icon { color: var(--info); background: var(--info-soft); }.trace-copy { min-width: 0; display: grid; gap: 2px; }.trace-copy strong { overflow: hidden; color: var(--text); font-size: 12px; text-overflow: ellipsis; white-space: nowrap; }.trace-copy small { overflow: hidden; color: var(--text-muted); font-size: 11px; line-height: 1.4; text-overflow: ellipsis; white-space: nowrap; }.trace-meta { display: grid; justify-items: end; gap: 3px; }.trace-meta small { color: var(--text-muted); font-size: 10px; }.trace-meta em { color: var(--text-secondary); font-size: 10px; font-style: normal; }.trace-event.issue .trace-meta em { color: var(--danger); }.trace-empty { padding: 28px 15px; margin: 0; color: var(--text-muted); font-size: 12px; text-align: center; }
.trace-detail { border-top: 1px solid var(--border-strong); background: var(--surface-muted); }.trace-detail-head { min-height: 45px; padding: 8px 12px 8px 15px; display: flex; align-items: center; justify-content: space-between; gap: 10px; }.trace-detail-head span { display: grid; gap: 1px; }.trace-detail-head strong { color: var(--text); font-size: 12px; }.trace-detail-head small { color: var(--text-muted); font-size: 10px; }.trace-detail-head button { width: 30px; height: 30px; display: grid; place-items: center; border: 0; border-radius: var(--radius-sm); color: var(--text-secondary); background: transparent; cursor: pointer; }.trace-detail-head button:hover { background: var(--surface-pressed); }.trace-detail-head button svg { transform: rotate(180deg); }
.spin { animation: spin .9s linear infinite; } @keyframes spin { to { transform: rotate(360deg); } }
@media (max-width: 600px) { .trace-head { align-items: flex-start; }.trace-state { margin-top: 4px; }.trace-event { padding: 9px 11px; grid-template-columns: 27px minmax(0, 1fr); }.trace-meta { display: none; }.trace-copy small { white-space: normal; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; }.trace-list { max-height: 320px; } }
@media (prefers-reduced-motion: reduce) { .spin { animation: none; } }
</style>
