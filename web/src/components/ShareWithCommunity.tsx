import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { api, ApiError, type Account } from "@/lib/api"
import { HeartHandshake } from "lucide-react"
import { useState } from "react"
import { toast } from "sonner"

interface Props {
  accounts: Account[]
  onChanged: () => void
}

/**
 * One-tap way to put every private account into the community rotation. Only
 * appears when there is actually something to share, so it never nags.
 */
export function ShareWithCommunity({ accounts, onChanged }: Props) {
  const [busy, setBusy] = useState(false)

  const privateAccounts = accounts.filter(
    (a) => a.visibility === "private" && a.status !== "disabled"
  )
  const shared = accounts.filter((a) => a.visibility === "public").length

  if (privateAccounts.length === 0) {
    if (shared === 0) return null
    return (
      <Card className="glass rounded-2xl border-0">
        <CardContent className="flex flex-wrap items-center gap-4 p-6">
          <div className="flex size-11 shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-aurora-teal/20 to-aurora-violet/20 ring-1 ring-aurora-teal/20">
            <HeartHandshake className="size-5 text-aurora-teal" />
          </div>
          <p className="flex-1 text-sm text-muted-foreground">
            <span className="font-medium text-foreground">
              You are sharing {shared} {shared === 1 ? "account" : "accounts"}.
            </span>{" "}
            Thank you — that is what keeps gplaydl working for everyone else.
          </p>
        </CardContent>
      </Card>
    )
  }

  async function shareAll() {
    setBusy(true)
    const results = await Promise.allSettled(
      privateAccounts.map((a) => api.updateAccount(a.id, { visibility: "public" }))
    )
    const failed = results.filter((r) => r.status === "rejected")
    if (failed.length === 0) {
      toast.success(
        `${privateAccounts.length} ${privateAccounts.length === 1 ? "account" : "accounts"} joined the community pool`
      )
    } else {
      const first = failed[0] as PromiseRejectedResult
      toast.error(
        first.reason instanceof ApiError ? first.reason.message : "Some accounts could not be shared"
      )
    }
    setBusy(false)
    onChanged()
  }

  return (
    <Card className="glass card-hover rounded-2xl border border-aurora-teal/25">
      <CardContent className="flex flex-wrap items-center gap-4 p-6">
        <div className="flex size-11 shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-aurora-teal/20 to-aurora-violet/20 ring-1 ring-aurora-teal/20">
          <HeartHandshake className="size-5 text-aurora-teal" />
        </div>
        <div className="min-w-0 flex-1">
          <h3 className="font-display text-base font-semibold">
            {privateAccounts.length} {privateAccounts.length === 1 ? "account is" : "accounts are"}{" "}
            private
          </h3>
          <p className="mt-1 text-sm text-muted-foreground">
            Sharing puts them into the rotation everyone downloads through. Keep any account
            holding purchases private instead.
          </p>
        </div>
        <Button onClick={shareAll} disabled={busy} className="btn-aurora rounded-xl">
          {busy ? "Sharing…" : "Share all with the community"}
        </Button>
      </CardContent>
    </Card>
  )
}
