package ollama

import (
	"testing"
)

func TestNewResponseError_JSONAndRaw(t *testing.T) {
	e := newResponseError(400, []byte(`{"error":"nope"}`))
	if re, ok := e.(*ResponseError); !ok || re.Message != "nope" || re.StatusCode != 400 {
		t.Fatalf("bad json response error: %#v", e)
	}
	e2 := newResponseError(500, []byte("something broke"))
	if re, ok := e2.(*ResponseError); !ok || re.Message != "something broke" || re.StatusCode != 500 {
		t.Fatalf("bad raw response error: %#v", e2)
	}
}
