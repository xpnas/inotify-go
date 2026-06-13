package main

import (
	"embed"
	"io/fs"
)

// uiDist holds the embedded frontend build output.
// When building the production image, the Dockerfile copies the frontend dist
// into cmd/inotify/ui/dist before running go build.
// In local development the directory exists but may only contain the placeholder,
// so LoadUI returns nil and the server skips static file serving.

//go:embed all:ui/dist
var uiDist embed.FS

// LoadUI returns the embedded frontend FS, or nil if not built yet.
func LoadUI() fs.FS {
	sub, err := fs.Sub(uiDist, "ui/dist")
	if err != nil {
		return nil
	}
	// Check that a real index.html was embedded (not just the placeholder)
	f, err := sub.Open("index.html")
	if err != nil {
		return nil
	}
	stat, _ := f.Stat()
	f.Close()
	if stat.Size() < 100 {
		return nil // placeholder only
	}
	return sub
}
