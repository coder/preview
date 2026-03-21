package preview

import (
	"bytes"
	"fmt"
	"io/fs"
	"path"
	"time"
)

var (
	_ fs.FS          = (*overrideFS)(nil)
	_ fs.File        = (*memFile)(nil)
	_ fs.FileInfo    = (*memFileInfo)(nil)
	_ fs.ReadDirFile = (*filteredDir)(nil)
)

// overrideFS wraps a base fs.FS, serving modified content for merged files
// and hiding consumed override files.
type overrideFS struct {
	base     fs.FS
	replaced map[string][]byte // path -> merged content
	hidden   map[string]bool   // paths to exclude
}

func (o *overrideFS) Open(name string) (fs.File, error) {
	if o.hidden[name] {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}

	if content, ok := o.replaced[name]; ok {
		return &memFile{
			name:    path.Base(name), // Stat().Name() returns base name
			content: bytes.NewReader(content),
			size:    int64(len(content)),
		}, nil
	}

	f, err := o.base.Open(name)
	if err != nil {
		return nil, err
	}

	// If this is a directory, wrap it to filter hidden entries.
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, &fs.PathError{Op: "open", Path: name, Err: fmt.Errorf("stat: %w", err)}
	}
	if info.IsDir() {
		if rdf, ok := f.(fs.ReadDirFile); ok {
			return &filteredDir{ReadDirFile: rdf, hidden: o.hidden, dirPath: name}, nil
		}
	}

	return f, nil
}

// memFile is an in-memory file that implements fs.File.
type memFile struct {
	name    string
	content *bytes.Reader
	size    int64
}

func (f *memFile) Stat() (fs.FileInfo, error) {
	return &memFileInfo{name: f.name, size: f.size}, nil
}
func (f *memFile) Read(b []byte) (int, error) { return f.content.Read(b) }
func (f *memFile) Close() error               { return nil }

// memFileInfo implements fs.FileInfo for in-memory files.
type memFileInfo struct {
	name string
	size int64
}

func (fi *memFileInfo) Name() string       { return fi.name }
func (fi *memFileInfo) Size() int64        { return fi.size }
func (fi *memFileInfo) Mode() fs.FileMode  { return 0o444 }
func (fi *memFileInfo) ModTime() time.Time { return time.Time{} }
func (fi *memFileInfo) IsDir() bool        { return false }
func (fi *memFileInfo) Sys() any           { return nil }

// filteredDir wraps a ReadDirFile to filter out hidden entries.
type filteredDir struct {
	fs.ReadDirFile
	hidden  map[string]bool
	dirPath string
}

func (d *filteredDir) ReadDir(n int) ([]fs.DirEntry, error) {
	entries, err := d.ReadDirFile.ReadDir(n)
	filtered := d.filter(entries)

	// If n > 0, an empty slice must have a non-nil error per the
	// fs.ReadDirFile contract. Keep reading until we have results
	// or the underlying reader signals EOF/error.
	// (len(entries) > 0 is defensive programming against underlying
	// FSs that break the contract.)
	for n > 0 && len(filtered) == 0 && err == nil && len(entries) > 0 {
		entries, err = d.ReadDirFile.ReadDir(n)
		filtered = d.filter(entries)
	}

	return filtered, err
}

func (d *filteredDir) filter(entries []fs.DirEntry) []fs.DirEntry {
	filtered := make([]fs.DirEntry, 0, len(entries))
	for _, entry := range entries {
		if !d.hidden[path.Join(d.dirPath, entry.Name())] {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}
