package webfs

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed all:dist
var embedded embed.FS

func Handler() http.Handler {
	sub, err := fs.Sub(embedded, "dist")
	if err != nil {
		return placeholder()
	}

	if _, err := fs.Stat(sub, "index.html"); err != nil {
		return placeholder()
	}

	files := http.FileServer(http.FS(sub))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := fs.Stat(sub, strings.TrimPrefix(r.URL.Path, "/")); err != nil {
			http.ServeFileFS(w, r, sub, "index.html")

			return
		}

		files.ServeHTTP(w, r)
	})
}

func placeholder() http.Handler {
	const msg = "maestro web UI is not embedded in this build"

	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)

		if _, err := w.Write([]byte(msg + "\n")); err != nil {
			return
		}
	})
}
