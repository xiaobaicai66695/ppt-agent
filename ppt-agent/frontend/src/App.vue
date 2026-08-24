<template>
  <a href="#main-content" class="skip-link">跳至主要内容</a>
  <RouterView />
</template>

<style>
*, *::before, *::after { box-sizing: border-box; }
html, body, #app { min-width: 320px; min-height: 100%; margin: 0; }

:root {
  color-scheme: light;
  --rail-width: 216px;
  --topbar-height: 58px;

  --canvas: #f4f6f2;
  --surface: #ffffff;
  --surface-muted: #f0f4f1;
  --surface-hover: #e7eee9;
  --surface-pressed: #dbe5df;
  --nav-surface: #17312e;
  --nav-text: #f6fbf8;
  --nav-muted: #a7bbb5;

  --text: #1b2523;
  --text-secondary: #4f5e59;
  --text-muted: #72817b;
  --text-disabled: #a9b5b0;

  --border: #d9e2de;
  --border-strong: #bdcbc6;
  --divider: #e5ece8;

  --action: #2dd4bf;
  --action-strong: #5eead4;
  --action-ink: #0f766e;
  --action-soft: #def7f1;
  --info: #2563eb;
  --info-soft: #e9efff;
  --accent-coral: #d97706;
  --accent-coral-soft: #fff3df;
  --success: #17845b;
  --success-soft: #e7f6ef;
  --warning: #a86f00;
  --warning-soft: #fff5d8;
  --danger: #c64242;
  --danger-soft: #fff0f0;

  --radius-xs: 3px;
  --radius-sm: 4px;
  --radius: 6px;
  --radius-md: 8px;
  --radius-lg: 8px;
  --radius-xl: 8px;
  --radius-full: 999px;

  --shadow-xs: 0 1px 2px rgba(18,25,28,0.05);
  --shadow-sm: 0 2px 8px rgba(18,25,28,0.07);
  --shadow: 0 8px 24px rgba(18,25,28,0.1);
  --shadow-md: 0 14px 36px rgba(18,25,28,0.13);
  --shadow-lg: 0 24px 56px rgba(18,25,28,0.16);
  --shadow-card: var(--shadow-xs);
  --shadow-card-hover: var(--shadow-sm);

  --motion-fast: 150ms cubic-bezier(0.2, 0, 0, 1);
  --motion-medium: 240ms cubic-bezier(0.2, 0, 0, 1);
  --transition: var(--motion-fast);
  --transition-md: var(--motion-medium);
  --transition-slow: 320ms cubic-bezier(0.2, 0, 0, 1);

  --z-base: 0;
  --z-header: 20;
  --z-nav: 40;
  --z-popover: 60;
  --z-modal: 100;

  /* Compatibility aliases while legacy scoped styles are migrated. */
  --bg-base: var(--surface);
  --bg-raised: var(--surface);
  --bg-overlay: var(--surface);
  --bg-muted: var(--surface-muted);
  --bg-sidebar: var(--surface-muted);
  --bg-input: var(--surface);
  --bg-primary: var(--surface);
  --bg-secondary: var(--surface-muted);
  --bg-hover: var(--surface-hover);
  --text-primary: var(--text);
  --accent: var(--action-ink);
  --accent-hover: #115e59;
  --accent-soft: var(--action-soft);
  --accent-light: var(--action-soft);
  --accent-border: #a7e6da;
  --accent-text: var(--action-ink);
  --color-primary: var(--action-ink);
  --error: var(--danger);
  --danger-border: #efb9b9;
  --success-border: #a8ddc8;
  --warning-border: #efd38f;
  --info-soft: #eaf0ff;
  --border-light: var(--divider);
  --border-focus: #2dd4bf;
  --status-running-bg: var(--info-soft);
  --status-running-text: var(--info);
  --status-pending-bg: var(--action-soft);
  --status-pending-text: var(--action-ink);
  --status-success-bg: var(--success-soft);
  --status-success-text: var(--success);
  --status-warning-bg: var(--warning-soft);
  --status-warning-text: var(--warning);
  --status-danger-bg: var(--danger-soft);
  --status-danger-text: var(--danger);
  --log-worker-bg: var(--warning);
  --log-file-bg: var(--success);
  --log-tool-bg: var(--info);
  --log-error-bg: var(--danger);
  --log-worker-text: var(--warning);
  --log-tool-text: var(--info);
  --log-error-text: var(--danger);
  --log-file-text: var(--success);
}

html { scroll-behavior: smooth; }
body {
  background: var(--canvas);
  color: var(--text);
  font-family: Inter, ui-sans-serif, -apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif;
  font-size: 14px;
  line-height: 1.5;
  letter-spacing: 0;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

button, input, textarea, select { font: inherit; letter-spacing: 0; }
button, a { touch-action: manipulation; }
button:disabled, input:disabled, textarea:disabled, select:disabled { cursor: not-allowed; opacity: 0.48; }
img { display: block; max-width: 100%; }
::selection { background: #b8eee8; color: #12322f; }

:focus-visible {
  outline: 2px solid var(--info);
  outline-offset: 2px;
}

.skip-link {
  position: fixed;
  top: 8px;
  left: 8px;
  z-index: 9999;
  padding: 9px 12px;
  border-radius: 6px;
  color: #ffffff;
  background: var(--text);
  text-decoration: none;
  transform: translateY(-150%);
  transition: transform var(--motion-fast);
}
.skip-link:focus { transform: translateY(0); }

.ui-button {
  min-height: 40px;
  padding: 0 14px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  border: 1px solid var(--border-strong);
  border-radius: 6px;
  color: var(--text);
  background: var(--surface);
  font-weight: 650;
  text-decoration: none;
  cursor: pointer;
  transition: border-color var(--motion-fast), background var(--motion-fast), transform var(--motion-fast);
}
.ui-button:hover { border-color: #aeb7ba; background: var(--surface-muted); }
.ui-button:active { transform: scale(0.98); }
.ui-button.primary { border-color: var(--action-ink); color: #ffffff; background: var(--action-ink); }
.ui-button.primary:hover { background: #064d48; }
.ui-button.danger { border-color: #e7b7b7; color: var(--danger); background: var(--danger-soft); }

.icon-button {
  width: 44px;
  height: 44px;
  padding: 0;
  display: inline-grid;
  place-items: center;
  border: 1px solid transparent;
  border-radius: 6px;
  color: var(--text-secondary);
  background: transparent;
  cursor: pointer;
  transition: color var(--motion-fast), border-color var(--motion-fast), background var(--motion-fast);
}
.icon-button:hover { color: var(--text); border-color: var(--border); background: var(--surface-muted); }

.visually-hidden {
  position: absolute !important;
  width: 1px !important;
  height: 1px !important;
  padding: 0 !important;
  margin: -1px !important;
  overflow: hidden !important;
  clip: rect(0, 0, 0, 0) !important;
  white-space: nowrap !important;
  border: 0 !important;
}

::-webkit-scrollbar { width: 10px; height: 10px; }
::-webkit-scrollbar-track { background: transparent; }
::-webkit-scrollbar-thumb { background: #c8ced0; border: 3px solid transparent; border-radius: 999px; background-clip: padding-box; }
::-webkit-scrollbar-thumb:hover { background-color: #aab2b5; }

@media (prefers-reduced-motion: reduce) {
  html { scroll-behavior: auto; }
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}
</style>
