import { Logo } from "@/components/Logo"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { useAuth } from "@/lib/auth"
import {
  ArrowRight,
  Github,
  Globe2,
  HeartHandshake,
  KeyRound,
  Lock,
  ShieldCheck,
  Smartphone,
  UserPlus
} from "lucide-react"
import { Link } from "react-router-dom"

const howItWorks = [
  {
    step: "1",
    title: "An app asks for a login",
    body: "Open-source Play Store clients (like Aurora Store) need a Google session to browse and download apps anonymously."
  },
  {
    step: "2",
    title: "The pool answers",
    body: "This service picks one of the community-shared accounts in turn and signs in on the app's behalf."
  },
  {
    step: "3",
    title: "The user stays anonymous",
    body: "The app receives a temporary session token. Contributors' account details are never exposed to anyone."
  }
]

const contributeSteps = [
  {
    icon: UserPlus,
    title: "Create a free account here",
    body: "Register with any email address — it only takes a moment and is just used to manage your contributions."
  },
  {
    icon: ShieldCheck,
    title: "Use a spare Google account",
    body: "Don't use your personal account. Create a fresh Google account just for this — it needs nothing in it, no payment method, no personal data."
  },
  {
    icon: Smartphone,
    title: "Generate a token",
    body: (
      <>
        <a
          href="/downloads/Authenticator_v1.0.4.apk"
          className="text-aurora-teal hover:underline"
        >
          Download the Authenticator app
        </a>{" "}
        on any Android device, sign in to the spare account, and copy the token it gives you (it
        starts with aas_et/). That token is what you contribute — not your password.
      </>
    )
  },
  {
    icon: HeartHandshake,
    title: "Add it as public",
    body: "Paste the token in your dashboard and mark the account as Public. That's it — you can pause, make it private, or remove it again at any time."
  }
]

const features = [
  {
    icon: Globe2,
    title: "Public or private",
    body: "Share accounts with the community, or keep them private and use the service just for your own devices. You decide per account, and can change your mind anytime."
  },
  {
    icon: Lock,
    title: "Encrypted by default",
    body: "Tokens are sealed with AES-256 encryption the moment you add them, and are only ever decrypted in memory while signing in."
  },
  {
    icon: ShieldCheck,
    title: "You stay in control",
    body: "Removing an account here, or changing the Google account's password, immediately stops the service from using it. No strings attached."
  },
  {
    icon: KeyRound,
    title: "Works out of the box",
    body: "Any Aurora Store-compatible client can use this service at /api/auth — no configuration needed."
  }
]

