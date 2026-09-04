<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { Activity, ArrowUpRight, CheckCircle2, Clock3, KeyRound, MessageSquareText, RefreshCw, ShieldAlert, Star, Users, Wrench, XCircle } from 'lucide-vue-next'
import AppShell from '../components/AppShell.vue'
import { fetchAdminFeedback, fetchAdminStats, fetchAdminTasks, fetchAdminUsers, fetchMe } from '../api'
import type { AdminFeedback, AdminStats, AdminTask, AdminUser } from '../api'
import type { AuthUser } from '../types'

const router = useRouter()
const user = ref<AuthUser>()
const error = ref('')
const loading = ref(true)
const refreshing = ref(false)
const stats = ref<AdminStats>({ user_count: 0, task_count: 0, running_count: 0, registered_user_count: 0, non_root_registered_user_count: 0, ppt_active_user_count: 0, custom_api_key_user_count: 0, ppt_generation_count: 0, non_root_ppt_generation_count: 0, feedback_count: 0, feedback_suggestion_count: 0 })
const users = ref<AdminUser[]>([])
const tasks = ref<AdminTask[]>([])
const feedbackItems = ref<AdminFeedback[]>([])
const lastUpdated = ref(new Date().toISOString())

const pptTasks = computed(() => tasks.value.filter(task => Boolean(task.generation_started_at)))
const terminalTasks = computed(() => pptTasks.value.filter(task => ['completed', 'failed', 'cancelled'].includes(task.status)))
const successfulTasks = computed(() => pptTasks.value.filter(task => task.status === 'completed'))
const averageDuration = (source: AdminTask[]) => source.length ? source.reduce((sum, task) => sum + (task.generation_duration_ms || 0), 0) / source.length / 1000 : undefined
const generation = computed(() => ({ total: pptTasks.value.length, success: successfulTasks.value.length, failed: pptTasks.value.filter(task => task.status === 'failed').length, cancelled: pptTasks.value.filter(task => task.status === 'cancelled').length, success_rate: terminalTasks.value.length ? successfulTasks.value.length / terminalTasks.value.length : undefined, avg_duration_seconds: averageDuration(terminalTasks.value), avg_success_duration_seconds: averageDuration(successfulTasks.value), non_system_total: stats.value.non_root_ppt_generation_count }))
const userMetrics = computed(() => ({ registered_total: stats.value.registered_user_count, registered_new: users.value.filter(account => Date.now() - new Date(account.created_at).getTime() <= 30 * 24 * 60 * 60 * 1000).length, active: { mau: stats.value.ppt_active_user_count }, personal_key_users: stats.value.custom_api_key_user_count }))
const averageRating = computed(() => feedbackItems.value.length ? feedbackItems.value.reduce((sum, item) => sum + item.rating, 0) / feedbackItems.value.length : undefined)
const feedback = computed(() => ({ total: stats.value.feedback_count, average_rating: averageRating.value, recent: feedbackItems.value }))
const fixer = computed(() => ({ total_runs: pptTasks.value.reduce((sum, task) => sum + (task.fixer_run_count || 0), 0), touched_task_count: pptTasks.value.filter(task => task.fixer_run_count > 0).length }))
const topGenerators = computed(() => [...users.value].sort((left, right) => right.ppt_generation_count - left.ppt_generation_count).slice(0, 6))
const taskTotal = computed(() => generation.value.total)
const successRate = computed(() => generation.value.success_rate)
const statusLabel = (status: string) => ({ completed: '已完成', failed: '失败', cancelled: '已取消', running: '生成中', conversation: '对话' }[status] || status)
const formatNumber = (value?: number) => new Intl.NumberFormat('zh-CN').format(value || 0)
const formatPercent = (value?: number) => value === undefined || value === null ? '—' : `${(value <= 1 ? value * 100 : value).toFixed(1)}%`
const formatDuration = (seconds?: number) => {
  if (seconds === undefined || seconds === null || seconds <= 0) return '—'
  const rounded = Math.round(seconds)
  return rounded >= 60 ? `${Math.floor(rounded / 60)}分${rounded % 60}秒` : `${rounded}秒`
}
const formatDate = (value?: string) => value ? new Date(value).toLocaleString('zh-CN', { month: 'numeric', day: 'numeric', hour: '2-digit', minute: '2-digit' }) : '—'
const width = (value: number) => `${Math.max(0, Math.min(100, value))}%`
const countShare = (value: number) => taskTotal.value ? (value / taskTotal.value) * 100 : 0

