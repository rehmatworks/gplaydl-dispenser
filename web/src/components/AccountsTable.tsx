import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { api, ApiError, type Account } from "@/lib/api"
import { AlertTriangle, CheckCircle2, Clock3, Lock, Trash2 } from "lucide-react"
import { toast } from "sonner"

interface Props {
  accounts: Account[]
  onDelete: (id: string) => void
}

function relativeTime(value: string | null) {
  if (!value) return "Not used yet"
  const date = new Date(value)
  const diff = Date.now() - date.getTime()
  const minutes = Math.max(1, Math.round(diff / 60_000))
  if (minutes < 60) return `${minutes}m ago`
  const hours = Math.round(minutes / 60)
  if (hours < 24) return `${hours}h ago`
  return `${Math.round(hours / 24)}d ago`
}

export function AccountsTable({ accounts, onDelete }: Props) {
  async function remove(account: Account) {
    if (!confirm(`Remove ${account.email} from the dispenser? This cannot be undone.`)) return
    try {
      await api.deleteAccount(account.id)
      onDelete(account.id)
      toast.success(`${account.email} removed`)
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Could not remove account")
    }
  }

  if (accounts.length === 0) {
    return (
      <div className="glass flex flex-col items-center rounded-3xl px-6 py-14 text-center">
        <div className="mb-4 flex size-14 items-center justify-center rounded-2xl bg-primary/12 ring-1 ring-primary/20">
          <Lock className="size-6 text-primary" />
        </div>
        <h3 className="text-lg font-semibold">No Google accounts yet</h3>
        <p className="mt-2 max-w-md text-sm leading-relaxed text-muted-foreground">
          Add an account in the gplaydl Authenticator app. It will appear here automatically.
        </p>
      </div>
    )
  }

  return (
    <div className="grid gap-4 md:grid-cols-2">
      {accounts.map((account) => {
        const proxyFailed = account.proxyConfigured && account.proxyTestStatus === "failed"
        const credentialFailed = account.status !== "active" || account.failureCount >= 5
        const needsAttention = credentialFailed || proxyFailed
        return (
          <article key={account.id} className="glass card-hover rounded-3xl p-5 sm:p-6">
            <div className="flex items-start justify-between gap-4">
              <div className="min-w-0">
                <p className="truncate text-base font-semibold">{account.email}</p>
                <div className="mt-2 flex flex-wrap items-center gap-2">
                  <Badge
                    variant="outline"
                    className={
                      needsAttention
                        ? "rounded-full border-chart-4/40 bg-chart-4/10 text-chart-4"
                        : "rounded-full border-primary/35 bg-primary/10 text-primary"
                    }
                  >
                    {needsAttention ? (
                      <AlertTriangle className="mr-1 size-3" />
                    ) : (
                      <CheckCircle2 className="mr-1 size-3" />
                    )}
                    {proxyFailed ? "Proxy check failed" : credentialFailed ? "Needs sign-in" : "Healthy"}
                  </Badge>
                  <span className="flex items-center gap-1 text-xs text-muted-foreground">
                    <Clock3 className="size-3" />
                    {relativeTime(account.lastUsedAt)}
                  </span>
                </div>
              </div>
              <Button
                variant="ghost"
                size="icon"
                aria-label={`Remove ${account.email}`}
                onClick={() => remove(account)}
                className="size-9 shrink-0 rounded-xl text-muted-foreground hover:bg-destructive/10 hover:text-destructive"
              >
                <Trash2 className="size-4" />
              </Button>
            </div>

            {needsAttention && (
              <div className="mt-4 rounded-2xl border border-chart-4/25 bg-chart-4/8 p-3 text-sm text-muted-foreground">
                {proxyFailed
                  ? `The assigned proxy has ${account.proxyFailureCount} consecutive connection ${
                      account.proxyFailureCount === 1 ? "failure" : "failures"
                    }. Token minting will retry it before using the production fallback.`
                  : "Open the Android app and add this account again to refresh its Google token."}
              </div>
            )}

            <div className="mt-5 flex items-center gap-3 rounded-2xl bg-muted/35 p-4">
              <div className="flex size-9 shrink-0 items-center justify-center rounded-xl bg-background/55">
                <Lock className="size-4 text-muted-foreground" />
              </div>
              <div className="min-w-0">
                <p className="text-sm font-medium">Private to you</p>
                <p className="truncate text-xs text-muted-foreground">
                  {`${account.mintCount} downloads served`}
                </p>
              </div>
            </div>
          </article>
        )
      })}
    </div>
  )
}
