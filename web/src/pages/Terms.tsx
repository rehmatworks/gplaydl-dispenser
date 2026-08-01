import { PolicyLayout } from "@/components/PolicyLayout"
import { Card, CardContent } from "@/components/ui/card"
import { ShieldAlert } from "lucide-react"

const terms = [
  {
    title: "Use only accounts you control",
    body: "Sign in with Google accounts that belong to you. Do not upload or use another person's credentials.",
  },
  {
    title: "Free and lawfully acquired apps",
    body: "Use gplaydl only for apps you are allowed to access. You remain responsible for publisher terms, local law, and any licenses that apply to downloaded files.",
  },
  {
    title: "Remove access whenever you like",
    body: "Deleting an account in the Authenticator app erases it from the dispenser and stops future sessions. Sessions already issued are short-lived and expire on their own.",
  },
  {
    title: "Best-effort service",
    body: "The service is provided without an uptime promise or warranty. Google can change its private Play protocol, and the service may change or stop working without notice.",
  },
]

export default function Terms() {
  return (
    <PolicyLayout
      title="Terms of use"
      description="A short, plain-language agreement for using gplaydl, the Authenticator app, and the dispenser."
    >
      <div
        className="flex items-start gap-4 rounded-2xl border border-chart-4/30 bg-chart-4/8 p-5"
        role="note"
      >
        <ShieldAlert className="mt-0.5 size-5 shrink-0 text-chart-4" />
        <div>
          <h2 className="font-semibold">Account risk</h2>
          <p className="mt-1 text-sm leading-relaxed text-muted-foreground">
            This software uses unofficial Google Play access. Google may flag, rate-limit,
            lock, or restrict accounts used with unofficial clients. Please use a separate
            account without payment methods and continue at your own risk.
          </p>
        </div>
      </div>

      {terms.map((term) => (
        <Card key={term.title} className="glass rounded-2xl border-0">
          <CardContent className="p-5 sm:p-6">
            <h2 className="font-semibold">{term.title}</h2>
            <p className="mt-2 text-sm leading-relaxed text-muted-foreground">{term.body}</p>
          </CardContent>
        </Card>
      ))}

      <p className="pt-2 text-sm leading-relaxed text-muted-foreground">
        gplaydl is an independent open-source project and is not affiliated with Google.
      </p>
    </PolicyLayout>
  )
}
