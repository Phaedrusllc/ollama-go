package main

import (
	"context"
	"fmt"

	"github.com/phaedrusllc/ollama-go/examples/internal/envutil"
	ollama "github.com/phaedrusllc/ollama-go/ollama"
)

func main() {
	envutil.LoadDotEnv(".env")
	c := envutil.NewClientFromEnv()
	model := envutil.EmbedModelFromEnv("nomic-embed-text")
	emb, err := c.Embed(context.Background(), &ollama.EmbedRequest{Model: model, Input: "embed me"})
	if err != nil {
		fmt.Println("EMBED ERROR:", err)
		return
	}
	if len(emb.Embeddings) == 0 {
		fmt.Println("EMBED: empty")
		return
	}
	fmt.Println("EMBED dims:", len(emb.Embeddings[0]))
}
