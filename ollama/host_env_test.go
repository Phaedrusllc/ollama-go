package ollama

import (
	"os"
	"strings"
	"testing"
)

func TestNewClient_UsesEnvHost(t *testing.T) {
    old := os.Getenv("OLLAMA_HOST")
    defer func(){ _ = os.Setenv("OLLAMA_HOST", old) }()
    _ = os.Setenv("OLLAMA_HOST", "example.com:5555/api")
	c := NewClient("")
	if !strings.HasPrefix(c.base, "http://example.com:5555/api") {
		t.Fatalf("unexpected base: %s", c.base)
	}
}
