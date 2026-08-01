import { PolicyLayout } from "@/components/PolicyLayout"
import { Card, CardContent } from "@/components/ui/card"

const sections = [
  {
    title: "What the dispenser stores",
    body: "The dispenser stores the Google account address and its AAS token, encrypted at rest with AES-256-GCM. It also keeps basic health information, such as how often an account has been used and when it was last used.",
  },
  {
    title: "What stays on your phone",
    body: "Google sign-in happens inside the Authenticator app. Your password, two-factor codes, and Google cookies are not sent to the dispenser. Only the resulting Play credential and account address are uploaded.",
  },
  {
    title: "Your accounts stay private",
    body: "An account you add is used only for your own paired gplaydl clients. It is never handed to another user and never enters a shared account pool.",
  },
  {
    title: "Browser and device sessions",
    body: "The Android app receives an anonymous device identity. A paired browser stores an encrypted, HTTP-only session cookie so it can access only the accounts connected to that identity.",
  },
  {
    title: "Operational logs",
    body: "IP addresses may appear in short-lived server logs and rate limiting. They are not written to the account database or attached to a Google account record.",
  },
  {
    title: "Deletion",
    body: "Removing an account erases its address and encrypted token from the database. Aggregate service-health counters may remain, with no account information left to connect them to you.",
  },
]

export default function Privacy() {
  return (
    <PolicyLayout
      title="Privacy policy"
      description="What gplaydl stores, what stays on your device, and how you can remove your data."
    >
      {sections.map((section) => (
        <Card key={section.title} className="glass rounded-2xl border-0">
          <CardContent className="p-5 sm:p-6">
            <h2 className="font-semibold">{section.title}</h2>
            <p className="mt-2 text-sm leading-relaxed text-muted-foreground">
              {section.body}
            </p>
          </CardContent>
        </Card>
      ))}

      <p className="pt-2 text-sm leading-relaxed text-muted-foreground">
        Questions about this policy can be raised in the{" "}
        <a
          className="text-primary hover:underline"
          href="https://github.com/rehmatworks/gplaydl-dispenser/issues"
        >
          public issue tracker
        </a>
        .
      </p>
    </PolicyLayout>
  )
}
