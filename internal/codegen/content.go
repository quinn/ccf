package codegen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type ContentType struct {
	Name       string
	PluralName string
	Fields     []*ast.Field
	DirName    string // The actual directory name found
	Config     string // The config string found in the doc
}

type ContentGenerator struct {
	ContentDir string
}

func NewContent(contentDir string) *ContentGenerator {
	return &ContentGenerator{
		ContentDir: contentDir,
	}
}

func (g *ContentGenerator) parseContentTypes(configPath string) ([]ContentType, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, configPath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	f := node

	var types []ContentType
	for _, d := range f.Decls {
		gDecl, ok := d.(*ast.GenDecl)
		if !ok || gDecl.Tok != token.TYPE {
			continue
		}

		for _, sp := range gDecl.Specs {
			tSpec, ok := sp.(*ast.TypeSpec)
			if !ok {
				continue
			}
			tStruct, ok := tSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			// Prefer the spec’s own doc, else the decl’s doc.
			conf := make(map[string]string)
			if gDecl.Doc != nil {
				doc := strings.TrimSpace(gDecl.Doc.Text())
				slog.Debug("Found struct doc", "name", tSpec.Name.Name, "doc", doc)

				strs := strings.Split(doc, ":")
				for i, s := range strs {
					if s == "ccf" {
						if len(strs) < i+1 {
							return nil, fmt.Errorf("incorrect struct doc format: %s", doc)
						}

						confStr := strings.Trim(strs[i+1], "\"")

						slog.Debug("Found struct doc", "name", tSpec.Name.Name, "doc", doc, "conf", confStr)

						for s := range strings.SplitSeq(confStr, " ") {
							kv := strings.Split(s, "=")
							if len(kv) != 2 {
								return nil, fmt.Errorf("incorrect struct doc format: %s", doc)
							}
							conf[kv[0]] = kv[1]
						}
						break
					}
				}
			}

			slog.Debug("Found struct", "name", tSpec.Name.Name, "conf", conf)

			types = append(types, ContentType{
				Name:       tSpec.Name.Name,
				PluralName: tSpec.Name.Name,
				Fields:     tStruct.Fields.List,
				DirName:    conf["dir"],
			})
		}
	}

	// var types []ContentType
	// ast.Inspect(node, func(n ast.Node) bool {
	// 	if typeSpec, ok := n.(*ast.TypeSpec); ok {
	// 		if structType, ok := typeSpec.Type.(*ast.StructType); ok {
	// 			slog.Debug("Found struct type", "structType", structType, "comments", typeSpec.Doc)
	// 			types = append(types, ContentType{
	// 				Name:       typeSpec.Name.Name,
	// 				PluralName: typeSpec.Name.Name,
	// 				Fields:     structType.Fields.List,
	// 			})
	// 		}
	// 	}
	// 	return true
	// })

	return types, nil
}

func (g *ContentGenerator) findMatchingDir(t ContentType, entries []os.DirEntry) (string, bool) {
	singular := strings.ToLower(t.Name)
	plural := singular + "s"

	if t.DirName != "" {
		return t.DirName, strings.HasSuffix(t.DirName, "s")
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dirName := entry.Name()

		if dirName == singular {
			return dirName, false
		}

		if dirName == plural {
			return dirName, true
		}
	}

	return "", true // default to plural if no matching directory found
}

func (g *ContentGenerator) getContentDirs(types []ContentType) ([]ContentType, error) {
	entries, err := os.ReadDir(g.ContentDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read content directory: %w", err)
	}

	// Find matching directories for each type
	for i, t := range types {
		dirName, isPlural := g.findMatchingDir(t, entries)
		types[i].DirName = dirName
		if isPlural {
			types[i].PluralName = t.Name + "s"
		}
	}

	return types, nil
}

var templateFuncs = template.FuncMap{
	"lower": strings.ToLower,
}

func (g *ContentGenerator) Generate() error {
	// Get content types from config
	types, err := g.parseContentTypes(filepath.Join(g.ContentDir, "config.go"))
	if err != nil {
		return fmt.Errorf("failed to parse content types: %w", err)
	}

	// Get content directories and match them to types
	types, err = g.getContentDirs(types)
	if err != nil {
		return fmt.Errorf("failed to get content directories: %w", err)
	}

	slog.Debug("Found content directories", "types", types)

	// Create space-separated list of directories for embed directive
	var dirs []string
	for _, t := range types {
		dirs = append(dirs, t.DirName)
	}

	// Create template data
	data := struct {
		Types []ContentType
		Dirs  string
	}{
		Types: types,
		Dirs:  strings.Join(dirs, " "),
	}

	// Read template file
	tmplContent, err := templates.ReadFile("templates/content.gotmpl")
	if err != nil {
		return fmt.Errorf("failed to read content template: %w", err)
	}

	tmpl, err := template.New("content.gotmpl").Funcs(templateFuncs).Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Set output path relative to content directory
	output := filepath.Join(g.ContentDir, "fs.go")

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create output file
	f, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}
