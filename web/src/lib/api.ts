export interface User {
  id: string
  email: string
  createdAt: string
}

export interface Account {
  id: string
  ownerId: string
  email: string
  visibility: "public" | "private"
  status: "active" | "flagged" | "disabled"
  lastUsedAt: string | null
  failureCount: number
  mintCount: number
  createdAt: string
}

export interface PoolStats {
  publicAccounts: number
  privateAccounts: number
  activeAccounts: number
  flaggedAccounts: number
  mints24h: number
  failures24h: number
  totalMints: number
}

export interface PublicStats {
  publicAccounts: number
  mints24h: number
  totalMints: number
}

export interface MintBucket {
  hour: string
  success: number
  failures: number
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
  register: (email: string, password: string) =>
    request<{ user: User; apiKey: string }>("/api/v1/register", {
      method: "POST",
      body: JSON.stringify({ email, password })
    }),

  login: (email: string, password: string) =>
    request<{ user: User }>("/api/v1/login", {
      method: "POST",
      body: JSON.stringify({ email, password })
    }),

  logout: () => request<{ status: string }>("/api/v1/logout", { method: "POST" }),

  me: () => request<{ user: User }>("/api/v1/me"),

  rotateApiKey: () => request<{ apiKey: string }>("/api/v1/me/api-key", { method: "POST" }),

  accounts: () => request<{ accounts: Account[] }>("/api/v1/accounts"),

  addAccount: (email: string, aasToken: string, visibility: "public" | "private") =>
    request<{ account: Account }>("/api/v1/accounts", {
      method: "POST",
      body: JSON.stringify({ email, aasToken, visibility })
    }),

  updateAccount: (id: string, patch: { visibility?: string; status?: string }) =>
    request<{ account: Account }>(`/api/v1/accounts/${id}`, {
      method: "PATCH",
      body: JSON.stringify(patch)
    }),

  deleteAccount: (id: string) =>
    request<{ status: string }>(`/api/v1/accounts/${id}`, { method: "DELETE" }),

  testAccount: (id: string) =>
    request<{ success: boolean; error: string; durationMs: number }>(
      `/api/v1/accounts/${id}/test`,
      { method: "POST" }
    ),

  stats: () => request<PoolStats>("/api/v1/stats"),

  publicStats: () => request<PublicStats>("/api/v1/public-stats"),

  timeline: () => request<{ timeline: MintBucket[] }>("/api/v1/timeline")
}
