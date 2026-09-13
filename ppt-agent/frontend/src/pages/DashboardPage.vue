<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Bot, CircleStop, FileDown, Image, MessageSquareText, Plus, RefreshCw, Send, Trash2, WandSparkles } from 'lucide-vue-next'
import AppShell from '../components/AppShell.vue'
import AppModal from '../components/AppModal.vue'
import ConversationTimelineItemCard from '../components/ConversationTimelineItem.vue'
import DeliveryFeedbackForm from '../components/DeliveryFeedbackForm.vue'
import TaskDeliveryPreview from '../components/TaskDeliveryPreview.vue'
import { cancelTask, continueTask, deleteTask, fetchConversation, fetchMe, fetchTasks, routeMessage, startTask, taskDownloadUrl } from '../api'
import type { AuthUser, TaskInfo, TaskStreamEvent } from '../types'
import { appendDeliveryDirectives, shouldStartPPTGeneration } from '../utils/messageRouting'
import { isTerminalTaskStreamEvent, taskStreamEventNames } from '../utils/taskStream'
import { appendExecutionStep, appendFinalAnswer, appendThought, appendTimelineError, appendTimelineMessage, appendToolCall, appendToolResult, finishStreamingEntries, finishTimelineEntries, resetConversationTimeline, toggleTimelineItem, type ConversationTimelineItem, type ExecutionState } from '../utils/conversationTimeline'

const router = useRouter()
const route = useRoute()
const user = ref<AuthUser>()
const tasks = ref<TaskInfo[]>([])
const selected = ref<TaskInfo>()
const timeline = ref<ConversationTimelineItem[]>([])
const prompt = ref('')
const busy = ref(false)
const mode = ref<'chat' | 'pptagent'>('chat')
const web = ref(false)
const images = ref(false)
const error = ref('')
const messagesContainer = ref<HTMLElement>()
const shouldFollowStream = ref(true)
const thumbnailRevision = ref(0)
const feedbackDialogOpen = ref(false)
const pendingDeletion = ref<TaskInfo>()
let source: EventSource | undefined
let reconnectTimer: ReturnType<typeof setTimeout> | undefined
let streamCursor = 0
let reconnectAttempts = 0
let streamGeneration = 0
let selectionGeneration = 0
let stickRequestPending = false

const activeTitle = computed(() => selected.value?.query || '新的创作会话')
const sorted = computed(() => [...tasks.value].sort((a, b) => Date.parse(b.updated_at || b.created_at) - Date.parse(a.updated_at || a.created_at)))
const hasTimeline = computed(() => busy.value || timeline.value.length > 0)
const hasDeliveryPreview = computed(() => selected.value?.status === 'completed' && Boolean(selected.value.files?.some(file => /\.pptx$/i.test(file))))
const taskLabel = (status: string) => ({ running: '生成中', completed: '已交付', paused_retryable: '可继续恢复', failed: '需要处理', conversation: '对话中', cancelled: '已取消' } as Record<string, string>)[status] || status
const toolLabel = (name = '') => ({ search: '联网检索', search_images: '图片搜索', generate_slide: '幻灯片渲染', slide_render: '幻灯片渲染', update_tasks_manifest: '写入任务清单', patch_tasks_draft: '修正规划草稿', read_file: '读取文件', shell: 'Shell', bash: 'Shell', command: '命令行', terminal: '终端' } as Record<string, string>)[name] || name || '调用工具'

function closeStream() {
  streamGeneration += 1
  source?.close()
  source = undefined
  if (reconnectTimer) clearTimeout(reconnectTimer)
  reconnectTimer = undefined
}

function addExecution(label: string, detail = '', state: ExecutionState = 'running', eventID?: number) {
  appendExecutionStep(timeline.value, label, detail, state, eventID)
  requestStickToLatestMessage()
}

function handleTimelineScroll() {
  const container = messagesContainer.value
  shouldFollowStream.value = !container || container.scrollHeight - container.scrollTop - container.clientHeight <= 96
}

function stickToLatestMessage() {
  const container = messagesContainer.value
  if (!container || !shouldFollowStream.value) return
  container.scrollTop = container.scrollHeight
}

