// Package goparser provides Go source parsing utilities for embedded filesystems.
// It offers modern, file-centric alternatives to the deprecated ast.Package approach.
package goparser

import (
	"embed"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
)

// ParsedFile represents a successfully parsed Go source file from an embedded filesystem.
type ParsedFile struct {
	Name   string
	Path   string
	File   *ast.File
	Errors []error // Parse errors for this specific file, if any
}

// ParseEmbedFiles parses all .go files in the specified directory from embedFS.
// It returns a slice of parsed files, collecting parse errors rather than failing immediately.
// This is the modern replacement for ast.Package-based approaches.
func ParseEmbedFiles(fset *token.FileSet, rootPath string, embedFS embed.FS, mode parser.Mode) ([]ParsedFile, error) {
	var results []ParsedFile

	err := fs.WalkDir(embedFS, rootPath, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		// Skip directories and non-.go files
		if d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Read file content from embed.FS
		data, readErr := embedFS.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("failed to read %s: %w", path, readErr)
		}

		// Parse the file
		file, parseErr := parser.ParseFile(fset, path, data, mode)

		result := ParsedFile{
			Name: filepath.Base(path),
			Path: path,
		}

		if parseErr != nil {
			result.Errors = append(result.Errors, parseErr)
			results = append(results, result) // Include file even with errors for visibility
			return nil
		}

		result.File = file
		results = append(results, result)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return results, nil
}

// ParseEmbedFilesStrict is like ParseEmbedFiles but returns an error if any files fail to parse.
func ParseEmbedFilesStrict(fset *token.FileSet, rootPath string, embedFS embed.FS, mode parser.Mode) ([]*ast.File, error) {
	results, err := ParseEmbedFiles(fset, rootPath, embedFS, mode)
	if err != nil {
		return nil, err
	}

	var parseErrors []error
	for _, result := range results {
		if len(result.Errors) > 0 {
			parseErrors = append(parseErrors, fmt.Errorf("%s: %w", result.Path, errors.Join(result.Errors...)))
		}
	}

	if len(parseErrors) > 0 {
		return nil, fmt.Errorf("parse errors encountered: %w", errors.Join(parseErrors...))
	}

	files := make([]*ast.File, len(results))
	for i, result := range results {
		files[i] = result.File
	}

	return files, nil
}

// GroupByPackage groups parsed files by their package name.
// This replaces the deprecated ast.Package grouping functionality.
func GroupByPackage(files []ParsedFile) map[string][]*ast.File {
	groups := make(map[string][]*ast.File)

	for _, result := range files {
		if result.File == nil {
			continue // Skip files that failed to parse
		}

		pkgName := result.File.Name.Name
		groups[pkgName] = append(groups[pkgName], result.File)
	}

	return groups
}

// GetPackageFiles retrieves all parsed files for a specific package name.
func GetPackageFiles(files []ParsedFile, pkgName string) []ParsedFile {
	var results []ParsedFile

	for _, result := range files {
		if result.File != nil && result.File.Name.Name == pkgName {
			results = append(results, result)
		}
	}

	return results
}

// HasParseErrors checks if any of the parsed files had parse errors.
func HasParseErrors(files []ParsedFile) bool {
	for _, f := range files {
		if len(f.Errors) > 0 {
			return true
		}
	}
	return false
}
