// Package mail sends transactional email through the Brevo API.
package mail

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

const brevoEndpoint = "https://api.brevo.com/v3/smtp/email"

type Mailer struct {
	apiKey    string
	fromEmail string
	fromName  string
	publicURL string
	http      *http.Client
	log       *slog.Logger
}

func New(apiKey, fromEmail, fromName, publicURL string, log *slog.Logger) *Mailer {
	return &Mailer{
		apiKey:    apiKey,
		fromEmail: fromEmail,
		fromName:  fromName,
		publicURL: publicURL,
		http:      &http.Client{Timeout: 15 * time.Second},
		log:       log,
	}
}

// Enabled reports whether a Brevo API key is configured. When disabled, the
// app skips email flows (new users are auto-verified).
func (m *Mailer) Enabled() bool { return m.apiKey != "" }

func (m *Mailer) SendVerification(ctx context.Context, to, token string) error {
	link := fmt.Sprintf("%s/verify-email?token=%s", m.publicURL, token)
	html := layout(
		"Confirm your email",
		"Thanks for joining the community pool. Click the button below to verify your email address and unlock contributing accounts.",
		link,
		"Verify email",
		"This link expires in 24 hours. If you didn't create an account, you can safely ignore this email.",
	)
	return m.send(ctx, to, "Verify your email — gplaydl dispenser", html)
}

func (m *Mailer) SendPasswordReset(ctx context.Context, to, token string) error {
	link := fmt.Sprintf("%s/reset-password?token=%s", m.publicURL, token)
	html := layout(
		"Reset your password",
		"We received a request to reset your password. Click the button below to choose a new one.",
		link,
		"Reset password",
		"This link expires in 1 hour. If you didn't request a reset, you can safely ignore this email — your password is unchanged.",
	)
	return m.send(ctx, to, "Reset your password — gplaydl dispenser", html)
}

type brevoRequest struct {
	Sender      brevoAddress   `json:"sender"`
	To          []brevoAddress `json:"to"`
	Subject     string         `json:"subject"`
	HTMLContent string         `json:"htmlContent"`
}

type brevoAddress struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

func (m *Mailer) send(ctx context.Context, to, subject, html string) error {
	if !m.Enabled() {
		return fmt.Errorf("mailer disabled: BREVO_API_KEY not set")
	}

	body, err := json.Marshal(brevoRequest{
		Sender:      brevoAddress{Email: m.fromEmail, Name: m.fromName},
		To:          []brevoAddress{{Email: to}},
		Subject:     subject,
		HTMLContent: html,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, brevoEndpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("api-key", m.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := m.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("brevo API returned %d: %s", resp.StatusCode, detail)
	}
	return nil
}

// SendAsync fires an email in the background so HTTP handlers never block on
// Brevo. Failures are logged, not surfaced to the caller.
func (m *Mailer) SendAsync(kind string, fn func(ctx context.Context) error) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		if err := fn(ctx); err != nil {
			m.log.Error("send email", "kind", kind, "err", err)
		}
	}()
}

func layout(title, intro, link, cta, footer string) string {
	return fmt.Sprintf(`<!doctype html>
<html>
  <body style="margin:0;padding:0;background:#0d0e16;font-family:-apple-system,'Segoe UI',Roboto,Helvetica,Arial,sans-serif;">
    <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="background:#0d0e16;padding:40px 16px;">
      <tr><td align="center">
        <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="max-width:480px;background:#171927;border:1px solid rgba(255,255,255,0.08);border-radius:16px;padding:36px;">
          <tr><td style="text-align:center;padding-bottom:24px;">
            <span style="font-size:18px;font-weight:700;color:#e8e9f1;">gplaydl<span style="color:#4fd1c5;">&middot;dispenser</span></span>
          </td></tr>
          <tr><td style="color:#e8e9f1;font-size:20px;font-weight:700;padding-bottom:12px;">%s</td></tr>
          <tr><td style="color:#9a9db1;font-size:14px;line-height:1.6;padding-bottom:28px;">%s</td></tr>
          <tr><td align="center" style="padding-bottom:28px;">
            <a href="%s" style="display:inline-block;background:linear-gradient(110deg,#4fd1c5,#8b5cf6);color:#0d0e16;font-size:14px;font-weight:600;text-decoration:none;padding:12px 28px;border-radius:12px;">%s</a>
          </td></tr>
          <tr><td style="color:#6b6e80;font-size:12px;line-height:1.6;border-top:1px solid rgba(255,255,255,0.08);padding-top:20px;">%s<br/><br/>If the button doesn't work, copy this link:<br/><a href="%s" style="color:#4fd1c5;word-break:break-all;">%s</a></td></tr>
        </table>
      </td></tr>
    </table>
  </body>
</html>`, title, intro, link, cta, footer, link, link)
}
