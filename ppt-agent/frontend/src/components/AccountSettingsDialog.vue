<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { Check, KeyRound, LoaderCircle, Save, Trash2, X } from 'lucide-vue-next';
import {
  deleteUserApiKey,
  fetchUserApiKeyStatus,
  updateUserApiKey,
  type UserApiKeyStatus,
} from '../api';

const props = defineProps<{ open: boolean }>();
const emit = defineEmits<{ close: [] }>();

const status = ref<UserApiKeyStatus | null>(null);
const draft = ref('');
const loading = ref(false);
const saving = ref(false);
const saved = ref(false);
const error = ref('');

const effectiveSource = computed(() => {
  if (status.value?.configured) return '当前账户 API Key';
  if (status.value?.default_configured) return '系统默认 API Key';
  return '未配置 API Key';
});

const effectiveHint = computed(() => {
  if (status.value?.configured) return '任务会优先使用当前账户的 Key。';
  if (status.value?.default_configured) return '当前账户未单独配置，将使用服务端环境变量中的默认 Key。';
  return '当前没有可用的模型 Key，生成任务会在模型初始化阶段失败。';
});

function formatTime(value?: string): string {
  if (!value) return '';
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? '' : date.toLocaleString('zh-CN');
}

async function loadStatus() {
  loading.value = true;
  error.value = '';
  saved.value = false;
  try {
    status.value = await fetchUserApiKeyStatus();
  } catch (e) {
    status.value = null;
    error.value = e instanceof Error ? e.message : 'API Key 状态加载失败';
  } finally {
    loading.value = false;
  }
}

async function save() {
  const apiKey = draft.value.trim();
  if (apiKey.length < 12) {
    error.value = '请输入有效的 API Key';
    return;
  }

  saving.value = true;
  error.value = '';
  saved.value = false;
  try {
    status.value = await updateUserApiKey(apiKey, status.value?.provider || 'ark');
    draft.value = '';
    saved.value = true;
    window.setTimeout(() => { saved.value = false; }, 2200);
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'API Key 保存失败';
  } finally {
    saving.value = false;
  }
}

async function clear() {
  if (!window.confirm('删除账户 API Key 后，将回退到系统默认 Key。继续吗？')) return;
  saving.value = true;
  error.value = '';
  try {
    await deleteUserApiKey();
    await loadStatus();
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'API Key 删除失败';
  } finally {
    saving.value = false;
  }
}

watch(() => props.open, (open) => {
  if (open) loadStatus();
});
</script>

<template>
  <Teleport to="body">
    <div v-if="open" class="account-overlay" @click.self="emit('close')">
      <section class="account-modal" role="dialog" aria-modal="true" aria-labelledby="account-settings-title">
        <header class="account-modal-head">
          <div class="account-title">
            <span class="account-icon" aria-hidden="true"><KeyRound :size="18" /></span>
            <span>
              <small>账户设置</small>
              <h2 id="account-settings-title">模型 API Key</h2>
            </span>
          </div>
          <button class="account-close" type="button" aria-label="关闭账户设置" @click="emit('close')">
            <X :size="19" />
          </button>
        </header>

        <div v-if="loading" class="account-loading" role="status">
          <LoaderCircle :size="20" class="spin" />
          正在读取账户配置
        </div>

        <div v-else class="account-body">
          <p class="account-intro">
            API Key 只作用于当前账户。完整密钥不会回显，系统仅保存脱敏状态供你确认。
          </p>

          <div class="account-status">
            <div class="status-heading">
              <span>当前生效来源</span>
              <strong :class="{ configured: status?.configured, fallback: !status?.configured && status?.default_configured }">
                {{ effectiveSource }}
              </strong>
            </div>
            <p>{{ effectiveHint }}</p>
            <div class="status-details">
              <span v-if="status?.configured">账户 Key：<code>{{ status.masked_key }}</code></span>
              <span v-else>账户 Key：未配置</span>
              <span v-if="status?.default_configured">系统默认 Key：可用</span>
              <span v-else>系统默认 Key：不可用</span>
              <span v-if="status?.updated_at">更新于 {{ formatTime(status.updated_at) }}</span>
            </div>
          </div>

          <label class="account-field">
            <span>设置账户 Key</span>
            <input
              v-model="draft"
              type="password"
              autocomplete="new-password"
              placeholder="粘贴新的模型 API Key"
              :disabled="saving"
            />
          </label>

          <p v-if="error" class="account-error" role="alert">{{ error }}</p>

          <footer class="account-actions">
            <button v-if="status?.configured" class="account-danger" type="button" :disabled="saving" @click="clear">
              <Trash2 :size="15" />
              删除账户 Key
            </button>
            <span class="account-actions-spacer"></span>
            <button class="account-secondary" type="button" :disabled="saving" @click="emit('close')">关闭</button>
            <button class="account-primary" type="button" :disabled="saving || !draft.trim()" @click="save">
              <LoaderCircle v-if="saving" :size="15" class="spin" />
              <Check v-else-if="saved" :size="15" />
              <Save v-else :size="15" />
              {{ saving ? '保存中' : saved ? '已保存' : '保存 Key' }}
            </button>
          </footer>
        </div>
      </section>
    </div>
  </Teleport>
