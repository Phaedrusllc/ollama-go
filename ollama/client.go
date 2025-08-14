package ollama

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"syscall"
)

const defaultBase = "http://127.0.0.1:11434"

// Client is a minimal HTTP client for the Ollama API.
// It is safe for concurrent use and honors OLLAMA_HOST when host is empty.
type Client struct {
	hc     *http.Client
	base   string
	header http.Header
}

// NewClient constructs a Client. If host is empty, it uses the OLLAMA_HOST
// environment variable; if that too is empty, it defaults to 127.0.0.1:11434.
func NewClient(host string, opts ...ClientOption) *Client {
	base := host
	if base == "" {
		base = os.Getenv("OLLAMA_HOST")
	}
	if base != "" {
		base = parseHost(base)
	} else {
		base = defaultBase
	}
	c := &Client{
		hc:   &http.Client{},
		base: base,
		header: http.Header{
			"Content-Type": []string{"application/json"},
			"Accept":       []string{"application/json"},
			"User-Agent":   []string{fmt.Sprintf("ollama-go/0.0.0 (%s %s) Go/%s", runtime.GOARCH, runtime.GOOS, runtime.Version())},
		},
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// ClientOption customizes a Client at construction.
type ClientOption func(*Client)

// WithHTTPClient sets a custom http.Client.
func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *Client) {
		if hc != nil {
			c.hc = hc
		}
	}
}

// WithHeader sets a default header value for all requests.
func WithHeader(k, v string) ClientOption { return func(c *Client) { c.header.Set(k, v) } }

func (c *Client) do(ctx context.Context, method, path string, body io.Reader, headers http.Header) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.base+path, body)
	if err != nil {
		return nil, err
	}
	// default headers
	for k, vv := range c.header {
		for _, v := range vv {
			req.Header.Add(k, v)
		}
	}
	for k, vv := range headers {
		for _, v := range vv {
			req.Header.Add(k, v)
		}
	}
	resp, err := c.hc.Do(req)
	if err != nil {
		if isConnectErr(err) {
			return nil, &ConnectionError{Message: connectionErrorMessage}
		}
		return nil, err
	}
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return nil, newResponseError(resp.StatusCode, b)
	}
	return resp, nil
}

func isConnectErr(err error) bool {
	// Common syscall errors indicating connectivity issues
	if errors.Is(err, syscall.ECONNREFUSED) || errors.Is(err, syscall.ENETUNREACH) || errors.Is(err, syscall.EHOSTUNREACH) {
		return true
	}
	// url.Error wrapping net errors
	var ue *url.Error
	if errors.As(err, &ue) {
		if errors.Is(ue.Err, syscall.ECONNREFUSED) || errors.Is(ue.Err, syscall.ENETUNREACH) || errors.Is(ue.Err, syscall.EHOSTUNREACH) {
			return true
		}
		// last resort substring check
		if strings.Contains(strings.ToLower(ue.Err.Error()), "connection refused") {
			return true
		}
	}
	// last resort substring on raw error
	if strings.Contains(strings.ToLower(err.Error()), "connection refused") {
		return true
	}
	return false
}

// requestJSON sends JSON and decodes JSON.
// It avoids HTML-escaping to match Python client's encoding behavior.
func requestJSON[Req any, Res any](ctx context.Context, c *Client, method, path string, in *Req) (*Res, error) {
    var body io.Reader
    if in != nil {
        b, err := json.Marshal(in)
        if err != nil {
            return nil, err
        }
        body = bytes.NewReader(b)
    }
    resp, err := c.do(ctx, method, path, body, nil)
    if err != nil {
        return nil, err
    }
    defer func() { _ = resp.Body.Close() }()
    var out Res
    dec := json.NewDecoder(resp.Body)
    if err := dec.Decode(&out); err != nil {
        return nil, err
    }
    return &out, nil
}

// Generate
// Generate performs a non-streaming generation request.
func (c *Client) Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	if err := ensureModel(req.BaseStreamableRequest.Model); err != nil {
		return nil, err
	}
	return requestJSON[GenerateRequest, GenerateResponse](ctx, c, http.MethodPost, "/api/generate", req)
}