// Coalesce scroll writes when SSE delivers many frames in a single turn.
function requestStickToLatestMessage() {
  if (stickRequestPending) return
  stickRequestPending = true
  void nextTick(() => {
    stickRequestPending = false
    stickToLatestMessage()
  })
}

function timelineMemoKey(item: ConversationTimelineItem) {
  if (item.type === 'message') return item.id
  if (item.type === 'thought') return `${item.id}:${item.content.length}:${item.state}:${item.expanded}:${item.streaming}`
  if (item.type === 'tool_call') return `${item.id}:${item.state}:${item.expanded}:${item.result?.length || 0}:${item.preview?.images?.length || 0}`
  if (item.type === 'final_answer') return `${item.id}:${item.content.length}:${item.streaming}`
  if (item.type === 'error') return `${item.id}:${item.content.length}`
  return `${item.id}:${item.state}:${item.detail || ''}`
}

async function loadTasks() {
  tasks.value = await fetchTasks()
}

async function select(task: TaskInfo) {
  const currentSelection = ++selectionGeneration
  closeStream()
  shouldFollowStream.value = true
  selected.value = task
  error.value = ''
  const session = await fetchConversation(task.id)
  if (currentSelection !== selectionGeneration || selected.value?.id !== task.id) return
  const history = session.messages || []
  timeline.value = resetConversationTimeline(history)
  if (session.conversation_streaming || task.status === 'running') {
    openStream(task.id, session.replay_after_event_id || session.latest_event_id || 0)
  }
}

function openStream(id: string, after = 0) {
  closeStream()
  const generation = streamGeneration
  streamCursor = after
  reconnectAttempts = 0
  busy.value = true
  const connect = () => {
    if (generation !== streamGeneration || selected.value?.id !== id) return
    source = new EventSource(`/api/tasks/${id}/stream${streamCursor ? `?after_id=${streamCursor}` : ''}`)
    const receive = (event: MessageEvent) => {
      const numericID = Number(event.lastEventId)
      // EventSource may redeliver the final frame while reconnecting. Event
      // ids are the authoritative client-side de-duplication boundary.
      if (Number.isFinite(numericID) && numericID > 0 && numericID <= streamCursor) return
      if (Number.isFinite(numericID) && numericID > streamCursor) streamCursor = numericID
      reconnectAttempts = 0
      consume(event.data, Number.isFinite(numericID) && numericID > 0 ? numericID : undefined)
    }
    for (const name of taskStreamEventNames) {
      source?.addEventListener(name, receive)
    }
    source.onerror = () => {
      source?.close()
      source = undefined
      if (generation !== streamGeneration || selected.value?.id !== id) return
      const delay = Math.min(10000, 500 * 2 ** reconnectAttempts++)
      reconnectTimer = setTimeout(() => {
        reconnectTimer = undefined
        connect()
      }, delay)
    }
  }
  connect()
}

