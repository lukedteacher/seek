package resources

import (
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// sets directory of static assets (styles, scripts, images, etc.)

// static asset handler
// func Handler() http.Handler {
// 	// allows direct subdirectory access for static assets
// 	// example: styles/global.css instead of static/styles/global.css
// 	files := http.StripPrefix("/static/", http.FileServer(http.Dir(StaticDirectoryPath)))
// 	// serves the files
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.Header().Set("Cache-Control", "no-cache") // changed from "no-store" to speed up page loading
// 		files.ServeHTTP(w, r)
// 	})
// }

var StaticDirectoryPath string

func init() {
	// if STATIC_DIR is set, use that absolute path.
	if dir := os.Getenv("STATIC_DIR"); dir != "" {
		StaticDirectoryPath = dir
		return
	}
	// default: relative to working directory (for local dev)
	StaticDirectoryPath = "static"
}

func Handler() http.Handler {
	fs := http.FileServer(http.Dir(StaticDirectoryPath))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// strip prefix and get relative path
		rel := strings.TrimPrefix(r.URL.Path, "/static/")
		// check if file exists
		if _, err := os.Stat(filepath.Join(StaticDirectoryPath, rel)); err != nil {
			log.Printf("file not found: %s", rel)
		}
		http.StripPrefix("/static/", fs).ServeHTTP(w, r)
	})
}

func StaticPath(assetPath string) string {
	return "/static/" + strings.TrimPrefix(path.Clean(assetPath), "/")
}
