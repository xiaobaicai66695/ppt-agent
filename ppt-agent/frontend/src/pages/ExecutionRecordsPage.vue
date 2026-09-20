<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft, ChevronLeft, ChevronRight, Database, RefreshCw, Search, ShieldAlert, X } from 'lucide-vue-next'
import AppShell from '../components/AppShell.vue'
import { fetchAdminExecutionRecords, fetchMe } from '../api'
import type { AdminExecutionRecord } from '../api'
import type { AuthUser } from '../types'

const router = useRouter()
const route = useRoute()
const user = ref<AuthUser>()
const records = ref<AdminExecutionRecord[]>([])
const total = ref(0)
const loading = ref(true)
const isFetching = ref(false)
const refreshing = ref(false)
const error = ref('')
const lastUpdated = ref('')
const userIDInput = ref(readUserID())
let latestRequestVersion = 0

const page = computed(() => positiveQueryNumber('page', 1))
const pageSize = computed(() => {
  const value = positiveQueryNumber('page_size', 50)
  return [50, 100, 200].includes(value) ? value : 50
})
const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))
const rangeLabel = computed(() => {
  if (!total.value) return '暂无记录'
  const first = (page.value - 1) * pageSize.value + 1
  return `第 ${first}–${Math.min(total.value, page.value * pageSize.value)} 条，共 ${total.value} 条`
})

function positiveQueryNumber(key: string, fallback: number) {
  const raw = Array.isArray(route.query[key]) ? route.query[key][0] : route.query[key]
  const parsed = Number(raw)
  return Number.isInteger(parsed) && parsed > 0 ? parsed : fallback
}

function readUserID() {
  const value = Array.isArray(route.query.user_id) ? route.query.user_id[0] : route.query.user_id
  return typeof value === 'string' ? value : ''
}

function updateQuery(nextPage = 1, nextPageSize = pageSize.value, userID = userIDInput.value.trim()) {
  const query: Record<string, string> = { page: String(nextPage), page_size: String(nextPageSize) }
  if (userID) query.user_id = userID
  router.replace({ query })
}

function applyFilter() {
  const value = userIDInput.value.trim()
  if (value && (!/^\d+$/.test(value) || !Number.isSafeInteger(Number(value)) || Number(value) < 1)) {
    error.value = '用户 ID 必须是大于 0 的整数。'
    return
  }
  updateQuery(1, pageSize.value, value)
}

function clearFilter() {
  userIDInput.value = ''
  updateQuery(1, pageSize.value, '')
}

function changePageSize(event: Event) {
  const nextPageSize = Number((event.target as HTMLSelectElement).value)
  updateQuery(1, nextPageSize)
}

async function loadRecords() {
  if (!user.value?.is_admin) return
  const requestVersion = ++latestRequestVersion
  loading.value = !records.value.length
  isFetching.value = true
  error.value = ''
  try {
    const userID = readUserID()
    userIDInput.value = userID
    const data = await fetchAdminExecutionRecords({ page: page.value, pageSize: pageSize.value, userId: userID ? Number(userID) : undefined })
    const nextRecords = data.records || []
    const nextTotal = data.pagination?.total || 0
    const nextPage = data.pagination?.page || page.value
    if (!nextRecords.length && nextTotal > 0 && nextPage > Math.ceil(nextTotal / pageSize.value)) {
      if (requestVersion === latestRequestVersion) updateQuery(Math.ceil(nextTotal / pageSize.value), pageSize.value, userID)
      return
    }
    if (requestVersion !== latestRequestVersion) return
    records.value = nextRecords
    total.value = nextTotal
    lastUpdated.value = new Date().toISOString()
  } catch (cause) {
    if (requestVersion === latestRequestVersion) error.value = cause instanceof Error ? cause.message : '无法加载执行记录。'
  } finally {
    if (requestVersion === latestRequestVersion) {
      loading.value = false
      isFetching.value = false
    }
  }
}

async function refresh() {
  refreshing.value = true
  await loadRecords()
  refreshing.value = false
}

function statusLabel(status: string) {
  return ({ completed: '已完成', failed: '失败', cancelled: '已取消', running: '执行中', conversation: '对话' }[status] || status)
}

function formatDate(value?: string) {
  return value ? new Date(value).toLocaleString('zh-CN', { month: 'numeric', day: 'numeric', hour: '2-digit', minute: '2-digit' }) : '—'
}

function formatDuration(milliseconds?: number) {
  if (!milliseconds || milliseconds <= 0) return '—'
  const seconds = Math.round(milliseconds / 1000)
  return seconds >= 60 ? `${Math.floor(seconds / 60)}分${seconds % 60}秒` : `${seconds}秒`
}

