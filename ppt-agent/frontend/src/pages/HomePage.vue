<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { ArrowRight, Clock3, LayoutPanelTop, LoaderCircle, Sparkles, WandSparkles } from 'lucide-vue-next';
import AppShell from '../components/AppShell.vue';
import { createTask, isLoggedIn, routeMessage } from '../api';
import { authState } from '../stores/auth';

type CreationMode = 'planned' | 'custom';

const router = useRouter();
const auth = authState;
const brief = ref('');
const mode = ref<CreationMode>('planned');
const creating = ref(false);
const createError = ref('');
const intentNotice = ref('');
const activeMode = ref<'chat' | 'pptagent'>('chat');

const canStart = computed(() => brief.value.trim().length > 0 && !creating.value);
const selectedLabel = computed(() => mode.value === 'planned' ? '智能规划' : '自定义编排');

onMounted(async () => {
  if (isLoggedIn() && !auth.user) await auth.init();
});

async function startCreation() {
  const query = brief.value.trim();
  if (!query || creating.value) return;
  createError.value = '';
  intentNotice.value = '';

  if (mode.value === 'custom') {
    await router.push({ path: '/compose', query: { brief: query, mode: 'custom' } });
    return;
  }

  creating.value = true;
  try {
    const routed = await routeMessage(query, '', activeMode.value);
    activeMode.value = routed.mode;
    if (routed.intent === 'chat' || routed.action === 'reply') {
      intentNotice.value = routed.reply || '这是普通对话，不会创建 PPT 任务。';
      return;
    }
    if (routed.intent === 'plan' || routed.action === 'save_plan') {
      await router.push({ path: '/compose', query: { brief: routed.normalized_request || query, mode: 'plan', draft: routed.draft_id || undefined } });
      return;
    }
    if (routed.intent === 'fix') {
      const candidates = (routed.task_candidates || []).map(item => `- ${item.title || item.id} (${item.id})`).join('\n');
      intentNotice.value = (routed.reply || '这是修复请求，请先在任务记录中选择要修改的 PPT。') + (candidates ? `\n\n最近可选任务：\n${candidates}` : '');
      return;
    }
    if (routed.needs_confirmation || routed.action === 'ask_clarification') {
      intentNotice.value = routed.reply || '已识别为 PPT 意图，但还需要补充信息。';
      return;
    }
    const task = await createTask(routed.normalized_request || query);
    await router.push({ path: '/dashboard', query: { select: task.id } });
  } catch (error) {
    createError.value = error instanceof Error ? error.message : '任务创建失败，请重试';
  } finally {
    creating.value = false;
  }
}
</script>

<template>
  <AppShell title="开始创作" eyebrow="PPT 工作区" content-class="home-workspace">
    <template #actions>
      <button class="ui-button" type="button" @click="router.push('/dashboard')">
        <Clock3 :size="17" />
        <span>任务记录</span>
      </button>
    </template>

    <section class="creation-zone" aria-labelledby="creation-title">
      <div class="creation-copy">
        <span class="section-kicker"><Sparkles :size="15" /> 新建演示</span>
        <h2 id="creation-title">今天要讲清楚什么？</h2>
        <p>写下受众、场景、页数和最重要的结论。Planner 会先完成叙事与组件规划，再进入审查和并发渲染。</p>
      </div>

      <div class="prompt-composer">
        <label for="presentation-brief">演示文稿需求</label>
        <textarea
          id="presentation-brief"
          v-model="brief"
          rows="6"
          placeholder="例如：为产品委员会准备一份 10 页的季度复盘，突出增长、用户反馈和下一阶段取舍，语气务实。"
          @keydown.ctrl.enter.prevent="startCreation"
          @keydown.meta.enter.prevent="startCreation"
        />

        <div class="mode-switch" aria-label="生成方式">
          <button type="button" :class="{ active: mode === 'planned' }" @click="mode = 'planned'">
            <WandSparkles :size="17" />
            <span><strong>智能规划</strong><small>动态拆页与组件编排</small></span>
          </button>
          <button type="button" :class="{ active: mode === 'custom' }" @click="mode = 'custom'">
            <LayoutPanelTop :size="17" />
            <span><strong>自定义编排</strong><small>手动确定页面结构</small></span>
          </button>
        </div>

        <div class="composer-footer">
          <span class="selected-mode">{{ selectedLabel }}</span>
          <button class="start-button" type="button" :disabled="!canStart" @click="startCreation">
            <LoaderCircle v-if="creating" :size="18" class="spin" />
            <span>{{ creating ? '创建中' : mode === 'custom' ? '打开编排' : '开始规划' }}</span>
            <ArrowRight v-if="!creating" :size="18" />
          </button>
        </div>
        <p v-if="intentNotice" class="intent-notice" role="status">{{ intentNotice }}</p>
        <p v-if="createError" class="creation-error" role="alert">{{ createError }}</p>
      </div>
    </section>

    <section class="workflow-strip" aria-label="生成流程">
      <div><span>01</span><strong>意图识别</strong><small>判断受众、场景与规模</small></div>
      <div><span>02</span><strong>DeckSpec 规划</strong><small>组织叙事与页面组件</small></div>
      <div><span>03</span><strong>质量审查</strong><small>校验密度、结构与契约</small></div>
      <div><span>04</span><strong>并发渲染</strong><small>按页生成并完成交付</small></div>
    </section>
  </AppShell>
