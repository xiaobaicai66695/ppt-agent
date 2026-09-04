<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Bot, CheckCircle2, ChevronDown, ChevronRight, CircleStop, FileDown, Image, LoaderCircle, MessageSquareText, Plus, RefreshCw, Send, Trash2, WandSparkles } from 'lucide-vue-next'
import AppShell from '../components/AppShell.vue'
import AppModal from '../components/AppModal.vue'
import MarkdownContent from '../components/MarkdownContent.vue'
import DeliveryFeedbackForm from '../components/DeliveryFeedbackForm.vue'
import TaskDeliveryPreview from '../components/TaskDeliveryPreview.vue'
import { cancelTask, continueTask, deleteTask, fetchConversation, fetchMe, fetchTasks, routeMessage, startTask, taskDownloadUrl } from '../api'
import type { AuthUser, TaskInfo } from '../types'
import { shouldStartPPTGeneration } from '../utils/messageRouting'
import { isTerminalTaskStreamEvent } from '../utils/taskStream'
import { appendExecutionStep, appendTimelineMessage, appendToolInvocation, beginObservablePhase, completeObservablePhase, hideCompletedToolTraces, resetConversationTimeline, resolveToolInvocation, toggleToolInvocation, type ConversationTimelineItem, type ExecutionState, type ToolPreview } from '../utils/conversationTimeline'

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
const streamingMessageID = ref<string>()
const activePhaseID = ref<string>()
const activePhase = ref('')
const thumbnailRevision = ref(0)
const feedbackDialogOpen = ref(false)
const pendingDeletion = ref<TaskInfo>()
let source: EventSource | undefined

const activeTitle = computed(() => selected.value?.query || '新的创作会话')
const sorted = computed(() => [...tasks.value].sort((a, b) => Date.parse(b.updated_at || b.created_at) - Date.parse(a.updated_at || a.created_at)))
const hasTimeline = computed(() => busy.value || timeline.value.length > 0)
const taskLabel = (status: string) => ({ running: '生成中', completed: '已交付', paused_retryable: '可继续恢复', failed: '需要处理', conversation: '对话中', cancelled: '已取消' } as Record<string, string>)[status] || status
const phaseLabel = (phase = '') => ({ analysis: '分析请求', answer: '组织回答' } as Record<string, string>)[phase] || '处理任务'
const toolLabel = (name = '') => ({ search: '联网检索', search_images: '图片搜索' } as Record<string, string>)[name] || name || '调用工具'
const executionLabel = (state: ExecutionState) => ({ running: '执行中', success: '已完成', error: '失败' } as Record<ExecutionState, string>)[state]

function closeStream() {
  source?.close()
  source = undefined
}

function addExecution(label: string, detail = '', state: ExecutionState = 'running') {
  appendExecutionStep(timeline.value, label, detail, state)
  void nextTick(stickToLatestMessage)
}

function stickToLatestMessage() {
  const container = messagesContainer.value
  if (!container || container.scrollHeight - container.scrollTop - container.clientHeight > 96) return
  container.scrollTop = container.scrollHeight
}

function appendAssistantChunk(content: string) {
  if (!content) return
  const item = timeline.value.find((candidate): candidate is Extract<ConversationTimelineItem, { type: 'message' }> => candidate.type === 'message' && candidate.id === streamingMessageID.value)
  if (item?.message.role === 'assistant') {
    item.message.content += content
  } else {
    appendTimelineMessage(timeline.value, { role: 'assistant', content, timestamp: new Date().toISOString() })
    streamingMessageID.value = timeline.value[timeline.value.length - 1]?.id
  }
  void nextTick(stickToLatestMessage)
}

async function loadTasks() {
  tasks.value = await fetchTasks()
}

async function select(task: TaskInfo) {
  closeStream()
  streamingMessageID.value = undefined
  activePhaseID.value = undefined
  activePhase.value = ''
  selected.value = task
  error.value = ''
  const session = await fetchConversation(task.id)
  const history = session.messages || []
  timeline.value = resetConversationTimeline(history)
  if (session.full_answer && !history.some(message => message.role === 'assistant' && message.content === session.full_answer)) {
    appendTimelineMessage(timeline.value, { role: 'assistant', content: session.full_answer, timestamp: '' })
  }
  if (session.conversation_streaming || task.status === 'running') {
    openStream(task.id, session.replay_after_event_id || session.latest_event_id || 0)
  }
}

