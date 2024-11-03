package assets

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/fs"
)

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
