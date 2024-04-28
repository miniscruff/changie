package shared

import (
	"os"
)

// WriteFiler is a func type for os.WriteFile
type WriteFiler func(filename string, data []byte, perm os.FileMode) error

// ReadFiler is a func type for os.ReadFile
type ReadFiler func(filename string) ([]byte, error)
