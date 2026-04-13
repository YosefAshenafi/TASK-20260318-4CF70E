import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

/** Mock permissions for UI scaffolding; replaced by API session later. */
const MOCK_ALL_PERMISSIONS = [
  'dashboard.view',
  'recruitment.view',
  'compliance.view',
  'cases.view',
  'files.view',
  'audit.view',
  'system.rbac',
] as const

export const useAuthStore = defineStore('auth', () => {
  const username = ref<string | null>(null)
  const permissions = ref<readonly string[]>([])

  const isAuthenticated = computed(() => username.value !== null)

  function hasPermission(key: string): boolean {
    return permissions.value.includes(key)
  }

  function login(user: string, grantAllForScaffold = true) {
    username.value = user
    permissions.value = grantAllForScaffold ? [...MOCK_ALL_PERMISSIONS] : []
  }

  function logout() {
    username.value = null
    permissions.value = []
  }

  return {
    username,
    permissions,
    isAuthenticated,
    hasPermission,
    login,
    logout,
  }
})
