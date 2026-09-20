package webfs

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
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

func TestServeFileSystem(t *testing.T) {
	t.Parallel()

	const (
		indexHTML = "<!doctype html><html><body>maestro</body></html>"
		appJS     = "console.log(\"maestro\");"
	)

	withIndex := fstest.MapFS{
		"index.html": {Data: []byte(indexHTML)},
		"app.js":     {Data: []byte(appJS)},
	}

	withoutIndex := fstest.MapFS{
		"app.js": {Data: []byte(appJS)},
	}

	tests := []struct {
		name       string
		fsys       fstest.MapFS
		path       string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "serves existing file",
			fsys:       withIndex,
			path:       "/app.js",
			wantStatus: http.StatusOK,
			wantBody:   appJS,
		},
		{
			name:       "falls back to index.html for unknown path",
			fsys:       withIndex,
			path:       "/does-not-exist",
			wantStatus: http.StatusOK,
			wantBody:   indexHTML,
		},
		{
			name:       "falls back to index.html for nested spa route",
			fsys:       withIndex,
			path:       "/connectors/123",
			wantStatus: http.StatusOK,
			wantBody:   indexHTML,
		},
		{
			name:       "falls back to index.html for root",
			fsys:       withIndex,
			path:       "/",
			wantStatus: http.StatusOK,
			wantBody:   indexHTML,
		},
		{
			name:       "returns placeholder without index.html for existing file path",
			fsys:       withoutIndex,
			path:       "/app.js",
			wantStatus: http.StatusNotFound,
			wantBody:   "maestro web UI is not embedded in this build\n",
		},
		{
			name:       "returns placeholder without index.html for unknown path",
			fsys:       withoutIndex,
			path:       "/does-not-exist",
			wantStatus: http.StatusNotFound,
			wantBody:   "maestro web UI is not embedded in this build\n",
		},
		{
			name:       "returns placeholder for empty filesystem",
			fsys:       fstest.MapFS{},
			path:       "/",
			wantStatus: http.StatusNotFound,
			wantBody:   "maestro web UI is not embedded in this build\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)

			h := serveFileSystem(tt.fsys)

			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.path, http.NoBody))

			must.Equal(tt.wantStatus, rec.Code)
			must.Equal(tt.wantBody, rec.Body.String())
		})
	}
}
