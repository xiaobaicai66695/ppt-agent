<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import {
  ArrowRight,
  Clock3,
  LayoutDashboard,
  LayoutTemplate,
  LoaderCircle,
  Palette,
  ShieldCheck,
  Sparkles,
  WandSparkles,
} from 'lucide-vue-next';
import AppShell from '../components/AppShell.vue';
import { createTask, fetchPresets, isLoggedIn, recommendTemplate, type PresetTemplate, type TemplateRecommendation } from '../api';
import { authState } from '../stores/auth';
import { resolveCreationSelection } from '../utils/creation';

type Choice =
  | { kind: 'recommended'; key: 'recommended' }
  | { kind: 'preset'; key: string; template: PresetTemplate }
  | { kind: 'custom'; key: 'custom' };

const router = useRouter();
const auth = authState;
const brief = ref('');
const presets = ref<PresetTemplate[]>([]);
const selectedKey = ref('recommended');
const catalogLoading = ref(false);
const catalogError = ref('');
const creating = ref(false);
const recommending = ref(false);
const createError = ref('');
const recommendation = ref<TemplateRecommendation | null>(null);
const recommendationBrief = ref('');

const choices = computed<Choice[]>(() => [
  { kind: 'recommended', key: 'recommended' },
  ...presets.value.map(template => ({ kind: 'preset' as const, key: `preset:${template.name}`, template })),
  { kind: 'custom', key: 'custom' },
]);

const selectedChoice = computed(() => choices.value.find(item => item.key === selectedKey.value) || choices.value[0]);
const selectedLabel = computed(() => {
  const choice = selectedChoice.value;
  if (choice.kind === 'recommended') return '智能推荐';
  if (choice.kind === 'custom') return '自定义编排';
  return choice.template.display_name;
});
const canStart = computed(() => brief.value.trim().length > 0 && !creating.value && !recommending.value);
const recommendationReady = computed(() => (
  selectedChoice.value.kind === 'recommended'
  && recommendation.value
  && recommendationBrief.value === brief.value.trim()
));
const startButtonLabel = computed(() => {
  if (creating.value) return '创建中';
  if (recommending.value) return '推荐中';
  if (selectedChoice.value.kind === 'custom') return '打开编排';
  if (selectedChoice.value.kind === 'recommended' && !recommendationReady.value) return '生成推荐';
  return '确认生成';
});

async function loadCatalog() {
  catalogLoading.value = true;
  catalogError.value = '';
  try {
    presets.value = await fetchPresets();
  } catch (error) {
    presets.value = [];
    catalogError.value = error instanceof Error ? error.message : '模板目录暂时无法加载';
  } finally {
    catalogLoading.value = false;
  }
}

onMounted(async () => {
  if (isLoggedIn() && !auth.user) await auth.init();
  await loadCatalog();
});

function handleThumbnailError(event: Event) {
  const image = event.target as HTMLImageElement;
  image.hidden = true;
  image.parentElement?.classList.add('missing');
}

