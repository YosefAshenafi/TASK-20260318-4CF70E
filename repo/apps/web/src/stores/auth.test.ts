import { beforeEach, describe, expect, it } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

import { PERMISSION_FULL_ACCESS, useAuthStore } from '@/stores/auth'

describe('auth store permissions', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('grants explicit permission when present', () => {
    const auth = useAuthStore()
    auth.permissions = ['recruitment.view']
    expect(auth.hasPermission('recruitment.view')).toBe(true)
    expect(auth.hasPermission('cases.manage')).toBe(false)
  })

  it('grants any permission with full access', () => {
    const auth = useAuthStore()
    auth.permissions = [PERMISSION_FULL_ACCESS]
    expect(auth.hasPermission('system.rbac')).toBe(true)
    expect(auth.hasPermission('compliance.manage')).toBe(true)
  })
})

