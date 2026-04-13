/** One row from GET /auth/me `scopes` (matches backend `meScopeDTO`). */
export type DataScope = {
  id: string
  scopeKey: string
  institutionId: string
  departmentId?: string
  teamId?: string
}

export type ScopeCreateContext = {
  institutionId: string
  departmentId?: string
  teamId?: string
}

/**
 * Pick a default institution (and optional dept/team) for create flows.
 * Prefers the widest scope per institution (institution-only before department/team),
 * then first institution id lexicographically.
 */
export function pickDefaultScopeContext(scopes: DataScope[]): ScopeCreateContext | null {
  if (!scopes.length) {
    return null
  }
  const sorted = [...scopes].sort((a, b) => {
    if (a.institutionId !== b.institutionId) {
      return a.institutionId.localeCompare(b.institutionId)
    }
    const narrow = (s: DataScope) => (s.departmentId ? 1 : 0) + (s.teamId ? 1 : 0)
    return narrow(a) - narrow(b)
  })
  const pick = sorted[0]
  return {
    institutionId: pick.institutionId,
    departmentId: pick.departmentId,
    teamId: pick.teamId,
  }
}
