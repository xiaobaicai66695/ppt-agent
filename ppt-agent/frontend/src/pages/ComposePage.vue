<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { fetchPresets, fetchLayouts, fetchThemes, fetchBackgrounds, createTaskWithOutline, expandWithAI, generateOutlineWithAI } from '../api';
import type { PresetTemplate, AtomicLayout, ThemeInfo, BackgroundTheme, TaskOutline, SlideOutline } from '../api';
import { authState } from '../stores/auth';

const router = useRouter();
const auth = authState;

// State
const activeTab = ref<'presets' | 'layouts' | 'backgrounds'>('presets');
const presets = ref<PresetTemplate[]>([]);
const layouts = ref<AtomicLayout[]>([]);
const themes = ref<ThemeInfo[]>([]);
const backgrounds = ref<BackgroundTheme[]>([]);
const loading = ref(false);
const generating = ref(false);
const testingOutline = ref(false);

// Background category filter
const bgCategories = ['全部', '商务', '科技', '教育', '党政', '生活', '自然'];
const bgCategoryMap: Record<string, string[]> = {
  '全部': [],
  '商务': ['biz', 'office', 'meeting', 'report'],
  '科技': ['tech', 'digital', 'data', 'ai'],
  '教育': ['edu', 'school', 'learning', 'knowledge'],
  '党政': ['gov', 'party', 'official', 'ceremony'],
  '生活': ['life', 'lifestyle', 'travel', 'food'],
  '自然': ['nature', 'landscape', 'environment', 'green'],
};
const selectedBgCategory = ref('全部');

// Editing state
const selectedTheme = ref('ocean_soft');
const pptTitle = ref('');
const topicInput = ref('');
const topicTrimmed = computed(() => topicInput.value.trim());
const slides = ref<SlideOutline[]>([]);
const selectedSlideIndex = ref(-1);
const editingSlide = ref<SlideOutline | null>(null);
const expanding = ref(false);

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

function onPresetThumbError(e: Event) {
  const img = e.target as HTMLImageElement;
  img.style.display = 'none';
  img.parentElement?.classList.add('thumb-missing');
}

// Category filter
const categories = ['全部', 'tech', 'biz', 'edu', 'gov', 'work', 'other'];
const categoryMap: Record<string, string> = {
  '全部': '',
  'tech': '技术',
  'biz': '商务',
  'edu': '教育',
  'gov': '党政',
  'work': '工作',
  'other': '其他',
};
const selectedCategory = ref('全部');

const filteredPresets = computed(() => {
  if (selectedCategory.value === '全部') return presets.value;
  return presets.value.filter(p => p.category === selectedCategory.value);
});

const validLayoutNames = computed(() => new Set(layouts.value.map(l => l.name)));
const validBackgroundNames = computed(() => new Set(backgrounds.value.map(b => b.name)));

// Filter backgrounds by category/scenario
const filteredBackgrounds = computed(() => {
  const cats = bgCategoryMap[selectedBgCategory.value] || [];
  if (selectedBgCategory.value === '全部' || cats.length === 0) return backgrounds.value;
  return backgrounds.value.filter(bg => {
    const scenarios = bg.scenarios || [];
    return cats.some(c => scenarios.some(s => s.toLowerCase().includes(c.toLowerCase())));
  });
});

const selectedPreset = ref<PresetTemplate | null>(null);

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

function getBackgroundPreview(name: string): string {
  const bg = backgrounds.value.find(b => b.name === name);
  return bg?.preview_path || '';
}

function getBgScenarios(name: string): string[] {
  const bg = backgrounds.value.find(b => b.name === name);
  return bg?.scenarios?.slice(0, 4) || [];
}

function getBgDescription(name: string): string {
  const bg = backgrounds.value.find(b => b.name === name);
  return bg?.description || '';
}

function getBgDisplayName(name: string): string {
  const bg = backgrounds.value.find(b => b.name === name);
  return bg?.display_name || name;
}

function applyBackgroundToSlide(bgName: string) {
  if (selectedSlideIndex.value >= 0 && editingSlide.value) {
    editingSlide.value.background = bgName;
    saveSlideEdit();
  } else if (selectedSlideIndex.value >= 0) {
    // Apply directly without opening editor
    slides.value[selectedSlideIndex.value].background = bgName;
  }
}

// Count of slides with empty descriptions
const emptyDescCount = computed(() => {
  return slides.value.filter(s => !s.description || !s.description.trim()).length;
});

const hasEmptyDescriptions = computed(() => emptyDescCount.value > 0);

// Load data
onMounted(async () => {
  loading.value = true;
  try {
    const [p, l, t, b] = await Promise.all([
      fetchPresets(),
      fetchLayouts(),
      fetchThemes(),
      fetchBackgrounds(),
    ]);
    presets.value = p;
    layouts.value = l;
    themes.value = t;
    backgrounds.value = b;
  } catch (e) {
    console.error('Failed to load templates:', e);
  } finally {
    loading.value = false;
  }
});

