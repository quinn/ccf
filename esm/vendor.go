package esm

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

	"github.com/evanw/esbuild/pkg/api"
)

// Module represents a single ESM module to vendor
type Module struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

// Template holds configuration for HTML template processing
type Template struct {
	Input  string `yaml:"input"`
	Output string `yaml:"output"`
}

// Config holds the configuration for vendoring ESM modules
type Config struct {
	OutputDir  string    `yaml:"output"`
	ImportPath string    `yaml:"import"`
	Modules    []Module  `yaml:"modules"`
	Template   *Template `yaml:"template,omitempty"`
}

func Run(config Config) error {
	if config.OutputDir == "" {
		return fmt.Errorf("missing field `output`: output directory is required")
	}
	if config.ImportPath == "" {
		return fmt.Errorf("missing field `import`: import path is required")
	}

	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		return err
	}

	// Create a shared cache for all dependencies
	cache := &moduleCache{
		outputDir:   config.OutputDir,
		importPath:  config.ImportPath,
		entryURLs:   make(map[string]string),
		modules:     make(map[string]string),
		contents:    make(map[string]string),
		importSpecs: make(map[string]string),
	}

	// Process each dependency
	for _, dep := range config.Modules {
		fmt.Fprintf(os.Stderr, "Processing %s from %s...\n", dep.Name, dep.URL)
		if err := processDependency(dep, cache); err != nil {
			return fmt.Errorf("failed to process %s: %w", dep.Name, err)
		}
	}

	// Write all downloaded modules to disk
	cache.mu.Lock()
	contentsCopy := make(map[string]string, len(cache.contents))
	for k, v := range cache.contents {
		contentsCopy[k] = v
	}
	cache.mu.Unlock()

	for moduleURL, content := range contentsCopy {
		cache.mu.Lock()
		localPath := cache.modules[moduleURL]
		cache.mu.Unlock()

		fullPath := filepath.Join(config.OutputDir, localPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return err
		}
	}

	// Generate import map once
	importMap := cache.generateImportMap()

	if err := cache.writeImportMap(importMap); err != nil {
		return err
	}

	// Process template if configured
	if config.Template != nil && config.Template.Input != "" && config.Template.Output != "" {
		if err := processTemplate(config, importMap); err != nil {
			return fmt.Errorf("failed to process template: %w", err)
		}
	}

	return nil
}

func processDependency(dep Module, cache *moduleCache) error {
	// Store the entry URL for this dependency
	cache.mu.Lock()
	cache.entryURLs[dep.Name] = dep.URL
	cache.mu.Unlock()

	result := api.Build(api.BuildOptions{
		EntryPoints: []string{dep.URL},
		Bundle:      true,
		Write:       false,
		Format:      api.FormatESModule,
		Platform:    api.PlatformBrowser,
		Plugins: []api.Plugin{
			{
				Name: "http-loader",
				Setup: func(build api.PluginBuild) {
					build.OnResolve(api.OnResolveOptions{Filter: ".*"},
						func(args api.OnResolveArgs) (api.OnResolveResult, error) {
							if strings.HasPrefix(args.Path, "http://") || strings.HasPrefix(args.Path, "https://") {
								cache.addImportSpec(args.Path, args.Path)
								return api.OnResolveResult{
									Path:      args.Path,
									Namespace: "http",
								}, nil
							}

							if args.Importer != "" {
								base, err := url.Parse(args.Importer)
								if err == nil && (base.Scheme == "http" || base.Scheme == "https") {
									resolved := resolveURL(base, args.Path)
									if !strings.HasPrefix(args.Path, ".") {
										cache.addImportSpec(args.Path, resolved)
									}
									return api.OnResolveResult{
										Path:      resolved,
										Namespace: "http",
									}, nil
								}
							}

							return api.OnResolveResult{}, nil
						})

					build.OnLoad(api.OnLoadOptions{Filter: ".*", Namespace: "http"},
						func(args api.OnLoadArgs) (api.OnLoadResult, error) {
							cache.mu.Lock()
							cached, ok := cache.contents[args.Path]
							cache.mu.Unlock()

							if ok {
								return api.OnLoadResult{
									Contents: &cached,
									Loader:   api.LoaderJS,
								}, nil
							}

							content, err := downloadURL(args.Path)
							if err != nil {
								return api.OnLoadResult{}, err
							}

							cache.addModule(args.Path, content)

							return api.OnLoadResult{
								Contents: &content,
								Loader:   api.LoaderJS,
							}, nil
						})
				},
			},
		},
	})

	if len(result.Errors) > 0 {
		for _, err := range result.Errors {
			fmt.Fprintf(os.Stderr, "%s\n", err.Text)
		}
		return fmt.Errorf("build failed with %d errors", len(result.Errors))
	}

	return nil
}

