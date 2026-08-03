<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import {
  ArrowLeft,
  CheckCircle2,
  Eye,
  EyeOff,
  LockKeyhole,
  Mail,
  ShieldCheck,
  Sparkles,
} from 'lucide-vue-next';
import { authState } from '../stores/auth';

const router = useRouter();
const route = useRoute();
const auth = authState;

const mode = ref<'code' | 'password'>('code');
const email = ref('');
const code = ref('');
const password = ref('');
const showPassword = ref(false);
const codeSent = ref(false);
const codeCountdown = ref(0);
const localError = ref('');
let countdownTimer: ReturnType<typeof setInterval> | null = null;

onMounted(() => auth.clearError());
onUnmounted(() => {
  if (countdownTimer) clearInterval(countdownTimer);
});

async function handleSendCode() {
  localError.value = '';
  auth.clearError();
  if (!email.value.trim()) {
    localError.value = '请先填写邮箱地址';
    return;
  }
  try {
    await auth.sendCode(email.value);
    codeSent.value = true;
    codeCountdown.value = 60;
    countdownTimer = setInterval(() => {
      codeCountdown.value--;
      if (codeCountdown.value <= 0 && countdownTimer) {
        clearInterval(countdownTimer);
        countdownTimer = null;
      }
    }, 1000);
  } catch (e) {
    localError.value = (e as Error).message;
  }
}

async function handleLogin() {
  localError.value = '';
  auth.clearError();

  if (!email.value.trim()) { localError.value = '请输入邮箱地址'; return; }
  if (mode.value === 'password' && !password.value) { localError.value = '请输入密码'; return; }
  if (mode.value === 'code' && !code.value.trim()) { localError.value = '请输入验证码'; return; }

  try {
    const credential = mode.value === 'code' ? code.value : password.value;
    await auth.login(email.value, credential, mode.value);
    code.value = '';
    password.value = '';
    codeSent.value = false;
    if (countdownTimer) { clearInterval(countdownTimer); countdownTimer = null; }
    const redirect = (route.query.redirect as string) || '/dashboard';
    router.push(redirect);
  } catch {
    localError.value = auth.error;
  }
}

function switchMode(next: 'code' | 'password') {
  mode.value = next;
  localError.value = '';
  codeSent.value = false;
  if (countdownTimer) { clearInterval(countdownTimer); countdownTimer = null; }
}
</script>