function selectPreset(preset: PresetTemplate) {
  selectedPreset.value = preset;
  // Don't overwrite pptTitle with preset display name — the user's topic
  // (from topicTrimmed input) should be the title. Only use as fallback.
  if (!pptTitle.value) pptTitle.value = topicTrimmed.value || '';
  selectedTheme.value = preset.default_palette || 'ocean_soft';
  slides.value = preset.default_slides.map(s => ({
    title: s.title,
    content_type: s.content_type,
    description: s.description || '',
  }));
}

function buildOutline(): TaskOutline {
  return {
    template: selectedPreset.value?.name || 'custom',
    theme: selectedTheme.value || 'ocean_soft',
    title: pptTitle.value || topicTrimmed.value || '未命名PPT',
    slides: slides.value.map(s => ({
      title: s.title,
      content_type: s.content_type,
      description: s.description || '',
      content_plan: s.content_plan,
      background: s.background,
    })),
  };
}

function mergeEnrichedSlides(enriched: SlideOutline[]) {
  for (let i = 0; i < slides.value.length && i < enriched.length; i++) {
    const e = enriched[i];
    const nextType = e.content_type && validLayoutNames.value.has(e.content_type)
      ? e.content_type
      : slides.value[i].content_type;
    const nextBackground = e.background && validBackgroundNames.value.has(e.background)
      ? e.background
      : slides.value[i].background;
    slides.value[i] = {
      title: e.title || slides.value[i].title,
      content_type: nextType,
      background: nextBackground,
      description: e.description || slides.value[i].description,
      content_plan: e.content_plan || slides.value[i].content_plan,
    };
  }
}

async function ensureOutlineFilled(): Promise<TaskOutline> {
  if (!hasEmptyDescriptions.value && slides.value.every(s => s.content_plan)) {
    return buildOutline();
  }
  batchFilling.value = true;
  batchFillProgress.value = '正在补齐内容规划...';
  try {
    const enriched = await generateOutlineWithAI(topicTrimmed.value, buildOutline());
    if (enriched && enriched.length > 0) {
      mergeEnrichedSlides(enriched);
      batchFillProgress.value = `已完成 ${enriched.length} 页内容规划`;
    }
    return buildOutline();
  } finally {
    batchFilling.value = false;
  }
}

