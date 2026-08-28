<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { ArrowDown, ArrowUp, Copy, GripVertical, LayoutTemplate, ListPlus, PanelsTopLeft, Play, Plus, Sparkles, Trash2, X } from 'lucide-vue-next';
import AppShell from '../components/AppShell.vue';
import { createTaskWithOutline, fetchLayouts } from '../api';
import type { AtomicLayout, TaskOutline, SlideOutline } from '../api';

const route = useRoute();
const router = useRouter();
// State
const layouts = ref<AtomicLayout[]>([]);
const loading = ref(false);
const loadError = ref('');
const generating = ref(false);
const generationError = ref('');
const intentMode = computed(() => route.query.mode === 'plan' ? 'plan' : 'create');
const planDraftId = computed(() => typeof route.query.draft === 'string' ? route.query.draft : '');

// Editing state
const pptTitle = ref('');
const topicInput = ref('');
const topicTrimmed = computed(() => topicInput.value.trim());
const slides = ref<SlideOutline[]>([]);
const selectedSlideIndex = ref(-1);
const editingSlide = ref<SlideOutline | null>(null);

// Drag-and-drop state
const draggingIndex = ref(-1);
const dragOverIndex = ref(-1);

function onDragStart(index: number) {
  draggingIndex.value = index;
}

function onDragOver(index: number) {
  dragOverIndex.value = index;
}

function onDragLeave(e: DragEvent) {
  const target = e.currentTarget as HTMLElement;
  const related = e.relatedTarget as HTMLElement | null;
  if (related && target.contains(related)) return;
  dragOverIndex.value = -1;
}

function onDrop(e: DragEvent, index: number) {
  e.preventDefault();
  if (draggingIndex.value >= 0 && draggingIndex.value !== index) {
    const temp = slides.value[draggingIndex.value];
    slides.value.splice(draggingIndex.value, 1);
    slides.value.splice(index, 0, temp);
    if (selectedSlideIndex.value === draggingIndex.value) {
      selectedSlideIndex.value = index;
    } else if (selectedSlideIndex.value > draggingIndex.value && selectedSlideIndex.value <= index) {
      selectedSlideIndex.value--;
    } else if (selectedSlideIndex.value < draggingIndex.value && selectedSlideIndex.value >= index) {
      selectedSlideIndex.value++;
    }
  }
  draggingIndex.value = -1;
  dragOverIndex.value = -1;
}

// Layout category groups
const layoutCategories = [
  { label: '封面/目录', names: ['title_slide', 'agenda', 'section_divider'] },
  { label: '内容', names: ['content_slide', 'quote_slide', 'deep_dive'] },
  { label: '布局', names: ['two_column', 'three_column', 'card_grid'] },
  { label: '图表', names: ['process_flow', 'stat_slide', 'kpi_dashboard'] },
  { label: '案例', names: ['case_study', 'image_text', 'summary_slide'] },
];

const groupedLayouts = computed(() => {
  const groups = layoutCategories.map(group => ({
    label: group.label,
    items: layouts.value.filter(l => group.names.includes(l.name)),
  })).filter(g => g.items.length > 0);
  const groupedNames = new Set(layoutCategories.flatMap(group => group.names));
  const remaining = layouts.value.filter(layout => !groupedNames.has(layout.name));
  if (remaining.length) groups.push({ label: '更多布局', items: remaining });
  return groups;
});

const selectedLayout = computed(() => layouts.value.find(
  layout => layout.name === editingSlide.value?.content_type,
) || null);

const capacityLabels: Record<string, string> = {
  density: '信息密度',
  max_analysis_chars: '分析文字上限',
  max_attribution_chars: '出处字数上限',
  max_body_chars: '正文单项上限',
  max_chars_per_item: '单项字数上限',
  max_datasets: '数据系列上限',
  max_header_chars: '小标题字数上限',
  max_intro_chars: '引导文字上限',
  max_items: '内容项上限',
  max_items_per_column: '每栏内容项上限',
  max_label_chars: '标签字数上限',
  max_labels: '分类标签上限',
  max_paragraph_chars: '段落字数上限',
  max_quote_chars: '引用字数上限',
  max_subtitle_chars: '副标题字数上限',
  max_title_chars: '标题字数上限',
  min_items: '内容项下限',
  min_paragraph_chars: '段落字数下限',
};

