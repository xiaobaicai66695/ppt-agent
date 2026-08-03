<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import {
  ArrowRight,
  Clock3,
  FilePlus2,
  LayoutTemplate,
  MessageSquareText,
  Presentation,
  Sparkles,
} from 'lucide-vue-next';
import AppShell from '../components/AppShell.vue';
import { authState } from '../stores/auth';
import { isLoggedIn } from '../api';

const router = useRouter();
const auth = authState;
const brief = ref('');
const selectedTemplate = ref('generic');

const templates = [
  { id: 'generic', name: '通用演示', meta: '清晰、克制、适合大多数主题', image: '/templates/thumbs/generic.jpg' },
  { id: 'pitch-deck', name: '商业路演', meta: '问题、方案、市场与增长', image: '/templates/thumbs/pitch-deck.jpg' },
  { id: 'tech-sharing', name: '技术分享', meta: '架构、流程、代码与结论', image: '/templates/thumbs/tech-sharing.jpg' },
  { id: 'research-report', name: '研究报告', meta: '数据、发现、分析与建议', image: '/templates/thumbs/research-report.jpg' },
  { id: 'weekly-report', name: '周报复盘', meta: '进展、指标、问题与计划', image: '/templates/thumbs/weekly-report.jpg' },
  { id: 'course-module', name: '课程模块', meta: '目标、知识点、练习与总结', image: '/templates/thumbs/course-module.jpg' },
];

const canStart = computed(() => brief.value.trim().length > 0);

onMounted(async () => {
  if (isLoggedIn() && !auth.user) await auth.init();
});

function startCompose() {
  router.push({
    path: '/compose',
    query: {
      template: selectedTemplate.value,
      ...(brief.value.trim() ? { brief: brief.value.trim() } : {}),
    },
  });
}

