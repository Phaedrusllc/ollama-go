package ollama

import "testing"

func TestParseHost_PythonParity(t *testing.T) {
	cases := map[string]string{
		"http://example.com":        "http://example.com:80",
		"https://example.com":       "https://example.com:443",
		"https://example.com:56789": "https://example.com:56789",
		"example.com":               "http://example.com:11434",
		"example.com:56789":         "http://example.com:56789",
		// duplicate removed
		"http://127.0.0.1":                          "http://127.0.0.1:80",
		"http://127.0.0.1:56789":                    "http://127.0.0.1:56789",
		"https://127.0.0.1":                         "https://127.0.0.1:443",
		"https://127.0.0.1:56789":                   "https://127.0.0.1:56789",
		"127.0.0.1":                                 "http://127.0.0.1:11434",
		"127.0.0.1:56789":                           "http://127.0.0.1:56789",
		"http://127.0.0.1:56789/":                   "http://127.0.0.1:56789",
		"http://localhost":                          "http://localhost:80",
		"http://localhost:56789":                    "http://localhost:56789",
		"https://localhost":                         "https://localhost:443",
		"https://localhost:56789":                   "https://localhost:56789",
		"localhost":                                 "http://localhost:11434",
		"localhost:56789":                           "http://localhost:56789",
		"http://localhost:56789/":                   "http://localhost:56789",
		"https://example.com:56789/":                "https://example.com:56789",
		"example.com/":                              "http://example.com:11434",
		"example.com:56789/":                        "http://example.com:56789",
		"example.com/path":                          "http://example.com:11434/path",
		"example.com:56789/path":                    "http://example.com:56789/path",
		"https://example.com:56789/path":            "https://example.com:56789/path",
		"example.com:56789/path/":                   "http://example.com:56789/path",
		"[0001:002:003:0004::1]":                    "http://[0001:002:003:0004::1]:11434",
		"[0001:002:003:0004::1]:56789":              "http://[0001:002:003:0004::1]:56789",
		"http://[0001:002:003:0004::1]":             "http://[0001:002:003:0004::1]:80",
		"https://[0001:002:003:0004::1]":            "https://[0001:002:003:0004::1]:443",
		"https://[0001:002:003:0004::1]:56789":      "https://[0001:002:003:0004::1]:56789",
		"[0001:002:003:0004::1]/":                   "http://[0001:002:003:0004::1]:11434",
		"[0001:002:003:0004::1]:56789/":             "http://[0001:002:003:0004::1]:56789",
		"[0001:002:003:0004::1]/path":               "http://[0001:002:003:0004::1]:11434/path",
		"[0001:002:003:0004::1]:56789/path":         "http://[0001:002:003:0004::1]:56789/path",
		"https://[0001:002:003:0004::1]:56789/path": "https://[0001:002:003:0004::1]:56789/path",
		"[0001:002:003:0004::1]:56789/path/":        "http://[0001:002:003:0004::1]:56789/path",
	}
	for in, want := range cases {
		if got := parseHost(in); got != want {
			t.Errorf("parseHost(%q) = %q; want %q", in, got, want)
		}
	}
}
