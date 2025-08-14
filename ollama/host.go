package ollama

import (
	"fmt"
	"net/url"
	"strings"
)

// parseHost replicates Python _parse_host behavior.
// - default scheme http
// - default port 11434 unless scheme http(80) or https(443) and port omitted
// - preserves paths, trims trailing slash
func parseHost(in string) string {
	host := strings.TrimSpace(in)
	if host == "" {
		return ""
	}

	// split scheme
	scheme := "http"
	rest := host
	if s, after, found := strings.Cut(host, "://"); found {
		scheme, rest = s, after
	}

	// urlsplit to reuse standard parsing
	u, err := url.Parse(fmt.Sprintf("%s://%s", scheme, rest))
	if err != nil {
		// fallback to default host
		return fmt.Sprintf("%s://127.0.0.1:11434", scheme)
	}

	h := u.Hostname()
	p := u.Port()
	path := strings.Trim(u.EscapedPath(), "/")

	// default port
	defPort := "11434"
	switch scheme {
	case "http":
		if p == "" {
			p = "80"
		} // as in Python when scheme set explicitly
	case "https":
		if p == "" {
			p = "443"
		}
	}
    // leave port as determined by scheme/explicit port or default

	if s, after, found := strings.Cut(host, "://"); !found {
		// If scheme wasn't specified, apply default port 11434 unless path gives port
		scheme = "http"
		if u.Port() == "" {
			p = defPort
		} else {
			p = u.Port()
		}
	} else {
		_ = s
		_ = after
		// If user gave scheme but no port, keep scheme standard port (handled above)
		if u.Port() == "" && (scheme != "http" && scheme != "https") {
			p = defPort
		}
	}

	// IPv6 without brackets? url.Parse already brackets hostname if constructed via scheme://rest
	if strings.Contains(h, ":") && !strings.HasPrefix(h, "[") {
		h = "[" + h + "]"
	}

	base := fmt.Sprintf("%s://%s:%s", scheme, h, func() string {
		if p == "" {
			return defPort
		}
		return p
	}())
	if path != "" {
		return base + "/" + path
	}
	return base
}
