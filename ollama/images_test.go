package ollama

import (
	"encoding/base64"
	"testing"
)

func TestImage_Marshal_PathAndBase64(t *testing.T) {
	// base64 input passes through
	raw := []byte("hello world")
	b64 := base64.StdEncoding.EncodeToString(raw)
	im := Image{Value: b64}
	js, err := im.MarshalJSON()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if string(js) != "\""+b64+"\"" {
		t.Fatalf("unexpected: %s", js)
	}

	// file path with image extension must exist
	path, err := tmpFileWith(".png", raw)
	if err != nil {
		t.Fatal(err)
	}
	im2 := Image{Value: path}
	js2, err := im2.MarshalJSON()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if string(js2) != "\""+b64+"\"" {
		t.Fatalf("unexpected: %s", js2)
	}

	// file path with non-image extension still reads as file if exists
	path2, err := tmpFileWith(".bin", raw)
	if err != nil {
		t.Fatal(err)
	}
	im3 := Image{Value: path2}
	js3, err := im3.MarshalJSON()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if string(js3) != "\""+b64+"\"" {
		t.Fatalf("unexpected: %s", js3)
	}
}
