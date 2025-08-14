package main

import (
	"context"
	"fmt"
	"os"

	"github.com/phaedrusllc/ollama-go/examples/internal/envutil"
)

func main() {
	envutil.LoadDotEnv(".env")
	c := envutil.NewClientFromEnv()
	f, _ := os.CreateTemp("", "blob-*.txt")
	defer os.Remove(f.Name())
	f.WriteString("hello blob")
	f.Close()
	digest, err := c.CreateBlob(context.Background(), f.Name())
	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}
	fmt.Println("BLOB:", digest)
}