function addSlideFromLayout(layout: AtomicLayout) {
  const slide: SlideOutline = {
    title: layout.display_name,
    content_type: layout.name,
    description: '',
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

async function handleAIAutoFill() {
  if (!editingSlide.value || expanding.value) return;
  expanding.value = true;
  try {
    const result = await expandWithAI(
      editingSlide.value.title,
      editingSlide.value.content_type,
      editingSlide.value.description,
      selectedTheme.value
    );
    if (result) {
      editingSlide.value.description = result;
    }
  } catch (e) {
    console.error('AI expand failed:', e);
  } finally {
    expanding.value = false;
  }
}

// AI batch fill: generate descriptions + content_plans for all slides with empty descriptions
const batchFilling = ref(false);
const batchFillProgress = ref('');

async function handleAIBatchFill() {
  if (slides.value.length === 0 || batchFilling.value) return;
  if (!topicTrimmed.value) {
    alert('请先在上方填写 PPT 内容主题，AI 将据此生成各页面的具体内容');
    return;
  }
  batchFilling.value = true;
  batchFillProgress.value = '正在生成内容规划...';
  try {
    const enriched = await generateOutlineWithAI(topicTrimmed.value, buildOutline());
    if (enriched && enriched.length > 0) {
      mergeEnrichedSlides(enriched);
      batchFillProgress.value = `已完成 ${enriched.length} 页内容生成`;
    }
  } catch (e) {
    console.error('AI batch fill failed:', e);
    alert('批量生成失败: ' + (e as Error).message);
    batchFillProgress.value = '';
  } finally {
    batchFilling.value = false;
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
    description: '',
  };
  slides.value.push(slide);
}

async function testGenerateOutline() {
  if (slides.value.length === 0) {
    alert('请先添加或选择至少一页幻灯片');
    return;
  }
  if (!topicTrimmed.value) {
    alert('请先填写 PPT 内容主题');
    return;
  }

  testingOutline.value = true;
  try {
    const enriched = await generateOutlineWithAI(topicTrimmed.value, buildOutline());
    if (enriched && enriched.length > 0) {
      mergeEnrichedSlides(enriched);
      alert(`测试成功：已生成 ${enriched.length} 页内容规划`);
    } else {
      alert('测试失败：未返回有效大纲');
    }
  } catch (e) {
    alert('测试失败: ' + (e as Error).message);
  } finally {
    testingOutline.value = false;
  }
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
  try {
    const outline = await ensureOutlineFilled();
    const query = topicTrimmed.value;
    const task = await createTaskWithOutline(query, outline);
    router.push({ name: 'dashboard', query: { select: task.id } });
  } catch (e) {
    alert('创建任务失败: ' + (e as Error).message);
  } finally {
    generating.value = false;
  }
}
</script>

<template>
  <div class="compose-page">
    <!-- Top Toolbar -->
    <div class="compose-toolbar">
      <div class="toolbar-left">
        <button class="btn-back" @click="router.push({ name: 'dashboard' })">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M15 18l-6-6 6-6"/>
          </svg>
          返回
        </button>
        <span class="toolbar-title">模板编排</span>
      </div>
      <div class="toolbar-center">
        <input
          v-model="pptTitle"
          class="title-input"
          placeholder="输入PPT标题..."
        />
        <select v-model="selectedTheme" class="theme-select">
          <option v-for="t in themes" :key="t.name" :value="t.name">
            {{ t.display_name }}
          </option>
        </select>
      </div>
      <div class="toolbar-right">
        <span class="slide-count">{{ slides.length }} 页</span>
        <span v-if="hasEmptyDescriptions" class="empty-desc-hint">
          {{ emptyDescCount }} 页待填充
        </span>
        <button class="btn-ai-batch" :disabled="batchFilling || slides.length === 0 || !topicTrimmed" @click="handleAIBatchFill">
          {{ batchFilling ? batchFillProgress : 'AI 批量续写' }}
        </button>
        <button class="btn-secondary" :disabled="testingOutline || slides.length === 0 || !topicTrimmed" @click="testGenerateOutline">
          {{ testingOutline ? '测试中...' : '测试' }}
        </button>
        <button class="btn-primary" :disabled="slides.length === 0 || generating || !topicTrimmed" @click="startGeneration">
          {{ generating ? '创建中...' : '开始生成' }}
        </button>
      </div>
    </div>

    <!-- Topic input — the core content that drives generation -->
    <div class="topic-bar">
      <label class="topic-label" for="topic-input">
        <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true">
          <circle cx="8" cy="8" r="6.5"/>
          <path d="M8 5v3.5M8 10.5v.5"/>
        </svg>
        PPT 内容主题
      </label>
      <textarea
        id="topic-input"
        v-model="topicInput"
        class="topic-input"
        placeholder="描述你希望这个 PPT 呈现的内容，例如：新能源汽车行业发展趋势分析，包含市场概况、竞争格局、技术路线、政策环境、未来展望..."
        rows="2"
        :disabled="slides.length === 0"
      ></textarea>
      <p class="topic-hint">编排好幻灯片结构后，输入你的 PPT 内容主题，AI 将据此生成各页面的具体内容</p>
    </div>

    <!-- Main Layout -->
    <div class="compose-body">
      <!-- Left Panel: Template Library -->
      <div class="left-panel">
        <div class="panel-tabs">
          <button
            :class="['tab-btn', { active: activeTab === 'presets' }]"
            @click="activeTab = 'presets'"
          >预设模板</button>
          <button
            :class="['tab-btn', { active: activeTab === 'layouts' }]"
            @click="activeTab = 'layouts'"
          >原子布局</button>
          <button
            :class="['tab-btn', { active: activeTab === 'backgrounds' }]"
            @click="activeTab = 'backgrounds'"
          >背景图片</button>
        </div>

        <!-- Presets Tab -->
        <div v-if="activeTab === 'presets'" class="preset-section">
          <div class="category-filter">
            <button
              v-for="cat in categories"
              :key="cat"
              :class="['cat-btn', { active: selectedCategory === cat }]"
              @click="selectedCategory = cat"
            >
              {{ categoryMap[cat] || cat }}
            </button>
          </div>

          <div v-if="loading" class="loading-placeholder">加载中...</div>
          <div v-else class="preset-grid">
            <button
              v-for="preset in filteredPresets"
              :key="preset.name"
              :class="['preset-card', { selected: selectedPreset?.name === preset.name }]"
              type="button"
              @click="selectPreset(preset)"
            >
              <div
                :class="['preset-thumb', { 'thumb-missing': !preset.thumbnail }]"
              >
                <img
                  v-if="preset.thumbnail"
                  :src="preset.thumbnail"
                  :alt="preset.display_name"
                  class="preset-thumb-img"
                  loading="lazy"
                  @error="onPresetThumbError"
                />
                <span class="preset-missing-label">预览缺失</span>
              </div>
              <div class="preset-info">
                <div class="preset-name">{{ preset.display_name }}</div>
                <div class="preset-meta">{{ preset.slide_count }}页 · {{ categoryMap[preset.category] || preset.category }}</div>
              </div>
            </button>
          </div>
        </div>

        <!-- Layouts Tab -->
        <div v-if="activeTab === 'layouts'" class="layout-section">
          <div v-for="group in groupedLayouts" :key="group.label" class="layout-group">
            <div class="layout-group-label">{{ group.label }}</div>
            <div class="layout-list">
              <button
                v-for="layout in group.items"
                :key="layout.name"
                class="layout-item"
                type="button"
                @click="addSlideFromLayout(layout)"
              >
                <span class="layout-icon" aria-hidden="true"></span>
                <span class="layout-name">{{ layout.display_name }}</span>
              </button>
            </div>
          </div>
        </div>

        <!-- Backgrounds Tab -->
        <div v-if="activeTab === 'backgrounds'" class="background-section">
          <div class="bg-category-filter">
            <button
              v-for="cat in bgCategories"
              :key="cat"
              :class="['cat-btn', { active: selectedBgCategory === cat }]"
              @click="selectedBgCategory = cat"
            >
              {{ cat }}
            </button>
          </div>
          <div class="bg-grid">
            <button
              v-for="bg in filteredBackgrounds"
              :key="bg.name"
              :class="[
                'bg-card',
                { selected: selectedSlideIndex >= 0 && slides[selectedSlideIndex]?.background === bg.name }
              ]"
              type="button"
              @click="applyBackgroundToSlide(bg.name)"
              :title="bg.description"
            >
              <div class="bg-thumb-wrapper">
                <img
                  v-if="bg.preview_path"
                  :src="bg.preview_path"
                  :alt="bg.display_name"
                  class="bg-thumb"
                  @error="(e) => { (e.target as HTMLImageElement).style.display = 'none'; (e.target as HTMLImageElement).parentElement?.classList.add('bg-thumb-placeholder'); }"
                />
                <div v-else class="bg-thumb-placeholder">
                  <span>预览缺失</span>
                </div>
              </div>
              <div class="bg-info">
                <div class="bg-name">{{ bg.display_name }}</div>
                <div class="bg-scenarios">
                  <span v-for="s in bg.scenarios?.slice(0, 2)" :key="s" class="bg-tag">{{ s }}</span>
                </div>
              </div>
            </button>
          </div>
          <div v-if="selectedSlideIndex < 0 && filteredBackgrounds.length > 0" class="bg-hint">
            请先在中间选择一张幻灯片，再应用背景
          </div>
        </div>
      </div>

      <!-- Center: Slide Canvas -->
      <div class="center-panel">
        <div class="canvas-header">
          <span>幻灯片列表</span>
          <button class="btn-add-slide" @click="addBlankSlide">+ 添加页面</button>
        </div>
          <div class="slide-list">
          <div
            v-for="(slide, index) in slides"
            :key="slide.title + '-' + index"
            :class="['slide-card', { selected: selectedSlideIndex === index, 'drag-over': dragOverIndex === index, 'is-dragging': draggingIndex === index }]"
            draggable="true"
            @dragstart="onDragStart(index)"
            @dragover.prevent="onDragOver(index)"
            @drop.prevent="onDrop($event, index)"
            @dragend="draggingIndex = -1"
            @dragleave="onDragLeave"
          >
            <div class="slide-index">{{ index + 1 }}</div>
            <button class="slide-open-btn" type="button" @click="openSlideEditor(index)">
              <div class="slide-info">
                <div class="slide-title">{{ slide.title || '未命名' }}</div>
                <div class="slide-type">{{ getLayoutDisplayName(slide.content_type) }}</div>
                <div v-if="slide.background" class="slide-bg-badge">
                  <span class="bg-dot"></span>
                  {{ getBgDisplayName(slide.background) }}
                </div>
              </div>
            </button>
            <div class="slide-actions" @click.stop>
              <button class="action-btn" @click="moveSlide(index, 'up')" title="上移" aria-label="上移">
                <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
                  <path d="M8 12V4M4 8l4-4 4 4"/>
                </svg>
              </button>
              <button class="action-btn" @click="moveSlide(index, 'down')" title="下移" aria-label="下移">
                <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
                  <path d="M8 4v8M4 8l4 4 4-4"/>
                </svg>
              </button>
              <button class="action-btn" @click="duplicateSlide(index)" title="复制" aria-label="复制">
                <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true">
                  <rect x="5" y="5" width="9" height="9" rx="1.5"/><path d="M11 5V3.5A1.5 1.5 0 0 0 9.5 2h-6A1.5 1.5 0 0 0 2 3.5v6A1.5 1.5 0 0 0 3.5 11H5"/>
                </svg>
              </button>
              <button class="action-btn danger" @click="deleteSlide(index)" title="删除" aria-label="删除">
                <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true">
                  <path d="M3 4h10M5.5 4V3a.5.5 0 0 1 .5-.5h4a.5.5 0 0 1 .5.5v1M6 6v5M10 6v5M4 4l1 9a1 1 0 0 0 1 1h4a1 1 0 0 0 1-1l1-9"/>
                </svg>
              </button>
            </div>
          </div>

          <div v-if="slides.length === 0" class="empty-canvas">
            <p>从左侧选择一个预设模板开始<br/>或点击下方按钮添加空白幻灯片</p>
            <button class="btn-primary" @click="activeTab = 'presets'" style="margin-top: 1rem;">
              浏览预设模板
            </button>
          </div>
        </div>
      </div>

      <!-- Right: Slide Editor -->
      <button
        v-if="editingSlide"
        class="editor-scrim"
        type="button"
        aria-label="关闭页面编辑器"
        @click="cancelEdit"
      ></button>
      <div class="right-panel" :class="{ open: editingSlide }">
        <div v-if="editingSlide" class="slide-editor">
          <div class="editor-header">
            <span>编辑页面</span>
            <button class="btn-close" type="button" aria-label="关闭页面编辑器" @click="cancelEdit">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
                <path d="M6 6l12 12M18 6L6 18"/>
              </svg>
            </button>
          </div>
          <div class="editor-body">
            <div class="field-group">
              <label for="field-title">页面标题</label>
              <input
                id="field-title"
                v-model="editingSlide.title"
                class="field-input"
                placeholder="输入页面标题"
              />
            </div>
            <div class="field-group">
              <label for="field-type">布局类型</label>
              <select id="field-type" v-model="editingSlide.content_type" class="field-select">
                <option v-for="layout in layouts" :key="layout.name" :value="layout.name">
                  {{ layout.display_name }}
                </option>
              </select>
            </div>
            <section v-if="selectedLayout" class="layout-guidance" aria-label="布局容量说明">
              <div class="guidance-head">
                <strong>{{ selectedLayout.display_name }}</strong>
                <span>{{ selectedLayout.description }}</span>
              </div>
              <div v-if="requiredFields.length" class="guidance-row">
                <span class="guidance-label">必填字段</span>
                <span class="guidance-values">
                  <span v-for="field in requiredFields" :key="field.name" class="guidance-chip">
                    {{ field.label }}
                  </span>
                </span>
              </div>
              <div v-if="capacityGuidance.length" class="guidance-row">
                <span class="guidance-label">容量建议</span>
                <span class="capacity-list">
                  <span v-for="item in capacityGuidance" :key="item.key">
                    {{ item.label }} {{ item.value }}
                  </span>
                </span>
              </div>
            </section>
            <div class="field-group">
              <label for="field-bg">背景图片</label>
              <div class="bg-selector">
                <select id="field-bg" v-model="editingSlide.background" class="field-select">
                  <option value="">不使用背景</option>
                  <option v-for="bg in backgrounds" :key="bg.name" :value="bg.name">
                    {{ bg.display_name }}
                  </option>
                </select>
                <div v-if="editingSlide.background" class="bg-preview-thumb">
                  <img
                    :src="getBackgroundPreview(editingSlide.background)"
                    :alt="editingSlide.background"
                    @error="(e) => (e.target as HTMLImageElement).style.display = 'none'"
                  />
                </div>
              </div>
              <div v-if="editingSlide.background" class="bg-scenarios">
                <span v-for="s in getBgScenarios(editingSlide.background)" :key="s" class="bg-tag">{{ s }}</span>
              </div>
            </div>
            <div class="field-group">
              <label for="field-desc">内容描述</label>
              <textarea
                id="field-desc"
                v-model="editingSlide.description"
                class="field-textarea"
                placeholder="描述此页面的内容要点，AI将根据此描述生成具体内容..."
                rows="6"
              ></textarea>
            </div>
          </div>
          <div class="editor-footer">
            <button class="btn-cancel" @click="cancelEdit">取消</button>
            <button class="btn-save" @click="saveSlideEdit">保存</button>
            <button class="btn-ai" :disabled="expanding" @click="handleAIAutoFill">
              {{ expanding ? '生成中...' : 'AI 续写' }}
            </button>
          </div>
        </div>
        <div v-else class="editor-empty">
          <p>点击幻灯片卡片进行编辑</p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.compose-page {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: var(--bg-primary);
  color: var(--text-primary);
}

.compose-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 20px;
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border);
  gap: 16px;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 200px;
}

