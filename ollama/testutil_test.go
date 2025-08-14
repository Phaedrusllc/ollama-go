package ollama

import "os"

// tmpFileWith creates a temp file with the given extension and writes data.
func tmpFileWith(ext string, data []byte) (string, error) {
	f, err := os.CreateTemp("", "img-*"+ext)
	if err != nil {
		return "", err
	}
    defer func(){ _ = f.Close() }()
	if _, err := f.Write(data); err != nil {
		return "", err
	}
	return f.Name(), nil
}
