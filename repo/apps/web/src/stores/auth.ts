import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import type { DataScope } from '@/utils/dataScope'

/** Matches backend `access.PermissionFullAccess`; grants all route/menu permission checks. */
export const PERMISSION_FULL_ACCESS = 'system.full_access'

const TOKEN_KEY = 'pharmaops_session_token'

function apiUrl(path: string): string {
  const base = import.meta.env.VITE_API_BASE_URL?.replace(/\/$/, '') ?? ''
  const p = path.startsWith('/') ? path : `/${path}`
  return `${base}${p}`
}

type Envelope = {
  code: string
  message: string
  requestId?: string
  data?: unknown
}

async function readEnvelope(res: Response): Promise<Envelope> {
  return (await res.json()) as Envelope
}

export const useAuthStore = defineStore('auth', () => {
  const userId = ref<string | null>(null)
  const username = ref<string | null>(null)
  const permissions = ref<string[]>([])
  const roles = ref<string[]>([])
  /** Data scopes from GET /auth/me — drives default institution/dept/team for creates. */
  const scopes = ref<DataScope[]>([])
  const loadingSession = ref(false)

  let sessionBootstrap: Promise<void> | null = null

  const isAuthenticated = computed(() => username.value !== null)

  function hasPermission(key: string): boolean {
    if (permissions.value.includes(PERMISSION_FULL_ACCESS)) {
      return true
    }
    return permissions.value.includes(key)
  }

  function clearLocal(): void {
    sessionStorage.removeItem(TOKEN_KEY)
    userId.value = null
    username.value = null
    permissions.value = []
    roles.value = []
    scopes.value = []
    sessionBootstrap = null
  }

  async function fetchMe(): Promise<void> {
    const token = sessionStorage.getItem(TOKEN_KEY)
    if (!token) {
      clearLocal()
      return
    }
    const res = await fetch(apiUrl('/api/v1/auth/me'), {
      headers: { Authorization: `Bearer ${token}` },
    })
    const env = await readEnvelope(res)
    if (!res.ok || env.code !== 'OK') {
      clearLocal()
      throw new Error(env.message?.trim() || 'Your session has expired. Please sign in again.')
    }
    const d = env.data as {
      id: string
      username: string
      roles?: string[]
      permissions?: string[]
      scopes?: DataScope[]
    }
    userId.value = d.id
    username.value = d.username
    roles.value = d.roles ?? []
    permissions.value = d.permissions ?? []
    scopes.value = Array.isArray(d.scopes) ? d.scopes : []
  }

  /** Call once before route guards; restores session from storage via GET /auth/me. */
  async function ensureSessionLoaded(): Promise<void> {
    if (!sessionBootstrap) {
      sessionBootstrap = (async () => {
        const token = sessionStorage.getItem(TOKEN_KEY)
        if (!token) {
          return
        }
        loadingSession.value = true
        try {
          await fetchMe()
        } catch {
          clearLocal()
        } finally {
          loadingSession.value = false
        }
      })()
    }
    await sessionBootstrap
  }

  async function login(user: string, password: string): Promise<void> {
    const res = await fetch(apiUrl('/api/v1/auth/login'), {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username: user, password }),
    })
    const env = await readEnvelope(res)
    if (!res.ok || env.code !== 'OK') {
      throw new Error(env.message?.trim() || 'Sign in failed. Check your username and password.')
    }
    const d = env.data as { token: string }
    sessionStorage.setItem(TOKEN_KEY, d.token)
    await fetchMe()
  }

  async function logout(): Promise<void> {
    const token = sessionStorage.getItem(TOKEN_KEY)
    if (token) {
      try {
        await fetch(apiUrl('/api/v1/auth/logout'), {
          method: 'POST',
          headers: { Authorization: `Bearer ${token}` },
        })
      } catch {
        /* ignore network errors */
      }
    }
    clearLocal()
  }

  return {
    userId,
    username,
    permissions,
    roles,
    scopes,
    loadingSession,
    isAuthenticated,
    hasPermission,
    ensureSessionLoaded,
    login,
    logout,
  }
})
