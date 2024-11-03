package assets

import (
	"bytes"
	"crypto/md5"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
)

type FingerprintedFS struct {
	fsys         fs.FS
	fingerprints map[string]string
}

func AttachAssets(e *echo.Echo, prefix string, assetDir string, embedFS embed.FS, embedded bool) {
	var inputFS fs.FS

	if embedded {
		inputFS = echo.MustSubFS(embedFS, prefix)
	} else {
		inputFS = os.DirFS(assetDir)
	}

	fingerprintFS, err := NewFFS(inputFS)
	if err != nil {
		log.Fatalf("failed to create fingerprinted FS: %s", err.Error())
	}

	e.StaticFS("/"+prefix, fingerprintFS)

	e.GET("/"+prefix+"/asset-manifest.json", func(c echo.Context) error {
		manifest, err := fingerprintFS.Manifest()
		if err != nil {
			return fmt.Errorf("failed to get asset manifest: %w", err)
		}

		return c.JSON(http.StatusOK, manifest)
	})
}

func NewFFS(fsys fs.FS) (*FingerprintedFS, error) {
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
	return &FingerprintedFS{fsys: fsys, fingerprints: fps}, nil
}

// URL returns the fingerprinted URL for a given asset path
func (ffs *FingerprintedFS) URL(path string) (string, error) {
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

func (ffs *FingerprintedFS) Manifest() (map[string]string, error) {
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

func (ffs *FingerprintedFS) Open(name string) (fs.File, error) {
	return ffs.open(name)
}

func (ffs *FingerprintedFS) open(name string) (fs.File, error) {
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
func (ffs *FingerprintedFS) HttpOpen(name string) (http.File, error) {
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

func (ffs *FingerprintedFS) ReadFile(file string) ([]byte, error) {
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

type fingerprintedFile struct {
	fs.File
	originalName string
	readSeeker   io.ReadSeeker
}

func (ff *fingerprintedFile) Read(p []byte) (n int, err error) {
	return ff.readSeeker.Read(p)
}

func (ff *fingerprintedFile) Seek(offset int64, whence int) (int64, error) {
	return ff.readSeeker.Seek(offset, whence)
}

func (ff *fingerprintedFile) Close() error {
	return ff.File.Close()
}

func (ff *fingerprintedFile) Stat() (fs.FileInfo, error) {
	info, err := ff.File.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}
	return &fingerprintedFileInfo{FileInfo: info, name: filepath.Base(ff.originalName)}, nil
}

type fingerprintedFileInfo struct {
	fs.FileInfo
	name string
}

func (ffi *fingerprintedFileInfo) Name() string {
	return ffi.name
}

func hashFile(fsys fs.FS, path string) (string, error) {
	f, err := fsys.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("failed to hash file: %w", err)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