.btn-back {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 6px 12px;
  background: transparent;
  border: 1px solid var(--border);
  border-radius: 6px;
  color: var(--text-secondary);
  cursor: pointer;
  font-size: 13px;
}

.btn-back:hover {
  background: var(--bg-hover);
}

.toolbar-title {
  font-weight: 600;
  font-size: 15px;
}

.toolbar-center {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
  justify-content: center;
}

.title-input {
  padding: 8px 16px;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--bg-base);
  color: var(--text);
  font-size: 14px;
  width: 100%;
  max-width: 400px;
  text-align: center;
}

.title-input:focus {
  outline: none;
  border-color: var(--accent);
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
}

.theme-select {
  padding: 8px 12px;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--bg-base);
  color: var(--text);
  font-size: 13px;
  appearance: auto;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 200px;
  justify-content: flex-end;
}

.slide-count {
  font-size: 13px;
  color: var(--text-secondary);
}

.empty-desc-hint {
  font-size: 12px;
  color: #f59e0b;
  background: #fef3c7;
  padding: 3px 10px;
  border-radius: 12px;
  font-weight: 500;
}

.btn-ai-batch {
  padding: 8px 16px;
  background: var(--color-primary);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: opacity 0.2s;
}

.btn-ai-batch:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-primary {
  padding: 8px 20px;
  background: var(--color-primary);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Topic bar */
.topic-bar {
  padding: 0.75rem 1.5rem;
  border-bottom: 1px solid var(--border);
  background: var(--bg-base);
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  flex-wrap: wrap;
}
.topic-label {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.72rem;
  font-weight: 600;
  color: var(--text-secondary);
  white-space: nowrap;
  padding-top: 0.5rem;
  flex-shrink: 0;
}
.topic-label svg { width: 14px; height: 14px; color: var(--accent); }
.topic-input {
  flex: 1;
  min-width: 300px;
  padding: 0.55rem 0.8rem;
  border: 1.5px solid var(--border);
  border-radius: var(--radius);
  background: var(--bg-base);
  color: var(--text);
  font-size: 0.82rem;
  font-family: inherit;
  line-height: 1.5;
  resize: none;
  outline: none;
  transition: border-color var(--transition), box-shadow var(--transition);
}
.topic-input::placeholder { color: var(--text-disabled); }
.topic-input:focus {
  border-color: var(--accent);
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
}
.topic-input:disabled { opacity: 0.5; cursor: not-allowed; }
.topic-hint {
  width: 100%;
  font-size: 0.65rem;
  color: var(--text-muted);
  margin-top: 0.25rem;
  padding-left: calc(14px + 0.35rem + 0.75rem);
}

.compose-body {
  display: flex;
  flex: 1;
  overflow: hidden;
}

/* Left Panel */
.left-panel {
  width: 280px;
  min-width: 280px;
  background: var(--bg-secondary);
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.panel-tabs {
  display: flex;
  border-bottom: 1px solid var(--border);
}

.tab-btn {
  flex: 1;
  padding: 12px;
  background: transparent;
  border: none;
  color: var(--text-secondary);
  font-size: 13px;
  cursor: pointer;
  border-bottom: 2px solid transparent;
}

.tab-btn.active {
  color: var(--color-primary);
  border-bottom-color: var(--color-primary);
}

.category-filter {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  padding: 12px;
  border-bottom: 1px solid var(--border);
}

.cat-btn {
  padding: 4px 10px;
  background: transparent;
  border: 1px solid var(--border);
  border-radius: 16px;
  font-size: 12px;
  color: var(--text-secondary);
  cursor: pointer;
}

.cat-btn.active {
  background: var(--color-primary);
  color: white;
  border-color: var(--color-primary);
}

.preset-section, .layout-section {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
}

.loading-placeholder {
  padding: 24px;
  text-align: center;
  color: var(--text-secondary);
  font-size: 13px;
}

.preset-grid {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.preset-card {
  display: flex;
  width: 100%;
  gap: 12px;
  padding: 10px;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--bg-primary);
  color: var(--text-primary);
  text-align: left;
  cursor: pointer;
  transition: all 0.15s;
}

.preset-card:hover {
  border-color: var(--color-primary);
  background: var(--bg-hover);
}

.preset-card.selected {
  border-color: var(--color-primary);
  background: color-mix(in srgb, var(--color-primary) 10%, transparent);
}

.preset-thumb {
  width: 48px;
  height: 36px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  overflow: hidden;
  position: relative;
  background: var(--bg-muted);
  border: 1px solid var(--border);
}

.preset-thumb-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  position: absolute;
  inset: 0;
}

.preset-missing-label {
  display: none;
  padding: 0 4px;
  color: var(--text-muted);
  font-size: 9px;
  text-align: center;
}
.preset-thumb.thumb-missing .preset-missing-label { display: block; }

.preset-info {
  flex: 1;
  min-width: 0;
}

.preset-name {
  font-size: 13px;
  font-weight: 500;
  margin-bottom: 2px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.preset-meta {
  font-size: 11px;
  color: var(--text-secondary);
}

.layout-section {
  padding: 0;
}

.layout-group {
  padding: 8px 12px;
  border-bottom: 1px solid var(--border);
}

.layout-group:last-child {
  border-bottom: none;
}

.layout-group-label {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-secondary);
  text-transform: uppercase;
  margin-bottom: 8px;
}

.layout-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.layout-item {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 5px 10px;
  background: var(--bg-primary);
  border: 1px solid var(--border);
  border-radius: 6px;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.15s;
  color: var(--text-primary);
}

.layout-item:hover {
  border-color: var(--color-primary);
  background: var(--bg-hover);
}

.layout-icon {
  width: 14px;
  height: 10px;
  border: 1.5px solid currentColor;
  box-shadow: inset 4px 0 0 color-mix(in srgb, currentColor 20%, transparent);
  opacity: 0.65;
}

/* Center Panel */
.center-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: var(--bg-primary);
}

