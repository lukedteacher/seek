package resources

import (
	"net/http"
	"path"
	"strings"
)

// sets directory of static assets (styles, scripts, images, etc.)
const StaticDirectoryPath = "static"

// static asset handler
func Handler() http.Handler {
	// allows direct subdirectory access for static assets
	// example: styles/global.css instead of static/styles/global.css
	files := http.StripPrefix("/static/", http.FileServer(http.Dir(StaticDirectoryPath)))
	// serves the files
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache") // changed from "no-store" to speed up page loading
		files.ServeHTTP(w, r)
	})
}

func StaticPath(assetPath string) string {
	return "/static/" + strings.TrimPrefix(path.Clean(assetPath), "/")
}
