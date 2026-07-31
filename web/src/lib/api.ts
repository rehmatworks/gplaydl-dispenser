export interface User {
  id: string
  createdAt: string
  kind: "device"
  label: string
  isAdmin: boolean
}

export interface Account {
  id: string
  ownerId: string
  email: string
  status: "active" | "flagged" | "disabled"
  lastUsedAt: string | null
  failureCount: number
  mintCount: number
  createdAt: string
  source: "web" | "app"
  lastSyncedAt: string | null
  proxyConfigured: boolean
  proxyTestStatus?: "passed" | "failed"
  proxyTestedAt?: string
  proxyFailureCount: number
  lastProxyFailureAt?: string
}

export interface AppRelease {
  version: string
  versionCode: number
  url: string
  sha256: string
}

export interface ProxySettings {
  proxyConfigured: boolean
}

export interface ProxyBackfillResult {
  targeted: number
  updated: number
  passed: number
  failed: number
  errors: number
}

export class ApiError extends Error {
  status: number
  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(path, {
    headers: { "Content-Type": "application/json" },
    credentials: "same-origin",
    ...init
  })

  const body = await res.json().catch(() => ({}))
  if (!res.ok) {
    throw new ApiError(res.status, body.error ?? `Request failed (${res.status})`)
  }
  return body as T
}

export const api = {
  logout: () => request<{ status: string }>("/api/v1/logout", { method: "POST" }),

  me: () => request<{ user: User }>("/api/v1/me"),

  accounts: () => request<{ accounts: Account[] }>("/api/v1/accounts"),

  deleteAccount: (id: string) =>
    request<{ status: string }>(`/api/v1/accounts/${id}`, { method: "DELETE" }),

  appLatest: () => request<AppRelease>("/api/v1/app/latest"),

  adminSettings: () => request<ProxySettings>("/api/v1/admin/settings"),

  updateProxy: (proxyTemplate: string) =>
    request<ProxySettings>("/api/v1/admin/settings/proxy", {
      method: "PUT",
      body: JSON.stringify({ proxyTemplate })
    }),

  clearProxy: () =>
    request<ProxySettings>("/api/v1/admin/settings/proxy", {
      method: "DELETE"
    }),

  backfillProxies: () =>
    request<ProxyBackfillResult>("/api/v1/admin/settings/proxy/backfill", {
      method: "POST"
    }),

  /** Trades a code shown by the Android app for a browser session. */
  claimPairing: (code: string) =>
    request<{ user: User }>("/api/v1/pair/claim", {
      method: "POST",
      body: JSON.stringify({ code })
    })
}
