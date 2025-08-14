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
	model := envutil.ModelFromEnv("llama3.2")
	out, err := c.Chat(context.Background(), &ollama.ChatRequest{
		BaseStreamableRequest: ollama.BaseStreamableRequest{Model: model},
		Messages:              []ollama.Message{{Role: "user", Content: ollama.StrPtr("A two-word greeting.")}},
	})
	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}
	fmt.Println("CHAT:", out.Message.GetContent())
}