</template>

<style scoped>
:global(.home-workspace) { width: min(100%, 1320px); margin: 0 auto; padding: 42px 48px 64px; }
.creation-zone { display: grid; grid-template-columns: minmax(250px, .72fr) minmax(460px, 1.28fr); gap: 48px; align-items: start; }
.creation-copy { padding-top: 18px; }
.section-kicker { display: inline-flex; align-items: center; gap: 7px; color: var(--action-ink); font-size: 11px; font-weight: 750; }
.creation-copy h2 { margin: 10px 0 0; color: var(--text); font-size: 34px; line-height: 1.16; letter-spacing: 0; }
.creation-copy p { max-width: 430px; margin: 14px 0 0; color: var(--text-secondary); font-size: 14px; line-height: 1.72; }
.prompt-composer { padding: 17px; border: 1px solid var(--border-strong); border-radius: 8px; background: var(--surface); box-shadow: var(--shadow-sm); }
.prompt-composer > label { display: block; margin-bottom: 8px; color: var(--text-secondary); font-size: 11px; font-weight: 700; }
.prompt-composer textarea { width: 100%; min-height: 150px; padding: 4px 2px 13px; resize: vertical; border: 0; outline: 0; color: var(--text); background: transparent; font: inherit; font-size: 16px; line-height: 1.65; }
.prompt-composer textarea::placeholder { color: var(--text-muted); }
.mode-switch { padding: 10px 0; display: grid; grid-template-columns: 1fr 1fr; gap: 8px; border-top: 1px solid var(--divider); }
.mode-switch button { min-width: 0; min-height: 58px; padding: 9px 10px; display: grid; grid-template-columns: 22px minmax(0, 1fr); align-items: center; gap: 8px; border: 1px solid var(--border); border-radius: 6px; color: var(--text-secondary); background: var(--surface-muted); text-align: left; cursor: pointer; }
.mode-switch button.active { border-color: var(--action-ink); color: var(--action-ink); background: var(--action-soft); box-shadow: 0 0 0 1px rgba(7, 94, 87, .08); }
.mode-switch span { min-width: 0; display: flex; flex-direction: column; gap: 2px; }
.mode-switch strong { color: var(--text); font-size: 12px; }
.mode-switch small { overflow: hidden; color: var(--text-muted); font-size: 10px; text-overflow: ellipsis; white-space: nowrap; }
.composer-footer { padding-top: 11px; display: flex; align-items: center; justify-content: space-between; gap: 12px; border-top: 1px solid var(--divider); }
.selected-mode { color: var(--text-secondary); font-size: 12px; }
.start-button { min-width: 128px; min-height: 44px; padding: 0 15px; display: inline-flex; align-items: center; justify-content: center; gap: 8px; border: 1px solid var(--action-ink); border-radius: 6px; color: #fff; background: var(--action-ink); font-weight: 700; cursor: pointer; }
.start-button:hover:not(:disabled) { background: #064d48; }
.start-button:disabled { border-color: var(--border-strong); color: var(--text-disabled); background: var(--surface-pressed); cursor: not-allowed; }
.creation-error { margin: 10px 0 0; color: var(--danger); font-size: 12px; }
.intent-notice { margin: 10px 0 0; padding: 10px 11px; border-left: 3px solid var(--info); color: var(--text-secondary); background: var(--info-soft); font-size: 12px; line-height: 1.6; white-space: pre-line; }
.workflow-strip { margin-top: 42px; display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); border-top: 1px solid var(--border); border-bottom: 1px solid var(--border); }
.workflow-strip div { min-width: 0; padding: 18px 20px; display: grid; grid-template-columns: 28px minmax(0, 1fr); gap: 3px 9px; border-right: 1px solid var(--divider); }
.workflow-strip div:last-child { border-right: 0; }
.workflow-strip span { grid-row: 1 / 3; color: var(--action-ink); font-size: 11px; font-weight: 800; }
.workflow-strip strong { color: var(--text); font-size: 12px; }
.workflow-strip small { overflow: hidden; color: var(--text-muted); font-size: 10px; text-overflow: ellipsis; white-space: nowrap; }
.spin { animation: spin .9s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
@media (max-width: 920px) {
  :global(.home-workspace) { padding: 28px 28px 52px; }
  .creation-zone { grid-template-columns: 1fr; gap: 18px; }
  .creation-copy { padding-top: 0; }
  .creation-copy h2 { font-size: 28px; }
  .workflow-strip { grid-template-columns: 1fr 1fr; }
  .workflow-strip div:nth-child(2) { border-right: 0; }
  .workflow-strip div:nth-child(-n+2) { border-bottom: 1px solid var(--divider); }
}
@media (max-width: 600px) {
  :global(.home-workspace) { padding: 20px 16px 40px; }
  .mode-switch, .workflow-strip { grid-template-columns: 1fr; }
  .workflow-strip div { border-right: 0; border-bottom: 1px solid var(--divider); }
  .workflow-strip div:last-child { border-bottom: 0; }
  .composer-footer { align-items: stretch; flex-direction: column; }
  .start-button { width: 100%; }
}
@media (prefers-reduced-motion: reduce) { .spin { animation: none; } }
</style>
