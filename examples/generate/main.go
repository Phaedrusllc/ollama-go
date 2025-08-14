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
	prompt := "Say a single cheerful word."
	resp, err := c.Generate(context.Background(), &ollama.GenerateRequest{
		BaseStreamableRequest: ollama.BaseStreamableRequest{Model: model},
		Prompt:                &prompt,
	})
	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}
	fmt.Println("GENERATE:", resp.Response)
}
