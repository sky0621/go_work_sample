package httpmw

import (
	"net/http"
	"strings"
)

// APIKeyConfig defines options for API key authentication middleware.
type APIKeyConfig struct {
	Header string
	Prefix string
	Key    string
}

// APIKey enforces presence of a static key on incoming requests.
func APIKey(cfg APIKeyConfig) func(http.Handler) http.Handler {
	header := cfg.Header
	if header == "" {
		header = "Authorization"
	}
	prefix := cfg.Prefix

	expected := strings.TrimSpace(cfg.Key)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if expected == "" {
				next.ServeHTTP(w, r)
				return
			}

			value := strings.TrimSpace(r.Header.Get(header))
			if prefix != "" {
				if !strings.HasPrefix(strings.ToLower(value), strings.ToLower(prefix)) {
					unauthorized(w)
					return
				}
				value = strings.TrimSpace(value[len(prefix):])
			}

			if value != expected {
				unauthorized(w)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
}
