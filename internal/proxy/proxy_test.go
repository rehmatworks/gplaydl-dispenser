package proxy

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestExpand(t *testing.T) {
	t.Run("concrete URL is unchanged", func(t *testing.T) {
		const concrete = "https://user:p%40ss@proxy.example:8443"
		got, err := Expand(concrete)
		if err != nil {
			t.Fatalf("Expand() error = %v", err)
		}
		if got != concrete {
			t.Fatalf("Expand() = %q, want %q", got, concrete)
		}
	})

	t.Run("fixed range uses public crypto path", func(t *testing.T) {
		got, err := Expand("http://proxy.example:{rand_int:8123-8123}")
		if err != nil {
			t.Fatalf("Expand() error = %v", err)
		}
		if got != "http://proxy.example:8123" {
			t.Fatalf("Expand() = %q", got)
		}
	})

	t.Run("range is inclusive", func(t *testing.T) {
		template := "http://proxy.example:{rand_int:2000-2002}"
		for _, endpoint := range []uint64{2000, 2002} {
			got, err := expand(template, func(min, max uint64) (uint64, error) {
				if min != 2000 || max != 2002 {
					t.Fatalf("chooser range = %d-%d", min, max)
				}
				return endpoint, nil
			})
			if err != nil {
				t.Fatalf("expand() error = %v", err)
			}
			want := fmt.Sprintf("http://proxy.example:%d", endpoint)
			if got != want {
				t.Fatalf("expand() = %q, want %q", got, want)
			}
		}
	})

	t.Run("https credentials and IPv6 are accepted", func(t *testing.T) {
		got, err := expand(
			"https://user:p%40ss@[2001:db8::1]:{rand_int:443-444}",
			func(_, _ uint64) (uint64, error) { return 444, nil },
		)
		if err != nil {
			t.Fatalf("expand() error = %v", err)
		}
		if got != "https://user:p%40ss@[2001:db8::1]:444" {
			t.Fatalf("expand() = %q", got)
		}
	})
}

func TestValidateTemplateRejectsInvalidInputWithoutLeakingIt(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"unsupported scheme", "socks5://user:top-secret@proxy.example:1080"},
		{"missing hostname", "http://user:top-secret@:8080"},
		{"invalid hostname", "http://user:top-secret@bad_host.example:8080"},
		{"missing port", "http://user:top-secret@proxy.example"},
		{"empty port", "http://user:top-secret@proxy.example:"},
		{"zero port", "http://user:top-secret@proxy.example:0"},
		{"large port", "http://user:top-secret@proxy.example:65536"},
		{"non-numeric port", "http://user:top-secret@proxy.example:eighty"},
		{"empty username", "http://:top-secret@proxy.example:8080"},
		{"empty password", "http://user:@proxy.example:8080"},
		{"control in password", "http://user:top%0Asecret@proxy.example:8080"},
		{"path", "http://user:top-secret@proxy.example:8080/path"},
		{"trailing slash", "http://user:top-secret@proxy.example:8080/"},
		{"query", "http://user:top-secret@proxy.example:8080?mode=fast"},
		{"empty query", "http://user:top-secret@proxy.example:8080?"},
		{"fragment", "http://user:top-secret@proxy.example:8080#fragment"},
		{"empty fragment", "http://user:top-secret@proxy.example:8080#"},
		{"malformed shortcode", "http://user:top-secret@proxy.example:{rand_int:1}"},
		{"unknown shortcode", "http://user:top-secret@proxy.example:{random:1-2}"},
		{"shortcode in username", "http://{rand_int:1-2}:top-secret@proxy.example:8080"},
		{"shortcode in password", "http://user:{rand_int:1-2}@proxy.example:8080"},
		{"shortcode in hostname", "http://proxy-{rand_int:1-2}.example:8080"},
		{"shortcode in path", "http://user:top-secret@proxy.example:8080/{rand_int:1-2}"},
		{"shortcode after digits", "http://user:top-secret@proxy.example:8{rand_int:1-2}"},
		{"two shortcodes", "http://{rand_int:1-2}@proxy.example:{rand_int:3-4}"},
		{"reversed range", "http://user:top-secret@proxy.example:{rand_int:9000-8000}"},
		{"zero range", "http://user:top-secret@proxy.example:{rand_int:0-10}"},
		{"large range", "http://user:top-secret@proxy.example:{rand_int:1-65536}"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateTemplate(test.input)
			if err == nil {
				t.Fatal("ValidateTemplate() error = nil")
			}
			if strings.Contains(err.Error(), test.input) || strings.Contains(err.Error(), "top-secret") {
				t.Fatalf("error leaked proxy URL or credentials: %q", err)
			}
		})
	}
}

