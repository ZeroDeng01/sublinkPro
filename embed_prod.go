//go:build prod
// +build prod

package main

import (
	"embed"
	"io/fs"
)

//go:embed static/assets/*
//go:embed static/images/*
//go:embed static/*
var embeddedFiles embed.FS

var StaticFiles fs.FS = embeddedFiles
