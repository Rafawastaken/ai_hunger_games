package static

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed all:*
var staticFiles embed.FS

// GetFileSystem returns an http.FileSystem for the embedded static files.
// The files are expected to be in the root of the embed (from Vite build output).
func GetFileSystem() http.FileSystem {
	// The embed includes everything in the static folder
	fsys, err := fs.Sub(staticFiles, ".")
	if err != nil {
		panic(err)
	}
	return http.FS(fsys)
}

// GetFS returns the raw embed.FS for advanced usage
func GetFS() embed.FS {
	return staticFiles
}