.canvas-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border);
  font-size: 13px;
  font-weight: 500;
}

.btn-add-slide {
  padding: 5px 12px;
  background: var(--color-primary);
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 12px;
  cursor: pointer;
}

.slide-list {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  display: flex;
  flex-wrap: wrap;
  align-content: flex-start;
  gap: 12px;
}

.slide-card {
  width: 160px;
  background: var(--bg-secondary);
  border: 2px solid var(--border);
  border-radius: 8px;
  padding: 10px;
  cursor: grab;
  transition: all 0.15s;
  position: relative;
}

.slide-card:active {
  cursor: grabbing;
}

.slide-card:hover {
  border-color: var(--color-primary);
}

.slide-card.selected {
  border-color: var(--color-primary);
  background: color-mix(in srgb, var(--color-primary) 8%, var(--bg-secondary));
}

.slide-card.drag-over {
  border-color: var(--color-primary);
  background: color-mix(in srgb, var(--color-primary) 15%, var(--bg-secondary));
  transform: scale(1.02);
}

.slide-card.is-dragging {
  opacity: 0.4;
  transform: scale(0.98);
}

.slide-index {
  position: absolute;
  top: -8px;
  left: -8px;
  width: 20px;
  height: 20px;
  background: var(--color-primary);
  color: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  font-weight: 600;
}

