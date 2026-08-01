import { useEffect } from "react"
import { useAuth } from "./lib/auth"
import Dashboard from "./pages/Dashboard"
import Landing from "./pages/Landing"
import Pair from "./pages/Pair"
import AdminSettings from "./pages/AdminSettings"
import Privacy from "./pages/Privacy"
import Terms from "./pages/Terms"

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

function ProtectedAdminSettings() {
  const { user, loading } = useAuth()

  useEffect(() => {
    if (!loading && (!user || !user.isAdmin)) window.location.replace("/dashboard")
  }, [loading, user])

  if (loading || !user?.isAdmin) {
    return (
      <div className="flex min-h-dvh items-center justify-center">
        <div className="size-8 animate-spin rounded-full border-2 border-primary border-t-transparent" />
      </div>
    )
  }
  return <AdminSettings />
}

export default function App() {
  switch (window.location.pathname.replace(/\/+$/, "") || "/") {
    case "/pair":
      return <Pair />
    case "/dashboard":
      return <ProtectedDashboard />
    case "/admin/settings":
      return <ProtectedAdminSettings />
    case "/terms":
      return <Terms />
    case "/privacy":
      return <Privacy />
    default:
      return <Landing />
  }
}
