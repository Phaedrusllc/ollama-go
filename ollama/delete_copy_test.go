package ollama

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDelete_StatusCodes(t *testing.T) {
	tests := []int{200, 201, 202, 204, 500}
	for _, code := range tests {
        srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.WriteHeader(code)
            _, _ = io.WriteString(w, "{}")
        }))
		c := NewClient("")
		c.base = srv.URL
		st, _ := c.Delete(context.Background(), "m")
        switch {
        case code == 200:
            if *st.Status != "success" { t.Fatalf("code=%d want success got %v", code, *st.Status) }
        default:
            if *st.Status != "error" { t.Fatalf("code=%d want error got %v", code, *st.Status) }
        }
		srv.Close()
	}
}

func TestCopy_StatusCodes(t *testing.T) {
	tests := []int{200, 204, 500}
	for _, code := range tests {
        srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.WriteHeader(code)
            _, _ = io.WriteString(w, "{}")
        }))
		c := NewClient("")
		c.base = srv.URL
		st, _ := c.Copy(context.Background(), "a", "b")
		if code == 200 {
			if *st.Status != "success" {
				t.Fatalf("code=%d want success got %v", code, *st.Status)
			}
		} else {
			if *st.Status != "error" {
				t.Fatalf("code=%d want error got %v", code, *st.Status)
			}
		}
		srv.Close()
	}
}
