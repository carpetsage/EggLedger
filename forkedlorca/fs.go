package forkedlorca

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// Embed is a helper function that embeds assets from the given directories
// into a Go source file. It is designed to be called from some generator
// script, see example project to find out how it can be used.
func Embed(packageName, file string, dirs ...string) error {
	w, err := os.Create(file)
	if err != nil {
		return err
	}
	defer w.Close()
	fmt.Fprintf(w, `// Code generated by Lorca. DO NOT EDIT.
package %s

import (
	"bytes"
	"errors"
	"net/http"
	"os"
	"time"
)

var assets = map[string][]byte{}

var FS = &fs{}

type fs struct {}

func (fs *fs) Open(name string) (http.File, error) {
	if name == "/" {
		return fs, nil;
	}
	b, ok := assets[name]
	if !ok {
		return nil, os.ErrNotExist
	}
	return &file{name: name, size: len(b), Reader: bytes.NewReader(b)}, nil
}

func (fs *fs) Close() error { return nil }
func (fs *fs) Read(p []byte) (int, error) { return 0, nil }
func (fs *fs) Seek(offset int64, whence int) (int64, error) { return 0, nil }
func (fs *fs) Stat() (os.FileInfo, error) { return fs, nil }
func (fs *fs) Name() string { return "/" }
func (fs *fs) Size() int64 { return 0 }
func (fs *fs) Mode() os.FileMode { return 0755}
func (fs *fs) ModTime() time.Time{ return time.Time{} }
func (fs *fs) IsDir() bool { return true }
func (fs *fs) Sys() interface{} { return nil }
func (fs *fs) Readdir(count int) ([]os.FileInfo, error) {
	files := []os.FileInfo{}
	for name, data := range assets {
		files = append(files, &file{name: name, size: len(data), Reader: bytes.NewReader(data)})
	}
	return files, nil
}

type file struct {
	name string
	size int
	*bytes.Reader 
}

func (f *file) Close() error { return nil }
func (f *file) Readdir(count int) ([]os.FileInfo, error) { return nil, errors.New("not supported") }
func (f *file) Stat() (os.FileInfo, error) { return f, nil }
func (f *file) Name() string { return f.name }
func (f *file) Size() int64 { return int64(f.size) }
func (f *file) Mode() os.FileMode { return 0644 }
func (f *file) ModTime() time.Time{ return time.Time{} }
func (f *file) IsDir() bool { return false }
func (f *file) Sys() interface{} { return nil }

func init() {
`, packageName)
	defer fmt.Fprintln(w, `}`)

	for _, dir := range dirs {
		filepath.Walk(dir, func(path string, info os.FileInfo, _ error) error {
			if info.IsDir() {
				return nil
			}
			b, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			path = filepath.ToSlash(path)
			fmt.Fprintf(w, `	assets[%q] = []byte{`, strings.TrimPrefix(path, dir))
			for i := 0; i < len(b); i++ {
				if i > 0 {
					fmt.Fprintf(w, `, `)
				}
				fmt.Fprintf(w, `0x%02x`, b[i])
			}
			fmt.Fprintln(w, `}`)
			return nil
		})
	}
	return nil
}
