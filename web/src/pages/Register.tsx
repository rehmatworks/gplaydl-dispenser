import { Logo } from "@/components/Logo"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { api, ApiError } from "@/lib/api"
import { useAuth } from "@/lib/auth"
import { useState } from "react"
import { Link, useNavigate } from "react-router-dom"
import { toast } from "sonner"

export default function Register() {
  const navigate = useNavigate()
  const { setUser } = useAuth()
  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const [busy, setBusy] = useState(false)

  async function submit(e: React.FormEvent) {
    e.preventDefault()
    setBusy(true)
    try {
      const res = await api.register(email, password)
      setUser(res.user)
      // Shown once on the dashboard; the server only stores a hash.
      sessionStorage.setItem("freshApiKey", res.apiKey)
      navigate("/dashboard")
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Could not create account")
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
            <CardTitle className="font-display text-2xl">Create your account</CardTitle>
            <p className="text-sm text-muted-foreground">
              Pool Google accounts, mint tokens, help the community
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
              <div className="space-y-2">
                <Label htmlFor="password">Password</Label>
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
              <Button type="submit" disabled={busy} className="btn-aurora h-11 w-full rounded-xl">
                {busy ? "Creating…" : "Create account"}
              </Button>
            </form>
            <p className="mt-5 text-center text-sm text-muted-foreground">
              Already registered?{" "}
              <Link to="/login" className="text-aurora-teal hover:underline">
                Sign in
              </Link>
            </p>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
