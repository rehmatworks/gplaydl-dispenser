import { StrictMode } from "react"
import { createRoot } from "react-dom/client"
import { Toaster } from "sonner"
import App from "./App"
import "./index.css"
import { AuthProvider } from "./lib/auth"

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <AuthProvider>
      <App />
      <Toaster
          theme="dark"
          position="bottom-right"
          toastOptions={{
            style: {
              background: "oklch(0.2 0.025 268 / 90%)",
              backdropFilter: "blur(18px)",
              border: "1px solid oklch(0.94 0.01 260 / 12%)",
              color: "oklch(0.94 0.01 260)"
            }
          }}
      />
    </AuthProvider>
  </StrictMode>
)
