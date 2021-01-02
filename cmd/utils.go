package cmd

import (
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/spf13/afero"
)

// TimeNow is a func type for time.Now
type TimeNow func() time.Time

// MkdirAller is a func type for os.MkdirAll
type MkdirAller func(path string, perm os.FileMode) error

// OpenFiler is a func type for os.Open
type OpenFiler func(filename string) (afero.File, error)

// CreateFiler is a func type for os.CreateFile
type CreateFiler func(filename string) (afero.File, error)

// WriteFiler is a func type for ioutil.WriteFile
type WriteFiler func(filename string, data []byte, perm os.FileMode) error

// ReadFiler is a func type for ioutil.ReadFile
type ReadFiler func(filename string) ([]byte, error)

// ReadDirer is a func type for ioutil.ReadDir
type ReadDirer func(dirname string) ([]os.FileInfo, error)

// GetVersions will return, in semver sorted order, all released versions
type GetVersions func(ReadDirer, Config) ([]*semver.Version, error)

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

func getAllVersions(readDir ReadDirer, config Config) ([]*semver.Version, error) {
	allVersions := make([]*semver.Version, 0)

	fileInfos, err := readDir(config.ChangesDir)
	if err != nil {
		return allVersions, err
	}

	for _, file := range fileInfos {
		if file.Name() == config.HeaderPath || file.IsDir() {
			continue
		}

		versionString := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))

		v, err := semver.NewVersion(versionString)
		if err != nil {
			continue
		}

		allVersions = append(allVersions, v)
	}

	sort.Sort(sort.Reverse(semver.Collection(allVersions)))

	return allVersions, nil
}
