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
	s, err := c.GenerateStream(context.Background(), &ollama.GenerateRequest{
		BaseStreamableRequest: ollama.BaseStreamableRequest{Model: model},
		Prompt:                ollama.StrPtr("Count from one to three, words only."),
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
		acc.WriteString(part.Response)
	}
	fmt.Println("GENERATE_STREAM:", acc.String())
}