<template>
  <main id="main-content" class="auth-page">
    <section class="auth-form-pane" aria-labelledby="auth-title">
      <button class="brand-link" type="button" @click="router.push('/')">
        <span class="brand-mark"><Sparkles :size="18" /></span>
        <span>PPT Agent</span>
      </button>

      <div class="auth-form-wrap">
        <div class="auth-heading">
          <span class="auth-kicker">个人工作区</span>
          <h1 id="auth-title">登录并继续你的演示</h1>
          <p>任务进度、模板偏好和生成结果会保留在你的工作区。</p>
        </div>

        <div class="auth-tabs" role="tablist" aria-label="登录方式">
          <button
            type="button"
            role="tab"
            :aria-selected="mode === 'code'"
            :class="{ active: mode === 'code' }"
            @click="switchMode('code')"
          >验证码</button>
          <button
            type="button"
            role="tab"
            :aria-selected="mode === 'password'"
            :class="{ active: mode === 'password' }"
            @click="switchMode('password')"
          >密码</button>
        </div>

        <form class="auth-form" novalidate @submit.prevent="handleLogin">
          <div class="field-group">
            <label for="auth-email">邮箱地址</label>
            <div class="field-control">
              <Mail :size="18" aria-hidden="true" />
              <input
                id="auth-email"
                v-model="email"
                type="email"
                placeholder="name@company.com"
                autocomplete="email"
                :aria-describedby="localError ? 'auth-error-msg' : undefined"
              />
            </div>
          </div>

          <div v-if="mode === 'code'" class="field-group">
            <label for="auth-code">验证码</label>
            <div class="code-field">
              <div class="field-control">
                <ShieldCheck :size="18" aria-hidden="true" />
                <input
                  id="auth-code"
                  v-model="code"
                  inputmode="numeric"
                  placeholder="6 位验证码"
                  maxlength="6"
                  autocomplete="one-time-code"
                />
              </div>
              <button
                type="button"
                class="send-code-button"
                :disabled="auth.loading || codeCountdown > 0"
                @click="handleSendCode"
              >
                {{ codeCountdown > 0 ? `${codeCountdown}s` : codeSent ? '重新发送' : '发送验证码' }}
              </button>
            </div>
            <span v-if="codeCountdown > 0" class="visually-hidden" aria-live="polite">{{ codeCountdown }} 秒后可重新发送</span>
          </div>

          <div v-else class="field-group">
            <label for="auth-password">密码</label>
            <div class="field-control password-control">
              <LockKeyhole :size="18" aria-hidden="true" />
              <input
                id="auth-password"
                v-model="password"
                :type="showPassword ? 'text' : 'password'"
                placeholder="输入登录密码"
                autocomplete="current-password"
              />
              <button
                type="button"
                :title="showPassword ? '隐藏密码' : '显示密码'"
                :aria-label="showPassword ? '隐藏密码' : '显示密码'"
                @click="showPassword = !showPassword"
              >
                <EyeOff v-if="showPassword" :size="18" />
                <Eye v-else :size="18" />
              </button>
            </div>
          </div>

          <p v-if="localError" id="auth-error-msg" class="auth-error" role="alert" aria-live="assertive">
            {{ localError }}
          </p>

          <button class="submit-button" type="submit" :disabled="auth.loading">
            <span v-if="auth.loading" class="button-spinner" aria-hidden="true" />
            <span>{{ auth.loading ? '正在登录' : mode === 'code' ? '登录或注册' : '登录工作区' }}</span>
          </button>
        </form>

        <button class="back-link" type="button" @click="router.push('/')">
          <ArrowLeft :size="16" />
          返回创建工作区
        </button>
      </div>
    </section>

    <aside class="auth-preview" aria-label="模板预览">
      <div class="preview-copy">
        <span>从结构到成稿</span>
        <h2>你的内容，一页一页抵达。</h2>
        <p>先确认叙事结构，再观察每张幻灯片渐进生成。无需等待整份文件完成。</p>
      </div>
      <div class="preview-stack" aria-hidden="true">
        <figure class="preview-main">
          <img src="/templates/thumbs/research-report.jpg" alt="" width="640" height="360" />
        </figure>
        <figure class="preview-secondary first">
          <img src="/templates/thumbs/pitch-deck.jpg" alt="" width="640" height="360" />
        </figure>
        <figure class="preview-secondary second">
          <img src="/templates/thumbs/tech-sharing.jpg" alt="" width="640" height="360" />
        </figure>
      </div>
      <div class="preview-facts">
        <span><CheckCircle2 :size="16" /> 真实模板预览</span>
        <span><CheckCircle2 :size="16" /> 单页渐进交付</span>
        <span><CheckCircle2 :size="16" /> 可继续修改</span>
      </div>
    </aside>
  </main>
</template>

<style scoped>
.auth-page {
  min-height: 100dvh;
  display: grid;
  grid-template-columns: minmax(420px, 0.82fr) minmax(520px, 1.18fr);
  background: var(--surface);
}

.auth-form-pane {
  min-width: 0;
  padding: 28px clamp(28px, 5vw, 72px) 42px;
  display: flex;
  flex-direction: column;
}

.brand-link {
  width: fit-content;
  min-height: 44px;
  padding: 0;
  display: inline-flex;
  align-items: center;
  gap: 10px;
  border: 0;
  color: var(--text);
  background: transparent;
  font-weight: 750;
  cursor: pointer;
}
.brand-mark {
  width: 34px;
  height: 34px;
  display: grid;
  place-items: center;
  border-radius: 7px;
  color: #102522;
  background: var(--action);
}

.auth-form-wrap {
  width: min(100%, 430px);
  margin: auto;
  padding: 50px 0 38px;
}

.auth-kicker { color: var(--action-ink); font-size: 11px; font-weight: 750; }
.auth-heading h1 { margin: 10px 0 0; font-size: 30px; line-height: 1.2; letter-spacing: 0; }
.auth-heading p { margin: 12px 0 0; color: var(--text-secondary); font-size: 14px; line-height: 1.7; }

.auth-tabs {
  margin-top: 30px;
  padding: 3px;
  display: grid;
  grid-template-columns: 1fr 1fr;
  border: 1px solid var(--border);
  border-radius: 7px;
  background: var(--surface-muted);
}
.auth-tabs button {
  min-height: 38px;
  border: 0;
  border-radius: 5px;
  color: var(--text-muted);
  background: transparent;
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
}
.auth-tabs button.active { color: var(--text); background: var(--surface); box-shadow: var(--shadow-xs); }

