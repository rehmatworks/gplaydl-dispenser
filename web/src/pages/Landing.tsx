import { CommandBlock } from "@/components/CommandBlock"
import { Logo } from "@/components/Logo"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { api, type AppRelease, type PublicStats } from "@/lib/api"
import { useAuth } from "@/lib/auth"
import {
  ArrowRight,
  Check,
  Download,
  Github,
  Globe2,
  KeyRound,
  Lock,
  ShieldCheck,
  Smartphone,
  Terminal,
  Users
} from "lucide-react"
import { QRCodeSVG } from "qrcode.react"
import { useEffect, useState } from "react"
import { Link } from "react-router-dom"

const steps = [
  {
    icon: Download,
    title: "Install the app",
    body: "One APK, no signup, no email address. It enrols itself with the dispenser the first time you open it."
  },
  {
    icon: Smartphone,
    title: "Sign in to a spare Google account",
    body: "The app opens Google's own sign-in page. Your password and 2FA codes stay on your phone; only the resulting token is uploaded."
  },
  {
    icon: Users,
    title: "The pool does the rest",
    body: "Your account joins the rotation automatically. Everyone running gplaydl can download again, and you can switch it back to private with one tap."
  }
]

const safety = [
  {
    icon: Lock,
    title: "Encrypted the moment it arrives",
    body: "Tokens are sealed with AES-256-GCM at rest and only decrypted in memory for the few seconds a sign-in takes."
  },
  {
    icon: Globe2,
    title: "Public or private, per account",
    body: "Share some accounts with the community and keep others to yourself. Private accounts are only ever dispensed to your own API key."
  },
  {
    icon: ShieldCheck,
    title: "Revocable at any time",
    body: "Flip an account to private, delete it from the dispenser, or change the Google password. Any of the three stops it being used immediately."
  },
  {
    icon: KeyRound,
    title: "Fair rotation, no favourites",
    body: "Accounts are handed out least-recently-used, and one that starts failing drops out of the pool on its own."
  }
]

const faqs = [
  {
    q: "Why does gplaydl need someone's Google account at all?",
    a: "Google Play will not answer requests that are not tied to a signed-in account. There is no anonymous mode. A shared pool of spare accounts is the only way an open-source client can work without asking every user to hand over their own credentials."
  },
  {
    q: "What exactly is uploaded?",
    a: "The account's email address and an AAS token. That is the same token Google issues to a phone during setup. Your password, session cookies and 2FA secrets never leave the app."
  },
  {
    q: "Can someone read my email or spend my money with the token?",
    a: "The token is scoped to the Play Store APIs the client uses — app metadata and downloads. It is not a Gmail credential. Even so, use a spare account: Google can flag accounts used by unofficial clients, and you do not want that to be an account you rely on."
  },
  {
    q: "I want to download apps I paid for. Do I have to share the account?",
    a: "No. Add it as private, then point gplaydl at your own account with your API key and the account's email. It never enters the community rotation."
  },
  {
    q: "Do I need to keep the app installed?",
    a: "No, the token keeps working after you uninstall. Keeping it around is handy though: if Google ever invalidates the token, signing in again refreshes it in a couple of taps."
  },
  {
    q: "How do I stop contributing?",
    a: "Open the app, switch the account to private, or remove it outright. You can also do both from the web dashboard after pairing your phone."
  }
]

