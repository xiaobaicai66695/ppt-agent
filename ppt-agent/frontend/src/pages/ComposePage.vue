<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { fetchPresets, fetchLayouts, fetchThemes, createTaskWithOutline, expandWithAI, generateOutlineWithAI } from '../api';
import type { PresetTemplate, AtomicLayout, ThemeInfo, TaskOutline, SlideOutline } from '../api';
import { authState } from '../stores/auth';

const router = useRouter();
const auth = authState;

// State
const activeTab = ref<'presets' | 'layouts'>('presets');
const presets = ref<PresetTemplate[]>([]);
const layouts = ref<AtomicLayout[]>([]);
const themes = ref<ThemeInfo[]>([]);
const loading = ref(false);

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

// Theme color map for preset thumbnails
const themeColors: Record<string, string> = {
  ocean_soft: '#0891b2', sage_calm: '#16a34a', charcoal_light: '#475569',
  government_red: '#dc2626', patriotic_blue: '#1d4ed8', warm_terracotta: '#ea580c',
  berry_cream: '#7c3aed', lavender_mist: '#9333ea', civic_gold: '#ca8a04',
  debate_purple: '#6d28d9', activity_orange: '#c2410c', report_green: '#15803d',
  simple_gray: '#374151', medical_blue: '#0284c7', finance_gold: '#b45309',
  education_blue: '#1e40af',
};

function getPresetGradient(palette: string): string {
  const color = themeColors[palette] || '#0891b2';
  const colorAlpha = color + '88';
  return 'linear-gradient(135deg, ' + color + ', ' + colorAlpha + ')';
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
  return layoutCategories.map(group => ({
    label: group.label,
    items: layouts.value.filter(l => group.names.includes(l.name)),
  })).filter(g => g.items.length > 0);
});

const contentTypeLabels: Record<string, string> = {
  title_slide: '封面', agenda: '目录', section_divider: '章节',
  content_slide: '内容', two_column: '双栏', three_column: '三栏',
  card_grid: '卡片', process_flow: '流程',
  stat_slide: '数据', quote_slide: '引用', image_text: '图文',
  case_study: '案例', kpi_dashboard: '仪表盘', summary_slide: '总结',
  deep_dive: '深度解析',
};

// Count of slides with empty descriptions
const emptyDescCount = computed(() => {
  return slides.value.filter(s => !s.description || !s.description.trim()).length;
});

const hasEmptyDescriptions = computed(() => emptyDescCount.value > 0);

