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
  { index: '/system/rbac', label: 'RBAC', permission: 'system.rbac' },
]

const visibleMenuItems = computed(() =>
  menuItems.filter((item) => !item.permission || auth.hasPermission(item.permission)),
)

function onSelect(index: string) {
  router.push(index)
}

function signOut() {
  auth.logout()
  router.push({ name: 'login' })
}
</script>

<template>
  <el-container class="app-shell">
    <el-aside width="220px" class="app-aside">
      <div class="brand">PharmaOps</div>
      <el-menu :default-active="activeMenu" class="app-menu" @select="onSelect">
        <el-menu-item v-for="item in visibleMenuItems" :key="item.index" :index="item.index">
          {{ item.label }}
        </el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header class="app-header">
        <span class="header-title">{{ (route.meta.title as string) || 'PharmaOps' }}</span>
        <div class="header-actions">
          <span v-if="auth.username" class="user-label">{{ auth.username }}</span>
          <el-button type="primary" link @click="signOut">Sign out</el-button>
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
}
.app-aside {
  border-right: 1px solid var(--el-border-color);
}
.brand {
  padding: 16px;
  font-weight: 600;
  font-size: 16px;
  border-bottom: 1px solid var(--el-border-color);
}
.app-menu {
  border-right: none;
}
.app-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid var(--el-border-color);
}
.header-title {
  font-weight: 600;
}
.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}
.user-label {
  color: var(--el-text-color-secondary);
  font-size: 14px;
}
.app-main {
  background: var(--el-fill-color-light);
}
</style>