export default function Landing() {
  const { user } = useAuth()
  const origin = window.location.origin

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
            <a href="#how-it-works" className="transition-colors hover:text-foreground">
              How it works
            </a>
            <a href="#contribute" className="transition-colors hover:text-foreground">
              Contribute
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
            A community project · Open source · GPL-3.0
          </Badge>

          <h1 className="animate-fade-up text-balance text-5xl font-bold leading-[1.05] tracking-tight md:text-7xl">
            Anonymous app store logins,
            <br />
            <span className="text-aurora">powered by people like you.</span>
          </h1>

          <p className="animate-fade-up mx-auto mt-6 max-w-2xl text-pretty text-lg text-muted-foreground [animation-delay:120ms]">
            Open-source Play Store clients need Google accounts to let people browse apps without
            signing in themselves. This small community pool makes that possible — one shared
            spare account at a time.
          </p>

          <div className="animate-fade-up mt-10 flex flex-wrap items-center justify-center gap-4 [animation-delay:240ms]">
            <Button asChild size="lg" className="btn-aurora h-12 rounded-2xl px-8 text-base">
              <Link to="/register">
                Contribute an account <ArrowRight className="size-4" />
              </Link>
            </Button>
            <Button
              asChild
              size="lg"
              variant="outline"
              className="glass h-12 rounded-2xl px-8 text-base"
            >
              <a href="#how-it-works">How it works</a>
            </Button>
          </div>
        </div>
      </section>

      {/* How it works */}
      <section id="how-it-works" className="px-6 py-24">
        <div className="mx-auto max-w-5xl">
          <h2 className="text-center text-3xl font-bold tracking-tight md:text-4xl">
            How it works
          </h2>
          <p className="mx-auto mt-3 max-w-xl text-center text-muted-foreground">
            Three things happen every time someone browses apps anonymously.
          </p>

          <div className="mt-14 grid gap-5 md:grid-cols-3">
            {howItWorks.map((s) => (
              <Card key={s.step} className="glass card-hover relative rounded-2xl border-0">
                <CardContent className="p-6">
                  <div className="font-display mb-4 text-5xl font-bold text-aurora opacity-90">
                    {s.step}
                  </div>
                  <h3 className="font-display text-lg font-semibold">{s.title}</h3>
                  <p className="mt-2 text-sm leading-relaxed text-muted-foreground">{s.body}</p>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      </section>

      {/* Contribute guide */}
      <section id="contribute" className="px-6 py-24">
        <div className="mx-auto max-w-5xl">
          <h2 className="text-center text-3xl font-bold tracking-tight md:text-4xl">
            Contribute in <span className="text-aurora">four easy steps</span>
          </h2>
          <p className="mx-auto mt-3 max-w-xl text-center text-muted-foreground">
            No technical knowledge needed. Ten minutes, start to finish.
          </p>

          <div className="mt-14 space-y-4">
            {contributeSteps.map((s, i) => (
              <Card key={s.title} className="glass card-hover rounded-2xl border-0">
                <CardContent className="flex items-start gap-5 p-6">
                  <div className="flex size-12 shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-aurora-teal/20 to-aurora-violet/20 ring-1 ring-aurora-teal/20">
                    <s.icon className="size-5 text-aurora-teal" />
                  </div>
                  <div>
                    <h3 className="font-display text-lg font-semibold">
                      <span className="text-aurora-teal">Step {i + 1}.</span> {s.title}
                    </h3>
                    <p className="mt-1.5 text-sm leading-relaxed text-muted-foreground">
                      {s.body}
                    </p>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>

          <div className="glass mt-8 rounded-2xl p-6">
            <p className="text-sm leading-relaxed text-muted-foreground">
              <span className="font-semibold text-foreground">Where do I get the token?</span>{" "}
              Install the free, open-source Authenticator app on any Android device —{" "}
              <a
                href="/downloads/Authenticator_v1.0.4.apk"
                className="text-aurora-teal hover:underline"
              >
                download the APK here
              </a>{" "}
              (source code on{" "}
              <a
                href="https://github.com/whyorean/Authenticator"
                target="_blank"
                rel="noreferrer"
                className="text-aurora-teal hover:underline"
              >
                GitHub
              </a>
              ) — then sign in with the spare account and it shows you the token to copy. The
              token keeps working until you change the account's password — which is also how you
              revoke it for good.
            </p>
          </div>

          <div className="mt-10 text-center">
            <Button asChild size="lg" className="btn-aurora h-12 rounded-2xl px-8 text-base">
              <Link to="/register">
                Start contributing <ArrowRight className="size-4" />
              </Link>
            </Button>
          </div>
        </div>
      </section>

      {/* Features */}
      <section id="features" className="px-6 py-24">
        <div className="mx-auto max-w-5xl">
          <h2 className="text-center text-3xl font-bold tracking-tight md:text-4xl">
            Safe by design
          </h2>
          <p className="mx-auto mt-3 max-w-xl text-center text-muted-foreground">
            Built so contributors never have to worry.
          </p>

          <div className="mt-14 grid gap-5 sm:grid-cols-2">
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
            For app developers
          </h2>
          <p className="mx-auto mt-3 max-w-xl text-center text-muted-foreground">
            Anonymous requests use the community pool. Pass your API key to use your own private
            accounts instead.
          </p>

          <Card className="glass-strong mt-10 overflow-hidden rounded-2xl border-0">
            <div className="flex items-center gap-2 border-b border-border px-5 py-3">
              <span className="size-3 rounded-full bg-destructive/60" />
              <span className="size-3 rounded-full bg-chart-4/60" />
              <span className="size-3 rounded-full bg-aurora-teal/60" />
              <span className="ml-3 font-mono text-xs text-muted-foreground">terminal</span>
            </div>
            <CardContent className="overflow-x-auto p-5 font-mono text-sm leading-7">
              <div className="text-muted-foreground"># Anonymous login from the community pool</div>
              <div>
                <span className="text-aurora-teal">curl</span> {origin}
                <span className="text-aurora-pink">/api/auth</span>
              </div>
              <div className="mt-4 text-muted-foreground"># Use your own private accounts</div>
              <div>
                <span className="text-aurora-teal">curl</span> -H{" "}
                <span className="text-chart-4">"X-Api-Key: $KEY"</span> "{origin}
                <span className="text-aurora-pink">/api/auth?pool=private</span>"
              </div>
            </CardContent>
          </Card>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t border-border px-6 py-10">
        <div className="mx-auto flex max-w-6xl flex-col items-center justify-between gap-4 md:flex-row">
          <Logo className="opacity-70" />
          <p className="text-sm text-muted-foreground">Open source under GPL-3.0.</p>
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
