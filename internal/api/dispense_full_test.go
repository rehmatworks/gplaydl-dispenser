package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// full=1 decides whether GET /api/auth hands back every token or just
// {email, auth}, so only an explicit truthy value may switch it on.
func TestQueryBool(t *testing.T) {
	cases := map[string]bool{
		"":            false,
		"?full=1":     true,
		"?full=true":  true,
		"?full=0":     false,
		"?full=false": false,
		"?full=":      false,
		"?full":       false,
		"?full=yes":   false,
		"?other=1":    false,
	}
	for query, want := range cases {
		r := httptest.NewRequest(http.MethodGet, "/api/auth"+query, nil)
		if got := queryBool(r, "full"); got != want {
			t.Errorf("queryBool(%q) = %v, want %v", query, got, want)
		}
	}
}
