<template>
  <a href="#main-content" class="skip-link">跳过导航，跳至主要内容</a>
  <RouterView />
</template>

<style>
/* ═══════════════════════════════════════════════════════════════════ */
/* Design Tokens — Modern Light Theme                                */
/* ═══════════════════════════════════════════════════════════════════ */
*, *::before, *::after { margin: 0; padding: 0; box-sizing: border-box; }

:root {
  /* Layout */
  --sidebar-w: 320px;

  /* Backgrounds — 4 elevation levels */
  --bg-base: #ffffff;
  --bg-raised: #ffffff;
  --bg-overlay: #ffffff;
  --bg-muted: #f8fafc;
  --bg-sidebar: #f8fafc;
  --bg-input: #ffffff;

  /* Brand / Primary — Indigo spectrum */
  --accent: #6366f1;
  --accent-hover: #4f46e5;
  --accent-soft: #eef2ff;
  --accent-light: #eef2ff;
  --accent-border: #c7d2fe;
  --accent-text: #4338ca;

  /* Semantic */
  --success: #10b981;
  --success-soft: #ecfdf5;
  --success-border: #a7f3d0;
  --danger: #ef4444;
  --error: #ef4444;
  --danger-soft: #fef2f2;
  --danger-border: #fecaca;
  --warning: #f59e0b;
  --warning-soft: #fffbeb;
  --warning-border: #fde68a;
  --info: #3b82f6;
  --info-soft: #eff6ff;

  /* Text */
  --text: #0f172a;
  --text-secondary: #475569;
  --text-muted: #94a3b8;
  --text-disabled: #cbd5e1;

  /* Borders */
  --border: #e2e8f0;
  --border-light: #f1f5f9;
  --border-focus: var(--accent);

  /* Shadows — 5 levels */
  --shadow-xs: 0 1px 2px rgba(15, 23, 42, 0.04);
  --shadow-sm: 0 1px 3px rgba(15, 23, 42, 0.06), 0 1px 2px rgba(15, 23, 42, 0.04);
  --shadow: 0 4px 12px rgba(15, 23, 42, 0.08), 0 2px 4px rgba(15, 23, 42, 0.04);
  --shadow-md: 0 8px 24px rgba(15, 23, 42, 0.10), 0 4px 8px rgba(15, 23, 42, 0.06);
  --shadow-lg: 0 16px 40px rgba(15, 23, 42, 0.12), 0 8px 16px rgba(15, 23, 42, 0.08);
  --shadow-card: 0 1px 3px rgba(15, 23, 42, 0.06), 0 0 0 1px rgba(15, 23, 42, 0.04);
  --shadow-card-hover: 0 8px 24px rgba(99, 102, 241, 0.12), 0 0 0 1px rgba(99, 102, 241, 0.08);

  /* Border Radius */
  --radius-xs: 4px;
  --radius-sm: 6px;
  --radius: 8px;
  --radius-md: 10px;
  --radius-lg: 14px;
  --radius-xl: 18px;
  --radius-full: 9999px;

  /* Transitions */
  --transition: 0.15s cubic-bezier(0.4, 0, 0.2, 1);
  --transition-md: 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  --transition-slow: 0.4s cubic-bezier(0.4, 0, 0.2, 1);

  /* Status colors */
  --status-running-bg: var(--info-soft);
  --status-running-text: var(--info);
  --status-pending-bg: var(--accent-soft);
  --status-pending-text: var(--accent);
  --status-success-bg: var(--success-soft);
  --status-success-text: var(--success);
  --status-warning-bg: var(--warning-soft);
  --status-warning-text: var(--warning);
  --status-danger-bg: var(--danger-soft);
  --status-danger-text: var(--danger);

  /* Log kind colors */
  --log-worker-bg: #f59e0b;
  --log-file-bg: #10b981;
  --log-tool-bg: #8b5cf6;
  --log-error-bg: #ef4444;
  --log-worker-text: #b45309;
  --log-tool-text: #6d28d9;
  --log-error-text: #dc2626;
  --log-file-text: #059669;

  /* Compose page aliases (backward compat) */
  --bg-primary: var(--bg-base);
  --bg-secondary: var(--bg-raised);
  --text-primary: var(--text);
  --color-primary: var(--accent);
  --bg-hover: var(--bg-muted);
}

/* ═══════════════════════════════════════════════════════════════════ */
/* Base Styles                                                      */
/* ═══════════════════════════════════════════════════════════════════ */
body {
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  background: var(--bg-muted);
  color: var(--text);
  line-height: 1.6;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

/* Text selection */
::selection { background: var(--accent-soft); color: var(--accent-text); }

/* Smooth scroll */
html { scroll-behavior: smooth; scroll-padding-top: 80px; }

/* Focus visible */
:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
  border-radius: var(--radius-sm);
}

/* Skip link */
.skip-link {
  position: absolute; top: -100%; left: 0.5rem;
  background: var(--accent); color: #fff;
  padding: 0.5rem 1rem; border-radius: 0 0 var(--radius) var(--radius);
  font-size: 0.8rem; font-weight: 600; text-decoration: none; z-index: 9999;
  transition: top 0.2s;
}
.skip-link:focus { top: 0; }

/* Respect reduced motion preference */
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
    scroll-behavior: auto !important;
  }
}

/* ═══════════════════════════════════════════════════════════════════ */
/* View Transitions                                                  */
/* ═══════════════════════════════════════════════════════════════════ */
.view-enter-active { transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1); }
.view-leave-active { transition: all 0.2s ease-in; }
.view-enter-from { opacity: 0; transform: translateY(10px); }
.view-leave-to { opacity: 0; transform: translateY(-6px); }

.fade-up-enter-active { transition: all 0.35s cubic-bezier(0.4, 0, 0.2, 1); }
.fade-up-leave-active { transition: all 0.2s ease-in; }
.fade-up-enter-from { opacity: 0; transform: translateY(12px); }
.fade-up-leave-to { opacity: 0; }
</style>