</template>

<style scoped>
.account-overlay {
  position: fixed;
  inset: 0;
  z-index: var(--z-modal);
  padding: 20px;
  display: grid;
  place-items: center;
  background: rgba(38, 48, 56, 0.22);
  backdrop-filter: blur(3px);
}

.account-modal {
  width: min(560px, 100%);
  max-height: min(680px, calc(100dvh - 40px));
  overflow: auto;
  border: 1px solid var(--border);
  border-radius: 10px;
  background: var(--surface);
  box-shadow: var(--shadow-lg);
}

.account-modal-head {
  min-height: 72px;
  padding: 13px 16px 13px 20px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid var(--divider);
}

.account-title { display: flex; align-items: center; gap: 10px; }
.account-title > span:last-child { display: flex; flex-direction: column; }
.account-title small { color: var(--text-muted); font-size: 10px; font-weight: 750; }
.account-title h2 { margin: 2px 0 0; color: var(--text); font-size: 16px; font-weight: 760; }
.account-icon { width: 34px; height: 34px; display: grid; place-items: center; border: 1px solid #cfe2e9; border-radius: 8px; color: var(--action-ink); background: var(--action-soft); }
.account-close { width: 40px; height: 40px; display: grid; place-items: center; border: 0; border-radius: 6px; color: var(--text-secondary); background: transparent; cursor: pointer; }
.account-close:hover { color: var(--text); background: var(--surface-muted); }

.account-loading { min-height: 260px; display: flex; align-items: center; justify-content: center; gap: 8px; color: var(--text-muted); }
.account-body { padding: 20px; display: grid; gap: 16px; }
.account-intro { margin: 0; color: var(--text-secondary); font-size: 12px; line-height: 1.7; }

.account-status { padding: 14px; display: grid; gap: 7px; border: 1px solid var(--border); border-radius: 8px; background: var(--surface-muted); }
.status-heading { display: flex; align-items: center; justify-content: space-between; gap: 12px; color: var(--text-muted); font-size: 11px; font-weight: 700; }
.status-heading strong { color: var(--danger); font-size: 12px; }.status-heading strong.configured { color: var(--success); }.status-heading strong.fallback { color: var(--info); }
.account-status p { margin: 0; color: var(--text-secondary); font-size: 11px; line-height: 1.6; }
.status-details { display: flex; flex-wrap: wrap; gap: 6px 14px; color: var(--text-muted); font-size: 10px; }
.status-details code { color: var(--text-secondary); font-family: ui-monospace, monospace; }

.account-field { display: grid; gap: 7px; color: var(--text-secondary); font-size: 11px; font-weight: 750; }
.account-field input { width: 100%; min-height: 42px; padding: 0 11px; border: 1px solid var(--border-strong); border-radius: 6px; outline: 0; color: var(--text); background: var(--surface); font-size: 13px; }
.account-field input:focus { border-color: var(--info); box-shadow: 0 0 0 3px var(--info-soft); }
.account-error { margin: 0; color: var(--danger); font-size: 11px; }

.account-actions { padding-top: 15px; display: flex; align-items: center; gap: 8px; border-top: 1px solid var(--divider); }
.account-actions-spacer { flex: 1; }
.account-actions button { min-height: 38px; padding: 0 11px; display: inline-flex; align-items: center; justify-content: center; gap: 7px; border: 1px solid var(--border-strong); border-radius: 6px; font-size: 11px; font-weight: 750; cursor: pointer; }
.account-danger { color: var(--danger); background: var(--danger-soft); }.account-danger:hover { border-color: #e3b5b5; }
.account-secondary { color: var(--text-secondary); background: var(--surface); }.account-secondary:hover { background: var(--surface-muted); }
.account-primary { border-color: var(--action-ink) !important; color: #fff; background: var(--action-ink); }.account-primary:hover:not(:disabled) { background: #2a5869; }

.spin { animation: spin 0.9s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

@media (max-width: 560px) {
  .account-overlay { padding: 0; align-items: end; }
  .account-modal { width: 100%; max-height: 92dvh; border-radius: 10px 10px 0 0; }
  .account-body { padding: 16px; }
  .account-actions { flex-wrap: wrap; }
  .account-danger { order: 3; }.account-actions-spacer { display: none; }
  .account-secondary, .account-primary { flex: 1; }
}
</style>