function consume(raw: string, eventID?: number) {
  try {
    const data = JSON.parse(raw) as TaskStreamEvent
    const sourceID = eventID || data.id

    if (data.type === 'thought') {
      appendThought(timeline.value, data.content || data.phase_detail || data.message || '正在推进任务', {
        eventID: sourceID,
        segmentID: data.segment_id,
        phase: data.phase,
        delta: data.delta,
      })
    } else if (data.type === 'llm_start') {
      finishTimelineEntries(timeline.value, { includeTools: false })
    } else if (data.type === 'llm_delta') {
      const isThought = data.phase === 'analysis' || data.phase === 'reasoning' || data.phase === 'thought'
      const text = data.content || data.phase_detail || data.message || ''
      if (isThought) {
        appendThought(timeline.value, text || '正在推进任务', {
          eventID: sourceID,
          segmentID: data.segment_id,
          phase: data.phase,
          delta: data.delta ?? true,
        })
      } else {
        appendFinalAnswer(timeline.value, text, {
          eventID: sourceID,
          segmentID: data.segment_id || 'llm-answer',
          delta: data.delta ?? true,
        })
      }
    } else if (data.type === 'llm_end') {
      finishTimelineEntries(timeline.value, { includeTools: false })
    } else if (data.type === 'final_answer' || data.type === 'answer') {
      appendFinalAnswer(timeline.value, data.content || '', {
        eventID: sourceID,
        segmentID: data.segment_id || 'legacy-final-answer',
        delta: data.delta ?? true,
      })
    } else if (data.type === 'error') {
      error.value = data.error || '任务出现错误'
      appendTimelineError(timeline.value, error.value, sourceID)
      finishStreamingEntries(timeline.value)
    } else if (data.type === 'system_step') {
      appendThought(timeline.value, data.content || data.phase_detail || data.message || '正在推进任务', {
        eventID: sourceID,
        segmentID: data.segment_id || `legacy-system-${sourceID || 'step'}`,
        phase: data.phase || '系统步骤',
        delta: false,
      })
    } else if (data.type === 'tool_call') {
      const toolState: ExecutionState = data.tool_status === 'error' ? 'error' : data.tool_status === 'success' ? 'success' : 'running'
      appendToolCall(timeline.value, {
        eventID: sourceID,
        callID: data.tool_call_id,
        name: data.tool_name || 'unknown',
        label: toolLabel(data.tool_name),
        args: data.tool_args,
        detail: data.phase_detail || data.message || '正在调用工具',
      })
      // During rolling deploys, an older backend can still merge the observation
      // into a tool_call frame. Split it locally without changing arrival order.
      if (data.tool_result || toolState !== 'running') {
        appendToolResult(timeline.value, {
          eventID: sourceID === undefined ? undefined : `${sourceID}-result`,
          callID: data.tool_call_id,
          name: data.tool_name || 'unknown',
          label: toolLabel(data.tool_name),
          args: data.tool_args,
          result: data.tool_result || (toolState === 'error' ? '工具调用失败' : '工具调用已完成'),
          state: toolState,
          preview: data.tool_preview,
        })
      }
    } else if (data.type === 'tool_result') {
      const toolState: ExecutionState = data.tool_status === 'error' ? 'error' : 'success'
      appendToolResult(timeline.value, {
        eventID: sourceID,
        callID: data.tool_call_id,
        name: data.tool_name || 'unknown',
        label: toolLabel(data.tool_name),
        args: data.tool_args,
        result: data.tool_result || data.error || data.phase_detail || (toolState === 'error' ? '工具调用失败' : '工具调用已完成'),
        state: toolState,
        preview: data.tool_preview,
      })
    } else if (data.type === 'progress') {
      addExecution(data.phase_detail || data.phase || '推进生成', data.message || '', 'running', sourceID)
    } else if (data.type === 'file_ready') {
      addExecution('演示文件已生成', '可以下载并继续修改', 'success', sourceID)
      void refreshSelected()
    } else if (data.type === 'thumbnail_ready') {
      thumbnailRevision.value += 1
      addExecution('缩略图已就绪', data.files?.length ? `已准备 ${data.files.length} 张预览` : '可以查看演示预览', 'success', sourceID)
    } else if (data.type === 'answer_end') {
      finishTimelineEntries(timeline.value, { includeTools: false })
      addExecution('规划说明已完成', '正在开始生成演示页面', 'success', sourceID)
    } else if (isTerminalTaskStreamEvent(data.type)) {
      busy.value = false
      closeStream()
      finishStreamingEntries(timeline.value)
      const paused = data.status === 'paused_retryable'
      addExecution(data.type === 'conversation_complete' ? '回答完成' : paused ? '生成已暂停，可继续恢复' : '生成阶段结束', '', paused ? 'error' : 'success', sourceID)
      void refreshSelected(data.type !== 'conversation_complete')
    }
    requestStickToLatestMessage()
  } catch {
    // Ignore SSE keepalive frames.
  }
}

async function refreshSelected(promptForFeedback = false) {
  await loadTasks()
  if (!selected.value) return
  const fresh = tasks.value.find(task => task.id === selected.value?.id)
  if (!fresh) return
  selected.value = fresh
  if (promptForFeedback && fresh.status === 'completed' && !fresh.feedback) feedbackDialogOpen.value = true
}

