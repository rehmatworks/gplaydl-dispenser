import { Card, CardContent } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import type { AppRelease, User } from "@/lib/api"
import { Download, Smartphone } from "lucide-react"

interface Props {
  user: User
  release: AppRelease | null
}

/**
 * Shown to contributors who reached the dashboard with a pairing code rather
 * than a password, so it is obvious which phone this session belongs to.
 */
export function DeviceCard({ user, release }: Props) {
  return (
    <Card className="glass card-hover rounded-2xl border-0">
      <CardContent className="flex flex-wrap items-center gap-5 p-6">
        <div className="flex size-12 shrink-0 items-center justify-center rounded-2xl bg-gradient-to-br from-aurora-teal/20 to-aurora-violet/20 ring-1 ring-aurora-teal/20">
          <Smartphone className="size-5 text-aurora-teal" />
        </div>
        <div className="min-w-0 flex-1">
          <h3 className="font-display text-base font-semibold">
            Paired with {user.label || "your phone"}
          </h3>
          <p className="mt-1 text-sm text-muted-foreground">
            This dashboard belongs to an app install, not an email address. Anything you change
            here shows up in the app, and vice versa.
          </p>
        </div>
        {release?.url && (
          <Button asChild variant="outline" className="glass rounded-xl">
            <a href={release.url}>
              <Download className="size-4" /> App {release.version}
            </a>
          </Button>
        )}
      </CardContent>
    </Card>
  )
}