onMounted(async () => {
  document.title = '执行记录 · PPTform'
  try {
    user.value = await fetchMe()
    if (!user.value.is_admin) {
      error.value = '当前账户没有查看执行记录的权限。'
      return
    }
    await loadRecords()
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : '无法确认账户权限。'
  } finally {
    loading.value = false
  }
})

watch(() => route.fullPath, () => {
  if (user.value?.is_admin) void loadRecords()
})
</script>

<template>
  <AppShell title="执行记录" subtitle="按用户审阅任务执行与资源消耗" :email="user?.email" :guest="user?.is_guest" @new="router.push('/dashboard')">
    <template #header>
      <button class="header-action" @click="router.push('/admin')"><ArrowLeft :size="15" />返回运营后台</button>
      <button class="header-action refresh" :disabled="isFetching || refreshing" @click="refresh"><RefreshCw :size="15" :class="{ spinning: refreshing }" />更新</button>
    </template>

    <main class="execution-records">
      <p v-if="error && !records.length" class="page-error" role="alert"><ShieldAlert :size="17" />{{ error }}<button v-if="user?.is_admin" @click="refresh">重试</button></p>
      <template v-else>
        <section class="intro">
          <div><p class="eyebrow">PPTform / execution ledger</p><h2>执行记录</h2><p>查看每次请求的状态、耗时和资源消耗；数据按创建时间倒序排列。</p></div>
          <p class="scope-note"><Database :size="15" />{{ rangeLabel }}</p>
        </section>

        <section class="records-panel" aria-labelledby="records-heading">
          <header class="records-header">
            <div><p class="panel-kicker">可核验运行数据</p><h3 id="records-heading">全部用户任务</h3></div>
            <p v-if="lastUpdated">更新于 {{ formatDate(lastUpdated) }}</p>
          </header>

          <form class="filters" novalidate @submit.prevent="applyFilter">
            <label for="execution-user-id">用户 ID</label>
            <div class="filter-input"><input id="execution-user-id" v-model="userIDInput" inputmode="numeric" autocomplete="off" placeholder="全部用户" aria-describedby="execution-filter-help" /> <button v-if="userIDInput" type="button" aria-label="清除用户 ID 筛选" @click="clearFilter"><X :size="14" /></button></div>
            <button class="filter-submit" type="submit"><Search :size="15" />筛选</button>
            <span id="execution-filter-help">按用户 ID 精确过滤</span>
            <label class="page-size-label" for="execution-page-size">每页</label>
            <select id="execution-page-size" :value="pageSize" aria-label="每页记录数" @change="changePageSize"><option :value="50">50 条</option><option :value="100">100 条</option><option :value="200">200 条</option></select>
          </form>

          <p v-if="error" class="inline-error" role="alert"><ShieldAlert :size="15" />{{ error }}<button @click="refresh">重试</button></p>
          <div class="table-wrap" :aria-busy="isFetching">
            <table>
              <thead><tr><th scope="col">用户 / 请求</th><th scope="col">类型</th><th scope="col">状态</th><th scope="col">进度</th><th scope="col">Token</th><th scope="col">执行耗时</th><th scope="col">创建时间</th><th scope="col">链路</th></tr></thead>
              <tbody v-if="records.length">
                <tr v-for="record in records" :key="record.id">
                  <td class="request-cell"><b>{{ record.query || '未命名请求' }}</b><small>{{ record.user_email || `用户 #${record.user_id}` }} · #{{ record.user_id }}</small><p v-if="record.error" class="task-error">{{ record.error }}</p></td>
                  <td><span class="intent">{{ record.intent || '—' }}</span></td>
                  <td><span class="status" :class="record.status">{{ statusLabel(record.status) }}</span></td>
                  <td>{{ record.total_count ? `${record.done_count}/${record.total_count} 页` : '—' }}</td>
                  <td><b class="mono">{{ record.total_tokens.toLocaleString('zh-CN') }}</b><small>输入 {{ record.prompt_tokens.toLocaleString('zh-CN') }} / 输出 {{ record.completion_tokens.toLocaleString('zh-CN') }}</small></td>
                  <td>{{ formatDuration(record.generation_duration_ms) }}</td>
                  <td>{{ formatDate(record.created_at) }}</td>
                  <td><details><summary>查看</summary><dl><div><dt>任务</dt><dd>{{ record.id }}</dd></div><div v-if="record.conversation_id"><dt>会话</dt><dd>{{ record.conversation_id }}</dd></div><div v-if="record.parent_task_id"><dt>父任务</dt><dd>{{ record.parent_task_id }}</dd></div><div v-if="record.fixer_run_count"><dt>Fixer</dt><dd>{{ record.fixer_run_count }} 次</dd></div></dl></details></td>
                </tr>
              </tbody>
              <tbody v-else-if="loading"><tr><td class="state-cell" colspan="8"><RefreshCw :size="18" class="spinning" />正在读取执行记录…</td></tr></tbody>
              <tbody v-else><tr><td class="state-cell" colspan="8"><Database :size="19" />{{ readUserID() ? '该用户暂无执行记录。' : '暂无执行记录。' }}</td></tr></tbody>
            </table>
          </div>

          <footer class="pagination" aria-label="执行记录分页"><span>{{ rangeLabel }}</span><div><button :disabled="isFetching || page <= 1" aria-label="上一页" @click="updateQuery(page - 1)"><ChevronLeft :size="17" />上一页</button><strong>第 {{ page }} / {{ totalPages }} 页</strong><button :disabled="isFetching || page >= totalPages" aria-label="下一页" @click="updateQuery(page + 1)">下一页<ChevronRight :size="17" /></button></div></footer>
        </section>
      </template>
    </main>
  </AppShell>
