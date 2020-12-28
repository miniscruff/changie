package cmd

import (
	"fmt"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/spf13/afero"
)

func TestCmd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cmd Suite")
}

func HaveContents(afs afero.Afero, content string) types.GomegaMatcher {
	return &haveContentMatcher{
		afs:     afs,
		content: content,
	}
}

type haveContentMatcher struct {
	afs      afero.Afero
	content  string
	expected string
}

func (matcher *haveContentMatcher) Match(actual interface{}) (success bool, err error) {
	pathStr, ok := actual.(string)
	if !ok {
		return false, fmt.Errorf("beAnEmptyFile matcher expects a string")
	}

	bs, err := matcher.afs.ReadFile(pathStr)
	if err != nil {
		return false, err
	}

	matcher.expected = string(bs)

	return matcher.expected == matcher.content, nil
}

func (matcher *haveContentMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf(
		"Expected '%s' to have contents\n\t'%s' but was \n\t'%s'",
		actual,
		matcher.content,
		matcher.expected,
	)
}

func (matcher *haveContentMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf(
		"Expected '%s' to not have contents\n\t'%s' but was \n\t'%s'",
		actual,
		matcher.content,
		matcher.expected,
	)
}

func BeAnEmptyFile(afs afero.Afero) types.GomegaMatcher {
	return &beAnEmptyFileMatcher{
		afs: afs,
	}
}

type beAnEmptyFileMatcher struct {
	afs afero.Afero
}

func (matcher *beAnEmptyFileMatcher) Match(actual interface{}) (success bool, err error) {
	pathStr, ok := actual.(string)
	if !ok {
		return false, fmt.Errorf("beAnEmptyFile matcher expects a string")
	}

	bs, err := matcher.afs.ReadFile(pathStr)
	if err != nil {
		return false, err
	}

	return len(bs) == 0, nil
}

func (*beAnEmptyFileMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected '%s' to be an empty file", actual)
}

func (*beAnEmptyFileMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected not to be '%s' to be an empty file", actual)
}

func BeADir(afs afero.Afero) types.GomegaMatcher {
	return &beADirMatcher{
		afs: afs,
	}
}

type beADirMatcher struct {
	afs afero.Afero
}

func (matcher *beADirMatcher) Match(actual interface{}) (success bool, err error) {
	pathStr, ok := actual.(string)
	if !ok {
		return false, fmt.Errorf("beADir matcher expects a string")
	}

	return matcher.afs.DirExists(pathStr)
}

func (*beADirMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected '%s' to be a directory", actual)
}

func (*beADirMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected not to be '%s' to be a directory", actual)
}

type mockFs struct {
	mockCreate   func(string) (afero.File, error)
	mockMkdirAll func(string, os.FileMode) error
	mockOpen     func(string) (afero.File, error)
	mockOpenFile func(string, int, os.FileMode) (afero.File, error)
	mockRemove   func(string) error
	mockChmod    func(string, os.FileMode) error
	memFs        afero.Fs
}

func newMockFs() *mockFs {
	return &mockFs{
		memFs: afero.NewMemMapFs(),
	}
}

func (m *mockFs) Create(name string) (afero.File, error) {
	if m.mockCreate != nil {
		return m.mockCreate(name)
	}

	return m.memFs.Create(name)
}

func (m *mockFs) Mkdir(name string, perm os.FileMode) error {
	panic("not implemented") // TODO: Implement
}

func (m *mockFs) MkdirAll(path string, perm os.FileMode) error {
	if m.mockMkdirAll != nil {
		return m.mockMkdirAll(path, perm)
	}

	return m.memFs.MkdirAll(path, perm)
}

func (m *mockFs) Open(name string) (afero.File, error) {
	if m.mockOpen != nil {
		return m.mockOpen(name)
	}

	return m.memFs.Open(name)
}

func (m *mockFs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	if m.mockOpenFile != nil {
		return m.mockOpenFile(name, flag, perm)
	}

	return m.memFs.OpenFile(name, flag, perm)
}

func (m *mockFs) Remove(name string) error {
	if m.mockRemove != nil {
		return m.mockRemove(name)
	}

	return m.memFs.Remove(name)
}

func (m *mockFs) RemoveAll(path string) error {
	panic("not implemented") // TODO: Implement
}

func (m *mockFs) Rename(oldname string, newname string) error {
	panic("not implemented") // TODO: Implement
}

func (m *mockFs) Stat(name string) (os.FileInfo, error) {
	panic("not implemented") // TODO: Implement
}

func (m *mockFs) Name() string {
	panic("not implemented") // TODO: Implement
}

func (m *mockFs) Chown(name string, uid, gid int) error {
	panic("not implemented") // TODO: Implement
}

func (m *mockFs) Chmod(name string, mode os.FileMode) error {
	if m.mockChmod != nil {
		return m.mockChmod(name, mode)
	}

	return m.memFs.Chmod(name, mode)
}

func (m *mockFs) Chtimes(name string, atime, mtime time.Time) error {
	panic("not implemented") // TODO: Implement
}

type mockFile struct {
	mockRead        func([]byte) (int, error)
	mockClose       func() error
	mockWrite       func([]byte) (int, error)
	mockWriteString func(string) (int, error)
	memFile         afero.File
}

func newMockFile(fs afero.Fs, filename string) *mockFile {
	f, _ := fs.Create(filename)

	return &mockFile{
		memFile: f,
	}
}

func (m *mockFile) Close() error {
	// always close our mock file
	m.memFile.Close()

	if m.mockClose != nil {
		return m.mockClose()
	}

	return nil
}

func (m *mockFile) Read(p []byte) (n int, err error) {
	if m.mockRead != nil {
		return m.mockRead(p)
	}

	return m.memFile.Read(p)
}

func (m *mockFile) ReadAt(p []byte, off int64) (n int, err error) {
	panic("not implemented") // TODO: Implement
}

func (m *mockFile) Seek(offset int64, whence int) (int64, error) {
	panic("not implemented") // TODO: Implement
}

func (m *mockFile) Write(p []byte) (n int, err error) {
	if m.mockWrite != nil {
		return m.mockWrite(p)
	}

	return m.memFile.Write(p)
}

func (m *mockFile) WriteAt(p []byte, off int64) (n int, err error) {
	panic("not implemented") // TODO: Implement
}

func (m *mockFile) Name() string {
	panic("not implemented") // TODO: Implement
}

func (m *mockFile) Readdir(count int) ([]os.FileInfo, error) {
	panic("not implemented") // TODO: Implement
}

func (m *mockFile) Readdirnames(n int) ([]string, error) {
	panic("not implemented") // TODO: Implement
}

func (m *mockFile) Stat() (os.FileInfo, error) {
	panic("not implemented") // TODO: Implement
}

func (m *mockFile) Sync() error {
	panic("not implemented") // TODO: Implement
}

func (m *mockFile) Truncate(size int64) error {
	panic("not implemented") // TODO: Implement
}

func (m *mockFile) WriteString(s string) (ret int, err error) {
	if m.mockWriteString != nil {
		return m.mockWriteString(s)
	}

	return m.memFile.WriteString(s)
}
