<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { FileStack, LayoutDashboard, LogOut, Menu, Moon, PanelLeftClose, Plus, Settings2, Sparkles, Sun } from 'lucide-vue-next'
import { clearToken, logout } from '../api'
import AccountSettingsDialog from './AccountSettingsDialog.vue'
import { useTheme } from '../composables/useTheme'

defineProps<{ title?: string; subtitle?: string; email?: string; guest?: boolean }>()
const emit = defineEmits<{ new: []; compose: [] }>()
const router = useRouter()
const open = ref(true)
const accountOpen = ref(false)
const { isLight, toggleTheme } = useTheme()
async function signOut() { await logout(); clearToken(); router.push('/') }
</script>

<template>
  <div class="studio-shell" :class="{ 'nav-closed': !open }">
    <aside class="studio-nav">
      <div class="nav-brand"><span class="brand-mark"><FileStack :size="18" /></span><span>Deckform</span></div>
      <button class="nav-create" @click="emit('new')"><Plus :size="18" />新建创作</button>
      <nav>
        <RouterLink to="/dashboard"><LayoutDashboard :size="18" />工作台</RouterLink>
        <RouterLink to="/compose" @click="emit('compose')"><Sparkles :size="18" />编排页面</RouterLink>
        <RouterLink to="/admin"><Settings2 :size="18" />运营后台</RouterLink>
      </nav>
      <div class="nav-bottom">
        <div class="identity"><span class="avatar">{{ guest ? 'G' : (email || 'U').slice(0, 1).toUpperCase() }}</span><span><b>{{ guest ? '访客工作区' : (email || '创作账户') }}</b><small>{{ guest ? '退出后无法恢复' : '已连接' }}</small></span></div>
        <button class="quiet-action" @click="accountOpen = true"><Settings2 :size="16" />模型 Key</button>
        <button class="quiet-action" @click="toggleTheme"><Sun v-if="!isLight" :size="16" /><Moon v-else :size="16" />{{ isLight ? '切换深色' : '切换亮色' }}</button>
        <button class="quiet-action" @click="signOut"><LogOut :size="16" />退出</button>
      </div>
    </aside>
    <main class="studio-main">
      <header class="studio-header"><button class="icon-button" :aria-label="open ? '收起导航' : '展开导航'" @click="open = !open"><PanelLeftClose v-if="open" :size="19" /><Menu v-else :size="19" /></button><div><h1>{{ title }}</h1><p v-if="subtitle">{{ subtitle }}</p></div><slot name="header" /></header>
      <slot />
    </main>
    <AccountSettingsDialog v-if="accountOpen" @close="accountOpen = false" />
  </div>
</template>

<style scoped>
.studio-shell { min-height: 100vh; display: grid; grid-template-columns: 248px minmax(0,1fr); background: #091825; transition: grid-template-columns .25s ease; }
.studio-shell.nav-closed { grid-template-columns: 0 minmax(0,1fr); }
.studio-nav { position: sticky; top: 0; height: 100vh; overflow: hidden; display: flex; flex-direction: column; padding: 26px 16px 18px; background: #07131f; border-right: 1px solid rgba(213,235,245,.1); transition: opacity .2s ease; }
.nav-closed .studio-nav { opacity: 0; pointer-events: none; }
.nav-brand { display: flex; gap: 10px; align-items: center; padding: 0 8px 26px; font-family: 'Noto Serif SC', serif; font-size: 21px; font-weight: 700; letter-spacing: -.04em; white-space: nowrap; }
.brand-mark { display: grid; place-items: center; width: 30px; height: 30px; color: #07131f; background: #6ce5ca; border-radius: 9px 9px 3px 9px; }
.nav-create { border: 0; display: flex; align-items: center; gap: 8px; justify-content: center; min-height: 45px; color: #03201f; background: #6ce5ca; border-radius: 8px; font-weight: 700; }
.nav-create:hover { background: #9af1dc; }
nav { margin-top: 28px; display: grid; gap: 4px; }
nav a, .quiet-action { display: flex; align-items: center; gap: 10px; padding: 11px 10px; color: #9bb0bf; border-radius: 7px; font-size: 14px; }
nav a.router-link-active { color: #ecf6f8; background: rgba(132,184,197,.13); }
nav a:hover, .quiet-action:hover { color: #ecf6f8; background: rgba(132,184,197,.08); }
.nav-bottom { margin-top: auto; padding-top: 18px; border-top: 1px solid rgba(213,235,245,.1); }
.identity { display: flex; gap: 9px; padding: 0 8px 16px; align-items: center; font-size: 12px; white-space: nowrap; }
.identity b, .identity small { display: block; overflow: hidden; text-overflow: ellipsis; max-width: 168px; }.identity small { margin-top: 2px; color: #8095a3; }.avatar { width: 29px; height: 29px; display: grid; place-items: center; border: 1px solid #497283; border-radius: 50%; color: #92f0db; font-family: 'DM Mono', monospace; }
.quiet-action { width: 100%; border: 0; background: none; text-align: left; }.studio-main { min-width: 0; }.studio-header { min-height: 78px; padding: 16px 30px; display: flex; gap: 13px; align-items: center; border-bottom: 1px solid rgba(213,235,245,.1); }.studio-header h1 { margin: 0; font-size: 16px; font-weight: 700; }.studio-header p { margin: 2px 0 0; font-size: 12px; color: #8ca2b1; }.icon-button { width: 36px; height: 36px; display: grid; place-items: center; border: 1px solid rgba(213,235,245,.14); border-radius: 7px; color: #aec2cd; background: transparent; }
.studio-main { min-height:0; height:100vh; display:flex; flex-direction:column; overflow:hidden; }
@media (max-width: 760px) { .studio-shell { grid-template-columns: 0 minmax(0,1fr); }.studio-nav { opacity: 0; pointer-events: none; }.studio-header { padding: 14px 18px; } }
</style>
