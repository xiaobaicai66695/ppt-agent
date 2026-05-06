import { createRouter, createWebHistory } from 'vue-router';
import { isLoggedIn } from '../api';

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'home',
      component: () => import('../pages/HomePage.vue'),
    },
    {
      path: '/auth',
      name: 'auth',
      component: () => import('../pages/AuthPage.vue'),
      beforeEnter: (_to, _from, next) => {
        if (isLoggedIn()) next({ name: 'dashboard' });
        else next();
      },
    },
    {
      path: '/dashboard',
      name: 'dashboard',
      component: () => import('../pages/DashboardPage.vue'),
      beforeEnter: (_to, _from, next) => {
        if (!isLoggedIn()) next({ name: 'auth', query: { redirect: '/dashboard' } });
        else next();
      },
    },
  ],
});

export default router;
