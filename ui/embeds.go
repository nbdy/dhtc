package ui

import "embed"

//go:embed templates/**/*.html
var templates embed.FS

//go:embed static
var static embed.FS
