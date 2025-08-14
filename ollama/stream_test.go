package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStream_FinalLineWithoutNewline(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-ndjson")
        _, _ = io.WriteString(w, `{"response":"hello"}`) // no trailing newline
	}))
	defer srv.Close()
	c := NewClient("")
	c.base = srv.URL
	s, err := c.GenerateStream(context.Background(), &GenerateRequest{BaseStreamableRequest: BaseStreamableRequest{Model: "m"}})
	if err != nil {
		t.Fatal(err)
	}
    defer func(){ _ = s.Close() }()
	part, err := s.Recv()
	if err != nil {
		t.Fatal(err)
	}
	if part.Response != "hello" {
		t.Fatalf("got: %q", part.Response)
	}
	if _, err := s.Recv(); err == nil {
		t.Fatal("expected EOF")
	}
}

func TestStream_HTTPErrorBeforeStreaming(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
        _, _ = io.WriteString(w, `{"error":"bad"}`)
	}))
	defer srv.Close()
	c := NewClient("")
	c.base = srv.URL
	_, err := c.GenerateStream(context.Background(), &GenerateRequest{BaseStreamableRequest: BaseStreamableRequest{Model: "m"}})
	if err == nil {
		t.Fatal("expected error")
	}
	if _, ok := err.(*ResponseError); !ok {
		t.Fatalf("want ResponseError, got %T", err)
	}
}

func TestStream_RequestBodyIncludesStreamTrue(t *testing.T) {
	var body bytes.Buffer
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        _, _ = io.Copy(&body, r.Body)
		w.Header().Set("Content-Type", "application/x-ndjson")
        _, _ = io.WriteString(w, `{"response":"x"}\n`)
	}))
	defer srv.Close()
	c := NewClient("")
	c.base = srv.URL
	s, err := c.GenerateStream(context.Background(), &GenerateRequest{BaseStreamableRequest: BaseStreamableRequest{Model: "m"}})
	if err != nil {
		t.Fatal(err)
	}
    defer func(){ _ = s.Close() }()
	_, _ = s.Recv()
	var got map[string]any
	if err := json.Unmarshal(body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if v, ok := got["stream"].(bool); !ok || !v {
		t.Fatalf("expected stream=true, got: %v", got["stream"])
	}
}
