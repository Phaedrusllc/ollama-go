package ollama

import "testing"

func TestImage_InvalidBase64(t *testing.T) {
	im := Image{Value: "@@notbase64@@"}
	if _, err := im.MarshalJSON(); err == nil {
		t.Fatal("expected error for invalid base64")
	}
}

func TestImage_MissingImagePath(t *testing.T) {
	im := Image{Value: "/definitely/not/there.png"}
	if _, err := im.MarshalJSON(); err == nil {
		t.Fatal("expected error for missing image path")
	}
}
