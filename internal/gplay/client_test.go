package gplay

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestExchangeAASTokenClassifiesCredentialErrors(t *testing.T) {
	client := &Client{}
	account := Account{Email: "person@example.com", AASToken: "aas_et/secret"}
	config := DeviceConfig{"GSF.version": "1", "Build.VERSION.SDK_INT": "35"}

	tests := []struct {
		name       string
		statusCode int
		body       string
		permanent  bool
	}{
		{"non-200 permanent rejection", http.StatusForbidden, "Error=BadAuthentication\n", true},
		{"200 permanent rejection", http.StatusOK, "Error=AccountDisabled\n", true},
		{"transient auth error", http.StatusOK, "Error=ServiceUnavailable\n", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			httpClient := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: test.statusCode,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader(test.body)),
					Request:    req,
				}, nil
			})}
			_, err := client.exchangeAASToken(
				context.Background(), httpClient, account, config, "abc123", "en",
			)
			if err == nil {
				t.Fatal("exchangeAASToken() error = nil")
			}
			if IsCredentialError(err) != test.permanent {
				t.Fatalf("IsCredentialError(%v) = %v, want %v", err, IsCredentialError(err), test.permanent)
			}
		})
	}
}

func TestIsProxyConnectionError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"network failure", &connectionError{}, true},
		{"attempt timeout", context.DeadlineExceeded, true},
		{"request canceled", context.Canceled, false},
		{"proxy authentication", &httpStatusError{statusCode: http.StatusProxyAuthRequired}, true},
		{"proxy gateway", &httpStatusError{statusCode: http.StatusBadGateway}, true},
		{"target server error", &httpStatusError{statusCode: http.StatusInternalServerError}, false},
		{"credential rejection", &CredentialError{Code: "BadAuthentication"}, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := IsProxyConnectionError(test.err); got != test.want {
				t.Fatalf("IsProxyConnectionError(%v) = %v, want %v", test.err, got, test.want)
			}
		})
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}