const densityLabels: Record<string, string> = { low: '低', normal: '中', high: '高' };

const capacityGuidance = computed(() => Object.entries(selectedLayout.value?.contract?.capacity || {}).map(
  ([key, value]) => ({
    key,
    label: capacityLabels[key] || key,
    value: key === 'density' ? (densityLabels[String(value)] || String(value)) : String(value),
  }),
));

const requiredFields = computed(() => {
  const layout = selectedLayout.value;
  if (!layout) return [];
  const contractFields = new Set(layout.contract?.required_fields || []);
  return layout.fields.filter(field => field.required || contractFields.has(field.name));
});

function getLayoutDisplayName(name: string): string {
  return layouts.value.find(layout => layout.name === name)?.display_name || name;
}

// Load data
async function loadWorkspaceData() {
  loading.value = true;
  loadError.value = '';
  try {
    const l = await fetchLayouts();
    layouts.value = l;

    const initialBrief = typeof route.query.brief === 'string' ? route.query.brief.trim() : '';
    if (initialBrief) {
      topicInput.value = initialBrief;
      pptTitle.value = initialBrief.length > 36 ? initialBrief.slice(0, 36) : initialBrief;
    }
    if (slides.value.length === 0) {
      slides.value = [
        { title: '', content_type: 'title_slide' },
        { title: '', content_type: 'content_slide' },
        { title: '', content_type: 'summary_slide' },
      ];
    }
  } catch (e) {
    console.error('Failed to load templates:', e);
    loadError.value = (e as Error).message || '模板资源加载失败';
  } finally {
    loading.value = false;
  }
}

onMounted(loadWorkspaceData);

function buildOutline(): TaskOutline {
  return {
    title: pptTitle.value || topicTrimmed.value || '未命名PPT',
    content_mode: 'user_outline',
    slides: slides.value.map(s => ({
      title: s.title,
      content_type: s.content_type,
      content_plan: s.content_plan,
    })),
  };
}

function addSlideFromLayout(layout: AtomicLayout) {
  const slide: SlideOutline = {
    title: layout.display_name,
    content_type: layout.name,
  };
  let insertIndex = slides.value.length;
  if (selectedSlideIndex.value >= 0) {
    insertIndex = selectedSlideIndex.value + 1;
    slides.value.splice(insertIndex, 0, slide);
  } else {
    slides.value.push(slide);
  }
  selectedSlideIndex.value = insertIndex;
  editingSlide.value = { ...slide };
}

function openSlideEditor(index: number) {
  selectedSlideIndex.value = index;
  editingSlide.value = { ...slides.value[index] };
}

function saveSlideEdit() {
  if (editingSlide.value && selectedSlideIndex.value >= 0) {
    slides.value[selectedSlideIndex.value] = { ...editingSlide.value };
    editingSlide.value = null;
    selectedSlideIndex.value = -1;
  }
}

function cancelEdit() {
  editingSlide.value = null;
  selectedSlideIndex.value = -1;
}

function deleteSlide(index: number) {
  slides.value.splice(index, 1);
  if (selectedSlideIndex.value === index) {
    selectedSlideIndex.value = -1;
    editingSlide.value = null;
  } else if (selectedSlideIndex.value > index) {
    selectedSlideIndex.value--;
  }
}

function duplicateSlide(index: number) {
  const copy = { ...slides.value[index] };
  slides.value.splice(index + 1, 0, copy);
}

function moveSlide(index: number, direction: 'up' | 'down') {
  const target = direction === 'up' ? index - 1 : index + 1;
  if (target < 0 || target >= slides.value.length) return;
  const temp = slides.value[index];
  slides.value[index] = slides.value[target];
  slides.value[target] = temp;
  if (selectedSlideIndex.value === index) selectedSlideIndex.value = target;
  else if (selectedSlideIndex.value === target) selectedSlideIndex.value = index;
}

function addBlankSlide() {
  const slide: SlideOutline = {
    title: '新页面',
    content_type: 'content_slide',
  };
  slides.value.push(slide);
}