// GenerateStream performs a streaming generation request and returns a Stream.
func (c *Client) GenerateStream(ctx context.Context, req *GenerateRequest) (*Stream[GenerateResponse], error) {
	if err := ensureModel(req.BaseStreamableRequest.Model); err != nil {
		return nil, err
	}
    s := true
    req.Stream = &s
    b, err := json.Marshal(req)
    if err != nil { return nil, err }
    resp, err := c.do(ctx, http.MethodPost, "/api/generate", bytes.NewReader(b), nil)
    if err != nil {
        return nil, err
    }
    return newStream[GenerateResponse](ctx, resp), nil
}

// Chat
// Chat performs a non-streaming chat request.
func (c *Client) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	if err := ensureModel(req.BaseStreamableRequest.Model); err != nil {
		return nil, err
	}
	return requestJSON[ChatRequest, ChatResponse](ctx, c, http.MethodPost, "/api/chat", req)
}

// ChatStream performs a streaming chat request and returns a Stream.
func (c *Client) ChatStream(ctx context.Context, req *ChatRequest) (*Stream[ChatResponse], error) {
	if err := ensureModel(req.BaseStreamableRequest.Model); err != nil {
		return nil, err
	}
    s := true
    req.Stream = &s
    b, err := json.Marshal(req)
    if err != nil { return nil, err }
    resp, err := c.do(ctx, http.MethodPost, "/api/chat", bytes.NewReader(b), nil)
    if err != nil {
        return nil, err
    }
    return newStream[ChatResponse](ctx, resp), nil
}

// Embed
// Embed requests embeddings from models that support /api/embed.
func (c *Client) Embed(ctx context.Context, req *EmbedRequest) (*EmbedResponse, error) {
	if err := ensureModel(req.Model); err != nil {
		return nil, err
	}
	return requestJSON[EmbedRequest, EmbedResponse](ctx, c, http.MethodPost, "/api/embed", req)
}

// Embeddings (deprecated)
// Embeddings requests embeddings via the deprecated /api/embeddings endpoint.
func (c *Client) Embeddings(ctx context.Context, req *EmbeddingsRequest) (*EmbeddingsResponse, error) {
	if err := ensureModel(req.Model); err != nil {
		return nil, err
	}
	return requestJSON[EmbeddingsRequest, EmbeddingsResponse](ctx, c, http.MethodPost, "/api/embeddings", req)
}

// Pull
// Pull pulls a model; returns a final progress snapshot.
func (c *Client) Pull(ctx context.Context, req *PullRequest) (*ProgressResponse, error) {
	if err := ensureModel(req.Model); err != nil {
		return nil, err
	}
	return requestJSON[PullRequest, ProgressResponse](ctx, c, http.MethodPost, "/api/pull", req)
}

// PullStream pulls a model and returns a progress stream.
func (c *Client) PullStream(ctx context.Context, req *PullRequest) (*Stream[ProgressResponse], error) {
	if err := ensureModel(req.Model); err != nil {
		return nil, err
	}
    s := true
    req.Stream = &s
    b, err := json.Marshal(req)
    if err != nil { return nil, err }
    resp, err := c.do(ctx, http.MethodPost, "/api/pull", bytes.NewReader(b), nil)
	if err != nil {
		return nil, err
	}
	return newStream[ProgressResponse](ctx, resp), nil
}

// Push
// Push pushes a model; returns a final progress snapshot.
func (c *Client) Push(ctx context.Context, req *PushRequest) (*ProgressResponse, error) {
	if err := ensureModel(req.Model); err != nil {
		return nil, err
	}
	return requestJSON[PushRequest, ProgressResponse](ctx, c, http.MethodPost, "/api/push", req)
}

// PushStream pushes a model and returns a progress stream.
func (c *Client) PushStream(ctx context.Context, req *PushRequest) (*Stream[ProgressResponse], error) {
	if err := ensureModel(req.Model); err != nil {
		return nil, err
	}
    s := true
    req.Stream = &s
    b, err := json.Marshal(req)
    if err != nil { return nil, err }
    resp, err := c.do(ctx, http.MethodPost, "/api/push", bytes.NewReader(b), nil)
	if err != nil {
		return nil, err
	}
	return newStream[ProgressResponse](ctx, resp), nil
}

