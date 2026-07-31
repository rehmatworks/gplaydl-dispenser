// Package proxy validates, expands, and uses HTTP proxy URLs.
package proxy

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	probeTargetURL   = "https://android.clients.google.com/"
	probeTimeout     = 5 * time.Second
	probeBodyLimit   = 32 << 10
	invalidRedaction = "<invalid-proxy-url>"
)

var randomPortPattern = regexp.MustCompile(`\{rand_int:([0-9]+)-([0-9]+)\}`)

type templateSpec struct {
	raw      string
	start    int
	end      int
	min, max uint64
}

// ValidateTemplate validates a proxy URL template without expanding it.
//
// A template must be an http or https URL with an explicit port. It may have
// one {rand_int:min-max} shortcode in place of that port.
func ValidateTemplate(proxyTemplate string) error {
	_, err := parseTemplate(proxyTemplate)
	return err
}

// Expand validates proxyTemplate and replaces its optional random-port
// shortcode using crypto/rand. Both endpoints of the configured range are
// eligible.
func Expand(proxyTemplate string) (string, error) {
	return expand(proxyTemplate, cryptoRandomPort)
}

func expand(proxyTemplate string, choose func(uint64, uint64) (uint64, error)) (string, error) {
	spec, err := parseTemplate(proxyTemplate)
	if err != nil {
		return "", err
	}
	if spec.start < 0 {
		return spec.raw, nil
	}

	port, err := choose(spec.min, spec.max)
	if err != nil {
		return "", errors.New("could not generate random proxy port")
	}
	if port < spec.min || port > spec.max {
		return "", errors.New("random proxy port is outside the configured range")
	}

	return spec.raw[:spec.start] + strconv.FormatUint(port, 10) + spec.raw[spec.end:], nil
}

func cryptoRandomPort(min, max uint64) (uint64, error) {
	width := max - min + 1
	offset, err := rand.Int(rand.Reader, new(big.Int).SetUint64(width))
	if err != nil {
		return 0, err
	}
	return min + offset.Uint64(), nil
}

func parseTemplate(raw string) (templateSpec, error) {
	spec := templateSpec{raw: raw, start: -1}
	matches := randomPortPattern.FindAllStringSubmatchIndex(raw, -1)
	if len(matches) > 1 {
		return spec, errors.New("proxy template may contain at most one random-port shortcode")
	}
	if len(matches) == 0 {
		if strings.ContainsAny(raw, "{}") {
			return spec, errors.New("proxy template contains an invalid shortcode")
		}
		if _, err := parseConcreteURL(raw); err != nil {
			return spec, err
		}
		return spec, nil
	}

	match := matches[0]
	start, end := match[0], match[1]
	if strings.ContainsAny(raw[:start], "{}") || strings.ContainsAny(raw[end:], "{}") {
		return spec, errors.New("proxy template contains an invalid shortcode")
	}

	schemeEnd := strings.Index(raw, "://")
	if schemeEnd < 0 {
		return spec, errors.New("proxy URL must include a scheme")
	}
	authorityStart := schemeEnd + 3
	authorityEnd := len(raw)
	if offset := strings.IndexAny(raw[authorityStart:], "/?#"); offset >= 0 {
		authorityEnd = authorityStart + offset
	}
	if start <= authorityStart || end != authorityEnd || raw[start-1] != ':' {
		return spec, errors.New("random-port shortcode must replace the URL port")
	}

	min, err := strconv.ParseUint(raw[match[2]:match[3]], 10, 16)
	if err != nil || min == 0 {
		return spec, errors.New("random proxy port minimum must be between 1 and 65535")
	}
	max, err := strconv.ParseUint(raw[match[4]:match[5]], 10, 16)
	if err != nil || max == 0 {
		return spec, errors.New("random proxy port maximum must be between 1 and 65535")
	}
	if min > max {
		return spec, errors.New("random proxy port minimum must not exceed maximum")
	}

	candidate := raw[:start] + strconv.FormatUint(min, 10) + raw[end:]
	if _, err := parseConcreteURL(candidate); err != nil {
		return spec, err
	}

	spec.start, spec.end, spec.min, spec.max = start, end, min, max
	return spec, nil
}

// Redact removes all user information from a concrete proxy URL. Invalid
// inputs return a fixed placeholder so callers can safely use the result in
// logs.
func Redact(concreteURL string) string {
	parsed, err := parseConcreteURL(concreteURL)
	if err != nil {
		return invalidRedaction
	}
	redacted := *parsed
	redacted.User = nil
	return redacted.String()
}

