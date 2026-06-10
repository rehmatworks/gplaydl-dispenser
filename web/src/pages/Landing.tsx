import { Logo } from "@/components/Logo"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { api, type PublicStats } from "@/lib/api"
import { useAuth } from "@/lib/auth"
import { ArrowRight, Github, Globe2, Lock, RefreshCcw, Shield, Users, Zap } from "lucide-react"
import { useEffect, useState } from "react"
import { Link } from "react-router-dom"

const features = [
  {
    icon: Users,
    title: "Community account pools",
    body: "Add your Google accounts and choose: share them with the open-source community in the public pool, or keep them private for your own tooling."
  },
  {
    icon: RefreshCcw,
    title: "Atomic LRU rotation",
    body: "Accounts rotate through Postgres with FOR UPDATE SKIP LOCKED — concurrent requests claim distinct accounts with zero contention, surviving restarts and replicas."
  },
  {
    icon: Lock,
    title: "Encrypted at rest",
    body: "AAS tokens are sealed with AES-256-GCM before they ever touch the database. Plaintext tokens exist only in memory, only during a mint."
  },
  {
    icon: Zap,
    title: "Built for concurrency",
    body: "A single Go binary minting tokens in parallel with bounded handshake concurrency, connection pooling, and graceful failover between accounts."
  },
  {
    icon: Shield,
    title: "Self-healing pool",
    body: "Dead credentials are flagged automatically after repeated failures and silently drop out of rotation — no more cycling through broken accounts."
  },
  {
    icon: Globe2,
    title: "Aurora Store compatible",
    body: "Drop-in replacement for the original dispenser API. Point any Aurora Store-compatible client at /api/auth and it just works."
  }
]

