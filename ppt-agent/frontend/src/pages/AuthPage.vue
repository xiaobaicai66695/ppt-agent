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

function switchMode(m: 'code' | 'password') {
  mode.value = m;
  localError.value = '';
  codeSent.value = false;
  if (countdownTimer) { clearInterval(countdownTimer); countdownTimer = null; }
}
</script>

<template>
  <div class="auth-page">
    <!-- Subtle background orbs -->
    <div class="auth-bg">
      <div class="bg-orb orb-1"></div>
      <div class="bg-orb orb-2"></div>
      <div class="bg-orb orb-3"></div>
    </div>

    <div class="auth-card">
      <div class="auth-logo" @click="router.push('/')">
        <svg viewBox="0 0 48 48" fill="none">
          <rect x="4" y="6" width="40" height="30" rx="5" fill="var(--accent-soft)" stroke="var(--accent-border)" stroke-width="1.5"/>
          <rect x="4" y="6" width="40" height="8" rx="5" fill="var(--accent)"/>
          <rect x="12" y="18" width="18" height="3" rx="1.5" fill="var(--accent)"/>
          <rect x="12" y="24" width="24" height="2.5" rx="1.25" fill="var(--border)"/>
          <rect x="12" y="29" width="20" height="2.5" rx="1.25" fill="var(--border)"/>
        </svg>
      </div>
      <h2 class="auth-title">欢迎使用 PPT Agent</h2>
      <p class="auth-sub">AI 驱动的智能幻灯片生成平台</p>

      <div class="demo-hint">
        <span class="demo-label">体验账号</span>
        <span class="demo-cred" aria-label="体验账号: root 密码">root@qq.com / root</span>
      </div>

      <form @submit.prevent="handleLogin" novalidate>
        <div class="input-group">
          <label for="auth-email" class="sr-only">邮箱地址</label>
          <input
            id="auth-email"
            class="auth-input"
            v-model="email"
            type="email"
            placeholder="邮箱地址"
            autocomplete="email"
            :aria-describedby="localError ? 'auth-error-msg' : undefined"
          />
        </div>

        <template v-if="mode === 'code'">
          <div class="code-row">
            <label for="auth-code" class="sr-only">验证码</label>
            <input
              id="auth-code"
              class="auth-input code-input"
              v-model="code"
              placeholder="验证码"
              maxlength="6"
              autocomplete="one-time-code"
            />
            <button
              type="button"
              class="send-code-btn"
              :disabled="auth.loading || codeCountdown > 0"
              @click="handleSendCode"
            >
              {{ codeCountdown > 0 ? codeCountdown + 's' : codeSent ? '重新发送' : '发送验证码' }}
            </button>
            <span v-if="codeCountdown > 0" class="sr-only" aria-live="polite">{{ codeCountdown }}秒后可重新发送</span>
          </div>
        </template>

        <template v-else>
          <label for="auth-password" class="sr-only">密码</label>
          <input
            id="auth-password"
            class="auth-input"
            v-model="password"
            type="password"
            placeholder="密码"
            autocomplete="current-password"
          />
        </template>

        <p v-if="localError" id="auth-error-msg" class="auth-error" role="alert" aria-live="assertive">{{ localError }}</p>

        <button class="auth-btn" type="submit" :disabled="auth.loading">
          <span v-if="auth.loading" class="btn-spinner" aria-hidden="true"></span>
          <span>{{ auth.loading ? '请稍候...' : mode === 'code' ? '登录 / 注册' : '登录' }}</span>
        </button>
      </form>

      <p class="auth-switch">
        <button type="button" class="switch-link" @click="switchMode(mode === 'code' ? 'password' : 'code')">
          {{ mode === 'code' ? '密码登录' : '验证码登录' }}
        </button>
        <span class="switch-sep">|</span>
        <button type="button" class="switch-link" @click="router.push('/')">返回首页</button>
      </p>
    </div>
  </div>
</template>

<style scoped>
.auth-page {
  min-height: 100vh;
  display: flex; align-items: center; justify-content: center;
  background: var(--bg-muted);
  position: relative; overflow: hidden;
}