export default function Landing() {
  const { user } = useAuth()
  const origin = window.location.origin

  const [stats, setStats] = useState<PublicStats | null>(null)
  const [release, setRelease] = useState<AppRelease | null>(null)

  useEffect(() => {
    api.publicStats().then(setStats).catch(() => {})
    api.appLatest().then(setRelease).catch(() => {})
  }, [])

  const apkUrl = release?.url ?? `${origin}/downloads/gplaydl-authenticator-latest.apk`

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
            <a href="#contribute" className="transition-colors hover:text-foreground">
              Contribute
            </a>
            <a href="#use" className="transition-colors hover:text-foreground">
              Use it
            </a>
            <a href="#safety" className="transition-colors hover:text-foreground">
              Safety
            </a>
            <a href="#faq" className="transition-colors hover:text-foreground">
              FAQ
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
                <Button asChild variant="ghost" className="hidden rounded-xl sm:inline-flex">
                  <Link to="/login">Sign in</Link>
                </Button>
                <Button asChild className="btn-aurora rounded-xl">
                  <a href={apkUrl}>
                    <Download className="size-4" /> Get the app
                  </a>
                </Button>
              </>
            )}
          </div>
        </div>
      </header>

      {/* Hero */}
      <section className="relative px-6 pb-20 pt-40">
        <div className="grid-dots absolute inset-0" aria-hidden />
        <div className="relative mx-auto max-w-5xl">
          <div className="text-center">
            <Badge
              variant="outline"
              className="animate-fade-up mb-6 gap-2 rounded-full border-aurora-teal/30 bg-aurora-teal/5 px-4 py-1.5 text-aurora-teal"
            >
              <span className="size-1.5 animate-pulse rounded-full bg-aurora-teal" />
              Community project · Open source · GPL-3.0
            </Badge>

            <h1 className="animate-fade-up text-balance text-5xl font-bold leading-[1.05] tracking-tight md:text-7xl">
              Keep gplaydl working.
              <br />
              <span className="text-aurora">Share one spare login.</span>
            </h1>

            <p className="animate-fade-up mx-auto mt-6 max-w-2xl text-pretty text-lg text-muted-foreground [animation-delay:120ms]">
              Google Play only answers requests from a signed-in account. Install the
              Authenticator app, sign in with an account you do not mind sharing, and it joins a
              rotating community pool that every gplaydl user downloads through.
            </p>
          </div>

          {/* Download + live numbers */}
          <div className="animate-fade-up mt-14 grid gap-5 md:grid-cols-[1.4fr_1fr] [animation-delay:240ms]">
            <Card className="glass-strong card-hover rounded-2xl border-0">
              <CardContent className="flex flex-col gap-6 p-7 sm:flex-row sm:items-center">
                <div className="shrink-0 rounded-2xl bg-white p-3">
                  <QRCodeSVG value={apkUrl} size={116} level="M" />
                </div>
                <div className="min-w-0 flex-1">
                  <h2 className="font-display text-xl font-semibold">gplaydl Authenticator</h2>
                  <p className="mt-1.5 text-sm leading-relaxed text-muted-foreground">
                    Scan with your phone, or download it directly. Android 7 and newer.
                    {release?.version && (
                      <>
                        {" "}
                        Current version{" "}
                        <span className="mono-chip">{release.version}</span>.
                      </>
                    )}
                  </p>
                  <div className="mt-5 flex flex-wrap gap-3">
                    <Button asChild className="btn-aurora h-11 rounded-xl px-6">
                      <a href={apkUrl}>
                        <Download className="size-4" /> Download APK
                      </a>
                    </Button>
                    <Button asChild variant="outline" className="glass h-11 rounded-xl px-5">
                      <a href="#contribute">How it works</a>
                    </Button>
                  </div>
                </div>
              </CardContent>
            </Card>

            <Card className="glass card-hover rounded-2xl border-0">
              <CardContent className="p-7">
                <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
                  Pool right now
                </p>
                <dl className="mt-5 space-y-4">
                  <Stat label="Shared accounts" value={stats?.publicAccounts} />
                  <Stat label="Contributors" value={stats?.contributors} />
                  <Stat label="Logins today" value={stats?.mints24h} />
                  <Stat label="Logins all time" value={stats?.totalMints} />
                </dl>
              </CardContent>
            </Card>
          </div>

          {stats?.publicAccounts === 0 && (
            <p className="animate-fade-up mt-4 text-center text-sm text-chart-4">
              The pool is empty right now — the next account shared will be the one keeping
              everyone else downloading.
            </p>
          )}
        </div>
      </section>

      {/* Contribute */}
      <section id="contribute" className="px-6 py-24">
        <div className="mx-auto max-w-5xl">
          <h2 className="text-center text-3xl font-bold tracking-tight md:text-4xl">
            Contributing takes <span className="text-aurora">about two minutes</span>
          </h2>
          <p className="mx-auto mt-3 max-w-2xl text-center text-muted-foreground">
            There is nothing to sign up for and nothing to copy and paste. The app handles the
            whole handshake.
          </p>

          <div className="mt-14 grid gap-5 md:grid-cols-3">
            {steps.map((s, i) => (
              <Card key={s.title} className="glass card-hover rounded-2xl border-0">
                <CardContent className="p-6">
                  <div className="mb-5 flex items-center gap-3">
                    <div className="flex size-11 items-center justify-center rounded-xl bg-gradient-to-br from-aurora-teal/20 to-aurora-violet/20 ring-1 ring-aurora-teal/20">
                      <s.icon className="size-5 text-aurora-teal" />
                    </div>
                    <span className="font-display text-3xl font-bold text-aurora opacity-80">
                      {i + 1}
                    </span>
                  </div>
                  <h3 className="font-display text-lg font-semibold">{s.title}</h3>
                  <p className="mt-2 text-sm leading-relaxed text-muted-foreground">{s.body}</p>
                </CardContent>
              </Card>
            ))}
          </div>

          <div className="glass mt-8 rounded-2xl p-6">
            <p className="text-sm leading-relaxed text-muted-foreground">
              <span className="font-semibold text-foreground">Please use a spare account.</span>{" "}
              Google may flag or suspend accounts that sign in through unofficial clients. Make a
              fresh Google account for this — it needs nothing in it, no payment method and no
              personal data. If you want to download apps you have <em>purchased</em>, add that
              account as <span className="mono-chip">private</span> instead and only your own API
              key will ever reach it.
            </p>
          </div>

          <div className="mt-10 flex flex-wrap justify-center gap-4">
            <Button asChild size="lg" className="btn-aurora h-12 rounded-2xl px-8 text-base">
              <a href={apkUrl}>
                <Download className="size-4" /> Download the app
              </a>
            </Button>
            <Button
              asChild
              size="lg"
              variant="outline"
              className="glass h-12 rounded-2xl px-8 text-base"
            >
              <Link to="/pair">Already have it? Pair your phone</Link>
            </Button>
          </div>

          <p className="mt-6 text-center text-sm text-muted-foreground">
            Prefer to paste a token by hand?{" "}
            <Link to="/register" className="text-aurora-teal hover:underline">
              Create a web account
            </Link>{" "}
            instead — the old workflow still works.
          </p>
        </div>
      </section>

      {/* Use it */}
      <section id="use" className="px-6 py-24">
        <div className="mx-auto max-w-3xl">
          <h2 className="text-center text-3xl font-bold tracking-tight md:text-4xl">
            Downloading with gplaydl
          </h2>
          <p className="mx-auto mt-3 max-w-xl text-center text-muted-foreground">
            Point the <span className="mono-chip">--dispenser</span> flag here. Nothing else to
            configure.
          </p>

          <div className="mt-12 space-y-8">
            <div>
              <h3 className="font-display flex items-center gap-2 text-lg font-semibold">
                <Terminal className="size-4 text-aurora-teal" /> Anyone, from the community pool
              </h3>
              <p className="mt-1.5 text-sm text-muted-foreground">
                No account and no API key needed.
              </p>
              <CommandBlock
                className="mt-4"
                command={`gplaydl download com.instagram.android -d ${origin}/api/auth`}
              />
            </div>

            <div>
              <h3 className="font-display flex items-center gap-2 text-lg font-semibold">
                <Lock className="size-4 text-aurora-teal" /> Your own account, by email
              </h3>
              <p className="mt-1.5 text-sm text-muted-foreground">
                Pins the download to one of your private accounts — this is how you reach apps you
                have purchased. Both values come from your dashboard.
              </p>
              <CommandBlock
                className="mt-4"
                command={`gplaydl download com.your.app -d "${origin}/api/auth?api_key=YOUR_KEY&email=you@gmail.com"`}
              />
            </div>

            <div>
              <h3 className="font-display flex items-center gap-2 text-lg font-semibold">
                <KeyRound className="size-4 text-aurora-teal" /> Any of your own accounts
              </h3>
              <p className="mt-1.5 text-sm text-muted-foreground">
                Rotates across your accounts only, skipping the community pool.
              </p>
              <CommandBlock
                className="mt-4"
                command={`gplaydl download com.your.app -d "${origin}/api/auth?api_key=YOUR_KEY&pool=private"`}
              />
            </div>
          </div>

          <div className="glass mt-10 rounded-2xl p-6">
            <p className="text-sm leading-relaxed text-muted-foreground">
              Writing your own client? <span className="mono-chip">GET /api/auth</span> returns{" "}
              <span className="mono-chip">{"{ email, auth }"}</span> for the community pool.{" "}
              <span className="mono-chip">POST /api/auth</span> takes a device properties object
              and returns the full auth bundle, including{" "}
              <span className="mono-chip">authToken</span>,{" "}
              <span className="mono-chip">gsfId</span> and{" "}
              <span className="mono-chip">deviceConfigToken</span>.
            </p>
          </div>
        </div>
      </section>

      {/* Safety */}
      <section id="safety" className="px-6 py-24">
        <div className="mx-auto max-w-5xl">
          <h2 className="text-center text-3xl font-bold tracking-tight md:text-4xl">
            What sharing actually means
          </h2>
          <p className="mx-auto mt-3 max-w-xl text-center text-muted-foreground">
            The app says all of this before it uploads anything. Here it is in writing.
          </p>

          <div className="mt-14 grid gap-5 sm:grid-cols-2">
            {safety.map((f) => (
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

          <Card className="glass-strong mt-8 rounded-2xl border-0">
            <CardContent className="p-7">
              <h3 className="font-display text-lg font-semibold">
                Uploaded when you share an account
              </h3>
              <ul className="mt-4 grid gap-2.5 text-sm text-muted-foreground sm:grid-cols-2">
                {[
                  "The account's email address",
                  "One AAS token (starts with aas_et/)",
                  "Which consent wording you agreed to",
                  "A random id for your app install"
                ].map((item) => (
                  <li key={item} className="flex items-center gap-2">
                    <Check className="size-4 shrink-0 text-aurora-teal" /> {item}
                  </li>
                ))}
              </ul>
              <h3 className="font-display mt-7 text-lg font-semibold">Never uploaded</h3>
              <ul className="mt-4 grid gap-2.5 text-sm text-muted-foreground sm:grid-cols-2">
                {[
                  "Your Google password",
                  "Your 2FA codes or recovery keys",
                  "Google session cookies",
                  "Contacts, mail, files or payment details"
                ].map((item) => (
                  <li key={item} className="flex items-center gap-2">
                    <span className="size-4 shrink-0 text-center text-destructive">×</span> {item}
                  </li>
                ))}
              </ul>
            </CardContent>
          </Card>
        </div>
      </section>

      {/* FAQ */}
      <section id="faq" className="px-6 py-24">
        <div className="mx-auto max-w-3xl">
          <h2 className="text-center text-3xl font-bold tracking-tight md:text-4xl">
            Questions people actually ask
          </h2>

          <div className="mt-12 space-y-3">
            {faqs.map((f) => (
              <details key={f.q} className="glass card-hover group rounded-2xl px-6 py-5">
                <summary className="font-display cursor-pointer list-none text-base font-semibold marker:content-none">
                  <span className="flex items-start justify-between gap-4">
                    {f.q}
                    <span className="mt-0.5 shrink-0 text-aurora-teal transition-transform group-open:rotate-45">
                      +
                    </span>
                  </span>
                </summary>
                <p className="mt-3 text-sm leading-relaxed text-muted-foreground">{f.a}</p>
              </details>
            ))}
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t border-border px-6 py-10">
        <div className="mx-auto flex max-w-6xl flex-col items-center justify-between gap-4 md:flex-row">
          <Logo className="opacity-70" />
          <div className="flex flex-wrap items-center justify-center gap-5 text-sm text-muted-foreground">
            <a href={apkUrl} className="transition-colors hover:text-foreground">
              Download the app
            </a>
            <Link to="/pair" className="transition-colors hover:text-foreground">
              Pair a phone
            </Link>
            <Link to="/login" className="transition-colors hover:text-foreground">
              Sign in
            </Link>
            <a
              href="https://github.com/rehmatworks/gplaydl"
              target="_blank"
              rel="noreferrer"
              className="transition-colors hover:text-foreground"
            >
              gplaydl
            </a>
          </div>
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
        <p className="mt-6 text-center text-xs text-muted-foreground">Open source under GPL-3.0.</p>
      </footer>
    </div>
  )
}

function Stat({ label, value }: { label: string; value?: number }) {
  return (
    <div className="flex items-baseline justify-between gap-3">
      <dt className="text-sm text-muted-foreground">{label}</dt>
      <dd className="font-display text-2xl font-bold text-aurora-teal">
        {value === undefined ? "—" : value.toLocaleString()}
      </dd>
    </div>
  )
}
