<script setup lang="ts">
import { RouterView } from 'vue-router'
</script>

<template><RouterView /></template>

<style>
@import url('https://fonts.googleapis.com/css2?family=Noto+Sans+SC:wght@400;500;600;700&family=Noto+Serif+SC:wght@600;700;900&family=DM+Mono:wght@400;500&display=swap');

:root { font-family: 'Noto Sans SC', system-ui, sans-serif; color: #e9f1f8; background: #07131f; font-synthesis: none; }
* { box-sizing: border-box; }
body { margin: 0; min-width: 320px; min-height: 100vh; background: #07131f; }
button, input, textarea, select { font: inherit; }
button { cursor: pointer; }
a { color: inherit; text-decoration: none; }
:focus-visible { outline: 2px solid #65e5cb; outline-offset: 3px; }
</style>