async function load() {
  loading.value = true
  error.value = ''
  try {
    user.value = await fetchMe()
    if (!user.value.is_admin) { error.value = '当前账户没有运营后台权限。'; return }
    const [baseStats, baseUsers, baseTasks, baseFeedback] = await Promise.all([fetchAdminStats(), fetchAdminUsers(), fetchAdminTasks(), fetchAdminFeedback()])
    stats.value = baseStats; users.value = baseUsers; tasks.value = baseTasks; feedbackItems.value = baseFeedback; lastUpdated.value = new Date().toISOString()
  } catch (cause) { error.value = cause instanceof Error ? cause.message : '无法加载后台' }
  finally { loading.value = false }
}
async function refresh() { refreshing.value = true; await load(); refreshing.value = false }
onMounted(load)
</script>

<template>
  <AppShell title="运营后台" subtitle="生成质量、采用情况与交付反馈" :email="user?.email" :guest="user?.is_guest" @new="router.push('/dashboard')">
    <template #header>
      <button class="refresh" :disabled="loading || refreshing" @click="refresh"><RefreshCw :size="15" :class="{ spinning: refreshing }" />更新数据</button>
    </template>
    <main class="admin">
      <p v-if="error" class="error"><ShieldAlert :size="17" />{{ error }}</p>
      <template v-else>
        <section class="intro">
          <div><p class="eyebrow">Deckform / operating signal</p><h2>运营脉搏</h2><p>从一次生成到用户复用，查看交付体验是否真正改善。</p></div>
          <p class="scope-note">实时总量 + 最近 100 条任务记录</p>
        </section>

        <section class="signal-strip" aria-label="核心运营指标">
          <article class="primary-signal"><span class="metric-icon"><CheckCircle2 :size="18" /></span><div><small>近期任务成功率</small><strong>{{ formatPercent(successRate) }}</strong><p>{{ formatNumber(generation.success) }} 个完成交付 / {{ formatNumber(taskTotal) }} 条可计时任务</p></div></article>
          <article><Clock3 :size="18" /><div><small>成功任务平均时长</small><strong>{{ formatDuration(generation.avg_success_duration_seconds) }}</strong><p>仅基于最近 100 条任务</p></div></article>
          <article><Users :size="18" /><div><small>累计生成用户</small><strong>{{ formatNumber(userMetrics.active.mau) }}</strong><p>访客按同一 IP 去重</p></div></article>
          <article><Activity :size="18" /><div><small>当前生成中</small><strong>{{ formatNumber(stats.running_count) }}</strong><p>实时运行负载</p></div></article>
        </section>

        <p class="sync-note">更新于 {{ formatDate(lastUpdated) }}。任务耗时和 Fixer 汇总来自可返回的最近 100 条任务；账户、Key、生成和反馈总量来自后端全量统计。</p>
          <section class="dashboard-grid generation-grid">
            <article class="panel generation-health">
              <header><div><p class="panel-kicker">交付健康度</p><h3>生成状态</h3></div><span>{{ formatNumber(taskTotal) }} 次创建</span></header>
              <div class="status-bar" aria-label="生成任务状态分布"><i class="complete" :style="{ width: width(countShare(generation.success)) }"></i><i class="failed" :style="{ width: width(countShare(generation.failed)) }"></i><i class="cancelled" :style="{ width: width(countShare(generation.cancelled)) }"></i></div>
              <div class="status-key"><span><i class="dot complete"></i>已完成 <b>{{ formatNumber(generation.success) }}</b></span><span><i class="dot failed"></i>失败 <b>{{ formatNumber(generation.failed) }}</b></span><span><i class="dot cancelled"></i>取消 <b>{{ formatNumber(generation.cancelled) }}</b></span></div>
              <div class="duration-compare"><div><small>全部终态任务</small><b>{{ formatDuration(generation.avg_duration_seconds) }}</b><p>从开始生成到完成、失败或取消</p></div><div><small>仅成功任务</small><b>{{ formatDuration(generation.avg_success_duration_seconds) }}</b><p>排除失败任务的实际交付耗时</p></div></div>
            </article>
            <article class="panel adoption-panel"><header><div><p class="panel-kicker">用户采用</p><h3>注册与使用</h3></div><ArrowUpRight :size="18" /></header><div class="registration"><div><small>累计注册</small><b>{{ formatNumber(userMetrics.registered_total) }}</b><span>近 30 天新增 {{ formatNumber(userMetrics.registered_new) }}</span></div><div class="active-window"><span><small>创建过 PPT</small><b>{{ formatNumber(userMetrics.active.mau) }}</b></span><span><small>非 root 注册</small><b>{{ formatNumber(stats.non_root_registered_user_count) }}</b></span></div></div><div class="adoption-foot"><span>访客同 IP 合并统计</span><span>历史总生成 <b>{{ formatNumber(stats.ppt_generation_count) }}</b></span></div></article>
          </section>

          <section class="dashboard-grid operations-grid">
            <article class="panel"><header><div><p class="panel-kicker">个人模型 Key</p><h3>自配 Key 使用</h3></div><KeyRound :size="18" /></header><div class="key-number"><b>{{ formatNumber(userMetrics.personal_key_users) }}</b><span>位用户已启用个人 Key</span></div><p class="quiet-copy">后端仅返回配置状态，不传输、存储或展示 Key 内容与供应商明细。</p></article>
            <article class="panel"><header><div><p class="panel-kicker">质量闭环</p><h3>用户反馈</h3></div><MessageSquareText :size="18" /></header><div class="feedback-summary"><div><Star :size="17" /><b>{{ feedback.average_rating?.toFixed(1) || '—' }}</b><span>/ 5 平均评分</span></div><span>{{ formatNumber(feedback.total) }} 条已提交</span></div><p class="quiet-copy">{{ formatNumber(stats.feedback_suggestion_count) }} 条附文字建议。反馈类型尚未由后端记录，因此不对内容、排版等分类作推断。</p></article>
            <article class="panel fixer-panel"><header><div><p class="panel-kicker">自动修复</p><h3>Fixer 执行</h3></div><Wrench :size="18" /></header><div class="fixer-summary"><div><b>{{ formatNumber(fixer.total_runs) }}</b><span>次实际触发</span></div><div><b>{{ formatNumber(fixer.touched_task_count) }}</b><span>个任务触发过</span></div></div><p class="quiet-copy">按 task_id 累加，当前接口未返回每次触发原因和修复成功结果。</p></article>
          </section>

          <section class="dashboard-grid detail-grid">
            <article class="panel ranking"><header><div><p class="panel-kicker">使用分布</p><h3>生成用户</h3></div><span>按创建次数</span></header><div v-if="topGenerators.length" class="ranking-list"><div v-for="(entry, index) in topGenerators" :key="String(entry.id)"><span class="rank">{{ String(index + 1).padStart(2, '0') }}</span><b>{{ entry.email || `用户 #${entry.id}` }}</b><span>{{ entry.ppt_generation_count || 0 }} 次创建</span><em>{{ entry.custom_api_key_configured ? '自配 Key' : '默认模型' }}</em></div></div><p v-else class="quiet-copy">单用户创建次数由后端全量统计，便于识别高频使用者。</p></article>
            <article class="panel recent-feedback"><header><div><p class="panel-kicker">最近声音</p><h3>反馈与建议</h3></div><span>{{ formatNumber(feedback.total) }} 条</span></header><div v-if="feedback.recent.length" class="feedback-list"><div v-for="item in feedback.recent.slice(0, 4)" :key="item.task_id"><span class="rating"><Star :size="13" />{{ item.rating }}</span><div><b>{{ item.task_query || item.task_id }}</b><p>{{ item.suggestion || '未填写文字建议' }}</p><small>{{ formatDate(item.updated_at || item.created_at) }}</small></div></div></div><p v-else class="quiet-copy">完成 PPT 后提交的评分与建议会在这里出现。</p></article>
          </section>

        <section class="panel records"><header><div><p class="panel-kicker">实时记录</p><h3>最近任务</h3></div><span>{{ tasks.length }} 条</span></header><div class="rows"><div v-for="task in tasks.slice(0, 12)" :key="task.id" class="row"><span class="task-title"><b>{{ task.query || '未命名任务' }}</b><small>{{ task.user_email || '未知账户' }} · {{ formatDate(task.created_at) }}</small></span><em :class="task.status"><CheckCircle2 v-if="task.status === 'completed'" :size="14" /><XCircle v-else-if="task.status === 'failed'" :size="14" /><Activity v-else :size="14" />{{ statusLabel(task.status) }}</em><span>{{ task.total_count ? `${task.done_count}/${task.total_count} 页` : '对话' }}</span></div></div></section>
        <section class="panel records accounts"><header><div><p class="panel-kicker">账户概况</p><h3>最近注册账户</h3></div><span>{{ users.length }} 个</span></header><div class="rows"><div v-for="account in users.slice(0, 10)" :key="account.id" class="row"><span class="task-title"><b>{{ account.email }}</b><small>注册于 {{ formatDate(account.created_at) }}</small></span><em :class="{ admin: account.is_admin }">{{ account.is_admin ? '管理员' : '成员' }}</em><span>#{{ account.id }}</span></div></div></section>
      </template>
    </main>
  </AppShell>
