import { Logo } from "@/components/Logo"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { api, ApiError } from "@/lib/api"
import { useState } from "react"
import { Link, useNavigate, useSearchParams } from "react-router-dom"
import { toast } from "sonner"

export default function ResetPassword() {
  const [params] = useSearchParams()
  const navigate = useNavigate()
  const token = params.get("token") ?? ""

  const [password, setPassword] = useState("")
  const [confirm, setConfirm] = useState("")
  const [busy, setBusy] = useState(false)

  async function submit(e: React.FormEvent) {
    e.preventDefault()
    if (password !== confirm) {
      toast.error("Passwords don't match")
      return
    }
    setBusy(true)
    try {
      await api.resetPassword(token, password)
      toast.success("Password updated — sign in with your new password")
      navigate("/login")
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Could not reset password")
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="relative flex min-h-dvh items-center justify-center px-6">
      <div className="aurora-bg" />
      <div className="w-full max-w-md">
        <Link to="/" className="mb-8 flex justify-center">
          <Logo />
        </Link>
        <Card className="glass-strong animate-fade-up rounded-3xl border-0">
          <CardHeader className="pb-2 text-center">
            <CardTitle className="font-display text-2xl">Choose a new password</CardTitle>
            <p className="text-sm text-muted-foreground">
              Resetting also signs you out everywhere else
            </p>
          </CardHeader>
          <CardContent className="p-6 pt-4">
            {token ? (
              <form onSubmit={submit} className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="password">New password</Label>
                  <Input
                    id="password"
                    type="password"
                    required
                    minLength={8}
                    autoComplete="new-password"
                    placeholder="At least 8 characters"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    className="h-11 rounded-xl bg-input/50"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="confirm">Confirm password</Label>
                  <Input
                    id="confirm"
                    type="password"
                    required
                    minLength={8}
                    autoComplete="new-password"
                    placeholder="••••••••"
                    value={confirm}
                    onChange={(e) => setConfirm(e.target.value)}
                    className="h-11 rounded-xl bg-input/50"
                  />
                </div>
                <Button type="submit" disabled={busy} className="btn-aurora h-11 w-full rounded-xl">
                  {busy ? "Saving…" : "Reset password"}
                </Button>
              </form>
            ) : (
              <p className="text-center text-sm text-muted-foreground">
                This link is missing its reset token. Request a new one from the{" "}
                <Link to="/forgot-password" className="text-aurora-teal hover:underline">
                  forgot password
                </Link>{" "}
                page.
              </p>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
