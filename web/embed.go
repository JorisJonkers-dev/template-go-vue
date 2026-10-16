// Package web embeds the built single-page app, so the Go binary serves it beside the API.
package web

import (
	"embed"
	"io/fs"
)

// The web build writes dist/; a tracked dist/.gitkeep keeps the pattern matching before it has.
//
//go:embed all:dist
var dist embed.FS

// Dist returns the files of the web build that was present when the binary was compiled.
func Dist() fs.FS {
	sub, _ := fs.Sub(dist, "dist") // "dist" is a valid name, so Sub cannot fail
	return sub
}
