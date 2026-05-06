import { createApp } from 'vue';
import App from './App.vue';
import router from './router';

const app = createApp(App);

// Global error handler
app.config.errorHandler = (err, _vm, info) => {
  console.error('[Vue Error]', err, info);
  const el = document.createElement('div');
  el.style.cssText = 'position:fixed;top:0;left:0;right:0;background:#dc2626;color:#fff;padding:12px 20px;z-index:99999;font:14px monospace;white-space:pre-wrap;';
  el.textContent = `Vue Error: ${(err as Error).message || String(err)}\nInfo: ${info}`;
  document.body.appendChild(el);
};

app.use(router);
app.mount('#app');
