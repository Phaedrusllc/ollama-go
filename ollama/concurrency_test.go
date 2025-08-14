package ollama

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestConcurrentRequests_DoNotRaceOrPanic(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            switch r.URL.Path {
            case "/api/generate":
                _, _ = io.Copy(io.Discard, r.Body)
                w.Header().Set("Content-Type", "application/json")
                _, _ = io.WriteString(w, `{"response":"ok"}`)
            case "/api/chat":
                _, _ = io.Copy(io.Discard, r.Body)
                if err := json.NewEncoder(w).Encode(map[string]any{"message": map[string]any{"role": "assistant", "content": "ok"}}); err != nil { t.Fatalf("encode: %v", err) }
            default:
                w.WriteHeader(404)
            }
	}))
	defer srv.Close()

	c := NewClient("")
	c.base = srv.URL
	ctx := context.Background()

	const N = 50
	wg := sync.WaitGroup{}
	wg.Add(N * 2)
	for i := 0; i < N; i++ {
		go func() {
			defer wg.Done()
			_, err := c.Generate(ctx, &GenerateRequest{BaseStreamableRequest: BaseStreamableRequest{Model: "m"}})
			if err != nil {
				t.Errorf("gen err: %v", err)
			}
		}()
		go func() {
			defer wg.Done()
			_, err := c.Chat(ctx, &ChatRequest{BaseStreamableRequest: BaseStreamableRequest{Model: "m"}})
			if err != nil {
				t.Errorf("chat err: %v", err)
			}
		}()
	}
	wg.Wait()
}

func TestMissingModel_ReturnsError(t *testing.T) {
	c := NewClient("")
	if _, err := c.Generate(context.Background(), &GenerateRequest{}); err == nil {
		t.Fatal("expected error")
	}
}
