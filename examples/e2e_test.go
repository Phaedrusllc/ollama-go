package examples

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func runExample(t *testing.T, rel string) string {
	t.Helper()
	cmd := exec.Command("go", "run", "./"+rel)
	cmd.Env = append(os.Environ(), "GO111MODULE=on")
	b, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s failed: %v\n%s", rel, err, string(b))
	}
	return string(b)
}

func TestE2E_Examples(t *testing.T) {
	if os.Getenv("RUN_E2E") == "" {
		t.Skip("set RUN_E2E=1 to run end-to-end examples against your Ollama instance")
	}
	// Basic env checks
	if os.Getenv("OLLAMA_HOST") == "" && os.Getenv("OLLAMA_BASE_URL") == "" {
		t.Skip("set OLLAMA_HOST or OLLAMA_BASE_URL for e2e tests")
	}
	if os.Getenv("OLLAMA_MODEL") == "" {
		t.Skip("set OLLAMA_MODEL for e2e tests")
	}

	root := findRepoRoot(t)
	chdir(t, root)

	out := runExample(t, "examples/generate")
	if !strings.Contains(out, "GENERATE:") {
		t.Fatalf("unexpected output: %s", out)
	}

	out = runExample(t, "examples/generate_stream")
	if !strings.Contains(out, "GENERATE_STREAM:") {
		t.Fatalf("unexpected output: %s", out)
	}

	out = runExample(t, "examples/chat")
	if !strings.Contains(out, "CHAT:") {
		t.Fatalf("unexpected output: %s", out)
	}

	out = runExample(t, "examples/chat_stream")
	if !strings.Contains(out, "CHAT_STREAM:") {
		t.Fatalf("unexpected output: %s", out)
	}

	out = runExample(t, "examples/embed")
	if !(strings.Contains(out, "EMBED dims:") || strings.Contains(out, "EMBED ERROR:")) {
		t.Fatalf("unexpected embed output: %s", out)
	}

	out = runExample(t, "examples/list_show_ps")
	if !(strings.Contains(out, "TAGS count:") && strings.Contains(out, "PS count:")) {
		t.Fatalf("unexpected list/ps output: %s", out)
	}

	out = runExample(t, "examples/blob")
	if !strings.Contains(out, "BLOB:") {
		t.Fatalf("unexpected blob output: %s", out)
	}
}

func chdir(t *testing.T, dir string) {
	t.Helper()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, _ := os.Getwd()
	for i := 0; i < 6; i++ {
		if fi, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil && !fi.IsDir() {
			return dir
		}
		dir = filepath.Dir(dir)
	}
	t.Fatal("go.mod not found")
	return ""
}