.auth-form { margin-top: 24px; display: grid; gap: 18px; }
.field-group { display: grid; gap: 8px; }
.field-group > label { color: var(--text-secondary); font-size: 11px; font-weight: 700; }
.field-control {
  min-height: 48px;
  padding: 0 13px;
  display: flex;
  align-items: center;
  gap: 10px;
  border: 1px solid var(--border-strong);
  border-radius: 6px;
  color: var(--text-muted);
  background: var(--surface);
  transition: border-color var(--motion-fast), box-shadow var(--motion-fast);
}
.field-control:focus-within { border-color: var(--action-ink); box-shadow: 0 0 0 3px rgba(7,94,87,0.1); }
.field-control input { min-width: 0; width: 100%; border: 0; outline: 0; color: var(--text); background: transparent; font-size: 16px; }
.field-control input::placeholder { color: #959da1; }

.password-control > button {
  width: 44px;
  height: 44px;
  margin-right: -10px;
  display: grid;
  place-items: center;
  flex: 0 0 auto;
  border: 0;
  color: var(--text-muted);
  background: transparent;
  cursor: pointer;
}

.code-field { display: grid; grid-template-columns: minmax(0, 1fr) auto; gap: 9px; }
.send-code-button {
  min-width: 112px;
  min-height: 48px;
  padding: 0 13px;
  border: 1px solid var(--border-strong);
  border-radius: 6px;
  color: var(--action-ink);
  background: var(--surface);
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
}
.send-code-button:hover:not(:disabled) { background: var(--action-soft); border-color: #9bded6; }

.auth-error { margin: -4px 0 0; padding: 10px 12px; border-left: 3px solid var(--danger); color: var(--danger); background: var(--danger-soft); font-size: 12px; }

.submit-button {
  min-height: 48px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 9px;
  border: 1px solid var(--action-ink);
  border-radius: 6px;
  color: #ffffff;
  background: var(--action-ink);
  font-weight: 750;
  cursor: pointer;
  transition: background var(--motion-fast), transform var(--motion-fast);
}
.submit-button:hover:not(:disabled) { background: #064d48; }
.submit-button:active:not(:disabled) { transform: scale(0.99); }
.button-spinner { width: 17px; height: 17px; border: 2px solid rgba(255,255,255,0.4); border-top-color: #ffffff; border-radius: 50%; animation: spin 0.7s linear infinite; }

.back-link {
  min-height: 44px;
  margin: 20px auto 0;
  display: flex;
  align-items: center;
  gap: 7px;
  border: 0;
  color: var(--text-muted);
  background: transparent;
  font-size: 12px;
  cursor: pointer;
}
.back-link:hover { color: var(--text); }

.auth-preview {
  position: relative;
  min-width: 0;
  overflow: hidden;
  padding: clamp(48px, 7vw, 96px);
  display: flex;
  flex-direction: column;
  justify-content: center;
  color: #f4f7f7;
  background: #1b1f21;
}
.preview-copy { position: relative; z-index: 2; width: min(100%, 550px); }
.preview-copy > span { color: var(--action); font-size: 11px; font-weight: 750; }
.preview-copy h2 { margin: 12px 0 0; font-size: clamp(30px, 4vw, 48px); line-height: 1.12; letter-spacing: 0; }
.preview-copy p { max-width: 500px; margin: 16px 0 0; color: #b8c0c4; font-size: 14px; line-height: 1.75; }

.preview-stack { position: relative; width: min(100%, 620px); aspect-ratio: 1.55; margin: 44px 0 28px; }
.preview-stack figure { position: absolute; margin: 0; overflow: hidden; border: 1px solid rgba(255,255,255,0.13); border-radius: 7px; background: #2b3033; box-shadow: 0 24px 60px rgba(0,0,0,0.28); }
.preview-stack img { width: 100%; height: 100%; object-fit: cover; }
.preview-main { inset: 0 12% 10% 0; z-index: 3; }
.preview-secondary { width: 48%; aspect-ratio: 16/9; right: 0; }
.preview-secondary.first { top: 8%; z-index: 2; }
.preview-secondary.second { bottom: 0; z-index: 4; }

.preview-facts { display: flex; flex-wrap: wrap; gap: 18px; color: #b8c0c4; font-size: 11px; }
.preview-facts span { display: inline-flex; align-items: center; gap: 6px; }
.preview-facts svg { color: var(--action); }

@keyframes spin { to { transform: rotate(360deg); } }

@media (max-width: 900px) {
  .auth-page { grid-template-columns: 1fr; }
  .auth-preview { display: none; }
  .auth-form-pane { min-height: 100dvh; }
}

@media (max-width: 520px) {
  .auth-form-pane { padding: 18px 16px 30px; }
  .auth-form-wrap { padding-top: 42px; }
  .auth-heading h1 { font-size: 27px; }
  .code-field { grid-template-columns: 1fr; }
  .send-code-button { width: 100%; }
}
</style>
