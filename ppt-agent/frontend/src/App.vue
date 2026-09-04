<script setup lang="ts">
import { RouterView } from 'vue-router'
import { useTheme } from './composables/useTheme'
useTheme()
</script>

<template><RouterView v-slot="{ Component }"><Transition name="page" mode="out-in"><component :is="Component" /></Transition></RouterView></template>

<style>
@import url('https://fonts.googleapis.com/css2?family=Noto+Sans+SC:wght@400;500;600;700&family=Noto+Serif+SC:wght@600;700;900&family=DM+Mono:wght@400;500&display=swap');

:root { font-family: 'Noto Sans SC', system-ui, sans-serif; color: #e9f1f8; background: #07131f; font-synthesis: none;
  --surface-base:#0c1d2a; --surface-raised:#0b1a27; --surface-accent:#0a2631; --surface-hover:#15303e;
  --border-subtle:rgba(213,235,245,.1); --border-strong:#466473;
  --text-strong:#e9f1f8; --text-muted:#b5cbd0; --text-subtle:#7898a1; --text-on-accent:#d9edf0;
  --accent:#6ce5ca; --accent-strong:#9af1dc; --accent-soft:rgba(102,225,202,.1); --accent-on:#03201f;
  --info:#79c6e5; --danger:#ef9892; --message-user:#173846; --message-user-label:#94d7ce; --message-assistant:#f0f1e9;
}
* { box-sizing: border-box; }
body { margin: 0; min-width: 320px; min-height: 100vh; background: #07131f; }
button, input, textarea, select { font: inherit; }
button { cursor: pointer; }
a { color: inherit; text-decoration: none; }
:focus-visible { outline: 2px solid #65e5cb; outline-offset: 3px; }

html[data-theme='light'] { color:#163841; background:#edf2f0;
  --surface-base:#eaf0ee; --surface-raised:#f9fbf7; --surface-accent:#dcebe7; --surface-hover:#dceae5;
  --border-subtle:#cfddd8; --border-strong:#9ebbb5;
  --text-strong:#183943; --text-muted:#43636a; --text-subtle:#52737a; --text-on-accent:#123b45;
  --accent:#176f72; --accent-strong:#13535a; --accent-soft:rgba(23,111,114,.14); --accent-on:#f7fffc;
  --info:#2f7da0; --danger:#b64e4b; --message-user:#cce5e2; --message-user-label:#2a7372; --message-assistant:#ffffff;
}
html[data-theme='light'] body{background:#edf2f0;color:#163841}
html[data-theme='light'] .studio-shell,html[data-theme='light'] .studio-main,html[data-theme='light'] .canvas,html[data-theme='light'] .compose,html[data-theme='light'] .composer{background:#eaf0ee!important;color:#153842}
html[data-theme='light'] .studio-nav,html[data-theme='light'] .conversations{background:#f9fbf7!important;border-color:#cfddd8!important}
html[data-theme='light'] .studio-header,html[data-theme='light'] .canvas-head,html[data-theme='light'] .slide-list,html[data-theme='light'] .delivery{border-color:#cfddd8!important}
html[data-theme='light'] .studio-header h1,html[data-theme='light'] .canvas-head h2,html[data-theme='light'] .nav-brand,html[data-theme='light'] .identity b,html[data-theme='light'] .editor h2{color:#183943!important}
html[data-theme='light'] nav a,html[data-theme='light'] .quiet-action,html[data-theme='light'] .task-row,html[data-theme='light'] .task-info b,html[data-theme='light'] .blank-canvas h3,html[data-theme='light'] .canvas-kicker{color:#43636a!important}
html[data-theme='light'] .task-row.selected,html[data-theme='light'] .task-row:hover{background:#dceae5!important;color:#173e45!important}
html[data-theme='light'] .activity{background:#dcebe7!important;color:#355b62!important}
html[data-theme='light'] .modebar button.on,html[data-theme='light'] .compose-link{background:#d4eae4!important;color:#13535a!important}
html[data-theme='light'] .composer{border-color:#cfddd8!important}
html[data-theme='light'] .modebar button,html[data-theme='light'] .modebar label{color:#52737a!important}
html[data-theme='light'] .composer-input{background:#ffffff!important;box-shadow:0 8px 22px rgba(35,74,72,.08)}
html[data-theme='light'] .composer-input textarea{color:#183943!important}
html[data-theme='light'] .composer-input button{background:#176f72!important;color:#f7fffc!important}
html[data-theme='light'] .slide-row.selected{background:#d7e9e4!important;color:#143e47!important}
html[data-theme='light'] .message.user{background:#cce5e2!important;color:#123b45!important}
html[data-theme='light'] .home-nav,html[data-theme='light'] .how{border-color:#c8d8d3!important}
html[data-theme='light'] .landing{background:#eaf1ee!important;color:#153941!important}
html[data-theme='light'] .landing .lede,html[data-theme='light'] .landing .sequence span,html[data-theme='light'] .landing .how>p,html[data-theme='light'] .landing .text-link{color:#4e6c70!important}
html[data-theme='light'] .auth-page{background:#eaf1ee!important}
html[data-theme='light'] .auth-aside{background:#164b56!important}
@media (prefers-reduced-motion:no-preference){.page-enter-active,.page-leave-active{transition:opacity .18s ease,transform .18s ease}.page-enter-from{opacity:0;transform:translateY(6px)}.page-leave-to{opacity:0;transform:translateY(-4px)}}
</style>