func TestExpandRejectsBadChooserResult(t *testing.T) {
	template := "http://proxy.example:{rand_int:100-200}"

	_, err := expand(template, func(_, _ uint64) (uint64, error) {
		return 0, errors.New("entropy unavailable")
	})
	if err == nil || strings.Contains(err.Error(), template) {
		t.Fatalf("expand() error = %v", err)
	}

	_, err = expand(template, func(_, _ uint64) (uint64, error) {
		return 201, nil
	})
	if err == nil || strings.Contains(err.Error(), template) {
		t.Fatalf("expand() error = %v", err)
	}
}

func TestRedactRemovesAllUserinfo(t *testing.T) {
	const concrete = "https://alice:p%40ssword@proxy.example:8443"
	got := Redact(concrete)
	if got != "https://proxy.example:8443" {
		t.Fatalf("Redact() = %q", got)
	}
	if strings.Contains(got, "alice") || strings.Contains(got, "ssword") || strings.Contains(got, "@") {
		t.Fatalf("Redact() retained userinfo: %q", got)
	}

	invalid := Redact("http://alice:top-secret@proxy.example:not-a-port")
	if invalid != invalidRedaction {
		t.Fatalf("Redact(invalid) = %q", invalid)
	}
	if strings.Contains(invalid, "alice") || strings.Contains(invalid, "top-secret") {
		t.Fatalf("Redact(invalid) leaked userinfo: %q", invalid)
	}
}

func TestNewTransportClonesDefaultsAndKeepsProxyURLImmutable(t *testing.T) {
	const concrete = "https://alice:p%40ssword@proxy.example:8443"
	transport, err := NewTransport(concrete)
	if err != nil {
		t.Fatalf("NewTransport() error = %v", err)
	}

	defaults := http.DefaultTransport.(*http.Transport)
	if transport == defaults {
		t.Fatal("NewTransport() returned http.DefaultTransport")
	}
	if transport.MaxIdleConns != defaults.MaxIdleConns ||
		transport.MaxIdleConnsPerHost != defaults.MaxIdleConnsPerHost ||
		transport.IdleConnTimeout != defaults.IdleConnTimeout ||
		transport.TLSHandshakeTimeout != defaults.TLSHandshakeTimeout ||
		transport.ExpectContinueTimeout != defaults.ExpectContinueTimeout ||
		transport.ForceAttemptHTTP2 != defaults.ForceAttemptHTTP2 {
		t.Fatal("NewTransport() did not preserve default transport settings")
	}

	req, err := http.NewRequest(http.MethodGet, "https://example.com/", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	first, err := transport.Proxy(req)
	if err != nil {
		t.Fatalf("Proxy() error = %v", err)
	}
	first.Host = "mutated.example:1"
	first.User = nil

	second, err := transport.Proxy(req)
	if err != nil {
		t.Fatalf("Proxy() error = %v", err)
	}
	if first == second {
		t.Fatal("Proxy() reused a mutable URL pointer")
	}
	if second.String() != concrete {
		t.Fatalf("second Proxy() = %q", second.String())
	}
}

func TestNewTransportErrorsDoNotLeakCredentials(t *testing.T) {
	const raw = "http://alice:top-secret@proxy.example:not-a-port"
	_, err := NewTransport(raw)
	if err == nil {
		t.Fatal("NewTransport() error = nil")
	}
	if strings.Contains(err.Error(), raw) || strings.Contains(err.Error(), "alice") || strings.Contains(err.Error(), "top-secret") {
		t.Fatalf("error leaked proxy URL or credentials: %q", err)
	}
}

func TestProbeAcceptsAnyResponseAndLimitsBody(t *testing.T) {
	body := &trackingBody{}
	cleanedUp := false
	factory := func(concreteURL string) (*http.Client, func(), error) {
		if concreteURL != "http://proxy.example:8080" {
			t.Fatalf("factory URL = %q", concreteURL)
		}
		client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.String() != probeTargetURL {
				t.Fatalf("probe target = %q", req.URL)
			}
			deadline, ok := req.Context().Deadline()
			if !ok || time.Until(deadline) > probeTimeout {
				t.Fatalf("probe context deadline = %v, present = %v", deadline, ok)
			}
			return &http.Response{
				StatusCode: http.StatusProxyAuthRequired,
				Header:     make(http.Header),
				Body:       body,
				Request:    req,
			}, nil
		})}
		return client, func() { cleanedUp = true }, nil
	}

	err := probe(context.Background(), "http://proxy.example:8080", probeTargetURL, factory)
	if err != nil {
		t.Fatalf("probe() error = %v", err)
	}
	if body.bytesRead != probeBodyLimit {
		t.Fatalf("probe read %d body bytes, want %d", body.bytesRead, probeBodyLimit)
	}
	if !body.closed {
		t.Fatal("probe did not close response body")
	}
	if !cleanedUp {
		t.Fatal("probe did not clean up its client")
	}
}