function openStream(id: string, after = 0) {
  closeStream()
  busy.value = true
  source = new EventSource(`/api/tasks/${id}/stream${after ? `?after_id=${after}` : ''}`)
  const receive = (event: MessageEvent) => consume(event.data)
  for (const name of ['answer', 'answer_end', 'system_step', 'tool_call', 'tool_result', 'progress', 'runtime_event', 'file_ready', 'thumbnail_ready', 'error', 'complete', 'continue_complete', 'continue_queued', 'conversation_complete']) {
    source.addEventListener(name, receive)
  }
  source.onerror = () => {
    closeStream()
    busy.value = false
    addExecution('连接已结束', '已从会话快照恢复', 'success')
    void refreshSelected(true)
  }
}

function startPhase(phase: string, detail: string) {
  activePhase.value = phase
  activePhaseID.value = beginObservablePhase(timeline.value, phase, phaseLabel(phase), detail)
  void nextTick(stickToLatestMessage)
}

function ensureAnalysisPhase() {
  if (!activePhaseID.value) startPhase('analysis', '正在分析请求与可用工具')
}

function consume(raw: string) {
  try {
    const data = JSON.parse(raw) as {
      type?: string
      content?: string
      error?: string
      message?: string
      phase?: string
      phase_detail?: string
      tool_name?: string
      files?: string[]
		status?: string
      runtime_event?: { kind?: string; name?: string; detail?: string; status?: string }
      tool_preview?: ToolPreview
    }

    if (data.type === 'answer') {
      appendAssistantChunk(data.content || '')
    } else if (data.type === 'error') {
      error.value = data.error || '任务出现错误'
      if (activePhase.value) completeObservablePhase(timeline.value, activePhase.value, 'error')
      addExecution('生成遇到错误', error.value, 'error')
    } else if (data.type === 'runtime_event') {
      const event = data.runtime_event
      const state: ExecutionState = event?.status === 'error' || event?.status === 'failed' ? 'error' : event?.status === 'ok' ? 'success' : 'running'
      addExecution(event?.name || event?.kind || '处理任务', event?.detail || '', state)
    } else if (data.type === 'system_step') {
      startPhase(data.phase || 'system', data.phase_detail || data.message || '正在推进')
    } else if (data.type === 'tool_call') {
      ensureAnalysisPhase()
      appendToolInvocation(timeline.value, activePhaseID.value, data.tool_name || 'unknown', toolLabel(data.tool_name), data.phase_detail || data.message || '正在调用', data.tool_preview)
      void nextTick(stickToLatestMessage)
    } else if (data.type === 'tool_result') {
      ensureAnalysisPhase()
      resolveToolInvocation(timeline.value, activePhaseID.value, data.tool_name || 'unknown', toolLabel(data.tool_name), data.error || data.phase_detail || data.message || '已完成', data.error ? 'error' : 'success', data.tool_preview)
      void nextTick(stickToLatestMessage)
    } else if (data.type === 'progress') {
      addExecution(data.phase_detail || data.phase || '推进生成', data.message || '')
    } else if (data.type === 'file_ready') {
      addExecution('演示文件已生成', '可以下载并继续修改', 'success')
      void refreshSelected()
    } else if (data.type === 'thumbnail_ready') {
      thumbnailRevision.value += 1
      addExecution('缩略图已就绪', data.files?.length ? `已准备 ${data.files.length} 张预览` : '可以查看演示预览', 'success')
    } else if (data.type === 'answer_end') {
      streamingMessageID.value = undefined
      completeObservablePhase(timeline.value, 'answer')
      addExecution('规划说明已完成', '正在开始生成演示页面', 'success')
    } else if (isTerminalTaskStreamEvent(data.type)) {
      busy.value = false
      closeStream()
      streamingMessageID.value = undefined
      completeObservablePhase(timeline.value, 'answer')
      hideCompletedToolTraces(timeline.value)
		const paused = data.status === 'paused_retryable'
      addExecution(data.type === 'conversation_complete' ? '回答完成' : paused ? '生成已暂停，可继续恢复' : '生成阶段结束', '', paused ? 'error' : 'success')
      void refreshSelected(data.type !== 'conversation_complete')
    }
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
  streamingMessageID.value = undefined
  activePhaseID.value = undefined
  activePhase.value = ''
  appendTimelineMessage(timeline.value, { role: 'user', content: text, timestamp: new Date().toISOString() })
  prompt.value = ''
  busy.value = true
  await nextTick(stickToLatestMessage)

  try {
    if (selected.value?.status === 'completed' || selected.value?.status === 'failed' || selected.value?.status === 'cancelled' || selected.value?.status === 'paused_retryable') {
      addExecution('继续处理任务')
      const result = await continueTask(selected.value.id, text)
      openStream(result.task_id, result.after_event_id || 0)
      return
    }
    const selectedMode = mode.value
    const result = await routeMessage(text, selected.value?.id || '', selectedMode, web.value, images.value)
    await loadTasks()
    selected.value = tasks.value.find(task => task.id === result.task_id) || selected.value
    if (shouldStartPPTGeneration(result, selectedMode)) {
      addExecution('已识别为 PPT 生成', selectedMode === 'pptagent' ? '已按手动选择直接创建' : '已按意图识别启动规划与交付')
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
    streamingMessageID.value = undefined
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
  hideCompletedToolTraces(timeline.value)
  addExecution('任务已停止', '', 'success')
  await refreshSelected()
}

function newConversation() {
  closeStream()
  streamingMessageID.value = undefined
  activePhaseID.value = undefined
  activePhase.value = ''
  selected.value = undefined
  timeline.value = []
  error.value = ''
  prompt.value = String(route.query.brief || '')
  router.replace({ query: {} })
}

function toggleTool(toolID: string) {
  toggleToolInvocation(timeline.value, toolID)
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
          <button v-for="task in sorted" :key="task.id" class="task-row" :class="{ selected: task.id === selected?.id }" @click="select(task)">
            <span class="status-dot" :class="task.status"></span><span class="task-info"><b>{{ task.query || '未命名会话' }}</b><small>{{ taskLabel(task.status) }} · {{ task.total_count ? `${task.done_count}/${task.total_count} 页` : '对话' }}</small></span><Trash2 :size="14" class="delete" @click.stop="requestRemove(task)" />
          </button>
          <p v-if="!tasks.length" class="empty-list">还没有会话。<br>从右侧写下第一个想法。</p>
        </div>
      </aside>
      <section class="canvas">
        <header class="canvas-head">
          <div><span class="canvas-kicker">{{ selected ? taskLabel(selected.status) : '准备就绪' }}</span><h2>{{ activeTitle }}</h2></div>
          <div v-if="selected" class="canvas-actions"><button v-if="selected.status === 'running'" class="outline-button" @click="stop"><CircleStop :size="15" />停止</button><button v-if="selected.status === 'paused_retryable'" class="outline-button" :disabled="busy" @click="resumePausedTask"><RefreshCw :size="15" />继续恢复</button><a v-for="file in selected.files" :key="file" class="download" :href="taskDownloadUrl(selected.id, file)"><FileDown :size="15" />下载</a></div>
        </header>
        <div ref="messagesContainer" class="messages">
          <div v-if="!hasTimeline" class="blank-canvas"><span><Bot :size="25" /></span><h3>从一个问题开始。</h3><p>可以让它解释、梳理资料，或直接开始一份演示。明确需求会让成稿更接近你的表达。</p><div><button @click="prompt = '为一场产品发布会规划 8 页叙事'">规划一份发布会演示</button><button @click="prompt = '总结这份资料的核心观点'">先梳理一个主题</button></div></div>
          <template v-for="item in timeline" :key="item.id">
            <article v-if="item.type === 'message'" class="message" :class="item.message.role">
              <div class="message-label">{{ item.message.role === 'user' ? '你' : 'Deckform' }}</div>
              <MarkdownContent :content="item.message.content" :streaming="busy && item.id === streamingMessageID" />
            </article>
            <section v-else-if="item.type === 'phase'" class="observable-phase" :class="item.state" :aria-label="`${item.label}阶段`">
              <header class="phase-head"><span class="trace-state" :class="item.state"><LoaderCircle v-if="item.state === 'running'" :size="14" /><CheckCircle2 v-else :size="14" /></span><div><span class="phase-kicker">可观察步骤</span><b>{{ item.label }}</b><small v-if="item.detail">{{ item.detail }}</small></div></header>
              <div v-if="item.tools.length" class="tool-list">
                <article v-for="tool in item.tools" :key="tool.id" class="tool-invocation" :class="tool.state">
                  <button type="button" class="tool-toggle" :aria-expanded="tool.expanded" :aria-controls="`tool-detail-${tool.id}`" @click="toggleTool(tool.id)">
                    <span class="trace-state" :class="tool.state"><LoaderCircle v-if="tool.state === 'running'" :size="14" /><CheckCircle2 v-else :size="14" /></span><span class="tool-copy"><small>工具调用</small><b>{{ tool.label }}</b></span><span class="tool-status">{{ executionLabel(tool.state) }}</span><ChevronDown v-if="tool.expanded" :size="16" aria-hidden="true" /><ChevronRight v-else :size="16" aria-hidden="true" />
                  </button>
                  <div v-if="tool.expanded" :id="`tool-detail-${tool.id}`" class="tool-detail">
                    <p v-if="tool.callDetail"><span>调用说明</span>{{ tool.callDetail }}</p><p v-if="tool.resultDetail"><span>{{ tool.state === 'error' ? '执行结果' : '已获得' }}</span>{{ tool.resultDetail }}</p><div v-if="tool.preview?.images?.length" class="tool-image-preview"><a v-for="(image, index) in tool.preview.images" :key="image.image_url || image.thumbnail_url || index" :href="image.source_url || image.image_url || image.thumbnail_url" target="_blank" rel="noopener noreferrer"><img :src="image.thumbnail_url || image.image_url" :alt="image.alt || '图片工具结果预览'" width="136" height="88"><small v-if="image.attribution">{{ image.attribution }}</small></a></div><p v-if="!tool.resultDetail && tool.state === 'running'"><span>当前状态</span>正在等待工具返回结果</p>
                  </div>
                </article>
              </div>
            </section>
            <div v-else class="execution-event" :class="item.state"><span class="trace-state" :class="item.state"><LoaderCircle v-if="item.state === 'running'" :size="13" /><CheckCircle2 v-else :size="13" /></span><b>{{ item.label }}</b><em v-if="item.detail">{{ item.detail }}</em></div>
          </template>
          <TaskDeliveryPreview v-if="selected?.status === 'completed'" :task="selected" :revision="thumbnailRevision" />
          <button v-if="selected?.status === 'completed'" type="button" class="feedback-trigger" @click="feedbackDialogOpen = true">{{ selected.feedback ? '修改评价' : '评价这份演示' }}</button>
        </div>
        <form class="composer" novalidate @submit.prevent="submit">
          <div class="modebar"><button type="button" :class="{ on: mode === 'chat' }" @click="mode = 'chat'"><MessageSquareText :size="14" />对话</button><button type="button" :class="{ on: mode === 'pptagent' }" @click="mode = 'pptagent'"><WandSparkles :size="14" />PPT 生成</button><label><input v-model="web" type="checkbox">联网资料</label><label><input v-model="images" type="checkbox"><Image :size="13" />图片参考</label></div>
          <div class="composer-input"><textarea class="resize-none" v-model="prompt" rows="2" :disabled="busy" placeholder="写下你想完成的事…" @keydown.enter="handleComposerEnter" /><button type="submit" :disabled="busy || !prompt.trim()"><Send :size="18" /></button></div>
          <p v-if="error" class="inline-error">{{ error }}</p>
        </form>
      </section>
    </div>
    <AppModal :open="feedbackDialogOpen" title="为这份演示评分" description="你的反馈会帮助我们改进下一次生成。" @close="feedbackDialogOpen = false"><DeliveryFeedbackForm v-if="selected" :task-id="selected.id" :feedback="selected.feedback" @saved="handleFeedbackSaved" /></AppModal>
    <AppModal :open="Boolean(pendingDeletion)" title="删除这个会话？" description="会话与已交付文件将被删除，且无法恢复。" @close="pendingDeletion = undefined"><div class="delete-confirm"><button type="button" class="outline-button" @click="pendingDeletion = undefined">取消</button><button type="button" class="danger-button" @click="remove">删除会话</button></div></AppModal>
  </AppShell>
</template>

<style scoped>
.compose-link{margin-left:auto;display:flex;align-items:center;gap:6px;border:1px solid var(--border-strong);border-radius:6px;padding:8px 10px;color:var(--accent-strong);background:var(--surface-raised);font-size:12px}.workbench{min-height:0;flex:1;display:grid;grid-template-columns:254px minmax(0,1fr);overflow:hidden;background:var(--surface-base)}.conversations{min-height:0;display:flex;flex-direction:column;border-right:1px solid var(--border-subtle);background:var(--surface-raised)}.conversation-head{display:flex;align-items:center;justify-content:space-between;padding:20px 16px 14px;color:var(--text-muted);font-size:12px}.conversation-head button{display:grid;place-items:center;width:27px;height:27px;color:var(--accent-strong);background:var(--surface-accent);border:0;border-radius:5px}.task-list{min-height:0;overflow:auto;padding:0 8px 16px}.task-row{position:relative;width:100%;display:flex;align-items:center;gap:8px;padding:11px 7px;border:0;border-radius:6px;color:var(--text-muted);background:transparent;text-align:left}.task-row:hover,.task-row.selected{background:var(--surface-hover);color:var(--text-strong)}.task-info{min-width:0;flex:1}.task-info b,.task-info small{display:block;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.task-info b{font-size:13px;font-weight:600}.task-info small{margin-top:3px;color:var(--text-subtle);font-size:11px}.status-dot{width:7px;height:7px;border-radius:50%;background:var(--text-subtle)}.status-dot.running{background:var(--accent);box-shadow:0 0 0 4px var(--accent-soft)}.status-dot.completed{background:var(--info)}.status-dot.failed{background:var(--danger)}.delete{opacity:0;color:var(--danger)}.task-row:hover .delete{opacity:1}.empty-list{padding:18px 8px;color:var(--text-subtle);font-size:12px;line-height:1.7}.canvas{min-height:0;display:grid;grid-template-rows:auto minmax(0,1fr) auto;background:var(--surface-base)}.canvas-head{display:flex;justify-content:space-between;gap:20px;padding:17px 30px;border-bottom:1px solid var(--border-subtle)}.canvas-kicker{font:500 10px 'DM Mono',monospace;color:var(--accent);letter-spacing:.08em}.canvas-head h2{max-width:690px;margin:4px 0 0;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;color:var(--text-strong);font:700 18px 'Noto Serif SC',serif}.canvas-actions{display:flex;align-items:center;gap:8px}.outline-button,.download,.feedback-trigger{display:flex;align-items:center;gap:5px;padding:7px 9px;border:1px solid var(--border-strong);border-radius:5px;color:var(--text-muted);background:var(--surface-raised);font-size:12px}.outline-button:hover,.feedback-trigger:hover{color:var(--text-strong);background:var(--surface-hover)}.download{color:var(--accent-on);background:var(--accent);border-color:var(--accent)}.messages{min-height:0;overflow:auto;padding:27px max(25px,7%)}.blank-canvas{max-width:570px;margin:8vh auto;text-align:center}.blank-canvas>span{display:grid;place-items:center;width:53px;height:53px;margin:auto;color:var(--accent-strong);background:var(--surface-accent);border-radius:16px 16px 4px 16px}.blank-canvas h3{margin:19px 0 9px;color:var(--text-strong);font:700 28px 'Noto Serif SC',serif}.blank-canvas p{margin:auto;max-width:450px;color:var(--text-muted);font-size:14px;line-height:1.8}.blank-canvas div{display:flex;justify-content:center;gap:8px;margin-top:25px}.blank-canvas button{border:1px solid var(--border-strong);border-radius:5px;padding:8px 10px;color:var(--text-muted);background:var(--surface-raised);font-size:12px}.message,.observable-phase,.execution-event{max-width:760px;margin:0 auto 16px}.message{padding:18px 20px;border-radius:3px 13px 13px 13px}.message.user{max-width:650px;margin-right:0;color:var(--text-on-accent);background:var(--message-user)}.message.assistant{color:var(--text-strong);background:var(--message-assistant);border:1px solid var(--border-subtle)}.message-label{margin-bottom:10px;font:500 10px 'DM Mono',monospace;letter-spacing:.08em;color:var(--accent-strong)}.message.user .message-label{color:var(--message-user-label)}.observable-phase{padding:13px 14px 12px;border:1px solid var(--border-subtle);border-left:2px solid var(--accent);border-radius:8px;background:var(--surface-accent)}.observable-phase.success{border-left-color:var(--info)}.observable-phase.error{border-left-color:var(--danger)}.phase-head{display:flex;align-items:flex-start;gap:9px}.phase-head>div{display:grid;gap:2px}.phase-kicker,.tool-copy small{color:var(--text-subtle);font:500 10px 'DM Mono',monospace;letter-spacing:.06em}.phase-head b{color:var(--text-strong);font-size:13px}.phase-head small:not(.phase-kicker){color:var(--text-muted);font-size:12px}.trace-state{display:grid;flex:none;place-items:center;margin-top:1px;color:var(--accent)}.trace-state.error{color:var(--danger)}.trace-state.success{color:var(--info)}.trace-state svg{animation:spin 1s linear infinite}.tool-list{display:grid;gap:6px;margin:11px 0 0 23px}.tool-invocation{overflow:hidden;border:1px solid var(--border-subtle);border-radius:6px;background:var(--surface-raised)}.tool-toggle{display:grid;grid-template-columns:auto minmax(0,1fr) auto auto;align-items:center;width:100%;gap:8px;padding:8px 9px;border:0;color:var(--text-muted);background:transparent;text-align:left}.tool-toggle:hover{background:var(--surface-hover)}.tool-copy{display:grid;gap:1px;min-width:0}.tool-copy b{overflow:hidden;color:var(--text-strong);font-size:12px;text-overflow:ellipsis;white-space:nowrap}.tool-status{color:var(--text-subtle);font-size:11px}.tool-detail{display:grid;gap:6px;padding:9px 10px;border-top:1px solid var(--border-subtle);color:var(--text-muted);background:rgba(0,0,0,.08);font-size:12px;line-height:1.55}.tool-detail p{margin:0}.tool-detail span{display:block;margin-bottom:2px;color:var(--text-subtle);font:500 10px 'DM Mono',monospace;letter-spacing:.04em}.execution-event{display:flex;align-items:center;gap:7px;padding:3px 3px;color:var(--text-muted);font-size:12px}.execution-event b{font-weight:600}.execution-event em{overflow:hidden;text-overflow:ellipsis;white-space:nowrap;color:var(--text-subtle);font-style:normal}.feedback-trigger{margin:0 auto 22px;color:var(--accent);font-weight:600}.composer{padding:12px 30px 18px;border-top:1px solid var(--border-subtle);background:var(--surface-base)}.modebar{display:flex;align-items:center;gap:5px;margin:0 0 7px}.modebar button,.modebar label{display:flex;align-items:center;gap:4px;border:0;padding:5px 7px;color:var(--text-subtle);background:transparent;font-size:11px}.modebar button.on{color:var(--accent-strong);background:var(--surface-accent);border-radius:4px}.modebar label{margin-left:5px}.modebar input{accent-color:var(--accent)}.composer-input{display:grid;grid-template-columns:1fr 43px;gap:8px;padding:8px;background:var(--message-assistant);border:1px solid var(--border-subtle);border-radius:8px}.composer textarea{width:100%;min-height:39px;resize:none;padding:6px 8px;border:0;outline:0;color:var(--text-strong);background:transparent;line-height:1.5}.composer-input button{display:grid;place-items:center;align-self:end;height:39px;border:0;border-radius:5px;color:var(--accent-on);background:var(--accent)}.composer-input button:disabled{opacity:.4;cursor:not-allowed}.inline-error{margin:7px 0 0;color:var(--danger);font-size:12px}.delete-confirm{display:flex;justify-content:flex-end;gap:9px}.danger-button{border:1px solid var(--danger);border-radius:5px;padding:7px 10px;color:#411515;background:var(--danger);font-size:12px;font-weight:700}.danger-button:hover{filter:brightness(1.08)}@keyframes spin{to{transform:rotate(360deg)}}@media(max-width:850px){.workbench{grid-template-columns:1fr}.conversations{display:none}.canvas-head,.composer{padding-left:18px;padding-right:18px}.messages{padding:22px 18px}.message{padding:15px}.tool-list{margin-left:9px}.modebar label{display:none}}
.tool-image-preview{display:flex;gap:8px;overflow:auto;padding-bottom:2px}.tool-image-preview a{display:grid;gap:3px;min-width:136px;color:var(--text-muted);font-size:10px;line-height:1.35}.tool-image-preview img{display:block;width:136px;height:88px;border:1px solid var(--border-subtle);border-radius:4px;object-fit:cover;background:var(--surface-hover)}.tool-image-preview a:focus-visible{outline:2px solid var(--focus-ring);outline-offset:2px}
</style>
