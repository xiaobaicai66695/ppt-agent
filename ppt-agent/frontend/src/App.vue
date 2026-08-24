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

  --canvas: #f5f7f8;
  --surface: #ffffff;
  --surface-muted: #f0f3f5;
  --surface-hover: #e8edf0;
  --surface-pressed: #dce4e8;
  --nav-surface: #ffffff;
  --nav-text: #2d3942;
  --nav-muted: #73808a;

  --text: #27323a;
  --text-secondary: #53616b;
  --text-muted: #7e8b94;
  --text-disabled: #b0bbc2;

  --border: #d9e1e5;
  --border-strong: #c3cdd3;
  --divider: #e9eef1;

  --action: #d9ecf0;
  --action-strong: #c5e2e8;
  --action-ink: #356b7d;
  --action-soft: #eaf5f7;
  --info: #4677c7;
  --info-soft: #edf3ff;
  --accent-coral: #b97948;
  --accent-coral-soft: #fff2e8;
  --success: #368568;
  --success-soft: #eaf7f0;
  --warning: #a87524;
  --warning-soft: #fff6e5;
  --danger: #c65c5c;
  --danger-soft: #fff1f1;

  --radius-xs: 3px;
  --radius-sm: 4px;
  --radius: 6px;
  --radius-md: 8px;
  --radius-lg: 8px;
  --radius-xl: 8px;
  --radius-full: 999px;

  --shadow-xs: 0 1px 2px rgba(55, 70, 80, 0.05);
  --shadow-sm: 0 2px 8px rgba(55, 70, 80, 0.07);
  --shadow: 0 8px 24px rgba(55, 70, 80, 0.1);
  --shadow-md: 0 14px 36px rgba(55, 70, 80, 0.12);
  --shadow-lg: 0 24px 56px rgba(55, 70, 80, 0.16);
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
::selection { background: #d6e9ee; color: #274d5b; }

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
.ui-button:hover { border-color: var(--border-strong); background: var(--surface-muted); }
.ui-button:active { transform: scale(0.98); }
.ui-button.primary { border-color: var(--action-ink); color: #ffffff; background: var(--action-ink); }
.ui-button.primary:hover { background: #2a5869; }
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
::-webkit-scrollbar-thumb { background: #c8d1d6; border: 3px solid transparent; border-radius: 999px; background-clip: padding-box; }
::-webkit-scrollbar-thumb:hover { background-color: #adb9c0; }

@media (prefers-reduced-motion: reduce) {
  html { scroll-behavior: auto; }
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}
</style>
