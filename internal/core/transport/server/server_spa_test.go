package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHTTPServer_RegisterSPA(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "root serves placeholder when dist is not embedded",
			path:       "/",
			wantStatus: http.StatusNotFound,
			wantBody:   "maestro web UI is not embedded in this build\n",
		},
		{
			name:       "api route is not intercepted by spa handler",
			path:       "/api/v1/x",
			wantStatus: http.StatusNoContent,
			wantBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)

			srv := NewHTTPServer(Config{Addr: "127.0.0.1:0"}, nopLogger())
			srv.RegisterSPA()
			srv.RegisterRoute(Route{
				Method: http.MethodGet,
				Path:   "/api/v1/x",
				Handler: func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusNoContent)
				},
			})

			rec := httptest.NewRecorder()
			srv.mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.path, http.NoBody))

			must.Equal(tt.wantStatus, rec.Code)
			must.Equal(tt.wantBody, rec.Body.String())
		})
	}
}
