package chroot

import (
	"errors"
	"io/fs"
	"log"
	"os"
	"path"

	. "github.com/nobodyawesomer/results"
)

// FileSystem represents a chrooted filesystem.
type FileSystem struct {
	Root      fs.FS
	Logger    *log.Logger
	Verbosity LogVerbosity

	// Underlying file
	// file *os.File
	rootName string

	// TODO: memoize
}

type LogVerbosity int // Could extract all of this out to separate logging utility, or use another pre-built one

const (
	DEFAULT LogVerbosity = iota
	INFO
	DEBUG
)

// New generates a new FileSystem rooted at dirpath in the operating system.
func New(dirpath string) FileSystem {
	err := os.Mkdir(dirpath, 0755)
	if err != nil && !errors.Is(err, os.ErrExist) {
		panic(err)
	}
	root := os.DirFS(dirpath) // TODO come up with a secure method?
	logger := log.New(os.Stdout, "CHROOT", log.LstdFlags /*|log.Lmsgprefix*/)
	return FileSystem{
		Root:      root,
		Logger:    logger,
		Verbosity: DEFAULT,
		rootName:  dirpath,
	}
}

// func (fsys *FileSystem) CreateFiles(filepaths ...string) map[string]*os.File {
// 	return nil // TODO impl
// }

// CreateFile creates a file rooted from fsys at path filepath and
// fills it with the given data.
func (fsys *FileSystem) CreateFile(filepath string, data []byte) *os.File {
	return Unwrap(os.Create(path.Join(fsys.rootName, filepath)))
}

// CreateDirectory creates a directory rooted within fsys. If dirpath
// already exists within fsys, then it succeeds quietly. Will automatically
// create intermediary directories.
func (fsys *FileSystem) CreateDirectory(dirpath string) {
	if err := os.MkdirAll(path.Join(fsys.rootName, dirpath), 0755); err != nil {
		panic(err)
	}
}
