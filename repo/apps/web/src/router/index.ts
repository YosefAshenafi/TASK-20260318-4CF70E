import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

import AppLayout from '@/layouts/AppLayout.vue'
import { useAuthStore } from '@/stores/auth'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'login',
    component: () => import('@/views/LoginView.vue'),
    meta: { public: true, title: 'Sign in' },
  },
  {
    path: '/',
    component: AppLayout,
    meta: { requiresAuth: true },
    children: [
      { path: '', redirect: '/dashboard' },
      {
        path: 'dashboard',
        name: 'dashboard',
        component: () => import('@/views/DashboardView.vue'),
        meta: { title: 'Dashboard', permission: 'dashboard.view' },
      },
      {
        path: 'recruitment/candidates',
        name: 'recruitment-candidates',
        component: () => import('@/views/ModulePlaceholderView.vue'),
        meta: { title: 'Candidates', permission: 'recruitment.view' },
      },
      {
        path: 'recruitment/positions',
        name: 'recruitment-positions',
        component: () => import('@/views/ModulePlaceholderView.vue'),
        meta: { title: 'Positions', permission: 'recruitment.view' },
      },
      {
        path: 'compliance/qualifications',
        name: 'compliance-qualifications',
        component: () => import('@/views/ModulePlaceholderView.vue'),
        meta: { title: 'Qualifications', permission: 'compliance.view' },
      },
      {
        path: 'compliance/restrictions',
        name: 'compliance-restrictions',
        component: () => import('@/views/ModulePlaceholderView.vue'),
        meta: { title: 'Purchase restrictions', permission: 'compliance.view' },
      },
      {
        path: 'cases',
        name: 'cases',
        component: () => import('@/views/ModulePlaceholderView.vue'),
        meta: { title: 'Cases', permission: 'cases.view' },
      },
      {
        path: 'files',
        name: 'files',
        component: () => import('@/views/ModulePlaceholderView.vue'),
        meta: { title: 'Files', permission: 'files.view' },
      },
      {
        path: 'audit-logs',
        name: 'audit-logs',
        component: () => import('@/views/ModulePlaceholderView.vue'),
        meta: { title: 'Audit logs', permission: 'audit.view' },
      },
      {
        path: 'system/rbac',
        name: 'system-rbac',
        component: () => import('@/views/ModulePlaceholderView.vue'),
        meta: { title: 'Roles & permissions', permission: 'system.rbac' },
      },
    ],
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'not-found',
    component: () => import('@/views/NotFoundView.vue'),
    meta: { public: true, title: 'Not found' },
  },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

router.beforeEach((to) => {
  const auth = useAuthStore()

  if (to.meta.public) {
    if (to.name === 'login' && auth.isAuthenticated) {
      return { path: '/dashboard' }
    }
    return true
  }

  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }

  const perm = to.meta.permission as string | undefined
  if (perm && !auth.hasPermission(perm)) {
    return { name: 'dashboard' }
  }

  return true
})

export default router