async function submit() {
  const text = prompt.value.trim()
  if (!text || busy.value) return
  error.value = ''
  shouldFollowStream.value = true
  appendTimelineMessage(timeline.value, { role: 'user', content: text, timestamp: new Date().toISOString() })
  prompt.value = ''
  busy.value = true
  await nextTick()
  requestStickToLatestMessage()

  try {
    if (selected.value?.status === 'completed' || selected.value?.status === 'failed' || selected.value?.status === 'cancelled' || selected.value?.status === 'paused_retryable') {
      addExecution('继续处理任务')
      const result = await continueTask(selected.value.id, text)
      openStream(result.task_id, result.after_event_id || 0)
      return
    }
    const selectedMode = mode.value
    const routedText = appendDeliveryDirectives(text, selectedMode, web.value, images.value)
    const result = await routeMessage(routedText, selected.value?.id || '')
    await loadTasks()
    selected.value = tasks.value.find(task => task.id === result.task_id) || selected.value
    if (shouldStartPPTGeneration(result)) {
      addExecution('已识别为 PPT 生成', selectedMode === 'pptagent' ? '已按显式生成指令启动规划与交付' : '已按意图识别启动规划与交付')
      const started = await startTask(result.task_id)
      selected.value = started
      openStream(started.id)
      return
    }
    if (result.reply) appendTimelineMessage(timeline.value, { role: 'assistant', content: result.reply, timestamp: new Date().toISOString() })
    openStream(result.task_id, result.after_event_id || 0)
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : '无法提交这条消息'
    busy.value = false
    addExecution('提交失败', error.value, 'error')
  }
}

async function resumePausedTask() {
  const current = selected.value
  if (!current || current.status !== 'paused_retryable' || busy.value) return
  busy.value = true
  error.value = ''
  addExecution('正在恢复任务', '将从最近的规划或渲染检查点继续')
  try {
    const result = await continueTask(current.id, '继续任务')
    openStream(result.task_id, result.after_event_id || 0)
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : '无法恢复任务'
    busy.value = false
    addExecution('恢复失败', error.value, 'error')
  }
}

function requestRemove(task: TaskInfo) {
  pendingDeletion.value = task
}

async function remove() {
  const task = pendingDeletion.value
  if (!task) return
  await deleteTask(task.id)
  if (selected.value?.id === task.id) {
    selected.value = undefined
    timeline.value = []
  }
  pendingDeletion.value = undefined
  await loadTasks()
}

function handleFeedbackSaved(task: TaskInfo) {
  selected.value = task
  feedbackDialogOpen.value = false
  void loadTasks()
}

async function stop() {
  if (!selected.value) return
  await cancelTask(selected.value.id)
  closeStream()
  busy.value = false
  finishStreamingEntries(timeline.value)
  addExecution('任务已停止', '', 'success')
  await refreshSelected()
}

function newConversation() {
  closeStream()
  shouldFollowStream.value = true
  selected.value = undefined
  timeline.value = []
  error.value = ''
  prompt.value = String(route.query.brief || '')
  router.replace({ query: {} })
}

function toggleTimelineEntry(itemID: string) {
  toggleTimelineItem(timeline.value, itemID)
}

function handleComposerEnter(event: KeyboardEvent) {
  if (event.isComposing || event.shiftKey) return
  event.preventDefault()
  void submit()
}

onMounted(async () => {
  try {
    user.value = await fetchMe()
    await loadTasks()
    const brief = String(route.query.brief || '')
    const wanted = String(route.query.task || '')
    if (brief) newConversation()
    else {
      const target = tasks.value.find(task => task.id === wanted) || tasks.value[0]
      if (target) await select(target)
    }
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : '加载工作台失败'
  }
})

onBeforeUnmount(closeStream)
watch(() => route.query.brief, value => { if (value) newConversation() })
</script>

