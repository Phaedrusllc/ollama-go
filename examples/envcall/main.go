package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	ollama "github.com/phaedrusllc/ollama-go/ollama"
)

func loadDotEnv(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
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
	return nil
}

func main() {
	// Load .env if present
	_ = loadDotEnv(".env")

	// Map OLLAMA_BASE_URL -> OLLAMA_HOST for our client
	if os.Getenv("OLLAMA_HOST") == "" && os.Getenv("OLLAMA_BASE_URL") != "" {
		os.Setenv("OLLAMA_HOST", os.Getenv("OLLAMA_BASE_URL"))
	}
	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		model = "llama3.2" // reasonable default if not provided
	}

	// 60s timeout client
	hc := &http.Client{Timeout: 60 * time.Second}
	c := ollama.NewClient("", ollama.WithHTTPClient(hc))

	ctx := context.Background()

	// 1) Generate (non-stream)
	prompt := "Say hello in one short sentence."
	gen, err := c.Generate(ctx, &ollama.GenerateRequest{
		BaseStreamableRequest: ollama.BaseStreamableRequest{Model: model},
		Prompt:                &prompt,
	})
	if err != nil {
		log.Fatalf("generate error: %v", err)
	}
	fmt.Println("GENERATE:", gen.Response)

	// 2) Chat (stream)
	s, err := c.ChatStream(ctx, &ollama.ChatRequest{
		BaseStreamableRequest: ollama.BaseStreamableRequest{Model: model},
		Messages:              []ollama.Message{{Role: "user", Content: ollama.StrPtr("Give me a haiku about Go.")}},
	})
	if err != nil {
		log.Fatalf("chat stream error: %v", err)
	}
	defer s.Close()
	var acc strings.Builder
	for {
		part, err := s.Recv()
		if err == ollama.EOF {
			break
		}
		if err != nil {
			log.Fatalf("stream recv error: %v", err)
		}
		acc.WriteString(part.Message.GetContent())
	}
	fmt.Println("CHAT:", acc.String())

	// 3) Embed
	emb, err := c.Embed(ctx, &ollama.EmbedRequest{
		Model: model,
		Input: "test embedding",
	})
	if err != nil {
		log.Printf("embed error (may not be supported by model): %v", err)
	} else {
		// print dimensions
		if len(emb.Embeddings) > 0 {
			fmt.Println("EMBED dims:", len(emb.Embeddings[0]))
		} else {
			fmt.Println("EMBED: empty")
		}
	}
}
