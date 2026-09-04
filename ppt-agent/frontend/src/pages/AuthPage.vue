<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft, ArrowRight, Check, FileStack, KeyRound, LockKeyhole, Mail, UserRound } from 'lucide-vue-next'
import { login, loginAsGuest, register, sendCode } from '../api'

type LoginMode = 'code' | 'password'

const router = useRouter()
const route = useRoute()
const email = ref('')
const code = ref('')
const password = ref('')
const confirmPassword = ref('')
const loginMode = ref<LoginMode>('code')
const sent = ref(false)
const busy = ref(false)
const error = ref('')
const notice = ref('')

const isRegister = computed(() => route.name === 'register')
const passwordRule = computed(() => /^(?=.{8,64}$)(?=.*[a-z])(?=.*[A-Z])(?=.*\d).+$/.test(password.value))
const emailValid = computed(() => /^\S+@\S+\.\S+$/.test(email.value.trim()))
const loginValid = computed(() => emailValid.value && (loginMode.value === 'code' ? code.value.trim().length > 0 : password.value.length > 0))
const registerValid = computed(() => emailValid.value && code.value.trim().length > 0 && passwordRule.value && password.value === confirmPassword.value)
const next = () => String(route.query.next || '/dashboard')

function resetFeedback() { error.value = ''; notice.value = '' }
async function send() { resetFeedback(); busy.value = true; try { await sendCode(email.value.trim()); sent.value = true; notice.value = '验证码已发送，请在 5 分钟内完成操作。' } catch (e) { error.value = e instanceof Error ? e.message : '验证码发送失败' } finally { busy.value = false } }
async function submit() { resetFeedback(); busy.value = true; try { if (isRegister.value) { await register(email.value.trim(), code.value.trim(), password.value) } else { await login(email.value.trim(), loginMode.value === 'code' ? code.value.trim() : password.value, loginMode.value === 'password') }; await router.push(next()) } catch (e) { error.value = e instanceof Error ? e.message : (isRegister.value ? '注册失败' : '登录失败') } finally { busy.value = false } }
async function guest() { resetFeedback(); busy.value = true; try { await loginAsGuest(); await router.push(next()) } catch (e) { error.value = e instanceof Error ? e.message : '访客登录失败' } finally { busy.value = false } }
function switchLoginMode(mode: LoginMode) { loginMode.value = mode; resetFeedback() }
</script>

<template>
  <main class="auth-page">
    <RouterLink to="/" class="back"><ArrowLeft :size="16" />返回首页</RouterLink>
    <section class="auth-panel" :aria-labelledby="isRegister ? 'register-title' : 'login-title'">
      <div class="brand"><span><FileStack :size="21" /></span><b>Deckform</b></div>
      <p class="overline">{{ isRegister ? '创建你的工作台' : '回到创作现场' }}</p>
      <h1 :id="isRegister ? 'register-title' : 'login-title'">{{ isRegister ? '从这里，建立你的\n演示工作台。' : '把你的想法\n排成作品。' }}</h1>
      <p class="intro">{{ isRegister ? '用邮箱验证身份，并设置一组只属于你的登录密码。' : '登录后可保存会话、继续生成，并管理你的交付文件。' }}</p>

      <div v-if="!isRegister" class="switch" aria-label="登录方式"><button type="button" :class="{ on: loginMode === 'code' }" @click="switchLoginMode('code')">验证码登录</button><button type="button" :class="{ on: loginMode === 'password' }" @click="switchLoginMode('password')">密码登录</button></div>

      <form novalidate @submit.prevent="submit">
        <div class="field"><label for="auth-email">邮箱地址</label><div class="input-wrap"><Mail :size="16" aria-hidden="true" /><input id="auth-email" v-model.trim="email" type="email" autocomplete="email" placeholder="name@example.com" :aria-invalid="Boolean(error) && !emailValid" required /></div></div>
        <div v-if="isRegister || loginMode === 'code'" class="field"><div class="field-head"><label for="auth-code">邮箱验证码</label><button type="button" class="inline-action" :disabled="busy || !emailValid" @click="send">{{ sent ? '重新发送' : '发送验证码' }}</button></div><div class="input-wrap"><KeyRound :size="16" aria-hidden="true" /><input id="auth-code" v-model.trim="code" inputmode="numeric" autocomplete="one-time-code" maxlength="6" placeholder="输入 6 位验证码" required /></div><p v-if="sent" class="field-note success"><Check :size="13" />{{ notice || '验证码已发送，请检查邮箱。' }}</p></div>
        <div v-if="isRegister || loginMode === 'password'" class="field"><label for="auth-password">{{ isRegister ? '设置密码' : '密码' }}</label><div class="input-wrap"><LockKeyhole :size="16" aria-hidden="true" /><input id="auth-password" v-model="password" type="password" :autocomplete="isRegister ? 'new-password' : 'current-password'" :placeholder="isRegister ? '8–64 位，含大小写字母和数字' : '输入密码'" required /></div><p v-if="isRegister" class="field-note" :class="{ success: passwordRule }"><Check v-if="passwordRule" :size="13" />密码需为 8–64 位，并包含大写字母、小写字母和数字。</p></div>
        <div v-if="isRegister" class="field"><label for="auth-confirm-password">确认密码</label><div class="input-wrap"><LockKeyhole :size="16" aria-hidden="true" /><input id="auth-confirm-password" v-model="confirmPassword" type="password" autocomplete="new-password" placeholder="再次输入密码" required :aria-invalid="confirmPassword.length > 0 && confirmPassword !== password" /></div><p v-if="confirmPassword.length > 0 && confirmPassword !== password" class="field-note error">两次输入的密码不一致。</p></div>
        <p v-if="error" class="form-error" role="alert">{{ error }}</p>
        <button class="submit" type="submit" :disabled="busy || (isRegister ? !registerValid : !loginValid)"><span>{{ busy ? (isRegister ? '正在创建…' : '正在登录…') : (isRegister ? '创建并进入工作台' : '进入工作台') }}</span><ArrowRight :size="17" /></button>
      </form>

      <p class="auth-change"><template v-if="isRegister">已有账号？<RouterLink :to="{ name: 'login', query: route.query }">去登录</RouterLink></template><template v-else>还没有账号？<RouterLink :to="{ name: 'register', query: route.query }">创建账号</RouterLink></template></p>
      <template v-if="!isRegister"><div class="rule"><span>或先体验</span></div><button class="guest" type="button" :disabled="busy" @click="guest"><UserRound :size="17" />以访客身份继续</button><p class="guest-hint">访客数据仅保留在当前工作区，退出后不可恢复。</p></template>
    </section>
    <aside class="auth-aside" aria-hidden="true"><p>{{ isRegister ? '先验证身份，再开始创作' : '从第一行想法开始' }}</p><div class="quote">不是把文字<br>塞进模板，<br><em>而是组织一场表达。</em></div><span>DECKFORM / CREATIVE STUDIO</span></aside>
  </main>
