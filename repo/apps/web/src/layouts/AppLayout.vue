<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const activeMenu = computed(() => route.path)

type MenuItem = { index: string; label: string; permission?: string }

const menuItems: MenuItem[] = [
  { index: '/dashboard', label: 'Dashboard', permission: 'dashboard.view' },
  { index: '/recruitment/candidates', label: 'Candidates', permission: 'recruitment.view' },
  { index: '/recruitment/positions', label: 'Positions', permission: 'recruitment.view' },
  { index: '/compliance/qualifications', label: 'Qualifications', permission: 'compliance.view' },
  { index: '/compliance/restrictions', label: 'Restrictions', permission: 'compliance.view' },
  { index: '/cases', label: 'Cases', permission: 'cases.view' },
  { index: '/files', label: 'Files', permission: 'files.view' },
  { index: '/audit-logs', label: 'Audit logs', permission: 'audit.view' },
  { index: '/system/rbac', label: 'Roles & permissions', permission: 'system.rbac' },
]

const visibleMenuItems = computed(() =>
  menuItems.filter((item) => !item.permission || auth.hasPermission(item.permission)),
)

function onSelect(index: string) {
  router.push(index)
}

async function signOut() {
  await auth.logout()
  router.push({ name: 'login' })
}
</script>

<template>
  <el-container class="app-shell">
    <el-aside width="248px" class="app-aside">
      <div class="brand">
        <span class="brand-mark" aria-hidden="true" />
        <div class="brand-text">
          <span class="brand-name">PharmaOps</span>
          <span class="brand-tag">Operations</span>
        </div>
      </div>
      <el-scrollbar class="aside-scroll">
        <el-menu
          :default-active="activeMenu"
          class="app-menu"
          :router="false"
          @select="onSelect"
        >
          <el-menu-item v-for="item in visibleMenuItems" :key="item.index" :index="item.index">
            {{ item.label }}
          </el-menu-item>
        </el-menu>
      </el-scrollbar>
    </el-aside>
    <el-container class="app-body">
      <el-header class="app-header" height="64px">
        <div class="header-left">
          <span class="header-title">{{ (route.meta.title as string) || 'Dashboard' }}</span>
        </div>
        <div class="header-actions">
          <span v-if="auth.username" class="user-label">{{ auth.username }}</span>
          <el-button round @click="signOut">Sign out</el-button>
        </div>
      </el-header>
      <el-main class="app-main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<style scoped>
.app-shell {
  min-height: 100vh;
  background: linear-gradient(180deg, #f8fafc 0%, #f1f5f9 100%);
}

.app-aside {
  display: flex;
  flex-direction: column;
  background: #ffffff;
  border-right: 1px solid var(--el-border-color-lighter);
  box-shadow: 4px 0 24px rgba(15, 23, 42, 0.04);
}

.brand {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 1.25rem 1rem 1rem;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.brand-mark {
  width: 36px;
  height: 36px;
  border-radius: 12px;
  background: linear-gradient(135deg, #0d9488 0%, #0f766e 100%);
  box-shadow: 0 4px 12px rgba(13, 148, 136, 0.35);
  flex-shrink: 0;
}

.brand-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.brand-name {
  font-weight: 700;
  font-size: 1rem;
  letter-spacing: -0.02em;
  color: var(--el-text-color-primary);
}

.brand-tag {
  font-size: 0.6875rem;
  font-weight: 600;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--el-text-color-secondary);
}

.aside-scroll {
  flex: 1;
  padding: 0.75rem 0.5rem 1rem;
}

.app-menu {
  border-right: none;
  background: transparent;
}

.app-menu :deep(.el-menu-item) {
  border-radius: 10px;
  margin: 2px 0;
  height: 42px;
  line-height: 42px;
  font-weight: 500;
}

.app-menu :deep(.el-menu-item.is-active) {
  background: var(--el-color-primary-light-9) !important;
  color: var(--el-color-primary-dark-2) !important;
}

.app-menu :deep(.el-menu-item:hover) {
  background: var(--el-fill-color-light) !important;
}

.app-body {
  flex-direction: column;
  min-width: 0;
}

.app-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 1.5rem;
  margin: 0.75rem 1rem 0;
  background: #ffffff;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 14px;
  box-shadow: var(--el-box-shadow);
}

.header-left {
  min-width: 0;
}

.header-title {
  font-weight: 650;
  font-size: 1.125rem;
  letter-spacing: -0.02em;
  color: var(--el-text-color-primary);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.user-label {
  color: var(--el-text-color-secondary);
  font-size: 0.875rem;
}

.app-main {
  padding: 1rem 1.25rem 1.5rem;
}
</style>