<template>
  <AppShell title="创作工作台" subtitle="对话、规划与交付在同一条轨迹里" :email="user?.email" :guest="user?.is_guest" @new="newConversation">
    <template #header><button class="compose-link" @click="router.push('/compose')"><WandSparkles :size="16" />编排页面</button></template>
    <div class="workbench">
      <aside class="conversations">
        <div class="conversation-head"><span>你的会话</span><button @click="newConversation" aria-label="新建会话"><Plus :size="17" /></button></div>
        <div class="task-list">
          <div v-for="task in sorted" :key="task.id" class="task-row" :class="{ selected: task.id === selected?.id }">
            <button type="button" class="task-select" @click="select(task)">
              <span class="status-dot" :class="task.status"></span><span class="task-info"><b>{{ task.query || '未命名会话' }}</b><small>{{ taskLabel(task.status) }} · {{ task.total_count ? `${task.done_count}/${task.total_count} 页` : '对话' }}</small></span>
            </button>
            <button type="button" class="delete-task" :aria-label="`删除会话：${task.query || '未命名会话'}`" @click="requestRemove(task)"><Trash2 :size="14" aria-hidden="true" /></button>
          </div>
          <p v-if="!tasks.length" class="empty-list">还没有会话。<br>从右侧写下第一个想法。</p>
        </div>
      </aside>
      <section class="canvas">
        <header class="canvas-head">
          <div><span class="canvas-kicker">{{ selected ? taskLabel(selected.status) : '准备就绪' }}</span><h2>{{ activeTitle }}</h2></div>
          <div v-if="selected" class="canvas-actions"><button v-if="selected.status === 'running'" class="outline-button" @click="stop"><CircleStop :size="15" />停止</button><button v-if="selected.status === 'paused_retryable'" class="outline-button" :disabled="busy" @click="resumePausedTask"><RefreshCw :size="15" />继续恢复</button><a v-for="file in selected.files" :key="file" class="download" :href="taskDownloadUrl(selected.id, file)"><FileDown :size="15" />下载</a></div>
        </header>
        <div ref="messagesContainer" class="messages" @scroll.passive="handleTimelineScroll">
          <div v-if="!hasTimeline" class="blank-canvas"><span><Bot :size="25" /></span><h3>从一个问题开始。</h3><p>可以让它解释、梳理资料，或直接开始一份演示。明确需求会让成稿更接近你的表达。</p><div><button @click="prompt = '为一场产品发布会规划 8 页叙事'">规划一份发布会演示</button><button @click="prompt = '总结这份资料的核心观点'">先梳理一个主题</button></div></div>
          <div v-if="timeline.length" class="timeline-list" role="list" aria-label="对话与执行时间线">
            <ConversationTimelineItemCard v-for="item in timeline" :key="item.id" v-memo="[timelineMemoKey(item)]" :item="item" @toggle="toggleTimelineEntry" />
          </div>
        </div>
        <form class="composer" novalidate @submit.prevent="submit">
          <div class="modebar"><button type="button" :class="{ on: mode === 'chat' }" @click="mode = 'chat'"><MessageSquareText :size="14" />对话</button><button type="button" :class="{ on: mode === 'pptagent' }" @click="mode = 'pptagent'"><WandSparkles :size="14" />PPT 生成</button><label><input v-model="web" type="checkbox">联网资料</label><label><input v-model="images" type="checkbox"><Image :size="13" />图片参考</label></div>
          <div class="composer-input"><textarea class="resize-none" v-model="prompt" rows="2" :disabled="busy" placeholder="写下你想完成的事… Enter 发送，Shift+Enter 换行" @keydown.enter="handleComposerEnter" /><button type="submit" :disabled="busy || !prompt.trim()" aria-label="发送消息"><Send :size="18" aria-hidden="true" /></button></div>
          <p v-if="error" class="inline-error">{{ error }}</p>
        </form>
      </section>
      <aside v-if="hasDeliveryPreview && selected" class="delivery-rail" aria-label="PPT 交付预览">
        <TaskDeliveryPreview :task="selected" :revision="thumbnailRevision" layout="side-rail" />
        <button type="button" class="feedback-trigger" @click="feedbackDialogOpen = true">{{ selected.feedback ? '修改评价' : '评价这份演示' }}</button>
      </aside>
    </div>
    <AppModal :open="feedbackDialogOpen" title="为这份演示评分" description="你的反馈会帮助我们改进下一次生成。" @close="feedbackDialogOpen = false"><DeliveryFeedbackForm v-if="selected" :task-id="selected.id" :feedback="selected.feedback" @saved="handleFeedbackSaved" /></AppModal>
    <AppModal :open="Boolean(pendingDeletion)" title="删除这个会话？" description="会话与已交付文件将被删除，且无法恢复。" @close="pendingDeletion = undefined"><div class="delete-confirm"><button type="button" class="outline-button" @click="pendingDeletion = undefined">取消</button><button type="button" class="danger-button" @click="remove">删除会话</button></div></AppModal>
  </AppShell>
