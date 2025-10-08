package httpmw_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sky0621/go_work_sample/core/pkg/httpmw"
)

func TestAPIKeyMiddleware(t *testing.T) {
	handler := httpmw.APIKey(httpmw.APIKeyConfig{Key: "secret", Prefix: "Bearer "})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Result().StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized, got %d", rr.Result().StatusCode)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("Authorization", "Bearer secret")
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	if rr2.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected ok, got %d", rr2.Result().StatusCode)
	}
}