func TestProbeClientDisablesRedirectsAndUsesShortTimeouts(t *testing.T) {
	client, cleanup, err := newProbeClient("http://proxy.example:8080")
	if err != nil {
		t.Fatalf("newProbeClient() error = %v", err)
	}
	defer cleanup()

	if client.Timeout != probeTimeout {
		t.Fatalf("client timeout = %v, want %v", client.Timeout, probeTimeout)
	}
	if err := client.CheckRedirect(nil, nil); !errors.Is(err, http.ErrUseLastResponse) {
		t.Fatalf("CheckRedirect() error = %v", err)
	}

	transport := client.Transport.(*http.Transport)
	if transport.ResponseHeaderTimeout <= 0 || transport.ResponseHeaderTimeout > probeTimeout {
		t.Fatalf("response header timeout = %v", transport.ResponseHeaderTimeout)
	}
	if transport.TLSHandshakeTimeout <= 0 || transport.TLSHandshakeTimeout > probeTimeout {
		t.Fatalf("TLS handshake timeout = %v", transport.TLSHandshakeTimeout)
	}
}

func TestProbeErrorsAreSanitized(t *testing.T) {
	const raw = "http://alice:top-secret@proxy.example:8080"
	factory := func(string) (*http.Client, func(), error) {
		client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("dial " + raw)
		})}
		return client, func() {}, nil
	}

	err := probe(context.Background(), raw, probeTargetURL, factory)
	if err == nil {
		t.Fatal("probe() error = nil")
	}
	if strings.Contains(err.Error(), raw) || strings.Contains(err.Error(), "alice") || strings.Contains(err.Error(), "top-secret") {
		t.Fatalf("error leaked proxy URL or credentials: %q", err)
	}

	err = Probe(context.Background(), "http://alice:top-secret@proxy.example:not-a-port")
	if err == nil {
		t.Fatal("Probe(invalid) error = nil")
	}
	if strings.Contains(err.Error(), "alice") || strings.Contains(err.Error(), "top-secret") {
		t.Fatalf("setup error leaked credentials: %q", err)
	}
}

func TestProbePreservesContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	factory := func(string) (*http.Client, func(), error) {
		client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			<-req.Context().Done()
			return nil, req.Context().Err()
		})}
		return client, func() {}, nil
	}

	err := probe(ctx, "http://proxy.example:8080", probeTargetURL, factory)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("probe() error = %v, want context.Canceled", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

type trackingBody struct {
	bytesRead int
	closed    bool
}

func (body *trackingBody) Read(buffer []byte) (int, error) {
	if body.closed {
		return 0, io.ErrClosedPipe
	}
	for index := range buffer {
		buffer[index] = 'x'
	}
	body.bytesRead += len(buffer)
	return len(buffer), nil
}

func (body *trackingBody) Close() error {
	body.closed = true
	return nil
}
