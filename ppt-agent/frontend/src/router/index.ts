import { createRouter, createWebHistory } from 'vue-router'
import HomePage from '../pages/HomePage.vue'
import AuthPage from '../pages/AuthPage.vue'
import DashboardPage from '../pages/DashboardPage.vue'
import ComposePage from '../pages/ComposePage.vue'
import { isLoggedIn } from '../api'

const router = createRouter({ history: createWebHistory(), routes: [
  { path: '/', component: HomePage },
  { path: '/auth', component: AuthPage },
  { path: '/dashboard', component: DashboardPage, meta: { auth: true } },
  { path: '/compose', component: ComposePage, meta: { auth: true } },
  { path: '/admin', component: () => import('../pages/AdminPage.vue'), meta: { auth: true } },
] })
router.beforeEach(to => to.meta.auth && !isLoggedIn() ? { path: '/auth', query: { next: to.fullPath } } : true)
export default router
