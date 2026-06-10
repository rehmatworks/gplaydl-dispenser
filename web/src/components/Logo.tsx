import { cn } from "@/lib/utils"

export function Logo({ className }: { className?: string }) {
  return (
    <div className={cn("flex items-center gap-2.5", className)}>
      <div className="ring-glow flex size-8 items-center justify-center rounded-xl bg-gradient-to-br from-aurora-teal to-aurora-violet">
        <svg viewBox="0 0 24 24" className="size-4.5 fill-background" aria-hidden>
          <path d="M5 3.5c0-1.1 1.2-1.8 2.2-1.2l13 7.5c1 .6 1 2 0 2.6l-13 7.5c-1 .6-2.2-.1-2.2-1.2v-15.2z" />
        </svg>
      </div>
      <span className="font-display text-lg font-700 font-bold tracking-tight">
        gplaydl<span className="text-aurora">·dispenser</span>
      </span>
    </div>
  )
}
