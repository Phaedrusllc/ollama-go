package ollama

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestServer(t *testing.T, handler func(w http.ResponseWriter, r *http.Request)) (*httptest.Server, *Client) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(handler))
	c := NewClient("")
	c.base = srv.URL
	return srv, c
}

func TestGenerate_JSON(t *testing.T) {
	ctx := context.Background()
	srv, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/generate" || r.Method != http.MethodPost {
			t.Fatalf("wrong path/method")
		}
		var got GenerateRequest
		_ = json.NewDecoder(r.Body).Decode(&got)
		if got.BaseStreamableRequest.Model != "m" {
			t.Fatalf("bad model")
		}
        if _, err := io.WriteString(w, `{"response":"hi","model":"m"}`); err != nil { t.Fatal(err) }
	})
	defer srv.Close()

	out, err := c.Generate(ctx, &GenerateRequest{BaseStreamableRequest: BaseStreamableRequest{Model: "m"}})
	if err != nil {
		t.Fatal(err)
	}
	if out.Response != "hi" {
		t.Fatalf("unexpected: %+v", out)
	}
}

func TestGenerate_Stream(t *testing.T) {
	ctx := context.Background()
	chunks := []string{
		"{\"response\":\"h\",\"model\":\"m\"}\n",
		"{\"response\":\"i\",\"model\":\"m\"}\n",
		"{\"done\":true,\"done_reason\":\"stop\"}\n",
	}
	srv, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/generate" || r.Method != http.MethodPost {
			t.Fatalf("wrong path/method")
		}
		w.Header().Set("Content-Type", "application/x-ndjson")
        for _, ch := range chunks {
                if _, err := io.WriteString(w, ch); err != nil { t.Fatal(err) }
        }
	})
	defer srv.Close()

	s, err := c.GenerateStream(ctx, &GenerateRequest{BaseStreamableRequest: BaseStreamableRequest{Model: "m"}})
	if err != nil {
		t.Fatal(err)
	}
    defer func(){ _ = s.Close() }()
	var acc strings.Builder
	for {
		part, err := s.Recv()
		if err == EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		acc.WriteString(part.Response)
	}
	if acc.String() != "hi" {
		t.Fatalf("acc=%q", acc.String())
	}
}

func TestChat_JSON(t *testing.T) {
	ctx := context.Background()
	srv, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" || r.Method != http.MethodPost {
			t.Fatalf("wrong path/method")
		}
        if _, err := io.WriteString(w, `{"message":{"role":"assistant","content":"pong"}}`); err != nil { t.Fatal(err) }
	})
	defer srv.Close()

	out, err := c.Chat(ctx, &ChatRequest{BaseStreamableRequest: BaseStreamableRequest{Model: "m"}, Messages: []Message{{Role: "user", Content: StrPtr("ping")}}})
	if err != nil {
		t.Fatal(err)
	}
	if out.Message.GetContent() != "pong" {
		t.Fatalf("unexpected: %+v", out)
	}
}

func TestErrors_HTTPStatus(t *testing.T) {
	ctx := context.Background()
	srv, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
        if _, err := io.WriteString(w, `{"error":"bad model"}`); err != nil { t.Fatal(err) }
	})
	defer srv.Close()

	_, err := c.Generate(ctx, &GenerateRequest{BaseStreamableRequest: BaseStreamableRequest{Model: "m"}})
	if err == nil {
		t.Fatal("want error")
	}
	re, ok := err.(*ResponseError)
	if !ok {
		t.Fatalf("wrong err type: %T", err)
	}
	if !strings.Contains(re.Error(), "bad model") {
		t.Fatalf("unexpected err: %v", re)
	}
}

func TestProgress_Stream_ErrorChunk(t *testing.T) {
	ctx := context.Background()
	srv, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-ndjson")
        if _, err := io.WriteString(w, "{\"status\":\"pulling\"}\n"); err != nil { t.Fatal(err) }
        if _, err := io.WriteString(w, "{\"error\":\"bad pull\"}\n"); err != nil { t.Fatal(err) }
	})
	defer srv.Close()
	s, err := c.PullStream(ctx, &PullRequest{Model: "m"})
	if err != nil {
		t.Fatal(err)
	}
    defer func(){ _ = s.Close() }()
	if _, err := s.Recv(); err != nil {
		t.Fatal(err)
	} // first ok
	if _, err := s.Recv(); err == nil {
		t.Fatal("want error from error chunk")
	}
}

func TestDelete_StatusMapping(t *testing.T) {
	ctx := context.Background()
	okSrv, okc := newTestServer(t, func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	defer okSrv.Close()
	st, err := okc.Delete(ctx, "m")
	if err != nil {
		t.Fatal(err)
	}
	if st.Status == nil || *st.Status != "success" {
		t.Fatalf("unexpected: %+v", st)
	}

    errSrv, errc := newTestServer(t, func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(500); _, _ = io.WriteString(w, "{}") })
	defer errSrv.Close()
	st2, err := errc.Delete(ctx, "m")
	if err != nil {
		t.Fatal(err)
	}
	if st2.Status == nil || *st2.Status != "error" {
		t.Fatalf("unexpected: %+v", st2)
	}
}

func TestCopy_StatusMapping(t *testing.T) {
	ctx := context.Background()
	okSrv, okc := newTestServer(t, func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	defer okSrv.Close()
	st, err := okc.Copy(ctx, "a", "b")
	if err != nil {
		t.Fatal(err)
	}
	if st.Status == nil || *st.Status != "success" {
		t.Fatalf("unexpected: %+v", st)
	}

    errSrv, errc := newTestServer(t, func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(500); _, _ = io.WriteString(w, "{}") })
	defer errSrv.Close()
	st2, err := errc.Copy(ctx, "a", "b")
	if err != nil {
		t.Fatal(err)
	}
	if st2.Status == nil || *st2.Status != "error" {
		t.Fatalf("unexpected: %+v", st2)
	}
}

func TestCreateBlob(t *testing.T) {
	ctx := context.Background()
	data := []byte("abc")
	path, err := tmpFileWith(".bin", data)
	if err != nil {
		t.Fatal(err)
	}
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("wrong method")
		}
		if !strings.HasPrefix(r.URL.Path, "/api/blobs/sha256:") {
			t.Fatalf("wrong path: %s", r.URL.Path)
		}
        _, _ = io.Copy(io.Discard, r.Body)
	}
	srv, c := newTestServer(t, handler)
	defer srv.Close()
	digest, err := c.CreateBlob(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(digest, "sha256:") {
		t.Fatalf("bad digest: %s", digest)
	}
}