async function startGeneration() {
  if (slides.value.length === 0) {
    alert('请先添加或选择至少一页幻灯片');
    return;
  }
  if (!topicTrimmed.value) {
    alert('请先填写 PPT 内容主题');
    return;
  }
  if (generating.value) return;

  generating.value = true;
  generationError.value = '';
  try {
    const query = topicTrimmed.value;
    const task = await createTaskWithOutline(query, buildOutline());
    await router.push({ name: 'dashboard', query: { select: task.id } });
  } catch (error) {
    generationError.value = error instanceof Error ? error.message : '任务创建失败，请重试';
  } finally {
    generating.value = false;
  }
}
</script>

<template>
  <AppShell :title="pptTitle || '高级编排'" eyebrow="Plan studio" content-class="compose-workspace">
    <template #actions>
      <span class="page-count">{{ slides.length }} 页</span>
      <button class="toolbar-action primary" type="button" :disabled="slides.length === 0 || generating || !topicTrimmed" @click="startGeneration">
        <Play :size="17" fill="currentColor" />
        <span>{{ generating ? '创建中' : '开始生成' }}</span>
      </button>
    </template>

    <section class="brief-strip" aria-label="演示文稿设置">
      <div>
        <label for="ppt-title">演示标题</label>
        <input id="ppt-title" v-model="pptTitle" placeholder="未命名演示" />
      </div>
      <div>
        <label for="topic-input">生成目标</label>
        <textarea id="topic-input" v-model="topicInput" rows="2" :disabled="slides.length === 0" placeholder="说明受众、场景、重点结论和期望页数。" />
      </div>
    </section>

    <p v-if="intentMode === 'plan'" class="workspace-notice" role="status">
      已识别为规划请求：当前只整理标题、页面结构和 DeckSpec 草稿；点击“开始生成”前不会创建或渲染 PPT 文件。
      <span v-if="planDraftId">服务端草稿 ID：{{ planDraftId }}</span>
    </p>

    <p v-if="loadError" class="workspace-error" role="alert">
      <span>{{ loadError }}</span>
      <button type="button" @click="loadWorkspaceData">重新加载</button>
    </p>
    <p v-if="generationError" class="workspace-error" role="alert">
      <span>{{ generationError }}</span>
      <button type="button" @click="startGeneration">重试</button>
    </p>

    <div class="editor-workspace">
      <aside class="resource-panel" aria-label="布局资源">
        <div class="resource-tabs" aria-label="资源类型">
          <span><PanelsTopLeft :size="17" />布局</span>
        </div>

        <div class="resource-scroll">
          <section>
            <div v-for="group in groupedLayouts" :key="group.label" class="layout-group">
              <h3>{{ group.label }}</h3>
              <div class="layout-list">
                <button v-for="layout in group.items" :key="layout.name" type="button" @click="addSlideFromLayout(layout)">
                  <span class="layout-icon"><PanelsTopLeft :size="17" /></span>
                  <span><strong>{{ layout.display_name }}</strong><small>{{ layout.description || '添加到页面列表' }}</small></span>
                  <Plus :size="16" />
                </button>
              </div>
            </div>
          </section>
        </div>
      </aside>

      <section class="outline-panel" aria-labelledby="outline-heading">
        <header class="panel-header">
          <div><span class="panel-kicker">高级模式</span><h2 id="outline-heading">页面结构</h2></div>
          <button class="add-slide-button" type="button" @click="addBlankSlide"><ListPlus :size="17" /><span>添加页面</span></button>
        </header>

        <div class="slide-list">
          <article v-for="(slide, index) in slides" :key="`${slide.title}-${index}`" class="slide-card" :class="{ selected: selectedSlideIndex === index, 'drag-over': dragOverIndex === index, 'is-dragging': draggingIndex === index }" draggable="true" @dragstart="onDragStart(index)" @dragover.prevent="onDragOver(index)" @drop.prevent="onDrop($event, index)" @dragend="draggingIndex = -1" @dragleave="onDragLeave">
            <span class="drag-handle" title="拖动排序" aria-hidden="true"><GripVertical :size="18" /></span>
            <button class="slide-open-button" type="button" @click="openSlideEditor(index)">
              <span class="slide-miniature" aria-hidden="true"><span>{{ index + 1 }}</span><LayoutTemplate :size="20" /></span>
              <span class="slide-copy">
                <strong>{{ slide.title || '未命名页面' }}</strong>
                <small>{{ getLayoutDisplayName(slide.content_type) }}</small>
              </span>
            </button>
            <div class="slide-actions">
              <button type="button" title="上移" aria-label="上移页面" @click="moveSlide(index, 'up')"><ArrowUp :size="16" /></button>
              <button type="button" title="下移" aria-label="下移页面" @click="moveSlide(index, 'down')"><ArrowDown :size="16" /></button>
              <button type="button" title="复制" aria-label="复制页面" @click="duplicateSlide(index)"><Copy :size="16" /></button>
              <button class="danger" type="button" title="删除" aria-label="删除页面" @click="deleteSlide(index)"><Trash2 :size="16" /></button>
            </div>
          </article>

          <div v-if="slides.length === 0 && !loading" class="empty-outline">
            <span><Sparkles :size="22" /></span>
            <h3>添加第一张页面</h3>
            <p>从左侧选择布局，或添加一张空白页面。</p>
            <button type="button" @click="addBlankSlide">添加页面</button>
          </div>
        </div>
      </section>

      <button v-if="editingSlide" class="editor-scrim" type="button" aria-label="关闭页面属性" @click="cancelEdit" />

      <aside class="property-panel" :class="{ open: editingSlide }" aria-label="页面属性">
        <template v-if="editingSlide">
          <header class="panel-header">
            <div><span class="panel-kicker">第 {{ selectedSlideIndex + 1 }} 页</span><h2>页面属性</h2></div>
            <button class="close-property" type="button" title="关闭" aria-label="关闭页面属性" @click="cancelEdit"><X :size="19" /></button>
          </header>

          <div class="property-body">
            <div class="field-group"><label for="field-title">页面标题</label><input id="field-title" v-model="editingSlide.title" placeholder="输入页面标题" /></div>
            <div class="field-group">
              <label for="field-type">布局类型</label>
              <select id="field-type" v-model="editingSlide.content_type">
                <option v-for="layout in layouts" :key="layout.name" :value="layout.name">{{ layout.display_name }}</option>
              </select>
            </div>

            <section v-if="selectedLayout" class="layout-guidance" aria-label="布局容量说明">
              <strong>{{ selectedLayout.display_name }}</strong>
              <p>{{ selectedLayout.description }}</p>
              <div v-if="requiredFields.length"><span>必填</span><p><small v-for="field in requiredFields" :key="field.name">{{ field.label }}</small></p></div>
              <div v-if="capacityGuidance.length"><span>容量</span><p><small v-for="item in capacityGuidance" :key="item.key">{{ item.label }} {{ item.value }}</small></p></div>
            </section>

          </div>

          <footer class="property-footer">
            <button type="button" @click="cancelEdit">取消</button>
            <button class="save-property" type="button" @click="saveSlideEdit">保存</button>
          </footer>
        </template>

        <div v-else class="property-empty">
          <span><PanelsTopLeft :size="22" /></span><h3>选择一张页面</h3><p>在这里编辑布局和标题。</p>
        </div>
      </aside>
    </div>
  </AppShell>
