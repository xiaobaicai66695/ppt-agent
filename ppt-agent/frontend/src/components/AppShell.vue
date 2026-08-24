<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import {
  Home,
  KeyRound,
  LogIn,
  LogOut,
  Menu,
  PanelsTopLeft,
  Plus,
  Presentation,
  ShieldCheck,
  Sparkles,
  X,
} from 'lucide-vue-next';
import { authState } from '../stores/auth';
import { isLoggedIn } from '../api';
import AccountSettingsDialog from './AccountSettingsDialog.vue';

defineProps<{
  title: string;
  eyebrow?: string;
  contentClass?: string;
}>();

const route = useRoute();
const router = useRouter();
const navOpen = ref(false);
const accountSettingsOpen = ref(false);
const auth = authState;

const navItems = computed(() => {
  const items = [
    { to: '/', label: '开始', icon: Home },
    { to: '/compose', label: '编排', icon: PanelsTopLeft },
    { to: '/dashboard', label: '任务', icon: Presentation },
  ];
  if (auth.isAdmin) items.push({ to: '/admin', label: '管理', icon: ShieldCheck });
  return items;
});

const userLabel = computed(() => auth.user?.email || '访客');
const userInitial = computed(() => userLabel.value.slice(0, 1).toUpperCase());

onMounted(async () => {
  if (isLoggedIn() && !auth.user) await auth.init();
});

function isActive(path: string) {
  return path === '/' ? route.path === '/' : route.path.startsWith(path);
}

function go(path: string) {
  navOpen.value = false;
  router.push(path);
}

async function handleAuthAction() {
  navOpen.value = false;
  if (auth.loggedIn) {
    await auth.logout();
    await router.push('/auth');
    return;
  }
  await router.push('/auth');
}
</script>

<template>
  <div class="app-shell" :class="{ 'nav-open': navOpen }">
    <aside id="app-navigation" class="app-rail" aria-label="主导航">
      <button class="brand-button" type="button" aria-label="返回开始页" @click="go('/')">
        <span class="brand-mark" aria-hidden="true">
          <Sparkles :size="18" :stroke-width="2" />
        </span>
        <span class="brand-copy">
          <strong>PPT Agent</strong>
          <small>Presentation studio</small>
        </span>
      </button>

      <button class="new-deck-button" type="button" @click="go('/compose')">
        <Plus :size="18" />
        <span>新建演示</span>
      </button>

      <nav class="rail-nav">
        <button
          v-for="item in navItems"
          :key="item.to"
          class="rail-link"
          :class="{ active: isActive(item.to) }"
          type="button"
          :aria-current="isActive(item.to) ? 'page' : undefined"
          @click="go(item.to)"
        >
          <component :is="item.icon" :size="18" :stroke-width="1.8" />
          <span>{{ item.label }}</span>
        </button>
      </nav>

      <div class="rail-footer">
        <div class="rail-user">
          <span class="user-avatar" aria-hidden="true">{{ userInitial }}</span>
          <span class="user-copy">
            <strong>{{ userLabel }}</strong>
            <small>{{ auth.isAdmin ? '管理员' : auth.loggedIn ? '个人工作区' : '尚未登录' }}</small>
          </span>
        </div>
        <button
          v-if="auth.loggedIn"
          class="rail-auth"
          type="button"
          title="账户设置"
          aria-label="账户设置"
          @click="accountSettingsOpen = true"
        >
          <KeyRound :size="17" />
        </button>
        <button
          class="rail-auth"
          type="button"
          :title="auth.loggedIn ? '退出登录' : '登录'"
          :aria-label="auth.loggedIn ? '退出登录' : '登录'"
          @click="handleAuthAction"
        >
          <LogOut v-if="auth.loggedIn" :size="18" />
          <LogIn v-else :size="18" />
        </button>
      </div>
    </aside>

    <button class="nav-scrim" type="button" aria-label="关闭导航" @click="navOpen = false" />

    <section class="app-stage">
      <header class="app-topbar">
        <button
          class="mobile-nav-button"
          type="button"
          :aria-expanded="navOpen"
          aria-controls="app-navigation"
          aria-label="打开主导航"
          @click="navOpen = !navOpen"
        >
          <X v-if="navOpen" :size="20" />
          <Menu v-else :size="20" />
        </button>
        <div class="page-context">
          <span v-if="eyebrow" class="page-eyebrow">{{ eyebrow }}</span>
          <h1>{{ title }}</h1>
        </div>
        <div class="topbar-actions">
          <slot name="actions" />
        </div>
      </header>

      <main id="main-content" class="app-content" :class="contentClass">
        <slot />
      </main>
    </section>

    <AccountSettingsDialog :open="accountSettingsOpen" @close="accountSettingsOpen = false" />
  </div>
</template>

<style scoped>
.app-shell {
  min-height: 100dvh;
  background: var(--canvas);
}

.app-rail {
  position: fixed;
  inset: 0 auto 0 0;
  z-index: var(--z-nav);
  width: var(--rail-width);
  padding: 14px 12px;
  display: flex;
  flex-direction: column;
  background: var(--nav-surface);
  color: var(--nav-text);
  border-right: 1px solid var(--border);
  box-shadow: 2px 0 16px rgba(65, 80, 90, 0.04);
}

