package shared

import (
	"os"
	"time"
)

// TimeNow is a func type for time.Now
type TimeNow func() time.Time

// MkdirAller is a func type for os.MkdirAll
type MkdirAller func(path string, perm os.FileMode) error

// OpenFiler is a func type for os.Open
type OpenFiler func(filename string) (*os.File, error)

// CreateFilerOS is a func type for os.CreateFile
type CreateFiler func(filename string) (*os.File, error)

// WriteFiler is a func type for os.WriteFile
type WriteFiler func(filename string, data []byte, perm os.FileMode) error

// FileExister checks whether a file exists or not
type FileExister func(filename string) (bool, error)

// ReadFiler is a func type for os.ReadFile
type ReadFiler func(filename string) ([]byte, error)

// ReadDirer is a func type for os.ReadDir
type ReadDirer func(dirname string) ([]os.DirEntry, error)
