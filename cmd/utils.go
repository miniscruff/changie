package cmd

import (
	"github.com/spf13/afero"
	"io"
)

func appendFile(fs afero.Fs, rootFile afero.File, path string) error {
	otherFile, err := fs.Open(path)
	if err != nil {
		return err
	}
	defer otherFile.Close()

	_, err = io.Copy(rootFile, otherFile)
	if err != nil {
		return err
	}

	return nil
}
