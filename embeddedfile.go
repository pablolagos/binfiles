package binfiles

import (
	"os"
	"path/filepath"
	"time"
)

// EmbeddedFile implements os.FileInfo interface for a given path and size
type EmbeddedFile struct {
	// Path is the path of this file
	Path string
	// Dir marks of the path is a directory
	Dir bool
	// Len is the length of the fake file, zero if it is a directory
	Len int64
	// Timestamp is the ModTime of this file
	Timestamp time.Time
}

func (f *EmbeddedFile) Name() string {
	_, name := filepath.Split(f.Path)
	return name
}

func (f *EmbeddedFile) Mode() os.FileMode {
	mode := os.FileMode(0644)
	if f.Dir {
		return mode | os.ModeDir
	}
	return mode
}

func (f *EmbeddedFile) ModTime() time.Time {
	return f.Timestamp
}

func (f *EmbeddedFile) Size() int64 {
	return f.Len
}

func (f *EmbeddedFile) IsDir() bool {
	return f.Mode().IsDir()
}

func (f *EmbeddedFile) Sys() interface{} {
	return nil
}