// Create
// Create creates a model; returns a final progress snapshot.
func (c *Client) Create(ctx context.Context, req *CreateRequest) (*ProgressResponse, error) {
	if err := ensureModel(req.Model); err != nil {
		return nil, err
	}
	return requestJSON[CreateRequest, ProgressResponse](ctx, c, http.MethodPost, "/api/create", req)
}

// CreateStream creates a model and returns a progress stream.
func (c *Client) CreateStream(ctx context.Context, req *CreateRequest) (*Stream[ProgressResponse], error) {
	if err := ensureModel(req.Model); err != nil {
		return nil, err
	}
    s := true
    req.Stream = &s
    b, err := json.Marshal(req)
    if err != nil { return nil, err }
    resp, err := c.do(ctx, http.MethodPost, "/api/create", bytes.NewReader(b), nil)
	if err != nil {
		return nil, err
	}
	return newStream[ProgressResponse](ctx, resp), nil
}

// Blobs
// CreateBlob uploads a blob content-addressed by sha256:hex digest.
func (c *Client) CreateBlob(ctx context.Context, path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
    defer func(){ _ = f.Close() }()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	digest := fmt.Sprintf("sha256:%x", h.Sum(nil))
	// upload
	if _, err := f.Seek(0, 0); err != nil {
		return "", err
	}
	resp, err := c.do(ctx, http.MethodPost, "/api/blobs/"+digest, f, http.Header{})
	if err != nil {
		return "", err
	}
    _ = resp.Body.Close()
	return digest, nil
}

// List models
// List returns installed model tags.
func (c *Client) List(ctx context.Context) (*ListResponse, error) {
	return requestJSON[struct{}, ListResponse](ctx, c, http.MethodGet, "/api/tags", nil)
}

// Delete removes a model by name and returns a Python-parity status.
func (c *Client) Delete(ctx context.Context, model string) (*StatusResponse, error) {
	resp, err := c.do(ctx, http.MethodDelete, "/api/delete", jsonBody(DeleteRequest{Model: model}), nil)
	if err != nil {
		// newResponseError already returned error; but we need status mapping like Python
		return &StatusResponse{Status: StrPtr("error")}, nil
	}
	st := "error"
	if resp.StatusCode == http.StatusOK {
		st = "success"
	}
    _, _ = io.Copy(io.Discard, resp.Body)
    _ = resp.Body.Close()
	return &StatusResponse{Status: StrPtr(st)}, nil
}

// Copy duplicates a model.
func (c *Client) Copy(ctx context.Context, source, destination string) (*StatusResponse, error) {
	resp, err := c.do(ctx, http.MethodPost, "/api/copy", jsonBody(CopyRequest{Source: source, Destination: destination}), nil)
	if err != nil {
		return &StatusResponse{Status: StrPtr("error")}, nil
	}
	st := "error"
	if resp.StatusCode == http.StatusOK {
		st = "success"
	}
    _, _ = io.Copy(io.Discard, resp.Body)
    _ = resp.Body.Close()
	return &StatusResponse{Status: StrPtr(st)}, nil
}

// Show returns model information for a given tag.
func (c *Client) Show(ctx context.Context, model string) (*ShowResponse, error) {
	return requestJSON[ShowRequest, ShowResponse](ctx, c, http.MethodPost, "/api/show", &ShowRequest{Model: model})
}

// PS lists running models/processes.
func (c *Client) PS(ctx context.Context) (*ProcessResponse, error) {
	return requestJSON[struct{}, ProcessResponse](ctx, c, http.MethodGet, "/api/ps", nil)
}

// helpers
// ensureModel validates a non-empty model.
func ensureModel(m string) error {
	if m == "" {
		return &RequestError{Message: "model is required"}
	}
	return nil
}

func jsonBody(v any) io.Reader {
    b, _ := json.Marshal(v)
    return bytes.NewReader(b)
}
