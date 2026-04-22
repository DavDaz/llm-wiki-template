// Package templates provides access to embedded wiki template assets.
package templates

import (
	"embed"
	"io/fs"
)

//go:embed assets/* assets/commands/*
var assetsFS embed.FS

// FS returns a sub-FS rooted at "assets", hiding the leading path segment.
func FS() fs.FS {
	sub, err := fs.Sub(assetsFS, "assets")
	if err != nil {
		// embed.FS.Sub only errors if the path doesn't exist, which would be a
		// compile-time mistake — panic is appropriate here.
		panic("templates: assets sub-FS not found: " + err.Error())
	}
	return sub
}

// ReadFile reads a named file from the embedded assets (relative to assets/).
func ReadFile(name string) ([]byte, error) {
	return assetsFS.ReadFile("assets/" + name)
}
