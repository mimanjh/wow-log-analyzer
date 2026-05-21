package handlers

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

//go:embed all:web
var webFS embed.FS

// NewStaticHandler serves the embedded Vite build with SPA-style fallback to
// index.html. Requests under /api/ are not handled here — they hit specific
// API routes registered earlier on the mux, and any unmatched /api/ path
// returns 404 from this handler rather than the SPA index.
func NewStaticHandler() http.Handler {
	sub, err := fs.Sub(webFS, "web")
	if err != nil {
		panic(err)
	}
	fileServer := http.FileServer(http.FS(sub))

	indexBytes, err := fs.ReadFile(sub, "index.html")
	if err != nil {
		panic("frontend bundle missing index.html: " + err.Error())
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		// Defensive: anything under /api/ that reached the catch-all is an
		// unknown API route — return 404 instead of the SPA shell.
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		clean := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if clean == "" {
			serveIndex(w, indexBytes)
			return
		}
		if _, err := fs.Stat(sub, clean); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}
		// Paths with a file extension are almost always missing assets, not
		// SPA routes — return 404 instead of the HTML shell.
		if path.Ext(clean) != "" {
			http.NotFound(w, r)
			return
		}
		// SPA fallback — let React Router resolve the path client-side.
		serveIndex(w, indexBytes)
	})
}

func serveIndex(w http.ResponseWriter, indexBytes []byte) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	_, _ = w.Write(indexBytes)
}
