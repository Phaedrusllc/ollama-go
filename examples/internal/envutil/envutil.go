package envutil

import (
	"net/http"
	"os"
	"strings"
	"time"

	ollama "github.com/phaedrusllc/ollama-go/ollama"
)

// LoadDotEnv loads a simple KEY=VALUE .env file into process env.
func LoadDotEnv(path string) {
	b, err := os.ReadFile(path)
	if err != nil {
		return
	}
	lines := strings.Split(string(b), "\n")
	for _, ln := range lines {
		s := strings.TrimSpace(ln)
		if s == "" || strings.HasPrefix(s, "#") {
			continue
		}
		kv := strings.SplitN(s, "=", 2)
		if len(kv) != 2 {
			continue
		}
		k := strings.TrimSpace(kv[0])
		v := strings.TrimSpace(kv[1])
		os.Setenv(k, v)
	}
}

// NewClientFromEnv creates a client using OLLAMA_HOST or OLLAMA_BASE_URL.
func NewClientFromEnv() *ollama.Client {
	if os.Getenv("OLLAMA_HOST") == "" && os.Getenv("OLLAMA_BASE_URL") != "" {
		os.Setenv("OLLAMA_HOST", os.Getenv("OLLAMA_BASE_URL"))
	}
	hc := &http.Client{Timeout: 60 * time.Second}
	return ollama.NewClient("", ollama.WithHTTPClient(hc))
}

// ModelFromEnv returns OLLAMA_MODEL or provided fallback.
func ModelFromEnv(fallback string) string {
	m := os.Getenv("OLLAMA_MODEL")
	if m == "" {
		return fallback
	}
	return m
}

// EmbedModelFromEnv prioritizes OLLAMA_EMBED_MODEL, then OLLAMA_MODEL.
func EmbedModelFromEnv(fallback string) string {
	if m := os.Getenv("OLLAMA_EMBED_MODEL"); m != "" {
		return m
	}
	if m := os.Getenv("OLLAMA_MODEL"); m != "" {
		return m
	}
	return fallback
}
