package then

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// WithTempDir creates a temporary directory and moves our working directory to it.
// At the end of our test it will remove the directory and revert our working directory
// to the original value.
func WithTempDir(t *testing.T) {
	t.Helper()

	startDir, err := os.Getwd()
	Nil(t, err)

	tempDir := t.TempDir()

	err = os.Chdir(tempDir)
	Nil(t, err)

	t.Cleanup(func() {
		err := os.Chdir(startDir)
		Nil(t, err)
	})
}

// FileExists checks whether a file exists at the path defined.
// Paths will be joined together using `filepath.Join`.
func FileExists(t *testing.T, paths ...string) {
	t.Helper()

	fullPath := filepath.Join(paths...)

	_, err := os.Stat(fullPath)
	if err == nil {
		return
	}

	switch {
	case os.IsNotExist(err):
		t.Errorf("expected file '%v' to exist", fullPath)
	default:
		t.Errorf("error getting info for '%v': %v", fullPath, err)
	}
}

// FileContentsNoAfero will check the contents of a file.
func FileContentsNoAfero(t *testing.T, contents string, paths ...string) {
	t.Helper()

	fullPath := filepath.Join(paths...)

	bs, err := os.ReadFile(fullPath)
	if err != nil {
		t.Errorf("reading file: '%v'", fullPath)
	}

	expected := string(bs)
	Equals(t, expected, contents)
}

// MockFileInfo is a simple struct to fake the `os.FileInfo`
type MockFileInfo struct {
	MockName  string
	MockIsDir bool
}

var _ os.FileInfo = (*MockFileInfo)(nil)

func (m *MockFileInfo) Name() string {
	return m.MockName
}

func (m *MockFileInfo) IsDir() bool {
	return m.MockIsDir
}

func (m *MockFileInfo) Size() int64 {
	return 0
}

func (m *MockFileInfo) Mode() os.FileMode {
	return os.ModeAppend
}

func (m *MockFileInfo) ModTime() time.Time {
	return time.Now()
}

func (m *MockFileInfo) Sys() any {
	return nil
}
