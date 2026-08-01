import { Logo } from "@/components/Logo"

type PublicFooterProps = {
  dashboardHref?: string
  downloadHref?: string
}

const linkClass =
  "text-sm text-muted-foreground transition-colors hover:text-foreground"

export function PublicFooter({
  dashboardHref = "/dashboard",
  downloadHref = "/downloads/gplaydl-authenticator-latest.apk",
}: PublicFooterProps) {
  return (
    <footer className="border-t border-border/60">
      <div className="mx-auto grid max-w-4xl grid-cols-2 gap-x-8 gap-y-9 px-6 py-10 sm:grid-cols-3 md:grid-cols-[1.4fr_1fr_1fr_0.8fr]">
        <div className="col-span-2 sm:col-span-3 md:col-span-1">
          <Logo className="opacity-90" />
          <p className="mt-3 max-w-xs text-sm leading-relaxed text-muted-foreground">
            Connect gplaydl to Google Play with accounts that stay private to you.
          </p>
        </div>

        <nav aria-label="Use gplaydl">
          <p className="text-xs font-semibold uppercase tracking-wider text-foreground">
            Use gplaydl
          </p>
          <div className="mt-3 flex flex-col items-start gap-2.5">
            <a href="https://gplaydl.com" className={linkClass}>
              Web downloader
            </a>
            <a href="https://github.com/rehmatworks/gplaydl" className={linkClass}>
              CLI
            </a>
          </div>
        </nav>

        <nav aria-label="Dispenser">
          <p className="text-xs font-semibold uppercase tracking-wider text-foreground">
            Dispenser
          </p>
          <div className="mt-3 flex flex-col items-start gap-2.5">
            <a href={dashboardHref} className={linkClass}>
              Dashboard
            </a>
            <a href={downloadHref} className={linkClass}>
              Download app
            </a>
            <a href="/#selfhost" className={linkClass}>
              Self-hosting
            </a>
            <a
              href="https://github.com/rehmatworks/gplaydl-dispenser"
              className={linkClass}
            >
              Source
            </a>
          </div>
        </nav>

        <nav aria-label="Legal">
          <p className="text-xs font-semibold uppercase tracking-wider text-foreground">
            Legal
          </p>
          <div className="mt-3 flex flex-col items-start gap-2.5">
            <a href="/terms" className={linkClass}>
              Terms
            </a>
            <a href="/privacy" className={linkClass}>
              Privacy
            </a>
          </div>
        </nav>
      </div>
    </footer>
  )
}
