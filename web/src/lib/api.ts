export interface User {
  id: string
  createdAt: string
  kind: "device"
  label: string
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
  source: "web" | "app"
  sharedAt: string | null
  lastSyncedAt: string | null
}

export interface PoolStats {
  publicAccounts: number
  privateAccounts: number
  activeAccounts: number
  flaggedAccounts: number
  mints24h: number
  failures24h: number
  totalMints: number
  contributors: number
  sharedAccounts: number
}

export interface PublicStats {
  publicAccounts: number
  mints24h: number
  totalMints: number
  contributors: number
}

export interface AppRelease {
  version: string
  versionCode: number
  url: string
  sha256: string
}

export interface MintBucket {
  hour: string
  success: number
  failures: number
}

/** Must match CONSENT_VERSION in the Android app and the wording on this site. */
export const CONSENT_VERSION = "2026-07-27"

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

  addAccount: (email: string, aasToken: string, visibility: "public" | "private") =>
    request<{ account: Account }>("/api/v1/accounts", {
      method: "POST",
      // Sharing publicly is only accepted alongside the consent wording the
      // contributor agreed to; the dialog shows the same text.
      body: JSON.stringify({ email, aasToken, visibility, consentVersion: CONSENT_VERSION })
    }),

  updateAccount: (id: string, patch: { visibility: "public" | "private" }) =>
    request<{ account: Account }>(`/api/v1/accounts/${id}`, {
      method: "PATCH",
      body: JSON.stringify({
        ...patch,
        ...(patch.visibility === "public" ? { consentVersion: CONSENT_VERSION } : {})
      })
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

  timeline: () => request<{ timeline: MintBucket[] }>("/api/v1/timeline"),

  appLatest: () => request<AppRelease>("/api/v1/app/latest"),

  /** Trades a code shown by the Android app for a browser session. */
  claimPairing: (code: string) =>
    request<{ user: User }>("/api/v1/pair/claim", {
      method: "POST",
      body: JSON.stringify({ code })
    })
}