/* Top accent bar */
.auth-page::before {
  content: '';
  position: absolute; top: 0; left: 0; right: 0; height: 4px;
  background: linear-gradient(90deg, var(--accent), #8b5cf6, var(--info));
}

.auth-bg { position: absolute; inset: 0; pointer-events: none; }
.bg-orb { position: absolute; border-radius: 50%; filter: blur(100px); opacity: 0.06; }
.orb-1 { width: 400px; height: 400px; background: var(--accent); top: -100px; right: -80px; animation: float 10s ease-in-out infinite; }
.orb-2 { width: 300px; height: 300px; background: var(--info); bottom: -50px; left: -60px; animation: float 12s ease-in-out infinite reverse; }
.orb-3 { width: 220px; height: 220px; background: #8b5cf6; top: 50%; left: 50%; animation: float 14s ease-in-out infinite 2s; }

.auth-card {
  background: var(--bg-base);
  border: 1px solid var(--border);
  border-radius: var(--radius-xl);
  padding: 2.5rem 2rem;
  width: 100%; max-width: 400px;
  text-align: center;
  z-index: 1;
  animation: fadeInUp 0.5s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: var(--shadow-md);
  position: relative;
}

.auth-logo {
  margin-bottom: 1rem; cursor: pointer;
  display: inline-block;
  transition: transform var(--transition-md);
}
.auth-logo:hover { transform: scale(1.05); }
.auth-logo svg { width: 52px; height: 52px; }

.auth-title { font-size: 1.3rem; font-weight: 700; color: var(--text); margin-bottom: 0.3rem; }
.auth-sub { font-size: 0.8rem; color: var(--text-muted); margin-bottom: 1.5rem; }

.auth-input {
  width: 100%; padding: 0.7rem 0.9rem;
  margin-bottom: 0.65rem;
  border: 1.5px solid var(--border);
  border-radius: var(--radius);
  background: var(--bg-base);
  color: var(--text); font-size: 0.85rem;
  outline: none; font-family: inherit;
  transition: border-color var(--transition), box-shadow var(--transition);
}
.auth-input::placeholder { color: var(--text-disabled); }
.auth-input:focus {
  border-color: var(--accent);
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
}

.code-row { display: flex; gap: 0.5rem; }
.code-input { flex: 1; }

.send-code-btn {
  white-space: nowrap; padding: 0.7rem 0.9rem;
  border: 1.5px solid var(--accent-border);
  border-radius: var(--radius);
  background: var(--accent-soft);
  color: var(--accent); font-size: 0.78rem; font-weight: 600;
  cursor: pointer; transition: all var(--transition);
  flex-shrink: 0;
}
.send-code-btn:hover { background: var(--accent); color: #fff; border-color: var(--accent); }
.send-code-btn:disabled { opacity: 0.45; cursor: not-allowed; }

.auth-error {
  color: var(--danger); font-size: 0.75rem;
  margin-bottom: 0.5rem; text-align: left;
}

.auth-btn {
  width: 100%; padding: 0.75rem;
  border: none; border-radius: var(--radius);
  background: var(--accent); color: #fff; font-size: 0.9rem; font-weight: 600;
  cursor: pointer; margin-top: 0.25rem;
  display: flex; align-items: center; justify-content: center; gap: 0.5rem;
  transition: transform var(--transition-md), box-shadow var(--transition-md), background var(--transition);
  box-shadow: 0 4px 12px rgba(99, 102, 241, 0.25);
}
.auth-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 6px 20px rgba(99, 102, 241, 0.35);
  background: var(--accent-hover);
}
.auth-btn:disabled { opacity: 0.6; cursor: not-allowed; transform: none; box-shadow: none; }

.btn-spinner {
  width: 16px; height: 16px;
  border: 2px solid rgba(255,255,255,0.3);
  border-top-color: #fff; border-radius: 50%;
  animation: spin 0.7s linear infinite;
}

.auth-switch {
  font-size: 0.75rem; color: var(--text-muted);
  margin-top: 1.25rem; display: flex; justify-content: center; gap: 0.5rem;
}
.auth-switch a { color: var(--accent); text-decoration: none; font-weight: 500; transition: color var(--transition); }
.auth-switch a:hover { color: var(--accent-hover); }
.switch-sep { color: var(--border); }
.switch-link {
  background: none; border: none; cursor: pointer;
  color: var(--accent); font-size: inherit; font-weight: 500;
  font-family: inherit; padding: 0;
  transition: color var(--transition);
}
.switch-link:hover { color: var(--accent-hover); }

.sr-only {
  position: absolute; width: 1px; height: 1px;
  padding: 0; margin: -1px; overflow: hidden;
  clip: rect(0,0,0,0); white-space: nowrap; border: 0;
}

.demo-hint {
  display: flex; align-items: center; justify-content: center; gap: 0.6rem;
  margin-bottom: 0.75rem; padding: 0.55rem 0.9rem;
  background: var(--accent-soft);
  border: 1px solid var(--accent-border);
  border-radius: var(--radius); font-size: 0.78rem;
}
.demo-label { color: var(--text-secondary); font-weight: 500; }
.demo-cred {
  color: var(--accent); font-weight: 600;
  font-family: 'JetBrains Mono', 'SF Mono', monospace;
  letter-spacing: 0.02em;
}

@keyframes float {
  0%, 100% { transform: translateY(0) scale(1); }
  50% { transform: translateY(-25px) scale(1.03); }
}
@keyframes fadeInUp {
  from { opacity: 0; transform: translateY(16px); }
  to { opacity: 1; transform: translateY(0); }
}
@keyframes spin { to { transform: rotate(360deg); } }
</style>
