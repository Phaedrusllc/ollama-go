package ollama

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestCreateRequest_FromField(t *testing.T) {
	from := "base:latest"
	r := &CreateRequest{Model: "m", From: &from}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	if !containsJSON(b, `"from":"base:latest"`) {
		t.Fatalf("missing from: %s", string(b))
	}
}

func TestToolParams_DefsToDollarDefs(t *testing.T) {
	tparams := &ToolJSONSchemaObject{Type: "object", Defs: map[string]any{"X": 1}}
	tf := &ToolFunction{Name: StrPtr("fn"), Parameters: tparams}
	tool := Tool{Function: tf}
	b, err := json.Marshal(tool)
	if err != nil {
		t.Fatal(err)
	}
	if !containsJSON(b, `"$defs"`) {
		t.Fatalf("expected $defs: %s", string(b))
	}
}

func containsJSON(b []byte, sub string) bool { return bytes.Contains(b, []byte(sub)) }