</template>

<style scoped>
.admin{--ink:#e8f2f5;--muted:#91a8b5;--faint:#65808c;--line:rgba(189,219,225,.14);--surface:#0d202d;--surface-raised:#102735;--mint:#75e4c9;--blue:#83c5f7;--red:#f29b97;--amber:#efbd69;max-width:1320px;width:100%;height:calc(100vh - 78px);overflow:auto;margin:0 auto;padding:32px 34px 70px;color:var(--ink)}
.intro{display:flex;justify-content:space-between;gap:24px;align-items:end;margin:3px 0 21px}.eyebrow,.panel-kicker{margin:0;color:var(--mint);font:500 10px 'DM Mono',monospace;letter-spacing:.08em;text-transform:uppercase}.intro h2{margin:5px 0 2px;font:800 32px 'Noto Serif SC',serif;letter-spacing:-.05em}.intro>div>p:not(.eyebrow),.scope-note{margin:0;color:var(--muted);font-size:13px}.scope-note{padding:7px 10px;border:1px solid var(--line);border-radius:7px;font-size:11px}.refresh{border:0;color:var(--muted);background:transparent;font-size:12px;display:flex;align-items:center;gap:7px;margin-left:auto;padding:8px 10px;border:1px solid var(--line);border-radius:7px}.refresh:hover{color:var(--ink);background:rgba(255,255,255,.04)}.refresh:disabled{opacity:.55;cursor:wait}.spinning{animation:spin 1s linear infinite}
.signal-strip{display:grid;grid-template-columns:1.2fr repeat(3,1fr);border:1px solid var(--line);border-radius:12px;overflow:hidden;background:var(--surface)}.signal-strip article{min-width:0;display:flex;gap:11px;padding:18px 19px;border-left:1px solid var(--line);color:var(--blue)}.signal-strip article:first-child{border-left:0}.signal-strip .primary-signal{background:linear-gradient(100deg,rgba(106,226,198,.14),transparent 80%);color:var(--mint)}.metric-icon{display:grid;place-items:center;width:28px;height:28px;border-radius:8px;background:rgba(117,228,201,.14)}.signal-strip small,.signal-strip p{display:block;margin:0;color:var(--muted);font-size:11px}.signal-strip strong{display:block;margin:3px 0 2px;color:var(--ink);font:700 24px 'Noto Serif SC',serif;letter-spacing:-.04em}.signal-strip p{white-space:nowrap;overflow:hidden;text-overflow:ellipsis}.analytics-waiting{display:flex;gap:11px;align-items:flex-start;margin-top:20px;padding:15px 17px;border:1px dashed rgba(131,197,247,.35);border-radius:10px;color:var(--blue);background:rgba(65,127,169,.08)}.analytics-waiting b{font-size:13px}.analytics-waiting p{margin:4px 0 0;color:var(--muted);font-size:12px}.sync-note{margin:15px 0 0;color:var(--faint);font-size:11px}.sync-note span{margin:0 7px}.dashboard-grid{display:grid;gap:15px;margin-top:15px}.generation-grid{grid-template-columns:1.25fr .75fr}.operations-grid{grid-template-columns:repeat(3,1fr)}.detail-grid{grid-template-columns:.9fr 1.1fr}.panel{min-width:0;border:1px solid var(--line);border-radius:11px;background:var(--surface);overflow:hidden}.panel>header{display:flex;align-items:flex-start;justify-content:space-between;gap:10px;padding:17px 18px 14px;color:var(--blue)}.panel h3{margin:4px 0 0;color:var(--ink);font:700 18px 'Noto Serif SC',serif;letter-spacing:-.03em}.panel>header>span{color:var(--muted);font-size:11px}.status-bar{display:flex;height:9px;margin:4px 18px 11px;border-radius:99px;overflow:hidden;background:rgba(255,255,255,.06)}.status-bar i{display:block;height:100%;transition:width .3s ease}.complete,.status-bar .complete{background:var(--mint)}.failed,.status-bar .failed{background:var(--red)}.cancelled,.status-bar .cancelled{background:var(--amber)}.status-key{display:flex;flex-wrap:wrap;gap:14px;padding:0 18px 18px;color:var(--muted);font-size:11px}.status-key span{display:flex;align-items:center;gap:5px}.status-key b{color:var(--ink);font-family:'DM Mono',monospace}.dot{display:inline-block;width:6px;height:6px;border-radius:50%}.duration-compare{display:grid;grid-template-columns:1fr 1fr;border-top:1px solid var(--line)}.duration-compare>div{padding:14px 18px}.duration-compare>div+div{border-left:1px solid var(--line)}.duration-compare small,.duration-compare p{display:block;margin:0;color:var(--muted);font-size:11px}.duration-compare b{display:block;margin:5px 0 2px;color:var(--ink);font:700 18px 'Noto Serif SC',serif}.registration{padding:2px 18px 15px}.registration>div:first-child{display:flex;align-items:baseline;gap:9px}.registration small,.active-window small{color:var(--muted);font-size:11px}.registration>div:first-child b{color:var(--ink);font:700 29px 'Noto Serif SC',serif;letter-spacing:-.05em}.registration>div:first-child span{color:var(--mint);font-size:11px}.active-window{display:flex;justify-content:space-between;margin-top:14px;padding:10px 0 0;border-top:1px solid var(--line)}.active-window span{display:grid;gap:2px}.active-window b{color:var(--ink);font:600 15px 'DM Mono',monospace}.adoption-foot{display:flex;justify-content:space-between;gap:10px;padding:12px 18px;border-top:1px solid var(--line);color:var(--muted);font-size:11px}.adoption-foot b{color:var(--ink);font-family:'DM Mono',monospace}.key-number{display:flex;align-items:baseline;gap:8px;padding:3px 18px 15px}.key-number b{font:700 31px 'Noto Serif SC',serif;letter-spacing:-.05em}.key-number span,.quiet-copy{color:var(--muted);font-size:11px}.provider-list{display:flex;flex-wrap:wrap;gap:7px;padding:13px 18px;border-top:1px solid var(--line)}.provider-list span{display:flex;align-items:center;gap:5px;padding:4px 7px;border-radius:5px;color:var(--muted);background:rgba(255,255,255,.04);font:10px 'DM Mono',monospace}.provider-list b{color:var(--ink);font-weight:500}.provider-list i{width:4px;height:4px;border-radius:50%;background:var(--mint)}.quiet-copy{min-height:58px;margin:0;padding:14px 18px;border-top:1px solid var(--line);line-height:1.6}.feedback-summary{display:flex;justify-content:space-between;align-items:center;padding:1px 18px 12px}.feedback-summary>div{display:flex;align-items:center;gap:5px;color:var(--amber)}.feedback-summary b{color:var(--ink);font:700 27px 'Noto Serif SC',serif}.feedback-summary span{color:var(--muted);font-size:11px}.feedback-types{display:grid;gap:7px;padding:11px 18px;border-top:1px solid var(--line)}.feedback-types span{display:grid;grid-template-columns:35px minmax(0,1fr) 20px;gap:7px;align-items:center;color:var(--muted);font-size:10px}.feedback-types b{font-weight:500}.feedback-types i{height:5px;border-radius:99px;background:rgba(255,255,255,.07);overflow:hidden}.feedback-types em{display:block;height:100%;border-radius:99px;background:var(--amber)}.feedback-types small{text-align:right;font:10px 'DM Mono',monospace}.fixer-summary{display:grid;grid-template-columns:1fr 1fr;margin:2px 18px 10px}.fixer-summary>div+div{padding-left:16px;border-left:1px solid var(--line)}.fixer-summary b{display:block;color:var(--ink);font:700 22px 'Noto Serif SC',serif}.fixer-summary span{color:var(--muted);font-size:11px}.reason-list{display:grid;gap:7px;padding:11px 18px;border-top:1px solid var(--line)}.reason-list span{display:flex;justify-content:space-between;gap:9px;font-size:11px}.reason-list b{overflow:hidden;color:var(--ink);text-overflow:ellipsis;white-space:nowrap;font-weight:500}.reason-list small{flex:none;color:var(--muted);font-size:10px}.ranking-list,.feedback-list{border-top:1px solid var(--line)}.ranking-list>div{display:grid;grid-template-columns:26px minmax(0,1fr) 72px 54px;gap:8px;align-items:center;padding:10px 18px;border-bottom:1px solid rgba(189,219,225,.08);font-size:11px}.ranking-list>div:last-child{border:0}.ranking-list .rank{color:var(--faint);font:10px 'DM Mono',monospace}.ranking-list b{overflow:hidden;text-overflow:ellipsis;white-space:nowrap;font-weight:500}.ranking-list span:not(.rank),.ranking-list em{color:var(--muted);font-style:normal;text-align:right}.feedback-list>div{display:flex;gap:10px;padding:11px 18px;border-bottom:1px solid rgba(189,219,225,.08)}.feedback-list>div:last-child{border:0}.rating{display:flex;align-items:center;gap:2px;height:21px;padding:0 6px;color:#2c2717;background:var(--amber);border-radius:5px;font:11px 'DM Mono',monospace}.feedback-list b,.feedback-list p,.feedback-list small{display:block;margin:0;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.feedback-list b{font-size:12px;font-weight:600}.feedback-list p{max-width:480px;margin:3px 0;color:var(--muted);font-size:11px}.feedback-list small{color:var(--faint);font-size:10px}.records{margin-top:15px}.accounts{margin-top:15px}.rows{border-top:1px solid var(--line)}.row{display:grid;grid-template-columns:minmax(0,1fr) 100px 70px;gap:16px;align-items:center;padding:12px 18px;border-bottom:1px solid rgba(189,219,225,.08);color:var(--muted);font-size:11px}.row:last-child{border:0}.task-title b,.task-title small{display:block;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.task-title b{color:var(--ink);font-size:12px;font-weight:600}.task-title small{margin-top:3px;color:var(--faint);font-size:10px}.row em{display:flex;align-items:center;gap:4px;font-style:normal;color:var(--blue)}.row em.completed{color:var(--mint)}.row em.failed{color:var(--red)}.row em.cancelled{color:var(--amber)}.row em.admin{color:var(--mint)}.error{display:flex;gap:8px;align-items:center;padding:14px;color:#ffc4c0;background:#3c2029;border:1px solid rgba(242,155,151,.25);border-radius:9px}@keyframes spin{to{transform:rotate(360deg)}}
@media(max-width:1050px){.signal-strip{grid-template-columns:repeat(2,1fr)}.signal-strip article:nth-child(3){border-left:0;border-top:1px solid var(--line)}.signal-strip article:nth-child(4){border-top:1px solid var(--line)}.operations-grid{grid-template-columns:1fr 1fr}.operations-grid>.panel:last-child{grid-column:span 2}.detail-grid{grid-template-columns:1fr}.admin{padding:26px 24px 60px}}
@media(max-width:670px){.admin{height:calc(100vh - 64px);padding:21px 15px 45px}.intro{align-items:flex-start;flex-direction:column;gap:15px}.intro h2{font-size:28px}.signal-strip,.generation-grid,.operations-grid{grid-template-columns:1fr}.signal-strip article{border-left:0;border-top:1px solid var(--line)}.signal-strip article:first-child{border-top:0}.operations-grid>.panel:last-child{grid-column:auto}.duration-compare{grid-template-columns:1fr}.duration-compare>div+div{border-top:1px solid var(--line);border-left:0}.row{grid-template-columns:minmax(0,1fr) 86px}.row>span:last-child{display:none}.ranking-list>div{grid-template-columns:22px minmax(0,1fr) 50px}.ranking-list em{display:none}.feedback-list p{max-width:250px}.sync-note{line-height:1.6}.signal-strip p{white-space:normal}.refresh{margin-left:auto}}
:global(html[data-theme='light']) .admin{--ink:#153742;--muted:#58727a;--faint:#769097;--line:rgba(33,81,91,.16);--surface:#f8fbf8;--surface-raised:#eef5f2;--mint:#148f78;--blue:#24719b;--red:#be554e;--amber:#b27b1e}.admin:deep(.refresh){color:var(--muted)}:global(html[data-theme='light']) .signal-strip .primary-signal{background:linear-gradient(100deg,rgba(20,143,120,.11),transparent 80%)}
</style>
