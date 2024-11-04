package assets

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
)

func newFFS(fsys fs.FS) (*fingerprintedFS, error) {
	fps := make(map[string]string)
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("failed to walk filesystem: %w", err)
		}
		if !d.IsDir() {
			hash, err := hashFile(fsys, path)
			if err != nil {
				return fmt.Errorf("failed to hash file: %w", err)
			}
			fps[path] = hash
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk filesystem: %w", err)
	}
	return &fingerprintedFS{fsys: fsys, fingerprints: fps}, nil
}

type fingerprintedFS struct {
	fsys         fs.FS
	fingerprints map[string]string
}

// URL returns the fingerprinted URL for a given asset path
func (ffs *fingerprintedFS) URL(path string) (string, error) {
	hash, ok := ffs.fingerprints[path]
	if !ok {
		return "", fmt.Errorf("file not found: %s", path)
	}

	ext := filepath.Ext(path)
	fingerprinted := fmt.Sprintf("%s.%s%s", strings.TrimSuffix(path, ext), hash, ext)
	// This seems to do nothing?
	// fingerprinted = strings.Replace(fingerprinted, "public/", spec.AssetPrefix+"/", 1)
	return "/" + fingerprinted, nil
}

func (ffs *fingerprintedFS) Manifest() (map[string]string, error) {
	manifest := make(map[string]string)

	for path := range ffs.fingerprints {
		fURL, err := ffs.URL(path)
		if err != nil {
			return nil, fmt.Errorf("failed to get fingerprinted URL: %w", err)
		}

		manifest[path] = fURL
	}

	return manifest, nil
}

func (ffs *fingerprintedFS) Open(name string) (fs.File, error) {
	return ffs.open(name)
}

func (ffs *fingerprintedFS) open(name string) (fs.File, error) {
	// Strip the fingerprint from the filename if present
	originalName := name
	parts := strings.Split(name, ".")
	if len(parts) > 2 {
		possibleHash := parts[len(parts)-2]
		if len(possibleHash) == 32 { // MD5 hash length
			name = strings.Join(append(parts[:len(parts)-2], parts[len(parts)-1]), ".")
		}
	}

	file, err := ffs.fsys.Open(name)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	// Read the entire file content
	content, err := io.ReadAll(file)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	// Create a ReadSeeker from the content
	readSeeker := bytes.NewReader(content)

	return &fingerprintedFile{
		File:         file,
		originalName: originalName,
		readSeeker:   readSeeker,
	}, nil
}

// Implement http.FileSystem interface
func (ffs *fingerprintedFS) HttpOpen(name string) (http.File, error) {
	f, err := ffs.open(name)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	hf, ok := f.(http.File)
	if !ok {
		return nil, fmt.Errorf("file does not implement http.File")
	}

	return hf, nil
}

func (ffs *fingerprintedFS) ReadFile(file string) ([]byte, error) {
	efs, ok := ffs.fsys.(embed.FS)
	if ok {
		return efs.ReadFile(file)
	}

	f, err := ffs.Open(file)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	return io.ReadAll(f)
}
