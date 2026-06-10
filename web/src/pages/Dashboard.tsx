import { AccountsTable } from "@/components/AccountsTable"
import { AddAccountDialog } from "@/components/AddAccountDialog"
import { ApiKeyCard } from "@/components/ApiKeyCard"
import { Logo } from "@/components/Logo"
import { MintChart } from "@/components/MintChart"
import { StatsCards } from "@/components/StatsCards"
import { Button } from "@/components/ui/button"
import { api, type Account, type MintBucket, type PoolStats } from "@/lib/api"
import { useAuth } from "@/lib/auth"
import { LogOut } from "lucide-react"
import { useCallback, useEffect, useState } from "react"
import { Link, useNavigate } from "react-router-dom"

export default function Dashboard() {
  const { user, setUser } = useAuth()
  const navigate = useNavigate()

  const [accounts, setAccounts] = useState<Account[]>([])
  const [stats, setStats] = useState<PoolStats | null>(null)
  const [timeline, setTimeline] = useState<MintBucket[]>([])

  const refresh = useCallback(() => {
    api.accounts().then((res) => setAccounts(res.accounts)).catch(() => {})
    api.stats().then(setStats).catch(() => {})
    api.timeline().then((res) => setTimeline(res.timeline)).catch(() => {})
  }, [])

  useEffect(() => {
    refresh()
    const interval = setInterval(refresh, 30_000)
    return () => clearInterval(interval)
  }, [refresh])

  async function logout() {
    await api.logout().catch(() => {})
    sessionStorage.removeItem("freshApiKey")
    setUser(null)
    navigate("/")
  }

  return (
    <div className="min-h-dvh">
      <header className="glass-strong sticky top-0 z-40">
        <div className="mx-auto flex h-16 max-w-6xl items-center justify-between px-6">
          <Link to="/">
            <Logo />
          </Link>
          <div className="flex items-center gap-4">
            <span className="hidden text-sm text-muted-foreground sm:block">{user?.email}</span>
            <Button variant="ghost" size="sm" onClick={logout} className="rounded-xl">
              <LogOut className="size-4" /> Sign out
            </Button>
          </div>
        </div>
      </header>

      <main className="mx-auto max-w-6xl space-y-8 px-6 py-10">
        <div className="animate-fade-up flex flex-wrap items-end justify-between gap-4">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Your account pool</h1>
            <p className="mt-1 text-sm text-muted-foreground">
              Manage Google accounts, sharing, and watch the dispenser at work.
            </p>
          </div>
          <AddAccountDialog
            onAdded={(account) => {
              setAccounts((prev) => [account, ...prev])
              refresh()
            }}
          />
        </div>

        <div className="animate-fade-up [animation-delay:80ms]">
          <StatsCards stats={stats} />
        </div>

        <div className="animate-fade-up grid gap-5 lg:grid-cols-3 [animation-delay:160ms]">
          <div className="lg:col-span-2">
            <MintChart timeline={timeline} />
          </div>
          <ApiKeyCard />
        </div>

        <section className="animate-fade-up space-y-4 [animation-delay:240ms]">
          <h2 className="font-display text-xl font-semibold">
            Google accounts{" "}
            <span className="text-sm font-normal text-muted-foreground">
              ({accounts.length})
            </span>
          </h2>
          <AccountsTable
            accounts={accounts}
            onChange={(updated) => {
              setAccounts((prev) => prev.map((a) => (a.id === updated.id ? updated : a)))
              refresh()
            }}
            onDelete={(id) => {
              setAccounts((prev) => prev.filter((a) => a.id !== id))
              refresh()
            }}
          />
        </section>
      </main>
    </div>
  )
}
