import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { api } from "@/lib/api"
import { Check, Copy, KeyRound, RefreshCcw } from "lucide-react"
import { useEffect, useState } from "react"
import { toast } from "sonner"

export function ApiKeyCard() {
  const [apiKey, setApiKey] = useState<string | null>(null)
  const [copied, setCopied] = useState(false)
  const [busy, setBusy] = useState(false)

  useEffect(() => {
    const fresh = sessionStorage.getItem("freshApiKey")
    if (fresh) setApiKey(fresh)
  }, [])

  async function copy() {
    if (!apiKey) return
    await navigator.clipboard.writeText(apiKey)
    setCopied(true)
    setTimeout(() => setCopied(false), 1600)
  }

  async function rotate() {
    if (
      apiKey === null &&
      !confirm("Rotating invalidates your current API key immediately. Continue?")
    ) {
      return
    }
    setBusy(true)
    try {
      const res = await api.rotateApiKey()
      setApiKey(res.apiKey)
      sessionStorage.setItem("freshApiKey", res.apiKey)
      toast.success("New API key generated")
    } catch {
      toast.error("Could not rotate API key")
    } finally {
      setBusy(false)
    }
  }

  return (
    <Card className="glass rounded-2xl border-0">
      <CardHeader className="pb-2">
        <CardTitle className="flex items-center gap-2 font-display text-base font-semibold">
          <KeyRound className="size-4 text-aurora-teal" /> API key
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4 pt-0">
        <p className="text-sm leading-relaxed text-muted-foreground">
          Use it as <span className="mono-chip">X-Api-Key</span> on{" "}
          <span className="mono-chip">/api/auth</span> to mint from your private accounts and skip
          anonymous rate limits.
        </p>

        {apiKey ? (
          <div className="flex items-center gap-2">
            <code className="glass min-w-0 flex-1 truncate rounded-xl px-3 py-2.5 font-mono text-xs text-aurora-teal">
              {apiKey}
            </code>
            <Button variant="outline" size="icon" onClick={copy} className="glass size-9 shrink-0 rounded-xl">
              {copied ? <Check className="size-4 text-aurora-teal" /> : <Copy className="size-4" />}
            </Button>
          </div>
        ) : (
          <p className="text-xs text-muted-foreground">
            Your key is only shown once. Lost it? Rotate to get a new one.
          </p>
        )}

        <Button
          variant="outline"
          onClick={rotate}
          disabled={busy}
          className="glass w-full rounded-xl"
        >
          <RefreshCcw className="size-4" /> {busy ? "Rotating…" : "Rotate key"}
        </Button>
      </CardContent>
    </Card>
  )
}
