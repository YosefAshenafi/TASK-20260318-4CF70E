/** Same storage key as `stores/auth.ts` (avoid importing Pinia here). */
const TOKEN_KEY = 'pharmaops_session_token'

function apiUrl(path: string): string {
  const base = import.meta.env.VITE_API_BASE_URL?.replace(/\/$/, '') ?? ''
  const p = path.startsWith('/') ? path : `/${path}`
  return `${base}${p}`
}

export type ApiEnvelope<T> = {
  code: string
  message: string
  requestId?: string
  data?: T
}

async function parseEnvelope<T>(res: Response): Promise<T> {
  const env = (await res.json()) as ApiEnvelope<T>
  if (!res.ok || env.code !== 'OK') {
    throw new Error(env.message || `HTTP ${res.status}`)
  }
  return env.data as T
}

function baseHeaders(): Record<string, string> {
  const h: Record<string, string> = {}
  const token = sessionStorage.getItem(TOKEN_KEY)
  if (token) {
    h.Authorization = `Bearer ${token}`
  }
  return h
}

export async function apiGet<T>(path: string): Promise<T> {
  const res = await fetch(apiUrl(path), { headers: baseHeaders() })
  return parseEnvelope<T>(res)
}

export async function apiPost<T>(path: string, body: unknown): Promise<T> {
  const res = await fetch(apiUrl(path), {
    method: 'POST',
    headers: { ...baseHeaders(), 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  return parseEnvelope<T>(res)
}

export async function apiPatch<T>(path: string, body: unknown): Promise<T> {
  const res = await fetch(apiUrl(path), {
    method: 'PATCH',
    headers: { ...baseHeaders(), 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  return parseEnvelope<T>(res)
}

export async function apiDelete<T>(path: string): Promise<T> {
  const res = await fetch(apiUrl(path), { method: 'DELETE', headers: baseHeaders() })
  return parseEnvelope<T>(res)
}
