package ollama

import (
	"encoding/json"
	"fmt"
)

// RequestError mirrors the Python client RequestError: client-side validation
// failure (e.g., missing model).
type RequestError struct{ Message string }

func (e *RequestError) Error() string { return e.Message }

// ResponseError mirrors the Python client ResponseError: extracts JSON `error`
// and carries the HTTP status code when available.
type ResponseError struct {
	Message    string
	StatusCode int
}

func (e *ResponseError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("%s (status code: %d)", e.Message, e.StatusCode)
	}
	return e.Message
}

// newResponseError tries to read `{ "error": "..." }` else uses raw body.
func newResponseError(status int, body []byte) error {
	var tmp struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(body, &tmp); err == nil && tmp.Error != "" {
		return &ResponseError{Message: tmp.Error, StatusCode: status}
	}
	return &ResponseError{Message: string(body), StatusCode: status}
}

// ConnectionError mirrors Python's friendly connect error message for failed
// connections to the Ollama server.
type ConnectionError struct{ Message string }

func (e *ConnectionError) Error() string { return e.Message }

const connectionErrorMessage = "Failed to connect to Ollama. Please check that Ollama is downloaded, running and accessible. https://ollama.com/download"
