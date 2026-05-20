<template>
  <RouterView />
</template>

<style>
/* ═══════════════════════════════════════════════════════════════════ */
/* Global Reset & Variables                                           */
/* ═══════════════════════════════════════════════════════════════════ */
*, *::before, *::after { margin: 0; padding: 0; box-sizing: border-box; }

:root {
  --sidebar-w: 320px;
  --c-bg: #f8fafc;
  --c-surface: #ffffff;
  --c-sidebar: #0f0d2e;
  --c-sidebar-hover: rgba(99,102,241,0.08);
  --c-sidebar-active: rgba(99,102,241,0.12);
  --c-primary: #3b82f6;
  --c-primary-light: #eff6ff;
  --c-success: #10b981;
  --c-success-light: #ecfdf5;
  --c-danger: #ef4444;
  --c-danger-light: #fef2f2;
  --c-warning: #f59e0b;
  --c-warning-light: #fffbeb;
  --c-text: #0f172a;
  --c-text-2: #334155;
  --c-text-muted: #94a3b8;
  --c-border: #e2e8f0;
  --c-border-light: #f1f5f9;
  --radius-sm: 6px;
  --radius: 10px;
  --radius-lg: 14px;
  --shadow-sm: 0 1px 2px rgba(0,0,0,0.04);
  --shadow: 0 1px 3px rgba(0,0,0,0.06), 0 1px 2px rgba(0,0,0,0.04);
  --shadow-md: 0 4px 12px rgba(0,0,0,0.06), 0 2px 4px rgba(0,0,0,0.04);
  --shadow-lg: 0 10px 30px rgba(0,0,0,0.08), 0 4px 10px rgba(0,0,0,0.04);
  --transition: 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  --transition-slow: 0.4s cubic-bezier(0.4, 0, 0.2, 1);

  /* Compose page aliases */
  --bg-primary: var(--c-bg);
  --bg-secondary: var(--c-surface);
  --border: var(--c-border);
  --text-primary: var(--c-text);
  --text-secondary: var(--c-text-2);
  --text-muted: var(--c-text-muted);
  --color-primary: var(--c-primary);
  --bg-hover: #f1f5f9;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', sans-serif;
  background: var(--c-bg);
  color: var(--c-text);
  line-height: 1.5;
  -webkit-font-smoothing: antialiased;
}

/* View transition */
.view-enter-active { transition: all 0.35s cubic-bezier(0.4, 0, 0.2, 1); }
.view-leave-active { transition: all 0.2s ease-in; }
.view-enter-from { opacity: 0; transform: translateY(12px); }
.view-leave-to { opacity: 0; transform: translateY(-8px); }

/* Fade-up */
.fade-up-enter-active { transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1); }
.fade-up-leave-active { transition: all 0.2s ease-in; }
.fade-up-enter-from { opacity: 0; transform: translateY(10px); }
.fade-up-leave-to { opacity: 0; }
</style>
