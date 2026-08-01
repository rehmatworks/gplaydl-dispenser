import { Logo } from "@/components/Logo"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { api, ApiError } from "@/lib/api"
import { useAuth } from "@/lib/auth"
import { Smartphone } from "lucide-react"
import { useEffect, useState } from "react"
import { toast } from "sonner"

/**
 * Opens the dashboard for a phone-enrolled user. They have no password, so the
 * app shows a one-shot code and this trades it for a session.
 */
export default function Pair() {
  const { user, loading, setUser } = useAuth()
  const [code, setCode] = useState("")
  const [busy, setBusy] = useState(false)

  useEffect(() => {
    if (!loading && user) window.location.replace("/dashboard")
  }, [loading, user])

  async function submit(e: React.FormEvent) {
    e.preventDefault()
    setBusy(true)
    try {
      const res = await api.claimPairing(code)
      setUser(res.user)
      window.location.assign("/dashboard")
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Could not pair this device")
    } finally {
      setBusy(false)
    }
  }

  if (loading || user) {
    return (
      <div className="flex min-h-dvh items-center justify-center">
        <div className="size-8 animate-spin rounded-full border-2 border-primary border-t-transparent" />
      </div>
    )
  }

  return (
    <div className="relative flex min-h-dvh items-center justify-center px-6">
      <div className="aurora-bg" />
      <div className="w-full max-w-md">
        <a href="/" className="mb-8 flex justify-center">
          <Logo />
        </a>
        <Card className="glass-strong animate-fade-up rounded-3xl border-0">
          <CardHeader className="pb-2 text-center">
            <div className="mx-auto mb-3 flex size-12 items-center justify-center rounded-2xl bg-gradient-to-br from-aurora-teal/20 to-aurora-violet/20 ring-1 ring-aurora-teal/20">
              <Smartphone className="size-5 text-aurora-teal" />
            </div>
            <CardTitle className="font-display text-2xl">Pair your phone</CardTitle>
            <p className="text-sm text-muted-foreground">
              In the Authenticator app, open the Link gplaydl tab and type the code it shows.
            </p>
          </CardHeader>
          <CardContent className="p-6 pt-4">
            <form onSubmit={submit} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="code">Pairing code</Label>
                <Input
                  id="code"
                  required
                  autoFocus
                  autoComplete="one-time-code"
                  placeholder="ABCD 2345"
                  value={code}
                  onChange={(e) => setCode(e.target.value.toUpperCase())}
                  className="h-14 rounded-xl bg-input/50 text-center font-mono text-2xl tracking-[0.3em]"
                />
              </div>
              <Button
                type="submit"
                disabled={busy || code.replace(/[\s-]/g, "").length < 8}
                className="btn-aurora h-11 w-full rounded-xl"
              >
                {busy ? "Pairing…" : "Open my dashboard"}
              </Button>
            </form>
            <p className="mt-5 text-center text-sm text-muted-foreground">
              Codes expire after 10 minutes and work once. Generate a new code in the Android
              app whenever you need to reconnect.
            </p>
            <p className="mt-3 text-center text-xs leading-relaxed text-muted-foreground">
              By pairing, you agree to the{" "}
              <a href="/terms" className="text-primary hover:underline">
                terms
              </a>{" "}
              and acknowledge the{" "}
              <a href="/privacy" className="text-primary hover:underline">
                privacy policy
              </a>
              .
            </p>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