</template>

<style scoped>
.compose-link { margin-left: auto; display: flex; align-items: center; gap: 6px; padding: 8px 10px; border: 1px solid var(--border-strong); border-radius: 6px; color: var(--accent-strong); background: var(--surface-raised); font-size: 12px; }
.workbench { display: grid; flex: 1; grid-template-columns: 254px minmax(460px, 1fr) minmax(300px, 360px); min-height: 0; overflow: hidden; background: var(--surface-base); }
.conversations { display: flex; flex-direction: column; min-height: 0; border-right: 1px solid var(--border-subtle); background: var(--surface-raised); }
.conversation-head { display: flex; align-items: center; justify-content: space-between; padding: 20px 16px 14px; color: var(--text-muted); font-size: 12px; }
.conversation-head button { display: grid; width: 27px; height: 27px; place-items: center; border: 0; border-radius: 5px; color: var(--accent-strong); background: var(--surface-accent); }
.task-list { min-height: 0; overflow: auto; padding: 0 8px 16px; }
.task-row { position: relative; display: grid; grid-template-columns: minmax(0, 1fr) 28px; align-items: center; border-radius: 6px; color: var(--text-muted); background: transparent; }
.task-row:hover, .task-row.selected { color: var(--text-strong); background: var(--surface-hover); }
.task-select { display: flex; min-width: 0; align-items: center; gap: 8px; padding: 11px 7px; border: 0; color: inherit; background: transparent; text-align: left; }
.task-select:focus-visible, .delete-task:focus-visible { position: relative; z-index: 1; }
.task-info { min-width: 0; flex: 1; }
.task-info b, .task-info small { display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.task-info b { font-size: 13px; font-weight: 600; }
.task-info small { margin-top: 3px; color: var(--text-subtle); font-size: 11px; }
.status-dot { width: 7px; height: 7px; border-radius: 50%; background: var(--text-subtle); }
.status-dot.running { background: var(--accent); box-shadow: 0 0 0 4px var(--accent-soft); }
.status-dot.completed { background: var(--info); }
.status-dot.failed { background: var(--danger); }
.delete-task { display: grid; width: 24px; height: 24px; place-items: center; border: 0; border-radius: 5px; color: var(--danger); background: transparent; opacity: 0; }
.delete-task:hover { background: color-mix(in srgb,var(--danger) 12%,transparent); opacity: 1; }
.task-row:hover .delete-task, .delete-task:focus-visible { opacity: 1; }
.empty-list { padding: 18px 8px; color: var(--text-subtle); font-size: 12px; line-height: 1.7; }
.canvas { display: grid; grid-template-rows: auto minmax(0, 1fr) auto; min-height: 0; background: var(--surface-base); }
.canvas-head { display: flex; justify-content: space-between; gap: 20px; padding: 17px 30px; border-bottom: 1px solid var(--border-subtle); }
.canvas-kicker { color: var(--accent); font: 500 10px 'DM Mono', monospace; letter-spacing: .08em; }
.canvas-head h2 { max-width: 690px; margin: 4px 0 0; overflow: hidden; color: var(--text-strong); font: 700 18px 'Noto Serif SC', serif; text-overflow: ellipsis; white-space: nowrap; }
.canvas-actions { display: flex; align-items: center; gap: 8px; }
.outline-button, .download, .feedback-trigger { display: flex; align-items: center; gap: 5px; padding: 7px 9px; border: 1px solid var(--border-strong); border-radius: 5px; color: var(--text-muted); background: var(--surface-raised); font-size: 12px; }
.outline-button:hover, .feedback-trigger:hover { color: var(--text-strong); background: var(--surface-hover); }
.download { border-color: var(--accent); color: var(--accent-on); background: var(--accent); }
.messages { min-height: 0; overflow: auto; padding: 27px max(25px, 7%); scrollbar-gutter: stable; }
.blank-canvas { max-width: 570px; margin: 8vh auto; text-align: center; }
.blank-canvas > span { display: grid; width: 53px; height: 53px; margin: auto; place-items: center; border-radius: 16px 16px 4px 16px; color: var(--accent-strong); background: var(--surface-accent); }
.blank-canvas h3 { margin: 19px 0 9px; color: var(--text-strong); font: 700 28px 'Noto Serif SC', serif; }
.blank-canvas p { max-width: 450px; margin: auto; color: var(--text-muted); font-size: 14px; line-height: 1.8; }
.blank-canvas div { display: flex; justify-content: center; gap: 8px; margin-top: 25px; }
.blank-canvas button { padding: 8px 10px; border: 1px solid var(--border-strong); border-radius: 5px; color: var(--text-muted); background: var(--surface-raised); font-size: 12px; }
.timeline-list { display: flow-root; max-width: 760px; margin: 0 auto; }
.delivery-rail { min-width: 0; overflow: auto; padding: 20px 16px; border-left: 1px solid var(--border-subtle); background: var(--surface-raised); scrollbar-gutter: stable; }
.feedback-trigger { width: 100%; justify-content: center; margin-top: 12px; color: var(--accent); font-weight: 600; }
.composer { padding: 12px 30px 18px; border-top: 1px solid var(--border-subtle); background: var(--surface-base); }
.modebar { display: flex; align-items: center; gap: 5px; margin: 0 0 7px; }
.modebar button, .modebar label { display: flex; align-items: center; gap: 4px; padding: 5px 7px; border: 0; color: var(--text-subtle); background: transparent; font-size: 11px; }
.modebar button.on { border-radius: 4px; color: var(--accent-strong); background: var(--surface-accent); }
.modebar label { margin-left: 5px; }
.modebar input { accent-color: var(--accent); }
.composer-input { display: grid; grid-template-columns: 1fr 43px; gap: 8px; padding: 8px; border: 1px solid var(--border-subtle); border-radius: 8px; background: var(--message-assistant); }
.composer textarea { width: 100%; min-height: 39px; padding: 6px 8px; resize: none; border: 0; outline: 0; color: var(--text-strong); background: transparent; line-height: 1.5; }
.composer-input button { display: grid; align-self: end; height: 39px; place-items: center; border: 0; border-radius: 5px; color: var(--accent-on); background: var(--accent); }
.composer-input button:disabled { cursor: not-allowed; opacity: .4; }
.inline-error { margin: 7px 0 0; color: var(--danger); font-size: 12px; }
.delete-confirm { display: flex; justify-content: flex-end; gap: 9px; }
.danger-button { padding: 7px 10px; border: 1px solid var(--danger); border-radius: 5px; color: #411515; background: var(--danger); font-size: 12px; font-weight: 700; }
.danger-button:hover { filter: brightness(1.08); }
@media (max-width: 1280px) {
  .workbench { grid-template-columns: minmax(0, 1fr) minmax(300px, 350px); }
  .conversations { display: none; }
}
@media (max-width: 920px) {
  .workbench { grid-template-columns: 1fr; grid-template-rows: minmax(520px, 1fr) auto; overflow: auto; }
  .delivery-rail { max-height: none; padding: 18px; border-top: 1px solid var(--border-subtle); border-left: 0; }
}
@media (max-width: 850px) {
  .canvas-head, .composer { padding-right: 18px; padding-left: 18px; }
  .messages { padding: 22px 18px; }
  .modebar label { display: none; }
}
</style>
