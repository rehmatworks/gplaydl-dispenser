import { Logo } from "@/components/Logo"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle
} from "@/components/ui/card"
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { api, ApiError } from "@/lib/api"
import {
  AlertTriangle,
  ArrowLeft,
  CheckCircle2,
  Eye,
  EyeOff,
  KeyRound,
  RefreshCw,
  Save,
  Server,
  ShieldCheck,
  Trash2
} from "lucide-react"
import { useCallback, useEffect, useState, type FormEvent } from "react"
import { toast } from "sonner"

export default function AdminSettings() {
  const [proxyTemplate, setProxyTemplate] = useState("")
  const [proxyConfigured, setProxyConfigured] = useState(false)
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState(false)
  const [saving, setSaving] = useState(false)
  const [clearing, setClearing] = useState(false)
  const [showTemplate, setShowTemplate] = useState(false)
  const [clearDialogOpen, setClearDialogOpen] = useState(false)

  const loadSettings = useCallback(async () => {
    setLoading(true)
    try {
      const result = await api.adminSettings()
      setProxyConfigured(result.proxyConfigured)
      setLoadError(false)
    } catch {
      setLoadError(true)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    loadSettings()
  }, [loadSettings])

  async function saveProxy(event: FormEvent) {
    event.preventDefault()
    const template = proxyTemplate.trim()
    if (!template) return

    setSaving(true)
    try {
      const wasConfigured = proxyConfigured
      const result = await api.updateProxy(template)
      setProxyConfigured(result.proxyConfigured)
      setProxyTemplate("")
      setShowTemplate(false)
      toast.success(wasConfigured ? "Proxy template replaced" : "Proxy template saved")
    } catch (error) {
      toast.error(error instanceof ApiError ? error.message : "Could not save the proxy template")
    } finally {
      setSaving(false)
    }
  }

  async function clearProxy() {
    setClearing(true)
    try {
      const result = await api.clearProxy()
      setProxyConfigured(result.proxyConfigured)
      setProxyTemplate("")
      setShowTemplate(false)
      setClearDialogOpen(false)
      toast.success("Proxy template cleared")
    } catch (error) {
      toast.error(error instanceof ApiError ? error.message : "Could not clear the proxy template")
    } finally {
      setClearing(false)
    }
  }

  return (
    <div className="min-h-dvh">
      <header className="sticky top-0 z-40 border-b border-border/60 bg-background/80 backdrop-blur-xl">
        <div className="mx-auto flex h-16 max-w-5xl items-center justify-between gap-3 px-4 sm:px-6">
          <a href="/">
            <Logo />
          </a>
          <Button asChild variant="ghost" size="sm" className="rounded-xl">
            <a href="/dashboard">
              <ArrowLeft className="size-4" />
              Dashboard
            </a>
          </Button>
        </div>
      </header>

      <main className="mx-auto max-w-5xl space-y-8 px-4 py-8 sm:px-6 sm:py-10">
        <div className="animate-fade-up">
          <div className="mb-3 flex items-center gap-2 text-sm font-medium text-primary">
            <ShieldCheck className="size-4" />
            Administrator
          </div>
          <h1 className="text-3xl font-bold sm:text-4xl">Proxy settings</h1>
          <p className="mt-3 max-w-2xl text-sm leading-relaxed text-muted-foreground sm:text-base">
            Configure an optional proxy template for download traffic. Leave it unconfigured to
            use the server&apos;s direct connection.
          </p>
        </div>

        {loadError && (
          <div className="flex flex-wrap items-center gap-3 rounded-2xl border border-chart-4/30 bg-chart-4/8 p-4 text-sm">
            <AlertTriangle className="size-5 shrink-0 text-chart-4" />
            <span className="min-w-60 flex-1 text-muted-foreground">
              The current proxy status could not be loaded.
            </span>
            <Button
              variant="outline"
              size="sm"
              onClick={loadSettings}
              disabled={loading}
              className="rounded-xl"
            >
              <RefreshCw className={`size-4 ${loading ? "animate-spin" : ""}`} />
              Retry
            </Button>
          </div>
        )}

        <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_18rem]">
          <Card className="glass-strong gap-0 rounded-3xl border-0">
            <CardHeader className="border-b border-border/70 pb-6">
              <div className="flex flex-wrap items-start justify-between gap-4">
                <div className="space-y-2">
                  <CardTitle className="flex items-center gap-2 text-xl">
                    <Server className="size-5 text-primary" />
                    Proxy template
                  </CardTitle>
                  <CardDescription>
                    Add credentials, host, port, and an optional random port range.
                  </CardDescription>
                </div>
                {loading ? (
                  <Badge variant="outline" className="border-border bg-muted/40 text-muted-foreground">
                    Checking…
                  </Badge>
                ) : proxyConfigured ? (
                  <Badge
                    variant="outline"
                    className="border-primary/30 bg-primary/10 text-primary"
                  >
                    <CheckCircle2 />
                    Configured
                  </Badge>
                ) : (
                  <Badge variant="outline" className="border-border bg-muted/40 text-muted-foreground">
                    Not configured
                  </Badge>
                )}
              </div>
            </CardHeader>

            <CardContent className="p-6">
              <form onSubmit={saveProxy} className="space-y-6">
                <div className="space-y-2">
                  <Label htmlFor="proxy-template">Proxy URL template</Label>
                  <div className="relative">
                    <Input
                      id="proxy-template"
                      type={showTemplate ? "text" : "password"}
                      autoComplete="off"
                      autoCapitalize="none"
                      spellCheck={false}
                      value={proxyTemplate}
                      onChange={(event) => setProxyTemplate(event.target.value)}
                      placeholder={
                        proxyConfigured
                          ? "Enter a new template to replace the current one"
                          : "Enter a proxy URL template"
                      }
                      disabled={loading || loadError || saving || clearing}
                      className="h-12 rounded-xl bg-input/40 pr-12 font-mono"
                      aria-describedby="proxy-template-help"
                    />
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon-sm"
                      className="absolute top-2 right-2 rounded-lg text-muted-foreground"
                      onClick={() => setShowTemplate((visible) => !visible)}
                      disabled={!proxyTemplate}
                      aria-label={showTemplate ? "Hide proxy template" : "Show proxy template"}
                    >
                      {showTemplate ? <EyeOff /> : <Eye />}
                    </Button>
                  </div>
                  <p id="proxy-template-help" className="text-xs leading-relaxed text-muted-foreground">
                    The saved value is never displayed again. Entering a new value replaces the
                    existing template.
                  </p>
                </div>

                <div className="rounded-2xl border border-border bg-background/35 p-4">
                  <p className="mb-2 text-xs font-medium tracking-wide text-muted-foreground uppercase">
                    Example
                  </p>
                  <code className="block overflow-x-auto font-mono text-sm text-foreground">
                    {"http://username:password@host:{rand_int:10001-49000}"}
                  </code>
                  <p className="mt-3 text-xs leading-relaxed text-muted-foreground">
                    The random integer placeholder selects a port from the inclusive range when
                    each account receives its assignment.
                  </p>
                </div>

                <div className="flex flex-col-reverse gap-3 border-t border-border pt-6 sm:flex-row sm:justify-between">
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => setClearDialogOpen(true)}
                    disabled={!proxyConfigured || loading || loadError || saving || clearing}
                    className="rounded-xl text-destructive hover:text-destructive"
                  >
                    <Trash2 className="size-4" />
                    Clear proxy
                  </Button>
                  <Button
                    type="submit"
                    disabled={
                      !proxyTemplate.trim() || loading || loadError || saving || clearing
                    }
                    className="btn-aurora rounded-xl"
                  >
                    <Save className="size-4" />
                    {saving
                      ? "Saving…"
                      : proxyConfigured
                        ? "Replace template"
                        : "Save template"}
                  </Button>
                </div>
              </form>
            </CardContent>
          </Card>

          <div className="space-y-4">
            <Card className="glass gap-4 rounded-2xl border-0">
              <CardHeader className="pb-0">
                <div className="flex size-10 items-center justify-center rounded-xl bg-primary/10">
                  <KeyRound className="size-4 text-primary" />
                </div>
                <CardTitle className="text-base">Keep it private</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-sm leading-relaxed text-muted-foreground">
                  Proxy URLs often contain credentials. The field is masked by default and the
                  current template is not returned by the API.
                </p>
              </CardContent>
            </Card>

            <Card className="glass gap-4 rounded-2xl border-0">
              <CardHeader className="pb-0">
                <div className="flex size-10 items-center justify-center rounded-xl bg-aurora-violet/10">
                  <Server className="size-4 text-aurora-violet" />
                </div>
                <CardTitle className="text-base">Optional by design</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-sm leading-relaxed text-muted-foreground">
                  Clearing the template stops assignments for new accounts. Existing account
                  assignments remain unchanged.
                </p>
              </CardContent>
            </Card>
          </div>
        </div>
      </main>

      <Dialog open={clearDialogOpen} onOpenChange={setClearDialogOpen}>
        <DialogContent className="glass-strong rounded-2xl">
          <DialogHeader>
            <DialogTitle>Clear the proxy template?</DialogTitle>
            <DialogDescription>
              New accounts will not receive a proxy until another template is saved. Existing
              account assignments remain unchanged.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <DialogClose asChild>
              <Button variant="outline" className="rounded-xl" disabled={clearing}>
                Cancel
              </Button>
            </DialogClose>
            <Button
              variant="destructive"
              className="rounded-xl"
              onClick={clearProxy}
              disabled={clearing}
            >
              <Trash2 className="size-4" />
              {clearing ? "Clearing…" : "Clear proxy"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
