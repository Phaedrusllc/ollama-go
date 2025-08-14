package ollama

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

// Options mirrors Python Options.
type Options struct {
	// Load-time
	Numa          *bool `json:"numa,omitempty"`
	NumCtx        *int  `json:"num_ctx,omitempty"`
	NumBatch      *int  `json:"num_batch,omitempty"`
	NumGPU        *int  `json:"num_gpu,omitempty"`
	MainGPU       *int  `json:"main_gpu,omitempty"`
	LowVRAM       *bool `json:"low_vram,omitempty"`
	F16KV         *bool `json:"f16_kv,omitempty"`
	LogitsAll     *bool `json:"logits_all,omitempty"`
	VocabOnly     *bool `json:"vocab_only,omitempty"`
	UseMMap       *bool `json:"use_mmap,omitempty"`
	UseMLock      *bool `json:"use_mlock,omitempty"`
	EmbeddingOnly *bool `json:"embedding_only,omitempty"`
	NumThread     *int  `json:"num_thread,omitempty"`
	// Runtime
	NumKeep          *int     `json:"num_keep,omitempty"`
	Seed             *int     `json:"seed,omitempty"`
	NumPredict       *int     `json:"num_predict,omitempty"`
	TopK             *int     `json:"top_k,omitempty"`
	TopP             *float64 `json:"top_p,omitempty"`
	TFSZ             *float64 `json:"tfs_z,omitempty"`
	TypicalP         *float64 `json:"typical_p,omitempty"`
	RepeatLastN      *int     `json:"repeat_last_n,omitempty"`
	Temperature      *float64 `json:"temperature,omitempty"`
	RepeatPenalty    *float64 `json:"repeat_penalty,omitempty"`
	PresencePenalty  *float64 `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float64 `json:"frequency_penalty,omitempty"`
	Mirostat         *int     `json:"mirostat,omitempty"`
	MirostatTau      *float64 `json:"mirostat_tau,omitempty"`
	MirostatEta      *float64 `json:"mirostat_eta,omitempty"`
	PenalizeNewline  *bool    `json:"penalize_newline,omitempty"`
	Stop             []string `json:"stop,omitempty"`
}

// BaseStreamableRequest holds common fields for streamable requests.
type BaseStreamableRequest struct {
	Model     string `json:"model"`
	Stream    *bool  `json:"stream,omitempty"`
	Options   any    `json:"options,omitempty"`
	Format    any    `json:"format,omitempty"`
	KeepAlive any    `json:"keep_alive,omitempty"`
}

// Image represents image input as path, raw bytes, or base64 string.
type Image struct{ Value any }

func (i Image) MarshalJSON() ([]byte, error) {
	switch v := i.Value.(type) {
	case []byte:
		s := base64.StdEncoding.EncodeToString(v)
		return json.Marshal(s)
	case string:
		// Try file path first (any extension) if it exists
		if fi, err := os.Stat(v); err == nil && !fi.IsDir() {
			b, err := os.ReadFile(v)
			if err != nil {
				return nil, err
			}
			return json.Marshal(base64.StdEncoding.EncodeToString(b))
		}
		// If extension looks image-like but file not found -> raise same as Python
		lower := strings.ToLower(v)
		if strings.HasSuffix(lower, ".png") || strings.HasSuffix(lower, ".jpg") || strings.HasSuffix(lower, ".jpeg") || strings.HasSuffix(lower, ".webp") {
			return nil, fmt.Errorf("File %s does not exist", v)
		}
		// Else try to decode as base64 to validate; if ok, pass through
		if _, err := base64.StdEncoding.DecodeString(v); err == nil {
			return json.Marshal(v)
		}
		return nil, errors.New("Invalid image data, expected base64 string or path to image file")
	default:
		return nil, errors.New("Invalid image data type")
	}
}

// GenerateRequest is the payload for /api/generate.
type GenerateRequest struct {
	BaseStreamableRequest
	Prompt   *string `json:"prompt,omitempty"`
	Suffix   *string `json:"suffix,omitempty"`
	System   *string `json:"system,omitempty"`
	Template *string `json:"template,omitempty"`
	Context  []int   `json:"context,omitempty"`
	Raw      *bool   `json:"raw,omitempty"`
	Images   []Image `json:"images,omitempty"`
	Think    any     `json:"think,omitempty"`
}

// BaseGenerateResponse contains timing and token accounting metadata.
type BaseGenerateResponse struct {
	Model           *string `json:"model,omitempty"`
	CreatedAt       *string `json:"created_at,omitempty"`
	Done            *bool   `json:"done,omitempty"`
	DoneReason      *string `json:"done_reason,omitempty"`
	TotalDuration   *int64  `json:"total_duration,omitempty"`
	LoadDuration    *int64  `json:"load_duration,omitempty"`
	PromptEvalCount *int    `json:"prompt_eval_count,omitempty"`
	PromptEvalDur   *int64  `json:"prompt_eval_duration,omitempty"`
	EvalCount       *int    `json:"eval_count,omitempty"`
	EvalDuration    *int64  `json:"eval_duration,omitempty"`
}

// GenerateResponse is returned by generate endpoints (including stream chunks).
type GenerateResponse struct {
	BaseGenerateResponse
	Response string  `json:"response"`
	Thinking *string `json:"thinking,omitempty"`
	Context  []int   `json:"context,omitempty"`
}

// Message represents a chat message or response fragment.
type Message struct {
	Role      string     `json:"role"`
	Content   *string    `json:"content,omitempty"`
	Thinking  *string    `json:"thinking,omitempty"`
	Images    []Image    `json:"images,omitempty"`
	ToolName  *string    `json:"tool_name,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// Accessors helpful in demos/tests
func (m Message) GetContent() string {
	if m.Content == nil {
		return ""
	}
	return *m.Content
}

// ToolCall represents a function call requested by the model.
type ToolCall struct {
	Function ToolCallFunction `json:"function"`
}

// ToolCallFunction describes the callable function and its arguments.
type ToolCallFunction struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

// Tool declares a function that the model can call.
type Tool struct {
	Type     *string       `json:"type,omitempty"`
	Function *ToolFunction `json:"function,omitempty"`
}

// ToolFunction describes the function metadata and parameters schema.
type ToolFunction struct {
	Name        *string               `json:"name,omitempty"`
	Description *string               `json:"description,omitempty"`
	Parameters  *ToolJSONSchemaObject `json:"parameters,omitempty"`
}

// ToolJSONSchemaObject supports mapping defs -> $defs on marshal.
type ToolJSONSchemaObject struct {
	Type       any                       `json:"type,omitempty"`
	Defs       any                       `json:"-"` // user sets here
	Items      any                       `json:"items,omitempty"`
	Required   []string                  `json:"required,omitempty"`
	Properties map[string]map[string]any `json:"properties,omitempty"`
}

func (o ToolJSONSchemaObject) MarshalJSON() ([]byte, error) {
	m := map[string]any{}
	if o.Type != nil {
		m["type"] = o.Type
	}
	if o.Items != nil {
		m["items"] = o.Items
	}
	if len(o.Required) > 0 {
		m["required"] = o.Required
	}
	if len(o.Properties) > 0 {
		m["properties"] = o.Properties
	}
	if o.Defs != nil {
		m["$defs"] = o.Defs
	}
	return json.Marshal(m)
}

// ChatRequest is the payload for /api/chat.
type ChatRequest struct {
	BaseStreamableRequest
	Messages []Message `json:"messages,omitempty"`
	Tools    []Tool    `json:"tools,omitempty"`
	Think    any       `json:"think,omitempty"`
}

// ChatResponse is returned by chat endpoints (including stream chunks).
type ChatResponse struct {
	BaseGenerateResponse
	Message Message `json:"message"`
}

// EmbedRequest is the payload for /api/embed (new embeddings API).
type EmbedRequest struct {
	Model     string `json:"model"`
	Input     any    `json:"input"`
	Truncate  *bool  `json:"truncate,omitempty"`
	Options   any    `json:"options,omitempty"`
	KeepAlive any    `json:"keep_alive,omitempty"`
}

// EmbedResponse returns a batch of embeddings.
type EmbedResponse struct {
	Embeddings [][]float64 `json:"embeddings"`
}

// EmbeddingsRequest is the deprecated embeddings payload.
// Deprecated: use EmbedRequest/EmbedResponse endpoints.
type EmbeddingsRequest struct {
	Model     string  `json:"model"`
	Prompt    *string `json:"prompt,omitempty"`
	Options   any     `json:"options,omitempty"`
	KeepAlive any     `json:"keep_alive,omitempty"`
}

// EmbeddingsResponse returns a single embedding vector.
type EmbeddingsResponse struct {
	Embedding []float64 `json:"embedding"`
}

// PullRequest requests pulling a model; supports streaming.
type PullRequest struct {
	Model    string `json:"model"`
	Stream   *bool  `json:"stream,omitempty"`
	Insecure *bool  `json:"insecure,omitempty"`
}

// PushRequest requests pushing a model; supports streaming.
type PushRequest struct {
	Model    string `json:"model"`
	Stream   *bool  `json:"stream,omitempty"`
	Insecure *bool  `json:"insecure,omitempty"`
}

// CreateRequest creates a model; supports streaming progress.
type CreateRequest struct {
	Model      string            `json:"model"`
	Stream     *bool             `json:"stream,omitempty"`
	Quantize   *string           `json:"quantize,omitempty"`
	From       *string           `json:"from,omitempty"` // maps from_ in Python
	Files      map[string]string `json:"files,omitempty"`
	Adapters   map[string]string `json:"adapters,omitempty"`
	Template   *string           `json:"template,omitempty"`
	License    any               `json:"license,omitempty"`
	System     *string           `json:"system,omitempty"`
	Parameters any               `json:"parameters,omitempty"`
	Messages   []Message         `json:"messages,omitempty"`
}

// ListResponse lists installed models.
type ListResponse struct {
	Models []ListModel `json:"models"`
}

// ListModel describes a single entry in ListResponse.
type ListModel struct {
	Model      *string       `json:"model,omitempty"`
	ModifiedAt *time.Time    `json:"modified_at,omitempty"`
	Digest     *string       `json:"digest,omitempty"`
	Size       *int64        `json:"size,omitempty"`
	Details    *ModelDetails `json:"details,omitempty"`
}

// ModelDetails holds metadata about a model.
type ModelDetails struct {
	ParentModel       *string  `json:"parent_model,omitempty"`
	Format            *string  `json:"format,omitempty"`
	Family            *string  `json:"family,omitempty"`
	Families          []string `json:"families,omitempty"`
	ParameterSize     *string  `json:"parameter_size,omitempty"`
	QuantizationLevel *string  `json:"quantization_level,omitempty"`
}

type DeleteRequest struct {
	Model string `json:"model"`
}

// CopyRequest requests copying a model.
type CopyRequest struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
}

