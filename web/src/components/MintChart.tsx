import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import type { MintBucket } from "@/lib/api"
import {
  Area,
  AreaChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis
} from "recharts"

export function MintChart({ timeline }: { timeline: MintBucket[] }) {
  const data = timeline.map((b) => ({
    ...b,
    label: new Date(b.hour).toLocaleTimeString([], { hour: "2-digit" })
  }))

  return (
    <Card className="glass rounded-2xl border-0">
      <CardHeader className="pb-0">
        <CardTitle className="font-display text-base font-semibold">
          Token mints — last 24 hours
        </CardTitle>
      </CardHeader>
      <CardContent className="h-64 p-2 pt-4">
        <ResponsiveContainer width="100%" height="100%">
          <AreaChart data={data} margin={{ top: 4, right: 16, left: -18, bottom: 0 }}>
            <defs>
              <linearGradient id="mintFill" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor="var(--color-chart-1)" stopOpacity={0.45} />
                <stop offset="100%" stopColor="var(--color-chart-1)" stopOpacity={0} />
              </linearGradient>
              <linearGradient id="failFill" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor="var(--color-chart-5)" stopOpacity={0.4} />
                <stop offset="100%" stopColor="var(--color-chart-5)" stopOpacity={0} />
              </linearGradient>
            </defs>
            <CartesianGrid stroke="oklch(0.94 0.01 260 / 7%)" vertical={false} />
            <XAxis
              dataKey="label"
              tick={{ fill: "oklch(0.65 0.02 262)", fontSize: 11 }}
              tickLine={false}
              axisLine={false}
              interval={3}
            />
            <YAxis
              tick={{ fill: "oklch(0.65 0.02 262)", fontSize: 11 }}
              tickLine={false}
              axisLine={false}
              allowDecimals={false}
            />
            <Tooltip
              contentStyle={{
                background: "oklch(0.2 0.025 268 / 95%)",
                border: "1px solid oklch(0.94 0.01 260 / 12%)",
                borderRadius: "0.75rem",
                color: "oklch(0.94 0.01 260)",
                fontSize: 12
              }}
            />
            <Area
              type="monotone"
              dataKey="success"
              name="Minted"
              stroke="var(--color-chart-1)"
              strokeWidth={2}
              fill="url(#mintFill)"
            />
            <Area
              type="monotone"
              dataKey="failures"
              name="Failed"
              stroke="var(--color-chart-5)"
              strokeWidth={2}
              fill="url(#failFill)"
            />
          </AreaChart>
        </ResponsiveContainer>
      </CardContent>
    </Card>
  )
}
