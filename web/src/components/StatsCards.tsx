import { Card, CardContent } from "@/components/ui/card"
import type { PoolStats } from "@/lib/api"
import { Activity, Globe2, KeyRound, ShieldCheck } from "lucide-react"

export function StatsCards({ stats }: { stats: PoolStats | null }) {
  const successRate =
    stats && stats.mints24h + stats.failures24h > 0
      ? Math.round((stats.mints24h / (stats.mints24h + stats.failures24h)) * 100) + "%"
      : "—"

  const items = [
    {
      icon: ShieldCheck,
      label: "Your active accounts",
      value: stats?.activeAccounts ?? "—",
      sub: stats ? `${stats.flaggedAccounts} flagged` : ""
    },
    {
      icon: Globe2,
      label: "Public pool",
      value: stats?.publicAccounts ?? "—",
      sub: "community accounts"
    },
    {
      icon: KeyRound,
      label: "Mints (24h)",
      value: stats?.mints24h ?? "—",
      sub: stats ? `${stats.failures24h} failed` : ""
    },
    {
      icon: Activity,
      label: "Success rate (24h)",
      value: successRate,
      sub: stats ? `${stats.totalMints} all-time` : ""
    }
  ]

  return (
    <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
      {items.map((s) => (
        <Card key={s.label} className="glass card-hover rounded-2xl border-0 py-0">
          <CardContent className="flex items-center gap-4 p-5">
            <div className="flex size-11 shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-aurora-teal/15 to-aurora-violet/15 ring-1 ring-aurora-teal/20">
              <s.icon className="size-5 text-aurora-teal" />
            </div>
            <div className="min-w-0">
              <div className="font-display text-2xl font-bold leading-tight">{s.value}</div>
              <div className="truncate text-xs text-muted-foreground">
                {s.label}
                {s.sub ? <span className="opacity-60"> · {s.sub}</span> : null}
              </div>
            </div>
          </CardContent>
        </Card>
      ))}
    </div>
  )
}