// StatusResponse reports success or error for management operations.
type StatusResponse struct {
	Status *string `json:"status,omitempty"`
}

// ProgressResponse reports progress for streaming operations.
type ProgressResponse struct {
	Status    *string `json:"status,omitempty"`
	Completed *int64  `json:"completed,omitempty"`
	Total     *int64  `json:"total,omitempty"`
	Digest    *string `json:"digest,omitempty"`
}

type ShowRequest struct {
	Model string `json:"model"`
}

// ShowResponse returns model metadata and details.
type ShowResponse struct {
	ModifiedAt   *time.Time     `json:"modified_at,omitempty"`
	Template     *string        `json:"template,omitempty"`
	Modelfile    *string        `json:"modelfile,omitempty"`
	License      *string        `json:"license,omitempty"`
	Details      *ModelDetails  `json:"details,omitempty"`
	ModelInfo    map[string]any `json:"model_info,omitempty"`
	Parameters   *string        `json:"parameters,omitempty"`
	Capabilities []string       `json:"capabilities,omitempty"`
}

// ProcessResponse lists currently running models.
type ProcessResponse struct {
	Models []ProcessModel `json:"models"`
}

// ProcessModel describes a single running model/process.
type ProcessModel struct {
	Model         *string       `json:"model,omitempty"`
	Name          *string       `json:"name,omitempty"`
	Digest        *string       `json:"digest,omitempty"`
	ExpiresAt     *time.Time    `json:"expires_at,omitempty"`
	Size          *int64        `json:"size,omitempty"`
	SizeVRAM      *int64        `json:"size_vram,omitempty"`
	Details       *ModelDetails `json:"details,omitempty"`
	ContextLength *int          `json:"context_length,omitempty"`
}

// StrPtr returns a pointer to the given string.
func StrPtr(s string) *string { return &s }
