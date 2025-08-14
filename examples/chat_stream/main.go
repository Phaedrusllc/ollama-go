package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/phaedrusllc/ollama-go/examples/internal/envutil"
	ollama "github.com/phaedrusllc/ollama-go/ollama"
)

func main() {
	envutil.LoadDotEnv(".env")
	c := envutil.NewClientFromEnv()
	model := envutil.ModelFromEnv("llama3.2")
	s, err := c.ChatStream(context.Background(), &ollama.ChatRequest{
		BaseStreamableRequest: ollama.BaseStreamableRequest{Model: model},
		Messages:              []ollama.Message{{Role: "user", Content: ollama.StrPtr("Write a 5-7-5 haiku.")}},
	})
	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}
	defer s.Close()
	var acc strings.Builder
	for {
		part, err := s.Recv()
		if err == ollama.EOF {
			break
		}
		if err != nil {
			fmt.Println("ERROR:", err)
			return
		}
		acc.WriteString(part.Message.GetContent())
	}
	fmt.Println("CHAT_STREAM:", acc.String())
}
