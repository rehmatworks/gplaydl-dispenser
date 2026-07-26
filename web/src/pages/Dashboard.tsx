import { AccountsTable } from "@/components/AccountsTable"
import { AddAccountDialog } from "@/components/AddAccountDialog"
import { ApiKeyCard } from "@/components/ApiKeyCard"
import { CommandBlock } from "@/components/CommandBlock"
import { DeviceCard } from "@/components/DeviceCard"
import { Logo } from "@/components/Logo"
import { MintChart } from "@/components/MintChart"
import { ShareWithCommunity } from "@/components/ShareWithCommunity"
import { StatsCards } from "@/components/StatsCards"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import {
  api,
  ApiError,
  type Account,
  type AppRelease,
  type MintBucket,
  type PoolStats
} from "@/lib/api"
import { useAuth } from "@/lib/auth"
import { LogOut, MailWarning } from "lucide-react"
import { useCallback, useEffect, useState } from "react"
import { Link, useNavigate } from "react-router-dom"
import { toast } from "sonner"

export default function Dashboard() {
  const { user, setUser } = useAuth()
  const navigate = useNavigate()

  const [accounts, setAccounts] = useState<Account[]>([])
  const [stats, setStats] = useState<PoolStats | null>(null)
  const [timeline, setTimeline] = useState<MintBucket[]>([])
  const [release, setRelease] = useState<AppRelease | null>(null)

  const refresh = useCallback(() => {
    api.accounts().then((res) => setAccounts(res.accounts)).catch(() => {})
    api.stats().then(setStats).catch(() => {})
    api.timeline().then((res) => setTimeline(res.timeline)).catch(() => {})
  }, [])

  useEffect(() => {
    refresh()
    api.appLatest().then(setRelease).catch(() => {})
    const interval = setInterval(refresh, 30_000)
    return () => clearInterval(interval)
  }, [refresh])

  const [resending, setResending] = useState(false)

  async function resendVerification() {
    setResending(true)
    try {
      await api.resendVerification()
      toast.success("Verification email sent — check your inbox")
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Could not send email")
    } finally {
      setResending(false)
    }
  }

  async function logout() {
    await api.logout().catch(() => {})
    sessionStorage.removeItem("freshApiKey")
    setUser(null)
    navigate("/")
  }

  const isDevice = user?.kind === "device"
  const origin = window.location.origin
  const firstPrivate = accounts.find((a) => a.visibility === "private")

  return (
    <div className="min-h-dvh">
      <header className="glass-strong sticky top-0 z-40">
        <div className="mx-auto flex h-16 max-w-6xl items-center justify-between px-6">
          <Link to="/">
            <Logo />
          </Link>
          <div className="flex items-center gap-4">
            <span className="hidden text-sm text-muted-foreground sm:block">
              {isDevice ? user?.label : user?.email}
            </span>
            <Button variant="ghost" size="sm" onClick={logout} className="rounded-xl">
              <LogOut className="size-4" /> Sign out
            </Button>
          </div>
        </div>
      </header>

      <main className="mx-auto max-w-6xl space-y-8 px-6 py-10">
        {/* Device users have no inbox, so the verification nag does not apply. */}
        {user && !isDevice && !user.emailVerified && (
          <div className="animate-fade-up glass flex flex-wrap items-center gap-3 rounded-2xl border border-aurora-teal/30 px-5 py-4">
            <MailWarning className="size-5 shrink-0 text-aurora-teal" />
            <p className="flex-1 text-sm text-muted-foreground">
              <span className="font-medium text-foreground">Verify your email</span> to start
              adding accounts. We sent a link to {user.email}.
            </p>
            <Button
              size="sm"
              variant="outline"
              onClick={resendVerification}
              disabled={resending}
              className="rounded-xl"
            >
              {resending ? "Sending…" : "Resend email"}
            </Button>
          </div>
        )}

        <div className="animate-fade-up flex flex-wrap items-end justify-between gap-4">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Your account pool</h1>
            <p className="mt-1 text-sm text-muted-foreground">
              {isDevice
                ? "Manage what your phone shares, and watch the dispenser at work."
                : "Manage Google accounts, sharing, and watch the dispenser at work."}
            </p>
          </div>
          {!isDevice && (
            <AddAccountDialog
              onAdded={(account) => {
                setAccounts((prev) => [account, ...prev])
                refresh()
              }}
            />
          )}
        </div>

        {user && isDevice && (
          <div className="animate-fade-up">
            <DeviceCard user={user} release={release} />
          </div>
        )}

        <div className="animate-fade-up [animation-delay:80ms]">
          <StatsCards stats={stats} />
        </div>

        <div className="animate-fade-up [animation-delay:120ms]">
          <ShareWithCommunity accounts={accounts} onChanged={refresh} />
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

        <section className="animate-fade-up space-y-4 [animation-delay:320ms]">
          <h2 className="font-display text-xl font-semibold">Downloading with gplaydl</h2>
          <Card className="glass rounded-2xl border-0">
            <CardContent className="space-y-5 p-6">
              <CommandBlock
                label="Community pool — no key needed"
                command={`gplaydl download com.instagram.android -d ${origin}/api/auth`}
              />
              <CommandBlock
                label={
                  firstPrivate
                    ? `Your own account (${firstPrivate.email}) — replace YOUR_KEY with the API key above`
                    : "One of your own accounts — replace YOUR_KEY and the email"
                }
                command={`gplaydl download com.your.app -d "${origin}/api/auth?api_key=YOUR_KEY&email=${
                  firstPrivate?.email ?? "you@gmail.com"
                }"`}
              />
              <p className="text-xs leading-relaxed text-muted-foreground">
                Pinning by email is how you reach apps you have purchased: the dispenser only
                looks at accounts your API key owns, and never puts a private account into the
                community rotation.
              </p>
            </CardContent>
          </Card>
        </section>
      </main>
    </div>
  )
}
