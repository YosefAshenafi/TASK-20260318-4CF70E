import { describe, expect, it } from 'vitest'

import { pickDefaultScopeContext, type DataScope } from '@/utils/dataScope'

describe('pickDefaultScopeContext', () => {
  it('returns null when no scopes exist', () => {
    expect(pickDefaultScopeContext([])).toBeNull()
  })

  it('prefers widest scope for same institution', () => {
    const scopes: DataScope[] = [
      { id: 's2', scopeKey: 'inst-a-dept', institutionId: 'inst-a', departmentId: 'dept-a' },
      { id: 's1', scopeKey: 'inst-a-root', institutionId: 'inst-a' },
    ]
    expect(pickDefaultScopeContext(scopes)).toEqual({
      institutionId: 'inst-a',
      departmentId: undefined,
      teamId: undefined,
    })
  })
})

