package main

import (
	"context"
	"fmt"

	"github.com/phaedrusllc/ollama-go/examples/internal/envutil"
)

func main() {
	envutil.LoadDotEnv(".env")
	c := envutil.NewClientFromEnv()
	ctx := context.Background()
	tags, err := c.List(ctx)
	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}
	fmt.Println("TAGS count:", len(tags.Models))
	ps, err := c.PS(ctx)
	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}
	fmt.Println("PS count:", len(ps.Models))
}
