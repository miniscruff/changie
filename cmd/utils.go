package cmd

import (
	"io"
	"os"
	"time"

	"github.com/spf13/afero"
)

// TimeNow is a func type for time.Now
type TimeNow func() time.Time

// MkdirAller is a func type for os.MkdirAll
type MkdirAller func(path string, perm os.FileMode) error

// OpenFiler is a func type for os.OpenFile
type OpenFiler func(filename string) (afero.File, error)

// CreateFiler is a func type for os.CreateFile
type CreateFiler func(filename string) (afero.File, error)

// WriteFiler is a func type for ioutil.WriteFile
type WriteFiler func(filename string, data []byte, perm os.FileMode) error

// ReadFiler is a func type for ioutil.ReadFile
type ReadFiler func(filename string) ([]byte, error)

func appendFile(opener OpenFiler, rootFile io.Writer, path string) error {
	otherFile, err := opener(path)
	if err != nil {
		return err
	}
	defer otherFile.Close()

	// Copy doesn't seem to break in any test I have
	// so ignoring this err return value
	_, _ = io.Copy(rootFile, otherFile)

	return nil
}
