/**
 * Readable labels for technical tokens (snake_case, dotted.permission.codes).
 * Matches product expectation: show names, not raw system identifiers, in the UI.
 */

function titleCaseSegment(segment: string): string {
  const s = segment.replace(/_/g, ' ').trim()
  if (!s) return ''
  return s
    .split(/\s+/)
    .map((w) => w.charAt(0).toUpperCase() + w.slice(1).toLowerCase())
    .join(' ')
}

/** Turns identifiers like `in_progress`, `recruitment.view`, or `system.full_access` into readable text. */
export function humanizeTechnicalLabel(value: string): string {
  if (!value) return ''
  if (value.includes('.')) {
    return value
      .split('.')
      .map((seg) => titleCaseSegment(seg))
      .filter(Boolean)
      .join(' — ')
  }
  return titleCaseSegment(value)
}

const CASE_STATUS: Record<string, string> = {
  submitted: 'Submitted',
  assigned: 'Assigned',
  in_progress: 'In progress',
  pending_review: 'Pending review',
  closed: 'Closed',
}

export function caseStatusLabel(code: string): string {
  return CASE_STATUS[code] ?? humanizeTechnicalLabel(code)
}

export function positionStatusLabel(code: string): string {
  if (code === 'open') return 'Open'
  if (code === 'closed') return 'Closed'
  return humanizeTechnicalLabel(code)
}

export function qualificationStatusLabel(code: string): string {
  if (code === 'active') return 'Active'
  if (code === 'inactive') return 'Inactive'
  return humanizeTechnicalLabel(code)
}

/** Scope keys like `inst:acme-dept` → easier to scan (still may contain technical bits for admins). */
export function humanizeScopeKeyLabel(key: string): string {
  if (!key) return ''
  return key.replace(/_/g, ' ').replace(/:/g, ' — ')
}