</template>

<style scoped>
:global(.compose-workspace) {
  height: calc(100dvh - var(--topbar-height));
  padding: 14px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  overflow: hidden;
}
.page-count, .outline-warning { color: var(--text-muted); font-size: 11px; font-variant-numeric: tabular-nums; white-space: nowrap; }
.outline-warning { color: var(--warning); }
.toolbar-action {
  min-height: 38px; padding: 0 12px; display: inline-flex; align-items: center; justify-content: center; gap: 7px;
  border: 1px solid var(--border-strong); border-radius: 6px; color: var(--text-secondary); background: var(--surface);
  font-size: 12px; font-weight: 700; cursor: pointer;
  transition: border-color var(--motion-fast), background var(--motion-fast), transform var(--motion-fast);
}
.toolbar-action:hover:not(:disabled) { border-color: #aeb7ba; background: var(--surface-muted); }
.toolbar-action:active:not(:disabled) { transform: scale(0.98); }
.toolbar-action.primary { border-color: var(--action-ink); color: #fff; background: var(--action-ink); }
.toolbar-action.primary:hover:not(:disabled) { background: #064d48; }

.brief-strip {
  min-height: 84px; padding: 12px 14px; display: grid;
  grid-template-columns: minmax(180px,.7fr) minmax(320px,1.6fr);
  gap: 14px; border: 1px solid var(--border); border-radius: 8px; background: var(--surface);
}
.brief-strip > div { min-width: 0; display: flex; flex-direction: column; gap: 6px; }
.brief-strip label, .field-group label { color: var(--text-muted); font-size: 10px; font-weight: 750; }
.brief-strip input, .brief-strip textarea { width: 100%; min-width: 0; border: 0; outline: 0; color: var(--text); background: transparent; }
.brief-strip input { height: 34px; font-size: 14px; font-weight: 700; }
.brief-strip textarea { min-height: 42px; resize: none; font-size: 13px; line-height: 1.5; }

.workspace-error { margin: 0; padding: 10px 12px; display: flex; align-items: center; justify-content: space-between; gap: 12px; border-left: 3px solid var(--danger); color: var(--danger); background: var(--danger-soft); font-size: 12px; }
.workspace-error button { min-height: 36px; border: 0; color: var(--danger); background: transparent; font-weight: 700; cursor: pointer; }
.workspace-notice { margin: 0; padding: 10px 12px; border-left: 3px solid var(--info); color: var(--text-secondary); background: var(--info-soft); font-size: 12px; line-height: 1.55; }
.workspace-notice span { display: block; margin-top: 4px; font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; font-size: 11px; }

.editor-workspace {
  min-height: 0; flex: 1; display: grid; grid-template-columns: 276px minmax(420px,1fr) 334px;
  overflow: hidden; border: 1px solid var(--border); border-radius: 8px; background: var(--surface);
}
.resource-panel, .outline-panel, .property-panel { min-width: 0; min-height: 0; display: flex; flex-direction: column; }
.resource-panel { border-right: 1px solid var(--border); background: var(--surface-muted); }
.property-panel { border-left: 1px solid var(--border); background: var(--surface); }

.resource-tabs { padding: 10px; border-bottom: 1px solid var(--border); }
.resource-tabs span {
  min-height: 40px; display: inline-flex; align-items: center; gap: 7px;
  color: var(--action-ink); font-size: 11px; font-weight: 750;
}
.resource-scroll { min-height: 0; flex: 1; overflow: auto; padding: 10px; }

.layout-list { display: grid; gap: 9px; }

.layout-group + .layout-group { margin-top: 16px; }
.layout-group h3 { margin: 0 0 7px; color: var(--text-muted); font-size: 10px; font-weight: 750; }
.layout-list button {
  min-height: 54px; padding: 7px 8px; display: grid; grid-template-columns: 32px minmax(0,1fr) auto;
  align-items: center; gap: 8px; border: 1px solid transparent; border-radius: 6px; color: var(--text-secondary); background: transparent; text-align: left; cursor: pointer;
}
.layout-list button:hover { border-color: var(--border); background: var(--surface); }
.layout-icon { width: 32px; height: 32px; display: grid; place-items: center; border-radius: 5px; color: var(--info); background: var(--info-soft); }
.layout-list button > span:nth-child(2) { min-width: 0; display: flex; flex-direction: column; }
.layout-list strong { color: var(--text); font-size: 11px; }
.layout-list small { margin-top: 2px; overflow: hidden; color: var(--text-muted); font-size: 9px; text-overflow: ellipsis; white-space: nowrap; }

.outline-panel { background: var(--canvas); }
.panel-header {
  min-height: 58px; padding: 10px 14px; display: flex; align-items: center; justify-content: space-between; gap: 12px;
  border-bottom: 1px solid var(--border); background: var(--surface);
}
.panel-kicker { display: block; color: var(--text-muted); font-size: 9px; font-weight: 750; }
.panel-header h2 { margin: 2px 0 0; font-size: 13px; }
.add-slide-button {
  min-height: 38px; padding: 0 11px; display: inline-flex; align-items: center; gap: 7px;
  border: 1px solid var(--border-strong); border-radius: 6px; color: var(--text-secondary); background: var(--surface); font-size: 11px; font-weight: 700; cursor: pointer;
}
.add-slide-button:hover { border-color: var(--action-ink); color: var(--action-ink); }

.slide-list { min-height: 0; flex: 1; overflow: auto; padding: 14px; display: grid; align-content: start; gap: 9px; }
.slide-card {
  min-width: 0; min-height: 106px; display: grid; grid-template-columns: 30px minmax(0,1fr) auto;
  align-items: stretch; overflow: hidden; border: 1px solid var(--border); border-radius: 7px;
  background: var(--surface); box-shadow: var(--shadow-xs);
  transition: border-color var(--motion-fast), box-shadow var(--motion-fast), opacity var(--motion-fast);
}
.slide-card:hover { border-color: var(--border-strong); box-shadow: var(--shadow-sm); }
.slide-card.selected { border-color: var(--action-ink); box-shadow: 0 0 0 2px rgba(7,94,87,.1); }
.slide-card.drag-over { border-color: var(--info); box-shadow: 0 0 0 2px rgba(47,111,237,.14); }
.slide-card.is-dragging { opacity: .45; }
.drag-handle { display: grid; place-items: center; color: var(--text-disabled); cursor: grab; }
.slide-open-button {
  min-width: 0; padding: 10px 6px; display: grid; grid-template-columns: 112px minmax(0,1fr) auto;
  align-items: center; gap: 13px; border: 0; color: var(--text); background: transparent; text-align: left; cursor: pointer;
}
.slide-miniature { width: 112px; aspect-ratio: 16/9; display: grid; grid-template-columns: 1fr auto; align-items: end; padding: 8px; border: 1px solid var(--border); border-radius: 4px; color: var(--text-muted); background: var(--surface-muted); font-size: 10px; }
.slide-copy { min-width: 0; display: flex; flex-direction: column; }
.slide-copy strong { overflow: hidden; font-size: 13px; text-overflow: ellipsis; white-space: nowrap; }
.slide-copy small { margin-top: 3px; color: var(--action-ink); font-size: 10px; }
.slide-actions { padding: 8px 7px; display: grid; grid-template-columns: repeat(2,34px); align-content: center; gap: 3px; border-left: 1px solid var(--divider); }
.slide-actions button, .close-property {
  width: 34px; height: 34px; padding: 0; display: grid; place-items: center;
  border: 0; border-radius: 5px; color: var(--text-muted); background: transparent; cursor: pointer;
}
.slide-actions button:hover, .close-property:hover { color: var(--text); background: var(--surface-hover); }
.slide-actions button.danger:hover { color: var(--danger); background: var(--danger-soft); }

.empty-outline, .property-empty {
  min-height: 260px; padding: 40px 20px; display: flex; flex-direction: column; align-items: center; justify-content: center;
  color: var(--text-muted); text-align: center;
}
.empty-outline > span, .property-empty > span { width: 44px; height: 44px; display: grid; place-items: center; border-radius: 7px; color: var(--action-ink); background: var(--action-soft); }
.empty-outline h3, .property-empty h3 { margin: 13px 0 0; color: var(--text); font-size: 13px; }
.empty-outline p, .property-empty p { max-width: 280px; margin: 6px 0 0; font-size: 11px; line-height: 1.6; }
.empty-outline button { min-height: 38px; margin-top: 15px; padding: 0 12px; border: 1px solid var(--action-ink); border-radius: 6px; color: #fff; background: var(--action-ink); font-size: 11px; font-weight: 700; cursor: pointer; }

.property-body { min-height: 0; flex: 1; overflow: auto; padding: 14px; display: grid; align-content: start; gap: 16px; }
.field-group { min-width: 0; display: grid; gap: 7px; }
.field-group input, .field-group select, .field-group textarea {
  width: 100%; min-height: 42px; padding: 9px 10px; border: 1px solid var(--border-strong); border-radius: 6px;
  outline: 0; color: var(--text); background: var(--surface); font-size: 12px;
}
.field-group textarea { min-height: 142px; resize: vertical; line-height: 1.55; }
.field-group input:focus, .field-group select:focus, .field-group textarea:focus { border-color: var(--action-ink); box-shadow: 0 0 0 3px rgba(7,94,87,.09); }

.layout-guidance { padding: 11px; border-left: 3px solid var(--info); background: var(--info-soft); }
.layout-guidance > strong { color: var(--text); font-size: 11px; }
.layout-guidance > p { margin: 4px 0 10px; color: var(--text-secondary); font-size: 10px; line-height: 1.5; }
.layout-guidance > div { margin-top: 8px; }
.layout-guidance > div > span { color: var(--text-muted); font-size: 9px; font-weight: 750; }
.layout-guidance div p { margin: 5px 0 0; display: flex; flex-wrap: wrap; gap: 4px; }
.layout-guidance small { padding: 3px 5px; border-radius: 4px; color: var(--info); background: rgba(255,255,255,.72); font-size: 8px; }

.property-footer { padding: 10px 12px; display: grid; grid-template-columns: 1fr 1.25fr; gap: 7px; border-top: 1px solid var(--border); }
.property-footer button { min-height: 40px; border: 1px solid var(--border-strong); border-radius: 6px; color: var(--text-secondary); background: var(--surface); font-size: 11px; font-weight: 700; cursor: pointer; }
.property-footer button:hover { background: var(--surface-muted); }
.property-footer .save-property { border-color: var(--action-ink); color: #fff; background: var(--action-ink); }

.editor-scrim { display: none; }
@media (max-width:1240px) {
  .editor-workspace { grid-template-columns: 250px minmax(400px,1fr) 310px; }
  .toolbar-action.optional span { display:none; }
  .toolbar-action.optional { width:40px; padding:0; }
}
@media (max-width:1000px) {
  :global(.compose-workspace) { height:auto; min-height:calc(100dvh - 56px); overflow:visible; }
  .brief-strip { grid-template-columns:1fr 1.5fr; }
  .editor-workspace { min-height:760px; grid-template-columns:250px minmax(0,1fr); overflow:visible; }
  .property-panel {
    position:fixed; inset:0 0 0 auto; z-index:var(--z-modal); width:min(92vw,380px);
    transform:translateX(102%); box-shadow:var(--shadow-lg); transition:transform var(--motion-medium);
  }
  .property-panel.open { transform:translateX(0); }
  .editor-scrim { position:fixed; inset:0; z-index:calc(var(--z-modal) - 1); display:block; border:0; background:rgba(15,17,18,.5); cursor:pointer; }
}
@media (max-width:720px) {
  :global(.compose-workspace) { padding:10px; }
  .page-count, .outline-warning, .toolbar-action.optional { display:none; }
  .toolbar-action.primary { min-width:44px; padding:0 10px; }
  .toolbar-action.primary span { display:none; }
  .brief-strip { grid-template-columns:1fr; }
  .editor-workspace { min-height:0; display:block; border:0; background:transparent; }
  .resource-panel { max-height:370px; margin-bottom:10px; border:1px solid var(--border); border-radius:8px; overflow:hidden; }
  .outline-panel { min-height:560px; border:1px solid var(--border); border-radius:8px; overflow:hidden; }
  .slide-list { padding:9px; }
  .slide-card { grid-template-columns:24px minmax(0,1fr); }
  .slide-open-button { grid-template-columns:80px minmax(0,1fr); gap:9px; }
  .slide-miniature { width:80px; }
  .slide-actions { grid-column:1/-1; padding:5px 7px; grid-template-columns:repeat(4,40px); justify-content:end; border-top:1px solid var(--divider); border-left:0; }
  .slide-actions button { width:40px; height:40px; }
}
@media (max-width:420px) {
  .slide-card { min-height:0; }
  .slide-open-button { grid-template-columns:1fr; padding:9px; }
  .slide-miniature { width:100%; }
  .property-panel { width:100vw; }
}
</style>
