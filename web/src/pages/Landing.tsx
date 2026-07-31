import { CommandBlock } from "@/components/CommandBlock"
import { Logo } from "@/components/Logo"
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

  // The browser resolves the initial hash before React has rendered the
  // target, so /#terms and /#privacy from the app would land at the top.
  useEffect(() => {
    const id = window.location.hash.slice(1)
    if (id) document.getElementById(id)?.scrollIntoView()
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
          <Button asChild variant="outline" size="sm" className="glass rounded-xl">
            <a href={dashboardHref}>
              <Smartphone className="size-4" />
              Dashboard
            </a>
          </Button>
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

        <section id="terms" className="mt-16 scroll-mt-20 border-t border-border/60 pt-10">
          <h2 className="text-lg font-semibold">Terms</h2>
          <dl className="mt-5 space-y-5 text-sm leading-relaxed">
            <div>
              <dt className="font-medium">Use a spare account</dt>
              <dd className="mt-1 text-muted-foreground">
                Prefer a throwaway over your primary, work, or payment-linked account. An AAS
                token can browse and download free Play apps as that account, and Google may
                rate-limit, lock, or revoke accounts it sees on unofficial clients.
              </dd>
            </div>
            <div>
              <dt className="font-medium">Add only accounts that are yours</dt>
              <dd className="mt-1 text-muted-foreground">
                Sign in with accounts you own and control. Do not upload someone else&apos;s
                credentials.
              </dd>
            </div>
            <div>
              <dt className="font-medium">You can remove an account at any time</dt>
              <dd className="mt-1 text-muted-foreground">
                Deleting an account in the app erases it from the dispenser and stops all
                future dispensing. Play sessions already handed to your own gplaydl are
                short-lived and expire on their own.
              </dd>
            </div>
            <div>
              <dt className="font-medium">No warranty</dt>
              <dd className="mt-1 text-muted-foreground">
                This is a best-effort service with no uptime promise. It may change or shut
                down, and it is not affiliated with Google. You can always self-host.
              </dd>
            </div>
          </dl>
        </section>

        <section id="privacy" className="mt-16 scroll-mt-20 border-t border-border/60 pt-10">
          <h2 className="text-lg font-semibold">Privacy</h2>
          <dl className="mt-5 space-y-5 text-sm leading-relaxed">
            <div>
              <dt className="font-medium">What is stored</dt>
              <dd className="mt-1 text-muted-foreground">
                The Google account address and its AAS token, encrypted at rest with
                AES-256-GCM, plus how often the account has been used and when it was last
                used. The app enrols your phone as an anonymous device, so there is no name,
                email, or password behind it.
              </dd>
            </div>
            <div>
              <dt className="font-medium">Your Google password never reaches us</dt>
              <dd className="mt-1 text-muted-foreground">
                Sign-in happens with Google inside the app. Passwords, two-factor codes, and
                cookies stay on your phone; only the resulting token is uploaded.
              </dd>
            </div>
            <div>
              <dt className="font-medium">Your accounts are yours alone</dt>
              <dd className="mt-1 text-muted-foreground">
                An account you add is only ever used to serve your own linked gplaydl. It is
                never handed to anyone else and never appears in a shared pool.
              </dd>
            </div>
            <div>
              <dt className="font-medium">Logs</dt>
              <dd className="mt-1 text-muted-foreground">
                IP addresses appear in short-lived server logs and in rate limiting. They are
                never written to the database or tied to an account.
              </dd>
            </div>
            <div>
              <dt className="font-medium">Deletion is deletion</dt>
              <dd className="mt-1 text-muted-foreground">
                Removing an account erases its address and token from the database. Usage
                counters kept for service health survive, with nothing left to link them to.
              </dd>
            </div>
          </dl>
        </section>
      </main>

      <footer className="mx-auto flex max-w-4xl flex-col gap-4 border-t border-border/60 px-6 py-8 text-sm text-muted-foreground sm:flex-row sm:items-center sm:justify-between">
        <Logo className="opacity-80" />
        <div className="flex flex-wrap gap-x-5 gap-y-2">
          <a href={dashboardHref} className="whitespace-nowrap hover:text-foreground">
            Dashboard
          </a>
          <a href={downloadUrl} className="whitespace-nowrap hover:text-foreground">
            Download app
          </a>
          <a href="#selfhost" className="whitespace-nowrap hover:text-foreground">
            Self-hosting
          </a>
          <a href="#terms" className="whitespace-nowrap hover:text-foreground">
            Terms
          </a>
          <a href="#privacy" className="whitespace-nowrap hover:text-foreground">
            Privacy
          </a>
        </div>
      </footer>
    </div>
  )
}