.slide-info {
  margin-top: 8px;
  margin-bottom: 8px;
}

.slide-title {
  font-size: 12px;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-bottom: 2px;
}

.slide-type {
  font-size: 11px;
  color: var(--text-secondary);
}

.slide-actions {
  display: flex;
  gap: 4px;
  opacity: 0;
  transition: opacity 0.15s;
}

.slide-card:hover .slide-actions {
  opacity: 1;
}

.slide-open-btn {
  width: 100%;
  padding: 0;
  border: 0;
  background: transparent;
  color: inherit;
  text-align: left;
  cursor: pointer;
  font: inherit;
}

.slide-open-btn:focus-visible,
.preset-card:focus-visible,
.layout-item:focus-visible,
.bg-card:focus-visible {
  outline: 3px solid color-mix(in srgb, var(--color-primary) 35%, transparent);
  outline-offset: 2px;
}

.action-btn {
  width: 26px;
  height: 26px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-base);
  border: 1px solid var(--border);
  border-radius: 4px;
  cursor: pointer;
  color: var(--text-secondary);
  padding: 0;
  transition: all var(--transition);
}
.action-btn svg { width: 14px; height: 14px; }

.action-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.action-btn.danger:hover {
  background: #fef2f2;
  color: #ef4444;
  border-color: #fecaca;
}

