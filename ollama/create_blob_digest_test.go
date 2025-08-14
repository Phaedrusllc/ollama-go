package ollama

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestCreateBlob_DigestAndUpload(t *testing.T) {
	data := []byte("content to hash")
	tmp, err := os.CreateTemp("", "blob-*.bin")
	if err != nil {
		t.Fatal(err)
	}
    defer func(){ _ = os.Remove(tmp.Name()) }()
	if _, err := tmp.Write(data); err != nil {
		t.Fatal(err)
	}
    if err := tmp.Close(); err != nil { t.Fatal(err) }

	var uploaded []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		uploaded = b
		w.WriteHeader(200)
	}))
	defer srv.Close()

	c := NewClient("")
	c.base = srv.URL
	digest, err := c.CreateBlob(context.Background(), tmp.Name())
	if err != nil {
		t.Fatal(err)
	}
	want := fmt.Sprintf("sha256:%x", sha256.Sum256(data))
	if digest != want {
		t.Fatalf("got %q want %q", digest, want)
	}
	if string(uploaded) != string(data) {
		t.Fatalf("uploaded mismatch")
	}
}
