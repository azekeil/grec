// Package commands provides Discord bot command implementations for grec.
package commands

import (
	"embed"
	"github.com/azekeil/grec/internal/goparser"
)

//go:embed *.go
var Files embed.FS

// ParseAll parses all .go files from the embedded commands directory using the modern file-centric approach.
func ParseAll() ([]goparser.ParsedFile, error) {
	return goparser.ParseEmbedFiles(nil, ".", Files, 0)
}
