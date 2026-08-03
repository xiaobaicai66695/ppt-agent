<script setup lang="ts">
import { computed, ref } from 'vue';
import { Check, Clock3, History, LoaderCircle, Plus, Save, Settings2, Trash2, X } from 'lucide-vue-next';
import type { TaskInfo } from '../types';
import { summarizeProfile, updateUserProfile, type PreferenceSummary } from '../api';
import { summarizeTaskTitle } from '../utils/workbench';

const props = defineProps<{
  user: { id: number; email: string; is_admin?: boolean } | null;
  tasks: TaskInfo[];
  selectedId: string | null;
}>();

const emit = defineEmits<{
  logout: [];
  selectTask: [id: string];
  deleteTask: [id: string];
  compose: [];
  newSession: [];
}>();

const taskCount = computed(() => props.tasks.length);

function fmtTime(iso: string): string {
  if (!iso) return '';
  const date = new Date(iso);
  return date.toDateString() === new Date().toDateString()
    ? date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
    : `${date.getMonth() + 1}/${date.getDate()}`;
}

function fmtTokens(value: number): string {
  if (!value || value <= 0) return '';
  if (value >= 1_000_000) return `${(value / 1_000_000).toFixed(1)}M`;
  if (value >= 1_000) return `${(value / 1000).toFixed(1)}K`;
  return String(value);
}

function statusLabel(status: string): string {
  return ({ running: '运行中', completed: '已完成', cancelled: '已中断', failed: '失败' } as Record<string, string>)[status] || status;
}

const showPrefs = ref(false);
const prefsLoading = ref(false);
const prefsSummary = ref<PreferenceSummary | null>(null);
const prefsSaved = ref(false);
const prefsError = ref('');

interface PreferenceDraft {
  preferred_themes: string;
  preferred_colors: string;
  content_patterns: string;
  layout_preferences: string;
  language_tone: string;
  typical_page_count: number;
}

const prefsEditing = ref<PreferenceDraft | null>(null);

function listToDraft(values?: string[]): string { return (values || []).join(', '); }
function draftToList(value: string): string[] { return value.split(/[,，]/).map(item => item.trim()).filter(Boolean); }

async function openPrefs() {
  showPrefs.value = true;
  prefsLoading.value = true;
  prefsSaved.value = false;
  prefsError.value = '';
  try {
    const data = await summarizeProfile();
    prefsSummary.value = data.summary;
    prefsEditing.value = {
      preferred_themes: listToDraft(data.summary.preferred_themes),
      preferred_colors: listToDraft(data.summary.preferred_colors),
      content_patterns: listToDraft(data.summary.content_patterns),
      layout_preferences: listToDraft(data.summary.layout_preferences),
      language_tone: data.summary.language_tone || '',
      typical_page_count: data.summary.typical_page_count || 0,
    };
  } catch (error) {
    prefsSummary.value = null;
    prefsEditing.value = null;
    prefsError.value = error instanceof Error ? error.message : '偏好加载失败';
  } finally {
    prefsLoading.value = false;
  }
}

async function savePrefs() {
  if (!prefsEditing.value) return;
  prefsError.value = '';
  try {
    await updateUserProfile({
      preferred_themes: draftToList(prefsEditing.value.preferred_themes),
      preferred_colors: draftToList(prefsEditing.value.preferred_colors),
      content_patterns: draftToList(prefsEditing.value.content_patterns),
      layout_preferences: draftToList(prefsEditing.value.layout_preferences),
      language_tone: prefsEditing.value.language_tone,
      typical_page_count: prefsEditing.value.typical_page_count,
    });
    prefsSaved.value = true;
    setTimeout(() => { prefsSaved.value = false; }, 2000);
  } catch (error) {
    prefsError.value = error instanceof Error ? error.message : '偏好保存失败';
  }
}

function closePrefs() {
  showPrefs.value = false;
  prefsSummary.value = null;
  prefsEditing.value = null;
}
</script>