.empty-canvas {
  width: 100%;
  text-align: center;
  padding: 48px;
  color: var(--text-secondary);
  font-size: 13px;
  line-height: 1.8;
}

/* Right Panel */
.right-panel {
  width: 320px;
  min-width: 320px;
  background: var(--bg-secondary);
  border-left: 1px solid var(--border);
  display: flex;
  flex-direction: column;
}

.editor-scrim { display: none; }

.slide-editor {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.editor-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border);
  font-size: 14px;
  font-weight: 500;
}

.btn-close {
  background: transparent;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0;
}
.btn-close svg { width: 20px; height: 20px; }

.editor-body {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.field-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.field-group label {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary);
}

.field-input, .field-select, .field-textarea {
  padding: 8px 12px;
  border: 1px solid var(--border);
  border-radius: 6px;
  background: var(--bg-primary);
  color: var(--text-primary);
  font-size: 13px;
  font-family: inherit;
}

.field-input:focus, .field-select:focus, .field-textarea:focus {
  outline: none;
  border-color: var(--color-primary);
}

.field-textarea {
  resize: vertical;
  min-height: 120px;
  line-height: 1.6;
}

.layout-guidance {
  padding: 12px;
  border: 1px solid var(--border);
  border-radius: 6px;
  background: var(--bg-primary);
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.guidance-head {
  display: flex;
  flex-direction: column;
  gap: 3px;
}
.guidance-head strong { font-size: 12px; }
.guidance-head > span { color: var(--text-secondary); font-size: 11px; line-height: 1.5; }
.guidance-row { display: flex; flex-direction: column; gap: 6px; }
.guidance-label { color: var(--text-muted); font-size: 10px; font-weight: 600; }
.guidance-values, .capacity-list { display: flex; flex-wrap: wrap; gap: 5px; }
.guidance-chip, .capacity-list > span {
  padding: 3px 7px;
  border: 1px solid var(--border);
  border-radius: 4px;
  background: var(--bg-secondary);
  color: var(--text-secondary);
  font-size: 10px;
}

.bg-selector {
  display: flex;
  gap: 8px;
  align-items: center;
}

.bg-preview-thumb {
  width: 48px;
  height: 36px;
  border-radius: 4px;
  overflow: hidden;
  border: 1px solid var(--border);
  flex-shrink: 0;
}

.bg-preview-thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.bg-scenarios {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  margin-top: 4px;
}

.bg-tag {
  font-size: 10px;
  padding: 1px 6px;
  background: var(--bg-hover);
  border-radius: 4px;
  color: var(--text-secondary);
}

.slide-bg-badge {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 10px;
  color: var(--text-secondary);
  margin-top: 2px;
}

.bg-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--accent);
  flex-shrink: 0;
}

