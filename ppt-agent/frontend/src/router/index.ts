import { createRouter, createWebHistory } from 'vue-router';
import { isLoggedIn } from '../api';
import { authState } from '../stores/auth';

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'home',
      // The workbench starts in normal chat mode and exposes PPT generation
      // as an explicit capability instead of making it the default action.
      component: () => import('../pages/DashboardPage.vue'),
    },
    {
      path: '/auth',
      name: 'auth',
      component: () => import('../pages/AuthPage.vue'),
      beforeEnter: (_to, _from, next) => {
        if (import.meta.env.DEV) {
          next();
        } else if (isLoggedIn()) {
          next({ name: 'dashboard' });
        } else {
          next();
        }
      },
    },
    {
      path: '/dashboard',
      name: 'dashboard',
      component: () => import('../pages/DashboardPage.vue'),
      beforeEnter: (_to, _from, next) => {
        if (import.meta.env.DEV) {
          next();
        } else if (!isLoggedIn()) {
          next({ name: 'auth', query: { redirect: '/dashboard' } });
        } else {
          next();
        }
      },
    },
    {
      path: '/compose',
      name: 'compose',
      component: () => import('../pages/ComposePage.vue'),
      beforeEnter: (_to, _from, next) => {
        if (import.meta.env.DEV) {
          next();
        } else if (!isLoggedIn()) {
          next({ name: 'auth', query: { redirect: '/compose' } });
        } else {
          next();
        }
      },
    },
    {
      path: '/admin',
      name: 'admin',
      component: () => import('../pages/AdminPage.vue'),
      beforeEnter: async (_to, _from, next) => {
        if (!isLoggedIn()) {
          next({ name: 'auth', query: { redirect: '/admin' } });
          return;
        }
        if (!authState.user) {
          await authState.init();
        }
        if (!authState.isAdmin) {
          next({ name: 'dashboard' });
        } else {
          next();
        }
      },
    },
  ],
});

export default router;
