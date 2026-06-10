import { Logo } from "@/components/Logo"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { api, ApiError } from "@/lib/api"
import { useAuth } from "@/lib/auth"
import { CheckCircle2, XCircle } from "lucide-react"
import { useEffect, useRef, useState } from "react"
import { Link, useSearchParams } from "react-router-dom"

export default function VerifyEmail() {
  const [params] = useSearchParams()
  const { setUser } = useAuth()
  const [state, setState] = useState<"working" | "ok" | "error">("working")
  const [message, setMessage] = useState("")
  const ran = useRef(false)

  useEffect(() => {
    if (ran.current) return
    ran.current = true

    const token = params.get("token") ?? ""
    if (!token) {
      setState("error")
      setMessage("This link is missing its verification token.")
      return
    }

    api
      .verifyEmail(token)
      .then(async () => {
        setState("ok")
        // Refresh the session user so the dashboard banner disappears.
        await api.me().then((res) => setUser(res.user)).catch(() => {})
      })
      .catch((err) => {
        setState("error")
        setMessage(
          err instanceof ApiError ? err.message : "Something went wrong. Please try again."
        )
      })
  }, [params, setUser])

  return (
    <div className="relative flex min-h-dvh items-center justify-center px-6">
      <div className="aurora-bg" />
      <div className="w-full max-w-md">
        <Link to="/" className="mb-8 flex justify-center">
          <Logo />
        </Link>
        <Card className="glass-strong animate-fade-up rounded-3xl border-0">
          <CardContent className="flex flex-col items-center gap-4 p-10 text-center">
            {state === "working" && (
              <>
                <div className="size-10 animate-spin rounded-full border-2 border-primary border-t-transparent" />
                <p className="text-sm text-muted-foreground">Verifying your email…</p>
              </>
            )}
            {state === "ok" && (
              <>
                <CheckCircle2 className="size-12 text-aurora-teal" />
                <h1 className="font-display text-2xl font-bold">Email verified</h1>
                <p className="text-sm text-muted-foreground">
                  You're all set — you can now contribute accounts to the pool.
                </p>
                <Button asChild className="btn-aurora mt-2 h-11 rounded-xl px-8">
                  <Link to="/dashboard">Go to dashboard</Link>
                </Button>
              </>
            )}
            {state === "error" && (
              <>
                <XCircle className="size-12 text-destructive" />
                <h1 className="font-display text-2xl font-bold">Verification failed</h1>
                <p className="text-sm text-muted-foreground">{message}</p>
                <p className="text-sm text-muted-foreground">
                  You can request a fresh link from your{" "}
                  <Link to="/dashboard" className="text-aurora-teal hover:underline">
                    dashboard
                  </Link>
                  .
                </p>
              </>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