export default function Landing() {
  const { user } = useAuth()
  const [stats, setStats] = useState<PublicStats | null>(null)
  const origin = window.location.origin

  useEffect(() => {
    api.publicStats().then(setStats).catch(() => {})
  }, [])

  return (
    <div className="relative">
      <div className="aurora-bg" />

      {/* Nav */}
      <header className="glass-strong fixed inset-x-0 top-0 z-50">
        <div className="mx-auto flex h-16 max-w-6xl items-center justify-between px-6">
          <Link to="/">
            <Logo />
          </Link>
          <nav className="hidden items-center gap-8 text-sm text-muted-foreground md:flex">
            <a href="#features" className="transition-colors hover:text-foreground">
              Features
            </a>
            <a href="#api" className="transition-colors hover:text-foreground">
              API
            </a>
          </nav>
          <div className="flex items-center gap-3">
            {user ? (
              <Button asChild className="btn-aurora rounded-xl">
                <Link to="/dashboard">
                  Dashboard <ArrowRight className="size-4" />
                </Link>
              </Button>
            ) : (
              <>
                <Button asChild variant="ghost" className="rounded-xl">
                  <Link to="/login">Sign in</Link>
                </Button>
                <Button asChild className="btn-aurora rounded-xl">
                  <Link to="/register">Get started</Link>
                </Button>
              </>
            )}
          </div>
        </div>
      </header>

      {/* Hero */}
      <section className="relative px-6 pb-24 pt-44">
        <div className="grid-dots absolute inset-0" aria-hidden />
        <div className="relative mx-auto max-w-4xl text-center">
          <Badge
            variant="outline"
            className="animate-fade-up mb-6 gap-2 rounded-full border-aurora-teal/30 bg-aurora-teal/5 px-4 py-1.5 text-aurora-teal"
          >
            <span className="size-1.5 animate-pulse rounded-full bg-aurora-teal" />
            Open source · GPL-3.0
          </Badge>

          <h1 className="animate-fade-up text-balance text-5xl font-bold leading-[1.05] tracking-tight md:text-7xl">
            Google Play tokens,
            <br />
            <span className="text-aurora">minted by the community.</span>
          </h1>

          <p className="animate-fade-up mx-auto mt-6 max-w-2xl text-pretty text-lg text-muted-foreground [animation-delay:120ms]">
            A high-concurrency token dispenser written in Go. Pool your Google accounts with the
            open-source community — or keep them private — and mint anonymous Play Store auth
            tokens at scale.
          </p>

          <div className="animate-fade-up mt-10 flex flex-wrap items-center justify-center gap-4 [animation-delay:240ms]">
            <Button asChild size="lg" className="btn-aurora h-12 rounded-2xl px-8 text-base">
              <Link to="/register">
                Add your accounts <ArrowRight className="size-4" />
              </Link>
            </Button>
            <Button
              asChild
              size="lg"
              variant="outline"
              className="glass h-12 rounded-2xl px-8 text-base"
            >
              <a href="#api">Use the API</a>
            </Button>
          </div>

          {/* Live stats */}
          <div className="animate-fade-up mx-auto mt-16 grid max-w-2xl grid-cols-3 gap-4 [animation-delay:360ms]">
            {[
              { label: "Public accounts", value: stats?.publicAccounts },
              { label: "Mints (24h)", value: stats?.mints24h },
              { label: "Tokens minted", value: stats?.totalMints }
            ].map((s) => (
              <Card key={s.label} className="glass card-hover rounded-2xl border-0 py-0">
                <CardContent className="px-4 py-5 text-center">
                  <div className="font-display text-3xl font-bold text-aurora">
                    {s.value ?? "—"}
                  </div>
                  <div className="mt-1 text-xs uppercase tracking-widest text-muted-foreground">
                    {s.label}
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      </section>

      {/* Features */}
      <section id="features" className="px-6 py-24">
        <div className="mx-auto max-w-6xl">
          <h2 className="text-center text-3xl font-bold tracking-tight md:text-4xl">
            Robust by design
          </h2>
          <p className="mx-auto mt-3 max-w-xl text-center text-muted-foreground">
            Everything the original NodeJS dispenser did, rebuilt on Go + PostgreSQL with
            multi-user pools.
          </p>

          <div className="mt-14 grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
            {features.map((f) => (
              <Card key={f.title} className="glass card-hover rounded-2xl border-0">
                <CardContent className="p-6">
                  <div className="mb-4 flex size-11 items-center justify-center rounded-xl bg-gradient-to-br from-aurora-teal/20 to-aurora-violet/20 ring-1 ring-aurora-teal/20">
                    <f.icon className="size-5 text-aurora-teal" />
                  </div>
                  <h3 className="font-display text-lg font-semibold">{f.title}</h3>
                  <p className="mt-2 text-sm leading-relaxed text-muted-foreground">{f.body}</p>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      </section>

      {/* API */}
      <section id="api" className="px-6 py-24">
        <div className="mx-auto max-w-3xl">
          <h2 className="text-center text-3xl font-bold tracking-tight md:text-4xl">
            One endpoint. <span className="text-aurora">Zero setup.</span>
          </h2>
          <p className="mx-auto mt-3 max-w-xl text-center text-muted-foreground">
            Anonymous requests draw from the public pool. Pass your API key to use your private
            accounts.
          </p>

          <Card className="glass-strong mt-10 overflow-hidden rounded-2xl border-0">
            <div className="flex items-center gap-2 border-b border-border px-5 py-3">
              <span className="size-3 rounded-full bg-destructive/60" />
              <span className="size-3 rounded-full bg-chart-4/60" />
              <span className="size-3 rounded-full bg-aurora-teal/60" />
              <span className="ml-3 font-mono text-xs text-muted-foreground">terminal</span>
            </div>
            <CardContent className="overflow-x-auto p-5 font-mono text-sm leading-7">
              <div className="text-muted-foreground"># Anonymous token from the public pool</div>
              <div>
                <span className="text-aurora-teal">curl</span> {origin}
                <span className="text-aurora-pink">/api/auth</span>
              </div>
              <div className="mt-4 text-muted-foreground"># Mint with your private accounts</div>
              <div>
                <span className="text-aurora-teal">curl</span> -H{" "}
                <span className="text-chart-4">"X-Api-Key: $KEY"</span> "{origin}
                <span className="text-aurora-pink">/api/auth?pool=private</span>"
              </div>
              <div className="mt-4 text-muted-foreground"># Full bundle with custom device</div>
              <div>
                <span className="text-aurora-teal">curl</span> -X POST -d{" "}
                <span className="text-chart-4">@device.json</span> {origin}
                <span className="text-aurora-pink">/api/auth</span>
              </div>
            </CardContent>
          </Card>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t border-border px-6 py-10">
        <div className="mx-auto flex max-w-6xl flex-col items-center justify-between gap-4 md:flex-row">
          <Logo className="opacity-70" />
          <p className="text-sm text-muted-foreground">
            GPL-3.0 — a community rewrite of Aurora Dispenser in Go.
          </p>
          <a
            href="https://github.com/rehmatworks/gplaydl-dispenser"
            target="_blank"
            rel="noreferrer"
            className="text-muted-foreground transition-colors hover:text-foreground"
            aria-label="GitHub"
          >
            <Github className="size-5" />
          </a>
        </div>
      </footer>
    </div>
  )
}