.editor-footer {
  padding: 12px 16px;
  border-top: 1px solid var(--border);
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}

.btn-cancel {
  padding: 8px 16px;
  background: transparent;
  border: 1px solid var(--border);
  border-radius: 6px;
  color: var(--text-secondary);
  font-size: 13px;
  cursor: pointer;
}

.btn-save {
  padding: 8px 20px;
  background: var(--color-primary);
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 13px;
  cursor: pointer;
}

.btn-ai {
  padding: 8px 20px;
  background: var(--color-primary);
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 13px;
  cursor: pointer;
}

.btn-ai:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.editor-empty {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-secondary);
  font-size: 13px;
}

/* Backgrounds Tab */
.background-section {
  flex: 1;
  overflow-y: auto;
  padding: 0;
}

.bg-category-filter {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  padding: 12px;
  border-bottom: 1px solid var(--border);
}

.bg-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
  padding: 12px;
}

.bg-card {
  border: 1.5px solid var(--border);
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  transition: all 0.15s;
  background: var(--bg-primary);
  color: var(--text-primary);
  padding: 0;
  text-align: left;
}

.bg-card:hover {
  border-color: var(--color-primary);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.bg-card.selected {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--color-primary) 25%, transparent);
}

.bg-thumb-wrapper {
  width: 100%;
  height: 72px;
  overflow: hidden;
  background: var(--bg-secondary);
}

.bg-thumb {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.bg-thumb-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-secondary);
  color: var(--text-muted);
  font-size: 11px;
}

@media (max-width: 1100px) {
  .left-panel { width: 260px; min-width: 260px; }
  .right-panel {
    position: fixed;
    top: 0;
    right: 0;
    bottom: 0;
    width: min(390px, 92vw);
    min-width: 0;
    z-index: 61;
    box-shadow: -12px 0 32px rgba(15, 23, 42, 0.16);
    transform: translateX(105%);
    transition: transform 0.2s ease;
  }
  .right-panel.open { transform: translateX(0); }
  .editor-scrim {
    display: block;
    position: fixed;
    inset: 0;
    z-index: 60;
    border: 0;
    background: rgba(15, 23, 42, 0.28);
  }
}

@media (max-width: 820px) {
  .compose-page { height: auto; min-height: 100dvh; overflow-x: hidden; }
  .compose-toolbar { flex-wrap: wrap; padding: 10px 12px; }
  .toolbar-left, .toolbar-center, .toolbar-right { width: 100%; min-width: 0; }
  .toolbar-left { justify-content: space-between; }
  .toolbar-center { order: 2; }
  .toolbar-right { order: 3; justify-content: flex-start; flex-wrap: wrap; gap: 8px; }
  .title-input { max-width: none; text-align: left; }
  .theme-select { flex: 0 0 auto; }
  .topic-bar { padding: 10px 12px; }
  .topic-input { width: 100%; min-width: 0; }
  .topic-hint { padding-left: 0; }
  .compose-body { flex-direction: column; overflow: visible; }
  .left-panel {
    width: 100%;
    min-width: 0;
    height: min(44vh, 410px);
    min-height: 300px;
    border-right: 0;
    border-bottom: 1px solid var(--border);
  }
  .preset-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .center-panel { min-height: 55vh; overflow: visible; }
  .slide-list {
    overflow: visible;
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
  .slide-card { width: auto; min-width: 0; }
  .slide-actions { opacity: 1; }
  .action-btn { width: 40px; height: 40px; }
  .right-panel { width: 100%; height: 100dvh; }
  .editor-header, .editor-footer { flex: 0 0 auto; background: var(--bg-secondary); }
  .editor-body { padding-bottom: 24px; }
  .btn-back, .btn-ai-batch, .btn-secondary, .btn-primary,
  .btn-add-slide, .btn-cancel, .btn-save, .btn-ai { min-height: 44px; }
}

@media (max-width: 520px) {
  .toolbar-center { flex-wrap: wrap; }
  .theme-select { width: 100%; min-height: 44px; }
  .preset-grid, .slide-list { grid-template-columns: minmax(0, 1fr); }
  .slide-card { padding: 14px; }
  .slide-actions { justify-content: flex-end; }
  .action-btn { width: 44px; height: 44px; }
  .editor-footer { justify-content: stretch; padding: 10px; }
  .editor-footer button { flex: 1; padding-left: 8px; padding-right: 8px; }
}

.bg-info {
  padding: 8px;
}

.bg-name {
  font-size: 11px;
  font-weight: 500;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-bottom: 4px;
}

.bg-scenarios {
  display: flex;
  flex-wrap: wrap;
  gap: 3px;
}

.bg-hint {
  margin: 12px;
  padding: 10px;
  background: #fffbeb;
  border: 1px solid #fde68a;
  border-radius: 6px;
  font-size: 11px;
  color: #92400e;
  text-align: center;
}
</style>
