package binfiles

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var (
	defaultFileTimestamp = time.Now()
)

// HttpFile implements http.File interface for a no-directory file with content
type HttpFile struct {
	*bytes.Reader
	io.Closer
	EmbeddedFile
}

func NewHttpFile(name string, content []byte, timestamp time.Time) *HttpFile {
	if timestamp.IsZero() {
		timestamp = defaultFileTimestamp
	}
	return &HttpFile{
		bytes.NewReader(content),
		ioutil.NopCloser(nil),
		EmbeddedFile{name, false, int64(len(content)), timestamp}}
}

func (f *HttpFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, errors.New("not a directory")
}

func (f *HttpFile) Size() int64 {
	return f.EmbeddedFile.Size()
}

func (f *HttpFile) Stat() (os.FileInfo, error) {
	return f, nil
}

// AssetDirectory implements http.File interface for a directory
type AssetDirectory struct {
	HttpFile
	ChildrenRead int
	Children     []os.FileInfo
}

func NewAssetDirectory(name string, children []string, fs *AssetFS) *AssetDirectory {
	fileinfos := make([]os.FileInfo, 0, len(children))
	for _, child := range children {
		_, err := fs.AssetDir(filepath.Join(name, child))
		fileinfos = append(fileinfos, &EmbeddedFile{child, err == nil, 0, time.Time{}})
	}
	return &AssetDirectory{
		HttpFile{
			bytes.NewReader(nil),
			ioutil.NopCloser(nil),
			EmbeddedFile{name, true, 0, time.Time{}},
		},
		0,
		fileinfos}
}

func (f *AssetDirectory) Readdir(count int) ([]os.FileInfo, error) {
	if count <= 0 {
		return f.Children, nil
	}
	if f.ChildrenRead+count > len(f.Children) {
		count = len(f.Children) - f.ChildrenRead
	}
	rv := f.Children[f.ChildrenRead : f.ChildrenRead+count]
	f.ChildrenRead += count
	return rv, nil
}

func (f *AssetDirectory) Stat() (os.FileInfo, error) {
	return f, nil
}

// AssetFS implements http.FileSystem, allowing
// embedded files to be served from net/http package.
type AssetFS struct {
	// Asset should return content of file in path if exists
	Asset func(path string) ([]byte, error)
	// AssetDir should return list of files in the path
	AssetDir func(path string) ([]string, error)
	// AssetInfo should return the info of file in path if exists
	AssetInfo func(path string) (os.FileInfo, error)
	// Prefix would be prepended to http requests
	Prefix string
	// Fallback file that is served if no other is found
	Fallback string
}

func (fs *AssetFS) Open(name string) (http.File, error) {
	name = path.Join(fs.Prefix, name)
	if len(name) > 0 && name[0] == '/' {
		name = name[1:]
	}
	if b, err := fs.Asset(name); err == nil {
		timestamp := defaultFileTimestamp
		if fs.AssetInfo != nil {
			if info, err := fs.AssetInfo(name); err == nil {
				timestamp = info.ModTime()
			}
		}
		return NewHttpFile(name, b, timestamp), nil
	}
	children, err := fs.AssetDir(name)

	if err != nil {
		if len(fs.Fallback) > 0 {
			return fs.Open(fs.Fallback)
		}

		// If the error is not found, return an error that will
		// result in a 404 error. Otherwise the server returns
		// a 500 error for files not found.
		if strings.Contains(err.Error(), "not found") {
			return nil, os.ErrNotExist
		}
		return nil, err
	}

	return NewAssetDirectory(name, children, fs), nil
}
