package shared

import (
	"os"
)

// MkdirAller is a func type for os.MkdirAll
type MkdirAller func(path string, perm os.FileMode) error

// WriteFiler is a func type for os.WriteFile
type WriteFiler func(filename string, data []byte, perm os.FileMode) error

// ReadFiler is a func type for os.ReadFile
type ReadFiler func(filename string) ([]byte, error)

// Renamer is a func type for os.Rename
type Renamer func(oldPath, newPath string) error

// ReadDirer is a func type for os.ReadDir
type ReadDirer func(dirname string) ([]os.DirEntry, error)
