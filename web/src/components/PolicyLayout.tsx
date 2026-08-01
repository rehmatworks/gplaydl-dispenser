import { Logo } from "@/components/Logo"
import { PublicFooter } from "@/components/PublicFooter"
import { Button } from "@/components/ui/button"
import { ArrowLeft, Smartphone } from "lucide-react"
import { useEffect, type ReactNode } from "react"

type PolicyLayoutProps = {
  title: string
  description: string
  children: ReactNode
}

export function PolicyLayout({ title, description, children }: PolicyLayoutProps) {
  useEffect(() => {
    document.title = `${title} · gplaydl dispenser`
  }, [title])

  return (
    <div className="min-h-dvh overflow-hidden">
      <div className="aurora-bg" />

      <header className="sticky top-0 z-40 border-b border-border/60 bg-background/80 backdrop-blur-xl">
        <div className="mx-auto flex h-16 max-w-3xl items-center justify-between gap-3 px-4 sm:px-6">
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
              <a href="/dashboard">
                <Smartphone className="size-4" />
                Dashboard
              </a>
            </Button>
          </div>
        </div>
      </header>

      <main className="mx-auto max-w-3xl px-6 py-12 sm:py-16">
        <a
          href="/"
          className="inline-flex items-center gap-2 text-sm text-muted-foreground transition-colors hover:text-foreground"
        >
          <ArrowLeft className="size-4" />
          Back to home
        </a>
        <p className="mt-10 text-sm font-medium text-primary">gplaydl dispenser</p>
        <h1 className="mt-2 font-display text-3xl font-bold tracking-tight sm:text-4xl">
          {title}
        </h1>
        <p className="mt-4 max-w-2xl leading-relaxed text-muted-foreground">{description}</p>
        <p className="mt-2 text-xs text-muted-foreground">Last updated August 1, 2026</p>

        <div className="mt-10 space-y-5">{children}</div>
      </main>

      <PublicFooter />
    </div>
  )
}
