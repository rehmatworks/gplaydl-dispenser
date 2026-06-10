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

export default function Login() {
  const navigate = useNavigate()
  const { setUser } = useAuth()
  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const [busy, setBusy] = useState(false)

  async function submit(e: React.FormEvent) {
    e.preventDefault()
    setBusy(true)
    try {
      const res = await api.login(email, password)
      setUser(res.user)
      navigate("/dashboard")
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Could not sign in")
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
            <CardTitle className="font-display text-2xl">Welcome back</CardTitle>
            <p className="text-sm text-muted-foreground">Sign in to manage your account pool</p>
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
                  autoComplete="current-password"
                  placeholder="••••••••"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  className="h-11 rounded-xl bg-input/50"
                />
              </div>
              <Button type="submit" disabled={busy} className="btn-aurora h-11 w-full rounded-xl">
                {busy ? "Signing in…" : "Sign in"}
              </Button>
            </form>
            <p className="mt-5 text-center text-sm text-muted-foreground">
              No account yet?{" "}
              <Link to="/register" className="text-aurora-teal hover:underline">
                Create one
              </Link>
            </p>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
