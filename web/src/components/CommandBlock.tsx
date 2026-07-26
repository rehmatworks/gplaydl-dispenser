import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"
import { Check, Copy } from "lucide-react"
import { useState } from "react"

interface Props {
  /** Shell command shown verbatim and copied to the clipboard. */
  command: string
  label?: string
  className?: string
}

/** A copyable one-liner. Used wherever the site tells someone what to run. */
export function CommandBlock({ command, label, className }: Props) {
  const [copied, setCopied] = useState(false)

  async function copy() {
    await navigator.clipboard.writeText(command)
    setCopied(true)
    setTimeout(() => setCopied(false), 1600)
  }

  return (
    <div className={cn("space-y-2", className)}>
      {label && <p className="text-xs font-medium text-muted-foreground">{label}</p>}
      <div className="glass flex items-center gap-2 rounded-xl p-1.5 pl-4">
        <code className="min-w-0 flex-1 overflow-x-auto whitespace-nowrap py-2 font-mono text-xs text-aurora-teal">
          {command}
        </code>
        <Button
          variant="outline"
          size="icon"
          onClick={copy}
          className="glass size-8 shrink-0 rounded-lg"
          aria-label="Copy command"
        >
          {copied ? <Check className="size-3.5 text-aurora-teal" /> : <Copy className="size-3.5" />}
        </Button>
      </div>
    </div>
  )
}