<template>
  <aside class="task-sidebar" aria-label="任务工作区">
    <header class="task-sidebar-head">
      <span class="head-copy">
        <History :size="17" />
        <strong>任务</strong>
        <small>{{ taskCount }}</small>
      </span>
      <span class="head-actions">
        <button type="button" title="生成偏好" aria-label="生成偏好" @click="openPrefs">
          <Settings2 :size="17" />
        </button>
        <button type="button" title="新建会话" aria-label="新建会话" @click="emit('newSession')">
          <Plus :size="18" />
        </button>
      </span>
    </header>

    <div class="task-list" role="list">
      <div v-if="tasks.length === 0" class="task-empty">
        <History :size="22" />
        <strong>还没有生成记录</strong>
        <span>使用工作台底部输入框创建演示</span>
      </div>

      <article
        v-for="task in tasks"
        :key="task.id"
        class="task-item"
        :class="{ active: task.id === selectedId }"
        role="listitem"
      >
        <button
          class="task-select"
          type="button"
          :aria-current="task.id === selectedId ? 'true' : undefined"
          @click="emit('selectTask', task.id)"
        >
          <strong :title="task.query || ''">{{ summarizeTaskTitle(task.query) }}</strong>
          <span class="task-meta">
            <span class="task-state" :class="task.status">
              <i aria-hidden="true"></i>{{ statusLabel(task.status) }}
            </span>
            <span><Clock3 :size="12" />{{ fmtTime(task.created_at) }}</span>
          </span>
          <span v-if="task.total_count > 0" class="mini-progress">
            <i :style="{ width: `${Math.round((task.done_count / task.total_count) * 100)}%` }"></i>
          </span>
          <small v-if="(task.total_tokens || 0) > 0">{{ fmtTokens(task.total_tokens || 0) }} tokens</small>
        </button>
        <button
          v-if="task.status !== 'running'"
          class="task-delete"
          type="button"
          :aria-label="`删除任务：${task.query || task.id}`"
          title="删除任务"
          @click="emit('deleteTask', task.id)"
        >
          <Trash2 :size="15" />
        </button>
      </article>
    </div>

    <Teleport to="body">
      <div v-if="showPrefs" class="prefs-overlay" @click.self="closePrefs">
        <section class="prefs-modal" role="dialog" aria-modal="true" aria-labelledby="prefs-title">
          <header>
            <span>
              <small>创作设置</small>
              <h2 id="prefs-title">生成偏好</h2>
            </span>
            <button type="button" aria-label="关闭偏好设置" @click="closePrefs"><X :size="20" /></button>
          </header>

          <div v-if="prefsLoading" class="prefs-loading" role="status">
            <LoaderCircle :size="22" class="spin" />
            正在读取偏好
          </div>

          <form v-else-if="prefsEditing" class="prefs-form" @submit.prevent="savePrefs">
            <label>配色主题<input v-model="prefsEditing.preferred_themes" :placeholder="(prefsSummary?.preferred_themes || []).join(', ') || 'ocean_soft, sage_calm'" /></label>
            <label>偏好颜色<input v-model="prefsEditing.preferred_colors" :placeholder="(prefsSummary?.preferred_colors || []).join(', ') || '蓝色系, 高对比度'" /></label>
            <div class="prefs-row">
              <label>语言风格
                <select v-model="prefsEditing.language_tone">
                  <option value="">自动</option><option value="formal">正式</option><option value="semi-formal">半正式</option><option value="casual">轻松</option>
                </select>
              </label>
              <label>典型页数<input v-model.number="prefsEditing.typical_page_count" type="number" min="4" max="50" /></label>
            </div>
            <label>布局偏好<input v-model="prefsEditing.layout_preferences" placeholder="图表优先, 双栏对比" /></label>
            <label>内容模式<input v-model="prefsEditing.content_patterns" placeholder="案例驱动, 数据支撑" /></label>
            <p v-if="prefsError" class="prefs-error" role="alert">{{ prefsError }}</p>
            <footer>
              <button type="button" class="secondary" @click="closePrefs">取消</button>
              <button type="submit" class="primary" :disabled="prefsSaved">
                <Check v-if="prefsSaved" :size="16" /><Save v-else :size="16" />
                {{ prefsSaved ? '已保存' : '保存偏好' }}
              </button>
            </footer>
          </form>

          <div v-else class="prefs-loading error" role="alert">{{ prefsError || '暂无可用偏好数据' }}</div>
        </section>
      </div>
    </Teleport>
  </aside>
</template>

<style scoped>
.task-sidebar { height: 100%; min-height: 0; display: flex; flex-direction: column; background: var(--surface-muted); border-right: 1px solid var(--border); }
.task-sidebar-head { min-height: 52px; padding: 0 12px 0 16px; display: flex; align-items: center; justify-content: space-between; border-bottom: 1px solid var(--border); }
.head-copy { display: flex; align-items: center; gap: 8px; color: var(--text-secondary); }
.head-copy strong { color: var(--text); font-size: 13px; }
.head-copy small { min-width: 20px; height: 20px; padding: 0 5px; display: grid; place-items: center; border-radius: 10px; color: var(--text-muted); background: var(--surface-pressed); font-size: 10px; }
.head-actions { display: flex; gap: 2px; }
.head-actions button { width: 40px; height: 40px; display: grid; place-items: center; border: 0; border-radius: 5px; color: var(--text-secondary); background: transparent; cursor: pointer; }
.head-actions button:hover { color: var(--text); background: var(--surface-hover); }

