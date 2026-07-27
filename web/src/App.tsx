import { useEffect } from "react"
import { useAuth } from "./lib/auth"
import Dashboard from "./pages/Dashboard"
import Landing from "./pages/Landing"
import Pair from "./pages/Pair"

function ProtectedDashboard() {
  const { user, loading } = useAuth()

  useEffect(() => {
    if (!loading && !user) window.location.replace("/pair")
  }, [loading, user])

  if (loading || !user) {
    return (
      <div className="flex min-h-dvh items-center justify-center">
        <div className="size-8 animate-spin rounded-full border-2 border-primary border-t-transparent" />
      </div>
    )
  }
  return <Dashboard />
}

export default function App() {
  switch (window.location.pathname.replace(/\/+$/, "") || "/") {
    case "/pair":
      return <Pair />
    case "/dashboard":
      return <ProtectedDashboard />
    default:
      return <Landing />
  }
}
