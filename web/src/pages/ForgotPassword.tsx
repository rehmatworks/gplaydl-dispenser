import { Logo } from "@/components/Logo"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { api, ApiError } from "@/lib/api"
import { MailCheck } from "lucide-react"
import { useState } from "react"
import { Link } from "react-router-dom"
import { toast } from "sonner"

export default function ForgotPassword() {
  const [email, setEmail] = useState("")
  const [busy, setBusy] = useState(false)
  const [sent, setSent] = useState(false)

  async function submit(e: React.FormEvent) {
    e.preventDefault()
    setBusy(true)
    try {
      await api.forgotPassword(email)
      setSent(true)
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Could not send reset email")
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
          {sent ? (
            <CardContent className="flex flex-col items-center gap-4 p-10 text-center">
              <MailCheck className="size-12 text-aurora-teal" />
              <h1 className="font-display text-2xl font-bold">Check your inbox</h1>
              <p className="text-sm text-muted-foreground">
                If <span className="text-foreground">{email}</span> is registered, a password
                reset link is on its way. The link expires in 1 hour.
              </p>
              <Link to="/login" className="text-sm text-aurora-teal hover:underline">
                Back to sign in
              </Link>
            </CardContent>
          ) : (
            <>
              <CardHeader className="pb-2 text-center">
                <CardTitle className="font-display text-2xl">Forgot your password?</CardTitle>
                <p className="text-sm text-muted-foreground">
                  Enter your email and we'll send you a reset link
                </p>
              </CardHeader>
              <CardContent className="p-6 pt-4">
                <form onSubmit={submit} className="space-y-4">
                  <div className="space-y-2">
                    <Label htmlFor="email">Email</Label>
                    <Input
                      id="email"
                      type="email"
                      required
                      autoComplete="email"
                      placeholder="you@example.com"
                      value={email}
                      onChange={(e) => setEmail(e.target.value)}
                      className="h-11 rounded-xl bg-input/50"
                    />
                  </div>
                  <Button type="submit" disabled={busy} className="btn-aurora h-11 w-full rounded-xl">
                    {busy ? "Sending…" : "Send reset link"}
                  </Button>
                </form>
                <p className="mt-5 text-center text-sm text-muted-foreground">
                  Remembered it?{" "}
                  <Link to="/login" className="text-aurora-teal hover:underline">
                    Sign in
                  </Link>
                </p>
              </CardContent>
            </>
          )}
        </Card>
      </div>
    </div>
  )
}