.task-list { min-height: 0; padding: 8px; display: grid; align-content: start; gap: 4px; overflow-y: auto; }
.task-empty { min-height: 150px; padding: 22px; display: grid; place-content: center; justify-items: center; gap: 5px; color: var(--text-muted); text-align: center; }
.task-empty strong { color: var(--text-secondary); font-size: 12px; }
.task-empty span { font-size: 11px; }
.task-item { position: relative; min-width: 0; border: 1px solid transparent; border-radius: 6px; }
.task-item:hover { background: var(--surface); }
.task-item.active { border-color: var(--border-strong); background: var(--surface); box-shadow: inset 3px 0 0 var(--action-ink); }
.task-select { width: 100%; min-height: 84px; padding: 10px 34px 10px 11px; display: flex; flex-direction: column; align-items: stretch; border: 0; color: inherit; background: transparent; text-align: left; cursor: pointer; }
.task-select strong { overflow: hidden; color: var(--text); font-size: 12px; font-weight: 650; line-height: 1.45; text-overflow: ellipsis; white-space: nowrap; }
.task-meta { margin-top: 7px; display: flex; align-items: center; justify-content: space-between; color: var(--text-muted); font-size: 10px; }
.task-meta > span { display: inline-flex; align-items: center; gap: 4px; }
.task-state i { width: 6px; height: 6px; border-radius: 50%; background: var(--text-disabled); }
.task-state.running { color: var(--info); }.task-state.running i { background: var(--info); animation: pulse 1.4s ease-in-out infinite; }
.task-state.completed { color: var(--success); }.task-state.completed i { background: var(--success); }
.task-state.failed { color: var(--danger); }.task-state.failed i { background: var(--danger); }
.mini-progress { height: 3px; margin-top: 9px; overflow: hidden; border-radius: 2px; background: var(--surface-pressed); }
.mini-progress i { display: block; height: 100%; background: var(--action-ink); }
.task-select > small { margin-top: 5px; color: var(--text-muted); font-size: 9px; }
.task-delete { position: absolute; top: 6px; right: 4px; width: 32px; height: 32px; display: grid; place-items: center; border: 0; border-radius: 4px; color: var(--text-muted); background: transparent; opacity: 0; cursor: pointer; }
.task-item:hover .task-delete, .task-delete:focus-visible { opacity: 1; }
.task-delete:hover { color: var(--danger); background: var(--danger-soft); }

.prefs-overlay { position: fixed; inset: 0; z-index: var(--z-modal); padding: 20px; display: grid; place-items: center; background: rgba(15, 17, 18, 0.55); }
.prefs-modal { width: min(560px, 100%); max-height: min(720px, calc(100dvh - 40px)); overflow: auto; border-radius: 8px; background: var(--surface); box-shadow: var(--shadow-lg); }
.prefs-modal > header { min-height: 64px; padding: 12px 16px 12px 20px; display: flex; align-items: center; justify-content: space-between; border-bottom: 1px solid var(--border); }
.prefs-modal header span { display: flex; flex-direction: column; }.prefs-modal header small { color: var(--text-muted); font-size: 10px; font-weight: 700; }.prefs-modal h2 { margin: 2px 0 0; font-size: 16px; }
.prefs-modal header button { width: 44px; height: 44px; display: grid; place-items: center; border: 0; border-radius: 5px; color: var(--text-secondary); background: transparent; cursor: pointer; }
.prefs-modal header button:hover { background: var(--surface-muted); }
.prefs-loading { min-height: 220px; display: flex; align-items: center; justify-content: center; gap: 8px; color: var(--text-muted); }.prefs-loading.error { color: var(--danger); }
.prefs-form { padding: 20px; display: grid; gap: 14px; }
.prefs-form label { display: grid; gap: 6px; color: var(--text-secondary); font-size: 11px; font-weight: 700; }
.prefs-form input, .prefs-form select { min-width: 0; width: 100%; min-height: 42px; padding: 0 10px; border: 1px solid var(--border-strong); border-radius: 5px; color: var(--text); background: var(--surface); }
.prefs-row { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
.prefs-error { margin: 0; color: var(--danger); font-size: 11px; }
.prefs-form footer { padding-top: 14px; display: flex; justify-content: flex-end; gap: 8px; border-top: 1px solid var(--divider); }
.prefs-form footer button { min-height: 40px; padding: 0 14px; display: inline-flex; align-items: center; gap: 7px; border: 1px solid var(--border-strong); border-radius: 5px; background: var(--surface); font-weight: 700; cursor: pointer; }
.prefs-form footer .primary { border-color: var(--action-ink); color: #fff; background: var(--action-ink); }

.spin { animation: spin 0.9s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
@keyframes pulse { 50% { opacity: 0.4; } }

@media (max-width: 520px) {
  .prefs-overlay { padding: 0; align-items: end; }
  .prefs-modal { width: 100%; max-height: 92dvh; border-radius: 8px 8px 0 0; }
  .prefs-row { grid-template-columns: 1fr; }
}
</style>
