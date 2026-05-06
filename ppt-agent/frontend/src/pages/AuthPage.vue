<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { authState } from '../stores/auth';

const router = useRouter();
const route = useRoute();
const auth = authState;

const mode = ref<'code' | 'password'>('code');
const email = ref('');
const code = ref('');
const password = ref('');
const codeSent = ref(false);
const codeCountdown = ref(0);
const localError = ref('');
let countdownTimer: ReturnType<typeof setInterval> | null = null;

onMounted(() => {
  auth.clearError();
});

async function handleSendCode() {
  localError.value = '';
  auth.clearError();
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

function switchMode(m: 'code' | 'password') {
  mode.value = m;
  localError.value = '';
  codeSent.value = false;
  if (countdownTimer) { clearInterval(countdownTimer); countdownTimer = null; }
}
</script>

<template>
  <div class="auth-page">
    <!-- Animated background -->
    <div class="auth-bg">
      <div class="bg-orb orb-1"></div>
      <div class="bg-orb orb-2"></div>
      <div class="bg-orb orb-3"></div>
    </div>

    <div class="auth-card">
      <div class="auth-logo" @click="router.push('/')">
        <svg viewBox="0 0 48 48" fill="none">
          <rect x="5" y="7" width="38" height="28" rx="4" fill="rgba(99,102,241,0.15)" stroke="rgba(99,102,241,0.3)" stroke-width="1.5"/>
          <rect x="5" y="7" width="38" height="7" rx="4" fill="rgba(99,102,241,0.3)"/>
          <rect x="13" y="18" width="16" height="2.5" rx="1.25" fill="rgba(99,102,241,0.2)"/>
          <rect x="13" y="23" width="22" height="2.5" rx="1.25" fill="rgba(99,102,241,0.12)"/>
          <rect x="13" y="28" width="18" height="2.5" rx="1.25" fill="rgba(99,102,241,0.1)"/>
        </svg>
      </div>
      <h2 class="auth-title">欢迎使用 PPT Agent</h2>
      <p class="auth-sub">AI 驱动的智能幻灯片生成平台</p>

      <form @submit.prevent="handleLogin">
        <input
          class="auth-input"
          v-model="email"
          type="email"
          placeholder="邮箱地址"
          autocomplete="email"
        />

        <template v-if="mode === 'code'">
          <div class="code-row">
            <input
              class="auth-input code-input"
              v-model="code"
              placeholder="验证码"
              maxlength="6"
            />
            <button
              type="button"
              class="send-code-btn"
              :disabled="auth.loading || codeCountdown > 0"
              @click="handleSendCode"
            >
              {{ codeCountdown > 0 ? codeCountdown + 's' : codeSent ? '重新发送' : '发送验证码' }}
            </button>
          </div>
        </template>

        <template v-else>
          <input
            class="auth-input"
            v-model="password"
            type="password"
            placeholder="密码"
            autocomplete="current-password"
          />
        </template>

        <p v-if="localError" class="auth-error">{{ localError }}</p>

        <button class="auth-btn" type="submit" :disabled="auth.loading">
          <span v-if="auth.loading" class="btn-spinner"></span>
          <span>{{ auth.loading ? '请稍候...' : mode === 'code' ? '登录 / 注册' : '登录' }}</span>
        </button>
      </form>

      <p class="auth-switch">
        <a href="#" @click.prevent="switchMode(mode === 'code' ? 'password' : 'code')">
          {{ mode === 'code' ? '密码登录' : '验证码登录' }}
        </a>
        <span class="switch-sep">|</span>
        <a href="#" @click.prevent="router.push('/')">返回首页</a>
      </p>
    </div>
  </div>
</template>

<style scoped>
.auth-page {
  min-height: 100vh;
  display: flex; align-items: center; justify-content: center;
  background: #0b1121;
  position: relative; overflow: hidden;
}

.auth-bg { position: absolute; inset: 0; pointer-events: none; }
.bg-orb {
  position: absolute; border-radius: 50%;
  filter: blur(80px); opacity: 0.15;
}
.orb-1 {
  width: 400px; height: 400px;
  background: #6366f1;
  top: -100px; right: -100px;
  animation: float 8s ease-in-out infinite;
}
.orb-2 {
  width: 300px; height: 300px;
  background: #3b82f6;
  bottom: -50px; left: -80px;
  animation: float 10s ease-in-out infinite reverse;
}
.orb-3 {
  width: 250px; height: 250px;
  background: #8b5cf6;
  top: 50%; left: 50%;
  animation: float 12s ease-in-out infinite 2s;
}

.auth-card {
  background: rgba(255,255,255,0.03);
  backdrop-filter: blur(24px);
  border: 1px solid rgba(255,255,255,0.08);
  border-radius: 20px;
  padding: 2.5rem 2rem;
  width: 100%; max-width: 400px;
  text-align: center;
  z-index: 1;
  animation: fadeInUp 0.6s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 0 20px 60px rgba(0,0,0,0.3);
}

.auth-logo {
  margin-bottom: 1rem; cursor: pointer;
  display: inline-block;
  transition: transform var(--transition);
}
.auth-logo:hover { transform: scale(1.05); }
.auth-logo svg { width: 56px; height: 56px; }

.auth-title {
  font-size: 1.3rem; font-weight: 700;
  color: #f1f5f9; margin-bottom: 0.3rem;
}
.auth-sub {
  font-size: 0.8rem; color: #64748b;
  margin-bottom: 1.75rem;
}

.auth-input {
  width: 100%; padding: 0.75rem 0.9rem;
  margin-bottom: 0.65rem;
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 10px;
  background: rgba(255,255,255,0.04);
  color: #e2e8f0; font-size: 0.85rem;
  outline: none; font-family: inherit;
  transition: border-color var(--transition), box-shadow var(--transition);
}
.auth-input::placeholder { color: #475569; }
.auth-input:focus {
  border-color: rgba(99,102,241,0.5);
  box-shadow: 0 0 0 3px rgba(99,102,241,0.1);
}

.code-row { display: flex; gap: 0.5rem; }
.code-input { flex: 1; }

.send-code-btn {
  white-space: nowrap; padding: 0.75rem 0.9rem;
  border: 1px solid rgba(99,102,241,0.4);
  border-radius: 10px;
  background: rgba(99,102,241,0.1);
  color: #818cf8; font-size: 0.8rem; font-weight: 500;
  cursor: pointer; transition: all var(--transition);
}
.send-code-btn:hover { background: rgba(99,102,241,0.2); border-color: #818cf8; }
.send-code-btn:disabled { opacity: 0.4; cursor: not-allowed; }

.auth-error {
  color: #f87171; font-size: 0.75rem;
  margin-bottom: 0.5rem; text-align: left;
}

.auth-btn {
  width: 100%; padding: 0.75rem;
  border: none; border-radius: 10px;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  color: #fff; font-size: 0.9rem; font-weight: 600;
  cursor: pointer; margin-top: 0.5rem;
  display: flex; align-items: center; justify-content: center; gap: 0.5rem;
  transition: transform var(--transition), box-shadow var(--transition);
}
.auth-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 20px rgba(99,102,241,0.4);
}
.auth-btn:disabled { opacity: 0.6; cursor: not-allowed; transform: none; }

.btn-spinner {
  width: 18px; height: 18px;
  border: 2px solid rgba(255,255,255,0.3);
  border-top-color: #fff; border-radius: 50%;
  animation: spin 0.7s linear infinite;
}

.auth-switch {
  font-size: 0.75rem; color: #64748b;
  margin-top: 1.25rem; display: flex; justify-content: center; gap: 0.5rem;
}
.auth-switch a {
  color: #818cf8; text-decoration: none; font-weight: 500;
  transition: color var(--transition);
}
.auth-switch a:hover { color: #a5b4fc; }
.switch-sep { color: #334155; }

@keyframes float {
  0%, 100% { transform: translateY(0) scale(1); }
  50% { transform: translateY(-30px) scale(1.05); }
}
@keyframes fadeInUp {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}
@keyframes spin { to { transform: rotate(360deg); } }
</style>
