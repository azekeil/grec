package self

import (
	"go/ast"
	"go/doc"
	"sort"
	"strings"

	"github.com/azekeil/grec/internal/goparser"
)

// DocFuncs represents documentation functions for commands.
type DocFuncs map[string]*doc.Func

func (f DocFuncs) Summary(name string) string {
	return strings.Split(f[name].Doc, "\n")[0]
}

func (f DocFuncs) CommandHelp(name string) string {
	return f[name].Doc
}

func (f DocFuncs) AllSummaries() []string {
	var s []string
	for k := range f {
		s = append(s, f.Summary(k))
	}
	sort.Strings(s)
	return s
}

// Capitalise does a case-insensitive comparison on the function names
// and returns the correctly-capitalised name if present
func (f DocFuncs) Capitalise(name string) string {
	lname := strings.ToLower(name)
	for _, k := range f {
		if lname == strings.ToLower(k.Name) {
			return k.Name
		}
	}
	return ""
}

// getDocPackage creates a doc.Package from parsed files for use with go/doc.
//
// NOTE: We must construct an ast.Package here because go/doc.New() requires it,
// even though the type is deprecated in Go 1.24+. This is a temporary workaround
// required by library compatibility until go/doc supports file-centric input directly.
// See: https://github.com/golang/go/issues/67895
func getDocPackage(files []goparser.ParsedFile, name string) *doc.Package {
	var filesToParse []*ast.File
	for _, f := range files {
		if f.File != nil && f.File.Name.Name == name {
			filesToParse = append(filesToParse, f.File)
		}
	}

	if len(filesToParse) == 0 {
		return nil
	}

	// TODO: Replace with file-centric approach when go/doc supports it directly
	pkg := &ast.Package{
		Name:  name,
		Files: make(map[string]*ast.File),
	}
	for _, f := range filesToParse {
		pkg.Files[f.Name.Name] = f
	}

	return doc.New(pkg, "./", 0)
}

func getDocMethods(p *doc.Package, typeName string) DocFuncs {
	s := make(map[string]*doc.Func, len(p.Types))
	for _, t := range p.Types {
		if t.Name == typeName {
			for _, m := range t.Methods {
				s[m.Name] = m
			}
		}
	}
	return DocFuncs(s)
}

// BuildDocFromFiles creates DocFuncs from pre-parsed files, avoiding circular imports.
func BuildDocFromFiles(files []goparser.ParsedFile) DocFuncs {
	pkg := getDocPackage(files, "commands")
	if pkg == nil {
		return make(DocFuncs)
	}

	docFuncs := getDocMethods(pkg, "")

	// Also add methods from Command type
	for _, t := range pkg.Types {
		if t.Name == "Command" {
			for _, m := range t.Methods {
				docFuncs[m.Name] = m
			}
		}
	}

	return docFuncs
}
