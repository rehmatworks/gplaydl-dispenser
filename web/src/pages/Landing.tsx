import { CommandBlock } from "@/components/CommandBlock"
import { Logo } from "@/components/Logo"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { api, type AppRelease } from "@/lib/api"
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
    title: "Choose Community, then sign in",
    body: "The app asks before every sign-in whether the account joins the shared pool or stays private to you. Use a spare Google account, never your main one.",
  },
  {
    title: "Change your mind whenever",
    body: "The account starts serving downloads right away. The Accounts screen in the app switches one back to private or deletes it.",
  },
]

export default function Landing() {
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
  const origin = window.location.origin

  return (
    <div className="min-h-dvh overflow-hidden">
      <div className="aurora-bg" />

      <header className="sticky top-0 z-40 border-b border-border/60 bg-background/80 backdrop-blur-xl">
        <div className="mx-auto flex h-16 max-w-4xl items-center justify-between gap-3 px-4 sm:px-6">
          <a href="/" aria-label="gplaydl dispenser home">
            <Logo />
          </a>
          <Button asChild variant="outline" size="sm" className="glass rounded-xl">
            <a href="/pair">
              <Smartphone className="size-4" />
              Dashboard
            </a>
          </Button>
        </div>
      </header>

      <main className="mx-auto max-w-4xl px-6 py-14">
        <h1 className="text-2xl font-bold tracking-tight sm:text-3xl">
          Contribute a Google account
        </h1>
        <p className="mt-3 leading-relaxed text-muted-foreground">
          Every Google Play login this dispenser hands out comes from a spare account
          somebody shared. Adding one takes about a minute.
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
          <h2 className="text-lg font-semibold">Using the pool</h2>
          <p className="mt-2 text-sm leading-relaxed text-muted-foreground">
            Public accounts need no signup and no API key. Point any client that speaks the
            dispenser API at this address.
          </p>
          <CommandBlock
            className="mt-5"
            command={`gplaydl download com.google.android.calculator -d ${origin}/api/auth`}
          />
        </section>

        <section id="terms" className="mt-16 scroll-mt-20 border-t border-border/60 pt-10">
          <h2 className="text-lg font-semibold">Sharing terms</h2>
          <dl className="mt-5 space-y-5 text-sm leading-relaxed">
            <div>
              <dt className="font-medium">Contribute a spare account only</dt>
              <dd className="mt-1 text-muted-foreground">
                Never your primary, work, payment-linked, or purchased-app account. An AAS
                token can browse and download free Play apps as that account, and Google may
                rate-limit, lock, or revoke accounts it sees on unofficial clients. Share
                nothing you would mind losing.
              </dd>
            </div>
            <div>
              <dt className="font-medium">Share only what is yours</dt>
              <dd className="mt-1 text-muted-foreground">
                Sign in with accounts you own and control. Do not upload someone else's
                credentials.
              </dd>
            </div>
            <div>
              <dt className="font-medium">You can withdraw at any time</dt>
              <dd className="mt-1 text-muted-foreground">
                Making an account private or deleting it stops all future dispensing. Play
                sessions already handed out are short-lived, but they cannot be recalled.
              </dd>
            </div>
            <div>
              <dt className="font-medium">No warranty</dt>
              <dd className="mt-1 text-muted-foreground">
                This is a community-run, best-effort service with no uptime promise. It may
                change or shut down, and it is not affiliated with Google.
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
              <dt className="font-medium">A public account address is public</dt>
              <dd className="mt-1 text-muted-foreground">
                Play needs both the address and the token, so the dispenser returns both to
                whoever asks the pool for a login. Anyone using the pool can see the address
                you contributed. This is exactly why it should be a throwaway one.
              </dd>
            </div>
            <div>
              <dt className="font-medium">Logs</dt>
              <dd className="mt-1 text-muted-foreground">
                IP addresses appear in short-lived server logs and in rate limiting. They are
                never written to the database or tied to a contributed account.
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
          <a href="/pair" className="whitespace-nowrap hover:text-foreground">
            Pair dashboard
          </a>
          <a href={downloadUrl} className="whitespace-nowrap hover:text-foreground">
            Download app
          </a>
          <a href="#terms" className="whitespace-nowrap hover:text-foreground">
            Sharing terms
          </a>
          <a href="#privacy" className="whitespace-nowrap hover:text-foreground">
            Privacy
          </a>
        </div>
      </footer>
    </div>
  )
}
