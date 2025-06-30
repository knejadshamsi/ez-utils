package gui

import (
	"embed"
	"io/fs"
)

//go:embed all:frontend/dist
var content embed.FS

// FrontendAssets is a sub-filesystem pointing to the embedded 'dist' directory.
var FrontendAssets fs.FS

func init() {
	var err error
	FrontendAssets, err = fs.Sub(content, "frontend/dist")
	if err != nil {
		panic(err)
	}
}