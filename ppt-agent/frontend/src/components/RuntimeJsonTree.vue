<script setup lang="ts">
import { computed, ref } from 'vue';

defineOptions({ name: 'RuntimeJsonTree' });

const props = withDefaults(defineProps<{
  label?: string;
  value: unknown;
  depth?: number;
  defaultOpen?: boolean;
}>(), {
  label: '',
  depth: 0,
  defaultOpen: false,
});

const open = ref(props.defaultOpen || props.depth < 1);

function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === 'object' && !Array.isArray(value);
}

const expandable = computed(() => Array.isArray(props.value) || isRecord(props.value));

const entries = computed(() => {
  if (Array.isArray(props.value)) {
    return props.value.map((value, index) => ({ key: String(index), value }));
  }
  if (isRecord(props.value)) {
    return Object.entries(props.value).map(([key, value]) => ({ key, value }));
  }
  return [];
});

const preview = computed(() => {
  if (Array.isArray(props.value)) return `Array(${props.value.length})`;
  if (isRecord(props.value)) {
    const keys = Object.keys(props.value);
    return keys.length ? `{ ${keys.slice(0, 4).join(', ')}${keys.length > 4 ? ', ...' : ''} }` : '{}';
  }
  return formatPrimitive(props.value);
});

function formatPrimitive(value: unknown): string {
  if (value === null) return 'null';
  if (value === undefined) return 'undefined';
  if (typeof value === 'string') return value;
  if (typeof value === 'number' || typeof value === 'boolean') return String(value);
  return JSON.stringify(value);
}

function primitiveClass(value: unknown): string {
  if (value === null || value === undefined) return 'nil';
  return typeof value;
}
</script>

<template>
  <div class="json-node" :style="{ '--depth': depth }">
    <button
      v-if="expandable"
      type="button"
      class="json-toggle"
      @click="open = !open"
    >
      <span class="json-caret" :class="{ open }">›</span>
      <span v-if="label" class="json-key">{{ label }}</span>
      <span class="json-preview">{{ preview }}</span>
    </button>
    <div v-else class="json-leaf">
      <span v-if="label" class="json-key">{{ label }}</span>
      <span class="json-colon" v-if="label">:</span>
      <span class="json-value" :class="primitiveClass(value)">{{ preview }}</span>
    </div>

    <div v-if="expandable && open" class="json-children">
      <RuntimeJsonTree
        v-for="entry in entries"
        :key="entry.key"
        :label="entry.key"
        :value="entry.value"
        :depth="depth + 1"
      />
    </div>
  </div>
</template>

<style scoped>
.json-node { font: 11px/1.45 ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; }
.json-toggle,
.json-leaf {
  width: 100%;
  min-height: 24px;
  padding: 2px 8px 2px calc(8px + var(--depth) * 14px);
  display: flex;
  align-items: flex-start;
  gap: 6px;
  border: 0;
  color: var(--text-secondary);
  background: transparent;
  text-align: left;
}
.json-toggle { cursor: pointer; }
.json-toggle:hover { background: var(--surface-muted); }
.json-caret {
  width: 10px;
  flex: 0 0 auto;
  color: var(--text-muted);
  transform: rotate(0deg);
  transition: transform 0.12s ease;
}
.json-caret.open { transform: rotate(90deg); }
.json-key { flex: 0 0 auto; color: var(--info); }
.json-colon { color: var(--text-muted); }
.json-preview { min-width: 0; overflow: hidden; color: var(--text-muted); text-overflow: ellipsis; white-space: nowrap; }
.json-value {
  min-width: 0;
  overflow-wrap: anywhere;
  white-space: pre-wrap;
}
.json-value.string { color: var(--success); }
.json-value.number,
.json-value.boolean { color: var(--warning); }
.json-value.nil { color: var(--text-muted); }
.json-children { border-left: 1px solid var(--divider); margin-left: calc(13px + var(--depth) * 14px); }
</style>