</template>

<style scoped>
.execution-records{--ink:#e8f2f5;--muted:#91a8b5;--faint:#65808c;--line:rgba(189,219,225,.14);--surface:#0d202d;--surface-raised:#102735;--mint:#75e4c9;--blue:#83c5f7;--red:#f29b97;--amber:#efbd69;max-width:1440px;width:100%;height:calc(100vh - 78px);overflow:auto;margin:0 auto;padding:32px 34px 70px;color:var(--ink)}
.header-action{display:inline-flex;align-items:center;gap:6px;border:1px solid transparent;padding:8px 10px;border-radius:7px;color:var(--text-muted);background:transparent;font-size:12px;font-weight:600;cursor:pointer}.header-action:hover{color:var(--text-strong);background:var(--surface-hover)}.header-action:focus-visible,.filters input:focus-visible,.filters select:focus-visible,.filters button:focus-visible,.pagination button:focus-visible,.page-error button:focus-visible,.inline-error button:focus-visible{outline:2px solid var(--accent);outline-offset:2px}.header-action:disabled,.pagination button:disabled{opacity:.48;cursor:not-allowed}.header-action.refresh{margin-left:2px}.spinning{animation:spin 1s linear infinite}
.intro{display:flex;align-items:flex-end;justify-content:space-between;gap:24px}.eyebrow,.panel-kicker{margin:0;color:var(--blue);font:10px 'DM Mono',monospace;letter-spacing:.13em;text-transform:uppercase}.intro h2{margin:6px 0;color:var(--ink);font:700 33px 'Noto Serif SC',serif;letter-spacing:-.05em}.intro p:not(.eyebrow){margin:0;color:var(--muted);font-size:13px}.scope-note{display:flex;align-items:center;gap:7px;flex:none;padding:8px 10px;border:1px solid var(--line);border-radius:6px;color:var(--muted)!important;background:var(--surface);font:11px 'DM Mono',monospace!important}
.records-panel{margin-top:22px;border:1px solid var(--line);border-radius:11px;background:var(--surface);overflow:hidden}.records-header{display:flex;justify-content:space-between;gap:18px;padding:18px 20px 16px}.records-header h3{margin:5px 0 0;color:var(--ink);font:700 19px 'Noto Serif SC',serif;letter-spacing:-.03em}.records-header>p{margin:4px 0 0;color:var(--faint);font-size:11px}
.filters{display:flex;align-items:center;gap:9px;padding:12px 20px;border-top:1px solid var(--line);border-bottom:1px solid var(--line);background:var(--surface-raised);font-size:12px}.filters>label{color:var(--muted);white-space:nowrap}.filter-input{display:flex;align-items:center;min-width:190px;border:1px solid var(--line);border-radius:6px;background:var(--surface)}.filter-input input{width:100%;min-width:0;border:0;padding:8px 10px;color:var(--ink);background:transparent;outline:0;font:12px 'DM Mono',monospace}.filter-input input::placeholder{color:var(--faint)}.filter-input button{display:grid;place-items:center;flex:none;width:29px;height:29px;border:0;color:var(--muted);background:transparent;cursor:pointer}.filter-input button:hover{color:var(--ink)}.filter-submit,.pagination button,.page-error button,.inline-error button{display:inline-flex;align-items:center;justify-content:center;gap:5px;border:1px solid var(--line);border-radius:6px;padding:8px 10px;color:var(--ink);background:var(--surface);font-size:12px;font-weight:600;cursor:pointer}.filter-submit{color:#10323b;border-color:transparent;background:var(--mint)}.filter-submit:hover{background:#a0f0db}.filters>span{color:var(--faint);font-size:11px}.page-size-label{margin-left:auto}.filters select{border:1px solid var(--line);border-radius:6px;padding:7px 28px 7px 9px;color:var(--ink);background:var(--surface);font:12px 'DM Mono',monospace}
.table-wrap{min-height:390px;overflow:auto;scrollbar-gutter:stable}.table-wrap[aria-busy='true']{opacity:.75}table{width:100%;min-width:1120px;border-collapse:collapse;text-align:left;font-size:12px}th{padding:11px 14px;color:var(--faint);background:color-mix(in srgb,var(--surface) 90%,transparent);font:10px 'DM Mono',monospace;letter-spacing:.08em;text-transform:uppercase;white-space:nowrap}td{padding:13px 14px;vertical-align:top;border-top:1px solid rgba(189,219,225,.09);color:var(--muted)}tbody tr:hover{background:rgba(117,228,201,.035)}.request-cell{min-width:250px;max-width:350px}.request-cell b,.request-cell small{display:block;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.request-cell b{color:var(--ink);font-weight:600}.request-cell small,td>small{margin-top:4px;color:var(--faint);font-size:10px}.task-error{display:-webkit-box;overflow:hidden;margin:6px 0 0;color:var(--red);font-size:11px;line-height:1.45;-webkit-box-orient:vertical;-webkit-line-clamp:2}.intent{color:var(--blue);font:11px 'DM Mono',monospace}.status{display:inline-flex;align-items:center;min-width:52px;border-radius:99px;padding:4px 7px;font-size:10px;font-weight:700;white-space:nowrap;background:rgba(131,197,247,.11);color:var(--blue)}.status.completed{background:rgba(117,228,201,.13);color:var(--mint)}.status.failed{background:rgba(242,155,151,.13);color:var(--red)}.status.cancelled{background:rgba(239,189,105,.13);color:var(--amber)}.mono{color:var(--ink);font:600 12px 'DM Mono',monospace}details{min-width:138px}summary{color:var(--blue);cursor:pointer;font-size:11px;white-space:nowrap}summary:focus-visible{outline:2px solid var(--accent);outline-offset:3px}dl{display:grid;gap:6px;margin:9px 0 0;padding:8px;border-left:2px solid var(--line);background:var(--surface-raised);font-size:10px}dl div{display:grid;gap:2px}dt{color:var(--faint)}dd{max-width:170px;overflow:hidden;margin:0;color:var(--muted);font-family:'DM Mono',monospace;text-overflow:ellipsis;white-space:nowrap}.state-cell{height:300px;text-align:center;color:var(--muted)}.state-cell svg{vertical-align:middle;margin-right:7px}.inline-error,.page-error{display:flex;align-items:center;gap:8px;margin:14px 20px;padding:10px 12px;border:1px solid rgba(242,155,151,.32);border-radius:7px;color:#ffc4c0;background:rgba(105,34,47,.5);font-size:12px}.page-error{margin:0}.page-error button,.inline-error button{margin-left:auto;color:#ffc4c0;background:transparent}.pagination{display:flex;align-items:center;justify-content:space-between;gap:16px;padding:13px 20px;border-top:1px solid var(--line);color:var(--faint);font-size:11px}.pagination>div{display:flex;align-items:center;gap:8px}.pagination strong{min-width:78px;color:var(--muted);font:10px 'DM Mono',monospace;text-align:center}.pagination button:hover:not(:disabled){border-color:var(--blue);color:var(--blue)}
@keyframes spin{to{transform:rotate(360deg)}}@media(prefers-reduced-motion:reduce){.spinning{animation:none}}@media(max-width:850px){.execution-records{padding:24px 20px 54px}.intro{align-items:flex-start;flex-direction:column;gap:12px}.filters{flex-wrap:wrap}.filters>span{order:4;flex-basis:100%}.page-size-label{margin-left:0}.records-header{align-items:flex-start;flex-direction:column;gap:3px}.table-wrap{min-height:360px}}@media(max-width:620px){.execution-records{height:calc(100vh - 64px);padding:20px 14px 45px}.intro h2{font-size:28px}.filters{align-items:flex-end}.filter-input{min-width:0;flex:1 1 155px}.filter-submit{flex:0 0 auto}.page-size-label{margin-left:auto}.pagination{align-items:flex-start;flex-direction:column}.pagination>div{width:100%;justify-content:space-between}.header-action{padding:7px}.header-action.refresh{margin-left:0}}:global(html[data-theme='light']) .execution-records{--ink:#153742;--muted:#58727a;--faint:#769097;--line:rgba(33,81,91,.16);--surface:#f8fbf8;--surface-raised:#eef5f2;--mint:#148f78;--blue:#24719b;--red:#be554e;--amber:#b27b1e}.execution-records:deep(.studio-header){color:var(--muted)}
</style>
