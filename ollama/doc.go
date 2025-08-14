// Package ollama provides a Go client for the Ollama HTTP API with runtime
// behavior aligned to the official Python client. It mirrors endpoint paths,
// request/response shapes, streaming semantics (NDJSON), error mapping, and
// host parsing so applications can expect consistent behavior across languages.
//
// Highlights:
//   - Generate and Chat with optional streaming
//   - Embed and (deprecated) Embeddings
//   - Model management: pull, push, create, delete, copy, show, list, ps
//   - Blob upload helper: CreateBlob
//   - Friendly connection errors and Python-parity ResponseError parsing
package ollama