.brand-button,
.new-deck-button,
.rail-link,
.rail-auth,
.mobile-nav-button,
.nav-scrim {
  appearance: none;
  border: 0;
  font: inherit;
}

.brand-button {
  min-height: 48px;
  padding: 4px 6px;
  display: flex;
  align-items: center;
  gap: 10px;
  color: inherit;
  background: transparent;
  text-align: left;
  cursor: pointer;
}

.brand-mark {
  width: 34px;
  height: 34px;
  display: grid;
  place-items: center;
  flex: 0 0 auto;
  color: var(--action-ink);
  background: var(--action-soft);
  border: 1px solid #cfe2e9;
  border-radius: 7px;
}

.brand-copy,
.user-copy {
  min-width: 0;
  display: flex;
  flex-direction: column;
  line-height: 1.2;
}

.brand-copy strong { font-size: 14px; font-weight: 700; }
.brand-copy small { margin-top: 3px; color: var(--nav-muted); font-size: 10px; }

.new-deck-button {
  min-height: 42px;
  margin: 18px 2px 14px;
  padding: 0 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  border-radius: 7px;
  color: var(--action-ink);
  border: 1px solid #cfe2e9;
  background: var(--action-soft);
  font-weight: 700;
  cursor: pointer;
  transition: background var(--motion-fast), transform var(--motion-fast);
}

.new-deck-button:hover { border-color: #b5d4df; background: var(--action-strong); }
.new-deck-button:active { transform: scale(0.98); }

.rail-nav {
  display: grid;
  gap: 3px;
}

.rail-link {
  min-height: 42px;
  padding: 0 11px;
  display: flex;
  align-items: center;
  gap: 11px;
  border-radius: 6px;
  color: var(--nav-muted);
  background: transparent;
  cursor: pointer;
  transition: color var(--motion-fast), background var(--motion-fast);
}

.rail-link:hover {
  color: var(--text);
  background: var(--surface-hover);
}

.rail-link.active {
  color: var(--action-ink);
  background: var(--action-soft);
  box-shadow: inset 3px 0 0 var(--action-ink);
}

.rail-link span { font-size: 13px; font-weight: 600; }

.rail-footer {
  margin-top: auto;
  padding: 12px 4px 2px;
  display: flex;
  align-items: center;
  gap: 8px;
  border-top: 1px solid var(--border);
}

.rail-user {
  min-width: 0;
  flex: 1;
  display: flex;
  align-items: center;
  gap: 9px;
}

.user-avatar {
  width: 32px;
  height: 32px;
  display: grid;
  place-items: center;
  flex: 0 0 auto;
  border-radius: 50%;
  color: var(--action-ink);
  background: var(--action-soft);
  border: 1px solid #cfe2e9;
  font-size: 12px;
  font-weight: 700;
}

.user-copy strong {
  overflow: hidden;
  color: var(--nav-text);
  font-size: 11px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.user-copy small { margin-top: 3px; color: var(--nav-muted); font-size: 10px; }

.rail-auth {
  width: 40px;
  height: 40px;
  display: grid;
  place-items: center;
  flex: 0 0 auto;
  border-radius: 6px;
  color: var(--nav-muted);
  background: transparent;
  cursor: pointer;
}

.rail-auth:hover { color: var(--action-ink); background: var(--action-soft); }

.app-stage {
  min-height: 100dvh;
  margin-left: var(--rail-width);
}

.app-topbar {
  position: sticky;
  top: 0;
  z-index: var(--z-header);
  min-height: var(--topbar-height);
  padding: 0 24px;
  display: flex;
  align-items: center;
  gap: 14px;
  background: rgba(255,255,255,0.96);
  border-bottom: 1px solid var(--border);
  backdrop-filter: blur(10px);
}

.page-context { min-width: 0; flex: 1; }
.page-eyebrow { display: block; color: var(--text-muted); font-size: 10px; font-weight: 700; }
.page-context h1 {
  overflow: hidden;
  color: var(--text);
  font-size: 15px;
  font-weight: 700;
  line-height: 1.25;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.topbar-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
}

.mobile-nav-button {
  width: 44px;
  height: 44px;
  display: none;
  place-items: center;
  flex: 0 0 auto;
  border-radius: 6px;
  color: var(--text);
  background: transparent;
  cursor: pointer;
}

.mobile-nav-button:hover { background: var(--surface-muted); }

.app-content { min-width: 0; }
.nav-scrim { display: none; }

@media (max-width: 1024px) {
  .app-rail {
    width: min(82vw, 240px);
    transform: translateX(-102%);
    transition: transform var(--motion-medium);
  }
  .nav-open .app-rail { transform: translateX(0); }
  .app-stage { margin-left: 0; }
  .mobile-nav-button { display: grid; }
  .nav-scrim {
    position: fixed;
    inset: 0;
    z-index: calc(var(--z-nav) - 1);
    display: block;
    visibility: hidden;
    background: rgba(38, 48, 56, 0.18);
    opacity: 0;
    cursor: pointer;
    transition: opacity var(--motion-medium), visibility var(--motion-medium);
  }
  .nav-open .nav-scrim { visibility: visible; opacity: 1; }
}

@media (max-width: 600px) {
  .app-topbar { min-height: 56px; padding: 0 12px; }
  .page-eyebrow { display: none; }
  .page-context h1 { font-size: 14px; }
}
</style>
