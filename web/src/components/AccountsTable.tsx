import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Switch } from "@/components/ui/switch"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from "@/components/ui/table"
import { api, ApiError, type Account } from "@/lib/api"
import { cn } from "@/lib/utils"
import { FlaskConical, Globe2, Loader2, Lock, Trash2 } from "lucide-react"
import { useState } from "react"
import { toast } from "sonner"

interface Props {
  accounts: Account[]
  onChange: (account: Account) => void
  onDelete: (id: string) => void
}

const statusStyles: Record<Account["status"], string> = {
  active: "border-aurora-teal/40 bg-aurora-teal/10 text-aurora-teal",
  flagged: "border-chart-4/40 bg-chart-4/10 text-chart-4",
  disabled: "border-muted-foreground/30 bg-muted/40 text-muted-foreground"
}

export function AccountsTable({ accounts, onChange, onDelete }: Props) {
  const [testing, setTesting] = useState<string | null>(null)

  async function toggleVisibility(account: Account) {
    const visibility = account.visibility === "public" ? "private" : "public"
    try {
      const res = await api.updateAccount(account.id, { visibility })
      onChange(res.account)
      toast.success(
        visibility === "public"
          ? `${account.email} is now shared with the community`
          : `${account.email} is now private`
      )
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Update failed")
    }
  }

  async function toggleEnabled(account: Account) {
    const status = account.status === "disabled" ? "active" : "disabled"
    try {
      const res = await api.updateAccount(account.id, { status })
      onChange(res.account)
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Update failed")
    }
  }

  async function test(account: Account) {
    setTesting(account.id)
    try {
      const res = await api.testAccount(account.id)
      if (res.success) {
        toast.success(`${account.email} minted a token in ${(res.durationMs / 1000).toFixed(1)}s`)
      } else {
        toast.error(`${account.email} failed: ${res.error}`)
      }
      const refreshed = await api.accounts()
      const updated = refreshed.accounts.find((a) => a.id === account.id)
      if (updated) onChange(updated)
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Test failed")
    } finally {
      setTesting(null)
    }
  }

  async function remove(account: Account) {
    if (!confirm(`Remove ${account.email} from the dispenser?`)) return
    try {
      await api.deleteAccount(account.id)
      onDelete(account.id)
      toast.success(`${account.email} removed`)
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Delete failed")
    }
  }

  if (accounts.length === 0) {
    return (
      <div className="glass flex flex-col items-center rounded-2xl px-6 py-16 text-center">
        <div className="mb-4 flex size-14 items-center justify-center rounded-2xl bg-gradient-to-br from-aurora-teal/15 to-aurora-violet/15 ring-1 ring-aurora-teal/20">
          <Lock className="size-6 text-aurora-teal" />
        </div>
        <h3 className="font-display text-lg font-semibold">No Google accounts yet</h3>
        <p className="mt-1 max-w-sm text-sm text-muted-foreground">
          Add your first account to start minting tokens. Keep it private, or share it with the
          open-source community.
        </p>
      </div>
    )
  }

  return (
    <div className="glass overflow-hidden rounded-2xl">
      <Table>
        <TableHeader>
          <TableRow className="border-border hover:bg-transparent">
            <TableHead className="pl-5">Account</TableHead>
            <TableHead>Pool</TableHead>
            <TableHead>Status</TableHead>
            <TableHead className="text-right">Mints</TableHead>
            <TableHead>Last used</TableHead>
            <TableHead>Enabled</TableHead>
            <TableHead className="pr-5 text-right">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {accounts.map((account) => (
            <TableRow key={account.id} className="border-border">
              <TableCell className="pl-5">
                <span className="mono-chip">{account.email}</span>
                {account.failureCount > 0 && (
                  <span className="ml-2 text-xs text-chart-4">
                    {account.failureCount} fails
                  </span>
                )}
              </TableCell>
              <TableCell>
                <button
                  onClick={() => toggleVisibility(account)}
                  className={cn(
                    "flex items-center gap-1.5 rounded-full border px-3 py-1 text-xs font-medium transition-all",
                    account.visibility === "public"
                      ? "border-aurora-violet/40 bg-aurora-violet/10 text-aurora-pink hover:bg-aurora-violet/20"
                      : "border-border bg-muted/40 text-muted-foreground hover:bg-muted/70"
                  )}
                  title="Click to toggle sharing"
                >
                  {account.visibility === "public" ? (
                    <>
                      <Globe2 className="size-3" /> Public
                    </>
                  ) : (
                    <>
                      <Lock className="size-3" /> Private
                    </>
                  )}
                </button>
              </TableCell>
              <TableCell>
                <Badge variant="outline" className={cn("rounded-full", statusStyles[account.status])}>
                  {account.status}
                </Badge>
              </TableCell>
              <TableCell className="text-right font-mono text-sm">{account.mintCount}</TableCell>
              <TableCell className="text-sm text-muted-foreground">
                {account.lastUsedAt ? new Date(account.lastUsedAt).toLocaleString() : "never"}
              </TableCell>
              <TableCell>
                <Switch
                  checked={account.status !== "disabled"}
                  onCheckedChange={() => toggleEnabled(account)}
                />
              </TableCell>
              <TableCell className="pr-5">
                <div className="flex justify-end gap-1.5">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => test(account)}
                    disabled={testing !== null}
                    className="glass rounded-lg"
                    title="Mint a test token with this account"
                  >
                    {testing === account.id ? (
                      <Loader2 className="size-3.5 animate-spin" />
                    ) : (
                      <FlaskConical className="size-3.5" />
                    )}
                    Test
                  </Button>
                  <Button
                    variant="outline"
                    size="icon"
                    onClick={() => remove(account)}
                    className="glass size-8 rounded-lg text-destructive hover:bg-destructive/10"
                    title="Remove account"
                  >
                    <Trash2 className="size-3.5" />
                  </Button>
                </div>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
