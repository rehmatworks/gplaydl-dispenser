import { CommandBlock } from "@/components/CommandBlock"
import { Logo } from "@/components/Logo"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { api, type AppRelease, type PublicStats } from "@/lib/api"
import {
  ArrowRight,
  CheckCircle2,
  Download,
  Globe2,
  KeyRound,
  Lock,
  ShieldAlert,
  Smartphone,
  Users
} from "lucide-react"
import { useEffect, useState } from "react"
import { QRCodeSVG } from "qrcode.react"

const fallbackDownload = "/downloads/gplaydl-authenticator-latest.apk"

export default function Landing() {
  const [release, setRelease] = useState<AppRelease | null>(null)
  const [stats, setStats] = useState<PublicStats | null>(null)

  useEffect(() => {
    api.appLatest().then(setRelease).catch(() => {})
    api.publicStats().then(setStats).catch(() => {})
  }, [])

  // The browser resolves the initial hash before React has rendered the
  // target, so /#terms and /#privacy from the app would land on the hero.
  useEffect(() => {
    const id = window.location.hash.slice(1)
    if (id) document.getElementById(id)?.scrollIntoView()
  }, [])

  const downloadUrl = release?.url || fallbackDownload
  const origin = window.location.origin

  return (
    <div className="min-h-dvh overflow-hidden">
      <div className="aurora-bg" />

      <header className="sticky top-0 z-40 border-b border-border/60 bg-background/80 backdrop-blur-xl">
        <div className="mx-auto flex h-16 max-w-6xl items-center justify-between gap-3 px-4 sm:px-6">
          <a href="/" aria-label="gplaydl dispenser home">
            <Logo />
          </a>
          <div className="flex items-center gap-2">
            <a
              href="#how-it-works"
              className="hidden rounded-lg px-3 py-2 text-sm text-muted-foreground hover:text-foreground sm:block"
            >
              How it works
            </a>
            <Button asChild variant="outline" size="sm" className="glass rounded-xl">
              <a href="/pair">
                <Smartphone className="size-4" />
                Dashboard
              </a>
            </Button>
          </div>
        </div>
      </header>

      <main>
        <section className="grid-dots relative">
          <div className="mx-auto grid max-w-6xl items-center gap-12 px-6 py-20 md:grid-cols-[1fr_320px] md:py-28">
            <div className="animate-fade-up">
              <div className="mb-6 inline-flex items-center gap-2 rounded-full border border-primary/25 bg-primary/8 px-3 py-1.5 text-sm text-primary">
                <CheckCircle2 className="size-4" />
                Passwordless and community-run
              </div>
              <h1 className="max-w-3xl text-4xl font-bold leading-[1.08] tracking-tight sm:text-6xl">
                Download from Google Play with{" "}
                <span className="text-aurora">gplaydl</span>
              </h1>
              <p className="mt-6 max-w-2xl text-lg leading-relaxed text-muted-foreground">
                Install the Authenticator, add a spare Google account, and choose whether to
                help the public pool or keep the account private. No dispenser signup required.
              </p>
              <div className="mt-8 flex flex-col gap-3 sm:flex-row">
                <Button asChild size="lg" className="btn-aurora h-12 rounded-xl px-6">
                  <a href={downloadUrl}>
                    <Download className="size-5" />
                    Download Authenticator {release ? `v${release.version}` : ""}
                  </a>
                </Button>
                <Button asChild size="lg" variant="outline" className="glass h-12 rounded-xl px-6">
                  <a href="#quick-start">
                    Try gplaydl
                    <ArrowRight className="size-4" />
                  </a>
                </Button>
              </div>
              <p className="mt-4 text-xs text-muted-foreground">
                Android 7.0 or newer · signed APK ·{" "}
                {release?.sha256 ? `SHA-256 ${release.sha256.slice(0, 12)}…` : "latest release"}
              </p>
            </div>

            <Card className="glass-strong hidden rounded-3xl border-0 p-0 md:block">
              <CardContent className="flex flex-col items-center p-7 text-center">
                <div className="rounded-2xl bg-white p-4">
                  <QRCodeSVG value={downloadUrl} size={176} level="M" />
                </div>
                <p className="mt-5 font-semibold">Scan on your Android phone</p>
                <p className="mt-1 text-sm text-muted-foreground">
                  Downloads the same signed APK as the button.
                </p>
              </CardContent>
            </Card>
          </div>
        </section>

        <section className="border-y border-border/70 bg-card/25">
          <div className="mx-auto grid max-w-6xl grid-cols-3 divide-x divide-border/70 px-6 py-6 text-center">
            <div>
              <p className="text-2xl font-bold text-primary">{stats?.publicAccounts ?? "—"}</p>
              <p className="mt-1 text-xs text-muted-foreground">shared accounts</p>
            </div>
            <div>
              <p className="text-2xl font-bold">{stats?.contributors ?? "—"}</p>
              <p className="mt-1 text-xs text-muted-foreground">contributors</p>
            </div>
            <div>
              <p className="text-2xl font-bold">{stats?.mints24h ?? "—"}</p>
              <p className="mt-1 text-xs text-muted-foreground">Play sessions today</p>
            </div>
          </div>
        </section>

        <section id="how-it-works" className="mx-auto max-w-6xl px-6 py-20">
          <div className="max-w-2xl">
            <p className="text-sm font-semibold uppercase tracking-[0.18em] text-primary">
              Three simple steps
            </p>
            <h2 className="mt-3 text-3xl font-bold sm:text-4xl">The app handles the sensitive part</h2>
            <p className="mt-4 text-muted-foreground">
              Accounts are added on your phone. The website is only an optional paired dashboard.
            </p>
          </div>
          <div className="mt-10 grid gap-4 md:grid-cols-3">
            {[
              {
                icon: Smartphone,
                title: "1. Install the app",
                body: "Accept the clear sharing terms and sign in to Google inside its embedded setup screen."
              },
              {
                icon: Lock,
                title: "2. Choose sharing",
                body: "Make each account Community or Private before it is saved. You can change this later."
              },
              {
                icon: Globe2,
                title: "3. Use gplaydl",
                body: "Use the public pool without an account, or copy your private API key from the app."
              }
            ].map((step) => (
              <Card key={step.title} className="glass card-hover rounded-3xl border-0">
                <CardContent className="p-6">
                  <div className="flex size-11 items-center justify-center rounded-2xl bg-primary/12 ring-1 ring-primary/20">
                    <step.icon className="size-5 text-primary" />
                  </div>
                  <h3 className="mt-5 text-lg font-semibold">{step.title}</h3>
                  <p className="mt-2 text-sm leading-relaxed text-muted-foreground">{step.body}</p>
                </CardContent>
              </Card>
            ))}
          </div>
        </section>

        <section id="quick-start" className="mx-auto max-w-4xl px-6 pb-20">
          <Card className="glass-strong rounded-3xl border-0">
            <CardContent className="p-6 sm:p-8">
              <div className="flex items-start gap-4">
                <div className="flex size-11 shrink-0 items-center justify-center rounded-2xl bg-primary/12">
                  <KeyRound className="size-5 text-primary" />
                </div>
                <div>
                  <h2 className="text-2xl font-bold">Try the community pool</h2>
                  <p className="mt-2 text-sm leading-relaxed text-muted-foreground">
                    No dispenser account or API key is needed for public downloads.
                  </p>
                </div>
              </div>
              <div className="mt-6">
                <CommandBlock
                  label="Download an app"
                  command={`gplaydl download com.google.android.calculator -d ${origin}/api/auth`}
                />
              </div>
            </CardContent>
          </Card>
        </section>

        <section className="border-y border-border/70 bg-card/25">
          <div className="mx-auto grid max-w-6xl gap-8 px-6 py-16 md:grid-cols-2">
            <div id="terms" className="scroll-mt-24">
              <ShieldAlert className="size-7 text-chart-4" />
              <h2 className="mt-4 text-2xl font-bold">Use a spare Google account</h2>
              <p className="mt-3 text-sm leading-relaxed text-muted-foreground">
                An AAS token is a powerful Google Play credential. Do not contribute your primary,
                work, purchased-app, or payment-linked account. Google can revoke unofficial
                client access at any time.
              </p>
            </div>
            <div id="privacy" className="scroll-mt-24">
              <Users className="size-7 text-primary" />
              <h2 className="mt-4 text-2xl font-bold">You stay in control</h2>
              <p className="mt-3 text-sm leading-relaxed text-muted-foreground">
                The dispenser stores the Google account email and encrypted AAS token. Making an
                account private or deleting it stops future public dispensing; already issued
                short-lived Play sessions cannot be recalled.
              </p>
            </div>
          </div>
        </section>
      </main>

      <footer className="mx-auto flex max-w-6xl flex-col gap-4 px-6 py-10 text-sm text-muted-foreground sm:flex-row sm:items-center sm:justify-between">
        <Logo className="opacity-80" />
        <div className="flex gap-5">
          <a href="/pair" className="hover:text-foreground">
            Pair dashboard
          </a>
          <a href={downloadUrl} className="hover:text-foreground">
            Download app
          </a>
          <a href="#terms" className="hover:text-foreground">
            Sharing terms
          </a>
          <a href="#privacy" className="hover:text-foreground">
            Privacy
          </a>
        </div>
      </footer>
    </div>
  )
}
