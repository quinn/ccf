package assets

import (
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
)

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
