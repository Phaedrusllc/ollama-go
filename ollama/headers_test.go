package ollama

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHeaderOverride(t *testing.T) {
	var ua, accept string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua = r.Header.Get("User-Agent")
		accept = r.Header.Get("Accept")
		w.WriteHeader(200)
	}))
	defer srv.Close()
	c := NewClient("", WithHeader("Accept", "application/problem+json"))
	c.base = srv.URL
	_, _ = c.List(context.Background())
	if ua == "" {
		t.Fatal("missing User-Agent")
	}
	if accept != "application/problem+json" {
		t.Fatalf("accept override failed: %q", accept)
	}
}