</template>

<style scoped>
.auth-page{min-height:100vh;display:grid;grid-template-columns:minmax(0,1fr) minmax(370px,.82fr);background:#eff1e8;color:#0e2d39}.back{position:fixed;z-index:1;top:25px;left:31px;display:flex;align-items:center;gap:6px;color:#62747b;font-size:13px}.auth-panel{width:min(420px,calc(100% - 48px));margin:auto;padding:85px 0 48px}.brand{display:flex;align-items:center;gap:9px;font:700 20px 'Noto Serif SC',serif}.brand span{display:grid;place-items:center;width:33px;height:33px;color:#eff1e8;background:#0f4a55;border-radius:9px 9px 3px 9px}.overline{margin:60px 0 12px;color:#44817d;font-size:12px;font-weight:700}.auth-panel h1{white-space:pre-line;margin:0;font:900 42px/1.18 'Noto Serif SC',serif;letter-spacing:-.07em}.intro{margin:17px 0 28px;color:#63777d;font-size:14px;line-height:1.8}.switch{display:flex;border-bottom:1px solid #d1dad6;margin-bottom:20px}.switch button{padding:10px 6px;border:0;border-bottom:2px solid transparent;margin-bottom:-1px;background:transparent;color:#73848a;font-size:13px}.switch button+button{margin-left:19px}.switch .on{color:#0d3f4a;border-color:#159587;font-weight:700}.field{margin-bottom:16px}.field label,.field-head{display:block;color:#39565d;font-size:13px;font-weight:700}.field-head{display:flex;justify-content:space-between;align-items:baseline}.input-wrap{display:flex;align-items:center;gap:10px;margin-top:7px;padding:0 13px;border:1px solid #c7d2ce;border-radius:6px;color:#71878a;background:#f7f8f2;transition:border-color .16s ease,box-shadow .16s ease}.input-wrap:focus-within{border-color:#159587;box-shadow:0 0 0 3px rgba(21,149,135,.12)}input{width:100%;height:45px;border:0;outline:0;background:transparent;color:#122f37}.inline-action{padding:0;border:0;background:transparent;color:#16766e;font-size:12px;text-decoration:underline}.inline-action:disabled{opacity:.45;cursor:not-allowed}.field-note{display:flex;align-items:center;gap:5px;margin:7px 0 0;color:#718187;font-size:12px;line-height:1.5}.field-note.success{color:#18776c}.field-note.error,.form-error{color:#b53936}.form-error{margin:0 0 12px;font-size:13px}.submit,.guest{width:100%;height:47px;display:flex;align-items:center;justify-content:center;gap:8px;border-radius:6px;font-weight:700}.submit{border:0;color:#f2f6ef;background:#0b4551}.submit:disabled{opacity:.45;cursor:not-allowed}.auth-change{margin:17px 0 0;color:#72848a;font-size:13px;text-align:center}.auth-change a{margin-left:5px;color:#16766e;font-weight:700}.rule{display:flex;align-items:center;gap:12px;margin:25px 0 0;color:#849595;font-size:11px}.rule:before,.rule:after{content:'';height:1px;flex:1;background:#d4dcd7}.guest{margin-top:17px;color:#174753;background:transparent;border:1px solid #b9c9c4}.guest:disabled{opacity:.45;cursor:not-allowed}.guest-hint{margin:9px 0 0;color:#899798;font-size:11px;line-height:1.6;text-align:center}.auth-aside{display:flex;flex-direction:column;justify-content:center;padding:72px clamp(34px,7vw,100px);color:#ddf7f0;background:#0c3442}.auth-aside>p{margin:0 0 20px;color:#7ccfbd;font-size:13px}.quote{font:900 clamp(38px,4.5vw,66px)/1.21 'Noto Serif SC',serif;letter-spacing:-.08em}.quote em{font-style:normal;color:#70d7c1}.auth-aside span{margin-top:60px;color:#78a0a3;font:11px 'DM Mono',monospace;letter-spacing:.08em}@media(max-width:760px){.auth-page{display:block}.auth-aside{display:none}.auth-panel{padding-top:95px}.back{left:18px;top:20px}}@media(prefers-reduced-motion:reduce){.input-wrap{transition:none}}
</style>
