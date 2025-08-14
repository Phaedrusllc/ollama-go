package ollama

import (
	"context"
	"net/http"
	"net/url"
	"syscall"
	"testing"
)

type failingRoundTripper struct{}

func (f failingRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, &url.Error{Op: "Post", URL: "http://127.0.0.1:11434/api/generate", Err: syscall.ECONNREFUSED}
}

func TestConnectionError_FriendlyMessage(t *testing.T) {
    c := NewClient("")
    c.hc = &http.Client{Transport: failingRoundTripper{}}
	_, err := c.Generate(context.Background(), &GenerateRequest{BaseStreamableRequest: BaseStreamableRequest{Model: "m"}})
	if err == nil {
		t.Fatal("expected error")
	}
	if _, ok := err.(*ConnectionError); !ok {
		t.Fatalf("expected ConnectionError, got %T", err)
	}
	if err.Error() != connectionErrorMessage {
		t.Fatalf("unexpected message: %q", err.Error())
	}
}