// Load data
onMounted(async () => {
  loading.value = true;
  try {
    const [p, l, t] = await Promise.all([
      fetchPresets(),
      fetchLayouts(),
      fetchThemes(),
    ]);
    presets.value = p;
    layouts.value = l;
    themes.value = t;
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
  selectedTheme.value = preset.default_palette;
  slides.value = preset.default_slides.map(s => ({
    title: s.title,
    content_type: s.content_type,
    description: '', // 预设只提供结构骨架，内容由 AI 根据用户主题生成
  }));
}

function addSlideFromLayout(layout: AtomicLayout) {
  const slide: SlideOutline = {
    title: layout.display_name,
    content_type: layout.name,
    description: '',
  };
  if (selectedSlideIndex.value >= 0) {
    slides.value.splice(selectedSlideIndex.value + 1, 0, slide);
  } else {
    slides.value.push(slide);
  }
  selectedSlideIndex.value = slides.value.length - 1;
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
    const outline: TaskOutline = {
      template: selectedPreset.value?.name || 'custom',
      theme: selectedTheme.value,
      title: pptTitle.value || '未命名PPT',
      slides: slides.value.map(s => ({
        title: s.title,
        content_type: s.content_type,
        description: s.description || '',
      })),
    };
    const enriched = await generateOutlineWithAI(topicTrimmed.value, outline);
    if (enriched && enriched.length > 0) {
      // Merge enriched data back into slides, preserving title and content_type
      for (let i = 0; i < slides.value.length && i < enriched.length; i++) {
        const e = enriched[i];
        // Always update description and content_plan from AI result
        slides.value[i] = {
          title: slides.value[i].title,       // keep original title
          content_type: slides.value[i].content_type, // keep original content_type
          description: e.description || slides.value[i].description,
          content_plan: e.content_plan,
        };
      }
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

async function startGeneration() {
  if (slides.value.length === 0) return;

  // Warn if there are empty descriptions
  if (hasEmptyDescriptions.value) {
    const confirmed = confirm(
      `有 ${emptyDescCount.value} 页尚未填写内容描述，AI 将根据主题自动生成。\n\n是否继续？`
    );
    if (!confirmed) return;
  }

  try {
    const outline: TaskOutline = {
      template: selectedPreset.value?.name || 'custom',
      theme: selectedTheme.value,
      title: pptTitle.value || '未命名PPT',
      slides: slides.value.map(s => ({
        title: s.title,
        content_type: s.content_type,
        description: s.description,
        content_plan: s.content_plan,
      })),
    };
    const query = topicTrimmed.value;
    const task = await createTaskWithOutline(query, outline);
    router.push({ name: 'dashboard', query: { select: task.id } });
  } catch (e) {
    alert('创建任务失败: ' + (e as Error).message);
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
        <button class="btn-primary" :disabled="slides.length === 0 || loading" @click="startGeneration">
          {{ loading ? '创建中...' : '开始生成' }}
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
            <div
              v-for="preset in filteredPresets"
              :key="preset.name"
              :class="['preset-card', { selected: selectedPreset?.name === preset.name }]"
              @click="selectPreset(preset)"
            >
              <div class="preset-thumb" :style="{ background: getPresetGradient(preset.default_palette) }">
                <span class="preset-icon">📄</span>
              </div>
              <div class="preset-info">
                <div class="preset-name">{{ preset.display_name }}</div>
                <div class="preset-meta">{{ preset.slide_count }}页 · {{ categoryMap[preset.category] || preset.category }}</div>
              </div>
            </div>
          </div>
        </div>

        <!-- Layouts Tab -->
        <div v-if="activeTab === 'layouts'" class="layout-section">
          <div v-for="group in groupedLayouts" :key="group.label" class="layout-group">
            <div class="layout-group-label">{{ group.label }}</div>
            <div class="layout-list">
              <div
                v-for="layout in group.items"
                :key="layout.name"
                class="layout-item"
                draggable="true"
                @click="addSlideFromLayout(layout)"
              >
                <span class="layout-icon">◫</span>
                <span class="layout-name">{{ layout.display_name }}</span>
              </div>
            </div>
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
            @click="openSlideEditor(index)"
            @dragstart="onDragStart(index)"
            @dragover.prevent="onDragOver(index)"
            @drop.prevent="onDrop($event, index)"
            @dragend="draggingIndex = -1"
            @dragleave="onDragLeave"
          >
            <div class="slide-index">{{ index + 1 }}</div>
            <div class="slide-info">
              <div class="slide-title">{{ slide.title || '未命名' }}</div>
              <div class="slide-type">{{ contentTypeLabels[slide.content_type] || slide.content_type }}</div>
            </div>
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
      <div class="right-panel">
        <div v-if="editingSlide" class="slide-editor">
          <div class="editor-header">
            <span>编辑页面</span>
            <button class="btn-close" @click="cancelEdit">✕</button>
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
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
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
  gap: 12px;
  padding: 10px;
  border: 1px solid var(--border);
  border-radius: 8px;
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
}

.preset-icon {
  font-size: 18px;
}

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
}

.layout-item:hover {
  border-color: var(--color-primary);
  background: var(--bg-hover);
}

.layout-icon {
  font-size: 14px;
  opacity: 0.6;
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
  font-size: 16px;
}

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
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
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
</style>
