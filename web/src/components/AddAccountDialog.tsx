import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { api, ApiError, type Account } from "@/lib/api"
import { cn } from "@/lib/utils"
import { Globe2, Lock, Plus } from "lucide-react"
import { useState } from "react"
import { toast } from "sonner"

export function AddAccountDialog({ onAdded }: { onAdded: (account: Account) => void }) {
  const [open, setOpen] = useState(false)
  const [email, setEmail] = useState("")
  const [aasToken, setAasToken] = useState("")
  const [visibility, setVisibility] = useState<"public" | "private">("private")
  const [busy, setBusy] = useState(false)

  async function submit(e: React.FormEvent) {
    e.preventDefault()
    setBusy(true)
    try {
      const res = await api.addAccount(email, aasToken, visibility)
      onAdded(res.account)
      toast.success(`${email} added to your ${visibility} pool`)
      setOpen(false)
      setEmail("")
      setAasToken("")
      setVisibility("private")
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Could not add account")
    } finally {
      setBusy(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button className="btn-aurora rounded-xl">
          <Plus className="size-4" /> Add account
        </Button>
      </DialogTrigger>
      <DialogContent className="glass-strong rounded-3xl border-border sm:max-w-lg">
        <DialogHeader>
          <DialogTitle className="font-display text-xl">Add a Google account</DialogTitle>
          <DialogDescription>
            Use a spare account — not your personal one. Sign in to it with the{" "}
            <a
              href="https://github.com/whyorean/Authenticator/releases"
              target="_blank"
              rel="noreferrer"
              className="text-aurora-teal hover:underline"
            >
              Authenticator app
            </a>{" "}
            and paste the token it shows you. It is encrypted before it is stored, and you can
            remove it anytime.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={submit} className="space-y-5">
          <div className="space-y-2">
            <Label htmlFor="acc-email">Google account email</Label>
            <Input
              id="acc-email"
              type="email"
              required
              placeholder="account@gmail.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="h-11 rounded-xl bg-input/50"
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="acc-token">AAS token</Label>
            <Input
              id="acc-token"
              required
              placeholder="aas_et/…"
              value={aasToken}
              onChange={(e) => setAasToken(e.target.value)}
              className="h-11 rounded-xl bg-input/50 font-mono text-xs"
            />
          </div>

          <div className="space-y-2">
            <Label>Sharing</Label>
            <div className="grid grid-cols-2 gap-3">
              {(
                [
                  {
                    value: "private",
                    icon: Lock,
                    title: "Private",
                    desc: "Only you can mint with it"
                  },
                  {
                    value: "public",
                    icon: Globe2,
                    title: "Public pool",
                    desc: "Shared with the community"
                  }
                ] as const
              ).map((opt) => (
                <button
                  key={opt.value}
                  type="button"
                  onClick={() => setVisibility(opt.value)}
                  className={cn(
                    "glass card-hover rounded-2xl p-4 text-left transition-all",
                    visibility === opt.value && "ring-glow border-aurora-teal/40"
                  )}
                >
                  <opt.icon
                    className={cn(
                      "mb-2 size-5",
                      visibility === opt.value ? "text-aurora-teal" : "text-muted-foreground"
                    )}
                  />
                  <div className="font-display text-sm font-semibold">{opt.title}</div>
                  <div className="mt-0.5 text-xs text-muted-foreground">{opt.desc}</div>
                </button>
              ))}
            </div>
          </div>

          <Button type="submit" disabled={busy} className="btn-aurora h-11 w-full rounded-xl">
            {busy ? "Adding…" : "Add account"}
          </Button>
        </form>
      </DialogContent>
    </Dialog>
  )
}
