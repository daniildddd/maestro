package webfs

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_ServesPlaceholderWithoutDist(t *testing.T) {
	t.Parallel()

	h := Handler()

	for _, path := range []string{"/", "/connectors", "/assets/app.js"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, http.NoBody))

		if rec.Code != http.StatusNotFound {
			t.Fatalf("GET %s: got status %d, want 404", path, rec.Code)
		}
	}
}
