package fonts

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"go.quinn.io/ccf/fonts"
	"gopkg.in/yaml.v3"
)

// FontsYAML represents the schema of fonts.yaml
// Example:
// fonts:
//   - family: "Roboto"
//     variants: ["regular", "700italic"]
//
// dir: "./webfonts"
// stylesheet: "./fonts.css"
type FontsYAML struct {
	Fonts      []FontEntry `yaml:"fonts"`
	Dir        string      `yaml:"dir"`
	Import     string      `yaml:"import"`
	Stylesheet string      `yaml:"stylesheet"`
}

type FontEntry struct {
	Family   string   `yaml:"family"`
	Variants []string `yaml:"variants"`
}

func Main() {
	configPathPtr := flag.String("config", "fonts.yaml", "Path to font configuration file")
	gfontsKeyPtr := flag.String("gfonts-key", os.Getenv("GFONTS_KEY"), "Google Fonts API key. (default \"GFONTS_KEY\" env var)")
	debugPtr := flag.Bool("debug", os.Getenv("DEBUG") == "true", "Enable debug logging")

	flag.Parse()

	if *debugPtr {
		h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
		slog.SetDefault(slog.New(h))
	}

	configPath := *configPathPtr
	gfontsKey := *gfontsKeyPtr

	if gfontsKey == "" {
		slog.Error("GFONTS_KEY env var or -gfonts-key flag not set")
		os.Exit(1)
	}

	slog.Debug("Reading font configuration", "path", configPath)

	cfg, err := readFontsYAML(configPath)
	if err != nil {
		slog.Error("Error reading YAML", "error", err)
		os.Exit(1)
	}

	slog.Debug("config", "dir", cfg.Dir, "stylesheet", cfg.Stylesheet, "import", cfg.Import)
	slog.Debug("Installing fonts to directory", "dir", cfg.Dir)

	if cfg.Dir == "" {
		slog.Error("Error: `dir` not specified in YAML")
		os.Exit(1)
	}

	if err := os.MkdirAll(cfg.Dir, 0755); err != nil {
		slog.Error("Failed to create directory", "dir", cfg.Dir, "error", err)
		os.Exit(1)
	}

	if cfg.Stylesheet == "" {
		slog.Error("Error: `stylesheet` not specified in YAML")
		os.Exit(1)
	}

	if err := os.MkdirAll(filepath.Dir(cfg.Stylesheet), 0755); err != nil {
		slog.Error("Failed to create directory", "dir", cfg.Stylesheet, "error", err)
		os.Exit(1)
	}

	// Track all font files that should exist after install
	wantedFiles := map[string]struct{}{}
	cssRules := []string{}
	if len(cfg.Fonts) == 0 {
		slog.Error("No fonts specified in YAML")
		os.Exit(1)
	}

	for _, entry := range cfg.Fonts {
		parsedFamily := fonts.ParseFontFamily(entry.Family)
		fontResponse := fonts.GetFontUrl(gfontsKey, parsedFamily)
		if len(fontResponse.Items) < 1 {
			slog.Warn("No font found for", "family", entry.Family)
			continue
		}

		item := fontResponse.Items[0]
		files := item.Files
		for _, variant := range entry.Variants {
			url, ok := files[variant]
			if !ok {
				slog.Warn("Variant not found for", "family", entry.Family, "variant", variant)
				slog.Warn("Available variants:", "variants", item.Variants)
				os.Exit(1)
			}
			fileName := item.Family + "_" + variant + ".woff2"
			filePath := filepath.Join(cfg.Dir, fileName)
			slog.Debug("Downloading", "family", entry.Family, "variant", variant, "path", filePath)
			if err := downloadToFile(url, filePath); err != nil {
				slog.Error("Failed to download", "file", fileName, "error", err)
				os.Exit(1)
			}
			wantedFiles[fileName] = struct{}{}

			cssImportFilename := fileName
			if cfg.Import != "" {
				cssImportFilename = path.Join(cfg.Import, fileName)
			}

			cssRules = append(cssRules, genCSS(item.Axes, item.Family, variant, cssImportFilename))
		}
	}

	// Remove any font files in dir not referenced in wantedFiles
	removeUnreferencedFiles(cfg.Dir, wantedFiles)

	// Write CSS file
	slog.Debug("Writing CSS to", "path", cfg.Stylesheet)

	if err := writeCSS(cfg.Stylesheet, cssRules); err != nil {
		slog.Error("Failed to write CSS", "path", cfg.Stylesheet, "error", err)
		os.Exit(1)
	}

	slog.Info("Install complete!")
}

func readFontsYAML(path string) (*FontsYAML, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var cfg FontsYAML
	dec := yaml.NewDecoder(f)
	if err := dec.Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func downloadToFile(url, filePath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("bad status: %s", resp.Status)
	}
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

func removeUnreferencedFiles(dir string, wanted map[string]struct{}) {
	d, err := os.Open(dir)
	if err != nil {
		slog.Error("Failed to open directory for cleanup", "dir", dir, "error", err)
		return
	}
	defer d.Close()
	files, err := d.Readdirnames(-1)
	if err != nil {
		slog.Error("Failed to list directory", "dir", dir, "error", err)
		return
	}
	for _, f := range files {
		if !strings.HasSuffix(f, ".woff2") {
			continue
		}
		if _, ok := wanted[f]; !ok {
			fullPath := filepath.Join(dir, f)
			slog.Debug("Removing unreferenced font file", "path", fullPath)
			os.Remove(fullPath)
		}
	}
}

func writeCSS(path string, rules []string) error {
	css := strings.Join(rules, "\n\n")
	return os.WriteFile(path, []byte(css), 0644)
}

func genCSS(axes []*fonts.Axes, family, variant, fileName string) string {
	format := "woff2"
	style := "normal"
	weight := "400"
	if variant == "italic" {
		style = "italic"
	} else if strings.HasSuffix(variant, "italic") {
		style = "italic"
		weight = strings.TrimSuffix(variant, "italic")
	} else if variant != "regular" {
		weight = variant
	}

	for _, axis := range axes {
		if axis.Tag == "wght" {
			weight = fmt.Sprintf("%d %d", axis.Start, axis.End)
			format = "woff2-variations"
			continue
		}
		slog.Warn("Unsupported axis", "tag", axis.Tag, "axis", axis)
	}

	return fmt.Sprintf(`@font-face {
  font-family: '%s';
  font-style: %s;
  font-weight: %s;
  src: url('%s') format('%s');
}`, family, style, weight, fileName, format)
}
