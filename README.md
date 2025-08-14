Ollama Go (Unofficial) — Python Parity Client

This package provides a Go client for Ollama that mirrors the public behavior of the official `ollama` Python client. It is designed for drop‑in parity on wire format and endpoint behavior with robust streaming and error handling.

Highlights
- Generate and Chat with optional streaming (NDJSON)
- Embed and (deprecated) Embeddings endpoints
- Model management: pull, push, create, delete, copy, show, list, ps
- Blob upload helper matching Python `create_blob`
- Host parsing semantics identical to Python `_parse_host`
- Error handling compatible with Python `ResponseError` (extracts JSON `error`)
- Friendly `ConnectionError` message aligned with Python client

Quick Start

  go get github.com/phaedrusllc/ollama-go/ollama

  package main

  import (
    "context"
    "fmt"
    "log"
    ollama "github.com/phaedrusllc/ollama-go/ollama"
  )

  func main() {
    prompt := "Say hello!"
    c := ollama.NewClient("") // honors OLLAMA_HOST
    resp, err := c.Generate(context.Background(), &ollama.GenerateRequest{BaseStreamableRequest: ollama.BaseStreamableRequest{Model: "llama3"}, Prompt: &prompt})
    if err != nil { log.Fatal(err) }
    fmt.Println(resp.Response)
  }

Streaming example

  stream, err := c.ChatStream(ctx, &ollama.ChatRequest{BaseStreamableRequest: ollama.BaseStreamableRequest{Model: "llama3"}, Messages: []ollama.Message{{Role: "user", Content: ollama.StrPtr("hi")}}})
  if err != nil { log.Fatal(err) }
  defer stream.Close()
  for {
    chunk, err := stream.Recv()
    if err == ollama.EOF { break }
    if err != nil { log.Fatal(err) }
    fmt.Print(chunk.Message.GetContent())
  }

Examples
- See the `examples/` folder for runnable programs (generate/chat/stream/embed/list/ps/blob). They read `.env` variables (`OLLAMA_BASE_URL`, `OLLAMA_MODEL`, and `OLLAMA_EMBED_MODEL`).

Testing
- Unit tests: `go test ./...` (robust suite including streaming edges, host parsing, errors, concurrency)
- End‑to‑end: `RUN_E2E=1 go test ./examples -v` (executes the examples against your Ollama instance)

How This Was Built
- We used the open-source Codex CLI to scaffold, iterate, and test.
- We studied the official `ollama` Python client to match behavior and wire format (HTTP routes, payload shapes, streaming semantics, error mapping, and host parsing).
- We referenced the official OpenAI Go client for idiomatic API design and streaming patterns.
- We created a rigorous test suite to ensure parity and robustness, including concurrency and edge‑case handling.

Notes
- This is an unofficial client implemented for parity with the Python client’s observable behavior. For production stability and long‑term support, verify against your Ollama version and models.

License
- MIT License © 2025 Phaedrus LLC
