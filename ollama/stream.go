package ollama

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

// EOF mirrors io.EOF for convenient stream loop checks.
var EOF = io.EOF

// Stream is a generic JSON lines stream decoder.
// Stream decodes newline-delimited JSON responses for streaming endpoints.
// Recv returns the next decoded value or EOF when the stream ends. Close must
// be called to release the underlying response body.
type Stream[T any] struct {
	ctx    context.Context
	resp   *http.Response
	rd     *bufio.Reader
	closer io.Closer
	decode func([]byte, *T) error
}

func newStream[T any](ctx context.Context, resp *http.Response) *Stream[T] {
	return &Stream[T]{
		ctx:    ctx,
		resp:   resp,
		rd:     bufio.NewReader(resp.Body),
		closer: resp.Body,
		decode: func(b []byte, v *T) error { return json.Unmarshal(b, v) },
	}
}

// Recv reads next JSON line chunk.
// Recv reads and decodes the next JSON line from the stream.
func (s *Stream[T]) Recv() (*T, error) {
	if s.rd == nil {
		return nil, io.EOF
	}
	line, err := s.rd.ReadBytes('\n')
	if len(line) == 0 && err != nil {
		if err == io.EOF {
			return nil, io.EOF
		}
		return nil, err
	}
	var out T
	if err := s.decode(bytesTrimSpace(line), &out); err != nil {
		return nil, err
	}

	// If the JSON has an "error" field, mirror Python behavior and return ResponseError
	var probe map[string]any
	if err := json.Unmarshal(bytesTrimSpace(line), &probe); err == nil {
		if e, ok := probe["error"].(string); ok && e != "" {
			return nil, &ResponseError{Message: e, StatusCode: s.resp.StatusCode}
		}
	}
	return &out, nil
}

// Close releases the underlying response body.
func (s *Stream[T]) Close() error {
	if s.closer == nil {
		return nil
	}
	return s.closer.Close()
}

func bytesTrimSpace(b []byte) []byte {
	for len(b) > 0 && (b[0] == ' ' || b[0] == '\t' || b[0] == '\n' || b[0] == '\r') {
		b = b[1:]
	}
	for len(b) > 0 && (b[len(b)-1] == ' ' || b[len(b)-1] == '\t' || b[len(b)-1] == '\n' || b[len(b)-1] == '\r') {
		b = b[:len(b)-1]
	}
	return b
}
