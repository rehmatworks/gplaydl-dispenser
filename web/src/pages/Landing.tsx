import { CommandBlock } from "@/components/CommandBlock"
import { Logo } from "@/components/Logo"
import { PublicFooter } from "@/components/PublicFooter"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { api, type AppRelease } from "@/lib/api"
import { useAuth } from "@/lib/auth"
import { Download, Smartphone } from "lucide-react"
import { useEffect, useState } from "react"
import { QRCodeSVG } from "qrcode.react"

const fallbackDownload = "/downloads/gplaydl-authenticator-latest.apk"

const steps = [
  {
    title: "Install the Authenticator app",
    body: "Android 7.0 or newer. It is not on Google Play, so Android will ask you to allow installs from this source.",
  },
  {
    title: "Sign in with a Google account",
    body: "Use a spare account rather than your main one. The account stays private to you; it is never shared with anyone else.",
  },
  {
    title: "Link gplaydl with the pairing code",
    body: "Run gplaydl link on your computer, open the Link gplaydl screen in the app, and type the short code it shows. That is the whole setup.",
  },
  {
    title: "Download",
    body: "gplaydl now downloads through your own account. Add several and pass --email to pick which one to use.",
  },
]

export default function Landing() {
  const { user, loading } = useAuth()
  const [release, setRelease] = useState<AppRelease | null>(null)

  useEffect(() => {
    api.appLatest().then(setRelease).catch(() => {})
  }, [])

  const downloadUrl = release?.url || fallbackDownload
  // Prefer the protected dashboard while auth is still loading. It will
  // redirect unauthenticated visitors to pairing once the session check ends.
  const dashboardHref = loading || user ? "/dashboard" : "/pair"

  return (
    <div className="min-h-dvh overflow-hidden">
      <div className="aurora-bg" />

      <header className="sticky top-0 z-40 border-b border-border/60 bg-background/80 backdrop-blur-xl">
        <div className="mx-auto flex h-16 max-w-4xl items-center justify-between gap-3 px-4 sm:px-6">
          <a href="/" aria-label="gplaydl dispenser home">
            <Logo />
          </a>
          <div className="flex items-center gap-3 sm:gap-5">
            <nav
              className="hidden items-center gap-4 text-sm text-muted-foreground sm:flex"
              aria-label="Get gplaydl"
            >
              <a href="https://gplaydl.com" className="hover:text-foreground">
                Web downloader
              </a>
              <a
                href="https://github.com/rehmatworks/gplaydl"
                className="hover:text-foreground"
              >
                CLI
              </a>
            </nav>
            <Button asChild variant="outline" size="sm" className="glass rounded-xl">
              <a href={dashboardHref}>
                <Smartphone className="size-4" />
                Dashboard
              </a>
            </Button>
          </div>
        </div>
      </header>

      <main className="mx-auto max-w-4xl px-6 py-14">
        <h1 className="text-2xl font-bold tracking-tight sm:text-3xl">
          Use gplaydl with your own Google account
        </h1>
        <p className="mt-3 leading-relaxed text-muted-foreground">
          This dispenser mints Google Play tokens from accounts you add yourself. Your
          accounts stay private to you, and setting up takes about two minutes.
        </p>

        <section className="mt-10 grid gap-10 md:grid-cols-[1fr_auto]">
          <ol className="space-y-7">
            {steps.map((step, i) => (
              <li key={step.title} className="flex gap-4">
                <span className="flex size-8 shrink-0 items-center justify-center rounded-full border border-primary/25 bg-primary/8 text-sm font-semibold text-primary">
                  {i + 1}
                </span>
                <div>
                  <p className="font-semibold leading-8">{step.title}</p>
                  <p className="mt-1 text-sm leading-relaxed text-muted-foreground">
                    {step.body}
                  </p>
                </div>
              </li>
            ))}
          </ol>

          <Card className="glass-strong hidden shrink-0 rounded-2xl border-0 p-0 md:block">
            <CardContent className="flex flex-col items-center p-5 text-center">
              <div className="rounded-xl bg-white p-3">
                <QRCodeSVG value={downloadUrl} size={140} level="M" />
              </div>
              <p className="mt-4 text-sm text-muted-foreground">Scan on your phone</p>
            </CardContent>
          </Card>
        </section>

        <div className="mt-10">
          <Button asChild size="lg" className="btn-aurora h-12 w-full rounded-xl px-6 sm:w-auto">
            <a href={downloadUrl}>
              <Download className="size-5" />
              Download the Authenticator app
            </a>
          </Button>
          <p className="mt-3 text-xs text-muted-foreground">
            {release ? `Version ${release.version}` : "Latest release"} · signed APK
            {release?.sha256 ? ` · SHA-256 ${release.sha256.slice(0, 12)}…` : ""}
            {" · "}
            <a
              href="https://github.com/rehmatworks/gplaydl-authenticator"
              className="hover:text-foreground hover:underline"
            >
              source
            </a>
          </p>
        </div>

        <section id="use" className="mt-16 scroll-mt-20 border-t border-border/60 pt-10">
          <h2 className="text-lg font-semibold">Using gplaydl</h2>
          <p className="mt-2 text-sm leading-relaxed text-muted-foreground">
            gplaydl 4.0 and newer uses this dispenser out of the box. Link it once with
            the pairing code from the app, and every download after that just works.
          </p>
          <CommandBlock className="mt-5" command="gplaydl link" />
          <CommandBlock
            className="mt-3"
            command="gplaydl download com.google.android.calculator"
          />
          <p className="mt-4 text-sm leading-relaxed text-muted-foreground">
            Added more than one account? Pass{" "}
            <code className="rounded bg-muted px-1.5 py-0.5">--email you@gmail.com</code> to
            download as a specific one. Otherwise gplaydl rotates through the accounts you
            added.
          </p>
          <p className="mt-3 text-sm leading-relaxed text-muted-foreground">
            Prefer a browser? Open{" "}
            <a className="text-primary hover:underline" href="https://gplaydl.com">
              gplaydl web
            </a>
            .
          </p>
        </section>

        <section id="selfhost" className="mt-16 scroll-mt-20 border-t border-border/60 pt-10">
          <h2 className="text-lg font-semibold">Self-hosting</h2>
          <p className="mt-2 text-sm leading-relaxed text-muted-foreground">
            The dispenser is open source. Run your own for a team or just for yourself,
            point the app&apos;s server setting at it, and link gplaydl with{" "}
            <code className="rounded bg-muted px-1.5 py-0.5">gplaydl link -d https://your.dispenser</code>.
          </p>
          <p className="mt-3 text-sm leading-relaxed">
            <a
              className="text-primary hover:underline"
              href="https://github.com/rehmatworks/gplaydl-dispenser"
            >
              github.com/rehmatworks/gplaydl-dispenser
            </a>
          </p>
        </section>

      </main>

      <PublicFooter dashboardHref={dashboardHref} downloadHref={downloadUrl} />
    </div>
  )
}