function useTemplate(id: string) {
  selectedTemplate.value = id;
  router.push({ path: '/compose', query: { template: id } });
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
        <span class="section-kicker"><Sparkles :size="15" /> AI 演示文稿</span>
        <h2 id="creation-title">今天要讲清楚什么？</h2>
        <p>描述受众、场景和希望传达的结论。你可以先生成结构，再逐页调整。</p>
      </div>

      <div class="prompt-composer">
        <label for="presentation-brief">演示文稿需求</label>
        <textarea
          id="presentation-brief"
          v-model="brief"
          rows="5"
          placeholder="例如：为产品委员会准备一份 10 页的季度复盘，突出增长、用户反馈和下一阶段取舍，语气务实。"
          @keydown.ctrl.enter.prevent="startCompose"
          @keydown.meta.enter.prevent="startCompose"
        />
        <div class="composer-footer">
          <div class="selected-mode">
            <LayoutTemplate :size="16" />
            <span>{{ templates.find(item => item.id === selectedTemplate)?.name }}</span>
          </div>
          <button class="start-button" type="button" :disabled="!canStart" @click="startCompose">
            <span>规划演示</span>
            <ArrowRight :size="18" />
          </button>
        </div>
      </div>

      <div class="workflow-strip" aria-label="生成流程">
        <div>
          <span class="workflow-index">01</span>
          <MessageSquareText :size="18" />
          <p><strong>描述目标</strong><small>受众、场景、页数与结论</small></p>
        </div>
        <div>
          <span class="workflow-index">02</span>
          <FilePlus2 :size="18" />
          <p><strong>确认结构</strong><small>选择模板并调整页面大纲</small></p>
        </div>
        <div>
          <span class="workflow-index">03</span>
          <Presentation :size="18" />
          <p><strong>渐进交付</strong><small>生成一页，立即预览一页</small></p>
        </div>
      </div>
    </section>

    <section class="template-section" aria-labelledby="template-heading">
      <div class="section-heading">
        <div>
          <span class="section-kicker">模板起点</span>
          <h2 id="template-heading">从熟悉的叙事结构开始</h2>
        </div>
        <button class="text-action" type="button" @click="router.push('/compose')">
          查看全部
          <ArrowRight :size="16" />
        </button>
      </div>

      <div class="template-grid">
        <button
          v-for="item in templates"
          :key="item.id"
          class="template-card"
          :class="{ selected: selectedTemplate === item.id }"
          type="button"
          :aria-pressed="selectedTemplate === item.id"
          @click="selectedTemplate = item.id"
          @dblclick="useTemplate(item.id)"
        >
          <span class="template-media">
            <img :src="item.image" :alt="`${item.name}模板预览`" loading="lazy" width="640" height="360" />
            <span class="template-use">选择模板</span>
          </span>
          <span class="template-copy">
            <strong>{{ item.name }}</strong>
            <small>{{ item.meta }}</small>
          </span>
        </button>
      </div>
    </section>

    <section class="resume-band" aria-label="继续工作">
      <div>
        <span class="resume-icon"><Clock3 :size="19" /></span>
        <p>
          <strong>{{ auth.loggedIn ? '继续之前的生成任务' : '登录后保存任务与偏好' }}</strong>
          <small>{{ auth.loggedIn ? '查看生成进度、下载页面或继续修改。' : '跨会话查看生成进度，并让模板选择更贴近你的习惯。' }}</small>
        </p>
      </div>
      <button class="ui-button" type="button" @click="router.push(auth.loggedIn ? '/dashboard' : '/auth')">
        {{ auth.loggedIn ? '打开任务工作台' : '登录工作区' }}
        <ArrowRight :size="17" />
      </button>
    </section>
  </AppShell>
</template>

<style scoped>
:global(.home-workspace) {
  width: min(100%, 1440px);
  margin: 0 auto;
  padding: 42px clamp(20px, 4vw, 64px) 64px;
}

.creation-zone {
  display: grid;
  grid-template-columns: minmax(240px, 0.7fr) minmax(420px, 1.3fr);
  gap: 28px 56px;
  align-items: start;
}

.creation-copy { padding-top: 14px; }
.section-kicker {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  color: var(--action-ink);
  font-size: 11px;
  font-weight: 750;
}

.creation-copy h2,
.section-heading h2 {
  margin: 10px 0 0;
  color: var(--text);
  font-size: clamp(26px, 3vw, 38px);
  font-weight: 720;
  line-height: 1.15;
  letter-spacing: 0;
}

.creation-copy p {
  max-width: 440px;
  margin: 16px 0 0;
  color: var(--text-secondary);
  font-size: 15px;
  line-height: 1.7;
}

.prompt-composer {
  padding: 16px;
  border: 1px solid var(--border-strong);
  border-radius: 8px;
  background: var(--surface);
  box-shadow: var(--shadow-sm);
}

.prompt-composer label {
  display: block;
  margin: 0 0 9px;
  color: var(--text-secondary);
  font-size: 11px;
  font-weight: 700;
}

.prompt-composer textarea {
  width: 100%;
  min-height: 132px;
  padding: 4px 2px 12px;
  resize: vertical;
  border: 0;
  outline: 0;
  color: var(--text);
  background: transparent;
  font-size: 16px;
  line-height: 1.65;
}

.prompt-composer textarea::placeholder { color: #939b9f; }

.composer-footer {
  padding-top: 12px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  border-top: 1px solid var(--divider);
}

.selected-mode {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 7px;
  color: var(--text-secondary);
  font-size: 12px;
}

.selected-mode span { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.start-button {
  min-height: 42px;
  padding: 0 16px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 9px;
  border: 1px solid var(--action-ink);
  border-radius: 6px;
  color: #ffffff;
  background: var(--action-ink);
  font-weight: 700;
  cursor: pointer;
  transition: background var(--motion-fast), transform var(--motion-fast);
}
.start-button:hover:not(:disabled) { background: #064d48; }
.start-button:active:not(:disabled) { transform: scale(0.98); }

.workflow-strip {
  grid-column: 1 / -1;
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  border-top: 1px solid var(--border);
  border-bottom: 1px solid var(--border);
}

.workflow-strip > div {
  min-width: 0;
  padding: 18px 22px;
  display: grid;
  grid-template-columns: auto auto 1fr;
  align-items: center;
  gap: 10px;
  color: var(--text-secondary);
}
.workflow-strip > div + div { border-left: 1px solid var(--border); }
.workflow-index { color: var(--text-muted); font-size: 10px; font-variant-numeric: tabular-nums; }
.workflow-strip svg { color: var(--action-ink); }
.workflow-strip p { min-width: 0; margin: 0; display: flex; flex-direction: column; }
.workflow-strip strong { color: var(--text); font-size: 12px; }
.workflow-strip small { margin-top: 2px; color: var(--text-muted); font-size: 10px; }

.template-section { margin-top: 48px; }
.section-heading {
  margin-bottom: 18px;
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 20px;
}
.section-heading h2 { margin-top: 6px; font-size: 21px; }

.text-action {
  min-height: 40px;
  padding: 0 4px;
  display: inline-flex;
  align-items: center;
  gap: 7px;
  border: 0;
  color: var(--action-ink);
  background: transparent;
  font-weight: 700;
  cursor: pointer;
}

.template-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16px;
}

.template-card {
  min-width: 0;
  padding: 0;
  overflow: hidden;
  border: 1px solid var(--border);
  border-radius: 8px;
  color: var(--text);
  background: var(--surface);
  text-align: left;
  cursor: pointer;
  transition: border-color var(--motion-fast), box-shadow var(--motion-fast), transform var(--motion-fast);
}
.template-card:hover { transform: translateY(-2px); border-color: var(--border-strong); box-shadow: var(--shadow-sm); }
.template-card.selected { border-color: var(--action-ink); box-shadow: 0 0 0 2px rgba(7,94,87,0.12); }

.template-media {
  position: relative;
  display: block;
  aspect-ratio: 16 / 9;
  overflow: hidden;
  background: var(--surface-muted);
  border-bottom: 1px solid var(--divider);
}
.template-media img { width: 100%; height: 100%; object-fit: cover; transition: transform var(--motion-medium); }
.template-card:hover img { transform: scale(1.015); }
.template-use {
  position: absolute;
  right: 9px;
  bottom: 9px;
  padding: 5px 8px;
  border-radius: 4px;
  color: #ffffff;
  background: rgba(23,26,28,0.86);
  font-size: 10px;
  font-weight: 700;
  opacity: 0;
  transition: opacity var(--motion-fast);
}
.template-card:hover .template-use,
.template-card:focus-visible .template-use,
.template-card.selected .template-use { opacity: 1; }

.template-copy { padding: 12px 13px 13px; display: flex; flex-direction: column; }
.template-copy strong { font-size: 13px; }
.template-copy small { margin-top: 4px; color: var(--text-muted); font-size: 11px; }

.resume-band {
  margin-top: 42px;
  padding: 20px 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  border-top: 1px solid var(--border);
  border-bottom: 1px solid var(--border);
}
.resume-band > div { display: flex; align-items: center; gap: 12px; }
.resume-icon { width: 38px; height: 38px; display: grid; place-items: center; border-radius: 6px; color: var(--info); background: var(--info-soft); }
.resume-band p { margin: 0; display: flex; flex-direction: column; }
.resume-band strong { font-size: 13px; }
.resume-band small { margin-top: 3px; color: var(--text-muted); font-size: 11px; }

@media (max-width: 900px) {
  .creation-zone { grid-template-columns: 1fr; gap: 22px; }
  .creation-copy { padding-top: 0; }
  .workflow-strip { grid-template-columns: 1fr; }
  .workflow-strip > div + div { border-left: 0; border-top: 1px solid var(--border); }
  .template-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
}

@media (max-width: 600px) {
  :global(.home-workspace) { padding: 24px 14px 40px; }
  :global(.home-workspace + *) { max-width: 100%; }
  .creation-copy h2 { font-size: 28px; }
  .creation-copy p { font-size: 14px; }
  .prompt-composer { padding: 13px; }
  .composer-footer { align-items: stretch; flex-direction: column; }
  .start-button { width: 100%; }
  .template-grid { grid-template-columns: 1fr; }
  .section-heading { align-items: flex-start; }
  .resume-band { align-items: stretch; flex-direction: column; }
  .resume-band .ui-button { width: 100%; }
}
</style>