// NewTransport returns a clone of http.DefaultTransport configured with a
// validated concrete proxy URL. The proxy URL captured by the transport
// cannot be changed through a URL returned by its Proxy callback.
func NewTransport(concreteURL string) (*http.Transport, error) {
	proxyURL, err := parseConcreteURL(concreteURL)
	if err != nil {
		return nil, err
	}

	proxyValue := *proxyURL
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = func(*http.Request) (*url.URL, error) {
		copy := proxyValue
		return &copy, nil
	}
	return transport, nil
}

// Probe verifies that concreteURL can reach the Google Play endpoint. Any HTTP
// response proves connectivity, regardless of its status code.
func Probe(ctx context.Context, concreteURL string) error {
	return probe(ctx, concreteURL, probeTargetURL, newProbeClient)
}

type probeClientFactory func(string) (*http.Client, func(), error)

func newProbeClient(concreteURL string) (*http.Client, func(), error) {
	transport, err := NewTransport(concreteURL)
	if err != nil {
		return nil, nil, err
	}

	dialer := &net.Dialer{
		Timeout:   3 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	transport.DialContext = dialer.DialContext
	transport.TLSHandshakeTimeout = 3 * time.Second
	transport.ResponseHeaderTimeout = 4 * time.Second

	client := &http.Client{
		Transport: transport,
		Timeout:   probeTimeout,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	return client, transport.CloseIdleConnections, nil
}

func probe(ctx context.Context, concreteURL, targetURL string, newClient probeClientFactory) error {
	client, closeClient, err := newClient(concreteURL)
	if err != nil {
		return errors.New("proxy probe setup failed")
	}
	defer closeClient()

	probeCtx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(probeCtx, http.MethodGet, targetURL, nil)
	if err != nil {
		return errors.New("could not create proxy probe request")
	}

	resp, requestErr := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, probeBodyLimit))
		return nil
	}
	if requestErr != nil {
		if contextErr := probeCtx.Err(); contextErr != nil {
			return fmt.Errorf("proxy probe failed: %w", contextErr)
		}
		return errors.New("proxy probe failed")
	}
	return errors.New("proxy probe returned no response")
}

func parseConcreteURL(raw string) (*url.URL, error) {
	if raw == "" {
		return nil, errors.New("proxy URL is required")
	}
	if strings.ContainsAny(raw, "{}") {
		return nil, errors.New("concrete proxy URL must not contain a shortcode")
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return nil, errors.New("proxy URL is invalid")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, errors.New("proxy URL scheme must be http or https")
	}
	if parsed.Opaque != "" || parsed.Host == "" {
		return nil, errors.New("proxy URL must include a hostname")
	}
	if parsed.Path != "" || parsed.RawPath != "" {
		return nil, errors.New("proxy URL must not include a path")
	}
	if parsed.RawQuery != "" || parsed.ForceQuery {
		return nil, errors.New("proxy URL must not include a query")
	}
	if parsed.Fragment != "" || parsed.RawFragment != "" || strings.Contains(raw, "#") {
		return nil, errors.New("proxy URL must not include a fragment")
	}

	hostname := parsed.Hostname()
	if !validHostname(hostname) {
		return nil, errors.New("proxy URL hostname is invalid")
	}
	if err := validateCredentials(parsed.User); err != nil {
		return nil, err
	}

	port := parsed.Port()
	if port == "" {
		return nil, errors.New("proxy URL must include a port")
	}
	portNumber, err := strconv.ParseUint(port, 10, 16)
	if err != nil || portNumber == 0 {
		return nil, errors.New("proxy URL port must be between 1 and 65535")
	}

	return parsed, nil
}

func validateCredentials(user *url.Userinfo) error {
	if user == nil {
		return nil
	}
	if user.Username() == "" {
		return errors.New("proxy URL credentials require a username")
	}
	if containsControl(user.Username()) {
		return errors.New("proxy URL username contains invalid characters")
	}
	if password, present := user.Password(); present {
		if password == "" {
			return errors.New("proxy URL password must not be empty")
		}
		if containsControl(password) {
			return errors.New("proxy URL password contains invalid characters")
		}
	}
	return nil
}

func containsControl(value string) bool {
	for _, char := range value {
		if char < 0x20 || char == 0x7f {
			return true
		}
	}
	return false
}

func validHostname(hostname string) bool {
	if hostname == "" {
		return false
	}
	if net.ParseIP(hostname) != nil {
		return true
	}

	hostname = strings.TrimSuffix(hostname, ".")
	if hostname == "" || len(hostname) > 253 {
		return false
	}
	for _, label := range strings.Split(hostname, ".") {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, char := range label {
			if (char < 'a' || char > 'z') &&
				(char < 'A' || char > 'Z') &&
				(char < '0' || char > '9') &&
				char != '-' {
				return false
			}
		}
	}
	return true
}