async function startCreation() {
  const query = brief.value.trim();
  if (!query || creating.value) return;
  createError.value = '';
  const choice = resolveCreationSelection(selectedKey.value);

  if (choice.kind === 'custom') {
    await router.push({ path: '/compose', query: { brief: query, mode: 'custom' } });
    return;
  }

  if (selectedChoice.value.kind === 'recommended' && !recommendationReady.value) {
    recommending.value = true;
    try {
      recommendation.value = await recommendTemplate(query);
      recommendationBrief.value = query;
    } catch (error) {
      createError.value = error instanceof Error ? error.message : '模板推荐失败，请重试';
    } finally {
      recommending.value = false;
    }
    return;
  }

  creating.value = true;
  try {
    const task = await createTask(query, choice.templateSelection);
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
        <p>写下受众、场景、页数和最重要的结论。</p>
      </div>

      <div class="prompt-composer">
        <label for="presentation-brief">演示文稿需求</label>
        <textarea
          id="presentation-brief"
          v-model="brief"
          rows="5"
          placeholder="例如：为产品委员会准备一份 10 页的季度复盘，突出增长、用户反馈和下一阶段取舍，语气务实。"
          @keydown.ctrl.enter.prevent="startCreation"
          @keydown.meta.enter.prevent="startCreation"
        />
        <div class="composer-footer">
          <div class="selected-mode" :title="selectedLabel">
            <LayoutTemplate :size="16" />
            <span>{{ selectedLabel }}</span>
          </div>
          <button class="start-button" type="button" :disabled="!canStart" @click="startCreation">
            <LoaderCircle v-if="creating || recommending" :size="18" class="spin" />
            <span>{{ startButtonLabel }}</span>
            <ArrowRight v-if="!creating && !recommending" :size="18" />
          </button>
        </div>
        <p v-if="createError" class="creation-error" role="alert">{{ createError }}</p>
      </div>
    </section>

    <section v-if="recommendationReady && recommendation" class="recommendation-panel" aria-label="推荐策略预览">
      <div class="recommendation-summary">
        <span class="recommendation-badge"><ShieldCheck :size="16" /> 推荐已生成</span>
        <h2>{{ recommendation.primary_template.display_name || '智能推荐方案' }}</h2>
        <p>{{ recommendation.strategy.reason }}</p>
      </div>
      <div class="recommendation-grid">
        <div>
          <span>模板</span>
          <strong>{{ recommendation.strategy.template }}</strong>
          <small>{{ recommendation.primary_template.description }}</small>
        </div>
        <div>
          <span>页数</span>
          <strong>{{ recommendation.strategy.page_count || '动态' }} 页</strong>
          <small>Planner 会按章节和内容密度细化</small>
        </div>
        <div>
          <span>配色</span>
          <strong>{{ recommendation.theme?.display_name || recommendation.strategy.theme }}</strong>
          <small>{{ recommendation.visual_policy }}</small>
        </div>
        <div>
          <span>视觉</span>
          <strong>{{ recommendation.strategy.use_visual_assets ? '图片素材优先' : '信息表面优先' }}</strong>
          <small>由 Planner 按页面内容规划图片检索和图文混排</small>
        </div>
      </div>
      <div class="component-focus">
        <span v-for="item in recommendation.component_focus" :key="item">{{ item }}</span>
      </div>
      <p v-if="recommendation.risks?.length" class="recommendation-risk">{{ recommendation.risks.join('；') }}</p>
    </section>

    <section class="template-section" aria-labelledby="template-heading">
      <div class="section-heading">
        <div>
          <span class="section-kicker">本次视觉方案</span>
          <h2 id="template-heading">选择一次，直接进入生成</h2>
        </div>
        <span class="catalog-count">{{ presets.length }} 套预设</span>
      </div>

      <p v-if="catalogError" class="catalog-error" role="alert">
        <span>{{ catalogError }}，仍可使用智能推荐或自定义编排。</span>
        <button type="button" :disabled="catalogLoading" @click="loadCatalog">重新加载</button>
      </p>

      <div class="template-grid" aria-label="模板目录">
        <button
          class="template-card special"
          :class="{ selected: selectedKey === 'recommended' }"
          type="button"
          :aria-pressed="selectedKey === 'recommended'"
          @click="selectedKey = 'recommended'"
        >
          <span class="special-media recommendation-media">
            <span class="recommendation-mark"><WandSparkles :size="27" /></span>
            <span class="recommendation-stack" aria-hidden="true">
              <i></i><i></i><i></i>
            </span>
          </span>
          <span class="template-copy">
            <span><strong>智能推荐</strong><em>推荐</em></span>
            <small>按主题选择模板、配色和背景策略</small>
          </span>
        </button>

        <template v-if="catalogLoading && presets.length === 0">
          <div v-for="index in 3" :key="index" class="template-skeleton" aria-hidden="true"><span></span><i></i></div>
        </template>

        <button
          v-for="template in presets"
          :key="template.name"
          class="template-card"
          :class="{ selected: selectedKey === `preset:${template.name}` }"
          type="button"
          :aria-pressed="selectedKey === `preset:${template.name}`"
          @click="selectedKey = `preset:${template.name}`"
        >
          <span class="template-media" :class="{ missing: !template.thumbnail }">
            <img
              v-if="template.thumbnail"
              :src="template.thumbnail"
              :alt="`${template.display_name}模板预览`"
              loading="lazy"
              width="640"
              height="360"
              @error="handleThumbnailError"
            />
            <span class="missing-preview"><LayoutDashboard :size="25" />预览暂缺</span>
          </span>
          <span class="template-copy">
            <span><strong>{{ template.display_name }}</strong><em>{{ template.slide_count }} 页</em></span>
            <small>{{ template.description || template.category }}</small>
          </span>
        </button>

        <button
          class="template-card special"
          :class="{ selected: selectedKey === 'custom' }"
          type="button"
          :aria-pressed="selectedKey === 'custom'"
          @click="selectedKey = 'custom'"
        >
          <span class="special-media custom-media">
            <span class="custom-canvas" aria-hidden="true"><i></i><i></i><i></i></span>
            <span class="custom-mark"><Palette :size="24" /></span>
          </span>
          <span class="template-copy">
            <span><strong>自定义编排</strong><em>自由</em></span>
            <small>自行添加页面、布局、文字和背景</small>
          </span>
        </button>
      </div>
    </section>
  </AppShell>
</template>

<style scoped>
:global(.home-workspace) {
  width: min(100%, 1440px);
  margin: 0 auto;
  padding: 36px 48px 64px;
}

.creation-zone {
  display: grid;
  grid-template-columns: minmax(230px, 0.68fr) minmax(420px, 1.32fr);
  gap: 44px;
  align-items: start;
}
.creation-copy { padding-top: 16px; }
.section-kicker { display: inline-flex; align-items: center; gap: 7px; color: var(--action-ink); font-size: 11px; font-weight: 750; }
.creation-copy h2 { margin: 10px 0 0; color: var(--text); font-size: 34px; line-height: 1.16; letter-spacing: 0; }
.creation-copy p { margin: 13px 0 0; color: var(--text-secondary); font-size: 14px; line-height: 1.65; }

.prompt-composer { padding: 16px; border: 1px solid var(--border-strong); border-radius: 8px; background: var(--surface); box-shadow: var(--shadow-sm); }
.prompt-composer label { display: block; margin-bottom: 8px; color: var(--text-secondary); font-size: 11px; font-weight: 700; }
.prompt-composer textarea { width: 100%; min-height: 126px; padding: 4px 2px 12px; resize: vertical; border: 0; outline: 0; color: var(--text); background: transparent; font: inherit; font-size: 16px; line-height: 1.65; }
.prompt-composer textarea::placeholder { color: var(--text-muted); }
.composer-footer { padding-top: 12px; display: flex; align-items: center; justify-content: space-between; gap: 12px; border-top: 1px solid var(--divider); }
.selected-mode { min-width: 0; display: flex; align-items: center; gap: 7px; color: var(--text-secondary); font-size: 12px; }
.selected-mode span { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.start-button { min-width: 124px; min-height: 44px; padding: 0 15px; display: inline-flex; align-items: center; justify-content: center; gap: 8px; border: 1px solid var(--action-ink); border-radius: 6px; color: #fff; background: var(--action-ink); font-weight: 700; cursor: pointer; }
.start-button:hover:not(:disabled) { background: #064d48; }
.start-button:disabled { border-color: var(--border-strong); color: var(--text-disabled); background: var(--surface-pressed); cursor: not-allowed; }
.creation-error { margin: 10px 0 0; color: var(--danger); font-size: 12px; }

.recommendation-panel {
  margin-top: 22px;
  padding: 16px;
  border: 1px solid #c9ddd7;
  border-radius: 8px;
  background: #fbfdfb;
  box-shadow: var(--shadow-xs);
}
.recommendation-summary {
  display: grid;
  gap: 7px;
}
.recommendation-badge {
  width: fit-content;
  min-height: 28px;
  padding: 0 9px;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  border-radius: 5px;
  color: var(--action-ink);
  background: var(--action-soft);
  font-size: 11px;
  font-weight: 750;
}
.recommendation-summary h2 { margin: 0; color: var(--text); font-size: 18px; }
.recommendation-summary p { margin: 0; color: var(--text-secondary); font-size: 13px; line-height: 1.65; }
.recommendation-grid {
  margin-top: 14px;
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 9px;
}
.recommendation-grid > div {
  min-width: 0;
  min-height: 104px;
  padding: 11px;
  display: flex;
  flex-direction: column;
  border: 1px solid var(--border);
  border-radius: 7px;
  background: var(--surface);
}
.recommendation-grid span { color: var(--text-muted); font-size: 10px; font-weight: 750; }
.recommendation-grid strong { margin-top: 8px; overflow: hidden; color: var(--text); font-size: 14px; text-overflow: ellipsis; white-space: nowrap; }
.recommendation-grid small {
  display: -webkit-box;
  margin-top: 7px;
  overflow: hidden;
  color: var(--text-secondary);
  font-size: 11px;
  line-height: 1.45;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}
.component-focus {
  margin-top: 12px;
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}
.component-focus span {
  padding: 5px 8px;
  border-radius: 5px;
  color: #315550;
  background: #edf5f2;
  font-size: 11px;
  font-weight: 650;
}
.recommendation-risk {
  margin: 11px 0 0;
  color: var(--warning);
  font-size: 12px;
}

.template-section { margin-top: 42px; }
.section-heading { margin-bottom: 16px; display: flex; align-items: end; justify-content: space-between; gap: 18px; }
.section-heading h2 { margin: 5px 0 0; color: var(--text); font-size: 20px; letter-spacing: 0; }
.catalog-count { color: var(--text-muted); font-size: 11px; }
.catalog-error { min-height: 42px; margin: 0 0 14px; padding: 8px 10px; display: flex; align-items: center; justify-content: space-between; gap: 12px; border: 1px solid var(--border-strong); border-radius: 6px; color: var(--text-secondary); background: var(--warning-soft); font-size: 12px; }
.catalog-error button { min-height: 34px; padding: 0 9px; flex: 0 0 auto; border: 1px solid var(--border-strong); border-radius: 5px; color: var(--text); background: var(--surface); cursor: pointer; }

.template-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 14px; }
.template-card { min-width: 0; padding: 0; overflow: hidden; border: 1px solid var(--border); border-radius: 8px; color: var(--text); background: var(--surface); text-align: left; cursor: pointer; transition: border-color var(--motion-fast), box-shadow var(--motion-fast), transform var(--motion-fast); }
.template-card:hover { transform: translateY(-2px); border-color: var(--border-strong); box-shadow: var(--shadow-sm); }
.template-card.selected { border-color: var(--action-ink); box-shadow: 0 0 0 2px rgba(7, 94, 87, 0.12); }
.template-media, .special-media { position: relative; display: block; aspect-ratio: 16 / 9; overflow: hidden; border-bottom: 1px solid var(--divider); background: var(--surface-muted); }
.template-media img { width: 100%; height: 100%; display: block; object-fit: cover; }
.missing-preview { position: absolute; inset: 0; display: none; place-items: center; align-content: center; gap: 6px; color: var(--text-muted); font-size: 11px; }
.template-media.missing .missing-preview { display: grid; }
.template-copy { min-width: 0; min-height: 68px; padding: 10px 11px 11px; display: flex; flex-direction: column; }
.template-copy > span { min-width: 0; display: flex; align-items: center; justify-content: space-between; gap: 8px; }
.template-copy strong { min-width: 0; overflow: hidden; color: var(--text); font-size: 13px; text-overflow: ellipsis; white-space: nowrap; }
.template-copy em { padding: 2px 5px; flex: 0 0 auto; border-radius: 3px; color: var(--action-ink); background: var(--action-soft); font-size: 9px; font-style: normal; font-weight: 750; }
.template-copy small { margin-top: 5px; overflow: hidden; color: var(--text-muted); font-size: 10px; line-height: 1.45; text-overflow: ellipsis; white-space: nowrap; }

.recommendation-media { display: grid; grid-template-columns: 1fr 1.25fr; align-items: center; padding: 18px 22px; background: #e8f3f1; }
.recommendation-mark { width: 54px; height: 54px; display: grid; place-items: center; border-radius: 8px; color: #fff; background: var(--action-ink); }
.recommendation-stack { display: grid; gap: 7px; transform: rotate(-3deg); }
.recommendation-stack i { height: 14px; display: block; border: 1px solid #a9c9c4; border-radius: 3px; background: #fff; }
.recommendation-stack i:nth-child(2) { width: 78%; background: #f6bf74; }
.recommendation-stack i:nth-child(3) { width: 58%; background: #d9654a; }
.custom-media { display: grid; place-items: center; background: #edf0f1; }
.custom-canvas { width: 64%; height: 62%; padding: 10px; display: grid; grid-template-columns: 1fr 1fr; gap: 7px; border: 1px solid #bbc4c7; border-radius: 5px; background: #fff; box-shadow: var(--shadow-sm); }
.custom-canvas i { display: block; border-radius: 2px; background: #dce3e4; }
.custom-canvas i:first-child { grid-column: 1 / -1; background: #075e57; }
.custom-canvas i:last-child { background: #d9654a; }
.custom-mark { position: absolute; right: 18%; bottom: 14%; width: 42px; height: 42px; display: grid; place-items: center; border: 1px solid var(--border); border-radius: 6px; color: var(--action-ink); background: #fff; }
.template-skeleton { overflow: hidden; border: 1px solid var(--border); border-radius: 8px; background: var(--surface); }
.template-skeleton span { display: block; aspect-ratio: 16 / 9; background: var(--surface-pressed); animation: pulse 1.3s ease-in-out infinite; }
.template-skeleton i { width: 48%; height: 13px; margin: 14px 11px 20px; display: block; border-radius: 3px; background: var(--surface-pressed); }
.spin { animation: spin .9s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
@keyframes pulse { 50% { opacity: .55; } }

@media (max-width: 1024px) {
  :global(.home-workspace) { padding: 28px 28px 52px; }
  .creation-zone { grid-template-columns: 1fr; gap: 18px; }
  .creation-copy { padding-top: 0; }
  .creation-copy h2 { font-size: 28px; }
  .template-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .recommendation-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
}
@media (max-width: 620px) {
  :global(.home-workspace) { padding: 20px 14px 42px; }
  .creation-copy h2 { font-size: 24px; }
  .prompt-composer { padding: 13px; }
  .composer-footer { align-items: stretch; flex-direction: column; }
  .start-button { width: 100%; }
  .section-heading { align-items: flex-start; }
  .template-grid { grid-template-columns: 1fr; }
  .recommendation-grid { grid-template-columns: 1fr; }
  .template-card { display: grid; grid-template-columns: 132px minmax(0, 1fr); }
  .template-media, .special-media { height: 100%; min-height: 86px; aspect-ratio: auto; border-right: 1px solid var(--divider); border-bottom: 0; }
  .template-copy { min-height: 86px; justify-content: center; }
  .recommendation-media { padding: 14px; }
  .recommendation-mark { width: 42px; height: 42px; }
  .template-skeleton { min-height: 86px; display: grid; grid-template-columns: 132px 1fr; }
  .template-skeleton span { aspect-ratio: auto; }
  .catalog-error { align-items: flex-start; flex-direction: column; }
  .catalog-error button { min-height: 40px; }
}
@media (prefers-reduced-motion: reduce) {
  .template-card, .spin, .template-skeleton span { transition: none; animation: none; }
}
</style>