type moduleCache struct {
	outputDir   string
	importPath  string
	entryURLs   map[string]string // maps import name to entry URL
	modules     map[string]string
	contents    map[string]string
	importSpecs map[string]string
	mu          sync.Mutex
}

func (c *moduleCache) addModule(moduleURL, content string) string {
	c.mu.Lock()
	defer c.mu.Unlock()

	if localPath, exists := c.modules[moduleURL]; exists {
		return localPath
	}

	localPath := urlToPath(moduleURL)
	c.modules[moduleURL] = localPath
	c.contents[moduleURL] = content
	return localPath
}

func (c *moduleCache) addImportSpec(spec, fullURL string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.importSpecs[spec] = fullURL
}

func (c *moduleCache) generateImportMap() map[string]map[string]string {
	c.mu.Lock()
	defer c.mu.Unlock()

	importMapEntries := make(map[string]string)

	// Add all import specs (transitive dependencies)
	for spec, fullURL := range c.importSpecs {
		if localPath, exists := c.modules[fullURL]; exists {
			res := c.importPath + "/" + localPath
			importMapEntries[spec] = res
		}
	}

	// Add named entry points for each dependency
	for importName, entryURL := range c.entryURLs {
		if localPath, exists := c.modules[entryURL]; exists {
			res := c.importPath + "/" + localPath
			importMapEntries[importName] = res
		}
	}

	return map[string]map[string]string{
		"imports": importMapEntries,
	}
}

func (c *moduleCache) writeImportMap(importMap map[string]map[string]string) error {
	data, err := json.MarshalIndent(importMap, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(c.outputDir, "importmap.json"), data, 0644)
}

func urlToPath(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		hash := sha256.Sum256([]byte(rawURL))
		return "invalid-" + hex.EncodeToString(hash[:8]) + ".js"
	}

	path := filepath.Join(u.Host, u.Path)

	if !strings.HasSuffix(path, ".js") && !strings.HasSuffix(path, ".mjs") {
		path = path + ".js"
	}

	return path
}

func downloadURL(rawURL string) (string, error) {
	resp, err := http.Get(rawURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, rawURL)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func resolveURL(base *url.URL, ref string) string {
	if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") {
		return ref
	}
	refURL, err := url.Parse(ref)
	if err != nil {
		return ref
	}
	return base.ResolveReference(refURL).String()
}

func processTemplate(config Config, importMap map[string]map[string]string) error {
	// Read template file
	tmplContent, err := os.ReadFile(config.Template.Input)
	if err != nil {
		return fmt.Errorf("failed to read template file: %w", err)
	}

	// Parse template
	tmpl, err := template.New("index").Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Generate import map JSON with proper indentation for HTML
	importMapJSON, err := json.MarshalIndent(importMap, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal import map: %w", err)
	}

	// Execute template
	data := map[string]string{
		"importmap": string(importMapJSON),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(config.Template.Output)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write output file
	if err := os.WriteFile(config.Template.Output, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	slog.Debug("Generated from template", "output", config.Template.Output)

	return nil
}
