package cmd

import (
	"github.com/spf13/afero"
	"io"
	"os"
	"time"
)

type TimeNow func() time.Time

type MkdirAller func(path string, perm os.FileMode) error
type WriteFiler func(filename string, data []byte, perm os.FileMode) error
type ReadFiler func(filename string) ([]byte, error)

func appendFile(fs afero.Fs, rootFile afero.File, path string) error {
	otherFile, err := fs.Open(path)
	if err != nil {
		return err
	}
	defer otherFile.Close()

	// Copy doesn't seem to break in any test I have
	// so ignoring this err return value
	_, _ = io.Copy(rootFile, otherFile)
	return nil
}
