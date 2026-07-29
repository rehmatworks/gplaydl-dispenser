import { AccountsTable } from "@/components/AccountsTable"
import { CommandBlock } from "@/components/CommandBlock"
import { Logo } from "@/components/Logo"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { api, type Account, type AppRelease } from "@/lib/api"
import { AlertTriangle, Download, LogOut, RefreshCw, ShieldCheck } from "lucide-react"
import { useAuth } from "@/lib/auth"
import { useCallback, useEffect, useState } from "react"
import { toast } from "sonner"

export default function Dashboard() {
  const { user, setUser } = useAuth()
  const [accounts, setAccounts] = useState<Account[]>([])
  const [release, setRelease] = useState<AppRelease | null>(null)
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [loadError, setLoadError] = useState(false)

  const refresh = useCallback(async (quiet = false) => {
    if (!quiet) setRefreshing(true)
    try {
      const accountResult = await api.accounts()
      setAccounts(accountResult.accounts)
      setLoadError(false)
    } catch {
      setLoadError(true)
      if (!quiet) toast.error("Could not refresh the dashboard")
    } finally {
      setLoading(false)
      setRefreshing(false)
    }
  }, [])

  useEffect(() => {
    refresh(true)
    api.appLatest().then(setRelease).catch(() => {})
    const interval = setInterval(() => refresh(true), 30_000)
    return () => clearInterval(interval)
  }, [refresh])

  async function disconnect() {
    await api.logout().catch(() => {})
    setUser(null)
    window.location.assign("/")
  }

  const needsAttention = accounts.filter(
    (account) => account.status !== "active" || account.failureCount >= 5
  ).length

  return (
    <div className="min-h-dvh">
      <header className="sticky top-0 z-40 border-b border-border/60 bg-background/80 backdrop-blur-xl">
        <div className="mx-auto flex h-16 max-w-6xl items-center justify-between gap-3 px-4 sm:px-6">
          <a href="/">
            <Logo />
          </a>
          <div className="flex items-center gap-2">
            <span className="hidden text-sm text-muted-foreground sm:block">{user?.label}</span>
            <Button
              variant="ghost"
              size="sm"
              onClick={disconnect}
              className="rounded-xl"
              aria-label="Disconnect this browser dashboard"
            >
              <LogOut className="size-4" />
              <span className="hidden sm:inline">Disconnect</span>
            </Button>
          </div>
        </div>
      </header>

      <main className="mx-auto max-w-6xl space-y-8 px-4 py-8 sm:px-6 sm:py-10">
        <div className="flex flex-wrap items-end justify-between gap-4">
          <div>
            <p className="text-sm font-medium text-primary">Paired with {user?.label || "Android app"}</p>
            <h1 className="mt-1 text-3xl font-bold">Google accounts</h1>
            <p className="mt-2 max-w-xl text-sm text-muted-foreground">
              These accounts are private to you and power your own gplaydl downloads. Add or
              refresh them in the Android app.
            </p>
          </div>
          <Button
            variant="outline"
            onClick={() => refresh()}
            disabled={refreshing}
            className="glass rounded-xl"
          >
            <RefreshCw className={`size-4 ${refreshing ? "animate-spin" : ""}`} />
            Refresh
          </Button>
        </div>

        {loadError && (
          <div className="flex items-center gap-3 rounded-2xl border border-chart-4/30 bg-chart-4/8 p-4 text-sm">
            <AlertTriangle className="size-5 shrink-0 text-chart-4" />
            <span className="flex-1 text-muted-foreground">
              Dashboard data could not be loaded. Your existing sharing settings have not changed.
            </span>
            <Button variant="outline" size="sm" onClick={() => refresh()} className="rounded-xl">
              Retry
            </Button>
          </div>
        )}

        <div className="grid gap-3 sm:grid-cols-2">
          {[
            { icon: ShieldCheck, label: "Your accounts", value: loading ? "—" : accounts.length },
            {
              icon: AlertTriangle,
              label: "Need attention",
              value: loading ? "—" : needsAttention
            }
          ].map((item) => (
            <Card key={item.label} className="glass rounded-2xl border-0">
              <CardContent className="flex items-center gap-4 p-5">
                <div className="flex size-10 items-center justify-center rounded-xl bg-primary/10">
                  <item.icon className="size-4 text-primary" />
                </div>
                <div>
                  <p className="text-2xl font-bold">{item.value}</p>
                  <p className="text-xs text-muted-foreground">{item.label}</p>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>

        <AccountsTable
          accounts={accounts}
          onDelete={(id) => setAccounts((current) => current.filter((account) => account.id !== id))}
        />

        <details className="glass group rounded-3xl">
          <summary className="cursor-pointer list-none p-5 font-semibold sm:p-6">
            Use gplaydl and update the app
            <span className="float-right text-muted-foreground transition-transform group-open:rotate-180">
              ↓
            </span>
          </summary>
          <div className="space-y-5 border-t border-border p-5 sm:p-6">
            <p className="text-sm leading-relaxed text-muted-foreground">
              Link this account to gplaydl once with the pairing code from the app&apos;s Link
              gplaydl screen, then download. Pass{" "}
              <code className="rounded bg-muted px-1.5 py-0.5">--email</code> if you added several
              accounts.
            </p>
            <CommandBlock label="Link once" command="gplaydl link" />
            <CommandBlock
              label="Then download"
              command="gplaydl download com.google.android.calculator"
            />
            {release?.url && (
              <Button asChild variant="outline" className="glass rounded-xl">
                <a href={release.url}>
                  <Download className="size-4" />
                  Download Authenticator {release.version}
                </a>
              </Button>
            )}
          </div>
        </details>
      </main>
    </div>
  )
}
